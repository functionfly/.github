package billing

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// Handler handles billing portal and subscription management (Stripe).
type Handler struct {
	repo storage.Repository
}

// NewHandler creates a new billing handler.
func NewHandler(repo storage.Repository) *Handler {
	return &Handler{repo: repo}
}

// CreatePortalSessionRequest is the request body for creating a billing portal session.
type CreatePortalSessionRequest struct {
	ReturnURL string `json:"return_url"`
}

// CreatePortalSessionResponse is the response with the Stripe portal URL.
type CreatePortalSessionResponse struct {
	URL string `json:"url"`
}

// HandleCreatePortalSession creates a Stripe Customer Billing Portal session and returns the URL.
// POST /v1/billing/portal-session
func (h *Handler) HandleCreatePortalSession(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if !payment.IsConfigured() {
		writeJSONError(w, http.StatusServiceUnavailable, "Billing is not configured")
		return
	}

	user, err := h.repo.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		logrus.WithError(err).WithField("user_id", claims.UserID).Warn("billing portal: user not found")
		writeJSONError(w, http.StatusNotFound, "User not found")
		return
	}

	var req CreatePortalSessionRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	returnURL := req.ReturnURL
	if returnURL == "" {
		returnURL = "/settings"
	}
	// Stripe requires a full URL; build from request if path-only
	if strings.HasPrefix(returnURL, "/") {
		scheme := "https"
		if r.TLS == nil && (r.URL == nil || r.URL.Scheme == "") {
			scheme = "http"
		}
		if r.URL != nil && r.URL.Scheme != "" {
			scheme = r.URL.Scheme
		}
		returnURL = scheme + "://" + r.Host + returnURL
	}

	name := user.Name
	if name == "" && user.ProviderData != nil {
		if n, ok := user.ProviderData["name"].(string); ok {
			name = n
		}
	}
	if name == "" {
		name = user.Email
	}

	customerID, err := payment.CreateOrGetStripeCustomer(r.Context(), h.repo, claims.TenantID, user.Email, name)
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", claims.TenantID).Error("billing portal: create or get stripe customer")
		msg := "Failed to prepare billing session"
		if os.Getenv("DEVELOPMENT") == "true" {
			msg = err.Error()
		}
		writeJSONError(w, http.StatusInternalServerError, msg)
		return
	}

	url, err := payment.CreateBillingPortalSession(r.Context(), customerID, returnURL)
	if err != nil {
		logrus.WithError(err).WithField("customer_id", customerID).Error("billing portal: create session")
		msg := "Failed to create billing session"
		if os.Getenv("DEVELOPMENT") == "true" {
			msg = err.Error()
		}
		writeJSONError(w, http.StatusInternalServerError, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CreatePortalSessionResponse{URL: url})
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
