package helpers

import (
	"context"
	"net/http"
	"strings"

	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type PCIAuditHelper struct {
	repo *storage.PCIAuditRepository
}

func NewPCIAuditHelper(repo *storage.PCIAuditRepository) *PCIAuditHelper {
	return &PCIAuditHelper{repo: repo}
}

type ActorContext struct {
	UserID    *uuid.UUID
	Email     string
	Role      string
	IP        string
	UserAgent string
	SessionID string
	RequestID string
	TenantID  *uuid.UUID
}

func (h *PCIAuditHelper) ActorContextFromRequest(r *http.Request, tenantID *uuid.UUID) ActorContext {
	claims := middleware.GetUserFromContext(r)
	ip := extractIP(r)
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		sessionID = r.Header.Get("X-Webbot-Session")
	}

	ctx := ActorContext{
		IP:        ip,
		UserAgent: r.Header.Get("User-Agent"),
		SessionID: sessionID,
		RequestID: middleware.GetRequestID(r.Context()),
		TenantID:  tenantID,
	}

	if claims != nil {
		ctx.UserID = &claims.UserID
		ctx.Email = claims.Email
		ctx.Role = string(claims.Role)
		if tenantID == nil && claims.TenantID != uuid.Nil {
			ctx.TenantID = &claims.TenantID
		}
	}

	return ctx
}

func extractIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[0])
	}
	if fwd := r.Header.Get("X-Real-IP"); fwd != "" {
		return fwd
	}
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

func (h *PCIAuditHelper) LogCardDataAccessAsync(ctx context.Context, actor ActorContext, params CardDataAccessParams) {
	if h.repo == nil {
		return
	}
	go func() {
		event, err := h.repo.LogCardholderDataAccess(ctx, params.ToCardholderDataAccessParams(actor))
		if err != nil {
			logrus.WithError(err).Warn("pci_audit: failed to log card data access")
		} else {
			logrus.WithFields(logrus.Fields{
				"event_id":   event.ID,
				"event_type": event.EventType,
				"severity":   event.Severity,
			}).Debug("pci_audit: card data access logged")
		}
	}()
}

func (h *PCIAuditHelper) LogPaymentFlowAsync(ctx context.Context, actor ActorContext, params PaymentFlowParams) {
	if h.repo == nil {
		return
	}
	go func() {
		event, err := h.repo.LogPaymentFlowEvent(ctx, params.ToPaymentFlowEventParams(actor))
		if err != nil {
			logrus.WithError(err).Warn("pci_audit: failed to log payment flow")
		} else {
			logrus.WithFields(logrus.Fields{
				"event_id":   event.ID,
				"event_type": event.EventType,
				"severity":   event.Severity,
			}).Debug("pci_audit: payment flow logged")
		}
	}()
}

func (h *PCIAuditHelper) LogAdminActionAsync(ctx context.Context, actor ActorContext, params AdminActionParams) {
	if h.repo == nil {
		return
	}
	go func() {
		event, err := h.repo.LogAdminAction(ctx, params.ToAdminActionParams(actor))
		if err != nil {
			logrus.WithError(err).Warn("pci_audit: failed to log admin action")
		} else {
			logrus.WithFields(logrus.Fields{
				"event_id":   event.ID,
				"event_type": event.EventType,
				"severity":   event.Severity,
			}).Debug("pci_audit: admin action logged")
		}
	}()
}

type CardDataAccessParams struct {
	AccessType      string
	DataType        string
	PaymentMethodID *uuid.UUID
	CardLastFour    *string
	CardBrand       *string
	CardExpiryMonth *int
	CardExpiryYear  *int
	TokenID         *string
	TransactionID   *string
	Purpose         string
	CDESection      string
	DataFlowStep    string
	Success         bool
	FailureReason   *string
	MFAUsed         bool
}

func (p CardDataAccessParams) ToCardholderDataAccessParams(actor ActorContext) storage.CardholderDataAccessParams {
	return storage.CardholderDataAccessParams{
		UserID:          actor.UserID,
		UserEmail:       actor.Email,
		UserRole:        actor.Role,
		IPAddress:       actor.IP,
		UserAgent:       actor.UserAgent,
		SessionID:       actor.SessionID,
		TenantID:        actor.TenantID,
		PaymentMethodID: p.PaymentMethodID,
		CardLastFour:    p.CardLastFour,
		CardBrand:       p.CardBrand,
		CardExpiryMonth: p.CardExpiryMonth,
		CardExpiryYear:  p.CardExpiryYear,
		TokenID:         p.TokenID,
		TransactionID:   p.TransactionID,
		RequestID:       actor.RequestID,
		AuthMethod:      "session",
		MFAUsed:         p.MFAUsed,
		AccessType:      p.AccessType,
		DataType:        p.DataType,
		Purpose:         p.Purpose,
		CDESection:      p.CDESection,
		DataFlowStep:    p.DataFlowStep,
		Success:         p.Success,
		FailureReason:   p.FailureReason,
	}
}

type PaymentFlowParams struct {
	EventType     string
	CardLastFour  *string
	CardBrand     *string
	CardExpMonth  *int
	CardExpYear   *int
	TokenID       *string
	TransactionID string
	StripeEventID *string
	AmountCents   int
	Currency      string
	PaymentMethod string
	Details       string
	Success       bool
	FailureReason *string
}

func (p PaymentFlowParams) ToPaymentFlowEventParams(actor ActorContext) storage.PaymentFlowEventParams {
	return storage.PaymentFlowEventParams{
		EventType:     p.EventType,
		UserID:        actor.UserID,
		UserEmail:     actor.Email,
		UserRole:      actor.Role,
		IPAddress:     actor.IP,
		TenantID:      actor.TenantID,
		CardLastFour:  p.CardLastFour,
		CardBrand:     p.CardBrand,
		CardExpMonth:  p.CardExpMonth,
		CardExpYear:   p.CardExpYear,
		TokenID:       p.TokenID,
		TransactionID: p.TransactionID,
		StripeEventID: p.StripeEventID,
		RequestID:     actor.RequestID,
		AuthMethod:    "session",
		MFAUsed:       false,
		AmountCents:   p.AmountCents,
		Currency:      p.Currency,
		PaymentMethod: p.PaymentMethod,
		Details:       p.Details,
		Success:       p.Success,
		FailureReason: p.FailureReason,
	}
}

type AdminActionParams struct {
	Action        string
	ResourceType  string
	ResourceID    *uuid.UUID
	Description   string
	CardLastFour  *string
	CardBrand     *string
	Success       bool
	FailureReason *string
}

func (p AdminActionParams) ToAdminActionParams(actor ActorContext) storage.AdminActionParams {
	return storage.AdminActionParams{
		EventType:     p.Action,
		ActorUserID:   actor.UserID,
		ActorEmail:    actor.Email,
		ActorRole:     actor.Role,
		ActorIP:       actor.IP,
		SessionID:     actor.SessionID,
		RequestID:     actor.RequestID,
		TenantID:      actor.TenantID,
		ResourceType:  p.ResourceType,
		ResourceID:    p.ResourceID,
		Description:   p.Description,
		CardLastFour:  p.CardLastFour,
		CardBrand:     p.CardBrand,
		Success:       p.Success,
		FailureReason: p.FailureReason,
	}
}
