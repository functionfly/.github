package fly

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFlyDeploymentClient_EnsureApp_Exists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/apps/myapp" || r.Method != "GET" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewFlyDeploymentClientWithBase("token", server.URL)
	ctx := context.Background()
	err := client.EnsureApp(ctx, "myapp", "personal")
	if err != nil {
		t.Fatalf("EnsureApp (existing) = %v", err)
	}
}

func TestFlyDeploymentClient_EnsureApp_CreateRequiresOrgSlug(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/apps/newapp" && r.Method == "GET" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewFlyDeploymentClientWithBase("token", server.URL)
	ctx := context.Background()
	err := client.EnsureApp(ctx, "newapp", "")
	if err == nil {
		t.Fatal("EnsureApp (create without org_slug) expected error")
	}
	if err.Error() != "org_slug is required when creating a new Fly app; set provider_config.org_slug (e.g. \"personal\")" {
		t.Errorf("EnsureApp error = %v", err)
	}
}

func TestFlyDeploymentClient_EnsureApp_Create(t *testing.T) {
	var postCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/apps/newapp" && r.Method == "GET" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.URL.Path == "/v1/apps" && r.Method == "POST" {
			postCalled = true
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["org_slug"] != "personal" || body["app_name"] != "newapp" {
				t.Errorf("create body = %v", body)
			}
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewFlyDeploymentClientWithBase("token", server.URL)
	ctx := context.Background()
	err := client.EnsureApp(ctx, "newapp", "personal")
	if err != nil {
		t.Fatalf("EnsureApp (create) = %v", err)
	}
	if !postCalled {
		t.Fatal("POST /v1/apps was not called")
	}
}

func TestFlyDeploymentClient_Deploy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/apps/myapp/machines" || r.Method != "POST" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		config, _ := body["config"].(map[string]interface{})
		if config["image"] != "registry.fly.io/myapp:v1" {
			t.Errorf("config.image = %v", config["image"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "machine-123",
			"state":  "started",
			"region": "iad",
		})
	}))
	defer server.Close()

	client := NewFlyDeploymentClientWithBase("token", server.URL)
	ctx := context.Background()
	result, err := client.Deploy(ctx, "myapp", "registry.fly.io/myapp:v1")
	if err != nil {
		t.Fatalf("Deploy = %v", err)
	}
	if result.DeploymentID != "machine-123" || result.Status != "success" {
		t.Errorf("Deploy result = %+v", result)
	}
}

func TestFlyDeploymentClient_Deploy_RequiresImageRef(t *testing.T) {
	client := NewFlyDeploymentClientWithBase("token", "http://localhost")
	ctx := context.Background()
	_, err := client.Deploy(ctx, "myapp", "")
	if err == nil {
		t.Fatal("Deploy with empty image ref expected error")
	}
}

func TestFlyDeploymentClient_ListMachines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/apps/myapp/machines" || r.Method != "GET" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"id":     "m1",
				"state":  "started",
				"region": "iad",
				"config": map[string]interface{}{"image": "registry.fly.io/myapp:old"},
			},
		})
	}))
	defer server.Close()

	client := NewFlyDeploymentClientWithBase("token", server.URL)
	ctx := context.Background()
	machines, err := client.ListMachines(ctx, "myapp")
	if err != nil {
		t.Fatalf("ListMachines = %v", err)
	}
	if len(machines) != 1 || machines[0].ID != "m1" {
		t.Errorf("ListMachines = %+v", machines)
	}
}

func TestFlyDeploymentClient_Rollback(t *testing.T) {
	var updateBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/apps/myapp/machines" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]map[string]interface{}{
				{
					"id":     "m1",
					"state":  "started",
					"region": "iad",
					"config": map[string]interface{}{"image": "registry.fly.io/myapp:current"},
				},
			})
			return
		}
		if r.URL.Path == "/v1/apps/myapp/machines/m1" && r.Method == "POST" {
			_ = json.NewDecoder(r.Body).Decode(&updateBody)
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewFlyDeploymentClientWithBase("token", server.URL)
	ctx := context.Background()
	result, err := client.Rollback(ctx, "myapp", "previous")
	if err != nil {
		t.Fatalf("Rollback = %v", err)
	}
	if result.Status != "success" || result.DeploymentID != "myapp" {
		t.Errorf("Rollback result = %+v", result)
	}
	config, _ := updateBody["config"].(map[string]interface{})
	if config["image"] != "registry.fly.io/myapp:previous" {
		t.Errorf("rollback update config.image = %v", config["image"])
	}
}

func TestGetFlyAPIBase(t *testing.T) {
	base := getFlyAPIBase()
	if base == "" {
		t.Error("getFlyAPIBase() returned empty")
	}
	// Should not contain trailing slash (TrimSuffix in getFlyAPIBase)
	if len(base) > 0 && base[len(base)-1] == '/' {
		t.Errorf("getFlyAPIBase() should not end with slash: %q", base)
	}
}

func TestFlyDeploymentClient_WaitForDeployment(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/apps/myapp/machines/m1" && r.Method == "GET" {
			callCount++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			state := "starting"
			if callCount >= 3 {
				state = "started"
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"state": state,
			})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewFlyDeploymentClientWithBase("token", server.URL)
	ctx := context.Background()
	status, err := client.WaitForDeployment(ctx, "myapp", "m1", 30*time.Second)
	if err != nil {
		t.Fatalf("WaitForDeployment = %v", err)
	}
	if status != "success" {
		t.Errorf("WaitForDeployment status = %v, want success", status)
	}
	if callCount < 3 {
		t.Errorf("WaitForDeployment callCount = %d, want >= 3", callCount)
	}
}

func TestFlyDeploymentClient_GetAppInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/apps/myapp" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "myapp",
				"status": "running",
			})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewFlyDeploymentClientWithBase("token", server.URL)
	ctx := context.Background()
	appInfo, err := client.GetAppInfo(ctx, "myapp")
	if err != nil {
		t.Fatalf("GetAppInfo = %v", err)
	}
	if appInfo["name"] != "myapp" {
		t.Errorf("GetAppInfo name = %v, want myapp", appInfo["name"])
	}
}

func TestFlyDeploymentClient_ListAppRegions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/apps/myapp/machines" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]map[string]interface{}{
				{
					"id":     "m1",
					"state":  "started",
					"region": "iad",
					"config": map[string]interface{}{"image": "registry.fly.io/myapp:v1"},
				},
				{
					"id":     "m2",
					"state":  "started",
					"region": "lhr",
					"config": map[string]interface{}{"image": "registry.fly.io/myapp:v1"},
				},
			})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewFlyDeploymentClientWithBase("token", server.URL)
	ctx := context.Background()
	regions, err := client.ListAppRegions(ctx, "myapp")
	if err != nil {
		t.Fatalf("ListAppRegions = %v", err)
	}
	if len(regions) != 2 {
		t.Errorf("ListAppRegions count = %d, want 2", len(regions))
	}
	regionSet := make(map[string]bool)
	for _, r := range regions {
		regionSet[r] = true
	}
	if !regionSet["iad"] || !regionSet["lhr"] {
		t.Errorf("ListAppRegions = %v, want [iad, lhr]", regions)
	}
}

func TestFlyDeploymentClient_ScaleApp(t *testing.T) {
	var createCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/apps/myapp/machines" && r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode([]map[string]interface{}{
				{
					"id":     "m1",
					"state":  "started",
					"region": "iad",
					"config": map[string]interface{}{"image": "registry.fly.io/myapp:v1"},
				},
			})
			return
		}
		if r.URL.Path == "/v1/apps/myapp/machines" && r.Method == "POST" {
			createCalled = true
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["region"] != "lhr" {
				t.Errorf("ScaleApp region = %v, want lhr", body["region"])
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":     "m2",
				"state":  "started",
				"region": "lhr",
			})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewFlyDeploymentClientWithBase("token", server.URL)
	ctx := context.Background()
	err := client.ScaleApp(ctx, "myapp", "lhr", 1)
	if err != nil {
		t.Fatalf("ScaleApp = %v", err)
	}
	if !createCalled {
		t.Fatal("ScaleApp did not create new machine")
	}
}

func TestFlyDeploymentClient_DeploymentArtifactCRUD(t *testing.T) {
	client := NewFlyDeploymentClientWithBase("token", "http://localhost")
	ctx := context.Background()
	now := time.Now().UTC()

	if err := client.StoreDeploymentArtifact(ctx, &DeploymentArtifact{
		AppName:    "myapp",
		ImageRef:   "registry.fly.io/myapp:v1",
		Version:    "v1",
		DeployedAt: now.Add(-time.Minute),
		Status:     "success",
	}); err != nil {
		t.Fatalf("StoreDeploymentArtifact(v1) = %v", err)
	}
	if err := client.StoreDeploymentArtifact(ctx, &DeploymentArtifact{
		AppName:    "myapp",
		ImageRef:   "registry.fly.io/myapp:v2",
		Version:    "v2",
		DeployedAt: now,
		Status:     "success",
	}); err != nil {
		t.Fatalf("StoreDeploymentArtifact(v2) = %v", err)
	}

	got, err := client.GetDeploymentArtifact(ctx, "myapp", "v2")
	if err != nil {
		t.Fatalf("GetDeploymentArtifact = %v", err)
	}
	if got.ImageRef != "registry.fly.io/myapp:v2" {
		t.Errorf("GetDeploymentArtifact image = %q", got.ImageRef)
	}

	artifacts, err := client.ListDeploymentArtifacts(ctx, "myapp", 1)
	if err != nil {
		t.Fatalf("ListDeploymentArtifacts = %v", err)
	}
	if len(artifacts) != 1 || artifacts[0].Version != "v2" {
		t.Fatalf("ListDeploymentArtifacts(limit=1) = %+v", artifacts)
	}

	if err := client.DeleteDeploymentArtifact(ctx, "myapp", "v2"); err != nil {
		t.Fatalf("DeleteDeploymentArtifact = %v", err)
	}
	if _, err := client.GetDeploymentArtifact(ctx, "myapp", "v2"); err == nil {
		t.Fatal("GetDeploymentArtifact after delete expected error")
	}
}

func TestFlyDeploymentClient_StoreDeploymentArtifact_Validation(t *testing.T) {
	client := NewFlyDeploymentClientWithBase("token", "http://localhost")
	ctx := context.Background()

	if err := client.StoreDeploymentArtifact(ctx, nil); err == nil {
		t.Fatal("StoreDeploymentArtifact(nil) expected error")
	}
	if err := client.StoreDeploymentArtifact(ctx, &DeploymentArtifact{
		AppName:  "myapp",
		ImageRef: "registry.fly.io/myapp:v1",
		Status:   "success",
	}); err == nil || !strings.Contains(err.Error(), "version") {
		t.Fatalf("StoreDeploymentArtifact missing version err = %v", err)
	}
}

func TestFlyDeploymentClient_RecordRollbackAndHistory(t *testing.T) {
	client := NewFlyDeploymentClientWithBase("token", "http://localhost")
	ctx := context.Background()

	if err := client.StoreDeploymentArtifact(ctx, &DeploymentArtifact{
		AppName:    "myapp",
		ImageRef:   "registry.fly.io/myapp:v1",
		Version:    "v1",
		DeployedAt: time.Now().UTC().Add(-time.Minute),
		Status:     "success",
	}); err != nil {
		t.Fatalf("StoreDeploymentArtifact = %v", err)
	}

	if err := client.RecordRollback(ctx, "myapp", "v2", "v1", true); err != nil {
		t.Fatalf("RecordRollback = %v", err)
	}
	if err := client.RecordRollback(ctx, "myapp", "v3", "v2", false); err != nil {
		t.Fatalf("RecordRollback(failed) = %v", err)
	}

	history, err := client.GetRollbackHistory(ctx, "myapp", 10)
	if err != nil {
		t.Fatalf("GetRollbackHistory = %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("GetRollbackHistory len = %d, want 2", len(history))
	}
	if history[0].Status != "rollback_failed" && history[0].Status != "rollback" {
		t.Fatalf("unexpected rollback status: %s", history[0].Status)
	}
}
