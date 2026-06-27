package scheduler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type StripeMeterUsageSchedulerConfig struct {
	Cron                    string
	Enabled                 bool
	DryRun                  bool
	LookbackHours           int
	MaxTenantsPerRun        int
	ReportRetryAttempts     int
	ReportRetryDelaySeconds int
}

func DefaultStripeMeterUsageSchedulerConfig() *StripeMeterUsageSchedulerConfig {
	return &StripeMeterUsageSchedulerConfig{
		Cron:                    "0 * * * *",
		Enabled:                 true,
		DryRun:                  false,
		LookbackHours:           1,
		MaxTenantsPerRun:        100,
		ReportRetryAttempts:     3,
		ReportRetryDelaySeconds: 5,
	}
}

func LoadStripeMeterUsageSchedulerConfig() *StripeMeterUsageSchedulerConfig {
	cfg := DefaultStripeMeterUsageSchedulerConfig()

	if v := os.Getenv("STRIPE_METER_SCHEDULER_ENABLED"); v != "" {
		if enabled, err := strconv.ParseBool(v); err == nil {
			cfg.Enabled = enabled
		}
	}

	if v := os.Getenv("STRIPE_METER_SCHEDULER_CRON"); v != "" {
		cfg.Cron = v
	}

	if v := os.Getenv("STRIPE_METER_SCHEDULER_DRY_RUN"); v != "" {
		if dryRun, err := strconv.ParseBool(v); err == nil {
			cfg.DryRun = dryRun
		}
	}

	if v := os.Getenv("STRIPE_METER_SCHEDULER_LOOKBACK_HOURS"); v != "" {
		if h, err := strconv.Atoi(v); err == nil && h > 0 {
			cfg.LookbackHours = h
		}
	}

	if v := os.Getenv("STRIPE_METER_SCHEDULER_MAX_TENANTS"); v != "" {
		if m, err := strconv.Atoi(v); err == nil && m > 0 {
			cfg.MaxTenantsPerRun = m
		}
	}

	return cfg
}

type StripeMeterUsageScheduler struct {
	cron       *cron.Cron
	repo       storage.Repository
	db         *storage.PostgresDB
	config     *StripeMeterUsageSchedulerConfig
	logger     *logrus.Logger
	stopOnce   sync.Once
	cancel     context.CancelFunc
	isRunning  bool
	mu         sync.RWMutex
}

func NewStripeMeterUsageScheduler(repo storage.Repository, db *storage.PostgresDB) *StripeMeterUsageScheduler {
	return &StripeMeterUsageScheduler{
		cron:   cron.New(),
		repo:   repo,
		db:     db,
		config: LoadStripeMeterUsageSchedulerConfig(),
		logger: logrus.New(),
	}
}

func (s *StripeMeterUsageScheduler) Start(ctx context.Context) error {
	if !s.config.Enabled {
		s.logger.Info("Stripe meter usage scheduler is disabled")
		return nil
	}

	if _, err := cron.ParseStandard(s.config.Cron); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	var ctxWithCancel context.Context
	ctxWithCancel, s.cancel = context.WithCancel(ctx)

	_, err := s.cron.AddFunc(s.config.Cron, func() {
		s.runMeteredBillingReporting(ctxWithCancel)
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	s.cron.Start()

	s.mu.Lock()
	s.isRunning = true
	s.mu.Unlock()

	s.logger.WithFields(logrus.Fields{
		"cron":        s.config.Cron,
		"lookback_hs": s.config.LookbackHours,
		"max_tenants": s.config.MaxTenantsPerRun,
		"dry_run":     s.config.DryRun,
	}).Info("Stripe meter usage scheduler started")

	return nil
}

func (s *StripeMeterUsageScheduler) Stop() error {
	s.stopOnce.Do(func() {
		s.mu.Lock()
		s.isRunning = false
		s.mu.Unlock()

		if s.cancel != nil {
			s.cancel()
		}
		<-s.cron.Stop().Done()
		s.logger.Info("Stripe meter usage scheduler stopped")
	})
	return nil
}

type tenantUsage struct {
	TenantID             uuid.UUID
	Plan                 string
	StripeCustomerID     string
	StripeSubscriptionID string
	TotalExecutions      int
	EventType            string
}

func (s *StripeMeterUsageScheduler) runMeteredBillingReporting(ctx context.Context) {
	start := time.Now()
	s.logger.Info("Starting Stripe metered billing usage reporting")

	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		s.logger.Warn("STRIPE_SECRET_KEY not set, skipping meter usage reporting")
		return
	}

	lookbackStart := time.Now().UTC().Add(-time.Duration(s.config.LookbackHours) * time.Hour)
	lookbackEnd := time.Now().UTC()

	tenants, err := s.getTenantsWithMeteredUsage(ctx, lookbackStart, lookbackEnd)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get tenants with metered usage")
		return
	}

	if len(tenants) == 0 {
		s.logger.Info("No tenants with metered usage found")
		return
	}

	processed := 0
	failed := 0

	for _, tu := range tenants {
		if processed >= s.config.MaxTenantsPerRun {
			s.logger.WithField("max_tenants", s.config.MaxTenantsPerRun).Info("Reached max tenants per run")
			break
		}

		if err := s.processTenantUsage(ctx, tu, lookbackStart, lookbackEnd, stripeKey); err != nil {
			s.logger.WithError(err).WithField("tenant_id", tu.TenantID).Error("Failed to process tenant usage")
			failed++
		} else {
			processed++
		}
	}

	duration := time.Since(start)
	s.logger.WithFields(logrus.Fields{
		"duration_ms":  duration.Milliseconds(),
		"processed":    processed,
		"failed":       failed,
		"total_found":  len(tenants),
	}).Info("Stripe metered billing reporting completed")
}

func (s *StripeMeterUsageScheduler) getTenantsWithMeteredUsage(ctx context.Context, start, end time.Time) ([]tenantUsage, error) {
	query := `
		SELECT DISTINCT
			cae.tenant_id,
			t.plan,
			t.stripe_customer_id,
			s.stripe_subscription_id,
			COUNT(*) as total_executions,
			'function_execution' as event_type
		FROM cost_allocation_entries cae
		JOIN tenants t ON t.id = cae.tenant_id
		LEFT JOIN subscriptions s ON s.tenant_id = cae.tenant_id
			AND s.status = 'active'
			AND s.stripe_subscription_id IS NOT NULL
			AND s.stripe_subscription_id != ''
		WHERE cae.timestamp >= $1
			AND cae.timestamp <= $2
			AND t.plan NOT IN ('free')
			AND t.stripe_customer_id IS NOT NULL
			AND s.stripe_subscription_id IS NOT NULL
		GROUP BY cae.tenant_id, t.plan, t.stripe_customer_id, s.stripe_subscription_id
		HAVING COUNT(*) > 0
		ORDER BY COUNT(*) DESC
	`

	rows, err := s.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenants with metered usage: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []tenantUsage
	for rows.Next() {
		var tu tenantUsage
		var customerID, subID sql.NullString

		err := rows.Scan(
			&tu.TenantID,
			&tu.Plan,
			&customerID,
			&subID,
			&tu.TotalExecutions,
			&tu.EventType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant usage: %w", err)
		}

		if customerID.Valid {
			tu.StripeCustomerID = customerID.String
		}
		if subID.Valid {
			tu.StripeSubscriptionID = subID.String
		}

		results = append(results, tu)
	}

	return results, rows.Err()
}

func (s *StripeMeterUsageScheduler) processTenantUsage(ctx context.Context, tu tenantUsage, periodStart, periodEnd time.Time, stripeKey string) error {
	tier := plans.GetUsagePricingTier(tu.Plan)

	if tier.IncludedRequestsMonthly == -1 {
		s.logger.WithFields(logrus.Fields{
			"tenant_id": tu.TenantID,
			"plan":      tu.Plan,
		}).Debug("Tenant has unlimited plan, no overage billing")
		return nil
	}

	if tier.OveragePricePer1000 == 0 {
		s.logger.WithFields(logrus.Fields{
			"tenant_id": tu.TenantID,
			"plan":      tu.Plan,
		}).Debug("Plan does not have overage billing")
		return nil
	}

	if s.config.DryRun {
		s.logger.WithFields(logrus.Fields{
			"tenant_id":         tu.TenantID,
			"plan":              tu.Plan,
			"total_executions":  tu.TotalExecutions,
			"included_requests": tier.IncludedRequestsMonthly,
			"action":            "dry_run",
		}).Info("Dry run: would report overage to Stripe")
		return nil
	}

	overageQuantity := tu.TotalExecutions

	idempotencyKey := fmt.Sprintf("%s_%s_%d_%d",
		tu.TenantID.String(),
		tu.EventType,
		periodStart.Unix(),
		periodEnd.Unix(),
	)

	meterEventName := os.Getenv("STRIPE_OVERAGE_METER_NAME")
	if meterEventName == "" {
		meterEventName = "functionfly_overage"
	}

	meterPayload := map[string]interface{}{
		"event_name": meterEventName,
		"timestamp":  time.Now().UTC().Unix(),
		"identifier": idempotencyKey,
		"payload": map[string]string{
			"value":              strconv.Itoa(overageQuantity),
			"stripe_customer_id": tu.StripeCustomerID,
			"event_type":         tu.EventType,
			"tenant_id":          tu.TenantID.String(),
			"plan":               tu.Plan,
		},
	}

	eventID, err := s.createMeterEvent(ctx, meterPayload, stripeKey)
	if err != nil {
		return fmt.Errorf("failed to create meter event: %w", err)
	}

	if err := s.recordUsageReport(ctx, tu, eventID, overageQuantity, periodStart, periodEnd, idempotencyKey, meterEventName); err != nil {
		s.logger.WithError(err).WithField("tenant_id", tu.TenantID).Warn("Failed to record usage report")
	}

	s.logger.WithFields(logrus.Fields{
		"tenant_id":          tu.TenantID,
		"event_id":           eventID,
		"overage_quantity":   overageQuantity,
		"meter_event_name":   meterEventName,
		"stripe_customer_id": tu.StripeCustomerID,
	}).Info("Successfully reported overage usage to Stripe")

	return nil
}

func (s *StripeMeterUsageScheduler) createMeterEvent(ctx context.Context, payload map[string]interface{}, stripeKey string) (string, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://meter-events.stripe.com/v1/billing/meter_events", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+stripeKey)
	req.Header.Set("Stripe-Version", "2024-04-15")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to Stripe: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		ID    string `json:"id"`
		Error *struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("stripe API error: %s (code: %s)", result.Error.Message, result.Error.Code)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return result.ID, nil
}

func (s *StripeMeterUsageScheduler) recordUsageReport(ctx context.Context, tu tenantUsage, eventID string, quantity int, periodStart, periodEnd time.Time, idempotencyKey, meterEventName string) error {
	report := &storage.StripeUsageReport{
		ID:                  uuid.New(),
		TenantID:            tu.TenantID,
		SubscriptionID:      tu.StripeSubscriptionID,
		SubscriptionItemID:  "",
		UsageQuantity:       quantity,
		UsagePeriodStart:    periodStart,
		UsagePeriodEnd:      periodEnd,
		StripeTimestamp:     time.Now().UTC().Unix(),
		StripeUsageRecordID: eventID,
		Status:              "reported",
		IdempotencyKey:      idempotencyKey,
		MeterEventName:      meterEventName,
		Metadata:            json.RawMessage(fmt.Sprintf(`{"plan":"%s","event_type":"%s"}`, tu.Plan, tu.EventType)),
	}

	usageRepo := storage.NewUsageReportingRepository(s.db)
	return usageRepo.CreateUsageReport(ctx, report)
}

func (s *StripeMeterUsageScheduler) GetSchedule() map[string]interface{} {
	nextRun := "unknown"
	if s.cron != nil {
		entries := s.cron.Entries()
		if len(entries) > 0 {
			nextRun = entries[0].Next.Format(time.RFC3339)
		}
	}

	return map[string]interface{}{
		"enabled":        s.config.Enabled,
		"cron":           s.config.Cron,
		"dry_run":        s.config.DryRun,
		"lookback_hours": s.config.LookbackHours,
		"max_tenants":    s.config.MaxTenantsPerRun,
		"next_run":       nextRun,
	}
}

func (s *StripeMeterUsageScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

func (s *StripeMeterUsageScheduler) TriggerNow(ctx context.Context) error {
	s.logger.Info("Manually triggering Stripe metered billing reporting")
	s.runMeteredBillingReporting(ctx)
	return nil
}