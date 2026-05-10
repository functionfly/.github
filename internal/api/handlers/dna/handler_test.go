package dna_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	handler "github.com/functionfly/functionfly/internal/api/handlers/dna"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/dna"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func TestGetProfile_Unauthorized(t *testing.T) {
	svc := &dna.Service{}
	h := handler.NewHandler(svc, logrus.StandardLogger())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/v1/functions/123/dna", nil)

	h.GetProfile(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestListMutations_Unauthorized(t *testing.T) {
	svc := &dna.Service{}
	h := handler.NewHandler(svc, logrus.StandardLogger())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/v1/functions/123/dna/mutations", nil)

	h.ListMutations(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAcceptVariant_Unauthorized(t *testing.T) {
	svc := &dna.Service{}
	h := handler.NewHandler(svc, logrus.StandardLogger())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/v1/functions/123/dna/variants/456/accept",
		bytes.NewBufferString(`{"canary_percentage": 10}`))

	h.AcceptVariant(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRejectVariant_Unauthorized(t *testing.T) {
	svc := &dna.Service{}
	h := handler.NewHandler(svc, logrus.StandardLogger())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/v1/functions/123/dna/variants/456/reject",
		bytes.NewBufferString(`{"reason": "test"}`))

	h.RejectVariant(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestToggleEvolution_Unauthorized(t *testing.T) {
	svc := &dna.Service{}
	h := handler.NewHandler(svc, logrus.StandardLogger())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/v1/functions/123/dna/evolution",
		bytes.NewBufferString(`{"enabled": true}`))

	h.ToggleEvolution(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestGetInsights_Unauthorized(t *testing.T) {
	svc := &dna.Service{}
	h := handler.NewHandler(svc, logrus.StandardLogger())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/v1/functions/123/dna/insights?period=7d", nil)

	h.GetInsights(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestGetEnterpriseInsights_Unauthorized(t *testing.T) {
	svc := &dna.Service{}
	h := handler.NewHandler(svc, logrus.StandardLogger())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/v1/dna/enterprise/insights?period=30d", nil)

	h.GetEnterpriseInsights(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestTriggerAnalysis_RateLimit(t *testing.T) {
	tenantID := uuid.New()
	svc := &dna.Service{}
	h := handler.NewHandler(svc, logrus.StandardLogger())

	claims := &auth.Claims{
		UserID:   uuid.New(),
		TenantID: tenantID,
	}

	// Send 10 requests that pass rate limiter but will fail at service level (nil repo)
	for i := 0; i < 10; i++ {
		func() {
			defer func() { recover() }() // recover from nil repo panic
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/v1/functions/test-func/dna/analyze", nil)
			r = middleware.SetUserInContext(r, claims)
			r = mux.SetURLVars(r, map[string]string{"id": "test-func"})
			h.TriggerAnalysis(w, r)
		}()
	}

	// 11th should be rate limited (429) before hitting the service
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/v1/functions/test-func/dna/analyze", nil)
	r = middleware.SetUserInContext(r, claims)
	r = mux.SetURLVars(r, map[string]string{"id": "test-func"})
	h.TriggerAnalysis(w, r)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("11th request: status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
}

func TestWriteError_Response(t *testing.T) {
	w := httptest.NewRecorder()
	handler.WriteError(w, http.StatusNotFound, "NOT_FOUND", "resource not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["code"] != "NOT_FOUND" {
		t.Errorf("code = %s, want NOT_FOUND", resp["code"])
	}
}
