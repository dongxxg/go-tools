package support

import (
	"fmt"
	"time"
)

func DayBegin() time.Time {
	end := time.Now()
	_, offset := end.Zone()
	s := end.Unix()
	s = s - (s+int64(offset))%(24*60*60)
	return time.Unix(s, 0)
}

// NowMs 当前时间戳（毫秒）
func NowMs() int64 {
	return time.Now().UnixNano() / 1e6
}

// NowS 当前时间戳（秒）
func NowS() int64 {
	return time.Now().Unix()
}

type DATE_FORMAT string

const DATE_FORMAT1 DATE_FORMAT = "2006-01-02 15:04:05"
const DATE_FORMAT_DEFAULT DATE_FORMAT = "2006-01-02 15:04:05"
const DATE_FORMAT_MONTH_DAY DATE_FORMAT = "01-02" //月日

// 弃用，因为没有支持使用format
func FormatDate(timestamp int64, format DATE_FORMAT) string {
	if len(format) == 0 {
		format = DATE_FORMAT_DEFAULT
	}
	t := time.Unix(timestamp/1000, timestamp%1000)

	return t.Format(string(format))
}

// 把 当前时间戳（毫秒）转为指定格式
func FormatDateTime(t time.Time, format DATE_FORMAT) string {
	if len(format) == 0 {
		format = DATE_FORMAT_DEFAULT
	}
	return t.Format(string(format))
}

// 将毫秒转化成时分秒的形式
func FormatDuration(milliseconds int64) string {
	duration := time.Duration(milliseconds) * time.Millisecond
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// 将2023.8.1转成毫秒时间戳
func ParseDate(date string) int64 {
	t, _ := time.ParseInLocation("2006.1.2", date, time.Local)
	return t.UnixNano() / 1e6
}

// 截取24小时时间戳针对时间对象
func TruncToTimeObj(t time.Time) int64 {
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return t.UnixMilli()
}

// 截取24小时时间戳针对时间戳
func TruncToDate(timestamp int64) int64 {
	t := time.UnixMilli(timestamp)
	return TruncToTimeObj(t)
}

// 获取截取的24小时的当前时间对象
func GetTruncToday() int64 {
	t := time.Now()
	return TruncToTimeObj(t)
}

// 获取昨天最后一秒
func GetYesterdayEndtime() int64 {
	t := time.Now()
	return TruncToTimeObj(t) - 1000
}

func ParseTimeStr(dateTimeStr string, locationStr string) (time.Time, error) {
	location, _ := time.LoadLocation(locationStr)
	t, e := time.ParseInLocation("2006-01-02 15:04:05", dateTimeStr, location)
	return t, e
}

func AfterXMinutes(x int64) int64 {
	return NowMs() + x*60*1000
}

func BeforeXMinutes(x int64) int64 {
	return NowMs() - x*60*1000
}
