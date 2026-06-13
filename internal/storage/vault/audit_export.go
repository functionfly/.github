package vault

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AuditExportFormat selects the output format for an export request.
type AuditExportFormat string

const (
	AuditExportFormatJSON AuditExportFormat = "json"
	AuditExportFormatCSV  AuditExportFormat = "csv"
	AuditExportFormatCEF  AuditExportFormat = "cef"
)

// AuditExportQuery is the input to ExportAudit. The caller may pass
// only the From / To timestamps and the desired Format.
type AuditExportQuery struct {
	TenantID uuid.UUID
	From     time.Time
	To       time.Time
	Format   AuditExportFormat
	SecretID *uuid.UUID
	Action   string
}

// AuditExportResult is the response from ExportAudit.
type AuditExportResult struct {
	Format    AuditExportFormat `json:"format"`
	Body      []byte            `json:"-"`
	RowCount  int               `json:"row_count"`
	Generated time.Time         `json:"generated_at"`
	// HMAC is the SHA-256 HMAC of the export body keyed by the
	// tenant's "audit-export" secret. The caller can verify
	// integrity by computing HMAC-SHA-256(Body, secret) and
	// comparing to this base64 value.
	HMAC string `json:"hmac_sha256"`
}

// ExportAudit streams the requested audit log slice in the chosen
// format. The function is bounded by the repository's
// ListAuditLogsByTenant, which is sufficient for the typical
// export window (≤ 1 year) when chunked.
func (r *Repository) ExportAudit(ctx context.Context, q AuditExportQuery, signingKey []byte) (*AuditExportResult, error) {
	if q.Format == "" {
		q.Format = AuditExportFormatJSON
	}
	rows, err := r.listAuditForExport(ctx, q)
	if err != nil {
		return nil, err
	}
	var body []byte
	switch q.Format {
	case AuditExportFormatJSON:
		body, err = encodeAuditJSON(rows)
	case AuditExportFormatCSV:
		body, err = encodeAuditCSV(rows)
	case AuditExportFormatCEF:
		body, err = encodeAuditCEF(rows)
	default:
		return nil, fmt.Errorf("vault: unsupported export format %q", q.Format)
	}
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, signingKey)
	mac.Write(body)
	digest := mac.Sum(nil)
	return &AuditExportResult{
		Format:    q.Format,
		Body:      body,
		RowCount:  len(rows),
		Generated: time.Now().UTC(),
		HMAC:      base64.StdEncoding.EncodeToString(digest),
	}, nil
}

// listAuditForExport runs the export query and returns the raw audit
// rows. We keep this separate so the encoders can be unit-tested.
func (r *Repository) listAuditForExport(ctx context.Context, q AuditExportQuery) ([]AuditLog, error) {
	query := r.db.WithContext(ctx).Where("tenant_id = ?", q.TenantID)
	if !q.From.IsZero() {
		query = query.Where("created_at >= ?", q.From)
	}
	if !q.To.IsZero() {
		query = query.Where("created_at <= ?", q.To)
	}
	if q.SecretID != nil {
		query = query.Where("secret_id = ?", *q.SecretID)
	}
	if q.Action != "" {
		query = query.Where("action = ?", q.Action)
	}
	var rows []AuditLog
	if err := query.Order("created_at ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func encodeAuditJSON(rows []AuditLog) ([]byte, error) {
	out := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		out[i] = auditRowToMap(r)
	}
	return json.MarshalIndent(map[string]interface{}{
		"format":  "json",
		"version": 1,
		"count":   len(rows),
		"rows":    out,
	}, "", "  ")
}

func encodeAuditCSV(rows []AuditLog) ([]byte, error) {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)
	if err := w.Write([]string{
		"id", "created_at", "action", "actor_id", "actor_type",
		"secret_id", "ip_address", "user_agent", "success",
		"error_message", "metadata_json",
	}); err != nil {
		return nil, err
	}
	for _, r := range rows {
		metaJSON, _ := json.Marshal(r.Metadata)
		_ = w.Write([]string{
			r.ID.String(),
			r.CreatedAt.UTC().Format(time.RFC3339Nano),
			string(r.Action),
			r.ActorID,
			string(r.ActorType),
			optionalUUID(r.SecretID),
			r.IPAddress,
			r.UserAgent,
			strconv.FormatBool(r.Success),
			r.ErrorMessage,
			string(metaJSON),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// encodeAuditCEF renders one row per line in Common Event Format
// (ArcSight / Splunk-friendly). See
// https://community.microfocus.com/dcvta051s/attachments/dcvta051s/Connector-Documentation/173/2/CommonEventFormat_v25.pdf
func encodeAuditCEF(rows []AuditLog) ([]byte, error) {
	buf := &bytes.Buffer{}
	for _, r := range rows {
		// Header: CEF:Version|Device Vendor|Device Product|Device Version|Signature ID|Name|Severity|Extension
		ts := r.CreatedAt.UTC().UnixMilli()
		sigID := sanitizeCEF("vault." + string(r.Action))
		name := fmt.Sprintf("vault %s by %s", r.Action, r.ActorType)
		severity := 5 // informational
		if !r.Success {
			severity = 7 // medium
		}
		ext := buildCEFExtension(map[string]string{
			"act":      r.ActorID,
			"cs1":      string(r.ActorType),
			"cs1Label": "actorType",
			"src":      r.IPAddress,
			"request":  r.RequestID,
			"cat":      "vault",
			"outcome":  ifThen(r.Success, "success", "failure"),
			"msg":      r.ErrorMessage,
			"rt":       strconv.FormatInt(ts, 10),
		})
		// Escape pipe and backslash in the extension field.
		ext = escapeCEF(ext)
		fmt.Fprintf(buf, "CEF:0|FunctionFly|Vault|1|%s|%s|%d|%s\n", sigID, name, severity, ext)
	}
	return buf.Bytes(), nil
}

func sanitizeCEF(s string) string {
	r := strings.NewReplacer("\\", "\\\\", "|", "\\|", "\n", " ", "\r", " ")
	return r.Replace(s)
}

func escapeCEF(s string) string {
	r := strings.NewReplacer("\\", "\\\\", "|", "\\|", "\n", " ", "\r", " ", "=", "\\=")
	return r.Replace(s)
}

func buildCEFExtension(kv map[string]string) string {
	var parts []string
	for k, v := range kv {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, " ")
}

func ifThen(b bool, yes, no string) string {
	if b {
		return yes
	}
	return no
}

func optionalUUID(p *uuid.UUID) string {
	if p == nil {
		return ""
	}
	return p.String()
}

func auditRowToMap(r AuditLog) map[string]interface{} {
	m := map[string]interface{}{
		"id":            r.ID,
		"created_at":    r.CreatedAt,
		"action":        r.Action,
		"actor_id":      r.ActorID,
		"actor_type":    r.ActorType,
		"ip_address":    r.IPAddress,
		"user_agent":    r.UserAgent,
		"request_id":    r.RequestID,
		"success":       r.Success,
		"error_message": r.ErrorMessage,
		"metadata":      r.Metadata,
	}
	if r.SecretID != nil {
		m["secret_id"] = *r.SecretID
	}
	return m
}

// ============================================================================
// SIEM dispatcher
// ============================================================================

// SIEMDispatcher pushes audit log events to registered webhooks in
// real time. One instance is constructed at process start; the
// handler calls Dispatch after writing each audit row.
type SIEMDispatcher struct {
	repo    *Repository
	client  *http.Client
	enabled bool
}

// NewSIEMDispatcher returns a dispatcher backed by the repository.
func NewSIEMDispatcher(repo *Repository) *SIEMDispatcher {
	return &SIEMDispatcher{
		repo:    repo,
		client:  &http.Client{Timeout: 10 * time.Second},
		enabled: true,
	}
}

// Dispatch reads all webhooks for the entry's tenant, formats the
// audit row, and POSTs it to each one. Best-effort: failures are
// recorded on the webhook row but never block the caller.
func (d *SIEMDispatcher) Dispatch(ctx context.Context, entry *AuditLog) {
	if d == nil || !d.enabled || entry == nil {
		return
	}
	hooks, err := d.repo.ListSIEMWebhooks(ctx, entry.TenantID)
	if err != nil || len(hooks) == 0 {
		return
	}
	for i := range hooks {
		hook := hooks[i]
		if !hook.Enabled {
			continue
		}
		body, err := renderSIEMPayload(hook.Format, entry)
		if err != nil {
			d.repo.MarkSIEMDelivery(ctx, hook.ID, 0, err.Error())
			continue
		}
		mac := hmac.New(sha256.New, hook.SecretHMAC)
		mac.Write(body)
		sig := hex.EncodeToString(mac.Sum(nil))
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, hook.URL, bytes.NewReader(body))
		if err != nil {
			d.repo.MarkSIEMDelivery(ctx, hook.ID, 0, err.Error())
			continue
		}
		req.Header.Set("Content-Type", contentTypeFor(hook.Format))
		req.Header.Set("X-Signature", sig)
		req.Header.Set("User-Agent", "functionfly-vault-siem/1.0")
		resp, err := d.client.Do(req)
		if err != nil {
			d.repo.MarkSIEMDelivery(ctx, hook.ID, 0, err.Error())
			continue
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		d.repo.MarkSIEMDelivery(ctx, hook.ID, resp.StatusCode, "")
	}
}

func renderSIEMPayload(format string, entry *AuditLog) ([]byte, error) {
	switch format {
	case "cef":
		ts := entry.CreatedAt.UTC().UnixMilli()
		ext := buildCEFExtension(map[string]string{
			"act":      entry.ActorID,
			"cs1":      string(entry.ActorType),
			"cs1Label": "actorType",
			"src":      entry.IPAddress,
			"request":  entry.RequestID,
			"cat":      "vault",
			"outcome":  ifThen(entry.Success, "success", "failure"),
			"msg":      entry.ErrorMessage,
			"rt":       strconv.FormatInt(ts, 10),
		})
		ext = escapeCEF(ext)
		name := fmt.Sprintf("vault %s by %s", entry.Action, entry.ActorType)
		out := fmt.Sprintf("CEF:0|FunctionFly|Vault|1|vault.%s|%s|%d|%s\n",
			entry.Action, name, 5, ext)
		return []byte(out), nil
	default:
		return json.Marshal(entry)
	}
}

func contentTypeFor(format string) string {
	if format == "cef" {
		return "text/plain; charset=utf-8"
	}
	return "application/json"
}
