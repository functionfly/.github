package storage

import (
	"fmt"
)

// SearchFunctions performs an optimized search with JOINs for better performance
func (r *RegistryRepository) SearchFunctions(query string, limit, offset int, includePrivate bool) ([]RegistryFunction, error) {
	var functions []RegistryFunction

	baseQuery := r.db.Joins("LEFT JOIN registry_function_ratings ON registry_functions.id = registry_function_ratings.function_id").
		Preload("Rating").
		Where("registry_functions.visibility = 'public'")

	if !includePrivate {
		baseQuery = baseQuery.Where("registry_functions.visibility = 'public'")
	}

	// Add full-text search if query provided
	if query != "" {
		baseQuery = baseQuery.Where("registry_functions.name ILIKE ? OR registry_functions.description ILIKE ? OR registry_functions.author ILIKE ?",
			"%"+query+"%", "%"+query+"%", "%"+query+"%")
	}

	if err := baseQuery.Order("COALESCE(registry_function_ratings.trust_score, 0) DESC, registry_functions.popularity_score DESC").
		Limit(limit).Offset(offset).
		Find(&functions).Error; err != nil {
		return nil, fmt.Errorf("failed to search functions: %w", err)
	}

	return functions, nil
}

// ListFunctions lists functions with filters
func (r *RegistryRepository) ListFunctions(author, category string, tags []string, visibility string, limit, offset int) ([]RegistryFunction, int, error) {
	query := r.db.Model(&RegistryFunction{})

	if author != "" {
		query = query.Where("author = ?", author)
	}

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if visibility == "" {
		visibility = "public"
	}
	query = query.Where("visibility = ?", visibility)

	// Count query
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count functions: %w", err)
	}

	// Add ordering and pagination
	var functions []RegistryFunction
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&functions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list functions: %w", err)
	}

	return functions, int(total), nil
}

// SearchFunctionsAdvanced searches functions by text query
func (r *RegistryRepository) SearchFunctionsAdvanced(query string, category, runtime string, minRating float64, limit, offset int) ([]RegistryFunction, int, error) {
	dbQuery := r.db.Model(&RegistryFunction{}).
		Where("visibility = ?", "public").
		Where("name ILIKE ? OR title ILIKE ? OR description ILIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%")

	if category != "" {
		dbQuery = dbQuery.Where("category = ?", category)
	}

	if minRating > 0 {
		dbQuery = dbQuery.Where("reliability_score >= ?", minRating)
	}

	// Count query
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Add ordering and pagination - sort by trust_score first, then popularity
	var functions []RegistryFunction
	if err := dbQuery.
		Order("trust_score DESC, popularity_score DESC, reliability_score DESC").
		Limit(limit).Offset(offset).Find(&functions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search functions: %w", err)
	}

	return functions, int(total), nil
}

// ListFunctionsByTrustScore lists functions ordered by trust score
func (r *RegistryRepository) ListFunctionsByTrustScore(category string, tags []string, visibility string, limit, offset int) ([]RegistryFunction, int, error) {
	dbQuery := r.db.Model(&RegistryFunction{})

	if category != "" {
		dbQuery = dbQuery.Where("category = ?", category)
	}

	if visibility == "" {
		visibility = "public"
	}
	dbQuery = dbQuery.Where("visibility = ?", visibility)

	// Count query
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count functions: %w", err)
	}

	// Add ordering by trust_score and pagination
	var functions []RegistryFunction
	if err := dbQuery.Order("trust_score DESC, popularity_score DESC").
		Limit(limit).Offset(offset).Find(&functions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list functions: %w", err)
	}

	return functions, int(total), nil
}
