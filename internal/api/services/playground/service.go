package playground

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/functionregistry"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ExecuteResponse struct {
	Success     bool        `json:"success"`
	Output      interface{} `json:"output,omitempty"`
	Error       string      `json:"error,omitempty"`
	LatencyMs   int64       `json:"latency_ms"`
	StatusCode  int         `json:"status_code"`
	Version     string      `json:"version,omitempty"`
	ExecutionID string      `json:"execution_id,omitempty"`
}

type RegistryRepo interface {
	GetFunctionByAuthorName(author, name string) (*registry.RegistryFunction, error)
	GetLatestFunctionVersion(fnID uuid.UUID) (*registry.RegistryFunctionVersion, error)
	GetFunctionVersion(fnID uuid.UUID, version string) (*registry.RegistryFunctionVersion, error)
	GetFunctionByID(id uuid.UUID) (*registry.RegistryFunction, error)
	CreateExecutionPublic(exec *registry.RegistryExecutionPublic) error
	GetExecutionPublicByID(id string) (*registry.RegistryExecutionPublic, error)
	GetRatingByFunctionID(fnID uuid.UUID) (*registry.RegistryFunctionRating, error)
}

type AppRepo interface {
	GetAppBySlug(slug string) (*AppInfo, error)
	GetFunctionByAppIDAndName(ctx context.Context, appID uuid.UUID, name string) (*FunctionConfig, error)
	GetActiveDeploymentForFunction(ctx context.Context, fnID uuid.UUID) (*FunctionDeployment, error)
}

type AppInfo struct {
	ID   uuid.UUID
	Slug string
}

type FunctionConfig struct {
	ID                uuid.UUID
	Name              string
	Version           string
	Status            string
	PlaygroundEnabled bool
	PlaygroundConfig  map[string]interface{}
}

type FunctionDeployment struct {
	Provider    string
	Region      string
	DeployedURL *string
}

type PlaygroundService struct {
	registryRepo RegistryRepo
	appRepo      AppRepo
	logger       *logrus.Logger
	wsHub        *WebSocketHub
}

func NewPlaygroundService(registryRepo RegistryRepo, appRepo AppRepo, logger *logrus.Logger) *PlaygroundService {
	svc := &PlaygroundService{
		registryRepo: registryRepo,
		appRepo:      appRepo,
		logger:       logger,
	}
	svc.wsHub = NewWebSocketHub(logger)
	go svc.wsHub.Run()
	return svc
}

func (s *PlaygroundService) WebSocketHub() *WebSocketHub {
	return s.wsHub
}

func (s *PlaygroundService) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := s.registryRepo.GetFunctionByAuthorName(author, name)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	fnVersion, err := s.registryRepo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function version not found"))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.WithError(err).Error("Failed to upgrade WebSocket")
		return
	}

	client := &WSClient{
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Hub:     s.wsHub,
		FuncID:  fn.ID,
		Version: fnVersion.Version,
	}

	s.wsHub.Register(client)

	go client.writePump()
	go client.readPump()

	s.sendInitialInfo(client, fn, fnVersion)
}

func (s *PlaygroundService) sendInitialInfo(client *WSClient, fn *registry.RegistryFunction, fnVersion *registry.RegistryFunctionVersion) {
	var manifest functionregistry.FunctionManifest
	if fnVersion.Manifest != nil {
		json.Unmarshal(fnVersion.Manifest, &manifest)
	}

	info := map[string]interface{}{
		"type": "init",
		"function": map[string]interface{}{
			"id":       fn.ID.String(),
			"author":   fn.Author,
			"name":     fn.Name,
			"version":  fnVersion.Version,
			"runtime":  fnVersion.Runtime,
			"manifest": manifest,
		},
	}
	data, _ := json.Marshal(info)
	client.Send <- data
}

func (s *PlaygroundService) HandleWebSocketExecute(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	fn, err := s.registryRepo.GetFunctionByAuthorName(author, name)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	fnVersion, err := s.registryRepo.GetLatestFunctionVersion(fn.ID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function version not found"))
		return
	}

	var req struct {
		Input json.RawMessage `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	result := s.ExecuteRegistry(fn, fnVersion, req.Input, false)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *PlaygroundService) ExecuteRegistry(fn *registry.RegistryFunction, fnVersion *registry.RegistryFunctionVersion, input json.RawMessage, record bool) *ExecuteResponse {
	return s.executeRegistryFunction(fn, fnVersion, input, record)
}

func (s *PlaygroundService) ExecuteApp(ctx context.Context, appSlug, functionName string, input json.RawMessage) *ExecuteResponse {
	app, err := s.appRepo.GetAppBySlug(appSlug)
	if err != nil {
		return &ExecuteResponse{Success: false, Error: "App not found"}
	}

	fn, err := s.appRepo.GetFunctionByAppIDAndName(ctx, app.ID, functionName)
	if err != nil {
		return &ExecuteResponse{Success: false, Error: "Function not found"}
	}

	if !fn.PlaygroundEnabled {
		return &ExecuteResponse{Success: false, Error: "Playground not enabled"}
	}

	deployment, err := s.appRepo.GetActiveDeploymentForFunction(ctx, fn.ID)
	if err != nil || deployment == nil || deployment.DeployedURL == nil {
		return &ExecuteResponse{Success: false, Error: "Function not deployed"}
	}

	start := time.Now()
	targetURL := *deployment.DeployedURL + "/execute"

	proxyReq, err := http.NewRequestWithContext(ctx, "POST", targetURL, bytes.NewReader(input))
	if err != nil {
		return &ExecuteResponse{Success: false, Error: "Failed to create request"}
	}

	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("X-FunctionFly-Playground", "true")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(proxyReq)
	latencyMs := time.Since(start).Milliseconds()

	if err != nil {
		return &ExecuteResponse{Success: false, Error: err.Error(), LatencyMs: latencyMs}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var output interface{}
	json.Unmarshal(body, &output)

	return &ExecuteResponse{
		Success:    resp.StatusCode >= 200 && resp.StatusCode < 300,
		Output:     output,
		LatencyMs:  latencyMs,
		StatusCode: resp.StatusCode,
	}
}

func (s *PlaygroundService) executeRegistryFunction(fn *registry.RegistryFunction, fnVersion *registry.RegistryFunctionVersion, input json.RawMessage, record bool) *ExecuteResponse {
	start := time.Now()

	execRequest := functionregistry.ExecutionRequest{
		Author:  fn.Author,
		Name:    fn.Name,
		Version: fnVersion.Version,
		Input:   input,
	}

	reqBytes, _ := json.Marshal(execRequest)
	serverPort := os.Getenv("PORT")
	if serverPort == "" {
		serverPort = "8090"
	}
	targetURL := fmt.Sprintf("http://localhost:%s/v1/fx/%s/%s@%s", serverPort, fn.Author, fn.Name, fnVersion.Version)

	proxyReq, _ := http.NewRequestWithContext(context.Background(), "POST", targetURL, bytes.NewReader(reqBytes))
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("X-FunctionFly-Playground", "true")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(proxyReq)
	latencyMs := time.Since(start).Milliseconds()

	if err != nil {
		return &ExecuteResponse{
			Success:   false,
			Error:     err.Error(),
			LatencyMs: latencyMs,
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var output interface{}
	json.Unmarshal(body, &output)

	var execResponse functionregistry.ExecutionResponse
	if err := json.Unmarshal(body, &execResponse); err != nil {
		var errorResp functionregistry.ExecutionError
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return &ExecuteResponse{
				Success:    resp.StatusCode >= 200 && resp.StatusCode < 300,
				Output:     output,
				LatencyMs:  latencyMs,
				StatusCode: resp.StatusCode,
			}
		}
		return &ExecuteResponse{
			Success:    errorResp.OK,
			Error:      errorResp.Error.Message,
			LatencyMs:  int64(errorResp.DurationMs),
			StatusCode: resp.StatusCode,
		}
	}

	response := &ExecuteResponse{
		Success:    execResponse.OK,
		Output:     execResponse.Data,
		LatencyMs:  int64(execResponse.DurationMs),
		StatusCode: resp.StatusCode,
		Version:    execResponse.Version,
	}

	if record {
		var outputJSON json.RawMessage
		if execResponse.Data != nil {
			outputJSON, _ = json.Marshal(execResponse.Data)
		}
		executionID := s.recordExecution(fn.ID.String(), fnVersion.ID.String(), input, outputJSON, resp.StatusCode, int(latencyMs))
		response.ExecutionID = executionID
	}

	return response
}

func (s *PlaygroundService) recordExecution(functionID, versionID string, input, output json.RawMessage, statusCode, durationMs int) string {
	publicID := generatePublicID()
	fnID, err := uuid.Parse(functionID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to parse function ID")
		return ""
	}

	exec := &registry.RegistryExecutionPublic{
		PublicID:   publicID,
		FunctionID: fnID,
		Version:    versionID,
		InputJSON:  input,
		OutputJSON: output,
		DurationMs: durationMs,
		Cached:     false,
		Shareable:  true,
	}

	if err := s.registryRepo.CreateExecutionPublic(exec); err != nil {
		s.logger.WithError(err).Error("Failed to create public execution")
		return ""
	}

	return publicID
}

func (s *PlaygroundService) ValidateInputAgainstSchema(input json.RawMessage, manifest *functionregistry.FunctionManifest) (bool, string) {
	if manifest == nil || manifest.Input == nil || manifest.Input.Schema == nil {
		return true, ""
	}

	var data interface{}
	if err := json.Unmarshal(input, &data); err != nil {
		return false, fmt.Sprintf("Invalid JSON input: %v", err)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(manifest.Input.Schema, &schema); err != nil {
		return true, ""
	}

	if properties, ok := schema["properties"].(map[string]interface{}); ok {
		if inputMap, ok := data.(map[string]interface{}); ok {
			for key, value := range inputMap {
				if prop, ok := properties[key].(map[string]interface{}); ok {
					if propType, ok := prop["type"].(string); ok {
						if !validateType(value, propType) {
							return false, fmt.Sprintf("Field '%s' must be of type %s", key, propType)
						}
					}
				}
			}
		}
	}

	return true, ""
}

func validateType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		switch value.(type) {
		case float64:
			return true
		case int:
			return true
		}
		return false
	case "integer":
		switch v := value.(type) {
		case int:
			return true
		case int64:
			return true
		case float64:
			return v == float64(int(v))
		}
		return false
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "array":
		_, ok := value.([]interface{})
		return ok
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	}
	return true
}

func generatePublicID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return "exec_" + base64.URLEncoding.EncodeToString(b)[:10]
}

type WebSocketHub struct {
	clients    map[*WSClient]bool
	register   chan *WSClient
	unregister chan *WSClient
	broadcast  chan []byte
	logger     *logrus.Logger
	stop       chan struct{}
}

type WSClient struct {
	Conn    *websocket.Conn
	Send    chan []byte
	Hub     *WebSocketHub
	FuncID  uuid.UUID
	Version string
}

func NewWebSocketHub(logger *logrus.Logger) *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*WSClient]bool),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
		broadcast:  make(chan []byte, 256),
		logger:     logger,
		stop:       make(chan struct{}),
	}
}

func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.logger.WithField("func_id", client.FuncID).Debug("WS client registered")
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
		case <-h.stop:
			return
		}
	}
}

func (h *WebSocketHub) Register(client *WSClient) {
	h.register <- client
}

func (h *WebSocketHub) Unregister(client *WSClient) {
	h.unregister <- client
}

func (c *WSClient) readPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.WithError(err).Error("WS read error")
			}
			break
		}

		var msg struct {
			Type  string          `json:"type"`
			Input json.RawMessage `json:"input"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		if msg.Type == "execute" {
			c.Hub.broadcast <- message
		}
	}
}

func (c *WSClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}