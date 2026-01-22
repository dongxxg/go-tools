package alarm

import (
	"github.com/pkg/errors"
	"strconv"
	"sync"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/unios-dice/uni-base/lib/stores/redis"
)

// NumberGenerator
// 生成递增的序列号
type NumberGenerator struct {
	start     int64        // 序列号起始值
	step      int64        // 序列号间隔值
	cur       int64        // 当前序列号
	max       int64        // 缓存的最大序列号
	client    *redis.Redis // Redis客户端
	mutex     sync.Mutex   // 互斥锁
	batchSize int64        // 一次拿多少数量
	prefix    string
}

func (gen *NumberGenerator) Get(ctx kctx.Ctx, ruleId, templateId, dstField string, start, interval int64) (int64, error) {
	//
	if gen.start == 0 {
		gen.start = start
	}
	if gen.step == 0 {
		gen.step = interval
	}

	return gen.next()
}

func NewNumberGenerator(start int64, step int64, rds *redis.Redis, prefix string) *NumberGenerator {
	gen := &NumberGenerator{
		prefix:    prefix,
		start:     start,
		step:      step,
		cur:       0,
		max:       0,
		batchSize: 100,
		mutex:     sync.Mutex{},
	}
	gen.client = rds
	return gen
}
func (gen *NumberGenerator) BatchSize(b int64) {
	gen.batchSize = b
}
func (gen *NumberGenerator) next() (int64, error) {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()
	if gen.cur+gen.step < gen.max {
		gen.cur += gen.step
		if gen.cur >= gen.start {
			return gen.cur, nil
		}
	}
	err := gen.Load()
	if err != nil {
		return 0, errors.Wrapf(err, "serial no load failed")
	}
	gen.cur += gen.step
	return gen.cur, nil
}

func (gen *NumberGenerator) Next() (string, error) {
	n, err := gen.next()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(n, 10), nil
}

func (gen *NumberGenerator) Load() error {
	resp, err := gen.client.Incrby(gen.prefix, gen.step*gen.batchSize)
	if err != nil {
		return errors.Wrapf(err, "Load serial No. failed")
	}
	if resp != 0 && resp <= gen.start {
		if err := gen.client.Set(gen.prefix, strconv.FormatInt(gen.start, 10)); err != nil {
			return errors.Wrapf(err, "set gen.Start %d", gen.start)
		}
		return gen.Load()
	}

	if gen.max+gen.batchSize < resp && gen.max != 0 {
		gen.max = resp
		return gen.Load()
	}
	gen.start = resp - gen.batchSize*gen.step
	gen.max = resp
	//gen.cur += gen.step
	gen.cur = gen.start
	return nil
}
