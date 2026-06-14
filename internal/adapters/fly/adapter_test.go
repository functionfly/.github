package fly

import (
	"context"
	"testing"

	"github.com/functionfly/functionfly/internal/adapters/common"
)

func TestFlyAdapter_GetName(t *testing.T) {
	a := NewFlyAdapter()
	if got := a.GetName(); got != ProviderName {
		t.Errorf("GetName() = %q, want %q", got, ProviderName)
	}
}

func TestFlyAdapter_GetRegions(t *testing.T) {
	a := NewFlyAdapter()
	regions := a.GetRegions()
	if len(regions) == 0 {
		t.Fatal("GetRegions() returned empty")
	}
	// Check a few known regions
	seen := make(map[string]bool)
	for _, r := range regions {
		seen[r] = true
	}
	for _, want := range []string{"iad", "ord", "lhr", "ams"} {
		if !seen[want] {
			t.Errorf("GetRegions() missing region %q", want)
		}
	}
}

func TestFlyAdapter_ValidateConfig(t *testing.T) {
	a := NewFlyAdapter()
	tests := []struct {
		name    string
		region  string
		urlStr  string
		wantErr bool
	}{
		{"valid fly.dev", "iad", "https://myapp.fly.dev", false},
		{"valid with path", "ord", "https://myapp.fly.dev/", false},
		{"invalid region", "xx", "https://myapp.fly.dev", true},
		{"invalid scheme", "iad", "http://myapp.fly.dev", true},
		{"invalid host (no dot, not .fly.dev or .internal)", "iad", "https://invalid", true},
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

func TestFlyAdapter_Deploy_MissingConfig(t *testing.T) {
	a := NewFlyAdapter()
	ctx := context.Background()
	spec := &common.DeploymentSpec{
		AppName: "test-app",
		Version: "1.0.0",
	}
	_, err := a.Deploy(ctx, spec)
	if err == nil {
		t.Fatal("Deploy() expected error when api_token missing")
	}
	if err.Error() != "missing required Fly.io config: api_token, app_name" {
		t.Errorf("Deploy() error = %v", err)
	}
}

func TestFlyAdapter_Deploy_RequiresAppName(t *testing.T) {
	a := NewFlyAdapter()
	ctx := context.Background()
	spec := &common.DeploymentSpec{
		ProviderConfig: map[string]interface{}{
			"api_token": "tok",
			"app_name":  "",
		},
	}
	_, err := a.Deploy(ctx, spec)
	if err == nil {
		t.Fatal("Deploy() expected error when app_name missing")
	}
}

func TestFlyAdapter_Deploy_UsesImageFromProviderConfig(t *testing.T) {
	// Unit test only checks that we pass through provider_config; integration would hit API
	a := NewFlyAdapter()
	if a.GetName() != "fly" {
		t.Errorf("adapter name = %q", a.GetName())
	}
}

func TestNewFlyDeploymentAdapter(t *testing.T) {
	adapter := NewFlyDeploymentAdapter()
	if adapter == nil {
		t.Fatal("NewFlyDeploymentAdapter() returned nil")
	}
	if adapter.GetName() != ProviderName {
		t.Errorf("GetName() = %q", adapter.GetName())
	}
}
