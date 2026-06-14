package recommendations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Service provides recommendation functionality
type Service struct {
	repo           *Repository
	registry       *registry.RegistryRepository
	config         *RecommendationConfig
	flyEmbedSvcURL string // AI service URL for FlyEmbed triple embeddings (e.g. http://localhost:8081)
	flyEmbedAPIKey string // API key for AI service authentication
	httpClient     *http.Client
}

// NewService creates a new recommendations service
func NewService(db *gorm.DB, registryRepo *registry.RegistryRepository, config *RecommendationConfig) *Service {
	if config == nil {
		config = DefaultRecommendationConfig()
	}

	return &Service{
		repo:           NewRepository(db),
		registry:       registryRepo,
		config:         config,
		flyEmbedSvcURL: getEnvOrDefault("FLYEMBED_SVC_URL", ""), // Must be set explicitly
		flyEmbedAPIKey: "",
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}
}

// WithFlyEmbedSvcURL sets the AI service URL for FlyEmbed calls.
func (s *Service) WithFlyEmbedSvcURL(url string) *Service {
	s.flyEmbedSvcURL = url
	return s
}

// WithFlyEmbedAPIKey sets the API key for AI service authentication.
func (s *Service) WithFlyEmbedAPIKey(apiKey string) *Service {
	s.flyEmbedAPIKey = apiKey
	return s
}

// GetRegistryRepository returns the underlying registry repository for cross-handler lookups.
func (s *Service) GetRegistryRepository() *registry.RegistryRepository {
	return s.registry
}

// GetRecommendations gets recommendations based on various strategies
func (s *Service) GetRecommendations(ctx context.Context, req *RecommendationRequest) (*RecommendationResponse, error) {
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Limit > 50 {
		req.Limit = 50
	}

	// Determine which types to include
	typesToInclude := req.Types
	if len(typesToInclude) == 0 {
		typesToInclude = []RecommendationType{
			RecommendationTypeSimilar,
			RecommendationTypeFrequentlyUsedTogether,
			RecommendationTypeSameCategory,
		}
		if s.config.EnableTrending {
			typesToInclude = append(typesToInclude, RecommendationTypeTrending)
		}
		if s.config.EnablePersonalized && req.UserID != nil {
			typesToInclude = append(typesToInclude, RecommendationTypePersonalized)
		}
	}

	// Collect recommendations from different strategies
	var allRecommendations []RecommendationResult

	// 1. Get recommendations based on a specific function
	if req.FunctionID != nil {
		funcRecs, err := s.getFunctionBasedRecommendations(ctx, *req.FunctionID, typesToInclude, req.Limit*2)
		if err != nil {
			logrus.WithError(err).Warn("failed to get function-based recommendations")
		} else {
			allRecommendations = append(allRecommendations, funcRecs...)
		}
	}

	// 2. Get category-based recommendations
	if req.Category != nil {
		catRecs, err := s.getCategoryBasedRecommendations(ctx, *req.Category, req.Limit*2)
		if err != nil {
			logrus.WithError(err).Warn("failed to get category-based recommendations")
		} else {
			allRecommendations = append(allRecommendations, catRecs...)
		}
	}

	// 3. Get personalized recommendations for user
	if s.config.EnablePersonalized && req.UserID != nil {
		personalizedRecs, err := s.getPersonalizedRecommendations(ctx, *req.UserID, req.Limit*2)
		if err != nil {
			logrus.WithError(err).Warn("failed to get personalized recommendations")
		} else {
			allRecommendations = append(allRecommendations, personalizedRecs...)
		}
	}

	// 4. Get trending recommendations
	if s.config.EnableTrending {
		trendingRecs, err := s.getTrendingRecommendations(ctx, req.Limit*2)
		if err != nil {
			logrus.WithError(err).Warn("failed to get trending recommendations")
		} else {
			allRecommendations = append(allRecommendations, trendingRecs...)
		}
	}

	// 5. Search-based recommendations (for use case queries)
	if req.Query != nil && *req.Query != "" {
		searchRecs, err := s.getSearchBasedRecommendations(ctx, *req.Query, req.Limit*2)
		if err != nil {
			logrus.WithError(err).Warn("failed to get search-based recommendations")
		} else {
			allRecommendations = append(allRecommendations, searchRecs...)
		}
	}

	// Deduplicate and re-rank
	allRecommendations = s.deduplicateAndRerank(allRecommendations, req.Limit)

	// Apply pagination
	start := req.Offset
	if start > len(allRecommendations) {
		allRecommendations = []RecommendationResult{}
	} else {
		end := start + req.Limit
		if end > len(allRecommendations) {
			end = len(allRecommendations)
		}
		allRecommendations = allRecommendations[start:end]
	}

	return &RecommendationResponse{
		Recommendations: allRecommendations,
		Total:           len(allRecommendations),
		Limit:           req.Limit,
		Offset:          req.Offset,
		TypesIncluded:   typesToInclude,
	}, nil
}

// GetRelatedFunctions gets related functions for a specific function
func (s *Service) GetRelatedFunctions(ctx context.Context, functionID uuid.UUID, limit int) ([]RecommendationResult, error) {
	if limit == 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}

	// First try to get cached recommendations
	recommendations, err := s.repo.GetRecommendationsForFunction(ctx, functionID, "", limit)
	if err == nil && len(recommendations) > 0 {
		// Convert to results
		return s.convertRecommendationsToResults(ctx, recommendations)
	}

	// Compute fresh recommendations if cache is empty
	if err := s.repo.ComputeAndStoreRecommendations(ctx, functionID, s.config); err != nil {
		logrus.WithError(err).Warn("failed to compute recommendations")
	}

	// Try again after computation
	recommendations, err = s.repo.GetRecommendationsForFunction(ctx, functionID, "", limit)
	if err != nil {
		return nil, err
	}

	if len(recommendations) == 0 {
		// Fall back to content-based similarity
		return s.getContentBasedRelated(ctx, functionID, limit)
	}

	return s.convertRecommendationsToResults(ctx, recommendations)
}

// ComputeSimilarity computes and stores similarity between two functions
func (s *Service) ComputeSimilarity(ctx context.Context, functionIDa, functionIDb uuid.UUID) error {
	// Get function details
	funcA, err := s.registry.GetFunctionByID(ctx, functionIDa)
	if err != nil {
		return err
	}

	funcB, err := s.registry.GetFunctionByID(ctx, functionIDb)
	if err != nil {
		return err
	}

	// Calculate content similarity
	contentSim := s.calculateContentSimilarity(funcA, funcB)

	// Get collaborative similarity
	collaborativeSim := 0.0
	sim, err := s.repo.GetSimilarityBetweenFunctions(ctx, functionIDa, functionIDb)
	if err == nil && sim != nil {
		collaborativeSim = sim.CollaborativeSimilarity
	}

	// Calculate category similarity
	categorySim := s.calculateCategorySimilarity(funcA.Category.String, funcB.Category.String)

	// Calculate combined similarity
	combinedSim := (s.config.ContentWeight * contentSim) +
		(s.config.CollaborativeWeight * collaborativeSim) +
		(s.config.CategoryWeight * categorySim)

	// Store similarity
	// Ensure consistent ordering
	functionIDaOrdered := functionIDa
	functionIDbOrdered := functionIDb
	if functionIDaOrdered.String() > functionIDbOrdered.String() {
		functionIDaOrdered, functionIDbOrdered = functionIDbOrdered, functionIDaOrdered
	}

	return s.repo.UpsertFunctionSimilarity(ctx, &FunctionSimilarity{
		ID:                      uuid.New(),
		FunctionIDA:             functionIDaOrdered,
		FunctionIDB:             functionIDbOrdered,
		ContentSimilarity:       contentSim,
		CollaborativeSimilarity: collaborativeSim,
		CategorySimilarity:      categorySim,
		CombinedSimilarity:      combinedSim,
		ComputedAt:              time.Now(),
	})
}

// RefreshAllRecommendations recomputes recommendations for all functions
func (s *Service) RefreshAllRecommendations(ctx context.Context) error {
	functions, _, err := s.registry.ListFunctions("", "", nil, "public", 1000, 0)
	if err != nil {
		return err
	}

	logrus.Infof("Computing recommendations for %d functions", len(functions))

	for _, fn := range functions {
		if err := s.repo.ComputeAndStoreRecommendations(ctx, fn.ID, s.config); err != nil {
			logrus.WithError(err).WithField("function_id", fn.ID).Warn("failed to compute recommendations")
		}
	}

	return nil
}

// trackCooccurrences tracks function co-occurrences in a session
func (s *Service) trackCooccurrences(ctx context.Context, sessionID string, functionID uuid.UUID) {
	// Get other functions used in this session
	usages, err := s.repo.GetSessionFunctions(ctx, sessionID)
	if err != nil || len(usages) == 0 {
		return
	}

	for _, usage := range usages {
		if usage.FunctionID != functionID {
			// Compute and store similarity
			go func() {
				if err := s.ComputeSimilarity(context.Background(), functionID, usage.FunctionID); err != nil {
					logrus.WithError(err).Warn("failed to compute similarity")
				}
			}()
		}
	}
}

// parseTags parses JSON tags into string slice
func (s *Service) parseTags(tags json.RawMessage) []string {
	if len(tags) == 0 || string(tags) == "null" {
		return nil
	}

	var tagList []string
	if err := json.Unmarshal(tags, &tagList); err != nil {
		return nil
	}
	return tagList
}

// containsType checks if a type is in the list
func (s *Service) containsType(types []RecommendationType, target RecommendationType) bool {
	for _, t := range types {
		if t == target {
			return true
		}
	}
	return false
}

// makeAuthenticatedRequest creates an HTTP request with optional API key and tenant headers
func (s *Service) makeAuthenticatedRequest(ctx context.Context, method, url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Add authentication headers if API key is configured
	if s.flyEmbedAPIKey != "" {
		req.Header.Set("X-API-Key", s.flyEmbedAPIKey)
	}

	// Add tenant context from context if available
	if tenantID, ok := ctx.Value("tenant_id").(string); ok && tenantID != "" {
		req.Header.Set("X-Tenant-ID", tenantID)
	}

	return req, nil
}
