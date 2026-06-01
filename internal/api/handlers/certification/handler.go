package certification

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Handler handles certification API endpoints
type Handler struct {
	repo              *storage.CertificationRepository
	userRepo          *storage.UserRepository
	verificationLimiter *CertVerificationRateLimiter
}

// CertVerificationRateLimiter applies per-IP rate limiting to public verification endpoints.
type CertVerificationRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	window   time.Duration
	limit    int
}

// NewCertVerificationRateLimiter creates a rate limiter allowing 30 requests per minute per IP.
func NewCertVerificationRateLimiter() *CertVerificationRateLimiter {
	rl := &CertVerificationRateLimiter{
		requests: make(map[string][]time.Time),
		window:   time.Minute,
		limit:    30,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *CertVerificationRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.window)
		for ip, times := range rl.requests {
			filtered := times[:0]
			for _, t := range times {
				if t.After(cutoff) {
					filtered = append(filtered, t)
				}
			}
			if len(filtered) == 0 {
				delete(rl.requests, ip)
			} else {
				rl.requests[ip] = filtered
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *CertVerificationRateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	var recent []time.Time
	for _, t := range rl.requests[ip] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	recent = append(recent, now)
	rl.requests[ip] = recent

	return len(recent) <= rl.limit
}

// Limit wraps a handler with cert verification rate limiting.
func (rl *CertVerificationRateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if ip == "" {
			ip = r.RemoteAddr
		}
		if !rl.Allow(ip) {
			w.Header().Set("X-RateLimit-Limit", "30")
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Retry-After", "60")
			writeJSONError(w, http.StatusTooManyRequests, "Rate limit exceeded. Try again later.")
			return
		}
		next.ServeHTTP(w, r)
	}
}

// NewHandler creates a new certification handler
func NewHandler(repo *storage.CertificationRepository, userRepo *storage.UserRepository) *Handler {
	return &Handler{
		repo:              repo,
		userRepo:          userRepo,
		verificationLimiter: NewCertVerificationRateLimiter(),
	}
}

// RegisterRoutes registers all certification routes on the given router
func (h *Handler) RegisterRoutes(api *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	cert := api.PathPrefix("/certification").Subrouter()

	cert.HandleFunc("/tiers", h.ListTiers).Methods("GET", "OPTIONS")
	cert.HandleFunc("/verify/{username}", h.verificationLimiter.Limit(h.VerifyCredential)).Methods("GET", "OPTIONS")
	cert.HandleFunc("/verify/number/{credentialNumber}", h.verificationLimiter.Limit(h.VerifyByNumber)).Methods("GET", "OPTIONS")
	cert.HandleFunc("/credentials/{username}/badges", h.verificationLimiter.Limit(h.PublicBadges)).Methods("GET", "OPTIONS")

	cert.HandleFunc("/tiers/{tierSlug}/start", authMiddleware.RequireAuth(h.StartExam)).Methods("POST", "OPTIONS")
	cert.HandleFunc("/exams/{examId}", authMiddleware.RequireAuth(h.GetExam)).Methods("GET", "OPTIONS")
	cert.HandleFunc("/exams/{examId}/answer", authMiddleware.RequireAuth(h.SubmitAnswer)).Methods("PUT", "OPTIONS")
	cert.HandleFunc("/exams/{examId}/submit", authMiddleware.RequireAuth(h.SubmitExam)).Methods("POST", "OPTIONS")
	cert.HandleFunc("/exams/{examId}/abandon", authMiddleware.RequireAuth(h.AbandonExam)).Methods("POST", "OPTIONS")
	cert.HandleFunc("/exams/{examId}/practical/{challengeId}/submit", authMiddleware.RequireAuth(h.SubmitPractical)).Methods("PUT", "OPTIONS")
	cert.HandleFunc("/exams", authMiddleware.RequireAuth(h.ListMyExams)).Methods("GET", "OPTIONS")

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
