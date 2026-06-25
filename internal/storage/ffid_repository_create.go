package storage

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
)

// GenerateFFID generates a new FFID in the format FF-{year}-{random4}-{sequence5}.
func (r *FFIDRepository) GenerateFFID(ctx context.Context, tenantID uuid.UUID) (string, error) {
	year := time.Now().Year() % 100
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	random4 := make([]byte, 4)
	for i := range random4 {
		random4[i] = chars[rand.Intn(len(chars))]
	}

	var seq int
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(CAST(SPLIT_PART(ffid, '-', 4) AS INTEGER)), 0) + 1
		FROM employees
		WHERE ffid LIKE $1
	`, fmt.Sprintf("FF-%02d-%s-%%", year, string(random4))).Scan(&seq)
	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to generate FFID sequence: %w", err)
	}
	if seq == 0 {
		seq = 1
	}

	return fmt.Sprintf("FF-%02d-%s-%05d", year, string(random4), seq), nil
}

// GetIdentityCard returns the full identity card data for an employee.
func (r *FFIDRepository) GetIdentityCard(ctx context.Context, employeeID uuid.UUID) (*types.IdentityCard, error) {
	emp, err := r.getEmployeeByID(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	if emp == nil {
		return nil, nil
	}

	card := &types.IdentityCard{
		Employee:               emp,
		ClearanceLevel:         types.ClearanceLevelFromText(emp.ClearanceLevel),
		IdentitySignature:      "",
		ReputationTotal:        0,
		TrustScore:             0,
		Achievements:           []*types.AchievementProgress{},
		Timeline:               []*types.CareerTimelineEvent{},
		ReputationByCategory:   make(map[string]float64),
	}

	_ = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(clearance_level_num, 1), COALESCE(identity_signature, ''),
			   COALESCE(reputation_total, 0), COALESCE(trust_score, 0)
		FROM employees WHERE id = $1
	`, employeeID).Scan(&card.ClearanceLevel, &card.IdentitySignature,
		&card.ReputationTotal, &card.TrustScore)

	achievements, err := r.GetAchievementProgress(ctx, employeeID)
	if err == nil {
		card.Achievements = achievements
	}

	timeline, err := r.GetCareerTimeline(ctx, employeeID)
	if err == nil {
		card.Timeline = timeline
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT category, score FROM reputation_scores WHERE employee_id = $1
	`, employeeID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cat string
			var score float64
			if err := rows.Scan(&cat, &score); err == nil {
				card.ReputationByCategory[cat] = score
			}
		}
	}

	return card, nil
}

func (r *FFIDRepository) getEmployeeByID(ctx context.Context, id uuid.UUID) (*types.Employee, error) {
	var emp types.Employee
	var deptID sql.NullInt64
	var managerID, workLoc, officeLoc, tz, bio, pronouns sql.NullString
	var hireDate sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, tenant_id, employee_number, ffid, department_id,
			   manager_id, hire_date, employment_type, clearance_level,
			   work_location, office_location, timezone, bio, pronouns,
			   emergency_contact, status, created_at, updated_at
		FROM employees WHERE id = $1
	`, id).Scan(
		&emp.ID, &emp.UserID, &emp.TenantID, &emp.EmployeeNumber, &emp.FFID,
		&deptID, &managerID, &hireDate, &emp.EmploymentType, &emp.ClearanceLevel,
		&workLoc, &officeLoc, &tz, &bio, &pronouns,
		&emp.EmergencyContact, &emp.Status, &emp.CreatedAt, &emp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	if deptID.Valid {
		d := deptID.Int64
		emp.DepartmentID = &d
	}
	if managerID.Valid {
		mid, err := uuid.Parse(managerID.String)
		if err == nil {
			emp.ManagerID = &mid
		}
	}
	if hireDate.Valid {
		emp.HireDate = &hireDate.Time
	}
	if workLoc.Valid {
		emp.WorkLocation = &workLoc.String
	}
	if officeLoc.Valid {
		emp.OfficeLocation = &officeLoc.String
	}
	if tz.Valid {
		emp.Timezone = &tz.String
	}
	if bio.Valid {
		emp.Bio = &bio.String
	}
	if pronouns.Valid {
		emp.Pronouns = &pronouns.String
	}

	return &emp, nil
}

// CreateAchievementDefinition creates a new achievement definition.
func (r *FFIDRepository) CreateAchievementDefinition(ctx context.Context, ach *types.AchievementDefinition) (*types.AchievementDefinition, error) {
	ach.ID = uuid.New()
	ach.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO achievement_definitions (id, tenant_id, slug, name, description, icon, category, criteria_type, criteria_threshold, points, tier, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, ach.ID, ach.TenantID, ach.Slug, ach.Name, ach.Description, ach.Icon,
		ach.Category, ach.CriteriaType, ach.CriteriaThreshold, ach.Points,
		ach.Tier, ach.IsActive, ach.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create achievement definition: %w", err)
	}
	return ach, nil
}

// ListAchievementDefinitions lists all achievement definitions for a tenant.
func (r *FFIDRepository) ListAchievementDefinitions(ctx context.Context, tenantID uuid.UUID) ([]*types.AchievementDefinition, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, slug, name, description, icon, category,
			   criteria_type, criteria_threshold, points, tier, is_active, created_at
		FROM achievement_definitions
		WHERE tenant_id = $1
		ORDER BY tier, points
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list achievement definitions: %w", err)
	}
	defer rows.Close()

	var results []*types.AchievementDefinition
	for rows.Next() {
		var a types.AchievementDefinition
		if err := rows.Scan(&a.ID, &a.TenantID, &a.Slug, &a.Name, &a.Description,
			&a.Icon, &a.Category, &a.CriteriaType, &a.CriteriaThreshold,
			&a.Points, &a.Tier, &a.IsActive, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan achievement definition: %w", err)
		}
		results = append(results, &a)
	}
	return results, nil
}

// CheckAndAwardAchievements checks employee progress and awards unlocked achievements.
func (r *FFIDRepository) CheckAndAwardAchievements(ctx context.Context, employeeID uuid.UUID) error {
	defs, err := r.db.QueryContext(ctx, `
		SELECT ad.id, ad.criteria_type, ad.criteria_threshold, ad.points
		FROM achievement_definitions ad
		JOIN employees e ON e.tenant_id = ad.tenant_id
		WHERE e.id = $1 AND ad.is_active = TRUE
	`, employeeID)
	if err != nil {
		return fmt.Errorf("failed to query achievement definitions: %w", err)
	}
	defer defs.Close()

	for defs.Next() {
		var achID uuid.UUID
		var criteriaType string
		var threshold, points int
		if err := defs.Scan(&achID, &criteriaType, &threshold, &points); err != nil {
			continue
		}

		currentValue, err := r.getCriteriaValue(ctx, employeeID, criteriaType)
		if err != nil {
			continue
		}

		_, err = r.db.ExecContext(ctx, `
			INSERT INTO achievement_progress (employee_id, achievement_id, current_value, awarded, awarded_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			ON CONFLICT (employee_id, achievement_id) DO UPDATE SET
				current_value = EXCLUDED.current_value,
				awarded = CASE WHEN EXCLUDED.current_value >= $6 THEN TRUE ELSE achievement_progress.awarded END,
				awarded_at = CASE WHEN EXCLUDED.current_value >= $6 AND NOT achievement_progress.awarded THEN NOW() ELSE achievement_progress.awarded_at END,
				updated_at = NOW()
		`, employeeID, achID, currentValue, currentValue >= threshold,
			func() interface{} {
				if currentValue >= threshold {
					return time.Now()
				}
				return nil
			}(), threshold)
		if err != nil {
			continue
		}
	}
	return nil
}

func (r *FFIDRepository) getCriteriaValue(ctx context.Context, employeeID uuid.UUID, criteriaType string) (int, error) {
	var count int
	var err error

	switch criteriaType {
	case "tasks_completed":
		err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tasks WHERE assignee_id = $1 AND status = 'done'`, employeeID).Scan(&count)
	case "projects_shipped":
		err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM projects WHERE owner_id = $1 AND status = 'completed'`, employeeID).Scan(&count)
	case "incidents_resolved":
		err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM fwos_incidents WHERE commander_id = $1 AND status = 'resolved'`, employeeID).Scan(&count)
	case "mentorships":
		err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM mentorship_matches WHERE mentor_id = $1 AND status = 'active'`, employeeID).Scan(&count)
	case "grants_funded":
		err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM innovation_grants WHERE proposer_id = $1 AND status = 'funded'`, employeeID).Scan(&count)
	case "articles_published":
		err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM knowledge_articles WHERE author_id = $1 AND status = 'published'`, employeeID).Scan(&count)
	case "peer_feedbacks":
		err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM peer_feedback WHERE to_employee_id = $1`, employeeID).Scan(&count)
	default:
		count = 0
	}

	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetAchievementProgress returns achievement progress for an employee.
func (r *FFIDRepository) GetAchievementProgress(ctx context.Context, employeeID uuid.UUID) ([]*types.AchievementProgress, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ap.id, ap.employee_id, ap.achievement_id, ap.current_value,
			   ap.awarded, ap.awarded_at, ap.created_at, ap.updated_at
		FROM achievement_progress ap
		WHERE ap.employee_id = $1
		ORDER BY ap.updated_at DESC
	`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievement progress: %w", err)
	}
	defer rows.Close()

	var results []*types.AchievementProgress
	for rows.Next() {
		var p types.AchievementProgress
		var awardedAt sql.NullTime
		if err := rows.Scan(&p.ID, &p.EmployeeID, &p.AchievementID,
			&p.CurrentValue, &p.Awarded, &awardedAt,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan achievement progress: %w", err)
		}
		if awardedAt.Valid {
			p.AwardedAt = &awardedAt.Time
		}
		results = append(results, &p)
	}
	return results, nil
}

// CreateCareerTimelineEvent creates a new career timeline event.
func (r *FFIDRepository) CreateCareerTimelineEvent(ctx context.Context, ev *types.CareerTimelineEvent) (*types.CareerTimelineEvent, error) {
	ev.ID = uuid.New()
	ev.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO career_timeline_events (id, employee_id, tenant_id, event_type, title, description, metadata, event_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, ev.ID, ev.EmployeeID, ev.TenantID, ev.EventType, ev.Title,
		ev.Description, ev.Metadata, ev.EventDate, ev.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create career timeline event: %w", err)
	}
	return ev, nil
}

// GetCareerTimeline returns career timeline events for an employee.
func (r *FFIDRepository) GetCareerTimeline(ctx context.Context, employeeID uuid.UUID) ([]*types.CareerTimelineEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, tenant_id, event_type, title, description, metadata, event_date, created_at
		FROM career_timeline_events
		WHERE employee_id = $1
		ORDER BY event_date DESC
	`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get career timeline: %w", err)
	}
	defer rows.Close()

	var results []*types.CareerTimelineEvent
	for rows.Next() {
		var ev types.CareerTimelineEvent
		var desc sql.NullString
		if err := rows.Scan(&ev.ID, &ev.EmployeeID, &ev.TenantID, &ev.EventType,
			&ev.Title, &desc, &ev.Metadata, &ev.EventDate, &ev.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan career timeline event: %w", err)
		}
		if desc.Valid {
			ev.Description = &desc.String
		}
		results = append(results, &ev)
	}
	return results, nil
}

// RecordReputationSnapshot records a reputation history entry.
func (r *FFIDRepository) RecordReputationSnapshot(ctx context.Context, employeeID, tenantID uuid.UUID, category string, score float64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO reputation_history (employee_id, tenant_id, category, score, recorded_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, employeeID, tenantID, category, score)
	if err != nil {
		return fmt.Errorf("failed to record reputation snapshot: %w", err)
	}
	return nil
}

// GetReputationHistory returns reputation history for an employee in a category.
func (r *FFIDRepository) GetReputationHistory(ctx context.Context, employeeID uuid.UUID, category string) ([]*types.ReputationHistory, error) {
	query := `
		SELECT id, employee_id, tenant_id, category, score, recorded_at
		FROM reputation_history
		WHERE employee_id = $1
	`
	args := []interface{}{employeeID}

	if category != "" {
		query += " AND category = $2"
		args = append(args, category)
	}
	query += " ORDER BY recorded_at DESC LIMIT 100"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get reputation history: %w", err)
	}
	defer rows.Close()

	var results []*types.ReputationHistory
	for rows.Next() {
		var h types.ReputationHistory
		if err := rows.Scan(&h.ID, &h.EmployeeID, &h.TenantID,
			&h.Category, &h.Score, &h.RecordedAt); err != nil {
			return nil, fmt.Errorf("failed to scan reputation history: %w", err)
		}
		results = append(results, &h)
	}
	return results, nil
}

// UpdateClearanceLevel updates the numeric clearance level for an employee.
func (r *FFIDRepository) UpdateClearanceLevel(ctx context.Context, employeeID uuid.UUID, level int) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE employees SET clearance_level_num = $1, updated_at = NOW()
		WHERE id = $2
	`, level, employeeID)
	if err != nil {
		return fmt.Errorf("failed to update clearance level: %w", err)
	}
	return nil
}

// UpdateIdentitySignature updates the identity signature for an employee.
func (r *FFIDRepository) UpdateIdentitySignature(ctx context.Context, employeeID uuid.UUID, signature string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE employees SET identity_signature = $1, updated_at = NOW()
		WHERE id = $2
	`, signature, employeeID)
	if err != nil {
		return fmt.Errorf("failed to update identity signature: %w", err)
	}
	return nil
}
