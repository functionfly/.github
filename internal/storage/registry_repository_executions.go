package storage

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RecordExecution records a function execution
func (r *RegistryRepository) RecordExecution(exec *RegistryFunctionExecution) error {
	exec.ID = uuid.New()
	exec.Timestamp = time.Now()

	if err := r.db.Create(exec).Error; err != nil {
		return fmt.Errorf("failed to record execution: %w", err)
	}

	return nil
}

// RecordResourceUsage records resource usage for an execution
func (r *RegistryRepository) RecordResourceUsage(usage *ExecutionResourceUsage) error {
	usage.ID = uuid.New()
	usage.CreatedAt = time.Now()

	if err := r.db.Create(usage).Error; err != nil {
		return fmt.Errorf("failed to record resource usage: %w", err)
	}

	return nil
}

// GetFunctionStats retrieves execution stats for a function
func (r *RegistryRepository) GetFunctionStats(functionID uuid.UUID, since time.Time) (totalCalls int, successRate float64, avgLatencyMs int, p95LatencyMs int, err error) {
	var result struct {
		TotalCalls  int     `json:"total_calls"`
		SuccessRate float64 `json:"success_rate"`
		AvgLatency  int     `json:"avg_latency"`
		P95Latency  int     `json:"p95_latency"`
	}

	query := `
		SELECT
			COUNT(*) as total_calls,
			COALESCE(SUM(CASE WHEN outcome = 'success' THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0) as success_rate,
			COALESCE(AVG(duration_ms), 0) as avg_latency,
			COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_ms)::INTEGER, 0) as p95_latency
		FROM registry_function_executions
		WHERE function_id = ? AND timestamp > ?
	`

	if err := r.db.Raw(query, functionID, since).Scan(&result).Error; err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to get function stats: %w", err)
	}

	return result.TotalCalls, result.SuccessRate, result.AvgLatency, result.P95Latency, nil
}

// GetFunctionTrustStats retrieves extended stats including trust score components
func (r *RegistryRepository) GetFunctionTrustStats(functionID uuid.UUID, since time.Time) (
	totalCalls int, successRate float64, avgLatencyMs int, p50LatencyMs int, p95LatencyMs int,
	timeoutRate float64, errorRate float64, err error) {
	var result struct {
		TotalCalls  int     `json:"total_calls"`
		SuccessRate float64 `json:"success_rate"`
		AvgLatency  int     `json:"avg_latency"`
		P50Latency  int     `json:"p50_latency"`
		P95Latency  int     `json:"p95_latency"`
		TimeoutRate float64 `json:"timeout_rate"`
		ErrorRate   float64 `json:"error_rate"`
	}

	query := `
		SELECT
			COUNT(*) as total_calls,
			COALESCE(SUM(CASE WHEN outcome = 'success' THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0) as success_rate,
			COALESCE(AVG(duration_ms), 0) as avg_latency,
			COALESCE(PERCENTILE_CONT(0.50) WITHIN GROUP (ORDER BY duration_ms)::INTEGER, 0) as p50_latency,
			COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_ms)::INTEGER, 0) as p95_latency,
			COALESCE(SUM(CASE WHEN outcome = 'timeout' THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0) as timeout_rate,
			COALESCE(SUM(CASE WHEN outcome = 'error' THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0) as error_rate
		FROM registry_function_executions
		WHERE function_id = ? AND timestamp > ?
	`

	if err := r.db.Raw(query, functionID, since).Scan(&result).Error; err != nil {
		return 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("failed to get function trust stats: %w", err)
	}

	return result.TotalCalls, result.SuccessRate, result.AvgLatency, result.P50Latency,
		result.P95Latency, result.TimeoutRate, result.ErrorRate, nil
}

// GetConsumerDiversity returns unique caller metrics for a function
func (r *RegistryRepository) GetConsumerDiversity(functionID uuid.UUID, since time.Time) (uniqueIPs int, uniqueTenants int, uniqueUsers int, err error) {
	var result struct {
		UniqueIPs     int `json:"unique_ips"`
		UniqueTenants int `json:"unique_tenants"`
		UniqueUsers   int `json:"unique_users"`
	}

	query := `
		SELECT
			COUNT(DISTINCT caller_ip) as unique_ips,
			COUNT(DISTINCT tenant_id) as unique_tenants,
			COUNT(DISTINCT user_id) as unique_users
		FROM registry_function_executions
		WHERE function_id = ? AND timestamp > ?
			AND caller_ip IS NOT NULL
	`

	if err := r.db.Raw(query, functionID, since).Scan(&result).Error; err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get consumer diversity: %w", err)
	}

	return result.UniqueIPs, result.UniqueTenants, result.UniqueUsers, nil
}

// CreateExecutionPublic creates a new shareable execution record
func (r *RegistryRepository) CreateExecutionPublic(exec *RegistryExecutionPublic) error {
	exec.ID = uuid.New()
	exec.CreatedAt = time.Now()

	if err := r.db.Create(exec).Error; err != nil {
		return fmt.Errorf("failed to create public execution: %w", err)
	}

	return nil
}

// GetExecutionPublicByID retrieves a shareable execution by its public ID
func (r *RegistryRepository) GetExecutionPublicByID(publicID string) (*RegistryExecutionPublic, error) {
	var exec RegistryExecutionPublic
	if err := r.db.Where("public_id = ? AND shareable = ?", publicID, true).First(&exec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("execution not found or not shareable")
		}
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	return &exec, nil
}

// GetExecutionCountForVersion returns the total number of executions for a specific function version
func (r *RegistryRepository) GetExecutionCountForVersion(functionID uuid.UUID, version string) (int, error) {
	var count int64
	if err := r.db.Model(&RegistryFunctionExecution{}).Where("function_id = ? AND version = ?", functionID, version).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count executions: %w", err)
	}
	return int(count), nil
}

// GetLastVerificationTimeForVersion returns the most recent verification time for a specific function version
func (r *RegistryRepository) GetLastVerificationTimeForVersion(functionID uuid.UUID, version string) (*time.Time, error) {
	var exec RegistryFunctionExecution
	if err := r.db.Where("function_id = ? AND version = ? AND verified_at IS NOT NULL", functionID, version).
		Order("verified_at DESC").First(&exec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No verification found
		}
		return nil, fmt.Errorf("failed to get last verification time: %w", err)
	}
	return &exec.VerifiedAt.Time, nil
}

// GetRecentVerificationFailureRate returns the failure rate for verifications in the last 7 days for a function version
func (r *RegistryRepository) GetRecentVerificationFailureRate(functionID uuid.UUID, version string) (float64, error) {
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	var total int64
	var failed int64

	// Count total verifications in the last 7 days
	if err := r.db.Model(&RegistryFunctionExecution{}).
		Where("function_id = ? AND version = ? AND verified_at IS NOT NULL AND verified_at > ?",
			functionID, version, sevenDaysAgo).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("failed to count recent verifications: %w", err)
	}

	if total == 0 {
		return 0, nil // No recent verifications
	}

	// Count failed verifications in the last 7 days
	if err := r.db.Model(&RegistryFunctionExecution{}).
		Where("function_id = ? AND version = ? AND verified_at IS NOT NULL AND verified_at > ? AND verification_status = ?",
			functionID, version, sevenDaysAgo, "failed").Count(&failed).Error; err != nil {
		return 0, fmt.Errorf("failed to count recent verification failures: %w", err)
	}

	return float64(failed) / float64(total), nil
}

// GetExecutionCountByUserInTimeWindow returns the number of executions by a user in the specified time window
func (r *RegistryRepository) GetExecutionCountByUserInTimeWindow(userID uuid.UUID, since time.Time) (int64, error) {
	var count int64
	if err := r.db.Model(&RegistryFunctionExecution{}).
		Where("user_id = ? AND timestamp >= ?", userID, since).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count executions by user: %w", err)
	}
	return count, nil
}

// GetExecutionCountByIPInTimeWindow returns the number of executions by IP address in the specified time window
func (r *RegistryRepository) GetExecutionCountByIPInTimeWindow(ipAddress string, since time.Time) (int64, error) {
	var count int64
	if err := r.db.Model(&RegistryFunctionExecution{}).
		Where("caller_ip = ? AND timestamp >= ?", ipAddress, since).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count executions by IP: %w", err)
	}
	return count, nil
}

// GetRecentExecutionRate returns the execution rate (executions per minute) for a user over the last time window
func (r *RegistryRepository) GetRecentExecutionRate(userID uuid.UUID, windowMinutes int) (float64, error) {
	since := time.Now().Add(-time.Duration(windowMinutes) * time.Minute)
	var count int64
	if err := r.db.Model(&RegistryFunctionExecution{}).
		Where("user_id = ? AND timestamp >= ?", userID, since).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count recent executions: %w", err)
	}
	return float64(count) / float64(windowMinutes), nil
}

// GetExecutionSpikes returns executions that exceed normal patterns for rate spike detection
func (r *RegistryRepository) GetExecutionSpikes(userID *uuid.UUID, ipAddress string, thresholdMultiplier float64, windowMinutes int) ([]RegistryFunctionExecution, error) {
	since := time.Now().Add(-time.Duration(windowMinutes) * time.Minute)

	// Get baseline rate (average executions per minute over the last 24 hours)
	baselineWindow := time.Now().Add(-24 * time.Hour)
	var baselineCount int64
	query := r.db.Model(&RegistryFunctionExecution{}).Where("timestamp >= ?", baselineWindow)
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else {
		query = query.Where("caller_ip = ?", ipAddress)
	}
	if err := query.Count(&baselineCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get baseline count: %w", err)
	}
	baselineRate := float64(baselineCount) / (24 * 60) // executions per minute over 24 hours

	// Get current window executions
	var executions []RegistryFunctionExecution
	query = r.db.Model(&RegistryFunctionExecution{}).Where("timestamp >= ?", since)
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else {
		query = query.Where("caller_ip = ?", ipAddress)
	}
	if err := query.Find(&executions).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent executions: %w", err)
	}

	// Filter for spikes (current rate exceeds baseline by threshold)
	currentRate := float64(len(executions)) / float64(windowMinutes)
	if currentRate > baselineRate*thresholdMultiplier {
		return executions, nil
	}

	return []RegistryFunctionExecution{}, nil
}