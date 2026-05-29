package studio

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

// BuyerLicenseItem is a license grant purchased by the caller (no secret key).
type BuyerLicenseItem struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	FunctionID      string `json:"functionId"`
	FunctionName    string `json:"functionName"`
	IssuerTenantID  string `json:"issuerTenantId,omitempty"`
	PurchaserName   string `json:"purchaserName"`
	IssuedAt        int64  `json:"issuedAt"`
	ExpiresAt       *int64 `json:"expiresAt,omitempty"`
	MaxActivations  *int   `json:"maxActivations,omitempty"`
	ActivationCount int    `json:"activationCount"`
	Revoked         bool   `json:"revoked"`
	KeyPrefix       string `json:"keyPrefix,omitempty"`
}

// BuyerSubscriptionItem is a plan subscription purchased by the caller.
type BuyerSubscriptionItem struct {
	ID                 string  `json:"id"`
	PlanName           string  `json:"planName"`
	CreatorTenantID    string  `json:"creatorTenantId,omitempty"`
	Status             string  `json:"status"`
	Amount             float64 `json:"amount"`
	Currency           string  `json:"currency"`
	BillingCycle       string  `json:"billingCycle"`
	CurrentPeriodStart int64   `json:"currentPeriodStart"`
	CurrentPeriodEnd   int64   `json:"currentPeriodEnd"`
	CancelAtPeriodEnd  bool    `json:"cancelAtPeriodEnd"`
}

// AgentHiringItem is an agent the tenant hired from the marketplace.
type AgentHiringItem struct {
	ID        string         `json:"id"`
	AgentID   string         `json:"agentId"`
	TaskType  string         `json:"taskType"`
	BudgetUSD float64        `json:"budgetUsd"`
	Status    string         `json:"status"`
	CreatedAt int64          `json:"createdAt"`
	Payload   map[string]any `json:"taskPayload,omitempty"`
}

// FunctionPurchaseItem is a function bought via agent wallet.
type FunctionPurchaseItem struct {
	ID             string  `json:"id"`
	AgentID        string  `json:"agentId"`
	FunctionAuthor string  `json:"functionAuthor"`
	FunctionName   string  `json:"functionName"`
	PricePaidUSD   float64 `json:"pricePaidUsd"`
	Status         string  `json:"status"`
	CreatedAt      int64   `json:"createdAt"`
}

// MyPurchasesSummary aggregates buyer-side marketplace assets.
type MyPurchasesSummary struct {
	Functions          []FunctionPurchaseItem  `json:"functions"`
	Agents             []AgentHiringItem       `json:"agents"`
	Licenses           []BuyerLicenseItem      `json:"licenses"`
	Subscriptions      []BuyerSubscriptionItem `json:"subscriptions"`
	TotalFunctions     int                     `json:"totalFunctions"`
	TotalAgents        int                     `json:"totalAgents"`
	TotalLicenses      int                     `json:"totalLicenses"`
	TotalSubscriptions int                     `json:"totalSubscriptions"`
}

// ListMyPurchases returns buyer-scoped marketplace purchases for a tenant/user.
func (r *MarketplaceRepository) ListMyPurchases(ctx context.Context, tenantID, userID string, limit, offset int) (*MyPurchasesSummary, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	out := &MyPurchasesSummary{
		Functions:     []FunctionPurchaseItem{},
		Agents:        []AgentHiringItem{},
		Licenses:      []BuyerLicenseItem{},
		Subscriptions: []BuyerSubscriptionItem{},
	}

	functions, err := r.listBuyerFunctionPurchases(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	out.Functions = functions
	out.TotalFunctions = len(functions)

	agents, err := r.listBuyerAgentHirings(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	out.Agents = agents
	out.TotalAgents = len(agents)

	licenses, err := r.listBuyerLicenses(ctx, tenantID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	out.Licenses = licenses
	out.TotalLicenses = len(licenses)

	subs, err := r.listBuyerSubscriptions(ctx, tenantID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	out.Subscriptions = subs
	out.TotalSubscriptions = len(subs)

	return out, nil
}

func (r *MarketplaceRepository) listBuyerFunctionPurchases(ctx context.Context, tenantID string, limit, offset int) ([]FunctionPurchaseItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT fp.id::text, fp.agent_id, fp.function_author, fp.function_name,
			fp.price_paid_usd, fp.status,
			EXTRACT(EPOCH FROM fp.created_at)::bigint * 1000
		FROM function_purchases fp
		INNER JOIN agent_identities ai ON ai.agent_id = fp.agent_id
		WHERE ai.tenant_id = $1::uuid AND fp.status = 'completed'
		ORDER BY fp.created_at DESC
		LIMIT $2 OFFSET $3`,
		tenantID, limit, offset,
	)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return []FunctionPurchaseItem{}, nil
		}
		return nil, fmt.Errorf("list buyer function purchases: %w", err)
	}
	defer rows.Close()

	items := make([]FunctionPurchaseItem, 0)
	for rows.Next() {
		var item FunctionPurchaseItem
		if err := rows.Scan(
			&item.ID, &item.AgentID, &item.FunctionAuthor, &item.FunctionName,
			&item.PricePaidUSD, &item.Status, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MarketplaceRepository) listBuyerAgentHirings(ctx context.Context, tenantID string, limit, offset int) ([]AgentHiringItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, agent_id, task_type, budget_usd, status,
			EXTRACT(EPOCH FROM created_at)::bigint * 1000,
			COALESCE(task_payload, '{}'::jsonb)
		FROM agent_hirings
		WHERE tenant_id = $1::uuid
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		tenantID, limit, offset,
	)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return []AgentHiringItem{}, nil
		}
		return nil, fmt.Errorf("list buyer agent hirings: %w", err)
	}
	defer rows.Close()

	items := make([]AgentHiringItem, 0)
	for rows.Next() {
		var item AgentHiringItem
		var payload []byte
		if err := rows.Scan(
			&item.ID, &item.AgentID, &item.TaskType, &item.BudgetUSD, &item.Status,
			&item.CreatedAt, &payload,
		); err != nil {
			return nil, err
		}
		if len(payload) > 0 {
			_ = json.Unmarshal(payload, &item.Payload)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MarketplaceRepository) listBuyerLicenses(ctx context.Context, tenantID, userID string, limit, offset int) ([]BuyerLicenseItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, license_type, function_id, function_name,
			tenant_id::text,
			purchaser_name,
			EXTRACT(EPOCH FROM created_at)::bigint * 1000,
			EXTRACT(EPOCH FROM expires_at)::bigint * 1000,
			max_activations, activation_count,
			(revoked_at IS NOT NULL) AS revoked,
			license_key_prefix
		FROM marketplace_license_grants
		WHERE (purchaser_tenant_id = $1::uuid OR purchaser_user_id = NULLIF($2, '')::uuid)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`,
		tenantID, userID, limit, offset,
	)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return []BuyerLicenseItem{}, nil
		}
		return nil, fmt.Errorf("list buyer licenses: %w", err)
	}
	defer rows.Close()

	items := make([]BuyerLicenseItem, 0)
	for rows.Next() {
		var item BuyerLicenseItem
		var expiresMs sql.NullInt64
		var maxAct sql.NullInt64
		if err := rows.Scan(
			&item.ID, &item.Type, &item.FunctionID, &item.FunctionName,
			&item.IssuerTenantID, &item.PurchaserName, &item.IssuedAt, &expiresMs,
			&maxAct, &item.ActivationCount, &item.Revoked, &item.KeyPrefix,
		); err != nil {
			return nil, err
		}
		if expiresMs.Valid {
			v := expiresMs.Int64
			item.ExpiresAt = &v
		}
		if maxAct.Valid {
			v := int(maxAct.Int64)
			item.MaxActivations = &v
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MarketplaceRepository) listBuyerSubscriptions(ctx context.Context, tenantID, userID string, limit, offset int) ([]BuyerSubscriptionItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, plan_name, creator_tenant_id::text, status, amount, currency,
			billing_cycle,
			EXTRACT(EPOCH FROM current_period_start)::bigint * 1000,
			EXTRACT(EPOCH FROM current_period_end)::bigint * 1000,
			cancel_at_period_end
		FROM marketplace_plan_subscriptions
		WHERE subscriber_tenant_id = $1::uuid OR subscriber_user_id = NULLIF($2, '')::uuid
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`,
		tenantID, userID, limit, offset,
	)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return []BuyerSubscriptionItem{}, nil
		}
		return nil, fmt.Errorf("list buyer subscriptions: %w", err)
	}
	defer rows.Close()

	items := make([]BuyerSubscriptionItem, 0)
	for rows.Next() {
		var item BuyerSubscriptionItem
		if err := rows.Scan(
			&item.ID, &item.PlanName, &item.CreatorTenantID, &item.Status, &item.Amount, &item.Currency,
			&item.BillingCycle, &item.CurrentPeriodStart, &item.CurrentPeriodEnd, &item.CancelAtPeriodEnd,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
