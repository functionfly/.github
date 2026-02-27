package storage

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// TeamRepository handles team-related database operations
type TeamRepository struct {
	db *gorm.DB
}

// NewTeamRepository creates a new team repository
func NewTeamRepository(db *gorm.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

// CreateTeam creates a new team
func (r *TeamRepository) CreateTeam(team *Team) error {
	return r.db.Create(team).Error
}

// GetTeamByID gets a team by ID
func (r *TeamRepository) GetTeamByID(teamID uuid.UUID) (*Team, error) {
	var team Team
	err := r.db.Preload("Members").Preload("Members.User").Preload("Permissions").First(&team, "id = ?", teamID).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// GetTeamsByTenantID gets all teams for a tenant
func (r *TeamRepository) GetTeamsByTenantID(tenantID uuid.UUID) ([]*Team, error) {
	var teams []*Team
	err := r.db.Preload("Members").Preload("Members.User").Where("tenant_id = ?", tenantID).Find(&teams).Error
	return teams, err
}

// UpdateTeam updates a team
func (r *TeamRepository) UpdateTeam(team *Team) error {
	return r.db.Save(team).Error
}

// DeleteTeam deletes a team
func (r *TeamRepository) DeleteTeam(teamID uuid.UUID) error {
	return r.db.Delete(&Team{}, "id = ?", teamID).Error
}

// AddTeamMember adds a user to a team
func (r *TeamRepository) AddTeamMember(membership *TeamMembership) error {
	return r.db.Create(membership).Error
}

// UpdateTeamMember updates a team member's role
func (r *TeamRepository) UpdateTeamMember(teamID, userID uuid.UUID, role string) error {
	return r.db.Model(&TeamMembership{}).Where("team_id = ? AND user_id = ?", teamID, userID).Update("role", role).Error
}

// RemoveTeamMember removes a user from a team
func (r *TeamRepository) RemoveTeamMember(teamID, userID uuid.UUID) error {
	return r.db.Delete(&TeamMembership{}, "team_id = ? AND user_id = ?", teamID, userID).Error
}

// GetTeamMembership gets a user's membership in a team
func (r *TeamRepository) GetTeamMembership(teamID, userID uuid.UUID) (*TeamMembership, error) {
	var membership TeamMembership
	err := r.db.Where("team_id = ? AND user_id = ?", teamID, userID).First(&membership).Error
	if err != nil {
		return nil, err
	}
	return &membership, nil
}

// GetUserTeams gets all teams a user belongs to
func (r *TeamRepository) GetUserTeams(userID uuid.UUID) ([]*Team, error) {
	var memberships []TeamMembership
	err := r.db.Preload("Team").Where("user_id = ?", userID).Find(&memberships).Error
	if err != nil {
		return nil, err
	}

	teams := make([]*Team, len(memberships))
	for i, membership := range memberships {
		teams[i] = membership.Team
	}
	return teams, nil
}

// GrantTeamPermission grants permissions to a team for a resource
func (r *TeamRepository) GrantTeamPermission(permission *TeamPermission) error {
	return r.db.Create(permission).Error
}

// RevokeTeamPermission revokes permissions from a team for a resource
func (r *TeamRepository) RevokeTeamPermission(teamID uuid.UUID, resourceType string, resourceID uuid.UUID) error {
	return r.db.Delete(&TeamPermission{}, "team_id = ? AND resource_type = ? AND resource_id = ?", teamID, resourceType, resourceID).Error
}

// GetTeamPermissions gets all permissions for a team
func (r *TeamRepository) GetTeamPermissions(teamID uuid.UUID) ([]*TeamPermission, error) {
	var permissions []*TeamPermission
	err := r.db.Where("team_id = ?", teamID).Find(&permissions).Error
	return permissions, err
}

// GetResourcePermissions gets permissions for a specific resource across all teams
func (r *TeamRepository) GetResourcePermissions(resourceType string, resourceID uuid.UUID) ([]*TeamPermission, error) {
	var permissions []*TeamPermission
	err := r.db.Preload("Team").Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).Find(&permissions).Error
	return permissions, err
}

// CheckUserResourcePermission checks if a user has permission for a resource through their teams
func (r *TeamRepository) CheckUserResourcePermission(userID uuid.UUID, resourceType string, resourceID uuid.UUID, requiredPerm string) (bool, error) {
	var count int64
	err := r.db.Table("team_memberships tm").
		Joins("JOIN team_permissions tp ON tm.team_id = tp.team_id").
		Where("tm.user_id = ? AND tp.resource_type = ? AND tp.resource_id = ? AND ? = ANY(string_to_array(tp.permissions, ','))",
			userID, resourceType, resourceID, requiredPerm).
		Count(&count).Error

	return count > 0, err
}

// GetUserPermissions gets all permissions a user has for a specific resource type
func (r *TeamRepository) GetUserPermissions(userID uuid.UUID, resourceType string) ([]string, error) {
	var permissions []string
	err := r.db.Table("team_memberships tm").
		Joins("JOIN team_permissions tp ON tm.team_id = tp.team_id").
		Where("tm.user_id = ? AND tp.resource_type = ?", userID, resourceType).
		Pluck("DISTINCT unnest(string_to_array(tp.permissions, ','))", &permissions).Error

	return permissions, err
}

// IsUserTeamOwner checks if a user is the owner of a team
func (r *TeamRepository) IsUserTeamOwner(userID, teamID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&TeamMembership{}).Where("team_id = ? AND user_id = ? AND role = ?", teamID, userID, "owner").Count(&count).Error
	return count > 0, err
}

// IsUserTeamAdmin checks if a user is an admin or owner of a team
func (r *TeamRepository) IsUserTeamAdmin(userID, teamID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&TeamMembership{}).Where("team_id = ? AND user_id = ? AND role IN ?", teamID, userID, pq.Array([]string{"owner", "admin"})).Count(&count).Error
	return count > 0, err
}