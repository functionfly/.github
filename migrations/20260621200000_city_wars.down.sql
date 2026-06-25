-- Migration: 20260621200000_city_wars.down.sql
-- Description: Drop the city_wars schema.

DROP TABLE IF EXISTS city_war_matches CASCADE;
DROP TABLE IF EXISTS city_wars CASCADE;
