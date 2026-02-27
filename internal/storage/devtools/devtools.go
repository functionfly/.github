// Package devtools provides database development tools for PostgreSQL schema analysis,
// visualization, and query optimization.
//
// This package offers comprehensive database introspection capabilities including:
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
//
// Example usage:
//
//	db := NewPostgresDB(config)
//	tools := NewDatabaseDevTools(db, logger)
//
//	// Analyze schema
//	schema, err := tools.AnalyzeSchema(ctx)
//	if err != nil {
//	    return err
//	}
//
//	// Generate schema diagram
//	err = tools.GenerateSchemaDiagram(ctx, "mermaid", "schema.md")
//	if err != nil {
//	    return err
//	}
//
//	// Analyze query performance
//	analysis, err := tools.AnalyzeQueryPlan(ctx, "SELECT * FROM users WHERE id = $1", 123)
//	if err != nil {
//	    return err
//	}
package devtools

// This file serves as the main entry point for the devtools package.
// All public APIs are exported from their respective files.
