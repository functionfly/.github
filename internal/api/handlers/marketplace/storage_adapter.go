package marketplace

import (
	"context"
	"encoding/json"

	"github.com/functionfly/functionfly/internal/storage"
)

type StorageAdapter struct {
	repo *storage.MarketplaceRepository
}

func NewStorageAdapter(repo *storage.MarketplaceRepository) *StorageAdapter {
	return &StorageAdapter{repo: repo}
}

func (a *StorageAdapter) List(ctx context.Context, params ListParams) ([]Extension, error) {
	listParams := storage.ListMarketplaceParams{
		CreatorID: params.CreatorID,
		Category:  params.Category,
		Status:   params.Status,
		Featured: params.Featured,
		Search:   params.Search,
		Tags:     params.Tags,
		Limit:    params.Limit,
		Offset:   params.Offset,
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

func toHandlerExtension(ext storage.MarketplaceExtension) Extension {
	manifestMap := make(map[string]interface{})
	if len(ext.Manifest) > 0 {
		_ = json.Unmarshal(ext.Manifest, &manifestMap)
	}
	return Extension{
		ID:              ext.ID,
		CreatorID:       ext.CreatorID,
		PluginID:        ext.PluginID,
		Name:            ext.Name,
		Version:         ext.Version,
		Description:     ext.Description,
		Category:        ext.Category,
		IconURL:         ext.IconURL,
		Screenshots:     ext.Screenshots,
		Manifest:        manifestMap,
		ManifestURL:     ext.ManifestURL,
		Signature:       ext.Signature,
		Verified:        ext.Verified,
		Status:          ext.Status,
		Featured:        ext.Featured,
		InstallCount:    ext.InstallCount,
		RatingAverage:   ext.RatingAverage,
		RatingCount:     ext.RatingCount,
		TrustScore:      ext.TrustScore,
		SandboxScore:    ext.SandboxScore,
		SecurityScore:   ext.SecurityScore,
		RuntimeScore:    ext.RuntimeScore,
		Compatibility:   ext.Compatibility,
		Tags:            ext.Tags,
		Changelog:       ext.Changelog,
		ReleaseNotes:    ext.ReleaseNotes,
	}
}

func toStorageExtension(ext *Extension) *storage.MarketplaceExtension {
	manifestBytes, _ := json.Marshal(ext.Manifest)
	return &storage.MarketplaceExtension{
		ID:              ext.ID,
		CreatorID:       ext.CreatorID,
		PluginID:        ext.PluginID,
		Name:            ext.Name,
		Version:         ext.Version,
		Description:     ext.Description,
		Category:        ext.Category,
		IconURL:         ext.IconURL,
		Screenshots:     ext.Screenshots,
		Manifest:        manifestBytes,
		ManifestURL:     ext.ManifestURL,
		Signature:       ext.Signature,
		Verified:        ext.Verified,
		Status:          ext.Status,
		Featured:        ext.Featured,
		InstallCount:    ext.InstallCount,
		RatingAverage:   ext.RatingAverage,
		RatingCount:     ext.RatingCount,
		TrustScore:      ext.TrustScore,
		SandboxScore:    ext.SandboxScore,
		SecurityScore:   ext.SecurityScore,
		RuntimeScore:    ext.RuntimeScore,
		Compatibility:   ext.Compatibility,
		Tags:            ext.Tags,
		Changelog:       ext.Changelog,
		ReleaseNotes:    ext.ReleaseNotes,
	}
}

type ListParams struct {
	CreatorID *string
	Category  *string
	Status    *string
	Featured  *bool
	Search    *string
	Tags      []string
	Limit     int
	Offset    int
}

type Extension struct {
	ID              string
	CreatorID       string
	PluginID        *string
	Name            string
	Version         string
	Description     string
	Category        string
	IconURL         string
	Screenshots     []string
	Manifest        map[string]interface{}
	ManifestURL     string
	Signature       string
	Verified        bool
	Status          string
	Featured        bool
	InstallCount    int
	RatingAverage   float64
	RatingCount     int
	TrustScore      float64
	SandboxScore    float64
	SecurityScore   float64
	RuntimeScore    float64
	Compatibility   map[string]interface{}
	Tags            []string
	Changelog       string
	ReleaseNotes    string
}

type CategoryCount struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

var _ HandlerRepo = (*StorageAdapter)(nil)