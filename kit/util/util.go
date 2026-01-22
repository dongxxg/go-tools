package util

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"net"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"unitechs.com/cno/kit/kctx"
)

// RequestId context 中存储的key值
const RequestId = "requestId"

func RidStr(ctx context.Context) string {
	if s, ok := ctx.Value(RequestId).(string); ok {
		return "rid: " + s + " "
	}

	return "rid: "
}
func Rid(ctx context.Context) string {
	if s, ok := ctx.Value(RequestId).(string); ok {
		return s
	}

	return ""
}

func NewContextWith(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, RequestId, rid)
}
func NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, RequestId, NewUUIDV4())
}
func NewUUIDV4() string {
	return uuid.NewV4().String()
}

func NewUUIDStr() string {
	return strings.ReplaceAll(uuid.NewV4().String(), "-", "")
}

// ParseTimeByLen 通过长度解析时间字符串
func ParseTimeByLen(t string) (time.Time, error) {
	switch len(t) {
	case 8:
		// YYYYMMDD
		return time.Parse(TimeFormatDay, t)
	case 10:
		// YYYYMMDDHH
		return time.Parse(TimeFormatHour, t)
	case 12:
		// YYYYMMDDHHMM
		return time.Parse(TimeFormatMin, t)
	case 14:
		// YYYYMMDDHHMMSS
		return time.Parse(TimeFormatSec, t)
	case 19:
		return time.Parse(TimeFormatDefault, t)
	default:
		return time.Time{}, fmt.Errorf("invalidate time formate: %s", t)
	}
}

// ParseTimeByLenWithZone 通过长度解析时间字符串
func ParseTimeByLenWithZone(t string, z *time.Location) (time.Time, error) {
	switch len(t) {
	case 8:
		// YYYYMMDD
		return time.ParseInLocation(TimeFormatDay, t, z)
	case 10:
		// YYYYMMDDHH
		return time.ParseInLocation(TimeFormatHour, t, z)
	case 12:
		// YYYYMMDDHHMM
		return time.ParseInLocation(TimeFormatMin, t, z)
	case 14:
		// YYYYMMDDHHMMSS
		return time.ParseInLocation(TimeFormatSec, t, z)
	case 19:
		return time.ParseInLocation(TimeFormatDefault, t, z)
	default:
		return time.Time{}, fmt.Errorf("invalidate time formate: %s", t)
	}
}

var TimeZone = time.UTC

func ParseInterfaceToTimeWith(t interface{}, zone string) (time.Time, error) {
	loc, err := time.LoadLocation(zone)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "parse location failed")
	}
	switch tt := t.(type) {
	case time.Time:
		return tt, nil
	default:
		ss := ConvertToString(t)
		return ParseToTimeWithZone(ss, loc)
	}
}

func ParseToTimeWithZone(s string, z *time.Location) (t time.Time, err error) {
	formatList := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		time.RFC3339,
		time.RFC3339Nano,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.Kitchen,
		time.StampMilli,
		time.Stamp,
		time.StampMicro,
		time.StampNano,
		time.Layout,
		"2006-01-02 15",
		"2006-01-02",
	}

	for _, v := range formatList {
		t, err = time.ParseInLocation(v, s, z)
		if err != nil {
			continue
		}
		return
	}
	// 解析时间戳
	n, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return time.Unix(n, 0), nil
	}

	return ParseTimeByLen(s)
}

// ParseToTime 解析得到时间
func ParseToTime(s string) (t time.Time, err error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("timeStr is empty")
	}
	formatList := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04",
		"2006-01-02 15",
		"2006-01-02",
	}

	for _, v := range formatList {
		t, err = time.Parse(v, s)
		if err != nil {
			continue
		}
		return
	}
	// 解析时间戳
	n, err := strconv.ParseInt(s, 10, 64)

	if err == nil {
		return time.Unix(n, 0), nil
	}

	return ParseTimeByLen(s)
}

func StringToSliceBySep(endPoint, sep string) []string {
	l := strings.Split(endPoint, sep)
	set := gset.NewStrSetFrom(l)
	set.Remove("")
	set.Remove(" ")
	return set.Slice()
}

func MarshalOutString(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func UnMarshalWithInterface(in interface{}, out interface{}) error {
	b, err := json.Marshal(in)
	if err != nil {
		return errors.Wrap(err, "UnMarshalWithInterface marshall failed")
	}
	if err := json.Unmarshal(b, out); err != nil {
		return errors.Wrap(err, "UnMarshalWithInterface unmarshal failed")
	}

	return nil
}

func HandlePanic(ctx kctx.Ctx, logLevel string, f func()) {
	if err := recover(); err != nil {
		f()
		switch logLevel {
		case "warning":
			ctx.Warningf("handlePanic err: %v \n%s", err, string(debug.Stack()))
		case "fatal":
			ctx.Fatalf("handlePanic err: %v \n%s", err, string(debug.Stack()))
		default:
			ctx.Errorf("handlePanic err: %v \n%s", err, string(debug.Stack()))
		}
	}
}

// GetVariable 获取变量
func GetVariable(content string) []string {
	r := regexp.MustCompile(`\$\{(\S+?)\}`)
	lstLst := r.FindAllStringSubmatch(content, -1)
	out := gset.NewStrSet()
	for _, lst := range lstLst {
		if len(lst) != 2 {
			continue
		}
		out.Add(lst[1])
	}

	return out.Slice()
}

func GetLocalIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip := GetIpFromAddr(addr)
			if ip == nil {
				continue
			}
			return ip, nil
		}
	}
	return nil, errors.New("connected to the network?")
}
func GetIpFromAddr(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() {
		return nil
	}
	ip = ip.To4()
	if ip == nil {
		return ip.To16()
	}
	return ip
}
