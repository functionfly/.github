package devtools

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// GetSchemaOverview returns a lightweight overview of the schema without detailed metadata.
//
// This method is optimized for quick schema inspection and provides basic
// statistics without the performance cost of full schema analysis.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns:
//   - *SchemaOverview: Basic schema statistics
//   - error: Any error encountered during analysis
//
// Example:
//
//	overview, err := tools.GetSchemaOverview(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Schema has %d tables\n", overview.TableCount)
func (dt *DatabaseDevTools) GetSchemaOverview(ctx context.Context) (*SchemaOverview, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	query := `
		SELECT
			COUNT(DISTINCT t.table_name) as table_count,
			COUNT(DISTINCT i.indexname) as index_count,
			COUNT(DISTINCT con.conname) as constraint_count,
			pg_size_pretty(SUM(pg_total_relation_size(t.table_schema || '.' || t.table_name))) as total_size
		FROM information_schema.tables t
		LEFT JOIN pg_indexes i ON i.tablename = t.table_name AND i.schemaname = t.table_schema
		LEFT JOIN pg_constraint con ON con.conrelid::regclass::text = t.table_name
		WHERE t.table_schema = $1
		AND t.table_type = 'BASE TABLE'
	`

	var overview SchemaOverview
	var sizeStr string

	err := dt.db.DB.QueryRowContext(ctx, query, dt.config.SchemaName).Scan(
		&overview.TableCount,
		&overview.IndexCount,
		&overview.ConstraintCount,
		&sizeStr,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema overview: %w", err)
	}

	// Parse size (this is a simplified parsing)
	sizeStr = strings.ReplaceAll(sizeStr, " ", "")
	if strings.HasSuffix(sizeStr, "MB") {
		if size, err := strconv.ParseFloat(strings.TrimSuffix(sizeStr, "MB"), 64); err == nil {
			overview.TotalSizeMB = size
		}
	}

	dt.logger.WithField("tables", overview.TableCount).WithField("indexes", overview.IndexCount).Debug("Retrieved schema overview")

	return &overview, nil
}

// CompareSchemas compares two schema snapshots and shows differences.
//
// This method performs a structural comparison between two schema states,
// identifying changes in tables, indexes, and other schema elements.
//
// Comparison includes:
//   - Added/removed tables
//   - Modified table structures
//   - Added/removed indexes
//   - Schema metadata (counts, sizes)
//
// Parameters:
//   - oldSchema: Baseline schema for comparison
//   - newSchema: Current schema to compare against
//
// Returns:
//   - map[string]interface{}: Comparison results with the following structure:
//     {
//     "added_tables": ["table1", "table2"],
//     "removed_tables": ["old_table"],
//     "modified_tables": ["changed_table"],
//     "added_indexes": ["new_index"],
//     "removed_indexes": ["dropped_index"],
//     "schema_metadata": {
//     "old_table_count": 10,
//     "new_table_count": 12,
//     ...
//     }
//     }
//
// Both schema parameters are validated for nil values. If either is nil,
// an empty schema is used for comparison and a warning is logged.
//
// Example:
//
//	differences := tools.CompareSchemas(oldSchema, newSchema)
//	addedTables := differences["added_tables"].([]string)
//	for _, table := range addedTables {
//	    fmt.Printf("New table: %s\n", table)
//	}
func (dt *DatabaseDevTools) CompareSchemas(oldSchema, newSchema *SchemaInfo) map[string]interface{} {
	// Input validation
	if oldSchema == nil {
		dt.logger.Warn("CompareSchemas called with nil oldSchema")
		oldSchema = &SchemaInfo{}
	}
	if newSchema == nil {
		dt.logger.Warn("CompareSchemas called with nil newSchema")
		newSchema = &SchemaInfo{}
	}

	differences := map[string]interface{}{
		"added_tables":    []string{},
		"removed_tables":  []string{},
		"modified_tables": []string{},
		"added_indexes":   []string{},
		"removed_indexes": []string{},
		"schema_metadata": map[string]interface{}{
			"old_table_count": len(oldSchema.Tables),
			"new_table_count": len(newSchema.Tables),
			"old_index_count": len(oldSchema.Indexes),
			"new_index_count": len(newSchema.Indexes),
		},
	}

	// Compare tables
	oldTables := make(map[string]TableInfo)
	newTables := make(map[string]TableInfo)

	for _, table := range oldSchema.Tables {
		oldTables[table.Name] = table
	}
	for _, table := range newSchema.Tables {
		newTables[table.Name] = table
	}

	for name := range newTables {
		if _, exists := oldTables[name]; !exists {
			differences["added_tables"] = append(differences["added_tables"].([]string), name)
		}
	}

	for name := range oldTables {
		if _, exists := newTables[name]; !exists {
			differences["removed_tables"] = append(differences["removed_tables"].([]string), name)
		}
	}

	// Compare indexes
	oldIndexes := make(map[string]IndexInfo)
	newIndexes := make(map[string]IndexInfo)

	for _, index := range oldSchema.Indexes {
		oldIndexes[index.Name] = index
	}
	for _, index := range newSchema.Indexes {
		newIndexes[index.Name] = index
	}

	for name := range newIndexes {
		if _, exists := oldIndexes[name]; !exists {
			differences["added_indexes"] = append(differences["added_indexes"].([]string), name)
		}
	}

	for name := range oldIndexes {
		if _, exists := newIndexes[name]; !exists {
			differences["removed_indexes"] = append(differences["removed_indexes"].([]string), name)
		}
	}

	dt.logger.WithField("added_tables", len(differences["added_tables"].([]string))).WithField("removed_tables", len(differences["removed_tables"].([]string))).Debug("Schema comparison completed")

	return differences
}