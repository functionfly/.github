package chat

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatSession struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID     uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index:idx_chat_sessions_tenant_user"`
	UserID       uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index:idx_chat_sessions_tenant_user"`
	Title        string     `json:"title" gorm:"size:255"`
	Model        string     `json:"model" gorm:"size:100;default:'gpt-4o-mini'"`
	ConnectorIDs StringSlice `json:"connector_ids" gorm:"type:jsonb;default:'[]'"`
	Metadata     JSONMap    `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt    time.Time  `json:"created_at" gorm:"default:now()"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"default:now()"`
}

func (ChatSession) TableName() string { return "chat_sessions" }

type ChatMessage struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SessionID    uuid.UUID  `json:"session_id" gorm:"type:uuid;not null;index:idx_chat_messages_session"`
	Role         string     `json:"role" gorm:"size:20;not null"`
	Content      string     `json:"content" gorm:"type:text"`
	Attachments  StringSlice `json:"attachments" gorm:"type:jsonb;default:'[]'"`
	Model        string     `json:"model" gorm:"size:100"`
	TokensUsed   int        `json:"tokens_used" gorm:"default:0"`
	LatencyMS    int        `json:"latency_ms" gorm:"default:0"`
	Metadata     JSONMap    `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt    time.Time  `json:"created_at" gorm:"default:now()"`
}

func (ChatMessage) TableName() string { return "chat_messages" }

type ChatConnector struct {
	ID                 uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID           uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index:idx_chat_connectors_tenant_user"`
	UserID             uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index:idx_chat_connectors_tenant_user"`
	Name               string     `json:"name" gorm:"size:100;not null"`
	Type               string     `json:"type" gorm:"size:50;not null"`
	Icon               string     `json:"icon" gorm:"size:50;default:'plug'"`
	Config             JSONMap    `json:"config" gorm:"type:jsonb;default:'{}'"`
	EncryptedCredentials string   `json:"-" gorm:"column:encrypted_credentials"`
	IsActive           bool       `json:"is_active" gorm:"default:true"`
	LastTestedAt       *time.Time `json:"last_tested_at"`
	LastTestSuccess    *bool      `json:"last_test_success"`
	CreatedAt          time.Time  `json:"created_at" gorm:"default:now()"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"default:now()"`
}

func (ChatConnector) TableName() string { return "chat_connectors" }

type ChatFunction struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID     uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index:idx_chat_functions_tenant_user"`
	UserID       uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index:idx_chat_functions_tenant_user"`
	SessionID    *uuid.UUID `json:"session_id" gorm:"type:uuid"`
	Name         string     `json:"name" gorm:"size:255;not null"`
	Description  string     `json:"description" gorm:"type:text"`
	Code         string     `json:"code" gorm:"type:text;not null"`
	Language     string     `json:"language" gorm:"size:50;default:'typescript'"`
	ConnectorID  *uuid.UUID `json:"connector_id" gorm:"type:uuid"`
	Metadata     JSONMap    `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt    time.Time  `json:"created_at" gorm:"default:now()"`
}

func (ChatFunction) TableName() string { return "chat_functions" }

type ChatBillingAdjustment struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID   uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	SessionID  *uuid.UUID `json:"session_id" gorm:"type:uuid"`
	TokensUsed int       `json:"tokens_used" gorm:"not null"`
	Model      string    `json:"model" gorm:"size:100;not null"`
	ChargedAt  time.Time `json:"charged_at" gorm:"default:now()"`
}

func (ChatBillingAdjustment) TableName() string { return "chat_billing_adjustments" }

type StringSlice []string

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, s)
}

type JSONMap map[string]interface{}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = map[string]interface{}{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, j)
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateSession(ctx context.Context, session *ChatSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *Repository) GetSession(ctx context.Context, id, tenantID uuid.UUID) (*ChatSession, error) {
	var session ChatSession
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&session).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *Repository) ListSessions(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]ChatSession, error) {
	var sessions []ChatSession
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Order("updated_at DESC").
		Limit(limit).Offset(offset).
		Find(&sessions).Error
	return sessions, err
}

func (r *Repository) UpdateSessionTimestamp(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&ChatSession{}).Where("id = ?", id).Update("updated_at", time.Now()).Error
}

func (r *Repository) DeleteSession(ctx context.Context, id, tenantID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).Delete(&ChatSession{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) CreateMessage(ctx context.Context, msg *ChatMessage) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *Repository) ListMessages(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]ChatMessage, error) {
	var messages []ChatMessage
	err := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Limit(limit).Offset(offset).
		Find(&messages).Error
	return messages, err
}

func (r *Repository) CreateConnector(ctx context.Context, conn *ChatConnector) error {
	return r.db.WithContext(ctx).Create(conn).Error
}

func (r *Repository) GetConnector(ctx context.Context, id uuid.UUID) (*ChatConnector, error) {
	var conn ChatConnector
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&conn).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &conn, nil
}

func (r *Repository) ListConnectors(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]ChatConnector, error) {
	var connectors []ChatConnector
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ? AND is_active = ?", tenantID, userID, true).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&connectors).Error
	return connectors, err
}

func (r *Repository) GetConnectorsByIDs(ctx context.Context, tenantID uuid.UUID, ids []string) ([]ChatConnector, error) {
	var uuids []uuid.UUID
	for _, id := range ids {
		if u, err := uuid.Parse(id); err == nil {
			uuids = append(uuids, u)
		}
	}
	if len(uuids) == 0 {
		return nil, nil
	}
	var connectors []ChatConnector
	err := r.db.WithContext(ctx).Where("id IN ? AND tenant_id = ? AND is_active = ?", uuids, tenantID, true).Find(&connectors).Error
	return connectors, err
}

func (r *Repository) DeleteConnector(ctx context.Context, id, tenantID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).Delete(&ChatConnector{}).Error
}

func (r *Repository) CreateFunction(ctx context.Context, fn *ChatFunction) error {
	return r.db.WithContext(ctx).Create(fn).Error
}

func (r *Repository) ListFunctions(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]ChatFunction, error) {
	var funcs []ChatFunction
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&funcs).Error
	return funcs, err
}

func (r *Repository) RecordBilling(ctx context.Context, tenantID, userID, sessionID uuid.UUID, tokens int, model string) error {
	adj := &ChatBillingAdjustment{
		TenantID:   tenantID,
		UserID:     userID,
		SessionID:  &sessionID,
		TokensUsed: tokens,
		Model:      model,
		ChargedAt:  time.Now(),
	}
	return r.db.WithContext(ctx).Create(adj).Error
}
