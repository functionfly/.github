/**
 * StateFabric - Composable Durable State for Stateless Functions
 *
 * A JavaScript/TypeScript SDK module for accessing and managing durable state
 * bound to function identities.
 *
 * Example:
 *     import { StateClient, StateManager } from '@functionfly/flypy';
 *
 *     // Direct client usage
 *     const client = new StateClient({ apiKey: 'your-api-key' });
 *
 *     // Get state value
 *     const cart = await client.getValue('my-tenant/cart/user123');
 *     console.log(cart);
 *
 *     // Set state value
 *     await client.setValue('my-tenant/cart/user123', { items: [{ id: 1, qty: 2 }] });
 *
 *     // Get state history
 *     const history = await client.getHistory('my-tenant/cart/user123');
 *
 *     // Create snapshot
 *     const snapshot = await client.createSnapshot('my-tenant/cart', { label: 'backup-001' });
 *
 *     // Restore from snapshot
 *     await client.restoreSnapshot('my-tenant/cart', { snapshotVersion: 1 });
 *
 *     // Using the state manager
 *     const manager = new StateManager({ tenantId: 'my-tenant' });
 *
 *     const cart = await manager.get('cart', 'user123');
 *     await manager.set('cart', 'user123', { items: [] });
 *     await manager.delete('cart', 'user123');
 */
import type { StateContainer, StateValue, StateEvent, StateSnapshot, StatePermission, StateTrigger, CreateStateRequest, CreateSnapshotRequest, RestoreSnapshotRequest, GrantPermissionRequest, CreateTriggerRequest, ListOptions, TimeTravelQuery } from './types.js';
export declare class StateError extends Error {
    statusCode?: number | undefined;
    constructor(message: string, statusCode?: number | undefined);
}
export declare class StateNotFoundError extends StateError {
    constructor(message: string);
}
export declare class StatePermissionError extends StateError {
    constructor(message: string);
}
export interface StateClientOptions {
    apiUrl?: string;
    tenantId?: string;
    apiKey?: string;
}
export interface GetValueOptions {
    includeMetadata?: boolean;
}
/**
 * Client for interacting with StateFabric API.
 *
 * Provides methods for CRUD operations on state, value management,
 * history tracking, snapshots, and permissions.
 */
export declare class StateClient {
    private client;
    private tenantId;
    constructor(options?: StateClientOptions);
    /**
     * Create a new state container.
     */
    createState(path: string, request: CreateStateRequest): Promise<StateContainer>;
    /**
     * Get a state container by path.
     */
    getState(path: string): Promise<StateContainer>;
    /**
     * List all states for a tenant.
     */
    listStates(options?: ListOptions): Promise<StateContainer[]>;
    /**
     * Delete a state container.
     */
    deleteState(path: string): Promise<void>;
    /**
     * Set a value in the state.
     */
    setValue(path: string, value: Record<string, unknown>, metadata?: Record<string, unknown>): Promise<StateValue>;
    /**
     * Get a value from the state.
     */
    getValue(path: string, options?: GetValueOptions): Promise<StateValue | null>;
    /**
     * Delete a value from the state.
     */
    deleteValue(path: string): Promise<void>;
    /**
     * Get all key-value pairs in a state container.
     */
    getAllValues(statePath: string): Promise<Record<string, unknown>>;
    /**
     * Get the event history for a state path.
     */
    getHistory(path: string, options?: ListOptions & {
        eventTypes?: string[];
    }): Promise<StateEvent[]>;
    /**
     * Query state at a specific point in time or version.
     */
    timeTravel(path: string, query: TimeTravelQuery): Promise<StateValue | null>;
    /**
     * Create a snapshot of the current state.
     */
    createSnapshot(path: string, request?: CreateSnapshotRequest): Promise<StateSnapshot>;
    /**
     * List all snapshots for a state container.
     */
    listSnapshots(path: string, options?: ListOptions): Promise<StateSnapshot[]>;
    /**
     * Restore state from a snapshot.
     */
    restoreSnapshot(path: string, request: RestoreSnapshotRequest): Promise<StateContainer>;
    /**
     * Grant permission to access a state.
     */
    grantPermission(path: string, request: GrantPermissionRequest): Promise<StatePermission>;
    /**
     * Get all permissions for a state.
     */
    getPermissions(path: string): Promise<StatePermission[]>;
    /**
     * Create a trigger for state changes.
     */
    createTrigger(request: CreateTriggerRequest): Promise<StateTrigger>;
    /**
     * Get triggers, optionally filtered by state path.
     */
    getTriggers(options?: {
        statePath?: string;
        isActive?: boolean;
    }): Promise<StateTrigger[]>;
    /**
     * Delete a trigger.
     */
    deleteTrigger(triggerId: string): Promise<void>;
}
/**
 * State manager for declarative state access.
 *
 * Provides a simpler interface for managing state with
 * convenience methods.
 */
export declare class StateManager {
    private client;
    private statePrefix;
    constructor(options?: StateClientOptions);
    /**
     * Get a value from state.
     */
    get(stateName: string, key: string, defaultValue?: unknown): Promise<unknown>;
    /**
     * Set a value in state.
     */
    set(stateName: string, key: string, value: unknown): Promise<StateValue>;
    /**
     * Delete a value from state.
     */
    delete(stateName: string, key: string): Promise<void>;
    /**
     * Create a snapshot of a state container.
     */
    snapshot(stateName: string, label?: string): Promise<StateSnapshot>;
    /**
     * Restore a state container from snapshot.
     */
    restore(stateName: string, snapshotVersion: number): Promise<StateContainer>;
    /**
     * Get the underlying client for advanced operations.
     */
    getClient(): StateClient;
}
export declare function getValue(path: string, options?: StateClientOptions): Promise<unknown>;
export declare function setValue(path: string, value: Record<string, unknown>, options?: StateClientOptions): Promise<StateValue>;
export declare function deleteValue(path: string, options?: StateClientOptions): Promise<void>;
export declare function getHistory(path: string, options?: StateClientOptions & ListOptions): Promise<StateEvent[]>;
export declare function createSnapshot(path: string, request?: CreateSnapshotRequest, options?: StateClientOptions): Promise<StateSnapshot>;
export declare function restoreSnapshot(path: string, request: RestoreSnapshotRequest, options?: StateClientOptions): Promise<StateContainer>;
//# sourceMappingURL=state.d.ts.map