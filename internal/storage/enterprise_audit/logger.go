package enterpriseaudit

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type Logger struct {
	repo *Repository
}

func NewLogger(repo *Repository) *Logger {
	return &Logger{repo: repo}
}

type LogInput struct {
	TenantID     uuid.UUID
	ServiceArea  ServiceArea
	Action       string
	ResourceType ResourceType
	ResourceID   *uuid.UUID
	ActorType    ActorType
	ActorID      string
	ActorName    string
	RequestID    string
	IPAddress    string
	UserAgent    string
	Metadata     map[string]interface{}
	Success      bool
	ErrorMessage string
}

func (l *Logger) Log(ctx context.Context, input LogInput) error {
	if input.Metadata == nil {
		input.Metadata = make(map[string]interface{})
	}

	auditLog := &AuditLog{
		TenantID:     input.TenantID,
		ServiceArea:  input.ServiceArea,
		Action:       input.Action,
		ResourceType: input.ResourceType,
		ResourceID:   input.ResourceID,
		ActorType:    input.ActorType,
		ActorID:      input.ActorID,
		ActorName:    input.ActorName,
		RequestID:    input.RequestID,
		IPAddress:    input.IPAddress,
		UserAgent:    input.UserAgent,
		Success:      input.Success,
		ErrorMessage: input.ErrorMessage,
	}
	auditLog.SetMetadata(input.Metadata)

	return l.repo.Create(ctx, auditLog)
}

func (l *Logger) LogAsync(ctx context.Context, input LogInput) {
	go func() {
		if err := l.Log(ctx, input); err != nil {
			// Silent fail - audit logging should not break operations
		}
	}()
}

type AuditRecorder struct {
	TenantID    uuid.UUID
	ActorType   ActorType
	ActorID     string
	ActorName   string
	RequestID   string
	IPAddress   string
	UserAgent   string
	logger      *Logger
	serviceArea ServiceArea
}

func NewAuditRecorder(logger *Logger, serviceArea ServiceArea) func(
	tenantID uuid.UUID,
	actorType ActorType,
	actorID string,
	actorName string,
	requestID string,
	ipAddress string,
	userAgent string,
) *AuditRecorder {
	return func(
		tenantID uuid.UUID,
		actorType ActorType,
		actorID string,
		actorName string,
		requestID string,
		ipAddress string,
		userAgent string,
	) *AuditRecorder {
		return &AuditRecorder{
			TenantID:    tenantID,
			ActorType:  actorType,
			ActorID:    actorID,
			ActorName:  actorName,
			RequestID:  requestID,
			IPAddress:  ipAddress,
			UserAgent:  userAgent,
			logger:     logger,
			serviceArea: serviceArea,
		}
	}
}

func (r *AuditRecorder) Record(ctx context.Context, action string, resourceType ResourceType, resourceID *uuid.UUID, success bool, errMsg string, metadata map[string]interface{}) {
	input := LogInput{
		TenantID:     r.TenantID,
		ServiceArea:  r.serviceArea,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ActorType:    r.ActorType,
		ActorID:      r.ActorID,
		ActorName:    r.ActorName,
		RequestID:    r.RequestID,
		IPAddress:    r.IPAddress,
		UserAgent:    r.UserAgent,
		Metadata:     metadata,
		Success:      success,
		ErrorMessage: errMsg,
	}
	if err := r.logger.Log(ctx, input); err != nil {
		// Silent fail - audit logging should not break operations
	}
}

func (r *AuditRecorder) RecordAsync(ctx context.Context, action string, resourceType ResourceType, resourceID *uuid.UUID, success bool, errMsg string, metadata map[string]interface{}) {
	go func() {
		r.Record(ctx, action, resourceType, resourceID, success, errMsg, metadata)
	}()
}

func CreateAuditLogFromRequest(ctx context.Context, logger *Logger, tenantID uuid.UUID, serviceArea ServiceArea, action string, resourceType ResourceType, resourceID *uuid.UUID, actorType ActorType, actorID, actorName, requestID, ipAddress, userAgent string, success bool, errMsg string, metadata map[string]interface{}) {
	input := LogInput{
		TenantID:     tenantID,
		ServiceArea:  serviceArea,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ActorType:    actorType,
		ActorID:      actorID,
		ActorName:    actorName,
		RequestID:    requestID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Metadata:     metadata,
		Success:      success,
		ErrorMessage: errMsg,
	}
	if err := logger.Log(ctx, input); err != nil {
		// Silent fail
	}
}

type Action string

const (
	ActionCreate Action = "create"
	ActionRead   Action = "read"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionList   Action = "list"
	ActionExport Action = "export"
	ActionLogin  Action = "login"
	ActionLogout Action = "logout"
)

func ActionFromHTTPMethod(method string, path string) string {
	switch method {
	case "GET":
		if path == "" {
			return string(ActionList)
		}
		return string(ActionRead)
	case "POST":
		return string(ActionCreate)
	case "PUT", "PATCH":
		return string(ActionUpdate)
	case "DELETE":
		return string(ActionDelete)
	default:
		return method
	}
}

func GetClientIPFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}

	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

var defaultLogger *Logger

func Log(ctx context.Context, input LogInput) error {
	if defaultLogger == nil {
		return nil
	}
	return defaultLogger.Log(ctx, input)
}

func LogAsync(ctx context.Context, input LogInput) {
	if defaultLogger == nil {
		return
	}
	defaultLogger.LogAsync(ctx, input)
}

func SetLogger(repo *Repository) {
	defaultLogger = NewLogger(repo)
}
