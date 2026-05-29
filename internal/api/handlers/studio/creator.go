package studio

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// CreatorProfile aggregates publisher stats for the authenticated tenant.
type CreatorProfile struct {
	ID              string  `json:"id"`
	Username        string  `json:"username"`
	Name            string  `json:"name"`
	Avatar          *string `json:"avatar,omitempty"`
	ProfileURL      string  `json:"profile_url"`
	TotalFunctions  int     `json:"total_functions"`
	TotalDownloads  int     `json:"total_downloads"`
	TotalRevenue    float64 `json:"total_revenue"`
	AverageRating   float64 `json:"average_rating"`
	ActiveUsers     int     `json:"active_users"`
}

// UsageMetric represents a usage billing metric for creator tools.
type UsageMetric struct {
	Name  string  `json:"name"`
	Used  float64 `json:"used"`
	Limit float64 `json:"limit"`
	Unit  string  `json:"unit"`
	Cost  float64 `json:"cost"`
}

// UsageBillingSummary is returned by GET /marketplace/usage.
type UsageBillingSummary struct {
	Metrics      []UsageMetric `json:"metrics"`
	TotalCost    float64       `json:"total_cost"`
	BillingCycle string        `json:"billing_cycle"`
	Currency     string        `json:"currency"`
}

// CustomerSubscription represents a subscriber to a creator plan.
type CustomerSubscription struct {
	ID                 string  `json:"id"`
	CustomerID         string  `json:"customer_id"`
	CustomerName       string  `json:"customer_name"`
	CustomerEmail      string  `json:"customer_email"`
	Plan               string  `json:"plan"`
	Status             string  `json:"status"`
	BillingCycle       string  `json:"billing_cycle"`
	Amount             float64 `json:"amount"`
	Currency           string  `json:"currency"`
	CurrentPeriodStart int64   `json:"current_period_start"`
	CurrentPeriodEnd   int64   `json:"current_period_end"`
	CancelAtPeriodEnd  bool    `json:"cancel_at_period_end"`
}

// RoyaltyRecord is the detailed royalty shape expected by the economy UI.
type RoyaltyRecord struct {
	ID                string  `json:"id"`
	FunctionID        string  `json:"function_id"`
	FunctionName      string  `json:"function_name"`
	Licensee          string  `json:"licensee"`
	LicenseType       string  `json:"license_type"`
	RoyaltyPercentage float64 `json:"royalty_percentage"`
	SaleAmount        float64 `json:"sale_amount"`
	RoyaltyAmount     float64 `json:"royalty_amount"`
	Currency          string  `json:"currency"`
	SaleDate          int64   `json:"sale_date"`
	PaidOut           bool    `json:"paid_out"`
}

// RevenueDataPoint is a time-series revenue bucket for analytics charts.
type RevenueDataPoint struct {
	Timestamp    int64   `json:"timestamp"`
	Revenue      float64 `json:"revenue"`
	Subscription float64 `json:"subscription"`
	OneTime      float64 `json:"one_time"`
	Royalty      float64 `json:"royalty"`
}

// LeaderboardEntry represents a ranked marketplace entry.
type LeaderboardEntry struct {
	Rank         int     `json:"rank"`
	CreatorID    string  `json:"creator_id"`
	CreatorName  string  `json:"creator_name"`
	CreatorAvatar string `json:"creator_avatar,omitempty"`
	FunctionID   string  `json:"function_id"`
	FunctionName string  `json:"function_name"`
	Sales        int     `json:"sales"`
	Revenue      float64 `json:"revenue"`
	Rating       float64 `json:"rating"`
	Trend        string  `json:"trend"`
}

func parseTimeRange(timeRange string) time.Time {
	now := time.Now().UTC()
	switch strings.ToLower(strings.TrimSpace(timeRange)) {
	case "7d":
		return now.AddDate(0, 0, -7)
	case "90d":
		return now.AddDate(0, 0, -90)
	case "all":
		return time.Time{}
	default:
		return now.AddDate(0, 0, -30)
	}
}

func parseRevenuePeriod(period string) time.Time {
	now := time.Now().UTC()
	switch strings.ToLower(strings.TrimSpace(period)) {
	case "7d":
		return now.AddDate(0, 0, -7)
	case "90d":
		return now.AddDate(0, 0, -90)
	case "1y":
		return now.AddDate(-1, 0, 0)
	default:
		return now.AddDate(0, 0, -30)
	}
}

// GetCreatorProfile returns aggregated creator stats for a tenant.
func (r *MarketplaceRepository) GetCreatorProfile(ctx context.Context, tenantID, userID, displayName string) (*CreatorProfile, error) {
	var (
		totalFunctions int
		totalDownloads int
		avgRating      sql.NullFloat64
		totalRevenue   sql.NullFloat64
		activeUsers    int
	)

	err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*)::int,
			COALESCE(SUM(rf.popularity_score), 0)::int,
			AVG(rfr.overall_score / 20.0)
		FROM registry_functions rf
		LEFT JOIN registry_function_ratings rfr ON rfr.function_id = rf.id
		WHERE rf.tenant_id = $1::uuid`,
		tenantID,
	).Scan(&totalFunctions, &totalDownloads, &avgRating)
	if err != nil {
		return nil, fmt.Errorf("creator profile functions: %w", err)
	}

	_ = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(net_amount_cents), 0) / 100.0
		FROM publisher_earnings
		WHERE tenant_id = $1::uuid`,
		tenantID,
	).Scan(&totalRevenue)

	_ = r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT e.caller_tenant_id)::int
		FROM registry_function_executions e
		INNER JOIN registry_functions rf ON rf.id = e.function_id
		WHERE rf.tenant_id = $1::uuid
		  AND e.timestamp >= NOW() - INTERVAL '30 days'`,
		tenantID,
	).Scan(&activeUsers)

	username := strings.ToLower(strings.ReplaceAll(displayName, " ", "-"))
	if username == "" {
		username = "creator"
	}

	profile := &CreatorProfile{
		ID:             tenantID,
		Username:       username,
		Name:           displayName,
		ProfileURL:     "/profile/" + username,
		TotalFunctions: totalFunctions,
		TotalDownloads: totalDownloads,
		ActiveUsers:    activeUsers,
	}
	if totalRevenue.Valid {
		profile.TotalRevenue = totalRevenue.Float64
	}
	if avgRating.Valid {
		profile.AverageRating = avgRating.Float64
	}
	return profile, nil
}

// GetUsageBilling returns marketplace usage metrics for a creator tenant.
func (r *MarketplaceRepository) GetUsageBilling(ctx context.Context, tenantID string) (*UsageBillingSummary, error) {
	var (
		apiCalls   int
		storageMB  float64
		bandwidthMB float64
		totalCost  float64
	)

	_ = r.db.QueryRowContext(ctx, `
		SELECT COUNT(e.id)::int
		FROM registry_function_executions e
		INNER JOIN registry_functions rf ON rf.id = e.function_id
		WHERE rf.tenant_id = $1::uuid
		  AND e.timestamp >= date_trunc('month', NOW())`,
		tenantID,
	).Scan(&apiCalls)

	_ = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(net_amount_cents), 0) / 100.0
		FROM publisher_earnings
		WHERE tenant_id = $1::uuid
		  AND earned_at >= date_trunc('month', NOW())`,
		tenantID,
	).Scan(&totalCost)

	metrics := []UsageMetric{
		{Name: "API Calls", Used: float64(apiCalls), Limit: 5000, Unit: "calls", Cost: 0},
		{Name: "Storage", Used: storageMB, Limit: 5, Unit: "GB", Cost: 0},
		{Name: "Bandwidth", Used: bandwidthMB, Limit: 10, Unit: "GB", Cost: 0},
	}

	return &UsageBillingSummary{
		Metrics:      metrics,
		TotalCost:    totalCost,
		BillingCycle: "monthly",
		Currency:     "USD",
	}, nil
}

// ListCustomerSubscriptions returns subscribers for a creator tenant.
func (r *MarketplaceRepository) ListCustomerSubscriptions(ctx context.Context, tenantID string) ([]CustomerSubscription, int, int, int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id::text,
			COALESCE(subscriber_user_id::text, subscriber_tenant_id::text, ''),
			subscriber_name,
			subscriber_email,
			plan_name,
			status,
			billing_cycle,
			amount,
			currency,
			EXTRACT(EPOCH FROM current_period_start)::bigint * 1000,
			EXTRACT(EPOCH FROM current_period_end)::bigint * 1000,
			cancel_at_period_end
		FROM marketplace_plan_subscriptions
		WHERE creator_tenant_id = $1::uuid
		ORDER BY created_at DESC
		LIMIT 100`,
		tenantID,
	)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return []CustomerSubscription{}, 0, 0, 0, nil
		}
		return nil, 0, 0, 0, fmt.Errorf("list subscriptions: %w", err)
	}
	defer rows.Close()

	subs := make([]CustomerSubscription, 0)
	var active, cancelled, pastDue int
	for rows.Next() {
		var sub CustomerSubscription
		if err := rows.Scan(
			&sub.ID, &sub.CustomerID, &sub.CustomerName, &sub.CustomerEmail,
			&sub.Plan, &sub.Status, &sub.BillingCycle, &sub.Amount, &sub.Currency,
			&sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CancelAtPeriodEnd,
		); err != nil {
			return nil, 0, 0, 0, fmt.Errorf("scan subscription: %w", err)
		}
		switch sub.Status {
		case "active", "trialing":
			active++
		case "cancelled":
			cancelled++
		case "past_due":
			pastDue++
		}
		subs = append(subs, sub)
	}
	return subs, active, cancelled, pastDue, rows.Err()
}

// ListRoyaltyRecords returns license-grant based royalty records for the UI.
func (r *MarketplaceRepository) ListRoyaltyRecords(ctx context.Context, tenantID string) ([]RoyaltyRecord, float64, float64, error) {
	const royaltyRate = 0.7

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			g.id::text,
			g.function_id,
			g.function_name,
			g.purchaser_name,
			g.license_type,
			COALESCE(rf.price_per_call, 0) * 100 AS sale_amount,
			COALESCE(rf.price_per_call, 0) * $2 AS royalty_amount,
			g.created_at,
			(g.revoked_at IS NOT NULL) AS paid_out
		FROM marketplace_license_grants g
		INNER JOIN registry_functions rf ON rf.id::text = g.function_id
		WHERE g.tenant_id = $1::uuid
		ORDER BY g.created_at DESC
		LIMIT 100`,
		tenantID, royaltyRate,
	)
	if err != nil {
		return r.listRoyaltyRecordsFromExecutions(ctx, tenantID)
	}
	defer rows.Close()

	records := make([]RoyaltyRecord, 0)
	var totalEarned, totalPending float64
	for rows.Next() {
		var rec RoyaltyRecord
		var saleDate time.Time
		var paidOut bool
		if err := rows.Scan(
			&rec.ID, &rec.FunctionID, &rec.FunctionName, &rec.Licensee, &rec.LicenseType,
			&rec.SaleAmount, &rec.RoyaltyAmount, &saleDate, &paidOut,
		); err != nil {
			return nil, 0, 0, fmt.Errorf("scan royalty record: %w", err)
		}
		rec.RoyaltyPercentage = royaltyRate * 100
		rec.Currency = "USD"
		rec.SaleDate = saleDate.UTC().UnixMilli()
		rec.PaidOut = paidOut
		records = append(records, rec)
		totalEarned += rec.RoyaltyAmount
		if !paidOut {
			totalPending += rec.RoyaltyAmount
		}
	}
	if len(records) == 0 {
		return r.listRoyaltyRecordsFromExecutions(ctx, tenantID)
	}
	return records, totalEarned, totalPending, rows.Err()
}

func (r *MarketplaceRepository) listRoyaltyRecordsFromExecutions(ctx context.Context, tenantID string) ([]RoyaltyRecord, float64, float64, error) {
	const royaltyRate = 0.7
	entries, total, pending, err := r.ListRoyalties(ctx, tenantID)
	if err != nil {
		return nil, 0, 0, err
	}
	records := make([]RoyaltyRecord, 0, len(entries))
	for _, e := range entries {
		records = append(records, RoyaltyRecord{
			ID:                e.ID,
			FunctionID:        e.FunctionID,
			FunctionName:      e.FunctionName,
			Licensee:          "Marketplace usage",
			LicenseType:       "commercial",
			RoyaltyPercentage: e.RoyaltyRate * 100,
			SaleAmount:        e.Earnings / royaltyRate,
			RoyaltyAmount:     e.Earnings,
			Currency:          "USD",
			SaleDate:          e.CreatedAt.UTC().UnixMilli(),
			PaidOut:           e.Paid,
		})
	}
	return records, total, pending, nil
}

// GetRevenueAnalytics returns time-series revenue for creator analytics.
func (r *MarketplaceRepository) GetRevenueAnalytics(ctx context.Context, tenantID, period string) ([]RevenueDataPoint, float64, error) {
	since := parseRevenuePeriod(period)
	bucket := "day"
	if period == "1y" || period == "90d" {
		bucket = "month"
	}

	var (
		rows *sql.Rows
		err  error
	)
	if since.IsZero() {
		rows, err = r.db.QueryContext(ctx, `
			SELECT date_trunc($2, earned_at), COALESCE(SUM(net_amount_cents), 0) / 100.0
			FROM publisher_earnings
			WHERE tenant_id = $1::uuid
			GROUP BY 1
			ORDER BY 1 ASC`, tenantID, bucket)
	} else {
		rows, err = r.db.QueryContext(ctx, `
			SELECT date_trunc($3, earned_at), COALESCE(SUM(net_amount_cents), 0) / 100.0
			FROM publisher_earnings
			WHERE tenant_id = $1::uuid AND earned_at >= $2
			GROUP BY 1
			ORDER BY 1 ASC`, tenantID, since, bucket)
	}
	if err != nil {
		return r.revenueFromExecutions(ctx, tenantID, since)
	}
	defer rows.Close()

	points := make([]RevenueDataPoint, 0)
	var total float64
	for rows.Next() {
		var bucketTime time.Time
		var revenue float64
		if err := rows.Scan(&bucketTime, &revenue); err != nil {
			return nil, 0, fmt.Errorf("scan revenue point: %w", err)
		}
		total += revenue
		points = append(points, RevenueDataPoint{
			Timestamp: bucketTime.UTC().UnixMilli(),
			Revenue:   revenue,
			Royalty:   revenue,
		})
	}
	if len(points) == 0 {
		return r.revenueFromExecutions(ctx, tenantID, since)
	}
	return points, total, rows.Err()
}

func (r *MarketplaceRepository) revenueFromExecutions(ctx context.Context, tenantID string, since time.Time) ([]RevenueDataPoint, float64, error) {
	query := `
		SELECT
			date_trunc('day', e.timestamp) AS bucket,
			COALESCE(SUM(rf.price_per_call * 0.7), 0) AS revenue
		FROM registry_function_executions e
		INNER JOIN registry_functions rf ON rf.id = e.function_id
		WHERE rf.tenant_id = $1::uuid`
	args := []interface{}{tenantID}
	if !since.IsZero() {
		query += ` AND e.timestamp >= $2`
		args = append(args, since)
	}
	query += ` GROUP BY bucket ORDER BY bucket ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return []RevenueDataPoint{}, 0, nil
	}
	defer rows.Close()

	points := make([]RevenueDataPoint, 0)
	var total float64
	for rows.Next() {
		var bucketTime time.Time
		var revenue float64
		if err := rows.Scan(&bucketTime, &revenue); err != nil {
			return nil, 0, err
		}
		total += revenue
		points = append(points, RevenueDataPoint{
			Timestamp: bucketTime.UTC().UnixMilli(),
			Revenue:   revenue,
			Royalty:   revenue,
		})
	}
	return points, total, rows.Err()
}

// GetLeaderboard returns ranked marketplace entries.
func (r *MarketplaceRepository) GetLeaderboard(ctx context.Context, category, timeRange string, limit int) ([]LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	since := parseTimeRange(timeRange)

	switch category {
	case "creators":
		return r.leaderboardByCreators(ctx, since, limit)
	case "revenue":
		return r.leaderboardByRevenue(ctx, since, limit)
	default:
		return r.leaderboardByFunctions(ctx, since, limit)
	}
}

func (r *MarketplaceRepository) leaderboardByFunctions(ctx context.Context, since time.Time, limit int) ([]LeaderboardEntry, error) {
	query := `
		SELECT
			rf.id::text,
			rf.author,
			rf.name,
			COUNT(e.id)::int AS sales,
			COALESCE(SUM(rf.price_per_call * 0.7), 0) AS revenue,
			COALESCE(rfr.overall_score / 20.0, 0) AS rating
		FROM registry_functions rf
		LEFT JOIN registry_function_executions e ON e.function_id = rf.id`
	args := []interface{}{}
	argIdx := 1
	if !since.IsZero() {
		query += fmt.Sprintf(` AND e.timestamp >= $%d`, argIdx)
		args = append(args, since)
		argIdx++
	}
	query += `
		LEFT JOIN registry_function_ratings rfr ON rfr.function_id = rf.id
		WHERE rf.visibility IN ('public', 'unlisted')
		GROUP BY rf.id, rf.author, rf.name, rfr.overall_score
		ORDER BY revenue DESC, sales DESC`
	query += fmt.Sprintf(` LIMIT $%d`, argIdx)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("leaderboard functions: %w", err)
	}
	defer rows.Close()
	return scanLeaderboardRows(rows)
}

func (r *MarketplaceRepository) leaderboardByCreators(ctx context.Context, since time.Time, limit int) ([]LeaderboardEntry, error) {
	query := `
		SELECT
			MIN(rf.id::text),
			rf.author,
			MIN(rf.name),
			COUNT(e.id)::int,
			COALESCE(SUM(rf.price_per_call * 0.7), 0),
			COALESCE(AVG(rfr.overall_score / 20.0), 0)
		FROM registry_functions rf
		LEFT JOIN registry_function_executions e ON e.function_id = rf.id`
	args := []interface{}{}
	argIdx := 1
	if !since.IsZero() {
		query += fmt.Sprintf(` AND e.timestamp >= $%d`, argIdx)
		args = append(args, since)
		argIdx++
	}
	query += `
		LEFT JOIN registry_function_ratings rfr ON rfr.function_id = rf.id
		WHERE rf.visibility IN ('public', 'unlisted')
		GROUP BY rf.author
		ORDER BY COALESCE(SUM(rf.price_per_call * 0.7), 0) DESC`
	query += fmt.Sprintf(` LIMIT $%d`, argIdx)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("leaderboard creators: %w", err)
	}
	defer rows.Close()
	return scanLeaderboardRows(rows)
}

func (r *MarketplaceRepository) leaderboardByRevenue(ctx context.Context, since time.Time, limit int) ([]LeaderboardEntry, error) {
	return r.leaderboardByFunctions(ctx, since, limit)
}

func scanLeaderboardRows(rows *sql.Rows) ([]LeaderboardEntry, error) {
	entries := make([]LeaderboardEntry, 0)
	rank := 0
	var prevRevenue float64
	for rows.Next() {
		var entry LeaderboardEntry
		if err := rows.Scan(
			&entry.FunctionID, &entry.CreatorName, &entry.FunctionName,
			&entry.Sales, &entry.Revenue, &entry.Rating,
		); err != nil {
			return nil, fmt.Errorf("scan leaderboard: %w", err)
		}
		rank++
		entry.Rank = rank
		entry.CreatorID = strings.ToLower(strings.ReplaceAll(entry.CreatorName, " ", "-"))
		if entry.Revenue > prevRevenue {
			entry.Trend = "up"
		} else if entry.Revenue < prevRevenue {
			entry.Trend = "down"
		} else {
			entry.Trend = "stable"
		}
		prevRevenue = entry.Revenue
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func ensureDefaultPlans(ctx context.Context, db *sql.DB, tenantID string) error {
	var count int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM marketplace_subscription_plans WHERE tenant_id = $1::uuid`,
		tenantID,
	).Scan(&count); err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			return nil
		}
		return err
	}
	if count > 0 {
		return nil
	}
	featuresFree, _ := json.Marshal([]string{"5 calls/min", "Basic support"})
	featuresPro, _ := json.Marshal([]string{"100 calls/min", "Priority support", "Analytics"})
	featuresEnt, _ := json.Marshal([]string{"Unlimited", "Dedicated support", "Custom SLA"})
	_, err := db.ExecContext(ctx, `
		INSERT INTO marketplace_subscription_plans (tenant_id, name, price, features, billing_cycle)
		VALUES
			($1::uuid, 'Free', 0, $2::jsonb, 'monthly'),
			($1::uuid, 'Pro', 9.99, $3::jsonb, 'monthly'),
			($1::uuid, 'Enterprise', 49.99, $4::jsonb, 'monthly')`,
		tenantID, featuresFree, featuresPro, featuresEnt,
	)
	return err
}
