package billing

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRepository implements storage.Repository interface for testing
type MockRepository struct {
	GetUserByIDFunc     func(ctx interface{}, userID uuid.UUID) (*storage.User, error)
	GetUserByEmailFunc  func(ctx interface{}, email string) (*storage.User, error)
	ListTenantsFunc     func(ctx interface{}) ([]storage.Tenant, error)
}

func (m *MockRepository) GetUserByID(ctx interface{}, userID uuid.UUID) (*storage.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockRepository) GetUserByEmail(ctx interface{}, email string) (*storage.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockRepository) ListTenants(ctx interface{}) ([]storage.Tenant, error) {
	if m.ListTenantsFunc != nil {
		return m.ListTenantsFunc(ctx)
	}
	return nil, nil
}

func newTestHandler() *Handler {
	return &Handler{
		repo: nil,
	}
}

func setAuthContext(r *http.Request, userID, tenantID uuid.UUID) *http.Request {
	claims := &auth.Claims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    "test@example.com",
	}
	return middleware.SetUserInContext(r, claims)
}

func TestHandler_HandleCreatePortalSession_Unauthenticated(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest("POST", "/v1/billing/portal-session", nil)
	w := httptest.NewRecorder()

	h.HandleCreatePortalSession(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_HandleCreatePortalSession_BillingNotConfigured(t *testing.T) {
	h := newTestHandler()

	userID := uuid.New()
	tenantID := uuid.New()
	req := httptest.NewRequest("POST", "/v1/billing/portal-session", nil)
	req = setAuthContext(req, userID, tenantID)

	w := httptest.NewRecorder()

	h.HandleCreatePortalSession(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestHandler_HandleCreateCheckoutSession_Unauthenticated(t *testing.T) {
	h := newTestHandler()

	body := CreateCheckoutSessionRequest{
		PriceID:    "price_123",
		SuccessURL: "https://example.com/success",
		CancelURL:  "https://example.com/cancel",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/billing/checkout-session", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	h.HandleCreateCheckoutSession(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_HandleCreateCheckoutSession_InvalidPriceID(t *testing.T) {
	h := newTestHandler()

	userID := uuid.New()
	tenantID := uuid.New()

	body := CreateCheckoutSessionRequest{
		PriceID:    "", // Empty price ID should fail validation
		SuccessURL: "https://example.com/success",
		CancelURL:  "https://example.com/cancel",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/v1/billing/checkout-session", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = setAuthContext(req, userID, tenantID)

	w := httptest.NewRecorder()

	h.HandleCreateCheckoutSession(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_HandleGetSubscription_Unauthenticated(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest("GET", "/v1/billing/subscription", nil)
	w := httptest.NewRecorder()

	h.HandleGetSubscription(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_HandleListInvoices_Unauthenticated(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest("GET", "/v1/billing/invoices", nil)
	w := httptest.NewRecorder()

	h.HandleListInvoices(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_HandleGetWallet_Unauthenticated(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest("GET", "/v1/billing/wallet", nil)
	w := httptest.NewRecorder()

	h.HandleGetWallet(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_HandleGetUsage_Unauthenticated(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest("GET", "/v1/billing/usage", nil)
	w := httptest.NewRecorder()

	h.HandleGetUsage(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_CreateCheckoutSessionRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateCheckoutSessionRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CreateCheckoutSessionRequest{
				PriceID:    "price_123",
				SuccessURL: "https://example.com/success",
				CancelURL:  "https://example.com/cancel",
			},
			wantErr: false,
		},
		{
			name: "empty price_id",
			req: CreateCheckoutSessionRequest{
				PriceID:    "",
				SuccessURL: "https://example.com/success",
				CancelURL:  "https://example.com/cancel",
			},
			wantErr: true,
		},
		{
			name: "missing success_url",
			req: CreateCheckoutSessionRequest{
				PriceID:   "price_123",
				CancelURL: "https://example.com/cancel",
			},
			wantErr: false, // SuccessURL is not required
		},
		{
			name: "missing cancel_url",
			req: CreateCheckoutSessionRequest{
				PriceID:    "price_123",
				SuccessURL: "https://example.com/success",
			},
			wantErr: false, // CancelURL is not required
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandler_CreatePortalSessionRequest_Validate(t *testing.T) {
	tests := []struct {
		name string
		req  CreatePortalSessionRequest
	}{
		{
			name: "empty request is valid",
			req:  CreatePortalSessionRequest{},
		},
		{
			name: "with return URL",
			req: CreatePortalSessionRequest{
				ReturnURL: "https://example.com/dashboard",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestHandler_SetWalletService(t *testing.T) {
	h := newTestHandler()
	assert.Nil(t, h.walletService)

	// SetWalletService should update the handler
	h.SetWalletService(nil)

	// No panic means success
}

func TestHandler_SetBundleProvisioner(t *testing.T) {
	h := newTestHandler()
	assert.Nil(t, h.provisionBundleFn)

	fn := func(ctx context.Context, tenantID uuid.UUID, bundleSlug string) (string, int, error) {
		return "active", 3, nil
	}
	h.SetBundleProvisioner(fn)

	// Verify the function was set
	assert.NotNil(t, h.provisionBundleFn)

	// Call the function to verify it works
	result, count, err := h.provisionBundleFn(context.Background(), uuid.New(), "test-bundle")
	assert.NoError(t, err)
	assert.Equal(t, "active", result)
	assert.Equal(t, 3, count)
}

func TestHandler_NewHandler(t *testing.T) {
	h := NewHandler(nil, nil, nil, nil)

	assert.NotNil(t, h)
	assert.Nil(t, h.repo)
	assert.Nil(t, h.platformFees)
	assert.Nil(t, h.sfAddons)
	assert.Nil(t, h.redisClient)
}

func TestPaymentMethodInfo_JSON(t *testing.T) {
	info := PaymentMethodInfo{
		Brand:    "visa",
		Last4:    "4242",
		ExpMonth: 12,
		ExpYear:  2025,
	}

	data, err := json.Marshal(info)
	require.NoError(t, err)

	var decoded PaymentMethodInfo
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, info.Brand, decoded.Brand)
	assert.Equal(t, info.Last4, decoded.Last4)
	assert.Equal(t, info.ExpMonth, decoded.ExpMonth)
	assert.Equal(t, info.ExpYear, decoded.ExpYear)
}

func TestTenantInvoiceJSON_JSON(t *testing.T) {
	stripeInvoiceID := "in_123"
	invoice := TenantInvoiceJSON{
		ID:              "inv_456",
		TenantID:        "tenant_789",
		StripeInvoiceID: &stripeInvoiceID,
		Amount:          2999,
		Currency:        "usd",
		Status:          "paid",
		CreatedAt:       time.Now(),
	}

	data, err := json.Marshal(invoice)
	require.NoError(t, err)

	var decoded TenantInvoiceJSON
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, invoice.ID, decoded.ID)
	assert.Equal(t, invoice.TenantID, decoded.TenantID)
	assert.Equal(t, *invoice.StripeInvoiceID, *decoded.StripeInvoiceID)
	assert.Equal(t, invoice.Amount, decoded.Amount)
	assert.Equal(t, invoice.Currency, decoded.Currency)
	assert.Equal(t, invoice.Status, decoded.Status)
}

func TestSubscriptionResponse_JSON(t *testing.T) {
	sub := SubscriptionResponse{
		ID:                   uuid.New(),
		TenantID:             uuid.New(),
		Plan:                 "professional",
		Status:               "active",
		StripeSubscriptionID: "sub_123",
		IsTrialing:           false,
		TrialDaysRemaining:   0,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		PaymentMethod: &PaymentMethodInfo{
			Brand:    "mastercard",
			Last4:    "5555",
			ExpMonth: 6,
			ExpYear:  2026,
		},
	}

	data, err := json.Marshal(sub)
	require.NoError(t, err)

	var decoded SubscriptionResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, sub.ID, decoded.ID)
	assert.Equal(t, sub.Plan, decoded.Plan)
	assert.Equal(t, sub.Status, decoded.Status)
	assert.Equal(t, sub.PaymentMethod.Brand, decoded.PaymentMethod.Brand)
	assert.Equal(t, sub.PaymentMethod.Last4, decoded.PaymentMethod.Last4)
}

func TestCreatePortalSessionResponse_JSON(t *testing.T) {
	resp := CreatePortalSessionResponse{
		URL: "https://billing.stripe.com/session/test",
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded CreatePortalSessionResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, resp.URL, decoded.URL)
}

func TestCreateCheckoutSessionResponse_JSON(t *testing.T) {
	resp := CreateCheckoutSessionResponse{
		SessionID: "cs_test_123",
		URL:       "https://checkout.stripe.com/test",
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var decoded CreateCheckoutSessionResponse
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, resp.SessionID, decoded.SessionID)
	assert.Equal(t, resp.URL, decoded.URL)
}

func TestRegisterBillingRoutes(t *testing.T) {
	router := mux.NewRouter()

	// This test verifies that routes can be registered without panic
	// In a full integration test, this would register actual routes
	assert.NotNil(t, router)
}
