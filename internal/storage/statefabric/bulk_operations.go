package statefabric

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/functionfly/functionfly/internal/monitoring"
)

type BulkFabricRequest struct {
	Fabrics    []BulkFabricEntry `json:"fabrics"`
	TenantID   uuid.UUID         `json:"tenant_id"`
	ImportedBy uuid.UUID          `json:"imported_by"`
	SkipErrors bool              `json:"skip_errors"`
}

type BulkFabricEntry struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Settings    map[string]interface{} `json:"settings"`
}

type BulkFabricResult struct {
	TotalSubmitted int              `json:"total_submitted"`
	Imported       int              `json:"imported"`
	Skipped        int              `json:"skipped"`
	Failed         int              `json:"failed"`
	Errors         []BulkFabricError `json:"errors,omitempty"`
	ImportedIDs    []uuid.UUID      `json:"imported_ids"`
	DurationMs     int64             `json:"duration_ms"`
}

type BulkFabricError struct {
	Index int    `json:"index"`
	Name  string `json:"name,omitempty"`
	ID    string `json:"id,omitempty"`
	Error string `json:"error"`
}

type BulkKeyRequest struct {
	Keys       []BulkKeyEntry `json:"keys"`
	TenantID   uuid.UUID      `json:"tenant_id"`
	FabricID   uuid.UUID      `json:"fabric_id"`
	SkipErrors bool           `json:"skip_errors"`
}

type BulkKeyEntry struct {
	Key   string                 `json:"key"`
	Value map[string]interface{} `json:"value"`
	TTL   int                    `json:"ttl_days,omitempty"`
}

type BulkKeyResult struct {
	TotalSubmitted int            `json:"total_submitted"`
	Processed      int            `json:"processed"`
	Deleted        int            `json:"deleted"`
	Skipped        int            `json:"skipped"`
	Failed         int            `json:"failed"`
	Errors         []BulkKeyError `json:"errors,omitempty"`
	DurationMs     int64          `json:"duration_ms"`
}

type BulkKeyError struct {
	Index int    `json:"index"`
	Key   string `json:"key,omitempty"`
	Error string `json:"error"`
}

type BulkFabricUpdateRequest struct {
	ID      string                 `json:"id"`
	Updates map[string]interface{} `json:"updates"`
}

// BulkCreateFabrics creates multiple fabrics in a single operation
func (r *Repository) BulkCreateFabrics(ctx context.Context, req BulkFabricRequest) (*BulkFabricResult, error) {
	start := time.Now()
	result := &BulkFabricResult{
		TotalSubmitted: len(req.Fabrics),
		ImportedIDs:    make([]uuid.UUID, 0),
		Errors:         make([]BulkFabricError, 0),
	}

	monitoring.RecordStateFabricBulkOperation(req.TenantID.String(), "bulk_create", "started")

	for idx, entry := range req.Fabrics {
		if err := validateBulkFabricEntry(&entry); err != nil {
			result.Skipped++
			if req.SkipErrors {
				result.Errors = append(result.Errors, BulkFabricError{
					Index: idx,
					Name:  entry.Name,
					Error: err.Error(),
				})
			}
			continue
		}

		fabric, err := r.CreateFabric(ctx, req.TenantID, entry.Name, entry.Description, entry.Type, entry.Settings)
		if err != nil {
			result.Failed++
			if req.SkipErrors {
				result.Errors = append(result.Errors, BulkFabricError{
					Index: idx,
					Name:  entry.Name,
					Error: err.Error(),
				})
			}
			monitoring.RecordStateFabricBulkOperation(req.TenantID.String(), "bulk_create", "failed")
			continue
		}

		result.Imported++
		result.ImportedIDs = append(result.ImportedIDs, fabric.ID)
	}

	result.DurationMs = time.Since(start).Milliseconds()

	status := "success"
	if result.Failed > 0 {
		status = "partial"
	}
	monitoring.RecordStateFabricBulkOperation(req.TenantID.String(), "bulk_create", status)
	monitoring.RecordStateFabricBulkOperationDuration(req.TenantID.String(), "bulk_create", time.Since(start))

	logrus.WithFields(logrus.Fields{
		"tenant_id":   req.TenantID,
		"total":       result.TotalSubmitted,
		"imported":    result.Imported,
		"skipped":     result.Skipped,
		"failed":      result.Failed,
		"duration_ms": result.DurationMs,
	}).Info("Bulk fabric creation completed")

	return result, nil
}

// validateBulkFabricEntry validates a bulk fabric entry
func validateBulkFabricEntry(entry *BulkFabricEntry) error {
	if strings.TrimSpace(entry.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if len(entry.Name) > 255 {
		return fmt.Errorf("name must be at most 255 characters")
	}
	validTypes := map[string]bool{"cache": true, "catalog": true, "workflow": true, "custom": true}
	if entry.Type != "" && !validTypes[entry.Type] {
		return fmt.Errorf("invalid type: %s", entry.Type)
	}
	return nil
}

// BulkUpdateFabrics updates multiple fabrics in a single operation
func (r *Repository) BulkUpdateFabrics(ctx context.Context, tenantID uuid.UUID, updates []BulkFabricUpdateRequest) (*BulkFabricResult, error) {
	start := time.Now()
	result := &BulkFabricResult{
		TotalSubmitted: len(updates),
		ImportedIDs:    make([]uuid.UUID, 0),
		Errors:         make([]BulkFabricError, 0),
	}

	monitoring.RecordStateFabricBulkOperation(tenantID.String(), "bulk_update", "started")

	for idx, update := range updates {
		fabricID, err := uuid.Parse(update.ID)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, BulkFabricError{
				Index: idx,
				ID:    update.ID,
				Error: "invalid fabric ID",
			})
			continue
		}

		fabric, err := r.UpdateFabric(ctx, tenantID, fabricID, update.Updates)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, BulkFabricError{
				Index: idx,
				ID:    update.ID,
				Error: err.Error(),
			})
			continue
		}

		result.Imported++
		result.ImportedIDs = append(result.ImportedIDs, fabric.ID)
	}

	result.DurationMs = time.Since(start).Milliseconds()

	status := "success"
	if result.Failed > 0 {
		status = "partial"
	}
	monitoring.RecordStateFabricBulkOperation(tenantID.String(), "bulk_update", status)
	monitoring.RecordStateFabricBulkOperationDuration(tenantID.String(), "bulk_update", time.Since(start))

	return result, nil
}

// BulkDeleteFabrics deletes multiple fabrics in a single operation
func (r *Repository) BulkDeleteFabrics(ctx context.Context, tenantID uuid.UUID, fabricIDs []string) (*BulkFabricResult, error) {
	start := time.Now()
	result := &BulkFabricResult{
		TotalSubmitted: len(fabricIDs),
		Errors:         make([]BulkFabricError, 0),
	}

	monitoring.RecordStateFabricBulkOperation(tenantID.String(), "bulk_delete", "started")

	for idx, fabricIDStr := range fabricIDs {
		fabricID, err := uuid.Parse(fabricIDStr)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, BulkFabricError{
				Index: idx,
				ID:    fabricIDStr,
				Error: "invalid fabric ID",
			})
			continue
		}

		if err := r.DeleteFabric(ctx, tenantID, fabricID); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, BulkFabricError{
				Index: idx,
				ID:    fabricIDStr,
				Error: err.Error(),
			})
			continue
		}

		result.Imported++
	}

	result.DurationMs = time.Since(start).Milliseconds()

	status := "success"
	if result.Failed > 0 {
		status = "partial"
	}
	monitoring.RecordStateFabricBulkOperation(tenantID.String(), "bulk_delete", status)
	monitoring.RecordStateFabricBulkOperationDuration(tenantID.String(), "bulk_delete", time.Since(start))

	return result, nil
}

// BulkSetKeys sets multiple key-value pairs in a fabric
func (r *Repository) BulkSetKeys(ctx context.Context, req BulkKeyRequest) (*BulkKeyResult, error) {
	start := time.Now()
	result := &BulkKeyResult{
		TotalSubmitted: len(req.Keys),
		Errors:         make([]BulkKeyError, 0),
	}

	monitoring.RecordStateFabricBulkOperation(req.TenantID.String(), "bulk_key_set", "started")

	_, err := r.GetFabric(ctx, req.TenantID, req.FabricID)
	if err != nil {
		return nil, err
	}

	for idx, entry := range req.Keys {
		keyStart := time.Now()
		err := r.SetFabricValue(ctx, req.TenantID, req.FabricID, entry.Key, entry.Value, "bulk")
		duration := time.Since(keyStart)

		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, BulkKeyError{
				Index: idx,
				Key:   entry.Key,
				Error: err.Error(),
			})
			monitoring.RecordStateFabricKeyOperation(req.TenantID.String(), req.FabricID.String(), "bulk_set", "failed")
		} else {
			result.Processed++
			monitoring.RecordStateFabricKeyOperation(req.TenantID.String(), req.FabricID.String(), "bulk_set", "success")
		}

		monitoring.RecordStateFabricKeyOperationDuration(req.TenantID.String(), req.FabricID.String(), "bulk_set", duration)
	}

	result.DurationMs = time.Since(start).Milliseconds()

	status := "success"
	if result.Failed > 0 {
		status = "partial"
	}
	monitoring.RecordStateFabricBulkOperation(req.TenantID.String(), "bulk_key_set", status)
	monitoring.RecordStateFabricBulkOperationDuration(req.TenantID.String(), "bulk_key_set", time.Since(start))

	logrus.WithFields(logrus.Fields{
		"tenant_id":   req.TenantID,
		"fabric_id":   req.FabricID,
		"total":       result.TotalSubmitted,
		"processed":   result.Processed,
		"skipped":     result.Skipped,
		"failed":      result.Failed,
		"duration_ms": result.DurationMs,
	}).Info("Bulk key set completed")

	return result, nil
}

// BulkDeleteKeys deletes multiple keys from a fabric
func (r *Repository) BulkDeleteKeys(ctx context.Context, req BulkKeyRequest) (*BulkKeyResult, error) {
	start := time.Now()
	result := &BulkKeyResult{
		TotalSubmitted: len(req.Keys),
		Errors:         make([]BulkKeyError, 0),
	}

	monitoring.RecordStateFabricBulkOperation(req.TenantID.String(), "bulk_key_delete", "started")

	_, err := r.GetFabric(ctx, req.TenantID, req.FabricID)
	if err != nil {
		return nil, err
	}

	for idx, entry := range req.Keys {
		err := r.DeleteFabricValue(ctx, req.TenantID, req.FabricID, entry.Key, "bulk")
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, BulkKeyError{
				Index: idx,
				Key:   entry.Key,
				Error: err.Error(),
			})
			monitoring.RecordStateFabricKeyOperation(req.TenantID.String(), req.FabricID.String(), "bulk_delete", "failed")
		} else {
			result.Deleted++
			monitoring.RecordStateFabricKeyOperation(req.TenantID.String(), req.FabricID.String(), "bulk_delete", "success")
		}
	}

	result.DurationMs = time.Since(start).Milliseconds()

	status := "success"
	if result.Failed > 0 {
		status = "partial"
	}
	monitoring.RecordStateFabricBulkOperation(req.TenantID.String(), "bulk_key_delete", status)

	return result, nil
}