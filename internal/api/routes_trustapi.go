package api

import (
	"context"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/handlers/trustapi"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apikey"
	"github.com/functionfly/functionfly/internal/logging"
	"github.com/functionfly/functionfly/internal/storage/registry"
	trustapirepo "github.com/functionfly/functionfly/internal/storage/trustapi"
	"github.com/gorilla/mux"
)

// trustAPIServerInterface defines the methods needed from Server for Trust API routes
type trustAPIServerInterface interface {
	GetTrustBillingService() *trustapi.BillingService
}

// GetTrustBillingService returns the Trust API billing service (implements interface)
func (s *Server) GetTrustBillingService() *trustapi.BillingService {
	return s.trustBillingService
}

// registerTrustAPIRoutes wires all Trust API routes for external platform partners
func registerTrustAPIRoutes(
	s *Server,
	api *mux.Router,
	registryRepo *registry.RegistryRepository,
) {
	// Initialize Trust API repository
	trustRepo := trustapirepo.NewRepository(s.postgresDB.GORM)

	// Initialize platform API key repository (unified system)
	apikeyRepo := apikey.NewRepository(s.postgresDB.GORM)

	// Initialize revocation repository for extended functionality
	revocationRepo := trustapirepo.NewRevocationRepository(s.postgresDB.GORM)

	// Initialize signing key repository for key rotation support
	signingKeyRepo := trustapirepo.NewSigningKeyRepository(s.postgresDB.GORM)
	trustapirepo.SetSigningKeyRepository(signingKeyRepo)

	// Initialize webhook repository and service
	webhookRepo := trustapirepo.NewWebhookRepository(s.postgresDB.GORM)
	webhookService := trustapirepo.NewWebhookService(webhookRepo)
	webhookService.SetLogger(logging.Logger())

	// Initialize handlers
	trustHandler := trustapi.NewHandler(apikeyRepo, trustRepo, registryRepo)
	webhookHandler := trustapi.NewWebhookHandler(webhookRepo)
	webhookHandler.SetLogger(logging.Logger())

	// Initialize extended handler with revocation and webhook capabilities
	extendedHandler := trustapi.NewExtendedHandler(apikeyRepo, trustRepo, registryRepo, revocationRepo, webhookService)

	// Initialize and wire audit log repository for compliance logging
	auditLogRepo := trustapirepo.NewAuditLogRepository(s.postgresDB.GORM)
	extendedHandler.SetAuditLogRepository(auditLogRepo)

	// Initialize Merkle audit trail repository
	merkleRepo := trustapirepo.NewMerkleRepository(s.postgresDB.GORM)
	extendedHandler.SetMerkleRepository(merkleRepo)

	// Start background job for automatic revocation expiration
	extendedHandler.StartRevocationExpirationJob(context.Background())

	// Start background job for automatic attestation expiration
	extendedHandler.StartAttestationExpirationJob(context.Background())

	// Initialize streaming handler
	trustStreamer := trustapi.NewTrustScoreStreamer(registryRepo, logging.Logger())
	go trustStreamer.Run()

	// Initialize middleware with both repos (apikey for auth, trust for partner/rate-limit data)
	authMiddleware := trustapi.NewAPIKeyAuthMiddleware(apikeyRepo, trustRepo)
	rateLimitMiddleware := trustapi.NewRateLimitMiddleware(trustRepo)
	usageTrackingMiddleware := trustapi.NewUsageTrackingMiddleware(trustRepo)

	// IP-based rate limiter for the public registration endpoint (5 per IP per hour)
	registrationRateLimiter := trustapi.NewRegistrationRateLimiter(5, 1*time.Hour)

	// Get internal auth middleware for JWT-protected routes
	internalAuthMiddleware := middleware.NewAuthMiddleware(s.authSvc)

	// Feature middleware for plan-based access control (JWT-authenticated users)
	featureMiddleware := middleware.NewFeatureMiddleware()

	// Create a subrouter for Trust API
	trustAPI := api.PathPrefix("/v1").Subrouter()

	// Apply global Trust API middleware
	// Order: Body Size Limit -> Usage Tracking -> Rate Limiting -> Authentication
	trustAPI.Use(trustapi.BodySizeLimit(1 << 20)) // 1MB max body size
	trustAPI.Use(usageTrackingMiddleware.Track())
	trustAPI.Use(rateLimitMiddleware.RateLimit())

	// Partner registration and management (create is public, read/update require JWT auth)
	// Registration has IP-based rate limiting (5/hour) + CAPTCHA verification in handler
	trustAPI.Handle("/partners",
		registrationRateLimiter.RateLimit()(http.HandlerFunc(trustHandler.HandleCreatePartner))).Methods("POST")
	trustAPI.Handle("/partners",
		internalAuthMiddleware.RequireAuth(trustHandler.HandleListPartners)).Methods("GET")
	trustAPI.Handle("/partners/{partner_id}",
		internalAuthMiddleware.RequireAuth(trustHandler.HandleGetPartner)).Methods("GET")
	trustAPI.Handle("/partners/{partner_id}",
		internalAuthMiddleware.RequireAuth(trustHandler.HandleUpdatePartner)).Methods("PATCH")
	trustAPI.Handle("/partners/{partner_id}/usage",
		internalAuthMiddleware.RequireAuth(trustHandler.HandleGetPartnerUsage)).Methods("GET")

	// API Key management (requires partner auth)
	apiKeyAuthRouter := trustAPI.PathPrefix("/partners/{partner_id}/api-keys").Subrouter()
	apiKeyAuthRouter.Use(authMiddleware.Authenticate())
	apiKeyAuthRouter.HandleFunc("", trustHandler.HandleCreateAPIKey).Methods("POST")
	apiKeyAuthRouter.HandleFunc("", trustHandler.HandleListAPIKeys).Methods("GET")
	apiKeyAuthRouter.HandleFunc("/{key_id}", trustHandler.HandleRevokeAPIKey).Methods("DELETE")

	// Trust endpoints (requires API key authentication)
	trustAuthRouter := trustAPI.PathPrefix("/trust").Subrouter()
	trustAuthRouter.Use(authMiddleware.Authenticate())

	// Trust score endpoints
	trustAuthRouter.HandleFunc("/score/{function_id}", trustHandler.HandleGetTrustScore).Methods("GET")
	trustAuthRouter.HandleFunc("/batch", trustHandler.HandleBatchTrustScore).Methods("POST")
	trustAuthRouter.HandleFunc("/history/{function_id}", trustHandler.HandleGetTrustHistory).Methods("GET")

	// Trust verification endpoints (requires verification:request scope or Starter+ plan)
	trustAuthRouter.Handle("/verify",
		authMiddleware.RequireScope("verification:request")(http.HandlerFunc(trustHandler.HandleSubmitVerification))).Methods("POST")
	trustAuthRouter.Handle("/verify",
		internalAuthMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			featureMiddleware.RequireFeature("trust_verification")(
				http.HandlerFunc(trustHandler.HandleSubmitVerification)).ServeHTTP(w, r)
		}))).Methods("POST")
	trustAuthRouter.HandleFunc("/verify/{verification_id}", trustHandler.HandleGetVerification).Methods("GET")

	// Trust reporting endpoints (requires reports:submit scope or Starter+ plan)
	trustAuthRouter.Handle("/report",
		authMiddleware.RequireScope("reports:submit")(http.HandlerFunc(trustHandler.HandleSubmitReport))).Methods("POST")
	trustAuthRouter.Handle("/report",
		internalAuthMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			featureMiddleware.RequireFeature("trust_reports")(
				http.HandlerFunc(trustHandler.HandleSubmitReport)).ServeHTTP(w, r)
		}))).Methods("POST")
	trustAuthRouter.HandleFunc("/report/{report_id}", trustHandler.HandleGetReport).Methods("GET")

	// ============================================
	// Extended Trust API Routes (New Functionality)
	// ============================================

	// Attestation read endpoints (API key auth, default trust:read scope)
	trustAuthRouter.HandleFunc("/attestations", extendedHandler.HandleGetAttestations).Methods("GET")
	trustAuthRouter.HandleFunc("/attestations/{attestation_id}", extendedHandler.HandleGetAttestation).Methods("GET")
	trustAuthRouter.HandleFunc("/attestations/{attestation_id}/verify", extendedHandler.HandleVerifyAttestation).Methods("GET")
	trustAuthRouter.HandleFunc("/attestations/chain/{function_id}", extendedHandler.HandleGetAttestationChain).Methods("GET")
	trustAuthRouter.HandleFunc("/attestations/chain/{function_id}/verify", extendedHandler.HandleVerifyChain).Methods("GET")
	trustAuthRouter.HandleFunc("/attestations/public-key", extendedHandler.HandleGetPublicKey).Methods("GET")
	// Blog-referenced aliases
	trustAuthRouter.HandleFunc("/keys", extendedHandler.HandleGetPublicKey).Methods("GET")

	// Signer status endpoint (JWT auth, admin or Pro+ plan)
	trustAuthRouter.HandleFunc("/signer/status", extendedHandler.HandleGetSignerStatus).Methods("GET")
	// Signer test endpoint (JWT auth, admin only)
	trustAuthRouter.HandleFunc("/signer/test", extendedHandler.HandleTestSigner).Methods("POST")
	// Signing key history (JWT auth)
	trustAuthRouter.HandleFunc("/signer/keys", extendedHandler.HandleListSigningKeys).Methods("GET")
	// Key rotation (JWT auth, admin only)
	trustAuthRouter.HandleFunc("/signer/rotate", extendedHandler.HandleRotateSigningKey).Methods("POST")

	// Attestation create: API key partners need attestation:create scope + startup tier,
	// JWT users need Professional+ plan (attestation_create feature)
	trustAuthRouter.Handle("/attestations",
		authMiddleware.RequireScope("attestation:create")(
			authMiddleware.RequireTier("startup")(
				http.HandlerFunc(extendedHandler.HandleCreateAttestation)))).Methods("POST")
	trustAuthRouter.Handle("/attestations",
		internalAuthMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			featureMiddleware.RequireFeature("attestation_create")(
				http.HandlerFunc(extendedHandler.HandleCreateAttestation)).ServeHTTP(w, r)
		}))).Methods("POST")

	// Attestation revoke: API key partners need attestation:revoke scope + business tier,
	// JWT users need Enterprise plan (attestation_revoke feature)
	trustAuthRouter.Handle("/attestations/{attestation_id}/revoke",
		authMiddleware.RequireScope("attestation:revoke")(
			authMiddleware.RequireTier("business")(
				http.HandlerFunc(extendedHandler.HandleRevokeAttestation)))).Methods("POST")
	trustAuthRouter.Handle("/attestations/{attestation_id}/revoke",
		internalAuthMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			featureMiddleware.RequireFeature("attestation_revoke")(
				http.HandlerFunc(extendedHandler.HandleRevokeAttestation)).ServeHTTP(w, r)
		}))).Methods("POST")

	// Verification completion: API key partners need verification:request scope,
	// JWT users need Starter+ plan (trust_verification feature)
	trustAuthRouter.Handle("/verify/{verification_id}/complete",
		authMiddleware.RequireScope("verification:request")(
			http.HandlerFunc(extendedHandler.HandleCompleteVerification))).Methods("POST")
	trustAuthRouter.Handle("/verify/{verification_id}/complete",
		internalAuthMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			featureMiddleware.RequireFeature("trust_verification")(
				http.HandlerFunc(extendedHandler.HandleCompleteVerification)).ServeHTTP(w, r)
		}))).Methods("POST")

	// Policy endpoints — evaluate requires API key (trust:read) or JWT (Pro+)
	trustAuthRouter.Handle("/policies",
		internalAuthMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			featureMiddleware.RequireFeature("trust_policy_evaluate")(
				http.HandlerFunc(extendedHandler.HandleListPolicies)).ServeHTTP(w, r)
		}))).Methods("GET")
	trustAuthRouter.HandleFunc("/policies/{policy_id}", extendedHandler.HandleGetPolicy).Methods("GET")
	trustAuthRouter.Handle("/policies/evaluate",
		authMiddleware.RequireScope("trust:read")(http.HandlerFunc(extendedHandler.HandleEvaluatePolicy))).Methods("POST")
	trustAuthRouter.Handle("/policies/evaluate",
		internalAuthMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			featureMiddleware.RequireFeature("trust_policy_evaluate")(
				http.HandlerFunc(extendedHandler.HandleEvaluatePolicy)).ServeHTTP(w, r)
		}))).Methods("POST")
	trustAuthRouter.Handle("/policies/evaluate/batch",
		authMiddleware.RequireScope("trust:read")(http.HandlerFunc(extendedHandler.HandleBatchEvaluatePolicy))).Methods("POST")
	trustAuthRouter.Handle("/policies/evaluate/batch",
		internalAuthMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			featureMiddleware.RequireFeature("trust_policy_evaluate")(
				http.HandlerFunc(extendedHandler.HandleBatchEvaluatePolicy)).ServeHTTP(w, r)
		}))).Methods("POST")

	// Policy management requires JWT authentication + Pro+ plan
	trustAuthRouter.Handle("/policies",
		internalAuthMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			featureMiddleware.RequireFeature("trust_policy_manage")(
				http.HandlerFunc(extendedHandler.HandleCreatePolicy)).ServeHTTP(w, r)
		}))).Methods("POST")
	trustAuthRouter.Handle("/policies/{policy_id}",
		internalAuthMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			featureMiddleware.RequireFeature("trust_policy_manage")(
				http.HandlerFunc(extendedHandler.HandleUpdatePolicy)).ServeHTTP(w, r)
		}))).Methods("PUT")
	trustAuthRouter.Handle("/policies/{policy_id}",
		internalAuthMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			featureMiddleware.RequireFeature("trust_policy_manage")(
				http.HandlerFunc(extendedHandler.HandleDeletePolicy)).ServeHTTP(w, r)
		}))).Methods("DELETE")

	// ============================================
	// Merkle Audit Trail Routes
	// ============================================
	// Read endpoints (API key auth with trust:read scope — default)
	trustAuthRouter.HandleFunc("/merkle/head", extendedHandler.HandleGetMerkleTreeHead).Methods("GET")
	trustAuthRouter.HandleFunc("/merkle/root", extendedHandler.HandleGetMerkleRoot).Methods("GET")
	trustAuthRouter.HandleFunc("/merkle/inclusion", extendedHandler.HandleGetMerkleInclusionProof).Methods("GET")
	trustAuthRouter.HandleFunc("/merkle/consistency", extendedHandler.HandleGetMerkleConsistencyProof).Methods("GET")
	trustAuthRouter.HandleFunc("/merkle/verify/inclusion", extendedHandler.HandleVerifyMerkleInclusion).Methods("POST")
	// Blog-referenced: GET /v1/trust/merkle/proof/{attestation_id} — looks up leaf index by attestation ID
	trustAuthRouter.HandleFunc("/merkle/proof/{attestation_id}", extendedHandler.HandleGetMerkleProofForAttestation).Methods("GET")

	// ============================================
	// Chain of Custody (Delegation) Routes
	// ============================================
	trustAuthRouter.HandleFunc("/delegation/chain/{chain_id}", extendedHandler.HandleGetDelegationChain).Methods("GET")
	trustAuthRouter.HandleFunc("/delegation/chain/{chain_id}/verify", extendedHandler.HandleVerifyDelegationChain).Methods("GET")
	trustAuthRouter.HandleFunc("/delegation/function/{function_id}", extendedHandler.HandleGetFunctionDelegationChains).Methods("GET")

	// Revocation endpoints require admin/internal JWT authentication
	revokeRouter := trustAPI.PathPrefix("/trust/revoke").Subrouter()

	// Revocation listing/checking (read-only, requires auth)
	revokeRouter.Handle("/revoked",
		internalAuthMiddleware.RequireAuth(extendedHandler.HandleListRevocations)).Methods("GET")
	revokeRouter.Handle("/revoked/{function_id}",
		internalAuthMiddleware.RequireAuth(extendedHandler.HandleCheckFunctionRevoked)).Methods("GET")
	revokeRouter.Handle("/{revocation_id}",
		internalAuthMiddleware.RequireAuth(extendedHandler.HandleGetRevocation)).Methods("GET")

	// Revocation create/lift (admin only, requires auth)
	revokeRouter.Handle("",
		internalAuthMiddleware.RequireAuth(extendedHandler.HandleRevokeTrust)).Methods("POST")
	revokeRouter.Handle("/{revocation_id}/lift",
		internalAuthMiddleware.RequireAuth(extendedHandler.HandleUnrevokeTrust)).Methods("POST")

	// ============================================
	// Webhook Management Routes (JWT authenticated)
	// ============================================
	webhookRouter := api.PathPrefix("/v1/webhooks").Subrouter()

	// Webhook CRUD (all require auth)
	webhookRouter.Handle("",
		internalAuthMiddleware.RequireAuth(webhookHandler.HandleCreateWebhook)).Methods("POST")
	webhookRouter.Handle("",
		internalAuthMiddleware.RequireAuth(webhookHandler.HandleListWebhooks)).Methods("GET")
	webhookRouter.Handle("/{webhook_id}",
		internalAuthMiddleware.RequireAuth(webhookHandler.HandleGetWebhook)).Methods("GET")
	webhookRouter.Handle("/{webhook_id}",
		internalAuthMiddleware.RequireAuth(webhookHandler.HandleUpdateWebhook)).Methods("PUT")
	webhookRouter.Handle("/{webhook_id}",
		internalAuthMiddleware.RequireAuth(webhookHandler.HandleDeleteWebhook)).Methods("DELETE")

	// Webhook testing and stats
	webhookRouter.Handle("/{webhook_id}/test",
		internalAuthMiddleware.RequireAuth(webhookHandler.HandleTestWebhook)).Methods("POST")
	webhookRouter.Handle("/{webhook_id}/deliveries",
		internalAuthMiddleware.RequireAuth(webhookHandler.HandleListDeliveries)).Methods("GET")
	webhookRouter.Handle("/{webhook_id}/stats",
		internalAuthMiddleware.RequireAuth(webhookHandler.HandleGetDeliveryStats)).Methods("GET")

	// ============================================
	// Real-time Trust Score Streaming Routes
	// ============================================
	// SSE streaming for trust score updates (requires JWT auth)

	// SSE endpoint for streaming trust score updates for watched functions
	api.HandleFunc("/v1/trust/stream/sse",
		internalAuthMiddleware.RequireAuth(trustStreamer.HandleSSE)).Methods("GET")

	// SSE endpoint for a specific function
	api.HandleFunc("/v1/trust/stream/functions/{function_id}/sse",
		internalAuthMiddleware.RequireAuth(trustStreamer.HandleSSEFunction)).Methods("GET")

	// ============================================
	// Trust API Billing Routes
	// ============================================
	// Partner billing endpoints (requires JWT auth)
	if s.GetTrustBillingService() != nil {
		// Initialize billing repository for the handler
		trustBillingRepo := trustapirepo.NewBillingRepository(s.postgresDB.GORM)
		billingHandler := trustapi.NewBillingHandler(s.GetTrustBillingService(), trustBillingRepo)

		// Stripe webhook (no auth - uses Stripe signature verification)
		stripeWebhookHandler := trustapi.NewStripeWebhookHandler(s.GetTrustBillingService(), trustBillingRepo)
		api.HandleFunc("/v1/webhooks/stripe", stripeWebhookHandler.HandleStripeWebhook).Methods("POST")

		// Tier pricing (public)
		trustAPI.HandleFunc("/partners/tiers", billingHandler.HandleGetTierPricing).Methods("GET")

		// Partner billing status and endpoints (requires JWT auth)
		trustAPI.Handle("/partners/{partner_id}/billing",
			internalAuthMiddleware.RequireAuth(billingHandler.HandleGetBillingStatus)).Methods("GET")
		trustAPI.Handle("/partners/{partner_id}/billing/checkout",
			internalAuthMiddleware.RequireAuth(billingHandler.HandleCreateCheckout)).Methods("POST")
		trustAPI.Handle("/partners/{partner_id}/billing/usage",
			internalAuthMiddleware.RequireAuth(billingHandler.HandleGetUsageReport)).Methods("GET")
		trustAPI.Handle("/partners/{partner_id}/billing/invoices",
			internalAuthMiddleware.RequireAuth(billingHandler.HandleGetInvoices)).Methods("GET")
		trustAPI.Handle("/partners/{partner_id}/founder",
			internalAuthMiddleware.RequireAuth(billingHandler.HandleEnrollFounderMode)).Methods("POST")
	}
}
