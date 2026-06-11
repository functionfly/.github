package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateFunction creates a new function in the registry
func (r *RegistryRepository) CreateFunction(fn *RegistryFunction) error {
	if fn.ID == uuid.Nil {
		fn.ID = uuid.New()
	}
	now := time.Now()
	if fn.CreatedAt.IsZero() {
		fn.CreatedAt = now
	}
	if fn.UpdatedAt.IsZero() {
		fn.UpdatedAt = now
	}

	if err := r.db.Create(fn).Error; err != nil {
		return fmt.Errorf("failed to create function: %w", err)
	}

	// Invalidate any related cache entries (though new functions won't have cache entries yet)
	if r.cache != nil {
		go func() {
			if err := r.cache.InvalidateSearchResults(context.Background()); err != nil {
				fmt.Printf("Failed to invalidate search cache after function creation: %v\n", err)
			}
		}()
	}

	return nil
}

// GetFunctionByID retrieves a function by ID with preloaded relationships.
// Uses batch queries instead of GORM's chained Preloads to avoid the N+1
// query problem (each Preload fires a separate sequential query, compounding
// latency with deeply nested relations like Versions → Signatures, Approvals → Comments).
func (r *RegistryRepository) GetFunctionByID(id uuid.UUID) (*RegistryFunction, error) {
	var fn RegistryFunction
	if err := r.db.First(&fn, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get function by ID: %w", err)
	}

	if err := r.preloadFunctionRelations(&fn); err != nil {
		return nil, err
	}

	return &fn, nil
}

// preloadFunctionRelations batch-loads all related data for a function to avoid N+1 queries.
func (r *RegistryRepository) preloadFunctionRelations(fn *RegistryFunction) error {
	// Load versions.
	var versions []RegistryFunctionVersion
	if err := r.db.Where("function_id = ?", fn.ID).Order("published_at DESC").Find(&versions).Error; err != nil {
		return fmt.Errorf("failed to load versions: %w", err)
	}
	if len(versions) == 0 {
		fn.Versions = []RegistryFunctionVersion{}
		return nil
	}

	// Collect version IDs for batch lookups.
	versionIDs := make([]uuid.UUID, len(versions))
	versionMap := make(map[uuid.UUID]*RegistryFunctionVersion, len(versions))
	for i, v := range versions {
		v := v // local copy for closure
		versionIDs[i] = v.ID
		versionMap[v.ID] = &versions[i]
	}

	// Batch load signatures.
	var signatures []RegistryFunctionSignature
	if err := r.db.Where("function_version_id IN ?", versionIDs).Find(&signatures).Error; err != nil {
		return fmt.Errorf("failed to load signatures: %w", err)
	}
	// Group signatures by version.
	sigMap := make(map[uuid.UUID][]RegistryFunctionSignature, len(versionIDs))
	for _, s := range signatures {
		s := s // local copy
		sigMap[s.FunctionVersionID] = append(sigMap[s.FunctionVersionID], s)
	}

	// Batch load malware scans.
	var malwareScans []RegistryFunctionMalwareScan
	if err := r.db.Where("function_version_id IN ?", versionIDs).Find(&malwareScans).Error; err != nil {
		return fmt.Errorf("failed to load malware scans: %w", err)
	}
	scanMap := make(map[uuid.UUID][]RegistryFunctionMalwareScan, len(versionIDs))
	for _, ms := range malwareScans {
		ms := ms
		scanMap[ms.FunctionVersionID] = append(scanMap[ms.FunctionVersionID], ms)
	}

	// Batch load approvals.
	var approvals []RegistryFunctionApproval
	if err := r.db.Where("function_version_id IN ?", versionIDs).Find(&approvals).Error; err != nil {
		return fmt.Errorf("failed to load approvals: %w", err)
	}
	approvalMap := make(map[uuid.UUID][]RegistryFunctionApproval, len(versionIDs))
	approvalIDs := make([]uuid.UUID, 0, len(approvals))
	approvalSet := make(map[uuid.UUID]struct{}, len(approvals))
	for _, a := range approvals {
		a := a
		approvalMap[a.FunctionVersionID] = append(approvalMap[a.FunctionVersionID], a)
		if _, ok := approvalSet[a.ID]; !ok {
			approvalIDs = append(approvalIDs, a.ID)
			approvalSet[a.ID] = struct{}{}
		}
	}

	// Batch load approval comments.
	var comments []RegistryFunctionApprovalComment
	if len(approvalIDs) > 0 {
		if err := r.db.Where("approval_id IN ?", approvalIDs).Find(&comments).Error; err != nil {
			return fmt.Errorf("failed to load approval comments: %w", err)
		}
	}
	commentMap := make(map[uuid.UUID][]RegistryFunctionApprovalComment, len(approvalIDs))
	for _, c := range comments {
		c := c
		commentMap[c.ApprovalID] = append(commentMap[c.ApprovalID], c)
	}

	// Batch load verification statuses.
	var verificationStatuses []RegistryFunctionVerificationStatus
	if err := r.db.Where("function_version_id IN ?", versionIDs).Find(&verificationStatuses).Error; err != nil {
		return fmt.Errorf("failed to load verification statuses: %w", err)
	}
	vsMap := make(map[uuid.UUID]*RegistryFunctionVerificationStatus, len(versionIDs))
	for i := range verificationStatuses {
		vsMap[verificationStatuses[i].FunctionVersionID] = &verificationStatuses[i]
	}

	// Wire up all relations.
	for id, version := range versionMap {
		version.Signatures = sigMap[id]
		version.MalwareScans = scanMap[id]
		version.Approvals = approvalMap[id]
		version.VerificationStatus = vsMap[id]
		for i := range version.Approvals {
			version.Approvals[i].Comments = commentMap[version.Approvals[i].ID]
		}
		versionMap[id] = version
	}

	fn.Versions = versions

	// Load rating if exists.
	var rating RegistryFunctionRating
	if err := r.db.Where("function_id = ?", fn.ID).First(&rating).Error; err == nil {
		fn.Rating = &rating
	}

	return nil
}

// GetFunctionByIDMinimal retrieves a function by ID without preloaded relationships (for performance)
func (r *RegistryRepository) GetFunctionByIDMinimal(id uuid.UUID) (*RegistryFunction, error) {
	var fn RegistryFunction
	if err := r.db.First(&fn, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get function by ID: %w", err)
	}

	return &fn, nil
}

// GetFunctionByAuthorName retrieves a function by author and name with preloaded relationships.
// Uses batch queries instead of GORM's chained Preloads to avoid the N+1 query problem.
func (r *RegistryRepository) GetFunctionByAuthorName(author, name string) (*RegistryFunction, error) {
	// Try cache first if available
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionInfo(author, name)
		var fn RegistryFunction
		if err := r.cache.GetJSON(context.Background(), cacheKey, &fn); err == nil {
			return &fn, nil
		}
		// Cache miss - continue to database
	}

	var fn RegistryFunction
	if err := r.db.Where("author = ? AND name = ?", author, name).First(&fn).Error; err != nil {
		return nil, fmt.Errorf("failed to get function by author/name: %w", err)
	}

	if err := r.preloadFunctionRelations(&fn); err != nil {
		return nil, err
	}

	// Cache the result if cache is available
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionInfo(author, name)
		go func() {
			if err := r.cache.SetJSON(context.Background(), cacheKey, fn); err != nil {
				// Log error but don't fail the operation
				fmt.Printf("Failed to cache function info: %v\n", err)
			}
		}()
	}

	return &fn, nil
}

// GetFunctionByAuthorNameMinimal retrieves a function by author and name without preloaded relationships (for performance)
func (r *RegistryRepository) GetFunctionByAuthorNameMinimal(author, name string) (*RegistryFunction, error) {
	// Try cache first if available
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionInfo(author, name)
		var fn RegistryFunction
		if err := r.cache.GetJSON(context.Background(), cacheKey, &fn); err == nil {
			return &fn, nil
		}
		// Cache miss - continue to database
	}

	var fn RegistryFunction
	if err := r.db.Where("author = ? AND name = ?", author, name).First(&fn).Error; err != nil {
		return nil, fmt.Errorf("failed to get function by author/name: %w", err)
	}

	// Cache the result if cache is available
	if r.cache != nil && r.keyGen != nil {
		cacheKey := r.keyGen.FunctionInfo(author, name)
		go func() {
			if err := r.cache.SetJSON(context.Background(), cacheKey, fn); err != nil {
				// Log error but don't fail the operation
				fmt.Printf("Failed to cache function info: %v\n", err)
			}
		}()
	}

	return &fn, nil
}

// UpdateFunctionLatestVersion updates the latest version pointer
func (r *RegistryRepository) UpdateFunctionLatestVersion(id uuid.UUID, version string) error {
	if err := r.db.Model(&RegistryFunction{}).Where("id = ?", id).Updates(map[string]interface{}{
		"latest_version": version,
		"updated_at":     time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to update latest version: %w", err)
	}

	// Invalidate cache for this function
	if r.cache != nil {
		go func() {
			if err := r.cache.InvalidateFunction(context.Background(), id.String()); err != nil {
				fmt.Printf("Failed to invalidate function cache after version update: %v\n", err)
			}
		}()
	}

	return nil
}
