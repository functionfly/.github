-- Drop Certification tables in reverse dependency order
DROP TABLE IF EXISTS cert_grading_queue CASCADE;
DROP TABLE IF EXISTS cert_team_badges CASCADE;
DROP TABLE IF EXISTS cert_subscriptions CASCADE;
DROP TABLE IF EXISTS cert_credentials CASCADE;
DROP TABLE IF EXISTS cert_exams CASCADE;
DROP TABLE IF EXISTS cert_practical_challenges CASCADE;
DROP TABLE IF EXISTS cert_questions CASCADE;
DROP TABLE IF EXISTS cert_tiers CASCADE;

-- Drop credential number sequences
DROP SEQUENCE IF EXISTS cert_credential_seq_architect;
DROP SEQUENCE IF EXISTS cert_credential_seq_professional;
DROP SEQUENCE IF EXISTS cert_credential_seq_associate;
