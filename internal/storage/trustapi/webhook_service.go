package trustapi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// WebhookService handles webhook delivery with retry logic
type WebhookService struct {
	repo   *WebhookRepository
	client *http.Client
	logger *logrus.Logger
}

// NewWebhookService creates a new webhook service
func NewWebhookService(repo *WebhookRepository) *WebhookService {
	// Create HTTP client with reasonable defaults
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	}

	return &WebhookService{
		repo: repo,
		client: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		logger: logrus.New(),
	}
}

// SetLogger sets the logger for the webhook service
func (s *WebhookService) SetLogger(logger *logrus.Logger) {
	s.logger = logger
}

// TriggerEvent triggers webhook deliveries for a trust event
func (s *WebhookService) TriggerEvent(
	eventType WebhookEventType,
	functionID *uuid.UUID,
	data map[string]interface{},
) error {
	// Get webhooks subscribed to this event
	var webhooks []TrustWebhook
	var err error

	if functionID != nil {
		webhooks, err = s.repo.GetActiveWebhooksForEventAndFunction(eventType, *functionID)
	} else {
		webhooks, err = s.repo.GetActiveWebhooksForEvent(eventType)
	}

	if err != nil {
		s.logger.WithError(err).Error("Failed to get webhooks for event")
		return err
	}

	// Create payload
	payload := NewWebhookPayload(eventType, data)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.logger.WithError(err).Error("Failed to marshal webhook payload")
		return err
	}

	// Create delivery records and send
	for _, webhook := range webhooks {
		delivery := &WebhookDelivery{
			WebhookID:     webhook.ID,
			EventType:     string(eventType),
			Payload:       payloadBytes,
			AttemptNumber: 1,
			MaxAttempts:   webhook.MaxRetries + 1,
			Status:        string(WebhookDeliveryStatusPending),
		}

		// Add entity ID if present in data
		if entityID, ok := data["id"]; ok {
			if idStr, ok := entityID.(string); ok {
				delivery.EntityID = idStr
			}
		} else if entityID, ok := data["revocation_id"]; ok {
			if idStr, ok := entityID.(string); ok {
				delivery.EntityID = idStr
			}
		} else if entityID, ok := data["attestation_id"]; ok {
			if idStr, ok := entityID.(string); ok {
				delivery.EntityID = idStr
			}
		} else if entityID, ok := data["policy_id"]; ok {
			if idStr, ok := entityID.(string); ok {
				delivery.EntityID = idStr
			}
		}

		if err := s.repo.CreateDelivery(delivery); err != nil {
			s.logger.WithError(err).WithField("webhook_id", webhook.ID).Error("Failed to create delivery record")
			continue
		}

		// Send immediately (async in production would use a queue)
		go s.sendDelivery(delivery, &webhook, payloadBytes)
	}

	return nil
}

// sendDelivery sends a webhook delivery
func (s *WebhookService) sendDelivery(delivery *WebhookDelivery, webhook *TrustWebhook, payload []byte) {
	start := time.Now()

	// Create request
	method := webhook.Method
	if method == "" {
		method = "POST"
	}

	req, err := http.NewRequest(method, webhook.URL, bytes.NewReader(payload))
	if err != nil {
		s.handleDeliveryError(delivery, webhook, err, 0)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "FunctionFly-Webhook/1.0")
	req.Header.Set("X-Webhook-ID", webhook.WebhookID)
	req.Header.Set("X-Delivery-ID", delivery.DeliveryID)
	req.Header.Set("X-Event-Type", delivery.EventType)

	// Add signature
	signature := webhook.GenerateSignature(payload)
	req.Header.Set("X-Webhook-Signature", "sha256="+signature)
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	// Add custom headers
	if len(webhook.CustomHeaders) > 0 {
		var headers map[string]string
		if err := json.Unmarshal(webhook.CustomHeaders, &headers); err == nil {
			for key, value := range headers {
				req.Header.Set(key, value)
			}
		}
	}

	// Send request
	resp, err := s.client.Do(req)
	responseTime := int(time.Since(start).Milliseconds())

	if err != nil {
		s.handleDeliveryError(delivery, webhook, err, responseTime)
		return
	}
	defer resp.Body.Close()

	// Read response body (limited to prevent memory issues)
	var responseBody bytes.Buffer
	_, _ = responseBody.ReadFrom(resp.Body)
	bodyStr := responseBody.String()
	if len(bodyStr) > 10000 {
		bodyStr = bodyStr[:10000] + "... [truncated]"
	}

	// Determine success
	isSuccess := resp.StatusCode >= 200 && resp.StatusCode < 300

	// Update delivery record
	if isSuccess {
		s.repo.UpdateDeliveryStatus(
			delivery.ID,
			WebhookDeliveryStatusDelivered,
			&resp.StatusCode,
			nil, // headers too large to store
			bodyStr,
			responseTime,
			"",
		)
		s.repo.RecordWebhookSuccess(webhook.ID)

		s.logger.WithFields(logrus.Fields{
			"webhook_id":       webhook.WebhookID,
			"delivery_id":      delivery.DeliveryID,
			"status_code":      resp.StatusCode,
			"response_time_ms": responseTime,
		}).Info("Webhook delivered successfully")
	} else {
		// Failed - schedule retry if attempts remain
		if delivery.AttemptNumber < delivery.MaxAttempts {
			s.scheduleRetry(delivery, webhook, resp.StatusCode, bodyStr, responseTime)
		} else {
			// Max retries reached
			s.repo.UpdateDeliveryStatus(
				delivery.ID,
				WebhookDeliveryStatusFailed,
				&resp.StatusCode,
				nil,
				bodyStr,
				responseTime,
				fmt.Sprintf("HTTP %d: %s", resp.StatusCode, bodyStr),
			)
			s.repo.RecordWebhookFailure(webhook.ID, fmt.Sprintf("HTTP %d", resp.StatusCode))

			s.logger.WithFields(logrus.Fields{
				"webhook_id":     webhook.WebhookID,
				"delivery_id":    delivery.DeliveryID,
				"status_code":    resp.StatusCode,
				"response_body":  bodyStr,
				"attempt_number": delivery.AttemptNumber,
				"max_attempts":   delivery.MaxAttempts,
			}).Warn("Webhook delivery failed permanently")
		}
	}
}

// handleDeliveryError handles delivery errors (network, timeout, etc.)
func (s *WebhookService) handleDeliveryError(
	delivery *WebhookDelivery,
	webhook *TrustWebhook,
	err error,
	responseTimeMs int,
) {
	errorMsg := err.Error()

	if delivery.AttemptNumber < delivery.MaxAttempts {
		s.scheduleRetry(delivery, webhook, 0, errorMsg, responseTimeMs)
	} else {
		s.repo.UpdateDeliveryStatus(
			delivery.ID,
			WebhookDeliveryStatusFailed,
			nil,
			nil,
			errorMsg,
			responseTimeMs,
			errorMsg,
		)
		s.repo.RecordWebhookFailure(webhook.ID, errorMsg)

		s.logger.WithFields(logrus.Fields{
			"webhook_id":     webhook.WebhookID,
			"delivery_id":    delivery.DeliveryID,
			"error":          errorMsg,
			"attempt_number": delivery.AttemptNumber,
		}).Error("Webhook delivery failed permanently")
	}
}

// scheduleRetry schedules a retry for a failed delivery
func (s *WebhookService) scheduleRetry(
	delivery *WebhookDelivery,
	webhook *TrustWebhook,
	statusCode int,
	responseBody string,
	responseTimeMs int,
) {
	// Update delivery status to retrying
	nextAttempt := delivery.AttemptNumber + 1
	statusCodePtr := &statusCode
	if statusCode == 0 {
		statusCodePtr = nil
	}

	s.repo.UpdateDeliveryStatus(
		delivery.ID,
		WebhookDeliveryStatusRetrying,
		statusCodePtr,
		nil,
		responseBody,
		responseTimeMs,
		fmt.Sprintf("Attempt %d failed, scheduling retry", delivery.AttemptNumber),
	)

	// Calculate retry delay with exponential backoff
	delay := webhook.RetryDelaySecs * (1 << (nextAttempt - 2)) // 2^attempt exponential backoff
	if delay > 3600 {                                          // Max 1 hour
		delay = 3600
	}

	s.repo.ScheduleRetry(delivery.ID, nextAttempt, delay)

	s.logger.WithFields(logrus.Fields{
		"webhook_id":      webhook.WebhookID,
		"delivery_id":     delivery.DeliveryID,
		"next_attempt":    nextAttempt,
		"retry_delay_sec": delay,
		"max_attempts":    delivery.MaxAttempts,
	}).Info("Webhook retry scheduled")
}

// ProcessRetries processes pending retries (should be called by a background worker)
func (s *WebhookService) ProcessRetries(limit int) error {
	deliveries, err := s.repo.ListRetriesDue(limit)
	if err != nil {
		return err
	}

	for _, delivery := range deliveries {
		webhook, err := s.repo.GetWebhookByID(delivery.WebhookID)
		if err != nil {
			s.logger.WithError(err).WithField("webhook_id", delivery.WebhookID).Error("Failed to get webhook for retry")
			continue
		}

		// Mark as sent for this attempt
		s.repo.UpdateDeliveryStatus(
			delivery.ID,
			WebhookDeliveryStatusSent,
			nil,
			nil,
			"",
			0,
			"",
		)

		// Send retry
		go s.sendDelivery(&delivery, webhook, delivery.Payload)
	}

	return nil
}

// ProcessPendingDeliveries processes pending deliveries (should be called by a background worker)
func (s *WebhookService) ProcessPendingDeliveries(limit int) error {
	deliveries, err := s.repo.ListPendingDeliveries(limit)
	if err != nil {
		return err
	}

	for _, delivery := range deliveries {
		webhook, err := s.repo.GetWebhookByID(delivery.WebhookID)
		if err != nil {
			s.logger.WithError(err).WithField("webhook_id", delivery.WebhookID).Error("Failed to get webhook for delivery")
			continue
		}

		// Mark as sent
		s.repo.UpdateDeliveryStatus(
			delivery.ID,
			WebhookDeliveryStatusSent,
			nil,
			nil,
			"",
			0,
			"",
		)

		// Send
		go s.sendDelivery(&delivery, webhook, delivery.Payload)
	}

	return nil
}

// TestWebhook tests a webhook endpoint with a test payload
func (s *WebhookService) TestWebhook(webhook *TrustWebhook, eventType string, testData map[string]interface{}) (*WebhookTestResponse, error) {
	// Create test payload
	payload := map[string]interface{}{
		"event_id":    "evt_test_" + uuid.New().String()[:24],
		"event_type":  eventType,
		"timestamp":   time.Now(),
		"api_version": "2024-04-12",
		"test":        true,
		"data":        testData,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Create request
	method := webhook.Method
	if method == "" {
		method = "POST"
	}

	req, err := http.NewRequest(method, webhook.URL, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "FunctionFly-Webhook/1.0")
	req.Header.Set("X-Webhook-ID", webhook.WebhookID)
	req.Header.Set("X-Delivery-ID", "del_test_"+uuid.New().String()[:24])
	req.Header.Set("X-Event-Type", eventType)
	req.Header.Set("X-Webhook-Test", "true")

	// Add signature
	signature := webhook.GenerateSignature(payloadBytes)
	req.Header.Set("X-Webhook-Signature", "sha256="+signature)
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	// Add custom headers
	if len(webhook.CustomHeaders) > 0 {
		var headers map[string]string
		if err := json.Unmarshal(webhook.CustomHeaders, &headers); err == nil {
			for key, value := range headers {
				req.Header.Set(key, value)
			}
		}
	}

	// Send with timeout
	start := time.Now()
	resp, err := s.client.Do(req)
	responseTime := int(time.Since(start).Milliseconds())

	if err != nil {
		return &WebhookTestResponse{
			Success:    false,
			Error:      err.Error(),
			DeliveryID: req.Header.Get("X-Delivery-ID"),
		}, nil
	}
	defer resp.Body.Close()

	// Read response
	var responseBody bytes.Buffer
	_, _ = responseBody.ReadFrom(resp.Body)
	bodyStr := responseBody.String()
	if len(bodyStr) > 10000 {
		bodyStr = bodyStr[:10000] + "... [truncated]"
	}

	isSuccess := resp.StatusCode >= 200 && resp.StatusCode < 300

	return &WebhookTestResponse{
		Success:        isSuccess,
		StatusCode:     resp.StatusCode,
		ResponseTimeMs: responseTime,
		ResponseBody:   bodyStr,
		DeliveryID:     req.Header.Get("X-Delivery-ID"),
	}, nil
}

// ValidateWebhookURL validates a webhook URL
func (s *WebhookService) ValidateWebhookURL(urlStr string) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("webhook URL must use HTTPS for security")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("webhook URL must have a valid host")
	}

	// Check for internal/private IP ranges
	host := parsedURL.Hostname()
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return fmt.Errorf("webhook URL cannot point to localhost")
	}

	// Check for private IP ranges
	if isPrivateIP(host) {
		return fmt.Errorf("webhook URL cannot point to private IP addresses")
	}

	return nil
}

// isPrivateIP checks if an IP is in a private range
func isPrivateIP(host string) bool {
	ip := net.ParseIP(host)
	if ip == nil {
		// Not an IP address - could be a hostname
		// Try to resolve it to check the resolved IP
		ips, err := net.LookupIP(host)
		if err != nil || len(ips) == 0 {
			return false
		}
		ip = ips[0]
	}

	return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}
