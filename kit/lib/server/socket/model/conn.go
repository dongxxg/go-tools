package model

import (
	"bufio"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net"
	"time"
)

// Conn wrap net.Conn
type Conn struct {
	sid        string
	rawConn    net.Conn
	sendCh     chan Message
	Done       chan error
	hbTimer    *time.Timer
	name       string
	MessageCh  chan []byte
	hbInterval time.Duration
	hbTimeout  time.Duration
	// readFunc 从连接中读取数据的函数
	readFunc func(r *bufio.Reader) ([]byte, error)
	Timeout  time.Duration // 写入消息后， 等待返回的 超时时间
}
type Message struct {
	Content []byte
	ErrChan chan error
}

// GetName Get conn name
func (c *Conn) GetName() string {
	return c.name
}
func (c *Conn) RemoteAddr() net.Addr {
	return c.rawConn.RemoteAddr()
}

// NewConn create new conn
func NewConn(c net.Conn, hbInterval, readTimeout time.Duration) *Conn {
	conn := &Conn{
		rawConn:    c,
		sendCh:     make(chan Message, 1),
		Done:       make(chan error),
		hbTimer:    nil,
		MessageCh:  make(chan []byte, 1),
		hbInterval: hbInterval,
		hbTimeout:  readTimeout,
		readFunc:   nil,
		Timeout:    0,
	}
	if conn.Timeout == 0 {
		conn.Timeout = time.Second * 5
	}

	conn.name = c.RemoteAddr().String()
	conn.hbTimer = time.NewTimer(conn.hbInterval)

	if conn.hbInterval == 0 {
		conn.hbTimer.Stop()
	}

	return conn
}

// Close connection
func (c *Conn) Close() {
	c.hbTimer.Stop()
	c.rawConn.Close()
}

// SendMessage send message
func (c *Conn) SendMessage(pkg []byte) error {
	errChan := make(chan error)

	c.sendCh <- Message{
		Content: pkg,
		ErrChan: errChan,
	}

	ticker := time.NewTicker(c.Timeout)
	select {
	case <-ticker.C:
		return fmt.Errorf("sendMessage timeout %s", c.Timeout.String())
	case err, ok := <-errChan:
		if !ok {
			// 直接被关闭， 说明没有写入异常
			return nil
		}

		return err
	}
}

func (c *Conn) SetReadFunc(r func(br *bufio.Reader) ([]byte, error)) {
	c.readFunc = r
}

// WriteCoroutine write coroutine
func (c *Conn) WriteCoroutine(ctx context.Context) {
	//hbData := make([]byte, 0)

	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-c.sendCh:

			if _, err := c.rawConn.Write(msg.Content); err != nil {

				msg.ErrChan <- err

				c.Done <- err
			}

			close(msg.ErrChan)
			//case <-c.hbTimer.C:
			// 不用给客户端发送心跳
			//hbMessage := NewMessage(MsgHeartbeat, hbData)
			//c.SendMessage(hbMessage)
			//// 设置心跳timer
			//if c.hbInterval > 0 {
			//	c.hbTimer.Reset(c.hbInterval)
			//}
		}
	}
}

// ReadCoroutine read coroutine
func (c *Conn) ReadCoroutine(ctx context.Context) {

	for {
		select {
		case <-ctx.Done():
			return

		default:
			// 设置超时
			if c.rawConn == nil {
				continue
			}

			//if c.hbInterval > 0 {
			//err := c.rawConn.SetReadDeadline(time.Now().Add(c.hbTimeout))
			//if err != nil {
			//	c.Done <- errors.Wrapf(err, "conn.SetReadDeadline failed")
			//	continue
			//}
			//}
			buf := bufio.NewReader(c.rawConn)
			data, err := c.readFunc(buf)
			if err != nil {
				c.Done <- errors.Wrapf(err, "readFunc failed")

				if errors.Is(err, io.EOF) {
					c.Close()
				}
				continue
			}

			c.MessageCh <- data
		}
	}
}
