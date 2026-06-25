package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateFWOSIncident creates a new incident.
func (r *Phase5Repository) CreateFWOSIncident(ctx context.Context, inc *FWOSIncident) (*FWOSIncident, error) {
	inc.ID = uuid.New()
	inc.CreatedAt = time.Now()
	inc.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO incidents (id, tenant_id, title, description, severity, status, commander_id, project_id, detected_at, acknowledged_at, resolved_at, closed_at, root_cause, impact, duration_minutes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
		inc.ID, inc.TenantID, inc.Title, inc.Description, inc.Severity, inc.Status, inc.CommanderID, inc.ProjectID, inc.DetectedAt, inc.AcknowledgedAt, inc.ResolvedAt, inc.ClosedAt, inc.RootCause, inc.Impact, inc.DurationMinutes, inc.CreatedAt, inc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create incident: %w", err)
	}
	return inc, nil
}

// CreateIncidentEvent creates a timeline event for an incident.
func (r *Phase5Repository) CreateIncidentEvent(ctx context.Context, ev *IncidentEvent) (*IncidentEvent, error) {
	ev.CreatedAt = time.Now()

	var metaParam interface{}
	if ev.Metadata != nil {
		b, _ := json.Marshal(ev.Metadata)
		metaParam = b
	}

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO incident_events (incident_id, author_id, event_type, body, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`,
		ev.IncidentID, ev.AuthorID, ev.EventType, ev.Body, metaParam, ev.CreatedAt,
	).Scan(&ev.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create incident event: %w", err)
	}
	return ev, nil
}

// AddIncidentResponder adds a responder to an incident.
func (r *Phase5Repository) AddIncidentResponder(ctx context.Context, resp *IncidentResponder) (*IncidentResponder, error) {
	resp.JoinedAt = time.Now()

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO incident_responders (incident_id, employee_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		resp.IncidentID, resp.EmployeeID, resp.Role, resp.JoinedAt,
	).Scan(&resp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to add incident responder: %w", err)
	}
	return resp, nil
}

// CreatePostmortem creates a new postmortem.
func (r *Phase5Repository) CreatePostmortem(ctx context.Context, pm *Postmortem) (*Postmortem, error) {
	pm.ID = uuid.New()
	pm.CreatedAt = time.Now()
	pm.UpdatedAt = time.Now()

	var actionItemsParam interface{}
	if pm.ActionItems != nil {
		b, _ := json.Marshal(pm.ActionItems)
		actionItemsParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO postmortems (id, incident_id, tenant_id, author_id, summary, root_cause, contributing_factors, what_went_well, what_went_wrong, action_items, lessons_learned, status, published_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		pm.ID, pm.IncidentID, pm.TenantID, pm.AuthorID, pm.Summary, pm.RootCause, pm.ContributingFactors, pm.WhatWentWell, pm.WhatWentWrong, actionItemsParam, pm.LessonsLearned, pm.Status, pm.PublishedAt, pm.CreatedAt, pm.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create postmortem: %w", err)
	}
	return pm, nil
}

// CreateLifecycleEvent creates a lifecycle event for an employee.
func (r *Phase5Repository) CreateLifecycleEvent(ctx context.Context, ev *LifecycleEvent) (*LifecycleEvent, error) {
	ev.ID = uuid.New()
	ev.CreatedAt = time.Now()

	var payloadParam interface{}
	if ev.Payload != nil {
		b, _ := json.Marshal(ev.Payload)
		payloadParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO lifecycle_events (id, employee_id, tenant_id, event_type, payload, triggered_by, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		ev.ID, ev.EmployeeID, ev.TenantID, ev.EventType, payloadParam, ev.TriggeredBy, ev.Notes, ev.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create lifecycle event: %w", err)
	}
	return ev, nil
}

// CreateLifecycleWorkflow creates a lifecycle workflow template.
func (r *Phase5Repository) CreateLifecycleWorkflow(ctx context.Context, wf *LifecycleWorkflow) (*LifecycleWorkflow, error) {
	wf.ID = uuid.New()
	wf.CreatedAt = time.Now()
	wf.UpdatedAt = time.Now()

	var stepsParam interface{}
	if wf.Steps != nil {
		b, _ := json.Marshal(wf.Steps)
		stepsParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO lifecycle_workflows (id, tenant_id, name, description, trigger_event, steps, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		wf.ID, wf.TenantID, wf.Name, wf.Description, wf.TriggerEvent, stepsParam, wf.IsActive, wf.CreatedAt, wf.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create lifecycle workflow: %w", err)
	}
	return wf, nil
}

// CreateLifecycleWorkflowInstance starts a new workflow instance.
func (r *Phase5Repository) CreateLifecycleWorkflowInstance(ctx context.Context, inst *LifecycleWorkflowInstance) (*LifecycleWorkflowInstance, error) {
	inst.ID = uuid.New()
	inst.CreatedAt = time.Now()
	inst.UpdatedAt = time.Now()

	var stepsStatusParam interface{}
	if inst.StepsStatus != nil {
		b, _ := json.Marshal(inst.StepsStatus)
		stepsStatusParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO lifecycle_workflow_instances (id, workflow_id, employee_id, tenant_id, status, current_step, steps_status, started_at, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		inst.ID, inst.WorkflowID, inst.EmployeeID, inst.TenantID, inst.Status, inst.CurrentStep, stepsStatusParam, inst.StartedAt, inst.CompletedAt, inst.CreatedAt, inst.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create lifecycle workflow instance: %w", err)
	}
	return inst, nil
}

// CreateFeatureFlag creates a new feature flag.
func (r *Phase5Repository) CreateFeatureFlag(ctx context.Context, ff *FeatureFlag) (*FeatureFlag, error) {
	ff.ID = uuid.New()
	ff.CreatedAt = time.Now()
	ff.UpdatedAt = time.Now()

	var variantsParam, audienceParam interface{}
	if ff.Variants != nil {
		b, _ := json.Marshal(ff.Variants)
		variantsParam = b
	}
	if ff.TargetAudience != nil {
		b, _ := json.Marshal(ff.TargetAudience)
		audienceParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO feature_flags (id, tenant_id, key, name, description, flag_type, is_enabled, rollout_pct, variants, target_audience, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		ff.ID, ff.TenantID, ff.Key, ff.Name, ff.Description, ff.FlagType, ff.IsEnabled, ff.RolloutPct, variantsParam, audienceParam, ff.CreatedAt, ff.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create feature flag: %w", err)
	}
	return ff, nil
}

// CreateDataClassification creates a data classification label.
func (r *Phase5Repository) CreateDataClassification(ctx context.Context, dc *DataClassification) (*DataClassification, error) {
	dc.ID = uuid.New()
	dc.CreatedAt = time.Now()
	dc.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO data_classifications (id, tenant_id, resource_type, resource_id, classification, classified_by, reason, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		dc.ID, dc.TenantID, dc.ResourceType, dc.ResourceID, dc.Classification, dc.ClassifiedBy, dc.Reason, dc.ExpiresAt, dc.CreatedAt, dc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create data classification: %w", err)
	}
	return dc, nil
}

// IssueCertificate issues a new employee certificate.
func (r *Phase5Repository) IssueCertificate(ctx context.Context, cert *EmployeeCertificate) (*EmployeeCertificate, error) {
	cert.ID = uuid.New()
	cert.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO employee_certificates (id, employee_id, tenant_id, certificate_serial, certificate_type, subject, issuer, public_key_pem, fingerprint, device_id, device_name, issued_at, expires_at, revoked_at, revoke_reason, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
		cert.ID, cert.EmployeeID, cert.TenantID, cert.CertificateSerial, cert.CertificateType, cert.Subject, cert.Issuer, cert.PublicKeyPEM, cert.Fingerprint, cert.DeviceID, cert.DeviceName, cert.IssuedAt, cert.ExpiresAt, cert.RevokedAt, cert.RevokeReason, cert.Status, cert.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to issue certificate: %w", err)
	}
	return cert, nil
}

// CreateFWOSEvent creates an event in the FWOS event log.
func (r *Phase5Repository) CreateFWOSEvent(ctx context.Context, ev *FWOSEvent) (*FWOSEvent, error) {
	ev.ID = uuid.New()
	ev.CreatedAt = time.Now()

	var payloadParam interface{}
	if ev.Payload != nil {
		b, _ := json.Marshal(ev.Payload)
		payloadParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO fwos_events (id, tenant_id, event_type, source, actor_id, resource_type, resource_id, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		ev.ID, ev.TenantID, ev.EventType, ev.Source, ev.ActorID, ev.ResourceType, ev.ResourceID, payloadParam, ev.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create FWOS event: %w", err)
	}
	return ev, nil
}
