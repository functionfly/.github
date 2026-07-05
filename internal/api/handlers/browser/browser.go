package browser

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/agent/browser"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles browser-related HTTP requests.
type Handler struct {
	browserSvc browser.Browser
}

// NewHandler creates a new browser handler.
func NewHandler(browserSvc browser.Browser) *Handler {
	return &Handler{browserSvc: browserSvc}
}

// RegisterRoutes registers browser routes.
func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/agents/{agent_id}/browser/sessions", h.CreateSession).Methods("POST")
	router.HandleFunc("/agents/{agent_id}/browser/sessions", h.ListSessions).Methods("GET")
	router.HandleFunc("/agents/{agent_id}/browser/sessions/{session_id}", h.GetSession).Methods("GET")
	router.HandleFunc("/agents/{agent_id}/browser/sessions/{session_id}", h.CloseSession).Methods("DELETE")

	// Browser actions
	router.HandleFunc("/agents/{agent_id}/browser/navigate", h.Navigate).Methods("POST")
	router.HandleFunc("/agents/{agent_id}/browser/click", h.Click).Methods("POST")
	router.HandleFunc("/agents/{agent_id}/browser/fill", h.Fill).Methods("POST")
	router.HandleFunc("/agents/{agent_id}/browser/extract", h.Extract).Methods("POST")
	router.HandleFunc("/agents/{agent_id}/browser/screenshot", h.Screenshot).Methods("POST")

	// Credential management
	router.HandleFunc("/agents/{agent_id}/browser/credentials", h.StoreCredential).Methods("POST")
	router.HandleFunc("/agents/{agent_id}/browser/credentials", h.ListCredentials).Methods("GET")
	router.HandleFunc("/agents/{agent_id}/browser/credentials/{credential_id}", h.GetCredential).Methods("GET")
	router.HandleFunc("/agents/{agent_id}/browser/credentials/{credential_id}", h.DeleteCredential).Methods("DELETE")

	// Configuration
	router.HandleFunc("/agents/{agent_id}/browser/config", h.GetConfig).Methods("GET")
	router.HandleFunc("/agents/{agent_id}/browser/config", h.UpsertConfig).Methods("PUT")

	// Usage stats
	router.HandleFunc("/agents/{agent_id}/browser/usage", h.GetUsage).Methods("GET")
}

// CreateSessionRequest is the request body for creating a session.
type CreateSessionRequest struct {
	Isolated bool `json:"isolated"`
}

// CreateSession creates a new browser session.
func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]
	ctx := r.Context()

	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Isolated = false
	}

	session, err := h.browserSvc.CreateSession(ctx, agentID, req.Isolated)
	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(session); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}

// ListSessions lists all sessions for an agent.
func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]
	ctx := r.Context()

	sessions, err := h.browserSvc.ListSessions(ctx, agentID)
	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sessions); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}

// GetSession gets a session by ID.
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID, err := uuid.Parse(vars["session_id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("invalid session ID"))
		return
	}

	session, err := h.browserSvc.GetSession(r.Context(), sessionID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("session not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(session); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}

// CloseSession closes a browser session.
func (h *Handler) CloseSession(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]
	sessionID := vars["session_id"]

	err := h.browserSvc.CloseSession(r.Context(), agentID, sessionID)
	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "session closed"}); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}

// NavigateRequest is the request body for navigation.
type NavigateRequest struct {
	URL       string  `json:"url"`
	SessionID *string `json:"session_id,omitempty"`
}

// Navigate navigates to a URL.
func (h *Handler) Navigate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]
	ctx := r.Context()

	var req NavigateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteErrorWithStatus(w, http.StatusBadRequest, apierror.ErrCodeBadRequest, "invalid request body")
		return
	}

	var sessionID *uuid.UUID
	if req.SessionID != nil {
		parsed, err := uuid.Parse(*req.SessionID)
		if err == nil {
			sessionID = &parsed
		}
	}

	result, err := h.browserSvc.Navigate(ctx, agentID, req.URL, sessionID)
	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}

// ClickRequest is the request body for clicking.
type ClickRequest struct {
	SessionID  string `json:"session_id"`
	ElementRef string `json:"element_ref"`
}

// Click clicks an element.
func (h *Handler) Click(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]
	ctx := r.Context()

	var req ClickRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteErrorWithStatus(w, http.StatusBadRequest, apierror.ErrCodeBadRequest, "invalid request body")
		return
	}

	err := h.browserSvc.Click(ctx, agentID, req.SessionID, req.ElementRef)
	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "clicked"}); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}

// FillRequest is the request body for filling forms.
type FillRequest struct {
	SessionID  string `json:"session_id"`
	ElementRef string `json:"element_ref"`
	Value     string `json:"value"`
}

// Fill fills a form field.
func (h *Handler) Fill(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]
	ctx := r.Context()

	var req FillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteErrorWithStatus(w, http.StatusBadRequest, apierror.ErrCodeBadRequest, "invalid request body")
		return
	}

	err := h.browserSvc.Fill(ctx, agentID, req.SessionID, req.ElementRef, req.Value)
	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "filled"}); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}

// ExtractRequest is the request body for extraction.
type ExtractRequest struct {
	SessionID string `json:"session_id"`
	Selector  string `json:"selector"`
}

// Extract extracts structured content.
func (h *Handler) Extract(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]
	ctx := r.Context()

	var req ExtractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteErrorWithStatus(w, http.StatusBadRequest, apierror.ErrCodeBadRequest, "invalid request body")
		return
	}

	content, err := h.browserSvc.Extract(ctx, agentID, req.SessionID, req.Selector)
	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"content": content}); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}

// ScreenshotRequest is the request body for screenshots.
type ScreenshotRequest struct {
	SessionID string `json:"session_id"`
}

// Screenshot captures a screenshot.
func (h *Handler) Screenshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]
	ctx := r.Context()

	var req ScreenshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteErrorWithStatus(w, http.StatusBadRequest, apierror.ErrCodeBadRequest, "invalid request body")
		return
	}

	screenshot, err := h.browserSvc.Screenshot(ctx, agentID, req.SessionID)
	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"screenshot": screenshot}); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}

// StoreCredentialRequest is the request body for storing credentials.
type StoreCredentialRequest struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Data   struct {
		Cookies    []browser.SessionCookie `json:"cookies,omitempty"`
		AuthHeader string                 `json:"auth_header,omitempty"`
		Tokens     map[string]string      `json:"tokens,omitempty"`
	} `json:"data"`
}

// StoreCredential stores a browser credential.
func (h *Handler) StoreCredential(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]
	ctx := r.Context()

	var req StoreCredentialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteErrorWithStatus(w, http.StatusBadRequest, apierror.ErrCodeBadRequest, "invalid request body")
		return
	}

	data := &browser.CredentialData{
		Cookies:    req.Data.Cookies,
		AuthHeader: req.Data.AuthHeader,
		Tokens:     req.Data.Tokens,
	}

	credential, err := h.browserSvc.StoreCredential(ctx, agentID, req.Name, req.Domain, data)
	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(credential); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}

// ListCredentials lists all credentials for an agent.
func (h *Handler) ListCredentials(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]

	credentials, err := h.browserSvc.ListCredentials(r.Context(), agentID)
	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(credentials); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}

// GetCredential gets a credential by ID.
func (h *Handler) GetCredential(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]
	credentialID := vars["credential_id"]

	credential, err := h.browserSvc.GetCredential(r.Context(), agentID, credentialID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("credential not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(credential); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}

// DeleteCredential deletes a credential.
func (h *Handler) DeleteCredential(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]
	credentialID := vars["credential_id"]

	err := h.browserSvc.DeleteCredential(r.Context(), agentID, credentialID)
	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "credential deleted"}); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}

// GetConfig gets browser configuration for an agent.
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]

	perm, err := h.browserSvc.GetPermission(r.Context(), agentID)
	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(perm); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}

// UpsertConfigRequest is the request body for updating config.
type UpsertConfigRequest struct {
	BrowserEnabled           bool     `json:"browser_enabled"`
	AllowedDomains           []string `json:"allowed_domains"`
	MaxSessions             int      `json:"max_sessions"`
	CredentialStorageEnabled bool    `json:"credential_storage_enabled"`
	DefaultTimeoutMs        int      `json:"default_timeout_ms"`
	HeadfulMode             bool     `json:"headful_mode"`
}

// UpsertConfig updates browser configuration for an agent.
func (h *Handler) UpsertConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]
	ctx := r.Context()

	var req UpsertConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteErrorWithStatus(w, http.StatusBadRequest, apierror.ErrCodeBadRequest, "invalid request body")
		return
	}

	perm := &browser.BrowserPermission{
		AgentID:               agentID,
		BrowserEnabled:        req.BrowserEnabled,
		AllowedDomains:        req.AllowedDomains,
		MaxSessions:          req.MaxSessions,
		CredentialStorage:    req.CredentialStorageEnabled,
		DefaultTimeoutMs:      req.DefaultTimeoutMs,
		HeadfulMode:           req.HeadfulMode,
	}

	err := h.browserSvc.UpsertPermission(ctx, perm)
	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "config updated"}); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}

// GetUsage gets browser usage statistics for an agent.
func (h *Handler) GetUsage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	agentID := vars["agent_id"]
	ctx := r.Context()

	svc, ok := h.browserSvc.(*browser.Service)
	if !ok {
		apierror.WriteError(w, apierror.NewInternal("service not available"))
		return
	}

	stats, err := svc.GetUsageStats(ctx, agentID)
	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		logrus.WithError(err).Warn("failed to encode response")
	}
}

func serverError(w http.ResponseWriter, r *http.Request, err error) {
	logrus.WithError(err).WithFields(logrus.Fields{
		"request_uri": r.RequestURI,
		"method":      r.Method,
	}).Error("internal server error")
	apierror.WriteError(w, apierror.NewInternal("internal server error"))
}

func clientError(w http.ResponseWriter, r *http.Request, err error) {
	apierror.LogAndBadRequest(w, r, err, "browser client error")
}
