package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateTeamHealthMetric creates a new team health metric.
func (r *Phase4Repository) CreateTeamHealthMetric(ctx context.Context, m *TeamHealthMetric) (*TeamHealthMetric, error) {
	m.ID = uuid.New()
	m.CreatedAt = time.Now()

	var metaParam interface{}
	if m.Metadata != nil {
		b, _ := json.Marshal(m.Metadata)
		metaParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO team_health_metrics (id, tenant_id, department_id, team_id, metric_date, workload_score, burnout_risk, velocity_score, collaboration_score, knowledge_sharing_score, pto_utilization_pct, avg_overtime_hours, headcount, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		m.ID, m.TenantID, m.DepartmentID, m.TeamID, m.MetricDate, m.WorkloadScore, m.BurnoutRisk, m.VelocityScore, m.CollaborationScore, m.KnowledgeSharingScore, m.PTOUtilizationPct, m.AvgOvertimeHours, m.Headcount, metaParam, m.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create team health metric: %w", err)
	}
	return m, nil
}

// CreateReputationScore creates a new reputation score.
func (r *Phase4Repository) CreateReputationScore(ctx context.Context, s *ReputationScore) (*ReputationScore, error) {
	s.ID = uuid.New()
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()

	var compParam interface{}
	if s.Components != nil {
		b, _ := json.Marshal(s.Components)
		compParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO reputation_scores (id, employee_id, tenant_id, category, score, rank, percentile, components, last_calculated, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		s.ID, s.EmployeeID, s.TenantID, s.Category, s.Score, s.Rank, s.Percentile, compParam, s.LastCalculated, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create reputation score: %w", err)
	}
	return s, nil
}

// CreateDigitalBadge creates a new digital badge.
func (r *Phase4Repository) CreateDigitalBadge(ctx context.Context, b *DigitalBadge) (*DigitalBadge, error) {
	b.ID = uuid.New()
	b.CreatedAt = time.Now()

	var critParam interface{}
	if b.Criteria != nil {
		c, _ := json.Marshal(b.Criteria)
		critParam = c
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO digital_badges (id, tenant_id, slug, name, description, icon_url, category, criteria, points, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		b.ID, b.TenantID, b.Slug, b.Name, b.Description, b.IconURL, b.Category, critParam, b.Points, b.IsActive, b.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create digital badge: %w", err)
	}
	return b, nil
}

// AwardEmployeeBadge awards a badge to an employee.
func (r *Phase4Repository) AwardEmployeeBadge(ctx context.Context, eb *EmployeeBadge) (*EmployeeBadge, error) {
	eb.AwardedAt = time.Now()

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO employee_badges (employee_id, badge_id, awarded_by, awarded_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		eb.EmployeeID, eb.BadgeID, eb.AwardedBy, eb.AwardedAt,
	).Scan(&eb.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to award employee badge: %w", err)
	}
	return eb, nil
}

// CreateLivingMemoryEntry creates a new living memory entry.
func (r *Phase4Repository) CreateLivingMemoryEntry(ctx context.Context, e *LivingMemoryEntry) (*LivingMemoryEntry, error) {
	e.ID = uuid.New()
	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()

	var tagsParam, participantsParam interface{}
	if e.Tags != nil {
		b, _ := json.Marshal(e.Tags)
		tagsParam = b
	}
	if e.Participants != nil {
		b, _ := json.Marshal(e.Participants)
		participantsParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO living_memory (id, tenant_id, author_id, title, body, memory_type, project_id, tags, participants, importance, searchable_text, view_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		e.ID, e.TenantID, e.AuthorID, e.Title, e.Body, e.MemoryType, e.ProjectID, tagsParam, participantsParam, e.Importance, e.SearchableText, e.ViewCount, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create living memory entry: %w", err)
	}
	return e, nil
}

// CreateMissionControlSnapshot creates a new mission control snapshot.
func (r *Phase4Repository) CreateMissionControlSnapshot(ctx context.Context, s *MissionControlSnapshot) (*MissionControlSnapshot, error) {
	s.ID = uuid.New()
	s.CreatedAt = time.Now()

	var metaParam interface{}
	if s.Metadata != nil {
		b, _ := json.Marshal(s.Metadata)
		metaParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO mission_control_snapshots (id, tenant_id, snapshot_date, total_employees, active_employees, new_hires_30d, departures_30d, total_projects, active_projects, completed_projects_30d, total_tasks, completed_tasks_30d, avg_task_completion_days, total_learning_hours, avg_skill_proficiency, innovation_grants_submitted, innovation_grants_funded, pto_days_used_30d, avg_burnout_risk, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`,
		s.ID, s.TenantID, s.SnapshotDate, s.TotalEmployees, s.ActiveEmployees, s.NewHires30d, s.Departures30d, s.TotalProjects, s.ActiveProjects, s.CompletedProjects30d, s.TotalTasks, s.CompletedTasks30d, s.AvgTaskCompletionDays, s.TotalLearningHours, s.AvgSkillProficiency, s.InnovationGrantsSubmitted, s.InnovationGrantsFunded, s.PTODaysUsed30d, s.AvgBurnoutRisk, metaParam, s.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create mission control snapshot: %w", err)
	}
	return s, nil
}

// CreateSkillsGraphEntry creates or updates a skills graph entry.
func (r *Phase4Repository) CreateSkillsGraphEntry(ctx context.Context, s *SkillsGraph) (*SkillsGraph, error) {
	s.ID = uuid.New()
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO skills_graph (id, tenant_id, skill_name, category, total_employees, avg_proficiency, demand_score, supply_score, gap_score, trending, last_calculated, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (tenant_id, skill_name) DO UPDATE SET
			category = EXCLUDED.category,
			total_employees = EXCLUDED.total_employees,
			avg_proficiency = EXCLUDED.avg_proficiency,
			demand_score = EXCLUDED.demand_score,
			supply_score = EXCLUDED.supply_score,
			gap_score = EXCLUDED.gap_score,
			trending = EXCLUDED.trending,
			last_calculated = EXCLUDED.last_calculated,
			updated_at = NOW()`,
		s.ID, s.TenantID, s.SkillName, s.Category, s.TotalEmployees, s.AvgProficiency, s.DemandScore, s.SupplyScore, s.GapScore, s.Trending, s.LastCalculated, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create skills graph entry: %w", err)
	}
	return s, nil
}
