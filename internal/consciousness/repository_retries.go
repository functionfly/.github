package consciousness

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DeliveryRetry represents a retry entry in the dead letter queue.
type DeliveryRetry struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	InsightID   uuid.UUID       `json:"insight_id" db:"insight_id"`
	TenantID    uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Channel     string          `json:"channel" db:"channel"`
	Payload     json.RawMessage `json:"payload" db:"payload"`
	AttemptCount int            `json:"attempt_count" db:"attempt_count"`
	MaxAttempts int             `json:"max_attempts" db:"max_attempts"`
	NextRetryAt time.Time       `json:"next_retry_at" db:"next_retry_at"`
	LastError   *string         `json:"last_error,omitempty" db:"last_error"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`

	// Joined fields
	Insight *Insight `json:"insight,omitempty" db:"-"`
}

// ScheduleRetry adds a failed delivery to the retry queue.
func (r *Repository) ScheduleRetry(ctx context.Context, retry *DeliveryRetry) error {
	query := `
		INSERT INTO consciousness_delivery_retries (
			insight_id, tenant_id, channel, payload, attempt_count, max_attempts, next_retry_at, last_error
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		retry.InsightID, retry.TenantID, retry.Channel, retry.Payload,
		retry.AttemptCount, retry.MaxAttempts, retry.NextRetryAt, retry.LastError,
	).Scan(&retry.ID, &retry.CreatedAt, &retry.UpdatedAt)
}

// GetDueRetries retrieves retries that are due for processing.
func (r *Repository) GetDueRetries(ctx context.Context, limit int) ([]*DeliveryRetry, error) {
	query := `
		SELECT id, insight_id, tenant_id, channel, payload, attempt_count, max_attempts, 
		       next_retry_at, last_error, created_at, updated_at
		FROM consciousness_delivery_retries
		WHERE next_retry_at <= NOW()
		  AND attempt_count < max_attempts
		ORDER BY next_retry_at ASC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get due retries: %w", err)
	}
	defer rows.Close()

	var retries []*DeliveryRetry
	for rows.Next() {
		retry := &DeliveryRetry{}
		var payload []byte
		err := rows.Scan(
			&retry.ID, &retry.InsightID, &retry.TenantID, &retry.Channel, &payload,
			&retry.AttemptCount, &retry.MaxAttempts, &retry.NextRetryAt, &retry.LastError,
			&retry.CreatedAt, &retry.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan retry: %w", err)
		}
		retry.Payload = payload
		retries = append(retries, retry)
	}

	return retries, rows.Err()
}

// UpdateRetryNextAttempt updates a retry for the next attempt with backoff.
func (r *Repository) UpdateRetryNextAttempt(ctx context.Context, id uuid.UUID, attemptCount int, delay time.Duration, lastError string) error {
	query := `
		UPDATE consciousness_delivery_retries
		SET attempt_count = $2,
		    next_retry_at = NOW() + $3::interval,
		    last_error = $4,
		    updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, attemptCount, delay.String(), lastError)
	return err
}

// MarkRetryCompleted marks a retry as successfully completed.
func (r *Repository) MarkRetryCompleted(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM consciousness_delivery_retries WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// MarkRetryExhausted marks a retry as failed after all attempts exhausted.
func (r *Repository) MarkRetryExhausted(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE consciousness_delivery_retries
		SET next_retry_at = NULL, updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetRetryQueueSize returns the current size of the retry queue.
func (r *Repository) GetRetryQueueSize(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM consciousness_delivery_retries WHERE next_retry_at <= NOW()`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// CleanupOldRetries removes retry entries older than the specified days.
func (r *Repository) CleanupOldRetries(ctx context.Context, olderThanDays int) (int64, error) {
	query := `
		DELETE FROM consciousness_delivery_retries
		WHERE next_retry_at IS NULL
		  AND updated_at < NOW() - INTERVAL '%d days'`
	_, err := r.db.ExecContext(ctx, fmt.Sprintf(query, olderThanDays))
	if err != nil {
		return 0, err
	}
	// Get affected rows from a separate count query
	countQuery := `SELECT COUNT(*) FROM consciousness_delivery_retries WHERE next_retry_at IS NULL`
	var count int64
	r.db.QueryRowContext(ctx, countQuery).Scan(&count)
	return count, nil
}
