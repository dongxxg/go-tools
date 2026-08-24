package timex

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRounding(t *testing.T) {
	cnLocation, _ := time.LoadLocation("Asia/Shanghai")
	tests := []struct {
		giveTime     time.Time
		giveInterval time.Duration
		want         string
	}{
		{
			giveTime:     time.Date(2023, 9, 22, 16, 31, 12, 0, time.UTC),
			giveInterval: 5 * time.Minute,
			want:         "2023-09-22 16:30:00 +0000 UTC",
		},
		{
			giveTime:     time.Date(2023, 9, 22, 16, 31, 12, 0, time.UTC),
			giveInterval: 1 * time.Minute,
			want:         "2023-09-22 16:31:00 +0000 UTC",
		},
		{
			giveTime:     time.Date(2023, 9, 22, 16, 31, 12, 0, time.UTC),
			giveInterval: 6 * time.Hour,
			want:         "2023-09-22 16:00:00 +0000 UTC",
		},
		{
			giveTime:     time.Date(2023, 9, 22, 16, 31, 12, 0, cnLocation),
			giveInterval: 6 * time.Hour,
			want:         "2023-09-22 04:00:00 +0000 UTC", // = 2023-09-22 12:00:00 +0800 CST
		},
		{
			giveTime:     time.Date(2023, 9, 22, 16, 31, 12, 0, time.UTC),
			giveInterval: 10 * time.Second,
			want:         "2023-09-22 16:31:10 +0000 UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.giveTime.String(), func(t *testing.T) {
			ts := Rounding(tt.giveTime, tt.giveInterval)
			assert.Equal(t, tt.want, ts.UTC().String())
		})
	}
}
