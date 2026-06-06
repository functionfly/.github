package studio

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// CodeEditorHandler handles code editing operations (format, save, undo, redo).
type CodeEditorHandler struct {
	repo          *CodeEditorRepository
	maxVersionAge time.Duration
}

// NewCodeEditorHandler creates a new CodeEditorHandler.
func NewCodeEditorHandler(repo *CodeEditorRepository) *CodeEditorHandler {
	return &CodeEditorHandler{
		repo:          repo,
		maxVersionAge: 30 * 24 * time.Hour, // Keep versions for 30 days
	}
}

type FormatRequest struct {
	Code     string                 `json:"code"`
	Language string                 `json:"language"`
	FilePath string                 `json:"file_path"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type FormatResponse struct {
	Formatted string `json:"formatted"`
	Version   int    `json:"version,omitempty"`
	Action    string `json:"action"`
}

// HandleFormatCode formats code and saves the formatted version.
func (h *CodeEditorHandler) HandleFormatCode(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	var req FormatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Code == "" {
		writeJSONError(w, http.StatusBadRequest, "Code is required")
		return
	}

	if req.Language == "" {
		req.Language = "typescript"
	}

	if req.FilePath == "" {
		req.FilePath = "main.ts"
	}

	// Format the code using basic formatting rules
	formatted, err := h.formatCode(req.Code, req.Language, req.Options)
	if err != nil {
		logrus.WithError(err).Warn("code format: formatting failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to format code")
		return
	}

	// Save formatted version to database
	metadata := map[string]interface{}{
		"language": req.Language,
		"original_length": len(req.Code),
		"formatted_length": len(formatted),
	}

	saved, err := h.repo.SaveVersion(r.Context(), tenantID, userID, environment, req.FilePath, formatted, "format", metadata)
	if err != nil {
		logrus.WithError(err).Error("code format: failed to save version")
		// Still return formatted code even if save fails
	}

	logrus.Infof("[CodeFormat] Formatted %s for tenant %s, file %s", req.Language, tenantID, req.FilePath)

	resp := FormatResponse{
		Formatted: formatted,
		Action:    "format",
	}
	if saved != nil {
		resp.Version = saved.Version
	}
	writeJSON(w, http.StatusOK, resp)
}

type SaveRequest struct {
	Code     string                 `json:"code"`
	FilePath string                 `json:"file_path"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type SaveResponse struct {
	Version   int       `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
}

// HandleSaveCode saves code and creates a version snapshot.
func (h *CodeEditorHandler) HandleSaveCode(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	var req SaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Code == "" {
		writeJSONError(w, http.StatusBadRequest, "Code is required")
		return
	}

	if req.FilePath == "" {
		req.FilePath = "main.ts"
	}

	// Add timestamp to metadata
	if req.Metadata == nil {
		req.Metadata = make(map[string]interface{})
	}
	req.Metadata["saved_at"] = time.Now().UTC().Format(time.RFC3339)

	saved, err := h.repo.SaveVersion(r.Context(), tenantID, userID, environment, req.FilePath, req.Code, "save", req.Metadata)
	if err != nil {
		logrus.WithError(err).Error("code save: failed to save version")
		writeJSONError(w, http.StatusInternalServerError, "Failed to save code")
		return
	}

	// Cleanup old versions (keep last 100)
	if err := h.repo.CleanupOldVersions(r.Context(), tenantID, userID, environment, req.FilePath, 100); err != nil {
		logrus.WithError(err).Warn("code save: failed to cleanup old versions")
	}

	logrus.Infof("[CodeSave] Saved for tenant %s, user %s, file %s, version %d",
		tenantID, userID, req.FilePath, saved.Version)

	writeJSON(w, http.StatusOK, SaveResponse{
		Version:   saved.Version,
		Timestamp: saved.CreatedAt,
		Action:    "save",
	})
}

type UndoRedoResponse struct {
	Code      string                 `json:"code"`
	Version   int                    `json:"version"`
	Action    string                 `json:"action"` // "undo" or "redo"
	Available bool                   `json:"available"`
	Metadata  map[string]interface{}  `json:"metadata,omitempty"`
}

// HandleUndo handles code undo operation.
func (h *CodeEditorHandler) HandleUndo(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	var req struct {
		FilePath       string `json:"file_path"`
		CurrentVersion int    `json:"current_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.FilePath == "" {
		req.FilePath = "main.ts"
	}

	// Get previous version
	prev, next, err := h.repo.GetUndoRedoVersions(r.Context(), tenantID, userID, environment, req.FilePath, req.CurrentVersion)
	if err != nil {
		logrus.WithError(err).Error("code undo: failed to get versions")
		writeJSONError(w, http.StatusInternalServerError, "Failed to perform undo")
		return
	}

	if prev == nil {
		writeJSON(w, http.StatusOK, UndoRedoResponse{
			Code:      "",
			Version:   req.CurrentVersion,
			Action:    "undo",
			Available: false,
		})
		return
	}

	// Save the undo action as a new version
	saved, err := h.repo.SaveVersion(r.Context(), tenantID, userID, environment, req.FilePath, prev.Content, "undo",
		map[string]interface{}{
			"restored_from_version": prev.Version,
			"original_version":       req.CurrentVersion,
		})
	if err != nil {
		logrus.WithError(err).Error("code undo: failed to save undo version")
	}

	logrus.Infof("[CodeUndo] Undone for tenant %s, user %s, file %s, version %d -> %d",
		tenantID, userID, req.FilePath, req.CurrentVersion, prev.Version)

	resp := UndoRedoResponse{
		Code:      prev.Content,
		Version:   prev.Version,
		Action:    "undo",
		Available: true,
	}
	if saved != nil {
		resp.Version = saved.Version
	}
	if next != nil {
		resp.Metadata = map[string]interface{}{"can_redo": true, "redo_version": next.Version}
	}
	writeJSON(w, http.StatusOK, resp)
}

// HandleRedo handles code redo operation.
func (h *CodeEditorHandler) HandleRedo(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	var req struct {
		FilePath       string `json:"file_path"`
		CurrentVersion int    `json:"current_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.FilePath == "" {
		req.FilePath = "main.ts"
	}

	// Get next version
	prev, next, err := h.repo.GetUndoRedoVersions(r.Context(), tenantID, userID, environment, req.FilePath, req.CurrentVersion)
	if err != nil {
		logrus.WithError(err).Error("code redo: failed to get versions")
		writeJSONError(w, http.StatusInternalServerError, "Failed to perform redo")
		return
	}

	if next == nil {
		writeJSON(w, http.StatusOK, UndoRedoResponse{
			Code:      "",
			Version:   req.CurrentVersion,
			Action:    "redo",
			Available: false,
		})
		return
	}

	// Save the redo action as a new version
	saved, err := h.repo.SaveVersion(r.Context(), tenantID, userID, environment, req.FilePath, next.Content, "redo",
		map[string]interface{}{
			"restored_from_version": next.Version,
			"original_version":       req.CurrentVersion,
		})
	if err != nil {
		logrus.WithError(err).Error("code redo: failed to save redo version")
	}

	logrus.Infof("[CodeRedo] Redone for tenant %s, user %s, file %s, version %d -> %d",
		tenantID, userID, req.FilePath, req.CurrentVersion, next.Version)

	resp := UndoRedoResponse{
		Code:      next.Content,
		Version:   next.Version,
		Action:    "redo",
		Available: true,
	}
	if saved != nil {
		resp.Version = saved.Version
	}
	if prev != nil {
		resp.Metadata = map[string]interface{}{"can_undo": true, "undo_version": prev.Version}
	}
	writeJSON(w, http.StatusOK, resp)
}

type VersionHistoryResponse struct {
	Versions []CodeVersionSummary `json:"versions"`
	Total    int                  `json:"total"`
}

type CodeVersionSummary struct {
	Version   int       `json:"version"`
	Action    string    `json:"action"`
	CreatedAt time.Time `json:"created_at"`
}

// HandleGetVersionHistory returns version history for a file.
func (h *CodeEditorHandler) HandleGetVersionHistory(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	environment := getEnvironment(r)

	filePath := r.URL.Query().Get("file_path")
	if filePath == "" {
		filePath = "main.ts"
	}

	limit := 50
	offset := 0

	versions, err := h.repo.ListVersions(r.Context(), tenantID, userID, environment, filePath, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("code history: failed to list versions")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get version history")
		return
	}

	summaries := make([]CodeVersionSummary, len(versions))
	for i, v := range versions {
		summaries[i] = CodeVersionSummary{
			Version:   v.Version,
			Action:    v.Action,
			CreatedAt: v.CreatedAt,
		}
	}

	writeJSON(w, http.StatusOK, VersionHistoryResponse{
		Versions: summaries,
		Total:     len(versions),
	})
}

// formatCode applies basic TypeScript/JavaScript formatting.
func (h *CodeEditorHandler) formatCode(code, language string, options map[string]interface{}) (string, error) {
	// Basic formatting based on language
	switch strings.ToLower(language) {
	case "typescript", "javascript", "tsx", "jsx":
		return h.formatTypeScript(code), nil
	case "python":
		return h.formatPython(code), nil
	default:
		return code, nil // Return as-is for unknown languages
	}
}

// formatTypeScript applies basic TypeScript formatting.
func (h *CodeEditorHandler) formatTypeScript(code string) string {
	lines := strings.Split(code, "\n")
	var formatted []string
	indentLevel := 0

	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")

		// Decrease indent for closing braces/brackets on their own line
		if strings.HasPrefix(trimmed, "}") || strings.HasPrefix(trimmed, "]") {
			if indentLevel > 0 {
				indentLevel--
			}
		}

		// Apply indentation
		if trimmed != "" {
			formatted = append(formatted, strings.Repeat("  ", indentLevel)+trimmed)
		} else {
			formatted = append(formatted, "")
		}

		// Increase indent after opening braces that aren't on the same line as closing
		if strings.HasSuffix(trimmed, "{") {
			indentLevel++
		}
		if strings.HasSuffix(trimmed, "[") && !strings.Contains(trimmed, "]") {
			indentLevel++
		}
	}

	// Remove trailing whitespace from each line
	for i, line := range formatted {
		formatted[i] = strings.TrimRight(line, " \t")
	}

	// Join lines, removing excessive blank lines (more than 2 consecutive)
	result := strings.Join(formatted, "\n")
	result = strings.ReplaceAll(result, "\n\n\n", "\n\n")

	return strings.Trim(result, "\n")
}

// formatPython applies basic Python formatting.
func (h *CodeEditorHandler) formatPython(code string) string {
	lines := strings.Split(code, "\n")
	var formatted []string
	indentLevel := 0

	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")

		// Apply indentation
		if trimmed != "" {
			formatted = append(formatted, strings.Repeat("    ", indentLevel)+trimmed)
		} else {
			formatted = append(formatted, "")
		}

		// Increase indent after lines ending with colon
		if strings.HasSuffix(trimmed, ":") {
			indentLevel++
		}
	}

	// Remove trailing whitespace from each line
	for i, line := range formatted {
		formatted[i] = strings.TrimRight(line, " \t")
	}

	result := strings.Join(formatted, "\n")
	result = strings.ReplaceAll(result, "\n\n\n", "\n\n")

	return strings.Trim(result, "\n")
}