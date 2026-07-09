-- IoT Device Registry Tables
-- Migration: 20260709141532_iot_devices.up.sql

-- Device registry
CREATE TABLE IF NOT EXISTS iot_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(256) NOT NULL,
    device_type VARCHAR(64) NOT NULL,
    auth_method VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'offline',
    cert_fingerprint TEXT,
    psk_hash TEXT,
    last_seen TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_iot_devices_tenant ON iot_devices(tenant_id);
CREATE INDEX IF NOT EXISTS idx_iot_devices_type ON iot_devices(device_type);
CREATE INDEX IF NOT EXISTS idx_iot_devices_status ON iot_devices(status);
CREATE INDEX IF NOT EXISTS idx_iot_devices_last_seen ON iot_devices(last_seen);
CREATE INDEX IF NOT EXISTS idx_iot_devices_cert_fingerprint ON iot_devices(cert_fingerprint);

-- Device state for current state queries
CREATE TABLE IF NOT EXISTS iot_device_states (
    device_id UUID PRIMARY KEY REFERENCES iot_devices(id) ON DELETE CASCADE,
    state JSONB NOT NULL DEFAULT '{}',
    last_telemetry JSONB,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_iot_device_states_updated ON iot_device_states(updated_at);

-- Device command history for auditing
CREATE TABLE IF NOT EXISTS iot_commands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL REFERENCES iot_devices(id) ON DELETE CASCADE,
    command JSONB NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    acknowledged_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_iot_commands_device ON iot_commands(device_id, created_at);
CREATE INDEX IF NOT EXISTS idx_iot_commands_status ON iot_commands(status);
CREATE INDEX IF NOT EXISTS idx_iot_commands_pending ON iot_commands(device_id, status) WHERE status = 'pending';
