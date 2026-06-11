// oversight.go - Admin oversight handlers for trust, execution audit, fraud detection, and economic monitoring
package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// TrustDashboardData represents the trust distribution and monitoring data
type TrustDashboardData struct {
	Distribution            TrustDistribution     `json:"distribution"`
	HighRiskFunctions       []HighRiskFunction    `json:"highRiskFunctions"`
	TrustSpikes             []TrustSpike          `json:"trustSpikes"`
	ReputationFarmingAlerts []ReputationFarmAlert `json:"reputationFarmingAlerts"`
}

type TrustDistribution struct {
	Excellent int `json:"excellent"`
	Good      int `json:"good"`
	Fair      int `json:"fair"`
	Poor      int `json:"poor"`
}

type HighRiskFunction struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Tenant      string   `json:"tenant"`
	TrustScore  float64  `json:"trustScore"`
	RiskFactors []string `json:"riskFactors"`
	LastUpdated string   `json:"lastUpdated"`
}

type TrustSpike struct {
	ID            string  `json:"id"`
	FunctionName  string  `json:"functionName"`
	Tenant        string  `json:"tenant"`
	PreviousScore float64 `json:"previousScore"`
	NewScore      float64 `json:"newScore"`
	SpikeAmount   float64 `json:"spikeAmount"`
	DetectedAt    string  `json:"detectedAt"`
}

type ReputationFarmAlert struct {
	ID                string                 `json:"id"`
	Type              string                 `json:"type"`
	Description       string                 `json:"description"`
	AffectedFunctions []string               `json:"affectedFunctions"`
	Severity          string                 `json:"severity"`
	DetectedAt        string                 `json:"detectedAt"`
	Details           map[string]interface{} `json:"details"`
}

// ExecutionAuditData represents execution audit trail
type ExecutionAuditData struct {
	Executions []ExecutionRecord `json:"executions"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
}

type ExecutionRecord struct {
	ID                string  `json:"id"`
	ExecutionRootHash string  `json:"executionRootHash"`
	Tenant            string  `json:"tenant"`
	FunctionName      string  `json:"functionName"`
	Timestamp         string  `json:"timestamp"`
	NodeSignature     string  `json:"nodeSignature"`
	Status            string  `json:"status"`
	Duration          int     `json:"duration"`
	InputSize         int     `json:"inputSize"`
	OutputSize        int     `json:"outputSize"`
	ErrorMessage      *string `json:"errorMessage,omitempty"`
}

// FraudDetectionData represents fraud detection results
type FraudDetectionData struct {
	BotPatterns         []BotPattern         `json:"botPatterns"`
	FakeDiversityAlerts []FakeDiversityAlert `json:"fakeDiversityAlerts"`
	IPClusters          []IPCluster          `json:"ipClusters"`
	WashUsagePatterns   []WashUsagePattern   `json:"washUsagePatterns"`
	Summary             FraudSummary         `json:"summary"`
}

type BotPattern struct {
	ID                string   `json:"id"`
	PatternType       string   `json:"patternType"`
	ConfidenceScore   int      `json:"confidenceScore"`
	AffectedFunctions []string `json:"affectedFunctions"`
	AffectedTenants   []string `json:"affectedTenants"`
	DetectedAt        string   `json:"detectedAt"`
	Pattern           string   `json:"pattern"`
}

type FakeDiversityAlert struct {
	ID             string   `json:"id"`
	TenantGroup    []string `json:"tenantGroup"`
	Indicators     []string `json:"indicators"`
	RiskLevel      string   `json:"riskLevel"`
	DetectedAt     string   `json:"detectedAt"`
	CommonPatterns []string `json:"commonPatterns"`
}

type IPCluster struct {
	ID                string   `json:"id"`
	IPRange           string   `json:"ipRange"`
	AssociatedTenants []string `json:"associatedTenants"`
	RiskLevel         string   `json:"riskLevel"`
	CommonPatterns    []string `json:"commonPatterns"`
	FirstSeen         string   `json:"firstSeen"`
	LastSeen          string   `json:"lastSeen"`
}

type WashUsagePattern struct {
	ID                   string `json:"id"`
	TenantA              string `json:"tenantA"`
	TenantB              string `json:"tenantB"`
	Function             string `json:"function"`
	Pattern              string `json:"pattern"`
	Confidence           int    `json:"confidence"`
	ReciprocalExecutions int    `json:"reciprocalExecutions"`
	DetectedAt           string `json:"detectedAt"`
}

type FraudSummary struct {
	TotalBotPatterns  int `json:"totalBotPatterns"`
	HighRiskClusters  int `json:"highRiskClusters"`
	SuspiciousTenants int `json:"suspiciousTenants"`
	WashUsageDetected int `json:"washUsageDetected"`
}

// EconomicLeaderboardData represents economic monitoring data
type EconomicLeaderboardData struct {
	TopRevenueGenerators []RevenueGenerator      `json:"topRevenueGenerators"`
	SuspiciousGrowth     []SuspiciousGrowthAlert `json:"suspiciousGrowth"`
	ArtificialBoosting   []ArtificialBoosting    `json:"artificialBoosting"`
}

type RevenueGenerator struct {
	ID             string  `json:"id"`
	Rank           int     `json:"rank"`
	TenantFunction string  `json:"tenantFunction"`
	Revenue30d     float64 `json:"revenue30d"`
	ExecutionCount int     `json:"executionCount"`
	GrowthRate     float64 `json:"growthRate"`
}

type SuspiciousGrowthAlert struct {
	ID             string `json:"id"`
	TenantFunction string `json:"tenantFunction"`
	Pattern        string `json:"pattern"`
	Details        string `json:"details"`
	DetectedAt     string `json:"detectedAt"`
}

type ArtificialBoosting struct {
	ID              string   `json:"id"`
	Function        string   `json:"function"`
	DetectedPattern string   `json:"detectedPattern"`
	Confidence      int      `json:"confidence"`
	RelatedAccounts []string `json:"relatedAccounts"`
}

// OversightHandler handles admin oversight operations
type OversightHandler struct {
	registryRepo *registry.RegistryRepository
	db           *gorm.DB
	logger       *logrus.Logger
}

// NewOversightHandler creates a new oversight handler
func NewOversightHandler(registryRepo *registry.RegistryRepository, db *gorm.DB, logger *logrus.Logger) *OversightHandler {
	if logger == nil {
		logger = logrus.New()
	}
	return &OversightHandler{
		registryRepo: registryRepo,
		db:           db,
		logger:       logger,
	}
}

// buildRiskFactors derives risk factor labels from high-risk function row metrics.
func buildRiskFactors(row registry.HighRiskFunctionRow) []string {
	var factors []string
	if row.ErrorRate > 0.05 {
		factors = append(factors, "Elevated error rate")
	}
	if row.TimeoutRate > 0.05 {
		factors = append(factors, "Elevated timeout rate")
	}
	if row.TenantDiversity < 3 && (row.ErrorRate > 0 || row.TimeoutRate > 0) {
		factors = append(factors, "Low tenant diversity")
	}
	if row.TrustScore > 0 && row.TrustScore < 0.2 {
		factors = append(factors, "Very low trust score")
	}
	if len(factors) == 0 {
		factors = append(factors, "Low trust score")
	}
	return factors
}

// blockUserOrTenant blocks a user or tenant by setting their execution quota BlockUntil
func (h *OversightHandler) blockUserOrTenant(entityID uuid.UUID, entityType string) error {
	var userID, tenantID *uuid.UUID

	if entityType == "user" {
		userID = &entityID
	} else {
		tenantID = &entityID
	}

	// Find the user's execution quota
	var quota storage.UserExecutionQuota
	query := h.db.Model(&storage.UserExecutionQuota{})
	if userID != nil {
		query = query.Where("user_id = ?", userID)
	}
	if tenantID != nil {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.First(&quota).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new quota record for blocking
			quota = storage.UserExecutionQuota{
				UserID:   userID,
				TenantID: tenantID,
			}
		} else {
			return fmt.Errorf("failed to find user execution quota: %w", err)
		}
	}

	// Set block until time (30 days by default)
	blockUntil := time.Now().Add(30 * 24 * time.Hour)
	quota.BlockUntil = &blockUntil
	quota.BlockReason = "Blocked by admin oversight"
	quota.SuspiciousActivityScore = 1000 // High score to ensure blocking

	if quota.ID == uuid.Nil {
		// New record
		quota.ID = uuid.New()
		quota.CreatedAt = time.Now()
		if err := h.db.Create(&quota).Error; err != nil {
			return fmt.Errorf("failed to create blocked user quota: %w", err)
		}
	} else {
		// Update existing
		quota.UpdatedAt = time.Now()
		if err := h.db.Save(&quota).Error; err != nil {
			return fmt.Errorf("failed to update blocked user quota: %w", err)
		}
	}

	// Log the blocking action
	abusePattern := &storage.AbusePattern{
		PatternType: "admin_user_block",
		Severity:    "critical",
		UserID:      userID,
		Description: fmt.Sprintf("User/tenant blocked by admin: %s", entityType),
		ActionTaken: "blocked",
		DetectedAt:  time.Now(),
	}

	if err := h.db.Create(abusePattern).Error; err != nil {
		h.logger.WithError(err).Warn("Failed to log user blocking action")
	}

	return nil
}

// blockIP blocks an IP address by creating an abuse pattern record
func (h *OversightHandler) blockIP(ipAddress string) error {
	// Create an abuse pattern record to track the IP block
	abusePattern := &storage.AbusePattern{
		PatternType: "admin_ip_block",
		Severity:    "critical",
		IPAddress:   ipAddress,
		Description: "IP blocked by admin oversight",
		ActionTaken: "blocked",
		DetectedAt:  time.Now(),
	}

	if err := h.db.Create(abusePattern).Error; err != nil {
		return fmt.Errorf("failed to create IP block record: %w", err)
	}

	return nil
}

// blockPattern blocks a specific fraud pattern instance
func (h *OversightHandler) blockPattern(patternType, patternID string, adminUserID uuid.UUID) error {
	// Create a blocked pattern record
	blockedPattern := &storage.BlockedPattern{
		PatternType: patternType,
		PatternID:   patternID,
		Description: fmt.Sprintf("Pattern %s:%s blocked by admin oversight", patternType, patternID),
		BlockedBy:   adminUserID,
		BlockedAt:   time.Now(),
	}

	if err := h.db.Create(blockedPattern).Error; err != nil {
		return fmt.Errorf("failed to create blocked pattern record: %w", err)
	}

	return nil
}

// openInvestigation creates an investigation record for suspicious activity
func (h *OversightHandler) openInvestigation(entityType, entityID, reason string, adminUserID uuid.UUID) error {
	// Determine priority based on entity type and reason
	priority := "medium"
	if strings.Contains(reason, "critical") || strings.Contains(reason, "high") {
		priority = "high"
	}
	if strings.Contains(reason, "suspicious") || entityType == "function" {
		priority = "medium"
	}

	investigation := &storage.Investigation{
		EntityType: entityType,
		EntityID:   entityID,
		Reason:     reason,
		OpenedBy:   adminUserID,
		Status:     "open",
		Priority:   priority,
		OpenedAt:   time.Now(),
	}

	if err := h.db.Create(investigation).Error; err != nil {
		return fmt.Errorf("failed to create investigation record: %w", err)
	}

	return nil
}

// HandleGetTrustDashboard returns trust distribution and monitoring data
func (h *OversightHandler) HandleGetTrustDashboard(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": claims.UserID,
	}).Info("Fetching trust dashboard data")

	dist, err := h.registryRepo.GetTrustDistribution()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get trust distribution")
		apierror.WriteError(w, apierror.NewInternal("Failed to load trust distribution"))
		return
	}

	highRiskRows, err := h.registryRepo.GetHighRiskFunctions(20)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get high-risk functions")
		apierror.WriteError(w, apierror.NewInternal("Failed to load high-risk functions"))
		return
	}

	highRisk := make([]HighRiskFunction, 0, len(highRiskRows))
	for _, row := range highRiskRows {
		riskFactors := buildRiskFactors(row)
		lastUpdated := ""
		if row.TrustUpdatedAt != nil {
			lastUpdated = row.TrustUpdatedAt.Format(time.RFC3339)
		}
		// API exposes trust score on 0-100 scale; DB stores 0-1
		trustScore100 := row.TrustScore * 100
		highRisk = append(highRisk, HighRiskFunction{
			ID:          row.FunctionID.String(),
			Name:        row.Name,
			Tenant:      row.Author,
			TrustScore:  trustScore100,
			RiskFactors: riskFactors,
			LastUpdated: lastUpdated,
		})
	}

	// Trust spikes and reputation farming alerts require historical/alert data not yet in schema; return empty.
	data := TrustDashboardData{
		Distribution: TrustDistribution{
			Excellent: dist.Excellent,
			Good:      dist.Good,
			Fair:      dist.Fair,
			Poor:      dist.Poor,
		},
		HighRiskFunctions:       highRisk,
		TrustSpikes:             []TrustSpike{},
		ReputationFarmingAlerts: []ReputationFarmAlert{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// HandleGetExecutionAudit returns execution audit trail
func (h *OversightHandler) HandleGetExecutionAudit(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	searchTerm := r.URL.Query().Get("search")
	tenantFilter := r.URL.Query().Get("tenant")
	statusFilter := r.URL.Query().Get("status")
	_ = r.URL.Query().Get("hash") // hash filter reserved for future use

	h.logger.WithFields(logrus.Fields{
		"user_id":       claims.UserID,
		"page":          page,
		"pageSize":      pageSize,
		"search":        searchTerm,
		"tenant_filter": tenantFilter,
		"status_filter": statusFilter,
	}).Info("Fetching execution audit data")

	offset := (page - 1) * pageSize
	rows, total, err := h.registryRepo.GetExecutionAuditData(searchTerm, tenantFilter, statusFilter, offset, pageSize)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get execution audit data")
		apierror.WriteError(w, apierror.NewInternal("Failed to load execution audit data"))
		return
	}

	executions := make([]ExecutionRecord, 0, len(rows))
	for _, row := range rows {
		// Derive execution root hash from function ID + timestamp if no certificate exists
		rootHash := row.ExecutionRootHash.String
		if rootHash == "" {
			rootHash = fmt.Sprintf("uncertified-%s-%d", row.FunctionID.String()[:8], row.Timestamp.Unix())
		}

		// Use node signature from certificate, or indicate uncertified if none
		nodeSig := row.NodeSignature.String
		if nodeSig == "" {
			nodeSig = "uncertified"
		}

		// Map error message from error_code
		var errorMsg *string
		if row.ErrorCode.Valid && row.ErrorCode.String != "" {
			errorMsg = &row.ErrorCode.String
		}

		executions = append(executions, ExecutionRecord{
			ID:                row.ID,
			ExecutionRootHash: rootHash,
			Tenant:            row.Author,
			FunctionName:      row.FunctionName,
			Timestamp:         row.Timestamp.Format(time.RFC3339),
			NodeSignature:     nodeSig,
			Status:            row.Outcome,
			Duration:          row.DurationMs,
			InputSize:         int(row.InputSize.Int64),
			OutputSize:        int(row.OutputSize.Int64),
			ErrorMessage:      errorMsg,
		})
	}

	data := ExecutionAuditData{
		Executions: executions,
		Total:      int(total),
		Page:       page,
		PageSize:   pageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// HandleGetFraudDetection returns fraud detection data
func (h *OversightHandler) HandleGetFraudDetection(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": claims.UserID,
	}).Info("Fetching fraud detection data")

	fraudResult, err := h.registryRepo.DetectFraudPatterns()
	if err != nil {
		h.logger.WithError(err).Error("Failed to detect fraud patterns")
		// Return empty fraud data so admin UI can load; caller can retry or check logs
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(FraudDetectionData{
			BotPatterns:         []BotPattern{},
			FakeDiversityAlerts: []FakeDiversityAlert{},
			IPClusters:          []IPCluster{},
			WashUsagePatterns:   []WashUsagePattern{},
			Summary:             FraudSummary{},
		})
		return
	}

	// Map bot patterns
	botPatterns := make([]BotPattern, 0, len(fraudResult.BotPatterns))
	for i, pattern := range fraudResult.BotPatterns {
		botPatterns = append(botPatterns, BotPattern{
			ID:                fmt.Sprintf("bot-%03d", i+1),
			PatternType:       pattern.PatternType,
			ConfidenceScore:   pattern.ConfidenceScore,
			AffectedFunctions: pattern.AffectedFunctions,
			AffectedTenants:   pattern.AffectedTenants,
			DetectedAt:        pattern.DetectedAt.Format(time.RFC3339),
			Pattern:           pattern.Pattern,
		})
	}

	// Map IP clusters
	ipClusters := make([]IPCluster, 0, len(fraudResult.IPClusters))
	for i, cluster := range fraudResult.IPClusters {
		ipClusters = append(ipClusters, IPCluster{
			ID:                fmt.Sprintf("ip-%03d", i+1),
			IPRange:           cluster.IPRange,
			AssociatedTenants: cluster.AssociatedTenants,
			RiskLevel:         cluster.RiskLevel,
			CommonPatterns:    cluster.CommonPatterns,
			FirstSeen:         cluster.FirstSeen.Format(time.RFC3339),
			LastSeen:          cluster.LastSeen.Format(time.RFC3339),
		})
	}

	// Map wash usage patterns
	washPatterns := make([]WashUsagePattern, 0, len(fraudResult.WashUsagePatterns))
	for i, pattern := range fraudResult.WashUsagePatterns {
		washPatterns = append(washPatterns, WashUsagePattern{
			ID:                   fmt.Sprintf("wash-%03d", i+1),
			TenantA:              pattern.TenantA,
			TenantB:              pattern.TenantB,
			Function:             pattern.Function,
			Pattern:              pattern.Pattern,
			Confidence:           pattern.Confidence,
			ReciprocalExecutions: pattern.ReciprocalExecutions,
			DetectedAt:           pattern.DetectedAt.Format(time.RFC3339),
		})
	}

	// Map fake diversity alerts from the repo
	fakeAlerts := make([]FakeDiversityAlert, 0, len(fraudResult.FakeDiversityAlerts))
	for i, alert := range fraudResult.FakeDiversityAlerts {
		fakeAlerts = append(fakeAlerts, FakeDiversityAlert{
			ID:             fmt.Sprintf("fake-%03d", i+1),
			TenantGroup:    alert.TenantGroup,
			Indicators:     alert.Indicators,
			RiskLevel:      alert.RiskLevel,
			DetectedAt:     alert.FirstSeen.Format(time.RFC3339),
			CommonPatterns: alert.Indicators,
		})
	}

	data := FraudDetectionData{
		BotPatterns:         botPatterns,
		FakeDiversityAlerts: fakeAlerts,
		IPClusters:          ipClusters,
		WashUsagePatterns:   washPatterns,
		Summary: FraudSummary{
			TotalBotPatterns:  fraudResult.Summary.TotalBotPatterns,
			HighRiskClusters:  fraudResult.Summary.HighRiskClusters,
			SuspiciousTenants: fraudResult.Summary.SuspiciousTenants,
			WashUsageDetected: fraudResult.Summary.WashUsageDetected,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// HandleGetEconomicLeaderboard returns economic monitoring data
func (h *OversightHandler) HandleGetEconomicLeaderboard(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": claims.UserID,
	}).Info("Fetching economic leaderboard data")

	economicResult, err := h.registryRepo.AnalyzeEconomicData()
	if err != nil {
		h.logger.WithError(err).Error("Failed to analyze economic data")
		apierror.WriteError(w, apierror.NewInternal("Failed to load economic leaderboard data"))
		return
	}

	// Map top revenue generators
	revenueGenerators := make([]RevenueGenerator, 0, len(economicResult.TopRevenueGenerators))
	for i, gen := range economicResult.TopRevenueGenerators {
		revenueGenerators = append(revenueGenerators, RevenueGenerator{
			ID:             gen.FunctionID.String(),
			Rank:           i + 1,
			TenantFunction: fmt.Sprintf("%s / %s", gen.Author, gen.FunctionName),
			Revenue30d:     gen.Revenue30d,
			ExecutionCount: gen.ExecutionCount,
			GrowthRate:     gen.GrowthRate,
		})
	}

	// Map suspicious growth alerts
	suspiciousGrowth := make([]SuspiciousGrowthAlert, 0, len(economicResult.SuspiciousGrowth))
	for i, growth := range economicResult.SuspiciousGrowth {
		suspiciousGrowth = append(suspiciousGrowth, SuspiciousGrowthAlert{
			ID:             fmt.Sprintf("growth-%03d", i+1),
			TenantFunction: fmt.Sprintf("%s / %s", growth.Author, growth.FunctionName),
			Pattern:        growth.Pattern,
			Details:        growth.Details,
			DetectedAt:     growth.DetectedAt.Format(time.RFC3339),
		})
	}

	// Map artificial boosting alerts
	artificialBoosting := make([]ArtificialBoosting, 0, len(economicResult.ArtificialBoosting))
	for i, boost := range economicResult.ArtificialBoosting {
		artificialBoosting = append(artificialBoosting, ArtificialBoosting{
			ID:              fmt.Sprintf("boost-%03d", i+1),
			Function:        boost.FunctionName,
			DetectedPattern: boost.DetectedPattern,
			Confidence:      boost.Confidence,
			RelatedAccounts: boost.RelatedAccounts,
		})
	}

	data := EconomicLeaderboardData{
		TopRevenueGenerators: revenueGenerators,
		SuspiciousGrowth:     suspiciousGrowth,
		ArtificialBoosting:   artificialBoosting,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// HandleBlockEntity blocks a suspicious entity (pattern, tenant, IP cluster)
func (h *OversightHandler) HandleBlockEntity(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	entityType := vars["type"]
	entityID := vars["id"]

	h.logger.WithFields(logrus.Fields{
		"user_id":     claims.UserID,
		"entity_type": entityType,
		"entity_id":   entityID,
	}).Info("Blocking entity")

	// Parse the entity ID based on type
	var err error
	switch entityType {
	case "function":
		// Block a function by its version ID
		functionVersionID, parseErr := uuid.Parse(entityID)
		if parseErr != nil {
			apierror.WriteError(w, apierror.NewBadRequest("Invalid function version ID"))
			return
		}
		err = h.registryRepo.BlockFunction(functionVersionID, "Blocked by admin oversight")

	case "tenant", "user":
		// Block a user/tenant by setting their execution quota BlockUntil
		entityUUID, parseErr := uuid.Parse(entityID)
		if parseErr != nil {
			apierror.WriteError(w, apierror.NewBadRequest("Invalid tenant/user ID"))
			return
		}

		err = h.blockUserOrTenant(entityUUID, entityType)

	case "ip":
		// Block an IP address
		err = h.blockIP(entityID)

	case "pattern":
		// Block a specific fraud pattern (e.g., "bot-001", "ip-001", "wash-001")
		// Parse pattern type from ID (e.g., "bot-001" -> "bot")
		patternParts := strings.Split(entityID, "-")
		if len(patternParts) < 2 {
			apierror.WriteError(w, apierror.NewBadRequest("Invalid pattern ID format. Expected format: type-number (e.g., bot-001)"))
			return
		}
		patternType := patternParts[0]
		err = h.blockPattern(patternType, entityID, claims.UserID)

	default:
		apierror.WriteError(w, apierror.NewBadRequest("Unknown entity type. Supported types: function, tenant, user, ip, pattern"))
		return
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to block entity")
		apierror.WriteError(w, apierror.NewInternal("Failed to block entity"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "blocked",
		"message": "Entity has been blocked successfully",
	})
}

// HandleInvestigateEntity opens an investigation on an entity
func (h *OversightHandler) HandleInvestigateEntity(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	entityType := vars["type"]
	entityID := vars["id"]

	h.logger.WithFields(logrus.Fields{
		"user_id":     claims.UserID,
		"entity_type": entityType,
		"entity_id":   entityID,
	}).Info("Opening investigation on entity")

	// Determine investigation reason based on entity type
	reason := fmt.Sprintf("Administrative investigation of %s: %s", entityType, entityID)

	// Add context based on entity type
	switch entityType {
	case "function":
		reason = fmt.Sprintf("Investigation of suspicious function activity: %s", entityID)
	case "user", "tenant":
		reason = fmt.Sprintf("Investigation of suspicious user/tenant activity: %s", entityID)
	case "ip":
		reason = fmt.Sprintf("Investigation of suspicious IP activity: %s", entityID)
	case "pattern":
		reason = fmt.Sprintf("Investigation of fraud pattern: %s", entityID)
	}

	err := h.openInvestigation(entityType, entityID, reason, claims.UserID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to open investigation")
		apierror.WriteError(w, apierror.NewInternal("Failed to open investigation"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "investigating",
		"message": "Investigation opened successfully",
	})
}
