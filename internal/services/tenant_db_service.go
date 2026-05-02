package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/functionfly/functionfly/internal/plans"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// TenantDBService handles automatic provisioning and management of dedicated tenant databases
type TenantDBService struct {
	provisioner *storage.TenantDBProvisioner
	poolManager  *storage.TenantPoolManager
	registry     *storage.TenantDBRegistry
	healthCheck *storage.TenantDBHealthChecker
	config      *storage.TenantDatabaseConfig
	// Track provisioning jobs
	provisionJobs sync.Map // map[uuid.UUID]*ProvisionJob
	// Idempotency locks for concurrent provisioning requests
	provisionLocks sync.Map // map[uuid.UUID]*sync.Mutex
	// Global lock for provisioning operations
	globalLock sync.Mutex
}

// ProvisionJob tracks the state of a database provisioning operation
type ProvisionJob struct {
	TenantID    uuid.UUID
	Status      string // "pending", "provisioning", "completed", "failed"
	StartedAt   time.Time
	CompletedAt *time.Time
	Error       string
}

// NewTenantDBService creates a new tenant database service
func NewTenantDBService(
	provisioner *storage.TenantDBProvisioner,
	poolManager *storage.TenantPoolManager,
	registry *storage.TenantDBRegistry,
	config *storage.TenantDatabaseConfig,
) *TenantDBService {
	return &TenantDBService{
		provisioner: provisioner,
		poolManager:  poolManager,
		registry:    registry,
		config:      config,
	}
}

// Start begins the health monitoring service
func (s *TenantDBService) Start(ctx context.Context) error {
	if s.healthCheck != nil {
		return s.healthCheck.Start(ctx)
	}
	return nil
}

// Stop stops all services
func (s *TenantDBService) Stop() {
	if s.healthCheck != nil {
		s.healthCheck.Stop()
	}
}

// ShouldProvisionDedicatedDB determines if a tenant should have a dedicated database
func (s *TenantDBService) ShouldProvisionDedicatedDB(plan string) bool {
	if !s.config.Enabled {
		return false
	}

	// Plans that qualify for dedicated databases
	dedicatedPlans := map[string]bool{
		plans.PlanStarter:    true,
		plans.PlanPro:        true,
		plans.PlanEnterprise: true,
	}

	return dedicatedPlans[plan]
}

// ProvisionForTenant provisions a dedicated database for a tenant
// Uses per-tenant locking to prevent race conditions during concurrent provisioning
func (s *TenantDBService) ProvisionForTenant(ctx context.Context, tenantID uuid.UUID) error {
	if !s.config.Enabled {
		return fmt.Errorf("tenant databases are disabled")
	}

	// Get or create per-tenant lock
	lockInterface, _ := s.provisionLocks.LoadOrStore(tenantID, &sync.Mutex{})
	lock := lockInterface.(*sync.Mutex)

	// Acquire per-tenant lock
	lock.Lock()
	defer lock.Unlock()

	// Double-check if already provisioned (within lock)
	existing, err := s.provisioner.GetTenantDBStatus(ctx, tenantID)
	if err == nil && existing != "" && existing != "failed" {
		logrus.Infof("Tenant %s already has a database (%s), skipping provisioning", tenantID, existing)
		return nil
	}

	// Create provision job
	job := &ProvisionJob{
		TenantID:  tenantID,
		Status:    "pending",
		StartedAt: time.Now(),
	}
	s.provisionJobs.Store(tenantID, job)

	// Update job status
	job.Status = "provisioning"

	// Provision the database
	err = s.provisioner.CreateTenantDB(ctx, tenantID)
	if err != nil {
		job.Status = "failed"
		job.Error = err.Error()
		logrus.Errorf("Failed to provision database for tenant %s: %v", tenantID, err)
		return fmt.Errorf("provisioning failed: %w", err)
	}

	// Mark as completed
	job.Status = "completed"
	now := time.Now()
	job.CompletedAt = &now

	logrus.Infof("Successfully provisioned dedicated database for tenant %s", tenantID)
	return nil
}

// DeprovisionForTenant removes a tenant's dedicated database
// Uses per-tenant locking to prevent race conditions
func (s *TenantDBService) DeprovisionForTenant(ctx context.Context, tenantID uuid.UUID) error {
	if !s.config.Enabled {
		return fmt.Errorf("tenant databases are disabled")
	}

	// Get or create per-tenant lock
	lockInterface, _ := s.provisionLocks.LoadOrStore(tenantID, &sync.Mutex{})
	lock := lockInterface.(*sync.Mutex)

	// Acquire per-tenant lock
	lock.Lock()
	defer func() {
		// Clean up lock after deprovisioning
		s.provisionLocks.Delete(tenantID)
		lock.Unlock()
	}()

	// Close the connection pool
	_ = s.poolManager.ClosePool(tenantID)

	// Delete the database
	err := s.provisioner.DeleteTenantDB(ctx, tenantID)
	if err != nil {
		logrus.Errorf("Failed to deprovision database for tenant %s: %v", tenantID, err)
		return fmt.Errorf("deprovisioning failed: %w", err)
	}

	logrus.Infof("Successfully deprovisioned database for tenant %s", tenantID)
	return nil
}

// SuspendTenant suspends a tenant's dedicated database (for payment failures, etc.)
// Uses per-tenant locking to prevent race conditions
func (s *TenantDBService) SuspendTenant(ctx context.Context, tenantID uuid.UUID) error {
	if !s.config.Enabled {
		return fmt.Errorf("tenant databases are disabled")
	}

	// Get or create per-tenant lock
	lockInterface, _ := s.provisionLocks.LoadOrStore(tenantID, &sync.Mutex{})
	lock := lockInterface.(*sync.Mutex)

	// Acquire per-tenant lock
	lock.Lock()
	defer lock.Unlock()

	// Close the connection pool first
	_ = s.poolManager.ClosePool(tenantID)

	// Suspend the database
	err := s.provisioner.SuspendTenantDB(ctx, tenantID)
	if err != nil {
		logrus.Errorf("Failed to suspend database for tenant %s: %v", tenantID, err)
		return fmt.Errorf("suspension failed: %w", err)
	}

	logrus.Infof("Suspended dedicated database for tenant %s", tenantID)
	return nil
}

// ResumeTenant resumes a suspended tenant's dedicated database
func (s *TenantDBService) ResumeTenant(ctx context.Context, tenantID uuid.UUID) error {
	if !s.config.Enabled {
		return fmt.Errorf("tenant databases are disabled")
	}

	err := s.provisioner.ResumeTenantDB(ctx, tenantID)
	if err != nil {
		logrus.Errorf("Failed to resume database for tenant %s: %v", tenantID, err)
		return fmt.Errorf("resume failed: %w", err)
	}

	logrus.Infof("Resumed dedicated database for tenant %s", tenantID)
	return nil
}

// GetPool returns a connection pool for a tenant's dedicated database
func (s *TenantDBService) GetPool(ctx context.Context, tenantID uuid.UUID) (interface{}, error) {
	return s.poolManager.GetPool(ctx, tenantID)
}

// GetHealthStatus returns the health status of a tenant's database
func (s *TenantDBService) GetHealthStatus(tenantID uuid.UUID) (*storage.TenantHealthStatus, error) {
	if s.healthCheck != nil {
		return s.healthCheck.GetTenantHealth(tenantID)
	}
	return nil, fmt.Errorf("health checker not available")
}

// GetAllHealthStatuses returns health status for all tenants
func (s *TenantDBService) GetAllHealthStatuses() []*storage.TenantHealthStatus {
	if s.healthCheck != nil {
		return s.healthCheck.GetAllTenantHealth()
	}
	return nil
}

// GetUnhealthyTenants returns all tenants with unhealthy databases
func (s *TenantDBService) GetUnhealthyTenants() []*storage.TenantHealthStatus {
	if s.healthCheck != nil {
		return s.healthCheck.GetUnhealthyTenants()
	}
	return nil
}

// GenerateHealthReport creates a comprehensive health report
func (s *TenantDBService) GenerateHealthReport() *storage.TenantHealthReport {
	if s.healthCheck != nil {
		return s.healthCheck.GenerateReport()
	}
	return nil
}

// HandlePlanChange is called when a tenant's plan changes
// This handles automatic provisioning/deprovisioning based on plan eligibility
func (s *TenantDBService) HandlePlanChange(ctx context.Context, tenantID uuid.UUID, oldPlan, newPlan string) error {
	oldQualifies := s.ShouldProvisionDedicatedDB(oldPlan)
	newQualifies := s.ShouldProvisionDedicatedDB(newPlan)

	// No change in qualification
	if oldQualifies == newQualifies {
		return nil
	}

	// Upgrading to a plan that qualifies
	if newQualifies && !oldQualifies {
		logrus.Infof("Tenant %s upgrading to %s - provisioning dedicated database", tenantID, newPlan)
		return s.ProvisionForTenant(ctx, tenantID)
	}

	// Downgrading to a plan that doesn't qualify
	if !newQualifies && oldQualifies {
		logrus.Infof("Tenant %s downgrading to %s - deprovisioning dedicated database", tenantID, newPlan)
		// Don't immediately deprovision - keep it for a grace period in case they upgrade again
		// Just suspend it for now
		return s.SuspendTenant(ctx, tenantID)
	}

	return nil
}

// HandleTenantSuspension handles tenant suspension (billing failure, etc.)
func (s *TenantDBService) HandleTenantSuspension(ctx context.Context, tenantID uuid.UUID) error {
	return s.SuspendTenant(ctx, tenantID)
}

// HandleTenantActivation handles tenant activation/resumption
func (s *TenantDBService) HandleTenantActivation(ctx context.Context, tenantID uuid.UUID) error {
	// Verify tenant exists and can be resumed
	_, err := s.registry.GetByTenantID(ctx, tenantID)
	if err != nil {
		return err
	}

	// Resume the tenant's dedicated database
	return s.ResumeTenant(ctx, tenantID)
}

// GetProvisionJob returns the status of a provisioning job
func (s *TenantDBService) GetProvisionJob(tenantID uuid.UUID) *ProvisionJob {
	if job, ok := s.provisionJobs.Load(tenantID); ok {
		return job.(*ProvisionJob)
	}
	return nil
}

// ListTenantDatabases returns information about all tenant databases
func (s *TenantDBService) ListTenantDatabases(ctx context.Context) ([]storage.TenantDBInfo, error) {
	return s.provisioner.ListTenantDatabases(ctx)
}

// GetPoolStats returns connection pool statistics
func (s *TenantDBService) GetPoolStats() storage.TenantPoolStats {
	return s.poolManager.GetStats()
}

// TenantDBServiceInterface defines the interface for tenant database management
type TenantDBServiceInterface interface {
	ProvisionForTenant(ctx context.Context, tenantID uuid.UUID) error
	DeprovisionForTenant(ctx context.Context, tenantID uuid.UUID) error
	SuspendTenant(ctx context.Context, tenantID uuid.UUID) error
	ResumeTenant(ctx context.Context, tenantID uuid.UUID) error
	HandlePlanChange(ctx context.Context, tenantID uuid.UUID, oldPlan, newPlan string) error
	GetPool(ctx context.Context, tenantID uuid.UUID) (interface{}, error)
	GetHealthStatus(tenantID uuid.UUID) (*storage.TenantHealthStatus, error)
	GetAllHealthStatuses() []*storage.TenantHealthStatus
	GetUnhealthyTenants() []*storage.TenantHealthStatus
	GenerateHealthReport() *storage.TenantHealthReport
	Start(ctx context.Context) error
	Stop()
}

// Verify implementation
var _ TenantDBServiceInterface = (*TenantDBService)(nil)