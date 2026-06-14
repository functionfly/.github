package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TenantAuthHandler handles per-tenant authentication API endpoints
type TenantAuthHandler struct {
	repo storage.Repository
}

// NewTenantAuthHandler creates a new tenant auth handler
func NewTenantAuthHandler(repo storage.Repository) *TenantAuthHandler {
	return &TenantAuthHandler{
		repo: repo,
	}
}

// GetSettings handles GET /v1/auth/settings
func (h *TenantAuthHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := getUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Tenant context required")
		return
	}

	settings, err := h.repo.GetAuthSettings(ctx, claims.TenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get auth settings")
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve settings")
		return
	}
	if settings == nil {
		writeJSONError(w, http.StatusNotFound, "Auth settings not found")
		return
	}

	writeJSON(w, http.StatusOK, settings)
}

// UpdateSettings handles PATCH /v1/auth/settings
func (h *TenantAuthHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := getUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Tenant context required")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate MFA mode
	if mode, ok := updates["mfa_mode"].(string); ok {
		if mode != storage.MFAModeOptional && mode != storage.MFAModeRequired && mode != storage.MFAModeEnforced {
			writeJSONError(w, http.StatusBadRequest, "Invalid MFA mode")
			return
		}
	}

	// Validate SSO provider
	if provider, ok := updates["sso_provider"].(string); ok {
		if provider != storage.SSOProviderNone && provider != storage.SSOProviderSAML && provider != storage.SSOProviderOIDC {
			writeJSONError(w, http.StatusBadRequest, "Invalid SSO provider")
			return
		}
	}

	updates["updated_at"] = time.Now()

	if err := 	h.repo.UpdateAuthSettings(ctx, claims.TenantID, updates); err != nil {
		logrus.WithError(err).Error("Failed to update auth settings")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update settings")
		return
	}

	// Return updated settings
	settings, _ := 	h.repo.GetAuthSettings(ctx, claims.TenantID)
	writeJSON(w, http.StatusOK, settings)
}

// ListOAuthProviders handles GET /v1/auth/oauth
func (h *TenantAuthHandler) ListOAuthProviders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := getUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Tenant context required")
		return
	}

	providers, err := 	h.repo.ListOAuthProviders(ctx, claims.TenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list OAuth providers")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list providers")
		return
	}

	// Return without secrets
	safeProviders := make([]map[string]interface{}, 0, len(providers))
	for _, p := range providers {
		safeProviders = append(safeProviders, map[string]interface{}{
			"id":           p.ID,
			"provider":     p.Provider,
			"client_id":    p.ClientID,
			"enabled":      p.Enabled,
			"callback_url": p.CallbackURL,
			"scopes":       p.Scopes,
			"created_at":   p.CreatedAt,
			"updated_at":   p.UpdatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"providers": safeProviders})
}

// ConfigureOAuthProvider handles POST /v1/auth/oauth
func (h *TenantAuthHandler) ConfigureOAuthProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := getUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Tenant context required")
		return
	}

	var config struct {
		Provider     string   `json:"provider"`
		ClientID     string   `json:"client_id"`
		ClientSecret string   `json:"client_secret"`
		Enabled      bool     `json:"enabled"`
		CallbackURL  string   `json:"callback_url"`
		Scopes       []string `json:"scopes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate provider
	if !storage.IsValidOAuthProvider(config.Provider) {
		writeJSONError(w, http.StatusBadRequest, "Invalid OAuth provider")
		return
	}

	// Validate required fields
	if config.ClientID == "" || config.ClientSecret == "" {
		writeJSONError(w, http.StatusBadRequest, "Client ID and secret are required")
		return
	}

	// Encrypt client secret using repo's encryption
	encryptedSecret, err := 	h.repo.EncryptField(r.Context(), config.ClientSecret)
	if err != nil {
		logrus.WithError(err).Error("Failed to encrypt client secret")
		writeJSONError(w, http.StatusInternalServerError, "Failed to encrypt secret")
		return
	}

	// Default scopes if not provided
	scopes := config.Scopes
	if len(scopes) == 0 {
		scopes = []string{"user:email", "read:user"}
	}
	scopesJSON, _ := json.Marshal(scopes)

	provider := &storage.TenantOAuthProvider{
		ID:                    uuid.New(),
		TenantID:              claims.TenantID,
		Provider:              config.Provider,
		ClientID:              config.ClientID,
		EncryptedClientSecret: encryptedSecret,
		Enabled:               config.Enabled,
		CallbackURL:           &config.CallbackURL,
		Scopes:                scopesJSON,
	}

	// Check if provider already exists
	existing, err := 	h.repo.GetOAuthProvider(ctx, claims.TenantID, config.Provider)
	if err != nil {
		logrus.WithError(err).Error("Failed to check existing provider")
		writeJSONError(w, http.StatusInternalServerError, "Failed to check provider")
		return
	}

	if existing != nil {
		// Update existing - we need to re-encrypt the secret
		updates := map[string]interface{}{
			"client_id":                 config.ClientID,
			"encrypted_client_secret": encryptedSecret,
			"enabled":                  config.Enabled,
			"callback_url":             config.CallbackURL,
			"scopes":                   scopesJSON,
			"updated_at":               time.Now(),
		}
		updated, err := 	h.repo.UpdateOAuthProvider(ctx, claims.TenantID, config.Provider, updates)
		if err != nil {
			logrus.WithError(err).Error("Failed to update OAuth provider")
			writeJSONError(w, http.StatusInternalServerError, "Failed to update provider")
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"provider": safeProviderResponse(updated)})
		return
	}

	// Create new
	if err := 	h.repo.CreateOAuthProvider(ctx, provider); err != nil {
		logrus.WithError(err).Error("Failed to create OAuth provider")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create provider")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"provider": safeProviderResponse(provider)})
}

// DeleteOAuthProvider handles DELETE /v1/auth/oauth/:provider
func (h *TenantAuthHandler) DeleteOAuthProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := getUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Tenant context required")
		return
	}

	provider := r.PathValue("provider")
	if !storage.IsValidOAuthProvider(provider) {
		writeJSONError(w, http.StatusBadRequest, "Invalid OAuth provider")
		return
	}

	if err := 	h.repo.DeleteOAuthProvider(ctx, claims.TenantID, provider); err != nil {
		logrus.WithError(err).Error("Failed to delete OAuth provider")
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete provider")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"deleted": true})
}

// ListMembers handles GET /v1/auth/members
func (h *TenantAuthHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := getUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Tenant context required")
		return
	}

	memberships, err := 	h.repo.ListMemberships(ctx, claims.TenantID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list members")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list members")
		return
	}

	// Enrich with user data
	members := make([]map[string]interface{}, 0, len(memberships))
	for _, m := range memberships {
		member := map[string]interface{}{
			"id":         m.ID,
			"tenant_id":  m.TenantID,
			"user_id":    m.UserID,
			"role":       m.Role,
			"status":    m.Status,
			"joined_at":  m.JoinedAt,
		}
		if m.User != nil {
			member["user"] = map[string]interface{}{
				"id":    m.User.ID,
				"email": m.User.Email,
				"name":  m.User.Name,
			}
		}
		members = append(members, member)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"members": members})
}

// InviteMember handles POST /v1/auth/members/invite
func (h *TenantAuthHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := getUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Tenant context required")
		return
	}

	var req struct {
		Email     string `json:"email"`
		Role      string `json:"role"`
		ExpiresIn int    `json:"expires_in_hours"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate email
	if req.Email == "" {
		writeJSONError(w, http.StatusBadRequest, "Email is required")
		return
	}

	// Validate role - default to team_member
	if req.Role == "" {
		req.Role = storage.RoleTeamMember
	}
	if !storage.IsValidRole(req.Role) {
		writeJSONError(w, http.StatusBadRequest, "Invalid role")
		return
	}

	// Check if user already has a membership
	existingUser, err := 	h.repo.GetUserByEmail(r.Context(), req.Email)
	if err == nil && existingUser != nil {
		membership, err := 	h.repo.GetMembership(ctx, claims.TenantID, existingUser.ID)
		if err == nil && membership != nil {
			writeJSONError(w, http.StatusConflict, "User is already a member")
			return
		}
	}

	// Check for existing pending invite
	existingInvite, err := 	h.repo.GetInviteCodeByEmail(ctx, claims.TenantID, req.Email)
	if err == nil && existingInvite != nil {
		writeJSONError(w, http.StatusConflict, "There is already a pending invite")
		return
	}

	// Generate secure invite code
	code := generateSecureCode()

	// Default expiration: 72 hours
	expiresIn := 72
	if req.ExpiresIn > 0 {
		expiresIn = req.ExpiresIn
	}

	invite := &storage.TenantInviteCode{
		ID:        uuid.New(),
		TenantID:  claims.TenantID,
		Code:      code,
		Email:     req.Email,
		Role:      req.Role,
		InvitedBy: claims.UserID,
		ExpiresAt: time.Now().Add(time.Duration(expiresIn) * time.Hour),
		MaxUses:   1,
	}

	if err := 	h.repo.CreateInviteCode(ctx, invite); err != nil {
		logrus.WithError(err).Error("Failed to create invite")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create invite")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"invite": invite,
		"url":    fmt.Sprintf("/invite/%s", code),
	})
}

// AcceptInvite handles POST /v1/auth/invites/:code/accept
func (h *TenantAuthHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := getUserFromContext(r)
	if claims == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	code := r.PathValue("code")

	invite, err := 	h.repo.GetInviteCode(ctx, code)
	if err != nil {
		logrus.WithError(err).Error("Failed to get invite")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get invite")
		return
	}
	if invite == nil {
		writeJSONError(w, http.StatusNotFound, "Invite not found")
		return
	}

	// Check expiration
	if time.Now().After(invite.ExpiresAt) {
		writeJSONError(w, http.StatusGone, "Invite has expired")
		return
	}

	// Check if already accepted
	if invite.AcceptedAt != nil {
		writeJSONError(w, http.StatusConflict, "Invite has already been used")
		return
	}

	// Create membership
	membership := &storage.TenantMembership{
		ID:        uuid.New(),
		TenantID:  invite.TenantID,
		UserID:    claims.UserID,
		Role:      invite.Role,
		InvitedBy: &invite.InvitedBy,
		InvitedAt: &invite.CreatedAt,
		Status:    storage.MembershipStatusActive,
	}

	if err := 	h.repo.CreateMembership(ctx, membership); err != nil {
		logrus.WithError(err).Error("Failed to create membership")
		writeJSONError(w, http.StatusInternalServerError, "Failed to join organization")
		return
	}

	// Accept the invite
	if err := 	h.repo.AcceptInviteCode(ctx, code, claims.UserID); err != nil {
		logrus.WithError(err).Warn("Failed to mark invite as accepted")
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"membership": membership,
		"tenant_id":   invite.TenantID,
	})
}

// UpdateMemberRole handles PATCH /v1/auth/members/:userId/role
func (h *TenantAuthHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := getUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Tenant context required")
		return
	}

	userIDStr := r.PathValue("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if !storage.IsValidRole(req.Role) {
		writeJSONError(w, http.StatusBadRequest, "Invalid role")
		return
	}

	membership, err := 	h.repo.GetMembership(ctx, claims.TenantID, userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get membership")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get membership")
		return
	}
	if membership == nil {
		writeJSONError(w, http.StatusNotFound, "User is not a member")
		return
	}

	// Prevent demoting the last owner
	if membership.Role == storage.RoleTeamOwner && req.Role != storage.RoleTeamOwner {
		ownerCount, err := 	h.repo.CountMembershipsByRole(ctx, claims.TenantID, storage.RoleTeamOwner)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Failed to count owners")
			return
		}
		if ownerCount <= 1 {
			writeJSONError(w, http.StatusConflict, "Cannot demote the last owner")
			return
		}
	}

	updates := map[string]interface{}{"role": req.Role}
	updated, err := 	h.repo.UpdateMembership(ctx, claims.TenantID, userID, updates)
	if err != nil {
		logrus.WithError(err).Error("Failed to update membership")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update role")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"member": updated})
}

// RemoveMember handles DELETE /v1/auth/members/:userId
func (h *TenantAuthHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := getUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Tenant context required")
		return
	}

	userIDStr := r.PathValue("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	membership, err := 	h.repo.GetMembership(ctx, claims.TenantID, userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get membership")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get membership")
		return
	}
	if membership == nil {
		writeJSONError(w, http.StatusNotFound, "User is not a member")
		return
	}

	// Prevent removing the last owner
	if membership.Role == storage.RoleTeamOwner {
		ownerCount, err := 	h.repo.CountMembershipsByRole(ctx, claims.TenantID, storage.RoleTeamOwner)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Failed to count owners")
			return
		}
		if ownerCount <= 1 {
			writeJSONError(w, http.StatusConflict, "Cannot remove the last owner")
			return
		}
	}

	if err := 	h.repo.DeleteMembership(ctx, claims.TenantID, userID); err != nil {
		logrus.WithError(err).Error("Failed to remove membership")
		writeJSONError(w, http.StatusInternalServerError, "Failed to remove member")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"removed": true})
}

// ListPendingInvites handles GET /v1/auth/invites
func (h *TenantAuthHandler) ListPendingInvites(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := getUserFromContext(r)
	if claims == nil || claims.TenantID == uuid.Nil {
		writeJSONError(w, http.StatusUnauthorized, "Tenant context required")
		return
	}

	invites, err := 	h.repo.GetInviteCodesByTenant(ctx, claims.TenantID, false)
	if err != nil {
		logrus.WithError(err).Error("Failed to list invites")
		writeJSONError(w, http.StatusInternalServerError, "Failed to list invites")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"invites": invites})
}

// RevokeInvite handles DELETE /v1/auth/invites/:code
func (h *TenantAuthHandler) RevokeInvite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := r.PathValue("code")

	if err := 	h.repo.RevokeInviteCode(ctx, code); err != nil {
		logrus.WithError(err).Error("Failed to revoke invite")
		writeJSONError(w, http.StatusInternalServerError, "Failed to revoke invite")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"revoked": true})
}

// Helper functions

func generateSecureCode() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return uuid.New().String()
	}
	return hex.EncodeToString(b)
}

func safeProviderResponse(p *storage.TenantOAuthProvider) map[string]interface{} {
	return map[string]interface{}{
		"id":           p.ID,
		"provider":     p.Provider,
		"client_id":    p.ClientID,
		"enabled":      p.Enabled,
		"callback_url": p.CallbackURL,
		"scopes":       p.Scopes,
		"created_at":   p.CreatedAt,
		"updated_at":   p.UpdatedAt,
	}
}

type userClaims struct {
	UserID   uuid.UUID
	TenantID uuid.UUID
	Email    string
	Role     string
}

func getUserFromContext(r *http.Request) *userClaims {
	ctx := r.Context()
	if claims, ok := ctx.Value("user").(*userClaims); ok {
		return claims
	}
	return nil
}
