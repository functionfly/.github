package security

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/sirupsen/logrus"
)

const (
	defaultClamAVTimeout   = 30 * time.Second
	defaultMaxFileSize     = 100 << 20 // 100 MiB
	defaultMaxResponseSize = 512 << 10 // 512 KiB
	maxFilenameLen         = 255
	maxVirusNameLen        = 256
)

type ClamAVConfig struct {
	URL             string
	Timeout         time.Duration
	FailOpen        bool
	MockMode        bool
	MockResult      string
	MaxFileSize     int64
	MaxResponseSize int64
}

type ClamAVScanner struct {
	config   ClamAVConfig
	client   *http.Client
	logger   *logrus.Logger
	healthy  bool
	healthMu sync.RWMutex
}

type ClamAVScanResult struct {
	Filename     string    `json:"filename"`
	FileHash     string    `json:"file_hash"`
	FileSize     int64     `json:"file_size"`
	Status       string    `json:"status"`
	VirusName    string    `json:"virus_name,omitempty"`
	Infected     bool      `json:"infected"`
	Error        string    `json:"error,omitempty"`
	ScanDuration int64     `json:"scan_duration_ms"`
	Engine       string    `json:"engine"`
	EngineVer    string    `json:"engine_version"`
	DatabaseVer  string    `json:"database_version"`
	ScannedAt    time.Time `json:"scanned_at"`
}

func NewClamAVScanner(config ClamAVConfig, logger *logrus.Logger) *ClamAVScanner {
	if config.Timeout == 0 {
		config.Timeout = defaultClamAVTimeout
	}
	if config.MaxFileSize <= 0 {
		config.MaxFileSize = defaultMaxFileSize
	}
	if config.MaxResponseSize <= 0 {
		config.MaxResponseSize = defaultMaxResponseSize
	}

	transport := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}
	scanner := &ClamAVScanner{
		config: config,
		client: &http.Client{
			Timeout:   config.Timeout,
			Transport: transport,
		},
		logger:  logger,
		healthy: false,
	}

	if config.MockMode {
		scanner.healthMu.Lock()
		scanner.healthy = true
		scanner.healthMu.Unlock()
		logger.Info("ClamAV scanner initialized in MOCK mode")
		return scanner
	}
	if config.URL != "" {
		if err := scanner.validateURL(config.URL); err != nil {
			logger.Warnf("ClamAV URL validation failed: %v", err)
		} else {
			scanner.checkHealth()
		}
	}

	return scanner
}

// validateURL ensures URL is http/https and well-formed to reduce SSRF risk.
func (c *ClamAVScanner) validateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	switch u.Scheme {
	case "http", "https":
		// allowed
	default:
		return fmt.Errorf("unsupported scheme %q (use http or https)", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("missing host in URL")
	}
	return nil
}

func (c *ClamAVScanner) checkHealth() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.URL+"/health", nil)
	if err != nil {
		c.logger.Warnf("ClamAV health check failed: %v", err)
		return
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Warnf("ClamAV health check failed: %v", err)
		return
	}
	defer resp.Body.Close()

	ok := resp.StatusCode == http.StatusOK
	c.healthMu.Lock()
	c.healthy = ok
	c.healthMu.Unlock()
	if ok {
		c.logger.Info("ClamAV scanner is healthy")
	} else {
		c.logger.Warnf("ClamAV health check returned status: %d", resp.StatusCode)
	}
}

func (c *ClamAVScanner) IsHealthy() bool {
	c.healthMu.RLock()
	defer c.healthMu.RUnlock()
	return c.healthy
}

func (c *ClamAVScanner) GetEngineInfo() (version, databaseVersion string) {
	if c.config.MockMode {
		return "ClamAV Mock 1.0.0", "MockDB 20240101"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.URL+"/version", nil)
	if err != nil {
		return "unknown", "unknown"
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "unknown", "unknown"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(io.LimitReader(resp.Body, defaultMaxResponseSize))
		if err != nil {
			return version, databaseVersion
		}
		var info struct {
			Version string `json:"version"`
			DBVer   string `json:"database_version"`
		}
		if json.Unmarshal(body, &info) == nil && (info.Version != "" || info.DBVer != "") {
			version, databaseVersion = info.Version, info.DBVer
		} else {
			_, _ = fmt.Sscanf(string(body), "ClamAV %s/%s", &version, &databaseVersion)
		}
	}

	return version, databaseVersion
}

func (c *ClamAVScanner) ScanFile(filename string, fileContent []byte) (*ClamAVScanResult, error) {
	return c.ScanFileContext(context.Background(), filename, fileContent)
}

func (c *ClamAVScanner) ScanFileContext(ctx context.Context, filename string, fileContent []byte) (*ClamAVScanResult, error) {
	startTime := time.Now()

	safeName := sanitizeFilename(filename)
	if safeName == "" {
		safeName = "unnamed"
	}

	result := &ClamAVScanResult{
		Filename:     safeName,
		FileSize:     int64(len(fileContent)),
		ScannedAt:    startTime,
		Engine:       "ClamAV",
		EngineVer:    "unknown",
		DatabaseVer:  "unknown",
		ScanDuration: 0,
	}

	if int64(len(fileContent)) > c.config.MaxFileSize {
		result.Error = fmt.Sprintf("file size %d exceeds max %d", len(fileContent), c.config.MaxFileSize)
		result.Status = "ERROR"
		result.ScanDuration = time.Since(startTime).Milliseconds()
		return result, fmt.Errorf("%s", result.Error)
	}

	if c.config.MockMode {
		return c.mockScan(safeName, fileContent, startTime), nil
	}

	if c.config.URL == "" {
		if c.config.FailOpen {
			c.logger.Warn("ClamAV not configured, failing open (treating as clean)")
			result.Status = "OK"
			result.Infected = false
			result.ScanDuration = time.Since(startTime).Milliseconds()
			return result, nil
		}
		result.Error = "ClamAV not configured"
		result.Status = "ERROR"
		return result, fmt.Errorf("ClamAV not configured")
	}

	c.healthMu.RLock()
	healthy := c.healthy
	c.healthMu.RUnlock()
	if !healthy {
		if c.config.FailOpen {
			c.logger.Warn("ClamAV unhealthy, failing open (treating as clean)")
			result.Status = "OK"
			result.Infected = false
			result.ScanDuration = time.Since(startTime).Milliseconds()
			return result, nil
		}
		result.Error = "ClamAV service unhealthy"
		result.Status = "ERROR"
		return result, fmt.Errorf("ClamAV service unhealthy")
	}

	return c.doScan(ctx, safeName, fileContent, startTime)
}

// sanitizeFilename returns a safe base name for multipart form (no path traversal, bounded length).
func sanitizeFilename(name string) string {
	base := filepath.Base(name)
	base = strings.TrimSpace(base)
	if base == "" || base == "." || base == ".." {
		return ""
	}
	// Restrict to valid UTF-8 and remove control chars / path separators
	var b strings.Builder
	for _, r := range base {
		if r == 0 || r == '/' || r == '\\' || r < 32 || r == 127 {
			continue
		}
		if !utf8.ValidRune(r) {
			continue
		}
		b.WriteRune(r)
	}
	out := b.String()
	if len(out) > maxFilenameLen {
		out = out[:maxFilenameLen]
	}
	return out
}

func (c *ClamAVScanner) mockScan(filename string, fileContent []byte, startTime time.Time) *ClamAVScanResult {
	result := &ClamAVScanResult{
		Filename:     filename,
		FileSize:     int64(len(fileContent)),
		ScannedAt:    startTime,
		Engine:       "ClamAV",
		EngineVer:    "1.0.0",
		DatabaseVer:  "20240101",
		ScanDuration: time.Since(startTime).Milliseconds(),
	}

	switch c.config.MockResult {
	case "infected":
		result.Status = "FOUND"
		result.VirusName = "Mock.Trojan.Test"
		result.Infected = true
	case "error":
		result.Status = "ERROR"
		result.Error = "Mock scan error"
	default:
		result.Status = "OK"
		result.Infected = false
	}

	return result
}

func (c *ClamAVScanner) doScan(ctx context.Context, filename string, fileContent []byte, startTime time.Time) (*ClamAVScanResult, error) {
	result := &ClamAVScanResult{
		Filename:  filename,
		FileSize:  int64(len(fileContent)),
		ScannedAt: startTime,
		Engine:    "ClamAV",
	}

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer func() {
			_ = writer.Close()
			_ = pw.Close()
		}()
		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			c.logger.Errorf("Failed to create form file: %v", err)
			return
		}
		if _, err := part.Write(fileContent); err != nil {
			c.logger.Errorf("Failed to write file content: %v", err)
			return
		}
	}()

	reqCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, c.config.URL+"/scan", pr)
	if err != nil {
		result.Error = err.Error()
		result.Status = "ERROR"
		return result, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		if c.config.FailOpen {
			c.logger.Warnf("ClamAV scan failed, failing open: %v", err)
			result.Status = "OK"
			result.Infected = false
			result.ScanDuration = time.Since(startTime).Milliseconds()
			return result, nil
		}
		result.Error = err.Error()
		result.Status = "ERROR"
		return result, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, c.config.MaxResponseSize))
	if err != nil {
		result.Error = err.Error()
		result.Status = "ERROR"
		return result, err
	}

	result.ScanDuration = time.Since(startTime).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		if c.config.FailOpen {
			c.logger.Warnf("ClamAV returned non-200 status: %d, failing open", resp.StatusCode)
			result.Status = "OK"
			result.Infected = false
			return result, nil
		}
		result.Error = fmt.Sprintf("ClamAV returned status: %d", resp.StatusCode)
		result.Status = "ERROR"
		return result, fmt.Errorf("ClamAV scan failed with status: %d", resp.StatusCode)
	}

	var scanResp struct {
		Status     string `json:"status"`
		VirusName  string `json:"virus_name,omitempty"`
		Version    string `json:"version,omitempty"`
		DBVersion  string `json:"db_version,omitempty"`
		ScanTimeMs int64  `json:"scan_time_ms,omitempty"`
	}

	if err := parseClamAVResponse(body, &scanResp); err != nil {
		result.Error = err.Error()
		result.Status = "ERROR"
		return result, err
	}

	result.Status = scanResp.Status
	result.VirusName = scanResp.VirusName
	result.EngineVer = scanResp.Version
	result.DatabaseVer = scanResp.DBVersion
	result.Infected = scanResp.Status == "FOUND"

	if scanResp.ScanTimeMs > 0 {
		result.ScanDuration = scanResp.ScanTimeMs
	}

	return result, nil
}

func parseClamAVResponse(body []byte, result *struct {
	Status     string `json:"status"`
	VirusName  string `json:"virus_name,omitempty"`
	Version    string `json:"version,omitempty"`
	DBVersion  string `json:"db_version,omitempty"`
	ScanTimeMs int64  `json:"scan_time_ms,omitempty"`
}) error {
	result.Status = "OK"
	if len(body) == 0 {
		return nil
	}

	// Prefer JSON response from REST API
	if json.Valid(body) {
		var j struct {
			Status     string `json:"status"`
			VirusName  string `json:"virus_name,omitempty"`
			Version    string `json:"version,omitempty"`
			DBVersion  string `json:"db_version,omitempty"`
			ScanTimeMs int64  `json:"scan_time_ms,omitempty"`
		}
		if err := json.Unmarshal(body, &j); err == nil && (j.Status != "" || j.VirusName != "") {
			result.Status = j.Status
			result.VirusName = truncateUTF8(j.VirusName, maxVirusNameLen)
			result.Version = j.Version
			result.DBVersion = j.DBVersion
			if j.ScanTimeMs > 0 {
				result.ScanTimeMs = j.ScanTimeMs
			}
			return nil
		}
	}

	// Fallback: text response (e.g. "FOUND: Eicar-Test-Signature")
	bodyStr := string(body)
	switch {
	case strings.Contains(bodyStr, "OK"):
		result.Status = "OK"
	case strings.Contains(bodyStr, "FOUND"):
		result.Status = "FOUND"
		if idx := strings.Index(bodyStr, ":"); idx >= 0 && idx+1 < len(bodyStr) {
			result.VirusName = truncateUTF8(strings.TrimSpace(bodyStr[idx+1:]), maxVirusNameLen)
		}
	case strings.Contains(bodyStr, "ERROR"):
		result.Status = "ERROR"
	default:
		result.Status = "UNKNOWN"
	}
	return nil
}

func truncateUTF8(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	for len(s) > maxBytes {
		_, size := utf8.DecodeLastRuneInString(s)
		s = s[:len(s)-size]
	}
	return s
}

func LoadClamAVConfig() ClamAVConfig {
	return ClamAVConfig{
		URL:             getEnvOrDefault("CLAMAV_URL", "http://clamav:3310"),
		Timeout:         getDurationEnvOrDefault("CLAMAV_TIMEOUT", defaultClamAVTimeout),
		FailOpen:        getEnvOrDefault("CLAMAV_FAIL_OPEN", "false") == "true",
		MockMode:        getEnvOrDefault("CLAMAV_MOCK", "false") == "true",
		MockResult:      getEnvOrDefault("CLAMAV_MOCK_RESULT", "clean"),
		MaxFileSize:     getInt64EnvOrDefault("CLAMAV_MAX_FILE_SIZE", defaultMaxFileSize),
		MaxResponseSize: getInt64EnvOrDefault("CLAMAV_MAX_RESPONSE_SIZE", defaultMaxResponseSize),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnvOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

func getInt64EnvOrDefault(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		var v int64
		if _, err := fmt.Sscanf(value, "%d", &v); err == nil && v > 0 {
			return v
		}
	}
	return defaultValue
}
