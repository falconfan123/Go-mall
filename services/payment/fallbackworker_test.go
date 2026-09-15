package main

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// ctx 取消后当前轮跑完即退出（不立即中断）——design 决策 1 优雅停机语义。
func TestRunLoopCompletesCurrentTickBeforeExit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var runs atomic.Int64
	var finished atomic.Bool

	done := make(chan struct{})
	go func() {
		runLoop(ctx, time.Hour, "test", func(context.Context) {
			// 首轮模拟长任务：cancel 在任务中途到达
			time.Sleep(200 * time.Millisecond)
			runs.Add(1)
			finished.Store(true)
		})
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("loop did not exit after cancel")
	}
	if !finished.Load() || runs.Load() != 1 {
		t.Fatalf("current tick must complete before exit: runs=%d finished=%v", runs.Load(), finished.Load())
	}
}

// panic 不外泄：本轮回收（error 日志），循环继续下一 tick。
func TestRunLoopRecoversFromPanic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var panics atomic.Int64
	var healthy atomic.Int64

	done := make(chan struct{})
	go func() {
		runLoop(ctx, 20*time.Millisecond, "test-panic", func(context.Context) {
			if panics.Load() == 0 {
				panics.Add(1)
				panic("tick panic (expected in test)")
			}
			healthy.Add(1)
		})
		close(done)
	}()

	time.Sleep(120 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("loop did not exit")
	}
	if panics.Load() != 1 || healthy.Load() < 1 {
		t.Fatalf("panic must not kill loop: panics=%d healthy=%d", panics.Load(), healthy.Load())
	}
}
