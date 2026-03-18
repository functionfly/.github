// Package flywheel provides HTTP handlers for the Flywheel Network
package flywheel

import (
	"regexp"
	"strings"
)

// sanitizer provides HTML/XSS sanitization for user-generated content

// Dangerous tags that can execute JavaScript or load external content
var dangerousTags = []string{
	"script", "iframe", "object", "embed", "applet", "form",
	"button", "select", "textarea", "base", "link", "meta",
	"head", "style", "xml", "xss", "noscript", "noframes",
	"frameset", "frame", "plaintext", "svg", "math",
}

// Event handler attribute pattern
var eventHandlerPattern = regexp.MustCompile(`(?i)\s+(on\w+)\s*=\s*`)

// JavaScript: URL pattern
var javascriptPattern = regexp.MustCompile(`(?i)\bjavascript\s*:`)

// Data: URL pattern
var dataPattern = regexp.MustCompile(`(?i)\bdata\s*:`)

// SVG with event handlers
var svgEventPattern = regexp.MustCompile(`(?i)<svg[^>]*\s+on\w+\s*=`)

// buildDangerousTagPattern creates a regex pattern for a specific tag
func buildDangerousTagPattern(tag string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)<` + tag + `[^>]*>[\s\S]*?</` + tag + `|<\s*` + tag + `[^>]*/?>`)
}

// buildSelfClosingTagPattern creates a pattern for self-closing tags
func buildSelfClosingTagPattern(tag string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)<\s*` + tag + `[^>]*/?>`)
}

// SanitizeString sanitizes a string to prevent XSS attacks
// It removes dangerous HTML tags, event handlers, and potentially dangerous URL schemes
func SanitizeString(input string) string {
	if input == "" {
		return input
	}

	output := input

	// Remove dangerous tags with content (multiple passes for nested tags)
	for i := 0; i < 3; i++ { // Multiple passes to handle nested tags
		prev := output
		for _, tag := range dangerousTags {
			output = buildDangerousTagPattern(tag).ReplaceAllString(output, "")
			output = buildSelfClosingTagPattern(tag).ReplaceAllString(output, "")
		}
		if output == prev {
			break
		}
	}

	// Remove remaining dangerous opening tags without content
	for _, tag := range dangerousTags {
		output = regexp.MustCompile(`(?i)<`+tag+`[^>]*>`).ReplaceAllString(output, "")
	}

	// Remove event handler attributes
	output = eventHandlerPattern.ReplaceAllString(output, " data-removed=")

	// Remove javascript: URLs
	output = javascriptPattern.ReplaceAllString(output, "removed:")

	// Remove data: URLs
	output = dataPattern.ReplaceAllString(output, "removed:")

	// Remove SVG event handlers
	output = svgEventPattern.ReplaceAllString(output, "<svg data-removed>")

	return output
}

// SanitizeContent sanitizes user content fields like thread/reply content
// This is the main entry point for sanitizing user-generated content
func SanitizeContent(content string) string {
	if content == "" {
		return content
	}

	// Apply sanitization
	sanitized := SanitizeString(content)

	// Additional cleanup: remove leading/trailing whitespace but preserve internal whitespace
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

// SanitizeThread sanitizes thread fields that contain user content
func SanitizeThread(title, content string) (sanitizedTitle, sanitizedContent string) {
	return SanitizeContent(title), SanitizeContent(content)
}

// SanitizeReply sanitizes reply content
func SanitizeReply(content string) string {
	return SanitizeContent(content)
}

// SanitizeCategory sanitizes category name and description
func SanitizeCategory(name, description string) (sanitizedName, sanitizedDescription string) {
	return SanitizeContent(name), SanitizeContent(description)
}

// SanitizeChallenge sanitizes challenge title and description
func SanitizeChallenge(title, description string) (sanitizedTitle, sanitizedDescription string) {
	return SanitizeContent(title), SanitizeContent(description)
}
