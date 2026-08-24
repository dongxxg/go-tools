package redis

import (
	"context"
	"github.com/dongxxg/go-tools/lib/mapping"
	"github.com/dongxxg/go-tools/lib/timex"
	"github.com/dongxxg/go-tools/log"
	"github.com/go-redis/redis/v8"
	"strings"
	"time"
)

type SlowCallHook struct{}

func (s SlowCallHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	ctx = context.WithValue(ctx, "start", time.Now())
	return ctx, nil
}
func (s SlowCallHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	data := ctx.Value("start")
	switch start := data.(type) {
	case time.Duration:
		duration := timex.Since(start)
		if duration > slowThreshold {
			var buf strings.Builder
			for i, arg := range cmd.Args() {
				if i > 0 {
					buf.WriteByte(' ')
				}
				buf.WriteString(mapping.Repr(arg))
			}
			log.WithDuration(duration).Infof("[REDIS] slowcall on executing: %s", buf.String())
		}
	default:

	}
	return nil
}
func (s SlowCallHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	ctx = context.WithValue(ctx, "start", time.Now())
	return ctx, nil
}
func (s SlowCallHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	data := ctx.Value("start")
	switch start := data.(type) {
	case time.Duration:
		duration := timex.Since(start)
		if duration > slowThreshold {
			var buf strings.Builder
			for _, cmd := range cmds {
				for i, arg := range cmd.Args() {
					if i > 0 {
						buf.WriteByte(' ')
					}
					buf.WriteString(mapping.Repr(arg))
				}
			}
			log.WithDuration(duration).Infof("[REDIS] slowcall on executing: %s", buf.String())
		}
	default:

	}
	return nil
}