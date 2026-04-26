-- Migration: Create revenue recognition tables for ASC 606/IFRS 15 compliance

DROP TRIGGER IF EXISTS trigger_po_updated_at ON performance_obligations;
DROP TRIGGER IF EXISTS trigger_ca_updated_at ON contract_assets;
DROP TRIGGER IF EXISTS trigger_rrs_updated_at ON revenue_recognition_schedules;

DROP TABLE IF EXISTS revenue_recognition_events;
DROP TABLE IF EXISTS revenue_recognition_schedules;
DROP TABLE IF EXISTS contract_assets;
DROP TABLE IF EXISTS performance_obligations;

DROP FUNCTION IF EXISTS update_revenue_recognition_updated_at();