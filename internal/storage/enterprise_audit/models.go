package enterpriseaudit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ServiceArea string

const (
	ServiceAreaAuth       ServiceArea = "auth"
	ServiceAreaVault      ServiceArea = "vault"
	ServiceAreaBilling    ServiceArea = "billing"
	ServiceAreaFunctions  ServiceArea = "functions"
	ServiceAreaRegistry   ServiceArea = "registry"
	ServiceAreaAgents     ServiceArea = "agents"
	ServiceAreaTeams      ServiceArea = "teams"
	ServiceAreaAPI        ServiceArea = "api"
	ServiceAreaSSO        ServiceArea = "sso"
	ServiceAreaSCIM       ServiceArea = "scim"
	ServiceAreaWebhook    ServiceArea = "webhook"
	ServiceAreaSettings   ServiceArea = "settings"
	ServiceAreaSystem     ServiceArea = "system"
)

type ActorType string

const (
	ActorTypeUser    ActorType = "user"
	ActorTypeToken   ActorType = "token"
	ActorTypeAPIKey  ActorType = "api_key"
	ActorTypeSystem  ActorType = "system"
	ActorTypeSCIM    ActorType = "scim"
)

type ResourceType string

const (
	ResourceTypeUser              ResourceType = "user"
	ResourceTypeTeam              ResourceType = "team"
	ResourceTypeFunction          ResourceType = "function"
	ResourceTypeSecret            ResourceType = "secret"
	ResourceTypeAPIKey            ResourceType = "api_key"
	ResourceTypeApp              ResourceType = "app"
	ResourceTypeBackend           ResourceType = "backend"
	ResourceTypeDeployment        ResourceType = "deployment"
	ResourceTypeSubscription      ResourceType = "subscription"
	ResourceTypeInvoice           ResourceType = "invoice"
	ResourceTypePaymentMethod     ResourceType = "payment_method"
	ResourceTypeAgent             ResourceType = "agent"
	ResourceTypeConnector         ResourceType = "connector"
	ResourceTypeWebhook           ResourceType = "webhook"
	ResourceTypeSSOConfig        ResourceType = "sso_config"
	ResourceTypeSCIMUser         ResourceType = "scim_user"
	ResourceTypeSCIMGroup        ResourceType = "scim_group"
	ResourceTypePolicy            ResourceType = "policy"
	ResourceTypeSettings         ResourceType = "settings"
)

type AuditLog struct {
	ID           uuid.UUID    `json:"id" db:"id"`
	TenantID     uuid.UUID    `json:"tenant_id" db:"tenant_id"`
	ServiceArea  ServiceArea  `json:"service_area" db:"service_area"`
	Action       string       `json:"action" db:"action"`
	ResourceType ResourceType `json:"resource_type" db:"resource_type"`
	ResourceID   *uuid.UUID   `json:"resource_id,omitempty" db:"resource_id"`
	ActorType    ActorType    `json:"actor_type" db:"actor_type"`
	ActorID      string       `json:"actor_id" db:"actor_id"`
	ActorName    string       `json:"actor_name,omitempty" db:"actor_name"`
	RequestID    string       `json:"request_id,omitempty" db:"request_id"`
	IPAddress    string       `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    string       `json:"user_agent,omitempty" db:"user_agent"`
	Metadata     []byte       `json:"metadata,omitempty" db:"metadata"`
	Success      bool         `json:"success" db:"success"`
	ErrorMessage string       `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
}

func (a *AuditLog) GetMetadata() map[string]interface{} {
	if a.Metadata == nil || len(a.Metadata) == 0 {
		return make(map[string]interface{})
	}
	var m map[string]interface{}
	if err := json.Unmarshal(a.Metadata, &m); err != nil {
		return make(map[string]interface{})
	}
	return m
}

func (a *AuditLog) SetMetadata(m map[string]interface{}) {
	if m == nil {
		a.Metadata = []byte("{}")
		return
	}
	b, err := json.Marshal(m)
	if err != nil {
		a.Metadata = []byte("{}")
		return
	}
	a.Metadata = b
}

type ListFilters struct {
	TenantID     uuid.UUID
	ServiceArea  *ServiceArea
	Action       *string
	ResourceType *ResourceType
	ResourceID   *uuid.UUID
	ActorType    *ActorType
	ActorID      *string
	Success      *bool
	StartTime    *time.Time
	EndTime      *time.Time
	Search       *string
	Limit        int
	Offset       int
}

type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatCEF ExportFormat = "cef"
)

type ExportQuery struct {
	TenantID    uuid.UUID
	Format      ExportFormat
	From        time.Time
	To          time.Time
	ServiceArea *ServiceArea
	Action      *string
}

type ExportResult struct {
	Format      ExportFormat
	Body        []byte
	HMAC        string
	Generated   time.Time
	RowCount    int
}

func (et ServiceArea) Valid() bool {
	switch et {
	case ServiceAreaAuth, ServiceAreaVault, ServiceAreaBilling, ServiceAreaFunctions,
		ServiceAreaRegistry, ServiceAreaAgents, ServiceAreaTeams, ServiceAreaAPI,
		ServiceAreaSSO, ServiceAreaSCIM, ServiceAreaWebhook, ServiceAreaSettings, ServiceAreaSystem:
		return true
	}
	return false
}

func (at ActorType) Valid() bool {
	switch at {
	case ActorTypeUser, ActorTypeToken, ActorTypeAPIKey, ActorTypeSystem, ActorTypeSCIM:
		return true
	}
	return false
}

func (rt ResourceType) Valid() bool {
	switch rt {
	case ResourceTypeUser, ResourceTypeTeam, ResourceTypeFunction, ResourceTypeSecret,
		ResourceTypeAPIKey, ResourceTypeApp, ResourceTypeBackend, ResourceTypeDeployment,
		ResourceTypeSubscription, ResourceTypeInvoice, ResourceTypePaymentMethod,
		ResourceTypeAgent, ResourceTypeConnector, ResourceTypeWebhook, ResourceTypeSSOConfig,
		ResourceTypeSCIMUser, ResourceTypeSCIMGroup, ResourceTypePolicy, ResourceTypeSettings:
		return true
	}
	return false
}
