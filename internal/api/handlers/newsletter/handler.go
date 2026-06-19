package newsletter

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"

	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/functionfly/functionfly/internal/apierror"
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	emailAddr := strings.ToLower(strings.TrimSpace(req.Email))
	if emailAddr == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Email is required"))
		return
	}

	if !isValidEmail(emailAddr) {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid email format"))
		return
	}

	name := strings.TrimSpace(req.Name)
	source := "landing_page"
	ipAddress := getClientIP(r)
	userAgent := r.UserAgent()

	confirmationToken, err := generateConfirmationToken()
	if err != nil {
		logrus.WithError(err).Error("Failed to generate confirmation token")
		apierror.WriteError(w, apierror.NewInternal("Failed to process subscription"))
		return
	}

	_, err = h.repo.CreatePendingNewsletterSubscriber(r.Context(), emailAddr, name, source, ipAddress, userAgent, confirmationToken)
	if err != nil {
		if err == storage.ErrSubscriberExists {
			apierror.WriteError(w, apierror.NewConflict("Email is already subscribed"))
			return
		}
		logrus.WithError(err).Error("Failed to create newsletter subscriber")
		apierror.WriteError(w, apierror.NewInternal("Failed to subscribe"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	token := strings.TrimSpace(req.Token)
	emailAddr := strings.ToLower(strings.TrimSpace(req.Email))

	if token == "" || emailAddr == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Token and email are required"))
		return
	}

	if !isValidEmail(emailAddr) {
		apierror.WriteError(w, apierror.NewBadRequest("Invalid email format"))
		return
	}

	subscriber, err := h.repo.GetNewsletterSubscriberByEmail(r.Context(), emailAddr)
	if err != nil {
		if err == storage.ErrSubscriberNotFound {
			apierror.WriteError(w, apierror.NewNotFound("Subscriber not found"))
			return
		}
		logrus.WithError(err).Error("Failed to get newsletter subscriber")
		apierror.WriteError(w, apierror.NewInternal("Failed to confirm subscription"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid subscription status"))
		return
	}

	if subscriber.ConfirmationToken == nil || *subscriber.ConfirmationToken != token {
		apierror.WriteError(w, apierror.NewUnauthorized("Invalid confirmation token"))
		return
	}

	if err := h.repo.ConfirmNewsletterSubscription(r.Context(), emailAddr); err != nil {
		logrus.WithError(err).Error("Failed to confirm newsletter subscription")
		apierror.WriteError(w, apierror.NewInternal("Failed to confirm subscription"))
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
		apierror.WriteError(w, apierror.NewBadRequest("Invalid request body"))
		return
	}

	emailAddr := strings.ToLower(strings.TrimSpace(req.Email))
	if emailAddr == "" {
		apierror.WriteError(w, apierror.NewBadRequest("Email is required"))
		return
	}

	if err := h.repo.UnsubscribeNewsletterSubscriber(r.Context(), emailAddr); err != nil {
		if err == storage.ErrSubscriberNotFound {
			apierror.WriteError(w, apierror.NewNotFound("Subscriber not found"))
			return
		}
		logrus.WithError(err).Error("Failed to unsubscribe newsletter subscriber")
		apierror.WriteError(w, apierror.NewInternal("Failed to unsubscribe"))
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
