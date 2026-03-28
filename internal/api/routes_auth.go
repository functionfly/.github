package api

import (
	"github.com/functionfly/functionfly/internal/api/handlers/apikeys"
	authHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/auth"
	"github.com/functionfly/functionfly/internal/api/handlers/billing"
	followHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/follow"
	mfaHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/mfa"
	notificationHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/notifications"
	usersHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/users"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
)

// registerAuthRoutes wires all authentication, user, follow, API-key, billing,
// MFA, and notification endpoints.
func registerAuthRoutes(
	router *mux.Router,
	api *mux.Router,
	authRateLimiter *middleware.AuthRateLimiter,
	authMiddleware *middleware.AuthMiddleware,
	authHandler *authHandlerPkg.Handler,
	apiKeyAuthHandler *apikeys.APIKeyAuthHandler,
	usersHandler *usersHandlerPkg.Handler,
	followHandler *followHandlerPkg.Handler,
	apiKeysHandler *apikeys.Handler,
	billingHandler *billing.Handler,
	mfaHandler *mfaHandlerPkg.MFAHandler,
	notificationHandler *notificationHandlerPkg.Handler,
	notificationWSHandler *notificationHandlerPkg.WebSocketHandler,
) {
	// ── Auth (public) ──────────────────────────────────────────────────────
	// Registered on both the bare router and /v1 so the CLI (fly login) works
	// with or without the prefix.
	router.HandleFunc("/auth/login", authRateLimiter.Limit(authHandler.HandleLogin)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/login", authRateLimiter.Limit(authHandler.HandleLogin)).Methods("POST", "OPTIONS")
	router.HandleFunc("/auth/refresh", authHandler.HandleRefreshToken).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/refresh", authHandler.HandleRefreshToken).Methods("POST", "OPTIONS")
	router.HandleFunc("/auth/signup", authRateLimiter.Limit(authHandler.HandleSignup)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/signup", authRateLimiter.Limit(authHandler.HandleSignup)).Methods("POST", "OPTIONS")
	router.HandleFunc("/auth/signup-config", authHandler.HandleSignupConfig).Methods("GET", "OPTIONS")
	api.HandleFunc("/auth/signup-config", authHandler.HandleSignupConfig).Methods("GET", "OPTIONS")
	router.HandleFunc("/auth/check-username", authHandler.HandleCheckUsernameAvailability).Methods("GET", "OPTIONS")
	router.HandleFunc("/auth/verify-email", authHandler.HandleVerifyEmail).Methods("GET", "OPTIONS")
	router.HandleFunc("/auth/resend-verification", authRateLimiter.Limit(authHandler.HandleResendVerification)).Methods("POST", "OPTIONS")
	router.HandleFunc("/auth/get-session", authHandler.HandleGetSession).Methods("GET", "OPTIONS")
	// Supabase-compatible path used by @neondatabase/neon-js (auth.url + /api/auth/get-session)
	router.HandleFunc("/api/auth/get-session", authHandler.HandleGetSession).Methods("GET", "OPTIONS")

	// OAuth (public)
	router.HandleFunc("/auth/oauth/providers", authHandler.HandleGetOAuthProviders).Methods("GET", "OPTIONS")
	router.HandleFunc("/auth/oauth/url", authRateLimiter.Limit(authHandler.HandleGetOAuthURL)).Methods("GET", "OPTIONS")
	router.HandleFunc("/auth/oauth/{provider}/callback", authHandler.HandleOAuthCallback).Methods("GET", "OPTIONS")
	api.HandleFunc("/auth/oauth/providers", authHandler.HandleGetOAuthProviders).Methods("GET", "OPTIONS")
	api.HandleFunc("/auth/oauth/url", authRateLimiter.Limit(authHandler.HandleGetOAuthURL)).Methods("GET", "OPTIONS")
	api.HandleFunc("/auth/oauth/{provider}/callback", authHandler.HandleOAuthCallback).Methods("GET", "OPTIONS")
	api.HandleFunc("/auth/validate", authMiddleware.RequireAuth(authHandler.HandleValidateToken)).Methods("GET", "OPTIONS")
	api.HandleFunc("/auth/logout", authMiddleware.RequireAuth(authHandler.HandleLogout)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/verify-password", authMiddleware.RequireAuth(authHandler.HandleVerifyPassword)).Methods("POST", "OPTIONS")

	// Password reset (public, rate-limited)
	api.HandleFunc("/auth/password-reset", authRateLimiter.Limit(authHandler.HandlePasswordResetRequest)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/password-reset/confirm", authRateLimiter.Limit(authHandler.HandlePasswordResetConfirm)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/api-key", apiKeyAuthHandler.HandleAuthenticate).Methods("POST", "OPTIONS")

	// MFA (protected)
	api.HandleFunc("/auth/mfa/setup", authMiddleware.RequireAuth(mfaHandler.SetupMFA)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/mfa/verify", authMiddleware.RequireAuth(mfaHandler.VerifyMFA)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/mfa/enable", authMiddleware.RequireAuth(mfaHandler.EnableMFA)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/mfa/disable", authMiddleware.RequireAuth(mfaHandler.DisableMFA)).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/mfa/status", authMiddleware.RequireAuth(mfaHandler.GetMFAStatus)).Methods("GET", "OPTIONS")

	// ── Users ──────────────────────────────────────────────────────────────
	// /users/me* must be registered before /users/{username} so "me" is never
	// captured as a username param (which would 404).
	api.HandleFunc("/users/me", authMiddleware.RequireAuth(usersHandler.HandleGetMe)).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/me", authMiddleware.RequireAuth(usersHandler.HandleUpdateMe)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/users/me/sessions", authMiddleware.RequireAuth(usersHandler.HandleListSessions)).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/me/sessions/revoke-others", authMiddleware.RequireAuth(usersHandler.HandleRevokeOtherSessions)).Methods("POST", "OPTIONS")
	api.HandleFunc("/users/me/sessions/{id}", authMiddleware.RequireAuth(usersHandler.HandleRevokeSession)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/users/me/settings", authMiddleware.RequireAuth(usersHandler.HandleGetUserSettingsMe)).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/me/settings/profile", authMiddleware.RequireAuth(usersHandler.HandlePatchUserSettingsProfileMe)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/users/me/settings/notifications", authMiddleware.RequireAuth(usersHandler.HandlePatchUserSettingsNotificationsMe)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/users/me/settings/privacy", authMiddleware.RequireAuth(usersHandler.HandlePatchUserSettingsPrivacyMe)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/users/me/settings/visibility", authMiddleware.RequireAuth(usersHandler.HandlePatchUserSettingsVisibilityMe)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/users/me/activity", authMiddleware.RequireAuth(usersHandler.HandleCreateUserActivity)).Methods("POST", "OPTIONS")
	api.HandleFunc("/users/me/skills", authMiddleware.RequireAuth(usersHandler.HandleAddUserSkill)).Methods("POST", "OPTIONS")
	api.HandleFunc("/users/me/skills/{id}", authMiddleware.RequireAuth(usersHandler.HandleRemoveUserSkill)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/users/me/notification-preferences", authMiddleware.RequireAuth(notificationHandler.HandleGetPreferences)).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/me/notification-preferences", authMiddleware.RequireAuth(notificationHandler.HandleUpdatePreferences)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/users/lookup-by-ids", authMiddleware.RequireAuth(usersHandler.HandleLookupUsersByIDs)).Methods("POST", "OPTIONS")
	api.HandleFunc("/users/search", authMiddleware.RequireAuth(usersHandler.HandleSearchUsers)).Methods("GET", "OPTIONS")

	// Public user profile (by username)
	api.HandleFunc("/users/{username}", usersHandler.HandleGetPublicProfile).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/{username}/settings", authMiddleware.RequireAuth(usersHandler.HandleGetUserSettings)).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/{username}/settings/profile", authMiddleware.RequireAuth(usersHandler.HandlePatchUserSettingsProfile)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/users/{username}/settings/notifications", authMiddleware.RequireAuth(usersHandler.HandlePatchUserSettingsNotifications)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/users/{username}/settings/privacy", authMiddleware.RequireAuth(usersHandler.HandlePatchUserSettingsPrivacy)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/users/{username}/settings/visibility", authMiddleware.RequireAuth(usersHandler.HandlePatchUserSettingsVisibility)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/users/{username}/analytics", usersHandler.HandleGetUserAnalytics).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/{username}/achievements", usersHandler.HandleGetUserAchievements).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/{username}/activity", usersHandler.HandleGetUserActivity).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/{username}/contributions", usersHandler.HandleGetUserContributions).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/{username}/skills", usersHandler.HandleGetUserSkills).Methods("GET", "OPTIONS")
	api.HandleFunc("/users/{username}/report", authMiddleware.RequireAuth(usersHandler.HandleReportProfile)).Methods("POST", "OPTIONS")
	api.HandleFunc("/@/{username}", usersHandler.HandleGetPublicProfileByAt).Methods("GET", "OPTIONS")

	// ── Follow ─────────────────────────────────────────────────────────────
	api.HandleFunc("/follow/users/{username}/follow", authMiddleware.RequireAuth(followHandler.HandleFollowUser)).Methods("POST", "OPTIONS")
	api.HandleFunc("/follow/users/{username}/follow", authMiddleware.RequireAuth(followHandler.HandleUnfollowUser)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/follow/users/{username}/followers", followHandler.HandleGetUserFollowers).Methods("GET", "OPTIONS")
	api.HandleFunc("/follow/users/{username}/following", followHandler.HandleGetUserFollowing).Methods("GET", "OPTIONS")
	api.HandleFunc("/follow/users/{username}/status", authMiddleware.RequireAuth(followHandler.HandleCheckFollowingStatus)).Methods("GET", "OPTIONS")
	api.HandleFunc("/follow/functions/{functionID}/follow", authMiddleware.RequireAuth(followHandler.HandleFollowFunction)).Methods("POST", "OPTIONS")
	api.HandleFunc("/follow/functions/{functionID}/follow", authMiddleware.RequireAuth(followHandler.HandleUnfollowFunction)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/follow/functions/{functionID}/followers", followHandler.HandleGetFunctionFollowers).Methods("GET", "OPTIONS")
	api.HandleFunc("/follow/functions/{functionID}/status", authMiddleware.RequireAuth(followHandler.HandleCheckFunctionFollowingStatus)).Methods("GET", "OPTIONS")
	api.HandleFunc("/follow/me/functions", authMiddleware.RequireAuth(followHandler.HandleGetMyFollowedFunctions)).Methods("GET", "OPTIONS")
	api.HandleFunc("/follow/me/stats", authMiddleware.RequireAuth(followHandler.HandleGetMyFollowStats)).Methods("GET", "OPTIONS")

	// ── API Keys (protected) ────────────────────────────────────────────────
	api.HandleFunc("/api-keys/environments/available", authMiddleware.RequireAuth(apiKeysHandler.HandleListAvailableEnvironments)).Methods("GET", "OPTIONS")
	api.HandleFunc("/api-keys", authMiddleware.RequireAuth(apiKeysHandler.HandleList)).Methods("GET", "OPTIONS")
	api.HandleFunc("/api-keys", authMiddleware.RequireAuth(apiKeysHandler.HandleCreate)).Methods("POST", "OPTIONS")
	api.HandleFunc("/api-keys/{id}", authMiddleware.RequireAuth(apiKeysHandler.HandleGet)).Methods("GET", "OPTIONS")
	api.HandleFunc("/api-keys/{id}", authMiddleware.RequireAuth(apiKeysHandler.HandleUpdate)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/api-keys/{id}", authMiddleware.RequireAuth(apiKeysHandler.HandleDelete)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/api-keys/{id}/rotate", authMiddleware.RequireAuth(apiKeysHandler.HandleRotate)).Methods("POST", "OPTIONS")
	api.HandleFunc("/api-keys/{id}/permissions", authMiddleware.RequireAuth(apiKeysHandler.HandleListPermissions)).Methods("GET", "OPTIONS")
	api.HandleFunc("/api-keys/{id}/permissions", authMiddleware.RequireAuth(apiKeysHandler.HandleAddPermission)).Methods("POST", "OPTIONS")
	api.HandleFunc("/api-keys/{id}/permissions/{perm_id}", authMiddleware.RequireAuth(apiKeysHandler.HandleRemovePermission)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/api-keys/{id}/environments", authMiddleware.RequireAuth(apiKeysHandler.HandleListEnvironments)).Methods("GET", "OPTIONS")
	api.HandleFunc("/api-keys/{id}/environments", authMiddleware.RequireAuth(apiKeysHandler.HandleAddEnvironment)).Methods("POST", "OPTIONS")
	api.HandleFunc("/api-keys/{id}/environments/{env_id}", authMiddleware.RequireAuth(apiKeysHandler.HandleRemoveEnvironment)).Methods("DELETE", "OPTIONS")

	// ── Billing ─────────────────────────────────────────────────────────────
	// State Fabric add-on catalog (public — pricing page)
	api.HandleFunc("/billing/state-fabric/add-ons", billingHandler.HandleListStateFabricAddOnCatalog).Methods("GET", "OPTIONS")
	// Billing (protected)
	api.HandleFunc("/billing/portal-session", authMiddleware.RequireAuth(billingHandler.HandleCreatePortalSession)).Methods("POST", "OPTIONS")
	api.HandleFunc("/billing/checkout", authMiddleware.RequireAuth(billingHandler.HandleCreateCheckoutSession)).Methods("POST", "OPTIONS")
	api.HandleFunc("/billing/subscription", authMiddleware.RequireAuth(billingHandler.HandleGetSubscription)).Methods("GET", "OPTIONS")
	api.HandleFunc("/billing/subscription/cancel", authMiddleware.RequireAuth(billingHandler.HandleCancelSubscription)).Methods("POST", "OPTIONS")
	api.HandleFunc("/billing/invoices", authMiddleware.RequireAuth(billingHandler.HandleListInvoices)).Methods("GET", "OPTIONS")
	api.HandleFunc("/billing/usage", authMiddleware.RequireAuth(billingHandler.HandleGetUsage)).Methods("GET", "OPTIONS")
	api.HandleFunc("/billing/wallet", authMiddleware.RequireAuth(billingHandler.HandleGetWallet)).Methods("GET", "OPTIONS")
	api.HandleFunc("/billing/wallet/top-up", authMiddleware.RequireAuth(billingHandler.HandleWalletTopUp)).Methods("POST", "OPTIONS")
	api.HandleFunc("/billing/fees", authMiddleware.RequireAuth(billingHandler.HandleListPlatformFees)).Methods("GET", "OPTIONS")

	// ── Revenue System Phase 1 - Trust Layer Monetization ──────────────────
	api.HandleFunc("/billing/plans", authMiddleware.RequireAuth(billingHandler.HandleGetPlans)).Methods("GET", "OPTIONS")
	api.HandleFunc("/billing/subscribe", authMiddleware.RequireAuth(billingHandler.HandleSubscribe)).Methods("POST", "OPTIONS")
	api.HandleFunc("/billing/verification-cost", authMiddleware.RequireAuth(billingHandler.HandleGetVerificationCost)).Methods("GET", "OPTIONS")
	api.HandleFunc("/billing/verify-function", authMiddleware.RequireAuth(billingHandler.HandleVerifyFunction)).Methods("POST", "OPTIONS")
	api.HandleFunc("/billing/earnings", authMiddleware.RequireAuth(billingHandler.HandleGetEarnings)).Methods("GET", "OPTIONS")
	api.HandleFunc("/billing/agent-usage", authMiddleware.RequireAuth(billingHandler.HandleGetAgentUsage)).Methods("GET", "OPTIONS")
	api.HandleFunc("/billing/state-fabric/add-ons/entitlements", authMiddleware.RequireAuth(billingHandler.HandleGetStateFabricAddOnEntitlements)).Methods("GET", "OPTIONS")
	api.HandleFunc("/billing/state-fabric/add-ons/checkout", authMiddleware.RequireAuth(billingHandler.HandleCreateStateFabricAddOnCheckout)).Methods("POST", "OPTIONS")
	// Internal webhook endpoint for subscription updates (called by Stripe webhook handler or admin)
	// Requires X-Internal-Webhook-Secret header matching INTERNAL_WEBHOOK_SECRET env var
	api.HandleFunc("/billing/subscription/webhook", billingHandler.HandleSubscriptionWebhook).Methods("POST", "OPTIONS")

	// ── Notifications (protected) ───────────────────────────────────────────
	api.HandleFunc("/notifications", authMiddleware.RequireAuth(notificationHandler.HandleListNotifications)).Methods("GET", "OPTIONS")
	api.HandleFunc("/notifications/unread-count", authMiddleware.RequireAuth(notificationHandler.HandleGetUnreadCount)).Methods("GET", "OPTIONS")
	api.HandleFunc("/notifications/read-all", authMiddleware.RequireAuth(notificationHandler.HandleMarkAllAsRead)).Methods("POST", "OPTIONS")
	api.HandleFunc("/notifications/{id}/read", authMiddleware.RequireAuth(notificationHandler.HandleMarkAsRead)).Methods("PATCH", "OPTIONS")
	api.HandleFunc("/notifications/{id}", authMiddleware.RequireAuth(notificationHandler.HandleDeleteNotification)).Methods("DELETE", "OPTIONS")
	api.HandleFunc("/notifications/stream", authMiddleware.RequireAuth(notificationWSHandler.HandleWebSocket))
}
