// Package registry provides the main registry handlers
package registry

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/handlers/registry/execution"
)

// HandleExecute handles executing a function
func (h *Handler) HandleExecute(w http.ResponseWriter, r *http.Request) {
	// Delegate to the execution subpackage
	execHandler := &execution.Handler{
		Repo:         h.repo,
		BackendRepo:  h.backendRepo,
		CacheService: h.cacheService,
		EdgeCache:    h.edgeCache,
	}
	execHandler.HandleExecute(w, r)
}

// HandleTest handles testing a function with validation data
func (h *Handler) HandleTest(w http.ResponseWriter, r *http.Request) {
	// Delegate to the execution subpackage
	execHandler := &execution.Handler{
		Repo:         h.repo,
		BackendRepo:  h.backendRepo,
		CacheService: h.cacheService,
		EdgeCache:    h.edgeCache,
	}
	execHandler.HandleTest(w, r)
}

// HandleGetReplay handles retrieving a shareable execution replay
func (h *Handler) HandleGetReplay(w http.ResponseWriter, r *http.Request) {
	// Delegate to the execution subpackage
	execHandler := &execution.Handler{
		Repo:         h.repo,
		BackendRepo:  h.backendRepo,
		CacheService: h.cacheService,
		EdgeCache:    h.edgeCache,
	}
	execHandler.HandleGetReplay(w, r)
}

// HandleVerifyReplay handles manual verification of a specific execution replay
func (h *Handler) HandleVerifyReplay(w http.ResponseWriter, r *http.Request) {
	// Delegate to the execution subpackage
	execHandler := &execution.Handler{
		Repo:         h.repo,
		BackendRepo:  h.backendRepo,
		CacheService: h.cacheService,
		EdgeCache:    h.edgeCache,
	}
	execHandler.HandleVerifyReplay(w, r)
}