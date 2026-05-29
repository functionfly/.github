package ghost

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/functionfly/functionfly/internal/agent/generation"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/sirupsen/logrus"
)

func (h *Handler) HandlePlanArchitecture(w http.ResponseWriter, r *http.Request) {
	var req PlanArchitectureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.Goal == "" {
		writeError(w, http.StatusBadRequest, "MISSING_GOAL", "goal is required for architecture planning")
		return
	}

	claims := middleware.GetUserFromContext(r)
	envCtx := h.buildEnvironmentContext(claims)

	if h.genSvc != nil {
		genReq := &generation.GenerationRequest{
			AgentID:     envCtx.AgentID,
			Name:        fmt.Sprintf("ghost-architecture-%d", time.Now().UnixNano()),
			Description: fmt.Sprintf("Architecture plan for: %s", req.Goal),
			Runtime:     "go",
			Prompt:      buildArchitecturePrompt(req.Goal, req.Domain),
			Model:       "anthropic/claude-3-opus",
			Tags:        []string{"ghost-mode", "architecture", envCtx.Environment},
		}

		result, err := h.genSvc.GenerateFunction(context.Background(), genReq)
		if err == nil && result.Success {
			plan := parseArchitectureFromCode(result.Code, req.Goal, req.Domain)
			if plan != nil {
				logrus.WithFields(logrus.Fields{
					"tenant": envCtx.TenantID,
					"goal":   req.Goal,
					"model":  result.ModelUsed,
				}).Info("Ghost Mode architecture generated via LLM")

				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":    true,
					"plan":  plan,
					"model": result.ModelUsed,
					"llm":   true,
				})
				return
			}
		}
	}

	plan := h.generateArchitecturePlan(req.Goal, req.Domain)

	logrus.WithFields(logrus.Fields{
		"tenant":  envCtx.TenantID,
		"goal":   req.Goal,
		"domain": req.Domain,
	}).Info("Ghost Mode architecture generated via template")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":    true,
		"plan":  plan,
		"llm":   false,
	})
}

func (h *Handler) generateArchitecturePlan(goal, domain string) *ArchitecturePlan {
	domainLower := ""
	if domain != "" {
		domainLower = domain
	}

	var components []ComponentSpec
	var entities []EntitySpec
	var endpoints []EndpointSpec

	switch {
	case containsKeyword(goal, "e-commerce", "shop", "store", "cart", "payment"):
		components = []ComponentSpec{
			{Name: "api-gateway", Type: "api", Description: "Main API gateway with auth and rate limiting", Technology: "Go"},
			{Name: "product-service", Type: "api", Description: "Product catalog and inventory management", Technology: "Go"},
			{Name: "order-service", Type: "api", Description: "Order processing and fulfillment", Technology: "Go"},
			{Name: "payment-service", Type: "api", Description: "Payment processing via Stripe", Technology: "Go"},
			{Name: "postgres-primary", Type: "database", Description: "Primary PostgreSQL 17 database for transactions", Technology: "PostgreSQL 17"},
			{Name: "redis-sessions", Type: "cache", Description: "Redis 7 for session and cart caching", Technology: "Redis 7"},
			{Name: "email-worker", Type: "worker", Description: "Background email processing", Technology: "Go + BullMQ"},
		}
		entities = []EntitySpec{
			{Name: "products", Fields: []FieldSpec{
				{Name: "id", Type: "uuid", Required: true},
				{Name: "name", Type: "string", Required: true},
				{Name: "price", Type: "decimal", Required: true},
				{Name: "inventory", Type: "int", Required: true},
				{Name: "category", Type: "string", Required: false},
				{Name: "created_at", Type: "timestamp", Required: true},
			}, Indexes: []string{"category", "created_at"}},
			{Name: "orders", Fields: []FieldSpec{
				{Name: "id", Type: "uuid", Required: true},
				{Name: "user_id", Type: "uuid", Required: true},
				{Name: "total", Type: "decimal", Required: true},
				{Name: "status", Type: "string", Required: true},
				{Name: "created_at", Type: "timestamp", Required: true},
			}, Indexes: []string{"user_id", "status"}},
			{Name: "payments", Fields: []FieldSpec{
				{Name: "id", Type: "uuid", Required: true},
				{Name: "order_id", Type: "uuid", Required: true},
				{Name: "amount", Type: "decimal", Required: true},
				{Name: "stripe_payment_id", Type: "string", Required: false},
				{Name: "status", Type: "string", Required: true},
				{Name: "created_at", Type: "timestamp", Required: true},
			}, Indexes: []string{"order_id", "status"}},
		}
		endpoints = []EndpointSpec{
			{Method: "GET", Path: "/products", Handler: "ListProducts", Auth: false},
			{Method: "GET", Path: "/products/{id}", Handler: "GetProduct", Auth: false},
			{Method: "POST", Path: "/orders", Handler: "CreateOrder", Auth: true},
			{Method: "GET", Path: "/orders/{id}", Handler: "GetOrder", Auth: true},
			{Method: "POST", Path: "/payments", Handler: "CreatePayment", Auth: true},
		}

	case containsKeyword(goal, "saas", "dashboard", "analytics", "metrics"):
		components = []ComponentSpec{
			{Name: "api-gateway", Type: "api", Description: "Main API gateway with auth and rate limiting", Technology: "Go"},
			{Name: "auth-service", Type: "api", Description: "Authentication and user management", Technology: "Go"},
			{Name: "analytics-service", Type: "api", Description: "Analytics data processing", Technology: "Go"},
			{Name: "postgres-primary", Type: "database", Description: "Primary PostgreSQL 17 database", Technology: "PostgreSQL 17"},
			{Name: "redis-cache", Type: "cache", Description: "Redis 7 for session caching", Technology: "Redis 7"},
			{Name: "cron-worker", Type: "worker", Description: "Scheduled report generation", Technology: "Go"},
		}
		entities = []EntitySpec{
			{Name: "users", Fields: []FieldSpec{
				{Name: "id", Type: "uuid", Required: true},
				{Name: "email", Type: "string", Required: true},
				{Name: "password_hash", Type: "string", Required: true},
				{Name: "plan", Type: "string", Required: true},
				{Name: "created_at", Type: "timestamp", Required: true},
			}, Indexes: []string{"email"}},
			{Name: "events", Fields: []FieldSpec{
				{Name: "id", Type: "uuid", Required: true},
				{Name: "user_id", Type: "uuid", Required: true},
				{Name: "event_type", Type: "string", Required: true},
				{Name: "properties", Type: "jsonb", Required: false},
				{Name: "timestamp", Type: "timestamp", Required: true},
			}, Indexes: []string{"user_id", "event_type", "timestamp"}},
			{Name: "reports", Fields: []FieldSpec{
				{Name: "id", Type: "uuid", Required: true},
				{Name: "user_id", Type: "uuid", Required: true},
				{Name: "type", Type: "string", Required: true},
				{Name: "data", Type: "jsonb", Required: true},
				{Name: "created_at", Type: "timestamp", Required: true},
			}, Indexes: []string{"user_id", "created_at"}},
		}
		endpoints = []EndpointSpec{
			{Method: "POST", Path: "/auth/register", Handler: "Register", Auth: false},
			{Method: "POST", Path: "/auth/login", Handler: "Login", Auth: false},
			{Method: "GET", Path: "/analytics/events", Handler: "ListEvents", Auth: true},
			{Method: "POST", Path: "/analytics/events", Handler: "TrackEvent", Auth: true},
			{Method: "GET", Path: "/reports", Handler: "ListReports", Auth: true},
		}

	case containsKeyword(goal, "api", "rest", "backend", "service"):
		components = []ComponentSpec{
			{Name: "api-gateway", Type: "api", Description: "Main API gateway with auth and rate limiting", Technology: "Go"},
			{Name: "postgres-primary", Type: "database", Description: "Primary PostgreSQL 17 database", Technology: "PostgreSQL 17"},
			{Name: "redis-cache", Type: "cache", Description: "Redis 7 for response caching", Technology: "Redis 7"},
		}
		entities = []EntitySpec{
			{Name: "resources", Fields: []FieldSpec{
				{Name: "id", Type: "uuid", Required: true},
				{Name: "name", Type: "string", Required: true},
				{Name: "data", Type: "jsonb", Required: false},
				{Name: "created_at", Type: "timestamp", Required: true},
				{Name: "updated_at", Type: "timestamp", Required: true},
			}, Indexes: []string{"name", "created_at"}},
		}
		endpoints = []EndpointSpec{
			{Method: "GET", Path: "/resources", Handler: "ListResources", Auth: false},
			{Method: "POST", Path: "/resources", Handler: "CreateResource", Auth: true},
			{Method: "GET", Path: "/resources/{id}", Handler: "GetResource", Auth: false},
			{Method: "PUT", Path: "/resources/{id}", Handler: "UpdateResource", Auth: true},
			{Method: "DELETE", Path: "/resources/{id}", Handler: "DeleteResource", Auth: true},
		}

	default:
		components = []ComponentSpec{
			{Name: "api-gateway", Type: "api", Description: "Main API gateway with auth and rate limiting", Technology: "Go"},
			{Name: "user-service", Type: "api", Description: "User management and authentication", Technology: "Go"},
			{Name: "postgres-primary", Type: "database", Description: "Primary PostgreSQL 17 database", Technology: "PostgreSQL 17"},
			{Name: "redis-cache", Type: "cache", Description: "Redis for session caching", Technology: "Redis 7"},
			{Name: "background-workers", Type: "worker", Description: "Background job processing", Technology: "Go + BullMQ"},
		}
		entities = []EntitySpec{
			{Name: "users", Fields: []FieldSpec{
				{Name: "id", Type: "uuid", Required: true},
				{Name: "email", Type: "string", Required: true},
				{Name: "password_hash", Type: "string", Required: true},
				{Name: "created_at", Type: "timestamp", Required: true},
				{Name: "updated_at", Type: "timestamp", Required: true},
			}, Indexes: []string{"email", "created_at"}},
		}
		endpoints = []EndpointSpec{
			{Method: "POST", Path: "/auth/register", Handler: "RegisterUser", Auth: false},
			{Method: "POST", Path: "/auth/login", Handler: "LoginUser", Auth: false},
			{Method: "GET", Path: "/users/me", Handler: "GetCurrentUser", Auth: true},
			{Method: "PUT", Path: "/users/me", Handler: "UpdateCurrentUser", Auth: true},
		}
	}

	return &ArchitecturePlan{
		Components:    components,
		DataModel:     entities,
		APIDesign:     endpoints,
		TechStack:     detectTechStack(goal, domainLower),
		Dependencies:  detectDependencies(goal, domainLower),
		EstimatedCost: estimateCost(components),
		RiskFactors:   detectRiskFactors(components),
	}
}

func containsKeyword(goal string, keywords ...string) bool {
	goalLower := normalizeString(goal)
	for _, kw := range keywords {
		if len(goalLower) > 0 && (contains(goalLower, kw) || contains(kw, goalLower)) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (len(s) >= len(substr)) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func normalizeString(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == ' ' || c == '-' || c == '_' {
			result = append(result, c)
		}
	}
	return string(result)
}

func detectTechStack(goal, domain string) []string {
	if containsKeyword(goal, "react", "frontend", "ui", "dashboard") {
		return []string{"Go", "PostgreSQL 17", "Redis 7", "React 19", "TailwindCSS", "TypeScript"}
	}
	if containsKeyword(goal, "mobile", "ios", "android") {
		return []string{"Go", "PostgreSQL 17", "Redis 7", "React Native", "Expo"}
	}
	return []string{"Go", "PostgreSQL 17", "Redis 7", "React 19", "TailwindCSS"}
}

func detectDependencies(goal, domain string) []string {
	deps := []string{
		"github.com/gin-gonic/gin",
		"github.com/lib/pq",
		"github.com/redis/go-redis/v9",
		"github.com/jmoiron/sqlx",
	}

	if containsKeyword(goal, "auth", "login", "user", "register") {
		deps = append(deps, "golang.org/x/crypto/bcrypt")
	}
	if containsKeyword(goal, "email", "notification", "sms") {
		deps = append(deps, "github.com/resend/resend-go")
	}
	if containsKeyword(goal, "payment", "stripe", "billing") {
		deps = append(deps, "github.com/stripe/stripe-go/v76")
	}
	if containsKeyword(goal, "file", "upload", "storage", "s3") {
		deps = append(deps, "github.com/aws/aws-sdk-go-v2")
	}

	return deps
}

func estimateCost(components []ComponentSpec) string {
	computeComponents := 0

	for _, c := range components {
		if c.Type == "worker" || c.Type == "api" {
			computeComponents++
		}
	}

	if computeComponents <= 2 {
		return "$25-75/month"
	}
	if computeComponents <= 4 {
		return "$75-200/month"
	}
	return "$200-500/month"
}

func detectRiskFactors(components []ComponentSpec) []string {
	var risks []string

	hasWorkers := false
	hasCache := false
	hasMultipleAPIs := false

	for _, c := range components {
		if c.Type == "worker" {
			hasWorkers = true
		}
		if c.Type == "cache" {
			hasCache = true
		}
		if c.Type == "api" {
			hasMultipleAPIs = true
		}
	}

	if hasWorkers {
		risks = append(risks, "Job queue backpressure during traffic spikes")
	}
	if hasCache {
		risks = append(risks, "Cache invalidation complexity")
	}
	if hasMultipleAPIs {
		risks = append(risks, "Service-to-service authentication and authorization")
	}

	if len(risks) == 0 {
		risks = []string{"Database connection pooling at scale"}
	}

	return risks
}

func parseArchitectureFromCode(code, goal, domain string) *ArchitecturePlan {
	return nil
}

func buildArchitecturePrompt(goal, domain string) string {
	prompt := fmt.Sprintf(`You are an expert software architect. Generate a comprehensive architecture plan for the following goal:

Goal: %s

`, goal)

	if domain != "" {
		prompt += fmt.Sprintf("Domain: %s\n", domain)
	}

	prompt += `
Provide a complete system design including:

1. **Components**: List all services, APIs, workers, and infrastructure needed
2. **Data Model**: Define all database entities with fields, types, and relationships
3. **API Design**: Specify all REST endpoints with methods, paths, authentication
4. **Tech Stack**: Recommend appropriate technologies
5. **Dependencies**: List required Go modules/packages
6. **Estimated Cost**: Provide a rough monthly cost estimate
7. **Risk Factors**: Identify potential technical risks and mitigation strategies

Format your response as a structured architecture plan. Focus on production-ready, scalable design.
`

	return prompt
}