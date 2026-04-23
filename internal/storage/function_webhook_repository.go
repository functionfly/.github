package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FunctionWebhookRepository struct {
	db *gorm.DB
}

func NewFunctionWebhookRepository(db *gorm.DB) *FunctionWebhookRepository {
	return &FunctionWebhookRepository{db: db}
}

func (r *FunctionWebhookRepository) Create(ctx context.Context, sub *FunctionWebhookSubscription) error {
	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	if sub.CreatedAt.IsZero() {
		sub.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *FunctionWebhookRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*FunctionWebhookSubscription, error) {
	var sub FunctionWebhookSubscription
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&sub).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *FunctionWebhookRepository) List(ctx context.Context, tenantID uuid.UUID, functionID *uuid.UUID, limit, offset int) ([]*FunctionWebhookSubscription, int64, error) {
	var subs []*FunctionWebhookSubscription
	var total int64

	query := r.db.WithContext(ctx).Model(&FunctionWebhookSubscription{}).Where("tenant_id = ?", tenantID)

	if functionID != nil {
		query = query.Where("function_id = ? OR function_id IS NULL", functionID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&subs).Error; err != nil {
		return nil, 0, err
	}

	return subs, total, nil
}

func (r *FunctionWebhookRepository) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&FunctionWebhookSubscription{}).Error
}

func (r *FunctionWebhookRepository) Update(ctx context.Context, sub *FunctionWebhookSubscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

func (r *FunctionWebhookRepository) RecordDelivery(ctx context.Context, delivery *FunctionWebhookDelivery) error {
	if delivery.ID == uuid.Nil {
		delivery.ID = uuid.New()
	}
	if delivery.AttemptedAt.IsZero() {
		delivery.AttemptedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(delivery).Error
}

func (r *FunctionWebhookRepository) GetActiveSubscriptionsForEvent(ctx context.Context, eventType string, functionID uuid.UUID) ([]*FunctionWebhookSubscription, error) {
	var subs []*FunctionWebhookSubscription

	err := r.db.WithContext(ctx).
		Where("active = ?", true).
		Where("? = ANY(event_types)", eventType).
		Find(&subs).Error

	if err != nil {
		return nil, err
	}

	var filtered []*FunctionWebhookSubscription
	for _, sub := range subs {
		if sub.FunctionID == nil || *sub.FunctionID == functionID {
			filtered = append(filtered, sub)
		}
	}

	return filtered, nil
}

func (r *FunctionWebhookRepository) ListDeliveries(ctx context.Context, subscriptionID uuid.UUID, limit, offset int) ([]*FunctionWebhookDelivery, int64, error) {
	var deliveries []*FunctionWebhookDelivery
	var total int64

	query := r.db.WithContext(ctx).Model(&FunctionWebhookDelivery{}).Where("subscription_id = ?", subscriptionID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("attempted_at DESC").Limit(limit).Offset(offset).Find(&deliveries).Error; err != nil {
		return nil, 0, err
	}

	return deliveries, total, nil
}

func (r *FunctionWebhookRepository) GenerateSecret() string {
	secretBytes := make([]byte, 32)
	_, _ = rand.Read(secretBytes)
	return hex.EncodeToString(secretBytes)
}

func (r *FunctionWebhookRepository) SetActive(ctx context.Context, id uuid.UUID, active bool) error {
	return r.db.WithContext(ctx).Model(&FunctionWebhookSubscription{}).Where("id = ?", id).Update("active", active).Error
}
