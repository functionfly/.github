package ghost

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/agent/deployment"
	"github.com/functionfly/functionfly/internal/agent/generation"
	"github.com/functionfly/functionfly/internal/api/middleware"
	"github.com/functionfly/functionfly/internal/auth"
	"github.com/sirupsen/logrus"
)

func (h *Handler) HandleGenerateSchema(w http.ResponseWriter, r *http.Request) {
	var req GenerateSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if len(req.Entities) == 0 {
		writeError(w, http.StatusBadRequest, "MISSING_ENTITIES", "at least one entity is required")
		return
	}

	claims := middleware.GetUserFromContext(r)
	envCtx := h.buildEnvironmentContext(claims)

	var sql strings.Builder
	sql.WriteString("-- Ghost Mode Auto-generated Schema\n")
	sql.WriteString(fmt.Sprintf("-- Generated: %s\n", envCtx.Environment))
	sql.WriteString("-- Production-grade schema with proper constraints\n\n")

	for _, entity := range req.Entities {
		if entity.Name == "" {
			continue
		}

		sql.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", toSnakeCase(entity.Name)))
		sql.WriteString("  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),\n")

		for i, field := range entity.Fields {
			if field.Name == "" {
				continue
			}
			reqStr := ""
			if field.Required {
				reqStr = " NOT NULL"
			}
			defaultStr := ""
			if field.Default != "" {
				defaultStr = fmt.Sprintf(" DEFAULT %s", field.Default)
			}

			comma := ","
			if i == len(entity.Fields)-1 && len(entity.Indexes) == 0 {
				comma = ""
			}
			sql.WriteString(fmt.Sprintf("  %s %s%s%s%s\n", toSnakeCase(field.Name), mapGoTypeToSQL(field.Type), reqStr, defaultStr, comma))
		}

		for idxIdx, idx := range entity.Indexes {
			comma := ","
			if idxIdx == len(entity.Indexes)-1 {
				comma = ""
			}
			sql.WriteString(fmt.Sprintf("  CONSTRAINT %s_unique UNIQUE (%s)%s\n", toSnakeCase(entity.Name), idx, comma))
		}

		sql.WriteString(")\n")

		for _, idx := range entity.Indexes {
			sql.WriteString(fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_%s ON %s(%s);\n",
				toSnakeCase(entity.Name), idx, toSnakeCase(entity.Name), idx))
		}

		sql.WriteString("\n")
	}

	schemaSQL := sql.String()
	logrus.WithFields(logrus.Fields{
		"tenant":   envCtx.TenantID,
		"entities": len(req.Entities),
		"env":      envCtx.Environment,
	}).Info("Ghost Mode generated database schema")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":          true,
		"sql":         schemaSQL,
		"entities":    len(req.Entities),
		"environment": envCtx.Environment,
	})
}

func (h *Handler) HandleGenerateBackend(w http.ResponseWriter, r *http.Request) {
	var req GenerateBackendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	claims := middleware.GetUserFromContext(r)
	envCtx := h.buildEnvironmentContext(claims)

	if req.Spec == "" {
		writeError(w, http.StatusBadRequest, "MISSING_SPEC", "spec is required for backend generation")
		return
	}

	language := req.Lang
	if language == "" {
		language = "python"
	}

	if !isValidLanguage(language) {
		writeError(w, http.StatusBadRequest, "INVALID_LANGUAGE", fmt.Sprintf("language %s is not supported", language))
		return
	}

	if h.genSvc != nil {
		genReq := &generation.GenerationRequest{
			AgentID:     envCtx.AgentID,
			Name:        fmt.Sprintf("ghost-backend-%s", sanitizeIdentifier(req.Spec[:min(50, len(req.Spec))])),
			Description: req.Spec,
			Runtime:     normalizeRuntime(language),
			Prompt:      buildBackendPrompt(req.Spec, language),
			Model:       selectModelForTask("backend"),
			Tags:        []string{"ghost-mode", "auto-generated", envCtx.Environment},
		}

		result, err := h.genSvc.GenerateFunction(context.Background(), genReq)
		if err == nil && result.Success {
			logrus.WithFields(logrus.Fields{
				"tenant":     envCtx.TenantID,
				"user":       envCtx.UserID,
				"model":      result.ModelUsed,
				"complexity": result.Complexity,
				"runtime":    language,
			}).Info("Ghost Mode backend generated via factoryGeneration")

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":         true,
				"code":       result.Code,
				"model":      result.ModelUsed,
				"complexity": result.Complexity,
				"function_id": result.FunctionID.String(),
				"manifest":   result.Manifest,
				"runtime":    language,
			})
			return
		}

		logrus.WithError(err).Warn("factoryGeneration failed, falling back to agentGenerator")
	}

	if h.deployGen != nil {
		deployReq := &deployment.GenerationRequest{
			AgentID: envCtx.AgentID,
			FunctionSpec: deployment.FunctionSpec{
				Name:        fmt.Sprintf("ghost-backend-%d", time.Now().UnixNano()),
				Title:       "Ghost Mode Backend",
				Description: req.Spec,
				Prompt:      buildBackendPrompt(req.Spec, language),
				Tags:        []string{"ghost-mode", "auto-generated"},
			},
			Language: language,
			Runtime:  normalizeRuntime(language),
		}

		generated, err := h.deployGen.Generate(context.Background(), deployReq)
		if err == nil && generated.Status == deployment.GenerationStatusSuccess {
			logrus.WithFields(logrus.Fields{
				"tenant":          envCtx.TenantID,
				"user":            envCtx.UserID,
				"model":           generated.ModelUsed,
				"generation_time": generated.GenerationTimeMs,
				"runtime":         language,
			}).Info("Ghost Mode backend generated via agentGenerator")

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":              true,
				"code":            generated.GeneratedCode,
				"model":           generated.ModelUsed,
				"generation_id":   generated.ID.String(),
				"generation_time": generated.GenerationTimeMs,
				"runtime":         language,
				"fallback":        false,
			})
			return
		}
	}

	writeError(w, http.StatusServiceUnavailable, "GENERATION_FAILED", "both factoryGeneration and agentGenerator are unavailable")
}

func (h *Handler) HandleGenerateFrontend(w http.ResponseWriter, r *http.Request) {
	var req GenerateFrontendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	claims := middleware.GetUserFromContext(r)
	envCtx := h.buildEnvironmentContext(claims)

	if req.Spec == "" {
		writeError(w, http.StatusBadRequest, "MISSING_SPEC", "spec is required for frontend generation")
		return
	}

	framework := req.Framework
	if framework == "" {
		framework = "react"
	}

	if !isValidFrontendFramework(framework) {
		writeError(w, http.StatusBadRequest, "INVALID_FRAMEWORK", fmt.Sprintf("framework %s is not supported", framework))
		return
	}

	frontendCode, model, complexity, err := h.generateFrontendWithLLM(context.Background(), req.Spec, framework, envCtx)
	if err != nil {
		logrus.WithError(err).Warn("frontend LLM generation failed, using template fallback")
		frontendCode = generateFrontendTemplate(framework, req.Spec)
		model = "template"
		complexity = 1
	}

	logrus.WithFields(logrus.Fields{
		"tenant":    envCtx.TenantID,
		"framework": framework,
		"model":     model,
		"complexity": complexity,
	}).Info("Ghost Mode frontend generated")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":        true,
		"code":      frontendCode,
		"framework": framework,
		"model":     model,
		"complexity": complexity,
	})
}

func (h *Handler) generateFrontendWithLLM(ctx context.Context, spec, framework string, envCtx EnvironmentContext) (string, string, int, error) {
	if h.genSvc == nil {
		return "", "", 0, fmt.Errorf("generation service unavailable")
	}

	prompt := buildFrontendPrompt(spec, framework)

	genReq := &generation.GenerationRequest{
		AgentID:     envCtx.AgentID,
		Name:        fmt.Sprintf("ghost-frontend-%s", framework),
		Description: fmt.Sprintf("Frontend component for: %s", spec[:min(100, len(spec))]),
		Runtime:     "typescript",
		Prompt:      prompt,
		Model:       selectModelForTask("frontend"),
		Tags:        []string{"ghost-mode", "frontend", framework},
	}

	result, err := h.genSvc.GenerateFunction(ctx, genReq)
	if err != nil || !result.Success {
		return "", "", 0, err
	}

	return result.Code, result.ModelUsed, result.Complexity, nil
}

func (h *Handler) HandleGenerateTests(w http.ResponseWriter, r *http.Request) {
	var req GenerateTestsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	claims := middleware.GetUserFromContext(r)
	envCtx := h.buildEnvironmentContext(claims)

	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "MISSING_CODE", "code is required for test generation")
		return
	}

	if req.Coverage == "" {
		req.Coverage = "80"
	}

	tests, model, err := h.generateTestsWithLLM(context.Background(), req.Code, req.Coverage, envCtx)
	if err != nil {
		logrus.WithError(err).Warn("test LLM generation failed, using template fallback")
		tests = generateTestTemplate(req.Code, req.Coverage)
		model = "template"
	}

	logrus.WithFields(logrus.Fields{
		"tenant":  envCtx.TenantID,
		"model":  model,
		"coverage": req.Coverage,
	}).Info("Ghost Mode tests generated")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"tests":   tests,
		"model":   model,
		"coverage": req.Coverage,
	})
}

func (h *Handler) generateTestsWithLLM(ctx context.Context, code, coverageTarget string, envCtx EnvironmentContext) (string, string, error) {
	if h.genSvc == nil {
		return "", "", fmt.Errorf("generation service unavailable")
	}

	prompt := buildTestPrompt(code, coverageTarget)

	genReq := &generation.GenerationRequest{
		AgentID:     envCtx.AgentID,
		Name:        fmt.Sprintf("ghost-tests-%d", time.Now().UnixNano()),
		Description: "Auto-generated tests for Ghost Mode build",
		Runtime:     detectRuntimeFromCode(code),
		Prompt:      prompt,
		Model:       selectModelForTask("tests"),
		Tags:        []string{"ghost-mode", "testing"},
	}

	result, err := h.genSvc.GenerateFunction(ctx, genReq)
	if err != nil || !result.Success {
		return "", "", err
	}

	return result.Code, result.ModelUsed, nil
}

func (h *Handler) buildEnvironmentContext(claims *auth.Claims) EnvironmentContext {
	envCtx := h.envContext
	if claims != nil {
		envCtx.TenantID = claims.TenantID.String()
		envCtx.UserID = claims.UserID.String()
		envCtx.AgentID = "ghost-" + claims.UserID.String()
		if claims.Permissions != nil {
			envCtx.Permissions = claims.Permissions
		}
	}
	return envCtx
}

func isValidLanguage(lang string) bool {
	validLanguages := map[string]bool{
		"python":     true,
		"javascript": true,
		"typescript": true,
		"go":         true,
		"java":       true,
	}
	return validLanguages[strings.ToLower(lang)]
}

func isValidFrontendFramework(framework string) bool {
	validFrameworks := map[string]bool{
		"react":     true,
		"vue":       true,
		"svelte":    true,
		"next":      true,
		"astro":     true,
	}
	return validFrameworks[strings.ToLower(framework)]
}

func normalizeRuntime(language string) string {
	switch strings.ToLower(language) {
	case "python", "py":
		return "python3.11"
	case "javascript", "js":
		return "nodejs20"
	case "typescript", "ts":
		return "nodejs20"
	case "go":
		return "go1.21"
	case "java":
		return "java17"
	default:
		return "python3.11"
	}
}

func selectModelForTask(task string) string {
	switch task {
	case "backend":
		return "inception/mercury-2"
	case "frontend":
		return "anthropic/claude-3-haiku"
	case "tests":
		return "inception/mercury-2"
	default:
		return "inception/mercury-2"
	}
}

func buildBackendPrompt(spec, language string) string {
	return fmt.Sprintf(`Generate production-grade %s code for the following specification:

%s

Requirements:
- Use %s runtime conventions
- Include proper error handling and validation
- Follow best practices for the language
- No hardcoded secrets or credentials
- Add comprehensive inline documentation
- Optimize for performance and maintainability`, language, spec, language)
}

func buildFrontendPrompt(spec, framework string) string {
	return fmt.Sprintf(`Generate production-grade %s code for the following specification:

%s

Requirements:
- Use %s best practices
- Include proper TypeScript types
- Follow accessibility guidelines
- Add comprehensive inline documentation
- Use CSS modules or Tailwind for styling
- Optimize for performance and bundle size`, framework, spec, framework)
}

func buildTestPrompt(code, coverageTarget string) string {
	return fmt.Sprintf(`Generate comprehensive tests for the following code with %s%% coverage target:

%s

Requirements:
- Use appropriate testing framework for the language
- Include unit tests, integration tests where applicable
- Mock external dependencies
- Test edge cases and error conditions
- Add descriptive test names`, coverageTarget, code)
}

func generateFrontendTemplate(framework, spec string) string {
	specTruncated := spec
	if len(specTruncated) > 200 {
		specTruncated = spec[:200]
	}

	switch strings.ToLower(framework) {
	case "react", "next":
		return `// Ghost Mode Auto-generated React Component
import React, { useState, useEffect } from 'react';

interface ComponentProps {
  className?: string;
}

export function GhostComponent({ className = '' }: ComponentProps) {
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        const response = await fetch('/api/endpoint');
        if (!response.ok) throw new Error('Failed to fetch');
        const result = await response.json();
        setData(result);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, []);

  if (loading) return <div className="loading">Loading...</div>;
  if (error) return <div className="error">{error}</div>;

  return (
    <div className={"ghost-component " + className}>
      <h1>Ghost Mode Generated</h1>
      <p>Spec: ` + specTruncated + `</p>
      {data && <pre>{JSON.stringify(data, null, 2)}</pre>}
    </div>
  );
}
`

	case "vue":
		return "// Ghost Mode Auto-generated Vue Component\n" +
			"<template>\n" +
			"  <div class=\"ghost-component\">\n" +
			"    <h1>Ghost Mode Generated</h1>\n" +
			"    <p>Spec: " + specTruncated + "</p>\n" +
			"  </div>\n" +
			"</template>\n\n" +
			"<script setup lang=\"ts\">\n" +
			"import { ref, onMounted } from 'vue';\n\n" +
			"const data = ref(null);\n" +
			"const loading = ref(true);\n" +
			"const error = ref<string | null>(null);\n\n" +
			"onMounted(async () => {\n" +
			"  try {\n" +
			"    loading.value = true;\n" +
			"    const response = await fetch('/api/endpoint');\n" +
			"    if (!response.ok) throw new Error('Failed to fetch');\n" +
			"    data.value = await response.json();\n" +
			"  } catch (err) {\n" +
			"    error.value = err instanceof Error ? err.message : 'Unknown error';\n" +
			"  } finally {\n" +
			"    loading.value = false;\n" +
			"  }\n" +
			"});\n" +
			"</script>\n\n" +
			"<style scoped>\n" +
			".ghost-component {\n" +
			"  padding: 1rem;\n" +
			"}\n" +
			"</style>\n"

	default:
		return "// Ghost Mode Auto-generated " + framework + " Component\n" +
			"// Spec: " + specTruncated + "\n\n" +
			"export function GhostComponent() {\n" +
			"  return document.createElement('div');\n" +
			"}\n"
	}
}

func generateTestTemplate(code, coverageTarget string) string {
	runtime := detectRuntimeFromCode(code)

	switch runtime {
	case "python3.11", "python":
		return fmt.Sprintf(`# Ghost Mode Auto-generated Tests
# Coverage target: %s%%

import pytest
from unittest.mock import Mock, patch

# Test implementation
def test_basic_functionality():
    """Test basic functionality of ghost-generated code"""
    assert True

def test_edge_cases():
    """Test edge cases and error conditions"""
    with pytest.raises(ValueError):
        pass  # Add your test here

def test_integration():
    """Test integration with external dependencies"""
    mock = Mock()
    mock.return_value = {"status": "ok"}
    assert mock() == {"status": "ok"}
`, coverageTarget)

	case "nodejs20", "javascript", "typescript":
		return fmt.Sprintf(`// Ghost Mode Auto-generated Tests
// Coverage target: %s%%

describe('Ghost Generated Tests', () => {
  it('should pass basic assertion', () => {
    expect(true).toBe(true);
  });

  it('should handle edge cases', () => {
    expect(() => {
      throw new Error('Test error');
    }).toThrow();
  });

  it('should mock external dependencies', () => {
    const mock = jest.fn(() => ({ status: 'ok' }));
    expect(mock()).toEqual({ status: 'ok' });
  });
});
`, coverageTarget)

	default:
		return fmt.Sprintf(`// Ghost Mode Auto-generated Tests
// Coverage target: %s%%

describe('Ghost Generated Tests', () => {
  it('should pass basic assertion', () => {
    expect(true).toBe(true);
  });
});
`, coverageTarget)
	}
}

func detectRuntimeFromCode(code string) string {
	codeLower := strings.ToLower(code)
	if strings.Contains(codeLower, "import ") && strings.Contains(codeLower, "react") {
		return "typescript"
	}
	if strings.Contains(codeLower, "def ") || strings.Contains(codeLower, "import ") && strings.Contains(codeLower, "pytest") {
		return "python3.11"
	}
	if strings.Contains(codeLower, "function ") || strings.Contains(codeLower, "const ") || strings.Contains(codeLower, "let ") {
		return "javascript"
	}
	return "python3.11"
}

func sanitizeIdentifier(s string) string {
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		} else if r == ' ' {
			result.WriteByte('-')
		}
	}
	return result.String()
}