package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SCIMService provides SCIM 2.0 protocol implementation
type SCIMService struct {
	configRepo  *storage.SCIMConfigRepository
	syncLogRepo *storage.SCIMSyncLogRepository
	userRepo    *storage.UserRepository
	teamRepo    *storage.TeamRepository
	db          *gorm.DB
	logger      *logrus.Logger
}

// SCIMServiceConfig holds the configuration for SCIM service
type SCIMServiceConfig struct {
	ConfigRepo  *storage.SCIMConfigRepository
	SyncLogRepo *storage.SCIMSyncLogRepository
	UserRepo    *storage.UserRepository
	TeamRepo    *storage.TeamRepository
	DB          *gorm.DB
	Logger      *logrus.Logger
}

// NewSCIMService creates a new SCIM service
func NewSCIMService(config SCIMServiceConfig) *SCIMService {
	logger := config.Logger
	if logger == nil {
		logger = logrus.New()
	}
	return &SCIMService{
		configRepo:  config.ConfigRepo,
		syncLogRepo: config.SyncLogRepo,
		userRepo:    config.UserRepo,
		teamRepo:    config.TeamRepo,
		db:          config.DB,
		logger:      logger,
	}
}

// SCIMUser represents a SCIM User resource (RFC 7643)
type SCIMUser struct {
	ID           string         `json:"id"`
	UserName     string         `json:"userName"`
	Name         SCIMName       `json:"name"`
	Emails       []SCIMEmail    `json:"emails"`
	PhoneNumbers []SCIMPhone    `json:"phoneNumbers,omitempty"`
	DisplayName  string         `json:"displayName,omitempty"`
	Active       bool           `json:"active"`
	Groups       []SCIMGroupRef `json:"groups,omitempty"`
	Roles        []SCIMRoleRef  `json:"roles,omitempty"`
	Meta         SCIMMeta       `json:"meta"`
}

// SCIMName represents the name sub-attribute
type SCIMName struct {
	GivenName  string `json:"givenName,omitempty"`
	FamilyName string `json:"familyName,omitempty"`
	Formatted  string `json:"formatted,omitempty"`
}

// SCIMEmail represents an email sub-attribute
type SCIMEmail struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// SCIMPhone represents a phone number sub-attribute
type SCIMPhone struct {
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

// SCIMGroupRef represents a group reference
type SCIMGroupRef struct {
	Value   string `json:"value"`
	Display string `json:"display"`
	Type    string `json:"type,omitempty"`
}

// SCIMRoleRef represents a role reference
type SCIMRoleRef struct {
	Value   string `json:"value"`
	Display string `json:"display"`
	Type    string `json:"type,omitempty"`
}

// SCIMMeta represents the meta sub-attribute
type SCIMMeta struct {
	ResourceType string    `json:"resourceType"`
	Created      time.Time `json:"created"`
	Modified     time.Time `json:"lastModified"`
	Version      string    `json:"version,omitempty"`
}

// SCIMGroup represents a SCIM Group resource (RFC 7643)
type SCIMGroup struct {
	ID          string          `json:"id"`
	DisplayName string          `json:"displayName"`
	Members     []SCIMMemberRef `json:"members,omitempty"`
	Meta        SCIMMeta        `json:"meta"`
}

// SCIMMemberRef represents a member reference
type SCIMMemberRef struct {
	Value   string `json:"value"`
	Display string `json:"display"`
	Type    string `json:"type,omitempty"`
}

// SCIMListResponse represents a SCIM list response
type SCIMListResponse struct {
	TotalResults int         `json:"totalResults"`
	ItemsPerPage int         `json:"itemsPerPage"`
	StartIndex   int         `json:"startIndex"`
	Resources    interface{} `json:"resources"`
}

// SCIMError represents a SCIM error response
type SCIMError struct {
	Schemas []string          `json:"schemas"`
	Status  string            `json:"status"`
	Detail  string            `json:"detail"`
	Errors  []SCIMErrorDetail `json:"Errors,omitempty"`
}

// SCIMErrorDetail represents a detailed SCIM error
type SCIMErrorDetail struct {
	Description string `json:"description"`
	Code        string `json:"code"`
}

// SCIMPatchOperation represents a PATCH operation
type SCIMPatchOperation struct {
	Op    string          `json:"op"`
	Path  string          `json:"path,omitempty"`
	Value json.RawMessage `json:"value"`
}

// NewSCIMError creates a new SCIM error
func NewSCIMError(status int, detail string) *SCIMError {
	return &SCIMError{
		Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
		Status:  fmt.Sprintf("%d", status),
		Detail:  detail,
	}
}

// ToError converts SCIMError to standard error
func (e *SCIMError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Status, e.Detail)
}

// ToError converts SCIMError to standard error (for compatibility)
func (e *SCIMError) ToError() error {
	return fmt.Errorf("[%s] %s", e.Status, e.Detail)
}

// GetUser retrieves a user by ID for a tenant
func (s *SCIMService) GetUser(tenantID, userID uuid.UUID) (*SCIMUser, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verify user belongs to tenant
	if user.TenantID != tenantID {
		return nil, fmt.Errorf("user not found")
	}

	return s.userToSCIM(user)
}

// ListUsers lists users for a tenant with pagination
func (s *SCIMService) ListUsers(tenantID uuid.UUID, startIndex, count int) (*SCIMListResponse, error) {
	if startIndex < 1 {
		startIndex = 1
	}
	if count < 1 {
		count = 100
	}
	if count > 1000 {
		count = 1000
	}

	// Get total count for pagination
	var total int64
	if err := s.db.Model(&storage.User{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Query users with tenant_id filter directly at database level
	offset := startIndex - 1 // SCIM uses 1-based indexing
	var users []*storage.User
	if err := s.db.Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Offset(offset).
		Limit(count).
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	scimUsers := make([]SCIMUser, len(users))
	for i, user := range users {
		scimUser, err := s.userToSCIM(user)
		if err != nil {
			s.logger.WithError(err).Warnf("failed to convert user %s to SCIM", user.ID)
			continue
		}
		scimUsers[i] = *scimUser
	}

	return &SCIMListResponse{
		TotalResults: int(total),
		ItemsPerPage: count,
		StartIndex:   startIndex,
		Resources:    scimUsers,
	}, nil
}

// CreateUser creates a new user from SCIM request
func (s *SCIMService) CreateUser(tenantID uuid.UUID, scimUser *SCIMUser) (*SCIMUser, error) {
	// Validate required fields
	if scimUser.UserName == "" {
		return nil, NewSCIMError(400, "userName is required").ToError()
	}

	// Check if user already exists
	existingUser, _ := s.userRepo.GetUserByEmail(scimUser.UserName)
	if existingUser != nil && existingUser.TenantID == tenantID {
		return nil, NewSCIMError(409, "user already exists").ToError()
	}

	// Determine email and name from SCIM data
	email := scimUser.UserName
	name := scimUser.DisplayName
	if name == "" && scimUser.Name.GivenName != "" {
		name = scimUser.Name.GivenName
		if scimUser.Name.FamilyName != "" {
			name += " " + scimUser.Name.FamilyName
		}
	}

	// Override with primary email if different
	if len(scimUser.Emails) > 0 {
		for _, em := range scimUser.Emails {
			if em.Primary || em.Type == "work" {
				email = em.Value
				break
			}
		}
		if email == scimUser.UserName && len(scimUser.Emails) > 0 {
			email = scimUser.Emails[0].Value
		}
	}

	// Create user with empty password (SCIM-provisioned users will need password reset)
	user, err := s.userRepo.CreateUser(email, "", tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Update user name if different from email
	if name != "" && name != email {
		updates := map[string]interface{}{
			"name": name,
		}
		_, err = s.userRepo.UpdateUser(context.Background(), user.ID, updates)
		if err != nil {
			s.logger.WithError(err).Warnf("failed to update user name")
		}
		user, _ = s.userRepo.GetUserByID(user.ID)
	}

	// Log the sync
	s.logSync(tenantID, "User", user.ID.String(), "create", true, "")

	return s.userToSCIM(user)
}

// UpdateUser updates an existing user from SCIM request
func (s *SCIMService) UpdateUser(tenantID, userID uuid.UUID, scimUser *SCIMUser) (*SCIMUser, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.TenantID != tenantID {
		return nil, fmt.Errorf("user not found")
	}

	// Build updates map
	updates := make(map[string]interface{})

	// Update name
	if scimUser.Name.GivenName != "" || scimUser.Name.FamilyName != "" {
		if scimUser.Name.Formatted != "" {
			updates["name"] = scimUser.Name.Formatted
		} else {
			name := scimUser.Name.GivenName
			if scimUser.Name.FamilyName != "" {
				name += " " + scimUser.Name.FamilyName
			}
			updates["name"] = name
		}
	}

	// Update display name (overrides name)
	if scimUser.DisplayName != "" {
		updates["name"] = scimUser.DisplayName
	}

	// Update email if different
	if len(scimUser.Emails) > 0 {
		for _, em := range scimUser.Emails {
			if em.Primary || em.Type == "work" {
				if em.Value != user.Email {
					updates["email"] = em.Value
				}
				break
			}
		}
		if _, ok := updates["email"]; !ok && len(scimUser.Emails) > 0 {
			newEmail := scimUser.Emails[0].Value
			if newEmail != user.Email {
				updates["email"] = newEmail
			}
		}
	}

	if len(updates) > 0 {
		user, err = s.userRepo.UpdateUser(context.Background(), userID, updates)
		if err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Log the sync
	s.logSync(tenantID, "User", user.ID.String(), "update", true, "")

	return s.userToSCIM(user)
}

// PatchUser applies a PATCH operation to a user
func (s *SCIMService) PatchUser(tenantID, userID uuid.UUID, operations []SCIMPatchOperation) (*SCIMUser, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.TenantID != tenantID {
		return nil, fmt.Errorf("user not found")
	}

	// Build updates map
	updates := make(map[string]interface{})

	for _, op := range operations {
		switch op.Op {
		case "replace":
			var replacement SCIMUser
			if err := json.Unmarshal(op.Value, &replacement); err != nil {
				return nil, NewSCIMError(400, "invalid patch value").ToError()
			}
			if replacement.Name.GivenName != "" || replacement.Name.FamilyName != "" {
				if replacement.Name.Formatted != "" {
					updates["name"] = replacement.Name.Formatted
				} else {
					name := replacement.Name.GivenName
					if replacement.Name.FamilyName != "" {
						name += " " + replacement.Name.FamilyName
					}
					updates["name"] = name
				}
			}
			if replacement.DisplayName != "" {
				updates["name"] = replacement.DisplayName
			}
			if len(replacement.Emails) > 0 {
				for _, em := range replacement.Emails {
					if em.Primary || em.Type == "work" {
						updates["email"] = em.Value
						break
					}
				}
			}
		case "add":
			var addition SCIMUser
			if err := json.Unmarshal(op.Value, &addition); err != nil {
				return nil, NewSCIMError(400, "invalid patch value").ToError()
			}
			if addition.Name.GivenName != "" && user.Name == "" {
				name := addition.Name.GivenName
				if addition.Name.FamilyName != "" {
					name += " " + addition.Name.FamilyName
				}
				updates["name"] = name
			}
		case "remove":
			// Handle remove - typically for removing attributes
			// For now, we don't support removing core fields
		}
	}

	if len(updates) > 0 {
		user, err = s.userRepo.UpdateUser(context.Background(), userID, updates)
		if err != nil {
			return nil, fmt.Errorf("failed to patch user: %w", err)
		}
	}

	// Log the sync
	s.logSync(tenantID, "User", user.ID.String(), "update", true, "")

	return s.userToSCIM(user)
}

// DeleteUser deletes a user with proper cascade deletion of related records
func (s *SCIMService) DeleteUser(tenantID, userID uuid.UUID) error {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if user.TenantID != tenantID {
		return fmt.Errorf("user not found")
	}

	// Use a transaction for cascade deletion to ensure atomicity
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Delete user sessions first
		if err := tx.Exec("DELETE FROM sessions WHERE user_id = $1", userID).Error; err != nil {
			return fmt.Errorf("failed to delete user sessions: %w", err)
		}

		// Delete user audit logs (actor references)
		if err := tx.Exec("DELETE FROM audit_events WHERE actor_user_id = $1", userID).Error; err != nil {
			return fmt.Errorf("failed to delete user audit logs: %w", err)
		}

		// Delete user executions
		if err := tx.Exec("DELETE FROM registry_function_executions WHERE user_id = $1", userID).Error; err != nil {
			return fmt.Errorf("failed to delete user executions: %w", err)
		}

		// Delete user MFA credentials (totp and backup codes)
		if err := tx.Exec("DELETE FROM mfa_totp WHERE user_id = $1", userID).Error; err != nil {
			return fmt.Errorf("failed to delete user MFA TOTP: %w", err)
		}
		if err := tx.Exec("DELETE FROM mfa_backup_codes WHERE user_id = $1", userID).Error; err != nil {
			return fmt.Errorf("failed to delete user backup codes: %w", err)
		}

		// Delete user login attempts history
		if err := tx.Exec("DELETE FROM login_attempts WHERE user_id = $1", userID).Error; err != nil {
			return fmt.Errorf("failed to delete user login attempts: %w", err)
		}

		// Delete the user itself
		if err := tx.Exec("DELETE FROM users WHERE id = $1", userID).Error; err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		return nil
	})

	if err != nil {
		s.logger.WithError(err).Errorf("Failed to delete user %s with cascade", userID)
		return err
	}

	// Log the sync
	s.logSync(tenantID, "User", userID.String(), "delete", true, "")

	return nil
}

// GetGroup retrieves a group by ID for a tenant
func (s *SCIMService) GetGroup(tenantID, groupID uuid.UUID) (*SCIMGroup, error) {
	group, err := s.teamRepo.GetTeamByID(groupID)
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	// Verify group belongs to tenant
	if group.TenantID != tenantID {
		return nil, fmt.Errorf("group not found")
	}

	return s.groupToSCIM(group)
}

// ListGroups lists groups for a tenant with pagination
func (s *SCIMService) ListGroups(tenantID uuid.UUID, startIndex, count int) (*SCIMListResponse, error) {
	if startIndex < 1 {
		startIndex = 1
	}
	if count < 1 {
		count = 100
	}
	if count > 1000 {
		count = 1000
	}

	groups, err := s.teamRepo.GetTeamsByTenantID(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	total := len(groups)
	startIdx := startIndex - 1
	endIdx := startIdx + count
	if startIdx >= len(groups) {
		groups = []*storage.Team{}
	} else {
		if endIdx > len(groups) {
			endIdx = len(groups)
		}
		groups = groups[startIdx:endIdx]
	}

	scimGroups := make([]SCIMGroup, len(groups))
	for i, group := range groups {
		scimGroup, err := s.groupToSCIM(group)
		if err != nil {
			s.logger.WithError(err).Warnf("failed to convert group %s to SCIM", group.ID)
			continue
		}
		scimGroups[i] = *scimGroup
	}

	return &SCIMListResponse{
		TotalResults: total,
		ItemsPerPage: count,
		StartIndex:   startIndex,
		Resources:    scimGroups,
	}, nil
}

// CreateGroup creates a new group from SCIM request
func (s *SCIMService) CreateGroup(tenantID uuid.UUID, scimGroup *SCIMGroup) (*SCIMGroup, error) {
	// Validate required fields
	if scimGroup.DisplayName == "" {
		return nil, NewSCIMError(400, "displayName is required").ToError()
	}

	// Create team (group) from SCIM data
	team := &storage.Team{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        scimGroup.DisplayName,
		Description: "",
		CreatedBy:   tenantID, // Use tenant ID as created by for SCIM-provisioned teams
	}

	err := s.teamRepo.CreateTeam(team)
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	// Add members if specified
	for _, member := range scimGroup.Members {
		memberID, err := uuid.Parse(member.Value)
		if err != nil {
			continue
		}
		membership := &storage.TeamMembership{
			ID:      uuid.New(),
			TeamID:  team.ID,
			UserID:  memberID,
			Role:    "member",
			AddedBy: tenantID,
		}
		_ = s.teamRepo.AddTeamMember(membership)
	}

	// Log the sync
	s.logSync(tenantID, "Group", team.ID.String(), "create", true, "")

	return s.groupToSCIM(team)
}

// UpdateGroup updates an existing group from SCIM request
func (s *SCIMService) UpdateGroup(tenantID, groupID uuid.UUID, scimGroup *SCIMGroup) (*SCIMGroup, error) {
	group, err := s.teamRepo.GetTeamByID(groupID)
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	if group.TenantID != tenantID {
		return nil, fmt.Errorf("group not found")
	}

	// Update group fields from SCIM
	if scimGroup.DisplayName != "" {
		group.Name = scimGroup.DisplayName
	}

	err = s.teamRepo.UpdateTeam(group)
	if err != nil {
		return nil, fmt.Errorf("failed to update group: %w", err)
	}

	// Log the sync
	s.logSync(tenantID, "Group", group.ID.String(), "update", true, "")

	return s.groupToSCIM(group)
}

// PatchGroup applies a PATCH operation to a group
func (s *SCIMService) PatchGroup(tenantID, groupID uuid.UUID, operations []SCIMPatchOperation) (*SCIMGroup, error) {
	group, err := s.teamRepo.GetTeamByID(groupID)
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	if group.TenantID != tenantID {
		return nil, fmt.Errorf("group not found")
	}

	for _, op := range operations {
		switch op.Op {
		case "replace":
			var replacement SCIMGroup
			if err := json.Unmarshal(op.Value, &replacement); err != nil {
				return nil, NewSCIMError(400, "invalid patch value").ToError()
			}
			if replacement.DisplayName != "" {
				group.Name = replacement.DisplayName
			}
		case "add":
			// Handle adding members
			if strings.Contains(op.Path, "members") {
				var members []SCIMMemberRef
				if err := json.Unmarshal(op.Value, &members); err != nil {
					return nil, NewSCIMError(400, "invalid patch value").ToError()
				}
				for _, member := range members {
					memberID, err := uuid.Parse(member.Value)
					if err != nil {
						continue
					}
					// Check if already a member
					_, err = s.teamRepo.GetTeamMembership(groupID, memberID)
					if err != nil {
						membership := &storage.TeamMembership{
							ID:      uuid.New(),
							TeamID:  group.ID,
							UserID:  memberID,
							Role:    "member",
							AddedBy: tenantID,
						}
						_ = s.teamRepo.AddTeamMember(membership)
					}
				}
			}
		case "remove":
			// Handle removing members
			if strings.Contains(op.Path, "members") {
				var members []SCIMMemberRef
				if err := json.Unmarshal(op.Value, &members); err != nil {
					return nil, NewSCIMError(400, "invalid patch value").ToError()
				}
				for _, member := range members {
					memberID, err := uuid.Parse(member.Value)
					if err != nil {
						continue
					}
					_ = s.teamRepo.RemoveTeamMember(groupID, memberID)
				}
			}
		}
	}

	err = s.teamRepo.UpdateTeam(group)
	if err != nil {
		return nil, fmt.Errorf("failed to patch group: %w", err)
	}

	// Log the sync
	s.logSync(tenantID, "Group", group.ID.String(), "update", true, "")

	return s.groupToSCIM(group)
}

// DeleteGroup deletes a group
func (s *SCIMService) DeleteGroup(tenantID, groupID uuid.UUID) error {
	group, err := s.teamRepo.GetTeamByID(groupID)
	if err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	if group.TenantID != tenantID {
		return fmt.Errorf("group not found")
	}

	err = s.teamRepo.DeleteTeam(groupID)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	// Log the sync
	s.logSync(tenantID, "Group", groupID.String(), "delete", true, "")

	return nil
}

// GetConfig retrieves SCIM config for a tenant
func (s *SCIMService) GetConfig(tenantID uuid.UUID) (*storage.SCIMConfig, error) {
	return s.configRepo.GetByTenantID(tenantID)
}

// UpdateConfig updates SCIM config for a tenant
func (s *SCIMService) UpdateConfig(tenantID uuid.UUID, config *storage.SCIMConfig) error {
	existing, err := s.configRepo.GetByTenantID(tenantID)
	if err != nil {
		// Create new config
		config.TenantID = tenantID
		config.ID = uuid.New()
		config.CreatedAt = time.Now()
		config.UpdatedAt = time.Now()
		return s.configRepo.Create(config)
	}

	// Update existing
	config.ID = existing.ID
	config.TenantID = tenantID
	config.CreatedAt = existing.CreatedAt
	config.UpdatedAt = time.Now()
	return s.configRepo.Update(config)
}

// EnableSCIM enables SCIM for a tenant
func (s *SCIMService) EnableSCIM(tenantID uuid.UUID, idpURL, idpToken string, secretKey []byte) error {
	return s.configRepo.Enable(tenantID, idpURL, idpToken, secretKey)
}

// DisableSCIM disables SCIM for a tenant
func (s *SCIMService) DisableSCIM(tenantID uuid.UUID) error {
	return s.configRepo.Disable(tenantID)
}

// userToSCIM converts a storage.User to SCIMUser
func (s *SCIMService) userToSCIM(user *storage.User) (*SCIMUser, error) {
	scimUser := &SCIMUser{
		ID:       user.ID.String(),
		UserName: user.Email,
		Name: SCIMName{
			GivenName:  user.Name,
			FamilyName: "",
			Formatted:  user.Name,
		},
		Active: true,
		Meta: SCIMMeta{
			ResourceType: "User",
			Created:      user.CreatedAt,
			Modified:     user.UpdatedAt,
			Version:      user.ID.String(),
		},
	}

	// Add email
	if user.Email != "" {
		scimUser.Emails = []SCIMEmail{
			{Value: user.Email, Type: "work", Primary: true},
		}
	}

	// Add display name
	if user.Name != "" {
		scimUser.DisplayName = user.Name
	}

	// Get user's groups
	teams, err := s.teamRepo.GetUserTeams(user.ID)
	if err == nil && len(teams) > 0 {
		scimUser.Groups = make([]SCIMGroupRef, len(teams))
		for i, team := range teams {
			scimUser.Groups[i] = SCIMGroupRef{
				Value:   team.ID.String(),
				Display: team.Name,
				Type:    "direct",
			}
		}
	}

	return scimUser, nil
}

// groupToSCIM converts a storage.Team to SCIMGroup
func (s *SCIMService) groupToSCIM(team *storage.Team) (*SCIMGroup, error) {
	scimGroup := &SCIMGroup{
		ID:          team.ID.String(),
		DisplayName: team.Name,
		Meta: SCIMMeta{
			ResourceType: "Group",
			Created:      team.CreatedAt,
			Modified:     team.UpdatedAt,
			Version:      team.ID.String(),
		},
	}

	// Add members
	if len(team.Members) > 0 {
		scimGroup.Members = make([]SCIMMemberRef, len(team.Members))
		for i, member := range team.Members {
			display := ""
			if member.User != nil {
				display = member.User.Email
			}
			scimGroup.Members[i] = SCIMMemberRef{
				Value:   member.UserID.String(),
				Display: display,
				Type:    "User",
			}
		}
	}

	return scimGroup, nil
}

// logSync logs a SCIM sync operation
func (s *SCIMService) logSync(tenantID uuid.UUID, resourceType, resourceID, action string, success bool, errorMsg string) {
	log := &storage.SCIMSyncLog{
		ID:           uuid.New(),
		TenantID:     tenantID,
		Direction:    "inbound",
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       action,
		Success:      success,
		CreatedAt:    time.Now(),
	}
	if errorMsg != "" {
		log.ErrorMessage = &errorMsg
	}

	_ = s.syncLogRepo.Create(log)
}
