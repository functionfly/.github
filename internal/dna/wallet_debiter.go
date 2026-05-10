package dna

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// WalletServiceDebiter implements WalletDebiter using the wallet service.
type WalletServiceDebiter struct {
	debitFn func(ctx context.Context, userID uuid.UUID, amountUSD float64, description string) error
}

// NewWalletServiceDebiter creates a new debiter wrapping the wallet service's DebitWalletUser.
func NewWalletServiceDebiter(debitFn func(ctx context.Context, userID uuid.UUID, amountUSD float64, description string) error) *WalletServiceDebiter {
	return &WalletServiceDebiter{debitFn: debitFn}
}

// DebitForDNAMutation debits credits for accepting a DNA mutation.
func (d *WalletServiceDebiter) DebitForDNAMutation(ctx context.Context, userID string, amountUSD float64, functionID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	return d.debitFn(ctx, uid, amountUSD, fmt.Sprintf("DNA mutation accepted for function %s", functionID))
}
