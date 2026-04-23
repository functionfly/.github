package storage

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type FunctionWebhookService struct {
	repo   *FunctionWebhookRepository
	client *http.Client
	logger *logrus.Logger
}

func NewFunctionWebhookService(repo *FunctionWebhookRepository) *FunctionWebhookService {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	}

	return &FunctionWebhookService{
		repo: repo,
		client: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		logger: logrus.New(),
	}
}

func (s *FunctionWebhookService) SetLogger(logger *logrus.Logger) {
	s.logger = logger
}

func (s *FunctionWebhookService) TriggerFunctionEvent(ctx context.Context, eventType string, functionID uuid.UUID, payload interface{}) error {
	subs, err := s.repo.GetActiveSubscriptionsForEvent(ctx, eventType, functionID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get subscriptions for event")
		return err
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.logger.WithError(err).Error("Failed to marshal payload")
		return err
	}

	for _, sub := range subs {
		delivery := &FunctionWebhookDelivery{
			SubscriptionID: sub.ID,
			EventType:      eventType,
			Payload:        payloadBytes,
			Success:        false,
		}

		if err := s.repo.RecordDelivery(ctx, delivery); err != nil {
			s.logger.WithError(err).WithField("subscription_id", sub.ID).Error("Failed to record delivery")
			continue
		}

		go s.sendWebhook(delivery, sub, payloadBytes)
	}

	return nil
}

func (s *FunctionWebhookService) sendWebhook(delivery *FunctionWebhookDelivery, sub *FunctionWebhookSubscription, payload []byte) {
	start := time.Now()

	req, err := http.NewRequest("POST", sub.URL, bytes.NewReader(payload))
	if err != nil {
		s.handleDeliveryError(delivery, sub, err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "FunctionFly-Webhook/1.0")
	req.Header.Set("X-Delivery-ID", delivery.ID.String())
	req.Header.Set("X-Event-Type", delivery.EventType)

	signature := sub.GenerateSignature(payload)
	req.Header.Set("X-FunctionFly-Signature", "sha256="+signature)
	req.Header.Set("X-FunctionFly-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	resp, err := s.client.Do(req)
	responseTime := int(time.Since(start).Milliseconds())

	if err != nil {
		s.handleDeliveryError(delivery, sub, err.Error())
		return
	}
	defer resp.Body.Close()

	var responseBody bytes.Buffer
	_, _ = responseBody.ReadFrom(resp.Body)
	bodyStr := responseBody.String()
	if len(bodyStr) > 10000 {
		bodyStr = bodyStr[:10000] + "... [truncated]"
	}

	isSuccess := resp.StatusCode >= 200 && resp.StatusCode < 300

	statusCode := resp.StatusCode
	delivery.Success = isSuccess
	delivery.ResponseStatus = &statusCode
	delivery.ResponseBody = &bodyStr

	if isSuccess {
		s.logger.WithFields(logrus.Fields{
			"subscription_id":  sub.ID,
			"delivery_id":      delivery.ID,
			"status_code":      resp.StatusCode,
			"response_time_ms": responseTime,
		}).Info("Function webhook delivered successfully")
	} else {
		errMsg := fmt.Sprintf("HTTP %d: %s", resp.StatusCode, bodyStr)
		s.handleDeliveryError(delivery, sub, errMsg)
	}
}

func (s *FunctionWebhookService) handleDeliveryError(delivery *FunctionWebhookDelivery, sub *FunctionWebhookSubscription, errMsg string) {
	delivery.Success = false
	errMsgCopy := errMsg
	delivery.ErrorMessage = &errMsgCopy

	s.logger.WithFields(logrus.Fields{
		"subscription_id": sub.ID,
		"delivery_id":     delivery.ID,
		"error":           errMsg,
	}).Error("Function webhook delivery failed")

	_ = s.repo.RecordDelivery(context.Background(), delivery)
}

func (s *FunctionWebhookService) ValidateURL(urlStr string) error {
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

	host := parsedURL.Hostname()
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return fmt.Errorf("webhook URL cannot point to localhost")
	}

	if isPrivateIP(host) {
		return fmt.Errorf("webhook URL cannot point to private IP addresses")
	}

	return nil
}

func isPrivateIP(host string) bool {
	privatePrefixes := []string{
		"10.",
		"172.16.", "172.17.", "172.18.", "172.19.",
		"172.20.", "172.21.", "172.22.", "172.23.",
		"172.24.", "172.25.", "172.26.", "172.27.",
		"172.28.", "172.29.", "172.30.", "172.31.",
		"192.168.",
		"127.",
		"0.",
		"::1",
		"fc00:",
		"fe80:",
	}

	for _, prefix := range privatePrefixes {
		if len(host) >= len(prefix) && host[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}

func (s *FunctionWebhookService) TestWebhook(sub *FunctionWebhookSubscription, eventType string, testData map[string]interface{}) (*FunctionWebhookTestResponse, error) {
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

	req, err := http.NewRequest("POST", sub.URL, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "FunctionFly-Webhook/1.0")
	req.Header.Set("X-Delivery-ID", "del_test_"+uuid.New().String()[:24])
	req.Header.Set("X-Event-Type", eventType)
	req.Header.Set("X-FunctionFly-Test", "true")

	signature := sub.GenerateSignature(payloadBytes)
	req.Header.Set("X-FunctionFly-Signature", "sha256="+signature)
	req.Header.Set("X-FunctionFly-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	start := time.Now()
	resp, err := s.client.Do(req)
	responseTime := int(time.Since(start).Milliseconds())

	if err != nil {
		return &FunctionWebhookTestResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}
	defer resp.Body.Close()

	var responseBody bytes.Buffer
	_, _ = responseBody.ReadFrom(resp.Body)
	bodyStr := responseBody.String()
	if len(bodyStr) > 10000 {
		bodyStr = bodyStr[:10000] + "... [truncated]"
	}

	isSuccess := resp.StatusCode >= 200 && resp.StatusCode < 300

	return &FunctionWebhookTestResponse{
		Success:        isSuccess,
		StatusCode:     resp.StatusCode,
		ResponseTimeMs: responseTime,
		ResponseBody:   bodyStr,
	}, nil
}
