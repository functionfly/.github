package certification

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// SubmitPractical handles PUT /v1/certification/exams/{examId}/practical/{challengeId}/submit
// Saves the candidate's practical results (deployment URL, outputs, etc.) for grading.
func (h *Handler) SubmitPractical(w http.ResponseWriter, r *http.Request) {
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
	challengeID, err := uuid.Parse(mux.Vars(r)["challengeId"])
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid challenge ID")
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
	if exam.Status != storage.CertExamStatusInProgress && exam.Status != storage.CertExamStatusSubmitted {
		writeJSONError(w, http.StatusBadRequest, "Exam is not in progress")
		return
	}

	challenge, err := h.repo.GetChallengeByID(r.Context(), challengeID)
	if err != nil || challenge == nil {
		writeJSONError(w, http.StatusNotFound, "Challenge not found")
		return
	}

	var results storage.JSONMap
	if err := json.NewDecoder(r.Body).Decode(&results); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if results == nil {
		results = storage.JSONMap{}
	}
	results["challenge_id"] = challengeID.String()
	results["submitted_at"] = time.Now().UTC().Format(time.RFC3339)

	if exam.PracticalResults == nil {
		exam.PracticalResults = storage.JSONMap{}
	}
	exam.PracticalResults[challengeID.String()] = results

	if err := h.repo.SubmitExam(r.Context(), examID, map[string]interface{}{
		"practical_results": exam.PracticalResults,
		"updated_at":        time.Now().UTC(),
	}); err != nil {
		logrus.WithError(err).Error("Failed to save practical results")
		writeJSONError(w, http.StatusInternalServerError, "Failed to save practical results")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":       "submitted",
		"exam_id":      examID,
		"challenge_id": challengeID,
	})
}
