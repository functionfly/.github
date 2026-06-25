package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateFeedbackRound creates a new feedback round.
func (r *RemainingRepository) CreateFeedbackRound(ctx context.Context, fr *FeedbackRound) (*FeedbackRound, error) {
	fr.ID = uuid.New()
	fr.CreatedAt = time.Now()
	fr.UpdatedAt = time.Now()

	var questionsParam interface{}
	if fr.Questions != nil {
		b, _ := json.Marshal(fr.Questions)
		questionsParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO feedback_rounds (id, tenant_id, name, description, review_period, round_type, status, start_date, end_date, questions, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		fr.ID, fr.TenantID, fr.Name, fr.Description, fr.ReviewPeriod, fr.RoundType, fr.Status, fr.StartDate, fr.EndDate, questionsParam, fr.CreatedBy, fr.CreatedAt, fr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create feedback round: %w", err)
	}
	return fr, nil
}

// CreateFeedbackRoundAssignment creates a reviewer-reviewee assignment.
func (r *RemainingRepository) CreateFeedbackRoundAssignment(ctx context.Context, a *FeedbackRoundAssignment) (*FeedbackRoundAssignment, error) {
	a.CreatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO feedback_round_assignments (round_id, reviewer_id, reviewee_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		a.RoundID, a.ReviewerID, a.RevieweeID, a.Status, a.CreatedAt,
	).Scan(&a.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create feedback round assignment: %w", err)
	}
	return a, nil
}

// CreateFeedbackRoundResponse creates a feedback response.
func (r *RemainingRepository) CreateFeedbackRoundResponse(ctx context.Context, resp *FeedbackRoundResponse) (*FeedbackRoundResponse, error) {
	resp.CreatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO feedback_round_responses (assignment_id, question_index, response_text, response_rating, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		resp.AssignmentID, resp.QuestionIndex, resp.ResponseText, resp.ResponseRating, resp.CreatedAt,
	).Scan(&resp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create feedback round response: %w", err)
	}
	return resp, nil
}

// CreateDocumentSignature creates a document signature request.
func (r *RemainingRepository) CreateDocumentSignature(ctx context.Context, ds *DocumentSignature) (*DocumentSignature, error) {
	ds.ID = uuid.New()
	ds.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO document_signatures (id, document_id, signer_id, signer_name, signer_email, signature_data, signed_at, status, decline_reason, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		ds.ID, ds.DocumentID, ds.SignerID, ds.SignerName, ds.SignerEmail, ds.SignatureData, ds.SignedAt, ds.Status, ds.DeclineReason, ds.ExpiresAt, ds.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create document signature: %w", err)
	}
	return ds, nil
}

// CreateCertificateKey creates a certificate key pair.
func (r *RemainingRepository) CreateCertificateKey(ctx context.Context, ck *CertificateKey) (*CertificateKey, error) {
	ck.ID = uuid.New()
	ck.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO certificate_keys (id, certificate_id, private_key_pem, public_key_pem, key_type, key_size, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		ck.ID, ck.CertificateID, ck.PrivateKeyPEM, ck.PublicKeyPEM, ck.KeyType, ck.KeySize, ck.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate key: %w", err)
	}
	return ck, nil
}

// CreateWalletPassTemplate creates a wallet pass template.
func (r *RemainingRepository) CreateWalletPassTemplate(ctx context.Context, t *WalletPassTemplate) (*WalletPassTemplate, error) {
	t.ID = uuid.New()
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()

	var templateDataParam interface{}
	if t.TemplateData != nil {
		b, _ := json.Marshal(t.TemplateData)
		templateDataParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO wallet_pass_templates (id, tenant_id, name, pass_type, platform, template_data, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		t.ID, t.TenantID, t.Name, t.PassType, t.Platform, templateDataParam, t.IsActive, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet pass template: %w", err)
	}
	return t, nil
}

// CreateOrgChartImport creates an org chart import job.
func (r *RemainingRepository) CreateOrgChartImport(ctx context.Context, imp *OrgChartImport) (*OrgChartImport, error) {
	imp.ID = uuid.New()
	imp.CreatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO org_chart_imports (id, tenant_id, uploaded_by, file_name, file_type, status, total_rows, processed_rows, error_rows, errors, created_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		imp.ID, imp.TenantID, imp.UploadedBy, imp.FileName, imp.FileType, imp.Status, imp.TotalRows, imp.ProcessedRows, imp.ErrorRows, nil, imp.CreatedAt, imp.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create org chart import: %w", err)
	}
	return imp, nil
}

// CreatePackage creates a new package in the registry.
func (r *RemainingRepository) CreatePackage(ctx context.Context, pkg *PackageRegistry) (*PackageRegistry, error) {
	pkg.ID = uuid.New()
	pkg.CreatedAt = time.Now()
	pkg.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO package_registry (id, tenant_id, name, scope, description, registry_type, latest_version, total_downloads, is_internal, repository_url, published_by, published_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		pkg.ID, pkg.TenantID, pkg.Name, pkg.Scope, pkg.Description, pkg.RegistryType, pkg.LatestVersion, pkg.TotalDownloads, pkg.IsInternal, pkg.RepositoryURL, pkg.PublishedBy, pkg.PublishedAt, pkg.CreatedAt, pkg.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create package: %w", err)
	}
	return pkg, nil
}

// CreatePackageVersion creates a new version for a package.
func (r *RemainingRepository) CreatePackageVersion(ctx context.Context, v *PackageVersion) (*PackageVersion, error) {
	v.ID = uuid.New()
	v.PublishedAt = time.Now()

	var depsParam interface{}
	if v.Dependencies != nil {
		b, _ := json.Marshal(v.Dependencies)
		depsParam = b
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO package_versions (id, package_id, version, description, dependencies, published_by, downloads, tarball_url, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		v.ID, v.PackageID, v.Version, v.Description, depsParam, v.PublishedBy, v.Downloads, v.TarballURL, v.PublishedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create package version: %w", err)
	}
	return v, nil
}
