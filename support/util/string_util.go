package util

import (
	"bytes"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/rivo/uniseg"
)

// 由于strconv.FormatFloat参数较多，所以写了个简化版的
// 转换后的字符串是原来浮点数的10进制的形式，精度和实际保持一致
// 7.2=>7.2
// 7.2000=>7.2
func ConvFloat2String(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func EncryptKey(key string) string {
	if len(key) < 10 {
		return "****"
	} else {
		return key[:3] + "****" + key[len(key)-3:]
	}
}
func CutString(s, sep string) (before, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, "", false
}

func CutBytes(s, sep []byte) (before, after []byte, found bool) {
	if i := bytes.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, nil, false
}

func String2Clusters(str string) []string {
	clusters := make([]string, 0, len(str))
	gr := uniseg.NewGraphemes(str)
	for gr.Next() {
		s := gr.Str()
		if s == "\r\n" {
			clusters = append(clusters, string(s[0]), string(s[1]))
		} else {
			clusters = append(clusters, s)
		}
	}
	return clusters
}

func Clusters2String(clusters []string) string {
	return strings.Join(clusters, "")
}

func SliceConvString2Int64(data []string) ([]int64, error) {
	if len(data) == 0 {
		return make([]int64, 0), nil
	}
	res := make([]int64, len(data))
	for i, d := range data {
		a, err := strconv.ParseInt(d, 10, 64)
		if err != nil {
			return nil, err
		}
		res[i] = a
	}
	return res, nil
}

const char = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandChar 随机生成指定长度的字符串
func RandChar(size int) string {
	rand.NewSource(time.Now().UnixNano()) // 产生随机种子
	var s bytes.Buffer
	for i := 0; i < size; i++ {
		s.WriteByte(char[rand.Int63()%int64(len(char))])
	}
	return s.String()
}

func CheckStringOrNilPointer(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		if len(v) > 0 {
			return v, true
		}
	case *string:
		if v != nil && len(*v) > 0 {
			return *v, true
		}
	default:
		break
	}

	return "", false
}

func Len(s string) int {
	return utf8.RuneCountInString(s)
}
