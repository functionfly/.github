package registry

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// humanizeFunctionName turns a function name like "text-truncate" into "Text truncate".
func humanizeFunctionName(name string) string {
	if name == "" {
		return "No description available"
	}
	parts := strings.Split(strings.ReplaceAll(name, "_", "-"), "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		}
	}
	return strings.Join(parts, " ")
}

// Helper function to parse tags from JSON
func ParseTags(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return []string{}, nil
	}
	var tags []string
	err := json.Unmarshal(raw, &tags)
	return tags, err
}

// Helper function to convert function to info
func (f *RegistryFunction) ToInfo(version *RegistryFunctionVersion) map[string]interface{} {
	return f.ToInfoWithRating(version, nil)
}

// ToInfoWithRating converts function to info with optional rating data
func (f *RegistryFunction) ToInfoWithRating(version *RegistryFunctionVersion, rating *RegistryFunctionRating) map[string]interface{} {
	versionStr := ""
	if version != nil {
		versionStr = version.Version
	}
	info := map[string]interface{}{
		"author":         f.Author,
		"name":           f.Name,
		"version":        versionStr,
		"visibility":     f.Visibility,
		"price_per_call": f.PricePerCall,
		"reliability":    f.ReliabilityScore,
	}

	if f.Title.Valid {
		info["title"] = f.Title.String
	}
	if f.Description.Valid {
		info["description"] = f.Description.String
	}
	if f.Category.Valid {
		info["category"] = f.Category.String
	}

	tags, _ := ParseTags(f.Tags)
	info["tags"] = tags

	info["created_at"] = f.CreatedAt.Format(time.RFC3339)
	info["updated_at"] = f.UpdatedAt.Format(time.RFC3339)

	if version != nil {
		info["runtime"] = version.Runtime
		info["timeout_ms"] = version.TimeoutMs
		info["memory_mb"] = version.MemoryMB
		info["deterministic"] = version.Deterministic

		// Add documentation URLs
		info["documentation_url"] = fmt.Sprintf("/docs/%s/%s", f.Author, f.Name)
		info["playground_url"] = fmt.Sprintf("/playground/%s/%s", f.Author, f.Name)
		info["side_effects"] = version.SideEffects
		info["idempotent"] = version.Idempotent
		info["cache_ttl"] = version.CacheTTL

		// Parse and add capabilities from version
		var capabilities []string
		if version.Capabilities != nil && len(version.Capabilities) > 0 {
			json.Unmarshal(version.Capabilities, &capabilities)
		}
		if len(capabilities) > 0 {
			info["capabilities"] = capabilities
		}

		// Add bundle size if present
		if version.BundleSize.Valid {
			info["bundle_size"] = version.BundleSize.Int32
		}

		// Add source hash if present
		if version.SourceHash.Valid {
			info["source_hash"] = version.SourceHash.String
		}

		// Add deployment ID if present
		if version.DeploymentID != nil {
			info["deployment_id"] = version.DeploymentID.String()
		}

		// Add backend ID if present
		if version.BackendID != nil {
			info["backend_id"] = version.BackendID.String()
		}

		// Parse manifest for input/output examples and to backfill description/category
		var manifest map[string]interface{}
		json.Unmarshal(version.Manifest, &manifest)
		if input, ok := manifest["input"].(map[string]interface{}); ok {
			info["input_type"] = input["type"]
			info["input_example"] = input["example"]
		}
		if output, ok := manifest["output"].(map[string]interface{}); ok {
			info["output_type"] = output["type"]
			info["output_example"] = output["example"]
		}
		// Backfill description from manifest when DB has none
		if !f.Description.Valid {
			if desc, ok := manifest["description"].(string); ok && desc != "" {
				info["description"] = desc
			}
		}
		// Backfill category from manifest when DB has none
		if !f.Category.Valid {
			if cat, ok := manifest["category"].(string); ok && cat != "" {
				info["category"] = cat
			}
		}
		// Fallback description when none in DB or manifest (e.g. batch-published functions)
		if desc, _ := info["description"].(string); desc == "" {
			info["description"] = humanizeFunctionName(f.Name)
		}
		info["published_at"] = version.PublishedAt.Format(time.RFC3339)
	}

	// Add trust score from rating if available (DB stores 0-1; API returns 0-100 for frontend)
	if rating != nil {
		trustScore := rating.TrustScore
		if trustScore > 0 && trustScore <= 1 {
			trustScore = trustScore * 100
		}
		info["trust_score"] = trustScore
		info["success_rate"] = rating.SuccessRate
		info["p50_latency_ms"] = rating.P50LatencyMs
		info["p95_latency_ms"] = rating.P95LatencyMs
		info["timeout_rate"] = rating.TimeoutRate
		info["error_rate"] = rating.ErrorRate
		info["consumer_diversity"] = rating.ConsumerDiversity
		info["tenant_diversity"] = rating.TenantDiversity
		info["user_diversity"] = rating.UserDiversity

		// Determine trust level from scaled score (0-100)
		trustLevel := "medium"
		switch {
		case trustScore >= 80:
			trustLevel = "high"
		case trustScore >= 60:
			trustLevel = "high"
		case trustScore >= 40:
			trustLevel = "medium"
		case trustScore >= 20:
			trustLevel = "low"
		case trustScore > 0:
			trustLevel = "untrusted"
		}
		info["trust_level"] = trustLevel
	}

	return info
}
