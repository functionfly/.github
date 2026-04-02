package api

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/handlers/content"
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
	contentHandler *content.Handler,
	feedbackHandler *feedbackHandlerPkg.Handler,
	recommendationHandler *recommendations.Handler,
) {
	// Inline init for registry-specific sub-handlers
	canaryHandler := registryhandler.NewCanaryHandler(
		storageregistry.NewCanaryConfigRepository(s.postgresDB.GORM),
		registryRepo,
	)
	versionManager := middleware.NewVersionManager()
	deprecationHandler := registryhandler.NewDeprecationHandler(versionManager)
	migrationHandler := registryhandler.NewMigrationHandler()
	dreHandler := drehandler.NewHandler(registryRepo)

	// ── App-based Playground (public) ───────────────────────────────────────
	api.HandleFunc("/run/{appSlug}/{functionName}", appPlaygroundHandler.HandlePlaygroundUI).Methods("GET", "OPTIONS")
	api.HandleFunc("/run/{appSlug}/{functionName}/info", appPlaygroundHandler.HandleGetFunctionInfo).Methods("GET", "OPTIONS")
	api.HandleFunc("/run/{appSlug}/{functionName}/execute", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		app, err := s.repo.GetAppBySlug(vars["appSlug"])
		if err != nil {
			http.Error(w, "App not found", http.StatusNotFound)
			return
		}
		fn, err := s.repo.GetFunctionByAppIDAndName(r.Context(), app.ID, vars["functionName"])
		if err != nil {
			http.Error(w, "Function not found", http.StatusNotFound)
			return
		}
		executionSecurityMW.SecureExecution(fn.ID, fn.Version)(appPlaygroundHandler.HandleExecute).ServeHTTP(w, r)
	}).Methods("POST", "OPTIONS")

	// ── Registry Playground (public) ────────────────────────────────────────
	api.HandleFunc("/fx/{author}/{name}", registryPlaygroundHandler.HandleFunctionPage).Methods("GET", "OPTIONS")
	api.HandleFunc("/run/{author}/{name}", registryPlaygroundHandler.HandlePlaygroundUI).Methods("GET", "OPTIONS")
	api.HandleFunc("/run/{author}/{name}/execute", registryPlaygroundHandler.HandlePlaygroundExecute).Methods("POST", "OPTIONS")
	api.HandleFunc("/run/{author}/{name}/share", registryPlaygroundHandler.HandlePlaygroundShare).Methods("POST", "OPTIONS")
	api.HandleFunc("/replay/{executionId}", registryPlaygroundHandler.HandleReplay).Methods("GET", "OPTIONS")
	api.HandleFunc("/fx/{author}/{name}/code", registryPlaygroundHandler.HandleCodeExamples).Methods("GET", "OPTIONS")
	api.HandleFunc("/fx/{author}/{name}/ai-schema", registryPlaygroundHandler.HandleAIToolSchema).Methods("GET", "OPTIONS")

	// @username clean-URL function routes
	api.HandleFunc("/@/{username}/v1/fx/{functionName}", registryPlaygroundHandler.HandleFunctionPageAt).Methods("GET", "OPTIONS")
	api.HandleFunc("/@/{username}/v1/fx/{functionName}/execute", registryPlaygroundHandler.HandleExecuteAt).Methods("POST", "OPTIONS")
	api.HandleFunc("/@/{username}/v1/fx/{functionName}/v/{version}", registryPlaygroundHandler.HandleFunctionPageAtVersion).Methods("GET", "OPTIONS")
	api.HandleFunc("/@/{username}/v1/fx/{functionName}/versions", registryHandler.HandleListVersionsAt).Methods("GET", "OPTIONS")
	api.HandleFunc("/@/{username}/v1/fx/{functionName}/stats", registryHandler.HandleGetFunctionStatsAt).Methods("GET", "OPTIONS")

	// ── Registry CRUD (public read) ──────────────────────────────────────────
	api.HandleFunc("/registry/functions", registryHandler.HandleListFunctions).Methods("GET")
	api.HandleFunc("/registry/functions", registryHandler.HandleDeleteAllFunctions).Methods("DELETE")
	api.HandleFunc("/registry/functions/{author}/{name}", registryHandler.HandleGetFunction).Methods("GET")
	api.HandleFunc("/registry/functions/{author}/{name}", registryHandler.HandleDeleteFunction).Methods("DELETE")
	api.HandleFunc("/registry/functions/{author}/{name}/versions", registryHandler.HandleListVersions).Methods("GET")
	api.HandleFunc("/registry/functions/{author}/{name}/changelogs", registryHandler.HandleGetChangelogs).Methods("GET")
	api.HandleFunc("/registry/functions/{author}/{name}/changelogs/{version}", registryHandler.HandleGetChangelogByVersion).Methods("GET")
	api.HandleFunc("/registry/functions/{author}/{name}/changelogs/category/{category}", registryHandler.HandleGetChangelogByCategory).Methods("GET")
	api.HandleFunc("/registry/functions/{author}/{name}/history", registryHandler.HandleGetVersionHistory).Methods("GET")
	api.HandleFunc("/registry/search", registryHandler.HandleSearchFunctions).Methods("GET")

	// ── Function Versions ────────────────────────────────────────────────────
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
	api.HandleFunc("/internal/contracts", versionHandler.HandleListServiceContracts).Methods("GET", "OPTIONS")
	api.HandleFunc("/internal/contracts/{service}", versionHandler.HandleGetServiceContracts).Methods("GET", "OPTIONS")
	api.HandleFunc("/internal/contracts/negotiate", versionHandler.HandleNegotiateContractVersion).Methods("POST", "OPTIONS")

	// ── Recommendations (public) ─────────────────────────────────────────────
	api.HandleFunc("/recommendations", recommendationHandler.HandleGetRecommendations).Methods("GET", "OPTIONS")
	api.HandleFunc("/recommendations/interactions", recommendationHandler.HandleRecordInteraction).Methods("POST", "OPTIONS")
	api.HandleFunc("/recommendations/executions", recommendationHandler.HandleRecordExecution).Methods("POST", "OPTIONS")
	api.HandleFunc("/recommendations/feedback", recommendationHandler.HandleRecordFeedback).Methods("POST", "OPTIONS")
	api.HandleFunc("/recommendations/refresh", authMiddleware.RequirePermission(auth.PermSystemWrite)(recommendationHandler.HandleRefreshRecommendations)).Methods("POST", "OPTIONS")
	api.HandleFunc("/recommendations/triple-search", recommendationHandler.HandleTripleSearch).Methods("POST", "OPTIONS")
	api.HandleFunc("/recommendations/composable/{function_id}", recommendationHandler.HandleFindComposable).Methods("GET", "OPTIONS")

	// ── Registry v2 ──────────────────────────────────────────────────────────
	apiV2.HandleFunc("/registry/functions", registryHandler.HandleListFunctions).Methods("GET")
	apiV2.HandleFunc("/registry/functions/{author}/{name}", registryHandler.HandleGetFunction).Methods("GET")
	apiV2.HandleFunc("/registry/functions/{author}/{name}/versions", registryHandler.HandleListVersions).Methods("GET")
	apiV2.HandleFunc("/registry/functions/{author}/{name}/changelogs", registryHandler.HandleGetChangelogs).Methods("GET")
	apiV2.HandleFunc("/registry/functions/{author}/{name}/changelogs/{version}", registryHandler.HandleGetChangelogByVersion).Methods("GET")
	apiV2.HandleFunc("/registry/functions/{author}/{name}/history", registryHandler.HandleGetVersionHistory).Methods("GET")
	apiV2.HandleFunc("/registry/search", registryHandler.HandleSearchFunctions).Methods("GET")

	// ── Canary Deployments ───────────────────────────────────────────────────
	api.HandleFunc("/registry/functions/{author}/{name}/canary", canaryHandler.HandleCreateCanary).Methods("POST")
	api.HandleFunc("/registry/functions/{author}/{name}/canary", canaryHandler.HandleGetCanary).Methods("GET")
	api.HandleFunc("/registry/functions/{author}/{name}/canary", canaryHandler.HandleUpdateCanary).Methods("PATCH")
	api.HandleFunc("/registry/functions/{author}/{name}/canary", canaryHandler.HandleCancelCanary).Methods("DELETE")
	api.HandleFunc("/registry/functions/{author}/{name}/canary/promote", canaryHandler.HandlePromoteCanary).Methods("POST")
	api.HandleFunc("/registry/functions/{author}/{name}/canary/rollback", canaryHandler.HandleRollbackCanary).Methods("POST")
	api.HandleFunc("/registry/functions/{author}/{name}/canary/history", canaryHandler.HandleGetCanaryHistory).Methods("GET")

	// ── Deprecation & Migration Guides ───────────────────────────────────────
	api.HandleFunc("/registry/deprecations", deprecationHandler.HandleGetAllDeprecations).Methods("GET")
	api.HandleFunc("/registry/deprecations/{endpoint}", deprecationHandler.HandleGetEndpointDeprecation).Methods("GET")
	api.HandleFunc("/registry/migration-guide", migrationHandler.HandleGetMigrationGuide).Methods("GET")
	api.HandleFunc("/registry/migration-guide/{endpoint}", migrationHandler.HandleGetEndpointMigration).Methods("GET")
	api.HandleFunc("/registry/versions", migrationHandler.HandleGetVersionInfo).Methods("GET")

	// ── Documentation (public) ───────────────────────────────────────────────
	api.HandleFunc("/docs", docsHandler.HandleIndex).Methods("GET")
	api.HandleFunc("/docs/{author}/{name}", docsHandler.HandleFunctionHTMLDocs).Methods("GET")
	api.HandleFunc("/docs/{author}/{name}/versions", docsHandler.HandleFunctionVersions).Methods("GET")
	api.HandleFunc("/docs/openapi.json", docsHandler.HandleOpenAPISpec).Methods("GET")
	api.HandleFunc("/docs/{author}/{name}/api", docsHandler.HandleFunctionDocs).Methods("GET")

	// Legacy playground routes
	api.HandleFunc("/playground/{author}/{name}", registryPlaygroundHandler.HandlePlaygroundUI).Methods("GET")
	api.HandleFunc("/playground/{author}/{name}/execute", registryPlaygroundHandler.HandlePlaygroundExecute).Methods("POST")
	api.HandleFunc("/playground/{author}/{name}/share", registryPlaygroundHandler.HandlePlaygroundShare).Methods("POST")

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

	// ── Embed (public read, protected write) ─────────────────────────────────
	api.HandleFunc("/embed/{author}/{nameVersion}", registryHandler.HandleServeEmbed).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/functions/{author}/{name}/embed", registryHandler.HandleGetEmbedConfig).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/functions/{author}/{name}/embed/snippet", registryHandler.HandleGetEmbedSnippet).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/functions/{author}/{name}/embed/analytics", registryHandler.HandleGetEmbedAnalytics).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/functions/{author}/{name}/embed", authMiddleware.RequireAuth(registryHandler.HandleUpdateEmbedConfig)).Methods("PUT", "OPTIONS")

	// Function settings (protected)
	api.HandleFunc("/functions/{author}/{name}/settings", authMiddleware.RequireAuth(registryHandler.HandleGetFunctionSettings)).Methods("GET", "OPTIONS")
	api.HandleFunc("/functions/{author}/{name}/settings", authMiddleware.RequireAuth(registryHandler.HandlePatchFunctionSettings)).Methods("PATCH", "OPTIONS")

	// Cache monitoring (public)
	api.HandleFunc("/cache/stats", registryHandler.HandleGetCacheStats).Methods("GET")

	// ── Execute /fx/ with security + optional verification ───────────────────
	secureExecuteHandler := func(w http.ResponseWriter, r *http.Request) {
		repo := storageregistry.NewRegistryRepository(s.postgresDB.GORM, nil)
		vars := mux.Vars(r)
		fn, err := repo.GetFunctionByAuthorName(vars["author"], vars["name"])
		if err != nil {
			http.Error(w, "Function not found", http.StatusNotFound)
			return
		}
		executionSecurityMW.SecureExecution(fn.ID, vars["version"])(registryHandler.HandleExecute).ServeHTTP(w, r)
	}
	if verificationMiddleware != nil {
		api.Handle("/fx/{author}/{name}", verificationMiddleware.RequireVerifiedFunction("standard")(http.HandlerFunc(secureExecuteHandler))).Methods("POST", "OPTIONS")
		api.Handle("/fx/{author}/{name}@{version}", verificationMiddleware.RequireVerifiedFunction("standard")(http.HandlerFunc(secureExecuteHandler))).Methods("POST", "OPTIONS")
	} else {
		api.HandleFunc("/fx/{author}/{name}", secureExecuteHandler).Methods("POST", "OPTIONS")
		api.HandleFunc("/fx/{author}/{name}@{version}", secureExecuteHandler).Methods("POST", "OPTIONS")
	}

	// Publish (protected)
	api.HandleFunc("/registry/publish", authMiddleware.RequireAuth(registryHandler.HandlePublish)).Methods("POST")

	// Stats / test / rating (public or protected)
	api.HandleFunc("/registry/functions/{author}/{name}/stats", registryHandler.HandleGetFunctionStats).Methods("GET")
	api.HandleFunc("/registry/functions/{author}/{name}/test", registryHandler.HandleTest).Methods("POST")
	api.HandleFunc("/registry/functions/{author}/{name}/rating", authMiddleware.RequireAuth(registryHandler.HandleSubmitRating)).Methods("POST")
	api.HandleFunc("/registry/functions/{author}/{name}/reviews", registryHandler.HandleListReviews).Methods("GET")
	api.HandleFunc("/registry/functions/{author}/{name}/reviews", authMiddleware.RequireAuth(registryHandler.HandleSubmitReview)).Methods("POST")
	api.HandleFunc("/registry/functions/{author}/{name}/aggregate", authMiddleware.RequireAuth(registryHandler.HandleAggregateStats)).Methods("POST")

	// ── Trust Scoring (public) ──────────────────────────────────────────────
	api.HandleFunc("/registry/functions/{author}/{name}/trust", registryHandler.HandleGetFunctionTrustByAuthorName).Methods("GET")
	api.HandleFunc("/functions/{functionId}/trust", registryHandler.HandleGetTrustScore).Methods("GET")
	api.HandleFunc("/functions/{functionId}/trust/history", registryHandler.HandleGetTrustHistory).Methods("GET")
	api.HandleFunc("/functions/{functionId}/trust/refresh", authMiddleware.RequireAuth(registryHandler.HandleRefreshTrustScore)).Methods("POST")

	// Replay (public)
	api.HandleFunc("/registry/replay/{execId}", registryHandler.HandleGetReplay).Methods("GET")

	// ── DRE 2.0 — Certificates & Passports (public) ──────────────────────────
	api.HandleFunc("/registry/{author}/{name}/cert/{cert_id}", dreHandler.HandleGetCertificate).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/cert/{cert_id}/verify", dreHandler.HandleVerifyCertificate).Methods("POST", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/cert/{cert_id}/anchor", dreHandler.HandleAnchorCertificate).Methods("POST", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/certs", dreHandler.HandleListCertificates).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/replay/{execution_id}", dreHandler.HandleReplay).Methods("POST", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/passport", dreHandler.HandleGetPassport).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/passport/public", dreHandler.HandleGetPassportPublic).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/diverge", dreHandler.HandleDivergenceSimulation).Methods("POST", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/executions", dreHandler.HandleListExecutions).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/executions/by-hash", dreHandler.HandleGetExecutionByHash).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/executions/timeline", dreHandler.HandleGetExecutionTimeline).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/executions/{execution_id}", dreHandler.HandleGetExecution).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/drift-reports", dreHandler.HandleListDriftReports).Methods("GET", "OPTIONS")
	api.HandleFunc("/registry/{author}/{name}/dre-stats", dreHandler.HandleGetDRESummary).Methods("GET", "OPTIONS")

	// Internal DRE endpoints (for platform services)
	api.HandleFunc("/internal/functions/{function_id}/passport", dreHandler.HandleGetPassportByFunctionID).Methods("GET", "OPTIONS")

	// Execution security (public)
	executionSecurityMW.CreateExecutionSecurityRoutes(api)

	// ── Verification (protected) ─────────────────────────────────────────────
	api.HandleFunc("/registry/verification/{functionVersionId}/status", authMiddleware.RequireAuth(registryHandler.HandleGetVerificationStatus)).Methods("GET")
	api.HandleFunc("/registry/verification/{functionVersionId}/sign", authMiddleware.RequireAuth(registryHandler.HandleSignFunction)).Methods("POST")
	api.HandleFunc("/registry/verification/signatures/{signatureId}/verify", authMiddleware.RequireAuth(registryHandler.HandleVerifySignature)).Methods("POST")
	api.HandleFunc("/registry/verification/{functionVersionId}/approval", authMiddleware.RequireAuth(registryHandler.HandleRequestApproval)).Methods("POST")
	api.HandleFunc("/registry/verification/approvals/{approvalId}/decide", authMiddleware.RequireAuth(registryHandler.HandleMakeApprovalDecision)).Methods("POST")
	api.HandleFunc("/registry/verification/{functionVersionId}/approvals", authMiddleware.RequireAuth(registryHandler.HandleGetApprovals)).Methods("GET")
	api.HandleFunc("/registry/verification/approvals/pending", authMiddleware.RequireAuth(registryHandler.HandleGetPendingApprovals)).Methods("GET")
	api.HandleFunc("/registry/verification/approvals/{approvalId}/comments", authMiddleware.RequireAuth(registryHandler.HandleAddApprovalComment)).Methods("POST")
	api.HandleFunc("/registry/verification/approvals/{approvalId}/comments", authMiddleware.RequireAuth(registryHandler.HandleGetApprovalComments)).Methods("GET")

	// ── Content Management (public) ──────────────────────────────────────────
	api.HandleFunc("/content/changelog", contentHandler.HandleGetPublishedChangelogEntries).Methods("GET")
	api.HandleFunc("/content/blog", contentHandler.HandleGetPublishedBlogPosts).Methods("GET")
	api.HandleFunc("/content/blog/{slug}", contentHandler.HandleGetPublishedBlogPostBySlug).Methods("GET")
	api.HandleFunc("/content/categories", contentHandler.HandleGetBlogCategories).Methods("GET")
	api.HandleFunc("/content/authors", contentHandler.HandleGetBlogAuthors).Methods("GET")

	// ── Feedback (public submit, protected read) ──────────────────────────────
	api.HandleFunc("/feedback", feedbackHandler.CreateFeedback).Methods("POST")
	api.HandleFunc("/feedback/history", authMiddleware.RequireAuth(feedbackHandler.GetFeedbackHistory)).Methods("GET")
	api.HandleFunc("/feedback/attachments/{id}/download", authMiddleware.RequireAuth(feedbackHandler.DownloadAttachment)).Methods("GET")
}
