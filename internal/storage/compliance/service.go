package compliance

import (
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// AuditComplianceService handles compliance-related audit operations
type AuditComplianceService struct {
	db           *storage.PostgresDB
	archiveStore storage.ArchiveStorage
	logger       *logrus.Logger
}

// NewAuditComplianceService creates a new audit compliance service
func NewAuditComplianceService(db *storage.PostgresDB, archiveStore storage.ArchiveStorage) *AuditComplianceService {
	return &AuditComplianceService{
		db:           db,
		archiveStore: archiveStore,
		logger:       logrus.New(),
	}
}
