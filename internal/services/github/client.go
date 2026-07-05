package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	defaultBaseURL  = "https://api.github.com"
	defaultUserAgent = "FunctionFly/1.0"
	maxRetries      = 3
	retryBaseDelay  = time.Second
)

type RateLimiter struct {
	remaining int
	limit     int
	resetAt   time.Time
	mu        sync.Mutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{remaining: 5000, limit: 5000}
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.remaining > 0 {
		rl.remaining--
		return true
	}
	if time.Now().Before(rl.resetAt) {
		return false
	}
	return true
}

func (rl *RateLimiter) UpdateFromHeaders(headers http.Header) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if remaining := headers.Get("X-RateLimit-Remaining"); remaining != "" {
		if v, err := strconv.Atoi(remaining); err == nil {
			rl.remaining = v
		}
	}
	if limit := headers.Get("X-RateLimit-Limit"); limit != "" {
		if v, err := strconv.Atoi(limit); err == nil {
			rl.limit = v
		}
	}
	if reset := headers.Get("X-RateLimit-Reset"); reset != "" {
		if v, err := strconv.ParseInt(reset, 10, 64); err == nil {
			rl.resetAt = time.Unix(v, 0)
		}
	}
}

func (rl *RateLimiter) WaitIfNeeded() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.remaining <= 0 && time.Now().Before(rl.resetAt) {
		waitDuration := time.Until(rl.resetAt) + time.Second
		if waitDuration > 5*time.Minute {
			waitDuration = 5 * time.Minute
		}
		rl.mu.Unlock()
		time.Sleep(waitDuration)
		rl.mu.Lock()
	}
}

type Client struct {
	httpClient  *http.Client
	baseURL     string
	token       string
	rateLimiter *RateLimiter
	logger      *logrus.Logger
	userAgent   string
}

type ClientOption func(*Client)

func WithBaseURL(url string) ClientOption {
	return func(c *Client) { c.baseURL = url }
}

func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) { c.httpClient = hc }
}

func WithLogger(l *logrus.Logger) ClientOption {
	return func(c *Client) { c.logger = l }
}

func NewClient(token string, opts ...ClientOption) *Client {
	c := &Client{
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		baseURL:     defaultBaseURL,
		token:       token,
		rateLimiter: NewRateLimiter(),
		logger:      logrus.New(),
		userAgent:   defaultUserAgent,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, http.Header, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = strings.NewReader(string(jsonBody))
	}

	url := path
	if !strings.HasPrefix(path, "http") {
		url = c.baseURL + path
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := retryBaseDelay * time.Duration(1<<(attempt-1))
			select {
			case <-ctx.Done():
				return nil, nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		c.rateLimiter.WaitIfNeeded()

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return nil, nil, fmt.Errorf("create request: %w", err)
		}

		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", c.userAgent)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			c.logger.WithError(err).WithField("attempt", attempt+1).Warn("GitHub API request failed")
			lastErr = err
			continue
		}

		c.rateLimiter.UpdateFromHeaders(resp.Header)

		respBody, err := io.ReadAll(resp.Body)
		if err := resp.Body.Close(); err != nil {
			c.logger.WithError(err).Warn("failed to close response body")
		}
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == 403 && strings.Contains(string(respBody), "rate limit") {
			c.logger.WithFields(logrus.Fields{
				"attempt": attempt + 1,
				"status":  403,
			}).Warn("GitHub API rate limited, retrying")
			lastErr = fmt.Errorf("rate limited (403)")
			continue
		}

		if resp.StatusCode >= 500 {
			c.logger.WithFields(logrus.Fields{
				"attempt": attempt + 1,
				"status":  resp.StatusCode,
			}).Warn("GitHub API server error, retrying")
			lastErr = fmt.Errorf("server error (%d): %s", resp.StatusCode, string(respBody))
			continue
		}

		if resp.StatusCode >= 400 {
			return nil, resp.Header, fmt.Errorf("github API error (%d): %s", resp.StatusCode, string(respBody))
		}

		return respBody, resp.Header, nil
	}

	return nil, nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	body, _, err := c.doRequest(ctx, "GET", path, nil)
	return body, err
}

func (c *Client) post(ctx context.Context, path string, reqBody interface{}) ([]byte, error) {
	body, _, err := c.doRequest(ctx, "POST", path, reqBody)
	return body, err
}

func (c *Client) delete(ctx context.Context, path string) error {
	_, _, err := c.doRequest(ctx, "DELETE", path, nil)
	return err
}

func (c *Client) GetAuthenticatedUser(ctx context.Context) (*GitHubUser, error) {
	data, err := c.get(ctx, "/user")
	if err != nil {
		return nil, err
	}
	var user GitHubUser
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	return &user, nil
}

func (c *Client) ListRepos(ctx context.Context, opts ListReposOptions) ([]*GitHubRepo, error) {
	page := opts.Page
	if page <= 0 {
		page = 1
	}
	perPage := opts.PerPage
	if perPage <= 0 {
		perPage = 30
	}
	sort := opts.Sort
	if sort == "" {
		sort = "updated"
	}
	direction := opts.Direction
	if direction == "" {
		direction = "desc"
	}
	typ := opts.Type
	if typ == "" {
		typ = "all"
	}

	path := fmt.Sprintf("/user/repos?page=%d&per_page=%d&sort=%s&direction=%s&type=%s",
		page, perPage, sort, direction, typ)

	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}
	var repos []*GitHubRepo
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, fmt.Errorf("unmarshal repos: %w", err)
	}
	return repos, nil
}

func (c *Client) GetRepo(ctx context.Context, owner, repo string) (*GitHubRepo, error) {
	data, err := c.get(ctx, fmt.Sprintf("/repos/%s/%s", owner, repo))
	if err != nil {
		return nil, err
	}
	var r GitHubRepo
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("unmarshal repo: %w", err)
	}
	return &r, nil
}

func (c *Client) GetRepoContent(ctx context.Context, owner, repo, path, ref string) ([]*GitHubContent, error) {
	url := fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path)
	if ref != "" {
		url += "?ref=" + ref
	}
	data, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}
	var contents []*GitHubContent
	if err := json.Unmarshal(data, &contents); err != nil {
		var single GitHubContent
		if err2 := json.Unmarshal(data, &single); err2 == nil {
			return []*GitHubContent{&single}, nil
		}
		return nil, fmt.Errorf("unmarshal contents: %w", err)
	}
	return contents, nil
}

func (c *Client) GetFileContent(ctx context.Context, owner, repo, path, ref string) ([]byte, error) {
	url := fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path)
	if ref != "" {
		url += "?ref=" + ref
	}
	data, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}
	var content GitHubContent
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("unmarshal content: %w", err)
	}
	if content.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(content.Content)
		if err != nil {
			return nil, fmt.Errorf("decode base64: %w", err)
		}
		return decoded, nil
	}
	return []byte(content.Content), nil
}

func (c *Client) GetLanguages(ctx context.Context, owner, repo string) (map[string]float64, error) {
	data, err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/languages", owner, repo))
	if err != nil {
		return nil, err
	}
	var raw map[string]int
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal languages: %w", err)
	}
	total := 0
	for _, v := range raw {
		total += v
	}
	result := make(map[string]float64)
	if total > 0 {
		for k, v := range raw {
			result[k] = float64(v) / float64(total) * 100
		}
	}
	return result, nil
}

func (c *Client) ListBranches(ctx context.Context, owner, repo string) ([]*GitHubBranch, error) {
	data, err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/branches", owner, repo))
	if err != nil {
		return nil, err
	}
	var branches []*GitHubBranch
	if err := json.Unmarshal(data, &branches); err != nil {
		return nil, fmt.Errorf("unmarshal branches: %w", err)
	}
	return branches, nil
}

func (c *Client) GetTree(ctx context.Context, owner, repo, sha string, recursive bool) (*GitHubTree, error) {
	url := fmt.Sprintf("/repos/%s/%s/git/trees/%s", owner, repo, sha)
	if recursive {
		url += "?recursive=1"
	}
	data, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}
	var tree GitHubTree
	if err := json.Unmarshal(data, &tree); err != nil {
		return nil, fmt.Errorf("unmarshal tree: %w", err)
	}
	return &tree, nil
}

func (c *Client) CreateWebhook(ctx context.Context, owner, repo string, req *GitHubWebhookRequest) (*GitHubWebhookResponse, error) {
	data, err := c.post(ctx, fmt.Sprintf("/repos/%s/%s/hooks", owner, repo), req)
	if err != nil {
		return nil, err
	}
	var wh GitHubWebhookResponse
	if err := json.Unmarshal(data, &wh); err != nil {
		return nil, fmt.Errorf("unmarshal webhook: %w", err)
	}
	return &wh, nil
}

func (c *Client) DeleteWebhook(ctx context.Context, owner, repo string, webhookID int64) error {
	return c.delete(ctx, fmt.Sprintf("/repos/%s/%s/hooks/%d", owner, repo, webhookID))
}

func (c *Client) CreateCommitStatus(ctx context.Context, owner, repo, sha string, status *CommitStatusRequest) error {
	_, err := c.post(ctx, fmt.Sprintf("/repos/%s/%s/statuses/%s", owner, repo, sha), status)
	return err
}

func (c *Client) GetCompareDiff(ctx context.Context, owner, repo, base, head string) (map[string]interface{}, error) {
	data, err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/compare/%s...%s", owner, repo, base, head))
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal compare: %w", err)
	}
	return result, nil
}

func (c *Client) FetchCommits(ctx context.Context, owner, repo, sha string, perPage int) ([]GitHubCommit, error) {
	if perPage <= 0 {
		perPage = 30
	}
	path := fmt.Sprintf("/repos/%s/%s/commits?sha=%s&per_page=%d", owner, repo, sha, perPage)
	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}
	var commits []GitHubCommit
	if err := json.Unmarshal(data, &commits); err != nil {
		return nil, fmt.Errorf("unmarshal commits: %w", err)
	}
	return commits, nil
}
