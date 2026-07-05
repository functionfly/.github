package companyranking

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// ListCompanies returns active companies ordered by name.
func (r *Repository) ListCompanies(ctx context.Context) ([]Company, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, slug, name, city_id, country_code, COALESCE(employee_count, 0), COALESCE(industry, ''), COALESCE(website, ''), is_active, created_at, updated_at
		FROM companies WHERE is_active = TRUE ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Company
	for rows.Next() {
		var c Company
		var cityID *int64
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name, &cityID, &c.CountryCode, &c.EmployeeCount, &c.Industry, &c.Website, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.CityID = cityID
		out = append(out, c)
	}
	return out, rows.Err()
}

// GetCompanyBySlug returns a company by slug.
func (r *Repository) GetCompanyBySlug(ctx context.Context, slug string) (*Company, error) {
	var c Company
	var cityID *int64
	err := r.pool.QueryRow(ctx, `
		SELECT id, slug, name, city_id, country_code, COALESCE(employee_count, 0), COALESCE(industry, ''), COALESCE(website, ''), is_active, created_at, updated_at
		FROM companies WHERE slug = $1
	`, slug).Scan(&c.ID, &c.Slug, &c.Name, &cityID, &c.CountryCode, &c.EmployeeCount, &c.Industry, &c.Website, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.CityID = cityID
	return &c, nil
}

// ListRankings returns the top companies for a country/category.
func (r *Repository) ListRankings(ctx context.Context, country string, limit int, category Category) ([]Ranking, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	args := []any{category, MinActiveUsersForPublic, limit}
	countryClause := ""
	if country != "" {
		countryClause = " AND co.country_code = $4"
		args = append(args, strings.ToUpper(country))
	}
	rows, err := r.pool.Query(ctx, `
		SELECT cr.company_id, co.slug, co.name, co.country_code, COALESCE(c.slug, ''),
			COALESCE(co.employee_count, 0),
			cr.rank_position, cr.prev_rank_position,
			cr.score_raw, cr.score_per_capita,
			cr.active_users, cr.deployments, cr.executions_30d,
			cr.revenue_cents, cr.new_users_30d,
			cr.period_start, cr.period_end
		FROM company_rankings cr
		JOIN companies co ON co.id = cr.company_id
		LEFT JOIN cities c ON c.id = co.city_id
		WHERE cr.ranking_category = $1
			AND cr.active_users >= $2
			AND co.is_active = TRUE`+countryClause+`
		ORDER BY cr.score_per_capita DESC
		LIMIT $3
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Ranking
	for rows.Next() {
		var rk Ranking
		var prev *int
		if err := rows.Scan(&rk.CompanyID, &rk.CompanySlug, &rk.CompanyName,
			&rk.CountryCode, &rk.CitySlug, &rk.EmployeeCount,
			&rk.RankPosition, &prev,
			&rk.ScoreRaw, &rk.ScorePerCapita,
			&rk.ActiveUsers, &rk.Deployments, &rk.Executions30d,
			&rk.RevenueCents, &rk.NewUsers30d,
			&rk.PeriodStart, &rk.PeriodEnd); err != nil {
			return nil, err
		}
		rk.PrevRank = prev
		rk.Category = category
		if prev != nil {
			rk.RankDelta = rk.RankPosition - *prev
		}
		out = append(out, rk)
	}
	return out, rows.Err()
}

// GetRankingBySlug returns a single company ranking.
func (r *Repository) GetRankingBySlug(ctx context.Context, slug string, category Category) (*Ranking, error) {
	var rk Ranking
	var prev *int
	err := r.pool.QueryRow(ctx, `
		SELECT cr.company_id, co.slug, co.name, co.country_code, COALESCE(c.slug, ''),
			COALESCE(co.employee_count, 0),
			cr.rank_position, cr.prev_rank_position,
			cr.score_raw, cr.score_per_capita,
			cr.active_users, cr.deployments, cr.executions_30d,
			cr.revenue_cents, cr.new_users_30d,
			cr.period_start, cr.period_end
		FROM company_rankings cr
		JOIN companies co ON co.id = cr.company_id
		LEFT JOIN cities c ON c.id = co.city_id
		WHERE co.slug = $1 AND cr.ranking_category = $2
		ORDER BY cr.period_end DESC
		LIMIT 1
	`, slug, category).Scan(&rk.CompanyID, &rk.CompanySlug, &rk.CompanyName,
		&rk.CountryCode, &rk.CitySlug, &rk.EmployeeCount,
		&rk.RankPosition, &prev,
		&rk.ScoreRaw, &rk.ScorePerCapita,
		&rk.ActiveUsers, &rk.Deployments, &rk.Executions30d,
		&rk.RevenueCents, &rk.NewUsers30d,
		&rk.PeriodStart, &rk.PeriodEnd)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rk.PrevRank = prev
	rk.Category = category
	if prev != nil {
		rk.RankDelta = rk.RankPosition - *prev
	}
	return &rk, nil
}

// UpsertRanking inserts or updates a ranking row.
func (r *Repository) UpsertRanking(ctx context.Context, rk Ranking) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO company_rankings
			(company_id, rank_position, prev_rank_position, score_raw, score_per_capita,
			 active_users, deployments, executions_30d, revenue_cents, new_users_30d,
			 period_start, period_end, ranking_category)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (company_id, period_end, ranking_category) DO UPDATE SET
			rank_position = EXCLUDED.rank_position,
			prev_rank_position = EXCLUDED.prev_rank_position,
			score_raw = EXCLUDED.score_raw,
			score_per_capita = EXCLUDED.score_per_capita,
			active_users = EXCLUDED.active_users,
			deployments = EXCLUDED.deployments,
			executions_30d = EXCLUDED.executions_30d,
			revenue_cents = EXCLUDED.revenue_cents,
			new_users_30d = EXCLUDED.new_users_30d,
			computed_at = NOW()
	`, rk.CompanyID, rk.RankPosition, rk.PrevRank, rk.ScoreRaw, rk.ScorePerCapita,
		rk.ActiveUsers, rk.Deployments, rk.Executions30d, rk.RevenueCents, rk.NewUsers30d,
		rk.PeriodStart, rk.PeriodEnd, rk.Category)
	return err
}

// ListAllCompanies returns every active company for recompute.
func (r *Repository) ListAllCompanies(ctx context.Context) ([]Company, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, slug, name, city_id, country_code, COALESCE(employee_count, 0), COALESCE(industry, ''), COALESCE(website, ''), is_active, created_at, updated_at
		FROM companies WHERE is_active = TRUE ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Company
	for rows.Next() {
		var c Company
		var cityID *int64
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name, &cityID, &c.CountryCode, &c.EmployeeCount, &c.Industry, &c.Website, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.CityID = cityID
		out = append(out, c)
	}
	return out, rows.Err()
}

// SignalsForCompany returns activity signals for a company in the last 30 days.
func (r *Repository) SignalsForCompany(ctx context.Context, companyID int64) (Signals, error) {
	var s Signals
	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT t.user_id),
			COUNT(DISTINCT f.id),
			COALESCE(COUNT(rfe.id), 0),
			COALESCE(SUM(rfe.revenue_cents), 0),
			COUNT(DISTINCT CASE WHEN u.created_at >= NOW() - INTERVAL '30 days' THEN u.id END)
		FROM tenants t
		JOIN users u ON u.tenant_id = t.id AND u.is_active = TRUE
		JOIN registry_functions f ON f.owner_user_id = u.id
		LEFT JOIN registry_function_executions rfe ON rfe.user_id = u.id AND rfe.timestamp >= NOW() - INTERVAL '30 days'
		WHERE t.company_id = $1
			AND COALESCE(u.company_ranking_opted_out, FALSE) = FALSE
	`, companyID).Scan(&s.ActiveUsers, &s.Deployments, &s.Executions30d, &s.RevenueCents, &s.NewUsers30d)
	if err != nil {
		return Signals{}, err
	}
	return s, nil
}

// SortRankings assigns rank positions to a slice of rankings sorted by per-capita descending.
func SortRankings(rs []Ranking) {
	for i := range rs {
		if i == 0 {
			rs[i].RankPosition = 1
			continue
		}
		if rs[i].ScorePerCapita == rs[i-1].ScorePerCapita {
			rs[i].RankPosition = rs[i-1].RankPosition
		} else {
			rs[i].RankPosition = i + 1
		}
	}
}

// TruncateHour truncates t to the hour.
func TruncateHour(t time.Time) time.Time {
	return t.Truncate(time.Hour)
}
