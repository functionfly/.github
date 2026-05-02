package templates

import (
	"fmt"
	"strings"
	"time"
)

type EmailTemplate struct {
	HTML string
	Text string
}

func BuildEmailURL(base, path, token string) string {
	separator := "?"
	if strings.Contains(base, "?") {
		separator = "&"
	}
	return fmt.Sprintf("%s%s%s=%s", base, separator, path, token)
}

func BaseURL(authURL, baseURL string) string {
	if authURL != "" {
		return authURL
	}
	return baseURL
}

func WithDefault(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func DefaultRiskLevel(level string) string {
	return WithDefault(level, "high")
}

func ExtractString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func ExtractInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(int); ok {
		return v
	}
	return 0
}

func TestBannerHTML(html string) string {
	banner := `<table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%;">
  <tr><td style="background:#92400e;padding:10px 20px;text-align:center;font-size:13px;font-weight:600;color:#fef3c7;">
    ⚠️ TEST EMAIL — FunctionFly Development Environment
  </td></tr>
</table>`
	return strings.Replace(html, `<table role="presentation" width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%%;">`, banner, 1)
}

func TestBannerText(text string) string {
	return "[TEST EMAIL - FunctionFly Development Environment]\n\n" + text
}

func TransactionalEmailCopyrightHTML() string {
	return fmt.Sprintf(`© %d FunctionFly. All rights reserved.`, time.Now().Year())
}

func TransactionalEmailCopyrightPlain() string {
	return fmt.Sprintf("© %d FunctionFly. All rights reserved.", time.Now().Year())
}

func TransactionalEmailCopyrightOrangeHTML() string {
	return fmt.Sprintf(`© %d FunctionFly. All rights reserved.<br>
<a href="https://functionfly.com/privacy" style="color:#f97316;text-decoration:none;">Privacy Policy</a> · 
<a href="https://functionfly.com/terms" style="color:#f97316;text-decoration:none;">Terms of Service</a>`, time.Now().Year())
}

func TransactionalEmailCopyrightOrangePlain() string {
	return fmt.Sprintf("© %d FunctionFly. All rights reserved.\nhttps://functionfly.com/privacy | https://functionfly.com/terms", time.Now().Year())
}