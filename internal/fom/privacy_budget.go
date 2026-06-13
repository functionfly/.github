package fom

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrBudgetExhausted = errors.New("privacy budget exhausted")

type PrivacyBudgetService struct {
	repo    *Repository
	db      *sql.DB
	periodDays int
}

func NewPrivacyBudgetService(repo *Repository, db *sql.DB) *PrivacyBudgetService {
	return &PrivacyBudgetService{
		repo:       repo,
		db:         db,
		periodDays: 30,
	}
}

func (s *PrivacyBudgetService) GetOrCreateBudget(ctx context.Context, tenantID uuid.UUID, tier UserTier) (*PrivacyBudget, error) {
	budget, err := s.repo.GetPrivacyBudget(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	if budget != nil {
		if time.Now().After(budget.PeriodEnd) {
			return s.refreshBudget(ctx, budget)
		}
		return budget, nil
	}

	newBudget := &PrivacyBudget{
		TenantID:    tenantID,
		TotalBudget: GetDefaultBudget(tier),
		UsedBudget:  0,
		PeriodStart: time.Now().UTC(),
		PeriodEnd:   time.Now().UTC().AddDate(0, 0, s.periodDays),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := s.repo.CreateOrUpdatePrivacyBudget(ctx, newBudget); err != nil {
		return nil, err
	}

	return newBudget, nil
}

func (s *PrivacyBudgetService) CheckAndConsume(ctx context.Context, tenantID uuid.UUID, tier UserTier) (bool, error) {
	budget, err := s.GetOrCreateBudget(ctx, tenantID, tier)
	if err != nil {
		return false, err
	}

	if budget.IsExhausted() {
		return false, ErrBudgetExhausted
	}

	consumed, err := s.repo.ConsumePrivacyBudget(ctx, tenantID)
	if err != nil {
		return false, err
	}

	return consumed, nil
}

func (s *PrivacyBudgetService) GetRemainingBudget(ctx context.Context, tenantID uuid.UUID, tier UserTier) (int, error) {
	budget, err := s.GetOrCreateBudget(ctx, tenantID, tier)
	if err != nil {
		return 0, err
	}

	return budget.Remaining(), nil
}

func (s *PrivacyBudgetService) refreshBudget(ctx context.Context, budget *PrivacyBudget) (*PrivacyBudget, error) {
	budget.UsedBudget = 0
	budget.PeriodStart = time.Now().UTC()
	budget.PeriodEnd = time.Now().UTC().AddDate(0, 0, s.periodDays)
	budget.UpdatedAt = time.Now().UTC()

	if err := s.repo.CreateOrUpdatePrivacyBudget(ctx, budget); err != nil {
		return nil, err
	}

	return budget, nil
}

func (s *PrivacyBudgetService) SetBudgetLimit(ctx context.Context, tenantID uuid.UUID, limit int) error {
	budget, err := s.repo.GetPrivacyBudget(ctx, tenantID)
	if err != nil {
		return err
	}

	if budget == nil {
		budget = &PrivacyBudget{
			TenantID:    tenantID,
			TotalBudget: limit,
			UsedBudget:  0,
			PeriodStart: time.Now().UTC(),
			PeriodEnd:   time.Now().UTC().AddDate(0, 0, s.periodDays),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}
	} else {
		budget.TotalBudget = limit
		budget.UpdatedAt = time.Now().UTC()
	}

	return s.repo.CreateOrUpdatePrivacyBudget(ctx, budget)
}

type BudgetInfo struct {
	TenantID     uuid.UUID `json:"tenant_id"`
	TotalBudget  int       `json:"total_budget"`
	UsedBudget   int       `json:"used_budget"`
	Remaining    int       `json:"remaining"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
	DaysLeft     int       `json:"days_left"`
	IsExhausted  bool      `json:"is_exhausted"`
}

func (s *PrivacyBudgetService) GetBudgetInfo(ctx context.Context, tenantID uuid.UUID, tier UserTier) (*BudgetInfo, error) {
	budget, err := s.GetOrCreateBudget(ctx, tenantID, tier)
	if err != nil {
		return nil, err
	}

	daysLeft := int(time.Until(budget.PeriodEnd).Hours() / 24)
	if daysLeft < 0 {
		daysLeft = 0
	}

	return &BudgetInfo{
		TenantID:    budget.TenantID,
		TotalBudget: budget.TotalBudget,
		UsedBudget:  budget.UsedBudget,
		Remaining:   budget.Remaining(),
		PeriodStart: budget.PeriodStart,
		PeriodEnd:   budget.PeriodEnd,
		DaysLeft:    daysLeft,
		IsExhausted: budget.IsExhausted(),
	}, nil
}

type BudgetAllocation struct {
	Tier       UserTier
	Limit      int
	PeriodDays int
}

func DefaultBudgetAllocations() []BudgetAllocation {
	return []BudgetAllocation{
		{UserTierFree, 1000, 30},
		{UserTierPro, 10000, 30},
		{UserTierEnterprise, 100000, 30},
	}
}

func (s *PrivacyBudgetService) EnforceBudgetMiddleware(ctx context.Context, tenantID uuid.UUID, tier UserTier) error {
	consumed, err := s.CheckAndConsume(ctx, tenantID, tier)
	if err != nil {
		if errors.Is(err, ErrBudgetExhausted) {
			return ErrBudgetExhausted
		}
		return err
	}

	if !consumed {
		return ErrBudgetExhausted
	}

	return nil
}