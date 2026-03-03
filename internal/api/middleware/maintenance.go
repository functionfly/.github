package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"hash/fnv"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/functionfly/functionfly/internal/types"
	"github.com/sirupsen/logrus"
)

// MaintenanceMiddleware handles platform-wide maintenance mode
type MaintenanceMiddleware struct {
	maintenanceRepo *storage.MaintenanceRepository
	cache           *MaintenanceCache
	logger          *logrus.Logger
}

// MaintenanceCache provides in-memory caching for maintenance state
type MaintenanceCache struct {
	enabled  bool
	config   *types.MaintenanceConfig
	ttl      time.Time
	cacheTTL time.Duration
}

// NewMaintenanceMiddleware creates a new maintenance middleware
func NewMaintenanceMiddleware(maintenanceRepo *storage.MaintenanceRepository) *MaintenanceMiddleware {
	return &MaintenanceMiddleware{
		maintenanceRepo: maintenanceRepo,
		cache: &MaintenanceCache{
			cacheTTL: 10 * time.Second,
		},
		logger: logrus.New(),
	}
}

// CheckMaintenanceMode returns middleware that checks for maintenance mode
func (m *MaintenanceMiddleware) CheckMaintenanceMode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip maintenance check for health endpoints
		if strings.HasPrefix(r.URL.Path, "/health") ||
			strings.HasPrefix(r.URL.Path, "/maintenance/status") ||
			strings.HasPrefix(r.URL.Path, "/v1/admin/maintenance") {
			next.ServeHTTP(w, r)
			return
		}

		// Check if maintenance mode is enabled
		config, err := m.getMaintenanceConfig(r.Context())
		if err != nil {
			m.logger.WithError(err).Warn("Failed to check maintenance mode - failing open")
			next.ServeHTTP(w, r)
			return
		}

		// If not enabled, continue
		if config == nil || !config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Check if scheduled maintenance should be active
		if config.IsScheduled && config.ScheduledStart != nil && config.ScheduledEnd != nil {
			now := time.Now()
			if now.Before(*config.ScheduledStart) || now.After(*config.ScheduledEnd) {
				// Outside scheduled window, continue normally
				next.ServeHTTP(w, r)
				return
			}
		}

		// Check rollout percentage for gradual enablement
		if config.RolloutPercentage < 100 {
			if !m.shouldShowMaintenance(r, config.RolloutPercentage) {
				next.ServeHTTP(w, r)
				return
			}
		}

		// Serve maintenance page
		m.serveMaintenancePage(w, r, config)
	})
}

// getMaintenanceConfig gets the current maintenance configuration (with caching)
func (m *MaintenanceMiddleware) getMaintenanceConfig(ctx context.Context) (*types.MaintenanceConfig, error) {
	// Check cache first
	if m.cache.ttl.After(time.Now()) && m.cache.config != nil {
		return m.cache.config, nil
	}

	// Get from database
	maintenance, err := m.maintenanceRepo.GetEnabledMaintenance(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance: %w", err)
	}

	if maintenance == nil {
		// No maintenance enabled, cache this state
		m.cache.enabled = false
		m.cache.config = nil
		m.cache.ttl = time.Now().Add(m.cache.cacheTTL)
		return nil, nil
	}

	// Get template
	template, err := m.maintenanceRepo.GetMaintenanceTemplate(ctx, maintenance.PageTemplate)
	if err != nil {
		m.logger.WithError(err).Warn("Failed to get maintenance template")
	}

	config := &types.MaintenanceConfig{
		PlatformMaintenance: *maintenance,
		Template:            template,
	}

	// Update cache
	m.cache.enabled = true
	m.cache.config = config
	m.cache.ttl = time.Now().Add(m.cache.cacheTTL)

	return config, nil
}

// shouldShowMaintenance determines if this request should see maintenance page
// based on rollout percentage
func (m *MaintenanceMiddleware) shouldShowMaintenance(r *http.Request, percentage int) bool {
	// Create a consistent identifier for this request
	identifier := m.getRequestIdentifier(r)

	// Hash the identifier
	h := sha256.New()
	h.Write([]byte(identifier))
	hash := h.Sum(nil)

	// Use first 4 bytes as number
	h2 := fnv.New32a()
	h2.Write(hash)
	hashNum := h2.Sum32()

	// Calculate percentage (0-100)
	requestPercentage := int(hashNum % 100)

	return requestPercentage < percentage
}

// getRequestIdentifier creates a consistent identifier for the request
func (m *MaintenanceMiddleware) getRequestIdentifier(r *http.Request) string {
	// Try to use cookie first for consistency
	cookie, err := r.Cookie("maintenance_seed")
	if err == nil && cookie != nil {
		return cookie.Value
	}

	// Fall back to IP address
	ip := GetRealIP(r)
	return ip
}

// serveMaintenancePage serves the maintenance page
func (m *MaintenanceMiddleware) serveMaintenancePage(w http.ResponseWriter, r *http.Request, config *types.MaintenanceConfig) {
	// Set appropriate headers
	w.Header().Set("Retry-After", fmt.Sprintf("%d", config.RetryAfterSeconds))
	w.Header().Set("X-Maintenance-Mode", "true")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render the maintenance page
	html := m.renderMaintenanceHTML(config)
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write([]byte(html))
}

// renderMaintenanceHTML renders the maintenance page HTML
func (m *MaintenanceMiddleware) renderMaintenanceHTML(config *types.MaintenanceConfig) string {
	template := config.Template
	if template == nil {
		template = &types.MaintenancePageTemplate{
			Name:            "default",
			Title:           stringPtr("We'll be back soon!"),
			MessageHTML:     stringPtr("<p>We're performing scheduled maintenance. We'll be back shortly.</p>"),
			BackgroundColor: "#1a1a2e",
			TextColor:       "#ffffff",
			AccentColor:     "#4ecdc4",
			ShowContactInfo: true,
			ShowSocialLinks: true,
		}
	}

	title := "We'll be back soon!"
	if template.Title != nil {
		title = *template.Title
	}

	message := "<p>We're performing scheduled maintenance. We'll be back shortly.</p>"
	if template.MessageHTML != nil {
		message = *template.MessageHTML
	}

	scheduledEnd := ""
	if config.ScheduledEnd != nil {
		scheduledEnd = config.ScheduledEnd.Format("January 2, 2006 at 3:04 PM MST")
	}

	contactInfo := ""
	if template.ShowContactInfo {
		contactEmail := "support@functionfly.com"
		if template.ContactEmail != nil {
			contactEmail = *template.ContactEmail
		}
		contactInfo = fmt.Sprintf(`<p class="contact">Contact us at <a href="mailto:%s">%s</a> for urgent matters.</p>`, contactEmail, contactEmail)
	}

	socialLinks := ""
	if template.ShowSocialLinks {
		socialLinks = `
		<div class="social-links">
			<a href="https://twitter.com/functionfly">Twitter</a>
		</div>`
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Maintenance - FunctionFly</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background-color: %s;
            color: %s;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            text-align: center;
            padding: 2rem;
        }
        .container {
            max-width: 600px;
        }
        .logo {
            font-size: 2rem;
            font-weight: 700;
            margin-bottom: 2rem;
            color: %s;
        }
        h1 {
            font-size: 2.5rem;
            margin-bottom: 1.5rem;
            color: %s;
        }
        .message {
            font-size: 1.25rem;
            line-height: 1.6;
            margin-bottom: 2rem;
            color: %s;
        }
        .message a {
            color: %s;
            text-decoration: underline;
        }
        .scheduled-end {
            font-size: 1rem;
            opacity: 0.8;
            margin-bottom: 2rem;
        }
        .contact {
            margin-top: 2rem;
            opacity: 0.8;
        }
        .contact a {
            color: %s;
        }
        .social-links {
            margin-top: 3rem;
            display: flex;
            gap: 1rem;
            justify-content: center;
        }
        .social-links a {
            color: %s;
            text-decoration: none;
            opacity: 0.7;
            transition: opacity 0.2s;
        }
        .social-links a:hover {
            opacity: 1;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">⚡ FunctionFly</div>
        <h1>%s</h1>
        <div class="message">%s</div>
        %s
        %s
        <div class="scheduled-end">
            Estimated return: %s
        </div>
    </div>
</body>
</html>`,
		template.BackgroundColor,
		template.TextColor,
		template.AccentColor,
		template.AccentColor,
		template.TextColor,
		template.AccentColor,
		template.AccentColor,
		template.AccentColor,
		title,
		message,
		contactInfo,
		socialLinks,
		scheduledEnd,
	)
}

// GetRealIP gets the real IP address from the request
func GetRealIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// stringPtr helper
func stringPtr(s string) *string {
	return &s
}
