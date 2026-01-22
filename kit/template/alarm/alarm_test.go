// // 任务号    日期     开发者      描述
package alarm

//
//import (
//	"context"
//	"testing"
//	"unitechs.com/cno/kit/kctx"
//	"unitechs.com/cno/kit/template/model"
//)
//
//func TestMatchAndConvert(t *testing.T) {
//	ctx := kctx.NewCtxWithField2(context.TODO(), "demo_test")
//	condKey := "k_ALARMTYPEID"
//	//SNo := &SerialNoDemo{n: 1000}
//	type args struct {
//		ctx kctx.Ctx
//		m   model.IMsg
//		r   model.ConvertRule
//		arg MatchArg
//	}
//	tests := []struct {
//		name    string
//		args    args
//		want    string
//		wantErr bool
//	}{
//		//
//		{
//			name: "demo01_fixed",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID": "DC0001",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "",
//							DstField:  "k_fixed",
//							AlarmType: "ALL",
//							MapType:   "fixed",
//							MapExpr:   "test_fixes",
//						},
//					},
//				},
//				arg: MatchArg{},
//			},
//			want: `{"k_fixed":"test_fixes"}`,
//		},
//		{
//			name: "demo02_reg",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID": "DC0001",
//					"k_RAWMSG":      "aaaaaaaasdfasdfwikenlucy",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "k_RAWMSG",
//							DstField:  "reg_test",
//							AlarmType: "ALL",
//							MapType:   "reg",
//							MapExpr:   "wiken(.{4})",
//						},
//					},
//				},
//				arg: MatchArg{},
//			},
//			want: `{"reg_test":"lucy"}`,
//		},
//		{
//			name: "demo03_case_string",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID": "DC0001",
//					"k_RAWMSG":      "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS": "1",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "",
//							DstField:  "status",
//							AlarmType: "ALL",
//							MapType:   "case",
//							MapExpr:   "k_ALARMSTATUS='1':活跃;k_ALARMSTATUS='0':清除;未知",
//						},
//					},
//				},
//				arg: MatchArg{},
//			},
//			want: `{"status":"活跃"}`,
//		},
//		{
//			name: "demo03_case_int", // 所有的返回都是string
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID": "DC0001",
//					"k_RAWMSG":      "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS": "1",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "",
//							DstField:  "status",
//							AlarmType: "ALL",
//							MapType:   "case",
//							MapExpr:   "k_ALARMSTATUS='1':1;k_ALARMSTATUS='0':0;2",
//						},
//					},
//				},
//				arg: MatchArg{},
//			},
//			want: `{"status":"1"}`,
//		},
//		{
//			name: "demo03_case_2", // 所有的返回都是string
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID": "DC0001",
//					"k_RAWMSG":      "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS": "0",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "",
//							DstField:  "status",
//							AlarmType: "ALL",
//							MapType:   "case",
//							MapExpr:   "k_ALARMSTATUS='1':1;k_ALARMSTATUS='0':0;2",
//						},
//					},
//				},
//				arg: MatchArg{},
//			},
//			want: `{"status":"0"}`,
//		},
//		{
//			name: "demo03_case_3", // 所有的返回都是string
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID": "DC0001",
//					"k_RAWMSG":      "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS": "asd",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "",
//							DstField:  "status",
//							AlarmType: "ALL",
//							MapType:   "case",
//							MapExpr:   "k_ALARMSTATUS='1':1;k_ALARMSTATUS='0':0;2",
//						},
//					},
//				},
//				arg: MatchArg{},
//			},
//			want: `{"status":"2"}`,
//		},
//		{
//			name: "demo04_join",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID": "DC0001",
//					"k_RAWMSG":      "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS": "1",
//					"k_TITTLE":      "测试",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "",
//							DstField:  "desc",
//							AlarmType: "ALL",
//							MapType:   "join",
//							MapExpr:   "${k_ALARMTYPEID}::${k_TITTLE}",
//						},
//					},
//				},
//				arg: MatchArg{},
//			},
//			want: `{"desc":"DC0001::测试"}`,
//		},
//		{
//			name: "demo05_nul",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID": "DC0001",
//					"k_RAWMSG":      "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS": "1",
//					"k_TITTLE":      "aa",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "",
//							DstField:  "title",
//							AlarmType: "ALL",
//							MapType:   "nvl",
//							MapExpr:   "k_TITTLE,不为空,为空",
//						},
//					},
//				},
//				arg: MatchArg{},
//			},
//			want: `{"title":"不为空"}`,
//		},
//		{
//			name: "demo05_nul2",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID": "DC0001",
//					"k_RAWMSG":      "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS": "1",
//					"k_TITTLE":      "",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "",
//							DstField:  "title",
//							AlarmType: "ALL",
//							MapType:   "nvl",
//							MapExpr:   "k_TITTLE,不为空,为空",
//						},
//					},
//				},
//				arg: MatchArg{},
//			},
//			want: `{"title":"为空"}`,
//		},
//		{
//			name: "demo06_serialNo",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID": "DC0001",
//					"k_RAWMSG":      "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS": "1",
//					"k_TITTLE":      "",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "",
//							DstField:  "serialNo",
//							AlarmType: "ALL",
//							MapType:   "serial",
//							MapExpr:   "1000,1",
//						},
//					},
//				},
//				arg: MatchArg{
//					RuleId:     "11",
//					TemplateId: "22",
//				},
//			},
//			want: `{"serialNo":1001}`,
//		},
//		{
//			name: "demo07_direct",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID": "DC0001",
//					"k_RAWMSG":      "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS": "1",
//					"k_TITTLE":      "测试",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "k_TITTLE",
//							DstField:  "title",
//							AlarmType: "ALL",
//							MapType:   "direct",
//							MapExpr:   "k_TITTLE",
//						},
//					},
//				},
//				arg: MatchArg{},
//			},
//			want: `{"title":"测试"}`,
//		},
//		{
//			name: "demo07_direct02",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID": "DC0001",
//					"k_RAWMSG":      "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS": "1",
//					"k_TITTLE":      "测试",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "k_TITTLE",
//							DstField:  "title",
//							AlarmType: "ALL",
//							MapType:   "direct",
//							MapExpr:   "k_TITTLEaa",
//						},
//					},
//				},
//				arg: MatchArg{},
//			},
//			want:    "",
//			wantErr: true,
//		},
//		{
//			name: "demo07_regCase",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID": "DC0001",
//					"k_RAWMSG":      "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS": "1",
//					"k_TITTLE":      "测试-01",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "testNo",
//							DstField:  "title",
//							AlarmType: "ALL",
//							MapType:   "regCase",
//							MapExpr:   `{k_TITTLE} match {测试} {k_TITTLE,'测试-(\d+)'}`,
//						},
//					},
//				},
//			},
//			want: `{"title":"01"}`,
//		},
//		{
//			name: "demo07_regCase02",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID": "DC0001",
//					"k_RAWMSG":      "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS": "1",
//					"k_TITTLE":      "测试-01",
//					"devName":       "华为设备",
//					"k_RESCLASS":    "DEV",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "testNo",
//							DstField:  "k_RESNAME",
//							AlarmType: "ALL",
//							MapType:   "regCase",
//							MapExpr:   `{k_RESCLASS} match {DEV} {devName,(\S+)};{k_RESCLASS} match {PORT} {cirName,(\S+)}`,
//						},
//					},
//				},
//			},
//			want: `{"k_RESNAME":"华为设备"}`,
//		},
//		{
//			name: "demo07_regCase03",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID": "DC0001",
//					"k_RAWMSG":      "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS": "1",
//					"k_TITTLE":      "测试-01",
//					"devName":       "华为设备",
//					"cirName":       "电路名称",
//					"k_RESCLASS":    "PORT",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "testNo",
//							DstField:  "k_RESNAME",
//							AlarmType: "ALL",
//							MapType:   "regCase",
//							MapExpr:   `{k_RESCLASS} match {DEV} {devName,(\S+)};{k_RESCLASS} match {PORT} {cirName,(\S+)}`,
//						},
//					},
//				},
//			},
//			want: `{"k_RESNAME":"电路名称"}`,
//		},
//		{
//			name: "demo07_regCase04",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID": "DC0001",
//					"k_RAWMSG":      "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS": "1",
//					"k_TITTLE":      "测试-01",
//					"devName":       "华为设备",
//					"cirName":       "电路名称",
//					"k_RESCLASS":    "PORT",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "testNo",
//							DstField:  "k_RESNAME",
//							AlarmType: "ALL",
//							MapType:   "regCase",
//							MapExpr:   `{k_RESCLASS} match {DEV} {devName,(\S+)};{k_RESCLASS} match {PORT} {cirName,(\S+)}`,
//						},
//					},
//				},
//			},
//			want: `{"k_RESNAME":"电路名称"}`,
//		},
//
//		{
//			name: "demo09_dict",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID":   "DC0001",
//					"k_ALARMTYPEDESC": "DC0001",
//					"k_RAWMSG":        "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS":   "1",
//					"k_TITTLE":        "测试-01",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "k_ALARMTYPEID",
//							DstField:  "k_ALARMTYPEDESC",
//							AlarmType: "ALL",
//							MapType:   "dict",
//							MapExpr:   `dict_1`,
//						},
//					},
//					Dict: model.Dicts{
//						{
//							DictCode: "dict_1",
//							DictList: [][]string{
//								{
//									"DC0001", "设备故障",
//								},
//							},
//						},
//					},
//				},
//			},
//			want: `{"k_ALARMTYPEDESC":"设备故障"}`,
//		},
//		{
//			name: "demo10_filter_01",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID":   "DC0001",
//					"k_ALARMTYPEDESC": "DC0001",
//					"k_RAWMSG":        "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS":   "1",
//					"k_TITTLE":        "测试-01",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "k_ALARMTYPEID",
//							DstField:  "k_ALARMTYPEDESC",
//							AlarmType: "DC0001",
//							MapType:   "dict",
//							MapExpr:   `dict_1`,
//						},
//					},
//					Dict: model.Dicts{
//						{
//							DictCode: "dict_1",
//							DictList: [][]string{
//								{
//									"DC0001", "设备故障",
//								},
//							},
//						},
//					},
//				},
//			},
//			want: `{"k_ALARMTYPEDESC":"设备故障"}`,
//		},
//		{
//			name: "demo10_filter_02",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID":   "DC0001",
//					"k_ALARMTYPEDESC": "DC0001",
//					"k_RAWMSG":        "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS":   "1",
//					"k_TITTLE":        "测试-01",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "k_ALARMTYPEID",
//							DstField:  "k_ALARMTYPEDESC",
//							AlarmType: "DC0002",
//							MapType:   "dict",
//							MapExpr:   `dict_1`,
//						},
//					},
//					Dict: model.Dicts{
//						{
//							DictCode: "dict_1",
//							DictList: [][]string{
//								{
//									"DC0001", "设备故障",
//								},
//							},
//						},
//					},
//				},
//			},
//			want:    "",
//			wantErr: true,
//		},
//		{
//			name: "demo10_filter_03",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID":   "DC0001",
//					"k_ALARMTYPEDESC": "DC0001",
//					"k_RAWMSG":        "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS":   "1",
//					"k_TITTLE":        "测试-01",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "k_ALARMTYPEID",
//							DstField:  "k_ALARMTYPEDESC",
//							AlarmType: "DC0002",
//							MapType:   "dict",
//							MapExpr:   `dict_1`,
//						},
//						{
//							SrcField:  "k_ALARMTYPEID",
//							DstField:  "k_ALARMTYPEDESC",
//							AlarmType: "DC0001",
//							MapType:   "dict",
//							MapExpr:   `dict_1`,
//						},
//					},
//					Dict: model.Dicts{
//						{
//							DictCode: "dict_1",
//							DictList: [][]string{
//								{
//									"DC0001", "设备故障",
//								},
//							},
//						},
//					},
//				},
//			},
//			want: `{"k_ALARMTYPEDESC":"设备故障"}`,
//		},
//		{
//			name: "demo10_filter_03",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID":   "DC0001",
//					"k_ALARMTYPEDESC": "DC0001",
//					"k_RAWMSG":        "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS":   "1",
//					"k_TITTLE":        "测试-01",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "k_ALARMTYPEID",
//							DstField:  "k_ALARMTYPEDESC",
//							AlarmType: "DC0001",
//							MapType:   "dict",
//							MapExpr:   `dict_1`,
//						},
//						{
//							SrcField:  "",
//							DstField:  "desc",
//							AlarmType: "DC0001",
//							MapType:   "join",
//							MapExpr:   "${k_ALARMTYPEID}::${k_TITTLE}",
//						},
//					},
//					Dict: model.Dicts{
//						{
//							DictCode: "dict_1",
//							DictList: [][]string{
//								{
//									"DC0001", "设备故障",
//								},
//							},
//						},
//					},
//				},
//			},
//			want: `{"k_ALARMTYPEDESC":"设备故障","desc":"DC0001::测试-01"}`,
//		},
//		{
//			name: "demo10_filter_04",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID":   "DC0001",
//					"k_ALARMTYPEDESC": "DC0001",
//					"k_RAWMSG":        "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS":   "1",
//					"k_TITTLE":        "测试-01",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "k_ALARMTYPEID",
//							DstField:  "k_ALARMTYPEDESC",
//							AlarmType: "DC0001",
//							MapType:   "dict",
//							MapExpr:   `dict_1`,
//						},
//						{
//							SrcField:  "",
//							DstField:  "k_RAWMSG",
//							AlarmType: "ALL",
//							MapType:   "direct",
//							MapExpr:   "k_ALARMTYPEID",
//						},
//						{
//							SrcField:  "",
//							DstField:  "k_NOTFOUND",
//							AlarmType: "ALL",
//							MapType:   "direct",
//							MapExpr:   "k_NOTFOUND",
//						},
//						{
//							SrcField:  "",
//							DstField:  "desc",
//							AlarmType: "DC0001",
//							MapType:   "join",
//							MapExpr:   "${k_RAWMSG}::${k_NOTFOUND}::${k_TITTLE}",
//						},
//					},
//					Dict: model.Dicts{
//						{
//							DictCode: "dict_1",
//							DictList: [][]string{
//								{
//									"DC0001", "设备故障",
//								},
//							},
//						},
//					},
//				},
//			},
//			want: `{"k_ALARMTYPEDESC":"设备故障","desc":"DC0001::测试-01"}`,
//		},
//		{
//			name: "demo10_filter_04",
//			args: args{
//				ctx: ctx,
//				m: NewMsgMapWithMap(condKey, map[string]interface{}{
//					"k_ALARMTYPEID":   "DC0001",
//					"k_ALARMTYPEDESC": "DC0001",
//					"k_RAWMSG":        "aaaaaaaasdfasdfwikenlucy",
//					"k_ALARMSTATUS":   "1",
//					"k_TITTLE":        "测试-01",
//				}),
//				r: model.ConvertRule{
//					Map: []model.Elem{
//						{
//							SrcField:  "k_ALARMTYPEID",
//							DstField:  "k_ALARMTYPEDESC",
//							AlarmType: "DC0001",
//							MapType:   "dict",
//							MapExpr:   `dict_1`,
//						},
//						{
//							SrcField:  "",
//							DstField:  "k_RAWMSG",
//							AlarmType: "ALL",
//							MapType:   "direct",
//							MapExpr:   "k_ALARMTYPEID",
//						},
//						{
//							SrcField:  "",
//							DstField:  "desc",
//							AlarmType: "DC0001",
//							MapType:   "join",
//							MapExpr:   "${k_RAWMSG}::${k_NOTFOUND}::${k_TITTLE}",
//						},
//					},
//					Dict: model.Dicts{
//						{
//							DictCode: "dict_1",
//							DictList: [][]string{
//								{
//									"DC0001", "设备故障",
//								},
//							},
//						},
//					},
//				},
//			},
//			want: `{"k_ALARMTYPEDESC":"设备故障","desc":"DC0001::测试-01"}`,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := NewTemplateConvertor(TemplateConvertorOption{
//				NotFoundPolicy: NotFoundPolicyExpr,
//				ApiCache: ApiCacheOption{
//					CacheKey: func(varLst []string) string {
//						return ""
//					},
//				},
//			}, &CacheDemo{})
//
//			got := c.MatchAndConvert(tt.args.ctx, tt.args.m, tt.args.r, tt.args.arg)
//			if got != tt.want {
//				t.Errorf("MatchAndConvert() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//type CacheDemo struct {
//}
//
//func (c CacheDemo) GetString(ctx kctx.Ctx, key string) (string, error) {
//	return "aaaa", nil
//}
//
//func (c CacheDemo) SetStringWithExpire(ctx kctx.Ctx, key, val string, expire int) error {
//	return nil
//}
//
//var (
//	redisAddr = "192.168.5.136:17001"
//	redisPwd  = "Jypzcqn!@)!*666"
//)
//
//func BenchmarkName(b *testing.B) {
//	ctx := kctx.NewContext(context.TODO(), "aaaaaaaaaaaaaaaaaaasdflk")
//
//	c, cf, err := NewTemplateConvertorWithRedis(ctx, "node", redisAddr, redisPwd, 0, TemplateConvertorOption{})
//	if err != nil {
//		b.Error(err)
//		return
//	}
//	defer cf()
//
//	for i := 0; i < b.N; i++ {
//		c.GetSerialNo(ctx, "bantch", "test", "wiken", 0, 1)
//		//if err != nil {
//		//	b.Log(err)
//		//} else {
//		//	b.Logf("serilaNo: %d", n)
//		//}
//	}
//}
