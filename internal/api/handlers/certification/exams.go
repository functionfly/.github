package certification

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/payment"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// StartExam handles POST /v1/certification/tiers/{tierSlug}/start
// Creates a new exam session for the authenticated user
// If the tier has a price, creates a checkout session and returns payment URL
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

	// Check for existing paid exam that's not expired
	paidExam, err := h.repo.GetPaidExamForUserTier(r.Context(), claims.UserID, tier.ID)
	if err == nil && paidExam != nil && paidExam.Status == storage.CertExamStatusInProgress && time.Now().Before(paidExam.ExpiresAt) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"exam": map[string]interface{}{
				"id":                 paidExam.ID,
				"tier":               map[string]interface{}{"slug": tier.Slug, "name": tier.Name},
				"status":             paidExam.Status,
				"started_at":         paidExam.StartedAt,
				"expires_at":         paidExam.ExpiresAt,
				"time_limit_minutes": tier.TimeLimitMinutes,
				"question_count":     tier.QuestionCount,
				"practical_count":    tier.PracticalCount,
				"already_paid":       true,
			},
		})
		return
	}

	// Enforce exam attempt limits from tier metadata (default max_attempts = 3)
	maxAttempts := 3
	if tier.Metadata != nil {
		if v, ok := tier.Metadata["max_attempts"].(float64); ok && v > 0 {
			maxAttempts = int(v)
		}
	}
	completedCount, countErr := h.repo.GetCompletedExamCountForUserTier(r.Context(), claims.UserID, tier.ID)
	if countErr != nil {
		logrus.WithError(countErr).Error("Failed to count completed exams for attempt limit")
		writeJSONError(w, http.StatusInternalServerError, "Failed to check exam eligibility")
		return
	}
	if completedCount >= maxAttempts {
		writeJSONError(w, http.StatusTooManyRequests, fmt.Sprintf("Maximum exam retakes (%d) reached for this tier. You must contact support for exceptions.", maxAttempts))
		return
	}

	// If there's a pending_payment exam, return its checkout URL so user can retry payment
	if err == nil && paidExam != nil && paidExam.Status == storage.CertExamStatusPendingPayment {
		user, userErr := h.userRepo.GetUserByID(claims.UserID)
		if userErr != nil {
			writeJSONError(w, http.StatusInternalServerError, "Failed to get user")
			return
		}
		stripeCustomerID, err := payment.GetOrCreateStripeCustomerForTenant(r.Context(), claims.TenantID, user.Email, user.Name)
		if err != nil {
			logrus.WithError(err).Error("Failed to get/create Stripe customer")
			writeJSONError(w, http.StatusInternalServerError, "Failed to create payment session")
			return
		}
		checkoutResp, err := payment.CreateCertExamCheckoutSessionSimple(
			r.Context(),
			stripeCustomerID,
			user.Email,
			payment.CreateCertExamCheckoutSessionRequest{
				ExamID:       paidExam.ID,
				TenantID:     claims.TenantID,
				UserID:       claims.UserID,
				TierSlug:     tierSlug,
				PriceCents:   tier.PriceCents,
				ProductName:  tier.Name + " Certification Exam",
				ProductDesc:  tier.Description,
			},
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to create checkout session for pending exam")
			writeJSONError(w, http.StatusInternalServerError, "Failed to create payment session")
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"exam": map[string]interface{}{
				"id":                 paidExam.ID,
				"tier":               map[string]interface{}{"slug": tier.Slug, "name": tier.Name},
				"status":             paidExam.Status,
				"started_at":         paidExam.StartedAt,
				"expires_at":         paidExam.ExpiresAt,
				"time_limit_minutes": tier.TimeLimitMinutes,
				"question_count":     tier.QuestionCount,
				"practical_count":    tier.PracticalCount,
				"already_paid":       false,
			},
			"checkout_url": checkoutResp.URL,
		})
		return
	}

	// If tier has a price, create checkout session
	if tier.PriceCents > 0 {
		logrus.WithFields(logrus.Fields{
			"tier_slug":    tierSlug,
			"price_cents": tier.PriceCents,
		}).Info("Starting paid exam flow - creating checkout session")

		user, userErr := h.userRepo.GetUserByID(claims.UserID)
		if userErr != nil {
			writeJSONError(w, http.StatusInternalServerError, "Failed to get user")
			return
		}

		// Create pending exam record first (without questions)
		now := time.Now()
		exam := &storage.CertExam{
			UserID:       claims.UserID,
			TierID:       tier.ID,
			Status:       storage.CertExamStatusPendingPayment,
			AmountCents:  tier.PriceCents,
			Currency:     tier.Currency,
			StartedAt:    now,
			ExpiresAt:    now.Add(time.Duration(tier.TimeLimitMinutes) * time.Minute),
			QuestionIDs:  []uuid.UUID{},
			PracticalIDs: []uuid.UUID{},
			Answers:      storage.JSONMap{},
			IPAddress:    strPtr(getClientIP(r)),
			UserAgent:    strPtr(r.UserAgent()),
			Metadata:     storage.JSONMap{},
		}

		if err := h.repo.CreateExam(r.Context(), exam); err != nil {
			logrus.WithError(err).Error("Failed to create pending exam")
			writeJSONError(w, http.StatusInternalServerError, "Failed to create exam session")
			return
		}

		// Create Stripe checkout session
		stripeCustomerID, err := payment.GetOrCreateStripeCustomerForTenant(r.Context(), claims.TenantID, user.Email, user.Name)
		if err != nil {
			logrus.WithError(err).Error("Failed to get/create Stripe customer")
			writeJSONError(w, http.StatusInternalServerError, "Failed to create payment session")
			return
		}

		checkoutResp, err := payment.CreateCertExamCheckoutSessionSimple(
			r.Context(),
			stripeCustomerID,
			user.Email,
			payment.CreateCertExamCheckoutSessionRequest{
				TenantID:    claims.TenantID,
				UserID:      claims.UserID,
				TierSlug:    tier.Slug,
				ExamID:      exam.ID,
				PriceCents:  tier.PriceCents,
				ProductName: tier.Name + " Certification Exam",
				ProductDesc: "One-time exam fee for " + tier.Name + " certification. Includes " + strconv.Itoa(tier.QuestionCount) + " questions and " + strconv.Itoa(tier.PracticalCount) + " practical challenges.",
			},
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to create checkout session")
			writeJSONError(w, http.StatusInternalServerError, "Failed to create payment session")
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"exam": map[string]interface{}{
				"id":            exam.ID,
				"tier":          map[string]interface{}{"slug": tier.Slug, "name": tier.Name},
				"status":        exam.Status,
				"amount_cents":  exam.AmountCents,
				"currency":      exam.Currency,
				"checkout_url":  checkoutResp.URL,
				"session_id":    checkoutResp.SessionID,
			},
			"requires_payment": true,
		})
		return
	}

	// Free tier - create exam directly with questions
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
		QuestionIDs:  questionIDs,
		PracticalIDs: practicalIDs,
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
	questionIDs := exam.QuestionIDs
	questions, err := h.repo.GetQuestionsByIDs(r.Context(), questionIDs)
	if err != nil {
		logrus.WithError(err).Error("Failed to get questions")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve questions")
		return
	}

	// Get practical challenge details
	practicalIDs := exam.PracticalIDs
	var challenges []map[string]interface{}
	for _, cid := range practicalIDs {
		ch, err := h.repo.GetChallengeByID(r.Context(), cid)
		if err != nil {
			logrus.WithError(err).Error("Failed to get challenge")
			continue
		}
		if ch != nil {
			challengeMap := map[string]interface{}{
				"id":                 ch.ID,
				"name":               ch.Name,
				"description":        ch.Description,
				"category":           ch.Category,
				"difficulty":         ch.Difficulty,
				"points":             ch.Points,
				"time_limit_minutes": ch.TimeLimitMinutes,
			}
			if ch.EnvironmentURL != "" {
				challengeMap["environment_url"] = ch.EnvironmentURL
			}
			challenges = append(challenges, challengeMap)
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

	var req struct {
		Answers map[string]interface{} `json:"answers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.Answers != nil {
		for k, v := range req.Answers {
			exam.Answers[k] = v
		}
	}

	if err := h.repo.SubmitExam(r.Context(), examID, exam.Answers); err != nil {
		logrus.WithError(err).Error("Failed to submit exam")
		writeJSONError(w, http.StatusInternalServerError, "Failed to submit exam")
		return
	}

	go h.gradeKnowledgeQuestions(context.Background(), examID, exam)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "submitted",
		"message": "Exam submitted successfully.",
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
	if exam.Status == storage.CertExamStatusPassed || exam.Status == storage.CertExamStatusFailed ||
		exam.Status == storage.CertExamStatusExpired || exam.Status == storage.CertExamStatusAbandoned {
		writeJSONError(w, http.StatusConflict, "Exam cannot be abandoned — already completed")
		return
	}

	if err := h.repo.AbandonExam(r.Context(), examID); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "completed") {
			writeJSONError(w, http.StatusConflict, "Exam cannot be abandoned — already completed")
			return
		}
		logrus.WithError(err).Error("Failed to abandon exam")
		writeJSONError(w, http.StatusInternalServerError, "Failed to abandon exam")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "abandoned",
		"message": "Exam abandoned successfully. This attempt has been counted.",
	})
}

// gradeKnowledgeQuestions grades the knowledge portion of an exam asynchronously,
// enqueues practical challenges, and issues a credential if the candidate passes.
func (h *Handler) gradeKnowledgeQuestions(ctx context.Context, examID uuid.UUID, exam *storage.CertExam) {
	questionIDs := exam.QuestionIDs
	correctAnswers, err := h.repo.GetCorrectAnswers(ctx, questionIDs)
	if err != nil {
		logrus.WithError(err).WithField("exam_id", examID).Error("Failed to get correct answers for grading")
		if gradeErr := h.repo.GradeExam(ctx, examID, 0, 0, 0, false); gradeErr != nil {
			logrus.WithError(gradeErr).Warn("Failed to mark exam as failed")
		}
		return
	}

	totalPoints := 0
	earnedPoints := 0

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

	passedThreshold := 70.0
	if exam.Tier != nil {
		passedThreshold = exam.Tier.PassThreshold
	}
	knowledgePassed := knowledgeScore >= passedThreshold

	logrus.WithFields(logrus.Fields{
		"exam_id":         examID,
		"knowledge_score": knowledgeScore,
		"earned_points":   earnedPoints,
		"total_points":    totalPoints,
	}).Info("Knowledge questions graded")

	tier, err := h.repo.GetTierByID(ctx, exam.TierID)
	if err != nil {
		logrus.WithError(err).WithField("exam_id", examID).Error("Failed to get tier for grading")
		return
	}

	if len(exam.PracticalIDs) == 0 {
		passed := knowledgePassed
		totalScore := knowledgeScore
		if err := h.repo.GradeExam(ctx, examID, knowledgeScore, 0, totalScore, passed); err != nil {
			logrus.WithError(err).WithField("exam_id", examID).Error("Failed to persist exam grade")
			return
		}
		if passed {
			h.issueCredential(ctx, exam, tier, knowledgeScore)
		}
		return
	}

	if err := h.repo.GradeExam(ctx, examID, knowledgeScore, 0, knowledgeScore, false); err != nil {
		logrus.WithError(err).WithField("exam_id", examID).Error("Failed to persist knowledge grade")
		return
	}

	for _, challengeID := range exam.PracticalIDs {
		if err := h.repo.EnqueueGrading(ctx, examID, challengeID); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"exam_id":      examID,
				"challenge_id": challengeID,
			}).Error("Failed to enqueue practical challenge for grading")
		}
	}

	h.waitForPracticalGrading(ctx, examID, tier, knowledgeScore, knowledgePassed)
}

// waitForPracticalGrading polls until all practical grading queue items for an exam are resolved,
// then computes the final grade and issues a credential if the candidate passed.
func (h *Handler) waitForPracticalGrading(ctx context.Context, examID uuid.UUID, tier *storage.CertTier, knowledgeScore float64, knowledgePassed bool) {
	timeout := time.After(30 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	exam, err := h.repo.GetExamByID(ctx, examID)
	if err != nil {
		logrus.WithError(err).WithField("exam_id", examID).Error("Failed to reload exam for practical grading")
		return
	}

	practicalCount := len(exam.PracticalIDs)
	practicalMaxPoints := 0
	for _, challengeID := range exam.PracticalIDs {
		challenge, chErr := h.repo.GetChallengeByID(ctx, challengeID)
		if chErr == nil && challenge != nil {
			practicalMaxPoints += challenge.Points
		}
	}
	if practicalMaxPoints == 0 {
		practicalMaxPoints = practicalCount * 10
	}

	knowledgeWeight := 0.6
	practicalWeight := 0.4

	for {
		select {
		case <-ctx.Done():
			logrus.WithField("exam_id", examID).Info("Grading cancelled due to context shutdown")
			return
		case <-timeout:
			logrus.WithField("exam_id", examID).Warn("Practical grading timed out after 30 minutes")
			if gradeErr := h.repo.GradeExam(ctx, examID, knowledgeScore, 0, knowledgeScore, false); gradeErr != nil {
				logrus.WithError(gradeErr).Warn("Failed to update timed-out exam grade")
			}
			return
		case <-ticker.C:
			pendingCount, countErr := h.countPendingGrading(ctx, examID)
			if countErr != nil {
				logrus.WithError(countErr).WithField("exam_id", examID).Warn("Failed to count pending grading items")
				continue
			}
			if pendingCount > 0 {
				continue
			}

			completedEarned := 0
			completedTotal := 0
			for _, challengeID := range exam.PracticalIDs {
				result, resultErr := h.repo.GetGradingResult(ctx, examID, challengeID)
				if resultErr != nil || result == nil {
					completedTotal += 10
					continue
				}
				scoreVal := 0
				if sv, ok := result["score"].(float64); ok {
					scoreVal = int(sv)
				}
				challenge, chErr := h.repo.GetChallengeByID(ctx, challengeID)
				if chErr == nil && challenge != nil {
					completedTotal += challenge.Points
					completedEarned += int(float64(scoreVal) / 100.0 * float64(challenge.Points))
				} else {
					completedTotal += 10
					completedEarned += int(float64(scoreVal) / 100.0 * 10.0)
				}
			}

			practicalScore := 0.0
			if completedTotal > 0 {
				practicalScore = (float64(completedEarned) / float64(completedTotal)) * 100
			}

			totalScore := knowledgeWeight*knowledgeScore + practicalWeight*practicalScore
			passed := knowledgePassed && practicalScore >= 50.0
			if tier != nil {
				passedThreshold := tier.PassThreshold
				passed = totalScore >= passedThreshold
			}

			logrus.WithFields(logrus.Fields{
				"exam_id":         examID,
				"knowledge_score": knowledgeScore,
				"practical_score": practicalScore,
				"total_score":    totalScore,
				"passed":         passed,
			}).Info("Final exam grade computed")

			if err := h.repo.GradeExam(ctx, examID, knowledgeScore, practicalScore, totalScore, passed); err != nil {
				logrus.WithError(err).WithField("exam_id", examID).Error("Failed to persist final exam grade")
				return
			}

			if passed {
				h.issueCredential(ctx, exam, tier, totalScore)
			}
			return
		}
	}
}

// countPendingGrading returns the number of pending/processing grading queue items for an exam.
func (h *Handler) countPendingGrading(ctx context.Context, examID uuid.UUID) (int, error) {
	return h.repo.CountPendingGrading(ctx, examID)
}

// issueCredential creates a credential record for a passed exam and sends a notification.
func (h *Handler) issueCredential(ctx context.Context, exam *storage.CertExam, tier *storage.CertTier, score float64) {
	existingCred, err := h.repo.GetActiveCredentialByUserTier(ctx, exam.UserID, exam.TierID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id": exam.UserID,
			"tier_id": exam.TierID,
		}).Warn("Failed to check for existing credential")
	}
	if existingCred != nil {
		logrus.WithFields(logrus.Fields{
			"exam_id":  exam.ID,
			"cred_id":  existingCred.ID,
		}).Info("Candidate already has active credential for this tier, skipping issuance")
		return
	}

	credNumber, err := h.repo.NextCredentialNumber(ctx, tier.Slug)
	if err != nil {
		logrus.WithError(err).WithField("tier_slug", tier.Slug).Error("Failed to generate credential number")
		return
	}

	user, userErr := h.userRepo.GetUserByID(exam.UserID)
	if userErr != nil || user == nil {
		logrus.WithError(userErr).WithField("user_id", exam.UserID).Error("Failed to fetch user for credential issuance")
		return
	}
	userEmail := user.Email

	issuedAt := time.Now()
	expiresAt := issuedAt.AddDate(tier.ValidityMonths, 0, 0)

	obaCredential := map[string]interface{}{
		"@context":    []string{"https://w3id.org/openbadges/v2"},
		"type":        "Assertion",
		"id":          fmt.Sprintf("https://api.functionfly.io/v1/certification/credentials/%s/badge", credNumber),
		"badge": map[string]interface{}{
			"type": "BadgeClass",
			"id":   fmt.Sprintf("https://api.functionfly.io/v1/certification/badges/%s", tier.Slug),
			"name": tier.Name + " Certification",
			"description": tier.Description,
			"image":        fmt.Sprintf("https://api.functionfly.io/v1/certification/badges/%s/image", tier.Slug),
			"criteria":     fmt.Sprintf("https://functionfly.io/certification/%s", tier.Slug),
			"issuer": map[string]interface{}{
				"type": "Issuer",
				"id":   "https://functionfly.io",
				"name": "FunctionFly",
				"url":  "https://functionfly.io",
			},
		},
		"issuedOn": issuedAt.Format(time.RFC3339),
		"recipient": map[string]interface{}{
			"type": "email",
			"identity": userEmail,
			"hashed":   false,
		},
	}

	verificationHash := fmt.Sprintf("%x", sha256.Sum256([]byte(credNumber+exam.ID.String()+exam.UserID.String())))[:64]
	verificationURL := fmt.Sprintf("https://functionfly.io/certification/verify/number/%s", credNumber)

	cred := &storage.CertCredential{
		UserID:           exam.UserID,
		TierID:           exam.TierID,
		ExamID:           exam.ID,
		CredentialNumber: credNumber,
		Status:           storage.CertCredentialStatusActive,
		IssuedAt:         issuedAt,
		ExpiresAt:        expiresAt,
		OBACredential:    obaCredential,
		VerificationHash: verificationHash,
		VerificationURL:  &verificationURL,
		Metadata:         storage.JSONMap{"score": score},
	}

	if createErr := h.repo.CreateCredential(ctx, cred); createErr != nil {
		logrus.WithError(createErr).WithField("exam_id", exam.ID).Error("Failed to create credential")
		return
	}

	logrus.WithFields(logrus.Fields{
		"exam_id":         exam.ID,
		"credential_id":   cred.ID,
		"credential_number": credNumber,
		"tier_slug":       tier.Slug,
		"score":           score,
	}).Info("Credential issued successfully")
}

// fmt is already imported

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
