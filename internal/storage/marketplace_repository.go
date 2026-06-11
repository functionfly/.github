package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ExtensionUpdate struct {
	InstalledPluginID string                 `json:"installed_plugin_id"`
	InstalledVersion  string                 `json:"installed_version"`
	ExtensionID       string                 `json:"extension_id"`
	LatestVersion     string                 `json:"latest_version"`
	Changelog         string                 `json:"changelog"`
	Manifest          map[string]interface{} `json:"manifest"`
}

type InstalledPlugin struct {
	ID      string
	Name    string
	Version string
}

func (r *MarketplaceRepository) FindUpdates(ctx context.Context, installed []InstalledPlugin) ([]ExtensionUpdate, error) {
	if len(installed) == 0 {
		return []ExtensionUpdate{}, nil
	}

	names := make([]string, len(installed))
	for i, p := range installed {
		names[i] = p.Name
	}

	query := `
		SELECT id, name, version, COALESCE(changelog, ''), manifest
		FROM marketplace_extensions
		WHERE name = ANY($1) AND status = 'published'
	`
	rows, err := r.db.QueryContext(ctx, query, names)
	if err != nil {
		return nil, fmt.Errorf("find updates: %w", err)
	}
	defer rows.Close()

	updates := []ExtensionUpdate{}
	installedByName := make(map[string]InstalledPlugin)
	for _, p := range installed {
		installedByName[p.Name] = p
	}

	for rows.Next() {
		var extID, name, version, changelog string
		var manifest []byte
		if err := rows.Scan(&extID, &name, &version, &changelog, &manifest); err != nil {
			return nil, fmt.Errorf("scan update: %w", err)
		}

		inst, ok := installedByName[name]
		if !ok {
			continue
		}

		if isNewerVersion(version, inst.Version) {
			var manifestMap map[string]interface{}
			if len(manifest) > 0 {
				_ = json.Unmarshal(manifest, &manifestMap)
			}
			updates = append(updates, ExtensionUpdate{
				InstalledPluginID: inst.ID,
				InstalledVersion:  inst.Version,
				ExtensionID:       extID,
				LatestVersion:     version,
				Changelog:         changelog,
				Manifest:          manifestMap,
			})
		}
	}

	return updates, rows.Err()
}

func isNewerVersion(latest, current string) bool {
	if latest == current {
		return false
	}
	l := parseSemver(latest)
	c := parseSemver(current)
	if l == nil || c == nil {
		return latest != current
	}
	if l[0] != c[0] {
		return l[0] > c[0]
	}
	if l[1] != c[1] {
		return l[1] > c[1]
	}
	return l[2] > c[2]
}

func parseSemver(v string) []int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	result := []int{0, 0, 0}
	for i := 0; i < len(parts) && i < 3; i++ {
		n := 0
		for _, ch := range parts[i] {
			if ch < '0' || ch > '9' {
				break
			}
			n = n*10 + int(ch-'0')
		}
		result[i] = n
	}
	return result
}

type MarketplaceExtension struct {
	ID             string                 `json:"id"`
	CreatorID      string                 `json:"creator_id"`
	PluginID       *string                `json:"plugin_id"`
	Name           string                 `json:"name"`
	Version        string                 `json:"version"`
	Description    string                 `json:"description"`
	Category       string                 `json:"category"`
	IconURL        string                 `json:"icon_url"`
	Screenshots    []string               `json:"screenshots"`
	Manifest       []byte                 `json:"-"`
	ManifestParsed map[string]interface{} `json:"manifest"`
	ManifestURL    string                 `json:"manifest_url"`
	Signature      string                 `json:"signature"`
	Verified       bool                   `json:"verified"`
	Status         string                 `json:"status"`
	Featured       bool                   `json:"featured"`
	InstallCount   int                    `json:"install_count"`
	RatingAverage  float64                `json:"rating_average"`
	RatingCount    int                    `json:"rating_count"`
	TrustScore     float64                `json:"trust_score"`
	SandboxScore   float64                `json:"sandbox_score"`
	SecurityScore  float64                `json:"security_score"`
	RuntimeScore   float64                `json:"runtime_score"`
	Compatibility  map[string]interface{} `json:"compatibility"`
	Tags           []string               `json:"tags"`
	Changelog      string                 `json:"changelog"`
	ReleaseNotes   string                 `json:"release_notes"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	PublishedAt    *time.Time             `json:"published_at"`
	UnpublishedAt  *time.Time             `json:"unpublished_at"`
}

type MarketplaceRepository struct {
	db *sql.DB
}

func NewMarketplaceRepository(db *sql.DB) *MarketplaceRepository {
	return &MarketplaceRepository{db: db}
}

type ListMarketplaceParams struct {
	CreatorID *string
	Category  *string
	Status    *string
	Featured  *bool
	Search    *string
	Tags      []string
	SortBy    string
	Limit     int
	Offset    int
}

func (r *MarketplaceRepository) List(ctx context.Context, params ListMarketplaceParams) ([]MarketplaceExtension, error) {
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

	if params.CreatorID != nil {
		conditions = append(conditions, fmt.Sprintf("creator_id = $%d", argIdx))
		args = append(args, *params.CreatorID)
		argIdx++
	}

	if params.Category != nil && *params.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *params.Category)
		argIdx++
	}

	if params.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *params.Status)
		argIdx++
	} else {
		conditions = append(conditions, "status = 'published'")
	}

	if params.Featured != nil && *params.Featured {
		conditions = append(conditions, "featured = true")
	}

	if params.Search != nil && *params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d OR $%d = ANY(tags))", argIdx, argIdx, argIdx))
		args = append(args, "%"+*params.Search+"%")
		argIdx++
	}

	if len(params.Tags) > 0 {
		for _, tag := range params.Tags {
			conditions = append(conditions, fmt.Sprintf("$%d = ANY(tags)", argIdx))
			args = append(args, tag)
			argIdx++
		}
	}

	conditionsStr := strings.Join(conditions, " AND ")
	if conditionsStr != "" {
		conditionsStr = "WHERE " + conditionsStr
	}

	orderBy := "ORDER BY featured DESC, install_count DESC, trust_score DESC"
	switch params.SortBy {
	case "newest":
		orderBy = "ORDER BY COALESCE(published_at, created_at) DESC"
	case "top_rated":
		orderBy = "ORDER BY rating_average DESC, rating_count DESC"
	case "most_installed":
		orderBy = "ORDER BY install_count DESC"
	case "trending":
		orderBy = "ORDER BY (install_count * 0.3 + rating_average * 100 * 0.7) DESC"
	}

	query := fmt.Sprintf(`
		SELECT id, creator_id, plugin_id, name, version, description, category, icon_url,
		       screenshots, manifest, manifest_url, signature, verified, status, featured,
		       install_count, rating_average, rating_count, trust_score, sandbox_score,
		       security_score, runtime_score, compatibility, tags, changelog, release_notes,
		       created_at, updated_at, published_at, unpublished_at
		FROM marketplace_extensions
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, conditionsStr, orderBy, argIdx, argIdx+1)
	args = append(args, params.Limit, params.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list marketplace extensions: %w", err)
	}
	defer rows.Close()

	var extensions []MarketplaceExtension
	for rows.Next() {
		ext, err := scanMarketplaceExtension(rows)
		if err != nil {
			return nil, fmt.Errorf("scan extension: %w", err)
		}
		extensions = append(extensions, *ext)
	}

	return extensions, rows.Err()
}

func (r *MarketplaceRepository) Get(ctx context.Context, id string) (*MarketplaceExtension, error) {
	query := `
		SELECT id, creator_id, plugin_id, name, version, description, category, icon_url,
		       screenshots, manifest, manifest_url, signature, verified, status, featured,
		       install_count, rating_average, rating_count, trust_score, sandbox_score,
		       security_score, runtime_score, compatibility, tags, changelog, release_notes,
		       created_at, updated_at, published_at, unpublished_at
		FROM marketplace_extensions
		WHERE id = $1
	`
	var ext MarketplaceExtension
	var screenshots, tags, compatibility []byte
	var pluginID, iconURL, manifestURL, signature, changelog, releaseNotes sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ext.ID, &ext.CreatorID, &pluginID, &ext.Name, &ext.Version, &ext.Description,
		&ext.Category, &iconURL, &screenshots, &ext.Manifest, &manifestURL, &signature,
		&ext.Verified, &ext.Status, &ext.Featured, &ext.InstallCount, &ext.RatingAverage,
		&ext.RatingCount, &ext.TrustScore, &ext.SandboxScore, &ext.SecurityScore, &ext.RuntimeScore,
		&compatibility, &tags, &changelog, &releaseNotes, &ext.CreatedAt, &ext.UpdatedAt,
		&ext.PublishedAt, &ext.UnpublishedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get marketplace extension: %w", err)
	}

	if pluginID.Valid {
		ext.PluginID = &pluginID.String
	}
	if iconURL.Valid {
		ext.IconURL = iconURL.String
	}
	if len(screenshots) > 0 {
		_ = json.Unmarshal(screenshots, &ext.Screenshots)
	}
	if manifestURL.Valid {
		ext.ManifestURL = manifestURL.String
	}
	if signature.Valid {
		ext.Signature = signature.String
	}
	if len(compatibility) > 0 {
		_ = json.Unmarshal(compatibility, &ext.Compatibility)
	}
	if len(tags) > 0 {
		_ = json.Unmarshal(tags, &ext.Tags)
	}
	if changelog.Valid {
		ext.Changelog = changelog.String
	}
	if releaseNotes.Valid {
		ext.ReleaseNotes = releaseNotes.String
	}

	return &ext, nil
}

func (r *MarketplaceRepository) Create(ctx context.Context, ext *MarketplaceExtension) error {
	if ext.ID == "" {
		ext.ID = uuid.New().String()
	}

	screenshotsRaw, _ := json.Marshal(ext.Screenshots)
	manifestRaw, _ := json.Marshal(ext.Manifest)
	compatibilityRaw, _ := json.Marshal(ext.Compatibility)
	tagsRaw, _ := json.Marshal(ext.Tags)
	now := time.Now()

	query := `
		INSERT INTO marketplace_extensions (id, creator_id, plugin_id, name, version, description,
		                                     category, icon_url, screenshots, manifest, manifest_url,
		                                     signature, verified, status, featured, install_count,
		                                     rating_average, rating_count, trust_score, sandbox_score,
		                                     security_score, runtime_score, compatibility, tags,
		                                     changelog, release_notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		ext.ID, ext.CreatorID, ext.PluginID, ext.Name, ext.Version, ext.Description,
		ext.Category, ext.IconURL, screenshotsRaw, manifestRaw, ext.ManifestURL, ext.Signature,
		ext.Verified, ext.Status, ext.Featured, ext.InstallCount, ext.RatingAverage,
		ext.RatingCount, ext.TrustScore, ext.SandboxScore, ext.SecurityScore, ext.RuntimeScore,
		compatibilityRaw, tagsRaw, ext.Changelog, ext.ReleaseNotes, now, now,
	).Scan(&ext.CreatedAt, &ext.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create marketplace extension: %w", err)
	}

	return nil
}

func (r *MarketplaceRepository) Update(ctx context.Context, ext *MarketplaceExtension) error {
	screenshotsRaw, _ := json.Marshal(ext.Screenshots)
	manifestRaw, _ := json.Marshal(ext.Manifest)
	compatibilityRaw, _ := json.Marshal(ext.Compatibility)
	tagsRaw, _ := json.Marshal(ext.Tags)
	now := time.Now()

	query := `
		UPDATE marketplace_extensions SET
			name = $1, version = $2, description = $3, category = $4, icon_url = $5,
			screenshots = $6, manifest = $7, manifest_url = $8, signature = $9, verified = $10,
			status = $11, featured = $12, tags = $13, changelog = $14, release_notes = $15,
			updated_at = $16, compatibility = $17
		WHERE id = $18
	`

	_, err := r.db.ExecContext(ctx, query,
		ext.Name, ext.Version, ext.Description, ext.Category, ext.IconURL,
		screenshotsRaw, manifestRaw, ext.ManifestURL, ext.Signature, ext.Verified,
		ext.Status, ext.Featured, tagsRaw, ext.Changelog, ext.ReleaseNotes, now, ext.ID,
		compatibilityRaw,
	)
	if err != nil {
		return fmt.Errorf("update marketplace extension: %w", err)
	}

	ext.UpdatedAt = now
	return nil
}

func (r *MarketplaceRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM marketplace_extensions WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete marketplace extension: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("extension not found")
	}

	return nil
}

func (r *MarketplaceRepository) IncrementInstallCount(ctx context.Context, id string) error {
	query := `UPDATE marketplace_extensions SET install_count = install_count + 1 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *MarketplaceRepository) GetInstallCounts(ctx context.Context, ids []string) (map[string]int, error) {
	if len(ids) == 0 {
		return make(map[string]int), nil
	}

	query := `
		SELECT id, install_count FROM marketplace_extensions
		WHERE id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("get install counts: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var id string
		var count int
		if err := rows.Scan(&id, &count); err != nil {
			return nil, fmt.Errorf("scan install count: %w", err)
		}
		counts[id] = count
	}
	return counts, rows.Err()
}

func (r *MarketplaceRepository) GetCategories(ctx context.Context) ([]struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}, error) {
	query := `
		SELECT category, COUNT(*) as count
		FROM marketplace_extensions
		WHERE status = 'published'
		GROUP BY category
		ORDER BY count DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		Category string `json:"category"`
		Count    int    `json:"count"`
	}
	for rows.Next() {
		var result struct {
			Category string `json:"category"`
			Count    int    `json:"count"`
		}
		if err := rows.Scan(&result.Category, &result.Count); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

type MarketplaceRating struct {
	ID          string    `json:"id"`
	ExtensionID string    `json:"extension_id"`
	TenantID    string    `json:"tenant_id"`
	Rating      int       `json:"rating"`
	Review      string    `json:"review,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (r *MarketplaceRepository) UpsertRating(ctx context.Context, rating *MarketplaceRating) error {
	if rating.ID == "" {
		rating.ID = uuid.New().String()
	}
	if rating.CreatedAt.IsZero() {
		rating.CreatedAt = time.Now()
	}
	rating.UpdatedAt = time.Now()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		INSERT INTO marketplace_ratings (id, extension_id, tenant_id, rating, review, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (extension_id, tenant_id) DO UPDATE SET
			rating = EXCLUDED.rating,
			review = EXCLUDED.review,
			updated_at = EXCLUDED.updated_at
	`
	_, err = tx.ExecContext(ctx, query,
		rating.ID, rating.ExtensionID, rating.TenantID, rating.Rating,
		rating.Review, rating.CreatedAt, rating.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert rating: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE marketplace_extensions SET
			rating_average = (SELECT AVG(rating)::float FROM marketplace_ratings WHERE extension_id = $1),
			rating_count = (SELECT COUNT(*) FROM marketplace_ratings WHERE extension_id = $1)
		WHERE id = $1
	`, rating.ExtensionID)
	if err != nil {
		return fmt.Errorf("recompute rating aggregate: %w", err)
	}

	return tx.Commit()
}

func (r *MarketplaceRepository) GetRating(ctx context.Context, extensionID, tenantID string) (*MarketplaceRating, error) {
	query := `
		SELECT id, extension_id, tenant_id, rating, COALESCE(review, ''), created_at, updated_at
		FROM marketplace_ratings
		WHERE extension_id = $1 AND tenant_id = $2
	`
	var rating MarketplaceRating
	err := r.db.QueryRowContext(ctx, query, extensionID, tenantID).Scan(
		&rating.ID, &rating.ExtensionID, &rating.TenantID, &rating.Rating,
		&rating.Review, &rating.CreatedAt, &rating.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get rating: %w", err)
	}
	return &rating, nil
}

func (r *MarketplaceRepository) ListRatings(ctx context.Context, extensionID string, limit int) ([]MarketplaceRating, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query := `
		SELECT id, extension_id, tenant_id, rating, COALESCE(review, ''), created_at, updated_at
		FROM marketplace_ratings
		WHERE extension_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, extensionID, limit)
	if err != nil {
		return nil, fmt.Errorf("list ratings: %w", err)
	}
	defer rows.Close()

	var ratings []MarketplaceRating
	for rows.Next() {
		var rating MarketplaceRating
		if err := rows.Scan(
			&rating.ID, &rating.ExtensionID, &rating.TenantID, &rating.Rating,
			&rating.Review, &rating.CreatedAt, &rating.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan rating: %w", err)
		}
		ratings = append(ratings, rating)
	}
	return ratings, rows.Err()
}

func scanMarketplaceExtension(rows interface {
	Scan(dst ...interface{}) error
}) (*MarketplaceExtension, error) {
	var ext MarketplaceExtension
	var screenshots, tags, compatibility []byte
	var pluginID, iconURL, manifestURL, signature, changelog, releaseNotes sql.NullString

	err := rows.Scan(
		&ext.ID, &ext.CreatorID, &pluginID, &ext.Name, &ext.Version, &ext.Description,
		&ext.Category, &iconURL, &screenshots, &ext.Manifest, &manifestURL, &signature,
		&ext.Verified, &ext.Status, &ext.Featured, &ext.InstallCount, &ext.RatingAverage,
		&ext.RatingCount, &ext.TrustScore, &ext.SandboxScore, &ext.SecurityScore, &ext.RuntimeScore,
		&compatibility, &tags, &changelog, &releaseNotes, &ext.CreatedAt, &ext.UpdatedAt,
		&ext.PublishedAt, &ext.UnpublishedAt,
	)
	if err != nil {
		return nil, err
	}

	if pluginID.Valid {
		ext.PluginID = &pluginID.String
	}
	if iconURL.Valid {
		ext.IconURL = iconURL.String
	}
	if len(screenshots) > 0 {
		_ = json.Unmarshal(screenshots, &ext.Screenshots)
	}
	if len(ext.Manifest) > 0 {
		_ = json.Unmarshal(ext.Manifest, &ext.ManifestParsed)
	}
	if manifestURL.Valid {
		ext.ManifestURL = manifestURL.String
	}
	if signature.Valid {
		ext.Signature = signature.String
	}
	if len(compatibility) > 0 {
		_ = json.Unmarshal(compatibility, &ext.Compatibility)
	}
	if len(tags) > 0 {
		_ = json.Unmarshal(tags, &ext.Tags)
	}
	if changelog.Valid {
		ext.Changelog = changelog.String
	}
	if releaseNotes.Valid {
		ext.ReleaseNotes = releaseNotes.String
	}

	return &ext, nil
}
