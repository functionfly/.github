package conversations

import "strings"

// MessagePreview returns a truncated preview suitable for inbox list rows.
func MessagePreview(content string) string {
	s := strings.TrimSpace(content)
	if s == "" {
		return "(empty message)"
	}
	r := []rune(s)
	const max = 120
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}
