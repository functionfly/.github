package registry

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// FraudDetectionResult holds detected fraud patterns.
type FraudDetectionResult struct {
	BotPatterns        []BotPatternRow
	IPClusters         []IPClusterRow
	WashUsagePatterns  []WashUsageRow
	FakeDiversityAlerts []FakeDiversityRow
	Summary            FraudSummaryCounts
}

type BotPatternRow struct {
	PatternType       string
	ConfidenceScore   int
	AffectedFunctions []string
	AffectedTenants   []string
	Pattern           string
	DetectedAt        time.Time
}

type IPClusterRow struct {
	IPRange           string
	AssociatedTenants []string
	RiskLevel         string
	CommonPatterns    []string
	FirstSeen         time.Time
	LastSeen          time.Time
}

type WashUsageRow struct {
	TenantA              string
	TenantB              string
	Function             string
	Pattern              string
	Confidence           int
	ReciprocalExecutions int
	DetectedAt           time.Time
}

type FakeDiversityRow struct {
	TenantGroup    []string
	Indicators     []string
	RiskLevel      string
	FirstSeen      time.Time
	LastSeen       time.Time
	FunctionCount  int
	TotalExecs     int
}

type FraudSummaryCounts struct {
	TotalBotPatterns  int
	HighRiskClusters  int
	SuspiciousTenants int
	WashUsageDetected int
}

// DetectFraudPatterns runs basic fraud detection algorithms on registry data.
func (r *RegistryRepository) DetectFraudPatterns() (*FraudDetectionResult, error) {
	result := &FraudDetectionResult{}

	// Detect bot patterns: functions with low diversity but high execution counts
	botPatterns, err := r.detectBotPatterns()
	if err != nil {
		return nil, fmt.Errorf("failed to detect bot patterns: %w", err)
	}
	result.BotPatterns = botPatterns

	// Detect IP clusters: group by IP ranges and look for suspicious patterns
	ipClusters, err := r.detectIPClusters()
	if err != nil {
		return nil, fmt.Errorf("failed to detect IP clusters: %w", err)
	}
	result.IPClusters = ipClusters

	// Detect wash usage patterns (simplified)
	washPatterns, err := r.detectWashUsagePatterns()
	if err != nil {
		return nil, fmt.Errorf("failed to detect wash usage patterns: %w", err)
	}
	result.WashUsagePatterns = washPatterns

	// Detect fake diversity: coordinated tenant groups with artificial diversity signals
	fakeDiversity, err := r.detectFakeDiversity()
	if err != nil {
		return nil, fmt.Errorf("failed to detect fake diversity: %w", err)
	}
	result.FakeDiversityAlerts = fakeDiversity

	// Calculate summary counts
	result.Summary = FraudSummaryCounts{
		TotalBotPatterns:   len(botPatterns),
		HighRiskClusters:  countHighRiskClusters(ipClusters),
		SuspiciousTenants: calculateSuspiciousTenants(botPatterns, ipClusters),
		WashUsageDetected: len(washPatterns),
	}

	return result, nil
}

// detectBotPatterns looks for functions with suspicious execution patterns.
func (r *RegistryRepository) detectBotPatterns() ([]BotPatternRow, error) {
	var patterns []BotPatternRow

	// Simple pattern: functions with very low tenant diversity (< 3 unique tenants)
	// but high execution counts (> 100 executions in last 30 days)
	query := `
		WITH function_stats AS (
			SELECT
				e.function_id,
				f.name as function_name,
				f.author as tenant_name,
				COUNT(DISTINCT e.tenant_id) as unique_tenants,
				COUNT(*) as total_executions,
				SUM(CASE WHEN e.outcome = 'success' THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as success_rate
			FROM registry_function_executions e
			JOIN registry_functions f ON e.function_id = f.id
			WHERE e.timestamp > NOW() - INTERVAL '30 days'
			GROUP BY e.function_id, f.name, f.author
			HAVING COUNT(*) > 100 AND COUNT(DISTINCT e.tenant_id) < 3
		)
		SELECT
			function_id,
			function_name,
			tenant_name,
			unique_tenants,
			total_executions,
			success_rate
		FROM function_stats
		ORDER BY total_executions DESC
		LIMIT 10
	`

	var rows []struct {
		FunctionID      uuid.UUID `json:"function_id"`
		FunctionName    string    `json:"function_name"`
		TenantName      string    `json:"tenant_name"`
		UniqueTenants   int       `json:"unique_tenants"`
		TotalExecutions int       `json:"total_executions"`
		SuccessRate     float64   `json:"success_rate"`
	}

	if err := r.db.Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		confidence := 50
		if row.UniqueTenants == 1 {
			confidence = 90
		} else if row.UniqueTenants == 2 {
			confidence = 70
		}

		patterns = append(patterns, BotPatternRow{
			PatternType:       "low_diversity_execution",
			ConfidenceScore:   confidence,
			AffectedFunctions: []string{row.FunctionName},
			AffectedTenants:   []string{row.TenantName},
			Pattern:           fmt.Sprintf("High execution volume (%d) with low tenant diversity (%d unique tenants)", row.TotalExecutions, row.UniqueTenants),
			DetectedAt:        time.Now(),
		})
	}

	return patterns, nil
}

// detectIPClusters groups executions by IP ranges and identifies suspicious clusters.
func (r *RegistryRepository) detectIPClusters() ([]IPClusterRow, error) {
	var clusters []IPClusterRow

	// Simple IP clustering: group by /24 subnets and look for patterns
	query := `
		WITH ip_clusters AS (
			SELECT
				SUBSTRING(caller_ip FROM '^(\d+\.\d+\.\d+)\.') || '.0/24' as ip_range,
				COUNT(DISTINCT tenant_id) as tenant_count,
				COUNT(DISTINCT function_id) as function_count,
				COUNT(*) as execution_count,
				MIN(timestamp) as first_seen,
				MAX(timestamp) as last_seen,
				array_agg(DISTINCT tenant_id::text) FILTER (WHERE tenant_id IS NOT NULL) as tenant_ids
			FROM registry_function_executions
			WHERE caller_ip IS NOT NULL AND caller_ip != ''
				AND timestamp > NOW() - INTERVAL '7 days'
			GROUP BY SUBSTRING(caller_ip FROM '^(\d+\.\d+\.\d+)\.')
			HAVING COUNT(DISTINCT tenant_id) > 5 OR COUNT(*) > 1000
		)
		SELECT * FROM ip_clusters
		ORDER BY execution_count DESC
		LIMIT 5
	`

	var rows []struct {
		IPRange        string         `json:"ip_range"`
		TenantCount    int            `json:"tenant_count"`
		FunctionCount  int            `json:"function_count"`
		ExecutionCount int            `json:"execution_count"`
		FirstSeen      time.Time      `json:"first_seen"`
		LastSeen       time.Time      `json:"last_seen"`
		TenantIDs      pq.StringArray `json:"tenant_ids" gorm:"type:text[]"`
	}

	if err := r.db.Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		riskLevel := "medium"
		if row.TenantCount > 10 || row.ExecutionCount > 5000 {
			riskLevel = "high"
		}

		patterns := []string{"High execution volume"}
		if row.TenantCount > 5 {
			patterns = append(patterns, "Many tenants from same IP range")
		}

		associated := make([]string, len(row.TenantIDs))
		copy(associated, row.TenantIDs)
		clusters = append(clusters, IPClusterRow{
			IPRange:           row.IPRange,
			AssociatedTenants: associated,
			RiskLevel:         riskLevel,
			CommonPatterns:    patterns,
			FirstSeen:         row.FirstSeen,
			LastSeen:          row.LastSeen,
		})
	}

	return clusters, nil
}

// detectWashUsagePatterns looks for reciprocal execution patterns between tenants.
func (r *RegistryRepository) detectWashUsagePatterns() ([]WashUsageRow, error) {
	var patterns []WashUsageRow

	// Simplified wash trading detection: look for tenants that execute each other's functions frequently
	query := `
		WITH tenant_pairs AS (
			SELECT
				e1.tenant_id as tenant_a,
				e2.tenant_id as tenant_b,
				e1.function_id,
				COUNT(*) as reciprocal_count
			FROM registry_function_executions e1
			JOIN registry_function_executions e2 ON e1.function_id = e2.function_id
				AND e1.tenant_id = e2.user_id
				AND e2.tenant_id = e1.user_id
			WHERE e1.timestamp > NOW() - INTERVAL '30 days'
				AND e2.timestamp > NOW() - INTERVAL '30 days'
				AND e1.tenant_id != e2.tenant_id
			GROUP BY e1.tenant_id, e2.tenant_id, e1.function_id
			HAVING COUNT(*) > 10
		)
		SELECT
			tp.tenant_a,
			tp.tenant_b,
			f.name as function_name,
			tp.reciprocal_count
		FROM tenant_pairs tp
		JOIN registry_functions f ON tp.function_id = f.id
		ORDER BY tp.reciprocal_count DESC
		LIMIT 5
	`

	var rows []struct {
		TenantA         uuid.UUID `json:"tenant_a"`
		TenantB         uuid.UUID `json:"tenant_b"`
		FunctionName    string    `json:"function_name"`
		ReciprocalCount int       `json:"reciprocal_count"`
	}

	if err := r.db.Raw(query).Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		confidence := 60
		if row.ReciprocalCount > 50 {
			confidence = 85
		}

		patterns = append(patterns, WashUsageRow{
			TenantA:              row.TenantA.String(),
			TenantB:              row.TenantB.String(),
			Function:             row.FunctionName,
			Pattern:              "Reciprocal execution pattern detected",
			Confidence:           confidence,
			ReciprocalExecutions: row.ReciprocalCount,
			DetectedAt:           time.Now(),
		})
	}

	return patterns, nil
}

// detectFakeDiversity identifies tenants that appear to be artificially boosting diversity metrics.
// It flags groups of tenants that execute the same small set of functions from the same IPs,
// with near-identical execution timing — a sign of coordinated fake activity rather than genuine organic usage.
func (r *RegistryRepository) detectFakeDiversity() ([]FakeDiversityRow, error) {
	var rows []FakeDiversityRow

	// Look for tenant groups that share IP patterns and execute the same functions
	// in synchronized windows, indicating scripted fake diversity.
	query := `
		WITH tenant_ip_functions AS (
			SELECT
				e.tenant_id,
				SUBSTRING(e.caller_ip FROM '^(\d+\.\d+)\.') as ip_prefix,
				e.function_id,
				COUNT(*) as exec_count,
				MIN(e.timestamp) as first_seen,
				MAX(e.timestamp) as last_seen,
				array_agg(DISTINCT date_trunc('hour', e.timestamp)::text) as hourly_slots
			FROM registry_function_executions e
			WHERE e.caller_ip IS NOT NULL
				AND e.tenant_id IS NOT NULL
				AND e.timestamp > NOW() - INTERVAL '14 days'
			GROUP BY e.tenant_id, SUBSTRING(e.caller_ip FROM '^(\d+\.\d+)\.'), e.function_id
			HAVING COUNT(*) > 20
		),
		grouped AS (
			SELECT
				t1.ip_prefix,
				t1.function_id,
				array_agg(DISTINCT t1.tenant_id::text) as tenant_list,
				COUNT(DISTINCT t1.tenant_id) as tenant_count,
				SUM(t1.exec_count) as total_execs,
				MIN(t1.first_seen) as first_seen,
				MAX(t1.last_seen) as last_seen,
				-- Measure synchrony: how many tenants share the same hour slots
				(SELECT COUNT(DISTINCT h)
				 FROM unnest(t1.hourly_slots || t2.hourly_slots) AS h) as shared_hour_slots
			FROM tenant_ip_functions t1
			JOIN tenant_ip_functions t2 ON t1.function_id = t2.function_id
				AND t1.tenant_id != t2.tenant_id
				AND t1.ip_prefix = t2.ip_prefix
			WHERE t1.tenant_id < t2.tenant_id  -- avoid duplicate pairs
			GROUP BY t1.ip_prefix, t1.function_id
			HAVING COUNT(DISTINCT t1.tenant_id) >= 3
		)
		SELECT
			ip_prefix,
			tenant_list,
			tenant_count,
			total_execs,
			first_seen,
			last_seen,
			shared_hour_slots
		FROM grouped
		WHERE shared_hour_slots > 0
		ORDER BY total_execs DESC
		LIMIT 10
	`

	// Scan raw
	type rawFakeDiversity struct {
		IPPrefix        string
		TenantList      string
		TenantCount     int
		TotalExecs      int
		FirstSeen       time.Time
		LastSeen        time.Time
		SharedHourSlots int
	}
	var raw []rawFakeDiversity
	if err := r.db.Raw(query).Scan(&raw).Error; err != nil {
		return nil, fmt.Errorf("failed to detect fake diversity: %w", err)
	}

	for _, r := range raw {
		riskLevel := "medium"
		if r.TenantCount >= 5 || r.SharedHourSlots > 10 {
			riskLevel = "high"
		}

		indicators := []string{}
		if r.SharedHourSlots > 0 {
			indicators = append(indicators, "Synchronized execution timing")
		}
		if r.TenantCount >= 3 {
			indicators = append(indicators, "Multiple tenants from same IP prefix")
		}
		indicators = append(indicators, "Low function diversity with high execution volume")

		// Parse tenant list from PostgreSQL array format
		tenantGroup := parsePostgresStringArray(r.TenantList)

		rows = append(rows, FakeDiversityRow{
			TenantGroup:   tenantGroup,
			Indicators:   indicators,
			RiskLevel:    riskLevel,
			FirstSeen:    r.FirstSeen,
			LastSeen:     r.LastSeen,
			FunctionCount: r.TenantCount,
			TotalExecs:   r.TotalExecs,
		})
	}

	return rows, nil
}

// parsePostgresStringArray parses a PostgreSQL array string like {a,b,c} into a slice.
func parsePostgresStringArray(s string) []string {
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		result = append(result, strings.Trim(p, "\""))
	}
	return result
}

// countHighRiskClusters counts clusters with high risk level.
func countHighRiskClusters(clusters []IPClusterRow) int {
	count := 0
	for _, cluster := range clusters {
		if cluster.RiskLevel == "high" {
			count++
		}
	}
	return count
}

// calculateSuspiciousTenants calculates total suspicious tenants from patterns.
func calculateSuspiciousTenants(botPatterns []BotPatternRow, ipClusters []IPClusterRow) int {
	tenantSet := make(map[string]bool)

	for _, pattern := range botPatterns {
		for _, tenant := range pattern.AffectedTenants {
			tenantSet[tenant] = true
		}
	}

	for _, cluster := range ipClusters {
		if cluster.RiskLevel != "high" {
			continue
		}
		for _, tenant := range cluster.AssociatedTenants {
			if tenant != "" {
				tenantSet[tenant] = true
			}
		}
	}

	return len(tenantSet)
}
