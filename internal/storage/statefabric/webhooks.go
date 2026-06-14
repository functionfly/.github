package statefabric

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/functionfly/functionfly/internal/monitoring"
)

type StateFabricWebhook struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FabricID      uuid.UUID `gorm:"type:uuid;not null;index"`
	Name          string    `gorm:"size:255"`
	URL           string    `gorm:"type:text;not null"`
	Secret        string    `gorm:"size:255"`
	EventTypes    JSONMap   `gorm:"type:jsonb"`
	Active        bool      `gorm:"default:true"`
	Headers       JSONMap   `gorm:"type:jsonb"`
	RetryPolicy   JSONMap   `gorm:"type:jsonb"`
	LastTriggered *time.Time `gorm:"column:last_triggered_at"`
	LastStatus    string    `gorm:"size:50"`
	LastError     string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (StateFabricWebhook) TableName() string { return "state_fabric_webhooks" }

func (s *StateFabricWebhook) BeforeCreate(tx interface{}) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type WebhookDelivery struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	WebhookID  uuid.UUID `gorm:"type:uuid;not null;index"`
	Status     string    `gorm:"size:50;default:pending"`
	Request    JSONMap   `gorm:"type:jsonb"`
	Response   JSONMap   `gorm:"type:jsonb"`
	StatusCode int       `gorm:"default:0"`
	Attempts   int       `gorm:"default:0"`
	Error      string    `gorm:"type:text"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func (WebhookDelivery) TableName() string { return "state_fabric_webhook_deliveries" }

func (s *WebhookDelivery) BeforeCreate(tx interface{}) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

const (
	EventKeySet           = "key_set"
	EventKeyDelete        = "key_delete"
	EventFabricUpdate     = "fabric_update"
	EventPipelineComplete  = "pipeline_complete"
	EventSnapshotCreated   = "snapshot_created"
	EventReplayComplete   = "replay_complete"
)

type WebhookPayload struct {
	EventType  string                 `json:"event_type"`
	Timestamp  time.Time              `json:"timestamp"`
	FabricID  uuid.UUID              `json:"fabric_id"`
	TenantID  uuid.UUID              `json:"tenant_id"`
	Data      map[string]interface{} `json:"data"`
	DeliveryID uuid.UUID              `json:"delivery_id,omitempty"`
}

func (r *Repository) CreateWebhook(ctx context.Context, fabricID uuid.UUID, name, url, secret string, eventTypes []string, headers map[string]interface{}) (*StateFabricWebhook, error) {
	eventTypesMap := JSONMap{}
	for _, et := range eventTypes {
		eventTypesMap[et] = true
	}

	webhook := &StateFabricWebhook{
		ID:         uuid.New(),
		FabricID:   fabricID,
		Name:       name,
		URL:        url,
		Secret:     secret,
		EventTypes: eventTypesMap,
		Active:     true,
		Headers:    JSONMap(headers),
		RetryPolicy: map[string]interface{}{
			"max_attempts":    3,
			"retry_delay_ms": 1000,
		},
	}

	if err := r.db.WithContext(ctx).Create(webhook).Error; err != nil {
		return nil, err
	}

	return webhook, nil
}

func (r *Repository) ListWebhooks(ctx context.Context, tenantID, fabricID uuid.UUID) ([]StateFabricWebhook, error) {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return nil, err
	}
	if state.TenantID != tenantID {
		return nil, fmt.Errorf("state fabric not found")
	}

	var webhooks []StateFabricWebhook
	if err := r.db.WithContext(ctx).Where("fabric_id = ?", fabricID).Find(&webhooks).Error; err != nil {
		return nil, err
	}

	return webhooks, nil
}

func (r *Repository) DeleteWebhook(ctx context.Context, tenantID, fabricID, webhookID uuid.UUID) error {
	state, err := r.stateRepo.GetStateByID(ctx, fabricID)
	if err != nil {
		return err
	}
	if state.TenantID != tenantID {
		return fmt.Errorf("state fabric not found")
	}

	return r.db.WithContext(ctx).Delete(&StateFabricWebhook{}, "id = ? AND fabric_id = ?", webhookID, fabricID).Error
}

func (r *Repository) TriggerWebhooks(ctx context.Context, fabricID uuid.UUID, eventType string, data map[string]interface{}) error {
	webhooks, err := r.ListWebhooks(ctx, uuid.Nil, fabricID)
	if err != nil {
		return err
	}

	tenantID := uuid.Nil
	state, stateErr := r.stateRepo.GetStateByID(ctx, fabricID)
	if stateErr == nil {
		tenantID = state.TenantID
	}

	for _, webhook := range webhooks {
		if !webhook.Active {
			continue
		}

		found := false
		for et := range webhook.EventTypes {
			if et == eventType || et == "*" {
				found = true
				break
			}
		}

		if !found {
			continue
		}

		go r.deliverWebhook(webhook, fabricID, tenantID, eventType, data)
	}

	return nil
}

func (r *Repository) deliverWebhook(webhook StateFabricWebhook, fabricID, tenantID uuid.UUID, eventType string, data map[string]interface{}) {
	ctx := context.Background()
	deliveryID := uuid.New()

	payload := WebhookPayload{
		EventType:  eventType,
		Timestamp:  time.Now(),
		FabricID:  fabricID,
		TenantID:  tenantID,
		Data:      data,
		DeliveryID: deliveryID,
	}

	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewReader(payloadBytes))
	if err != nil {
		logrus.WithError(err).WithField("webhook_id", webhook.ID).Error("failed to create webhook request")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-StateFabric-Event", eventType)
	req.Header.Set("X-StateFabric-Delivery-ID", deliveryID.String())

	if webhook.Secret != "" {
		signature := computeHMAC(payloadBytes, webhook.Secret)
		req.Header.Set("X-StateFabric-Signature", signature)
	}

	for k, v := range webhook.Headers {
		if s, ok := v.(string); ok {
			req.Header.Set(k, s)
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)

	statusCode := 0
	responseBody := map[string]interface{}{}
	if err == nil {
		statusCode = resp.StatusCode
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(&responseBody)
	}

	delivery := &WebhookDelivery{
		ID:         deliveryID,
		WebhookID:  webhook.ID,
		Status:     "success",
		Request:    data,
		Response:   responseBody,
		StatusCode: statusCode,
		Attempts:   1,
	}
	if err != nil {
		delivery.Status = "failed"
		delivery.Error = err.Error()
	} else if statusCode >= 400 {
		delivery.Status = "failed"
		delivery.Error = fmt.Sprintf("HTTP %d", statusCode)
	}

	r.db.WithContext(ctx).Create(delivery)

	now := time.Now()
	update := map[string]interface{}{
		"last_triggered_at": now,
		"last_status":      delivery.Status,
	}
	if delivery.Error != "" {
		update["last_error"] = delivery.Error
	}
	r.db.WithContext(ctx).Model(&webhook).Updates(update)

	if delivery.Status == "success" {
		monitoring.RecordStateFabricWebhookInvocation(tenantID.String(), fabricID.String(), eventType, "success")
	} else {
		monitoring.RecordStateFabricWebhookInvocation(tenantID.String(), fabricID.String(), eventType, "failed")
	}
}

func computeHMAC(message []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}
