export type Runtime =
  | 'typescript'
  | 'javascript'
  | 'python'
  | 'python-wasm'
  | 'rust-wasm'
  | 'go'
  | 'deno'
  | 'bun'
  | 'browser-wasm';
export type HttpMethod = 'ANY' | 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
export type Visibility = 'public' | 'private';
export type BackoffStrategy = 'linear' | 'exponential' | 'fixed';

export interface EnvironmentVariable {
  id: string;
  key: string;
  value: string;
  isSecret: boolean;
}

export interface DeploymentLog {
  id: string;
  timestamp: string;
  level: 'info' | 'warn' | 'error' | 'success';
  message: string;
}

export interface HttpTrigger {
  enabled: boolean;
  method: HttpMethod;
  path: string;
}

export interface ScheduleTrigger {
  enabled: boolean;
  cron: string;
  timezone: string;
}

export interface ResourceLimits {
  memoryMb: number;
  timeoutMs: number;
  maxConcurrency: number;
}

export interface RetryPolicy {
  maxRetries: number;
  backoffMs: number;
  backoffStrategy: BackoffStrategy;
}

export interface AdvancedSettings {
  retryPolicy: RetryPolicy;
  warmInstances: number;
}

export interface ProviderOption {
  id: string;
  name: string;
  regions: string[];
}

export interface DraftState {
  functionName: string;
  slug: string;
  description: string;
  runtime: Runtime;
  runtimeVersion: string;
  code: string;
  visibility: Visibility;
  tags: string[];
  envVars: EnvironmentVariable[];
  resources: ResourceLimits;
  httpTrigger: HttpTrigger;
  scheduleTrigger: ScheduleTrigger;
  retryPolicy: RetryPolicy;
  warmInstances: number;
  savedAt: string;
}
