// 任务号    日期     开发者      描述

package kredis

import (
	"errors"
	"github.com/bsm/redislock"
	rds "github.com/redis/go-redis/v9"
	"strings"
	"time"
	"unitechs.com/cno/kit/kctx"
)

var LoopTimes = 1000

// SetLoopTime 设备获取全局锁的次数
func SetLoopTime(n int) {
	LoopTimes = n
}

var redisLocker *redislock.Client

// InitDistributedLock 初始化一个分布式锁, 返回cancelFunc, 退出时调用
func InitDistributedLock(addr, pwd string, db int) func() {
	if strings.Contains(addr, ",") {
		// 如果有多个， 使用一个， 多个可能不支持
		addr = strings.Split(addr, ",")[0]
	}
	client := rds.NewClient(&rds.Options{
		//"127.0.0.1:6379"
		Addr:     addr,
		Password: pwd,
		DB:       db,
	})

	redisLocker = redislock.New(client)
	return func() {
		_ = client.Close()
	}
}

func GetDistributeLock(ctx kctx.Ctx, key string, ttl time.Duration, opt *redislock.Options) (releaseFunc func(), err error) {

	for i := 0; i < LoopTimes; i++ {
		ctx.Debugf("get distribute lock %d times", i)
		lock, err := redisLocker.Obtain(ctx, key, ttl, opt)
		if err == redislock.ErrNotObtained {
			time.Sleep(time.Millisecond * 10)
			continue
		}
		if err != nil {
			kctx.Errorf("Release DistributeLock key failed, key=%s, err: %v", key, err)
			break
		}
		return func() {
			if err := lock.Release(ctx); err != nil {
				kctx.Errorf("Release DistributeLock key failed, key=%s, err: %s", key, err.Error())
			}
		}, nil
	}
	return func() {

	}, errors.New("get distribute lock failed")
}
