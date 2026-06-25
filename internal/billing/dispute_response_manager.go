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
		AutoRefundThresholdCents:        getEnvInt("CHARGEBACK_AUTO_REFUND_THRESHOLD_CENTS", 5000),
		AutoRefundAllowedReasons:        getEnvList("CHARGEBACK_AUTO_REFUND_ALLOWED_REASONS", []string{"duplicate", "product_not_received"}),
		EvidenceAutoSubmit:              getEnvBool("CHARGEBACK_EVIDENCE_AUTO_SUBMIT", true),
		CustomerNotificationEnabled:     getEnvBool("CHARGEBACK_CUSTOMER_NOTIFICATION_ENABLED", true),
		AdminEscalationEnabled:         getEnvBool("CHARGEBACK_ADMIN_ESCALATION_ENABLED", true),
		AdminEscalationThresholdCents:  getEnvInt("CHARGEBACK_ADMIN_ESCALATION_THRESHOLD_CENTS", 50000),
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
		ProductDescription:    "FunctionFly AI Agent Platform Services",
		CustomerEmail:         "customer@example.com",
		CustomerName:          "Customer",
		RefundPolicyURL:      "https://functionfly.com/refund-policy",
		RefundPolicyDisclosed: true,
		RefundPolicyDisclosedText: "Customers are shown our refund policy at checkout and agree to it before completing their purchase.",
		ServiceDate:           time.Now().Format("2006-01-02"),
		CancellationReason:    "Customer initiated cancellation request",
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
			if evidence.CustomerEmail == "customer@example.com" {
				evidence.CustomerEmail = user.Email
			}
			if evidence.CustomerName == "Customer" {
				evidence.CustomerName = user.Name
			}
			evidence.CustomerPurchaseIP = user.LastLoginIP
		}
	}

	evidence.ProductDescription = fmt.Sprintf("FunctionFly AI Platform - %s dispute - Amount: $%.2f",
		paymentDispute.Reason, float64(paymentDispute.AmountCents)/100.0)
	evidence.AccessActivityLog = m.getAccessActivityLog(ctx, paymentDispute)
	evidence.CustomerCommunication = "Customer has contacted support regarding this charge. All support tickets and communications are on file."

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

	var logs []string
	db := m.disputeRepo.DB().WithContext(ctx).Raw(`
		SELECT created_at::text || ' - API activity recorded for tenant'
		FROM execution_logs
		WHERE tenant_id = ?
		AND created_at >= NOW() - INTERVAL '30 days'
		ORDER BY created_at DESC
		LIMIT 50
	`, *paymentDispute.TenantID).Scan(&logs)

	if db.Error == nil && len(logs) > 0 {
		if len(logs) > 10 {
			logs = logs[:10]
			logs = append(logs, "... (additional activity in logs)")
		}
		return strings.Join(logs, "\n")
	}

	return "Regular API and platform activity over the past 30 days"
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

	logrus.WithFields(logrus.Fields{
		"dispute_id": paymentDispute.StripeDisputeID,
		"new_status": newStatus,
		"user_id":    paymentDispute.UserID,
	}).Info("DisputeResponseManager: customer status change notification (not implemented - requires customer email lookup)")
}

func (m *DisputeResponseManager) notifyCustomerDisputeResolved(ctx context.Context, paymentDispute *storage.PaymentDispute, won bool, amountUSD float64) {
	if m.notificationSvc == nil || paymentDispute.UserID == nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"dispute_id": paymentDispute.StripeDisputeID,
		"won":        won,
		"amount_usd": amountUSD,
		"user_id":    paymentDispute.UserID,
	}).Info("DisputeResponseManager: customer dispute resolved notification (not implemented - requires customer email lookup)")
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
