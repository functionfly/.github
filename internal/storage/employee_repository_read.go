package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/types"
	"github.com/google/uuid"
)

type (
	Department                        = types.Department
	Employee                          = types.Employee
	EmployeeSkill                     = types.EmployeeSkill
	EmployeeCertification             = types.EmployeeCertification
	EmployeeAchievement               = types.EmployeeAchievement
	Project                           = types.Project
	Task                              = types.Task
	TaskComment                       = types.TaskComment
	LearningCourse                    = types.LearningCourse
	EmployeeLearning                  = types.EmployeeLearning
	KnowledgeArticle                  = types.KnowledgeArticle
	CompensationRecord                = types.CompensationRecord
	EquityGrant                       = types.EquityGrant
	CompensationAccessLog            = types.CompensationAccessLog
	FWOSNotification                 = types.FWOSNotification
	ListEmployeesOpts                 = types.ListEmployeesOpts
	ListProjectsOpts                  = types.ListProjectsOpts
	ListTasksOpts                     = types.ListTasksOpts
	ListKnowledgeOpts                 = types.ListKnowledgeOpts
	AIChatSession                     = types.AIChatSession
	AIChatMessage                     = types.AIChatMessage
	PerformanceGoal                    = types.PerformanceGoal
	PerformanceReview                = types.PerformanceReview
	PeerFeedback                      = types.PeerFeedback
	TimeEntry                         = types.TimeEntry
	PTORequest                        = types.PTORequest
	ListAIChatSessionsOpts            = types.ListAIChatSessionsOpts
	ListPerformanceGoalsOpts          = types.ListPerformanceGoalsOpts
	ListPerformanceReviewsOpts        = types.ListPerformanceReviewsOpts
	ListTimeEntriesOpts               = types.ListTimeEntriesOpts
	ListPTORequestsOpts               = types.ListPTORequestsOpts
	InnovationGrant                   = types.InnovationGrant
	InnovationGrantVote               = types.InnovationGrantVote
	MarketplaceOpportunity            = types.MarketplaceOpportunity
	MarketplaceApplication            = types.MarketplaceApplication
	CareerPath                        = types.CareerPath
	EmployeeCareerProgress            = types.EmployeeCareerProgress
	MentorshipMatch                   = types.MentorshipMatch
	Document                          = types.Document
	DocumentShare                      = types.DocumentShare
	ListInnovationGrantsOpts          = types.ListInnovationGrantsOpts
	ListMarketplaceOpportunitiesOpts   = types.ListMarketplaceOpportunitiesOpts
	ListCareerPathsOpts               = types.ListCareerPathsOpts
	ListMentorshipMatchesOpts         = types.ListMentorshipMatchesOpts
	ListDocumentsOpts                 = types.ListDocumentsOpts
	TeamHealthMetric                  = types.TeamHealthMetric
	SkillsGraph                       = types.SkillsGraph
	ReputationScore                    = types.ReputationScore
	DigitalBadge                      = types.DigitalBadge
	EmployeeBadge                     = types.EmployeeBadge
	LivingMemoryEntry                 = types.LivingMemoryEntry
	MissionControlSnapshot             = types.MissionControlSnapshot
	ListTeamHealthOpts                = types.ListTeamHealthOpts
	ListReputationOpts                = types.ListReputationOpts
	ListBadgesOpts                    = types.ListBadgesOpts
	ListLivingMemoryOpts              = types.ListLivingMemoryOpts
	SearchLivingMemoryOpts            = types.SearchLivingMemoryOpts
	FWOSIncident                     = types.FWOSIncident
	IncidentEvent                    = types.IncidentEvent
	IncidentResponder                 = types.IncidentResponder
	Postmortem                       = types.Postmortem
	LifecycleEvent                   = types.LifecycleEvent
	LifecycleWorkflow                = types.LifecycleWorkflow
	LifecycleWorkflowInstance        = types.LifecycleWorkflowInstance
	FeatureFlag                      = types.FeatureFlag
	DataClassification               = types.DataClassification
	EmployeeCertificate              = types.EmployeeCertificate
	FWOSEvent                        = types.FWOSEvent
	ListIncidentsOpts                = types.ListIncidentsOpts
	ListLifecycleEventsOpts          = types.ListLifecycleEventsOpts
	ListFeatureFlagsOpts             = types.ListFeatureFlagsOpts
	ListDataClassificationsOpts      = types.ListDataClassificationsOpts
	ListCertificatesOpts             = types.ListCertificatesOpts
	ListFWOSEventsOpts               = types.ListFWOSEventsOpts
	EmailAccount                     = types.EmailAccount
	Device                           = types.Device
	SSOProvisioningConfig           = types.SSOProvisioningConfig
	SSOProvisioningLog              = types.SSOProvisioningLog
	WalletPass                       = types.WalletPass
	PushSubscription                 = types.PushSubscription
	NotificationPreference           = types.NotificationPreference
	ListEmailAccountsOpts            = types.ListEmailAccountsOpts
	ListDevicesOpts                  = types.ListDevicesOpts
	ListSSOProvisioningConfigsOpts   = types.ListSSOProvisioningConfigsOpts
	ListSSOProvisioningLogsOpts     = types.ListSSOProvisioningLogsOpts
	ListWalletPassesOpts             = types.ListWalletPassesOpts
	ListNotificationPreferencesOpts  = types.ListNotificationPreferencesOpts
	FeedbackRound                    = types.FeedbackRound
	FeedbackRoundAssignment          = types.FeedbackRoundAssignment
	FeedbackRoundResponse            = types.FeedbackRoundResponse
	DocumentSignature                = types.DocumentSignature
	CertificateKey                   = types.CertificateKey
	WalletPassTemplate               = types.WalletPassTemplate
	OrgChartImport                   = types.OrgChartImport
	PackageRegistry                  = types.PackageRegistry
	PackageVersion                   = types.PackageVersion
	ListFeedbackRoundsOpts           = types.ListFeedbackRoundsOpts
	ListDocumentSignaturesOpts       = types.ListDocumentSignaturesOpts
	ListWalletPassTemplatesOpts      = types.ListWalletPassTemplatesOpts
	ListOrgChartImportsOpts          = types.ListOrgChartImportsOpts
	ListPackageRegistryOpts          = types.ListPackageRegistryOpts
	ListPackageVersionsOpts          = types.ListPackageVersionsOpts
	EmployeeDepartment               = types.EmployeeDepartment
	IdentityCard                     = types.IdentityCard
	AchievementDefinition            = types.AchievementDefinition
	AchievementProgress              = types.AchievementProgress
	CareerTimelineEvent              = types.CareerTimelineEvent
	ReputationHistory                = types.ReputationHistory
)

type (
	FounderVote struct {
		ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
		TenantID    uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
		Title       string     `json:"title" gorm:"not null"`
		Description string     `json:"description" gorm:"type:text"`
		Options     string     `json:"options" gorm:"type:jsonb;default:'[]'"` // JSON array of options
		ChangeDiff  string     `json:"change_diff" gorm:"type:jsonb;default:'{}'"` // JSON structured change info
		Status      string     `json:"status" gorm:"default:'active'"`
		Quorum      int        `json:"quorum" gorm:"default:0"` // 0 = no quorum required
		CreatedBy   uuid.UUID  `json:"created_by" gorm:"type:uuid;not null"`
		CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
		UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	}

	FounderVoteResponse struct {
		ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
		VoteID    uuid.UUID `json:"vote_id" gorm:"type:uuid;not null;index"`
		UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
		OptionID  string    `json:"option_id" gorm:"not null"`
		CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	}

	FounderEarlyAccessFeature struct {
		ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
		TenantID    uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`
		Slug        string    `json:"slug" gorm:"uniqueIndex;not null"`
		Name        string    `json:"name" gorm:"not null"`
		Description string    `json:"description" gorm:"type:text"`
		IsActive    bool      `json:"is_active" gorm:"default:true"`
		LaunchedAt  *time.Time `json:"launched_at,omitempty"`
		CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
		UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	}

	FounderEarlyAccess struct {
		ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
		UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
		FeatureSlug string    `json:"feature_slug" gorm:"not null"`
		FeatureName string    `json:"feature_name" gorm:"not null"`
		AccessedAt  time.Time `json:"accessed_at" gorm:"autoCreateTime"`
	}
)

func (FounderEarlyAccess) TableName() string { return "founder_early_access" }

// GetDepartmentByID retrieves a department by ID.
func (r *EmployeeRepository) GetDepartmentByID(ctx context.Context, id int64) (*Department, error) {
	dept := &Department{}
	var description sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, slug, description, parent_id, head_id, is_active, created_at, updated_at
		FROM departments WHERE id = $1`, id).Scan(
		&dept.ID, &dept.TenantID, &dept.Name, &dept.Slug, &description, &dept.ParentID, &dept.HeadID, &dept.IsActive, &dept.CreatedAt, &dept.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get department: %w", err)
	}
	if description.Valid {
		dept.Description = &description.String
	}
	return dept, nil
}

// ListDepartments lists all departments for a tenant.
func (r *EmployeeRepository) ListDepartments(ctx context.Context, tenantID uuid.UUID) ([]*Department, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, name, slug, description, parent_id, head_id, is_active, created_at, updated_at
		FROM departments WHERE tenant_id = $1 ORDER BY name`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list departments: %w", err)
	}
	defer rows.Close()

	var depts []*Department
	for rows.Next() {
		dept := &Department{}
		var description sql.NullString
		if err := rows.Scan(&dept.ID, &dept.TenantID, &dept.Name, &dept.Slug, &description, &dept.ParentID, &dept.HeadID, &dept.IsActive, &dept.CreatedAt, &dept.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan department: %w", err)
		}
		if description.Valid {
			dept.Description = &description.String
		}
		depts = append(depts, dept)
	}
	return depts, nil
}

// GetEmployeeByID retrieves an employee by ID.
func (r *EmployeeRepository) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*Employee, error) {
	return r.getEmployeeByColumn(ctx, "id", id)
}

// GetEmployeeByUserID retrieves an employee by user ID.
func (r *EmployeeRepository) GetEmployeeByUserID(ctx context.Context, userID uuid.UUID) (*Employee, error) {
	return r.getEmployeeByColumn(ctx, "user_id", userID)
}

// GetEmployeeByFFID retrieves an employee by FFID.
func (r *EmployeeRepository) GetEmployeeByFFID(ctx context.Context, ffid string) (*Employee, error) {
	return r.getEmployeeByColumn(ctx, "ffid", ffid)
}

func (r *EmployeeRepository) getEmployeeByColumn(ctx context.Context, column string, value interface{}) (*Employee, error) {
	emp := &Employee{}
	var departmentID sql.NullInt64
	var managerID sql.NullString
	var hireDate sql.NullTime
	var workLocation, officeLocation, timezone, bio, pronouns sql.NullString
	var emergencyContactData []byte

	query := fmt.Sprintf(`
		SELECT id, user_id, tenant_id, employee_number, ffid, department_id, manager_id, hire_date, employment_type, clearance_level, work_location, office_location, timezone, bio, pronouns, emergency_contact, status, created_at, updated_at
		FROM employees WHERE %s = $1`, column)

	err := r.db.QueryRowContext(ctx, query, value).Scan(
		&emp.ID, &emp.UserID, &emp.TenantID, &emp.EmployeeNumber, &emp.FFID, &departmentID, &managerID, &hireDate,
		&emp.EmploymentType, &emp.ClearanceLevel, &workLocation, &officeLocation, &timezone, &bio, &pronouns,
		&emergencyContactData, &emp.Status, &emp.CreatedAt, &emp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get employee by %s: %w", column, err)
	}

	if departmentID.Valid {
		did := departmentID.Int64
		emp.DepartmentID = &did
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
	if workLocation.Valid {
		emp.WorkLocation = &workLocation.String
	}
	if officeLocation.Valid {
		emp.OfficeLocation = &officeLocation.String
	}
	if timezone.Valid {
		emp.Timezone = &timezone.String
	}
	if bio.Valid {
		emp.Bio = &bio.String
	}
	if pronouns.Valid {
		emp.Pronouns = &pronouns.String
	}
	if emergencyContactData != nil {
		var ec JSONMap
		if err := json.Unmarshal(emergencyContactData, &ec); err == nil {
			emp.EmergencyContact = ec
		}
	}
	return emp, nil
}

// ListEmployeesOpts and related types are defined in types/fwos.go.

// ListEmployees lists employees for a tenant with filtering.
func (r *EmployeeRepository) ListEmployees(ctx context.Context, tenantID uuid.UUID, opts ListEmployeesOpts) ([]*Employee, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.DepartmentID != nil {
		where += fmt.Sprintf(" AND department_id = $%d", argIdx)
		args = append(args, *opts.DepartmentID)
		argIdx++
	}
	if opts.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *opts.Status)
		argIdx++
	}
	if opts.Search != nil {
		where += fmt.Sprintf(" AND (employee_number ILIKE $%d OR ffid ILIKE $%d OR bio ILIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, "%"+*opts.Search+"%")
		argIdx++
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM employees %s", where)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count employees: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := opts.Offset

	query := fmt.Sprintf(`
		SELECT id, user_id, tenant_id, employee_number, ffid, department_id, manager_id, hire_date, employment_type, clearance_level, work_location, office_location, timezone, bio, pronouns, emergency_contact, status, created_at, updated_at
		FROM employees %s ORDER BY employee_number LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list employees: %w", err)
	}
	defer rows.Close()

	var employees []*Employee
	for rows.Next() {
		emp := &Employee{}
		var departmentID sql.NullInt64
		var managerID sql.NullString
		var hireDate sql.NullTime
		var workLocation, officeLocation, timezone, bio, pronouns sql.NullString
		var emergencyContactData []byte

		if err := rows.Scan(
			&emp.ID, &emp.UserID, &emp.TenantID, &emp.EmployeeNumber, &emp.FFID, &departmentID, &managerID, &hireDate,
			&emp.EmploymentType, &emp.ClearanceLevel, &workLocation, &officeLocation, &timezone, &bio, &pronouns,
			&emergencyContactData, &emp.Status, &emp.CreatedAt, &emp.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan employee: %w", err)
		}

		if departmentID.Valid {
			did := departmentID.Int64
			emp.DepartmentID = &did
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
		if workLocation.Valid {
			emp.WorkLocation = &workLocation.String
		}
		if officeLocation.Valid {
			emp.OfficeLocation = &officeLocation.String
		}
		if timezone.Valid {
			emp.Timezone = &timezone.String
		}
		if bio.Valid {
			emp.Bio = &bio.String
		}
		if pronouns.Valid {
			emp.Pronouns = &pronouns.String
		}
		if emergencyContactData != nil {
			var ec JSONMap
			if err := json.Unmarshal(emergencyContactData, &ec); err == nil {
				emp.EmergencyContact = ec
			}
		}
		employees = append(employees, emp)
	}
	return employees, total, nil
}

// GetEmployeeSkills retrieves skills for an employee.
func (r *EmployeeRepository) GetEmployeeSkills(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeSkill, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, skill_name, category, proficiency, years_exp, endorsements, verified, created_at, updated_at
		FROM employee_skills WHERE employee_id = $1 ORDER BY skill_name`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee skills: %w", err)
	}
	defer rows.Close()

	var skills []*EmployeeSkill
	for rows.Next() {
		s := &EmployeeSkill{}
		var category sql.NullString
		var yearsExp sql.NullFloat64
		if err := rows.Scan(&s.ID, &s.EmployeeID, &s.SkillName, &category, &s.Proficiency, &yearsExp, &s.Endorsements, &s.Verified, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan employee skill: %w", err)
		}
		if category.Valid {
			s.Category = &category.String
		}
		if yearsExp.Valid {
			s.YearsExp = &yearsExp.Float64
		}
		skills = append(skills, s)
	}
	return skills, nil
}

// GetEmployeeCertifications retrieves certifications for an employee.
func (r *EmployeeRepository) GetEmployeeCertifications(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeCertification, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, name, issuer, credential_id, credential_url, issued_date, expiry_date, verified, created_at
		FROM employee_certifications WHERE employee_id = $1 ORDER BY created_at DESC`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee certifications: %w", err)
	}
	defer rows.Close()

	var certs []*EmployeeCertification
	for rows.Next() {
		c := &EmployeeCertification{}
		var credentialID, credentialURL sql.NullString
		var issuedDate, expiryDate sql.NullTime
		if err := rows.Scan(&c.ID, &c.EmployeeID, &c.Name, &c.Issuer, &credentialID, &credentialURL, &issuedDate, &expiryDate, &c.Verified, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan employee certification: %w", err)
		}
		if credentialID.Valid {
			c.CredentialID = &credentialID.String
		}
		if credentialURL.Valid {
			c.CredentialURL = &credentialURL.String
		}
		if issuedDate.Valid {
			c.IssuedDate = &issuedDate.Time
		}
		if expiryDate.Valid {
			c.ExpiryDate = &expiryDate.Time
		}
		certs = append(certs, c)
	}
	return certs, nil
}

// GetEmployeeAchievements retrieves achievements for an employee.
func (r *EmployeeRepository) GetEmployeeAchievements(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeAchievement, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, title, description, type, awarded_by, points, badge_url, awarded_at
		FROM employee_achievements WHERE employee_id = $1 ORDER BY awarded_at DESC`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee achievements: %w", err)
	}
	defer rows.Close()

	var achievements []*EmployeeAchievement
	for rows.Next() {
		a := &EmployeeAchievement{}
		var description, badgeURL sql.NullString
		if err := rows.Scan(&a.ID, &a.EmployeeID, &a.Title, &description, &a.Type, &a.AwardedBy, &a.Points, &badgeURL, &a.AwardedAt); err != nil {
			return nil, fmt.Errorf("failed to scan employee achievement: %w", err)
		}
		if description.Valid {
			a.Description = &description.String
		}
		if badgeURL.Valid {
			a.BadgeURL = &badgeURL.String
		}
		achievements = append(achievements, a)
	}
	return achievements, nil
}

// GetProjectByID retrieves a project by ID.
func (r *EmployeeRepository) GetProjectByID(ctx context.Context, id uuid.UUID) (*Project, error) {
	proj := &Project{}
	var description sql.NullString
	var tagsData, metadataData []byte
	var startDate, targetDate sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, slug, description, status, priority, owner_id, start_date, target_date, tags, metadata, created_at, updated_at
		FROM projects WHERE id = $1`, id).Scan(
		&proj.ID, &proj.TenantID, &proj.Name, &proj.Slug, &description, &proj.Status, &proj.Priority, &proj.OwnerID,
		&startDate, &targetDate, &tagsData, &metadataData, &proj.CreatedAt, &proj.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if description.Valid {
		proj.Description = &description.String
	}
	if startDate.Valid {
		proj.StartDate = &startDate.Time
	}
	if targetDate.Valid {
		proj.TargetDate = &targetDate.Time
	}
	if tagsData != nil {
		var tags JSONMap
		if err := json.Unmarshal(tagsData, &tags); err == nil {
			proj.Tags = tags
		}
	}
	if metadataData != nil {
		var metadata JSONMap
		if err := json.Unmarshal(metadataData, &metadata); err == nil {
			proj.Metadata = metadata
		}
	}
	return proj, nil
}

// ListProjects lists projects for a tenant with filtering.
func (r *EmployeeRepository) ListProjects(ctx context.Context, tenantID uuid.UUID, opts ListProjectsOpts) ([]*Project, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *opts.Status)
		argIdx++
	}
	if opts.OwnerID != nil {
		where += fmt.Sprintf(" AND owner_id = $%d", argIdx)
		args = append(args, *opts.OwnerID)
		argIdx++
	}
	if opts.Search != nil {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR slug ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+*opts.Search+"%")
		argIdx++
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM projects %s", where)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := opts.Offset

	query := fmt.Sprintf(`
		SELECT id, tenant_id, name, slug, description, status, priority, owner_id, start_date, target_date, tags, metadata, created_at, updated_at
		FROM projects %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		proj := &Project{}
		var description sql.NullString
		var tagsData, metadataData []byte
		var startDate, targetDate sql.NullTime

		if err := rows.Scan(
			&proj.ID, &proj.TenantID, &proj.Name, &proj.Slug, &description, &proj.Status, &proj.Priority, &proj.OwnerID,
			&startDate, &targetDate, &tagsData, &metadataData, &proj.CreatedAt, &proj.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan project: %w", err)
		}

		if description.Valid {
			proj.Description = &description.String
		}
		if startDate.Valid {
			proj.StartDate = &startDate.Time
		}
		if targetDate.Valid {
			proj.TargetDate = &targetDate.Time
		}
		if tagsData != nil {
			var tags JSONMap
			if err := json.Unmarshal(tagsData, &tags); err == nil {
				proj.Tags = tags
			}
		}
		if metadataData != nil {
			var metadata JSONMap
			if err := json.Unmarshal(metadataData, &metadata); err == nil {
				proj.Metadata = metadata
			}
		}
		projects = append(projects, proj)
	}
	return projects, total, nil
}

// GetTaskByID retrieves a task by ID.
func (r *EmployeeRepository) GetTaskByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	task := &Task{}
	var description sql.NullString
	var tagsData []byte
	var dueDate sql.NullTime
	var estimatedHours, actualHours sql.NullFloat64

	err := r.db.QueryRowContext(ctx, `
		SELECT id, project_id, tenant_id, title, description, status, priority, assignee_id, reporter_id, parent_id, due_date, estimated_hours, actual_hours, tags, position, created_at, updated_at
		FROM tasks WHERE id = $1`, id).Scan(
		&task.ID, &task.ProjectID, &task.TenantID, &task.Title, &description, &task.Status, &task.Priority,
		&task.AssigneeID, &task.ReporterID, &task.ParentID, &dueDate, &estimatedHours, &actualHours,
		&tagsData, &task.Position, &task.CreatedAt, &task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if description.Valid {
		task.Description = &description.String
	}
	if dueDate.Valid {
		task.DueDate = &dueDate.Time
	}
	if estimatedHours.Valid {
		task.EstimatedHours = &estimatedHours.Float64
	}
	if actualHours.Valid {
		task.ActualHours = &actualHours.Float64
	}
	if tagsData != nil {
		var tags JSONMap
		if err := json.Unmarshal(tagsData, &tags); err == nil {
			task.Tags = tags
		}
	}
	return task, nil
}

// ListTasks lists tasks for a tenant with filtering.
func (r *EmployeeRepository) ListTasks(ctx context.Context, tenantID uuid.UUID, opts ListTasksOpts) ([]*Task, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.ProjectID != nil {
		where += fmt.Sprintf(" AND project_id = $%d", argIdx)
		args = append(args, *opts.ProjectID)
		argIdx++
	}
	if opts.AssigneeID != nil {
		where += fmt.Sprintf(" AND assignee_id = $%d", argIdx)
		args = append(args, *opts.AssigneeID)
		argIdx++
	}
	if opts.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *opts.Status)
		argIdx++
	}
	if opts.Priority != nil {
		where += fmt.Sprintf(" AND priority = $%d", argIdx)
		args = append(args, *opts.Priority)
		argIdx++
	}
	if opts.Search != nil {
		where += fmt.Sprintf(" AND title ILIKE $%d", argIdx)
		args = append(args, "%"+*opts.Search+"%")
		argIdx++
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks %s", where)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := opts.Offset

	query := fmt.Sprintf(`
		SELECT id, project_id, tenant_id, title, description, status, priority, assignee_id, reporter_id, parent_id, due_date, estimated_hours, actual_hours, tags, position, created_at, updated_at
		FROM tasks %s ORDER BY position, created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		var description sql.NullString
		var tagsData []byte
		var dueDate sql.NullTime
		var estimatedHours, actualHours sql.NullFloat64

		if err := rows.Scan(
			&task.ID, &task.ProjectID, &task.TenantID, &task.Title, &description, &task.Status, &task.Priority,
			&task.AssigneeID, &task.ReporterID, &task.ParentID, &dueDate, &estimatedHours, &actualHours,
			&tagsData, &task.Position, &task.CreatedAt, &task.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan task: %w", err)
		}

		if description.Valid {
			task.Description = &description.String
		}
		if dueDate.Valid {
			task.DueDate = &dueDate.Time
		}
		if estimatedHours.Valid {
			task.EstimatedHours = &estimatedHours.Float64
		}
		if actualHours.Valid {
			task.ActualHours = &actualHours.Float64
		}
		if tagsData != nil {
			var tags JSONMap
			if err := json.Unmarshal(tagsData, &tags); err == nil {
				task.Tags = tags
			}
		}
		tasks = append(tasks, task)
	}
	return tasks, total, nil
}

// GetTaskComments retrieves comments for a task.
func (r *EmployeeRepository) GetTaskComments(ctx context.Context, taskID uuid.UUID) ([]*TaskComment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, task_id, author_id, body, created_at, updated_at
		FROM task_comments WHERE task_id = $1 ORDER BY created_at`, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task comments: %w", err)
	}
	defer rows.Close()

	var comments []*TaskComment
	for rows.Next() {
		c := &TaskComment{}
		if err := rows.Scan(&c.ID, &c.TaskID, &c.AuthorID, &c.Body, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan task comment: %w", err)
		}
		comments = append(comments, c)
	}
	return comments, nil
}

// GetLearningCourseByID retrieves a learning course by ID.
func (r *EmployeeRepository) GetLearningCourseByID(ctx context.Context, id uuid.UUID) (*LearningCourse, error) {
	course := &LearningCourse{}
	var description, category, difficulty, contentURL, thumbnailURL sql.NullString
	var durationMin sql.NullInt64

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, title, description, category, difficulty, duration_min, content_url, thumbnail_url, is_mandatory, is_active, created_at, updated_at
		FROM learning_courses WHERE id = $1`, id).Scan(
		&course.ID, &course.TenantID, &course.Title, &description, &category, &difficulty, &durationMin,
		&contentURL, &thumbnailURL, &course.IsMandatory, &course.IsActive, &course.CreatedAt, &course.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get learning course: %w", err)
	}

	if description.Valid {
		course.Description = &description.String
	}
	if category.Valid {
		course.Category = &category.String
	}
	if difficulty.Valid {
		course.Difficulty = &difficulty.String
	}
	if durationMin.Valid {
		dm := int(durationMin.Int64)
		course.DurationMin = &dm
	}
	if contentURL.Valid {
		course.ContentURL = &contentURL.String
	}
	if thumbnailURL.Valid {
		course.ThumbnailURL = &thumbnailURL.String
	}
	return course, nil
}

// ListLearningCourses lists learning courses for a tenant.
func (r *EmployeeRepository) ListLearningCourses(ctx context.Context, tenantID uuid.UUID) ([]*LearningCourse, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, title, description, category, difficulty, duration_min, content_url, thumbnail_url, is_mandatory, is_active, created_at, updated_at
		FROM learning_courses WHERE tenant_id = $1 AND is_active = TRUE ORDER BY title`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list learning courses: %w", err)
	}
	defer rows.Close()

	var courses []*LearningCourse
	for rows.Next() {
		course := &LearningCourse{}
		var description, category, difficulty, contentURL, thumbnailURL sql.NullString
		var durationMin sql.NullInt64

		if err := rows.Scan(
			&course.ID, &course.TenantID, &course.Title, &description, &category, &difficulty, &durationMin,
			&contentURL, &thumbnailURL, &course.IsMandatory, &course.IsActive, &course.CreatedAt, &course.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan learning course: %w", err)
		}

		if description.Valid {
			course.Description = &description.String
		}
		if category.Valid {
			course.Category = &category.String
		}
		if difficulty.Valid {
			course.Difficulty = &difficulty.String
		}
		if durationMin.Valid {
			dm := int(durationMin.Int64)
			course.DurationMin = &dm
		}
		if contentURL.Valid {
			course.ContentURL = &contentURL.String
		}
		if thumbnailURL.Valid {
			course.ThumbnailURL = &thumbnailURL.String
		}
		courses = append(courses, course)
	}
	return courses, nil
}

// GetEmployeeLearning retrieves learning enrollments for an employee.
func (r *EmployeeRepository) GetEmployeeLearning(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeLearning, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, course_id, status, progress_pct, started_at, completed_at, score, created_at, updated_at
		FROM employee_learning WHERE employee_id = $1 ORDER BY created_at DESC`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee learning: %w", err)
	}
	defer rows.Close()

	var enrollments []*EmployeeLearning
	for rows.Next() {
		el := &EmployeeLearning{}
		var startedAt, completedAt sql.NullTime
		var score sql.NullFloat64

		if err := rows.Scan(&el.ID, &el.EmployeeID, &el.CourseID, &el.Status, &el.ProgressPct, &startedAt, &completedAt, &score, &el.CreatedAt, &el.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan employee learning: %w", err)
		}
		if startedAt.Valid {
			el.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			el.CompletedAt = &completedAt.Time
		}
		if score.Valid {
			el.Score = &score.Float64
		}
		enrollments = append(enrollments, el)
	}
	return enrollments, nil
}

// GetKnowledgeArticleByID retrieves a knowledge article by ID.
func (r *EmployeeRepository) GetKnowledgeArticleByID(ctx context.Context, id uuid.UUID) (*KnowledgeArticle, error) {
	article := &KnowledgeArticle{}
	var category sql.NullString
	var tagsData []byte
	var publishedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, title, slug, body, category, tags, author_id, status, view_count, published_at, created_at, updated_at
		FROM knowledge_articles WHERE id = $1`, id).Scan(
		&article.ID, &article.TenantID, &article.Title, &article.Slug, &article.Body, &category,
		&tagsData, &article.AuthorID, &article.Status, &article.ViewCount, &publishedAt,
		&article.CreatedAt, &article.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge article: %w", err)
	}

	if category.Valid {
		article.Category = &category.String
	}
	if tagsData != nil {
		var tags JSONMap
		if err := json.Unmarshal(tagsData, &tags); err == nil {
			article.Tags = tags
		}
	}
	if publishedAt.Valid {
		article.PublishedAt = &publishedAt.Time
	}
	return article, nil
}

// GetKnowledgeArticleBySlug retrieves a knowledge article by tenant and slug.
func (r *EmployeeRepository) GetKnowledgeArticleBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*KnowledgeArticle, error) {
	article := &KnowledgeArticle{}
	var category sql.NullString
	var tagsData []byte
	var publishedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, title, slug, body, category, tags, author_id, status, view_count, published_at, created_at, updated_at
		FROM knowledge_articles WHERE tenant_id = $1 AND slug = $2`, tenantID, slug).Scan(
		&article.ID, &article.TenantID, &article.Title, &article.Slug, &article.Body, &category,
		&tagsData, &article.AuthorID, &article.Status, &article.ViewCount, &publishedAt,
		&article.CreatedAt, &article.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge article by slug: %w", err)
	}

	if category.Valid {
		article.Category = &category.String
	}
	if tagsData != nil {
		var tags JSONMap
		if err := json.Unmarshal(tagsData, &tags); err == nil {
			article.Tags = tags
		}
	}
	if publishedAt.Valid {
		article.PublishedAt = &publishedAt.Time
	}
	return article, nil
}

// ListKnowledgeArticles lists knowledge articles for a tenant with filtering.
func (r *EmployeeRepository) ListKnowledgeArticles(ctx context.Context, tenantID uuid.UUID, opts ListKnowledgeOpts) ([]*KnowledgeArticle, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.Category != nil {
		where += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, *opts.Category)
		argIdx++
	}
	if opts.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *opts.Status)
		argIdx++
	}
	if opts.Search != nil {
		where += fmt.Sprintf(" AND (title ILIKE $%d OR body ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+*opts.Search+"%")
		argIdx++
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM knowledge_articles %s", where)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count knowledge articles: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := opts.Offset

	query := fmt.Sprintf(`
		SELECT id, tenant_id, title, slug, body, category, tags, author_id, status, view_count, published_at, created_at, updated_at
		FROM knowledge_articles %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list knowledge articles: %w", err)
	}
	defer rows.Close()

	var articles []*KnowledgeArticle
	for rows.Next() {
		article := &KnowledgeArticle{}
		var category sql.NullString
		var tagsData []byte
		var publishedAt sql.NullTime

		if err := rows.Scan(
			&article.ID, &article.TenantID, &article.Title, &article.Slug, &article.Body, &category,
			&tagsData, &article.AuthorID, &article.Status, &article.ViewCount, &publishedAt,
			&article.CreatedAt, &article.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan knowledge article: %w", err)
		}

		if category.Valid {
			article.Category = &category.String
		}
		if tagsData != nil {
			var tags JSONMap
			if err := json.Unmarshal(tagsData, &tags); err == nil {
				article.Tags = tags
			}
		}
		if publishedAt.Valid {
			article.PublishedAt = &publishedAt.Time
		}
		articles = append(articles, article)
	}
	return articles, total, nil
}

// SearchKnowledgeArticles performs a text search on knowledge articles.
func (r *EmployeeRepository) SearchKnowledgeArticles(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]*KnowledgeArticle, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, title, slug, body, category, tags, author_id, status, view_count, published_at, created_at, updated_at
		FROM knowledge_articles
		WHERE tenant_id = $1 AND status = 'published' AND (title ILIKE $2 OR body ILIKE $2)
		ORDER BY view_count DESC
		LIMIT $3`, tenantID, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search knowledge articles: %w", err)
	}
	defer rows.Close()

	var articles []*KnowledgeArticle
	for rows.Next() {
		article := &KnowledgeArticle{}
		var category sql.NullString
		var tagsData []byte
		var publishedAt sql.NullTime

		if err := rows.Scan(
			&article.ID, &article.TenantID, &article.Title, &article.Slug, &article.Body, &category,
			&tagsData, &article.AuthorID, &article.Status, &article.ViewCount, &publishedAt,
			&article.CreatedAt, &article.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan knowledge article: %w", err)
		}

		if category.Valid {
			article.Category = &category.String
		}
		if tagsData != nil {
			var tags JSONMap
			if err := json.Unmarshal(tagsData, &tags); err == nil {
				article.Tags = tags
			}
		}
		if publishedAt.Valid {
			article.PublishedAt = &publishedAt.Time
		}
		articles = append(articles, article)
	}
	return articles, nil
}

// GetActiveCompensation retrieves the active compensation record for an employee.
func (r *EmployeeRepository) GetActiveCompensation(ctx context.Context, employeeID uuid.UUID) (*CompensationRecord, error) {
	rec := &CompensationRecord{}
	var endDate, reviewDate sql.NullTime
	var notes sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, employee_id, tenant_id, base_salary_cents, currency, pay_frequency, effective_date, end_date, review_date, notes, created_by, created_at, updated_at
		FROM compensation_records
		WHERE employee_id = $1 AND (end_date IS NULL OR end_date > CURRENT_DATE)
		ORDER BY effective_date DESC LIMIT 1`, employeeID).Scan(
		&rec.ID, &rec.EmployeeID, &rec.TenantID, &rec.BaseSalaryCents, &rec.Currency, &rec.PayFrequency,
		&rec.EffectiveDate, &endDate, &reviewDate, &notes, &rec.CreatedBy, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active compensation: %w", err)
	}

	if endDate.Valid {
		rec.EndDate = &endDate.Time
	}
	if reviewDate.Valid {
		rec.ReviewDate = &reviewDate.Time
	}
	if notes.Valid {
		rec.Notes = &notes.String
	}
	return rec, nil
}

// ListEquityGrants lists equity grants for an employee.
func (r *EmployeeRepository) ListEquityGrants(ctx context.Context, employeeID uuid.UUID) ([]*EquityGrant, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, tenant_id, grant_type, shares, strike_price_cents, vesting_start, vesting_end, cliff_date, vested_shares, status, grant_date, created_at, updated_at
		FROM equity_grants WHERE employee_id = $1 ORDER BY grant_date DESC`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list equity grants: %w", err)
	}
	defer rows.Close()

	var grants []*EquityGrant
	for rows.Next() {
		g := &EquityGrant{}
		var strikePriceCents sql.NullInt64
		var cliffDate sql.NullTime

		if err := rows.Scan(
			&g.ID, &g.EmployeeID, &g.TenantID, &g.GrantType, &g.Shares, &strikePriceCents,
			&g.VestingStart, &g.VestingEnd, &cliffDate, &g.VestedShares, &g.Status, &g.GrantDate,
			&g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan equity grant: %w", err)
		}

		if strikePriceCents.Valid {
			g.StrikePriceCents = &strikePriceCents.Int64
		}
		if cliffDate.Valid {
			g.CliffDate = &cliffDate.Time
		}
		grants = append(grants, g)
	}
	return grants, nil
}

// GetCompensationAccessLog retrieves compensation access logs for an employee.
func (r *EmployeeRepository) GetCompensationAccessLog(ctx context.Context, employeeID uuid.UUID) ([]*CompensationAccessLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, accessor_id, target_id, action, ip_address, user_agent, created_at
		FROM compensation_access_log WHERE target_id = $1 ORDER BY created_at DESC`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get compensation access log: %w", err)
	}
	defer rows.Close()

	var logs []*CompensationAccessLog
	for rows.Next() {
		l := &CompensationAccessLog{}
		var ipAddress, userAgent sql.NullString
		if err := rows.Scan(&l.ID, &l.AccessorID, &l.TargetID, &l.Action, &ipAddress, &userAgent, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan compensation access log: %w", err)
		}
		if ipAddress.Valid {
			l.IPAddress = &ipAddress.String
		}
		if userAgent.Valid {
			l.UserAgent = &userAgent.String
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// ListNotifications lists notifications for a user.
func (r *EmployeeRepository) ListNotifications(ctx context.Context, userID uuid.UUID, unreadOnly bool, limit, offset int) ([]*FWOSNotification, int, error) {
	where := "WHERE user_id = $1"
	args := []interface{}{userID}
	argIdx := 2

	if unreadOnly {
		where += " AND is_read = FALSE"
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM fwos_notifications %s", where)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	if limit <= 0 {
		limit = 50
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, tenant_id, type, title, body, action_url, is_read, read_at, created_at
		FROM fwos_notifications %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*FWOSNotification
	for rows.Next() {
		n := &FWOSNotification{}
		var body, actionURL sql.NullString
		var readAt sql.NullTime

		if err := rows.Scan(&n.ID, &n.UserID, &n.TenantID, &n.Type, &n.Title, &body, &actionURL, &n.IsRead, &readAt, &n.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan notification: %w", err)
		}
		if body.Valid {
			n.Body = &body.String
		}
		if actionURL.Valid {
			n.ActionURL = &actionURL.String
		}
		if readAt.Valid {
			n.ReadAt = &readAt.Time
		}
		notifications = append(notifications, n)
	}
	return notifications, total, nil
}

// CountUnreadNotifications counts unread notifications for a user.
func (r *EmployeeRepository) CountUnreadNotifications(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM fwos_notifications WHERE user_id = $1 AND is_read = FALSE`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}
	return count, nil
}

// GetDirectReports retrieves employees reporting to a manager.
func (r *EmployeeRepository) GetDirectReports(ctx context.Context, managerID uuid.UUID) ([]*Employee, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, tenant_id, employee_number, ffid, department_id, manager_id, hire_date, employment_type, clearance_level, work_location, office_location, timezone, bio, pronouns, emergency_contact, status, created_at, updated_at
		FROM employees WHERE manager_id = $1 ORDER BY employee_number`, managerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get direct reports: %w", err)
	}
	defer rows.Close()

	return scanEmployeeRows(rows)
}

// GetOrgChart retrieves all active employees for a tenant (for org chart rendering).
func (r *EmployeeRepository) GetOrgChart(ctx context.Context, tenantID uuid.UUID) ([]*Employee, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, tenant_id, employee_number, ffid, department_id, manager_id, hire_date, employment_type, clearance_level, work_location, office_location, timezone, bio, pronouns, emergency_contact, status, created_at, updated_at
		FROM employees WHERE tenant_id = $1 AND status = 'active' ORDER BY employee_number`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get org chart: %w", err)
	}
	defer rows.Close()

	return scanEmployeeRows(rows)
}

// scanEmployeeRows is a helper to scan multiple employee rows.
func scanEmployeeRows(rows *sql.Rows) ([]*Employee, error) {
	var employees []*Employee
	for rows.Next() {
		emp := &Employee{}
		var departmentID sql.NullInt64
		var managerID sql.NullString
		var hireDate sql.NullTime
		var workLocation, officeLocation, timezone, bio, pronouns sql.NullString
		var emergencyContactData []byte

		if err := rows.Scan(
			&emp.ID, &emp.UserID, &emp.TenantID, &emp.EmployeeNumber, &emp.FFID, &departmentID, &managerID, &hireDate,
			&emp.EmploymentType, &emp.ClearanceLevel, &workLocation, &officeLocation, &timezone, &bio, &pronouns,
			&emergencyContactData, &emp.Status, &emp.CreatedAt, &emp.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan employee: %w", err)
		}

		if departmentID.Valid {
			did := departmentID.Int64
			emp.DepartmentID = &did
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
		if workLocation.Valid {
			emp.WorkLocation = &workLocation.String
		}
		if officeLocation.Valid {
			emp.OfficeLocation = &officeLocation.String
		}
		if timezone.Valid {
			emp.Timezone = &timezone.String
		}
		if bio.Valid {
			emp.Bio = &bio.String
		}
		if pronouns.Valid {
			emp.Pronouns = &pronouns.String
		}
		if emergencyContactData != nil {
			var ec JSONMap
			if err := json.Unmarshal(emergencyContactData, &ec); err == nil {
				emp.EmergencyContact = ec
			}
		}
		employees = append(employees, emp)
	}
	return employees, nil
}
