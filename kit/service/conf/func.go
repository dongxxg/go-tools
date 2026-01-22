package conf

import (
	"strings"
)

func GenNacosKey(key string) string {
	return strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
}

type ConfigKey struct {
	prefix string
}

func NewConfigKey(prefix string) *ConfigKey {
	return &ConfigKey{prefix: prefix}
}

func (c ConfigKey) Key(k string) string {
	return c.prefix + "." + k
}
