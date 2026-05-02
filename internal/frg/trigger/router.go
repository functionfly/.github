// Package trigger provides the event trigger system for FRG (Function Runtime Graph).
//
// The trigger system makes graphs reactive by:
// - Webhook triggers: HTTP endpoints that start graph execution
// - Cron triggers: Scheduled execution
// - State triggers: Database changes that trigger execution
//
// Part of the "Backend as a Graph" vision - "Event-Driven Everything"
package trigger

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/frg"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

// Router manages trigger registration and routing for graph execution
type Router struct {
	frgRepo       *frg.Repository
	engine        *frg.ExecutionEngine
	cron          *cron.Cron
	webhooks      map[string]*WebhookTrigger // path -> trigger
	webhookSecret string                    // Secret for webhook signature verification
	mu            sync.RWMutex
}

// WebhookTrigger represents a registered webhook trigger
type WebhookTrigger struct {
	GraphID   uuid.UUID
	GraphName string // author/name format
	Path      string // e.g., "/api/signup"
	Method    string // POST, GET, etc.
	Config    map[string]interface{}
}

// CronTrigger represents a registered cron trigger
type CronTrigger struct {
	GraphID   uuid.UUID
	GraphName string
	CronExpr  string
	EntryID   cron.EntryID
}

// NewRouter creates a new trigger router
func NewRouter(frgRepo *frg.Repository, engine *frg.ExecutionEngine) *Router {
	webhookSecret := os.Getenv("FRG_WEBHOOK_SECRET")
	return &Router{
		frgRepo:       frgRepo,
		engine:        engine,
		cron:          cron.New(cron.WithSeconds()),
		webhooks:      make(map[string]*WebhookTrigger),
		webhookSecret: webhookSecret,
	}
}

// Start starts the trigger router (cron scheduler, etc.)
func (r *Router) Start() {
	r.cron.Start()
	logrus.Info("Trigger router started")
}

// Stop stops the trigger router
func (r *Router) Stop() {
	r.cron.Stop()
	logrus.Info("Trigger router stopped")
}

// RegisterWebhook registers a webhook trigger for a graph
func (r *Router) RegisterWebhook(graphID uuid.UUID, graphName string, triggerConfig json.RawMessage) error {
	var config struct {
		Type   string                 `json:"type"`
		Config map[string]interface{} `json:"config"`
	}
	if err := json.Unmarshal(triggerConfig, &config); err != nil {
		return fmt.Errorf("failed to unmarshal trigger config: %w", err)
	}

	if config.Type != "webhook" {
		return fmt.Errorf("trigger type is not webhook: %s", config.Type)
	}

	path, ok := config.Config["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("webhook config missing 'path'")
	}

	method := "POST"
	if m, ok := config.Config["method"].(string); ok && m != "" {
		method = m
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for conflicts
	if existing, ok := r.webhooks[path]; ok {
		return fmt.Errorf("webhook path already registered: %s (graph: %s)", path, existing.GraphName)
	}

	r.webhooks[path] = &WebhookTrigger{
		GraphID:   graphID,
		GraphName: graphName,
		Path:      path,
		Method:    method,
		Config:    config.Config,
	}

	logrus.WithFields(logrus.Fields{
		"graph_id":   graphID,
		"graph_name": graphName,
		"path":       path,
		"method":     method,
	}).Info("Registered webhook trigger")

	return nil
}

// UnregisterWebhook removes a webhook trigger
func (r *Router) UnregisterWebhook(graphID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for path, trigger := range r.webhooks {
		if trigger.GraphID == graphID {
			delete(r.webhooks, path)
			logrus.WithFields(logrus.Fields{
				"graph_id": graphID,
				"path":     path,
			}).Info("Unregistered webhook trigger")
			return nil
		}
	}

	return fmt.Errorf("webhook not found for graph: %s", graphID)
}

// GetWebhookHandler returns the HTTP handler for webhooks
func (r *Router) GetWebhookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path

		r.mu.RLock()
		trigger, ok := r.webhooks[path]
		r.mu.RUnlock()

		if !ok {
			http.Error(w, "Webhook not found", http.StatusNotFound)
			return
		}

		if req.Method != trigger.Method {
			http.Error(w, fmt.Sprintf("Method not allowed. Expected: %s", trigger.Method), http.StatusMethodNotAllowed)
			return
		}

		// Verify webhook signature if secret is configured
		if r.webhookSecret != "" {
			signature := req.Header.Get("X-Webhook-Signature")
			if signature == "" {
				logrus.Warn("Webhook signature missing")
				http.Error(w, "Missing webhook signature", http.StatusUnauthorized)
				return
			}

			// Read body for signature verification
			body, err := io.ReadAll(req.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}

			if !r.verifyWebhookSignature(body, signature) {
				logrus.Warn("Webhook signature verification failed")
				http.Error(w, "Invalid webhook signature", http.StatusUnauthorized)
				return
			}
		}

		// Parse input from request
		var inputData map[string]interface{}
		if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" {
			contentType := req.Header.Get("Content-Type")
			if contentType == "application/json" {
				if err := json.NewDecoder(req.Body).Decode(&inputData); err != nil {
					logrus.WithError(err).Warn("Failed to decode webhook body")
					// Continue with empty input
					inputData = make(map[string]interface{})
				}
			}
		}

		// Add query parameters
		for key, values := range req.URL.Query() {
			if len(values) > 0 {
				inputData[key] = values[0]
			}
		}

		// Add headers
		headers := make(map[string]string)
		for key, values := range req.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		inputData["_headers"] = headers

		// Get graph definition
		ctx := req.Context()
		parts := strings.Split(trigger.GraphName, "/")
		if len(parts) != 2 {
			http.Error(w, "Invalid graph name", http.StatusInternalServerError)
			return
		}

		author, name := parts[0], parts[1]
		graphDef, err := r.frgRepo.GetDefinitionByName(ctx, author, name, "v1")
		if err != nil {
			logrus.WithError(err).WithField("graph", trigger.GraphName).Error("Failed to get graph definition")
			http.Error(w, "Graph not found", http.StatusNotFound)
			return
		}

		// Execute based on execution mode
		switch graphDef.ExecutionMode {
		case frg.ExecutionModeSync:
			// Execute synchronously and return result
			result, err := r.engine.ExecuteSync(ctx, graphDef, inputData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(result)

		case frg.ExecutionModeAsync, frg.ExecutionModeEventDriven, frg.ExecutionModeStreaming:
			// Start async execution and return instance ID
			instance, err := r.engine.ExecuteAsync(ctx, graphDef, inputData)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"instance_id": instance.ID,
				"status":      "started",
				"graph":       trigger.GraphName,
			})

		default:
			http.Error(w, "Unsupported execution mode", http.StatusInternalServerError)
		}
	}
}

// verifyWebhookSignature verifies the webhook signature using HMAC-SHA256
func (r *Router) verifyWebhookSignature(payload []byte, signature string) bool {
	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(r.webhookSecret))
	mac.Write(payload)
	expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	// Compare signatures using constant-time comparison to prevent timing attacks
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// RegisterCron registers a cron trigger for a graph
func (r *Router) RegisterCron(graphID uuid.UUID, graphName string, triggerConfig json.RawMessage) error {
	var config struct {
		Type   string                 `json:"type"`
		Config map[string]interface{} `json:"config"`
	}
	if err := json.Unmarshal(triggerConfig, &config); err != nil {
		return fmt.Errorf("failed to unmarshal trigger config: %w", err)
	}

	if config.Type != "schedule" {
		return fmt.Errorf("trigger type is not schedule: %s", config.Type)
	}

	cronExpr, ok := config.Config["cron"].(string)
	if !ok || cronExpr == "" {
		return fmt.Errorf("schedule config missing 'cron' expression")
	}

	// Validate cron expression
	if _, err := cron.ParseStandard(cronExpr); err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	entryID, err := r.cron.AddFunc(cronExpr, func() {
		if err := r.triggerGraph(graphID, graphName); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"graph_id":   graphID,
				"graph_name": graphName,
			}).Error("Cron trigger failed")
		}
	})

	if err != nil {
		return fmt.Errorf("failed to register cron: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"graph_id":   graphID,
		"graph_name": graphName,
		"cron":       cronExpr,
		"entry_id":   entryID,
	}).Info("Registered cron trigger")

	return nil
}

// triggerGraph triggers a graph execution
func (r *Router) triggerGraph(graphID uuid.UUID, graphName string) error {
	ctx := context.Background()

	parts := strings.Split(graphName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid graph name format: %s", graphName)
	}

	author, name := parts[0], parts[1]
	graphDef, err := r.frgRepo.GetDefinitionByName(ctx, author, name, "v1")
	if err != nil {
		return fmt.Errorf("failed to get graph definition: %w", err)
	}

	input := map[string]interface{}{
		"trigger_type": "cron",
		"timestamp":    time.Now().UTC(),
	}

	_, err = r.engine.ExecuteAsync(ctx, graphDef, input)
	if err != nil {
		return fmt.Errorf("failed to execute graph: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"graph_id":   graphID,
		"graph_name": graphName,
	}).Info("Triggered graph execution")

	return nil
}

// LoadGraphTriggers loads all triggers for a graph from the database
func (r *Router) LoadGraphTriggers(graphID uuid.UUID, graphName string, triggerConfig json.RawMessage) error {
	if len(triggerConfig) == 0 {
		return nil // No trigger config
	}

	var config struct {
		Type   string                 `json:"type"`
		Config map[string]interface{} `json:"config"`
	}
	if err := json.Unmarshal(triggerConfig, &config); err != nil {
		return fmt.Errorf("failed to unmarshal trigger config: %w", err)
	}

	switch config.Type {
	case "webhook":
		return r.RegisterWebhook(graphID, graphName, triggerConfig)
	case "schedule":
		return r.RegisterCron(graphID, graphName, triggerConfig)
	case "state_trigger":
		// State triggers are handled separately via database triggers
		logrus.WithField("graph_id", graphID).Info("State trigger configured (database-level)")
		return nil
	default:
		return fmt.Errorf("unknown trigger type: %s", config.Type)
	}
}

// SetupWebhookRoutes sets up the webhook routes on the given router
func SetupWebhookRoutes(r *mux.Router, triggerRouter *Router) {
	// Dynamic webhook handler - all webhooks go through this
	r.HandleFunc("/webhook/{path:.*}", triggerRouter.GetWebhookHandler())

	// Also register specific known webhooks
	r.HandleFunc("/api/webhooks/graph/{graph_id}", triggerRouter.GetWebhookHandler()).Methods("POST")
}

// StartTriggerLoader loads all published graphs with triggers on startup
func (r *Router) StartTriggerLoader(ctx context.Context) error {
	// Get all published graphs with triggers from the database
	// Query for graphs where trigger_config is not null and visibility is public/unlisted
	var graphs []*frg.GraphDefinition
	err := r.frgRepo.QueryPublishedGraphsWithTriggers(ctx, &graphs)
	if err != nil {
		return fmt.Errorf("failed to list graphs with triggers: %w", err)
	}

	for _, graph := range graphs {
		graphName := fmt.Sprintf("%s/%s", graph.Author, graph.Name)
		if err := r.LoadGraphTriggers(graph.ID, graphName, graph.TriggerConfig); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"graph_id":   graph.ID,
				"graph_name": graphName,
			}).Warn("Failed to load graph triggers")
			// Continue with other graphs
		}
	}

	logrus.WithField("count", len(graphs)).Info("Loaded triggers for graphs")
	return nil
}
