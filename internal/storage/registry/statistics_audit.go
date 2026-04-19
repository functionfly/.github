package registry

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ExecutionAuditRow represents a row for execution audit with function details.
type ExecutionAuditRow struct {
	ID                string
	FunctionID        uuid.UUID
	FunctionName      string
	Author            string
	Version           string
	Timestamp         time.Time
	DurationMs        int
	Outcome           string
	ErrorCode         sql.NullString
	TenantID          *uuid.UUID
	InputSize         sql.NullInt64 // From wasm_execution_audit if available
	OutputSize        sql.NullInt64 // From wasm_execution_audit if available
	ExecutionRootHash sql.NullString // From certificates if available
	NodeSignature     sql.NullString // From certificates if available
}

// GetExecutionAuditData returns paginated execution audit data with function details and filtering.
// Supports search (function name/author), tenant filter, status filter, pagination.
func (r *RegistryRepository) GetExecutionAuditData(
	searchTerm, tenantFilter, statusFilter string,
	offset, limit int,
) ([]ExecutionAuditRow, int64, error) {
	// Build the base query with joins
	query := r.db.Table("registry_function_executions e").
		Joins("JOIN registry_functions f ON e.function_id = f.id").
		Joins("LEFT JOIN execution_certificates c ON c.execution_id = e.id").
		Joins("LEFT JOIN wasm_execution_audit w ON w.execution_id = e.id").
		Select(`
			e.id::text as id,
			e.function_id,
			f.name as function_name,
			f.author,
			e.version,
			e.timestamp,
			e.duration_ms,
			e.outcome,
			e.error_code,
			e.tenant_id,
			w.input_size,
			w.output_size,
			c.execution_root_hash,
			c.node_signature
		`)

	// Apply filters
	if searchTerm != "" {
		q := "%" + searchTerm + "%"
		query = query.Where("(f.name ILIKE ? OR f.author ILIKE ?)", q, q)
	}
	if tenantFilter != "" {
		query = query.Where("f.author = ?", tenantFilter)
	}
	if statusFilter != "" {
		query = query.Where("e.outcome = ?", statusFilter)
	}

	// Get total count for pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count executions: %w", err)
	}

	// Apply ordering and pagination
	var rows []ExecutionAuditRow
	err := query.Order("e.timestamp DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get execution audit data: %w", err)
	}

	return rows, total, nil
}
