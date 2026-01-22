// 任务号    日期     开发者      描述

package util

import (
	"encoding/json"
	"fmt"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/pkg/errors"
	"sort"
	"strconv"
	"strings"
	"time"
)

func ConvertToInt(value interface{}) (val int, err error) {
	switch d := value.(type) {
	case string:
		// 将interface转为string字符串类型
		val, err = strconv.Atoi(d)
	case int:
		val = d
	case int8:
		val = int(d)
	case int16:
		val = int(d)
	case int32:
		val = int(d)
	case int64:
		// 将interface转为int64类型
		strInt64 := strconv.FormatInt(d, 10)
		val, err = strconv.Atoi(strInt64)
	case float64:
		val = int(d)
	case float32:
		val = int(d)
	default:
		s := ConvertToString(value)
		val, err = strconv.Atoi(s)
	}
	return
}

func ConvertToInt64(v interface{}) (int64, error) {
	s := ConvertToString(v)
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "ConvertToInt parse failed")
	}
	return n, nil
}

func ConvertToString(value interface{}) string {
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case time.Time:
		return value.(time.Time).Format(time.RFC3339)
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}

func SplitBy(s, sep string) []string {
	l := strings.Split(s, sep)
	lst := gset.NewStrSet()
	for _, v := range l {
		if v == "" {
			continue
		}
		lst.Add(v)
	}
	return lst.Slice()
}
func StringType2Slice(lst ...interface{}) []string {
	s := gset.NewStrSet()

	for _, v := range lst {
		if ss, ok := v.(string); ok {
			s.Add(ss)
		}
	}

	return s.Slice()
}

func MarshalWith(in interface{}, out interface{}) error {

	b, err := json.Marshal(in)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal err: %v", err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		return errors.Wrapf(err, "failed, to unmarshal err: %v", err)
	}
	return nil
}

// ConvertJsonString  将配置文件json转换为 环境变量
// prefix： KeyValue.
func ConvertJsonString(jsonString, prefix string) ([]string, error) {
	var input interface{}
	err := json.Unmarshal([]byte(jsonString), &input)
	if err != nil {
		return nil, err
	}

	output := make(map[string]interface{})
	traverseJsonNode(input, "", output)

	out := make([]string, 0, len(output))
	for k, v := range output {
		//temp[prefix+"."+k] = v
		if v == "" {
			out = append(out, fmt.Sprintf("%s.%s=\"\"", prefix, k))
			continue
		}

		if v == "true" || v == "false" {
			out = append(out, fmt.Sprintf("%s.%s=%v", prefix, k, v))
			continue
		}
		s := ConvertToString(v)
		n, err := strconv.Atoi(s)
		if err == nil {
			out = append(out, fmt.Sprintf("%s.%s=%v", prefix, k, n))
			continue
		}
		out = append(out, fmt.Sprintf("%s.%s=\"%v\"", prefix, k, v))
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i] < out[j]
	})
	return out, nil
	//return strings.Join(out, "\n"), nil
}

func convertToUpperCase(s string) string {
	return strings.ReplaceAll(strings.ToUpper(s), ".", "_")
}

func traverseJsonNode(node interface{}, prefix string, result map[string]interface{}) {
	switch v := node.(type) {
	case map[string]interface{}:
		for key, value := range v {
			traverseJsonNode(value, prefix+"."+key, result)
		}
	case []interface{}:
		for i, value := range v {
			traverseJsonNode(value, fmt.Sprintf("%s[%d]", prefix, i), result)
		}
	default:
		result[convertToUpperCase(prefix[1:])] = v
	}
}
