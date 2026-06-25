package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type FounderAnniversaryScheduler struct {
	cron        *cron.Cron
	userRepo    storage.Repository
	notifySvc   *notification.Service
	redisClient *redis.Client
	db          *sql.DB
	logger      *logrus.Logger

	CheckInterval string
}

type FounderAnniversary struct {
	UserID        uuid.UUID `json:"user_id"`
	FounderNumber int       `json:"founder_number"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	YearsActive   int       `json:"years_active"`
	FounderSince  time.Time `json:"founder_since"`
}

func NewFounderAnniversaryScheduler(
	userRepo storage.Repository,
	notifySvc *notification.Service,
	redisClient *redis.Client,
	db *sql.DB,
) *FounderAnniversaryScheduler {
	return &FounderAnniversaryScheduler{
		cron:           cron.New(cron.WithSeconds()),
		userRepo:       userRepo,
		notifySvc:       notifySvc,
		redisClient:     redisClient,
		db:             db,
		logger:         logrus.New(),
		CheckInterval:  "0 0 * * *",
	}
}

func (s *FounderAnniversaryScheduler) WithLogger(logger *logrus.Logger) *FounderAnniversaryScheduler {
	s.logger = logger
	return s
}

func (s *FounderAnniversaryScheduler) Start(ctx context.Context) error {
	if err := s.initAnniversaryTable(ctx); err != nil {
		s.logger.WithError(err).Warn("Failed to init anniversary table, continuing")
	}

	_, err := s.cron.AddFunc(s.CheckInterval, func() {
		s.runAnniversaryCheck(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to schedule founder anniversary check: %w", err)
	}

	s.cron.Start()
	s.logger.WithField("interval", s.CheckInterval).Info("Founder anniversary scheduler started")
	return nil
}

func (s *FounderAnniversaryScheduler) Stop() {
	s.cron.Stop()
	s.logger.Info("Founder anniversary scheduler stopped")
}

func (s *FounderAnniversaryScheduler) initAnniversaryTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS founder_anniversary_notifications (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			years_active INT NOT NULL,
			notification_type VARCHAR(50) NOT NULL,
			channel VARCHAR(50) NOT NULL DEFAULT 'email',
			sent_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			status VARCHAR(20) NOT NULL DEFAULT 'sent',
			error_message TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(user_id, years_active, notification_type)
		);
		CREATE INDEX IF NOT EXISTS idx_founder_anniversary_user_id ON founder_anniversary_notifications(user_id);
		CREATE INDEX IF NOT EXISTS idx_founder_anniversary_sent_at ON founder_anniversary_notifications(sent_at);
	`

	if _, err := s.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create anniversary table: %w", err)
	}
	return nil
}

func (s *FounderAnniversaryScheduler) runAnniversaryCheck(ctx context.Context) {
	s.logger.Info("Running founder anniversary check")

	founders, err := s.getFoundersWithAnniversaryToday(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get founders with anniversary")
		return
	}

	for _, founder := range founders {
		if founder.YearsActive < 1 {
			continue
		}

		alreadySent, err := s.hasAlreadySentNotification(ctx, founder.UserID, founder.YearsActive)
		if err != nil {
			s.logger.WithError(err).WithField("user_id", founder.UserID).Warn("Failed to check notification history")
			continue
		}
		if alreadySent {
			s.logger.WithFields(logrus.Fields{
				"user_id":      founder.UserID,
				"years_active": founder.YearsActive,
			}).Debug("Anniversary notification already sent")
			continue
		}

		if err := s.sendAnniversaryNotification(ctx, founder); err != nil {
			s.logger.WithError(err).WithField("user_id", founder.UserID).Error("Failed to send anniversary notification")
			continue
		}

		if err := s.recordNotificationSent(ctx, founder, "anniversary", "in_app"); err != nil {
			s.logger.WithError(err).Warn("Failed to record notification sent")
		}

		s.logger.WithFields(logrus.Fields{
			"user_id":      founder.UserID,
			"founder_num":  founder.FounderNumber,
			"years_active": founder.YearsActive,
		}).Info("Founder anniversary notification sent")
	}

	s.logger.WithField("founders_checked", len(founders)).Info("Founder anniversary check completed")
}

func (s *FounderAnniversaryScheduler) getFoundersWithAnniversaryToday(ctx context.Context) ([]*FounderAnniversary, error) {
	query := `
		SELECT
			u.id as user_id,
			COALESCE(u.founder_number, 0) as founder_number,
			u.email,
			COALESCE(u.name, '') as name,
			DATE_PART('year', CURRENT_DATE) - DATE_PART('year', u.created_at)::int as years_active,
			u.created_at as founder_since
		FROM users u
		WHERE u.is_founder = true
		AND u.deactivated_at IS NULL
		AND DATE_PART('month', u.created_at) = DATE_PART('month', CURRENT_DATE)
		AND DATE_PART('day', u.created_at) = DATE_PART('day', CURRENT_DATE)
		AND DATE_PART('year', CURRENT_DATE) > DATE_PART('year', u.created_at)
		ORDER BY u.founder_number ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var founders []*FounderAnniversary
	for rows.Next() {
		f := &FounderAnniversary{}
		err := rows.Scan(&f.UserID, &f.FounderNumber, &f.Email, &f.Name, &f.YearsActive, &f.FounderSince)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to scan founder row")
			continue
		}
		founders = append(founders, f)
	}

	return founders, rows.Err()
}

func (s *FounderAnniversaryScheduler) hasAlreadySentNotification(ctx context.Context, userID uuid.UUID, yearsActive int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM founder_anniversary_notifications
			WHERE user_id = $1 AND years_active = $2 AND notification_type = 'anniversary'
		)
	`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, userID, yearsActive).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("query failed: %w", err)
	}
	return exists, nil
}

func (s *FounderAnniversaryScheduler) sendAnniversaryNotification(ctx context.Context, founder *FounderAnniversary) error {
	if s.notifySvc == nil {
		return fmt.Errorf("notification service not configured")
	}

	data := map[string]interface{}{
		"founder_number": founder.FounderNumber,
		"years_active":  founder.YearsActive,
		"founder_since":  founder.FounderSince.Format("January 2, 2006"),
		"user_name":      founder.Name,
	}

	title := fmt.Sprintf("🎉 Happy %d-Year Founder Anniversary!", founder.YearsActive)
	body := fmt.Sprintf("Thank you for being a founder since %s. Your permanent status is valued and will never be revoked.",
		founder.FounderSince.Format("January 2, 2006"))

	if founder.YearsActive >= 5 {
		title = fmt.Sprintf("🎉 Happy %d-Year Founder Anniversary! 🎉", founder.YearsActive)
		body = fmt.Sprintf("Incredible! You've been a FunctionFly founder for %d years. Thank you for your incredible impact and dedication.", founder.YearsActive)
	}

	_, err := s.notifySvc.Send(ctx, notification.SendRequest{
		UserID:   founder.UserID,
		Type:     "founder_anniversary",
		Category: "founder",
		Title:    title,
		Body:     body,
		Channels: []string{notification.ChannelInApp, notification.ChannelEmail},
		Priority: notification.PriorityNormal,
		Data:     data,
	})

	return err
}

func (s *FounderAnniversaryScheduler) recordNotificationSent(ctx context.Context, founder *FounderAnniversary, notifType, channel string) error {
	query := `
		INSERT INTO founder_anniversary_notifications (user_id, years_active, notification_type, channel, status)
		VALUES ($1, $2, $3, $4, 'sent')
		ON CONFLICT (user_id, years_active, notification_type) DO NOTHING
	`

	_, err := s.db.ExecContext(ctx, query, founder.UserID, founder.YearsActive, notifType, channel)
	if err != nil {
		return fmt.Errorf("failed to record notification: %w", err)
	}
	return nil
}

type FounderTierUpdateScheduler struct {
	cron        *cron.Cron
	userRepo    storage.Repository
	db          *sql.DB
	logger      *logrus.Logger
	UpdateInterval string
}

func NewFounderTierUpdateScheduler(
	userRepo storage.Repository,
	db *sql.DB,
) *FounderTierUpdateScheduler {
	return &FounderTierUpdateScheduler{
		cron:           cron.New(cron.WithSeconds()),
		userRepo:       userRepo,
		db:             db,
		logger:         logrus.New(),
		UpdateInterval: "0 2 * * *",
	}
}

func (s *FounderTierUpdateScheduler) WithLogger(logger *logrus.Logger) *FounderTierUpdateScheduler {
	s.logger = logger
	return s
}

func (s *FounderTierUpdateScheduler) Start(ctx context.Context) error {
	_, err := s.cron.AddFunc(s.UpdateInterval, func() {
		s.runTierUpdate(ctx)
	})
	if err != nil {
		return fmt.Errorf("failed to schedule founder tier update: %w", err)
	}

	s.cron.Start()
	s.logger.WithField("interval", s.UpdateInterval).Info("Founder tier update scheduler started")
	return nil
}

func (s *FounderTierUpdateScheduler) Stop() {
	s.cron.Stop()
	s.logger.Info("Founder tier update scheduler stopped")
}

func (s *FounderTierUpdateScheduler) runTierUpdate(ctx context.Context) {
	s.logger.Info("Running founder tier update")

	if err := s.userRepo.CalculateAndUpdateFounderTiers(ctx); err != nil {
		s.logger.WithError(err).Error("Failed to calculate and update founder tiers")
		return
	}

	stats, err := s.getTierStats(ctx)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get tier stats")
	}

	s.logger.WithFields(logrus.Fields{
		"total":      stats.Total,
		"elite":      stats.Elite,
		"pro":        stats.Pro,
		"standard":   stats.Standard,
	}).Info("Founder tier update completed")
}

type tierStats struct {
	Elite    int64
	Pro      int64
	Standard int64
	Total    int64
}

func (s *FounderTierUpdateScheduler) getTierStats(ctx context.Context) (*tierStats, error) {
	query := `
		SELECT
			COUNT(CASE WHEN tier = 'founder_elite' THEN 1 END),
			COUNT(CASE WHEN tier = 'founder_pro' THEN 1 END),
			COUNT(CASE WHEN tier = 'founder' THEN 1 END),
			COUNT(*)
		FROM founder_tiers
	`

	var stats tierStats
	err := s.db.QueryRowContext(ctx, query).Scan(&stats.Elite, &stats.Pro, &stats.Standard, &stats.Total)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}
