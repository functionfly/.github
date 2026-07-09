package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type IoTDeviceRecord struct {
	ID              uuid.UUID      `json:"id"`
	TenantID        uuid.UUID      `json:"tenant_id"`
	Name            string         `json:"name"`
	DeviceType      string         `json:"device_type"`
	AuthMethod      string         `json:"auth_method"`
	Status          string         `json:"status"`
	CertFingerprint string         `json:"cert_fingerprint,omitempty"`
	PSKHash         string         `json:"psk_hash,omitempty"`
	LastSeen        *time.Time     `json:"last_seen,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type IoTDeviceStateRecord struct {
	DeviceID      uuid.UUID      `json:"device_id"`
	State         map[string]any `json:"state"`
	LastTelemetry map[string]any `json:"last_telemetry,omitempty"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type IoTCommandRecord struct {
	ID             uuid.UUID      `json:"id"`
	DeviceID       uuid.UUID      `json:"device_id"`
	Command        map[string]any `json:"command"`
	Status         string         `json:"status"`
	CreatedAt      time.Time      `json:"created_at"`
	AcknowledgedAt *time.Time     `json:"acknowledged_at,omitempty"`
}

var (
	ErrIoTDeviceNotFound = errors.New("iot device not found")
)

type IoTDeviceRepository struct {
	db *PostgresDB
}

func NewIoTDeviceRepository(db *PostgresDB) *IoTDeviceRepository {
	return &IoTDeviceRepository{db: db}
}

func (r *IoTDeviceRepository) Create(ctx context.Context, device *IoTDeviceRecord) error {
	if device.ID == uuid.Nil {
		device.ID = uuid.New()
	}
	device.CreatedAt = time.Now()
	device.UpdatedAt = time.Now()
	if device.Status == "" {
		device.Status = "offline"
	}

	metadata := []byte("{}")
	if device.Metadata != nil {
		b, err := json.Marshal(device.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadata = b
	}

	query := `
		INSERT INTO iot_devices (id, tenant_id, name, device_type, auth_method, status, cert_fingerprint, psk_hash, last_seen, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, ''), $9, $10, $11, $12)`

	_, err := r.db.ExecContext(ctx, query,
		device.ID, device.TenantID, device.Name, device.DeviceType, device.AuthMethod,
		device.Status, device.CertFingerprint, device.PSKHash, device.LastSeen,
		metadata, device.CreatedAt, device.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}

	state := &IoTDeviceStateRecord{
		DeviceID:  device.ID,
		State:     map[string]any{},
		UpdatedAt: time.Now(),
	}
	if err := r.UpsertState(ctx, state); err != nil {
		return fmt.Errorf("failed to create initial state: %w", err)
	}

	return nil
}

func (r *IoTDeviceRepository) Get(ctx context.Context, deviceID uuid.UUID) (*IoTDeviceRecord, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, device_type, auth_method, status,
			COALESCE(cert_fingerprint, ''), COALESCE(psk_hash, ''),
			last_seen, COALESCE(metadata, '{}'::jsonb), created_at, updated_at
		FROM iot_devices
		WHERE id = $1 AND deleted_at IS NULL`, deviceID)

	return scanDevice(row)
}

func (r *IoTDeviceRepository) GetByCertFingerprint(ctx context.Context, fingerprint string) (*IoTDeviceRecord, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, device_type, auth_method, status,
			COALESCE(cert_fingerprint, ''), COALESCE(psk_hash, ''),
			last_seen, COALESCE(metadata, '{}'::jsonb), created_at, updated_at
		FROM iot_devices
		WHERE cert_fingerprint = $1 AND deleted_at IS NULL`, fingerprint)
	return scanDevice(row)
}

func (r *IoTDeviceRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, deviceType, status string, limit, offset int) ([]*IoTDeviceRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, tenant_id, name, device_type, auth_method, status,
			COALESCE(cert_fingerprint, ''), COALESCE(psk_hash, ''),
			last_seen, COALESCE(metadata, '{}'::jsonb), created_at, updated_at
		FROM iot_devices
		WHERE tenant_id = $1 AND deleted_at IS NULL`
	args := []any{tenantID}
	argIdx := 2

	if deviceType != "" {
		query += fmt.Sprintf(" AND device_type = $%d", argIdx)
		args = append(args, deviceType)
		argIdx++
	}
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}
	defer rows.Close()

	var devices []*IoTDeviceRecord
	for rows.Next() {
		device, err := scanDevice(rows)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}
	return devices, rows.Err()
}

func (r *IoTDeviceRepository) UpdateStatus(ctx context.Context, deviceID uuid.UUID, status string) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		UPDATE iot_devices
		SET status = $1, updated_at = $2,
		    last_seen = CASE WHEN $1 = 'online' THEN $2 ELSE last_seen END
		WHERE id = $3 AND deleted_at IS NULL`,
		status, now, deviceID)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrIoTDeviceNotFound
	}
	return nil
}

func (r *IoTDeviceRepository) UpdateMetadata(ctx context.Context, deviceID uuid.UUID, metadata map[string]any) error {
	meta, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE iot_devices SET metadata = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL`,
		meta, time.Now(), deviceID)
	return err
}

func (r *IoTDeviceRepository) SoftDelete(ctx context.Context, deviceID uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE iot_devices SET deleted_at = $1, updated_at = $1
		WHERE id = $2 AND deleted_at IS NULL`,
		time.Now(), deviceID)
	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrIoTDeviceNotFound
	}
	return nil
}

func (r *IoTDeviceRepository) UpsertState(ctx context.Context, state *IoTDeviceStateRecord) error {
	stateJSON, err := json.Marshal(state.State)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	var telemetryJSON []byte
	if state.LastTelemetry != nil {
		telemetryJSON, err = json.Marshal(state.LastTelemetry)
		if err != nil {
			return fmt.Errorf("failed to marshal telemetry: %w", err)
		}
	}

	state.UpdatedAt = time.Now()
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO iot_device_state (device_id, state, last_telemetry, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (device_id) DO UPDATE
		SET state = EXCLUDED.state,
		    last_telemetry = COALESCE(EXCLUDED.last_telemetry, iot_device_state.last_telemetry),
		    updated_at = EXCLUDED.updated_at`,
		state.DeviceID, stateJSON, telemetryJSON, state.UpdatedAt)
	return err
}

func (r *IoTDeviceRepository) GetState(ctx context.Context, deviceID uuid.UUID) (*IoTDeviceStateRecord, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT device_id, state, last_telemetry, updated_at
		FROM iot_device_state WHERE device_id = $1`, deviceID)

	var state IoTDeviceStateRecord
	var stateJSON, telemetryJSON []byte
	if err := row.Scan(&state.DeviceID, &stateJSON, &telemetryJSON, &state.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrIoTDeviceNotFound
		}
		return nil, err
	}
	if err := json.Unmarshal(stateJSON, &state.State); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}
	if len(telemetryJSON) > 0 {
		_ = json.Unmarshal(telemetryJSON, &state.LastTelemetry)
	}
	return &state, nil
}

func (r *IoTDeviceRepository) CreateCommand(ctx context.Context, cmd *IoTCommandRecord) error {
	if cmd.ID == uuid.Nil {
		cmd.ID = uuid.New()
	}
	cmd.CreatedAt = time.Now()
	if cmd.Status == "" {
		cmd.Status = "pending"
	}

	cmdJSON, err := json.Marshal(cmd.Command)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO iot_commands (id, device_id, command, status, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		cmd.ID, cmd.DeviceID, cmdJSON, cmd.Status, cmd.CreatedAt)
	return err
}

func (r *IoTDeviceRepository) AcknowledgeCommand(ctx context.Context, commandID uuid.UUID) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE iot_commands
		SET status = 'acknowledged', acknowledged_at = $1
		WHERE id = $2 AND status IN ('pending', 'sent')`,
		now, commandID)
	return err
}

func (r *IoTDeviceRepository) ListPendingCommands(ctx context.Context, deviceID uuid.UUID, limit int) ([]*IoTCommandRecord, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, device_id, command, status, created_at, acknowledged_at
		FROM iot_commands
		WHERE device_id = $1 AND status = 'pending'
		ORDER BY created_at ASC LIMIT $2`, deviceID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list commands: %w", err)
	}
	defer rows.Close()

	var cmds []*IoTCommandRecord
	for rows.Next() {
		var c IoTCommandRecord
		var cmdJSON []byte
		if err := rows.Scan(&c.ID, &c.DeviceID, &cmdJSON, &c.Status, &c.CreatedAt, &c.AcknowledgedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(cmdJSON, &c.Command); err != nil {
			return nil, err
		}
		cmds = append(cmds, &c)
	}
	return cmds, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDevice(s rowScanner) (*IoTDeviceRecord, error) {
	var d IoTDeviceRecord
	var metadata []byte
	if err := s.Scan(
		&d.ID, &d.TenantID, &d.Name, &d.DeviceType, &d.AuthMethod, &d.Status,
		&d.CertFingerprint, &d.PSKHash, &d.LastSeen, &metadata, &d.CreatedAt, &d.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrIoTDeviceNotFound
		}
		return nil, err
	}
	if len(metadata) > 0 {
		_ = json.Unmarshal(metadata, &d.Metadata)
	}
	return &d, nil
}
