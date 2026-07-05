package api

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/handlers/blog"
	"github.com/functionfly/functionfly/internal/api/handlers/content"
	"github.com/functionfly/functionfly/internal/api/handlers/demo"
	feedbackHandlerPkg "github.com/functionfly/functionfly/internal/api/handlers/feedback"
	"github.com/functionfly/functionfly/internal/api/handlers/playground"
	"github.com/functionfly/functionfly/internal/api/handlers/recommendations"
	registryhandler "github.com/functionfly/functionfly/internal/api/handlers/registry"
	drehandler "github.com/functionfly/functionfly/internal/api/handlers/registry/dre"
	versionhandler "github.com/functionfly/functionfly/internal/api/handlers/version"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/gorilla/mux"
	"github.com/functionfly/functionfly/internal/apierror"
)

// registerRegistryRoutes wires all function registry, playground, documentation,
// embed, versioning, canary, DRE, verification, content, and feedback endpoints.
func registerRegistryRoutes(
	s *Server,
	api *mux.Router,
	apiV2 *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	executionSecurityMW *middleware.ExecutionCoordinatorMiddleware,
	verificationMiddleware *middleware.VerificationMiddleware,
	registryRepo *storageregistry.RegistryRepository,
	registryHandler *registryhandler.Handler,
	registryPlaygroundHandler *registryhandler.PlaygroundHandler,
	appPlaygroundHandler *playground.Handler,
	docsHandler *registryhandler.DocumentationHandler,
	tutorialsHandler *registryhandler.TutorialsHandler,
	versionHandler *versionhandler.Handler,
	blogHandler *blog.Handler,
	contentHandler *content.Handler,
	feedbackHandler *feedbackHandlerPkg.Handler,
	recommendationHandler *recommendations.Handler,
	anchoringService drehandler.AnchorServicer,
	demoHandler *demo.Handler,
	localUploadHandler *registryhandler.LocalUploadHandler,
) {
	// Inline init for registry-specific sub-handlers
	canaryHandler := registryhandler.NewCanaryHandler(
		storageregistry.NewCanaryConfigRepository(s.postgresDB.GORM),
		registryRepo,
	)
	versionManager := middleware.NewVersionManager()
	deprecationHandler := registryhandler.NewDeprecationHandler(versionManager)
	migrationHandler := registryhandler.NewMigrationHandler()
	var dreHandler *drehandler.Handler
	if anchoringService != nil {
		dreHandler = drehandler.NewHandlerWithAnchoring(registryRepo, anchoringService)
	} else {
		dreHandler = drehandler.NewHandler(registryRepo)
	}

	// ── App-based Playground (public) ───────────────────────────────────────
	// NOTE: uses /app-run/ prefix to avoid collision with registry playground /run/ routes.
	api.HandleFunc("/app-run/{appSlug}/{functionName}", appPlaygroundHandler.HandlePlaygroundUI).Methods("GET", "OPTIONS")
	api.HandleFunc("/app-run/{appSlug}/{functionName}/info", appPlaygroundHandler.HandleGetFunctionInfo).Methods("GET", "OPTIONS")
	api.HandleFunc("/app-run/{appSlug}/{functionName}/execute", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		app, err := s.repo.GetAppBySlug(r.Context(), vars["appSlug"])
		if err != nil || app == nil {
			apierror.WriteError(w, apierror.NewNotFound("App not found"))
			return
		}
		fn, err := s.repo.GetFunctionByAppIDAndName(r.Context(), app.ID, vars["functionName"])
		if err != nil || fn == nil {
			apierror.WriteError(w, apierror.NewNotFound("Function not found"))
			return
		}
		executionSecurityMW.SecureExecution(fn.ID, fn.Version)(appPlaygroundHandler.HandleExecute).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	// ── Zero-Friction Demo API (public, no auth, no signup) ────────────────
	// Must be registered BEFORE the /{author}/{name} catchall below, otherwise
	// POST /v1/demo/execute is matched as author="demo", name="execute" and
	// returns "Function not found" before the demo handler ever sees it.
	if demoHandler != nil {
		api.HandleFunc("/demo", demoHandler.ListFunctions).Methods("GET", "OPTIONS")
		api.HandleFunc("/demo/execute", demoHandler.HandleExecute).Methods("POST", "OPTIONS")
	}

	// ── Public Execute (v1 prefix, /fx alias for compatibility) ─────────────
	secureExecuteHandler := func(w http.ResponseWriter, r *http.Request) {
		repo := storageregistry.NewRegistryRepository(s.postgresDB.GORM, nil)
		vars := mux.Vars(r)
		fn, err := repo.GetFunctionByAuthorName(r.Context(), vars["author"], vars["name"])
		if err != nil {
			apierror.WriteError(w, apierror.NewNotFound("Function not found"))
			return
		}
		version := vars["version"]
		executionSecurityMW.SecureExecution(fn.ID, version)(registryHandler.HandleExecute).ServeHTTP(w, r)
	}

	// NOTE: Publish routes MUST be registered BEFORE the {author}/{name}
	// catch-alls below, otherwise POST /v1/registry/publish is matched as
	// author="registry", name="publish" by the execute handler.
	api.HandleFunc("/registry/publish", authMiddleware.RequireAuth(registryHandler.HandlePublish)).Methods("POST", "OPTIONS")
	apiV2.HandleFunc("/registry/publish", authMiddleware.RequireAuth(registryHandler.HandlePublish)).Methods("POST", "OPTIONS")

	// Direct-to-R2 artifact upload. The dashboard calls this first for large
	// payloads, gets a presigned URL + the storage keys to use, uploads the
	// file to R2, then calls POST /registry/publish with presigned_upload_complete=true.
	api.HandleFunc("/registry/publish/presign", authMiddleware.RequireAuth(registryHandler.HandlePublishPresign)).Methods("POST", "OPTIONS")
	apiV2.HandleFunc("/registry/publish/presign", authMiddleware.RequireAuth(registryHandler.HandlePublishPresign)).Methods("POST", "OPTIONS")

	// Presigned GET for browser-side source/WASM preview.
	api.HandleFunc("/artifacts/download", authMiddleware.RequireAuth(registryHandler.HandleArtifactDownload)).Methods("POST", "OPTIONS")

	// Health probe for the artifact store wiring.
	api.HandleFunc("/artifacts/health", registryHandler.HandleArtifactHealth).Methods("GET", "OPTIONS")

	// Local-upload/download are only available when the artifact store is the
	// local filesystem backend (dev / --skip-migrations). The wrapping is a
	// no-op when LocalUploadHandler is nil; the routes return 503 cleanly.
	if localUploadHandler != nil {
		api.Handle("/artifacts/local-upload", localUploadHandler).Methods("PUT", "OPTIONS")
		api.Handle("/artifacts/local-download", localUploadHandler.LocalDownloadHandler()).Methods("GET", "OPTIONS")
	}

	if verificationMiddleware != nil {
		api.Handle("/{author}/{name}", verificationMiddleware.RequireVerifiedFunction("standard")(http.HandlerFunc(secureExecuteHandler))).Methods("POST", "OPTIONS")
		api.Handle("/{author}/{name}@{version}", verificationMiddleware.RequireVerifiedFunction("standard")(http.HandlerFunc(secureExecuteHandler))).Methods("POST", "OPTIONS")
		api.Handle("/fx/{author}/{name}", verificationMiddleware.RequireVerifiedFunction("standard")(http.HandlerFunc(secureExecuteHandler))).Methods("POST", "OPTIONS")
		api.Handle("/fx/{author}/{name}@{version}", verificationMiddleware.RequireVerifiedFunction("standard")(http.HandlerFunc(secureExecuteHandler))).Methods("POST", "OPTIONS")
	} else {
		api.HandleFunc("/{author}/{name}", secureExecuteHandler).Methods("POST", "OPTIONS")
		api.HandleFunc("/{author}/{name}@{version}", secureExecuteHandler).Methods("POST", "OPTIONS")
		api.HandleFunc("/fx/{author}/{name}", secureExecuteHandler).Methods("POST", "OPTIONS")
		api.HandleFunc("/fx/{author}/{name}@{version}", secureExecuteHandler).Methods("POST", "OPTIONS")
	}

	// ── Registry Playground (public) ────────────────────────────────────────
	api.HandleFunc("/fx/{author}/{name}", registryPlaygroundHandler.HandleFunctionPage).Methods("GET", "OPTIONS")
	api.HandleFunc("/fx/{author}/{name}/code", registryPlaygroundHandler.HandleCodeExamples).Methods("GET", "OPTIONS")
	api.HandleFunc("/fx/{author}/{name}/ai-schema", registryPlaygroundHandler.HandleAIToolSchema).Methods("GET", "OPTIONS")
	api.HandleFunc("/run/{author}/{name}", registryPlaygroundHandler.HandlePlaygroundUI).Methods("GET", "OPTIONS")
	api.HandleFunc("/run/{author}/{name}/execute", registryPlaygroundHandler.HandlePlaygroundExecute).Methods("POST", "OPTIONS")
	api.HandleFunc("/run/{author}/{name}/execute/stream", registryPlaygroundHandler.HandlePlaygroundExecuteStream).Methods("POST", "OPTIONS")
	api.HandleFunc("/run/{author}/{name}/share", registryPlaygroundHandler.HandlePlaygroundShare).Methods("POST", "OPTIONS")
	api.HandleFunc("/replay/{executionId}", registryPlaygroundHandler.HandleReplay).Methods("GET", "OPTIONS")

	// ── @username Social Profile Style (primary route) ───────────────────────
	api.HandleFunc("/@/{username}/v1/fx/{functionName}", registryPlaygroundHandler.HandleFunctionPageAt).Methods("GET", "OPTIONS")
	api.HandleFunc("/@/{username}/v1/fx/{functionName}/execute", registryPlaygroundHandler.HandleExecuteAt).Methods("POST", "OPTIONS")
	api.HandleFunc("/@/{username}/v1/fx/{functionName}@{version}", registryPlaygroundHandler.HandleFunctionPageAtVersion).Methods("GET", "OPTIONS")
	api.HandleFunc("/@/{username}/v1/fx/{functionName}/versions", registryHandler.HandleListVersionsAt).Methods("GET", "OPTIONS")
	api.HandleFunc("/@/{username}/v1/fx/{functionName}/stats", registryHandler.HandleGetFunctionStatsAt).Methods("GET", "OPTIONS")

	// ── Registry CRUD (public read) - /functions prefix ──────────────────────
	api.HandleFunc("/functions", registryHandler.HandleListFunctions).Methods("GET")
	api.HandleFunc("/functions/search", registryHandler.HandleSearchFunctions).Methods("GET")
	api.HandleFunc("/functions/{author}/{name}", registryHandler.HandleGetFunction).Methods("GET")
	api.HandleFunc("/functions/{author}/{name}", authMiddleware.RequireAuth(registryHandler.HandleDeleteFunction)).Methods("DELETE")
	api.HandleFunc("/functions/{author}/{name}/versions", registryHandler.HandleListVersions).Methods("GET")
	api.HandleFunc("/functions/{author}/{name}/source", registryHandler.HandleGetFunctionSource).Methods("GET")
	api.HandleFunc("/functions/{author}/{name}/history", registryHandler.HandleGetVersionHistory).Methods("GET")
	api.HandleFunc("/functions/{author}/{name}/changelogs", registryHandler.HandleGetChangelogs).Methods("GET")
	api.HandleFunc("/functions/{author}/{name}/changelogs/{version}", registryHandler.HandleGetChangelogByVersion).Methods("GET")
	api.HandleFunc("/functions/{author}/{name}/changelogs/category/{category}", registryHandler.HandleGetChangelogByCategory).Methods("GET")

	// ── Function Versions (registry management) ─────────────────────────────
	api.HandleFunc("/functions/{functionId}/versions", versionHandler.HandleListFunctionVersions).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/versions/{version}", versionHandler.HandleGetFunctionVersion).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/versions/{version}/changelog", authMiddleware.RequireAuth(versionHandler.HandleCreateChangelog)).Methods("POST", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/changelogs", versionHandler.HandleGetChangelogs).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/versions/{version}/publish", authMiddleware.RequireAuth(versionHandler.HandlePublishVersion)).Methods("POST", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/versions/{version}/archive", authMiddleware.RequireAuth(versionHandler.HandleArchiveVersion)).Methods("POST", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/versions/{version}/deprecate", authMiddleware.RequireAuth(versionHandler.HandleDeprecateFunctionVersion)).Methods("POST", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/versions/{version}/alias/{alias}", authMiddleware.RequireAuth(versionHandler.HandleSetAlias)).Methods("POST", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/versions/{version}/rollback", authMiddleware.RequireAuth(versionHandler.HandleRollbackVersion)).Methods("POST", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/rollback", authMiddleware.RequireAuth(versionHandler.HandleRollbackLatest)).Methods("POST", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/rollbacks", versionHandler.HandleGetRollbackHistory).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/versions/{version}/deployments", versionHandler.HandleListDeployments).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/versions/{version}/deployments/{deploymentId}", versionHandler.HandleGetDeployment).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/versions/{version}/lineage", versionHandler.HandleGetVersionLineage).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/versions/compare", versionHandler.HandleCompareVersions).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{functionId}/trust", registryHandler.HandleGetTrustScore).Methods("GET")
	api.HandleFunc("/functions/{functionId}/trust/history", registryHandler.HandleGetTrustHistory).Methods("GET")
	api.HandleFunc("/functions/{functionId}/trust/refresh", authMiddleware.RequireAuth(registryHandler.HandleRefreshTrustScore)).Methods("POST")

	// ── Service Contracts (internal) ─────────────────────────────────────────
	api.HandleFunc("/internal/contracts", versionHandler.HandleListServiceContracts).Methods("GET", "OPTIONS")
	api.HandleFunc("/internal/contracts/{service}", versionHandler.HandleGetServiceContracts).Methods("GET", "OPTIONS")
	api.HandleFunc("/internal/contracts/negotiate", versionHandler.HandleNegotiateContractVersion).Methods("POST", "OPTIONS")

	// ── Recommendations (query params for filtering) ─────────────────────────
	api.HandleFunc("/recommendations", authMiddleware.RequireAuth(recommendationHandler.HandleGetRecommendations)).Methods("GET", "OPTIONS")
	api.HandleFunc("/recommendations/interactions", authMiddleware.RequireAuth(recommendationHandler.HandleRecordInteraction)).Methods("POST", "OPTIONS")
	api.HandleFunc("/recommendations/executions", authMiddleware.RequireAuth(recommendationHandler.HandleRecordExecution)).Methods("POST", "OPTIONS")
	api.HandleFunc("/recommendations/feedback", authMiddleware.RequireAuth(recommendationHandler.HandleRecordFeedback)).Methods("POST", "OPTIONS")
	api.HandleFunc("/recommendations/refresh", authMiddleware.RequirePermission(auth.PermSystemWrite)(recommendationHandler.HandleRefreshRecommendations)).Methods("POST", "OPTIONS")
	api.HandleFunc("/recommendations/triple-search", authMiddleware.RequireAuth(recommendationHandler.HandleTripleSearch)).Methods("POST", "OPTIONS")
	api.HandleFunc("/recommendations/composable/{function_id}", authMiddleware.RequireAuth(recommendationHandler.HandleFindComposable)).Methods("GET", "OPTIONS")

	// ── Registry v2 ──────────────────────────────────────────────────────────
	apiV2.HandleFunc("/functions/mine", authMiddleware.RequireAuth(registryHandler.HandleListMyFunctions)).Methods("GET")
	apiV2.HandleFunc("/functions", registryHandler.HandleListFunctions).Methods("GET")
	apiV2.HandleFunc("/functions/search", registryHandler.HandleSearchFunctions).Methods("GET")
	apiV2.HandleFunc("/functions/{author}/{name}", registryHandler.HandleGetFunction).Methods("GET")
	apiV2.HandleFunc("/functions/{author}/{name}/versions", registryHandler.HandleListVersions).Methods("GET")
	apiV2.HandleFunc("/functions/{author}/{name}/changelogs", registryHandler.HandleGetChangelogs).Methods("GET")
	apiV2.HandleFunc("/functions/{author}/{name}/changelogs/{version}", registryHandler.HandleGetChangelogByVersion).Methods("GET")
	apiV2.HandleFunc("/functions/{author}/{name}/history", registryHandler.HandleGetVersionHistory).Methods("GET")

	// ── Canary Deployments ───────────────────────────────────────────────────
	api.HandleFunc("/functions/{author}/{name}/canary", canaryHandler.HandleCreateCanary).Methods("POST")
	api.HandleFunc("/functions/{author}/{name}/canary", canaryHandler.HandleGetCanary).Methods("GET")
	api.HandleFunc("/functions/{author}/{name}/canary", canaryHandler.HandleUpdateCanary).Methods("PATCH")
	api.HandleFunc("/functions/{author}/{name}/canary", canaryHandler.HandleCancelCanary).Methods("DELETE")
	api.HandleFunc("/functions/{author}/{name}/canary/promote", canaryHandler.HandlePromoteCanary).Methods("POST")
	api.HandleFunc("/functions/{author}/{name}/canary/rollback", canaryHandler.HandleRollbackCanary).Methods("POST")
	api.HandleFunc("/functions/{author}/{name}/canary/history", canaryHandler.HandleGetCanaryHistory).Methods("GET")

	// ── Deprecation & Migration Guides ───────────────────────────────────────
	api.HandleFunc("/deprecations", deprecationHandler.HandleGetAllDeprecations).Methods("GET")
	api.HandleFunc("/deprecations/{endpoint}", deprecationHandler.HandleGetEndpointDeprecation).Methods("GET")
	api.HandleFunc("/migration-guide", migrationHandler.HandleGetMigrationGuide).Methods("GET")
	api.HandleFunc("/migration-guide/{endpoint}", migrationHandler.HandleGetEndpointMigration).Methods("GET")
	api.HandleFunc("/versions", migrationHandler.HandleGetVersionInfo).Methods("GET")

	// ── Documentation (public) ───────────────────────────────────────────────
	api.HandleFunc("/docs", docsHandler.HandleIndex).Methods("GET")
	api.HandleFunc("/docs/{author}/{name}", docsHandler.HandleFunctionHTMLDocs).Methods("GET")
	api.HandleFunc("/docs/{author}/{name}/versions", docsHandler.HandleFunctionVersions).Methods("GET")
	api.HandleFunc("/docs/openapi.json", docsHandler.HandleOpenAPISpec).Methods("GET")
	api.HandleFunc("/docs/{author}/{name}/api", docsHandler.HandleFunctionDocs).Methods("GET")

	// ── Tutorials (public) ───────────────────────────────────────────────────
	api.HandleFunc("/tutorials", tutorialsHandler.HandleIndex).Methods("GET")
	api.HandleFunc("/tutorials/getting-started", tutorialsHandler.HandleGettingStarted).Methods("GET")
	api.HandleFunc("/tutorials/api-usage", tutorialsHandler.HandleAPIUsage).Methods("GET")
	api.HandleFunc("/tutorials/function-development", tutorialsHandler.HandleFunctionDevelopment).Methods("GET")
	api.HandleFunc("/tutorials/examples/{example}", tutorialsHandler.HandleInteractiveExample).Methods("GET")

	// ── CDN Static Assets (public) ───────────────────────────────────────────
	api.HandleFunc("/sdk/{sdk}/{version}/{filename}", registryHandler.HandleServeSDK).Methods("GET")
	api.HandleFunc("/docs/{type}/{version}/{path}", registryHandler.HandleServeDocs).Methods("GET")
	api.HandleFunc("/static/{category}/{path}", registryHandler.HandleServeStatic).Methods("GET")

	// ── Embed (public read for script serving, protected for config/analytics) ─
	api.HandleFunc("/embed/{author}/{nameVersion}", registryHandler.HandleServeEmbed).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/embed", authMiddleware.RequireAuth(registryHandler.HandleGetEmbedConfig)).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/embed/snippet", registryHandler.HandleGetEmbedSnippet).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/embed/analytics", authMiddleware.RequireAuth(registryHandler.HandleGetEmbedAnalytics)).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/embed", authMiddleware.RequireAuth(registryHandler.HandleUpdateEmbedConfig)).Methods("PUT", "OPTIONS")

	// Function settings (protected)
	api.HandleFunc("/functions/{author}/{name}/settings", authMiddleware.RequireAuth(registryHandler.HandleGetFunctionSettings)).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/settings", authMiddleware.RequireAuth(registryHandler.HandlePatchFunctionSettings)).Methods("PATCH", "OPTIONS")

	// Environment variables (protected)
	api.HandleFunc("/functions/{author}/{name}/env", authMiddleware.RequireAuth(registryHandler.HandleGetEnvVars)).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/env", authMiddleware.RequireAuth(registryHandler.HandlePutEnvVars)).Methods("PUT", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/env/{key}", authMiddleware.RequireAuth(registryHandler.HandleDeleteEnvVar)).Methods("DELETE", "OPTIONS")

	// Secrets (protected)
	api.HandleFunc("/functions/{author}/{name}/secrets", authMiddleware.RequireAuth(registryHandler.HandleGetSecrets)).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/secrets", authMiddleware.RequireAuth(registryHandler.HandlePutSecrets)).Methods("PUT", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/secrets/{key}", authMiddleware.RequireAuth(registryHandler.HandleDeleteSecret)).Methods("DELETE", "OPTIONS")

	// ── MCP Settings (per-function) ──────────────────────────────────────────
	api.HandleFunc("/functions/{author}/{name}/mcp", authMiddleware.RequireAuth(registryHandler.HandleGetMCPSettings)).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/mcp", authMiddleware.RequireAuth(registryHandler.HandleUpdateMCPSettings)).Methods("PATCH", "OPTIONS")

	// ── MCP Functions list (all functions with MCP settings) ──────────────────
	api.HandleFunc("/functions/mcp", registryHandler.HandleListFunctionsWithMCP).Methods("GET", "OPTIONS")

	// Cache monitoring (public)
	api.HandleFunc("/cache/stats", registryHandler.HandleGetCacheStats).Methods("GET")

	// ── Stats / test / rating / reviews (public or protected) ────────────────
	api.HandleFunc("/functions/{author}/{name}/stats", registryHandler.HandleGetFunctionStats).Methods("GET")
	api.HandleFunc("/functions/{author}/{name}/test", authMiddleware.RequireAuth(registryHandler.HandleTest)).Methods("POST")
	api.HandleFunc("/functions/{author}/{name}/rating", authMiddleware.RequireAuth(registryHandler.HandleSubmitRating)).Methods("POST")
	api.HandleFunc("/functions/{author}/{name}/reviews", registryHandler.HandleListReviews).Methods("GET")
	api.HandleFunc("/functions/{author}/{name}/reviews", authMiddleware.RequireAuth(registryHandler.HandleSubmitReview)).Methods("POST")
	api.HandleFunc("/functions/{author}/{name}/aggregate", authMiddleware.RequireAuth(registryHandler.HandleAggregateStats)).Methods("POST")

	// ── Trust Scoring (public) ──────────────────────────────────────────────
	api.HandleFunc("/functions/{author}/{name}/trust", registryHandler.HandleGetFunctionTrustByAuthorName).Methods("GET")

	// Replay (public)
	api.HandleFunc("/replay/{execId}", registryHandler.HandleGetReplay).Methods("GET")

	// ── DRE 2.0 — Certificates & Passports (public) ──────────────────────────
	api.HandleFunc("/functions/{author}/{name}/cert/{cert_id}", dreHandler.HandleGetCertificate).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/cert/{cert_id}/verify", dreHandler.HandleVerifyCertificate).Methods("POST", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/cert/{cert_id}/anchor", dreHandler.HandleAnchorCertificate).Methods("POST", "OPTIONS")
	api.HandleFunc("/dre/anchoring/status", dreHandler.HandleGetAnchoringStatus).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/certs", dreHandler.HandleListCertificates).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/replay/{execution_id}", dreHandler.HandleReplay).Methods("POST", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/passport", dreHandler.HandleGetPassport).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/passport/public", dreHandler.HandleGetPassportPublic).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/diverge", dreHandler.HandleDivergenceSimulation).Methods("POST", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/executions", dreHandler.HandleListExecutions).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/executions/by-hash", dreHandler.HandleGetExecutionByHash).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/executions/timeline", dreHandler.HandleGetExecutionTimeline).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/executions/{execution_id}", dreHandler.HandleGetExecution).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/drift-reports", dreHandler.HandleListDriftReports).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/dre-stats", dreHandler.HandleGetDRESummary).Methods("GET", "OPTIONS")

	// Internal DRE endpoints (for platform services)
	api.HandleFunc("/internal/functions/{function_id}/passport", dreHandler.HandleGetPassportByFunctionID).Methods("GET", "OPTIONS")

	// Execution security (public)
	executionSecurityMW.CreateExecutionSecurityRoutes(api)

	// ── Verification (protected) ─────────────────────────────────────────────
	api.HandleFunc("/verification/{functionVersionId}/status", authMiddleware.RequireAuth(registryHandler.HandleGetVerificationStatus)).Methods("GET")
	api.HandleFunc("/verification/{functionVersionId}/sign", authMiddleware.RequireAuth(registryHandler.HandleSignFunction)).Methods("POST")
	api.HandleFunc("/verification/signatures/{signatureId}/verify", authMiddleware.RequireAuth(registryHandler.HandleVerifySignature)).Methods("POST")
	api.HandleFunc("/verification/{functionVersionId}/approval", authMiddleware.RequireAuth(registryHandler.HandleRequestApproval)).Methods("POST")
	api.HandleFunc("/verification/approvals/{approvalId}/decide", authMiddleware.RequireAuth(registryHandler.HandleMakeApprovalDecision)).Methods("POST")
	api.HandleFunc("/verification/{functionVersionId}/approvals", authMiddleware.RequireAuth(registryHandler.HandleGetApprovals)).Methods("GET")
	api.HandleFunc("/verification/approvals/pending", authMiddleware.RequireAuth(registryHandler.HandleGetPendingApprovals)).Methods("GET")
	api.HandleFunc("/verification/approvals/{approvalId}/comments", authMiddleware.RequireAuth(registryHandler.HandleAddApprovalComment)).Methods("POST")
	api.HandleFunc("/verification/approvals/{approvalId}/comments", authMiddleware.RequireAuth(registryHandler.HandleGetApprovalComments)).Methods("GET")

	// ── Content Management (public) ──────────────────────────────────────────
	api.HandleFunc("/content/changelog", contentHandler.HandleGetPublishedChangelogEntries).Methods("GET")
	api.HandleFunc("/content/blog", contentHandler.HandleGetPublishedBlogPosts).Methods("GET")
	api.HandleFunc("/content/blog/{slug}", contentHandler.HandleGetPublishedBlogPostBySlug).Methods("GET")
	api.HandleFunc("/content/categories", contentHandler.HandleGetBlogCategories).Methods("GET")
	api.HandleFunc("/content/authors", contentHandler.HandleGetBlogAuthors).Methods("GET")

	// ── Blog API (NestJS migration - public) ───────────────────────────────
	api.HandleFunc("/blog/posts", blogHandler.HandleListPosts).Methods("GET")
	api.HandleFunc("/blog/posts/{slug}", blogHandler.HandleGetPostBySlug).Methods("GET")
	api.HandleFunc("/blog/categories", blogHandler.HandleGetCategories).Methods("GET")
	api.HandleFunc("/blog/authors", blogHandler.HandleGetAuthors).Methods("GET")

	// ── Feedback (public submit, protected read) ──────────────────────────────
	api.HandleFunc("/feedback", feedbackHandler.CreateFeedback).Methods("POST")
	api.HandleFunc("/feedback/history", authMiddleware.RequireAuth(feedbackHandler.GetFeedbackHistory)).Methods("GET")
	api.HandleFunc("/feedback/attachments/{id}/download", authMiddleware.RequireAuth(feedbackHandler.DownloadAttachment)).Methods("GET")

	// ── Trending & Discovery ─────────────────────────────────────────────────
	api.HandleFunc("/functions/trending", registryHandler.HandleGetTrendingFunctions).Methods("GET", "OPTIONS")

	// ── Remix/fork functionality (protected) ────────────────────────────────
	api.HandleFunc("/functions/{author}/{name}/remix/cost", authMiddleware.RequireAuth(registryHandler.HandleGetRemixCost)).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/remix", authMiddleware.RequireAuth(registryHandler.HandleRemix)).Methods("POST", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/remix/history", registryHandler.HandleGetRemixHistory).Methods("GET", "OPTIONS")

	// ── Social features (likes) ─────────────────────────────────────────────
	api.HandleFunc("/functions/{author}/{name}/likes", registryHandler.HandleGetFunctionLikes).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/likes", authMiddleware.RequireAuth(registryHandler.HandleLikeFunction)).Methods("POST", "OPTIONS")

	// ── FRG (Function Registry + Live Runtime Graph) ──────────────────────────
	// These routes are registered when FRG is initialized in routes.go
}

// registerAtlasRoutes wires Atlas Memory Engine trace replay endpoints.
func registerAtlasRoutes(
	api *mux.Router,
	authMiddleware *middleware.AuthMiddleware,
	atlasHandler *registryhandler.AtlasHandler,
) {
	api.HandleFunc("/atlas/traces", atlasHandler.HandleListTraces).Methods("GET", "OPTIONS")
	api.HandleFunc("/atlas/traces/health", atlasHandler.HandleHealth).Methods("GET", "OPTIONS")
	api.HandleFunc("/atlas/traces/search", authMiddleware.RequireAuth(atlasHandler.HandleSearchTraces)).Methods("POST", "OPTIONS")
	api.HandleFunc("/atlas/traces/{runId}", atlasHandler.HandleGetTrace).Methods("GET", "OPTIONS")
	api.HandleFunc("/atlas/traces/{runId}/graph", atlasHandler.HandleGetGraph).Methods("GET", "OPTIONS")
}
