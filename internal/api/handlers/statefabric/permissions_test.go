package statefabric

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/functionfly/functionfly/internal/plans"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type stubPlanResolver struct {
	plan string
}

func (s stubPlanResolver) GetTenantPlan(_ context.Context, _ uuid.UUID) string {
	return s.plan
}

func TestRequireFabricQuota_FreePlanDenied(t *testing.T) {
	h := &Handler{
		planResolver: stubPlanResolver{plan: plans.PlanFree},
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	ok := h.requireFabricQuota(w, r, uuid.New())
	assert.False(t, ok)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireFabricQuota_EnterpriseUnlimited(t *testing.T) {
	h := &Handler{
		planResolver: stubPlanResolver{plan: plans.PlanEnterprise},
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	ok := h.requireFabricQuota(w, r, uuid.New())
	assert.True(t, ok)
}
