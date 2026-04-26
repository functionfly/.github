package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// StackOverflowScanner is a scanner that fetches questions from StackOverflow.
// It uses the Stack Exchange API to discover unanswered questions with specific tags.
// When multiple tags are configured via STACKOVERFLOW_TAGS, scans each one.
type StackOverflowScanner struct {
	// Configuration
	Tag        string
	Tags       []string // Multiple tags when STACKOVERFLOW_TAGS is set
	MaxResults int

	// Internal state
	client    *http.Client
	baseURL   string
	apiKey    string
	rateLimit *StackOverflowRateLimiter
}

// StackOverflowRateLimiter implements rate limiting for StackExchange API.
// The API allows 300 requests per day for unauthenticated requests.
type StackOverflowRateLimiter struct {
	mu             sync.Mutex
	remaining      int
	limit          int
	resetTime      time.Time
	requestsMade   int
	windowStart    time.Time
	windowRequests int
}

// StackOverflowQuestion represents a question from the StackOverflow API
type StackOverflowQuestion struct {
	QuestionID   int      `json:"question_id"`
	Title        string   `json:"title"`
	Body         string   `json:"body_markdown"`
	Tags         []string `json:"tags"`
	ViewCount    int      `json:"view_count"`
	AnswerCount  int      `json:"answer_count"`
	Score        int      `json:"score"`
	Link         string   `json:"link"`
	IsAnswered   bool     `json:"is_answered"`
	CreationDate int64    `json:"creation_date"`
	LastActivity int64    `json:"last_activity_date"`
}

// StackOverflowAPIResponse represents the response from StackExchange API
type StackOverflowAPIResponse struct {
	Items          []StackOverflowQuestion `json:"items"`
	HasMore        bool                    `json:"has_more"`
	QuotaMax       int                     `json:"quota_max"`
	QuotaRemaining int                     `json:"quota_remaining"`
}

// NewStackOverflowScanner creates a new StackOverflow scanner with configuration from environment variables.
// It reads STACKOVERFLOW_TAG or STACKOVERFLOW_TAGS (comma-separated) from the environment.
func NewStackOverflowScanner() *StackOverflowScanner {
	tagsEnv := os.Getenv("STACKOVERFLOW_TAGS")
	tag := getEnvOrDefault("STACKOVERFLOW_TAG", "python")
	apiKey := os.Getenv("STACKOVERFLOW_API_KEY")

	maxResults := 50
	if v := os.Getenv("DISCOVERY_STACKOVERFLOW_MAX_RESULTS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			maxResults = parsed
		}
	}

	var tags []string
	if tagsEnv != "" {
		tags = strings.Split(tagsEnv, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	} else if tag != "" {
		tags = []string{tag}
	}

	return &StackOverflowScanner{
		Tag:        tag,
		Tags:       tags,
		MaxResults: maxResults,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   "https://api.stackexchange.com/2.3",
		apiKey:    apiKey,
		rateLimit: NewStackOverflowRateLimiter(),
	}
}

// NewStackOverflowScannerWithConfig creates a new StackOverflow scanner with explicit configuration.
func NewStackOverflowScannerWithConfig(tag string, maxResults int) *StackOverflowScanner {
	if tag == "" {
		tag = "python"
	}
	if maxResults <= 0 {
		maxResults = 50
	}

	return &StackOverflowScanner{
		Tag:        tag,
		MaxResults: maxResults,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   "https://api.stackexchange.com/2.3",
		rateLimit: NewStackOverflowRateLimiter(),
	}
}

// NewStackOverflowRateLimiter creates a new rate limiter for StackOverflow API
func NewStackOverflowRateLimiter() *StackOverflowRateLimiter {
	return &StackOverflowRateLimiter{
		mu:             sync.Mutex{},
		remaining:      300,
		limit:          300,
		resetTime:      time.Now().Add(24 * time.Hour),
		windowStart:    time.Now(),
		windowRequests: 0,
	}
}

// Allow checks if we can make a request without exceeding rate limits
func (rl *StackOverflowRateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Reset window if needed (rolling window for additional limiting)
	if now.Sub(rl.windowStart) > time.Minute*5 {
		rl.windowStart = now
		rl.windowRequests = 0
	}

	// Primary check: StackOverflow's remaining quota
	if rl.remaining <= 0 {
		// Check if we've passed the reset time
		if now.After(rl.resetTime) {
			rl.remaining = rl.limit
			return true
		}
		return false
	}

	// Secondary check: our rolling window (max 30 requests per 5 minutes to be safe)
	if rl.windowRequests >= 30 {
		return false
	}

	rl.remaining--
	rl.windowRequests++
	return true
}

// WaitUntilReady waits until rate limit allows making a request
func (rl *StackOverflowRateLimiter) WaitUntilReady() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for !rl.Allow() {
		rl.mu.Unlock()
		time.Sleep(10 * time.Second)
		rl.mu.Lock()
	}
}

// UpdateFromResponse updates rate limit info from StackOverflow response
func (rl *StackOverflowRateLimiter) UpdateFromResponse(resp *StackOverflowAPIResponse) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.limit = resp.QuotaMax
	rl.remaining = resp.QuotaRemaining
}

// WaitIfNeeded waits if we're close to rate limit
func (rl *StackOverflowRateLimiter) WaitIfNeeded() {
	rl.mu.Lock()
	remaining := rl.remaining
	rl.mu.Unlock()

	// If we're running low, wait a bit to space out requests
	if remaining < 10 {
		time.Sleep(time.Second)
	}
}

// Name returns the scanner name
func (s StackOverflowScanner) Name() string {
	if len(s.Tags) > 1 {
		return fmt.Sprintf("stackoverflow:multi:%d", len(s.Tags))
	}
	if s.Tag == "" {
		return "stackoverflow"
	}
	return fmt.Sprintf("stackoverflow:%s", s.Tag)
}

// Scan fetches questions from StackOverflow and converts them to opportunity candidates.
// When multiple tags are configured, scans each one and aggregates results.
func (s StackOverflowScanner) Scan(ctx context.Context) ([]OpportunityCandidate, error) {
	// Determine which tags to scan
	var tags []string
	if len(s.Tags) > 1 {
		tags = s.Tags
	} else if s.Tag != "" {
		tags = []string{s.Tag}
	} else {
		// Return empty results if no tag configured - don't error
		return []OpportunityCandidate{}, nil
	}

	// Wait for rate limit if needed
	if !s.rateLimit.Allow() {
		return []OpportunityCandidate{}, fmt.Errorf("StackOverflow API rate limit exceeded, try again later")
	}

	// Search for unanswered questions with each tag
	results := make([]OpportunityCandidate, 0)
	for _, tag := range tags {
		questions, err := s.searchQuestions(ctx, tag, s.MaxResults/len(tags)+1)
		if err != nil {
			// Don't fail entirely - continue with other tags
			continue
		}

		for _, question := range questions {
			// Skip answered questions - we want to find problems people need help with
			if question.IsAnswered && question.Score < 10 {
				continue
			}

			// Calculate demand signal based on views, score, and lack of answers
			demandSignal := s.calculateDemandSignal(question)

			candidate := OpportunityCandidate{
				Source:           s.Name(),
				SourceID:         fmt.Sprintf("%d", question.QuestionID),
				Title:            question.Title,
				Description:      question.Body,
				Category:         "developer_problem",
				Tags:             append([]string{"stackoverflow"}, question.Tags...),
				DemandSignal:     demandSignal,
				ComplexitySignal: estimateComplexity(question.Title+" "+question.Body, question.Tags),
				Metadata: map[string]any{
					"link":          question.Link,
					"views":         question.ViewCount,
					"answer_count":  question.AnswerCount,
					"score":         question.Score,
					"is_answered":   question.IsAnswered,
					"creation_date": time.Unix(question.CreationDate, 0),
					"tag":           tag,
				},
				DiscoveredAt: time.Now().UTC(),
			}
			results = append(results, candidate)
		}
	}

	return results, nil
}

// searchQuestions searches StackOverflow for unanswered questions with the specified tag
func (s StackOverflowScanner) searchQuestions(ctx context.Context, tag string, maxResults int) ([]StackOverflowQuestion, error) {
	// Build the API URL
	// We filter for questions that are not answered or have low score
	params := url.Values{}
	params.Set("order", "desc")
	params.Set("sort", "activity")
	params.Set("tagged", tag)
	params.Set("site", "stackoverflow")
	params.Set("filter", "withbody") // Include body in response
	params.Set("pagesize", strconv.Itoa(min(maxResults, 100)))

	// Add API key if available
	if s.apiKey != "" {
		params.Set("key", s.apiKey)
	}

	url := fmt.Sprintf("%s/questions?%s", s.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "FunctionFly/1.0 (AI Function Factory)")

	// Make request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch questions: %w", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("StackOverflow API rate limited or access forbidden")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("StackOverflow API endpoint not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("StackOverflow API returned status %d", resp.StatusCode)
	}

	// Parse response
	var apiResp StackOverflowAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Update rate limit info
	s.rateLimit.UpdateFromResponse(&apiResp)
	s.rateLimit.WaitIfNeeded()

	return apiResp.Items, nil
}

// calculateDemandSignal calculates a demand score based on question metrics
func (s StackOverflowScanner) calculateDemandSignal(question StackOverflowQuestion) float64 {
	signal := float64(15) // Base score

	// View count indicates interest
	signal += float64(min(question.ViewCount/100, 30))

	// Score indicates question quality and interest
	signal += float64(question.Score * 3)

	// Fewer answers = more opportunity for help (if question is interesting)
	if question.AnswerCount == 0 {
		signal += 15 // Unanswered questions are high opportunity
	} else if question.AnswerCount < 2 && question.Score < 5 {
		signal += 5 // Low engagement with answers
	}

	// Cap the signal
	if signal > 100 {
		signal = 100
	}
	if signal < 0 {
		signal = 0
	}

	return signal
}
