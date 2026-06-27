package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/billing"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type DisputeScheduler struct {
	cron              *cron.Cron
	disputeRepo       *storage.DisputeRepository
	disputeRespMgr    *billing.DisputeResponseManager
	notificationSvc   *notification.Service
	db                *sql.DB
	logger            *logrus.Logger
	stop              chan struct{}
}

func NewDisputeScheduler(
	disputeRepo *storage.DisputeRepository,
	disputeRespMgr *billing.DisputeResponseManager,
	notificationSvc *notification.Service,
	db *sql.DB,
) *DisputeScheduler {
	logger := logrus.WithField("scheduler", "dispute").Logger

	return &DisputeScheduler{
		cron:            cron.New(),
		disputeRepo:     disputeRepo,
		disputeRespMgr:  disputeRespMgr,
		notificationSvc: notificationSvc,
		db:              db,
		logger:          logger,
		stop:            make(chan struct{}),
	}
}

func (s *DisputeScheduler) Start() error {
	evidenceDeadlineCheck := disputeEnvStr("DISPUTE_EVIDENCE_DEADLINE_CRON", "0 */4 * * *")
	pendingReviewCheck := disputeEnvStr("DISPUTE_PENDING_REVIEW_CRON", "0 * * * *")
	fraudPatternAnalysis := disputeEnvStr("DISPUTE_FRAUD_ANALYSIS_CRON", "0 6 * * *")

	_, err := s.cron.AddFunc(evidenceDeadlineCheck, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		s.checkEvidenceDeadlines(ctx)
	})
	if err != nil {
		s.logger.WithError(err).Error("Failed to schedule evidence deadline check")
		return err
	}
	s.logger.Infof("Scheduled evidence deadline check: %s", evidenceDeadlineCheck)

	_, err = s.cron.AddFunc(pendingReviewCheck, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		s.checkPendingReviews(ctx)
	})
	if err != nil {
		s.logger.WithError(err).Error("Failed to schedule pending review check")
		return err
	}
	s.logger.Infof("Scheduled pending review check: %s", pendingReviewCheck)

	_, err = s.cron.AddFunc(fraudPatternAnalysis, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		s.analyzeFraudPatterns(ctx)
	})
	if err != nil {
		s.logger.WithError(err).Error("Failed to schedule fraud pattern analysis")
		return err
	}
	s.logger.Infof("Scheduled fraud pattern analysis: %s", fraudPatternAnalysis)

	s.cron.Start()
	s.logger.Info("Dispute scheduler started")

	go func() {
		<-s.stop
		s.cron.Stop()
		s.logger.Info("Dispute scheduler stopped")
	}()

	return nil
}

func (s *DisputeScheduler) Stop() {
	close(s.stop)
}

func (s *DisputeScheduler) checkEvidenceDeadlines(ctx context.Context) {
	disputes, err := s.disputeRepo.GetDisputesWithApproachingDeadline(ctx, 48*time.Hour)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get disputes with approaching deadlines")
		return
	}

	if len(disputes) == 0 {
		s.logger.Debug("No disputes with approaching evidence deadlines")
		return
	}

	s.logger.WithField("count", len(disputes)).Info("Found disputes with approaching evidence deadlines")

	for _, dispute := range disputes {
		if dispute.EvidenceDueBy == nil {
			continue
		}

		daysRemaining := int(time.Until(*dispute.EvidenceDueBy).Hours() / 24)
		if daysRemaining < 0 {
			daysRemaining = 0
		}

		adminUsers := s.getAdminUsers(ctx)
		if len(adminUsers) > 0 {
			err := s.notificationSvc.SendDisputeEvidenceDueSoon(ctx, adminUsers, dispute.StripeDisputeID, daysRemaining)
			if err != nil {
				s.logger.WithError(err).WithField("dispute_id", dispute.StripeDisputeID).Error("Failed to send evidence due soon notification")
			}
		}

		s.logger.WithFields(logrus.Fields{
			"dispute_id":     dispute.StripeDisputeID,
			"days_remaining": daysRemaining,
		}).Info("Evidence deadline approaching")
	}
}

func (s *DisputeScheduler) checkPendingReviews(ctx context.Context) {
	disputes, err := s.disputeRepo.GetStalePendingDisputes(ctx, 24*time.Hour)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get stale pending disputes")
		return
	}

	if len(disputes) == 0 {
		s.logger.Debug("No stale pending disputes")
		return
	}

	s.logger.WithField("count", len(disputes)).Info("Found stale pending disputes")

	for _, dispute := range disputes {
		adminUsers := s.getAdminUsers(ctx)
		if len(adminUsers) > 0 {
			err := s.notificationSvc.SendDisputePendingReminder(ctx, adminUsers, dispute.StripeDisputeID)
			if err != nil {
				s.logger.WithError(err).WithField("dispute_id", dispute.StripeDisputeID).Error("Failed to send pending reminder notification")
			}
		}

		s.logger.WithField("dispute_id", dispute.StripeDisputeID).Info("Dispute pending review for > 24 hours")
	}
}

func (s *DisputeScheduler) analyzeFraudPatterns(ctx context.Context) {
	windowDays := disputeEnvInt("DISPUTE_FRAUD_WINDOW_DAYS", 180)
	window := time.Duration(windowDays) * 24 * time.Hour

	disputes, err := s.disputeRepo.GetRecentDisputes(ctx, window)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get recent disputes for fraud analysis")
		return
	}

	if len(disputes) == 0 {
		s.logger.Debug("No recent disputes for fraud analysis")
		return
	}

	offenderMap := make(map[string][]*storage.PaymentDispute)
	for _, d := range disputes {
		if d.TenantID == nil {
			continue
		}
		key := d.TenantID.String()
		offenderMap[key] = append(offenderMap[key], d)
	}

	flaggedCount := 0
	for tenantID, disputeList := range offenderMap {
		if len(disputeList) >= 2 {
			flaggedCount++
			s.logger.WithFields(logrus.Fields{
				"tenant_id":           tenantID,
				"dispute_count":       len(disputeList),
				"total_disputed_cents": sumDisputeAmounts(disputeList),
			}).Warn("Repeat chargeback offender detected")

			s.flagRepeatOffender(ctx, tenantID, disputeList)
		}
	}

	if flaggedCount > 0 {
		s.logger.WithField("flagged_accounts", flaggedCount).Info("Fraud pattern analysis complete")
	}
}

func (s *DisputeScheduler) flagRepeatOffender(ctx context.Context, tenantID string, disputes []*storage.PaymentDispute) {
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		s.logger.WithError(err).WithField("tenant_id", tenantID).Error("Invalid tenant ID format")
		return
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO fraud_flags (id, tenant_id, flag_type, severity, details, created_at)
		VALUES (gen_random_uuid(), $1, 'repeat_chargeback_offender', 'high', $2::jsonb, NOW())
		ON CONFLICT DO NOTHING
	`, tenantUUID, toJSON(map[string]interface{}{
		"dispute_count":        len(disputes),
		"total_disputed_cents": sumDisputeAmounts(disputes),
		"first_dispute_at":     disputes[0].CreatedAt,
		"last_dispute_at":      disputes[len(disputes)-1].CreatedAt,
	}))
	if err != nil {
		s.logger.WithError(err).WithField("tenant_id", tenantID).Error("Failed to flag repeat offender")
	}
}

func (s *DisputeScheduler) getAdminUsers(ctx context.Context) []uuid.UUID {
	var userIDs []uuid.UUID
	rows, err := s.db.QueryContext(ctx, `
		SELECT id FROM users WHERE role IN ('admin', 'owner') AND email IS NOT NULL LIMIT 10
	`)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get admin users")
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err == nil {
			userIDs = append(userIDs, id)
		}
	}

	return userIDs
}

func sumDisputeAmounts(disputes []*storage.PaymentDispute) int64 {
	var sum int64
	for _, d := range disputes {
		sum += int64(d.AmountCents)
	}
	return sum
}

func disputeEnvStr(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func disputeEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func toJSON(m map[string]interface{}) string {
	if m == nil {
		return "{}"
	}
	s := "{"
	first := true
	for k, v := range m {
		if !first {
			s += ", "
		}
		first = false
		s += "\"" + k + "\": "
		switch val := v.(type) {
		case string:
			s += "\"" + val + "\""
		case int:
			s += strconv.Itoa(val)
		case int64:
			s += strconv.FormatInt(val, 10)
		case float64:
			s += strconv.FormatFloat(val, 'f', -1, 64)
		case bool:
			s += strconv.FormatBool(val)
		case time.Time:
			s += "\"" + val.Format(time.RFC3339) + "\""
		default:
			s += "\"" + fmt.Sprintf("%v", val) + "\""
		}
	}
	s += "}"
	return s
}
