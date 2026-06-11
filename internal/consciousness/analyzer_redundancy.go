package consciousness

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// RedundancyAnalyzer detects functions performing duplicate or overlapping work.
//
// It uses two signals:
//  1. Co-occurrence analysis — functions always invoked together with identical
//     inputs are likely doing redundant work (from function_co_occurrences table).
//  2. FlyEmbed triple-vector similarity — functions with high contract, semantic,
//     and code overlap (>0.85 combined) are candidates for merging
//     (from function_embedding_triples table).
type RedundancyAnalyzer struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewRedundancyAnalyzer(db *sql.DB, logger *logrus.Logger) *RedundancyAnalyzer {
	return &RedundancyAnalyzer{db: db, logger: logger}
}

func (a *RedundancyAnalyzer) Name() string              { return "redundancy" }
func (a *RedundancyAnalyzer) Category() InsightCategory { return CategoryRedundancy }

const (
	// redundancySimilarityThreshold is the minimum combined FlyEmbed score
	// to flag two functions as potentially redundant.
	redundancySimilarityThreshold = 0.82

	// redundancyCoOccurrenceMin is the minimum co-occurrence count to consider.
	redundancyCoOccurrenceMin = 50

	// redundancyMaxPairs is the maximum number of redundant pairs to report per run.
	redundancyMaxPairs = 5
)

// Analyze finds function pairs with high code/semantic overlap.
func (a *RedundancyAnalyzer) Analyze(ctx context.Context, tenantID uuid.UUID, params AnalysisParams) ([]*Insight, error) {
	var insights []*Insight

	// ── Signal 1: FlyEmbed triple-vector similarity ──────────────────────────
	// Find pairs of functions owned by this tenant with combined similarity > threshold.
	tripleInsights, err := a.findTripleSimilarPairs(ctx, tenantID)
	if err != nil {
		a.logger.WithError(err).Warn("Triple-vector redundancy scan failed")
	} else {
		insights = append(insights, tripleInsights...)
	}

	// ── Signal 2: High co-occurrence with same category ──────────────────────
	// Functions that are always called together and share a category are suspicious.
	coInsights, err := a.findCoOccurrenceRedundancy(ctx, tenantID)
	if err != nil {
		a.logger.WithError(err).Warn("Co-occurrence redundancy scan failed")
	} else {
		insights = append(insights, coInsights...)
	}

	// Cap output
	if len(insights) > redundancyMaxPairs {
		insights = insights[:redundancyMaxPairs]
	}

	return insights, nil
}

// findTripleSimilarPairs uses the function_embedding_triples table to find
// functions with high combined similarity across contract, semantic, and code vectors.
func (a *RedundancyAnalyzer) findTripleSimilarPairs(ctx context.Context, tenantID uuid.UUID) ([]*Insight, error) {
	query := `
		WITH tenant_functions AS (
			SELECT rf.id, rf.author, rf.name, rf.title, rf.description,
				t.contract_embedding, t.semantic_embedding, t.code_embedding
			FROM registry_functions rf
			JOIN function_embedding_triples t ON t.function_id = rf.id
			WHERE rf.tenant_id = $1
			AND rf.visibility = 'public'
			AND t.contract_embedding IS NOT NULL
			AND t.semantic_embedding IS NOT NULL
			AND t.code_embedding IS NOT NULL
		)
		SELECT
			a.id AS func_a_id, a.name AS func_a_name, a.title AS func_a_title,
			b.id AS func_b_id, b.name AS func_b_name, b.title AS func_b_title,
			(1 - (a.contract_embedding <=> b.contract_embedding)) AS contract_sim,
			(1 - (a.semantic_embedding <=> b.semantic_embedding)) AS semantic_sim,
			(1 - (a.code_embedding <=> b.code_embedding)) AS code_sim,
			(0.35 * (1 - (a.contract_embedding <=> b.contract_embedding)) +
			  0.40 * (1 - (a.semantic_embedding <=> b.semantic_embedding)) +
			  0.25 * (1 - (a.code_embedding <=> b.code_embedding))) AS combined_sim
		FROM tenant_functions a
		JOIN tenant_functions b ON a.id < b.id
		WHERE (0.35 * (1 - (a.contract_embedding <=> b.contract_embedding)) +
			   0.40 * (1 - (a.semantic_embedding <=> b.semantic_embedding)) +
			   0.25 * (1 - (a.code_embedding <=> b.code_embedding))) > $2
		ORDER BY combined_sim DESC
		LIMIT $3`

	rows, err := a.db.QueryContext(ctx, query, tenantID, redundancySimilarityThreshold, redundancyMaxPairs)
	if err != nil {
		if isRelationNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("triple-vector redundancy query: %w", err)
	}
	defer rows.Close()

	var insights []*Insight
	for rows.Next() {
		var (
			funcAID, funcBID         uuid.UUID
			funcAName, funcBName     string
			funcATitle, funcBTitle   sql.NullString
			contractSim, semanticSim float64
			codeSim, combinedSim     float64
		)

		if err := rows.Scan(
			&funcAID, &funcAName, &funcATitle,
			&funcBID, &funcBName, &funcBTitle,
			&contractSim, &semanticSim, &codeSim, &combinedSim,
		); err != nil {
			a.logger.WithError(err).Warn("Failed to scan triple redundancy row")
			continue
		}

		displayA := funcAName
		if funcATitle.Valid && funcATitle.String != "" {
			displayA = funcATitle.String
		}
		displayB := funcBName
		if funcBTitle.Valid && funcBTitle.String != "" {
			displayB = funcBTitle.String
		}

		overlapPct := combinedSim * 100
		severity := SeverityWarning
		if overlapPct > 92 {
			severity = SeverityOpportunity
		}

		confidence := combinedSim
		trajectory := TrajectoryStable

		insights = append(insights, &Insight{
			TenantID: tenantID,
			Category: CategoryRedundancy,
			Severity: severity,
			Priority: SeverityWeight(severity)*10 + int(overlapPct/10),
			Title:    fmt.Sprintf("%.0f%% overlap between %s and %s", overlapPct, displayA, displayB),
			Message: fmt.Sprintf(
				"%s and %s share %.0f%% similarity across contract (%.0f%%), semantics (%.0f%%), and code (%.0f%%). Merging them would reduce maintenance surface and save on execution costs.",
				displayA, displayB, overlapPct,
				contractSim*100, semanticSim*100, codeSim*100,
			),
			Summary: strPtr(fmt.Sprintf("%.0f%% overlap — merge candidates", overlapPct)),
			InsightData: JSONMap{
				"function_a_id":   funcAID.String(),
				"function_a_name": funcAName,
				"function_b_id":   funcBID.String(),
				"function_b_name": funcBName,
				"contract_sim":    contractSim,
				"semantic_sim":    semanticSim,
				"code_sim":        codeSim,
				"combined_sim":    combinedSim,
				"signal":          "flyembed_triple_vector",
			},
			ActionType: ActionMergeFunctions,
			ActionData: JSONMap{
				"merge_from": funcBID.String(),
				"merge_into": funcAID.String(),
			},
			ActionPreview: JSONMap{
				"would_remove":                  funcBName,
				"would_keep":                    funcAName,
				"estimated_savings_description": "Reduced duplicate executions and maintenance",
			},
			RelatedFunctionIDs: []uuid.UUID{funcAID, funcBID},
			Trajectory:         &trajectory,
			Confidence:         &confidence,
			Status:             StatusActive,
			ExpiresAt:          timePtr(time.Now().Add(14 * 24 * time.Hour)),
		})
	}

	return insights, rows.Err()
}

// findCoOccurrenceRedundancy finds functions frequently called together in the
// same session that also share a category — a strong signal of duplication.
func (a *RedundancyAnalyzer) findCoOccurrenceRedundancy(ctx context.Context, tenantID uuid.UUID) ([]*Insight, error) {
	query := `
		SELECT
			rf_a.id AS func_a_id, rf_a.name AS func_a_name,
			rf_b.id AS func_b_id, rf_b.name AS func_b_name,
			co.co_occurrence_count,
			rf_a.category AS shared_category
		FROM function_co_occurrences co
		JOIN registry_functions rf_a ON rf_a.id = co.function_id_a
		JOIN registry_functions rf_b ON rf_b.id = co.function_id_b
		WHERE (rf_a.tenant_id = $1 OR rf_b.tenant_id = $1)
		AND rf_a.category = rf_b.category
		AND rf_a.category IS NOT NULL
		AND rf_a.category != ''
		AND co.co_occurrence_count >= $2
		AND co.last_cooccurred_at > NOW() - INTERVAL '30 days'
		ORDER BY co.co_occurrence_count DESC
		LIMIT $3`

	rows, err := a.db.QueryContext(ctx, query, tenantID, redundancyCoOccurrenceMin, redundancyMaxPairs)
	if err != nil {
		// Table might not exist in all deployments — gracefully degrade.
		if isRelationNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("co-occurrence redundancy query: %w", err)
	}
	defer rows.Close()

	var insights []*Insight
	for rows.Next() {
		var (
			funcAID, funcBID     uuid.UUID
			funcAName, funcBName string
			coCount              int
			sharedCategory       string
		)

		if err := rows.Scan(&funcAID, &funcAName, &funcBID, &funcBName, &coCount, &sharedCategory); err != nil {
			continue
		}

		confidence := 0.70
		if coCount > 200 {
			confidence = 0.85
		}
		trajectory := TrajectoryStable

		insights = append(insights, &Insight{
			TenantID: tenantID,
			Category: CategoryRedundancy,
			Severity: SeverityOpportunity,
			Priority: 20,
			Title:    fmt.Sprintf("%s and %s are always called together", funcAName, funcBName),
			Message: fmt.Sprintf(
				"These two %s functions have been co-invoked %d times in the last 30 days. They may be doing overlapping work — consider consolidating.",
				sharedCategory, coCount,
			),
			Summary: strPtr(fmt.Sprintf("%d co-invocations — consolidation candidate", coCount)),
			InsightData: JSONMap{
				"function_a_id":       funcAID.String(),
				"function_a_name":     funcAName,
				"function_b_id":       funcBID.String(),
				"function_b_name":     funcBName,
				"co_occurrence_count": coCount,
				"shared_category":     sharedCategory,
				"signal":              "co_occurrence",
			},
			ActionType: ActionMergeFunctions,
			ActionData: JSONMap{
				"merge_from": funcBID.String(),
				"merge_into": funcAID.String(),
			},
			RelatedFunctionIDs: []uuid.UUID{funcAID, funcBID},
			Trajectory:         &trajectory,
			Confidence:         &confidence,
			Status:             StatusActive,
			ExpiresAt:          timePtr(time.Now().Add(14 * 24 * time.Hour)),
		})
	}

	return insights, rows.Err()
}

// ensure interface compliance
var _ Analyzer = (*RedundancyAnalyzer)(nil)

// isRelationNotExist checks for PostgreSQL "relation does not exist" errors.
func isRelationNotExist(err error) bool {
	if err == nil {
		return false
	}
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "42P01"
	}
	return false
}
