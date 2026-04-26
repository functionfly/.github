package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// GitHubScanner is a scanner that fetches issues from GitHub repositories.
// It uses the GitHub REST API to discover enhancement requests, feature requests,
// and other opportunities for function automation.
type GitHubScanner struct {
	// Configuration
	Owner      string
	Repo       string
	Token      string
	Labels     []string
	MaxResults int

	// Internal state
	client    *http.Client
	baseURL   string
	rateLimit *GitHubRateLimiter
}

// GitHubRateLimiter implements rate limiting based on GitHub's rate limit headers.
// Authenticated requests get 5000 requests/hour, unauthenticated get 60/hour.
type GitHubRateLimiter struct {
	mu             sync.Mutex
	remaining      int
	limit          int
	resetTime      time.Time
	requestsMade   int
	windowStart    time.Time
	windowRequests int
}

// GitHubIssue represents an issue returned from the GitHub API
type GitHubIssue struct {
	ID      int    `json:"id"`
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
	State   string `json:"state"`
	Labels  []struct {
		Name string `json:"name"`
	} `json:"labels"`
	Comments    int       `json:"comments"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	PullRequest struct {
		URL string `json:"url"`
	} `json:"pull_request"`
}

// GitHubSearchResponse represents the response from GitHub search API
type GitHubSearchResponse struct {
	TotalCount        int           `json:"total_count"`
	IncompleteResults bool          `json:"incomplete_results"`
	Items             []GitHubIssue `json:"items"`
}

// NewGitHubScanner creates a new GitHub scanner with configuration from environment variables.
// It reads GITHUB_TOKEN, GITHUB_OWNER, and GITHUB_REPO from the environment.
// If no token is provided, it will work with unauthenticated rate limits (60/hour).
func NewGitHubScanner() *GitHubScanner {
	token := os.Getenv("GITHUB_TOKEN")
	owner := getEnvOrDefault("GITHUB_OWNER", "")
	repo := getEnvOrDefault("GITHUB_REPO", "")

	// Default labels to search for
	labels := []string{"enhancement", "feature-request", "help-wanted", "good first issue"}

	// Default max results
	maxResults := 100
	if v := os.Getenv("DISCOVERY_GITHUB_MAX_RESULTS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			maxResults = parsed
		}
	}

	return &GitHubScanner{
		Owner:      owner,
		Repo:       repo,
		Token:      token,
		Labels:     labels,
		MaxResults: maxResults,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   "https://api.github.com",
		rateLimit: NewGitHubRateLimiter(),
	}
}

// NewGitHubScannerWithConfig creates a new GitHub scanner with explicit configuration.
// This is useful for testing or when you want to scan specific repositories.
func NewGitHubScannerWithConfig(owner, repo, token string, labels []string, maxResults int) *GitHubScanner {
	if len(labels) == 0 {
		labels = []string{"enhancement", "feature-request", "help-wanted", "good first issue"}
	}
	if maxResults <= 0 {
		maxResults = 100
	}

	return &GitHubScanner{
		Owner:      owner,
		Repo:       repo,
		Token:      token,
		Labels:     labels,
		MaxResults: maxResults,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   "https://api.github.com",
		rateLimit: NewGitHubRateLimiter(),
	}
}

// NewGitHubGlobalScanner creates a scanner that searches GitHub globally for issues
// matching the specified labels across all public repositories.
// This is used when GITHUB_OWNER/GITHUB_REPO are not set but GITHUB_TOKEN is available.
func NewGitHubGlobalScanner(token string, labels []string, maxResults int) *GitHubScanner {
	if len(labels) == 0 {
		labels = []string{"enhancement", "feature-request", "help-wanted", "good first issue"}
	}
	if maxResults <= 0 {
		maxResults = 100
	}

	return &GitHubScanner{
		Owner:      "", // empty signals global search mode
		Repo:       "",
		Token:      token,
		Labels:     labels,
		MaxResults: maxResults,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL:   "https://api.github.com",
		rateLimit: NewGitHubRateLimiter(),
	}
}

// NewGitHubRateLimiter creates a new rate limiter for GitHub API
func NewGitHubRateLimiter() *GitHubRateLimiter {
	return &GitHubRateLimiter{
		mu:             sync.Mutex{},
		remaining:      5000, // Assume authenticated by default
		limit:          5000,
		resetTime:      time.Now().Add(time.Hour),
		windowStart:    time.Now(),
		windowRequests: 0,
	}
}

// Allow checks if we can make a request without exceeding rate limits
func (rl *GitHubRateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Reset window if needed (rolling window for additional limiting)
	if now.Sub(rl.windowStart) > time.Minute {
		rl.windowStart = now
		rl.windowRequests = 0
	}

	// Primary check: GitHub's remaining count
	if rl.remaining <= 0 {
		// Check if we've passed the reset time
		if now.After(rl.resetTime) {
			rl.remaining = rl.limit
			return true
		}
		return false
	}

	// Secondary check: our rolling window (max 100 requests per minute to be safe)
	if rl.windowRequests >= 100 {
		return false
	}

	rl.remaining--
	rl.windowRequests++
	return true
}

// WaitUntilReady waits until rate limit allows making a request
func (rl *GitHubRateLimiter) WaitUntilReady() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for !rl.Allow() {
		rl.mu.Unlock()
		time.Sleep(10 * time.Second)
		rl.mu.Lock()
	}
}

// UpdateFromResponse updates rate limit info from GitHub response headers
func (rl *GitHubRateLimiter) UpdateFromResponse(resp *http.Response) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if limit := resp.Header.Get("X-RateLimit-Limit"); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil {
			rl.limit = parsed
		}
	}

	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		if parsed, err := strconv.Atoi(remaining); err == nil {
			rl.remaining = parsed
		}
	}

	if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
		if parsed, err := strconv.ParseInt(reset, 10, 64); err == nil {
			rl.resetTime = time.Unix(parsed, 0)
		}
	}
}

// WaitIfNeeded waits if we're close to rate limit
func (rl *GitHubRateLimiter) WaitIfNeeded() {
	rl.mu.Lock()
	remaining := rl.remaining
	rl.mu.Unlock()

	// If we're running low, wait a bit to space out requests
	if remaining < 10 {
		time.Sleep(time.Second)
	}
}

// Name returns the scanner name
func (g GitHubScanner) Name() string {
	if g.Owner == "" && g.Repo == "" {
		return "github:global"
	}
	if g.Owner == "" || g.Repo == "" {
		return "github"
	}
	return fmt.Sprintf("github:%s/%s", g.Owner, g.Repo)
}

// Scan fetches issues from GitHub and converts them to opportunity candidates.
// It searches for issues with the configured labels (enhancement, feature-request, etc.)
// and returns them as candidates for function automation.
// When Owner/Repo are empty but Token is set, performs a global search across GitHub.
func (g GitHubScanner) Scan(ctx context.Context) ([]OpportunityCandidate, error) {
	// Check if we have valid configuration
	if g.Owner == "" && g.Repo == "" {
		// Global search mode - requires token for reasonable rate limits
		if g.Token == "" {
			// No token and no owner/repo - return empty silently
			return []OpportunityCandidate{}, nil
		}
		return g.scanGlobal(ctx)
	}

	if g.Owner == "" || g.Repo == "" {
		// Partial configuration - return empty
		return []OpportunityCandidate{}, nil
	}

	// Wait for rate limit if needed
	if !g.rateLimit.Allow() {
		return []OpportunityCandidate{}, fmt.Errorf("GitHub API rate limit exceeded, try again later")
	}

	// Build search query for issues
	labelQuery := strings.Join(g.Labels, ",")
	query := fmt.Sprintf("repo:%s/%s is:issue is:open label:%s",
		g.Owner, g.Repo, labelQuery)

	// Use search API to find issues
	issues, err := g.searchIssues(ctx, query, g.MaxResults)
	if err != nil {
		// Don't return error, just log and return empty
		// This ensures the pipeline continues even if GitHub API fails
		return []OpportunityCandidate{}, nil
	}

	return g.convertToCandidates(issues)
}

// scanGlobal searches GitHub globally for issues matching the configured labels.
func (g GitHubScanner) scanGlobal(ctx context.Context) ([]OpportunityCandidate, error) {
	if !g.rateLimit.Allow() {
		return []OpportunityCandidate{}, fmt.Errorf("GitHub API rate limit exceeded, try again later")
	}

	// Build global search query - search across all of GitHub for issues with our labels
	// GitHub search uses commas to match any of several labels (OR behavior)
	// Labels with spaces must be quoted
	labelParts := make([]string, len(g.Labels))
	for i, label := range g.Labels {
		if strings.Contains(label, " ") {
			labelParts[i] = fmt.Sprintf("label:%q", label)
		} else {
			labelParts[i] = "label:" + label
		}
	}
	labelQuery := strings.Join(labelParts, ",")
	query := fmt.Sprintf("is:issue is:open %s", labelQuery)

	issues, err := g.searchIssues(ctx, query, g.MaxResults)
	if err != nil {
		return []OpportunityCandidate{}, nil
	}

	return g.convertToCandidates(issues)
}

// convertToCandidates converts GitHub issues to opportunity candidates.
func (g GitHubScanner) convertToCandidates(issues []GitHubIssue) ([]OpportunityCandidate, error) {
	results := make([]OpportunityCandidate, 0, len(issues))
	for _, issue := range issues {
		// Skip pull requests
		if issue.PullRequest.URL != "" {
			continue
		}

		// Extract label names
		labelNames := make([]string, len(issue.Labels))
		for i, label := range issue.Labels {
			labelNames[i] = label.Name
		}

		// Calculate demand signal based on comments and recency
		demandSignal := g.calculateDemandSignal(issue)

		candidate := OpportunityCandidate{
			Source:           g.Name(),
			SourceID:         fmt.Sprintf("%d", issue.ID),
			Title:            issue.Title,
			Description:      issue.Body,
			Category:         "github_issue",
			Tags:             append([]string{"github"}, labelNames...),
			DemandSignal:     demandSignal,
			ComplexitySignal: estimateComplexity(issue.Title+" "+issue.Body, labelNames),
			Metadata: map[string]any{
				"url":        issue.HTMLURL,
				"number":     issue.Number,
				"state":      issue.State,
				"created_at": issue.CreatedAt,
				"updated_at": issue.UpdatedAt,
				"comments":   issue.Comments,
				"labels":     labelNames,
			},
			DiscoveredAt: time.Now().UTC(),
		}
		results = append(results, candidate)
	}

	return results, nil
}

// searchIssues searches GitHub for issues matching the query
func (g GitHubScanner) searchIssues(ctx context.Context, query string, maxResults int) ([]GitHubIssue, error) {
	// URL encode the query
	encodedQuery := strings.ReplaceAll(query, " ", "+")

	url := fmt.Sprintf("%s/search/issues?q=%s&per_page=%d&sort=created&order=desc",
		g.baseURL, encodedQuery, min(maxResults, 100))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if g.Token != "" {
		req.Header.Set("Authorization", "Bearer "+g.Token)
	}

	// Make request
	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limit info
	g.rateLimit.UpdateFromResponse(resp)
	g.rateLimit.WaitIfNeeded()

	// Check for errors
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("GitHub API rate limited or access forbidden")
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("repository not found or no access: %s/%s", g.Owner, g.Repo)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	// Parse response
	var searchResp GitHubSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return searchResp.Items, nil
}

// fetchReactions fetches reactions for an issue (optional, for more accurate demand signal)
func (g GitHubScanner) fetchReactions(ctx context.Context, issueNumber int) (int, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/reactions",
		g.baseURL, g.Owner, g.Repo, issueNumber)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if g.Token != "" {
		req.Header.Set("Authorization", "Bearer "+g.Token)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	g.rateLimit.UpdateFromResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return 0, nil // Don't fail, just return 0
	}

	var reactions []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&reactions); err != nil {
		return 0, nil
	}

	return len(reactions), nil
}

// calculateDemandSignal calculates a demand score based on issue metrics
func (g GitHubScanner) calculateDemandSignal(issue GitHubIssue) float64 {
	signal := float64(15) // Base score

	// Comments indicate active interest
	signal += float64(issue.Comments * 5)

	// Recency: more recent issues are more relevant
	daysOld := time.Since(issue.CreatedAt).Hours() / 24
	if daysOld < 7 {
		signal += 20
	} else if daysOld < 30 {
		signal += 10
	} else if daysOld > 365 {
		signal -= 10
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

// matchesAnyLabel checks if any of the issue's labels match the expected labels
func matchesAnyLabel(labels []string, expected map[string]struct{}) bool {
	for _, label := range labels {
		if _, ok := expected[strings.ToLower(strings.TrimSpace(label))]; ok {
			return true
		}
	}
	return false
}

// estimateComplexity estimates the complexity of implementing a function based on issue content
func estimateComplexity(text string, tags []string) int {
	complexity := 3
	lower := strings.ToLower(text + " " + strings.Join(tags, " "))

	keywords := map[string]int{
		"oauth": 9, "auth": 8, "async": 6, "stream": 7,
		"webhook": 8, "deployment": 7, "infra": 8,
		"security": 9, "database": 6, "api": 5,
		"integration": 8, "machine learning": 9,
		"ai": 8, "ml": 9, "blockchain": 10,
	}

	for keyword, complexityBonus := range keywords {
		if strings.Contains(lower, keyword) {
			complexity += complexityBonus / 3
		}
	}

	if complexity > 10 {
		complexity = 10
	}
	return complexity
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
