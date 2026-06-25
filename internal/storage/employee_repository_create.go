package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/functionfly/functionfly/internal/types"
)

// CreateDepartment creates a new department.
func (r *EmployeeRepository) CreateDepartment(ctx context.Context, dept *types.Department) (*types.Department, error) {
	dept.CreatedAt = time.Now()
	dept.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO departments (tenant_id, name, slug, description, parent_id, head_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		dept.TenantID, dept.Name, dept.Slug, dept.Description, dept.ParentID, dept.HeadID, dept.IsActive, dept.CreatedAt, dept.UpdatedAt,
	).Scan(&dept.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create department: %w", err)
	}
	return dept, nil
}

// CreateEmployee creates a new employee record.
func (r *EmployeeRepository) CreateEmployee(ctx context.Context, emp *types.Employee) (*types.Employee, error) {
	emp.ID = uuid.New()
	emp.CreatedAt = time.Now()
	emp.UpdatedAt = time.Now()

	var emergencyContactParam interface{}
	if emp.EmergencyContact != nil {
		b, _ := json.Marshal(emp.EmergencyContact)
		emergencyContactParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO employees (id, user_id, tenant_id, employee_number, ffid, department_id, manager_id, hire_date, employment_type, clearance_level, work_location, office_location, timezone, bio, pronouns, emergency_contact, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)`,
		emp.ID, emp.UserID, emp.TenantID, emp.EmployeeNumber, emp.FFID, emp.DepartmentID, emp.ManagerID, emp.HireDate, emp.EmploymentType, emp.ClearanceLevel, emp.WorkLocation, emp.OfficeLocation, emp.Timezone, emp.Bio, emp.Pronouns, emergencyContactParam, emp.Status, emp.CreatedAt, emp.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}
	return emp, nil
}

// AddEmployeeDepartment adds an employee-department membership.
func (r *EmployeeRepository) AddEmployeeDepartment(ctx context.Context, ed *types.EmployeeDepartment) error {
	ed.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO employee_departments (employee_id, department_id, role_in_dept, is_primary, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		ed.EmployeeID, ed.DepartmentID, ed.RoleInDept, ed.IsPrimary, ed.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add employee department: %w", err)
	}
	return nil
}

// AddEmployeeSkill adds a skill for an employee.
func (r *EmployeeRepository) AddEmployeeSkill(ctx context.Context, skill *types.EmployeeSkill) (*types.EmployeeSkill, error) {
	skill.CreatedAt = time.Now()
	skill.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO employee_skills (employee_id, skill_name, category, proficiency, years_exp, endorsements, verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		skill.EmployeeID, skill.SkillName, skill.Category, skill.Proficiency, skill.YearsExp, skill.Endorsements, skill.Verified, skill.CreatedAt, skill.UpdatedAt,
	).Scan(&skill.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to add employee skill: %w", err)
	}
	return skill, nil
}

// AddEmployeeCertification adds a certification for an employee.
func (r *EmployeeRepository) AddEmployeeCertification(ctx context.Context, cert *types.EmployeeCertification) (*types.EmployeeCertification, error) {
	cert.CreatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO employee_certifications (employee_id, name, issuer, credential_id, credential_url, issued_date, expiry_date, verified, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		cert.EmployeeID, cert.Name, cert.Issuer, cert.CredentialID, cert.CredentialURL, cert.IssuedDate, cert.ExpiryDate, cert.Verified, cert.CreatedAt,
	).Scan(&cert.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to add employee certification: %w", err)
	}
	return cert, nil
}

// AwardEmployeeAchievement awards an achievement to an employee.
func (r *EmployeeRepository) AwardEmployeeAchievement(ctx context.Context, ach *types.EmployeeAchievement) (*types.EmployeeAchievement, error) {
	ach.AwardedAt = time.Now()

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO employee_achievements (employee_id, title, description, type, awarded_by, points, badge_url, awarded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		ach.EmployeeID, ach.Title, ach.Description, ach.Type, ach.AwardedBy, ach.Points, ach.BadgeURL, ach.AwardedAt,
	).Scan(&ach.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to award achievement: %w", err)
	}
	return ach, nil
}

// CreateProject creates a new project.
func (r *EmployeeRepository) CreateProject(ctx context.Context, proj *types.Project) (*types.Project, error) {
	proj.ID = uuid.New()
	proj.CreatedAt = time.Now()
	proj.UpdatedAt = time.Now()

	var tagsParam, metadataParam interface{}
	if proj.Tags != nil {
		b, _ := json.Marshal(proj.Tags)
		tagsParam = b
	}
	if proj.Metadata != nil {
		b, _ := json.Marshal(proj.Metadata)
		metadataParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO projects (id, tenant_id, name, slug, description, status, priority, owner_id, start_date, target_date, tags, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		proj.ID, proj.TenantID, proj.Name, proj.Slug, proj.Description, proj.Status, proj.Priority, proj.OwnerID, proj.StartDate, proj.TargetDate, tagsParam, metadataParam, proj.CreatedAt, proj.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	return proj, nil
}

// CreateTask creates a new task.
func (r *EmployeeRepository) CreateTask(ctx context.Context, task *types.Task) (*types.Task, error) {
	task.ID = uuid.New()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	var tagsParam interface{}
	if task.Tags != nil {
		b, _ := json.Marshal(task.Tags)
		tagsParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (id, project_id, tenant_id, title, description, status, priority, assignee_id, reporter_id, parent_id, due_date, estimated_hours, actual_hours, tags, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
		task.ID, task.ProjectID, task.TenantID, task.Title, task.Description, task.Status, task.Priority, task.AssigneeID, task.ReporterID, task.ParentID, task.DueDate, task.EstimatedHours, task.ActualHours, tagsParam, task.Position, task.CreatedAt, task.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	return task, nil
}

// CreateTaskComment creates a comment on a task.
func (r *EmployeeRepository) CreateTaskComment(ctx context.Context, comment *types.TaskComment) (*types.TaskComment, error) {
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO task_comments (task_id, author_id, body, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		comment.TaskID, comment.AuthorID, comment.Body, comment.CreatedAt, comment.UpdatedAt,
	).Scan(&comment.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create task comment: %w", err)
	}
	return comment, nil
}

// CreateLearningCourse creates a new learning course.
func (r *EmployeeRepository) CreateLearningCourse(ctx context.Context, course *types.LearningCourse) (*types.LearningCourse, error) {
	course.ID = uuid.New()
	course.CreatedAt = time.Now()
	course.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO learning_courses (id, tenant_id, title, description, category, difficulty, duration_min, content_url, thumbnail_url, is_mandatory, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		course.ID, course.TenantID, course.Title, course.Description, course.Category, course.Difficulty, course.DurationMin, course.ContentURL, course.ThumbnailURL, course.IsMandatory, course.IsActive, course.CreatedAt, course.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create learning course: %w", err)
	}
	return course, nil
}

// EnrollCourse enrolls an employee in a learning course.
func (r *EmployeeRepository) EnrollCourse(ctx context.Context, el *types.EmployeeLearning) (*types.EmployeeLearning, error) {
	el.CreatedAt = time.Now()
	el.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO employee_learning (employee_id, course_id, status, progress_pct, started_at, completed_at, score, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		el.EmployeeID, el.CourseID, el.Status, el.ProgressPct, el.StartedAt, el.CompletedAt, el.Score, el.CreatedAt, el.UpdatedAt,
	).Scan(&el.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to enroll course: %w", err)
	}
	return el, nil
}

// CreateKnowledgeArticle creates a new knowledge base article.
func (r *EmployeeRepository) CreateKnowledgeArticle(ctx context.Context, article *types.KnowledgeArticle) (*types.KnowledgeArticle, error) {
	article.ID = uuid.New()
	article.CreatedAt = time.Now()
	article.UpdatedAt = time.Now()

	var tagsParam interface{}
	if article.Tags != nil {
		b, _ := json.Marshal(article.Tags)
		tagsParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO knowledge_articles (id, tenant_id, title, slug, body, category, tags, author_id, status, view_count, published_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		article.ID, article.TenantID, article.Title, article.Slug, article.Body, article.Category, tagsParam, article.AuthorID, article.Status, article.ViewCount, article.PublishedAt, article.CreatedAt, article.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create knowledge article: %w", err)
	}
	return article, nil
}

// CreateCompensationRecord creates a compensation record for an employee.
func (r *EmployeeRepository) CreateCompensationRecord(ctx context.Context, rec *types.CompensationRecord) (*types.CompensationRecord, error) {
	rec.ID = uuid.New()
	rec.CreatedAt = time.Now()
	rec.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO compensation_records (id, employee_id, tenant_id, base_salary_cents, currency, pay_frequency, effective_date, end_date, review_date, notes, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		rec.ID, rec.EmployeeID, rec.TenantID, rec.BaseSalaryCents, rec.Currency, rec.PayFrequency, rec.EffectiveDate, rec.EndDate, rec.ReviewDate, rec.Notes, rec.CreatedBy, rec.CreatedAt, rec.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create compensation record: %w", err)
	}
	return rec, nil
}

// CreateEquityGrant creates an equity grant for an employee.
func (r *EmployeeRepository) CreateEquityGrant(ctx context.Context, grant *types.EquityGrant) (*types.EquityGrant, error) {
	grant.ID = uuid.New()
	grant.CreatedAt = time.Now()
	grant.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO equity_grants (id, employee_id, tenant_id, grant_type, shares, strike_price_cents, vesting_start, vesting_end, cliff_date, vested_shares, status, grant_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		grant.ID, grant.EmployeeID, grant.TenantID, grant.GrantType, grant.Shares, grant.StrikePriceCents, grant.VestingStart, grant.VestingEnd, grant.CliffDate, grant.VestedShares, grant.Status, grant.GrantDate, grant.CreatedAt, grant.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create equity grant: %w", err)
	}
	return grant, nil
}

// LogCompensationAccess logs access to compensation data.
func (r *EmployeeRepository) LogCompensationAccess(ctx context.Context, log *types.CompensationAccessLog) error {
	log.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO compensation_access_log (accessor_id, target_id, action, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		log.AccessorID, log.TargetID, log.Action, log.IPAddress, log.UserAgent, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to log compensation access: %w", err)
	}
	return nil
}

// CreateNotification creates an FWOS notification.
func (r *EmployeeRepository) CreateNotification(ctx context.Context, notif *types.FWOSNotification) (*types.FWOSNotification, error) {
	notif.ID = uuid.New()
	notif.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO fwos_notifications (id, user_id, tenant_id, type, title, body, action_url, is_read, read_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		notif.ID, notif.UserID, notif.TenantID, notif.Type, notif.Title, notif.Body, notif.ActionURL, notif.IsRead, notif.ReadAt, notif.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}
	return notif, nil
}
