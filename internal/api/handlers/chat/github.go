package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type GitHubConnector struct {
	logger *logrus.Logger
}

func NewGitHubConnector(logger *logrus.Logger) *GitHubConnector {
	if logger == nil {
		logger = logrus.New()
	}
	return &GitHubConnector{logger: logger}
}

func (c *GitHubConnector) Name() string { return "GitHub" }
func (c *GitHubConnector) Icon() string { return "github" }
func (c *GitHubConnector) IsConfigured() bool { return true }

func (c *GitHubConnector) Authenticate(ctx context.Context, creds map[string]string) error {
	token := creds["token"]
	if token == "" {
		return fmt.Errorf("GitHub token is required")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub authentication failed: status %d", resp.StatusCode)
	}
	return nil
}

func (c *GitHubConnector) FetchData(ctx context.Context, config map[string]interface{}) (interface{}, error) {
	token := config["token"]
	if token == nil {
		token = ""
	}

	owner := config["owner"]
	repo := config["repo"]

	url := "https://api.github.com/repos/" + fmt.Sprintf("%v/%v", owner, repo) + "/issues?state=open"
	if owner == nil || repo == nil {
		url = "https://api.github.com/user/repos?per_page=10"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.(string))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

func (c *GitHubConnector) Search(ctx context.Context, query string, config map[string]interface{}) ([]SearchResult, error) {
	token := config["token"]
	if token == nil {
		token = ""
	}

	url := "https://api.github.com/search/issues?q=" + query
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.(string))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			Title   string `json:"title"`
			Body    string `json:"body"`
			HTMLURL string `json:"html_url"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, item := range result.Items {
		results = append(results, SearchResult{
			Title:   item.Title,
			Content: item.Body,
			URL:     item.HTMLURL,
		})
	}
	return results, nil
}
