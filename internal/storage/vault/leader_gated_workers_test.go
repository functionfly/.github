package vault

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunLeaderGatedWorkers_NoRedisStillCancels(t *testing.T) {
	// With a nil elector and no redis, the function runs each
	// worker directly. This test verifies the cleanup path: when
	// the context is cancelled, the worker functions should be
	// able to observe it.
	var ran atomic.Int32
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		RunLeaderGatedWorkers(ctx, nil, nil, func(ctx context.Context) {
			ran.Add(1)
			<-ctx.Done()
		})
		close(done)
	}()
	// Give the goroutine a moment to start.
	time.Sleep(50 * time.Millisecond)
	if ran.Load() != 1 {
		t.Fatalf("expected worker to run once, got %d", ran.Load())
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("RunLeaderGatedWorkers did not return after ctx cancel")
	}
}

func TestRunLeaderGatedWorkers_NilElector(t *testing.T) {
	// When elector is nil, all workers should run unconditionally.
	var ran atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	RunLeaderGatedWorkers(ctx, nil, nil,
		func(ctx context.Context) { ran.Add(1) },
		func(ctx context.Context) { ran.Add(1) },
	)
	time.Sleep(50 * time.Millisecond)
	if ran.Load() != 2 {
		t.Fatalf("expected 2 workers to run, got %d", ran.Load())
	}
}
