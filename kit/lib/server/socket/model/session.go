package model

import (
	"github.com/gogf/gf/v2/container/gmap"
	"time"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/util"
)

// Session struct
type Session struct {
	Ctx  kctx.Ctx
	sID  string
	uID  string
	conn *Conn
	// settings
	// isLogin  是否登录
	// isHangUp 是事挂起
	settings                *gmap.StrStrMap
	latestBeatTime          time.Time // 最后一次心跳的时间
	heartBeatIntervalSecond int       // 心中间隔
	//heartBeatFailedTimes    int       //  多少次没有心跳， 判定断线
	curHeartBeatTime int // 当前心跳失败的次数
	loginFailedTimes int // 登录失败的次数,  登录重试次数：若客户端连续$loginFailRetries次登录失败，则主动断开该客户端的TCP连接。
}

const (
	IsLogin  = "isLogin"
	IsHangUp = "isHangUp"
)

// NewSession create a new session
func NewSession(ctx kctx.Ctx, hbInterval int, conn *Conn) *Session {
	id := util.NewUUIDStr()
	if hbInterval == 0 {
		hbInterval = 30
	}

	session := &Session{
		Ctx:                     ctx,
		sID:                     id,
		conn:                    conn,
		settings:                gmap.NewStrStrMap(true),
		latestBeatTime:          time.Time{},
		heartBeatIntervalSecond: hbInterval,
	}

	return session
}

// Dead 外层框架根据是否dead来维护 连接池
func (s *Session) Dead() bool {
	if s.conn == nil && !s.IsLogin() && s.IsHangUp() {
		return true
	}

	return false
}

// IsHeartbeatExpire  心跳是否超时， true: 超时  false: 没有超时
func (s *Session) IsHeartbeatExpire(beatTime time.Time) bool {
	if s.latestBeatTime.IsZero() {
		s.latestBeatTime = beatTime
	}
	if s.HeartbeatExpire(beatTime) {
		s.latestBeatTime = beatTime

		deadTime := s.latestBeatTime.Add(time.Second * time.Duration(s.heartBeatIntervalSecond))

		s.Ctx.Infof("heartbeat is expire, deadTime at %v, beatTime %v",
			deadTime.Format(time.RFC3339),
			beatTime.Format(time.RFC3339),
		)
		return true
	}
	s.latestBeatTime = beatTime
	return false
}

// HeartbeatExpire 单纯判断是否心跳超时, true: 超时， false 没有超时
func (s *Session) HeartbeatExpire(n time.Time) bool {
	if s.latestBeatTime.IsZero() {
		return false
	}
	deadTime := s.latestBeatTime.Add(
		time.Second *
			time.Duration(s.heartBeatIntervalSecond))

	return n.After(deadTime)
}

func (s *Session) Close() {
	// 清空登录， hangUp 的标记
	s.settings = gmap.NewStrStrMap(true)
	// 用户id 清空
	s.Ctx.Infof("session %s userId %s closed", s.sID, s.uID)

	s.uID = ""

	if s.conn != nil {
		s.conn.Close()
	}
}

// GetSessionID get session ID
func (s *Session) GetSessionID() string {
	return s.sID
}

// BindUserID bind a user ID to session
func (s *Session) BindUserID(uid string) {
	s.uID = uid
}

// GetUserID get user ID
func (s *Session) GetUserID() string {
	return s.uID
}

// GetConn get zero.Conn pointer
func (s *Session) GetConn() (*Conn, bool) {
	if !s.IsLogin() {
		return nil, false
	}

	if s.IsHangUp() {
		return nil, false
	}

	return s.conn, true
}

// SetConn set a zero.Conn to session
func (s *Session) SetConn(conn *Conn) {
	s.conn = conn
}

// GetSetting get setting
func (s *Session) GetSetting(key string) string {
	return s.settings.Get(key)
}

// SetSetting set setting
func (s *Session) SetSetting(key, value string) {
	s.settings.Set(key, value)
}

// SetLoginFlag 设置登录标记， 如果登录失败次数到达上限， 则返回true
// flag: yes/no
func (s *Session) SetLoginFlag(maxFailedTimes int, flag string) (isMax bool) {
	switch flag {
	case "yes":
		s.loginFailedTimes = 0
		s.settings.Set(IsLogin, "yes")

	default:
		s.loginFailedTimes++
		s.settings.Set(IsLogin, "no")

		if s.loginFailedTimes >= maxFailedTimes {
			return true
		}
	}

	return false
}
func (s *Session) SetHangUp() {
	s.settings.Set(IsHangUp, "yes")
}

func (s *Session) IsLogin() bool {
	return s.GetSetting(IsLogin) == "yes"
}
func (s *Session) IsHangUp() bool {
	return s.GetSetting(IsHangUp) == "yes"
}

// Check 检查登录状态， 心跳状态等是否正常
func (s *Session) Check() bool {
	// 是否登录
	if !s.IsLogin() {
		return false
	}

	// 心跳是否超时
	if s.HeartbeatExpire(time.Now()) {
		return false
	}
	// 是否已经挂起
	if s.IsHangUp() {
		return false
	}

	return true
}
