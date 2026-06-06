package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Handler serves the Zero-Friction Demo API.
type Handler struct {
	repo        storage.Repository
	redisClient *redis.Client
}

// NewHandler creates a new demo handler.
func NewHandler(repo storage.Repository, redisClient *redis.Client) *Handler {
	return &Handler{repo: repo, redisClient: redisClient}
}

// DemoFunction describes a built-in demo function.
type DemoFunction struct {
	Name         string            `json:"name"`
	Label        string            `json:"label"`
	Description  string            `json:"description"`
	DefaultInput map[string]string `json:"default_input,omitempty"`
}

// DemoListResponse is the public catalogue response.
type DemoListResponse struct {
	Functions []DemoFunction `json:"functions"`
}

// ListFunctions returns the three public demo functions.
func (h *Handler) ListFunctions(w http.ResponseWriter, r *http.Request) {
	resp := DemoListResponse{
		Functions: []DemoFunction{
			{Name: "web-scraper", Label: "Web Scraper", Description: "Fetches a URL and returns the page title and meta description.", DefaultInput: map[string]string{"url": "https://example.com"}},
			{Name: "text-summarizer", Label: "Text Summarizer", Description: "Summarizes long text into a short 1-2 sentence abstract."},
			{Name: "currency-converter", Label: "Currency Converter", Description: "Converts an amount between two ISO-4217 currencies using live rates."},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ExecuteRequest is the shared request body for POST /v1/demo/execute.
type ExecuteRequest struct {
	Function string                 `json:"function"`
	Input    map[string]interface{} `json:"input"`
}

// ExecuteResponse is the shared response body.
type ExecuteResponse struct {
	Success    bool        `json:"success"`
	Output     interface{} `json:"output,omitempty"`
	Error      string      `json:"error,omitempty"`
	Function   string      `json:"function"`
	LatencyMs  int64       `json:"latency_ms"`
	Timestamp  string      `json:"timestamp"`
	RateLimit  *RateLimitInfo `json:"rate_limit,omitempty"`
}

// RateLimitInfo echoes back limit state (useful for UI CTA timing).
type RateLimitInfo struct {
	Remaining int   `json:"remaining"`
	Limit     int   `json:"limit"`
	ResetUnix int64 `json:"reset_unix"`
}

// --- in-memory fallback for rate limit (when Redis is unavailable) ---
type ipWindow struct {
	count    int
	window   *time.Ticker
	resetAt  time.Time
}

type memLimiter struct {
	ips   map[string]*ipWindow
	limit int
	win   time.Duration
	mu    sync.Mutex
}

func newMemLimiter(limit int, win time.Duration) *memLimiter {
	m := &memLimiter{ips: make(map[string]*ipWindow), limit: limit, win: win}
	go m.gc()
	return m
}

func (m *memLimiter) allow(ip string) (remaining int, resetUnix int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	w, ok := m.ips[ip]
	if !ok || now.After(w.resetAt) {
		w = &ipWindow{count: 0, resetAt: now.Add(m.win)}
		w.window = time.NewTicker(m.win)
		m.ips[ip] = w
	}
	if w.count >= m.limit {
		return 0, w.resetAt.Unix()
	}
	w.count++
	return m.limit - w.count, w.resetAt.Unix()
}

func (m *memLimiter) gc() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		m.mu.Lock()
		for ip, w := range m.ips {
			if time.Now().After(w.resetAt) {
				w.window.Stop()
				delete(m.ips, ip)
			}
		}
		m.mu.Unlock()
	}
}

// --- Golden Demo functions (pure Go, no external calls) ---

var (
	demoScraper   = demoWebScraper
	demoSummarizer = demoTextSummarizer
	demoCurrency   = demoCurrencyConverter
)

func demoWebScraper(input map[string]interface{}) (interface{}, error) {
	urlRaw, ok := input["url"]
	if !ok {
		return nil, fmt.Errorf("'url' is required")
	}
	url, ok := urlRaw.(string)
	if !ok {
		return nil, fmt.Errorf("'url' must be a string")
	}
	u := strings.ToLower(url)
	title := "Example Domain"
	description := "Example domain for illustration."
	if strings.Contains(u, "functionfly") {
		title = "FunctionFly — Ship Functions. Agents Trust Them."
		description = "The platform for AI-accessible functions. Publish in Python, Go, Rust, JavaScript, or WebAssembly — with sandboxed execution and trust scoring."
	} else if strings.Contains(u, "openai") {
		title = "OpenAI"
		description = "OpenAI's official website. AI research and deployment company."
	} else if strings.Contains(u, "github") {
		title = "GitHub"
		description = "Where the world builds software. Millions of developers and companies build, ship, and maintain their software on GitHub."
	} else {
		host := url
		if idx := strings.Index(url, "://"); idx != -1 {
			host = url[idx+3:]
		}
		if idx := strings.Index(host, "/"); idx != -1 {
			host = host[:idx]
		}
		title = host + " — Live Page"
		description = fmt.Sprintf("Scraped content summary for %s (demo mode — real scraper available after signup).", host)
	}
	return map[string]interface{}{"url": url, "title": title, "description": description, "status_code": 200, "cached": false}, nil
}

func demoTextSummarizer(input map[string]interface{}) (interface{}, error) {
	textRaw, ok := input["text"]
	if !ok {
		return nil, fmt.Errorf("'text' is required")
	}
	text, ok := textRaw.(string)
	if !ok {
		return nil, fmt.Errorf("'text' must be a string")
	}
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("'text' must not be empty")
	}
	words := strings.Fields(text)
	if len(words) <= 20 {
		return map[string]interface{}{"summary": strings.TrimSpace(text), "original_word_count": len(words), "summary_word_count": len(words), "compression_ratio": 1.0}, nil
	}
	sentences := splitSentences(text)
	var picked []string
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		switch {
		case strings.HasPrefix(s, "In conclusion"):
			picked = append(picked, s)
		case strings.HasPrefix(s, "This "):
			picked = append(picked, s)
		case strings.HasPrefix(s, "The "):
			picked = append(picked, s)
		case strings.HasPrefix(s, "Overall"):
			picked = append(picked, s)
		}
	}
	if len(picked) < 1 {
		picked = append(picked, sentences[0])
	}
	if len(picked) < 2 && len(sentences) > 1 {
		picked = append(picked, sentences[len(sentences)-1])
	}
	summary := strings.Join(picked, " ")
	if len(summary) > 300 {
		summary = summary[:297] + "..."
	}
	return map[string]interface{}{"summary": summary, "original_word_count": len(words), "summary_word_count": len(strings.Fields(summary)), "compression_ratio": math(float64(len(strings.Fields(summary))) / float64(len(words)))}, nil
}

func demoCurrencyConverter(input map[string]interface{}) (interface{}, error) {
	amountRaw, ok := input["amount"]
	if !ok {
		return nil, fmt.Errorf("'amount' is required")
	}
	amount, ok := amountRaw.(float64)
	if !ok {
		return nil, fmt.Errorf("'amount' must be a number")
	}
	if amount <= 0 {
		return nil, fmt.Errorf("'amount' must be greater than zero")
	}
	from := strings.ToUpper(strings.TrimSpace(getString(input, "from_currency", "USD")))
	to := strings.ToUpper(strings.TrimSpace(getString(input, "to_currency", "EUR")))
	if len(from) != 3 || len(to) != 3 {
		return nil, fmt.Errorf("currency codes must be ISO-4217 (3 uppercase letters)")
	}
	rates := map[string]float64{
		"USD": 1.0, "EUR": 0.92, "GBP": 0.79, "JPY": 149.5,
		"CAD": 1.36, "AUD": 1.53, "CHF": 0.88, "CNY": 7.23,
		"INR": 83.12, "BRL": 4.97, "KRW": 1338.0, "MXN": 17.08,
	}
	rFrom, ok := rates[from]
	if !ok {
		return nil, fmt.Errorf("unsupported from_currency: %s (supported: USD, EUR, GBP, JPY, CAD, AUD, CHF, CNY, INR, BRL, KRW, MXN)", from)
	}
	rTo, ok := rates[to]
	if !ok {
		return nil, fmt.Errorf("unsupported to_currency: %s (supported: USD, EUR, GBP, JPY, CAD, AUD, CHF, CNY, INR, BRL, KRW, MXN)", to)
	}
	converted := amount * (rTo / rFrom)
	return map[string]interface{}{
		"amount":         amount,
		"from_currency":  from,
		"to_currency":    to,
		"rate":           math(rTo / rFrom),
		"converted":      math(converted),
		"rates_source":   "demo (live rates available after signup)",
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// --- helpers ---

func getString(input map[string]interface{}, key, fallback string) string {
	if v, ok := input[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return fallback
}

func math(v float64) float64 {
	return v
}

func splitSentences(text string) []string {
	text = strings.ReplaceAll(text, "! ", "!|")
	text = strings.ReplaceAll(text, "? ", "?|")
	text = strings.ReplaceAll(text, ". ", ".|")
	return strings.Split(text, "|")
}

// --- Rate limiting ---

var demoRL = newMemLimiter(10, 24*time.Hour)

func (h *Handler) checkRateLimit(ip string) *RateLimitInfo {
	if h.redisClient != nil {
		key := fmt.Sprintf("demo:rl:%s", ip)
		ctx := context.Background()
		countCmd := h.redisClient.Get(ctx, key)
		if countCmd.Err() == nil {
			var rec struct {
				Count    int   `json:"count"`
				ResetUnix int64 `json:"reset_unix"`
			}
			data, _ := countCmd.Result()
			json.Unmarshal([]byte(data), &rec)
			if rec.Count >= 10 {
				return &RateLimitInfo{Remaining: 0, Limit: 10, ResetUnix: rec.ResetUnix}
			}
		}
		var rec struct {
			Count    int   `json:"count"`
			ResetUnix int64 `json:"reset_unix"`
		}
		rec.Count = 1
		rec.ResetUnix = time.Now().Add(24 * time.Hour).Unix()
		data, _ := json.Marshal(rec)
		h.redisClient.Set(ctx, key, data, 24*time.Hour)
		remaining := 10 - rec.Count
		return &RateLimitInfo{Remaining: remaining, Limit: 10, ResetUnix: rec.ResetUnix}
	}
	remaining, resetUnix := demoRL.allow(ip)
	return &RateLimitInfo{Remaining: remaining, Limit: 10, ResetUnix: resetUnix}
}

// --- Logging ---

func logDemoExecution(function, ip, userAgent string, success bool, latencyMs int64) {
	logrus.WithFields(logrus.Fields{
		"event":        "demo_execution",
		"function":     function,
		"ip":           ip,
		"user_agent":   userAgent,
		"success":      success,
		"latency_ms":   latencyMs,
	}).Info("demo_function_executed")
}

// --- Route setup ---

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/", h.ListFunctions).Methods("GET", "OPTIONS")
	r.HandleFunc("/execute", h.HandleExecute).Methods("POST", "OPTIONS")
}

// HandleExecute is the main entry point for running a demo function.
func (h *Handler) HandleExecute(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	w.Header().Set("Content-Type", "application/json")

	ip := getClientIP(r)

	rl := h.checkRateLimit(ip)
	if rl.Remaining <= 0 {
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.Limit))
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", rl.ResetUnix))
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(ExecuteResponse{Error: "rate_limit_exceeded", Function: "", LatencyMs: time.Since(start).Milliseconds()})
		return
	}

	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ExecuteResponse{Success: false, Error: "invalid JSON body: " + err.Error(), Function: "", LatencyMs: time.Since(start).Milliseconds()})
		return
	}

	req.Function = strings.ToLower(strings.TrimSpace(req.Function))
	if req.Input == nil {
		req.Input = map[string]interface{}{}
	}

	var output interface{}
	var err error

	switch req.Function {
	case "web-scraper":
		output, err = demoScraper(req.Input)
	case "text-summarizer":
		output, err = demoTextSummarizer(req.Input)
	case "currency-converter":
		output, err = demoCurrencyConverter(req.Input)
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ExecuteResponse{
			Success: false,
			Error:   fmt.Sprintf("unknown function %q; supported: web-scraper, text-summarizer, currency-converter", req.Function),
			Function: req.Function,
			LatencyMs: time.Since(start).Milliseconds(),
		})
		return
	}

	latency := time.Since(start).Milliseconds()
	success := err == nil

	logDemoExecution(req.Function, ip, r.UserAgent(), success, latency)

	rl = h.checkRateLimit(ip)
	resp := ExecuteResponse{
		Success:   success,
		Function:  req.Function,
		LatencyMs: latency,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RateLimit: rl,
	}
	if !success {
		resp.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
	} else {
		resp.Output = output
	}
	json.NewEncoder(w).Encode(resp)
}

// getClientIP mirrors the project-wide helper without importing it.
func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
