package vault

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// AuditEntry is one row of the vault audit log.
type AuditEntry struct {
	ID        string                 `json:"id"`
	SecretID  string                 `json:"secret_id,omitempty"`
	Action    string                 `json:"action"`
	ActorID   string                 `json:"actor_id"`
	ActorType string                 `json:"actor_type"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error_message,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// AuditListOptions controls Audit.List.
type AuditListOptions struct {
	SecretID string
	Action   string
	ActorID  string
	Limit    int
	Offset   int
}

// AuditList is the response from Audit.List.
type AuditList struct {
	Entries []AuditEntry `json:"entries"`
	Total   int64        `json:"total"`
	Limit   int          `json:"limit"`
	Offset  int          `json:"offset"`
}

// AuditService queries the vault audit log.
type AuditService struct{ client *Client }

// List returns paginated audit entries.
func (a *AuditService) List(ctx context.Context, opts AuditListOptions) (*AuditList, error) {
	v := url.Values{}
	if opts.SecretID != "" {
		v.Set("secret_id", opts.SecretID)
	}
	if opts.Action != "" {
		v.Set("action", opts.Action)
	}
	if opts.ActorID != "" {
		v.Set("actor_id", opts.ActorID)
	}
	if opts.Limit > 0 {
		v.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		v.Set("offset", strconv.Itoa(opts.Offset))
	}
	path := "/v1/vault/audit"
	if encoded := v.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out AuditList
	if err := a.client.do(ctx, "GET", path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
