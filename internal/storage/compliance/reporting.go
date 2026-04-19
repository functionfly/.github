package compliance

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetComplianceReports retrieves compliance assessment reports with optional framework filter and pagination
func (acs *AuditComplianceService) GetComplianceReports(ctx context.Context, framework ComplianceFramework, limit int) ([]*ComplianceReport, error) {
	var models []*ComplianceReportModel

	query := acs.db.GORM.WithContext(ctx).Model(&ComplianceReportModel{})

	if framework != "" {
		query = query.Where("framework = ?", framework)
	}

	err := query.
		Order("assessment_date DESC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance reports: %w", err)
	}

	return convertModelsToReports(models), nil
}

// GetComplianceReportByID retrieves a specific compliance report by ID
func (acs *AuditComplianceService) GetComplianceReportByID(ctx context.Context, id uuid.UUID) (*ComplianceReport, error) {
	var model ComplianceReportModel
	err := acs.db.GORM.WithContext(ctx).
		Where("id = ?", id).
		First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get compliance report: %w", err)
	}

	report := convertModelToReport(&model)
	return report, nil
}

// GenerateComplianceReport generates and stores a compliance report
func (acs *AuditComplianceService) GenerateComplianceReport(ctx context.Context, framework ComplianceFramework) (*ComplianceReport, error) {
	report, err := acs.AssessCompliance(ctx, framework)
	if err != nil {
		return nil, err
	}

	// Convert to database model and store
	model := &ComplianceReportModel{
		ID:               uuid.MustParse(report.ID),
		Framework:        report.Framework,
		AssessmentDate:   report.AssessmentDate,
		OverallScore:     report.OverallScore,
		Status:           report.Status,
		Requirements:     report.Requirements,
		CriticalFindings: report.CriticalFindings,
		Recommendations:  report.Recommendations,
		NextAssessment:   report.NextAssessment,
		Assessor:         report.Assessor,
	}

	err = acs.db.GORM.WithContext(ctx).Create(model).Error
	if err != nil {
		return nil, fmt.Errorf("failed to store compliance report: %w", err)
	}

	return report, nil
}

// convertModelsToReports converts database models to report structs
func convertModelsToReports(models []*ComplianceReportModel) []*ComplianceReport {
	reports := make([]*ComplianceReport, len(models))
	for i, m := range models {
		reports[i] = convertModelToReport(m)
	}
	return reports
}

// convertModelToReport converts a database model to a report struct
func convertModelToReport(m *ComplianceReportModel) *ComplianceReport {
	return &ComplianceReport{
		ID:               m.ID.String(),
		Framework:        m.Framework,
		AssessmentDate:   m.AssessmentDate,
		OverallScore:     m.OverallScore,
		Status:           m.Status,
		Requirements:     m.Requirements,
		CriticalFindings: m.CriticalFindings,
		Recommendations:  m.Recommendations,
		NextAssessment:   m.NextAssessment,
		Assessor:         m.Assessor,
	}
}

// GetLatestComplianceReport retrieves the most recent report for a given framework
func (acs *AuditComplianceService) GetLatestComplianceReport(ctx context.Context, framework ComplianceFramework) (*ComplianceReport, error) {
	var model ComplianceReportModel
	err := acs.db.GORM.WithContext(ctx).
		Where("framework = ?", framework).
		Order("assessment_date DESC").
		First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest compliance report: %w", err)
	}

	return convertModelToReport(&model), nil
}

// DeleteOldComplianceReports deletes reports older than the specified retention period
func (acs *AuditComplianceService) DeleteOldComplianceReports(ctx context.Context, olderThan time.Time) (int64, error) {
	result := acs.db.GORM.WithContext(ctx).
		Where("assessment_date < ?", olderThan).
		Delete(&ComplianceReportModel{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete old compliance reports: %w", result.Error)
	}

	return result.RowsAffected, nil
}
