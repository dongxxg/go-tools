package redis

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"go-tools/support/logger"
	"reflect"
	"strconv"
	"time"
)

var ctx = context.Background()
var rdbClient *RedisClient
var rdb *redis.Client

const errMsgNotInitRedis string = "redis is not init success"

type RedisConfig struct {
	Addr     string
	Password string
	Db       string
	UserName string
}

type RedisClient struct {
	rdb *redis.Client
}

func InitRedis(config *RedisConfig) error {
	dbNum, err := strconv.Atoi(config.Db)
	if err != nil {
		return err
	}
	options := &redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       dbNum,
	}
	if config.UserName != "" {
		options.Username = config.UserName
	}
	rdb = redis.NewClient(options)
	rdbClient = &RedisClient{
		rdb: rdb,
	}
	return nil
}

func GetRedisClient() (*RedisClient, error) {
	if rdbClient == nil {
		return nil, errors.New("redis client is not exist")
	}
	// 检查是否连通，不连通就保存，尝试连接
	return rdbClient, nil
}

func GetRawRedisClient() *redis.Client {
	// 没有初始化的原生redis客户端禁止获取
	if rdb == nil {
		panic("raw redis client is not init")
	}
	return rdb
}

// 加锁
func (r *RedisClient) Lock(key string, value string, expireTimeSecond time.Duration) bool {
	if !r.preCheck() {
		panic(errMsgNotInitRedis)
	}
	res := r.rdb.SetNX(ctx, key, value, expireTimeSecond)
	return res.Val()
}

// 解锁
func (r *RedisClient) UnLock(key string) bool {
	if !r.preCheck() {
		panic(errMsgNotInitRedis)
	}
	res := r.rdb.Del(ctx, key)
	return res.Err() == nil
}

// ReleaseLock 使用Lua脚本原子性地判断锁的所有者并释放锁
func (r *RedisClient) ReleaseLock(key string, identifier string) bool {
	if !r.preCheck() {
		panic(errMsgNotInitRedis)
	}
	// identifier是唯一标识，也就是key对应的value，和当前客户端标识相同时才能删除
	script := redis.NewScript(`
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
    `)
	result, err := script.Run(ctx, r.rdb, []string{key}, identifier).Result()
	if err != nil {
		return false
	}

	return result == int64(1)
}

func (r *RedisClient) preCheck() bool {
	return r.rdb != nil
}

type RedisValue struct {
	Value interface{} `json:"value"`
}

func NewRedisValue(v interface{}) RedisValue {
	return RedisValue{Value: v}
}

func (v *RedisValue) GetValue() interface{} {
	if v.Value == nil {
		return ""
	}
	switch reflect.TypeOf(v.Value).Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array, reflect.Ptr, reflect.Interface:
		res, err := json.Marshal(v.Value)
		if err != nil {
			//Channel, complex, and function values 这三个值
			return v.Value
		} else {
			return res
		}
	case reflect.Func, reflect.Chan:
		return ""
	default:
		return v.Value
	}
}

// value 支持传入字符、数值、布尔、slice、数组、结构体、map;如果RedisValue 改为interface{},无法支持map、结构体、slice等复杂数据类型
// if expireTimeSecond is 0 ,means the key has no expiration time
func (r *RedisClient) Set(key string, value RedisValue, expireTime time.Duration) error {
	if !r.preCheck() {
		return errors.New(errMsgNotInitRedis)
	} else if len(key) == 0 {
		return errors.New("cant support key len is 0")
	} else {
		res := r.rdb.Set(ctx, key, value.GetValue(), expireTime)
		return res.Err()
	}
}

func (r *RedisClient) Get(key string) (data string, isExist bool) {
	if !r.preCheck() {
		logger.Error(errMsgNotInitRedis)
		return "", false
	} else {
		res := r.rdb.Get(ctx, key)
		return res.Val(), res.Err() != redis.Nil
	}
}

// 删除key
func (r *RedisClient) Del(key string) {
	if !r.preCheck() {
		return
	} else {
		r.rdb.Del(ctx, key)
	}
}

// 获取key的过期时间
func (r *RedisClient) GetExpire(key string) (time.Duration, error) {
	if !r.preCheck() {
		return 0, errors.New(errMsgNotInitRedis)
	} else {
		res := r.rdb.TTL(ctx, key)
		return res.Val(), res.Err()
	}
}

// 执行Ping操作
func (r *RedisClient) Ping(ctx context.Context) error {
	_, err := r.rdb.Ping(ctx).Result()
	return err
}

// SetNX 不存在才插入
func (r *RedisClient) SetNX(key string, value RedisValue, expireTimeSecond time.Duration) error {
	if !r.preCheck() {
		panic(errMsgNotInitRedis)
	} else if len(key) == 0 {
		return errors.New("cant support key len is 0")
	} else {
		res := r.rdb.SetNX(ctx, key, value.GetValue(), expireTimeSecond)
		return res.Err()
	}
}

// Expire 刷新过期时间
func (r *RedisClient) Expire(key string, expireTimeSecond time.Duration) error {
	if !r.preCheck() {
		return errors.New(errMsgNotInitRedis)
	} else {
		res := r.rdb.Expire(ctx, key, expireTimeSecond)
		return res.Err()
	}
}

// 获取原始RDB
func (r *RedisClient) GetRdb() *redis.Client {
	return r.rdb
}
