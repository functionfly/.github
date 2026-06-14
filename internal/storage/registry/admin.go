package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ListFunctionsForAdmin lists all registry functions for admin (optional visibility/category/search; empty visibility = all).
// Does not use cache.
func (r *RegistryRepository) ListFunctionsForAdmin(visibility, category, search string, limit, offset int) ([]RegistryFunction, int, error) {
	query := r.db.Model(&RegistryFunction{})

	// If visibility is specified and not "all", filter by it
	// If visibility is empty or "all", show all visibilities (no filter)
	if visibility != "" && visibility != "all" {
		query = query.Where("visibility = ?", visibility)
	}
	// If category is specified and not "all", filter by it
	// If category is empty or "all", show all categories (no filter)
	if category != "" && category != "all" {
		query = query.Where("category = ?", category)
	}
	if search != "" {
		q := "%" + search + "%"
		query = query.Where("name ILIKE ? OR title ILIKE ? OR description ILIKE ? OR author ILIKE ?", q, q, q, q)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count functions: %w", err)
	}

	var functions []RegistryFunction
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&functions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list functions: %w", err)
	}
	return functions, int(total), nil
}

// GetRegistryStats returns aggregate stats for admin (total and by visibility).
func (r *RegistryRepository) GetRegistryStats() (total int64, byVisibility map[string]int64, err error) {
	if err := r.db.Model(&RegistryFunction{}).Count(&total).Error; err != nil {
		return 0, nil, fmt.Errorf("failed to count total: %w", err)
	}
	byVisibility = make(map[string]int64)
	for _, v := range []string{"public", "private", "unlisted"} {
		var c int64
		if err := r.db.Model(&RegistryFunction{}).Where("visibility = ?", v).Count(&c).Error; err != nil {
			continue
		}
		byVisibility[v] = c
	}
	return total, byVisibility, nil
}

// UpdateRegistryFunction updates registry function fields (visibility, price_per_call, title, description, category, tags).
func (r *RegistryRepository) UpdateRegistryFunction(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*RegistryFunction, error) {
	allowed := make(map[string]interface{})
	for k, v := range updates {
		switch k {
		case "visibility", "price_per_call", "title", "description", "category", "tags":
			allowed[k] = v
		}
	}
	if len(allowed) == 0 {
		return r.GetFunctionByID(ctx, id)
	}
	if err := r.db.Model(&RegistryFunction{}).Where("id = ?", id).Updates(allowed).Error; err != nil {
		return nil, fmt.Errorf("failed to update function: %w", err)
	}
	return r.GetFunctionByID(ctx, id)
}

// AdminStatsResult holds the extended admin stats returned by GetAdminRegistryStats.
type AdminStatsResult struct {
	TotalFunctions   int64            `json:"total_functions"`
	PublicFunctions  int64            `json:"public_functions"`
	PrivateFunctions int64            `json:"private_functions"`
	UnlistedFunctions int64           `json:"unlisted_functions"`
	FlaggedFunctions int64            `json:"flagged_functions"`
	TotalCalls       int64            `json:"total_calls"`
	TotalRevenueUSD  float64          `json:"total_revenue_usd"`
	AvgRating        float64          `json:"avg_rating"`
	ByVisibility     map[string]int64 `json:"by_visibility"`
}

// GetAdminRegistryStats returns aggregate stats for admin: counts by visibility, flagged count, total calls, revenue, avg rating.
func (r *RegistryRepository) GetAdminRegistryStats() (*AdminStatsResult, error) {
	result := &AdminStatsResult{
		ByVisibility: make(map[string]int64),
	}

	// Total and by-visibility counts
	if err := r.db.Model(&RegistryFunction{}).Count(&result.TotalFunctions).Error; err != nil {
		return nil, fmt.Errorf("failed to count total functions: %w", err)
	}
	for _, v := range []string{"public", "private", "unlisted"} {
		var c int64
		if err := r.db.Model(&RegistryFunction{}).Where("visibility = ?", v).Count(&c).Error; err != nil {
			continue
		}
		result.ByVisibility[v] = c
	}
	result.PublicFunctions = result.ByVisibility["public"]
	result.PrivateFunctions = result.ByVisibility["private"]
	result.UnlistedFunctions = result.ByVisibility["unlisted"]

	// Flagged count — functions with a manual_review_queue entry in a non-terminal status
	var flagged int64
	if err := r.db.Model(&ManualReviewQueue{}).
		Where("status IN ?", []string{"pending", "in_review", "escalated"}).
		Distinct("function_id").
		Count(&flagged).Error; err == nil {
		result.FlaggedFunctions = flagged
	}

	// Total calls (all-time)
	if err := r.db.Model(&RegistryFunctionExecution{}).Count(&result.TotalCalls).Error; err != nil {
		logrus.WithError(err).Warn("GetAdminRegistryStats: failed to count total calls")
	}

	// Total revenue: sum of price_per_call * execution count per function
	var revenue struct {
		Total float64
	}
	if err := r.db.Raw(`
		SELECT COALESCE(SUM(price_per_call * exec_count), 0) as total
		FROM (
			SELECT f.price_per_call, COUNT(e.id) as exec_count
			FROM registry_functions f
			LEFT JOIN registry_function_executions e ON e.function_id = f.id
			WHERE f.price_per_call > 0
			GROUP BY f.id, f.price_per_call
		) sub
	`).Scan(&revenue).Error; err == nil {
		result.TotalRevenueUSD = revenue.Total
	}

	// Average overall_score across all rated functions
	var rating struct {
		Avg float64
	}
	if err := r.db.Raw(`
		SELECT COALESCE(AVG(overall_score), 0) as avg
		FROM registry_function_ratings
		WHERE overall_score > 0
	`).Scan(&rating).Error; err == nil {
		result.AvgRating = rating.Avg
	}

	return result, nil
}

// FlagFunctionFlags represents the flaggable reasons for a function.
type FlagFunctionFlags string

const (
	FlagReasonSpam           FlagFunctionFlags = "spam"
	FlagReasonMalware        FlagFunctionFlags = "malware"
	FlagReasonIPInfringement FlagFunctionFlags = "ip_infringement"
	FlagReasonAbuse          FlagFunctionFlags = "abuse"
	FlagReasonViolation      FlagFunctionFlags = "policy_violation"
)

// FlagFunction creates or updates a manual review queue entry to flag a function for review.
func (r *RegistryRepository) FlagFunction(ctx context.Context, functionID uuid.UUID, reason FlagFunctionFlags, reviewerID uuid.UUID, notes string) error {
	existing, err := r.GetManualReviewQueueByFunctionVersion(ctx, functionID)
	if err != nil && err != ErrReviewNotFound {
		return fmt.Errorf("failed to check existing review: %w", err)
	}

	if existing != nil {
		// Update existing entry instead of creating duplicate
		if err := r.UpdateManualReviewQueue(ctx, existing.ID, map[string]interface{}{
			"status":       "pending",
			"priority":     "high",
			"review_notes": notes,
			"review_type":  string(reason),
		}); err != nil {
			return fmt.Errorf("failed to update existing review queue entry: %w", err)
		}
		return nil
	}

	// Get latest version for the function
	fn, err := r.GetFunctionByID(ctx, functionID)
	if err != nil {
		return fmt.Errorf("failed to get function: %w", err)
	}

	// Get the function version ID for the latest version
	var versionID uuid.UUID
	if fn.LatestVersion.Valid && fn.LatestVersion.String != "" {
		var version RegistryFunctionVersion
		if err := r.db.Where("function_id = ? AND version = ?", functionID, fn.LatestVersion.String).First(&version).Error; err == nil {
			versionID = version.ID
		}
	}
	if versionID == uuid.Nil {
		// Fall back to latest version by published_at
		var version RegistryFunctionVersion
		if err := r.db.Where("function_id = ?", functionID).Order("published_at DESC").First(&version).Error; err == nil {
			versionID = version.ID
		}
	}

	review := &ManualReviewQueue{
		ID:                 uuid.New(),
		FunctionID:         functionID,
		FunctionVersionID:  versionID,
		Status:             "pending",
		Priority:           "high",
		ReviewType:         string(reason),
		ReviewNotes:        notes,
		ReviewComments:     "",
		AssignedTo:         nil,
		AutoApproveIfNoResponseDays: 3,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	if reviewerID != uuid.Nil {
		review.AssignedTo = &reviewerID
	}

	return r.CreateManualReviewQueue(ctx, review)
}

// DeactivateFunctionVersion sets a version's IsActive flag to false.
func (r *RegistryRepository) DeactivateFunctionVersion(versionID uuid.UUID) error {
	if err := r.db.Model(&RegistryFunctionVersion{}).
		Where("id = ?", versionID).
		Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate version: %w", err)
	}
	return nil
}

// GetVersionByID retrieves a specific version by its primary key.
func (r *RegistryRepository) GetVersionByID(versionID uuid.UUID) (*RegistryFunctionVersion, error) {
	var v RegistryFunctionVersion
	if err := r.db.First(&v, versionID).Error; err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}
	return &v, nil
}
