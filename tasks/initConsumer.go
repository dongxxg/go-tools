/**
  @Author: dongxx
  @Since: 2025/8/18 17:10
  @Desc: //TODO
**/

package tasks

import (
	"context"
	"os"
	"strconv"

	"github.com/dongxxg/go-tools/log"
	"github.com/hibiken/asynq"
)

// RedisConfig holds the minimal redis connection info for building an asynq client.
type RedisConfig struct {
	Addr     string
	Password string
	Db       string
	UserName string
}

// asynqLogger adapts uni-base log to asynq's Logger interface.
type asynqLogger struct{}

func (asynqLogger) Debug(args ...interface{}) { log.Debug(args...) }
func (asynqLogger) Info(args ...interface{})  { log.Info(args...) }
func (asynqLogger) Warn(args ...interface{})  { log.Warn(args...) }
func (asynqLogger) Error(args ...interface{}) { log.Error(args...) }
func (asynqLogger) Fatal(args ...interface{}) { log.Fatal(args...) }

// 参数说明:
// queue 任务消费的队列, key:队列名, value:优先级, 必须要大于0, 如果<=0, 则任务不会被消费, 如果不指定则只会消费default队列的任务
// handles 任务处理函数, key:任务类型, value:处理函数
func InitAsynqConsumer(
	redis *RedisConfig,
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
	srv := asynq.NewServer(
		opt,
		asynq.Config{
			Concurrency: 10,
			Logger:      asynqLogger{},
			Queues:      queue,
		},
	)

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	for i, v := range handles {
		mux.HandleFunc(i, v)
	}

	if err := srv.Run(mux); err != nil {
		log.Error("asynq server run error: %s", err.Error())
		os.Exit(1)
	}
}