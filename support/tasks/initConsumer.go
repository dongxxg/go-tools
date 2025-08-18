/**
  @Author: dongxx
  @Since: 2025/8/18 17:10
  @Desc: //TODO
**/

package tasks

import (
	"context"
	"github.com/hibiken/asynq"
	"go-tools/support/logger"
	"go-tools/support/redis"
	"os"
	"strconv"
)

// 参数说明:
// queue 任务消费的队列, key:队列名, value:优先级, 必须要大于0, 如果<=0, 则任务不会被消费, 如果不指定则只会消费default队列的任务
// handles 任务处理函数, key:任务类型, value:处理函数
func InitAsynqConsumer(
	log *logger.Logger,
	redis *redis.RedisConfig,
	queue map[string]int,
	handles map[string]func(context.Context, *asynq.Task) error,
) {
	for s := range queue {
		if queue[s] > 0 {
			log.Info("asynq queue name %s, task will exec", s)
		} else {
			log.Warn("asynq queue name %s, task will not exec", s)
		}
	}
	dbNum, _ := strconv.Atoi(redis.Db)
	opt := asynq.RedisClientOpt{
		Addr:     redis.Addr,
		Password: redis.Password,
		DB:       dbNum,
	}
	if redis.UserName != "" {
		opt.Username = redis.UserName
	}
	asynqLog := adapterLog.AsynqLog{
		Logger: log,
	}
	srv := asynq.NewServer(
		opt,
		asynq.Config{
			Concurrency: 10,
			Logger:      asynqLog,
			Queues:      queue,
		},
	)

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	for i, v := range handles {
		mux.HandleFunc(i, v)
	}

	if err := srv.Run(mux); err != nil {
		logger.Error("asynq server run error: %s", err.Error())
		os.Exit(1)
	}
}
