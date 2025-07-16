package media

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestNewInfoPlist(t *testing.T) {
	type args struct {
		duration int64
	}
	tests := []struct {
		name string
		args args
		want *InfoPlist
	}{
		// TODO: Add test cases.
		{name: "test", args: args{duration: 1000}, want: NewInfoPlist(1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewInfoPlist(tt.args.duration)
			infoByte, _ := json.Marshal(got)
			fmt.Println(string(infoByte))
		})
	}
}
