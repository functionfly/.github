package registry

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/functionfly/functionfly/internal/api/handlers/registry/execution"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/artifacts"
	"github.com/functionfly/functionfly/internal/atlas"
	"github.com/functionfly/functionfly/internal/bundler"
	"github.com/functionfly/functionfly/internal/cache"
	"github.com/functionfly/functionfly/internal/dre"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/functionfly/functionfly/internal/wallet"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler contains registry API handlers
type Handler struct {
	repo              *registry.RegistryRepository
	backendRepo       storage.Repository
	functionRepo      *storage.FunctionRepository
	cacheService      *cache.CacheService
	cdnService        *cache.CDNService
	edgeCache         *cache.EdgeCacheService
	realtimeMonitor   *monitoring.RealtimeMonitor
	platformFeeRepo   *registry.PlatformFeeRepository
	walletSvc         *wallet.Service
	recommendationSvc interface {
		EmbedFunctionViaAIService(ctx context.Context, functionID uuid.UUID, name, title, description, category string, tags []string, manifest map[string]interface{}, sourceCode, runtime string, capabilities []string) error
	}
	// RealtimeUsageTracker provides real-time quota enforcement and usage tracking
	realtimeUsageTracker services.RealtimeUsageTrackerInterface
	// PrivacyService provides privacy and compliance features
	privacySvc execution.PrivacyService
	// MicroVMRepo provides MicroVM execution tracking and billing (optional)
	MicroVMRepo *storage.MicroVMRepository
	// DRE execution node config (optional): when set, FXCERTs are signed
	dreNodeKey     ed25519.PrivateKey
	drePlatformKey ed25519.PrivateKey
	dreNodeID      string
	dreRegion      string
	// DNA recorder for execution pipeline (optional)
	dnaRecorder execution.DNARecorder
	// BillingController handles wallet operations for paid function execution (optional)
	billingCtrl execution.BillingController
	// RuntimeRouter routes execution to the appropriate engine (optional; wired at startup)
	runtimeRouter *execution.RuntimeRouter
	// BundleService provides eager bundling at publish time (optional)
	bundleService *bundler.BundleService
	// AutoReadmeService generates README docs for functions
	autoReadmeSvc interface {
		GenerateForVersion(ctx context.Context, functionID uuid.UUID, version string, force bool) (string, error)
		BackfillAll(ctx context.Context, batchSize int, force bool) (int, error)
	}
	// ReceiptMilestoneHook is called after a successful public execution receipt is created.
	// Set by the receipt milestone wiring in routes.go.
	ReceiptMilestoneHook func(ctx context.Context, functionID uuid.UUID, tenantID *uuid.UUID, publicID string)
	// repHooker awards reputation points for platform actions (optional)
	repHooker *services.ReputationHooker
	// atlasTracer records execution traces to Atlas Memory Engine (optional)
	atlasTracer *atlas.Tracer
	// artifactStore holds user-uploaded source/WASM/readme bytes in object
	// storage (R2 by default). When nil the publish path falls back to the
	// legacy Postgres bytea/text columns.
	artifactStore artifacts.Store
	// artifactResolver fetches artifact bytes during execution, transparently
	// falling back to DB columns for legacy rows during the cutover window.
	artifactResolver *artifacts.Resolver
}

// SetWalletService sets the wallet service for unified wallet operations
func (h *Handler) SetWalletService(walletSvc *wallet.Service) {
	h.walletSvc = walletSvc
}

// SetReputationHooker sets the reputation hooker for awarding points
func (h *Handler) SetReputationHooker(reputer *services.ReputationHooker) {
	h.repHooker = reputer
}

// NewHandler creates a new registry handler.
// DRE node key is loaded from env (DRE_NODE_PRIVATE_KEY or DRE_NODE_PRIVATE_KEY_PATH); when set, FXCERTs are signed.
func NewHandler(
	repo *registry.RegistryRepository,
	backendRepo storage.Repository,
	functionRepo *storage.FunctionRepository,
	cacheService *cache.CacheService,
	cdnService *cache.CDNService,
	edgeCache *cache.EdgeCacheService,
	realtimeMonitor *monitoring.RealtimeMonitor,
	platformFeeRepo *registry.PlatformFeeRepository,
	recommendationSvc interface {
		EmbedFunctionViaAIService(ctx context.Context, functionID uuid.UUID, name, title, description, category string, tags []string, manifest map[string]interface{}, sourceCode, runtime string, capabilities []string) error
	},
	realtimeUsageTracker services.RealtimeUsageTrackerInterface,
) *Handler {
	h := &Handler{
		repo:                 repo,
		backendRepo:          backendRepo,
		functionRepo:         functionRepo,
		cacheService:         cacheService,
		cdnService:           cdnService,
		edgeCache:            edgeCache,
		realtimeMonitor:      realtimeMonitor,
		platformFeeRepo:      platformFeeRepo,
		recommendationSvc:    recommendationSvc,
		realtimeUsageTracker: realtimeUsageTracker,
		privacySvc:           nil, // Set later via SetPrivacyService
	}
	key, nodeID, region, err := dre.LoadNodeKeyFromEnv()
	if err != nil {
		logrus.WithError(err).Warn("DRE: failed to load node key from env; FXCERTs will be unsigned")
	}
	if key != nil {
		h.dreNodeKey = key
		logrus.Info("DRE: node key loaded; FXCERTs will be signed")
	}
	platformKey, err := dre.LoadPlatformKeyFromEnv()
	if err != nil {
		logrus.WithError(err).Warn("DRE: failed to load platform key from env")
	}
	if platformKey != nil {
		h.drePlatformKey = platformKey
		logrus.Info("DRE: platform key loaded; FXCERTs will include platform signature")
	}
	h.dreNodeID = nodeID
	h.dreRegion = region
	return h
}

// SetPrivacyService sets the privacy service for compliance features
func (h *Handler) SetPrivacyService(svc execution.PrivacyService) {
	h.privacySvc = svc
}

// SetDNARecorder sets the DNA recorder for execution pipeline integration
func (h *Handler) SetDNARecorder(recorder execution.DNARecorder) {
	h.dnaRecorder = recorder
}

// SetBillingController sets the billing controller for paid function execution
func (h *Handler) SetBillingController(ctrl execution.BillingController) {
	h.billingCtrl = ctrl
}

// SetRuntimeRouter wires the runtime router into the execution pipeline.
func (h *Handler) SetRuntimeRouter(router *execution.RuntimeRouter) {
	h.runtimeRouter = router
}

// SetBundleService wires the eager-bundling service for publish-time compilation.
func (h *Handler) SetBundleService(svc *bundler.BundleService) {
	h.bundleService = svc
}

// SetMicroVMRepo wires the MicroVM repository for execution tracking and billing.
func (h *Handler) SetMicroVMRepo(repo *storage.MicroVMRepository) {
	h.MicroVMRepo = repo
}

// SetAtlasTracer wires the Atlas Memory Engine tracer for execution tracing.
func (h *Handler) SetAtlasTracer(tracer *atlas.Tracer) {
	h.atlasTracer = tracer
}

// SetArtifactStore wires the object-storage backend used for user-uploaded
// function artifacts. When store is nil the publish path keeps the legacy
// Postgres bytea/text storage (no behavior change).
func (h *Handler) SetArtifactStore(store artifacts.Store, lru *artifacts.DiskLRU) {
	h.artifactStore = store
	if store != nil {
		h.artifactResolver = artifacts.NewResolver(store, lru)
	}
}

// ArtifactResolver returns the read-side resolver wired in by
// SetArtifactStore, or nil if no artifact store is configured.
func (h *Handler) ArtifactResolver() *artifacts.Resolver {
	return h.artifactResolver
}

// SetAutoReadmeService sets the auto-README generator service.
func (h *Handler) SetAutoReadmeService(svc interface {
	GenerateForVersion(ctx context.Context, functionID uuid.UUID, version string, force bool) (string, error)
	BackfillAll(ctx context.Context, batchSize int, force bool) (int, error)
}) {
	h.autoReadmeSvc = svc
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in case of multiple
		if idx := strings.Index(xff, ","); idx > 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx > 0 {
		return r.RemoteAddr[:idx]
	}

	return r.RemoteAddr
}

// HandleGetSDKCode handles generating SDK code for a function
func (h *Handler) HandleGetSDKCode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]
	sdk := r.URL.Query().Get("sdk") // javascript, python, go

	fn, err := h.repo.GetFunctionByAuthorName(context.Background(), author, name)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	fnVersion, err := h.repo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("No versions available"))
		return
	}

	var code string
	switch sdk {
	case "python":
		code = generatePythonSDK(fn.Author, fn.Name, fnVersion.Version, fn.Title.String, fn.Description.String)
	case "go":
		code = generateGoSDK(fn.Author, fn.Name, fnVersion.Version, fn.Title.String, fn.Description.String)
	default:
		code = generateJavaScriptSDK(fn.Author, fn.Name, fnVersion.Version, fn.Title.String, fn.Description.String)
	}

	response := map[string]interface{}{
		"sdk":    sdk,
		"code":   code,
		"author": author,
		"name":   name,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleServeSDK serves static SDK files via CDN
func (h *Handler) HandleServeSDK(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sdkType := vars["sdk"]       // javascript, python, go
	version := vars["version"]   // latest, v1.0.0, etc.
	filename := vars["filename"] // functionfly.js, functionfly.py, etc.

	if h.cdnService == nil || !h.cdnService.IsCDNEnabled() {
		cache.RecordCDNMiss()
		h.serveSDKLocally(w, r, sdkType, version, filename)
		return
	}

	// Construct CDN path and redirect to CDN when enabled
	cdnPath := fmt.Sprintf("/sdk/%s/%s/%s", sdkType, version, filename)
	cdnURL := h.cdnService.GetCDNURL(cdnPath)
	if cdnURL != cdnPath {
		// CDN is configured: redirect so the client fetches from CDN
		cache.RecordCDNHit()
		http.Redirect(w, r, cdnURL, http.StatusTemporaryRedirect)
		return
	}
	// CDN URL same as path (disabled): serve locally with CDN headers for consistency
	h.cdnService.SetCDNHeaders(w, cdnPath)
	cache.RecordCDNHit()
	h.serveSDKLocally(w, r, sdkType, version, filename)
}

// HandleServeDocs serves documentation files via CDN
func (h *Handler) HandleServeDocs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	docType := vars["type"]    // api, sdk, guides
	version := vars["version"] // latest, v1.0.0, etc.
	path := vars["path"]       // getting-started.md, api-reference.html, etc.

	if h.cdnService == nil || !h.cdnService.IsCDNEnabled() {
		cache.RecordCDNMiss()
		h.serveDocsLocally(w, r, docType, version, path)
		return
	}

	// Construct CDN path and redirect to CDN when enabled
	cdnPath := fmt.Sprintf("/docs/%s/%s/%s", docType, version, path)
	cdnURL := h.cdnService.GetCDNURL(cdnPath)
	if cdnURL != cdnPath {
		cache.RecordCDNHit()
		http.Redirect(w, r, cdnURL, http.StatusTemporaryRedirect)
		return
	}
	h.cdnService.SetCDNHeaders(w, cdnPath)
	cache.RecordCDNHit()
	h.serveDocsLocally(w, r, docType, version, path)
}

// HandleServeStatic serves other static assets via CDN
func (h *Handler) HandleServeStatic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["category"] // images, css, js, fonts, etc.
	path := vars["path"]         // logo.png, styles.css, etc.

	if h.cdnService == nil || !h.cdnService.IsCDNEnabled() {
		cache.RecordCDNMiss()
		h.serveStaticLocally(w, r, category, path)
		return
	}

	// Construct CDN path and redirect to CDN when enabled
	cdnPath := fmt.Sprintf("/static/%s/%s", category, path)
	cdnURL := h.cdnService.GetCDNURL(cdnPath)
	if cdnURL != cdnPath {
		cache.RecordCDNHit()
		http.Redirect(w, r, cdnURL, http.StatusTemporaryRedirect)
		return
	}
	h.cdnService.SetCDNHeaders(w, cdnPath)

	// Record CDN hit
	cache.RecordCDNHit()

	// For now, serve locally with CDN headers
	h.serveStaticLocally(w, r, category, path)
}

// HandleGetCacheStats returns comprehensive cache statistics
// This endpoint provides public cache metrics for monitoring and debugging
func (h *Handler) HandleGetCacheStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"cache_enabled":      h.cacheService != nil,
		"redis_enabled":      h.cacheService != nil && h.cacheService.IsRedisCacheEnabled(),
		"cdn_enabled":        h.cdnService != nil && h.cdnService.IsCDNEnabled(),
		"edge_cache_enabled": h.edgeCache != nil,
	}

	// Get comprehensive cache service stats if available
	if h.cacheService != nil {
		// Memory cache stats (L1)
		if memStats := h.cacheService.GetMemoryStats(); memStats != nil {
			stats["memory_cache"] = map[string]interface{}{
				"layer":      "L1",
				"type":       "memory",
				"hits":       memStats.Hits,
				"misses":     memStats.Misses,
				"hit_ratio":  memStats.Ratio,
				"size_bytes": memStats.SizeBytes,
				"evictions":  memStats.Evictions,
			}
		}

		// Disk cache stats (L2)
		if diskStats, err := h.cacheService.GetDiskStats(); err == nil && diskStats != nil {
			var hitRatio float64
			if diskStats.TotalHits > 0 {
				totalLookups := diskStats.TotalEntries + diskStats.TotalHits // Approximation
				if totalLookups > 0 {
					hitRatio = float64(diskStats.TotalHits) / float64(totalLookups)
				}
			}
			stats["disk_cache"] = map[string]interface{}{
				"layer":            "L2",
				"type":             "disk",
				"total_entries":    diskStats.TotalEntries,
				"total_size_bytes": diskStats.TotalSizeBytes,
				"total_hits":       diskStats.TotalHits,
				"hit_ratio":        hitRatio,
				"expired_entries":  diskStats.ExpiredEntries,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// certJSONMinimal is used to parse CertJSON for bootstrap detection.
type certJSONMinimal struct {
	Execution struct {
		FunctionID string `json:"function_id"`
		NodeID     string `json:"node_id"`
	} `json:"execution"`
}

// HandleRegenerateBootstrap regenerates bootstrap FXCERTs for registry functions so they are signed with the current node key.
// POST /v1/admin/registry/dre/regenerate-bootstrap?author=functionfly (optional author filter)
func (h *Handler) HandleRegenerateBootstrap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}
	author := strings.TrimSpace(r.URL.Query().Get("author"))

	functions, _, err := h.repo.ListFunctionsForAdmin("", "", "", 5000, 0)
	if err != nil {
		logrus.WithError(err).Error("DRE: regenerate bootstrap failed to list functions")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to list functions"})
		return
	}

	regenerated := 0
	for _, fn := range functions {
		if author != "" && !strings.EqualFold(fn.Author, author) {
			continue
		}
		versions, err := h.repo.ListFunctionVersions(fn.ID)
		if err != nil {
			logrus.WithError(err).WithField("function_id", fn.ID).Warn("DRE: list versions failed")
			continue
		}
		for _, v := range versions {
			fnVersion, err := h.repo.GetFunctionVersion(fn.ID, v.Version)
			if err != nil || fnVersion == nil {
				continue
			}
			expectedFunctionID := fmt.Sprintf("fx://%s/%s/%s", fn.Author, fn.Name, fnVersion.Version)
			certs, err := h.repo.GetCertificatesByFunctionID(fn.ID, 200, 0)
			if err != nil {
				continue
			}
			for _, cert := range certs {
				var minimal certJSONMinimal
				if err := json.Unmarshal(cert.CertJSON, &minimal); err != nil {
					continue
				}
				if minimal.Execution.NodeID != "bootstrap" || minimal.Execution.FunctionID != expectedFunctionID {
					continue
				}
				if err := h.repo.DeleteCertificate(cert.CertificateID); err != nil {
					logrus.WithError(err).WithField("certificate_id", cert.CertificateID).Warn("DRE: delete cert failed")
					continue
				}
				if err := h.repo.DeleteMEGRecord(cert.MEGRecordID); err != nil {
					logrus.WithError(err).WithField("meg_id", cert.MEGRecordID).Warn("DRE: delete MEG failed")
				}
				execution.BootstrapFXCERT(h.repo, &fn, fnVersion, "bootstrap", "internal", h.dreNodeKey, h.drePlatformKey)
				regenerated++
				logrus.WithFields(logrus.Fields{
					"function": fmt.Sprintf("%s/%s", fn.Author, fn.Name),
					"version":  fnVersion.Version,
					"cert_id":  cert.CertificateID,
				}).Info("DRE: regenerated bootstrap FXCERT")
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"regenerated": regenerated,
		"message":     fmt.Sprintf("regenerated %d bootstrap FXCERT(s)", regenerated),
	})
}

// HandleBackfillReadmes auto-generates READMEs for functions that are missing them.
// POST /v1/admin/registry/readmes/backfill?batch_size=100&force=false
func (h *Handler) HandleBackfillReadmes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	if h.autoReadmeSvc == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "auto-readme service not available"})
		return
	}

	batchSize := 100
	if bs := r.URL.Query().Get("batch_size"); bs != "" {
		if v, err := strconv.Atoi(bs); err == nil && v > 0 && v <= 1000 {
			batchSize = v
		}
	}
	force := r.URL.Query().Get("force") == "true"

	ctx := r.Context()
	total, err := h.autoReadmeSvc.BackfillAll(ctx, batchSize, force)
	if err != nil {
		logrus.WithError(err).Error("Auto-readme backfill failed")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":        true,
		"total_versions": total,
		"batch_size":     batchSize,
		"force":          force,
		"message":        fmt.Sprintf("Backfilled %d version(s)", total),
	})
}
