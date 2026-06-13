package vault

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// RunLeaderGatedWorkers runs each of the supplied worker functions
// in its own goroutine, but only on the goroutines that are elected
// leader. Each worker is gated by the elector's IsLeader() check
// and is started/stopped as the process gains/loses leadership.
//
// Single-instance deployments are always leader (Redis is unused).
// Multi-instance deployments elect one leader via SETNX; the
// non-leader goroutines return immediately and the leader keeps
// both workers running.
func RunLeaderGatedWorkers(
	ctx context.Context,
	elector *LeaderElector,
	logger *logrus.Logger,
	workers ...func(context.Context),
) {
	if elector == nil {
		// No election — run all workers unconditionally.
		for _, w := range workers {
			go w(ctx)
		}
		return
	}
	if elector.redis == nil {
		// No Redis — we can never become leader. Skip workers.
		logger.Warn("Leader elector has no Redis client; gated workers will not run")
		<-ctx.Done()
		return
	}
	go elector.Run(ctx)
	// Wait for first election attempt to settle, then start
	// workers whenever we become leader and stop when we lose it.
	started := make([]bool, len(workers))
	workerCtxs := make([]context.CancelFunc, len(workers))
	var mu sync.Mutex

	start := func(i int) {
		mu.Lock()
		defer mu.Unlock()
		if started[i] {
			return
		}
		wctx, cancel := context.WithCancel(ctx)
		workerCtxs[i] = cancel
		started[i] = true
		go workers[i](wctx)
	}
	stop := func(i int) {
		mu.Lock()
		defer mu.Unlock()
		if !started[i] {
			return
		}
		workerCtxs[i]()
		started[i] = false
	}

	for {
		select {
		case <-ctx.Done():
			mu.Lock()
			for i := range workers {
				if started[i] {
					workerCtxs[i]()
					started[i] = false
				}
			}
			mu.Unlock()
			return
		case <-time.After(1 * time.Second):
			if elector.IsLeader() {
				for i := range workers {
					start(i)
				}
			} else {
				for i := range workers {
					stop(i)
				}
			}
		}
	}
}
