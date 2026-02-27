package devtools

import (
	"context"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

// getTables retrieves table metadata
func (dt *DatabaseDevTools) getTables(ctx context.Context) ([]TableInfo, error) {
	// First try the configured schema
	query := `
		SELECT
			t.table_name,
			t.table_schema,
			t.table_type,
			pg_size_pretty(pg_total_relation_size(t.table_schema || '.' || t.table_name)) as size,
			pg_total_relation_size(t.table_schema || '.' || t.table_name) as size_bytes,
			obj_description((t.table_schema || '.' || t.table_name)::regclass, 'pg_class') as description
		FROM information_schema.tables t
		WHERE t.table_schema = $1
		AND t.table_type = 'BASE TABLE'
		AND t.table_name NOT LIKE 'pg_%'  -- Exclude PostgreSQL system tables
		AND t.table_name NOT LIKE 'sql_%' -- Exclude SQL standard tables
		ORDER BY t.table_name;
	`

	rows, err := dt.db.DB.QueryContext(ctx, query, dt.config.SchemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		var sizeStr string
		err := rows.Scan(&table.Name, &table.Schema, &table.Type, &sizeStr, &table.SizeBytes, &table.Description)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}

		// Get columns for this table
		columns, err := dt.getTableColumns(ctx, dt.config.SchemaName, table.Name)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf("failed to get columns for table %s: %w", table.Name, err)
		}
		table.Columns = columns

		// Get row count (use proper table name quoting and context)
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", pq.QuoteIdentifier(table.Name))
		err = dt.db.DB.QueryRowContext(ctx, countQuery).Scan(&table.RowCount)
		if err != nil {
			dt.logger.WithError(err).WithField("table", table.Name).Warn("Failed to get row count, setting to 0")
			table.RowCount = 0 // Set to 0 instead of leaving uninitialized
		}

		tables = append(tables, table)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during table rows iteration: %w", err)
	}

	rows.Close()

	// If no tables found in the primary schema, try searching all non-system schemas
	if len(tables) == 0 {
		dt.logger.WithField("primary_schema", dt.config.SchemaName).Debug("No tables found in primary schema, searching all schemas")
		allSchemasQuery := `
			SELECT
				t.table_name,
				t.table_schema,
				t.table_type,
				pg_size_pretty(pg_total_relation_size(t.table_schema || '.' || t.table_name)) as size,
				pg_total_relation_size(t.table_schema || '.' || t.table_name) as size_bytes,
				obj_description((t.table_schema || '.' || t.table_name)::regclass, 'pg_class') as description
			FROM information_schema.tables t
			WHERE t.table_schema NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
			AND t.table_type = 'BASE TABLE'
			AND t.table_name NOT LIKE 'pg_%'
			AND t.table_name NOT LIKE 'sql_%'
			ORDER BY t.table_schema, t.table_name;
		`

		allRows, err := dt.db.DB.QueryContext(ctx, allSchemasQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to query all schemas: %w", err)
		}
		defer allRows.Close()

		for allRows.Next() {
			var table TableInfo
			var sizeStr string
			err := allRows.Scan(&table.Name, &table.Schema, &table.Type, &sizeStr, &table.SizeBytes, &table.Description)
			if err != nil {
				return nil, fmt.Errorf("failed to scan all-schemas table row: %w", err)
			}

			// Get columns for this table
			columns, err := dt.getTableColumns(ctx, table.Schema, table.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to get columns for table %s.%s: %w", table.Schema, table.Name, err)
			}
			table.Columns = columns

			// Get row count - need to use schema-qualified name for cross-schema queries
			countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", pq.QuoteIdentifier(table.Schema), pq.QuoteIdentifier(table.Name))
			err = dt.db.DB.QueryRowContext(ctx, countQuery).Scan(&table.RowCount)
			if err != nil {
				dt.logger.WithError(err).WithField("table", table.Name).WithField("schema", table.Schema).Warn("Failed to get row count, setting to 0")
				table.RowCount = 0
			}

			tables = append(tables, table)
		}

		// Check for any errors during iteration
		if err := allRows.Err(); err != nil {
			return nil, fmt.Errorf("error during all-schemas table rows iteration: %w", err)
		}
	}

	dt.logger.WithField("table_count", len(tables)).Debug("Retrieved table metadata")
	return tables, nil
}

// getTableColumns retrieves column metadata for a table
func (dt *DatabaseDevTools) getTableColumns(ctx context.Context, schemaName, tableName string) ([]ColumnInfo, error) {
	if strings.TrimSpace(tableName) == "" {
		return nil, fmt.Errorf("table name cannot be empty")
	}
	if strings.TrimSpace(schemaName) == "" {
		schemaName = dt.config.SchemaName
	}

	query := `
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable = 'YES' as nullable,
			c.column_default,
			pgd.description
		FROM information_schema.columns c
		LEFT JOIN pg_catalog.pg_statio_all_tables st ON c.table_schema = st.schemaname AND c.table_name = st.relname
		LEFT JOIN pg_catalog.pg_description pgd ON pgd.objoid = st.relid AND pgd.objsubid = c.ordinal_position
		WHERE c.table_schema = $1
		AND c.table_name = $2
		ORDER BY c.ordinal_position;
	`

	rows, err := dt.db.DB.QueryContext(ctx, query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns for table %s.%s: %w", schemaName, tableName, err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			dt.logger.WithError(closeErr).WithField("table", tableName).Warn("Failed to close column rows")
		}
	}()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var defaultValue *string
		err := rows.Scan(&col.Name, &col.Type, &col.Nullable, &defaultValue, &col.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column row for table %s.%s: %w", schemaName, tableName, err)
		}
		if defaultValue != nil {
			col.DefaultValue = *defaultValue
		}
		columns = append(columns, col)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during column rows iteration for table %s.%s: %w", schemaName, tableName, err)
	}

	return columns, nil
}

// getIndexes retrieves index metadata
func (dt *DatabaseDevTools) getIndexes(ctx context.Context) ([]IndexInfo, error) {
	query := `
		SELECT
			i.indexname,
			t.tablename,
			pg_get_indexdef(i.indexrelid) as definition,
			pg_size_pretty(pg_relation_size(i.indexrelid)) as size,
			pg_relation_size(i.indexrelid) as size_bytes
		FROM pg_indexes i
		JOIN pg_class c ON i.indexname = c.relname
		JOIN pg_stat_user_indexes sui ON sui.indexrelname = i.indexname
		JOIN information_schema.tables t ON t.table_name = i.tablename AND t.table_schema = i.schemaname
		WHERE i.schemaname = $1
		ORDER BY i.indexname;
	`

	rows, err := dt.db.DB.QueryContext(ctx, query, dt.config.SchemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			dt.logger.WithError(closeErr).Warn("Failed to close index rows")
		}
	}()

	var indexes []IndexInfo
	for rows.Next() {
		var idx IndexInfo
		var definition, sizeStr string
		err := rows.Scan(&idx.Name, &idx.Table, &definition, &sizeStr, &idx.SizeBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to scan index row: %w", err)
		}

		// Parse index definition to extract columns and properties
		idx.Columns = dt.parseIndexColumns(definition)
		idx.Type = dt.parseIndexType(definition)
		idx.IsUnique = strings.Contains(definition, "UNIQUE")
		idx.IsPrimary = strings.Contains(definition, "PRIMARY KEY")

		indexes = append(indexes, idx)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during index rows iteration: %w", err)
	}

	dt.logger.WithField("index_count", len(indexes)).Debug("Retrieved index metadata")
	return indexes, nil
}

// getConstraints retrieves constraint metadata
func (dt *DatabaseDevTools) getConstraints(ctx context.Context) ([]ConstraintInfo, error) {
	query := `
		SELECT
			con.conname,
			t.table_name,
			con.contype,
			pg_get_constraintdef(con.oid)
		FROM pg_constraint con
		JOIN information_schema.tables t ON t.table_name = con.conrelid::regclass::text
		WHERE t.table_schema = $1
		ORDER BY con.conname;
	`

	rows, err := dt.db.DB.QueryContext(ctx, query, dt.config.SchemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to query constraints: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			dt.logger.WithError(closeErr).Warn("Failed to close constraint rows")
		}
	}()

	var constraints []ConstraintInfo
	for rows.Next() {
		var constraint ConstraintInfo
		err := rows.Scan(&constraint.Name, &constraint.Table, &constraint.Type, &constraint.Definition)
		if err != nil {
			return nil, fmt.Errorf("failed to scan constraint row: %w", err)
		}
		constraints = append(constraints, constraint)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during constraint rows iteration: %w", err)
	}

	dt.logger.WithField("constraint_count", len(constraints)).Debug("Retrieved constraint metadata")
	return constraints, nil
}

// getFunctions retrieves database function metadata
func (dt *DatabaseDevTools) getFunctions(ctx context.Context) ([]FunctionInfo, error) {
	query := `
		SELECT
			p.proname,
			n.nspname,
			l.lanname,
			pg_get_functiondef(p.oid)
		FROM pg_proc p
		JOIN pg_namespace n ON p.pronamespace = n.oid
		JOIN pg_language l ON p.prolang = l.oid
		WHERE n.nspname = $1
		ORDER BY p.proname;
	`

	rows, err := dt.db.DB.QueryContext(ctx, query, dt.config.SchemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to query functions: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			dt.logger.WithError(closeErr).Warn("Failed to close function rows")
		}
	}()

	var functions []FunctionInfo
	for rows.Next() {
		var fn FunctionInfo
		err := rows.Scan(&fn.Name, &fn.Schema, &fn.Language, &fn.Definition)
		if err != nil {
			return nil, fmt.Errorf("failed to scan function row: %w", err)
		}
		functions = append(functions, fn)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during function rows iteration: %w", err)
	}

	dt.logger.WithField("function_count", len(functions)).Debug("Retrieved function metadata")
	return functions, nil
}

// getRLSPolicies retrieves RLS policy metadata
func (dt *DatabaseDevTools) getRLSPolicies(ctx context.Context) ([]RLSPolicyInfo, error) {
	query := `
		SELECT
			p.policyname,
			p.tablename,
			p.cmd,
			pg_get_policydef(p.oid)
		FROM pg_policies p
		WHERE p.schemaname = $1
		ORDER BY p.tablename, p.policyname;
	`

	rows, err := dt.db.DB.QueryContext(ctx, query, dt.config.SchemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to query RLS policies: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			dt.logger.WithError(closeErr).Warn("Failed to close RLS policy rows")
		}
	}()

	var policies []RLSPolicyInfo
	for rows.Next() {
		var policy RLSPolicyInfo
		err := rows.Scan(&policy.Name, &policy.Table, &policy.Command, &policy.Definition)
		if err != nil {
			return nil, fmt.Errorf("error during RLS policy rows iteration: %w", err)
		}
		policies = append(policies, policy)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during RLS policy rows iteration: %w", err)
	}

	dt.logger.WithField("rls_policy_count", len(policies)).Debug("Retrieved RLS policy metadata")
	return policies, nil
}
