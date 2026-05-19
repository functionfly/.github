-- Create studio_tasks table for autonomous task management in Studio

CREATE TABLE IF NOT EXISTS studio_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'todo',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    assignee_id UUID,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    CONSTRAINT chk_studio_task_status CHECK (status IN ('todo', 'in-progress', 'done', 'blocked', 'review')),
    CONSTRAINT chk_studio_task_priority CHECK (priority IN ('low', 'medium', 'high'))
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_studio_tasks_tenant ON studio_tasks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_studio_tasks_status ON studio_tasks(status);
CREATE INDEX IF NOT EXISTS idx_studio_tasks_assignee ON studio_tasks(assignee_id);
CREATE INDEX IF NOT EXISTS idx_studio_tasks_created_by ON studio_tasks(created_by);
CREATE INDEX IF NOT EXISTS idx_studio_tasks_created_at ON studio_tasks(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_studio_tasks_tenant_status ON studio_tasks(tenant_id, status);

COMMENT ON TABLE studio_tasks IS 'Autonomous task management for FunctionFly Studio';