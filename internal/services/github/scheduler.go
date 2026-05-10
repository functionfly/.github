package github

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type ImportJob struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	UserID        uuid.UUID
	ImportID      uuid.UUID
	RepoFullName  string
	Branch        string
	CommitSHA     string
	Priority      int
	ScheduledAt   time.Time
	StartedAt     *time.Time
	CompletedAt   *time.Time
	Status        string
}

type ImportScheduler struct {
	queue         chan *ImportJob
	maxConcurrent int
	tokensPerHour int
	costPerImport int
	redisClient   *redis.Client
	logger        *logrus.Logger
	activeJobs    map[uuid.UUID]*ImportJob
	mu            sync.Mutex
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

func NewImportScheduler(redisClient *redis.Client, logger *logrus.Logger, maxConcurrent int) *ImportScheduler {
	s := &ImportScheduler{
		queue:         make(chan *ImportJob, 1000),
		maxConcurrent: maxConcurrent,
		tokensPerHour: 5000,
		costPerImport: 10,
		redisClient:   redisClient,
		logger:        logger,
		activeJobs:    make(map[uuid.UUID]*ImportJob),
		stopCh:        make(chan struct{}),
	}
	return s
}

func (s *ImportScheduler) Start() {
	s.wg.Add(1)
	go s.processQueue()
	s.wg.Add(1)
	go s.tokenRefiller()
}

func (s *ImportScheduler) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}

func (s *ImportScheduler) Schedule(job *ImportJob) error {
	select {
	case s.queue <- job:
		s.logger.WithFields(logrus.Fields{
			"job_id":   job.ID,
			"import_id": job.ImportID,
		}).Info("Job scheduled")
		return nil
	default:
		return ErrQueueFull
	}
}

func (s *ImportScheduler) processQueue() {
	defer s.wg.Done()

	for {
		select {
		case <-s.stopCh:
			return
		case job := <-s.queue:
			s.processJob(job)
		}
	}
}

func (s *ImportScheduler) processJob(job *ImportJob) {
	s.mu.Lock()
	if len(s.activeJobs) >= s.maxConcurrent {
		s.queue <- job
		s.mu.Unlock()
		return
	}

	tokens, err := s.getAvailableTokens(context.Background(), job.TenantID)
	if err != nil || tokens < int64(s.costPerImport) {
		delay := 30 * time.Second
		s.logger.WithFields(logrus.Fields{
			"job_id":      job.ID,
			"tenant_id":   job.TenantID,
			"delay_secs":  delay.Seconds(),
		}).Warn("Rate limited, re-scheduling")

		time.Sleep(delay)
		s.queue <- job
		s.mu.Unlock()
		return
	}

	job.Status = "running"
	now := time.Now()
	job.StartedAt = &now
	s.activeJobs[job.ID] = job
	s.mu.Unlock()

	s.consumeTokens(context.Background(), job.TenantID, int64(s.costPerImport))

	s.logger.WithFields(logrus.Fields{
		"job_id":   job.ID,
		"import_id": job.ImportID,
	}).Info("Processing import job")

	go func() {
		s.executeJob(job)
	}()
}

func (s *ImportScheduler) executeJob(job *ImportJob) {
	defer func() {
		s.mu.Lock()
		delete(s.activeJobs, job.ID)
		now := time.Now()
		job.CompletedAt = &now
		s.mu.Unlock()
	}()

	time.Sleep(2 * time.Second)

	completed := &time.Time{}
	_ = completed
}

func (s *ImportScheduler) tokenRefiller() {
	defer s.wg.Done()
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.refillTokens(context.Background())
		}
	}
}

func (s *ImportScheduler) getAvailableTokens(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	if s.redisClient == nil {
		return int64(s.tokensPerHour), nil
	}

	key := s.tokenKey(tenantID)
	tokens, err := s.redisClient.Get(ctx, key).Int64()
	if err == redis.Nil {
		return int64(s.tokensPerHour), nil
	}
	if err != nil {
		return 0, err
	}
	return tokens, nil
}

func (s *ImportScheduler) consumeTokens(ctx context.Context, tenantID uuid.UUID, cost int64) error {
	if s.redisClient == nil {
		return nil
	}

	key := s.tokenKey(tenantID)
	return s.redisClient.DecrBy(ctx, key, cost).Err()
}

func (s *ImportScheduler) refillTokens(ctx context.Context) {
	if s.redisClient == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	pattern := "import:tokens:*"
	iter := s.redisClient.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		s.redisClient.Set(ctx, key, s.tokensPerHour, time.Hour)
	}
}

func (s *ImportScheduler) tokenKey(tenantID uuid.UUID) string {
	return "import:tokens:" + tenantID.String()
}

func (s *ImportScheduler) GetActiveJobsCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.activeJobs)
}

func (s *ImportScheduler) GetQueueLength() int {
	return len(s.queue)
}

type TenantPlan int

const (
	TierFree TenantPlan = iota
	TierStarter
	TierPro
	TierEnterprise
)

func GetTenantPlan(tenantID uuid.UUID) TenantPlan {
	return TierPro
}

type RateLimitConfig struct {
	MaxConcurrent int
	TokensPerHour int
	CostPerImport int
}

func GetRateLimitConfig(plan TenantPlan) RateLimitConfig {
	switch plan {
	case TierFree:
		return RateLimitConfig{
			MaxConcurrent: 2,
			TokensPerHour: 1000,
			CostPerImport: 20,
		}
	case TierStarter:
		return RateLimitConfig{
			MaxConcurrent: 5,
			TokensPerHour: 3000,
			CostPerImport: 15,
		}
	case TierPro:
		return RateLimitConfig{
			MaxConcurrent: 10,
			TokensPerHour: 5000,
			CostPerImport: 10,
		}
	case TierEnterprise:
		return RateLimitConfig{
			MaxConcurrent: 50,
			TokensPerHour: 10000,
			CostPerImport: 5,
		}
	default:
		return RateLimitConfig{
			MaxConcurrent: 2,
			TokensPerHour: 1000,
			CostPerImport: 20,
		}
	}
}

var ErrQueueFull = &SchedulerError{Message: "import queue is full, please retry later"}

type SchedulerError struct {
	Message string
}

func (e *SchedulerError) Error() string {
	return e.Message
}