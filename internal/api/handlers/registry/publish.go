package registry

import (
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req functionregistry.PublishRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Author == "" || req.Name == "" || req.Version == "" {
		http.Error(w, "author, name, and version are required", http.StatusBadRequest)
		return
	}

	// Validate semver format
	if err := functionregistry.ValidateSemVer(req.Version); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
			http.Error(w, "invalid conflict_strategy: must be 'error', 'overwrite', or 'create_new'", http.StatusBadRequest)
			return
		}
	}

	// For registry functions, source code is required for sandbox execution
	if req.Source == nil || req.Source.Code == "" {
		http.Error(w, "source code is required for registry functions", http.StatusBadRequest)
		return
	}

	// Parse manifest early so we can set/update function metadata (title, description, category, tags)
	cleanManifest := manifest.StripComments(string(req.Manifest))
	var m functionregistry.FunctionManifest
	if err := json.Unmarshal([]byte(cleanManifest), &m); err != nil {
		http.Error(w, "Invalid manifest JSON", http.StatusBadRequest)
		return
	}

	// Determine trust level from request or manifest
	trustLevel := req.TrustLevel
	if trustLevel == "" {
		trustLevel = "standard" // Default trust level
	}

	// Check if function already exists
	existingFn, err := h.repo.GetFunctionByAuthorName(req.Author, req.Name)
	if err != nil && !isRecordNotFound(err) {
		logrus.WithError(err).Error("Failed to check existing function")
		http.Error(w, "Failed to check function: "+err.Error(), http.StatusInternalServerError)
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
			PricePerCall:       0,
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

		if err := h.repo.CreateFunction(fn); err != nil {
			logrus.WithError(err).Error("Failed to create function")
			http.Error(w, "Failed to create function: "+err.Error(), http.StatusInternalServerError)
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
		// FunctionFly functions get default high trust on function record if not already set
		if strings.EqualFold(req.Author, "functionfly") && existingFn.ReliabilityScore == 0 && existingFn.DeterministicScore == 0 {
			meta["reliability_score"] = 90.0
			meta["deterministic_score"] = 90.0
		}
		if len(meta) > 0 {
			if _, err := h.repo.UpdateRegistryFunction(fnID, meta); err != nil {
				logrus.WithError(err).Warn("Failed to update function metadata from manifest")
			}
		}
	}

	// Validate capabilities against allowed list
	for _, cap := range m.Capabilities {
		if !functionregistry.IsValidCapability(cap) {
			http.Error(w, "Invalid capability: "+cap+". Allowed: "+strings.Join(functionregistry.AllowedCapabilities, ", "), http.StatusBadRequest)
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
			http.Error(w, "Invalid WASM binary encoding: "+decodeErr.Error(), http.StatusBadRequest)
			return
		}
		// Calculate source hash from the code
		sourceHash = h.calculateSourceHash(req.Source)
		logrus.WithField("size", len(wasmBinary)).Info("Using pre-compiled WASM binary")
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
			http.Error(w, upsertErr.Error(), http.StatusConflict)
			return
		}
		http.Error(w, "Failed to create version: "+upsertErr.Error(), http.StatusInternalServerError)
		return
	}

	// Update latest version pointer
	if err := h.repo.UpdateFunctionLatestVersion(fnID, req.Version); err != nil {
		logrus.WithError(err).Error("Failed to update latest version")
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

	// Auto-verify official FunctionFly functions so they are always verified and get FXCERTs
	if strings.EqualFold(req.Author, "functionfly") {
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
			logrus.WithError(err).WithField("function_version_id", version.ID).Warn("Failed to auto-verify FunctionFly function")
		} else {
			logrus.WithFields(logrus.Fields{
				"function": fmt.Sprintf("%s/%s", req.Author, req.Name),
				"version":  req.Version,
			}).Info("Auto-verified FunctionFly function with FXCERT status")
		}
		// Generate bootstrap FXCERT so History and Certificates show a cert after publish
		go func() {
			fn, err := h.repo.GetFunctionByID(fnID)
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

	// NEW: Skip synchronous verification during publish - verify lazily at execute time
	// This significantly speeds up publish by avoiding expensive security scans
	// Verification will happen on first execution instead
	logrus.WithFields(logrus.Fields{
		"function_id":   version.ID,
		"function_name": fmt.Sprintf("%s/%s", req.Author, req.Name),
		"version":       req.Version,
		"lazy_verify":   true,
	}).Info("Skipping synchronous verification - will verify at execute time")

	var verificationStatusStr string
	if trustLevel == "high" || trustLevel == "enterprise" {
		// For high-trust levels, we still need verification but can do it async
		verificationStatusStr = "pending"
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
