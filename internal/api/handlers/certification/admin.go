package certification

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// RegisterAdminRoutes registers admin certification routes on adminRouter.
func (h *Handler) RegisterAdminRoutes(adminRouter *mux.Router, adminMW *middleware.AuthMiddleware) {
	adminRouter.HandleFunc("/tiers", adminMW.RequirePermission(auth.PermSystemWrite)(h.AdminListTiers)).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/tiers", adminMW.RequirePermission(auth.PermSystemWrite)(h.AdminCreateTier)).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/tiers/{tierID}", adminMW.RequirePermission(auth.PermSystemWrite)(h.AdminGetTier)).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/tiers/{tierID}", adminMW.RequirePermission(auth.PermSystemWrite)(h.AdminUpdateTier)).Methods("PUT", "PATCH", "OPTIONS")
	adminRouter.HandleFunc("/tiers/{tierID}", adminMW.RequirePermission(auth.PermSystemWrite)(h.AdminDeleteTier)).Methods("DELETE", "OPTIONS")
	adminRouter.HandleFunc("/tiers/{tierID}/questions", adminMW.RequirePermission(auth.PermSystemWrite)(h.AdminListQuestionsForTier)).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/tiers/{tierID}/questions", adminMW.RequirePermission(auth.PermSystemWrite)(h.AdminCreateQuestionForTier)).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/questions/{questionID}", adminMW.RequirePermission(auth.PermSystemWrite)(h.AdminGetQuestion)).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/questions/{questionID}", adminMW.RequirePermission(auth.PermSystemWrite)(h.AdminUpdateQuestion)).Methods("PUT", "PATCH", "OPTIONS")
	adminRouter.HandleFunc("/questions/{questionID}", adminMW.RequirePermission(auth.PermSystemWrite)(h.AdminDeleteQuestion)).Methods("DELETE", "OPTIONS")
	adminRouter.HandleFunc("/exams", adminMW.RequirePermission(auth.PermSystemRead)(h.AdminListExams)).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/exams/{examID}", adminMW.RequirePermission(auth.PermSystemRead)(h.AdminGetExam)).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/exams/{examID}/grade", adminMW.RequirePermission(auth.PermSystemWrite)(h.AdminGradeExam)).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/credentials", adminMW.RequirePermission(auth.PermSystemRead)(h.AdminListCredentials)).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/credentials/{credentialID}", adminMW.RequirePermission(auth.PermSystemRead)(h.AdminGetCredential)).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/credentials/{credentialID}/revoke", adminMW.RequirePermission(auth.PermSystemWrite)(h.AdminRevokeCredential)).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/exams/{examID}/reset", adminMW.RequirePermission(auth.PermSystemWrite)(h.AdminResetExam)).Methods("POST", "OPTIONS")
	adminRouter.HandleFunc("/users/{userID}/exams/tier/{tierSlug}/reset", adminMW.RequirePermission(auth.PermSystemWrite)(h.AdminResetUserTierExamAttempts)).Methods("POST", "OPTIONS")
}

func (h *Handler) AdminListTiers(w http.ResponseWriter, r *http.Request) {
	tiers, err := h.repo.ListTiers(r.Context())
	if err != nil {
		logrus.WithError(err).Error("adminListTiers failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list tiers")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"tiers": tiers})
}

func (h *Handler) AdminCreateTier(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Slug             string  `json:"slug"`
		Name             string  `json:"name"`
		Description      string  `json:"description"`
		Icon             string  `json:"icon"`
		Color            string  `json:"color"`
		SortOrder        int     `json:"sort_order"`
		PriceCents       int     `json:"price_cents"`
		PassThreshold    float64 `json:"pass_threshold"`
		TimeLimitMinutes int     `json:"time_limit_minutes"`
		QuestionCount    int     `json:"question_count"`
		PracticalCount   int     `json:"practical_count"`
		ValidityMonths   int     `json:"validity_months"`
		IsActive         *bool   `json:"is_active"`
		IsComingSoon     *bool   `json:"is_coming_soon"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Slug == "" || req.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "slug and name are required")
		return
	}
	if req.PassThreshold <= 0 {
		req.PassThreshold = 70
	}
	if req.TimeLimitMinutes <= 0 {
		req.TimeLimitMinutes = 90
	}
	if req.QuestionCount <= 0 {
		req.QuestionCount = 50
	}
	if req.PracticalCount <= 0 {
		req.PracticalCount = 3
	}
	if req.ValidityMonths <= 0 {
		req.ValidityMonths = 24
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	isComingSoon := false
	if req.IsComingSoon != nil {
		isComingSoon = *req.IsComingSoon
	}
	tier := &storage.CertTier{
		Slug:             req.Slug,
		Name:             req.Name,
		Description:      req.Description,
		Icon:             req.Icon,
		Color:            req.Color,
		SortOrder:        req.SortOrder,
		PriceCents:       req.PriceCents,
		Currency:         "USD",
		PassThreshold:    req.PassThreshold,
		TimeLimitMinutes: req.TimeLimitMinutes,
		QuestionCount:    req.QuestionCount,
		PracticalCount:   req.PracticalCount,
		ValidityMonths:   req.ValidityMonths,
		IsActive:         isActive,
		IsComingSoon:     isComingSoon,
		Metadata:         storage.JSONMap{},
	}
	if err := h.repo.CreateTier(r.Context(), tier); err != nil {
		logrus.WithError(err).Error("adminCreateTier failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create tier")
		return
	}
	writeJSON(w, http.StatusCreated, tier)
}

func (h *Handler) AdminGetTier(w http.ResponseWriter, r *http.Request) {
	tierID, err := uuid.Parse(mux.Vars(r)["tierID"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid tier ID")
		return
	}
	tier, err := h.repo.GetTierByID(r.Context(), tierID)
	if err != nil || tier == nil {
		writeJSONError(w, http.StatusNotFound, "Tier not found")
		return
	}
	writeJSON(w, http.StatusOK, tier)
}

func (h *Handler) AdminUpdateTier(w http.ResponseWriter, r *http.Request) {
	tierID, err := uuid.Parse(mux.Vars(r)["tierID"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid tier ID")
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.repo.UpdateTier(r.Context(), tierID, updates); err != nil {
		logrus.WithError(err).Error("adminUpdateTier failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update tier")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Tier updated"})
}

func (h *Handler) AdminDeleteTier(w http.ResponseWriter, r *http.Request) {
	tierID, err := uuid.Parse(mux.Vars(r)["tierID"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid tier ID")
		return
	}
	// Soft-delete via is_active=false
	if err := h.repo.UpdateTier(r.Context(), tierID, storage.JSONMap{"is_active": false}); err != nil {
		logrus.WithError(err).Error("adminDeleteTier failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete tier")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Tier deactivated"})
}

func (h *Handler) AdminListQuestionsForTier(w http.ResponseWriter, r *http.Request) {
	tierID, err := uuid.Parse(mux.Vars(r)["tierID"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid tier ID")
		return
	}
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	questions, err := h.repo.GetQuestionsByTier(r.Context(), tierID, limit)
	if err != nil {
		logrus.WithError(err).Error("adminListQuestionsForTier failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list questions")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"questions": questions})
}

func (h *Handler) AdminCreateQuestionForTier(w http.ResponseWriter, r *http.Request) {
	tierID, err := uuid.Parse(mux.Vars(r)["tierID"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid tier ID")
		return
	}
	var req struct {
		Category       string                 `json:"category"`
		Difficulty     string                 `json:"difficulty"`
		QuestionText   string                 `json:"question_text"`
		QuestionFormat string                 `json:"question_format"`
		Options        interface{}            `json:"options"`
		CorrectAnswers storage.JSONMap        `json:"correct_answers"`
		Explanation    string                 `json:"explanation"`
		Points         int                    `json:"points"`
		IsActive       *bool                  `json:"is_active"`
		Metadata       storage.JSONMap        `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.QuestionText == "" || req.Category == "" {
		writeJSONError(w, http.StatusBadRequest, "category and question_text are required")
		return
	}
	if req.Options == nil {
		req.Options = map[string]interface{}{}
	}
	if req.CorrectAnswers == nil {
		req.CorrectAnswers = storage.JSONMap{}
	}
	if req.Points <= 0 {
		req.Points = 1
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	question := &storage.CertQuestion{
		TierID:         tierID,
		Category:       req.Category,
		Difficulty:     req.Difficulty,
		QuestionText:   req.QuestionText,
		QuestionFormat: req.QuestionFormat,
		Options:        req.Options,
		CorrectAnswers: req.CorrectAnswers,
		Explanation:    req.Explanation,
		Points:         req.Points,
		IsActive:       isActive,
		Metadata:       req.Metadata,
	}
	if err := h.repo.CreateQuestion(r.Context(), question); err != nil {
		logrus.WithError(err).Error("adminCreateQuestionForTier failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create question")
		return
	}
	writeJSON(w, http.StatusCreated, question)
}

func (h *Handler) AdminGetQuestion(w http.ResponseWriter, r *http.Request) {
	questionID, err := uuid.Parse(mux.Vars(r)["questionID"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid question ID")
		return
	}
	q, err := h.repo.GetQuestionByID(r.Context(), questionID)
	if err != nil || q == nil {
		writeJSONError(w, http.StatusNotFound, "Question not found")
		return
	}
	writeJSON(w, http.StatusOK, q)
}

func (h *Handler) AdminUpdateQuestion(w http.ResponseWriter, r *http.Request) {
	questionID, err := uuid.Parse(mux.Vars(r)["questionID"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid question ID")
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.repo.UpdateQuestion(r.Context(), questionID, updates); err != nil {
		logrus.WithError(err).Error("adminUpdateQuestion failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update question")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Question updated"})
}

func (h *Handler) AdminDeleteQuestion(w http.ResponseWriter, r *http.Request) {
	questionID, err := uuid.Parse(mux.Vars(r)["questionID"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid question ID")
		return
	}
	if err := h.repo.DeleteQuestion(r.Context(), questionID); err != nil {
		logrus.WithError(err).Error("adminDeleteQuestion failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete question")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Question deleted"})
}

func (h *Handler) AdminListExams(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	status := r.URL.Query().Get("status")
	tierIDStr := r.URL.Query().Get("tier_id")
	userIDStr := r.URL.Query().Get("user_id")

	exams, total, err := h.repo.ListExams(r.Context(), storage.CertExamListFilter{
		Limit:  limit,
		Offset: offset,
		Status: status,
		TierID: ptrUUID(tierIDStr),
		UserID: ptrUUID(userIDStr),
	})
	if err != nil {
		logrus.WithError(err).Error("adminListExams failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list exams")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"exams":  exams,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) AdminGetExam(w http.ResponseWriter, r *http.Request) {
	examID, err := uuid.Parse(mux.Vars(r)["examID"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid exam ID")
		return
	}
	exam, err := h.repo.GetExamByID(r.Context(), examID)
	if err != nil || exam == nil {
		writeJSONError(w, http.StatusNotFound, "Exam not found")
		return
	}
	writeJSON(w, http.StatusOK, exam)
}

type adminGradeExamRequest struct {
	KnowledgeScore *float64 `json:"knowledge_score"`
	PracticalScore *float64 `json:"practical_score"`
	TotalScore     *float64 `json:"total_score"`
	Passed         *bool    `json:"passed"`
}

func (h *Handler) AdminGradeExam(w http.ResponseWriter, r *http.Request) {
	examID, err := uuid.Parse(mux.Vars(r)["examID"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid exam ID")
		return
	}
	var req adminGradeExamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	exam, err := h.repo.GetExamByID(r.Context(), examID)
	if err != nil || exam == nil {
		writeJSONError(w, http.StatusNotFound, "Exam not found")
		return
	}
	var knowledgeScore, practicalScore, totalScore float64
	var passed bool
	if exam.KnowledgeScore != nil {
		knowledgeScore = *exam.KnowledgeScore
	}
	if exam.PracticalScore != nil {
		practicalScore = *exam.PracticalScore
	}
	if exam.TotalScore != nil {
		totalScore = *exam.TotalScore
	}
	if exam.Passed != nil {
		passed = *exam.Passed
	}
	if req.KnowledgeScore != nil {
		knowledgeScore = *req.KnowledgeScore
	}
	if req.PracticalScore != nil {
		practicalScore = *req.PracticalScore
	}
	if req.TotalScore != nil {
		totalScore = *req.TotalScore
	}
	if req.Passed != nil {
		passed = *req.Passed
	}
	if err := h.repo.GradeExam(r.Context(), examID, knowledgeScore, practicalScore, totalScore, passed); err != nil {
		logrus.WithError(err).Error("adminGradeExam failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to grade exam")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Exam graded"})
}

func (h *Handler) AdminListCredentials(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	status := r.URL.Query().Get("status")
	tierIDStr := r.URL.Query().Get("tier_id")
	userIDStr := r.URL.Query().Get("user_id")
	creds, err := h.repo.ListCredentials(r.Context(), storage.CertCredentialListFilter{
		Limit:  limit,
		Offset: offset,
		Status: status,
		TierID: ptrUUID(tierIDStr),
		UserID: ptrUUID(userIDStr),
	})
	if err != nil {
		logrus.WithError(err).Error("adminListCredentials failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list credentials")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"credentials": creds})
}

func (h *Handler) AdminGetCredential(w http.ResponseWriter, r *http.Request) {
	credID, err := uuid.Parse(mux.Vars(r)["credentialID"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid credential ID")
		return
	}
	cred, err := h.repo.GetCredentialByID(r.Context(), credID)
	if err != nil || cred == nil {
		writeJSONError(w, http.StatusNotFound, "Credential not found")
		return
	}
	writeJSON(w, http.StatusOK, cred)
}

func (h *Handler) AdminRevokeCredential(w http.ResponseWriter, r *http.Request) {
	credID, err := uuid.Parse(mux.Vars(r)["credentialID"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid credential ID")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.Reason == "" {
		req.Reason = "revoked by admin"
	}
	now := time.Now()
	if err := h.repo.UpdateCredential(r.Context(), credID, storage.JSONMap{
		"status":          storage.CertCredentialStatusRevoked,
		"revoked_at":      now,
		"revoked_reason":  req.Reason,
		"updated_at":      now,
	}); err != nil {
		logrus.WithError(err).Error("adminRevokeCredential failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to revoke credential")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Credential revoked"})
}

func (h *Handler) AdminResetExam(w http.ResponseWriter, r *http.Request) {
	examID, err := uuid.Parse(mux.Vars(r)["examID"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid exam ID")
		return
	}
	exam, err := h.repo.GetExamByID(r.Context(), examID)
	if err != nil || exam == nil {
		writeJSONError(w, http.StatusNotFound, "Exam not found")
		return
	}
	// Reset exam status to abandoned so user can start a new attempt
	if err := h.repo.UpdateExamStatus(r.Context(), examID, storage.CertExamStatusAbandoned); err != nil {
		logrus.WithError(err).Error("adminResetExam failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to reset exam")
		return
	}
	logrus.WithFields(logrus.Fields{"exam_id": examID, "user_id": exam.UserID}).Info("Admin reset exam")
	writeJSON(w, http.StatusOK, map[string]string{"message": "Exam reset successfully"})
}

func (h *Handler) AdminResetUserTierExamAttempts(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(mux.Vars(r)["userID"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	tierSlug := mux.Vars(r)["tierSlug"]

	tier, err := h.repo.GetTierBySlug(r.Context(), tierSlug)
	if err != nil || tier == nil {
		writeJSONError(w, http.StatusNotFound, "Certification tier not found")
		return
	}

	// Set all non-passed exams for this user+tier to abandoned
	if err := h.repo.ResetUserTierExamAttempts(r.Context(), userID, tier.ID); err != nil {
		logrus.WithError(err).Error("adminResetUserTierExamAttempts failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to reset exam attempts")
		return
	}

	logrus.WithFields(logrus.Fields{"user_id": userID, "tier_slug": tierSlug}).Info("Admin reset all exam attempts for user+tier")
	writeJSON(w, http.StatusOK, map[string]string{"message": "All exam attempts reset successfully"})
}

func ptrUUID(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &u
}

