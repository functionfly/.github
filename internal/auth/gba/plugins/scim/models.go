// Package scim provides SCIM 2.0 provisioning support for GoBetterAuth
// This is Phase 4 of the Better Auth migration plan - Enterprise user lifecycle management
package scim

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SCIMConfig stores SCIM configuration for a tenant
type SCIMConfig struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID   uuid.UUID      `gorm:"type:uuid;not null;index;uniqueIndex:idx_gba_scim_config_tenant;column:tenant_id"`
	Enabled    bool           `gorm:"default:false"`
	TokenHash  string         `gorm:"type:text;not null;column:token_hash"` // Bearer token (bcrypt hashed)
	SyncGroups bool           `gorm:"default:true;column:sync_groups"`
	SyncUsers  bool           `gorm:"default:true;column:sync_users"`
	LastSyncAt *time.Time     `gorm:"column:last_sync_at"`
	CreatedAt  time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for the SCIMConfig model
func (SCIMConfig) TableName() string {
	return "gba_scim_configs"
}

// BeforeCreate hook to set timestamps and ID
func (c *SCIMConfig) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.Must(uuid.NewRandom())
	}
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook to update the updated_at timestamp
func (c *SCIMConfig) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}

// SCIMUser represents a SCIM user resource
type SCIMUser struct {
	ID          uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID    uuid.UUID       `gorm:"type:uuid;not null;index;column:tenant_id"`
	ExternalID  string          `gorm:"type:varchar(255);column:external_id"`
	UserName    string          `gorm:"type:varchar(255);not null;column:user_name"`
	DisplayName string          `gorm:"type:varchar(255);column:display_name"`
	Emails      SCIMAttributes  `gorm:"type:jsonb;column:emails"`
	Active      bool            `gorm:"default:true"`
	Groups      SCIMGroups      `gorm:"type:jsonb;column:groups"`
	Raw         json.RawMessage `gorm:"type:jsonb;column:raw"`
	CreatedAt   time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt  `gorm:"index"`
}

// TableName returns the table name for the SCIMUser model
func (SCIMUser) TableName() string {
	return "gba_scim_users"
}

// BeforeCreate hook to set timestamps and ID
func (u *SCIMUser) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.Must(uuid.NewRandom())
	}
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook to update the updated_at timestamp
func (u *SCIMUser) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

// ToSCIMResponse converts the SCIMUser to a SCIM API response
func (u *SCIMUser) ToSCIMResponse(baseURL string) map[string]interface{} {
	return map[string]interface{}{
		"schemas":     []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		"id":          u.ID.String(),
		"externalId":  u.ExternalID,
		"userName":    u.UserName,
		"displayName": u.DisplayName,
		"emails":      u.Emails.ToSlice(),
		"active":      u.Active,
		"groups":      u.Groups.ToSlice(),
		"meta": map[string]interface{}{
			"resourceType": "User",
			"created":      u.CreatedAt.Format(time.RFC3339),
			"lastModified": u.UpdatedAt.Format(time.RFC3339),
			"location":     baseURL + "/Users/" + u.ID.String(),
		},
	}
}

// SCIMGroup represents a SCIM group resource
type SCIMGroup struct {
	ID          uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID    uuid.UUID       `gorm:"type:uuid;not null;index;column:tenant_id"`
	ExternalID  string          `gorm:"type:varchar(255);column:external_id"`
	DisplayName string          `gorm:"type:varchar(255);not null;column:display_name"`
	Members     SCIMMembers     `gorm:"type:jsonb;column:members"`
	Raw         json.RawMessage `gorm:"type:jsonb;column:raw"`
	CreatedAt   time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time       `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt  `gorm:"index"`
}

// TableName returns the table name for the SCIMGroup model
func (SCIMGroup) TableName() string {
	return "gba_scim_groups"
}

// BeforeCreate hook to set timestamps and ID
func (g *SCIMGroup) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.Must(uuid.NewRandom())
	}
	g.CreatedAt = time.Now()
	g.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook to update the updated_at timestamp
func (g *SCIMGroup) BeforeUpdate(tx *gorm.DB) error {
	g.UpdatedAt = time.Now()
	return nil
}

// ToSCIMResponse converts the SCIMGroup to a SCIM API response
func (g *SCIMGroup) ToSCIMResponse(baseURL string) map[string]interface{} {
	return map[string]interface{}{
		"schemas":     []string{"urn:ietf:params:scim:schemas:core:2.0:Group"},
		"id":          g.ID.String(),
		"externalId":  g.ExternalID,
		"displayName": g.DisplayName,
		"members":     g.Members.ToSlice(),
		"meta": map[string]interface{}{
			"resourceType": "Group",
			"created":      g.CreatedAt.Format(time.RFC3339),
			"lastModified": g.UpdatedAt.Format(time.RFC3339),
			"location":     baseURL + "/Groups/" + g.ID.String(),
		},
	}
}

// SCIMAttributes represents SCIM email attributes
type SCIMAttributes []SCIMAttribute

// SCIMAttribute represents a single SCIM attribute (e.g., email)
type SCIMAttribute struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

// ToSlice converts SCIMAttributes to a slice of maps for JSON response
func (a SCIMAttributes) ToSlice() []map[string]interface{} {
	result := make([]map[string]interface{}, len(a))
	for i, attr := range a {
		result[i] = map[string]interface{}{
			"value":   attr.Value,
			"type":    attr.Type,
			"primary": attr.Primary,
		}
	}
	return result
}

// SCIMGroups represents a list of SCIM group memberships
type SCIMGroups []SCIMGroupRef

// SCIMGroupRef represents a reference to a SCIM group
type SCIMGroupRef struct {
	Value   string `json:"value"`
	Display string `json:"display,omitempty"`
}

// ToSlice converts SCIMGroups to a slice of maps for JSON response
func (g SCIMGroups) ToSlice() []map[string]interface{} {
	result := make([]map[string]interface{}, len(g))
	for i, group := range g {
		result[i] = map[string]interface{}{
			"value":   group.Value,
			"display": group.Display,
		}
	}
	return result
}

// SCIMMembers represents a list of SCIM group members
type SCIMMembers []SCIMMember

// SCIMMember represents a single group member
type SCIMMember struct {
	Value   string `json:"value"`
	Display string `json:"display,omitempty"`
	Type    string `json:"type,omitempty"` // "User" or "Group"
}

// ToSlice converts SCIMMembers to a slice of maps for JSON response
func (m SCIMMembers) ToSlice() []map[string]interface{} {
	result := make([]map[string]interface{}, len(m))
	for i, member := range m {
		result[i] = map[string]interface{}{
			"value":   member.Value,
			"display": member.Display,
			"type":    member.Type,
		}
	}
	return result
}

// SCIMError represents a SCIM error response
type SCIMError struct {
	Schemas  []string `json:"schemas"`
	Status   string   `json:"status"`
	Detail   string   `json:"detail,omitempty"`
	ScimType string   `json:"scimType,omitempty"`
}

// NewSCIMError creates a new SCIM error response
func NewSCIMError(status int, detail string) *SCIMError {
	return &SCIMError{
		Schemas: []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
		Status:  fmt.Sprintf("%d", status),
		Detail:  detail,
	}
}

// SCIMListResponse represents a list response for SCIM resources
type SCIMListResponse struct {
	Schemas      []string                 `json:"schemas"`
	TotalResults int                      `json:"totalResults"`
	StartIndex   int                      `json:"startIndex"`
	ItemsPerPage int                      `json:"itemsPerPage"`
	Resources    []map[string]interface{} `json:"Resources"`
}

// NewSCIMListResponse creates a new SCIM list response
func NewSCIMListResponse(startIndex, itemsPerPage, totalResults int, resources []map[string]interface{}) *SCIMListResponse {
	return &SCIMListResponse{
		Schemas:      []string{"urn:ietf:params:scim:api:messages:2.0:ListResponse"},
		TotalResults: totalResults,
		StartIndex:   startIndex,
		ItemsPerPage: itemsPerPage,
		Resources:    resources,
	}
}

// SCIMServiceProviderConfig represents the SCIM service provider configuration
type SCIMServiceProviderConfig struct {
	Schemas               []string                 `json:"schemas"`
	DocumentationURI      string                   `json:"documentationUri,omitempty"`
	Patch                 map[string]interface{}   `json:"patch"`
	Bulk                  map[string]interface{}   `json:"bulk"`
	Filter                map[string]interface{}   `json:"filter"`
	ChangePassword        map[string]interface{}   `json:"changePassword"`
	Sort                  map[string]interface{}   `json:"sort"`
	ETag                  map[string]interface{}   `json:"etag"`
	AuthenticationSchemes []map[string]interface{} `json:"authenticationSchemes"`
	Meta                  map[string]interface{}   `json:"meta"`
}

// DefaultServiceProviderConfig returns the default service provider config
func DefaultServiceProviderConfig(baseURL string) *SCIMServiceProviderConfig {
	return &SCIMServiceProviderConfig{
		Schemas: []string{"urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"},
		Patch: map[string]interface{}{
			"supported": true,
		},
		Bulk: map[string]interface{}{
			"supported":      true,
			"maxOperations":  100,
			"maxPayloadSize": 1048576,
		},
		Filter: map[string]interface{}{
			"supported":  true,
			"maxResults": 200,
		},
		ChangePassword: map[string]interface{}{
			"supported": false,
		},
		Sort: map[string]interface{}{
			"supported": false,
		},
		ETag: map[string]interface{}{
			"supported": false,
		},
		AuthenticationSchemes: []map[string]interface{}{
			{
				"type":             "oauthbearertoken",
				"name":             "OAuth Bearer Token",
				"description":      "Authentication using OAuth Bearer Token",
				"specURI":          "https://www.rfc-editor.org/rfc/rfc6750",
				"documentationURI": "",
				"primary":          true,
			},
		},
		Meta: map[string]interface{}{
			"resourceType": "ServiceProviderConfig",
			"location":     baseURL + "/ServiceProviderConfig",
		},
	}
}

// SCIMResourceType represents a SCIM resource type
type SCIMResourceType struct {
	ID               string                   `json:"id"`
	Name             string                   `json:"name"`
	Endpoint         string                   `json:"endpoint"`
	Description      string                   `json:"description"`
	Schema           string                   `json:"schema"`
	SchemaExtensions []map[string]interface{} `json:"schemaExtensions,omitempty"`
	Meta             map[string]interface{}   `json:"meta"`
}

// SCIMResourceTypes returns the list of supported resource types
func SCIMResourceTypes(baseURL string) []SCIMResourceType {
	return []SCIMResourceType{
		{
			ID:          "User",
			Name:        "User",
			Endpoint:    baseURL + "/Users",
			Description: "User Account",
			Schema:      "urn:ietf:params:scim:schemas:core:2.0:User",
			Meta: map[string]interface{}{
				"resourceType": "ResourceType",
				"location":     baseURL + "/ResourceTypes/User",
			},
		},
		{
			ID:          "Group",
			Name:        "Group",
			Endpoint:    baseURL + "/Groups",
			Description: "Group",
			Schema:      "urn:ietf:params:scim:schemas:core:2.0:Group",
			Meta: map[string]interface{}{
				"resourceType": "ResourceType",
				"location":     baseURL + "/ResourceTypes/Group",
			},
		},
	}
}

// SCIMSchema represents a SCIM schema definition
type SCIMSchema struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Attributes  []SCIMSchemaAttribute  `json:"attributes"`
	Meta        map[string]interface{} `json:"meta"`
}

// SCIMSchemaAttribute represents a SCIM schema attribute
type SCIMSchemaAttribute struct {
	Name            string                `json:"name"`
	Type            string                `json:"type"`
	MultiValued     bool                  `json:"multiValued"`
	Description     string                `json:"description"`
	Required        bool                  `json:"required"`
	CanonicalValues []string              `json:"canonicalValues,omitempty"`
	CaseExact       bool                  `json:"caseExact"`
	Mutability      string                `json:"mutability"`
	Returned        string                `json:"returned"`
	Uniqueness      string                `json:"uniqueness"`
	SubAttributes   []SCIMSchemaAttribute `json:"subAttributes,omitempty"`
}

// SCIMSchemas returns the list of supported schemas
func SCIMSchemas(baseURL string) []SCIMSchema {
	// Return standard SCIM 2.0 schemas
	return []SCIMSchema{
		{
			ID:          "urn:ietf:params:scim:schemas:core:2.0:User",
			Name:        "User",
			Description: "User Account",
			Attributes:  []SCIMSchemaAttribute{},
			Meta: map[string]interface{}{
				"resourceType": "Schema",
				"location":     baseURL + "/Schemas/urn:ietf:params:scim:schemas:core:2.0:User",
			},
		},
		{
			ID:          "urn:ietf:params:scim:schemas:core:2.0:Group",
			Name:        "Group",
			Description: "Group",
			Attributes:  []SCIMSchemaAttribute{},
			Meta: map[string]interface{}{
				"resourceType": "Schema",
				"location":     baseURL + "/Schemas/urn:ietf:params:scim:schemas:core:2.0:Group",
			},
		},
	}
}

// SCIMUserRequest represents a request to create/update a SCIM user
type SCIMUserRequest struct {
	Schemas     []string                 `json:"schemas"`
	ExternalID  string                   `json:"externalId,omitempty"`
	UserName    string                   `json:"userName"`
	DisplayName string                   `json:"displayName,omitempty"`
	Emails      []map[string]interface{} `json:"emails,omitempty"`
	Active      bool                     `json:"active"`
	Groups      []map[string]interface{} `json:"groups,omitempty"`
	Name        map[string]interface{}   `json:"name,omitempty"`
}

// SCIMGroupRequest represents a request to create/update a SCIM group
type SCIMGroupRequest struct {
	Schemas     []string                 `json:"schemas"`
	ExternalID  string                   `json:"externalId,omitempty"`
	DisplayName string                   `json:"displayName"`
	Members     []map[string]interface{} `json:"members,omitempty"`
}

// SCIMPatchRequest represents a SCIM patch request
type SCIMPatchRequest struct {
	Schemas    []string             `json:"schemas"`
	Operations []SCIMPatchOperation `json:"Operations"`
}

// SCIMPatchOperation represents a single patch operation
type SCIMPatchOperation struct {
	Op    string      `json:"op"` // "add", "remove", "replace"
	Path  string      `json:"path,omitempty"`
	Value interface{} `json:"value,omitempty"`
}

// SCIMConfigRequest represents a request to configure SCIM for a tenant
type SCIMConfigRequest struct {
	Enabled    bool `json:"enabled"`
	SyncGroups bool `json:"sync_groups"`
	SyncUsers  bool `json:"sync_users"`
}

// SCIMConfigResponse represents the SCIM configuration response
type SCIMConfigResponse struct {
	ID         string     `json:"id"`
	TenantID   string     `json:"tenant_id"`
	Enabled    bool       `json:"enabled"`
	SyncGroups bool       `json:"sync_groups"`
	SyncUsers  bool       `json:"sync_users"`
	LastSyncAt *time.Time `json:"last_sync_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// SCIMTokenResponse represents a generated SCIM token response
type SCIMTokenResponse struct {
	Token     string    `json:"token"`
	TokenHash string    `json:"token_hash,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// SCIMStatusResponse represents the SCIM status for a tenant
type SCIMStatusResponse struct {
	Enabled    bool `json:"enabled"`
	Configured bool `json:"configured"`
}

// Validate validates a SCIM user request
func (r *SCIMUserRequest) Validate() error {
	if r.UserName == "" {
		return fmt.Errorf("userName is required")
	}
	return nil
}

// Validate validates a SCIM group request
func (r *SCIMGroupRequest) Validate() error {
	if r.DisplayName == "" {
		return fmt.Errorf("displayName is required")
	}
	return nil
}

// ParseEmails parses emails from the request
func (r *SCIMUserRequest) ParseEmails() []SCIMAttribute {
	emails := make([]SCIMAttribute, 0, len(r.Emails))
	for _, e := range r.Emails {
		email := SCIMAttribute{}
		if val, ok := e["value"].(string); ok {
			email.Value = val
		}
		if val, ok := e["type"].(string); ok {
			email.Type = val
		}
		if val, ok := e["primary"].(bool); ok {
			email.Primary = val
		}
		emails = append(emails, email)
	}
	return emails
}

// ParseGroups parses group memberships from the request
func (r *SCIMUserRequest) ParseGroups() []SCIMGroupRef {
	groups := make([]SCIMGroupRef, 0, len(r.Groups))
	for _, g := range r.Groups {
		group := SCIMGroupRef{}
		if val, ok := g["value"].(string); ok {
			group.Value = val
		}
		if val, ok := g["display"].(string); ok {
			group.Display = val
		}
		groups = append(groups, group)
	}
	return groups
}

// ParseMembers parses members from the request
func (r *SCIMGroupRequest) ParseMembers() []SCIMMember {
	members := make([]SCIMMember, 0, len(r.Members))
	for _, m := range r.Members {
		member := SCIMMember{}
		if val, ok := m["value"].(string); ok {
			member.Value = val
		}
		if val, ok := m["display"].(string); ok {
			member.Display = val
		}
		if val, ok := m["type"].(string); ok {
			member.Type = val
		}
		members = append(members, member)
	}
	return members
}
