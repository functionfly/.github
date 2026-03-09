package factory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/agent/categorization"
	"github.com/functionfly/functionfly/internal/agent/deployment"
	"github.com/functionfly/functionfly/internal/agent/discovery"
	"github.com/functionfly/functionfly/internal/agent/generation"
	"github.com/functionfly/functionfly/internal/agent/testing"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type FactoryRun struct {
	ID                   uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID              string         `json:"agent_id" gorm:"not null;index"`
	Status               string         `json:"status" gorm:"not null;default:'pending';index"`
	OpportunitiesScanned int            `json:"opportunities_scanned"`
	FunctionsGenerated   int            `json:"functions_generated"`
	FunctionsPublished   int            `json:"functions_published"`
	AverageQualityScore  float64        `json:"average_quality_score" gorm:"type:decimal(5,2);default:0"`
	ErrorMessage         *string        `json:"error_message" gorm:"type:text"`
	Metadata             map[string]any `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt            time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	CompletedAt          *time.Time     `json:"completed_at"`
}

func (FactoryRun) TableName() string { return "factory_runs" }

type FactoryVersion struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RunID          uuid.UUID      `json:"run_id" gorm:"type:uuid;not null;index"`
	OpportunityID  uuid.UUID      `json:"opportunity_id" gorm:"type:uuid;not null;index"`
	FunctionID     string         `json:"function_id" gorm:"not null"`
	GeneratedCode  string         `json:"generated_code" gorm:"type:text"`
	Manifest       string         `json:"manifest" gorm:"type:text"`
	ModelUsed      string         `json:"model_used"`
	QualityScore   float64        `json:"quality_score" gorm:"type:decimal(5,2);default:0"`
	TestScore      float64        `json:"test_score" gorm:"type:decimal(5,2);default:0"`
	ReviewRequired bool           `json:"review_required" gorm:"not null;default:false"`
	Metadata       map[string]any `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
}

func (FactoryVersion) TableName() string { return "factory_versions" }

type Service struct {
	db          *gorm.DB
	config      Config
	discovery   *discovery.Service
	generation  *generation.Service
	testing     *testing.Service
	publisher   *deployment.Publisher
	categorizer *categorization.Service
	// Experiment support
	experimentAdapter *GenerationExperimentAdapter
}

func NewService(db *gorm.DB, config Config, discoverySvc *discovery.Service, generationSvc *generation.Service, testingSvc *testing.Service, publisher *deployment.Publisher) *Service {
	return &Service{
		db:          db,
		config:      config,
		discovery:   discoverySvc,
		generation:  generationSvc,
		testing:     testingSvc,
		publisher:   publisher,
		categorizer: categorization.NewService(db),
	}
}

// NewServiceWithCategorizer creates a factory service with a custom categorizer
func NewServiceWithCategorizer(db *gorm.DB, config Config, discoverySvc *discovery.Service, generationSvc *generation.Service, testingSvc *testing.Service, publisher *deployment.Publisher, categorizer *categorization.Service) *Service {
	return &Service{
		db:          db,
		config:      config,
		discovery:   discoverySvc,
		generation:  generationSvc,
		testing:     testingSvc,
		publisher:   publisher,
		categorizer: categorizer,
	}
}

// NewServiceWithExperiment creates a factory service with experiment support
func NewServiceWithExperiment(db *gorm.DB, config Config, discoverySvc *discovery.Service, generationSvc *generation.Service, testingSvc *testing.Service, publisher *deployment.Publisher, experimentAdapter *GenerationExperimentAdapter) *Service {
	svc := NewServiceWithCategorizer(db, config, discoverySvc, generationSvc, testingSvc, publisher, categorization.NewService(db))
	svc.experimentAdapter = experimentAdapter
	return svc
}

func (s *Service) AutoMigrate(ctx context.Context) error {
	if err := s.db.WithContext(ctx).AutoMigrate(&FactoryRun{}, &FactoryVersion{}, &FactoryConfig{}); err != nil {
		return err
	}
	// Also migrate experiment tables
	if err := s.db.WithContext(ctx).AutoMigrate(&Experiment{}, &ExperimentVariant{}, &ExperimentMetric{}); err != nil {
		return err
	}
	// Also migrate categorization tables
	return s.categorizer.AutoMigrate(ctx)
}

func (s *Service) Run(ctx context.Context) (*FactoryRun, error) {
	run := &FactoryRun{ID: uuid.New(), AgentID: s.config.AgentID, Status: RunStatusRunning, Metadata: map[string]any{}}
	if err := s.db.WithContext(ctx).Create(run).Error; err != nil {
		return nil, err
	}
	if err := s.execute(ctx, run); err != nil {
		msg := err.Error()
		run.Status = RunStatusFailed
		run.ErrorMessage = &msg
		now := time.Now().UTC()
		run.CompletedAt = &now
		run.UpdatedAt = now
		_ = s.db.WithContext(ctx).Save(run).Error
		return run, err
	}
	now := time.Now().UTC()
	run.Status = RunStatusSucceeded
	run.CompletedAt = &now
	run.UpdatedAt = now
	return run, s.db.WithContext(ctx).Save(run).Error
}

func (s *Service) execute(ctx context.Context, run *FactoryRun) error {
	opportunities, err := s.discovery.ListQualified(ctx, s.config.DiscoveryBatchSize)
	if err != nil {
		return err
	}
	run.OpportunitiesScanned = len(opportunities)
	processed := 0
	totalQuality := 0.0
	for _, opportunity := range opportunities {
		if processed >= s.config.MaxOpportunitiesPerRun {
			break
		}
		version, publishCount, err := s.processOpportunity(ctx, run.ID, opportunity)
		if err != nil {
			return err
		}
		processed++
		run.FunctionsGenerated++
		run.FunctionsPublished += publishCount
		totalQuality += version.QualityScore
	}
	if processed > 0 {
		run.AverageQualityScore = totalQuality / float64(processed)
	}
	return s.db.WithContext(ctx).Save(run).Error
}

func (s *Service) processOpportunity(ctx context.Context, runID uuid.UUID, opportunity discovery.Opportunity) (*FactoryVersion, int, error) {
	request := &generation.GenerationRequest{
		AgentID:       s.config.AgentID,
		Name:          opportunity.Title,
		Description:   opportunity.Description,
		Category:      opportunity.Category,
		Runtime:       "python3.11",
		Prompt:        opportunity.Description,
		Deterministic: true,
		Tags:          opportunity.Tags,
	}

	// Experiment: Try to assign a variant and use its prompt template
	var experimentID, variantID *uuid.UUID
	if s.experimentAdapter != nil {
		prompt, expID, varID, err := s.experimentAdapter.GetPromptForGeneration(ctx, s.config.AgentID, request.Prompt)
		if err == nil && expID != nil {
			request.Prompt = prompt
			experimentID = expID
			variantID = varID
			logrus.Debugf("Using experiment variant for generation: experiment=%s variant=%s", expID, varID)
		}
	}

	genResult, err := s.generation.GenerateFunction(ctx, request)
	if err != nil {
		return nil, 0, err
	}
	testResults, err := s.testing.RunTests(ctx, genResult.FunctionID, genResult.Code, request.Runtime)
	if err != nil {
		return nil, 0, err
	}
	testScore := testing.AggregateScore(testResults)
	qualityScore := opportunity.QualityScore
	reviewRequired := genResult.ReviewRequired || qualityScore < s.config.MinimumQualityScore || testScore < s.config.MinimumTestScore

	// Auto-categorize the generated function
	catResult, err := s.categorizer.CategorizeFunction(ctx, &categorization.FunctionSpec{
		Name:         request.Name,
		Description:  request.Description,
		InputSchema:  request.InputSchema,
		OutputSchema: request.OutputSchema,
		Code:         genResult.Code,
		Runtime:      request.Runtime,
	})
	if err != nil {
		logrus.WithError(err).Warn("failed to auto-categorize function, using opportunity category")
	} else {
		// Override category and tags with auto-categorization results if confidence is high
		if catResult.Confidence > 0.6 {
			request.Category = catResult.PrimaryCategory
			if len(catResult.Tags) > 0 {
				request.Tags = catResult.Tags
			}
		}
		// Store categorization result
		if _, err := s.categorizer.CategorizeAndStore(ctx, genResult.FunctionID, &categorization.FunctionSpec{
			Name:         request.Name,
			Description:  request.Description,
			InputSchema:  request.InputSchema,
			OutputSchema: request.OutputSchema,
			Code:         genResult.Code,
			Runtime:      request.Runtime,
		}); err != nil {
			logrus.WithError(err).Warn("failed to store categorization result")
		}
	}

	version := &FactoryVersion{
		ID:             uuid.New(),
		RunID:          runID,
		OpportunityID:  opportunity.ID,
		FunctionID:     genResult.FunctionID.String(),
		GeneratedCode:  genResult.Code,
		Manifest:       genResult.Manifest,
		ModelUsed:      genResult.ModelUsed,
		QualityScore:   qualityScore,
		TestScore:      testScore,
		ReviewRequired: reviewRequired,
		Metadata: map[string]any{
			"review_reason":       genResult.ReviewReason,
			"test_results":        testResults,
			"auto_category":       catResult.PrimaryCategory,
			"auto_tags":           catResult.Tags,
			"category_confidence": catResult.Confidence,
		},
	}
	if err := s.db.WithContext(ctx).Create(version).Error; err != nil {
		return nil, 0, err
	}

	// Record experiment metrics if an experiment variant was used
	if s.experimentAdapter != nil && experimentID != nil && variantID != nil {
		result := GenerationResult{
			FunctionID:     genResult.FunctionID,
			Success:        genResult.Success,
			QualityScore:   qualityScore,
			TestScore:      testScore,
			AllTestsPassed: testing.AllPassed(testResults),
		}
		if err := s.experimentAdapter.RecordGenerationResult(ctx, *experimentID, *variantID, result); err != nil {
			logrus.WithError(err).Warnf("failed to record experiment metrics for variant %s", *variantID)
		}
	}

	if err := s.discovery.MarkGenerated(ctx, opportunity.ID.String(), genResult.FunctionID.String()); err != nil {
		return nil, 0, err
	}
	if reviewRequired || !testing.AllPassed(testResults) || !s.config.AutoPublish {
		return version, 0, s.db.WithContext(ctx).Model(&FactoryRun{}).Where("id = ?", runID).Update("status", RunStatusReview).Error
	}
	_, err = s.publisher.Publish(ctx, &deployment.PublishRequest{
		AgentID:         s.config.AgentID,
		GeneratedCodeID: genResult.FunctionID,
		GeneratedCode: &deployment.GeneratedCode{
			ID:            genResult.FunctionID,
			AgentID:       s.config.AgentID,
			GeneratedCode: genResult.Code,
			Language:      runtimeToLanguage(request.Runtime),
			Runtime:       request.Runtime,
			ModelUsed:     genResult.ModelUsed,
			Status:        deployment.GenerationStatusSuccess,
			Request: deployment.FunctionSpec{
				Name:         request.Name,
				Title:        request.Name,
				Description:  request.Description,
				Prompt:       request.Prompt,
				InputSchema:  request.InputSchema,
				OutputSchema: request.OutputSchema,
				Category:     request.Category,
				Tags:         request.Tags,
			},
		},
		Author:      "functionfly-ai",
		Name:        request.Name,
		Title:       request.Name,
		Description: request.Description,
		Category:    request.Category,
		Tags:        request.Tags,
		IsPublic:    true,
	})
	if err != nil {
		return version, 0, fmt.Errorf("publish generated function: %w", err)
	}
	return version, 1, nil
}

func runtimeToLanguage(runtime string) string {
	runtime = strings.ToLower(runtime)
	if strings.Contains(runtime, "node") || strings.Contains(runtime, "javascript") || strings.Contains(runtime, "js") {
		return "javascript"
	}
	return "python"
}

// GetConfig returns the current factory configuration from the database.
// If no config exists, it returns the default config.
func (s *Service) GetConfig(ctx context.Context) (Config, error) {
	var fc FactoryConfig
	// Use silent logger for the lookup so expected "record not found" is not logged
	db := s.db.WithContext(ctx).Session(&gorm.Session{Logger: s.db.Logger.LogMode(logger.Silent)})
	err := db.Unscoped().First(&fc, "agent_id = ?", s.config.AgentID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create default config in database if not exists
			fc = FactoryConfig{
				ID:                     s.config.AgentID,
				AgentID:                s.config.AgentID,
				DiscoveryBatchSize:     s.config.DiscoveryBatchSize,
				MinimumQualityScore:    s.config.MinimumQualityScore,
				MinimumTestScore:       s.config.MinimumTestScore,
				RequireAllTestsPass:    s.config.RequireAllTestsPass,
				AutoPublish:            s.config.AutoPublish,
				MaxOpportunitiesPerRun: s.config.MaxOpportunitiesPerRun,
				RetryAttempts:          s.config.RetryAttempts,
				RetryBackoffMs:         int(s.config.RetryBackoff.Milliseconds()),
				ScheduleEnabled:        s.config.ScheduleEnabled,
				ScheduleCron:           s.config.ScheduleCron,
				ScheduleTimezone:       s.config.ScheduleTimezone,
			}
			if err := s.db.WithContext(ctx).Unscoped().Create(&fc).Error; err != nil {
				return s.config, err
			}
			return s.config, nil
		}
		return s.config, err
	}
	return fc.ToConfig(), nil
}

// SaveConfig saves the factory configuration to the database.
func (s *Service) SaveConfig(ctx context.Context, cfg Config) error {
	var fc FactoryConfig
	err := s.db.WithContext(ctx).Unscoped().First(&fc, "agent_id = ?", cfg.AgentID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new config
			fc = FactoryConfig{
				ID: cfg.AgentID,
			}
			fc.UpdateFrom(cfg)
			return s.db.WithContext(ctx).Unscoped().Create(&fc).Error
		}
		return err
	}
	// Update existing config
	fc.UpdateFrom(cfg)
	return s.db.WithContext(ctx).Unscoped().Save(&fc).Error
}
