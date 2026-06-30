package users

import (
	"net/http"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/apierror"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ReferralCodeResponse struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	CommissionType  string  `json:"commission_type"`
	CommissionValue float64 `json:"commission_value"`
	IsActive     bool   `json:"is_active"`
	ShareURL     string `json:"share_url"`
}

type ReferralStatsResponse struct {
	TotalReferrals    int     `json:"total_referrals"`
	ConvertedCount    int     `json:"converted_count"`
	PendingCount      int     `json:"pending_count"`
	ConversionRate    float64 `json:"conversion_rate"`
	TotalCommissionEarned int64 `json:"total_commission_earned_cents"`
	PendingCommission int64   `json:"pending_commission_cents"`
	AverageCommission float64 `json:"average_commission_cents"`
}

const foundersSignupURL = "https://functionfly.com/signup"

func (h *Handler) HandleGetMyReferralCode(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	ctx := r.Context()

	user, err := h.repo.GetUserByID(ctx, claims.UserID)
	if err != nil || user == nil {
		logrus.WithError(err).Error("Failed to get user for referral code")
		apierror.WriteError(w, apierror.NewInternal("Failed to get user"))
		return
	}

	codes, err := h.repo.ListAffiliateCodesByPublisher(ctx, claims.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list affiliate codes")
		apierror.WriteError(w, apierror.NewInternal("Failed to get referral code"))
		return
	}

	var code *storage.AffiliateCode
	for _, c := range codes {
		if c.Name == "Founder Referral" && c.IsActive {
			code = c
			break
		}
	}

	if code == nil && len(codes) > 0 {
		code = codes[0]
	}

	if code == nil {
		code = &storage.AffiliateCode{
			Code:            generateAffiliateCode(claims.UserID),
			PublisherID:      claims.UserID,
			Name:            "Founder Referral",
			CommissionType:  "percent",
			CommissionValue: 10.0,
			IsActive:        true,
		}

		created, err := h.repo.CreateAffiliateCode(ctx, code)
		if err != nil {
			logrus.WithError(err).Error("Failed to create affiliate code")
			apierror.WriteError(w, apierror.NewInternal("Failed to create referral code"))
			return
		}
		code = created
	}

	refSlug := code.Code
	if user.Username != nil && *user.Username != "" {
		refSlug = "f-" + *user.Username
	}

	response := ReferralCodeResponse{
		Code:            code.Code,
		Name:            code.Name,
		CommissionType:  code.CommissionType,
		CommissionValue: code.CommissionValue,
		IsActive:        code.IsActive,
		ShareURL:        foundersSignupURL + "?ref=" + refSlug,
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) HandleGetMyReferralStats(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r)
	if claims == nil {
		apierror.WriteError(w, apierror.NewUnauthorized("Unauthorized"))
		return
	}

	ctx := r.Context()

	codes, err := h.repo.ListAffiliateCodesByPublisher(ctx, claims.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list affiliate codes")
		apierror.WriteError(w, apierror.NewInternal("Failed to get referral stats"))
		return
	}

	if len(codes) == 0 {
		writeJSON(w, http.StatusOK, ReferralStatsResponse{
			TotalReferrals:    0,
			ConvertedCount:    0,
			PendingCount:      0,
			ConversionRate:    0,
			TotalCommissionEarned: 0,
			PendingCommission: 0,
			AverageCommission: 0,
		})
		return
	}

	var code *storage.AffiliateCode
	for _, c := range codes {
		if c.Name == "Founder Referral" {
			code = c
			break
		}
	}
	if code == nil {
		code = codes[0]
	}

	referrals, err := h.repo.ListAffiliateReferralsByCode(ctx, code.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list referrals")
		apierror.WriteError(w, apierror.NewInternal("Failed to get referral stats"))
		return
	}

	commissions, err := h.repo.ListAffiliateCommissionsByCode(ctx, code.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to list commissions")
		apierror.WriteError(w, apierror.NewInternal("Failed to get referral stats"))
		return
	}

	var totalCommission int64
	var pendingCommission int64
	for _, c := range commissions {
		totalCommission += c.CommissionCents
		if c.Status == "pending" {
			pendingCommission += c.CommissionCents
		}
	}

	totalReferrals := len(referrals)
	convertedCount := 0
	for _, r := range referrals {
		if r.Status == "converted" || r.Status == "qualified" {
			convertedCount++
		}
	}

	pendingCount := totalReferrals - convertedCount
	var conversionRate float64
	if totalReferrals > 0 {
		conversionRate = float64(convertedCount) / float64(totalReferrals) * 100
	}

	var avgCommission float64
	if convertedCount > 0 {
		avgCommission = float64(totalCommission) / float64(convertedCount) / 100.0
	}

	response := ReferralStatsResponse{
		TotalReferrals:        totalReferrals,
		ConvertedCount:        convertedCount,
		PendingCount:          pendingCount,
		ConversionRate:        conversionRate,
		TotalCommissionEarned: totalCommission,
		PendingCommission:     pendingCommission,
		AverageCommission:     avgCommission,
	}

	writeJSON(w, http.StatusOK, response)
}

func generateAffiliateCode(userID uuid.UUID) string {
	return "F" + userID.String()[0:8]
}
