package services

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type RevenueRecognitionWorker struct {
	repo *storage.RevenueRecognitionRepository
}

func NewRevenueRecognitionWorker(repo *storage.RevenueRecognitionRepository) *RevenueRecognitionWorker {
	return &RevenueRecognitionWorker{
		repo: repo,
	}
}

func (w *RevenueRecognitionWorker) Run(ctx context.Context) error {
	currentPeriod := time.Now().Format("2006-01")

	logrus.WithField("period", currentPeriod).Info("Starting revenue recognition processing")

	schedules, err := w.repo.GetAllUnrecognizedSchedules(ctx, currentPeriod)
	if err != nil {
		return fmt.Errorf("failed to get schedules: %w", err)
	}

	processed := 0
	failed := 0

	for _, schedule := range schedules {
		if err := w.processSchedule(ctx, schedule); err != nil {
			logrus.WithError(err).WithField("schedule_id", schedule.ID).Warn("Failed to process schedule")
			failed++
			continue
		}
		processed++
	}

	logrus.WithFields(logrus.Fields{
		"period":    currentPeriod,
		"processed": processed,
		"failed":    failed,
	}).Info("Revenue recognition processing completed")

	return nil
}

func (w *RevenueRecognitionWorker) processSchedule(ctx context.Context, schedule *storage.RevenueRecognitionSchedule) error {
	if schedule.IsRecognized {
		return nil
	}

	if err := w.repo.MarkScheduleRecognized(ctx, schedule.ID); err != nil {
		return fmt.Errorf("failed to mark schedule recognized: %w", err)
	}

	event := &storage.RevenueRecognitionEvent{
		ID:                    uuid.New(),
		TenantID:              schedule.TenantID,
		InvoiceID:             schedule.InvoiceID,
		EventType:             "scheduled_recognition",
		RevenueType:           schedule.RevenueType,
		GrossAmountCents:      schedule.AllocatedAmountCents,
		DeferredAmountCents:   0,
		RecognizedAmountCents: schedule.AllocatedAmountCents,
		EventDate:             time.Now(),
		ReportingPeriod:       time.Now().Format("2006-01"),
		ScheduleID:            &schedule.ID,
		Description:            fmt.Sprintf("Scheduled revenue recognition for period %s", schedule.RecognitionMonth),
	}

	return w.repo.CreateRecognitionEvent(ctx, event)
}

func (w *RevenueRecognitionWorker) ProcessTenant(ctx context.Context, tenantID uuid.UUID, period string) error {
	logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"period":    period,
	}).Info("Processing tenant revenue recognition")

	schedules, err := w.repo.GetRecognitionSchedulesByPeriod(ctx, tenantID, period)
	if err != nil {
		return fmt.Errorf("failed to get schedules: %w", err)
	}

	for _, schedule := range schedules {
		if schedule.IsRecognized {
			continue
		}
		if err := w.processSchedule(ctx, schedule); err != nil {
			logrus.WithError(err).WithField("schedule_id", schedule.ID).Warn("Failed to process schedule")
			continue
		}
	}

	return nil
}

func (w *RevenueRecognitionWorker) GetPendingSchedulesCount(ctx context.Context, period string) (int, error) {
	schedules, err := w.repo.GetAllUnrecognizedSchedules(ctx, period)
	if err != nil {
		return 0, err
	}
	return len(schedules), nil
}

func (w *RevenueRecognitionWorker) StartCron(ctx context.Context, period string) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Revenue recognition worker shutting down")
			return
		case <-ticker.C:
			if err := w.Run(ctx); err != nil {
				logrus.WithError(err).Error("Revenue recognition cron run failed")
			}
		}
	}
}