package studio

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/handlers/registry/execution"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// MarketplaceExecutionResult is returned after invoking a registry function.
type MarketplaceExecutionResult struct {
	ExecutionID string          `json:"execution_id"`
	Status      string          `json:"status"`
	Output      json.RawMessage `json:"output,omitempty"`
	DurationMs  int             `json:"duration_ms"`
	Version     string          `json:"version"`
	Error       string          `json:"error,omitempty"`
}

// MarketplaceExecutor runs marketplace functions through the SAR sandbox runtime.
type MarketplaceExecutor struct {
	registry *storageregistry.RegistryRepository
}

// NewMarketplaceExecutor creates an executor backed by the registry repository.
func NewMarketplaceExecutor(registry *storageregistry.RegistryRepository) *MarketplaceExecutor {
	return &MarketplaceExecutor{registry: registry}
}

// Execute invokes a marketplace function and records the execution.
func (e *MarketplaceExecutor) Execute(
	ctx context.Context,
	callerTenantID, callerUserID, functionID string,
	input map[string]interface{},
	callerIP string,
) (*MarketplaceExecutionResult, error) {
	if e == nil || e.registry == nil {
		return nil, fmt.Errorf("marketplace executor is not configured")
	}

	fnUUID, err := uuid.Parse(functionID)
	if err != nil {
		return nil, fmt.Errorf("invalid function id")
	}

	fn, err := e.registry.GetFunctionByID(context.Background(), fnUUID)
	if err != nil {
		return nil, fmt.Errorf("function not found")
	}

	callerTenantUUID, _ := uuid.Parse(callerTenantID)
	if fn.Visibility != "public" && fn.Visibility != "unlisted" {
		if fn.TenantID == nil || *fn.TenantID != callerTenantUUID {
			return nil, fmt.Errorf("function not found")
		}
	}

	latestVersion := ""
	if fn.LatestVersion.Valid {
		latestVersion = fn.LatestVersion.String
	}
	fnVersion, err := e.resolveExecutableVersion(fnUUID, latestVersion)
	if err != nil {
		return nil, err
	}

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshal input: %w", err)
	}

	timeoutMs := fnVersion.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 30000
	}

	start := time.Now()
	executor, err := execution.NewSandboxExecutor()
	if err != nil {
		return nil, fmt.Errorf("create sandbox executor: %w", err)
	}

	output, execErr := executor.ExecuteFunction(fnVersion, inputBytes, timeoutMs)
	durationMs := int(time.Since(start).Milliseconds())

	execID := uuid.New()
	statusCode := http.StatusOK
	outcome := "success"
	var errCode sql.NullString
	if execErr != nil {
		statusCode = http.StatusInternalServerError
		outcome = "error"
		errCode = sql.NullString{String: "execution_failed", Valid: true}
	}

	var callerTenantPtr *uuid.UUID
	if callerTenantUUID != uuid.Nil {
		callerTenantPtr = &callerTenantUUID
	}
	var callerUserPtr *uuid.UUID
	if callerUserUUID, parseErr := uuid.Parse(callerUserID); parseErr == nil {
		callerUserPtr = &callerUserUUID
	}

	record := &storageregistry.RegistryFunctionExecution{
		ID:         execID,
		FunctionID: fnUUID,
		Version:    fnVersion.Version,
		DurationMs: durationMs,
		StatusCode: statusCode,
		Outcome:    outcome,
		ErrorCode:  errCode,
		TenantID:   callerTenantPtr,
		UserID:     callerUserPtr,
		Timestamp:  time.Now(),
	}
	if callerIP != "" {
		record.CallerIP = sql.NullString{String: callerIP, Valid: true}
	}

	if recordErr := e.registry.RecordExecution(context.Background(), record); recordErr != nil {
		logrus.WithError(recordErr).Warn("marketplace: failed to record execution")
	}

	result := &MarketplaceExecutionResult{
		ExecutionID: execID.String(),
		Status:      outcome,
		DurationMs:  durationMs,
		Version:     fnVersion.Version,
	}
	if execErr != nil {
		result.Error = execErr.Error()
		result.Status = "error"
		return result, nil
	}
	result.Output = json.RawMessage(output)
	return result, nil
}

func (e *MarketplaceExecutor) resolveExecutableVersion(functionID uuid.UUID, latestVersion string) (*storageregistry.RegistryFunctionVersion, error) {
	if latestVersion != "" {
		if v, err := e.registry.GetFunctionVersion(functionID, latestVersion); err == nil && len(v.WasmBinary) > 0 {
			return v, nil
		}
	}

	versions, err := e.registry.ListFunctionVersions(functionID)
	if err != nil {
		return nil, fmt.Errorf("list function versions: %w", err)
	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("function has no published versions")
	}

	var selected *storageregistry.RegistryFunctionVersion
	for i := range versions {
		v := versions[i]
		if len(v.WasmBinary) == 0 {
			continue
		}
		if selected == nil || v.Version > selected.Version {
			copy := v
			selected = &copy
		}
	}
	if selected == nil {
		return nil, fmt.Errorf("function has no executable WASM binary")
	}
	return selected, nil
}
