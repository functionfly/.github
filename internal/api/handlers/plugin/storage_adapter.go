package plugin

import (
	"context"

	"github.com/functionfly/functionfly/internal/storage"
)

type StorageAdapter struct {
	repo *storage.PluginRepository
}

func NewStorageAdapter(repo *storage.PluginRepository) *StorageAdapter {
	return &StorageAdapter{repo: repo}
}

func (a *StorageAdapter) List(ctx context.Context, params ListPluginsParams) ([]Plugin, error) {
	storageParams := storage.ListPluginsParams{
		TenantID: params.TenantID,
		Category: params.Category,
		Search:   params.Search,
		Limit:    params.Limit,
		Offset:   params.Offset,
	}
	if params.PluginType != nil {
		pt := storage.PluginType(*params.PluginType)
		storageParams.PluginType = &pt
	}
	if params.Status != nil {
		s := storage.PluginStatus(*params.Status)
		storageParams.Status = &s
	}

	plugins, err := a.repo.List(ctx, storageParams)
	if err != nil {
		return nil, err
	}

	return convertPlugins(plugins), nil
}

func (a *StorageAdapter) Get(ctx context.Context, tenantID, pluginID string) (*Plugin, error) {
	plugin, err := a.repo.Get(ctx, tenantID, pluginID)
	if err != nil || plugin == nil {
		return nil, err
	}
	p := convertPlugin(plugin)
	return &p, nil
}

func (a *StorageAdapter) Create(ctx context.Context, plugin *Plugin) error {
	p := convertToStoragePlugin(plugin)
	return a.repo.Create(ctx, p)
}

func (a *StorageAdapter) Update(ctx context.Context, plugin *Plugin) error {
	p := convertToStoragePlugin(plugin)
	return a.repo.Update(ctx, p)
}

func (a *StorageAdapter) Delete(ctx context.Context, tenantID, pluginID string) error {
	return a.repo.Delete(ctx, tenantID, pluginID)
}

func (a *StorageAdapter) SetStatus(ctx context.Context, tenantID, pluginID string, status PluginStatus) error {
	return a.repo.SetStatus(ctx, tenantID, pluginID, storage.PluginStatus(status))
}

func (a *StorageAdapter) SetError(ctx context.Context, tenantID, pluginID string, errMsg string) error {
	return a.repo.SetError(ctx, tenantID, pluginID, errMsg)
}

func (a *StorageAdapter) UpdateConfig(ctx context.Context, tenantID, pluginID string, config map[string]string) error {
	return a.repo.UpdateConfig(ctx, tenantID, pluginID, config)
}

func (a *StorageAdapter) GetEnabledByType(ctx context.Context, tenantID string, pluginType PluginType) (*Plugin, error) {
	p, err := a.repo.GetEnabledByType(ctx, tenantID, storage.PluginType(pluginType))
	if err != nil || p == nil {
		return nil, err
	}
	plugin := convertPlugin(p)
	return &plugin, nil
}

func (a *StorageAdapter) GetSandbox(ctx context.Context, pluginID string) (*PluginSandbox, error) {
	sandbox, err := a.repo.GetSandbox(ctx, pluginID)
	if err != nil || sandbox == nil {
		return nil, err
	}
	return convertSandbox(sandbox), nil
}

func (a *StorageAdapter) UpsertSandbox(ctx context.Context, sandbox *PluginSandbox) error {
	s := &storage.PluginSandbox{
		ID:             sandbox.ID,
		PluginID:       sandbox.PluginID,
		Tier:           storage.SandboxTier(sandbox.Tier),
		CPULimit:       sandbox.CPULimit,
		MemoryLimitMB:  sandbox.MemoryLimitMB,
		TimeoutSeconds: sandbox.TimeoutSeconds,
		MaxInstances:   sandbox.MaxInstances,
		EnvVars:        sandbox.EnvVars,
		AllowedDomains: sandbox.AllowedDomains,
		BlockedDomains: sandbox.BlockedDomains,
		RateLimitRPM:   sandbox.RateLimitRPM,
	}
	return a.repo.UpsertSandbox(ctx, s)
}

func (a *StorageAdapter) ListPermissions(ctx context.Context, pluginID string) ([]PluginPermission, error) {
	perms, err := a.repo.ListPermissions(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	return convertPermissions(perms), nil
}

func (a *StorageAdapter) SetPermission(ctx context.Context, perm *PluginPermission) error {
	p := &storage.PluginPermission{
		ID:               perm.ID,
		PluginID:         perm.PluginID,
		PermissionType:   perm.PermissionType,
		PermissionAction: perm.PermissionAction,
		Resource:         perm.Resource,
		Granted:          perm.Granted,
		GrantedAt:        perm.GrantedAt,
		GrantedBy:        perm.GrantedBy,
		ExpiresAt:        perm.ExpiresAt,
	}
	return a.repo.SetPermission(ctx, p)
}

func (a *StorageAdapter) CreateVersion(ctx context.Context, version *PluginVersion) error {
	v := &storage.PluginVersion{
		ID:        version.ID,
		PluginID:  version.PluginID,
		Version:   version.Version,
		Changelog: version.Changelog,
		Manifest:  version.Manifest,
		SizeBytes: version.SizeBytes,
		Signature: version.Signature,
		ReleaseAt: version.ReleaseAt,
	}
	return a.repo.CreateVersion(ctx, v)
}

func (a *StorageAdapter) ListVersions(ctx context.Context, pluginID string) ([]PluginVersion, error) {
	versions, err := a.repo.ListVersions(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	return convertVersions(versions), nil
}

func (a *StorageAdapter) GetPreviousVersion(ctx context.Context, pluginID, currentVersion string) (*PluginVersion, error) {
	version, err := a.repo.GetPreviousVersion(ctx, pluginID, currentVersion)
	if err != nil || version == nil {
		return nil, err
	}
	v := convertVersion(version)
	return &v, nil
}

func (a *StorageAdapter) GetTelemetrySummary(ctx context.Context, tenantID, pluginID string, timeRange string) (*TelemetrySummary, error) {
	s, err := a.repo.GetTelemetrySummary(ctx, tenantID, pluginID, timeRange)
	if err != nil {
		return nil, err
	}
	return &TelemetrySummary{
		Executions:         s.Executions,
		Errors:             s.Errors,
		ErrorRate:          s.ErrorRate,
		AvgLatencyMs:       s.AvgLatencyMs,
		CPUUsageSeconds:    s.CPUUsageSeconds,
		AvgMemoryUsageMB:   s.AvgMemoryUsageMB,
		NetworkBytes:       s.NetworkBytes,
		PreviousExecutions: s.PreviousExecutions,
		LatencyTrend:       s.LatencyTrend,
		ExecutionsTrend:    s.ExecutionsTrend,
	}, nil
}

func (a *StorageAdapter) RecordAnalytics(ctx context.Context, analytics *PluginAnalytics) error {
	return a.repo.RecordAnalytics(ctx, &storage.PluginAnalytics{
		ID:               analytics.ID,
		PluginID:         analytics.PluginID,
		TenantID:         analytics.TenantID,
		EventType:        analytics.EventType,
		ExecutionsCount:  analytics.ExecutionsCount,
		ErrorsCount:      analytics.ErrorsCount,
		TotalLatencyMs:   analytics.TotalLatencyMs,
		CPUUsageSeconds:  analytics.CPUUsageSeconds,
		MemoryUsageMBAvg: analytics.MemoryUsageMBAvg,
		NetworkBytes:     analytics.NetworkBytes,
		PeriodStart:      analytics.PeriodStart,
		PeriodEnd:        analytics.PeriodEnd,
		Metadata:         analytics.Metadata,
		CreatedAt:        analytics.CreatedAt,
	})
}

func convertPlugin(p *storage.Plugin) Plugin {
	return Plugin{
		ID:            p.ID,
		TenantID:      p.TenantID,
		Manifest:      p.Manifest,
		PluginType:    PluginType(p.PluginType),
		Name:          p.Name,
		Version:       p.Version,
		Description:   p.Description,
		AuthorName:    p.AuthorName,
		AuthorEmail:   p.AuthorEmail,
		AuthorWebsite: p.AuthorWebsite,
		Category:      p.Category,
		Status:        PluginStatus(p.Status),
		IconURL:       p.IconURL,
		HomepageURL:   p.HomepageURL,
		RepositoryURL: p.RepositoryURL,
		License:       p.License,
		SizeBytes:     p.SizeBytes,
		Signature:     p.Signature,
		Verified:      p.Verified,
		Config:        p.Config,
		Metadata:      p.Metadata,
		InstalledAt:   p.InstalledAt,
		UpdatedAt:     p.UpdatedAt,
		EnabledAt:     p.EnabledAt,
		ErrorMessage:  p.ErrorMessage,
	}
}

func convertPlugins(plugins []storage.Plugin) []Plugin {
	result := make([]Plugin, len(plugins))
	for i, p := range plugins {
		result[i] = convertPlugin(&p)
	}
	return result
}

func convertToStoragePlugin(p *Plugin) *storage.Plugin {
	return &storage.Plugin{
		ID:            p.ID,
		TenantID:      p.TenantID,
		Manifest:      p.Manifest,
		PluginType:    storage.PluginType(p.PluginType),
		Name:          p.Name,
		Version:       p.Version,
		Description:   p.Description,
		AuthorName:    p.AuthorName,
		AuthorEmail:   p.AuthorEmail,
		AuthorWebsite: p.AuthorWebsite,
		Category:      p.Category,
		Status:        storage.PluginStatus(p.Status),
		IconURL:       p.IconURL,
		HomepageURL:   p.HomepageURL,
		RepositoryURL: p.RepositoryURL,
		License:       p.License,
		SizeBytes:     p.SizeBytes,
		Signature:     p.Signature,
		Verified:      p.Verified,
		Config:        p.Config,
		Metadata:      p.Metadata,
		InstalledAt:   p.InstalledAt,
		UpdatedAt:     p.UpdatedAt,
		EnabledAt:     p.EnabledAt,
		ErrorMessage:  p.ErrorMessage,
	}
}

func convertSandbox(s *storage.PluginSandbox) *PluginSandbox {
	return &PluginSandbox{
		ID:              s.ID,
		PluginID:        s.PluginID,
		Tier:            SandboxTier(s.Tier),
		CPULimit:        s.CPULimit,
		MemoryLimitMB:   s.MemoryLimitMB,
		TimeoutSeconds:  s.TimeoutSeconds,
		NetworkIsolated: s.NetworkIsolated,
		FilesystemScope: s.FilesystemScope,
		MaxInstances:    s.MaxInstances,
		EnvVars:         s.EnvVars,
		AllowedDomains:  s.AllowedDomains,
		BlockedDomains:  s.BlockedDomains,
		RateLimitRPM:    s.RateLimitRPM,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}

func convertPermissions(perms []storage.PluginPermission) []PluginPermission {
	result := make([]PluginPermission, len(perms))
	for i, p := range perms {
		result[i] = PluginPermission{
			ID:               p.ID,
			PluginID:         p.PluginID,
			PermissionType:   p.PermissionType,
			PermissionAction: p.PermissionAction,
			Resource:         p.Resource,
			Granted:          p.Granted,
			GrantedAt:        p.GrantedAt,
			GrantedBy:        p.GrantedBy,
			ExpiresAt:        p.ExpiresAt,
			CreatedAt:        p.CreatedAt,
		}
	}
	return result
}

func convertVersion(v *storage.PluginVersion) PluginVersion {
	return PluginVersion{
		ID:        v.ID,
		PluginID:  v.PluginID,
		Version:   v.Version,
		Changelog: v.Changelog,
		Manifest:  v.Manifest,
		SizeBytes: v.SizeBytes,
		Signature: v.Signature,
		ReleaseAt: v.ReleaseAt,
		CreatedAt: v.CreatedAt,
	}
}

func convertVersions(versions []storage.PluginVersion) []PluginVersion {
	result := make([]PluginVersion, len(versions))
	for i, v := range versions {
		result[i] = convertVersion(&v)
	}
	return result
}
