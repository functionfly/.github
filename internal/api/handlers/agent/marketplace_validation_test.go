package agent

import "testing"

func TestValidateMarketplaceFunctionRef(t *testing.T) {
	if err := validateMarketplaceFunctionRef("acme", "hello-world"); err != nil {
		t.Fatalf("expected valid ref: %v", err)
	}
	if err := validateMarketplaceFunctionRef("", "x"); err == nil {
		t.Fatal("expected error for empty author")
	}
	if err := validateMarketplaceFunctionRef("bad author", "fn"); err == nil {
		t.Fatal("expected error for invalid author charset")
	}
}
