-- Add verification and tenant/user columns to registry_function_executions
-- and create execution_resource_usages table for resource tracking.

-- Columns added by this migration may already exist if a later schema was applied;
-- use DO blocks to add columns only when missing (idempotent).

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'registry_function_executions' AND column_name = 'tenant_id'
  ) THEN
    ALTER TABLE registry_function_executions ADD COLUMN tenant_id UUID;
    CREATE INDEX IF NOT EXISTS idx_registry_function_executions_tenant_id ON registry_function_executions(tenant_id) WHERE tenant_id IS NOT NULL;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'registry_function_executions' AND column_name = 'user_id'
  ) THEN
    ALTER TABLE registry_function_executions ADD COLUMN user_id UUID;
    CREATE INDEX IF NOT EXISTS idx_registry_function_executions_user_id ON registry_function_executions(user_id) WHERE user_id IS NOT NULL;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'registry_function_executions' AND column_name = 'verified_at'
  ) THEN
    ALTER TABLE registry_function_executions ADD COLUMN verified_at TIMESTAMP WITH TIME ZONE;
    CREATE INDEX IF NOT EXISTS idx_registry_function_executions_verified_at ON registry_function_executions(verified_at DESC) WHERE verified_at IS NOT NULL;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'registry_function_executions' AND column_name = 'verification_status'
  ) THEN
    ALTER TABLE registry_function_executions ADD COLUMN verification_status TEXT;
    CREATE INDEX IF NOT EXISTS idx_registry_function_executions_verification_status ON registry_function_executions(verification_status) WHERE verification_status IS NOT NULL;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'registry_function_executions' AND column_name = 'verification_error'
  ) THEN
    ALTER TABLE registry_function_executions ADD COLUMN verification_error TEXT;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'registry_function_executions' AND column_name = 'replayed_duration_ms'
  ) THEN
    ALTER TABLE registry_function_executions ADD COLUMN replayed_duration_ms INTEGER;
  END IF;
END $$;

-- Relax outcome check if it only allows 'success','error','timeout' (some code may send 'execution_failed')
-- Leave outcome as-is; if needed, a separate migration can alter the check.

-- Create execution_resource_usages table (GORM default table name for ExecutionResourceUsage)
CREATE TABLE IF NOT EXISTS execution_resource_usages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID REFERENCES registry_function_executions(id) ON DELETE SET NULL,

    max_memory_mb INTEGER NOT NULL DEFAULT 128,
    max_cpu_time_ms INTEGER NOT NULL DEFAULT 5000,

    memory_used_mb DOUBLE PRECISION NOT NULL DEFAULT 0,
    cpu_time_used_ms INTEGER NOT NULL DEFAULT 0,
    wall_time_used_ms INTEGER NOT NULL DEFAULT 0,

    terminated_by VARCHAR(50) DEFAULT 'normal',

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_execution_resource_usages_execution_id ON execution_resource_usages(execution_id) WHERE execution_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_execution_resource_usages_created_at ON execution_resource_usages(created_at DESC);
