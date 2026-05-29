package studio

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrLicenseNotFound    = errors.New("license not found")
	ErrLicenseRevoked     = errors.New("license revoked")
	ErrLicenseExpired     = errors.New("license expired")
	ErrActivationLimit    = errors.New("activation limit reached")
	ErrInvalidLicenseKey  = errors.New("invalid license key")
	ErrInvalidLicenseType = errors.New("invalid license type")
	ErrInvalidSPDXLicense = errors.New("invalid spdx license")
)

// LicensePolicy represents SPDX/commercial policy for a marketplace function.
type LicensePolicy struct {
	FunctionID        string    `json:"function_id"`
	SPDXLicense       string    `json:"spdx_license"`
	CustomLicenseText *string   `json:"custom_license_text,omitempty"`
	CommercialType    string    `json:"commercial_type"`
	MaxActivations    *int      `json:"max_activations_default,omitempty"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// LicenseGrant represents an issued license key grant.
type LicenseGrant struct {
	ID              string `json:"id"`
	Key             string `json:"key,omitempty"`
	Type            string `json:"type"`
	FunctionID      string `json:"functionId"`
	FunctionName    string `json:"functionName"`
	PurchaserID     string `json:"purchaserId"`
	PurchaserName   string `json:"purchaserName"`
	IssuedAt        int64  `json:"issuedAt"`
	ExpiresAt       *int64 `json:"expiresAt,omitempty"`
	MaxActivations  *int   `json:"maxActivations,omitempty"`
	ActivationCount int    `json:"activationCount"`
	Revoked         bool   `json:"revoked"`
}

type ListLicenseGrantsParams struct {
	TenantID   string
	FunctionID *string
	Revoked    *bool
	Limit      int
	Offset     int
}

type CreateLicenseGrantParams struct {
	TenantID       string
	UserID         string
	FunctionID     string
	FunctionName   string
	LicenseType    string
	PurchaserName  string
	PurchaserID    string
	MaxActivations *int
	ExpiresAt      *time.Time
}

type ActivateLicenseParams struct {
	LicenseKey      string
	TenantID        string
	UserID          string
	ActivationLabel string
	IPAddress       string
}

func validSPDXLicense(license string) bool {
	switch strings.ToLower(strings.TrimSpace(license)) {
	case "mit", "apache", "gpl", "proprietary", "custom":
		return true
	default:
		return false
	}
}

func validCommercialType(licenseType string) bool {
	switch strings.ToLower(strings.TrimSpace(licenseType)) {
	case "open", "restricted", "commercial":
		return true
	default:
		return false
	}
}

func generateLicenseKey() (string, string, string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", "", "", fmt.Errorf("generate license key: %w", err)
	}
	key := "FFLIC-" + strings.ToUpper(hex.EncodeToString(buf))
	hash := hashLicenseKey(key)
	prefix := key
	if len(prefix) > 12 {
		prefix = prefix[:12]
	}
	return key, hash, prefix, nil
}

func hashLicenseKey(key string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(key)))
	return hex.EncodeToString(sum[:])
}

func maskLicenseKey(prefix string) string {
	if prefix == "" {
		return "FFLIC-****"
	}
	return prefix + "****"
}

func scanLicenseGrant(
	id, licenseType, functionID, functionName, purchaserID, purchaserName, prefix string,
	createdAt time.Time,
	expiresAt sql.NullTime,
	maxActivations sql.NullInt64,
	activationCount int,
	revokedAt sql.NullTime,
) LicenseGrant {
	grant := LicenseGrant{
		ID:              id,
		Type:            licenseType,
		FunctionID:      functionID,
		FunctionName:    functionName,
		PurchaserID:     purchaserID,
		PurchaserName:   purchaserName,
		IssuedAt:        createdAt.UTC().UnixMilli(),
		ActivationCount: activationCount,
		Key:             maskLicenseKey(prefix),
		Revoked:         revokedAt.Valid,
	}
	if expiresAt.Valid {
		ms := expiresAt.Time.UTC().UnixMilli()
		grant.ExpiresAt = &ms
	}
	if maxActivations.Valid {
		v := int(maxActivations.Int64)
		grant.MaxActivations = &v
	}
	return grant
}

func (r *MarketplaceRepository) GetLicensePolicy(ctx context.Context, tenantID, functionID string) (*LicensePolicy, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT function_id, spdx_license, custom_license_text, commercial_type, max_activations_default, updated_at
		FROM marketplace_function_license_policies
		WHERE tenant_id = $1 AND function_id = $2
	`, tenantID, functionID)

	var policy LicensePolicy
	var customText sql.NullString
	var maxActivations sql.NullInt64
	if err := row.Scan(
		&policy.FunctionID,
		&policy.SPDXLicense,
		&customText,
		&policy.CommercialType,
		&maxActivations,
		&policy.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &LicensePolicy{
				FunctionID:     functionID,
				SPDXLicense:    "mit",
				CommercialType: "open",
				UpdatedAt:      time.Now().UTC(),
			}, nil
		}
		return nil, err
	}

	if customText.Valid {
		policy.CustomLicenseText = &customText.String
	}
	if maxActivations.Valid {
		v := int(maxActivations.Int64)
		policy.MaxActivations = &v
	}
	return &policy, nil
}

func (r *MarketplaceRepository) UpsertLicensePolicy(
	ctx context.Context,
	tenantID, functionID, spdxLicense string,
	customText *string,
	commercialType string,
	maxActivations *int,
) (*LicensePolicy, error) {
	if !validSPDXLicense(spdxLicense) {
		return nil, ErrInvalidSPDXLicense
	}
	if commercialType == "" {
		commercialType = "open"
	}
	if !validCommercialType(commercialType) {
		return nil, ErrInvalidLicenseType
	}

	var custom sql.NullString
	if customText != nil && strings.TrimSpace(*customText) != "" {
		custom = sql.NullString{String: strings.TrimSpace(*customText), Valid: true}
	}
	var maxAct sql.NullInt64
	if maxActivations != nil {
		maxAct = sql.NullInt64{Int64: int64(*maxActivations), Valid: true}
	}

	row := r.db.QueryRowContext(ctx, `
		INSERT INTO marketplace_function_license_policies (
			tenant_id, function_id, spdx_license, custom_license_text, commercial_type, max_activations_default
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (tenant_id, function_id) DO UPDATE SET
			spdx_license = EXCLUDED.spdx_license,
			custom_license_text = EXCLUDED.custom_license_text,
			commercial_type = EXCLUDED.commercial_type,
			max_activations_default = EXCLUDED.max_activations_default,
			updated_at = NOW()
		RETURNING function_id, spdx_license, custom_license_text, commercial_type, max_activations_default, updated_at
	`, tenantID, functionID, strings.ToLower(spdxLicense), custom, commercialType, maxAct)

	policy := &LicensePolicy{}
	var customOut sql.NullString
	var maxOut sql.NullInt64
	if err := row.Scan(
		&policy.FunctionID,
		&policy.SPDXLicense,
		&customOut,
		&policy.CommercialType,
		&maxOut,
		&policy.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if customOut.Valid {
		policy.CustomLicenseText = &customOut.String
	}
	if maxOut.Valid {
		v := int(maxOut.Int64)
		policy.MaxActivations = &v
	}
	return policy, nil
}

func (r *MarketplaceRepository) UpdateLicense(ctx context.Context, tenantID, functionID, license string) error {
	_, err := r.UpsertLicensePolicy(ctx, tenantID, functionID, license, nil, "open", nil)
	return err
}

func (r *MarketplaceRepository) ListLicenseGrants(ctx context.Context, params ListLicenseGrantsParams) ([]LicenseGrant, int, int, error) {
	if params.Limit <= 0 {
		params.Limit = 50
	}
	if params.Limit > 200 {
		params.Limit = 200
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIdx))
	args = append(args, params.TenantID)
	argIdx++

	if params.FunctionID != nil && *params.FunctionID != "" {
		conditions = append(conditions, fmt.Sprintf("function_id = $%d", argIdx))
		args = append(args, *params.FunctionID)
		argIdx++
	}
	if params.Revoked != nil {
		if *params.Revoked {
			conditions = append(conditions, "revoked_at IS NOT NULL")
		} else {
			conditions = append(conditions, "revoked_at IS NULL")
		}
	}

	where := strings.Join(conditions, " AND ")

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*),
			COUNT(*) FILTER (WHERE revoked_at IS NULL),
			COUNT(*) FILTER (WHERE revoked_at IS NOT NULL)
		FROM marketplace_license_grants WHERE %s`, where)

	var total, active, revoked int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total, &active, &revoked); err != nil {
		return nil, 0, 0, err
	}

	listQuery := fmt.Sprintf(`
		SELECT id, license_type, function_id, function_name,
			COALESCE(purchaser_user_id::text, purchaser_tenant_id::text, '') AS purchaser_id,
			purchaser_name, created_at, expires_at, max_activations, activation_count,
			revoked_at, license_key_prefix
		FROM marketplace_license_grants
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, params.Limit, params.Offset)
	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	grants := make([]LicenseGrant, 0)
	for rows.Next() {
		var (
			id, licenseType, functionID, functionName, purchaserID, purchaserName, prefix string
			createdAt                                                                      time.Time
			expires                                                                        sql.NullTime
			maxActivations                                                                 sql.NullInt64
			revokedAt                                                                      sql.NullTime
			activationCount                                                                int
		)
		if err := rows.Scan(
			&id, &licenseType, &functionID, &functionName, &purchaserID, &purchaserName,
			&createdAt, &expires, &maxActivations, &activationCount, &revokedAt, &prefix,
		); err != nil {
			return nil, 0, 0, err
		}
		grants = append(grants, scanLicenseGrant(
			id, licenseType, functionID, functionName, purchaserID, purchaserName, prefix,
			createdAt, expires, maxActivations, activationCount, revokedAt,
		))
	}

	return grants, active, revoked, rows.Err()
}

func (r *MarketplaceRepository) CreateLicenseGrant(ctx context.Context, params CreateLicenseGrantParams) (*LicenseGrant, error) {
	if params.FunctionID == "" {
		return nil, errors.New("function_id is required")
	}
	if params.LicenseType == "" {
		params.LicenseType = "commercial"
	}
	if !validCommercialType(params.LicenseType) {
		return nil, ErrInvalidLicenseType
	}
	if params.PurchaserName == "" {
		params.PurchaserName = "Unassigned"
	}

	key, hash, prefix, err := generateLicenseKey()
	if err != nil {
		return nil, err
	}

	var purchaserUser sql.NullString
	if params.PurchaserID != "" {
		if parsed, parseErr := uuid.Parse(params.PurchaserID); parseErr == nil {
			purchaserUser = sql.NullString{String: parsed.String(), Valid: true}
		}
	}

	var expires sql.NullTime
	if params.ExpiresAt != nil {
		expires = sql.NullTime{Time: params.ExpiresAt.UTC(), Valid: true}
	}
	var maxAct sql.NullInt64
	if params.MaxActivations != nil {
		maxAct = sql.NullInt64{Int64: int64(*params.MaxActivations), Valid: true}
	}

	row := r.db.QueryRowContext(ctx, `
		INSERT INTO marketplace_license_grants (
			tenant_id, function_id, function_name, license_key_hash, license_key_prefix,
			license_type, purchaser_user_id, purchaser_name, issued_by_user_id,
			expires_at, max_activations
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, license_type, function_id, function_name,
			COALESCE(purchaser_user_id::text, '') AS purchaser_id,
			purchaser_name, created_at, expires_at, max_activations, activation_count, license_key_prefix
	`, params.TenantID, params.FunctionID, params.FunctionName, hash, prefix,
		params.LicenseType, purchaserUser, params.PurchaserName, params.UserID, expires, maxAct)

	var (
		grant      LicenseGrant
		expiresOut sql.NullTime
		maxOut     sql.NullInt64
		createdAt  time.Time
	)
	if err := row.Scan(
		&grant.ID, &grant.Type, &grant.FunctionID, &grant.FunctionName,
		&grant.PurchaserID, &grant.PurchaserName, &createdAt,
		&expiresOut, &maxOut, &grant.ActivationCount, &prefix,
	); err != nil {
		return nil, err
	}

	grant.Key = key
	grant.IssuedAt = createdAt.UTC().UnixMilli()
	if expiresOut.Valid {
		ms := expiresOut.Time.UTC().UnixMilli()
		grant.ExpiresAt = &ms
	}
	if maxOut.Valid {
		v := int(maxOut.Int64)
		grant.MaxActivations = &v
	}
	return &grant, nil
}

func (r *MarketplaceRepository) RevokeLicenseGrant(ctx context.Context, tenantID, userID, grantID string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE marketplace_license_grants
		SET revoked_at = NOW(), revoked_by_user_id = $1, updated_at = NOW()
		WHERE id = $2 AND tenant_id = $3 AND revoked_at IS NULL
	`, userID, grantID, tenantID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrLicenseNotFound
	}
	return nil
}

func (r *MarketplaceRepository) ActivateLicenseGrant(ctx context.Context, params ActivateLicenseParams) (*LicenseGrant, error) {
	hash := hashLicenseKey(params.LicenseKey)

	row := r.db.QueryRowContext(ctx, `
		SELECT id, license_type, function_id, function_name,
			COALESCE(purchaser_user_id::text, '') AS purchaser_id,
			purchaser_name, created_at, expires_at, max_activations, activation_count, revoked_at, license_key_prefix
		FROM marketplace_license_grants
		WHERE license_key_hash = $1
	`, hash)

	var (
		grant            LicenseGrant
		expiresAt        sql.NullTime
		maxActivations   sql.NullInt64
		revokedAt        sql.NullTime
		createdAt        time.Time
		prefix           string
	)
	if err := row.Scan(
		&grant.ID, &grant.Type, &grant.FunctionID, &grant.FunctionName,
		&grant.PurchaserID, &grant.PurchaserName, &createdAt,
		&expiresAt, &maxActivations, &grant.ActivationCount, &revokedAt, &prefix,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidLicenseKey
		}
		return nil, err
	}

	if revokedAt.Valid {
		return nil, ErrLicenseRevoked
	}
	if expiresAt.Valid && time.Now().UTC().After(expiresAt.Time.UTC()) {
		return nil, ErrLicenseExpired
	}
	if maxActivations.Valid && grant.ActivationCount >= int(maxActivations.Int64) {
		return nil, ErrActivationLimit
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE marketplace_license_grants
		SET activation_count = activation_count + 1, updated_at = NOW()
		WHERE id = $1
	`, grant.ID); err != nil {
		return nil, err
	}

	var userID sql.NullString
	if params.UserID != "" {
		userID = sql.NullString{String: params.UserID, Valid: true}
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO marketplace_license_activations (grant_id, tenant_id, user_id, activation_label, ip_address)
		VALUES ($1, $2, $3, $4, NULLIF($5, '')::inet)
	`, grant.ID, params.TenantID, userID, params.ActivationLabel, params.IPAddress); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	grant.ActivationCount++
	result := scanLicenseGrant(
		grant.ID, grant.Type, grant.FunctionID, grant.FunctionName,
		grant.PurchaserID, grant.PurchaserName, prefix,
		createdAt, expiresAt, maxActivations, grant.ActivationCount, revokedAt,
	)
	return &result, nil
}
