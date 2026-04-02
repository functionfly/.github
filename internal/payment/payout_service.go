package payment

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v83"
	"github.com/stripe/stripe-go/v83/account"
	"github.com/stripe/stripe-go/v83/accountlink"
	stripePayout "github.com/stripe/stripe-go/v83/payout"
	"github.com/stripe/stripe-go/v83/transfer"
)

const (
	// Minimum payout amount in cents ($10.00)
	MinPayoutAmountCents = 1000
	// Maximum payout amount in cents ($50,000.00)
	MaxPayoutAmountCents = 5000000
)

// ErrConnectPlatformNotReady is returned when the platform Stripe account has not completed
// Connect onboarding (dashboard.stripe.com/connect). End users see a generic message; ops
// should enable Connect on the platform account.
var ErrConnectPlatformNotReady = errors.New(
	"publisher payouts are not available yet; platform payout setup is still in progress",
)

// PayoutService handles Stripe Connect account management and payouts.
type PayoutService struct {
	payoutRepo *storage.PayoutRepository
}

// NewPayoutService creates a new payout service.
func NewPayoutService(payoutRepo *storage.PayoutRepository) *PayoutService {
	return &PayoutService{payoutRepo: payoutRepo}
}

// OnboardingResult holds the result of creating a Connect account onboarding link.
type OnboardingResult struct {
	AccountID     string `json:"account_id"`
	OnboardingURL string `json:"onboarding_url"`
	Status        string `json:"status"`
}

// PayoutResult holds the result of a payout request.
type PayoutResult struct {
	PayoutRequestID uuid.UUID `json:"payout_request_id"`
	AmountCents     int       `json:"amount_cents"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
}

// ConnectAccountStatus holds the display info for a user's connect account.
type ConnectAccountStatus struct {
	HasAccount      bool    `json:"has_account"`
	AccountID       *string `json:"account_id,omitempty"`
	Status          string  `json:"status"`
	PayoutsEnabled  bool    `json:"payouts_enabled"`
	BankName        *string `json:"bank_name,omitempty"`
	BankLast4       *string `json:"bank_last4,omitempty"`
	Country         *string `json:"country,omitempty"`
	OnboardingURL   *string `json:"onboarding_url,omitempty"`
	NeedsOnboarding bool    `json:"needs_onboarding"`
}

// StartOnboarding creates a Stripe Express connected account and returns an onboarding link.
// This is idempotent — if the user already has a connect account, it refreshes the onboarding URL.
func (s *PayoutService) StartOnboarding(ctx context.Context, userID uuid.UUID, email string) (*OnboardingResult, error) {
	if stripeKey() == "" {
		return nil, fmt.Errorf("Stripe is not configured")
	}

	// Check if user already has a connect account
	existing, err := s.payoutRepo.GetConnectAccountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing account: %w", err)
	}

	if existing != nil {
		// If onboarding is already complete, return current status
		if existing.DetailsSubmitted && existing.PayoutsEnabled {
			return &OnboardingResult{
				AccountID:     existing.StripeAccountID,
				OnboardingURL: "",
				Status:        existing.AccountStatus,
			}, nil
		}

		// Refresh the onboarding link
		url, expiresAt, err := s.createAccountLink(existing.StripeAccountID)
		if err != nil {
			return nil, fmt.Errorf("failed to create onboarding link: %w", err)
		}

		if err := s.payoutRepo.UpdateConnectAccountOnboardingURL(ctx, existing.ID, url, expiresAt); err != nil {
			logrus.WithError(err).Warn("failed to store refreshed onboarding URL")
		}

		return &OnboardingResult{
			AccountID:     existing.StripeAccountID,
			OnboardingURL: url,
			Status:        existing.AccountStatus,
		}, nil
	}

	// Create a new Stripe Express connected account
	stripeAccount, err := s.createConnectedAccount(ctx, email)
	if err != nil {
		if isStripeConnectPlatformNotReady(err) {
			return nil, ErrConnectPlatformNotReady
		}
		return nil, fmt.Errorf("failed to create connected account: %w", err)
	}

	// Create onboarding link
	onboardingURL, expiresAt, err := s.createAccountLink(stripeAccount.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create onboarding link: %w", err)
	}

	// Store in database
	account := &storage.StripeConnectAccount{
		UserID:                 userID,
		StripeAccountID:        stripeAccount.ID,
		AccountStatus:          "onboarding",
		PayoutsEnabled:         stripeAccount.PayoutsEnabled,
		DetailsSubmitted:       stripeAccount.DetailsSubmitted,
		ChargesEnabled:         stripeAccount.ChargesEnabled,
		Currency:               "usd",
		OnboardingURL:          &onboardingURL,
		OnboardingURLExpiresAt: &expiresAt,
	}

	if stripeAccount.Country != "" {
		account.Country = &stripeAccount.Country
	}

	if err := s.payoutRepo.CreateConnectAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to store connect account: %w", err)
	}

	return &OnboardingResult{
		AccountID:     stripeAccount.ID,
		OnboardingURL: onboardingURL,
		Status:        "onboarding",
	}, nil
}

// GetConnectAccountStatus returns the user's connect account status for the dashboard.
func (s *PayoutService) GetConnectAccountStatus(ctx context.Context, userID uuid.UUID) (*ConnectAccountStatus, error) {
	account, err := s.payoutRepo.GetConnectAccountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if account == nil {
		return &ConnectAccountStatus{
			HasAccount:      false,
			Status:          "none",
			NeedsOnboarding: true,
		}, nil
	}

	needsOnboarding := !account.DetailsSubmitted || !account.PayoutsEnabled

	return &ConnectAccountStatus{
		HasAccount:      true,
		AccountID:       &account.StripeAccountID,
		Status:          account.AccountStatus,
		PayoutsEnabled:  account.PayoutsEnabled,
		BankName:        account.BankName,
		BankLast4:       account.BankLast4,
		Country:         account.Country,
		OnboardingURL:   account.OnboardingURL,
		NeedsOnboarding: needsOnboarding,
	}, nil
}

// RequestPayout creates a payout request, debits the user's ledger balance,
// and initiates a Stripe Transfer to the connected account.
func (s *PayoutService) RequestPayout(ctx context.Context, userID uuid.UUID, amountCents int, idempotencyKey string) (*PayoutResult, error) {
	if stripeKey() == "" {
		return nil, fmt.Errorf("Stripe is not configured")
	}

	if amountCents < MinPayoutAmountCents {
		return nil, fmt.Errorf("minimum payout is $%.2f", float64(MinPayoutAmountCents)/100.0)
	}
	if amountCents > MaxPayoutAmountCents {
		return nil, fmt.Errorf("maximum payout is $%.2f", float64(MaxPayoutAmountCents)/100.0)
	}

	// Get connect account
	connectAccount, err := s.payoutRepo.GetConnectAccountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connect account: %w", err)
	}
	if connectAccount == nil {
		return nil, fmt.Errorf("no connected account found; please complete onboarding first")
	}
	if !connectAccount.PayoutsEnabled {
		return nil, fmt.Errorf("payouts are not enabled; please complete onboarding first")
	}

	// Check available balance
	balance, err := s.payoutRepo.GetPayoutBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	availableCents := int(balance.AvailableBalanceUSD * 100)
	if availableCents < amountCents {
		return nil, fmt.Errorf("insufficient balance: available $%.2f, requested $%.2f",
			balance.AvailableBalanceUSD, float64(amountCents)/100.0)
	}

	// Create payout request in pending state
	payoutReq := &storage.PayoutRequest{
		UserID:           userID,
		ConnectAccountID: connectAccount.ID,
		AmountCents:      amountCents,
		Currency:         "usd",
		Status:           "pending",
		IdempotencyKey:   idempotencyKey,
	}
	if err := s.payoutRepo.CreatePayoutRequest(ctx, payoutReq); err != nil {
		return nil, fmt.Errorf("failed to create payout request: %w", err)
	}

	// Debit the user's ledger balance
	if err := s.payoutRepo.DebitForPayout(ctx, userID, amountCents, payoutReq.ID); err != nil {
		return nil, fmt.Errorf("failed to debit balance: %w", err)
	}

	// Update status to processing
	if err := s.payoutRepo.UpdatePayoutRequestStatus(ctx, payoutReq.ID, "processing", nil, nil); err != nil {
		logrus.WithError(err).Error("failed to update payout status to processing")
	}

	// Create Stripe Transfer
	transferParams := &stripe.TransferParams{
		Amount:      stripe.Int64(int64(amountCents)),
		Currency:    stripe.String("usd"),
		Destination: stripe.String(connectAccount.StripeAccountID),
	}
	transferParams.IdempotencyKey = stripe.String(idempotencyKey)

	stripeTransfer, err := transfer.New(transferParams)
	if err != nil {
		// Reverse the debit since transfer failed
		reverseErr := s.payoutRepo.ReversePayoutDebit(ctx, userID, amountCents, payoutReq.ID,
			fmt.Sprintf("Transfer failed: %s", err.Error()))
		if reverseErr != nil {
			logrus.WithError(reverseErr).Error("CRITICAL: failed to reverse payout debit after transfer failure")
		}
		_ = s.payoutRepo.MarkPayoutRequestFailed(ctx, payoutReq.ID, err.Error())
		return nil, fmt.Errorf("stripe transfer failed: %w", err)
	}

	// Update payout request with transfer ID and mark completed
	transferID := stripeTransfer.ID
	if err := s.payoutRepo.UpdatePayoutRequestStatus(ctx, payoutReq.ID, "completed", &transferID, nil); err != nil {
		logrus.WithError(err).Error("failed to update payout request after successful transfer")
	}

	return &PayoutResult{
		PayoutRequestID: payoutReq.ID,
		AmountCents:     amountCents,
		Currency:        "usd",
		Status:          "completed",
	}, nil
}

// RefreshAccountStatus syncs the connect account status from Stripe to our database.
// Called from webhook handlers or on-demand.
func (s *PayoutService) RefreshAccountStatus(ctx context.Context, stripeAccountID string) error {
	stripeAccount, err := account.GetByID(stripeAccountID, nil)
	if err != nil {
		return fmt.Errorf("failed to get stripe account: %w", err)
	}

	// Find our local record by stripe account ID
	// (We need a lookup method; for now we use a direct query)
	var accountID uuid.UUID
	err = s.payoutRepo.GetAccountIDByStripeID(ctx, stripeAccountID, &accountID)
	if err != nil {
		return fmt.Errorf("failed to find local account: %w", err)
	}

	status := "pending"
	switch {
	case stripeAccount.PayoutsEnabled && stripeAccount.DetailsSubmitted:
		status = "active"
	case stripeAccount.DetailsSubmitted && !stripeAccount.PayoutsEnabled:
		status = "restricted"
	case !stripeAccount.DetailsSubmitted:
		status = "onboarding"
	}

	if err := s.payoutRepo.UpdateConnectAccountStatus(ctx, accountID, status,
		stripeAccount.PayoutsEnabled, stripeAccount.DetailsSubmitted, stripeAccount.ChargesEnabled); err != nil {
		return fmt.Errorf("failed to update account status: %w", err)
	}

	// Update bank info if available
	if stripeAccount.ExternalAccounts != nil {
		for _, acct := range stripeAccount.ExternalAccounts.Data {
			if acct.BankAccount != nil {
				last4 := acct.BankAccount.Last4
				bankName := acct.BankAccount.BankName
				var last4Ptr, bankNamePtr *string
				if last4 != "" {
					last4Ptr = &last4
				}
				if bankName != "" {
					bankNamePtr = &bankName
				}
				_ = s.payoutRepo.UpdateConnectAccountBankInfo(ctx, accountID, last4Ptr, bankNamePtr)
				break
			}
		}
	}

	return nil
}

// ─── Private helpers ─────────────────────────────────────────────────────────

func (s *PayoutService) createConnectedAccount(ctx context.Context, email string) (*stripe.Account, error) {
	params := &stripe.AccountParams{
		Type:  stripe.String(string(stripe.AccountTypeExpress)),
		Email: stripe.String(email),
		Capabilities: &stripe.AccountCapabilitiesParams{
			Transfers: &stripe.AccountCapabilitiesTransfersParams{
				Requested: stripe.Bool(true),
			},
		},
		BusinessType: stripe.String(string(stripe.AccountBusinessTypeIndividual)),
		Metadata: map[string]string{
			"platform": "functionfly",
		},
	}

	return account.New(params)
}

func isStripeConnectPlatformNotReady(err error) bool {
	if err == nil {
		return false
	}
	var se *stripe.Error
	if errors.As(err, &se) && se != nil {
		m := strings.ToLower(se.Msg)
		if strings.Contains(m, "signed up for connect") {
			return true
		}
	}
	return strings.Contains(strings.ToLower(err.Error()), "signed up for connect")
}

func (s *PayoutService) createAccountLink(stripeAccountID string) (string, time.Time, error) {
	returnURL := os.Getenv("APP_URL")
	if returnURL == "" {
		returnURL = "https://functionfly.com"
	}

	params := &stripe.AccountLinkParams{
		Account:    stripe.String(stripeAccountID),
		RefreshURL: stripe.String(returnURL + "/settings/payouts?refresh=true"),
		ReturnURL:  stripe.String(returnURL + "/settings/payouts?connected=true"),
		Type:       stripe.String(string(stripe.AccountLinkTypeAccountOnboarding)),
	}

	link, err := accountlink.New(params)
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt := time.Unix(link.ExpiresAt, 0)
	return link.URL, expiresAt, nil
}

// ProcessTransferReversed handles a transfer.reversed webhook event by reversing the payout debit.
func (s *PayoutService) ProcessTransferReversed(ctx context.Context, stripeTransferID string) error {
	// Find the payout request by transfer ID
	var payoutReqID, userIDStr string
	var amountCents int
	err := s.payoutRepo.FindPayoutRequestByTransferID(ctx, stripeTransferID, &payoutReqID, &userIDStr, &amountCents)
	if err != nil {
		return fmt.Errorf("failed to find payout request for transfer %s: %w", stripeTransferID, err)
	}

	userID, _ := uuid.Parse(userIDStr)
	payoutReqIDParsed, _ := uuid.Parse(payoutReqID)

	// Reverse the debit
	if err := s.payoutRepo.ReversePayoutDebit(ctx, userID, amountCents, payoutReqIDParsed,
		fmt.Sprintf("Stripe transfer %s reversed", stripeTransferID)); err != nil {
		return fmt.Errorf("failed to reverse debit: %w", err)
	}

	// Update payout request status
	if err := s.payoutRepo.MarkPayoutRequestFailed(ctx, payoutReqIDParsed,
		fmt.Sprintf("Transfer reversed: %s", stripeTransferID)); err != nil {
		return fmt.Errorf("failed to mark payout as failed: %w", err)
	}

	return nil
}

// ProcessPayoutPaid handles a payout.paid webhook event by recording the Stripe payout ID.
func (s *PayoutService) ProcessPayoutPaid(ctx context.Context, stripePayoutID, stripeAccountID string) error {
	// Find the associated transfer for this payout to link it to our payout request
	payoutParams := &stripe.PayoutParams{}
	payoutParams.SetStripeAccount(stripeAccountID)

	po, err := stripePayout.Get(stripePayoutID, payoutParams)
	if err != nil {
		return fmt.Errorf("failed to get stripe payout: %w", err)
	}

	// The payout to bank is confirmed. We log it but don't change the payout request status
	// since the transfer to the connected account was already completed.
	logrus.WithFields(logrus.Fields{
		"stripe_payout_id":  stripePayoutID,
		"stripe_account_id": stripeAccountID,
		"amount":            po.Amount,
		"status":            po.Status,
	}).Info("Stripe payout to bank confirmed")

	return nil
}
