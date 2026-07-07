package mailchimp

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	mailchimpSyncQueueKey = "mailchimp:sync"
	mailchimpSyncLockKey  = "mailchimp:sync:lock"
	maxRetries            = 3
	pollInterval          = 5 * time.Second
	lockDuration          = 30 * time.Second
)

type SyncAction string

const (
	ActionSubscribe   SyncAction = "subscribe"
	ActionUnsubscribe SyncAction = "unsubscribe"
	ActionUpdate      SyncAction = "update"
)

type SyncJob struct {
	SubscriberID string      `json:"subscriber_id"`
	Email        string      `json:"email"`
	Action       SyncAction  `json:"action"`
	MergeFields  MergeFields `json:"merge_fields,omitempty"`
	RetryCount   int         `json:"retry_count"`
	CreatedAt    time.Time   `json:"created_at"`
}

type SyncStats struct {
	TotalProcessed  int64     `json:"total_processed"`
	TotalFailed    int64     `json:"total_failed"`
	LastProcessedAt time.Time `json:"last_processed_at"`
	LastError      string    `json:"last_error,omitempty"`
}

type MailchimpSyncRepository interface {
	UpdateNewsletterSubscriberMailchimp(ctx context.Context, id, mailchimpID string, syncStatus string) error
}

type SyncService struct {
	client   *redis.Client
	mcClient *Client
	repo     MailchimpSyncRepository
	stopCh   chan struct{}
	logger   *logrus.Logger
}

func NewSyncService(client *redis.Client, mcClient *Client, repo MailchimpSyncRepository, logger *logrus.Logger) *SyncService {
	return &SyncService{
		client:   client,
		mcClient: mcClient,
		repo:     repo,
		stopCh:   make(chan struct{}),
		logger:   logger,
	}
}

func (s *SyncService) EnqueueSync(ctx context.Context, job SyncJob) error {
	if s.mcClient == nil || !s.mcClient.config.SyncEnabled {
		s.logger.Debug("Mailchimp sync is disabled, skipping enqueue")
		return nil
	}

	job.CreatedAt = time.Now()
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return s.client.LPush(ctx, mailchimpSyncQueueKey, string(data)).Err()
}

func (s *SyncService) EnqueueSubscribe(ctx context.Context, subscriberID, email string, mergeFields MergeFields) error {
	job := SyncJob{
		SubscriberID: subscriberID,
		Email:        email,
		Action:       ActionSubscribe,
		MergeFields:  mergeFields,
		RetryCount:   0,
	}
	return s.EnqueueSync(ctx, job)
}

func (s *SyncService) EnqueueUnsubscribe(ctx context.Context, subscriberID, email string) error {
	job := SyncJob{
		SubscriberID: subscriberID,
		Email:        email,
		Action:       ActionUnsubscribe,
		RetryCount:   0,
	}
	return s.EnqueueSync(ctx, job)
}

func (s *SyncService) EnqueueUpdate(ctx context.Context, subscriberID, email string, mergeFields MergeFields) error {
	job := SyncJob{
		SubscriberID: subscriberID,
		Email:        email,
		Action:       ActionUpdate,
		MergeFields:  mergeFields,
		RetryCount:   0,
	}
	return s.EnqueueSync(ctx, job)
}

func (s *SyncService) StartWorker(ctx context.Context) {
	if s.mcClient == nil || !s.mcClient.config.SyncEnabled {
		s.logger.Info("Mailchimp sync is disabled, worker not started")
		return
	}

	s.logger.Info("Starting Mailchimp sync worker")

	go func() {
		for {
			select {
			case <-ctx.Done():
				s.logger.Info("Mailchimp sync worker stopping (context cancelled)")
				return
			case <-s.stopCh:
				s.logger.Info("Mailchimp sync worker stopping")
				return
			default:
				s.processNextJob(ctx)
			}
		}
	}()
}

func (s *SyncService) Stop() {
	close(s.stopCh)
}

func (s *SyncService) processNextJob(ctx context.Context) {
	result, err := s.client.BRPop(ctx, pollInterval, mailchimpSyncQueueKey).Result()
	if err != nil {
		if err != redis.Nil {
			s.logger.WithError(err).Debug("BRPop returned error")
		}
		return
	}

	if len(result) < 2 {
		return
	}

	jobData := result[1]
	var job SyncJob
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		s.logger.WithError(err).Error("Failed to unmarshal sync job")
		return
	}

	s.processJob(ctx, job)
}

func (s *SyncService) processJob(ctx context.Context, job SyncJob) {
	var err error
	switch job.Action {
	case ActionSubscribe:
		err = s.handleSubscribe(ctx, job)
	case ActionUnsubscribe:
		err = s.handleUnsubscribe(ctx, job)
	case ActionUpdate:
		err = s.handleUpdate(ctx, job)
	default:
		s.logger.WithField("action", job.Action).Warn("Unknown sync action")
		return
	}

	if err != nil {
		s.handleRetry(ctx, job, err)
		return
	}

	if job.SubscriberID != "" {
		s.repo.UpdateNewsletterSubscriberMailchimp(ctx, job.SubscriberID, "", "synced")
	}

	s.logger.WithFields(logrus.Fields{
		"action":   job.Action,
		"email":    job.Email,
		"retry":    job.RetryCount,
	}).Debug("Sync job processed successfully")
}

func (s *SyncService) handleSubscribe(ctx context.Context, job SyncJob) error {
	email := job.Email
	firstName := job.MergeFields["FNAME"]
	lastName := job.MergeFields["LNAME"]
	delete(job.MergeFields, "FNAME")
	delete(job.MergeFields, "LNAME")

	tags := []string{}
	if source := job.MergeFields["SOURCE"]; source != "" {
		tags = append(tags, "source:"+source)
		delete(job.MergeFields, "SOURCE")
	}

	resp, err := s.mcClient.Subscribe(ctx, email, firstName, lastName, tags, job.MergeFields)
	if err != nil {
		return err
	}

	if job.SubscriberID != "" && resp != nil {
		s.repo.UpdateNewsletterSubscriberMailchimp(ctx, job.SubscriberID, resp.ID, "synced")
	}

	return nil
}

func (s *SyncService) handleUnsubscribe(ctx context.Context, job SyncJob) error {
	return s.mcClient.Unsubscribe(ctx, job.Email)
}

func (s *SyncService) handleUpdate(ctx context.Context, job SyncJob) error {
	email := job.Email
	delete(job.MergeFields, "FNAME")
	delete(job.MergeFields, "LNAME")

	resp, err := s.mcClient.UpdateSubscriber(ctx, email, nil, job.MergeFields)
	if err != nil {
		return err
	}

	if job.SubscriberID != "" && resp != nil {
		s.repo.UpdateNewsletterSubscriberMailchimp(ctx, job.SubscriberID, resp.ID, "synced")
	}

	return nil
}

func (s *SyncService) handleRetry(ctx context.Context, job SyncJob, err error) {
	if job.RetryCount >= maxRetries {
		s.logger.WithFields(logrus.Fields{
			"action":      job.Action,
			"email":       job.Email,
			"retry_count": job.RetryCount,
			"error":       err.Error(),
		}).Error("Max retries reached for sync job, marking as failed")

		if job.SubscriberID != "" {
			s.repo.UpdateNewsletterSubscriberMailchimp(ctx, job.SubscriberID, "", "failed")
		}
		return
	}

	job.RetryCount++
	data, _ := json.Marshal(job)

	backoff := time.Duration(job.RetryCount*job.RetryCount) * time.Second
	time.AfterFunc(backoff, func() {
		s.client.LPush(ctx, mailchimpSyncQueueKey, string(data))
	})

	s.logger.WithFields(logrus.Fields{
		"action":      job.Action,
		"email":       job.Email,
		"retry_count": job.RetryCount,
		"backoff":     backoff.String(),
		"error":       err.Error(),
	}).Warn("Sync job failed, requeuing with backoff")
}

func (s *SyncService) GetStats(ctx context.Context) (*SyncStats, error) {
	queueLen, err := s.client.LLen(ctx, mailchimpSyncQueueKey).Result()
	if err != nil {
		return nil, err
	}

	return &SyncStats{
		TotalProcessed: queueLen,
		TotalFailed:    0,
		LastProcessedAt: time.Time{},
	}, nil
}

func (s *SyncService) GetQueueLength(ctx context.Context) (int64, error) {
	return s.client.LLen(ctx, mailchimpSyncQueueKey).Result()
}

func (s *SyncService) ValidateWebhook(signature, timestamp, body string) bool {
	if s.mcClient == nil || s.mcClient.config.WebhookSecret == "" {
		return false
	}

	expectedSig := ComputeHMACSHA1(timestamp+body, s.mcClient.config.WebhookSecret)
	return signature == expectedSig
}

func ComputeHMACSHA1(data, secret string) string {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func ParseWebhookPayload(payload string) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return nil, err
	}
	return data, nil
}

func ExtractWebhookEventType(data map[string]interface{}) string {
	if typeVal, ok := data["type"].(string); ok {
		return typeVal
	}
	return ""
}

func ExtractWebhookEmail(data map[string]interface{}) string {
	if email, ok := data["data"].(map[string]interface{}); ok {
		if emailVal, ok := email["email"].(string); ok {
			return strings.ToLower(emailVal)
		}
	}
	return ""
}
