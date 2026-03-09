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

// RedditScanner is a scanner that fetches posts from Reddit subreddits.
// It uses Reddit's public JSON API to discover posts that indicate pain points
// or feature requests that could be automated with functions.
type RedditScanner struct {
	// Configuration
	Subreddit  string
	MaxResults int

	// Internal state
	client    *http.Client
	baseURL   string
	rateLimit *RedditRateLimiter
}

// RedditRateLimiter implements rate limiting for Reddit's API.
// Reddit allows ~60 requests per minute for public access.
type RedditRateLimiter struct {
	mu             sync.Mutex
	remaining      int
	limit          int
	resetTime      time.Time
	requestsMade   int
	windowStart    time.Time
	windowRequests int
}

// RedditPost represents a post from the Reddit API
type RedditPost struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Body         string  `json:"selftext"`
	Score        int     `json:"score"`
	CommentCount int     `json:"num_comments"`
	Permalink    string  `json:"permalink"`
	Flair        string  `json:"link_flair_text"`
	CreatedUTC   float64 `json:"created_utc"`
	Subreddit    string  `json:"subreddit"`
	URL          string  `json:"url"`
	IsSelf       bool    `json:"is_self"`
}

// RedditAPIResponse represents the response from Reddit's JSON API
type RedditAPIResponse struct {
	Data struct {
		Children []struct {
			Data RedditPost `json:"data"`
		} `json:"children"`
		After string `json:"after"`
	} `json:"data"`
}

// NewRedditScanner creates a new Reddit scanner with configuration from environment variables.
// It reads REDDIT_SUBREDDIT from the environment.
func NewRedditScanner() *RedditScanner {
	subreddit := getEnvOrDefault("REDDIT_SUBREDDIT", "learnprogramming")

	maxResults := 50
	if v := os.Getenv("DISCOVERY_REDDIT_MAX_RESULTS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			maxResults = parsed
		}
	}

	return &RedditScanner{
		Subreddit:  subreddit,
		MaxResults: maxResults,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   "https://www.reddit.com",
		rateLimit: NewRedditRateLimiter(),
	}
}

// NewRedditScannerWithConfig creates a new Reddit scanner with explicit configuration.
func NewRedditScannerWithConfig(subreddit string, maxResults int) *RedditScanner {
	if subreddit == "" {
		subreddit = "learnprogramming"
	}
	if maxResults <= 0 {
		maxResults = 50
	}

	return &RedditScanner{
		Subreddit:  subreddit,
		MaxResults: maxResults,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   "https://www.reddit.com",
		rateLimit: NewRedditRateLimiter(),
	}
}

// NewRedditRateLimiter creates a new rate limiter for Reddit API
func NewRedditRateLimiter() *RedditRateLimiter {
	return &RedditRateLimiter{
		mu:             sync.Mutex{},
		remaining:      60,
		limit:          60,
		resetTime:      time.Now().Add(time.Minute),
		windowStart:    time.Now(),
		windowRequests: 0,
	}
}

// Allow checks if we can make a request without exceeding rate limits
func (rl *RedditRateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Reset window if needed (rolling window for additional limiting)
	if now.Sub(rl.windowStart) > time.Minute {
		rl.windowStart = now
		rl.windowRequests = 0
	}

	// Primary check: our remaining quota
	if rl.remaining <= 0 {
		// Check if we've passed the reset time
		if now.After(rl.resetTime) {
			rl.remaining = rl.limit
			return true
		}
		return false
	}

	// Secondary check: our rolling window (max 30 requests per minute to be safe)
	if rl.windowRequests >= 30 {
		return false
	}

	rl.remaining--
	rl.windowRequests++
	return true
}

// WaitUntilReady waits until rate limit allows making a request
func (rl *RedditRateLimiter) WaitUntilReady() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for !rl.Allow() {
		rl.mu.Unlock()
		time.Sleep(10 * time.Second)
		rl.mu.Lock()
	}
}

// WaitIfNeeded waits if we're close to rate limit
func (rl *RedditRateLimiter) WaitIfNeeded() {
	rl.mu.Lock()
	remaining := rl.remaining
	rl.mu.Unlock()

	// If we're running low, wait a bit to space out requests
	if remaining < 10 {
		time.Sleep(time.Second)
	}
}

// Name returns the scanner name
func (r RedditScanner) Name() string {
	if r.Subreddit == "" {
		return "reddit"
	}
	return fmt.Sprintf("reddit:%s", r.Subreddit)
}

// Scan fetches posts from Reddit and converts them to opportunity candidates.
func (r RedditScanner) Scan(ctx context.Context) ([]OpportunityCandidate, error) {
	// Check if we have valid configuration
	if r.Subreddit == "" {
		// Return empty results if no subreddit configured - don't error
		return []OpportunityCandidate{}, nil
	}

	// Wait for rate limit if needed
	if !r.rateLimit.Allow() {
		return []OpportunityCandidate{}, fmt.Errorf("Reddit API rate limit exceeded, try again later")
	}

	// Fetch posts from the subreddit
	posts, err := r.fetchPosts(ctx, r.Subreddit, r.MaxResults)
	if err != nil {
		// Don't return error, just log and return empty
		// This ensures the pipeline continues even if Reddit API fails
		return []OpportunityCandidate{}, nil
	}

	// Convert to candidates
	results := make([]OpportunityCandidate, 0, len(posts))
	for _, post := range posts {
		// Skip posts that don't indicate a pain point or opportunity
		text := strings.ToLower(post.Title + " " + post.Body)
		isOpportunity := r.isOpportunityPost(post, text)

		if !isOpportunity {
			continue
		}

		// Calculate demand signal
		demandSignal := r.calculateDemandSignal(post)

		candidate := OpportunityCandidate{
			Source:           r.Name(),
			SourceID:         post.ID,
			Title:            post.Title,
			Description:      post.Body,
			Category:         "community_pain_point",
			Tags:             uniqueStrings([]string{"reddit", r.Subreddit, post.Flair}),
			DemandSignal:     demandSignal,
			ComplexitySignal: estimateComplexity(post.Title+" "+post.Body, []string{post.Flair, r.Subreddit}),
			Metadata: map[string]any{
				"permalink":     "https://reddit.com" + post.Permalink,
				"score":         post.Score,
				"comment_count": post.CommentCount,
				"flair":         post.Flair,
				"created_utc":   post.CreatedUTC,
				"is_self":       post.IsSelf,
			},
			DiscoveredAt: time.Now().UTC(),
		}
		results = append(results, candidate)
	}

	return results, nil
}

// isOpportunityPost determines if a post indicates a pain point or opportunity
func (r RedditScanner) isOpportunityPost(post RedditPost, text string) bool {
	// Look for keywords that indicate someone needs help or wants automation
	indicators := []string{
		"need help",
		"looking for",
		"how do i",
		"how can i",
		"wish there was",
		"anyone know",
		"can someone",
		"trying to",
		"need to",
		"automate",
		"is there a way",
		"any tool",
		"recommend",
		"best way",
		"struggling with",
		" frustrated",
		" pain point",
		"annoying",
		"tedious",
		"repetitive",
		"manual process",
	}

	for _, indicator := range indicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}

	// High engagement posts are also worth considering
	if post.Score >= 50 || post.CommentCount >= 20 {
		return true
	}

	return false
}

// fetchPosts fetches posts from a subreddit
func (r RedditScanner) fetchPosts(ctx context.Context, subreddit string, maxResults int) ([]RedditPost, error) {
	// Build the API URL - use .json endpoint for public access
	params := url.Values{}
	params.Set("limit", strconv.Itoa(min(maxResults, 100)))
	params.Set("sort", "hot") // Get hot posts for popular topics

	url := fmt.Sprintf("%s/r/%s/new.json?%s", r.baseURL, subreddit, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers - Reddit requires a User-Agent
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "FunctionFly/1.0 (AI Function Factory; https://functionfly.com)")

	// Make request
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("Reddit API rate limited or access forbidden")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("subreddit not found: %s", subreddit)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Reddit API returned status %d", resp.StatusCode)
	}

	// Parse response
	var apiResp RedditAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract posts from response
	posts := make([]RedditPost, 0, len(apiResp.Data.Children))
	for _, child := range apiResp.Data.Children {
		posts = append(posts, child.Data)
	}

	// Update rate limit
	r.rateLimit.WaitIfNeeded()

	return posts, nil
}

// calculateDemandSignal calculates a demand score based on post metrics
func (r RedditScanner) calculateDemandSignal(post RedditPost) float64 {
	signal := float64(15) // Base score

	// Score indicates community validation
	signal += float64(min(post.Score, 50))

	// Comments indicate active discussion
	signal += float64(min(post.CommentCount/2, 25))

	// Check for urgency indicators in text
	text := strings.ToLower(post.Title + " " + post.Body)
	urgentKeywords := []string{"urgent", "asap", "emergency", "critical", "broken", "help!"}
	for _, keyword := range urgentKeywords {
		if strings.Contains(text, keyword) {
			signal += 10
			break
		}
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
