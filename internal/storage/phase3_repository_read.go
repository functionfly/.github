package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// GetInnovationGrantByID retrieves an innovation grant by ID.
func (r *Phase3Repository) GetInnovationGrantByID(ctx context.Context, id uuid.UUID) (*InnovationGrant, error) {
	grant := &InnovationGrant{}
	var category sql.NullString
	var requestedAmount sql.NullInt64
	var feasibilityScore sql.NullFloat64
	var reviewedBy sql.NullString
	var reviewedAt sql.NullTime
	var rejectionReason sql.NullString
	var fundedAt sql.NullTime
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, proposer_id, title, description, category, requested_amount_cents, status, feasibility_score, votes_for, votes_against, reviewed_by, reviewed_at, rejection_reason, funded_at, completed_at, created_at, updated_at
		FROM innovation_grants WHERE id = $1`, id).Scan(
		&grant.ID, &grant.TenantID, &grant.ProposerID, &grant.Title, &grant.Description, &category, &requestedAmount, &grant.Status, &feasibilityScore, &grant.VotesFor, &grant.VotesAgainst, &reviewedBy, &reviewedAt, &rejectionReason, &fundedAt, &completedAt, &grant.CreatedAt, &grant.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get innovation grant: %w", err)
	}
	if category.Valid {
		grant.Category = category.String
	}
	if requestedAmount.Valid {
		grant.RequestedAmountCents = &requestedAmount.Int64
	}
	if feasibilityScore.Valid {
		grant.FeasibilityScore = &feasibilityScore.Float64
	}
	if reviewedBy.Valid {
		rid, err := uuid.Parse(reviewedBy.String)
		if err == nil {
			grant.ReviewedBy = &rid
		}
	}
	if reviewedAt.Valid {
		grant.ReviewedAt = &reviewedAt.Time
	}
	if rejectionReason.Valid {
		grant.RejectionReason = &rejectionReason.String
	}
	if fundedAt.Valid {
		grant.FundedAt = &fundedAt.Time
	}
	if completedAt.Valid {
		grant.CompletedAt = &completedAt.Time
	}
	return grant, nil
}

// ListInnovationGrants lists innovation grants for a tenant.
func (r *Phase3Repository) ListInnovationGrants(ctx context.Context, tenantID uuid.UUID, opts ListInnovationGrantsOpts) ([]*InnovationGrant, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *opts.Status)
		argIdx++
	}
	if opts.Category != nil {
		where += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, *opts.Category)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM innovation_grants %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count innovation grants: %w", err)
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
		SELECT id, tenant_id, proposer_id, title, description, category, requested_amount_cents, status, feasibility_score, votes_for, votes_against, reviewed_by, reviewed_at, rejection_reason, funded_at, completed_at, created_at, updated_at
		FROM innovation_grants %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list innovation grants: %w", err)
	}
	defer rows.Close()

	var grants []*InnovationGrant
	for rows.Next() {
		grant := &InnovationGrant{}
		var category sql.NullString
		var requestedAmount sql.NullInt64
		var feasibilityScore sql.NullFloat64
		var reviewedBy sql.NullString
		var reviewedAt sql.NullTime
		var rejectionReason sql.NullString
		var fundedAt sql.NullTime
		var completedAt sql.NullTime
		if err := rows.Scan(&grant.ID, &grant.TenantID, &grant.ProposerID, &grant.Title, &grant.Description, &category, &requestedAmount, &grant.Status, &feasibilityScore, &grant.VotesFor, &grant.VotesAgainst, &reviewedBy, &reviewedAt, &rejectionReason, &fundedAt, &completedAt, &grant.CreatedAt, &grant.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan innovation grant: %w", err)
		}
		if category.Valid {
			grant.Category = category.String
		}
		if requestedAmount.Valid {
			grant.RequestedAmountCents = &requestedAmount.Int64
		}
		if feasibilityScore.Valid {
			grant.FeasibilityScore = &feasibilityScore.Float64
		}
		if reviewedBy.Valid {
			rid, err := uuid.Parse(reviewedBy.String)
			if err == nil {
				grant.ReviewedBy = &rid
			}
		}
		if reviewedAt.Valid {
			grant.ReviewedAt = &reviewedAt.Time
		}
		if rejectionReason.Valid {
			grant.RejectionReason = &rejectionReason.String
		}
		if fundedAt.Valid {
			grant.FundedAt = &fundedAt.Time
		}
		if completedAt.Valid {
			grant.CompletedAt = &completedAt.Time
		}
		grants = append(grants, grant)
	}
	return grants, total, nil
}

// GetInnovationGrantVoteByVoter retrieves a vote by grant and voter.
func (r *Phase3Repository) GetInnovationGrantVoteByVoter(ctx context.Context, grantID, voterID uuid.UUID) (*InnovationGrantVote, error) {
	vote := &InnovationGrantVote{}
	var comment sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, grant_id, voter_id, vote, comment, created_at
		FROM innovation_grant_votes WHERE grant_id = $1 AND voter_id = $2`, grantID, voterID).Scan(
		&vote.ID, &vote.GrantID, &vote.VoterID, &vote.Vote, &comment, &vote.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get innovation grant vote: %w", err)
	}
	if comment.Valid {
		vote.Comment = &comment.String
	}
	return vote, nil
}

// GetMarketplaceOpportunityByID retrieves a marketplace opportunity by ID.
func (r *Phase3Repository) GetMarketplaceOpportunityByID(ctx context.Context, id uuid.UUID) (*MarketplaceOpportunity, error) {
	opp := &MarketplaceOpportunity{}
	var deptID sql.NullInt64
	var skillsBytes []byte
	var hoursPerWeek sql.NullFloat64
	var durationWeeks sql.NullInt64
	var maxApplicants sql.NullInt64

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, posted_by, department_id, title, description, opportunity_type, skills_required, hours_per_week, duration_weeks, is_remote, status, max_applicants, created_at, updated_at
		FROM marketplace_opportunities WHERE id = $1`, id).Scan(
		&opp.ID, &opp.TenantID, &opp.PostedBy, &deptID, &opp.Title, &opp.Description, &opp.OpportunityType, &skillsBytes, &hoursPerWeek, &durationWeeks, &opp.IsRemote, &opp.Status, &maxApplicants, &opp.CreatedAt, &opp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get marketplace opportunity: %w", err)
	}
	if deptID.Valid {
		opp.DepartmentID = &deptID.Int64
	}
	if skillsBytes != nil {
		var skills JSONMap
		if err := json.Unmarshal(skillsBytes, &skills); err == nil {
			opp.SkillsRequired = skills
		}
	}
	if hoursPerWeek.Valid {
		opp.HoursPerWeek = &hoursPerWeek.Float64
	}
	if durationWeeks.Valid {
		dw := int(durationWeeks.Int64)
		opp.DurationWeeks = &dw
	}
	if maxApplicants.Valid {
		ma := int(maxApplicants.Int64)
		opp.MaxApplicants = &ma
	}
	return opp, nil
}

// ListMarketplaceOpportunities lists marketplace opportunities for a tenant.
func (r *Phase3Repository) ListMarketplaceOpportunities(ctx context.Context, tenantID uuid.UUID, opts ListMarketplaceOpportunitiesOpts) ([]*MarketplaceOpportunity, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.Type != nil {
		where += fmt.Sprintf(" AND opportunity_type = $%d", argIdx)
		args = append(args, *opts.Type)
		argIdx++
	}
	if opts.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *opts.Status)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM marketplace_opportunities %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count marketplace opportunities: %w", err)
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
		SELECT id, tenant_id, posted_by, department_id, title, description, opportunity_type, skills_required, hours_per_week, duration_weeks, is_remote, status, max_applicants, created_at, updated_at
		FROM marketplace_opportunities %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list marketplace opportunities: %w", err)
	}
	defer rows.Close()

	var opps []*MarketplaceOpportunity
	for rows.Next() {
		opp := &MarketplaceOpportunity{}
		var deptID sql.NullInt64
		var skillsBytes []byte
		var hoursPerWeek sql.NullFloat64
		var durationWeeks sql.NullInt64
		var maxApplicants sql.NullInt64
		if err := rows.Scan(&opp.ID, &opp.TenantID, &opp.PostedBy, &deptID, &opp.Title, &opp.Description, &opp.OpportunityType, &skillsBytes, &hoursPerWeek, &durationWeeks, &opp.IsRemote, &opp.Status, &maxApplicants, &opp.CreatedAt, &opp.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan marketplace opportunity: %w", err)
		}
		if deptID.Valid {
			opp.DepartmentID = &deptID.Int64
		}
		if skillsBytes != nil {
			var skills JSONMap
			if err := json.Unmarshal(skillsBytes, &skills); err == nil {
				opp.SkillsRequired = skills
			}
		}
		if hoursPerWeek.Valid {
			opp.HoursPerWeek = &hoursPerWeek.Float64
		}
		if durationWeeks.Valid {
			dw := int(durationWeeks.Int64)
			opp.DurationWeeks = &dw
		}
		if maxApplicants.Valid {
			ma := int(maxApplicants.Int64)
			opp.MaxApplicants = &ma
		}
		opps = append(opps, opp)
	}
	return opps, total, nil
}

// GetMarketplaceApplicationByID retrieves a marketplace application by ID.
func (r *Phase3Repository) GetMarketplaceApplicationByID(ctx context.Context, id uuid.UUID) (*MarketplaceApplication, error) {
	app := &MarketplaceApplication{}
	var message sql.NullString
	var reviewedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, opportunity_id, applicant_id, message, status, reviewed_at, created_at
		FROM marketplace_applications WHERE id = $1`, id).Scan(
		&app.ID, &app.OpportunityID, &app.ApplicantID, &message, &app.Status, &reviewedAt, &app.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get marketplace application: %w", err)
	}
	if message.Valid {
		app.Message = &message.String
	}
	if reviewedAt.Valid {
		app.ReviewedAt = &reviewedAt.Time
	}
	return app, nil
}

// ListMarketplaceApplications lists applications for an opportunity.
func (r *Phase3Repository) ListMarketplaceApplications(ctx context.Context, opportunityID uuid.UUID, limit, offset int) ([]*MarketplaceApplication, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, opportunity_id, applicant_id, message, status, reviewed_at, created_at
		FROM marketplace_applications WHERE opportunity_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		opportunityID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list marketplace applications: %w", err)
	}
	defer rows.Close()

	var apps []*MarketplaceApplication
	for rows.Next() {
		app := &MarketplaceApplication{}
		var message sql.NullString
		var reviewedAt sql.NullTime
		if err := rows.Scan(&app.ID, &app.OpportunityID, &app.ApplicantID, &message, &app.Status, &reviewedAt, &app.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan marketplace application: %w", err)
		}
		if message.Valid {
			app.Message = &message.String
		}
		if reviewedAt.Valid {
			app.ReviewedAt = &reviewedAt.Time
		}
		apps = append(apps, app)
	}
	return apps, nil
}

// GetCareerPathByID retrieves a career path by ID.
func (r *Phase3Repository) GetCareerPathByID(ctx context.Context, id uuid.UUID) (*CareerPath, error) {
	path := &CareerPath{}
	var description sql.NullString
	var reqBytes []byte
	var salaryMin, salaryMax sql.NullInt64
	var nextPathID sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, title, track, level, description, requirements, salary_range_min_cents, salary_range_max_cents, next_path_id, is_active, created_at, updated_at
		FROM career_paths WHERE id = $1`, id).Scan(
		&path.ID, &path.TenantID, &path.Title, &path.Track, &path.Level, &description, &reqBytes, &salaryMin, &salaryMax, &nextPathID, &path.IsActive, &path.CreatedAt, &path.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get career path: %w", err)
	}
	if description.Valid {
		path.Description = &description.String
	}
	if reqBytes != nil {
		var req JSONMap
		if err := json.Unmarshal(reqBytes, &req); err == nil {
			path.Requirements = req
		}
	}
	if salaryMin.Valid {
		path.SalaryRangeMinCents = &salaryMin.Int64
	}
	if salaryMax.Valid {
		path.SalaryRangeMaxCents = &salaryMax.Int64
	}
	if nextPathID.Valid {
		npid, err := uuid.Parse(nextPathID.String)
		if err == nil {
			path.NextPathID = &npid
		}
	}
	return path, nil
}

// ListCareerPaths lists career paths for a tenant.
func (r *Phase3Repository) ListCareerPaths(ctx context.Context, tenantID uuid.UUID, opts ListCareerPathsOpts) ([]*CareerPath, int, error) {
	where := "WHERE tenant_id = $1 AND is_active = TRUE"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.Track != nil {
		where += fmt.Sprintf(" AND track = $%d", argIdx)
		args = append(args, *opts.Track)
		argIdx++
	}
	if opts.Level != nil {
		where += fmt.Sprintf(" AND level = $%d", argIdx)
		args = append(args, *opts.Level)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM career_paths %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count career paths: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY track, level LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, tenant_id, title, track, level, description, requirements, salary_range_min_cents, salary_range_max_cents, next_path_id, is_active, created_at, updated_at
		FROM career_paths %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list career paths: %w", err)
	}
	defer rows.Close()

	var paths []*CareerPath
	for rows.Next() {
		path := &CareerPath{}
		var description sql.NullString
		var reqBytes []byte
		var salaryMin, salaryMax sql.NullInt64
		var nextPathID sql.NullString
		if err := rows.Scan(&path.ID, &path.TenantID, &path.Title, &path.Track, &path.Level, &description, &reqBytes, &salaryMin, &salaryMax, &nextPathID, &path.IsActive, &path.CreatedAt, &path.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan career path: %w", err)
		}
		if description.Valid {
			path.Description = &description.String
		}
		if reqBytes != nil {
			var req JSONMap
			if err := json.Unmarshal(reqBytes, &req); err == nil {
				path.Requirements = req
			}
		}
		if salaryMin.Valid {
			path.SalaryRangeMinCents = &salaryMin.Int64
		}
		if salaryMax.Valid {
			path.SalaryRangeMaxCents = &salaryMax.Int64
		}
		if nextPathID.Valid {
			npid, err := uuid.Parse(nextPathID.String)
			if err == nil {
				path.NextPathID = &npid
			}
		}
		paths = append(paths, path)
	}
	return paths, total, nil
}

// GetEmployeeCareerProgressByEmployee retrieves career progress for an employee.
func (r *Phase3Repository) GetEmployeeCareerProgressByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeCareerProgress, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, career_path_id, status, gap_analysis, started_at, target_date, created_at, updated_at
		FROM employee_career_progress WHERE employee_id = $1 ORDER BY created_at DESC`,
		employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee career progress: %w", err)
	}
	defer rows.Close()

	var progs []*EmployeeCareerProgress
	for rows.Next() {
		prog := &EmployeeCareerProgress{}
		var gapBytes []byte
		var startedAt sql.NullTime
		var targetDate sql.NullTime
		if err := rows.Scan(&prog.ID, &prog.EmployeeID, &prog.CareerPathID, &prog.Status, &gapBytes, &startedAt, &targetDate, &prog.CreatedAt, &prog.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan employee career progress: %w", err)
		}
		if gapBytes != nil {
			var gap JSONMap
			if err := json.Unmarshal(gapBytes, &gap); err == nil {
				prog.GapAnalysis = gap
			}
		}
		if startedAt.Valid {
			prog.StartedAt = &startedAt.Time
		}
		if targetDate.Valid {
			prog.TargetDate = &targetDate.Time
		}
		progs = append(progs, prog)
	}
	return progs, nil
}

// GetMentorshipMatchByID retrieves a mentorship match by ID.
func (r *Phase3Repository) GetMentorshipMatchByID(ctx context.Context, id uuid.UUID) (*MentorshipMatch, error) {
	match := &MentorshipMatch{}
	var focusArea, meetingFreq, notes sql.NullString
	var endedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, mentor_id, mentee_id, focus_area, status, started_at, ended_at, meeting_frequency, notes, created_at, updated_at
		FROM mentorship_matches WHERE id = $1`, id).Scan(
		&match.ID, &match.TenantID, &match.MentorID, &match.MenteeID, &focusArea, &match.Status, &match.StartedAt, &endedAt, &meetingFreq, &notes, &match.CreatedAt, &match.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get mentorship match: %w", err)
	}
	if focusArea.Valid {
		match.FocusArea = &focusArea.String
	}
	if endedAt.Valid {
		match.EndedAt = &endedAt.Time
	}
	if meetingFreq.Valid {
		match.MeetingFrequency = &meetingFreq.String
	}
	if notes.Valid {
		match.Notes = &notes.String
	}
	return match, nil
}

// ListMentorshipMatches lists mentorship matches for an employee (as mentor or mentee).
func (r *Phase3Repository) ListMentorshipMatches(ctx context.Context, employeeID uuid.UUID, opts ListMentorshipMatchesOpts) ([]*MentorshipMatch, int, error) {
	where := "WHERE (mentor_id = $1 OR mentee_id = $1)"
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
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM mentorship_matches %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count mentorship matches: %w", err)
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
		SELECT id, tenant_id, mentor_id, mentee_id, focus_area, status, started_at, ended_at, meeting_frequency, notes, created_at, updated_at
		FROM mentorship_matches %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list mentorship matches: %w", err)
	}
	defer rows.Close()

	var matches []*MentorshipMatch
	for rows.Next() {
		match := &MentorshipMatch{}
		var focusArea, meetingFreq, notes sql.NullString
		var endedAt sql.NullTime
		if err := rows.Scan(&match.ID, &match.TenantID, &match.MentorID, &match.MenteeID, &focusArea, &match.Status, &match.StartedAt, &endedAt, &meetingFreq, &notes, &match.CreatedAt, &match.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan mentorship match: %w", err)
		}
		if focusArea.Valid {
			match.FocusArea = &focusArea.String
		}
		if endedAt.Valid {
			match.EndedAt = &endedAt.Time
		}
		if meetingFreq.Valid {
			match.MeetingFrequency = &meetingFreq.String
		}
		if notes.Valid {
			match.Notes = &notes.String
		}
		matches = append(matches, match)
	}
	return matches, total, nil
}

// GetDocumentByID retrieves a document by ID.
func (r *Phase3Repository) GetDocumentByID(ctx context.Context, id uuid.UUID) (*Document, error) {
	doc := &Document{}
	var body, category sql.NullString
	var tagsBytes []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, author_id, title, body, doc_type, category, tags, is_template, status, view_count, created_at, updated_at
		FROM documents WHERE id = $1`, id).Scan(
		&doc.ID, &doc.TenantID, &doc.AuthorID, &doc.Title, &body, &doc.DocType, &category, &tagsBytes, &doc.IsTemplate, &doc.Status, &doc.ViewCount, &doc.CreatedAt, &doc.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	if body.Valid {
		doc.Body = &body.String
	}
	if category.Valid {
		doc.Category = &category.String
	}
	if tagsBytes != nil {
		var tags JSONMap
		if err := json.Unmarshal(tagsBytes, &tags); err == nil {
			doc.Tags = tags
		}
	}
	return doc, nil
}

// ListDocuments lists documents for a tenant.
func (r *Phase3Repository) ListDocuments(ctx context.Context, tenantID uuid.UUID, opts ListDocumentsOpts) ([]*Document, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.DocType != nil {
		where += fmt.Sprintf(" AND doc_type = $%d", argIdx)
		args = append(args, *opts.DocType)
		argIdx++
	}
	if opts.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *opts.Status)
		argIdx++
	}
	if opts.Search != nil {
		where += fmt.Sprintf(" AND title ILIKE $%d", argIdx)
		args = append(args, "%"+*opts.Search+"%")
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM documents %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY updated_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, tenant_id, author_id, title, body, doc_type, category, tags, is_template, status, view_count, created_at, updated_at
		FROM documents %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var docs []*Document
	for rows.Next() {
		doc := &Document{}
		var body, category sql.NullString
		var tagsBytes []byte
		if err := rows.Scan(&doc.ID, &doc.TenantID, &doc.AuthorID, &doc.Title, &body, &doc.DocType, &category, &tagsBytes, &doc.IsTemplate, &doc.Status, &doc.ViewCount, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan document: %w", err)
		}
		if body.Valid {
			doc.Body = &body.String
		}
		if category.Valid {
			doc.Category = &category.String
		}
		if tagsBytes != nil {
			var tags JSONMap
			if err := json.Unmarshal(tagsBytes, &tags); err == nil {
				doc.Tags = tags
			}
		}
		docs = append(docs, doc)
	}
	return docs, total, nil
}

// ListDocumentShares lists shares for a document.
func (r *Phase3Repository) ListDocumentShares(ctx context.Context, documentID uuid.UUID) ([]*DocumentShare, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, document_id, shared_with, permission, created_at
		FROM document_shares WHERE document_id = $1 ORDER BY created_at DESC`,
		documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list document shares: %w", err)
	}
	defer rows.Close()

	var shares []*DocumentShare
	for rows.Next() {
		share := &DocumentShare{}
		if err := rows.Scan(&share.ID, &share.DocumentID, &share.SharedWith, &share.Permission, &share.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan document share: %w", err)
		}
		shares = append(shares, share)
	}
	return shares, nil
}
