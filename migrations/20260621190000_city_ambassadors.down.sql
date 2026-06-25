-- Migration: 20260621190000_city_ambassadors.down.sql

DROP INDEX IF EXISTS idx_city_ambassadors_user;
DROP INDEX IF EXISTS uq_city_ambassadors_active;
DROP TABLE IF EXISTS city_ambassadors;
