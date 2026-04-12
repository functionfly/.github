package api

import (
	"github.com/functionfly/functionfly/internal/api/handlers/privacy"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/gorilla/mux"
)

// registerPrivacyRoutes registers all privacy-related API routes
// This includes user-facing privacy endpoints for GDPR compliance,
// data export/deletion, consent management, and privacy settings
func registerPrivacyRoutes(
	api *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	privacyHandler *privacy.Handler,
) {
	// Privacy settings routes (user-facing)
	api.HandleFunc("/privacy/settings", authMiddleware.RequireAuth(privacyHandler.HandleGetPrivacySettings)).Methods("GET", "OPTIONS")
	api.HandleFunc("/privacy/settings", authMiddleware.RequireAuth(privacyHandler.HandleUpdatePrivacySettings)).Methods("PUT", "PATCH", "OPTIONS")
	api.HandleFunc("/privacy/settings", authMiddleware.RequireAuth(privacyHandler.HandleDeletePrivacySettings)).Methods("DELETE", "OPTIONS")

	// Consent management routes
	api.HandleFunc("/privacy/consent", authMiddleware.RequireAuth(privacyHandler.HandleGetConsentStatus)).Methods("GET", "OPTIONS")
	api.HandleFunc("/privacy/consent", authMiddleware.RequireAuth(privacyHandler.HandleRecordConsent)).Methods("POST", "OPTIONS")
	api.HandleFunc("/privacy/consent/withdraw", authMiddleware.RequireAuth(privacyHandler.HandleWithdrawConsent)).Methods("POST", "OPTIONS")

	// Data export routes (GDPR Article 20)
	api.HandleFunc("/privacy/export", authMiddleware.RequireAuth(privacyHandler.HandleRequestDataExport)).Methods("POST", "OPTIONS")
	api.HandleFunc("/privacy/export/{id}", authMiddleware.RequireAuth(privacyHandler.HandleGetExportStatus)).Methods("GET", "OPTIONS")
	api.HandleFunc("/privacy/export/{id}/download", authMiddleware.RequireAuth(privacyHandler.HandleGetExportDownload)).Methods("GET", "OPTIONS")

	// Data deletion routes (GDPR Article 17 - Right to erasure)
	api.HandleFunc("/privacy/deletion", authMiddleware.RequireAuth(privacyHandler.HandleRequestDataDeletion)).Methods("POST", "OPTIONS")
	api.HandleFunc("/privacy/deletion/{id}", authMiddleware.RequireAuth(privacyHandler.HandleGetDeletionStatus)).Methods("GET", "OPTIONS")

	// PII scanning utility
	api.HandleFunc("/privacy/scan-pii", authMiddleware.RequireAuth(privacyHandler.HandleScanForPII)).Methods("POST", "OPTIONS")

	// Admin privacy routes
	registerAdminPrivacyRoutes(api, authMiddleware, privacyHandler)
}

// registerAdminPrivacyRoutes registers admin-only privacy management routes
func registerAdminPrivacyRoutes(
	api *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	privacyHandler *privacy.Handler,
) {
	adminRoutes := api.PathPrefix("/admin").Subrouter()

	// Global privacy settings management
	adminRoutes.HandleFunc("/privacy/settings", authMiddleware.RequirePermission(auth.PermSystemRead)(privacyHandler.HandleGetGlobalPrivacySettings)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/privacy/settings", authMiddleware.RequirePermission(auth.PermSystemWrite)(privacyHandler.HandleUpdateGlobalPrivacySettings)).Methods("PUT", "PATCH", "OPTIONS")

	// Admin data export/deletion management
	adminRoutes.HandleFunc("/privacy/exports", authMiddleware.RequirePermission(auth.PermAuditRead)(privacyHandler.HandleListExportRequests)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/privacy/exports/{id}", authMiddleware.RequirePermission(auth.PermAuditRead)(privacyHandler.HandleAdminGetExportStatus)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/privacy/deletions", authMiddleware.RequirePermission(auth.PermAuditRead)(privacyHandler.HandleListDeletionRequests)).Methods("GET", "OPTIONS")
	adminRoutes.HandleFunc("/privacy/deletions/{id}", authMiddleware.RequirePermission(auth.PermAuditRead)(privacyHandler.HandleAdminGetDeletionStatus)).Methods("GET", "OPTIONS")

	// Privacy audit log
	adminRoutes.HandleFunc("/privacy/audit", authMiddleware.RequirePermission(auth.PermAuditRead)(privacyHandler.HandleListPrivacyAuditLogs)).Methods("GET", "OPTIONS")
}
