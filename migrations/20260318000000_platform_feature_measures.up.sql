-- Platform feature/security measures (admin can toggle enabled)
CREATE TABLE IF NOT EXISTS platform_feature_measures (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key         TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category    TEXT NOT NULL DEFAULT 'General',
    icon        TEXT NOT NULL DEFAULT 'Shield',
    enabled     BOOLEAN NOT NULL DEFAULT true,
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_platform_feature_measures_category ON platform_feature_measures(category);
CREATE INDEX IF NOT EXISTS idx_platform_feature_measures_enabled ON platform_feature_measures(enabled);

-- Seed with current static security measures (all enabled by default)
INSERT INTO platform_feature_measures (key, name, description, category, icon, sort_order) VALUES
('infra-multi-cloud', 'Multi-cloud deployment with automatic failover', 'Infrastructure Security', 'Infrastructure Security', 'Server', 10),
('infra-encryption', 'End-to-end encryption (AES-256)', 'Infrastructure Security', 'Infrastructure Security', 'Server', 20),
('infra-patching', 'Automated security patching and updates', 'Infrastructure Security', 'Infrastructure Security', 'Server', 30),
('infra-ddos', 'DDoS protection with global CDN', 'Infrastructure Security', 'Infrastructure Security', 'Server', 40),
('infra-zero-trust', 'Zero-trust network architecture', 'Infrastructure Security', 'Infrastructure Security', 'Server', 50),
('infra-container-scan', 'Container security scanning', 'Infrastructure Security', 'Infrastructure Security', 'Server', 60),
('app-owasp', 'OWASP Top 10 compliance', 'Application Security', 'Application Security', 'Code', 110),
('app-vuln-scan', 'Automated vulnerability scanning', 'Application Security', 'Application Security', 'Code', 120),
('app-secure-coding', 'Secure coding practices and reviews', 'Application Security', 'Application Security', 'Code', 130),
('app-rasp', 'Runtime Application Self-Protection (RASP)', 'Application Security', 'Application Security', 'Code', 140),
('app-rate-limit', 'API rate limiting and throttling', 'Application Security', 'Application Security', 'Code', 150),
('app-input-validation', 'Input validation and sanitization', 'Application Security', 'Application Security', 'Code', 160),
('data-encryption', 'Data encryption at rest and in transit', 'Data Protection', 'Data Protection', 'Database', 210),
('data-access-controls', 'Database access controls and auditing', 'Data Protection', 'Data Protection', 'Database', 220),
('data-assessments', 'Regular security assessments', 'Data Protection', 'Data Protection', 'Database', 230),
('data-backup', 'Backup encryption and integrity checks', 'Data Protection', 'Data Protection', 'Database', 240),
('data-classification', 'Data classification and handling procedures', 'Data Protection', 'Data Protection', 'Database', 250),
('data-secure-deletion', 'Secure deletion protocols', 'Data Protection', 'Data Protection', 'Database', 260),
('access-mfa', 'Multi-factor authentication (MFA)', 'Access Control', 'Access Control', 'Key', 310),
('access-rbac', 'Role-based access control (RBAC)', 'Access Control', 'Access Control', 'Key', 320),
('access-sso', 'Single sign-on (SSO) integration', 'Access Control', 'Access Control', 'Key', 330),
('access-session', 'Session management and timeout', 'Access Control', 'Access Control', 'Key', 340),
('access-audit-log', 'Audit logging for all access events', 'Access Control', 'Access Control', 'Key', 350),
('access-least-privilege', 'Least privilege principle enforcement', 'Access Control', 'Access Control', 'Key', 360)
ON CONFLICT (key) DO NOTHING;
