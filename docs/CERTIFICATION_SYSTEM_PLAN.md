# Developer Certification & Credentialing System — Production Plan

> **Status:** Architecture Plan  
> **Author:** System Architecture  
> **Date:** 2026-05-03  
> **Scope:** Full-stack certification system with exam engine, credential verification, Open Badges, and billing integration

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architecture Overview](#architecture-overview)
3. [Database Schema](#database-schema)
4. [API Contracts](#api-contracts)
5. [Exam Engine Design](#exam-engine-design)
6. [Credential Verification & Open Badges](#credential-verification--open-badges)
7. [Billing Integration](#billing-integration)
8. [Frontend Architecture](#frontend-architecture)
9. [Background Jobs](#background-jobs)
10. [DRY Reuse Map](#dry-reuse-map)
11. [Rollout Phases](#rollout-phases)
12. [File Manifest](#file-manifest)

---

## Executive Summary

This plan introduces a **FunctionFly Certified Developer** program with three tiers (Associate → Professional → Architect). The system is designed to be **DRY-first**, reusing existing patterns for handlers, storage, billing, achievements, and profiles. It includes:

- **Knowledge questions** (multiple choice) + **practical challenges** (auto-graded platform tasks)
- **Timed exam interface** in the dashboard
- **Verifiable credentials** via Open Badges 3.0 standard
- **Public verification URL** (`functionfly.com/verify/:username`)
- **Stripe billing** for exam attempts ($50–$200 per attempt)
- **Enterprise "Certified Team" badges** (annual subscription)
- **Automatic credential renewal** subscriptions

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Dashboard (React)                        │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────────────┐  │
│  │ Exam UI      │  │ Cert Profile │  │ Public Verify Page    │  │
│  │ (timed)      │  │ (badges)     │  │ (no auth required)    │  │
│  └──────┬───────┘  └──────┬───────┘  └───────────┬───────────┘  │
└─────────┼──────────────────┼──────────────────────┼──────────────┘
          │                  │                      │
          ▼                  ▼                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    API Layer (Go / Gorilla Mux)                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │           internal/api/handlers/certification/          │    │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌───────────┐  │    │
│  │  │ exams.go │ │ creds.go │ │ verify.go│ │ admin.go  │  │    │
│  │  └──────────┘ └──────────┘ └──────────┘ └───────────┘  │    │
│  └─────────────────────────────────────────────────────────┘    │
│  ┌──────────────────────┐  ┌──────────────────────────────┐    │
│  │  middleware.RequireAuth│  │ middleware.RequirePermission │    │
│  └──────────────────────┘  └──────────────────────────────┘    │
└────────────────────────┬───────────────────────────┬────────────┘
                         │                           │
          ┌──────────────▼──────────┐   ┌────────────▼────────────┐
          │  certification_repo.go  │   │  billing handler (reuse)│
          │  (PostgreSQL + GORM)    │   │  (Stripe checkout)      │
          └──────────────┬──────────┘   └────────────┬────────────┘
                         │                           │
          ┌──────────────▼───────────────────────────▼────────────┐
          │                    PostgreSQL                          │
          │  cert_tiers │ cert_questions │ cert_exams │ cert_      │
          │  credentials│ cert_attempts  │ practical_ │ challenges │
          └──────────────────────────────────────────────────────┘
                         │
          ┌──────────────▼────────────────────────────────────────┐
          │              Background Jobs (cron + queue)            │
          │  ┌──────────────┐  ┌──────────────┐  ┌────────────┐  │
          │  │ Exam Grader  │  │ Cert Expiry  │  │ Badge Sync │  │
          │  │ (FOR UPDATE  │  │ (cron daily) │  │ (on cert)  │  │
          │  │  SKIP LOCKED)│  │              │  │            │  │
          │  └──────────────┘  └──────────────┘  └────────────┘  │
          └──────────────────────────────────────────────────────┘
```

---

## Database Schema

### Design Principles
- **UUID PKs** everywhere (matches existing `gen_random_uuid()` pattern)
- **JSONB metadata** for flexible extension (matches `JSONMap` pattern)
- **Idempotent migrations** with `IF NOT EXISTS` (matches project convention)
- **GORM struct tags** + `json` tags (matches `models_core.go`)

### Tables

```sql
-- 1. Certification tiers (Associate, Professional, Architect)
CREATE TABLE IF NOT EXISTS cert_tiers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug            VARCHAR(50) UNIQUE NOT NULL,   -- 'associate', 'professional', 'architect'
    name            VARCHAR(255) NOT NULL,          -- 'FunctionFly Associate'
    description     TEXT NOT NULL,
    icon            VARCHAR(100),                   -- Lucide icon name
    color           VARCHAR(20),                    -- 'blue', 'purple', 'gold'
    sort_order      INT NOT NULL DEFAULT 0,
    price_cents     INT NOT NULL DEFAULT 0,         -- 0 = free (launch promo)
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    pass_threshold  NUMERIC(5,2) NOT NULL DEFAULT 70.00, -- percentage to pass
    time_limit_minutes INT NOT NULL DEFAULT 90,
    question_count  INT NOT NULL DEFAULT 50,        -- questions per exam session
    practical_count INT NOT NULL DEFAULT 3,         -- practical challenges per session
    validity_months INT NOT NULL DEFAULT 24,        -- credential validity period
    is_active       BOOLEAN NOT NULL DEFAULT true,
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 2. Question bank (knowledge questions)
CREATE TABLE IF NOT EXISTS cert_questions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tier_id         UUID NOT NULL REFERENCES cert_tiers(id) ON DELETE CASCADE,
    category        VARCHAR(100) NOT NULL,          -- 'sandboxing', 'state', 'deployment', 'security', etc.
    difficulty      VARCHAR(20) NOT NULL DEFAULT 'medium', -- 'easy', 'medium', 'hard'
    question_text   TEXT NOT NULL,
    question_format VARCHAR(20) NOT NULL DEFAULT 'multiple_choice', -- 'multiple_choice', 'true_false', 'multi_select'
    options         JSONB NOT NULL,                 -- [{id:'a', text:'...'}, ...]
    correct_answers JSONB NOT NULL,                 -- ['a'] or ['a','c'] for multi-select
    explanation     TEXT,                            -- shown after exam for learning
    points          INT NOT NULL DEFAULT 1,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_by      UUID REFERENCES users(id),
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_cert_questions_tier_active ON cert_questions(tier_id, is_active);
CREATE INDEX IF NOT EXISTS idx_cert_questions_category ON cert_questions(tier_id, category);

-- 3. Practical challenges (hands-on tasks)
CREATE TABLE IF NOT EXISTS cert_practical_challenges (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tier_id         UUID NOT NULL REFERENCES cert_tiers(id) ON DELETE CASCADE,
    slug            VARCHAR(100) UNIQUE NOT NULL,    -- 'webhook_validator', 'graph_chain_3', etc.
    name            VARCHAR(255) NOT NULL,
    description     TEXT NOT NULL,                   -- Markdown instructions shown to candidate
    category        VARCHAR(100) NOT NULL,
    difficulty      VARCHAR(20) NOT NULL DEFAULT 'medium',
    points          INT NOT NULL DEFAULT 10,
    time_limit_minutes INT NOT NULL DEFAULT 30,
    -- Grading config (JSONB for flexibility — DRY with execution patterns)
    grading_config  JSONB NOT NULL,                 -- see GradingConfig struct below
    -- Validation function (deployed by FunctionFly for auto-grading)
    validator_function_id UUID,                     -- optional: links to a registry function
    is_active       BOOLEAN NOT NULL DEFAULT true,
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 4. Exam sessions (one per user per attempt)
CREATE TABLE IF NOT EXISTS cert_exams (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tier_id         UUID NOT NULL REFERENCES cert_tiers(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'in_progress',
                    -- 'in_progress', 'submitted', 'grading', 'passed', 'failed', 'expired', 'abandoned'
    -- Payment
    stripe_payment_id VARCHAR(255),
    amount_cents    INT NOT NULL DEFAULT 0,
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    -- Timing
    started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    submitted_at    TIMESTAMPTZ,
    graded_at       TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ NOT NULL,           -- started_at + time_limit
    -- Results
    knowledge_score NUMERIC(5,2),                   -- percentage
    practical_score NUMERIC(5,2),                   -- percentage
    total_score     NUMERIC(5,2),                   -- weighted combination
    passed          BOOLEAN,
    -- Question/practical selection (randomized per session)
    question_ids    JSONB NOT NULL DEFAULT '[]',    -- [uuid, uuid, ...]
    practical_ids   JSONB NOT NULL DEFAULT '[]',    -- [uuid, uuid, ...]
    -- Answers (submitted by candidate)
    answers         JSONB DEFAULT '{}',             -- {question_id: answer_id, ...}
    practical_results JSONB DEFAULT '{}',           -- {challenge_id: {deployed_url, status, score}, ...}
    -- Metadata
    ip_address      INET,
    user_agent      VARCHAR(500),
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_cert_exams_user_tier ON cert_exams(user_id, tier_id);
CREATE INDEX IF NOT EXISTS idx_cert_exams_status ON cert_exams(status) WHERE status = 'in_progress';
CREATE INDEX IF NOT EXISTS idx_cert_exams_expires ON cert_exams(expires_at) WHERE status = 'in_progress';

-- 5. Credentials (earned certifications)
CREATE TABLE IF NOT EXISTS cert_credentials (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tier_id         UUID NOT NULL REFERENCES cert_tiers(id),
    exam_id         UUID NOT NULL REFERENCES cert_exams(id),
    credential_number VARCHAR(20) UNIQUE NOT NULL,  -- 'FFC-2026-000001' (sequential per tier)
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
                    -- 'active', 'expired', 'revoked', 'suspended'
    issued_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,
    revoked_reason  TEXT,
    -- Open Badges 3.0 compliant JSON
    oba_credential  JSONB,                          -- full Open Badges assertion
    -- Verification
    verification_hash VARCHAR(64) NOT NULL,         -- SHA-256 of credential payload
    verification_url VARCHAR(500),                  -- 'https://functionfly.com/verify/username?cert=...'
    -- Renewal
    renewal_exam_id UUID REFERENCES cert_exams(id), -- links to renewal exam if renewed
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_cert_credentials_user ON cert_credentials(user_id);
CREATE INDEX IF NOT EXISTS idx_cert_credentials_status ON cert_credentials(user_id, status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cert_credentials_active_per_tier
    ON cert_credentials(user_id, tier_id) WHERE status = 'active';

-- 6. Credential renewal subscriptions
CREATE TABLE IF NOT EXISTS cert_subscriptions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tier_id             UUID NOT NULL REFERENCES cert_tiers(id),
    stripe_subscription_id VARCHAR(255),
    status              VARCHAR(20) NOT NULL DEFAULT 'active',
                        -- 'active', 'canceled', 'past_due', 'paused'
    renewal_price_cents INT NOT NULL DEFAULT 4900,  -- $49/year
    currency            VARCHAR(3) NOT NULL DEFAULT 'USD',
    current_period_start TIMESTAMPTZ,
    current_period_end  TIMESTAMPTZ,
    canceled_at         TIMESTAMPTZ,
    metadata            JSONB DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_cert_subscriptions_user ON cert_subscriptions(user_id);

-- 7. Enterprise team certifications
CREATE TABLE IF NOT EXISTS cert_team_badges (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    tier_id         UUID NOT NULL REFERENCES cert_tiers(id),
    badge_name      VARCHAR(255) NOT NULL,           -- 'Certified Team - Professional'
    min_certified   INT NOT NULL DEFAULT 5,          -- minimum certified members required
    stripe_subscription_id VARCHAR(255),
    annual_price_cents INT NOT NULL DEFAULT 49900,   -- $499/year
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    issued_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ NOT NULL,
    verified_count  INT NOT NULL DEFAULT 0,          -- current count of certified team members
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_cert_team_badges_tenant ON cert_team_badges(tenant_id);

-- 8. Exam grading queue (for async practical challenge grading)
CREATE TABLE IF NOT EXISTS cert_grading_queue (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exam_id         UUID NOT NULL REFERENCES cert_exams(id) ON DELETE CASCADE,
    challenge_id    UUID NOT NULL REFERENCES cert_practical_challenges(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
                    -- 'pending', 'processing', 'completed', 'failed'
    attempts        INT NOT NULL DEFAULT 0,
    max_attempts    INT NOT NULL DEFAULT 3,
    result          JSONB,                          -- {score, feedback, execution_log}
    error_message   TEXT,
    locked_at       TIMESTAMPTZ,
    locked_by       VARCHAR(255),                   -- worker ID
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_cert_grading_queue_status ON cert_grading_queue(status)
    WHERE status IN ('pending', 'processing');
```

### Go Models (`internal/storage/models_certification.go`)

```go
package storage

import (
    "time"
    "github.com/google/uuid"
)

// CertTier represents a certification level
type CertTier struct {
    ID                uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Slug              string    `json:"slug" gorm:"uniqueIndex;size:50;not null"`
    Name              string    `json:"name" gorm:"size:255;not null"`
    Description       string    `json:"description" gorm:"type:text;not null"`
    Icon              string    `json:"icon" gorm:"size:100"`
    Color             string    `json:"color" gorm:"size:20"`
    SortOrder         int       `json:"sort_order" gorm:"default:0"`
    PriceCents        int       `json:"price_cents" gorm:"default:0"`
    Currency          string    `json:"currency" gorm:"size:3;default:'USD'"`
    PassThreshold     float64   `json:"pass_threshold" gorm:"type:numeric(5,2);default:70"`
    TimeLimitMinutes  int       `json:"time_limit_minutes" gorm:"default:90"`
    QuestionCount     int       `json:"question_count" gorm:"default:50"`
    PracticalCount    int       `json:"practical_count" gorm:"default:3"`
    ValidityMonths    int       `json:"validity_months" gorm:"default:24"`
    IsActive          bool      `json:"is_active" gorm:"default:true"`
    Metadata          JSONMap   `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
    CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CertTier) TableName() string { return "cert_tiers" }

// CertQuestion represents a knowledge question in the question bank
type CertQuestion struct {
    ID             uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    TierID         uuid.UUID `json:"tier_id" gorm:"type:uuid;not null;index"`
    Category       string    `json:"category" gorm:"size:100;not null"`
    Difficulty     string    `json:"difficulty" gorm:"size:20;default:'medium'"`
    QuestionText   string    `json:"question_text" gorm:"type:text;not null"`
    QuestionFormat string    `json:"question_format" gorm:"size:20;default:'multiple_choice'"`
    Options        JSONMap   `json:"options" gorm:"type:jsonb;not null"`
    CorrectAnswers JSONMap   `json:"correct_answers" gorm:"type:jsonb;not null"`
    Explanation    string    `json:"explanation,omitempty" gorm:"type:text"`
    Points         int       `json:"points" gorm:"default:1"`
    IsActive       bool      `json:"is_active" gorm:"default:true"`
    CreatedBy      *uuid.UUID `json:"created_by,omitempty" gorm:"type:uuid"`
    Metadata       JSONMap   `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
    CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CertQuestion) TableName() string { return "cert_questions" }

// CertPracticalChallenge represents a hands-on grading challenge
type CertPracticalChallenge struct {
    ID                uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    TierID            uuid.UUID `json:"tier_id" gorm:"type:uuid;not null;index"`
    Slug              string    `json:"slug" gorm:"uniqueIndex;size:100;not null"`
    Name              string    `json:"name" gorm:"size:255;not null"`
    Description       string    `json:"description" gorm:"type:text;not null"`
    Category          string    `json:"category" gorm:"size:100;not null"`
    Difficulty        string    `json:"difficulty" gorm:"size:20;default:'medium'"`
    Points            int       `json:"points" gorm:"default:10"`
    TimeLimitMinutes  int       `json:"time_limit_minutes" gorm:"default:30"`
    GradingConfig     JSONMap   `json:"grading_config" gorm:"type:jsonb;not null"`
    ValidatorFuncID   *uuid.UUID `json:"validator_function_id,omitempty" gorm:"type:uuid"`
    IsActive          bool      `json:"is_active" gorm:"default:true"`
    Metadata          JSONMap   `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
    CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CertPracticalChallenge) TableName() string { return "cert_practical_challenges" }

// CertExam represents a single exam session
type CertExam struct {
    ID               uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    UserID           uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
    TierID           uuid.UUID  `json:"tier_id" gorm:"type:uuid;not null"`
    Status           string     `json:"status" gorm:"size:20;default:'in_progress';not null"`
    StripePaymentID  *string    `json:"stripe_payment_id,omitempty" gorm:"size:255"`
    AmountCents      int        `json:"amount_cents" gorm:"default:0"`
    Currency         string     `json:"currency" gorm:"size:3;default:'USD'"`
    StartedAt        time.Time  `json:"started_at" gorm:"not null;default:now()"`
    SubmittedAt      *time.Time `json:"submitted_at,omitempty"`
    GradedAt         *time.Time `json:"graded_at,omitempty"`
    ExpiresAt        time.Time  `json:"expires_at" gorm:"not null"`
    KnowledgeScore   *float64   `json:"knowledge_score,omitempty" gorm:"type:numeric(5,2)"`
    PracticalScore   *float64   `json:"practical_score,omitempty" gorm:"type:numeric(5,2)"`
    TotalScore       *float64   `json:"total_score,omitempty" gorm:"type:numeric(5,2)"`
    Passed           *bool      `json:"passed,omitempty"`
    QuestionIDs      JSONMap    `json:"question_ids" gorm:"type:jsonb;default:'[]'"`
    PracticalIDs     JSONMap    `json:"practical_ids" gorm:"type:jsonb;default:'[]'"`
    Answers          JSONMap    `json:"answers,omitempty" gorm:"type:jsonb;default:'{}'"`
    PracticalResults JSONMap    `json:"practical_results,omitempty" gorm:"type:jsonb;default:'{}'"`
    IPAddress        *string    `json:"ip_address,omitempty" gorm:"type:inet"`
    UserAgent        *string    `json:"user_agent,omitempty" gorm:"size:500"`
    Metadata         JSONMap    `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
    CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
    // Relations
    Tier             *CertTier  `json:"tier,omitempty" gorm:"foreignKey:TierID"`
}

func (CertExam) TableName() string { return "cert_exams" }

// CertCredential represents an earned certification
type CertCredential struct {
    ID               uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    UserID           uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
    TierID           uuid.UUID  `json:"tier_id" gorm:"type:uuid;not null"`
    ExamID           uuid.UUID  `json:"exam_id" gorm:"type:uuid;not null"`
    CredentialNumber string     `json:"credential_number" gorm:"uniqueIndex;size:20;not null"`
    Status           string     `json:"status" gorm:"size:20;default:'active';not null"`
    IssuedAt         time.Time  `json:"issued_at" gorm:"not null;default:now()"`
    ExpiresAt        time.Time  `json:"expires_at" gorm:"not null"`
    RevokedAt        *time.Time `json:"revoked_at,omitempty"`
    RevokedReason    *string    `json:"revoked_reason,omitempty" gorm:"type:text"`
    OBACredential    JSONMap    `json:"oba_credential,omitempty" gorm:"type:jsonb"`
    VerificationHash string     `json:"verification_hash" gorm:"size:64;not null"`
    VerificationURL  *string    `json:"verification_url,omitempty" gorm:"size:500"`
    RenewalExamID    *uuid.UUID `json:"renewal_exam_id,omitempty" gorm:"type:uuid"`
    Metadata         JSONMap    `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
    CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
    // Relations
    Tier             *CertTier  `json:"tier,omitempty" gorm:"foreignKey:TierID"`
    User             *User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (CertCredential) TableName() string { return "cert_credentials" }

// CertSubscription represents a credential renewal subscription
type CertSubscription struct {
    ID                   uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    UserID               uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
    TierID               uuid.UUID  `json:"tier_id" gorm:"type:uuid;not null"`
    StripeSubscriptionID *string    `json:"stripe_subscription_id,omitempty" gorm:"size:255"`
    Status               string     `json:"status" gorm:"size:20;default:'active';not null"`
    RenewalPriceCents    int        `json:"renewal_price_cents" gorm:"default:4900"`
    Currency             string     `json:"currency" gorm:"size:3;default:'USD'"`
    CurrentPeriodStart   *time.Time `json:"current_period_start,omitempty"`
    CurrentPeriodEnd     *time.Time `json:"current_period_end,omitempty"`
    CanceledAt           *time.Time `json:"canceled_at,omitempty"`
    Metadata             JSONMap    `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
    CreatedAt            time.Time  `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt            time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CertSubscription) TableName() string { return "cert_subscriptions" }

// CertTeamBadge represents an enterprise team certification
type CertTeamBadge struct {
    ID                   uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    TenantID             uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
    TierID               uuid.UUID  `json:"tier_id" gorm:"type:uuid;not null"`
    BadgeName            string     `json:"badge_name" gorm:"size:255;not null"`
    MinCertified         int        `json:"min_certified" gorm:"default:5"`
    StripeSubscriptionID *string    `json:"stripe_subscription_id,omitempty" gorm:"size:255"`
    AnnualPriceCents     int        `json:"annual_price_cents" gorm:"default:49900"`
    Currency             string     `json:"currency" gorm:"size:3;default:'USD'"`
    Status               string     `json:"status" gorm:"size:20;default:'active';not null"`
    IssuedAt             time.Time  `json:"issued_at" gorm:"not null;default:now()"`
    ExpiresAt            time.Time  `json:"expires_at" gorm:"not null"`
    VerifiedCount        int        `json:"verified_count" gorm:"default:0"`
    Metadata             JSONMap    `json:"metadata,omitempty" gorm:"type:jsonb;default:'{}'"`
    CreatedAt            time.Time  `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt            time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CertTeamBadge) TableName() string { return "cert_team_badges" }

// CertGradingQueueItem represents a practical challenge grading task
type CertGradingQueueItem struct {
    ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    ExamID       uuid.UUID  `json:"exam_id" gorm:"type:uuid;not null;index"`
    ChallengeID  uuid.UUID  `json:"challenge_id" gorm:"type:uuid;not null"`
    Status       string     `json:"status" gorm:"size:20;default:'pending';not null"`
    Attempts     int        `json:"attempts" gorm:"default:0"`
    MaxAttempts  int        `json:"max_attempts" gorm:"default:3"`
    Result       JSONMap    `json:"result,omitempty" gorm:"type:jsonb"`
    ErrorMessage *string    `json:"error_message,omitempty" gorm:"type:text"`
    LockedAt     *time.Time `json:"locked_at,omitempty"`
    LockedBy     *string    `json:"locked_by,omitempty" gorm:"size:255"`
    CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CertGradingQueueItem) TableName() string { return "cert_grading_queue" }

// GradingConfig defines how a practical challenge is auto-graded
// Stored as JSONB in cert_practical_challenges.grading_config
type GradingConfig struct {
    Type            string            `json:"type"`             // 'http_response', 'state_check', 'graph_execution', 'deployment_check'
    Endpoint        string            `json:"endpoint"`         // function URL to call
    Method          string            `json:"method"`           // HTTP method
    Headers         map[string]string `json:"headers"`          // required headers
    Body            string            `json:"body"`             // request body
    ExpectedStatus  int               `json:"expected_status"`  // expected HTTP status
    ExpectedBody    *string           `json:"expected_body"`    // expected response body (substring match)
    ExpectedJSON    map[string]interface{} `json:"expected_json"` // expected JSON structure
    StateChecks     []StateCheck      `json:"state_checks"`     // state verification rules
    TimeoutSeconds  int               `json:"timeout_seconds"`  // max execution time
}

type StateCheck struct {
    Store   string      `json:"store"`   // state store name
    Key     string      `json:"key"`     // key to check
    Value   interface{} `json:"value"`   // expected value
    Operator string     `json:"operator"` // 'eq', 'contains', 'exists', 'gt', 'lt'
}
```

---

## API Contracts

All endpoints follow existing patterns: `gorilla/mux` routing, `middleware.RequireAuth`, JSON responses with `writeJSON`/`writeJSONError`.

### Route Registration (in `routes.go`)

```go
func registerCertificationRoutes(api *mux.Router, authMiddleware *middleware.AuthMiddleware, certHandler *certification.Handler) {
    cert := api.PathPrefix("/certification").Subrouter()
    cert.Use(authMiddleware.OptionalAuth)

    // Public
    cert.HandleFunc("/tiers", certHandler.ListTiers).Methods("GET", "OPTIONS")
    cert.HandleFunc("/verify/{username}", certHandler.VerifyCredential).Methods("GET", "OPTIONS")
    cert.HandleFunc("/verify/{credentialNumber}", certHandler.VerifyByNumber).Methods("GET", "OPTIONS")
    cert.HandleFunc("/credentials/{username}/badges", certHandler.PublicBadges).Methods("GET", "OPTIONS")
    cert.HandleFunc("/openbadges/{username}.json", certHandler.OpenBadgesProfile).Methods("GET", "OPTIONS")

    // Authenticated
    auth := cert.PathPrefix("").Subrouter()
    auth.Use(authMiddleware.RequireAuth)

    // Exams
    auth.HandleFunc("/tiers/{tierSlug}/start", certHandler.StartExam).Methods("POST", "OPTIONS")
    auth.HandleFunc("/exams/{examId}", certHandler.GetExam).Methods("GET", "OPTIONS")
    auth.HandleFunc("/exams/{examId}/answer", certHandler.SubmitAnswer).Methods("PUT", "OPTIONS")
    auth.HandleFunc("/exams/{examId}/submit", certHandler.SubmitExam).Methods("POST", "OPTIONS")
    auth.HandleFunc("/exams/{examId}/practical/{challengeId}/deploy", certHandler.DeployPractical).Methods("POST", "OPTIONS")
    auth.HandleFunc("/exams/{examId}/practical/{challengeId}/submit", certHandler.SubmitPractical).Methods("POST", "OPTIONS")
    auth.HandleFunc("/exams", certHandler.ListMyExams).Methods("GET", "OPTIONS")

    // Credentials
    auth.HandleFunc("/credentials", certHandler.ListMyCredentials).Methods("GET", "OPTIONS")
    auth.HandleFunc("/credentials/{credentialId}", certHandler.GetCredential).Methods("GET", "OPTIONS")
    auth.HandleFunc("/credentials/{credentialId}/renew", certHandler.RenewCredential).Methods("POST", "OPTIONS")

    // Renewal subscriptions
    auth.HandleFunc("/subscriptions", certHandler.ListMySubscriptions).Methods("GET", "OPTIONS")
    auth.HandleFunc("/subscriptions/create", certHandler.CreateSubscription).Methods("POST", "OPTIONS")
    auth.HandleFunc("/subscriptions/{subId}/cancel", certHandler.CancelSubscription).Methods("POST", "OPTIONS")

    // Team badges (enterprise)
    auth.HandleFunc("/team-badge", certHandler.GetTeamBadge).Methods("GET", "OPTIONS")
    auth.HandleFunc("/team-badge/create", certHandler.CreateTeamBadge).Methods("POST", "OPTIONS")
    auth.HandleFunc("/team-badge/members", certHandler.TeamBadgeMembers).Methods("GET", "OPTIONS")

    // Admin
    admin := cert.PathPrefix("/admin").Subrouter()
    admin.Use(authMiddleware.RequireAuth)
    admin.Use(authMiddleware.RequirePermission("certification:admin"))
    admin.HandleFunc("/tiers", certHandler.AdminCreateTier).Methods("POST", "OPTIONS")
    admin.HandleFunc("/tiers/{tierId}", certHandler.AdminUpdateTier).Methods("PUT", "OPTIONS")
    admin.HandleFunc("/questions", certHandler.AdminCreateQuestion).Methods("POST", "OPTIONS")
    admin.HandleFunc("/questions/import", certHandler.AdminImportQuestions).Methods("POST", "OPTIONS")
    admin.HandleFunc("/questions/{questionId}", certHandler.AdminUpdateQuestion).Methods("PUT", "OPTIONS")
    admin.HandleFunc("/questions/{questionId}", certHandler.AdminDeleteQuestion).Methods("DELETE", "OPTIONS")
    admin.HandleFunc("/challenges", certHandler.AdminCreateChallenge).Methods("POST", "OPTIONS")
    admin.HandleFunc("/challenges/{challengeId}", certHandler.AdminUpdateChallenge).Methods("PUT", "OPTIONS")
    admin.HandleFunc("/credentials/{credentialId}/revoke", certHandler.AdminRevokeCredential).Methods("POST", "OPTIONS")
    admin.HandleFunc("/stats", certHandler.AdminStats).Methods("GET", "OPTIONS")
}
```

### Endpoint Specifications

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/v1/certification/tiers` | No | List all active certification tiers |
| `POST` | `/v1/certification/tiers/{tierSlug}/start` | Yes | Start exam session (creates Stripe checkout if paid) |
| `GET` | `/v1/certification/exams/{examId}` | Yes | Get exam session with questions (no answers) |
| `PUT` | `/v1/certification/exams/{examId}/answer` | Yes | Submit/update answer for a question |
| `POST` | `/v1/certification/exams/{examId}/submit` | Yes | Submit entire exam for grading |
| `POST` | `/v1/certification/exams/{examId}/practical/{challengeId}/deploy` | Yes | Deploy practical challenge environment |
| `POST` | `/v1/certification/exams/{examId}/practical/{challengeId}/submit` | Yes | Submit practical challenge solution |
| `GET` | `/v1/certification/exams` | Yes | List user's exam history |
| `GET` | `/v1/certification/credentials` | Yes | List user's active credentials |
| `GET` | `/v1/certification/credentials/{credentialId}` | Yes | Get credential detail + Open Badges JSON |
| `POST` | `/v1/certification/credentials/{credentialId}/renew` | Yes | Start renewal process |
| `GET` | `/v1/certification/verify/{username}` | No | Public verification page data |
| `GET` | `/v1/certification/verify/{credentialNumber}` | No | Verify by credential number |
| `GET` | `/v1/certification/credentials/{username}/badges` | No | Public badge list for profile |
| `GET` | `/v1/certification/openbadges/{username}.json` | No | Open Badges 3.0 compliant JSON |
| `POST` | `/v1/certification/subscriptions/create` | Yes | Create renewal subscription |
| `POST` | `/v1/certification/team-badge/create` | Yes | Create enterprise team badge |

### Request/Response Examples

**Start Exam:**
```json
POST /v1/certification/tiers/associate/start
{
  "payment_method_id": "pm_123"  // optional if free
}

Response 201:
{
  "exam": {
    "id": "uuid",
    "tier": { "slug": "associate", "name": "FunctionFly Associate" },
    "status": "in_progress",
    "started_at": "2026-05-03T00:00:00Z",
    "expires_at": "2026-05-03T01:30:00Z",
    "time_limit_minutes": 90,
    "question_count": 50,
    "practical_count": 3,
    "total_points": 100
  },
  "checkout_url": "https://checkout.stripe.com/..." // if paid
}
```

**Get Exam (with randomized questions, no correct answers):**
```json
GET /v1/certification/exams/{examId}

Response 200:
{
  "exam": {
    "id": "uuid",
    "status": "in_progress",
    "time_remaining_seconds": 4320,
    "questions": [
      {
        "id": "uuid",
        "category": "sandboxing",
        "difficulty": "medium",
        "question_text": "What isolation boundary does WASM sandboxing provide?",
        "question_format": "multiple_choice",
        "options": [
          {"id": "a", "text": "Process-level isolation"},
          {"id": "b", "text": "Memory-safe module isolation"},
          {"id": "c", "text": "Container-level isolation"},
          {"id": "d", "text": "VM-level isolation"}
        ],
        "points": 1
      }
    ],
    "practical_challenges": [
      {
        "id": "uuid",
        "name": "Webhook Validator",
        "description": "Deploy a function that...",
        "category": "deployment",
        "difficulty": "medium",
        "points": 10,
        "time_limit_minutes": 30
      }
    ]
  }
}
```

**Submit Answer:**
```json
PUT /v1/certification/exams/{examId}/answer
{
  "question_id": "uuid",
  "answer": "b"
}

Response 200:
{
  "saved": true,
  "answers_submitted": 12,
  "answers_total": 50
}
```

**Public Verification:**
```json
GET /v1/certification/verify/johndoe

Response 200:
{
  "user": {
    "username": "johndoe",
    "name": "John Doe",
    "avatar_url": "..."
  },
  "credentials": [
    {
      "credential_number": "FFC-2026-000142",
      "tier": { "slug": "professional", "name": "FunctionFly Professional" },
      "issued_at": "2026-03-15T00:00:00Z",
      "expires_at": "2028-03-15T00:00:00Z",
      "status": "active",
      "verification_hash": "abc123..."
    }
  ],
  "verified_at": "2026-05-03T00:25:00Z"
}
```

---

## Exam Engine Design

### Question Selection Algorithm

```
For each exam session:
1. Fetch all active questions for the requested tier
2. Stratify by category × difficulty matrix:
   - 60% medium, 25% hard, 15% easy (weighted random)
   - Ensure minimum coverage per category (at least 2 per category)
3. Fisher-Yates shuffle the selected pool
4. Take first N questions (configurable per tier)
5. Store selected question_ids in cert_exams.question_ids
6. Return questions WITHOUT correct_answers/explanation fields
```

### Practical Challenge Auto-Grading

The grading system uses the existing platform execution infrastructure:

1. **Candidate deploys** a function during the exam (via existing function deployment API)
2. **Candidate submits** the function URL as their answer
3. **Grading queue** picks up the task via `SELECT ... FOR UPDATE SKIP LOCKED`
4. **Grader calls** the deployed function with test inputs from `GradingConfig`
5. **Grader verifies** response status, body, JSON structure, and state checks
6. **Score recorded** in `cert_grading_queue.result`

```go
// GradingConfig stored in cert_practical_challenges.grading_config
type GradingConfig struct {
    Type           string            `json:"type"`
    Endpoint       string            `json:"endpoint"`
    Method         string            `json:"method"`
    Headers        map[string]string `json:"headers"`
    Body           string            `json:"body"`
    ExpectedStatus int               `json:"expected_status"`
    ExpectedBody   *string           `json:"expected_body"`
    ExpectedJSON   map[string]interface{} `json:"expected_json"`
    StateChecks    []StateCheck      `json:"state_checks"`
    TimeoutSeconds int               `json:"timeout_seconds"`
}
```

### Exam Lifecycle State Machine

```
[User clicks "Start Exam"]
         │
         ▼
    ┌──────────┐   timeout    ┌──────────┐
    │ IN_PROGRESS├───────────►│ EXPIRED  │
    └────┬─────┘              └──────────┘
         │ submit
         ▼
    ┌──────────┐   practical  ┌──────────┐
    │ SUBMITTED├─────────────►│ GRADING  │
    └────┬────┘               └────┬─────┘
         │ all graded              │
         ▼                         ▼
    ┌──────────┐              ┌──────────┐
    │  PASSED  │              │  FAILED  │
    └────┬─────┘              └──────────┘
         │
         ▼
    ┌──────────┐
    │CREDENTIAL│
    │ ISSUED   │
    └──────────┘
```

### Scoring Formula

```
knowledge_score = (correct_answers / total_questions) * 100
practical_score = (sum_of_practical_points_earned / sum_of_practical_points_possible) * 100
total_score     = (knowledge_score * 0.6) + (practical_score * 0.4)
passed          = total_score >= tier.pass_threshold
```

---

## Credential Verification & Open Badges

### Open Badges 3.0 Integration

FunctionFly credentials comply with [Open Badges 3.0](https://www.imsglobal.org/spec/ob/v3p0/) for LinkedIn/native badge platform compatibility.

```json
// OB 3.0 Credential (stored in cert_credentials.oba_credential)
{
  "@context": [
    "https://www.w3.org/2018/credentials/v1",
    "https://purl.imsglobal.org/spec/ob/v3p0/context-3.0.3.json"
  ],
  "type": ["VerifiableCredential", "AchievementCredential"],
  "issuer": {
    "id": "https://functionfly.com/issuers/functionfly",
    "type": "Profile",
    "name": "FunctionFly",
    "url": "https://functionfly.com",
    "image": "https://functionfly.com/logo.png"
  },
  "issuanceDate": "2026-03-15T00:00:00Z",
  "expirationDate": "2028-03-15T00:00:00Z",
  "credentialSubject": {
    "id": "did:web:functionfly.com:u:johndoe",
    "type": "AchievementSubject",
    "achievement": {
      "id": "https://functionfly.com/certification/professional",
      "type": "Achievement",
      "name": "FunctionFly Professional",
      "description": "Demonstrates proficiency in orchestration, agents, state management, and security on the FunctionFly platform.",
      "criteria": {
        "narrative": "Pass the FunctionFly Professional certification exam with a score of 70% or higher."
      },
      "image": {
        "id": "https://functionfly.com/badges/professional.svg",
        "type": "Image"
      }
    }
  },
  "credentialSchema": {
    "id": "https://functionfly.com/schemas/certification/v1",
    "type": "JsonSchemaValidator2018"
  },
  "proof": {
    "type": "Ed25519Signature2020",
    "created": "2026-03-15T00:00:00Z",
    "verificationMethod": "https://functionfly.com/issuers/functionfly#key-1",
    "proofPurpose": "assertionMethod",
    "proofValue": "z58DAdFfa9SkqZ6VPz8H5..."
  }
}
```

### Verification Flow

```
Public URL: functionfly.com/verify/{username}
                │
                ▼
    ┌───────────────────────┐
    │ Load credentials      │
    │ from cert_credentials │
    │ WHERE user = username │
    │ AND status = 'active' │
    └───────────┬───────────┘
                │
                ▼
    ┌───────────────────────┐
    │ Verify hash chain:    │
    │ SHA-256(credential    │
    │   payload) matches    │
    │   stored hash         │
    └───────────┬───────────┘
                │
                ▼
    ┌───────────────────────┐
    │ Check expiry date     │
    │ against now()         │
    └───────────┬───────────┘
                │
                ▼
    ┌───────────────────────┐
    │ Return verification   │
    │ result + badge data   │
    └───────────────────────┘
```

### Credential Number Format

```
FFC-{YEAR}-{SEQUENCE}
Example: FFC-2026-000142

Generated via:
SELECT nextval('cert_credential_seq_' || tier_slug)
Padded to 6 digits
```

---

## Billing Integration

### Reuses Existing Patterns

| Pattern | Reused From | Application |
|---------|------------|-------------|
| Stripe Checkout | `billing/handler.go` CreateCheckoutSession | Exam payment |
| Payment Method | `billing/handler.go` SetupIntent | Card collection |
| Subscriptions | `billing/models_billing.go` Subscription | Renewal subscriptions |
| Webhook Processing | `billing/handler.go` webhook handler | Exam payment confirmation |
| Invoice Generation | `billing_repository.go` | Exam receipts |
| Usage Events | `billing_repository_usage.go` | Exam attempt tracking |

### Exam Pricing

| Tier | Launch Price | Regular Price | Renewal/yr |
|------|-------------|---------------|------------|
| Associate | FREE (first 500) | $50 | $29 |
| Professional | FREE (first 500) | $150 | $49 |
| Architect | FREE (first 500) | $200 | $69 |
| Enterprise Team | — | $499/yr | $499/yr |

### Exam Start Flow (with billing)

```
POST /v1/certification/tiers/{tierSlug}/start
         │
         ▼
    ┌──────────────────┐
    │ Check: has active│──yes──► Return existing exam
    │ in_progress exam?│
    └────────┬─────────┘
             │ no
             ▼
    ┌──────────────────┐
    │ Check: price = 0?│──yes──► Create exam directly
    └────────┬─────────┘
             │ no
             ▼
    ┌──────────────────┐
    │ Create Stripe    │
    │ Checkout Session │
    │ (exam_pending)   │
    └────────┬─────────┘
             │
             ▼
    ┌──────────────────┐
    │ Return checkout  │
    │ URL to frontend  │
    └────────┬─────────┘
             │ (user completes payment)
             ▼
    ┌──────────────────┐
    │ Stripe webhook   │
    │ checkout.completed│
    └────────┬─────────┘
             │
             ▼
    ┌──────────────────┐
    │ Create exam      │
    │ session + select │
    │ questions        │
    └──────────────────┘
```

---

## Frontend Architecture

### New Files

```
web/dashboard/src/
├── api/
│   └── certification.ts              # API client module
├── hooks/
│   ├── useCertification.ts           # React Query hooks
│   └── useExamTimer.ts              # Exam countdown timer
├── pages/
│   ├── CertificationPage/
│   │   ├── index.tsx                 # Landing page (tier cards)
│   │   ├── ExamPage.tsx             # Active exam UI
│   │   └── ExamResultsPage.tsx      # Results + score breakdown
│   ├── CredentialsPage.tsx          # My credentials list
│   └── VerifyPage.tsx              # Public verification page
├── components/
│   └── certification/
│       ├── TierCard.tsx             # Tier selection card
│       ├── ExamQuestion.tsx         # Question renderer (MC/TF/Multi)
│       ├── ExamTimer.tsx            # Countdown timer bar
│       ├── ExamProgress.tsx         # Progress indicator
│       ├── PracticalChallenge.tsx   # Challenge instructions + deploy
│       ├── CredentialCard.tsx       # Earned credential display
│       ├── CredentialBadge.tsx      # Embeddable badge (profile)
│       ├── VerifyResult.tsx         # Verification page result
│       └── TeamBadgeCard.tsx        # Enterprise team badge
├── lib/
│   └── certification-constants.ts   # Tier colors, icons, routes
└── stores/
    └── examStore.ts                 # Zustand store for exam state
```

### Component Design Notes

- **CredentialBadge** extends existing `AchievementBadge` pattern with certification-specific tier styling
- **ExamQuestion** uses existing `Card` + `RadioGroup`/`Checkbox` from `components/ui/`
- **ExamTimer** uses `useExamTimer` hook with `setInterval` + `visibilityState` API (pause when tab hidden)
- **TierCard** uses existing `glass-card glow hover-lift` pattern
- All pages use `PageLayout` + `PageHeader` wrappers
- Data fetching via `@tanstack/react-query` with query key factory pattern

### Route Additions (in `App.tsx`)

```tsx
{/* Certification routes */}
<Route path="/certification" element={<CertificationPage />} />
<Route path="/certification/exam/:examId" element={<ProtectedRoute><ExamPage /></ProtectedRoute>} />
<Route path="/certification/exam/:examId/results" element={<ProtectedRoute><ExamResultsPage /></ProtectedRoute>} />
<Route path="/credentials" element={<ProtectedRoute><CredentialsPage /></ProtectedRoute>} />
<Route path="/verify/:username" element={<VerifyPage />} />
```

### Sidebar Addition

Add to the "Build" section in `Sidebar.tsx`:
```tsx
{ path: '/certification', label: 'Certification', icon: Award, badge: 'new' },
```

---

## Background Jobs

### 1. Exam Expiry Checker (Cron — every 5 minutes)

```go
// internal/scheduler/cert_exam_expiry.go
// Finds in_progress exams past expires_at, marks as 'expired'
// Pattern: matches existing TrustScoreScheduler
```

### 2. Practical Challenge Grader (DB Queue Worker)

```go
// internal/certification/grader_worker.go
// Polls cert_grading_queue WHERE status='pending'
// Uses SELECT ... FOR UPDATE SKIP LOCKED (matches DNA analysis pattern)
// Calls deployed function, checks response against GradingConfig
// Updates exam practical_results when all challenges graded
```

### 3. Credential Expiry Checker (Cron — daily at 2 AM)

```go
// internal/scheduler/cert_expiry.go
// Finds credentials WHERE expires_at < now() AND status='active'
// Marks as 'expired'
// Sends email notification via existing email service
// Pattern: matches DataRetentionScheduler
```

### 4. Badge Sync (On credential change)

```go
// Internal event: when credential is issued/revoked/expired
// Syncs with achievement system (new 'certification' category)
// Updates user's profile badges
// Updates marketplace listing badges
```

---

## DRY Reuse Map

| What | Reuse From | How |
|------|-----------|-----|
| Handler struct + constructor | `newsletter/handler.go` | Same `Handler` + `NewHandler()` pattern |
| Route registration | `routes.go` → `registerCertificationRoutes()` | Same `PathPrefix` + `Subrouter` pattern |
| Auth middleware | `middleware/auth.go` | `RequireAuth`, `RequirePermission`, `GetUserFromContext` |
| JSON responses | `users/achievements.go` | `writeJSON`, `writeJSONError` helpers |
| Models with GORM | `models_core.go` | Same struct tag pattern, `TableName()` methods |
| JSONB fields | `models_core.go` → `JSONMap` | Reuse existing `JSONMap` type |
| UUID PKs | All models | `gen_random_uuid()` default |
| Migrations | `scripts/create-migration.sh` | Timestamp format, idempotent SQL |
| Stripe integration | `billing/handler.go` | Checkout sessions, webhooks, subscriptions |
| Achievement system | `models_core.go` Achievement | New category `certification` |
| User profile badges | `AchievementBadge` component | Extend for cert badges |
| Background jobs | `scheduler/` + DNA queue pattern | Same cron + FOR UPDATE SKIP LOCKED |
| Email notifications | `internal/email/` | Template-based emails for pass/fail/expiry |
| Rate limiting | `middleware/auth.go` | Exam-specific rate limiter |
| Activity tracking | `UserActivity` model | `activity_type: 'certification_earned'` |
| Page layout | `PageLayout` + `PageHeader` | Same wrappers |
| Data fetching | `@tanstack/react-query` | Query key factory pattern |
| Forms | `react-hook-form` + `zod` | Same validation pattern |
| Toast notifications | `sonner` | `toast.success()` / `toast.error()` |
| State management | Zustand stores | `examStore` for exam session state |

---

## Rollout Phases

### Phase 1 — Foundation (Weeks 1-3)

**Goal:** Working exam engine with Associate tier, free for first 500 users

- [ ] Database migration (all 8 tables)
- [ ] `models_certification.go` (all models)
- [ ] `certification_repository.go` (CRUD + queries)
- [ ] `handlers/certification/exams.go` (start, get, answer, submit)
- [ ] `handlers/certification/tiers.go` (list tiers)
- [ ] Seed Associate tier + initial question bank (50 questions)
- [ ] Seed 3 practical challenges for Associate
- [ ] Exam expiry cron scheduler
- [ ] Dashboard: `CertificationPage` (tier cards)
- [ ] Dashboard: `ExamPage` (timed exam UI)
- [ ] Dashboard: `ExamResultsPage`
- [ ] Auto-grader worker for practical challenges

### Phase 2 — Credentials & Verification (Weeks 4-5)

**Goal:** Verifiable credentials + public verification page

- [ ] `handlers/certification/credentials.go` (list, get, verify)
- [ ] `handlers/certification/verify.go` (public verification)
- [ ] Open Badges 3.0 JSON generation
- [ ] Credential number sequence generator
- [ ] Verification hash computation (SHA-256)
- [ ] Credential expiry cron scheduler
- [ ] Dashboard: `CredentialsPage`
- [ ] Dashboard: `VerifyPage` (public, no auth)
- [ ] Dashboard: `CredentialBadge` (profile integration)
- [ ] Email templates: exam passed, exam failed, credential expiring

### Phase 3 — Billing & Enterprise (Weeks 6-8)

**Goal:** Paid exams + enterprise team badges

- [ ] Stripe Checkout integration for exam payment
- [ ] Stripe webhook handler for exam payment confirmation
- [ ] Renewal subscription system
- [ ] Enterprise team badge creation + management
- [ ] Question bank expansion (200+ per tier)
- [ ] Professional + Architect tier seeds
- [ ] Dashboard: billing flow in exam start
- [ ] Dashboard: subscription management
- [ ] Dashboard: `TeamBadgeCard`

### Phase 4 — Community & Scale (Weeks 9-12)

**Goal:** Community question contributions + analytics

- [ ] Community question submission flow (moderated)
- [ ] Admin dashboard: question management, stats
- [ ] Exam analytics (pass rates by category, question difficulty analysis)
- [ ] Anti-cheating measures (question pool rotation, IP tracking, tab-switch detection)
- [ ] LinkedIn badge sharing integration (OAuth + API)
- [ ] Certificate PDF generation (optional)
- [ ] University/bootcamp partner API (bulk registration)

---

## File Manifest

### Backend (Go)

```
internal/
├── api/
│   ├── handlers/
│   │   └── certification/
│   │       ├── handler.go          # Handler struct, NewHandler(), RegisterRoutes()
│   │       ├── tiers.go            # ListTiers, AdminCreateTier, AdminUpdateTier
│   │       ├── exams.go            # StartExam, GetExam, SubmitAnswer, SubmitExam, ListMyExams
│   │       ├── practical.go        # DeployPractical, SubmitPractical
│   │       ├── credentials.go      # ListMyCredentials, GetCredential, RenewCredential
│   │       ├── verify.go           # VerifyCredential, VerifyByNumber, PublicBadges, OpenBadgesProfile
│   │       ├── subscriptions.go    # CreateSubscription, CancelSubscription, ListMySubscriptions
│   │       ├── team_badges.go      # CreateTeamBadge, GetTeamBadge, TeamBadgeMembers
│   │       ├── admin.go            # AdminCreateQuestion, AdminImportQuestions, AdminUpdateQuestion, etc.
│   │       └── helpers.go          # writeJSON, writeJSONError, selectQuestions, computeScore
│   └── routes.go                   # registerCertificationRoutes() added
├── certification/
│   ├── grader_worker.go            # Practical challenge grading worker
│   ├── question_selector.go        # Stratified random question selection
│   ├── scoring.go                  # Score computation logic
│   └── openbadges.go               # Open Badges 3.0 JSON generation
├── storage/
│   ├── models_certification.go     # All cert models (CertTier, CertQuestion, etc.)
│   ├── certification_repository.go # CRUD + queries for all cert tables
│   └── sql/
│       └── certification_queries.go # Raw SQL for complex queries
├── scheduler/
│   ├── cert_exam_expiry.go         # Exam session expiry checker
│   └── cert_credential_expiry.go   # Credential expiry checker
migrations/
└── 20260503000000_create_certification_tables.up.sql
└── 20260503000000_create_certification_tables.down.sql
```

### Frontend (React/TypeScript)

```
web/dashboard/src/
├── api/
│   └── certification.ts
├── hooks/
│   ├── useCertification.ts
│   └── useExamTimer.ts
├── pages/
│   ├── CertificationPage/
│   │   └── index.tsx
│   ├── ExamPage.tsx
│   ├── ExamResultsPage.tsx
│   ├── CredentialsPage.tsx
│   └── VerifyPage.tsx
├── components/
│   └── certification/
│       ├── TierCard.tsx
│       ├── ExamQuestion.tsx
│       ├── ExamTimer.tsx
│       ├── ExamProgress.tsx
│       ├── PracticalChallenge.tsx
│       ├── CredentialCard.tsx
│       ├── CredentialBadge.tsx
│       ├── VerifyResult.tsx
│       └── TeamBadgeCard.tsx
├── lib/
│   └── certification-constants.ts
└── stores/
    └── examStore.ts
```

### Seed Data

```
scripts/
├── seed-cert-associate.sql       # 50 knowledge questions + 3 practical challenges
├── seed-cert-professional.sql    # 50 knowledge questions + 3 practical challenges
└── seed-cert-architect.sql       # 50 knowledge questions + 3 practical challenges
```

---

## Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `CERTIFICATION_ENABLED` | Feature flag for entire system | `false` |
| `CERTIFICATION_FREE_SLOTS` | Free exam slots for launch promo | `500` |
| `CERTIFICATION_GRADER_WORKERS` | Number of grading workers | `2` |
| `CERTIFICATION_GRADER_POLL_INTERVAL` | Grading queue poll interval | `10s` |
| `CERTIFICATION_OPENBADGES_ISSUER_URL` | Issuer URL for OB credentials | `https://functionfly.com/issuers/functionfly` |
| `CERTIFICATION_OPENBADGES_PRIVATE_KEY` | Ed25519 private key for signing | (required) |
| `CERTIFICATION_STRIPE_PRODUCT_ID` | Stripe product for exam pricing | (required for paid) |

---

## Anti-Cheating Measures

1. **Question pool rotation** — 200+ questions per tier, randomized each session
2. **Answer order randomization** — option positions shuffled per question
3. **Tab-switch detection** — frontend tracks focus/blur events, logged to exam metadata
4. **Time pressure** — strict time limits with server-side expiry enforcement
5. **IP logging** — exam sessions record IP address for anomaly detection
6. **One active exam** — users can only have one in_progress exam per tier at a time
7. **Cooldown period** — 24-hour cooldown between failed attempts (configurable)
8. **Rate limiting** — exam endpoints get their own rate limiter (5 attempts/hour)
