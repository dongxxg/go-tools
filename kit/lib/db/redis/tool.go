package redis

import (
	"fmt"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/pkg/errors"
	errors2 "unitechs.com/cno/kit/errors"
	"unitechs.com/cno/kit/kctx"
	"unitechs.com/unios-dice/uni-base/lib/stores/redis"
)

// Hgetall  使用hscan 的方式获取代替hgetall
func Hgetall(ctx kctx.Ctx, db *redis.Redis, hashName string, batchSize int64, depth int) (map[string]string, error) {
	n := 0
	var cursor uint64
	out := make(map[string]string)

	for {
		vals, cur, err := db.Hscan(hashName, cursor, "*", batchSize)
		if err != nil {
			return nil, fmt.Errorf("hgetall by hscan failed, key %s err: %v", hashName, err)
		}
		if len(vals) != 0 {
			for i := 0; i < len(vals)/2; i++ {
				out[vals[i*2]] = vals[i*2+1]
			}
		}

		if cur == 0 {
			break
		}
		cursor = cur
		n++
		if n >= depth {
			break
		}
	}

	return out, nil
}

// Keys 因为很多现场不能使用keys命令， 因为使用scan来实现一下
func Keys(ctx kctx.Ctx, db *redis.Redis, match string, batchSize int64, maxDepth int) ([]string, error) {
	var curDepth int
	var cursor uint64
	out := gset.NewStrSet()

	for {
		lst, c, err := db.Scan(cursor, match, batchSize)
		if err != nil {
			return nil, fmt.Errorf("scan field err: %v", err)
		}
		for _, v := range lst {
			out.Add(v)
		}
		if c == 0 {
			return out.Slice(), nil
		}
		cursor = c

		if curDepth >= maxDepth {
			return out.Slice(), errors.Wrapf(errors2.ErrMaxRecursion{}, "is %d", maxDepth)
		}
		curDepth++
	}
}
