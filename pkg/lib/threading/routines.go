package threading

import (
	"bytes"
	"context"
	"runtime"
	"strconv"

	"unitechs.com/unios-dice/uni-base/lib/rescue"
)

// 解决裸goroutine的几个问题：
// 1. 如果协程异常退出，无法追踪异常栈
// 2. 某个异常请求触发 panic，应该做故障隔离，而不是整个进程退出，容易被攻击
func GoSafe(fn func()) {
	go RunSafe(fn)
}

// GoSafeCtx runs the given fn using another goroutine, recovers if fn panics with ctx.
func GoSafeCtx(ctx context.Context, fn func()) {
	go RunSafeCtx(ctx, fn)
}

// RoutineId is only for debug, never use it in production.
func RoutineId() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	// if error, just return 0
	n, _ := strconv.ParseUint(string(b), 10, 64)

	return n
}

// RunSafe runs the given fn, recovers if fn panics.
func RunSafe(fn func()) {
	defer rescue.Recover()

	fn()
}

// RunSafeCtx runs the given fn, recovers if fn panics with ctx.
func RunSafeCtx(ctx context.Context, fn func()) {
	defer rescue.RecoverCtx(ctx)

	fn()
}
