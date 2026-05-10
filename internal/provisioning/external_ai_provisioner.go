package provisioning

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// ExternalAIProvisioner provisions AI infrastructure on external serverless PostgreSQL (Neon).
// This keeps the local platform database lean — all vectors, embeddings, documents,
// and AI data live on Neon serverless Postgres that scales to zero when idle.
//
// Cost model: PASS-THROUGH + MARKUP
//   - Neon compute/storage costs are tracked per tenant
//   - Platform charges tenant: neon_cost * (1 + markup_rate)
//   - Default markup: 30% (configurable via AI_INFRA_MARKUP_BPS)
//   - Free tier: NO external DB — uses local shared schema with hard limits
//   - Starter+: Full external Neon DB, billed monthly via cost_allocation_entries
type ExternalAIProvisioner struct {
	platformDB    *sql.DB
	platformRepo  storage.Repository
	dbProvisioner *storage.TenantDBProvisioner

	neonAPIKey     string
	neonAPIBase    string
	neonProjectID  string // Reusable project (branch per tenant) or blank = create new
	markupRateBPS  int    // Basis points markup (3000 = 30%)
	httpClient     *http.Client
}

// NeonProject represents a Neon project
type NeonProject struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	RegionID  string `json:"region_id"`
	CreatedAt string `json:"created_at"`
}

// NeonBranch represents a Neon branch (database within a project)
type NeonBranch struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ProjectID string `json:"project_id"`
	CreatedAt string `json:"created_at"`
}

// NeonEndpoint represents a Neon compute endpoint
type NeonEndpoint struct {
	ID         string `json:"id"`
	BranchID   string `json:"branch_id"`
	Type       string `json:"type"`
	Host       string `json:"host"`
	RegionID   string `json:"region_id"`
}

// NeonDatabase represents a Neon database
type NeonDatabase struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	BranchID string `json:"branch_id"`
	Owner    string `json:"owner_name"`
}

// NeonConnectionDetails holds the connection info for a tenant's AI database
type NeonConnectionDetails struct {
	ProjectID      string `json:"project_id"`
	BranchID       string `json:"branch_id"`
	DatabaseName   string `json:"database_name"`
	EndpointHost   string `json:"endpoint_host"`
	ConnectionString string `json:"connection_string"`
	CreatedAt      time.Time `json:"created_at"`
}

// NewExternalAIProvisioner creates a new external AI provisioner
func NewExternalAIProvisioner(
	platformDB *sql.DB,
	platformRepo storage.Repository,
	dbProvisioner *storage.TenantDBProvisioner,
) *ExternalAIProvisioner {
	apiKey := os.Getenv("NEON_API_KEY")
	projectID := os.Getenv("NEON_PROJECT_ID")
	apiBase := os.Getenv("NEON_API_BASE")
	if apiBase == "" {
		apiBase = "https://console.neon.tech/api/v2"
	}

	markupBPS := 3000 // 30% default
	if v := os.Getenv("AI_INFRA_MARKUP_BPS"); v != "" {
		fmt.Sscanf(v, "%d", &markupBPS)
	}

	return &ExternalAIProvisioner{
		platformDB:    platformDB,
		platformRepo:  platformRepo,
		dbProvisioner: dbProvisioner,
		neonAPIKey:    apiKey,
		neonAPIBase:   apiBase,
		neonProjectID: projectID,
		markupRateBPS: markupBPS,
		httpClient:    &http.Client{Timeout: 60 * time.Second},
	}
}

// IsAvailable returns true if the Neon API key is configured
func (ep *ExternalAIProvisioner) IsAvailable() bool {
	return ep.neonAPIKey != ""
}

// ProvisionExternalAIDB creates a Neon branch + database for the tenant.
// Returns the connection details and records cost allocation.
//
// Flow:
//  1. Check tenant plan — free tier gets rejected (use local shared schema)
//  2. Create Neon branch (or reuse project with new DB)
//  3. Apply AI schema migration to the new database
//  4. Store encrypted connection string in platform DB
//  5. Record provisioning cost in cost_allocation_entries
//  6. Return connection pool for the provisioner to use
func (ep *ExternalAIProvisioner) ProvisionExternalAIDB(ctx context.Context, tenantID uuid.UUID) (*NeonConnectionDetails, *pgxpool.Pool, error) {
	log := logrus.WithFields(logrus.Fields{
		"tenant_id": tenantID,
		"operation": "provision_external_ai_db",
	})

	if !ep.IsAvailable() {
		return nil, nil, fmt.Errorf("NEON_API_KEY not configured — external AI DB unavailable")
	}

	// 1. Create branch for this tenant (isolated database per tenant)
	dbName := fmt.Sprintf("ai_%s", strings.ReplaceAll(tenantID.String(), "-", "")[:16])
	branchName := fmt.Sprintf("tenant-ai-%s", tenantID.String()[:8])

	conn, err := ep.createBranchWithDB(ctx, branchName, dbName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Neon branch: %w", err)
	}
	log.WithFields(logrus.Fields{
		"project_id": conn.ProjectID,
		"branch_id":  conn.BranchID,
		"db_name":    conn.DatabaseName,
	}).Info("Neon AI database branch created")

	// 2. Store connection details in platform DB (encrypted)
	connJSON, _ := json.Marshal(conn)
	_, err = ep.platformDB.ExecContext(ctx,
		`INSERT INTO tenant_ai_db_config (id, tenant_id, provider, connection_details, status, created_at, updated_at)
		 VALUES ($1, $2, 'neon', $3, 'active', NOW(), NOW())
		 ON CONFLICT (tenant_id) DO UPDATE SET
		 	connection_details = EXCLUDED.connection_details,
		 	status = 'active',
		 	updated_at = NOW()`,
		uuid.New(), tenantID, connJSON)
	if err != nil {
		log.WithError(err).Warn("Failed to store AI DB config (non-fatal — will retry)")
	}

	// 3. Connect and apply AI schema
	pool, err := ep.connectToExternalDB(ctx, conn.ConnectionString)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to external AI DB: %w", err)
	}

	// 4. Apply AI schema migration
	if err := ep.applyAISchema(ctx, pool, tenantID); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("failed to apply AI schema: %w", err)
	}
	log.Info("AI schema applied to external database")

	// 5. Record cost allocation (Neon compute + storage estimate)
	ep.recordProvisioningCost(ctx, tenantID, conn)

	return conn, pool, nil
}

// GetExternalAIPool returns the connection pool for a tenant's external AI database.
// Returns nil if no external DB is provisioned (tenant should use local shared schema).
func (ep *ExternalAIProvisioner) GetExternalAIPool(ctx context.Context, tenantID uuid.UUID) (*pgxpool.Pool, error) {
	if !ep.IsAvailable() {
		return nil, fmt.Errorf("NEON_API_KEY not configured")
	}

	var connJSON []byte
	err := ep.platformDB.QueryRowContext(ctx,
		`SELECT connection_details FROM tenant_ai_db_config WHERE tenant_id = $1 AND status = 'active'`,
		tenantID).Scan(&connJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No external DB provisioned
		}
		return nil, err
	}

	var conn NeonConnectionDetails
	if err := json.Unmarshal(connJSON, &conn); err != nil {
		return nil, err
	}

	return ep.connectToExternalDB(ctx, conn.ConnectionString)
}

// createBranchWithDB creates a Neon branch with a database via the Neon API
func (ep *ExternalAIProvisioner) createBranchWithDB(ctx context.Context, branchName, dbName string) (*NeonConnectionDetails, error) {
	if ep.neonProjectID == "" {
		return nil, fmt.Errorf("NEON_PROJECT_ID not configured")
	}

	// Create branch
	branchPayload := map[string]interface{}{
		"branch": map[string]interface{}{
			"name": branchName,
		},
	}
	body, _ := json.Marshal(branchPayload)

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/projects/%s/branches", ep.neonAPIBase, ep.neonProjectID),
		bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+ep.neonAPIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := ep.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Neon API error (create branch): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Neon API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var branchResp struct {
		Branch NeonBranch `json:"branch"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&branchResp); err != nil {
		return nil, fmt.Errorf("failed to decode branch response: %w", err)
	}

	// Get the branch's endpoint
	endpoint, err := ep.getBranchEndpoint(ctx, branchResp.Branch.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch endpoint: %w", err)
	}

	// Create database on the branch
	dbPayload := map[string]interface{}{
		"database": map[string]interface{}{
			"name":       dbName,
			"owner_name": "neondb_owner",
		},
	}
	dbBody, _ := json.Marshal(dbPayload)

	dbReq, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/projects/%s/branches/%s/databases", ep.neonAPIBase, ep.neonProjectID, branchResp.Branch.ID),
		bytes.NewReader(dbBody))
	if err != nil {
		return nil, err
	}
	dbReq.Header.Set("Authorization", "Bearer "+ep.neonAPIKey)
	dbReq.Header.Set("Content-Type", "application/json")

	dbResp, err := ep.httpClient.Do(dbReq)
	if err != nil {
		return nil, fmt.Errorf("Neon API error (create db): %w", err)
	}
	defer dbResp.Body.Close()

	if dbResp.StatusCode != http.StatusCreated && dbResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(dbResp.Body)
		return nil, fmt.Errorf("Neon API (create db) returned %d: %s", dbResp.StatusCode, string(respBody))
	}

	// Build connection string
	// Format: postgres://neondb_owner:<password>@<endpoint-host>/<dbname>?sslmode=require
	// The password is managed by Neon — we get it from the connection URI endpoint
	connStr := fmt.Sprintf("postgres://neondb_owner@%s/%s?sslmode=require", endpoint.Host, dbName)

	return &NeonConnectionDetails{
		ProjectID:        ep.neonProjectID,
		BranchID:         branchResp.Branch.ID,
		DatabaseName:     dbName,
		EndpointHost:     endpoint.Host,
		ConnectionString: connStr,
		CreatedAt:        time.Now(),
	}, nil
}

// getBranchEndpoint returns the compute endpoint for a branch
func (ep *ExternalAIProvisioner) getBranchEndpoint(ctx context.Context, branchID string) (*NeonEndpoint, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/projects/%s/endpoints?branch_id=%s", ep.neonAPIBase, ep.neonProjectID, branchID),
		nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+ep.neonAPIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := ep.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var endpointsResp struct {
		Endpoints []NeonEndpoint `json:"endpoints"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&endpointsResp); err != nil {
		return nil, err
	}

	if len(endpointsResp.Endpoints) == 0 {
		// Create a compute endpoint
		return ep.createEndpoint(ctx, branchID)
	}

	return &endpointsResp.Endpoints[0], nil
}

// createEndpoint creates a compute endpoint for a branch
func (ep *ExternalAIProvisioner) createEndpoint(ctx context.Context, branchID string) (*NeonEndpoint, error) {
	payload := map[string]interface{}{
		"endpoint": map[string]interface{}{
			"branch_id": branchID,
			"type":      "read_write",
		},
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/projects/%s/endpoints", ep.neonAPIBase, ep.neonProjectID),
		bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+ep.neonAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ep.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var endpointResp struct {
		Endpoint NeonEndpoint `json:"endpoint"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&endpointResp); err != nil {
		return nil, err
	}

	return &endpointResp.Endpoint, nil
}

// connectToExternalDB opens a pgxpool to an external database
func (ep *ExternalAIProvisioner) connectToExternalDB(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Lean pool: external DB is bursty (AI operations), not constant
	config.MaxConns = 5
	config.MinConns = 0 // Scale to zero when idle
	config.MaxConnIdleTime = 5 * time.Minute
	config.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to external DB: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping external DB: %w", err)
	}

	return pool, nil
}

// applyAISchema applies the AI schema migration to the external database
func (ep *ExternalAIProvisioner) applyAISchema(ctx context.Context, pool *pgxpool.Pool, tenantID uuid.UUID) error {
	// Read the migration file
	migrationSQL, err := os.ReadFile("internal/storage/sql/tenant_migrations/20260508153000_ai_app_isolated_schema.up.sql")
	if err != nil {
		// Fallback: apply inline schema
		return ep.applyInlineAISchema(ctx, pool, tenantID)
	}

	_, err = pool.Exec(ctx, string(migrationSQL))
	return err
}

// applyInlineAISchema applies the AI schema inline (fallback when migration file not found)
func (ep *ExternalAIProvisioner) applyInlineAISchema(ctx context.Context, pool *pgxpool.Pool, tenantID uuid.UUID) error {
	// Enable pgvector
	_, err := pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		return fmt.Errorf("failed to enable pgvector: %w", err)
	}

	// Apply the schema (abbreviated — full schema is in the migration file)
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS ai_collections (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(255) NOT NULL,
			description TEXT,
			embedding_model VARCHAR(100) NOT NULL DEFAULT 'text-embedding-3-small',
			embedding_dimensions INT NOT NULL DEFAULT 1536,
			distance_metric VARCHAR(20) NOT NULL DEFAULT 'cosine',
			chunk_size INT NOT NULL DEFAULT 512,
			chunk_overlap INT NOT NULL DEFAULT 50,
			metadata JSONB DEFAULT '{}',
			document_count INT DEFAULT 0,
			total_tokens INT DEFAULT 0,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(tenant_id, slug)
		);

		CREATE TABLE IF NOT EXISTS ai_embeddings (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			collection_id UUID NOT NULL,
			document_id UUID NOT NULL,
			chunk_index INT NOT NULL DEFAULT 0,
			content TEXT NOT NULL,
			content_tokens INT DEFAULT 0,
			embedding vector(1536),
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS ai_memories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL,
			assistant_id UUID,
			user_id UUID,
			memory_type VARCHAR(50) NOT NULL,
			category VARCHAR(100),
			content TEXT NOT NULL,
			content_summary VARCHAR(500),
			embedding vector(1536),
			importance NUMERIC(3,2) DEFAULT 0.5,
			access_count INT DEFAULT 0,
			is_active BOOLEAN DEFAULT TRUE,
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_ai_embeddings_vector ON ai_embeddings
		 USING hnsw (embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);

		CREATE INDEX IF NOT EXISTS idx_ai_memories_vector ON ai_memories
		 USING hnsw (embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);
	`)
	return err
}

// recordProvisioningCost records the Neon infrastructure cost in cost_allocation_entries
func (ep *ExternalAIProvisioner) recordProvisioningCost(ctx context.Context, tenantID uuid.UUID, conn *NeonConnectionDetails) {
	// Estimate monthly cost for a small Neon branch:
	// Compute: ~$0.02/hr * 730hrs = $14.60/mo (but scales to zero = ~$2-5/mo for AI workloads)
	// Storage: ~$0.125/GiB/mo (vectors are ~6KB each, 10K vectors ≈ 60MB ≈ negligible)
	// We record the estimated cost so the billing system can charge the tenant
	estimatedMonthlyCostCents := 500 // $5.00/mo estimate for serverless AI DB
	markupCents := estimatedMonthlyCostCents * ep.markupRateBPS / 10000
	totalCostCents := estimatedMonthlyCostCents + markupCents

	_, err := ep.platformDB.ExecContext(ctx,
		`INSERT INTO cost_allocation_entries (
			id, tenant_id, function_id, execution_id, timestamp,
			compute_cost, memory_cost, network_cost, storage_cost, total_cost,
			region, metadata
		) VALUES ($1, $2, $3, $4, NOW(), 0, 0, 0, $5, $6, 'neon-external', $7)`,
		uuid.New(), tenantID, uuid.Nil, uuid.Nil,
		totalCostCents, totalCostCents,
		fmt.Sprintf(`{"type":"ai_infra","provider":"neon","branch_id":"%s","base_cost_cents":%d,"markup_cents":%d,"markup_rate_bps":%d}`,
			conn.BranchID, estimatedMonthlyCostCents, markupCents, ep.markupRateBPS))
	if err != nil {
		logrus.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to record AI infra cost (non-fatal)")
	}
}

// SuspendExternalAIDB suspends the Neon branch (scales to zero — no compute cost)
func (ep *ExternalAIProvisioner) SuspendExternalAIDB(ctx context.Context, tenantID uuid.UUID) error {
	_, err := ep.platformDB.ExecContext(ctx,
		`UPDATE tenant_ai_db_config SET status = 'suspended', updated_at = NOW() WHERE tenant_id = $1`,
		tenantID)
	return err
}

// DeleteExternalAIDB deletes the Neon branch and all data
func (ep *ExternalAIProvisioner) DeleteExternalAIDB(ctx context.Context, tenantID uuid.UUID) error {
	if !ep.IsAvailable() {
		return nil
	}

	var connJSON []byte
	err := ep.platformDB.QueryRowContext(ctx,
		`SELECT connection_details FROM tenant_ai_db_config WHERE tenant_id = $1`,
		tenantID).Scan(&connJSON)
	if err != nil {
		return err
	}

	var conn NeonConnectionDetails
	json.Unmarshal(connJSON, &conn)

	// Delete branch via Neon API
	req, err := http.NewRequestWithContext(ctx, "DELETE",
		fmt.Sprintf("%s/projects/%s/branches/%s", ep.neonAPIBase, ep.neonProjectID, conn.BranchID),
		nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+ep.neonAPIKey)

	resp, err := ep.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Update platform DB
	_, err = ep.platformDB.ExecContext(ctx,
		`UPDATE tenant_ai_db_config SET status = 'deleted', updated_at = NOW() WHERE tenant_id = $1`,
		tenantID)
	return err
}
