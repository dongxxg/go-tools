package threading

import (
	"github.com/dongxxg/go-tools/lib/lang"
	"github.com/dongxxg/go-tools/lib/rescue"
)

type TaskRunner struct {
	limitChan chan lang.PlaceholderType
}

// NewTaskRunner 通过令牌机制，控制最大并发执行数
func NewTaskRunner(concurrency int) *TaskRunner {
	return &TaskRunner{
		limitChan: make(chan lang.PlaceholderType, concurrency),
	}
}

// Schedule schedules a task to run under concurrency control.
func (rp *TaskRunner) Schedule(task func()) {
	rp.limitChan <- lang.Placeholder

	go func() {
		defer rescue.Recover(func() {
			<-rp.limitChan
		})

		task()
	}()
}