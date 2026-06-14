package connectors

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/crypto"
	connectorEngine "github.com/functionfly/functionfly/internal/agent/connectors"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	connectorRepo *storage.ConnectorRepository
	brainRepo     *storage.BrainRepository
	scheduler     *connectorEngine.SyncScheduler
	logger        *logrus.Logger
}

// TokenResponse contains OAuth tokens from providers
type TokenResponse struct {
	AccessToken  string                 `json:"access_token"`
	RefreshToken string                `json:"refresh_token,omitempty"`
	ExpiresIn    int                    `json:"expires_in"`
	TokenType    string                 `json:"token_type"`
	Raw          map[string]interface{} `json:"raw,omitempty"`
}

func NewHandler(
	connectorRepo *storage.ConnectorRepository,
	brainRepo *storage.BrainRepository,
	scheduler *connectorEngine.SyncScheduler,
	logger *logrus.Logger,
) *Handler {
	return &Handler{
		connectorRepo: connectorRepo,
		brainRepo:     brainRepo,
		scheduler:     scheduler,
		logger:        logger,
	}
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) respondError(w http.ResponseWriter, status int, code, message string) {
	h.respondJSON(w, status, map[string]interface{}{
		"error":   http.StatusText(status),
		"code":    code,
		"message": message,
	})
}

func (h *Handler) getTenantPlan(ctx context.Context, tenantID uuid.UUID) (string, error) {
	var tier string
	err := h.connectorRepo.GetDB().QueryRowContext(ctx, `
		SELECT COALESCE(pt.name, 'free')
		FROM subscriptions s
		JOIN pricing_tiers pt ON pt.id = s.pricing_tier_id
		WHERE s.tenant_id = $1 AND s.status = 'active'
		ORDER BY s.created_at DESC LIMIT 1`, tenantID).Scan(&tier)
	if err != nil {
		return plans.PlanFree, err
	}
	return tier, nil
}

// HandleListCatalog returns all available connectors
func (h *Handler) HandleListCatalog(w http.ResponseWriter, r *http.Request) {
	connectors, err := h.connectorRepo.ListCatalog(r.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to list connectors catalog")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to list connectors")
		return
	}

	h.respondJSON(w, 200, map[string]interface{}{
		"connectors": connectors,
	})
}

// HandleListUserConnectors returns the user's linked connectors
func (h *Handler) HandleListUserConnectors(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	connectors, err := h.connectorRepo.GetUserConnectors(r.Context(), claims.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list user connectors")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to list connectors")
		return
	}

	h.respondJSON(w, 200, map[string]interface{}{
		"connectors": connectors,
	})
}

// HandleLinkConnector initiates linking a connector
func (h *Handler) HandleLinkConnector(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req storage.LinkConnectorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, 400, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.ConnectorSlug == "" {
		h.respondError(w, 400, "MISSING_FIELD", "connector_slug is required")
		return
	}

	// Check tier access
	tier := plans.PlanFree
	sub, err := h.getTenantPlan(r.Context(), claims.TenantID)
	if err == nil && sub != "" {
		tier = sub
	}
	if !plans.IsConnectorAvailableForPlan(req.ConnectorSlug, tier) {
		h.respondError(w, 403, "TIER_RESTRICTED", "This connector is not available on your plan")
		return
	}

	// Check connector limit
	existing, err := h.connectorRepo.CountUserConnectors(r.Context(), claims.TenantID)
	if err != nil {
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to check connector count")
		return
	}
	maxConnectors := plans.GetMaxConnectors(tier)
	if maxConnectors > 0 && existing >= maxConnectors {
		h.respondError(w, 403, "LIMIT_REACHED", "Connector limit reached for your plan")
		return
	}

	// Get connector catalog entry
	connector, err := h.connectorRepo.GetConnectorBySlug(r.Context(), req.ConnectorSlug)
	if err != nil {
		h.respondError(w, 404, "NOT_FOUND", "Connector not found")
		return
	}

	// Create user connector
	uc := &storage.UserConnector{
		TenantID:             claims.TenantID,
		ConnectorID:          connector.ID,
		DisplayName:          req.DisplayName,
		Status:               "active",
		EncryptedCredentials: req.EncryptedCredentials,
	}

	if uc.DisplayName == "" {
		uc.DisplayName = connector.Name
	}
	if uc.EncryptedCredentials == nil {
		uc.EncryptedCredentials = []byte("{}")
	}

	created, err := h.connectorRepo.CreateUserConnector(r.Context(), uc)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create user connector")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to link connector")
		return
	}

	h.respondJSON(w, 201, map[string]interface{}{
		"connector": created,
		"message":   "Connector linked successfully",
	})
}

// HandleOAuthCallback handles OAuth callback from frontend (POST with session)
func (h *Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	var req struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, 400, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Code == "" {
		h.respondError(w, 400, "MISSING_FIELD", "code is required")
		return
	}
	if req.State == "" {
		h.respondError(w, 400, "MISSING_FIELD", "state is required")
		return
	}

	// Consume OAuth state (validates and deletes in one step)
	oauthState, err := h.connectorRepo.ConsumeOAuthState(r.Context(), req.State)
	if err != nil {
		h.logger.WithError(err).Error("Failed to consume OAuth state")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to validate OAuth state")
		return
	}
	if oauthState == nil {
		h.respondError(w, 400, "INVALID_STATE", "OAuth state expired or invalid. Please try connecting again.")
		return
	}

	// Verify tenant ownership
	if oauthState.TenantID != claims.TenantID {
		h.respondError(w, 403, "FORBIDDEN", "OAuth state does not match current session")
		return
	}

	// Process callback
	h.processOAuthCallback(w, r, oauthState, req.Code)
}

// HandleOAuthCallbackGet handles OAuth callback from provider redirect (GET without session)
// Tenant is identified from the state parameter, not from session
// This creates the user_connector directly since there's no way to communicate with the opener
func (h *Handler) HandleOAuthCallbackGet(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		h.showCallbackPage(w, "error", "Missing code parameter", "", "")
		return
	}
	if state == "" {
		h.showCallbackPage(w, "error", "Missing state parameter", "", "")
		return
	}

	// Check for OAuth provider error (e.g., user denied access)
	if oauthError := r.URL.Query().Get("error"); oauthError != "" {
		errorDesc := r.URL.Query().Get("error_description")
		if errorDesc == "" {
			errorDesc = oauthError
		}
		h.showCallbackPage(w, "error", fmt.Sprintf("Authorization denied: %s", errorDesc), "", "")
		return
	}

	// Consume OAuth state (validates and deletes in one step)
	oauthState, err := h.connectorRepo.ConsumeOAuthState(r.Context(), state)
	if err != nil {
		h.logger.WithError(err).Error("Failed to consume OAuth state")
		h.showCallbackPage(w, "error", "Failed to validate OAuth state. Please try again.", "", "")
		return
	}
	if oauthState == nil {
		h.showCallbackPage(w, "error", "OAuth session expired or invalid. Please try connecting again.", "", "")
		return
	}

	// Get connector info by ID (ConnectorID is a UUID, not a slug)
	connector, err := h.connectorRepo.GetConnectorByID(r.Context(), oauthState.ConnectorID)
	if err != nil {
		h.logger.WithError(err).WithField("connector_id", oauthState.ConnectorID).Error("Connector not found by ID")
		h.showCallbackPage(w, "error", "Connector not found. Please try again.", "", "")
		return
	}

	// Build redirect URI for token exchange
	// Must match the redirect_uri used when the OAuth URL was generated
	redirectURI := os.Getenv("APP_BASE_URL")
	if redirectURI == "" {
		redirectURI = "https://app.functionfly.com"
	}
	redirectURI = redirectURI + "/v1/connectors/callback"

	// Exchange code for tokens with the OAuth provider
	tokens, err := h.exchangeOAuthCode(r.Context(), connector.Slug, code, redirectURI)
	if err != nil {
		h.logger.WithError(err).WithField("slug", connector.Slug).Error("Failed to exchange OAuth code")
		h.showCallbackPage(w, "error", fmt.Sprintf("Failed to authorize with %s: %v", connector.Name, err), connector.Name, connector.Slug)
		return
	}

	// Encrypt tokens with server-side key for background sync
	tokenJSON, err := json.Marshal(tokens)
	if err != nil {
		h.showCallbackPage(w, "error", "Failed to process tokens", connector.Name, connector.Slug)
		return
	}

	serverEncrypted, err := crypto.EncryptToJSON(tokenJSON, oauthState.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to encrypt tokens server-side")
		h.showCallbackPage(w, "error", "Failed to secure tokens", connector.Name, connector.Slug)
		return
	}

	// Create user_connector with encrypted credentials
	encryptedCredJSON, _ := json.Marshal(map[string]interface{}{
		"server_sync": serverEncrypted,
		"provider":    connector.Slug,
		"linked_at":   time.Now().UTC().Format(time.RFC3339),
	})

	uc := &storage.UserConnector{
		TenantID:             oauthState.TenantID,
		ConnectorID:          connector.ID,
		DisplayName:          connector.Name,
		Status:               "active",
		EncryptedCredentials: encryptedCredJSON,
	}

	// Check if connector already linked for this tenant
	existing, _ := h.connectorRepo.GetUserConnectorBySlug(r.Context(), oauthState.TenantID, connector.Slug)
	if existing != nil {
		h.showCallbackPage(w, "success", fmt.Sprintf("%s is already connected!", connector.Name), connector.Name, connector.Slug)
		return
	}

	created, err := h.connectorRepo.CreateUserConnector(r.Context(), uc)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create user connector")
		h.showCallbackPage(w, "error", "Failed to save connector. Please try again.", connector.Name, connector.Slug)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"connector_id": created.ID,
		"tenant_id":    oauthState.TenantID,
		"slug":         connector.Slug,
	}).Info("Connector linked successfully via OAuth callback")

	h.showCallbackPage(w, "success", fmt.Sprintf("%s connected successfully!", connector.Name), connector.Name, connector.Slug)
}

// showCallbackPage renders a simple HTML page for the OAuth callback popup.
// It uses window.opener.postMessage() to notify the parent window of the result
// and auto-closes after a short delay.
func (h *Handler) showCallbackPage(w http.ResponseWriter, status, message, connectorName, connectorSlug string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	icon := "✓"
	color := "#22c55e" // green
	if status == "error" {
		icon = "✗"
		color = "#ef4444" // red
	}

	// HTML-escape user-controlled strings to prevent XSS
	safeMessage := html.EscapeString(message)
	safeStatus := html.EscapeString(status)

	// Determine the frontend origin for postMessage target.
	// Use APP_FRONTEND_URL if set; otherwise '*' as safe fallback since we only
	// message window.opener (our own dashboard popup) with non-sensitive data.
	frontendOrigin := os.Getenv("APP_FRONTEND_URL")
	if frontendOrigin == "" {
		if os.Getenv("DEVELOPMENT") == "true" {
			frontendOrigin = "*"
		} else {
			frontendOrigin = "https://app.functionfly.com"
		}
	}

	// Build the postMessage payload as JSON
	msgPayload, _ := json.Marshal(map[string]interface{}{
		"type":            "oauth_callback",
		"status":          status,
		"connector_slug":  connectorSlug,
		"connector_name":  connectorName,
		"message":         message,
	})

	// Auto-close delay: 3s for success (to show confirmation), 0 for error (user needs to read it)
	autoCloseDelay := "3000"
	if status == "error" {
		autoCloseDelay = "0" // Don't auto-close errors
	}

	htmlPage := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
	<title>Connector Callback</title>
	<meta name="robots" content="noindex, nofollow">
	<style>
		body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #0f172a; color: #fff; min-height: 100vh; margin: 0; display: flex; align-items: center; justify-content: center; }
		.card { background: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 32px; max-width: 400px; text-align: center; }
		.icon { font-size: 48px; margin-bottom: 16px; color: %s; }
		h2 { margin: 0 0 12px; font-size: 20px; font-weight: 600; }
		p { margin: 0 0 24px; color: #94a3b8; line-height: 1.5; }
		button { background: #3b82f6; color: #fff; border: none; padding: 12px 24px; border-radius: 8px; font-size: 14px; cursor: pointer; }
		button:hover { background: #2563eb; }
		.countdown { color: #64748b; font-size: 12px; margin-top: 8px; }
	</style>
</head>
<body>
	<div class="card">
		<div class="icon">%s</div>
		<h2>%s</h2>
		<p>%s</p>
		<button onclick="closeWindow()">Close Window</button>
		<div id="countdown" class="countdown"></div>
	</div>
	<script>
		var payload = %s;
		var targetOrigin = %q;

		function notifyOpener() {
			try {
				if (window.opener && !window.opener.closed) {
					window.opener.postMessage(payload, targetOrigin);
				}
			} catch (e) {
				console.error('postMessage failed:', e);
			}
		}

		function closeWindow() {
			try { notifyOpener(); } catch(e) {}
			try { window.close(); } catch(e) {}
		}

		// Notify parent immediately
		notifyOpener();

		%s
	</script>
</body>
</html>`, color, icon, safeStatus, safeMessage, string(msgPayload), frontendOrigin,
		func() string {
			if autoCloseDelay != "0" {
				return fmt.Sprintf(`
		// Auto-close after delay
		var seconds = %s / 1000;
		var countdownEl = document.getElementById('countdown');
		if (countdownEl) countdownEl.textContent = 'Closing in ' + seconds + 's...';
		var interval = setInterval(function() {
			seconds--;
			if (seconds <= 0) {
				clearInterval(interval);
				if (countdownEl) countdownEl.textContent = '';
			} else {
				if (countdownEl) countdownEl.textContent = 'Closing in ' + seconds + 's...';
			}
		}, 1000);
		setTimeout(function() { window.close(); }, %s);`, autoCloseDelay, autoCloseDelay)
			}
			return ""
		}())

	// For errors, set appropriate status code
	if status == "error" {
		w.WriteHeader(400)
	}
	w.Write([]byte(htmlPage))
}

// processOAuthCallback does the actual token exchange and encryption
func (h *Handler) processOAuthCallback(w http.ResponseWriter, r *http.Request, oauthState *storage.ConnectorOAuthState, code string) {
	// Get connector by ID (ConnectorID is a UUID, not a slug)
	connector, err := h.connectorRepo.GetConnectorByID(r.Context(), oauthState.ConnectorID)
	if err != nil {
		h.respondError(w, 404, "NOT_FOUND", "Connector not found")
		return
	}

	// Build redirect URI for token exchange
	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "https://app.functionfly.com"
	}
	redirectURI := baseURL + "/v1/connectors/callback"

	// Exchange code for tokens with the OAuth provider
	tokens, err := h.exchangeOAuthCode(r.Context(), connector.Slug, code, redirectURI)
	if err != nil {
		h.logger.WithError(err).WithField("slug", connector.Slug).Error("Failed to exchange OAuth code")
		h.respondError(w, 502, "PROVIDER_ERROR", fmt.Sprintf("Failed to authorize with %s: %v", connector.Name, err))
		return
	}

	// Encrypt tokens with server-side key for background sync
	tokenJSON, err := json.Marshal(tokens)
	if err != nil {
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to process tokens")
		return
	}

	serverEncrypted, err := crypto.EncryptToJSON(tokenJSON, oauthState.TenantID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to encrypt tokens server-side")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to secure tokens")
		return
	}

	// Return tokens and encrypted blob to frontend
	// Frontend will encrypt with user's vault key and call linkConnector
	h.respondJSON(w, 200, map[string]interface{}{
		"status":           "success",
		"tokens":           tokens,
		"server_encrypted": serverEncrypted,
		"connector_slug":   connector.Slug,
	})
}

// exchangeOAuthCode exchanges an authorization code for tokens with the OAuth provider
func (h *Handler) exchangeOAuthCode(ctx context.Context, slug, code, redirectURI string) (*TokenResponse, error) {
	switch slug {
	case "github":
		return h.exchangeGitHubToken(ctx, code, redirectURI)
	case "slack":
		return h.exchangeSlackToken(ctx, code, redirectURI)
	case "notion":
		return h.exchangeNotionToken(ctx, code, redirectURI)
	case "gmail":
		return h.exchangeGmailToken(ctx, code, redirectURI)
	case "linear":
		return h.exchangeLinearToken(ctx, code, redirectURI)
	default:
		return nil, fmt.Errorf("unsupported connector: %s", slug)
	}
}

// GitHub token exchange
func (h *Handler) exchangeGitHubToken(ctx context.Context, code, redirectURI string) (*TokenResponse, error) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	resp, err := http.PostForm("https://github.com/login/oauth/access_token", url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	})
	if err != nil {
		return nil, fmt.Errorf("GitHub token request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub response: %w", err)
	}

	return &TokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    result.TokenType,
	}, nil
}

// Slack token exchange
func (h *Handler) exchangeSlackToken(ctx context.Context, code, redirectURI string) (*TokenResponse, error) {
	clientID := os.Getenv("SLACK_CLIENT_ID")
	clientSecret := os.Getenv("SLACK_CLIENT_SECRET")

	resp, err := http.PostForm("https://slack.com/api/oauth.v2.access", url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	})
	if err != nil {
		return nil, fmt.Errorf("Slack token request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		BotToken    string `json:"bot_access_token"` // Slack uses bot_access_token
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		TeamID       string `json:"team_id"`
		Ok           bool   `json:"ok"`
		Error        string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Slack response: %w", err)
	}

	if !result.Ok {
		return nil, fmt.Errorf("Slack error: %s", result.Error)
	}

	// Use bot token if available, otherwise use regular token
	accessToken := result.AccessToken
	if result.BotToken != "" {
		accessToken = result.BotToken
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    result.TokenType,
		Raw: map[string]interface{}{
			"team_id": result.TeamID,
			"bot_token": result.BotToken,
		},
	}, nil
}

// Notion token exchange
func (h *Handler) exchangeNotionToken(ctx context.Context, code, redirectURI string) (*TokenResponse, error) {
	clientID := os.Getenv("NOTION_CLIENT_ID")
	clientSecret := os.Getenv("NOTION_CLIENT_SECRET")

	// Notion uses Basic auth with client_id:client_secret in Authorization header
	auth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))

	body := bytes.NewReader([]byte(fmt.Sprintf(`{
		"grant_type": "authorization_code",
		"code": %q,
		"redirect_uri": %q
	}`, code, redirectURI)))

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.notion.com/v1/oauth/token", body)
	if err != nil {
		return nil, fmt.Errorf("Notion token request creation failed: %w", err)
	}
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Notion token request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Notion response: %w", err)
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		WorkspaceID  string `json:"workspace_id"`
		WorkspaceName string `json:"workspace_name"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode Notion response: %w", err)
	}

	return &TokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    result.TokenType,
		Raw: map[string]interface{}{
			"workspace_id":   result.WorkspaceID,
			"workspace_name": result.WorkspaceName,
		},
	}, nil
}

// Gmail token exchange
func (h *Handler) exchangeGmailToken(ctx context.Context, code, redirectURI string) (*TokenResponse, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		return nil, fmt.Errorf("Google token request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Google response: %w", err)
	}

	return &TokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    result.TokenType,
	}, nil
}

// Linear token exchange
func (h *Handler) exchangeLinearToken(ctx context.Context, code, redirectURI string) (*TokenResponse, error) {
	clientID := os.Getenv("LINEAR_CLIENT_ID")
	clientSecret := os.Getenv("LINEAR_CLIENT_SECRET")

	resp, err := http.PostForm("https://api.linear.app/oauth/token", url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	})
	if err != nil {
		return nil, fmt.Errorf("Linear token request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Linear response: %w", err)
	}

	return &TokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    result.TokenType,
	}, nil
}

// HandleUnlinkConnector removes a linked connector
func (h *Handler) HandleUnlinkConnector(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	connectorID, err := uuid.Parse(vars["connector_id"])
	if err != nil {
		h.respondError(w, 400, "INVALID_ID", "Invalid connector ID")
		return
	}

	if err := h.connectorRepo.DeleteUserConnector(r.Context(), claims.TenantID, connectorID); err != nil {
		h.respondError(w, 404, "NOT_FOUND", "Connector not found")
		return
	}

	h.respondJSON(w, 200, map[string]interface{}{
		"message": "Connector unlinked successfully",
	})
}

// HandleTriggerSync triggers a manual sync for a connector
func (h *Handler) HandleTriggerSync(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	connectorID, err := uuid.Parse(vars["connector_id"])
	if err != nil {
		h.respondError(w, 400, "INVALID_ID", "Invalid connector ID")
		return
	}

	uc, err := h.connectorRepo.GetUserConnector(r.Context(), claims.TenantID, connectorID)
	if err != nil {
		h.respondError(w, 404, "NOT_FOUND", "Connector not found")
		return
	}

	h.respondJSON(w, 200, storage.SyncTriggerResponse{
		Status:    "syncing",
		Message:   "Sync started for " + uc.ConnectorName,
		StartedAt: time.Now().UTC(),
	})
}

// HandleUpdateUserConnector updates connector settings
func (h *Handler) HandleUpdateUserConnector(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	vars := mux.Vars(r)
	connectorID, err := uuid.Parse(vars["connector_id"])
	if err != nil {
		h.respondError(w, 400, "INVALID_ID", "Invalid connector ID")
		return
	}

	var req struct {
		Enabled       *bool   `json:"enabled"`
		DisplayName   *string `json:"display_name"`
		SyncFrequency *string `json:"sync_frequency"`
		AutoSync      *bool   `json:"auto_sync"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, 400, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.connectorRepo.UpdateUserConnectorSettings(r.Context(), claims.TenantID, connectorID, req.Enabled, req.DisplayName, req.SyncFrequency, req.AutoSync); err != nil {
		h.logger.WithError(err).Error("Failed to update connector settings")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to update connector")
		return
	}

	uc, err := h.connectorRepo.GetUserConnector(r.Context(), claims.TenantID, connectorID)
	if err != nil {
		h.respondError(w, 404, "NOT_FOUND", "Connector not found")
		return
	}

	h.respondJSON(w, 200, map[string]interface{}{
		"connector": uc,
	})
}

// HandleGetConnectorOAuthURL generates OAuth URL for connector linking
func (h *Handler) HandleGetConnectorOAuthURL(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		h.respondError(w, 401, "UNAUTHORIZED", "Authentication required")
		return
	}

	slug := r.URL.Query().Get("slug")
	if slug == "" {
		h.respondError(w, 400, "MISSING_FIELD", "slug query parameter is required")
		return
	}

	connector, err := h.connectorRepo.GetConnectorBySlug(r.Context(), slug)
	if err != nil {
		h.respondError(w, 404, "NOT_FOUND", "Connector not found")
		return
	}

	state := uuid.New().String()

	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "https://app.functionfly.com"
	}
	redirectURI := baseURL + "/v1/connectors/callback"

	var clientID string
	var clientSecret string
	switch connector.Slug {
	case "github":
		clientID = os.Getenv("GITHUB_CLIENT_ID")
		clientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
	case "slack":
		clientID = os.Getenv("SLACK_CLIENT_ID")
		clientSecret = os.Getenv("SLACK_CLIENT_SECRET")
	case "notion":
		clientID = os.Getenv("NOTION_CLIENT_ID")
		clientSecret = os.Getenv("NOTION_CLIENT_SECRET")
	case "gmail":
		clientID = os.Getenv("GOOGLE_CLIENT_ID")
		clientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	case "linear":
		clientID = os.Getenv("LINEAR_CLIENT_ID")
		clientSecret = os.Getenv("LINEAR_CLIENT_SECRET")
	default:
		clientID = os.Getenv("GITHUB_CLIENT_ID")
		clientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
	}

	if clientID == "" || clientSecret == "" {
		h.logger.WithField("slug", connector.Slug).Warn("OAuth client credentials not configured")
		h.respondError(w, 503, "NOT_CONFIGURED", fmt.Sprintf("OAuth for %s is not configured. Please contact support.", connector.Name))
		return
	}

	oauthURL := buildOAuthURL(connector.Slug, clientID, redirectURI, state)


	if err := h.connectorRepo.StoreOAuthState(r.Context(), state, claims.TenantID, connector.ID); err != nil {
		h.logger.WithError(err).Error("Failed to store OAuth state")
		h.respondError(w, 500, "INTERNAL_ERROR", "Failed to store OAuth state")
		return
	}

	h.respondJSON(w, 200, map[string]interface{}{
		"oauth_url": oauthURL,
		"state":    state,
	})
}

func buildOAuthURL(slug, clientID, redirectURI, state string) string {
	switch slug {
	case "github":
		return "https://github.com/login/oauth/authorize?client_id=" + clientID + "&redirect_uri=" + url.QueryEscape(redirectURI) + "&scope=repo,read:user,notifications&state=" + state
	case "notion":
		return "https://api.notion.com/v1/oauth/authorize?client_id=" + clientID + "&owner=user&redirect_uri=" + url.QueryEscape(redirectURI) + "&state=" + state
	case "slack":
		return "https://slack.com/oauth/v2/authorize?client_id=" + clientID + "&scope=channels:history,users:read,reactions:read&redirect_uri=" + url.QueryEscape(redirectURI) + "&state=" + state
	case "gmail":
		return "https://accounts.google.com/o/oauth2/v2/auth?client_id=" + clientID + "&redirect_uri=" + url.QueryEscape(redirectURI) + "&scope=https://www.googleapis.com/auth/gmail.readonly&response_type=code&state=" + state
	case "linear":
		return "https://linear.app/oauthAuthorize?client_id=" + clientID + "&redirect_uri=" + url.QueryEscape(redirectURI) + "&scope=read,issues:read&state=" + state
	default:
		return strings.Replace(connectorOAuthURLs[slug], "{client_id}", clientID, 1)
	}
}

var connectorOAuthURLs = map[string]string{
	"github":  "https://github.com/login/oauth/authorize?client_id={client_id}&redirect_uri={redirect_uri}&scope=repo,read:user,notifications&state={state}",
	"notion":  "https://api.notion.com/v1/oauth/authorize?client_id={client_id}&owner=user&redirect_uri={redirect_uri}&state={state}",
	"slack":   "https://slack.com/oauth/v2/authorize?client_id={client_id}&scope=channels:history,users:read,reactions:read&redirect_uri={redirect_uri}&state={state}",
	"gmail":   "https://accounts.google.com/o/oauth2/v2/auth?client_id={client_id}&redirect_uri={redirect_uri}&scope=https://www.googleapis.com/auth/gmail.readonly&response_type=code&state={state}",
	"linear":  "https://linear.app/oauthAuthorize?client_id={client_id}&redirect_uri={redirect_uri}&scope=read,issues:read&state={state}",
}
