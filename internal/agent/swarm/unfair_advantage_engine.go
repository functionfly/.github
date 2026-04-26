package swarm

import (
	"context"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/discovery"
	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	UnfairAdvantageEngineAgentID = "functionfly/unfair-advantage-engine"
)

type UnfairAdvantageEngine struct {
	db            *gorm.DB
	platformCtrl  *PlatformController
	metricsColl   *MetricsCollector
	initialized   bool
}

func NewUnfairAdvantageEngine(db *gorm.DB, platformCtrl *PlatformController, metricsColl *MetricsCollector) *UnfairAdvantageEngine {
	return &UnfairAdvantageEngine{
		db:           db,
		platformCtrl: platformCtrl,
		metricsColl:  metricsColl,
	}
}

func (uae *UnfairAdvantageEngine) Initialize(ctx context.Context) error {
	if uae.initialized {
		return nil
	}

	if err := uae.ensureEngineAgent(ctx); err != nil {
		return err
	}

	if err := uae.runMigrations(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	uae.initialized = true
	return nil
}

func (uae *UnfairAdvantageEngine) runMigrations(ctx context.Context) error {
	return uae.db.WithContext(ctx).AutoMigrate(
		&InternalOpportunity{},
		&InternalFunction{},
		&RDLabRun{},
		&StealthRun{},
	)
}

func (uae *UnfairAdvantageEngine) ensureEngineAgent(ctx context.Context) error {
	var agent struct {
		ID        uuid.UUID
		AgentID   string
		Name      string
		Status    string
		CreatedAt time.Time
	}

	err := uae.db.WithContext(ctx).Raw(`
		SELECT id, agent_id, name, status, created_at
		FROM agent_identities
		WHERE agent_id = ?
	`, UnfairAdvantageEngineAgentID).Scan(&agent).Error

	if err == gorm.ErrRecordNotFound {
		return uae.createEngineAgent(ctx)
	}
	if err != nil {
		return err
	}
	return nil
}

func (uae *UnfairAdvantageEngine) createEngineAgent(ctx context.Context) error {
	agent := &identity.AgentIdentity{
		ID:                uuid.New(),
		AgentID:           UnfairAdvantageEngineAgentID,
		Name:              "FunctionFly Unfair Advantage Engine",
		Description:       "Private internal R&D lab and money printer for FunctionFly team",
		PlanTier:          "agent_enterprise",
		Status:            identity.AgentStatusActive,
		SwarmRole:         identity.SwarmRoleManager,
		MaxChildAgents:    50,
		Capabilities:      identity.JSONBMap{
			"rd_lab":           true,
			"internal_only":    true,
			"money_printer":    true,
			"competitive_moat": true,
			"stealth_mode":     true,
		},
		AutonomousEnabled: true,
		EvolutionEnabled:  true,
		TrustScore:        100.0,
		EconomicScore:     100.0,
	}
	return uae.db.WithContext(ctx).Create(agent).Error
}

func (uae *UnfairAdvantageEngine) GetDashboard(ctx context.Context) (*AdvantageDashboard, error) {
	status, err := uae.platformCtrl.GetStatus(ctx)
	if err != nil {
		return nil, err
	}

	dailyMetrics, err := uae.metricsColl.CollectDailyMetrics(ctx)
	if err != nil {
		return nil, err
	}

	var internalFuncs int64
	uae.db.WithContext(ctx).Model(&struct{}{}).Table("registry_functions").
		Where("author = ?", "functionfly-ai").Count(&internalFuncs)

	var totalValueGenerated float64
	uae.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(estimated_value_usd), 0)
		FROM internal_opportunities
		WHERE status IN ('generated', 'published')
	`).Scan(&totalValueGenerated)

	var competitiveMeters CompetitiveMeters
	uae.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(SUM(CASE WHEN category = 'cost_savings' THEN estimated_value_usd ELSE 0 END), 0) as cost_savings,
			COALESCE(SUM(CASE WHEN category = 'revenue_boost' THEN estimated_value_usd ELSE 0 END), 0) as revenue_boost,
			COALESCE(SUM(CASE WHEN category = 'speed_gain' THEN estimated_value_usd ELSE 0 END), 0) as speed_gain,
			COALESCE(SUM(CASE WHEN category = 'competitive_moat' THEN estimated_value_usd ELSE 0 END), 0) as competitive_moat
		FROM internal_opportunities
		WHERE status IN ('generated', 'published')
	`).Scan(&competitiveMeters)

	dashboard := &AdvantageDashboard{
		EngineID:          UnfairAdvantageEngineAgentID,
		GeneratedAt:      time.Now(),
		SwarmStatus:      status,
		DailyMetrics:     dailyMetrics,
		InternalFunctions: int(internalFuncs),
		TotalValueGenerated: totalValueGenerated,
		CompetitiveMeters: competitiveMeters,
		Momentum:          calculateMomentum(dailyMetrics),
	}

	return dashboard, nil
}

func (uae *UnfairAdvantageEngine) SeedInternalOpportunity(ctx context.Context, req *SeedOpportunityRequest) (*InternalOpportunity, error) {
	opportunity := &InternalOpportunity{
		ID:                uuid.New(),
		Source:            "internal_rd_lab",
		Title:             req.Title,
		Description:       req.Description,
		Category:          req.Category,
		Tags:              req.Tags,
		Status:            discovery.OpportunityStatusQualified,
		QualityScore:      95.0,
		DemandScore:       90.0,
		EstimatedValueUSD: req.EstimatedValue,
		Priority:          req.Priority,
		Confidential:      true,
		Metadata: map[string]any{
			"seeded_by":       req.SeededBy,
			"business_impact":  req.BusinessImpact,
			"competitive_edge": req.CompetitiveEdge,
			"time_to_market":   req.TimeToMarket,
			"rd_phase":         "ideation",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uae.db.WithContext(ctx).Create(opportunity).Error; err != nil {
		return nil, err
	}

	return opportunity, nil
}

func (uae *UnfairAdvantageEngine) ListInternalOpportunities(ctx context.Context, filter *OpportunityFilter) ([]InternalOpportunity, int64, error) {
	query := uae.db.WithContext(ctx).Model(&InternalOpportunity{})

	if filter != nil {
		if filter.Category != "" {
			query = query.Where("category = ?", filter.Category)
		}
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.Priority != "" {
			query = query.Where("priority = ?", filter.Priority)
		}
		if filter.ConfidentialOnly {
			query = query.Where("confidential = ?", true)
		}
	}

	var total int64
	query.Count(&total)

	var opportunities []InternalOpportunity
	err := query.Order("priority DESC, estimated_value_usd DESC, created_at DESC").
		Limit(filter.Limit).Offset(filter.Offset).Find(&opportunities).Error

	return opportunities, total, err
}

func (uae *UnfairAdvantageEngine) RunInternalRDLab(ctx context.Context) (*RDLabRun, error) {
	run := &RDLabRun{
		ID:           uuid.New(),
		Status:       "running",
		StartedAt:    time.Now(),
		IdeasScouted:  0,
		IdeasFunded:  0,
		Prototypes:   0,
	}

	opportunities, _, err := uae.ListInternalOpportunities(ctx, &OpportunityFilter{
		Status:     discovery.OpportunityStatusQualified,
		Limit:      10,
		Offset:     0,
		Priority:   "high",
	})
	if err != nil {
		return nil, err
	}

	run.IdeasScouted = len(opportunities)

	for _, opp := range opportunities {
		run.IdeasFunded++
		run.TotalValueTracked += opp.EstimatedValueUSD

		uae.db.WithContext(ctx).Model(&InternalOpportunity{}).
			Where("id = ?", opp.ID).
			Update("metadata", gorm.Expr("metadata || ?", map[string]any{
				"rd_lab_processed":    true,
				"rd_lab_run_id":       run.ID.String(),
				"processed_at":       time.Now().Unix(),
			}))
	}

	run.Status = "completed"
	run.CompletedAt = &[]time.Time{time.Now()}[0]
	run.DurationMs = time.Now().UnixMilli() - run.StartedAt.UnixMilli()

	uae.db.WithContext(ctx).Create(run)

	return run, nil
}

func (uae *UnfairAdvantageEngine) GenerateInternalFunction(ctx context.Context, oppID uuid.UUID) (*InternalFunction, error) {
	var opportunity InternalOpportunity
	if err := uae.db.WithContext(ctx).Where("id = ?", oppID).First(&opportunity).Error; err != nil {
		return nil, err
	}

	function := &InternalFunction{
		ID:              uuid.New(),
		Name:            sanitizeFunctionName(opportunity.Title),
		Title:           opportunity.Title,
		Description:     opportunity.Description,
		Category:        opportunity.Category,
		Tags:            opportunity.Tags,
		Status:          "prototype",
		SourceOpprID:    &oppID,
		GeneratedCode:   generateTemplateCode(opportunity),
		EstimatedValue:  opportunity.EstimatedValueUSD,
		TimeToValue:     estimateTimeToValue(opportunity),
		PrivacyLevel:    "internal_team_only",
		CompetitiveEdge: opportunity.Metadata["competitive_edge"],
		RDPhase:         "prototype",
		CreatedAt:       time.Now(),
	}

	if err := uae.db.WithContext(ctx).Create(function).Error; err != nil {
		return nil, err
	}

	uae.db.WithContext(ctx).Model(&InternalOpportunity{}).
		Where("id = ?", oppID).
		Update("status", discovery.OpportunityStatusGenerated)

	return function, nil
}

func (uae *UnfairAdvantageEngine) GetValueReport(ctx context.Context) (*ValueReport, error) {
	var totalInternalValue, totalCostSavings, totalRevenueBoost, totalSpeedGain, totalMoat float64

	var valueResult struct {
		Total           float64
		CostSavings     float64
		RevenueBoost    float64
		SpeedGain       float64
		CompetitiveMoat float64
	}
	err := uae.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(SUM(estimated_value_usd), 0) as total,
			COALESCE(SUM(CASE WHEN category = 'cost_savings' THEN estimated_value_usd ELSE 0 END), 0) as cost_savings,
			COALESCE(SUM(CASE WHEN category = 'revenue_boost' THEN estimated_value_usd ELSE 0 END), 0) as revenue_boost,
			COALESCE(SUM(CASE WHEN category = 'speed_gain' THEN estimated_value_usd ELSE 0 END), 0) as speed_gain,
			COALESCE(SUM(CASE WHEN category = 'competitive_moat' THEN estimated_value_usd ELSE 0 END), 0) as competitive_moat
		FROM internal_opportunities
	`).Scan(&valueResult).Error
	if err != nil {
		return nil, err
	}
	totalInternalValue = valueResult.Total
	totalCostSavings = valueResult.CostSavings
	totalRevenueBoost = valueResult.RevenueBoost
	totalSpeedGain = valueResult.SpeedGain
	totalMoat = valueResult.CompetitiveMoat

	var totalHoursSaved float64
	uae.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(
			(CASE WHEN metadata->>'manual_hours_saved' IS NOT NULL
				THEN (metadata->>'manual_hours_saved')::float * 150 ELSE 0 END) +
			(CASE WHEN metadata->>'dev_hours_saved' IS NOT NULL
				THEN (metadata->>'dev_hours_saved')::float * 200 ELSE 0 END)
		), 0) FROM internal_opportunities WHERE status IN ('generated', 'published')
	`).Scan(&totalHoursSaved)

	var functionsCreated int64
	uae.db.WithContext(ctx).Model(&InternalFunction{}).Count(&functionsCreated)

	return &ValueReport{
		PeriodStart:      time.Now().AddDate(0, 0, -30),
		PeriodEnd:        time.Now(),
		TotalValue:       totalInternalValue,
		CostSavings:      totalCostSavings,
		RevenueBoost:     totalRevenueBoost,
		SpeedGain:        totalSpeedGain,
		CompetitiveMoat:  totalMoat,
		HoursSaved:       totalHoursSaved / 150,
		FunctionsCreated: int(functionsCreated),
		ROI:              (totalInternalValue / 1) * 100,
		MoatDepth:        calculateMoatDepth(totalMoat),
		GeneratedAt:      time.Now(),
	}, nil
}

func (uae *UnfairAdvantageEngine) SeedCustomOpportunity(ctx context.Context, seed *CustomOpportunitySeed) (*discovery.Opportunity, error) {
	opportunity := &discovery.Opportunity{
		ID:                uuid.New(),
		Source:            "internal_custom",
		SourceID:          "custom-" + uuid.New().String()[:8],
		Title:             seed.Title,
		Description:       seed.Description,
		Category:          seed.Category,
		Tags:              seed.Tags,
		DemandScore:       seed.DemandScore,
		QualityScore:      seed.QualityScore,
		Complexity:        seed.Complexity,
		Status:            discovery.OpportunityStatusQualified,
		Validated:         true,
		Metadata: map[string]any{
			"seed_type":      seed.SeedType,
			"target_metric": seed.TargetMetric,
			"expected_moi":  seed.ExpectedMOI,
			"experiment_id": seed.ExperimentID,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uae.db.WithContext(ctx).Create(opportunity).Error; err != nil {
		return nil, err
	}

	return opportunity, nil
}

func (uae *UnfairAdvantageEngine) RunStealthPipeline(ctx context.Context, cfg *StealthConfig) (*StealthRun, error) {
	run := &StealthRun{
		ID:            uuid.New(),
		Status:        "running",
		Mode:          cfg.Mode,
		StartedAt:     time.Now(),
		OpsExecuted:   0,
		FunctionsBuilt: 0,
		ValueGenerated: 0,
	}

	scannerChildren := uae.getChildrenByRole(ctx, "github-scanner")
	for _, scanner := range scannerChildren {
		uae.platformCtrl.DispatchTask(ctx, scanner.AgentID, "stealth_scan", map[string]any{
			"mode":    cfg.Mode,
			"targets": cfg.ScanTargets,
			"depth":   cfg.ScanDepth,
		})
		run.OpsExecuted++
	}

	for i := 0; i < cfg.GeneratorWorkers; i++ {
		genWorkers := uae.getChildrenByRole(ctx, "generator-worker")
		if len(genWorkers) > i {
			uae.platformCtrl.DispatchTask(ctx, genWorkers[i].AgentID, "stealth_generate", map[string]any{
				"quality_floor": cfg.QualityFloor,
				"auto_publish":  cfg.AutoPublish,
			})
			run.OpsExecuted++
		}
	}

	run.Status = "completed"
	run.CompletedAt = &[]time.Time{time.Now()}[0]

	uae.db.WithContext(ctx).Create(run)

	return run, nil
}

func (uae *UnfairAdvantageEngine) getChildrenByRole(ctx context.Context, role string) []*identity.AgentIdentity {
	status, _ := uae.platformCtrl.GetStatus(ctx)
	if status == nil {
		return nil
	}
	var filtered []*identity.AgentIdentity
	for _, child := range status.Children {
		if cap, ok := child.Capabilities["role"].(string); ok && cap == role {
			filtered = append(filtered, &identity.AgentIdentity{
				AgentID: child.AgentID,
				Name:    child.Name,
			})
		}
	}
	return filtered
}

func calculateMomentum(metrics *DailySwarmMetrics) string {
	if metrics == nil {
		return "unknown"
	}
	if metrics.ConversionRate >= 50 && metrics.AverageQualityScore >= 85 {
		return "accelerating"
	}
	if metrics.ConversionRate >= 30 && metrics.AverageQualityScore >= 70 {
		return "steady"
	}
	return "slow"
}

func calculateMoatDepth(moatValue float64) string {
	if moatValue >= 100000 {
		return "massive"
	}
	if moatValue >= 50000 {
		return "significant"
	}
	if moatValue >= 10000 {
		return "growing"
	}
	return "nascent"
}

func sanitizeFunctionName(title string) string {
	name := ""
	for _, c := range title {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			name += string(c)
		}
	}
	if name == "" {
		name = "internal-fn"
	}
	return name
}

func generateTemplateCode(opp InternalOpportunity) string {
	return `# Internal Function Generated by FunctionFly Unfair Advantage Engine
# Source: ` + opp.Source + ` | Priority: ` + opp.Priority + `
# Estimated Value: $` + fmt.Sprintf("%.2f", opp.EstimatedValueUSD) + `

def handle_` + sanitizeFunctionName(opp.Title) + `(event, context):
    """
    ` + opp.Description + `
    """
    # TODO: Implement based on internal requirements
    return {"status": "prototype", "source": "` + opp.Source + `"}
`
}

func estimateTimeToValue(opp InternalOpportunity) string {
	if opp.EstimatedValueUSD >= 50000 {
		return "3-6 months"
	}
	if opp.EstimatedValueUSD >= 10000 {
		return "1-3 months"
	}
	return "2-4 weeks"
}

type EngineAgent = identity.AgentIdentity

type AdvantageDashboard struct {
	EngineID            string             `json:"engine_id"`
	GeneratedAt        time.Time          `json:"generated_at"`
	SwarmStatus        *PlatformStatus    `json:"swarm_status"`
	DailyMetrics       *DailySwarmMetrics `json:"daily_metrics"`
	InternalFunctions   int                `json:"internal_functions"`
	TotalValueGenerated float64           `json:"total_value_generated"`
	CompetitiveMeters  CompetitiveMeters   `json:"competitive_meters"`
	Momentum           string             `json:"momentum"`
}

type CompetitiveMeters struct {
	CostSavings     float64 `json:"cost_savings"`
	RevenueBoost    float64 `json:"revenue_boost"`
	SpeedGain       float64 `json:"speed_gain"`
	CompetitiveMoat float64 `json:"competitive_moat"`
}

type InternalOpportunity struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Source            string         `json:"source" gorm:"not null"`
	SourceID          string         `json:"source_id" gorm:"not null"`
	Title             string         `json:"title" gorm:"not null"`
	Description       string         `json:"description" gorm:"type:text"`
	Category          string         `json:"category" gorm:"not null"`
	Tags              []string       `json:"tags" gorm:"serializer:json"`
	DemandScore       float64        `json:"demand_score" gorm:"type:decimal(5,2)"`
	QualityScore      float64        `json:"quality_score" gorm:"type:decimal(5,2)"`
	Complexity        int            `json:"complexity" gorm:"not null;default:1"`
	Validated         bool           `json:"validated" gorm:"not null;default:true"`
	Status            string         `json:"status" gorm:"not null;default:'pending'"`
	EstimatedValueUSD float64        `json:"estimated_value_usd" gorm:"type:decimal(12,2)"`
	Priority          string         `json:"priority" gorm:"default:'medium'"`
	Confidential      bool           `json:"confidential" gorm:"not null;default:true"`
	Metadata          map[string]any `json:"metadata" gorm:"serializer:json;default:'{}'"`
	CreatedAt         time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

func (InternalOpportunity) TableName() string { return "internal_opportunities" }

type SeedOpportunityRequest struct {
	Title           string   `json:"title" validate:"required"`
	Description     string   `json:"description"`
	Category        string   `json:"category"`
	Tags            []string `json:"tags"`
	EstimatedValue  float64  `json:"estimated_value"`
	Priority        string   `json:"priority"`
	SeededBy        string   `json:"seeded_by"`
	BusinessImpact  string   `json:"business_impact"`
	CompetitiveEdge string   `json:"competitive_edge"`
	TimeToMarket    string   `json:"time_to_market"`
}

type OpportunityFilter struct {
	Category         string
	Status           string
	Priority         string
	ConfidentialOnly bool
	Limit            int
	Offset           int
}

type InternalFunction struct {
	ID              uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name            string     `json:"name" gorm:"not null"`
	Title           string     `json:"title" gorm:"not null"`
	Description     string     `json:"description" gorm:"type:text"`
	Category        string     `json:"category" gorm:"not null"`
	Tags            []string   `json:"tags" gorm:"serializer:json"`
	Status          string     `json:"status" gorm:"not null;default:'prototype'"`
	SourceOpprID    *uuid.UUID `json:"source_opportunity_id" gorm:"type:uuid"`
	GeneratedCode   string     `json:"generated_code" gorm:"type:text"`
	EstimatedValue  float64    `json:"estimated_value" gorm:"type:decimal(12,2)"`
	TimeToValue     string     `json:"time_to_value"`
	PrivacyLevel    string     `json:"privacy_level" gorm:"default:'internal_team_only'"`
	CompetitiveEdge any        `json:"competitive_edge"`
	RDPhase         string     `json:"rd_phase" gorm:"default:'prototype'"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

func (InternalFunction) TableName() string { return "internal_functions" }

type RDLabRun struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Status        string    `json:"status" gorm:"not null"`
	StartedAt     time.Time `json:"started_at" gorm:"autoCreateTime"`
	CompletedAt   *time.Time `json:"completed_at"`
	DurationMs    int64     `json:"duration_ms"`
	IdeasScouted  int       `json:"ideas_scouted"`
	IdeasFunded   int       `json:"ideas_funded"`
	Prototypes    int       `json:"prototypes"`
	TotalValueTracked float64 `json:"total_value_tracked"`
}

func (RDLabRun) TableName() string { return "rd_lab_runs" }

type ValueReport struct {
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
	TotalValue       float64   `json:"total_value"`
	CostSavings      float64   `json:"cost_savings"`
	RevenueBoost     float64   `json:"revenue_boost"`
	SpeedGain        float64   `json:"speed_gain"`
	CompetitiveMoat  float64   `json:"competitive_moat"`
	HoursSaved       float64   `json:"hours_saved"`
	FunctionsCreated int       `json:"functions_created"`
	ROI              float64   `json:"roi"`
	MoatDepth        string    `json:"moat_depth"`
	GeneratedAt      time.Time `json:"generated_at"`
}

type CustomOpportunitySeed struct {
	Title         string   `json:"title" validate:"required"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Tags          []string `json:"tags"`
	DemandScore   float64  `json:"demand_score"`
	QualityScore  float64  `json:"quality_score"`
	Complexity    int      `json:"complexity"`
	SeedType      string   `json:"seed_type"`
	TargetMetric  string   `json:"target_metric"`
	ExpectedMOI   string   `json:"expected_moi"`
	ExperimentID  string   `json:"experiment_id"`
}

type StealthConfig struct {
	Mode            string   `json:"mode"`
	ScanTargets     []string `json:"scan_targets"`
	ScanDepth       int      `json:"scan_depth"`
	GeneratorWorkers int     `json:"generator_workers"`
	QualityFloor    float64  `json:"quality_floor"`
	AutoPublish     bool     `json:"auto_publish"`
}

type StealthRun struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Status         string    `json:"status"`
	Mode           string    `json:"mode"`
	StartedAt      time.Time `json:"started_at" gorm:"autoCreateTime"`
	CompletedAt    *time.Time `json:"completed_at"`
	OpsExecuted    int       `json:"ops_executed"`
	FunctionsBuilt int       `json:"functions_built"`
	ValueGenerated float64   `json:"value_generated"`
}

func (StealthRun) TableName() string { return "stealth_runs" }