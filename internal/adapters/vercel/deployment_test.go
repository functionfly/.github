package vercel

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVercelDeploymentClient_Deploy(t *testing.T) {
	var reqPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath = r.URL.Path
		if r.Method != "POST" || r.URL.Path != "/v13/deployments" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Consume body so client can finish (multipart); don't rely on parsing
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"uid":       "dpl_abc123",
			"url":       "https://myapp-abc123.vercel.app",
			"state":     "READY",
			"createdAt": 1234567890,
		})
	}))
	defer server.Close()

	client := NewVercelDeploymentClientWithBase("token", "", server.URL)
	ctx := context.Background()
	result, err := client.Deploy(ctx, []byte("export default () => {}"), "myapp", nil)
	if err != nil {
		t.Fatalf("Deploy = %v", err)
	}
	if result.DeploymentID != "dpl_abc123" || result.Status != "success" {
		t.Errorf("Deploy result = %+v", result)
	}
	if result.DeploymentURL != "https://myapp-abc123.vercel.app" {
		t.Errorf("DeploymentURL = %q", result.DeploymentURL)
	}
	if reqPath != "/v13/deployments" {
		t.Errorf("request path = %q", reqPath)
	}
}

func TestVercelDeploymentClient_GetDeploymentStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v13/deployments/dpl_xyz" || r.Method != "GET" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"state": "READY"})
	}))
	defer server.Close()

	client := NewVercelDeploymentClientWithBase("token", "", server.URL)
	ctx := context.Background()
	status, err := client.GetDeploymentStatus(ctx, "dpl_xyz")
	if err != nil {
		t.Fatalf("GetDeploymentStatus = %v", err)
	}
	if status != "success" {
		t.Errorf("status = %q", status)
	}
}

func TestVercelDeploymentClient_GetProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v6/projects/myproject" || r.Method != "GET" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        "prj_abc",
			"name":      "myproject",
			"framework": "nextjs",
			"createdAt": 1234567890,
			"updatedAt": 1234567890,
		})
	}))
	defer server.Close()

	client := NewVercelDeploymentClientWithBase("token", "", server.URL)
	ctx := context.Background()
	project, err := client.GetProject(ctx, "myproject")
	if err != nil {
		t.Fatalf("GetProject = %v", err)
	}
	if project == nil || project.ID != "prj_abc" || project.Name != "myproject" {
		t.Errorf("GetProject = %+v", project)
	}
}

func TestVercelDeploymentClient_GetProject_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewVercelDeploymentClientWithBase("token", "", server.URL)
	ctx := context.Background()
	project, err := client.GetProject(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetProject = %v", err)
	}
	if project != nil {
		t.Errorf("GetProject expected nil for 404, got %+v", project)
	}
}

func TestGetVercelAPIBase(t *testing.T) {
	base := getVercelAPIBase()
	if base == "" {
		t.Error("getVercelAPIBase() returned empty")
	}
	if len(base) > 0 && base[len(base)-1] == '/' {
		t.Errorf("getVercelAPIBase() should not end with slash: %q", base)
	}
}
