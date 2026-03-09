package security

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type ClamAVConfig struct {
	URL        string
	Timeout    time.Duration
	FailOpen   bool
	MockMode   bool
	MockResult string
}

type ClamAVScanner struct {
	config  ClamAVConfig
	client  *http.Client
	logger  *logrus.Logger
	healthy bool
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
		config.Timeout = 30 * time.Second
	}

	scanner := &ClamAVScanner{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger:  logger,
		healthy: false,
	}

	if config.MockMode {
		scanner.healthy = true
		logger.Info("ClamAV scanner initialized in MOCK mode")
	} else if config.URL != "" {
		scanner.checkHealth()
	}

	return scanner
}

func (c *ClamAVScanner) checkHealth() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.config.URL+"/health", nil)
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

	if resp.StatusCode == http.StatusOK {
		c.healthy = true
		c.logger.Info("ClamAV scanner is healthy")
	} else {
		c.logger.Warnf("ClamAV health check returned status: %d", resp.StatusCode)
	}
}

func (c *ClamAVScanner) IsHealthy() bool {
	return c.healthy
}

func (c *ClamAVScanner) GetEngineInfo() (version, databaseVersion string) {
	if c.config.MockMode {
		return "ClamAV Mock 1.0.0", "MockDB 20240101"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.config.URL+"/version", nil)
	if err != nil {
		return "unknown", "unknown"
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "unknown", "unknown"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
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
			fmt.Sscanf(string(body), "ClamAV %s/%s", &version, &databaseVersion)
		}
	}

	return version, databaseVersion
}

func (c *ClamAVScanner) ScanFile(filename string, fileContent []byte) (*ClamAVScanResult, error) {
	startTime := time.Now()

	result := &ClamAVScanResult{
		Filename:     filename,
		FileSize:     int64(len(fileContent)),
		ScannedAt:    startTime,
		Engine:       "ClamAV",
		EngineVer:    "unknown",
		DatabaseVer:  "unknown",
		ScanDuration: 0,
	}

	if c.config.MockMode {
		return c.mockScan(filename, fileContent, startTime), nil
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

	if !c.healthy {
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

	return c.doScan(filename, fileContent, startTime)
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

func (c *ClamAVScanner) doScan(filename string, fileContent []byte, startTime time.Time) (*ClamAVScanResult, error) {
	result := &ClamAVScanResult{
		Filename:  filename,
		FileSize:  int64(len(fileContent)),
		ScannedAt: startTime,
		Engine:    "ClamAV",
	}

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer writer.Close()

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

	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.URL+"/scan", pr)
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

	body, err := io.ReadAll(resp.Body)
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

	bodyStr := string(body)

	switch {
	case contains(bodyStr, "OK"):
		result.Status = "OK"
	case contains(bodyStr, "FOUND"):
		result.Status = "FOUND"
		if idx := indexOf(bodyStr, ":"); idx > 0 {
			result.VirusName = bodyStr[idx+1:]
		}
	case contains(bodyStr, "ERROR"):
		result.Status = "ERROR"
	default:
		result.Status = "UNKNOWN"
	}

	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func LoadClamAVConfig() ClamAVConfig {
	return ClamAVConfig{
		URL:        getEnvOrDefault("CLAMAV_URL", "http://clamav:3310"),
		Timeout:    getDurationEnvOrDefault("CLAMAV_TIMEOUT", 30*time.Second),
		FailOpen:   getEnvOrDefault("CLAMAV_FAIL_OPEN", "true") == "true",
		MockMode:   getEnvOrDefault("CLAMAV_MOCK", "false") == "true",
		MockResult: getEnvOrDefault("CLAMAV_MOCK_RESULT", "clean"),
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
