package registry

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type submitReviewRequest struct {
	Stars int    `json:"stars"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

// HandleListReviews returns recent user reviews for a function (public).
func (h *Handler) HandleListReviews(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	limit := 20
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil || fn == nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	reviews, total, err := h.repo.ListFunctionReviews(r.Context(), fn.ID, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list reviews")
		http.Error(w, "Failed to list reviews", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"reviews": reviews,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// HandleSubmitReview creates/updates the current user's review for a function (protected).
func (h *Handler) HandleSubmitReview(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	author := vars["author"]
	name := vars["name"]

	var req submitReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)

	if req.Stars < 1 || req.Stars > 5 {
		http.Error(w, "stars must be between 1 and 5", http.StatusBadRequest)
		return
	}
	// Keep payload sane; UI can still show full text if needed later.
	if len(req.Title) > 120 {
		req.Title = req.Title[:120]
	}
	if len(req.Body) > 5000 {
		req.Body = req.Body[:5000]
	}

	fn, err := h.repo.GetFunctionByAuthorName(author, name)
	if err != nil || fn == nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	executed, err := h.repo.HasUserExecutedFunction(r.Context(), fn.ID, user.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check review eligibility")
		http.Error(w, "Failed to submit review", http.StatusInternalServerError)
		return
	}
	if !executed {
		// Allow platform admins/support to review without executing (useful for moderation and QA).
		switch user.Role {
		case "super_admin", "admin", "support", "billing_admin", "developer_admin":
			// ok: proceed
		default:
			http.Error(w, "Please execute this function at least once before leaving a review.", http.StatusForbidden)
			return
		}
	}

	review, err := h.repo.UpsertFunctionReview(r.Context(), fn.ID, user.UserID, req.Stars, req.Title, req.Body)
	if err != nil {
		logrus.WithError(err).Error("Failed to submit review")
		http.Error(w, "Failed to submit review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":     true,
		"review": review,
	})
}

