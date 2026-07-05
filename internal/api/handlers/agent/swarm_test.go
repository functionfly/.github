package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/functionfly/functionfly/internal/agent/identity"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestSwarmHandler_requireAgentTenant(t *testing.T) {
	// Create a handler with a mock identity repository
	handler := &SwarmHandler{}

	router := mux.NewRouter()
	router.HandleFunc("/agent/{id}/wallet", func(w http.ResponseWriter, r *http.Request) {
		agentID := mux.Vars(r)["id"]
		if !handler.requireAgentTenant(w, r, agentID) {
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"ok": true}`)); err != nil {
			t.Logf("failed to write response: %v", err)
		}
	}).Methods("GET")

	t.Run("returns 401 when no auth claims in context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/agent/test-agent/wallet", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("returns 400 when agent_id is empty", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/agent//wallet", nil)
		rr := httptest.NewRecorder()

		// Add auth claims to context
		claims := &auth.Claims{
			UserID:   uuid.New(),
			TenantID: uuid.New(),
		}
		req = middleware.SetUserInContext(req, claims)

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

// Mock identity repository for testing
type mockIdentityRepo struct {
	agents map[string]*identity.AgentIdentity
}

func (m *mockIdentityRepo) GetAgent(ctx context.Context, agentID string) (*identity.AgentIdentity, error) {
	agent, ok := m.agents[agentID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return agent, nil
}
