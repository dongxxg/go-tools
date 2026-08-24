package threading

// 包装WorkGroup，开发者只需要关注：
// 1. 任务逻辑【job】
// 2. 任务数【workers】
type WorkerGroup struct {
	job     func()
	workers int
}

// NewWorkerGroup returns a WorkerGroup with given job and workers.
func NewWorkerGroup(job func(), workers int) WorkerGroup {
	return WorkerGroup{
		job:     job,
		workers: workers,
	}
}

// Start starts a WorkerGroup.
func (wg WorkerGroup) Start() {
	// 包装了sync.WaitGroup
	group := NewRoutineGroup()
	for i := 0; i < wg.workers; i++ {
		// 内部维护了 wg.Add(1) wg.Done()
		// 同时也是 goroutine 安全模式下进行的
		group.RunSafe(wg.job)
	}
	group.Wait()
}
