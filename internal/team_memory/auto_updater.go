package team_memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/conversations"
	"github.com/functionfly/functionfly/internal/monitoring"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AIClientConfig holds configuration for the AI service client
type AIClientConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
	Enabled bool
}

// DefaultAIClientConfig returns default configuration from environment
func DefaultAIClientConfig() *AIClientConfig {
	baseURL := os.Getenv("AI_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081" // Default ai-service port
	}

	apiKey := os.Getenv("AI_SERVICE_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("FLYMIND_API_KEY") // Fallback
	}

	timeout := 30 * time.Second
	if t := os.Getenv("AI_SERVICE_TIMEOUT"); t != "" {
		if d, err := time.ParseDuration(t); err == nil {
			timeout = d
		}
	}

	return &AIClientConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Timeout: timeout,
		Enabled: os.Getenv("AI_SERVICE_ENABLED") != "false",
	}
}

// AIServiceClient implements AIAnalyzer by calling the FlyMind AI service
type AIServiceClient struct {
	config AIClientConfig
	client *http.Client
}

// MemoryExtractionRequest represents the request to ai-service
type MemoryExtractionRequest struct {
	Transcript     string                 `json:"transcript"`
	TeamID         string                 `json:"team_id,omitempty"`
	ConversationID string                 `json:"conversation_id,omitempty"`
	Context        map[string]interface{} `json:"context,omitempty"`
}

// MemoryExtractionResponse represents the response from ai-service
type MemoryExtractionResponse struct {
	Memories   []ExtractedMemory `json:"memories"`
	Confidence float64           `json:"confidence"`
	TokensUsed int               `json:"tokens_used"`
	Model      string            `json:"model"`
	LatencyMs  float64           `json:"latency_ms"`
}

// ExtractedMemory represents a single extracted memory from ai-service
type ExtractedMemory struct {
	Type       string                 `json:"type"`
	Category   *string                `json:"category,omitempty"`
	Summary    string                 `json:"summary"`
	Content    map[string]interface{} `json:"content"`
	Confidence float64                `json:"confidence"`
	Rationale  string                 `json:"rationale"`
}

// NewAIServiceClient creates a new AI service client for memory extraction
func NewAIServiceClient(config *AIClientConfig) *AIServiceClient {
	if config == nil {
		config = DefaultAIClientConfig()
	}

	return &AIServiceClient{
		config: *config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// IsEnabled returns whether the AI service is enabled
func (c *AIServiceClient) IsEnabled() bool {
	return c.config.Enabled
}

// AnalyzeConversation sends transcript to ai-service for memory extraction
func (c *AIServiceClient) AnalyzeConversation(ctx context.Context, transcript string) (*ExtractionResult, error) {
	if !c.config.Enabled {
		return nil, fmt.Errorf("AI service is disabled")
	}

	reqBody := MemoryExtractionRequest{
		Transcript: transcript,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/memory/extract", c.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("X-API-Key", c.config.APIKey)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		if isNetworkError(err) {
			logrus.WithError(err).Warn("AI service unreachable, using fallback")
			return nil, fmt.Errorf("ai service unreachable: %w", err)
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("AI service authentication failed")
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, fmt.Errorf("AI service rate limited")
		}
		return nil, fmt.Errorf("AI service returned status %d", resp.StatusCode)
	}

	var result MemoryExtractionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to internal format
	memories := make([]PotentialMemory, 0, len(result.Memories))
	for _, m := range result.Memories {
		memories = append(memories, PotentialMemory{
			Type:       m.Type,
			Category:   dereferenceString(m.Category),
			Summary:    m.Summary,
			Content:    m.Content,
			Confidence: m.Confidence,
			Rationale:  m.Rationale,
		})
	}

	return &ExtractionResult{
		Memories:   memories,
		Confidence: result.Confidence,
	}, nil
}

// isNetworkError checks if the error is a network connectivity issue
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "i/o timeout") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "dial tcp")
}

// dereferenceString safely dereferences a string pointer
func dereferenceString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ============================================
// Auto Updater with AI Service Integration
// ============================================

// AutoUpdater automatically extracts team memories from conversations
type AutoUpdater struct {
	repo     storage.Repository
	convRepo *storage.ConversationRepository
	aiClient AIAnalyzer
	enabled  bool
	fallback AIAnalyzer // Fallback when AI service is unavailable
}

// AIAnalyzer interface for analyzing conversations and extracting memories
type AIAnalyzer interface {
	AnalyzeConversation(ctx context.Context, transcript string) (*ExtractionResult, error)
}

// ExtractionResult contains AI-extracted potential memories
type ExtractionResult struct {
	Memories   []PotentialMemory `json:"memories"`
	Confidence float64           `json:"confidence"`
}

// PotentialMemory represents a memory extracted by AI
type PotentialMemory struct {
	Type       string                 `json:"type"`
	Category   string                 `json:"category,omitempty"`
	Summary    string                 `json:"summary"`
	Content    map[string]interface{} `json:"content"`
	Confidence float64                `json:"confidence"`
	Rationale  string                 `json:"rationale"`
}

// NewAutoUpdater creates a new auto-updater with AI service integration
// Uses GPT-4o-mini via FlyMind AI service for extraction (2026 pricing).
// Cost per conversation: ~$0.0006 (2K input + 0.5K output tokens at ~$0.15/1M input, ~$0.60/1M output)
// Monthly cost: ~$1.50 for 2,500 conversations
func NewAutoUpdater(
	repo storage.Repository,
	convRepo *storage.ConversationRepository,
	aiClient AIAnalyzer,
) *AutoUpdater {
	// Create AI service client if not provided
	if aiClient == nil {
		config := DefaultAIClientConfig()
		aiClient = NewAIServiceClient(config)
	}

	return &AutoUpdater{
		repo:     repo,
		convRepo: convRepo,
		aiClient: aiClient,
		enabled:  true,
		fallback: NewSimpleMemoryExtractor(), // Fallback extractor (zero cost)
	}
}

// SetEnabled enables or disables the auto-updater
func (u *AutoUpdater) SetEnabled(enabled bool) {
	u.enabled = enabled
}

// SetFallback sets a fallback analyzer for when AI service is unavailable
func (u *AutoUpdater) SetFallback(fallback AIAnalyzer) {
	u.fallback = fallback
}

// ProcessConversation analyzes a completed conversation for memory-worthy content
func (u *AutoUpdater) ProcessConversation(ctx context.Context, conversationID uuid.UUID, teamID uuid.UUID) error {
	if !u.enabled {
		return nil
	}

	start := time.Now()
	teamIDStr := teamID.String()

	// Fetch conversation using conversation repository
	conversation, err := u.convRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		monitoring.RecordTeamMemoryExtraction(teamIDStr, "unknown", "failed")
		return fmt.Errorf("failed to fetch conversation: %w", err)
	}
	if conversation == nil {
		monitoring.RecordTeamMemoryExtraction(teamIDStr, "unknown", "failed")
		return fmt.Errorf("conversation not found: %s", conversationID)
	}

	// Fetch messages for the conversation
	messages, err := u.convRepo.ListMessages(ctx, conversationID, 1000, 0)
	if err != nil {
		monitoring.RecordTeamMemoryExtraction(teamIDStr, "unknown", "failed")
		return fmt.Errorf("failed to fetch messages: %w", err)
	}

	// Build transcript from messages
	transcript := u.buildTranscriptFromMessages(messages)
	if len(transcript) < 100 {
		return nil // Too short, skip
	}

	// Try AI service first, fallback to rule-based if unavailable
	var result *ExtractionResult
	var extractionSource string

	if aiClient, ok := u.aiClient.(*AIServiceClient); ok && aiClient.IsEnabled() {
		result, err = u.aiClient.AnalyzeConversation(ctx, transcript)
		if err != nil {
			logrus.WithError(err).Warn("AI service failed, using fallback extractor")
			extractionSource = "fallback"
			result, err = u.fallback.AnalyzeConversation(ctx, transcript)
		} else {
			extractionSource = "ai_service"
		}
	} else {
		// Use fallback directly
		extractionSource = "fallback"
		result, err = u.fallback.AnalyzeConversation(ctx, transcript)
	}

	if err != nil {
		logrus.WithError(err).Error("Failed to analyze conversation for memories")
		monitoring.RecordTeamMemoryExtraction(teamIDStr, "unknown", "failed")
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Record extraction duration
	monitoring.RecordTeamMemoryExtractionDuration(teamIDStr, extractionSource, time.Since(start))

	// Process each potential memory
	extractionCount := 0
	for _, potential := range result.Memories {
		if potential.Confidence < 0.7 {
			continue // Skip low-confidence extractions
		}

		// Record extraction confidence metric
		monitoring.RecordTeamMemoryExtractionConfidence(teamIDStr, potential.Type, potential.Confidence)

		// Check for similar existing memory
		duplicate, err := u.findSimilarMemory(ctx, teamID, potential)
		if err != nil {
			logrus.WithError(err).Warn("Failed to check for duplicate memory")
		}

		if duplicate != nil {
			// Merge/update existing if confidence higher
			if potential.Confidence > duplicate.ConfidenceScore {
				u.mergeMemoryUpdate(ctx, duplicate, potential)
			}
			continue
		}

		// Create extraction record
		extraction := &storage.MemoryExtraction{
			TeamID:         teamID,
			ConversationID: conversationID,
			MemoryType:     potential.Type,
			Category:       &potential.Category,
			Content:        potential.Content,
			Summary:        potential.Summary,
			Confidence:     potential.Confidence,
			Rationale:      potential.Rationale,
			Status:         "pending",
		}

		_, err = u.repo.CreateMemoryExtraction(ctx, extraction)
		if err != nil {
			logrus.WithError(err).Error("Failed to create memory extraction")
			monitoring.RecordTeamMemoryExtraction(teamIDStr, potential.Type, "failed")
			continue
		}

		monitoring.RecordTeamMemoryExtraction(teamIDStr, potential.Type, "success")
		extractionCount++

		// Auto-apply if high confidence (MVP default: >= 0.9)
		if potential.Confidence >= 0.9 {
			memory, err := u.repo.ApproveMemoryExtraction(ctx, extraction.ID, uuid.Nil) // System approval
			if err != nil {
				logrus.WithError(err).Error("Failed to auto-approve extraction")
			} else {
				monitoring.RecordTeamMemoryCreated(teamIDStr, memory.MemoryType, "auto_extraction")
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"conversation_id":   conversationID,
		"team_id":           teamID,
		"extractions":       extractionCount,
		"extraction_source": extractionSource,
		"duration_ms":       time.Since(start).Milliseconds(),
	}).Info("Processed conversation for memory extraction")

	return nil
}

// buildTranscriptFromMessages creates a text transcript from conversation messages
func (u *AutoUpdater) buildTranscriptFromMessages(messages []*conversations.ConversationMessage) string {
	var parts []string

	for _, msg := range messages {
		if msg == nil {
			continue
		}
		// Extract role from message or determine from author
		role := "unknown"
		if msg.AuthorID != uuid.Nil {
			// Could look up user to get role, but for now use author ID prefix
			role = msg.AuthorID.String()[:8]
		}

		timestamp := msg.CreatedAt.Format("15:04")
		parts = append(parts, fmt.Sprintf("[%s] %s: %s", timestamp, role, msg.Content))
	}

	return strings.Join(parts, "\n")
}

// findSimilarMemory checks for semantically similar existing memories
func (u *AutoUpdater) findSimilarMemory(ctx context.Context, teamID uuid.UUID, potential PotentialMemory) (*storage.TeamMemory, error) {
	// Search by summary text
	searchText := potential.Summary
	if len(searchText) < 10 {
		return nil, nil
	}

	// Extract tenant ID from context for proper isolation
	tenantID := GetTenantIDFromContext(ctx)

	// Use the team memory repository for the search
	// Use empty embedding for now (would need to generate one)
	results, err := u.repo.SearchTeamMemoriesByVector(ctx, tenantID, teamID, nil, 5)
	if err != nil {
		return nil, err
	}

	// Check for high similarity (relevance score > 0.85)
	for _, result := range results {
		if result.RelevanceScore > 0.85 {
			return &result.TeamMemory, nil
		}
	}

	return nil, nil
}

// mergeMemoryUpdate merges new extraction into existing memory
func (u *AutoUpdater) mergeMemoryUpdate(ctx context.Context, existing *storage.TeamMemory, potential PotentialMemory) {
	// Only update if extraction has higher confidence
	if potential.Confidence <= existing.ConfidenceScore {
		return
	}

	// Update content and confidence
	existing.Content = potential.Content
	existing.ConfidenceScore = potential.Confidence
	if existing.Summary != nil && potential.Summary != "" {
		summary := potential.Summary
		existing.Summary = &summary
	}

	_, err := u.repo.UpdateTeamMemory(ctx, existing)
	if err != nil {
		logrus.WithError(err).Error("Failed to merge memory update")
	}
}

// StartBackgroundProcessing starts a background worker for auto-processing
func (u *AutoUpdater) StartBackgroundProcessing(ctx context.Context, interval time.Duration) {
	if interval == 0 {
		interval = 5 * time.Minute
	}

	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				u.processPendingExtractions(ctx)
			}
		}
	}()
}

// processPendingExtractions processes extractions in the queue
func (u *AutoUpdater) processPendingExtractions(ctx context.Context) {
	count, err := u.repo.ProcessAutoApplyExtractions(ctx, 10)
	if err != nil {
		logrus.WithError(err).Error("Failed to process auto-apply extractions")
		return
	}
	if count > 0 {
		logrus.WithField("count", count).Info("Auto-applied memory extractions")
	}
}

// SimpleMemoryExtractor provides a basic rule-based extractor without AI
// Used as fallback when AI service is unavailable
type SimpleMemoryExtractor struct{}

// NewSimpleMemoryExtractor creates a simple extractor
func NewSimpleMemoryExtractor() *SimpleMemoryExtractor {
	return &SimpleMemoryExtractor{}
}

// AnalyzeConversation extracts memories using simple pattern matching
func (e *SimpleMemoryExtractor) AnalyzeConversation(ctx context.Context, transcript string) (*ExtractionResult, error) {
	var memories []PotentialMemory
	lines := strings.Split(transcript, "\n")

	for _, line := range lines {
		line = strings.ToLower(line)

		// Decision patterns
		if containsAny(line, "we decided", "we agree", "decision:", "let's go with", "let's use") {
			memories = append(memories, PotentialMemory{
				Type:       "decision",
				Summary:    extractSummary(line, 100),
				Content:    map[string]interface{}{"decision": extractContent(line)},
				Confidence: 0.85,
				Rationale:  "Explicit decision statement found in conversation",
			})
		}

		// Preference patterns
		if containsAny(line, "prefers", "likes", "wants", "always use", "never use") {
			memories = append(memories, PotentialMemory{
				Type:       "preference",
				Summary:    extractSummary(line, 100),
				Content:    map[string]interface{}{"preference": extractContent(line)},
				Confidence: 0.8,
				Rationale:  "Preference statement found",
			})
		}

		// Process patterns
		if containsAny(line, "process", "workflow", "steps:", "first we", "then we") {
			memories = append(memories, PotentialMemory{
				Type:       "process",
				Summary:    extractSummary(line, 100),
				Content:    map[string]interface{}{"process": extractContent(line)},
				Confidence: 0.75,
				Rationale:  "Process description found",
			})
		}

		// Client context patterns
		if containsAny(line, "client", "customer", "acme", "corp") {
			memories = append(memories, PotentialMemory{
				Type:       "client_context",
				Summary:    extractSummary(line, 100),
				Content:    map[string]interface{}{"context": extractContent(line)},
				Confidence: 0.7,
				Rationale:  "Client reference found",
			})
		}
	}

	return &ExtractionResult{
		Memories:   memories,
		Confidence: 0.75,
	}, nil
}

// Helper functions
func containsAny(s string, substrs ...string) bool {
	lower := strings.ToLower(s)
	for _, substr := range substrs {
		if strings.Contains(lower, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}

func extractSummary(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

func extractContent(s string) string {
	return strings.TrimSpace(s)
}
