package atlas

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/sirupsen/logrus"
)

const (
	DefaultBaseURL       = "http://localhost:7447"
	tokenRefreshBuffer   = 30 * time.Second
	defaultTokenTTL      = 3600 * time.Second
	maxBatchSize         = 100
	flushInterval        = 2 * time.Second
)

type EventKind string

const (
	EventKindInput    EventKind = "input"
	EventKindDecision EventKind = "decision"
	EventKindAction   EventKind = "action"
	EventKindResult   EventKind = "result"
	EventKindError    EventKind = "error"
)

type Event struct {
	EventID        string          `json:"event_id"`
	RunID          string          `json:"run_id"`
	TimestampNs    uint64          `json:"timestamp_ns"`
	Sequence       uint64          `json:"sequence"`
	SystemID       string          `json:"system_id"`
	TargetSystemID *string         `json:"target_system_id,omitempty"`
	Kind           EventKind       `json:"kind"`
	Parent         *string         `json:"parent,omitempty"`
	Payload        json.RawMessage `json:"payload"`
}

type RunRecord struct {
	TenantID       string          `json:"tenant_id"`
	RunID          string          `json:"run_id"`
	Labels         json.RawMessage `json:"labels"`
	FirstEventTsNs *uint64         `json:"first_event_ts_ns,omitempty"`
	LastEventTsNs  *uint64         `json:"last_event_ts_ns,omitempty"`
	EventCount     uint64          `json:"event_count"`
	CreatedAtNs    uint64          `json:"created_at_ns"`
}

type GraphNode struct {
	ID          string  `json:"id"`
	Kind        string  `json:"kind"`
	SystemID    string  `json:"system_id"`
	Sequence    uint64  `json:"sequence"`
	TimestampNs uint64  `json:"timestamp_ns"`
	Parent      *string `json:"parent,omitempty"`
}

type GraphEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation"`
}

type GraphResponse struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type SearchRequest struct {
	Kind           *EventKind     `json:"kind,omitempty"`
	SystemID       *string        `json:"system_id,omitempty"`
	SinceNs        *uint64        `json:"since_ns,omitempty"`
	UntilNs        *uint64        `json:"until_ns,omitempty"`
	PayloadMatches []PayloadMatch `json:"payload_matches,omitempty"`
	Limit          uint32         `json:"limit,omitempty"`
}

type PayloadMatch struct {
	Path   string          `json:"path"`
	Equals json.RawMessage `json:"equals"`
}

type SearchResponse struct {
	Events      []Event `json:"events"`
	ScannedRuns uint32  `json:"scanned_runs"`
	Truncated   bool    `json:"truncated"`
}

type apiError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (e *apiError) String() string {
	return fmt.Sprintf("atlas api error: %s: %s", e.Error, e.Message)
}

type tokenCache struct {
	mu         sync.RWMutex
	accessToken string
	expiresAt   time.Time
}

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	logger     *logrus.Logger
	token      tokenCache
}

func NewClient(baseURL, apiKey string, _ ...interface{}) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logrus.StandardLogger(),
	}
}

func newULID() string {
	return ulid.Make().String()
}

func newULIDAt(t time.Time) string {
	return ulid.MustNew(ulid.Timestamp(t), nil).String()
}

func (c *Client) SetLogger(logger *logrus.Logger) {
	c.logger = logger
}

func (c *Client) getToken(ctx context.Context) (string, error) {
	c.token.mu.RLock()
	if c.token.accessToken != "" && time.Now().Add(tokenRefreshBuffer).Before(c.token.expiresAt) {
		t := c.token.accessToken
		c.token.mu.RUnlock()
		return t, nil
	}
	c.token.mu.RUnlock()

	return c.refreshToken(ctx)
}

func (c *Client) refreshToken(ctx context.Context) (string, error) {
	c.token.mu.Lock()
	defer c.token.mu.Unlock()

	if c.token.accessToken != "" && time.Now().Add(tokenRefreshBuffer).Before(c.token.expiresAt) {
		return c.token.accessToken, nil
	}

	body, _ := json.Marshal(map[string]string{"secret": c.apiKey})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/auth/token", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("exchange token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange failed (status %d): %s", resp.StatusCode, string(b))
	}

	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   uint64 `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	ttl := defaultTokenTTL
	if tok.ExpiresIn > 0 {
		ttl = time.Duration(tok.ExpiresIn) * time.Second
	}
	c.token.accessToken = tok.AccessToken
	c.token.expiresAt = time.Now().Add(ttl)

	return tok.AccessToken, nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, err
	}

	var reqBody *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		token, err = c.refreshToken(ctx)
		if err != nil {
			return nil, err
		}
		if body != nil {
			b, _ := json.Marshal(body)
			reqBody = bytes.NewBuffer(b)
		} else {
			reqBody = &bytes.Buffer{}
		}
		req, err = http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
		if err != nil {
			return nil, fmt.Errorf("create retry request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		return c.httpClient.Do(req)
	}

	return resp, nil
}

func (c *Client) CreateRun(ctx context.Context, labels map[string]interface{}) (string, error) {
	type createRunReq struct {
		RunID  *string                `json:"run_id,omitempty"`
		Labels map[string]interface{} `json:"labels,omitempty"`
	}

	runID := newULID()
	resp, err := c.doRequest(ctx, http.MethodPost, "/v1/runs", createRunReq{RunID: &runID, Labels: labels})
	if err != nil {
		return "", fmt.Errorf("create run: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create run failed (status %d): %s", resp.StatusCode, string(b))
	}

	var result struct {
		RunID string `json:"run_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode create run response: %w", err)
	}

	if result.RunID != "" {
		return result.RunID, nil
	}
	return runID, nil
}

func (c *Client) AppendEvents(ctx context.Context, runID string, events []Event) (int, error) {
	type wireEvent struct {
		EventID        string          `json:"event_id"`
		RunID          string          `json:"run_id"`
		TimestampNs    uint64          `json:"timestamp_ns"`
		Sequence       uint64          `json:"sequence"`
		SystemID       string          `json:"system_id"`
		TargetSystemID *string         `json:"target_system_id,omitempty"`
		Kind           string          `json:"kind"`
		Parent         *string         `json:"parent,omitempty"`
		Payload        json.RawMessage `json:"payload"`
	}

	type batchReq struct {
		Events []wireEvent `json:"events"`
	}

	wire := make([]wireEvent, len(events))
	for i, e := range events {
		eid := e.EventID
		if eid == "" {
			eid = newULID()
		}
		wire[i] = wireEvent{
			EventID:     eid,
			RunID:       runID,
			TimestampNs: e.TimestampNs,
			Sequence:    e.Sequence,
			SystemID:    e.SystemID,
			Kind:        string(e.Kind),
			Parent:      e.Parent,
			Payload:     e.Payload,
		}
	}

	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/v1/runs/%s/events/batch", runID), batchReq{Events: wire})
	if err != nil {
		return 0, fmt.Errorf("append events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("append events failed (status %d): %s", resp.StatusCode, string(b))
	}

	var result struct {
		Accepted int `json:"accepted"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode append response: %w", err)
	}

	return result.Accepted, nil
}

func (c *Client) GetRun(ctx context.Context, runID string) (*RunRecord, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/v1/runs/%s", runID), nil)
	if err != nil {
		return nil, fmt.Errorf("get run: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get run failed (status %d): %s", resp.StatusCode, string(b))
	}

	var run RunRecord
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("decode run: %w", err)
	}

	return &run, nil
}

func (c *Client) ListRuns(ctx context.Context, limit int, after string) ([]RunRecord, error) {
	path := fmt.Sprintf("/v1/runs?limit=%d", limit)
	if after != "" {
		path += "&after=" + after
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("list runs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list runs failed (status %d): %s", resp.StatusCode, string(b))
	}

	var runs []RunRecord
	if err := json.NewDecoder(resp.Body).Decode(&runs); err != nil {
		return nil, fmt.Errorf("decode runs: %w", err)
	}

	return runs, nil
}

func (c *Client) GetEvents(ctx context.Context, runID string, afterSeq uint64, limit int) ([]Event, error) {
	path := fmt.Sprintf("/v1/runs/%s/events?limit=%d", runID, limit)
	if afterSeq > 0 {
		path += fmt.Sprintf("&after_seq=%d", afterSeq)
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get events failed (status %d): %s", resp.StatusCode, string(b))
	}

	var events []Event
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("decode events: %w", err)
	}

	return events, nil
}

func (c *Client) GetGraph(ctx context.Context, runID string) (*GraphResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/v1/runs/%s/graph", runID), nil)
	if err != nil {
		return nil, fmt.Errorf("get graph: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get graph failed (status %d): %s", resp.StatusCode, string(b))
	}

	var graph GraphResponse
	if err := json.NewDecoder(resp.Body).Decode(&graph); err != nil {
		return nil, fmt.Errorf("decode graph: %w", err)
	}

	return &graph, nil
}

func (c *Client) GetAncestors(ctx context.Context, runID, eventID string, maxDepth int) ([]Event, error) {
	path := fmt.Sprintf("/v1/runs/%s/graph/ancestors?event_id=%s&max_depth=%d", runID, eventID, maxDepth)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get ancestors: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get ancestors failed (status %d): %s", resp.StatusCode, string(b))
	}

	var events []Event
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("decode ancestors: %w", err)
	}

	return events, nil
}

func (c *Client) GetDescendants(ctx context.Context, runID, eventID string, maxDepth int) ([]Event, error) {
	path := fmt.Sprintf("/v1/runs/%s/graph/descendants?event_id=%s&max_depth=%d", runID, eventID, maxDepth)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get descendants: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get descendants failed (status %d): %s", resp.StatusCode, string(b))
	}

	var events []Event
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("decode descendants: %w", err)
	}

	return events, nil
}

func (c *Client) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/v1/search", req)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed (status %d): %s", resp.StatusCode, string(b))
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode search: %w", err)
	}

	return &result, nil
}

func (c *Client) HealthCheck(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Status, nil
}

// DeriveAtlasTenantID derives a deterministic tenant ID for Atlas from a UUID.
// Kept for backward compatibility with agent_observability handler.
func DeriveAtlasTenantID(tenantID interface{}) string {
	switch v := tenantID.(type) {
	case string:
		h := sha256Sum(v + "atlas-salt-v1")
		return h[:16]
	default:
		h := sha256Sum(fmt.Sprintf("%v", tenantID) + "atlas-salt-v1")
		return h[:16]
	}
}

func sha256Sum(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// AppendEvent appends a single event. Backward-compatible wrapper around AppendEvents.
func (c *Client) AppendEvent(ctx context.Context, runID string, kind string, payload json.RawMessage, systemID string, parentID string) (*Event, error) {
	var parent *string
	if parentID != "" {
		parent = &parentID
	}

	event := Event{
		TimestampNs: uint64(time.Now().UnixNano()),
		Sequence:    0,
		SystemID:    systemID,
		Kind:        EventKind(kind),
		Parent:      parent,
		Payload:     payload,
	}

	accepted, err := c.AppendEvents(ctx, runID, []Event{event})
	if err != nil {
		return nil, err
	}
	if accepted == 0 {
		return nil, fmt.Errorf("event not accepted")
	}

	return &event, nil
}

// Replay retrieves events for a run after a given sequence number.
// Backward-compatible wrapper around GetEvents.
func (c *Client) Replay(ctx context.Context, runID string, afterSequence uint64) ([]*Event, error) {
	events, err := c.GetEvents(ctx, runID, afterSequence, 10000)
	if err != nil {
		return nil, err
	}
	result := make([]*Event, len(events))
	for i := range events {
		result[i] = &events[i]
	}
	return result, nil
}

// StreamEvents returns a channel of events for a run (polling-based fallback).
func (c *Client) StreamEvents(ctx context.Context, runID string) (<-chan *Event, <-chan error) {
	events := make(chan *Event, 100)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		var lastSeq uint64
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			evts, err := c.GetEvents(ctx, runID, lastSeq, 100)
			if err != nil {
				errs <- err
				return
			}

			for i := range evts {
				events <- &evts[i]
				if evts[i].Sequence > lastSeq {
					lastSeq = evts[i].Sequence
				}
			}

			time.Sleep(2 * time.Second)
		}
	}()

	return events, errs
}

// RunStats contains aggregated stats for a run.
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

// GetStats computes basic stats from a run's events.
func (c *Client) GetStats(ctx context.Context, runID string) (*RunStats, error) {
	events, err := c.GetEvents(ctx, runID, 0, 10000)
	if err != nil {
		return nil, err
	}

	stats := &RunStats{
		AtlasRunID: runID,
		EventCount: len(events),
	}

	for _, e := range events {
		if e.Kind == EventKindError {
			stats.ErrorCount++
		}
	}

	if len(events) >= 2 {
		first := events[0].TimestampNs
		last := events[len(events)-1].TimestampNs
		if last > first {
			stats.DurationMs = int64((last - first) / 1_000_000)
		}
	}

	return stats, nil
}

// EndRun is a no-op in the current Atlas API (runs are implicitly ended).
func (c *Client) EndRun(ctx context.Context, runID string, status string) error {
	return nil
}

// GetGraphWithDepth retrieves the decision graph with a specified max depth.
func (c *Client) GetGraphWithDepth(ctx context.Context, runID string, eventID string, maxDepth int) (*GraphResponse, error) {
	return c.GetGraph(ctx, runID)
}

