package marketplace

import (
	"context"
	"encoding/json"

	agentmarketplace "github.com/functionfly/functionfly/internal/agent/marketplace"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
)

type AgentSearcherAdapter struct {
	svc *agentmarketplace.Service
}

func NewAgentSearcherAdapter(svc *agentmarketplace.Service) *AgentSearcherAdapter {
	return &AgentSearcherAdapter{svc: svc}
}

func (a *AgentSearcherAdapter) SearchAgents(ctx context.Context, req UnifiedSearchRequest) ([]UnifiedAgentResult, int64, error) {
	agentReq := &agentmarketplace.SearchAgentsRequest{
		Capabilities: splitQuery(req.Query),
		SortBy:       "rating_score",
		Limit:        req.Limit,
		Offset:       req.Offset,
	}
	results, total, err := a.svc.SearchAgents(ctx, agentReq)
	if err != nil {
		return nil, 0, err
	}
	out := make([]UnifiedAgentResult, len(results))
	for i, r := range results {
		out[i] = agentListingToUnified(r)
	}
	return out, total, nil
}

func agentListingToUnified(r agentmarketplace.AgentSearchResult) UnifiedAgentResult {
	l := r.Listing
	var name, desc string
	var caps []string
	var verified bool
	if l.Agent != nil {
		name = l.Agent.Name
		desc = l.Agent.Description
		for k := range l.Agent.Capabilities {
			caps = append(caps, k)
		}
	}
	if name == "" {
		name = l.AgentID
	}
	return UnifiedAgentResult{
		AgentID:      l.AgentID,
		Name:         name,
		Description:  desc,
		Capabilities: caps,
		PricingModel: l.PricingModel,
		PricePerCall: l.PricePerCall,
		SubMonthly:   l.SubscriptionMonthlyUSD,
		RatingScore:  l.RatingScore,
		TotalCalls:   l.TotalCalls,
		ROIScore:     l.ROIScore,
		ListingType:  l.ListingType,
		Verified:     verified,
		RankScore:    r.RankScore,
	}
}

func (a *AgentSearcherAdapter) SearchFunctions(ctx context.Context, req UnifiedSearchRequest) ([]UnifiedFunctionResult, int64, error) {
	funcReq := &agentmarketplace.SearchFunctionsRequest{
		Limit:  req.Limit,
		Offset: req.Offset,
	}
	results, total, err := a.svc.SearchFunctions(ctx, funcReq)
	if err != nil {
		return nil, 0, err
	}
	out := make([]UnifiedFunctionResult, len(results))
	for i, r := range results {
		out[i] = functionListingToUnified(r)
	}
	return out, total, nil
}

func functionListingToUnified(r agentmarketplace.FunctionSearchResult) UnifiedFunctionResult {
	l := r.Listing
	name := l.FunctionID.String()
	var desc string
	if l.Function != nil {
		name = l.Function.Name
		desc = l.Function.Description
	}
	return UnifiedFunctionResult{
		FunctionID:   l.FunctionID.String(),
		Name:         name,
		Description:  desc,
		Runtime:      "",
		PricingModel: l.PricingModel,
		PricePerCall: l.PricePerCall,
		SubMonthly:   l.SubscriptionMonthlyUSD,
		RatingScore:  l.RatingScore,
		CallVolume:   l.CallVolume,
		Verified:     l.DeterministicVerified,
	}
}

func splitQuery(q string) []string {
	if q == "" {
		return nil
	}
	return []string{q}
}

var _ AgentSearcher = (*AgentSearcherAdapter)(nil)

// AgentRaterAdapter adapts marketplace.Service to the AgentRater interface
type AgentRaterAdapter struct {
	svc *agentmarketplace.Service
}

func NewAgentRaterAdapter(svc *agentmarketplace.Service) *AgentRaterAdapter {
	return &AgentRaterAdapter{svc: svc}
}

func (a *AgentRaterAdapter) RateAgent(ctx context.Context, agentID string, tenantID uuid.UUID, rating int, review string) error {
	return a.svc.RateAgent(ctx, agentID, tenantID, rating, review)
}

func (a *AgentRaterAdapter) GetAgentRating(ctx context.Context, agentID string, tenantID uuid.UUID) (*AgentRatingResult, error) {
	r, err := a.svc.GetAgentRating(ctx, agentID, tenantID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return &AgentRatingResult{
		ID:        r.ID.String(),
		AgentID:   r.AgentID,
		TenantID:  r.TenantID.String(),
		Rating:    r.Rating,
		Review:    r.Review,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}, nil
}

func (a *AgentRaterAdapter) ListAgentRatings(ctx context.Context, agentID string, limit int) ([]AgentRatingResult, error) {
	ratings, err := a.svc.ListAgentRatings(ctx, agentID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]AgentRatingResult, len(ratings))
	for i, r := range ratings {
		out[i] = AgentRatingResult{
			ID:        r.ID.String(),
			AgentID:   r.AgentID,
			TenantID:  r.TenantID.String(),
			Rating:    r.Rating,
			Review:    r.Review,
			Username:  r.Username,
			UserName:  r.UserName,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		}
	}
	return out, nil
}

var _ AgentRater = (*AgentRaterAdapter)(nil)

// FunctionRaterAdapter adapts marketplace.Service to the FunctionRater interface
type FunctionRaterAdapter struct {
	svc *agentmarketplace.Service
}

func NewFunctionRaterAdapter(svc *agentmarketplace.Service) *FunctionRaterAdapter {
	return &FunctionRaterAdapter{svc: svc}
}

func (a *FunctionRaterAdapter) RateFunction(ctx context.Context, functionID uuid.UUID, tenantID uuid.UUID, rating int, review string) error {
	return a.svc.RateFunction(ctx, functionID, tenantID, rating, review)
}

func (a *FunctionRaterAdapter) GetFunctionRating(ctx context.Context, functionID uuid.UUID, tenantID uuid.UUID) (*FunctionRatingResult, error) {
	r, err := a.svc.GetFunctionRating(ctx, functionID, tenantID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	return &FunctionRatingResult{
		ID:         r.ID.String(),
		FunctionID: r.FunctionID.String(),
		TenantID:   r.TenantID.String(),
		Rating:     r.Rating,
		Review:     r.Review,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}, nil
}

func (a *FunctionRaterAdapter) ListFunctionRatings(ctx context.Context, functionID uuid.UUID, limit int) ([]FunctionRatingResult, error) {
	ratings, err := a.svc.ListFunctionRatings(ctx, functionID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]FunctionRatingResult, len(ratings))
	for i, r := range ratings {
		out[i] = FunctionRatingResult{
			ID:         r.ID.String(),
			FunctionID: r.FunctionID.String(),
			TenantID:   r.TenantID.String(),
			Rating:     r.Rating,
			Review:     r.Review,
			Username:   r.Username,
			UserName:   r.UserName,
			CreatedAt:  r.CreatedAt,
			UpdatedAt:  r.UpdatedAt,
		}
	}
	return out, nil
}

var _ FunctionRater = (*FunctionRaterAdapter)(nil)

type FunctionSearcherAdapter struct {
	repo *registry.RegistryRepository
}

func NewFunctionSearcherAdapter(repo *registry.RegistryRepository) *FunctionSearcherAdapter {
	return &FunctionSearcherAdapter{repo: repo}
}

func (f *FunctionSearcherAdapter) SearchFunctions(ctx context.Context, req UnifiedSearchRequest) ([]UnifiedFunctionResult, int, error) {
	functions, total, err := f.repo.SearchFunctionsWithSort(req.Query, "", "", 0, req.Limit, req.Offset, "")
	if err != nil {
		return nil, 0, err
	}
	out := make([]UnifiedFunctionResult, len(functions))
	for i, fn := range functions {
		name := fn.Name
		desc := ""
		if fn.Description.Valid {
			desc = fn.Description.String
		}
		author := fn.Author
		category := ""
		if fn.Category.Valid {
			category = fn.Category.String
		}

		var pricingModel string
		var pricePerCall *float64
		if fn.PricePerCall > 0 {
			pricingModel = "per_call"
			pricePerCall = &fn.PricePerCall
		} else {
			pricingModel = "free"
		}

		rating := fn.ReliabilityScore

		var tags []string
		if fn.Tags != nil {
			_ = json.Unmarshal(fn.Tags, &tags)
		}

		out[i] = UnifiedFunctionResult{
			FunctionID:   fn.ID.String(),
			Name:         name,
			Description:  desc,
			Author:       author,
			Category:     category,
			Runtime:      fn.Region,
			PricingModel: pricingModel,
			PricePerCall: pricePerCall,
			SubMonthly:   nil,
			RatingScore:  rating,
			CallVolume:   fn.PopularityScore,
			Verified:     fn.TrustScore > 0,
			Tags:         tags,
		}
	}
	return out, total, nil
}

var _ FunctionSearcher = (*FunctionSearcherAdapter)(nil)
