package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PostgresDB methods: function registry, deployments, logs, dashboard aggregates, teams.

// Function operations
func (db *PostgresDB) CreateFunction(ctx context.Context, function *FunctionConfig) (*FunctionConfig, error) {
	return db.functionRepository.CreateFunction(ctx, function)
}

func (db *PostgresDB) GetFunctionByID(ctx context.Context, functionID uuid.UUID) (*FunctionConfig, error) {
	return db.functionRepository.GetFunctionByID(ctx, functionID)
}

func (db *PostgresDB) ListFunctionsByTenant(ctx context.Context, tenantID uuid.UUID) ([]*FunctionConfig, error) {
	return db.functionRepository.ListFunctionsByTenant(ctx, tenantID)
}

func (db *PostgresDB) ListAllFunctions(ctx context.Context, limit, offset int, tenantID *uuid.UUID, status *string) ([]*FunctionConfig, int, error) {
	return db.functionRepository.ListAllFunctions(ctx, limit, offset, tenantID, status)
}

func (db *PostgresDB) UpdateFunction(ctx context.Context, functionID uuid.UUID, updates map[string]interface{}) (*FunctionConfig, error) {
	return db.functionRepository.UpdateFunction(ctx, functionID, updates)
}

func (db *PostgresDB) DeleteFunction(ctx context.Context, functionID uuid.UUID) error {
	return db.functionRepository.DeleteFunction(ctx, functionID)
}

func (db *PostgresDB) GetFunctionByAppIDAndName(ctx context.Context, appID uuid.UUID, name string) (*FunctionConfig, error) {
	return db.functionRepository.GetFunctionByAppIDAndName(ctx, appID, name)
}

func (db *PostgresDB) GetActiveDeploymentForFunction(ctx context.Context, functionID uuid.UUID) (*FunctionDeployment, error) {
	return db.functionRepository.GetActiveDeploymentForFunction(ctx, functionID)
}

// Function deployment operations
func (db *PostgresDB) CreateFunctionDeployment(ctx context.Context, deployment *FunctionDeployment) (*FunctionDeployment, error) {
	return db.functionRepository.CreateFunctionDeployment(ctx, deployment)
}

func (db *PostgresDB) GetFunctionDeploymentByID(ctx context.Context, deploymentID uuid.UUID) (*FunctionDeployment, error) {
	return db.functionRepository.GetFunctionDeploymentByID(ctx, deploymentID)
}

func (db *PostgresDB) ListFunctionDeployments(ctx context.Context, functionID uuid.UUID, limit int) ([]*FunctionDeployment, error) {
	return db.functionRepository.ListFunctionDeployments(ctx, functionID, limit)
}

func (db *PostgresDB) UpdateFunctionDeploymentStatus(ctx context.Context, deploymentID uuid.UUID, status string, deployedURL, errorMessage *string) error {
	return db.functionRepository.UpdateFunctionDeploymentStatus(ctx, deploymentID, status, deployedURL, errorMessage)
}

// Function log operations
func (db *PostgresDB) CreateFunctionLog(ctx context.Context, log *FunctionLog) error {
	return db.functionRepository.CreateFunctionLog(ctx, log)
}

func (db *PostgresDB) GetFunctionLogs(ctx context.Context, functionID *uuid.UUID, deploymentID *uuid.UUID, limit int, since *time.Time, level *string) ([]*FunctionLog, error) {
	return db.functionRepository.GetFunctionLogs(ctx, functionID, deploymentID, limit, since, level)
}

func (db *PostgresDB) DeleteFunctionLogsOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	return db.functionRepository.DeleteFunctionLogsOlderThan(ctx, cutoff)
}

// Dashboard aggregations
func (db *PostgresDB) GetUsageByDay(ctx context.Context, tenantID uuid.UUID, days int) ([]UsageByDay, error) {
	return db.functionRepository.GetUsageByDay(ctx, tenantID, days)
}

func (db *PostgresDB) GetExecutionRateByHour(ctx context.Context, tenantID uuid.UUID, hours int) ([]ExecutionRateByHour, error) {
	return db.functionRepository.GetExecutionRateByHour(ctx, tenantID, hours)
}

func (db *PostgresDB) GetRecentActivityForTenant(ctx context.Context, tenantID uuid.UUID, limit int) ([]DashboardActivityItem, error) {
	return db.functionRepository.GetRecentActivityForTenant(ctx, tenantID, limit)
}

func (db *PostgresDB) GetDashboardMetrics(ctx context.Context, tenantID uuid.UUID) (*DashboardMetrics, error) {
	return db.functionRepository.GetDashboardMetrics(ctx, tenantID)
}

// Team operations
func (db *PostgresDB) CreateTeam(ctx context.Context, team *Team) error {
	return db.teamRepository.CreateTeam(ctx, team)
}

func (db *PostgresDB) GetTeamByID(ctx context.Context, teamID uuid.UUID) (*Team, error) {
	return db.teamRepository.GetTeamByID(ctx, teamID)
}

func (db *PostgresDB) GetTeamsByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*Team, error) {
	return db.teamRepository.GetTeamsByTenantID(ctx, tenantID)
}

func (db *PostgresDB) UpdateTeam(ctx context.Context, team *Team) error {
	return db.teamRepository.UpdateTeam(ctx, team)
}

func (db *PostgresDB) DeleteTeam(ctx context.Context, teamID uuid.UUID) error {
	return db.teamRepository.DeleteTeam(ctx, teamID)
}

func (db *PostgresDB) AddTeamMember(ctx context.Context, membership *TeamMembership) error {
	return db.teamRepository.AddTeamMember(ctx, membership)
}

func (db *PostgresDB) UpdateTeamMember(ctx context.Context, teamID, userID uuid.UUID, role string) error {
	return db.teamRepository.UpdateTeamMember(ctx, teamID, userID, role)
}

func (db *PostgresDB) RemoveTeamMember(ctx context.Context, teamID, userID uuid.UUID) error {
	return db.teamRepository.RemoveTeamMember(ctx, teamID, userID)
}

func (db *PostgresDB) GetTeamMembership(ctx context.Context, teamID, userID uuid.UUID) (*TeamMembership, error) {
	return db.teamRepository.GetTeamMembership(ctx, teamID, userID)
}

func (db *PostgresDB) GetUserTeams(ctx context.Context, userID uuid.UUID) ([]*Team, error) {
	return db.teamRepository.GetUserTeams(ctx, userID)
}

func (db *PostgresDB) GrantTeamPermission(ctx context.Context, permission *TeamPermission) error {
	return db.teamRepository.GrantTeamPermission(ctx, permission)
}

func (db *PostgresDB) RevokeTeamPermission(ctx context.Context, teamID uuid.UUID, resourceType string, resourceID uuid.UUID) error {
	return db.teamRepository.RevokeTeamPermission(ctx, teamID, resourceType, resourceID)
}

func (db *PostgresDB) GetTeamPermissions(ctx context.Context, teamID uuid.UUID) ([]*TeamPermission, error) {
	return db.teamRepository.GetTeamPermissions(ctx, teamID)
}

func (db *PostgresDB) GetResourcePermissions(ctx context.Context, resourceType string, resourceID uuid.UUID) ([]*TeamPermission, error) {
	return db.teamRepository.GetResourcePermissions(ctx, resourceType, resourceID)
}

func (db *PostgresDB) CheckUserResourcePermission(ctx context.Context, userID uuid.UUID, resourceType string, resourceID uuid.UUID, requiredPerm string) (bool, error) {
	return db.teamRepository.CheckUserResourcePermission(ctx, userID, resourceType, resourceID, requiredPerm)
}

func (db *PostgresDB) GetUserPermissions(ctx context.Context, userID uuid.UUID, resourceType string) ([]string, error) {
	return db.teamRepository.GetUserPermissions(ctx, userID, resourceType)
}

func (db *PostgresDB) IsUserTeamOwner(ctx context.Context, userID, teamID uuid.UUID) (bool, error) {
	return db.teamRepository.IsUserTeamOwner(ctx, userID, teamID)
}

func (db *PostgresDB) IsUserTeamAdmin(ctx context.Context, userID, teamID uuid.UUID) (bool, error) {
	return db.teamRepository.IsUserTeamAdmin(ctx, userID, teamID)
}
