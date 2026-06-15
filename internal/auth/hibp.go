package auth

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// HIBPService checks passwords against the Have I Been Pwned API
type HIBPService struct {
	httpClient *http.Client
	enabled    bool
}

// NewHIBPService creates a new HIBP service
func NewHIBPService() *HIBPService {
	return &HIBPService{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		enabled: os.Getenv("ENABLE_PASSWORD_BREACH_CHECK") == "true",
	}
}

// IsEnabled returns whether HIBP checking is enabled
func (h *HIBPService) IsEnabled() bool {
	return h.enabled
}

// CheckPassword checks if a password has appeared in known data breaches
// Uses the k-anonymity model: only the first 5 characters of the SHA-1 hash are sent to the API
func (h *HIBPService) CheckPassword(password string) (bool, error) {
	if !h.enabled {
		return false, nil
	}

	// Hash the password with SHA-1
	hash := sha1.New()
	hash.Write([]byte(password))
	sha1Hash := hex.EncodeToString(hash.Sum(nil))
	prefix := strings.ToUpper(sha1Hash[:5])
	suffix := strings.ToUpper(sha1Hash[5:])

	// Query the HIBP API with the prefix
	url := fmt.Sprintf("https://api.pwnedpasswords.com/range/%s", prefix)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "FunctionFly-Auth/1.0")
	req.Header.Set("Add-Padding", "true") // Request padding for privacy

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to query HIBP API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HIBP API returned status %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read HIBP response: %w", err)
	}

	// Parse the response - each line is in the format "SUFFIX:COUNT"
	lines := strings.Split(string(body), "\r\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		if strings.ToUpper(parts[0]) == suffix {
			return true, nil // Password found in breaches
		}
	}

	return false, nil // Password not found in breaches
}
