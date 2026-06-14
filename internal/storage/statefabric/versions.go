package statefabric

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/functionfly/functionfly/internal/monitoring"
)

type FabricVersionFilter struct {
	Limit       int
	Offset      int
	ChangeType  string
	StartTime   *time.Time
	EndTime     *time.Time
}

type ListFabricVersionsResult struct {
	Versions     []StateFabricVersion `json:"versions"`
	Total        int64               `json:"total"`
	CurrentVersion int               `json:"current_version"`
}

type KeyVersionFilter struct {
	Limit     int
	Offset    int
	StartTime *time.Time
	EndTime   *time.Time
}

type ListKeyVersionsResult struct {
	Versions     []StateFabricKeyVersion `json:"versions"`
	Total        int64                 `json:"total"`
	CurrentValue JSONMap                `json:"current_value"`
}

// ListFabricVersions returns version history for a fabric
func (r *Repository) ListFabricVersions(ctx context.Context, tenantID, fabricID uuid.UUID, filter FabricVersionFilter) (*ListFabricVersionsResult, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf("state fabric not found")
	}

	query := r.db.WithContext(ctx).Model(&StateFabricVersion{}).Where("fabric_id = ?", fabricID)

	if filter.ChangeType != "" {
		query = query.Where("change_type = ?", filter.ChangeType)
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	var versions []StateFabricVersion
	if err := query.Order("version_number DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&versions).Error; err != nil {
		return nil, err
	}

	// Get current version number
	var currentVersion int
	if len(versions) > 0 {
		currentVersion = versions[0].VersionNumber
	}

	return &ListFabricVersionsResult{
		Versions:      versions,
		Total:          total,
		CurrentVersion: currentVersion,
	}, nil
}

// GetFabricVersion returns a specific version of a fabric
func (r *Repository) GetFabricVersion(ctx context.Context, tenantID, fabricID uuid.UUID, versionNumber int) (*StateFabricVersion, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf("state fabric not found")
	}

	var version StateFabricVersion
	if err := r.db.WithContext(ctx).Where("fabric_id = ? AND version_number = ?", fabricID, versionNumber).First(&version).Error; err != nil {
		return nil, err
	}

	return &version, nil
}

// CreateFabricVersion creates a version snapshot of the fabric configuration
func (r *Repository) CreateFabricVersion(ctx context.Context, tenantID, fabricID uuid.UUID, changeType, changeSummary string, actorID uuid.UUID, actorType string) (*StateFabricVersion, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf("state fabric not found")
	}

	// Get the next version number
	var maxVersion int
	r.db.WithContext(ctx).Model(&StateFabricVersion{}).Where("fabric_id = ?", fabricID).Select("COALESCE(MAX(version_number), 0)").Scan(&maxVersion)

	version := &StateFabricVersion{
		ID:            uuid.New(),
		FabricID:      fabricID,
		VersionNumber: maxVersion + 1,
		Name:          state.Name,
		Description:   normalizeDescription(state.Description),
		Type:          state.StorageType,
		Settings:      JSONMap(state.Tags),
		ChangeType:    changeType,
		ChangeSummary: changeSummary,
		ActorID:       actorID,
		ActorType:     actorType,
	}

	if err := r.db.WithContext(ctx).Create(version).Error; err != nil {
		return nil, err
	}

	monitoring.RecordStateFabricOperation(tenantID.String(), fabricID.String(), "version_create", "success")

	return version, nil
}

// DiffFabricVersions compares two versions of a fabric
func (r *Repository) DiffFabricVersions(ctx context.Context, tenantID, fabricID uuid.UUID, fromVersion, toVersion int) (*FabricVersionDiff, error) {
	from, err := r.GetFabricVersion(ctx, tenantID, fabricID, fromVersion)
	if err != nil {
		return nil, err
	}
	to, err := r.GetFabricVersion(ctx, tenantID, fabricID, toVersion)
	if err != nil {
		return nil, err
	}

	diff := &FabricVersionDiff{
		FromVersion:     fromVersion,
		ToVersion:       toVersion,
		NameChanged:     from.Name != to.Name,
		NameFrom:        from.Name,
		NameTo:          to.Name,
		DescChanged:     from.Description != to.Description,
		DescFrom:        from.Description,
		DescTo:          to.Description,
		TypeChanged:     from.Type != to.Type,
		TypeFrom:        from.Type,
		TypeTo:          to.Type,
		SettingsChanged: !settingsEqual(from.Settings, to.Settings),
		SettingsFrom:    from.Settings,
		SettingsTo:      to.Settings,
		ActorID:         to.ActorID,
		ActorType:       to.ActorType,
		CreatedAt:       to.CreatedAt,
	}

	diff.HasChanges = diff.NameChanged || diff.DescChanged || diff.TypeChanged || diff.SettingsChanged

	return diff, nil
}

type FabricVersionDiff struct {
	FromVersion     int         `json:"from_version"`
	ToVersion       int         `json:"to_version"`
	HasChanges      bool        `json:"has_changes"`
	NameChanged     bool        `json:"name_changed"`
	NameFrom        string      `json:"name_from"`
	NameTo          string      `json:"name_to"`
	DescChanged     bool        `json:"desc_changed"`
	DescFrom        string      `json:"desc_from"`
	DescTo          string      `json:"desc_to"`
	TypeChanged     bool        `json:"type_changed"`
	TypeFrom        string      `json:"type_from"`
	TypeTo          string      `json:"type_to"`
	SettingsChanged bool        `json:"settings_changed"`
	SettingsFrom    JSONMap     `json:"settings_from"`
	SettingsTo      JSONMap     `json:"settings_to"`
	ActorID         uuid.UUID   `json:"actor_id"`
	ActorType       string      `json:"actor_type"`
	CreatedAt       time.Time   `json:"created_at"`
}

func settingsEqual(a, b JSONMap) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// ListKeyVersions returns version history for a key
func (r *Repository) ListKeyVersions(ctx context.Context, tenantID, fabricID uuid.UUID, key string, filter KeyVersionFilter) (*ListKeyVersionsResult, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf("state fabric not found")
	}

	query := r.db.WithContext(ctx).Model(&StateFabricKeyVersion{}).Where("fabric_id = ? AND key = ?", fabricID, key)

	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	var versions []StateFabricKeyVersion
	if err := query.Order("version_number DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&versions).Error; err != nil {
		return nil, err
	}

	// Get current value
	var currentValue JSONMap
	if len(versions) > 0 {
		currentValue = versions[0].Value
	}

	return &ListKeyVersionsResult{
		Versions:     versions,
		Total:        total,
		CurrentValue: currentValue,
	}, nil
}

// GetKeyVersion returns a specific version of a key
func (r *Repository) GetKeyVersion(ctx context.Context, tenantID, fabricID uuid.UUID, key string, versionNumber int) (*StateFabricKeyVersion, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf("state fabric not found")
	}

	var version StateFabricKeyVersion
	if err := r.db.WithContext(ctx).Where("fabric_id = ? AND key = ? AND version_number = ?", fabricID, key, versionNumber).First(&version).Error; err != nil {
		return nil, err
	}

	return &version, nil
}

// CreateKeyVersion creates a version snapshot of a key value
func (r *Repository) CreateKeyVersion(ctx context.Context, tenantID, fabricID uuid.UUID, key string, changeType string, value JSONMap, actorID uuid.UUID, actorType string) (*StateFabricKeyVersion, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf("state fabric not found")
	}

	// Get the next version number
	var maxVersion int
	r.db.WithContext(ctx).Model(&StateFabricKeyVersion{}).Where("fabric_id = ? AND key = ?", fabricID, key).Select("COALESCE(MAX(version_number), 0)").Scan(&maxVersion)

	version := &StateFabricKeyVersion{
		ID:            uuid.New(),
		FabricID:      fabricID,
		Key:           key,
		VersionNumber: maxVersion + 1,
		Value:         value,
		ChangeType:    changeType,
		ChangeSummary: fmt.Sprintf("Key %s: version %d", key, maxVersion+1),
		ActorID:       actorID,
		ActorType:     actorType,
	}

	if err := r.db.WithContext(ctx).Create(version).Error; err != nil {
		return nil, err
	}

	return version, nil
}

// DiffKeyVersions compares two versions of a key
func (r *Repository) DiffKeyVersions(ctx context.Context, tenantID, fabricID uuid.UUID, key string, fromVersion, toVersion int) (*KeyVersionDiff, error) {
	from, err := r.GetKeyVersion(ctx, tenantID, fabricID, key, fromVersion)
	if err != nil {
		return nil, err
	}
	to, err := r.GetKeyVersion(ctx, tenantID, fabricID, key, toVersion)
	if err != nil {
		return nil, err
	}

	diff := &KeyVersionDiff{
		FromVersion:  fromVersion,
		ToVersion:    toVersion,
		Key:          key,
		ValueChanged: !valueEqual(from.Value, to.Value),
		ValueFrom:    from.Value,
		ValueTo:      to.Value,
		ActorID:      to.ActorID,
		ActorType:    to.ActorType,
		CreatedAt:    to.CreatedAt,
	}

	diff.HasChanges = diff.ValueChanged

	return diff, nil
}

type KeyVersionDiff struct {
	FromVersion  int         `json:"from_version"`
	ToVersion    int         `json:"to_version"`
	Key          string      `json:"key"`
	HasChanges   bool        `json:"has_changes"`
	ValueChanged bool        `json:"value_changed"`
	ValueFrom    JSONMap     `json:"value_from"`
	ValueTo      JSONMap     `json:"value_to"`
	ActorID      uuid.UUID   `json:"actor_id"`
	ActorType    string      `json:"actor_type"`
	CreatedAt    time.Time   `json:"created_at"`
}

func valueEqual(a, b JSONMap) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}

func (r *Repository) recordFabricVersion(ctx context.Context, tenantID, fabricID uuid.UUID, changeType, changeSummary string, actorID uuid.UUID, actorType string) error {
	_, err := r.CreateFabricVersion(ctx, tenantID, fabricID, changeType, changeSummary, actorID, actorType)
	if err != nil {
		logrus.WithError(err).Warn("failed to record fabric version")
	}
	return err
}

func (r *Repository) recordKeyVersion(ctx context.Context, tenantID, fabricID uuid.UUID, key string, changeType string, value JSONMap, actorID uuid.UUID, actorType string) error {
	_, err := r.CreateKeyVersion(ctx, tenantID, fabricID, key, changeType, value, actorID, actorType)
	if err != nil {
		logrus.WithError(err).Warn("failed to record key version")
	}
	return err
}

// RollbackFabric rolls back a fabric to a previous version
func (r *Repository) RollbackFabric(ctx context.Context, tenantID, fabricID uuid.UUID, targetVersion int, actorID uuid.UUID, actorType string) (*Fabric, error) {
	version, err := r.GetFabricVersion(ctx, tenantID, fabricID, targetVersion)
	if err != nil {
		return nil, err
	}

	// Update the fabric with the old configuration
	updates := map[string]interface{}{
		"name":        version.Name,
		"description": version.Description,
		"settings":    version.Settings,
	}

	fabric, err := r.UpdateFabric(ctx, tenantID, fabricID, updates)
	if err != nil {
		return nil, err
	}

	// Record the rollback as a new version
	summary := fmt.Sprintf("Rolled back to version %d", targetVersion)
	r.CreateFabricVersion(ctx, tenantID, fabricID, "rollback", summary, actorID, actorType)

	// Invalidate cache
	if r.cache != nil {
		r.cache.InvalidateFabric(ctx, tenantID, fabricID)
		r.cache.InvalidateFabricList(ctx, tenantID)
	}

	return fabric, nil
}

// RollbackKey rolls back a key to a previous version
func (r *Repository) RollbackKey(ctx context.Context, tenantID, fabricID uuid.UUID, key string, targetVersion int, actorID uuid.UUID, actorType string) error {
	version, err := r.GetKeyVersion(ctx, tenantID, fabricID, key, targetVersion)
	if err != nil {
		return err
	}

	// Set the old value
	_, err = r.SetFabricValue(ctx, tenantID, fabricID, key, version.Value, "rollback")
	if err != nil {
		return err
	}

	// Record the rollback as a new version
	_, err = r.CreateKeyVersion(ctx, tenantID, fabricID, key, "rollback", version.Value, actorID, actorType)
	if err != nil {
		return err
	}

	return nil
}