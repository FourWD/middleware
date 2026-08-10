package infra

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"
)

// RunLoop calls fn immediately and then every tick until ctx is done. A panic
// inside fn is recovered and logged, so one bad iteration cannot take down the
// worker goroutine that owns the loop.
//
// The first call happens before the first sleep: a worker registered at boot
// does its work right away rather than after one tick of silence.
func RunLoop(ctx context.Context, log Scoped, tick time.Duration, fn func(context.Context)) {
	for {
		runLoopOnce(ctx, log, fn)
		select {
		case <-ctx.Done():
			return
		case <-time.After(tick):
		}
	}
}

func runLoopOnce(ctx context.Context, log Scoped, fn func(context.Context)) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(ctx, fmt.Errorf("panic: %v", r), "WORKER_PANIC_RECOVERED",
				map[string]any{"stack": string(debug.Stack()), "severity": "critical"},
				WithOperation("run_loop"),
				WithLogKind(LogKindError))
		}
	}()
	fn(ctx)
}
