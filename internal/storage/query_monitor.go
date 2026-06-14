package storage

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// QueryMonitor provides query performance monitoring and slow query logging
type QueryMonitor struct {
	db             *PostgresDB
	slowQueryThreshold time.Duration
	enabled           bool
	logQueries       bool
	collectStats     bool
	queryStats       map[string]*QueryStats
}

// QueryStats holds statistics for a specific query pattern
type QueryStats struct {
	QueryPattern string
	Count        int64
	TotalTime    time.Duration
	AvgTime      time.Duration
	MaxTime      time.Duration
	MinTime      time.Duration
	LastExecuted time.Time
}

// QueryExecution represents a single query execution
type QueryExecution struct {
	Query     string
	Args      []interface{}
	Duration  time.Duration
	StartTime time.Time
	EndTime   time.Time
	Error     error
}

// NewQueryMonitor creates a new query performance monitor
func NewQueryMonitor(db *PostgresDB) *QueryMonitor {
	return &QueryMonitor{
		db:                 db,
		slowQueryThreshold: 100 * time.Millisecond, // Default 100ms
		enabled:           true,
		logQueries:       false, // Disabled by default for performance
		collectStats:     true,
		queryStats:       make(map[string]*QueryStats),
	}
}

// EnableSlowQueryLogging enables slow query logging
func (qm *QueryMonitor) EnableSlowQueryLogging(threshold time.Duration) {
	qm.enabled = true
	qm.slowQueryThreshold = threshold
	logrus.WithField("threshold", threshold).Info("Slow query logging enabled")
}

// EnableQueryLogging enables detailed query logging
func (qm *QueryMonitor) EnableQueryLogging() {
	qm.logQueries = true
	logrus.Info("Detailed query logging enabled")
}

// Disable disables all query monitoring
func (qm *QueryMonitor) Disable() {
	qm.enabled = false
	qm.logQueries = false
	qm.collectStats = false
	logrus.Info("Query monitoring disabled")
}

// MonitorQuery monitors a single query execution
func (qm *QueryMonitor) MonitorQuery(query string, args []interface{}, fn func() error) error {
	if !qm.enabled {
		return fn()
	}

	start := time.Now()
	err := fn()
	duration := time.Since(start)

	execution := QueryExecution{
		Query:     query,
		Args:      args,
		Duration:  duration,
		StartTime: start,
		EndTime:   time.Now(),
		Error:     err,
	}

	qm.recordExecution(execution)
	return err
}

// MonitorGORMQuery monitors GORM database operations
func (qm *QueryMonitor) MonitorGORMQuery(db *gorm.DB) *gorm.DB {
	if !qm.enabled {
		return db
	}

	// Register query callbacks
	db.Callback().Query().Before("gorm:query").Register("query_monitor:before_query", qm.beforeQuery)
	db.Callback().Query().After("gorm:query").Register("query_monitor:after_query", qm.afterQuery)

	// Register create callbacks
	db.Callback().Create().Before("gorm:create").Register("query_monitor:before_create", qm.beforeQuery)
	db.Callback().Create().After("gorm:create").Register("query_monitor:after_create", qm.afterQuery)

	// Register update callbacks
	db.Callback().Update().Before("gorm:update").Register("query_monitor:before_update", qm.beforeQuery)
	db.Callback().Update().After("gorm:update").Register("query_monitor:after_update", qm.afterQuery)

	// Register delete callbacks
	db.Callback().Delete().Before("gorm:delete").Register("query_monitor:before_delete", qm.beforeQuery)
	db.Callback().Delete().After("gorm:delete").Register("query_monitor:after_delete", qm.afterQuery)

	// Register row query callbacks
	db.Callback().Row().Before("gorm:row_query").Register("query_monitor:before_row", qm.beforeQuery)
	db.Callback().Row().After("gorm:row_query").Register("query_monitor:after_row", qm.afterQuery)

	return db
}

// GORM callback functions
func (qm *QueryMonitor) beforeQuery(db *gorm.DB) {
	if !qm.enabled {
		return
	}

	// Store start time in DB instance
	db.InstanceSet("query_monitor_start", time.Now())
}

func (qm *QueryMonitor) afterQuery(db *gorm.DB) {
	if !qm.enabled {
		return
	}

	startTime, ok := db.Get("query_monitor_start")
	if !ok {
		return
	}

	start, ok := startTime.(time.Time)
	if !ok {
		return
	}

	duration := time.Since(start)

	// Fast path: skip all overhead if below threshold and stats collection is disabled
	if duration < qm.slowQueryThreshold && !qm.logQueries && !qm.collectStats {
		return
	}

	query := db.Statement.SQL.String()

	// Get query args if available
	var args []interface{}
	if db.Statement.Vars != nil {
		args = db.Statement.Vars
	}

	execution := QueryExecution{
		Query:     query,
		Args:      args,
		Duration:  duration,
		StartTime: start,
		EndTime:   time.Now(),
		Error:     db.Error,
	}

	qm.recordExecution(execution)
}

// recordExecution records a query execution for monitoring
func (qm *QueryMonitor) recordExecution(execution QueryExecution) {
	// Log slow queries
	if execution.Duration >= qm.slowQueryThreshold {
		qm.logSlowQuery(execution)
	}

	// Log all queries if enabled
	if qm.logQueries {
		qm.logQuery(execution)
	}

	// Collect statistics
	if qm.collectStats {
		qm.collectQueryStats(execution)
	}
}

// logSlowQuery logs slow query information
func (qm *QueryMonitor) logSlowQuery(execution QueryExecution) {
	logrus.WithFields(logrus.Fields{
		"duration_ms": execution.Duration.Milliseconds(),
		"query":       qm.sanitizeQuery(execution.Query),
		"args_count":  len(execution.Args),
		"start_time":  execution.StartTime,
		"end_time":    execution.EndTime,
	}).Warn("Slow query detected")
}

// logQuery logs detailed query information
func (qm *QueryMonitor) logQuery(execution QueryExecution) {
	level := logrus.InfoLevel
	if execution.Error != nil {
		level = logrus.ErrorLevel
	}

	logrus.WithFields(logrus.Fields{
		"duration_ms": execution.Duration.Milliseconds(),
		"query":       qm.sanitizeQuery(execution.Query),
		"args":        qm.sanitizeArgs(execution.Args),
		"error":       execution.Error,
	}).Log(level, "Query executed")
}

// collectQueryStats collects statistics for query patterns
func (qm *QueryMonitor) collectQueryStats(execution QueryExecution) {
	pattern := qm.extractQueryPattern(execution.Query)

	stats, exists := qm.queryStats[pattern]
	if !exists {
		stats = &QueryStats{
			QueryPattern: pattern,
			MinTime:      execution.Duration,
			MaxTime:      execution.Duration,
		}
		qm.queryStats[pattern] = stats
	}

	stats.Count++
	stats.TotalTime += execution.Duration
	stats.AvgTime = stats.TotalTime / time.Duration(stats.Count)
	stats.LastExecuted = execution.StartTime

	if execution.Duration > stats.MaxTime {
		stats.MaxTime = execution.Duration
	}
	if execution.Duration < stats.MinTime {
		stats.MinTime = execution.Duration
	}
}

// extractQueryPattern extracts a normalized pattern from a query
func (qm *QueryMonitor) extractQueryPattern(query string) string {
	// Remove extra whitespace
	query = regexp.MustCompile(`\s+`).ReplaceAllString(query, " ")
	query = strings.TrimSpace(query)

	// Replace literal values with placeholders
	query = regexp.MustCompile(`\$[0-9]+`).ReplaceAllString(query, "?")
	query = regexp.MustCompile(`'[^\']*'`).ReplaceAllString(query, "?")
	query = regexp.MustCompile(`"[^\"]*"`).ReplaceAllString(query, "?")

	return strings.ToUpper(query)
}

// sanitizeQuery removes sensitive data from query strings
func (qm *QueryMonitor) sanitizeQuery(query string) string {
	// Remove potential sensitive data patterns
	query = regexp.MustCompile(`(password|token|secret|key)\s*=\s*[^,\s)]+`).ReplaceAllString(query, "$1=***")
	return query
}

// sanitizeArgs removes sensitive data from query arguments
func (qm *QueryMonitor) sanitizeArgs(args []interface{}) []interface{} {
	if len(args) == 0 {
		return args
	}

	sanitized := make([]interface{}, len(args))
	for i, arg := range args {
		argStr := fmt.Sprintf("%v", arg)
		// Simple check for potential sensitive data
		if strings.Contains(strings.ToLower(argStr), "password") ||
		   strings.Contains(strings.ToLower(argStr), "token") ||
		   strings.Contains(strings.ToLower(argStr), "secret") {
			sanitized[i] = "***"
		} else {
			sanitized[i] = arg
		}
	}

	return sanitized
}

// GetQueryStats returns current query statistics
func (qm *QueryMonitor) GetQueryStats() map[string]*QueryStats {
	// Return a copy to avoid concurrent map writes
	stats := make(map[string]*QueryStats)
	for k, v := range qm.queryStats {
		stats[k] = &QueryStats{
			QueryPattern: v.QueryPattern,
			Count:        v.Count,
			TotalTime:    v.TotalTime,
			AvgTime:      v.AvgTime,
			MaxTime:      v.MaxTime,
			MinTime:      v.MinTime,
			LastExecuted: v.LastExecuted,
		}
	}
	return stats
}

// GetSlowQueries returns queries that exceed the threshold
func (qm *QueryMonitor) GetSlowQueries() []*QueryStats {
	var slowQueries []*QueryStats

	for _, stats := range qm.queryStats {
		if stats.AvgTime >= qm.slowQueryThreshold {
			slowQueries = append(slowQueries, stats)
		}
	}

	return slowQueries
}

// ResetStats resets all collected statistics
func (qm *QueryMonitor) ResetStats() {
	qm.queryStats = make(map[string]*QueryStats)
	logrus.Info("Query statistics reset")
}

// ConfigurePostgresLogging configures PostgreSQL server-side logging
func (qm *QueryMonitor) ConfigurePostgresLogging() error {
	// Set PostgreSQL parameters for better query logging
	queries := []string{
		"SET log_statement = 'all'",
		"SET log_duration = on",
		"SET log_min_duration_statement = 100", // Log queries taking longer than 100ms
		"SET log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '",
		"SET log_statement_stats = on",
	}

	for _, query := range queries {
		if _, err := qm.db.DB.Exec(query); err != nil {
			logrus.WithError(err).WithField("query", query).Warn("Failed to configure PostgreSQL logging")
			// Don't return error, continue with other settings
		}
	}

	logrus.Info("PostgreSQL server-side logging configured")
	return nil
}

// CreateQueryPerformanceView creates a database view for query performance analysis
func (qm *QueryMonitor) CreateQueryPerformanceView() error {
	viewSQL := `
	CREATE OR REPLACE VIEW query_performance_summary AS
	SELECT
		query,
		calls,
		total_time,
		mean_time,
		stddev_time,
		min_time,
		max_time,
		rows,
		temp_blks_written,
		temp_blks_read,
		blk_write_time,
		blk_read_time
	FROM pg_stat_statements
	WHERE query NOT LIKE '%pg_stat_statements%'
	ORDER BY mean_time DESC;
	`

	if _, err := qm.db.DB.Exec(viewSQL); err != nil {
		return fmt.Errorf("failed to create query performance view: %w", err)
	}

	logrus.Info("Query performance summary view created")
	return nil
}