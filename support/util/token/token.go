package token

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math/rand"
	"time"
)

const signLen = 20
const fixedLen = 30 // 30 = validBegin(4)+validTime(2)+rand(4)+sign(20)

var endian = binary.BigEndian // 采用大端对齐

var (
	ErrMissKey    = errors.New("miss sign key")
	ErrMissToken  = errors.New("miss token")
	ErrTokenShort = errors.New("token is too short")
	ErrNotBase64  = errors.New("token not base64")
	ErrFormat     = errors.New("format invalid")
	ErrEarly      = errors.New("before valid begin")
	ErrExpired    = errors.New("token expired")
	ErrWrongKey   = errors.New("sign key is wrong")
)

// GenToken 生成token
// token规则: {validBegin}{validTime}{rand}{userId}{sign}
// validBegin: 4字节，校验开始时间，单位秒
// validTime: 2字节，有效期，单位分钟
// rand: 4字节，随机数
// info: 不定长，可以为meetingcode、用户Id、设备id（可靠的签名参数）
// sign: 20字节，前面4个字段与signKey通过hmac加密生成的签名
// 最终生成base64
func GenToken(signKey string, info string, validBegin time.Time, validTime time.Duration) (string, error) {
	if signKey == "" {
		return "", ErrMissKey
	}

	var uLen = len(info)
	b := make([]byte, fixedLen+uLen)
	// 固定长度
	vb := intToBytes(uint32(validBegin.Unix())) // 校验开始时间
	copy(b[:4], vb)
	vt := intToBytes(uint16(validTime / time.Minute)) // 有效期
	copy(b[4:6], vt)
	r := intToBytes(rand.Uint32()) // 32位随机数
	copy(b[6:10], r)

	// userId
	copy(b[10:10+uLen], info)

	// sign
	mac := hmac.New(sha1.New, []byte(signKey))
	mac.Write(b[:10+uLen])
	copy(b[10+uLen:], genSign(signKey, b[:10+uLen]))

	return base64.StdEncoding.EncodeToString(b), nil
}

// CheckToken 校验token
func CheckToken(signKey string, token string) error {
	if signKey == "" {
		return ErrMissKey
	}
	if token == "" {
		return ErrMissToken
	}

	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return ErrNotBase64
	}
	if len(b) < fixedLen { // 不能比固定字段长度短
		return ErrTokenShort
	}
	now := time.Now().Unix()
	// 验证开始时间
	validBegin := int64(bytesToUint32(b[:4]))
	if validBegin == 0 {
		return ErrFormat
	}
	if now < validBegin {
		return ErrEarly
	}
	// 验证结束时间
	validMin := int64(bytesToUint16(b[4:6]))
	if validMin == 0 {
		return ErrFormat
	}
	if now > validBegin+validMin*60 {
		return ErrExpired
	}
	// 验证signKey
	sign := make([]byte, signLen)
	copy(sign, genSign(signKey, b[:len(b)-signLen]))

	signInToken := b[len(b)-signLen:]
	if !bytes.Equal(sign, signInToken) {
		return ErrWrongKey
	}

	return nil
}

func genSign(signKey string, val []byte) []byte {
	mac := hmac.New(sha1.New, []byte(signKey))
	mac.Write(val)
	return mac.Sum(nil)
}

func intToBytes(n interface{}) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, n)
	return bytesBuffer.Bytes()
}
func bytesToUint32(b []byte) uint32 {
	bytesBuffer := bytes.NewBuffer(b)

	var x uint32
	_ = binary.Read(bytesBuffer, endian, &x)

	return x
}
func bytesToUint16(b []byte) uint16 {
	bytesBuffer := bytes.NewBuffer(b)

	var x uint16
	_ = binary.Read(bytesBuffer, endian, &x)

	return x
}
