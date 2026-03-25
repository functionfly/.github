---
title: "TRUST PROTOCOL SPEC"
---

# FunctionFly Trust Protocol Specification

**Version**: 1.0.0-draft  
**Status**: Draft for Q4 2026 Publication  
**Date**: 2026-03-21  

---

## Table of Contents

1. [Overview](#1-overview)
2. [Trust Score Calculation](#2-trust-score-calculation)
3. [Verification Levels](#3-verification-levels)
4. [Agent Communication Protocol](#4-agent-communication-protocol)
5. [API Specification](#5-api-specification)
6. [Trust Tiers](#6-trust-tiers)
7. [Data Models](#7-data-models)
8. [Security Considerations](#8-security-considerations)
9. [Appendix: Reference Implementation](#9-appendix-reference-implementation)

---

## 1. Overview

### 1.1 Purpose and Scope

The FunctionFly Trust Protocol establishes a standardized framework for evaluating, verifying, and communicating the trustworthiness of executable functions in a distributed AI agent ecosystem. This protocol enables:

- **AI agents** to make informed decisions about which functions to invoke based on trustworthiness
- **Function publishers** to demonstrate quality through independent verification
- **Platform operators** to enforce trust-based access policies
- **End users** to understand the reliability of AI tool integrations

### 1.2 Key Concepts

| Term | Definition |
|------|------------|
| **Trust Score** | A composite 0-100 score representing overall function trustworthiness |
| **Trust Tier** | A categorical ranking (Untrusted → Excellent) derived from the Trust Score |
| **Verification** | The process of independently validating function behavior and security |
| **Verification Level** | The depth/rigor of verification performed (Basic → Enterprise) |
| **Agent** | An AI system capable of discovering and executing functions via this protocol |
| **Function** | A callable unit of executable code published to the FunctionFly registry |

### 1.3 Design Principles

1. **Transparency**: All trust components are measurable and independently verifiable
2. **Composability**: Trust scores combine multiple factors with configurable weights
3. **Incremental Trust**: Functions can improve their trust standing over time
4. **Agent-Native**: Protocol designed for programmatic trust evaluation by AI systems
5. **Privacy-Preserving**: Trust data does not expose proprietary function internals

---

## 2. Trust Score Calculation

### 2.1 Score Components

The Trust Score is a weighted composite of five components:

| Component | Weight | Description | Range |
|-----------|--------|-------------|-------|
| **Reliability** | 0.30 | Historical execution success rate | 0-100 |
| **Latency** | 0.20 | Performance consistency and speed | 0-100 |
| **Error Rate** | 0.20 | Frequency of execution errors | 0-100 |
| **User Rating** | 0.15 | Aggregate user satisfaction scores | 0-100 |
| **Verification** | 0.15 | Verification status bonus | 0-15 |

### 2.2 Component Calculations

#### 2.2.1 Reliability Score

```
ReliabilityScore = (SuccessfulExecutions / TotalExecutions) × 100
```

- **Window**: Last 100 executions or 24 hours, whichever is longer
- **Successful**: Executions that returned valid responses without errors
- **Minimum sample size**: 10 executions for score to be meaningful

#### 2.2.2 Latency Score

```
LatencyScore = max(0, 100 - (AverageLatencyMs / MaxAcceptableLatencyMs) × 100)
```

- **MaxAcceptableLatencyMs**: 5000ms (configurable per function category)
- **Calculation**: Exponential moving average with α = 0.1 for recent bias
- **Penalty**: 50% score reduction for functions exceeding p99 latency threshold

#### 2.2.3 Error Rate Score

```
ErrorRateScore = max(0, 100 - (ErrorRate × 100))
```

- **ErrorRate**: Errors / TotalExecutions in the measurement window
- **Error types counted**: Timeouts, exceptions, validation failures, system errors
- **Error types excluded**: Client errors (4xx), intentional rejections

#### 2.2.4 User Rating Score

```
UserRatingScore = WeightedAverage(Ratings, Weights)
```

- **Weight factors**: Recent ratings weighted higher (time decay α = 0.05)
- **Minimum ratings**: 3 ratings required for score
- **Diversity bonus**: +5 points if ratings come from 5+ unique users

#### 2.2.5 Verification Bonus

```
VerificationBonus = VerificationLevelBonus × VerificationFreshnessFactor
```

| Verification Level | Base Bonus |
|--------------------|------------|
| None | 0 |
| Basic | 3 |
| Standard | 7 |
| Advanced | 11 |
| Enterprise | 15 |

- **FreshnessFactor**: `min(1.0, daysSinceVerification / 30)` — decays over 30 days
- **Re-verification** resets the freshness factor

### 2.3 Trust Score Formula

```
TrustScore = (ReliabilityScore × 0.30) +
             (LatencyScore × 0.20) +
             (ErrorRateScore × 0.20) +
             (UserRatingScore × 0.15) +
             (VerificationBonus × 0.15)
```

**Note**: The verification component contributes 0-15 points (not 0-100), making it a bonus that can push otherwise-equivalent functions above unverified alternatives.

### 2.4 Weight Configuration

Default weights are recommended but implementers MAY use custom weights:

```yaml
trust_score_weights:
  reliability: 0.30
  latency: 0.20
  error_rate: 0.20
  user_rating: 0.15
  verification: 0.15
```

---

## 3. Verification Levels

### 3.1 Level Definitions

| Level | Name | Description | Use Case |
|-------|------|-------------|----------|
| **0** | None | No verification performed | Testing, internal functions |
| **1** | Basic | Automated security scan | Public functions, prototypes |
| **2** | Standard | Automated + deterministic execution test | Production functions |
| **3** | Advanced | Standard + manual code review | High-stakes functions |
| **4** | Enterprise | Advanced + compliance audit | Enterprise deployments |

### 3.2 Requirements by Level

#### Level 0: None

- No requirements
- Trust score based entirely on execution metrics and user ratings
- Appropriate for: Internal functions, development/testing

#### Level 1: Basic

- Automated malware scan (static analysis)
- Dependency vulnerability check
- Basic input validation testing
- **Turnaround**: Automated, immediate

#### Level 2: Standard

- All Level 1 requirements
- Deterministic execution verification (DRE)
- Output consistency testing across multiple runs
- Runtime resource consumption validation
- **Turnaround**: < 1 hour

#### Level 3: Advanced

- All Level 2 requirements
- Manual code review by human reviewer
- Security penetration testing
- Documentation review
- **Turnaround**: 1-3 business days

#### Level 4: Enterprise

- All Level 3 requirements
- SOC 2 Type II audit compatibility
- Custom compliance requirements
- Legal/contractual review
- **Turnaround**: 1-4 weeks

### 3.3 Verification Process

```
┌─────────────────────────────────────────────────────────────┐
│                    Verification Pipeline                     │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌─────────┐ │
│  │ Submit   │───▶│ Security │───▶│   DRE    │───▶│ Manual  │ │
│  │ Request  │    │  Scan    │    │  Test    │    │ Review  │ │
│  └──────────┘    └──────────┘    └──────────┘    └─────────┘ │
│       │              │              │              │        │
│       ▼              ▼              ▼              ▼        │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Verification Certificate                 │   │
│  │  (FXCERT with trust_score, determinism, signature)   │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 3.4 Verification Certificate (FXCERT)

Upon successful verification, functions receive an FXCERT containing:

```json
{
  "certificate_id": "vfy_abc123...",
  "function_id": "uuid",
  "version": "1.0.0",
  "issued_at": "2026-03-21T00:00:00Z",
  "expires_at": "2026-04-21T00:00:00Z",
  "level": "standard",
  "trust": {
    "trust_score": 87.5,
    "determinism_score": 0.95,
    "reproducibility": true
  },
  "signature": "base64_signature"
}
```

---

## 4. Agent Communication Protocol

### 4.1 Function Discovery

Agents discover functions through the registry with trust-aware filtering:

```
GET /v1/registry/functions?trust_score_min={threshold}&verified={bool}&category={name}
```

**Trust-Aware Discovery Algorithm**:

1. Query registry with minimum trust threshold
2. Sort by composite score: `(trust_score × 0.7) + (popularity × 0.3)`
3. Filter by verification level if required by policy
4. Return ranked list with trust metadata

### 4.2 Trust Evaluation Flow

```
┌─────────┐     ┌──────────────┐     ┌─────────────┐     ┌────────────┐
│  Agent  │────▶│  Discover    │────▶│  Evaluate   │────▶│  Execute   │
│         │     │  Functions   │     │  Trust      │     │  + Monitor │
└─────────┘     └──────────────┘     └─────────────┘     └────────────┘
                                          │
                                          ▼
                                   ┌──────────────┐
                                   │  Update      │
                                   │  Trust Score │
                                   └──────────────┘
```

### 4.3 Trust-Aware Fallback Protocol

When a function invocation fails, agents SHOULD:

1. **Check if failure is trust-related**:
   - Error type: `TRUST_VIOLATION` indicates verification failure
   - Error type: `EXECUTION_ERROR` indicates performance issue

2. **Find fallback candidates**:

   ```json
   GET /v1/registry/functions?capability={capability}&trust_score_min={score}
   ```

3. **Evaluate fallback trust**:

   ```json
   GET /v1/trust/score/{function_id}
   ```

4. **Select best fallback**:
   - Prefer higher trust scores
   - Consider latency/cost tradeoffs
   - Verify fallback capability equivalence

### 4.4 Trust Metadata in Agent Context

When an agent caches function metadata, include:

```json
{
  "function_id": "uuid",
  "name": "function_name",
  "author": "publisher_id",
  "trust": {
    "score": 87.5,
    "tier": "excellent",
    "verified": true,
    "verification_level": "standard",
    "last_verified": "2026-03-15T00:00:00Z"
  },
  "metrics": {
    "success_rate": 0.98,
    "avg_latency_ms": 245,
    "total_executions": 15420
  }
}
```

---

## 5. API Specification

### 5.1 Authentication

Trust API uses API key authentication:

- **Header**: `X-API-Key: {api_key}`
- **Partner ID**: `X-Partner-ID: {partner_id}`
- **Rate Limits**: Per partner tier (see Section 5.6)

### 5.2 Endpoints Overview

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/v1/trust/score/{function_id}` | Get trust score for a function | API Key |
| POST | `/v1/trust/batch` | Get trust scores for multiple functions | API Key |
| GET | `/v1/trust/history/{function_id}` | Get trust score history | API Key |
| POST | `/v1/trust/verify` | Submit function for verification | API Key + Scope |
| GET | `/v1/trust/verify/{verification_id}` | Get verification status | API Key |
| POST | `/v1/trust/report` | Report trust issue | API Key + Scope |
| GET | `/v1/trust/report/{report_id}` | Get report status | API Key |

### 5.3 Get Trust Score

```
GET /v1/trust/score/{function_id}
```

**Response**:

```json
{
  "function_id": "550e8400-e29b-41d4-a716-446655440000",
  "trust_score": 87.5,
  "trust_tier": "excellent",
  "is_verified": true,
  "verification_level": "standard",
  "last_updated": "2026-03-21T10:30:00Z",
  "components": {
    "reliability": 95.0,
    "latency": 82.5,
    "error_rate": 98.0,
    "user_rating": 78.0,
    "verification": 7.0
  },
  "execution_stats": {
    "total_calls": 15420,
    "success_rate": 0.98,
    "avg_latency_ms": 245,
    "p99_latency_ms": 890
  }
}
```

### 5.4 Batch Trust Score

```
POST /v1/trust/batch
```

**Request**:

```json
{
  "function_ids": [
    "550e8400-e29b-41d4-a716-446655440000",
    "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
  ]
}
```

**Response**:

```json
{
  "scores": [
    {
      "function_id": "550e8400-e29b-41d4-a716-446655440000",
      "trust_score": 87.5,
      "trust_tier": "excellent",
      "is_verified": true,
      "last_updated": "2026-03-21T10:30:00Z"
    },
    {
      "function_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
      "trust_score": 62.3,
      "trust_tier": "good",
      "is_verified": false,
      "last_updated": "2026-03-20T15:45:00Z"
    }
  ],
  "errors": []
}
```

### 5.5 Get Trust History

```
GET /v1/trust/history/{function_id}?page=1&page_size=20
```

**Response**:

```json
{
  "function_id": "550e8400-e29b-41d4-a716-446655440000",
  "history": [
    {
      "trust_score": 87.5,
      "trust_tier": "excellent",
      "calculated_at": "2026-03-21T10:00:00Z",
      "components": {
        "reliability": 95.0,
        "latency": 82.5,
        "error_rate": 98.0,
        "user_rating": 78.0,
        "verification": 7.0
      }
    },
    {
      "trust_score": 85.2,
      "trust_tier": "excellent",
      "calculated_at": "2026-03-20T10:00:00Z",
      "components": {
        "reliability": 94.0,
        "latency": 80.0,
        "error_rate": 97.0,
        "user_rating": 78.0,
        "verification": 7.0
      }
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 45
  }
}
```

### 5.6 Submit Verification

```
POST /v1/trust/verify
```

**Request**:

```json
{
  "function_id": "550e8400-e29b-41d4-a716-446655440000",
  "function_version": "1.0.0",
  "verification_level": "standard",
  "metadata": {
    "repository_url": "https://github.com/author/repo",
    "documentation_url": "https://docs.example.com/function"
  }
}
```

**Response**:

```json
{
  "id": "uuid",
  "verification_id": "vfy_abc123def456",
  "function_id": "550e8400-e29b-41d4-a716-446655440000",
  "verification_level": "standard",
  "status": "pending",
  "created_at": "2026-03-21T10:00:00Z",
  "estimated_completion": "2026-03-21T11:00:00Z"
}
```

### 5.7 Partner Rate Limits

| Tier | Monthly Requests | Rate (req/min) | Verification |
|------|------------------|----------------|--------------|
| Developer | 1,000 | 10 | No |
| Startup | 10,000 | 100 | Yes |
| Business | 100,000 | 1,000 | Yes |
| Enterprise | Unlimited | Custom | Yes |

---

## 6. Trust Tiers

### 6.1 Tier Definitions

| Tier | Score Range | Badge Color | Description |
|------|-------------|-------------|-------------|
| **Excellent** | 80-100 | Green | Highly reliable, verified function |
| **Good** | 60-79 | Blue | Solid performance, minor concerns |
| **Fair** | 40-59 | Yellow | Average performance, use with caution |
| **Poor** | 20-39 | Orange | Significant issues, not recommended |
| **Very Poor** | 1-19 | Red | Severe issues, avoid if possible |
| **Untrusted** | 0 | Gray | New function with no data |

### 6.2 Tier Display Requirements

| Tier | Show Trust Badge | Show Components | Require Verification Prompt |
|------|------------------|-----------------|----------------------------|
| Excellent | Yes | Optional | No |
| Good | Yes | Optional | No |
| Fair | Yes | Recommended | Yes (first use) |
| Poor | Yes | Required | Yes |
| Very Poor | Yes | Required | Yes |
| Untrusted | No | N/A | Yes |

---

## 7. Data Models

### 7.1 TrustScoreResponse

```go
type TrustScoreResponse struct {
    FunctionID        uuid.UUID   `json:"function_id"`
    TrustScore        float64     `json:"trust_score"`         // 0-100
    TrustTier         TrustTier   `json:"trust_tier"`          // e.g., "excellent"
    IsVerified        bool        `json:"is_verified"`
    VerificationLevel string      `json:"verification_level"`  // e.g., "standard"
    LastUpdated       time.Time   `json:"last_updated"`
    Components        Components  `json:"components"`
}

type Components struct {
    Reliability  float64 `json:"reliability"`  // 0-100
    Latency      float64 `json:"latency"`      // 0-100
    ErrorRate    float64 `json:"error_rate"`  // 0-100
    UserRating   float64 `json:"user_rating"` // 0-100
    Verification float64 `json:"verification"` // 0-15
}
```

### 7.2 TrustHistory

```go
type TrustHistory struct {
    FunctionID        uuid.UUID   `json:"function_id"`
    TrustScore        float64     `json:"trust_score"`
    ReliabilityScore  float64     `json:"reliability_score"`
    LatencyScore      float64     `json:"latency_score"`
    ErrorRateScore    float64     `json:"error_rate_score"`
    UserRatingScore   float64     `json:"user_rating_score"`
    VerificationBonus float64     `json:"verification_bonus"`
    TotalCalls        int         `json:"total_calls"`
    TrustTier         TrustTier   `json:"trust_tier"`
    CalculatedAt      time.Time   `json:"calculated_at"`
    IsVerified        bool        `json:"is_verified"`
    VerificationLevel string      `json:"verification_level"`
}
```

### 7.3 TrustTier Enum

```go
type TrustTier string

const (
    TrustTierExcellent  TrustTier = "excellent"
    TrustTierGood       TrustTier = "good"
    TrustTierFair       TrustTier = "fair"
    TrustTierPoor       TrustTier = "poor"
    TrustTierVeryPoor   TrustTier = "very_poor"
    TrustTierUntrusted  TrustTier = "untrusted"
)
```

---

## 8. Security Considerations

### 8.1 Trust Score Manipulation Prevention

- **Execution counting**: Server-side only, clients cannot spoof
- **User ratings**: One rating per user per function, rate-limited
- **Verification**: Cryptographically signed certificates
- **History immutability**: Trust score history is append-only

### 8.2 Privacy

- Trust scores do NOT reveal:
  - Function source code or internals
  - Publisher business logic
  - User identity (for ratings)
  - Execution payloads or results

### 8.3 API Security

- All Trust API endpoints require API key authentication
- Partner keys have scoped permissions (trust:read, verification:request, reports:submit)
- Rate limiting per partner tier prevents abuse
- Usage tracking for billing and anomaly detection

---

## 9. Appendix: Reference Implementation

### 9.1 Default Weights (Go)

Reference implementation from [`internal/storage/registry/types.go:571`](internal/storage/registry/types.go:571):

```go
func DefaultTrustScoreWeights() TrustScoreWeights {
    return TrustScoreWeights{
        Reliability:  0.30,
        Latency:     0.20,
        ErrorRate:    0.20,
        UserRating:  0.15,
        Verification: 0.15,
    }
}
```

### 9.2 Trust Score Calculation

Reference implementation from [`internal/storage/registry/trust_repository.go:19`](internal/storage/registry/trust_repository.go:19):

```go
func (r *RegistryRepository) CalculateTrustScore(functionID uuid.UUID, windowStart, windowEnd time.Time) (*TrustHistory, error) {
    // Get execution metrics for the window
    metrics, err := r.GetExecutionMetrics(functionID, windowStart, windowEnd)
    
    // Get user ratings
    rating, _ := r.GetRatingByFunctionID(functionID)
    
    // Get verification status
    isVerified, verificationLevel, _ := r.GetFunctionVerificationStatus(functionID)
    
    // Calculate components
    reliabilityScore := calculateReliabilityScore(metrics)
    latencyScore := calculateLatencyScore(metrics)
    errorRateScore := calculateErrorRateScore(metrics)
    userRatingScore := calculateUserRatingScore(rating)
    verificationBonus := calculateVerificationBonus(isVerified, verificationLevel)
    
    // Composite score
    trustScore := (reliabilityScore * 0.30) +
                  (latencyScore * 0.20) +
                  (errorRateScore * 0.20) +
                  (userRatingScore * 0.15) +
                  (verificationBonus * 0.15)
    
    trustTier := getTrustTierFromScore(trustScore)
    
    return &TrustHistory{
        FunctionID:         functionID,
        TrustScore:         trustScore,
        ReliabilityScore:    reliabilityScore,
        LatencyScore:       latencyScore,
        ErrorRateScore:     errorRateScore,
        UserRatingScore:    userRatingScore,
        VerificationBonus:  verificationBonus,
        TotalCalls:         metrics.TotalCalls,
        TrustTier:          trustTier,
        CalculatedAt:       time.Now(),
        IsVerified:         isVerified,
        VerificationLevel:  verificationLevel,
    }, nil
}
```

### 9.3 Verification Levels

Reference implementation from [`internal/storage/trustapi/models.go:329`](internal/storage/trustapi/models.go:329):

```go
type VerificationLevel string

const (
    VerificationLevelBasic      VerificationLevel = "basic"
    VerificationLevelStandard   VerificationLevel = "standard"
    VerificationLevelAdvanced   VerificationLevel = "advanced"
    VerificationLevelEnterprise VerificationLevel = "enterprise"
)
```

---

## Document History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0-draft | 2026-03-21 | Initial draft for Q4 2026 publication |

---

**Next Review**: Q2 2026  
**Owner**: FunctionFly Platform Team  
**Contact**: <trust-protocol@functionfly.dev>
