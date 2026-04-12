package privacy

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/auth/gba"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Context keys for privacy context
type privacyContextKey string

const (
	// PrivacySettingsContextKey stores privacy settings in context
	PrivacySettingsContextKey privacyContextKey = "privacy_settings"
	// PrivacyHeadersContextKey stores privacy headers in context
	PrivacyHeadersContextKey privacyContextKey = "privacy_headers"
)

// Middleware provides privacy middleware for HTTP handlers
type Middleware struct {
	service     *Service
	globalSettings *GlobalPrivacySettings
}

// NewMiddleware creates a new privacy middleware
func NewMiddleware(service *Service) *Middleware {
	return &Middleware{
		service: service,
	}
}

// Handler wraps an HTTP handler with privacy context
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract privacy headers
		privacyHeaders := m.extractPrivacyHeaders(r)

		// Add to context
		ctx := context.WithValue(r.Context(), PrivacyHeadersContextKey, privacyHeaders)

		// Try to get user privacy settings if authenticated
		userID := m.extractUserID(r)
		if userID != uuid.Nil {
			settings, err := m.service.GetPrivacySettings(userID)
			if err != nil {
				logrus.WithError(err).Warn("Failed to get user privacy settings")
			} else {
				ctx = context.WithValue(ctx, PrivacySettingsContextKey, settings)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ExecutionPrivacyMiddleware is specific middleware for execution endpoints
func (m *Middleware) ExecutionPrivacyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract privacy headers
		privacyHeaders := m.extractPrivacyHeaders(r)

		// Get global settings to check if consent is required
		globalSettings, err := m.service.repo.GetGlobalPrivacySettings()
		if err != nil {
			logrus.WithError(err).Warn("Failed to get global privacy settings")
		}

		// Check if consent is required
		if globalSettings != nil && globalSettings.RequireConsent {
			userID := m.extractUserID(r)
			if userID != uuid.Nil {
				// Check if user has given consent for execution logging
				if !m.service.HasActiveConsent(userID, "execution_logging") {
					// Add consent requirement header
					w.Header().Set("X-Consent-Required", "true")
					w.Header().Set("X-Consent-Type", "execution_logging")

					// In strict mode, we could reject the request
					// For now, just add warning headers
				}
			}
		}

		// Add privacy context
		ctx := context.WithValue(r.Context(), PrivacyHeadersContextKey, privacyHeaders)

		// Get user settings
		userID := m.extractUserID(r)
		if userID != uuid.Nil {
			settings, err := m.service.GetPrivacySettings(userID)
			if err != nil {
				logrus.WithError(err).Warn("Failed to get user privacy settings")
			} else {
				ctx = context.WithValue(ctx, PrivacySettingsContextKey, settings)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractPrivacyHeaders extracts privacy-related headers from request
func (m *Middleware) extractPrivacyHeaders(r *http.Request) *PrivacyHeaders {
	headers := make(map[string]string)

	// Extract relevant headers
	privacyHeaderNames := []string{
		"Dnt", "Gdpr", "Ccpa", "Consent", "Privacy-Level",
		"X-Do-Not-Track", "X-GDPR-Applies", "X-CCPA-Applies",
		"Sec-Gpc", // Global Privacy Control
	}

	for _, header := range privacyHeaderNames {
		if value := r.Header.Get(header); value != "" {
			headers[header] = value
		}
	}

	return m.service.GetPrivacyHeaders(headers)
}

// extractUserID extracts user ID from request context (from auth)
func (m *Middleware) extractUserID(r *http.Request) uuid.UUID {
	// Try to get from context (set by auth middleware)
	if userID, ok := r.Context().Value(gba.ContextKeyUserID).(string); ok {
		if id, err := uuid.Parse(userID); err == nil {
			return id
		}
	}

	// Try to get from auth context
	if claims, ok := r.Context().Value("claims").(*auth.Claims); ok {
		return claims.UserID
	}

	return uuid.Nil
}

// GetPrivacySettingsFromContext retrieves privacy settings from context
func GetPrivacySettingsFromContext(ctx context.Context) *PrivacySettings {
	if settings, ok := ctx.Value(PrivacySettingsContextKey).(*PrivacySettings); ok {
		return settings
	}
	return nil
}

// GetPrivacyHeadersFromContext retrieves privacy headers from context
func GetPrivacyHeadersFromContext(ctx context.Context) *PrivacyHeaders {
	if headers, ok := ctx.Value(PrivacyHeadersContextKey).(*PrivacyHeaders); ok {
		return headers
	}
	return nil
}

// ShouldAnonymizeIP checks if IP should be anonymized based on context
func ShouldAnonymizeIP(ctx context.Context) bool {
	// Check privacy settings
	if settings := GetPrivacySettingsFromContext(ctx); settings != nil {
		if settings.AnonymizeIP || settings.PrivacyLevel == PrivacyLevelMaximum || settings.PrivacyLevel == PrivacyLevelGDPR {
			return true
		}
	}

	// Check privacy headers
	if headers := GetPrivacyHeadersFromContext(ctx); headers != nil {
		if headers.DoNotTrack || headers.RequestAnonymization {
			return true
		}
	}

	return false
}

// ShouldAnonymizeUserAgent checks if user agent should be anonymized
func ShouldAnonymizeUserAgent(ctx context.Context) bool {
	if settings := GetPrivacySettingsFromContext(ctx); settings != nil {
		if settings.AnonymizeUserAgent || settings.PrivacyLevel == PrivacyLevelMaximum || settings.PrivacyLevel == PrivacyLevelGDPR {
			return true
		}
	}
	return false
}

// GetIPMaskType returns the IP mask type from context
func GetIPMaskType(ctx context.Context) PIIMaskType {
	if settings := GetPrivacySettingsFromContext(ctx); settings != nil {
		return settings.IPMaskType
	}
	return PIIMaskTypeNone
}

// GetUserAgentMaskType returns the user agent mask type from context
func GetUserAgentMaskType(ctx context.Context) PIIMaskType {
	if settings := GetPrivacySettingsFromContext(ctx); settings != nil {
		return settings.UserAgentMaskType
	}
	return PIIMaskTypeNone
}

// PrivacyResponseWriter wraps http.ResponseWriter to add privacy headers
type PrivacyResponseWriter struct {
	http.ResponseWriter
	privacyHeaders *PrivacyHeaders
}

// NewPrivacyResponseWriter creates a new privacy response writer
func NewPrivacyResponseWriter(w http.ResponseWriter, headers *PrivacyHeaders) *PrivacyResponseWriter {
	return &PrivacyResponseWriter{
		ResponseWriter: w,
		privacyHeaders: headers,
	}
}

// WriteHeader adds privacy headers before writing status
func (prw *PrivacyResponseWriter) WriteHeader(code int) {
	// Add privacy-related response headers
	if prw.privacyHeaders != nil {
		if prw.privacyHeaders.DoNotTrack {
			prw.ResponseWriter.Header().Set("Tk", "N") // Not tracking
		}

		if prw.privacyHeaders.GDPRApplies {
			prw.ResponseWriter.Header().Set("X-GDPR-Applies", "true")
		}

		if prw.privacyHeaders.CCPAApplies {
			prw.ResponseWriter.Header().Set("X-CCPA-Applies", "true")
		}
	}

	prw.ResponseWriter.WriteHeader(code)
}

// ConsentMiddleware handles consent requirements
type ConsentMiddleware struct {
	service *Service
}

// NewConsentMiddleware creates a new consent middleware
func NewConsentMiddleware(service *Service) *ConsentMiddleware {
	return &ConsentMiddleware{service: service}
}

// RequireConsent returns middleware that requires consent for specific operations
func (cm *ConsentMiddleware) RequireConsent(consentType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := cm.extractUserID(r)
			if userID == uuid.Nil {
				next.ServeHTTP(w, r)
				return
			}

			if !cm.service.HasActiveConsent(userID, consentType) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"error":"Consent required","message":"User consent is required for this operation","consent_type":"` + consentType + `"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractUserID extracts user ID from request
func (cm *ConsentMiddleware) extractUserID(r *http.Request) uuid.UUID {
	if userID, ok := r.Context().Value(gba.ContextKeyUserID).(string); ok {
		if id, err := uuid.Parse(userID); err == nil {
			return id
		}
	}
	return uuid.Nil
}

// IPAnonymizationMiddleware anonymizes IP addresses in requests
func IPAnonymizationMiddleware(anonymizer *Anonymizer, maskType PIIMaskType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Anonymize IP in headers if needed
			if maskType != PIIMaskTypeNone {
				// Note: We can't easily modify r.RemoteAddr, but we can set context
				ctx := context.WithValue(r.Context(), "anonymize_ip", true)
				ctx = context.WithValue(ctx, "ip_mask_type", maskType)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetClientIPAnonymized gets the client IP, anonymized if required
func GetClientIPAnonymized(r *http.Request, anonymizer *Anonymizer, maskType PIIMaskType) string {
	// Get original IP
	ip := getClientIP(r)

	// Anonymize if needed
	if maskType != PIIMaskTypeNone {
		return anonymizer.AnonymizeIP(ip, maskType)
	}

	return ip
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-Ip header
	xri := r.Header.Get("X-Real-Ip")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

