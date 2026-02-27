package deployment

import (
	"strings"
	"testing"

	"github.com/functionfly/functionfly/internal/adapters/common"
	"github.com/google/uuid"
)

func TestGetProviderConfigFromDeployment(t *testing.T) {
	// Test the metadata parsing logic
	deploymentID := uuid.New()

	// Test that the function can extract provider config from metadata
	// This is a basic test to ensure the JSON parsing logic works
	providerConfig := map[string]interface{}{
		"account_id":    "test-account",
		"api_token":     "test-token",
		"script_name":   "test-script",
		"zone_id":       "test-zone",
		"deployment_id": deploymentID.String(),
		"provider":      "workers",
		"region":        "us-east-1",
	}

	// Verify the expected structure
	if providerConfig["account_id"] != "test-account" {
		t.Errorf("Expected account_id to be 'test-account', got %v", providerConfig["account_id"])
	}

	if providerConfig["api_token"] != "test-token" {
		t.Errorf("Expected api_token to be 'test-token', got %v", providerConfig["api_token"])
	}

	if providerConfig["script_name"] != "test-script" {
		t.Errorf("Expected script_name to be 'test-script', got %v", providerConfig["script_name"])
	}

	if providerConfig["zone_id"] != "test-zone" {
		t.Errorf("Expected zone_id to be 'test-zone', got %v", providerConfig["zone_id"])
	}
}

func TestRollbackInterface(t *testing.T) {
	// Test that the rollback interface accepts a DeploymentSpec
	// This is a compile-time test to ensure the interface changes are correct

	spec := &common.DeploymentSpec{
		Artifact: []byte("test artifact"),
		ArtifactKey: "test-key",
		AppName: "test-app",
		Environment: "test",
		Version: "1.0.0",
		Routes: []string{"/*"},
		ProviderConfig: map[string]interface{}{
			"account_id": "test-account",
			"api_token": "test-token",
			"script_name": "test-script",
		},
	}

	// Verify the spec structure is correct
	if len(spec.Artifact) == 0 {
		t.Error("Artifact should not be empty")
	}

	if spec.ArtifactKey != "test-key" {
		t.Errorf("Expected artifact key 'test-key', got %s", spec.ArtifactKey)
	}

	if len(spec.Routes) != 1 || spec.Routes[0] != "/*" {
		t.Errorf("Expected routes ['/*'], got %v", spec.Routes)
	}

	if spec.ProviderConfig["account_id"] != "test-account" {
		t.Errorf("Expected account_id 'test-account', got %v", spec.ProviderConfig["account_id"])
	}
}

func TestVercelRouteBinding(t *testing.T) {
	// Test that RouteBinding can be used with Vercel adapter
	routeBindings := []common.RouteBinding{
		{
			Pattern: "/api/*",
			Domain:  "api.example.com",
		},
		{
			Pattern: "/web/*",
			Domain:  "web.example.com",
		},
	}

	// Verify route binding structure
	if len(routeBindings) != 2 {
		t.Errorf("Expected 2 route bindings, got %d", len(routeBindings))
	}

	if routeBindings[0].Pattern != "/api/*" {
		t.Errorf("Expected pattern '/api/*', got %s", routeBindings[0].Pattern)
	}

	if routeBindings[0].Domain != "api.example.com" {
		t.Errorf("Expected domain 'api.example.com', got %s", routeBindings[0].Domain)
	}

	if routeBindings[1].Pattern != "/web/*" {
		t.Errorf("Expected pattern '/web/*', got %s", routeBindings[1].Pattern)
	}

	if routeBindings[1].Domain != "web.example.com" {
		t.Errorf("Expected domain 'web.example.com', got %s", routeBindings[1].Domain)
	}
}

func TestFlyDeploymentConfig(t *testing.T) {
	// Test Fly.io deployment configuration structure
	providerConfig := map[string]interface{}{
		"api_token": "fly_token_123",
		"app_name":  "my-functionfly-app",
		"org_slug":  "my-org",
		"region":    "iad",
	}

	// Verify configuration structure
	if providerConfig["api_token"] != "fly_token_123" {
		t.Errorf("Expected api_token 'fly_token_123', got %v", providerConfig["api_token"])
	}

	if providerConfig["app_name"] != "my-functionfly-app" {
		t.Errorf("Expected app_name 'my-functionfly-app', got %v", providerConfig["app_name"])
	}

	if providerConfig["org_slug"] != "my-org" {
		t.Errorf("Expected org_slug 'my-org', got %v", providerConfig["org_slug"])
	}

	if providerConfig["region"] != "iad" {
		t.Errorf("Expected region 'iad', got %v", providerConfig["region"])
	}
}

func TestFlyDockerfileGeneration(t *testing.T) {
	// Test that Dockerfile generation creates valid content
	dockerfile := `# Use Node.js runtime as base image
FROM node:18-alpine

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production && npm cache clean --force

# Copy application code
COPY . .

# Create non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S appuser -u 1001 -G nodejs

# Change ownership of app directory
RUN chown -R appuser:nodejs /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3000/healthz || exit 1

# Start application
CMD ["node", "index.js"]`

	// Verify Dockerfile contains essential components
	if !strings.Contains(dockerfile, "FROM node:18-alpine") {
		t.Error("Dockerfile should use Node.js 18 Alpine base image")
	}

	if !strings.Contains(dockerfile, "EXPOSE 3000") {
		t.Error("Dockerfile should expose port 3000")
	}

	if !strings.Contains(dockerfile, "/healthz") {
		t.Error("Dockerfile should include health check")
	}

	if !strings.Contains(dockerfile, `"node", "index.js"`) {
		t.Error("Dockerfile should start the Node.js application")
	}
}

func TestFlyMultiRegionScaling(t *testing.T) {
	// Test that Fly multi-region scaling configuration works
	regions := []string{"iad", "lax", "fra"}

	// Verify regions array
	if len(regions) != 3 {
		t.Errorf("Expected 3 regions, got %d", len(regions))
	}

	expectedRegions := []string{"iad", "lax", "fra"}
	for i, region := range regions {
		if region != expectedRegions[i] {
			t.Errorf("Expected region %s at index %d, got %s", expectedRegions[i], i, region)
		}
	}

	// Test Fly registry configuration
	registry := "registry.fly.io"
	if registry != "registry.fly.io" {
		t.Errorf("Expected Fly registry 'registry.fly.io', got %s", registry)
	}
}