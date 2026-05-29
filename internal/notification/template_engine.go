package notification

import (
	"bytes"
	"context"
	"fmt"
	htmltemplate "html/template"
	"sort"
	"strings"
	texttemplate "text/template"
)

// TemplateEngine handles template rendering
type TemplateEngine struct {
	repo Repository
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine(repo Repository) *TemplateEngine {
	return &TemplateEngine{repo: repo}
}

// Render renders a template with data using Go's text/template and html/template.
// Templates may use {{.Key}} (e.g. {{.UserName}}) or legacy {{Key}}; both are supported.
// HTML body is executed with html/template so injected values are escaped for safety.
func (e *TemplateEngine) Render(ctx context.Context, notificationType, channel string, data JSONMap) (subject, bodyHTML, bodyText string, err error) {
	tmpl, err := e.repo.GetTemplate(ctx, notificationType, channel)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get template: %w", err)
	}
	if tmpl == nil {
		return "", "", "", nil
	}

	subject = normalizeTemplatePlaceholders(tmpl.Subject, data)
	bodyHTML = normalizeTemplatePlaceholders(tmpl.BodyHTML, data)
	bodyText = normalizeTemplatePlaceholders(tmpl.BodyText, data)

	subject, err = executeTextTemplate("subject", subject, data)
	if err != nil {
		return "", "", "", fmt.Errorf("subject template: %w", err)
	}
	bodyText, err = executeTextTemplate("bodyText", bodyText, data)
	if err != nil {
		return "", "", "", fmt.Errorf("body text template: %w", err)
	}
	bodyHTML, err = executeHTMLTemplate("bodyHTML", bodyHTML, data)
	if err != nil {
		return "", "", "", fmt.Errorf("body HTML template: %w", err)
	}

	return subject, bodyHTML, bodyText, nil
}

// normalizeTemplatePlaceholders rewrites {{key}} to {{.key}} for each key in data (longest keys first to avoid partial matches).
func normalizeTemplatePlaceholders(s string, data JSONMap) string {
	if len(data) == 0 {
		return s
	}
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	for _, k := range keys {
		old := fmt.Sprintf("{{%s}}", k)
		new := fmt.Sprintf("{{.%s}}", k)
		s = strings.ReplaceAll(s, old, new)
	}
	return s
}

func executeTextTemplate(name, text string, data JSONMap) (string, error) {
	t, err := texttemplate.New(name).Parse(text)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func executeHTMLTemplate(name, text string, data JSONMap) (string, error) {
	t, err := htmltemplate.New(name).Parse(text)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
