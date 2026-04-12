package team_memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ============================================
// Bulk Import/Export for Team Memories
// ============================================

// BulkImportRequest represents a bulk import operation
type BulkImportRequest struct {
	Memories    []MemoryImportEntry `json:"memories"`
	TeamID      uuid.UUID           `json:"team_id"`
	TenantID    uuid.UUID           `json:"tenant_id"`
	ImportedBy  uuid.UUID           `json:"imported_by"`
	SkipInvalid bool                `json:"skip_invalid"` // If true, skip invalid entries; if false, fail entire batch
	Source      string              `json:"source"`       // e.g., "notion", "confluence", "csv", "manual"
}

// MemoryImportEntry represents a single memory to import
type MemoryImportEntry struct {
	MemoryType   string                 `json:"memory_type"`   // 'decision', 'preference', 'process', 'client_context'
	Category     *string                `json:"category"`      // e.g., "client:acme-corp"
	Summary      string                 `json:"summary"`       // Human-readable summary
	Content      map[string]interface{} `json:"content"`       // Structured content
	Confidence   float64                `json:"confidence"`    // 0.0-1.0
	TTLDays      int                    `json:"ttl_days"`      // 0 = never expire
	IsValidated  bool                   `json:"is_validated"`  // Pre-validated if true
	AutoValidate bool                   `json:"auto_validate"` // Auto-validate if confidence >= threshold
	ExternalID   *string                `json:"external_id"`   // ID from external system (for deduplication)
}

// BulkImportResult represents the result of a bulk import
type BulkImportResult struct {
	TotalSubmitted int           `json:"total_submitted"`
	Imported       int           `json:"imported"`
	Skipped        int           `json:"skipped"` // Deduplicated or invalid
	Failed         int           `json:"failed"`  // Failed to import
	Errors         []ImportError `json:"errors,omitempty"`
	ImportedIDs    []uuid.UUID   `json:"imported_ids"`
	DurationMs     int64         `json:"duration_ms"`
	SourceChecksum string        `json:"source_checksum"` // For idempotency
}

// ImportError represents an import error for a specific entry
type ImportError struct {
	Index   int    `json:"index"`
	EntryID string `json:"entry_id,omitempty"`
	Error   string `json:"error"`
}

// BulkImporter handles bulk import operations for team memories
type BulkImporter struct {
	repo             storage.Repository
	embeddingService storage.EmbeddingService
	duplicateChecker DuplicateChecker
}

// DuplicateChecker checks for duplicate memories during import
type DuplicateChecker interface {
	CheckDuplicate(ctx context.Context, teamID uuid.UUID, externalID string, content map[string]interface{}) (*storage.TeamMemory, error)
}

// NewBulkImporter creates a new bulk importer
func NewBulkImporter(repo storage.Repository, embeddingService storage.EmbeddingService) *BulkImporter {
	return &BulkImporter{
		repo:             repo,
		embeddingService: embeddingService,
	}
}

// Import imports memories in bulk
func (i *BulkImporter) Import(ctx context.Context, req BulkImportRequest) (*BulkImportResult, error) {
	start := time.Now()
	result := &BulkImportResult{
		TotalSubmitted: len(req.Memories),
		ImportedIDs:    make([]uuid.UUID, 0),
		Errors:         make([]ImportError, 0),
	}

	teamIDStr := req.TeamID.String()

	for idx, entry := range req.Memories {
		// Validate entry
		if err := i.validateEntry(&entry); err != nil {
			result.Skipped++
			if !req.SkipInvalid {
				result.Errors = append(result.Errors, ImportError{
					Index:   idx,
					EntryID: dereferenceString(entry.ExternalID),
					Error:   err.Error(),
				})
			}
			continue
		}

		// Check for duplicates if external ID provided
		if entry.ExternalID != nil && i.duplicateChecker != nil {
			duplicate, err := i.duplicateChecker.CheckDuplicate(ctx, req.TeamID, *entry.ExternalID, entry.Content)
			if err == nil && duplicate != nil {
				result.Skipped++
				continue // Skip duplicate
			}
		}

		// Create memory
		memory := &storage.TeamMemory{
			TenantID:        req.TenantID,
			TeamID:          req.TeamID,
			MemoryType:      entry.MemoryType,
			Category:        entry.Category,
			Content:         entry.Content,
			Summary:         &entry.Summary,
			CreatedBy:       req.ImportedBy,
			ConfidenceScore: entry.Confidence,
			IsValidated:     entry.IsValidated || (entry.AutoValidate && entry.Confidence >= 0.9),
			TTLDays:         entry.TTLDays,
		}

		// Generate embedding if service available
		if i.embeddingService != nil {
			searchText := entry.Summary + " " + extractTextFromContent(entry.Content, entry.MemoryType)
			embedding, err := i.embeddingService.GenerateEmbedding(ctx, searchText)
			if err == nil {
				memory.Embedding = embedding
			}
		}

		// Save to repository
		created, err := i.repo.CreateTeamMemory(ctx, memory)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				Index:   idx,
				EntryID: dereferenceString(entry.ExternalID),
				Error:   fmt.Sprintf("failed to create memory: %v", err),
			})
			continue
		}

		result.Imported++
		result.ImportedIDs = append(result.ImportedIDs, created.ID)

		// Record metric
		source := req.Source
		if source == "" {
			source = "bulk_import"
		}
		monitoring.RecordTeamMemoryCreated(teamIDStr, entry.MemoryType, source)
	}

	result.DurationMs = time.Since(start).Milliseconds()

	logrus.WithFields(logrus.Fields{
		"team_id":     req.TeamID,
		"total":       result.TotalSubmitted,
		"imported":    result.Imported,
		"skipped":     result.Skipped,
		"failed":      result.Failed,
		"duration_ms": result.DurationMs,
		"source":      req.Source,
	}).Info("Bulk memory import completed")

	return result, nil
}

// validateEntry validates a memory import entry
func (i *BulkImporter) validateEntry(entry *MemoryImportEntry) error {
	if entry.Summary == "" {
		return fmt.Errorf("summary is required")
	}

	validTypes := map[string]bool{
		"decision":       true,
		"preference":     true,
		"process":        true,
		"client_context": true,
	}

	if !validTypes[entry.MemoryType] {
		return fmt.Errorf("invalid memory_type: %s", entry.MemoryType)
	}

	if entry.Confidence < 0 || entry.Confidence > 1.0 {
		return fmt.Errorf("confidence must be between 0.0 and 1.0")
	}

	if entry.Content == nil {
		entry.Content = make(map[string]interface{})
	}

	return nil
}

// ============================================
// Bulk Export
// ============================================

// BulkExportRequest represents a bulk export operation
type BulkExportRequest struct {
	TeamID       uuid.UUID `json:"team_id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	MemoryTypes  []string  `json:"memory_types,omitempty"`
	Categories   []string  `json:"categories,omitempty"`
	IncludeAll   bool      `json:"include_all"`   // Include all, including unvalidated
	Format       string    `json:"format"`        // "json", "csv", "markdown"
	IncludeStats bool      `json:"include_stats"` // Include summary statistics
}

// BulkExportResult represents the result of a bulk export
type BulkExportResult struct {
	Memories     []MemoryExportEntry `json:"memories"`
	TotalCount   int                 `json:"total_count"`
	ExportFormat string              `json:"export_format"`
	ExportedAt   time.Time           `json:"exported_at"`
	ExportedBy   uuid.UUID           `json:"exported_by"`
	Stats        *ExportStats        `json:"stats,omitempty"`
	Checksum     string              `json:"checksum"`
	DurationMs   int64               `json:"duration_ms"`
}

// MemoryExportEntry represents a single exported memory
type MemoryExportEntry struct {
	ID             uuid.UUID              `json:"id"`
	MemoryType     string                 `json:"memory_type"`
	Category       *string                `json:"category,omitempty"`
	Summary        string                 `json:"summary"`
	Content        map[string]interface{} `json:"content"`
	Confidence     float64                `json:"confidence_score"`
	IsValidated    bool                   `json:"is_validated"`
	Importance     float64                `json:"importance_score"`
	AccessCount    int                    `json:"access_count"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	LastAccessedAt *time.Time             `json:"last_accessed_at,omitempty"`
	TTLDays        int                    `json:"ttl_days"`
}

// ExportStats represents statistics for the export
type ExportStats struct {
	ByType           map[string]int `json:"by_type"`
	ValidatedCount   int            `json:"validated_count"`
	UnvalidatedCount int            `json:"unvalidated_count"`
	AvgConfidence    float64        `json:"avg_confidence"`
}

// BulkExporter handles bulk export operations for team memories
type BulkExporter struct {
	repo storage.Repository
}

// NewBulkExporter creates a new bulk exporter
func NewBulkExporter(repo storage.Repository) *BulkExporter {
	return &BulkExporter{repo: repo}
}

// Export exports memories in bulk
func (e *BulkExporter) Export(ctx context.Context, req BulkExportRequest, exportedBy uuid.UUID) (*BulkExportResult, error) {
	start := time.Now()
	result := &BulkExportResult{
		Memories:     make([]MemoryExportEntry, 0),
		ExportFormat: req.Format,
		ExportedAt:   time.Now(),
		ExportedBy:   exportedBy,
	}

	// Build filter
	filter := storage.TeamMemoryFilter{
		Limit: 10000, // Max export size
	}

	if !req.IncludeAll {
		validated := true
		filter.IsValidated = &validated
	}

	// Fetch memories
	memories, total, err := e.repo.ListTeamMemories(ctx, req.TenantID, req.TeamID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list memories: %w", err)
	}

	// Filter by type and category if specified
	for _, memory := range memories {
		if len(req.MemoryTypes) > 0 && !containsString(req.MemoryTypes, memory.MemoryType) {
			continue
		}
		if len(req.Categories) > 0 && (memory.Category == nil || !containsString(req.Categories, *memory.Category)) {
			continue
		}

		entry := MemoryExportEntry{
			ID:             memory.ID,
			MemoryType:     memory.MemoryType,
			Category:       memory.Category,
			Summary:        dereferenceString(memory.Summary),
			Content:        memory.Content,
			Confidence:     memory.ConfidenceScore,
			IsValidated:    memory.IsValidated,
			Importance:     memory.ImportanceScore,
			AccessCount:    memory.AccessCount,
			CreatedAt:      memory.CreatedAt,
			UpdatedAt:      memory.UpdatedAt,
			LastAccessedAt: memory.LastAccessedAt,
			TTLDays:        memory.TTLDays,
		}

		result.Memories = append(result.Memories, entry)
	}

	result.TotalCount = len(result.Memories)
	result.DurationMs = time.Since(start).Milliseconds()

	// Calculate stats if requested
	if req.IncludeStats {
		result.Stats = e.calculateStats(result.Memories)
	}

	// Calculate checksum
	data, _ := json.Marshal(result.Memories)
	result.Checksum = fmt.Sprintf("%x", data)[:16] // Simple truncated hash

	logrus.WithFields(logrus.Fields{
		"team_id":       req.TeamID,
		"total":         result.TotalCount,
		"filtered_from": total,
		"duration_ms":   result.DurationMs,
		"format":        req.Format,
	}).Info("Bulk memory export completed")

	return result, nil
}

// calculateStats calculates export statistics
func (e *BulkExporter) calculateStats(entries []MemoryExportEntry) *ExportStats {
	stats := &ExportStats{
		ByType: make(map[string]int),
	}

	var totalConfidence float64
	for _, entry := range entries {
		stats.ByType[entry.MemoryType]++
		totalConfidence += entry.Confidence

		if entry.IsValidated {
			stats.ValidatedCount++
		} else {
			stats.UnvalidatedCount++
		}
	}

	if len(entries) > 0 {
		stats.AvgConfidence = totalConfidence / float64(len(entries))
	}

	return stats
}

// ExportAsMarkdown exports memories as formatted Markdown
func (e *BulkExporter) ExportAsMarkdown(ctx context.Context, req BulkExportRequest, exportedBy uuid.UUID) (string, error) {
	result, err := e.Export(ctx, req, exportedBy)
	if err != nil {
		return "", err
	}

	var md string
	md += fmt.Sprintf("# Team Memory Export\n\n")
	md += fmt.Sprintf("**Team ID:** %s\n\n", req.TeamID)
	md += fmt.Sprintf("**Exported At:** %s\n\n", result.ExportedAt.Format("2006-01-02 15:04:05"))
	md += fmt.Sprintf("**Total Memories:** %d\n\n", result.TotalCount)

	if result.Stats != nil {
		md += "## Statistics\n\n"
		md += fmt.Sprintf("- Validated: %d\n", result.Stats.ValidatedCount)
		md += fmt.Sprintf("- Unvalidated: %d\n", result.Stats.UnvalidatedCount)
		md += fmt.Sprintf("- Average Confidence: %.1f%%\n", result.Stats.AvgConfidence*100)
		md += "\n### By Type\n\n"
		for t, c := range result.Stats.ByType {
			md += fmt.Sprintf("- %s: %d\n", t, c)
		}
		md += "\n"
	}

	md += "## Memories\n\n"

	for i, entry := range result.Memories {
		md += fmt.Sprintf("### %d. %s\n\n", i+1, entry.Summary)
		md += fmt.Sprintf("- **Type:** %s\n", entry.MemoryType)
		md += fmt.Sprintf("- **ID:** %s\n", entry.ID)
		if entry.Category != nil {
			md += fmt.Sprintf("- **Category:** %s\n", *entry.Category)
		}
		md += fmt.Sprintf("- **Confidence:** %.0f%%\n", entry.Confidence*100)
		md += fmt.Sprintf("- **Validated:** %v\n", entry.IsValidated)
		md += fmt.Sprintf("- **Created:** %s\n", entry.CreatedAt.Format("2006-01-02"))
		md += "\n**Content:**\n\n```json\n"
		contentJSON, _ := json.MarshalIndent(entry.Content, "", "  ")
		md += string(contentJSON)
		md += "\n```\n\n---\n\n"
	}

	return md, nil
}

// Helper functions

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// extractTextFromContent extracts searchable text from structured content
func extractTextFromContent(content map[string]interface{}, memoryType string) string {
	switch memoryType {
	case "decision":
		if title, ok := content["title"].(string); ok {
			return title
		}
		if rationale, ok := content["rationale"].(string); ok {
			return rationale
		}
	case "preference":
		if subject, ok := content["subject"].(string); ok {
			if value, ok := content["value"].(string); ok {
				return subject + " " + value
			}
		}
	case "process":
		if name, ok := content["name"].(string); ok {
			return name
		}
	case "client_context":
		if clientName, ok := content["client_name"].(string); ok {
			return clientName
		}
	}

	// Fallback: convert entire content to string representation
	if content != nil {
		data, _ := json.Marshal(content)
		return string(data)
	}
	return ""
}
