package notification

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Queue manages the notification processing queue
type Queue struct {
	repo     Repository
	logger   *logrus.Logger
	queue    chan *Notification
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	workers  int
	stopOnce sync.Once
}

// NewQueue creates a new notification queue
func NewQueue(repo Repository, logger *logrus.Logger) *Queue {
	return &Queue{
		repo:    repo,
		logger:  logger,
		queue:   make(chan *Notification, 1000),
		workers: 5,
	}
}

// Enqueue adds a notification to the queue (blocks up to 30s rather than dropping).
func (q *Queue) Enqueue(n *Notification) {
	if q.ctx == nil {
		q.logger.WithField("notification_id", n.ID).Warn("Queue not started; notification remains pending in DB")
		return
	}
	select {
	case q.queue <- n:
		q.logger.WithField("notification_id", n.ID).Debug("Notification queued")
	case <-q.ctx.Done():
		return
	default:
		timer := time.NewTimer(30 * time.Second)
		defer timer.Stop()
		select {
		case q.queue <- n:
			q.logger.WithField("notification_id", n.ID).Debug("Notification queued after wait")
		case <-timer.C:
			q.logger.WithField("notification_id", n.ID).Warn("Notification queue saturated; will retry from DB pending sweep")
		case <-q.ctx.Done():
		}
	}
}

// Start begins processing the queue
func (q *Queue) Start(ctx context.Context, dispatcher *Dispatcher) {
	q.ctx, q.cancel = context.WithCancel(ctx)

	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(dispatcher)
	}

	q.logger.WithField("workers", q.workers).Info("Notification queue started")
}

// Stop stops the queue processing
func (q *Queue) Stop() {
	q.stopOnce.Do(func() {
		if q.cancel != nil {
			q.cancel()
		}
		close(q.queue)
		q.wg.Wait()
		q.logger.Info("Notification queue stopped")
	})
}

// worker processes notifications from the queue
func (q *Queue) worker(dispatcher *Dispatcher) {
	defer q.wg.Done()

	for {
		select {
		case n, ok := <-q.queue:
			if !ok {
				return
			}
			if err := dispatcher.Dispatch(q.ctx, n); err != nil {
				q.logger.WithError(err).WithField("notification_id", n.ID).Error("Failed to dispatch notification")
			}
		case <-q.ctx.Done():
			return
		}
	}
}
