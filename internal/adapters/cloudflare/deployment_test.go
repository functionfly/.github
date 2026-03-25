package cloudflare

import (
	"context"
	"testing"

	"github.com/functionfly/functionfly/internal/adapters/common"
)

func TestDeploy_EmptyContent(t *testing.T) {
	client := NewCloudflareDeploymentClient("token", "account-id")
	ctx := context.Background()
	_, err := client.Deploy(ctx, []byte{}, "myscript", common.RuntimeJavaScript)
	if err == nil {
		t.Fatal("expected error for empty script content")
	}
	if err.Error() != "script content cannot be empty" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeploy_EmptyScriptName(t *testing.T) {
	client := NewCloudflareDeploymentClient("token", "account-id")
	ctx := context.Background()
	_, err := client.Deploy(ctx, []byte("x"), "", common.RuntimeJavaScript)
	if err == nil {
		t.Fatal("expected error for empty script name")
	}
	if err.Error() != "script name cannot be empty" {
		t.Errorf("unexpected error: %v", err)
	}
}
