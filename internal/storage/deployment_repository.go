package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DeploymentRepository handles deployment and artifact operations
type DeploymentRepository struct {
	db *PostgresDB
}

// NewDeploymentRepository creates a new deployment repository
func NewDeploymentRepository(db *PostgresDB) *DeploymentRepository {
	return &DeploymentRepository{db: db}
}

// CreateDeployment creates a new deployment
func (r *DeploymentRepository) CreateDeployment(ctx context.Context, appID uuid.UUID, provider, region, deploymentID, artifactKey string, routes []string) (*Deployment, error) {
	routesJSON, err := json.Marshal(routes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal routes: %w", err)
	}

	deployment := &Deployment{
		ID:           uuid.New(),
		AppID:        appID,
		Provider:     provider,
		Region:       region,
		DeploymentID: deploymentID,
		Status:       "pending",
		ArtifactKey:  artifactKey,
		Routes:       routes,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO deployments (id, app_id, provider, region, deployment_id, status, artifact_key, routes, message, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		deployment.ID, deployment.AppID, deployment.Provider, deployment.Region,
		deployment.DeploymentID, deployment.Status, deployment.ArtifactKey,
		string(routesJSON), deployment.Message, deployment.Metadata,
		deployment.CreatedAt, deployment.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create deployment: %w", err)
	}

	return deployment, nil
}

// UpdateDeploymentStatus updates deployment status
func (r *DeploymentRepository) UpdateDeploymentStatus(ctx context.Context, id uuid.UUID, status, message string, metadata map[string]interface{}) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE deployments
		SET status = $2, message = $3, metadata = $4, updated_at = $5
		WHERE id = $1`,
		id, status, message, string(metadataJSON), time.Now())

	if err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	return nil
}

// GetDeploymentByID retrieves a deployment by ID
func (r *DeploymentRepository) GetDeploymentByID(ctx context.Context, id uuid.UUID) (*Deployment, error) {
	var deployment Deployment
	var routesJSON string
	var metadataJSON string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, app_id, provider, region, deployment_id, status, artifact_key, routes, message, metadata, created_at, updated_at
		FROM deployments WHERE id = $1`, id).Scan(
		&deployment.ID, &deployment.AppID, &deployment.Provider, &deployment.Region,
		&deployment.DeploymentID, &deployment.Status, &deployment.ArtifactKey,
		&routesJSON, &deployment.Message, &metadataJSON,
		&deployment.CreatedAt, &deployment.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	// Unmarshal routes
	if err := json.Unmarshal([]byte(routesJSON), &deployment.Routes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal routes: %w", err)
	}

	return &deployment, nil
}

// ListDeploymentsByAppID lists deployments for an app
func (r *DeploymentRepository) ListDeploymentsByAppID(ctx context.Context, appID uuid.UUID, limit int) ([]*Deployment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, app_id, provider, region, deployment_id, status, artifact_key, routes, message, metadata, created_at, updated_at
		FROM deployments
		WHERE app_id = $1
		ORDER BY created_at DESC
		LIMIT $2`, appID, limit)

	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}
	defer rows.Close()

	var deployments []*Deployment
	for rows.Next() {
		var deployment Deployment
		var routesJSON string
		var metadataJSON string

		err := rows.Scan(
			&deployment.ID, &deployment.AppID, &deployment.Provider, &deployment.Region,
			&deployment.DeploymentID, &deployment.Status, &deployment.ArtifactKey,
			&routesJSON, &deployment.Message, &metadataJSON,
			&deployment.CreatedAt, &deployment.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment: %w", err)
		}

		// Unmarshal routes
		if err := json.Unmarshal([]byte(routesJSON), &deployment.Routes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal routes: %w", err)
		}

		deployments = append(deployments, &deployment)
	}

	return deployments, nil
}

// GetLatestSuccessfulDeployment gets the latest successful deployment for an app and provider
func (r *DeploymentRepository) GetLatestSuccessfulDeployment(ctx context.Context, appID uuid.UUID, provider string) (*Deployment, error) {
	var deployment Deployment
	var routesJSON string
	var metadataJSON string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, app_id, provider, region, deployment_id, status, artifact_key, routes, message, metadata, created_at, updated_at
		FROM deployments
		WHERE app_id = $1 AND provider = $2 AND status = 'success'
		ORDER BY created_at DESC
		LIMIT 1`, appID, provider).Scan(
		&deployment.ID, &deployment.AppID, &deployment.Provider, &deployment.Region,
		&deployment.DeploymentID, &deployment.Status, &deployment.ArtifactKey,
		&routesJSON, &deployment.Message, &metadataJSON,
		&deployment.CreatedAt, &deployment.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest successful deployment: %w", err)
	}

	// Unmarshal routes
	if err := json.Unmarshal([]byte(routesJSON), &deployment.Routes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal routes: %w", err)
	}

	return &deployment, nil
}

// StoreDeploymentArtifact stores a deployment artifact
func (r *DeploymentRepository) StoreDeploymentArtifact(ctx context.Context, appID uuid.UUID, provider, key, contentType, checksum string, size int64) (*DeploymentArtifact, error) {
	artifact := &DeploymentArtifact{
		Key:         key,
		AppID:       appID,
		Provider:    provider,
		ContentType: contentType,
		Size:        size,
		Checksum:    checksum,
		CreatedAt:   time.Now(),
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO deployment_artifacts (key, app_id, provider, content_type, size, checksum, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		artifact.Key, artifact.AppID, artifact.Provider, artifact.ContentType,
		artifact.Size, artifact.Checksum, artifact.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to store deployment artifact: %w", err)
	}

	return artifact, nil
}

// GetDeploymentArtifact retrieves a deployment artifact by key
func (r *DeploymentRepository) GetDeploymentArtifact(ctx context.Context, key string) (*DeploymentArtifact, error) {
	var artifact DeploymentArtifact

	err := r.db.QueryRowContext(ctx, `
		SELECT key, app_id, provider, content_type, size, checksum, created_at
		FROM deployment_artifacts WHERE key = $1`, key).Scan(
		&artifact.Key, &artifact.AppID, &artifact.Provider, &artifact.ContentType,
		&artifact.Size, &artifact.Checksum, &artifact.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get deployment artifact: %w", err)
	}

	return &artifact, nil
}