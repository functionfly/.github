-- IoT Device Registry Tables - Rollback
-- Migration: 20260709141532_iot_devices.down.sql

DROP INDEX IF EXISTS idx_iot_commands_pending;
DROP INDEX IF EXISTS idx_iot_commands_status;
DROP INDEX IF EXISTS idx_iot_commands_device;
DROP TABLE IF EXISTS iot_commands;
DROP INDEX IF EXISTS idx_iot_device_states_updated;
DROP TABLE IF EXISTS iot_device_states;
DROP INDEX IF EXISTS idx_iot_devices_cert_fingerprint;
DROP INDEX IF EXISTS idx_iot_devices_last_seen;
DROP INDEX IF EXISTS idx_iot_devices_status;
DROP INDEX IF EXISTS idx_iot_devices_type;
DROP INDEX IF EXISTS idx_iot_devices_tenant;
DROP TABLE IF EXISTS iot_devices;
