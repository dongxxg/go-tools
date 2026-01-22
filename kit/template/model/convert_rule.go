// 任务号    日期     开发者      描述

package model

// ConvertRule 转换规则
type ConvertRule struct {
	Map  Elems `json:"map"`
	Dict Dicts `json:"dict"`
}

// Dicts 字段列表
type Dicts []Dict

type Dict struct {
	DictCode string     `json:"dictCode"`
	DictList [][]string `json:"dictList"`
}

type Elems []Elem

type Elem struct {
	SrcField  string  `json:"srcField"`  // 源字段
	DstField  string  `json:"dstField"`  // 目标字段
	AlarmType Type    `json:"alarmType"` // 告警类型
	MapType   MapType `json:"mapType"`   // map 类型
	MapExpr   string  `json:"mapExpr"`   //表达式
}
type Type string

func (t Type) String() string {
	return string(t)
}

type MapType string

const (
	TypeAll Type = "ALL"
)

const (
	MapTypeDict              MapType = "dict"
	MapTypeFixed             MapType = "fixed"
	MapTypeDirect            MapType = "direct"
	MapTypeJoin              MapType = "join"
	MapTypeReg               MapType = "reg"
	MapTypeCase              MapType = "case"
	MapTypeNvl               MapType = "nvl"
	MapTypeSerial            MapType = "serial"
	MapTypeAPI               MapType = "api"
	MapTypeRegCase           MapType = "regCase"
	MapTypeFindAttr          MapType = "findAttr"
	MapTypeApiCache          MapType = "apiCache"          // 调用api 带缓存
	MapTypeTimeStdFormat     MapType = "timeStdFormat"     // 转换为yyyy-mm-dd hh:mi:ss
	MapTypeTimeRfc3339       MapType = "timeRfc3339"       // 转换为RFC3339约定时间格式，例如2006-01-02T15:04:05Z07:00
	MapTypeTimeCompactFormat MapType = "timeCompactFormat" // 转换为yyyymmddhhmiss。
	MapTypeTimestamp         MapType = "timeStamp"         // 时间戳
	MapTypeAutoSwitch        MapType = "autoSwitch"
)

type Filter struct {
	FilterName string `json:"fillterName"` // fillterName
	FilterType int    `json:"filterType"`
	FilterExpr string `json:"filterExpr"`
	IsDelay    int    `json:"isDelay"`   // 是否延迟转发 1-是 0-否
	DelayTime  int    `json:"delayTime"` // 延迟时间 单位s
	IsMatched  bool   `json:"-"`         // 是否匹配上
}
