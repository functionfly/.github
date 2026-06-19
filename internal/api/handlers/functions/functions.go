package functions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/api/types"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/bundler"
	deployPkg "github.com/functionfly/functionfly/internal/deployment"
	"github.com/functionfly/functionfly/internal/flypy"
	"github.com/functionfly/functionfly/internal/manifest"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/wasm"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler contains function management handlers
type Handler struct {
	repo          storage.Repository
	deploySvc     *deployPkg.Orchestrator
	pasteHandler  *PasteHandler
}

// NewHandler creates a new functions handler
func NewHandler(repo storage.Repository, deploySvc *deployPkg.Orchestrator, pasteHandler *PasteHandler) *Handler {
	return &Handler{
		repo:         repo,
		deploySvc:    deploySvc,
		pasteHandler: pasteHandler,
	}
}

// HandleListFunctions handles GET /v1/functions
func (h *Handler) HandleListFunctions(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	functions, err := h.repo.ListFunctionsByTenant(r.Context(), user.TenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list functions")
		apierror.WriteError(w, apierror.NewInternal("Failed to list functions"))
		return
	}

	response := map[string]interface{}{
		"functions": functions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetFunction handles GET /v1/functions/{id}
func (h *Handler) HandleGetFunction(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	function, err := h.repo.GetFunctionByID(r.Context(), functionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get function")
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	// Check if function belongs to user's tenant
	if function.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(function)
}

// HandleCreateFunction handles POST /v1/functions
func (h *Handler) HandleCreateFunction(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req types.CreateFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	if req.Name == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Function name is required"))
		return
	}

	if len(req.Providers) == 0 {
		apierror.WriteError(w, apierror.NewBadRequest("At least one provider is required"))
		return
	}

	function := &storage.FunctionConfig{
		TenantID:  user.TenantID,
		Name:      req.Name,
		Providers: req.Providers,
		Region:    req.Region,
		Code:      req.Code,
		EnvVars:   req.EnvVars,
		Status:    "draft",
	}

	createdFunction, err := h.repo.CreateFunction(r.Context(), function)
	if err != nil {
		logrus.WithError(err).Error("Failed to create function")
		apierror.WriteError(w, apierror.NewInternal("Failed to create function"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(types.CreateFunctionResponse{
		FunctionID: createdFunction.ID.String(),
	})
}

// HandleUpdateFunction handles PUT /v1/functions/{id}
func (h *Handler) HandleUpdateFunction(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	var req types.UpdateFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Check if function exists and belongs to user
	existingFunction, err := h.repo.GetFunctionByID(r.Context(), functionID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if existingFunction.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Providers != nil {
		updates["providers"] = req.Providers
	}
	if req.Region != nil {
		updates["region"] = *req.Region
	}
	if req.Code != nil {
		updates["code"] = *req.Code
	}
	if req.EnvVars != nil {
		updates["env_vars"] = req.EnvVars
	}

	updatedFunction, err := h.repo.UpdateFunction(r.Context(), functionID, updates)
	if err != nil {
		logrus.WithError(err).Error("Failed to update function")
		apierror.WriteError(w, apierror.NewInternal("Failed to update function"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedFunction)
}

// HandleDeleteFunction handles DELETE /v1/functions/{id}
func (h *Handler) HandleDeleteFunction(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	// Check if function exists and belongs to user
	existingFunction, err := h.repo.GetFunctionByID(r.Context(), functionID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if existingFunction.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	err = h.repo.DeleteFunction(r.Context(), functionID)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete function")
		apierror.WriteError(w, apierror.NewInternal("Failed to delete function"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleDeployFunction handles POST /v1/functions/deploy
func (h *Handler) HandleDeployFunction(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req types.DeployFunctionRequest // Declare request variable
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// Validate function belongs to user
	function, err := h.repo.GetFunctionByID(r.Context(), req.FunctionId)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if function.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	// Get backend information
	backendID, err := uuid.Parse(req.BackendID)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid backend ID format"))
		return
	}
	backend, err := h.repo.GetBackendByID(r.Context(), backendID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get backend")
		apierror.WriteError(w, apierror.NewNotFound("Backend not found"))
		return
	}

	// Verify backend belongs to same tenant
	if function.AppID != nil && backend.AppID != *function.AppID {
		apierror.WriteError(w, apierror.NewBadRequest("Backend does not belong to the same app as the function"))
		return
	}

	// Create initial deployment record
	deployment := &storage.FunctionDeployment{
		FunctionID: req.FunctionId,
		Version:    req.Version,
		Status:     "pending",
		Provider:   backend.Provider,
		Region:     backend.Region,
	}

	if req.Version == "" {
		deployment.Version = function.Version
	}

	createdDeployment, err := h.repo.CreateFunctionDeployment(r.Context(), deployment)
	if err != nil {
		logrus.WithError(err).Error("Failed to create deployment")
		apierror.WriteError(w, apierror.NewInternal("Failed to create deployment"))
		return
	}

	// Default to production if not specified
	environment := req.Environment
	if environment == "" {
		environment = "prod"
	}

	// Trigger actual deployment asynchronously
	go h.deployFunctionAsync(r.Context(), function, backend, createdDeployment, user.TenantID, environment)

	// Build the deployment URL
	deploymentURL := fmt.Sprintf("https://%s.%s.functionfly.app", function.Name, backend.Region)

	response := types.DeployFunctionResponse{
		FunctionID:   function.ID.String(),
		DeploymentID: createdDeployment.ID.String(),
		URL:         deploymentURL,
		Region:      backend.Region,
		Providers:   []string{backend.Provider},
		Status:      "pending",
		Deployments: []*storage.FunctionDeployment{createdDeployment},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleTestFunction handles POST /v1/functions/test
func (h *Handler) HandleTestFunction(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	var req types.TestFunctionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	// If functionId is provided, validate it belongs to user
	if req.FunctionId != nil {
		function, err := h.repo.GetFunctionByID(r.Context(), *req.FunctionId)
		if err != nil {
			apierror.WriteError(w, apierror.NewNotFound("Function not found"))
			return
		}

		if function.TenantID != user.TenantID {
			apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
			return
		}
	}

	// Get full user record for execution
	fullUser, err := h.repo.GetUserByID(r.Context(), user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get user")
		apierror.WriteError(w, apierror.NewNotFound("User not found"))
		return
	}

	// Execute the function and return the result
	response, err := h.executeTestFunction(r.Context(), &req, fullUser)
	if err != nil {
		logrus.WithError(err).Error("Failed to execute test function")
		apierror.WriteError(w, apierror.NewInternal("failed to execute function"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetFunctionLogs handles GET /v1/functions/{id}/logs
func (h *Handler) HandleGetFunctionLogs(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	// Check if function belongs to user
	function, err := h.repo.GetFunctionByID(r.Context(), functionID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if function.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	// Parse query parameters
	limit := 50 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	logs, err := h.repo.GetFunctionLogs(r.Context(), &functionID, nil, limit, nil, nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to get function logs")
		apierror.WriteError(w, apierror.NewInternal("Failed to get function logs"))
		return
	}

	response := map[string]interface{}{
		"logs": logs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetFunctionDeployments handles GET /v1/functions/{id}/deployments
func (h *Handler) HandleGetFunctionDeployments(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	functionID, err := uuid.Parse(vars["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid function ID"))
		return
	}

	// Check if function belongs to user
	function, err := h.repo.GetFunctionByID(r.Context(), functionID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if function.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	// Parse query parameters
	limit := 20 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	deployments, err := h.repo.ListFunctionDeployments(r.Context(), functionID, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to get function deployments")
		apierror.WriteError(w, apierror.NewInternal("Failed to get function deployments"))
		return
	}

	response := map[string]interface{}{
		"deployments": deployments,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleGetFunctionDeployment handles GET /v1/functions/deployments/{deploymentId}
func (h *Handler) HandleGetFunctionDeployment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	vars := mux.Vars(r)
	deploymentID, err := uuid.Parse(vars["deploymentId"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid deployment ID"))
		return
	}

	deployment, err := h.repo.GetFunctionDeploymentByID(r.Context(), deploymentID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get function deployment")
		apierror.WriteError(w, apierror.NewInternal("Failed to get function deployment"))
		return
	}
	if deployment == nil {
		apierror.WriteError(w, apierror.NewNotFound("Function deployment not found"))
		return
	}

	// Verify user has access to the function
	function, err := h.repo.GetFunctionByID(r.Context(), deployment.FunctionID)
	if err != nil {
		apierror.WriteError(w, apierror.NewNotFound("Function not found"))
		return
	}

	if function.TenantID != user.TenantID {
		apierror.WriteError(w, apierror.NewForbidden("Forbidden"))
		return
	}

	response := map[string]interface{}{
		"deployment": deployment,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// deployFunctionAsync performs the actual function deployment asynchronously
func (h *Handler) deployFunctionAsync(ctx context.Context, function *storage.FunctionConfig, backend *storage.Backend, deployment *storage.FunctionDeployment, tenantID uuid.UUID, environment string) {
	logrus.WithFields(logrus.Fields{
		"function_id":   function.ID,
		"deployment_id": deployment.ID,
		"backend_id":    backend.ID,
	}).Info("Starting function deployment")

	// Update deployment status to deploying
	if err := h.repo.UpdateFunctionDeploymentStatus(ctx, deployment.ID, "deploying", nil, stringPtr("Starting deployment")); err != nil {
		logrus.WithError(err).WithField("deployment_id", deployment.ID).Error("Failed to update deployment status to deploying")
		return
	}

	// Create temporary directory for bundling
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("function-deploy-%s", function.ID.String()))
	if err != nil {
		logrus.WithError(err).WithField("deployment_id", deployment.ID).Error("Failed to create temp directory for bundling")
		h.updateDeploymentStatus(ctx, deployment.ID, "failed", nil, fmt.Sprintf("Failed to create temp directory: %v", err))
		return
	}
	defer os.RemoveAll(tempDir)

	// Write function code to a file
	funcFile := filepath.Join(tempDir, "function.py")
	if err := os.WriteFile(funcFile, []byte(function.Code), 0644); err != nil {
		logrus.WithError(err).WithField("deployment_id", deployment.ID).Error("Failed to write function code to temp file")
		h.updateDeploymentStatus(ctx, deployment.ID, "failed", nil, fmt.Sprintf("Failed to write function code: %v", err))
		return
	}

	// Create manifest for bundling
	deterministic := function.Status == "deterministic"
	idempotent := determineSideEffects(function.Capabilities) == "none"
	funcManifest := &manifest.Manifest{
		Name:          function.Name,
		Version:       deployment.Version,
		Runtime:       "python3.11",
		Deterministic: &deterministic,
		InputSchema:   map[string]interface{}{},
		OutputSchema:  map[string]interface{}{},
		Idempotent:    &idempotent,
		SideEffects:   determineSideEffects(function.Capabilities),
		Capabilities:  function.Capabilities,
		MainFile:      "function.py",
	}

	// Bundle the function code from the temp directory
	artifact, err := bundler.BundleWithWorkingDirectory(funcManifest, tempDir)
	if err != nil {
		logrus.WithError(err).WithField("deployment_id", deployment.ID).Error("Failed to bundle function")
		h.updateDeploymentStatus(ctx, deployment.ID, "failed", nil, fmt.Sprintf("Bundling failed: %v", err))
		return
	}

	// Prepare environment variables and secrets
	envVars := make(map[string]string)
	secrets := make(map[string]string)
	for _, envVar := range function.EnvVars {
		if envVar.IsSecret {
			secrets[envVar.Key] = envVar.Value
		} else {
			envVars[envVar.Key] = envVar.Value
		}
	}

	// Create deployment spec
	spec := &deployPkg.DeploySpec{
		AppID:       *function.AppID,
		Provider:    deployment.Provider,
		Region:      deployment.Region,
		AppName:     function.Name,
		Environment: environment,
		Version:     deployment.Version,
		Artifact:    artifact,
		Routes:      []string{"/*"},
		EnvVars:     envVars,
		Secrets:     secrets,
		ProviderConfig: map[string]interface{}{
			"backend_url":   backend.URL,
			"shared_secret": backend.SharedSecret,
		},
	}

	// Trigger deployment
	result, err := h.deploySvc.Deploy(ctx, spec)
	if err != nil {
		logrus.WithError(err).WithField("deployment_id", deployment.ID).Error("Deployment failed")
		h.updateDeploymentStatus(ctx, deployment.ID, "failed", nil, fmt.Sprintf("Deployment failed: %v", err))
		return
	}

	// Update deployment status to success
	if err := h.repo.UpdateFunctionDeploymentStatus(ctx, deployment.ID, "success", stringPtr(result.Message), nil); err != nil {
		logrus.WithError(err).WithField("deployment_id", deployment.ID).Error("Failed to update deployment status to success")
		return
	}

	logrus.WithFields(logrus.Fields{
		"function_id":   function.ID,
		"deployment_id": deployment.ID,
		"result_status": result.Status,
	}).Info("Function deployment completed successfully")
}

// updateDeploymentStatus is a helper to update deployment status with error handling
func (h *Handler) updateDeploymentStatus(ctx context.Context, deploymentID uuid.UUID, status string, deployedURL *string, errorMessage string) {
	var errMsg *string
	if errorMessage != "" {
		errMsg = &errorMessage
	}

	if err := h.repo.UpdateFunctionDeploymentStatus(ctx, deploymentID, status, deployedURL, errMsg); err != nil {
		logrus.WithError(err).WithField("deployment_id", deploymentID).Error("Failed to update deployment status")
	}
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}

// determineSideEffects determines side effects based on function capabilities
func determineSideEffects(capabilities []string) string {
	// Network-related capabilities
	networkCapabilities := map[string]bool{
		"fetch:read":   true,
		"fetch:write":  true,
		"webhook":      true,
		"email":        true,
		"external_api": true,
	}

	// External state capabilities
	stateCapabilities := map[string]bool{
		"cache:read":  true,
		"cache:write": true,
		"kv":          true,
		"storage":     true,
	}

	hasNetwork := false
	hasExternalState := false

	for _, cap := range capabilities {
		if networkCapabilities[cap] {
			hasNetwork = true
		}
		if stateCapabilities[cap] {
			hasExternalState = true
		}
	}

	if hasNetwork && hasExternalState {
		// If both, prioritize network (more restrictive)
		return "network"
	} else if hasNetwork {
		return "network"
	} else if hasExternalState {
		return "external_state"
	}

	return "none"
}

// executeTestFunction executes a function for testing purposes
func (h *Handler) executeTestFunction(ctx context.Context, req *types.TestFunctionRequest, user *storage.User) (*types.TestFunctionResponse, error) {
	startTime := time.Now()

	var functionCode string
	var functionName string

	// Get function code either from database or request
	if req.FunctionId != nil {
		// Load function from database
		function, err := h.repo.GetFunctionByID(ctx, *req.FunctionId)
		if err != nil {
			return nil, fmt.Errorf("failed to get function: %w", err)
		}

		if function.TenantID != user.TenantID {
			return nil, fmt.Errorf("function does not belong to user")
		}

		functionCode = function.Code
		functionName = function.Name
	} else {
		// Use code from request (for testing arbitrary code)
		functionCode = req.Input
		functionName = "test-function"
	}

	// Create temporary directory for compilation
	tempDir, err := os.MkdirTemp("", fmt.Sprintf("test-function-%d", time.Now().Unix()))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Compile the function
	config := &flypy.Config{
		Mode:      flypy.CompatibleMode, // Allow some non-deterministic operations for testing
		OutputDir: tempDir,
		Verbose:   false,
	}

	compiler := flypy.NewCompiler(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create compiler: %w", err)
	}

	compileCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := compiler.Compile(compileCtx, functionCode, functionName)
	if err != nil {
		return &types.TestFunctionResponse{
			Success:         false,
			Output:          nil,
			ExecutionTimeMs: int(time.Since(startTime).Milliseconds()),
			Logs:            []*storage.FunctionLog{{Message: fmt.Sprintf("Compilation failed: %v", err)}},
		}, nil
	}

	// Check for compilation warnings/errors
	if len(result.Warnings) > 0 {
		logs := make([]*storage.FunctionLog, len(result.Warnings))
		for i, warning := range result.Warnings {
			logs[i] = &storage.FunctionLog{Message: fmt.Sprintf("Warning: %s", warning)}
		}
		return &types.TestFunctionResponse{
			Success:         false,
			Output:          nil,
			ExecutionTimeMs: int(time.Since(startTime).Milliseconds()),
			Logs:            logs,
		}, nil
	}

	// Execute the compiled function
	wasmPath := filepath.Join(tempDir, "state_transition.wasm")

	// Create output buffers for capturing logs
	var stdoutBuf, stderrBuf bytes.Buffer

	// Create runtime
	runtime, err := wasm.NewPythonRuntimeWithDebug(wasmPath, &stdoutBuf, &stderrBuf, nil, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create runtime: %w", err)
	}
	defer runtime.Close()

	// Initialize runtime
	if err := runtime.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize runtime: %w", err)
	}

	// Load the compiled code
	if err := runtime.LoadCode(functionCode); err != nil {
		return nil, fmt.Errorf("failed to load code: %w", err)
	}

	// Prepare input for execution
	var inputData []byte
	if req.FunctionId != nil {
		// For database functions, use the input as JSON
		inputData = []byte(req.Input)
	} else {
		// For direct code testing, the input is already the code
		inputData = []byte("{}") // Empty JSON object as default input
	}

	// Execute the function
	output, err := runtime.Execute(inputData)
	executionTime := time.Since(startTime)

	var response *types.TestFunctionResponse
	if err != nil {
		response = &types.TestFunctionResponse{
			Success:         false,
			Output:          nil,
			ExecutionTimeMs: int(executionTime.Milliseconds()),
			Logs:            []*storage.FunctionLog{{Message: fmt.Sprintf("Execution failed: %v", err)}},
		}
	} else {
		// Parse output as JSON
		var parsedOutput interface{}
		if err := json.Unmarshal(output, &parsedOutput); err != nil {
			parsedOutput = string(output) // Fallback to string if not JSON
		}

		response = &types.TestFunctionResponse{
			Success:         true,
			Output:          parsedOutput,
			ExecutionTimeMs: int(executionTime.Milliseconds()),
			Logs:            []*storage.FunctionLog{}, // Could add execution logs here
		}
	}

	// Add any stdout/stderr output as logs
	if stdoutBuf.Len() > 0 {
		response.Logs = append(response.Logs, &storage.FunctionLog{
			Message: fmt.Sprintf("stdout: %s", stdoutBuf.String()),
		})
	}
	if stderrBuf.Len() > 0 {
		response.Logs = append(response.Logs, &storage.FunctionLog{
			Message: fmt.Sprintf("stderr: %s", stderrBuf.String()),
		})
	}

	return response, nil
}

func (h *Handler) HandleParseCode(w http.ResponseWriter, r *http.Request) {
	if h.pasteHandler != nil {
		h.pasteHandler.HandleParseCode(w, r)
		return
	}
	apierror.WriteError(w, apierror.NewServiceUnavailable("Paste handler not available"))
}

func (h *Handler) HandleCreateFromCode(w http.ResponseWriter, r *http.Request) {
	if h.pasteHandler != nil {
		h.pasteHandler.HandleCreateFromCode(w, r)
		return
	}
	apierror.WriteError(w, apierror.NewServiceUnavailable("Paste handler not available"))
}
