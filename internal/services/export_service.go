package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ExportService handles the business logic for executing export jobs
type ExportService struct {
	exportRepo     *storage.ExportRepository
	billingRepo    storage.Repository
	baseURL        string
	exportBasePath string
	logger         *logrus.Logger
}

// NewExportService creates a new export service
func NewExportService(exportRepo *storage.ExportRepository, billingRepo storage.Repository, baseURL string) *ExportService {
	return &ExportService{
		exportRepo:     exportRepo,
		billingRepo:    billingRepo,
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
			"date":       startDate.Format("2006-01-02"),
			"tenant_id":  tenantID.String(),
			"cost_type":  "compute",
			"amount":     1.23,
			"currency":   "USD",
			"description": "Compute costs",
		},
	}, nil
}

// fetchExecutionData fetches execution data for the export
func (s *ExportService) fetchExecutionData(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	// This would use the execution repository to get execution data
	return []map[string]interface{}{
		{
			"execution_id": "exec-123",
			"function_id":  "func-123",
			"function_name": "Example Function",
			"started_at":   startDate.Format(time.RFC3339),
			"ended_at":     endDate.Format(time.RFC3339),
			"status":       "success",
			"runtime_ms":   500,
			"memory_mb":    128,
			"result":       "Success",
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
			"forecast_id":         forecasts.ID.String(),
			"tenant_id":           forecasts.TenantID.String(),
			"forecast_type":       forecasts.ForecastType,
			"period_start":        forecasts.PeriodStart.Format("2006-01-02"),
			"period_end":          forecasts.PeriodEnd.Format("2006-01-02"),
			"method_used":         forecasts.MethodUsed,
			"predicted_value":     forecasts.PredictedValue,
			"lower_bound":         forecasts.LowerBound,
			"upper_bound":         forecasts.UpperBound,
			"confidence":          forecasts.Confidence,
			"growth_rate":         forecasts.GrowthRate,
			"days_of_history":     forecasts.DaysOfHistory,
			"created_at":          forecasts.CreatedAt.Format(time.RFC3339),
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

// generateExcelFile generates an Excel export file (placeholder)
func (s *ExportService) generateExcelFile(jobID uuid.UUID, data map[string][]map[string]interface{}, config *storage.UsageExportConfiguration) (string, error) {
	// TODO: Implement Excel generation using a library like excelize
	// For now, fall back to CSV
	s.logger.Warn("Excel generation not yet implemented, falling back to CSV")
	return s.generateCSVFile(jobID, data, config)
}

// generateParquetFile generates a Parquet export file (placeholder)
func (s *ExportService) generateParquetFile(jobID uuid.UUID, data map[string][]map[string]interface{}, config *storage.UsageExportConfiguration) (string, error) {
	// TODO: Implement Parquet generation using a library like parquet-go
	// For now, fall back to JSON
	s.logger.Warn("Parquet generation not yet implemented, falling back to JSON")
	return s.generateJSONFile(jobID, data, config)
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

// deliverViaEmail delivers the export via email
func (s *ExportService) deliverViaEmail(ctx context.Context, job *storage.UsageExportJob, filePath, fileName string) error {
	// TODO: Implement email delivery using email service
	// This would attach the file and send to configured recipients
	s.logger.Infof("Email delivery not yet implemented for job %s", job.ID.String())
	return nil
}

// deliverViaWebhook delivers the export via webhook
func (s *ExportService) deliverViaWebhook(ctx context.Context, job *storage.UsageExportJob, filePath, fileName string) error {
	// Note: Webhook URL is not part of UsageExportJob, it's part of the config
	// This would need to be passed differently or fetched from config
	s.logger.Infof("Webhook delivery not yet implemented for job %s", job.ID.String())
	return nil
}

// deliverViaS3 delivers the export to S3
func (s *ExportService) deliverViaS3(ctx context.Context, job *storage.UsageExportJob, filePath, fileName string) error {
	// TODO: Implement S3 upload using AWS SDK
	s.logger.Infof("S3 delivery not yet implemented for job %s", job.ID.String())
	return nil
}

// ExportScheduler handles scheduled export job execution
type ExportScheduler struct {
	exportRepo  *storage.ExportRepository
	exportSvc   *ExportService
	running     bool
	ticker      *time.Ticker
	stopChan    chan bool
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
	JobID          uuid.UUID
	DownloadURL    string
	RecordCount    int64
	FileSizeBytes  int64
	Checksum       string
	Format         storage.UsageExportFormat
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
