package registry

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// GalleryStats holds aggregate registry metrics for the public gallery HUD.
type GalleryStats struct {
	TotalFunctions   int64 `json:"total_functions"`
	TotalRemixes     int64 `json:"total_remixes"`
	DistinctRuntimes int64 `json:"distinct_runtimes"`
}

// normalizeGalleryRuntime maps backend runtime strings to gallery palette keys
// (mirrors web/dashboard/src/pages/GalleryPage/constants.ts).
func normalizeGalleryRuntime(runtime string) string {
	r := strings.ToLower(strings.TrimSpace(runtime))
	if r == "" {
		return ""
	}
	if strings.HasPrefix(r, "python") {
		return "python"
	}
	if strings.HasPrefix(r, "node") || r == "javascript" {
		return "nodejs"
	}
	if strings.HasPrefix(r, "typescript") || r == "deno" || r == "bun" {
		return "typescript"
	}
	if strings.HasPrefix(r, "go") {
		return "go"
	}
	if strings.HasPrefix(r, "rust") {
		return "rust"
	}
	if strings.HasPrefix(r, "java") {
		return "java"
	}
	if strings.HasPrefix(r, "csharp") || strings.HasPrefix(r, "c#") {
		return "csharp"
	}
	if strings.HasPrefix(r, "ruby") {
		return "ruby"
	}
	if strings.HasPrefix(r, "php") {
		return "php"
	}
	parts := strings.FieldsFunc(r, func(c rune) bool {
		return (c < 'a' || c > 'z') && (c < '0' || c > '9')
	})
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return "python"
}

// GetGalleryStats returns aggregate public-registry metrics for the gallery hero.
func (r *RegistryRepository) GetGalleryStats() (*GalleryStats, error) {
	if r.cache != nil {
		cacheKey := r.keyGen.GalleryStats()
		var cached GalleryStats
		if err := r.cache.GetJSON(context.Background(), cacheKey, &cached); err == nil {
			return &cached, nil
		}
	}

	stats := &GalleryStats{}

	if err := r.db.Model(&RegistryFunction{}).
		Where("visibility = ?", "public").
		Count(&stats.TotalFunctions).Error; err != nil {
		return nil, fmt.Errorf("failed to count public functions: %w", err)
	}

	if err := r.db.Model(&RemixHistory{}).Count(&stats.TotalRemixes).Error; err != nil {
		return nil, fmt.Errorf("failed to count remixes: %w", err)
	}

	var runtimes []string
	// Join on denormalized latest_version instead of DISTINCT ON over all version rows.
	if err := r.db.Raw(`
		SELECT DISTINCT v.runtime
		FROM registry_functions f
		INNER JOIN registry_function_versions v
			ON v.function_id = f.id AND v.version = f.latest_version
		WHERE f.visibility = 'public'
			AND f.latest_version IS NOT NULL
			AND f.latest_version <> ''
			AND v.runtime <> ''
	`).Scan(&runtimes).Error; err != nil {
		return nil, fmt.Errorf("failed to load public function runtimes: %w", err)
	}

	seen := make(map[string]struct{}, len(runtimes))
	for _, rt := range runtimes {
		normalized := normalizeGalleryRuntime(rt)
		if normalized == "" {
			continue
		}
		seen[normalized] = struct{}{}
	}
	stats.DistinctRuntimes = int64(len(seen))

	if r.cache != nil {
		cacheKey := r.keyGen.GalleryStats()
		if err := r.cache.SetJSONWithTTL(context.Background(), cacheKey, stats, 5*time.Minute); err != nil {
			fmt.Printf("Failed to cache gallery stats: %v\n", err)
		}
	}

	return stats, nil
}
