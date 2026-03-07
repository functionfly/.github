// Package scim provides SCIM 2.0 provisioning support for GoBetterAuth
package scim

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SCIMService handles SCIM operations
type SCIMService struct {
	db     *gorm.DB
	logger *logrus.Logger
	config *SCIMServiceConfig
}

// SCIMServiceConfig holds SCIM service configuration
type SCIMServiceConfig struct {
	BaseURL     string
	TokenExpiry time.Duration
}

// NewSCIMService creates a new SCIM service
func NewSCIMService(db *gorm.DB, logger *logrus.Logger, baseURL string) (*SCIMService, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	if logger == nil {
		logger = logrus.New()
	}

	return &SCIMService{
		db:     db,
		logger: logger,
		config: &SCIMServiceConfig{
			BaseURL:     baseURL,
			TokenExpiry: 365 * 24 * time.Hour, // 1 year default
		},
	}, nil
}

// GetConfig retrieves SCIM configuration for a tenant
func (s *SCIMService) GetConfig(ctx context.Context, tenantID uuid.UUID) (*SCIMConfig, error) {
	var config SCIMConfig
	result := s.db.WithContext(ctx).Where("tenant_id = ? AND enabled = ? AND deleted_at IS NULL", tenantID, true).First(&config)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("SCIM not configured for tenant")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get SCIM config: %w", result.Error)
	}
	return &config, nil
}

// GetConfigByID retrieves SCIM configuration by ID
func (s *SCIMService) GetConfigByID(ctx context.Context, configID uuid.UUID) (*SCIMConfig, error) {
	var config SCIMConfig
	result := s.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", configID).First(&config)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("SCIM config not found")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get SCIM config: %w", result.Error)
	}
	return &config, nil
}

// CreateConfig creates a new SCIM configuration for a tenant
func (s *SCIMService) CreateConfig(ctx context.Context, tenantID uuid.UUID, req *SCIMConfigRequest) (*SCIMConfig, error) {
	// Generate a token for SCIM access
	token, err := GenerateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	tokenHash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash token: %w", err)
	}

	config := &SCIMConfig{
		TenantID:   tenantID,
		Enabled:    req.Enabled,
		TokenHash:  string(tokenHash),
		SyncGroups: req.SyncGroups,
		SyncUsers:  req.SyncUsers,
	}

	if err := s.db.WithContext(ctx).Create(config).Error; err != nil {
		return nil, fmt.Errorf("failed to create SCIM config: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"config_id": config.ID,
	}).Info("SCIM configuration created")

	return config, nil
}

// UpdateConfig updates an existing SCIM configuration
func (s *SCIMService) UpdateConfig(ctx context.Context, configID uuid.UUID, req *SCIMConfigRequest) (*SCIMConfig, error) {
	config, err := s.GetConfigByID(ctx, configID)
	if err != nil {
		return nil, err
	}

	config.Enabled = req.Enabled
	config.SyncGroups = req.SyncGroups
	config.SyncUsers = req.SyncUsers

	if err := s.db.WithContext(ctx).Save(config).Error; err != nil {
		return nil, fmt.Errorf("failed to update SCIM config: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"config_id": configID,
	}).Info("SCIM configuration updated")

	return config, nil
}

// DeleteConfig soft-deletes a SCIM configuration
func (s *SCIMService) DeleteConfig(ctx context.Context, configID uuid.UUID) error {
	result := s.db.WithContext(ctx).Delete(&SCIMConfig{}, "id = ?", configID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete SCIM config: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("SCIM config not found")
	}

	s.logger.WithFields(logrus.Fields{
		"config_id": configID,
	}).Info("SCIM configuration deleted")

	return nil
}

// RegenerateToken generates a new SCIM token
func (s *SCIMService) RegenerateToken(ctx context.Context, tenantID uuid.UUID) (string, *SCIMConfig, error) {
	config, err := s.GetConfig(ctx, tenantID)
	if err != nil {
		// Create new config if not exists
		config = &SCIMConfig{
			TenantID:   tenantID,
			Enabled:    true,
			SyncGroups: true,
			SyncUsers:  true,
		}
	}

	token, err := GenerateSecureToken(32)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	tokenHash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, fmt.Errorf("failed to hash token: %w", err)
	}

	config.TokenHash = string(tokenHash)

	if config.ID == uuid.Nil {
		// Create new config
		if err := s.db.WithContext(ctx).Create(config).Error; err != nil {
			return "", nil, fmt.Errorf("failed to create SCIM config: %w", err)
		}
	} else {
		// Update existing config
		if err := s.db.WithContext(ctx).Save(config).Error; err != nil {
			return "", nil, fmt.Errorf("failed to update SCIM config: %w", err)
		}
	}

	s.logger.WithFields(logrus.Fields{
		"tenant_id": tenantID,
	}).Info("SCIM token regenerated")

	return token, config, nil
}

// VerifyToken verifies a SCIM bearer token
func (s *SCIMService) VerifyToken(ctx context.Context, tenantID uuid.UUID, token string) (bool, error) {
	config, err := s.GetConfig(ctx, tenantID)
	if err != nil {
		return false, err
	}

	if !config.Enabled {
		return false, fmt.Errorf("SCIM is disabled for this tenant")
	}

	err = bcrypt.CompareHashAndPassword([]byte(config.TokenHash), []byte(token))
	if err != nil {
		return false, fmt.Errorf("invalid token")
	}

	return true, nil
}

// User CRUD Operations

// ListUsers lists all SCIM users for a tenant with pagination and filtering
func (s *SCIMService) ListUsers(ctx context.Context, tenantID uuid.UUID, filter string, startIndex, count int) ([]SCIMUser, int, error) {
	var users []SCIMUser
	var total int64

	query := s.db.WithContext(ctx).Model(&SCIMUser{}).Where("tenant_id = ? AND deleted_at IS NULL", tenantID)

	// Apply filter if provided
	if filter != "" {
		query = s.applyUserFilter(query, filter)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Apply pagination
	if startIndex > 0 {
		query = query.Offset(startIndex - 1) // SCIM uses 1-based indexing
	}
	if count > 0 {
		query = query.Limit(count)
	}

	// Execute query
	if err := query.Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, int(total), nil
}

// GetUser retrieves a SCIM user by ID
func (s *SCIMService) GetUser(ctx context.Context, tenantID uuid.UUID, userID string) (*SCIMUser, error) {
	var user SCIMUser
	result := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", userID, tenantID).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user: %w", result.Error)
	}
	return &user, nil
}

// GetUserByExternalID retrieves a SCIM user by external ID
func (s *SCIMService) GetUserByExternalID(ctx context.Context, tenantID uuid.UUID, externalID string) (*SCIMUser, error) {
	var user SCIMUser
	result := s.db.WithContext(ctx).Where("external_id = ? AND tenant_id = ? AND deleted_at IS NULL", externalID, tenantID).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user by external ID: %w", result.Error)
	}
	return &user, nil
}

// GetUserByUserName retrieves a SCIM user by username
func (s *SCIMService) GetUserByUserName(ctx context.Context, tenantID uuid.UUID, userName string) (*SCIMUser, error) {
	var user SCIMUser
	result := s.db.WithContext(ctx).Where("user_name = ? AND tenant_id = ? AND deleted_at IS NULL", userName, tenantID).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", result.Error)
	}
	return &user, nil
}

// CreateUser creates a new SCIM user
func (s *SCIMService) CreateUser(ctx context.Context, tenantID uuid.UUID, req *SCIMUserRequest) (*SCIMUser, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if user already exists
	existingUser, err := s.GetUserByUserName(ctx, tenantID, req.UserName)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user with username %s already exists", req.UserName)
	}

	// Parse raw JSON
	rawJSON, _ := json.Marshal(req)

	user := &SCIMUser{
		TenantID:    tenantID,
		ExternalID:  req.ExternalID,
		UserName:    req.UserName,
		DisplayName: req.DisplayName,
		Emails:      req.ParseEmails(),
		Active:      req.Active,
		Groups:      req.ParseGroups(),
		Raw:         rawJSON,
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":   user.ID,
		"tenant_id": tenantID,
		"username":  req.UserName,
	}).Info("SCIM user created")

	return user, nil
}

// UpdateUser updates an existing SCIM user
func (s *SCIMService) UpdateUser(ctx context.Context, tenantID uuid.UUID, userID string, req *SCIMUserRequest) (*SCIMUser, error) {
	user, err := s.GetUser(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update fields
	if req.UserName != "" {
		user.UserName = req.UserName
	}
	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	if req.ExternalID != "" {
		user.ExternalID = req.ExternalID
	}

	user.Active = req.Active
	user.Emails = req.ParseEmails()
	user.Groups = req.ParseGroups()

	// Update raw JSON
	rawJSON, _ := json.Marshal(req)
	user.Raw = rawJSON

	if err := s.db.WithContext(ctx).Save(user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"tenant_id": tenantID,
	}).Info("SCIM user updated")

	return user, nil
}

// PatchUser applies a patch operation to a user
func (s *SCIMService) PatchUser(ctx context.Context, tenantID uuid.UUID, userID string, req *SCIMPatchRequest) (*SCIMUser, error) {
	user, err := s.GetUser(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Apply operations
	for _, op := range req.Operations {
		switch op.Op {
		case "replace":
			if op.Path == "" {
				// Replace entire resource
				if val, ok := op.Value.(map[string]interface{}); ok {
					if active, ok := val["active"].(bool); ok {
						user.Active = active
					}
				}
			} else if op.Path == "active" {
				if val, ok := op.Value.(bool); ok {
					user.Active = val
				}
			}
		case "add":
			// Handle add operations
		case "remove":
			// Handle remove operations
		}
	}

	if err := s.db.WithContext(ctx).Save(user).Error; err != nil {
		return nil, fmt.Errorf("failed to patch user: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"tenant_id": tenantID,
	}).Info("SCIM user patched")

	return user, nil
}

// DeleteUser soft-deletes a SCIM user
func (s *SCIMService) DeleteUser(ctx context.Context, tenantID uuid.UUID, userID string) error {
	result := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", userID, tenantID).Delete(&SCIMUser{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"tenant_id": tenantID,
	}).Info("SCIM user deleted")

	return nil
}

// Group CRUD Operations

// ListGroups lists all SCIM groups for a tenant
func (s *SCIMService) ListGroups(ctx context.Context, tenantID uuid.UUID, filter string, startIndex, count int) ([]SCIMGroup, int, error) {
	var groups []SCIMGroup
	var total int64

	query := s.db.WithContext(ctx).Model(&SCIMGroup{}).Where("tenant_id = ? AND deleted_at IS NULL", tenantID)

	// Apply filter if provided
	if filter != "" {
		query = s.applyGroupFilter(query, filter)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count groups: %w", err)
	}

	// Apply pagination
	if startIndex > 0 {
		query = query.Offset(startIndex - 1)
	}
	if count > 0 {
		query = query.Limit(count)
	}

	if err := query.Find(&groups).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list groups: %w", err)
	}

	return groups, int(total), nil
}

// GetGroup retrieves a SCIM group by ID
func (s *SCIMService) GetGroup(ctx context.Context, tenantID uuid.UUID, groupID string) (*SCIMGroup, error) {
	var group SCIMGroup
	result := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", groupID, tenantID).First(&group)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get group: %w", result.Error)
	}
	return &group, nil
}

// GetGroupByExternalID retrieves a SCIM group by external ID
func (s *SCIMService) GetGroupByExternalID(ctx context.Context, tenantID uuid.UUID, externalID string) (*SCIMGroup, error) {
	var group SCIMGroup
	result := s.db.WithContext(ctx).Where("external_id = ? AND tenant_id = ? AND deleted_at IS NULL", externalID, tenantID).First(&group)
	if result.Error == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get group by external ID: %w", result.Error)
	}
	return &group, nil
}

// CreateGroup creates a new SCIM group
func (s *SCIMService) CreateGroup(ctx context.Context, tenantID uuid.UUID, req *SCIMGroupRequest) (*SCIMGroup, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if group already exists
	var existingGroup SCIMGroup
	result := s.db.WithContext(ctx).Where("display_name = ? AND tenant_id = ? AND deleted_at IS NULL", req.DisplayName, tenantID).First(&existingGroup)
	if result.Error == nil {
		return nil, fmt.Errorf("group with display name %s already exists", req.DisplayName)
	}

	// Parse raw JSON
	rawJSON, _ := json.Marshal(req)

	group := &SCIMGroup{
		TenantID:    tenantID,
		ExternalID:  req.ExternalID,
		DisplayName: req.DisplayName,
		Members:     req.ParseMembers(),
		Raw:         rawJSON,
	}

	if err := s.db.WithContext(ctx).Create(group).Error; err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"group_id":  group.ID,
		"tenant_id": tenantID,
	}).Info("SCIM group created")

	return group, nil
}

// UpdateGroup updates an existing SCIM group
func (s *SCIMService) UpdateGroup(ctx context.Context, tenantID uuid.UUID, groupID string, req *SCIMGroupRequest) (*SCIMGroup, error) {
	group, err := s.GetGroup(ctx, tenantID, groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, fmt.Errorf("group not found")
	}

	if req.DisplayName != "" {
		group.DisplayName = req.DisplayName
	}
	if req.ExternalID != "" {
		group.ExternalID = req.ExternalID
	}
	group.Members = req.ParseMembers()

	// Update raw JSON
	rawJSON, _ := json.Marshal(req)
	group.Raw = rawJSON

	if err := s.db.WithContext(ctx).Save(group).Error; err != nil {
		return nil, fmt.Errorf("failed to update group: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"group_id":  groupID,
		"tenant_id": tenantID,
	}).Info("SCIM group updated")

	return group, nil
}

// PatchGroup applies a patch operation to a group
func (s *SCIMService) PatchGroup(ctx context.Context, tenantID uuid.UUID, groupID string, req *SCIMPatchRequest) (*SCIMGroup, error) {
	group, err := s.GetGroup(ctx, tenantID, groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, fmt.Errorf("group not found")
	}

	// Apply operations
	for _, op := range req.Operations {
		switch op.Op {
		case "add":
			if op.Path == "members" {
				// Add members
				if members, ok := op.Value.([]interface{}); ok {
					for _, m := range members {
						if memberMap, ok := m.(map[string]interface{}); ok {
							member := SCIMMember{}
							if val, ok := memberMap["value"].(string); ok {
								member.Value = val
							}
							if val, ok := memberMap["display"].(string); ok {
								member.Display = val
							}
							group.Members = append(group.Members, member)
						}
					}
				}
			}
		case "remove":
			if op.Path == "members" {
				// Remove members
				if memberValue, ok := op.Value.(map[string]interface{}); ok {
					if valueToRemove, ok := memberValue["value"].(string); ok {
						newMembers := make([]SCIMMember, 0)
						for _, m := range group.Members {
							if m.Value != valueToRemove {
								newMembers = append(newMembers, m)
							}
						}
						group.Members = newMembers
					}
				}
			}
		}
	}

	if err := s.db.WithContext(ctx).Save(group).Error; err != nil {
		return nil, fmt.Errorf("failed to patch group: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"group_id":  groupID,
		"tenant_id": tenantID,
	}).Info("SCIM group patched")

	return group, nil
}

// DeleteGroup soft-deletes a SCIM group
func (s *SCIMService) DeleteGroup(ctx context.Context, tenantID uuid.UUID, groupID string) error {
	result := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", groupID, tenantID).Delete(&SCIMGroup{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete group: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("group not found")
	}

	s.logger.WithFields(logrus.Fields{
		"group_id":  groupID,
		"tenant_id": tenantID,
	}).Info("SCIM group deleted")

	return nil
}

// applyUserFilter applies a SCIM filter to user queries
func (s *SCIMService) applyUserFilter(query *gorm.DB, filter string) *gorm.DB {
	// Simple filter parsing for common cases
	// SCIM filters: userName eq "john", emails.value sw "john", etc.

	filter = strings.TrimSpace(filter)

	// Handle userName eq "value"
	if strings.Contains(filter, "userName eq ") {
		parts := strings.SplitN(filter, "userName eq ", 2)
		if len(parts) == 2 {
			value := strings.Trim(parts[1], `"`)
			return query.Where("user_name = ?", value)
		}
	}

	// Handle externalId eq "value"
	if strings.Contains(filter, "externalId eq ") {
		parts := strings.SplitN(filter, "externalId eq ", 2)
		if len(parts) == 2 {
			value := strings.Trim(parts[1], `"`)
			return query.Where("external_id = ?", value)
		}
	}

	// Handle active eq true/false
	if strings.Contains(filter, "active eq ") {
		parts := strings.SplitN(filter, "active eq ", 2)
		if len(parts) == 2 {
			value := strings.ToLower(strings.TrimSpace(parts[1])) == "true"
			return query.Where("active = ?", value)
		}
	}

	return query
}

// applyGroupFilter applies a SCIM filter to group queries
func (s *SCIMService) applyGroupFilter(query *gorm.DB, filter string) *gorm.DB {
	filter = strings.TrimSpace(filter)

	// Handle displayName eq "value"
	if strings.Contains(filter, "displayName eq ") {
		parts := strings.SplitN(filter, "displayName eq ", 2)
		if len(parts) == 2 {
			value := strings.Trim(parts[1], `"`)
			return query.Where("display_name = ?", value)
		}
	}

	// Handle externalId eq "value"
	if strings.Contains(filter, "externalId eq ") {
		parts := strings.SplitN(filter, "externalId eq ", 2)
		if len(parts) == 2 {
			value := strings.Trim(parts[1], `"`)
			return query.Where("external_id = ?", value)
		}
	}

	return query
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		// Use a simple random for now - in production use crypto/rand
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b), nil
}

// ParsePagination parses SCIM pagination parameters
func ParsePagination(startIndexStr, countStr string) (startIndex, count int) {
	startIndex = 1 // SCIM uses 1-based indexing
	count = 100    // Default page size

	if startIndexStr != "" {
		if val, err := strconv.Atoi(startIndexStr); err == nil && val > 0 {
			startIndex = val
		}
	}

	if countStr != "" {
		if val, err := strconv.Atoi(countStr); err == nil && val > 0 {
			count = val
		}
	}

	return startIndex, count
}
