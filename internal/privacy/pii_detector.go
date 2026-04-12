package privacy

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// PIIDetector provides PII detection and redaction capabilities
type PIIDetector struct {
	patterns     map[string]*regexp.Regexp
	redactMode   bool
	confidenceThreshold float64
}

// NewPIIDetector creates a new PII detector
func NewPIIDetector() *PIIDetector {
	d := &PIIDetector{
		patterns:            make(map[string]*regexp.Regexp),
		confidenceThreshold: 0.8,
	}
	d.initPatterns()
	return d
}

// SetRedactMode enables/disables redaction mode
func (d *PIIDetector) SetRedactMode(enabled bool) {
	d.redactMode = enabled
}

// SetConfidenceThreshold sets the detection confidence threshold (0.0-1.0)
func (d *PIIDetector) SetConfidenceThreshold(threshold float64) {
	if threshold < 0 {
		threshold = 0
	}
	if threshold > 1 {
		threshold = 1
	}
	d.confidenceThreshold = threshold
}

// initPatterns initializes the regex patterns for PII detection
func (d *PIIDetector) initPatterns() {
	// Email addresses
	d.patterns["email"] = regexp.MustCompile(`(?i)[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}`)

	// Phone numbers (various formats)
	d.patterns["phone"] = regexp.MustCompile(`(\+?1[-.\s]?)?\(?[0-9]{3}\)?[-.\s]?[0-9]{3}[-.\s]?[0-9]{4}`)

	// SSN (Social Security Number)
	d.patterns["ssn"] = regexp.MustCompile(`\b\d{3}[-.\s]?\d{2}[-.\s]?\d{4}\b`)

	// Credit card numbers (major card types)
	d.patterns["credit_card"] = regexp.MustCompile(`\b(?:\d[ -]*?){13,16}\b`)

	// IP addresses (IPv4)
	d.patterns["ip_address"] = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)

	// API keys (common patterns)
	d.patterns["api_key"] = regexp.MustCompile(`(?i)(api[_-]?key|apikey)[:\s=]+["']?[a-zA-Z0-9_\-]{16,}["']?`)

	// Passwords in various formats
	d.patterns["password"] = regexp.MustCompile(`(?i)(password|passwd|pwd)[:\s=]+["']?[^"'\s]{4,}["']?`)

	// Secret tokens
	d.patterns["secret"] = regexp.MustCompile(`(?i)(secret|token|auth)[:\s=]+["']?[a-zA-Z0-9_\-]{16,}["']?`)

	// AWS Access Key ID
	d.patterns["aws_key"] = regexp.MustCompile(`(?i)AKIA[0-9A-Z]{16}`)

	// GitHub tokens
	d.patterns["github_token"] = regexp.MustCompile(`(?i)gh[pousr]_[A-Za-z0-9_]{36,}`)

	// Slack tokens
	d.patterns["slack_token"] = regexp.MustCompile(`(?i)xox[baprs]-[0-9]{10,13}-[0-9]{10,13}-[a-zA-Z0-9]{24}`)

	// Private keys
	d.patterns["private_key"] = regexp.MustCompile(`(?i)-----BEGIN (RSA |DSA |EC |OPENSSH )?PRIVATE KEY-----`)

	// URL with credentials
	d.patterns["url_with_creds"] = regexp.MustCompile(`[a-zA-Z]+://[^:]+:[^@]+@[a-zA-Z0-9.-]+`)

	// Address patterns (simplified)
	d.patterns["address"] = regexp.MustCompile(`(?i)\d+\s+[a-zA-Z\s]+(?:street|st|avenue|ave|road|rd|boulevard|blvd|lane|ln|drive|dr|way|court|ct|circle|cir|trail|trl|parkway|pkwy|highway|hwy)\b`)

	// Date of birth patterns
	d.patterns["dob"] = regexp.MustCompile(`(?i)(birth[\s_-]?date|dob|date[\s_-]?of[\s_-]?birth)[:\s=]+["']?\d{1,4}[/-]\d{1,2}[/-]\d{1,4}["']?`)
}

// DetectPII scans data for PII and returns detection results
func (d *PIIDetector) DetectPII(data interface{}) *PIIDetectionResult {
	result := &PIIDetectionResult{
		HasPII:     false,
		Categories: []string{},
		Confidence: 0.0,
		Matches:    []PIIMatch{},
	}

	// Convert data to string for scanning
	var content string
	switch v := data.(type) {
	case string:
		content = v
	case []byte:
		content = string(v)
	case map[string]interface{}, []interface{}:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return result
		}
		content = string(jsonBytes)
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return result
		}
		content = string(jsonBytes)
	}

	// Check each pattern
	totalMatches := 0
	for category, pattern := range d.patterns {
		matches := pattern.FindAllStringIndex(content, -1)
		if len(matches) > 0 {
			result.HasPII = true
			result.Categories = append(result.Categories, category)
			totalMatches += len(matches)

			for _, match := range matches {
				matchedValue := content[match[0]:match[1]]
				// Calculate confidence based on pattern specificity
				confidence := d.calculateConfidence(category, matchedValue)
				if confidence >= d.confidenceThreshold {
					result.Matches = append(result.Matches, PIIMatch{
						Type:       category,
						Value:      matchedValue,
						Position:   match[0],
						Length:     match[1] - match[0],
						Confidence: confidence,
					})
				}
			}
		}
	}

	// Calculate overall confidence
	if len(result.Matches) > 0 {
		var totalConfidence float64
		for _, match := range result.Matches {
			totalConfidence += match.Confidence
		}
		result.Confidence = totalConfidence / float64(len(result.Matches))
	}

	// Generate redacted version if enabled
	if d.redactMode && result.HasPII {
		result.RedactedData = d.RedactData(content, result.Matches)
	}

	return result
}

// DetectPIIInJSON scans JSON data for PII
func (d *PIIDetector) DetectPIIInJSON(jsonData []byte) (*PIIDetectionResult, error) {
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	return d.DetectPII(data), nil
}

// RedactData redacts PII from content based on matches
func (d *PIIDetector) RedactData(content string, matches []PIIMatch) string {
	if len(matches) == 0 {
		return content
	}

	// Sort matches by position (descending) to replace from end to start
	// This prevents position shifts during replacement
	sortedMatches := make([]PIIMatch, len(matches))
	copy(sortedMatches, matches)

	for i, j := 0, len(sortedMatches)-1; i < j; i, j = i+1, j-1 {
		sortedMatches[i], sortedMatches[j] = sortedMatches[j], sortedMatches[i]
	}

	result := content
	for _, match := range sortedMatches {
		redaction := d.getRedactionString(match.Type, match.Value)
		result = result[:match.Position] + redaction + result[match.Position+match.Length:]
	}

	return result
}

// RedactJSON redacts PII from JSON data
func (d *PIIDetector) RedactJSON(jsonData []byte) ([]byte, *PIIDetectionResult, error) {
	result := d.DetectPII(string(jsonData))
	if !result.HasPII {
		return jsonData, result, nil
	}

	redacted := d.RedactData(string(jsonData), result.Matches)
	return []byte(redacted), result, nil
}

// RedactJSONFields redacts specific fields in JSON data
func (d *PIIDetector) RedactJSONFields(jsonData []byte, fieldsToRedact []string) ([]byte, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return jsonData, err
	}

	d.redactFields(data, fieldsToRedact, "")

	return json.Marshal(data)
}

// redactFields recursively redacts fields in a map
func (d *PIIDetector) redactFields(data map[string]interface{}, fieldsToRedact []string, prefix string) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		// Check if this key should be redacted
		shouldRedact := false
		for _, field := range fieldsToRedact {
			if strings.EqualFold(key, field) || strings.EqualFold(fullKey, field) {
				shouldRedact = true
				break
			}
		}

		if shouldRedact {
			data[key] = "[REDACTED]"
			continue
		}

		// Recursively process nested objects
		switch v := value.(type) {
		case map[string]interface{}:
			d.redactFields(v, fieldsToRedact, fullKey)
		case []interface{}:
			for i, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					d.redactFields(itemMap, fieldsToRedact, fmt.Sprintf("%s[%d]", fullKey, i))
				}
			}
		}
	}
}

// calculateConfidence calculates confidence score for a match
func (d *PIIDetector) calculateConfidence(category, value string) float64 {
	switch category {
	case "email":
		// Valid email format
		if strings.Contains(value, "@") && strings.Contains(value, ".") {
			return 0.95
		}
		return 0.5
	case "ssn":
		// Valid SSN format (not all zeros, etc.)
		cleaned := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(value, "-", ""), ".", ""), " ", "")
		if len(cleaned) == 9 && cleaned != "000000000" && !strings.HasPrefix(cleaned, "000") {
			return 0.95
		}
		return 0.3
	case "credit_card":
		// Check Luhn algorithm (simplified)
		cleaned := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(value, "-", ""), " ", ""), "-", "")
		if len(cleaned) >= 13 && len(cleaned) <= 16 {
			return 0.85
		}
		return 0.4
	case "api_key", "secret", "aws_key", "github_token", "slack_token":
		return 0.9
	case "private_key":
		return 1.0
	case "password":
		return 0.8
	case "url_with_creds":
		return 0.9
	case "ip_address":
		// Validate IP format
		parts := strings.Split(value, ".")
		if len(parts) == 4 {
			valid := true
			for _, part := range parts {
				if part == "" {
					valid = false
					break
				}
			}
			if valid {
				return 0.85
			}
		}
		return 0.4
	case "phone":
		// Valid phone format
		digits := regexp.MustCompile(`\d`).FindAllString(value, -1)
		if len(digits) == 10 {
			return 0.85
		}
		if len(digits) == 11 && digits[0] == "1" {
			return 0.85
		}
		return 0.4
	case "dob":
		return 0.75
	case "address":
		return 0.7
	default:
		return 0.6
	}
	return 0.6
}

// getRedactionString returns an appropriate redaction string for the PII type
func (d *PIIDetector) getRedactionString(piiType, originalValue string) string {
	switch piiType {
	case "email":
		parts := strings.Split(originalValue, "@")
		if len(parts) == 2 {
			return "***@" + parts[1]
		}
		return "[EMAIL REDACTED]"
	case "phone":
		return "[PHONE REDACTED]"
	case "ssn":
		return "[SSN REDACTED]"
	case "credit_card":
		return "[CC REDACTED]"
	case "api_key", "secret", "aws_key", "github_token", "slack_token":
		return "[API KEY REDACTED]"
	case "private_key":
		return "[PRIVATE KEY REDACTED]"
	case "password":
		return "[PASSWORD REDACTED]"
	case "ip_address":
		return "[IP REDACTED]"
	case "url_with_creds":
		// Keep the URL structure but redact credentials
		return regexp.MustCompile(`:[^@]+@`).ReplaceAllString(originalValue, ":***@")
	case "dob":
		return "[DOB REDACTED]"
	case "address":
		return "[ADDRESS REDACTED]"
	default:
		return "[PII REDACTED]"
	}
}

// SensitiveFieldNames returns common sensitive field names to check
func SensitiveFieldNames() []string {
	return []string{
		"password", "passwd", "pwd",
		"secret", "api_key", "apikey", "api-key",
		"token", "auth_token", "access_token", "refresh_token",
		"private_key", "privatekey",
		"credit_card", "creditcard", "cc_number", "card_number",
		"ssn", "social_security",
		"dob", "birth_date", "date_of_birth",
		"phone", "phone_number", "mobile",
		"email", "email_address",
		"address", "street_address",
		"zip", "postal_code", "zipcode",
	}
}

// IsSensitiveField checks if a field name is commonly sensitive
func IsSensitiveField(fieldName string) bool {
	fieldLower := strings.ToLower(fieldName)
	for _, sensitive := range SensitiveFieldNames() {
		if strings.Contains(fieldLower, sensitive) {
			return true
		}
	}
	return false
}

// SanitizeJSON sanitizes JSON data by checking field names and redacting PII
func (d *PIIDetector) SanitizeJSON(jsonData []byte) ([]byte, *PIIDetectionResult, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return jsonData, nil, err
	}

	// First, redact sensitive fields by name
	d.redactSensitiveFieldsByName(data)

	// Then detect and redact PII in values
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return jsonData, nil, err
	}

	return d.RedactJSON(jsonBytes)
}

// redactSensitiveFieldsByName redacts fields based on their names
func (d *PIIDetector) redactSensitiveFieldsByName(data map[string]interface{}) {
	for key, value := range data {
		if IsSensitiveField(key) {
			data[key] = "[REDACTED BY FIELD NAME]"
			continue
		}

		// Recursively process nested objects
		switch v := value.(type) {
		case map[string]interface{}:
			d.redactSensitiveFieldsByName(v)
		case []interface{}:
			for _, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					d.redactSensitiveFieldsByName(itemMap)
				}
			}
		}
	}
}
