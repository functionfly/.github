package providers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/auth"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo storage.Repository
}

type ProviderValidationRequest struct {
	Provider string `json:"provider"`
	Token    string `json:"token"`
}

type ProviderValidationResponse struct {
	IsValid  bool   `json:"is_valid"`
	Message  string `json:"message,omitempty"`
	UserID   string `json:"user_id,omitempty"`
	Email    string `json:"email,omitempty"`
}

type CostEstimationRequest struct {
	Provider        string `json:"provider"`
	FunctionName    string `json:"function_name"`
	Runtime         string `json:"runtime"`
	MemoryMB        int    `json:"memory_mb"`
	RequestsPerDay  int    `json:"requests_per_day"`
	ComputeDuration int    `json:"compute_duration_ms"`
	Regions         []string `json:"regions"`
}

type CostEstimationResponse struct {
	MonthlyCost  float64               `json:"monthly_cost"`
	Currency     string                `json:"currency"`
	Breakdown    map[string]float64    `json:"breakdown"`
	ProviderData map[string]interface{} `json:"provider_data,omitempty"`
}

type TeamInviteRequest struct {
	Emails    []string `json:"emails"`
	Role      string   `json:"role"`
	Message   string   `json:"message,omitempty"`
}

type TeamInviteResponse struct {
	Invites []TeamInvite `json:"invites"`
}

type TeamInvite struct {
	Email   string `json:"email"`
	Token   string `json:"token"`
	Expires int64  `json:"expires"`
}

func NewHandler(repo storage.Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

// HandleValidateProvider validates a provider API token
func (h *Handler) HandleValidateProvider(w http.ResponseWriter, r *http.Request) {
	var req ProviderValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user from context
	user, ok := r.Context().Value("user").(*storage.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response := h.validateProviderToken(req.Provider, req.Token)

	// Store validated provider if successful
	if response.IsValid {
		// Generate a random ID for the provider
		idBytes := make([]byte, 16)
		rand.Read(idBytes)
		providerID := hex.EncodeToString(idBytes)

		provider := &storage.Provider{
			ID:       providerID,
			UserID:   user.ID,
			Provider: req.Provider,
			Token:    req.Token, // Token will be encrypted by the repository layer
			Status:   "active",
		}

		if err := h.repo.CreateProvider(provider); err != nil {
			logrus.WithError(err).Error("Failed to store provider")
			response.IsValid = false
			response.Message = "Failed to save provider configuration"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleEstimateCost provides cost estimation for function deployment
func (h *Handler) HandleEstimateCost(w http.ResponseWriter, r *http.Request) {
	var req CostEstimationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user from context
	user, ok := r.Context().Value("user").(*storage.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user has access to this provider
	provider, err := h.repo.GetProviderByUserAndType(user.ID, req.Provider)
	if err != nil {
		http.Error(w, "Provider not configured", http.StatusBadRequest)
		return
	}

	if provider.Status != "active" {
		http.Error(w, "Provider not active", http.StatusBadRequest)
		return
	}

	estimation := h.estimateCost(req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(estimation)
}

// HandleCreateTeamInvite creates team invitations during onboarding
func (h *Handler) HandleCreateTeamInvite(w http.ResponseWriter, r *http.Request) {
	var req TeamInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user from context
	user, ok := r.Context().Value("user").(*storage.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Create team if user doesn't have one
	team, err := h.repo.GetTeamByUserID(user.ID)
	if err != nil {
		// Create new team for user
		team = &storage.Team{
			Name:      fmt.Sprintf("%s's Team", user.Email),
			TenantID:  user.TenantID,
			CreatedBy: user.ID,
		}
		if err := h.repo.CreateTeam(team); err != nil {
			logrus.WithError(err).Error("Failed to create team")
			http.Error(w, "Failed to create team", http.StatusInternalServerError)
			return
		}

		// Add user as admin to team
		teamMember := &storage.TeamMembership{
			TeamID: team.ID,
			UserID: user.ID,
			Role:   "admin",
		}
		if err := h.repo.AddTeamMember(teamMember); err != nil {
			logrus.WithError(err).Error("Failed to add user to team")
			http.Error(w, "Failed to setup team", http.StatusInternalServerError)
			return
		}
	}

	var invites []TeamInvite
	for _, email := range req.Emails {
		// Create invitation token
		token, expires := auth.GenerateInviteToken()

		invite := &storage.TeamInvite{
			TeamID:    team.ID,
			Email:     email,
			Token:     token,
			Role:      req.Role,
			InvitedBy: user.ID,
			ExpiresAt: expires,
			Message:   req.Message,
		}

		if err := h.repo.CreateTeamInvite(invite); err != nil {
			logrus.WithError(err).WithField("email", email).Error("Failed to create team invite")
			continue
		}

		invites = append(invites, TeamInvite{
			Email:   email,
			Token:   token,
			Expires: expires.Unix(),
		})
	}

	response := TeamInviteResponse{
		Invites: invites,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleShareProvider shares a provider configuration with team members
func (h *Handler) HandleShareProvider(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	providerID := vars["providerId"]

	var req struct {
		TeamID string `json:"team_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user from context
	user, ok := r.Context().Value("user").(*storage.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Verify user owns the provider or is team admin
	provider, err := h.repo.GetProviderByID(providerID)
	if err != nil {
		http.Error(w, "Provider not found", http.StatusNotFound)
		return
	}

	if provider.UserID != user.ID {
		// Check if user is team admin
		isAdmin, err := h.repo.IsTeamAdmin(user.ID, req.TeamID)
		if err != nil || !isAdmin {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	// Share provider with team
	if err := h.repo.ShareProviderWithTeam(providerID, req.TeamID); err != nil {
		logrus.WithError(err).Error("Failed to share provider")
		http.Error(w, "Failed to share provider", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "shared"})
}

// validateProviderToken validates API tokens for different providers
func (h *Handler) validateProviderToken(provider, token string) ProviderValidationResponse {
	switch provider {
	case "cloudflare":
		return h.validateCloudflareToken(token)
	case "vercel":
		return h.validateVercelToken(token)
	case "fly":
		return h.validateFlyToken(token)
	default:
		return ProviderValidationResponse{
			IsValid: false,
			Message: "Unsupported provider",
		}
	}
}

// validateCloudflareToken validates Cloudflare API token
func (h *Handler) validateCloudflareToken(token string) ProviderValidationResponse {
	// Validate the token by making a request to Cloudflare's API
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Use the /user/tokens/verify endpoint to validate the token
	req, err := http.NewRequest("GET", "https://api.cloudflare.com/client/v4/user/tokens/verify", nil)
	if err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to create validation request: %v", err),
		}
	}

	// Set authorization header with the token
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to validate token: %v", err),
		}
	}
	defer resp.Body.Close()

	// Parse the response
	var result struct {
		Success bool `json:"success"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
		Result struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to parse validation response: %v", err),
		}
	}

	if !result.Success || len(result.Errors) > 0 {
		errorMsg := "token validation failed"
		if len(result.Errors) > 0 {
			errorMsg = result.Errors[0].Message
		}
		return ProviderValidationResponse{
			IsValid: false,
			Message: errorMsg,
		}
	}

	// Token is valid
	return ProviderValidationResponse{
		IsValid: true,
		Message: "Cloudflare token validated successfully",
		UserID:  result.Result.ID,
		Email:   "", // Cloudflare API doesn't return email in token verification
	}
}

// validateVercelToken validates Vercel API token
func (h *Handler) validateVercelToken(token string) ProviderValidationResponse {
	// Validate the token by making a request to Vercel's API
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Use the /v2/user endpoint to validate the token and get user info
	req, err := http.NewRequest("GET", "https://api.vercel.com/v2/user", nil)
	if err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to create validation request: %v", err),
		}
	}

	// Set authorization header with the token
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to validate token: %v", err),
		}
	}
	defer resp.Body.Close()

	// Parse the response
	var result struct {
		User struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"user"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to parse validation response: %v", err),
		}
	}

	// Check for API errors
	if result.Error.Code != "" {
		return ProviderValidationResponse{
			IsValid: false,
			Message: result.Error.Message,
		}
	}

	// Check if we got a successful response with user data
	if result.User.ID == "" {
		return ProviderValidationResponse{
			IsValid: false,
			Message: "invalid token or unable to retrieve user information",
		}
	}

	// Token is valid
	return ProviderValidationResponse{
		IsValid: true,
		Message: "Vercel token validated successfully",
		UserID:  result.User.ID,
		Email:   result.User.Email,
	}
}

// validateFlyToken validates Fly.io API token
func (h *Handler) validateFlyToken(token string) ProviderValidationResponse {
	// Validate the token by making a request to Fly.io's API
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Use the /api/v1/user endpoint to validate the token and get user info
	req, err := http.NewRequest("GET", "https://api.fly.io/api/v1/user", nil)
	if err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to create validation request: %v", err),
		}
	}

	// Set authorization header with the token
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to validate token: %v", err),
		}
	}
	defer resp.Body.Close()

	// Parse the response
	var result struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Error string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ProviderValidationResponse{
			IsValid: false,
			Message: fmt.Sprintf("failed to parse validation response: %v", err),
		}
	}

	// Check for API errors
	if result.Error != "" {
		return ProviderValidationResponse{
			IsValid: false,
			Message: result.Error,
		}
	}

	// Check if we got a successful response with user data
	if result.ID == "" {
		return ProviderValidationResponse{
			IsValid: false,
			Message: "invalid token or unable to retrieve user information",
		}
	}

	// Token is valid
	return ProviderValidationResponse{
		IsValid: true,
		Message: "Fly.io token validated successfully",
		UserID:  result.ID,
		Email:   result.Email,
	}
}

// estimateCost calculates cost estimation for function deployment
func (h *Handler) estimateCost(req CostEstimationRequest) CostEstimationResponse {
	// Base costs per provider (monthly estimates)
	baseCosts := map[string]float64{
		"cloudflare": 0.0,  // Free tier
		"vercel":     0.0,  // Free tier
		"fly":        2.67, // ~$2.67 for basic usage
	}

	// Compute costs per million requests
	computeCosts := map[string]float64{
		"cloudflare": 0.30, // $0.30 per million requests
		"vercel":     0.40, // $0.40 per million requests
		"fly":        0.22, // $0.22 per million requests
	}

	// Storage costs (per GB/month)
	storageCosts := map[string]float64{
		"cloudflare": 0.055,
		"vercel":     0.10,
		"fly":        0.15,
	}

	// Calculate requests per month
	requestsPerMonth := float64(req.RequestsPerDay) * 30

	// Calculate compute cost
	computeCost := (requestsPerMonth / 1000000) * computeCosts[req.Provider]

	// Calculate storage cost (assuming 10MB per function)
	storageCost := 0.01 * storageCosts[req.Provider]

	// Calculate bandwidth cost (assuming 1KB per request)
	bandwidthMB := (requestsPerMonth * 1024) / (1024 * 1024) // Convert to GB
	bandwidthCost := bandwidthMB * 0.09 // ~$0.09 per GB

	totalCost := baseCosts[req.Provider] + computeCost + storageCost + bandwidthCost

	breakdown := map[string]float64{
		"base":      baseCosts[req.Provider],
		"compute":   computeCost,
		"storage":   storageCost,
		"bandwidth": bandwidthCost,
	}

	return CostEstimationResponse{
		MonthlyCost: totalCost,
		Currency:    "USD",
		Breakdown:   breakdown,
		ProviderData: map[string]interface{}{
			"requests_per_month": requestsPerMonth,
			"estimated_bandwidth_gb": bandwidthMB,
			"regions_count": len(req.Regions),
		},
	}
}