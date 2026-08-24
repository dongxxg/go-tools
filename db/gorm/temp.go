package db

import (
	"bytes"
	"fmt"
	"github.com/araddon/dateparse"
	"reflect"
	"time"
)

const valueSql = `
{{$length := len .AllValues}}
{{ range $r,$row := .AllValues}}{{$cols := len $row}}{{if lastRow $length $r}}({{ range $index,$v := $row }}{{if lastCol $cols $index}}{{dealData $v}}{{else}}{{dealData $v}},{{end}}{{end}});{{else}}({{ range $index,$v := $row }}{{if lastCol $cols $index}}{{dealData $v}}{{else}}{{dealData $v}},{{end}}{{end}}),{{end}}
{{end}}
`

//处理最后一条数据
func DealLastRow(length, max int) bool {
	return length-1 == max
}

//处理最后一列数据
func DealLastData(length, max int) bool {
	return length-1 == max
}

func DealData(value interface{}) interface{} {
	if value == "NULL" { //空值处理
		return value
	}
	switch reflect.ValueOf(value).Kind() {
	case reflect.String:
		rs := bytes.Runes([]byte(value.(string)))
		if len(rs) == 1 {
			if rs[0] == 0 || rs[0] == 1 { //排除0和1的情况
				return rs[0]
			} else {
				value = string(rs)
			}

		}
		return fmt.Sprintf("'%s'", value)
	case reflect.Struct:
		str := fmt.Sprintf("%+v", value)
		t1, err := DealTime(str)
		if err != nil {
			fmt.Println(err)
			return value
		} else {
			return fmt.Sprintf("'%+v'", t1)
		}
	default:
		fmt.Println("current value is ", value, "and type is ", reflect.ValueOf(value).Kind())
		return value
	}
	return value
}

func DealTime(data string) (string, error) {
	col, _ := time.LoadLocation("Asia/Shanghai")
	t1, err_t1 := dateparse.ParseIn(data, col)
	if err_t1 != nil {
		fmt.Println("deal time is err ", err_t1)
		return "", err_t1
	}
	format := "2006-01-02 15:04:05"
	return t1.Format(format), nil
}
