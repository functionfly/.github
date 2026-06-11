package wallet

import (
	"context"
	"fmt"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"testing"
)

func TestWalletPaymentFlow(t *testing.T) {
	db, err := storage.NewPostgresDB()
	if err != nil {
		t.Fatalf("DB connect error: %v", err)
	}

	walletRepo := NewRepository(db.GORM)
	walletSvc := NewService(walletRepo, nil)

	userIDStr := "acc3d00f-163f-4804-9047-69064feb1de6"
	userID := uuid.MustParse(userIDStr)

	fmt.Println("=== Test 1: GetWalletByOwner ===")
	w, err := walletSvc.GetWalletByOwner(context.Background(), OwnerTypeUser, userIDStr)
	if err != nil {
		t.Errorf("Error: %v", err)
	} else if w == nil {
		t.Error("Wallet not found")
	} else {
		fmt.Printf("Wallet: ID=%s, Balance=%.4f\n", w.ID, w.BalanceUSD)
	}

	fmt.Println("\n=== Test 2: GetUserBalance ===")
	balance, err := walletSvc.GetUserBalance(context.Background(), userID)
	if err != nil {
		t.Errorf("Error: %v", err)
	} else {
		fmt.Printf("Balance: %.4f\n", balance)
	}

	fmt.Println("\n=== Test 3: ConsumeUserOrAgentCredits ===")
	update, err := walletSvc.ConsumeUserOrAgentCredits(context.Background(), userIDStr, 0.015)
	if err != nil {
		t.Errorf("Error: %v", err)
	} else {
		fmt.Printf("Consumed 0.015: Previous=%.4f, Current=%.4f\n", update.PreviousBalance, update.CurrentBalance)
	}

	fmt.Println("\n=== Test 4: GetWalletByOwner (after consume) ===")
	w, err = walletSvc.GetWalletByOwner(context.Background(), OwnerTypeUser, userIDStr)
	if err != nil {
		t.Errorf("Error: %v", err)
	} else if w == nil {
		t.Error("Wallet not found")
	} else {
		fmt.Printf("Wallet: ID=%s, Balance=%.4f\n", w.ID, w.BalanceUSD)
	}
}
