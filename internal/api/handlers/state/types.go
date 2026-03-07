package state

import "github.com/google/uuid"

// Request/Response types

type CreateStateRequest struct {
	Name          string                 `json:"name"`
	StorageType   string                 `json:"storage_type,omitempty"`
	TTLDays       int                    `json:"ttl_days,omitempty"`
	MaxSizeMB     int                    `json:"max_size_mb,omitempty"`
	IsVersioned   bool                   `json:"is_versioned,omitempty"`
	IsEncrypted   bool                   `json:"is_encrypted,omitempty"`
	IsPublic      bool                   `json:"is_public,omitempty"`
	Description   string                 `json:"description,omitempty"`
	Tags          map[string]interface{} `json:"tags,omitempty"`
}

type SetValueRequest struct {
	Value        map[string]interface{} `json:"value"`
	IsEncrypted  bool                   `json:"is_encrypted,omitempty"`
	ExpiresInDays int                   `json:"expires_in_days,omitempty"`
}

type PatchValueRequest struct {
	Patch []map[string]interface{} `json:"patch"`
}

type UpdateStateRequest struct {
	Name          string                 `json:"name,omitempty"`
	Description   string                 `json:"description,omitempty"`
	Tags          map[string]interface{} `json:"tags,omitempty"`
	TTLDays       *int                   `json:"ttl_days,omitempty"`
	MaxSizeMB     *int                   `json:"max_size_mb,omitempty"`
	IsPublic      *bool                  `json:"is_public,omitempty"`
	IsEncrypted   *bool                  `json:"is_encrypted,omitempty"`
}

type CreateSnapshotRequest struct {
	Label string `json:"label,omitempty"`
}

type RestoreSnapshotRequest struct {
	SnapshotVersion int `json:"snapshot_version"`
}

type GrantPermissionRequest struct {
	PrincipalType string    `json:"principal_type"`
	PrincipalID   uuid.UUID `json:"principal_id"`
	CanRead       bool      `json:"can_read"`
	CanWrite      bool      `json:"can_write"`
	CanDelete     bool      `json:"can_delete"`
	CanAdmin      bool      `json:"can_admin"`
	CanTrigger    bool      `json:"can_trigger"`
}

type CreateTriggerRequest struct {
	StatePath                string                 `json:"state_path,omitempty"`
	TriggerType             string                 `json:"trigger_type"`
	KeyPattern              string                 `json:"key_pattern,omitempty"`
	Condition               map[string]interface{} `json:"condition,omitempty"`
	TargetFunctionID        *uuid.UUID             `json:"target_function_id,omitempty"`
	TargetFunction          string                 `json:"target_function"`
	IncludePrevious         bool                   `json:"include_previous"`
	IncludeNew              bool                   `json:"include_new"`
	MaxInvocationsPerMinute int                    `json:"max_invocations_per_minute"`
	IsActive                bool                   `json:"is_active"`
}
