package employee

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleSearchMemory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		apierror.WriteError(w, apierror.NewBadRequest("q (query) parameter is required"))
		return
	}

	opts := storage.SearchLivingMemoryOpts{
		Query: query,
		Limit: 20,
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 50 {
			opts.Limit = n
		}
	}
	if mt := r.URL.Query().Get("memory_type"); mt != "" {
		opts.MemoryType = &mt
	}
	if pid := r.URL.Query().Get("project_id"); pid != "" {
		if id, err := uuid.Parse(pid); err == nil {
			opts.ProjectID = &id
		}
	}

	entries, err := h.repo.SearchLivingMemory(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to search living memory")
		apierror.WriteError(w, apierror.NewInternal("Failed to search memory"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": entries,
		"total":   len(entries),
		"query":   query,
	})
}

type createMemoryRequest struct {
	Title        string   `json:"title"`
	Body         string   `json:"body"`
	MemoryType   string   `json:"memory_type"`
	ProjectID    *string  `json:"project_id,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Participants []string `json:"participants,omitempty"`
	Importance   string   `json:"importance,omitempty"`
}

func (h *Handler) HandleCreateMemory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	emp, err := h.repo.GetEmployeeByUserID(r.Context(), claims.UserID)
	if err != nil || emp == nil {
		apierror.WriteError(w, apierror.NewNotFound("Employee profile not found"))
		return
	}

	var req createMemoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Title == "" || req.Body == "" {
		apierror.WriteError(w, apierror.NewBadRequest("title and body are required"))
		return
	}

	validTypes := map[string]bool{
		"meeting": true, "decision": true, "design": true,
		"lesson": true, "discovery": true, "note": true,
	}
	memoryType := req.MemoryType
	if memoryType == "" {
		memoryType = "note"
	}
	if !validTypes[memoryType] {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid memory_type. Must be one of: meeting, decision, design, lesson, discovery, note"))
		return
	}

	entry := &storage.LivingMemoryEntry{
		TenantID:   claims.TenantID,
		AuthorID:   emp.ID,
		Title:      req.Title,
		Body:       req.Body,
		MemoryType: memoryType,
		Importance: "normal",
	}

	if req.Importance != "" {
		entry.Importance = req.Importance
	}
	if req.ProjectID != nil {
		pid, err := uuid.Parse(*req.ProjectID)
		if err == nil {
			entry.ProjectID = &pid
		}
	}
	if req.Tags != nil {
		tags := make(storage.JSONMap)
		for i, t := range req.Tags {
			tags[string(rune('0'+i))] = t
		}
		entry.Tags = tags
	}
	if req.Participants != nil {
		parts := make(storage.JSONMap)
		for i, p := range req.Participants {
			parts[string(rune('0'+i))] = p
		}
		entry.Participants = parts
	}

	// Set searchable text for full-text search
	searchable := req.Title + " " + req.Body
	entry.SearchableText = &searchable

	created, err := h.repo.CreateLivingMemoryEntry(r.Context(), entry)
	if err != nil {
		h.log.WithError(err).Error("Failed to create living memory entry")
		apierror.WriteError(w, apierror.NewInternal("Failed to create memory"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entry": created,
	})
}

func (h *Handler) HandleGetMemory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid memory entry ID"))
		return
	}

	entry, err := h.repo.GetLivingMemoryEntry(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get living memory entry")
		apierror.WriteError(w, apierror.NewInternal("Failed to get memory"))
		return
	}
	if entry == nil {
		apierror.WriteError(w, apierror.NewNotFound("Memory entry not found"))
		return
	}

	h.repo.IncrementLivingMemoryViewCount(r.Context(), id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entry": entry,
	})
}

func (h *Handler) HandleListMemory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListLivingMemoryOpts{
		Limit:  20,
		Offset: 0,
	}
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
			opts.Limit = n
		}
	}
	if o := q.Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			opts.Offset = n
		}
	}
	if mt := q.Get("memory_type"); mt != "" {
		opts.MemoryType = &mt
	}
	if pid := q.Get("project_id"); pid != "" {
		if id, err := uuid.Parse(pid); err == nil {
			opts.ProjectID = &id
		}
	}
	if imp := q.Get("importance"); imp != "" {
		opts.Importance = &imp
	}

	entries, total, err := h.repo.ListLivingMemoryEntries(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list living memory entries")
		apierror.WriteError(w, apierror.NewInternal("Failed to list memory"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": entries,
		"total":   total,
		"limit":   opts.Limit,
		"offset":  opts.Offset,
	})
}
