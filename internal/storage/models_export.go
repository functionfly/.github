package storage

import (
	"time"

	"github.com/google/uuid"
)

// UsageExportFormat represents the supported export formats
type UsageExportFormat string

const (
	ExportFormatCSV    UsageExportFormat = "csv"
	ExportFormatJSON   UsageExportFormat = "json"
	ExportFormatParquet UsageExportFormat = "parquet"
	ExportFormatExcel  UsageExportFormat = "excel"
)

// UsageExportStatus represents the status of an export job
type UsageExportStatus string

const (
	ExportStatusPending    UsageExportStatus = "pending"
	ExportStatusProcessing UsageExportStatus = "processing"
	ExportStatusCompleted  UsageExportStatus = "completed"
	ExportStatusFailed     UsageExportStatus = "failed"
	ExportStatusExpired    UsageExportStatus = "expired"
)

// UsageExportConfiguration represents a saved export configuration
type UsageExportConfiguration struct {
	ID                 uuid.UUID              `json:"id"`
	TenantID           uuid.UUID              `json:"tenant_id"`
	Name               string                 `json:"name"`
	Description        string                 `json:"description,omitempty"`
	
	// Export settings
	Format             UsageExportFormat      `json:"format"`
	DataTypes          []string               `json:"data_types"` // usage, costs, executions, forecasts
	Granularity        string                 `json:"granularity"` // raw, hourly, daily, monthly
	IncludeMetadata    bool                   `json:"include_metadata"`
	IncludeBreakdown   bool                   `json:"include_breakdown"` // by-function, by-region, etc.
	
	// Filtering
	DateRangeType      string                 `json:"date_range_type"` // current_period, last_30d, last_90d, custom, rolling_30d
	FunctionFilter     []uuid.UUID            `json:"function_filter,omitempty"`
	RegionFilter       []string               `json:"region_filter,omitempty"`
	OutcomeFilter      []string               `json:"outcome_filter,omitempty"`
	
	// Scheduling
	IsScheduled        bool                   `json:"is_scheduled"`
	ScheduleFrequency  string                 `json:"schedule_frequency,omitempty"` // daily, weekly, monthly
	ScheduleDayOfMonth *int                   `json:"schedule_day_of_month,omitempty"`
	ScheduleDayOfWeek  *int                   `json:"schedule_day_of_week,omitempty"`
	ScheduleHour       *int                   `json:"schedule_hour,omitempty"`
	
	// Delivery
	DeliveryMethod     string                 `json:"delivery_method"` // download, email, s3, webhook
	EmailRecipients    []string               `json:"email_recipients,omitempty"`
	WebhookURL         string                 `json:"webhook_url,omitempty"`
	S3Bucket           string                 `json:"s3_bucket,omitempty"`
	S3Prefix           string                 `json:"s3_prefix,omitempty"`
	
	// Integration settings
	ExternalSystemID   *uuid.UUID             `json:"external_system_id,omitempty"` // Link to external billing system
	FieldMapping       map[string]string      `json:"field_mapping,omitempty"`        // Custom field mappings
	TransformConfig    map[string]interface{} `json:"transform_config,omitempty"`   // Transformation rules
	
	IsActive           bool                   `json:"is_active"`
	CreatedBy          uuid.UUID              `json:"created_by"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	LastExecutedAt     *time.Time             `json:"last_executed_at,omitempty"`
	LastExportID       *uuid.UUID             `json:"last_export_id,omitempty"`
}

// UsageExportJob represents an instance of an export job
type UsageExportJob struct {
	ID                   uuid.UUID              `json:"id"`
	ConfigurationID      uuid.UUID              `json:"configuration_id"`
	TenantID             uuid.UUID              `json:"tenant_id"`
	Status               UsageExportStatus      `json:"status"`
	
	// Export parameters (snapshot of config at execution time)
	Format               UsageExportFormat      `json:"format"`
	DataTypes            []string               `json:"data_types"`
	PeriodStart          time.Time              `json:"period_start"`
	PeriodEnd            time.Time              `json:"period_end"`
	RecordCount          int64                  `json:"record_count"`
	FileSizeBytes        int64                  `json:"file_size_bytes"`
	
	// File storage
	StorageProvider      string                 `json:"storage_provider"` // local, s3, r2
	StoragePath          string                 `json:"storage_path"`
	StorageURL           string                 `json:"storage_url,omitempty"`
	Checksum             string                 `json:"checksum,omitempty"`
	
	// Processing metrics
	StartedAt            *time.Time             `json:"started_at,omitempty"`
	CompletedAt          *time.Time             `json:"completed_at,omitempty"`
	ExpiresAt            *time.Time             `json:"expires_at,omitempty"`
	ErrorMessage         string                 `json:"error_message,omitempty"`
	RetryCount           int                    `json:"retry_count"`
	
	// Delivery tracking
	DeliveredAt          *time.Time             `json:"delivered_at,omitempty"`
	DeliveryMethod       string                 `json:"delivery_method"`
	DeliveryStatus       string                 `json:"delivery_status"`
	DeliveryError        string                 `json:"delivery_error,omitempty"`
	
	CreatedAt            time.Time              `json:"created_at"`
	TriggeredBy          string                 `json:"triggered_by"` // user, schedule, api
}

// ExternalBillingSystem represents a configured external billing integration
type ExternalBillingSystem struct {
	ID                  uuid.UUID              `json:"id"`
	TenantID            uuid.UUID              `json:"tenant_id"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description,omitempty"`
	
	// System type and configuration
	SystemType          string                 `json:"system_type"` // stripe, chargebee, recurly, zuora, net_suite, custom
	APIEndpoint         string                 `json:"api_endpoint,omitempty"`
	AuthType            string                 `json:"auth_type"` // api_key, oauth2, basic_auth
	
	// Encrypted credentials (stored securely)
	APICredentialKey    string                 `json:"-"` // Encrypted, never exposed in JSON
	APICredentialSecret string                 `json:"-"` // Encrypted, never exposed in JSON
	OAuthToken          string                 `json:"-"` // Encrypted
	OAuthRefreshToken   string                 `json:"-"` // Encrypted
	OAuthExpiresAt      *time.Time             `json:"oauth_expires_at,omitempty"`
	
	// Connection settings
	IsActive            bool                   `json:"is_active"`
	LastTestedAt        *time.Time             `json:"last_tested_at,omitempty"`
	LastTestStatus      string                 `json:"last_test_status,omitempty"`
	LastTestError       string                 `json:"last_test_error,omitempty"`
	
	// Sync settings
	SyncEnabled         bool                   `json:"sync_enabled"`
	SyncFrequency       string                 `json:"sync_frequency,omitempty"` // hourly, daily, weekly
	SyncDirection       string                 `json:"sync_direction,omitempty"` // push, pull, bidirectional
	LastSyncAt          *time.Time             `json:"last_sync_at,omitempty"`
	LastSyncStatus      string                 `json:"last_sync_status,omitempty"`
	
	// Field mappings and transformations
	FieldMappings       map[string]string      `json:"field_mappings"`        // local_field -> external_field
	ValueMappings       map[string]interface{} `json:"value_mappings,omitempty"` // field -> {local: external}
	TransformRules      []TransformRule        `json:"transform_rules,omitempty"`
	
	// Event webhooks
	WebhookSecret       string                 `json:"-"` // Encrypted webhook secret
	WebhookURL          string                 `json:"webhook_url,omitempty"`
	
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
	CreatedBy           uuid.UUID              `json:"created_by"`
}

// TransformRule represents a data transformation rule for external integrations
type TransformRule struct {
	Field       string                 `json:"field"`
	Operation   string                 `json:"operation"` // multiply, divide, add, subtract, format, map
	Value       interface{}            `json:"value,omitempty"`
	Format      string                 `json:"format,omitempty"` // for date/number formatting
	Conditions  []TransformCondition   `json:"conditions,omitempty"`
}

// TransformCondition represents a conditional transformation
type TransformCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // eq, neq, gt, gte, lt, lte, contains, starts_with, ends_with
	Value    interface{} `json:"value"`
	Then     interface{} `json:"then"`
	Else     interface{} `json:"else,omitempty"`
}

// BillingIntegrationSync represents a sync operation with an external billing system
type BillingIntegrationSync struct {
	ID                    uuid.UUID              `json:"id"`
	ExternalSystemID      uuid.UUID              `json:"external_system_id"`
	TenantID              uuid.UUID              `json:"tenant_id"`
	
	// Sync configuration
	SyncType              string                 `json:"sync_type"` // usage, invoices, customers, payments, all
	Direction             string                 `json:"direction"` // push, pull
	
	// Status and results
	Status                string                 `json:"status"` // pending, running, completed, failed, partial
	StartedAt             *time.Time             `json:"started_at,omitempty"`
	CompletedAt           *time.Time             `json:"completed_at,omitempty"`
	
	// Statistics
	RecordsProcessed      int64                  `json:"records_processed"`
	RecordsCreated        int64                  `json:"records_created"`
	RecordsUpdated        int64                  `json:"records_updated"`
	RecordsFailed         int64                  `json:"records_failed"`
	RecordsSkipped        int64                  `json:"records_skipped"`
	
	// Error tracking
	ErrorMessage          string                 `json:"error_message,omitempty"`
	ErrorDetails          []SyncErrorDetail      `json:"error_details,omitempty"`
	
	// External references
	ExternalBatchID       string                 `json:"external_batch_id,omitempty"`
	ExternalReferences    map[string]string      `json:"external_references,omitempty"`
	
	CreatedAt             time.Time              `json:"created_at"`
	TriggeredBy           string                 `json:"triggered_by"` // schedule, manual, webhook, api
}

// SyncErrorDetail represents an individual error during sync
type SyncErrorDetail struct {
	RecordID      string                 `json:"record_id"`
	RecordType    string                 `json:"record_type"`
	ErrorCode     string                 `json:"error_code"`
	ErrorMessage  string                 `json:"error_message"`
	Context       map[string]interface{} `json:"context,omitempty"`
}

// UsageExportTemplate represents a predefined export template
type UsageExportTemplate struct {
	ID              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description,omitempty"`
	Category        string                 `json:"category"` // financial, operational, compliance, custom
	
	// Template configuration
	Format          UsageExportFormat      `json:"format"`
	DataTypes       []string               `json:"data_types"`
	Granularity     string                 `json:"granularity"`
	IncludeMetadata bool                   `json:"include_metadata"`
	IncludeBreakdown bool                  `json:"include_breakdown"`
	DefaultFields   []string               `json:"default_fields"`
	FieldOrder      []string               `json:"field_order,omitempty"`
	
	// Transformations
	ColumnHeaders   map[string]string      `json:"column_headers,omitempty"` // field -> display name
	DataTransforms  []TransformRule        `json:"data_transforms,omitempty"`
	
	IsActive        bool                   `json:"is_active"`
	IsSystem        bool                   `json:"is_system"` // System templates can't be deleted
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}
