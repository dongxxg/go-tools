package util

import (
	"fmt"
	"reflect"
)

// StructToMap 注意该函数中的入参s,不支持指针；使用时注意
func StructToMap(s interface{}) map[string]string {
	data := make(map[string]string)

	t := reflect.TypeOf(s)
	v := reflect.ValueOf(s)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()

		// 忽略未导出的字段和非字符串类型的 tag
		if field.PkgPath != "" {
			continue
		}

		if field.Type.Kind() == reflect.Struct {
			// 递归处理嵌套的结构体
			for k, v := range StructToMap(value) {
				data[k] = v
			}
			continue
		}

		tag := field.Tag.Get("json")
		if tag == "" {
			tag = field.Name
		}

		data[tag] = fmt.Sprintf("%v", value)
	}

	return data
}
