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

func (h *Handler) HandleListDocuments(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListDocumentsOpts{
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
	if t := q.Get("doc_type"); t != "" {
		opts.DocType = &t
	}
	if s := q.Get("status"); s != "" {
		opts.Status = &s
	}
	if search := q.Get("search"); search != "" {
		opts.Search = &search
	}

	docs, total, err := h.repo.ListDocuments(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list documents")
		apierror.WriteError(w, apierror.NewInternal("Failed to list documents"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"documents": docs,
		"total":     total,
		"limit":     opts.Limit,
		"offset":    opts.Offset,
	})
}

func (h *Handler) HandleGetDocument(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid document ID"))
		return
	}

	doc, err := h.repo.GetDocumentByID(r.Context(), id)
	if err != nil {
		h.log.WithError(err).Error("Failed to get document")
		apierror.WriteError(w, apierror.NewInternal("Failed to get document"))
		return
	}
	if doc == nil {
		apierror.WriteError(w, apierror.NewNotFound("Document not found"))
		return
	}

	h.repo.IncrementDocumentViewCount(r.Context(), id)

	shares, _ := h.repo.ListDocumentShares(r.Context(), id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"document": doc,
		"shares":   shares,
	})
}

type createDocumentRequest struct {
	Title      string   `json:"title"`
	Body       string   `json:"body,omitempty"`
	DocType    string   `json:"doc_type,omitempty"`
	Category   string   `json:"category,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	IsTemplate bool     `json:"is_template,omitempty"`
}

func (h *Handler) HandleCreateDocument(w http.ResponseWriter, r *http.Request) {
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

	var req createDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.Title == "" {
		apierror.WriteError(w, apierror.NewBadRequest("title is required"))
		return
	}

	doc := &storage.Document{
		TenantID:   claims.TenantID,
		AuthorID:   emp.ID,
		Title:      req.Title,
		DocType:    "note",
		Status:     "draft",
		IsTemplate: req.IsTemplate,
	}
	if req.Body != "" {
		doc.Body = &req.Body
	}
	if req.DocType != "" {
		doc.DocType = req.DocType
	}
	if req.Category != "" {
		doc.Category = &req.Category
	}
	if req.Tags != nil {
		doc.Tags = storage.JSONMap{"tags": req.Tags}
	}

	created, err := h.repo.CreateDocument(r.Context(), doc)
	if err != nil {
		h.log.WithError(err).Error("Failed to create document")
		apierror.WriteError(w, apierror.NewInternal("Failed to create document"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"document": created,
	})
}

type updateDocumentRequest struct {
	Title    *string  `json:"title,omitempty"`
	Body     *string  `json:"body,omitempty"`
	DocType  *string  `json:"doc_type,omitempty"`
	Category *string  `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Status   *string  `json:"status,omitempty"`
}

func (h *Handler) HandleUpdateDocument(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	id, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid document ID"))
		return
	}

	doc, err := h.repo.GetDocumentByID(r.Context(), id)
	if err != nil || doc == nil {
		apierror.WriteError(w, apierror.NewNotFound("Document not found"))
		return
	}

	var req updateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Body != nil {
		updates["body"] = *req.Body
	}
	if req.DocType != nil {
		updates["doc_type"] = *req.DocType
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Tags != nil {
		updates["tags"] = storage.JSONMap{"tags": req.Tags}
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := h.repo.UpdateDocument(r.Context(), id, updates); err != nil {
		h.log.WithError(err).Error("Failed to update document")
		apierror.WriteError(w, apierror.NewInternal("Failed to update document"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

type shareDocumentRequest struct {
	SharedWith string `json:"shared_with"`
	Permission string `json:"permission,omitempty"`
}

func (h *Handler) HandleShareDocument(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	docID, err := uuid.Parse(mux.Vars(r)["id"])
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid document ID"))
		return
	}

	doc, err := h.repo.GetDocumentByID(r.Context(), docID)
	if err != nil || doc == nil {
		apierror.WriteError(w, apierror.NewNotFound("Document not found"))
		return
	}

	var req shareDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}
	if req.SharedWith == "" {
		apierror.WriteError(w, apierror.NewBadRequest("shared_with is required"))
		return
	}

	sharedWithID, err := uuid.Parse(req.SharedWith)
	if err != nil {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid shared_with ID"))
		return
	}

	share := &storage.DocumentShare{
		DocumentID: docID,
		SharedWith: sharedWithID,
		Permission: "read",
	}
	if req.Permission != "" {
		share.Permission = req.Permission
	}

	created, err := h.repo.CreateDocumentShare(r.Context(), share)
	if err != nil {
		h.log.WithError(err).Error("Failed to share document")
		apierror.WriteError(w, apierror.NewInternal("Failed to share document"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"share": created,
	})
}

func (h *Handler) HandleListTemplates(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	q := r.URL.Query()
	opts := storage.ListDocumentsOpts{
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
	docType := "template"
	opts.DocType = &docType

	docs, total, err := h.repo.ListDocuments(r.Context(), claims.TenantID, opts)
	if err != nil {
		h.log.WithError(err).Error("Failed to list templates")
		apierror.WriteError(w, apierror.NewInternal("Failed to list templates"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"templates": docs,
		"total":     total,
		"limit":     opts.Limit,
		"offset":    opts.Offset,
	})
}
