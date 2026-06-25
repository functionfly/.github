package api

import (
	"github.com/functionfly/functionfly/internal/api/handlers/employee"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

// registerEmployeeRoutes registers all FWOS employee-related API routes.
func registerEmployeeRoutes(
	api *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	csrfMiddleware *middleware.CSRFMiddleware,
	handler *employee.Handler,
) {
	// ── Employee CRUD ──────────────────────────────────────────────────────
	api.HandleFunc("/employees", authMiddleware.RequireAuth(handler.HandleListEmployees)).Methods("GET", "OPTIONS")
	api.HandleFunc("/employees/{id}", authMiddleware.RequireAuth(handler.HandleGetEmployee)).Methods("GET", "OPTIONS")
	api.HandleFunc("/employees", authMiddleware.RequireAuth(handler.HandleCreateEmployee)).Methods("POST", "OPTIONS")
	api.HandleFunc("/employees/{id}", authMiddleware.RequireAuth(handler.HandleUpdateEmployee)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/employees/{id}/generate-access", authMiddleware.RequireAuth(handler.HandleGenerateAccess)).Methods("POST", "OPTIONS")

	// ── Employee by FFID ──────────────────────────────────────────────────
	api.HandleFunc("/employees/ffid/{ffid}", authMiddleware.RequireAuth(handler.HandleGetByFFID)).Methods("GET", "OPTIONS")

	// ── Skills, Certifications, Achievements ──────────────────────────────
	api.HandleFunc("/employees/{id}/skills", authMiddleware.RequireAuth(handler.HandleListSkills)).Methods("GET", "OPTIONS")
	api.HandleFunc("/employees/{id}/skills", authMiddleware.RequireAuth(handler.HandleAddSkill)).Methods("POST", "OPTIONS")
	api.HandleFunc("/employees/{id}/certifications", authMiddleware.RequireAuth(handler.HandleListCertifications)).Methods("GET", "OPTIONS")
	api.HandleFunc("/employees/{id}/certifications", authMiddleware.RequireAuth(handler.HandleAddCertification)).Methods("POST", "OPTIONS")
	api.HandleFunc("/employees/{id}/achievements", authMiddleware.RequireAuth(handler.HandleListAchievements)).Methods("GET", "OPTIONS")

	// ── Projects ──────────────────────────────────────────────────────────
	api.HandleFunc("/projects", authMiddleware.RequireAuth(handler.HandleListProjects)).Methods("GET", "OPTIONS")
	api.HandleFunc("/projects/{id}", authMiddleware.RequireAuth(handler.HandleGetProject)).Methods("GET", "OPTIONS")
	api.HandleFunc("/projects", authMiddleware.RequireAuth(handler.HandleCreateProject)).Methods("POST", "OPTIONS")
	api.HandleFunc("/projects/{id}", authMiddleware.RequireAuth(handler.HandleUpdateProject)).Methods("PATCH", "OPTIONS")

	// ── Tasks ─────────────────────────────────────────────────────────────
	api.HandleFunc("/tasks", authMiddleware.RequireAuth(handler.HandleListTasks)).Methods("GET", "OPTIONS")
	api.HandleFunc("/tasks/{id}", authMiddleware.RequireAuth(handler.HandleGetTask)).Methods("GET", "OPTIONS")
	api.HandleFunc("/tasks", authMiddleware.RequireAuth(handler.HandleCreateTask)).Methods("POST", "OPTIONS")
	api.HandleFunc("/tasks/{id}", authMiddleware.RequireAuth(handler.HandleUpdateTask)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/tasks/{id}/assign", authMiddleware.RequireAuth(handler.HandleAssignTask)).Methods("POST", "OPTIONS")
	api.HandleFunc("/tasks/{id}/comments", authMiddleware.RequireAuth(handler.HandleListTaskComments)).Methods("GET", "OPTIONS")
	api.HandleFunc("/tasks/{id}/comments", authMiddleware.RequireAuth(handler.HandleCreateTaskComment)).Methods("POST", "OPTIONS")

	// ── Learning ──────────────────────────────────────────────────────────
	api.HandleFunc("/learning/courses", authMiddleware.RequireAuth(handler.HandleListCourses)).Methods("GET", "OPTIONS")
	api.HandleFunc("/learning/courses", authMiddleware.RequireAuth(handler.HandleCreateCourse)).Methods("POST", "OPTIONS")
	api.HandleFunc("/learning/progress", authMiddleware.RequireAuth(handler.HandleMyLearningProgress)).Methods("GET", "OPTIONS")
	api.HandleFunc("/learning/courses/{id}/enroll", authMiddleware.RequireAuth(handler.HandleEnrollCourse)).Methods("POST", "OPTIONS")
	api.HandleFunc("/learning/progress/{id}", authMiddleware.RequireAuth(handler.HandleUpdateProgress)).Methods("PATCH", "OPTIONS")

	// ── Knowledge Base ────────────────────────────────────────────────────
	api.HandleFunc("/knowledge", authMiddleware.RequireAuth(handler.HandleListArticles)).Methods("GET", "OPTIONS")
	api.HandleFunc("/knowledge/search", authMiddleware.RequireAuth(handler.HandleSearchKnowledge)).Methods("GET", "OPTIONS")
	api.HandleFunc("/knowledge/{slug}", authMiddleware.RequireAuth(handler.HandleGetArticle)).Methods("GET", "OPTIONS")
	api.HandleFunc("/knowledge", authMiddleware.RequireAuth(handler.HandleCreateArticle)).Methods("POST", "OPTIONS")
	api.HandleFunc("/knowledge/{id}", authMiddleware.RequireAuth(handler.HandleUpdateArticle)).Methods("PATCH", "OPTIONS")

	// ── Compensation (sensitive — logged access) ──────────────────────────
	api.HandleFunc("/compensation/{employeeId}", authMiddleware.RequireAuth(handler.HandleGetCompensation)).Methods("GET", "OPTIONS")
	api.HandleFunc("/compensation/{employeeId}", authMiddleware.RequireAuth(handler.HandleUpdateCompensation)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/compensation/{employeeId}/equity", authMiddleware.RequireAuth(handler.HandleListEquityGrants)).Methods("GET", "OPTIONS")
	api.HandleFunc("/compensation/{employeeId}/audit", authMiddleware.RequireAuth(handler.HandleGetCompensationAudit)).Methods("GET", "OPTIONS")

	// ── Org Chart ─────────────────────────────────────────────────────────
	api.HandleFunc("/orgchart", authMiddleware.RequireAuth(handler.HandleGetOrgChart)).Methods("GET", "OPTIONS")
	api.HandleFunc("/orgchart/{employeeId}/reports", authMiddleware.RequireAuth(handler.HandleGetDirectReports)).Methods("GET", "OPTIONS")

	// ── Notifications ─────────────────────────────────────────────────────
	api.HandleFunc("/notifications", authMiddleware.RequireAuth(handler.HandleListNotifications)).Methods("GET", "OPTIONS")
	api.HandleFunc("/notifications/unread-count", authMiddleware.RequireAuth(handler.HandleUnreadCount)).Methods("GET", "OPTIONS")
	api.HandleFunc("/notifications/{id}/read", authMiddleware.RequireAuth(handler.HandleMarkRead)).Methods("POST", "OPTIONS")
	api.HandleFunc("/notifications/read-all", authMiddleware.RequireAuth(handler.HandleMarkAllRead)).Methods("POST", "OPTIONS")

	// ── AI Chat ──────────────────────────────────────────────────────────
	api.HandleFunc("/ai/chat/sessions", authMiddleware.RequireAuth(handler.HandleListChatSessions)).Methods("GET", "OPTIONS")
	api.HandleFunc("/ai/chat/sessions", authMiddleware.RequireAuth(handler.HandleCreateChatSession)).Methods("POST", "OPTIONS")
	api.HandleFunc("/ai/chat/sessions/{id}", authMiddleware.RequireAuth(handler.HandleGetChatSession)).Methods("GET", "OPTIONS")
	api.HandleFunc("/ai/chat/sessions/{id}/messages", authMiddleware.RequireAuth(handler.HandleSendMessage)).Methods("POST", "OPTIONS")
	api.HandleFunc("/ai/chat/sessions/{id}", authMiddleware.RequireAuth(handler.HandleDeleteChatSession)).Methods("DELETE", "OPTIONS")

	// ── Performance Goals ────────────────────────────────────────────────
	api.HandleFunc("/performance/goals", authMiddleware.RequireAuth(handler.HandleListGoals)).Methods("GET", "OPTIONS")
	api.HandleFunc("/performance/goals", authMiddleware.RequireAuth(handler.HandleCreateGoal)).Methods("POST", "OPTIONS")
	api.HandleFunc("/performance/goals/{id}", authMiddleware.RequireAuth(handler.HandleUpdateGoal)).Methods("PATCH", "OPTIONS")

	// ── Performance Reviews ──────────────────────────────────────────────
	api.HandleFunc("/performance/reviews", authMiddleware.RequireAuth(handler.HandleListReviews)).Methods("GET", "OPTIONS")
	api.HandleFunc("/performance/reviews", authMiddleware.RequireAuth(handler.HandleCreateReview)).Methods("POST", "OPTIONS")
	api.HandleFunc("/performance/reviews/{id}/submit", authMiddleware.RequireAuth(handler.HandleSubmitReview)).Methods("POST", "OPTIONS")

	// ── Peer Feedback ───────────────────────────────────────────────────
	api.HandleFunc("/performance/feedback", authMiddleware.RequireAuth(handler.HandleListFeedback)).Methods("GET", "OPTIONS")
	api.HandleFunc("/performance/feedback", authMiddleware.RequireAuth(handler.HandleGiveFeedback)).Methods("POST", "OPTIONS")

	// ── Time Tracking ───────────────────────────────────────────────────
	api.HandleFunc("/time-entries", authMiddleware.RequireAuth(handler.HandleListTimeEntries)).Methods("GET", "OPTIONS")
	api.HandleFunc("/time-entries", authMiddleware.RequireAuth(handler.HandleCreateTimeEntry)).Methods("POST", "OPTIONS")
	api.HandleFunc("/time-entries/{id}", authMiddleware.RequireAuth(handler.HandleUpdateTimeEntry)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/time-entries/{id}", authMiddleware.RequireAuth(handler.HandleDeleteTimeEntry)).Methods("DELETE", "OPTIONS")

	// ── PTO Requests ────────────────────────────────────────────────────
	api.HandleFunc("/pto", authMiddleware.RequireAuth(handler.HandleListPTO)).Methods("GET", "OPTIONS")
	api.HandleFunc("/pto", authMiddleware.RequireAuth(handler.HandleRequestPTO)).Methods("POST", "OPTIONS")
	api.HandleFunc("/pto/{id}/approve", authMiddleware.RequireAuth(handler.HandleApprovePTO)).Methods("POST", "OPTIONS")
	api.HandleFunc("/pto/{id}/reject", authMiddleware.RequireAuth(handler.HandleRejectPTO)).Methods("POST", "OPTIONS")

	// ── Analytics ───────────────────────────────────────────────────────
	api.HandleFunc("/analytics/team", authMiddleware.RequireAuth(handler.HandleGetTeamAnalytics)).Methods("GET", "OPTIONS")
	api.HandleFunc("/analytics/skills/{employeeId}", authMiddleware.RequireAuth(handler.HandleGetSkillGapAnalysis)).Methods("GET", "OPTIONS")
	api.HandleFunc("/analytics/time-report", authMiddleware.RequireAuth(handler.HandleGetTimeReport)).Methods("GET", "OPTIONS")

	// ── Innovation Grants (Phase 3) ────────────────────────────────────
	api.HandleFunc("/innovation/grants", authMiddleware.RequireAuth(handler.HandleListGrants)).Methods("GET", "OPTIONS")
	api.HandleFunc("/innovation/grants", authMiddleware.RequireAuth(handler.HandleCreateGrant)).Methods("POST", "OPTIONS")
	api.HandleFunc("/innovation/grants/{id}/submit", authMiddleware.RequireAuth(handler.HandleSubmitGrant)).Methods("POST", "OPTIONS")
	api.HandleFunc("/innovation/grants/{id}/vote", authMiddleware.RequireAuth(handler.HandleVoteGrant)).Methods("POST", "OPTIONS")
	api.HandleFunc("/innovation/grants/{id}/review", authMiddleware.RequireAuth(handler.HandleReviewGrant)).Methods("POST", "OPTIONS")

	// ── Talent Marketplace (Phase 3) ───────────────────────────────────
	api.HandleFunc("/marketplace/opportunities", authMiddleware.RequireAuth(handler.HandleListOpportunities)).Methods("GET", "OPTIONS")
	api.HandleFunc("/marketplace/opportunities", authMiddleware.RequireAuth(handler.HandleCreateOpportunity)).Methods("POST", "OPTIONS")
	api.HandleFunc("/marketplace/opportunities/{id}/apply", authMiddleware.RequireAuth(handler.HandleApplyOpportunity)).Methods("POST", "OPTIONS")
	api.HandleFunc("/marketplace/applications/{id}/review", authMiddleware.RequireAuth(handler.HandleReviewApplication)).Methods("POST", "OPTIONS")

	// ── Career Paths (Phase 3) ─────────────────────────────────────────
	api.HandleFunc("/career/paths", authMiddleware.RequireAuth(handler.HandleListCareerPaths)).Methods("GET", "OPTIONS")
	api.HandleFunc("/career/paths/{id}", authMiddleware.RequireAuth(handler.HandleGetCareerPath)).Methods("GET", "OPTIONS")
	api.HandleFunc("/career/progress", authMiddleware.RequireAuth(handler.HandleGetMyCareerProgress)).Methods("GET", "OPTIONS")
	api.HandleFunc("/career/target", authMiddleware.RequireAuth(handler.HandleSetCareerTarget)).Methods("POST", "OPTIONS")
	api.HandleFunc("/career/paths/{id}/gap-analysis", authMiddleware.RequireAuth(handler.HandleGetGapAnalysis)).Methods("GET", "OPTIONS")

	// ── Mentorship (Phase 3) ───────────────────────────────────────────
	api.HandleFunc("/mentorships", authMiddleware.RequireAuth(handler.HandleListMentorships)).Methods("GET", "OPTIONS")
	api.HandleFunc("/mentorships", authMiddleware.RequireAuth(handler.HandleRequestMentorship)).Methods("POST", "OPTIONS")
	api.HandleFunc("/mentorships/{id}", authMiddleware.RequireAuth(handler.HandleUpdateMentorship)).Methods("PATCH", "OPTIONS")

	// ── Documents (Phase 3) ────────────────────────────────────────────
	api.HandleFunc("/documents", authMiddleware.RequireAuth(handler.HandleListDocuments)).Methods("GET", "OPTIONS")
	api.HandleFunc("/documents/templates", authMiddleware.RequireAuth(handler.HandleListTemplates)).Methods("GET", "OPTIONS")
	api.HandleFunc("/documents/{id}", authMiddleware.RequireAuth(handler.HandleGetDocument)).Methods("GET", "OPTIONS")
	api.HandleFunc("/documents", authMiddleware.RequireAuth(handler.HandleCreateDocument)).Methods("POST", "OPTIONS")
	api.HandleFunc("/documents/{id}", authMiddleware.RequireAuth(handler.HandleUpdateDocument)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/documents/{id}/share", authMiddleware.RequireAuth(handler.HandleShareDocument)).Methods("POST", "OPTIONS")

	// ── Mission Control (Phase 4) ─────────────────────────────────────
	api.HandleFunc("/mission-control", authMiddleware.RequireAuth(handler.HandleGetMissionControl)).Methods("GET", "OPTIONS")
	api.HandleFunc("/mission-control/refresh", authMiddleware.RequireAuth(handler.HandleRefreshSnapshot)).Methods("POST", "OPTIONS")
	api.HandleFunc("/mission-control/snapshots", authMiddleware.RequireAuth(handler.HandleListSnapshots)).Methods("GET", "OPTIONS")

	// ── Team Health (Phase 4) ─────────────────────────────────────────
	api.HandleFunc("/team-health", authMiddleware.RequireAuth(handler.HandleGetTeamHealth)).Methods("GET", "OPTIONS")
	api.HandleFunc("/team-health/burnout", authMiddleware.RequireAuth(handler.HandleGetBurnoutRisk)).Methods("GET", "OPTIONS")

	// ── Skills Graph (Phase 4) ────────────────────────────────────────
	api.HandleFunc("/skills/graph", authMiddleware.RequireAuth(handler.HandleGetSkillsGraph)).Methods("GET", "OPTIONS")
	api.HandleFunc("/skills/gaps", authMiddleware.RequireAuth(handler.HandleGetSkillGap)).Methods("GET", "OPTIONS")

	// ── Reputation (Phase 4) ──────────────────────────────────────────
	api.HandleFunc("/reputation/{employeeId}", authMiddleware.RequireAuth(handler.HandleGetReputation)).Methods("GET", "OPTIONS")
	api.HandleFunc("/reputation/leaderboard", authMiddleware.RequireAuth(handler.HandleGetLeaderboard)).Methods("GET", "OPTIONS")
	api.HandleFunc("/reputation/{employeeId}/trends", authMiddleware.RequireAuth(handler.HandleGetReputationTrends)).Methods("GET", "OPTIONS")

	// ── Badges (Phase 4) ──────────────────────────────────────────────
	api.HandleFunc("/badges", authMiddleware.RequireAuth(handler.HandleListBadges)).Methods("GET", "OPTIONS")
	api.HandleFunc("/badges", authMiddleware.RequireAuth(handler.HandleCreateBadge)).Methods("POST", "OPTIONS")
	api.HandleFunc("/badges/award", authMiddleware.RequireAuth(handler.HandleAwardBadge)).Methods("POST", "OPTIONS")
	api.HandleFunc("/badges/my", authMiddleware.RequireAuth(handler.HandleGetMyBadges)).Methods("GET", "OPTIONS")
	api.HandleFunc("/badges/{employeeId}/{badgeId}", authMiddleware.RequireAuth(handler.HandleRevokeBadge)).Methods("DELETE", "OPTIONS")

	// ── Living Memory (Phase 4) ───────────────────────────────────────
	api.HandleFunc("/memory", authMiddleware.RequireAuth(handler.HandleListMemory)).Methods("GET", "OPTIONS")
	api.HandleFunc("/memory", authMiddleware.RequireAuth(handler.HandleCreateMemory)).Methods("POST", "OPTIONS")
	api.HandleFunc("/memory/search", authMiddleware.RequireAuth(handler.HandleSearchMemory)).Methods("GET", "OPTIONS")
	api.HandleFunc("/memory/{id}", authMiddleware.RequireAuth(handler.HandleGetMemory)).Methods("GET", "OPTIONS")

	// ── Incidents (Phase 5) ─────────────────────────────────────────
	api.HandleFunc("/incidents", authMiddleware.RequireAuth(handler.HandleListIncidents)).Methods("GET", "OPTIONS")
	api.HandleFunc("/incidents/{id}", authMiddleware.RequireAuth(handler.HandleGetIncident)).Methods("GET", "OPTIONS")
	api.HandleFunc("/incidents", authMiddleware.RequireAuth(handler.HandleCreateIncident)).Methods("POST", "OPTIONS")
	api.HandleFunc("/incidents/{id}", authMiddleware.RequireAuth(handler.HandleUpdateIncident)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/incidents/{id}/events", authMiddleware.RequireAuth(handler.HandleListIncidentEvents)).Methods("GET", "OPTIONS")
	api.HandleFunc("/incidents/{id}/events", authMiddleware.RequireAuth(handler.HandleAddIncidentEvent)).Methods("POST", "OPTIONS")
	api.HandleFunc("/incidents/{id}/responders", authMiddleware.RequireAuth(handler.HandleAddResponder)).Methods("POST", "OPTIONS")

	// ── Postmortems (Phase 5) ───────────────────────────────────────
	api.HandleFunc("/incidents/{id}/postmortem", authMiddleware.RequireAuth(handler.HandleGetPostmortem)).Methods("GET", "OPTIONS")
	api.HandleFunc("/postmortems", authMiddleware.RequireAuth(handler.HandleCreatePostmortem)).Methods("POST", "OPTIONS")

	// ── Lifecycle (Phase 5) ────────────────────────────────────────
	api.HandleFunc("/lifecycle/events/{employeeId}", authMiddleware.RequireAuth(handler.HandleListLifecycleEvents)).Methods("GET", "OPTIONS")
	api.HandleFunc("/lifecycle/trigger", authMiddleware.RequireAuth(handler.HandleTriggerLifecycle)).Methods("POST", "OPTIONS")
	api.HandleFunc("/lifecycle/workflows", authMiddleware.RequireAuth(handler.HandleListWorkflows)).Methods("GET", "OPTIONS")
	api.HandleFunc("/lifecycle/workflows", authMiddleware.RequireAuth(handler.HandleCreateWorkflow)).Methods("POST", "OPTIONS")
	api.HandleFunc("/lifecycle/instances/{id}", authMiddleware.RequireAuth(handler.HandleGetWorkflowInstance)).Methods("GET", "OPTIONS")
	api.HandleFunc("/lifecycle/instances/{id}/complete-step", authMiddleware.RequireAuth(handler.HandleCompleteWorkflowStep)).Methods("POST", "OPTIONS")

	// ── Feature Flags (Phase 5) ───────────────────────────────────
	api.HandleFunc("/feature-flags", authMiddleware.RequireAuth(handler.HandleListFlags)).Methods("GET", "OPTIONS")
	api.HandleFunc("/feature-flags/{key}", authMiddleware.RequireAuth(handler.HandleGetFlag)).Methods("GET", "OPTIONS")
	api.HandleFunc("/feature-flags", authMiddleware.RequireAuth(handler.HandleCreateFlag)).Methods("POST", "OPTIONS")
	api.HandleFunc("/feature-flags/{id}", authMiddleware.RequireAuth(handler.HandleUpdateFlag)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/feature-flags/{key}/evaluate", authMiddleware.RequireAuth(handler.HandleEvaluateFlag)).Methods("GET", "OPTIONS")

	// ── Data Classification (Phase 5) ─────────────────────────────
	api.HandleFunc("/data-classifications", authMiddleware.RequireAuth(handler.HandleListClassifications)).Methods("GET", "OPTIONS")
	api.HandleFunc("/data-classifications", authMiddleware.RequireAuth(handler.HandleClassifyResource)).Methods("POST", "OPTIONS")
	api.HandleFunc("/data-classifications/{resourceType}/{resourceId}", authMiddleware.RequireAuth(handler.HandleGetClassification)).Methods("GET", "OPTIONS")

	// ── Employee Certificates (Phase 5) ──────────────────────────
	api.HandleFunc("/certificates/{employeeId}", authMiddleware.RequireAuth(handler.HandleListCertificates)).Methods("GET", "OPTIONS")
	api.HandleFunc("/certificates", authMiddleware.RequireAuth(handler.HandleIssueCertificate)).Methods("POST", "OPTIONS")
	api.HandleFunc("/certificates/{id}/revoke", authMiddleware.RequireAuth(handler.HandleRevokeCertificate)).Methods("POST", "OPTIONS")
	api.HandleFunc("/certificates/serial/{serial}", authMiddleware.RequireAuth(handler.HandleGetCertificate)).Methods("GET", "OPTIONS")

	// ── FWOS Events (Phase 5) ─────────────────────────────────────
	api.HandleFunc("/events", authMiddleware.RequireAuth(handler.HandleListEvents)).Methods("GET", "OPTIONS")
	api.HandleFunc("/events", authMiddleware.RequireAuth(handler.HandleCreateEvent)).Methods("POST", "OPTIONS")

	// ── Email Accounts (Phase 6) ─────────────────────────────────
	api.HandleFunc("/email-accounts", authMiddleware.RequireAuth(handler.HandleListEmails)).Methods("GET", "OPTIONS")
	api.HandleFunc("/email-accounts", authMiddleware.RequireAuth(handler.HandleProvisionEmail)).Methods("POST", "OPTIONS")
	api.HandleFunc("/email-accounts/{id}/status", authMiddleware.RequireAuth(handler.HandleUpdateEmailStatus)).Methods("PATCH", "OPTIONS")

	// ── Devices (Phase 6) ───────────────────────────────────────
	api.HandleFunc("/devices", authMiddleware.RequireAuth(handler.HandleListDevices)).Methods("GET", "OPTIONS")
	api.HandleFunc("/devices", authMiddleware.RequireAuth(handler.HandleRegisterDevice)).Methods("POST", "OPTIONS")
	api.HandleFunc("/devices/{id}", authMiddleware.RequireAuth(handler.HandleGetDevice)).Methods("GET", "OPTIONS")
	api.HandleFunc("/devices/{id}", authMiddleware.RequireAuth(handler.HandleUpdateDevice)).Methods("PATCH", "OPTIONS")

	// ── SSO Provisioning (Phase 6) ─────────────────────────────
	api.HandleFunc("/sso/configs", authMiddleware.RequireAuth(handler.HandleListSSOConfigs)).Methods("GET", "OPTIONS")
	api.HandleFunc("/sso/configs", authMiddleware.RequireAuth(handler.HandleCreateSSOConfig)).Methods("POST", "OPTIONS")
	api.HandleFunc("/sso/configs/{id}/sync", authMiddleware.RequireAuth(handler.HandleSyncSSO)).Methods("POST", "OPTIONS")
	api.HandleFunc("/sso/configs/{id}/logs", authMiddleware.RequireAuth(handler.HandleGetProvisioningLogs)).Methods("GET", "OPTIONS")

	// ── Wallet Passes (Phase 6) ────────────────────────────────
	api.HandleFunc("/wallet/passes/{employeeId}", authMiddleware.RequireAuth(handler.HandleListWalletPasses)).Methods("GET", "OPTIONS")
	api.HandleFunc("/wallet/passes", authMiddleware.RequireAuth(handler.HandleGenerateWalletPass)).Methods("POST", "OPTIONS")
	api.HandleFunc("/wallet/passes/{id}/revoke", authMiddleware.RequireAuth(handler.HandleRevokeWalletPass)).Methods("POST", "OPTIONS")
	api.HandleFunc("/wallet/verify/{token}", authMiddleware.RequireAuth(handler.HandleVerifyBadgeQR)).Methods("GET", "OPTIONS")

	// ── Push & Notification Prefs (Phase 6) ───────────────────
	api.HandleFunc("/push/subscribe", authMiddleware.RequireAuth(handler.HandleSubscribePush)).Methods("POST", "OPTIONS")
	api.HandleFunc("/notifications/preferences", authMiddleware.RequireAuth(handler.HandleGetNotificationPrefs)).Methods("GET", "OPTIONS")
	api.HandleFunc("/notifications/preferences", authMiddleware.RequireAuth(handler.HandleUpdateNotificationPrefs)).Methods("POST", "OPTIONS")

	// ── 360 Feedback Rounds (Remaining) ──────────────────────
	api.HandleFunc("/feedback-rounds", authMiddleware.RequireAuth(handler.HandleListFeedbackRounds)).Methods("GET", "OPTIONS")
	api.HandleFunc("/feedback-rounds", authMiddleware.RequireAuth(handler.HandleCreateFeedbackRound)).Methods("POST", "OPTIONS")
	api.HandleFunc("/feedback-rounds/{id}/start", authMiddleware.RequireAuth(handler.HandleStartFeedbackRound)).Methods("POST", "OPTIONS")
	api.HandleFunc("/feedback-rounds/responses", authMiddleware.RequireAuth(handler.HandleSubmitFeedbackResponse)).Methods("POST", "OPTIONS")
	api.HandleFunc("/feedback-rounds/{id}/results", authMiddleware.RequireAuth(handler.HandleGetFeedbackResults)).Methods("GET", "OPTIONS")

	// ── Goal Cascade (Remaining) ─────────────────────────────
	api.HandleFunc("/performance/goals/tree", authMiddleware.RequireAuth(handler.HandleGetGoalTree)).Methods("GET", "OPTIONS")
	api.HandleFunc("/performance/goals/{id}/cascade", authMiddleware.RequireAuth(handler.HandleCascadeGoal)).Methods("POST", "OPTIONS")

	// ── Document Signatures (Remaining) ──────────────────────
	api.HandleFunc("/documents/signatures", authMiddleware.RequireAuth(handler.HandleRequestSignature)).Methods("POST", "OPTIONS")
	api.HandleFunc("/documents/signatures/{id}/sign", authMiddleware.RequireAuth(handler.HandleSignDocument)).Methods("POST", "OPTIONS")
	api.HandleFunc("/documents/signatures/{id}/decline", authMiddleware.RequireAuth(handler.HandleDeclineSignature)).Methods("POST", "OPTIONS")
	api.HandleFunc("/documents/signatures/{id}", authMiddleware.RequireAuth(handler.HandleGetSignatureStatus)).Methods("GET", "OPTIONS")

	// ── Certificate PKI (Remaining) ─────────────────────────
	api.HandleFunc("/certificates/keys", authMiddleware.RequireAuth(handler.HandleGenerateCertificateKey)).Methods("POST", "OPTIONS")
	api.HandleFunc("/certificates/{certificateId}/chain", authMiddleware.RequireAuth(handler.HandleGetCertificateChain)).Methods("GET", "OPTIONS")

	// ── Wallet Pass Templates (Remaining) ────────────────────
	api.HandleFunc("/wallet/templates", authMiddleware.RequireAuth(handler.HandleListPassTemplates)).Methods("GET", "OPTIONS")
	api.HandleFunc("/wallet/templates", authMiddleware.RequireAuth(handler.HandleGeneratePassFile)).Methods("POST", "OPTIONS")

	// ── Org Chart Import (Remaining) ─────────────────────────
	api.HandleFunc("/orgchart/import", authMiddleware.RequireAuth(handler.HandleUploadOrgChart)).Methods("POST", "OPTIONS")
	api.HandleFunc("/orgchart/import/{id}", authMiddleware.RequireAuth(handler.HandleGetImportStatus)).Methods("GET", "OPTIONS")

	// ── Package Registry (Remaining) ─────────────────────────
	api.HandleFunc("/packages", authMiddleware.RequireAuth(handler.HandleListPackages)).Methods("GET", "OPTIONS")
	api.HandleFunc("/packages/{id}", authMiddleware.RequireAuth(handler.HandleGetPackage)).Methods("GET", "OPTIONS")
	api.HandleFunc("/packages", authMiddleware.RequireAuth(handler.HandlePublishPackage)).Methods("POST", "OPTIONS")
	api.HandleFunc("/packages/{id}/versions", authMiddleware.RequireAuth(handler.HandleListVersions)).Methods("GET", "OPTIONS")

	// ── FFID Identity System ──────────────────────────────────────
	api.HandleFunc("/identity/{employeeId}", authMiddleware.RequireAuth(handler.HandleGetIdentityCard)).Methods("GET", "OPTIONS")
	api.HandleFunc("/identity/{employeeId}/timeline", authMiddleware.RequireAuth(handler.HandleListCareerTimeline)).Methods("GET", "OPTIONS")
	api.HandleFunc("/identity/{employeeId}/timeline", authMiddleware.RequireAuth(handler.HandleCreateTimelineEvent)).Methods("POST", "OPTIONS")
	api.HandleFunc("/identity/{employeeId}/achievements", authMiddleware.RequireAuth(handler.HandleGetAchievementProgress)).Methods("GET", "OPTIONS")
	api.HandleFunc("/identity/{employeeId}/achievements/check", authMiddleware.RequireAuth(handler.HandleCheckAchievements)).Methods("POST", "OPTIONS")
	api.HandleFunc("/achievements", authMiddleware.RequireAuth(handler.HandleListAchievementDefinitions)).Methods("GET", "OPTIONS")
	api.HandleFunc("/achievements", authMiddleware.RequireAuth(handler.HandleCreateAchievementDefinition)).Methods("POST", "OPTIONS")
	api.HandleFunc("/achievements/seed", authMiddleware.RequireAuth(handler.HandleSeedAchievements)).Methods("POST", "OPTIONS")
	api.HandleFunc("/employees/{id}/clearance", authMiddleware.RequireAuth(handler.HandleUpdateClearanceLevel)).Methods("PATCH", "OPTIONS")
}
