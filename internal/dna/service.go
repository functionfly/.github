package dna

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/storage/dna"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// SourceCodeFetcher retrieves function source code by ID.
type SourceCodeFetcher interface {
	GetFunctionSourceCode(functionID string) (sourceCode string, runtime string, err error)
}

// WalletDebiter debits credits from a user's wallet.
type WalletDebiter interface {
	DebitForDNAMutation(ctx context.Context, userID string, amountUSD float64, functionID string) error
}

// MutationNotifier sends notifications when mutations are proposed.
type MutationNotifier interface {
	NotifyMutationProposed(ctx context.Context, tenantID, functionID, mutationType, triggerReason string) error
}

// Service orchestrates DNA analysis, AI calls, and mutation management.
type Service struct {
	repo       *dna.Repository
	logger     *logrus.Logger
	aiBaseURL  string
	aiAPIKey   string
	httpClient *http.Client
	// sourceCodeFetcher is optional; when set, the service includes source code
	// in AI requests so the LLM can generate real code variants.
	sourceCodeFetcher SourceCodeFetcher
	// walletDebiter is optional; when set, accepting a mutation debits credits.
	walletDebiter WalletDebiter
	// mutationNotifier is optional; when set, sends real-time notifications on new mutations.
	mutationNotifier MutationNotifier
	// canaryTriggerer is optional; when set, triggers canary deployment on mutation acceptance.
	canaryTriggerer CanaryTriggerer
	// mutationValidator validates mutations before acceptance via sandbox testing.
	mutationValidator *MutationValidator
	// platformSettingsProvider retrieves user platform settings (auto-evolve, canary, etc.).
	// If not set, defaults are used for all settings.
	platformSettingsProvider PlatformSettingsProvider
	// serverCtx is cancelled on server shutdown. Use for fire-and-forget goroutines
	// instead of context.Background() to prevent goroutine leaks.
	serverCtx context.Context
	// aiCircuitBreaker prevents cascading failures when the AI service is down.
	aiCircuitBreaker *circuitBreaker
}

// NewService creates a new DNA service.
func NewService(repo *dna.Repository, logger *logrus.Logger) *Service {
	cbThreshold := getEnvInt("AI_CIRCUIT_BREAKER_THRESHOLD", 5)
	cbCooldownMinutes := getEnvInt("AI_CIRCUIT_BREAKER_COOLDOWN_MINUTES", 2)
	return &Service{
		repo:      repo,
		logger:    logger,
		aiBaseURL: getEnvOrDefault("AI_SERVICE_URL", ""), // Must be set explicitly
		aiAPIKey:  os.Getenv("AI_SERVICE_API_KEY"),
		httpClient: &http.Client{Timeout: 2 * time.Minute},
		aiCircuitBreaker: newCircuitBreaker(cbThreshold, time.Duration(cbCooldownMinutes)*time.Minute),
	}
}

// SetSourceCodeFetcher sets the source code fetcher for real AI code generation.
func (s *Service) SetSourceCodeFetcher(fetcher SourceCodeFetcher) {
	s.sourceCodeFetcher = fetcher
}

// SetWalletDebiter sets the wallet debiter for credit deduction on mutation acceptance.
func (s *Service) SetWalletDebiter(debiter WalletDebiter) {
	s.walletDebiter = debiter
}

// SetMutationNotifier sets the notifier for real-time mutation notifications.
func (s *Service) SetMutationNotifier(notifier MutationNotifier) {
	s.mutationNotifier = notifier
}

// SetCanaryTriggerer sets the canary triggerer for deployment on mutation acceptance.
func (s *Service) SetCanaryTriggerer(triggerer CanaryTriggerer) {
	s.canaryTriggerer = triggerer
}

// SetMutationValidator sets the validator for pre-acceptance mutation testing.
func (s *Service) SetMutationValidator(validator *MutationValidator) {
	s.mutationValidator = validator
}

// SetPlatformSettingsProvider sets the provider for user platform settings.
func (s *Service) SetPlatformSettingsProvider(provider PlatformSettingsProvider) {
	s.platformSettingsProvider = provider
}

// SetServerContext sets the server-lifetime context for fire-and-forget goroutines.
// This context should be cancelled on server shutdown.
func (s *Service) SetServerContext(ctx context.Context) {
	s.serverCtx = ctx
}

// serverContext returns the server context, falling back to Background if not set.
func (s *Service) serverContext() context.Context {
	if s.serverCtx != nil {
		return s.serverCtx
	}
	return context.Background()
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// ──────────────────────────────────────────────────────────────────────────────
// Public API
// ──────────────────────────────────────────────────────────────────────────────

// GetProfile returns the DNA profile for a function, creating one if needed.
func (s *Service) GetProfile(ctx context.Context, functionID, functionType, tenantID string) (*dna.DNAProfile, error) {
	return s.repo.GetOrCreateProfile(ctx, functionID, functionType, tenantID)
}

// ListMutations returns mutations with filtering.
func (s *Service) ListMutations(ctx context.Context, functionID, status string, limit, offset int) ([]*dna.Mutation, int, error) {
	return s.repo.ListMutations(ctx, functionID, status, limit, offset)
}

// GetMutation returns a single mutation with full details.
func (s *Service) GetMutation(ctx context.Context, mutationID string) (*dna.Mutation, error) {
	return s.repo.GetMutation(ctx, mutationID)
}

// AcceptMutation accepts a proposed variant and triggers deployment.
func (s *Service) AcceptMutation(ctx context.Context, mutationID, userID, tenantID string, canaryPct int) error {
	m, err := s.repo.GetMutation(ctx, mutationID)
	if err != nil {
		return fmt.Errorf("get mutation: %w", err)
	}
	if m == nil {
		return fmt.Errorf("mutation not found")
	}
	if m.TenantID != tenantID {
		return fmt.Errorf("access denied")
	}
	if m.Status != "proposed" {
		return fmt.Errorf("mutation is not in proposed status: %s", m.Status)
	}

	// Validate mutation in sandbox before acceptance
	if s.mutationValidator != nil && m.OriginalCode != nil && m.MutatedCode != nil {
		validationCtx, validationCancel := context.WithTimeout(ctx, 60*time.Second)
		defer validationCancel()

		runtime := "python"
		if m.FunctionType != "" {
			runtime = m.FunctionType
		}

		report, err := s.mutationValidator.ValidateMutation(
			validationCtx,
			*m.OriginalCode,
			*m.MutatedCode,
			runtime,
			nil,
		)
		if err != nil {
			s.logger.WithError(err).WithField("mutation_id", mutationID).Warn("dna: mutation validation failed")
			return fmt.Errorf("mutation validation failed: %w", err)
		}

		if !report.Passed {
			s.logger.WithFields(logrus.Fields{
				"mutation_id":   mutationID,
				"failed_tests":  report.FailedTests,
				"security_fail": len(report.Errors) > 0,
			}).Warn("dna: mutation rejected by validator")

			if updateErr := s.repo.UpdateMutationStatus(ctx, mutationID, "rejected", map[string]interface{}{
				"reason": fmt.Sprintf("Validation failed: %d/%d tests passed, security checks: %v",
					report.PassedTests, report.TotalTests, report.Errors),
			}); updateErr != nil {
				s.logger.WithError(updateErr).Warn("dna: failed to update mutation status after validation failure")
			}
			return fmt.Errorf("mutation failed validation: %d tests failed", report.FailedTests)
		}

		s.logger.WithFields(logrus.Fields{
			"mutation_id":      mutationID,
			"passed_tests":     report.PassedTests,
			"total_tests":      report.TotalTests,
			"performance_diff": report.PerformanceDiff.ImprovementPct,
		}).Info("dna: mutation passed validation")
	}

	// Update status first (cheap, atomic). If debit fails after this,
	// the mutation is accepted but unpaid — safer than the reverse.
	if err := s.repo.UpdateMutationStatus(ctx, mutationID, "accepted", map[string]interface{}{
		"accepted_by": userID,
	}); err != nil {
		return fmt.Errorf("update mutation status: %w", err)
	}

	// Debit credits from wallet (50 credits per architecture doc)
	if s.walletDebiter != nil {
		if err := s.walletDebiter.DebitForDNAMutation(ctx, userID, 50.0, m.FunctionID); err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"mutation_id":  mutationID,
				"user_id":      userID,
				"function_id":  m.FunctionID,
			}).Error("dna: mutation accepted but wallet debit failed — manual review required")
			// Don't revert the acceptance. Log for manual reconciliation.
		}
	}

	// Get platform settings to determine canary percentage and approval mode
	canaryPctToUse := canaryPct
	requireApproval := true
	if s.platformSettingsProvider != nil {
		settings, err := s.platformSettingsProvider.GetPlatformSettings(ctx, userID)
		if err == nil && settings != nil {
			if canaryPctToUse <= 0 {
				canaryPctToUse = settings.DefaultCanaryPct
			}
			requireApproval = settings.RequireApproval
		}
	}

	// If platform auto-approval is disabled, just accept and don't auto-deploy
	if !requireApproval {
		s.logger.WithFields(logrus.Fields{
			"mutation_id": mutationID,
			"function_id": m.FunctionID,
		}).Info("dna: auto-approving mutation (require_approval=false)")
		// Mark as auto-approved
		if err := s.repo.UpdateMutationStatus(ctx, mutationID, "accepted", map[string]interface{}{
			"accepted_by": "system-auto",
		}); err != nil {
			s.logger.WithError(err).Warn("dna: failed to mark auto-approval")
		}
		// Don't trigger canary when auto-approved — user will manually deploy
		return nil
	}

	// Trigger canary deployment if triggerer is configured
	if s.canaryTriggerer != nil {
		srvCtx := s.serverContext()
		go func() {
			version := safePrefix(mutationID, 8)
			if m.MutatedHash != nil && *m.MutatedHash != "" {
				version = safePrefix(*m.MutatedHash, 8)
			}
			version = "dna-" + version
			if err := s.canaryTriggerer.TriggerCanary(srvCtx, m.FunctionID, version, canaryPctToUse); err != nil {
				s.logger.WithError(err).WithFields(logrus.Fields{
					"mutation_id":  mutationID,
					"function_id":  m.FunctionID,
					"canary_pct":   canaryPctToUse,
				}).Warn("dna: failed to trigger canary deployment")
			}
		}()
	}

	return nil
}

// RejectMutation rejects a proposed variant.
func (s *Service) RejectMutation(ctx context.Context, mutationID, tenantID, reason string) error {
	m, err := s.repo.GetMutation(ctx, mutationID)
	if err != nil {
		return fmt.Errorf("get mutation: %w", err)
	}
	if m == nil {
		return fmt.Errorf("mutation not found")
	}
	if m.TenantID != tenantID {
		return fmt.Errorf("access denied")
	}
	if m.Status != "proposed" {
		return fmt.Errorf("mutation is not in proposed status: %s", m.Status)
	}

	return s.repo.UpdateMutationStatus(ctx, mutationID, "rejected", map[string]interface{}{
		"reason": reason,
	})
}

// SetEvolutionEnabled toggles evolution for a function.
func (s *Service) SetEvolutionEnabled(ctx context.Context, functionID, functionType string, enabled bool) error {
	return s.repo.SetEvolutionEnabled(ctx, functionID, functionType, enabled)
}

// CheckFunctionOwnership verifies that a function's DNA profile belongs to the given tenant.
// Returns nil if the profile doesn't exist (function has no DNA yet) or belongs to the tenant.
// Returns an error if the profile exists but belongs to a different tenant.
func (s *Service) CheckFunctionOwnership(ctx context.Context, functionID, functionType, tenantID string) error {
	profile, err := s.repo.GetProfileReadOnly(ctx, functionID, functionType)
	if err != nil {
		return fmt.Errorf("check ownership: %w", err)
	}
	if profile != nil && profile.TenantID != tenantID {
		return fmt.Errorf("access denied")
	}
	return nil
}

// VerifyDNAHash recomputes the DNA hash for a function and compares it to the stored hash.
// Returns (matches bool, storedHash, computedHash, error).
// The DNA hash is now computed from the function's source code (if available) combined with
// metrics, providing true code integrity verification rather than just metric drift detection.
func (s *Service) VerifyDNAHash(ctx context.Context, functionID, functionType string) (bool, string, string, error) {
	profile, err := s.repo.GetOrCreateProfile(ctx, functionID, functionType, "")
	if err != nil {
		return false, "", "", fmt.Errorf("get profile: %w", err)
	}
	if profile.DNAHash == nil || *profile.DNAHash == "" {
		return true, "", "", nil // no hash to verify
	}

	// Try to get source code for code-based hash
	var codeHash string
	if s.sourceCodeFetcher != nil {
		sourceCode, _, err := s.sourceCodeFetcher.GetFunctionSourceCode(functionID)
		if err == nil && sourceCode != "" {
			h := sha256.Sum256([]byte(sourceCode))
			codeHash = fmt.Sprintf("code:%x", h[:8])
		}
	}

	// Fall back to metric-based hash if no source code available
	if codeHash == "" {
		metrics, err := s.repo.AggregateMetrics(ctx, functionID, 48*time.Hour)
		if err != nil {
			return false, "", "", fmt.Errorf("aggregate metrics: %w", err)
		}
		avgLatency := metrics.AvgLatencyMs
		hashData := fmt.Sprintf("%s:%d:%.2f:%.4f", functionID, metrics.TotalExecutions, avgLatency, metrics.SuccessRate)
		hash := sha256.Sum256([]byte(hashData))
		codeHash = fmt.Sprintf("sha256:%x", hash[:16])
	}

	return *profile.DNAHash == codeHash, *profile.DNAHash, codeHash, nil
}

// RecordExecutionFromPipeline is the simplified entry point called from the execution handler.
// It maps minimal execution data into a full ExecutionMetric and records it.
func (s *Service) RecordExecutionFromPipeline(ctx context.Context, functionID, functionType string, durationMs int, statusCode int, coldStart bool, region string) {
	execID := uuid.New().String()
	m := &dna.ExecutionMetric{
		FunctionID:   functionID,
		FunctionType: functionType,
		ExecutionID:  &execID,
		DurationMs:   durationMs,
		StatusCode:   &statusCode,
		ColdStart:    coldStart,
		Region:       &region,
	}
	if statusCode >= 500 {
		m.ErrorCategory = "runtime"
	} else if statusCode == 408 {
		m.ErrorCategory = "timeout"
	} else {
		m.ErrorCategory = "none"
	}
	if err := s.RecordExecution(ctx, m); err != nil {
		s.logger.WithError(err).WithField("function_id", functionID).Warn("dna: failed to record execution from pipeline")
	}
}

// RecordExecution records micro-data from a function execution (called from execution pipeline).
func (s *Service) RecordExecution(ctx context.Context, m *dna.ExecutionMetric) error {
	if err := s.repo.InsertExecutionMetric(ctx, m); err != nil {
		return err
	}

	// Check if we should trigger analysis (every 10,000 executions)
	profile, err := s.repo.GetOrCreateProfile(ctx, m.FunctionID, m.FunctionType, "")
	if err != nil {
		s.logger.WithError(err).Warn("dna: failed to get profile for analysis check")
		return nil
	}
	if profile.EvolutionEnabled && profile.TotalExecutions > 0 && profile.TotalExecutions%10000 == 0 {
		if err := s.repo.EnqueueAnalysis(ctx, m.FunctionID, m.FunctionType, profile.TenantID, 5); err != nil {
			s.logger.WithError(err).Warn("dna: failed to enqueue analysis")
		}
	}
	return nil
}

// TriggerAnalysis manually queues a DNA analysis for a function.
func (s *Service) TriggerAnalysis(ctx context.Context, functionID, functionType, tenantID string) error {
	return s.repo.EnqueueAnalysis(ctx, functionID, functionType, tenantID, 1)
}

// GetInsights returns time-series DNA insights for a function.
func (s *Service) GetInsights(ctx context.Context, functionID, period string) (map[string]interface{}, error) {
	var since time.Duration
	switch period {
	case "7d":
		since = 7 * 24 * time.Hour
	case "30d":
		since = 30 * 24 * time.Hour
	case "90d":
		since = 90 * 24 * time.Hour
	default:
		since = 30 * 24 * time.Hour
	}

	metrics, err := s.repo.AggregateMetrics(ctx, functionID, since)
	if err != nil {
		return nil, err
	}

	mutations, total, err := s.repo.ListMutations(ctx, functionID, "", 100, 0)
	if err != nil {
		return nil, err
	}

	outcomes := map[string]int{
		"accepted": 0, "rejected": 0, "proposed": 0, "deployed": 0, "rolled_back": 0,
	}
	for _, m := range mutations {
		outcomes[m.Status]++
	}

	return map[string]interface{}{
		"function_id": functionID,
		"period":      period,
		"metrics":     metrics,
		"mutation_outcomes": map[string]interface{}{
			"total":      total,
			"outcomes":   outcomes,
		},
	}, nil
}

// GetEnterpriseInsights returns tenant-wide DNA analytics.
func (s *Service) GetEnterpriseInsights(ctx context.Context, tenantID, period string) (*dna.TenantInsights, error) {
	var since time.Duration
	switch period {
	case "7d":
		since = 7 * 24 * time.Hour
	case "30d":
		since = 30 * 24 * time.Hour
	case "90d":
		since = 90 * 24 * time.Hour
	default:
		since = 30 * 24 * time.Hour
	}
	return s.repo.GetTenantInsights(ctx, tenantID, since)
}

// ──────────────────────────────────────────────────────────────────────────────
// Analysis Worker
// ──────────────────────────────────────────────────────────────────────────────

// RunAnalysisWorker starts the background analysis worker loop.
func (s *Service) RunAnalysisWorker(ctx context.Context) {
	s.logger.Info("dna: analysis worker started")
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("dna: analysis worker stopped")
			return
		case <-ticker.C:
			s.processNextAnalysis(ctx)
		}
	}
}

func (s *Service) processNextAnalysis(ctx context.Context) {
	queueID, functionID, functionType, err := s.repo.DequeueAnalysis(ctx)
	if err != nil {
		s.logger.WithError(err).Error("dna: dequeue analysis failed")
		return
	}
	if queueID == "" {
		return // nothing to process
	}

	s.logger.WithFields(logrus.Fields{
		"queue_id":     queueID,
		"function_id":  functionID,
		"function_type": functionType,
	}).Info("dna: processing analysis")

	if err := s.runAnalysis(ctx, functionID, functionType); err != nil {
		s.logger.WithError(err).Error("dna: analysis failed")
		if failErr := s.repo.FailAnalysis(ctx, queueID, err.Error()); failErr != nil {
			s.logger.WithError(failErr).Error("dna: failed to mark analysis as failed")
		}
		return
	}

	if err := s.repo.CompleteAnalysis(ctx, queueID); err != nil {
		s.logger.WithError(err).Error("dna: failed to mark analysis as completed")
	}
}

func (s *Service) runAnalysis(ctx context.Context, functionID, functionType string) error {
	// 1. Aggregate metrics over 48h window
	metrics, err := s.repo.AggregateMetrics(ctx, functionID, 48*time.Hour)
	if err != nil {
		return fmt.Errorf("aggregate metrics: %w", err)
	}

	if metrics.TotalExecutions < 100 {
		s.logger.WithField("function_id", functionID).Info("dna: not enough executions for analysis")
		return nil
	}

	// 2. Get current profile
	profile, err := s.repo.GetOrCreateProfile(ctx, functionID, functionType, "")
	if err != nil {
		return fmt.Errorf("get profile: %w", err)
	}

	// 3. Compute fitness score
	fitness := s.computeFitness(metrics)

	// 4. Update profile with latest metrics
	avgLatency := metrics.AvgLatencyMs
	p99Latency := metrics.P99LatencyMs
	errorDist, _ := json.Marshal(metrics.ErrorDistribution)
	inputPatterns, _ := json.Marshal(metrics.InputPatterns)

	profile.FitnessScore = fitness
	profile.AvgLatencyMs = &avgLatency
	profile.P99LatencyMs = &p99Latency
	profile.SuccessRate = metrics.SuccessRate
	profile.ErrorDistribution = errorDist
	profile.InputPatterns = inputPatterns
	profile.TotalExecutions = metrics.TotalExecutions
	now := time.Now()
	profile.LastAnalyzedAt = &now

	// Compute DNA hash — prefer code-based hash for integrity verification
	var newHash string
	if s.sourceCodeFetcher != nil {
		sourceCode, _, err := s.sourceCodeFetcher.GetFunctionSourceCode(functionID)
		if err == nil && sourceCode != "" {
			h := sha256.Sum256([]byte(sourceCode))
			newHash = fmt.Sprintf("code:%x", h[:8])
		}
	}
	if newHash == "" {
		// Fall back to metric-based hash
		hashData := fmt.Sprintf("%s:%d:%.2f:%.4f", functionID, metrics.TotalExecutions, avgLatency, metrics.SuccessRate)
		hash := sha256.Sum256([]byte(hashData))
		newHash = fmt.Sprintf("sha256:%x", hash[:16])
	}

	// Verify hash: compare old vs new to detect drift
	if profile.DNAHash != nil && *profile.DNAHash != "" && *profile.DNAHash != newHash {
		s.logger.WithFields(logrus.Fields{
			"function_id": functionID,
			"old_hash":    *profile.DNAHash,
			"new_hash":    newHash,
		}).Info("dna: hash drift detected — function DNA has changed")
	}
	profile.DNAHash = &newHash

	if err := s.repo.UpdateProfile(ctx, profile); err != nil {
		return fmt.Errorf("update profile: %w", err)
	}

	// 5. Check if auto-evolution is enabled (skip if disabled globally)
	if !s.isAutoEvolveEnabled(ctx, profile.TenantID) {
		s.logger.WithField("function_id", functionID).Debug("dna: auto-evolve disabled, skipping mutation proposal")
		return nil
	}

	// Check if we should propose a mutation
	shouldMutate, mutationType, reason := s.shouldMutate(metrics, profile)
	if !shouldMutate {
		return nil
	}

	// 6. Call AI service to generate variant
	proposal, err := s.callAIForVariant(ctx, functionID, functionType, mutationType, reason, metrics)
	if err != nil {
		s.logger.WithError(err).Warn("dna: AI variant generation failed")
		return nil // non-fatal
	}

	// 7. Store mutation
	mutation := &dna.Mutation{
		FunctionID:         functionID,
		FunctionType:       functionType,
		TenantID:           profile.TenantID,
		Generation:         profile.Generation,
		MutationType:       mutationType,
		Status:             "proposed",
		TriggerReason:      &reason,
		OriginalCode:       proposal.OriginalCode,
		MutatedCode:        proposal.MutatedCode,
		OriginalHash:       proposal.OriginalHash,
		MutatedHash:        proposal.MutatedHash,
		Diff:               proposal.Diff,
		EstimatedImpact:    proposal.EstimatedImpact,
		Confidence:         &proposal.Confidence,
		ModelUsed:          &proposal.ModelUsed,
		AnalysisWindowHours: intPtr(48),
		ExecutionsAnalyzed: intPtr(int(metrics.TotalExecutions)),
	}
	if err := s.repo.CreateMutation(ctx, mutation); err != nil {
		return fmt.Errorf("create mutation: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"function_id":  functionID,
		"mutation_id":  mutation.ID,
		"mutation_type": mutationType,
		"confidence":   proposal.Confidence,
	}).Info("dna: mutation proposed")

	// Send notification if configured
	if s.mutationNotifier != nil && s.shouldNotifyProposal(ctx, profile.TenantID) {
		srvCtx := s.serverContext()
		go func() {
			if err := s.mutationNotifier.NotifyMutationProposed(
				srvCtx, profile.TenantID, functionID, mutationType, reason,
			); err != nil {
				s.logger.WithError(err).Warn("dna: failed to send mutation notification")
			}
		}()
	}

	// Send real-time notification to the developer
	if s.mutationNotifier != nil {
		srvCtx := s.serverContext()
		go func() {
			if err := s.mutationNotifier.NotifyMutationProposed(
				srvCtx, profile.TenantID, functionID, mutationType, reason,
			); err != nil {
				s.logger.WithError(err).Warn("dna: failed to send mutation notification")
			}
		}()
	}

	return nil
}

// computeFitness calculates a 0-100 fitness score from aggregated metrics.
func (s *Service) computeFitness(m *dna.AggregatedMetrics) float64 {
	score := 50.0

	// Success rate contribution (0-30 points)
	score += m.SuccessRate * 30.0

	// Latency contribution (0-20 points)
	if m.P99LatencyMs < 100 {
		score += 20.0
	} else if m.P99LatencyMs < 500 {
		score += 20.0 * (1.0 - (m.P99LatencyMs-100)/400.0)
	}

	// Cold start penalty (-5 points if >20%)
	if m.ColdStartRate > 0.2 {
		score -= 5.0
	}

	// Error diversity penalty (-5 per error category)
	if len(m.ErrorDistribution) > 0 {
		score -= float64(len(m.ErrorDistribution)) * 2.5
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

// shouldMutate decides if a mutation should be proposed based on metrics.
// isAutoEvolveEnabled checks if auto-evolution is enabled for the tenant.
func (s *Service) isAutoEvolveEnabled(ctx context.Context, tenantID string) bool {
	if s.platformSettingsProvider == nil {
		return true // default: enabled
	}
	// We use userID = tenantID for platform settings (system-wide)
	settings, err := s.platformSettingsProvider.GetPlatformSettings(ctx, tenantID)
	if err != nil || settings == nil {
		return true
	}
	return settings.AutoEvolve
}

// shouldNotifyProposal checks if notifications are enabled for mutation proposals.
func (s *Service) shouldNotifyProposal(ctx context.Context, userID string) bool {
	if s.platformSettingsProvider == nil {
		return true // default: enabled
	}
	settings, err := s.platformSettingsProvider.GetPlatformSettings(ctx, userID)
	if err != nil || settings == nil {
		return true
	}
	return settings.NotifyOnProposal
}

func (s *Service) shouldMutate(m *dna.AggregatedMetrics, p *dna.DNAProfile) (bool, string, string) {
	// Latency regression
	if m.P99LatencyMs > 500 {
		return true, "optimize_latency",
			fmt.Sprintf("P99 latency is %.0fms (threshold: 500ms). Detected %d executions with high tail latency.", m.P99LatencyMs, m.TotalExecutions)
	}

	// Error rate
	if m.SuccessRate < 0.95 {
		return true, "fix_error_pattern",
			fmt.Sprintf("Success rate is %.1f%% (threshold: 95%%). Error distribution: %v", m.SuccessRate*100, m.ErrorDistribution)
	}

	// High cold start rate
	if m.ColdStartRate > 0.3 {
		return true, "reduce_memory",
			fmt.Sprintf("Cold start rate is %.1f%% (threshold: 30%%). Consider memory optimization.", m.ColdStartRate*100)
	}

	// Memory pressure
	if m.AvgMemoryPeakMb > 256 {
		return true, "reduce_memory",
			fmt.Sprintf("Average peak memory is %.0fMB (threshold: 256MB).", m.AvgMemoryPeakMb)
	}

	// Fitness-based refactoring
	if p.FitnessScore < 60 && p.TotalMutations == 0 {
		return true, "refactor_hotpath",
			fmt.Sprintf("Function fitness score is %.0f/100 with no prior evolutions. Opportunities for improvement detected.", p.FitnessScore)
	}

	return false, "", ""
}

// AI Service integration

type variantRequest struct {
	FunctionID      string                 `json:"function_id"`
	MutationType    string                 `json:"mutation_type"`
	TriggerReason   string                 `json:"trigger_reason"`
	Metrics         *dna.AggregatedMetrics `json:"metrics"`
	CurrentCode     string                 `json:"current_code,omitempty"`
	Runtime         string                 `json:"runtime,omitempty"`
}

type variantResponse struct {
	OriginalCode    *string         `json:"original_code"`
	MutatedCode     *string         `json:"mutated_code"`
	OriginalHash    *string         `json:"original_hash"`
	MutatedHash     *string         `json:"mutated_hash"`
	Diff            *string         `json:"diff"`
	EstimatedImpact json.RawMessage `json:"estimated_impact"`
	Confidence      float64         `json:"confidence"`
	ModelUsed       string          `json:"model_used"`
}

func (s *Service) callAIForVariant(ctx context.Context, functionID, functionType, mutationType, reason string, metrics *dna.AggregatedMetrics) (*variantResponse, error) {
	if !s.aiCircuitBreaker.allow() {
		return nil, fmt.Errorf("ai service circuit breaker open — service temporarily unavailable")
	}

	reqBody := variantRequest{
		FunctionID:    functionID,
		MutationType:  mutationType,
		TriggerReason: reason,
		Metrics:       metrics,
	}

	// Fetch source code for real LLM code generation
	if s.sourceCodeFetcher != nil {
		sourceCode, runtime, err := s.sourceCodeFetcher.GetFunctionSourceCode(functionID)
		if err != nil {
			s.logger.WithError(err).WithField("function_id", functionID).Warn("dna: failed to fetch source code for AI generation")
		} else if sourceCode != "" {
			reqBody.CurrentCode = sourceCode
			reqBody.Runtime = runtime
		}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.aiBaseURL+"/api/dna/generate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.aiAPIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.aiAPIKey)
	}
	req.Header.Set("X-Function-ID", functionID)
	req.Header.Set("X-Mutation-Type", mutationType)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.aiCircuitBreaker.recordFailure()
		return nil, fmt.Errorf("ai service call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		s.aiCircuitBreaker.recordFailure()
		return nil, fmt.Errorf("ai service returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result variantResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		s.aiCircuitBreaker.recordFailure()
		return nil, fmt.Errorf("decode ai response: %w", err)
	}

	s.aiCircuitBreaker.recordSuccess()
	return &result, nil
}

func intPtr(v int) *int {
	return &v
}

// safePrefix returns the first n characters of s, or all of s if shorter.
func safePrefix(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// PlatformSettings defines per-user DNA platform preferences.
type PlatformSettings struct {
	AutoEvolve               bool
	RequireApproval          bool
	SandboxValidation        bool
	DefaultCanaryPct         int
	MaxMutationsPerDay       int
	NotifyOnProposal         bool
	NotifyOnDeploy          bool
	NotifyOnRollback         bool
	AutoRollbackOnError      bool
	AutoRollbackErrorThreshold int // percentage, e.g. 5 = 5%
}

// PlatformSettingsProvider retrieves a user's DNA platform settings.
type PlatformSettingsProvider interface {
	GetPlatformSettings(ctx context.Context, userID string) (*PlatformSettings, error)
}

// DefaultPlatformSettings returns safe defaults for all settings.
func DefaultPlatformSettings() *PlatformSettings {
	return &PlatformSettings{
		AutoEvolve:               true,
		RequireApproval:          true,
		SandboxValidation:        true,
		DefaultCanaryPct:         10,
		MaxMutationsPerDay:       5,
		NotifyOnProposal:         true,
		NotifyOnDeploy:           true,
		NotifyOnRollback:         true,
		AutoRollbackOnError:      true,
		AutoRollbackErrorThreshold: 5,
	}
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if intVal, err := strconv.Atoi(v); err == nil {
			return intVal
		}
	}
	return defaultVal
}
