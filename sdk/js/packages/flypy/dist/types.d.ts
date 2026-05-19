/**
 * Core types for FunctionFly SDK
 */
export interface FunctionMetadata {
    name: string;
    version?: string;
    description?: string;
    deterministic?: boolean;
    idempotent?: boolean;
    pure?: boolean;
    cacheTtl?: number;
    capabilities?: string[];
    maxExecutionTime?: number;
    inputSchema?: Record<string, unknown>;
    outputSchema?: Record<string, unknown>;
}
export interface FunctionDefinition {
    metadata: FunctionMetadata;
    sourceCode: string;
    astJson?: string;
    dependencies?: string[];
    imports?: string[];
}
export interface BuildResult {
    success: boolean;
    functionName: string;
    outputDir: string;
    warnings: string[];
    errors: string[];
    wasmFile?: string;
    manifestFile?: string;
    determinismHash?: string;
    buildTimeMs?: number;
    wasmSizeBytes?: number;
    optimizationStats?: Record<string, unknown>;
    bundleAnalysis?: Record<string, unknown>;
    coldStartStats?: Record<string, unknown>;
}
export interface StateContainer {
    id: string;
    tenantId: string;
    name: string;
    path: string;
    storageType: 'durable' | 'ephemeral' | 'cached';
    ttlDays?: number;
    maxSizeMB?: number;
    isVersioned: boolean;
    isPublic: boolean;
    description?: string;
    tags?: Record<string, unknown>;
    createdAt: string;
    updatedAt: string;
}
export interface StateValue {
    id: string;
    stateId: string;
    key: string;
    value: Record<string, unknown>;
    version: number;
    metadata?: Record<string, unknown>;
    createdAt: string;
    updatedAt: string;
}
export interface StateEvent {
    id: string;
    stateId: string;
    key: string;
    eventType: 'set' | 'delete' | 'restore' | 'snapshot';
    previousValue?: Record<string, unknown>;
    newValue?: Record<string, unknown>;
    version: number;
    actor?: string;
    timestamp: string;
}
export interface StateSnapshot {
    id: string;
    stateId: string;
    version: number;
    label?: string;
    values: Record<string, unknown>;
    createdAt: string;
}
export interface StatePermission {
    id: string;
    stateId: string;
    principalType: 'user' | 'team' | 'service';
    principalId: string;
    canRead: boolean;
    canWrite: boolean;
    canDelete: boolean;
    canAdmin: boolean;
    canTrigger: boolean;
    createdAt: string;
}
export interface StateTrigger {
    id: string;
    stateId: string;
    triggerType: 'on_set' | 'on_delete' | 'on_change';
    keyPattern?: string;
    condition?: Record<string, unknown>;
    targetFunctionId?: string;
    targetFunction: string;
    includePrevious: boolean;
    includeNew: boolean;
    maxInvocationsPerMinute: number;
    isActive: boolean;
    lastFiredAt?: string;
    createdAt: string;
    updatedAt: string;
}
export interface CreateStateRequest {
    name: string;
    storageType?: 'durable' | 'ephemeral' | 'cached';
    ttlDays?: number;
    maxSizeMB?: number;
    isVersioned?: boolean;
    isPublic?: boolean;
    description?: string;
    tags?: Record<string, unknown>;
}
export interface SetValueRequest {
    value: Record<string, unknown>;
    metadata?: Record<string, unknown>;
}
export interface CreateSnapshotRequest {
    label?: string;
}
export interface RestoreSnapshotRequest {
    snapshotVersion: number;
}
export interface GrantPermissionRequest {
    principalType: 'user' | 'team' | 'service';
    principalId: string;
    canRead?: boolean;
    canWrite?: boolean;
    canDelete?: boolean;
    canAdmin?: boolean;
    canTrigger?: boolean;
}
export interface CreateTriggerRequest {
    triggerType: 'on_set' | 'on_delete' | 'on_change';
    keyPattern?: string;
    condition?: Record<string, unknown>;
    targetFunctionId?: string;
    targetFunction: string;
    includePrevious?: boolean;
    includeNew?: boolean;
    maxInvocationsPerMinute?: number;
    isActive?: boolean;
}
export interface ListOptions {
    limit?: number;
    offset?: number;
}
export interface TimeTravelQuery {
    timestamp?: string;
    version?: number;
}
//# sourceMappingURL=types.d.ts.map