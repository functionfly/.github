//! PostgreSQL metadata storage

use sqlx::{postgres::PgPool, Row};
use uuid::Uuid as UuidTrait;

use crate::models::{EventMetadata, SnapshotMetadata, State};

/// Repository for state metadata in PostgreSQL
#[derive(Debug)]
pub struct PostgresStateRepository {
    pool: PgPool,
}

impl PostgresStateRepository {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }

    /// Create a new state
    pub async fn create(&self, state: &State) -> Result<(), sqlx::Error> {
        sqlx::query(
            r#"
            INSERT INTO states (
                id, tenant_id, path, full_path, current_version,
                storage_type, state_hash, size_bytes, key_count,
                deterministic, agent_id, config,
                created_at, updated_at, last_accessed_at
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
            "#,
        )
        .bind(state.id)
        .bind(state.tenant_id)
        .bind(&state.path)
        .bind(&state.full_path)
        .bind(state.current_version)
        .bind(&state.storage_type)
        .bind(&state.state_hash)
        .bind(state.size_bytes)
        .bind(state.key_count)
        .bind(state.deterministic)
        .bind(state.agent_id)
        .bind(&state.config)
        .bind(state.created_at)
        .bind(state.updated_at)
        .bind(state.last_accessed_at)
        .execute(&self.pool)
        .await?;

        Ok(())
    }

    /// Get state by ID
    pub async fn get_by_id(&self, id: UuidTrait) -> Result<Option<State>, sqlx::Error> {
        let row = sqlx::query(
            r#"
            SELECT id, tenant_id, path, full_path, current_version,
                   storage_type, state_hash, size_bytes, key_count,
                   deterministic, agent_id, config,
                   created_at, updated_at, last_accessed_at
            FROM states WHERE id = $1
            "#,
        )
        .bind(id)
        .fetch_optional(&self.pool)
        .await?;

        Ok(row.map(|r| State {
            id: r.get("id"),
            tenant_id: r.get("tenant_id"),
            path: r.get("path"),
            full_path: r.get("full_path"),
            current_version: r.get("current_version"),
            storage_type: r.get("storage_type"),
            state_hash: r.get("state_hash"),
            size_bytes: r.get("size_bytes"),
            key_count: r.get("key_count"),
            deterministic: r.get("deterministic"),
            agent_id: r.get("agent_id"),
            config: r.get("config"),
            created_at: r.get("created_at"),
            updated_at: r.get("updated_at"),
            last_accessed_at: r.get("last_accessed_at"),
        }))
    }

    /// Update state version and hash
    pub async fn update_version(
        &self,
        id: UuidTrait,
        version: i64,
        state_hash: &str,
        size_bytes: i64,
        key_count: i32,
    ) -> Result<(), sqlx::Error> {
        sqlx::query(
            r#"
            UPDATE states
            SET current_version = $2, state_hash = $3, size_bytes = $4, key_count = $5,
                updated_at = NOW()
            WHERE id = $1
            "#,
        )
        .bind(id)
        .bind(version)
        .bind(state_hash)
        .bind(size_bytes)
        .bind(key_count)
        .execute(&self.pool)
        .await?;

        Ok(())
    }

    /// SECURITY P0: Verify a state_id belongs to the given tenant.
    ///
    /// Returns `Ok(true)` if the state exists and `tenant_id` matches,
    /// `Ok(false)` if the state does not exist OR belongs to a different tenant
    /// (we deliberately do not differentiate to avoid an enumeration oracle).
    ///
    /// This is the authoritative cross-tenant access gate.
    pub async fn verify_tenant_ownership(
        &self,
        state_id: UuidTrait,
        tenant_id: UuidTrait,
    ) -> Result<bool, sqlx::Error> {
        if tenant_id.is_nil() {
            return Ok(false);
        }

        let row = sqlx::query(
            r#"
            SELECT EXISTS(
                SELECT 1 FROM states
                WHERE id = $1 AND tenant_id = $2
            ) AS owned
            "#,
        )
        .bind(state_id)
        .bind(tenant_id)
        .fetch_one(&self.pool)
        .await?;

        Ok(row.get::<bool, _>("owned"))
    }
}

/// Repository for event metadata in PostgreSQL
#[derive(Debug)]
pub struct PostgresEventRepository {
    pool: PgPool,
}

impl PostgresEventRepository {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }

    /// Insert event metadata (actual event data goes to object storage)
    pub async fn insert_metadata(&self, metadata: &EventMetadata) -> Result<(), sqlx::Error> {
        sqlx::query(
            r#"
            INSERT INTO event_metadata (
                id, state_id, event_type, key, correlation_id,
                source_type, source_id, deterministic, sequence_num,
                timestamp, storage_key, input_hash, output_hash
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
            "#,
        )
        .bind(metadata.id)
        .bind(metadata.state_id)
        .bind(&metadata.event_type)
        .bind(&metadata.key)
        .bind(&metadata.correlation_id)
        .bind(&metadata.source_type)
        .bind(&metadata.source_id)
        .bind(metadata.deterministic)
        .bind(metadata.sequence_num)
        .bind(metadata.timestamp)
        .bind(&metadata.storage_key)
        .bind(&metadata.input_hash)
        .bind(&metadata.output_hash)
        .execute(&self.pool)
        .await?;

        Ok(())
    }

    /// Get event metadata by ID
    pub async fn get_metadata(&self, id: UuidTrait) -> Result<Option<EventMetadata>, sqlx::Error> {
        let row = sqlx::query(
            r#"
            SELECT id, state_id, event_type, key, correlation_id,
                   source_type, source_id, deterministic, sequence_num,
                   timestamp, storage_key, input_hash, output_hash
            FROM event_metadata WHERE id = $1
            "#,
        )
        .bind(id)
        .fetch_optional(&self.pool)
        .await?;

        Ok(row.map(|r| EventMetadata {
            id: r.get("id"),
            state_id: r.get("state_id"),
            event_type: r.get("event_type"),
            key: r.get("key"),
            correlation_id: r.get("correlation_id"),
            source_type: r.get("source_type"),
            source_id: r.get("source_id"),
            deterministic: r.get("deterministic"),
            sequence_num: r.get("sequence_num"),
            timestamp: r.get("timestamp"),
            storage_key: r.get("storage_key"),
            input_hash: r.get("input_hash"),
            output_hash: r.get("output_hash"),
        }))
    }

    /// Get event metadata by state ID and sequence range
    pub async fn get_range(
        &self,
        state_id: UuidTrait,
        from_seq: i64,
        to_seq: i64,
    ) -> Result<Vec<EventMetadata>, sqlx::Error> {
        let rows = sqlx::query(
            r#"
            SELECT id, state_id, event_type, key, correlation_id,
                   source_type, source_id, deterministic, sequence_num,
                   timestamp, storage_key, input_hash, output_hash
            FROM event_metadata
            WHERE state_id = $1 AND sequence_num >= $2 AND sequence_num <= $3
            ORDER BY sequence_num ASC
            "#,
        )
        .bind(state_id)
        .bind(from_seq)
        .bind(to_seq)
        .fetch_all(&self.pool)
        .await?;

        Ok(rows.into_iter().map(|r| EventMetadata {
            id: r.get("id"),
            state_id: r.get("state_id"),
            event_type: r.get("event_type"),
            key: r.get("key"),
            correlation_id: r.get("correlation_id"),
            source_type: r.get("source_type"),
            source_id: r.get("source_id"),
            deterministic: r.get("deterministic"),
            sequence_num: r.get("sequence_num"),
            timestamp: r.get("timestamp"),
            storage_key: r.get("storage_key"),
            input_hash: r.get("input_hash"),
            output_hash: r.get("output_hash"),
        }).collect())
    }

    /// Get next sequence number for a state
    pub async fn get_next_sequence(&self, state_id: UuidTrait) -> Result<i64, sqlx::Error> {
        let row = sqlx::query(
            r#"
            SELECT COALESCE(MAX(sequence_num), 0) + 1 as next_seq
            FROM event_metadata WHERE state_id = $1
            "#,
        )
        .bind(state_id)
        .fetch_one(&self.pool)
        .await?;

        Ok(row.get("next_seq"))
    }
}

/// Repository for snapshot metadata in PostgreSQL
#[derive(Debug)]
pub struct PostgresSnapshotRepository {
    pool: PgPool,
}

impl PostgresSnapshotRepository {
    pub fn new(pool: PgPool) -> Self {
        Self { pool }
    }

    /// Insert snapshot metadata
    pub async fn insert_metadata(&self, metadata: &SnapshotMetadata) -> Result<(), sqlx::Error> {
        sqlx::query(
            r#"
            INSERT INTO snapshot_metadata (
                id, state_id, snapshot_version, label, key_count,
                size_bytes, first_sequence, last_sequence, root_event_id,
                is_compressed, compression_algo, checksum, storage_key, created_at
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
            "#,
        )
        .bind(metadata.id)
        .bind(metadata.state_id)
        .bind(metadata.snapshot_version)
        .bind(&metadata.label)
        .bind(metadata.key_count)
        .bind(metadata.size_bytes)
        .bind(metadata.first_sequence)
        .bind(metadata.last_sequence)
        .bind(metadata.root_event_id)
        .bind(metadata.is_compressed)
        .bind(&metadata.compression_algo)
        .bind(&metadata.checksum)
        .bind(&metadata.storage_key)
        .bind(metadata.created_at)
        .execute(&self.pool)
        .await?;

        Ok(())
    }

    /// Get latest snapshot for a state
    pub async fn get_latest(&self, state_id: UuidTrait) -> Result<Option<SnapshotMetadata>, sqlx::Error> {
        let row = sqlx::query(
            r#"
            SELECT id, state_id, snapshot_version, label, key_count,
                   size_bytes, first_sequence, last_sequence, root_event_id,
                   is_compressed, compression_algo, checksum, storage_key, created_at
            FROM snapshot_metadata
            WHERE state_id = $1
            ORDER BY snapshot_version DESC
            LIMIT 1
            "#,
        )
        .bind(state_id)
        .fetch_optional(&self.pool)
        .await?;

        Ok(row.map(|r| SnapshotMetadata {
            id: r.get("id"),
            state_id: r.get("state_id"),
            snapshot_version: r.get("snapshot_version"),
            label: r.get("label"),
            key_count: r.get("key_count"),
            size_bytes: r.get("size_bytes"),
            first_sequence: r.get("first_sequence"),
            last_sequence: r.get("last_sequence"),
            root_event_id: r.get("root_event_id"),
            is_compressed: r.get("is_compressed"),
            compression_algo: r.get("compression_algo"),
            checksum: r.get("checksum"),
            storage_key: r.get("storage_key"),
            created_at: r.get("created_at"),
        }))
    }

    /// Get snapshot by version
    pub async fn get_by_version(
        &self,
        state_id: UuidTrait,
        version: i64,
    ) -> Result<Option<SnapshotMetadata>, sqlx::Error> {
        let row = sqlx::query(
            r#"
            SELECT id, state_id, snapshot_version, label, key_count,
                   size_bytes, first_sequence, last_sequence, root_event_id,
                   is_compressed, compression_algo, checksum, storage_key, created_at
            FROM snapshot_metadata
            WHERE state_id = $1 AND snapshot_version = $2
            "#,
        )
        .bind(state_id)
        .bind(version)
        .fetch_optional(&self.pool)
        .await?;

        Ok(row.map(|r| SnapshotMetadata {
            id: r.get("id"),
            state_id: r.get("state_id"),
            snapshot_version: r.get("snapshot_version"),
            label: r.get("label"),
            key_count: r.get("key_count"),
            size_bytes: r.get("size_bytes"),
            first_sequence: r.get("first_sequence"),
            last_sequence: r.get("last_sequence"),
            root_event_id: r.get("root_event_id"),
            is_compressed: r.get("is_compressed"),
            compression_algo: r.get("compression_algo"),
            checksum: r.get("checksum"),
            storage_key: r.get("storage_key"),
            created_at: r.get("created_at"),
        }))
    }

    /// Get snapshot by ID
    pub async fn get_by_id(&self, id: UuidTrait) -> Result<Option<SnapshotMetadata>, sqlx::Error> {
        let row = sqlx::query(
            r#"
            SELECT id, state_id, snapshot_version, label, key_count,
                   size_bytes, first_sequence, last_sequence, root_event_id,
                   is_compressed, compression_algo, checksum, storage_key, created_at
            FROM snapshot_metadata
            WHERE id = $1
            "#,
        )
        .bind(id)
        .fetch_optional(&self.pool)
        .await?;

        Ok(row.map(|r| SnapshotMetadata {
            id: r.get("id"),
            state_id: r.get("state_id"),
            snapshot_version: r.get("snapshot_version"),
            label: r.get("label"),
            key_count: r.get("key_count"),
            size_bytes: r.get("size_bytes"),
            first_sequence: r.get("first_sequence"),
            last_sequence: r.get("last_sequence"),
            root_event_id: r.get("root_event_id"),
            is_compressed: r.get("is_compressed"),
            compression_algo: r.get("compression_algo"),
            checksum: r.get("checksum"),
            storage_key: r.get("storage_key"),
            created_at: r.get("created_at"),
        }))
    }

    /// Get all snapshots for a state, ordered by version descending (newest first)
    pub async fn get_all(&self, state_id: UuidTrait) -> Result<Vec<SnapshotMetadata>, sqlx::Error> {
        let rows = sqlx::query(
            r#"
            SELECT id, state_id, snapshot_version, label, key_count,
                   size_bytes, first_sequence, last_sequence, root_event_id,
                   is_compressed, compression_algo, checksum, storage_key, created_at
            FROM snapshot_metadata
            WHERE state_id = $1
            ORDER BY snapshot_version DESC
            "#,
        )
        .bind(state_id)
        .fetch_all(&self.pool)
        .await?;

        Ok(rows.into_iter().map(|r| SnapshotMetadata {
            id: r.get("id"),
            state_id: r.get("state_id"),
            snapshot_version: r.get("snapshot_version"),
            label: r.get("label"),
            key_count: r.get("key_count"),
            size_bytes: r.get("size_bytes"),
            first_sequence: r.get("first_sequence"),
            last_sequence: r.get("last_sequence"),
            root_event_id: r.get("root_event_id"),
            is_compressed: r.get("is_compressed"),
            compression_algo: r.get("compression_algo"),
            checksum: r.get("checksum"),
            storage_key: r.get("storage_key"),
            created_at: r.get("created_at"),
        }).collect())
    }

    /// Get next snapshot version for a state
    pub async fn get_next_version(&self, state_id: UuidTrait) -> Result<i64, sqlx::Error> {
        let row = sqlx::query(
            r#"
            SELECT COALESCE(MAX(snapshot_version), 0) + 1 as next_ver
            FROM snapshot_metadata WHERE state_id = $1
            "#,
        )
        .bind(state_id)
        .fetch_one(&self.pool)
        .await?;

        Ok(row.get("next_ver"))
    }
}
