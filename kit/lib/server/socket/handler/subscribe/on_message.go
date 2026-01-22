package subscribe

import (
	"fmt"
	"strings"
	"time"
	"unitechs.com/cno/kit/lib/server/socket/model"
)

func (s *Handler) OnMessage(sess *model.Session, data []byte) {
	header := data[:s.lenOfHeader]
	msgType, err := s.p.GetMsgType(header)
	if err != nil {
		s.ctx.Errorf("GetMsgType %v", err)
		return
	}
	body := data[s.lenOfHeader:]

	sess.Ctx.Infof("request msg, msgType %d body  %s", msgType, string(body))

	var res string
	switch msgType {
	case MsgTypeReqLoginAlarm:
		// 消息样例：reqLoginAlarm;user=yiy;key=qw#$@;type=msg
		if s.parseLogin(sess.Ctx, body) {
			sess.SetSetting(model.IsLogin, "yes")
			// 发送成功的返回
			res = s.p.GetParam(ParamLoginResMsgSucc)

			sess.SetLoginFlag(s.loginFailedRetries, "yes")
		} else {
			// 发送失败的返回
			res = s.p.GetParam(ParamLoginResMsgFail)
			sess.Ctx.Errorf("login failed")
			if sess.SetLoginFlag(s.loginFailedRetries, "no") {
				sess.Ctx.Errorf("login failures exceeded %s", s.loginFailedRetries)
				sess.Close()
			}
		}

	case MsgTypeReqHeartBeat:
		// reqHeartBeat;reqId=33
		/* 心跳：若$heartbeatInterval秒后仍然未收到心跳请求报文，
		则主动断开该客户端的TCP连接；若正常收到心跳请求报文，
		则采用$seqNoExpr截取心跳请求报文，得到的值替换$heartbeatResMsg中的变量seqNoExpr，
		作为最终心跳响应报文发送（其中msgType=$heartbeatMsgType）。
		备注：心跳请求报文需满足正则$heartbeatReqMsgExpr。
		*/
		if !s.heartbeatReg.Match(body) {
			sess.Ctx.Errorf("invalid heartbeat msg not  match %s", s.heartbeatReg.String())
			res = fmt.Sprintf("ackHeartBeat;reqId=;resDesc=invalid heartbeat msg, not match regexpr.")

		} else {
			id, err := s.getHeartbeatReqId(body)
			if err != nil {
				sess.Ctx.Errorf("invalid heartbeat get ReqId failed: %s", body)
				id = "1" // 如果获取不到reqId， 默认返回 1
			}
			hb := s.p.GetParam(ParamHeartbeatResMsg)
			hb = strings.ReplaceAll(hb, "${seqNoExpr}", id)

			// ackHeartBeat;reqId=33

			if sess.IsHeartbeatExpire(time.Now()) {
				// 心跳超时
				sess.Ctx.Errorf("heartbeat is expire")
				sess.Close()

				res = fmt.Sprintf("ackHeartBeat;reqId=%s;resDesc=heartbeat-expired", id)
			} else {
				// 没有超时
				res = hb
				if !sess.IsLogin() {
					sess.Ctx.Info("is not login, no need to heartbeat response")
					return
				}
			}
		}

	case MsgTypeCloseConnAlarm:
		sess.Close()
		return

	case MsgTypeReqSyncAlarmFile,
		MsgTypeAckSyncAlarmFileResult,
		MsgTypeReqSyncAlarmMsg:
		sess.Ctx.Infof("not support msgType %s yet", msgType)
		return
	}
	sess.Ctx.Infof("response sessionId %s msgType %d msg %s", sess.GetSessionID(), msgType, res)
	msg := s.p.GetHeaderByType(sess.Ctx, msgType, len([]byte(res)))
	msg = append(msg, []byte(res)...)
	session, ok := sess.GetConn()
	if !ok {
		sess.Ctx.Errorf("get conn failed")
		return
	}

	session.SendMessage(msg)
}
