package chat

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

type WebConnector struct {
	logger *logrus.Logger
}

func NewWebConnector(logger *logrus.Logger) *WebConnector {
	if logger == nil {
		logger = logrus.New()
	}
	return &WebConnector{logger: logger}
}

func (c *WebConnector) Name() string { return "Web Search" }
func (c *WebConnector) Icon() string { return "globe" }
func (c *WebConnector) IsConfigured() bool { return true }

func (c *WebConnector) Authenticate(ctx context.Context, creds map[string]string) error {
	return nil
}

func (c *WebConnector) FetchData(ctx context.Context, config map[string]interface{}) (interface{}, error) {
	urlStr := config["url"]
	if urlStr == nil {
		return nil, fmt.Errorf("url is required")
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr.(string), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body := make([]byte, 4096)
	n, _ := resp.Body.Read(body)
	return strings.TrimSpace(string(body[:n])), nil
}

func (c *WebConnector) Search(ctx context.Context, query string, config map[string]interface{}) ([]SearchResult, error) {
	return []SearchResult{
		{
			Title:   "Web search: " + query,
			Content: "Web search functionality - integrate with search API for full results",
			URL:     "https://www.google.com/search?q=" + query,
		},
	}, nil
}
