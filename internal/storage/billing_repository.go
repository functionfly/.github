package storage

import (
	"database/sql"
	"strings"

	"github.com/google/uuid"
)

// BillingRepository handles billing-related database operations
type BillingRepository struct {
	db *PostgresDB
}

// NewBillingRepository creates a new billing repository
func NewBillingRepository(db *PostgresDB) *BillingRepository {
	return &BillingRepository{db: db}
}

// nullStringPtr converts a sql.NullString to *string
func nullStringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	s := ns.String
	return &s
}

// normalizeInvoiceCurrency normalizes currency code to 3-letter uppercase
func normalizeInvoiceCurrency(currency string) string {
	c := strings.ToUpper(strings.TrimSpace(currency))
	if c == "" {
		return "USD"
	}
	if len(c) > 3 {
		return c[:3]
	}
	return c
}

// StringPtr returns a pointer to the string value, or nil if empty
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// UUIDPtr returns a pointer to the UUID value
func UUIDPtr(id uuid.UUID) *uuid.UUID {
	return &id
}
