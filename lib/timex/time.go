package timex

import (
	"strconv"
	"time"
)

func NowStr() string {
	return time.Now().Format("20060102150405")
}

func NowStrMsec() string {
	return time.Now().Format("20060102150405.000")
}

func NowUnixStr() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func NowStrDay() string {
	return time.Now().Format("20060102")
}

func IsLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func FormatTimeToList(t time.Time) []int {
	hour, min, sec := t.Clock()
	year, month, day := t.Date()
	return []int{t.Nanosecond(), sec, min, hour, day, int(month), year}
}

func UnixTSFormat(ts int64) string {
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}

// 将时间戳根据采集间隔取整为当时间隔的0秒，比如16:34,按5分钟取整后是16:30
// time.Round()方法进行四舍五入
func Rounding(ts time.Time, interval time.Duration) time.Time {
	intv := int64(interval.Seconds())
	s := ts.Unix() / intv * intv

	// 如果interval大于1小时，处理时区偏移，固定按照东八区处理
	if interval > time.Hour {
		s = (ts.Unix()+8*60*60)/intv*intv - 8*60*60
	}

	return time.Unix(s, 0)
}
