// 任务号    日期     开发者      描述

package kredis

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/pkg/errors"
	"strings"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/cno/kit/util"
	"unitechs.com/unios-dice/uni-base/lib/stores/redis"
)

func NewRedisClient(ctx kctx.Ctx, Type, host, pwd string, db int) (*redis.Redis, error) {
	conf := redis.RedisConf{Type: Type,
		Host: host,
		Pass: pwd,
		Db:   db,
	}
	ctx.Infof("redis info: %s, host: %s, database: %d", Type, host, db)

	conn := conf.NewRedis()
	if conn == nil {
		return nil, errors.New("new redis client failed, conn is nil")
	}
	ctx.Infof("Successfully connected to redis type: %s, host: %s, database: %d", Type, host, db)
	return conn, nil
}

func NewRedisRepo(ctx kctx.Ctx, Type, host, pwd string, db int) (*RedisRepo, error) {
	cli, err := NewRedisClient(ctx, Type, host, pwd, db)
	if err != nil {
		return nil, errors.Wrapf(err, "new redis client failed err: %v", err)
	}
	return &RedisRepo{conn: cli}, nil
}

func NewRedisRepoWithCli(cli *redis.Redis) *RedisRepo {
	return &RedisRepo{conn: cli}
}

type RedisRepo struct {
	conn *redis.Redis
}

func (r RedisRepo) Get(key string) (string, error) {
	return r.conn.GetAssertValue(key)
}

func (r RedisRepo) Set(key, val string) error {
	return r.conn.Set(key, val)
}

func (r RedisRepo) SetStringWithExpire(ctx kctx.Ctx, key, val string, expire int) error {
	return r.conn.Setex(key, val, expire)
}

func (r RedisRepo) GetDB(ctx kctx.Ctx) *redis.Redis {
	return r.conn
}

func (r RedisRepo) GetString(ctx kctx.Ctx, key string) (string, error) {
	val, err := r.conn.GetAssertValue(key)
	if err != nil {
		return "", errors.Wrapf(err, "get string failed, err: %v", err)
	}
	return val, nil
}

func (r RedisRepo) SetString(ctx kctx.Ctx, key string, val string) error {
	return r.conn.Set(key, val)
}

func (r RedisRepo) HGetString(ctx kctx.Ctx, key, field string) (string, error) {
	val, err := r.conn.Hget(key, field)
	if err != nil {
		return "", errors.Wrapf(err, "hget string failed, err: %v", err)
	}
	return val, nil
}

func (r RedisRepo) GenSerialNoWithLua(ctx kctx.Ctx, key string, start, interval int64) (int64, error) {
	luaScript := `
local base=tonumber(ARGV[1]);
local no=redis.call('get', KEYS[1]);
if not no then
	redis.call('set', KEYS[1], base);
	return base
end
local interval = tonumber(ARGV[2]);
no = tonumber(no); 
if no >= base then 
	no = no + interval; 
	redis.call('set', KEYS[1], no); 
	return no 
else 
	redis.call('set', KEYS[1], base);	 
	return base
end
`
	v, err := r.conn.Eval(luaScript, []string{key}, start, interval)
	if err != nil {
		return 0, errors.Wrapf(err, "redis eval lua failed")
	}
	n, err := util.ConvertToInt64(v)
	if err != nil {
		return 0, errors.Wrapf(err, "convert to int failed %v", v)
	}

	return n, nil
}

const (
	instanceFieldLatestTime = "latest_time"
	instanceFieldStatus     = "status" // 1: running, 2: stopped
)

// PushHeartbeat
// 推送心跳
// hash
// hashName: unios-alarm:alarm:instanceHash:id
// 心跳存储格式
//
//	type redisInstance struct {
//		LatestTime time.Time `json:"latest_time"`
//		Status     int       `json:"status"`  1: running  2 stopped
//	}

type hashTemp struct {
	Id     string
	hashNo uint32
}

func (r RedisRepo) GetHeartbeat(key string) ([]map[string]string, error) {
	ctx := kctx.NewContext(context.TODO(), "get_heartbeat")
	keys, err := r.conn.Get(key)

	//ids, err := keys(ctx, r.conn, key+":*", 1000, 100)
	if err != nil {
		return nil, fmt.Errorf("get instance list failed: %v", err)
	}
	ids := gset.NewStrSetFrom(strings.Split(keys, ","))
	ids.Remove("")

	out := make([]map[string]string, 0, ids.Size())
	for _, v := range ids.Slice() {
		m, err := hgetall(ctx, r.conn, r.instanceName(key, v), 1000, 100)
		if err != nil {
			return nil, errors.Wrapf(err, "get heart beat failed")
		}
		m["id"] = v
		out = append(out, m)
	}
	return out, nil
}

func (r RedisRepo) Hmset(key string, vals map[string]string) error {
	return r.conn.Hmset(key, vals)
}
func (r RedisRepo) instanceName(key, id string) string {
	return key + ":" + id
}

func (r RedisRepo) DelInstance(key string, id string) error {
	if _, err := r.conn.Del(r.instanceName(key, id)); err != nil {
		return errors.Wrapf(err, "del instance failed")
	}
	val, err := r.Get(key)
	if err != nil {
		return errors.Wrapf(err, "get hash list in del instance")
	}
	lst := gset.NewStrSetFrom(strings.Split(val, ","))
	lst.Remove(id)
	if err = r.Set(key, strings.Join(lst.Slice(), ",")); err != nil {
		return errors.Wrapf(err, "set hash list in del isntance")
	}
	return nil
}
