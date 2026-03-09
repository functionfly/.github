package discovery

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Scorer allows discovery scoring to be reused or replaced.
type Scorer interface {
	Score(ctx context.Context, candidate OpportunityCandidate) (demand float64, quality float64, complexity int, err error)
}

// Service orchestrates discovery source scanning and opportunity persistence.
type Service struct {
	db      *gorm.DB
	sources []Source
	scorer  Scorer
}

func NewService(db *gorm.DB, sources ...Source) *Service {
	return &Service{db: db, sources: sources, scorer: DefaultScorer{}}
}

func NewServiceWithScorer(db *gorm.DB, scorer Scorer, sources ...Source) *Service {
	if scorer == nil {
		scorer = DefaultScorer{}
	}
	return &Service{db: db, sources: sources, scorer: scorer}
}

func (s *Service) AutoMigrate(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(&Opportunity{})
}

func (s *Service) ScanAll(ctx context.Context) ([]DiscoveryBatch, error) {
	results := make([]DiscoveryBatch, 0, len(s.sources))
	for _, source := range s.sources {
		batch, err := s.ScanSource(ctx, source)
		if err != nil {
			return results, err
		}
		results = append(results, batch)
	}
	return results, nil
}

func (s *Service) ScanSource(ctx context.Context, source Source) (DiscoveryBatch, error) {
	started := time.Now()
	if source == nil {
		return DiscoveryBatch{}, errors.New("source is required")
	}
	candidates, err := source.Scan(ctx)
	if err != nil {
		return DiscoveryBatch{}, fmt.Errorf("scan %s: %w", source.Name(), err)
	}

	batch := DiscoveryBatch{Source: source.Name(), Discovered: len(candidates)}
	for _, candidate := range candidates {
		candidate.Normalize()
		persisted, created, err := s.upsertCandidate(ctx, candidate)
		if err != nil {
			return batch, err
		}
		if created {
			batch.Persisted++
		} else if persisted != nil {
			batch.Deduplicated++
		}
	}
	batch.Duration = time.Since(started)
	return batch, nil
}

func (s *Service) ListQualified(ctx context.Context, limit int) ([]Opportunity, error) {
	if limit <= 0 {
		limit = DefaultDiscoveryBatchLimit
	}
	var opportunities []Opportunity
	err := s.db.WithContext(ctx).
		Where("status IN ?", []string{OpportunityStatusQualified, OpportunityStatusNeedsReview}).
		Order("quality_score DESC, demand_score DESC, created_at ASC").
		Limit(limit).
		Find(&opportunities).Error
	return opportunities, err
}

func (s *Service) MarkGenerated(ctx context.Context, sourceID string, generatedFunctionID string) error {
	return s.db.WithContext(ctx).Model(&Opportunity{}).
		Where("id = ?", sourceID).
		Updates(map[string]any{
			"status":            OpportunityStatusGenerated,
			"generated_func_id": generatedFunctionID,
			"updated_at":        time.Now().UTC(),
		}).Error
}

func (s *Service) ApplyReviewDecision(ctx context.Context, opportunityID string, decision ReviewDecision) error {
	updates := map[string]any{
		"updated_at": time.Now().UTC(),
	}
	if decision.Approved {
		updates["review_status"] = ReviewStatusApproved
		updates["status"] = OpportunityStatusQualified
		updates["validated"] = true
	} else {
		updates["review_status"] = ReviewStatusRejected
		updates["status"] = OpportunityStatusRejected
	}
	if reason := strings.TrimSpace(decision.Reason); reason != "" {
		updates["review_reason"] = reason
	}
	if actor := strings.TrimSpace(decision.Actor); actor != "" {
		updates["metadata"] = gorm.Expr("COALESCE(metadata, '{}'::jsonb) || jsonb_build_object('review_actor', ?)", actor)
	}
	return s.db.WithContext(ctx).Model(&Opportunity{}).Where("id = ?", opportunityID).Updates(updates).Error
}

func (s *Service) upsertCandidate(ctx context.Context, candidate OpportunityCandidate) (*Opportunity, bool, error) {
	demand, quality, complexity, err := s.scorer.Score(ctx, candidate)
	if err != nil {
		return nil, false, fmt.Errorf("score candidate %s/%s: %w", candidate.Source, candidate.SourceID, err)
	}

	reviewStatus, reviewReason, reviewRequestedAt := evaluateReviewNeed(demand, quality, complexity)
	status := OpportunityStatusQualified
	validated := quality >= 70
	if !validated {
		status = OpportunityStatusRejected
	}
	if reviewStatus == ReviewStatusPending {
		status = OpportunityStatusNeedsReview
	}

	record := Opportunity{
		Source:            candidate.Source,
		SourceID:          candidate.SourceID,
		Title:             candidate.Title,
		Description:       candidate.Description,
		Category:          candidate.Category,
		Tags:              uniqueStrings(candidate.Tags),
		DemandScore:       roundScore(demand),
		Complexity:        complexity,
		Validated:         validated,
		Status:            status,
		QualityScore:      roundScore(quality),
		ReviewStatus:      reviewStatus,
		ReviewReason:      reviewReason,
		ReviewRequestedAt: reviewRequestedAt,
		Metadata:          candidate.Metadata,
	}

	result := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "source"}, {Name: "source_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "description", "category", "tags", "demand_score", "complexity", "validated", "status", "quality_score", "review_status", "review_reason", "review_requested_at", "metadata", "updated_at"}),
	}).Create(&record)
	if result.Error != nil {
		return nil, false, result.Error
	}
	var existing Opportunity
	if err := s.db.WithContext(ctx).Where("source = ? AND source_id = ?", candidate.Source, candidate.SourceID).First(&existing).Error; err != nil {
		return nil, false, err
	}
	created := existing.CreatedAt.Equal(existing.UpdatedAt)
	return &existing, created, nil
}

// DefaultScorer implements a simple heuristic scorer that can be refined later.
type DefaultScorer struct{}

func (DefaultScorer) Score(_ context.Context, candidate OpportunityCandidate) (float64, float64, int, error) {
	text := strings.ToLower(strings.Join([]string{candidate.Title, candidate.Description, strings.Join(candidate.Tags, " ")}, " "))
	demand := candidate.DemandSignal
	quality := 45.0 + demand*0.35
	complexity := candidate.ComplexitySignal

	weights := map[string]float64{
		"api": 8, "automation": 7, "convert": 5, "json": 4, "csv": 4, "parser": 6,
		"integration": 9, "error": 3, "workflow": 6, "oauth": 9, "webhook": 8,
	}
	for token, bonus := range weights {
		if strings.Contains(text, token) {
			demand += bonus
			quality += bonus * 0.7
			if bonus >= 8 {
				complexity++
			}
		}
	}

	if strings.Contains(text, "urgent") || strings.Contains(text, "popular") || strings.Contains(text, "high demand") {
		demand += 10
	}
	if strings.Contains(text, "simple") || strings.Contains(text, "one step") {
		complexity--
	}

	if quality > 100 {
		quality = 100
	}
	if demand > 100 {
		demand = 100
	}
	if complexity < 1 {
		complexity = 1
	}
	if complexity > 10 {
		complexity = 10
	}
	return demand, quality, complexity, nil
}

func evaluateReviewNeed(demand, quality float64, complexity int) (string, *string, *time.Time) {
	if quality < 70 {
		return ReviewStatusNotRequired, nil, nil
	}
	if complexity >= 8 || demand >= 85 {
		reason := "high-impact opportunity requires manual review"
		now := time.Now().UTC()
		return ReviewStatusPending, &reason, &now
	}
	return ReviewStatusNotRequired, nil, nil
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		key := strings.ToLower(normalized)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

func roundScore(value float64) float64 {
	return math.Round(value*100) / 100
}
