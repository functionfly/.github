package recommendations

import (
	"context"
	"math"
	"strings"

	"github.com/functionfly/functionfly/internal/storage/registry"
)

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
