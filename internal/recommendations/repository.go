package recommendations

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Repository handles database operations for recommendations
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new recommendations repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// AutoMigrate runs auto migration for recommendation tables
func (r *Repository) AutoMigrate() error {
	return r.db.AutoMigrate(
		&FunctionCooccurrence{},
		&FunctionSimilarity{},
		&UserFunctionInteraction{},
		&FunctionRecommendation{},
		&SessionFunctionUsage{},
		&CategorySimilarity{},
		&RecommendationFeedback{},
		&FunctionEmbedding{},
	)
}

// CreateFunctionCooccurrence creates or updates a co-occurrence
func (r *Repository) CreateFunctionCooccurrence(ctx context.Context, co *FunctionCooccurrence) error {
	return r.db.WithContext(ctx).Create(co).Error
}

// GetFunctionCooccurrences gets co-occurrences for a function
func (r *Repository) GetFunctionCooccurrences(ctx context.Context, functionID uuid.UUID, limit int) ([]FunctionCooccurrence, error) {
	var cooccurrences []FunctionCooccurrence
	query := r.db.WithContext(ctx).
		Where("function_id_a = ? OR function_id_b = ?", functionID, functionID).
		Order("co_occurrence_count DESC").
		Limit(limit)

	if err := query.Find(&cooccurrences).Error; err != nil {
		return nil, err
	}

	// Ensure consistent ordering
	for i := range cooccurrences {
		if cooccurrences[i].FunctionIDB == functionID {
			cooccurrences[i].FunctionIDA, cooccurrences[i].FunctionIDB = cooccurrences[i].FunctionIDB, cooccurrences[i].FunctionIDA
		}
	}

	return cooccurrences, nil
}

// UpsertFunctionSimilarity creates or updates a similarity score
func (r *Repository) UpsertFunctionSimilarity(ctx context.Context, sim *FunctionSimilarity) error {
	return r.db.WithContext(ctx).
		Where("function_id_a = ? AND function_id_b = ?", sim.FunctionIDA, sim.FunctionIDB).
		Assign(FunctionSimilarity{
			ContentSimilarity:       sim.ContentSimilarity,
			CollaborativeSimilarity: sim.CollaborativeSimilarity,
			CategorySimilarity:      sim.CategorySimilarity,
			CombinedSimilarity:      sim.CombinedSimilarity,
			ComputedAt:              time.Now(),
		}).
		Create(sim).Error
}

// GetSimilaritiesForFunction gets all similarity records for a function
func (r *Repository) GetSimilaritiesForFunction(ctx context.Context, functionID uuid.UUID, limit int) ([]FunctionSimilarity, error) {
	var similarities []FunctionSimilarity
	err := r.db.WithContext(ctx).
		Where("function_id_a = ? OR function_id_b = ?", functionID, functionID).
		Order("combined_similarity DESC").
		Limit(limit).
		Find(&similarities).Error
	return similarities, err
}

// GetSimilarityBetweenFunctions gets similarity between two functions
func (r *Repository) GetSimilarityBetweenFunctions(ctx context.Context, functionIDa, functionIDb uuid.UUID) (*FunctionSimilarity, error) {
	var sim FunctionSimilarity
	err := r.db.WithContext(ctx).
		Where("(function_id_a = ? AND function_id_b = ?) OR (function_id_a = ? AND function_id_b = ?)",
			functionIDa, functionIDb, functionIDb, functionIDa).
		First(&sim).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &sim, err
}

// RecordUserInteraction records a user interaction with a function
func (r *Repository) RecordUserInteraction(ctx context.Context, interaction *UserFunctionInteraction) error {
	interaction.Timestamp = time.Now()
	return r.db.WithContext(ctx).Create(interaction).Error
}

// GetUserInteractionHistory gets a user's interaction history
func (r *Repository) GetUserInteractionHistory(ctx context.Context, userID uuid.UUID, limit int) ([]UserFunctionInteraction, error) {
	var interactions []UserFunctionInteraction
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("timestamp DESC").
		Limit(limit).
		Find(&interactions).Error
	return interactions, err
}

// GetUserExecutedFunctions gets functions a user has executed
func (r *Repository) GetUserExecutedFunctions(ctx context.Context, userID uuid.UUID, limit int) ([]uuid.UUID, error) {
	var functionIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&UserFunctionInteraction{}).
		Where("user_id = ? AND interaction_type = ?", userID, InteractionTypeExecute).
		Distinct("function_id").
		Order("timestamp DESC").
		Limit(limit).
		Pluck("function_id", &functionIDs).Error
	return functionIDs, err
}

// GetUserViewedFunctions gets functions a user has viewed
func (r *Repository) GetUserViewedFunctions(ctx context.Context, userID uuid.UUID, limit int) ([]uuid.UUID, error) {
	var functionIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&UserFunctionInteraction{}).
		Where("user_id = ? AND interaction_type = ?", userID, InteractionTypeView).
		Distinct("function_id").
		Order("timestamp DESC").
		Limit(limit).
		Pluck("function_id", &functionIDs).Error
	return functionIDs, err
}

// UpsertRecommendation creates or updates a recommendation
func (r *Repository) UpsertRecommendation(ctx context.Context, rec *FunctionRecommendation) error {
	return r.db.WithContext(ctx).
		Where("function_id = ? AND recommended_function_id = ? AND recommendation_type = ?",
			rec.FunctionID, rec.RecommendedFunctionID, rec.RecommendationType).
		Assign(FunctionRecommendation{
			RecommendationScore: rec.RecommendationScore,
			RankPosition:        rec.RankPosition,
			ComputedAt:          time.Now(),
			ExpiresAt:           rec.ExpiresAt,
		}).
		Create(rec).Error
}

// GetRecommendationsForFunction gets cached recommendations for a function
func (r *Repository) GetRecommendationsForFunction(ctx context.Context, functionID uuid.UUID, recType RecommendationType, limit int) ([]FunctionRecommendation, error) {
	var recommendations []FunctionRecommendation
	query := r.db.WithContext(ctx).
		Where("function_id = ?", functionID)

	if recType != "" {
		query = query.Where("recommendation_type = ?", recType)
	}

	err := query.
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("recommendation_score DESC").
		Limit(limit).
		Find(&recommendations).Error
	return recommendations, err
}

// DeleteRecommendationsForFunction deletes recommendations for a function
func (r *Repository) DeleteRecommendationsForFunction(ctx context.Context, functionID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("function_id = ?", functionID).
		Delete(&FunctionRecommendation{}).Error
}

// RecordSessionUsage records function usage in a session
func (r *Repository) RecordSessionUsage(ctx context.Context, usage *SessionFunctionUsage) error {
	usage.LastUsedAt = time.Now()

	var existing SessionFunctionUsage
	err := r.db.WithContext(ctx).
		Where("session_id = ? AND function_id = ?", usage.SessionID, usage.FunctionID).
		First(&existing).Error

	if err == nil {
		// Update existing
		return r.db.WithContext(ctx).
			Model(&existing).
			Updates(map[string]interface{}{
				"execution_count": gorm.Expr("execution_count + 1"),
				"last_used_at":    usage.LastUsedAt,
			}).Error
	}

	// Create new
	usage.FirstUsedAt = time.Now()
	return r.db.WithContext(ctx).Create(usage).Error
}

// GetSessionFunctions gets functions used in a session
func (r *Repository) GetSessionFunctions(ctx context.Context, sessionID string) ([]SessionFunctionUsage, error) {
	var usages []SessionFunctionUsage
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("last_used_at DESC").
		Find(&usages).Error
	return usages, err
}

// GetCategorySimilarity gets similarity between two categories
func (r *Repository) GetCategorySimilarity(ctx context.Context, categoryA, categoryB string) (*CategorySimilarity, error) {
	var sim CategorySimilarity
	err := r.db.WithContext(ctx).
		Where("(category_a = ? AND category_b = ?) OR (category_a = ? AND category_b = ?)",
			categoryA, categoryB, categoryB, categoryA).
		First(&sim).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &sim, err
}

// GetRelatedCategories gets related categories for a given category
func (r *Repository) GetRelatedCategories(ctx context.Context, category string, limit int) ([]CategorySimilarity, error) {
	var similarities []CategorySimilarity
	err := r.db.WithContext(ctx).
		Where("category_a = ? OR category_b = ?", category, category).
		Order("similarity_score DESC").
		Limit(limit).
		Find(&similarities).Error
	return similarities, err
}

// RecordFeedback records recommendation feedback
func (r *Repository) RecordFeedback(ctx context.Context, feedback *RecommendationFeedback) error {
	feedback.Timestamp = time.Now()
	return r.db.WithContext(ctx).Create(feedback).Error
}

// GetPopularFunctions gets popular functions (for trending recommendations), ordered by total co-occurrence count.
// When category is set, only functions in that category with visibility = 'public' are returned.
func (r *Repository) GetPopularFunctions(ctx context.Context, category *string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 50
	}

	// Unpivot co-occurrences so each function gets its total count, then order by popularity.
	// Join registry_functions when filtering by category so we only return public functions in that category.
	baseQuery := `
		SELECT fn_id::text
		FROM (
			SELECT function_id_a AS fn_id, co_occurrence_count AS cnt FROM function_co_occurrences
			UNION ALL
			SELECT function_id_b AS fn_id, co_occurrence_count AS cnt FROM function_co_occurrences
		) t
	`
	args := []interface{}{}
	if category != nil && *category != "" {
		baseQuery += ` JOIN registry_functions rf ON rf.id = t.fn_id AND rf.category = ? AND rf.visibility = 'public'`
		args = append(args, *category)
	}
	baseQuery += ` GROUP BY fn_id ORDER BY SUM(cnt) DESC LIMIT ?`
	args = append(args, limit)

	var functionIDs []string
	err := r.db.WithContext(ctx).Raw(baseQuery, args...).Scan(&functionIDs).Error
	return functionIDs, err
}

// GetTrendingFunctions gets trending functions based on recent co-occurrences
func (r *Repository) GetTrendingFunctions(ctx context.Context, limit int) ([]uuid.UUID, error) {
	var functionIDs []uuid.UUID
	oneDayAgo := time.Now().Add(-24 * time.Hour)

	err := r.db.WithContext(ctx).
		Model(&FunctionCooccurrence{}).
		Select("DISTINCT CASE WHEN last_co_occurred_at > ? THEN CASE WHEN function_id_a < function_id_b THEN function_id_a ELSE function_id_b END END", oneDayAgo).
		Where("last_co_occurred_at > ?", oneDayAgo).
		Order("co_occurrence_count DESC").
		Limit(limit).
		Pluck("function_id_a", &functionIDs).Error

	return functionIDs, err
}

// ComputeAndStoreRecommendations computes and stores recommendations for a function
func (r *Repository) ComputeAndStoreRecommendations(ctx context.Context, functionID uuid.UUID, config *RecommendationConfig) error {
	// Get all similarities
	similarities, err := r.GetSimilaritiesForFunction(ctx, functionID, config.MaxRecommendationsPerFunction*2)
	if err != nil {
		return err
	}

	// Get co-occurrences
	cooccurrences, err := r.GetFunctionCooccurrences(ctx, functionID, config.MaxRecommendationsPerFunction*2)
	if err != nil {
		logrus.WithError(err).Warn("failed to get co-occurrences")
	}

	// Get category-based recommendations
	var categoryRecommendations []uuid.UUID
	funcCategory, err := r.getFunctionCategory(ctx, functionID)
	if err == nil && funcCategory != "" {
		categoryFunctions, err := r.getFunctionsInCategory(ctx, funcCategory, config.MaxRecommendationsPerFunction*2)
		if err == nil {
			categoryRecommendations = categoryFunctions
		}
	}

	// Combine recommendations with scores
	type scoredRecommendation struct {
		functionID uuid.UUID
		score      float64
		recType    RecommendationType
	}

	var allRecommendations []scoredRecommendation

	// Add similarity-based recommendations
	for _, sim := range similarities {
		var otherID uuid.UUID
		if sim.FunctionIDA == functionID {
			otherID = sim.FunctionIDB
		} else {
			otherID = sim.FunctionIDA
		}

		if otherID != functionID && sim.CombinedSimilarity >= config.MinSimilarityScore {
			allRecommendations = append(allRecommendations, scoredRecommendation{
				functionID: otherID,
				score:      sim.CombinedSimilarity,
				recType:    RecommendationTypeSimilar,
			})
		}
	}

	// Add collaborative filtering recommendations
	maxCooccurrence := 1
	for _, co := range cooccurrences {
		if co.CooccurrenceCount > maxCooccurrence {
			maxCooccurrence = co.CooccurrenceCount
		}
	}

	for _, co := range cooccurrences {
		var otherID uuid.UUID
		if co.FunctionIDA == functionID {
			otherID = co.FunctionIDB
		} else {
			otherID = co.FunctionIDA
		}

		if otherID != functionID {
			score := float64(co.CooccurrenceCount) / float64(maxCooccurrence)
			allRecommendations = append(allRecommendations, scoredRecommendation{
				functionID: otherID,
				score:      score,
				recType:    RecommendationTypeFrequentlyUsedTogether,
			})
		}
	}

	// Add category-based recommendations
	for _, catFuncID := range categoryRecommendations {
		if catFuncID != functionID {
			// Check if already added
			found := false
			for _, r := range allRecommendations {
				if r.functionID == catFuncID {
					found = true
					break
				}
			}
			if !found {
				allRecommendations = append(allRecommendations, scoredRecommendation{
					functionID: catFuncID,
					score:      0.5, // Default category score
					recType:    RecommendationTypeSameCategory,
				})
			}
		}
	}

	// Deduplicate and keep highest score per function
	recommendationMap := make(map[uuid.UUID]scoredRecommendation)
	for _, rec := range allRecommendations {
		if existing, ok := recommendationMap[rec.functionID]; !ok || rec.score > existing.score {
			recommendationMap[rec.functionID] = rec
		}
	}

	// Sort by score and keep top N
	var finalRecommendations []scoredRecommendation
	for _, rec := range recommendationMap {
		finalRecommendations = append(finalRecommendations, rec)
	}

	sort.Slice(finalRecommendations, func(i, j int) bool {
		return finalRecommendations[i].score > finalRecommendations[j].score
	})

	if len(finalRecommendations) > config.MaxRecommendationsPerFunction {
		finalRecommendations = finalRecommendations[:config.MaxRecommendationsPerFunction]
	}

	// Store recommendations
	expiresAt := time.Now().Add(time.Duration(config.CacheTTLMinutes) * time.Minute)
	for i, rec := range finalRecommendations {
		err := r.UpsertRecommendation(ctx, &FunctionRecommendation{
			ID:                    uuid.New(),
			FunctionID:            functionID,
			RecommendedFunctionID: rec.functionID,
			RecommendationScore:   rec.score,
			RecommendationType:    rec.recType,
			RankPosition:          i + 1,
			ComputedAt:            time.Now(),
			ExpiresAt:             &expiresAt,
		})
		if err != nil {
			logrus.WithError(err).WithField("function_id", functionID).Warn("failed to store recommendation")
		}
	}

	return nil
}

// getFunctionCategory gets the category for a function
func (r *Repository) getFunctionCategory(ctx context.Context, functionID uuid.UUID) (string, error) {
	var category string
	err := r.db.WithContext(ctx).
		Model(&FunctionRecommendation{}).
		Raw("SELECT category FROM registry_functions WHERE id = ?", functionID).
		Scan(&category).Error
	return category, err
}

// getFunctionsInCategory gets functions in a specific category
func (r *Repository) getFunctionsInCategory(ctx context.Context, category string, limit int) ([]uuid.UUID, error) {
	var functionIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&FunctionRecommendation{}).
		Raw("SELECT id FROM registry_functions WHERE category = ? AND visibility = 'public' LIMIT ?", category, limit).
		Scan(&functionIDs).Error
	return functionIDs, err
}

// BatchComputeRecommendations computes recommendations for multiple functions
func (r *Repository) BatchComputeRecommendations(ctx context.Context, functionIDs []uuid.UUID, config *RecommendationConfig) error {
	for _, functionID := range functionIDs {
		if err := r.ComputeAndStoreRecommendations(ctx, functionID, config); err != nil {
			logrus.WithError(err).WithField("function_id", functionID).Warn("failed to compute recommendations")
		}
	}
	return nil
}

// UpsertFunctionEmbedding creates or updates the embedding for a function (pgvector vector(1536)).
func (r *Repository) UpsertFunctionEmbedding(ctx context.Context, e *FunctionEmbedding) error {
	e.ComputedAt = time.Now()
	return r.db.WithContext(ctx).
		Where("function_id = ?", e.FunctionID).
		Assign(FunctionEmbedding{
			Embedding:      e.Embedding,
			EmbeddedText:   e.EmbeddedText,
			EmbeddingModel: e.EmbeddingModel,
			ComputedAt:     e.ComputedAt,
		}).
		FirstOrCreate(e).Error
}

// GetFunctionEmbeddingByFunctionID returns the embedding for a function, or nil if not found.
func (r *Repository) GetFunctionEmbeddingByFunctionID(ctx context.Context, functionID uuid.UUID) (*FunctionEmbedding, error) {
	var e FunctionEmbedding
	err := r.db.WithContext(ctx).Where("function_id = ?", functionID).First(&e).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// FunctionEmbeddingWithDistance is a function embedding plus its cosine distance to the query vector (0 = identical, 2 = opposite).
type FunctionEmbeddingWithDistance struct {
	FunctionEmbedding
	Distance float64 `json:"distance" gorm:"column:distance"`
}

// SearchFunctionEmbeddingsByVector returns function embeddings ordered by cosine similarity (<=>), with distance.
// excludeFunctionID optionally excludes one function from results (e.g. the query function).
func (r *Repository) SearchFunctionEmbeddingsByVector(ctx context.Context, queryVector []float32, limit int, excludeFunctionID *uuid.UUID) ([]*FunctionEmbeddingWithDistance, error) {
	vectorStr := "["
	for i, v := range queryVector {
		if i > 0 {
			vectorStr += ","
		}
		vectorStr += fmt.Sprintf("%.6f", v)
	}
	vectorStr += "]"

	var args []interface{}
	sql := `SELECT id, function_id, embedding, embedded_text, embedding_model, computed_at,
		(embedding <=> ?) AS distance
		FROM function_embeddings
		WHERE embedding IS NOT NULL`
	if excludeFunctionID != nil {
		sql += ` AND function_id != ?`
		args = append(args, *excludeFunctionID)
	}
	sql += ` ORDER BY embedding <=> ? LIMIT ?`
	args = append(args, vectorStr, vectorStr, limit)

	var out []*FunctionEmbeddingWithDistance
	err := r.db.WithContext(ctx).Raw(sql, args...).Scan(&out).Error
	return out, err
}
