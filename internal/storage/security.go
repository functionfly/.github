package storage

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
)

// CreateSecurityScan creates a new security scan record
func (db *PostgresDB) CreateSecurityScan(ctx context.Context, scan *SecurityScan) (*SecurityScan, error) {
	configJSON, err := json.Marshal(scan.Config)
	if err != nil {
		return nil, err
	}

	summaryJSON, err := json.Marshal(scan.Summary)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO security_scans (
			tenant_id, user_id, scan_type, status, target, config, summary,
			started_at, completed_at, duration_ms
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	err = db.QueryRowContext(ctx, query,
		scan.TenantID,
		scan.UserID,
		scan.ScanType,
		scan.Status,
		scan.Target,
		configJSON,
		summaryJSON,
		scan.StartedAt,
		scan.CompletedAt,
		scan.DurationMs,
	).Scan(&scan.ID, &scan.CreatedAt, &scan.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return scan, nil
}

// UpdateSecurityScan updates an existing security scan
func (db *PostgresDB) UpdateSecurityScan(ctx context.Context, scanID uuid.UUID, updates map[string]interface{}) (*SecurityScan, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argCount := 1

	for field, value := range updates {
		switch field {
		case "status", "target", "completed_at", "duration_ms":
			setParts = append(setParts, field+" = $"+string(rune(argCount+'0')))
			args = append(args, value)
			argCount++
		case "config", "summary":
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			setParts = append(setParts, field+" = $"+string(rune(argCount+'0')))
			args = append(args, jsonBytes)
			argCount++
		}
	}

	if len(setParts) == 0 {
		return nil, nil // No updates
	}

	args = append(args, scanID)
	query := "UPDATE security_scans SET " + strings.Join(setParts, ", ") +
		", updated_at = NOW() WHERE id = $" + string(rune(argCount+'0')) +
		" RETURNING id, tenant_id, user_id, scan_type, status, target, config, summary, started_at, completed_at, duration_ms, created_at, updated_at"

	scan := &SecurityScan{}
	var configJSON, summaryJSON []byte

	err := db.QueryRowContext(ctx, query, args...).Scan(
		&scan.ID,
		&scan.TenantID,
		&scan.UserID,
		&scan.ScanType,
		&scan.Status,
		&scan.Target,
		&configJSON,
		&summaryJSON,
		&scan.StartedAt,
		&scan.CompletedAt,
		&scan.DurationMs,
		&scan.CreatedAt,
		&scan.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	if len(configJSON) > 0 {
		err = json.Unmarshal(configJSON, &scan.Config)
		if err != nil {
			return nil, err
		}
	}

	if len(summaryJSON) > 0 {
		err = json.Unmarshal(summaryJSON, &scan.Summary)
		if err != nil {
			return nil, err
		}
	}

	return scan, nil
}

// GetSecurityScan retrieves a security scan by ID
func (db *PostgresDB) GetSecurityScan(ctx context.Context, scanID uuid.UUID) (*SecurityScan, error) {
	query := `
		SELECT id, tenant_id, user_id, scan_type, status, target, config, summary,
			   started_at, completed_at, duration_ms, created_at, updated_at
		FROM security_scans WHERE id = $1`

	scan := &SecurityScan{}
	var configJSON, summaryJSON []byte

	err := db.QueryRowContext(ctx, query, scanID).Scan(
		&scan.ID,
		&scan.TenantID,
		&scan.UserID,
		&scan.ScanType,
		&scan.Status,
		&scan.Target,
		&configJSON,
		&summaryJSON,
		&scan.StartedAt,
		&scan.CompletedAt,
		&scan.DurationMs,
		&scan.CreatedAt,
		&scan.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	if len(configJSON) > 0 {
		err = json.Unmarshal(configJSON, &scan.Config)
		if err != nil {
			return nil, err
		}
	}

	if len(summaryJSON) > 0 {
		err = json.Unmarshal(summaryJSON, &scan.Summary)
		if err != nil {
			return nil, err
		}
	}

	return scan, nil
}

// ListSecurityScans retrieves a list of security scans with optional filtering
func (db *PostgresDB) ListSecurityScans(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*SecurityScan, error) {
	query := `
		SELECT id, tenant_id, user_id, scan_type, status, target, config, summary,
			   started_at, completed_at, duration_ms, created_at, updated_at
		FROM security_scans`

	whereParts := []string{}
	args := []interface{}{}
	argCount := 1

	// Add filters
	if tenantID, ok := filters["tenant_id"]; ok {
		whereParts = append(whereParts, "tenant_id = $"+string(rune(argCount+'0')))
		args = append(args, tenantID)
		argCount++
	}

	if userID, ok := filters["user_id"]; ok {
		whereParts = append(whereParts, "user_id = $"+string(rune(argCount+'0')))
		args = append(args, userID)
		argCount++
	}

	if scanType, ok := filters["scan_type"]; ok {
		whereParts = append(whereParts, "scan_type = $"+string(rune(argCount+'0')))
		args = append(args, scanType)
		argCount++
	}

	if status, ok := filters["status"]; ok {
		whereParts = append(whereParts, "status = $"+string(rune(argCount+'0')))
		args = append(args, status)
		argCount++
	}

	if len(whereParts) > 0 {
		query += " WHERE " + strings.Join(whereParts, " AND ")
	}

	query += " ORDER BY started_at DESC LIMIT $" + string(rune(argCount+'0')) +
		" OFFSET $" + string(rune(argCount+1+'0'))
	args = append(args, limit, offset)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scans []*SecurityScan
	for rows.Next() {
		scan := &SecurityScan{}
		var configJSON, summaryJSON []byte

		err := rows.Scan(
			&scan.ID,
			&scan.TenantID,
			&scan.UserID,
			&scan.ScanType,
			&scan.Status,
			&scan.Target,
			&configJSON,
			&summaryJSON,
			&scan.StartedAt,
			&scan.CompletedAt,
			&scan.DurationMs,
			&scan.CreatedAt,
			&scan.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		if len(configJSON) > 0 {
			err = json.Unmarshal(configJSON, &scan.Config)
			if err != nil {
				return nil, err
			}
		}

		if len(summaryJSON) > 0 {
			err = json.Unmarshal(summaryJSON, &scan.Summary)
			if err != nil {
				return nil, err
			}
		}

		scans = append(scans, scan)
	}

	return scans, rows.Err()
}

// CreateVulnerability creates a new vulnerability record
func (db *PostgresDB) CreateVulnerability(ctx context.Context, vuln *Vulnerability) (*Vulnerability, error) {
	referenceUrlsJSON, err := json.Marshal(vuln.ReferenceUrls)
	if err != nil {
		return nil, err
	}

	metadataJSON, err := json.Marshal(vuln.Metadata)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO vulnerabilities (
			scan_id, title, description, severity, cvss_score, cve, category,
			component, location, status, remediation, reference_urls, metadata, discovered_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at`

	err = db.QueryRowContext(ctx, query,
		vuln.ScanID,
		vuln.Title,
		vuln.Description,
		vuln.Severity,
		vuln.CVSS,
		vuln.CVE,
		vuln.Category,
		vuln.Component,
		vuln.Location,
		vuln.Status,
		vuln.Remediation,
		referenceUrlsJSON,
		metadataJSON,
		vuln.DiscoveredAt,
	).Scan(&vuln.ID, &vuln.CreatedAt, &vuln.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return vuln, nil
}

// UpdateVulnerability updates an existing vulnerability
func (db *PostgresDB) UpdateVulnerability(ctx context.Context, vulnID uuid.UUID, updates map[string]interface{}) (*Vulnerability, error) {
	setParts := []string{}
	args := []interface{}{}
	argCount := 1

	for field, value := range updates {
		switch field {
		case "title", "description", "severity", "category", "component", "status", "remediation":
			setParts = append(setParts, field+" = $"+string(rune(argCount+'0')))
			args = append(args, value)
			argCount++
		case "cvss_score":
			setParts = append(setParts, "cvss_score = $"+string(rune(argCount+'0')))
			args = append(args, value)
			argCount++
		case "cve", "location":
			setParts = append(setParts, field+" = $"+string(rune(argCount+'0')))
			args = append(args, value)
			argCount++
		case "reference_urls":
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			setParts = append(setParts, "reference_urls = $"+string(rune(argCount+'0')))
			args = append(args, jsonBytes)
			argCount++
		case "metadata":
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				return nil, err
			}
			setParts = append(setParts, "metadata = $"+string(rune(argCount+'0')))
			args = append(args, jsonBytes)
			argCount++
		}
	}

	if len(setParts) == 0 {
		return nil, nil // No updates
	}

	args = append(args, vulnID)
	query := "UPDATE vulnerabilities SET " + strings.Join(setParts, ", ") +
		", updated_at = NOW() WHERE id = $" + string(rune(argCount+'0')) +
		" RETURNING id, scan_id, title, description, severity, cvss_score, cve, category, component, location, status, remediation, references, metadata, discovered_at, created_at, updated_at"

	vuln := &Vulnerability{}
	var referenceUrlsJSON, metadataJSON []byte

	err := db.QueryRowContext(ctx, query, args...).Scan(
		&vuln.ID,
		&vuln.ScanID,
		&vuln.Title,
		&vuln.Description,
		&vuln.Severity,
		&vuln.CVSS,
		&vuln.CVE,
		&vuln.Category,
		&vuln.Component,
		&vuln.Location,
		&vuln.Status,
		&vuln.Remediation,
		&referenceUrlsJSON,
		&metadataJSON,
		&vuln.DiscoveredAt,
		&vuln.CreatedAt,
		&vuln.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	if len(referenceUrlsJSON) > 0 {
		err = json.Unmarshal(referenceUrlsJSON, &vuln.ReferenceUrls)
		if err != nil {
			return nil, err
		}
	}

	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &vuln.Metadata)
		if err != nil {
			return nil, err
		}
	}

	return vuln, nil
}

// GetVulnerabilities retrieves vulnerabilities with optional filtering
func (db *PostgresDB) GetVulnerabilities(ctx context.Context, filters map[string]interface{}) ([]*Vulnerability, error) {
	query := `
		SELECT id, scan_id, title, description, severity, cvss_score, cve, category,
			   component, location, status, remediation, reference_urls, metadata,
			   discovered_at, created_at, updated_at
		FROM vulnerabilities`

	whereParts := []string{}
	args := []interface{}{}
	argCount := 1

	// Add filters
	if scanID, ok := filters["scan_id"]; ok {
		whereParts = append(whereParts, "scan_id = $"+string(rune(argCount+'0')))
		args = append(args, scanID)
		argCount++
	}

	if severity, ok := filters["severity"]; ok {
		whereParts = append(whereParts, "severity = $"+string(rune(argCount+'0')))
		args = append(args, severity)
		argCount++
	}

	if status, ok := filters["status"]; ok {
		whereParts = append(whereParts, "status = $"+string(rune(argCount+'0')))
		args = append(args, status)
		argCount++
	}

	if category, ok := filters["category"]; ok {
		whereParts = append(whereParts, "category = $"+string(rune(argCount+'0')))
		args = append(args, category)
		argCount++
	}

	if component, ok := filters["component"]; ok {
		whereParts = append(whereParts, "component ILIKE $"+string(rune(argCount+'0')))
		args = append(args, "%"+component.(string)+"%")
		argCount++
	}

	if len(whereParts) > 0 {
		query += " WHERE " + strings.Join(whereParts, " AND ")
	}

	query += " ORDER BY discovered_at DESC"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vulnerabilities []*Vulnerability
	for rows.Next() {
		vuln := &Vulnerability{}
		var referenceUrlsJSON, metadataJSON []byte

		err := rows.Scan(
			&vuln.ID,
			&vuln.ScanID,
			&vuln.Title,
			&vuln.Description,
			&vuln.Severity,
			&vuln.CVSS,
			&vuln.CVE,
			&vuln.Category,
			&vuln.Component,
			&vuln.Location,
			&vuln.Status,
			&vuln.Remediation,
			&referenceUrlsJSON,
			&metadataJSON,
			&vuln.DiscoveredAt,
			&vuln.CreatedAt,
			&vuln.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		if len(referenceUrlsJSON) > 0 {
			err = json.Unmarshal(referenceUrlsJSON, &vuln.ReferenceUrls)
			if err != nil {
				return nil, err
			}
		}

		if len(metadataJSON) > 0 {
			err = json.Unmarshal(metadataJSON, &vuln.Metadata)
			if err != nil {
				return nil, err
			}
		}

		vulnerabilities = append(vulnerabilities, vuln)
	}

	return vulnerabilities, rows.Err()
}

// GetVulnerabilityByID retrieves a vulnerability by its ID
func (db *PostgresDB) GetVulnerabilityByID(ctx context.Context, vulnID uuid.UUID) (*Vulnerability, error) {
	query := `
		SELECT id, scan_id, title, description, severity, cvss_score, cve, category,
			   component, location, status, remediation, reference_urls, metadata,
			   discovered_at, created_at, updated_at
		FROM vulnerabilities WHERE id = $1`

	vuln := &Vulnerability{}
	var referenceUrlsJSON, metadataJSON []byte

	err := db.QueryRowContext(ctx, query, vulnID).Scan(
		&vuln.ID,
		&vuln.ScanID,
		&vuln.Title,
		&vuln.Description,
		&vuln.Severity,
		&vuln.CVSS,
		&vuln.CVE,
		&vuln.Category,
		&vuln.Component,
		&vuln.Location,
		&vuln.Status,
		&vuln.Remediation,
		&referenceUrlsJSON,
		&metadataJSON,
		&vuln.DiscoveredAt,
		&vuln.CreatedAt,
		&vuln.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	if len(referenceUrlsJSON) > 0 {
		err = json.Unmarshal(referenceUrlsJSON, &vuln.ReferenceUrls)
		if err != nil {
			return nil, err
		}
	}

	if len(metadataJSON) > 0 {
		err = json.Unmarshal(metadataJSON, &vuln.Metadata)
		if err != nil {
			return nil, err
		}
	}

	return vuln, nil
}