package storage

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type MicroVMExecution struct {
	ID               uuid.UUID      `json:"id" db:"id"`
	TenantID         uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	FunctionID       uuid.UUID      `json:"function_id" db:"function_id"`
	FunctionVersion  string         `json:"function_version" db:"function_version"`
	ExecutionID      uuid.UUID      `json:"execution_id" db:"execution_id"`
	FlyMachineID     string         `json:"fly_machine_id" db:"fly_machine_id"`
	StartedAt        time.Time      `json:"started_at" db:"started_at"`
	CompletedAt      sql.NullTime   `json:"completed_at" db:"completed_at"`
	DurationMs       int            `json:"duration_ms" db:"duration_ms"`
	MemoryMB         int            `json:"memory_mb" db:"memory_mb"`
	VCPUs            int            `json:"vcpus" db:"vcpus"`
	Status           string         `json:"status" db:"status"`
	Outcome          sql.NullString `json:"outcome" db:"outcome"`
	ErrorMessage     sql.NullString `json:"error_message" db:"error_message"`
	NetworkAllowed   bool           `json:"network_allowed" db:"network_allowed"`
	PackagesCached   bool           `json:"packages_cached" db:"packages_cached"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
}

type MicroVMBillingRecord struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	TenantID           uuid.UUID `json:"tenant_id" db:"tenant_id"`
	BillingPeriod      string    `json:"billing_period" db:"billing_period"`
	TotalExecutions    int       `json:"total_executions" db:"total_executions"`
	TotalComputeSeconds float64  `json:"total_compute_seconds" db:"total_compute_seconds"`
	TotalMemorySeconds  float64  `json:"total_memory_seconds" db:"total_memory_seconds"`
	AvgMemoryMB        int       `json:"avg_memory_mb" db:"avg_memory_mb"`
	AvgVCPUs           float64   `json:"avg_vcpus" db:"avg_vcpus"`
	BaseFeeCents       int       `json:"base_fee_cents" db:"base_fee_cents"`
	ComputeChargeCents int       `json:"compute_charge_cents" db:"compute_charge_cents"`
	MemoryChargeCents  int       `json:"memory_charge_cents" db:"memory_charge_cents"`
	TotalChargeCents   int       `json:"total_charge_cents" db:"total_charge_cents"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type MicroVMAuditLog struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	TenantID     uuid.UUID      `json:"tenant_id" db:"tenant_id"`
	UserID       uuid.NullUUID  `json:"user_id" db:"user_id"`
	Action       string         `json:"action" db:"action"`
	ResourceType string         `json:"resource_type" db:"resource_type"`
	ResourceID   uuid.NullUUID  `json:"resource_id" db:"resource_id"`
	Details      json.RawMessage `json:"details" db:"details"`
	IPAddress    sql.NullString `json:"ip_address" db:"ip_address"`
	UserAgent    sql.NullString `json:"user_agent" db:"user_agent"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
}

type MicroVMTenantQuota struct {
	TenantID            uuid.UUID `json:"tenant_id" db:"tenant_id"`
	MaxConcurrentVMs   int       `json:"max_concurrent_vms" db:"max_concurrent_vms"`
	MaxMemoryMB        int       `json:"max_memory_mb" db:"max_memory_mb"`
	MaxVCPUs           int       `json:"max_vcpus" db:"max_vcpus"`
	MaxTimeoutMs       int       `json:"max_timeout_ms" db:"max_timeout_ms"`
	CurrentComputeUsage float64   `json:"current_compute_usage" db:"current_compute_usage"`
	CurrentMemoryUsage float64   `json:"current_memory_usage" db:"current_memory_usage"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type MicroVMUsageSummary struct {
	TenantID           uuid.UUID `json:"tenant_id"`
	TotalExecutions    int       `json:"total_executions"`
	ActiveVMs          int       `json:"active_vms"`
	TotalComputeSeconds float64  `json:"total_compute_seconds"`
	TotalMemorySeconds  float64  `json:"total_memory_seconds"`
	AvgDurationMs      int       `json:"avg_duration_ms"`
	CurrentPeriod      string    `json:"current_period"`
}

type MicroVMStats struct {
	TotalExecutions    int     `json:"total_executions"`
	RunningVMs         int     `json:"running_vms"`
	AvgDurationMs      int     `json:"avg_duration_ms"`
	SuccessRate        float64 `json:"success_rate"`
	TotalComputeSeconds float64 `json:"total_compute_seconds"`
}
