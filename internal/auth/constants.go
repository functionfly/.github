package auth

const (
	// JWT claims
	Issuer = "functionfly"
	// Password hashing parameters
	saltLength = 32
	keyLength  = 32
	timeCost   = 1
	memoryCost = 64 * 1024
	threads    = 4
)

// Admin roles - platform level permissions
const (
	RoleSuperAdmin     = "super_admin"
	RoleAdmin          = "admin" // Treated same as super_admin for RBAC; MFA and other code already treat "admin" as admin
	RoleSupport        = "support"
	RoleBillingAdmin   = "billing_admin"
	RoleDeveloperAdmin = "developer_admin"
	RoleReadOnly       = "read_only"
)

// Admin permissions - explicit string permissions
const (
	PermTenantsRead      = "tenants.read"
	PermTenantsWrite     = "tenants.write"
	PermUsersRead        = "users.read"
	PermUsersWrite       = "users.write"
	PermBillingRead      = "billing.read"
	PermBillingWrite     = "billing.write"
	PermDeploymentsRead  = "deployments.read"
	PermDeploymentsWrite = "deployments.write"
	PermAuditRead        = "audit.read"
	PermSystemRead       = "system.read"
	PermSystemWrite      = "system.write"
)

// Team roles
const (
	TeamRoleOwner  = "owner"
	TeamRoleAdmin  = "admin"
	TeamRoleMember = "member"
	TeamRoleViewer = "viewer"
)

// Resource permissions
const (
	PermAppsRead          = "apps.read"
	PermAppsWrite         = "apps.write"
	PermAppsDelete        = "apps.delete"
	PermFunctionsRead     = "functions.read"
	PermFunctionsWrite    = "functions.write"
	PermFunctionsDelete   = "functions.delete"
	PermBackendsRead      = "backends.read"
	PermBackendsWrite     = "backends.write"
	PermBackendsDelete    = "backends.delete"
	PermDeploymentsDelete = "deployments.delete"
	PermTeamsRead         = "teams.read"
	PermTeamsWrite        = "teams.write"
	PermTeamsDelete       = "teams.delete"

	// StateFabric permissions
	PermStateRead         = "state.read"
	PermStateWrite        = "state.write"
	PermStateDelete       = "state.delete"
	PermStateAdmin        = "state.admin"
	PermTriggersManage    = "triggers.manage"
	PermSnapshotsCreate   = "snapshots.create"
	PermSnapshotsRestore  = "snapshots.restore"
	PermReplayAccess      = "replay.access"
	PermMemoryRead        = "memory.read"
	PermMemoryWrite       = "memory.write"

	// Function Registry permissions
	PermRegistryPublish   = "registry.publish"
	PermRegistryVerify    = "registry.verify"
	PermRegistryApprove   = "registry.approve"
	PermRegistrySign      = "registry.sign"
	PermRegistryManage    = "registry.manage"

	// Monitoring & Observability permissions
	PermMonitoringAlerts  = "monitoring.alerts"
	PermMonitoringManage  = "monitoring.manage"
	PermMonitoringMetrics = "monitoring.metrics"
	PermMonitoringAdmin   = "monitoring.admin"
	PermMonitoringHealth  = "monitoring.health"

	// Security Operations permissions
	PermSecurityIncidents = "security.incidents"
	PermSecurityScans     = "security.scans"
	PermSecurityAudit     = "security.audit"
	PermSecurityAdmin     = "security.admin"

	// Content Management permissions
	PermContentCreate     = "content.create"
	PermContentEdit       = "content.edit"
	PermContentPublish    = "content.publish"
	PermContentManage     = "content.manage"
	PermChangelogManage   = "changelog.manage"
	PermBlogManage        = "blog.manage"

	// Granular Team Management permissions
	PermTeamMembersManage = "team.members.manage"
	PermTeamRolesAssign   = "team.roles.assign"
	PermTeamResourcesShare = "team.resources.share"

	// Function Verification permissions
	PermVerificationApprove = "verification.approve"
	PermVerificationSign    = "verification.sign"
	PermVerificationOverride = "verification.override"

	// Feedback Management permissions
	PermFeedbackModerate    = "feedback.moderate"
	PermFeedbackAnalytics   = "feedback.analytics"
)
