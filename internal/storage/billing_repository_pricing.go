package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CreatePricingTier creates a new pricing tier
func (r *BillingRepository) CreatePricingTier(ctx context.Context, tier *PricingTier) (*PricingTier, error) {
	tier.ID = uuid.New()
	tier.CreatedAt = time.Now()
	tier.UpdatedAt = time.Now()

	query := `
		INSERT INTO pricing_tiers (id, name, description, price_cents, currency, features, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, name, description, price_cents, currency, features, is_active, created_at, updated_at`

	var features []byte
	if tier.Features != nil {
		features, _ = json.Marshal(tier.Features)
	}

	err := r.db.QueryRow(query, tier.ID, tier.Name, tier.Description, tier.PriceCents,
		tier.Currency, features, tier.IsActive, tier.CreatedAt, tier.UpdatedAt).Scan(
		&tier.ID, &tier.Name, &tier.Description, &tier.PriceCents,
		&tier.Currency, &features, &tier.IsActive, &tier.CreatedAt, &tier.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create pricing tier: %w", err)
	}

	if len(features) > 0 {
		json.Unmarshal(features, &tier.Features)
	}

	return tier, nil
}

// ListPricingTiers lists all active pricing tiers
func (r *BillingRepository) ListPricingTiers() ([]*PricingTier, error) {
	query := `SELECT id, name, description, price_cents, currency, features, is_active, created_at, updated_at
			  FROM pricing_tiers WHERE is_active = true ORDER BY price_cents ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list pricing tiers: %w", err)
	}
	defer rows.Close()

	var tiers []*PricingTier
	for rows.Next() {
		tier := &PricingTier{}
		var features []byte
		err := rows.Scan(&tier.ID, &tier.Name, &tier.Description, &tier.PriceCents,
			&tier.Currency, &features, &tier.IsActive, &tier.CreatedAt, &tier.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pricing tier: %w", err)
		}

		if len(features) > 0 {
			json.Unmarshal(features, &tier.Features)
		}

		tiers = append(tiers, tier)
	}

	return tiers, nil
}

// GetPricingTierByID retrieves a pricing tier by ID
func (r *BillingRepository) GetPricingTierByID(id uuid.UUID) (*PricingTier, error) {
	query := `SELECT id, name, description, price_cents, currency, features, is_active, created_at, updated_at
			  FROM pricing_tiers WHERE id = $1`

	tier := &PricingTier{}
	var features []byte
	err := r.db.QueryRow(query, id).Scan(&tier.ID, &tier.Name, &tier.Description, &tier.PriceCents,
		&tier.Currency, &features, &tier.IsActive, &tier.CreatedAt, &tier.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing tier: %w", err)
	}

	if len(features) > 0 {
		json.Unmarshal(features, &tier.Features)
	}

	return tier, nil
}

// UpdatePricingTier updates pricing tier fields dynamically
func (r *BillingRepository) UpdatePricingTier(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*PricingTier, error) {
	// Get current tier
	current, err := r.GetPricingTierByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get current pricing tier: %w", err)
	}
	if current == nil {
		return nil, fmt.Errorf("pricing tier not found")
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if name, ok := updates["name"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, name)
		argIndex++
	}

	if description, ok := updates["description"].(string); ok {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, description)
		argIndex++
	}

	if priceCents, ok := updates["price_cents"].(int); ok {
		setParts = append(setParts, fmt.Sprintf("price_cents = $%d", argIndex))
		args = append(args, priceCents)
		argIndex++
	}

	if isActive, ok := updates["is_active"].(bool); ok {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, isActive)
		argIndex++
	}

	if len(setParts) == 0 {
		return current, nil
	}

	setParts = append(setParts, "updated_at = NOW()")

	query := fmt.Sprintf("UPDATE pricing_tiers SET %s WHERE id = $%d RETURNING id, name, description, price_cents, currency, features, is_active, created_at, updated_at",
		strings.Join(setParts, ", "), argIndex)

	args = append(args, id)

	updated := &PricingTier{}
	var features []byte
	err = r.db.QueryRow(query, args...).Scan(&updated.ID, &updated.Name, &updated.Description,
		&updated.PriceCents, &updated.Currency, &features, &updated.IsActive, &updated.CreatedAt, &updated.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update pricing tier: %w", err)
	}

	if len(features) > 0 {
		json.Unmarshal(features, &updated.Features)
	}

	return updated, nil
}

// DeletePricingTier soft deletes a pricing tier
func (r *BillingRepository) DeletePricingTier(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec("UPDATE pricing_tiers SET is_active = false WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete pricing tier: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("pricing tier not found")
	}

	return nil
}
