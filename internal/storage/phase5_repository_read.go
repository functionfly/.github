package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// GetFWOSIncidentByID retrieves an incident by ID.
func (r *Phase5Repository) GetFWOSIncidentByID(ctx context.Context, id uuid.UUID) (*FWOSIncident, error) {
	inc := &FWOSIncident{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, title, description, severity, status, commander_id, project_id, detected_at, acknowledged_at, resolved_at, closed_at, root_cause, impact, duration_minutes, created_at, updated_at
		FROM incidents WHERE id = $1`, id).Scan(
		&inc.ID, &inc.TenantID, &inc.Title, &inc.Description, &inc.Severity, &inc.Status, &inc.CommanderID, &inc.ProjectID, &inc.DetectedAt, &inc.AcknowledgedAt, &inc.ResolvedAt, &inc.ClosedAt, &inc.RootCause, &inc.Impact, &inc.DurationMinutes, &inc.CreatedAt, &inc.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get incident: %w", err)
	}
	return inc, nil
}

// ListFWOSIncidents lists incidents for a tenant.
func (r *Phase5Repository) ListFWOSIncidents(ctx context.Context, tenantID uuid.UUID, opts ListIncidentsOpts) ([]*FWOSIncident, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *opts.Status)
		argIdx++
	}
	if opts.Severity != nil {
		where += fmt.Sprintf(" AND severity = $%d", argIdx)
		args = append(args, *opts.Severity)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM incidents %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count incidents: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY detected_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, tenant_id, title, description, severity, status, commander_id, project_id, detected_at, acknowledged_at, resolved_at, closed_at, root_cause, impact, duration_minutes, created_at, updated_at
		FROM incidents %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list incidents: %w", err)
	}
	defer rows.Close()

	var incidents []*FWOSIncident
	for rows.Next() {
		inc := &FWOSIncident{}
		if err := rows.Scan(&inc.ID, &inc.TenantID, &inc.Title, &inc.Description, &inc.Severity, &inc.Status, &inc.CommanderID, &inc.ProjectID, &inc.DetectedAt, &inc.AcknowledgedAt, &inc.ResolvedAt, &inc.ClosedAt, &inc.RootCause, &inc.Impact, &inc.DurationMinutes, &inc.CreatedAt, &inc.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan incident: %w", err)
		}
		incidents = append(incidents, inc)
	}
	return incidents, total, nil
}

// ListIncidentEvents lists timeline events for an incident.
func (r *Phase5Repository) ListIncidentEvents(ctx context.Context, incidentID uuid.UUID, limit, offset int) ([]*IncidentEvent, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, incident_id, author_id, event_type, body, metadata, created_at
		FROM incident_events WHERE incident_id = $1 ORDER BY created_at ASC LIMIT $2 OFFSET $3`,
		incidentID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list incident events: %w", err)
	}
	defer rows.Close()

	var events []*IncidentEvent
	for rows.Next() {
		ev := &IncidentEvent{}
		var metaBytes []byte
		if err := rows.Scan(&ev.ID, &ev.IncidentID, &ev.AuthorID, &ev.EventType, &ev.Body, &metaBytes, &ev.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan incident event: %w", err)
		}
		if metaBytes != nil {
			var meta JSONMap
			if err := json.Unmarshal(metaBytes, &meta); err == nil {
				ev.Metadata = meta
			}
		}
		events = append(events, ev)
	}
	return events, nil
}

// ListIncidentResponders lists responders for an incident.
func (r *Phase5Repository) ListIncidentResponders(ctx context.Context, incidentID uuid.UUID) ([]*IncidentResponder, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, incident_id, employee_id, role, joined_at, left_at
		FROM incident_responders WHERE incident_id = $1 ORDER BY joined_at ASC`,
		incidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list incident responders: %w", err)
	}
	defer rows.Close()

	var responders []*IncidentResponder
	for rows.Next() {
		resp := &IncidentResponder{}
		if err := rows.Scan(&resp.ID, &resp.IncidentID, &resp.EmployeeID, &resp.Role, &resp.JoinedAt, &resp.LeftAt); err != nil {
			return nil, fmt.Errorf("failed to scan incident responder: %w", err)
		}
		responders = append(responders, resp)
	}
	return responders, nil
}

// GetPostmortemByIncident retrieves a postmortem for an incident.
func (r *Phase5Repository) GetPostmortemByIncident(ctx context.Context, incidentID uuid.UUID) (*Postmortem, error) {
	pm := &Postmortem{}
	var actionItemsBytes []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, incident_id, tenant_id, author_id, summary, root_cause, contributing_factors, what_went_well, what_went_wrong, action_items, lessons_learned, status, published_at, created_at, updated_at
		FROM postmortems WHERE incident_id = $1`, incidentID).Scan(
		&pm.ID, &pm.IncidentID, &pm.TenantID, &pm.AuthorID, &pm.Summary, &pm.RootCause, &pm.ContributingFactors, &pm.WhatWentWell, &pm.WhatWentWrong, &actionItemsBytes, &pm.LessonsLearned, &pm.Status, &pm.PublishedAt, &pm.CreatedAt, &pm.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get postmortem: %w", err)
	}
	if actionItemsBytes != nil {
		var items JSONMap
		if err := json.Unmarshal(actionItemsBytes, &items); err == nil {
			pm.ActionItems = items
		}
	}
	return pm, nil
}

// ListLifecycleEvents lists lifecycle events for an employee.
func (r *Phase5Repository) ListLifecycleEvents(ctx context.Context, employeeID uuid.UUID, opts ListLifecycleEventsOpts) ([]*LifecycleEvent, int, error) {
	where := "WHERE employee_id = $1"
	args := []interface{}{employeeID}
	argIdx := 2

	if opts.EventType != nil {
		where += fmt.Sprintf(" AND event_type = $%d", argIdx)
		args = append(args, *opts.EventType)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM lifecycle_events %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count lifecycle events: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, employee_id, tenant_id, event_type, payload, triggered_by, notes, created_at
		FROM lifecycle_events %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list lifecycle events: %w", err)
	}
	defer rows.Close()

	var events []*LifecycleEvent
	for rows.Next() {
		ev := &LifecycleEvent{}
		var payloadBytes []byte
		if err := rows.Scan(&ev.ID, &ev.EmployeeID, &ev.TenantID, &ev.EventType, &payloadBytes, &ev.TriggeredBy, &ev.Notes, &ev.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan lifecycle event: %w", err)
		}
		if payloadBytes != nil {
			var payload JSONMap
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				ev.Payload = payload
			}
		}
		events = append(events, ev)
	}
	return events, total, nil
}

// ListLifecycleWorkflows lists lifecycle workflow templates for a tenant.
func (r *Phase5Repository) ListLifecycleWorkflows(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*LifecycleWorkflow, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM lifecycle_workflows WHERE tenant_id = $1", tenantID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count lifecycle workflows: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, name, description, trigger_event, steps, is_active, created_at, updated_at
		FROM lifecycle_workflows WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list lifecycle workflows: %w", err)
	}
	defer rows.Close()

	var workflows []*LifecycleWorkflow
	for rows.Next() {
		wf := &LifecycleWorkflow{}
		var stepsBytes []byte
		if err := rows.Scan(&wf.ID, &wf.TenantID, &wf.Name, &wf.Description, &wf.TriggerEvent, &stepsBytes, &wf.IsActive, &wf.CreatedAt, &wf.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan lifecycle workflow: %w", err)
		}
		if stepsBytes != nil {
			var steps JSONMap
			if err := json.Unmarshal(stepsBytes, &steps); err == nil {
				wf.Steps = steps
			}
		}
		workflows = append(workflows, wf)
	}
	return workflows, total, nil
}

// GetLifecycleWorkflowByID retrieves a lifecycle workflow by ID.
func (r *Phase5Repository) GetLifecycleWorkflowByID(ctx context.Context, id uuid.UUID) (*LifecycleWorkflow, error) {
	wf := &LifecycleWorkflow{}
	var stepsBytes []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, description, trigger_event, steps, is_active, created_at, updated_at
		FROM lifecycle_workflows WHERE id = $1`, id).Scan(
		&wf.ID, &wf.TenantID, &wf.Name, &wf.Description, &wf.TriggerEvent, &stepsBytes, &wf.IsActive, &wf.CreatedAt, &wf.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lifecycle workflow: %w", err)
	}
	if stepsBytes != nil {
		var steps JSONMap
		if err := json.Unmarshal(stepsBytes, &steps); err == nil {
			wf.Steps = steps
		}
	}
	return wf, nil
}

// GetLifecycleWorkflowInstance retrieves a workflow instance by ID.
func (r *Phase5Repository) GetLifecycleWorkflowInstance(ctx context.Context, id uuid.UUID) (*LifecycleWorkflowInstance, error) {
	inst := &LifecycleWorkflowInstance{}
	var stepsStatusBytes []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, workflow_id, employee_id, tenant_id, status, current_step, steps_status, started_at, completed_at, created_at, updated_at
		FROM lifecycle_workflow_instances WHERE id = $1`, id).Scan(
		&inst.ID, &inst.WorkflowID, &inst.EmployeeID, &inst.TenantID, &inst.Status, &inst.CurrentStep, &stepsStatusBytes, &inst.StartedAt, &inst.CompletedAt, &inst.CreatedAt, &inst.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lifecycle workflow instance: %w", err)
	}
	if stepsStatusBytes != nil {
		var status JSONMap
		if err := json.Unmarshal(stepsStatusBytes, &status); err == nil {
			inst.StepsStatus = status
		}
	}
	return inst, nil
}

// GetFeatureFlagByKey retrieves a feature flag by tenant and key.
func (r *Phase5Repository) GetFeatureFlagByKey(ctx context.Context, tenantID uuid.UUID, key string) (*FeatureFlag, error) {
	ff := &FeatureFlag{}
	var variantsBytes, audienceBytes []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, key, name, description, flag_type, is_enabled, rollout_pct, variants, target_audience, created_at, updated_at
		FROM feature_flags WHERE tenant_id = $1 AND key = $2`, tenantID, key).Scan(
		&ff.ID, &ff.TenantID, &ff.Key, &ff.Name, &ff.Description, &ff.FlagType, &ff.IsEnabled, &ff.RolloutPct, &variantsBytes, &audienceBytes, &ff.CreatedAt, &ff.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get feature flag: %w", err)
	}
	if variantsBytes != nil {
		var v JSONMap
		if err := json.Unmarshal(variantsBytes, &v); err == nil {
			ff.Variants = v
		}
	}
	if audienceBytes != nil {
		var a JSONMap
		if err := json.Unmarshal(audienceBytes, &a); err == nil {
			ff.TargetAudience = a
		}
	}
	return ff, nil
}

// ListFeatureFlags lists feature flags for a tenant.
func (r *Phase5Repository) ListFeatureFlags(ctx context.Context, tenantID uuid.UUID, opts ListFeatureFlagsOpts) ([]*FeatureFlag, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.IsEnabled != nil {
		where += fmt.Sprintf(" AND is_enabled = $%d", argIdx)
		args = append(args, *opts.IsEnabled)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM feature_flags %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count feature flags: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, tenant_id, key, name, description, flag_type, is_enabled, rollout_pct, variants, target_audience, created_at, updated_at
		FROM feature_flags %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list feature flags: %w", err)
	}
	defer rows.Close()

	var flags []*FeatureFlag
	for rows.Next() {
		ff := &FeatureFlag{}
		var variantsBytes, audienceBytes []byte
		if err := rows.Scan(&ff.ID, &ff.TenantID, &ff.Key, &ff.Name, &ff.Description, &ff.FlagType, &ff.IsEnabled, &ff.RolloutPct, &variantsBytes, &audienceBytes, &ff.CreatedAt, &ff.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan feature flag: %w", err)
		}
		if variantsBytes != nil {
			var v JSONMap
			if err := json.Unmarshal(variantsBytes, &v); err == nil {
				ff.Variants = v
			}
		}
		if audienceBytes != nil {
			var a JSONMap
			if err := json.Unmarshal(audienceBytes, &a); err == nil {
				ff.TargetAudience = a
			}
		}
		flags = append(flags, ff)
	}
	return flags, total, nil
}

// GetDataClassification retrieves a classification for a specific resource.
func (r *Phase5Repository) GetDataClassification(ctx context.Context, tenantID uuid.UUID, resourceType string, resourceID uuid.UUID) (*DataClassification, error) {
	dc := &DataClassification{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, resource_type, resource_id, classification, classified_by, reason, expires_at, created_at, updated_at
		FROM data_classifications WHERE tenant_id = $1 AND resource_type = $2 AND resource_id = $3`, tenantID, resourceType, resourceID).Scan(
		&dc.ID, &dc.TenantID, &dc.ResourceType, &dc.ResourceID, &dc.Classification, &dc.ClassifiedBy, &dc.Reason, &dc.ExpiresAt, &dc.CreatedAt, &dc.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get data classification: %w", err)
	}
	return dc, nil
}

// ListDataClassifications lists data classifications for a tenant.
func (r *Phase5Repository) ListDataClassifications(ctx context.Context, tenantID uuid.UUID, opts ListDataClassificationsOpts) ([]*DataClassification, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.ResourceType != nil {
		where += fmt.Sprintf(" AND resource_type = $%d", argIdx)
		args = append(args, *opts.ResourceType)
		argIdx++
	}
	if opts.Classification != nil {
		where += fmt.Sprintf(" AND classification = $%d", argIdx)
		args = append(args, *opts.Classification)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM data_classifications %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count data classifications: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, tenant_id, resource_type, resource_id, classification, classified_by, reason, expires_at, created_at, updated_at
		FROM data_classifications %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list data classifications: %w", err)
	}
	defer rows.Close()

	var dcs []*DataClassification
	for rows.Next() {
		dc := &DataClassification{}
		if err := rows.Scan(&dc.ID, &dc.TenantID, &dc.ResourceType, &dc.ResourceID, &dc.Classification, &dc.ClassifiedBy, &dc.Reason, &dc.ExpiresAt, &dc.CreatedAt, &dc.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan data classification: %w", err)
		}
		dcs = append(dcs, dc)
	}
	return dcs, total, nil
}

// GetCertificateBySerial retrieves a certificate by serial number.
func (r *Phase5Repository) GetCertificateBySerial(ctx context.Context, serial string) (*EmployeeCertificate, error) {
	cert := &EmployeeCertificate{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, employee_id, tenant_id, certificate_serial, certificate_type, subject, issuer, public_key_pem, fingerprint, device_id, device_name, issued_at, expires_at, revoked_at, revoke_reason, status, created_at
		FROM employee_certificates WHERE certificate_serial = $1`, serial).Scan(
		&cert.ID, &cert.EmployeeID, &cert.TenantID, &cert.CertificateSerial, &cert.CertificateType, &cert.Subject, &cert.Issuer, &cert.PublicKeyPEM, &cert.Fingerprint, &cert.DeviceID, &cert.DeviceName, &cert.IssuedAt, &cert.ExpiresAt, &cert.RevokedAt, &cert.RevokeReason, &cert.Status, &cert.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}
	return cert, nil
}

// ListCertificates lists certificates for an employee.
func (r *Phase5Repository) ListCertificates(ctx context.Context, employeeID uuid.UUID, opts ListCertificatesOpts) ([]*EmployeeCertificate, int, error) {
	where := "WHERE employee_id = $1"
	args := []interface{}{employeeID}
	argIdx := 2

	if opts.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *opts.Status)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM employee_certificates %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count certificates: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY issued_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, employee_id, tenant_id, certificate_serial, certificate_type, subject, issuer, public_key_pem, fingerprint, device_id, device_name, issued_at, expires_at, revoked_at, revoke_reason, status, created_at
		FROM employee_certificates %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list certificates: %w", err)
	}
	defer rows.Close()

	var certs []*EmployeeCertificate
	for rows.Next() {
		cert := &EmployeeCertificate{}
		if err := rows.Scan(&cert.ID, &cert.EmployeeID, &cert.TenantID, &cert.CertificateSerial, &cert.CertificateType, &cert.Subject, &cert.Issuer, &cert.PublicKeyPEM, &cert.Fingerprint, &cert.DeviceID, &cert.DeviceName, &cert.IssuedAt, &cert.ExpiresAt, &cert.RevokedAt, &cert.RevokeReason, &cert.Status, &cert.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan certificate: %w", err)
		}
		certs = append(certs, cert)
	}
	return certs, total, nil
}

// ListFWOSEvents lists events from the FWOS event log.
func (r *Phase5Repository) ListFWOSEvents(ctx context.Context, tenantID uuid.UUID, opts ListFWOSEventsOpts) ([]*FWOSEvent, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.EventType != nil {
		where += fmt.Sprintf(" AND event_type = $%d", argIdx)
		args = append(args, *opts.EventType)
		argIdx++
	}
	if opts.Source != nil {
		where += fmt.Sprintf(" AND source = $%d", argIdx)
		args = append(args, *opts.Source)
		argIdx++
	}
	if opts.ResourceType != nil {
		where += fmt.Sprintf(" AND resource_type = $%d", argIdx)
		args = append(args, *opts.ResourceType)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM fwos_events %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count FWOS events: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, tenant_id, event_type, source, actor_id, resource_type, resource_id, payload, created_at
		FROM fwos_events %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list FWOS events: %w", err)
	}
	defer rows.Close()

	var events []*FWOSEvent
	for rows.Next() {
		ev := &FWOSEvent{}
		var payloadBytes []byte
		if err := rows.Scan(&ev.ID, &ev.TenantID, &ev.EventType, &ev.Source, &ev.ActorID, &ev.ResourceType, &ev.ResourceID, &payloadBytes, &ev.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan FWOS event: %w", err)
		}
		if payloadBytes != nil {
			var payload JSONMap
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				ev.Payload = payload
			}
		}
		events = append(events, ev)
	}
	return events, total, nil
}
