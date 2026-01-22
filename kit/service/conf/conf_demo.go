package conf

import (
	"github.com/spf13/cast"
	"strconv"
	"strings"
)

// MapConfig 这是同map 实现的配置文件，主要用于测试使用， 不可使用在正式代码中
type MapConfig struct {
	Config map[string]string
}

func NewMapConfig(config map[string]string) *MapConfig {
	return &MapConfig{Config: config}
}

func (d *MapConfig) Set(key string, val interface{}) {
	d.Config[key] = cast.ToString(val)
	return
}

func (d *MapConfig) IsNacosSync() bool {
	return false
}

func (d *MapConfig) NacosGroup() string {
	return ""
}

func (d *MapConfig) GetString(key string) string {
	return d.Config[key]
}

func (d *MapConfig) GetBool(key string) bool {
	v, ok := d.Config[key]
	if ok {
		if strings.ToLower(v) == "true" {
			return true
		}
	}
	return false
}

func (d *MapConfig) GetInt(key string) int {
	v, ok := d.Config[key]
	if ok {
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0
		}
		return n
	}
	return 0
}

func (d *MapConfig) GetStringWithDefault(key string, v string) string {
	n := d.GetString(key)
	if n == "" {
		return v
	}

	return n
}

func (d *MapConfig) GetIntWithDefault(key string, v int) int {
	n := d.GetInt(key)
	if n == 0 {
		return v
	}
	return n
}

func (d *MapConfig) GetInt64WithDefault(key string, v int64) int64 {
	n := d.GetIntWithDefault(key, int(v))
	return int64(n)
}

func (d *MapConfig) GetInt32WithDefault(key string, v int32) int32 {
	n := d.GetIntWithDefault(key, int(v))
	return int32(n)
}

func (d *MapConfig) GetBoolWithDefault(key string, v bool) bool {
	return d.GetBool(key)
}
