import { z } from 'zod';

// ============================================================================
// Zod Schemas
// ============================================================================

export const githubConnectionSchema = z.object({
  id: z.string().uuid(),
  user_id: z.string().uuid(),
  tenant_id: z.string().uuid(),
  github_user_id: z.number(),
  github_username: z.string(),
  github_avatar_url: z.string().nullable(),
  github_profile_url: z.string().nullable(),
  token_scope: z.string().nullable(),
  token_expires_at: z.string().nullable(),
  github_app_install: z.boolean(),
  status: z.enum(['active', 'expired', 'revoked', 'error']),
  last_synced_at: z.string().nullable(),
  created_at: z.string(),
  updated_at: z.string(),
});

export const detectedFunctionSchema = z.object({
  name: z.string(),
  entry_point: z.string(),
  runtime: z.string(),
  sub_directory: z.string().nullable(),
  confidence: z.number(),
  strategy: z.string(),
  manifest: z.record(z.string(), z.unknown()).nullable(),
  dependencies: z.array(z.string()).nullable(),
});

export const githubRepoSchema = z.object({
  id: z.string().uuid(),
  github_repo_id: z.number(),
  full_name: z.string(),
  name: z.string(),
  owner: z.string(),
  description: z.string().nullable(),
  default_branch: z.string(),
  language: z.string().nullable(),
  languages: z.record(z.string(), z.number()).nullable(),
  is_private: z.boolean(),
  is_fork: z.boolean(),
  is_archived: z.boolean(),
  topics: z.array(z.string()).nullable(),
  stars_count: z.number(),
  forks_count: z.number(),
  size_kb: z.number(),
  pushed_at: z.string().nullable(),
  html_url: z.string(),
  detected_functions: z.array(detectedFunctionSchema).nullable(),
  detected_runtime: z.string().nullable(),
  has_functionfly_json: z.boolean(),
  import_status: z.enum(['not_imported', 'importing', 'imported', 'partial', 'error']).nullable(),
  last_scanned_at: z.string().nullable(),
  created_at: z.string(),
});

export const scanResultSchema = z.object({
  repo_id: z.string().uuid(),
  functions: z.array(detectedFunctionSchema),
  primary_runtime: z.string(),
  overall_confidence: z.number(),
  strategy_used: z.string(),
  warnings: z.array(z.string()),
  estimated_import_time_seconds: z.number(),
  estimated_cost_usd: z.number(),
});

export const githubImportSchema = z.object({
  id: z.string().uuid(),
  user_id: z.string().uuid(),
  tenant_id: z.string().uuid(),
  connection_id: z.string().uuid(),
  repo_id: z.string().uuid(),
  source_branch: z.string(),
  source_path: z.string(),
  function_name: z.string(),
  function_id: z.string().uuid().nullable(),
  function_version_id: z.string().uuid().nullable(),
  visibility: z.enum(['public', 'private', 'unlisted']),
  runtime_override: z.string().nullable(),
  manifest_overrides: z.record(z.string(), z.unknown()).nullable(),
  auto_sync_enabled: z.boolean(),
  sync_branches: z.array(z.string()).nullable(),
  environment_mappings: z.record(z.string(), z.string()).nullable(),
  status: z.enum(['pending', 'scanning', 'configuring', 'fetching', 'building', 'publishing', 'completed', 'failed', 'cancelled']),
  progress: z.number(),
  error_message: z.string().nullable(),
  error_details: z.record(z.string(), z.unknown()).nullable(),
  content_hash: z.string().nullable(),
  commit_sha: z.string().nullable(),
  files_imported: z.number(),
  total_size_bytes: z.number(),
  created_at: z.string(),
  updated_at: z.string(),
  completed_at: z.string().nullable(),
});

export const githubSyncLogSchema = z.object({
  id: z.string().uuid(),
  import_id: z.string().uuid(),
  function_id: z.string().uuid(),
  trigger_type: z.enum(['push', 'pr_open', 'pr_sync', 'pr_close', 'manual']),
  trigger_branch: z.string().nullable(),
  trigger_commit_sha: z.string().nullable(),
  trigger_pr_number: z.number().nullable(),
  status: z.enum(['pending', 'building', 'deploying', 'completed', 'failed', 'skipped']),
  version_published: z.string().nullable(),
  duration_ms: z.number(),
  error_message: z.string().nullable(),
  created_at: z.string(),
  completed_at: z.string().nullable(),
});

export const githubTemplateSchema = z.object({
  id: z.string().uuid(),
  tenant_id: z.string().uuid(),
  user_id: z.string().uuid(),
  name: z.string(),
  description: z.string().nullable(),
  config: z.record(z.string(), z.unknown()),
  detection_rules: z.array(z.record(z.string(), z.unknown())).nullable(),
  is_default: z.boolean(),
  usage_count: z.number(),
  created_at: z.string(),
  updated_at: z.string(),
});

export const importConflictSchema = z.object({
  function_name: z.string(),
  existing_function_id: z.string().uuid(),
  existing_version: z.string(),
  resolution: z.enum(['overwrite', 'skip', 'rename']).optional(),
});

export const importPreviewSchema = z.object({
  functions: z.array(detectedFunctionSchema),
  total_estimated_cost_usd: z.number(),
  warnings: z.array(z.string()),
  conflicts: z.array(importConflictSchema),
});

export const githubConnectionListSchema = z.array(githubConnectionSchema);
export const githubRepoListSchema = z.array(githubRepoSchema);
export const githubImportListSchema = z.array(githubImportSchema);
export const githubSyncLogListSchema = z.array(githubSyncLogSchema);
export const githubTemplateListSchema = z.array(githubTemplateSchema);
export const detectedFunctionListSchema = z.array(detectedFunctionSchema);
export const branchSchema = z.object({
  name: z.string(),
  sha: z.string(),
  protected: z.boolean(),
});
export const branchListSchema = z.array(branchSchema);

export const treeEntrySchema = z.object({
  path: z.string(),
  type: z.enum(['blob', 'tree']),
  sha: z.string(),
  size: z.number().nullable(),
  url: z.string(),
});

export const treeResponseSchema = z.object({
  sha: z.string(),
  truncated: z.boolean(),
  tree: z.array(treeEntrySchema),
});

// ============================================================================
// Inferred TypeScript Types
// ============================================================================

export type GitHubConnection = z.infer<typeof githubConnectionSchema>;
export type DetectedFunction = z.infer<typeof detectedFunctionSchema>;
export type GitHubRepo = z.infer<typeof githubRepoSchema>;
export type ScanResult = z.infer<typeof scanResultSchema>;
export type GitHubImport = z.infer<typeof githubImportSchema>;
export type GitHubSyncLog = z.infer<typeof githubSyncLogSchema>;
export type GitHubTemplate = z.infer<typeof githubTemplateSchema>;
export type ImportConflict = z.infer<typeof importConflictSchema>;
export type ImportPreview = z.infer<typeof importPreviewSchema>;
export type TreeEntry = z.infer<typeof treeEntrySchema>;
export type TreeResponse = z.infer<typeof treeResponseSchema>;
export type Branch = z.infer<typeof branchSchema>;

// ============================================================================
// SSE Event Interfaces
// ============================================================================

export interface ImportProgressEvent {
  stage: 'scanning' | 'fetching' | 'building' | 'publishing';
  progress: number;
  message: string;
}

export interface ImportCompleteEvent {
  stage: 'completed';
  progress: number;
  function_id?: string | null;
  function_name?: string;
  commit_sha?: string | null;
  files_imported?: number;
}

export interface ImportErrorEvent {
  stage: 'error';
  progress: number;
  message: string;
  details?: {
    code?: string;
    retryable?: boolean;
    failed_stage?: string;
  };
}

// ============================================================================
// Request/Response Types
// ============================================================================

export interface ScanRepoOptions {
  force_rescan?: boolean;
  strategy?: string;
  include_dependencies?: boolean;
}

export interface StartImportRequest {
  connection_id: string;
  repo_id: string;
  source_branch: string;
  source_path: string;
  function_name: string;
  visibility?: 'public' | 'private' | 'unlisted';
  runtime_override?: string;
  manifest_overrides?: Record<string, unknown>;
  auto_sync_enabled?: boolean;
  sync_branches?: string[];
  environment_mappings?: Record<string, string>;
}

export interface BulkImportRequest {
  imports: StartImportRequest[];
}

export interface UpdateSyncRequest {
  auto_sync_enabled?: boolean;
  sync_branches?: string[];
  environment_mappings?: Record<string, string>;
}

export interface ListReposParams {
  page?: number;
  per_page?: number;
  search?: string;
  sort?: 'full_name' | 'updated_at' | 'pushed_at' | 'stars';
  direction?: 'asc' | 'desc';
  type?: 'all' | 'owner' | 'member';
}

export interface ListImportsParams {
  page?: number;
  per_page?: number;
  status?: GitHubImport['status'];
  repo_id?: string;
}

export interface GetTreeParams {
  ref?: string;
  recursive?: boolean;
}

export interface ListSyncLogsParams {
  page?: number;
  per_page?: number;
  trigger_type?: GitHubSyncLog['trigger_type'];
  status?: GitHubSyncLog['status'];
}

export interface CreateTemplateRequest {
  name: string;
  description?: string;
  config: Record<string, unknown>;
  detection_rules?: Record<string, unknown>[];
  is_default?: boolean;
}

export interface UpdateTemplateRequest {
  name?: string;
  description?: string;
  config?: Record<string, unknown>;
  detection_rules?: Record<string, unknown>[];
  is_default?: boolean;
}

export interface ListResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
}

export interface PaginatedImportsResponse {
  imports: GitHubImport[];
  total: number;
  page: number;
  per_page: number;
}

export interface PaginatedSyncLogsResponse {
  logs: GitHubSyncLog[];
  total: number;
  page: number;
  per_page: number;
}

export interface PaginatedReposResponse {
  repos: GitHubRepo[];
  total: number;
  page: number;
  per_page: number;
}
