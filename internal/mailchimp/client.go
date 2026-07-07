package mailchimp

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	ErrSubscriberNotFound = errors.New("subscriber not found")
	ErrInvalidAPIKey      = errors.New("invalid API key")
)

type Client struct {
	config      *Config
	httpClient  *http.Client
	apiBaseURL  string
}

func NewClient(cfg *Config) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiBaseURL: fmt.Sprintf("https://%s.api.mailchimp.com/3.0", cfg.ServerPrefix),
	}
}

type SubscriberStatus string

const (
	StatusSubscribed   SubscriberStatus = "subscribed"
	StatusUnsubscribed SubscriberStatus = "unsubscribed"
	StatusPending      SubscriberStatus = "pending"
	StatusCleaned      SubscriberStatus = "cleaned"
)

type MergeFields map[string]string

type Subscriber struct {
	ID              string          `json:"id,omitempty"`
	EmailAddress    string          `json:"email_address"`
	Status          SubscriberStatus `json:"status"`
	StatusIfNew     SubscriberStatus `json:"status_if_new,omitempty"`
	MergeFields     MergeFields     `json:"merge_fields,omitempty"`
	Tags            []string        `json:"tags,omitempty"`
	EmailType       string          `json:"email_type,omitempty"`
	VIP             bool            `json:"vip,omitempty"`
	Location        *Location       `json:"location,omitempty"`
	MarketingPerm   string          `json:"marketing_permissions,omitempty"`
	TimestampSignup string          `json:"timestamp_signup,omitempty"`
	TimestampOpt    string          `json:"timestamp_opt,omitempty"`
	LastChange      string          `json:"last_changed,omitempty"`
	LastNote        *Note           `json:"last_note,omitempty"`
	Source          string          `json:"source,omitempty"`
	Language        string          `json:"language,omitempty"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Note struct {
	NoteID    int    `json:"note_id"`
	CreatedAt string `json:"created_at"`
	CreatedBy string `json:"created_by"`
	Text      string `json:"text"`
}

type SubscriberResponse struct {
	ID           string            `json:"id"`
	EmailAddress  string            `json:"email_address"`
	Status       SubscriberStatus  `json:"status"`
	MergeFields  map[string]string `json:"merge_fields,omitempty"`
	LastChanged  string            `json:"last_changed"`
}

type ListStats struct {
	MemberCount               int     `json:"member_count"`
	UnsubscribeCount          int     `json:"unsubscribe_count"`
	CleanedCount              int     `json:"cleaned_count"`
	MemberCountSinceSend      int     `json:"member_count_since_send"`
	UnsubscribeCountSinceSend int     `json:"unsubscribe_count_since_send"`
	CleanedCountSinceSend     int     `json:"cleaned_count_since_send"`
	CampaignCount             int     `json:"campaign_count"`
	AvgSubRate                float64 `json:"avg_sub_rate"`
	AvgUnsubRate              float64 `json:"avg_unsub_rate"`
	TargetSubRate             float64 `json:"target_sub_rate"`
	OpenRate                  float64 `json:"open_rate"`
	ClickRate                 float64 `json:"click_rate"`
	LastSubDate              string  `json:"last_sub_date"`
	LastUnsubDate            string  `json:"last_unsub_date"`
}

type ActivityInfo struct {
	EmailAddress  string      `json:"email_address"`
	Activity     []Activity  `json:"activity"`
}

type Activity struct {
	Action    string `json:"action"`
	Timestamp string `json:"timestamp"`
	URL       string `json:"url,omitempty"`
}

type MergeField struct {
	MergeID   int    `json:"merge_id"`
	Tag       string `json:"tag"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Required  bool   `json:"required"`
	DefaultValue string `json:"default_value,omitempty"`
	Public    bool   `json:"public"`
	DisplayOrder int  `json:"display_order"`
}

type BatchOperation struct {
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Body        string            `json:"body,omitempty"`
	Params      map[string]string `json:"params,omitempty"`
}

type BatchResponse struct {
	ID          string   `json:"id"`
	Status      string   `json:"status"`
	TotalOps    int      `json:"total_operations"`
	FinishedOps int      `json:"finished_operations"`
	ErroredOps  int      `json:"errored_operations"`
	Response    []BatchResp `json:"response,omitempty"`
}

type BatchResp struct {
	StatusCode int             `json:"status_code"`
	OperationID string         `json:"operation_id"`
	Body        string         `json:"body,omitempty"`
}

func (c *Client) emailHash(email string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	hash := md5.Sum([]byte(email))
	return fmt.Sprintf("%x", hash)
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	url := c.apiBaseURL + path
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("anystring", c.config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func (c *Client) handleError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}

	var errResp struct {
		Title   string `json:"title"`
		Detail  string `json:"detail"`
		Status  int    `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("mailchimp error: status %d", resp.StatusCode)
	}

	return fmt.Errorf("mailchimp error [%d]: %s - %s", errResp.Status, errResp.Title, errResp.Detail)
}

func (c *Client) Subscribe(ctx context.Context, email, firstName, lastName string, tags []string, mergeFields MergeFields) (*SubscriberResponse, error) {
	if !c.config.IsConfigured() {
		return nil, errors.New("mailchimp client not configured")
	}

	subscriber := Subscriber{
		EmailAddress: email,
		Status:       StatusSubscribed,
		StatusIfNew:  StatusSubscribed,
		MergeFields:  mergeFields,
		Tags:         tags,
	}

	if firstName != "" || lastName != "" {
		if subscriber.MergeFields == nil {
			subscriber.MergeFields = make(MergeFields)
		}
		if firstName != "" {
			subscriber.MergeFields["FNAME"] = firstName
		}
		if lastName != "" {
			subscriber.MergeFields["LNAME"] = lastName
		}
	}

	resp, err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/lists/%s/members", c.config.DefaultListID), subscriber)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.handleError(resp); err != nil {
		if strings.Contains(err.Error(), "already a list member") {
			return c.UpdateSubscriber(ctx, email, nil, mergeFields)
		}
		return nil, err
	}

	var result SubscriberResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) Unsubscribe(ctx context.Context, email string) error {
	if !c.config.IsConfigured() {
		return errors.New("mailchimp client not configured")
	}

	hash := c.emailHash(email)
	path := fmt.Sprintf("/lists/%s/members/%s", c.config.DefaultListID, hash)

	body := map[string]string{"status": "unsubscribed"}
	resp, err := c.doRequest(ctx, http.MethodPatch, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.handleError(resp); err != nil {
		if strings.Contains(err.Error(), "already unsubscribed") {
			return nil
		}
		return err
	}

	return nil
}

func (c *Client) UpdateSubscriber(ctx context.Context, email string, tags *[]string, mergeFields MergeFields) (*SubscriberResponse, error) {
	if !c.config.IsConfigured() {
		return nil, errors.New("mailchimp client not configured")
	}

	hash := c.emailHash(email)
	path := fmt.Sprintf("/lists/%s/members/%s", c.config.DefaultListID, hash)

	updateBody := make(map[string]interface{})
	if mergeFields != nil {
		updateBody["merge_fields"] = mergeFields
	}
	if tags != nil {
		updateBody["tags"] = tags
	}

	resp, err := c.doRequest(ctx, http.MethodPatch, path, updateBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.handleError(resp); err != nil {
		return nil, err
	}

	var result SubscriberResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetSubscriber(ctx context.Context, email string) (*Subscriber, error) {
	if !c.config.IsConfigured() {
		return nil, errors.New("mailchimp client not configured")
	}

	hash := c.emailHash(email)
	path := fmt.Sprintf("/lists/%s/members/%s", c.config.DefaultListID, hash)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.handleError(resp); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrSubscriberNotFound
		}
		return nil, err
	}

	var subscriber Subscriber
	if err := json.NewDecoder(resp.Body).Decode(&subscriber); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &subscriber, nil
}

func (c *Client) GetListStats(ctx context.Context) (*ListStats, error) {
	if !c.config.IsConfigured() {
		return nil, errors.New("mailchimp client not configured")
	}

	path := fmt.Sprintf("/lists/%s", c.config.DefaultListID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.handleError(resp); err != nil {
		return nil, err
	}

	var stats ListStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &stats, nil
}

func (c *Client) GetSubscriberActivity(ctx context.Context, email string) (*ActivityInfo, error) {
	if !c.config.IsConfigured() {
		return nil, errors.New("mailchimp client not configured")
	}

	hash := c.emailHash(email)
	path := fmt.Sprintf("/lists/%s/members/%s/activity", c.config.DefaultListID, hash)

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.handleError(resp); err != nil {
		return nil, err
	}

	var activity ActivityInfo
	if err := json.NewDecoder(resp.Body).Decode(&activity); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &activity, nil
}

func (c *Client) GetListMergeFields(ctx context.Context) ([]MergeField, error) {
	if !c.config.IsConfigured() {
		return nil, errors.New("mailchimp client not configured")
	}

	path := fmt.Sprintf("/lists/%s/merge-fields", c.config.DefaultListID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.handleError(resp); err != nil {
		return nil, err
	}

	var result struct {
		MergeFields []MergeField `json:"merge_fields"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.MergeFields, nil
}

func (c *Client) CreateMergeField(ctx context.Context, tag, name, fieldType string) (*MergeField, error) {
	if !c.config.IsConfigured() {
		return nil, errors.New("mailchimp client not configured")
	}

	path := fmt.Sprintf("/lists/%s/merge-fields", c.config.DefaultListID)
	body := map[string]string{
		"tag": tag,
		"name": name,
		"type": fieldType,
	}

	resp, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.handleError(resp); err != nil {
		return nil, err
	}

	var mergeField MergeField
	if err := json.NewDecoder(resp.Body).Decode(&mergeField); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &mergeField, nil
}

func (c *Client) EnsureFrequencyMergeField(ctx context.Context) error {
	fields, err := c.GetListMergeFields(ctx)
	if err != nil {
		return err
	}

	for _, f := range fields {
		if f.Tag == "FREQUENCY" {
			return nil
		}
	}

	_, err = c.CreateMergeField(ctx, "FREQUENCY", "Email Frequency", "text")
	return err
}

func (c *Client) GetWebhookSecret() string {
	if c == nil || c.config == nil {
		return ""
	}
	return c.config.WebhookSecret
}

func (c *Client) IsConfigured() bool {
	return c != nil && c.config != nil && c.config.IsConfigured()
}

func (c *Client) IsSyncEnabled() bool {
	if c == nil || c.config == nil {
		return false
	}
	return c.config.SyncEnabled
}
