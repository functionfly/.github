// Package registry provides the main registry handlers
package registry

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/handlers/registry/execution"
)

// buildExecHandler assembles a execution.Handler wired with runtime router and bundle service.
func (h *Handler) buildExecHandler() *execution.Handler {
	return &execution.Handler{
		Repo:              h.repo,
		BackendRepo:       h.backendRepo,
		CacheService:      h.cacheService,
		EdgeCache:         h.edgeCache,
		UsageTracker:      h.realtimeUsageTracker,
		PrivacyService:    h.privacySvc,
		MicroVMRepo:       h.MicroVMRepo,
		DNARecorder:       h.dnaRecorder,
		NodeID:            h.dreNodeID,
		Region:            h.dreRegion,
		NodeKey:           h.dreNodeKey,
		PlatformKey:       h.drePlatformKey,
		BillingController: h.billingCtrl,
		RuntimeRouter:     h.runtimeRouter,
		BundleService:     h.bundleService,
	}
}

// HandleExecute handles executing a function
func (h *Handler) HandleExecute(w http.ResponseWriter, r *http.Request) {
	h.buildExecHandler().HandleExecute(w, r)
}

// HandleTest handles testing a function with validation data
func (h *Handler) HandleTest(w http.ResponseWriter, r *http.Request) {
	h.buildExecHandler().HandleTest(w, r)
}

// HandleGetReplay handles retrieving a shareable execution replay
func (h *Handler) HandleGetReplay(w http.ResponseWriter, r *http.Request) {
	h.buildExecHandler().HandleGetReplay(w, r)
}

// HandleVerifyReplay handles manual verification of a specific execution replay
func (h *Handler) HandleVerifyReplay(w http.ResponseWriter, r *http.Request) {
	h.buildExecHandler().HandleVerifyReplay(w, r)
}
