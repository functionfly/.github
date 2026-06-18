package statefabric

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/functionfly/functionfly/internal/agent/circuitbreaker"
	"github.com/functionfly/functionfly/internal/api/middleware"
	statestore "github.com/functionfly/functionfly/internal/storage/state"
	vault "github.com/functionfly/functionfly/internal/storage/vault"
)

const (
	defaultmaxStoreSizeBytes = 500 * 1024 * 1024 * 1024 // 500 GB max store size
	defaultHTTPTimeout       = 30 * time.Second
)

const (
	ErrFabricNotFound   = "state fabric not found"
	ErrSnapshotNotFound = "snapshot not found"
	ErrReplayNotFound   = "replay not found"
	ErrTriggerNotFound  = "trigger not found"
)

var (
	maxStoreSizeBytes    uint64
	httpTimeout          time.Duration
	circuitBreakerConfig circuitbreaker.Config
)

func init() {
	maxStoreSizeBytes = defaultmaxStoreSizeBytes
	if val := os.Getenv("STATEFABRIC_MAX_STORE_SIZE"); val != "" {
		if parsed, err := parseBytes(val); err == nil {
			maxStoreSizeBytes = parsed
		}
	}

	httpTimeout = defaultHTTPTimeout
	if val := os.Getenv("STATEFABRIC_HTTP_TIMEOUT"); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil && parsed > 0 {
			httpTimeout = parsed
		}
	}

	circuitBreakerConfig = circuitbreaker.DefaultConfig()
	if val := os.Getenv("STATEFABRIC_CB_FAILURE_THRESHOLD"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			circuitBreakerConfig.FailureThreshold = parsed
		}
	}
	if val := os.Getenv("STATEFABRIC_CB_COOLDOWN"); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil && parsed > 0 {
			circuitBreakerConfig.CooldownDuration = parsed
		}
	}
}

func clearString(s *string) {
	if s != nil && len(*s) > 0 {
		b := []byte(*s)
		for i := range b {
			b[i] = 0
		}
		*s = ""
	}
}

func parseBytes(s string) (uint64, error) {
	var val uint64
	_, err := fmt.Sscanf(s, "%d", &val)
	if err != nil {
		if strings.HasSuffix(s, "GB") || strings.HasSuffix(s, "gb") {
			var gb float64
			_, err := fmt.Sscanf(s, "%fGB", &gb)
			if err == nil {
				return uint64(gb * 1024 * 1024 * 1024), nil
			}
		}
		if strings.HasSuffix(s, "MB") || strings.HasSuffix(s, "mb") {
			var mb float64
			_, err := fmt.Sscanf(s, "%fMB", &mb)
			if err == nil {
				return uint64(mb * 1024 * 1024), nil
			}
		}
	}
	return val, err
}

func getmaxStoreSizeBytes() uint64 {
	return maxStoreSizeBytes
}

func getHTTPTimeout() time.Duration {
	return httpTimeout
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: getHTTPTimeout(),
		Transport: &http.Transport{
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func requestIDFromContext(ctx context.Context) string {
	if reqID, ok := ctx.Value("request_id").(string); ok {
		return reqID
	}
	if reqID, ok := ctx.Value(middleware.RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

type contextKey string

const RequestIDKey contextKey = "request_id"

func getLoggerWithRequestID(ctx context.Context) *logrus.Entry {
	fields := logrus.Fields{}
	if reqID := requestIDFromContext(ctx); reqID != "" {
		fields["request_id"] = reqID
	}
	return logrus.WithFields(fields)
}

type Repository struct {
	db        *gorm.DB
	stateRepo *statestore.StateRepository
	r2Backend *R2StorageBackend // Optional R2 backend for large data (events, snapshots, memory, replays)

	// Function execution client
	httpClient     *http.Client
	baseURL        string
	circuitBreaker *circuitbreaker.Breaker

	// API key security: store vault secret reference instead of actual key
	// The actual key is retrieved from vault when needed
	vaultRepo      *vault.Repository
	vaultSecretRef string
	apiKey         string // Fallback: cached only during initialization, cleared after

	// Redis cache for StateFabric data
	cache *StateFabricCache

	// Trigger engine for processing state change triggers
	triggerEngine *statestore.TriggerEngine

	// Active replay tracking for graceful shutdown
	mu                sync.Mutex
	replayCancelFuncs map[uuid.UUID]context.CancelFunc
}

func NewRepository(db *gorm.DB) *Repository {
	repo := &Repository{
		db:                db,
		stateRepo:         statestore.NewStateRepository(db),
		replayCancelFuncs: make(map[uuid.UUID]context.CancelFunc),
	}

	// Initialize R2 backend if configured
	if IsR2StorageConfigured() {
		if r2Backend, err := NewR2StorageBackend(); err == nil {
			repo.r2Backend = r2Backend
		}
	}

	// Initialize HTTP client for function execution with TLS verification
	repo.httpClient = newHTTPClient()

	// Initialize circuit breaker for pipeline execution
	repo.circuitBreaker = circuitbreaker.New(circuitBreakerConfig)

	return repo
}

// NewRepositoryWithRedis creates a repository with Redis caching
func NewRepositoryWithRedis(db *gorm.DB, redisClient *redis.Client) *Repository {
	repo := NewRepository(db)
	repo.cache = NewStateFabricCache(redisClient)
	return repo
}

// NewRepositoryWithVault creates a repository with vault integration for secret management
func NewRepositoryWithVault(db *gorm.DB, vaultRepo *vault.Repository) *Repository {
	repo := NewRepository(db)
	repo.vaultRepo = vaultRepo
	return repo
}

// NewRepositoryWithR2 creates a repository with an explicit R2 backend (for testing or custom config)
func NewRepositoryWithR2(db *gorm.DB, r2Backend *R2StorageBackend) *Repository {
	return &Repository{
		db:                db,
		stateRepo:         statestore.NewStateRepository(db),
		r2Backend:         r2Backend,
		replayCancelFuncs: make(map[uuid.UUID]context.CancelFunc),
		httpClient:        newHTTPClient(),
		circuitBreaker:    circuitbreaker.New(circuitBreakerConfig),
	}
}

// NewRepositoryWithVaultAndR2 creates a repository with vault and R2 backends
func NewRepositoryWithVaultAndR2(db *gorm.DB, vaultRepo *vault.Repository, r2Backend *R2StorageBackend) *Repository {
	repo := &Repository{
		db:                db,
		stateRepo:         statestore.NewStateRepository(db),
		r2Backend:         r2Backend,
		vaultRepo:         vaultRepo,
		replayCancelFuncs: make(map[uuid.UUID]context.CancelFunc),
		httpClient:        newHTTPClient(),
		circuitBreaker:    circuitbreaker.New(circuitBreakerConfig),
	}
	return repo
}

func (r *Repository) ConfigureExecution(baseURL, apiKey string) {
	r.baseURL = strings.TrimSuffix(baseURL, "/")
	if apiKey != "" {
		r.apiKey = apiKey
		clearString(&apiKey)
	}
}

// ConfigureExecutionWithVault sets execution config with vault secret reference
// The actual API key is retrieved from vault when needed, not stored in struct
func (r *Repository) ConfigureExecutionWithVault(ctx context.Context, baseURL, vaultSecretRef string) error {
	r.baseURL = strings.TrimSuffix(baseURL, "/")
	r.vaultSecretRef = vaultSecretRef

	if r.vaultRepo == nil {
		return fmt.Errorf("vault repository not configured")
	}

	secretID, err := uuid.Parse(vaultSecretRef)
	if err != nil {
		return fmt.Errorf("invalid vault secret ID: %w", err)
	}

	var systemTenantID uuid.UUID
	secret, err := r.vaultRepo.GetSecretByID(ctx, secretID, systemTenantID)
	if err != nil {
		return fmt.Errorf("failed to retrieve secret from vault: %w", err)
	}
	if secret == nil {
		return fmt.Errorf("vault secret not found: %s", vaultSecretRef)
	}

	return nil
}

// SetTriggerEngine configures the trigger engine for processing state change triggers
func (r *Repository) SetTriggerEngine(engine *statestore.TriggerEngine) {
	r.triggerEngine = engine
}

// getAPIKey retrieves the API key, preferring vault over cached value
func (r *Repository) getAPIKey(ctx context.Context) (string, error) {
	if r.vaultSecretRef != "" && r.vaultRepo != nil {
		secretID, err := uuid.Parse(r.vaultSecretRef)
		if err != nil {
			return "", fmt.Errorf("invalid vault secret ID: %w", err)
		}

		var systemTenantID uuid.UUID
		secret, err := r.vaultRepo.GetSecretByID(ctx, secretID, systemTenantID)
		if err != nil {
			return "", fmt.Errorf("failed to retrieve secret from vault: %w", err)
		}
		if secret == nil {
			return "", fmt.Errorf("vault secret not found: %s", r.vaultSecretRef)
		}

		// TODO: Implement proper server-side decryption for system secrets
		// For now, return the encrypted value (caller must decrypt)
		return string(secret.EncryptedValue), nil
	}

	if r.apiKey == "" {
		return "", fmt.Errorf("API key not configured")
	}

	return r.apiKey, nil
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
	if v < 0 {
		return fallback
	}
	return v
}

func (r *Repository) ListFabrics(ctx context.Context, opts ListOptions) ([]Fabric, int64, error) {
	// Try cache first (only for default query without filters)
	if opts.Status == "" && opts.Search == "" && r.cache != nil && r.cache.IsEnabled() {
		if cached, err := r.cache.GetFabricList(ctx, opts.TenantID); err == nil && cached != nil {
			r.cache.RecordCacheHit(opts.TenantID.String(), "", "fabric_list")
			total := int64(len(cached))
			offset := opts.Offset
			if offset > len(cached) {
				offset = len(cached)
			}
			end := offset + opts.Limit
			if end > len(cached) {
				end = len(cached)
			}
			if opts.Limit > 0 && end < len(cached) {
				return cached[offset:end], total, nil
			}
			if offset == 0 && (opts.Limit <= 0 || opts.Limit >= len(cached)) {
				return cached, total, nil
			}
		} else if cached == nil {
			r.cache.RecordCacheMiss(opts.TenantID.String(), "", "fabric_list")
		}
	}

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

	// Cache the result if it's a full list query
	if opts.Status == "" && opts.Search == "" && r.cache != nil && r.cache.IsEnabled() {
		r.cache.SetFabricList(ctx, opts.TenantID, items)
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
		return nil, fmt.Errorf(ErrFabricNotFound)
	}
	metrics, _ := r.GetMetrics(ctx, state.ID, "")
	pipelines, _ := r.ListPipelines(ctx, state.ID)
	fabric := buildFabric(state, metrics, pipelines)
	return &fabric, nil
}

var allowedSettingsKeys = map[string]bool{
	"region":         true,
	"replication":    true,
	"backupEnabled":  true,
	"backupInterval": true,
	"compression":    true,
	"encryption":     true,
	"ttl":            true,
	"maxConnections": true,
	"readOnly":       true,
	"autoScale":      true,
}

func (r *Repository) validateSettings(settings map[string]interface{}) error {
	if settings == nil {
		return nil
	}

	for key, value := range settings {
		if !allowedSettingsKeys[key] {
			return fmt.Errorf("unknown setting key: %s", key)
		}

		switch key {
		case "region":
			if region, ok := value.(string); ok {
				if strings.TrimSpace(region) == "" {
					return fmt.Errorf("region cannot be empty")
				}
				if len(region) > 64 {
					return fmt.Errorf("region name too long (max 64 chars)")
				}
			} else {
				return fmt.Errorf("region must be a string")
			}
		case "replication":
			if rep, ok := value.(float64); ok {
				if rep < 1 || rep > 10 {
					return fmt.Errorf("replication factor must be between 1 and 10")
				}
			} else {
				return fmt.Errorf("replication must be a number")
			}
		case "backupEnabled", "readOnly", "autoScale":
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("%s must be a boolean", key)
			}
		case "backupInterval":
			if interval, ok := value.(float64); ok {
				if interval < 1 || interval > 168 {
					return fmt.Errorf("backupInterval must be between 1 and 168 hours")
				}
			} else {
				return fmt.Errorf("backupInterval must be a number")
			}
		case "compression":
			if comp, ok := value.(string); ok {
				if comp != "none" && comp != "gzip" && comp != "lz4" && comp != "zstd" {
					return fmt.Errorf("compression must be one of: none, gzip, lz4, zstd")
				}
			} else {
				return fmt.Errorf("compression must be a string")
			}
		case "encryption":
			if enc, ok := value.(string); ok {
				if enc != "none" && enc != "aes-256" && enc != "chacha20" {
					return fmt.Errorf("encryption must be one of: none, aes-256, chacha20")
				}
			} else {
				return fmt.Errorf("encryption must be a string")
			}
		case "ttl":
			if ttl, ok := value.(float64); ok {
				if ttl < 0 || ttl > 31536000 {
					return fmt.Errorf("ttl must be between 0 and 31536000 seconds (1 year)")
				}
			} else {
				return fmt.Errorf("ttl must be a number")
			}
		case "maxConnections":
			if mc, ok := value.(float64); ok {
				if mc < 1 || mc > 10000 {
					return fmt.Errorf("maxConnections must be between 1 and 10000")
				}
			} else {
				return fmt.Errorf("maxConnections must be a number")
			}
		}
	}

	return nil
}

func (r *Repository) UpdateFabric(ctx context.Context, tenantID, fabricID uuid.UUID, updates map[string]interface{}) (*Fabric, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf(ErrFabricNotFound)
	}
	if name, ok := updates["name"].(string); ok && strings.TrimSpace(name) != "" {
		state.Name = name
		state.FullPath = fmt.Sprintf("%s/%s", tenantID.String()[:8], name)
	}
	if description, ok := updates["description"].(string); ok {
		state.Description = stringPtr(description)
	}
	if settings, ok := updates["settings"].(map[string]interface{}); ok {
		if err := r.validateSettings(settings); err != nil {
			return nil, fmt.Errorf("invalid settings: %w", err)
		}
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
	// Try cache first
	if r.cache != nil && r.cache.IsEnabled() {
		if cached, err := r.cache.GetMetrics(ctx, fabricID); err == nil && cached != nil {
			r.cache.RecordCacheHit("", fabricID.String(), "metrics")
			return *cached, nil
		}
		r.cache.RecordCacheMiss("", fabricID.String(), "metrics")
	}

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

	// Cache the result
	if r.cache != nil && r.cache.IsEnabled() {
		r.cache.SetMetrics(ctx, fabricID, &metrics)
	}

	return metrics, nil
}

func (r *Repository) ListStores(ctx context.Context, tenantID, fabricID uuid.UUID) ([]FabricStore, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf(ErrFabricNotFound)
	}
	return []FabricStore{buildStore(state)}, nil
}

func (r *Repository) CreateStore(ctx context.Context, tenantID, fabricID uuid.UUID, name, storeType string, maxSize int64, region string) (*FabricStore, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf(ErrFabricNotFound)
	}
	if strings.TrimSpace(name) != "" {
		state.Name = name
	}
	if maxSize < 0 {
		return nil, fmt.Errorf("maxSize cannot be negative")
	}
	if maxSize > 0 && maxSize > int64(maxStoreSizeBytes) {
		return nil, fmt.Errorf("maxSize exceeds maximum allowed size of %d bytes", maxStoreSizeBytes)
	}
	if maxSize > 0 {
		state.MaxSizeMB = int(maxSize / (1024 * 1024))
	}

	validStoreTypes := map[string]bool{
		"queue":      true,
		"persistent": true,
		"memory":     true,
	}
	if storeType != "" && !validStoreTypes[storeType] {
		return nil, fmt.Errorf("invalid store type: %s (allowed: queue, persistent, memory)", storeType)
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

	executionID := uuid.New()
	execution := &StateFabricPipelineExecution{
		ID:         executionID,
		PipelineID: pipelineID,
		Status:     "running",
		InputData:  input,
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

		if err := r.validateTargetFunctionForPipeline(ctx, tenantID, targetFunction); err != nil {
			execution.Status = "failed"
			execution.OutputData = map[string]interface{}{
				"error":     err.Error(),
				"stepIndex": i,
				"stepName":  stepName,
			}
			if saveErr := r.db.WithContext(ctx).Save(execution).Error; saveErr != nil {
				return nil, fmt.Errorf("failed to update execution record: %w", saveErr)
			}
			return nil, err
		}

		stepInput := input
		if i > 0 && lastOutput != nil {
			stepInput = lastOutput
		}

		var stepErr error
		var stepOutput map[string]interface{}
		retryConfig := DefaultRetryConfig()
		retryConfig.MaxAttempts = retryCount + 1

		for attempt := 0; attempt < retryConfig.MaxAttempts; attempt++ {
			if attempt > 0 {
				delay := time.Duration(0)
				if attempt == 1 {
					delay = retryConfig.InitialDelay
				} else {
					delay = time.Duration(float64(retryConfig.InitialDelay) * pow(retryConfig.BackoffMultiplier, float64(attempt-1)))
					if delay > retryConfig.MaxDelay {
						delay = retryConfig.MaxDelay
					}
				}
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(delay):
				}
			}

			stepOutput, stepErr = r.executeFunction(ctx, targetFunction, stepInput, timeoutMs)
			if stepErr == nil {
				break
			}

			if attempt == retryConfig.MaxAttempts-1 {
				r.recordDeadLetter(ctx, fabricID, &pipelineID, "pipeline_execution", input, stepErr, retryConfig.MaxAttempts)
			}
		}

		if stepErr != nil {
			execution.Status = "failed"
			execution.OutputData = map[string]interface{}{
				"error":     stepErr.Error(),
				"stepIndex": i,
				"stepName":  stepName,
				"stepType":  stepType,
				"attempts":  retryCount + 1,
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

	now := time.Now()
	trigger.LastTriggeredAt = &now
	if _, err := r.stateRepo.UpdateTrigger(ctx, trigger); err != nil {
		logrus.WithError(err).Warn("failed to update trigger last triggered time")
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

	if r.circuitBreaker == nil {
		return nil, fmt.Errorf("pipeline executor circuit breaker not initialized")
	}

	log := getLoggerWithRequestID(ctx)
	log.WithFields(logrus.Fields{
		"targetFunction": targetFunction,
		"timeoutMs":      timeoutMs,
	}).Debug("executing pipeline function via circuit breaker")

	var result map[string]interface{}
	err := r.circuitBreaker.Execute(func() error {
		execErr := r.doExecuteFunction(ctx, targetFunction, input, &result)
		return execErr
	})
	if err != nil {
		if errors.Is(err, circuitbreaker.ErrCircuitOpen) {
			log.WithFields(logrus.Fields{
				"targetFunction": targetFunction,
			}).Warn("pipeline execution circuit breaker open")
			return nil, fmt.Errorf("pipeline execution circuit breaker open: function %s temporarily unavailable", targetFunction)
		}
		log.WithError(err).WithFields(logrus.Fields{
			"targetFunction": targetFunction,
		}).Error("pipeline function execution failed")
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"targetFunction": targetFunction,
	}).Debug("pipeline function executed successfully")
	return result, nil
}

func (r *Repository) doExecuteFunction(ctx context.Context, targetFunction string, input map[string]interface{}, result *map[string]interface{}) error {
	log := getLoggerWithRequestID(ctx)

	var url string
	if strings.Contains(targetFunction, "/") {
		url = fmt.Sprintf("%s/v1/functions/by-name/%s/execute", r.baseURL, targetFunction)
	} else {
		url = fmt.Sprintf("%s/v1/functions/%s/execute", r.baseURL, targetFunction)
	}

	log.WithFields(logrus.Fields{
		"url":            url,
		"targetFunction": targetFunction,
	}).Debug("making HTTP request to execute function")

	jsonInput, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonInput))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	apiKey, err := r.getAPIKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}
	if apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"targetFunction": targetFunction,
		}).Error("HTTP request to execute function failed")
		return fmt.Errorf("failed to execute function: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		log.WithFields(logrus.Fields{
			"statusCode": resp.StatusCode,
			"body":       string(body),
		}).Error("function execution returned error status")
		return fmt.Errorf("function execution failed with status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, result); err != nil {
		log.WithError(err).Warn("failed to unmarshal function response, storing as raw")
		*result = map[string]interface{}{"raw": string(body)}
	}

	return nil
}

func (r *Repository) validateTargetFunctionForPipeline(ctx context.Context, tenantID uuid.UUID, targetFunction string) error {
	if targetFunction == "" {
		return fmt.Errorf("target function cannot be empty")
	}

	if strings.Contains(targetFunction, "..") || strings.Contains(targetFunction, "~") {
		return fmt.Errorf("invalid characters in target function name")
	}

	if strings.HasPrefix(targetFunction, "/") || strings.HasSuffix(targetFunction, "/") {
		return fmt.Errorf("target function name cannot start or end with slash")
	}

	var count int64
	query := `
		SELECT COUNT(*) FROM functions
		WHERE tenant_id = $1
		AND (
			-- Match by function name (author/name format)
			(CASE WHEN $2 LIKE '%/%' THEN $2 ELSE 'tenant/' || $2 END) = (owner || '/' || name)
			-- Or match by function ID directly if it's a UUID
			OR id = $3::uuid
		)
		AND status = 'deployed'
	`

	var functionID interface{}
	if id, err := uuid.Parse(targetFunction); err == nil {
		functionID = id.String()
	} else {
		functionID = uuid.Nil.String()
	}

	if err := r.db.WithContext(ctx).Raw(query, tenantID, targetFunction, functionID).Scan(&count).Error; err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"tenantID":       tenantID.String(),
			"targetFunction": targetFunction,
		}).Error("failed to validate target function for pipeline")
		return fmt.Errorf("failed to validate target function authorization")
	}

	if count == 0 {
		logrus.WithFields(logrus.Fields{
			"tenantID":       tenantID.String(),
			"targetFunction": targetFunction,
		}).Warn("pipeline execution blocked: target function not authorized for tenant")
		return fmt.Errorf("target function '%s' is not authorized for pipeline execution", targetFunction)
	}

	return nil
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
	const maxEventQueryLimit = 1000
	if limit > maxEventQueryLimit {
		limit = maxEventQueryLimit
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
		return nil, fmt.Errorf(ErrFabricNotFound)
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
		return nil, fmt.Errorf(ErrFabricNotFound)
	}

	// Create snapshot in PostgreSQL
	created, err := r.stateRepo.CreateSnapshot(ctx, fabricID, name)
	if err != nil {
		return nil, err
	}

	const maxSnapshotSizeBytes = 1 * 1024 * 1024 * 1024 // 1 GB limit
	stateDataSize := estimateJSONSize(created.StateData)
	if stateDataSize > maxSnapshotSizeBytes {
		return nil, fmt.Errorf("snapshot size %d exceeds maximum allowed size of %d bytes", stateDataSize, maxSnapshotSizeBytes)
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
				if err := r.db.WithContext(ctx).Model(&StateFabricSnapshot{}).Where("id = ?", created.ID).Updates(map[string]interface{}{
					"r2_object_key":   r2Object.Key,
					"r2_bucket":       r2Object.Bucket,
					"r2_content_hash": r2Object.ContentHash,
				}).Error; err != nil {
					logrus.WithError(err).Warn("failed to update snapshot R2 metadata")
				}
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
		return fmt.Errorf(ErrSnapshotNotFound)
	}
	return nil
}

func (r *Repository) ListReplays(ctx context.Context, tenantID, fabricID uuid.UUID) ([]ReplaySession, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf(ErrFabricNotFound)
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
		return nil, fmt.Errorf(ErrFabricNotFound)
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
		ID:             replayID,
		TenantID:       tenantID,
		FabricID:       fabricID,
		SnapshotID:     snapshotUUID,
		StartEventID:   startEventUUID,
		EndEventID:     endEventUUID,
		Status:         "running",
		Progress:       0,
		EventsReplayed: 0,
		StartedAt:      startedAt,
	}
	if err := r.db.WithContext(ctx).Create(dbReplay).Error; err != nil {
		return nil, fmt.Errorf("failed to create replay record: %w", err)
	}

	session := &ReplaySession{
		ID:             replayID.String(),
		FabricID:       fabricID.String(),
		SnapshotID:     req.SnapshotID,
		StartEventID:   req.StartEventID,
		EndEventID:     req.EndEventID,
		Status:         "running",
		Progress:       0,
		EventsReplayed: 0,
		StartedAt:      startedAt,
	}

	replayCtx, cancel := context.WithCancel(context.Background())
	r.mu.Lock()
	r.replayCancelFuncs[replayID] = cancel
	r.mu.Unlock()
	go r.executeReplay(replayCtx, replayID, fabricID, req)

	return session, nil
}

func (r *Repository) executeReplay(ctx context.Context, replayID, fabricID uuid.UUID, req ReplayCreateRequest) {
	defer func() {
		r.mu.Lock()
		if cancel, ok := r.replayCancelFuncs[replayID]; ok {
			cancel()
			delete(r.replayCancelFuncs, replayID)
		}
		r.mu.Unlock()
	}()

	logger := logrus.WithFields(logrus.Fields{"replay_id": replayID, "fabric_id": fabricID})

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
		select {
		case <-ctx.Done():
			r.updateReplayStatus(ctx, replayID, "cancelled", int(float64(i+1)/float64(totalEvents)*100), processed, "shutdown")
			logger.Info("replay cancelled due to shutdown")
			return
		default:
		}

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
			if err := r.db.WithContext(ctx).Model(&StateFabricReplay{}).Where("id = ?", replayID).Updates(map[string]interface{}{
				"r2_object_key":   obj.Key,
				"r2_bucket":       obj.Bucket,
				"r2_content_hash": obj.ContentHash,
			}).Error; err != nil {
				logrus.WithError(err).Warn("failed to update replay R2 metadata")
			}
		}
		eventsReplayed = int64(len(events))
	}

	r.updateReplayStatus(ctx, replayID, "completed", 100, eventsReplayed, "")
	logger.Infof("replay completed: %d events processed", eventsReplayed)
}

func (r *Repository) ShutdownReplays() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, cancel := range r.replayCancelFuncs {
		cancel()
	}
	r.replayCancelFuncs = make(map[uuid.UUID]context.CancelFunc)
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
	if err := r.db.WithContext(ctx).Model(&statestore.StateEvent{}).Select("COALESCE(MIN(sequence_num), 0)").Where("state_id = ?", fabricID).Scan(&firstSeq).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Model(&statestore.StateEvent{}).Select("COALESCE(MAX(sequence_num), 0)").Where("state_id = ?", fabricID).Scan(&lastSeq).Error; err != nil {
		return nil, err
	}

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
	if err := r.db.WithContext(ctx).Model(&StateFabricReplay{}).Where("id = ?", replayID).Updates(updates).Error; err != nil {
		logrus.WithError(err).Warn("failed to update replay status")
	}
}

func (r *Repository) updateReplayProgress(ctx context.Context, replayID uuid.UUID, progress int, eventsReplayed int64) {
	if err := r.db.WithContext(ctx).Model(&StateFabricReplay{}).Where("id = ?", replayID).Updates(map[string]interface{}{
		"progress":        progress,
		"events_replayed": eventsReplayed,
	}).Error; err != nil {
		logrus.WithError(err).Warn("failed to update replay progress")
	}
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
	if state == nil {
		return nil, fmt.Errorf("state not found")
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
	maxFabricsPerTenant := 10
	if val := os.Getenv("STATEFABRIC_MAX_FABRICS_PER_TENANT"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			maxFabricsPerTenant = parsed
		}
	}

	snapshotRetentionDays := 30
	if val := os.Getenv("STATEFABRIC_SNAPSHOT_RETENTION_DAYS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= 0 {
			snapshotRetentionDays = parsed
		}
	}

	allowPublicPipelines := false
	if val := os.Getenv("STATEFABRIC_ALLOW_PUBLIC_PIPELINES"); val != "" {
		allowPublicPipelines = val == "true" || val == "1"
	}

	maintenanceMode := false
	if val := os.Getenv("STATEFABRIC_MAINTENANCE_MODE"); val != "" {
		maintenanceMode = val == "true" || val == "1"
	}

	return map[string]interface{}{
		"maxFabricsPerTenant":          maxFabricsPerTenant,
		"defaultSnapshotRetentionDays": snapshotRetentionDays,
		"allowPublicPipelines":         allowPublicPipelines,
		"maintenanceMode":              maintenanceMode,
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
				"is_archived":   true,
				"archived_at":   now,
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
		// Events were stored in R2 but DB update failed - R2 data is now orphaned
		// Cleanup will happen via retention policy (~90 days) or manual intervention
		logrus.WithError(err).WithFields(logrus.Fields{
			"tenant_id":   tenantID,
			"fabric_id":   fabricID,
			"batch_id":    batchID,
			"event_count": len(events),
			"r2_key":      r2Object.Key,
			"r2_bucket":   r2Object.Bucket,
		}).Error("Failed to update event archival status in DB - R2 data is orphaned and will be cleaned up by retention policy")
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
	if r2Object == nil {
		return nil, fmt.Errorf("no data to store")
	}

	if err := r.db.WithContext(ctx).Model(&StateFabricReplay{}).Where("id = ?", replayID).Updates(map[string]interface{}{
		"r2_object_key":   r2Object.Key,
		"r2_bucket":       r2Object.Bucket,
		"r2_content_hash": r2Object.ContentHash,
	}).Error; err != nil {
		logrus.WithError(err).Warn("failed to update replay R2 metadata")
	}

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

// recordDeadLetter creates a dead letter entry for a failed operation
func (r *Repository) recordDeadLetter(ctx context.Context, fabricID uuid.UUID, pipelineID *uuid.UUID, operationType string, inputData map[string]interface{}, err error, attempts int) {
	deadLetter := &StateFabricDeadLetter{
		ID:            uuid.New(),
		FabricID:      fabricID,
		PipelineID:    pipelineID,
		OperationType: operationType,
		InputData:     inputData,
		ErrorMessage:  err.Error(),
		ErrorCode:     classifyError(err),
		Attempts:      attempts,
		MaxAttempts:   attempts,
		Status:        "pending",
	}
	if dlErr := r.db.WithContext(ctx).Create(deadLetter).Error; dlErr != nil {
		logrus.WithError(dlErr).WithFields(logrus.Fields{
			"fabric_id": fabricID,
			"operation": operationType,
		}).Error("failed to record dead letter entry")
	}
}

// classifyError classifies an error into a code for tracking
func classifyError(err error) string {
	if err == nil {
		return "unknown"
	}
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "timeout"):
		return "timeout"
	case strings.Contains(errStr, "circuit breaker"):
		return "circuit_breaker"
	case strings.Contains(errStr, "not found"):
		return "not_found"
	case strings.Contains(errStr, "unauthorized"):
		return "unauthorized"
	case strings.Contains(errStr, "rate limit"):
		return "rate_limited"
	default:
		return "execution_error"
	}
}

// ListDeadLetters returns dead letter entries for a fabric
func (r *Repository) ListDeadLetters(ctx context.Context, tenantID, fabricID uuid.UUID, limit, offset int) ([]StateFabricDeadLetter, int64, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, 0, err
	}
	if state.TenantID != tenantID {
		return nil, 0, fmt.Errorf("state fabric not found")
	}

	var total int64
	if err := r.db.WithContext(ctx).Model(&StateFabricDeadLetter{}).Where("fabric_id = ?", fabricID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var deadLetters []StateFabricDeadLetter
	if err := r.db.WithContext(ctx).Where("fabric_id = ?", fabricID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&deadLetters).Error; err != nil {
		return nil, 0, err
	}

	return deadLetters, total, nil
}

// GetDeadLetter returns a specific dead letter entry
func (r *Repository) GetDeadLetter(ctx context.Context, tenantID, fabricID, deadLetterID uuid.UUID) (*StateFabricDeadLetter, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf(ErrFabricNotFound)
	}

	var deadLetter StateFabricDeadLetter
	if err := r.db.WithContext(ctx).Where("id = ? AND fabric_id = ?", deadLetterID, fabricID).First(&deadLetter).Error; err != nil {
		return nil, err
	}

	return &deadLetter, nil
}

// RetryDeadLetter retries a dead letter operation
func (r *Repository) RetryDeadLetter(ctx context.Context, tenantID, fabricID, deadLetterID uuid.UUID) error {
	deadLetter, err := r.GetDeadLetter(ctx, tenantID, fabricID, deadLetterID)
	if err != nil {
		return err
	}

	deadLetter.Status = "retrying"
	deadLetter.Attempts++
	nextRetry := time.Now().Add(time.Duration(deadLetter.Attempts) * time.Minute)
	deadLetter.NextRetryAt = &nextRetry

	if err := r.db.WithContext(ctx).Save(deadLetter).Error; err != nil {
		return err
	}

	go r.processDeadLetterRetry(ctx, tenantID, fabricID, deadLetter.PipelineID, deadLetter.ID, deadLetter.OperationType, deadLetter.InputData)

	return nil
}

// processDeadLetterRetry processes a dead letter retry in the background
func (r *Repository) processDeadLetterRetry(ctx context.Context, tenantID, fabricID uuid.UUID, pipelineID *uuid.UUID, deadLetterID uuid.UUID, operationType string, inputData map[string]interface{}) {
	retryCtx := context.Background()

	var err error
	switch operationType {
	case "pipeline_execution":
		if pipelineID == nil {
			err = fmt.Errorf("pipeline ID is nil, cannot retry")
			break
		}
		_, err = r.ExecutePipeline(retryCtx, tenantID, fabricID, *pipelineID, inputData)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"pipeline_id": *pipelineID,
				"fabric_id":   fabricID,
			}).Warn("pipeline retry execution failed")
		}
	default:
		err = fmt.Errorf("unknown operation type: %s", operationType)
	}

	dl := &StateFabricDeadLetter{}
	if dbErr := r.db.WithContext(ctx).Where("id = ?", deadLetterID).First(dl).Error; dbErr != nil {
		logrus.WithError(dbErr).Error("failed to load dead letter for retry update")
		return
	}

	if err != nil {
		dl.Status = "failed"
		dl.ErrorMessage = err.Error()
	} else {
		dl.Status = "resolved"
		now := time.Now()
		dl.ResolvedAt = &now
	}
	dl.NextRetryAt = nil

	if updateErr := r.db.WithContext(ctx).Save(dl).Error; updateErr != nil {
		logrus.WithError(updateErr).WithField("dead_letter_id", deadLetterID).Error("failed to update dead letter after retry")
	}
}

// ResolveDeadLetter marks a dead letter as resolved
func (r *Repository) ResolveDeadLetter(ctx context.Context, tenantID, fabricID, deadLetterID uuid.UUID) error {
	deadLetter, err := r.GetDeadLetter(ctx, tenantID, fabricID, deadLetterID)
	if err != nil {
		return err
	}

	deadLetter.Status = "resolved"
	now := time.Now()
	deadLetter.ResolvedAt = &now
	deadLetter.NextRetryAt = nil

	return r.db.WithContext(ctx).Save(deadLetter).Error
}

// DeleteDeadLetter deletes a dead letter entry
func (r *Repository) DeleteDeadLetter(ctx context.Context, tenantID, fabricID, deadLetterID uuid.UUID) error {
	_, err := r.GetDeadLetter(ctx, tenantID, fabricID, deadLetterID)
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Delete(&StateFabricDeadLetter{}, "id = ?", deadLetterID).Error
}

// CleanupResolvedDeadLetters removes resolved dead letters older than the retention period
func (r *Repository) CleanupResolvedDeadLetters(ctx context.Context, retentionDays int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	result := r.db.WithContext(ctx).Where("status = ? AND resolved_at < ?", "resolved", cutoff).Delete(&StateFabricDeadLetter{})
	return result.RowsAffected, result.Error
}

// pow returns base raised to the power exp
func pow(base float64, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// Ping checks database connectivity
func (r *Repository) Ping(ctx context.Context) error {
	var result int
	return r.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
}

// IsCacheEnabled returns whether Redis cache is enabled
func (r *Repository) IsCacheEnabled() bool {
	return r.cache != nil && r.cache.IsEnabled()
}

// PingCache checks Redis connectivity
func (r *Repository) PingCache(ctx context.Context) error {
	if r.cache == nil || r.cache.redis == nil {
		return fmt.Errorf("cache not configured")
	}
	return r.cache.redis.Ping(ctx).Err()
}

// IsR2Enabled returns whether R2 backend is configured
func (r *Repository) IsR2Enabled() bool {
	return r.r2Backend != nil
}

// PingR2 checks R2 connectivity
func (r *Repository) PingR2(ctx context.Context) error {
	if r.r2Backend == nil {
		return fmt.Errorf("R2 backend not configured")
	}
	// R2 backend health check - try a simple operation
	return r.r2Backend.HealthCheck(ctx)
}

// CountDeadLetters returns the total count of pending dead letters across all fabrics
func (r *Repository) CountDeadLetters(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&StateFabricDeadLetter{}).Where("status IN ?", []string{"pending", "retrying"}).Count(&count).Error
	return count, err
}
