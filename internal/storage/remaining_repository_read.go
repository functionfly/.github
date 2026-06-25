package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// GetFeedbackRoundByID retrieves a feedback round by ID.
func (r *RemainingRepository) GetFeedbackRoundByID(ctx context.Context, id uuid.UUID) (*FeedbackRound, error) {
	fr := &FeedbackRound{}
	var questionsBytes []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, description, review_period, round_type, status, start_date, end_date, questions, created_by, created_at, updated_at
		FROM feedback_rounds WHERE id = $1`, id).Scan(
		&fr.ID, &fr.TenantID, &fr.Name, &fr.Description, &fr.ReviewPeriod, &fr.RoundType, &fr.Status, &fr.StartDate, &fr.EndDate, &questionsBytes, &fr.CreatedBy, &fr.CreatedAt, &fr.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback round: %w", err)
	}
	if questionsBytes != nil {
		var q JSONMap
		if err := json.Unmarshal(questionsBytes, &q); err == nil {
			fr.Questions = q
		}
	}
	return fr, nil
}

// ListFeedbackRounds lists feedback rounds for a tenant.
func (r *RemainingRepository) ListFeedbackRounds(ctx context.Context, tenantID uuid.UUID, opts ListFeedbackRoundsOpts) ([]*FeedbackRound, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *opts.Status)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM feedback_rounds %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count feedback rounds: %w", err)
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
		SELECT id, tenant_id, name, description, review_period, round_type, status, start_date, end_date, questions, created_by, created_at, updated_at
		FROM feedback_rounds %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list feedback rounds: %w", err)
	}
	defer rows.Close()

	var rounds []*FeedbackRound
	for rows.Next() {
		fr := &FeedbackRound{}
		var questionsBytes []byte
		if err := rows.Scan(&fr.ID, &fr.TenantID, &fr.Name, &fr.Description, &fr.ReviewPeriod, &fr.RoundType, &fr.Status, &fr.StartDate, &fr.EndDate, &questionsBytes, &fr.CreatedBy, &fr.CreatedAt, &fr.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan feedback round: %w", err)
		}
		if questionsBytes != nil {
			var q JSONMap
			if err := json.Unmarshal(questionsBytes, &q); err == nil {
				fr.Questions = q
			}
		}
		rounds = append(rounds, fr)
	}
	return rounds, total, nil
}

// ListFeedbackRoundAssignments lists assignments for a feedback round.
func (r *RemainingRepository) ListFeedbackRoundAssignments(ctx context.Context, roundID uuid.UUID) ([]*FeedbackRoundAssignment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, round_id, reviewer_id, reviewee_id, status, submitted_at, created_at
		FROM feedback_round_assignments WHERE round_id = $1 ORDER BY id`, roundID)
	if err != nil {
		return nil, fmt.Errorf("failed to list feedback round assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*FeedbackRoundAssignment
	for rows.Next() {
		a := &FeedbackRoundAssignment{}
		if err := rows.Scan(&a.ID, &a.RoundID, &a.ReviewerID, &a.RevieweeID, &a.Status, &a.SubmittedAt, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan feedback round assignment: %w", err)
		}
		assignments = append(assignments, a)
	}
	return assignments, nil
}

// ListFeedbackRoundResponses lists responses for an assignment.
func (r *RemainingRepository) ListFeedbackRoundResponses(ctx context.Context, assignmentID int64) ([]*FeedbackRoundResponse, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, assignment_id, question_index, response_text, response_rating, created_at
		FROM feedback_round_responses WHERE assignment_id = $1 ORDER BY question_index`, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list feedback round responses: %w", err)
	}
	defer rows.Close()

	var responses []*FeedbackRoundResponse
	for rows.Next() {
		resp := &FeedbackRoundResponse{}
		if err := rows.Scan(&resp.ID, &resp.AssignmentID, &resp.QuestionIndex, &resp.ResponseText, &resp.ResponseRating, &resp.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan feedback round response: %w", err)
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

// GetFeedbackRoundResults aggregates results for a feedback round.
func (r *RemainingRepository) GetFeedbackRoundResults(ctx context.Context, roundID uuid.UUID) ([]map[string]interface{}, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.reviewee_id, e.employee_number, r.question_index, AVG(r.response_rating) as avg_rating, COUNT(r.id) as response_count
		FROM feedback_round_assignments a
		JOIN feedback_round_responses r ON r.assignment_id = a.id
		JOIN employees e ON e.id = a.reviewee_id
		WHERE a.round_id = $1 AND r.response_rating IS NOT NULL
		GROUP BY a.reviewee_id, e.employee_number, r.question_index
		ORDER BY a.reviewee_id, r.question_index`, roundID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback round results: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var revieweeID uuid.UUID
		var employeeNumber string
		var questionIndex int
		var avgRating float64
		var responseCount int
		if err := rows.Scan(&revieweeID, &employeeNumber, &questionIndex, &avgRating, &responseCount); err != nil {
			return nil, fmt.Errorf("failed to scan feedback round result: %w", err)
		}
		results = append(results, map[string]interface{}{
			"reviewee_id":     revieweeID,
			"employee_number": employeeNumber,
			"question_index":  questionIndex,
			"avg_rating":      avgRating,
			"response_count":  responseCount,
		})
	}
	return results, nil
}

// GetGoalTree retrieves all goals for a tenant to build a hierarchy tree.
func (r *RemainingRepository) GetGoalTree(ctx context.Context, tenantID uuid.UUID) ([]*PerformanceGoal, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, tenant_id, title, description, category, status, priority, target_date, completed_at, progress_pct, created_at, updated_at, parent_goal_id, goal_level, cascade_visibility
		FROM performance_goals WHERE tenant_id = $1 ORDER BY goal_level, created_at`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get goal tree: %w", err)
	}
	defer rows.Close()

	var goals []*PerformanceGoal
	for rows.Next() {
		g := &PerformanceGoal{}
		var parentGoalID *uuid.UUID
		var goalLevel, cascadeVisibility *string
		if err := rows.Scan(&g.ID, &g.EmployeeID, &g.TenantID, &g.Title, &g.Description, &g.Category, &g.Status, &g.Priority, &g.TargetDate, &g.CompletedAt, &g.ProgressPct, &g.CreatedAt, &g.UpdatedAt, &parentGoalID, &goalLevel, &cascadeVisibility); err != nil {
			return nil, fmt.Errorf("failed to scan goal: %w", err)
		}
		goals = append(goals, g)
	}
	return goals, nil
}

// GetDocumentSignatureByID retrieves a document signature by ID.
func (r *RemainingRepository) GetDocumentSignatureByID(ctx context.Context, id uuid.UUID) (*DocumentSignature, error) {
	ds := &DocumentSignature{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, document_id, signer_id, signer_name, signer_email, signature_data, signed_at, status, decline_reason, expires_at, created_at
		FROM document_signatures WHERE id = $1`, id).Scan(
		&ds.ID, &ds.DocumentID, &ds.SignerID, &ds.SignerName, &ds.SignerEmail, &ds.SignatureData, &ds.SignedAt, &ds.Status, &ds.DeclineReason, &ds.ExpiresAt, &ds.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document signature: %w", err)
	}
	return ds, nil
}

// ListDocumentSignatures lists signatures for a document.
func (r *RemainingRepository) ListDocumentSignatures(ctx context.Context, documentID uuid.UUID, opts ListDocumentSignaturesOpts) ([]*DocumentSignature, int, error) {
	where := "WHERE document_id = $1"
	args := []interface{}{documentID}
	argIdx := 2

	if opts.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *opts.Status)
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM document_signatures %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count document signatures: %w", err)
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
		SELECT id, document_id, signer_id, signer_name, signer_email, signature_data, signed_at, status, decline_reason, expires_at, created_at
		FROM document_signatures %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list document signatures: %w", err)
	}
	defer rows.Close()

	var sigs []*DocumentSignature
	for rows.Next() {
		ds := &DocumentSignature{}
		if err := rows.Scan(&ds.ID, &ds.DocumentID, &ds.SignerID, &ds.SignerName, &ds.SignerEmail, &ds.SignatureData, &ds.SignedAt, &ds.Status, &ds.DeclineReason, &ds.ExpiresAt, &ds.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan document signature: %w", err)
		}
		sigs = append(sigs, ds)
	}
	return sigs, total, nil
}

// GetCertificateKeysByCertID retrieves certificate keys by certificate ID.
func (r *RemainingRepository) GetCertificateKeysByCertID(ctx context.Context, certificateID uuid.UUID) ([]*CertificateKey, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, certificate_id, private_key_pem, public_key_pem, key_type, key_size, created_at
		FROM certificate_keys WHERE certificate_id = $1 ORDER BY created_at DESC`, certificateID)
	if err != nil {
		return nil, fmt.Errorf("failed to list certificate keys: %w", err)
	}
	defer rows.Close()

	var keys []*CertificateKey
	for rows.Next() {
		ck := &CertificateKey{}
		if err := rows.Scan(&ck.ID, &ck.CertificateID, &ck.PrivateKeyPEM, &ck.PublicKeyPEM, &ck.KeyType, &ck.KeySize, &ck.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan certificate key: %w", err)
		}
		keys = append(keys, ck)
	}
	return keys, nil
}

// ListWalletPassTemplates lists wallet pass templates for a tenant.
func (r *RemainingRepository) ListWalletPassTemplates(ctx context.Context, tenantID uuid.UUID, opts ListWalletPassTemplatesOpts) ([]*WalletPassTemplate, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.Platform != nil {
		where += fmt.Sprintf(" AND platform = $%d", argIdx)
		args = append(args, *opts.Platform)
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
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM wallet_pass_templates %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count wallet pass templates: %w", err)
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
		SELECT id, tenant_id, name, pass_type, platform, template_data, is_active, created_at, updated_at
		FROM wallet_pass_templates %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list wallet pass templates: %w", err)
	}
	defer rows.Close()

	var templates []*WalletPassTemplate
	for rows.Next() {
		t := &WalletPassTemplate{}
		var templateDataBytes []byte
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Name, &t.PassType, &t.Platform, &templateDataBytes, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan wallet pass template: %w", err)
		}
		if templateDataBytes != nil {
			var td JSONMap
			if err := json.Unmarshal(templateDataBytes, &td); err == nil {
				t.TemplateData = td
			}
		}
		templates = append(templates, t)
	}
	return templates, total, nil
}

// GetOrgChartImportByID retrieves an org chart import by ID.
func (r *RemainingRepository) GetOrgChartImportByID(ctx context.Context, id uuid.UUID) (*OrgChartImport, error) {
	imp := &OrgChartImport{}
	var errorsBytes []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, uploaded_by, file_name, file_type, status, total_rows, processed_rows, error_rows, errors, created_at, completed_at
		FROM org_chart_imports WHERE id = $1`, id).Scan(
		&imp.ID, &imp.TenantID, &imp.UploadedBy, &imp.FileName, &imp.FileType, &imp.Status, &imp.TotalRows, &imp.ProcessedRows, &imp.ErrorRows, &errorsBytes, &imp.CreatedAt, &imp.CompletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get org chart import: %w", err)
	}
	if errorsBytes != nil {
		var e JSONMap
		if err := json.Unmarshal(errorsBytes, &e); err == nil {
			imp.Errors = e
		}
	}
	return imp, nil
}

// GetPackageByID retrieves a package by ID.
func (r *RemainingRepository) GetPackageByID(ctx context.Context, id uuid.UUID) (*PackageRegistry, error) {
	pkg := &PackageRegistry{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, scope, description, registry_type, latest_version, total_downloads, is_internal, repository_url, published_by, published_at, created_at, updated_at
		FROM package_registry WHERE id = $1`, id).Scan(
		&pkg.ID, &pkg.TenantID, &pkg.Name, &pkg.Scope, &pkg.Description, &pkg.RegistryType, &pkg.LatestVersion, &pkg.TotalDownloads, &pkg.IsInternal, &pkg.RepositoryURL, &pkg.PublishedBy, &pkg.PublishedAt, &pkg.CreatedAt, &pkg.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get package: %w", err)
	}
	return pkg, nil
}

// GetPackageByName retrieves a package by name and type.
func (r *RemainingRepository) GetPackageByName(ctx context.Context, tenantID uuid.UUID, name, registryType string) (*PackageRegistry, error) {
	pkg := &PackageRegistry{}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, scope, description, registry_type, latest_version, total_downloads, is_internal, repository_url, published_by, published_at, created_at, updated_at
		FROM package_registry WHERE tenant_id = $1 AND name = $2 AND registry_type = $3`, tenantID, name, registryType).Scan(
		&pkg.ID, &pkg.TenantID, &pkg.Name, &pkg.Scope, &pkg.Description, &pkg.RegistryType, &pkg.LatestVersion, &pkg.TotalDownloads, &pkg.IsInternal, &pkg.RepositoryURL, &pkg.PublishedBy, &pkg.PublishedAt, &pkg.CreatedAt, &pkg.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get package by name: %w", err)
	}
	return pkg, nil
}

// ListPackages lists packages for a tenant.
func (r *RemainingRepository) ListPackages(ctx context.Context, tenantID uuid.UUID, opts ListPackageRegistryOpts) ([]*PackageRegistry, int, error) {
	where := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIdx := 2

	if opts.RegistryType != nil {
		where += fmt.Sprintf(" AND registry_type = $%d", argIdx)
		args = append(args, *opts.RegistryType)
		argIdx++
	}
	if opts.Search != nil {
		where += fmt.Sprintf(" AND name ILIKE $%d", argIdx)
		args = append(args, "%"+*opts.Search+"%")
		argIdx++
	}

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int
	err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM package_registry %s", where), countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count packages: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	where += fmt.Sprintf(" ORDER BY name ASC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, tenant_id, name, scope, description, registry_type, latest_version, total_downloads, is_internal, repository_url, published_by, published_at, created_at, updated_at
		FROM package_registry %s`, where), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list packages: %w", err)
	}
	defer rows.Close()

	var packages []*PackageRegistry
	for rows.Next() {
		pkg := &PackageRegistry{}
		if err := rows.Scan(&pkg.ID, &pkg.TenantID, &pkg.Name, &pkg.Scope, &pkg.Description, &pkg.RegistryType, &pkg.LatestVersion, &pkg.TotalDownloads, &pkg.IsInternal, &pkg.RepositoryURL, &pkg.PublishedBy, &pkg.PublishedAt, &pkg.CreatedAt, &pkg.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan package: %w", err)
		}
		packages = append(packages, pkg)
	}
	return packages, total, nil
}

// ListPackageVersions lists versions for a package.
func (r *RemainingRepository) ListPackageVersions(ctx context.Context, packageID uuid.UUID, opts ListPackageVersionsOpts) ([]*PackageVersion, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM package_versions WHERE package_id = $1", packageID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count package versions: %w", err)
	}

	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, package_id, version, description, dependencies, published_by, downloads, tarball_url, published_at
		FROM package_versions WHERE package_id = $1 ORDER BY published_at DESC LIMIT $2 OFFSET $3`,
		packageID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list package versions: %w", err)
	}
	defer rows.Close()

	var versions []*PackageVersion
	for rows.Next() {
		v := &PackageVersion{}
		var depsBytes []byte
		if err := rows.Scan(&v.ID, &v.PackageID, &v.Version, &v.Description, &depsBytes, &v.PublishedBy, &v.Downloads, &v.TarballURL, &v.PublishedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan package version: %w", err)
		}
		if depsBytes != nil {
			var d JSONMap
			if err := json.Unmarshal(depsBytes, &d); err == nil {
				v.Dependencies = d
			}
		}
		versions = append(versions, v)
	}
	return versions, total, nil
}
