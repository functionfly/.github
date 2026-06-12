package newsletter

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo         storage.Repository
	emailService email.Service
	baseURL      string
}

func NewHandler(repo storage.Repository, emailService email.Service, baseURL string) *Handler {
	return &Handler{
		repo:         repo,
		emailService: emailService,
		baseURL:      baseURL,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/newsletter/subscribe", h.Subscribe).Methods("POST", "OPTIONS")
	router.HandleFunc("/newsletter/unsubscribe", h.Unsubscribe).Methods("POST", "OPTIONS")
	router.HandleFunc("/newsletter/confirm", h.ConfirmSubscription).Methods("POST", "OPTIONS")
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

	confirmationToken, err := generateConfirmationToken()
	if err != nil {
		logrus.WithError(err).Error("Failed to generate confirmation token")
		http.Error(w, `{"error": "Failed to process subscription"}`, http.StatusInternalServerError)
		return
	}

	subscriber, err := h.repo.CreatePendingNewsletterSubscriber(r.Context(), emailAddr, name, source, ipAddress, userAgent, confirmationToken)
	if err != nil {
		if err == storage.ErrSubscriberExists {
			http.Error(w, `{"error": "Email is already subscribed"}`, http.StatusConflict)
			return
		}
		logrus.WithError(err).Error("Failed to create newsletter subscriber")
		http.Error(w, `{"error": "Failed to subscribe"}`, http.StatusInternalServerError)
		return
	}

	confirmationURL := buildConfirmationURL(h.baseURL, confirmationToken, emailAddr)
	if err := h.emailService.SendNewsletterConfirmationEmail(emailAddr, name, confirmationURL); err != nil {
		logrus.WithError(err).Warn("Failed to send newsletter confirmation email")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Please check your email to confirm your subscription",
	})
}

type ConfirmRequest struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

func (h *Handler) ConfirmSubscription(w http.ResponseWriter, r *http.Request) {
	var req ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	token := strings.TrimSpace(req.Token)
	emailAddr := strings.ToLower(strings.TrimSpace(req.Email))

	if token == "" || emailAddr == "" {
		http.Error(w, `{"error": "Token and email are required"}`, http.StatusBadRequest)
		return
	}

	if !isValidEmail(emailAddr) {
		http.Error(w, `{"error": "Invalid email format"}`, http.StatusBadRequest)
		return
	}

	subscriber, err := h.repo.GetNewsletterSubscriberByEmail(r.Context(), emailAddr)
	if err != nil {
		if err == storage.ErrSubscriberNotFound {
			http.Error(w, `{"error": "Subscriber not found"}`, http.StatusNotFound)
			return
		}
		logrus.WithError(err).Error("Failed to get newsletter subscriber")
		http.Error(w, `{"error": "Failed to confirm subscription"}`, http.StatusInternalServerError)
		return
	}

	if subscriber.Status == "active" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Email is already confirmed",
		})
		return
	}

	if subscriber.Status != "pending" {
		http.Error(w, `{"error": "Invalid subscription status"}`, http.StatusBadRequest)
		return
	}

	if subscriber.ConfirmationToken == nil || *subscriber.ConfirmationToken != token {
		http.Error(w, `{"error": "Invalid confirmation token"}`, http.StatusUnauthorized)
		return
	}

	if err := h.repo.ConfirmNewsletterSubscription(r.Context(), emailAddr); err != nil {
		logrus.WithError(err).Error("Failed to confirm newsletter subscription")
		http.Error(w, `{"error": "Failed to confirm subscription"}`, http.StatusInternalServerError)
		return
	}

	if err := h.emailService.SendNewsletterSubscriptionConfirmation(emailAddr, subscriber.Name); err != nil {
		logrus.WithError(err).Warn("Failed to send welcome email after confirmation")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Successfully confirmed your subscription",
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

func isValidEmail(emailStr string) bool {
	_, err := mail.ParseAddress(emailStr)
	return err == nil
}

func generateConfirmationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func buildConfirmationURL(baseURL, token, email string) string {
	separator := "?"
	if strings.Contains(baseURL, "?") {
		separator = "&"
	}
	return strings.TrimSuffix(baseURL, "/") + separator + "token=" + token + "&email=" + email
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
