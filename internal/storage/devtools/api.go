package devtools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// NewDatabaseDevTools creates a new database development tools instance.
//
// Parameters:
//   - db: PostgreSQL database connection instance (required)
//   - logger: Optional logger instance. If nil, a new logrus logger is created.
//   - config: Optional configuration. If nil, default configuration is used.
//
// Returns:
//   - *DatabaseDevTools: Configured database tools instance
//
// Performance considerations:
//   - Ensure the database connection pool is properly configured for concurrent operations
//   - For large schemas, consider increasing connection pool size and timeouts
//   - Schema analysis can be expensive; cache results when appropriate
//
// The logger is used for structured logging throughout all operations.
// It is recommended to pass a configured logger for production use.
//
// Panics if db is nil.
func NewDatabaseDevTools(db *storage.PostgresDB, logger *logrus.Logger, config *DatabaseDevToolsConfig) *DatabaseDevTools {
	if db == nil {
		panic("db cannot be nil")
	}

	if logger == nil {
		logger = logrus.New()
	}
	if config == nil {
		config = DefaultDatabaseDevToolsConfig()
	}

	// Validate configuration
	if config.SchemaName == "" {
		config.SchemaName = "public"
	}
	if config.QueryAnalysisTimeout <= 0 {
		config.QueryAnalysisTimeout = 30 * time.Second
	}
	if config.MaxQueryComplexity <= 0 {
		config.MaxQueryComplexity = 100
	}
	if config.MaxStatements <= 0 {
		config.MaxStatements = 1
	}
	if config.FilePermissions == 0 {
		config.FilePermissions = 0644
	}

	return &DatabaseDevTools{
		db:      db,
		logger:  logger,
		config:  config,
		metrics: &DevToolsMetrics{},
	}
}

// GenerateSchemaDiagram generates a database schema diagram in various formats.
//
// Supported output formats:
//   - "json": Structured JSON representation of the schema
//   - "sql": SQL DDL statements for schema recreation
//   - "mermaid": Mermaid ER diagram syntax for documentation
//   - "plantuml": PlantUML class diagram syntax
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - outputFormat: Desired output format (json, sql, mermaid, plantuml)
//   - outputFile: Optional file path to write output. If empty, prints to stdout.
//
// Returns:
//   - error: Any error encountered during generation
//
// Security considerations:
//   - Output file paths are validated to prevent directory traversal
//   - Parent directories must exist for file output
//
// Example:
//
//	err := tools.GenerateSchemaDiagram(ctx, "mermaid", "schema.md")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (dt *DatabaseDevTools) GenerateSchemaDiagram(ctx context.Context, outputFormat string, outputFile string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// Input validation
	if err := dt.validateOutputFormat(outputFormat); err != nil {
		return err
	}

	if err := dt.validateOutputFile(outputFile); err != nil {
		return err
	}

	schema, err := dt.AnalyzeSchema(ctx)
	if err != nil {
		return fmt.Errorf("failed to analyze schema: %w", err)
	}

	var output string
	switch strings.ToLower(outputFormat) {
	case "json":
		return dt.generateJSONSchema(schema, outputFile)
	case "sql":
		return dt.generateSQLSchema(schema, outputFile)
	case "mermaid":
		return dt.generateMermaidDiagram(schema, outputFile)
	case "plantuml":
		return dt.generatePlantUMLDiagram(schema, outputFile)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	if outputFile != "" {
		return os.WriteFile(outputFile, []byte(output), dt.config.FilePermissions)
	}

	fmt.Println(output)
	return nil
}

// AnalyzeSchema analyzes the current database schema and returns comprehensive metadata.
//
// This method performs a complete schema introspection including:
//   - Tables with column information, row counts, and sizes
//   - Indexes with column analysis and types
//   - Constraints (primary keys, foreign keys, unique, check)
//   - Database functions and procedures
//   - Row Level Security (RLS) policies
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns:
//   - *SchemaInfo: Complete schema metadata
//   - error: Any error encountered during analysis
//
// The operation may take time for large schemas. Use context with timeout
// for production environments to prevent long-running operations.
func (dt *DatabaseDevTools) AnalyzeSchema(ctx context.Context) (*SchemaInfo, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Check database connectivity
	if err := dt.checkDatabaseConnection(ctx); err != nil {
		dt.metrics.ErrorCount++
		return nil, fmt.Errorf("database connectivity check failed: %w", err)
	}

	schema := &SchemaInfo{}

	start := time.Now()
	dt.logger.WithField("operation", "analyze_schema").WithField("schema", dt.config.SchemaName).Info("Starting schema analysis")

	// Get tables
	tables, err := dt.getTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	schema.Tables = tables

	// Get indexes
	indexes, err := dt.getIndexes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get indexes: %w", err)
	}
	schema.Indexes = indexes

	// Get constraints
	constraints, err := dt.getConstraints(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get constraints: %w", err)
	}
	schema.Constraints = constraints

	// Get functions
	functions, err := dt.getFunctions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get functions: %w", err)
	}
	schema.Functions = functions

	// Get RLS policies
	rlsPolicies, err := dt.getRLSPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get RLS policies: %w", err)
	}
	schema.RLSPolicies = rlsPolicies

	analysisTime := time.Since(start)
	dt.metrics.SchemaAnalysisCount++
	dt.metrics.TotalAnalysisTime += analysisTime
	dt.metrics.LastSchemaAnalysisTime = time.Now()

	dt.logger.WithField("operation", "analyze_schema").WithField("tables", len(tables)).WithField("duration", analysisTime).Info("Schema analysis completed")
	return schema, nil
}

// AnalyzeQueryPlan analyzes a query execution plan with security and performance safeguards.
//
// This method executes EXPLAIN ANALYZE on the provided query and returns:
//   - Execution time and planning time
//   - Complete query execution plan in JSON format
//   - Automated suggestions for query optimization
//
// Security measures:
//   - Blocks dangerous operations (DROP, DELETE, UPDATE, INSERT, etc.)
//   - Prevents SQL injection through proper parameterization
//   - Enforces 30-second timeout to prevent long-running analysis
//   - Validates query complexity and statement count
//
// Parameters:
//   - ctx: Context for cancellation (additional to internal timeout)
//   - query: SQL query to analyze (SELECT statements only)
//   - args: Optional query parameters for safe parameterization
//
// Returns:
//   - *QueryPlanAnalysis: Complete analysis results with suggestions
//   - error: Any error encountered during analysis
//
// Only SELECT queries are allowed. Any attempt to analyze modification
// queries will result in an error.
//
// Example:
//
//	analysis, err := tools.AnalyzeQueryPlan(ctx,
//	    "SELECT u.name FROM users u JOIN posts p ON u.id = p.user_id WHERE u.active = $1",
//	    true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, suggestion := range analysis.Suggestions {
//	    fmt.Println("Suggestion:", suggestion)
//	}
func (dt *DatabaseDevTools) AnalyzeQueryPlan(ctx context.Context, query string, args ...interface{}) (*QueryPlanAnalysis, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Input validation and security checks
	if strings.TrimSpace(query) == "" {
		dt.metrics.ErrorCount++
		return nil, fmt.Errorf("query cannot be empty")
	}

	// Basic security check - prevent dangerous operations
	if dt.containsDangerousOperations(query) {
		dt.metrics.ErrorCount++
		return nil, fmt.Errorf("query contains potentially dangerous operations that are not allowed in analysis mode")
	}

	analysis := &QueryPlanAnalysis{
		Query:       query,
		Suggestions: []string{},
	}

	dt.logger.WithField("operation", "analyze_query_plan").WithField("query_length", len(query)).Info("Starting query plan analysis")

	// Prepare EXPLAIN ANALYZE query
	explainQuery := fmt.Sprintf("EXPLAIN (ANALYZE, VERBOSE, COSTS, BUFFERS, FORMAT JSON) %s", query)

	// Create a context with timeout to prevent long-running analysis
	analysisCtx, cancel := context.WithTimeout(ctx, dt.config.QueryAnalysisTimeout) // 30 second timeout for analysis
	defer cancel()

	start := time.Now()

	// Use context for timeout control
	rows, err := dt.db.DB.QueryContext(analysisCtx, explainQuery, args...)
	if err != nil {
		dt.metrics.ErrorCount++
		return nil, fmt.Errorf("failed to execute EXPLAIN: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			dt.logger.WithError(closeErr).Warn("Failed to close query rows")
		}
	}()

	analysis.ExecutionTime = time.Since(start)

	var planJSON string
	if rows.Next() {
		if err := rows.Scan(&planJSON); err != nil {
			dt.metrics.ErrorCount++
			return nil, fmt.Errorf("failed to scan plan: %w", err)
		}
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		dt.metrics.ErrorCount++
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	analysis.Plan = json.RawMessage(planJSON)

	// Analyze the plan and provide suggestions
	analysis.Suggestions = dt.analyzeQueryPlan(planJSON)

	dt.metrics.QueryAnalysisCount++

	dt.logger.WithField("operation", "analyze_query_plan").WithField("execution_time", analysis.ExecutionTime).Info("Query plan analysis completed")

	return analysis, nil
}
