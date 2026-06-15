package storage

import (

	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/functionfly/functionfly/internal/types"
)

// CreditNoteRepository handles credit note database operations (SQL-based)
type CreditNoteRepository struct {
	db *PostgresDB
}

// NewCreditNoteRepository creates a new credit note repository
func NewCreditNoteRepositorySQL(db *PostgresDB) *CreditNoteRepository {
	return &CreditNoteRepository{db: db}
}

// GenerateReferenceNumber generates a unique reference number for a credit note
func (r *CreditNoteRepository) GenerateReferenceNumber(ctx context.Context) (string, error) {
	year := time.Now().Year()
	prefix := fmt.Sprintf("CN-%d-", year)

	var maxSeq int
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(CAST(SUBSTRING(reference_number FROM LENGTH($1)+1 FOR position('-' IN reference_number) - LENGTH($1) - 1) AS INTEGER)), 0)
		FROM credit_notes
		WHERE reference_number LIKE $2
	`, prefix, prefix+"%").Scan(&maxSeq)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	return fmt.Sprintf("%s%06d", prefix, maxSeq+1), nil
}

// Create creates a new credit note
func (r *CreditNoteRepository) Create(ctx context.Context, creditNote *CreditNote) (*CreditNote, error) {
	if creditNote.ID == uuid.Nil {
		creditNote.ID = uuid.New()
	}
	if creditNote.CreatedAt.IsZero() {
		creditNote.CreatedAt = time.Now().UTC()
	}
	creditNote.UpdatedAt = time.Now().UTC()

	if creditNote.Status == "" {
		creditNote.Status = CreditNoteStatusDraft
	}

	refNumber, err := r.GenerateReferenceNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate reference number: %w", err)
	}
	creditNote.ReferenceNumber = refNumber

	query := `
		INSERT INTO credit_notes (id, tenant_id, invoice_id, reference_number, status,
			subtotal_cents, tax_cents, total_cents, currency, reason, description,
			payment_refund_id, issued_at, applied_at, voided_at, issued_by, notes,
			created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING id, tenant_id, invoice_id, reference_number, status,
			subtotal_cents, tax_cents, total_cents, currency, reason, description,
			payment_refund_id, issued_at, applied_at, voided_at, issued_by, notes,
			created_at, updated_at
	`

	err = r.db.QueryRowContext(ctx, query,
		creditNote.ID, creditNote.TenantID, creditNote.InvoiceID, creditNote.ReferenceNumber, creditNote.Status,
		creditNote.SubtotalCents, creditNote.TaxCents, creditNote.TotalCents, creditNote.Currency,
		creditNote.Reason, creditNote.Description, creditNote.PaymentRefundID,
		creditNote.IssuedAt, creditNote.AppliedAt, creditNote.VoidedAt,
		creditNote.IssuedBy, creditNote.Notes, creditNote.CreatedAt, creditNote.UpdatedAt,
	).Scan(
		&creditNote.ID, &creditNote.TenantID, &creditNote.InvoiceID, &creditNote.ReferenceNumber, &creditNote.Status,
		&creditNote.SubtotalCents, &creditNote.TaxCents, &creditNote.TotalCents, &creditNote.Currency,
		&creditNote.Reason, &creditNote.Description, &creditNote.PaymentRefundID,
		&creditNote.IssuedAt, &creditNote.AppliedAt, &creditNote.VoidedAt,
		&creditNote.IssuedBy, &creditNote.Notes, &creditNote.CreatedAt, &creditNote.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create credit note: %w", err)
	}

	return creditNote, nil
}

// GetByID retrieves a credit note by ID
func (r *CreditNoteRepository) GetByID(ctx context.Context, id uuid.UUID) (*CreditNote, error) {
	query := `
		SELECT id, tenant_id, invoice_id, reference_number, status,
			subtotal_cents, tax_cents, total_cents, currency, reason, description,
			payment_refund_id, issued_at, applied_at, voided_at, issued_by, notes,
			created_at, updated_at
		FROM credit_notes WHERE id = $1
	`

	creditNote := &CreditNote{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&creditNote.ID, &creditNote.TenantID, &creditNote.InvoiceID, &creditNote.ReferenceNumber, &creditNote.Status,
		&creditNote.SubtotalCents, &creditNote.TaxCents, &creditNote.TotalCents, &creditNote.Currency,
		&creditNote.Reason, &creditNote.Description, &creditNote.PaymentRefundID,
		&creditNote.IssuedAt, &creditNote.AppliedAt, &creditNote.VoidedAt,
		&creditNote.IssuedBy, &creditNote.Notes, &creditNote.CreatedAt, &creditNote.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get credit note: %w", err)
	}

	return creditNote, nil
}

// GetByReferenceNumber retrieves a credit note by reference number
func (r *CreditNoteRepository) GetByReferenceNumber(ctx context.Context, refNumber string) (*CreditNote, error) {
	query := `
		SELECT id, tenant_id, invoice_id, reference_number, status,
			subtotal_cents, tax_cents, total_cents, currency, reason, description,
			payment_refund_id, issued_at, applied_at, voided_at, issued_by, notes,
			created_at, updated_at
		FROM credit_notes WHERE reference_number = $1
	`

	creditNote := &CreditNote{}
	err := r.db.QueryRowContext(ctx, query, refNumber).Scan(
		&creditNote.ID, &creditNote.TenantID, &creditNote.InvoiceID, &creditNote.ReferenceNumber, &creditNote.Status,
		&creditNote.SubtotalCents, &creditNote.TaxCents, &creditNote.TotalCents, &creditNote.Currency,
		&creditNote.Reason, &creditNote.Description, &creditNote.PaymentRefundID,
		&creditNote.IssuedAt, &creditNote.AppliedAt, &creditNote.VoidedAt,
		&creditNote.IssuedBy, &creditNote.Notes, &creditNote.CreatedAt, &creditNote.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get credit note by reference: %w", err)
	}

	return creditNote, nil
}

// List retrieves credit notes with optional filtering
func (r *CreditNoteRepository) List(ctx context.Context, filter *CreditNoteFilter, limit, offset int) ([]*CreditNote, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}

	whereClause := "1=1"
	args := []interface{}{}
	argIndex := 1

	if filter != nil {
		if filter.TenantID != nil {
			whereClause += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
			args = append(args, *filter.TenantID)
			argIndex++
		}
		if filter.InvoiceID != nil {
			whereClause += fmt.Sprintf(" AND invoice_id = $%d", argIndex)
			args = append(args, *filter.InvoiceID)
			argIndex++
		}
		if filter.Status != nil && *filter.Status != "" {
			whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
			args = append(args, filter.Status)
			argIndex++
		}
		if filter.StartDate != nil {
			whereClause += fmt.Sprintf(" AND created_at >= $%d", argIndex)
			args = append(args, *filter.StartDate)
			argIndex++
		}
		if filter.EndDate != nil {
			whereClause += fmt.Sprintf(" AND created_at <= $%d", argIndex)
			args = append(args, *filter.EndDate)
			argIndex++
		}
		if filter.PaymentRefundID != nil {
			whereClause += fmt.Sprintf(" AND payment_refund_id = $%d", argIndex)
			args = append(args, *filter.PaymentRefundID)
			argIndex++
		}
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM credit_notes WHERE %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count credit notes: %w", err)
	}

	listQuery := fmt.Sprintf(`
		SELECT id, tenant_id, invoice_id, reference_number, status,
			subtotal_cents, tax_cents, total_cents, currency, reason, description,
			payment_refund_id, issued_at, applied_at, voided_at, issued_by, notes,
			created_at, updated_at
		FROM credit_notes
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list credit notes: %w", err)
	}
	defer rows.Close()

	var creditNotes []*CreditNote
	for rows.Next() {
		cn := &CreditNote{}
		err := rows.Scan(
			&cn.ID, &cn.TenantID, &cn.InvoiceID, &cn.ReferenceNumber, &cn.Status,
			&cn.SubtotalCents, &cn.TaxCents, &cn.TotalCents, &cn.Currency,
			&cn.Reason, &cn.Description, &cn.PaymentRefundID,
			&cn.IssuedAt, &cn.AppliedAt, &cn.VoidedAt,
			&cn.IssuedBy, &cn.Notes, &cn.CreatedAt, &cn.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan credit note: %w", err)
		}
		creditNotes = append(creditNotes, cn)
	}

	return creditNotes, total, nil
}

// ListByTenant retrieves all credit notes for a tenant
func (r *CreditNoteRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*CreditNote, int64, error) {
	return r.List(ctx, &CreditNoteFilter{TenantID: &tenantID}, limit, offset)
}

// ListByInvoice retrieves all credit notes for an invoice
func (r *CreditNoteRepository) ListByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]*CreditNote, error) {
	query := `
		SELECT id, tenant_id, invoice_id, reference_number, status,
			subtotal_cents, tax_cents, total_cents, currency, reason, description,
			payment_refund_id, issued_at, applied_at, voided_at, issued_by, notes,
			created_at, updated_at
		FROM credit_notes
		WHERE invoice_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list credit notes by invoice: %w", err)
	}
	defer rows.Close()

	var creditNotes []*CreditNote
	for rows.Next() {
		cn := &CreditNote{}
		err := rows.Scan(
			&cn.ID, &cn.TenantID, &cn.InvoiceID, &cn.ReferenceNumber, &cn.Status,
			&cn.SubtotalCents, &cn.TaxCents, &cn.TotalCents, &cn.Currency,
			&cn.Reason, &cn.Description, &cn.PaymentRefundID,
			&cn.IssuedAt, &cn.AppliedAt, &cn.VoidedAt,
			&cn.IssuedBy, &cn.Notes, &cn.CreatedAt, &cn.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan credit note: %w", err)
		}
		creditNotes = append(creditNotes, cn)
	}

	return creditNotes, nil
}

// Update updates a credit note
func (r *CreditNoteRepository) Update(ctx context.Context, creditNote *CreditNote) error {
	creditNote.UpdatedAt = time.Now().UTC()

	query := `
		UPDATE credit_notes SET
			tenant_id = $1, invoice_id = $2, status = $3,
			subtotal_cents = $4, tax_cents = $5, total_cents = $6, currency = $7,
			reason = $8, description = $9, payment_refund_id = $10,
			issued_at = $11, applied_at = $12, voided_at = $13,
			issued_by = $14, notes = $15, updated_at = $16
		WHERE id = $17
	`

	_, err := r.db.ExecContext(ctx, query,
		creditNote.TenantID, creditNote.InvoiceID, creditNote.Status,
		creditNote.SubtotalCents, creditNote.TaxCents, creditNote.TotalCents, creditNote.Currency,
		creditNote.Reason, creditNote.Description, creditNote.PaymentRefundID,
		creditNote.IssuedAt, creditNote.AppliedAt, creditNote.VoidedAt,
		creditNote.IssuedBy, creditNote.Notes, creditNote.UpdatedAt, creditNote.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update credit note: %w", err)
	}

	return nil
}

// UpdateStatus updates the status of a credit note
func (r *CreditNoteRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now().UTC(),
	}

	now := time.Now().UTC()
	switch status {
	case CreditNoteStatusIssued:
		updates["issued_at"] = now
	case CreditNoteStatusApplied:
		updates["applied_at"] = now
	case CreditNoteStatusVoid:
		updates["voided_at"] = now
	}

	query := fmt.Sprintf("UPDATE credit_notes SET status = $1, updated_at = $2")
	args := []interface{}{status, time.Now().UTC()}
	argIndex := 3

	if updates["issued_at"] != nil {
		query += fmt.Sprintf(", issued_at = $%d", argIndex)
		args = append(args, now)
		argIndex++
	}
	if updates["applied_at"] != nil {
		query += fmt.Sprintf(", applied_at = $%d", argIndex)
		args = append(args, now)
		argIndex++
	}
	if updates["voided_at"] != nil {
		query += fmt.Sprintf(", voided_at = $%d", argIndex)
		args = append(args, now)
		argIndex++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update credit note status: %w", err)
	}

	return nil
}

// Void voids a credit note
func (r *CreditNoteRepository) Void(ctx context.Context, id uuid.UUID) error {
	return r.UpdateStatus(ctx, id, CreditNoteStatusVoid)
}

// Apply marks a credit note as applied
func (r *CreditNoteRepository) Apply(ctx context.Context, id uuid.UUID) error {
	return r.UpdateStatus(ctx, id, CreditNoteStatusApplied)
}

// CreateLineItem creates a line item for a credit note
func (r *CreditNoteRepository) CreateLineItem(ctx context.Context, item *CreditNoteLineItem) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	item.CreatedAt = time.Now().UTC()
	item.UpdatedAt = time.Now().UTC()

	query := `
		INSERT INTO credit_note_line_items (id, credit_note_id, description, quantity,
			unit_price_cents, tax_cents, amount_cents, total_cents, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.CreditNoteID, item.Description, item.Quantity,
		item.UnitPriceCents, item.TaxCents, item.AmountCents, item.TotalCents,
		item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create credit note line item: %w", err)
	}

	return nil
}

// GetLineItems retrieves all line items for a credit note
func (r *CreditNoteRepository) GetLineItems(ctx context.Context, creditNoteID uuid.UUID) ([]*CreditNoteLineItem, error) {
	query := `
		SELECT id, credit_note_id, description, quantity, unit_price_cents,
			tax_cents, amount_cents, total_cents, created_at, updated_at
		FROM credit_note_line_items
		WHERE credit_note_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, creditNoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credit note line items: %w", err)
	}
	defer rows.Close()

	var items []*CreditNoteLineItem
	for rows.Next() {
		item := &CreditNoteLineItem{}
		err := rows.Scan(
			&item.ID, &item.CreditNoteID, &item.Description, &item.Quantity,
			&item.UnitPriceCents, &item.TaxCents, &item.AmountCents, &item.TotalCents,
			&item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan credit note line item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// GetWithRelations retrieves a credit note with line items
func (r *CreditNoteRepository) GetWithRelations(ctx context.Context, id uuid.UUID) (*CreditNote, error) {
	creditNote, err := r.GetByID(ctx, id)
	if err != nil || creditNote == nil {
		return creditNote, err
	}

	lineItems, err := r.GetLineItems(ctx, id)
	if err != nil {
		return nil, err
	}
	creditNote.LineItems = lineItems

	return creditNote, nil
}

// DeleteLineItem deletes a line item
func (r *CreditNoteRepository) DeleteLineItem(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM credit_note_line_items WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete credit note line item: %w", err)
	}
	return nil
}

// DeleteLineItems deletes all line items for a credit note
func (r *CreditNoteRepository) DeleteLineItems(ctx context.Context, creditNoteID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM credit_note_line_items WHERE credit_note_id = $1", creditNoteID)
	if err != nil {
		return fmt.Errorf("failed to delete credit note line items: %w", err)
	}
	return nil
}

// GetCreditNoteStats returns aggregate statistics about credit notes
func (r *CreditNoteRepository) GetCreditNoteStats(ctx context.Context, tenantID *uuid.UUID) (*CreditNoteStats, error) {
	stats := &CreditNoteStats{}

	whereClause := "1=1"
	args := []interface{}{}
	if tenantID != nil {
		whereClause += " AND tenant_id = $1"
		args = append(args, *tenantID)
	}

	var totalNotes int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM credit_notes WHERE %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalNotes); err != nil {
		return nil, fmt.Errorf("failed to count credit notes: %w", err)
	}
	stats.TotalCreditNotes = totalNotes

	statusQuery := fmt.Sprintf(`
		SELECT status, COUNT(*) as count
		FROM credit_notes
		WHERE %s
		GROUP BY status
	`, whereClause)

	rows, err := r.db.QueryContext(ctx, statusQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get credit note stats by status: %w", err)
	}
	defer rows.Close()

	stats.ByStatus = make(map[string]int64)
	for rows.Next() {
		var status string
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats.ByStatus[status] = count
	}

	totalQuery := fmt.Sprintf("SELECT COALESCE(SUM(total_cents), 0) FROM credit_notes WHERE %s", whereClause)
	var totalCents int64
	if err := r.db.QueryRowContext(ctx, totalQuery, args...).Scan(&totalCents); err != nil {
		return nil, fmt.Errorf("failed to get total credited: %w", err)
	}
	stats.TotalCreditedCents = totalCents

	return stats, nil
}
// Status constant aliases from the types package.
const CreditNoteStatusIssued = types.CreditNoteStatusIssued
const CreditNoteStatusApplied = types.CreditNoteStatusApplied
const CreditNoteStatusDraft = types.CreditNoteStatusDraft
const CreditNoteStatusVoid = types.CreditNoteStatusVoid
