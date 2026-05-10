-- FunctionFly Developer Certification & Credentialing System
-- Creates 8 tables: cert_tiers, cert_questions, cert_practical_challenges,
-- cert_exams, cert_credentials, cert_subscriptions, cert_team_badges, cert_grading_queue

-- Sequences for credential numbering per tier
CREATE SEQUENCE IF NOT EXISTS cert_credential_seq_associate START 1;
CREATE SEQUENCE IF NOT EXISTS cert_credential_seq_professional START 1;
CREATE SEQUENCE IF NOT EXISTS cert_credential_seq_architect START 1;

-- 1. Certification tiers (Associate, Professional, Architect)
CREATE TABLE IF NOT EXISTS cert_tiers (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug                VARCHAR(50) UNIQUE NOT NULL,
    name                VARCHAR(255) NOT NULL,
    description         TEXT NOT NULL,
    icon                VARCHAR(100),
    color               VARCHAR(20),
    sort_order          INT NOT NULL DEFAULT 0,
    price_cents         INT NOT NULL DEFAULT 0,
    currency            VARCHAR(3) NOT NULL DEFAULT 'USD',
    pass_threshold      NUMERIC(5,2) NOT NULL DEFAULT 70.00,
    time_limit_minutes  INT NOT NULL DEFAULT 90,
    question_count      INT NOT NULL DEFAULT 50,
    practical_count     INT NOT NULL DEFAULT 3,
    validity_months     INT NOT NULL DEFAULT 24,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    metadata            JSONB NOT NULL DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cert_tiers_active ON cert_tiers(is_active, sort_order);

-- 2. Question bank (knowledge questions)
CREATE TABLE IF NOT EXISTS cert_questions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tier_id           UUID NOT NULL REFERENCES cert_tiers(id) ON DELETE CASCADE,
    category          VARCHAR(100) NOT NULL,
    difficulty        VARCHAR(20) NOT NULL DEFAULT 'medium',
    question_text     TEXT NOT NULL,
    question_format   VARCHAR(20) NOT NULL DEFAULT 'multiple_choice',
    options           JSONB NOT NULL,
    correct_answers   JSONB NOT NULL,
    explanation       TEXT,
    points            INT NOT NULL DEFAULT 1,
    is_active         BOOLEAN NOT NULL DEFAULT true,
    created_by        UUID REFERENCES users(id),
    metadata          JSONB NOT NULL DEFAULT '{}',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cert_questions_tier_active ON cert_questions(tier_id, is_active);
CREATE INDEX IF NOT EXISTS idx_cert_questions_category ON cert_questions(tier_id, category);

-- 3. Practical challenges (hands-on tasks)
CREATE TABLE IF NOT EXISTS cert_practical_challenges (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tier_id               UUID NOT NULL REFERENCES cert_tiers(id) ON DELETE CASCADE,
    slug                  VARCHAR(100) UNIQUE NOT NULL,
    name                  VARCHAR(255) NOT NULL,
    description           TEXT NOT NULL,
    category              VARCHAR(100) NOT NULL,
    difficulty            VARCHAR(20) NOT NULL DEFAULT 'medium',
    points                INT NOT NULL DEFAULT 10,
    time_limit_minutes    INT NOT NULL DEFAULT 30,
    grading_config        JSONB NOT NULL,
    validator_function_id UUID,
    is_active             BOOLEAN NOT NULL DEFAULT true,
    metadata              JSONB NOT NULL DEFAULT '{}',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cert_practical_challenges_tier ON cert_practical_challenges(tier_id, is_active);

-- 4. Exam sessions
CREATE TABLE IF NOT EXISTS cert_exams (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tier_id           UUID NOT NULL REFERENCES cert_tiers(id),
    status            VARCHAR(20) NOT NULL DEFAULT 'in_progress',
    stripe_payment_id VARCHAR(255),
    amount_cents      INT NOT NULL DEFAULT 0,
    currency          VARCHAR(3) NOT NULL DEFAULT 'USD',
    started_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    submitted_at      TIMESTAMPTZ,
    graded_at         TIMESTAMPTZ,
    expires_at        TIMESTAMPTZ NOT NULL,
    knowledge_score   NUMERIC(5,2),
    practical_score   NUMERIC(5,2),
    total_score       NUMERIC(5,2),
    passed            BOOLEAN,
    question_ids      JSONB NOT NULL DEFAULT '[]',
    practical_ids     JSONB NOT NULL DEFAULT '[]',
    answers           JSONB NOT NULL DEFAULT '{}',
    practical_results JSONB NOT NULL DEFAULT '{}',
    ip_address        INET,
    user_agent        VARCHAR(500),
    metadata          JSONB NOT NULL DEFAULT '{}',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cert_exams_user_tier ON cert_exams(user_id, tier_id);
CREATE INDEX IF NOT EXISTS idx_cert_exams_status ON cert_exams(status) WHERE status = 'in_progress';
CREATE INDEX IF NOT EXISTS idx_cert_exams_expires ON cert_exams(expires_at) WHERE status = 'in_progress';
CREATE INDEX IF NOT EXISTS idx_cert_exams_created ON cert_exams(created_at DESC);

-- 5. Credentials (earned certifications)
CREATE TABLE IF NOT EXISTS cert_credentials (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tier_id            UUID NOT NULL REFERENCES cert_tiers(id),
    exam_id            UUID NOT NULL REFERENCES cert_exams(id),
    credential_number  VARCHAR(20) UNIQUE NOT NULL,
    status             VARCHAR(20) NOT NULL DEFAULT 'active',
    issued_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at         TIMESTAMPTZ NOT NULL,
    revoked_at         TIMESTAMPTZ,
    revoked_reason     TEXT,
    oba_credential     JSONB,
    verification_hash  VARCHAR(64) NOT NULL,
    verification_url   VARCHAR(500),
    renewal_exam_id    UUID REFERENCES cert_exams(id),
    metadata           JSONB NOT NULL DEFAULT '{}',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cert_credentials_user ON cert_credentials(user_id);
CREATE INDEX IF NOT EXISTS idx_cert_credentials_status ON cert_credentials(user_id, status);
CREATE INDEX IF NOT EXISTS idx_cert_credentials_number ON cert_credentials(credential_number);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cert_credentials_active_per_tier
    ON cert_credentials(user_id, tier_id) WHERE status = 'active';

-- 6. Credential renewal subscriptions
CREATE TABLE IF NOT EXISTS cert_subscriptions (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tier_id                  UUID NOT NULL REFERENCES cert_tiers(id),
    stripe_subscription_id   VARCHAR(255),
    status                   VARCHAR(20) NOT NULL DEFAULT 'active',
    renewal_price_cents      INT NOT NULL DEFAULT 4900,
    currency                 VARCHAR(3) NOT NULL DEFAULT 'USD',
    current_period_start     TIMESTAMPTZ,
    current_period_end       TIMESTAMPTZ,
    canceled_at              TIMESTAMPTZ,
    metadata                 JSONB NOT NULL DEFAULT '{}',
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cert_subscriptions_user ON cert_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_cert_subscriptions_status ON cert_subscriptions(status);

-- 7. Enterprise team certifications
CREATE TABLE IF NOT EXISTS cert_team_badges (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id                UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    tier_id                  UUID NOT NULL REFERENCES cert_tiers(id),
    badge_name               VARCHAR(255) NOT NULL,
    min_certified            INT NOT NULL DEFAULT 5,
    stripe_subscription_id   VARCHAR(255),
    annual_price_cents       INT NOT NULL DEFAULT 49900,
    currency                 VARCHAR(3) NOT NULL DEFAULT 'USD',
    status                   VARCHAR(20) NOT NULL DEFAULT 'active',
    issued_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at               TIMESTAMPTZ NOT NULL,
    verified_count           INT NOT NULL DEFAULT 0,
    metadata                 JSONB NOT NULL DEFAULT '{}',
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cert_team_badges_tenant ON cert_team_badges(tenant_id);

-- 8. Exam grading queue (for async practical challenge grading)
CREATE TABLE IF NOT EXISTS cert_grading_queue (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exam_id         UUID NOT NULL REFERENCES cert_exams(id) ON DELETE CASCADE,
    challenge_id    UUID NOT NULL REFERENCES cert_practical_challenges(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts        INT NOT NULL DEFAULT 0,
    max_attempts    INT NOT NULL DEFAULT 3,
    result          JSONB,
    error_message   TEXT,
    locked_at       TIMESTAMPTZ,
    locked_by       VARCHAR(255),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cert_grading_queue_status ON cert_grading_queue(status)
    WHERE status IN ('pending', 'processing');
CREATE INDEX IF NOT EXISTS idx_cert_grading_queue_exam ON cert_grading_queue(exam_id);

-- Seed the three certification tiers
INSERT INTO cert_tiers (slug, name, description, icon, color, sort_order, price_cents, pass_threshold, time_limit_minutes, question_count, practical_count, validity_months)
VALUES
    ('associate', 'FunctionFly Associate',
     'Demonstrates foundational knowledge of the FunctionFly platform: deploying functions, using the CLI, navigating the marketplace, and understanding core concepts.',
     'Award', 'blue', 1, 5000, 70.00, 60, 40, 2, 24),
    ('professional', 'FunctionFly Professional',
     'Demonstrates proficiency in orchestration, agent management, state fabric, security best practices, and production deployment patterns on FunctionFly.',
     'Shield', 'purple', 2, 15000, 70.00, 90, 50, 3, 24),
    ('architect', 'FunctionFly Architect',
     'Demonstrates expert-level skills in graph design, swarm agent coordination, enterprise architecture patterns, performance optimization, and platform extension.',
     'Crown', 'gold', 3, 20000, 75.00, 120, 60, 4, 24)
ON CONFLICT (slug) DO NOTHING;

-- Seed Associate knowledge questions (batch 1 — core concepts)
INSERT INTO cert_questions (tier_id, category, difficulty, question_text, question_format, options, correct_answers, explanation, points)
SELECT t.id, q.category, q.difficulty, q.question_text, q.question_format,
       q.options::jsonb, q.correct_answers::jsonb, q.explanation, q.points
FROM cert_tiers t
CROSS JOIN (VALUES
     ('deployment', 'easy',
      'What command deploys a function to FunctionFly?',
      'multiple_choice',
      '[{"id":"a","text":"fly deploy"},{"id":"b","text":"ff deploy"},{"id":"c","text":"functionfly push"},{"id":"d","text":"ffly up"}]',
      '["d"]',
      'The CLI command is "ffly up" which builds and deploys your function to the FunctionFly cloud. The CLI was renamed from "ff" to "ffly" to avoid confusion with the fly.io CLI.',
      1),
    ('sandboxing', 'medium',
     'What isolation mechanism does FunctionFly use to execute untrusted function code?',
     'multiple_choice',
     '[{"id":"a","text":"Docker containers"},{"id":"b","text":"WASM sandboxing"},{"id":"c","text":"Virtual machines"},{"id":"d","text":"Process isolation"}]',
     '["b"]',
     'FunctionFly uses WebAssembly (WASM) sandboxing for memory-safe, lightweight isolation of function execution.',
     1),
    ('state', 'medium',
     'When should you use State Fabric instead of the Key-Value store?',
     'multiple_choice',
     '[{"id":"a","text":"For simple key lookups"},{"id":"b","text":"For cross-function state coordination and conflict resolution"},{"id":"c","text":"For caching HTTP responses"},{"id":"d","text":"For storing large binary blobs"}]',
     '["b"]',
     'State Fabric provides cross-function state coordination with CRDT-based conflict resolution, while KV store is for simple key-value lookups.',
     1),
    ('deployment', 'medium',
     'What is the difference between canary and standard deployments?',
     'multiple_choice',
     '[{"id":"a","text":"Canary deploys to all regions at once"},{"id":"b","text":"Canary gradually routes a percentage of traffic to the new version"},{"id":"c","text":"Standard deployments require manual approval"},{"id":"d","text":"There is no difference"}]',
     '["b"]',
     'Canary deployments route a configurable percentage of traffic to the new version, allowing you to catch issues before full rollout.',
     1),
    ('security', 'medium',
     'How does the FunctionFly Secrets Vault protect sensitive data?',
     'multiple_choice',
     '[{"id":"a","text":"Server-side encryption with AES-256"},{"id":"b","text":"Zero-knowledge client-side encryption — the server never sees plaintext"},{"id":"c","text":"Base64 encoding"},{"id":"d","text":"Environment variable obfuscation"}]',
     '["b"]',
     'The Secrets Vault uses zero-knowledge client-side AES-256-GCM encryption. The server stores only ciphertext and never has access to the decryption passphrase.',
     1),
    ('marketplace', 'easy',
     'What is the purpose of the FunctionFly Marketplace?',
     'multiple_choice',
     '[{"id":"a","text":"To buy physical hardware"},{"id":"b","text":"To discover, share, and monetize reusable serverless functions"},{"id":"c","text":"To manage DNS records"},{"id":"d","text":"To deploy Kubernetes clusters"}]',
     '["b"]',
     'The Marketplace is a community-driven platform where developers publish, discover, and monetize reusable serverless functions.',
     1),
    ('cli', 'easy',
     'Which CLI command lists your deployed functions?',
     'multiple_choice',
     '[{"id":"a","text":"ff list"},{"id":"b","text":"ff functions"},{"id":"c","text":"ff ls"},{"id":"d","text":"ff status"}]',
     '["c"]',
     '"ff ls" lists all functions deployed in your current project context.',
     1),
    ('agents', 'medium',
     'What is a FunctionFly Agent?',
     'multiple_choice',
     '[{"id":"a","text":"A monitoring daemon"},{"id":"b","text":"An autonomous entity that can execute functions, manage state, and make decisions"},{"id":"c","text":"A CI/CD pipeline"},{"id":"d","text":"A user account type"}]',
     '["b"]',
     'Agents are autonomous entities on FunctionFly that can execute functions, maintain state, interact with other agents, and make decisions based on configurable policies.',
     1),
    ('security', 'easy',
     'What authentication method does the FunctionFly CLI use?',
     'multiple_choice',
     '[{"id":"a","text":"SSH keys"},{"id":"b","text":"API key stored locally after browser-based OAuth login"},{"id":"c","text":"Username and password"},{"id":"d","text":"Client certificates"}]',
     '["b"]',
     'The CLI authenticates via browser-based OAuth flow, storing a secure API token locally for subsequent requests.',
     1),
    ('state', 'easy',
     'What is the default TTL for Key-Value store entries?',
     'multiple_choice',
     '[{"id":"a","text":"1 hour"},{"id":"b","text":"24 hours"},{"id":"c","text":"No expiry (persistent)"},{"id":"d","text":"7 days"}]',
     '["c"]',
     'KV store entries are persistent by default with no TTL. You can optionally set an expiry time.',
     1),
    ('deployment', 'medium',
     'What happens when you roll back a canary deployment?',
     'multiple_choice',
    '[{"id":"a","text":"The old version is deleted"},{"id":"b","text":"Traffic is routed back to the previous stable version"},{"id":"c","text":"A new deployment is created"},{"id":"d","text":"The function is disabled"}]',
     '["b"]',
     'Rolling back redirects all traffic to the previous stable version. The canary version is kept for debugging.',
     1),
    ('graph', 'medium',
     'What is a Function Graph on FunctionFly?',
     'multiple_choice',
     '[{"id":"a","text":"A monitoring dashboard"},{"id":"b","text":"A DAG of functions connected by data flow edges"},{"id":"c","text":"A database schema"},{"id":"d","text":"A billing chart"}]',
     '["b"]',
     'A Function Graph is a directed acyclic graph (DAG) where nodes are functions and edges define data flow and execution dependencies.',
     1),
    ('observability', 'easy',
     'Where can you view real-time execution logs for a deployed function?',
     'multiple_choice',
     '[{"id":"a","text":"Only in the CLI"},{"id":"b","text":"In the dashboard under the function detail page"},{"id":"c","text":"In a separate monitoring tool"},{"id":"d","text":"Logs are not available"}]',
     '["b"]',
     'The dashboard provides real-time execution logs, metrics, and traces under each function detail page.',
     1),
    ('pricing', 'easy',
     'How does FunctionFly bill for function executions?',
     'multiple_choice',
     '[{"id":"a","text":"Flat monthly fee only"},{"id":"b","text":"Per-execution with compute time and platform fee"},{"id":"c","text":"Per-line of code"},{"id":"d","text":"Per-developer seat"}]',
     '["b"]',
     'FunctionFly uses a pay-per-execution model based on invocation count, compute time (ms × memory), and a platform fee.',
     1),
    ('deployment', 'medium',
     'What is the purpose of deploy keys in FunctionFly?',
     'multiple_choice',
     '[{"id":"a","text":"To encrypt function code"},{"id":"b","text":"To enable CI/CD pipelines to deploy without interactive login"},{"id":"c","text":"To manage DNS"},{"id":"d","text":"To configure environment variables"}]',
     '["b"]',
     'Deploy keys are machine credentials that allow CI/CD systems to deploy functions without requiring interactive browser authentication.',
     1),
    ('security', 'hard',
     'What is the recommended approach for storing database credentials in FunctionFly?',
     'multiple_choice',
     '[{"id":"a","text":"Hard-code them in the function source"},{"id":"b","text":"Store them as environment variables"},{"id":"c","text":"Use the Secrets Vault with client-side encryption"},{"id":"d","text":"Pass them via HTTP headers"}]',
     '["c"]',
     'The Secrets Vault provides zero-knowledge encryption. Environment variables are visible in the dashboard; the Vault ensures even FunctionFly cannot read your secrets.',
     1),
    ('marketplace', 'medium',
     'What revenue share does FunctionFly offer to marketplace publishers?',
     'multiple_choice',
     '[{"id":"a","text":"50/50"},{"id":"b","text":"70/30 (70% to publisher)"},{"id":"c","text":"80/20 (80% to publisher)"},{"id":"d","text":"90/10 (90% to publisher)"}]',
     '["c"]',
     'FunctionFly offers an 80/20 revenue split, with 80% going to the function publisher.',
     1),
    ('agents', 'hard',
     'What is a Swarm Agent on FunctionFly?',
     'multiple_choice',
     '[{"id":"a","text":"A single agent with multiple functions"},{"id":"b","text":"A coordinated group of agents that collaborate on complex tasks"},{"id":"c","text":"A load balancer"},{"id":"d","text":"A monitoring cluster"}]',
     '["b"]',
     'Swarm Agents are coordinated groups of individual agents that communicate and collaborate to solve complex tasks, with built-in consensus and delegation patterns.',
     1),
    ('cli', 'medium',
     'How do you set environment variables for a deployed function?',
     'multiple_choice',
     '[{"id":"a","text":"ff env set KEY=value"},{"id":"b","text":"ff config set KEY=value"},{"id":"c","text":"ff secrets set KEY=value"},{"id":"d","text":"ff vars set KEY=value"}]',
     '["a"]',
     '"ff env set KEY=value" sets an environment variable for the current function context.',
     1),
    ('observability', 'medium',
     'What is Function DNA?',
     'multiple_choice',
     '[{"id":"a","text":"A code generation tool"},{"id":"b","text":"An automated analysis that profiles function behavior, dependencies, and performance characteristics"},{"id":"c","text":"A genetic algorithm"},{"id":"d","text":"A version control system"}]',
     '["b"]',
     'Function DNA is an automated analysis system that creates a behavioral profile of your function including dependency graph, performance characteristics, and risk assessment.',
     1)
) AS q(category, difficulty, question_text, question_format, options, correct_answers, explanation, points)
WHERE t.slug = 'associate'
ON CONFLICT DO NOTHING;
