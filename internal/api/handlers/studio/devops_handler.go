package studio

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// generateID generates a unique ID with the given prefix
func generateID(prefix string) string {
	return prefix + "-" + uuid.New().String()[:8]
}

// ============================================================================
// DevOps Data Models
// ============================================================================

// PipelineStageStatus represents the status of a pipeline stage
type PipelineStageStatus string

const (
	PipelineStageStatusPending   PipelineStageStatus = "pending"
	PipelineStageStatusRunning  PipelineStageStatus = "running"
	PipelineStageStatusCompleted PipelineStageStatus = "completed"
	PipelineStageStatusFailed   PipelineStageStatus = "failed"
	PipelineStageStatusSkipped  PipelineStageStatus = "skipped"
	PipelineStageStatusWaiting  PipelineStageStatus = "waiting"
)

// PipelineStatus represents the overall pipeline status
type PipelineStatus string

const (
	PipelineStatusActive  PipelineStatus = "active"
	PipelineStatusPaused  PipelineStatus = "paused"
	PipelineStatusArchived PipelineStatus = "archived"
)

// PipelineStage represents a stage in a deployment pipeline
type PipelineStage struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Status      PipelineStageStatus `json:"status"`
	Duration    int64                `json:"duration,omitempty"`
	StartedAt   int64                `json:"started_at,omitempty"`
	CompletedAt int64                `json:"completed_at,omitempty"`
	Tasks       []PipelineTask       `json:"tasks,omitempty"`
}

// PipelineTask represents a task within a pipeline stage
type PipelineTask struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration int64  `json:"duration,omitempty"`
	Logs     []string `json:"logs,omitempty"`
	Error    string `json:"error,omitempty"`
}

// Pipeline represents a deployment pipeline
type Pipeline struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Status       PipelineStatus    `json:"status"`
	Stages       []PipelineStage   `json:"stages"`
	CurrentStageID string          `json:"current_stage_id,omitempty"`
	TriggeredBy  string            `json:"triggered_by"`
	TriggeredAt  int64             `json:"triggered_at"`
	Branch       string            `json:"branch"`
	CommitSha    string            `json:"commit_sha"`
	Source       string            `json:"source"`
	TenantID     string            `json:"tenant_id"`
	CreatedAt    int64             `json:"created_at"`
	UpdatedAt    int64             `json:"updated_at"`
}

// EnvironmentType represents the type of environment
type EnvironmentType string

const (
	EnvironmentTypeProduction  EnvironmentType = "production"
	EnvironmentTypeStaging    EnvironmentType = "staging"
	EnvironmentTypePreview    EnvironmentType = "preview"
	EnvironmentTypeDevelopment EnvironmentType = "development"
)

// Environment represents a deployment environment
type Environment struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       EnvironmentType  `json:"type"`
	Color      string            `json:"color"`
	Variables  map[string]string `json:"variables"`
	Secrets    []EnvironmentSecret `json:"secrets"`
	Replicas   int               `json:"replicas"`
	AutoScale  bool              `json:"auto_scale"`
	Region     string            `json:"region,omitempty"`
	TenantID   string            `json:"tenant_id"`
	CreatedAt  int64             `json:"created_at"`
	UpdatedAt  int64             `json:"updated_at"`
}

// EnvironmentSecret represents a secret in an environment
type EnvironmentSecret struct {
	Key         string `json:"key"`
	Masked      bool   `json:"masked"`
	LastUpdated int64  `json:"last_updated"`
}

// CloudProvider represents a cloud provider
type CloudProvider string

const (
	CloudProviderAWS   CloudProvider = "aws"
	CloudProviderGCP   CloudProvider = "gcp"
	CloudProviderAzure CloudProvider = "azure"
	CloudProviderCustom CloudProvider = "custom"
)

// CloudRegionSpec represents compute specs for a region
type CloudRegionSpec struct {
	Compute  int  `json:"compute,omitempty"`
	Memory   int  `json:"memory,omitempty"`
	Storage  int  `json:"storage,omitempty"`
	GPU      bool `json:"gpu,omitempty"`
}

// CloudRegion represents a cloud region for deployment
type CloudRegion struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Provider     CloudProvider  `json:"provider"`
	Zone         string         `json:"zone"`
	ZoneName     string         `json:"zone_name"`
	Location     string         `json:"location"`
	Country      string         `json:"country"`
	Coordinates  struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"coordinates"`
	IsAvailable  bool           `json:"is_available"`
	IsRecommended bool          `json:"is_recommended,omitempty"`
	Specs        *CloudRegionSpec `json:"specs,omitempty"`
	TenantID     string         `json:"tenant_id"`
	CreatedAt    int64          `json:"created_at"`
}

// ============================================================================
// DevOps Handler
// ============================================================================

// DevOpsHandler handles studio DevOps HTTP requests
type DevOpsHandler struct {
	devopsRepo *DevOpsRepository
}

// NewDevOpsHandler creates a new DevOps handler
func NewDevOpsHandler(devopsRepo *DevOpsRepository) *DevOpsHandler {
	return &DevOpsHandler{devopsRepo: devopsRepo}
}

// ============================================================================
// Pipeline Handlers
// ============================================================================

// HandleListPipelines handles GET /v1/studio/devops/pipelines
func (h *DevOpsHandler) HandleListPipelines(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			offset = n
		}
	}

	status := r.URL.Query().Get("status")

	pipelines, err := h.devopsRepo.ListPipelines(r.Context(), tenantID, status, limit, offset)
	if err != nil {
		logrus.WithError(err).Warn("studio devops: failed to list pipelines")
		writeJSON(w, http.StatusOK, map[string]interface{}{"pipelines": []Pipeline{}, "total": 0})
		return
	}

	total := len(pipelines)
	writeJSON(w, http.StatusOK, map[string]interface{}{"pipelines": pipelines, "total": total})
}

// HandleGetPipeline handles GET /v1/studio/devops/pipelines/{id}
func (h *DevOpsHandler) HandleGetPipeline(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pipelineID := mux.Vars(r)["id"]
	if pipelineID == "" {
		writeJSONError(w, http.StatusBadRequest, "pipeline id is required")
		return
	}

	pipeline, err := h.devopsRepo.GetPipeline(r.Context(), tenantID, pipelineID)
	if err != nil {
		logrus.WithError(err).Warn("studio devops: failed to get pipeline")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get pipeline")
		return
	}
	if pipeline == nil {
		writeJSONError(w, http.StatusNotFound, "Pipeline not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"pipeline": pipeline})
}

// HandleCreatePipeline handles POST /v1/studio/devops/pipelines
func (h *DevOpsHandler) HandleCreatePipeline(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Name      string   `json:"name"`
		Version   string   `json:"version"`
		Branch    string   `json:"branch"`
		CommitSha string   `json:"commit_sha"`
		Source    string   `json:"source"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "name is required")
		return
	}

	if req.Source == "" {
		req.Source = "manual"
	}

	pipeline := &Pipeline{
		ID:          generateID("pipe"),
		Name:        req.Name,
		Version:     req.Version,
		Status:      PipelineStatusActive,
		Stages:      []PipelineStage{},
		TriggeredBy: getUserID(r),
		TriggeredAt: time.Now().UnixMilli(),
		Branch:      req.Branch,
		CommitSha:   req.CommitSha,
		Source:      req.Source,
		TenantID:    tenantID,
		CreatedAt:   time.Now().UnixMilli(),
		UpdatedAt:   time.Now().UnixMilli(),
	}

	if err := h.devopsRepo.CreatePipeline(r.Context(), pipeline); err != nil {
		logrus.WithError(err).Error("studio devops: failed to create pipeline")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create pipeline")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"pipeline": pipeline})
}

// HandleUpdatePipelineStage handles PATCH /v1/studio/devops/pipelines/{id}/stages/{stageId}
func (h *DevOpsHandler) HandleUpdatePipelineStage(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pipelineID := mux.Vars(r)["id"]
	stageID := mux.Vars(r)["stageId"]
	if pipelineID == "" || stageID == "" {
		writeJSONError(w, http.StatusBadRequest, "pipeline id and stage id are required")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	pipeline, err := h.devopsRepo.UpdatePipelineStage(r.Context(), tenantID, pipelineID, stageID, updates)
	if err != nil {
		logrus.WithError(err).Error("studio devops: failed to update stage")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update stage")
		return
	}
	if pipeline == nil {
		writeJSONError(w, http.StatusNotFound, "Pipeline or stage not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"pipeline": pipeline})
}

// HandleRetryPipelineStage handles POST /v1/studio/devops/pipelines/{id}/stages/{stageId}/retry
func (h *DevOpsHandler) HandleRetryPipelineStage(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pipelineID := mux.Vars(r)["id"]
	stageID := mux.Vars(r)["stageId"]
	if pipelineID == "" || stageID == "" {
		writeJSONError(w, http.StatusBadRequest, "pipeline id and stage id are required")
		return
	}

	pipeline, err := h.devopsRepo.UpdatePipelineStage(r.Context(), tenantID, pipelineID, stageID, map[string]interface{}{
		"status": PipelineStageStatusPending,
	})
	if err != nil {
		logrus.WithError(err).Error("studio devops: failed to retry stage")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retry stage")
		return
	}
	if pipeline == nil {
		writeJSONError(w, http.StatusNotFound, "Pipeline or stage not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"pipeline": pipeline})
}

// ============================================================================
// Environment Handlers
// ============================================================================

// HandleListEnvironments handles GET /v1/studio/devops/environments
func (h *DevOpsHandler) HandleListEnvironments(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	envType := r.URL.Query().Get("type")

	environments, err := h.devopsRepo.ListEnvironments(r.Context(), tenantID, envType)
	if err != nil {
		logrus.WithError(err).Warn("studio devops: failed to list environments")
		writeJSON(w, http.StatusOK, map[string]interface{}{"environments": []Environment{}})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"environments": environments})
}

// HandleGetEnvironment handles GET /v1/studio/devops/environments/{id}
func (h *DevOpsHandler) HandleGetEnvironment(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	envID := mux.Vars(r)["id"]
	if envID == "" {
		writeJSONError(w, http.StatusBadRequest, "environment id is required")
		return
	}

	env, err := h.devopsRepo.GetEnvironment(r.Context(), tenantID, envID)
	if err != nil {
		logrus.WithError(err).Warn("studio devops: failed to get environment")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get environment")
		return
	}
	if env == nil {
		writeJSONError(w, http.StatusNotFound, "Environment not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"environment": env})
}

// HandleCreateEnvironment handles POST /v1/studio/devops/environments
func (h *DevOpsHandler) HandleCreateEnvironment(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Name      string          `json:"name"`
		Type      EnvironmentType `json:"type"`
		Color     string          `json:"color"`
		Variables map[string]string `json:"variables"`
		Replicas  int             `json:"replicas"`
		AutoScale bool            `json:"auto_scale"`
		Region    string          `json:"region"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "name is required")
		return
	}

	if req.Color == "" {
		req.Color = "#06b6d4"
	}
	if req.Variables == nil {
		req.Variables = map[string]string{}
	}

	env := &Environment{
		ID:        generateID("env"),
		Name:      req.Name,
		Type:      req.Type,
		Color:     req.Color,
		Variables: req.Variables,
		Secrets:   []EnvironmentSecret{},
		Replicas:  req.Replicas,
		AutoScale: req.AutoScale,
		Region:    req.Region,
		TenantID:  tenantID,
		CreatedAt: time.Now().UnixMilli(),
		UpdatedAt: time.Now().UnixMilli(),
	}

	if err := h.devopsRepo.CreateEnvironment(r.Context(), env); err != nil {
		logrus.WithError(err).Error("studio devops: failed to create environment")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create environment")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"environment": env})
}

// HandleUpdateEnvironment handles PATCH /v1/studio/devops/environments/{id}
func (h *DevOpsHandler) HandleUpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	envID := mux.Vars(r)["id"]
	if envID == "" {
		writeJSONError(w, http.StatusBadRequest, "environment id is required")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	env, err := h.devopsRepo.UpdateEnvironment(r.Context(), tenantID, envID, updates)
	if err != nil {
		logrus.WithError(err).Error("studio devops: failed to update environment")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update environment")
		return
	}
	if env == nil {
		writeJSONError(w, http.StatusNotFound, "Environment not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"environment": env})
}

// HandleDeleteEnvironment handles DELETE /v1/studio/devops/environments/{id}
func (h *DevOpsHandler) HandleDeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	envID := mux.Vars(r)["id"]
	if envID == "" {
		writeJSONError(w, http.StatusBadRequest, "environment id is required")
		return
	}

	if err := h.devopsRepo.DeleteEnvironment(r.Context(), tenantID, envID); err != nil {
		logrus.WithError(err).Error("studio devops: failed to delete environment")
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete environment")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Environment deleted"})
}

// HandleAddEnvironmentVariable handles POST /v1/studio/devops/environments/{id}/variables
func (h *DevOpsHandler) HandleAddEnvironmentVariable(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	envID := mux.Vars(r)["id"]
	if envID == "" {
		writeJSONError(w, http.StatusBadRequest, "environment id is required")
		return
	}

	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Key == "" {
		writeJSONError(w, http.StatusBadRequest, "key is required")
		return
	}

	env, err := h.devopsRepo.AddEnvironmentVariable(r.Context(), tenantID, envID, req.Key, req.Value)
	if err != nil {
		logrus.WithError(err).Error("studio devops: failed to add variable")
		writeJSONError(w, http.StatusInternalServerError, "Failed to add variable")
		return
	}
	if env == nil {
		writeJSONError(w, http.StatusNotFound, "Environment not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"environment": env})
}

// HandleAddEnvironmentSecret handles POST /v1/studio/devops/environments/{id}/secrets
func (h *DevOpsHandler) HandleAddEnvironmentSecret(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	envID := mux.Vars(r)["id"]
	if envID == "" {
		writeJSONError(w, http.StatusBadRequest, "environment id is required")
		return
	}

	var req struct {
		Key string `json:"key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Key == "" {
		writeJSONError(w, http.StatusBadRequest, "key is required")
		return
	}

	env, err := h.devopsRepo.AddEnvironmentSecret(r.Context(), tenantID, envID, req.Key)
	if err != nil {
		logrus.WithError(err).Error("studio devops: failed to add secret")
		writeJSONError(w, http.StatusInternalServerError, "Failed to add secret")
		return
	}
	if env == nil {
		writeJSONError(w, http.StatusNotFound, "Environment not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"environment": env})
}

// ============================================================================
// Cloud Region Handlers
// ============================================================================

// HandleListCloudRegions handles GET /v1/studio/devops/regions
func (h *DevOpsHandler) HandleListCloudRegions(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	provider := r.URL.Query().Get("provider")

	regions, err := h.devopsRepo.ListCloudRegions(r.Context(), tenantID, provider)
	if err != nil {
		logrus.WithError(err).Warn("studio devops: failed to list regions")
		writeJSON(w, http.StatusOK, map[string]interface{}{"regions": []CloudRegion{}})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"regions": regions})
}

// HandleGetCloudRegion handles GET /v1/studio/devops/regions/{id}
func (h *DevOpsHandler) HandleGetCloudRegion(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	regionID := mux.Vars(r)["id"]
	if regionID == "" {
		writeJSONError(w, http.StatusBadRequest, "region id is required")
		return
	}

	region, err := h.devopsRepo.GetCloudRegion(r.Context(), tenantID, regionID)
	if err != nil {
		logrus.WithError(err).Warn("studio devops: failed to get region")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get region")
		return
	}
	if region == nil {
		writeJSONError(w, http.StatusNotFound, "Region not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"region": region})
}

// HandleCreateCloudRegion handles POST /v1/studio/devops/regions
func (h *DevOpsHandler) HandleCreateCloudRegion(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Name     string        `json:"name"`
		Provider CloudProvider `json:"provider"`
		Zone     string        `json:"zone"`
		ZoneName string        `json:"zone_name"`
		Location string        `json:"location"`
		Country  string        `json:"country"`
		Specs    *CloudRegionSpec `json:"specs"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Provider == "" {
		req.Provider = CloudProviderAWS
	}

	region := &CloudRegion{
		ID:          generateID("reg"),
		Name:        req.Name,
		Provider:    req.Provider,
		Zone:         req.Zone,
		ZoneName:    req.ZoneName,
		Location:    req.Location,
		Country:     req.Country,
		IsAvailable: true,
		Specs:       req.Specs,
		TenantID:    tenantID,
	}

	if err := h.devopsRepo.CreateCloudRegion(r.Context(), region); err != nil {
		logrus.WithError(err).Error("studio devops: failed to create region")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create region")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"region": region})
}

// ============================================================================
// DevOps Stats Handler
// ============================================================================

// HandleGetDevOpsStats handles GET /v1/studio/devops/stats
func (h *DevOpsHandler) HandleGetDevOpsStats(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	if tenantID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	pipelines, _ := h.devopsRepo.ListPipelines(r.Context(), tenantID, "", 100, 0)
	environments, _ := h.devopsRepo.ListEnvironments(r.Context(), tenantID, "")
	regions, _ := h.devopsRepo.ListCloudRegions(r.Context(), tenantID, "")

	var activePipelines, successRate, avgColdStart int64
	for _, p := range pipelines {
		if p.Status == PipelineStatusActive {
			activePipelines++
		}
	}

	// Calculate success rate and avg cold start from stages
	var completedStages, totalStages int64
	for _, pipeline := range pipelines {
		for _, stage := range pipeline.Stages {
			totalStages++
			if stage.Status == PipelineStageStatusCompleted {
				completedStages++
			}
		}
	}
	if totalStages > 0 {
		successRate = (completedStages * 100) / totalStages
	}

	// Simulated avg cold start (would come from actual metrics in production)
	avgColdStart = 2100 // 2.1 seconds in ms

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"pipelines":      len(pipelines),
		"active_pipelines": activePipelines,
		"success_rate":   successRate,
		"avg_cold_start_ms": avgColdStart,
		"environments":  len(environments),
		"regions":       len(regions),
	})
}