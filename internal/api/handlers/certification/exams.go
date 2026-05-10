package certification

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// StartExam handles POST /v1/certification/tiers/{tierSlug}/start
// Creates a new exam session for the authenticated user
func (h *Handler) StartExam(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	tierSlug := mux.Vars(r)["tierSlug"]
	tier, err := h.repo.GetTierBySlug(r.Context(), tierSlug)
	if err != nil {
		logrus.WithError(err).WithField("tier", tierSlug).Error("Failed to get cert tier")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve tier")
		return
	}
	if tier == nil {
		writeJSONError(w, http.StatusNotFound, "Certification tier not found")
		return
	}

	// Check for existing in-progress exam
	existing, err := h.repo.GetActiveExamForUserTier(r.Context(), claims.UserID, tier.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to check existing exam")
		writeJSONError(w, http.StatusInternalServerError, "Failed to check existing exam")
		return
	}
	if existing != nil {
		writeJSONError(w, http.StatusConflict, "You already have an in-progress exam for this tier")
		return
	}

	// Select random questions
	questionIDs, err := h.repo.SelectRandomQuestions(r.Context(), tier.ID, tier.QuestionCount)
	if err != nil {
		logrus.WithError(err).Error("Failed to select questions")
		writeJSONError(w, http.StatusInternalServerError, "Failed to prepare exam")
		return
	}
	if len(questionIDs) < tier.QuestionCount {
		writeJSONError(w, http.StatusServiceUnavailable, "Not enough questions available for this tier")
		return
	}

	// Select random practical challenges
	practicalIDs, err := h.repo.SelectRandomChallenges(r.Context(), tier.ID, tier.PracticalCount)
	if err != nil {
		logrus.WithError(err).Error("Failed to select challenges")
		writeJSONError(w, http.StatusInternalServerError, "Failed to prepare exam")
		return
	}

	now := time.Now()
	exam := &storage.CertExam{
		UserID:       claims.UserID,
		TierID:       tier.ID,
		Status:       storage.CertExamStatusInProgress,
		AmountCents:  tier.PriceCents,
		Currency:     tier.Currency,
		StartedAt:    now,
		ExpiresAt:    now.Add(time.Duration(tier.TimeLimitMinutes) * time.Minute),
		QuestionIDs:  uuidsToJSONMap(questionIDs),
		PracticalIDs: uuidsToJSONMap(practicalIDs),
		Answers:      storage.JSONMap{},
		IPAddress:    strPtr(getClientIP(r)),
		UserAgent:    strPtr(r.UserAgent()),
		Metadata:     storage.JSONMap{},
	}

	if err := h.repo.CreateExam(r.Context(), exam); err != nil {
		logrus.WithError(err).Error("Failed to create exam")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create exam session")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"exam": map[string]interface{}{
			"id":                 exam.ID,
			"tier":               map[string]interface{}{"slug": tier.Slug, "name": tier.Name},
			"status":             exam.Status,
			"started_at":         exam.StartedAt,
			"expires_at":         exam.ExpiresAt,
			"time_limit_minutes": tier.TimeLimitMinutes,
			"question_count":     tier.QuestionCount,
			"practical_count":    tier.PracticalCount,
		},
	})
}

// GetExam handles GET /v1/certification/exams/{examId}
// Returns exam details with questions (without correct answers)
func (h *Handler) GetExam(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	examID, err := uuid.Parse(mux.Vars(r)["examId"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid exam ID")
		return
	}

	exam, err := h.repo.GetExamByID(r.Context(), examID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get exam")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve exam")
		return
	}
	if exam == nil {
		writeJSONError(w, http.StatusNotFound, "Exam not found")
		return
	}
	if exam.UserID != claims.UserID {
		writeJSONError(w, http.StatusForbidden, "Access denied")
		return
	}

	// Get question details (without answers)
	questionIDs := jsonMapToUUIDs(exam.QuestionIDs)
	questions, err := h.repo.GetQuestionsByIDs(r.Context(), questionIDs)
	if err != nil {
		logrus.WithError(err).Error("Failed to get questions")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve questions")
		return
	}

	// Get practical challenge details
	practicalIDs := jsonMapToUUIDs(exam.PracticalIDs)
	var challenges []map[string]interface{}
	for _, cid := range practicalIDs {
		ch, err := h.repo.GetChallengeByID(r.Context(), cid)
		if err != nil {
			logrus.WithError(err).Error("Failed to get challenge")
			continue
		}
		if ch != nil {
			challenges = append(challenges, map[string]interface{}{
				"id":                 ch.ID,
				"name":               ch.Name,
				"description":        ch.Description,
				"category":           ch.Category,
				"difficulty":         ch.Difficulty,
				"points":             ch.Points,
				"time_limit_minutes": ch.TimeLimitMinutes,
			})
		}
	}

	// Build response
	remaining := time.Until(exam.ExpiresAt)
	timeRemainingSeconds := int(math.Max(0, remaining.Seconds()))

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"exam": map[string]interface{}{
			"id":                     exam.ID,
			"status":                 exam.Status,
			"started_at":             exam.StartedAt,
			"expires_at":             exam.ExpiresAt,
			"time_remaining_seconds": timeRemainingSeconds,
			"answers":                exam.Answers,
			"questions":              questions,
			"practical_challenges":   challenges,
		},
	})
}

// SubmitAnswer handles PUT /v1/certification/exams/{examId}/answer
// Submits or updates an answer for a single question
func (h *Handler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	examID, err := uuid.Parse(mux.Vars(r)["examId"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid exam ID")
		return
	}

	var req struct {
		QuestionID string      `json:"question_id"`
		Answer     interface{} `json:"answer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.QuestionID == "" {
		writeJSONError(w, http.StatusBadRequest, "question_id is required")
		return
	}

	questionID, err := uuid.Parse(req.QuestionID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid question ID")
		return
	}

	// Verify exam ownership
	exam, err := h.repo.GetExamByID(r.Context(), examID)
	if err != nil || exam == nil {
		writeJSONError(w, http.StatusNotFound, "Exam not found")
		return
	}
	if exam.UserID != claims.UserID {
		writeJSONError(w, http.StatusForbidden, "Access denied")
		return
	}
	if exam.Status != storage.CertExamStatusInProgress {
		writeJSONError(w, http.StatusBadRequest, "Exam is not in progress")
		return
	}
	if time.Now().After(exam.ExpiresAt) {
		writeJSONError(w, http.StatusBadRequest, "Exam has expired")
		return
	}

	if err := h.repo.UpdateExamAnswer(r.Context(), examID, questionID, req.Answer); err != nil {
		logrus.WithError(err).Error("Failed to update answer")
		writeJSONError(w, http.StatusInternalServerError, "Failed to save answer")
		return
	}

	// Count submitted answers
	answerCount := len(exam.Answers)
	if _, exists := exam.Answers[questionID.String()]; !exists {
		answerCount++
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"saved":             true,
		"answers_submitted": answerCount,
	})
}

// SubmitExam handles POST /v1/certification/exams/{examId}/submit
// Submits the entire exam for grading
func (h *Handler) SubmitExam(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	examID, err := uuid.Parse(mux.Vars(r)["examId"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid exam ID")
		return
	}

	exam, err := h.repo.GetExamByID(r.Context(), examID)
	if err != nil || exam == nil {
		writeJSONError(w, http.StatusNotFound, "Exam not found")
		return
	}
	if exam.UserID != claims.UserID {
		writeJSONError(w, http.StatusForbidden, "Access denied")
		return
	}
	if exam.Status != storage.CertExamStatusInProgress {
		writeJSONError(w, http.StatusBadRequest, "Exam is not in progress")
		return
	}

	// Merge any answers from request body
	var req struct {
		Answers map[string]interface{} `json:"answers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.Answers != nil {
		for k, v := range req.Answers {
			exam.Answers[k] = v
		}
	}

	// Submit the exam
	if err := h.repo.SubmitExam(r.Context(), examID, exam.Answers); err != nil {
		logrus.WithError(err).Error("Failed to submit exam")
		writeJSONError(w, http.StatusInternalServerError, "Failed to submit exam")
		return
	}

	// Grade knowledge questions immediately
	go h.gradeKnowledgeQuestions(examID, exam)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "submitted",
		"message": "Exam submitted successfully. Knowledge questions will be graded immediately.",
	})
}

// AbandonExam handles POST /v1/certification/exams/{examId}/abandon
// Marks an exam as abandoned by the user
func (h *Handler) AbandonExam(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	examID, err := uuid.Parse(mux.Vars(r)["examId"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid exam ID")
		return
	}

	exam, err := h.repo.GetExamByID(r.Context(), examID)
	if err != nil || exam == nil {
		writeJSONError(w, http.StatusNotFound, "Exam not found")
		return
	}
	if exam.UserID != claims.UserID {
		writeJSONError(w, http.StatusForbidden, "Access denied")
		return
	}
	if exam.Status != storage.CertExamStatusInProgress {
		writeJSONError(w, http.StatusBadRequest, "Exam is not in progress")
		return
	}

	if err := h.repo.AbandonExam(r.Context(), examID); err != nil {
		logrus.WithError(err).Error("Failed to abandon exam")
		writeJSONError(w, http.StatusInternalServerError, "Failed to abandon exam")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "abandoned",
		"message": "Exam abandoned successfully.",
	})
}

// gradeKnowledgeQuestions grades the knowledge portion of an exam asynchronously
func (h *Handler) gradeKnowledgeQuestions(examID uuid.UUID, exam *storage.CertExam) {
	// This runs in a goroutine — uses background context
	ctx := context.Background()

	questionIDs := jsonMapToUUIDs(exam.QuestionIDs)
	correctAnswers, err := h.repo.GetCorrectAnswers(ctx, questionIDs)
	if err != nil {
		logrus.WithError(err).Error("Failed to get correct answers for grading")
		return
	}

	totalPoints := 0
	earnedPoints := 0

	// Parse user answers from exam
	userAnswers := make(map[string]string)
	for k, v := range exam.Answers {
		if s, ok := v.(string); ok {
			userAnswers[k] = s
		}
	}

	for qID, correct := range correctAnswers {
		points := 1
		if p, ok := correct["_points"].(int); ok {
			points = p
		}
		totalPoints += points

		userAnswer, exists := userAnswers[qID.String()]
		if !exists {
			continue
		}

		// Check if answer matches any correct answer
		if correctAnswersList, ok := correct["answers"]; ok {
			if answers, ok := correctAnswersList.([]interface{}); ok {
				for _, a := range answers {
					if s, ok := a.(string); ok && s == userAnswer {
						earnedPoints += points
						break
					}
				}
			}
		}
	}

	knowledgeScore := 0.0
	if totalPoints > 0 {
		knowledgeScore = (float64(earnedPoints) / float64(totalPoints)) * 100
	}

	// For now, practical score defaults to 0 (graded asynchronously)
	// The final grade will be computed once practical challenges are graded
	logrus.WithFields(logrus.Fields{
		"exam_id":          examID,
		"knowledge_score":  knowledgeScore,
		"earned_points":    earnedPoints,
		"total_points":     totalPoints,
	}).Info("Knowledge questions graded")
}

// ListMyExams handles GET /v1/certification/exams
// Returns the authenticated user's exam history
func (h *Handler) ListMyExams(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limit := 20
	offset := 0
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && o >= 0 {
		offset = o
	}

	exams, total, err := h.repo.ListExamsByUser(r.Context(), claims.UserID, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list exams")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve exams")
		return
	}

	result := make([]map[string]interface{}, 0, len(exams))
	for _, e := range exams {
		entry := map[string]interface{}{
			"id":           e.ID,
			"tier_id":      e.TierID,
			"status":       e.Status,
			"started_at":   e.StartedAt,
			"expires_at":   e.ExpiresAt,
			"amount_cents": e.AmountCents,
			"created_at":   e.CreatedAt,
		}
		if e.SubmittedAt != nil {
			entry["submitted_at"] = e.SubmittedAt
		}
		if e.GradedAt != nil {
			entry["graded_at"] = e.GradedAt
		}
		if e.KnowledgeScore != nil {
			entry["knowledge_score"] = e.KnowledgeScore
		}
		if e.TotalScore != nil {
			entry["total_score"] = e.TotalScore
		}
		if e.Passed != nil {
			entry["passed"] = e.Passed
		}
		result = append(result, entry)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"exams":  result,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

func uuidsToJSONMap(ids []uuid.UUID) storage.JSONMap {
	arr := make([]interface{}, len(ids))
	for i, id := range ids {
		arr[i] = id.String()
	}
	return storage.JSONMap{"_ids": arr}
}

func jsonMapToUUIDs(m storage.JSONMap) []uuid.UUID {
	raw, ok := m["_ids"]
	if !ok {
		return nil
	}
	arr, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, v := range arr {
		s, ok := v.(string)
		if !ok {
			continue
		}
		if uid, err := uuid.Parse(s); err == nil {
			ids = append(ids, uid)
		}
	}
	return ids
}

func strPtr(s string) *string {
	return &s
}
