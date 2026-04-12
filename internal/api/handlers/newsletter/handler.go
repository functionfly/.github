package newsletter

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo         storage.Repository
	emailService email.Service
}

func NewHandler(repo storage.Repository, emailService email.Service) *Handler {
	return &Handler{
		repo:         repo,
		emailService: emailService,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/newsletter/subscribe", h.Subscribe).Methods("POST", "OPTIONS")
	router.HandleFunc("/newsletter/unsubscribe", h.Unsubscribe).Methods("POST", "OPTIONS")
}

type SubscribeRequest struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	var req SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	emailAddr := strings.ToLower(strings.TrimSpace(req.Email))
	if emailAddr == "" {
		http.Error(w, `{"error": "Email is required"}`, http.StatusBadRequest)
		return
	}

	if !isValidEmail(emailAddr) {
		http.Error(w, `{"error": "Invalid email format"}`, http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.Name)
	source := "landing_page"
	ipAddress := getClientIP(r)
	userAgent := r.UserAgent()

	subscriber, err := h.repo.CreateNewsletterSubscriber(r.Context(), emailAddr, name, source, ipAddress, userAgent)
	if err != nil {
		if err == storage.ErrSubscriberExists {
			http.Error(w, `{"error": "Email is already subscribed"}`, http.StatusConflict)
			return
		}
		logrus.WithError(err).Error("Failed to create newsletter subscriber")
		http.Error(w, `{"error": "Failed to subscribe"}`, http.StatusInternalServerError)
		return
	}

	if err := h.emailService.SendNewsletterSubscriptionConfirmation(emailAddr, name); err != nil {
		logrus.WithError(err).Warn("Failed to send newsletter subscription confirmation email")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"message":    "Successfully subscribed to newsletter",
		"subscriber": subscriber,
	})
}

type UnsubscribeRequest struct {
	Email string `json:"email"`
}

func (h *Handler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	var req UnsubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	emailAddr := strings.ToLower(strings.TrimSpace(req.Email))
	if emailAddr == "" {
		http.Error(w, `{"error": "Email is required"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.UnsubscribeNewsletterSubscriber(r.Context(), emailAddr); err != nil {
		if err == storage.ErrSubscriberNotFound {
			http.Error(w, `{"error": "Subscriber not found"}`, http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to unsubscribe newsletter subscriber")
		http.Error(w, `{"error": "Failed to unsubscribe"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Successfully unsubscribed from newsletter",
	})
}

func isValidEmail(email string) bool {
	atIndex := strings.LastIndex(email, "@")
	if atIndex <= 0 || atIndex == len(email)-1 {
		return false
	}
	domain := email[atIndex+1:]
	if !strings.Contains(domain, ".") {
		return false
	}
	return true
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}
	forwarded = r.Header.Get("X-Real-IP")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}
