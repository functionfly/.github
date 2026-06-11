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

// GetTeamByID gets a team by ID.
// Uses batch queries instead of GORM's chained Preloads to avoid the N+1
// query problem (each Preload fires a separate sequential query, compounding
// latency with nested relations like Members → User).
func (r *TeamRepository) GetTeamByID(teamID uuid.UUID) (*Team, error) {
	var team Team
	if err := r.db.Where("id = ?", teamID).First(&team).Error; err != nil {
		return nil, err
	}

	// Batch-load all memberships for this team.
	var memberships []TeamMembership
	if err := r.db.Where("team_id = ?", teamID).Find(&memberships).Error; err != nil {
		return nil, err
	}
	if len(memberships) == 0 {
		team.Members = []TeamMembership{}
		team.Permissions = []TeamPermission{}
		return &team, nil
	}

	// Collect all user IDs referenced by memberships.
	userIDs := make([]uuid.UUID, 0, len(memberships))
	userSet := make(map[uuid.UUID]struct{}, len(memberships))
	for _, m := range memberships {
		if _, ok := userSet[m.UserID]; !ok {
			userIDs = append(userIDs, m.UserID)
			userSet[m.UserID] = struct{}{}
		}
	}

	// Batch-load all users in one query.
	users := make(map[uuid.UUID]*User, len(userIDs))
	if len(userIDs) > 0 {
		var userRecords []User
		if err := r.db.Where("id IN ?", userIDs).Find(&userRecords).Error; err != nil {
			return nil, err
		}
		for i := range userRecords {
			users[userRecords[i].ID] = &userRecords[i]
		}
	}

	// Batch-load all permissions for this team in one query.
	var permissions []TeamPermission
	if err := r.db.Where("team_id = ?", teamID).Find(&permissions).Error; err != nil {
		return nil, err
	}

	// Wire up memberships → user and assign to team.
	members := make([]TeamMembership, len(memberships))
	for i, m := range memberships {
		m := m // local copy for closure
		if u, ok := users[m.UserID]; ok {
			m.User = u
		}
		members[i] = m
	}
	team.Members = members
	team.Permissions = permissions

	return &team, nil
}

// GetTeamsByTenantID gets all teams for a tenant.
// Uses batch queries instead of GORM's chained Preloads to avoid the N+1
// query problem (each Preload fires a separate sequential query, compounding
// latency with nested relations like Members → User).
func (r *TeamRepository) GetTeamsByTenantID(tenantID uuid.UUID) ([]*Team, error) {
	var teams []*Team
	if err := r.db.Where("tenant_id = ?", tenantID).Find(&teams).Error; err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return teams, nil
	}

	// Collect all team IDs for batch lookups.
	teamIDs := make([]uuid.UUID, len(teams))
	teamMap := make(map[uuid.UUID]*Team, len(teams))
	for i, t := range teams {
		teamIDs[i] = t.ID
		teamMap[t.ID] = t
	}

	// Batch-load all memberships for these teams in one query.
	var memberships []TeamMembership
	if err := r.db.Where("team_id IN ?", teamIDs).Find(&memberships).Error; err != nil {
		return nil, err
	}

	// Collect all user IDs referenced by memberships.
	userIDs := make([]uuid.UUID, 0, len(memberships))
	userSet := make(map[uuid.UUID]struct{}, len(memberships))
	for _, m := range memberships {
		if _, ok := userSet[m.UserID]; !ok {
			userIDs = append(userIDs, m.UserID)
			userSet[m.UserID] = struct{}{}
		}
	}

	// Batch-load all users in one query.
	users := make(map[uuid.UUID]*User, len(userIDs))
	if len(userIDs) > 0 {
		var userRecords []User
		if err := r.db.Where("id IN ?", userIDs).Find(&userRecords).Error; err != nil {
			return nil, err
		}
		for i := range userRecords {
			users[userRecords[i].ID] = &userRecords[i]
		}
	}

	// Batch-load all permissions for these teams in one query.
	var permissions []TeamPermission
	if err := r.db.Where("team_id IN ?", teamIDs).Find(&permissions).Error; err != nil {
		return nil, err
	}

	// Wire up memberships → user and assign to teams.
	memberMap := make(map[uuid.UUID][]TeamMembership, len(teams))
	for _, m := range memberships {
		m := m // local copy for closure
		if u, ok := users[m.UserID]; ok {
			m.User = u
		}
		memberMap[m.TeamID] = append(memberMap[m.TeamID], m)
	}
	for id, team := range teamMap {
		team.Members = memberMap[id]
	}

	// Wire up permissions to teams.
	permMap := make(map[uuid.UUID][]TeamPermission, len(teams))
	for _, p := range permissions {
		p := p
		permMap[p.TeamID] = append(permMap[p.TeamID], p)
	}
	for id, team := range teamMap {
		team.Permissions = permMap[id]
	}

	return teams, nil
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
