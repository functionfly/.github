package registry

import (
	"fmt"

	"github.com/google/uuid"
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
func (r *RegistryRepository) UpdateRegistryFunction(id uuid.UUID, updates map[string]interface{}) (*RegistryFunction, error) {
	allowed := make(map[string]interface{})
	for k, v := range updates {
		switch k {
		case "visibility", "price_per_call", "title", "description", "category", "tags":
			allowed[k] = v
		}
	}
	if len(allowed) == 0 {
		return r.GetFunctionByID(id)
	}
	if err := r.db.Model(&RegistryFunction{}).Where("id = ?", id).Updates(allowed).Error; err != nil {
		return nil, fmt.Errorf("failed to update function: %w", err)
	}
	return r.GetFunctionByID(id)
}
