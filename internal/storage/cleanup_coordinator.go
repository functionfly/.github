package storage

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// CleanupCoordinator coordinates multiple cleanup jobs to avoid conflicts
// during heavy write load by staggering their execution times.
type CleanupCoordinator struct {
	mu         sync.Mutex
	jobs       []*CleanupJob
	logger     *logrus.Logger
	maxJitter  time.Duration // Maximum random jitter to add to stagger jobs
}

// CleanupJob represents a cleanup task that can be coordinated
type CleanupJob struct {
	Name      string
	Interval  time.Duration
	Retention time.Duration
	RunFunc   func(ctx context.Context, retention time.Duration) error
}

// NewCleanupCoordinator creates a new cleanup coordinator
func NewCleanupCoordinator() *CleanupCoordinator {
	return &CleanupCoordinator{
		logger:    logrus.New(),
		maxJitter: 5 * time.Minute, // Default 5 minute max jitter
	}
}

// SetMaxJitter sets the maximum random jitter to add to job staggering
func (c *CleanupCoordinator) SetMaxJitter(jitter time.Duration) {
	c.maxJitter = jitter
}

// RegisterJob registers a cleanup job with the coordinator
func (c *CleanupCoordinator) RegisterJob(job *CleanupJob) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.jobs = append(c.jobs, job)
}

// StartAll starts all registered cleanup jobs with staggered timing
func (c *CleanupCoordinator) StartAll(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, job := range c.jobs {
		// Calculate initial delay to stagger jobs across the interval
		// This prevents all jobs from running at the same time
		jitter := time.Duration(rand.Int63n(int64(c.maxJitter)))
		delay := (job.Interval / time.Duration(len(c.jobs))) * time.Duration(i)
		initialDelay := delay + jitter

		c.logger.WithFields(logrus.Fields{
			"job":           job.Name,
			"interval":      job.Interval.String(),
			"initial_delay": initialDelay.String(),
		}).Info("Starting coordinated cleanup job")

		go func(j *CleanupJob, delay time.Duration) {
			// Wait for initial stagger delay
			time.Sleep(delay)

			// Run initial cleanup immediately
			if err := j.RunFunc(ctx, j.Retention); err != nil {
				c.logger.WithError(err).Errorf("Initial %s cleanup failed", j.Name)
			}

			// Set up periodic cleanup
			ticker := time.NewTicker(j.Interval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if err := j.RunFunc(ctx, j.Retention); err != nil {
						c.logger.WithError(err).Errorf("Periodic %s cleanup failed", j.Name)
					}
				case <-ctx.Done():
					c.logger.Infof("%s cleanup routine stopping", j.Name)
					return
				}
			}
		}(job, initialDelay)
	}
}

// CleanupStats holds statistics about cleanup operations
type CleanupStats struct {
	Name         string
	LastRunAt    time.Time
	LastDuration time.Duration
	LastDeleted  int64
	LastError    error
}

// GetStats returns statistics for all cleanup jobs
func (c *CleanupCoordinator) GetStats() map[string]CleanupStats {
	c.mu.Lock()
	defer c.mu.Unlock()

	stats := make(map[string]CleanupStats)
	for _, job := range c.jobs {
		// This would need to be enhanced to track actual stats per job
		stats[job.Name] = CleanupStats{
			Name: job.Name,
		}
	}
	return stats
}
