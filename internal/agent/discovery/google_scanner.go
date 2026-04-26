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

type GoogleScanner struct {
	Query       string
	Queries     []string // Multiple queries when GOOGLE_TRENDS_QUERIES is set
	MaxResults  int
	ApiKey      string
	Client      *http.Client
	BaseURL     string
	RateLimit   *GoogleRateLimiter
}

type GoogleRateLimiter struct {
	mu             sync.Mutex
	remaining      int
	limit          int
	resetTime      time.Time
	windowStart    time.Time
	windowRequests int
}

type GoogleTrendsResponse struct {
	Default struct {
		TrendingSearches []struct {
			Title             struct {
				Query string `json:"query"`
			} `json:"title"`
			SummaryText      string `json:"summaryText"`
			Traffic          string `json:"traffic"`
			RelatedQueries   []struct {
				Query string `json:"query"`
			} `json:"relatedQueries"`
			Image struct {
				NewsUrl   string `json:"newsUrl"`
				Source    string `json:"source"`
						} `json:"image"`
			} `json:"trendingSearches"`
		} `json:"default"`
	}

type GoogleSearchResult struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Snippet     string `json:"snippet"`
}

type GoogleSearchResponse struct {
	Items []GoogleSearchResult `json:"items"`
}

func NewGoogleScanner() *GoogleScanner {
	queriesEnv := os.Getenv("GOOGLE_TRENDS_QUERIES")
	query := getEnvOrDefault("GOOGLE_TRENDS_QUERY", "automation API integration webhook")
	maxResults := 50
	if v := os.Getenv("DISCOVERY_GOOGLE_MAX_RESULTS"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			maxResults = parsed
		}
	}

	var queries []string
	if queriesEnv != "" {
		queries = strings.Split(queriesEnv, ",")
		for i := range queries {
			queries[i] = strings.TrimSpace(queries[i])
		}
	} else {
		queries = []string{query}
	}

	return &GoogleScanner{
		Query:      query,
		Queries:    queries,
		MaxResults: maxResults,
		ApiKey:     os.Getenv("GOOGLE_SEARCH_API_KEY"),
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		BaseURL:   "https://www.googleapis.com/customsearch/v1",
		RateLimit: NewGoogleRateLimiter(),
	}
}

func NewGoogleRateLimiter() *GoogleRateLimiter {
	return &GoogleRateLimiter{
		mu:             sync.Mutex{},
		remaining:      100,
		limit:          100,
		resetTime:      time.Now().Add(time.Minute),
		windowStart:    time.Now(),
		windowRequests: 0,
	}
}

func (g *GoogleRateLimiter) Allow() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := time.Now()
	if now.Sub(g.windowStart) > time.Minute {
		g.windowStart = now
		g.windowRequests = 0
	}
	if g.remaining <= 0 {
		if now.After(g.resetTime) {
			g.remaining = g.limit
			return true
		}
		return false
	}
	if g.windowRequests >= 50 {
		return false
	}
	g.remaining--
	g.windowRequests++
	return true
}

func (g GoogleScanner) Name() string {
	if len(g.Queries) > 1 {
		return "google:trends:multi"
	}
	return "google:trends"
}

func (g GoogleScanner) Scan(ctx context.Context) ([]OpportunityCandidate, error) {
	if !g.RateLimit.Allow() {
		return []OpportunityCandidate{}, fmt.Errorf("Google API rate limit exceeded")
	}

	candidates := make([]OpportunityCandidate, 0)
	apiKey := os.Getenv("GOOGLE_SEARCH_API_KEY")
	searchEngineID := os.Getenv("GOOGLE_SEARCH_ENGINE_ID")

	if apiKey == "" || searchEngineID == "" {
		return []OpportunityCandidate{}, nil
	}

	// Fetch from all configured queries
	for _, q := range g.Queries {
		trendsCandidates, err := g.fetchFromQuery(ctx, apiKey, searchEngineID, q)
		if err == nil && len(trendsCandidates) > 0 {
			candidates = append(candidates, trendsCandidates...)
		}
		if len(candidates) >= g.MaxResults {
			break
		}
	}

	return candidates, nil
}

func (g GoogleScanner) fetchFromQuery(ctx context.Context, apiKey, searchEngineID, query string) ([]OpportunityCandidate, error) {
	searchQueries := []string{
		"automation workflow API integration",
		"webhook automation no-code",
		"API data transformation pipeline",
		"scheduled task automation",
		"business process automation API",
	}

	if query != "" {
		searchQueries = []string{query}
	}

	results := make([]OpportunityCandidate, 0, g.MaxResults)

	for _, q := range searchQueries {
		params := url.Values{}
		params.Set("key", apiKey)
		params.Set("cx", searchEngineID)
		params.Set("q", q)
		params.Set("num", strconv.Itoa(min(g.MaxResults/len(searchQueries), 10)))
		params.Set("sort", "date")

		req, err := http.NewRequestWithContext(ctx, "GET", g.BaseURL+"?"+params.Encode(), nil)
		if err != nil {
			continue
		}
		req.Header.Set("Accept", "application/json")

		resp, err := g.Client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		var searchResp GoogleSearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
			continue
		}

		for _, item := range searchResp.Items {
			text := strings.ToLower(item.Title + " " + item.Snippet)
			if !isAutomationOpportunity(text) {
				continue
			}

			demandSignal := calculateDemandFromGoogle(item)

			candidate := OpportunityCandidate{
				Source:           g.Name(),
				SourceID:         fmt.Sprintf("google_%d", hashStrings(item.Link)),
				Title:            cleanTitle(item.Title),
				Description:      item.Snippet,
				Category:         "google_trends",
				Tags:             uniqueStrings([]string{"google", "trends", extractTopic(item.Title)}),
				DemandSignal:     demandSignal,
				ComplexitySignal: estimateComplexityFromText(text),
				Metadata: map[string]any{
					"link":   item.Link,
					"source": "google_search",
					"query":  q,
				},
				DiscoveredAt: time.Now().UTC(),
			}
			results = append(results, candidate)
		}

		if len(results) >= g.MaxResults {
			break
		}

		time.Sleep(time.Second)
	}

	return results, nil
}

func isAutomationOpportunity(text string) bool {
	indicators := []string{
		"automate", "automation", "integration", "webhook", "api",
		"workflow", "pipeline", "schedule", "cron", "sync",
		"convert", "transform", "import", "export", "sync",
		"notify", "alert", "report", "dashboard", "no-code",
		"i need", "looking for", "tool to", "how to", "best way to",
		"manual", "tedious", "repetitive", "time consuming",
	}
	for _, ind := range indicators {
		if strings.Contains(text, ind) {
			return true
		}
	}
	return false
}

func calculateDemandFromGoogle(item GoogleSearchResult) float64 {
	signal := float64(20)
	text := strings.ToLower(item.Title + " " + item.Snippet)

	highValue := []string{"enterprise", "business", "production", "scale", "million"}
	for _, kw := range highValue {
		if strings.Contains(text, kw) {
			signal += 15
		}
	}

	urgent := []string{"urgent", "critical", "broken", "failing", "error"}
	for _, kw := range urgent {
		if strings.Contains(text, kw) {
			signal += 10
		}
	}

	popular := []string{"popular", "trending", "viral", "top"}
	for _, kw := range popular {
		if strings.Contains(text, kw) {
			signal += 10
		}
	}

	if signal > 100 {
		signal = 100
	}
	return signal
}

func estimateComplexityFromText(text string) int {
	complex := []string{"machine learning", "ai", "nlp", "image processing", "video", "real-time", "streaming", "distributed"}
	simple := []string{"simple", "basic", "one step", "single", "easy"}

	score := 5
	for _, kw := range complex {
		if strings.Contains(text, kw) {
			score++
		}
	}
	for _, kw := range simple {
		if strings.Contains(text, kw) {
			score--
		}
	}

	if score < 1 {
		score = 1
	}
	if score > 10 {
		score = 10
	}
	return score
}

func cleanTitle(title string) string {
	title = strings.TrimSpace(title)
	title = strings.ReplaceAll(title, "|", "-")
	title = strings.ReplaceAll(title, "  ", " ")
	return title
}

func extractTopic(title string) string {
	title = strings.ToLower(title)
	topics := []string{"api", "webhook", "automation", "integration", "workflow", "data", "sync", "import", "export"}
	for _, t := range topics {
		if strings.Contains(title, t) {
			return t
		}
	}
	return "general"
}

func hashStrings(s string) int {
	h := 0
	for _, c := range s {
		h = h*31 + int(c)
	}
	if h < 0 {
		h = -h
	}
	return h
}
