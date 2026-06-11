package billing

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ExternalExporter is the interface for exporting billing data to external accounting systems.
type ExternalExporter interface {
	// Name returns the system name (e.g. "quickbooks", "xero").
	Name() string
	// TestConnection verifies credentials and API access.
	TestConnection(ctx context.Context, system *storage.ExternalBillingSystem) error
	// ExportLineItems sends cost allocation line items to the external system.
	ExportLineItems(ctx context.Context, system *storage.ExternalBillingSystem, items []LineItemExport) (*ExportResult, error)
	// SyncCustomers syncs tenant users as customers in the external system.
	SyncCustomers(ctx context.Context, system *storage.ExternalBillingSystem, customers []CustomerExport) (*ExportResult, error)
	// SyncInvoices syncs invoices to the external system.
	SyncInvoices(ctx context.Context, system *storage.ExternalBillingSystem, invoices []InvoiceExport) (*ExportResult, error)
	// SyncPayments syncs payments to the external system.
	SyncPayments(ctx context.Context, system *storage.ExternalBillingSystem, payments []PaymentExport) (*ExportResult, error)
}

// InvoiceExport represents an invoice for external export.
type InvoiceExport struct {
	ExternalID      string     `json:"external_id,omitempty"`
	TenantID        uuid.UUID  `json:"tenant_id"`
	TenantName      string     `json:"tenant_name"`
	InvoiceNumber   string     `json:"invoice_number"`
	Status          string     `json:"status"`
	AmountDueCents  int64      `json:"amount_due_cents"`
	AmountPaidCents int64      `json:"amount_paid_cents"`
	Currency        string     `json:"currency"`
	Description     string     `json:"description"`
	PeriodStart     time.Time  `json:"period_start"`
	PeriodEnd       time.Time  `json:"period_end"`
	DueDate         time.Time  `json:"due_date"`
	PaidDate        *time.Time `json:"paid_date,omitempty"`
}

// PaymentExport represents a payment for external export.
type PaymentExport struct {
	ExternalID      string    `json:"external_id,omitempty"`
	TenantID        uuid.UUID `json:"tenant_id"`
	TenantName      string    `json:"tenant_name"`
	InvoiceNumber   string    `json:"invoice_number"`
	AmountCents     int64     `json:"amount_cents"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	PaymentDate     time.Time `json:"payment_date"`
	Description     string    `json:"description"`
	ReferenceNumber string    `json:"reference_number,omitempty"`
}

// LineItemExport represents a single billable line item for external export.
type LineItemExport struct {
	ExternalID    string            `json:"external_id,omitempty"`
	TenantID      uuid.UUID         `json:"tenant_id"`
	TenantName    string            `json:"tenant_name"`
	Description   string            `json:"description"`
	Quantity      float64           `json:"quantity"`
	UnitCostCents int64             `json:"unit_cost_cents"`
	TotalCents    int64             `json:"total_cents"`
	ServiceDate   time.Time         `json:"service_date"`
	CustomerRef   string            `json:"customer_ref,omitempty"`
	AccountRef    string            `json:"account_ref,omitempty"`
	DepartmentRef string            `json:"department_ref,omitempty"`
	Tags          map[string]string `json:"tags,omitempty"`
}

// CustomerExport represents a tenant/user exported as a customer in an external system.
type CustomerExport struct {
	ExternalID   string `json:"external_id,omitempty"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	ExternalRef  string `json:"external_ref,omitempty"`
	AddressLine1 string `json:"address_line1,omitempty"`
	AddressLine2 string `json:"address_line2,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	PostalCode   string `json:"postal_code,omitempty"`
	Country      string `json:"country,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Currency     string `json:"currency"`
}

// ExportResult holds the outcome of an export operation.
type ExportResult struct {
	RecordsProcessed int64             `json:"records_processed"`
	RecordsCreated   int64             `json:"records_created"`
	RecordsUpdated   int64             `json:"records_updated"`
	RecordsFailed    int64             `json:"records_failed"`
	ExternalBatchID  string            `json:"external_batch_id,omitempty"`
	ExternalRefs     map[string]string `json:"external_refs,omitempty"` // our ID → their ID
	Errors           []string          `json:"errors,omitempty"`
}

// QuickBooksExporter implements ExternalExporter for QuickBooks Online.
type QuickBooksExporter struct {
	client *http.Client
}

// NewQuickBooksExporter creates a new QuickBooks exporter.
func NewQuickBooksExporter() *QuickBooksExporter {
	return &QuickBooksExporter{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (e *QuickBooksExporter) Name() string { return "quickbooks" }

// TestConnection verifies the QuickBooks connection by calling the CompanyInfo endpoint.
func (e *QuickBooksExporter) TestConnection(ctx context.Context, system *storage.ExternalBillingSystem) error {
	if system.APIEndpoint == "" {
		return fmt.Errorf("quickbooks: api_endpoint is required")
	}
	if system.OAuthToken == "" {
		return fmt.Errorf("quickbooks: oauth_token is required")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", system.APIEndpoint+"/companyinfo/1", nil)
	if err != nil {
		return fmt.Errorf("quickbooks: failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+system.OAuthToken)
	req.Header.Set("Accept", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("quickbooks: connection test failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("quickbooks: API returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ExportLineItems creates bills in QuickBooks for the given line items.
func (e *QuickBooksExporter) ExportLineItems(
	ctx context.Context,
	system *storage.ExternalBillingSystem,
	items []LineItemExport,
) (*ExportResult, error) {
	if len(items) == 0 {
		return &ExportResult{}, nil
	}

	result := &ExportResult{RecordsProcessed: int64(len(items))}

	// Build Bill payload — QuickBooks uses Purchases/Bills for expenses
	for _, item := range items {
		payload := map[string]interface{}{
			"VendorRef": map[string]interface{}{
				"value": item.CustomerRef,
				"name":  item.TenantName,
			},
			"TxnDate": item.ServiceDate.Format("2006-01-02"),
			"DueDate": item.ServiceDate.AddDate(0, 0, 30).Format("2006-01-02"),
			"Line": []map[string]interface{}{
				{
					"Amount":      float64(item.TotalCents) / 100,
					"DetailType":  "AccountBasedExpenseLineDetail",
					"Description": item.Description,
					"AccountBasedExpenseLineDetail": map[string]interface{}{
						"AccountRef": map[string]interface{}{
							"value": getMappedValue(system.FieldMappings, "expense_account", "1"), // default COGS
						},
					},
				},
			},
			"Memo": fmt.Sprintf("FunctionFly usage %s", item.ServiceDate.Format("Jan 2006")),
		}

		body, err := json.Marshal(payload)
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("item %s: marshal error: %v", item.ExternalID, err))
			continue
		}

		reqURL := system.APIEndpoint + "/purchase"
		if system.FieldMappings != nil {
			if v, ok := system.FieldMappings["bill_endpoint"]; ok && v != "" {
				reqURL = system.APIEndpoint + v
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewReader(body))
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("item %s: request error: %v", item.ExternalID, err))
			continue
		}
		req.Header.Set("Authorization", "Bearer "+system.OAuthToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := e.client.Do(req)
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("item %s: request failed: %v", item.ExternalID, err))
			continue
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		if resp.StatusCode >= 400 {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("item %s: QB returned %d: %s", item.ExternalID, resp.StatusCode, string(respBody)))
			continue
		}

		var qbResp map[string]interface{}
		if err := json.Unmarshal(respBody, &qbResp); err == nil {
			if id, ok := qbResp["Id"].(string); ok {
				if result.ExternalRefs == nil {
					result.ExternalRefs = make(map[string]string)
				}
				result.ExternalRefs[item.ExternalID] = id
			}
		}
		result.RecordsCreated++
	}

	return result, nil
}

// SyncCustomers creates or updates customer records in QuickBooks.
func (e *QuickBooksExporter) SyncCustomers(
	ctx context.Context,
	system *storage.ExternalBillingSystem,
	customers []CustomerExport,
) (*ExportResult, error) {
	if len(customers) == 0 {
		return &ExportResult{}, nil
	}

	result := &ExportResult{RecordsProcessed: int64(len(customers))}

	for _, c := range customers {
		payload := map[string]interface{}{
			"DisplayName": c.Name,
			"PrimaryEmailAddr": map[string]interface{}{
				"Address": c.Email,
			},
			"PrimaryPhone": map[string]interface{}{
				"FreeFormNumber": c.Phone,
			},
		}

		if c.AddressLine1 != "" {
			payload["BillAddr"] = map[string]interface{}{
				"Line1":      c.AddressLine1,
				"Line2":      c.AddressLine2,
				"City":       c.City,
				"Country":    c.Country,
				"PostalCode": c.PostalCode,
			}
		}

		body, err := json.Marshal(payload)
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("customer %s: %v", c.Email, err))
			continue
		}

		method := "POST"
		endpoint := system.APIEndpoint + "/customer"
		if c.ExternalRef != "" {
			method = "POST" // QuickBooks uses same endpoint for create; update uses PUT /customer/{id}
			endpoint = system.APIEndpoint + "/customer?operation=update"
		}

		req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
		if err != nil {
			result.RecordsFailed++
			continue
		}
		req.Header.Set("Authorization", "Bearer "+system.OAuthToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := e.client.Do(req)
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("customer %s: %v", c.Email, err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			result.RecordsFailed++
			continue
		}
		result.RecordsCreated++
	}

	return result, nil
}

// SyncInvoices creates or updates invoices in QuickBooks.
func (e *QuickBooksExporter) SyncInvoices(
	ctx context.Context,
	system *storage.ExternalBillingSystem,
	invoices []InvoiceExport,
) (*ExportResult, error) {
	if len(invoices) == 0 {
		return &ExportResult{}, nil
	}

	result := &ExportResult{RecordsProcessed: int64(len(invoices))}

	for _, inv := range invoices {
		payload := map[string]interface{}{
			"DocNumber": inv.InvoiceNumber,
			"TxnDate":   inv.PeriodEnd.Format("2006-01-02"),
			"DueDate":   inv.DueDate.Format("2006-01-02"),
			"Currency":  inv.Currency,
			"Line": []map[string]interface{}{
				{
					"Amount":      float64(inv.AmountDueCents) / 100,
					"Description": inv.Description,
					"DetailType":  "SalesItemLineDetail",
					"SalesItemLineDetail": map[string]interface{}{
						"ItemRef": map[string]interface{}{
							"value": getMappedValue(system.FieldMappings, "revenue_account", "200"),
						},
					},
				},
			},
		}

		body, err := json.Marshal(payload)
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("invoice %s: %v", inv.InvoiceNumber, err))
			continue
		}

		req, err := http.NewRequestWithContext(ctx, "POST", system.APIEndpoint+"/invoice", bytes.NewReader(body))
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("invoice %s: %v", inv.InvoiceNumber, err))
			continue
		}
		req.Header.Set("Authorization", "Bearer "+system.OAuthToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := e.client.Do(req)
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("invoice %s: %v", inv.InvoiceNumber, err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			result.RecordsFailed++
			continue
		}
		result.RecordsCreated++
	}

	return result, nil
}

// SyncPayments creates or updates payments in QuickBooks.
func (e *QuickBooksExporter) SyncPayments(
	ctx context.Context,
	system *storage.ExternalBillingSystem,
	payments []PaymentExport,
) (*ExportResult, error) {
	if len(payments) == 0 {
		return &ExportResult{}, nil
	}

	result := &ExportResult{RecordsProcessed: int64(len(payments))}

	for _, p := range payments {
		payload := map[string]interface{}{
			"TxnDate":   p.PaymentDate.Format("2006-01-02"),
			"TotalAmt":  float64(p.AmountCents) / 100,
			"DocNumber": p.ReferenceNumber,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("payment %s: %v", p.ReferenceNumber, err))
			continue
		}

		req, err := http.NewRequestWithContext(ctx, "POST", system.APIEndpoint+"/payment", bytes.NewReader(body))
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("payment %s: %v", p.ReferenceNumber, err))
			continue
		}
		req.Header.Set("Authorization", "Bearer "+system.OAuthToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := e.client.Do(req)
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("payment %s: %v", p.ReferenceNumber, err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			result.RecordsFailed++
			continue
		}
		result.RecordsCreated++
	}

	return result, nil
}

// XeroExporter implements ExternalExporter for Xero.
type XeroExporter struct {
	client *http.Client
}

// NewXeroExporter creates a new Xero exporter.
func NewXeroExporter() *XeroExporter {
	return &XeroExporter{client: &http.Client{Timeout: 30 * time.Second}}
}

func (e *XeroExporter) Name() string { return "xero" }

// TestConnection calls Xero's Connections endpoint to verify the token.
func (e *XeroExporter) TestConnection(ctx context.Context, system *storage.ExternalBillingSystem) error {
	if system.APIEndpoint == "" {
		return fmt.Errorf("xero: api_endpoint is required (e.g. https://api.xero.com)")
	}
	if system.OAuthToken == "" {
		return fmt.Errorf("xero: oauth_token is required")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", system.APIEndpoint+"/connections", nil)
	if err != nil {
		return fmt.Errorf("xero: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+system.OAuthToken)
	req.Header.Set("Accept", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("xero: connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("xero: API returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// ExportLineItems creates repeating invoices in Xero.
func (e *XeroExporter) ExportLineItems(
	ctx context.Context,
	system *storage.ExternalBillingSystem,
	items []LineItemExport,
) (*ExportResult, error) {
	if len(items) == 0 {
		return &ExportResult{}, nil
	}

	result := &ExportResult{RecordsProcessed: int64(len(items))}

	tenantID := system.APIEndpoint[strings.Index(system.APIEndpoint, "xero.com/api.xro")+14:]
	if tenantID == "" {
		tenantID = system.APIEndpoint // fallback
	}

	for _, item := range items {
		invoice := map[string]interface{}{
			"Type":            "ACCREC", // Accounts Receivable
			"Contact":         map[string]string{"Name": item.TenantName},
			"Date":            item.ServiceDate.Format("2006-01-02"),
			"DueDate":         item.ServiceDate.AddDate(0, 0, 30).Format("2006-01-02"),
			"Status":          "AUTHORISED",
			"LineAmountTypes": "Exclusive",
			"LineItems": []map[string]interface{}{
				{
					"Description": item.Description,
					"Quantity":    item.Quantity,
					"UnitAmount":  float64(item.UnitCostCents) / 100,
					"AccountCode": getMappedValue(system.FieldMappings, "revenue_account", "200"),
				},
			},
		}

		body, err := json.Marshal([]map[string]interface{}{invoice})
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("item %s: %v", item.ExternalID, err))
			continue
		}

		reqURL := system.APIEndpoint + "/invoices"
		req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewReader(body))
		if err != nil {
			result.RecordsFailed++
			continue
		}
		req.Header.Set("Authorization", "Bearer "+system.OAuthToken)
		req.Header.Set("xero-tenant-id", tenantID)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := e.client.Do(req)
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("item %s: %v", item.ExternalID, err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			result.RecordsFailed++
			continue
		}
		result.RecordsCreated++
	}

	return result, nil
}

// SyncCustomers syncs contacts in Xero.
func (e *XeroExporter) SyncCustomers(
	ctx context.Context,
	system *storage.ExternalBillingSystem,
	customers []CustomerExport,
) (*ExportResult, error) {
	if len(customers) == 0 {
		return &ExportResult{}, nil
	}

	result := &ExportResult{RecordsProcessed: int64(len(customers))}
	tenantID := ""
	if idx := strings.Index(system.APIEndpoint, "xero.com"); idx != -1 {
		tenantID = system.APIEndpoint[idx+14:]
	}

	for _, c := range customers {
		contact := map[string]interface{}{
			"Name":         c.Name,
			"EmailAddress": c.Email,
			"Phones": []map[string]string{
				{"PhoneNumber": c.Phone},
			},
		}

		body, err := json.Marshal([]map[string]interface{}{contact})
		if err != nil {
			result.RecordsFailed++
			continue
		}

		req, err := http.NewRequestWithContext(ctx, "POST", system.APIEndpoint+"/contacts", bytes.NewReader(body))
		if err != nil {
			result.RecordsFailed++
			continue
		}
		req.Header.Set("Authorization", "Bearer "+system.OAuthToken)
		req.Header.Set("xero-tenant-id", tenantID)
		req.Header.Set("Content-Type", "application/json")

		resp, err := e.client.Do(req)
		if err != nil {
			result.RecordsFailed++
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 400 {
			result.RecordsFailed++
			continue
		}
		result.RecordsCreated++
	}

	return result, nil
}

// SyncInvoices syncs invoices to Xero.
func (e *XeroExporter) SyncInvoices(
	ctx context.Context,
	system *storage.ExternalBillingSystem,
	invoices []InvoiceExport,
) (*ExportResult, error) {
	if len(invoices) == 0 {
		return &ExportResult{}, nil
	}

	result := &ExportResult{RecordsProcessed: int64(len(invoices))}
	tenantID := extractXeroTenantID(system.APIEndpoint)

	for _, inv := range invoices {
		invoice := map[string]interface{}{
			"Type":            "ACCREC",
			"Contact":         map[string]string{"Name": inv.TenantName},
			"Date":            inv.PeriodEnd.Format("2006-01-02"),
			"DueDate":         inv.DueDate.Format("2006-01-02"),
			"Status":          mapInvoiceStatusToXero(inv.Status),
			"LineAmountTypes": "Exclusive",
			"Reference":       inv.InvoiceNumber,
			"LineItems": []map[string]interface{}{
				{
					"Description": inv.Description,
					"Quantity":    1,
					"UnitAmount":  float64(inv.AmountDueCents) / 100,
					"AccountCode": getMappedValue(system.FieldMappings, "revenue_account", "200"),
					"TaxType":     "NONE",
				},
			},
		}

		if inv.PaidDate != nil {
			invoice["Payments"] = []map[string]interface{}{
				{
					"Date":   inv.PaidDate.Format("2006-01-02"),
					"Amount": float64(inv.AmountPaidCents) / 100,
				},
			}
		}

		body, err := json.Marshal([]map[string]interface{}{invoice})
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("invoice %s: %v", inv.InvoiceNumber, err))
			continue
		}

		req, err := http.NewRequestWithContext(ctx, "POST", system.APIEndpoint+"/invoices", bytes.NewReader(body))
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("invoice %s: %v", inv.InvoiceNumber, err))
			continue
		}
		req.Header.Set("Authorization", "Bearer "+system.OAuthToken)
		req.Header.Set("xero-tenant-id", tenantID)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := e.client.Do(req)
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("invoice %s: %v", inv.InvoiceNumber, err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("invoice %s: Xero returned %d: %s", inv.InvoiceNumber, resp.StatusCode, string(errBody)))
			continue
		}
		result.RecordsCreated++
	}

	return result, nil
}

// SyncPayments syncs payments to Xero.
func (e *XeroExporter) SyncPayments(
	ctx context.Context,
	system *storage.ExternalBillingSystem,
	payments []PaymentExport,
) (*ExportResult, error) {
	if len(payments) == 0 {
		return &ExportResult{}, nil
	}

	result := &ExportResult{RecordsProcessed: int64(len(payments))}
	tenantID := extractXeroTenantID(system.APIEndpoint)

	for _, p := range payments {
		payment := map[string]interface{}{
			"Type:":     "ACCREC",
			"Invoice":   map[string]string{"InvoiceNumber": p.InvoiceNumber},
			"Date":      p.PaymentDate.Format("2006-01-02"),
			"Amount":    float64(p.AmountCents) / 100,
			"Reference": p.ReferenceNumber,
			"Status":    mapPaymentStatusToXero(p.Status),
		}

		body, err := json.Marshal([]map[string]interface{}{payment})
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("payment %s: %v", p.ReferenceNumber, err))
			continue
		}

		req, err := http.NewRequestWithContext(ctx, "POST", system.APIEndpoint+"/payments", bytes.NewReader(body))
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("payment %s: %v", p.ReferenceNumber, err))
			continue
		}
		req.Header.Set("Authorization", "Bearer "+system.OAuthToken)
		req.Header.Set("xero-tenant-id", tenantID)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := e.client.Do(req)
		if err != nil {
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("payment %s: %v", p.ReferenceNumber, err))
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			result.RecordsFailed++
			result.Errors = append(result.Errors, fmt.Sprintf("payment %s: Xero returned %d: %s", p.ReferenceNumber, resp.StatusCode, string(errBody)))
			continue
		}
		result.RecordsCreated++
	}

	return result, nil
}

// extractXeroTenantID extracts the tenant ID from the Xero API endpoint.
// The tenant ID is stored in the APIEndpoint field after the first "/" following "xero.com".
func extractXeroTenantID(apiEndpoint string) string {
	if idx := strings.Index(apiEndpoint, "xero.com"); idx != -1 {
		remaining := apiEndpoint[idx+9:]
		if slashIdx := strings.Index(remaining, "/"); slashIdx != -1 {
			return remaining[slashIdx+1:]
		}
	}
	return apiEndpoint
}

// mapInvoiceStatusToXero maps internal invoice status to Xero invoice status.
func mapInvoiceStatusToXero(status string) string {
	switch status {
	case "paid", "Paid", "PAID":
		return "PAID"
	case "due", "Due", "DUE":
		return "AUTHORISED"
	case "overdue", "Overdue", "OVERDUE":
		return "AUTHORISED"
	case "draft", "Draft", "DRAFT":
		return "DRAFT"
	case "cancelled", "Cancelled", "CANCELLED", "void", "Void", "VOID":
		return "VOIDED"
	default:
		return "AUTHORISED"
	}
}

// mapPaymentStatusToXero maps internal payment status to Xero payment status.
func mapPaymentStatusToXero(status string) string {
	switch status {
	case "completed", "Completed", "COMPLETED":
		return "AUTHORISED"
	case "pending", "Pending", "PENDING":
		return "PENDING"
	case "failed", "Failed", "FAILED":
		return "DECLINED"
	default:
		return "AUTHORISED"
	}
}

// getMappedValue returns the mapped value from field mappings or a default.
func getMappedValue(mappings map[string]string, key, defaultVal string) string {
	if mappings == nil {
		return defaultVal
	}
	if v, ok := mappings[key]; ok && v != "" {
		return v
	}
	return defaultVal
}

// ExporterRegistry maps system types to their exporter implementations.
var ExporterRegistry = map[storage.BillingSystemType]func() ExternalExporter{
	storage.BillingSystemQuickBooks: func() ExternalExporter { return NewQuickBooksExporter() },
	storage.BillingSystemXero:       func() ExternalExporter { return NewXeroExporter() },
}

// GetExporter returns an exporter for the given system type.
func GetExporter(systemType storage.BillingSystemType) (ExternalExporter, bool) {
	factory, ok := ExporterRegistry[systemType]
	if !ok {
		return nil, false
	}
	return factory(), true
}

// ==================== CSV/JSON Export Generator ====================

// GenerateCSV generates a CSV export from cost allocation entries.
func GenerateCSV(entries []*storage.CostAllocationEntry) ([]byte, error) {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)

	// Header
	header := []string{
		"Date", "TenantID", "FunctionID", "FunctionName", "FunctionAuthor",
		"ExecutionID", "Outcome", "Cached",
		"DurationMs", "CPUTimeMs", "MemoryUsedMB",
		"ExecutionCost", "ComputeCost", "PlatformFee", "DataTransfer", "TotalCost",
		"Region",
	}
	if err := w.Write(header); err != nil {
		return nil, fmt.Errorf("csv header: %w", err)
	}

	for _, e := range entries {
		row := []string{
			e.Timestamp.Format("2006-01-02 15:04:05"),
			e.TenantID.String(),
			e.FunctionID.String(),
			e.FunctionName,
			e.FunctionAuthor,
			e.ExecutionID.String(),
			e.ExecutionOutcome,
			fmt.Sprintf("%v", e.Cached),
			fmt.Sprintf("%d", e.DurationMs),
			fmt.Sprintf("%d", e.CPUTimeMs),
			fmt.Sprintf("%d", e.MemoryUsedMB),
			fmt.Sprintf("%.4f", float64(e.ExecutionCostCents)/100),
			fmt.Sprintf("%.4f", float64(e.ComputeCostCents)/100),
			fmt.Sprintf("%.4f", float64(e.PlatformFeeCents)/100),
			fmt.Sprintf("%.4f", float64(e.DataTransferCents)/100),
			fmt.Sprintf("%.4f", float64(e.TotalCostCents)/100),
			e.Region,
		}
		if err := w.Write(row); err != nil {
			return nil, fmt.Errorf("csv row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("csv flush: %w", err)
	}
	return buf.Bytes(), nil
}

// GenerateJSON generates a JSON export from cost allocation entries.
func GenerateJSON(entries []*storage.CostAllocationEntry) ([]byte, error) {
	type jsonEntry struct {
		Date           string  `json:"date"`
		TenantID       string  `json:"tenant_id"`
		FunctionID     string  `json:"function_id"`
		FunctionName   string  `json:"function_name"`
		FunctionAuthor string  `json:"function_author"`
		ExecutionID    string  `json:"execution_id"`
		Outcome        string  `json:"outcome"`
		Cached         bool    `json:"cached"`
		DurationMs     int64   `json:"duration_ms"`
		CPUTimeMs      int64   `json:"cpu_time_ms"`
		MemoryUsedMB   int64   `json:"memory_used_mb"`
		TotalCostUSD   float64 `json:"total_cost_usd"`
		Region         string  `json:"region"`
	}

	out := make([]jsonEntry, len(entries))
	for i, e := range entries {
		out[i] = jsonEntry{
			Date:           e.Timestamp.Format(time.RFC3339),
			TenantID:       e.TenantID.String(),
			FunctionID:     e.FunctionID.String(),
			FunctionName:   e.FunctionName,
			FunctionAuthor: e.FunctionAuthor,
			ExecutionID:    e.ExecutionID.String(),
			Outcome:        e.ExecutionOutcome,
			Cached:         e.Cached,
			DurationMs:     e.DurationMs,
			CPUTimeMs:      e.CPUTimeMs,
			MemoryUsedMB:   e.MemoryUsedMB,
			TotalCostUSD:   float64(e.TotalCostCents) / 100,
			Region:         e.Region,
		}
	}
	return json.MarshalIndent(out, "", "  ")
}

// ==================== Cost Allocation Service ====================

// CostAllocationService orchestrates cost tracking, budget checks, and anomaly detection.
type CostAllocationService struct {
	billingRepo *storage.BillingRepository
	exportRepo  *storage.ExportRepository
}

// NewCostAllocationService creates a new cost allocation service.
func NewCostAllocationService(billingRepo *storage.BillingRepository, exportRepo *storage.ExportRepository) *CostAllocationService {
	return &CostAllocationService{billingRepo: billingRepo, exportRepo: exportRepo}
}

// CheckBudgets iterates active department budgets and records alerts when thresholds are crossed.
func (s *CostAllocationService) CheckBudgets(ctx context.Context) error {
	budgets, err := s.billingRepo.ListDepartmentBudgets(ctx, uuid.Nil)
	if err != nil {
		return fmt.Errorf("list budgets: %w", err)
	}

	now := time.Now()
	for _, budget := range budgets {
		if !budget.IsActive || now.Before(budget.PeriodStart) || now.After(budget.PeriodEnd) {
			continue
		}

		// Fetch current spend
		updated, err := s.billingRepo.GetDepartmentBudgetSpend(ctx, budget.ID)
		if err != nil {
			logrus.WithError(err).WithField("budget_id", budget.ID).Warn("budget spend check failed")
			continue
		}

		pct := updated.SpentPercent

		// Determine alert level
		var level string
		if pct >= float64(budget.CriticalThresholdPct) {
			level = "critical"
		} else if pct >= float64(budget.WarningThresholdPct) {
			level = "warning"
		} else {
			continue // within budget
		}

		alert := &storage.BudgetAlert{
			BudgetID:     budget.ID,
			TenantID:     budget.TenantID,
			Level:        level,
			SpentPct:     pct,
			SpentCents:   updated.SpentCents,
			BudgetCents:  budget.BudgetCents,
			AlertSentTo:  budget.AlertEmail,
			AlertMessage: fmt.Sprintf("Department budget %q is at %.1f%% ($%.2f of $%.2f)", budget.Name, pct, float64(updated.SpentCents)/100, float64(budget.BudgetCents)/100),
		}
		if err := s.billingRepo.UpsertBudgetAlert(ctx, alert); err != nil {
			logrus.WithError(err).WithField("budget_id", budget.ID).Warn("failed to record budget alert")
		} else {
			logrus.WithFields(logrus.Fields{
				"budget_id": budget.ID,
				"level":     level,
				"spent_pct": pct,
			}).Info("budget alert recorded")
		}
	}
	return nil
}

// DetectCostAnomalies compares current spend to historical averages and records spikes.
// This is a simple threshold-based detector; for production use a statistical model (e.g. Z-score, Prophet).
func (s *CostAllocationService) DetectCostAnomalies(ctx context.Context, tenantID uuid.UUID) error {
	now := time.Now()
	currentStart := now.AddDate(0, 0, -1)  // last 24h
	baselineStart := now.AddDate(0, 0, -8) // last 8 days (for weekly baseline)

	// Current period spend
	current, err := s.billingRepo.GetTenantCostSummary(ctx, tenantID, currentStart, now)
	if err != nil || current == nil {
		return err
	}

	// Baseline average daily spend
	baseline, err := s.billingRepo.GetTenantCostSummary(ctx, tenantID, baselineStart, currentStart)
	if err != nil || baseline == nil {
		return err
	}

	if baseline.TotalCostCents == 0 {
		return nil // no baseline data
	}

	// Average daily baseline vs. current day
	baselineDailyAvg := float64(baseline.TotalCostCents) / 7
	if baselineDailyAvg == 0 {
		return nil
	}

	currentDaily := float64(current.TotalCostCents)
	deltaPct := ((currentDaily - baselineDailyAvg) / baselineDailyAvg) * 100

	// Flag if today's spend is more than 3x the daily average (>200% delta)
	if deltaPct > 200 {
		anomaly := &storage.CostAnomaly{
			TenantID:          tenantID,
			AnomalyType:       "spike",
			Severity:          "high",
			ExpectedCostCents: int64(baselineDailyAvg),
			ActualCostCents:   int64(currentDaily),
			DeltaCents:        int64(currentDaily - baselineDailyAvg),
			DeltaPercent:      deltaPct,
			Description:       fmt.Sprintf("Daily spend is %.0f%% above 7-day average ($%.2f vs avg $%.2f)", deltaPct, currentDaily/100, baselineDailyAvg/100),
		}
		if err := s.billingRepo.UpsertCostAnomaly(ctx, anomaly); err != nil {
			logrus.WithError(err).Warn("failed to record cost anomaly")
		}
	}

	return nil
}

// ==================== URL-safe base64 encoding ====================

// EncodeForExport encodes a byte slice as a URL-safe base64 string.
func EncodeForExport(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}
