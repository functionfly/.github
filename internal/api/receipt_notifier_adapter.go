// Package api — notifier adapter for the receipt milestone worker.
//
// Bridges the small `receipt.Notifier` interface (which the milestone
// worker depends on) to a direct GORM insert against the `notifications`
// table. We avoid notification.Service.Send here because that method
// passes a Go `[]string` to pgx which doesn't know how to bind it
// (`unsupported type []string, a slice of string`). The receipt feature
// is high-volume and best-effort — direct insert with a channels literal
// is the simplest, most reliable path.
package api

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	receiptpkg "github.com/functionfly/functionfly/internal/api/handlers/receipt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// receiptNotifierAdapter satisfies receipt.Notifier by inserting directly
// into the `notifications` table. nil-safe: if db is nil, every call
// returns nil so the milestone worker degrades to "log only".
type receiptNotifierAdapter struct {
	db *gorm.DB
}

func newReceiptNotifierAdapter(_ /* svc */ interface{}, db *gorm.DB) receiptpkg.Notifier {
	return &receiptNotifierAdapter{db: db}
}

// channelsLiteral converts a Go []string into a Postgres text[] literal
// (e.g. `{"in_app","email"}`). We build this as a string and cast in
// the SQL so pgx never sees a Go slice parameter.
func channelsLiteral(channels []string) string {
	if len(channels) == 0 {
		return "{}"
	}
	parts := make([]string, 0, len(channels))
	for _, c := range channels {
		parts = append(parts, `"`+strings.ReplaceAll(c, `"`, `\"`)+`"`)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

// dataJSON serialises a map[string]interface{} for the jsonb column. We
// use encoding/json so all the corner cases (escaping, UTF-8, etc.) are
// handled correctly. The notification table column is `data jsonb` with
// default `{}`; we explicitly pass `{}` (not NULL) when the map is
// empty so the notification row is always valid.
func dataJSON(data map[string]interface{}) string {
	if len(data) == 0 {
		return "{}"
	}
	b, err := json.Marshal(data)
	if err != nil {
		// Fall back to a minimal literal so the insert still goes through
		// (the dashboard will see a less-rich payload, but the milestone
		// notification is preserved).
		return "{}"
	}
	return string(b)
}

// Insert stores an in-app notification by writing directly to the
// `notifications` table with the `in_app` channel. The notification
// dispatcher picks it up on its next pass and renders it in the
// dashboard's notification list.
func (a *receiptNotifierAdapter) Insert(ctx context.Context, userID uuid.UUID, kind, title, body string, data map[string]interface{}) error {
	if a == nil || a.db == nil {
		return nil
	}
	if data == nil {
		data = map[string]interface{}{}
	}
	return a.insertRaw(ctx, a.db, userID, kind, title, body, data, []string{"in_app"})
}

// SendEmail dispatches an email by writing a notification row with the
// `email` channel. The notification dispatcher's email channel sends the
// actual email on its next pass (typically within seconds via the queue
// worker).
func (a *receiptNotifierAdapter) SendEmail(ctx context.Context, toUserID uuid.UUID, subject, htmlBody, plainBody string) error {
	if a == nil || a.db == nil {
		return nil
	}
	data := map[string]interface{}{
		"html_body": htmlBody,
		"subject":   subject,
	}
	return a.insertRaw(ctx, a.db, toUserID, "receipt_milestone_email", subject, plainBody, data, []string{"email"})
}

// insertRaw writes a single notification row. Returns nil on success so
// the milestone worker can record the channel as fired.
func (a *receiptNotifierAdapter) insertRaw(
	ctx context.Context,
	db *gorm.DB,
	userID uuid.UUID,
	kind, title, body string,
	data map[string]interface{},
	channels []string,
) error {
	now := time.Now()
	channelsLit := channelsLiteral(channels)
	dataLit := dataJSON(data)

	// Use raw SQL because GORM doesn't know how to bind Go []string for
	// the channels column (the same bug the notification repo has). We
	// cast the channels literal to text[] and the data literal to jsonb.
	res := db.WithContext(ctx).Exec(`
		INSERT INTO notifications
			(id, user_id, type, category, title, body, data, channels, priority, status, created_at, updated_at)
		VALUES (gen_random_uuid(), ?, ?, ?, ?, ?, ?::jsonb, ?::text[], 'low', 'pending', ?, ?)
	`, userID, kind, "receipt_milestone", title, body, dataLit, channelsLit, now, now)
	return res.Error
}
