// Package api — receipt route registration.
//
// Mounts the public Execution Receipt HTTP surface on the supplied root
// router. All paths are public; rate limits are applied per-IP via the
// receipt handler's own limiter. The owner-only revoke route is
// registered separately via registerReceiptAuthedRoutes so the caller
// chooses the auth middleware.
package api

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/handlers/receipt"
	"github.com/gorilla/mux"
)

// registerReceiptPublicRoutes wires the public receipt surface onto the
// root router. No auth required; per-IP rate limits are applied inside the
// handler.
func registerReceiptPublicRoutes(root *mux.Router, h *receipt.Handler) {
	if h == nil {
		return
	}
	h.RegisterRoutes(root)
}

// registerReceiptAuthedRoutes wires the owner-only revoke route behind the
// supplied auth middleware. The auth middleware is the same one used for
// other authed registry routes (e.g. function settings).
func registerReceiptAuthedRoutes(root *mux.Router, h *receipt.Handler, authMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	if h == nil {
		return
	}
	h.RegisterAuthedRoutes(root, authMiddleware)
}
