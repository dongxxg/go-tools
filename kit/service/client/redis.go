package client

import (
	"strconv"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/unios-dice/uni-base/lib/stores/redis"
)

// NewRedisCli
/*
Type = "node"
Host = "localhost:6379"
Password = ""
Db = 0
PoolSize = 500
*/
func NewRedisCli(ctx kctx.Ctx, getter IConfigGetter, prefix string) *redis.Redis {
	Type := prefix + ".Type"
	host := prefix + ".Host"
	pwd := prefix + ".Password"
	db := prefix + ".Db"
	poolSize := prefix + ".PoolSize"

	conf := redis.RedisConf{
		Type:     getter.GetString(Type),
		Host:     getter.GetString(host),
		Pass:     getter.GetString(pwd),
		Db:       getter.GetInt(db),
		PoolSize: getter.GetInt(poolSize),
	}

	conn := conf.NewRedis()
	if getter.IsNacosSync() {
		lst := []string{Type, host, pwd, db}
		for _, v := range lst {
			func(v string) {
				if err := ListenNacos(getter, v, func(namespace, group, dataId, data string) {
					ctx.Infof("es config change %s  %s", v, data)
					getter.Set(v, data)
					switch v {
					case Type:
						conn.Type = data
					case host:
						conn.Addr = data
					case pwd:
						conn.Pass = data
					case db:
						d, err := strconv.Atoi(data)
						if err != nil {
							ctx.Errorf("convert nacos config %s failed:%v", v, err)
							return
						}
						conn.Db = d
					}
				}); err != nil {
					ctx.Errorf("listen nacos config failed: %v", err)
				}
			}(v)
		}
	}

	return conn
}

// NewRedisCliFormat2
/*
Type = "node"
Host = "localhost:6379"
Password = ""
Database = 0
*/
func NewRedisCliFormat2(ctx kctx.Ctx, getter IConfigGetter, prefix string) *redis.Redis {
	conf := redis.RedisConf{
		Type: getter.GetString(prefix + ".Type"),
		Host: getter.GetString(prefix + ".Host"),
		Pass: getter.GetString(prefix + ".Password"),
		Db:   getter.GetInt(prefix + ".Database"),
	}

	return conf.NewRedis()
}
