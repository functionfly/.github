---
title: Trust API for AI Models & Tools
description: Integrate trust verification into your AI model pipelines and toolchains.
sidebar:
  order: 10
---

The FunctionFly Trust API provides comprehensive trust scoring, verification, and revocation management for AI models and tools. This guide covers integration patterns for AI applications, model marketplaces, and toolchains.

## Overview

The Trust API enables AI platforms to:

- **Verify AI model authenticity** before execution
- **Check trust scores** for tools and functions
- **Validate attestation chains** for compliance
- **Enforce trust policies** based on model provenance
- **Monitor revocation status** in real-time

## Authentication

### API Key Authentication

All Trust API requests require a partner API key:

```bash
curl -H "X-API-Key: fft_your_partner_key" \
  https://api.functionfly.com/v1/trust/score/{function_id}
```

### Available Scopes

| Scope | Description |
|-------|-------------|
| `trust:read` | Read trust scores and history |
| `trust:write` | Submit trust reports |
| `verification:request` | Request function verification |
| `reports:submit` | Submit trust issue reports |
| `partners:manage` | Manage partner account (admin) |

## Trust Score Endpoints

### Get Trust Score

Retrieve the trust score for a specific function or AI model:

```bash
curl -X GET "https://api.functionfly.com/v1/trust/score/{function_id}" \
  -H "X-API-Key: fft_your_key"
```

**Response:**

```json
{
  "function_id": "550e8400-e29b-41d4-a716-446655440000",
  "trust_score": 87.5,
  "trust_tier": "verified",
  "is_verified": true,
  "verification_level": "standard",
  "last_updated": "2026-06-08T12:00:00Z",
  "components": {
    "reliability": 92.0,
    "latency": 88.5,
    "error_rate": 85.0,
    "user_rating": 90.0,
    "verification": 95.0
  },
  "metrics": {
    "total_calls": 125000,
    "success_rate": 99.2,
    "p50_latency_ms": 45,
    "p95_latency_ms": 120,
    "p99_latency_ms": 250,
    "error_rate": 0.8,
    "timeout_rate": 0.2
  }
}
```

### Batch Trust Score Lookup

Query trust scores for multiple functions at once:

```bash
curl -X POST "https://api.functionfly.com/v1/trust/batch" \
  -H "X-API-Key: fft_your_key" \
  -H "Content-Type: application/json" \
  -d '{
    "function_ids": [
      "550e8400-e29b-41d4-a716-446655440000",
      "660e8400-e29b-41d4-a716-446655440001"
    ]
  }'
```

**Response:**

```json
{
  "scores": [
    {
      "function_id": "550e8400-e29b-41d4-a716-446655440000",
      "trust_score": 87.5,
      "trust_tier": "verified",
      "is_verified": true,
      "components": { ... },
      "metrics": { ... }
    }
  ],
  "errors": [
    {
      "function_id": "660e8400-e29b-41d4-a716-446655440001",
      "error": "Trust score not found"
    }
  ]
}
```

### Trust Score History

Get historical trust scores for trend analysis:

```bash
curl -X GET "https://api.functionfly.com/v1/trust/history/{function_id}?page=1&page_size=20" \
  -H "X-API-Key: fft_your_key"
```

## Verification Endpoints

### Submit Verification Request

Request verification for an AI model or tool:

```bash
curl -X POST "https://api.functionfly.com/v1/trust/verify" \
  -H "X-API-Key: fft_your_key" \
  -H "Content-Type: application/json" \
  -d '{
    "function_id": "550e8400-e29b-41d4-a716-446655440000",
    "function_version": "2.1.0",
    "verification_level": "standard",
    "metadata": {
      "model_type": "llm",
      "training_data_source": "verified",
      "safety_evaluation": "passed"
    }
  }'
```

**Verification Levels:**

| Level | Description |
|-------|-------------|
| `basic` | Identity verification only |
| `standard` | Security + performance review |
| `advanced` | Comprehensive audit + safety testing |
| `enterprise` | Full compliance + custom requirements |

### Get Verification Status

Check the status of a verification request:

```bash
curl -X GET "https://api.functionfly.com/v1/trust/verify/ver_abc123" \
  -H "X-API-Key: fft_your_key"
```

## Trust Reports

### Submit a Trust Issue Report

Report issues with an AI model or tool:

```bash
curl -X POST "https://api.functionfly.com/v1/trust/report" \
  -H "X-API-Key: fft_your_key" \
  -H "Content-Type: application/json" \
  -d '{
    "function_id": "550e8400-e29b-41d4-a716-446655440000",
    "report_type": "misinformation",
    "severity": "high",
    "title": "Hallucination in medical domain",
    "description": "Model produces confident but incorrect medical claims...",
    "evidence": {
      "example_prompts": ["What is the treatment for..."],
      "incorrect_outputs": ["Claimed X is Y when actually Z"],
      "frequency": "frequent"
    }
  }'
```

**Report Types:**

| Type | Description |
|------|-------------|
| `malware` | Model exhibits malicious behavior |
| `phishing` | Model used for credential harvesting |
| `data_leak` | Model leaks sensitive training data |
| `abuse` | Model promotes harmful content |
| `misinformation` | Model produces false claims |
| `other` | Other issues |

### Get Report Status

Check the status of a submitted report:

```bash
curl -X GET "https://api.functionfly.com/v1/trust/report/rpt_abc123" \
  -H "X-API-Key: fft_your_key"
```

## Revocation Management

### Check Revocation Status

Verify if a function is revoked:

```bash
curl -X GET "https://api.functionfly.com/v1/trust/revoked/{function_id}" \
  -H "X-API-Key: fft_your_key"
```

**Response (revoked):**

```json
{
  "function_id": "550e8400-e29b-41d4-a716-446655440000",
  "is_revoked": true,
  "revocation_id": "rev_abc123",
  "reason": "safety_concern",
  "severity": "critical",
  "revoked_at": "2026-06-01T10:00:00Z",
  "revocation_type": "emergency",
  "impact_description": "Model exhibits unsafe behavior under specific inputs"
}
```

### List All Revocations

Get a list of all revoked functions (admin only):

```bash
curl -X GET "https://api.functionfly.com/v1/trust/revoke/revoked?page=1&page_size=20" \
  -H "X-API-Key: fft_admin_key"
```

## Attestation Endpoints

### Get Attestations

List attestations for a function:

```bash
curl -X GET "https://api.functionfly.com/v1/trust/attestations?function_id={function_id}" \
  -H "X-API-Key: fft_your_key"
```

### Verify Attestation Integrity

Cryptographically verify an attestation:

```bash
curl -X GET "https://api.functionfly.com/v1/trust/attestations/{attestation_id}/verify" \
  -H "X-API-Key: fft_your_key"
```

### Get Attestation Chain

Get the full chain of attestations for audit:

```bash
curl -X GET "https://api.functionfly.com/v1/trust/attestations/{function_id}/chain" \
  -H "X-API-Key: fft_your_key"
```

## Policy Evaluation

### Evaluate Against Trust Policy

Evaluate a function against your trust policy:

```bash
curl -X POST "https://api.functionfly.com/v1/trust/policies/evaluate" \
  -H "X-API-Key: fft_your_key" \
  -H "Content-Type: application/json" \
  -d '{
    "function_id": "550e8400-e29b-41d4-a716-446655440000",
    "policy_id": "pol_abc123"
  }'
```

**Response:**

```json
{
  "result": {
    "evaluation_id": "eval_xyz789",
    "policy_id": "pol_abc123",
    "function_id": "550e8400-e29b-41d4-a716-446655440000",
    "function_author": "model-provider",
    "function_name": "gpt4-wrapper",
    "result": "allowed",
    "decision": "policy_rule_passed",
    "reason": "Rule 'min_trust_score' passed",
    "trust_score": 87.5,
    "trust_tier": "verified",
    "is_verified": true,
    "is_revoked": false,
    "rule_results": [
      {
        "rule_id": "min_trust_score",
        "type": "min_trust_score",
        "passed": true,
        "actual_value": 87.5,
        "expected_value": 50.0,
        "reason": ""
      },
      {
        "rule_id": "verification_required",
        "type": "verification_required",
        "passed": true,
        "actual_value": true,
        "expected_value": true,
        "reason": ""
      }
    ],
    "evaluated_at": "2026-06-08T16:00:00Z",
    "cache_valid_until": "2026-06-08T16:05:00Z"
  },
  "cached": false
}
```

### Batch Policy Evaluation

Evaluate multiple functions at once:

```bash
curl -X POST "https://api.functionfly.com/v1/trust/policies/evaluate/batch" \
  -H "X-API-Key: fft_your_key" \
  -H "Content-Type: application/json" \
  -d '{
    "function_ids": [
      "550e8400-e29b-41d4-a716-446655440000",
      "660e8400-e29b-41d4-a716-446655440001"
    ],
    "policy_id": "pol_abc123"
  }'
```

## AI Model Integration Examples

### Python: Trust Score Lookup

```python
import requests

class AITrustClient:
    def __init__(self, api_key, base_url="https://api.functionfly.com"):
        self.api_key = api_key
        self.base_url = base_url
        self.headers = {"X-API-Key": api_key}
    
    def get_trust_score(self, function_id):
        """Get trust score for an AI model"""
        response = requests.get(
            f"{self.base_url}/v1/trust/score/{function_id}",
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()
    
    def verify_model(self, function_id, version, level="standard"):
        """Request verification for an AI model"""
        response = requests.post(
            f"{self.base_url}/v1/trust/verify",
            headers=self.headers,
            json={
                "function_id": function_id,
                "function_version": version,
                "verification_level": level,
                "metadata": {
                    "model_type": "llm",
                    "safety_evaluation": "pending"
                }
            }
        )
        response.raise_for_status()
        return response.json()
    
    def evaluate_policy(self, function_id, policy_id=None):
        """Evaluate model against trust policy"""
        payload = {"function_id": function_id}
        if policy_id:
            payload["policy_id"] = policy_id
        
        response = requests.post(
            f"{self.base_url}/v1/trust/policies/evaluate",
            headers=self.headers,
            json=payload
        )
        response.raise_for_status()
        return response.json()["result"]

# Usage
client = AITrustClient(api_key="fft_your_key")

# Check trust score before using model
score = client.get_trust_score("550e8400-e29b-41d4-a716-446655440000")
if score["trust_score"] < 50:
    print(f"Warning: Low trust score ({score['trust_score']})")
    print(f"Tier: {score['trust_tier']}")
```

### Python: Pre-Execution Trust Check

```python
import requests

def execute_with_trust_check(function_id, payload, policy_id=None):
    """
    Execute a function only if it passes trust evaluation.
    Ideal for AI agent tool selection.
    """
    client = AITrustClient(api_key="fft_your_key")
    
    # Evaluate trust before execution
    result = client.evaluate_policy(function_id, policy_id)
    
    if result["result"] == "denied":
        raise RuntimeError(
            f"Function {function_id} blocked by trust policy: {result['reason']}"
        )
    
    if result["result"] == "warned":
        print(f"Warning: {result['reason']}")
    
    if result["is_revoked"]:
        raise RuntimeError(f"Function {function_id} has been revoked")
    
    # Proceed with execution
    response = requests.post(
        f"https://api.functionfly.com/v1/execute/{function_id}",
        json=payload,
        headers={"X-API-Key": "fft_your_key"}
    )
    return response.json()

# Usage in AI agent
try:
    result = execute_with_trust_check(
        "550e8400-e29b-41d4-a716-446655440000",
        {"prompt": "Analyze this data..."}
    )
except RuntimeError as e:
    print(f"Tool blocked: {e}")
    # Fall back to alternative tool
```

### JavaScript: Batch Model Evaluation

```javascript
class AITrustVerifier {
  constructor(apiKey) {
    this.apiKey = apiKey;
    this.baseUrl = 'https://api.functionfly.com';
  }

  async evaluateModels(functionIds, policyId) {
    const response = await fetch(`${this.baseUrl}/v1/trust/policies/evaluate/batch`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey
      },
      body: JSON.stringify({
        function_ids: functionIds,
        policy_id: policyId
      })
    });

    const data = await response.json();
    return data.results.filter(r => r.result === 'allowed');
  }

  async selectTrustedTools(tools, policyId) {
    const functionIds = tools.map(t => t.function_id);
    const trusted = await this.evaluateModels(functionIds, policyId);
    
    const trustedIds = new Set(trusted.map(t => t.function_id));
    return tools.filter(t => trustedIds.has(t.function_id));
  }
}

// Usage in AI agent
const verifier = new AITrustVerifier('fft_your_key');

const availableTools = [
  { function_id: '550e8400-e29b-41d4-a716-446655440000', name: 'gpt4-analyzer' },
  { function_id: '660e8400-e29b-41d4-a716-446655440001', name: 'claude-summarizer' },
  { function_id: '770e8400-e29b-41d4-a716-446655440002', name: 'local-llm' }
];

// Filter to only trusted tools
const trustedTools = await verifier.selectTrustedTools(
  availableTools,
  'pol_your_policy_id'
);

console.log(`Using ${trustedTools.length} trusted tools`);
```

### Go: Trust-Aware Tool Selection

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type TrustClient struct {
    APIKey    string
    BaseURL   string
    Client    *http.Client
}

type TrustScore struct {
    FunctionID    string  `json:"function_id"`
    TrustScore    float64 `json:"trust_score"`
    TrustTier     string  `json:"trust_tier"`
    IsVerified    bool    `json:"is_verified"`
    IsRevoked     bool    `json:"is_revoked"`
}

func (c *TrustClient) GetTrustScore(functionID string) (*TrustScore, error) {
    req, _ := http.NewRequest("GET", 
        fmt.Sprintf("%s/v1/trust/score/%s", c.BaseURL, functionID), nil)
    req.Header.Set("X-API-Key", c.APIKey)
    
    resp, err := c.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var score TrustScore
    if err := json.NewDecoder(resp.Body).Decode(&score); err != nil {
        return nil, err
    }
    return &score, nil
}

func (c *TrustClient) SelectTrustedTools(toolIDs []string, minScore float64) ([]string, error) {
    trusted := []string{}
    
    for _, id := range toolIDs {
        score, err := c.GetTrustScore(id)
        if err != nil {
            continue
        }
        
        if score.TrustScore >= minScore && !score.IsRevoked {
            trusted = append(trusted, id)
        }
    }
    
    return trusted, nil
}

func main() {
    client := &TrustClient{
        APIKey:  "fft_your_key",
        BaseURL: "https://api.functionfly.com",
        Client:  &http.Client{},
    }
    
    tools := []string{
        "550e8400-e29b-41d4-a716-446655440000",
        "660e8400-e29b-41d4-a716-446655440001",
    }
    
    trusted, err := client.SelectTrustedTools(tools, 70.0)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    
    fmt.Printf("Trusted tools: %v\n", trusted)
}
```

## Trust Tiers

| Tier | Score Range | Description |
|------|-------------|-------------|
| `untrusted` | 0-25 | Function blocked or revoked |
| `trusted` | 26-50 | Basic trust established |
| `verified` | 51-75 | Verified with standard checks |
| `highly_trusted` | 76-100 | Full verification + audit |

## Rate Limits

Rate limits are based on your partner tier:

| Tier | Requests/Minute | Requests/Day |
|------|-----------------|--------------|
| Developer | 60 | 10,000 |
| Startup | 300 | 100,000 |
| Business | 1,000 | 500,000 |
| Enterprise | 10,000 | 10,000,000 |

## Webhooks

Subscribe to trust events for real-time updates:

```bash
# Subscribe to revocation events
ffly events subscribe trust.revocation \
  --webhook https://your-app.com/webhooks/trust
```

**Trust Event Types:**

| Event | Description |
|-------|-------------|
| `trust.score.updated` | Trust score changed |
| `trust.verified` | Function verified |
| `trust.revoked` | Function revoked |
| `trust.revocation.lifted` | Revocation removed |

## Error Codes

| Code | Description |
|------|-------------|
| `invalid_function_id` | Function ID format invalid |
| `trust_not_found` | No trust data for function |
| `function_not_found` | Function does not exist |
| `unauthorized` | Invalid or missing API key |
| `rate_limit_exceeded` | Rate limit hit |
| `policy_not_found` | Trust policy not found |

## SDK Support

Official SDKs for Trust API integration:

- [Python SDK](/sdks/python/)
- [JavaScript/TypeScript SDK](/sdks/javascript/)
- [Go SDK](/sdks/go/)
- [Rust SDK](/sdks/rust/)

## Next Steps

- [Trust API Reference](/trust-api/) - Complete API documentation
- [Trust Protocol Spec](/trust-protocol-spec/) - Technical specification
- [Security Guide](/security/) - Best practices for AI trust
