package statefabric

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	statestore "github.com/functionfly/functionfly/internal/storage/state"
)

type Repository struct {
	db        *gorm.DB
	stateRepo *statestore.StateRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db, stateRepo: statestore.NewStateRepository(db)}
}

type Fabric struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Type        string                 `json:"type"`
	TenantID    uuid.UUID              `json:"tenantId"`
	Stores      []FabricStore          `json:"stores"`
	Pipelines   []Pipeline             `json:"pipelines"`
	Throughput  int64                  `json:"throughput"`
	Latency     float64                `json:"latency"`
	LastUpdated time.Time              `json:"lastUpdated"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	Settings    map[string]interface{} `json:"settings"`
	Metrics     FabricMetrics          `json:"metrics"`
}

type FabricStore struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	Size       int64     `json:"size"`
	MaxSize    int64     `json:"maxSize"`
	Region     string    `json:"region"`
	Provider   string    `json:"provider"`
	Throughput float64   `json:"throughput"`
	Latency    float64   `json:"latency"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type Pipeline struct {
	ID             string                   `json:"id"`
	Name           string                   `json:"name"`
	Description    string                   `json:"description"`
	Status         string                   `json:"status"`
	Steps          []map[string]interface{} `json:"steps"`
	InputSchema    map[string]interface{}   `json:"inputSchema,omitempty"`
	OutputSchema   map[string]interface{}   `json:"outputSchema,omitempty"`
	Throughput     float64                  `json:"throughput"`
	ErrorRate      float64                  `json:"errorRate"`
	LastExecutedAt *time.Time               `json:"lastExecutedAt,omitempty"`
	CreatedAt      time.Time                `json:"createdAt"`
	UpdatedAt      time.Time                `json:"updatedAt"`
}

type EventLog struct {
	ID             string                 `json:"id"`
	FabricID       string                 `json:"fabricId"`
	StoreID        string                 `json:"storeId,omitempty"`
	EventType      string                 `json:"eventType"`
	Payload        map[string]interface{} `json:"payload"`
	Timestamp      time.Time              `json:"timestamp"`
	SequenceNumber int64                  `json:"sequenceNumber"`
	CorrelationID  string                 `json:"correlationId,omitempty"`
}

type Snapshot struct {
	ID          string                 `json:"id"`
	FabricID    string                 `json:"fabricId"`
	StoreID     string                 `json:"storeId,omitempty"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	State       map[string]interface{} `json:"state"`
	EventCount  int                    `json:"eventCount"`
	SizeBytes   int64                  `json:"sizeBytes"`
	CreatedAt   time.Time              `json:"createdAt"`
	ExpiresAt   *time.Time             `json:"expiresAt,omitempty"`
}

type ReplaySession struct {
	ID             string     `json:"id"`
	FabricID       string     `json:"fabricId"`
	SnapshotID     string     `json:"snapshotId,omitempty"`
	StartEventID   string     `json:"startEventId,omitempty"`
	EndEventID     string     `json:"endEventId,omitempty"`
	Status         string     `json:"status"`
	Progress       int        `json:"progress"`
	EventsReplayed int        `json:"eventsReplayed"`
	StartedAt      time.Time  `json:"startedAt"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	Error          string     `json:"error,omitempty"`
}

type FabricMetrics struct {
	TotalOperations     int64     `json:"totalOperations"`
	OperationsPerSecond float64   `json:"operationsPerSecond"`
	AverageLatency      float64   `json:"averageLatency"`
	ErrorRate           float64   `json:"errorRate"`
	CacheHitRate        *float64  `json:"cacheHitRate,omitempty"`
	StorageUsed         int64     `json:"storageUsed"`
	LastCalculatedAt    time.Time `json:"lastCalculatedAt"`
}

type ListOptions struct {
	TenantID uuid.UUID
	Limit    int
	Offset   int
	Status   string
	Search   string
}

type EventListOptions struct {
	StoreID   string
	EventType string
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
	Offset    int
}

type ReplayCreateRequest struct {
	SnapshotID   string
	StartEventID string
	EndEventID   string
}

func defaultSettings(state *statestore.State) map[string]interface{} {
	retention := state.TTLDays
	if retention <= 0 {
		retention = 30
	}
	return map[string]interface{}{
		"autoSnapshot":            false,
		"snapshotIntervalMinutes": 60,
		"retentionDays":           retention,
		"enableReplication":       false,
		"regions":                 []string{},
		"conflictResolution":      "last-write-wins",
	}
}

func stateType(storageType string) string {
	switch storageType {
	case "timeseries":
		return "workflow"
	case "document":
		return "catalog"
	case "graph":
		return "custom"
	default:
		return "cache"
	}
}

func stateStatus(state *statestore.State) string {
	if state == nil {
		return "offline"
	}
	if state.UpdatedAt.Before(time.Now().Add(-24 * time.Hour)) {
		return "degraded"
	}
	return "online"
}

func normalizeDescription(description *string) string {
	if description == nil {
		return ""
	}
	return *description
}

func safeJSONMap(value statestore.JSONMap) map[string]interface{} {
	if value == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}(value)
}

func buildStore(state *statestore.State) FabricStore {
	region := "global"
	if tags := safeJSONMap(state.Tags); tags != nil {
		if taggedRegion, ok := tags["region"].(string); ok && taggedRegion != "" {
			region = taggedRegion
		}
	}
	provider := "functionfly"
	storeType := "persistent"
	if state.StorageType == "keyvalue" {
		storeType = "cache"
	}
	return FabricStore{
		ID:         state.ID.String(),
		Name:       state.Name,
		Type:       storeType,
		Status:     "active",
		Size:       state.StorageUsedMB * 1024 * 1024,
		MaxSize:    int64(state.MaxSizeMB) * 1024 * 1024,
		Region:     region,
		Provider:   provider,
		Throughput: float64(state.WriteOpsMonth+state.ReadOpsMonth) / float64(maxInt(daysSince(state.CreatedAt), 1)),
		Latency:    0,
		CreatedAt:  state.CreatedAt,
		UpdatedAt:  state.UpdatedAt,
	}
}

func buildFabric(state *statestore.State, metrics FabricMetrics, pipelines []Pipeline) Fabric {
	store := buildStore(state)
	return Fabric{
		ID:          state.ID,
		Name:        state.Name,
		Description: normalizeDescription(state.Description),
		Status:      stateStatus(state),
		Type:        stateType(state.StorageType),
		TenantID:    state.TenantID,
		Stores:      []FabricStore{store},
		Pipelines:   pipelines,
		Throughput:  state.WriteOpsMonth + state.ReadOpsMonth,
		Latency:     metrics.AverageLatency,
		LastUpdated: state.UpdatedAt,
		CreatedAt:   state.CreatedAt,
		UpdatedAt:   state.UpdatedAt,
		Settings:    defaultSettings(state),
		Metrics:     metrics,
	}
}

func daysSince(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return int64(time.Since(t).Hours() / 24)
}

func maxInt(v, fallback int64) int64 {
	if v <= 0 {
		return fallback
	}
	return v
}

func (r *Repository) ListFabrics(ctx context.Context, opts ListOptions) ([]Fabric, int64, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 100
	}
	states, total, err := r.stateRepo.ListStatesByTenant(ctx, opts.TenantID, limit, opts.Offset)
	if err != nil {
		return nil, 0, err
	}
	items := make([]Fabric, 0, len(states))
	for _, state := range states {
		if opts.Status != "" && opts.Status != "all" && stateStatus(state) != opts.Status {
			continue
		}
		if opts.Search != "" {
			query := strings.ToLower(opts.Search)
			if !strings.Contains(strings.ToLower(state.Name), query) && !strings.Contains(strings.ToLower(normalizeDescription(state.Description)), query) {
				continue
			}
		}
		metrics, _ := r.GetMetrics(ctx, state.ID, "")
		pipelines, _ := r.ListPipelines(ctx, state.ID)
		items = append(items, buildFabric(state, metrics, pipelines))
	}
	return items, total, nil
}

func (r *Repository) CreateFabric(ctx context.Context, tenantID uuid.UUID, name, description, fabricType string, settings map[string]interface{}) (*Fabric, error) {
	storageType := "keyvalue"
	switch fabricType {
	case "catalog":
		storageType = "document"
	case "workflow":
		storageType = "timeseries"
	case "custom":
		storageType = "graph"
	}
	state := &statestore.State{
		TenantID:    tenantID,
		Name:        name,
		FullPath:    fmt.Sprintf("%s/%s", tenantID.String()[:8], name),
		StorageType: storageType,
		Description: stringPtr(description),
		Tags: statestore.JSONMap{
			"fabric_type": fabricType,
			"settings":    settings,
		},
	}
	created, err := r.stateRepo.CreateState(ctx, state)
	if err != nil {
		return nil, err
	}
	metrics, _ := r.GetMetrics(ctx, created.ID, "")
	fabric := buildFabric(created, metrics, nil)
	return &fabric, nil
}

func (r *Repository) GetFabric(ctx context.Context, tenantID, fabricID uuid.UUID) (*Fabric, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf("state fabric not found")
	}
	metrics, _ := r.GetMetrics(ctx, state.ID, "")
	pipelines, _ := r.ListPipelines(ctx, state.ID)
	fabric := buildFabric(state, metrics, pipelines)
	return &fabric, nil
}

func (r *Repository) UpdateFabric(ctx context.Context, tenantID, fabricID uuid.UUID, updates map[string]interface{}) (*Fabric, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf("state fabric not found")
	}
	if name, ok := updates["name"].(string); ok && strings.TrimSpace(name) != "" {
		state.Name = name
		state.FullPath = fmt.Sprintf("%s/%s", tenantID.String()[:8], name)
	}
	if description, ok := updates["description"].(string); ok {
		state.Description = stringPtr(description)
	}
	if settings, ok := updates["settings"].(map[string]interface{}); ok {
		tags := safeJSONMap(state.Tags)
		tags["settings"] = settings
		state.Tags = statestore.JSONMap(tags)
	}
	updated, err := r.stateRepo.UpdateState(ctx, state)
	if err != nil {
		return nil, err
	}
	metrics, _ := r.GetMetrics(ctx, updated.ID, "")
	pipelines, _ := r.ListPipelines(ctx, updated.ID)
	fabric := buildFabric(updated, metrics, pipelines)
	return &fabric, nil
}

func (r *Repository) DeleteFabric(ctx context.Context, tenantID, fabricID uuid.UUID) error {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return err
	}
	if state.TenantID != tenantID {
		return fmt.Errorf("state fabric not found")
	}
	return r.stateRepo.DeleteState(ctx, fabricID)
}

func (r *Repository) GetMetrics(ctx context.Context, fabricID uuid.UUID, _ string) (FabricMetrics, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return FabricMetrics{}, err
	}
	var eventCount int64
	if err := r.db.WithContext(ctx).Model(&statestore.StateEvent{}).Where("state_id = ?", fabricID).Count(&eventCount).Error; err != nil {
		return FabricMetrics{}, err
	}
	var snapshotCount int64
	_ = r.db.WithContext(ctx).Model(&statestore.StateSnapshot{}).Where("state_id = ?", fabricID).Count(&snapshotCount).Error
	days := float64(maxInt(daysSince(state.CreatedAt), 1))
	avgLatency := 5.0
	metrics := FabricMetrics{
		TotalOperations:     state.WriteOpsMonth + state.ReadOpsMonth,
		OperationsPerSecond: float64(state.WriteOpsMonth+state.ReadOpsMonth) / (days * 86400),
		AverageLatency:      avgLatency,
		ErrorRate:           0,
		StorageUsed:         state.StorageUsedMB * 1024,
		LastCalculatedAt:    time.Now(),
	}
	if snapshotCount > 0 {
		cache := float64(100)
		metrics.CacheHitRate = &cache
	}
	if eventCount == 0 {
		metrics.OperationsPerSecond = 0
	}
	return metrics, nil
}

func (r *Repository) ListStores(ctx context.Context, tenantID, fabricID uuid.UUID) ([]FabricStore, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf("state fabric not found")
	}
	return []FabricStore{buildStore(state)}, nil
}

func (r *Repository) CreateStore(ctx context.Context, tenantID, fabricID uuid.UUID, name, storeType string, maxSize int64, region string) (*FabricStore, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf("state fabric not found")
	}
	if strings.TrimSpace(name) != "" {
		state.Name = name
	}
	if maxSize > 0 {
		state.MaxSizeMB = int(maxSize / (1024 * 1024))
	}
	state.StorageType = mapStoreType(storeType)
	tags := safeJSONMap(state.Tags)
	if region != "" {
		tags["region"] = region
	}
	state.Tags = statestore.JSONMap(tags)
	updated, err := r.stateRepo.UpdateState(ctx, state)
	if err != nil {
		return nil, err
	}
	store := buildStore(updated)
	return &store, nil
}

func mapStoreType(storeType string) string {
	switch storeType {
	case "queue":
		return "timeseries"
	case "persistent":
		return "document"
	case "memory":
		return "keyvalue"
	default:
		return "graph"
	}
}

func (r *Repository) DeleteStore(ctx context.Context, tenantID, fabricID uuid.UUID, _ string) error {
	_, err := r.GetFabric(ctx, tenantID, fabricID)
	return err
}

func (r *Repository) ListPipelines(ctx context.Context, fabricID uuid.UUID) ([]Pipeline, error) {
	triggers, err := r.stateRepo.GetTriggers(ctx, fabricID)
	if err != nil {
		return nil, nil
	}
	pipelines := make([]Pipeline, 0, len(triggers))
	for _, trigger := range triggers {
		steps := []map[string]interface{}{
			{
				"id":   trigger.ID.String(),
				"name": defaultTriggerName(trigger),
				"type": "custom",
				"config": map[string]interface{}{
					"targetFunction": trigger.TargetFunction,
					"keyPattern":     derefString(trigger.KeyPattern),
				},
				"order":      1,
				"enabled":    trigger.IsActive,
				"timeoutMs":  30000,
				"retryCount": 1,
			},
		}
		pipelines = append(pipelines, Pipeline{
			ID:          trigger.ID.String(),
			Name:        defaultTriggerName(trigger),
			Description: fmt.Sprintf("Trigger-driven pipeline for %s", trigger.TargetFunction),
			Status:      pipelineStatus(trigger.IsActive),
			Steps:       steps,
			Throughput:  float64(trigger.MaxInvocationsPerMinute) / 60.0,
			ErrorRate:   0,
			CreatedAt:   trigger.CreatedAt,
			UpdatedAt:   trigger.UpdatedAt,
		})
	}
	sort.Slice(pipelines, func(i, j int) bool { return pipelines[i].CreatedAt.After(pipelines[j].CreatedAt) })
	return pipelines, nil
}

func pipelineStatus(active bool) string {
	if active {
		return "active"
	}
	return "paused"
}

func defaultTriggerName(trigger *statestore.StateTrigger) string {
	if trigger.TargetFunction != "" {
		return trigger.TargetFunction
	}
	return fmt.Sprintf("trigger-%s", trigger.ID.String()[:8])
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (r *Repository) CreatePipeline(ctx context.Context, tenantID, fabricID uuid.UUID, name, description string, steps []map[string]interface{}) (*Pipeline, error) {
	_, err := r.GetFabric(ctx, tenantID, fabricID)
	if err != nil {
		return nil, err
	}
	trigger := &statestore.StateTrigger{
		TenantID:       tenantID,
		SourceStateID:  &fabricID,
		TriggerType:    "on_write",
		TargetFunction: name,
		Condition: statestore.JSONMap{
			"description": description,
			"steps":       steps,
		},
		IncludeNew:              true,
		MaxInvocationsPerMinute: 60,
		IsActive:                false,
	}
	created, err := r.stateRepo.CreateTrigger(ctx, trigger)
	if err != nil {
		return nil, err
	}
	pipeline := Pipeline{
		ID:          created.ID.String(),
		Name:        name,
		Description: description,
		Status:      "draft",
		Steps:       steps,
		CreatedAt:   created.CreatedAt,
		UpdatedAt:   created.UpdatedAt,
	}
	return &pipeline, nil
}

func (r *Repository) UpdatePipeline(ctx context.Context, tenantID, fabricID, pipelineID uuid.UUID, updates map[string]interface{}) (*Pipeline, error) {
	trigger, err := r.stateRepo.GetTrigger(ctx, pipelineID)
	if err != nil {
		return nil, err
	}
	if trigger.TenantID != tenantID || trigger.SourceStateID == nil || *trigger.SourceStateID != fabricID {
		return nil, fmt.Errorf("pipeline not found")
	}
	if name, ok := updates["name"].(string); ok && strings.TrimSpace(name) != "" {
		trigger.TargetFunction = name
	}
	if description, ok := updates["description"].(string); ok {
		condition := safeJSONMap(trigger.Condition)
		condition["description"] = description
		trigger.Condition = statestore.JSONMap(condition)
	}
	if status, ok := updates["status"].(string); ok {
		trigger.IsActive = status == "active"
	}
	if steps, ok := updates["steps"].([]map[string]interface{}); ok {
		condition := safeJSONMap(trigger.Condition)
		condition["steps"] = steps
		trigger.Condition = statestore.JSONMap(condition)
	}
	updated, err := r.stateRepo.UpdateTrigger(ctx, trigger)
	if err != nil {
		return nil, err
	}
	pipeline := Pipeline{
		ID:          updated.ID.String(),
		Name:        defaultTriggerName(updated),
		Description: stringFromAny(safeJSONMap(updated.Condition)["description"]),
		Status:      pipelineStatus(updated.IsActive),
		Steps:       stepsFromCondition(updated.Condition),
		CreatedAt:   updated.CreatedAt,
		UpdatedAt:   updated.UpdatedAt,
	}
	return &pipeline, nil
}

func stringFromAny(value interface{}) string {
	if s, ok := value.(string); ok {
		return s
	}
	return ""
}

func stepsFromCondition(condition statestore.JSONMap) []map[string]interface{} {
	stepsAny, ok := safeJSONMap(condition)["steps"]
	if !ok {
		return nil
	}
	stepsSlice, ok := stepsAny.([]interface{})
	if !ok {
		if typed, ok := stepsAny.([]map[string]interface{}); ok {
			return typed
		}
		return nil
	}
	steps := make([]map[string]interface{}, 0, len(stepsSlice))
	for _, step := range stepsSlice {
		if m, ok := step.(map[string]interface{}); ok {
			steps = append(steps, m)
		}
	}
	return steps
}

func (r *Repository) DeletePipeline(ctx context.Context, tenantID, fabricID, pipelineID uuid.UUID) error {
	trigger, err := r.stateRepo.GetTrigger(ctx, pipelineID)
	if err != nil {
		return err
	}
	if trigger.TenantID != tenantID || trigger.SourceStateID == nil || *trigger.SourceStateID != fabricID {
		return fmt.Errorf("pipeline not found")
	}
	return r.stateRepo.DeleteTrigger(ctx, pipelineID)
}

func (r *Repository) ExecutePipeline(ctx context.Context, tenantID, fabricID, pipelineID uuid.UUID, input map[string]interface{}) (map[string]interface{}, error) {
	trigger, err := r.stateRepo.GetTrigger(ctx, pipelineID)
	if err != nil {
		return nil, err
	}
	if trigger.TenantID != tenantID || trigger.SourceStateID == nil || *trigger.SourceStateID != fabricID {
		return nil, fmt.Errorf("pipeline not found")
	}
	return map[string]interface{}{
		"executionId": uuid.New().String(),
		"status":      "completed",
		"input":       input,
		"pipelineId":  pipelineID.String(),
	}, nil
}

func (r *Repository) ListEvents(ctx context.Context, tenantID, fabricID uuid.UUID, opts EventListOptions) ([]EventLog, int64, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, 0, err
	}
	if state.TenantID != tenantID {
		return nil, 0, fmt.Errorf("state fabric not found")
	}
	query := r.db.WithContext(ctx).Model(&statestore.StateEvent{}).Where("state_id = ?", fabricID)
	if opts.EventType != "" {
		query = query.Where("event_type = ?", mapEventTypeToStateEvent(opts.EventType))
	}
	if opts.StartTime != nil {
		query = query.Where("timestamp >= ?", *opts.StartTime)
	}
	if opts.EndTime != nil {
		query = query.Where("timestamp <= ?", *opts.EndTime)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = 100
	}
	var events []statestore.StateEvent
	if err := query.Order("sequence_num DESC").Limit(limit).Offset(opts.Offset).Find(&events).Error; err != nil {
		return nil, 0, err
	}
	items := make([]EventLog, 0, len(events))
	for _, event := range events {
		payload := map[string]interface{}{}
		if event.NewValue != nil {
			payload["newValue"] = map[string]interface{}(*event.NewValue)
		}
		if event.PreviousValue != nil {
			payload["previousValue"] = map[string]interface{}(*event.PreviousValue)
		}
		if event.Key != nil {
			payload["key"] = *event.Key
		}
		items = append(items, EventLog{
			ID:             event.ID.String(),
			FabricID:       fabricID.String(),
			EventType:      mapStateEventType(event.EventType),
			Payload:        payload,
			Timestamp:      event.Timestamp,
			SequenceNumber: event.SequenceNum,
			CorrelationID:  event.CorrelationID,
		})
	}
	return items, total, nil
}

func mapStateEventType(eventType string) string {
	switch eventType {
	case "set":
		return "update"
	case "restore":
		return "sync"
	default:
		return eventType
	}
}

func mapEventTypeToStateEvent(eventType string) string {
	switch eventType {
	case "update":
		return "set"
	case "sync":
		return "restore"
	default:
		return eventType
	}
}

func (r *Repository) ListSnapshots(ctx context.Context, tenantID, fabricID uuid.UUID) ([]Snapshot, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf("state fabric not found")
	}
	snapshots, _, err := r.stateRepo.ListSnapshots(ctx, fabricID, 100, 0)
	if err != nil {
		return nil, err
	}
	items := make([]Snapshot, 0, len(snapshots))
	for _, snapshot := range snapshots {
		name := fmt.Sprintf("snapshot-v%d", snapshot.SnapshotVersion)
		if snapshot.Label != nil && *snapshot.Label != "" {
			name = *snapshot.Label
		}
		items = append(items, Snapshot{
			ID:          snapshot.ID.String(),
			FabricID:    fabricID.String(),
			Name:        name,
			Description: "",
			State:       map[string]interface{}(snapshot.StateData),
			EventCount:  snapshot.KeyCount,
			SizeBytes:   snapshot.StateSizeBytes,
			CreatedAt:   snapshot.CreatedAt,
		})
	}
	return items, nil
}

func (r *Repository) CreateSnapshot(ctx context.Context, tenantID, fabricID uuid.UUID, name string) (*Snapshot, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf("state fabric not found")
	}
	created, err := r.stateRepo.CreateSnapshot(ctx, fabricID, name)
	if err != nil {
		return nil, err
	}
	snapshotName := name
	if snapshotName == "" {
		snapshotName = fmt.Sprintf("snapshot-v%d", created.SnapshotVersion)
	}
	return &Snapshot{
		ID:         created.ID.String(),
		FabricID:   fabricID.String(),
		Name:       snapshotName,
		State:      map[string]interface{}(created.StateData),
		EventCount: created.KeyCount,
		SizeBytes:  created.StateSizeBytes,
		CreatedAt:  created.CreatedAt,
	}, nil
}

func (r *Repository) DeleteSnapshot(ctx context.Context, tenantID, fabricID, snapshotID uuid.UUID) error {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return err
	}
	if state.TenantID != tenantID {
		return fmt.Errorf("state fabric not found")
	}
	result := r.db.WithContext(ctx).Delete(&statestore.StateSnapshot{}, "id = ? AND state_id = ?", snapshotID, fabricID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("snapshot not found")
	}
	return nil
}

func (r *Repository) ListReplays(ctx context.Context, tenantID, fabricID uuid.UUID) ([]ReplaySession, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf("state fabric not found")
	}
	snapshots, err := r.ListSnapshots(ctx, tenantID, fabricID)
	if err != nil {
		return nil, err
	}
	items := make([]ReplaySession, 0, len(snapshots))
	for _, snapshot := range snapshots {
		items = append(items, ReplaySession{
			ID:             snapshot.ID,
			FabricID:       fabricID.String(),
			SnapshotID:     snapshot.ID,
			Status:         "completed",
			Progress:       100,
			EventsReplayed: snapshot.EventCount,
			StartedAt:      snapshot.CreatedAt,
			CompletedAt:    &snapshot.CreatedAt,
		})
	}
	return items, nil
}

func (r *Repository) CreateReplay(ctx context.Context, tenantID, fabricID uuid.UUID, req ReplayCreateRequest) (*ReplaySession, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf("state fabric not found")
	}
	now := time.Now()
	completed := now
	return &ReplaySession{
		ID:             uuid.New().String(),
		FabricID:       fabricID.String(),
		SnapshotID:     req.SnapshotID,
		StartEventID:   req.StartEventID,
		EndEventID:     req.EndEventID,
		Status:         "completed",
		Progress:       100,
		EventsReplayed: 0,
		StartedAt:      now,
		CompletedAt:    &completed,
	}, nil
}

func (r *Repository) GetReplay(ctx context.Context, tenantID, fabricID uuid.UUID, replayID string) (*ReplaySession, error) {
	replays, err := r.ListReplays(ctx, tenantID, fabricID)
	if err != nil {
		return nil, err
	}
	for _, replay := range replays {
		if replay.ID == replayID {
			copy := replay
			return &copy, nil
		}
	}
	return nil, fmt.Errorf("replay not found")
}

// ListFabricsAdmin lists all fabrics for admin (optional tenant filter).
func (r *Repository) ListFabricsAdmin(ctx context.Context, tenantID *uuid.UUID, status string, limit, offset int) ([]Fabric, int64, error) {
	return r.ListAllFabrics(ctx, limit, offset, tenantID, status)
}

// ListStoresByFabric returns the store(s) for a fabric by ID (admin, no tenant check).
func (r *Repository) ListStoresByFabric(ctx context.Context, fabricID uuid.UUID) ([]FabricStore, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	return []FabricStore{buildStore(state)}, nil
}

// ListPipelinesByFabric returns pipelines for a fabric by ID.
func (r *Repository) ListPipelinesByFabric(ctx context.Context, fabricID uuid.UUID) ([]Pipeline, error) {
	return r.ListPipelines(ctx, fabricID)
}

// GetFabricByID returns a fabric by ID without tenant check (admin).
func (r *Repository) GetFabricByID(ctx context.Context, fabricID uuid.UUID) (*Fabric, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	metrics, _ := r.GetMetrics(ctx, state.ID, "")
	pipelines, _ := r.ListPipelines(ctx, state.ID)
	fabric := buildFabric(state, metrics, pipelines)
	return &fabric, nil
}

// GetAdminStats returns admin dashboard counts.
func (r *Repository) GetAdminStats(ctx context.Context) (totalFabrics, activeFabrics, totalStores, totalPipelines, totalEvents, storageUsed int64, err error) {
	if err = r.db.WithContext(ctx).Model(&statestore.State{}).Count(&totalFabrics).Error; err != nil {
		return 0, 0, 0, 0, 0, 0, err
	}
	activeFabrics = totalFabrics
	totalStores = totalFabrics
	if err = r.db.WithContext(ctx).Model(&statestore.StateTrigger{}).Count(&totalPipelines).Error; err != nil {
		return 0, 0, 0, 0, 0, 0, err
	}
	if err = r.db.WithContext(ctx).Model(&statestore.StateEvent{}).Count(&totalEvents).Error; err != nil {
		return 0, 0, 0, 0, 0, 0, err
	}
	if err = r.db.WithContext(ctx).Model(&statestore.State{}).Select("COALESCE(SUM(storage_used_mb), 0)").Scan(&storageUsed).Error; err != nil {
		return 0, 0, 0, 0, 0, 0, err
	}
	return totalFabrics, activeFabrics, totalStores, totalPipelines, totalEvents, storageUsed, nil
}

func (r *Repository) ListAllFabrics(ctx context.Context, limit, offset int, tenantID *uuid.UUID, status string) ([]Fabric, int64, error) {
	query := r.db.WithContext(ctx).Model(&statestore.State{})
	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 100
	}
	var states []statestore.State
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&states).Error; err != nil {
		return nil, 0, err
	}
	items := make([]Fabric, 0, len(states))
	for i := range states {
		state := states[i]
		if status != "" && status != "all" && stateStatus(&state) != status && !(status == "suspended" && stateStatus(&state) == "offline") {
			continue
		}
		metrics, _ := r.GetMetrics(ctx, state.ID, "")
		items = append(items, buildFabric(&state, metrics, nil))
	}
	return items, total, nil
}

func (r *Repository) Stats(ctx context.Context) (map[string]int64, error) {
	stats := map[string]int64{}
	var total int64
	if err := r.db.WithContext(ctx).Model(&statestore.State{}).Count(&total).Error; err != nil {
		return nil, err
	}
	stats["totalFabrics"] = total
	stats["activeFabrics"] = total
	var stores int64
	stats["totalStores"] = total
	var pipelines int64
	if err := r.db.WithContext(ctx).Model(&statestore.StateTrigger{}).Count(&pipelines).Error; err != nil {
		return nil, err
	}
	stats["totalPipelines"] = pipelines
	var events int64
	if err := r.db.WithContext(ctx).Model(&statestore.StateEvent{}).Count(&events).Error; err != nil {
		return nil, err
	}
	stats["totalEvents"] = events
	var storageUsed int64
	if err := r.db.WithContext(ctx).Model(&statestore.State{}).Select("COALESCE(SUM(storage_used_mb), 0)").Scan(&storageUsed).Error; err != nil {
		return nil, err
	}
	stats["storageUsed"] = storageUsed
	_ = stores
	return stats, nil
}

func (r *Repository) SetFabricSuspended(ctx context.Context, fabricID uuid.UUID, suspended bool, reason string) error {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return err
	}
	description := normalizeDescription(state.Description)
	if suspended {
		description = strings.TrimSpace(description + "\n[SUSPENDED] " + reason)
	} else {
		description = strings.ReplaceAll(description, "[SUSPENDED] "+reason, "")
		description = strings.TrimSpace(description)
	}
	state.Description = stringPtr(description)
	_, err = r.stateRepo.UpdateState(ctx, state)
	return err
}

// platformSettingsRow is used to scan the single settings row.
type platformSettingsRow struct {
	Config []byte `gorm:"column:config;type:jsonb"`
}

// GetPlatformSettings returns the platform-wide state fabric settings (single row).
func (r *Repository) GetPlatformSettings(ctx context.Context) (map[string]interface{}, error) {
	var row platformSettingsRow
	err := r.db.WithContext(ctx).Raw(
		"SELECT config FROM state_fabric_platform_settings WHERE id = 1",
	).Scan(&row).Error
	if err != nil {
		return nil, fmt.Errorf("get platform settings: %w", err)
	}
	if len(row.Config) == 0 {
		return defaultPlatformSettings(), nil
	}
	var out map[string]interface{}
	if err := json.Unmarshal(row.Config, &out); err != nil {
		return defaultPlatformSettings(), nil
	}
	return mergeWithDefaults(out), nil
}

// UpdatePlatformSettings updates the platform-wide state fabric settings.
func (r *Repository) UpdatePlatformSettings(ctx context.Context, config map[string]interface{}) error {
	if config == nil {
		config = defaultPlatformSettings()
	}
	merged := mergeWithDefaults(config)
	payload, err := json.Marshal(merged)
	if err != nil {
		return fmt.Errorf("marshal platform settings: %w", err)
	}
	res := r.db.WithContext(ctx).Exec(
		"UPDATE state_fabric_platform_settings SET config = $1, updated_at = NOW() WHERE id = 1",
		payload,
	)
	if res.Error != nil {
		return fmt.Errorf("update platform settings: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return r.db.WithContext(ctx).Exec(
			"INSERT INTO state_fabric_platform_settings (id, config) VALUES (1, $1) ON CONFLICT (id) DO UPDATE SET config = EXCLUDED.config, updated_at = NOW()",
			payload,
		).Error
	}
	return nil
}

func defaultPlatformSettings() map[string]interface{} {
	return map[string]interface{}{
		"maxFabricsPerTenant":          10,
		"defaultSnapshotRetentionDays": 30,
		"allowPublicPipelines":         false,
		"maintenanceMode":              false,
	}
}

func mergeWithDefaults(in map[string]interface{}) map[string]interface{} {
	def := defaultPlatformSettings()
	out := make(map[string]interface{}, len(def))
	for k, v := range def {
		out[k] = v
	}
	for k, v := range in {
		if v != nil {
			out[k] = v
		}
	}
	return out
}

func stringPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	v := value
	return &v
}
