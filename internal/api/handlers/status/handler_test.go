package status

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/auth"
)

func TestDeterminePlatformStatus(t *testing.T) {
	tests := []struct {
		name          string
		healthPercent float64
		incidents     []Incident
		wantStatus    string
		wantIndicator string
	}{
		{
			name:          "Healthy system",
			healthPercent: 100,
			incidents:     []Incident{},
			wantStatus:    "operational",
			wantIndicator: "none",
		},
		{
			name:          "Critical incident",
			healthPercent: 100,
			incidents:     []Incident{{Severity: "critical", Status: "investigating"}},
			wantStatus:    "major_outage",
			wantIndicator: "critical",
		},
		{
			name:          "High severity incident",
			healthPercent: 100,
			incidents:     []Incident{{Severity: "high", Status: "investigating"}},
			wantStatus:    "degraded",
			wantIndicator: "major",
		},
		{
			name:          "Degraded health",
			healthPercent: 70,
			incidents:     []Incident{},
			wantStatus:    "degraded",
			wantIndicator: "major",
		},
		{
			name:          "Low health",
			healthPercent: 40,
			incidents:     []Incident{},
			wantStatus:    "major_outage",
			wantIndicator: "critical",
		},
		{
			name:          "Minor degradation",
			healthPercent: 90,
			incidents:     []Incident{},
			wantStatus:    "degraded",
			wantIndicator: "minor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{}
			status, indicator, _ := h.determinePlatformStatus(tt.healthPercent, tt.incidents)
			if status != tt.wantStatus {
				t.Errorf("determinePlatformStatus() status = %v, want %v", status, tt.wantStatus)
			}
			if indicator != tt.wantIndicator {
				t.Errorf("determinePlatformStatus() indicator = %v, want %v", indicator, tt.wantIndicator)
			}
		})
	}
}

func TestIsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		claims   *auth.Claims
		expected bool
	}{
		{
			name:     "Admin user",
			claims:   &auth.Claims{Role: "admin"},
			expected: true,
		},
		{
			name:     "Super admin user",
			claims:   &auth.Claims{Role: "super_admin"},
			expected: true,
		},
		{
			name:     "Regular user",
			claims:   &auth.Claims{Role: "user"},
			expected: false,
		},
		{
			name:     "Empty role",
			claims:   &auth.Claims{Role: ""},
			expected: false,
		},
		{
			name:     "Guest role",
			claims:   &auth.Claims{Role: "guest"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{}
			result := h.isAdmin(tt.claims)
			if result != tt.expected {
				t.Errorf("isAdmin() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParsePrometheusTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected time.Time
	}{
		{
			name:     "Float64 timestamp",
			input:    float64(1609459200),
			expected: time.Unix(1609459200, 0),
		},
		{
			name:     "String timestamp",
			input:    "1609459200",
			expected: time.Unix(1609459200, 0),
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: time.Now(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePrometheusTimestamp(tt.input)
			if tt.input != nil {
				diff := result.Sub(tt.expected)
				if diff < 0 {
					diff = -diff
				}
				if diff > time.Second {
					t.Errorf("parsePrometheusTimestamp() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestMapBoolStatus(t *testing.T) {
	tests := []struct {
		input    bool
		expected string
	}{
		{true, "operational"},
		{false, "major_outage"},
	}

	for _, tt := range tests {
		result := mapBoolStatus(tt.input)
		if result != tt.expected {
			t.Errorf("mapBoolStatus(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestRespondJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	respondJSON(rr, http.StatusOK, data)

	if rr.Code != http.StatusOK {
		t.Errorf("respondJSON() status = %v, want %v", rr.Code, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("respondJSON() Content-Type = %v, want application/json", contentType)
	}

	var result map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if result["key"] != "value" {
		t.Errorf("respondJSON() body = %v, want %v", result, data)
	}
}

func TestFormatProviderName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"workers", "Cloudflare Workers"},
		{"vercel", "Vercel"},
		{"fly", "Fly.io"},
		{"deno-deploy", "Deno Deploy"},
		{"functionfly-edge", "FunctionFly Edge"},
		{"custom", "Custom"},
		{"", ""},
	}

	for _, tt := range tests {
		result := formatProviderName(tt.input)
		if result != tt.expected {
			t.Errorf("formatProviderName(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestFormatRegionName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"us-east", "US East"},
		{"us-west", "US West"},
		{"eu-west", "EU West"},
		{"eu-central", "EU Central"},
		{"ap-south", "Asia Pacific South"},
		{"ap-northeast", "Asia Pacific Northeast"},
		{"custom-region", "custom-region"},
		{"", ""},
	}

	for _, tt := range tests {
		result := formatRegionName(tt.input)
		if result != tt.expected {
			t.Errorf("formatRegionName(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestCalculateProviderStatus(t *testing.T) {
	tests := []struct {
		healthy   int
		degraded  int
		unhealthy int
		expected  string
	}{
		{10, 0, 0, "operational"},
		{8, 2, 0, "operational"},
		{5, 3, 2, "degraded"},
		{3, 2, 5, "degraded"},
		{0, 0, 10, "outage"},
		{0, 0, 0, "operational"},
		{5, 0, 5, "degraded"},
	}

	for _, tt := range tests {
		result := calculateProviderStatus(tt.healthy, tt.degraded, tt.unhealthy)
		if result != tt.expected {
			t.Errorf("calculateProviderStatus(%d, %d, %d) = %s, want %s",
				tt.healthy, tt.degraded, tt.unhealthy, result, tt.expected)
		}
	}
}

func TestCalculateRegionStatus(t *testing.T) {
	tests := []struct {
		healthy  int
		total    int
		expected string
	}{
		{10, 10, "operational"},
		{9, 10, "degraded_performance"},
		{8, 10, "degraded_performance"},
		{5, 10, "partial_outage"},
		{3, 10, "major_outage"},
		{0, 10, "major_outage"},
		{0, 0, "operational"},
	}

	for _, tt := range tests {
		result := calculateRegionStatus(tt.healthy, tt.total)
		if result != tt.expected {
			t.Errorf("calculateRegionStatus(%d, %d) = %s, want %s",
				tt.healthy, tt.total, result, tt.expected)
		}
	}
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		strs     []string
		sep      string
		expected string
	}{
		{[]string{}, ", ", ""},
		{[]string{"a"}, ", ", "a"},
		{[]string{"a", "b", "c"}, ", ", "a, b, c"},
	}

	for _, tt := range tests {
		result := joinStrings(tt.strs, tt.sep)
		if result != tt.expected {
			t.Errorf("joinStrings(%v, %s) = %s, want %s", tt.strs, tt.sep, result, tt.expected)
		}
	}
}

func TestMapCheckTypeToComponentType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"database", "database"},
		{"api", "api"},
		{"cache", "cache"},
		{"external_service", "provider"},
		{"monitoring", "monitoring"},
		{"disk_space", "infrastructure"},
		{"memory", "infrastructure"},
	}

	for _, tt := range tests {
		result := mapCheckTypeToComponentType(tt.input)
		if result != tt.expected {
			t.Errorf("mapCheckTypeToComponentType(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestMapHealthStatusToComponentStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"healthy", "operational"},
		{"degraded", "degraded_performance"},
		{"unhealthy", "major_outage"},
		{"unknown", "operational"},
		{"", "operational"},
	}

	for _, tt := range tests {
		result := mapHealthStatusToComponentStatus(tt.input)
		if result != tt.expected {
			t.Errorf("mapHealthStatusToComponentStatus(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestParseValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
	}{
		{"string value", "45.2", 45.2},
		{"float64 value", float64(45.2), 45.2},
		{"nil value", nil, 0},
		{"int value", int(42), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseValue(tt.input)
			if result != tt.expected {
				t.Errorf("parseValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPrometheusCache(t *testing.T) {
	cache := &PrometheusCache{
		entries: make(map[string]cacheEntry),
	}

	// Test set and get
	response := &PrometheusResponse{Status: "success"}
	cache.set("test-key", response, 1*time.Minute)

	result := cache.get("test-key")
	if result == nil {
		t.Error("Expected to get cached value")
	}
	if result.Status != "success" {
		t.Error("Cached value doesn't match")
	}

	// Test expired entry
	cache.set("expired-key", response, 1*time.Nanosecond)
	time.Sleep(2 * time.Nanosecond)
	result = cache.get("expired-key")
	if result != nil {
		t.Error("Expected expired entry to return nil")
	}

	// Test non-existent key
	result = cache.get("non-existent")
	if result != nil {
		t.Error("Expected non-existent key to return nil")
	}
}
