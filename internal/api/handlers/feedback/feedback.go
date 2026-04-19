package feedback

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/services"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Handler handles feedback-related HTTP requests
type Handler struct {
	repo           storage.Repository
	storageService *services.StorageService
}

// NewHandler creates a new feedback handler
func NewHandler(repo storage.Repository, storageService *services.StorageService) *Handler {
	return &Handler{
		repo:           repo,
		storageService: storageService,
	}
}

// getStorageBucketName returns the configured bucket name for attachment metadata (env-driven).
func getStorageBucketName() string {
	if b := os.Getenv("STORAGE_BUCKET"); b != "" {
		return b
	}
	return "functionfly-uploads"
}

// CreateFeedback handles POST /v1/feedback
func (h *Handler) CreateFeedback(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (for file uploads)
	err := r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Extract basic fields
	feedbackType := r.FormValue("feedbackType")
	subject := r.FormValue("subject")
	message := r.FormValue("message")
	priority := r.FormValue("priority")
	browserInfo := r.FormValue("browserInfo")

	// Validate required fields
	if feedbackType == "" || subject == "" || message == "" {
		http.Error(w, `{"error":"feedbackType, subject, and message are required"}`, http.StatusBadRequest)
		return
	}

	// Launch waitlist: optional email, short message (interests)
	if feedbackType == "launch_waitlist" {
		email := strings.TrimSpace(r.FormValue("email"))
		if email != "" && (len(email) > 254 || !strings.Contains(email, "@") || !strings.Contains(email, ".")) {
			http.Error(w, `{"error":"Invalid email"}`, http.StatusBadRequest)
			return
		}
	} else {
		// Smart form validation for other types
		if feedbackType == "bug" && !strings.Contains(strings.ToLower(message), "steps to reproduce") {
			http.Error(w, `{"error":"Bug reports should include steps to reproduce"}`, http.StatusBadRequest)
			return
		}
		if feedbackType == "feature" && len(message) < 50 {
			http.Error(w, `{"error":"Feature requests need more detail (minimum 50 characters)"}`, http.StatusBadRequest)
			return
		}
	}

	if len(message) > 1000 {
		http.Error(w, `{"error":"Message must be 1000 characters or less"}`, http.StatusBadRequest)
		return
	}

	// Get user context (if authenticated)
	var userID *uuid.UUID
	var userEmail *string

	if user := middleware.GetUserFromContext(r); user != nil {
		userID = &user.UserID
	} else {
		// For anonymous feedback (e.g. launch waitlist), use optional form email
		emailForm := strings.TrimSpace(r.FormValue("email"))
		if emailForm != "" {
			userEmail = &emailForm
		}
	}

	// Rate limiting check
	if userID != nil {
		// For authenticated users, check if they've submitted feedback recently
		feedbacks, err := h.repo.GetFeedbackByUser(userID, nil, 1, 0)
		if err == nil && len(feedbacks) > 0 {
			lastSubmission := feedbacks[0].CreatedAt
			if time.Since(lastSubmission) < time.Hour {
				http.Error(w, `{"error":"Please wait an hour before submitting another feedback"}`, http.StatusTooManyRequests)
				return
			}
		}
	}

	// Normalize IP for Postgres INET (address only, no port). Use nil if not parseable so DB accepts NULL.
	ipAddr := r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		ipAddr = host
	}
	if ipAddr != "" && net.ParseIP(ipAddr) == nil {
		ipAddr = "" // invalid for INET; store as NULL
	}

	// Create feedback
	feedback := &storage.Feedback{
		FeedbackType: feedbackType,
		Subject:      subject,
		Message:      message,
		Priority:     priority,
		BrowserInfo:  browserInfo,
		UserID:       userID,
		UserEmail:    userEmail,
		IPAddress:    ipAddr,
		UserAgent:    r.Header.Get("User-Agent"),
	}

	createdFeedback, err := h.repo.CreateFeedback(feedback)
	if err != nil {
		logrus.WithError(err).WithField("feedback_type", feedbackType).Error("CreateFeedback failed")
		// If DB rejects feedback_type (e.g. launch_waitlist not in CHECK), tell operator to run migration
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23514" && strings.Contains(strings.ToLower(pqErr.Message), "feedback_type") {
			http.Error(w, `{"error":"Launch waitlist is not enabled. Run migration 20260328000000_feedback_launch_waitlist on the database."}`, http.StatusBadRequest)
			return
		}
		http.Error(w, `{"error":"Failed to create feedback"}`, http.StatusInternalServerError)
		return
	}

	// Handle file uploads
	files := r.MultipartForm.File
	if len(files) > 0 {
		for fieldName, fileHeaders := range files {
			if strings.HasPrefix(fieldName, "attachment_") {
				for _, fileHeader := range fileHeaders {
					// Validate file
					if fileHeader.Size > 10*1024*1024 { // 10MB
						continue // Skip oversized files
					}

					allowedTypes := []string{
						"image/jpeg", "image/png", "image/gif", "image/webp",
						"text/plain", "text/log",
					}

					contentType := fileHeader.Header.Get("Content-Type")
					isAllowed := false
					for _, allowedType := range allowedTypes {
						if contentType == allowedType || strings.HasSuffix(fileHeader.Filename, allowedType[5:]) {
							isAllowed = true
							break
						}
					}

				if !isAllowed {
					continue // Skip unsupported files
				}

				file, err := fileHeader.Open()
					if err != nil {
						continue
					}
					defer file.Close()

					// Generate unique path for storage
					storagePath := h.storageService.GenerateUniquePath(
						fmt.Sprintf("feedback/%s", createdFeedback.ID),
						fileHeader.Filename,
					)

					// Upload file to storage backend (local, S3, R2, or S3-compatible e.g. B2)
					_, err = h.storageService.UploadFile(r.Context(), fileHeader, storagePath)
					if err != nil {
						// Log error but don't fail the entire request
						fmt.Printf("Failed to upload file to storage: %v\n", err)
						continue
					}

					// Create attachment record with actual storage info (bucket from env for metadata)
					attachment := &storage.FeedbackAttachment{
						FeedbackID:  createdFeedback.ID,
						Filename:    fileHeader.Filename,
						ContentType: contentType,
						Size:        fileHeader.Size,
						S3Key:       storagePath,
						S3Bucket:    getStorageBucketName(),
					}

					_, err = h.repo.CreateFeedbackAttachment(attachment)
					if err != nil {
						// Log error but don't fail the entire request
						fmt.Printf("Failed to create attachment: %v\n", err)
					}
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdFeedback)
}

// GetFeedbackHistory handles GET /v1/feedback/history
func (h *Handler) GetFeedbackHistory(w http.ResponseWriter, r *http.Request) {
	// Get user context
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	limit := 10 // Default limit
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l >= 1 && l <= 50 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 && o <= 1000 {
			offset = o
		}
	}

	// Get feedback history
	feedbacks, err := h.repo.GetFeedbackByUser(&user.UserID, nil, limit, offset)
	if err != nil {
		http.Error(w, `{"error":"Failed to retrieve feedback history"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"feedback": feedbacks,
		"limit":    limit,
		"offset":   offset,
	})
}

// ListFeedback handles GET /v1/admin/feedback (admin only).
// Query params: limit, offset, status, feedback_type (e.g. feedback_type=launch_waitlist for waitlist signups).
func (h *Handler) ListFeedback(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limit := 50 // Default limit for admin
	offset := 0
	var statusFilter *string
	var typeFilter *string

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l >= 1 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 && o <= 10000 {
			offset = o
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		statusFilter = &status
	}

	if ft := r.URL.Query().Get("feedback_type"); ft != "" {
		typeFilter = &ft
	}

	// Get feedback list
	feedbacks, err := h.repo.ListFeedback(limit, offset, statusFilter, typeFilter)
	if err != nil {
		http.Error(w, `{"error":"Failed to retrieve feedback"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"feedback": feedbacks,
		"limit":    limit,
		"offset":   offset,
	})
}

// UpdateFeedbackStatus handles PATCH /v1/admin/feedback/{id}/status (admin only)
func (h *Handler) UpdateFeedbackStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	feedbackIDStr := vars["id"]

	feedbackID, err := uuid.Parse(feedbackIDStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid feedback ID"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate status
	validStatuses := []string{"submitted", "in-review", "resolved", "closed"}
	isValid := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValid = true
			break
		}
	}

	if !isValid {
		http.Error(w, `{"error":"Invalid status. Must be one of: submitted, in-review, resolved, closed"}`, http.StatusBadRequest)
		return
	}

	// Update status
	err = h.repo.UpdateFeedbackStatus(feedbackID, req.Status)
	if err != nil {
		http.Error(w, `{"error":"Failed to update feedback status"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Status updated successfully"})
}

// DownloadAttachment handles GET /v1/feedback/attachments/{id}/download (auth required: owner or admin).
// Streams the file from storage (local, S3, R2, or B2) for private buckets.
func (h *Handler) DownloadAttachment(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	attachmentIDStr := vars["id"]
	attachmentID, err := uuid.Parse(attachmentIDStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid attachment ID"}`, http.StatusBadRequest)
		return
	}

	attachment, err := h.repo.GetFeedbackAttachmentByID(attachmentID)
	if err != nil || attachment == nil {
		http.Error(w, `{"error":"Attachment not found"}`, http.StatusNotFound)
		return
	}

	feedback, err := h.repo.GetFeedbackByID(attachment.FeedbackID)
	if err != nil || feedback == nil {
		http.Error(w, `{"error":"Feedback not found"}`, http.StatusNotFound)
		return
	}

	// Allow if user is the feedback owner or has system read (admin)
	allowed := feedback.UserID != nil && *feedback.UserID == user.UserID
	if !allowed && user.Permissions != nil {
		for _, p := range user.Permissions {
			if p == auth.PermSystemRead {
				allowed = true
				break
			}
		}
	}
	if !allowed {
		http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
		return
	}

	rc, err := h.storageService.GetFile(r.Context(), attachment.S3Key)
	if err != nil {
		http.Error(w, `{"error":"Failed to load file"}`, http.StatusInternalServerError)
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", attachment.Filename))
	if attachment.ContentType != "" {
		w.Header().Set("Content-Type", attachment.ContentType)
	}
	w.Header().Set("Content-Length", strconv.FormatInt(attachment.Size, 10))
	_, _ = io.Copy(w, rc)
}

// GetFeedbackStats handles GET /v1/admin/feedback/stats (admin only)
func (h *Handler) GetFeedbackStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.GetFeedbackStats()
	if err != nil {
		http.Error(w, `{"error":"Failed to retrieve feedback stats"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetFeedbackAnalytics handles GET /v1/admin/feedback/analytics (admin only)
func (h *Handler) GetFeedbackAnalytics(w http.ResponseWriter, r *http.Request) {
	analytics, err := h.repo.GetFeedbackAnalytics()
	if err != nil {
		http.Error(w, `{"error":"Failed to retrieve feedback analytics"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

// ExportFeedback handles GET /v1/admin/feedback/export (admin only)
func (h *Handler) ExportFeedback(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json" // default format
	}

	if format != "json" && format != "csv" {
		http.Error(w, `{"error":"Invalid format. Must be 'json' or 'csv'"}`, http.StatusBadRequest)
		return
	}

	// Parse filters
	statusFilter := r.URL.Query().Get("status")
	typeFilter := r.URL.Query().Get("type")
	priorityFilter := r.URL.Query().Get("priority")
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	// Get all feedback (admin can see all)
	feedbacks, err := h.repo.ListFeedback(10000, 0, nil, nil) // Get up to 10k records
	if err != nil {
		http.Error(w, `{"error":"Failed to retrieve feedback"}`, http.StatusInternalServerError)
		return
	}

	// Apply filters
	filteredFeedback := make([]storage.Feedback, 0)
	for _, fb := range feedbacks {
		// Status filter
		if statusFilter != "" && fb.Status != statusFilter {
			continue
		}

		// Type filter
		if typeFilter != "" && fb.FeedbackType != typeFilter {
			continue
		}

		// Priority filter
		if priorityFilter != "" && fb.Priority != priorityFilter {
			continue
		}

		// Date filters
		if dateFrom != "" {
			if fromDate, err := time.Parse("2006-01-02", dateFrom); err == nil {
				if fb.CreatedAt.Before(fromDate) {
					continue
				}
			}
		}

		if dateTo != "" {
			if toDate, err := time.Parse("2006-01-02", dateTo); err == nil {
				toDate = toDate.AddDate(0, 0, 1) // Include the end date
				if fb.CreatedAt.After(toDate) {
					continue
				}
			}
		}

		filteredFeedback = append(filteredFeedback, fb)
	}

	// Export based on format
	if format == "csv" {
		h.exportCSV(w, filteredFeedback)
	} else {
		h.exportJSON(w, filteredFeedback)
	}
}

func (h *Handler) exportCSV(w http.ResponseWriter, feedbacks []storage.Feedback) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=feedback_export.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"ID", "User Email", "Feedback Type", "Subject", "Message", "Priority",
		"Status", "Browser Info", "IPAddress", "User Agent", "Created At", "Updated At",
	}
	writer.Write(header)

	// Write data
	for _, fb := range feedbacks {
		userEmail := ""
		if fb.UserEmail != nil {
			userEmail = *fb.UserEmail
		}
		record := []string{
			fb.ID.String(),
			userEmail,
			fb.FeedbackType,
			fb.Subject,
			fb.Message,
			fb.Priority,
			fb.Status,
			fb.BrowserInfo,
			fb.IPAddress,
			fb.UserAgent,
			fb.CreatedAt.Format(time.RFC3339),
			fb.UpdatedAt.Format(time.RFC3339),
		}
		writer.Write(record)
	}
}

func (h *Handler) exportJSON(w http.ResponseWriter, feedbacks []storage.Feedback) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=feedback_export.json")

	// Convert to export format (exclude sensitive internal fields)
	exportData := make([]map[string]interface{}, 0, len(feedbacks))
	for _, fb := range feedbacks {
		userEmail := ""
		if fb.UserEmail != nil {
			userEmail = *fb.UserEmail
		}
		exportItem := map[string]interface{}{
			"id":              fb.ID.String(),
			"user_email":      userEmail,
			"feedback_type":   fb.FeedbackType,
			"subject":         fb.Subject,
			"message":         fb.Message,
			"priority":        fb.Priority,
			"status":          fb.Status,
			"browser_info":    fb.BrowserInfo,
			"created_at":      fb.CreatedAt.Format(time.RFC3339),
			"updated_at":      fb.UpdatedAt.Format(time.RFC3339),
			"has_attachments": len(fb.Attachments) > 0,
		}
		exportData = append(exportData, exportItem)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"total_count": len(feedbacks),
		"data":        exportData,
	})
}
