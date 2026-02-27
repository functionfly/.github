// Package devtools provides database development tools for PostgreSQL schema analysis,
// visualization, and query optimization.
//
// This package contains configuration, types, and core structures for DatabaseDevTools.
// It offers comprehensive database introspection capabilities including:
//   - Schema analysis and visualization (JSON, SQL, Mermaid, PlantUML)
//   - Index and constraint metadata extraction
//   - Query execution plan analysis with security safeguards
//   - Schema comparison and diffing
//   - RLS (Row Level Security) policy inspection
//
// All operations are designed with production safety in mind:
//   - Comprehensive error handling with proper context
//   - SQL injection prevention
//   - Input validation and sanitization
//   - Resource cleanup and timeout controls
//   - Structured logging for observability
package devtools

import (
	"encoding/json"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// DatabaseDevToolsConfig contains configuration options for DatabaseDevTools
type DatabaseDevToolsConfig struct {
	// SchemaName is the PostgreSQL schema to analyze (default: "public")
	SchemaName string
	// QueryAnalysisTimeout is the maximum time allowed for query analysis (default: 30s)
	QueryAnalysisTimeout time.Duration
	// MaxQueryComplexity is the maximum number of words allowed in analyzed queries (default: 100)
	MaxQueryComplexity int
	// MaxStatements is the maximum number of SQL statements allowed in analysis (default: 1)
	MaxStatements int
	// FilePermissions are the permissions used when writing output files (default: 0644)
	FilePermissions os.FileMode
}

// DefaultDatabaseDevToolsConfig returns a configuration with sensible defaults
func DefaultDatabaseDevToolsConfig() *DatabaseDevToolsConfig {
	return &DatabaseDevToolsConfig{
		SchemaName:           "public",
		QueryAnalysisTimeout: 30 * time.Second,
		MaxQueryComplexity:   100,
		MaxStatements:        1,
		FilePermissions:      0644,
	}
}

// DatabaseDevTools provides development tools for database schema visualization and query analysis
type DatabaseDevTools struct {
	db      *storage.PostgresDB
	logger  *logrus.Logger
	config  *DatabaseDevToolsConfig
	metrics *DevToolsMetrics
}

// DevToolsMetrics tracks operational metrics for database development tools
type DevToolsMetrics struct {
	SchemaAnalysisCount    int           `json:"schema_analysis_count"`
	QueryAnalysisCount     int           `json:"query_analysis_count"`
	TotalAnalysisTime      time.Duration `json:"total_analysis_time"`
	LastSchemaAnalysisTime time.Time     `json:"last_schema_analysis_time"`
	ErrorCount             int           `json:"error_count"`
}

// GetMetrics returns current operational metrics
func (dt *DatabaseDevTools) GetMetrics() *DevToolsMetrics {
	return dt.metrics
}

// ResetMetrics resets all operational metrics
func (dt *DatabaseDevTools) ResetMetrics() {
	dt.metrics = &DevToolsMetrics{}
}

// SchemaInfo represents database schema information
type SchemaInfo struct {
	Tables      []TableInfo      `json:"tables"`
	Indexes     []IndexInfo      `json:"indexes"`
	Constraints []ConstraintInfo `json:"constraints"`
	Functions   []FunctionInfo   `json:"functions"`
	RLSPolicies []RLSPolicyInfo  `json:"rls_policies"`
}

// TableInfo represents table metadata
type TableInfo struct {
	Name        string       `json:"name"`
	Schema      string       `json:"schema"`
	Type        string       `json:"type"`
	Columns     []ColumnInfo `json:"columns"`
	RowCount    int64        `json:"row_count"`
	SizeBytes   int64        `json:"size_bytes"`
	Description string       `json:"description"`
}

// ColumnInfo represents column metadata
type ColumnInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"default_value"`
	Description  string `json:"description"`
}

// IndexInfo represents index metadata
type IndexInfo struct {
	Name      string   `json:"name"`
	Table     string   `json:"table"`
	Columns   []string `json:"columns"`
	Type      string   `json:"type"`
	IsUnique  bool     `json:"is_unique"`
	IsPrimary bool     `json:"is_primary"`
	SizeBytes int64    `json:"size_bytes"`
}

// ConstraintInfo represents constraint metadata
type ConstraintInfo struct {
	Name       string `json:"name"`
	Table      string `json:"table"`
	Type       string `json:"type"`
	Definition string `json:"definition"`
}

// FunctionInfo represents database function metadata
type FunctionInfo struct {
	Name       string `json:"name"`
	Schema     string `json:"schema"`
	Language   string `json:"language"`
	Definition string `json:"definition"`
}

// RLSPolicyInfo represents RLS policy metadata
type RLSPolicyInfo struct {
	Name       string `json:"name"`
	Table      string `json:"table"`
	Command    string `json:"command"`
	Definition string `json:"definition"`
}

// SchemaOverview provides a lightweight summary of schema statistics
type SchemaOverview struct {
	TableCount      int     `json:"table_count"`
	IndexCount      int     `json:"index_count"`
	ConstraintCount int     `json:"constraint_count"`
	TotalSizeMB     float64 `json:"total_size_mb"`
}

// QueryPlanAnalysis represents EXPLAIN ANALYZE results
type QueryPlanAnalysis struct {
	Query         string          `json:"query"`
	ExecutionTime time.Duration   `json:"execution_time"`
	PlanningTime  time.Duration   `json:"planning_time"`
	Plan          json.RawMessage `json:"plan"`
	Suggestions   []string        `json:"suggestions"`
}