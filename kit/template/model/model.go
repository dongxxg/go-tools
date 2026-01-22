package model

import "unitechs.com/cno/kit/kctx"

const (

	// ModuleName 日志的模块名称
	ModuleName = "module"
)

type IMsg interface {
	Get(k string) (interface{}, bool)
	GetCondKey() (key, val string, exist bool)
	GetStr(key string) string
}

// ICache 缓存接口
type ICache interface {
	HGetString(ctx kctx.Ctx, key, field string) (string, error)
	GetString(ctx kctx.Ctx, key string) (string, error)
	SetStringWithExpire(ctx kctx.Ctx, key, val string, expire int) error
}
