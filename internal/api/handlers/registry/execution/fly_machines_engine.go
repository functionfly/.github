package execution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/fly"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type FlyMachinesEngine struct {
	client       *fly.FlyMachinesClient
	microvmRepo *storage.MicroVMRepository
	imageName   string
}

func NewFlyMachinesEngine(client *fly.FlyMachinesClient, microvmRepo *storage.MicroVMRepository) *FlyMachinesEngine {
	imageName := os.Getenv("FLY_ENTERPRISE_IMAGE")
	if imageName == "" {
		imageName = "registry.fly.io/functionfly-enterprise:latest"
	}
	return &FlyMachinesEngine{
		client:     client,
		microvmRepo: microvmRepo,
		imageName: imageName,
	}
}

func (e *FlyMachinesEngine) Execute(ctx context.Context, req ExecutionRequest) (ExecutionResult, error) {
	fnVersion := req.FunctionVersion
	if fnVersion == nil {
		return ExecutionResult{}, fmt.Errorf("function version is required")
	}

	sourceCode := fnVersion.SourceCode.String
	if sourceCode == "" {
		return ExecutionResult{}, fmt.Errorf("function has no source code to execute")
	}

	orchestratorURL := os.Getenv("FUNCTIONFLY_ORCHESTRATOR_URL")
	if orchestratorURL == "" {
		orchestratorURL = "http://localhost:8080"
	}

	machineSecret := os.Getenv("FUNCTIONFLY_MACHINE_SECRET")
	executionID := uuid.New().String()
	tenantID := req.TenantID

	machineKind := "shared"
	cpus := 2
	memoryMB := req.MaxMemoryMB

	if req.Tier == plans.PlanMicroVMEnterprise || req.Tier == "enterprise" {
		machineKind = "performance"
		cpus = 4
	}

	if memoryMB == 0 {
		memoryMB = plans.EnterpriseDefaultMemoryMB
	}

	timeoutMs := req.MaxCPUTimeMs
	if timeoutMs == 0 {
		timeoutMs = plans.EnterpriseDefaultTimeoutMs
	}

	envVars := map[string]string{
		"FUNCTIONFLY_ORCHESTRATOR_URL": orchestratorURL,
		"FUNCTIONFLY_MACHINE_SECRET":   machineSecret,
		"FLY_EXECUTION_ID":            executionID,
		"FLY_TIMEOUT_SECONDS":         fmt.Sprintf("%d", timeoutMs/1000),
	}

	metadata := map[string]string{
		"ff_customer_id":  tenantID,
		"ff_execution_id": executionID,
		"ff_plan":         req.Tier,
		"ff_created_at":   time.Now().UTC().Format(time.RFC3339),
	}

	createReq := &fly.CreateMachineRequest{
		Config: fly.MachineConfig{
			Image: e.imageName,
			Guest: fly.GuestConfig{
				CPUKind:  machineKind,
				CPUs:     cpus,
				MemoryMB: memoryMB,
			},
			Metadata: metadata,
			Env:      envVars,
		},
	}

	logrus.WithFields(logrus.Fields{
		"execution_id": executionID,
		"tenant_id":    tenantID,
		"tier":         req.Tier,
		"memory_mb":    memoryMB,
		"cpus":         cpus,
	}).Info("Creating Fly Machine for execution")

	machine, err := e.client.CreateMachine(ctx, createReq)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("failed to create Fly Machine: %w", err)
	}

	if e.microvmRepo != nil {
		microvmExec := &storage.MicroVMExecution{
			ID:              uuid.New(),
			TenantID:        uuid.MustParse(tenantID),
			FunctionID:      fnVersion.FunctionID,
			FunctionVersion: fnVersion.Version,
			ExecutionID:    uuid.MustParse(executionID),
			FlyMachineID:   machine.ID,
			StartedAt:      time.Now(),
			MemoryMB:        memoryMB,
			VCPUs:           cpus,
			Status:          "starting",
			CreatedAt:       time.Now(),
		}
		if err := e.microvmRepo.CreateExecution(ctx, microvmExec); err != nil {
			logrus.WithError(err).Error("Failed to create MicroVM execution record")
		}
	}

	logrus.WithFields(logrus.Fields{
		"machine_id":   machine.ID,
		"execution_id": executionID,
		"region":       machine.Region,
	}).Info("Fly Machine created, waiting for start")

	if err := e.client.WaitForMachine(ctx, machine.ID, "started", 10*time.Second); err != nil {
		e.client.StopMachine(ctx, machine.ID)
		e.client.DeleteMachine(ctx, machine.ID)
		return ExecutionResult{}, fmt.Errorf("machine failed to start: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"machine_id":   machine.ID,
		"execution_id": executionID,
	}).Debug("Fly Machine started, executing")

	inputJSON, _ := json.Marshal(map[string]interface{}{
		"input": string(req.Input),
	})

	result, err := e.client.ExecInMachine(ctx, machine.ID, "", bytes.NewReader(inputJSON))
	if err != nil {
		e.client.StopMachine(ctx, machine.ID)
		e.client.DeleteMachine(ctx, machine.ID)
		return ExecutionResult{}, fmt.Errorf("failed to exec in machine: %w", err)
	}

	e.client.StopMachine(ctx, machine.ID)
	e.client.DeleteMachine(ctx, machine.ID)

	logrus.WithFields(logrus.Fields{
		"machine_id":   machine.ID,
		"execution_id": executionID,
		"result_len":   len(result),
	}).Debug("Fly Machine execution completed")

	var output json.RawMessage
	if len(result) > 0 {
		var execResult struct {
			Result   string `json:"result"`
			ExitCode int    `json:"exit_code"`
			Error    string `json:"error,omitempty"`
		}
		if err := json.Unmarshal(result, &execResult); err == nil {
			if execResult.ExitCode != 0 {
				return ExecutionResult{}, fmt.Errorf("execution failed: %s", execResult.Error)
			}
			output = json.RawMessage(execResult.Result)
		} else {
			output = result
		}
	}

	return ExecutionResult{
		Output:     output,
		DurationMs: int(timeoutMs),
	}, nil
}

func (e *FlyMachinesEngine) Healthy(ctx context.Context) bool {
	machines, err := e.client.ListMachines(ctx)
	if err != nil {
		return false
	}
	return len(machines) >= 0
}

func (e *FlyMachinesEngine) Close() error {
	return nil
}
