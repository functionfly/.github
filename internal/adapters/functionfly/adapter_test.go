package functionfly

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/common"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	mock "github.com/stretchr/testify/mock"
)

func TestValidateConfig(t *testing.T) {
	a := NewFunctionFlyAdapter()
	tests := []struct {
		name    string
		region  string
		url     string
		wantErr bool
	}{
		{"empty url", "us-east-1", "", true},
		{"invalid url", "us-east-1", "://bad", true},
		{"http rejected", "us-east-1", "http://edge.example.com", true},
		{"invalid region", "mars", "https://edge.functionfly.com", true},
		{"ok default", "us-east-1", "https://edge.functionfly.com", false},
		{"ok custom", "eu-central-1", "https://edge.mycompany.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := a.ValidateConfig(tt.region, tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHealthCheck_NilBackend(t *testing.T) {
	a := NewFunctionFlyAdapter()
	ctx := context.Background()
	result, err := a.HealthCheck(ctx, nil)
	if err != nil {
		t.Fatalf("HealthCheck with nil backend should not return error: %v", err)
	}
	if result.OK {
		t.Error("expected OK false for nil backend")
	}
	if result.ErrorMessage == "" {
		t.Error("expected non-empty ErrorMessage for nil backend")
	}
}

func TestHealthCheck_EmptyURL(t *testing.T) {
	a := NewFunctionFlyAdapter()
	ctx := context.Background()
	backend := &storage.Backend{URL: "", Region: "us-east-1"}
	result, err := a.HealthCheck(ctx, backend)
	if err != nil {
		t.Fatalf("HealthCheck should not return error: %v", err)
	}
	if result.OK {
		t.Error("expected OK false for empty URL")
	}
}

func TestDeploy_ValidAppName_WithMock(t *testing.T) {
	mockClient := NewMockEdgeDeploymentClientInterface(t)

	adapter := NewFunctionFlyAdapter()
	adapter = adapter.WithDeploymentClient(mockClient)

	mockClient.On("RegisterFunction", context.Background(), &common.DeploymentSpec{
		AppName: "test-app",
		Runtime: common.RuntimeJavaScript,
	}).Return(&DeployResponse{
		Success:       true,
		DeploymentID:  "test-deployment-id",
		Status:        common.DeploymentStatusSuccess,
		Message:       "Function registered successfully",
		DeploymentURL: "https://edge.functionfly.com/test-app",
	}, nil)

	ctx := context.Background()
	spec := &common.DeploymentSpec{
		AppName: "test-app",
		Runtime: common.RuntimeJavaScript,
	}
	res, err := adapter.Deploy(ctx, spec)
	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
	if res.Status != common.DeploymentStatusSuccess {
		t.Errorf("expected success, got %s: %s", res.Status, res.Message)
	}
	if res.DeploymentID != "test-deployment-id" {
		t.Errorf("unexpected DeploymentID: %s", res.DeploymentID)
	}
}

func TestDeploy_ApiUnavailable_GracefulFallback(t *testing.T) {
	mockClient := NewMockEdgeDeploymentClientInterface(t)

	adapter := NewFunctionFlyAdapter()
	adapter = adapter.WithDeploymentClient(mockClient)

	mockClient.On("RegisterFunction", context.Background(), &common.DeploymentSpec{
		AppName: "test-app",
		Runtime: common.RuntimeJavaScript,
	}).Return(nil, context.DeadlineExceeded)

	ctx := context.Background()
	spec := &common.DeploymentSpec{
		AppName: "test-app",
		Runtime: common.RuntimeJavaScript,
	}
	res, err := adapter.Deploy(ctx, spec)
	if err != nil {
		t.Fatalf("Deploy should not return error even if API unavailable: %v", err)
	}
	if res.Status != common.DeploymentStatusSuccess {
		t.Errorf("expected success (graceful fallback), got %s: %s", res.Status, res.Message)
	}
	if res.DeploymentURL != "https://edge.functionfly.com/test-app" {
		t.Errorf("unexpected DeploymentURL: %s", res.DeploymentURL)
	}
	registered, ok := res.Metadata["registered"].(bool)
	if !ok || registered {
		t.Error("expected registered=false in metadata when API unavailable")
	}
}

func TestDeploy_InvalidAppNames(t *testing.T) {
	a := NewFunctionFlyAdapter()
	ctx := context.Background()
	bad := []string{"", "../etc", "a b", "a/b", "a.b", " leading", "trailing "}
	for _, appName := range bad {
		res, err := a.Deploy(ctx, &common.DeploymentSpec{AppName: appName})
		if err != nil {
			continue
		}
		if res.Status == common.DeploymentStatusSuccess {
			t.Errorf("expected failure for app name %q", appName)
		}
	}
	long := "a"
	for i := 0; i < AppNameMaxLen+1; i++ {
		long += "a"
	}
	res, _ := a.Deploy(ctx, &common.DeploymentSpec{AppName: long})
	if res.Status == common.DeploymentStatusSuccess {
		t.Error("expected failure for overlong app name")
	}
}

func TestDeploy_CustomBaseURL(t *testing.T) {
	mockClient := NewMockEdgeDeploymentClientInterface(t)

	adapter := NewFunctionFlyAdapter()
	adapter = adapter.WithDeploymentClient(mockClient)

	mockClient.On("RegisterFunction", context.Background(), mock.Anything).Return(&DeployResponse{
		Success:       true,
		DeploymentID:  "test-deployment-id",
		Status:        common.DeploymentStatusSuccess,
		DeploymentURL: "https://edge.custom.com/foo",
	}, nil)

	ctx := context.Background()
	spec := &common.DeploymentSpec{
		AppName: "foo",
		ProviderConfig: map[string]interface{}{
			"url": "https://edge.custom.com",
		},
	}
	res, err := adapter.Deploy(ctx, spec)
	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
	if res.DeploymentURL != "https://edge.custom.com/foo" {
		t.Errorf("expected custom URL, got %s", res.DeploymentURL)
	}
}

func TestDeploy_WASM(t *testing.T) {
	mockClient := NewMockEdgeDeploymentClientInterface(t)

	adapter := NewFunctionFlyAdapter()
	adapter = adapter.WithDeploymentClient(mockClient)

	mockClient.On("RegisterFunction", context.Background(), &common.DeploymentSpec{
		AppName:  "test-wasm",
		Runtime:  common.RuntimeWASM,
		Artifact: []byte("fake wasm binary"),
	}).Return(&DeployResponse{
		Success:       true,
		DeploymentID:  "test-wasm",
		Status:        common.DeploymentStatusSuccess,
		DeploymentURL: "http://localhost:8080/test-wasm",
		Message:       "WASM function deployed",
	}, nil)

	ctx := context.Background()
	spec := &common.DeploymentSpec{
		AppName:  "test-wasm",
		Runtime:  common.RuntimeWASM,
		Artifact: []byte("fake wasm binary"),
	}
	res, err := adapter.Deploy(ctx, spec)
	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
	if res.Status != common.DeploymentStatusSuccess {
		t.Errorf("expected success, got %s: %s", res.Status, res.Message)
	}
	if res.Metadata["runtime"] != string(common.RuntimeWASM) {
		t.Errorf("unexpected runtime in metadata: %v", res.Metadata["runtime"])
	}
}

func TestGetDeploymentStatus_ApiAvailable(t *testing.T) {
	mockClient := NewMockEdgeDeploymentClientInterface(t)

	adapter := NewFunctionFlyAdapter()
	adapter = adapter.WithDeploymentClient(mockClient)

	mockClient.On("GetFunctionStatus", context.Background(), "my-app").Return(&StatusResponse{
		Exists:   true,
		Status:   common.DeploymentStatusSuccess,
		Deployed: true,
	}, nil)

	ctx := context.Background()
	status, err := adapter.GetDeploymentStatus(ctx, "my-app", nil)
	if err != nil {
		t.Fatalf("GetDeploymentStatus failed: %v", err)
	}
	if status != common.DeploymentStatusSuccess {
		t.Errorf("expected success, got %s", status)
	}
}

func TestGetDeploymentStatus_NotFound(t *testing.T) {
	mockClient := NewMockEdgeDeploymentClientInterface(t)

	adapter := NewFunctionFlyAdapter()
	adapter = adapter.WithDeploymentClient(mockClient)

	mockClient.On("GetFunctionStatus", context.Background(), "nonexistent").Return(&StatusResponse{
		Exists:   false,
		Status:   common.DeploymentStatusFailed,
		Message:  "function not found",
		Deployed: false,
	}, nil)

	ctx := context.Background()
	status, err := adapter.GetDeploymentStatus(ctx, "nonexistent", nil)
	if err == nil {
		t.Fatalf("GetDeploymentStatus should return error for not found")
	}
	if status != common.DeploymentStatusFailed {
		t.Errorf("expected failed status, got %s", status)
	}
}

func TestGetDeploymentStatus_InvalidAppName(t *testing.T) {
	a := NewFunctionFlyAdapter()
	ctx := context.Background()
	status, err := a.GetDeploymentStatus(ctx, "", nil)
	if err == nil || status != common.DeploymentStatusFailed {
		t.Errorf("expected error and failed status for empty deploymentID")
	}

	status, err = a.GetDeploymentStatus(ctx, "../etc", nil)
	if err == nil || status != common.DeploymentStatusFailed {
		t.Errorf("expected error for invalid app name")
	}
}

func TestSetEnv_Mock(t *testing.T) {
	mockClient := NewMockEdgeDeploymentClientInterface(t)

	adapter := NewFunctionFlyAdapter()
	adapter = adapter.WithDeploymentClient(mockClient)

	envVars := map[string]string{"NODE_ENV": "production"}
	secrets := map[string]string{"API_KEY": "secret123"}

	mockClient.On("SetEnvironment", context.Background(), "test-app", envVars, secrets).Return(nil)

	ctx := context.Background()
	err := adapter.SetEnv(ctx, "test-app", nil, envVars, secrets)
	if err != nil {
		t.Fatalf("SetEnv failed: %v", err)
	}
}

func TestBindRoutes_Mock(t *testing.T) {
	mockClient := NewMockEdgeDeploymentClientInterface(t)

	adapter := NewFunctionFlyAdapter()
	adapter = adapter.WithDeploymentClient(mockClient)

	routes := []common.RouteBinding{
		{Pattern: "/api/*"},
		{Pattern: "/webhook"},
	}

	mockClient.On("BindRoutes", context.Background(), "test-app", routes).Return(nil)

	ctx := context.Background()
	err := adapter.BindRoutes(ctx, "test-app", nil, routes)
	if err != nil {
		t.Fatalf("BindRoutes failed: %v", err)
	}
}

func TestRollback_NotFound(t *testing.T) {
	mockClient := NewMockEdgeDeploymentClientInterface(t)

	adapter := NewFunctionFlyAdapter()
	adapter = adapter.WithDeploymentClient(mockClient)

	mockClient.On("GetFunctionStatus", context.Background(), "test-app").Return(&StatusResponse{
		Exists:   false,
		Status:   common.DeploymentStatusFailed,
		Message:  "function not found",
	}, nil)

	ctx := context.Background()
	spec := &common.DeploymentSpec{
		AppName: "test-app",
		Version: "v1.0.0",
		Runtime: common.RuntimeJavaScript,
	}
	res, err := adapter.Rollback(ctx, spec)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
	if res.Status != common.DeploymentStatusFailed {
		t.Errorf("expected failure for non-existent function, got %s", res.Status)
	}
}

func TestRollback_SameVersion(t *testing.T) {
	mockClient := NewMockEdgeDeploymentClientInterface(t)

	adapter := NewFunctionFlyAdapter()
	adapter = adapter.WithDeploymentClient(mockClient)

	mockClient.On("GetFunctionStatus", context.Background(), "test-app").Return(&StatusResponse{
		Exists:   true,
		Status:   common.DeploymentStatusSuccess,
		Version:  "v1.0.0",
		Deployed: true,
	}, nil)

	ctx := context.Background()
	spec := &common.DeploymentSpec{
		AppName: "test-app",
		Version: "v1.0.0",
		Runtime: common.RuntimeJavaScript,
	}
	res, err := adapter.Rollback(ctx, spec)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
	if res.Status != common.DeploymentStatusSuccess {
		t.Errorf("expected success for same version, got %s", res.Status)
	}
	noChange, ok := res.Metadata["no_change"].(bool)
	if !ok || !noChange {
		t.Error("expected no_change=true in metadata")
	}
}

func TestRollback_ReDeploy(t *testing.T) {
	mockClient := NewMockEdgeDeploymentClientInterface(t)

	adapter := NewFunctionFlyAdapter()
	adapter = adapter.WithDeploymentClient(mockClient)

	mockClient.On("GetFunctionStatus", context.Background(), "test-app").Return(&StatusResponse{
		Exists:   true,
		Status:   common.DeploymentStatusSuccess,
		Version:  "v2.0.0",
		Deployed: true,
	}, nil)

	mockClient.On("RegisterFunction", context.Background(), &common.DeploymentSpec{
		AppName: "test-app",
		Version: "v1.0.0",
		Runtime: common.RuntimeJavaScript,
	}).Return(&DeployResponse{
		Success:       true,
		DeploymentID:  "test-deployment-id",
		Status:        common.DeploymentStatusSuccess,
		Message:       "Rolled back to v1.0.0",
		DeploymentURL: "https://edge.functionfly.com/test-app",
	}, nil)

	ctx := context.Background()
	spec := &common.DeploymentSpec{
		AppName: "test-app",
		Version: "v1.0.0",
		Runtime: common.RuntimeJavaScript,
	}
	res, err := adapter.Rollback(ctx, spec)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
	if res.Status != common.DeploymentStatusSuccess {
		t.Errorf("expected success, got %s", res.Status)
	}
}

func TestHealthCheck_RealHTTP(t *testing.T) {
	ok := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = r.URL.Path == HealthPath
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	backend := &storage.Backend{
		ID:           uuid.New(),
		URL:          server.URL,
		Region:       "us-east-1",
		SharedSecret: "test-secret",
	}
	client := &http.Client{Timeout: 5 * time.Second}
	a := NewFunctionFlyAdapterWithClient(client)
	ctx := context.Background()
	result, err := a.HealthCheck(ctx, backend)
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}
	if !result.OK {
		t.Errorf("expected OK true, got false (status=%d, msg=%s)", result.StatusCode, result.ErrorMessage)
	}
	if !ok {
		t.Error("handler was not called with /healthz path")
	}
}
