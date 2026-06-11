package dna

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository handles all DNA-related database operations.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new DNA repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ──────────────────────────────────────────────────────────────────────────────
// DNA Profile
// ──────────────────────────────────────────────────────────────────────────────

// DNAProfile is the master genetic identity for a function.
type DNAProfile struct {
	ID                  string          `json:"id"`
	FunctionID          string          `json:"function_id"`
	FunctionType        string          `json:"function_type"`
	TenantID            string          `json:"tenant_id"`
	Generation          int             `json:"generation"`
	FitnessScore        float64         `json:"fitness_score"`
	TotalExecutions     int64           `json:"total_executions"`
	TotalMutations      int             `json:"total_mutations"`
	AvgLatencyMs        *float64        `json:"avg_latency_ms"`
	P99LatencyMs        *float64        `json:"p99_latency_ms"`
	SuccessRate         float64         `json:"success_rate"`
	ErrorDistribution   json.RawMessage `json:"error_distribution"`
	InputPatterns       json.RawMessage `json:"input_patterns"`
	BottleneckSignature json.RawMessage `json:"bottleneck_signature"`
	DNAHash             *string         `json:"dna_hash"`
	LastAnalyzedAt      *time.Time      `json:"last_analyzed_at"`
	EvolutionEnabled    bool            `json:"evolution_enabled"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

// GetOrCreateProfile returns the DNA profile or creates one if it doesn't exist.
func (r *Repository) GetOrCreateProfile(ctx context.Context, functionID, functionType, tenantID string) (*DNAProfile, error) {
	p := &DNAProfile{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, function_id, function_type, tenant_id, generation, fitness_score,
			total_executions, total_mutations, avg_latency_ms, p99_latency_ms,
			success_rate, error_distribution, input_patterns, bottleneck_signature,
			dna_hash, last_analyzed_at, evolution_enabled, created_at, updated_at
		FROM function_dna_profiles
		WHERE function_id = $1 AND function_type = $2
	`, functionID, functionType).Scan(
		&p.ID, &p.FunctionID, &p.FunctionType, &p.TenantID, &p.Generation,
		&p.FitnessScore, &p.TotalExecutions, &p.TotalMutations,
		&p.AvgLatencyMs, &p.P99LatencyMs, &p.SuccessRate,
		&p.ErrorDistribution, &p.InputPatterns, &p.BottleneckSignature,
		&p.DNAHash, &p.LastAnalyzedAt, &p.EvolutionEnabled,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err == nil {
		return p, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("get dna profile: %w", err)
	}

	// Create new profile (ON CONFLICT handles concurrent inserts safely)
	p = &DNAProfile{
		ID:                  uuid.New().String(),
		FunctionID:          functionID,
		FunctionType:        functionType,
		TenantID:            tenantID,
		Generation:          1,
		FitnessScore:        50.0,
		SuccessRate:         1.0,
		ErrorDistribution:   json.RawMessage(`{}`),
		InputPatterns:       json.RawMessage(`[]`),
		BottleneckSignature: json.RawMessage(`[]`),
		EvolutionEnabled:    true,
	}
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO function_dna_profiles
			(id, function_id, function_type, tenant_id, generation, fitness_score,
			 success_rate, error_distribution, input_patterns, bottleneck_signature, evolution_enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (function_id, function_type) DO NOTHING
		RETURNING id, function_id, function_type, tenant_id, generation, fitness_score,
			total_executions, total_mutations, avg_latency_ms, p99_latency_ms,
			success_rate, error_distribution, input_patterns, bottleneck_signature,
			dna_hash, last_analyzed_at, evolution_enabled, created_at, updated_at
	`, p.ID, p.FunctionID, p.FunctionType, p.TenantID, p.Generation,
		p.FitnessScore, p.SuccessRate, p.ErrorDistribution,
		p.InputPatterns, p.BottleneckSignature, p.EvolutionEnabled).Scan(
		&p.ID, &p.FunctionID, &p.FunctionType, &p.TenantID, &p.Generation,
		&p.FitnessScore, &p.TotalExecutions, &p.TotalMutations,
		&p.AvgLatencyMs, &p.P99LatencyMs, &p.SuccessRate,
		&p.ErrorDistribution, &p.InputPatterns, &p.BottleneckSignature,
		&p.DNAHash, &p.LastAnalyzedAt, &p.EvolutionEnabled,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err == nil {
		return p, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("insert dna profile: %w", err)
	}
	// ON CONFLICT DO NOTHING returned no rows — another goroutine inserted it. Read it back.
	return r.GetOrCreateProfile(ctx, functionID, functionType, tenantID)
}

// UpdateProfile updates mutable fields on a DNA profile after analysis.
func (r *Repository) UpdateProfile(ctx context.Context, p *DNAProfile) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE function_dna_profiles SET
			generation = $3, fitness_score = $4, total_executions = $5,
			total_mutations = $6, avg_latency_ms = $7, p99_latency_ms = $8,
			success_rate = $9, error_distribution = $10, input_patterns = $11,
			bottleneck_signature = $12, dna_hash = $13, last_analyzed_at = $14,
			evolution_enabled = $15, updated_at = NOW()
		WHERE function_id = $1 AND function_type = $2
	`, p.FunctionID, p.FunctionType, p.Generation, p.FitnessScore,
		p.TotalExecutions, p.TotalMutations, p.AvgLatencyMs, p.P99LatencyMs,
		p.SuccessRate, p.ErrorDistribution, p.InputPatterns,
		p.BottleneckSignature, p.DNAHash, p.LastAnalyzedAt, p.EvolutionEnabled)
	return err
}

// SetEvolutionEnabled toggles evolution for a function.
func (r *Repository) SetEvolutionEnabled(ctx context.Context, functionID, functionType string, enabled bool) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE function_dna_profiles SET evolution_enabled = $3, updated_at = NOW()
		WHERE function_id = $1 AND function_type = $2
	`, functionID, functionType, enabled)
	return err
}

// GetProfileReadOnly returns the DNA profile without creating one if missing.
// Returns nil, nil if the profile does not exist.
func (r *Repository) GetProfileReadOnly(ctx context.Context, functionID, functionType string) (*DNAProfile, error) {
	p := &DNAProfile{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, function_id, function_type, tenant_id, generation, fitness_score,
			total_executions, total_mutations, avg_latency_ms, p99_latency_ms,
			success_rate, error_distribution, input_patterns, bottleneck_signature,
			dna_hash, last_analyzed_at, evolution_enabled, created_at, updated_at
		FROM function_dna_profiles
		WHERE function_id = $1 AND function_type = $2
	`, functionID, functionType).Scan(
		&p.ID, &p.FunctionID, &p.FunctionType, &p.TenantID, &p.Generation,
		&p.FitnessScore, &p.TotalExecutions, &p.TotalMutations,
		&p.AvgLatencyMs, &p.P99LatencyMs, &p.SuccessRate,
		&p.ErrorDistribution, &p.InputPatterns, &p.BottleneckSignature,
		&p.DNAHash, &p.LastAnalyzedAt, &p.EvolutionEnabled,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get dna profile: %w", err)
	}
	return p, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Execution Metrics
// ──────────────────────────────────────────────────────────────────────────────

// ExecutionMetric represents a single execution's micro-data.
type ExecutionMetric struct {
	FunctionID     string    `json:"function_id"`
	FunctionType   string    `json:"function_type"`
	ExecutionID    *string   `json:"execution_id"`
	DurationMs     int       `json:"duration_ms"`
	MemoryPeakMb   *float64  `json:"memory_peak_mb"`
	CPUTimeMs      *int      `json:"cpu_time_ms"`
	InputSizeBytes *int      `json:"input_size_bytes"`
	OutputSizeBytes *int     `json:"output_size_bytes"`
	InputShapeHash *string   `json:"input_shape_hash"`
	StatusCode     *int      `json:"status_code"`
	ErrorCategory  string    `json:"error_category"`
	ColdStart      bool      `json:"cold_start"`
	CacheHit       bool      `json:"cache_hit"`
	Region         *string   `json:"region"`
}

// InsertExecutionMetric records a single execution's DNA data.
func (r *Repository) InsertExecutionMetric(ctx context.Context, m *ExecutionMetric) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO function_dna_execution_metrics
			(function_id, function_type, execution_id, duration_ms, memory_peak_mb,
			 cpu_time_ms, input_size_bytes, output_size_bytes, input_shape_hash,
			 status_code, error_category, cold_start, cache_hit, region)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, m.FunctionID, m.FunctionType, m.ExecutionID, m.DurationMs,
		m.MemoryPeakMb, m.CPUTimeMs, m.InputSizeBytes, m.OutputSizeBytes,
		m.InputShapeHash, m.StatusCode, m.ErrorCategory,
		m.ColdStart, m.CacheHit, m.Region)
	return err
}

// AggregatedMetrics holds pre-computed statistics for a function.
type AggregatedMetrics struct {
	TotalExecutions  int64              `json:"total_executions"`
	AvgLatencyMs     float64            `json:"avg_latency_ms"`
	P50LatencyMs     float64            `json:"p50_latency_ms"`
	P95LatencyMs     float64            `json:"p95_latency_ms"`
	P99LatencyMs     float64            `json:"p99_latency_ms"`
	SuccessRate      float64            `json:"success_rate"`
	ErrorDistribution map[string]int64  `json:"error_distribution"`
	InputPatterns    []InputPattern     `json:"input_patterns"`
	ColdStartRate    float64            `json:"cold_start_rate"`
	AvgMemoryPeakMb  float64            `json:"avg_memory_peak_mb"`
}

// InputPattern represents a detected input shape and its frequency.
type InputPattern struct {
	Shape     string  `json:"shape"`
	Hash      string  `json:"hash"`
	Frequency float64 `json:"frequency"`
	Count     int64   `json:"count"`
}

// AggregateMetrics computes statistics over a time window for a function.
// Uses a single CTE query to minimize round trips to the database.
func (r *Repository) AggregateMetrics(ctx context.Context, functionID string, since time.Duration) (*AggregatedMetrics, error) {
	m := &AggregatedMetrics{
		ErrorDistribution: make(map[string]int64),
	}
	cutoff := time.Now().Add(-since)

	// Single query with CTEs for main stats, error distribution, and input patterns
	var errorDistJSON, inputPatternsJSON sql.NullString
	err := r.db.QueryRowContext(ctx, `
		WITH base AS (
			SELECT * FROM function_dna_execution_metrics
			WHERE function_id = $1 AND recorded_at > $2
		),
		main_stats AS (
			SELECT
				COUNT(*) AS total_executions,
				COALESCE(AVG(duration_ms), 0) AS avg_latency,
				COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY duration_ms), 0) AS p50,
				COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_ms), 0) AS p95,
				COALESCE(PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY duration_ms), 0) AS p99,
				COALESCE(AVG(CASE WHEN error_category = 'none' THEN 1.0 ELSE 0.0 END), 1.0) AS success_rate,
				COALESCE(AVG(CASE WHEN cold_start THEN 1.0 ELSE 0.0 END), 0.0) AS cold_start_rate,
				COALESCE(AVG(memory_peak_mb), 0) AS avg_memory
			FROM base
		),
		errors AS (
			SELECT COALESCE(jsonb_object_agg(error_category, cnt), '{}'::jsonb) AS dist
			FROM (
				SELECT error_category, COUNT(*) AS cnt
				FROM base WHERE error_category != 'none'
				GROUP BY error_category
			) sub
		),
		patterns AS (
			SELECT COALESCE(jsonb_agg(jsonb_build_object('hash', input_shape_hash, 'cnt', cnt)), '[]'::jsonb) AS pats
			FROM (
				SELECT input_shape_hash, COUNT(*) AS cnt
				FROM base WHERE input_shape_hash IS NOT NULL
				GROUP BY input_shape_hash ORDER BY cnt DESC LIMIT 10
			) sub
		)
		SELECT
			ms.total_executions, ms.avg_latency, ms.p50, ms.p95, ms.p99,
			ms.success_rate, ms.cold_start_rate, ms.avg_memory,
			e.dist::text, p.pats::text
		FROM main_stats ms
		CROSS JOIN errors e
		CROSS JOIN patterns p
	`, functionID, cutoff).Scan(
		&m.TotalExecutions, &m.AvgLatencyMs, &m.P50LatencyMs,
		&m.P95LatencyMs, &m.P99LatencyMs, &m.SuccessRate,
		&m.ColdStartRate, &m.AvgMemoryPeakMb,
		&errorDistJSON, &inputPatternsJSON,
	)
	if err != nil {
		return nil, fmt.Errorf("aggregate metrics: %w", err)
	}

	// Parse error distribution JSON
	if errorDistJSON.Valid && errorDistJSON.String != "" && errorDistJSON.String != "{}" {
		var distMap map[string]int64
		if err := json.Unmarshal([]byte(errorDistJSON.String), &distMap); err == nil {
			m.ErrorDistribution = distMap
		}
	}

	// Parse input patterns JSON
	if inputPatternsJSON.Valid && inputPatternsJSON.String != "" && inputPatternsJSON.String != "[]" {
		var rawPatterns []struct {
			Hash string `json:"hash"`
			Cnt  int64  `json:"cnt"`
		}
		if err := json.Unmarshal([]byte(inputPatternsJSON.String), &rawPatterns); err == nil {
			var totalPatterns int64
			for _, p := range rawPatterns {
				totalPatterns += p.Cnt
			}
			for _, p := range rawPatterns {
				freq := 0.0
				if totalPatterns > 0 {
					freq = float64(p.Cnt) / float64(totalPatterns)
				}
				m.InputPatterns = append(m.InputPatterns, InputPattern{
					Hash:      p.Hash,
					Count:     p.Cnt,
					Frequency: freq,
				})
			}
		}
	}

	return m, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Mutations
// ──────────────────────────────────────────────────────────────────────────────

// Mutation represents an evolution event.
type Mutation struct {
	ID                 string          `json:"id"`
	FunctionID         string          `json:"function_id"`
	FunctionType       string          `json:"function_type"`
	TenantID           string          `json:"tenant_id"`
	Generation         int             `json:"generation"`
	MutationType       string          `json:"mutation_type"`
	Status             string          `json:"status"`
	TriggerReason      *string         `json:"trigger_reason"`
	OriginalCode       *string         `json:"original_code"`        // Deprecated: nullable, use OriginalHash
	MutatedCode        *string         `json:"mutated_code"`         // Deprecated: nullable, use MutatedHash
	OriginalHash       *string         `json:"original_hash"`
	MutatedHash        *string         `json:"mutated_hash"`
	CodeHashAlgo       string          `json:"code_hash_algo"`
	OriginalCodeHash   *string         `json:"original_code_hash"`
	MutatedCodeHash    *string         `json:"mutated_code_hash"`
	CodeSizeBytes      *int            `json:"code_size_bytes"`
	LineCount          *int            `json:"line_count"`
	Diff               *string         `json:"diff"`
	EstimatedImpact    json.RawMessage `json:"estimated_impact"`
	ActualImpact       json.RawMessage `json:"actual_impact"`
	Confidence         *float64        `json:"confidence"`
	ModelUsed          *string         `json:"model_used"`
	AnalysisWindowHours *int           `json:"analysis_window_hours"`
	ExecutionsAnalyzed *int            `json:"executions_analyzed"`
	AcceptedBy         *string         `json:"accepted_by"`
	AcceptedAt         *time.Time      `json:"accepted_at"`
	DeployedAt         *time.Time      `json:"deployed_at"`
	RolledBackAt       *time.Time      `json:"rolled_back_at"`
	RejectedReason     *string         `json:"rejected_reason"`
	CreatedAt          time.Time       `json:"created_at"`
	// Payment tracking fields
	PaymentStatus       string    `json:"payment_status"`
	PaymentRetryCount   int       `json:"payment_retry_count"`
	PaymentFailedAt      *time.Time `json:"payment_failed_at"`
	PaymentFailureReason *string   `json:"payment_failure_reason"`
}

// ListMutations returns mutations for a function with optional filters.
func (r *Repository) ListMutations(ctx context.Context, functionID, status string, limit, offset int) ([]*Mutation, int, error) {
	where := "function_id = $1"
	args := []interface{}{functionID}
	argIdx := 2

	if status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	var total int
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM function_dna_mutations WHERE %s", where),
		args...,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count mutations: %w", err)
	}

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, function_id, function_type, tenant_id, generation, mutation_type,
			status, trigger_reason, original_code, mutated_code,
			original_hash, mutated_hash, code_hash_algo,
			original_code_hash, mutated_code_hash, code_size_bytes, line_count,
			estimated_impact, actual_impact, confidence,
			model_used, analysis_window_hours, executions_analyzed,
			accepted_by, accepted_at, deployed_at, rolled_back_at, rejected_reason,
			payment_status, payment_retry_count, payment_failed_at, payment_failure_reason,
			created_at
		FROM function_dna_mutations
		WHERE %s
		ORDER BY generation DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list mutations: %w", err)
	}
	defer rows.Close()

	var mutations []*Mutation
	for rows.Next() {
		m := &Mutation{}
		var actualImpact sql.NullString
		var codeHashAlgo sql.NullString
		var originalCodeHash, mutatedCodeHash sql.NullString
		var codeSizeBytes, lineCount sql.NullInt64
		var paymentStatus sql.NullString
		var paymentRetryCount sql.NullInt64
		var paymentFailedAt sql.NullTime
		var paymentFailureReason sql.NullString
		if err := rows.Scan(
			&m.ID, &m.FunctionID, &m.FunctionType, &m.TenantID, &m.Generation,
			&m.MutationType, &m.Status, &m.TriggerReason,
			&m.OriginalCode, &m.MutatedCode,
			&m.OriginalHash, &m.MutatedHash, &codeHashAlgo,
			&originalCodeHash, &mutatedCodeHash, &codeSizeBytes, &lineCount,
			&m.EstimatedImpact, &actualImpact, &m.Confidence,
			&m.ModelUsed, &m.AnalysisWindowHours, &m.ExecutionsAnalyzed,
			&m.AcceptedBy, &m.AcceptedAt, &m.DeployedAt, &m.RolledBackAt,
			&m.RejectedReason, &paymentStatus, &paymentRetryCount, &paymentFailedAt,
			&paymentFailureReason, &m.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		if actualImpact.Valid {
			m.ActualImpact = json.RawMessage(actualImpact.String)
		}
		if codeHashAlgo.Valid {
			m.CodeHashAlgo = codeHashAlgo.String
		}
		if originalCodeHash.Valid {
			m.OriginalCodeHash = &originalCodeHash.String
		}
		if mutatedCodeHash.Valid {
			m.MutatedCodeHash = &mutatedCodeHash.String
		}
		if codeSizeBytes.Valid {
			v := int(codeSizeBytes.Int64)
			m.CodeSizeBytes = &v
		}
		if lineCount.Valid {
			v := int(lineCount.Int64)
			m.LineCount = &v
		}
		if paymentStatus.Valid {
			m.PaymentStatus = paymentStatus.String
		}
		if paymentRetryCount.Valid {
			m.PaymentRetryCount = int(paymentRetryCount.Int64)
		}
		if paymentFailedAt.Valid {
			m.PaymentFailedAt = &paymentFailedAt.Time
		}
		if paymentFailureReason.Valid {
			m.PaymentFailureReason = &paymentFailureReason.String
		}
		mutations = append(mutations, m)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("list mutations iteration: %w", err)
	}
	return mutations, total, nil
}

// GetMutation returns a single mutation with full code details.
func (r *Repository) GetMutation(ctx context.Context, mutationID string) (*Mutation, error) {
	m := &Mutation{}
	var actualImpact sql.NullString
	var codeHashAlgo sql.NullString
	var originalCodeHash, mutatedCodeHash sql.NullString
	var codeSizeBytes, lineCount sql.NullInt64
	var paymentStatus sql.NullString
	var paymentRetryCount sql.NullInt64
	var paymentFailedAt sql.NullTime
	var paymentFailureReason sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, function_id, function_type, tenant_id, generation, mutation_type,
			status, trigger_reason, original_code, mutated_code, original_hash,
			mutated_hash, code_hash_algo, original_code_hash, mutated_code_hash,
			code_size_bytes, line_count, diff, estimated_impact, actual_impact,
			confidence, model_used, analysis_window_hours, executions_analyzed,
			accepted_by, accepted_at, deployed_at, rolled_back_at, rejected_reason,
			payment_status, payment_retry_count, payment_failed_at, payment_failure_reason,
			created_at
		FROM function_dna_mutations WHERE id = $1
	`, mutationID).Scan(
		&m.ID, &m.FunctionID, &m.FunctionType, &m.TenantID, &m.Generation,
		&m.MutationType, &m.Status, &m.TriggerReason,
		&m.OriginalCode, &m.MutatedCode, &m.OriginalHash, &m.MutatedHash,
		&codeHashAlgo, &originalCodeHash, &mutatedCodeHash, &codeSizeBytes, &lineCount,
		&m.Diff, &m.EstimatedImpact, &actualImpact, &m.Confidence,
		&m.ModelUsed, &m.AnalysisWindowHours, &m.ExecutionsAnalyzed,
		&m.AcceptedBy, &m.AcceptedAt, &m.DeployedAt, &m.RolledBackAt,
		&m.RejectedReason, &paymentStatus, &paymentRetryCount, &paymentFailedAt,
		&paymentFailureReason, &m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get mutation: %w", err)
	}
	if actualImpact.Valid {
		m.ActualImpact = json.RawMessage(actualImpact.String)
	}
	if codeHashAlgo.Valid {
		m.CodeHashAlgo = codeHashAlgo.String
	}
	if originalCodeHash.Valid {
		m.OriginalCodeHash = &originalCodeHash.String
	}
	if mutatedCodeHash.Valid {
		m.MutatedCodeHash = &mutatedCodeHash.String
	}
	if codeSizeBytes.Valid {
		v := int(codeSizeBytes.Int64)
		m.CodeSizeBytes = &v
	}
	if lineCount.Valid {
		v := int(lineCount.Int64)
		m.LineCount = &v
	}
	if paymentStatus.Valid {
		m.PaymentStatus = paymentStatus.String
	}
	if paymentRetryCount.Valid {
		m.PaymentRetryCount = int(paymentRetryCount.Int64)
	}
	if paymentFailedAt.Valid {
		m.PaymentFailedAt = &paymentFailedAt.Time
	}
	if paymentFailureReason.Valid {
		m.PaymentFailureReason = &paymentFailureReason.String
	}
	return m, nil
}

// CreateMutation inserts a new proposed mutation.
func (r *Repository) CreateMutation(ctx context.Context, m *Mutation) error {
	m.ID = uuid.New().String()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO function_dna_mutations
			(id, function_id, function_type, tenant_id, generation, mutation_type,
			 status, trigger_reason, original_code, mutated_code, original_hash,
			 mutated_hash, diff, estimated_impact, confidence, model_used,
			 analysis_window_hours, executions_analyzed)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`, m.ID, m.FunctionID, m.FunctionType, m.TenantID, m.Generation,
		m.MutationType, m.Status, m.TriggerReason,
		m.OriginalCode, m.MutatedCode, m.OriginalHash, m.MutatedHash,
		m.Diff, m.EstimatedImpact, m.Confidence, m.ModelUsed,
		m.AnalysisWindowHours, m.ExecutionsAnalyzed)
	return err
}

// validMutationStatuses is the allowlist for mutation status transitions.
var validMutationStatuses = map[string]bool{
	"proposed":                 true,
	"accepted_pending_payment": true,
	"accepted":                 true,
	"rejected":                 true,
	"payment_failed":           true,
	"deploying":                true,
	"deployed":                 true,
	"rolled_back":              true,
}

// UpdateMutationStatus updates a mutation's status and related timestamps.
func (r *Repository) UpdateMutationStatus(ctx context.Context, mutationID, status string, extra map[string]interface{}) error {
	if !validMutationStatuses[status] {
		return fmt.Errorf("invalid mutation status: %s", status)
	}
	query := "UPDATE function_dna_mutations SET status = $2"
	args := []interface{}{mutationID, status}
	argIdx := 3

	switch status {
	case "accepted_pending_payment":
		if v, ok := extra["accepted_by"]; ok {
			query += fmt.Sprintf(", accepted_by = $%d, accepted_at = NOW()", argIdx)
			args = append(args, v)
			argIdx++
		}
	case "accepted":
		if v, ok := extra["accepted_by"]; ok {
			query += fmt.Sprintf(", accepted_by = $%d, accepted_at = NOW()", argIdx)
			args = append(args, v)
			argIdx++
		}
	case "payment_failed":
		if v, ok := extra["rejected_reason"]; ok {
			query += fmt.Sprintf(", rejected_reason = $%d", argIdx)
			args = append(args, v)
			argIdx++
		}
	case "deploying", "deployed":
		query += ", deployed_at = NOW()"
	case "rolled_back":
		query += ", rolled_back_at = NOW()"
		if v, ok := extra["rollback_reason"]; ok {
			query += fmt.Sprintf(", rejected_reason = $%d", argIdx)
			args = append(args, v)
			argIdx++
		}
	case "rejected":
		if v, ok := extra["reason"]; ok {
			query += fmt.Sprintf(", rejected_reason = $%d", argIdx)
			args = append(args, v)
			argIdx++
		}
	}

	query += fmt.Sprintf(" WHERE id = $1")
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// ──────────────────────────────────────────────────────────────────────────────
// Analysis Queue
// ──────────────────────────────────────────────────────────────────────────────

// EnqueueAnalysis adds a function to the analysis queue.
func (r *Repository) EnqueueAnalysis(ctx context.Context, functionID, functionType, tenantID string, priority int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO function_dna_analysis_queue (function_id, analysis_type, tenant_id, priority)
		VALUES ($1, $2, $3, $4)
	`, functionID, functionType, tenantID, priority)
	return err
}

// DequeueAnalysis picks the next pending analysis task (SELECT ... FOR UPDATE SKIP LOCKED).
func (r *Repository) DequeueAnalysis(ctx context.Context) (string, string, string, error) {
	var id, functionID, functionType, tenantID string
	err := r.db.QueryRowContext(ctx, `
		UPDATE function_dna_analysis_queue
		SET status = 'processing', started_at = NOW(), attempts = attempts + 1
		WHERE id = (
			SELECT id FROM function_dna_analysis_queue
			WHERE status = 'pending' AND queued_at <= NOW() AND attempts < max_attempts
			ORDER BY priority ASC, queued_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, function_id, analysis_type, tenant_id
	`).Scan(&id, &functionID, &functionType, &tenantID)
	if err == sql.ErrNoRows {
		return "", "", "", nil
	}
	return id, functionID, functionType, err
}

// CompleteAnalysis marks an analysis task as completed.
func (r *Repository) CompleteAnalysis(ctx context.Context, queueID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE function_dna_analysis_queue SET status = 'completed', completed_at = NOW()
		WHERE id = $1
	`, queueID)
	return err
}

// FailAnalysis marks an analysis task as failed.
func (r *Repository) FailAnalysis(ctx context.Context, queueID, errMsg string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE function_dna_analysis_queue SET status = 'failed', last_error = $2
		WHERE id = $1
	`, queueID, errMsg)
	return err
}

// ──────────────────────────────────────────────────────────────────────────────
// Enterprise Insights
// ──────────────────────────────────────────────────────────────────────────────

// TenantInsights holds enterprise-wide DNA analytics.
type TenantInsights struct {
	TotalFunctionsAnalyzed   int              `json:"total_functions_analyzed"`
	TotalMutationsProposed   int              `json:"total_mutations_proposed"`
	TotalMutationsAccepted   int              `json:"total_mutations_accepted"`
	AvgFitnessScore          float64          `json:"avg_fitness_score"`
	AvgLatencyImprovementPct float64          `json:"avg_latency_improvement_pct"`
	TotalCostSavingsUSD      float64          `json:"total_cost_savings_usd"`
	TopBottleneckCategories  json.RawMessage  `json:"top_bottleneck_categories"`
	EvolutionLeaderboard     json.RawMessage  `json:"evolution_leaderboard"`
}

// GetTenantInsights computes enterprise-wide DNA analytics for a tenant.
func (r *Repository) GetTenantInsights(ctx context.Context, tenantID string, since time.Duration) (*TenantInsights, error) {
	ins := &TenantInsights{}
	cutoff := time.Now().Add(-since)

	// Profile aggregates
	err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COALESCE(AVG(fitness_score), 50),
			COALESCE(SUM(total_mutations), 0)
		FROM function_dna_profiles
		WHERE tenant_id = $1
	`, tenantID).Scan(&ins.TotalFunctionsAnalyzed, &ins.AvgFitnessScore, &ins.TotalMutationsProposed)
	if err != nil {
		return nil, fmt.Errorf("profile aggregates: %w", err)
	}

	// Mutation acceptance
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM function_dna_mutations
		WHERE tenant_id = $1 AND status = 'accepted' AND created_at > $2
	`, tenantID, cutoff).Scan(&ins.TotalMutationsAccepted)
	if err != nil {
		return nil, fmt.Errorf("mutation count: %w", err)
	}

	// Average latency improvement from actual_impact
	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(AVG((actual_impact->>'latency_improvement_pct')::float), 0)
		FROM function_dna_mutations
		WHERE tenant_id = $1 AND actual_impact IS NOT NULL AND created_at > $2
	`, tenantID, cutoff).Scan(&ins.AvgLatencyImprovementPct)
	if err != nil {
		return nil, fmt.Errorf("latency improvement: %w", err)
	}

	// Evolution leaderboard
	rows, err := r.db.QueryContext(ctx, `
		SELECT function_id, generation, fitness_score
		FROM function_dna_profiles
		WHERE tenant_id = $1
		ORDER BY generation DESC, fitness_score DESC
		LIMIT 10
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("leaderboard: %w", err)
	}
	defer rows.Close()
	var leaderboard []map[string]interface{}
	for rows.Next() {
		var fid string
		var gen int
		var fitness float64
		if err := rows.Scan(&fid, &gen, &fitness); err != nil {
			return nil, err
		}
		leaderboard = append(leaderboard, map[string]interface{}{
			"function_id": fid, "generation": gen, "fitness_score": fitness,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("leaderboard iteration: %w", err)
	}
	ins.EvolutionLeaderboard, _ = json.Marshal(leaderboard)

	// Top bottleneck categories
	rows2, err := r.db.QueryContext(ctx, `
		SELECT category, cnt FROM (
			SELECT jsonb_object_keys(bottleneck_signature) as category,
				SUM(jsonb_array_length(bottleneck_signature)) as cnt
			FROM function_dna_profiles
			WHERE tenant_id = $1 AND bottleneck_signature != '[]'::jsonb
			GROUP BY category
			ORDER BY cnt DESC
			LIMIT 5
		) sub
	`, tenantID)
	if err == nil {
		defer rows2.Close()
		var bottlenecks []map[string]interface{}
		for rows2.Next() {
			var cat string
			var cnt int
			if err := rows2.Scan(&cat, &cnt); err != nil {
				return nil, fmt.Errorf("bottleneck scan: %w", err)
			}
			bottlenecks = append(bottlenecks, map[string]interface{}{
				"category": cat, "count": cnt,
			})
		}
		if err := rows2.Err(); err != nil {
			return nil, fmt.Errorf("bottleneck iteration: %w", err)
		}
		ins.TopBottleneckCategories, _ = json.Marshal(bottlenecks)
	}

	return ins, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Partition Management
// ──────────────────────────────────────────────────────────────────────────────

// CreateFuturePartitions creates monthly partitions for the execution metrics table
// for the given number of months ahead from the current month.
func (r *Repository) CreateFuturePartitions(ctx context.Context, monthsAhead int) (int, error) {
	created := 0
	now := time.Now()

	for i := 0; i < monthsAhead; i++ {
		t := now.AddDate(0, i, 0)
		partName := fmt.Sprintf("function_dna_execution_metrics_%s", t.Format("2006_01"))
		fromDate := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		toDate := fromDate.AddDate(0, 1, 0)

		_, err := r.db.ExecContext(ctx, fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s
			PARTITION OF function_dna_execution_metrics
			FOR VALUES FROM ('%s') TO ('%s')
		`, partName, fromDate.Format("2006-01-02"), toDate.Format("2006-01-02")))
		if err != nil {
			return created, fmt.Errorf("create partition %s: %w", partName, err)
		}
		created++
	}

	return created, nil
}

// DropOldPartitions drops monthly partitions for execution metrics older than
// the given number of months. Returns the number of partitions dropped.
func (r *Repository) DropOldPartitions(ctx context.Context, retentionMonths int) (int, error) {
	cutoff := time.Now().AddDate(0, -retentionMonths, 0)

	rows, err := r.db.QueryContext(ctx, `
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public'
		AND tablename LIKE 'function_dna\_execution\_metrics\_%'
		AND tablename < $1
		ORDER BY tablename
	`, fmt.Sprintf("function_dna_execution_metrics_%s", cutoff.Format("2006_01")))
	if err != nil {
		return 0, fmt.Errorf("list old partitions: %w", err)
	}
	defer rows.Close()

	var partitions []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return 0, err
		}
		partitions = append(partitions, name)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("list old partitions iteration: %w", err)
	}

	dropped := 0
	for _, partName := range partitions {
		_, err := r.db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", partName))
		if err != nil {
			return dropped, fmt.Errorf("drop partition %s: %w", partName, err)
		}
		dropped++
	}

	return dropped, nil
}

// ListPartitions returns the names of all current execution metric partitions.
func (r *Repository) ListPartitions(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public'
		AND tablename LIKE 'function_dna\_execution\_metrics\_%'
		ORDER BY tablename
	`)
	if err != nil {
		return nil, fmt.Errorf("list partitions: %w", err)
	}
	defer rows.Close()

	var partitions []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		partitions = append(partitions, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list partitions iteration: %w", err)
	}
	return partitions, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// DNA Insights (Pre-computed Enterprise Analytics)
// ──────────────────────────────────────────────────────────────────────────────

// InsertInsight inserts a pre-computed enterprise insight row.
func (r *Repository) InsertInsight(ctx context.Context, ins *TenantInsights, tenantID string, periodStart, periodEnd time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO function_dna_insights
			(tenant_id, period_start, period_end, total_functions_analyzed,
			 total_mutations_proposed, total_mutations_accepted, avg_fitness_score,
			 avg_latency_improvement_pct, total_cost_savings_usd,
			 top_bottleneck_categories, evolution_leaderboard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, tenantID, periodStart, periodEnd,
		ins.TotalFunctionsAnalyzed, ins.TotalMutationsProposed, ins.TotalMutationsAccepted,
		ins.AvgFitnessScore, ins.AvgLatencyImprovementPct, ins.TotalCostSavingsUSD,
		ins.TopBottleneckCategories, ins.EvolutionLeaderboard)
	return err
}

// GetDistinctTenantIDs returns all unique tenant IDs from the DNA profiles table.
func (r *Repository) GetDistinctTenantIDs(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT tenant_id FROM function_dna_profiles WHERE tenant_id != ''
	`)
	if err != nil {
		return nil, fmt.Errorf("get distinct tenants: %w", err)
	}
	defer rows.Close()

	var tenants []string
	for rows.Next() {
		var tid string
		if err := rows.Scan(&tid); err != nil {
			return nil, err
		}
		tenants = append(tenants, tid)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get distinct tenants iteration: %w", err)
	}
	return tenants, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Payment Reconciliation
// ──────────────────────────────────────────────────────────────────────────────

// PendingPayment represents a mutation with pending payment that needs reconciliation.
type PendingPayment struct {
	ID                    string
	FunctionID            string
	TenantID              string
	AcceptedBy            string
	PaymentRetryCount     int
	PaymentFailureReason  *string
	CreatedAt             time.Time
}

// GetPendingPayments retrieves mutations with pending payment status for reconciliation.
func (r *Repository) GetPendingPayments(ctx context.Context, maxAge time.Duration, maxRetries int) ([]PendingPayment, error) {
	cutoff := time.Now().Add(-maxAge)
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, function_id, tenant_id, accepted_by, payment_retry_count,
			   payment_failure_reason, created_at
		FROM function_dna_mutations
		WHERE status = 'accepted_pending_payment'
		  AND payment_status = 'pending'
		  AND created_at < $1
		  AND payment_retry_count < $2
		ORDER BY created_at ASC
		LIMIT 100
	`, cutoff, maxRetries)
	if err != nil {
		return nil, fmt.Errorf("get pending payments: %w", err)
	}
	defer rows.Close()

	var payments []PendingPayment
	for rows.Next() {
		var p PendingPayment
		if err := rows.Scan(&p.ID, &p.FunctionID, &p.TenantID, &p.AcceptedBy,
			&p.PaymentRetryCount, &p.PaymentFailureReason, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan pending payment: %w", err)
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}

// UpdateMutationPaymentStatus updates the payment status and related fields for a mutation.
func (r *Repository) UpdateMutationPaymentStatus(ctx context.Context, mutationID string, paymentStatus string, failureReason string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE function_dna_mutations
		SET payment_status = $2,
			payment_failure_reason = $3,
			payment_retry_count = payment_retry_count + 1,
			payment_failed_at = CASE WHEN $2 = 'failed' THEN NOW() ELSE payment_failed_at END
		WHERE id = $1
	`, mutationID, paymentStatus, failureReason)
	return err
}

// MarkMutationAsReconciled marks a mutation as reconciled after successful payment retry.
func (r *Repository) MarkMutationAsReconciled(ctx context.Context, mutationID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE function_dna_mutations
		SET status = 'accepted',
			payment_status = 'reconciled',
			payment_failure_reason = NULL
		WHERE id = $1 AND status = 'accepted_pending_payment'
	`, mutationID)
	return err
}

// GetFailedPaymentsForManualReview retrieves mutations with failed payments that need manual review.
func (r *Repository) GetFailedPaymentsForManualReview(ctx context.Context) ([]PendingPayment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, function_id, tenant_id, accepted_by, payment_retry_count,
			   payment_failure_reason, created_at
		FROM function_dna_mutations
		WHERE status = 'accepted_pending_payment'
		  AND payment_status = 'failed'
		ORDER BY created_at ASC
		LIMIT 100
	`)
	if err != nil {
		return nil, fmt.Errorf("get failed payments: %w", err)
	}
	defer rows.Close()

	var payments []PendingPayment
	for rows.Next() {
		var p PendingPayment
		if err := rows.Scan(&p.ID, &p.FunctionID, &p.TenantID, &p.AcceptedBy,
			&p.PaymentRetryCount, &p.PaymentFailureReason, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan failed payment: %w", err)
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}
