package support

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ContextCollector gathers contextual information for support conversations
type ContextCollector struct {
	functionGetter FunctionGetter
	logGetter     LogGetter
	deployGetter  DeployGetter
	userGetter    UserGetter
}

// FunctionGetter interface for fetching function information
type FunctionGetter interface {
	GetFunction(ctx context.Context, author, name, version string) (*FunctionInfo, error)
	GetFunctionCode(ctx context.Context, functionID uuid.UUID) (string, error)
	GetFunctionLogs(ctx context.Context, functionID uuid.UUID, limit int) ([]string, error)
	GetRecentDeployments(ctx context.Context, functionID uuid.UUID, limit int) ([]*DeploymentInfo, error)
}

// LogGetter interface for fetching logs
type LogGetter interface {
	GetExecutionLogs(ctx context.Context, executionID uuid.UUID) ([]string, error)
}

// DeployGetter interface for deployment information
type DeployGetter interface {
	GetDeploymentStatus(ctx context.Context, deploymentID uuid.UUID) (*DeploymentStatus, error)
}

// UserGetter interface for user information
type UserGetter interface {
	GetUser(ctx context.Context, userID uuid.UUID) (*UserInfo, error)
}

// FunctionInfo represents function metadata
type FunctionInfo struct {
	ID          uuid.UUID
	Author      string
	Name        string
	Version     string
	Runtime     string
	MemoryLimit int
	Timeout     int
	Status      string
	CreatedAt   string
	UpdatedAt   string
}

// DeploymentInfo represents deployment information
type DeploymentInfo struct {
	ID        uuid.UUID
	Status    string
	Region    string
	CreatedAt string
	Error     string
}

// DeploymentStatus represents current deployment status
type DeploymentStatus struct {
	ID            uuid.UUID
	Status        string
	Endpoint      string
	ActiveVersion string
	Failures      int
	LastDeployAt  string
}

// UserInfo represents user information
type UserInfo struct {
	ID        uuid.UUID
	Email     string
	Plan      string
	OrgID     uuid.UUID
	OrgName   string
	CreatedAt string
}

// NewContextCollector creates a new context collector
func NewContextCollector(
	functionGetter FunctionGetter,
	logGetter LogGetter,
	deployGetter DeployGetter,
	userGetter UserGetter,
) *ContextCollector {
	return &ContextCollector{
		functionGetter: functionGetter,
		logGetter:     logGetter,
		deployGetter:  deployGetter,
		userGetter:    userGetter,
	}
}

// CollectContext gathers all relevant context for a support conversation
func (c *ContextCollector) CollectContext(ctx context.Context, userID uuid.UUID, ref *FunctionRef, deploymentID *uuid.UUID) (*SupportContext, error) {
	supportCtx := &SupportContext{
		UserInfo: &UserContext{},
	}

	// Get user info
	if c.userGetter != nil && userID != uuid.Nil {
		userInfo, err := c.userGetter.GetUser(ctx, userID)
		if err == nil && userInfo != nil {
			createdAt, _ := time.Parse("2006-01-02T15:04:05Z", userInfo.CreatedAt)
			supportCtx.UserInfo = &UserContext{
				ID:        userInfo.ID,
				Email:     userInfo.Email,
				Plan:     userInfo.Plan,
				OrgID:     userInfo.OrgID,
				CreatedAt: createdAt,
			}
		}
	}

	// Get function info if provided
	if ref != nil && c.functionGetter != nil {
		fn, err := c.functionGetter.GetFunction(ctx, ref.Author, ref.Name, ref.Version)
		if err == nil && fn != nil {
			// Get function logs
			logs, err := c.functionGetter.GetFunctionLogs(ctx, fn.ID, 50)
			if err == nil {
				supportCtx.FunctionLogs = logs
			}

			// Get recent deployments
			deployments, err := c.functionGetter.GetRecentDeployments(ctx, fn.ID, 5)
			if err == nil && len(deployments) > 0 {
				// Check for failed deployments
				for _, d := range deployments {
					if d.Status == "failed" && d.Error != "" {
						supportCtx.DeploymentError = d.Error
						break
					}
				}
			}
		}
	}

	// Get deployment-specific info
	if deploymentID != nil && c.deployGetter != nil {
		status, err := c.deployGetter.GetDeploymentStatus(ctx, *deploymentID)
		if err == nil && status != nil {
			if status.Failures > 0 {
				supportCtx.DeploymentError = fmt.Sprintf("Deployment has %d failures. Last deployed: %s", status.Failures, status.LastDeployAt)
			}
		}
	}

	return supportCtx, nil
}

// CollectFunctionCode retrieves and optionally masks sensitive code
func (c *ContextCollector) CollectFunctionCode(ctx context.Context, functionID uuid.UUID) (string, error) {
	if c.functionGetter == nil {
		return "", nil
	}

	code, err := c.functionGetter.GetFunctionCode(ctx, functionID)
	if err != nil {
		return "", err
	}

	// Basic masking of potential secrets
	code = maskSecrets(code)

	return code, nil
}

// maskSecrets masks potential secrets in code
func maskSecrets(code string) string {
	// Common patterns for secrets - simple string replacement
	replacements := []struct {
		from string
		to   string
	}{
		{`"api_key": "`, `"api_key": "[REDACTED]"`},
		{`"password": "`, `"password": "[REDACTED]"`},
		{`"secret": "`, `"secret": "[REDACTED]"`},
		{`"token": "`, `"token": "[REDACTED]"`},
		{`'api_key': '`, `'api_key': '[REDACTED]'`},
		{`'password': '`, `'password': '[REDACTED]'`},
		{`'secret': '`, `'secret': '[REDACTED]'`},
		{`'token': '`, `'token': '[REDACTED]'`},
	}

	result := code
	for _, r := range replacements {
		result = strings.ReplaceAll(result, r.from, r.to)
	}

	return result
}

// FormatContextForAI formats context into a prompt-friendly string
func FormatContextForAI(ctx *SupportContext) string {
	var builder strings.Builder

	builder.WriteString("## Support Context\n\n")

	if ctx.UserInfo != nil {
		builder.WriteString("### User\n")
		builder.WriteString(fmt.Sprintf("- User ID: %s\n", ctx.UserInfo.ID))
		builder.WriteString(fmt.Sprintf("- Email: %s\n", ctx.UserInfo.Email))
		builder.WriteString(fmt.Sprintf("- Plan: %s\n", ctx.UserInfo.Plan))
		builder.WriteString("\n")
	}

	if ctx.FunctionLogs != nil && len(ctx.FunctionLogs) > 0 {
		builder.WriteString("### Recent Logs\n")
		for i, log := range ctx.FunctionLogs {
			if i >= 20 { // Limit to last 20 logs
				builder.WriteString(fmt.Sprintf("... and %d more\n", len(ctx.FunctionLogs)-20))
				break
			}
			builder.WriteString(fmt.Sprintf("  %s\n", log))
		}
		builder.WriteString("\n")
	}

	if ctx.DeploymentError != "" {
		builder.WriteString("### Deployment Error\n")
		builder.WriteString(ctx.DeploymentError)
		builder.WriteString("\n\n")
	}

	if ctx.EnvironmentVars != nil {
		builder.WriteString("### Environment Variables\n")
		for key, value := range ctx.EnvironmentVars {
			// Mask sensitive values
			if isSensitiveKey(key) {
				value = "[REDACTED]"
			}
			builder.WriteString(fmt.Sprintf("- %s: %s\n", key, value))
		}
		builder.WriteString("\n")
	}

	if ctx.ExecutionHistory != nil && len(ctx.ExecutionHistory) > 0 {
		builder.WriteString("### Recent Executions\n")
		for _, exec := range ctx.ExecutionHistory {
			status := "SUCCESS"
			if !exec.Success {
				status = "FAILED"
			}
			builder.WriteString(fmt.Sprintf("  [%s] %s - %s", status, exec.Timestamp.Format(time.RFC3339), exec.ID))
			if exec.Error != "" {
				builder.WriteString(fmt.Sprintf(" - Error: %s", exec.Error))
			}
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

// isSensitiveKey checks if an environment variable key is sensitive
func isSensitiveKey(key string) bool {
	sensitive := []string{
		"api_key", "apikey", "api-key",
		"secret", "password", "passwd",
		"token", "auth", "credential",
		"private", "key", "jwt",
	}

	lower := strings.ToLower(key)
	for _, s := range sensitive {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}
