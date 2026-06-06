package storage

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Connector struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	IconURL   string    `json:"icon_url"`
	OAuthURL  string    `json:"oauth_url"`
	Scopes    string    `json:"scopes"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type UserConnector struct {
	ID                    uuid.UUID       `json:"id"`
	TenantID              uuid.UUID       `json:"tenant_id"`
	ConnectorID           uuid.UUID       `json:"connector_id"`
	ConnectorSlug         string          `json:"connector_slug,omitempty"`
	ConnectorName         string          `json:"connector_name,omitempty"`
	ConnectorIconURL      string          `json:"connector_icon_url,omitempty"`
	DisplayName           string          `json:"display_name"`
	Status                string          `json:"status"`
	EncryptedCredentials  json.RawMessage `json:"encrypted_credentials"`
	LastSyncAt            *time.Time      `json:"last_sync_at,omitempty"`
	SyncError             string          `json:"sync_error,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

type LinkConnectorRequest struct {
	ConnectorSlug        string          `json:"connector_slug"`
	DisplayName          string          `json:"display_name"`
	EncryptedCredentials json.RawMessage `json:"encrypted_credentials"`
	OAuthCode            string          `json:"oauth_code,omitempty"`
	RedirectURI          string          `json:"redirect_uri,omitempty"`
}

type ConnectorCallbackRequest struct {
	Code        string `json:"code"`
	State       string `json:"state"`
	RedirectURI string `json:"redirect_uri"`
}

type SyncTriggerResponse struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	StartedAt time.Time `json:"started_at"`
}
