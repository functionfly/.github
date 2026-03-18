package auth

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// SIEMExporter is the interface for exporting audit logs to SIEM systems
type SIEMExporter interface {
	Export(ctx context.Context, logs []map[string]interface{}) error
	TestConnection(ctx context.Context) error
}

// SIEMService handles SIEM integration for audit log export
type SIEMService struct {
	siemRepo      *storage.SIEMRepository
	auditRepo     *storage.AuthAuditRepository
	httpClient    *http.Client
	encryptionKey []byte
	// Background scheduler
	stopCh       chan struct{}
	wg           sync.WaitGroup
	runScheduled bool
}

// NewSIEMService creates a new SIEM service
func NewSIEMService(siemRepo *storage.SIEMRepository, auditRepo *storage.AuthAuditRepository, encryptionKey []byte) *SIEMService {
	return &SIEMService{
		siemRepo:      siemRepo,
		auditRepo:     auditRepo,
		encryptionKey: encryptionKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SIEM Destination Types
const (
	DestinationWebhook       = "webhook"
	DestinationCloudWatch    = "cloudwatch"
	DestinationAzureSentinel = "azure_sentinel"
	DestinationGCPChronicle  = "gcp_chronicle"
	DestinationSplunk        = "splunk"
)

// Export Formats
const (
	FormatJSON = "json"
	FormatCEF  = "cef"
	FormatLEEF = "leef"
)

// ExportStatus
const (
	ExportStatusSuccess = "success"
	ExportStatusFailed  = "failed"
	ExportStatusPartial = "partial"
)

// GetExporter returns the appropriate exporter for a SIEM config
func (s *SIEMService) GetExporter(config *storage.SIEMConfig) (SIEMExporter, error) {
	// Decrypt sensitive config values
	decryptedConfig := s.decryptConfig(config.Config)

	switch config.DestinationType {
	case DestinationWebhook:
		return NewWebhookExporter(decryptedConfig, s.httpClient)
	case DestinationCloudWatch:
		return NewCloudWatchExporter(decryptedConfig)
	case DestinationAzureSentinel:
		return NewAzureSentinelExporter(decryptedConfig, s.httpClient)
	case DestinationGCPChronicle:
		return NewGCPChronicleExporter(decryptedConfig, s.httpClient)
	case DestinationSplunk:
		return NewSplunkExporter(decryptedConfig, s.httpClient)
	default:
		return nil, fmt.Errorf("unsupported destination type: %s", config.DestinationType)
	}
}

// Export sends audit logs to the configured SIEM destination
func (s *SIEMService) Export(ctx context.Context, configID uuid.UUID) error {
	config, err := s.siemRepo.GetByID(ctx, configID)
	if err != nil {
		return fmt.Errorf("failed to get SIEM config: %w", err)
	}
	if config == nil {
		return fmt.Errorf("SIEM config not found")
	}
	if !config.Enabled {
		return fmt.Errorf("SIEM config is disabled")
	}

	// Get logs since last export
	var since time.Time
	if config.LastExportAt != nil {
		since = *config.LastExportAt
	} else {
		// Default to last 24 hours if no previous export
		since = time.Now().Add(-24 * time.Hour)
	}

	logs, err := s.auditRepo.GetByTenantSince(ctx, config.TenantID, since)
	if err != nil {
		return fmt.Errorf("failed to get audit logs: %w", err)
	}

	if len(logs) == 0 {
		// No logs to export, update timestamp anyway
		return s.siemRepo.UpdateLastExportAt(ctx, configID, time.Now())
	}

	// Get exporter
	exporter, err := s.GetExporter(config)
	if err != nil {
		s.siemRepo.CreateExportLog(ctx, &storage.SIEMExportLog{
			SIEMConfigID: configID,
			Status:       ExportStatusFailed,
			ErrorMessage: err.Error(),
		})
		return err
	}

	// Transform logs to appropriate format
	transformedLogs, err := s.transformLogs(logs, config.ExportFormat)
	if err != nil {
		s.siemRepo.CreateExportLog(ctx, &storage.SIEMExportLog{
			SIEMConfigID: configID,
			Status:       ExportStatusFailed,
			ErrorMessage: err.Error(),
		})
		return err
	}

	// Export to SIEM
	err = exporter.Export(ctx, *transformedLogs)
	exportTime := time.Now()

	if err != nil {
		s.siemRepo.CreateExportLog(ctx, &storage.SIEMExportLog{
			SIEMConfigID: configID,
			Status:       ExportStatusFailed,
			RecordsSent:  0,
			ErrorMessage: err.Error(),
		})
		s.siemRepo.UpdateLastExportAt(ctx, configID, exportTime)
		return err
	}

	// Log success
	s.siemRepo.CreateExportLog(ctx, &storage.SIEMExportLog{
		SIEMConfigID: configID,
		Status:       ExportStatusSuccess,
		RecordsSent:  len(logs),
	})
	s.siemRepo.UpdateLastExportAt(ctx, configID, exportTime)

	return nil
}

// TestConnection tests the connection to a SIEM destination
func (s *SIEMService) TestConnection(ctx context.Context, configID uuid.UUID) error {
	config, err := s.siemRepo.GetByID(ctx, configID)
	if err != nil {
		return fmt.Errorf("failed to get SIEM config: %w", err)
	}
	if config == nil {
		return fmt.Errorf("SIEM config not found")
	}

	exporter, err := s.GetExporter(config)
	if err != nil {
		return err
	}

	return exporter.TestConnection(ctx)
}

// transformLogs transforms audit logs to the specified format
func (s *SIEMService) transformLogs(logs []*storage.AuthAuditLog, format string) (*[]map[string]interface{}, error) {
	result := make([]map[string]interface{}, len(logs))

	for i, log := range logs {
		switch format {
		case FormatJSON:
			result[i] = s.toJSONFormat(log)
		case FormatCEF:
			result[i] = s.toCEFFormat(log)
		case FormatLEEF:
			result[i] = s.toLEEFFormat(log)
		default:
			result[i] = s.toJSONFormat(log)
		}
	}

	return &result, nil
}

func (s *SIEMService) toJSONFormat(log *storage.AuthAuditLog) map[string]interface{} {
	return map[string]interface{}{
		"id":             log.ID.String(),
		"tenant_id":      log.TenantID.String(),
		"user_id":        log.UserID.String(),
		"event_type":     log.EventType,
		"event_data":     log.EventData,
		"ip_address":     log.IPAddress,
		"user_agent":     log.UserAgent,
		"success":        log.Success,
		"failure_reason": log.FailureReason,
		"timestamp":      log.CreatedAt.Format(time.RFC3339),
	}
}

func (s *SIEMService) toCEFFormat(log *storage.AuthAuditLog) map[string]interface{} {
	// Common Event Format: CEF:Version|Device Vendor|Device Product|Device Version|Signature ID|Name|Severity|Extension
	severity := "5"
	if !log.Success {
		severity = "8"
	}

	eventDataJSON, _ := json.Marshal(log.EventData)

	cefEvent := fmt.Sprintf("CEF:0|FunctionFly|Auth|%s|100|%s|%s|",
		"1.0", log.EventType, severity)

	return map[string]interface{}{
		"cef":            cefEvent,
		"device_vendor":  "FunctionFly",
		"device_product": "Auth",
		"device_version": "1.0",
		"signature_id":   "100",
		"name":           log.EventType,
		"severity":       severity,
		"src":            log.IPAddress,
		"suser":          log.UserID.String(),
		"msg":            string(eventDataJSON),
		"rt":             log.CreatedAt.Format(time.RFC3339),
	}
}

func (s *SIEMService) toLEEFFormat(log *storage.AuthAuditLog) map[string]interface{} {
	// Log Event Extended Event Format: LEEF:Version|Vendor|Product|Version|EventID|Name|
	eventDataJSON, _ := json.Marshal(log.EventData)

	return map[string]interface{}{
		"leef":      fmt.Sprintf("LEEF:1.0|FunctionFly|Auth|1.0|%s", log.EventType),
		"vendor":    "FunctionFly",
		"product":   "Auth",
		"version":   "1.0",
		"event_id":  log.EventType,
		"src":       log.IPAddress,
		"usrName":   log.UserID.String(),
		"extension": string(eventDataJSON),
		"timestamp": log.CreatedAt.Format(time.RFC3339),
	}
}

// decryptConfig decrypts sensitive values in the config
func (s *SIEMService) decryptConfig(config map[string]interface{}) map[string]interface{} {
	if s.encryptionKey == nil {
		return config
	}

	decrypted := make(map[string]interface{})
	for k, v := range config {
		if str, ok := v.(string); ok && isEncryptedValue(str) {
			decrypted[k] = s.decryptValue(str)
		} else {
			decrypted[k] = v
		}
	}
	return decrypted
}

func isEncryptedValue(s string) bool {
	return strings.HasPrefix(s, "enc:")
}

func (s *SIEMService) decryptValue(encrypted string) string {
	if !strings.HasPrefix(encrypted, "enc:") {
		return encrypted
	}

	encoded := strings.TrimPrefix(encrypted, "enc:")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return encrypted
	}

	mac := hmac.New(sha256.New, s.encryptionKey)
	mac.Write(decoded[:len(decoded)-32])
	expectedMAC := mac.Sum(nil)

	if !hmac.Equal(decoded[len(decoded)-32:], expectedMAC) {
		return encrypted
	}

	return string(decoded[:len(decoded)-32])
}

// ============== Webhook Exporter ==============

// WebhookExporter exports logs to a generic webhook endpoint
type WebhookExporter struct {
	Endpoint string
	APIKey   string
	Headers  map[string]string
	client   *http.Client
}

// NewWebhookExporter creates a new Webhook exporter
func NewWebhookExporter(config map[string]interface{}, client *http.Client) (*WebhookExporter, error) {
	endpoint, ok := config["endpoint"].(string)
	if !ok || endpoint == "" {
		return nil, fmt.Errorf("webhook endpoint is required")
	}

	apiKey, _ := config["api_key"].(string)
	headers := make(map[string]string)
	if h, ok := config["headers"].(map[string]interface{}); ok {
		for k, v := range h {
			if str, ok := v.(string); ok {
				headers[k] = str
			}
		}
	}

	return &WebhookExporter{
		Endpoint: endpoint,
		APIKey:   apiKey,
		Headers:  headers,
		client:   client,
	}, nil
}

func (e *WebhookExporter) Export(ctx context.Context, logs []map[string]interface{}) error {
	payload, err := json.Marshal(logs)
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Audit-Log-Format", "json")

	if e.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.APIKey)
	}

	for k, v := range e.Headers {
		req.Header.Set(k, v)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (e *WebhookExporter) TestConnection(ctx context.Context) error {
	// Send a test event
	testLog := []map[string]interface{}{
		{
			"id":         uuid.New().String(),
			"event_type": "siem_test",
			"success":    true,
			"timestamp":  time.Now().Format(time.RFC3339),
			"test_event": true,
		},
	}

	return e.Export(ctx, testLog)
}

// ============== AWS CloudWatch Exporter ==============

// CloudWatchExporter exports logs to AWS CloudWatch
type CloudWatchExporter struct {
	Region      string
	LogGroup    string
	AccessKeyID string
	SecretKey   string
	StreamName  string
}

// NewCloudWatchExporter creates a new CloudWatch exporter
func NewCloudWatchExporter(config map[string]interface{}) (*CloudWatchExporter, error) {
	region, _ := config["region"].(string)
	logGroup, _ := config["log_group"].(string)
	accessKeyID, _ := config["access_key_id"].(string)
	secretKey, _ := config["secret_access_key"].(string)
	streamName, _ := config["stream_name"].(string)

	if region == "" || logGroup == "" {
		return nil, fmt.Errorf("region and log_group are required")
	}

	return &CloudWatchExporter{
		Region:      region,
		LogGroup:    logGroup,
		AccessKeyID: accessKeyID,
		SecretKey:   secretKey,
		StreamName:  streamName,
	}, nil
}

func (e *CloudWatchExporter) Export(ctx context.Context, logs []map[string]interface{}) error {
	if len(logs) == 0 {
		return nil
	}

	cfg, err := e.awsConfig(ctx)
	if err != nil {
		return fmt.Errorf("cloudwatch config: %w", err)
	}
	client := cloudwatchlogs.NewFromConfig(cfg)

	streamName := e.StreamName
	if streamName == "" {
		streamName = "functionfly-auth-" + time.Now().UTC().Format("2006-01-02")
	}

	_, _ = client.CreateLogStream(ctx, &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(e.LogGroup),
		LogStreamName: aws.String(streamName),
	})
	// Ignore error: ResourceAlreadyExistsException is expected when stream exists

	// Build log events with timestamps (ms); CloudWatch requires chronological order
	events := make([]cwtypes.InputLogEvent, 0, len(logs))
	for _, log := range logs {
		data, err := json.Marshal(log)
		if err != nil {
			continue
		}
		msg := string(data)
		if len(msg) > 1024*1024 {
			msg = msg[:1024*1024]
		}
		ts := time.Now().UnixMilli()
		if t, ok := log["timestamp"]; ok {
			switch v := t.(type) {
			case string:
				if parsed, err := time.Parse(time.RFC3339, v); err == nil {
					ts = parsed.UnixMilli()
				}
			case float64:
				ts = int64(v)
			}
		}
		events = append(events, cwtypes.InputLogEvent{
			Message:   aws.String(msg),
			Timestamp: aws.Int64(ts),
		})
	}
	sort.Slice(events, func(i, j int) bool { return *events[i].Timestamp < *events[j].Timestamp })

	const maxPerBatch = 10000
	var nextSeq *string
	for i := 0; i < len(events); i += maxPerBatch {
		end := i + maxPerBatch
		if end > len(events) {
			end = len(events)
		}
		batch := events[i:end]
		input := &cloudwatchlogs.PutLogEventsInput{
			LogGroupName:  aws.String(e.LogGroup),
			LogStreamName: aws.String(streamName),
			LogEvents:     batch,
		}
		if nextSeq != nil {
			input.SequenceToken = nextSeq
		}
		out, err := client.PutLogEvents(ctx, input)
		if err != nil {
			return fmt.Errorf("cloudwatch PutLogEvents: %w", err)
		}
		nextSeq = out.NextSequenceToken
	}
	return nil
}

func (e *CloudWatchExporter) TestConnection(ctx context.Context) error {
	cfg, err := e.awsConfig(ctx)
	if err != nil {
		return fmt.Errorf("cloudwatch config: %w", err)
	}
	client := cloudwatchlogs.NewFromConfig(cfg)
	_, err = client.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(e.LogGroup),
		Limit:              aws.Int32(1),
	})
	if err != nil {
		return fmt.Errorf("cloudwatch describe log groups: %w", err)
	}
	return nil
}

func (e *CloudWatchExporter) awsConfig(ctx context.Context) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(e.Region),
	}
	if e.AccessKeyID != "" && e.SecretKey != "" {
		opts = append(opts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			e.AccessKeyID, e.SecretKey, "",
		)))
	}
	return config.LoadDefaultConfig(ctx, opts...)
}

// ============== Azure Sentinel Exporter ==============

// AzureSentinelExporter exports logs to Azure Sentinel (Log Analytics)
type AzureSentinelExporter struct {
	WorkspaceID string
	APIKey      string
	TableName   string
	client      *http.Client
}

// NewAzureSentinelExporter creates a new Azure Sentinel exporter
func NewAzureSentinelExporter(config map[string]interface{}, client *http.Client) (*AzureSentinelExporter, error) {
	workspaceID, _ := config["workspace_id"].(string)
	apiKey, _ := config["api_key"].(string)
	tableName, _ := config["table_name"].(string)

	if workspaceID == "" || apiKey == "" {
		return nil, fmt.Errorf("workspace_id and api_key are required")
	}

	if tableName == "" {
		tableName = "FunctionFlyAuth_CL"
	}

	return &AzureSentinelExporter{
		WorkspaceID: workspaceID,
		APIKey:      apiKey,
		TableName:   tableName,
		client:      client,
	}, nil
}

func (e *AzureSentinelExporter) Export(ctx context.Context, logs []map[string]interface{}) error {
	// Azure Log Analytics API endpoint
	endpoint := fmt.Sprintf("https://%s.ods.opinsights.azure.com/api/logs?api-version=2016-04-01",
		e.WorkspaceID)

	// Convert logs to Azure format
	records := make([]map[string]interface{}, len(logs))
	for i, log := range logs {
		records[i] = log
		records[i]["TimeGenerated"] = log["timestamp"]
	}

	payload := map[string]interface{}{
		"records": records,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.APIKey)
	req.Header.Set("Log-Type", e.TableName)
	req.Header.Set("x-ms-date", time.Now().UTC().Format(http.TimeFormat))

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send to Azure: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Azure returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (e *AzureSentinelExporter) TestConnection(ctx context.Context) error {
	testLog := []map[string]interface{}{
		{
			"id":            uuid.New().String(),
			"event_type":    "siem_test",
			"success":       true,
			"timestamp":     time.Now().Format(time.RFC3339),
			"TimeGenerated": time.Now().Format(time.RFC3339),
		},
	}

	return e.Export(ctx, testLog)
}

// ============== GCP Chronicle Exporter ==============

// GCPChronicleExporter exports logs to GCP Chronicle
type GCPChronicleExporter struct {
	CustomerID string
	APIKey     string
	client     *http.Client
}

// NewGCPChronicleExporter creates a new GCP Chronicle exporter
func NewGCPChronicleExporter(config map[string]interface{}, client *http.Client) (*GCPChronicleExporter, error) {
	customerID, _ := config["customer_id"].(string)
	apiKey, _ := config["api_key"].(string)

	if customerID == "" || apiKey == "" {
		return nil, fmt.Errorf("customer_id and api_key are required")
	}

	return &GCPChronicleExporter{
		CustomerID: customerID,
		APIKey:     apiKey,
		client:     client,
	}, nil
}

func (e *GCPChronicleExporter) Export(ctx context.Context, logs []map[string]interface{}) error {
	// Chronicle Ingestion API
	endpoint := fmt.Sprintf("https://ingestion.googleapis.com/v1/uploads/%s/entries",
		e.CustomerID)

	// Convert to Chronicle format
	entries := make([]map[string]interface{}, len(logs))
	for i, log := range logs {
		entries[i] = map[string]interface{}{
			"timestamp": log["timestamp"],
			"metadata": map[string]interface{}{
				"event_type":   log["event_type"],
				"source_type":  "FUNCTIONFLY_AUTH",
				"vendor_name":  "FunctionFly",
				"product_name": "Auth",
			},
			"data": log,
		}
	}

	payload := map[string]interface{}{
		"entries": entries,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.APIKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send to Chronicle: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Chronicle returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (e *GCPChronicleExporter) TestConnection(ctx context.Context) error {
	testLog := []map[string]interface{}{
		{
			"id":         uuid.New().String(),
			"event_type": "siem_test",
			"success":    true,
			"timestamp":  time.Now().Format(time.RFC3339),
		},
	}

	return e.Export(ctx, testLog)
}

// ============== Splunk HEC Exporter ==============

// SplunkExporter exports logs to Splunk via HEC
type SplunkExporter struct {
	HECURL   string
	HECToken string
	Index    string
	Source   string
	client   *http.Client
}

// NewSplunkExporter creates a new Splunk exporter
func NewSplunkExporter(config map[string]interface{}, client *http.Client) (*SplunkExporter, error) {
	hecURL, _ := config["hec_url"].(string)
	hecToken, _ := config["hec_token"].(string)
	index, _ := config["index"].(string)
	source, _ := config["source"].(string)

	if hecURL == "" || hecToken == "" {
		return nil, fmt.Errorf("hec_url and hec_token are required")
	}

	if source == "" {
		source = "functionfly-auth"
	}

	return &SplunkExporter{
		HECURL:   hecURL,
		HECToken: hecToken,
		Index:    index,
		Source:   source,
		client:   client,
	}, nil
}

func (e *SplunkExporter) Export(ctx context.Context, logs []map[string]interface{}) error {
	for _, log := range logs {
		event := map[string]interface{}{
			"event":      log,
			"host":       "functionfly",
			"index":      e.Index,
			"source":     e.Source,
			"sourcetype": "functionfly:auth",
			"time":       log["timestamp"],
		}

		payload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.HECURL+"/services/collector", bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Splunk "+e.HECToken)

		resp, err := e.client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send to Splunk: %w", err)
		}
		resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("Splunk returned status %d", resp.StatusCode)
		}
	}

	return nil
}

func (e *SplunkExporter) TestConnection(ctx context.Context) error {
	testEvent := map[string]interface{}{
		"event": map[string]interface{}{
			"id":         uuid.New().String(),
			"event_type": "siem_test",
			"success":    true,
			"timestamp":  time.Now().Format(time.RFC3339),
		},
		"host":       "functionfly",
		"index":      e.Index,
		"source":     e.Source,
		"sourcetype": "functionfly:auth",
	}

	payload, _ := json.Marshal(testEvent)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.HECURL+"/services/collector", bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Splunk "+e.HECToken)

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Splunk connection test failed with status %d", resp.StatusCode)
	}

	return nil
}

// ============== Background Scheduler ==============

// StartScheduler starts the background SIEM export scheduler
// interval specifies how often to run exports (default: 5 minutes)
func (s *SIEMService) StartScheduler(interval time.Duration) {
	if s.runScheduled {
		logrus.Warn("SIEM scheduler already running")
		return
	}

	if interval <= 0 {
		interval = 5 * time.Minute
	}

	s.stopCh = make(chan struct{})
	s.runScheduled = true

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		logrus.Info("SIEM export scheduler started")

		// Run immediately on start
		s.runScheduledExport()

		for {
			select {
			case <-ticker.C:
				s.runScheduledExport()
			case <-s.stopCh:
				logrus.Info("SIEM export scheduler stopped")
				return
			}
		}
	}()
}

// StopScheduler stops the background SIEM export scheduler
func (s *SIEMService) StopScheduler() {
	if !s.runScheduled {
		return
	}

	close(s.stopCh)
	s.wg.Wait()
	s.runScheduled = false
	logrus.Info("SIEM export scheduler stopped")
}

// runScheduledExport runs exports for all enabled SIEM configs
func (s *SIEMService) runScheduledExport() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	configs, err := s.siemRepo.GetEnabled(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to get enabled SIEM configs")
		return
	}

	if len(configs) == 0 {
		logrus.Debug("No enabled SIEM configs found")
		return
	}

	logrus.Infof("Running SIEM export for %d enabled configs", len(configs))

	for _, config := range configs {
		s.runExportWithRetry(ctx, config.ID)
	}
}

// runExportWithRetry runs an export with retry logic
func (s *SIEMService) runExportWithRetry(ctx context.Context, configID uuid.UUID) {
	maxRetries := 3
	retryDelay := 5 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := s.Export(ctx, configID)
		if err == nil {
			logrus.Infof("SIEM export successful for config %s (attempt %d)", configID, attempt)
			return
		}

		logrus.WithError(err).Warnf("SIEM export failed for config %s (attempt %d/%d)",
			configID, attempt, maxRetries)

		if attempt < maxRetries {
			select {
			case <-ctx.Done():
				return
			case <-time.After(retryDelay):
			}
		}
	}

	logrus.Errorf("SIEM export failed after %d attempts for config %s", maxRetries, configID)
}

// ExportAllEnabled exports logs to all enabled SIEM configs
func (s *SIEMService) ExportAllEnabled(ctx context.Context) error {
	configs, err := s.siemRepo.GetEnabled(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled SIEM configs: %w", err)
	}

	if len(configs) == 0 {
		return nil
	}

	var lastErr error
	for _, config := range configs {
		if err := s.Export(ctx, config.ID); err != nil {
			logrus.WithError(err).Errorf("Failed to export to SIEM config %s", config.ID)
			lastErr = err
		}
	}

	return lastErr
}
