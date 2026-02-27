package devtools

import (
	"encoding/json"
	"strings"
)

// analyzeQueryPlan analyzes EXPLAIN output and provides suggestions
func (dt *DatabaseDevTools) analyzeQueryPlan(planJSON string) []string {
	var suggestions []string

	// Parse the JSON plan
	var plan interface{}
	if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
		dt.logger.WithError(err).Warn("Failed to parse query plan JSON")
		return []string{"Failed to parse query plan"}
	}

	// Enhanced analysis with better pattern matching
	planStr := strings.ToLower(planJSON)

	// Sequential scan detection
	if strings.Contains(planStr, "seq scan") {
		suggestions = append(suggestions, "Consider adding indexes for sequential scans on large tables")
	}

	// Nested loop join analysis
	if strings.Contains(planStr, "nested loop") {
		suggestions = append(suggestions, "Nested loop joins detected - review join conditions and consider additional indexes")
	}

	// Sort operation without index
	if strings.Contains(planStr, "\"sort\"") && !strings.Contains(planStr, "index") {
		suggestions = append(suggestions, "In-memory sorting detected - consider indexed ordering or increasing work_mem")
	}

	// High cost operations
	if strings.Contains(planStr, "\"total cost\":") {
		// Could be enhanced to parse actual cost values
		suggestions = append(suggestions, "Review query cost - high cost operations may benefit from optimization")
	}

	// Temporary file usage
	if strings.Contains(planStr, "temp file") {
		suggestions = append(suggestions, "Query uses temporary files - consider increasing work_mem or optimizing the query")
	}

	return suggestions
}