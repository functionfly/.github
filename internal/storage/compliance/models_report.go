package compliance

import (
	"time"

	"github.com/google/uuid"
)

// ComplianceReportModel is the database model for compliance reports
type ComplianceReportModel struct {
	ID               uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Framework        ComplianceFramework    `json:"framework" gorm:"type:varchar(50);not null;index"`
	AssessmentDate   time.Time              `json:"assessment_date" gorm:"not null;index"`
	OverallScore     float64                `json:"overall_score" gorm:"type:double precision"`
	Status           string                 `json:"status" gorm:"type:varchar(50);not null"`
	Requirements     []ComplianceRequirement `json:"requirements" gorm:"type:jsonb"`
	CriticalFindings []string               `json:"critical_findings" gorm:"type:jsonb"`
	Recommendations  []string               `json:"recommendations" gorm:"type:jsonb"`
	NextAssessment   time.Time              `json:"next_assessment" gorm:"index"`
	Assessor         string                 `json:"assessor" gorm:"type:varchar(255)"`
	CreatedAt        time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for ComplianceReportModel
func (ComplianceReportModel) TableName() string {
	return "compliance_reports"
}
