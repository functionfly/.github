package providers

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
)

// ProviderValidationRequest validates a provider API token
type ProviderValidationRequest struct {
	Provider string `json:"provider"`
	Token    string `json:"token"`
}

// ProviderValidationResponse is the result of token validation
type ProviderValidationResponse struct {
	IsValid bool   `json:"is_valid"`
	Message string `json:"message,omitempty"`
	UserID  string `json:"user_id,omitempty"`
	Email   string `json:"email,omitempty"`
}

// CostEstimationRequest calculates deployment costs
type CostEstimationRequest struct {
	Provider        string   `json:"provider"`
	FunctionName    string   `json:"function_name"`
	Runtime         string   `json:"runtime"`
	MemoryMB        int      `json:"memory_mb"`
	RequestsPerDay  int      `json:"requests_per_day"`
	ComputeDuration int      `json:"compute_duration_ms"`
	Regions         []string `json:"regions"`
}

// CostEstimationResponse contains cost breakdown
type CostEstimationResponse struct {
	MonthlyCost  float64                `json:"monthly_cost"`
	Currency     string                 `json:"currency"`
	Breakdown    map[string]float64     `json:"breakdown"`
	ProviderData map[string]interface{} `json:"provider_data,omitempty"`
}

// TeamInviteRequest creates team invitations
type TeamInviteRequest struct {
	Emails  []string `json:"emails"`
	Role    string   `json:"role"`
	Message string   `json:"message,omitempty"`
}

// TeamInviteResponse returns created invites
type TeamInviteResponse struct {
	Invites []TeamInvite `json:"invites"`
}

// TeamInvite represents a single invitation
type TeamInvite struct {
	Email   string `json:"email"`
	Token   string `json:"token"`
	Expires int64  `json:"expires"`
}

// Handler manages provider operations
type Handler struct {
	repo   storage.Repository
	notify *notification.Service
}

// NewHandler creates a new provider handler
func NewHandler(repo storage.Repository, notify *notification.Service) *Handler {
	return &Handler{
		repo:   repo,
		notify: notify,
	}
}

// HTTPHandler defines the interface for provider HTTP handlers
type HTTPHandler interface {
	HandleConnectProvider(w http.ResponseWriter, r *http.Request)
	HandleDisconnectProvider(w http.ResponseWriter, r *http.Request)
	HandleTestConnection(w http.ResponseWriter, r *http.Request)
	HandleListProviders(w http.ResponseWriter, r *http.Request)
	HandleValidateProvider(w http.ResponseWriter, r *http.Request)
	HandleEstimateCost(w http.ResponseWriter, r *http.Request)
	HandleCreateTeamInvite(w http.ResponseWriter, r *http.Request)
	HandleShareProvider(w http.ResponseWriter, r *http.Request)
}

// Ensure Handler implements HTTPHandler
var _ HTTPHandler = (*Handler)(nil)
