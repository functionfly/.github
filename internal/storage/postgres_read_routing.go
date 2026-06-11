package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ============================================================================
// Read Replica Query Routing Helpers
// ============================================================================

// ReadPreference indicates the query's tolerance for replication lag
type ReadPreference int

const (
	// ReadStrong requires primary database (zero lag tolerance)
	// Use for: user writes, billing updates, critical reads immediately after writes
	ReadStrong ReadPreference = iota

	// ReadDefault allows healthy replicas (accepts normal replication lag ~100ms)
	// Use for: dashboard data, listing queries, most read operations
	ReadDefault

	// ReadStale allows any replica including lagging ones (accepts high lag ~5s)
	// Use for: analytics, reports, background processing, non-urgent aggregations
	ReadStale
)

// String representation for logging
func (rp ReadPreference) String() string {
	switch rp {
	case ReadStrong:
		return "strong"
	case ReadDefault:
		return "default"
	case ReadStale:
		return "stale"
	default:
		return "unknown"
	}
}

// QueryContext executes a read query with automatic replica routing
// This is the main entry point for replica-aware queries
//
// Usage:
//
//	rows, err := db.QueryWithPreference(ctx, ReadDefault, "SELECT * FROM users WHERE tenant_id = $1", tenantID)
func (db *PostgresDB) QueryWithPreference(
	ctx context.Context,
	preference ReadPreference,
	query string,
	args ...interface{},
) (*sql.Rows, error) {
	start := time.Now()
	conn := db.selectConnection(preference)

	rows, err := conn.QueryContext(ctx, query, args...)

	// Track metrics
	db.trackReadQuery(conn, preference, time.Since(start), err)

	return rows, err
}

// QueryRowWithPreference executes a single-row read query with replica routing
func (db *PostgresDB) QueryRowWithPreference(
	ctx context.Context,
	preference ReadPreference,
	query string,
	args ...interface{},
) *sql.Row {
	start := time.Now()
	conn := db.selectConnection(preference)

	row := conn.QueryRowContext(ctx, query, args...)

	// Track metrics (best effort, ignore errors for single row)
	db.trackReadQuery(conn, preference, time.Since(start), nil)

	return row
}

// selectConnection chooses the appropriate database connection based on preference
func (db *PostgresDB) selectConnection(preference ReadPreference) *sql.DB {
	switch preference {
	case ReadStrong:
		return db.DB // Always use primary

	case ReadDefault:
		// Use healthy replicas if available
		return db.getReadReplica()

	case ReadStale:
		// Use any available replica even if marked unhealthy
		// (may have high lag but still useful for analytics)
		return db.getAnyReplica()

	default:
		return db.getReadReplica()
	}
}

// getAnyReplica returns any replica connection, even if marked unhealthy
func (db *PostgresDB) getAnyReplica() *sql.DB {
	if !db.readReplicaEnabled || len(db.readReplicas) == 0 {
		return db.DB
	}

	// Return first replica regardless of health
	return db.readReplicas[0].DB
}

// trackReadQuery logs metrics for read query routing decisions
func (db *PostgresDB) trackReadQuery(conn *sql.DB, preference ReadPreference, duration time.Duration, err error) {
	// Determine if this hit primary or replica
	isPrimary := conn == db.DB

	logFields := logrus.Fields{
		"read_preference": preference.String(),
		"duration_ms":     duration.Milliseconds(),
		"is_primary":      isPrimary,
	}

	if err != nil {
		logFields["error"] = err.Error()
		logrus.WithFields(logFields).Warn("Read query error")
	} else if duration > 100*time.Millisecond {
		// Log slow queries
		logrus.WithFields(logFields).Info("Slow read query")
	}

	// Future: Send metrics to Prometheus/Datadog
	// metrics.ReadQueryDuration.WithLabelValues(preference.String(), fmt.Sprint(isPrimary)).Observe(duration.Seconds())
}

// ============================================================================
// Repository-Level Read Routing Helpers
// ============================================================================

// RepositoryReadHelper provides convenient methods for repositories
// to route queries to appropriate databases
type RepositoryReadHelper struct {
	db *PostgresDB
}

// NewRepositoryReadHelper creates a read routing helper for a repository
func NewRepositoryReadHelper(db *PostgresDB) *RepositoryReadHelper {
	return &RepositoryReadHelper{db: db}
}

// ForDashboard returns a connection suitable for dashboard queries
// (uses replicas for most operations, primary for critical updates)
func (h *RepositoryReadHelper) ForDashboard() *sql.DB {
	return h.db.selectConnection(ReadDefault)
}

// ForAnalytics returns a connection for analytics/reporting queries
// (aggressively uses replicas, tolerates lag)
func (h *RepositoryReadHelper) ForAnalytics() *sql.DB {
	return h.db.selectConnection(ReadStale)
}

// ForCritical returns the primary connection for strong consistency
func (h *RepositoryReadHelper) ForCritical() *sql.DB {
	return h.db.selectConnection(ReadStrong)
}

// Query executes a SELECT with automatic routing based on query characteristics
func (h *RepositoryReadHelper) Query(
	ctx context.Context,
	query string,
	args ...interface{},
) (*sql.Rows, error) {
	preference := h.classifyQuery(query)
	return h.db.QueryWithPreference(ctx, preference, query, args...)
}

// QueryRow executes a single-row SELECT with routing
func (h *RepositoryReadHelper) QueryRow(
	ctx context.Context,
	query string,
	args ...interface{},
) *sql.Row {
	preference := h.classifyQuery(query)
	return h.db.QueryRowWithPreference(ctx, preference, query, args...)
}

// classifyQuery determines read preference based on query characteristics
func (h *RepositoryReadHelper) classifyQuery(query string) ReadPreference {
	// Simple heuristics - can be enhanced with query parsing
	lowerQuery := fmt.Sprint(query)

	// Always use primary for these patterns (need strong consistency)
	strongPatterns := []string{
		"FOR UPDATE",
		"FOR SHARE",
		"NOW()",           // May need current timestamp
		"current_setting", // Configuration queries
	}

	for _, pattern := range strongPatterns {
		if containsIgnoreCase(lowerQuery, pattern) {
			return ReadStrong
		}
	}

	// Use stale reads for heavy analytics patterns
	stalePatterns := []string{
		"COUNT(*)",
		"GROUP BY",
		"SUM(",
		"AVG(",
		"PERCENTILE",
		"WINDOW",
		"OVER (",
		"mv_", // Materialized views are already pre-computed
	}

	for _, pattern := range stalePatterns {
		if containsIgnoreCase(lowerQuery, pattern) {
			return ReadStale
		}
	}

	// Default: standard read preference
	return ReadDefault
}

// ============================================================================
// Billing Repository Read Extensions
// ============================================================================

// BillingAnalyticsRepository provides read-optimized methods for billing queries
type BillingAnalyticsRepository struct {
	db *PostgresDB
}

// NewBillingAnalyticsRepository creates a billing analytics repo with replica routing
func NewBillingAnalyticsRepository(db *PostgresDB) *BillingAnalyticsRepository {
	return &BillingAnalyticsRepository{db: db}
}

// GetDailySummary uses materialized view via replica (stale OK)
func (r *BillingAnalyticsRepository) GetDailySummary(ctx context.Context, tenantID string, days int) (*sql.Rows, error) {
	return r.db.QueryWithPreference(ctx, ReadStale,
		`SELECT * FROM mv_tenant_daily_billing_summary 
		 WHERE tenant_id = $1 AND billing_date >= CURRENT_DATE - INTERVAL '$2 days'
		 ORDER BY billing_date DESC`,
		tenantID, days)
}

// GetFunctionStats uses materialized view via replica
func (r *BillingAnalyticsRepository) GetFunctionStats(ctx context.Context, functionID string) *sql.Row {
	return r.db.QueryRowWithPreference(ctx, ReadStale,
		`SELECT * FROM mv_function_usage_stats WHERE function_id = $1`,
		functionID)
}

// GetCurrentMRR uses subscription summary via replica
func (r *BillingAnalyticsRepository) GetCurrentMRR(ctx context.Context, tenantID string) *sql.Row {
	return r.db.QueryRowWithPreference(ctx, ReadDefault,
		`SELECT SUM(recognized_mrr_cents) 
		 FROM mv_subscription_revenue_summary 
		 WHERE tenant_id = $1 AND revenue_status = 'healthy'`,
		tenantID)
}

// GetCohortRetention uses cohort analysis view via replica
func (r *BillingAnalyticsRepository) GetCohortRetention(ctx context.Context, cohortMonth string) (*sql.Rows, error) {
	return r.db.QueryWithPreference(ctx, ReadStale,
		`SELECT * FROM mv_tenant_cohort_analysis 
		 WHERE cohort_month = $1 
		 ORDER BY months_since_first`,
		cohortMonth)
}

// ============================================================================
// Analytics Repository Read Extensions
// ============================================================================

// GetRegionalPerformanceMV returns regional performance from materialized view
// Uses ReadStale preference for analytics queries
func (r *AnalyticsRepository) GetRegionalPerformanceMV(ctx context.Context, region string, days int) (*sql.Rows, error) {
	return r.db.QueryWithPreference(ctx, ReadStale,
		`SELECT * FROM mv_regional_performance_summary 
		 WHERE region = $1 AND stat_date >= CURRENT_DATE - INTERVAL '$2 days'
		 ORDER BY stat_date DESC`,
		region, days)
}

// GetPlatformMetricsMV returns platform metrics from materialized view
func (r *AnalyticsRepository) GetPlatformMetricsMV(ctx context.Context, days int) (*sql.Rows, error) {
	return r.db.QueryWithPreference(ctx, ReadStale,
		`SELECT * FROM mv_platform_daily_metrics 
		 WHERE metric_date >= CURRENT_DATE - INTERVAL '$1 days'
		 ORDER BY metric_date DESC`,
		days)
}

// GetTeamCostAllocationMV returns team cost data from materialized view
func (r *AnalyticsRepository) GetTeamCostAllocationMV(ctx context.Context, teamID string, days int) (*sql.Rows, error) {
	return r.db.QueryWithPreference(ctx, ReadStale,
		`SELECT * FROM mv_team_cost_allocation 
		 WHERE team_id = $1 AND usage_date >= CURRENT_DATE - INTERVAL '$2 days'
		 ORDER BY usage_date DESC`,
		teamID, days)
}

// ============================================================================
// User Repository Read Extensions
// ============================================================================

// UserReadRepository provides read-optimized user queries
type UserReadRepository struct {
	db *PostgresDB
}

// NewUserReadRepository creates a user read repo with smart routing
func NewUserReadRepository(db *PostgresDB) *UserReadRepository {
	return &UserReadRepository{db: db}
}

// SearchUsers uses replica for search (eventually consistent OK)
func (r *UserReadRepository) SearchUsers(ctx context.Context, query string, limit int) (*sql.Rows, error) {
	return r.db.QueryWithPreference(ctx, ReadDefault,
		`SELECT id, username, email, name
		 FROM users
		 WHERE username ILIKE $1 ESCAPE '\' OR email ILIKE $1 ESCAPE '\' OR name ILIKE $1 ESCAPE '\'
		 LIMIT $2`,
		"%"+escapeLikeWildcards(query)+"%", limit)
}

func escapeLikeWildcards(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// GetUserTeams uses replica for listing (acceptable lag)
func (r *UserReadRepository) GetUserTeams(ctx context.Context, userID string) (*sql.Rows, error) {
	return r.db.QueryWithPreference(ctx, ReadDefault,
		`SELECT t.*, tm.role 
		 FROM teams t 
		 JOIN team_memberships tm ON t.id = tm.team_id 
		 WHERE tm.user_id = $1`,
		userID)
}

// GetUserActivity uses replica for activity feed (acceptable lag)
func (r *UserReadRepository) GetUserActivity(ctx context.Context, userID string, limit int) (*sql.Rows, error) {
	return r.db.QueryWithPreference(ctx, ReadStale,
		`SELECT * FROM user_activity 
		 WHERE user_id = $1 
		 ORDER BY created_at DESC 
		 LIMIT $2`,
		userID, limit)
}

// GetUserByID uses primary for authentication-critical reads
func (r *UserReadRepository) GetUserByID(ctx context.Context, userID string) *sql.Row {
	return r.db.QueryRowWithPreference(ctx, ReadStrong,
		`SELECT * FROM users WHERE id = $1`,
		userID)
}

// GetUserByEmail uses primary for authentication-critical reads
func (r *UserReadRepository) GetUserByEmail(ctx context.Context, email string) *sql.Row {
	return r.db.QueryRowWithPreference(ctx, ReadStrong,
		`SELECT * FROM users WHERE email = $1`,
		email)
}

// ============================================================================
// Helper Functions
// ============================================================================

func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive contains
	return len(s) >= len(substr) &&
		(findSubstrIgnoreCase(s, substr) >= 0)
}

func findSubstrIgnoreCase(s, substr string) int {
	// Naive implementation - can be optimized
	ls := len(s)
	lsub := len(substr)
	if lsub > ls {
		return -1
	}

	for i := 0; i <= ls-lsub; i++ {
		match := true
		for j := 0; j < lsub; j++ {
			c1 := s[i+j]
			c2 := substr[j]

			// Case-insensitive comparison
			if c1 >= 'A' && c1 <= 'Z' {
				c1 = c1 + ('a' - 'A')
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 = c2 + ('a' - 'A')
			}

			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
