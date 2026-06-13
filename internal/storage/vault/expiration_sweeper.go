package vault

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ExpirationSweeperConfig configures the background expiration worker.
type ExpirationSweeperConfig struct {
	// Interval between sweeps. Default 1h. Set to 0 to disable.
	Interval time.Duration
	// WarningWindow is how far ahead a secret is marked "expiring_soon".
	// Default 7d.
	WarningWindow time.Duration
	// BatchSize limits secrets handled per sweep. Default 200.
	BatchSize int
	// NotificationHook is called whenever a secret is expired or
	// transitioned to expiring_soon. The hook is fired from the sweeper
	// goroutine; it must not block for long. If nil, only DB state is
	// updated and audit rows are written.
	NotificationHook func(secret *Secret, event ExpirationEvent)
	// Logger is the structured logger used by the sweeper. If nil,
	// logrus.New() is used.
	Logger *logrus.Logger
}

// ExpirationEvent identifies why a notification hook was called.
type ExpirationEvent string

const (
	ExpirationEventExpiringSoon ExpirationEvent = "expiring_soon"
	ExpirationEventExpired      ExpirationEvent = "expired"
)

// ExpirationSweeper runs the background sweep loop for secret expiration.
// It is safe to construct once at process start and share across the
// application. A single instance must not be Start()'d more than once.
type ExpirationSweeper struct {
	repo   *Repository
	cfg    ExpirationSweeperConfig
	logger *logrus.Logger

	mu      sync.Mutex
	running bool
}

// NewExpirationSweeper constructs a sweeper with sane defaults.
func NewExpirationSweeper(repo *Repository, cfg ExpirationSweeperConfig) *ExpirationSweeper {
	if cfg.Interval == 0 {
		cfg.Interval = time.Hour
	}
	if cfg.WarningWindow == 0 {
		cfg.WarningWindow = 7 * 24 * time.Hour
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 200
	}
	logger := cfg.Logger
	if logger == nil {
		logger = logrus.New()
	}
	return &ExpirationSweeper{repo: repo, cfg: cfg, logger: logger}
}

// Start runs the sweeper until ctx is cancelled. It returns immediately
// when cfg.Interval == 0.
func (s *ExpirationSweeper) Start(ctx context.Context) {
	if s.cfg.Interval <= 0 {
		s.logger.Info("Vault expiration sweeper disabled (interval = 0)")
		return
	}
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	s.logger.WithFields(logrus.Fields{
		"interval":       s.cfg.Interval.String(),
		"warning_window": s.cfg.WarningWindow.String(),
		"batch_size":     s.cfg.BatchSize,
	}).Info("Starting vault expiration sweeper")

	// Run once immediately so freshly-restarted processes catch up.
	s.runOnce(ctx)

	ticker := time.NewTicker(s.cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping vault expiration sweeper")
			return
		case <-ticker.C:
			s.runOnce(ctx)
		}
	}
}

// RunOnce runs a single sweep cycle. Useful for tests and ad-hoc
// triggers.
func (s *ExpirationSweeper) RunOnce(ctx context.Context) (ExpiringCounts, error) {
	return s.sweep(ctx)
}

// ExpiringCounts reports the work done in a single sweep.
type ExpiringCounts struct {
	MarkedExpiringSoon int64
	MarkedExpired      int64
	TokensRevoked      int64
	NotificationsSent  int
	BreakGlassExpired  int64
}

func (s *ExpirationSweeper) runOnce(ctx context.Context) {
	counts, err := s.sweep(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Vault expiration sweep failed")
		return
	}
	s.logger.WithFields(logrus.Fields{
		"expiring_soon":  counts.MarkedExpiringSoon,
		"expired":        counts.MarkedExpired,
		"tokens_revoked": counts.TokensRevoked,
		"notifications":  counts.NotificationsSent,
		"break_glass":    counts.BreakGlassExpired,
	}).Info("Vault expiration sweep complete")
}

func (s *ExpirationSweeper) sweep(ctx context.Context) (ExpiringCounts, error) {
	var counts ExpiringCounts

	// 1) Secrets about to expire — flip to expiring_soon.
	expiring, err := s.repo.ListExpiringSoon(ctx, s.cfg.WarningWindow, s.cfg.BatchSize)
	if err != nil {
		return counts, err
	}
	for i := range expiring {
		secret := expiring[i]
		if err := s.repo.MarkSecretExpiringSoon(ctx, secret.ID, secret.TenantID); err != nil {
			s.logger.WithError(err).WithField("secret_id", secret.ID).Warn("Failed to mark expiring_soon")
			continue
		}
		counts.MarkedExpiringSoon++
		if s.cfg.NotificationHook != nil {
			s.safeNotify(&secret, ExpirationEventExpiringSoon)
			counts.NotificationsSent++
		}
		s.writeAuditLog(ctx, &secret, AuditActionUpdate, "expiring_soon", true)
	}

	// 2) Secrets past expires_at — flip to expired, revoke tokens.
	expired, err := s.repo.ListExpired(ctx, s.cfg.BatchSize)
	if err != nil {
		return counts, err
	}
	for i := range expired {
		secret := expired[i]
		changed, err := s.repo.MarkSecretExpired(ctx, secret.ID, secret.TenantID)
		if err != nil {
			s.logger.WithError(err).WithField("secret_id", secret.ID).Warn("Failed to mark expired")
			continue
		}
		if !changed {
			continue
		}
		counts.MarkedExpired++
		revoked, err := s.repo.RevokeAllTokensForSecret(ctx, secret.ID, secret.TenantID, "secret_expired")
		if err != nil {
			s.logger.WithError(err).WithField("secret_id", secret.ID).Warn("Failed to revoke tokens on expiry")
		}
		counts.TokensRevoked += revoked
		if s.cfg.NotificationHook != nil {
			s.safeNotify(&secret, ExpirationEventExpired)
			counts.NotificationsSent++
		}
		s.writeAuditLog(ctx, &secret, AuditActionExpire, "expired", revoked > 0)
	}

	// 3) Break-glass grants that have passed their window.
	if n, err := s.repo.ExpireBreakGlassRequests(ctx); err == nil {
		counts.BreakGlassExpired = n
	} else {
		s.logger.WithError(err).Warn("Failed to expire break-glass requests")
	}

	return counts, nil
}

func (s *ExpirationSweeper) safeNotify(secret *Secret, event ExpirationEvent) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.WithField("panic", r).Warn("Notification hook panicked")
		}
	}()
	s.cfg.NotificationHook(secret, event)
}

func (s *ExpirationSweeper) writeAuditLog(ctx context.Context, secret *Secret, action AuditAction, event string, success bool) {
	entry := &AuditLog{
		TenantID:  secret.TenantID,
		SecretID:  &secret.ID,
		ActorID:   "system:expiration-sweeper",
		ActorType: ActorTypeSystem,
		Action:    action,
		Success:   success,
		Metadata: JSONMap{
			"event":       event,
			"secret_name": secret.Name,
		},
	}
	if err := s.repo.CreateAuditLog(ctx, entry); err != nil {
		s.logger.WithError(err).Debug("Failed to write expiration audit log")
	}
}
