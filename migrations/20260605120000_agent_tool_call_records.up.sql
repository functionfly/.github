-- =============================================================================
-- Migration: 20260605120000_agent_tool_call_records.up.sql
-- Purpose: Persist agent tool call attribution records for policy, billing,
--          and observability. These records power tool usage dashboards,
--          spend tracking per tool, and replay/debug workflows.
-- =============================================================================

-- Records for every tool invocation executed through HandleExecute tool chains
-- or the direct /agent/tools/{name}/call endpoint.

create table if not exists agent_tool_call_records (
    id           uuid primary key default gen_random_uuid(),
    agent_id     varchar(255) not null,
    tenant_id    uuid not null,
    session_id   varchar(255),
    execution_id varchar(255),
    call_depth   int not null default 0,
    tool_name    varchar(255) not null,
    input_hash   varchar(255),
    output_hash  varchar(255),
    latency_ms   bigint not null default 0,
    outcome      varchar(64) not null default 'success',
    error_code   varchar(255),
    cost_usd     numeric(10,6) not null default 0,
    timestamp    timestamp not null default now()
);

create index if not exists idx_agent_tool_calls_agent_ts
    on agent_tool_call_records (agent_id, timestamp desc);

create index if not exists idx_agent_tool_calls_tenant_ts
    on agent_tool_call_records (tenant_id, timestamp desc);

create index if not exists idx_agent_tool_calls_execution
    on agent_tool_call_records (execution_id);

create index if not exists idx_agent_tool_calls_tool
    on agent_tool_call_records (tool_name);
