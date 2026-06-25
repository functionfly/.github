-- Migration: 20260621234014_company_rankings.down.sql
-- Description: Drop company rankings schema

ALTER TABLE tenants DROP COLUMN IF EXISTS company_id;

DROP TABLE IF EXISTS company_aliases;
DROP TABLE IF EXISTS company_rankings;
DROP TABLE IF EXISTS companies;
