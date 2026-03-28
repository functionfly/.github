package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PostgresDB methods: feedback and monitoring/notifications.

// Feedback operations
func (db *PostgresDB) CreateFeedback(feedback *Feedback) (*Feedback, error) {
	return db.feedbackRepository.CreateFeedback(feedback)
}

func (db *PostgresDB) GetFeedbackByID(id uuid.UUID) (*Feedback, error) {
	return db.feedbackRepository.GetFeedbackByID(id)
}

func (db *PostgresDB) GetFeedbackByUser(userID *uuid.UUID, userEmail *string, limit, offset int) ([]Feedback, error) {
	return db.feedbackRepository.GetFeedbackByUser(userID, userEmail, limit, offset)
}

func (db *PostgresDB) ListFeedback(limit, offset int, statusFilter *string, typeFilter *string) ([]Feedback, error) {
	return db.feedbackRepository.ListFeedback(limit, offset, statusFilter, typeFilter)
}

func (db *PostgresDB) UpdateFeedbackStatus(id uuid.UUID, status string) error {
	return db.feedbackRepository.UpdateFeedbackStatus(id, status)
}

func (db *PostgresDB) CreateFeedbackAttachment(attachment *FeedbackAttachment) (*FeedbackAttachment, error) {
	return db.feedbackRepository.CreateFeedbackAttachment(attachment)
}

func (db *PostgresDB) GetFeedbackAttachments(feedbackID uuid.UUID) ([]FeedbackAttachment, error) {
	return db.feedbackRepository.GetFeedbackAttachments(feedbackID)
}

func (db *PostgresDB) GetFeedbackAttachmentByID(attachmentID uuid.UUID) (*FeedbackAttachment, error) {
	return db.feedbackRepository.GetFeedbackAttachmentByID(attachmentID)
}

func (db *PostgresDB) GetFeedbackStats() (map[string]interface{}, error) {
	return db.feedbackRepository.GetFeedbackStats()
}

func (db *PostgresDB) GetFeedbackAnalytics() (map[string]interface{}, error) {
	return db.feedbackRepository.GetFeedbackAnalytics()
}

// Monitoring operations
func (db *PostgresDB) InsertPerformanceMetric(metric *PerformanceMetric) error {
	return db.monitoringRepository.InsertPerformanceMetric(metric)
}

func (db *PostgresDB) InsertAlert(alert *Alert) error {
	return db.monitoringRepository.InsertAlert(alert)
}

func (db *PostgresDB) InsertSystemHealthCheck(check *SystemHealthCheck) error {
	return db.monitoringRepository.InsertSystemHealthCheck(check)
}

func (db *PostgresDB) InsertMonitoringEvent(event *MonitoringEvent) error {
	return db.monitoringRepository.InsertMonitoringEvent(event)
}

func (db *PostgresDB) QueryMonitoringEvents(eventType string, tenantID *uuid.UUID, since time.Time, limit int) ([]*MonitoringEvent, error) {
	return db.monitoringRepository.QueryMonitoringEvents(eventType, tenantID, since, limit)
}

func (db *PostgresDB) UpdateAlertStatus(alert *Alert) error {
	return db.monitoringRepository.UpdateAlertStatus(alert)
}

func (db *PostgresDB) QueryPerformanceMetrics(metricType string, tenantID *uuid.UUID, since time.Time, limit int) ([]*PerformanceMetric, error) {
	return db.monitoringRepository.QueryPerformanceMetrics(metricType, tenantID, since, limit)
}

func (db *PostgresDB) QueryActiveAlerts(tenantID *uuid.UUID) ([]*Alert, error) {
	return db.monitoringRepository.QueryActiveAlerts(tenantID)
}

func (db *PostgresDB) QueryLatestSystemHealthChecks() (map[string]*SystemHealthCheck, error) {
	return db.monitoringRepository.QueryLatestSystemHealthChecks()
}

func (db *PostgresDB) PgNotify(channel, payload string) error {
	return db.monitoringRepository.PgNotify(channel, payload)
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
