package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIClient is a simple HTTP client for the FunctionFly API.
type APIClient struct {
	BaseURL string
	Token   string
	client  *http.Client
}

// NewAPIClient creates a new API client using stored credentials.
func NewAPIClient() (*APIClient, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return nil, err
	}
	cfg, _ := LoadConfig()
	baseURL := "https://api.functionfly.com"
	if cfg != nil && cfg.API.URL != "" {
		baseURL = cfg.API.URL
	}
	return &APIClient{BaseURL: baseURL, Token: creds.Token, client: &http.Client{Timeout: 30 * time.Second}}, nil
}

// NewAPIClientWithToken creates a new API client with an explicit token.
func NewAPIClientWithToken(token string) *APIClient {
	cfg, _ := LoadConfig()
	baseURL := "https://api.functionfly.com"
	if cfg != nil && cfg.API.URL != "" {
		baseURL = cfg.API.URL
	}
	return &APIClient{BaseURL: baseURL, Token: token, client: &http.Client{Timeout: 30 * time.Second}}
}

func (c *APIClient) Get(path string, out interface{}) error {
	req, err := http.NewRequest("GET", c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *APIClient) Post(path string, body interface{}, out interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out)
}

func (c *APIClient) Put(path string, body interface{}, out interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", c.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out)
}

func (c *APIClient) Delete(path string, out interface{}) error {
	req, err := http.NewRequest("DELETE", c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *APIClient) do(req *http.Request, out interface{}) error {
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("User-Agent", "fly-cli/1.0.0")
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %w\n   → Check your internet connection", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		var errResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(body, &errResp)
		msg := errResp.Error
		if msg == "" {
			msg = errResp.Message
		}
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		hint := ""
		switch resp.StatusCode {
		case 401:
			hint = "\n   → Your session may have expired — run: fly login"
		case 403:
			hint = "\n   → You don't have permission to perform this action"
		case 404:
			hint = "\n   → The resource was not found"
		case 409:
			hint = "\n   → This version already exists — run: fly update patch"
		case 429:
			hint = "\n   → Rate limited — please wait a moment and try again"
		}
		return fmt.Errorf("%s%s", msg, hint)
	}
	if out != nil && len(body) > 0 {
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("could not parse response: %w", err)
		}
	}
	return nil
}

// StreamLines opens a streaming GET connection and calls fn for each line.
func (c *APIClient) StreamLines(path string, fn func(line string) bool) error {
	req, err := http.NewRequest("GET", c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("User-Agent", "fly-cli/1.0.0")
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	buf := make([]byte, 4096)
	var line []byte
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			for _, b := range buf[:n] {
				if b == '\n' {
					if len(line) > 0 {
						if !fn(string(line)) {
							return nil
						}
						line = line[:0]
					}
				} else {
					line = append(line, b)
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return nil
}
