package marketplace

import (
	"context"
	"encoding/json"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

type StorageAdapter struct {
	repo       *storage.MarketplaceRepository
	pluginRepo *storage.PluginRepository
}

func NewStorageAdapter(repo *storage.MarketplaceRepository) *StorageAdapter {
	return &StorageAdapter{repo: repo}
}

func NewStorageAdapterWithPlugins(marketplaceRepo *storage.MarketplaceRepository, pluginRepo *storage.PluginRepository) *StorageAdapter {
	return &StorageAdapter{repo: marketplaceRepo, pluginRepo: pluginRepo}
}

func (a *StorageAdapter) List(ctx context.Context, params ListParams) ([]Extension, error) {
	listParams := storage.ListMarketplaceParams{
		CreatorID: params.CreatorID,
		Category:  params.Category,
		Status:    params.Status,
		Featured:  params.Featured,
		Search:    params.Search,
		Tags:      params.Tags,
		SortBy:    params.SortBy,
		Limit:     params.Limit,
		Offset:    params.Offset,
	}

	extensions, err := a.repo.List(ctx, listParams)
	if err != nil {
		return nil, err
	}

	result := make([]Extension, len(extensions))
	for i, ext := range extensions {
		result[i] = toHandlerExtension(ext)
	}
	return result, nil
}

func (a *StorageAdapter) Get(ctx context.Context, id string) (*Extension, error) {
	ext, err := a.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if ext == nil {
		return nil, nil
	}
	result := toHandlerExtension(*ext)
	return &result, nil
}

func (a *StorageAdapter) Create(ctx context.Context, ext *Extension) error {
	storageExt := toStorageExtension(ext)
	return a.repo.Create(ctx, storageExt)
}

func (a *StorageAdapter) Update(ctx context.Context, ext *Extension) error {
	storageExt := toStorageExtension(ext)
	return a.repo.Update(ctx, storageExt)
}

func (a *StorageAdapter) Delete(ctx context.Context, id string) error {
	return a.repo.Delete(ctx, id)
}

func (a *StorageAdapter) IncrementInstallCount(ctx context.Context, id string) error {
	return a.repo.IncrementInstallCount(ctx, id)
}

func (a *StorageAdapter) GetInstallCounts(ctx context.Context, ids []string) (map[string]int, error) {
	return a.repo.GetInstallCounts(ctx, ids)
}

func (a *StorageAdapter) GetCategories(ctx context.Context) ([]CategoryCount, error) {
	cats, err := a.repo.GetCategories(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]CategoryCount, len(cats))
	for i, c := range cats {
		result[i] = CategoryCount{Category: c.Category, Count: c.Count}
	}
	return result, nil
}

func (a *StorageAdapter) UpsertRating(ctx context.Context, rating *Rating) error {
	r := &storage.MarketplaceRating{
		ID:          rating.ID,
		ExtensionID: rating.ExtensionID,
		TenantID:    rating.TenantID,
		Rating:      rating.Rating,
		Review:      rating.Review,
		CreatedAt:   rating.CreatedAt,
		UpdatedAt:   rating.UpdatedAt,
	}
	return a.repo.UpsertRating(ctx, r)
}

func (a *StorageAdapter) GetRating(ctx context.Context, extensionID, tenantID string) (*Rating, error) {
	r, err := a.repo.GetRating(ctx, extensionID, tenantID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return &Rating{
		ID:          r.ID,
		ExtensionID: r.ExtensionID,
		TenantID:    r.TenantID,
		Rating:      r.Rating,
		Review:      r.Review,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}, nil
}

func (a *StorageAdapter) ListRatings(ctx context.Context, extensionID string, limit int) ([]Rating, error) {
	rows, err := a.repo.ListRatings(ctx, extensionID, limit)
	if err != nil {
		return nil, err
	}
	result := make([]Rating, len(rows))
	for i, r := range rows {
		result[i] = Rating{
			ID:          r.ID,
			ExtensionID: r.ExtensionID,
			TenantID:    r.TenantID,
			Rating:      r.Rating,
			Review:      r.Review,
			Username:    r.Username,
			UserName:    r.UserName,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		}
	}
	return result, nil
}

func (a *StorageAdapter) FindUpdates(ctx context.Context, installed []InstalledPlugin) ([]ExtensionUpdate, error) {
	storageInstalled := make([]storage.InstalledPlugin, len(installed))
	for i, p := range installed {
		storageInstalled[i] = storage.InstalledPlugin{
			ID:      p.ID,
			Name:    p.Name,
			Version: p.Version,
		}
	}
	updates, err := a.repo.FindUpdates(ctx, storageInstalled)
	if err != nil {
		return nil, err
	}
	result := make([]ExtensionUpdate, len(updates))
	for i, u := range updates {
		result[i] = ExtensionUpdate{
			InstalledPluginID: u.InstalledPluginID,
			InstalledVersion:  u.InstalledVersion,
			ExtensionID:       u.ExtensionID,
			LatestVersion:     u.LatestVersion,
			Changelog:         u.Changelog,
			Manifest:          u.Manifest,
		}
	}
	return result, nil
}

func (a *StorageAdapter) CreatePluginFromExtension(ctx context.Context, tenantID, extensionID string) (*Extension, error) {
	ext, err := a.repo.Get(ctx, extensionID)
	if err != nil {
		return nil, err
	}
	if ext == nil {
		return nil, nil
	}

	_ = a.repo.IncrementInstallCount(ctx, extensionID)

	if a.pluginRepo == nil {
		return toHandlerExtensionPtr(*ext), nil
	}

	pluginType := ext.Category
	if pluginType == "" {
		pluginType = "marketplace"
	}
	if pluginType != "ui" && pluginType != "graph" && pluginType != "ai_tool" &&
		pluginType != "runtime" && pluginType != "infrastructure" && pluginType != "marketplace" {
		pluginType = "marketplace"
	}

	manifestMap := ext.ManifestParsed
	if manifestMap == nil && len(ext.Manifest) > 0 {
		_ = json.Unmarshal(ext.Manifest, &manifestMap)
	}
	if manifestMap == nil {
		manifestMap = map[string]interface{}{"name": ext.Name, "version": ext.Version}
	}

	plugin := &storage.Plugin{
		TenantID:      tenantID,
		Manifest:      manifestMap,
		PluginType:    storage.PluginType(pluginType),
		Name:          ext.Name,
		Version:       ext.Version,
		Description:   ext.Description,
		AuthorName:    ext.CreatorID,
		Category:      ext.Category,
		Status:        storage.PluginStatusDisabled,
		IconURL:       ext.IconURL,
		RepositoryURL: "",
		License:       "",
		SizeBytes:     0,
		Signature:     ext.Signature,
		Verified:      ext.Verified,
		Config:        map[string]string{},
		Metadata:      map[string]interface{}{"extension_id": ext.ID},
	}

	if err := a.pluginRepo.Create(ctx, plugin); err != nil {
		logrus.WithError(err).Warn("marketplace: failed to create plugin from extension (will retry upsert)")
	}

	ext.PluginID = &plugin.ID
	if err := a.repo.Update(ctx, ext); err != nil {
		logrus.WithError(err).Warn("marketplace: failed to link extension to plugin")
	}

	updated, err := a.repo.Get(ctx, extensionID)
	if err != nil {
		return toHandlerExtensionPtr(*ext), nil
	}
	return toHandlerExtensionPtr(*updated), nil
}

func toHandlerExtensionPtr(ext storage.MarketplaceExtension) *Extension {
	result := toHandlerExtension(ext)
	return &result
}

func toHandlerExtension(ext storage.MarketplaceExtension) Extension {
	manifestMap := make(map[string]interface{})
	if len(ext.Manifest) > 0 {
		_ = json.Unmarshal(ext.Manifest, &manifestMap)
	}
	return Extension{
		ID:            ext.ID,
		CreatorID:     ext.CreatorID,
		PluginID:      ext.PluginID,
		Name:          ext.Name,
		Version:       ext.Version,
		Description:   ext.Description,
		Category:      ext.Category,
		IconURL:       ext.IconURL,
		Screenshots:   ext.Screenshots,
		Manifest:      manifestMap,
		ManifestURL:   ext.ManifestURL,
		Signature:     ext.Signature,
		Verified:      ext.Verified,
		Status:        ext.Status,
		Featured:      ext.Featured,
		InstallCount:  ext.InstallCount,
		RatingAverage: ext.RatingAverage,
		RatingCount:   ext.RatingCount,
		TrustScore:    ext.TrustScore,
		SandboxScore:  ext.SandboxScore,
		SecurityScore: ext.SecurityScore,
		RuntimeScore:  ext.RuntimeScore,
		Compatibility: ext.Compatibility,
		Tags:          ext.Tags,
		Changelog:     ext.Changelog,
		ReleaseNotes:  ext.ReleaseNotes,
		CreatedAt:     ext.CreatedAt,
		UpdatedAt:     ext.UpdatedAt,
		PublishedAt:   ext.PublishedAt,
	}
}

func toStorageExtension(ext *Extension) *storage.MarketplaceExtension {
	manifestBytes, _ := json.Marshal(ext.Manifest)
	return &storage.MarketplaceExtension{
		ID:            ext.ID,
		CreatorID:     ext.CreatorID,
		PluginID:      ext.PluginID,
		Name:          ext.Name,
		Version:       ext.Version,
		Description:   ext.Description,
		Category:      ext.Category,
		IconURL:       ext.IconURL,
		Screenshots:   ext.Screenshots,
		Manifest:      manifestBytes,
		ManifestURL:   ext.ManifestURL,
		Signature:     ext.Signature,
		Verified:      ext.Verified,
		Status:        ext.Status,
		Featured:      ext.Featured,
		InstallCount:  ext.InstallCount,
		RatingAverage: ext.RatingAverage,
		RatingCount:   ext.RatingCount,
		TrustScore:    ext.TrustScore,
		SandboxScore:  ext.SandboxScore,
		SecurityScore: ext.SecurityScore,
		RuntimeScore:  ext.RuntimeScore,
		Compatibility: ext.Compatibility,
		Tags:          ext.Tags,
		Changelog:     ext.Changelog,
		ReleaseNotes:  ext.ReleaseNotes,
	}
}

type ListParams struct {
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

type Extension struct {
	ID            string                 `json:"id"`
	CreatorID     string                 `json:"creator_id"`
	PluginID      *string                `json:"plugin_id"`
	Name          string                 `json:"name"`
	Version       string                 `json:"version"`
	Description   string                 `json:"description"`
	Category      string                 `json:"category"`
	IconURL       string                 `json:"icon_url"`
	Screenshots   []string               `json:"screenshots"`
	Manifest      map[string]interface{} `json:"manifest"`
	ManifestURL   string                 `json:"manifest_url"`
	Signature     string                 `json:"signature"`
	Verified      bool                   `json:"verified"`
	Status        string                 `json:"status"`
	Featured      bool                   `json:"featured"`
	InstallCount  int                    `json:"install_count"`
	RatingAverage float64                `json:"rating_average"`
	RatingCount   int                    `json:"rating_count"`
	TrustScore    float64                `json:"trust_score"`
	SandboxScore  float64                `json:"sandbox_score"`
	SecurityScore float64                `json:"security_score"`
	RuntimeScore  float64                `json:"runtime_score"`
	Compatibility map[string]interface{} `json:"compatibility"`
	Tags          []string               `json:"tags"`
	Changelog     string                 `json:"changelog"`
	ReleaseNotes  string                 `json:"release_notes"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	PublishedAt   *time.Time             `json:"published_at,omitempty"`
}

type CategoryCount struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

var _ HandlerRepo = (*StorageAdapter)(nil)
