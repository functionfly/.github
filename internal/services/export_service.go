package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/functionfly/functionfly/internal/email"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/writer"
	"github.com/xuri/excelize/v2"
)

// ExportService handles the business logic for executing export jobs
type ExportService struct {
	exportRepo     *storage.ExportRepository
	billingRepo    storage.Repository
	emailSvc       email.Service
	httpClient     *http.Client
	baseURL        string
	exportBasePath string
	logger         *logrus.Logger
}

// NewExportService creates a new export service
func NewExportService(exportRepo *storage.ExportRepository, billingRepo storage.Repository, emailSvc email.Service, baseURL string) *ExportService {
	return &ExportService{
		exportRepo:     exportRepo,
		billingRepo:    billingRepo,
		emailSvc:       emailSvc,
		httpClient:     &http.Client{Timeout: 60 * time.Second},
		baseURL:        baseURL,
		exportBasePath: "./exports",
		logger:         logrus.New(),
	}
}

// ExecuteExportJob executes a usage export job
func (s *ExportService) ExecuteExportJob(ctx context.Context, jobID uuid.UUID) error {
	// Get the job
	job, err := s.exportRepo.GetUsageExportJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get export job: %w", err)
	}

	// Update status to running
	if err := s.exportRepo.UpdateUsageExportJobStatus(ctx, jobID, storage.ExportStatusProcessing, ""); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Get the configuration
	config, err := s.exportRepo.GetUsageExportConfiguration(ctx, job.ConfigurationID)
	if err != nil {
		s.exportRepo.UpdateUsageExportJobStatus(ctx, jobID, storage.ExportStatusFailed, fmt.Sprintf("Failed to get configuration: %v", err))
		return fmt.Errorf("failed to get export configuration: %w", err)
	}

	// Determine date range
	startDate, endDate := s.calculateDateRange(job, config)

	// Fetch data based on data types
	exportData := make(map[string][]map[string]interface{})

	for _, dataType := range config.DataTypes {
		switch dataType {
		case "usage":
			usageData, err := s.fetchUsageData(ctx, config.TenantID, startDate, endDate)
			if err != nil {
				s.logger.Warnf("Failed to fetch usage data: %v", err)
			} else {
				exportData["usage"] = usageData
			}

		case "costs":
			costData, err := s.fetchCostData(ctx, config.TenantID, startDate, endDate)
			if err != nil {
				s.logger.Warnf("Failed to fetch cost data: %v", err)
			} else {
				exportData["costs"] = costData
			}

		case "executions":
			execData, err := s.fetchExecutionData(ctx, config.TenantID, startDate, endDate)
			if err != nil {
				s.logger.Warnf("Failed to fetch execution data: %v", err)
			} else {
				exportData["executions"] = execData
			}

		case "forecasts":
			forecastData, err := s.fetchForecastData(ctx, config.TenantID, startDate, endDate)
			if err != nil {
				s.logger.Warnf("Failed to fetch forecast data: %v", err)
			} else {
				exportData["forecasts"] = forecastData
			}
		}
	}

	// Apply field mappings and transformations
	if len(config.FieldMapping) > 0 {
		exportData = s.applyFieldMappings(exportData, config.FieldMapping)
	}

	// Generate the export file
	filePath, fileName, fileSize, recordCount, err := s.generateExportFile(job, exportData, config)
	if err != nil {
		s.exportRepo.UpdateUsageExportJobStatus(ctx, jobID, storage.ExportStatusFailed, fmt.Sprintf("Failed to generate export file: %v", err))
		return fmt.Errorf("failed to generate export file: %w", err)
	}

	// Generate download URL
	storageURL := fmt.Sprintf("%s/api/v1/exports/jobs/%s/download", s.baseURL, jobID.String())

	// Update job with success
	if err := s.exportRepo.CompleteUsageExportJob(ctx, jobID, filePath, storageURL, "", int64(recordCount), fileSize); err != nil {
		s.logger.Warnf("Failed to complete export job: %v", err)
	}

	// Handle delivery
	if err := s.handleDelivery(ctx, job, filePath, fileName); err != nil {
		s.logger.Warnf("Failed to handle delivery: %v", err)
	}

	// Update last execution on configuration
	s.exportRepo.UpdateLastExecution(ctx, config.ID, jobID, time.Now())

	return nil
}

// calculateDateRange determines the date range for the export
func (s *ExportService) calculateDateRange(job *storage.UsageExportJob, config *storage.UsageExportConfiguration) (time.Time, time.Time) {
	// Use the job's period dates which are set when creating the job
	return job.PeriodStart, job.PeriodEnd
}

// fetchUsageData fetches usage data for the export
func (s *ExportService) fetchUsageData(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	// This would use the billing repository to get usage data
	// For now, return mock data
	return []map[string]interface{}{
		{
			"date":          startDate.Format("2006-01-02"),
			"tenant_id":     tenantID.String(),
			"function_id":   "func-123",
			"function_name": "Example Function",
			"executions":    100,
			"runtime_ms":    5000,
			"memory_mb":     128,
			"cost":          0.12,
		},
	}, nil
}

// fetchCostData fetches cost data for the export
func (s *ExportService) fetchCostData(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	// This would use the billing repository to get cost data
	return []map[string]interface{}{
		{
			"date":        startDate.Format("2006-01-02"),
			"tenant_id":   tenantID.String(),
			"cost_type":   "compute",
			"amount":      1.23,
			"currency":    "USD",
			"description": "Compute costs",
		},
	}, nil
}

// fetchExecutionData fetches execution data for the export
func (s *ExportService) fetchExecutionData(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	// This would use the execution repository to get execution data
	return []map[string]interface{}{
		{
			"execution_id":  "exec-123",
			"function_id":   "func-123",
			"function_name": "Example Function",
			"started_at":    startDate.Format(time.RFC3339),
			"ended_at":      endDate.Format(time.RFC3339),
			"status":        "success",
			"runtime_ms":    500,
			"memory_mb":     128,
			"result":        "Success",
		},
	}, nil
}

// fetchForecastData fetches forecast data for the export
func (s *ExportService) fetchForecastData(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	// Get forecasts from the repository
	forecasts, err := s.billingRepo.GetLatestForecast(ctx, tenantID, "spend")
	if err != nil {
		return nil, err
	}

	if forecasts == nil {
		return []map[string]interface{}{}, nil
	}

	return []map[string]interface{}{
		{
			"forecast_id":     forecasts.ID.String(),
			"tenant_id":       forecasts.TenantID.String(),
			"forecast_type":   forecasts.ForecastType,
			"period_start":    forecasts.PeriodStart.Format("2006-01-02"),
			"period_end":      forecasts.PeriodEnd.Format("2006-01-02"),
			"method_used":     forecasts.MethodUsed,
			"predicted_value": forecasts.PredictedValue,
			"lower_bound":     forecasts.LowerBound,
			"upper_bound":     forecasts.UpperBound,
			"confidence":      forecasts.Confidence,
			"growth_rate":     forecasts.GrowthRate,
			"days_of_history": forecasts.DaysOfHistory,
			"created_at":      forecasts.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

// applyFieldMappings applies field mappings to the export data
func (s *ExportService) applyFieldMappings(data map[string][]map[string]interface{}, mappings map[string]string) map[string][]map[string]interface{} {
	result := make(map[string][]map[string]interface{})

	for dataType, records := range data {
		mappedRecords := make([]map[string]interface{}, len(records))
		for i, record := range records {
			mappedRecord := make(map[string]interface{})

			// Apply mappings
			for oldKey, newKey := range mappings {
				if value, exists := record[oldKey]; exists {
					mappedRecord[newKey] = value
				}
			}

			// Copy unmapped fields
			for key, value := range record {
				if _, mapped := mappings[key]; !mapped {
					mappedRecord[key] = value
				}
			}

			mappedRecords[i] = mappedRecord
		}
		result[dataType] = mappedRecords
	}

	return result
}

// generateExportFile generates the export file in the specified format
func (s *ExportService) generateExportFile(job *storage.UsageExportJob, data map[string][]map[string]interface{}, config *storage.UsageExportConfiguration) (string, string, int64, int, error) {
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("export_%s_%s", job.ID.String(), timestamp)

	var filePath string
	var err error

	switch config.Format {
	case storage.ExportFormatCSV:
		fileName = fileName + ".csv"
		filePath, err = s.generateCSVFile(job.ID, data, config)
	case storage.ExportFormatJSON:
		fileName = fileName + ".json"
		filePath, err = s.generateJSONFile(job.ID, data, config)
	case storage.ExportFormatExcel:
		fileName = fileName + ".xlsx"
		filePath, err = s.generateExcelFile(job.ID, data, config)
	case storage.ExportFormatParquet:
		fileName = fileName + ".parquet"
		filePath, err = s.generateParquetFile(job.ID, data, config)
	default:
		fileName = fileName + ".json"
		filePath, err = s.generateJSONFile(job.ID, data, config)
	}

	if err != nil {
		return "", "", 0, 0, err
	}

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", "", 0, 0, fmt.Errorf("failed to get file info: %w", err)
	}

	// Count records
	recordCount := 0
	for _, records := range data {
		recordCount += len(records)
	}

	return filePath, fileName, fileInfo.Size(), recordCount, nil
}

// generateCSVFile generates a CSV export file
func (s *ExportService) generateCSVFile(jobID uuid.UUID, data map[string][]map[string]interface{}, config *storage.UsageExportConfiguration) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("export_%s_%s.csv", jobID.String(), timestamp)
	filePath := filepath.Join(s.exportBasePath, fileName)

	// Ensure directory exists
	if err := os.MkdirAll(s.exportBasePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create export directory: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Determine headers from first record
	var headers []string
	for _, records := range data {
		if len(records) > 0 {
			for key := range records[0] {
				headers = append(headers, key)
			}
			break
		}
	}

	// Write headers
	if len(headers) > 0 {
		writer.Write(headers)
	}

	// Write data
	for _, records := range data {
		for _, record := range records {
			row := make([]string, len(headers))
			for i, header := range headers {
				if value, exists := record[header]; exists {
					row[i] = fmt.Sprintf("%v", value)
				}
			}
			writer.Write(row)
		}
	}

	return filePath, nil
}

// generateJSONFile generates a JSON export file
func (s *ExportService) generateJSONFile(jobID uuid.UUID, data map[string][]map[string]interface{}, config *storage.UsageExportConfiguration) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("export_%s_%s.json", jobID.String(), timestamp)
	filePath := filepath.Join(s.exportBasePath, fileName)

	// Ensure directory exists
	if err := os.MkdirAll(s.exportBasePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create export directory: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return "", fmt.Errorf("failed to encode JSON: %w", err)
	}

	return filePath, nil
}

// generateExcelFile generates an Excel export file with multiple sheets using excelize
func (s *ExportService) generateExcelFile(jobID uuid.UUID, data map[string][]map[string]interface{}, config *storage.UsageExportConfiguration) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("export_%s_%s.xlsx", jobID.String(), timestamp)
	filePath := filepath.Join(s.exportBasePath, fileName)

	// Ensure directory exists
	if err := os.MkdirAll(s.exportBasePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create export directory: %w", err)
	}

	// Create new Excel file
	f := excelize.NewFile()
	defer f.Close()

	// Track if we add any sheets (to delete default sheet later)
	sheetCount := 0

	// Create a sheet for each data type
	for dataType, records := range data {
		if len(records) == 0 {
			continue
		}

		// Sanitize sheet name (max 31 chars, no special chars)
		sheetName := sanitizeSheetName(dataType)

		// Create new sheet
		var sheetIndex int
		if sheetCount == 0 {
			// First sheet: rename the default "Sheet1"
			f.SetSheetName("Sheet1", sheetName)
			sheetIndex = 0
		} else {
			// Additional sheets
			index, err := f.NewSheet(sheetName)
			if err != nil {
				s.logger.Warnf("Failed to create sheet %s: %v", sheetName, err)
				continue
			}
			sheetIndex = index
		}
		sheetCount++

		// Get headers from first record
		var headers []string
		for key := range records[0] {
			headers = append(headers, key)
		}
		sort.Strings(headers) // Consistent ordering

		// Write headers
		for colIdx, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
			f.SetCellValue(sheetName, cell, header)
		}

		// Style header row
		style, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
		})
		f.SetRowStyle(sheetName, 1, 1, style)

		// Write data rows
		for rowIdx, record := range records {
			for colIdx, header := range headers {
				cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
				value := record[header]
				// Handle nil values
				if value == nil {
					f.SetCellValue(sheetName, cell, "")
				} else {
					f.SetCellValue(sheetName, cell, value)
				}
			}
		}

		// Auto-fit columns (set a reasonable width based on content)
		for colIdx := range headers {
			col, _ := excelize.ColumnNumberToName(colIdx + 1)
			f.SetColWidth(sheetName, col, col, 18) // Default 18 char width
		}

		// Set column width for specific known wide columns
		for colIdx, header := range headers {
			if header == "description" || header == "function_name" || header == "result" {
				col, _ := excelize.ColumnNumberToName(colIdx + 1)
				f.SetColWidth(sheetName, col, col, 40)
			}
		}

		// Freeze header row
		f.SetPanes(sheetName, &excelize.Panes{
			Freeze:      true,
			Split:       false,
			TopLeftCell: "A2",
			XSplit:      0,
			YSplit:      1,
		})

		_ = sheetIndex // avoid unused variable
	}

	// If no data was added, add an empty "Data" sheet
	if sheetCount == 0 {
		f.SetSheetName("Sheet1", "Data")
		f.SetCellValue("Data", "A1", "No data available")
	}

	// Save file
	if err := f.SaveAs(filePath); err != nil {
		return "", fmt.Errorf("failed to save Excel file: %w", err)
	}

	return filePath, nil
}

// sanitizeSheetName ensures sheet name is valid for Excel (max 31 chars, no special chars)
func sanitizeSheetName(name string) string {
	// Replace invalid characters
	replacer := strings.NewReplacer(
		":", "_",
		"\\", "_",
		"/", "_",
		"?", "_",
		"*", "_",
		"[", "_",
		"]", "_",
	)
	name = replacer.Replace(name)

	// Truncate to 31 characters (Excel limit)
	if len(name) > 31 {
		name = name[:31]
	}

	// Ensure not empty
	if name == "" {
		name = "Data"
	}

	return name
}

// generateParquetFile generates a Parquet export file with multiple data types as separate row groups
func (s *ExportService) generateParquetFile(jobID uuid.UUID, data map[string][]map[string]interface{}, config *storage.UsageExportConfiguration) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("export_%s_%s.parquet", jobID.String(), timestamp)
	filePath := filepath.Join(s.exportBasePath, fileName)

	// Ensure directory exists
	if err := os.MkdirAll(s.exportBasePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create export directory: %w", err)
	}

	// Flatten all data into a single table with a "_data_type" column
	// This allows multiple datasets in a single Parquet file
	var allHeaders []string
	allHeaders = append(allHeaders, "_data_type") // First column indicates the data type

	// Collect all unique headers across all data types
	headerSet := make(map[string]bool)
	for _, records := range data {
		if len(records) > 0 {
			for key := range records[0] {
				headerSet[key] = true
			}
		}
	}

	// Convert to sorted slice for consistent ordering
	for key := range headerSet {
		allHeaders = append(allHeaders, key)
	}
	sort.Strings(allHeaders[1:]) // Sort all except "_data_type"

	// Create metadata for CSV writer (format: "name=Name, type=BYTE_ARRAY, convertedtype=UTF8")
	md := buildParquetMetadata(allHeaders)

	// Create local file writer using parquet-go's local package
	fw, err := local.NewLocalFileWriter(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create parquet file: %w", err)
	}
	defer fw.Close()

	// Create CSV writer (writes Parquet with CSV-like interface)
	pw, err := writer.NewCSVWriter(md, fw, 4)
	if err != nil {
		return "", fmt.Errorf("failed to create parquet writer: %w", err)
	}

	// Write records
	recordCount := 0
	for dataType, records := range data {
		for _, record := range records {
			// Build row as []*string for WriteString method
			row := make([]*string, len(allHeaders))
			row[0] = strPtr(dataType) // First column is the data type

			for i, header := range allHeaders[1:] {
				if value, exists := record[header]; exists && value != nil {
					strValue := fmt.Sprintf("%v", value)
					row[i+1] = &strValue
				} // nil values stay nil (NULL in parquet)
			}

			if err := pw.WriteString(row); err != nil {
				s.logger.Warnf("Failed to write parquet row: %v", err)
				continue
			}
			recordCount++
		}
	}

	// Flush writer
	if err := pw.WriteStop(); err != nil {
		return "", fmt.Errorf("failed to finalize parquet file: %w", err)
	}

	s.logger.Infof("Generated Parquet file with %d records", recordCount)
	return filePath, nil
}

// buildParquetMetadata creates metadata definitions for CSV writer
// Format: "name=Name, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"
func buildParquetMetadata(headers []string) []string {
	md := make([]string, len(headers))
	for i, header := range headers {
		// Sanitize field name (remove special chars)
		fieldName := sanitizeParquetFieldName(header)
		// Use BYTE_ARRAY with UTF8 for all string columns
		md[i] = fmt.Sprintf("name=%s, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY", fieldName)
	}
	return md
}

// sanitizeParquetFieldName ensures field name is valid for Parquet schema
func sanitizeParquetFieldName(name string) string {
	// Parquet field names must start with letter or underscore, contain only alphanumeric and underscore
	result := strings.Builder{}
	firstChar := true
	for _, ch := range name {
		if firstChar {
			// First character: must be letter or underscore
			if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' {
				result.WriteRune(ch)
				firstChar = false
			} else if ch >= '0' && ch <= '9' {
				// Starts with digit: prefix with underscore
				result.WriteRune('_')
				result.WriteRune(ch)
				firstChar = false
			} else {
				// Invalid first char: use underscore
				result.WriteRune('_')
				firstChar = false
				// Write valid char if possible
				if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
					result.WriteRune(ch)
				}
			}
		} else {
			// Subsequent characters: alphanumeric or underscore
			if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' {
				result.WriteRune(ch)
			} else {
				result.WriteRune('_')
			}
		}
	}
	// Ensure not empty
	if result.Len() == 0 {
		return "field"
	}
	return result.String()
}

// handleDelivery handles the delivery of the export file
func (s *ExportService) handleDelivery(ctx context.Context, job *storage.UsageExportJob, filePath, fileName string) error {
	switch job.DeliveryMethod {
	case "email":
		if err := s.deliverViaEmail(ctx, job, filePath, fileName); err != nil {
			s.logger.Warnf("Failed to deliver via email: %v", err)
		}
	case "webhook":
		if err := s.deliverViaWebhook(ctx, job, filePath, fileName); err != nil {
			s.logger.Warnf("Failed to deliver via webhook: %v", err)
		}
	case "s3":
		if err := s.deliverViaS3(ctx, job, filePath, fileName); err != nil {
			s.logger.Warnf("Failed to deliver via S3: %v", err)
		}
	}
	return nil
}

// deliverViaEmail delivers the export via email using Resend email service
func (s *ExportService) deliverViaEmail(ctx context.Context, job *storage.UsageExportJob, filePath, fileName string) error {
	// Get the configuration to find email recipients
	config, err := s.exportRepo.GetUsageExportConfiguration(ctx, job.ConfigurationID)
	if err != nil {
		return fmt.Errorf("failed to get export configuration for email delivery: %w", err)
	}

	// Determine recipient email addresses
	var recipients []string
	if len(config.EmailRecipients) > 0 {
		recipients = config.EmailRecipients
	} else {
		// Fallback: get the user who created the configuration
		user, err := s.billingRepo.GetUserByID(ctx, config.CreatedBy)
		if err != nil {
			return fmt.Errorf("failed to get user for email delivery: %w", err)
		}
		if user == nil {
			return fmt.Errorf("configuration creator user not found")
		}
		recipients = []string{user.Email}
	}

	if len(recipients) == 0 {
		return fmt.Errorf("no email recipients configured for export job")
	}

	// Determine expiration time (default to 7 days from now if not set)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if job.ExpiresAt != nil {
		expiresAt = *job.ExpiresAt
	}

	// Send email to each recipient
	for _, email := range recipients {
		if err := s.emailSvc.SendUsageExportReady(email, job.ID.String(), job.StorageURL, expiresAt, job.FileSizeBytes); err != nil {
			s.logger.WithError(err).Warnf("Failed to send export ready email to %s", email)
			// Continue to try other recipients even if one fails
		}
	}

	s.logger.Infof("Email delivery completed for job %s to %d recipients", job.ID.String(), len(recipients))
	return nil
}

// deliverViaWebhook delivers the export via webhook
func (s *ExportService) deliverViaWebhook(ctx context.Context, job *storage.UsageExportJob, filePath, fileName string) error {
	config, err := s.exportRepo.GetUsageExportConfiguration(ctx, job.ConfigurationID)
	if err != nil {
		return fmt.Errorf("failed to get export configuration for webhook delivery: %w", err)
	}

	webhookURL := config.WebhookURL
	if webhookURL == "" {
		return fmt.Errorf("webhook URL not configured for export job")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	payload := map[string]interface{}{
		"job_id":      job.ID.String(),
		"tenant_id":   job.TenantID.String(),
		"format":      string(job.Format),
		"file_name":   fileName,
		"file_size":   job.FileSizeBytes,
		"storage_url": job.StorageURL,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"data":        string(content),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Export-Job-ID", job.ID.String())

	if secret := os.Getenv("EXPORT_WEBHOOK_SECRET"); secret != "" {
		req.Header.Set("X-Webhook-Secret", secret)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(respBody))
	}

	s.logger.Infof("Webhook delivery successful for job %s to %s", job.ID.String(), webhookURL)
	return nil
}

// deliverViaS3 delivers the export to S3
func (s *ExportService) deliverViaS3(ctx context.Context, job *storage.UsageExportJob, filePath, fileName string) error {
	bucket := os.Getenv("EXPORT_S3_BUCKET")
	if bucket == "" {
		return fmt.Errorf("EXPORT_S3_BUCKET not configured")
	}

	s3Region := os.Getenv("EXPORT_S3_REGION")
	if s3Region == "" {
		s3Region = "us-east-1"
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(s3Region))
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Endpoint := os.Getenv("EXPORT_S3_ENDPOINT")
	opts := func(o *s3.Options) {
		if s3Endpoint != "" {
			o.UsePathStyle = true
			o.BaseEndpoint = aws.String(s3Endpoint)
		}
	}
	s3Client := s3.NewFromConfig(cfg, opts)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	key := fmt.Sprintf("exports/%s/%s/%s", job.TenantID.String(), job.ID.String(), fileName)
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	s.logger.Infof("Export uploaded to S3: bucket=%s, key=%s", bucket, key)
	return nil
}

// ExportScheduler handles scheduled export job execution
type ExportScheduler struct {
	exportRepo *storage.ExportRepository
	exportSvc  *ExportService
	running    bool
	ticker     *time.Ticker
	stopChan   chan bool
}

// NewExportScheduler creates a new export scheduler
func NewExportScheduler(exportRepo *storage.ExportRepository, exportSvc *ExportService) *ExportScheduler {
	return &ExportScheduler{
		exportRepo: exportRepo,
		exportSvc:  exportSvc,
		stopChan:   make(chan bool),
	}
}

// Start begins the export scheduler
func (s *ExportScheduler) Start() {
	if s.running {
		return
	}

	s.running = true
	s.ticker = time.NewTicker(1 * time.Minute) // Check every minute

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.checkScheduledExports()
			case <-s.stopChan:
				s.ticker.Stop()
				return
			}
		}
	}()

	s.exportSvc.logger.Info("Export scheduler started")
}

// Stop stops the export scheduler
func (s *ExportScheduler) Stop() {
	if !s.running {
		return
	}

	s.running = false
	s.stopChan <- true
	s.exportSvc.logger.Info("Export scheduler stopped")
}

// checkScheduledExports checks for and executes scheduled exports
func (s *ExportScheduler) checkScheduledExports() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get pending scheduled configurations
	configs, err := s.exportRepo.GetPendingScheduledConfigs(ctx, time.Now())
	if err != nil {
		s.exportSvc.logger.Warnf("Failed to get pending scheduled configurations: %v", err)
		return
	}

	now := time.Now()

	for _, config := range configs {
		// Create a new export job
		job := &storage.UsageExportJob{
			ID:              uuid.New(),
			ConfigurationID: config.ID,
			TenantID:        config.TenantID,
			Status:          storage.ExportStatusPending,
			Format:          config.Format,
			DataTypes:       config.DataTypes,
			PeriodStart:     now.AddDate(0, 0, -30), // Default 30 days
			PeriodEnd:       now,
			CreatedAt:       now,
			TriggeredBy:     "schedule",
		}

		if err := s.exportRepo.CreateUsageExportJob(ctx, job); err != nil {
			s.exportSvc.logger.Warnf("Failed to create scheduled export job: %v", err)
			continue
		}

		// Execute the job asynchronously
		go func(jobID uuid.UUID) {
			if err := s.exportSvc.ExecuteExportJob(ctx, jobID); err != nil {
				s.exportSvc.logger.Warnf("Scheduled export job failed: %v", err)
			}
		}(job.ID)

		// Update last execution time
		s.exportRepo.UpdateLastExecution(ctx, config.ID, job.ID, now)
	}
}

// strPtr is a helper function to create a string pointer
func strPtr(s string) *string {
	return &s
}

// ExecuteExportResult contains the result of an export execution
type ExecuteExportResult struct {
	JobID         uuid.UUID
	DownloadURL   string
	RecordCount   int64
	FileSizeBytes int64
	Checksum      string
	Format        storage.UsageExportFormat
}

// ExecuteExport executes a usage export and returns the result
func (s *ExportService) ExecuteExport(ctx context.Context, config *storage.UsageExportConfiguration, triggeredBy string) (*ExecuteExportResult, error) {
	now := time.Now()

	// Create a new export job
	job := &storage.UsageExportJob{
		ID:              uuid.New(),
		ConfigurationID: config.ID,
		TenantID:        config.TenantID,
		Status:          storage.ExportStatusPending,
		Format:          config.Format,
		DataTypes:       config.DataTypes,
		PeriodStart:     now.AddDate(0, 0, -30), // Default 30 days
		PeriodEnd:       now,
		TriggeredBy:     triggeredBy,
		CreatedAt:       now,
	}

	if err := s.exportRepo.CreateUsageExportJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create export job: %w", err)
	}

	// Execute the job
	if err := s.ExecuteExportJob(ctx, job.ID); err != nil {
		return nil, err
	}

	// Refresh the job data
	job, err := s.exportRepo.GetUsageExportJob(ctx, job.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated job: %w", err)
	}

	return &ExecuteExportResult{
		JobID:         job.ID,
		DownloadURL:   job.StorageURL,
		RecordCount:   job.RecordCount,
		FileSizeBytes: job.FileSizeBytes,
		Checksum:      job.Checksum,
		Format:        job.Format,
	}, nil
}

// GetExportFile retrieves the exported file data
func (s *ExportService) GetExportFile(ctx context.Context, jobID uuid.UUID) ([]byte, string, error) {
	// Get the job
	job, err := s.exportRepo.GetUsageExportJob(ctx, jobID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get export job: %w", err)
	}

	if job.Status != storage.ExportStatusCompleted {
		return nil, "", fmt.Errorf("export job not completed")
	}

	// Read the file
	data, err := os.ReadFile(job.StoragePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read export file: %w", err)
	}

	// Determine content type based on format
	contentType := "application/octet-stream"
	switch job.Format {
	case storage.ExportFormatCSV:
		contentType = "text/csv"
	case storage.ExportFormatJSON:
		contentType = "application/json"
	case storage.ExportFormatExcel:
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case storage.ExportFormatParquet:
		contentType = "application/octet-stream"
	}

	return data, contentType, nil
}
