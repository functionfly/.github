# GitHub Repo Import — Complete UI & Component Specification

**Date:** 2026-05-01
**Status:** Implementation-Ready Specification

---

## Table of Contents

1. [Architecture Suggestions & Improvements](#1-architecture-suggestions--improvements)
2. [File Structure](#2-file-structure)
3. [API Client Layer](#3-api-client-layer)
4. [Zod Schemas & Types](#4-zod-schemas--types)
5. [React Query Hooks](#5-react-query-hooks)
6. [Zustand Store](#6-zustand-store)
7. [UI Components (shadcn/ui + Custom)](#7-ui-components)
8. [Page Components](#8-page-components)
9. [Routing & Navigation](#9-routing--navigation)
10. [Sidebar Integration](#10-sidebar-integration)
11. [i18n Keys](#11-i18n-keys)

---

## 1. Architecture Suggestions & Improvements

### 1.1 Critical Additions to the Plan

| # | Suggestion | Why |
|---|-----------|-----|
| **1** | **Add SSE (Server-Sent Events) for import progress** | The existing codebase uses polling for deploy status. Import is multi-stage (30s-5min) — SSE gives real-time progress without polling overhead. Use `EventSource` on the frontend, `http.Flusher` on the backend. |
| **2** | **Add GitHub App support from day one** | OAuth `repo` scope gives access to ALL repos. Enterprise users won't accept this. Design the schema with `github_app_install_id` from the start. |
| **3** | **Add `dry-run` mode for imports** | Before committing to a full import, let users preview exactly what will be created (function name, runtime, manifest, estimated cost). Reduces wasted platform fees. |
| **4** | **Add conflict detection** | If a user imports a repo that creates a function with the same `author/name` as an existing one, detect this and offer: overwrite, rename, skip, or create new version. |
| **5** | **Add `functionfly.jsonc` generation on import** | If the repo doesn't have a manifest, auto-generate one and offer to create a PR back to the repo with the generated `functionfly.jsonc`. This drives adoption. |
| **6** | **Add rollback-on-failure** | If a multi-function import fails halfway, roll back all partially-created functions. Use the existing saga pattern. |
| **7** | **Add import quota enforcement** | Gate imports per plan: Free (5 imports/mo), Pro (50), Enterprise (unlimited). Reuse existing `RequireFeature` middleware. |
| **8** | **Add dependency caching layer** | Cache `node_modules`, `pip` packages, etc. in Redis/R2 keyed by lockfile hash. Avoids re-downloading on every sync. |
| **9** | **Add fork detection** | If a repo is a fork, detect the upstream and offer to compare changes. Useful for contributing to open-source functions. |
| **10** | **Add workspace-level connections** | Currently the plan scopes connections to `user_id`. Add `team_id` as an alternative scope so entire teams can share a GitHub connection. |

### 1.2 Enhanced Data Flow

```
User clicks "Connect GitHub"
       │
       ▼
OAuth with scopes: repo, read:user, user:email
       │
       ▼
Callback → encrypt token → store in github_connections
       │
       ▼
Async: Fetch repos → cache in github_repos (background)
       │
       ▼
User browses repos (instant from cache)
       │
       ▼
User clicks repo → POST /scan → AI detection pipeline
       │
       ▼
Shows detected functions with confidence scores
       │
       ▼
User selects functions + sets visibility + configures sync
       │
       ▼
POST /import (dry-run first) → shows preview
       │
       ▼
User confirms → full import pipeline (SSE progress)
       │
       ▼
Creates registry_functions + versions → registers webhook
       │
       ▼
Reports GitHub commit status → shows in dashboard
```

---

## 2. File Structure

```
web/dashboard/src/
├── api/
│   └── github.ts                              # API client (all endpoints)
├── hooks/
│   ├── useGitHubConnection.ts                 # Connection query + mutations
│   ├── useGitHubRepos.ts                      # Repo listing + scanning
│   ├── useGitHubImport.ts                     # Import operations
│   ├── useGitHubSync.ts                       # Sync management
│   └── useGitHubTemplates.ts                  # Import templates
├── stores/
│   └── githubStore.ts                         # Zustand store (connection state)
├── types/
│   └── github.ts                              # Zod schemas + TypeScript types
├── components/
│   └── github/
│       ├── GitHubConnectButton.tsx            # "Connect GitHub" CTA
│       ├── GitHubConnectionCard.tsx           # Connected account display
│       ├── GitHubRepoCard.tsx                 # Individual repo card
│       ├── GitHubRepoGrid.tsx                 # Grid of repo cards
│       ├── GitHubRepoListItem.tsx             # List view repo row
│       ├── DetectedFunctionCard.tsx           # Detected function in scan results
│       ├── FunctionDetectionResults.tsx       # Scan results panel
│       ├── ImportConfigForm.tsx               # Per-function import config
│       ├── ImportPreviewCard.tsx              # Dry-run preview
│       ├── ImportProgressStepper.tsx          # Multi-stage progress (SSE)
│       ├── ImportStatusBadge.tsx              # Import status indicator
│       ├── SyncStatusIndicator.tsx            # Auto-sync on/off with status
│       ├── SyncLogEntry.tsx                   # Individual sync log row
│       ├── WebhookStatusBadge.tsx             # Webhook health indicator
│       ├── EnvironmentMappingEditor.tsx       # Branch → environment mapper
│       ├── VisibilitySelector.tsx             # Public/Private/Unlisted picker
│       ├── GitHubStatsBar.tsx                 # Summary stats (repos, imports, syncs)
│       ├── ImportTemplateCard.tsx             # Template card
│       ├── ImportTemplateForm.tsx             # Template create/edit form
│       ├── ConflictResolutionDialog.tsx       # Handle name conflicts
│       ├── DryRunPreviewDialog.tsx            # Preview before import
│       └── EmptyStates/
│           ├── NoGitHubConnection.tsx         # "Connect GitHub to get started"
│           ├── NoReposFound.tsx               # "No repos found"
│           └── NoImportsYet.tsx               # "Import your first function"
├── pages/
│   ├── GitHubPage/
│   │   ├── index.tsx                          # Main GitHub hub page
│   │   ├── GitHubReposTab.tsx                 # Repos sub-tab
│   │   ├── GitHubImportsTab.tsx               # Imports sub-tab
│   │   └── GitHubTemplatesTab.tsx             # Templates sub-tab
│   ├── GitHubRepoImportPage/
│   │   ├── index.tsx                          # Repo scan + import config
│   │   └── ImportExecutionStep.tsx            # Import progress with SSE
│   └── GitHubSettingsPage/
│       └── index.tsx                          # Connection settings (in Settings)
└── styles/
    └── github.css                             # GitHub-specific styles (optional)
```

---

## 3. API Client Layer

### `web/dashboard/src/api/github.ts`

```typescript
import { apiClient } from './client';
import {
  githubConnectionSchema,
  githubRepoSchema,
  githubImportSchema,
  githubSyncLogSchema,
  githubTemplateSchema,
  type GitHubConnection,
  type GitHubRepo,
  type GitHubImport,
  type GitHubSyncLog,
  type GitHubTemplate,
  type ScanResult,
  type ImportPreview,
  scanResultSchema,
  importPreviewSchema,
} from '@/types/github';

// ──────────────────────────────────────────────
// Connection Management
// ──────────────────────────────────────────────

export const githubApi = {
  // OAuth flow — returns redirect URL
  getConnectUrl: (): Promise<{ url: string }> =>
    apiClient.get('/v1/github/connect'),

  // Current connection status
  getConnection: (): Promise<GitHubConnection | null> =>
    apiClient.getValidated<GitHubConnection>(
      githubConnectionSchema,
      '/v1/github/connection',
      undefined,
      null // fallback if not connected
    ),

  // Disconnect GitHub account
  disconnect: (): Promise<void> =>
    apiClient.delete('/v1/github/connection'),

  // Force token refresh
  refreshToken: (): Promise<{ expires_at: string }> =>
    apiClient.post('/v1/github/connection/refresh'),

  // ──────────────────────────────────────────────
  // Repository Browsing
  // ──────────────────────────────────────────────

  // List repos (cached, paginated)
  listRepos: (params?: {
    page?: number;
    per_page?: number;
    sort?: 'updated' | 'pushed' | 'full_name' | 'stars';
    direction?: 'asc' | 'desc';
    language?: string;
    visibility?: 'all' | 'public' | 'private';
    search?: string;
  }): Promise<{ repos: GitHubRepo[]; total: number; page: number; per_page: number }> =>
    apiClient.get('/v1/github/repos', { params }),

  // Force re-fetch from GitHub
  refreshRepos: (): Promise<{ count: number }> =>
    apiClient.post('/v1/github/repos/refresh'),

  // Get single repo details
  getRepo: (repoId: string): Promise<GitHubRepo> =>
    apiClient.getValidatedData<GitHubRepo>(githubRepoSchema, `/v1/github/repos/${repoId}`),

  // Deep scan for functions
  scanRepo: (repoId: string, options?: {
    branch?: string;
    use_ai?: boolean;
  }): Promise<ScanResult> =>
    apiClient.postValidatedData<ScanResult>(
      scanResultSchema,
      `/v1/github/repos/${repoId}/scan`,
      options
    ),

  // List branches
  listBranches: (repoId: string): Promise<{ branches: Array<{ name: string; sha: string; is_default: boolean }> }> =>
    apiClient.get(`/v1/github/repos/${repoId}/branches`),

  // Browse file tree
  getTree: (repoId: string, params?: {
    branch?: string;
    path?: string;
    recursive?: boolean;
  }): Promise<{ tree: Array<{ path: string; type: string; sha: string; size?: number }> }> =>
    apiClient.get(`/v1/github/repos/${repoId}/tree`, { params }),

  // ──────────────────────────────────────────────
  // Import Operations
  // ──────────────────────────────────────────────

  // Start import
  startImport: (data: {
    repo_id: string;
    branch?: string;
    functions: Array<{
      detected_name: string;
      custom_name?: string;
      entry_point: string;
      sub_directory?: string;
      runtime: string;
      visibility: 'public' | 'private' | 'unlisted';
      manifest_override?: Record<string, unknown>;
    }>;
    visibility?: 'public' | 'private' | 'unlisted';
    auto_sync?: boolean;
    sync_branches?: string[];
    environment_mappings?: Record<string, string>;
    template_id?: string;
  }): Promise<{ import_id: string }> =>
    apiClient.post('/v1/github/import', data),

  // Dry-run preview (no actual import)
  previewImport: (data: Parameters<typeof githubApi.startImport>[0]): Promise<ImportPreview> =>
    apiClient.postValidatedData<ImportPreview>('/v1/github/import/preview', importPreviewSchema, data),

  // Bulk import multiple repos
  bulkImport: (data: {
    imports: Array<Parameters<typeof githubApi.startImport>[0]>;
  }): Promise<{ import_ids: string[] }> =>
    apiClient.post('/v1/github/import/bulk', data),

  // List all imports
  listImports: (params?: {
    page?: number;
    per_page?: number;
    status?: string;
    repo_id?: string;
  }): Promise<{ imports: GitHubImport[]; total: number }> =>
    apiClient.get('/v1/github/imports', { params }),

  // Get single import
  getImport: (importId: string): Promise<GitHubImport> =>
    apiClient.getValidatedData<GitHubImport>(githubImportSchema, `/v1/github/imports/${importId}`),

  // Cancel import
  cancelImport: (importId: string): Promise<void> =>
    apiClient.post(`/v1/github/imports/${importId}/cancel`),

  // Retry failed import
  retryImport: (importId: string): Promise<{ import_id: string }> =>
    apiClient.post(`/v1/github/imports/${importId}/retry`),

  // Force re-sync from latest commit
  resyncImport: (importId: string): Promise<{ import_id: string }> =>
    apiClient.post(`/v1/github/imports/${importId}/resync`),

  // ──────────────────────────────────────────────
  // Sync Management
  // ──────────────────────────────────────────────

  // Update sync settings
  updateSync: (importId: string, data: {
    auto_sync_enabled?: boolean;
    sync_branches?: string[];
    environment_mappings?: Record<string, string>;
  }): Promise<void> =>
    apiClient.put(`/v1/github/imports/${importId}/sync`, data),

  // Get sync logs
  getSyncLogs: (importId: string, params?: {
    page?: number;
    per_page?: number;
    status?: string;
  }): Promise<{ logs: GitHubSyncLog[]; total: number }> =>
    apiClient.get(`/v1/github/imports/${importId}/sync-logs`, { params }),

  // ──────────────────────────────────────────────
  // Import Templates
  // ──────────────────────────────────────────────

  listTemplates: (): Promise<{ templates: GitHubTemplate[] }> =>
    apiClient.get('/v1/github/templates'),

  createTemplate: (data: {
    name: string;
    description?: string;
    config: Record<string, unknown>;
    detection_rules?: Record<string, unknown>;
    is_default?: boolean;
  }): Promise<GitHubTemplate> =>
    apiClient.postValidatedData<GitHubTemplate>(
      githubTemplateSchema,
      '/v1/github/templates',
      data
    ),

  updateTemplate: (id: string, data: Partial<Parameters<typeof githubApi.createTemplate>[0]>): Promise<GitHubTemplate> =>
    apiClient.putValidatedData<GitHubTemplate>(
      githubTemplateSchema,
      `/v1/github/templates/${id}`,
      data
    ),

  deleteTemplate: (id: string): Promise<void> =>
    apiClient.delete(`/v1/github/templates/${id}`),

  // ──────────────────────────────────────────────
  // SSE Import Progress Stream
  // ──────────────────────────────────────────────

  /**
   * Opens an SSE connection for real-time import progress.
   * Returns a cleanup function to close the EventSource.
   *
   * Progress events:
   *   { stage: 'scanning' | 'fetching' | 'building' | 'publishing', progress: 0-100, message: string }
   * Completion event:
   *   { stage: 'completed', function_id: string, version: string }
   * Error event:
   *   { stage: 'error', message: string, details?: object }
   */
  streamImportProgress: (
    importId: string,
    onProgress: (event: ImportProgressEvent) => void,
    onComplete: (event: ImportCompleteEvent) => void,
    onError: (event: ImportErrorEvent) => void
  ): (() => void) => {
    const baseUrl = apiClient.defaults.baseURL || window.location.origin;
    const token = localStorage.getItem('ff-access-token');
    const es = new EventSource(
      `${baseUrl}/v1/github/imports/${importId}/progress?token=${token}`
    );

    es.addEventListener('progress', (e) => onProgress(JSON.parse(e.data)));
    es.addEventListener('complete', (e) => { onComplete(JSON.parse(e.data)); es.close(); });
    es.addEventListener('error', (e) => {
      if (e.data) {
        onError(JSON.parse(e.data));
      }
      es.close();
    });

    return () => es.close();
  },
};

// ──────────────────────────────────────────────
// Event Types
// ──────────────────────────────────────────────

export interface ImportProgressEvent {
  stage: 'scanning' | 'fetching' | 'building' | 'publishing';
  progress: number; // 0-100
  message: string;
  details?: {
    files_processed?: number;
    total_files?: number;
    current_file?: string;
    bytes_downloaded?: number;
    build_output?: string;
  };
}

export interface ImportCompleteEvent {
  stage: 'completed';
  function_id: string;
  function_name: string;
  version: string;
  visibility: string;
  url: string;
}

export interface ImportErrorEvent {
  stage: 'error';
  message: string;
  details?: {
    code?: string;
    retryable?: boolean;
    failed_stage?: string;
  };
}
```

---

## 4. Zod Schemas & Types

### `web/dashboard/src/types/github.ts`

```typescript
import { z } from 'zod';

// ──────────────────────────────────────────────
// Connection
// ──────────────────────────────────────────────

export const githubConnectionSchema = z.object({
  id: z.string().uuid(),
  user_id: z.string().uuid(),
  tenant_id: z.string().uuid(),
  github_user_id: z.number(),
  github_username: z.string(),
  github_avatar_url: z.string().nullable(),
  github_profile_url: z.string().nullable(),
  token_scope: z.string().nullable(),
  token_expires_at: z.string().datetime().nullable(),
  github_app_install: z.boolean(),
  status: z.enum(['active', 'expired', 'revoked', 'error']),
  last_synced_at: z.string().datetime().nullable(),
  created_at: z.string().datetime(),
  updated_at: z.string().datetime(),
});
export type GitHubConnection = z.infer<typeof githubConnectionSchema>;

// ──────────────────────────────────────────────
// Repository
// ──────────────────────────────────────────────

export const detectedFunctionSchema = z.object({
  name: z.string(),
  entry_point: z.string(),
  runtime: z.string(),
  sub_directory: z.string().optional(),
  confidence: z.number().min(0).max(1),
  strategy: z.string(),
  manifest: z.record(z.unknown()).optional(),
  dependencies: z.object({
    manager: z.string(), // npm, pip, go, cargo, etc.
    lockfile: z.string().optional(),
    packages: z.array(z.string()).optional(),
  }).optional(),
});
export type DetectedFunction = z.infer<typeof detectedFunctionSchema>;

export const githubRepoSchema = z.object({
  id: z.string().uuid(),
  github_repo_id: z.number(),
  full_name: z.string(),
  name: z.string(),
  owner: z.string(),
  description: z.string().nullable(),
  default_branch: z.string(),
  language: z.string().nullable(),
  languages: z.record(z.number()).nullable(),
  is_private: z.boolean(),
  is_fork: z.boolean(),
  is_archived: z.boolean(),
  topics: z.array(z.string()),
  stars_count: z.number(),
  forks_count: z.number(),
  size_kb: z.number(),
  pushed_at: z.string().datetime().nullable(),
  html_url: z.string(),
  detected_functions: z.array(detectedFunctionSchema),
  detected_runtime: z.string().nullable(),
  has_functionfly_json: z.boolean(),
  import_status: z.enum(['not_imported', 'importing', 'imported', 'partial', 'error']),
  last_scanned_at: z.string().datetime().nullable(),
  created_at: z.string().datetime(),
});
export type GitHubRepo = z.infer<typeof githubRepoSchema>;

// ──────────────────────────────────────────────
// Scan Result
// ──────────────────────────────────────────────

export const scanResultSchema = z.object({
  repo_id: z.string().uuid(),
  functions: z.array(detectedFunctionSchema),
  primary_runtime: z.string(),
  overall_confidence: z.number().min(0).max(1),
  strategy_used: z.string(),
  warnings: z.array(z.string()),
  estimated_import_time_seconds: z.number(),
  estimated_cost_usd: z.number(),
});
export type ScanResult = z.infer<typeof scanResultSchema>;

// ──────────────────────────────────────────────
// Import
// ──────────────────────────────────────────────

export const githubImportSchema = z.object({
  id: z.string().uuid(),
  user_id: z.string().uuid(),
  tenant_id: z.string().uuid(),
  connection_id: z.string().uuid(),
  repo_id: z.string().uuid(),
  source_branch: z.string(),
  source_path: z.string().nullable(),
  function_name: z.string(),
  function_id: z.string().uuid().nullable(),
  function_version_id: z.string().uuid().nullable(),
  visibility: z.enum(['public', 'private', 'unlisted']),
  runtime_override: z.string().nullable(),
  manifest_overrides: z.record(z.unknown()),
  auto_sync_enabled: z.boolean(),
  sync_branches: z.array(z.string()),
  environment_mappings: z.record(z.string()),
  status: z.enum([
    'pending', 'scanning', 'configuring', 'fetching',
    'building', 'publishing', 'completed', 'failed', 'cancelled'
  ]),
  progress: z.number().min(0).max(100),
  error_message: z.string().nullable(),
  error_details: z.record(z.unknown()).nullable(),
  content_hash: z.string().nullable(),
  commit_sha: z.string().nullable(),
  files_imported: z.number(),
  total_size_bytes: z.number(),
  created_at: z.string().datetime(),
  updated_at: z.string().datetime(),
  completed_at: z.string().datetime().nullable(),
});
export type GitHubImport = z.infer<typeof githubImportSchema>;

// ──────────────────────────────────────────────
// Sync Log
// ──────────────────────────────────────────────

export const githubSyncLogSchema = z.object({
  id: z.string().uuid(),
  import_id: z.string().uuid(),
  function_id: z.string().uuid().nullable(),
  trigger_type: z.enum(['push', 'pr_open', 'pr_sync', 'pr_close', 'manual']),
  trigger_branch: z.string().nullable(),
  trigger_commit_sha: z.string().nullable(),
  trigger_pr_number: z.number().nullable(),
  status: z.enum(['pending', 'building', 'deploying', 'completed', 'failed', 'skipped']),
  version_published: z.string().nullable(),
  duration_ms: z.number().nullable(),
  error_message: z.string().nullable(),
  created_at: z.string().datetime(),
  completed_at: z.string().datetime().nullable(),
});
export type GitHubSyncLog = z.infer<typeof githubSyncLogSchema>;

// ──────────────────────────────────────────────
// Import Preview (dry-run)
// ──────────────────────────────────────────────

export const importPreviewSchema = z.object({
  functions: z.array(z.object({
    name: z.string(),
    runtime: z.string(),
    visibility: z.string(),
    estimated_size_bytes: z.number(),
    estimated_cost_usd: z.number(),
    has_conflict: z.boolean(),
    conflict_type: z.enum(['none', 'name_exists', 'version_exists']).optional(),
    existing_function_id: z.string().uuid().optional(),
  })),
  total_estimated_cost_usd: z.number(),
  warnings: z.array(z.string()),
  conflicts: z.array(z.object({
    function_name: z.string(),
    existing_id: z.string().uuid(),
    resolution_options: z.array(z.enum(['overwrite', 'rename', 'skip', 'new_version'])),
  })),
});
export type ImportPreview = z.infer<typeof importPreviewSchema>;

// ──────────────────────────────────────────────
// Template
// ──────────────────────────────────────────────

export const githubTemplateSchema = z.object({
  id: z.string().uuid(),
  tenant_id: z.string().uuid(),
  user_id: z.string().uuid(),
  name: z.string(),
  description: z.string().nullable(),
  config: z.record(z.unknown()),
  detection_rules: z.record(z.unknown()),
  is_default: z.boolean(),
  usage_count: z.number(),
  created_at: z.string().datetime(),
  updated_at: z.string().datetime(),
});
export type GitHubTemplate = z.infer<typeof githubTemplateSchema>;
```

---

## 5. React Query Hooks

### `web/dashboard/src/hooks/useGitHubConnection.ts`

```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { githubApi } from '@/api/github';
import { toast } from 'sonner';

export const githubKeys = {
  all: ['github'] as const,
  connection: () => [...githubKeys.all, 'connection'] as const,
  repos: (params?: Record<string, unknown>) => [...githubKeys.all, 'repos', params] as const,
  repo: (id: string) => [...githubKeys.all, 'repo', id] as const,
  branches: (id: string) => [...githubKeys.all, 'branches', id] as const,
  tree: (id: string, params?: Record<string, unknown>) => [...githubKeys.all, 'tree', id, params] as const,
  scan: (id: string) => [...githubKeys.all, 'scan', id] as const,
  imports: (params?: Record<string, unknown>) => [...githubKeys.all, 'imports', params] as const,
  import: (id: string) => [...githubKeys.all, 'import', id] as const,
  syncLogs: (importId: string, params?: Record<string, unknown>) =>
    [...githubKeys.all, 'syncLogs', importId, params] as const,
  templates: () => [...githubKeys.all, 'templates'] as const,
};

// Connection
export function useGitHubConnection() {
  return useQuery({
    queryKey: githubKeys.connection(),
    queryFn: () => githubApi.getConnection(),
    staleTime: 5 * 60 * 1000, // 5 min
    retry: false, // Don't retry on 404 (not connected)
  });
}

export function useGitHubConnect() {
  return useMutation({
    mutationFn: () => githubApi.getConnectUrl(),
    onSuccess: (data) => {
      window.location.href = data.url;
    },
    onError: (error: Error) => {
      toast.error(`Failed to connect GitHub: ${error.message}`);
    },
  });
}

export function useGitHubDisconnect() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => githubApi.disconnect(),
    onSuccess: () => {
      queryClient.removeQueries({ queryKey: githubKeys.all });
      toast.success('GitHub account disconnected');
    },
    onError: (error: Error) => {
      toast.error(`Failed to disconnect: ${error.message}`);
    },
  });
}

export function useGitHubTokenRefresh() {
  return useMutation({
    mutationFn: () => githubApi.refreshToken(),
    onSuccess: () => toast.success('Token refreshed'),
    onError: (error: Error) => toast.error(`Refresh failed: ${error.message}`),
  });
}
```

### `web/dashboard/src/hooks/useGitHubRepos.ts`

```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { githubApi } from '@/api/github';
import { githubKeys } from './useGitHubConnection';
import { toast } from 'sonner';

export function useGitHubRepos(params?: {
  page?: number;
  per_page?: number;
  sort?: string;
  language?: string;
  visibility?: string;
  search?: string;
}) {
  return useQuery({
    queryKey: githubKeys.repos(params),
    queryFn: () => githubApi.listRepos(params),
    staleTime: 60 * 1000, // 1 min
    enabled: !!params, // Only fetch when explicitly requested
  });
}

export function useGitHubRepo(repoId: string) {
  return useQuery({
    queryKey: githubKeys.repo(repoId),
    queryFn: () => githubApi.getRepo(repoId),
    staleTime: 2 * 60 * 1000,
    enabled: !!repoId,
  });
}

export function useRefreshGitHubRepos() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: () => githubApi.refreshRepos(),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: githubKeys.all });
      toast.success(`Refreshed ${data.count} repositories`);
    },
    onError: (error: Error) => {
      toast.error(`Failed to refresh: ${error.message}`);
    },
  });
}

export function useScanGitHubRepo(repoId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (options?: { branch?: string; use_ai?: boolean }) =>
      githubApi.scanRepo(repoId, options),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: githubKeys.repo(repoId) });
      toast.success(`Found ${data.functions.length} function(s)`);
    },
    onError: (error: Error) => {
      toast.error(`Scan failed: ${error.message}`);
    },
  });
}

export function useGitHubBranches(repoId: string) {
  return useQuery({
    queryKey: githubKeys.branches(repoId),
    queryFn: () => githubApi.listBranches(repoId),
    staleTime: 5 * 60 * 1000,
    enabled: !!repoId,
  });
}

export function useGitHubTree(repoId: string, params?: { branch?: string; path?: string }) {
  return useQuery({
    queryKey: githubKeys.tree(repoId, params),
    queryFn: () => githubApi.getTree(repoId, params),
    staleTime: 2 * 60 * 1000,
    enabled: !!repoId,
  });
}
```

### `web/dashboard/src/hooks/useGitHubImport.ts`

```typescript
import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from '@tanstack/react-query';
import { githubApi, type ImportProgressEvent, type ImportCompleteEvent, type ImportErrorEvent } from '@/api/github';
import { githubKeys } from './useGitHubConnection';
import { toast } from 'sonner';
import { useCallback, useEffect, useRef, useState } from 'react';

export function useGitHubImports(params?: {
  status?: string;
  repo_id?: string;
  page?: number;
  per_page?: number;
}) {
  return useQuery({
    queryKey: githubKeys.imports(params),
    queryFn: () => githubApi.listImports(params),
    staleTime: 30 * 1000, // 30s — imports change frequently
  });
}

export function useGitHubImport(importId: string) {
  return useQuery({
    queryKey: githubKeys.import(importId),
    queryFn: () => githubApi.getImport(importId),
    staleTime: 10 * 1000,
    enabled: !!importId,
    // Poll while in progress
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      if (status && ['pending', 'scanning', 'fetching', 'building', 'publishing'].includes(status)) {
        return 2000; // 2s
      }
      return false;
    },
  });
}

export function useStartImport() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Parameters<typeof githubApi.startImport>[0]) =>
      githubApi.startImport(data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: githubKeys.imports() });
      toast.success('Import started');
    },
    onError: (error: Error) => {
      toast.error(`Import failed: ${error.message}`);
    },
  });
}

export function usePreviewImport() {
  return useMutation({
    mutationFn: (data: Parameters<typeof githubApi.previewImport>[0]) =>
      githubApi.previewImport(data),
    onError: (error: Error) => {
      toast.error(`Preview failed: ${error.message}`);
    },
  });
}

export function useBulkImport() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: { imports: Parameters<typeof githubApi.startImport>[0][] }) =>
      githubApi.bulkImport(data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: githubKeys.imports() });
      toast.success(`Started ${data.import_ids.length} imports`);
    },
    onError: (error: Error) => {
      toast.error(`Bulk import failed: ${error.message}`);
    },
  });
}

export function useCancelImport() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (importId: string) => githubApi.cancelImport(importId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: githubKeys.imports() });
      toast.success('Import cancelled');
    },
    onError: (error: Error) => toast.error(`Cancel failed: ${error.message}`),
  });
}

export function useRetryImport() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (importId: string) => githubApi.retryImport(importId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: githubKeys.imports() });
      toast.success('Import retried');
    },
    onError: (error: Error) => toast.error(`Retry failed: ${error.message}`),
  });
}

export function useResyncImport() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (importId: string) => githubApi.resyncImport(importId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: githubKeys.imports() });
      toast.success('Re-sync started');
    },
    onError: (error: Error) => toast.error(`Resync failed: ${error.message}`),
  });
}

// ──────────────────────────────────────────────
// SSE Progress Hook
// ──────────────────────────────────────────────

export interface ImportProgressState {
  stage: ImportProgressEvent['stage'] | 'completed' | 'error' | 'idle';
  progress: number;
  message: string;
  details?: ImportProgressEvent['details'];
  result?: ImportCompleteEvent;
  error?: ImportErrorEvent;
}

export function useImportProgress(importId: string | null) {
  const [state, setState] = useState<ImportProgressState>({
    stage: 'idle',
    progress: 0,
    message: '',
  });
  const cleanupRef = useRef<(() => void) | null>(null);

  useEffect(() => {
    if (!importId) return;

    cleanupRef.current = githubApi.streamImportProgress(
      importId,
      (event) => setState({
        stage: event.stage,
        progress: event.progress,
        message: event.message,
        details: event.details,
      }),
      (event) => setState({
        stage: 'completed',
        progress: 100,
        message: 'Import completed!',
        result: event,
      }),
      (event) => setState({
        stage: 'error',
        progress: 0,
        message: event.message,
        error: event,
      })
    );

    return () => {
      cleanupRef.current?.();
    };
  }, [importId]);

  return state;
}
```

### `web/dashboard/src/hooks/useGitHubSync.ts`

```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { githubApi } from '@/api/github';
import { githubKeys } from './useGitHubConnection';
import { toast } from 'sonner';

export function useUpdateSync(importId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: {
      auto_sync_enabled?: boolean;
      sync_branches?: string[];
      environment_mappings?: Record<string, string>;
    }) => githubApi.updateSync(importId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: githubKeys.import(importId) });
      toast.success('Sync settings updated');
    },
    onError: (error: Error) => toast.error(`Update failed: ${error.message}`),
  });
}

export function useSyncLogs(importId: string, params?: { page?: number; status?: string }) {
  return useQuery({
    queryKey: githubKeys.syncLogs(importId, params),
    queryFn: () => githubApi.getSyncLogs(importId, params),
    staleTime: 15 * 1000,
    enabled: !!importId,
  });
}
```

### `web/dashboard/src/hooks/useGitHubTemplates.ts`

```typescript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { githubApi } from '@/api/github';
import { githubKeys } from './useGitHubConnection';
import { toast } from 'sonner';

export function useGitHubTemplates() {
  return useQuery({
    queryKey: githubKeys.templates(),
    queryFn: () => githubApi.listTemplates(),
    staleTime: 5 * 60 * 1000,
  });
}

export function useCreateTemplate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Parameters<typeof githubApi.createTemplate>[0]) =>
      githubApi.createTemplate(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: githubKeys.templates() });
      toast.success('Template created');
    },
    onError: (error: Error) => toast.error(`Failed: ${error.message}`),
  });
}

export function useUpdateTemplate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: { id: string } & Partial<Parameters<typeof githubApi.createTemplate>[0]>) =>
      githubApi.updateTemplate(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: githubKeys.templates() });
      toast.success('Template updated');
    },
    onError: (error: Error) => toast.error(`Failed: ${error.message}`),
  });
}

export function useDeleteTemplate() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => githubApi.deleteTemplate(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: githubKeys.templates() });
      toast.success('Template deleted');
    },
    onError: (error: Error) => toast.error(`Failed: ${error.message}`),
  });
}
```

---

## 6. Zustand Store

### `web/dashboard/src/stores/githubStore.ts`

```typescript
import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { GitHubConnection, GitHubRepo, ScanResult } from '@/types/github';

interface GitHubState {
  // Connection
  connection: GitHubConnection | null;
  isConnected: boolean;

  // Selected repo for import
  selectedRepo: GitHubRepo | null;
  scanResult: ScanResult | null;

  // Import config (multi-step wizard state)
  importConfig: {
    selectedFunctions: string[]; // detected function names
    visibilityOverrides: Record<string, 'public' | 'private' | 'unlisted'>;
    globalVisibility: 'public' | 'private' | 'unlisted';
    autoSync: boolean;
    syncBranches: string[];
    environmentMappings: Record<string, string>;
    templateId: string | null;
  };

  // Active import tracking
  activeImportId: string | null;

  // Actions
  setConnection: (conn: GitHubConnection | null) => void;
  setSelectedRepo: (repo: GitHubRepo | null) => void;
  setScanResult: (result: ScanResult | null) => void;
  toggleFunctionSelection: (name: string) => void;
  selectAllFunctions: () => void;
  deselectAllFunctions: () => void;
  setVisibilityOverride: (name: string, visibility: 'public' | 'private' | 'unlisted') => void;
  setGlobalVisibility: (visibility: 'public' | 'private' | 'unlisted') => void;
  setAutoSync: (enabled: boolean) => void;
  setSyncBranches: (branches: string[]) => void;
  setEnvironmentMappings: (mappings: Record<string, string>) => void;
  setTemplateId: (id: string | null) => void;
  setActiveImportId: (id: string | null) => void;
  resetImportConfig: () => void;
  resetAll: () => void;
}

const defaultImportConfig = {
  selectedFunctions: [] as string[],
  visibilityOverrides: {} as Record<string, 'public' | 'private' | 'unlisted'>,
  globalVisibility: 'private' as const,
  autoSync: false,
  syncBranches: ['main'],
  environmentMappings: {} as Record<string, string>,
  templateId: null as string | null,
};

export const useGitHubStore = create<GitHubState>()(
  persist(
    (set, get) => ({
      connection: null,
      isConnected: false,
      selectedRepo: null,
      scanResult: null,
      importConfig: { ...defaultImportConfig },
      activeImportId: null,

      setConnection: (conn) => set({
        connection: conn,
        isConnected: !!conn && conn.status === 'active',
      }),

      setSelectedRepo: (repo) => set({ selectedRepo: repo }),
      setScanResult: (result) => set({ scanResult: result }),

      toggleFunctionSelection: (name) => set((state) => {
        const selected = state.importConfig.selectedFunctions;
        const newSelected = selected.includes(name)
          ? selected.filter((n) => n !== name)
          : [...selected, name];
        return { importConfig: { ...state.importConfig, selectedFunctions: newSelected } };
      }),

      selectAllFunctions: () => set((state) => {
        const allNames = state.scanResult?.functions.map((f) => f.name) ?? [];
        return { importConfig: { ...state.importConfig, selectedFunctions: allNames } };
      }),

      deselectAllFunctions: () => set((state) => ({
        importConfig: { ...state.importConfig, selectedFunctions: [] },
      })),

      setVisibilityOverride: (name, visibility) => set((state) => ({
        importConfig: {
          ...state.importConfig,
          visibilityOverrides: { ...state.importConfig.visibilityOverrides, [name]: visibility },
        },
      })),

      setGlobalVisibility: (visibility) => set((state) => ({
        importConfig: { ...state.importConfig, globalVisibility: visibility },
      })),

      setAutoSync: (enabled) => set((state) => ({
        importConfig: { ...state.importConfig, autoSync: enabled },
      })),

      setSyncBranches: (branches) => set((state) => ({
        importConfig: { ...state.importConfig, syncBranches: branches },
      })),

      setEnvironmentMappings: (mappings) => set((state) => ({
        importConfig: { ...state.importConfig, environmentMappings: mappings },
      })),

      setTemplateId: (id) => set((state) => ({
        importConfig: { ...state.importConfig, templateId: id },
      })),

      setActiveImportId: (id) => set({ activeImportId: id }),

      resetImportConfig: () => set({ importConfig: { ...defaultImportConfig } }),

      resetAll: () => set({
        connection: null,
        isConnected: false,
        selectedRepo: null,
        scanResult: null,
        importConfig: { ...defaultImportConfig },
        activeImportId: null,
      }),
    }),
    {
      name: 'github-storage',
      partialize: (state) => ({
        // Only persist connection + import config between page navigations
        importConfig: state.importConfig,
      }),
    }
  )
);
```

---

## 7. UI Components

All components follow the existing patterns: `cn()` for className merging, `framer-motion` for animations, `lucide-react` for icons, shadcn/ui base components, and the aviation/velocity brand colors.

### 7.1 `GitHubConnectButton.tsx`

**Purpose:** Primary CTA to initiate GitHub OAuth. Shows different states based on connection status.

```typescript
interface GitHubConnectButtonProps {
  variant?: 'default' | 'outline' | 'ghost';
  size?: 'default' | 'sm' | 'lg';
  onConnected?: () => void;
}
```

**Visual:**
- Not connected: `Button` with GitHub icon + "Connect GitHub" — gradient `linear-gradient(135deg, #24292e, #586069)`
- Connected: Green checkmark + "GitHub Connected" — `variant="outline"` with `border-success`
- Loading: `Loader2` spinner + "Connecting..."
- Error: Red border + "Connection Failed — Retry"

**Behavior:**
1. On click → calls `useGitHubConnect` mutation → redirects to GitHub OAuth
2. On return (via callback page) → invalidates connection query → shows success toast

### 7.2 `GitHubConnectionCard.tsx`

**Purpose:** Shows the connected GitHub account details. Used in Settings and the GitHub hub page.

```typescript
interface GitHubConnectionCardProps {
  connection: GitHubConnection;
  onDisconnect: () => void;
  onRefresh: () => void;
}
```

**Visual:**
```
┌──────────────────────────────────────────────────────┐
│  ┌──────┐                                            │
│  │ Avatar│  octocat                                   │
│  └──────┘  Connected since Jan 15, 2026              │
│            Scopes: repo, read:user, user:email        │
│            Token expires: in 28 days                  │
│                                                      │
│  Status: ● Active                                     │
│                                                      │
│  [Refresh Token]  [Disconnect]                        │
└──────────────────────────────────────────────────────┘
```

- Uses shadcn `Card` with avatar, username, scope badges, expiry info
- Status dot: green for active, amber for expiring-soon (< 7 days), red for expired
- Disconnect opens `AlertDialog` with warning

### 7.3 `GitHubRepoCard.tsx`

**Purpose:** Individual repo card in the repo browser grid.

```typescript
interface GitHubRepoCardProps {
  repo: GitHubRepo;
  selected?: boolean;
  onSelect: (repo: GitHubRepo) => void;
  onScan: (repo: GitHubRepo) => void;
}
```

**Visual:**
```
┌─────────────────────────────────────────────────────┐
│  📦 owner/repo-name                    ⭐ 1.2k     │
│  A short description of the repo                     │
│  ┌────────┐ ┌──────────┐ ┌─────────┐ ┌──────────┐ │
│  │ Go 85% │ │ Shell 15%│ │ Private │ │ Archived │ │
│  └────────┘ └──────────┘ └─────────┘ └──────────┘ │
│                                                      │
│  ● Detected 2 functions      Last pushed: 2h ago     │
│                                                      │
│  ┌──────────────────────────────────────────────┐   │
│  │ 🟢 greet-handler  src/handler.go  [Go] 95%  │   │
│  │ 🟡 health-check   src/health.go   [Go] 72%  │   │
│  └──────────────────────────────────────────────┘   │
│                                                      │
│  [ Import Selected ]  [ Scan ]  [ View on GitHub ]   │
└─────────────────────────────────────────────────────┘
```

- Uses shadcn `Card` with hover elevation (`hover:-translate-y-0.5 hover:shadow-lg`)
- Language badges: colored by language (Go = cyan, JS = yellow, Python = blue, etc.)
- Status badges: `Private`/`Public`, `Fork`, `Archived` (amber warning)
- Detected functions: expandable section with confidence indicator (green ≥80%, yellow ≥50%, red <50%)
- Default branch badge, stars count, last push time

### 7.4 `GitHubRepoGrid.tsx` / `GitHubRepoListItem.tsx`

**Purpose:** Grid and list views for repos. Follows the `ToggleButtonGroup` pattern from FunctionsPage.

```typescript
interface GitHubRepoGridProps {
  repos: GitHubRepo[];
  isLoading: boolean;
  selectedRepoIds: Set<string>;
  onSelectRepo: (repo: GitHubRepo) => void;
  onScanRepo: (repo: GitHubRepo) => void;
  viewMode: 'grid' | 'list';
}
```

- Grid: `grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4`
- List: `DataTable` with columns: Name, Language, Visibility, Stars, Functions Detected, Last Push, Actions
- Empty state: `NoReposFound` component
- Loading: `SkeletonCard` × 6

### 7.5 `DetectedFunctionCard.tsx`

**Purpose:** Shows a single detected function from a repo scan, with selection checkbox and config overrides.

```typescript
interface DetectedFunctionCardProps {
  func: DetectedFunction;
  selected: boolean;
  visibility: 'public' | 'private' | 'unlisted';
  onToggle: () => void;
  onVisibilityChange: (v: 'public' | 'private' | 'unlisted') => void;
  onNameChange: (name: string) => void;
}
```

**Visual:**
```
┌─────────────────────────────────────────────────────┐
│  ☑  greet-handler                                    │
│     Entry: src/handler.go    Runtime: Go    95%      │
│     ┌─────────────┐ ┌──────────────────────────┐    │
│     │ 🔒 Private ▼│ │ Rename: [greet-handler ] │    │
│     └─────────────┘ └──────────────────────────┘    │
│     Dependencies: go.mod (3 packages)                 │
└─────────────────────────────────────────────────────┘
```

- Checkbox for selection
- Confidence bar (green/yellow/red gradient)
- Visibility dropdown (Globe = public, Lock = private, Eye-off = unlisted)
- Inline rename field
- Runtime icon from `@/components/icons/`
- Dependency info (manager + package count)

### 7.6 `FunctionDetectionResults.tsx`

**Purpose:** Panel showing all detected functions from a scan, with bulk actions.

```typescript
interface FunctionDetectionResultsProps {
  scanResult: ScanResult;
  selectedFunctions: string[];
  onToggleFunction: (name: string) => void;
  onSelectAll: () => void;
  onDeselectAll: () => void;
  onRescan: () => void;
}
```

**Visual:**
```
┌──────────────────────────────────────────────────────┐
│  🔍 Scan Results                           [Rescan]  │
│  Found 3 functions · Confidence: 87% · Strategy: AI  │
│  ┌────────────────────────────────────────────────┐  │
│  │ ☑ All  (3 selected)                            │  │
│  └────────────────────────────────────────────────┘  │
│                                                       │
│  [DetectedFunctionCard × 3]                          │
│                                                       │
│  ⚠ Warnings:                                         │
│  • No functionfly.jsonc found — will auto-generate   │
│  • Large dependency tree (142 packages)               │
│                                                       │
│  Estimated import time: ~45 seconds                   │
│  Estimated cost: $0.10                                │
└──────────────────────────────────────────────────────┘
```

- Header with scan metadata (strategy, confidence, time estimate)
- Select all / deselect all checkboxes
- List of `DetectedFunctionCard` components
- Warnings section with amber alert icons
- Cost estimation

### 7.7 `ImportConfigForm.tsx`

**Purpose:** Configuration form for the import (global settings + sync config).

```typescript
interface ImportConfigFormProps {
  repo: GitHubRepo;
  config: GitHubState['importConfig'];
  onChange: (config: Partial<GitHubState['importConfig']>) => void;
  templates: GitHubTemplate[];
  branches: Array<{ name: string; is_default: boolean }>;
}
```

**Sections:**
1. **Global Visibility** — `VisibilitySelector` (default for all functions)
2. **Auto-Sync** — `Switch` + branch multi-select
3. **Environment Mappings** — `EnvironmentMappingEditor` (branch → env table)
4. **Template** — `Select` dropdown of saved templates, "Save as Template" button

### 7.8 `ImportPreviewCard.tsx`

**Purpose:** Shows dry-run results before confirming import.

```typescript
interface ImportPreviewCardProps {
  preview: ImportPreview;
  onConfirm: () => void;
  onCancel: () => void;
  isImporting: boolean;
}
```

**Visual:**
```
┌──────────────────────────────────────────────────────┐
│  📋 Import Preview                                    │
│                                                       │
│  ┌────────────────────────────────────────────────┐  │
│  │ Function        │ Runtime │ Visibility │ Cost  │  │
│  │─────────────────│─────────│────────────│───────│  │
│  │ greet-handler   │ Go      │ 🔒 Private │ $0.05│  │
│  │ health-check    │ Go      │ 🌐 Public  │ $0.03│  │
│  │ api-router      │ Go      │ 🔒 Private │ $0.07│  │
│  └────────────────────────────────────────────────┘  │
│                                                       │
│  Total estimated cost: $0.15                          │
│                                                       │
│  ⚠️ Conflicts:                                        │
│  • "api-router" already exists → [Overwrite] [Rename]│
│                                                       │
│  [Cancel]              [Confirm Import — $0.15]       │
└──────────────────────────────────────────────────────┘
```

### 7.9 `ImportProgressStepper.tsx`

**Purpose:** Real-time import progress display using SSE. The centerpiece of the import UX.

```typescript
interface ImportProgressStepperProps {
  importId: string;
  onComplete: (result: ImportCompleteEvent) => void;
  onError: (error: ImportErrorEvent) => void;
}
```

**Visual (uses existing `Stepper` component):**
```
┌──────────────────────────────────────────────────────┐
│  Importing: owner/repo → greet-handler               │
│                                                       │
│  ● Scanning          ✓ Complete (2 functions found)  │
│  ● Fetching          ✓ Complete (14 files, 23 KB)    │
│  ● Building          ◐ In Progress...                 │
│     ┌────────────────────────────────────────────┐   │
│     │  Compiling handler.go → WASM (4.2 MB)      │   │
│     │  ████████████████████░░░░░░  67%            │   │
│     └────────────────────────────────────────────┘   │
│  ○ Publishing        Pending                          │
│                                                       │
│  ───────────────────────────────────────────────────  │
│  Build output (scrollable):                           │
│  > go: downloading github.com/some/dep v1.2.3        │
│  > go: compiling to WASM...                           │
│  > wasm-opt: optimizing...                            │
└──────────────────────────────────────────────────────┘
```

- Uses `useImportProgress` hook (SSE)
- `Stepper` component with 4 steps: Scanning, Fetching, Building, Publishing
- Active step shows: progress bar, current operation text, build output (scrollable `ScrollArea`)
- Completed steps show: checkmark + summary
- Error state: red step indicator + error message + "Retry" button
- Success state: confetti + "View Function" button (links to function detail page)

### 7.10 `ImportStatusBadge.tsx`

**Purpose:** Badge showing import status. Used in imports list.

```typescript
interface ImportStatusBadgeProps {
  status: GitHubImport['status'];
  size?: 'sm' | 'default';
}
```

**Visual mapping:**
| Status | Badge |
|--------|-------|
| `pending` | Gray dot + "Pending" |
| `scanning` | Blue dot + "Scanning" + spinner |
| `fetching` | Blue dot + "Fetching" + spinner |
| `building` | Blue dot + "Building" + spinner |
| `publishing` | Blue dot + "Publishing" + spinner |
| `completed` | Green dot + "Completed" |
| `failed` | Red dot + "Failed" |
| `cancelled` | Gray dot + "Cancelled" |

Uses `Badge variant="outline"` with custom color classes.

### 7.11 `SyncStatusIndicator.tsx`

**Purpose:** Shows auto-sync status and last sync info.

```typescript
interface SyncStatusIndicatorProps {
  importId: string;
  autoSyncEnabled: boolean;
  syncBranches: string[];
  lastSyncAt?: string;
  lastSyncStatus?: string;
  onToggle: (enabled: boolean) => void;
}
```

**Visual:**
```
┌─────────────────────────────────────────────────────┐
│  🔄 Auto-Sync  [ON/OFF toggle]                      │
│  Branches: main, staging                             │
│  Last sync: 5 minutes ago · ✓ success · v1.2.3+abc  │
│  [View Sync History]                                 │
└─────────────────────────────────────────────────────┘
```

### 7.12 `VisibilitySelector.tsx`

**Purpose:** Dropdown/toggle for public/private/unlisted visibility.

```typescript
interface VisibilitySelectorProps {
  value: 'public' | 'private' | 'unlisted';
  onChange: (value: 'public' | 'private' | 'unlisted') => void;
  variant?: 'dropdown' | 'segmented';
}
```

**Visual (segmented variant):**
```
┌────────────┬────────────┬────────────┐
│ 🌐 Public  │ 🔒 Private │ 👁 Unlisted│
└────────────┴────────────┴────────────┘
```

- Uses `ToggleButtonGroup` pattern
- Globe icon (green), Lock icon (muted), Eye-off icon (yellow)

### 7.13 `EnvironmentMappingEditor.tsx`

**Purpose:** Maps branches to environment names.

```typescript
interface EnvironmentMappingEditorProps {
  mappings: Record<string, string>;
  branches: Array<{ name: string; is_default: boolean }>;
  onChange: (mappings: Record<string, string>) => void;
}
```

**Visual:**
```
┌─────────────────────────────────────────────────────┐
│  Branch → Environment Mapping                        │
│                                                      │
│  main     → [production ▼]     ●                     │
│  staging  → [staging    ▼]                           │
│  dev      → [development▼]                           │
│                                                      │
│  [+ Add Mapping]                                     │
└─────────────────────────────────────────────────────┘
```

- Table with branch name (from repo) → environment `Select` dropdown
- Default mapping: `main` → `production`
- Add/remove rows
- Environment options: `production`, `staging`, `development`, `preview`, custom

### 7.14 `GitHubStatsBar.tsx`

**Purpose:** Summary stats bar at the top of the GitHub hub page.

```typescript
interface GitHubStatsBarProps {
  totalRepos: number;
  importedCount: number;
  activeSyncs: number;
  recentImports: number; // last 7 days
}
```

**Visual (4 stat cards in a row):**
```
┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ 📦 42        │ │ ✅ 12        │ │ 🔄 5         │ │ 📈 3         │
│ Repos        │ │ Imported     │ │ Active Syncs │ │ This Week    │
└──────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
```

- Uses `GlassmorphismCard` or standard shadcn `Card`
- Animated number transitions (`motion.span` with `layout`)

### 7.15 `ImportTemplateCard.tsx` / `ImportTemplateForm.tsx`

**Purpose:** Manage reusable import configurations.

**Card Visual:**
```
┌─────────────────────────────────────────────────────┐
│  📋 Node.js API Template                             │
│  Default config for Node.js API repos                │
│  Runtime: node20 · Visibility: private · Sync: on   │
│  Used 8 times                                        │
│  [Edit]  [Delete]  [Apply to Import]                 │
└─────────────────────────────────────────────────────┘
```

### 7.16 `ConflictResolutionDialog.tsx`

**Purpose:** Modal to resolve naming conflicts when importing.

```typescript
interface ConflictResolutionDialogProps {
  conflicts: ImportPreview['conflicts'];
  onResolve: (resolutions: Record<string, 'overwrite' | 'rename' | 'skip' | 'new_version'>) => void;
  onCancel: () => void;
}
```

**Visual:**
```
┌──────────────────────────────────────────────────────┐
│  ⚠️ Import Conflicts Detected                        │
│                                                       │
│  "api-router" already exists in your registry         │
│  Current version: 2.1.0 · Published 3 days ago       │
│                                                       │
│  How would you like to resolve this?                  │
│                                                       │
│  ○ Overwrite — Replace the existing function          │
│  ○ Rename    — Create as "api-router-v2"              │
│  ○ New Version — Add as version 2.2.0                 │
│  ○ Skip     — Don't import this function              │
│                                                       │
│  [Cancel]                    [Apply & Continue]        │
└──────────────────────────────────────────────────────┘
```

### 7.17 Empty States

**`NoGitHubConnection.tsx`:**
```
┌──────────────────────────────────────────────────────┐
│                                                       │
│            [GitHub logo - 64px]                       │
│                                                       │
│        Connect your GitHub account                    │
│   Import repositories and auto-create functions       │
│                                                       │
│        [Connect GitHub]                               │
│                                                       │
│   We only request access to your repositories.        │
│   Your token is encrypted with AES-256-GCM.          │
└──────────────────────────────────────────────────────┘
```

Uses `EmptyState` component with `variant="card"`, GitHub icon.

**`NoReposFound.tsx`:**
```
┌──────────────────────────────────────────────────────┐
│            [Search icon - 64px]                      │
│        No repositories found                          │
│   Try adjusting your filters or refresh your repos   │
│        [Refresh Repos]                                │
└──────────────────────────────────────────────────────┘
```

**`NoImportsYet.tsx`:**
```
┌──────────────────────────────────────────────────────┐
│            [Import icon - 64px]                      │
│        No imports yet                                 │
│   Import your first function from a GitHub repo      │
│        [Browse Repos]                                 │
└──────────────────────────────────────────────────────┘
```

---

## 8. Page Components

### 8.1 `GitHubPage/index.tsx` — Main Hub

**Route:** `/github` (protected)
**Layout:** `PageLayout` → `PageHeader` → `Tabs`

```
PageHeader:
  title: "GitHub Integration"
  subtitle: "Import and sync functions from GitHub repositories"
  actions: [Connect GitHub] (if not connected) or [Refresh Repos] (if connected)
  badges: [Beta]

Tabs:
  Tab 1: "Repositories" → GitHubReposTab
  Tab 2: "Imports" → GitHubImportsTab
  Tab 3: "Templates" → GitHubTemplatesTab
```

**If not connected:** Show `NoGitHubConnection` empty state.

### 8.2 `GitHubPage/GitHubReposTab.tsx`

**Layout:**
```
GitHubStatsBar (top)
    │
GitHubRepoSearchFilter (search + language + visibility filters)
    │
ToggleButtonGroup (grid/list) + Refresh button
    │
GitHubRepoGrid or DataTable (based on viewMode)
    │
Pagination
```

**Behavior:**
- Repos load from cache (`useGitHubRepos`)
- Search filters client-side (debounced 300ms)
- Language filter: `Select` dropdown populated from repo data
- Visibility filter: `Select` (All/Public/Private)
- Click repo card → navigate to `/github/repos/:repoId/import`

### 8.3 `GitHubPage/GitHubImportsTab.tsx`

**Layout:**
```
Filter bar: Status filter (all/completed/failed/in-progress) + Search
    │
DataTable with columns:
  | Repository | Function | Visibility | Status | Last Sync | Created | Actions |
    │
Pagination
```

**Row actions (DropdownMenu):**
- View Function (link to function detail)
- View Sync History
- Resync Now
- Update Sync Settings
- Cancel (if in-progress)
- Retry (if failed)
- Delete Import

### 8.4 `GitHubPage/GitHubTemplatesTab.tsx`

**Layout:**
```
[Create Template] button (top-right)
    │
Grid of ImportTemplateCard components
    │
Empty state if no templates
```

### 8.5 `GitHubRepoImportPage/index.tsx` — Import Wizard

**Route:** `/github/repos/:repoId/import` (protected)
**Layout:** Full-page wizard with `Stepper`

**Steps:**

**Step 1: Scan**
- Shows `GitHubRepoCard` (expanded) with repo details
- "Scan for Functions" button → calls `useScanGitHubRepo`
- Shows `FunctionDetectionResults` when scan completes
- Auto-advances to Step 2 if functions found

**Step 2: Configure**
- `DetectedFunctionCard` list with selection + visibility
- `ImportConfigForm` (global settings)
- `ImportTemplateCard` selection (optional)
- "Preview Import" button → shows `DryRunPreviewDialog`

**Step 3: Confirm & Import**
- `ImportPreviewCard` with dry-run results
- Conflict resolution if needed (`ConflictResolutionDialog`)
- Cost display
- "Confirm Import" button → starts import

**Step 4: Progress**
- `ImportProgressStepper` with SSE real-time updates
- Build output stream
- On complete: confetti + "View Function" / "Import Another" buttons
- On error: error details + "Retry" button

**Navigation:**
- Back button (disabled on Step 4)
- Skip button (for Step 1 if already scanned)
- Cancel → AlertDialog "Discard import?"

### 8.6 `GitHubSettingsPage/index.tsx`

**Route:** `/settings#github` (within existing Settings tabs)
**OR:** New tab "GitHub" in Settings page

**Layout:**
```
GitHubConnectionCard (if connected)
    │
Section: "Connection Details"
  - Username, avatar, scopes, token expiry
  - [Refresh Token] [Disconnect]

Section: "GitHub App (Optional)"
  - For fine-grained per-repo access
  - [Install GitHub App] button

Section: "Import Defaults"
  - Default visibility for new imports
  - Default auto-sync setting
  - Default sync branches

Section: "Danger Zone"
  - [Disconnect GitHub] with confirmation
  - [Delete All Import Records] with confirmation
```

---

## 9. Routing & Navigation

Add to `App.tsx`:

```typescript
// Inside the protected routes (under DashboardLayout)
<Route path="/github" element={<GitHubPage />} />
<Route path="/github/repos/:repoId/import" element={<GitHubRepoImportPage />} />
```

Add to Settings page tabs:

```typescript
// In SettingsContent.tsx — add "GitHub" tab
{ value: 'github', label: t('settings.tabs.github'), icon: Github },
```

---

## 10. Sidebar Integration

Add to `Sidebar.tsx` in the "Deploy" section:

```typescript
{
  path: '/github',
  label: 'GitHub Import',
  icon: Github, // from lucide-react
  badge: isNew ? 'new' : undefined,
  description: 'Import functions from GitHub repositories',
}
```

Position: After "Functions" in the "Build" section, before "Deploy" section.

---

## 11. i18n Keys

Add to translation files (`en.json` and other locales):

```json
{
  "github": {
    "title": "GitHub Integration",
    "subtitle": "Import and sync functions from GitHub repositories",
    "connect": "Connect GitHub",
    "disconnect": "Disconnect",
    "connected": "Connected",
    "notConnected": "Not Connected",
    "repos": {
      "title": "Repositories",
      "search": "Search repositories...",
      "refresh": "Refresh",
      "noRepos": "No repositories found",
      "noReposDesc": "Connect your GitHub account to see your repositories",
      "private": "Private",
      "public": "Public",
      "fork": "Fork",
      "archived": "Archived",
      "stars": "Stars",
      "lastPush": "Last pushed",
      "detectedFunctions": "Detected Functions",
      "importSelected": "Import Selected ({{count}})",
      "importAll": "Import All",
      "scan": "Scan for Functions"
    },
    "import": {
      "title": "Import Functions",
      "scanning": "Scanning for functions...",
      "configuring": "Configure Import",
      "preview": "Import Preview",
      "confirm": "Confirm Import",
      "progress": "Importing...",
      "completed": "Import Completed!",
      "failed": "Import Failed",
      "estimatedCost": "Estimated cost",
      "estimatedTime": "Estimated time",
      "selectFunctions": "Select functions to import",
      "selectAll": "Select All",
      "deselectAll": "Deselect All"
    },
    "visibility": {
      "public": "Public",
      "private": "Private",
      "unlisted": "Unlisted",
      "publicDesc": "Visible in the public registry",
      "privateDesc": "Only visible to your team",
      "unlistedDesc": "Accessible by direct link only"
    },
    "sync": {
      "title": "Auto-Sync",
      "enable": "Enable auto-sync",
      "branches": "Sync branches",
      "lastSync": "Last sync",
      "never": "Never",
      "history": "Sync History",
      "envMappings": "Environment Mappings",
      "addMapping": "Add Mapping"
    },
    "templates": {
      "title": "Import Templates",
      "create": "Create Template",
      "noTemplates": "No templates yet",
      "noTemplatesDesc": "Save an import configuration as a template for reuse"
    },
    "conflicts": {
      "title": "Import Conflicts",
      "alreadyExists": "\"{{name}}\" already exists",
      "overwrite": "Overwrite",
      "rename": "Rename",
      "newVersion": "New Version",
      "skip": "Skip"
    },
    "stepper": {
      "scan": "Scan",
      "configure": "Configure",
      "confirm": "Confirm",
      "import": "Import"
    }
  }
}
```

---

## Summary: Complete Component Inventory

| # | Component | Type | Location |
|---|-----------|------|----------|
| 1 | `GitHubConnectButton` | Interactive | `components/github/` |
| 2 | `GitHubConnectionCard` | Display | `components/github/` |
| 3 | `GitHubRepoCard` | Interactive | `components/github/` |
| 4 | `GitHubRepoGrid` | Layout | `components/github/` |
| 5 | `GitHubRepoListItem` | Display | `components/github/` |
| 6 | `DetectedFunctionCard` | Interactive | `components/github/` |
| 7 | `FunctionDetectionResults` | Display | `components/github/` |
| 8 | `ImportConfigForm` | Form | `components/github/` |
| 9 | `ImportPreviewCard` | Display | `components/github/` |
| 10 | `ImportProgressStepper` | Interactive | `components/github/` |
| 11 | `ImportStatusBadge` | Display | `components/github/` |
| 12 | `SyncStatusIndicator` | Interactive | `components/github/` |
| 13 | `VisibilitySelector` | Interactive | `components/github/` |
| 14 | `EnvironmentMappingEditor` | Form | `components/github/` |
| 15 | `GitHubStatsBar` | Display | `components/github/` |
| 16 | `ImportTemplateCard` | Interactive | `components/github/` |
| 17 | `ImportTemplateForm` | Form | `components/github/` |
| 18 | `ConflictResolutionDialog` | Modal | `components/github/` |
| 19 | `DryRunPreviewDialog` | Modal | `components/github/` |
| 20 | `NoGitHubConnection` | Empty State | `components/github/EmptyStates/` |
| 21 | `NoReposFound` | Empty State | `components/github/EmptyStates/` |
| 22 | `NoImportsYet` | Empty State | `components/github/EmptyStates/` |
| 23 | `GitHubPage` | Page | `pages/GitHubPage/` |
| 24 | `GitHubReposTab` | Tab Content | `pages/GitHubPage/` |
| 25 | `GitHubImportsTab` | Tab Content | `pages/GitHubPage/` |
| 26 | `GitHubTemplatesTab` | Tab Content | `pages/GitHubPage/` |
| 27 | `GitHubRepoImportPage` | Page (Wizard) | `pages/GitHubRepoImportPage/` |
| 28 | `ImportExecutionStep` | Step Content | `pages/GitHubRepoImportPage/` |
| 29 | `GitHubSettingsPage` | Page | `pages/GitHubSettingsPage/` |
| 30 | `github.ts` (API) | API Client | `api/` |
| 31 | `github.ts` (types) | Zod Schemas | `types/` |
| 32 | `useGitHubConnection.ts` | Hook | `hooks/` |
| 33 | `useGitHubRepos.ts` | Hook | `hooks/` |
| 34 | `useGitHubImport.ts` | Hook | `hooks/` |
| 35 | `useGitHubSync.ts` | Hook | `hooks/` |
| 36 | `useGitHubTemplates.ts` | Hook | `hooks/` |
| 37 | `githubStore.ts` | Zustand Store | `stores/` |

**Total: 37 files** (22 components, 3 pages, 2 page sections, 5 hooks, 1 store, 1 API client, 1 types file, 2 empty states)
