package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v83"
	stripeDispute "github.com/stripe/stripe-go/v83/dispute"
	"github.com/stripe/stripe-go/v83/refund"
)

type DisputeResponseManager struct {
	disputeRepo     *storage.DisputeRepository
	notificationSvc  DisputeResponseNotificationService
	stripeKey       string
	stop            chan struct{}
	config          *DisputeAutomationConfig
}

type DisputeResponseNotificationService interface {
	SendBillingAlert(ctx context.Context, email, alertType string, data map[string]interface{}) error
	Send(ctx context.Context, req notification.SendRequest) (*notification.Notification, error)
	SendDisputeCreated(ctx context.Context, adminUserIDs []uuid.UUID, disputeID, amountUSD, currency, reason, evidenceDueBy string) error
	SendDisputeEvidenceDueSoon(ctx context.Context, adminUserIDs []uuid.UUID, disputeID string, daysRemaining int) error
	SendDisputeResolved(ctx context.Context, adminUserIDs []uuid.UUID, disputeID, outcome string, amountUSD float64, won bool) error
	SendAutoRefundExecuted(ctx context.Context, adminUserIDs []uuid.UUID, disputeID string, amountUSD float64, reason string) error
	SendDisputeEvidenceSubmitted(ctx context.Context, adminUserIDs []uuid.UUID, disputeID string) error
}

type DisputeAutomationConfig struct {
	AutoRefundEnabled              bool
	AutoRefundThresholdCents      int
	AutoRefundAllowedReasons      []string
	EvidenceAutoSubmit            bool
	CustomerNotificationEnabled    bool
	AdminEscalationEnabled        bool
	AdminEscalationThresholdCents int
}

func NewDisputeResponseManager(
	disputeRepo *storage.DisputeRepository,
	notificationSvc DisputeResponseNotificationService,
) *DisputeResponseManager {
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")

	config := &DisputeAutomationConfig{
		AutoRefundEnabled:              getEnvBool("CHARGEBACK_AUTO_REFUND_ENABLED", true),
		AutoRefundThresholdCents:        getEnvInt("CHARGEBACK_AUTO_REFUND_THRESHOLD_CENTS", 2500), // $25
		AutoRefundAllowedReasons:        getEnvList("CHARGEBACK_AUTO_REFUND_ALLOWED_REASONS", []string{"duplicate", "product_not_received"}),
		EvidenceAutoSubmit:              getEnvBool("CHARGEBACK_EVIDENCE_AUTO_SUBMIT", false), // Manual review by default
		CustomerNotificationEnabled:     getEnvBool("CHARGEBACK_CUSTOMER_NOTIFICATION_ENABLED", true),
		AdminEscalationEnabled:         getEnvBool("CHARGEBACK_ADMIN_ESCALATION_ENABLED", true),
		AdminEscalationThresholdCents:  getEnvInt("CHARGEBACK_ADMIN_ESCALATION_THRESHOLD_CENTS", 15000), // $150
	}

	return &DisputeResponseManager{
		disputeRepo:     disputeRepo,
		notificationSvc: notificationSvc,
		stripeKey:       stripeKey,
		stop:            make(chan struct{}),
		config:          config,
	}
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		return strings.ToLower(val) == "true" || val == "1"
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		var i int
		if _, err := fmt.Sscanf(val, "%d", &i); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvList(key string, defaultVal []string) []string {
	if val := os.Getenv(key); val != "" {
		return strings.Split(val, ",")
	}
	return defaultVal
}

func (m *DisputeResponseManager) IsStripeConfigured() bool {
	return m.stripeKey != ""
}

func (m *DisputeResponseManager) GetConfig() *DisputeAutomationConfig {
	return m.config
}

func (m *DisputeResponseManager) UpdateConfig(config *DisputeAutomationConfig) {
	m.config = config
}

func (m *DisputeResponseManager) Stop() {
	close(m.stop)
}

func (m *DisputeResponseManager) StopChan() <-chan struct{} {
	return m.stop
}

func (m *DisputeResponseManager) HandleDisputeCreated(ctx context.Context, dispute *stripe.Dispute, paymentDispute *storage.PaymentDispute) error {
	logrus.WithFields(logrus.Fields{
		"dispute_id":   dispute.ID,
		"amount_cents": dispute.Amount,
		"reason":       dispute.Reason,
		"status":       dispute.Status,
	}).Info("DisputeResponseManager: processing new dispute")

	if err := m.disputeRepo.UpsertDispute(ctx, paymentDispute); err != nil {
		return fmt.Errorf("failed to upsert dispute: %w", err)
	}

	m.logAutomationAction(ctx, paymentDispute.ID, "dispute_received", "success", map[string]interface{}{
		"amount_cents": dispute.Amount,
		"reason":       dispute.Reason,
		"status":       dispute.Status,
	})

	if m.config.AdminEscalationEnabled {
		m.notifyAdminNewDispute(ctx, dispute, paymentDispute)
	}

	if m.shouldAutoRefund(dispute) {
		if err := m.executeAutoRefund(ctx, dispute, paymentDispute); err != nil {
			logrus.WithError(err).WithField("dispute_id", dispute.ID).Error("DisputeResponseManager: auto-refund failed")
		}
	} else if m.config.EvidenceAutoSubmit {
		if err := m.submitEvidence(ctx, dispute, paymentDispute); err != nil {
			logrus.WithError(err).WithField("dispute_id", dispute.ID).Error("DisputeResponseManager: evidence submission failed")
		}
	}

	return nil
}

func (m *DisputeResponseManager) HandleDisputeUpdated(ctx context.Context, dispute *stripe.Dispute, paymentDispute *storage.PaymentDispute) error {
	logrus.WithFields(logrus.Fields{
		"dispute_id": dispute.ID,
		"status":     dispute.Status,
	}).Info("DisputeResponseManager: processing dispute update")

	existingDispute, err := m.disputeRepo.GetDisputeByStripeID(ctx, dispute.ID)
	if err != nil || existingDispute == nil {
		return fmt.Errorf("dispute not found: %w", err)
	}

	newStatus := string(dispute.Status)
	if existingDispute.Status != newStatus {
		if err := m.disputeRepo.UpdateDisputeStatus(ctx, existingDispute.ID, newStatus, "", ""); err != nil {
			return fmt.Errorf("failed to update dispute status: %w", err)
		}

		m.logAutomationAction(ctx, existingDispute.ID, "status_updated", "success", map[string]interface{}{
			"old_status": existingDispute.Status,
			"new_status": newStatus,
		})

		if m.config.CustomerNotificationEnabled {
			m.notifyCustomerStatusChange(ctx, existingDispute, newStatus)
		}
	}

	return nil
}

func (m *DisputeResponseManager) HandleDisputeClosed(ctx context.Context, dispute *stripe.Dispute, paymentDispute *storage.PaymentDispute) error {
	logrus.WithFields(logrus.Fields{
		"dispute_id": dispute.ID,
		"status":     dispute.Status,
	}).Info("DisputeResponseManager: processing dispute closed")

	existingDispute, err := m.disputeRepo.GetDisputeByStripeID(ctx, dispute.ID)
	if err != nil || existingDispute == nil {
		return fmt.Errorf("dispute not found: %w", err)
	}

	status := string(dispute.Status)
	outcome := status
	if err := m.disputeRepo.UpdateDisputeStatus(ctx, existingDispute.ID, status, outcome, ""); err != nil {
		return fmt.Errorf("failed to update dispute status: %w", err)
	}

	m.logAutomationAction(ctx, existingDispute.ID, "dispute_closed", "success", map[string]interface{}{
		"outcome": outcome,
		"status":  status,
	})

	amountUSD := float64(dispute.Amount) / 100.0
	won := status == "won" || status == "charge_refunded"

	if m.config.CustomerNotificationEnabled {
		m.notifyCustomerDisputeResolved(ctx, existingDispute, won, amountUSD)
	}

	m.notifyAdminDisputeResolved(ctx, existingDispute, status, amountUSD)

	return nil
}

func (m *DisputeResponseManager) shouldAutoRefund(dispute *stripe.Dispute) bool {
	if !m.config.AutoRefundEnabled {
		return false
	}

	if dispute.Amount > int64(m.config.AutoRefundThresholdCents) {
		return false
	}

	reason := string(dispute.Reason)
	for _, allowed := range m.config.AutoRefundAllowedReasons {
		if reason == allowed {
			return true
		}
	}

	return false
}

func (m *DisputeResponseManager) executeAutoRefund(ctx context.Context, dispute *stripe.Dispute, paymentDispute *storage.PaymentDispute) error {
	amountUSD := float64(dispute.Amount) / 100.0

	logrus.WithFields(logrus.Fields{
		"dispute_id": dispute.ID,
		"charge_id":  dispute.Charge.ID,
		"amount_usd": amountUSD,
		"reason":    dispute.Reason,
	}).Info("DisputeResponseManager: executing auto-refund")

	refundParams := &stripe.RefundParams{
		Charge: stripe.String(dispute.Charge.ID),
	}

	reasonStr := string(dispute.Reason)
	if dispute.Reason == stripe.DisputeReasonDuplicate {
		refundParams.Reason = stripe.String(string(stripe.RefundReasonDuplicate))
	} else {
		refundParams.Reason = stripe.String(string(stripe.RefundReasonRequestedByCustomer))
	}

	result, err := refund.New(refundParams)
	if err != nil {
		m.logAutomationAction(ctx, paymentDispute.ID, "auto_refund", "failed", map[string]interface{}{
			"error":        err.Error(),
			"amount_cents": dispute.Amount,
		})
		m.notifyAdminAutoRefundFailed(ctx, paymentDispute, amountUSD, err.Error())
		return fmt.Errorf("failed to submit refund to stripe: %w", err)
	}

	m.logAutomationAction(ctx, paymentDispute.ID, "auto_refund", "success", map[string]interface{}{
		"refund_id":     result.ID,
		"amount_cents": dispute.Amount,
	})

	m.notifyAdminAutoRefundExecuted(ctx, paymentDispute, amountUSD, reasonStr)

	logrus.WithFields(logrus.Fields{
		"dispute_id": dispute.ID,
		"refund_id":  result.ID,
	}).Info("DisputeResponseManager: auto-refund successful")

	return nil
}

func (m *DisputeResponseManager) submitEvidence(ctx context.Context, dispute *stripe.Dispute, paymentDispute *storage.PaymentDispute) error {
	evidence, err := m.compileEvidenceData(ctx, paymentDispute)
	if err != nil {
		m.logAutomationAction(ctx, paymentDispute.ID, "evidence_compilation", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to compile evidence: %w", err)
	}

	params := &stripe.DisputeParams{
		Evidence: &stripe.DisputeEvidenceParams{
			AccessActivityLog:        stripe.String(evidence.AccessActivityLog),
			BillingAddress:           stripe.String(evidence.BillingAddress),
			CancellationPolicy:       stripe.String(evidence.RefundPolicyURL),
			CancellationPolicyDisclosure: stripe.String(evidence.CancellationReason),
			CustomerCommunication:     stripe.String(evidence.CustomerCommunication),
			CustomerEmailAddress:     stripe.String(evidence.CustomerEmail),
			CustomerName:             stripe.String(evidence.CustomerName),
			CustomerPurchaseIP:       stripe.String(evidence.CustomerPurchaseIP),
			ProductDescription:       stripe.String(evidence.ProductDescription),
			Receipt:                  stripe.String(evidence.ReceiptURL),
			RefundPolicy:             stripe.String(evidence.RefundPolicyURL),
			RefundPolicyDisclosure:   stripe.String(evidence.RefundPolicyDisclosedText),
			ServiceDate:              stripe.String(evidence.ServiceDate),
			ServiceDocumentation:     stripe.String(evidence.ServiceDocument),
			ShippingAddress:          stripe.String(evidence.ShippingAddress),
			ShippingCarrier:         stripe.String(evidence.ShippingCarrier),
			ShippingDate:             stripe.String(evidence.ShippingDate),
			ShippingTrackingNumber:   stripe.String(evidence.ShippingTracking),
		},
		Submit: stripe.Bool(true),
	}

	_, err = stripeDispute.Update(dispute.ID, params)
	if err != nil {
		m.logAutomationAction(ctx, paymentDispute.ID, "evidence_submission", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to submit evidence to stripe: %w", err)
	}

	evidenceDetails := &storage.EvidenceDetails{
		ProductDescription:  evidence.ProductDescription,
		CustomerEmail:       evidence.CustomerEmail,
		CustomerName:        evidence.CustomerName,
		CustomerPurchaseIP: evidence.CustomerPurchaseIP,
		BillingAddress:     evidence.BillingAddress,
		ReceiptURL:         evidence.ReceiptURL,
		ServiceDate:        evidence.ServiceDate,
		ServiceDocument:    evidence.ServiceDocument,
		ShippingAddress:    evidence.ShippingAddress,
		ShippingDate:       evidence.ShippingDate,
		ShippingTracking:   evidence.ShippingTracking,
		ShippingCarrier:    evidence.ShippingCarrier,
		RefundPolicyURL:   evidence.RefundPolicyURL,
		RefundPolicyDisclosed: evidence.RefundPolicyDisclosed,
		CancellationReason: evidence.CancellationReason,
		AccessActivityLog: evidence.AccessActivityLog,
	}

	if err := m.disputeRepo.UpdateDisputeEvidence(ctx, paymentDispute.ID, evidenceDetails); err != nil {
		logrus.WithError(err).WithField("dispute_id", dispute.ID).Warn("DisputeResponseManager: failed to update evidence in DB")
	}

	m.logAutomationAction(ctx, paymentDispute.ID, "evidence_submitted", "success", nil)
	m.notifyAdminEvidenceSubmitted(ctx, paymentDispute)

	logrus.WithField("dispute_id", dispute.ID).Info("DisputeResponseManager: evidence submitted successfully")
	return nil
}

type CompiledEvidence struct {
	AccessActivityLog        string
	BillingAddress          string
	RefundPolicyURL         string
	RefundPolicyDisclosed   bool
	RefundPolicyDisclosedText string
	CancellationReason      string
	CustomerEmail           string
	CustomerName           string
	CustomerPurchaseIP     string
	CustomerCommunication   string
	ProductDescription      string
	ReceiptURL             string
	ServiceDate            string
	ServiceDocument        string
	ShippingAddress        string
	ShippingCarrier        string
	ShippingDate           string
	ShippingTracking       string
}

func (m *DisputeResponseManager) compileEvidenceData(ctx context.Context, paymentDispute *storage.PaymentDispute) (*CompiledEvidence, error) {
	evidence := &CompiledEvidence{
		ProductDescription:      "FunctionFly AI Agent Platform Services",
		CustomerEmail:           "",
		CustomerName:            "",
		RefundPolicyURL:        "https://functionfly.com/refund-policy",
		RefundPolicyDisclosed:   true,
		RefundPolicyDisclosedText: "Customers are shown our refund policy at checkout and agree to it before completing their purchase.",
		ServiceDate:             time.Now().Format("2006-01-02"),
	}

	if paymentDispute.TenantID != nil {
		tenant, err := m.getTenantInfo(ctx, *paymentDispute.TenantID)
		if err == nil && tenant != nil {
			evidence.CustomerName = tenant.Name
			evidence.BillingAddress = tenant.BillingAddress
		}
	}

	if paymentDispute.UserID != nil {
		user, err := m.getUserInfo(ctx, *paymentDispute.UserID)
		if err == nil && user != nil {
			evidence.CustomerEmail = user.Email
			if evidence.CustomerName == "" {
				evidence.CustomerName = user.Name
			}
			evidence.CustomerPurchaseIP = user.LastLoginIP
		}
	}

	// Get receipt URL from Stripe charge or local invoice
	evidence.ReceiptURL = m.getReceiptURL(ctx, paymentDispute)

	// Get subscription info for product description
	subInfo := m.getSubscriptionInfo(ctx, paymentDispute)
	evidence.ProductDescription = fmt.Sprintf("FunctionFly AI Platform - %s - %s",
		subInfo.PlanName, paymentDispute.Reason)

	// Get access activity log
	evidence.AccessActivityLog = m.getAccessActivityLog(ctx, paymentDispute)

	// Get support ticket communications
	evidence.CustomerCommunication = m.getSupportCommunications(ctx, paymentDispute)

	// Get service documentation URL
	evidence.ServiceDocument = "https://functionfly.com/terms"

	return evidence, nil
}

func (m *DisputeResponseManager) getTenantInfo(ctx context.Context, tenantID uuid.UUID) (*TenantInfo, error) {
	var tenant TenantInfo
	db := m.disputeRepo.DB().WithContext(ctx).Raw(`
		SELECT name, billing_address FROM tenants WHERE id = ?
	`, tenantID).Scan(&tenant)
	if db.Error != nil {
		return nil, db.Error
	}
	return &tenant, nil
}

func (m *DisputeResponseManager) getUserInfo(ctx context.Context, userID uuid.UUID) (*UserInfo, error) {
	var user UserInfo
	db := m.disputeRepo.DB().WithContext(ctx).Raw(`
		SELECT email, username as name, last_login_ip FROM users WHERE id = ?
	`, userID).Scan(&user)
	if db.Error != nil {
		return nil, db.Error
	}
	return &user, nil
}

func (m *DisputeResponseManager) getAccessActivityLog(ctx context.Context, paymentDispute *storage.PaymentDispute) string {
	if paymentDispute.TenantID == nil {
		return "Activity log not available"
	}

	var logs []struct {
		CreatedAt time.Time
		Action    string
	}

	db := m.disputeRepo.DB().WithContext(ctx).Raw(`
		SELECT created_at, action
		FROM execution_logs
		WHERE tenant_id = ?
		AND created_at >= NOW() - INTERVAL '30 days'
		ORDER BY created_at DESC
		LIMIT 20
	`, *paymentDispute.TenantID).Scan(&logs)

	if db.Error == nil && len(logs) > 0 {
		var lines []string
		for _, log := range logs {
			lines = append(lines, fmt.Sprintf("%s - %s", log.CreatedAt.Format("2006-01-02 15:04"), log.Action))
		}
		if len(lines) > 10 {
			lines = lines[:10]
			lines = append(lines, fmt.Sprintf("... and %d more events in the past 30 days", len(logs)-10))
		}
		return strings.Join(lines, "\n")
	}

	return "Regular API and platform activity over the past 30 days"
}

func (m *DisputeResponseManager) getReceiptURL(ctx context.Context, paymentDispute *storage.PaymentDispute) string {
	if paymentDispute.StripeChargeID == "" {
		return ""
	}

	// Try to get receipt URL from invoices table
	var receiptURL string
	db := m.disputeRepo.DB().WithContext(ctx).Raw(`
		SELECT receipt_url FROM invoices
		WHERE stripe_charge_id = ?
		LIMIT 1
	`, paymentDispute.StripeChargeID).Scan(&receiptURL)

	if db.Error == nil && receiptURL != "" {
		return receiptURL
	}

	// If not found locally, construct Stripe receipt URL
	return fmt.Sprintf("https://dashboard.stripe.com/payments/%s", paymentDispute.StripeChargeID)
}

type SubscriptionInfo struct {
	PlanName   string
	TermAccept string
}

func (m *DisputeResponseManager) getSubscriptionInfo(ctx context.Context, paymentDispute *storage.PaymentDispute) *SubscriptionInfo {
	info := &SubscriptionInfo{
		PlanName:   "Professional Plan",
		TermAccept: time.Now().Format("2006-01-02"),
	}

	if paymentDispute.TenantID == nil {
		return info
	}

	// Get bundle subscription plan name
	var planName string
	db := m.disputeRepo.DB().WithContext(ctx).Raw(`
		SELECT pb.display_name
		FROM bundle_subscriptions bs
		JOIN pricing_bundles pb ON pb.id = bs.bundle_id
		WHERE bs.tenant_id = ?
		AND bs.status = 'active'
		LIMIT 1
	`, *paymentDispute.TenantID).Scan(&planName)

	if db.Error == nil && planName != "" {
		info.PlanName = planName
	}

	// Get terms acceptance date
	var termsAccepted *time.Time
	db = m.disputeRepo.DB().WithContext(ctx).Raw(`
		SELECT terms_accepted_at FROM tenants WHERE id = ?
	`, *paymentDispute.TenantID).Scan(&termsAccepted)

	if db.Error == nil && termsAccepted != nil {
		info.TermAccept = termsAccepted.Format("2006-01-02")
	}

	return info
}

func (m *DisputeResponseManager) getSupportCommunications(ctx context.Context, paymentDispute *storage.PaymentDispute) string {
	if paymentDispute.TenantID == nil {
		return "No support tickets found for this account."
	}

	var tickets []struct {
		CreatedAt time.Time
		Subject   string
		Status    string
	}

	db := m.disputeRepo.DB().WithContext(ctx).Raw(`
		SELECT created_at, subject, status
		FROM support_tickets
		WHERE tenant_id = ?
		AND created_at >= NOW() - INTERVAL '90 days'
		ORDER BY created_at DESC
		LIMIT 5
	`, *paymentDispute.TenantID).Scan(&tickets)

	if db.Error == nil && len(tickets) > 0 {
		var lines []string
		for _, t := range tickets {
			lines = append(lines, fmt.Sprintf("%s - [%s] %s", t.CreatedAt.Format("2006-01-02"), t.Status, t.Subject))
		}
		return strings.Join(lines, "\n")
	}

	return "No recent support tickets. Customer has not contacted support regarding this charge."
}

type TenantInfo struct {
	Name           string
	BillingAddress string
}

type UserInfo struct {
	Email       string
	Name        string
	LastLoginIP string
}

func (m *DisputeResponseManager) logAutomationAction(ctx context.Context, disputeID uuid.UUID, action, outcome string, details map[string]interface{}) {
	detailsJSON, _ := json.Marshal(details)

	db := m.disputeRepo.DB().WithContext(ctx).Exec(`
		INSERT INTO dispute_automation_log (id, dispute_id, action, outcome, details, created_at)
		VALUES (gen_random_uuid(), ?, ?, ?, ?::jsonb, NOW())
	`, disputeID, action, outcome, detailsJSON)
	if db.Error != nil {
		logrus.WithError(db.Error).WithField("dispute_id", disputeID).Warn("DisputeResponseManager: failed to log automation action")
	}
}

func (m *DisputeResponseManager) GetAutomationLog(ctx context.Context, disputeID uuid.UUID) ([]map[string]interface{}, error) {
	var logs []map[string]interface{}
	rows, err := m.disputeRepo.DB().WithContext(ctx).Raw(`
		SELECT action, outcome, details, created_at
		FROM dispute_automation_log
		WHERE dispute_id = ?
		ORDER BY created_at DESC
	`, disputeID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var action, outcome string
		var detailsJSON []byte
		var createdAt time.Time
		rows.Scan(&action, &outcome, &detailsJSON, &createdAt)

		var details map[string]interface{}
		json.Unmarshal(detailsJSON, &details)

		logs = append(logs, map[string]interface{}{
			"action":     action,
			"outcome":    outcome,
			"details":    details,
			"created_at": createdAt,
		})
	}

	return logs, nil
}

func (m *DisputeResponseManager) PreviewEvidence(ctx context.Context, disputeID uuid.UUID) (*storage.EvidenceDetails, error) {
	dispute, err := m.disputeRepo.GetDisputeByID(ctx, disputeID)
	if err != nil || dispute == nil {
		return nil, fmt.Errorf("dispute not found")
	}

	evidence, err := m.compileEvidenceData(ctx, dispute)
	if err != nil {
		return nil, err
	}

	return &storage.EvidenceDetails{
		ProductDescription:  evidence.ProductDescription,
		CustomerEmail:       evidence.CustomerEmail,
		CustomerName:        evidence.CustomerName,
		CustomerPurchaseIP: evidence.CustomerPurchaseIP,
		BillingAddress:     evidence.BillingAddress,
		ReceiptURL:         evidence.ReceiptURL,
		ServiceDate:        evidence.ServiceDate,
		ServiceDocument:    evidence.ServiceDocument,
		ShippingAddress:    evidence.ShippingAddress,
		ShippingDate:       evidence.ShippingDate,
		ShippingTracking:   evidence.ShippingTracking,
		ShippingCarrier:    evidence.ShippingCarrier,
		RefundPolicyURL:   evidence.RefundPolicyURL,
		RefundPolicyDisclosed: evidence.RefundPolicyDisclosed,
		CancellationReason: evidence.CancellationReason,
		AccessActivityLog: evidence.AccessActivityLog,
	}, nil
}

func (m *DisputeResponseManager) SkipAutoRefund(ctx context.Context, disputeID uuid.UUID) error {
	_, err := m.disputeRepo.GetDisputeByID(ctx, disputeID)
	if err != nil {
		return fmt.Errorf("dispute not found")
	}

	m.logAutomationAction(ctx, disputeID, "auto_refund_skipped", "success", map[string]interface{}{
		"skipped_by": "admin",
	})

	logrus.WithField("dispute_id", disputeID).Info("DisputeResponseManager: auto-refund skipped by admin")
	return nil
}

func (m *DisputeResponseManager) ForceRefund(ctx context.Context, disputeID uuid.UUID) error {
	dispute, err := m.disputeRepo.GetDisputeByID(ctx, disputeID)
	if err != nil || dispute == nil {
		return fmt.Errorf("dispute not found")
	}

	evidenceParams := &stripe.DisputeEvidenceParams{
		CancellationRebuttal: stripe.String("Service was provided as described and contracted. Customer agreed to terms at signup."),
	}

	disputeParams := &stripe.DisputeParams{
		Evidence: evidenceParams,
		Submit:  stripe.Bool(true),
	}

	_, err = stripeDispute.Update(dispute.StripeDisputeID, disputeParams)
	if err != nil {
		return fmt.Errorf("failed to update dispute with evidence: %w", err)
	}

	refundParams := &stripe.RefundParams{
		Charge: stripe.String(dispute.StripeChargeID),
		Reason: stripe.String(string(stripe.RefundReasonFraudulent)),
	}

	result, err := refund.New(refundParams)
	if err != nil {
		m.logAutomationAction(ctx, disputeID, "force_refund", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to submit refund: %w", err)
	}

	m.logAutomationAction(ctx, disputeID, "force_refund", "success", map[string]interface{}{
		"refund_id": result.ID,
	})

	logrus.WithFields(logrus.Fields{
		"dispute_id": disputeID,
		"refund_id":  result.ID,
	}).Info("DisputeResponseManager: force refund successful")

	return nil
}

func (m *DisputeResponseManager) GetDisputeByStripeID(ctx context.Context, stripeDisputeID string) (*storage.PaymentDispute, error) {
	return m.disputeRepo.GetDisputeByStripeID(ctx, stripeDisputeID)
}

func (m *DisputeResponseManager) notifyAdminNewDispute(ctx context.Context, dispute *stripe.Dispute, paymentDispute *storage.PaymentDispute) {
	if m.notificationSvc == nil {
		return
	}

	adminUsers := m.getAdminUsers(ctx)
	if len(adminUsers) == 0 {
		return
	}

	amountUSD := float64(dispute.Amount) / 100.0
	evidenceDueBy := "unknown"
	if dispute.EvidenceDetails != nil && dispute.EvidenceDetails.DueBy > 0 {
		evidenceDueBy = time.Unix(dispute.EvidenceDetails.DueBy, 0).Format("2006-01-02")
	}

	m.notificationSvc.SendDisputeCreated(ctx, adminUsers, dispute.ID, fmt.Sprintf("%.2f", amountUSD),
		string(dispute.Currency), string(dispute.Reason), evidenceDueBy)
}

func (m *DisputeResponseManager) notifyAdminAutoRefundExecuted(ctx context.Context, paymentDispute *storage.PaymentDispute, amountUSD float64, reason string) {
	if m.notificationSvc == nil {
		return
	}

	adminUsers := m.getAdminUsers(ctx)
	if len(adminUsers) == 0 {
		return
	}

	m.notificationSvc.SendAutoRefundExecuted(ctx, adminUsers, paymentDispute.StripeDisputeID, amountUSD, reason)
}

func (m *DisputeResponseManager) notifyAdminAutoRefundFailed(ctx context.Context, paymentDispute *storage.PaymentDispute, amountUSD float64, errorMsg string) {
	if m.notificationSvc == nil {
		return
	}

	adminUsers := m.getAdminUsers(ctx)
	if len(adminUsers) == 0 {
		return
	}

	logrus.WithFields(logrus.Fields{
		"dispute_id": paymentDispute.StripeDisputeID,
		"error":      errorMsg,
	}).Warn("DisputeResponseManager: auto-refund failed, manual intervention required")
}

func (m *DisputeResponseManager) notifyAdminEvidenceSubmitted(ctx context.Context, paymentDispute *storage.PaymentDispute) {
	if m.notificationSvc == nil {
		return
	}

	adminUsers := m.getAdminUsers(ctx)
	if len(adminUsers) == 0 {
		return
	}

	m.notificationSvc.SendDisputeEvidenceSubmitted(ctx, adminUsers, paymentDispute.StripeDisputeID)
}

func (m *DisputeResponseManager) notifyAdminDisputeResolved(ctx context.Context, paymentDispute *storage.PaymentDispute, status string, amountUSD float64) {
	if m.notificationSvc == nil {
		return
	}

	adminUsers := m.getAdminUsers(ctx)
	if len(adminUsers) == 0 {
		return
	}

	won := status == "won" || status == "charge_refunded"
	m.notificationSvc.SendDisputeResolved(ctx, adminUsers, paymentDispute.StripeDisputeID, status, amountUSD, won)
}

func (m *DisputeResponseManager) notifyCustomerStatusChange(ctx context.Context, paymentDispute *storage.PaymentDispute, newStatus string) {
	if m.notificationSvc == nil || paymentDispute.UserID == nil {
		return
	}

	user, err := m.getUserInfo(ctx, *paymentDispute.UserID)
	if err != nil || user == nil || user.Email == "" {
		logrus.WithField("user_id", paymentDispute.UserID).Warn("Cannot notify customer: user not found")
		return
	}

	message := getCustomerStatusMessage(newStatus, paymentDispute.Reason)
	_, err = m.notificationSvc.Send(ctx, notification.SendRequest{
		UserID:   *paymentDispute.UserID,
		Type:     notification.TypeBillingAlert,
		Category: notification.CategoryBilling,
		Title:    "Update on Your Disputed Charge",
		Body:     message,
		Data: map[string]interface{}{
			"dispute_id": paymentDispute.StripeDisputeID,
			"status":     newStatus,
			"reason":     paymentDispute.Reason,
		},
		Channels: []string{notification.ChannelEmail, notification.ChannelInApp},
		Priority: notification.PriorityNormal,
	})

	if err != nil {
		logrus.WithError(err).WithField("dispute_id", paymentDispute.StripeDisputeID).Warn("Failed to notify customer of status change")
	} else {
		logrus.WithField("dispute_id", paymentDispute.StripeDisputeID).Info("Customer notified of dispute status change")
	}
}

func (m *DisputeResponseManager) notifyCustomerDisputeResolved(ctx context.Context, paymentDispute *storage.PaymentDispute, won bool, amountUSD float64) {
	if m.notificationSvc == nil || paymentDispute.UserID == nil {
		return
	}

	user, err := m.getUserInfo(ctx, *paymentDispute.UserID)
	if err != nil || user == nil || user.Email == "" {
		logrus.WithField("user_id", paymentDispute.UserID).Warn("Cannot notify customer: user not found")
		return
	}

	var title, body string
	if won {
		title = "Dispute Resolved in Your Favor"
		body = fmt.Sprintf("Good news! The dispute for charge %s has been resolved in your favor. No action is needed from you.", paymentDispute.StripeDisputeID)
	} else {
		title = "Dispute Update"
		body = fmt.Sprintf("The dispute for charge %s has been resolved. A refund of $%.2f has been processed to your original payment method.", paymentDispute.StripeDisputeID, amountUSD)
	}

	_, err = m.notificationSvc.Send(ctx, notification.SendRequest{
		UserID:   *paymentDispute.UserID,
		Type:     notification.TypeBillingAlert,
		Category: notification.CategoryBilling,
		Title:    title,
		Body:     body,
		Data: map[string]interface{}{
			"dispute_id": paymentDispute.StripeDisputeID,
			"won":        won,
			"amount_usd": amountUSD,
		},
		Channels: []string{notification.ChannelEmail, notification.ChannelInApp},
		Priority: notification.PriorityNormal,
	})

	if err != nil {
		logrus.WithError(err).WithField("dispute_id", paymentDispute.StripeDisputeID).Warn("Failed to notify customer of dispute resolution")
	} else {
		logrus.WithField("dispute_id", paymentDispute.StripeDisputeID).Info("Customer notified of dispute resolution")
	}
}

func (m *DisputeResponseManager) getAdminUsers(ctx context.Context) []uuid.UUID {
	var userIDs []uuid.UUID
	rows, err := m.disputeRepo.DB().WithContext(ctx).Raw(`
		SELECT id FROM users WHERE role IN ('admin', 'owner') AND email IS NOT NULL LIMIT 10
	`).Rows()
	if err != nil {
		logrus.WithError(err).Warn("DisputeResponseManager: failed to get admin users")
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		rows.Scan(&id)
		userIDs = append(userIDs, id)
	}

	return userIDs
}

func getCustomerStatusMessage(status, reason string) string {
	switch status {
	case "needs_response", "warning_needs_response":
		return fmt.Sprintf("We received notice of a dispute for your recent charge. Reason: %s. We are reviewing this matter and will take appropriate action.", reason)
	case "needs_review", "under_review":
		return fmt.Sprintf("Your dispute (reason: %s) is under review. We will keep you updated on any developments.", reason)
	case "won", "charge_refunded":
		return "Great news! The dispute has been resolved in your favor. No refund will be issued."
	case "lost":
		return "The dispute has been resolved. A refund has been processed to your original payment method."
	case "closed":
		return "The dispute case has been closed."
	default:
		return fmt.Sprintf("There has been an update to your dispute (reason: %s). Please contact support if you have questions.", reason)
	}
}

func (m *DisputeResponseManager) GetConfigFromDB(ctx context.Context) (*DisputeAutomationConfig, error) {
	var config DisputeAutomationConfigRow
	db := m.disputeRepo.DB().WithContext(ctx).Raw(`
		SELECT auto_refund_enabled, auto_refund_threshold_cents, auto_refund_allowed_reasons,
			   evidence_auto_submit, customer_notification_enabled,
			   admin_escalation_enabled, admin_escalation_threshold_cents
		FROM dispute_automation_config
		LIMIT 1
	`).Scan(&config)
	if db.Error != nil {
		return nil, db.Error
	}

	m.config.AutoRefundEnabled = config.AutoRefundEnabled
	m.config.AutoRefundThresholdCents = config.AutoRefundThresholdCents
	m.config.AutoRefundAllowedReasons = config.AutoRefundAllowedReasons
	m.config.EvidenceAutoSubmit = config.EvidenceAutoSubmit
	m.config.CustomerNotificationEnabled = config.CustomerNotificationEnabled
	m.config.AdminEscalationEnabled = config.AdminEscalationEnabled
	m.config.AdminEscalationThresholdCents = config.AdminEscalationThresholdCents

	return m.config, nil
}

func (m *DisputeResponseManager) UpdateConfigInDB(ctx context.Context, config *DisputeAutomationConfig) error {
	db := m.disputeRepo.DB().WithContext(ctx).Exec(`
		UPDATE dispute_automation_config SET
			auto_refund_enabled = ?,
			auto_refund_threshold_cents = ?,
			auto_refund_allowed_reasons = ?,
			evidence_auto_submit = ?,
			customer_notification_enabled = ?,
			admin_escalation_enabled = ?,
			admin_escalation_threshold_cents = ?,
			updated_at = NOW()
	`, config.AutoRefundEnabled, config.AutoRefundThresholdCents, config.AutoRefundAllowedReasons,
		config.EvidenceAutoSubmit, config.CustomerNotificationEnabled,
		config.AdminEscalationEnabled, config.AdminEscalationThresholdCents)
	return db.Error
}

type DisputeAutomationConfigRow struct {
	AutoRefundEnabled              bool
	AutoRefundThresholdCents       int
	AutoRefundAllowedReasons     []string
	EvidenceAutoSubmit            bool
	CustomerNotificationEnabled    bool
	AdminEscalationEnabled         bool
	AdminEscalationThresholdCents int
}
