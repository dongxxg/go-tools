package conf

import (
	"github.com/spf13/viper"
	"strings"
	"sync"
	"unitechs.com/unios-dice/uni-base/core/config"
)

type KitConf struct {
	*viper.Viper
	// keyMap 缓存nacos与本地配置文件的环境变量的key的映射
	// 如 {"DEMO_NAME":"Demo.Name"}, 当nacos 配置变化的时候，可以通过nacos的dataId 找到配置文件的key
	keyMap sync.Map
	option Option
}

type Option struct {
	isNacosSync bool
	nacosGroup  string // nacos的配置组
}
type OptionFunc func(o *Option)

// WithNacosSync 是否同步nacos 数据变化到本地配置， 默认false
func WithNacosSync(isSync bool) OptionFunc {
	return func(o *Option) {
		o.isNacosSync = isSync
	}
}
func WithNacosGroup(group string) OptionFunc {
	return func(o *Option) {
		o.nacosGroup = group
	}
}

var kitConf *KitConf

func GetKitConf() *KitConf {
	return kitConf
}

func NewKitConf(viper *viper.Viper, options ...OptionFunc) *KitConf {
	op := new(Option)
	for _, o := range options {
		o(op)
	}
	kitConf = &KitConf{
		Viper: viper,
	}
	if op.isNacosSync {
		if op.nacosGroup == "" {
			op.nacosGroup = kitConf.GetStringWithDefault("SD.Client.Group", "DEFAULT_GROUP")
		}

		for _, v := range viper.AllKeys() {
			key := strings.ToUpper(strings.ReplaceAll(v, ".", "_"))
			kitConf.keyMap.Store(key, v)
			func(v string) {
				config.ListenConfig(op.nacosGroup, key, func(namespace, group, dataId, data string) {
					viper.Set(v, data)
				})
			}(v)
		}
	}
	kitConf.option = *op

	return kitConf
}
func (k *KitConf) UnmarshalKey(key string, rawVal interface{}) error {
	return k.Viper.UnmarshalKey(key, rawVal)
}

func (k *KitConf) GetStringWithDefault(key string, v string) string {
	s := k.GetString(key)
	if s == "" {
		s = v
	}
	return s
}
func (k *KitConf) Get(key string) (string, bool) {
	v, ok := k.keyMap.Load(key)
	if !ok {
		return "", false
	}

	if val, ok := v.(string); ok {
		return val, true
	}

	return "", false
}

func (k *KitConf) IsNacosSync() bool {
	return k.option.isNacosSync
}
func (k *KitConf) NacosGroup() string {
	return k.option.nacosGroup
}

func (k *KitConf) GetIntWithDefault(key string, v int) int {
	s := k.GetInt(key)
	if s == 0 {
		s = v
	}
	return s
}
func (k *KitConf) GetInt64WithDefault(key string, v int64) int64 {
	s := k.GetInt64(key)
	if s == 0 {
		s = v
	}
	return s
}
func (k *KitConf) GetInt32WithDefault(key string, v int32) int32 {
	s := k.GetInt32(key)
	if s == 0 {
		s = v
	}
	return s
}
func (k *KitConf) GetBoolWithDefault(key string, v bool) bool {
	s := k.GetBool(key)
	return s
}
