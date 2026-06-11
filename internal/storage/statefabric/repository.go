package statefabric

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	statestore "github.com/functionfly/functionfly/internal/storage/state"
)

const (
	MaxStoreSizeBytes    = 500 * 1024 * 1024 * 1024 // 500 GB max store size
	MaxSnapshotSizeBytes = 1 * 1024 * 1024 * 1024   // 1 GB max snapshot size
	MaxEventListLimit    = 1000                     // Max events returned per query
)

type Repository struct {
	db        *gorm.DB
	stateRepo *statestore.StateRepository
	r2Backend *R2StorageBackend // Optional R2 backend for large data (events, snapshots, memory, replays)

	// Function execution client
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

func NewRepository(db *gorm.DB) *Repository {
	repo := &Repository{db: db, stateRepo: statestore.NewStateRepository(db)}

	// Initialize R2 backend if configured
	if IsR2StorageConfigured() {
		if r2Backend, err := NewR2StorageBackend(); err == nil {
			repo.r2Backend = r2Backend
		}
	}

	// Initialize HTTP client for function execution with TLS configuration
	repo.httpClient = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	return repo
}

// NewRepositoryWithR2 creates a repository with an explicit R2 backend (for testing or custom config)
func NewRepositoryWithR2(db *gorm.DB, r2Backend *R2StorageBackend) *Repository {
	return &Repository{
		db:        db,
		stateRepo: statestore.NewStateRepository(db),
		r2Backend: r2Backend,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			},
		},
	}
}

// ConfigureExecution sets the base URL and API key for function execution
func (r *Repository) ConfigureExecution(baseURL, apiKey string) {
	r.baseURL = strings.TrimSuffix(baseURL, "/")
	r.apiKey = apiKey
}

// R2Backend returns the R2 storage backend if configured
func (r *Repository) R2Backend() *R2StorageBackend {
	return r.r2Backend
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
	TenantID    uuid.UUID
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
	var storageType string
	switch fabricType {
	case "catalog":
		storageType = "document"
	case "workflow":
		storageType = "timeseries"
	case "custom":
		storageType = "graph"
	default:
		return nil, fmt.Errorf("unknown fabric type: %s", fabricType)
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
	if maxSize < 0 {
		return nil, fmt.Errorf("maxSize cannot be negative")
	}
	if maxSize > MaxStoreSizeBytes {
		return nil, fmt.Errorf("maxSize exceeds maximum allowed size of %d bytes", MaxStoreSizeBytes)
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

func (r *Repository) DeleteStore(ctx context.Context, tenantID, fabricID uuid.UUID, storeID string) error {
	state, err := r.GetFabric(ctx, tenantID, fabricID)
	if err != nil {
		return err
	}

	if storeID == "" {
		return fmt.Errorf("store ID is required")
	}

	storeUUID, err := uuid.Parse(storeID)
	if err != nil {
		return fmt.Errorf("invalid store ID: %w", err)
	}

	result := r.db.WithContext(ctx).Delete(&StateFabricStore{}, "id = ? AND fabric_id = ?", storeUUID, fabricID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete store: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("store not found")
	}

	logrus.WithFields(logrus.Fields{
		"store_id":   storeID,
		"fabric_id":  fabricID,
		"tenant_id":  tenantID,
		"rows_count": result.RowsAffected,
	}).Info("Store deleted from fabric")

	_ = state
	return nil
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

	executionID := uuid.New()
	execution := &StateFabricPipelineExecution{
		ID:        executionID,
		PipelineID: pipelineID,
		Status:    "running",
		InputData: input,
	}

	if err := r.db.WithContext(ctx).Create(execution).Error; err != nil {
		return nil, fmt.Errorf("failed to create execution record: %w", err)
	}

	steps := stepsFromCondition(trigger.Condition)
	if len(steps) == 0 {
		execution.Status = "completed"
		execution.OutputData = map[string]interface{}{"result": "no steps to execute"}
		if err := r.db.WithContext(ctx).Save(execution).Error; err != nil {
			return nil, fmt.Errorf("failed to update execution record: %w", err)
		}
		return map[string]interface{}{
			"executionId": executionID.String(),
			"status":      "completed",
			"input":       input,
			"pipelineId":  pipelineID.String(),
			"output":      execution.OutputData,
		}, nil
	}

	var lastOutput map[string]interface{}
	for i, step := range steps {
		stepName, _ := step["name"].(string)
		stepType, _ := step["type"].(string)

		config, ok := step["config"].(map[string]interface{})
		if !ok {
			config = map[string]interface{}{}
		}
		targetFunction, _ := config["targetFunction"].(string)

		timeoutMs := 30000
		if tm, ok := config["timeoutMs"].(float64); ok {
			timeoutMs = int(tm)
		}

		retryCount := 1
		if rc, ok := config["retryCount"].(float64); ok {
			retryCount = int(rc)
		}

		if targetFunction == "" {
			execution.Status = "failed"
			execution.OutputData = map[string]interface{}{
				"error":     fmt.Sprintf("step %d (%s) has no target function", i, stepName),
				"stepIndex": i,
			}
			if err := r.db.WithContext(ctx).Save(execution).Error; err != nil {
				return nil, fmt.Errorf("failed to update execution record: %w", err)
			}
			return nil, fmt.Errorf("step %d (%s) has no target function", i, stepName)
		}

		stepInput := input
		if i > 0 && lastOutput != nil {
			stepInput = lastOutput
		}

		var stepErr error
		var stepOutput map[string]interface{}
		for attempt := 0; attempt <= retryCount; attempt++ {
			if attempt > 0 {
				time.Sleep(time.Duration(attempt*100) * time.Millisecond)
			}

			stepOutput, stepErr = r.executeFunction(ctx, targetFunction, stepInput, timeoutMs)
			if stepErr == nil {
				break
			}
		}

		if stepErr != nil {
			execution.Status = "failed"
			execution.OutputData = map[string]interface{}{
				"error":      stepErr.Error(),
				"stepIndex":  i,
				"stepName":   stepName,
				"stepType":   stepType,
				"attempts":   retryCount + 1,
			}
			if err := r.db.WithContext(ctx).Save(execution).Error; err != nil {
				return nil, fmt.Errorf("failed to update execution record: %w", err)
			}
			return map[string]interface{}{
				"executionId": executionID.String(),
				"status":      "failed",
				"error":       stepErr.Error(),
				"stepIndex":   i,
				"stepName":    stepName,
				"pipelineId":  pipelineID.String(),
				"input":       input,
			}, stepErr
		}

		lastOutput = stepOutput
	}

	execution.Status = "completed"
	execution.OutputData = lastOutput
	if err := r.db.WithContext(ctx).Save(execution).Error; err != nil {
		return nil, fmt.Errorf("failed to update execution record: %w", err)
	}

	if trigger.LastTriggeredAt != nil {
		now := time.Now()
		trigger.LastTriggeredAt = &now
		if _, err := r.stateRepo.UpdateTrigger(ctx, trigger); err != nil {
			logrus.WithError(err).Warn("failed to update trigger last triggered time")
		}
	}

	return map[string]interface{}{
		"executionId": executionID.String(),
		"status":      "completed",
		"input":       input,
		"pipelineId":  pipelineID.String(),
		"output":      lastOutput,
	}, nil
}

func (r *Repository) executeFunction(ctx context.Context, targetFunction string, input map[string]interface{}, timeoutMs int) (map[string]interface{}, error) {
	if r.baseURL == "" || r.httpClient == nil {
		return nil, fmt.Errorf("pipeline executor not configured: baseURL or httpClient is missing")
	}

	var url string
	if strings.Contains(targetFunction, "/") {
		url = fmt.Sprintf("%s/v1/functions/by-name/%s/execute", r.baseURL, targetFunction)
	} else {
		url = fmt.Sprintf("%s/v1/functions/%s/execute", r.baseURL, targetFunction)
	}

	jsonInput, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonInput))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if r.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.apiKey))
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute function: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("function execution failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return map[string]interface{}{"raw": string(body)}, nil
	}

	return result, nil
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
	if limit > MaxEventListLimit {
		limit = MaxEventListLimit
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

	// Create snapshot in PostgreSQL
	created, err := r.stateRepo.CreateSnapshot(ctx, fabricID, name)
	if err != nil {
		return nil, err
	}

	// Check snapshot size limit after creation
	stateDataSize := estimateJSONSize(created.StateData)
	if stateDataSize > MaxSnapshotSizeBytes {
		// Delete the created snapshot since it exceeds size limit
		r.db.WithContext(ctx).Delete(&StateFabricSnapshot{}, "id = ?", created.ID)
		return nil, fmt.Errorf("snapshot size exceeds maximum allowed size of %d bytes", MaxSnapshotSizeBytes)
	}

	snapshotName := name
	if snapshotName == "" {
		snapshotName = fmt.Sprintf("snapshot-v%d", created.SnapshotVersion)
	}

	// If R2 backend is configured and snapshot data is large, offload to R2
	if r.r2Backend != nil && len(created.StateData) > 0 {
		// Calculate size threshold (100KB) for R2 offloading
		stateDataSize := estimateJSONSize(created.StateData)
		if stateDataSize > 100*1024 { // 100KB threshold
			snapshotData := JSONMap(created.StateData)
			metadata := JSONMap{
				"snapshot_version": created.SnapshotVersion,
				"key_count":        created.KeyCount,
				"original_size":    stateDataSize,
			}

			r2Object, err := r.r2Backend.StoreSnapshotData(ctx, tenantID, fabricID, created.ID, snapshotData, metadata)
			if err == nil && r2Object != nil {
				// Update the snapshot record with R2 reference (stored in state_fabric_snapshots table)
				r.db.WithContext(ctx).Model(&StateFabricSnapshot{}).Where("id = ?", created.ID).Updates(map[string]interface{}{
					"r2_object_key":   r2Object.Key,
					"r2_bucket":       r2Object.Bucket,
					"r2_content_hash": r2Object.ContentHash,
				})
			}
		}
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

// estimateJSONSize estimates the size of JSON data in bytes
func estimateJSONSize(data statestore.JSONMap) int {
	if data == nil {
		return 0
	}
	jsonBytes, _ := json.Marshal(data)
	return len(jsonBytes)
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
	var dbReplays []StateFabricReplay
	if err := r.db.WithContext(ctx).Where("fabric_id = ?", fabricID).Order("started_at DESC").Find(&dbReplays).Error; err != nil {
		return nil, err
	}
	items := make([]ReplaySession, 0, len(dbReplays))
	for _, replay := range dbReplays {
		var snapshotID, startEventID, endEventID string
		if replay.SnapshotID != nil {
			snapshotID = replay.SnapshotID.String()
		}
		if replay.StartEventID != nil {
			startEventID = replay.StartEventID.String()
		}
		if replay.EndEventID != nil {
			endEventID = replay.EndEventID.String()
		}
		items = append(items, ReplaySession{
			ID:             replay.ID.String(),
			FabricID:       fabricID.String(),
			SnapshotID:     snapshotID,
			StartEventID:   startEventID,
			EndEventID:     endEventID,
			Status:         replay.Status,
			Progress:       replay.Progress,
			EventsReplayed: int(replay.EventsReplayed),
			StartedAt:      replay.StartedAt,
			CompletedAt:    replay.CompletedAt,
			Error:          "",
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

	replayID := uuid.New()
	startedAt := time.Now()

	var snapshotUUID, startEventUUID, endEventUUID *uuid.UUID
	if req.SnapshotID != "" {
		if su, err := uuid.Parse(req.SnapshotID); err == nil {
			snapshotUUID = &su
		}
	}
	if req.StartEventID != "" {
		if su, err := uuid.Parse(req.StartEventID); err == nil {
			startEventUUID = &su
		}
	}
	if req.EndEventID != "" {
		if eu, err := uuid.Parse(req.EndEventID); err == nil {
			endEventUUID = &eu
		}
	}

	dbReplay := &StateFabricReplay{
		ID:           replayID,
		FabricID:     fabricID,
		SnapshotID:   snapshotUUID,
		StartEventID: startEventUUID,
		EndEventID:   endEventUUID,
		Status:       "pending",
		Progress:     0,
		EventsReplayed: 0,
		StartedAt:    startedAt,
	}
	if err := r.db.WithContext(ctx).Create(dbReplay).Error; err != nil {
		return nil, fmt.Errorf("failed to create replay record: %w", err)
	}

	session := &ReplaySession{
		ID:           replayID.String(),
		FabricID:     fabricID.String(),
		SnapshotID:   req.SnapshotID,
		StartEventID: req.StartEventID,
		EndEventID:   req.EndEventID,
		Status:       "running",
		Progress:     0,
		EventsReplayed: 0,
		StartedAt:    startedAt,
	}

	go r.executeReplay(replayID, fabricID, req)

	return session, nil
}

func (r *Repository) executeReplay(replayID, fabricID uuid.UUID, req ReplayCreateRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	logger := logrus.WithFields(logrus.Fields{"replay_id": replayID, "fabric_id": fabricID, "tenant_id": req.TenantID})

	var events []statestore.StateEvent
	var err error

	if req.SnapshotID != "" {
		events, err = r.getEventsFromSnapshot(ctx, fabricID, req.SnapshotID)
	} else {
		events, err = r.getEventsForReplay(ctx, fabricID, req.StartEventID, req.EndEventID)
	}

	if err != nil {
		r.updateReplayStatus(ctx, replayID, "failed", 0, 0, err.Error())
		logger.WithError(err).Error("replay failed to fetch events")
		return
	}

	totalEvents := int64(len(events))
	processed := int64(0)

	for i, event := range events {
		if err := r.applyEventToState(ctx, fabricID, event); err != nil {
			logger.WithError(err).Warnf("failed to apply event %s, continuing", event.ID)
		}
		processed++
		progress := int(float64(i+1) / float64(totalEvents) * 100)
		if i%100 == 0 || i == len(events)-1 {
			r.updateReplayProgress(ctx, replayID, progress, processed)
		}
	}

	var eventsReplayed int64
	if r.r2Backend != nil && len(events) > 0 {
		metadata := JSONMap{
			"fabric_id":    fabricID.String(),
			"snapshot_id":  req.SnapshotID,
			"start_event":  req.StartEventID,
			"end_event":    req.EndEventID,
			"total_events": len(events),
		}
		if obj, storeErr := r.r2Backend.StoreReplayData(ctx, uuid.Nil, replayID, events, metadata); storeErr == nil {
			r.db.WithContext(ctx).Model(&StateFabricReplay{}).Where("id = ?", replayID).Updates(map[string]interface{}{
				"r2_object_key":   obj.Key,
				"r2_bucket":       obj.Bucket,
				"r2_content_hash": obj.ContentHash,
			})
		}
		eventsReplayed = int64(len(events))
	}

	r.updateReplayStatus(ctx, replayID, "completed", 100, eventsReplayed, "")
	logger.Infof("replay completed: %d events processed", eventsReplayed)
}

func (r *Repository) getEventsFromSnapshot(ctx context.Context, fabricID uuid.UUID, snapshotID string) ([]statestore.StateEvent, error) {
	snapshots, _, err := r.stateRepo.ListSnapshots(ctx, fabricID, 100, 0)
	if err != nil {
		return nil, err
	}
	var target *statestore.StateSnapshot
	for _, s := range snapshots {
		if s.ID.String() == snapshotID {
			target = s
			break
		}
	}
	if target == nil {
		return nil, fmt.Errorf("snapshot not found")
	}

	var firstSeq, lastSeq int64
	r.db.WithContext(ctx).Model(&statestore.StateEvent{}).Select("COALESCE(MIN(sequence_num), 0)").Where("state_id = ?", fabricID).Scan(&firstSeq)
	r.db.WithContext(ctx).Model(&statestore.StateEvent{}).Select("COALESCE(MAX(sequence_num), 0)").Where("state_id = ?", fabricID).Scan(&lastSeq)

	if target.FirstSequence > 0 && target.LastSequence > 0 {
		firstSeq, lastSeq = target.FirstSequence, target.LastSequence
	}

	var events []statestore.StateEvent
	err = r.db.WithContext(ctx).
		Where("state_id = ? AND sequence_num >= ? AND sequence_num <= ?", fabricID, firstSeq, lastSeq).
		Order("sequence_num ASC").
		Find(&events).Error
	return events, err
}

func (r *Repository) getEventsForReplay(ctx context.Context, fabricID uuid.UUID, startEventID, endEventID string) ([]statestore.StateEvent, error) {
	query := r.db.WithContext(ctx).Model(&statestore.StateEvent{}).Where("state_id = ?", fabricID)

	if startEventID != "" {
		if startUUID, err := uuid.Parse(startEventID); err == nil {
			var seq int64
			r.db.WithContext(ctx).Model(&statestore.StateEvent{}).Select("sequence_num").Where("id = ?", startUUID).Scan(&seq)
			if seq > 0 {
				query = query.Where("sequence_num >= ?", seq)
			}
		}
	}
	if endEventID != "" {
		if endUUID, err := uuid.Parse(endEventID); err == nil {
			var seq int64
			r.db.WithContext(ctx).Model(&statestore.StateEvent{}).Select("sequence_num").Where("id = ?", endUUID).Scan(&seq)
			if seq > 0 {
				query = query.Where("sequence_num <= ?", seq)
			}
		}
	}

	var events []statestore.StateEvent
	err := query.Order("sequence_num ASC").Find(&events).Error
	return events, err
}

func (r *Repository) applyEventToState(ctx context.Context, fabricID uuid.UUID, event statestore.StateEvent) error {
	switch event.EventType {
	case "set":
		if event.NewValue != nil && event.Key != nil {
			_, err := r.stateRepo.SetStateValue(ctx, fabricID, *event.Key, *event.NewValue, "replay", event.ID.String())
			return err
		}
	case "delete":
		if event.Key != nil {
			return r.stateRepo.DeleteStateValue(ctx, fabricID, *event.Key, "replay", event.ID.String())
		}
	}
	return nil
}

func (r *Repository) updateReplayStatus(ctx context.Context, replayID uuid.UUID, status string, progress int, eventsReplayed int64, errMsg string) {
	updates := map[string]interface{}{
		"status":          status,
		"progress":        progress,
		"events_replayed": eventsReplayed,
	}
	if status == "completed" || status == "failed" {
		now := time.Now()
		updates["completed_at"] = &now
	}
	if errMsg != "" {
		updates["error_message"] = &errMsg
	}
	r.db.WithContext(ctx).Model(&StateFabricReplay{}).Where("id = ?", replayID).Updates(updates)
}

func (r *Repository) updateReplayProgress(ctx context.Context, replayID uuid.UUID, progress int, eventsReplayed int64) {
	r.db.WithContext(ctx).Model(&StateFabricReplay{}).Where("id = ?", replayID).Updates(map[string]interface{}{
		"progress":        progress,
		"events_replayed": eventsReplayed,
	})
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

// ArchiveEventsToR2 archives a batch of events to R2 storage for long-term retention.
// This is typically called by a background job or when events exceed local retention limits.
func (r *Repository) ArchiveEventsToR2(ctx context.Context, tenantID, fabricID uuid.UUID, batchID string, events []statestore.StateEvent) error {
	if r.r2Backend == nil {
		return fmt.Errorf("R2 backend not configured")
	}
	if len(events) == 0 {
		return nil
	}

	// Store events in R2
	r2Object, err := r.r2Backend.StoreEventLogs(ctx, tenantID, fabricID, events)
	if err != nil {
		return fmt.Errorf("failed to store events in R2: %w", err)
	}
	if r2Object == nil {
		return nil
	}

	// Mark events as archived in PostgreSQL using transaction
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		for _, event := range events {
			if err := tx.Model(&statestore.StateEvent{}).Where("id = ?", event.ID).Updates(map[string]interface{}{
				"is_archived": true,
				"archived_at": now,
				"r2_object_key": r2Object.Key,
				"r2_bucket":     r2Object.Bucket,
				"batch_id":      batchID,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		// Log the partial failure - events are in R2 but DB update failed
		logrus.WithError(err).WithFields(logrus.Fields{
			"tenant_id":  tenantID,
			"fabric_id":  fabricID,
			"batch_id":   batchID,
			"event_count": len(events),
			"r2_key":     r2Object.Key,
		}).Error("Failed to update event archival status in DB - R2 data may be orphaned")
		return fmt.Errorf("failed to update event archival status: %w", err)
	}

	return nil
}

// RestoreSnapshotFromR2 retrieves snapshot data from R2 if it's been offloaded.
// This is used when the snapshot data in PostgreSQL is empty but R2 reference exists.
func (r *Repository) RestoreSnapshotFromR2(ctx context.Context, tenantID, snapshotID uuid.UUID) (JSONMap, error) {
	if r.r2Backend == nil {
		return nil, fmt.Errorf("R2 backend not configured")
	}

	// Find the R2 object key for this snapshot
	var snapshot StateFabricSnapshot
	if err := r.db.WithContext(ctx).Where("id = ? AND fabric_id = ?", snapshotID, tenantID).First(&snapshot).Error; err != nil {
		return nil, err
	}

	if snapshot.R2ObjectKey == nil || *snapshot.R2ObjectKey == "" {
		return nil, fmt.Errorf("snapshot not found in R2")
	}

	return r.r2Backend.GetSnapshotData(ctx, *snapshot.R2ObjectKey)
}

// StoreMemoryBlobToR2 stores a memory blob to R2 for large memory content.
func (r *Repository) StoreMemoryBlobToR2(ctx context.Context, tenantID, memoryID uuid.UUID, content []byte, memoryType string, metadata JSONMap) (*R2StorageObject, error) {
	if r.r2Backend == nil {
		return nil, fmt.Errorf("R2 backend not configured")
	}

	r2Object, err := r.r2Backend.StoreMemoryBlob(ctx, tenantID, memoryID, content, memoryType, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to store memory blob in R2: %w", err)
	}

	// Update the agent memory record with R2 reference
	r.db.WithContext(ctx).Model(&statestore.AgentMemory{}).Where("id = ?", memoryID).Updates(map[string]interface{}{
		"r2_object_key":   r2Object.Key,
		"r2_bucket":       r2Object.Bucket,
		"r2_content_hash": r2Object.ContentHash,
		"is_offloaded":    true,
		"offloaded_at":    time.Now(),
	})

	return r2Object, nil
}

// GetMemoryBlobFromR2 retrieves a memory blob from R2.
func (r *Repository) GetMemoryBlobFromR2(ctx context.Context, memoryID uuid.UUID) ([]byte, error) {
	if r.r2Backend == nil {
		return nil, fmt.Errorf("R2 backend not configured")
	}

	// Find the R2 object key for this memory
	var memory statestore.AgentMemory
	if err := r.db.WithContext(ctx).Where("id = ?", memoryID).First(&memory).Error; err != nil {
		return nil, err
	}

	if memory.R2ObjectKey == nil || *memory.R2ObjectKey == "" {
		return nil, fmt.Errorf("memory blob not found in R2")
	}

	return r.r2Backend.GetMemoryBlob(ctx, *memory.R2ObjectKey)
}

// StoreReplayDataToR2 stores replay session data to R2.
func (r *Repository) StoreReplayDataToR2(ctx context.Context, tenantID, replayID uuid.UUID, events []statestore.StateEvent, metadata JSONMap) (*R2StorageObject, error) {
	if r.r2Backend == nil {
		return nil, fmt.Errorf("R2 backend not configured")
	}

	r2Object, err := r.r2Backend.StoreReplayData(ctx, tenantID, replayID, events, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to store replay data in R2: %w", err)
	}

	// Update the replay record with R2 reference
	r.db.WithContext(ctx).Model(&StateFabricReplay{}).Where("id = ?", replayID).Updates(map[string]interface{}{
		"r2_object_key":   r2Object.Key,
		"r2_bucket":       r2Object.Bucket,
		"r2_content_hash": r2Object.ContentHash,
	})

	return r2Object, nil
}

// GetReplayDataFromR2 retrieves replay session data from R2.
func (r *Repository) GetReplayDataFromR2(ctx context.Context, replayID uuid.UUID) (*ReplayData, error) {
	if r.r2Backend == nil {
		return nil, fmt.Errorf("R2 backend not configured")
	}

	// Find the R2 object key for this replay
	var replay StateFabricReplay
	if err := r.db.WithContext(ctx).Where("id = ?", replayID).First(&replay).Error; err != nil {
		return nil, err
	}

	if replay.R2ObjectKey == nil || *replay.R2ObjectKey == "" {
		return nil, fmt.Errorf("replay data not found in R2")
	}

	return r.r2Backend.GetReplayData(ctx, *replay.R2ObjectKey)
}

// StateFabricAuditEvent represents an audit log entry for state fabric operations
type StateFabricAuditEvent struct {
	ID            uuid.UUID              `json:"id"`
	TenantID      uuid.UUID              `json:"tenant_id"`
	FabricID      uuid.UUID              `json:"fabric_id"`
	StoreID       *uuid.UUID             `json:"store_id,omitempty"`
	Action        string                 `json:"action"` // create, read, update, delete, execute, snapshot, replay
	ActorUserID   *uuid.UUID             `json:"actor_user_id,omitempty"`
	ActorEmail    string                 `json:"actor_email,omitempty"`
	ResourceType  string                 `json:"resource_type"` // fabric, store, pipeline, snapshot, replay, edge_state
	ResourceID    *uuid.UUID             `json:"resource_id,omitempty"`
	RequestID     string                 `json:"request_id,omitempty"`
	BeforeState   map[string]interface{} `json:"before_state,omitempty"`
	AfterState    map[string]interface{} `json:"after_state,omitempty"`
	IPAddress     string                 `json:"ip_address,omitempty"`
	UserAgent     string                 `json:"user_agent,omitempty"`
	Timestamp     time.Time             `json:"timestamp"`
	Success       bool                   `json:"success"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	OperationType string                 `json:"operation_type"` // mutation, query, execution
}

// LogStateFabricAudit logs a state fabric audit event
func (r *Repository) LogStateFabricAudit(ctx context.Context, event *StateFabricAuditEvent) error {
	event.ID = uuid.New()
	event.Timestamp = time.Now()

	query := `
		INSERT INTO state_fabric_audit_log (id, tenant_id, fabric_id, store_id, action, actor_user_id, actor_email,
			resource_type, resource_id, request_id, before_state, after_state, ip_address, user_agent, timestamp, success, error_message, operation_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`

	var beforeState, afterState []byte
	if event.BeforeState != nil {
		beforeState, _ = json.Marshal(event.BeforeState)
	}
	if event.AfterState != nil {
		afterState, _ = json.Marshal(event.AfterState)
	}

	result := r.db.WithContext(ctx).Exec(query,
		event.ID, event.TenantID, event.FabricID, event.StoreID, event.Action,
		event.ActorUserID, event.ActorEmail, event.ResourceType, event.ResourceID,
		event.RequestID, beforeState, afterState, event.IPAddress, event.UserAgent,
		event.Timestamp, event.Success, event.ErrorMessage, event.OperationType)

	if result.Error != nil {
		logrus.WithError(result.Error).WithFields(logrus.Fields{
			"action":   event.Action,
			"fabric_id": event.FabricID,
			"tenant_id": event.TenantID,
		}).Error("Failed to log state fabric audit event")
		return result.Error
	}

	return nil
}

// LogStateFabricMutation logs a state fabric mutation (create, update, delete)
func (r *Repository) LogStateFabricMutation(ctx context.Context, tenantID, fabricID uuid.UUID, action, resourceType string, beforeState, afterState map[string]interface{}, userID *uuid.UUID, userEmail, ipAddress string) error {
	event := &StateFabricAuditEvent{
		TenantID:      tenantID,
		FabricID:      fabricID,
		Action:        action,
		ActorUserID:   userID,
		ActorEmail:    userEmail,
		ResourceType:  resourceType,
		BeforeState:   beforeState,
		AfterState:    afterState,
		IPAddress:     ipAddress,
		Success:       true,
		OperationType: "mutation",
	}
	return r.LogStateFabricAudit(ctx, event)
}

// LogStateFabricQuery logs a state fabric query operation (read, list)
func (r *Repository) LogStateFabricQuery(ctx context.Context, tenantID, fabricID uuid.UUID, resourceType string, userID *uuid.UUID, userEmail, ipAddress string) error {
	event := &StateFabricAuditEvent{
		TenantID:      tenantID,
		FabricID:      fabricID,
		Action:        "read",
		ActorUserID:   userID,
		ActorEmail:    userEmail,
		ResourceType:  resourceType,
		IPAddress:     ipAddress,
		Success:       true,
		OperationType: "query",
	}
	return r.LogStateFabricAudit(ctx, event)
}

// LogStateFabricExecution logs a state fabric execution operation (pipeline execute, snapshot, replay)
func (r *Repository) LogStateFabricExecution(ctx context.Context, tenantID, fabricID uuid.UUID, action, resourceType string, beforeState, afterState map[string]interface{}, userID *uuid.UUID, userEmail, ipAddress string, success bool, errorMsg string) error {
	event := &StateFabricAuditEvent{
		TenantID:      tenantID,
		FabricID:      fabricID,
		Action:        action,
		ActorUserID:   userID,
		ActorEmail:    userEmail,
		ResourceType:  resourceType,
		BeforeState:   beforeState,
		AfterState:    afterState,
		IPAddress:     ipAddress,
		Success:       success,
		ErrorMessage:  errorMsg,
		OperationType: "execution",
	}
	return r.LogStateFabricAudit(ctx, event)
}

// ListStateFabricAuditLogs lists audit logs for a state fabric with filtering
func (r *Repository) ListStateFabricAuditLogs(ctx context.Context, tenantID, fabricID uuid.UUID, limit, offset int, action, resourceType string) ([]StateFabricAuditEvent, int64, error) {
	query := `SELECT id, tenant_id, fabric_id, store_id, action, actor_user_id, actor_email, resource_type,
		resource_id, request_id, before_state, after_state, ip_address, user_agent, timestamp, success, error_message, operation_type
		FROM state_fabric_audit_log WHERE tenant_id = $1 AND fabric_id = $2`
	countQuery := `SELECT COUNT(*) FROM state_fabric_audit_log WHERE tenant_id = $1 AND fabric_id = $2`

	args := []interface{}{tenantID, fabricID}
	argIndex := 3

	if action != "" {
		query += fmt.Sprintf(" AND action = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND action = $%d", argIndex)
		args = append(args, action)
		argIndex++
	}
	if resourceType != "" {
		query += fmt.Sprintf(" AND resource_type = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND resource_type = $%d", argIndex)
		args = append(args, resourceType)
		argIndex++
	}

	var total int64
	if err := r.db.WithContext(ctx).Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var events []StateFabricAuditEvent
	for rows.Next() {
		var event StateFabricAuditEvent
		var beforeState, afterState []byte
		var storeID, actorUserID, resourceID *uuid.UUID
		var actorEmail, ipAddress, userAgent, errorMsg sql.NullString

		err := rows.Scan(
			&event.ID, &event.TenantID, &event.FabricID, &storeID, &event.Action,
			&actorUserID, &actorEmail, &event.ResourceType, &resourceID,
			&event.RequestID, &beforeState, &afterState, &ipAddress, &userAgent,
			&event.Timestamp, &event.Success, &errorMsg, &event.OperationType)
		if err != nil {
			return nil, 0, err
		}

		if storeID != nil {
			event.StoreID = storeID
		}
		if actorUserID != nil {
			event.ActorUserID = actorUserID
		}
		if actorEmail.Valid {
			event.ActorEmail = actorEmail.String
		}
		if resourceID != nil {
			event.ResourceID = resourceID
		}
		if ipAddress.Valid {
			event.IPAddress = ipAddress.String
		}
		if userAgent.Valid {
			event.UserAgent = userAgent.String
		}
		if errorMsg.Valid {
			event.ErrorMessage = errorMsg.String
		}
		if len(beforeState) > 0 {
			_ = json.Unmarshal(beforeState, &event.BeforeState)
		}
		if len(afterState) > 0 {
			_ = json.Unmarshal(afterState, &event.AfterState)
		}

		events = append(events, event)
	}

	return events, total, nil
}
