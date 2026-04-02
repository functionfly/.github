package identity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// JSONBMap scans from and writes to PostgreSQL JSONB for map fields (e.g. capabilities).
type JSONBMap map[string]any

// Scan implements sql.Scanner for JSONB.
func (m *JSONBMap) Scan(value interface{}) error {
	if value == nil {
		*m = make(JSONBMap)
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("expected []byte or string for JSONB")
	}
	if len(b) == 0 {
		*m = make(JSONBMap)
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return err
	}
	if out == nil {
		out = make(map[string]any)
	}
	*m = JSONBMap(out)
	return nil
}

// Value implements driver.Valuer for JSONB.
func (m JSONBMap) Value() (driver.Value, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

// AgentIdentity represents a registered agent in the system
type AgentIdentity struct {
	ID                uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID          uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`
	AgentID           string    `json:"agent_id" gorm:"uniqueIndex;not null"` // "org/agent-name"
	Name              string    `json:"name" gorm:"not null"`
	Description       string    `json:"description"`
	PlanTier          string    `json:"plan_tier" gorm:"not null;default:'agent_starter'"`
	Status            string    `json:"status" gorm:"not null;default:'active'"` // active | suspended | deleted
	APIKeyHash        string    `json:"-" gorm:"column:api_key_hash"`            // hashed API key, never returned
	ParentAgentID     *string   `json:"parent_agent_id" gorm:"column:parent_agent_id"`
	SwarmRole         string    `json:"swarm_role" gorm:"not null;default:'worker'"` // worker | manager | infrastructure
	MaxChildAgents    int       `json:"max_child_agents" gorm:"not null;default:0"`
	Capabilities      JSONBMap  `json:"capabilities" gorm:"type:jsonb;default:'{}'"`
	AutonomousEnabled bool      `json:"autonomous_enabled" gorm:"not null;default:false"`
	EvolutionEnabled  bool      `json:"evolution_enabled" gorm:"not null;default:false"`
	TrustScore        float64   `json:"trust_score" gorm:"type:decimal(5,2);default:0"`
	EconomicScore     float64   `json:"economic_score" gorm:"type:decimal(5,2);default:0"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name
func (AgentIdentity) TableName() string {
	return "agent_identities"
}

// AgentQuotaConfig holds per-agent quota configuration
type AgentQuotaConfig struct {
	ID                  uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID             string    `json:"agent_id" gorm:"uniqueIndex;not null"`
	MaxCallsPerMinute   int       `json:"max_calls_per_minute" gorm:"not null;default:100"`
	MaxCallsPerDay      int       `json:"max_calls_per_day" gorm:"not null;default:16667"`
	MaxStateWritesPerHr int       `json:"max_state_writes_per_hr" gorm:"not null;default:1000"`
	MaxCostPerExecution float64   `json:"max_cost_per_execution" gorm:"type:decimal(10,6);not null;default:0.01"`
	MaxDailySpendUSD    float64   `json:"max_daily_spend_usd" gorm:"type:decimal(10,2);not null;default:5.00"`
	AllowedFunctions    []string  `json:"allowed_functions,omitempty" gorm:"type:text[]"`
	ForbiddenFunctions  []string  `json:"forbidden_functions,omitempty" gorm:"type:text[]"`
	CreatedAt           time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name
func (AgentQuotaConfig) TableName() string {
	return "agent_quota_configs"
}

// RegisterAgentRequest is the request body for registering a new agent
type RegisterAgentRequest struct {
	AgentID     string `json:"agent_id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
	PlanTier    string `json:"plan_tier,omitempty"`
}

// RegisterAgentResponse is the response after registering an agent
type RegisterAgentResponse struct {
	OK      bool           `json:"ok"`
	Agent   *AgentIdentity `json:"agent"`
	APIKey  string         `json:"api_key,omitempty"` // Only returned on creation
	Message string         `json:"message,omitempty"`
}

// AgentStatus constants
const (
	AgentStatusActive    = "active"
	AgentStatusSuspended = "suspended"
	AgentStatusDeleted   = "deleted"
)

// SwarmRole constants
const (
	SwarmRoleWorker         = "worker"
	SwarmRoleManager        = "manager"
	SwarmRoleInfrastructure = "infrastructure"
)

// MessageType constants for A2A communication
const (
	MessageTypeTaskDelegation      = "task_delegation"
	MessageTypeTaskResult          = "task_result"
	MessageTypeQuery               = "query"
	MessageTypeResponse            = "response"
	MessageTypeCapabilityDiscovery = "capability_discovery"
	MessageTypeHeartbeat           = "heartbeat"
	MessageTypeEvolutionProposal   = "evolution_proposal"
	MessageTypeBudgetRequest       = "budget_request"
)

// PricingModel constants
const (
	PricingModelFree         = "free"
	PricingModelPerCall      = "per_call"
	PricingModelSubscription = "subscription"
	PricingModelRevenueShare = "revenue_share"
)

// TransactionType constants
const (
	TransactionTypeDelegationPayment = "delegation_payment"
	TransactionTypeFunctionCall      = "function_call"
	TransactionTypeSubscription      = "subscription"
	TransactionTypeRevenueShare      = "revenue_share"
	TransactionTypeRefund            = "refund"
)

// EvolutionProposalType constants
const (
	EvolutionTypeSpawnSpecialist     = "spawn_specialist"
	EvolutionTypeModifyPolicy        = "modify_policy"
	EvolutionTypeAdjustTimeout       = "adjust_timeout"
	EvolutionTypeGenerateFunction    = "generate_function"
	EvolutionTypeRetireChild         = "retire_child"
	EvolutionTypeUpgradeCapabilities = "upgrade_capabilities"
)

// AutonomyScheduleType constants
const (
	AutonomyScheduleRecurring    = "recurring"
	AutonomyScheduleOneTime      = "one_time"
	AutonomyScheduleTriggerBased = "trigger_based"
)

// AutonomyActionType constants
const (
	AutonomyActionExecuteFunction = "execute_function"
	AutonomyActionSpawnAgent      = "spawn_agent"
	AutonomyActionSendMessage     = "send_message"
	AutonomyActionUpdateState     = "update_state"
	AutonomyActionEvolve          = "evolve"
)

// AgentRelationship represents a parent-child relationship between agents
type AgentRelationship struct {
	ID                 uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ParentAgentID      string    `json:"parent_agent_id" gorm:"not null"`
	ChildAgentID       string    `json:"child_agent_id" gorm:"not null"`
	RelationshipType   string    `json:"relationship_type" gorm:"not null;default:'parent'"` // parent | supervisor | collaborator
	MaxDelegationDepth int       `json:"max_delegation_depth" gorm:"not null;default:5"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the GORM table name
func (AgentRelationship) TableName() string {
	return "agent_relationships"
}

// AgentMessage represents a message between agents
type AgentMessage struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FromAgentID string         `json:"from_agent_id" gorm:"not null"`
	ToAgentID   string         `json:"to_agent_id" gorm:"not null"`
	MessageType string         `json:"message_type" gorm:"not null"`
	Payload     map[string]any `json:"payload" gorm:"type:jsonb;default:'{}'"`
	SessionID   *string        `json:"session_id"`
	TTLSeconds  int            `json:"ttl_seconds" gorm:"not null;default:3600"`
	Status      string         `json:"status" gorm:"not null;default:'pending'"` // pending | delivered | read | expired | failed
	CreatedAt   time.Time      `json:"created_at" gorm:"autoCreateTime"`
	DeliveredAt *time.Time     `json:"delivered_at"`
}

// TableName returns the GORM table name
func (AgentMessage) TableName() string {
	return "agent_messages"
}

// AgentWallet represents the economic wallet for an agent
type AgentWallet struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID          string     `json:"agent_id" gorm:"uniqueIndex;not null"`
	BalanceUSD       float64    `json:"balance_usd" gorm:"type:decimal(12,2);not null;default:0"`
	EscrowBalanceUSD float64    `json:"escrow_balance_usd" gorm:"type:decimal(12,2);not null;default:0"`
	TotalEarnedUSD   float64    `json:"total_earned_usd" gorm:"type:decimal(12,2);not null;default:0"`
	TotalSpentUSD    float64    `json:"total_spent_usd" gorm:"type:decimal(12,2);not null;default:0"`
	LastEarningAt    *time.Time `json:"last_earning_at"`
	LastSpendingAt   *time.Time `json:"last_spending_at"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name
func (AgentWallet) TableName() string {
	return "agent_wallets"
}

// AgentListing represents a marketplace listing for an agent
type AgentListing struct {
	ID                     uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID                string         `json:"agent_id" gorm:"uniqueIndex;not null"`
	Agent                  *AgentIdentity `json:"agent,omitempty" gorm:"foreignKey:AgentID;references:AgentID"`
	ListingType            string         `json:"listing_type" gorm:"not null;default:'worker'"` // worker | manager | infrastructure
	PricingModel           string         `json:"pricing_model" gorm:"not null;default:'per_call'"` // free | per_call | subscription | revenue_share | tiered | dynamic | auction
	PricePerCall           *float64       `json:"price_per_call" gorm:"type:decimal(10,4)"`
	SubscriptionMonthlyUSD  *float64       `json:"subscription_monthly_usd" gorm:"type:decimal(10,2)"`
	RevenueSharePercent    *float64       `json:"revenue_share_percent" gorm:"type:decimal(5,2)"`
	// Tiered pricing
	Tiers                  []PricingTier  `json:"tiers" gorm:"-"`
	TiersJSON              string         `json:"-" gorm:"column:tiers;type:jsonb"`
	// Dynamic pricing
	DynamicMinPrice        *float64       `json:"dynamic_min_price" gorm:"type:decimal(10,4)"`
	DynamicMaxPrice        *float64       `json:"dynamic_max_price" gorm:"type:decimal(10,4)"`
	DynamicDemandFactor    *float64       `json:"dynamic_demand_factor" gorm:"type:decimal(5,2)"`
	// Auction pricing
	AuctionStartPrice      *float64       `json:"auction_start_price" gorm:"type:decimal(10,4)"`
	AuctionReservePrice    *float64       `json:"auction_reserve_price" gorm:"type:decimal(10,4)"`
	AuctionEndTime         *time.Time     `json:"auction_end_time"`
	AuctionCurrentBid      *float64       `json:"auction_current_bid" gorm:"type:decimal(10,4)"`
	AuctionBidCount        int            `json:"auction_bid_count" gorm:"default:0"`
	//
	RatingScore            float64        `json:"rating_score" gorm:"type:decimal(3,2);default:0"`
	TotalCalls             int            `json:"total_calls" gorm:"not null;default:0"`
	ROIScore               float64        `json:"roi_score" gorm:"type:decimal(5,2);default:0"`
	IsActive               bool           `json:"is_active" gorm:"not null;default:true"`
	ListedAt               time.Time      `json:"listed_at" gorm:"autoCreateTime"`
	UpdatedAt              time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// PricingTier represents a volume-based pricing tier
type PricingTier struct {
	CallsPerMonth int     `json:"calls_per_month" gorm:"not null"`
	PricePerCall  float64 `json:"price_per_call" gorm:"type:decimal(10,4);not null"`
	DiscountPct   float64 `json:"discount_pct" gorm:"type:decimal(5,2);default:0"`
}

// TableName returns the GORM table name
func (AgentListing) TableName() string {
	return "agent_listings"
}

// FunctionListing represents a marketplace listing for a function
type FunctionListing struct {
	ID                     uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FunctionID             uuid.UUID `json:"function_id" gorm:"not null"`
	PricingModel           string    `json:"pricing_model" gorm:"not null;default:'per_call'"` // free | per_call | subscription | revenue_share | tiered | dynamic | auction
	PricePerCall           *float64  `json:"price_per_call" gorm:"type:decimal(10,4)"`
	SubscriptionMonthlyUSD  *float64  `json:"subscription_monthly_usd" gorm:"type:decimal(10,2)"`
	RevenueSharePercent     *float64  `json:"revenue_share_percent" gorm:"type:decimal(5,2)"`
	// Tiered pricing
	Tiers                  []PricingTier `json:"tiers" gorm:"-"`
	TiersJSON              string        `json:"-" gorm:"column:tiers;type:jsonb"`
	// Dynamic pricing
	DynamicMinPrice        *float64      `json:"dynamic_min_price" gorm:"type:decimal(10,4)"`
	DynamicMaxPrice        *float64      `json:"dynamic_max_price" gorm:"type:decimal(10,4)"`
	DynamicDemandFactor    *float64      `json:"dynamic_demand_factor" gorm:"type:decimal(5,2)"`
	// Auction pricing
	AuctionStartPrice      *float64      `json:"auction_start_price" gorm:"type:decimal(10,4)"`
	AuctionReservePrice    *float64      `json:"auction_reserve_price" gorm:"type:decimal(10,4)"`
	AuctionEndTime         *time.Time    `json:"auction_end_time"`
	AuctionCurrentBid      *float64      `json:"auction_current_bid" gorm:"type:decimal(10,4)"`
	AuctionBidCount        int           `json:"auction_bid_count" gorm:"default:0"`
	//
	IsActive               bool         `json:"is_active" gorm:"not null;default:true"`
	RatingScore            float64      `json:"rating_score" gorm:"type:decimal(3,2);default:0"`
	CallVolume             int          `json:"call_volume" gorm:"not null;default:0"`
	DeterministicVerified  bool         `json:"deterministic_verified" gorm:"not null;default:false"`
	ListedAt               time.Time    `json:"listed_at" gorm:"autoCreateTime"`
	UpdatedAt              time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name
func (FunctionListing) TableName() string {
	return "function_listings"
}

// RevenueTransaction represents a financial transaction in the agent economy
type RevenueTransaction struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FromAgentID       *string    `json:"from_agent_id"`
	ToAgentID         string     `json:"to_agent_id" gorm:"not null"`
	FunctionID        *uuid.UUID `json:"function_id"`
	AmountUSD         float64    `json:"amount_usd" gorm:"type:decimal(10,4);not null"`
	TransactionType   string     `json:"transaction_type" gorm:"not null"`
	SessionID         *string    `json:"session_id"`
	ExecutionID       *string    `json:"execution_id"`
	ParentExecutionID *string    `json:"parent_execution_id"`
	Status            string     `json:"status" gorm:"not null;default:'completed'"` // pending | completed | failed | refunded
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// TableName returns the GORM table name
func (RevenueTransaction) TableName() string {
	return "agent_revenue_transactions"
}

// EvolutionProposal represents an agent evolution proposal
type EvolutionProposal struct {
	ID                     uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID                string         `json:"agent_id" gorm:"not null"`
	ProposalType           string         `json:"proposal_type" gorm:"not null"`
	ProposalData           map[string]any `json:"proposal_data" gorm:"type:jsonb;not null;default:'{}'"`
	Status                 string         `json:"status" gorm:"not null;default:'pending'"` // pending | approved | rejected | implemented | expired
	ParentApprovalRequired bool           `json:"parent_approval_required" gorm:"not null;default:true"`
	SimulatedOutcome       map[string]any `json:"simulated_outcome" gorm:"type:jsonb"`
	ApprovedBy             *string        `json:"approved_by"`
	ImplementedAt          *time.Time     `json:"implemented_at"`
	CreatedAt              time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt              time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name
func (EvolutionProposal) TableName() string {
	return "agent_evolution_proposals"
}

// SetProposalData sets the ProposalData field from a generic map.
// It marshals the map to JSON and stores it, allowing dynamic proposal data to be persisted.
func (p *EvolutionProposal) SetProposalData(data map[string]any) error {
	if data == nil {
		p.ProposalData = make(map[string]any)
		return nil
	}
	p.ProposalData = data
	return nil
}

// AutonomySchedule represents a scheduled or triggered execution
type AutonomySchedule struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AgentID          string         `json:"agent_id" gorm:"not null"`
	ScheduleType     string         `json:"schedule_type" gorm:"not null"` // recurring | one_time | trigger_based
	CronExpression   *string        `json:"cron_expression"`
	TriggerEvent     *string        `json:"trigger_event"`
	TriggerCondition map[string]any `json:"trigger_condition" gorm:"type:jsonb"`
	ActionType       string         `json:"action_type" gorm:"not null"` // execute_function | spawn_agent | send_message | update_state | evolve
	ActionPayload    map[string]any `json:"action_payload" gorm:"type:jsonb;not null;default:'{}'"`
	IsActive         bool           `json:"is_active" gorm:"not null;default:true"`
	NextRunAt        *time.Time     `json:"next_run_at"`
	LastRunAt        *time.Time     `json:"last_run_at"`
	CreatedAt        time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name
func (AutonomySchedule) TableName() string {
	return "agent_autonomy_schedules"
}

// Function is a minimal GORM model for marketplace/agent use (e.g. FunctionListing FK).
// The canonical function model is storage/registry.RegistryFunction (table: registry_functions).
type Function struct {
	ID                    string    `json:"id" gorm:"primaryKey"`
	Author                string    `json:"author" gorm:"not null"`
	Name                  string    `json:"name" gorm:"not null"`
	LatestVersion         string    `json:"latest_version,omitempty"`
	Title                 string    `json:"title,omitempty"`
	Description           string    `json:"description,omitempty"`
	Category              string    `json:"category,omitempty"`
	Tags                  []string  `json:"tags,omitempty" gorm:"type:text[]"`
	Visibility            string    `json:"visibility" gorm:"default:'public'"`
	PricePerCall          float64   `json:"price_per_call" gorm:"type:decimal(20,8);default:0"`
	PopularityScore       int       `json:"popularity_score" gorm:"default:0"`
	ReliabilityScore      float64   `json:"reliability_score" gorm:"type:decimal(5,2);default:0"`
	DeterministicScore    float64   `json:"deterministic_score" gorm:"type:decimal(5,2);default:0"`
	TenantID              *string   `json:"tenant_id,omitempty"`
	OwnerUserID           *string   `json:"owner_user_id,omitempty"`
	OwnerAgentID          *string   `json:"owner_agent_id,omitempty"`
	AgentGenerated        bool      `json:"agent_generated" gorm:"default:false"`
	GenerationPromptHash  *string   `json:"generation_prompt_hash,omitempty"`
	GenerationModel       *string   `json:"generation_model,omitempty"`
	DeterministicCertHash *string   `json:"deterministic_cert_hash,omitempty"`
	RevenueTotalUSD       float64   `json:"revenue_total_usd" gorm:"type:decimal(12,2);default:0"`
	CreatedAt             time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the GORM table name
func (Function) TableName() string {
	return "registry_functions"
}
