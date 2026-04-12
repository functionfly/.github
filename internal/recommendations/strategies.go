package recommendations

import (
	"context"
	"sort"
	"strings"

	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

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

// getFunctionsInCategory gets all functions in a category
func (s *Service) getFunctionsInCategory(ctx context.Context, category string, limit int) ([]registry.RegistryFunction, error) {
	functions, _, err := s.registry.ListFunctions("", category, nil, "public", limit, 0)
	if err != nil {
		return nil, err
	}
	return functions, nil
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
