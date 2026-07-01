package agent

import (
	"encoding/json"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// MCPServerHandler handles agent MCP server CRUD endpoints
type MCPServerHandler struct {
	repo *storage.AgentMCPServerRepository
}

// NewMCPServerHandler creates a new MCP server handler
func NewMCPServerHandler(repo *storage.AgentMCPServerRepository) *MCPServerHandler {
	return &MCPServerHandler{repo: repo}
}

// RegisterRoutes registers MCP server routes under /agent/{agent_id}/mcp-servers
func (h *MCPServerHandler) RegisterRoutes(r *mux.Router, prefix string, auth *middleware.AuthMiddleware) {
	s := r.PathPrefix(prefix + "/agent/{agent_id}/mcp-servers").Subrouter()
	s.HandleFunc("", auth.RequireAuth(h.handleList)).Methods("GET", "OPTIONS")
	s.HandleFunc("", auth.RequireAuth(h.handleCreate)).Methods("POST", "OPTIONS")
	s.HandleFunc("/{server_id}", auth.RequireAuth(h.handleGet)).Methods("GET", "OPTIONS")
	s.HandleFunc("/{server_id}", auth.RequireAuth(h.handleUpdate)).Methods("PATCH", "OPTIONS")
	s.HandleFunc("/{server_id}", auth.RequireAuth(h.handleDelete)).Methods("DELETE", "OPTIONS")
}

type createMCPServerRequest struct {
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Transport   string            `json:"transport"`
	Description string            `json:"description"`
	Headers     map[string]string `json:"headers"`
}

type updateMCPServerRequest struct {
	Name        *string            `json:"name"`
	URL         *string            `json:"url"`
	Transport   *string            `json:"transport"`
	Description *string            `json:"description"`
	Enabled     *bool              `json:"enabled"`
	Headers     *map[string]string `json:"headers"`
}

func (h *MCPServerHandler) handleList(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_AGENT_ID", "agent_id is required")
		return
	}

	servers, err := h.repo.ListByAgent(r.Context(), agentID, claims.TenantID)
	if err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "LIST_FAILED", "failed to list MCP servers", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"servers": servers,
	})
}

func (h *MCPServerHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	agentID := mux.Vars(r)["agent_id"]
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_AGENT_ID", "agent_id is required")
		return
	}

	var req createMCPServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "MISSING_NAME", "name is required")
		return
	}
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "MISSING_URL", "url is required")
		return
	}

	transport := req.Transport
	if transport == "" {
		transport = "streamable-http"
	}
	if transport != "streamable-http" && transport != "stdio" && transport != "sse" {
		writeError(w, http.StatusBadRequest, "INVALID_TRANSPORT", "transport must be streamable-http, stdio, or sse")
		return
	}

	existing, _ := h.repo.GetByAgentAndURL(r.Context(), agentID, req.URL, claims.TenantID)
	if existing != nil {
		writeError(w, http.StatusConflict, "DUPLICATE_SERVER", "an MCP server with this URL already exists for this agent")
		return
	}

	headers := make(storage.JSONMap)
	for k, v := range req.Headers {
		headers[k] = v
	}

	server := &storage.AgentMCPServer{
		ID:          uuid.New(),
		AgentID:     agentID,
		TenantID:    claims.TenantID,
		Name:        req.Name,
		URL:         req.URL,
		Transport:   transport,
		Description: req.Description,
		Enabled:     true,
		Headers:     headers,
	}

	if err := h.repo.Create(r.Context(), server); err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "CREATE_FAILED", "failed to create MCP server", err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"ok":     true,
		"server": server,
	})
}

func (h *MCPServerHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	serverID, err := uuid.Parse(mux.Vars(r)["server_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid server id")
		return
	}

	server, err := h.repo.GetByID(r.Context(), serverID, claims.TenantID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"server": server,
	})
}

func (h *MCPServerHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	serverID, err := uuid.Parse(mux.Vars(r)["server_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid server id")
		return
	}

	existing, err := h.repo.GetByID(r.Context(), serverID, claims.TenantID)
	if err != nil || existing == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server not found")
		return
	}

	var req updateMCPServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.URL != nil {
		updates["url"] = *req.URL
	}
	if req.Transport != nil {
		updates["transport"] = *req.Transport
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Headers != nil {
		updates["headers"] = *req.Headers
	}

	if len(updates) == 0 {
		writeError(w, http.StatusBadRequest, "NO_FIELDS", "no fields to update")
		return
	}

	if err := h.repo.Update(r.Context(), serverID, updates); err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "UPDATE_FAILED", "failed to update MCP server", err)
		return
	}

	updated, _ := h.repo.GetByID(r.Context(), serverID, claims.TenantID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":     true,
		"server": updated,
	})
}

func (h *MCPServerHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	serverID, err := uuid.Parse(mux.Vars(r)["server_id"])
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "invalid server id")
		return
	}

	existing, err := h.repo.GetByID(r.Context(), serverID, claims.TenantID)
	if err != nil || existing == nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "MCP server not found")
		return
	}

	if err := h.repo.Delete(r.Context(), serverID); err != nil {
		writeErrorFromErr(r, w, http.StatusInternalServerError, "DELETE_FAILED", "failed to delete MCP server", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"message": "MCP server deleted",
	})
}
