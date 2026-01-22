package kafka_group

import (
	"fmt"
	"github.com/pkg/errors"
	"time"
	redisLock "unitechs.com/cno/kit/db/kredis/lock"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/unios-dice/uni-base/lib/stores/redis"
)

// GroupGetter
// 生成kafka消费组group
// 对于告警的通知，不同的实例，需要使用不同的group。但是现场的group需要提前申请
// 所以写这个消费组的获取器
// 1. 先将所有的group 设置进redis
// 2. 每一个消费组， 有状态，心跳时间等信息
// 3. 每一个实例，定期向redis自己的消费组发送心跳
// 4. 如果一个实例，没有发送心跳，那么认为其已经宕机了
// 5. 如果一个实例，已经宕机，那么删除redis实例信息
// 6. 需要兼容高并发
type GroupGetter struct {
	ctx          kctx.Ctx
	redisPrefix  string //  消费组hash的前缀
	rds          *redis.Redis
	rdsLock      *redisLock.Lock
	groups       []string
	pushInterval time.Duration //  心跳时间间隔, 即推送每两次心跳之间的间隔
	maxInterval  time.Duration //  最大心跳间隔， 如当前获取的心跳间隔超过这个值，那么认为其已经宕机了， 则删除掉该实例
}

func NewGroupGetter(ctx kctx.Ctx, hashPrefix, lockKey string, rds *redis.Redis, maxInterval, pushInterval time.Duration) (*GroupGetter, error) {
	// lockKey 全局锁的名称
	rdsLock, err := redisLock.NewRedisLock(rds, lockKey, 2)
	if err != nil {
		return nil, errors.Wrapf(err, "NewRedisLock")
	}
	if maxInterval == 0 {
		maxInterval = time.Second * 10
	}
	if pushInterval == 0 {
		pushInterval = time.Second * 5
	}

	return &GroupGetter{
		ctx:          ctx,
		rds:          rds,
		rdsLock:      rdsLock,
		redisPrefix:  hashPrefix,
		maxInterval:  maxInterval,
		pushInterval: pushInterval,
	}, nil
}
func (g *GroupGetter) Run(ctx kctx.Ctx) {
	defer func() {
		if err := recover(); err != nil {
			ctx.Errorf("GroupGetter.Run panic: %v", err)
		}
	}()
	if g.pushInterval == 0 {
		g.pushInterval = time.Second * 5
	}
	// 这里需要使用协程
	go func() {
		ticker := time.NewTicker(g.pushInterval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for _, v := range g.groups {
					// 生成hashName
					hashName := g.genHashName(v)
					if err := g.rds.Hmset(hashName, map[string]string{
						"timestamp": time.Now().Format(time.RFC3339),
					}); err != nil {
						ctx.Errorf("update hash failed: %v", err)
					}
				}
			}
		}
	}()
}

func (g *GroupGetter) SetGroups(groups ...string) {
	g.ctx.Infof("GroupGetter.SetGroups set groups %v", groups)
	g.groups = groups
}

func (g *GroupGetter) GetGroup(ctx kctx.Ctx, groups ...string) (string, error) {
	if len(groups) == 0 {
		return "", fmt.Errorf("GroupGetter.GetGroup no groups input")
	}
	ok, err := g.rdsLock.AcquireLock()
	if err != nil {
		return "", errors.Wrap(err, "GroupGetter.GetGroup")
	}
	if !ok {
		return "", errors.New("get redisLock failed")
	}
	defer func() {
		err = g.rdsLock.ReleaseLock()
		if err != nil {
			ctx.Errorf("GroupGetter.GetGroup release lock failed: %v", err)
		}
	}()

	for _, v := range groups {
		// 生成hashName
		hashName := g.genHashName(v)

		// 从redis中获取hash
		hash, err := g.rds.Hgetall(hashName)
		if err != nil {
			ctx.Errorf("GroupGetter.GetGroup %s get hash failed: %v", v, err)
			continue
		}

		// 如果hash为空，那么说明没有设置过, 那么这里将使用这个
		if len(hash) == 0 {
			// 设置redis
			if err := g.DelAndGenNewInfoAndSet(v, "used"); err != nil {
				ctx.Errorf("GroupGetter.GetGroup %s del and gen new info failed: %v", v, err)
				continue
			}

			return v, nil
		}
		latestTime, ok := hash["timestamp"]
		if !ok {
			// 设置redis
			if err := g.DelAndGenNewInfoAndSet(v, "used"); err != nil {
				ctx.Errorf("GroupGetter.GetGroup %s del and gen new info failed: %v", v, err)
				continue
			}
			return v, nil
		}
		latestTimeInt, err := time.Parse(time.RFC3339, latestTime)
		if err != nil {
			// 设置redis
			if err := g.DelAndGenNewInfoAndSet(v, "used"); err != nil {
				ctx.Errorf("GroupGetter.GetGroup %s del and gen new info failed: %v", v, err)
				continue
			}
			return v, nil
		}
		if time.Now().Sub(latestTimeInt) > g.maxInterval {
			// 删除redis
			if err := g.DelAndGenNewInfoAndSet(v, "used"); err != nil {
				ctx.Errorf("GroupGetter.GetGroup %s del and gen new info failed: %v", v, err)
				continue
			}
			return v, nil
		} else {
			// 小于等于
			if hash["status"] != "used" {
				if err := g.DelAndGenNewInfoAndSet(v, "used"); err != nil {
					ctx.Errorf("GroupGetter.GetGroup %s del and gen new info failed: %v", v, err)
					continue
				}
				return v, nil
			}
			// status == used
			continue
		}
	}

	return "", fmt.Errorf("GroupGetter.GetGroup no groups")
}

func (g *GroupGetter) GetKafkaGrop(group, nacosPort string) (string, error) {
	return g.GetGroup(g.ctx, g.groups...)
}

func (g *GroupGetter) DelAndGenNewInfoAndSet(group string, status string) error {
	hashName := g.genHashName(group)
	_ = g.delInfo(hashName)

	info := g.genInfo(status, group)

	if err := g.rds.Hmset(hashName, info); err != nil {
		return err
	}

	return nil
}

func (g *GroupGetter) delInfo(hashName string) error {
	_, err := g.rds.Del(hashName)
	return err
}

func (g *GroupGetter) genInfo(status, group string) map[string]string {
	return map[string]string{
		"status":    status,                          // 状态 ready, used, dead
		"timestamp": time.Now().Format(time.RFC3339), // 最后一次的时间
		"group":     group,
	}
}

func (g *GroupGetter) genHashName(group string) string {
	return g.redisPrefix + ":" + group
}
