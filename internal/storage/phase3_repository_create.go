package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateInnovationGrant creates a new innovation grant.
func (r *Phase3Repository) CreateInnovationGrant(ctx context.Context, grant *InnovationGrant) (*InnovationGrant, error) {
	grant.ID = uuid.New()
	grant.CreatedAt = time.Now()
	grant.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO innovation_grants (id, tenant_id, proposer_id, title, description, category, requested_amount_cents, status, feasibility_score, votes_for, votes_against, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		grant.ID, grant.TenantID, grant.ProposerID, grant.Title, grant.Description, grant.Category, grant.RequestedAmountCents, grant.Status, grant.FeasibilityScore, grant.VotesFor, grant.VotesAgainst, grant.CreatedAt, grant.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create innovation grant: %w", err)
	}
	return grant, nil
}

// CreateInnovationGrantVote creates a vote on an innovation grant.
func (r *Phase3Repository) CreateInnovationGrantVote(ctx context.Context, vote *InnovationGrantVote) (*InnovationGrantVote, error) {
	vote.CreatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO innovation_grant_votes (grant_id, voter_id, vote, comment, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		vote.GrantID, vote.VoterID, vote.Vote, vote.Comment, vote.CreatedAt,
	).Scan(&vote.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create innovation grant vote: %w", err)
	}
	return vote, nil
}

// CreateMarketplaceOpportunity creates a new marketplace opportunity.
func (r *Phase3Repository) CreateMarketplaceOpportunity(ctx context.Context, opp *MarketplaceOpportunity) (*MarketplaceOpportunity, error) {
	opp.ID = uuid.New()
	opp.CreatedAt = time.Now()
	opp.UpdatedAt = time.Now()

	var skillsParam interface{}
	if opp.SkillsRequired != nil {
		b, _ := json.Marshal(opp.SkillsRequired)
		skillsParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO marketplace_opportunities (id, tenant_id, posted_by, department_id, title, description, opportunity_type, skills_required, hours_per_week, duration_weeks, is_remote, status, max_applicants, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		opp.ID, opp.TenantID, opp.PostedBy, opp.DepartmentID, opp.Title, opp.Description, opp.OpportunityType, skillsParam, opp.HoursPerWeek, opp.DurationWeeks, opp.IsRemote, opp.Status, opp.MaxApplicants, opp.CreatedAt, opp.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create marketplace opportunity: %w", err)
	}
	return opp, nil
}

// CreateMarketplaceApplication creates a new marketplace application.
func (r *Phase3Repository) CreateMarketplaceApplication(ctx context.Context, app *MarketplaceApplication) (*MarketplaceApplication, error) {
	app.ID = uuid.New()
	app.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO marketplace_applications (id, opportunity_id, applicant_id, message, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		app.ID, app.OpportunityID, app.ApplicantID, app.Message, app.Status, app.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create marketplace application: %w", err)
	}
	return app, nil
}

// CreateCareerPath creates a new career path.
func (r *Phase3Repository) CreateCareerPath(ctx context.Context, path *CareerPath) (*CareerPath, error) {
	path.ID = uuid.New()
	path.CreatedAt = time.Now()
	path.UpdatedAt = time.Now()

	var reqParam interface{}
	if path.Requirements != nil {
		b, _ := json.Marshal(path.Requirements)
		reqParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO career_paths (id, tenant_id, title, track, level, description, requirements, salary_range_min_cents, salary_range_max_cents, next_path_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		path.ID, path.TenantID, path.Title, path.Track, path.Level, path.Description, reqParam, path.SalaryRangeMinCents, path.SalaryRangeMaxCents, path.NextPathID, path.IsActive, path.CreatedAt, path.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create career path: %w", err)
	}
	return path, nil
}

// CreateEmployeeCareerProgress creates a career progress record.
func (r *Phase3Repository) CreateEmployeeCareerProgress(ctx context.Context, prog *EmployeeCareerProgress) (*EmployeeCareerProgress, error) {
	prog.CreatedAt = time.Now()
	prog.UpdatedAt = time.Now()

	var gapParam interface{}
	if prog.GapAnalysis != nil {
		b, _ := json.Marshal(prog.GapAnalysis)
		gapParam = b
	}

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO employee_career_progress (employee_id, career_path_id, status, gap_analysis, started_at, target_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		prog.EmployeeID, prog.CareerPathID, prog.Status, gapParam, prog.StartedAt, prog.TargetDate, prog.CreatedAt, prog.UpdatedAt,
	).Scan(&prog.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create employee career progress: %w", err)
	}
	return prog, nil
}

// CreateMentorshipMatch creates a new mentorship match.
func (r *Phase3Repository) CreateMentorshipMatch(ctx context.Context, match *MentorshipMatch) (*MentorshipMatch, error) {
	match.ID = uuid.New()
	match.CreatedAt = time.Now()
	match.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO mentorship_matches (id, tenant_id, mentor_id, mentee_id, focus_area, status, started_at, ended_at, meeting_frequency, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		match.ID, match.TenantID, match.MentorID, match.MenteeID, match.FocusArea, match.Status, match.StartedAt, match.EndedAt, match.MeetingFrequency, match.Notes, match.CreatedAt, match.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create mentorship match: %w", err)
	}
	return match, nil
}

// CreateDocument creates a new document.
func (r *Phase3Repository) CreateDocument(ctx context.Context, doc *Document) (*Document, error) {
	doc.ID = uuid.New()
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()

	var tagsParam interface{}
	if doc.Tags != nil {
		b, _ := json.Marshal(doc.Tags)
		tagsParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO documents (id, tenant_id, author_id, title, body, doc_type, category, tags, is_template, status, view_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		doc.ID, doc.TenantID, doc.AuthorID, doc.Title, doc.Body, doc.DocType, doc.Category, tagsParam, doc.IsTemplate, doc.Status, doc.ViewCount, doc.CreatedAt, doc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}
	return doc, nil
}

// CreateDocumentShare creates a document share record.
func (r *Phase3Repository) CreateDocumentShare(ctx context.Context, share *DocumentShare) (*DocumentShare, error) {
	share.CreatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO document_shares (document_id, shared_with, permission, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		share.DocumentID, share.SharedWith, share.Permission, share.CreatedAt,
	).Scan(&share.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create document share: %w", err)
	}
	return share, nil
}
