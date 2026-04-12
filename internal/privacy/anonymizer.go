package privacy

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"regexp"
	"strings"
)

// Anonymizer provides PII anonymization utilities
type Anonymizer struct {
	salt string
}

// NewAnonymizer creates a new anonymizer with the given salt
// The salt should be a stable, secret value stored securely
func NewAnonymizer(salt string) *Anonymizer {
	return &Anonymizer{
		salt: salt,
	}
}

// AnonymizeIP anonymizes an IP address based on the mask type
func (a *Anonymizer) AnonymizeIP(ip string, maskType PIIMaskType) string {
	if ip == "" {
		return ""
	}

	switch maskType {
	case PIIMaskTypeNone:
		return ip
	case PIIMaskTypeHash:
		return a.hashString(ip)
	case PIIMaskTypePartial:
		return a.partialMaskIP(ip)
	case PIIMaskTypeRedact:
		return "[REDACTED]"
	default:
		return ip
	}
}

// AnonymizeUserAgent anonymizes a user agent string based on the mask type
func (a *Anonymizer) AnonymizeUserAgent(ua string, maskType PIIMaskType) string {
	if ua == "" {
		return ""
	}

	switch maskType {
	case PIIMaskTypeNone:
		return ua
	case PIIMaskTypeHash:
		return a.hashString(ua)
	case PIIMaskTypePartial:
		return a.partialMaskUserAgent(ua)
	case PIIMaskTypeRedact:
		return "[REDACTED]"
	default:
		return ua
	}
}

// AnonymizeEmbedOrigin anonymizes an embed origin domain
func (a *Anonymizer) AnonymizeEmbedOrigin(origin string, maskType PIIMaskType) string {
	if origin == "" {
		return ""
	}

	switch maskType {
	case PIIMaskTypeNone:
		return origin
	case PIIMaskTypeHash:
		return a.hashString(origin)
	case PIIMaskTypePartial:
		return a.partialMaskDomain(origin)
	case PIIMaskTypeRedact:
		return "[REDACTED]"
	default:
		return origin
	}
}

// HashIPForUniqueness returns a consistent hash for an IP (used for uniqueness tracking)
func (a *Anonymizer) HashIPForUniqueness(ip string) string {
	if ip == "" {
		return ""
	}
	return a.hashString(ip)
}

// HashUserAgentForUniqueness returns a consistent hash for user agent (used for fingerprinting)
func (a *Anonymizer) HashUserAgentForUniqueness(ua string) string {
	if ua == "" {
		return ""
	}
	return a.hashString(ua)
}

// GetIPPrefix returns a privacy-preserving IP prefix for rough geolocation
// IPv4: returns first 2 octets (e.g., "192.168.x.x")
// IPv6: returns first 32 bits
func (a *Anonymizer) GetIPPrefix(ip string) string {
	if ip == "" {
		return ""
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "invalid"
	}

	// IPv4
	if parsedIP.To4() != nil {
		octets := strings.Split(ip, ".")
		if len(octets) >= 2 {
			return octets[0] + "." + octets[1] + ".x.x"
		}
		return "x.x.x.x"
	}

	// IPv6 - return first group
	groups := strings.Split(ip, ":")
	if len(groups) >= 1 {
		return groups[0] + ":xxxx:xxxx:xxxx"
	}
	return "xxxx:xxxx:xxxx:xxxx"
}

// partialMaskIP masks the last octets of an IP address
func (a *Anonymizer) partialMaskIP(ip string) string {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return ip
	}

	// IPv4
	if parsedIP.To4() != nil {
		octets := strings.Split(ip, ".")
		if len(octets) == 4 {
			return octets[0] + "." + octets[1] + ".0.0"
		}
	}

	// IPv6 - mask last 64 bits
	if strings.Contains(ip, ":") {
		parts := strings.Split(ip, ":")
		if len(parts) >= 4 {
			return strings.Join(parts[:4], ":") + ":0000:0000:0000:0000"
		}
	}

	return ip
}

// partialMaskUserAgent masks identifying info in user agent
func (a *Anonymizer) partialMaskUserAgent(ua string) string {
	// Mask specific identifiers but keep general browser/OS info
	masks := []struct {
		pattern *regexp.Regexp
		replace string
	}{
		// Mask Windows user-specific data
		{regexp.MustCompile(`Windows NT \d+\.\d+;`), "Windows;"},
		// Mask specific build numbers
		{regexp.MustCompile(`\bWin64\b`), "Win64"},
		// Mask macOS specific versions
		{regexp.MustCompile(`Mac OS X \d+[_\.]\d+([_\.]\d+)?`), "Mac OS X"},
		// Mask specific browser versions
		{regexp.MustCompile(`(Chrome|Firefox|Safari|Edge)/\d+\.\d+\.\d+\.\d+`), "$1/xx.x.x.x"},
		{regexp.MustCompile(`Version/\d+\.\d+\.\d+`), "Version/x.x.x"},
		// Mask Android/iOS device specifics
		{regexp.MustCompile(`Android \d+\.\d+; [^;)]+`), "Android x.x"},
		{regexp.MustCompile(`iPhone OS \d+[_\.]\d+`), "iOS x.x"},
		{regexp.MustCompile(`iPad; CPU OS \d+[_\.]\d+`), "iPad; CPU OS x.x"},
		// Mask any numeric IDs that might be unique
		{regexp.MustCompile(`\b\d{6,}\b`), "[ID]"},
	}

	result := ua
	for _, mask := range masks {
		result = mask.pattern.ReplaceAllString(result, mask.replace)
	}

	return result
}

// partialMaskDomain masks the full domain, keeping only TLD and partial domain
func (a *Anonymizer) partialMaskDomain(domain string) string {
	// Remove protocol if present
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimSuffix(domain, "/")

	parts := strings.Split(domain, ".")
	if len(parts) >= 2 {
		// Keep last 2 parts (domain.tld) and mask subdomain
		return "***." + parts[len(parts)-2] + "." + parts[len(parts)-1]
	}

	return "[MASKED]"
}

// hashString creates a SHA256 hash of the input with salt
func (a *Anonymizer) hashString(input string) string {
	h := sha256.New()
	h.Write([]byte(a.salt + input))
	return hex.EncodeToString(h.Sum(nil))
}

// QuickHash creates a quick hash without salt (for non-sensitive uniqueness tracking)
func QuickHash(input string) string {
	if input == "" {
		return ""
	}
	h := sha256.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))[:16] // First 16 chars for shorter storage
}

// IsPrivateIP checks if an IP is in a private range (should not be logged externally)
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// Check for private IPv4 ranges
	privateRanges := []*net.IPNet{
		// 10.0.0.0/8
		{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)},
		// 172.16.0.0/12
		{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)},
		// 192.168.0.0/16
		{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)},
		// 127.0.0.0/8 (loopback)
		{IP: net.ParseIP("127.0.0.0"), Mask: net.CIDRMask(8, 32)},
		// 169.254.0.0/16 (link-local)
		{IP: net.ParseIP("169.254.0.0"), Mask: net.CIDRMask(16, 32)},
	}

	for _, ipNet := range privateRanges {
		if ipNet.Contains(parsedIP) {
			return true
		}
	}

	// Check for IPv6 loopback
	if parsedIP.IsLoopback() {
		return true
	}

	// Check for IPv6 link-local
	if parsedIP.IsLinkLocalUnicast() || parsedIP.IsLinkLocalMulticast() {
		return true
	}

	return false
}

// GetRegionFromIP extracts a privacy-preserving region code
// Returns a general region (e.g., "US", "EU", "APAC") rather than specific geo
// Note: This is a simplified fallback. For accurate detection, use GeoIPService
// which integrates MaxMind GeoLite2 (free tier available).
//
// To use GeoLite2:
//   1. Sign up for free at https://www.maxmind.com/en/geolite2/signup
//   2. Set MAXMIND_LICENSE_KEY environment variable
//   3. The service will auto-download the database on first run
//   4. Or manually download from https://dev.maxmind.com/geoip/geolite2-free-geolocation-data
//
// Alternative: IP2Location LITE (free, CC BY-SA 4.0) - https://lite.ip2location.com
func GetRegionFromIP(ip string) string {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "unknown"
	}

	// Private IPs
	if IsPrivateIP(ip) {
		return "private"
	}

	// Simplified region detection - use GeoIPService for accurate detection
	return simplifiedRegionDetection(ip)
}

// simplifiedRegionDetection provides basic region detection without GeoLite2
// This is used as a fallback when the GeoLite2 database is not available.
// For production, use GeoIPService which provides accurate, privacy-preserving
// region detection using MaxMind GeoLite2.
func simplifiedRegionDetection(ip string) string {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "unknown"
	}

	ipStr := parsedIP.String()

	// IPv6 detection
	if parsedIP.To4() == nil {
		ipStr := parsedIP.String()
		if strings.HasPrefix(ipStr, "2001:4860:") || // Google
			strings.HasPrefix(ipStr, "2600:") || // US AWS
			strings.HasPrefix(ipStr, "2601:") || // US Comcast
			strings.HasPrefix(ipStr, "2602:") || // US
			strings.HasPrefix(ipStr, "2603:") || // US
			strings.HasPrefix(ipStr, "2604:") || // US
			strings.HasPrefix(ipStr, "2605:") || // US
			strings.HasPrefix(ipStr, "2606:") || // US
			strings.HasPrefix(ipStr, "2607:") || // US
			strings.HasPrefix(ipStr, "2608:") || // US
			strings.HasPrefix(ipStr, "2609:") { // US
			return "US"
		}
		if strings.HasPrefix(ipStr, "2a02:") || // EU
			strings.HasPrefix(ipStr, "2a01:") || // EU
			strings.HasPrefix(ipStr, "2001:4c28:") { // EU
			return "EU"
		}
		return "unknown"
	}

	// Simplified IPv4 region detection
	// Check for common US ranges
	if strings.HasPrefix(ipStr, "3.") || strings.HasPrefix(ipStr, "13.") ||
		strings.HasPrefix(ipStr, "34.") || strings.HasPrefix(ipStr, "35.") {
		return "US"
	}

	// Check for common EU ranges
	if strings.HasPrefix(ipStr, "2.") || strings.HasPrefix(ipStr, "5.") ||
		strings.HasPrefix(ipStr, "31.") || strings.HasPrefix(ipStr, "46.") ||
		strings.HasPrefix(ipStr, "51.") || strings.HasPrefix(ipStr, "77.") ||
		strings.HasPrefix(ipStr, "78.") || strings.HasPrefix(ipStr, "79.") ||
		strings.HasPrefix(ipStr, "80.") || strings.HasPrefix(ipStr, "81.") {
		return "EU"
	}

	return "unknown"
}

// AnonymizeExecutionRecord anonymizes an execution record based on privacy settings
func (a *Anonymizer) AnonymizeExecutionRecord(record *AnonymizedExecution, settings *PrivacySettings) *AnonymizedExecution {
	if record == nil || settings == nil {
		return record
	}

	// Anonymize IP if needed
	if settings.AnonymizeIP && record.IPHashPrefix != "" {
		record.IPHashPrefix = a.AnonymizeIP(record.IPHashPrefix, settings.IPMaskType)
	}

	// Anonymize User Agent if needed
	if settings.AnonymizeUserAgent && record.UserAgentHash != "" {
		record.UserAgentHash = a.AnonymizeUserAgent(record.UserAgentHash, settings.UserAgentMaskType)
	}

	// Clear specific identifiers in maximum/GDPR mode
	if settings.PrivacyLevel == PrivacyLevelMaximum || settings.PrivacyLevel == PrivacyLevelGDPR {
		record.UserID = nil
	}

	return record
}
