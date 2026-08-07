package timex

import (
	"fmt"
	"time"
)

// Use the long enough past time as start time, in case timex.Now() - lastTime equals 0.
var initTime = time.Now().AddDate(-1, -1, -1)

// Now returns a relative time duration since initTime, which is not important.
// The caller only needs to care about the relative value.
func Now() time.Duration {
	return time.Since(initTime)
}

// Since returns a diff since given d.
func Since(d time.Duration) time.Duration {
	return time.Since(initTime) - d
}

func Time() time.Time {
	return initTime.Add(Now())
}

func TimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)
	hours := diff.Hours()
	if hours < 1.0 {
		return fmt.Sprintf("约 %.0f 分钟前", diff.Minutes())
	}

	if hours < 24.0 {
		return fmt.Sprintf("约 %.0f 小时前", hours)
	}

	if hours < 72.0 {
		return fmt.Sprintf("约 %.0f 天前", hours/24.0)
	}

	// 同一年，不用年份
	if now.Year() == t.Year() {
		return t.Format("01-02 15:04")
	}

	return t.Format("2006-01-02")
}

func FormatTimedelta(t time.Time) string {
	d := time.Now().Unix() - t.Unix()
	u := uint64(d)
	neg := d < 0
	if neg {
		u = -u
	}
	secs := u % 60
	u /= 60
	mins := u % 60
	u /= 60
	hours := u % 24
	days := u / 24

	if days == 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, mins, secs)
	}
	return fmt.Sprintf("%dd ", days) + fmt.Sprintf("%02d:%02d:%02d", hours, mins, secs)
}

func Yesterday(current time.Time) time.Time {
	yesterday := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
	return yesterday
}

func SixMonthAgo(current time.Time) time.Time {
	currentYear := current.Year()
	currentMonth := current.Month()
	for i := 0; i < 6; i++ {
		if currentMonth == time.January {
			currentMonth = time.December
			currentYear--
		} else {
			currentMonth--
		}
	}
	sixMonthsAgo := time.Date(currentYear, currentMonth, current.Day(), 0, 0, 0, 0, time.UTC)
	return sixMonthsAgo
}

func Today(current time.Time) time.Time {
	today := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, time.UTC)
	return today
}

func NextMonth(today time.Time) time.Time {
	currentYear := today.Year()
	currentMonth := today.Month()
	if currentMonth == time.December {
		currentMonth = time.January
		currentYear++
	} else {
		currentMonth++
	}
	nextMonth := time.Date(currentYear, currentMonth, today.Day(), 0, 0, 0, 0, time.UTC)
	return nextMonth
}

func LastMonth(today time.Time) time.Time {
	currentYear := today.Year()
	currentMonth := today.Month()
	if currentMonth == time.January {
		currentMonth = time.December
		currentYear--
	} else {
		currentMonth--
	}
	lastMonth := time.Date(currentYear, currentMonth, today.Day(), 0, 0, 0, 0, time.UTC)
	return lastMonth
}

// FirstDayOfTheWeek gets the first day (Monday) of the current week.
func FirstDayOfTheWeek(today time.Time) time.Time {
	dayDiff := int(today.Weekday())
	if today.Weekday() == time.Sunday {
		dayDiff += 7
	}
	dayDiff = -dayDiff + 1
	targetDay := today.AddDate(0, 0, dayDiff)
	FirstDayOfTheWeek := time.Date(targetDay.Year(), targetDay.Month(), targetDay.Day(), 0, 0, 0, 0, time.UTC)
	return FirstDayOfTheWeek
}

// GetTheFirstDayOfMonth return the first day of the giving month.
func GetTheFirstDayOfMonth(today time.Time) time.Time {
	firstDayOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	return firstDayOfMonth
}

// GetLastWeek return the dateTime from last week.
func GetLastWeek(today time.Time) time.Time {
	lastWeek := today.AddDate(0, 0, -6)
	lastWeek = time.Date(lastWeek.Year(), lastWeek.Month(), lastWeek.Day(), 0, 0, 0, 0, time.UTC)
	return lastWeek
}
