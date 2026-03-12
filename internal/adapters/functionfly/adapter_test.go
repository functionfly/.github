package functionfly

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/adapters/common"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
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
		{"ok custom", "eu-west-1", "https://edge.mycompany.com", false},
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

func TestDeploy_ValidAppName(t *testing.T) {
	a := NewFunctionFlyAdapter()
	ctx := context.Background()
	spec := &common.DeploymentSpec{AppName: "my-app"}
	res, err := a.Deploy(ctx, spec)
	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
	if res.Status != common.DeploymentStatusSuccess {
		t.Errorf("expected success, got %s: %s", res.Status, res.Message)
	}
	if res.DeploymentURL != "https://edge.functionfly.com/my-app" {
		t.Errorf("unexpected DeploymentURL: %s", res.DeploymentURL)
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
	// Overlong name (129 chars)
	long := strings.Repeat("a", AppNameMaxLen+1)
	res, _ := a.Deploy(ctx, &common.DeploymentSpec{AppName: long})
	if res.Status == common.DeploymentStatusSuccess {
		t.Error("expected failure for overlong app name")
	}
}

func TestDeploy_CustomBaseURL(t *testing.T) {
	a := NewFunctionFlyAdapter()
	ctx := context.Background()
	spec := &common.DeploymentSpec{
		AppName: "foo",
		ProviderConfig: map[string]interface{}{
			"url": "https://edge.custom.com",
		},
	}
	res, err := a.Deploy(ctx, spec)
	if err != nil {
		t.Fatalf("Deploy failed: %v", err)
	}
	if res.DeploymentURL != "https://edge.custom.com/foo" {
		t.Errorf("expected custom URL, got %s", res.DeploymentURL)
	}
}

func TestGetDeploymentStatus(t *testing.T) {
	a := NewFunctionFlyAdapter()
	ctx := context.Background()
	status, err := a.GetDeploymentStatus(ctx, "", nil)
	if err == nil || status != common.DeploymentStatusFailed {
		t.Errorf("expected error and failed status for empty deploymentID")
	}
	status, err = a.GetDeploymentStatus(ctx, "my-app", nil)
	if err != nil || status != common.DeploymentStatusSuccess {
		t.Errorf("expected success for non-empty deploymentID: status=%s err=%v", status, err)
	}
}
