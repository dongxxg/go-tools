// 任务号    日期     开发者      描述

package kredis

import (
	"fmt"

	"github.com/pkg/errors"
	"unitechs.com/cno/kit/kctx"
)

// ISerialRepo 通过该接口实现对缓存库的读写 如redis 的读写
type ISerialRepo interface {
	GenSerialNoWithLua(ctx kctx.Ctx, key string, start, interval int64) (int64, error)
}

// SerialNoRepo 全局序列号No
type SerialNoRepo struct {
	redisRepo ISerialRepo
}

func NewSerialNoRepo(repo ISerialRepo) *SerialNoRepo {
	return &SerialNoRepo{redisRepo: repo}
}
func NewSerialNo(ctx kctx.Ctx, Type, host, pwd string, db int) (*SerialNoRepo, error) {
	repo, err := NewRedisRepo(ctx, Type, host, pwd, db)
	if err != nil {
		return nil, errors.Wrapf(err, "new redis repo failed err: %v", err)
	}

	return NewSerialNoRepo(repo), nil
}

func (s SerialNoRepo) Get(ctx kctx.Ctx, ruleId, templateId, dstField string, start, interval int64) (int64, error) {
	key := GenSerialNoKey(ruleId, templateId, dstField)
	return s.redisRepo.GenSerialNoWithLua(ctx, key, start, interval)
}

// GenSerialNoKey 生成序列号在redis中存储的key
func GenSerialNoKey(ruleId, templateId, dstField string) string {
	if dstField == "" {
		return fmt.Sprintf("%s:%s", ruleId, templateId)
	}
	return fmt.Sprintf("%s:%s:%s", ruleId, templateId, dstField)
}
