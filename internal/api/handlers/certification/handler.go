package certification

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles certification API endpoints
type Handler struct {
	repo      *storage.CertificationRepository
	userRepo  *storage.UserRepository
}

// NewHandler creates a new certification handler
func NewHandler(repo *storage.CertificationRepository, userRepo *storage.UserRepository) *Handler {
	return &Handler{
		repo:     repo,
		userRepo: userRepo,
	}
}

// RegisterRoutes registers all certification routes on the given router
func (h *Handler) RegisterRoutes(api *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	cert := api.PathPrefix("/certification").Subrouter()

	// Public endpoints (no auth required)
	cert.HandleFunc("/tiers", h.ListTiers).Methods("GET", "OPTIONS")
	cert.HandleFunc("/verify/{username}", h.VerifyCredential).Methods("GET", "OPTIONS")
	cert.HandleFunc("/verify/number/{credentialNumber}", h.VerifyByNumber).Methods("GET", "OPTIONS")
	cert.HandleFunc("/credentials/{username}/badges", h.PublicBadges).Methods("GET", "OPTIONS")

	// Exams (authenticated)
	cert.HandleFunc("/tiers/{tierSlug}/start", authMiddleware.RequireAuth(h.StartExam)).Methods("POST", "OPTIONS")
	cert.HandleFunc("/exams/{examId}", authMiddleware.RequireAuth(h.GetExam)).Methods("GET", "OPTIONS")
	cert.HandleFunc("/exams/{examId}/answer", authMiddleware.RequireAuth(h.SubmitAnswer)).Methods("PUT", "OPTIONS")
	cert.HandleFunc("/exams/{examId}/submit", authMiddleware.RequireAuth(h.SubmitExam)).Methods("POST", "OPTIONS")
	cert.HandleFunc("/exams/{examId}/abandon", authMiddleware.RequireAuth(h.AbandonExam)).Methods("POST", "OPTIONS")
	cert.HandleFunc("/exams", authMiddleware.RequireAuth(h.ListMyExams)).Methods("GET", "OPTIONS")

	// Credentials (authenticated)
	cert.HandleFunc("/credentials", authMiddleware.RequireAuth(h.ListMyCredentials)).Methods("GET", "OPTIONS")
	cert.HandleFunc("/credentials/{credentialId}", authMiddleware.RequireAuth(h.GetCredential)).Methods("GET", "OPTIONS")
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logrus.WithError(err).Error("Failed to encode JSON response")
	}
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for i, c := range xff {
			if c == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
