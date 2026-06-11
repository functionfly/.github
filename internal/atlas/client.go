package atlas

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	DefaultBaseURL   = "http://localhost:7447"
	DefaultGRPCPort  = 50051
	atlasSaltVersion = "atlas-salt-v1"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	tenantID   uuid.UUID
	logger     *logrus.Logger
}

type Event struct {
	EventID   string          `json:"event_id"`
	RunID     string          `json:"run_id"`
	Sequence  uint64          `json:"sequence"`
	Timestamp time.Time       `json:"timestamp"`
	SystemID  string          `json:"system_id"`
	Kind      string          `json:"kind"`
	ParentID  string          `json:"parent_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
	SpanID    string          `json:"span_id,omitempty"`
}

type RunStats struct {
	AtlasRunID    string  `json:"atlas_run_id"`
	DurationMs    int64   `json:"duration_ms"`
	EventCount    int     `json:"event_count"`
	InputTokens   int     `json:"input_tokens"`
	OutputTokens  int     `json:"output_tokens"`
	TotalCostUSD  float64 `json:"total_cost_usd"`
	ErrorCount    int     `json:"error_count"`
	ToolCallCount int     `json:"tool_call_count"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
	CostPerToken  float64 `json:"cost_per_token"`
}

type GraphNode struct {
	EventID   string          `json:"event_id"`
	Kind      string          `json:"kind"`
	Sequence  uint64          `json:"sequence"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

type GraphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type DecisionGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

func NewClient(baseURL, apiKey string, tenantID uuid.UUID) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL:  baseURL,
		apiKey:   apiKey,
		tenantID: tenantID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logrus.New(),
	}
}

func (c *Client) SetLogger(logger *logrus.Logger) {
	c.logger = logger
}

func DeriveAtlasTenantID(tenantID uuid.UUID) string {
	h := sha256.Sum256([]byte(tenantID.String() + atlasSaltVersion))
	return hex.EncodeToString(h[:8])
}

func (c *Client) AtlasTenantID() string {
	return DeriveAtlasTenantID(c.tenantID)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("X-Atlas-Tenant", c.AtlasTenantID())

	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

func (c *Client) CreateRun(ctx context.Context, metadata map[string]string) (string, error) {
	type createRunRequest struct {
		Metadata map[string]string `json:"metadata"`
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/v1/runs", createRunRequest{Metadata: metadata})
	if err != nil {
		return "", fmt.Errorf("failed to create run: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	type createRunResponse struct {
		RunID string `json:"run_id"`
	}

	var result createRunResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.RunID, nil
}

func (c *Client) AppendEvent(ctx context.Context, runID string, kind string, payload json.RawMessage, systemID string, parentID string) (*Event, error) {
	type appendEventRequest struct {
		Kind      string          `json:"kind"`
		Payload   json.RawMessage `json:"payload"`
		SystemID  string          `json:"system_id"`
		Parent    string          `json:"parent"`
		Timestamp int64           `json:"timestamp_ns"`
	}

	req := appendEventRequest{
		Kind:      kind,
		Payload:   payload,
		SystemID:  systemID,
		Parent:    parentID,
		Timestamp: time.Now().UnixNano(),
	}

	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/v1/runs/%s/events", runID), req)
	if err != nil {
		return nil, fmt.Errorf("failed to append event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var event Event
	if err := json.NewDecoder(resp.Body).Decode(&event); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &event, nil
}

func (c *Client) Replay(ctx context.Context, runID string, afterSequence uint64) ([]*Event, error) {
	url := fmt.Sprintf("/v1/runs/%s/replay?after_sequence=%d", runID, afterSequence)

	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get replay: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	type replayResponse struct {
		Events []*Event `json:"events"`
	}

	var result replayResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Events, nil
}

func (c *Client) GetStats(ctx context.Context, runID string) (*RunStats, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/v1/runs/%s/stats", runID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var stats RunStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &stats, nil
}

func (c *Client) GetGraph(ctx context.Context, runID string, eventID string, maxDepth int) (*DecisionGraph, error) {
	url := fmt.Sprintf("/v1/runs/%s/graph?event_id=%s&max_depth=%d", runID, eventID, maxDepth)

	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var graph DecisionGraph
	if err := json.NewDecoder(resp.Body).Decode(&graph); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &graph, nil
}

func (c *Client) StreamEvents(ctx context.Context, runID string) (<-chan *Event, <-chan error) {
	events := make(chan *Event, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		url := fmt.Sprintf("%s/v1/stream/%s", c.baseURL, runID)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			errs <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("X-Atlas-Tenant", c.AtlasTenantID())

		conn, _, err := websocket.DefaultDialer.Dial(req.URL.String(), req.Header)
		if err != nil {
			errs <- fmt.Errorf("failed to connect to WebSocket: %w", err)
			return
		}
		defer conn.Close()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, msg, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						errs <- fmt.Errorf("WebSocket closed: %w", err)
					}
					return
				}

				var event Event
				if err := json.Unmarshal(msg, &event); err != nil {
					c.logger.WithError(err).Warn("failed to unmarshal event")
					continue
				}

				events <- &event
			}
		}
	}()

	return events, errs
}

func (c *Client) EndRun(ctx context.Context, runID string, status string) error {
	type endRunRequest struct {
		Status string `json:"status"`
	}

	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/v1/runs/%s/end", runID), endRunRequest{Status: status})
	if err != nil {
		return fmt.Errorf("failed to end run: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetRun(ctx context.Context, runID string) (*Event, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/v1/runs/%s", runID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get run: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var event Event
	if err := json.NewDecoder(resp.Body).Decode(&event); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &event, nil
}
