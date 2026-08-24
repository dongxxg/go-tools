package media

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// #EXT-X-PROGRAM-DATE-TIME:2023-09-14T01:35:18.601+00:00
func ParseExtXProgramDateTime(line string) (startTime int64, err error) {
	if !strings.HasPrefix(line, "#EXT-X-PROGRAM-DATE-TIME") {
		return 0, errors.New("format error ,should contain #EXT-X-PROGRAM-DATE-TIME")
	}
	if !strings.HasSuffix(line, "+00:00") {
		return 0, errors.New("only support +00:00 time zone,but got " + line)
	}
	//2023-09-14T01:35:18.601+00:00
	timeStr := strings.TrimPrefix(line, "#EXT-X-PROGRAM-DATE-TIME:")
	year, _ := strconv.Atoi(timeStr[0:4])
	month, _ := strconv.Atoi(timeStr[5:7])
	day, _ := strconv.Atoi(timeStr[8:10])
	hour, _ := strconv.Atoi(timeStr[11:13])
	min, _ := strconv.Atoi(timeStr[14:16])
	sec, _ := strconv.Atoi(timeStr[17:19])
	ms, _ := strconv.Atoi(timeStr[20:23])
	parseTime := time.Date(year, time.Month(month), day, hour, min, sec, ms*1e6, time.UTC)
	return parseTime.UnixNano() / 1e6, nil
}
