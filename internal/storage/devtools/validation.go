package devtools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// checkDatabaseConnection validates that the database connection is working
func (dt *DatabaseDevTools) checkDatabaseConnection(ctx context.Context) error {
	// Simple connectivity check
	err := dt.db.DB.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

// validateOutputFormat checks if the output format is supported
func (dt *DatabaseDevTools) validateOutputFormat(format string) error {
	validFormats := map[string]bool{
		"json":     true,
		"sql":      true,
		"mermaid":  true,
		"plantuml": true,
	}

	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		return fmt.Errorf("output format cannot be empty")
	}

	if !validFormats[format] {
		return fmt.Errorf("unsupported output format: %s. Supported formats: json, sql, mermaid, plantuml", format)
	}

	return nil
}

// validateOutputFile checks if the output file path is safe to use
func (dt *DatabaseDevTools) validateOutputFile(filePath string) error {
	if filePath == "" {
		return nil // Empty path means stdout, which is valid
	}

	if strings.Contains(filePath, "..") {
		return fmt.Errorf("output file path cannot contain '..' for security reasons")
	}

	// Check if directory exists and is writable
	if dir := filepath.Dir(filePath); dir != "." {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("output directory does not exist: %s", dir)
		}
	}

	return nil
}

// containsDangerousOperations checks if a query contains potentially dangerous operations
func (dt *DatabaseDevTools) containsDangerousOperations(query string) bool {
	dangerousPatterns := []string{
		"drop",
		"delete",
		"truncate",
		"update",
		"insert",
		"create",
		"alter",
		"grant",
		"revoke",
		"execute",
		"call",
		"begin",
		"commit",
		"rollback",
		"savepoint",
		"declare",
		"fetch",
		"close",
		"listen",
		"unlisten",
		"notify",
		"load",
		"copy",
		"do",
	}

	queryLower := strings.ToLower(strings.TrimSpace(query))

	// Check for dangerous keywords at the beginning of statements
	for _, pattern := range dangerousPatterns {
		if strings.HasPrefix(queryLower, pattern+" ") ||
			strings.Contains(queryLower, ";"+pattern+" ") ||
			strings.Contains(queryLower, " "+pattern+" ") {
			dt.logger.WithField("blocked_pattern", pattern).Warn("Query blocked due to dangerous operation")
			return true
		}
	}

	// Check for multiple statements (semicolon injection)
	statementCount := strings.Count(query, ";")
	if statementCount > dt.config.MaxStatements {
		dt.logger.WithField("statement_count", statementCount).WithField("limit", dt.config.MaxStatements).Warn("Query blocked due to multiple statements")
		return true
	}

	// Check for overly complex queries (basic heuristic)
	if len(strings.Split(query, " ")) > dt.config.MaxQueryComplexity {
		dt.logger.WithField("complexity", len(strings.Split(query, " "))).WithField("limit", dt.config.MaxQueryComplexity).Warn("Query blocked due to excessive complexity")
		return true
	}

	return false
}
