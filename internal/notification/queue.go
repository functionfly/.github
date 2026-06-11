package notification

import (
	"context"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/tracing"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	queueCapacity       = 1000
	pendingSweepInterval = 5 * time.Minute
	pendingSweepLimit    = 100
	maxRetries           = 5
	dlqSweepInterval    = 15 * time.Minute
	dlqMaxAge           = 24 * time.Hour
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
	metrics  *NotificationMetrics
	sweepTicker *time.Ticker
	dlqTicker   *time.Ticker
	retryCounts map[uuid.UUID]int
	retryMu     sync.Mutex
}

// NewQueue creates a new notification queue
func NewQueue(repo Repository, logger *logrus.Logger) *Queue {
	return &Queue{
		repo:         repo,
		logger:       logger,
		queue:        make(chan *Notification, queueCapacity),
		workers:      5,
		metrics:      GetNotificationMetrics(),
		retryCounts:  make(map[uuid.UUID]int),
	}
}

// Enqueue adds a notification to the queue (blocks up to 30s rather than dropping).
func (q *Queue) Enqueue(n *Notification) {
	if q.ctx == nil {
		q.logger.WithField("notification_id", n.ID).Warn("Queue not started; notification remains pending in DB")
		return
	}

	traceID := uuid.New().String()
	spanID := uuid.New().String()
	ctx := tracing.WithTraceContext(context.Background(), traceID, spanID, "")
	defer tracing.Finish(ctx)
	tracing.SetAttribute(ctx, "notification_id", n.ID.String())

	select {
	case q.queue <- n:
		q.metrics.RecordEnqueue()
		q.metrics.RecordQueueDepth(len(q.queue))
		q.metrics.RecordQueueSaturation(float64(len(q.queue)) / float64(queueCapacity) * 100)
		tracing.AddEvent(ctx, "enqueued", map[string]interface{}{"queue_depth": len(q.queue)})
		q.logger.WithField("notification_id", n.ID).Debug("Notification queued")
	case <-q.ctx.Done():
		return
	default:
		timer := time.NewTimer(30 * time.Second)
		defer timer.Stop()
		select {
		case q.queue <- n:
			q.metrics.RecordEnqueue()
			q.metrics.RecordQueueDepth(len(q.queue))
			q.metrics.RecordQueueSaturation(float64(len(q.queue)) / float64(queueCapacity) * 100)
			tracing.AddEvent(ctx, "enqueued_after_wait", map[string]interface{}{"queue_depth": len(q.queue)})
			q.logger.WithField("notification_id", n.ID).Debug("Notification queued after wait")
		case <-timer.C:
			q.metrics.RecordDropped()
			tracing.SetAttribute(ctx, "dropped", true)
			tracing.AddEvent(ctx, "dropped_saturation", nil)
			q.logger.WithField("notification_id", n.ID).Error("Notification queue saturated; requeuing in DB for pending sweep")
			if err := q.repo.RequeueNotification(ctx, n.ID); err != nil {
				q.logger.WithError(err).WithField("notification_id", n.ID).Error("Failed to requeue notification in DB")
			}
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

	q.sweepTicker = time.NewTicker(pendingSweepInterval)
	q.wg.Add(1)
	go q.pendingSweeper(dispatcher)

	q.dlqTicker = time.NewTicker(dlqSweepInterval)
	q.wg.Add(1)
	go q.dlqSweeper()

	q.logger.WithField("workers", q.workers).Info("Notification queue started")
}

// pendingSweeper periodically requeues pending notifications that may have been dropped
func (q *Queue) pendingSweeper(dispatcher *Dispatcher) {
	defer q.wg.Done()
	for {
		select {
		case <-q.sweepTicker.C:
			pendingNotifications, err := q.repo.GetPendingNotifications(q.ctx, pendingSweepInterval, pendingSweepLimit)
			if err != nil {
				q.logger.WithError(err).Error("Failed to get pending notifications for sweep")
				continue
			}
			for _, n := range pendingNotifications {
				select {
				case q.queue <- n:
					q.logger.WithField("notification_id", n.ID).Debug("Pending notification requeued by sweep")
				default:
					q.logger.WithField("notification_id", n.ID).Debug("Queue full, skipping pending sweep for now")
					break
				}
			}
		case <-q.ctx.Done():
			return
		}
	}
}

// dlqSweeper periodically cleans up old DLQ entries
func (q *Queue) dlqSweeper() {
	defer q.wg.Done()
	for {
		select {
		case <-q.dlqTicker.C:
			if err := q.repo.CleanupDeadLetterQueue(q.ctx, dlqMaxAge); err != nil {
				q.logger.WithError(err).Error("Failed to cleanup dead letter queue")
			} else {
				q.logger.Debug("Dead letter queue cleanup completed")
			}
		case <-q.ctx.Done():
			return
		}
	}
}

// Stop stops the queue processing
func (q *Queue) Stop() {
	q.stopOnce.Do(func() {
		if q.cancel != nil {
			q.cancel()
		}
		if q.sweepTicker != nil {
			q.sweepTicker.Stop()
		}
		if q.dlqTicker != nil {
			q.dlqTicker.Stop()
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
			q.metrics.RecordQueueDepth(len(q.queue))
			q.metrics.RecordQueueSaturation(float64(len(q.queue)) / float64(queueCapacity) * 100)

			traceID := uuid.New().String()
			spanID := uuid.New().String()
			ctx := tracing.WithTraceContext(q.ctx, traceID, spanID, "")
			tracing.SetAttribute(ctx, "notification_id", n.ID.String())
			tracing.SetAttribute(ctx, "notification_type", n.Type)

			start := time.Now()
			if err := dispatcher.Dispatch(ctx, n); err != nil {
				tracing.RecordError(ctx, err)
				q.logger.WithError(err).WithField("notification_id", n.ID).Error("Failed to dispatch notification")
				q.metrics.RecordDispatchError(err)

				q.retryMu.Lock()
				q.retryCounts[n.ID]++
				retries := q.retryCounts[n.ID]
				q.retryMu.Unlock()

				if retries >= maxRetries {
					q.logger.WithFields(logrus.Fields{
						"notification_id": n.ID,
						"retries":        retries,
					}).Error("Notification moved to DLQ after max retries")
					if dlqErr := q.repo.MoveToDeadLetterQueue(ctx, n.ID, err.Error()); dlqErr != nil {
						q.logger.WithError(dlqErr).WithField("notification_id", n.ID).Error("Failed to move notification to DLQ")
					}
					q.retryMu.Lock()
					delete(q.retryCounts, n.ID)
					q.retryMu.Unlock()
				} else {
					q.logger.WithFields(logrus.Fields{
						"notification_id": n.ID,
						"retry":          retries,
						"max_retries":    maxRetries,
					}).Warn("Notification failed, will retry")
				}
			} else {
				q.metrics.RecordDispatch(time.Since(start), true)
				q.retryMu.Lock()
				delete(q.retryCounts, n.ID)
				q.retryMu.Unlock()
			}
			tracing.Finish(ctx)
		case <-q.ctx.Done():
			return
		}
	}
}

// HealthCheck returns the current health status of the queue
func (q *Queue) HealthCheck() map[string]interface{} {
	queueLen := len(q.queue)
	saturation := float64(queueLen) / float64(queueCapacity) * 100

	status := "healthy"
	if saturation > 90 {
		status = "critical"
	} else if saturation > 75 {
		status = "degraded"
	}

	return map[string]interface{}{
		"status":           status,
		"queue_depth":      queueLen,
		"queue_capacity":   queueCapacity,
		"saturation_pct":   saturation,
		"workers":          q.workers,
		"running":          q.ctx != nil && q.ctx.Err() == nil,
	}
}
