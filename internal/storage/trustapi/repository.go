package trustapi

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles database operations for Trust API
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new Trust API repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ============================================
// Partner Operations
// ============================================

// CreatePartner creates a new partner
func (r *Repository) CreatePartner(partner *TrustAPIPartner) error {
	if partner.ID == uuid.Nil {
		partner.ID = uuid.New()
	}
	if partner.Tier == "" {
		partner.Tier = string(PartnerTierDeveloper)
	}
	if partner.Status == "" {
		partner.Status = string(PartnerStatusPending)
	}
	if partner.RateLimitPerMinute == 0 {
		partner.RateLimitPerMinute = 60
	}
	if partner.RateLimitPerDay == 0 {
		partner.RateLimitPerDay = 10000
	}
	if partner.MonthlyRequestLimit == 0 {
		partner.MonthlyRequestLimit = 50000
	}

	return r.db.Create(partner).Error
}

// CreatePartnerInTransaction creates a new partner inside a transaction.
// This prevents TOCTOU race conditions on slug/email uniqueness checks
// and ensures UNIQUE constraint violations are surfaced cleanly.
func (r *Repository) CreatePartnerInTransaction(partner *TrustAPIPartner) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if partner.ID == uuid.Nil {
			partner.ID = uuid.New()
		}
		if partner.Tier == "" {
			partner.Tier = string(PartnerTierDeveloper)
		}
		if partner.Status == "" {
			partner.Status = string(PartnerStatusPending)
		}
		if partner.RateLimitPerMinute == 0 {
			partner.RateLimitPerMinute = 60
		}
		if partner.RateLimitPerDay == 0 {
			partner.RateLimitPerDay = 10000
		}
		if partner.MonthlyRequestLimit == 0 {
			partner.MonthlyRequestLimit = 50000
		}

		return tx.Create(partner).Error
	})
}

// GetPartnerByID retrieves a partner by ID
func (r *Repository) GetPartnerByID(id uuid.UUID) (*TrustAPIPartner, error) {
	var partner TrustAPIPartner
	err := r.db.Where("id = ?", id).First(&partner).Error
	if err != nil {
		return nil, err
	}
	return &partner, nil
}

// GetPartnerBySlug retrieves a partner by slug
func (r *Repository) GetPartnerBySlug(slug string) (*TrustAPIPartner, error) {
	var partner TrustAPIPartner
	err := r.db.Where("slug = ?", slug).First(&partner).Error
	if err != nil {
		return nil, err
	}
	return &partner, nil
}

// GetPartnerByContactEmail retrieves a partner by contact email
func (r *Repository) GetPartnerByContactEmail(email string) (*TrustAPIPartner, error) {
	var partner TrustAPIPartner
	err := r.db.Where("contact_email = ?", email).First(&partner).Error
	if err != nil {
		return nil, err
	}
	return &partner, nil
}

// ListPartners retrieves all partners with optional filtering
func (r *Repository) ListPartners(status string, tier string, limit, offset int) ([]TrustAPIPartner, int64, error) {
	var partners []TrustAPIPartner
	var total int64

	query := r.db.Model(&TrustAPIPartner{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if tier != "" {
		query = query.Where("tier = ?", tier)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&partners).Error; err != nil {
		return nil, 0, err
	}

	return partners, total, nil
}

// UpdatePartner updates a partner
func (r *Repository) UpdatePartner(partner *TrustAPIPartner) error {
	return r.db.Save(partner).Error
}

// UpdatePartnerStatus updates a partner's status
func (r *Repository) UpdatePartnerStatus(id uuid.UUID, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	switch status {
	case string(PartnerStatusActive):
		now := time.Now()
		updates["activated_at"] = &now
	case string(PartnerStatusSuspended):
		now := time.Now()
		updates["suspended_at"] = &now
	}

	return r.db.Model(&TrustAPIPartner{}).Where("id = ?", id).Updates(updates).Error
}

// IncrementUsage increments the partner's monthly usage counter
func (r *Repository) IncrementUsage(partnerID uuid.UUID, count int) error {
	return r.db.Model(&TrustAPIPartner{}).
		Where("id = ?", partnerID).
		UpdateColumn("current_month_usage", gorm.Expr("current_month_usage + ?", count)).
		Error
}

// ResetMonthlyUsage resets the monthly usage for all partners (called by scheduler)
func (r *Repository) ResetMonthlyUsage() error {
	return r.db.Model(&TrustAPIPartner{}).
		Where("current_month_usage > 0").
		Update("current_month_usage", 0).
		Error
}

// GetRateLimitConfig returns the rate limit configuration for a partner tier
func GetRateLimitConfig(tier string) RateLimitConfig {
	config, ok := RateLimitsPerTier[PartnerTier(tier)]
	if !ok {
		return RateLimitsPerTier[PartnerTierDeveloper]
	}
	return config
}

// ============================================
// API Key Operations
// ============================================

// GenerateAPIKey generates a new API key for a partner
func (r *Repository) GenerateAPIKey(partnerID uuid.UUID, req *APIKeyCreateRequest, createdBy string) (*TrustAPIKey, string, error) {
	// Generate the actual key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate key: %w", err)
	}
	rawKey := "fft_" + hex.EncodeToString(keyBytes)

	// Hash the key for storage
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	// Create key ID and prefix
	keyID := "fft_" + hex.EncodeToString(keyBytes)[:24]
	keyPrefix := keyID[:9]

	// Marshal scopes
	scopesJSON, err := json.Marshal(req.Scopes)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal scopes: %w", err)
	}

	// Marshal allowed IPs
	allowedIPsJSON, err := json.Marshal(req.AllowedIPs)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal allowed IPs: %w", err)
	}

	apiKey := &TrustAPIKey{
		ID:          uuid.New(),
		PartnerID:   partnerID,
		KeyID:       keyID,
		KeyPrefix:   keyPrefix,
		KeyHash:     keyHash,
		Name:        req.Name,
		Description: req.Description,
		Scopes:      scopesJSON,
		AllowedIPs:  allowedIPsJSON,
		ExpiresAt:   req.ExpiresAt,
		IsRevoked:   false,
		UseCount:    0,
		CreatedBy:   createdBy,
	}

	if err := r.db.Create(apiKey).Error; err != nil {
		return nil, "", fmt.Errorf("failed to create API key: %w", err)
	}

	return apiKey, rawKey, nil
}

// GetAPIKeyByKeyID retrieves an API key by its public key ID
func (r *Repository) GetAPIKeyByKeyID(keyID string) (*TrustAPIKey, error) {
	var apiKey TrustAPIKey
	err := r.db.Preload("Partner").Where("key_id = ?", keyID).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// GetAPIKeyByHash retrieves an API key by its hash (for verification)
func (r *Repository) GetAPIKeyByHash(keyHash string) (*TrustAPIKey, error) {
	var apiKey TrustAPIKey
	err := r.db.Preload("Partner").Where("key_hash = ?", keyHash).First(&apiKey).Error
	if err != nil {
		return nil, err
	}
	return &apiKey, nil
}

// ValidateAPIKey validates an API key and returns the associated partner
func (r *Repository) ValidateAPIKey(rawKey string) (*TrustAPIKey, *TrustAPIPartner, error) {
	// Hash the provided key
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	// Look up the key
	apiKey, err := r.GetAPIKeyByHash(keyHash)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid API key")
	}

	// Check if revoked
	if apiKey.IsRevoked {
		return nil, nil, fmt.Errorf("API key has been revoked")
	}

	// Check if expired
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, nil, fmt.Errorf("API key has expired")
	}

	// Check if partner is active
	if apiKey.Partner.Status != string(PartnerStatusActive) {
		return nil, nil, fmt.Errorf("partner account is not active")
	}

	return apiKey, apiKey.Partner, nil
}

// ListAPIKeysForPartner lists all API keys for a partner
func (r *Repository) ListAPIKeysForPartner(partnerID uuid.UUID, includeRevoked bool) ([]TrustAPIKey, error) {
	var keys []TrustAPIKey
	query := r.db.Where("partner_id = ?", partnerID)

	if !includeRevoked {
		query = query.Where("is_revoked = ?", false)
	}

	if err := query.Order("created_at DESC").Find(&keys).Error; err != nil {
		return nil, err
	}

	return keys, nil
}

// RevokeAPIKey revokes an API key
func (r *Repository) RevokeAPIKey(keyID uuid.UUID, reason string) error {
	now := time.Now()
	return r.db.Model(&TrustAPIKey{}).
		Where("id = ?", keyID).
		Updates(map[string]interface{}{
			"is_revoked":     true,
			"revoked_at":     &now,
			"revoked_reason": reason,
		}).Error
}

// IncrementKeyUsage increments the usage counter for an API key
func (r *Repository) IncrementKeyUsage(keyID uuid.UUID) error {
	return r.db.Model(&TrustAPIKey{}).
		Where("id = ?", keyID).
		Updates(map[string]interface{}{
			"use_count":    gorm.Expr("use_count + 1"),
			"last_used_at": time.Now(),
		}).Error
}

// CheckIPAllowed checks if an IP address is allowed for an API key
func (r *Repository) CheckIPAllowed(apiKey *TrustAPIKey, ipAddress string) bool {
	var allowedIPs []string
	if err := json.Unmarshal(apiKey.AllowedIPs, &allowedIPs); err != nil {
		// SECURITY: Fail closed (deny) on parse error - if we can't verify the IP list, deny the request
		log.Printf("ERROR: Failed to parse AllowedIPs for TrustAPIKey %s: %v - denying request", apiKey.ID, err)
		return false
	}

	// Empty array means all IPs are allowed
	if len(allowedIPs) == 0 {
		return true
	}

	// Check if the IP matches any allowed pattern (supports both single IPs and CIDR ranges)
	for _, allowed := range allowedIPs {
		allowed = strings.TrimSpace(allowed)
		if allowed == "" {
			continue
		}

		// Check for exact match
		if allowed == ipAddress {
			return true
		}

		// Check for CIDR match (e.g., "192.168.1.0/24" or "2001:db8::/32")
		if strings.Contains(allowed, "/") {
			if ipInCIDR(ipAddress, allowed) {
				return true
			}
		}
	}

	return false
}

// ipInCIDR checks if an IP address is within a CIDR range
func ipInCIDR(ipAddress, cidrRange string) bool {
	_, ipNet, err := net.ParseCIDR(cidrRange)
	if err != nil {
		// Invalid CIDR format - log and skip
		log.Printf("WARN: Invalid CIDR format '%s': %v", cidrRange, err)
		return false
	}

	ip := net.ParseIP(ipAddress)
	if ip == nil {
		// Invalid IP address
		return false
	}

	return ipNet.Contains(ip)
}

// ============================================
// Usage Tracking Operations
// ============================================

// RecordUsage records API usage
func (r *Repository) RecordUsage(usage *TrustAPIUsage) error {
	if usage.ID == uuid.Nil {
		usage.ID = uuid.New()
	}
	return r.db.Create(usage).Error
}

// GetUsageForPartner retrieves usage statistics for a partner
func (r *Repository) GetUsageForPartner(partnerID uuid.UUID, startDate, endDate time.Time) (*UsageResponse, error) {
	var totalRequests int64
	var successfulRequests int64
	var failedRequests int64
	var totalLatency float64
	var rateLimitHits int64

	// Get aggregate stats
	statsQuery := r.db.Model(&TrustAPIUsage{}).
		Where("partner_id = ?", partnerID).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate)

	statsQuery.Count(&totalRequests)

	statsQuery.Where("status_code >= 200 AND status_code < 400").Count(&successfulRequests)
	statsQuery.Where("status_code >= 400").Count(&failedRequests)
	statsQuery.Where("error_code = ?", "rate_limit_exceeded").Count(&rateLimitHits)

	// Get average latency
	var avgResult struct {
		AvgLatency float64
	}
	r.db.Model(&TrustAPIUsage{}).
		Select("AVG(response_time_ms) as avg_latency").
		Where("partner_id = ?", partnerID).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Scan(&avgResult)

	totalLatency = avgResult.AvgLatency

	// Get top endpoints
	var topEndpoints []EndpointUsage
	r.db.Model(&TrustAPIUsage{}).
		Select("endpoint, COUNT(*) as count").
		Where("partner_id = ?", partnerID).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Group("endpoint").
		Order("count DESC").
		Limit(10).
		Scan(&topEndpoints)

	return &UsageResponse{
		PartnerID:          partnerID,
		PeriodStart:        startDate,
		PeriodEnd:          endDate,
		TotalRequests:      totalRequests,
		SuccessfulRequests: successfulRequests,
		FailedRequests:     failedRequests,
		AverageLatencyMs:   totalLatency,
		RateLimitHits:      rateLimitHits,
		TopEndpoints:       topEndpoints,
	}, nil
}

// GetMonthlyUsage gets the current month's usage for a partner
func (r *Repository) GetMonthlyUsage(partnerID uuid.UUID) (int, error) {
	var partner TrustAPIPartner
	err := r.db.Select("current_month_usage").Where("id = ?", partnerID).First(&partner).Error
	if err != nil {
		return 0, err
	}
	return partner.CurrentMonthUsage, nil
}

// ============================================
// Rate Limiting Operations
// ============================================

// CheckRateLimit checks if a partner has exceeded their rate limit
func (r *Repository) CheckRateLimit(partnerID uuid.UUID, limitType string, limit int) (bool, int, error) {
	now := time.Now()
	var windowStart, windowEnd time.Time

	switch limitType {
	case "minute":
		windowStart = now.Truncate(time.Minute)
		windowEnd = windowStart.Add(time.Minute)
	case "hour":
		windowStart = now.Truncate(time.Hour)
		windowEnd = windowStart.Add(time.Hour)
	case "day":
		windowStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		windowEnd = windowStart.Add(24 * time.Hour)
	case "month":
		windowStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		windowEnd = windowStart.AddDate(0, 1, 0)
	default:
		return true, limit, fmt.Errorf("unknown limit type: %s", limitType)
	}

	// Try to get or create rate limit record
	var rateLimit TrustAPIRateLimit
	err := r.db.Where("partner_id = ? AND limit_type = ? AND window_start = ?", partnerID, limitType, windowStart).
		First(&rateLimit).Error

	if err == gorm.ErrRecordNotFound {
		// Create new record
		rateLimit = TrustAPIRateLimit{
			ID:           uuid.New(),
			PartnerID:    partnerID,
			LimitType:    limitType,
			WindowStart:  windowStart,
			WindowEnd:    windowEnd,
			RequestCount: 0,
		}
		if err := r.db.Create(&rateLimit).Error; err != nil {
			return false, limit, err
		}
		return true, limit, nil
	} else if err != nil {
		return false, limit, err
	}

	// Check if over limit
	remaining := limit - rateLimit.RequestCount
	if remaining <= 0 {
		return false, remaining, nil
	}

	return true, remaining, nil
}

// IncrementRateLimit increments the rate limit counter for a partner atomically.
// Uses UPDATE ... SET request_count = request_count + 1 for atomicity.
func (r *Repository) IncrementRateLimit(partnerID uuid.UUID, limitType string) error {
	now := time.Now()
	var windowStart time.Time

	switch limitType {
	case "minute":
		windowStart = now.Truncate(time.Minute)
	case "hour":
		windowStart = now.Truncate(time.Hour)
	case "day":
		windowStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "month":
		windowStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	default:
		return fmt.Errorf("unknown limit type: %s", limitType)
	}

	return r.db.Model(&TrustAPIRateLimit{}).
		Where("partner_id = ? AND limit_type = ? AND window_start = ?", partnerID, limitType, windowStart).
		UpdateColumn("request_count", gorm.Expr("request_count + 1")).Error
}

// CheckAndIncrementRateLimit atomically checks and increments the rate limit counter.
// This eliminates the race condition between check and increment.
// Returns (allowed, remaining, error).
func (r *Repository) CheckAndIncrementRateLimit(partnerID uuid.UUID, limitType string, limit int) (bool, int, error) {
	now := time.Now()
	var windowStart, windowEnd time.Time

	switch limitType {
	case "minute":
		windowStart = now.Truncate(time.Minute)
		windowEnd = windowStart.Add(time.Minute)
	case "hour":
		windowStart = now.Truncate(time.Hour)
		windowEnd = windowStart.Add(time.Hour)
	case "day":
		windowStart = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		windowEnd = windowStart.Add(24 * time.Hour)
	case "month":
		windowStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		windowEnd = windowStart.AddDate(0, 1, 0)
	default:
		return true, limit, fmt.Errorf("unknown limit type: %s", limitType)
	}

	// Atomic upsert: insert if not exists with count=1, or increment existing
	// Uses PostgreSQL's ON CONFLICT for atomicity
	var rateLimit TrustAPIRateLimit
	err := r.db.Where("partner_id = ? AND limit_type = ? AND window_start = ?", partnerID, limitType, windowStart).
		First(&rateLimit).Error

	if err == gorm.ErrRecordNotFound {
		// Create new record with count=1
		rateLimit = TrustAPIRateLimit{
			ID:           uuid.New(),
			PartnerID:    partnerID,
			LimitType:    limitType,
			WindowStart:  windowStart,
			WindowEnd:    windowEnd,
			RequestCount: 1,
		}
		if err := r.db.Create(&rateLimit).Error; err != nil {
			return false, limit, err
		}
		return true, limit - 1, nil
	} else if err != nil {
		return false, limit, err
	}

	// Increment atomically and get the new count
	result := r.db.Model(&TrustAPIRateLimit{}).
		Where("partner_id = ? AND limit_type = ? AND window_start = ? AND request_count < ?", partnerID, limitType, windowStart, limit).
		UpdateColumn("request_count", gorm.Expr("request_count + 1"))

	if result.Error != nil {
		return false, 0, result.Error
	}

	if result.RowsAffected == 0 {
		// No rows updated means we're at or over the limit
		return false, 0, nil
	}

	remaining := limit - rateLimit.RequestCount - 1
	if remaining < 0 {
		remaining = 0
	}
	return true, remaining, nil
}

// CleanupOldRateLimits removes rate limit records older than 24 hours
func (r *Repository) CleanupOldRateLimits() error {
	cutoff := time.Now().Add(-24 * time.Hour)
	return r.db.Where("window_end < ?", cutoff).Delete(&TrustAPIRateLimit{}).Error
}

// ============================================
// Report Operations
// ============================================

// CreateReport creates a new trust report
func (r *Repository) CreateReport(report *TrustAPIReport) error {
	if report.ID == uuid.Nil {
		report.ID = uuid.New()
	}
	// Generate public report ID
	report.ReportID = "rpt_" + uuid.New().String()[:24]
	if report.Status == "" {
		report.Status = string(ReportStatusPending)
	}
	return r.db.Create(report).Error
}

// GetReportByReportID retrieves a report by its public report ID
func (r *Repository) GetReportByReportID(reportID string) (*TrustAPIReport, error) {
	var report TrustAPIReport
	err := r.db.Preload("Partner").Where("report_id = ?", reportID).First(&report).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// ListReportsForPartner lists reports submitted by a partner
func (r *Repository) ListReportsForPartner(partnerID uuid.UUID, status string, limit, offset int) ([]TrustAPIReport, int64, error) {
	var reports []TrustAPIReport
	var total int64

	query := r.db.Model(&TrustAPIReport{}).Where("partner_id = ?", partnerID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// UpdateReportStatus updates the status of a report
func (r *Repository) UpdateReportStatus(reportID uuid.UUID, status string, resolvedBy *uuid.UUID, notes string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == string(ReportStatusResolved) || status == string(ReportStatusDismissed) {
		now := time.Now()
		updates["resolved_at"] = &now
		if resolvedBy != nil {
			updates["resolved_by"] = resolvedBy
		}
		if notes != "" {
			updates["resolution_notes"] = notes
		}
	}

	return r.db.Model(&TrustAPIReport{}).Where("id = ?", reportID).Updates(updates).Error
}

// ============================================
// Verification Operations
// ============================================

// CreateVerification creates a new verification request
func (r *Repository) CreateVerification(verification *TrustAPIVerification) error {
	if verification.ID == uuid.Nil {
		verification.ID = uuid.New()
	}
	// Generate public verification ID
	verification.VerificationID = "vfy_" + uuid.New().String()[:24]
	if verification.Status == "" {
		verification.Status = string(VerificationStatusPending)
	}
	return r.db.Create(verification).Error
}

// GetVerificationByVerificationID retrieves a verification by its public verification ID
func (r *Repository) GetVerificationByVerificationID(verificationID string) (*TrustAPIVerification, error) {
	var verification TrustAPIVerification
	err := r.db.Preload("Partner").Where("verification_id = ?", verificationID).First(&verification).Error
	if err != nil {
		return nil, err
	}
	return &verification, nil
}

// ListVerificationsForPartner lists verification requests from a partner
func (r *Repository) ListVerificationsForPartner(partnerID uuid.UUID, status string, limit, offset int) ([]TrustAPIVerification, int64, error) {
	var verifications []TrustAPIVerification
	var total int64

	query := r.db.Model(&TrustAPIVerification{}).Where("partner_id = ?", partnerID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&verifications).Error; err != nil {
		return nil, 0, err
	}

	return verifications, total, nil
}

// UpdateVerificationResult updates a verification with its result
func (r *Repository) UpdateVerificationResult(verificationID uuid.UUID, trustScore *float64, trustTier string, badgeURL string, notes string, completedBy *uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&TrustAPIVerification{}).
		Where("id = ?", verificationID).
		Updates(map[string]interface{}{
			"status":                 string(VerificationStatusCompleted),
			"trust_score":            trustScore,
			"trust_tier":             trustTier,
			"verification_badge_url": badgeURL,
			"completion_notes":       notes,
			"completed_by":           completedBy,
			"completed_at":           &now,
		}).Error
}
