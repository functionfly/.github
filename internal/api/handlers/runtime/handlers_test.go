package runtime

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

// ─── ListRuntimes Tests ───────────────────────────────────────────────────────

func TestListRuntimes_IncludesSwift(t *testing.T) {
	t.Parallel()
	h := New()
	req := httptest.NewRequest(http.MethodGet, "/runtimes", nil)
	w := httptest.NewRecorder()

	h.ListRuntimes(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Runtimes []RuntimeInfo `json:"runtimes"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	var swift *RuntimeInfo
	for i := range resp.Runtimes {
		if resp.Runtimes[i].ID == "swift" {
			swift = &resp.Runtimes[i]
			break
		}
	}

	if swift == nil {
		t.Fatal("Swift runtime not found in ListRuntimes response")
	}
	if swift.Status != "stable" {
		t.Errorf("expected Swift status 'stable', got %q", swift.Status)
	}
	if swift.Name != "Swift" {
		t.Errorf("expected name 'Swift', got %q", swift.Name)
	}
	if swift.Version != "5.9+" {
		t.Errorf("expected version '5.9+', got %q", swift.Version)
	}
	if swift.MemoryLimit != 2048 {
		t.Errorf("expected memory limit 2048, got %d", swift.MemoryLimit)
	}
	if swift.Timeout != 300000 {
		t.Errorf("expected timeout 300000, got %d", swift.Timeout)
	}
}

func TestListRuntimes_SwiftFeatures(t *testing.T) {
	t.Parallel()
	h := New()
	req := httptest.NewRequest(http.MethodGet, "/runtimes", nil)
	w := httptest.NewRecorder()

	h.ListRuntimes(w, req)

	var resp struct {
		Runtimes []RuntimeInfo `json:"runtimes"`
	}
	json.NewDecoder(w.Body).Decode(&resp)

	for _, rt := range resp.Runtimes {
		if rt.ID == "swift" {
			requiredFeatures := []string{"SwiftWasm", "WASM sandbox"}
			for _, f := range requiredFeatures {
				found := false
				for _, feat := range rt.Features {
					if feat == f {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Swift runtime missing feature %q", f)
				}
			}
			return
		}
	}
	t.Fatal("Swift runtime not found")
}

// ─── GetRuntimeInfo Tests ─────────────────────────────────────────────────────

func TestGetRuntimeInfo_Swift(t *testing.T) {
	t.Parallel()
	h := New()
	router := mux.NewRouter()
	router.HandleFunc("/runtimes/{id}", h.GetRuntimeInfo)

	req := httptest.NewRequest(http.MethodGet, "/runtimes/swift", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var info RuntimeInfo
	if err := json.NewDecoder(w.Body).Decode(&info); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if info.ID != "swift" {
		t.Errorf("expected ID 'swift', got %q", info.ID)
	}
	if info.Status != "stable" {
		t.Errorf("expected status 'stable', got %q", info.Status)
	}
	if info.Name != "Swift" {
		t.Errorf("expected name 'Swift', got %q", info.Name)
	}
}

func TestGetRuntimeInfo_SwiftNotFound_NoLonger404(t *testing.T) {
	t.Parallel()
	h := New()
	router := mux.NewRouter()
	router.HandleFunc("/runtimes/{id}", h.GetRuntimeInfo)

	req := httptest.NewRequest(http.MethodGet, "/runtimes/swift", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Before the fix, Swift was missing from the map and returned 404.
	// Now it should return 200.
	if w.Code == http.StatusNotFound {
		t.Fatal("Swift runtime should NOT return 404 (was previously broken)")
	}
}

func TestGetRuntimeInfo_Nonexistent(t *testing.T) {
	t.Parallel()
	h := New()
	router := mux.NewRouter()
	router.HandleFunc("/runtimes/{id}", h.GetRuntimeInfo)

	req := httptest.NewRequest(http.MethodGet, "/runtimes/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for nonexistent runtime, got %d", w.Code)
	}
}

// ─── ListToolchains Tests ─────────────────────────────────────────────────────

func TestListToolchains_SwiftPresent(t *testing.T) {
	t.Parallel()
	h := New()
	req := httptest.NewRequest(http.MethodGet, "/runtimes/toolchains", nil)
	w := httptest.NewRecorder()

	h.ListToolchains(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Toolchains []struct {
			Name      string `json:"name"`
			Language  string `json:"language"`
			Available bool   `json:"available"`
			Toolchain string `json:"toolchain"`
		} `json:"toolchains"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	var swift *struct {
		Name      string `json:"name"`
		Language  string `json:"language"`
		Available bool   `json:"available"`
		Toolchain string `json:"toolchain"`
	}
	for i := range resp.Toolchains {
		if resp.Toolchains[i].Language == "swift" {
			swift = &resp.Toolchains[i]
			break
		}
	}

	if swift == nil {
		t.Fatal("Swift toolchain not found in ListToolchains response")
	}
	if swift.Toolchain != "swiftwasm" {
		t.Errorf("expected toolchain 'swiftwasm', got %q", swift.Toolchain)
	}
	// Available depends on whether carton/swiftc is installed — just verify it's a bool
	t.Logf("Swift toolchain available: %v", swift.Available)
}

// ─── Nil Registry Guard Tests ─────────────────────────────────────────────────

func TestGetDiagnostics_NilRegistry(t *testing.T) {
	t.Parallel()
	h := New() // registry is nil
	router := mux.NewRouter()
	router.HandleFunc("/functions/{function_id}/diagnostics", h.GetDiagnostics)

	req := httptest.NewRequest(http.MethodGet, "/functions/test-fn/diagnostics", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 500, not panic
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for nil registry, got %d", w.Code)
	}
}

func TestUpdateRuntimeConfig_NilRegistry(t *testing.T) {
	t.Parallel()
	h := New() // registry is nil
	router := mux.NewRouter()
	router.HandleFunc("/functions/{function_id}/runtime", h.UpdateRuntimeConfig)

	req := httptest.NewRequest(http.MethodPut, "/functions/test-fn/runtime", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 500, not panic
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for nil registry, got %d", w.Code)
	}
}

// ─── Route Registration Tests ─────────────────────────────────────────────────

func TestRegisterRoutes(t *testing.T) {
	t.Parallel()
	h := New()
	router := mux.NewRouter()
	h.RegisterRoutes(router)

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/runtimes"},
		{http.MethodGet, "/runtimes/swift"},
		{http.MethodGet, "/runtimes/toolchains"},
		{http.MethodGet, "/functions/test/diagnostics"},
		{http.MethodPut, "/functions/test/runtime"},
	}

	for _, rt := range routes {
		req := httptest.NewRequest(rt.method, rt.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		// We just verify the route exists (doesn't 405)
		if w.Code == http.StatusMethodNotAllowed {
			t.Errorf("route %s %s returned 405 MethodNotAllowed", rt.method, rt.path)
		}
	}
}
