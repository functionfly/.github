package consciousness

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	MarketplaceSimilarityThresholdEnv  = "CONSCIOUSNESS_MARKETPLACE_SIMILARITY_THRESHOLD"
	MarketplaceMaxInsightsEnv          = "CONSCIOUSNESS_MARKETPLACE_MAX_INSIGHTS"
	DefaultMarketplaceSimilarityThreshold = 0.75
	DefaultMarketplaceMaxInsights         = 5
)

// MarketplaceAnalyzer discovers marketplace functions that could replace or
// improve the tenant's existing custom functions.
//
// It uses two strategies:
//  1. FlyEmbed triple-vector matching — finds registry functions whose contract,
//     semantic, and code embeddings are highly similar to the tenant's managed
//     functions, then compares pricing and performance.
//  2. Category trending — surfaces recently published, high-trust functions
//     in categories where the tenant already has functions.
type MarketplaceAnalyzer struct {
	db                  *sql.DB
	logger              *logrus.Logger
	maxInsights         int
	similarityThreshold float64
}

func NewMarketplaceAnalyzer(db *sql.DB, logger *logrus.Logger) *MarketplaceAnalyzer {
	return &MarketplaceAnalyzer{
		db:                  db,
		logger:              logger,
		maxInsights:         loadMarketplaceMaxInsights(),
		similarityThreshold: loadMarketplaceSimilarityThreshold(),
	}
}

func (a *MarketplaceAnalyzer) Name() string              { return "marketplace" }
func (a *MarketplaceAnalyzer) Category() InsightCategory { return CategoryMarketplace }

func loadMarketplaceSimilarityThreshold() float64 {
	if v := os.Getenv(MarketplaceSimilarityThresholdEnv); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 && f <= 1 {
			return f
		}
	}
	return DefaultMarketplaceSimilarityThreshold
}

func loadMarketplaceMaxInsights() int {
	if v := os.Getenv(MarketplaceMaxInsightsEnv); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			return i
		}
	}
	return DefaultMarketplaceMaxInsights
}

// Analyze finds marketplace functions that match the tenant's existing functions.
func (a *MarketplaceAnalyzer) Analyze(ctx context.Context, tenantID uuid.UUID, params AnalysisParams) ([]*Insight, error) {
	var insights []*Insight

	embedInsights, err := a.findFlyEmbedMatches(ctx, tenantID)
	if err != nil {
		a.logger.WithError(err).Warn("FlyEmbed marketplace matching failed")
	} else {
		insights = append(insights, embedInsights...)
	}

	trendInsights, err := a.findCategoryTrends(ctx, tenantID, params)
	if err != nil {
		a.logger.WithError(err).Warn("Category trend scan failed")
	} else {
		insights = append(insights, trendInsights...)
	}

	if len(insights) > a.maxInsights {
		insights = insights[:a.maxInsights]
	}

	return insights, nil
}

// findFlyEmbedMatches finds registry functions with high triple-vector similarity
// to the tenant's managed functions. These are potential drop-in replacements.
func (a *MarketplaceAnalyzer) findFlyEmbedMatches(ctx context.Context, tenantID uuid.UUID) ([]*Insight, error) {
	query := `
		WITH tenant_triples AS (
			SELECT t.function_id, t.contract_embedding, t.semantic_embedding, t.code_embedding,
				rf.name, rf.title, rf.price_per_call, rf.category
			FROM function_embedding_triples t
			JOIN registry_functions rf ON rf.id = t.function_id
			WHERE rf.tenant_id = $1
			AND rf.visibility = 'public'
			AND t.contract_embedding IS NOT NULL
			AND t.semantic_embedding IS NOT NULL
			AND t.code_embedding IS NOT NULL
			LIMIT 50
		),
		tenant_categories AS (
			SELECT DISTINCT category FROM tenant_triples WHERE category IS NOT NULL
		),
		marketplace_triples AS (
			SELECT t.function_id, t.contract_embedding, t.semantic_embedding, t.code_embedding,
				rf.author, rf.name, rf.title, rf.description, rf.category,
				rf.price_per_call, rf.reliability_score, rf.trust_score,
				rf.popularity_score, rf.created_at
			FROM function_embedding_triples t
			JOIN registry_functions rf ON rf.id = t.function_id
			WHERE rf.visibility = 'public'
			AND rf.tenant_id != $1
			AND rf.trust_score >= 0.5
			AND t.contract_embedding IS NOT NULL
			AND t.semantic_embedding IS NOT NULL
			AND t.code_embedding IS NOT NULL
			AND rf.category IN (SELECT category FROM tenant_categories)
			LIMIT 500
		),
		scored_matches AS (
			SELECT
				tt.function_id AS tenant_func_id, tt.name AS tenant_func_name, tt.title AS tenant_func_title,
				tt.price_per_call AS tenant_price,
				mt.function_id AS market_func_id, mt.author AS market_author, mt.name AS market_name,
				mt.title AS market_title, mt.description AS market_description,
				mt.category AS market_category,
				mt.price_per_call AS market_price, mt.reliability_score AS market_reliability,
				mt.trust_score AS market_trust, mt.popularity_score AS market_popularity,
				mt.created_at AS market_published,
				(0.35 * (1 - (tt.contract_embedding <=> mt.contract_embedding)) +
				  0.40 * (1 - (tt.semantic_embedding <=> mt.semantic_embedding)) +
				  0.25 * (1 - (tt.code_embedding <=> mt.code_embedding))) AS combined_sim
			FROM tenant_triples tt
			CROSS JOIN LATERAL (
				SELECT * FROM marketplace_triples mt
				WHERE (0.35 * (1 - (tt.contract_embedding <=> mt.contract_embedding)) +
					   0.40 * (1 - (tt.semantic_embedding <=> mt.semantic_embedding)) +
					   0.25 * (1 - (tt.code_embedding <=> mt.code_embedding))) > $2
				ORDER BY (0.35 * (1 - (tt.contract_embedding <=> mt.contract_embedding)) +
					     0.40 * (1 - (tt.semantic_embedding <=> mt.semantic_embedding)) +
					     0.25 * (1 - (tt.code_embedding <=> mt.code_embedding))) DESC
				LIMIT 5
			) mt
		)
		SELECT
			tenant_func_id, tenant_func_name, tenant_func_title, tenant_price,
			market_func_id, market_author, market_name,
			market_title, market_description, market_category,
			market_price, market_reliability, market_trust, market_popularity,
			market_published, combined_sim
		FROM scored_matches
		ORDER BY combined_sim DESC
		LIMIT $3`

	rows, err := a.db.QueryContext(ctx, query, tenantID, a.similarityThreshold, a.maxInsights)
	if err != nil {
		return nil, fmt.Errorf("flyembed marketplace query: %w", err)
	}
	defer rows.Close()

	var insights []*Insight
	for rows.Next() {
		var (
			tenantFuncID                            uuid.UUID
			tenantFuncName                          string
			tenantFuncTitle                         sql.NullString
			tenantPrice                             float64
			marketFuncID                            uuid.UUID
			marketAuthor, marketName                string
			marketTitle, marketDesc, marketCategory sql.NullString
			marketPrice                             float64
			marketReliability, marketTrust          float64
			marketPopularity                        int
			marketPublished                         time.Time
			combinedSim                             float64
		)

		if err := rows.Scan(
			&tenantFuncID, &tenantFuncName, &tenantFuncTitle, &tenantPrice,
			&marketFuncID, &marketAuthor, &marketName,
			&marketTitle, &marketDesc, &marketCategory,
			&marketPrice, &marketReliability, &marketTrust, &marketPopularity,
			&marketPublished, &combinedSim,
		); err != nil {
			a.logger.WithError(err).Error("Failed to scan marketplace match row")
			continue
		}

		displayTenant := tenantFuncName
		if tenantFuncTitle.Valid && tenantFuncTitle.String != "" {
			displayTenant = tenantFuncTitle.String
		}
		displayMarket := marketName
		if marketTitle.Valid && marketTitle.String != "" {
			displayMarket = marketTitle.String
		}

		// Determine the improvement narrative
		var improvements []string
		if marketPrice < tenantPrice && tenantPrice > 0 {
			savingsPct := ((tenantPrice - marketPrice) / tenantPrice) * 100
			improvements = append(improvements, fmt.Sprintf("%.0f%% cheaper per call", savingsPct))
		}
		if marketReliability > 0.95 {
			improvements = append(improvements, fmt.Sprintf("%.0f%% reliability", marketReliability*100))
		}
		if marketTrust > 0.8 {
			improvements = append(improvements, "high trust score")
		}

		improvementText := "a comparable alternative"
		if len(improvements) > 0 {
			improvementText = ""
			for i, imp := range improvements {
				if i > 0 {
					improvementText += ", "
				}
				improvementText += imp
			}
		}

		matchPct := combinedSim * 100
		severity := SeverityInfo
		if matchPct > 90 {
			severity = SeverityOpportunity
		}

		confidence := combinedSim * 0.9 // Slight discount for cross-tenant comparison
		trajectory := TrajectoryStable
		publishedAgo := time.Since(marketPublished).Truncate(time.Hour)

		insights = append(insights, &Insight{
			TenantID: tenantID,
			Category: CategoryMarketplace,
			Severity: severity,
			Priority: SeverityWeight(severity)*10 + int(matchPct/10),
			Title:    fmt.Sprintf("Marketplace alternative for %s: %s/%s", displayTenant, marketAuthor, marketName),
			Message: fmt.Sprintf(
				"%s/%s (%s) matches your %s at %.0f%% similarity. It offers %s at $%.4f/call. Published %s ago with %d uses.",
				marketAuthor, marketName, displayMarket, displayTenant,
				matchPct, improvementText, marketPrice,
				publishedAgo, marketPopularity,
			),
			Summary:    strPtr(fmt.Sprintf("%s/%s — %s", marketAuthor, marketName, improvementText)),
			FunctionID: &tenantFuncID,
			InsightData: JSONMap{
				"tenant_function_id":      tenantFuncID.String(),
				"tenant_function_name":    tenantFuncName,
				"tenant_price":            tenantPrice,
				"marketplace_author":      marketAuthor,
				"marketplace_name":        marketName,
				"marketplace_title":       displayMarket,
				"marketplace_price":       marketPrice,
				"marketplace_trust":       marketTrust,
				"marketplace_reliability": marketReliability,
				"marketplace_popularity":  marketPopularity,
				"combined_similarity":     combinedSim,
				"improvements":            improvements,
			},
			ActionType: ActionSwapMarketplace,
			ActionData: JSONMap{
				"swap_from":      tenantFuncID.String(),
				"swap_to_author": marketAuthor,
				"swap_to_name":   marketName,
			},
			ActionPreview: JSONMap{
				"current_function":  displayTenant,
				"replacement":       fmt.Sprintf("%s/%s", marketAuthor, marketName),
				"current_price":     tenantPrice,
				"replacement_price": marketPrice,
			},
			Trajectory: &trajectory,
			Confidence: &confidence,
			Status:     StatusActive,
			ExpiresAt:  timePtr(time.Now().Add(14 * 24 * time.Hour)),
		})
	}

	return insights, rows.Err()
}

// findCategoryTrends finds recently published, high-trust marketplace functions
// in categories where the tenant already has functions.
func (a *MarketplaceAnalyzer) findCategoryTrends(ctx context.Context, tenantID uuid.UUID, params AnalysisParams) ([]*Insight, error) {
	// Find categories where the tenant has functions
	query := `
		WITH tenant_categories AS (
			SELECT DISTINCT category
			FROM registry_functions
			WHERE tenant_id = $1
			AND category IS NOT NULL
			AND category != ''
			AND visibility = 'public'
		),
		recent_marketplace AS (
			SELECT rf.id, rf.author, rf.name, rf.title, rf.description,
				rf.category, rf.price_per_call, rf.reliability_score,
				rf.trust_score, rf.popularity_score, rf.created_at
			FROM registry_functions rf
			JOIN tenant_categories tc ON tc.category = rf.category
			WHERE rf.tenant_id != $1
			AND rf.visibility = 'public'
			AND rf.trust_score >= 0.6
			AND rf.created_at > NOW() - INTERVAL '14 days'
			ORDER BY rf.trust_score DESC, rf.popularity_score DESC
			LIMIT 5
		)
		SELECT id, author, name, title, description, category,
			price_per_call, reliability_score, trust_score, popularity_score, created_at
		FROM recent_marketplace`

	rows, err := a.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("category trend query: %w", err)
	}
	defer rows.Close()

	var insights []*Insight
	for rows.Next() {
		var (
			funcID                       uuid.UUID
			author, name                 string
			title, description, category sql.NullString
			price                        float64
			reliability, trust           float64
			popularity                   int
			createdAt                    time.Time
		)

		if err := rows.Scan(
			&funcID, &author, &name, &title, &description,
			&category, &price, &reliability, &trust, &popularity, &createdAt,
		); err != nil {
			a.logger.WithError(err).Error("Failed to scan category trend row")
			continue
		}

		displayName := name
		if title.Valid && title.String != "" {
			displayName = title.String
		}

		catName := "your category"
		if category.Valid && category.String != "" {
			catName = category.String
		}

		descText := ""
		if description.Valid && description.String != "" {
			descText = description.String
			if len(descText) > 120 {
				descText = descText[:120] + "..."
			}
		}

		publishedAgo := time.Since(createdAt).Truncate(time.Hour)
		confidence := trust * 0.8
		trajectory := TrajectoryStable

		insights = append(insights, &Insight{
			TenantID: tenantID,
			Category: CategoryMarketplace,
			Severity: SeverityInfo,
			Priority: 10,
			Title:    fmt.Sprintf("New in %s: %s/%s", catName, author, name),
			Message: fmt.Sprintf(
				"A new marketplace function %s/%s (%s) was published %s ago in %s. %s Price: $%.4f/call, trust: %.0f%%, %d uses.",
				author, name, displayName, publishedAgo, catName,
				descText, price, trust*100, popularity,
			),
			Summary: strPtr(fmt.Sprintf("%s/%s — new in %s", author, name, catName)),
			InsightData: JSONMap{
				"marketplace_author":     author,
				"marketplace_name":       name,
				"marketplace_title":      displayName,
				"marketplace_category":   catName,
				"marketplace_price":      price,
				"marketplace_trust":      trust,
				"marketplace_popularity": popularity,
				"published_ago":          publishedAgo.String(),
				"signal":                 "category_trend",
			},
			Trajectory: &trajectory,
			Confidence: &confidence,
			Status:     StatusActive,
			ExpiresAt:  timePtr(time.Now().Add(7 * 24 * time.Hour)),
		})
	}

	return insights, rows.Err()
}

// ensure interface compliance
var _ Analyzer = (*MarketplaceAnalyzer)(nil)

// unused import guard
var _ = json.Marshal
