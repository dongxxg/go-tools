package subscribe

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"regexp"
	"strings"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/lib/server/socket/model"
)

type Handler struct {
	ctx                kctx.Ctx
	p                  Param // 规则参数
	lenOfHeader        int
	isLoginSuccess     bool                         // 是否登录成功
	loginMsg           string                       // 登录成功的消息
	Decryption         func(key, val string) string // 解密函数， 有的key , 可能根据key 进行base64 解密
	loginFailedRetries int
	heartbeatReg       *regexp.Regexp
}

func NewHandler(p Param) (*Handler, error) {
	s := &Handler{
		p: p,
	}
	if s.loginFailedRetries == 0 {
		times, err := s.p.GetParamInt(ParamLoginFailRetries)
		if err != nil {
			times = 3
		}
		s.loginFailedRetries = times
	}
	expr := s.p.GetParam(ParamHeartbeatReqMsgExpr)
	r, err := regexp.Compile(expr)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid ParamHeartbeatReqMsgExpr %s", expr)
	}

	s.heartbeatReg = r

	return s, nil
}

func (s *Handler) OnConnect(sess *model.Session) {
	sess.Ctx.Infof("connect success")
}

func (s *Handler) OnDisConnect(sess *model.Session, err error) {
	sess.Ctx.Error(err)
}
func (s *Handler) WrapperMsg(msg []byte, msgType int) []byte {
	header := s.p.GetHeaderByType(s.ctx, MsgType(msgType), len(msg))
	header = append(header, msg...)

	return header
}

func (s *Handler) ReadFunc(r *bufio.Reader) ([]byte, error) {
	var err error
	if s.lenOfHeader == 0 {
		s.lenOfHeader, err = s.p.GetLenOfHeader()
		if err != nil {
			return nil, errors.Wrapf(err, "GetLenOfHeader")
		}
	}
	// 获取指定长度的header
	header, err := r.Peek(s.lenOfHeader)
	if err != nil {
		return nil, errors.Wrapf(err, "Peek")
	}
	// 获取 body 的长度
	length, err := s.p.GetLenOfBody(header)
	if err != nil {
		return nil, errors.Wrapf(err, "GetLenOfBody")
	}

	info := make([]byte, length+s.lenOfHeader)

	_, err = r.Read(info)
	if err != nil {
		return nil, errors.Wrapf(err, "Read")
	}

	return info, nil
}

func (s *Handler) parseLogin(ctx kctx.Ctx, body []byte) bool {
	// 规则中： reqLoginAlarm;user=${loginUser};key=${loginPass};type=msg
	// 规则中： reqLoginAlarm;user=${loginUser};key=${loginPass};type=msg^^reqLoginAlarm;user=${loginUser};key=${loginPass};type=msg

	bodyStr := string(body)
	// body：reqLoginAlarm;user=yiy;key=qw#$@;type=msg
	if s.loginMsg == "" {
		s.loginMsg = s.p.GetLoginMsg()
	}
	lst := strings.Split(s.loginMsg, "^^")
	ctx.Debugf("login info %s", s.loginMsg)
	for _, v := range lst {
		if v == bodyStr {
			return true
		}
	}

	return false
}

// getHeartbeatReqId
// ackHeartBeat;reqId=33
func (s *Handler) getHeartbeatReqId(body []byte) (string, error) {
	// 获取 reqId 的值
	ss := string(body)
	expr := s.p.GetParam(ParamSeqNoExpr)
	reg, err := regexp.Compile(expr)
	if err != nil {
		return "", fmt.Errorf("invalid seqNoExpr %s, err: %v", expr, err)
	}
	lst := reg.FindStringSubmatch(ss)
	if len(lst) < 2 {
		return "", fmt.Errorf("invalid seqNoExpr %s, lst %v", expr, lst)
	}
	var SeqNo string
	for i := 1; i < len(lst); i++ {
		if lst[i] != "" {
			SeqNo = lst[i]
			break
		}
	}

	return SeqNo, nil
}
