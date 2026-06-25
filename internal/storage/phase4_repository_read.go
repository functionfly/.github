package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// GetTeamHealthMetrics lists team health metrics for a tenant.
func (r *Phase4Repository) GetTeamHealthMetrics(ctx context.Context, tenantID uuid.UUID, opts ListTeamHealthOpts) ([]*TeamHealthMetric, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.DepartmentID != nil {
		where += fmt.Sprintf(" AND department_id = $%d", argIdx)
		args = append(args, *opts.DepartmentID)
		argIdx++
	}
	if opts.StartDate != nil {
		where += fmt.Sprintf(" AND metric_date >= $%d", argIdx)
		args = append(args, *opts.StartDate)
		argIdx++
	}
	if opts.EndDate != nil {
		where += fmt.Sprintf(" AND metric_date <= $%d", argIdx)
		args = append(args, *opts.EndDate)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM team_health_metrics %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count team health metrics: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY metric_date DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, tenant_id, department_id, team_id, metric_date, workload_score, burnout_risk, velocity_score, collaboration_score, knowledge_sharing_score, pto_utilization_pct, avg_overtime_hours, headcount, metadata, created_at
		FROM team_health_metrics %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list team health metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*TeamHealthMetric
	for rows.Next() {
		m := &TeamHealthMetric{}
		var deptID sql.NullInt64
		var teamID sql.NullString
		var metaBytes []byte
		if err := rows.Scan(&m.ID, &m.TenantID, &deptID, &teamID, &m.MetricDate, &m.WorkloadScore, &m.BurnoutRisk, &m.VelocityScore, &m.CollaborationScore, &m.KnowledgeSharingScore, &m.PTOUtilizationPct, &m.AvgOvertimeHours, &m.Headcount, &metaBytes, &m.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan team health metric: %w", err)
		}
		if deptID.Valid {
			m.DepartmentID = &deptID.Int64
		}
		if teamID.Valid {
			tid, err := uuid.Parse(teamID.String)
			if err == nil {
				m.TeamID = &tid
			}
		}
		if metaBytes != nil {
			var meta JSONMap
			if err := json.Unmarshal(metaBytes, &meta); err == nil {
				m.Metadata = meta
			}
		}
		metrics = append(metrics, m)
	}
	return metrics, total, nil
}

// GetLatestTeamHealth retrieves the most recent team health metric for a department.
func (r *Phase4Repository) GetLatestTeamHealth(ctx context.Context, tenantID uuid.UUID, departmentID int64) (*TeamHealthMetric, error) {
	m := &TeamHealthMetric{}
	var deptID sql.NullInt64
	var teamID sql.NullString
	var metaBytes []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, department_id, team_id, metric_date, workload_score, burnout_risk, velocity_score, collaboration_score, knowledge_sharing_score, pto_utilization_pct, avg_overtime_hours, headcount, metadata, created_at
		FROM team_health_metrics WHERE tenant_id = $1 AND department_id = $2 ORDER BY metric_date DESC LIMIT 1`,
		tenantID, departmentID).Scan(
		&m.ID, &m.TenantID, &deptID, &teamID, &m.MetricDate, &m.WorkloadScore, &m.BurnoutRisk, &m.VelocityScore, &m.CollaborationScore, &m.KnowledgeSharingScore, &m.PTOUtilizationPct, &m.AvgOvertimeHours, &m.Headcount, &metaBytes, &m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest team health: %w", err)
	}
	if deptID.Valid {
		m.DepartmentID = &deptID.Int64
	}
	if teamID.Valid {
		tid, err := uuid.Parse(teamID.String)
		if err == nil {
			m.TeamID = &tid
		}
	}
	if metaBytes != nil {
		var meta JSONMap
		if err := json.Unmarshal(metaBytes, &meta); err == nil {
			m.Metadata = meta
		}
	}
	return m, nil
}

// GetSkillsGraph lists all skills in the graph for a tenant.
func (r *Phase4Repository) GetSkillsGraph(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*SkillsGraph, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM skills_graph WHERE tenant_id = $1", tenantID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count skills graph: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, skill_name, category, total_employees, avg_proficiency, demand_score, supply_score, gap_score, trending, last_calculated, created_at, updated_at
		FROM skills_graph WHERE tenant_id = $1 ORDER BY gap_score DESC NULLS LAST LIMIT $2 OFFSET $3`,
		tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list skills graph: %w", err)
	}
	defer rows.Close()

	var skills []*SkillsGraph
	for rows.Next() {
		s := &SkillsGraph{}
		var category sql.NullString
		var lastCalc sql.NullTime
		if err := rows.Scan(&s.ID, &s.TenantID, &s.SkillName, &category, &s.TotalEmployees, &s.AvgProficiency, &s.DemandScore, &s.SupplyScore, &s.GapScore, &s.Trending, &lastCalc, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan skills graph: %w", err)
		}
		if category.Valid {
			s.Category = &category.String
		}
		if lastCalc.Valid {
			s.LastCalculated = &lastCalc.Time
		}
		skills = append(skills, s)
	}
	return skills, total, nil
}

// GetSkillGaps returns skills with the highest gap scores.
func (r *Phase4Repository) GetSkillGaps(ctx context.Context, tenantID uuid.UUID, limit int) ([]*SkillsGraph, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, skill_name, category, total_employees, avg_proficiency, demand_score, supply_score, gap_score, trending, last_calculated, created_at, updated_at
		FROM skills_graph WHERE tenant_id = $1 AND gap_score > 0 ORDER BY gap_score DESC LIMIT $2`,
		tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get skill gaps: %w", err)
	}
	defer rows.Close()

	var skills []*SkillsGraph
	for rows.Next() {
		s := &SkillsGraph{}
		var category sql.NullString
		var lastCalc sql.NullTime
		if err := rows.Scan(&s.ID, &s.TenantID, &s.SkillName, &category, &s.TotalEmployees, &s.AvgProficiency, &s.DemandScore, &s.SupplyScore, &s.GapScore, &s.Trending, &lastCalc, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan skill gap: %w", err)
		}
		if category.Valid {
			s.Category = &category.String
		}
		if lastCalc.Valid {
			s.LastCalculated = &lastCalc.Time
		}
		skills = append(skills, s)
	}
	return skills, nil
}

// GetReputationScores retrieves reputation scores for an employee.
func (r *Phase4Repository) GetReputationScores(ctx context.Context, employeeID uuid.UUID) ([]*ReputationScore, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, tenant_id, category, score, rank, percentile, components, last_calculated, created_at, updated_at
		FROM reputation_scores WHERE employee_id = $1 ORDER BY score DESC`,
		employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reputation scores: %w", err)
	}
	defer rows.Close()

	var scores []*ReputationScore
	for rows.Next() {
		s := &ReputationScore{}
		var rank sql.NullInt64
		var percentile sql.NullFloat64
		var compBytes []byte
		var lastCalc sql.NullTime
		if err := rows.Scan(&s.ID, &s.EmployeeID, &s.TenantID, &s.Category, &s.Score, &rank, &percentile, &compBytes, &lastCalc, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan reputation score: %w", err)
		}
		if rank.Valid {
			r := int(rank.Int64)
			s.Rank = &r
		}
		if percentile.Valid {
			s.Percentile = &percentile.Float64
		}
		if compBytes != nil {
			var comp JSONMap
			if err := json.Unmarshal(compBytes, &comp); err == nil {
				s.Components = comp
			}
		}
		if lastCalc.Valid {
			s.LastCalculated = &lastCalc.Time
		}
		scores = append(scores, s)
	}
	return scores, nil
}

// GetReputationLeaderboard returns top employees by reputation category.
func (r *Phase4Repository) GetReputationLeaderboard(ctx context.Context, tenantID uuid.UUID, category string, limit, offset int) ([]*ReputationScore, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM reputation_scores WHERE tenant_id = $1 AND category = $2", tenantID, category).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count reputation leaderboard: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, tenant_id, category, score, rank, percentile, components, last_calculated, created_at, updated_at
		FROM reputation_scores WHERE tenant_id = $1 AND category = $2 ORDER BY score DESC LIMIT $3 OFFSET $4`,
		tenantID, category, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get reputation leaderboard: %w", err)
	}
	defer rows.Close()

	var scores []*ReputationScore
	for rows.Next() {
		s := &ReputationScore{}
		var rank sql.NullInt64
		var percentile sql.NullFloat64
		var compBytes []byte
		var lastCalc sql.NullTime
		if err := rows.Scan(&s.ID, &s.EmployeeID, &s.TenantID, &s.Category, &s.Score, &rank, &percentile, &compBytes, &lastCalc, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan reputation leaderboard: %w", err)
		}
		if rank.Valid {
			r := int(rank.Int64)
			s.Rank = &r
		}
		if percentile.Valid {
			s.Percentile = &percentile.Float64
		}
		if compBytes != nil {
			var comp JSONMap
			if err := json.Unmarshal(compBytes, &comp); err == nil {
				s.Components = comp
			}
		}
		if lastCalc.Valid {
			s.LastCalculated = &lastCalc.Time
		}
		scores = append(scores, s)
	}
	return scores, total, nil
}

// ListDigitalBadges lists digital badges for a tenant.
func (r *Phase4Repository) ListDigitalBadges(ctx context.Context, tenantID uuid.UUID, opts ListBadgesOpts) ([]*DigitalBadge, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.Category != nil {
		where += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, *opts.Category)
		argIdx++
	}
	if opts.IsActive != nil {
		where += fmt.Sprintf(" AND is_active = $%d", argIdx)
		args = append(args, *opts.IsActive)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM digital_badges %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count digital badges: %w", err)
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
		SELECT id, tenant_id, slug, name, description, icon_url, category, criteria, points, is_active, created_at
		FROM digital_badges %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list digital badges: %w", err)
	}
	defer rows.Close()

	var badges []*DigitalBadge
	for rows.Next() {
		b := &DigitalBadge{}
		var description, iconURL sql.NullString
		var critBytes []byte
		if err := rows.Scan(&b.ID, &b.TenantID, &b.Slug, &b.Name, &description, &iconURL, &b.Category, &critBytes, &b.Points, &b.IsActive, &b.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan digital badge: %w", err)
		}
		if description.Valid {
			b.Description = &description.String
		}
		if iconURL.Valid {
			b.IconURL = &iconURL.String
		}
		if critBytes != nil {
			var crit JSONMap
			if err := json.Unmarshal(critBytes, &crit); err == nil {
				b.Criteria = crit
			}
		}
		badges = append(badges, b)
	}
	return badges, total, nil
}

// GetDigitalBadgeBySlug retrieves a badge by slug.
func (r *Phase4Repository) GetDigitalBadgeBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*DigitalBadge, error) {
	b := &DigitalBadge{}
	var description, iconURL sql.NullString
	var critBytes []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, slug, name, description, icon_url, category, criteria, points, is_active, created_at
		FROM digital_badges WHERE tenant_id = $1 AND slug = $2`,
		tenantID, slug).Scan(
		&b.ID, &b.TenantID, &b.Slug, &b.Name, &description, &iconURL, &b.Category, &critBytes, &b.Points, &b.IsActive, &b.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get digital badge: %w", err)
	}
	if description.Valid {
		b.Description = &description.String
	}
	if iconURL.Valid {
		b.IconURL = &iconURL.String
	}
	if critBytes != nil {
		var crit JSONMap
		if err := json.Unmarshal(critBytes, &crit); err == nil {
			b.Criteria = crit
		}
	}
	return b, nil
}

// GetEmployeeBadges retrieves badges awarded to an employee.
func (r *Phase4Repository) GetEmployeeBadges(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeBadge, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, badge_id, awarded_by, awarded_at
		FROM employee_badges WHERE employee_id = $1 ORDER BY awarded_at DESC`,
		employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee badges: %w", err)
	}
	defer rows.Close()

	var badges []*EmployeeBadge
	for rows.Next() {
		eb := &EmployeeBadge{}
		var awardedBy sql.NullString
		if err := rows.Scan(&eb.ID, &eb.EmployeeID, &eb.BadgeID, &awardedBy, &eb.AwardedAt); err != nil {
			return nil, fmt.Errorf("failed to scan employee badge: %w", err)
		}
		if awardedBy.Valid {
			aid, err := uuid.Parse(awardedBy.String)
			if err == nil {
				eb.AwardedBy = &aid
			}
		}
		badges = append(badges, eb)
	}
	return badges, nil
}

// GetLivingMemoryEntry retrieves a living memory entry by ID.
func (r *Phase4Repository) GetLivingMemoryEntry(ctx context.Context, id uuid.UUID) (*LivingMemoryEntry, error) {
	e := &LivingMemoryEntry{}
	var projectID sql.NullString
	var tagsBytes, participantsBytes []byte
	var searchableText sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, author_id, title, body, memory_type, project_id, tags, participants, importance, searchable_text, view_count, created_at, updated_at
		FROM living_memory WHERE id = $1`, id).Scan(
		&e.ID, &e.TenantID, &e.AuthorID, &e.Title, &e.Body, &e.MemoryType, &projectID, &tagsBytes, &participantsBytes, &e.Importance, &searchableText, &e.ViewCount, &e.CreatedAt, &e.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get living memory entry: %w", err)
	}
	if projectID.Valid {
		pid, err := uuid.Parse(projectID.String)
		if err == nil {
			e.ProjectID = &pid
		}
	}
	if tagsBytes != nil {
		var tags JSONMap
		if err := json.Unmarshal(tagsBytes, &tags); err == nil {
			e.Tags = tags
		}
	}
	if participantsBytes != nil {
		var participants JSONMap
		if err := json.Unmarshal(participantsBytes, &participants); err == nil {
			e.Participants = participants
		}
	}
	if searchableText.Valid {
		e.SearchableText = &searchableText.String
	}
	return e, nil
}

// ListLivingMemoryEntries lists living memory entries for a tenant.
func (r *Phase4Repository) ListLivingMemoryEntries(ctx context.Context, tenantID uuid.UUID, opts ListLivingMemoryOpts) ([]*LivingMemoryEntry, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.MemoryType != nil {
		where += fmt.Sprintf(" AND memory_type = $%d", argIdx)
		args = append(args, *opts.MemoryType)
		argIdx++
	}
	if opts.ProjectID != nil {
		where += fmt.Sprintf(" AND project_id = $%d", argIdx)
		args = append(args, *opts.ProjectID)
		argIdx++
	}
	if opts.Importance != nil {
		where += fmt.Sprintf(" AND importance = $%d", argIdx)
		args = append(args, *opts.Importance)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM living_memory %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count living memory entries: %w", err)
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
		SELECT id, tenant_id, author_id, title, body, memory_type, project_id, tags, participants, importance, searchable_text, view_count, created_at, updated_at
		FROM living_memory %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list living memory entries: %w", err)
	}
	defer rows.Close()

	var entries []*LivingMemoryEntry
	for rows.Next() {
		e := &LivingMemoryEntry{}
		var projectID sql.NullString
		var tagsBytes, participantsBytes []byte
		var searchableText sql.NullString
		if err := rows.Scan(&e.ID, &e.TenantID, &e.AuthorID, &e.Title, &e.Body, &e.MemoryType, &projectID, &tagsBytes, &participantsBytes, &e.Importance, &searchableText, &e.ViewCount, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan living memory entry: %w", err)
		}
		if projectID.Valid {
			pid, err := uuid.Parse(projectID.String)
			if err == nil {
				e.ProjectID = &pid
			}
		}
		if tagsBytes != nil {
			var tags JSONMap
			if err := json.Unmarshal(tagsBytes, &tags); err == nil {
				e.Tags = tags
			}
		}
		if participantsBytes != nil {
			var participants JSONMap
			if err := json.Unmarshal(participantsBytes, &participants); err == nil {
				e.Participants = participants
			}
		}
		if searchableText.Valid {
			e.SearchableText = &searchableText.String
		}
		entries = append(entries, e)
	}
	return entries, total, nil
}

// SearchLivingMemory performs full-text search on living memory entries.
func (r *Phase4Repository) SearchLivingMemory(ctx context.Context, tenantID uuid.UUID, opts SearchLivingMemoryOpts) ([]*LivingMemoryEntry, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	where := "WHERE tenant_id = $1 AND searchable_text @@ plainto_tsquery('english', $2)"
	args := []interface{}{tenantID, opts.Query}
	argIdx := 3

	if opts.MemoryType != nil {
		where += fmt.Sprintf(" AND memory_type = $%d", argIdx)
		args = append(args, *opts.MemoryType)
		argIdx++
	}
	if opts.ProjectID != nil {
		where += fmt.Sprintf(" AND project_id = $%d", argIdx)
		args = append(args, *opts.ProjectID)
		argIdx++
	}

	args = append(args, limit)
	query := fmt.Sprintf(`
		SELECT id, tenant_id, author_id, title, body, memory_type, project_id, tags, participants, importance, searchable_text, view_count, created_at, updated_at
		FROM living_memory %s
		ORDER BY ts_rank(to_tsvector('english', searchable_text), plainto_tsquery('english', $2)) DESC
		LIMIT $%d`, where, argIdx)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search living memory: %w", err)
	}
	defer rows.Close()

	var entries []*LivingMemoryEntry
	for rows.Next() {
		e := &LivingMemoryEntry{}
		var projectID sql.NullString
		var tagsBytes, participantsBytes []byte
		var searchableText sql.NullString
		if err := rows.Scan(&e.ID, &e.TenantID, &e.AuthorID, &e.Title, &e.Body, &e.MemoryType, &projectID, &tagsBytes, &participantsBytes, &e.Importance, &searchableText, &e.ViewCount, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		if projectID.Valid {
			pid, err := uuid.Parse(projectID.String)
			if err == nil {
				e.ProjectID = &pid
			}
		}
		if tagsBytes != nil {
			var tags JSONMap
			if err := json.Unmarshal(tagsBytes, &tags); err == nil {
				e.Tags = tags
			}
		}
		if participantsBytes != nil {
			var participants JSONMap
			if err := json.Unmarshal(participantsBytes, &participants); err == nil {
				e.Participants = participants
			}
		}
		if searchableText.Valid {
			e.SearchableText = &searchableText.String
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// GetLatestMissionControlSnapshot retrieves the most recent snapshot for a tenant.
func (r *Phase4Repository) GetLatestMissionControlSnapshot(ctx context.Context, tenantID uuid.UUID) (*MissionControlSnapshot, error) {
	s := &MissionControlSnapshot{}
	var metaBytes []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, snapshot_date, total_employees, active_employees, new_hires_30d, departures_30d, total_projects, active_projects, completed_projects_30d, total_tasks, completed_tasks_30d, avg_task_completion_days, total_learning_hours, avg_skill_proficiency, innovation_grants_submitted, innovation_grants_funded, pto_days_used_30d, avg_burnout_risk, metadata, created_at
		FROM mission_control_snapshots WHERE tenant_id = $1 ORDER BY snapshot_date DESC LIMIT 1`,
		tenantID).Scan(
		&s.ID, &s.TenantID, &s.SnapshotDate, &s.TotalEmployees, &s.ActiveEmployees, &s.NewHires30d, &s.Departures30d, &s.TotalProjects, &s.ActiveProjects, &s.CompletedProjects30d, &s.TotalTasks, &s.CompletedTasks30d, &s.AvgTaskCompletionDays, &s.TotalLearningHours, &s.AvgSkillProficiency, &s.InnovationGrantsSubmitted, &s.InnovationGrantsFunded, &s.PTODaysUsed30d, &s.AvgBurnoutRisk, &metaBytes, &s.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get mission control snapshot: %w", err)
	}
	if metaBytes != nil {
		var meta JSONMap
		if err := json.Unmarshal(metaBytes, &meta); err == nil {
			s.Metadata = meta
		}
	}
	return s, nil
}

// ListMissionControlSnapshots lists snapshots for a tenant.
func (r *Phase4Repository) ListMissionControlSnapshots(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*MissionControlSnapshot, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM mission_control_snapshots WHERE tenant_id = $1", tenantID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count mission control snapshots: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, snapshot_date, total_employees, active_employees, new_hires_30d, departures_30d, total_projects, active_projects, completed_projects_30d, total_tasks, completed_tasks_30d, avg_task_completion_days, total_learning_hours, avg_skill_proficiency, innovation_grants_submitted, innovation_grants_funded, pto_days_used_30d, avg_burnout_risk, metadata, created_at
		FROM mission_control_snapshots WHERE tenant_id = $1 ORDER BY snapshot_date DESC LIMIT $2 OFFSET $3`,
		tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list mission control snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []*MissionControlSnapshot
	for rows.Next() {
		s := &MissionControlSnapshot{}
		var metaBytes []byte
		if err := rows.Scan(&s.ID, &s.TenantID, &s.SnapshotDate, &s.TotalEmployees, &s.ActiveEmployees, &s.NewHires30d, &s.Departures30d, &s.TotalProjects, &s.ActiveProjects, &s.CompletedProjects30d, &s.TotalTasks, &s.CompletedTasks30d, &s.AvgTaskCompletionDays, &s.TotalLearningHours, &s.AvgSkillProficiency, &s.InnovationGrantsSubmitted, &s.InnovationGrantsFunded, &s.PTODaysUsed30d, &s.AvgBurnoutRisk, &metaBytes, &s.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan mission control snapshot: %w", err)
		}
		if metaBytes != nil {
			var meta JSONMap
			if err := json.Unmarshal(metaBytes, &meta); err == nil {
				s.Metadata = meta
			}
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, total, nil
}
