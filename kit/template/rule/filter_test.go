package rule

import (
	"context"
	"fmt"
	"testing"
	"time"
	"unitechs.com/cno/kit/kctx"
)

func TestIsMatch(t *testing.T) {
	type args struct {
		exp  string
		data map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "demo01",
			args: args{
				exp: "K_SUMALARMID==1",
				data: map[string]interface{}{
					"K_SUMALARMID": "1",
				},
			}, want: true,
			wantErr: false,
		},
		{
			name: "demo01",
			args: args{
				exp: "K_SUMALARMID==1",
				data: map[string]interface{}{
					"K_SUMALARMID": "2",
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "demo01",
			args: args{
				exp:  "K_SUMALARMID==1",
				data: map[string]interface{}{},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "demo02",
			args: args{
				exp: "K_SUMALARMID==1&&K_SERIALNO==2 ",
				data: map[string]interface{}{
					"K_SUMALARMID": "1",
					"K_SERIALNO":   "2",
				},
			}, want: true,
			wantErr: false,
		},
		{
			name: "demo03",
			args: args{
				exp: "K_SUMALARMID==1&&K_SERIALNO=='2'",
				data: map[string]interface{}{
					"K_SUMALARMID": "1",
					"K_SERIALNO":   "2",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "demo04",
			args: args{
				exp: "K_SUMALARMID==1&&K_SERIALNO=='2'",
				data: map[string]interface{}{
					"K_SUMALARMID": 1,
					"K_SERIALNO":   2,
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "demo05",
			args: args{
				exp: "K_SUMALARMID==1&&K_SERIALNO=='2'",
				data: map[string]interface{}{
					"K_SUMALARMID": "1",
					"K_SERIALNO":   "2",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "demo07",
			args: args{

				exp: "(K_SUMALARMID==1&&K_SERIALNO=='2')&& MsgType in ['add', 'delete', 'mod'] ",
				data: map[string]interface{}{
					"K_SUMALARMID": "1",
					"K_SERIALNO":   "2",
					"MsgType":      "add",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "demo08",
			args: args{
				exp: "(K_SUMALARMID==1&&K_SERIALNO=='2')&& MsgType in ['add', 'delete', 'mod'] ",
				data: map[string]interface{}{
					"K_SUMALARMID": "1",
					"K_SERIALNO":   "2",
					"MsgType":      "add2",
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "test regexp true",
			args: args{
				exp: "AlarmLevelOrig matches '.*fo' ",
				data: map[string]interface{}{
					"K_SUMALARMID": "1",
					"K_SERIALNO":   "2",
					"MsgType":      "add", // 如果没有该字段， 返回是true
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "test regexp true",
			args: args{
				exp: "AlarmLevelOrig matches '.*fo' ",
				data: map[string]interface{}{
					"K_SUMALARMID":   "1",
					"K_SERIALNO":     "2",
					"MsgType":        "add", // 如果没有该字段， 返回是true
					"AlarmLevelOrig": "wiken",
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "test regexp true",
			args: args{
				exp: "AlarmLevelOrig matches '.*fo' ",
				data: map[string]interface{}{
					"K_SUMALARMID":   "1",
					"K_SERIALNO":     "2",
					"MsgType":        "add", // 如果没有该字段， 返回是true
					"AlarmLevelOrig": "wikenfo",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "test_regexp_false",
			args: args{
				exp: "(K_SUMALARMID==1&&K_SERIALNO=='2')&& MsgType in ['add', 'delete', 'mod'] && AlarmLevelOrig matches '.*afo' ",
				data: map[string]interface{}{
					"K_SUMALARMID":   "1",
					"K_SERIALNO":     "2",
					"MsgType":        "add",
					"AlarmLevelOrig": "wikenafo",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "test_or_false",
			args: args{
				exp: "(K_SUMALARMID==1&&K_SERIALNO=='2')&& MsgType in ['add', 'delete', 'mod'] && AlarmLevelOrig matches '.*afo'&& L_TIMES!=4",
				data: map[string]interface{}{
					"K_SUMALARMID": "1",
					"K_SERIALNO":   "2",
					"MsgType":      "add",
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "not_equal_true",
			args: args{
				exp: "K_SERIALNO!='2'",
				data: map[string]interface{}{
					"K_SUMALARMID": "1",
					"K_SERIALNO":   "3",
					"MsgType":      "add",
				},
			},
			want:    true,
			wantErr: false,
		},

		{
			name: "not_equal_true",
			args: args{
				exp: "K_SERIALNO!=2",
				data: map[string]interface{}{
					"K_SUMALARMID": "1",
					"K_SERIALNO":   "1",
					"MsgType":      "add",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "not_equal_true",
			args: args{
				exp: "K_SERIALNO!='2'",
				data: map[string]interface{}{
					"K_SUMALARMID": "1",
					"K_SERIALNO":   "1",
					"MsgType":      "add",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "time range",
			args: args{
				exp: "t_STARTTIME>'2020-08-10 13:00:00'",
				data: map[string]interface{}{
					"t_STARTTIME": time.Now(),
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "demo 10",
			args: args{
				exp: "l_ALARMSTATIS== '1' && k_ENGIID matches '^$'",
				data: map[string]interface{}{
					"t_STARTTIME":   time.Now(),
					"k_ENGIID":      "", // 匹配空字符串
					"l_ALARMSTATIS": "1",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "demo 11",
			args: args{
				exp: "l_ALARMSTATIS== '1' && k_ENGIID matches '^$'",
				data: map[string]interface{}{
					"t_STARTTIME":   time.Now(),
					"k_ENGIID":      "", // 匹配空字符串
					"l_ALARMSTATIS": "1",
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "demo 12",
			args: args{
				exp: "k_ENGIID matches ''",
				data: map[string]interface{}{
					"k_ENGIID": "kehu", // 匹配空字符串
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "demo 13",
			args: args{
				exp: "k_ENGIID in  ['kehu', 'b', 'c']",
				data: map[string]interface{}{
					"k_ENGIID": "kehu", // 匹配空字符串
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "demo 14",
			args: args{
				exp: "k_ENGIID in  ['kehu', 'b', 'c']",
				data: map[string]interface{}{
					"k_ENGIID": "kehua", // 匹配空字符串
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "demo 15",
			args: args{
				exp: "k_ENGIID not in  ['kehu', 'b', 'c']",
				data: map[string]interface{}{
					"k_ENGIID": "asdfasdf", // 匹配空字符串
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "demo 16",
			args: args{
				exp: "k_ENGIID not in  ['kehu', '阿斯蒂芬', 'c']",
				data: map[string]interface{}{
					"k_ENGIID": "阿斯蒂芬", // 匹配空字符串
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "demo 17",
			args: args{
				exp: "k_ENGIID not in  ['kehu', '阿斯蒂芬', 'c']",
				data: map[string]interface{}{
					"k_ENGIID": "苛阿斯蒂芬", // 匹配空字符串
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "demo 17",
			args: args{
				exp: "k_ENGIID include  '斯柯达'",
				data: map[string]interface{}{
					"k_ENGIID": "苛阿斯斯柯达蒂芬", // 匹配空字符串
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "demo 18",
			args: args{
				exp: "k_ENGIID exclude  '斯柯达'",
				data: map[string]interface{}{
					"k_ENGIID": "苛阿柯达蒂芬", // 匹配空字符串
				},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := kctx.NewCtxWithField2(context.TODO(), "test_trace_id")
			msg := NewFilterMsg(tt.args.data)
			got, err := IsMatch(ctx, tt.args.exp, msg, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsMatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsMatch() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDemo(t *testing.T) {
	var aa interface{}
	aa = time.Now()
	switch d := aa.(type) {
	case time.Time:
		fmt.Println(d)
	case string:

	}

}
