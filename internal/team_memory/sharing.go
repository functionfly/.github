package team_memory

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ============================================
// Cross-Team Memory Sharing
// ============================================

// ShareType defines how a memory is shared
type ShareType string

const (
	ShareTypeReference ShareType = "reference" // Reference to original (updates sync)
	ShareTypeCopy      ShareType = "copy"      // Independent copy (no sync)
	ShareTypeFork      ShareType = "fork"      // Copy that tracks ancestry but independent
)

// Permission defines access level
type Permission string

const (
	PermissionRead  Permission = "read"  // Can view and use
	PermissionWrite Permission = "write" // Can view and suggest updates
	PermissionAdmin Permission = "admin" // Full control including revoke
)

// ShareStatus defines the status of a share
type ShareStatus string

const (
	ShareStatusPending  ShareStatus = "pending"
	ShareStatusAccepted ShareStatus = "accepted"
	ShareStatusRejected ShareStatus = "rejected"
	ShareStatusRevoked  ShareStatus = "revoked"
)

// SharingPolicy defines the policy for auto-accepting shares
type SharingPolicy struct {
	AutoAcceptSameTenant bool     // Auto-accept shares between teams in the same tenant
	AutoAcceptSameOrg    bool     // Auto-accept shares within same organization (if applicable)
	RequireApprovalFor   []string // Memory types that always require approval (e.g., ["client_context"])
}

// DefaultSharingPolicy returns the default sharing policy
// Can be overridden via environment variables
func DefaultSharingPolicy() *SharingPolicy {
	return &SharingPolicy{
		AutoAcceptSameTenant: getEnvBool("MEMORY_SHARE_AUTO_ACCEPT_SAME_TENANT", true),
		AutoAcceptSameOrg:    getEnvBool("MEMORY_SHARE_AUTO_ACCEPT_SAME_ORG", true),
		RequireApprovalFor:   []string{}, // By default, no types require manual approval
	}
}

func getEnvBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return strings.ToLower(val) == "true" || val == "1"
}

// MemorySharingManager manages cross-team memory sharing
type MemorySharingManager struct {
	repo   storage.Repository
	policy *SharingPolicy
}

// NewMemorySharingManager creates a new sharing manager
func NewMemorySharingManager(repo storage.Repository) *MemorySharingManager {
	return &MemorySharingManager{
		repo:   repo,
		policy: DefaultSharingPolicy(),
	}
}

// NewMemorySharingManagerWithPolicy creates a sharing manager with custom policy
func NewMemorySharingManagerWithPolicy(repo storage.Repository, policy *SharingPolicy) *MemorySharingManager {
	return &MemorySharingManager{
		repo:   repo,
		policy: policy,
	}
}

// SetPolicy updates the sharing policy
func (m *MemorySharingManager) SetPolicy(policy *SharingPolicy) {
	m.policy = policy
}

// teamsInSameTenant checks if two teams belong to the same tenant
func (m *MemorySharingManager) teamsInSameTenant(ctx context.Context, teamID1, teamID2 uuid.UUID) (bool, error) {
	// Get both teams
	team1, err := m.repo.GetTeamByID(ctx, teamID1)
	if err != nil {
		return false, fmt.Errorf("failed to get source team: %w", err)
	}
	if team1 == nil {
		return false, fmt.Errorf("source team not found")
	}

	team2, err := m.repo.GetTeamByID(ctx, teamID2)
	if err != nil {
		return false, fmt.Errorf("failed to get target team: %w", err)
	}
	if team2 == nil {
		return false, fmt.Errorf("target team not found")
	}

	return team1.TenantID == team2.TenantID, nil
}

// requiresManualApproval checks if a memory type requires manual approval
func (m *MemorySharingManager) requiresManualApproval(memoryType string) bool {
	for _, t := range m.policy.RequireApprovalFor {
		if t == memoryType {
			return true
		}
	}
	return false
}

// ShareMemoryRequest represents a request to share a memory
type ShareMemoryRequest struct {
	MemoryID     uuid.UUID  `json:"memory_id"`
	SourceTeamID uuid.UUID  `json:"source_team_id"`
	TargetTeamID uuid.UUID  `json:"target_team_id"`
	SharedBy     uuid.UUID  `json:"shared_by"`
	ShareType    ShareType  `json:"share_type"`
	Permission   Permission `json:"permission"`
	Message      *string    `json:"message,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// ShareMemory shares a memory with another team
func (m *MemorySharingManager) ShareMemory(ctx context.Context, req ShareMemoryRequest) (*storage.MemoryShare, error) {
	// Extract tenant ID from context for proper isolation
	tenantID := GetTenantIDFromContext(ctx)

	// Validate the memory exists and belongs to the source team
	memory, err := m.repo.GetTeamMemoryByID(ctx, tenantID, req.SourceTeamID, req.MemoryID)
	if err != nil {
		return nil, fmt.Errorf("memory not found or access denied: %w", err)
	}

	// Check if already shared
	existing, err := m.GetShareBetweenTeams(ctx, req.MemoryID, req.SourceTeamID, req.TargetTeamID)
	if err == nil && existing != nil && existing.Status == string(ShareStatusAccepted) {
		return nil, fmt.Errorf("memory already shared with this team")
	}

	share := &storage.MemoryShare{
		ID:           uuid.New(),
		MemoryID:     req.MemoryID,
		SourceTeamID: req.SourceTeamID,
		TargetTeamID: req.TargetTeamID,
		SharedBy:     req.SharedBy,
		ShareType:    string(req.ShareType),
		Permission:   string(req.Permission),
		Status:       string(ShareStatusPending),
		Message:      req.Message,
		ExpiresAt:    req.ExpiresAt,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save the share request
	err = m.repo.CreateMemoryShare(ctx, share)
	if err != nil {
		return nil, fmt.Errorf("failed to create share: %w", err)
	}

	// Check if auto-accept should apply
	shouldAutoAccept := false
	autoAcceptReason := ""

	// Check if teams are in the same tenant and policy allows auto-accept
	if m.policy.AutoAcceptSameTenant {
		sameTenant, err := m.teamsInSameTenant(ctx, req.SourceTeamID, req.TargetTeamID)
		if err != nil {
			logrus.WithError(err).Warn("Failed to check if teams are in same tenant, leaving share as pending")
		} else if sameTenant {
			// Check if memory type requires manual approval
			if !m.requiresManualApproval(memory.MemoryType) {
				shouldAutoAccept = true
				autoAcceptReason = "auto_accept_same_tenant"
			}
		}
	}

	// Auto-accept if conditions are met
	if shouldAutoAccept {
		now := time.Now()
		share.Status = string(ShareStatusAccepted)
		share.AcceptedBy = &req.SharedBy // The sharer becomes the accepter in auto-accept
		share.AcceptedAt = &now
		share.UpdatedAt = now

		// Handle different share types
		switch share.ShareType {
		case string(ShareTypeReference):
			err = m.createReferenceMemory(ctx, share)
		case string(ShareTypeCopy):
			err = m.createMemoryCopy(ctx, share)
		case string(ShareTypeFork):
			err = m.createMemoryFork(ctx, share)
		}

		if err != nil {
			logrus.WithError(err).Warn("Failed to create shared memory in auto-accept, leaving share as pending")
			// Reset to pending if creation failed
			share.Status = string(ShareStatusPending)
			share.AcceptedBy = nil
			share.AcceptedAt = nil
		} else {
			// Update the share
			err = m.repo.UpdateMemoryShare(ctx, share)
			if err != nil {
				logrus.WithError(err).Warn("Failed to update share in auto-accept")
			}
		}

		logrus.WithFields(logrus.Fields{
			"share_id":           share.ID,
			"memory_id":          req.MemoryID,
			"source_team":        req.SourceTeamID,
			"target_team":        req.TargetTeamID,
			"auto_accept":        true,
			"auto_accept_reason": autoAcceptReason,
		}).Info("Memory share auto-accepted")
	}

	logrus.WithFields(logrus.Fields{
		"share_id":      share.ID,
		"memory_id":     req.MemoryID,
		"source_team":   req.SourceTeamID,
		"target_team":   req.TargetTeamID,
		"share_type":    req.ShareType,
		"status":        share.Status,
		"auto_accepted": shouldAutoAccept,
	}).Info("Memory shared with team")

	// Record metric
	monitoring.RecordTeamMemoryCreated(req.TargetTeamID.String(), memory.MemoryType, "shared")

	return share, nil
}

// AcceptShare accepts a pending memory share
func (m *MemorySharingManager) AcceptShare(ctx context.Context, shareID, acceptedBy uuid.UUID, message *string) (*storage.MemoryShare, error) {
	share, err := m.repo.GetMemoryShareByID(ctx, shareID)
	if err != nil {
		return nil, fmt.Errorf("share not found: %w", err)
	}

	if share.Status != string(ShareStatusPending) {
		return nil, fmt.Errorf("share is not pending")
	}

	now := time.Now()
	share.Status = string(ShareStatusAccepted)
	share.AcceptedBy = &acceptedBy
	share.AcceptedAt = &now
	share.UpdatedAt = now

	if message != nil {
		share.Message = message
	}

	// Handle different share types
	switch share.ShareType {
	case string(ShareTypeReference):
		// Create a reference memory in the target team
		err = m.createReferenceMemory(ctx, share)
	case string(ShareTypeCopy):
		// Create an independent copy
		err = m.createMemoryCopy(ctx, share)
	case string(ShareTypeFork):
		// Create a fork that tracks ancestry
		err = m.createMemoryFork(ctx, share)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create shared memory: %w", err)
	}

	// Update the share
	err = m.repo.UpdateMemoryShare(ctx, share)
	if err != nil {
		return nil, fmt.Errorf("failed to update share: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"share_id":    shareID,
		"memory_id":   share.MemoryID,
		"accepted_by": acceptedBy,
	}).Info("Memory share accepted")

	return share, nil
}

// RejectShare rejects a pending memory share
func (m *MemorySharingManager) RejectShare(ctx context.Context, shareID uuid.UUID, reason *string) error {
	share, err := m.repo.GetMemoryShareByID(ctx, shareID)
	if err != nil {
		return fmt.Errorf("share not found: %w", err)
	}

	if share.Status != string(ShareStatusPending) {
		return fmt.Errorf("share is not pending")
	}

	share.Status = string(ShareStatusRejected)
	share.UpdatedAt = time.Now()
	if reason != nil {
		share.Message = reason
	}

	return m.repo.UpdateMemoryShare(ctx, share)
}

// RevokeShare revokes an accepted share (source team only)
func (m *MemorySharingManager) RevokeShare(ctx context.Context, shareID, revokedBy uuid.UUID, reason string) error {
	share, err := m.repo.GetMemoryShareByID(ctx, shareID)
	if err != nil {
		return fmt.Errorf("share not found: %w", err)
	}

	// Only source team or admin can revoke
	// (Permission check should be done at handler level)

	share.Status = string(ShareStatusRevoked)
	share.UpdatedAt = time.Now()

	// Remove or disable the shared memory in target team
	err = m.revokeSharedMemory(ctx, share)
	if err != nil {
		logrus.WithError(err).Warn("Failed to revoke shared memory, but marking share as revoked")
	}

	return m.repo.UpdateMemoryShare(ctx, share)
}

// ListSharedMemories lists memories shared with a team
func (m *MemorySharingManager) ListSharedMemories(ctx context.Context, teamID uuid.UUID, status ShareStatus, limit, offset int) ([]*storage.MemoryShare, error) {
	return m.repo.ListMemorySharesByTargetTeam(ctx, teamID, string(status), limit, offset)
}

// ListOutgoingShares lists memories shared by a team
func (m *MemorySharingManager) ListOutgoingShares(ctx context.Context, teamID uuid.UUID, status ShareStatus, limit, offset int) ([]*storage.MemoryShare, error) {
	return m.repo.ListMemorySharesBySourceTeam(ctx, teamID, string(status), limit, offset)
}

// GetShareBetweenTeams gets an existing share between two teams
func (m *MemorySharingManager) GetShareBetweenTeams(ctx context.Context, memoryID, sourceTeamID, targetTeamID uuid.UUID) (*storage.MemoryShare, error) {
	return m.repo.GetMemoryShareBetweenTeams(ctx, memoryID, sourceTeamID, targetTeamID)
}

// CreateMemoryShareForTeam creates a shared memory entry in the target team
func (m *MemorySharingManager) createReferenceMemory(ctx context.Context, share *storage.MemoryShare) error {
	// Extract tenant ID from context for proper isolation
	tenantID := GetTenantIDFromContext(ctx)

	// Get the original memory
	memory, err := m.repo.GetTeamMemoryByID(ctx, tenantID, share.SourceTeamID, share.MemoryID)
	if err != nil {
		return fmt.Errorf("failed to get original memory: %w", err)
	}

	// Create a reference entry (with shared metadata)
	summary := fmt.Sprintf("[Shared] %s", dereferenceString(memory.Summary))
	sharedMemory := &storage.TeamMemory{
		TenantID:        memory.TenantID, // May need adjustment for cross-tenant
		TeamID:          share.TargetTeamID,
		MemoryType:      memory.MemoryType,
		Category:        memory.Category,
		Content:         memory.Content,
		Summary:         &summary,
		CreatedBy:       share.SharedBy,
		IsValidated:     memory.IsValidated,
		ConfidenceScore: memory.ConfidenceScore,
		TTLDays:         memory.TTLDays,
	}

	created, err := m.repo.CreateTeamMemory(ctx, sharedMemory)
	if err != nil {
		return fmt.Errorf("failed to create reference memory: %w", err)
	}

	// Store the target memory ID for revocation
	share.TargetMemoryID = &created.ID

	return nil
}

// CreateMemoryCopy creates an independent copy of a memory
func (m *MemorySharingManager) createMemoryCopy(ctx context.Context, share *storage.MemoryShare) error {
	// Extract tenant ID from context for proper isolation
	tenantID := GetTenantIDFromContext(ctx)

	memory, err := m.repo.GetTeamMemoryByID(ctx, tenantID, share.SourceTeamID, share.MemoryID)
	if err != nil {
		return fmt.Errorf("failed to get original memory: %w", err)
	}

	// Create independent copy
	summary := fmt.Sprintf("[From %s] %s", share.SourceTeamID.String()[:8], dereferenceString(memory.Summary))
	copy := &storage.TeamMemory{
		TenantID:        memory.TenantID,
		TeamID:          share.TargetTeamID,
		MemoryType:      memory.MemoryType,
		Category:        memory.Category,
		Content:         memory.Content,
		Summary:         &summary,
		CreatedBy:       share.SharedBy,
		IsValidated:     false, // Require validation in target team
		ConfidenceScore: memory.ConfidenceScore,
	}

	created, err := m.repo.CreateTeamMemory(ctx, copy)
	if err != nil {
		return fmt.Errorf("failed to create memory copy: %w", err)
	}

	// Store the target memory ID for revocation
	share.TargetMemoryID = &created.ID

	return nil
}

// CreateMemoryFork creates a fork (copy with ancestry tracking)
func (m *MemorySharingManager) createMemoryFork(ctx context.Context, share *storage.MemoryShare) error {
	// Extract tenant ID from context for proper isolation
	tenantID := GetTenantIDFromContext(ctx)

	memory, err := m.repo.GetTeamMemoryByID(ctx, tenantID, share.SourceTeamID, share.MemoryID)
	if err != nil {
		return fmt.Errorf("failed to get original memory: %w", err)
	}

	// Create fork with ancestry info in content
	forkContent := make(map[string]interface{})
	for k, v := range memory.Content {
		forkContent[k] = v
	}
	forkContent["_fork_info"] = map[string]interface{}{
		"source_memory_id": share.MemoryID,
		"source_team_id":   share.SourceTeamID,
		"forked_at":        time.Now().Format(time.RFC3339),
	}

	summary := fmt.Sprintf("[Fork] %s", dereferenceString(memory.Summary))
	fork := &storage.TeamMemory{
		TenantID:        memory.TenantID,
		TeamID:          share.TargetTeamID,
		MemoryType:      memory.MemoryType,
		Category:        memory.Category,
		Content:         forkContent,
		Summary:         &summary,
		CreatedBy:       share.SharedBy,
		IsValidated:     false,
		ConfidenceScore: memory.ConfidenceScore,
	}

	created, err := m.repo.CreateTeamMemory(ctx, fork)
	if err != nil {
		return fmt.Errorf("failed to create memory fork: %w", err)
	}

	// Store the target memory ID for revocation
	share.TargetMemoryID = &created.ID

	return nil
}

// RevokeSharedMemory removes or disables a shared memory in the target team
func (m *MemorySharingManager) revokeSharedMemory(ctx context.Context, share *storage.MemoryShare) error {
	var targetMemoryID *uuid.UUID

	// First, try to use the stored target_memory_id (preferred, direct lookup)
	if share.TargetMemoryID != nil {
		targetMemoryID = share.TargetMemoryID
		logrus.WithFields(logrus.Fields{
			"share_id":         share.ID,
			"target_memory_id": *targetMemoryID,
		}).Debug("Using stored target_memory_id for revocation")
	} else {
		// Fallback: Find by searching (for backward compatibility with older shares)
		// This is less reliable but handles shares created before target_memory_id was added
		logrus.WithFields(logrus.Fields{
			"share_id": share.ID,
		}).Warn("No target_memory_id stored, falling back to search for revocation")
		targetMemoryID = m.findSharedMemoryBySearch(ctx, share)
	}

	if targetMemoryID == nil {
		logrus.WithFields(logrus.Fields{
			"share_id":    share.ID,
			"memory_id":   share.MemoryID,
			"source_team": share.SourceTeamID,
			"target_team": share.TargetTeamID,
		}).Warn("Could not find shared memory in target team to revoke")
		return nil // Not an error, may have been deleted already
	}

	// Extract tenant ID from context for proper isolation
	tenantID := GetTenantIDFromContext(ctx)

	// Delete the shared memory based on share type
	var err error
	switch share.ShareType {
	case string(ShareTypeReference):
		// For references, we can safely delete as they're just pointers
		err = m.repo.DeleteTeamMemory(ctx, tenantID, share.TargetTeamID, *targetMemoryID)
		if err != nil {
			return fmt.Errorf("failed to delete reference memory: %w", err)
		}

	case string(ShareTypeCopy), string(ShareTypeFork):
		// For copies and forks, we should mark as revoked rather than delete
		// to preserve audit trail
		memory, err := m.repo.GetTeamMemoryByID(ctx, tenantID, share.TargetTeamID, *targetMemoryID)
		if err != nil {
			return fmt.Errorf("failed to get shared memory: %w", err)
		}

		// Mark as disabled/revoked by updating content
		if memory.Content == nil {
			memory.Content = make(map[string]interface{})
		}
		memory.Content["_revoked"] = map[string]interface{}{
			"revoked_at": time.Now().Format(time.RFC3339),
			"revoked_by": share.SharedBy,
			"share_id":   share.ID,
			"reason":     "share_revoked",
		}

		// Update summary to indicate revocation
		revokedSummary := fmt.Sprintf("[REVOKED] %s", dereferenceString(memory.Summary))
		memory.Summary = &revokedSummary

		_, err = m.repo.UpdateTeamMemory(ctx, memory)
		if err != nil {
			return fmt.Errorf("failed to mark memory as revoked: %w", err)
		}
	}

	logrus.WithFields(logrus.Fields{
		"share_id":         share.ID,
		"memory_id":        share.MemoryID,
		"target_memory_id": *targetMemoryID,
		"target_team":      share.TargetTeamID,
		"share_type":       share.ShareType,
	}).Info("Shared memory revoked in target team")

	return nil
}

// SyncSharedReferences syncs updates to referenced memories across teams
func (m *MemorySharingManager) SyncSharedReferences(ctx context.Context, memoryID, teamID uuid.UUID) error {
	// Get all shares for this memory
	shares, err := m.repo.ListMemorySharesByMemoryID(ctx, memoryID, string(ShareStatusAccepted))
	if err != nil {
		return err
	}

	// For each reference-type share, update the shared memory
	for _, share := range shares {
		if share.ShareType == string(ShareTypeReference) {
			// Update the shared memory
			// (Implementation depends on how we track the reference)
			logrus.WithFields(logrus.Fields{
				"memory_id":   memoryID,
				"target_team": share.TargetTeamID,
			}).Debug("Syncing shared memory reference")
		}
	}

	return nil
}

// findSharedMemoryBySearch searches for a shared memory in the target team using heuristics
// This is a fallback for shares created before target_memory_id was tracked
func (m *MemorySharingManager) findSharedMemoryBySearch(ctx context.Context, share *storage.MemoryShare) *uuid.UUID {
	filter := storage.TeamMemoryFilter{
		Limit: 100,
	}

	// Extract tenant ID from context for proper isolation
	tenantID := GetTenantIDFromContext(ctx)

	// Search by summary prefixes that indicate shared memories
	prefixes := []string{
		"[Shared] ", // Reference shares
		fmt.Sprintf("[From %s] ", share.SourceTeamID.String()[:8]), // Copy shares
		"[Fork] ", // Fork shares
	}

	memories, _, err := m.repo.ListTeamMemories(ctx, tenantID, share.TargetTeamID, filter)
	if err != nil {
		logrus.WithError(err).Error("Failed to list memories for revocation search")
		return nil
	}

	for _, mem := range memories {
		if mem.Summary == nil {
			continue
		}
		summary := *mem.Summary

		// Check if this memory matches any of our share prefixes
		for _, prefix := range prefixes {
			if len(summary) >= len(prefix) && summary[:len(prefix)] == prefix {
				// Additional check: verify this was created by the sharer
				if mem.CreatedBy == share.SharedBy {
					return &mem.ID
				}
			}
		}
	}

	return nil
}

// dereferenceString is defined in auto_updater.go
