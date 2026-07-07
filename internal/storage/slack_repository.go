package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SlackConfigRepository struct {
	db *sql.DB
}

func NewSlackConfigRepository(db *sql.DB) *SlackConfigRepository {
	return &SlackConfigRepository{db: db}
}

type SlackConfig struct {
	ID             uuid.UUID       `json:"id"`
	TenantID       uuid.UUID       `json:"tenant_id"`
	BotTokenEnc    []byte          `json:"bot_token_enc"`
	SigningSecret  []byte          `json:"signing_secret"`
	WebhookURL     string          `json:"webhook_url"`
	AlertChannel   string          `json:"alert_channel"`
	ReportChannel  string          `json:"report_channel"`
	ChannelRouting json.RawMessage `json:"channel_routing"`
	SeverityConfig json.RawMessage `json:"severity_config"`
	QuietHours     json.RawMessage `json:"quiet_hours"`
	Enabled        bool            `json:"enabled"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func (r *SlackConfigRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*SlackConfig, error) {
	query := `
		SELECT id, tenant_id, bot_token_enc, signing_secret, webhook_url, 
		       alert_channel, report_channel, channel_routing, severity_config,
		       quiet_hours, enabled, created_at, updated_at
		FROM slack_config
		WHERE tenant_id = $1
	`

	var cfg SlackConfig
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&cfg.ID, &cfg.TenantID, &cfg.BotTokenEnc, &cfg.SigningSecret,
		&cfg.WebhookURL, &cfg.AlertChannel, &cfg.ReportChannel,
		&cfg.ChannelRouting, &cfg.SeverityConfig, &cfg.QuietHours,
		&cfg.Enabled, &cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *SlackConfigRepository) Create(ctx context.Context, cfg *SlackConfig) error {
	query := `
		INSERT INTO slack_config (tenant_id, bot_token_enc, signing_secret, webhook_url,
		                         alert_channel, report_channel, channel_routing,
		                         severity_config, quiet_hours, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRowContext(ctx, query,
		cfg.TenantID, cfg.BotTokenEnc, cfg.SigningSecret, cfg.WebhookURL,
		cfg.AlertChannel, cfg.ReportChannel, cfg.ChannelRouting,
		cfg.SeverityConfig, cfg.QuietHours, cfg.Enabled,
	).Scan(&cfg.ID, &cfg.CreatedAt, &cfg.UpdatedAt)
}

func (r *SlackConfigRepository) Update(ctx context.Context, cfg *SlackConfig) error {
	query := `
		UPDATE slack_config 
		SET bot_token_enc = $2, signing_secret = $3, webhook_url = $4,
		    alert_channel = $5, report_channel = $6, channel_routing = $7,
		    severity_config = $8, quiet_hours = $9, enabled = $10, updated_at = NOW()
		WHERE tenant_id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		cfg.TenantID, cfg.BotTokenEnc, cfg.SigningSecret, cfg.WebhookURL,
		cfg.AlertChannel, cfg.ReportChannel, cfg.ChannelRouting,
		cfg.SeverityConfig, cfg.QuietHours, cfg.Enabled,
	)
	return err
}

type MonitoredComponent struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Enabled     bool      `json:"enabled"`
	SlackChannel string   `json:"slack_channel"`
	CreatedAt   time.Time `json:"created_at"`
}

type SlackAlertLogRepository struct {
	db *sql.DB
}

func NewSlackAlertLogRepository(db *sql.DB) *SlackAlertLogRepository {
	return &SlackAlertLogRepository{db: db}
}

type SlackAlertLog struct {
	ID           uuid.UUID  `json:"id"`
	ComponentID  string     `json:"component_id"`
	OldStatus    string     `json:"old_status"`
	NewStatus    string     `json:"new_status"`
	Severity     string     `json:"severity"`
	Channel      string     `json:"channel"`
	MessageTS    string     `json:"message_ts"`
	Delivered    bool       `json:"delivered"`
	Error        string     `json:"error"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (r *SlackAlertLogRepository) Create(ctx context.Context, log *SlackAlertLog) error {
	query := `
		INSERT INTO slack_alert_log (component_id, old_status, new_status, severity,
		                            channel, message_ts, delivered, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	return r.db.QueryRowContext(ctx, query,
		log.ComponentID, log.OldStatus, log.NewStatus, log.Severity,
		log.Channel, log.MessageTS, log.Delivered, log.Error,
	).Scan(&log.ID, &log.CreatedAt)
}

func (r *SlackAlertLogRepository) UpdateDeliveryStatus(ctx context.Context, id uuid.UUID, delivered bool, messageTS, errorMsg string) error {
	query := `
		UPDATE slack_alert_log 
		SET delivered = $2, message_ts = $3, error = $4
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, delivered, messageTS, errorMsg)
	return err
}

func (r *SlackAlertLogRepository) GetByComponent(ctx context.Context, componentID string, limit int) ([]SlackAlertLog, error) {
	query := `
		SELECT id, component_id, old_status, new_status, severity, channel,
		       message_ts, delivered, error, created_at
		FROM slack_alert_log
		WHERE component_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, componentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []SlackAlertLog
	for rows.Next() {
		var log SlackAlertLog
		var oldStatus, msgTS, errMsg sql.NullString
		if err := rows.Scan(
			&log.ID, &log.ComponentID, &oldStatus, &log.NewStatus, &log.Severity,
			&log.Channel, &msgTS, &log.Delivered, &errMsg, &log.CreatedAt,
		); err != nil {
			return nil, err
		}
		log.OldStatus = oldStatus.String
		log.MessageTS = msgTS.String
		log.Error = errMsg.String
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func GetSlackConfigEnabledTenants(ctx context.Context, db *sql.DB) ([]uuid.UUID, error) {
	query := `SELECT tenant_id FROM slack_config WHERE enabled = true`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenantIDs []uuid.UUID
	for rows.Next() {
		var tenantID uuid.UUID
		if err := rows.Scan(&tenantID); err != nil {
			return nil, err
		}
		tenantIDs = append(tenantIDs, tenantID)
	}
	return tenantIDs, rows.Err()
}
