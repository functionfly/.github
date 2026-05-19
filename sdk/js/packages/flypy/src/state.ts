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

import axios, { AxiosInstance, AxiosError } from 'axios';
import type {
  StateContainer,
  StateValue,
  StateEvent,
  StateSnapshot,
  StatePermission,
  StateTrigger,
  CreateStateRequest,
  SetValueRequest,
  CreateSnapshotRequest,
  RestoreSnapshotRequest,
  GrantPermissionRequest,
  CreateTriggerRequest,
  ListOptions,
  TimeTravelQuery,
} from './types.js';

const DEFAULT_API_URL = process.env.FLYPY_API_URL || 'http://localhost:8080/api';

export class StateError extends Error {
  constructor(message: string, public statusCode?: number) {
    super(message);
    this.name = 'StateError';
  }
}

export class StateNotFoundError extends StateError {
  constructor(message: string) {
    super(message, 404);
    this.name = 'StateNotFoundError';
  }
}

export class StatePermissionError extends StateError {
  constructor(message: string) {
    super(message, 403);
    this.name = 'StatePermissionError';
  }
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
export class StateClient {
  private client: AxiosInstance;
  private tenantId: string;

  constructor(options: StateClientOptions = {}) {
    const apiUrl = options.apiUrl || DEFAULT_API_URL;
    this.tenantId = options.tenantId || '';

    this.client = axios.create({
      baseURL: apiUrl,
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        ...(options.apiKey && { 'Authorization': `Bearer ${options.apiKey}` }),
      },
    });

    // Add response interceptor for error handling
    this.client.interceptors.response.use(
      response => response,
      (error: AxiosError) => {
        if (error.response) {
          const status = error.response.status;
          const data = error.response.data as Record<string, unknown>;
          const message = (data?.error as string) || error.message;

          if (status === 404) {
            throw new StateNotFoundError(message);
          } else if (status === 403) {
            throw new StatePermissionError(message);
          } else {
            throw new StateError(`HTTP ${status}: ${message}`, status);
          }
        } else if (error.request) {
          throw new StateError(`Network error: ${error.message}`);
        }
        throw error;
      }
    );
  }

  // State Management

  /**
   * Create a new state container.
   */
  async createState(path: string, request: CreateStateRequest): Promise<StateContainer> {
    const response = await this.client.post<StateContainer>('/v1/state', request);
    return response.data;
  }

  /**
   * Get a state container by path.
   */
  async getState(path: string): Promise<StateContainer> {
    const response = await this.client.get<StateContainer>(`/v1/state/${path}`);
    return response.data;
  }

  /**
   * List all states for a tenant.
   */
  async listStates(options: ListOptions = {}): Promise<StateContainer[]> {
    const params = {
      tenant_id: this.tenantId,
      ...options,
    };
    const response = await this.client.get<StateContainer[]>('/v1/state', { params });
    return response.data;
  }

  /**
   * Delete a state container.
   */
  async deleteState(path: string): Promise<void> {
    await this.client.delete(`/v1/state/${path}`);
  }

  // Value Operations

  /**
   * Set a value in the state.
   */
  async setValue(path: string, value: Record<string, unknown>, metadata?: Record<string, unknown>): Promise<StateValue> {
    const request: SetValueRequest = { value };
    if (metadata) {
      request.metadata = metadata;
    }
    const response = await this.client.put<StateValue>(`/v1/state/${path}/value`, request);
    return response.data;
  }

  /**
   * Get a value from the state.
   */
  async getValue(path: string, options: GetValueOptions = {}): Promise<StateValue | null> {
    try {
      const response = await this.client.get<StateValue>(`/v1/state/${path}/value`, {
        params: { include_metadata: options.includeMetadata },
      });
      return response.data;
    } catch (error) {
      if (error instanceof StateNotFoundError) {
        return null;
      }
      throw error;
    }
  }

  /**
   * Delete a value from the state.
   */
  async deleteValue(path: string): Promise<void> {
    await this.client.delete(`/v1/state/${path}/value`);
  }

  /**
   * Get all key-value pairs in a state container.
   */
  async getAllValues(statePath: string): Promise<Record<string, unknown>> {
    const response = await this.client.get<Record<string, unknown>>(`/v1/state/${statePath}`);
    return response.data;
  }

  // History & Events

  /**
   * Get the event history for a state path.
   */
  async getHistory(path: string, options: ListOptions & { eventTypes?: string[] } = {}): Promise<StateEvent[]> {
    const params = {
      ...options,
      event_types: options.eventTypes?.join(','),
    };
    const response = await this.client.get<StateEvent[]>(`/v1/state/${path}/history`, { params });
    return response.data;
  }

  /**
   * Query state at a specific point in time or version.
   */
  async timeTravel(path: string, query: TimeTravelQuery): Promise<StateValue | null> {
    try {
      const response = await this.client.get<StateValue>(`/v1/state/${path}/time-travel`, {
        params: query,
      });
      return response.data;
    } catch (error) {
      if (error instanceof StateNotFoundError) {
        return null;
      }
      throw error;
    }
  }

  // Snapshots

  /**
   * Create a snapshot of the current state.
   */
  async createSnapshot(path: string, request: CreateSnapshotRequest = {}): Promise<StateSnapshot> {
    const response = await this.client.post<StateSnapshot>(`/v1/state/${path}/snapshot`, request);
    return response.data;
  }

  /**
   * List all snapshots for a state container.
   */
  async listSnapshots(path: string, options: ListOptions = {}): Promise<StateSnapshot[]> {
    const response = await this.client.get<StateSnapshot[]>(`/v1/state/${path}/snapshots`, {
      params: options,
    });
    return response.data;
  }

  /**
   * Restore state from a snapshot.
   */
  async restoreSnapshot(path: string, request: RestoreSnapshotRequest): Promise<StateContainer> {
    const response = await this.client.post<StateContainer>(`/v1/state/${path}/restore`, request);
    return response.data;
  }

  // Permissions

  /**
   * Grant permission to access a state.
   */
  async grantPermission(path: string, request: GrantPermissionRequest): Promise<StatePermission> {
    const response = await this.client.post<StatePermission>(`/v1/state/${path}/permissions`, request);
    return response.data;
  }

  /**
   * Get all permissions for a state.
   */
  async getPermissions(path: string): Promise<StatePermission[]> {
    const response = await this.client.get<StatePermission[]>(`/v1/state/${path}/permissions`);
    return response.data;
  }

  // Triggers

  /**
   * Create a trigger for state changes.
   */
  async createTrigger(request: CreateTriggerRequest): Promise<StateTrigger> {
    const response = await this.client.post<StateTrigger>('/v1/triggers', request);
    return response.data;
  }

  /**
   * Get triggers, optionally filtered by state path.
   */
  async getTriggers(options: { statePath?: string; isActive?: boolean } = {}): Promise<StateTrigger[]> {
    const response = await this.client.get<StateTrigger[]>('/v1/triggers', { params: options });
    return response.data;
  }

  /**
   * Delete a trigger.
   */
  async deleteTrigger(triggerId: string): Promise<void> {
    await this.client.delete(`/v1/triggers/${triggerId}`);
  }
}

/**
 * State manager for declarative state access.
 * 
 * Provides a simpler interface for managing state with
 * convenience methods.
 */
export class StateManager {
  private client: StateClient;
  private statePrefix: string;

  constructor(options: StateClientOptions = {}) {
    this.client = new StateClient(options);
    this.statePrefix = options.tenantId || 'default';
  }

  /**
   * Get a value from state.
   */
  async get(stateName: string, key: string, defaultValue: unknown = null): Promise<unknown> {
    const path = `${this.statePrefix}/${stateName}/${key}`;
    const value = await this.client.getValue(path);
    return value?.value ?? defaultValue;
  }

  /**
   * Set a value in state.
   */
  async set(stateName: string, key: string, value: unknown): Promise<StateValue> {
    const path = `${this.statePrefix}/${stateName}/${key}`;
    return this.client.setValue(path, value as Record<string, unknown>);
  }

  /**
   * Delete a value from state.
   */
  async delete(stateName: string, key: string): Promise<void> {
    const path = `${this.statePrefix}/${stateName}/${key}`;
    await this.client.deleteValue(path);
  }

  /**
   * Create a snapshot of a state container.
   */
  async snapshot(stateName: string, label?: string): Promise<StateSnapshot> {
    const path = `${this.statePrefix}/${stateName}`;
    return this.client.createSnapshot(path, { label });
  }

  /**
   * Restore a state container from snapshot.
   */
  async restore(stateName: string, snapshotVersion: number): Promise<StateContainer> {
    const path = `${this.statePrefix}/${stateName}`;
    return this.client.restoreSnapshot(path, { snapshotVersion });
  }

  /**
   * Get the underlying client for advanced operations.
   */
  getClient(): StateClient {
    return this.client;
  }
}

// Convenience functions using default client
let defaultClient: StateClient | null = null;

function getDefaultClient(options: StateClientOptions = {}): StateClient {
  if (!defaultClient) {
    defaultClient = new StateClient(options);
  }
  return defaultClient;
}

export async function getValue(path: string, options?: StateClientOptions): Promise<unknown> {
  const client = getDefaultClient(options);
  const value = await client.getValue(path);
  return value?.value ?? null;
}

export async function setValue(path: string, value: Record<string, unknown>, options?: StateClientOptions): Promise<StateValue> {
  const client = getDefaultClient(options);
  return client.setValue(path, value);
}

export async function deleteValue(path: string, options?: StateClientOptions): Promise<void> {
  const client = getDefaultClient(options);
  return client.deleteValue(path);
}

export async function getHistory(path: string, options?: StateClientOptions & ListOptions): Promise<StateEvent[]> {
  const client = getDefaultClient(options);
  return client.getHistory(path, options);
}

export async function createSnapshot(path: string, request?: CreateSnapshotRequest, options?: StateClientOptions): Promise<StateSnapshot> {
  const client = getDefaultClient(options);
  return client.createSnapshot(path, request);
}

export async function restoreSnapshot(path: string, request: RestoreSnapshotRequest, options?: StateClientOptions): Promise<StateContainer> {
  const client = getDefaultClient(options);
  return client.restoreSnapshot(path, request);
}
