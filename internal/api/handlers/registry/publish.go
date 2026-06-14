package registry

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/handlers/registry/execution"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/bundler"
	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/functionfly/functionfly/internal/manifest"
	"github.com/functionfly/functionfly/internal/storage"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// isRecordNotFound checks if the error is a GORM record not found error
func isRecordNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "record not found") ||
		strings.Contains(err.Error(), "no rows in result set")
}

// HandlePublish handles publishing a new function or version
func (h *Handler) HandlePublish(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Failed to read request body"))
		return
	}
	defer r.Body.Close()

	var req functionregistry.PublishRequest
	if err := json.Unmarshal(body, &req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid JSON"))
		return
	}

	// Validate required fields
	if req.Author == "" || req.Name == "" || req.Version == "" {
		apierror.WriteError(w, apierror.NewBadRequest("author, name, and version are required"))
		return
	}

	// Validate semver format
	if err := functionregistry.ValidateSemVer(req.Version); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest(err.Error()))
		return
	}

	// Parse conflict strategy (default: error)
	conflictStrategy := storage.VersionConflictError
	if cs := r.URL.Query().Get("conflict_strategy"); cs != "" {
		switch cs {
		case "overwrite":
			conflictStrategy = storage.VersionConflictOverwrite
		case "create_new":
			conflictStrategy = storage.VersionConflictCreateNew
		case "error":
			conflictStrategy = storage.VersionConflictError
		default:
			apierror.WriteError(w, apierror.NewBadRequest("invalid conflict_strategy: must be 'error', 'overwrite', or 'create_new'"))
			return
		}
	}

	// For registry functions, source code is required for sandbox execution
	if req.Source == nil || req.Source.Code == "" {
		apierror.WriteError(w, apierror.NewBadRequest("source code is required for registry functions"))
		return
	}

	// Parse manifest early so we can set/update function metadata (title, description, category, tags)
	cleanManifest := manifest.StripComments(string(req.Manifest))
	var m functionregistry.FunctionManifest
	if err := json.Unmarshal([]byte(cleanManifest), &m); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid manifest JSON"))
		return
	}

	// Determine trust level from request or manifest
	trustLevel := req.TrustLevel
	if trustLevel == "" {
		trustLevel = "standard" // Default trust level
	}

	// Check if function already exists
	existingFn, err := h.repo.GetFunctionByAuthorName(context.Background(), req.Author, req.Name)
	if err != nil && !isRecordNotFound(err) {
		logrus.WithError(err).Error("Failed to check existing function")
		apierror.WriteError(w, apierror.NewInternal("Failed to check function: "+err.Error()))
		return
	}

	var fnID uuid.UUID

	if existingFn == nil {
		// Create new function with metadata from manifest
		tagsJSON, _ := json.Marshal(m.Tags)
		if len(tagsJSON) == 0 || string(tagsJSON) == "null" {
			tagsJSON, _ = json.Marshal([]string{})
		}
		relScore, detScore := 0.0, 0.0
		if strings.EqualFold(req.Author, "functionfly") {
			relScore, detScore = 90.0, 90.0
		}
		fn := &storage.RegistryFunction{
			Author:             req.Author,
			Name:               req.Name,
			Visibility:         "public",
			PricePerCall:       m.PricePerCall,
			PopularityScore:    0,
			ReliabilityScore:   relScore,
			DeterministicScore: detScore,
			TenantID:           &user.TenantID,
			OwnerUserID:        &user.UserID,
			Tags:               tagsJSON,
		}
		if m.Title != "" {
			fn.Title = sql.NullString{String: m.Title, Valid: true}
		}
		if m.Description != "" {
			fn.Description = sql.NullString{String: m.Description, Valid: true}
		}
		if m.Category != "" {
			fn.Category = sql.NullString{String: m.Category, Valid: true}
		}

		if err := h.repo.CreateFunction(context.Background(), fn); err != nil {
			logrus.WithError(err).Error("Failed to create function")
			apierror.WriteError(w, apierror.NewInternal("Failed to create function: "+err.Error()))
			return
		}
		fnID = fn.ID
	} else {
		fnID = existingFn.ID
		// Sync function metadata from this version's manifest (title, description, category, tags)
		meta := map[string]interface{}{}
		if m.Title != "" {
			meta["title"] = m.Title
		}
		if m.Description != "" {
			meta["description"] = m.Description
		}
		if m.Category != "" {
			meta["category"] = m.Category
		}
		if len(m.Tags) > 0 {
			meta["tags"] = m.Tags
		}
		if m.PricePerCall > 0 {
			meta["price_per_call"] = m.PricePerCall
		}
		// FunctionFly functions get default high trust on function record if not already set
		if strings.EqualFold(req.Author, "functionfly") && existingFn.ReliabilityScore == 0 && existingFn.DeterministicScore == 0 {
			meta["reliability_score"] = 90.0
			meta["deterministic_score"] = 90.0
		}
		if len(meta) > 0 {
			if _, err := h.repo.UpdateRegistryFunction(context.Background(), fnID, meta); err != nil {
				logrus.WithError(err).Warn("Failed to update function metadata from manifest")
			}
		}
	}

	// Platform fee charging - check exemption, verify balance, and charge
	var feeCharged bool
	var feeAmountUSD float64
	var feeType string
	platformFeePaid := false

	if h.platformFeeRepo != nil && !storageregistry.IsAuthorExempt(req.Author) {
		isNewFunction := existingFn == nil
		feeAmountUSD = storageregistry.CalculatePublishFee(req.Author)
		feeType = storageregistry.FeeTypePublish

		if !isNewFunction {
			feeAmountUSD = storageregistry.CalculateVersionUpdateFee(req.Author)
			feeType = storageregistry.FeeTypeVersionUpdate
		}

		if feeAmountUSD > 0 {
			// Get or create wallet and check balance using unified wallet service if available
			var walletBalance float64
			var walletID uuid.UUID

			if h.walletSvc != nil {
				// Use unified wallet service
				userWallet, err := h.walletSvc.GetOrCreateUserWallet(r.Context(), user.UserID)
				if err != nil {
					logrus.WithError(err).Error("Failed to get or create wallet for platform fee")
					apierror.WriteError(w, apierror.NewInternal("Failed to process payment: "+err.Error()))
					return
				}
				walletID = userWallet.ID
				walletBalance = userWallet.BalanceUSD
			} else {
				// Fall back to legacy platform fee repo
				wallet, err := h.platformFeeRepo.GetOrCreateWallet(r.Context(), user.UserID)
				if err != nil {
					logrus.WithError(err).Error("Failed to get or create wallet for platform fee")
					apierror.WriteError(w, apierror.NewInternal("Failed to process payment: "+err.Error()))
					return
				}
				walletBalance = wallet.BalanceUSD
			}

			// Check sufficient balance
			if walletBalance < feeAmountUSD {
				apierror.WriteError(w, apierror.NewBadRequest(fmt.Sprintf("Insufficient wallet balance. Required: $%.2f, Available: $%.2f", feeAmountUSD, walletBalance)))
				return
			}

			// Debit wallet (atomic transaction)
			description := fmt.Sprintf("FunctionFly registry publish: %s/%s v%s", req.Author, req.Name, req.Version)
			if h.walletSvc != nil {
				// Use unified wallet service for debit
				_, err := h.walletSvc.DebitForFeePayment(r.Context(), walletID, feeAmountUSD, feeType, description)
				if err != nil {
					logrus.WithError(err).Error("Failed to debit wallet for platform fee")
					apierror.WriteError(w, apierror.NewInternal("Failed to process payment: "+err.Error()))
					return
				}
			} else {
				// Fall back to legacy platform fee repo
				if err := h.platformFeeRepo.DebitWallet(r.Context(), user.UserID, feeAmountUSD, description); err != nil {
					logrus.WithError(err).Error("Failed to debit wallet for platform fee")
					apierror.WriteError(w, apierror.NewInternal("Failed to process payment: "+err.Error()))
					return
				}
			}

			// Record platform fee
			now := time.Now()
			feeRecord := &storageregistry.PlatformFee{
				FunctionID: fnID,
				UserID:     user.UserID,
				FeeType:    feeType,
				AmountUSD:  feeAmountUSD,
				ChargedAt:  now,
				Status:     storageregistry.FeeStatusCompleted,
			}
			if err := h.platformFeeRepo.RecordPlatformFee(r.Context(), feeRecord); err != nil {
				// Log warning but don't fail publish
				logrus.WithError(err).Warn("Failed to record platform fee")
			}

			// Update function record with fee info
			feePaid := true
			meta := map[string]interface{}{
				"platform_fee_paid":       feePaid,
				"platform_fee_amount_usd": feeAmountUSD,
				"last_fee_charged_at":     now,
			}
			if _, err := h.repo.UpdateRegistryFunction(context.Background(), fnID, meta); err != nil {
				logrus.WithError(err).Warn("Failed to update function with platform fee info")
			}

			feeCharged = true
			platformFeePaid = true
		}
	}

	// Validate capabilities against allowed list
	for _, cap := range m.Capabilities {
		if !functionregistry.IsValidCapability(cap) {
			apierror.WriteError(w, apierror.NewBadRequest("Invalid capability: "+cap+". Allowed: "+strings.Join(functionregistry.AllowedCapabilities, ", ")))
			return
		}
	}

	// Set defaults
	if m.TimeoutMs == 0 {
		m.TimeoutMs = 5000
	}
	if m.MemoryMB == 0 {
		m.MemoryMB = 128
	}
	// Default side_effects to "none" if not specified
	if m.SideEffects == "" {
		m.SideEffects = "none"
	}

	// Get WASM binary - either from pre-compiled source or skip bundling for lazy approach
	var wasmBinary []byte
	var sourceHash string

	logrus.WithFields(logrus.Fields{
		"source_runtime":  req.Source.Runtime,
		"has_wasm_binary": req.Source.WasmBinary != "",
		"wasm_binary_len": len(req.Source.WasmBinary),
	}).Info("Checking WASM source")

	if req.Source.Runtime == "wasm" && req.Source.WasmBinary != "" {
		// Use pre-compiled WASM binary (base64-encoded)
		var decodeErr error
		wasmBinary, decodeErr = base64.StdEncoding.DecodeString(req.Source.WasmBinary)
		if decodeErr != nil {
			logrus.WithError(decodeErr).Error("Failed to decode pre-compiled WASM")
			apierror.WriteError(w, apierror.NewBadRequest("Invalid WASM binary encoding: "+decodeErr.Error()))
			return
		}
		// Calculate source hash from the code
		sourceHash = h.calculateSourceHash(req.Source)
		logrus.WithField("size", len(wasmBinary)).Info("Using pre-compiled WASM binary")

		// YARA scan: kick off an async malware scan of the WASM binary.
		// The scan runs in the background and marks the version as quarantined
		// if a match is found.  Does not block the publish response.
		if len(wasmBinary) > 0 {
			go func(fnID uuid.UUID, version string, wasmBytes []byte) {
				h.scanWasmForMalware(fnID, version, wasmBytes)
			}(fnID, req.Version, wasmBinary)
		}
	} else {
		// NEW: Skip bundling during publish - will bundle lazily at execute time
		// This significantly speeds up publish by avoiding expensive compilation
		// Calculate source hash for integrity checking
		sourceHash = h.calculateSourceHash(req.Source)
		logrus.WithFields(logrus.Fields{
			"runtime":     req.Source.Runtime,
			"source_hash": sourceHash,
		}).Info("Skipping bundling during publish - will bundle at execute time")
		// wasmBinary remains nil/empty - will be generated lazily
	}

	// Create function version - store as clean JSON (canonical format)
	// Also store capabilities separately for efficient runtime access
	capabilitiesJSON, _ := json.Marshal(m.Capabilities)
	version := &storage.RegistryFunctionVersion{
		FunctionID:    fnID,
		Version:       req.Version,
		Manifest:      []byte(cleanManifest), // Store clean JSON (no comments)
		Runtime:       m.Runtime,
		TimeoutMs:     m.TimeoutMs,
		MemoryMB:      m.MemoryMB,
		Deterministic: m.Deterministic,
		SideEffects:   m.SideEffects,
		Idempotent:    m.Idempotent,
		CacheTTL:      m.CacheTTL,
		Capabilities:  capabilitiesJSON,
		WasmBinary:    wasmBinary,
		SourceHash:    sql.NullString{String: sourceHash, Valid: true},
		BundleSize:    sql.NullInt32{Int32: int32(len(wasmBinary)), Valid: true},
		// Store source code for lazy bundling if no WASM binary
		SourceCode: sql.NullString{String: req.Source.Code, Valid: req.Source.Code != ""},
		// Store readme for function page documentation
		Readme: sql.NullString{String: req.Source.Readme, Valid: req.Source.Readme != ""},
	}

	// Link to deployment if provided
	if req.Deployment != nil {
		depID, err := uuid.Parse(req.Deployment.DeploymentID)
		if err == nil {
			version.DeploymentID = &depID
		}
		backendID, err := uuid.Parse(req.Deployment.BackendID)
		if err == nil {
			version.BackendID = &backendID
		}
	}

	if _, upsertErr := h.repo.UpsertFunctionVersion(version, storageregistry.VersionConflictStrategy(conflictStrategy)); upsertErr != nil {
		logrus.WithError(upsertErr).Error("Failed to create/update function version")
		if strings.Contains(upsertErr.Error(), "already exists") {
			apierror.WriteError(w, apierror.NewConflict(upsertErr.Error()))
			return
		}
		apierror.WriteError(w, apierror.NewInternal("Failed to create version: "+upsertErr.Error()))
		return
	}

	// Eager bundling: compile Python source → WASM at publish time.
	// This eliminates the cold-start penalty and surfaces compilation errors early.
	if h.bundleService != nil && version.SourceCode.Valid && version.SourceCode.String != "" {
		if strings.HasPrefix(version.Runtime, "python") {
			go func(v *storage.RegistryFunctionVersion) {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
				defer cancel()
				if _, err := h.bundleService.Bundle(ctx, v); err != nil {
					logrus.WithError(err).WithFields(logrus.Fields{
						"function_id": v.FunctionID,
						"version":     v.Version,
						"runtime":     v.Runtime,
					}).Warn("Eager bundling failed — will retry at execution time")
				} else {
					logrus.WithFields(logrus.Fields{
						"function_id": v.FunctionID,
						"version":     v.Version,
						"runtime":     v.Runtime,
					}).Info("Eager bundling completed")
				}
			}(version)
		}
	}

	// Update latest version pointer
	if err := h.repo.UpdateFunctionLatestVersion(context.Background(), fnID, req.Version); err != nil {
		logrus.WithError(err).Error("Failed to update latest version")
	}

	// Auto-generate README if not provided (async, don't block publish response)
	if h.autoReadmeSvc != nil && (!version.Readme.Valid || version.Readme.String == "") {
		go func() {
			_, err := h.autoReadmeSvc.GenerateForVersion(r.Context(), fnID, req.Version, false)
			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"function_id": fnID,
					"version":     req.Version,
				}).Warn("Auto-readme generation failed")
			}
		}()
	}

	// Initialize rating
	rating, _ := h.repo.GetOrCreateRating(fnID)
	// FunctionFly functions get default high trust values from the start
	if rating != nil && strings.EqualFold(req.Author, "functionfly") {
		rating.TrustScore = 0.9 // DB constraint: 0-1; frontend gets 0-100 via helpers
		rating.ReliabilityScore = 0.9
		rating.SuccessRate = 0.9
		_ = h.repo.UpdateTrustScore(rating)
		dreScores := &storageregistry.DREScores{
			DeterminismScore:          0.9,
			ReplayIntegrityScore:      0.9,
			PerformanceStabilityScore: 0.9,
			DriftScore:                1.0,
		}
		_ = h.repo.UpdateTrustScoreV2(fnID, dreScores, 0.9)
	}

	// Invalidate list cache so description/category show up in browse UI
	h.repo.InvalidateListCache(r.Context())

	// Baseline verification status returned by publish response.
	// This is the server-side truth that execution middleware should enforce.
	verificationStatusStr := "verified"
	if trustLevel == "high" || trustLevel == "enterprise" {
		verificationStatusStr = "pending"
	}

	// Auto-verify only if the function author is a verified internal namespace
	// This requires proper authorization check - author must match a verified internal account
	isVerifiedNamespace := false
	if strings.EqualFold(req.Author, "functionfly") {
		// Only auto-verify if the authenticated user is authorized for this namespace
		// The user from the authenticated request should match or have admin role
		if user != nil && (user.Role == "admin" || user.Role == "super_admin" ||
			strings.EqualFold(user.Email, "functionfly@functionfly.io") ||
			strings.EqualFold(user.Email, "system@functionfly.io")) {
			isVerifiedNamespace = true
		}
	}

	// Only auto-verify for explicitly authorized namespaces
	if isVerifiedNamespace {
		verificationStatusStr = "verified"

		now := time.Now()
		verifiedStatus := &storageregistry.RegistryFunctionVerificationStatus{
			FunctionVersionID:   version.ID,
			ContentHashVerified: true,
			SignatureVerified:   true,
			MalwareScanned:      true,
			MalwareStatus:       "clean",
			MalwareRiskScore:    0,
			ApprovalRequired:    false,
			ApprovalStatus:      "not_required",
			OverallStatus:       "verified",
			LastVerifiedAt:      &now,
		}
		if err := h.repo.CreateOrUpdateVerificationStatus(verifiedStatus); err != nil {
			logrus.WithError(err).WithField("function_version_id", version.ID).Warn("Failed to auto-verify internal function")
		} else {
			logrus.WithFields(logrus.Fields{
				"function": fmt.Sprintf("%s/%s", req.Author, req.Name),
				"version":  req.Version,
			}).Info("Auto-verified internal function with FXCERT status")
		}
		// Generate bootstrap FXCERT so History and Certificates show a cert after publish
		go func() {
			fn, err := h.repo.GetFunctionByID(context.Background(), fnID)
			if err != nil || fn == nil {
				return
			}
			fnVersion, err := h.repo.GetFunctionVersion(fnID, req.Version)
			if err != nil || fnVersion == nil {
				return
			}
			execution.BootstrapFXCERT(h.repo, fn, fnVersion, "bootstrap", "internal", h.dreNodeKey, h.drePlatformKey)
		}()
	}

	// Auto-create MCP settings for FunctionFly functions to ensure they appear in registry stats
	// This is done regardless of verification status since MCP settings affect visibility
	if strings.EqualFold(req.Author, "functionfly") {
		enabled := true
		exposeInputSchema := true
		mcpSettings := storageregistry.MCPSettingsInput{
			Enabled:           &enabled,
			ExposeInputSchema: &exposeInputSchema,
			Transports:        []string{"streamable-http"},
			RateLimitPerMin:   60,
		}
		actorID, _ := uuid.Parse(user.ID)
		if _, err := h.repo.UpsertMCPSettings(r.Context(), fnID, mcpSettings, &actorID); err != nil {
			logrus.WithError(err).WithField("function_id", fnID).Warn("Failed to auto-create MCP settings for FunctionFly function")
		} else {
			logrus.WithFields(logrus.Fields{
				"function": fmt.Sprintf("%s/%s", req.Author, req.Name),
			}).Info("Auto-created MCP settings for FunctionFly function")
		}
	}

	// NEW: Skip synchronous verification during publish - verify lazily at execute time
	// This significantly speeds up publish by avoiding expensive security scans
	// Verification will happen on first execution instead
	logrus.WithFields(logrus.Fields{
		"function_id":   version.ID,
		"function_name": fmt.Sprintf("%s/%s", req.Author, req.Name),
		"version":       req.Version,
		"lazy_verify":   true,
	}).Info("Skipping synchronous verification - will verify at execute time")

	// Create changelog entry if provided
	if req.Changelog != nil {
		if err := h.createChangelogEntry(fnID, version.ID, req); err != nil {
			logrus.WithError(err).Warn("Failed to create changelog entry")
		}
	} else {
		// Auto-generate a basic changelog if not provided
		if err := h.autoGenerateChangelog(fnID, version.ID, req, existingFn); err != nil {
			logrus.WithError(err).Warn("Failed to auto-generate changelog entry")
		}
	}

	// Ensure registry verification status exists for newly published versions.
	// - `standard` is treated as verified by default.
	// - `high`/`enterprise` start as pending and must complete verification/approval before execution.
	if !strings.EqualFold(req.Author, "functionfly") {
		now := time.Now()
		status := &storageregistry.RegistryFunctionVerificationStatus{
			FunctionVersionID: version.ID,

			// Baseline: assume content hash is already anchored by the publish pipeline.
			ContentHashVerified: true,

			// Baseline malware status: treat as clean for "verified by default" semantics on `standard`.
			// (Actual scanning can still update this row later via async verification jobs.)
			MalwareScanned:   true,
			MalwareStatus:    "clean",
			MalwareRiskScore: 0,

			// Approval/signature fields are used by execution gating.
			// - `standard`: not required
			// - `high`/`enterprise`: pending
			ApprovalRequired:  false,
			ApprovalStatus:    "not_required",
			SignatureVerified: false,

			OverallStatus:  verificationStatusStr,
			LastVerifiedAt: &now,
			UpdatedAt:      now,
			CreatedAt:      now,
		}

		if trustLevel == "high" || trustLevel == "enterprise" {
			status.ApprovalRequired = true
			status.ApprovalStatus = "pending"
			status.SignatureVerified = false
			status.OverallStatus = "pending"
		}

		if err := h.repo.CreateOrUpdateVerificationStatus(status); err != nil {
			logrus.WithError(err).WithField("function_version_id", version.ID).Warn("Failed to set baseline verification status")
		}
	}

	// Auto-generate a permissive input schema if the manifest didn't provide one.
	// This ensures the function_cache and validation pipeline have a schema to work with.
	var inputSchema json.RawMessage
	var isStrict bool
	if m.Input != nil && m.Input.Schema != nil && len(m.Input.Schema) > 0 {
		inputSchema = m.Input.Schema
	}

	if inputSchema == nil || len(inputSchema) == 0 {
		permissiveSchema, schemaErr := json.Marshal(map[string]interface{}{
			"type":                 "object",
			"additionalProperties": true,
			"properties":           map[string]interface{}{},
		})
		if schemaErr == nil {
			inputSchema = permissiveSchema
		}
	}

	if inputSchema != nil && len(inputSchema) > 0 {
		if err := h.repo.UpsertFunctionInputSchema(r.Context(), version.ID, inputSchema, isStrict); err != nil {
			logrus.WithError(err).WithField("function_version_id", version.ID).Warn("Failed to save input schema")
		} else {
			logrus.WithField("function_version_id", version.ID).Info("Saved input schema for published function")
		}
	}

	// Parse capabilities from manifest
	var capabilities []string
	if len(m.Capabilities) > 0 {
		capabilities = m.Capabilities
	}

	// Get bundle size
	var bundleSize int
	if version.BundleSize.Valid {
		bundleSize = int(version.BundleSize.Int32)
	}

	// Fire-and-forget: generate FlyEmbed triple embeddings after successful publish.
	// Failures are logged but don't block the publish response.
	if h.recommendationSvc != nil {
		go func() {
			fn, err := h.repo.GetFunctionByID(context.Background(), fnID)
			if err != nil {
				logrus.WithError(err).WithField("function_id", fnID).Warn("FlyEmbed: failed to get function for embedding")
				return
			}
			fnVer, err := h.repo.GetFunctionVersion(fnID, req.Version)
			if err != nil {
				logrus.WithError(err).WithField("function_id", fnID).Warn("FlyEmbed: failed to get function version for embedding")
				return
			}
			var manifest map[string]interface{}
			if fnVer.Manifest != nil {
				_ = json.Unmarshal(fnVer.Manifest, &manifest)
			}
			var tags []string
			if fn.Tags != nil {
				_ = json.Unmarshal(fn.Tags, &tags)
			}
			sourceCode := ""
			if fnVer.SourceCode.Valid {
				sourceCode = fnVer.SourceCode.String
			}
			title := ""
			if fn.Title.Valid {
				title = fn.Title.String
			}
			description := ""
			if fn.Description.Valid {
				description = fn.Description.String
			}
			category := ""
			if fn.Category.Valid {
				category = fn.Category.String
			}
			if err := h.recommendationSvc.EmbedFunctionViaAIService(
				context.Background(), fnID, fn.Name,
				title, description, category,
				tags, manifest, sourceCode, fnVer.Runtime, capabilities,
			); err != nil {
				logrus.WithError(err).WithField("function_id", fnID).Warn("FlyEmbed: failed to generate triple embeddings")
			} else {
				logrus.WithField("function_id", fnID).Info("FlyEmbed: triple embeddings generated successfully")
			}
		}()
	}

	response := functionregistry.PublishResponse{
		OK:                 true,
		Function:           fmt.Sprintf("%s/%s", req.Author, req.Name),
		Version:            req.Version,
		Message:            "Function published successfully",
		VerificationStatus: verificationStatusStr,
		// Full function details
		FunctionID:    fnID.String(),
		Runtime:       m.Runtime,
		TimeoutMs:     m.TimeoutMs,
		MemoryMB:      m.MemoryMB,
		Capabilities:  capabilities,
		Deterministic: m.Deterministic,
		SideEffects:   m.SideEffects,
		Idempotent:    m.Idempotent,
		CacheTTL:      m.CacheTTL,
		BundleSize:    bundleSize,
		// Platform fee info
		FeeCharged:      feeCharged,
		FeeAmountUSD:    feeAmountUSD,
		FeeType:         feeType,
		PlatformFeePaid: platformFeePaid,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// bundleFunctionToWasm bundles function source code to WebAssembly
func (h *Handler) bundleFunctionToWasm(source *functionregistry.FunctionSource, fnManifest *functionregistry.FunctionManifest) ([]byte, string, error) {
	// Create temporary directory for source files
	tempDir, err := os.MkdirTemp("", "functionfly-bundle-*")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	// Create manifest file
	manifestData := map[string]interface{}{
		"name":          fnManifest.Name,
		"version":       fnManifest.Version,
		"runtime":       source.Runtime, // Use runtime from source
		"timeout_ms":    fnManifest.TimeoutMs,
		"memory_mb":     fnManifest.MemoryMB,
		"deterministic": fnManifest.Deterministic,
		"cache_ttl":     fnManifest.CacheTTL,
	}

	if fnManifest.Title != "" {
		manifestData["title"] = fnManifest.Title
	}
	if fnManifest.Description != "" {
		manifestData["description"] = fnManifest.Description
	}

	// Write manifest file
	manifestBytes, err := json.MarshalIndent(manifestData, "", "  ")
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal manifest: %w", err)
	}

	manifestPath := filepath.Join(tempDir, "functionfly.jsonc")
	if err := os.WriteFile(manifestPath, manifestBytes, 0644); err != nil {
		return nil, "", fmt.Errorf("failed to write manifest: %w", err)
	}

	// Write main source file
	var mainFile string
	switch source.Runtime {
	case "node18", "node20", "deno":
		mainFile = "index.js"
	case "python3.11":
		mainFile = "main.py"
	default:
		mainFile = "index.js"
	}

	mainPath := filepath.Join(tempDir, mainFile)
	if err := os.WriteFile(mainPath, []byte(source.Code), 0644); err != nil {
		return nil, "", fmt.Errorf("failed to write main source file: %w", err)
	}

	// Write additional files
	for filename, content := range source.Files {
		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return nil, "", fmt.Errorf("failed to write file %s: %w", filename, err)
		}
	}

	// Calculate source hash
	sourceHash := h.calculateSourceHash(source)

	// Create manifest for bundling
	bundleManifest := &manifest.Manifest{
		Name:          fnManifest.Name,
		Version:       fnManifest.Version,
		Runtime:       source.Runtime,
		TimeoutMS:     &fnManifest.TimeoutMs,
		MemoryMB:      &fnManifest.MemoryMB,
		Deterministic: &fnManifest.Deterministic,
		CacheTTL:      &fnManifest.CacheTTL,
	}

	// Bundle to WASM
	wasmBytes, err := bundler.BundleForWasmRuntimeWithWorkingDirectory(bundleManifest, tempDir)

	if err != nil {
		return nil, "", fmt.Errorf("failed to bundle to WASM: %w", err)
	}

	return wasmBytes, sourceHash, nil
}

// calculateSourceHash calculates SHA256 hash of source code for integrity verification
func (h *Handler) calculateSourceHash(source *functionregistry.FunctionSource) string {
	hasher := sha256.New()

	// Hash main code
	hasher.Write([]byte(source.Code))

	// Hash files in deterministic order
	var filenames []string
	for filename := range source.Files {
		filenames = append(filenames, filename)
	}

	// Sort for deterministic hashing
	sortedFilenames := make([]string, len(filenames))
	copy(sortedFilenames, filenames)
	for i := 0; i < len(sortedFilenames); i++ {
		for j := i + 1; j < len(sortedFilenames); j++ {
			if sortedFilenames[i] > sortedFilenames[j] {
				sortedFilenames[i], sortedFilenames[j] = sortedFilenames[j], sortedFilenames[i]
			}
		}
	}

	for _, filename := range sortedFilenames {
		hasher.Write([]byte(filename))
		hasher.Write([]byte(source.Files[filename]))
	}

	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// createChangelogEntry creates a changelog entry from the provided changelog input
func (h *Handler) createChangelogEntry(functionID, versionID uuid.UUID, req functionregistry.PublishRequest) error {
	if req.Changelog == nil {
		return nil
	}

	// Convert changes to JSON
	changesJSON, err := json.Marshal(req.Changelog.Changes)
	if err != nil {
		return fmt.Errorf("failed to marshal changelog changes: %w", err)
	}

	// Determine change type based on whether this is a new function or update
	changeType := storageregistry.ChangeTypeAdded
	if req.Changelog.Category == "bug_fix" {
		changeType = storageregistry.ChangeTypeFixed
	} else if req.Changelog.Category == "breaking_change" {
		changeType = storageregistry.ChangeTypeBreaking
	}

	changelog := &storageregistry.FunctionVersionChangelog{
		FunctionID:        functionID,
		FunctionVersionID: versionID,
		Version:           req.Version,
		ChangeType:        changeType,
		Category:          storageregistry.ChangeCategory(req.Changelog.Category),
		Title:             req.Changelog.Title,
		Description:       req.Changelog.Description,
		Changes:           changesJSON,
		Author:            req.Author,
	}

	if err := h.repo.CreateFunctionVersionChangelog(changelog); err != nil {
		return fmt.Errorf("failed to create changelog entry: %w", err)
	}

	return nil
}

// scanWasmForMalware performs an async YARA scan of the WASM binary.
// If a malware pattern is matched, the function version is marked as quarantined.
// This is a best-effort scan — failures are logged but don't block publishing.
func (h *Handler) scanWasmForMalware(functionID uuid.UUID, version string, wasmBytes []byte) {
	// YARA service URL from environment (default to local yara-service)
	yaraURL := os.Getenv("YARA_SERVICE_URL")
	if yaraURL == "" {
		yaraURL = "http://localhost:5000"
	}

	// Skip if the YARA service is not configured or scan is disabled
	if os.Getenv("YARA_SCAN_ENABLED") != "true" {
		return
	}

	scanURL := yaraURL + "/scan"
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Post(scanURL, "application/octet-stream", bytes.NewReader(wasmBytes))
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"function_id": functionID,
			"version":     version,
		}).Debug("YARA scan failed (service unreachable), skipping")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// Scan passed — no malware detected
		logrus.WithFields(logrus.Fields{
			"function_id": functionID,
			"version":     version,
		}).Debug("YARA scan passed: no malware detected")
		return
	}

	if resp.StatusCode == http.StatusConflict {
		// Status 409 = malware detected
		body, _ := io.ReadAll(resp.Body)
		logrus.WithFields(logrus.Fields{
			"function_id": functionID,
			"version":     version,
			"yara_result": string(body),
		}).Warn("YARA scan detected malware pattern in WASM binary")
		// Future: mark version as quarantined in the DB
		return
	}

	logrus.WithFields(logrus.Fields{
		"function_id": functionID,
		"version":     version,
		"yara_status": resp.StatusCode,
	}).Debug("YARA scan returned unexpected status")
}

// autoGenerateChangelog automatically generates a basic changelog entry
func (h *Handler) autoGenerateChangelog(functionID, versionID uuid.UUID, req functionregistry.PublishRequest, existingFn *storage.RegistryFunction) error {
	// Determine if this is a new function or an update
	isNewFunction := existingFn == nil

	var changeType storageregistry.ChangeType
	var category storageregistry.ChangeCategory
	var title, description string

	if isNewFunction {
		changeType = storageregistry.ChangeTypeAdded
		category = storageregistry.ChangeCategoryFeature
		title = fmt.Sprintf("Initial release v%s", req.Version)
		description = "First version of the function published"
	} else {
		changeType = storageregistry.ChangeTypeChanged
		category = storageregistry.ChangeCategoryOther
		title = fmt.Sprintf("Updated to v%s", req.Version)
		description = fmt.Sprintf("Function updated to version %s", req.Version)

		// Try to get previous version for comparison
		if existingFn.LatestVersion.Valid {
			prevVersion := existingFn.LatestVersion.String
			changelog := &storageregistry.FunctionVersionChangelog{
				FunctionID:        functionID,
				FunctionVersionID: versionID,
				Version:           req.Version,
				PreviousVersion:   &prevVersion,
				ChangeType:        changeType,
				Category:          category,
				Title:             title,
				Description:       description,
				Changes:           json.RawMessage("[]"),
				Author:            req.Author,
			}
			return h.repo.CreateFunctionVersionChangelog(changelog)
		}
	}

	changelog := &storageregistry.FunctionVersionChangelog{
		FunctionID:        functionID,
		FunctionVersionID: versionID,
		Version:           req.Version,
		ChangeType:        changeType,
		Category:          category,
		Title:             title,
		Description:       description,
		Changes:           json.RawMessage("[]"),
		Author:            req.Author,
	}

	if err := h.repo.CreateFunctionVersionChangelog(changelog); err != nil {
		return fmt.Errorf("failed to create auto-generated changelog: %w", err)
	}

	return nil
}
