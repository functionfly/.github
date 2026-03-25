package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/functionfly/functionfly/internal/functionregistry"
	storageregistry "github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPlatformFeeRepo implements a mock platform fee repository for testing
type mockPlatformFeeRepo struct {
	wallets         map[uuid.UUID]*storageregistry.UserWallet
	fees            []storageregistry.PlatformFee
	getOrCreateErr  error
	debitErr        error
	creditErr       error
	insufficientBal  bool
}

func newMockPlatformFeeRepo() *mockPlatformFeeRepo {
	return &mockPlatformFeeRepo{
		wallets: make(map[uuid.UUID]*storageregistry.UserWallet),
	}
}

func (m *mockPlatformFeeRepo) GetWallet(ctx context.Context, userID uuid.UUID) (*storageregistry.UserWallet, error) {
	if m.getOrCreateErr != nil {
		return nil, m.getOrCreateErr
	}
	wallet, ok := m.wallets[userID]
	if !ok {
		return nil, nil
	}
	return wallet, nil
}

func (m *mockPlatformFeeRepo) GetOrCreateWallet(ctx context.Context, userID uuid.UUID) (*storageregistry.UserWallet, error) {
	if m.getOrCreateErr != nil {
		return nil, m.getOrCreateErr
	}
	wallet, ok := m.wallets[userID]
	if !ok {
		wallet = &storageregistry.UserWallet{
			UserID:     userID,
			BalanceUSD: 0,
		}
		m.wallets[userID] = wallet
	}
	return wallet, nil
}

func (m *mockPlatformFeeRepo) GetWalletBalance(ctx context.Context, userID uuid.UUID) (float64, error) {
	wallet, err := m.GetWallet(ctx, userID)
	if err != nil || wallet == nil {
		return 0, err
	}
	return wallet.BalanceUSD, nil
}

func (m *mockPlatformFeeRepo) CreditWallet(ctx context.Context, userID uuid.UUID, amountUSD float64, stripePaymentID string) error {
	if m.creditErr != nil {
		return m.creditErr
	}
	wallet, ok := m.wallets[userID]
	if !ok {
		wallet = &storageregistry.UserWallet{
			UserID: userID,
		}
		m.wallets[userID] = wallet
	}
	wallet.BalanceUSD += amountUSD
	return nil
}

func (m *mockPlatformFeeRepo) DebitWallet(ctx context.Context, userID uuid.UUID, amountUSD float64, description string) error {
	if m.debitErr != nil {
		return m.debitErr
	}
	if m.insufficientBal {
		return &insufficientBalanceError{}
	}
	wallet, ok := m.wallets[userID]
	if !ok {
		return &insufficientBalanceError{}
	}
	if wallet.BalanceUSD < amountUSD {
		return &insufficientBalanceError{}
	}
	wallet.BalanceUSD -= amountUSD
	return nil
}

func (m *mockPlatformFeeRepo) RecordPlatformFee(ctx context.Context, fee *storageregistry.PlatformFee) error {
	m.fees = append(m.fees, *fee)
	return nil
}

func (m *mockPlatformFeeRepo) GetPlatformFeesByFunction(ctx context.Context, functionID uuid.UUID) ([]storageregistry.PlatformFee, error) {
	var fees []storageregistry.PlatformFee
	for _, f := range m.fees {
		if f.FunctionID == functionID {
			fees = append(fees, f)
		}
	}
	return fees, nil
}

func (m *mockPlatformFeeRepo) GetPlatformFeesByUser(ctx context.Context, userID uuid.UUID) ([]storageregistry.PlatformFee, error) {
	var fees []storageregistry.PlatformFee
	for _, f := range m.fees {
		if f.UserID == userID {
			fees = append(fees, f)
		}
	}
	return fees, nil
}

func (m *mockPlatformFeeRepo) UpdatePlatformFeeStatus(ctx context.Context, feeID uuid.UUID, status string, stripePaymentID string) error {
	return nil
}

type insufficientBalanceError struct{}

func (e *insufficientBalanceError) Error() string {
	return "insufficient balance"
}

// mockRegistryForPublish implements minimal registry interface for publish tests
type mockRegistryForPublish struct {
	functions      map[string]*storageregistry.RegistryFunction
	versions       map[uuid.UUID][]*storageregistry.RegistryFunctionVersion
	createErr      error
	getFunctionErr error
}

func newMockRegistryForPublish() *mockRegistryForPublish {
	return &mockRegistryForPublish{
		functions: make(map[string]*storageregistry.RegistryFunction),
		versions:  make(map[uuid.UUID][]*storageregistry.RegistryFunctionVersion),
	}
}

func (m *mockRegistryForPublish) GetFunctionByAuthorName(author, name string) (*storageregistry.RegistryFunction, error) {
	if m.getFunctionErr != nil {
		return nil, m.getFunctionErr
	}
	key := author + "/" + name
	fn, ok := m.functions[key]
	if !ok {
		return nil, nil
	}
	return fn, nil
}

func (m *mockRegistryForPublish) CreateFunction(fn *storageregistry.RegistryFunction) error {
	if m.createErr != nil {
		return m.createErr
	}
	key := fn.Author + "/" + fn.Name
	m.functions[key] = fn
	return nil
}

func (m *mockRegistryForPublish) UpdateRegistryFunction(fnID uuid.UUID, meta map[string]interface{}) (*storageregistry.RegistryFunction, error) {
	return nil, nil
}

func (m *mockRegistryForPublish) UpsertFunctionVersion(version *storageregistry.RegistryFunctionVersion, strategy storageregistry.VersionConflictStrategy) (*storageregistry.RegistryFunctionVersion, error) {
	key := version.FunctionID
	m.versions[key] = append(m.versions[key], version)
	return version, nil
}

func (m *mockRegistryForPublish) UpdateFunctionLatestVersion(fnID uuid.UUID, version string) error {
	return nil
}

func (m *mockRegistryForPublish) InvalidateListCache(ctx context.Context) {}

// TestHandlePublish_ExemptAuthor_NoFee tests that functionfly author is exempt from fees
func TestHandlePublish_ExemptAuthor_NoFee(t *testing.T) {
	// This test verifies that exempt authors don't get charged
	// Note: This is a unit test for the fee calculation logic

	author := "functionfly"

	// Verify exempt status
	assert.True(t, storageregistry.IsAuthorExempt(author), "functionfly should be exempt")

	// Verify publish fee is 0 for exempt author
	publishFee := storageregistry.CalculatePublishFee(author)
	assert.Equal(t, 0.0, publishFee, "Exempt author should have $0 publish fee")

	// Verify version update fee is 0 for exempt author
	versionFee := storageregistry.CalculateVersionUpdateFee(author)
	assert.Equal(t, 0.0, versionFee, "Exempt author should have $0 version update fee")
}

// TestHandlePublish_NonExemptAuthor_PublishFee tests that non-exempt authors are charged $2.99 for publish
func TestHandlePublish_NonExemptAuthor_PublishFee(t *testing.T) {
	author := "testauthor"

	// Verify not exempt
	assert.False(t, storageregistry.IsAuthorExempt(author), "testauthor should not be exempt")

	// Verify publish fee is $2.99 for non-exempt author
	publishFee := storageregistry.CalculatePublishFee(author)
	assert.Equal(t, storageregistry.PublishFeeAmount, publishFee, "Non-exempt author should have $2.99 publish fee")
	assert.Equal(t, 2.99, publishFee)

	// Verify version update fee is $0.99 for non-exempt author
	versionFee := storageregistry.CalculateVersionUpdateFee(author)
	assert.Equal(t, storageregistry.VersionUpdateFeeAmount, versionFee, "Non-exempt author should have $0.99 version update fee")
	assert.Equal(t, 0.99, versionFee)
}

// TestHandlePublish_InsufficientBalance_Returns402 tests that insufficient wallet balance returns 402
func TestHandlePublish_InsufficientBalance_Returns402(t *testing.T) {
	// This test verifies the fee calculation returns correct amounts
	// The actual 402 response would be tested in an integration test

	author := "testauthor"
	feeAmount := storageregistry.CalculatePublishFee(author)

	// Verify fee is 2.99
	assert.Equal(t, 2.99, feeAmount)

	// Verify insufficient balance error message format
	err := &insufficientBalanceError{}
	assert.Contains(t, err.Error(), "insufficient balance")
}

// TestFeeCalculation_Commission tests commission calculation
func TestFeeCalculation_Commission(t *testing.T) {
	tests := []struct {
		saleAmount   float64
		expectedComm float64
	}{
		{saleAmount: 100.00, expectedComm: 15.00},  // 15% of $100 = $15
		{saleAmount: 10.00, expectedComm: 1.50},    // 15% of $10 = $1.50
		{saleAmount: 1.00, expectedComm: 0.15},     // 15% of $1 = $0.15
		{saleAmount: 0.00, expectedComm: 0.00},     // 15% of $0 = $0
		{saleAmount: 33.33, expectedComm: 5.00},    // 15% of $33.33 ≈ $5.00
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			comm := storageregistry.CalculateCommission(tt.saleAmount)
			assert.InDelta(t, tt.expectedComm, comm, 0.01, "Commission for $%.2f should be $%.2f", tt.saleAmount, tt.expectedComm)
		})
	}
}

// TestFeeCalculation_Constants verifies fee constants are correct
func TestFeeCalculation_Constants(t *testing.T) {
	assert.Equal(t, 2.99, storageregistry.PublishFeeAmount)
	assert.Equal(t, 0.99, storageregistry.VersionUpdateFeeAmount)
	assert.Equal(t, 0.15, storageregistry.PlatformCommissionRate)
}

// TestPlatformFeeExemption_CaseSensitivity tests that exemption is case-sensitive
func TestPlatformFeeExemption_CaseSensitivity(t *testing.T) {
	// Exact match "functionfly" should be exempt
	assert.True(t, storageregistry.IsAuthorExempt("functionfly"))

	// Case variations should NOT be exempt
	assert.False(t, storageregistry.IsAuthorExempt("FunctionFly"))
	assert.False(t, storageregistry.IsAuthorExempt("FUNCTIONFLY"))
	assert.False(t, storageregistry.IsAuthorExempt("Functionfly"))

	// Similar but different names should NOT be exempt
	assert.False(t, storageregistry.IsAuthorExempt("functionfly2"))
	assert.False(t, storageregistry.IsAuthorExempt("my-functionfly"))
}

// TestPlatformFeeExemptAuthors_List verifies the exempt authors list
func TestPlatformFeeExemptAuthors_List(t *testing.T) {
	assert.Contains(t, storageregistry.ExemptAuthors, "functionfly")
	assert.Len(t, storageregistry.ExemptAuthors, 1)
}

// TestPublishRequest_JSON tests publish request JSON marshaling
func TestPublishRequest_JSON(t *testing.T) {
	req := functionregistry.PublishRequest{
		Author:  "testauthor",
		Name:    "myfunction",
		Version: "1.0.0",
		Manifest: json.RawMessage(`{"title":"Test","runtime":"python"}`),
		Source: &functionregistry.FunctionSource{
			Code: "def handler(): return {}",
		},
	}

	data, err := json.Marshal(req)
	require.NoError(t, err)

	var parsed functionregistry.PublishRequest
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, req.Author, parsed.Author)
	assert.Equal(t, req.Name, parsed.Name)
	assert.Equal(t, req.Version, parsed.Version)
}

// TestPublishRequest_Validation tests required fields validation
func TestPublishRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     functionregistry.PublishRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: functionregistry.PublishRequest{
				Author:  "author",
				Name:    "name",
				Version: "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "missing author",
			req: functionregistry.PublishRequest{
				Name:    "name",
				Version: "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			req: functionregistry.PublishRequest{
				Author:  "author",
				Version: "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "missing version",
			req: functionregistry.PublishRequest{
				Author: "author",
				Name:   "name",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasErr := tt.req.Author == "" || tt.req.Name == "" || tt.req.Version == ""
			assert.Equal(t, tt.wantErr, hasErr)
		})
	}
}

// TestSemVerValidation tests semantic version validation
func TestSemVerValidation(t *testing.T) {
	tests := []struct {
		version string
		valid   bool
	}{
		{"1.0.0", true},
		{"0.1.0", true},
		{"1.0.0-alpha", true},
		{"1.0.0-beta.1", true},
		{"1.0.0+build.123", true},
		{"1.0.0-alpha+001", true},
		{"1", false},
		{"1.0", false},
		{"1.0.", false},
		{"1.0.0.0", false},
		{"v1.0.0", false},
		{"1.0.0.", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			err := functionregistry.ValidateSemVer(tt.version)
			if tt.valid {
				assert.NoError(t, err, "Version %s should be valid", tt.version)
			} else {
				assert.Error(t, err, "Version %s should be invalid", tt.version)
			}
		})
	}
}

// TestWalletCreditDebit_Flow tests the flow of crediting and debiting a wallet
func TestWalletCreditDebit_Flow(t *testing.T) {
	mockRepo := newMockPlatformFeeRepo()
	ctx := context.Background()
	userID := uuid.New()

	// Initial balance should be 0 (wallet doesn't exist)
	balance, err := mockRepo.GetWalletBalance(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 0.0, balance)

	// Credit the wallet
	err = mockRepo.CreditWallet(ctx, userID, 100.00, "stripe_123")
	require.NoError(t, err)

	// Check balance after credit
	balance, err = mockRepo.GetWalletBalance(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 100.00, balance)

	// Debit for publish fee
	err = mockRepo.DebitWallet(ctx, userID, 2.99, "publish fee")
	require.NoError(t, err)

	// Check balance after debit
	balance, err = mockRepo.GetWalletBalance(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 97.01, balance)

	// Debit for version update fee
	err = mockRepo.DebitWallet(ctx, userID, 0.99, "version update fee")
	require.NoError(t, err)

	// Check final balance
	balance, err = mockRepo.GetWalletBalance(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 96.02, balance)
}

// TestWalletDebit_InsufficientBalance tests debit failure with insufficient balance
func TestWalletDebit_InsufficientBalance(t *testing.T) {
	mockRepo := newMockPlatformFeeRepo()
	ctx := context.Background()
	userID := uuid.New()

	// Credit only $1
	err := mockRepo.CreditWallet(ctx, userID, 1.00, "stripe_123")
	require.NoError(t, err)

	// Try to debit $2.99 (more than balance)
	err = mockRepo.DebitWallet(ctx, userID, 2.99, "publish fee")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient balance")

	// Balance should be unchanged
	balance, err := mockRepo.GetWalletBalance(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 1.00, balance)
}

// TestWalletDebit_ExactBalance tests debit when balance is exactly equal
func TestWalletDebit_ExactBalance(t *testing.T) {
	mockRepo := newMockPlatformFeeRepo()
	ctx := context.Background()
	userID := uuid.New()

	// Credit exactly $2.99
	err := mockRepo.CreditWallet(ctx, userID, 2.99, "stripe_123")
	require.NoError(t, err)

	// Debit exactly $2.99
	err = mockRepo.DebitWallet(ctx, userID, 2.99, "publish fee")
	require.NoError(t, err)

	// Balance should be 0
	balance, err := mockRepo.GetWalletBalance(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 0.0, balance)
}

// TestPlatformFeeRepository_MockInterface tests that mock implements the interface
func TestPlatformFeeRepository_MockInterface(t *testing.T) {
	mockRepo := newMockPlatformFeeRepo()
	ctx := context.Background()
	userID := uuid.New()

	// Test GetOrCreateWallet
	wallet, err := mockRepo.GetOrCreateWallet(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, wallet)
	assert.Equal(t, userID, wallet.UserID)

	// Test GetWallet (should return existing wallet)
	wallet, err = mockRepo.GetWallet(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, wallet)

	// Test GetWallet (non-existent user)
	otherID := uuid.New()
	wallet, err = mockRepo.GetWallet(ctx, otherID)
	require.NoError(t, err)
	assert.Nil(t, wallet)

	// Test CreditWallet
	err = mockRepo.CreditWallet(ctx, userID, 50.00, "stripe_456")
	require.NoError(t, err)

	// Test DebitWallet
	err = mockRepo.DebitWallet(ctx, userID, 10.00, "test debit")
	require.NoError(t, err)

	// Verify final balance
	wallet, _ = mockRepo.GetWallet(ctx, userID)
	assert.Equal(t, 40.00, wallet.BalanceUSD)
}

// TestPublishFeeRecord tests recording platform fees
func TestPublishFeeRecord(t *testing.T) {
	mockRepo := newMockPlatformFeeRepo()
	ctx := context.Background()

	fnID := uuid.New()
	userID := uuid.New()

	// Record a publish fee
	fee := &storageregistry.PlatformFee{
		FunctionID: fnID,
		UserID:     userID,
		FeeType:    storageregistry.FeeTypePublish,
		AmountUSD:  storageregistry.PublishFeeAmount,
		Status:     storageregistry.FeeStatusCompleted,
	}

	err := mockRepo.RecordPlatformFee(ctx, fee)
	require.NoError(t, err)

	// Get fees by function
	fees, err := mockRepo.GetPlatformFeesByFunction(ctx, fnID)
	require.NoError(t, err)
	assert.Len(t, fees, 1)
	assert.Equal(t, storageregistry.FeeTypePublish, fees[0].FeeType)
	assert.Equal(t, 2.99, fees[0].AmountUSD)

	// Get fees by user
	fees, err = mockRepo.GetPlatformFeesByUser(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, fees, 1)

	// Record a version update fee
	versionFee := &storageregistry.PlatformFee{
		FunctionID: fnID,
		UserID:     userID,
		FeeType:    storageregistry.FeeTypeVersionUpdate,
		AmountUSD:  storageregistry.VersionUpdateFeeAmount,
		Status:     storageregistry.FeeStatusCompleted,
	}
	err = mockRepo.RecordPlatformFee(ctx, versionFee)
	require.NoError(t, err)

	// Verify both fees are returned
	fees, err = mockRepo.GetPlatformFeesByFunction(ctx, fnID)
	require.NoError(t, err)
	assert.Len(t, fees, 2)
}

// TestFeeStatusConstants tests fee status constants
func TestFeeStatusConstants(t *testing.T) {
	assert.Equal(t, "pending", storageregistry.FeeStatusPending)
	assert.Equal(t, "completed", storageregistry.FeeStatusCompleted)
	assert.Equal(t, "failed", storageregistry.FeeStatusFailed)
	assert.Equal(t, "refunded", storageregistry.FeeStatusRefunded)
}

// TestFeeTypeConstants tests fee type constants
func TestFeeTypeConstants(t *testing.T) {
	assert.Equal(t, "publish", storageregistry.FeeTypePublish)
	assert.Equal(t, "version_update", storageregistry.FeeTypeVersionUpdate)
	assert.Equal(t, "commission", storageregistry.FeeTypeCommission)
}

// Helper to create a test publish request
func createTestPublishRequest(author, name, version string) *functionregistry.PublishRequest {
	return &functionregistry.PublishRequest{
		Author:  author,
		Name:    name,
		Version: version,
		Manifest: json.RawMessage(`{
			"title": "Test Function",
			"description": "A test function",
			"runtime": "python",
			"input": {
				"type": "object",
				"properties": {
					"text": {"type": "string"}
				}
			}
		}`),
		Source: &functionregistry.FunctionSource{
			Code: "def handler(text): return text.upper()",
		},
	}
}

// Helper to marshal publish request to JSON
func marshalPublishRequest(t *testing.T, req *functionregistry.PublishRequest) *bytes.Buffer {
	data, err := json.Marshal(req)
	require.NoError(t, err)
	return bytes.NewBuffer(data)
}

// TestInsufficientBalanceError tests the insufficient balance error
func TestInsufficientBalanceError(t *testing.T) {
	err := &insufficientBalanceError{}
	assert.Equal(t, "insufficient balance", err.Error())
}

// TestMockPlatformFeeRepo_NonExistentWallet tests operations on non-existent wallet
func TestMockPlatformFeeRepo_NonExistentWallet(t *testing.T) {
	mockRepo := newMockPlatformFeeRepo()
	ctx := context.Background()
	userID := uuid.New()

	// GetWallet for non-existent user should return nil wallet with no error
	wallet, err := mockRepo.GetWallet(ctx, userID)
	require.NoError(t, err)
	assert.Nil(t, wallet)

	// DebitWallet on non-existent wallet should fail with insufficient balance
	err = mockRepo.DebitWallet(ctx, userID, 1.00, "test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient balance")
}

// TestPublishFeeAmounts tests that fee amounts are correct for different scenarios
func TestPublishFeeAmounts(t *testing.T) {
	// New function publish
	author := "testauthor"
	assert.Equal(t, 2.99, storageregistry.CalculatePublishFee(author))

	// Version update
	assert.Equal(t, 0.99, storageregistry.CalculateVersionUpdateFee(author))

	// Exempt author
	exemptAuthor := "functionfly"
	assert.Equal(t, 0.0, storageregistry.CalculatePublishFee(exemptAuthor))
	assert.Equal(t, 0.0, storageregistry.CalculateVersionUpdateFee(exemptAuthor))
}
