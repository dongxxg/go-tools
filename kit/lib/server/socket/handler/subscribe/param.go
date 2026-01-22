package subscribe

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/container/gmap"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"sync"
	"time"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/util"
)

type MsgType int

const (
	MsgTypeRealTimeAlarm          MsgType = 0
	MsgTypeReqLoginAlarm          MsgType = 1
	MsgTypeAckLoginAlarm          MsgType = 2
	MsgTypeReqSyncAlarmMsg        MsgType = 3
	MsgTypeAckSyncAlarmMsg        MsgType = 4
	MsgTypeReqSyncAlarmFile       MsgType = 5
	MsgTypeAckSyncAlarmFile       MsgType = 6
	MsgTypeAckSyncAlarmFileResult MsgType = 7
	MsgTypeReqHeartBeat           MsgType = 8
	MsgTypeAckHeartBeat           MsgType = 9
	MsgTypeCloseConnAlarm         MsgType = 10
)

const (
	ParamSockMode             = "sockMode"             // 通信模式
	ParamMsgHeaderFormat      = "msgHeaderFormat"      // 消息头格式
	ParamLoginReqMsg          = "loginReqMsg"          // 登录请求报文
	ParamLoginUserName        = "loginUserName"        // 登录用户名
	ParamLoginPassword        = "loginPassword"        // 登录密码
	ParamLoginResMsgSucc      = "loginResMsgSucc"      // 登录成功响应报文
	ParamLoginResMsgFail      = "loginResMsgFail"      // 登录失败响应报文
	ParamLoginResMsgType      = "loginResMsgType"      // 登录响应消息类型
	ParamLoginFailRetries     = "loginFailRetries"     // 登录允许失败次数
	ParamHeartbeatResMsg      = "heartbeatResMsg"      // 心跳响应报文
	ParamSeqNoExpr            = "seqNoExpr"            // 请求标识截取正则
	ParamHeartbeatReqMsgExpr  = "heartbeatReqMsgExpr"  // 心跳请求报文校验正则
	ParamCloseConnReqMsg      = "closeConnReqMsg"      // 关闭连接请求报文
	ParamRealTimeAlarmMsgType = "realTimeAlarmMsgType" // 实时告警消息类型
	ParamShakeHandFails       = "shakehandFails"       // 握手失败次数
	ParamHeartbeatInterval    = "heartbeatInterval"    // 心中间隔
)

func NewParam(data *gmap.StrStrMap, df func(key string, val string) string) *Param {
	return &Param{
		Data:       data,
		Decryption: df,
	}
}

type Param struct {
	Data         *gmap.StrStrMap
	lenOfHeader  int // header字段长度
	headerFormat *msgHeaders
	Decryption   func(key, val string) string
}

func (p *Param) GetLoginMsg() string {
	return p.getParam(ParamLoginReqMsg)
}
func (p *Param) GetHeartbeatMsg() string {
	return p.getParam(ParamHeartbeatResMsg)
}
func (p *Param) GetHeaderByType(ctx kctx.Ctx, m MsgType, lenOfBody int) []byte {
	return p.genHeader(ctx, m, lenOfBody)
}
func (p *Param) GetParam(key string) string {
	return p.getParam(key)
}

// getParam
func (p *Param) getParam(key string) string {
	v := p.Data.Get(key)
	vs := util.GetVariable(v)
	ignore := gset.NewStrSetFrom([]string{"seqNoExpr"})
	for _, variable := range vs {
		if ignore.Contains(variable) {
			continue
		}
		v2 := p.Data.Get(variable)
		// 进行解密操作
		v2 = p.Decryption(variable, v2)
		v = strings.ReplaceAll(v, fmt.Sprintf("${%s}", variable), v2)
	}

	return v
}

// GetParamInt 获取int类型参数
func (p *Param) GetParamInt(key string) (int, error) {
	n := p.getParam(key)
	if n == "" {
		// 默认登录失败3次
		return 3, nil
	}
	nn, err := strconv.Atoi(n)
	if err != nil {
		return 0, fmt.Errorf("invalid key %s", n)
	}

	return nn, nil
}

func (p *Param) GetMsgType(header []byte) (MsgType, error) {
	if p.headerFormat == nil {
		_, err := p.msgHeaderFormat()
		if err != nil {
			return 0, errors.Wrapf(err, "msgHeaderFormat")
		}
	}
	begin, end, err := p.headerFormat.indexOfMsgType()
	if err != nil {
		return 0, errors.Wrapf(err, "indexOfMsgType")
	}
	if end-begin == 1 {
		return MsgType(int(header[begin:end][0])), nil
	}

	h := binary.BigEndian.Uint16(header[begin:end])

	return MsgType(int(h)), nil
}
func (p *Param) GetLenOfBody(header []byte) (int, error) {
	// 找到lenOfBody 字段的长度及起点
	if p.headerFormat == nil {
		_, err := p.msgHeaderFormat()
		if err != nil {
			return 0, errors.Wrapf(err, "msgHeaderFormat")
		}
	}

	begin, end, err := p.headerFormat.indexOfLenOfBody()
	if err != nil {
		return 0, errors.Wrapf(err, "get indexOfLenOfBody")
	}

	h := binary.BigEndian.Uint16(header[begin:end])

	return int(h), nil
}
func (p *Param) GetLenOfHeader() (int, error) {
	if p.lenOfHeader != 0 {
		return p.lenOfHeader, nil
	}
	mh, err := p.msgHeaderFormat()
	if err != nil {
		return 0, errors.Wrapf(err, "get msgHeaderFormat")
	}
	// 遍历所有, 长度加起来
	l := 0
	for _, v := range mh.h {
		l += v.Length
	}
	if l == 0 {
		return 0, fmt.Errorf("invalid msgHeaderFormat, len of header is 0")
	}

	p.lenOfHeader = l
	return l, nil
}

type msgHeaders struct {
	h []msgHeader
	sync.Mutex
}

func (h *msgHeaders) Range(f func(header msgHeader) bool) {
	h.Lock()
	defer h.Unlock()
	for _, v := range h.h {
		if !f(v) {
			break
		}

	}
}

/*
[

	{"name":"startSign", "type": "16", "length": 2, "value": "0xFFFF"},
	{"name":"msgType", "type": "10", "length": 1, "value": ""},
	{"name":"timeStamp", "type": "10", "length": 4, "value": ""},
	{"name":"lenOfBody", "type": "10", "length": 2, "value": ""}

]
*/
func (h *msgHeaders) startSign() (msgHeader, bool) {
	return h.getByName("startSign")
}
func (h *msgHeaders) msgType() (msgHeader, bool) {
	return h.getByName("msgType")
}
func (h *msgHeaders) timeStamp() (msgHeader, bool) {
	return h.getByName("timeStamp")
}
func (h *msgHeaders) lenOfBody() (msgHeader, bool) {
	return h.getByName("timeStamp")
}

func (h *msgHeaders) getByName(name string) (msgHeader, bool) {
	for _, v := range h.h {
		if strings.ToLower(v.Name) == strings.ToLower(name) {
			return v, true
		}
	}
	return msgHeader{}, false

}

func (h *msgHeaders) indexOfMsgType() (int, int, error) {
	return h.beginAndEnd("msgType")
}

func (h *msgHeaders) indexOfLenOfBody() (int, int, error) {
	return h.beginAndEnd("lenOfBody")
}

func (h *msgHeaders) beginAndEnd(key string) (int, int, error) {
	begin := 0
	for _, v := range h.h {
		if v.Name == key {
			return begin, begin + v.Length, nil
		}
		begin += v.Length
	}

	return 0, 0, fmt.Errorf("invalid msgHeaderFormat, %s not found in msgHeaderFormat", key)
}

type msgHeader struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Length int    `json:"length"`
	Value  string `json:"value"`
}

// msgHeaderFormat
func (p *Param) msgHeaderFormat() (msgHeaders, error) {
	v := p.Data.Get(ParamMsgHeaderFormat)
	out := make([]msgHeader, 0)
	if err := json.Unmarshal([]byte(v), &out); err != nil {
		return msgHeaders{}, errors.Wrapf(err, "invalid msgHeaderFormat, %s", v)
	}

	p.headerFormat = &msgHeaders{h: out}

	return msgHeaders{h: out}, nil
}

// genHeader
/*
[
    {"name":"startSign", "type": "16", "length": 2, "value": "0xFFFF"},
    {"name":"MsgType", "type": "10", "length": 1, "value": ""},
    {"name":"timeStamp", "type": "10", "length": 4, "value": ""},
    {"name":"lenOfBody", "type": "10", "length": 2, "value": ""}
]
*/
// 每个消息的消息头由16个字节(128bits)来表示，均采用大字节序上报，如下：
func (p *Param) genHeader(ctx kctx.Ctx, msgType MsgType, lenOfBody int) []byte {

	if p.lenOfHeader == 0 {
		n, err := p.GetLenOfHeader()
		if err != nil {
			ctx.Errorf("genHeader: %v", err)
			return nil
		}
		p.lenOfHeader = n
	}
	// StartSign   开始标志   16 固定为0x7EE7  用于跟踪调试使用，代表一条消息的开始
	headerBytes := make([]byte, p.lenOfHeader)

	begin := 0

	if p.headerFormat == nil {
		h, err := p.msgHeaderFormat()
		if err != nil {
			ctx.Errorf("genHeader: %v", err)
			return nil
		}
		p.headerFormat = &h
	}
	p.headerFormat.Range(func(header msgHeader) bool {
		if header.Length == 0 {
			return true
		}
		if header.Length == 1 {
			headerBytes = append(headerBytes)
		}

		var data uint16
		switch strings.ToLower(header.Name) {
		case "msgtype":
			data = uint16(msgType)
		case "timestamp":
			data = uint16(time.Now().Unix())
		case "lenofbody":
			data = uint16(lenOfBody)
		default:
			switch header.Type {
			case "16":
				i, err := strconv.ParseInt(header.Value, 0, 0)
				if err != nil {
					ctx.Errorf("convert header %v  to int failed: %v", header, err)
					return true
				}
				data = uint16(i)
			case "10":
				i, err := strconv.ParseInt(header.Value, 10, 64)
				if err != nil {
					ctx.Errorf("convert header %v  to int failed: %v", header, err)
					return true
				}
				data = uint16(i)
			}
		}
		b := headerBytes[begin : begin+header.Length]
		switch header.Length {
		case 1:
			headerBytes[begin : begin+header.Length][0] = byte(data)
		case 2:
			binary.BigEndian.PutUint16(b, data)
		case 4:
			binary.BigEndian.PutUint32(b, uint32(data))
		case 8:
			binary.BigEndian.PutUint64(b, uint64(data))
		default:
			binary.BigEndian.PutUint16(b, data)
		}
		begin += header.Length
		return true
	})

	return headerBytes
}
