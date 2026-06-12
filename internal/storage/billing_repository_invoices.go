package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CreateInvoice creates a new invoice
func (r *BillingRepository) CreateInvoice(ctx context.Context, invoice *Invoice) (*Invoice, error) {
	invoice.ID = uuid.New()
	invoice.CreatedAt = time.Now()
	invoice.UpdatedAt = time.Now()
	if invoice.Status == "" {
		invoice.Status = "draft"
	}

	query := `
		INSERT INTO invoices (id, tenant_id, subscription_id, status, amount_due_cents, amount_paid_cents, currency, period_start, period_end, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, tenant_id, subscription_id, status, amount_due_cents, amount_paid_cents, currency,
			stripe_invoice_id, external_reference, invoice_pdf_url, hosted_invoice_url, period_start, period_end, due_date, paid_at, created_at, updated_at`

	var stripeInvID, extRef sql.NullString
	err := r.db.QueryRowContext(ctx, query, invoice.ID, invoice.TenantID, invoice.SubscriptionID, invoice.Status,
		invoice.AmountDueCents, invoice.AmountPaidCents, invoice.Currency, invoice.PeriodStart,
		invoice.PeriodEnd, invoice.DueDate, invoice.CreatedAt, invoice.UpdatedAt).Scan(
		&invoice.ID, &invoice.TenantID, &invoice.SubscriptionID, &invoice.Status,
		&invoice.AmountDueCents, &invoice.AmountPaidCents, &invoice.Currency,
		&stripeInvID, &extRef,
		&invoice.InvoicePdfURL, &invoice.HostedInvoiceURL, &invoice.PeriodStart,
		&invoice.PeriodEnd, &invoice.DueDate, &invoice.PaidAt, &invoice.CreatedAt, &invoice.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}
	invoice.StripeInvoiceID = nullStringPtr(stripeInvID)
	invoice.ExternalReference = nullStringPtr(extRef)

	return invoice, nil
}

// CreatePaidInvoiceForStripeCheckoutSession inserts a paid invoice row for a completed Checkout Session (one-time payment).
// Safe to call on webhook retries: duplicate checkoutSessionID is ignored.
func (r *BillingRepository) CreatePaidInvoiceForStripeCheckoutSession(ctx context.Context, tenantID uuid.UUID, amountCents int, currency, checkoutSessionID, receiptURL string) error {
	if checkoutSessionID == "" || amountCents <= 0 {
		return fmt.Errorf("invalid checkout session id or amount")
	}
	var exists int
	err := r.db.QueryRowContext(ctx,
		`SELECT 1 FROM invoices WHERE external_reference = $1 LIMIT 1`, checkoutSessionID).Scan(&exists)
	if err == nil {
		return nil
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("invoice idempotency check: %w", err)
	}

	curr := normalizeInvoiceCurrency(currency)
	now := time.Now().UTC()
	id := uuid.New()
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO invoices (
			id, tenant_id, subscription_id, status,
			amount_due_cents, amount_paid_cents, currency,
			invoice_pdf_url, hosted_invoice_url,
			paid_at, created_at, updated_at, external_reference
		) VALUES ($1, $2, NULL, 'paid', $3, $3, $4, $5, $6, $7, $7, $7, $8)`,
		id, tenantID, amountCents, curr, receiptURL, receiptURL, now, checkoutSessionID,
	)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return nil
		}
		return fmt.Errorf("insert paid invoice: %w", err)
	}
	return nil
}

// ListInvoicesByTenant lists invoices for a tenant
func (r *BillingRepository) ListInvoicesByTenant(tenantID uuid.UUID, limit, offset int) ([]*Invoice, error) {
	query := `
		SELECT id, tenant_id, subscription_id, status, amount_due_cents, amount_paid_cents, currency,
			   invoice_pdf_url, hosted_invoice_url, period_start, period_end, due_date, paid_at, created_at, updated_at
		FROM invoices
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*Invoice
	for rows.Next() {
		invoice := &Invoice{}
		err := rows.Scan(&invoice.ID, &invoice.TenantID, &invoice.SubscriptionID, &invoice.Status,
			&invoice.AmountDueCents, &invoice.AmountPaidCents, &invoice.Currency,
			&invoice.InvoicePdfURL, &invoice.HostedInvoiceURL, &invoice.PeriodStart,
			&invoice.PeriodEnd, &invoice.DueDate, &invoice.PaidAt, &invoice.CreatedAt, &invoice.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

// CountInvoicesByTenant returns the number of invoice rows for a tenant.
func (r *BillingRepository) CountInvoicesByTenant(tenantID uuid.UUID) (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM invoices WHERE tenant_id = $1`, tenantID).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("failed to count invoices: %w", err)
	}
	return n, nil
}

// ListAllInvoices lists all invoices across tenants (for admin dashboard)
func (r *BillingRepository) ListAllInvoices(limit, offset int) ([]*Invoice, error) {
	query := `
		SELECT id, tenant_id, subscription_id, status, amount_due_cents, amount_paid_cents, currency,
			   invoice_pdf_url, hosted_invoice_url, period_start, period_end, due_date, paid_at, created_at, updated_at
		FROM invoices
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*Invoice
	for rows.Next() {
		invoice := &Invoice{}
		err := rows.Scan(&invoice.ID, &invoice.TenantID, &invoice.SubscriptionID, &invoice.Status,
			&invoice.AmountDueCents, &invoice.AmountPaidCents, &invoice.Currency,
			&invoice.InvoicePdfURL, &invoice.HostedInvoiceURL, &invoice.PeriodStart,
			&invoice.PeriodEnd, &invoice.DueDate, &invoice.PaidAt, &invoice.CreatedAt, &invoice.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

// GetInvoiceByID retrieves an invoice by ID
func (r *BillingRepository) GetInvoiceByID(id uuid.UUID) (*Invoice, error) {
	query := `
		SELECT id, tenant_id, subscription_id, status, amount_due_cents, amount_paid_cents, currency,
			   invoice_pdf_url, hosted_invoice_url, period_start, period_end, due_date, paid_at, created_at, updated_at
		FROM invoices WHERE id = $1`

	invoice := &Invoice{}
	err := r.db.QueryRow(query, id).Scan(&invoice.ID, &invoice.TenantID, &invoice.SubscriptionID, &invoice.Status,
		&invoice.AmountDueCents, &invoice.AmountPaidCents, &invoice.Currency,
		&invoice.InvoicePdfURL, &invoice.HostedInvoiceURL, &invoice.PeriodStart,
		&invoice.PeriodEnd, &invoice.DueDate, &invoice.PaidAt, &invoice.CreatedAt, &invoice.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return invoice, nil
}

// GetInvoiceByPeriod retrieves an invoice by tenant and period
func (r *BillingRepository) GetInvoiceByPeriod(ctx context.Context, tenantID uuid.UUID, periodStart, periodEnd time.Time) (*Invoice, error) {
	query := `
		SELECT id, tenant_id, subscription_id, status, amount_due_cents, amount_paid_cents, currency,
			invoice_pdf_url, hosted_invoice_url,
			period_start, period_end, due_date, paid_at, created_at, updated_at
		FROM invoices
		WHERE tenant_id = $1
			AND status = 'draft'
			AND period_start = $2
			AND period_end = $3
		LIMIT 1
	`

	invoice := &Invoice{}
	err := r.db.QueryRowContext(ctx, query, tenantID, periodStart, periodEnd).Scan(
		&invoice.ID, &invoice.TenantID, &invoice.SubscriptionID, &invoice.Status,
		&invoice.AmountDueCents, &invoice.AmountPaidCents, &invoice.Currency,
		&invoice.InvoicePdfURL, &invoice.HostedInvoiceURL, &invoice.PeriodStart,
		&invoice.PeriodEnd, &invoice.DueDate, &invoice.PaidAt, &invoice.CreatedAt, &invoice.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice by period: %w", err)
	}

	return invoice, nil
}

// UpdateInvoice updates invoice fields dynamically
func (r *BillingRepository) UpdateInvoice(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*Invoice, error) {
	// Get current invoice
	current, err := r.GetInvoiceByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get current invoice: %w", err)
	}
	if current == nil {
		return nil, fmt.Errorf("invoice not found")
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if status, ok := updates["status"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if amountPaidCents, ok := updates["amount_paid_cents"].(int); ok {
		setParts = append(setParts, fmt.Sprintf("amount_paid_cents = $%d", argIndex))
		args = append(args, amountPaidCents)
		argIndex++
		if amountPaidCents == current.AmountDueCents {
			setParts = append(setParts, "paid_at = NOW()")
		}
	}

	if len(setParts) == 0 {
		return current, nil
	}

	setParts = append(setParts, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE invoices SET %s WHERE id = $%d RETURNING id, tenant_id, subscription_id, status, amount_due_cents, amount_paid_cents, currency, invoice_pdf_url, hosted_invoice_url, period_start, period_end, due_date, paid_at, created_at, updated_at",
		strings.Join(setParts, ", "), argIndex)

	args = append(args, id)

	updated := &Invoice{}
	err = r.db.QueryRow(query, args...).Scan(&updated.ID, &updated.TenantID, &updated.SubscriptionID, &updated.Status,
		&updated.AmountDueCents, &updated.AmountPaidCents, &updated.Currency,
		&updated.InvoicePdfURL, &updated.HostedInvoiceURL, &updated.PeriodStart,
		&updated.PeriodEnd, &updated.DueDate, &updated.PaidAt, &updated.CreatedAt, &updated.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	return updated, nil
}
