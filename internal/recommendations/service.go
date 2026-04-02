package recommendations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Service provides recommendation functionality
type Service struct {
	repo           *Repository
	registry       *registry.RegistryRepository
	config         *RecommendationConfig
	flyEmbedSvcURL string // AI service URL for FlyEmbed triple embeddings (e.g. http://localhost:8081)
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
		flyEmbedSvcURL: "http://localhost:8081",
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}
}

// WithFlyEmbedSvcURL sets the AI service URL for FlyEmbed calls.
func (s *Service) WithFlyEmbedSvcURL(url string) *Service {
	s.flyEmbedSvcURL = url
	return s
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

// UpsertFunctionEmbedding stores or updates the vector embedding for a function (pgvector vector(1536)).
func (s *Service) UpsertFunctionEmbedding(ctx context.Context, functionID uuid.UUID, embedding []float32, embeddedText *string, model string) error {
	if model == "" {
		model = "text-embedding-ada-002"
	}
	return s.repo.UpsertFunctionEmbedding(ctx, &FunctionEmbedding{
		FunctionID:     functionID,
		Embedding:      embedding,
		EmbeddedText:   embeddedText,
		EmbeddingModel: model,
	})
}

// SearchSimilarByEmbedding returns recommendations by vector similarity (cosine). Score = 1 - cosine_distance.
// excludeFunctionID is optional (e.g. the query function). Requires registry to resolve function details.
func (s *Service) SearchSimilarByEmbedding(ctx context.Context, queryEmbedding []float32, limit int, excludeFunctionID *uuid.UUID) ([]RecommendationResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	withDist, err := s.repo.SearchFunctionEmbeddingsByVector(ctx, queryEmbedding, limit, excludeFunctionID)
	if err != nil || len(withDist) == 0 {
		return nil, err
	}
	if s.registry == nil {
		return nil, nil
	}
	var results []RecommendationResult
	for _, e := range withDist {
		fn, err := s.registry.GetFunctionByID(e.FunctionID)
		if err != nil {
			continue
		}
		// Cosine distance is in [0, 2]; similarity = 1 - distance, clamped to [0, 1].
		score := 1 - e.Distance
		if score < 0 {
			score = 0
		}
		if score > 1 {
			score = 1
		}
		results = append(results, RecommendationResult{
			FunctionID:         fn.ID,
			Author:             fn.Author,
			Name:               fn.Name,
			Title:              fn.Title.String,
			Description:        fn.Description.String,
			Category:           fn.Category.String,
			Tags:               s.parseTags(fn.Tags),
			PopularityScore:    fn.PopularityScore,
			ReliabilityScore:   fn.ReliabilityScore,
			Score:              score,
			RecommendationType: RecommendationTypeSimilar,
		})
	}
	return results, nil
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

// RecordInteraction records a user interaction for future recommendations
func (s *Service) RecordInteraction(ctx context.Context, userID *uuid.UUID, functionID uuid.UUID, interactionType InteractionType, sessionID *string, referrerFunctionID *uuid.UUID) error {
	interaction := &UserFunctionInteraction{
		ID:                 uuid.New(),
		UserID:             userID,
		FunctionID:         functionID,
		InteractionType:    interactionType,
		SessionID:          sessionID,
		ReferrerFunctionID: referrerFunctionID,
		Timestamp:          time.Now(),
	}

	if err := s.repo.RecordUserInteraction(ctx, interaction); err != nil {
		return err
	}

	// If it's an execute interaction, also track co-occurrences
	if interactionType == InteractionTypeExecute && sessionID != nil {
		s.trackCooccurrences(ctx, *sessionID, functionID)
	}

	return nil
}

// RecordExecution records a function execution for recommendations
func (s *Service) RecordExecution(ctx context.Context, userID *uuid.UUID, functionID uuid.UUID, sessionID string) error {
	// Record session usage
	usage := &SessionFunctionUsage{
		ID:             uuid.New(),
		SessionID:      sessionID,
		FunctionID:     functionID,
		UserID:         userID,
		ExecutionCount: 1,
		FirstUsedAt:    time.Now(),
		LastUsedAt:     time.Now(),
	}

	if err := s.repo.RecordSessionUsage(ctx, usage); err != nil {
		logrus.WithError(err).Warn("failed to record session usage")
	}

	// Track co-occurrences
	s.trackCooccurrences(ctx, sessionID, functionID)

	// Record interaction
	return s.RecordInteraction(ctx, userID, functionID, InteractionTypeExecute, &sessionID, nil)
}

// RecordFeedback records recommendation feedback
func (s *Service) RecordFeedback(ctx context.Context, userID *uuid.UUID, functionID, recommendedFunctionID uuid.UUID, feedbackType string, recommendationType *string) error {
	feedback := &RecommendationFeedback{
		ID:                    uuid.New(),
		UserID:                userID,
		FunctionID:            functionID,
		RecommendedFunctionID: recommendedFunctionID,
		FeedbackType:          feedbackType,
		RecommendationType:    recommendationType,
		Timestamp:             time.Now(),
	}

	return s.repo.RecordFeedback(ctx, feedback)
}

// ComputeSimilarity computes and stores similarity between two functions
func (s *Service) ComputeSimilarity(ctx context.Context, functionIDa, functionIDb uuid.UUID) error {
	// Get function details
	funcA, err := s.registry.GetFunctionByID(functionIDa)
	if err != nil {
		return err
	}

	funcB, err := s.registry.GetFunctionByID(functionIDb)
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

// getFunctionBasedRecommendations gets recommendations based on a function
func (s *Service) getFunctionBasedRecommendations(ctx context.Context, functionID uuid.UUID, types []RecommendationType, limit int) ([]RecommendationResult, error) {
	var results []RecommendationResult

	// Get similarity-based recommendations
	if s.containsType(types, RecommendationTypeSimilar) {
		sims, err := s.repo.GetSimilaritiesForFunction(ctx, functionID, limit)
		if err == nil && len(sims) > 0 {
			for _, sim := range sims {
				if sim.CombinedSimilarity < s.config.MinSimilarityScore {
					continue
				}

				otherID := sim.FunctionIDB
				if sim.FunctionIDA == functionID {
					otherID = sim.FunctionIDB
				} else {
					otherID = sim.FunctionIDA
				}

				fn, err := s.registry.GetFunctionByID(otherID)
				if err != nil {
					continue
				}

				results = append(results, RecommendationResult{
					FunctionID:         fn.ID,
					Author:             fn.Author,
					Name:               fn.Name,
					Title:              fn.Title.String,
					Description:        fn.Description.String,
					Category:           fn.Category.String,
					Tags:               s.parseTags(fn.Tags),
					PopularityScore:    fn.PopularityScore,
					ReliabilityScore:   fn.ReliabilityScore,
					Score:              sim.CombinedSimilarity,
					RecommendationType: RecommendationTypeSimilar,
				})
			}
		}
	}

	// Get collaborative filtering recommendations
	if s.containsType(types, RecommendationTypeFrequentlyUsedTogether) {
		coocs, err := s.repo.GetFunctionCooccurrences(ctx, functionID, limit)
		if err == nil && len(coocs) > 0 {
			maxCount := 1
			for _, co := range coocs {
				if co.CooccurrenceCount > maxCount {
					maxCount = co.CooccurrenceCount
				}
			}

			for _, co := range coocs {
				otherID := co.FunctionIDB
				if co.FunctionIDA == functionID {
					otherID = co.FunctionIDB
				} else {
					otherID = co.FunctionIDA
				}

				fn, err := s.registry.GetFunctionByID(otherID)
				if err != nil {
					continue
				}

				score := float64(co.CooccurrenceCount) / float64(maxCount)

				results = append(results, RecommendationResult{
					FunctionID:         fn.ID,
					Author:             fn.Author,
					Name:               fn.Name,
					Title:              fn.Title.String,
					Description:        fn.Description.String,
					Category:           fn.Category.String,
					Tags:               s.parseTags(fn.Tags),
					PopularityScore:    fn.PopularityScore,
					ReliabilityScore:   fn.ReliabilityScore,
					Score:              score,
					RecommendationType: RecommendationTypeFrequentlyUsedTogether,
				})
			}
		}
	}

	// Get category-based recommendations
	if s.containsType(types, RecommendationTypeSameCategory) {
		fn, err := s.registry.GetFunctionByID(functionID)
		if err == nil && fn.Category.Valid && fn.Category.String != "" {
			catFuncs, err := s.getFunctionsInCategory(ctx, fn.Category.String, limit)
			if err == nil {
				for _, catFn := range catFuncs {
					if catFn.ID == functionID {
						continue
					}
					results = append(results, RecommendationResult{
						FunctionID:         catFn.ID,
						Author:             catFn.Author,
						Name:               catFn.Name,
						Title:              catFn.Title.String,
						Description:        catFn.Description.String,
						Category:           catFn.Category.String,
						Tags:               s.parseTags(catFn.Tags),
						PopularityScore:    catFn.PopularityScore,
						ReliabilityScore:   catFn.ReliabilityScore,
						Score:              0.5,
						RecommendationType: RecommendationTypeSameCategory,
					})
				}
			}
		}
	}

	return results, nil
}

// getCategoryBasedRecommendations gets recommendations for a category
func (s *Service) getCategoryBasedRecommendations(ctx context.Context, category string, limit int) ([]RecommendationResult, error) {
	functions, err := s.getFunctionsInCategory(ctx, category, limit)
	if err != nil {
		return nil, err
	}

	var results []RecommendationResult
	for _, fn := range functions {
		results = append(results, RecommendationResult{
			FunctionID:         fn.ID,
			Author:             fn.Author,
			Name:               fn.Name,
			Title:              fn.Title.String,
			Description:        fn.Description.String,
			Category:           fn.Category.String,
			Tags:               s.parseTags(fn.Tags),
			PopularityScore:    fn.PopularityScore,
			ReliabilityScore:   fn.ReliabilityScore,
			Score:              0.8,
			RecommendationType: RecommendationTypeSameCategory,
		})
	}

	return results, nil
}

// getPersonalizedRecommendations gets personalized recommendations for a user
func (s *Service) getPersonalizedRecommendations(ctx context.Context, userID uuid.UUID, limit int) ([]RecommendationResult, error) {
	// Get functions the user has executed
	executedFuncs, err := s.repo.GetUserExecutedFunctions(ctx, userID, 20)
	if err != nil || len(executedFuncs) == 0 {
		// Fall back to viewed functions
		executedFuncs, err = s.repo.GetUserViewedFunctions(ctx, userID, 20)
		if err != nil || len(executedFuncs) == 0 {
			return nil, err
		}
	}

	// Get similar functions to what the user has interacted with
	var results []RecommendationResult
	for _, funcID := range executedFuncs {
		sims, err := s.repo.GetSimilaritiesForFunction(ctx, funcID, 5)
		if err != nil {
			continue
		}

		for _, sim := range sims {
			if sim.CombinedSimilarity < s.config.MinSimilarityScore {
				continue
			}

			otherID := sim.FunctionIDB
			if sim.FunctionIDA == funcID {
				otherID = sim.FunctionIDB
			} else {
				otherID = sim.FunctionIDA
			}

			fn, err := s.registry.GetFunctionByID(otherID)
			if err != nil {
				continue
			}

			results = append(results, RecommendationResult{
				FunctionID:         fn.ID,
				Author:             fn.Author,
				Name:               fn.Name,
				Title:              fn.Title.String,
				Description:        fn.Description.String,
				Category:           fn.Category.String,
				Tags:               s.parseTags(fn.Tags),
				PopularityScore:    fn.PopularityScore,
				ReliabilityScore:   fn.ReliabilityScore,
				Score:              sim.CombinedSimilarity * 0.8, // Boost for personalization
				RecommendationType: RecommendationTypePersonalized,
			})
		}

		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

// getTrendingRecommendations gets trending functions
func (s *Service) getTrendingRecommendations(ctx context.Context, limit int) ([]RecommendationResult, error) {
	// Use repository's trending function
	trendingFuncIDs, err := s.repo.GetTrendingFunctions(ctx, limit)
	if err != nil {
		return nil, err
	}

	var results []RecommendationResult
	for _, funcID := range trendingFuncIDs {
		fn, err := s.registry.GetFunctionByID(funcID)
		if err != nil {
			continue
		}

		results = append(results, RecommendationResult{
			FunctionID:         fn.ID,
			Author:             fn.Author,
			Name:               fn.Name,
			Title:              fn.Title.String,
			Description:        fn.Description.String,
			Category:           fn.Category.String,
			Tags:               s.parseTags(fn.Tags),
			PopularityScore:    fn.PopularityScore,
			ReliabilityScore:   fn.ReliabilityScore,
			Score:              0.7,
			RecommendationType: RecommendationTypeTrending,
		})
	}

	return results, nil
}

// getSearchBasedRecommendations gets recommendations based on a search query
func (s *Service) getSearchBasedRecommendations(ctx context.Context, query string, limit int) ([]RecommendationResult, error) {
	functions, _, err := s.registry.SearchFunctions(query, "", "", 0, limit, 0)
	if err != nil {
		return nil, err
	}

	var results []RecommendationResult
	for _, fn := range functions {
		results = append(results, RecommendationResult{
			FunctionID:         fn.ID,
			Author:             fn.Author,
			Name:               fn.Name,
			Title:              fn.Title.String,
			Description:        fn.Description.String,
			Category:           fn.Category.String,
			Tags:               s.parseTags(fn.Tags),
			PopularityScore:    fn.PopularityScore,
			ReliabilityScore:   fn.ReliabilityScore,
			Score:              0.6,
			RecommendationType: RecommendationTypePersonalized, // Using this type as it's derived from query
		})
	}

	return results, nil
}

// getContentBasedRelated gets related functions using content similarity
func (s *Service) getContentBasedRelated(ctx context.Context, functionID uuid.UUID, limit int) ([]RecommendationResult, error) {
	fn, err := s.registry.GetFunctionByID(functionID)
	if err != nil {
		return nil, err
	}

	// Search for similar functions
	query := fn.Title.String + " " + fn.Description.String
	if fn.Category.Valid {
		query += " " + fn.Category.String
	}

	functions, _, err := s.registry.SearchFunctions(query, fn.Category.String, "", 0, limit*2, 0)
	if err != nil {
		return nil, err
	}

	var results []RecommendationResult
	for _, sf := range functions {
		if sf.ID == functionID {
			continue
		}

		score := s.calculateContentSimilarity(fn, &sf)

		results = append(results, RecommendationResult{
			FunctionID:         sf.ID,
			Author:             sf.Author,
			Name:               sf.Name,
			Title:              sf.Title.String,
			Description:        sf.Description.String,
			Category:           sf.Category.String,
			Tags:               s.parseTags(sf.Tags),
			PopularityScore:    sf.PopularityScore,
			ReliabilityScore:   sf.ReliabilityScore,
			Score:              score,
			RecommendationType: RecommendationTypeSimilar,
		})
	}

	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// convertRecommendationsToResults converts stored recommendations to results
func (s *Service) convertRecommendationsToResults(ctx context.Context, recommendations []FunctionRecommendation) ([]RecommendationResult, error) {
	var results []RecommendationResult

	for _, rec := range recommendations {
		fn, err := s.registry.GetFunctionByID(rec.RecommendedFunctionID)
		if err != nil {
			continue
		}

		results = append(results, RecommendationResult{
			FunctionID:         fn.ID,
			Author:             fn.Author,
			Name:               fn.Name,
			Title:              fn.Title.String,
			Description:        fn.Description.String,
			Category:           fn.Category.String,
			Tags:               s.parseTags(fn.Tags),
			PopularityScore:    fn.PopularityScore,
			ReliabilityScore:   fn.ReliabilityScore,
			Score:              rec.RecommendationScore,
			RecommendationType: rec.RecommendationType,
		})
	}

	return results, nil
}

// deduplicateAndRerank deduplicates recommendations and re-ranks them
func (s *Service) deduplicateAndRerank(recommendations []RecommendationResult, limit int) []RecommendationResult {
	// Deduplicate by function ID, keeping highest score
	seen := make(map[uuid.UUID]RecommendationResult)
	for _, rec := range recommendations {
		if existing, ok := seen[rec.FunctionID]; !ok || rec.Score > existing.Score {
			seen[rec.FunctionID] = rec
		}
	}

	// Convert back to slice
	var unique []RecommendationResult
	for _, rec := range seen {
		unique = append(unique, rec)
	}

	// Sort by score
	sort.Slice(unique, func(i, j int) bool {
		return unique[i].Score > unique[j].Score
	})

	// Limit results
	if len(unique) > limit {
		unique = unique[:limit]
	}

	return unique
}

// calculateContentSimilarity calculates similarity between two functions based on content
func (s *Service) calculateContentSimilarity(a, b *registry.RegistryFunction) float64 {
	var similarity float64

	// Category match (30%)
	if a.Category.Valid && b.Category.Valid && a.Category.String == b.Category.String {
		similarity += 0.3
	}

	// Tag overlap (40%)
	if len(a.Tags) > 0 && len(b.Tags) > 0 {
		aTags := s.parseTags(a.Tags)
		bTags := s.parseTags(b.Tags)

		overlap := 0
		tagSet := make(map[string]bool)
		for _, tag := range aTags {
			tagSet[tag] = true
		}
		for _, tag := range bTags {
			if tagSet[tag] {
				overlap++
			}
		}

		if len(aTags)+len(bTags) > 0 {
			tagSim := 2.0 * float64(overlap) / float64(len(aTags)+len(bTags))
			similarity += 0.4 * tagSim
		}
	}

	// Description similarity using simple word matching (30%)
	if a.Description.Valid && b.Description.Valid && a.Description.String != "" && b.Description.String != "" {
		aWords := s.extractWords(a.Description.String)
		bWords := s.extractWords(b.Description.String)

		overlap := 0
		wordSet := make(map[string]bool)
		for _, word := range aWords {
			wordSet[word] = true
		}
		for _, word := range bWords {
			if wordSet[word] {
				overlap++
			}
		}

		if len(aWords)+len(bWords) > 0 {
			descSim := 2.0 * float64(overlap) / float64(len(aWords)+len(bWords))
			similarity += 0.3 * descSim
		}
	}

	return math.Min(similarity, 1.0)
}

// calculateCategorySimilarity calculates similarity between two categories
func (s *Service) calculateCategorySimilarity(catA, catB string) float64 {
	if catA == "" || catB == "" {
		return 0
	}

	if catA == catB {
		return 1.0
	}

	// Check predefined category similarities
	sim, err := s.repo.GetCategorySimilarity(context.Background(), catA, catB)
	if err == nil && sim != nil {
		return sim.SimilarityScore
	}

	// Simple substring matching as fallback
	if strings.Contains(catA, catB) || strings.Contains(catB, catA) {
		return 0.5
	}

	return 0
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

// getFunctionsInCategory gets all functions in a category
func (s *Service) getFunctionsInCategory(ctx context.Context, category string, limit int) ([]registry.RegistryFunction, error) {
	functions, _, err := s.registry.ListFunctions("", category, nil, "public", limit, 0)
	if err != nil {
		return nil, err
	}
	return functions, nil
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

// extractWords extracts meaningful words from text
func (s *Service) extractWords(text string) []string {
	words := strings.Fields(strings.ToLower(text))

	// Filter out common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
		"being": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "could": true,
		"should": true, "may": true, "might": true, "can": true, "this": true,
		"that": true, "these": true, "those": true, "it": true, "its": true,
	}

	var result []string
	for _, word := range words {
		word = strings.Trim(word, ".,!?;:\"'()[]{}")
		if len(word) > 2 && !stopWords[word] {
			result = append(result, word)
		}
	}

	return result
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

// ---------------------------------------------------------------------------
// FlyEmbed Triple-Vector Methods
// ---------------------------------------------------------------------------

// flyEmbedTripleEmbeddingRequest is the request body for the AI service triple embed endpoint.
type flyEmbedTripleEmbeddingRequest struct {
	FunctionID string  `json:"function_id"`
	Name       string  `json:"name"`
	Title      string  `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Category   string `json:"category,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Manifest   map[string]interface{} `json:"manifest,omitempty"`
	SourceCode string `json:"source_code,omitempty"`
	Runtime    string `json:"runtime,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// flyEmbedTripleEmbeddingResponse is the response from the AI service triple embed endpoint.
type flyEmbedTripleEmbeddingResponse struct {
	FunctionID        string    `json:"function_id"`
	ContractEmbedding []float64 `json:"contract_embedding"`
	SemanticEmbedding []float64 `json:"semantic_embedding"`
	CodeEmbedding     []float64 `json:"code_embedding"`
	ContractText      string    `json:"contract_text"`
	SemanticText      string    `json:"semantic_text"`
	CodeText          string    `json:"code_text"`
	EmbeddingModel    string    `json:"embedding_model"`
	LatencyMs         float64   `json:"latency_ms"`
}

// flyEmbedTripleQueryRequest is the request body for the AI service triple query endpoint.
type flyEmbedTripleQueryRequest struct {
	Query string `json:"query"`
}

// flyEmbedTripleQueryResponse is the response from the AI service triple query endpoint.
type flyEmbedTripleQueryResponse struct {
	Query            string    `json:"query"`
	ContractVector   []float64 `json:"contract_vector"`
	SemanticVector   []float64 `json:"semantic_vector"`
	CodeVector       []float64 `json:"code_vector"`
	Dimensions       int       `json:"dimensions"`
	LatencyMs        float64   `json:"latency_ms"`
}

// EmbedFunctionViaAIService calls the AI service to generate triple embeddings and stores them.
// This is called by the backfill CLI and the publish goroutine.
func (s *Service) EmbedFunctionViaAIService(ctx context.Context, functionID uuid.UUID, name, title, description, category string, tags []string, manifest map[string]interface{}, sourceCode, runtime string, capabilities []string) error {
	reqBody := flyEmbedTripleEmbeddingRequest{
		FunctionID:   functionID.String(),
		Name:         name,
		Title:        title,
		Description:  description,
		Category:     category,
		Tags:         tags,
		Manifest:     manifest,
		SourceCode:   sourceCode,
		Runtime:      runtime,
		Capabilities: capabilities,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/flyembed/embed", s.flyEmbedSvcURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call ai-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ai-service returned status %d", resp.StatusCode)
	}

	var embedResp flyEmbedTripleEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert []float64 to []float32 for pgvector
	contractVec := make([]float32, len(embedResp.ContractEmbedding))
	semanticVec := make([]float32, len(embedResp.SemanticEmbedding))
	codeVec := make([]float32, len(embedResp.CodeEmbedding))
	for i, v := range embedResp.ContractEmbedding {
		contractVec[i] = float32(v)
	}
	for i, v := range embedResp.SemanticEmbedding {
		semanticVec[i] = float32(v)
	}
	for i, v := range embedResp.CodeEmbedding {
		codeVec[i] = float32(v)
	}

	embeddingModel := embedResp.EmbeddingModel
	if embeddingModel == "" {
		embeddingModel = "flyembed-v1"
	}

	triple := &FunctionEmbeddingTriple{
		FunctionID:        functionID,
		ContractEmbedding: contractVec,
		SemanticEmbedding: semanticVec,
		CodeEmbedding:     codeVec,
		ContractText:     &embedResp.ContractText,
		SemanticText:     &embedResp.SemanticText,
		CodeText:         &embedResp.CodeText,
		EmbeddingModel:   embeddingModel,
		EmbeddingVersion: 1,
	}

	return s.repo.UpsertFunctionEmbeddingTriple(ctx, triple)
}

// SearchSimilarByTripleEmbedding uses triple vectors when available, falls back to single vector.
func (s *Service) SearchSimilarByTripleEmbedding(ctx context.Context, query string, limit int, excludeFunctionID *uuid.UUID) ([]RecommendationResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// Check if triples exist
	tripleCount, err := s.repo.CountTripleEmbeddings(ctx)
	if err != nil || tripleCount == 0 {
		// Fall back to single-vector search via AI service query
		logrus.Debug("No triple embeddings found, falling back to single-vector search")
		return s.searchBySingleVector(ctx, query, limit, excludeFunctionID)
	}

	// Generate triple query vectors via AI service
	queryVec, err := s.generateTripleQueryVectors(ctx, query)
	if err != nil {
		logrus.WithError(err).Warn("Failed to generate triple query vectors, falling back to single-vector")
		return s.searchBySingleVector(ctx, query, limit, excludeFunctionID)
	}

	// Run triple-vector search
	weights := DefaultTripleSearchWeights()
	tripleResults, err := s.repo.SearchByTripleVector(ctx, queryVec.contract, queryVec.semantic, queryVec.code, weights, limit, excludeFunctionID)
	if err != nil {
		return nil, fmt.Errorf("triple vector search failed: %w", err)
	}

	// Resolve function details from registry
	return s.tripleResultsToRecommendations(ctx, tripleResults)
}

// tripleQueryVectors holds the three query vectors from the AI service.
type tripleQueryVectors struct {
	contract []float32
	semantic []float32
	code     []float32
}

func (s *Service) generateTripleQueryVectors(ctx context.Context, query string) (*tripleQueryVectors, error) {
	reqBody := flyEmbedTripleQueryRequest{Query: query}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/flyembed/query", s.flyEmbedSvcURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ai-service returned status %d", resp.StatusCode)
	}

	var queryResp flyEmbedTripleQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, err
	}

	contract := make([]float32, len(queryResp.ContractVector))
	semantic := make([]float32, len(queryResp.SemanticVector))
	code := make([]float32, len(queryResp.CodeVector))
	for i, v := range queryResp.ContractVector {
		contract[i] = float32(v)
	}
	for i, v := range queryResp.SemanticVector {
		semantic[i] = float32(v)
	}
	for i, v := range queryResp.CodeVector {
		code[i] = float32(v)
	}

	return &tripleQueryVectors{contract: contract, semantic: semantic, code: code}, nil
}

func (s *Service) tripleResultsToRecommendations(ctx context.Context, results []TripleSearchResult) ([]RecommendationResult, error) {
	if s.registry == nil {
		return nil, nil
	}
	var recs []RecommendationResult
	for _, r := range results {
		fn, err := s.registry.GetFunctionByID(r.FunctionID)
		if err != nil {
			continue
		}
		recs = append(recs, RecommendationResult{
			FunctionID:         fn.ID,
			Author:             fn.Author,
			Name:               fn.Name,
			Title:              fn.Title.String,
			Description:        fn.Description.String,
			Category:           fn.Category.String,
			Tags:               s.parseTags(fn.Tags),
			PopularityScore:    fn.PopularityScore,
			ReliabilityScore:   fn.ReliabilityScore,
			Score:              r.CombinedScore,
			RecommendationType: RecommendationTypeSimilar,
		})
	}
	return recs, nil
}

// searchBySingleVector falls back to single-vector search via the AI service embed endpoint.
func (s *Service) searchBySingleVector(ctx context.Context, query string, limit int, excludeFunctionID *uuid.UUID) ([]RecommendationResult, error) {
	// Generate a single embedding for the query
	embedReq := map[string]interface{}{"text": query}
	body, err := json.Marshal(embedReq)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/embed", s.flyEmbedSvcURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embed endpoint returned status %d", resp.StatusCode)
	}

	var embedResp struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, err
	}

	vec := make([]float32, len(embedResp.Embedding))
	for i, v := range embedResp.Embedding {
		vec[i] = float32(v)
	}

	return s.SearchSimilarByEmbedding(ctx, vec, limit, excludeFunctionID)
}

// UpsertTripleEmbedding stores a triple embedding for a function (called directly, not via AI service).
func (s *Service) UpsertTripleEmbedding(ctx context.Context, functionID uuid.UUID, contract, semantic, code []float32, contractText, semanticText, codeText *string, model string) error {
	if model == "" {
		model = "flyembed-v1"
	}
	return s.repo.UpsertFunctionEmbeddingTriple(ctx, &FunctionEmbeddingTriple{
		FunctionID:        functionID,
		ContractEmbedding: contract,
		SemanticEmbedding: semantic,
		CodeEmbedding:     code,
		ContractText:     contractText,
		SemanticText:     semanticText,
		CodeText:         codeText,
		EmbeddingModel:   model,
		EmbeddingVersion: 1,
	})
}

// SearchByTripleEmbedding generates 3 query embeddings via AI service, runs triple MaxSim.
func (s *Service) SearchByTripleEmbedding(ctx context.Context, query string, weights *TripleSearchWeights, limit int) ([]RecommendationResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	if weights == nil {
		w := DefaultTripleSearchWeights()
		weights = &w
	}

	queryVecs, err := s.generateTripleQueryVectors(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query vectors: %w", err)
	}

	results, err := s.repo.SearchByTripleVector(ctx, queryVecs.contract, queryVecs.semantic, queryVecs.code, *weights, limit, nil)
	if err != nil {
		return nil, err
	}

	return s.tripleResultsToRecommendations(ctx, results)
}

// FindComposableFunctions finds functions whose contract inputs match the target function's outputs.
func (s *Service) FindComposableFunctions(ctx context.Context, functionID uuid.UUID, limit int) ([]RecommendationResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// Get the function's embedding triple
	triple, err := s.repo.GetFunctionEmbeddingTriple(ctx, functionID)
	if err != nil || triple == nil {
		return nil, fmt.Errorf("function has no triple embedding: %w", err)
	}

	// Search by contract vector only (matches output schema to input schema)
	results, err := s.repo.SearchByContractVector(ctx, triple.SemanticEmbedding, limit)
	if err != nil {
		return nil, err
	}

	// Filter out the source function
	filtered := make([]TripleSearchResult, 0, len(results))
	for _, r := range results {
		if r.FunctionID != functionID {
			filtered = append(filtered, r)
		}
		if len(filtered) >= limit {
			break
		}
	}

	return s.tripleResultsToRecommendations(ctx, filtered)
}
