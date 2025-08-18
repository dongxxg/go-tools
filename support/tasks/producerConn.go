/**
  @Author: dongxx
  @Since: 2025/8/18 17:14
  @Desc: //TODO
**/

package tasks

import (
	"encoding/json"
	"errors"
	"go-tools/support/redis"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
)

var asy *asynq.Client
var asyClient *AsynqProducer

type AsynqProducer struct {
	AsynqProducer *asynq.Client
}

func initAsynqProducer(redis *redis.RedisConfig) {
	dbNum, _ := strconv.Atoi(redis.Db)
	opt := asynq.RedisClientOpt{
		Addr:     redis.Addr,
		Password: redis.Password,
		DB:       dbNum,
	}
	if redis.UserName != "" {
		opt.Username = redis.UserName
	}
	asy = asynq.NewClient(opt)
	asyClient = &AsynqProducer{asy}
}

func GetAsynqProducer() (*AsynqProducer, error) {
	if asyClient == nil {
		if &config.Config.Redis == nil {
			return nil, errors.New("redis config is nil")
		}
		initAsynqProducer(&config.Config.Redis)
	}
	return asyClient, nil
}

// 其它参数待补充
type TaskConfig struct {
	Queue     string        // Queue 队列名称，为空则使用默认队列
	RetryNum  int           // RetryNum 重试次数，<=0 表示0重试
	TimeOut   time.Duration // TimeOut 超时时间，0< timeOut <=24小时 则设置最大限制1天
	ProcessIn time.Duration // 任务的延迟执行时间（相对于当前时间）
}

func (t *TaskConfig) parse() []asynq.Option {
	var opt []asynq.Option

	if t.Queue != "" {
		opt = append(opt, asynq.Queue(t.Queue))
	}

	opt = append(opt, asynq.MaxRetry(t.RetryNum))

	if t.TimeOut > 0 && t.TimeOut <= time.Hour*24 {
		opt = append(opt, asynq.Timeout(t.TimeOut))
	} else {
		opt = append(opt, asynq.Timeout(time.Hour*24))
	}

	if t.ProcessIn > 0 {
		opt = append(opt, asynq.ProcessIn(t.ProcessIn))
	}
	return opt
}

// Produce 投入任务, 和基本参数校验
// task 任务参数
// config 任务配置
func (p *AsynqProducer) Produce(
	taskType string,
	info interface{},
	config *TaskConfig,
) (*asynq.TaskInfo, error) {
	if p == nil || p.AsynqProducer == nil {
		return nil, errors.New("asynq producer is nil")
	}
	payload, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}

	task := asynq.NewTask(taskType, payload)

	defaultOpt := TaskConfig{
		Queue:     "",
		RetryNum:  0,
		TimeOut:   0,
		ProcessIn: 0,
	}

	if config != nil {
		if len(config.Queue) > 0 {
			defaultOpt.Queue = config.Queue
		}
		if config.RetryNum > 0 {
			defaultOpt.RetryNum = config.RetryNum
		}
		if config.TimeOut > 0 {
			defaultOpt.TimeOut = config.TimeOut
		}
		if config.ProcessIn > 0 {
			defaultOpt.ProcessIn = config.ProcessIn
		}
	}

	opt := defaultOpt.parse()

	return p.AsynqProducer.Enqueue(
		task,
		opt...,
	)

}
