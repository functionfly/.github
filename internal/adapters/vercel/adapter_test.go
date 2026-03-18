package vercel

import (
	"context"
	"testing"

	"github.com/functionfly/functionfly/internal/adapters/common"
)

func TestVercelAdapter_GetName(t *testing.T) {
	a := NewVercelAdapter()
	if got := a.GetName(); got != ProviderName {
		t.Errorf("GetName() = %q, want %q", got, ProviderName)
	}
}

func TestVercelAdapter_GetRegions(t *testing.T) {
	a := NewVercelAdapter()
	regions := a.GetRegions()
	if len(regions) == 0 {
		t.Fatal("GetRegions() returned empty")
	}
	seen := make(map[string]bool)
	for _, r := range regions {
		seen[r] = true
	}
	for _, want := range []string{"iad1", "sfo1", "lhr1"} {
		if !seen[want] {
			t.Errorf("GetRegions() missing region %q", want)
		}
	}
}

func TestVercelAdapter_ValidateConfig(t *testing.T) {
	a := NewVercelAdapter()
	tests := []struct {
		name    string
		region  string
		urlStr  string
		wantErr bool
	}{
		{"valid vercel.app", "iad1", "https://myapp.vercel.app", false},
		{"valid custom domain", "sfo1", "https://app.example.com", false},
		{"invalid region", "xx", "https://myapp.vercel.app", true},
		{"invalid scheme", "iad1", "http://myapp.vercel.app", true},
		{"invalid host (no dot)", "iad1", "https://invalid", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := a.ValidateConfig(tt.region, tt.urlStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVercelAdapter_Deploy_MissingToken(t *testing.T) {
	a := NewVercelAdapter()
	ctx := context.Background()
	spec := &common.DeploymentSpec{
		AppName:        "my-project",
		ProviderConfig: map[string]interface{}{"project_name": "my-project"},
		Artifact:       []byte("export default function handler() {}"),
	}
	_, err := a.Deploy(ctx, spec)
	if err == nil {
		t.Fatal("Deploy() expected error when api_token missing")
	}
	if err.Error() != "missing required Vercel config: api_token" {
		t.Errorf("Deploy() error = %v", err)
	}
}

func TestVercelAdapter_Deploy_MissingProjectName(t *testing.T) {
	a := NewVercelAdapter()
	ctx := context.Background()
	spec := &common.DeploymentSpec{
		ProviderConfig: map[string]interface{}{"api_token": "tok"},
		Artifact:       []byte("x"),
	}
	_, err := a.Deploy(ctx, spec)
	if err == nil {
		t.Fatal("Deploy() expected error when project_name missing")
	}
	if err.Error() != "missing required Vercel config: project_name (or app_name in spec)" {
		t.Errorf("Deploy() error = %v", err)
	}
}

func TestVercelAdapter_Deploy_EmptyArtifact(t *testing.T) {
	a := NewVercelAdapter()
	ctx := context.Background()
	spec := &common.DeploymentSpec{
		AppName:        "my-project",
		ProviderConfig: map[string]interface{}{"api_token": "tok", "project_name": "my-project"},
		Artifact:       nil,
	}
	result, err := a.Deploy(ctx, spec)
	if err != nil {
		t.Fatalf("Deploy() unexpected error: %v", err)
	}
	if result.Status != common.DeploymentStatusFailed {
		t.Errorf("Deploy() status = %v, want failed", result.Status)
	}
}
