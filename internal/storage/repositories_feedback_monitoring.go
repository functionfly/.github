package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PostgresDB methods: feedback and monitoring/notifications.

// Feedback operations
func (db *PostgresDB) CreateFeedback(ctx context.Context, feedback *Feedback) (*Feedback, error) {
	return db.feedbackRepository.CreateFeedback(ctx, feedback)
}

func (db *PostgresDB) GetFeedbackByID(ctx context.Context, id uuid.UUID) (*Feedback, error) {
	return db.feedbackRepository.GetFeedbackByID(ctx, id)
}

func (db *PostgresDB) GetFeedbackByUser(ctx context.Context, userID *uuid.UUID, userEmail *string, limit, offset int) ([]Feedback, error) {
	return db.feedbackRepository.GetFeedbackByUser(ctx, userID, userEmail, limit, offset)
}

func (db *PostgresDB) ListFeedback(ctx context.Context, limit, offset int, statusFilter *string, typeFilter *string) ([]Feedback, error) {
	return db.feedbackRepository.ListFeedback(ctx, limit, offset, statusFilter, typeFilter)
}

func (db *PostgresDB) UpdateFeedbackStatus(id uuid.UUID, status string) error {
	return db.feedbackRepository.UpdateFeedbackStatus(context.Background(), id, status)
}

func (db *PostgresDB) CreateFeedbackAttachment(ctx context.Context, attachment *FeedbackAttachment) (*FeedbackAttachment, error) {
	return db.feedbackRepository.CreateFeedbackAttachment(ctx, attachment)
}

func (db *PostgresDB) GetFeedbackAttachments(ctx context.Context, feedbackID uuid.UUID) ([]FeedbackAttachment, error) {
	return db.feedbackRepository.GetFeedbackAttachments(ctx, feedbackID)
}

func (db *PostgresDB) GetFeedbackAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*FeedbackAttachment, error) {
	return db.feedbackRepository.GetFeedbackAttachmentByID(ctx, attachmentID)
}

func (db *PostgresDB) GetFeedbackStats(ctx context.Context) (map[string]interface{}, error) {
	return db.feedbackRepository.GetFeedbackStats(ctx)
}

func (db *PostgresDB) GetFeedbackAnalytics(ctx context.Context) (map[string]interface{}, error) {
	return db.feedbackRepository.GetFeedbackAnalytics(ctx)
}

// Monitoring operations
func (db *PostgresDB) InsertPerformanceMetric(metric *PerformanceMetric) error {
	return db.monitoringRepository.InsertPerformanceMetric(context.Background(), metric)
}

func (db *PostgresDB) InsertAlert(alert *Alert) error {
	return db.monitoringRepository.InsertAlert(context.Background(), alert)
}

func (db *PostgresDB) InsertSystemHealthCheck(check *SystemHealthCheck) error {
	return db.monitoringRepository.InsertSystemHealthCheck(context.Background(), check)
}

func (db *PostgresDB) InsertMonitoringEvent(event *MonitoringEvent) error {
	return db.monitoringRepository.InsertMonitoringEvent(context.Background(), event)
}

func (db *PostgresDB) QueryMonitoringEvents(ctx context.Context, eventType string, tenantID *uuid.UUID, since time.Time, limit int) ([]*MonitoringEvent, error) {
	return db.monitoringRepository.QueryMonitoringEvents(ctx, eventType, tenantID, since, limit)
}

func (db *PostgresDB) UpdateAlertStatus(alert *Alert) error {
	return db.monitoringRepository.UpdateAlertStatus(context.Background(), alert)
}

func (db *PostgresDB) QueryPerformanceMetrics(ctx context.Context, metricType string, tenantID *uuid.UUID, since time.Time, limit int) ([]*PerformanceMetric, error) {
	return db.monitoringRepository.QueryPerformanceMetrics(ctx, metricType, tenantID, since, limit)
}

func (db *PostgresDB) QueryActiveAlerts(ctx context.Context, tenantID *uuid.UUID) ([]*Alert, error) {
	return db.monitoringRepository.QueryActiveAlerts(ctx, tenantID)
}

func (db *PostgresDB) QueryLatestSystemHealthChecks(ctx context.Context) (map[string]*SystemHealthCheck, error) {
	return db.monitoringRepository.QueryLatestSystemHealthChecks(ctx)
}

func (db *PostgresDB) PgNotify(channel, payload string) error {
	return db.monitoringRepository.PgNotify(context.Background(), channel, payload)
}

func (db *PostgresDB) PgListen(ctx context.Context, channel string) error {
	if db == nil || db.monitoringRepository == nil {
		return fmt.Errorf("monitoring repository not initialized")
	}
	return db.monitoringRepository.PgListen(ctx, channel)
}

func (db *PostgresDB) PgWaitForNotification(ctx context.Context) (*PgNotification, error) {
	return db.monitoringRepository.PgWaitForNotification(ctx)
}

func (db *PostgresDB) GetDatabaseHealthMetrics(ctx context.Context) (map[string]interface{}, error) {
	return db.monitoringRepository.GetDatabaseHealthMetrics(ctx)
}

func (db *PostgresDB) StoreDatabaseMetrics(ctx context.Context, metrics map[string]interface{}) error {
	return db.monitoringRepository.StoreDatabaseMetrics(ctx, metrics)
}

func (db *PostgresDB) QueryDatabaseMetrics(ctx context.Context, metricType string, since time.Time, limit int) ([]*DatabaseMetric, error) {
	return db.monitoringRepository.QueryDatabaseMetrics(ctx, metricType, since, limit)
}
