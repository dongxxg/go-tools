package socket

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/pkg/errors"
	"net"
	"sync"
	"time"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/lib/server/socket/handler/subscribe"
	"unitechs.com/cno/kit/lib/server/socket/model"
	"unitechs.com/cno/kit/template/cache"
)

// IHandler 处理接口， 注入不同的handler, 实现不同的socketServer
type IHandler interface {
	//OnMessage 接收到msg的真正的逻辑
	OnMessage(sess *model.Session, data []byte)
	OnConnect(sess *model.Session)
	OnDisConnect(sess *model.Session, err error)
	// ReadFunc 读取
	ReadFunc(r *bufio.Reader) ([]byte, error)
	// WrapperMsg 发送的消息， 需要包裹头等消息
	WrapperMsg(msg []byte, msgType int) []byte
}

func NewServerWithInterface(ctx kctx.Ctx, network, addr string, param interface{}) (*Server, error) {
	b, _ := json.Marshal(param)
	pMap := gmap.NewStrStrMap(true)
	if err := pMap.UnmarshalJSON(b); err != nil {
		return nil, errors.Wrapf(err, "unmarshal socket param to map")
	}
	p := subscribe.NewParam(pMap, func(key string, val string) string {
		// 是否加密
		return cache.GetPassword(ctx, val, key)
	})
	handler, err := subscribe.NewHandler(*p)
	if err != nil {
		return nil, errors.Wrapf(err, "new socket server handler")
	}
	server, err := NewServerWith(ctx, network, addr, handler)
	if err != nil {
		ctx.Errorf("new socket server failed: %v", err)
		return nil, errors.Wrapf(err, "new socket server")
	}

	return server, nil
}

func NewServerWith(ctx kctx.Ctx, network, laddr string, h IHandler) (*Server, error) {
	srv, err := NewServer(ctx, network, laddr)
	if err != nil {
		return nil, err
	}
	srv.Set(func(s *Server) {
		s.HeartBeatIntervalSecond = 60
		s.Timeout = 10
		s.WrapMsg = h.WrapperMsg
	})
	srv.RegMessageHandler(h.OnMessage)
	srv.RegDisconnectHandler(h.OnDisConnect)
	srv.RegConnectHandler(h.OnConnect)
	srv.RegReadFunc(h.ReadFunc)

	return srv, nil
}

// Server struct
type Server struct {
	ctx      kctx.Ctx
	status   int
	listener net.Listener
	stopCh   chan error

	OnMessage               func(sess *model.Session, data []byte)
	OnConnect               func(sess *model.Session)
	OnDisconnect            func(sess *model.Session, err error)
	WrapMsg                 func(msg []byte, msgType int) []byte
	sessions                sync.Map
	laddr                   string
	ReadFunc                func(r *bufio.Reader) ([]byte, error)
	Timeout                 time.Duration // 读取conn 的超时时间
	HeartBeatIntervalSecond int           // 每次心跳的间隔
	SendTimeout             int
}

// NewServer create a new socket service
func NewServer(ctx kctx.Ctx, network, laddr string) (*Server, error) {

	// network tcp tcp6  udp
	l, err := net.Listen(network, laddr)

	if err != nil {
		return nil, err
	}

	ctx.Infof("listening at %s", laddr)
	s := &Server{
		ctx:                     ctx,
		sessions:                sync.Map{},
		HeartBeatIntervalSecond: 0,
		laddr:                   laddr,
		status:                  STInited,
		listener:                l,
		stopCh:                  make(chan error),
		ReadFunc:                nil,
		Timeout:                 time.Second * time.Duration(10), // 10s 的超时时间
	}

	return s, nil
}

// Serv Start socket service
func (s *Server) Serv() {

	s.status = STRunning
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		s.status = STStop
		cancel()
		s.listener.Close()
	}()

	go s.acceptHandler(ctx)
	go s.clear()

	for {
		select {

		case info, ok := <-s.stopCh:
			if !ok {
				return
			}
			s.ctx.Infof("recv stop signal %v", info)
			s.ctx.Infof("start to close all session %v", info)
			s.sessions.Range(func(key, value any) bool {
				ss, ok := value.(*model.Session)
				if ok {
					ss.Close()
				}

				return true
			})

			return
		}
	}
}

func (s *Server) clear() {
	ticker := time.NewTimer(time.Minute)
	for {
		select {
		case <-ticker.C:
			deleteId := gset.NewSet()
			s.sessions.Range(func(key, value any) bool {
				sess, ok := value.(*model.Session)
				if ok {
					if !sess.IsLogin() && sess.IsHangUp() {
						deleteId.Add(sess.GetSessionID())
					}
				}
				return true
			})
			deleteId.Iterator(func(v interface{}) bool {
				id, ok := v.(string)
				if ok {
					s.ctx.Infof("delete session %s", id)
					s.sessions.Delete(id)
				}
				return true
			})
		}
	}
}

// Stop socket service with reason
func (s *Server) Stop(reason string) {
	s.stopCh <- errors.New(reason)
}

// GetStatus get socket service status
func (s *Server) GetStatus() int {
	return s.status
}

// GetConnsCount get connect count
func (s *Server) GetConnsCount() int {
	var count int
	s.sessions.Range(func(k, v interface{}) bool {
		count++
		return true
	})
	return count
}

// Unicast with session ID
func (s *Server) Unicast(sid string, msg []byte) error {
	v, ok := s.sessions.Load(sid)
	if ok {
		session := v.(*model.Session)
		conn, ok := session.GetConn()
		if ok {
			return conn.SendMessage(msg)
		}
		s.sessions.Delete(sid)
		return nil
	}
	s.sessions.Delete(sid)
	return nil
}

// Broadcast to all connections
func (s *Server) Broadcast(ctx kctx.Ctx, msg []byte, msgType int) []map[string]string {
	msg = s.WrapMsg(msg, msgType)

	deleteIds := gset.NewSet(true)
	out := make([]map[string]string, 0)
	lock := sync.Mutex{}

	wg := sync.WaitGroup{}

	s.sessions.Range(func(k, v interface{}) bool {
		wg.Add(1)

		go func(k, v interface{}) {
			defer wg.Done()

			ss, ok := v.(*model.Session)
			if !ok {
				deleteIds.Add(k)
				return
			}

			conn, ok := ss.GetConn()
			if ok {
				temp := make(map[string]string)

				addr := conn.RemoteAddr()
				temp["remoteAddr"] = addr.String()
				defer func() {
					lock.Lock()
					out = append(out, temp)
					lock.Unlock()
				}()

				if !ss.Check() {
					temp["error"] = "bad session sessionId: " + ss.GetSessionID()
					return
				}

				if err := conn.SendMessage(msg); err != nil {
					ctx.Errorf("send to session %s failed: %v", ss.GetSessionID(), err)
					temp["error"] = fmt.Sprintf("send to session %s failed: %v", ss.GetSessionID(), err)
					return
				}
				ctx.Errorf("send to sesson %s success", ss.GetSessionID())
			}

			return

		}(k, v)

		return true
	})
	wg.Wait()
	return out
}

// Set 设置其他属性
func (s *Server) Set(f func(s *Server)) {
	f(s)
}

// RegMessageHandler register message handler
func (s *Server) RegMessageHandler(handler func(*model.Session, []byte)) {
	s.OnMessage = handler
}

// RegConnectHandler register connect handler
func (s *Server) RegConnectHandler(handler func(*model.Session)) {
	s.OnConnect = handler
}

// RegDisconnectHandler register disconnect handler
func (s *Server) RegDisconnectHandler(handler func(*model.Session, error)) {
	s.OnDisconnect = handler
}

// RegReadFunc register ReadFunc
func (s *Server) RegReadFunc(r func(r *bufio.Reader) ([]byte, error)) {
	s.ReadFunc = r
}

func (s *Server) acceptHandler(ctx context.Context) {
	for {
		c, err := s.listener.Accept()
		if err != nil {
			s.stopCh <- err
			return
		}

		go s.connectHandler(ctx, c)
	}
}

func (s *Server) connectHandler(ctx context.Context, c net.Conn) {
	s.ctx.Infof("coming a connection, remote addr is %s", c.RemoteAddr().String())
	conn := model.NewConn(c, time.Second*time.Duration(s.HeartBeatIntervalSecond), s.Timeout)

	if s.SendTimeout != 0 {
		conn.Timeout = time.Second * time.Duration(s.SendTimeout)
	}

	conn.SetReadFunc(s.ReadFunc)
	ctx_ := kctx.NewDefaultContext()

	session := model.NewSession(ctx_, s.HeartBeatIntervalSecond, conn)
	s.sessions.Store(session.GetSessionID(), session)

	connctx, cancel := context.WithCancel(ctx)

	defer func() {
		cancel()
		conn.Close()
		s.sessions.Delete(session.GetSessionID())
	}()

	go conn.ReadCoroutine(connctx)
	go conn.WriteCoroutine(connctx)

	if s.OnConnect != nil {
		s.OnConnect(session)
	}

	for {
		select {
		case err := <-conn.Done:

			if s.OnDisconnect != nil {
				s.OnDisconnect(session, err)
			}
			return

		case msg := <-conn.MessageCh:
			if s.OnMessage != nil {
				s.OnMessage(session, msg)
			}
		}
	}
}
