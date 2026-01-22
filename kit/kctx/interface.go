package kctx

type IConfigGetter interface {
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	GetStringWithDefault(key string, v string) string
	GetIntWithDefault(key string, v int) int
	GetInt64WithDefault(key string, v int64) int64
	GetInt32WithDefault(key string, v int32) int32
	GetBoolWithDefault(key string, v bool) bool
	UnmarshalKey(key string, rawVal interface{}) error
}
