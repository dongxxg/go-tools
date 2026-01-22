// 任务号    日期     开发者      描述

package util

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"net"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestParseTimeByLen(t *testing.T) {
	a := time.Now().Unix()
	b := time.Now().UnixNano()
	c := time.Now().UnixMilli()
	fmt.Printf("%d, %d, %d", len(strconv.Itoa(int(a))),
		len(strconv.Itoa(int(b))), len(strconv.Itoa(int(c))))
	aa := time.Unix(a, 0)
	bb := time.Unix(b, 0)
	cc := time.Unix(c, 0)
	fmt.Println("here, ", aa, bb, cc)
}

type person struct {
	Name    string `json:"name"`
	Friends string `json:"friends"`
}

func TestName(t *testing.T) {

	mm := person{
		Name:    "wiken",
		Friends: "",
	}
	liker := person{
		Name:    "liker",
		Friends: "",
	}
	b, _ := json.Marshal(liker)
	mm.Friends = string(b)
	wiken, _ := json.Marshal(mm)
	r := gjson.GetBytes(wiken, "friends.name")
	name := r.String()
	fmt.Println(name)

}

func TestGetLocalIP(t *testing.T) {
	tests := []struct {
		name    string
		want    net.IP
		wantErr bool
	}{
		{
			name:    "1",
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLocalIP()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLocalIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLocalIP() got = %v, want %v", got, tt.want)
			}
		})
	}
}
