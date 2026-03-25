# Agent Critical Phases Implementation Roadmap

## Overview

This roadmap provides detailed implementation steps for Phase 1 (Resilience), Phase 2 (Observability), and Phase 3 (Security) - the critical foundation for production-ready agent capabilities.

---

## Phase 1: Resilience & Fault Tolerance

### Sprint 1.1: Circuit Breaker (Week 1)

#### Day 1-2: Core Implementation

**File**: `internal/agent/circuitbreaker/breaker.go`

```go
package circuitbreaker

import (
    "sync"
    "time"
)

type State int

const (
    StateClosed State = iota
    StateOpen
    StateHalfOpen
)

type Config struct {
    FailureThreshold    int           // failures before opening
    SuccessThreshold    int           // successes before closing
    CooldownDuration    time.Duration // time in OPEN before HALF_OPEN
    HalfOpenMaxRequests int           // max requests in HALF_OPEN
    OnStateChange       func(from, to State)
}

type Breaker struct {
    mu              sync.RWMutex
    state           State
    failures        int
    successes       int
    lastFailure     time.Time
    halfOpenCount   int
    config          Config
}

func New(config Config) *Breaker {
    return &Breaker{
        state:  StateClosed,
        config: config,
    }
}

func (b *Breaker) Execute(fn func() error) error {
    if !b.Allow() {
        return ErrCircuitOpen
    }
    
    err := fn()
    b.Record(err)
    return err
}

func (b *Breaker) Allow() bool {
    b.mu.RLock()
    defer b.mu.RUnlock()
    
    switch b.state {
    case StateClosed:
        return true
    case StateOpen:
        if time.Since(b.lastFailure) > b.config.CooldownDuration {
            b.mu.RUnlock()
            b.mu.Lock()
            b.transitionTo(StateHalfOpen)
            b.mu.Unlock()
            b.mu.RLock()
            return true
        }
        return false
    case StateHalfOpen:
        return b.halfOpenCount < b.config.HalfOpenMaxRequests
    }
    return false
}

func (b *Breaker) Record(err error) {
    b.mu.Lock()
    defer b.mu.Unlock()
    
    if err != nil {
        b.failures++
        b.successes = 0
        b.lastFailure = time.Now()
        
        if b.state == StateClosed && b.failures >= b.config.FailureThreshold {
            b.transitionTo(StateOpen)
        } else if b.state == StateHalfOpen {
            b.transitionTo(StateOpen)
        }
    } else {
        b.successes++
        b.failures = 0
        
        if b.state == StateHalfOpen {
            b.halfOpenCount++
            if b.successes >= b.config.SuccessThreshold {
                b.transitionTo(StateClosed)
            }
        }
    }
}

func (b *Breaker) transitionTo(newState State) {
    oldState := b.state
    b.state = newState
    b.failures = 0
    b.successes = 0
    b.halfOpenCount = 0
    
    if b.config.OnStateChange != nil {
        go b.config.OnStateChange(oldState, newState)
    }
}
```

**File**: `internal/agent/circuitbreaker/breaker_test.go`

```go
package circuitbreaker

import (
    "errors"
    "testing"
    "time"
)

func TestBreaker_ClosedToOpen(t *testing.T) {
    b := New(Config{
        FailureThreshold: 3,
        SuccessThreshold: 2,
        CooldownDuration: time.Second,
        HalfOpenMaxRequests: 1,
    })
    
    // Record failures
    for i := 0; i < 3; i++ {
        b.Record(errors.New("fail"))
    }
    
    if b.State() != StateOpen {
        t.Errorf("expected OPEN, got %v", b.State())
    }
}

func TestBreaker_HalfOpenToClosed(t *testing.T) {
    b := New(Config{
        FailureThreshold: 3,
        SuccessThreshold: 2,
        CooldownDuration: time.Millisecond,
        HalfOpenMaxRequests: 2,
    })
    
    // Force to OPEN
    for i := 0; i < 3; i++ {
        b.Record(errors.New("fail"))
    }
    
    // Wait for cooldown
    time.Sleep(2 * time.Millisecond)
    
    // Record successes
    b.Record(nil)
    b.Record(nil)
    
    if b.State() != StateClosed {
        t.Errorf("expected CLOSED, got %v", b.State())
    }
}
```

#### Day 3-4: Integration with Agent Execution

**File**: `internal/api/handlers/agent/execute.go` (modifications)

```go
// Add to Handler struct
type Handler struct {
    // ... existing fields
    circuitBreakers map[string]*circuitbreaker.Breaker
    breakerMu       sync.RWMutex
}

// Add initialization
func NewHandler(...) *Handler {
    h := &Handler{
        // ... existing initialization
        circuitBreakers: make(map[string]*circuitbreaker.Breaker),
    }
    return h
}

// Add breaker retrieval
func (h *Handler) getBreaker(agentID string) *circuitbreaker.Breaker {
    h.breakerMu.RLock()
    breaker, exists := h.circuitBreakers[agentID]
    h.breakerMu.RUnlock()
    
    if exists {
        return breaker
    }
    
    h.breakerMu.Lock()
    defer h.breakerMu.Unlock()
    
    // Double-check
    if breaker, exists := h.circuitBreakers[agentID]; exists {
        return breaker
    }
    
    breaker = circuitbreaker.New(circuitbreaker.Config{
        FailureThreshold:    5,
        SuccessThreshold:    2,
        CooldownDuration:    30 * time.Second,
        HalfOpenMaxRequests: 1,
        OnStateChange: func(from, to circuitbreaker.State) {
            logrus.WithFields(logrus.Fields{
                "agent_id":  agentID,
                "from":      from,
                "to":        to,
            }).Warn("circuit breaker state change")
            
            // Emit metric
            metrics.AgentCircuitState.WithLabelValues(agentID).Set(float64(to))
        },
    })
    
    h.circuitBreakers[agentID] = breaker
    return breaker
}

// Modify HandleExecute
func (h *Handler) HandleExecute(w http.ResponseWriter, r *http.Request) {
    // ... existing code until after authentication
    
    // Check circuit breaker
    breaker := h.getBreaker(agentID)
    if !breaker.Allow() {
        w.Header().Set("Retry-After", "30")
        writeError(w, http.StatusServiceUnavailable, "CIRCUIT_OPEN", "agent execution circuit breaker is open")
        return
    }
    
    // ... existing execution code
    
    // Record result in breaker
    if execErr != nil {
        breaker.Record(execErr)
    } else {
        breaker.Record(nil)
    }
}
```

#### Day 5: Metrics Integration

**File**: `internal/agent/metrics/circuit.go`

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
)

var (
    AgentCircuitState = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "agent_circuit_state",
            Help: "Agent circuit breaker state (0=closed, 1=open, 2=half-open)",
        },
        []string{"agent_id"},
    )
    
    AgentCircuitTransitions = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "agent_circuit_transitions_total",
            Help: "Total circuit breaker state transitions",
        },
        []string{"agent_id", "from", "to"},
    )
)

func init() {
    prometheus.MustRegister(AgentCircuitState)
    prometheus.MustRegister(AgentCircuitTransitions)
}
```

---

### Sprint 1.2: Retry Policies (Week 2)

#### Day 1-2: Retry Policy Implementation

**File**: `internal/agent/retry/policy.go`

```go
package retry

import (
    "context"
    "math"
    "time"
)

type Policy struct {
    MaxRetries      int
    InitialDelay    time.Duration
    MaxDelay        time.Duration
    BackoffFactor   float64
    RetryableErrors []error
}

var DefaultPolicy = Policy{
    MaxRetries:    3,
    InitialDelay:  100 * time.Millisecond,
    MaxDelay:      5 * time.Second,
    BackoffFactor: 2.0,
}

func (p *Policy) Execute(ctx context.Context, fn func() error) error {
    var lastErr error
    
    for attempt := 0; attempt <= p.MaxRetries; attempt++ {
        if attempt > 0 {
            delay := p.calculateDelay(attempt)
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(delay):
            }
        }
        
        lastErr = fn()
        if lastErr == nil {
            return nil
        }
        
        if !p.isRetryable(lastErr) {
            return lastErr
        }
    }
    
    return lastErr
}

func (p *Policy) calculateDelay(attempt int) time.Duration {
    delay := float64(p.InitialDelay) * math.Pow(p.BackoffFactor, float64(attempt-1))
    if delay > float64(p.MaxDelay) {
        delay = float64(p.MaxDelay)
    }
    return time.Duration(delay)
}

func (p *Policy) isRetryable(err error) bool {
    for _, retryableErr := range p.RetryableErrors {
        if err == retryableErr {
            return true
        }
    }
    return false
}
```

#### Day 3-4: Integration with Agent Execution

**File**: `internal/api/handlers/agent/execute.go` (modifications)

```go
// Add retry execution
func (h *Handler) executeWithRetry(ctx context.Context, agentID string, fn func() error) error {
    policy := retry.Policy{
        MaxRetries:    3,
        InitialDelay:  100 * time.Millisecond,
        MaxDelay:      5 * time.Second,
        BackoffFactor: 2.0,
        RetryableErrors: []error{
            ErrTimeout,
            ErrRegistryUnavailable,
            ErrWASMSandboxUnavailable,
        },
    }
    
    return policy.Execute(ctx, fn)
}

// Modify HandleExecute to use retry
func (h *Handler) HandleExecute(w http.ResponseWriter, r *http.Request) {
    // ... existing code
    
    // Execute with retry
    var execResult json.RawMessage
    var execErr error
    
    err := h.executeWithRetry(r.Context(), agentID, func() error {
        execResult, execErr = h.executeViaRegistry(r, author, name, fnVersion.Version, req.Input)
        return execErr
    })
    
    if err != nil {
        // All retries exhausted
        writeError(w, http.StatusInternalServerError, "EXECUTION_FAILED", err.Error())
        return
    }
    
    // ... rest of code
}
```

#### Day 5: Testing

**File**: `internal/agent/retry/policy_test.go`

```go
package retry

import (
    "context"
    "errors"
    "testing"
    "time"
)

func TestPolicy_RetryOnTransientError(t *testing.T) {
    attempts := 0
    policy := Policy{
        MaxRetries:    3,
        InitialDelay:  10 * time.Millisecond,
        MaxDelay:      100 * time.Millisecond,
        BackoffFactor: 2.0,
        RetryableErrors: []error{errors.New("transient")},
    }
    
    err := policy.Execute(context.Background(), func() error {
        attempts++
        if attempts < 3 {
            return errors.New("transient")
        }
        return nil
    })
    
    if err != nil {
        t.Errorf("expected success, got %v", err)
    }
    if attempts != 3 {
        t.Errorf("expected 3 attempts, got %d", attempts)
    }
}
```

---

## Phase 2: Observability & Monitoring

### Sprint 2.1: OpenTelemetry Tracing (Week 3)

#### Day 1-2: Tracer Setup

**File**: `internal/agent/telemetry/tracer.go`

```go
package telemetry

import (
    "context"
    
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func InitTracer(serviceName string) (*sdktrace.TracerProvider, error) {
    ctx := context.Background()
    
    exporter, err := otlptracegrpc.New(ctx)
    if err != nil {
        return nil, err
    }
    
    resource := resource.NewWithAttributes(
        semconv.SchemaURL,
        semconv.ServiceNameKey.String(serviceName),
    )
    
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(resource),
    )
    
    otel.SetTracerProvider(tp)
    return tp, nil
}

func Tracer() trace.Tracer {
    return otel.Tracer("agent")
}
```

#### Day 3-4: Trace Points in Agent Execution

**File**: `internal/api/handlers/agent/execute.go` (modifications)

```go
import "go.opentelemetry.io/otel/attribute"

func (h *Handler) HandleExecute(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    tracer := telemetry.Tracer()
    
    // Start root span
    ctx, span := tracer.Start(ctx, "agent.execute")
    defer span.End()
    
    // Add span attributes
    span.SetAttributes(
        attribute.String("agent.id", agentID),
        attribute.String("agent.tenant_id", tenantID.String()),
        attribute.String("function.uri", functionURI),
        attribute.String("execution.id", executionID),
        attribute.Int("call.depth", req.CallDepth),
    )
    
    // Trace authentication
    ctx, authSpan := tracer.Start(ctx, "agent.authenticate")
    agentID, tenantID, err := h.authenticateAgent(r)
    authSpan.End()
    
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "authentication failed")
        // ... error handling
    }
    
    // Trace quota check
    ctx, quotaSpan := tracer.Start(ctx, "agent.quota.check")
    quotaResult, err := h.quotaEnforcer.CheckAndConsume(ctx, agentID, functionURI, 0)
    quotaSpan.End()
    
    // Trace policy check
    ctx, policySpan := tracer.Start(ctx, "agent.policy.check")
    policyResult, err := h.policyEngine.CheckPolicy(ctx, agentID, policyReq)
    policySpan.End()
    
    // Trace function execution
    ctx, execSpan := tracer.Start(ctx, "agent.function.execute")
    execResult, execErr := h.executeViaRegistry(r, author, name, fnVersion.Version, req.Input)
    execSpan.End()
    
    if execErr != nil {
        span.RecordError(execErr)
        span.SetStatus(codes.Error, "execution failed")
    } else {
        span.SetStatus(codes.Ok, "success")
    }
    
    // Trace attribution
    ctx, attrSpan := tracer.Start(ctx, "agent.attribution.record")
    h.attributionRepo.RecordExecution(ctx, record)
    attrSpan.End()
}
```

#### Day 5: Configuration

**File**: `internal/agent/telemetry/config.go`

```go
package telemetry

type Config struct {
    Enabled     bool
    Endpoint    string
    SampleRate  float64
    ServiceName string
}

func FromEnv() Config {
    return Config{
        Enabled:     os.Getenv("OTEL_ENABLED") == "true",
        Endpoint:    os.Getenv("OTEL_ENDPOINT"),
        SampleRate:  parseFloat(os.Getenv("OTEL_SAMPLE_RATE"), 1.0),
        ServiceName: os.Getenv("OTEL_SERVICE_NAME"),
    }
}
```

---

### Sprint 2.2: Prometheus Metrics (Week 4)

#### Day 1-2: Metrics Definition

**File**: `internal/agent/metrics/metrics.go`

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
)

var (
    // Execution metrics
    AgentExecutions = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "agent_executions_total",
            Help: "Total agent executions",
        },
        []string{"agent_id", "function_uri", "outcome"},
    )
    
    AgentExecutionDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "agent_execution_duration_seconds",
            Help:    "Agent execution duration",
            Buckets: prometheus.ExponentialBuckets(0.01, 2, 15),
        },
        []string{"agent_id", "function_uri"},
    )
    
    AgentExecutionErrors = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "agent_execution_errors_total",
            Help: "Total agent execution errors",
        },
        []string{"agent_id", "function_uri", "error_code"},
    )
    
    // Quota metrics
    AgentQuotaUsage = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "agent_quota_usage_ratio",
            Help: "Agent quota usage ratio (0-1)",
        },
        []string{"agent_id", "quota_type"},
    )
    
    AgentQuotaViolations = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "agent_quota_violations_total",
            Help: "Total quota violations",
        },
        []string{"agent_id", "quota_type"},
    )
    
    // Policy metrics
    AgentPolicyViolations = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "agent_policy_violations_total",
            Help: "Total policy violations",
        },
        []string{"agent_id", "violation_code"},
    )
    
    // Billing metrics
    AgentCostUSD = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "agent_cost_usd_total",
            Help: "Total cost in USD",
        },
        []string{"agent_id", "function_uri"},
    )
    
    // Concurrency metrics
    AgentConcurrencyActive = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "agent_concurrency_active",
            Help: "Active concurrent executions",
        },
        []string{"agent_id"},
    )
    
    AgentConcurrencyLimit = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "agent_concurrency_limit",
            Help: "Concurrency limit",
        },
        []string{"agent_id"},
    )
)

func init() {
    prometheus.MustRegister(
        AgentExecutions,
        AgentExecutionDuration,
        AgentExecutionErrors,
        AgentQuotaUsage,
        AgentQuotaViolations,
        AgentPolicyViolations,
        AgentCostUSD,
        AgentConcurrencyActive,
        AgentConcurrencyLimit,
    )
}
```

#### Day 3-4: Metrics Integration

**File**: `internal/api/handlers/agent/execute.go` (modifications)

```go
func (h *Handler) HandleExecute(w http.ResponseWriter, r *http.Request) {
    startTime := time.Now()
    
    // ... existing code
    
    // Record metrics
    defer func() {
        duration := time.Since(startTime).Seconds()
        
        metrics.AgentExecutions.WithLabelValues(
            agentID,
            functionURI,
            string(outcome),
        ).Inc()
        
        metrics.AgentExecutionDuration.WithLabelValues(
            agentID,
            functionURI,
        ).Observe(duration)
        
        if execErr != nil {
            metrics.AgentExecutionErrors.WithLabelValues(
                agentID,
                functionURI,
                execErr.Error(),
            ).Inc()
        }
        
        metrics.AgentCostUSD.WithLabelValues(
            agentID,
            functionURI,
        ).Add(record.CostUSD)
    }()
    
    // ... rest of code
}
```

#### Day 5: Grafana Dashboard

**File**: `deploy/monitoring/grafana/agent-dashboard.json`

```json
{
  "dashboard": {
    "title": "Agent Execution Dashboard",
    "panels": [
      {
        "title": "Execution Rate",
        "targets": [{
          "expr": "rate(agent_executions_total[5m])"
        }]
      },
      {
        "title": "Error Rate",
        "targets": [{
          "expr": "rate(agent_execution_errors_total[5m]) / rate(agent_executions_total[5m])"
        }]
      },
      {
        "title": "P95 Latency",
        "targets": [{
          "expr": "histogram_quantile(0.95, rate(agent_execution_duration_seconds_bucket[5m]))"
        }]
      },
      {
        "title": "Circuit Breaker States",
        "targets": [{
          "expr": "agent_circuit_state"
        }]
      },
      {
        "title": "Quota Usage",
        "targets": [{
          "expr": "agent_quota_usage_ratio"
        }]
      }
    ]
  }
}
```

---

### Sprint 2.3: Alerting (Week 5)

#### Day 1-2: Alert Rules

**File**: `deploy/monitoring/alerts/agent-alerts.yml`

```yaml
groups:
  - name: agent_alerts
    rules:
      - alert: AgentHighErrorRate
        expr: |
          rate(agent_execution_errors_total[5m]) 
          / rate(agent_executions_total[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Agent {{ $labels.agent_id }} has high error rate"
          description: "Error rate is {{ $value | humanizePercentage }}"
      
      - alert: AgentCircuitOpen
        expr: agent_circuit_state == 1
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Agent {{ $labels.agent_id }} circuit breaker is OPEN"
          description: "Agent is temporarily unavailable"
      
      - alert: AgentQuotaExhausted
        expr: agent_quota_usage_ratio > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Agent {{ $labels.agent_id }} quota nearly exhausted"
          description: "Quota usage is {{ $value | humanizePercentage }}"
      
      - alert: AgentHighLatency
        expr: |
          histogram_quantile(0.95, rate(agent_execution_duration_seconds_bucket[5m])) > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Agent {{ $labels.agent_id }} has high latency"
          description: "P95 latency is {{ $value }}s"
      
      - alert: AgentPolicyViolations
        expr: rate(agent_policy_violations_total[5m]) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Agent {{ $labels.agent_id }} has high policy violations"
          description: "{{ $value }} violations per second"
```

#### Day 3-4: Notification Channels

**File**: `deploy/monitoring/alertmanager/config.yml`

```yaml
global:
  slack_api_url: '${SLACK_WEBHOOK_URL}'

route:
  group_by: ['alertname', 'agent_id']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'slack-notifications'

receivers:
  - name: 'slack-notifications'
    slack_configs:
      - channel: '#agent-alerts'
        title: '{{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}\n{{ end }}'
  
  - name: 'pagerduty-critical'
    pagerduty_configs:
      - service_key: '${PAGERDUTY_KEY}'
        severity: 'critical'
```

#### Day 5: Testing Alerts

**File**: `deploy/monitoring/alerts/test-alerts.sh`

```bash
#!/bin/bash

# Test alert by simulating high error rate
curl -X POST http://localhost:8080/v1/agent/execute/test/test-func \
  -H "X-Agent-API-Key: test-key" \
  -d '{"input": {"simulate_error": true}}'

# Wait for alert to fire
sleep 120

# Check alert status
curl http://localhost:9093/api/v1/alerts | jq '.data.alerts[] | select(.labels.alertname=="AgentHighErrorRate")'
```

---

## Phase 3: Security Hardening

### Sprint 3.1: Capability Scoping (Week 6)

#### Day 1-2: Capability Model

**File**: `internal/agent/capabilities/scope.go`

```go
package capabilities

import (
    "path"
    "strings"
)

type Capabilities struct {
    AgentID           string
    AllowedFunctions  []string // glob patterns: "fx://functionfly/*"
    DeniedFunctions   []string
    AllowedProviders  []string // ["workers", "vercel"]
    MaxCallDepth      int
    MaxConcurrent     int
    AllowedRegions    []string
    MaxExecutionTime  int // seconds
    MaxMemoryMB       int
}

func (c *Capabilities) CanExecute(functionURI string) bool {
    // Check denied first
    for _, pattern := range c.DeniedFunctions {
        if matchPattern(pattern, functionURI) {
            return false
        }
    }
    
    // Check allowed
    if len(c.AllowedFunctions) == 0 {
        return true // no restrictions
    }
    
    for _, pattern := range c.AllowedFunctions {
        if matchPattern(pattern, functionURI) {
            return true
        }
    }
    
    return false
}

func matchPattern(pattern, value string) bool {
    // Support glob patterns
    if strings.Contains(pattern, "*") {
        matched, _ := path.Match(pattern, value)
        return matched
    }
    return pattern == value
}
```

#### Day 3-4: Integration with Policy Engine

**File**: `internal/agent/policy/engine.go` (modifications)

```go
func (e *Engine) CheckPolicy(ctx context.Context, agentID string, req *AgentExecutionRequest) (*PolicyResult, error) {
    // Get capabilities
    caps, err := e.capabilityRepo.GetCapabilities(ctx, agentID)
    if err != nil {
        return nil, err
    }
    
    // Check function permission
    if !caps.CanExecute(req.FunctionURI) {
        return &PolicyResult{
            Allowed: false,
            Violation: &PolicyViolation{
                Code:    "FUNCTION_NOT_ALLOWED",
                Message: "agent does not have permission to execute this function",
            },
        }, nil
    }
    
    // Check call depth
    if req.CallDepth > caps.MaxCallDepth {
        return &PolicyResult{
            Allowed: false,
            Violation: &PolicyViolation{
                Code:    "CALL_DEPTH_EXCEEDED",
                Message: fmt.Sprintf("call depth %d exceeds limit %d", req.CallDepth, caps.MaxCallDepth),
            },
        }, nil
    }
    
    // ... existing policy checks
}
```

#### Day 5: API for Managing Capabilities

**File**: `internal/api/handlers/agent/capabilities.go`

```go
// HandleUpdateCapabilities updates agent capabilities
// PUT /v1/agent/{agent_id}/capabilities
func (h *Handler) HandleUpdateCapabilities(w http.ResponseWriter, r *http.Request) {
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
        return
    }
    
    agentID := mux.Vars(r)["agent_id"]
    agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
    if err != nil {
        writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
        return
    }
    
    if agent.TenantID != claims.TenantID {
        writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
        return
    }
    
    var caps capabilities.Capabilities
    if err := json.NewDecoder(r.Body).Decode(&caps); err != nil {
        writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
        return
    }
    
    caps.AgentID = agentID
    
    if err := h.capabilityRepo.UpsertCapabilities(r.Context(), &caps); err != nil {
        writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", "failed to update capabilities")
        return
    }
    
    writeJSON(w, http.StatusOK, map[string]interface{}{
        "ok":           true,
        "capabilities": caps,
    })
}
```

---

### Sprint 3.2: Input/Output Validation (Week 7)

#### Day 1-2: Validator Implementation

**File**: `internal/agent/validation/validator.go`

```go
package validation

import (
    "encoding/json"
    "regexp"
    "strings"
)

type InputValidator struct {
    MaxSizeBytes    int
    AllowedMimeTypes []string
    PIIPatterns     []*regexp.Regexp
}

type OutputValidator struct {
    MaxSizeBytes    int
    RedactPatterns  []*regexp.Regexp
    PIIPatterns     []*regexp.Regexp
}

var DefaultInputValidator = InputValidator{
    MaxSizeBytes: 1024 * 1024, // 1MB
    PIIPatterns: []*regexp.Regexp{
        regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`), // email
        regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`), // SSN
        regexp.MustCompile(`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`), // credit card
    },
}

var DefaultOutputValidator = OutputValidator{
    MaxSizeBytes: 10 * 1024 * 1024, // 10MB
    RedactPatterns: []*regexp.Regexp{
        regexp.MustCompile(`(?i)password["\s:=]+[^\s"]+`),
        regexp.MustCompile(`(?i)secret["\s:=]+[^\s"]+`),
        regexp.MustCompile(`(?i)token["\s:=]+[^\s"]+`),
    },
}

func (v *InputValidator) Validate(input json.RawMessage) error {
    // Check size
    if len(input) > v.MaxSizeBytes {
        return ErrInputTooLarge
    }
    
    // Check for PII
    inputStr := string(input)
    for _, pattern := range v.PIIPatterns {
        if pattern.MatchString(inputStr) {
            return ErrPIIDetected
        }
    }
    
    return nil
}

func (v *OutputValidator) Validate(output json.RawMessage) (json.RawMessage, error) {
    // Check size
    if len(output) > v.MaxSizeBytes {
        return nil, ErrOutputTooLarge
    }
    
    // Redact sensitive data
    outputStr := string(output)
    for _, pattern := range v.RedactPatterns {
        outputStr = pattern.ReplaceAllString(outputStr, "[REDACTED]")
    }
    
    return json.RawMessage(outputStr), nil
}
```

#### Day 3-4: Integration with Agent Execution

**File**: `internal/api/handlers/agent/execute.go` (modifications)

```go
func (h *Handler) HandleExecute(w http.ResponseWriter, r *http.Request) {
    // ... existing code
    
    // Validate input
    if err := h.inputValidator.Validate(req.Input); err != nil {
        writeError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
        return
    }
    
    // ... execute function
    
    // Validate and sanitize output
    if execResult != nil {
        sanitized, err := h.outputValidator.Validate(execResult)
        if err != nil {
            logrus.WithError(err).Warn("output validation failed")
            // Don't fail execution, just log warning
        } else {
            execResult = sanitized
        }
    }
    
    // ... rest of code
}
```

#### Day 5: Testing

**File**: `internal/agent/validation/validator_test.go`

```go
package validation

import (
    "encoding/json"
    "testing"
)

func TestInputValidator_PIIDetection(t *testing.T) {
    validator := DefaultInputValidator
    
    // Test email detection
    input := json.RawMessage(`{"email": "user@example.com"}`)
    err := validator.Validate(input)
    if err != ErrPIIDetected {
        t.Errorf("expected PII detection, got %v", err)
    }
}

func TestOutputValidator_Redaction(t *testing.T) {
    validator := DefaultOutputValidator
    
    output := json.RawMessage(`{"password": "secret123", "data": "safe"}`)
    sanitized, err := validator.Validate(output)
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
    
    if !strings.Contains(string(sanitized), "[REDACTED]") {
        t.Error("expected redaction")
    }
}
```

---

### Sprint 3.3: Enhanced Audit Logging (Week 8)

#### Day 1-2: Audit Logger

**File**: `internal/agent/audit/logger.go`

```go
package audit

import (
    "context"
    "time"
    
    "github.com/google/uuid"
    "github.com/sirupsen/logrus"
)

type EventType string

const (
    EventExecute         EventType = "execute"
    EventSpawn           EventType = "spawn"
    EventMessage         EventType = "message"
    EventPolicyViolation EventType = "policy_violation"
    EventQuotaViolation  EventType = "quota_violation"
    EventAuthFailure     EventType = "auth_failure"
    EventCapabilityViolation EventType = "capability_violation"
)

type AuditEvent struct {
    ID           uuid.UUID
    Timestamp    time.Time
    EventType    EventType
    AgentID      string
    TenantID     uuid.UUID
    FunctionURI  string
    IPAddress    string
    UserAgent    string
    Outcome      string
    Details      map[string]interface{}
    RiskScore    float64
}

type Logger struct {
    repo   Repository
    logger *logrus.Logger
}

func (l *Logger) Log(ctx context.Context, event *AuditEvent) {
    // Calculate risk score
    event.RiskScore = l.calculateRiskScore(event)
    
    // Log to database
    if err := l.repo.Create(ctx, event); err != nil {
        l.logger.WithError(err).Error("failed to write audit log")
    }
    
    // Log high-risk events to stdout
    if event.RiskScore > 0.7 {
        l.logger.WithFields(logrus.Fields{
            "event_type": event.EventType,
            "agent_id":   event.AgentID,
            "risk_score": event.RiskScore,
            "details":    event.Details,
        }).Warn("high-risk audit event")
    }
}

func (l *Logger) calculateRiskScore(event *AuditEvent) float64 {
    score := 0.0
    
    switch event.EventType {
    case EventPolicyViolation:
        score += 0.5
    case EventQuotaViolation:
        score += 0.3
    case EventAuthFailure:
        score += 0.7
    case EventCapabilityViolation:
        score += 0.6
    }
    
    // Increase score for rapid events
    // (would need to query recent events)
    
    return score
}
```

#### Day 3-4: Integration Points

**File**: `internal/api/handlers/agent/execute.go` (modifications)

```go
func (h *Handler) HandleExecute(w http.ResponseWriter, r *http.Request) {
    // ... existing code
    
    // Log audit event
    h.auditLogger.Log(r.Context(), &audit.AuditEvent{
        ID:          uuid.New(),
        Timestamp:   time.Now(),
        EventType:   audit.EventExecute,
        AgentID:     agentID,
        TenantID:    tenantID,
        FunctionURI: functionURI,
        IPAddress:   r.RemoteAddr,
        UserAgent:   r.UserAgent(),
        Outcome:     string(outcome),
        Details: map[string]interface{}{
            "execution_id": executionID,
            "latency_ms":   latencyMs,
            "cost_usd":     record.CostUSD,
        },
    })
    
    // ... rest of code
}
```

#### Day 5: Query API

**File**: `internal/api/handlers/agent/audit.go`

```go
// HandleListAuditEvents lists audit events for an agent
// GET /v1/agent/{agent_id}/audit
func (h *Handler) HandleListAuditEvents(w http.ResponseWriter, r *http.Request) {
    claims := middleware.GetUserFromContext(r)
    if claims == nil {
        writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
        return
    }
    
    agentID := mux.Vars(r)["agent_id"]
    agent, err := h.identityRepo.GetAgent(r.Context(), agentID)
    if err != nil {
        writeError(w, http.StatusNotFound, "NOT_FOUND", "agent not found")
        return
    }
    
    if agent.TenantID != claims.TenantID {
        writeError(w, http.StatusForbidden, "FORBIDDEN", "access denied")
        return
    }
    
    // Parse filters
    eventType := r.URL.Query().Get("event_type")
    since := r.URL.Query().Get("since")
    limit := 50
    offset := 0
    
    events, total, err := h.auditLogger.repo.List(r.Context(), agentID, eventType, since, limit, offset)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "LIST_FAILED", "failed to list audit events")
        return
    }
    
    writeJSON(w, http.StatusOK, map[string]interface{}{
        "ok":     true,
        "events": events,
        "total":  total,
    })
}
```

---

## Summary Timeline

| Week | Phase | Deliverables |
|------|-------|--------------|
| 1 | Resilience | Circuit breaker implementation + tests |
| 2 | Resilience | Retry policies + integration |
| 3 | Observability | OpenTelemetry tracing |
| 4 | Observability | Prometheus metrics + Grafana dashboard |
| 5 | Observability | Alerting rules + notification channels |
| 6 | Security | Capability scoping model |
| 7 | Security | Input/output validation |
| 8 | Security | Enhanced audit logging |

---

## Success Criteria

### Phase 1 Complete When

- [ ] Circuit breaker prevents cascading failures
- [ ] Retry policies handle transient errors
- [ ] All tests pass
- [ ] Metrics show circuit breaker state

### Phase 2 Complete When

- [ ] 95%+ of executions have traces
- [ ] Grafana dashboard shows real-time metrics
- [ ] Alerts fire correctly on anomalies
- [ ] No false positive alerts in 24h

### Phase 3 Complete When

- [ ] Capability scoping enforced
- [ ] PII detected and blocked in inputs
- [ ] Sensitive data redacted in outputs
- [ ] Audit logs capture all security events
- [ ] Risk scoring identifies suspicious patterns
