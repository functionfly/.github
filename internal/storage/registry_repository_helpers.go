package storage

import (
	"encoding/json"
)

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
func ToInfo(f *RegistryFunction, version *RegistryFunctionVersion) map[string]interface{} {
	return ToInfoWithRating(f, version, nil)
}

// ToInfoWithRating converts function to info with optional rating data
func ToInfoWithRating(f *RegistryFunction, version *RegistryFunctionVersion, rating *RegistryFunctionRating) map[string]interface{} {
	info := map[string]interface{}{
		"author":         f.Author,
		"name":           f.Name,
		"version":        version.Version,
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

	if version != nil {
		info["runtime"] = version.Runtime
		info["timeout_ms"] = version.TimeoutMs
		info["memory_mb"] = version.MemoryMB
		info["deterministic"] = version.Deterministic
		info["side_effects"] = version.SideEffects
		info["idempotent"] = version.Idempotent
		info["cache_ttl"] = version.CacheTTL

		// Parse manifest for input/output examples
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
		if examples, ok := manifest["examples"].([]interface{}); ok {
			info["examples"] = examples
		}
	}

	// Add trust score from rating if available
	if rating != nil {
		info["trust_score"] = rating.TrustScore
		info["success_rate"] = rating.SuccessRate
		info["p50_latency_ms"] = rating.P50LatencyMs
		info["p95_latency_ms"] = rating.P95LatencyMs
		info["timeout_rate"] = rating.TimeoutRate
		info["error_rate"] = rating.ErrorRate
		info["consumer_diversity"] = rating.ConsumerDiversity
		info["tenant_diversity"] = rating.TenantDiversity
		info["user_diversity"] = rating.UserDiversity

		// Determine trust level
		trustLevel := "insufficient_data"
		switch {
		case rating.TrustScore >= 80:
			trustLevel = "excellent"
		case rating.TrustScore >= 60:
			trustLevel = "good"
		case rating.TrustScore >= 40:
			trustLevel = "fair"
		case rating.TrustScore >= 20:
			trustLevel = "poor"
		case rating.TrustScore > 0:
			trustLevel = "very_poor"
		}
		info["trust_level"] = trustLevel
	}

	return info
}