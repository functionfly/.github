package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Connector interface {
	Name() string
	Icon() string
	IsConfigured() bool
	Authenticate(ctx context.Context, creds map[string]string) error
	FetchData(ctx context.Context, config map[string]interface{}) (interface{}, error)
	Search(ctx context.Context, query string, config map[string]interface{}) ([]SearchResult, error)
}

type SearchResult struct {
	Title    string                 `json:"title"`
	Content  string                 `json:"content"`
	URL      string                 `json:"url"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata"`
}

type ConnectorRegistry struct {
	connectors map[string]Connector
	mu         sync.RWMutex
}

func NewConnectorRegistry(logger *logrus.Logger) *ConnectorRegistry {
	registry := &ConnectorRegistry{
		connectors: make(map[string]Connector),
	}

	registry.registerConnector("github", NewGitHubConnector(logger))
	registry.registerConnector("slack", NewSlackConnector(logger))
	registry.registerConnector("postgres", NewPostgresConnector(logger))
	registry.registerConnector("web", NewWebConnector(logger))
	registry.registerConnector("files", NewFilesConnector(logger))

	return registry
}

func (r *ConnectorRegistry) registerConnector(name string, connector Connector) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connectors[name] = connector
}

func (r *ConnectorRegistry) Get(name string) Connector {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.connectors[name]
}

func (r *ConnectorRegistry) ListConnectors() []map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []map[string]string
	for name, conn := range r.connectors {
		result = append(result, map[string]string{
			"name": name,
			"icon": conn.Icon(),
		})
	}
	return result
}

type ConnectorHandler struct {
	repo     *Repository
	registry *ConnectorRegistry
	logger   *logrus.Logger
}

func NewConnectorHandler(repo *Repository, registry *ConnectorRegistry, logger *logrus.Logger) *ConnectorHandler {
	if logger == nil {
		logger = logrus.New()
	}
	return &ConnectorHandler{
		repo:     repo,
		registry: registry,
		logger:   logger,
	}
}

type RegisterConnectorRequest struct {
	Name        string                  `json:"name"`
	Type        string                  `json:"type"`
	Config      map[string]interface{}  `json:"config"`
	Credentials map[string]string      `json:"credentials"`
}

func (h *ConnectorHandler) RegisterConnector(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	var req RegisterConnectorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	connector := h.registry.Get(req.Type)
	if connector == nil {
		http.Error(w, `{"error":"Unknown connector type"}`, http.StatusBadRequest)
		return
	}

	conn := &ChatConnector{
		TenantID: user.TenantID,
		UserID:   user.UserID,
		Name:     req.Name,
		Type:     req.Type,
		Config:   req.Config,
	}

	if err := h.repo.CreateConnector(r.Context(), conn); err != nil {
		h.logger.WithError(err).Error("Failed to save connector")
		http.Error(w, `{"error":"Failed to register connector"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(conn)
}

func (h *ConnectorHandler) ListConnectors(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	connectors, err := h.repo.ListConnectors(r.Context(), user.TenantID, user.UserID, 50, 0)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list connectors")
		http.Error(w, `{"error":"Failed to list connectors"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"connectors": connectors})
}

func (h *ConnectorHandler) DeleteConnector(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]
	connID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, `{"error":"Invalid connector ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.DeleteConnector(r.Context(), connID, user.TenantID, user.UserID); err != nil {
		h.logger.WithError(err).Error("Failed to delete connector")
		http.Error(w, `{"error":"Failed to delete connector"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h *ConnectorHandler) TestConnector(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	id := mux.Vars(r)["id"]
	connID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, `{"error":"Invalid connector ID"}`, http.StatusBadRequest)
		return
	}

	conn, err := h.repo.GetConnector(r.Context(), connID)
	if err != nil || conn == nil {
		http.Error(w, `{"error":"Connector not found"}`, http.StatusNotFound)
		return
	}

	connector := h.registry.Get(conn.Type)
	if connector == nil {
		http.Error(w, `{"error":"Connector implementation not found"}`, http.StatusInternalServerError)
		return
	}

	var testReq struct {
		Credentials map[string]string `json:"credentials"`
	}
	json.NewDecoder(r.Body).Decode(&testReq)

	err = connector.Authenticate(r.Context(), testReq.Credentials)
	success := err == nil

	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"error":   errMsg,
	})
}

func (h *ConnectorHandler) RegisterRoutes(router *mux.Router, authMw *middleware.AuthMiddleware) {
	router.HandleFunc("/chat/connectors", authMw.RequireAuth(h.RegisterConnector)).Methods("POST")
	router.HandleFunc("/chat/connectors", authMw.RequireAuth(h.ListConnectors)).Methods("GET")
	router.HandleFunc("/chat/connectors/{id}", authMw.RequireAuth(h.DeleteConnector)).Methods("DELETE")
	router.HandleFunc("/chat/connectors/{id}/test", authMw.RequireAuth(h.TestConnector)).Methods("POST")
}
