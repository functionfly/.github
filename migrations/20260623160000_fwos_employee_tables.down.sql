-- Migration: 20260623160000_fwos_employee_tables.down.sql
-- Description: Drop all FWOS employee tables in reverse dependency order.

DROP TABLE IF EXISTS fwos_notifications;
DROP TABLE IF EXISTS compensation_access_log;
DROP TABLE IF EXISTS equity_grants;
DROP TABLE IF EXISTS compensation_records;
DROP TABLE IF EXISTS knowledge_articles;
DROP TABLE IF EXISTS employee_learning;
DROP TABLE IF EXISTS learning_courses;
DROP TABLE IF EXISTS task_comments;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS employee_achievements;
DROP TABLE IF EXISTS employee_certifications;
DROP TABLE IF EXISTS employee_skills;
DROP TABLE IF EXISTS employee_departments;
ALTER TABLE departments DROP CONSTRAINT IF EXISTS departments_head_id_fkey;
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS departments;
