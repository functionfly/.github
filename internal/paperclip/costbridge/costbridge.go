// Package costbridge sends FunctionFly agent execution costs to Paperclip
// so Paperclip can enforce monthly budgets and auto-pause.
package costbridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Config holds Paperclip cost-events API configuration (from env).
type Config struct {
	BaseURL    string // PAPERCLIP_BASE_URL
	APIKey     string // PAPERCLIP_API_KEY
	CompanyID  string // PAPERCLIP_COMPANY_ID
	AgentID    string // PAPERCLIP_AGENT_ID (optional; default agent for cost attribution)
}

// FromEnv returns config from environment variables.
func FromEnv() Config {
	return Config{
		BaseURL:   os.Getenv("PAPERCLIP_BASE_URL"),
		APIKey:    os.Getenv("PAPERCLIP_API_KEY"),
		CompanyID: os.Getenv("PAPERCLIP_COMPANY_ID"),
		AgentID:   os.Getenv("PAPERCLIP_AGENT_ID"),
	}
}

// ReportCost sends a single execution cost to Paperclip POST /api/companies/:companyId/cost-events.
// Cost is converted to integer cents. Metadata can include execution_id, function_uri, issue_id.
func ReportCost(ctx context.Context, cfg Config, costUSD float64, metadata map[string]string) error {
	if cfg.BaseURL == "" || cfg.APIKey == "" || cfg.CompanyID == "" {
		return nil
	}
	costCents := int(costUSD * 100)
	if costCents < 0 {
		costCents = 0
	}
	payload := map[string]interface{}{
		"costCents": costCents,
		"occurredAt": time.Now().UTC().Format(time.RFC3339),
	}
	if cfg.AgentID != "" {
		payload["agentId"] = cfg.AgentID
	}
	if len(metadata) > 0 {
		payload["metadata"] = metadata
	}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/api/companies/%s/cost-events", cfg.BaseURL, cfg.CompanyID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("paperclip cost-events: %d", resp.StatusCode)
	}
	return nil
}
