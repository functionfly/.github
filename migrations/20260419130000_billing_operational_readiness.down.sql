-- Migration: Drop billing operational readiness tables
-- Created: 2026-04-19

DROP TABLE IF EXISTS eu_vat_validations CASCADE;
DROP TABLE IF EXISTS tax_exemption_certificates CASCADE;
DROP TABLE IF EXISTS webhook_replay_requests CASCADE;
DROP TABLE IF EXISTS stored_webhook_payloads CASCADE;

DROP FUNCTION IF EXISTS cleanup_expired_webhook_payloads();
