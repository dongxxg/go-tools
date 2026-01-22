package alarm

import (
	"reflect"
	"strings"
	"sync"
	"testing"
	"unitechs.com/cno/kit/db/kredis"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/template/model"
)

func Test_getParamVal(t *testing.T) {
	type args struct {
		expr string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "1",
			args: args{
				expr: "rpc/res/alarm/getOnuOpticalSplitter?oltdevid=${k_RESID}&ponport=${k_RESPARA}&onuid=${k_LOCATEINFO}&ostype=1",
			},
			want: []string{"1"},
		},
		{
			name: "2",
			args: args{
				expr: "rpc/res/alarm/getOnuOpticalSplitter?oltdevid=${k_RESID}&ponport=${k_RESPARA}&onuid=${k_LOCATEINFO}&ostype=",
			},
			want: []string{},
		},
		{
			name: "2",
			args: args{
				expr: "rpc/res/alarm/getOnuOpticalSplitter?name=wiken&oltdevid=${k_RESID}&ponport=${k_RESPARA}&onuid=${k_LOCATEINFO}&ostype=1",
			},
			want: []string{"wiken", "1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getParamVal(tt.args.expr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getParamVal() = %v, want %v", got, tt.want)
			}
		})
	}
}

type msg struct {
}

func (m msg) Get(k string) (interface{}, bool) {
	return "NA-0-1-0:8", true
}

func (m msg) GetCondKey() (key, val string, exist bool) {
	//TODO implement me
	panic("implement me")
}

func (m msg) GetStr(key string) string {
	//TODO implement me
	panic("implement me")
}

func TestTemplateConvertor_APIFunc(t1 *testing.T) {
	type fields struct {
		Option     TemplateConvertorOption
		cache      model.ICache
		SerialNo   ISerialNo
		ExpireHour int
		Zone       string
		writeLock  sync.RWMutex
	}
	type args struct {
		ctx  kctx.Ctx
		temp model.Elem
		msg  model.IMsg
		s    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "1",
			fields: fields{
				Option:     TemplateConvertorOption{},
				cache:      nil,
				SerialNo:   nil,
				ExpireHour: 0,
				Zone:       "",
				writeLock:  sync.RWMutex{},
			},
			args: args{
				ctx: kctx.NewDefaultContext(),
				temp: model.Elem{
					SrcField:  "aa",
					DstField:  "aa",
					AlarmType: "",
					MapType:   "api",
					MapExpr:   "http://localhost:123/name=${name}&age=${age}::name",
				},
				msg: nil,
				s:   "",
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &TemplateConvertor{
				Option:     tt.fields.Option,
				cache:      tt.fields.cache,
				SerialNo:   tt.fields.SerialNo,
				ExpireHour: tt.fields.ExpireHour,
				Zone:       tt.fields.Zone,
				writeLock:  tt.fields.writeLock,
			}
			got, err := t.APIFunc(tt.args.ctx, tt.args.temp, tt.args.msg, tt.args.s)
			if (err != nil) != tt.wantErr {
				t1.Errorf("APIFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("APIFunc() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateConvertor_CaseFunc(t1 *testing.T) {

	type args struct {
		ctx  kctx.Ctx
		temp model.Elem
		m    model.IMsg
		s    string
	}
	tests := []struct {
		name    string
		args    args
		wantVal interface{}
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				ctx: kctx.NewDefaultContext(),
				temp: model.Elem{
					SrcField:  "",
					DstField:  "ResPara",
					AlarmType: "case",
					MapType:   "",
					MapExpr:   "${RACK}=NA:1;${RACK}",
				},
				m: &MsgMap{
					m: map[string]interface{}{
						"RACK": "NA",
					},
				},
			},
			wantVal: "1",
			wantErr: false,
		},
		{
			name: "2",
			args: args{
				ctx: kctx.NewDefaultContext(),
				temp: model.Elem{
					SrcField:  "",
					DstField:  "ResPara",
					AlarmType: "case",
					MapType:   "",
					MapExpr:   "${RACK}=NA:1;${RACK}",
				},
				m: &MsgMap{
					m: map[string]interface{}{
						"RACK": "2",
					},
				},
			},
			wantVal: "2",
			wantErr: false,
		},
		{
			name: "3",
			args: args{
				ctx: kctx.NewDefaultContext(),
				temp: model.Elem{
					SrcField:  "",
					DstField:  "ResPara",
					AlarmType: "case",
					MapType:   "",
					MapExpr:   "${RACK}=NA:${name};${RACK}",
				},
				m: &MsgMap{
					m: map[string]interface{}{
						"RACK": "NA",
						"name": "wiken",
					},
				},
			},
			wantVal: "wiken",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &TemplateConvertor{}
			gotVal, err := t.CaseFunc(tt.args.ctx, tt.args.temp, tt.args.m, tt.args.s)
			if (err != nil) != tt.wantErr {
				t1.Errorf("CaseFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotVal, tt.wantVal) {
				t1.Errorf("CaseFunc() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestTemplateConvertor_RegFunc(t1 *testing.T) {
	type fields struct {
		Option     TemplateConvertorOption
		cache      model.ICache
		SerialNo   ISerialNo
		ExpireHour int
		Zone       string
		writeLock  sync.RWMutex
	}
	type args struct {
		ctx kctx.Ctx
		e   model.Elem
		m   model.IMsg
		s   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantVal interface{}
		wantErr bool
	}{
		{
			name:   "1",
			fields: fields{},
			args: args{
				ctx: kctx.NewDefaultContext(),
				e: model.Elem{
					SrcField:  "aa",
					DstField:  "bb",
					AlarmType: "",
					MapType:   "reg",
					MapExpr:   `\[Session\s+Create\s+But\s+Not\s+Up\]Ld:\d+\/Rd:\d+,Protocol:PIM,singleHop,IfName:(\S+),VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:|\[Session\s+Create\s+But\s+Not\s+Up\]Ld:\d+\/Rd:\d+,Protocol:INSTANCE,multiHop,VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:(\S+)\.|\[Session\s+Down\s+For\s+\S+\]Ld:\d+\/Rd:\d+,Protocol:PIM,singleHop,IfName:(\S+),VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:|\[Session\s+Down\s+For\s+\S+\]Ld:\d+\/Rd:\d+,Protocol:INSTANCE,multiHop,VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:(\S+)\.`,
				},
				m: &MsgMap{
					m: map[string]interface{}{
						"aa": "[Session Create But Not Up]Ld:16783/Rd:0,Protocol:PIM,singleHop,IfName:smartgroup78.100,VpnId:0,Local/Peer:192.168.241.37/192.168.240.253,SessionName:.",
					},
				},
				s: "",
			},
			wantVal: "smartgroup78.100",
			wantErr: false,
		},
		{
			name:   "2",
			fields: fields{},
			args: args{
				ctx: kctx.NewDefaultContext(),
				e: model.Elem{
					SrcField:  "aa",
					DstField:  "bb",
					AlarmType: "",
					MapType:   "reg",
					MapExpr:   `\[Session\s+Create\s+But\s+Not\s+Up\]Ld:\d+\/Rd:\d+,Protocol:PIM,singleHop,IfName:(\S+),VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:|\[Session\s+Create\s+But\s+Not\s+Up\]Ld:\d+\/Rd:\d+,Protocol:INSTANCE,multiHop,VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:(\S+)\.|\[Session\s+Down\s+For\s+\S+\]Ld:\d+\/Rd:\d+,Protocol:PIM,singleHop,IfName:(\S+),VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:|\[Session\s+Down\s+For\s+\S+\]Ld:\d+\/Rd:\d+,Protocol:INSTANCE,multiHop,VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:(\S+)\.`,
				},
				m: &MsgMap{
					m: map[string]interface{}{
						"aa": "[Session Down For FTM_UP_DOWN]Ld:16550/Rd:16555,Protocol:INSTANCE,multiHop,VpnId:0,Local/Peer:172.16.32.203/172.16.32.202,SessionName:bfd-vrrp91.",
					},
				},
				s: "",
			},
			wantVal: "bfd-vrrp91",
			wantErr: false,
		},
		{
			name:   "3",
			fields: fields{},
			args: args{
				ctx: kctx.NewDefaultContext(),
				e: model.Elem{
					SrcField:  "aa",
					DstField:  "bb",
					AlarmType: "",
					MapType:   "reg",
					MapExpr:   `\[Session\s+Create\s+But\s+Not\s+Up\]Ld:\d+\/Rd:\d+,Protocol:PIM,singleHop,IfName:(\S+),VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:|\[Session\s+Create\s+But\s+Not\s+Up\]Ld:\d+\/Rd:\d+,Protocol:INSTANCE,multiHop,VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:(\S+)\.|\[Session\s+Down\s+For\s+\S+\]Ld:\d+\/Rd:\d+,Protocol:PIM,singleHop,IfName:(\S+),VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:|\[Session\s+Down\s+For\s+\S+\]Ld:\d+\/Rd:\d+,Protocol:INSTANCE,multiHop,VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:(\S+)\.`,
				},
				m: &MsgMap{
					m: map[string]interface{}{
						"aa": "[Session Down For RCV_ADMINDOWN]Ld:16386/Rd:16468,Protocol:PIM,singleHop,IfName:smartgroup12.100,VpnId:0,Local/Peer:192.168.240.42/192.168.240.41,SessionName:.",
					},
				},
				s: "smartgroup12.100",
			},
			wantVal: "smartgroup12.100",
			wantErr: false,
		},
		{
			name:   "4",
			fields: fields{},
			args: args{
				ctx: kctx.NewDefaultContext(),
				e: model.Elem{
					SrcField:  "aa",
					DstField:  "bb",
					AlarmType: "",
					MapType:   "reg",
					MapExpr:   `\[Session\s+Create\s+But\s+Not\s+Up\]Ld:\d+\/Rd:\d+,Protocol:PIM,singleHop,IfName:(\S+),VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:|\[Session\s+Create\s+But\s+Not\s+Up\]Ld:\d+\/Rd:\d+,Protocol:INSTANCE,multiHop,VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:(\S+)\.|\[Session\s+Down\s+For\s+\S+\]Ld:\d+\/Rd:\d+,Protocol:PIM,singleHop,IfName:(\S+),VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:|\[Session\s+Down\s+For\s+\S+\]Ld:\d+\/Rd:\d+,Protocol:INSTANCE,multiHop,VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:(\S+)\.`,
				},
				m: &MsgMap{
					m: map[string]interface{}{
						"aa": "[Session Down For FTM_UP_DOWN]Ld:19358/Rd:19257,Protocol:PIM,singleHop,IfName:smartgroup36.100,VpnId:0,Local/Peer:192.168.240.129/192.168.240.130,SessionName:.",
					},
				},
				s: "",
			},
			wantVal: "smartgroup36.100",
			wantErr: false,
		},
		{
			name:   "5",
			fields: fields{},
			args: args{
				ctx: kctx.NewDefaultContext(),
				e: model.Elem{
					SrcField:  "aa",
					DstField:  "bb",
					AlarmType: "",
					MapType:   "reg",
					MapExpr:   `\[Session\s+Create\s+But\s+Not\s+Up\]Ld:\d+\/Rd:\d+,Protocol:PIM,singleHop,IfName:(\S+),VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:|\[Session\s+Create\s+But\s+Not\s+Up\]Ld:\d+\/Rd:\d+,Protocol:INSTANCE,multiHop,VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:(\S+)\.|\[Session\s+Down\s+For\s+\S+\]Ld:\d+\/Rd:\d+,Protocol:PIM,singleHop,IfName:(\S+),VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:|\[Session\s+Down\s+For\s+\S+\]Ld:\d+\/Rd:\d+,Protocol:INSTANCE,multiHop,VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:(\S+)\.`,
				},
				m: &MsgMap{
					m: map[string]interface{}{
						"aa": "[Session Create But Not Up]Ld:16783/Rd:0,Protocol:PIM,singleHop,IfName:smartgroup78.100,VpnId:0,Local/Peer:192.168.241.37/192.168.240.253,SessionName:.",
					},
				},
				s: "",
			},
			wantVal: "smartgroup78.100",
			wantErr: false,
		},
		{
			name:   "6",
			fields: fields{},
			args: args{
				ctx: kctx.NewDefaultContext(),
				e: model.Elem{
					SrcField:  "aa",
					DstField:  "bb",
					AlarmType: "",
					MapType:   "reg",
					MapExpr:   `\[Session\s+Create\s+But\s+Not\s+Up\]Ld:\d+\/Rd:\d+,Protocol:PIM,singleHop,IfName:(\S+),VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:|\[Session\s+Create\s+But\s+Not\s+Up\]Ld:\d+\/Rd:\d+,Protocol:INSTANCE,multiHop,VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:(\S+)\.|\[Session\s+Down\s+For\s+\S+\]Ld:\d+\/Rd:\d+,Protocol:PIM,singleHop,IfName:(\S+),VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:|\[Session\s+Down\s+For\s+\S+\]Ld:\d+\/Rd:\d+,Protocol:INSTANCE,multiHop,VpnId:\d+,Local\/Peer:\S+\/\S+,SessionName:(\S+)\.`,
				},
				m: &MsgMap{
					m: map[string]interface{}{
						"aa": "[Session Create But Not Up]Ld:16398/Rd:0,Protocol:INSTANCE,multiHop,VpnId:0,Local/Peer:192.168.200.99/192.168.200.100,SessionName:bfd-vrrp15.",
					},
				},
				s: "bfd-vrrp15",
			},
			wantVal: "bfd-vrrp15",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &TemplateConvertor{
				Option:     tt.fields.Option,
				cache:      tt.fields.cache,
				SerialNo:   tt.fields.SerialNo,
				ExpireHour: tt.fields.ExpireHour,
				Zone:       tt.fields.Zone,
				writeLock:  tt.fields.writeLock,
			}
			gotVal, err := t.RegFunc(tt.args.ctx, tt.args.e, tt.args.m, tt.args.s)
			if (err != nil) != tt.wantErr {
				t1.Errorf("RegFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotVal, tt.wantVal) {
				t1.Errorf("RegFunc() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}

func TestTemplateConvertor_NvlFunc(t1 *testing.T) {
	type fields struct {
		Option     TemplateConvertorOption
		cache      model.ICache
		SerialNo   ISerialNo
		ExpireHour int
		Zone       string
		writeLock  sync.RWMutex
	}
	type args struct {
		ctx  kctx.Ctx
		temp model.Elem
		msg  model.IMsg
		s    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name:   "1",
			fields: fields{},
			args: args{
				ctx: kctx.NewDefaultContext(),
				temp: model.Elem{
					SrcField:  "aa",
					DstField:  "bb",
					AlarmType: "ALL",
					MapType:   "nvl",
					MapExpr:   `aa,${aa},${bb}`,
				},
				msg: &MsgMap{
					m: map[string]interface{}{
						"aa": "aaValue",
						"bb": "bbValue",
					},
				},
				s: "",
			},
			want:    "aaValue",
			wantErr: false,
		},
		{
			name:   "1",
			fields: fields{},
			args: args{
				ctx: kctx.NewDefaultContext(),
				temp: model.Elem{
					SrcField:  "aa",
					DstField:  "bb",
					AlarmType: "ALL",
					MapType:   "nvl",
					MapExpr:   `ac,${aa},${bb}`,
				},
				msg: &MsgMap{
					m: map[string]interface{}{
						"aa": "aaValue",
						"bb": "bbValue",
					},
				},
				s: "",
			},
			want:    "bbValue",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &TemplateConvertor{
				Option:     tt.fields.Option,
				cache:      tt.fields.cache,
				SerialNo:   tt.fields.SerialNo,
				ExpireHour: tt.fields.ExpireHour,
				Zone:       tt.fields.Zone,
				writeLock:  tt.fields.writeLock,
			}
			got, err := t.NvlFunc(tt.args.ctx, tt.args.temp, tt.args.msg, tt.args.s)
			if (err != nil) != tt.wantErr {
				t1.Errorf("NvlFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("NvlFunc() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateConvertor_ApiCache(t1 *testing.T) {
	ctx := kctx.NewDefaultContext()
	cli, _ := kredis.NewRedisRepo(ctx, "node", "127.0.0.1:6379", "wiken", 0)
	if err := cli.Set("bb", ""); err != nil {
		t1.Fatal(err)
	}
	convertor := TemplateConvertor{
		cache: cli,
		Option: TemplateConvertorOption{
			NotFoundPolicy: "",
			ApiCache: ApiCacheOption{
				CacheKey: func(varLst []string) string {
					return strings.Join(varLst, ":")
				},
				ExpireHour: 0,
			},
			SerialNo: SerialNoOption{},
		},
	}

	type args struct {
		ctx  kctx.Ctx
		temp model.Elem
		msg  model.IMsg
		s    string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				ctx: kctx.NewDefaultContext(),
				temp: model.Elem{
					SrcField:  "",
					DstField:  "",
					AlarmType: "",
					MapType:   "apiCache",
					MapExpr:   "ttp://RESOURCE-JAVA-SERVICE/api/res/alarm/getRemoteInfo?localDevId=${k_RESID}&peerIp=${k_RESPARA}&osType=1::data.remote",
				},
				msg: &MsgMap{
					m: map[string]interface{}{
						"k_RESID":   "aaaaa",
						"k_RESPARA": "bbbbbb",
					}},
				s: "",
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "2",
			args: args{
				ctx: kctx.NewDefaultContext(),
				temp: model.Elem{
					SrcField:  "",
					DstField:  "",
					AlarmType: "",
					MapType:   "apiCache",
					MapExpr:   "http://RESOURCE-JAVA-SERVICE/api/res/alarm/getRemoteInfo?localDevId=${k_RESID}&peerIp=${k_RESPARA}::data.remote",
				},
				msg: &MsgMap{
					m: map[string]interface{}{
						"k_RESID":   "aaaaa",
						"k_RESPARA": "bbbbbb",
					}},
				s: "",
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			got, err := convertor.ApiCache(tt.args.ctx, tt.args.temp, tt.args.msg, tt.args.s)
			if (err != nil) != tt.wantErr {
				t1.Errorf("ApiCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("ApiCache() got = %v, want %v", got, tt.want)
			}
		})
	}
}
