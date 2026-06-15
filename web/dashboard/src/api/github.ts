import { z } from 'zod';
import { tokenVault } from '@/utils/token-vault';
import { apiClient } from './client';
import {
  githubConnectionSchema,
  githubRepoSchema,
  githubRepoListSchema,
  scanResultSchema,
  githubImportSchema,
  githubImportListSchema,
  bulkImportResultSchema,
  githubSyncLogListSchema,
  githubTemplateSchema,
  githubTemplateListSchema,
  importPreviewSchema,
  branchListSchema,
  treeResponseSchema,
} from '@/types/github';
import type {
  GitHubConnection,
  GitHubRepo,
  Branch,
  ScanResult,
  GitHubImport,
  BulkImportResult,
  GitHubSyncLog,
  GitHubTemplate,
  ImportPreview,
  TreeResponse,
  ScanRepoOptions,
  StartImportRequest,
  BulkImportRequest,
  UpdateSyncRequest,
  ListReposParams,
  ListImportsParams,
  GetTreeParams,
  ListSyncLogsParams,
  CreateTemplateRequest,
  UpdateTemplateRequest,
  PaginatedImportsResponse,
  PaginatedSyncLogsResponse,
  PaginatedReposResponse,
  ImportProgressEvent,
  ImportCompleteEvent,
  ImportErrorEvent,
} from '@/types/github';

export const githubApi = {
  // ==========================================================================
  // Connection
  // ==========================================================================

  getConnectUrl: async (): Promise<{ url: string }> => {
    return apiClient.get<{ url: string }>('/v1/github/connect');
  },

  getConnection: async (): Promise<GitHubConnection | null> => {
    try {
      return await apiClient.getValidatedData<GitHubConnection>(
        githubConnectionSchema,
        '/v1/github/connection'
      );
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : String(error);
      if (
        errorMsg.includes('not_found') ||
        errorMsg.includes('No GitHub') ||
        errorMsg.includes('404')
      ) {
        return null;
      }
      throw error;
    }
  },

  disconnect: async (): Promise<void> => {
    await apiClient.delete('/v1/github/connection');
  },

  refreshToken: async (): Promise<GitHubConnection> => {
    return apiClient.postValidatedData<GitHubConnection>(
      githubConnectionSchema,
      '/v1/github/connection/refresh'
    );
  },

  // ==========================================================================
  // Repositories
  // ==========================================================================

  listRepos: async (params?: ListReposParams): Promise<PaginatedReposResponse> => {
    console.debug('[githubApi] listRepos params:', params);
    const response = await apiClient.get<{
      repos: unknown[];
      total: number;
      page: number;
      per_page: number;
    }>('/v1/github/repos', { params });
    console.debug('[githubApi] listRepos raw response:', { total: response.total, page: response.page, per_page: response.per_page, repoCount: response.repos?.length });
    try {
      const parsed = githubRepoListSchema.parse(response.repos);
      console.debug('[githubApi] listRepos parse OK, returning', parsed.length, 'repos');
      return {
        repos: parsed,
        total: response.total,
        page: response.page,
        per_page: response.per_page,
      };
    } catch (parseErr) {
      console.error('[githubApi] listRepos ZOD PARSE ERROR:', parseErr);
      console.error('[githubApi] First repo sample that failed:', response.repos?.[0]);
      throw parseErr;
    }
  },

  refreshRepos: async (): Promise<{ refreshed: number }> => {
    return apiClient.post<{ refreshed: number }>('/v1/github/repos/refresh');
  },

  getRepo: async (repoId: string): Promise<GitHubRepo> => {
    return apiClient.getValidatedData<GitHubRepo>(
      githubRepoSchema,
      `/v1/github/repos/${repoId}`
    );
  },

  scanRepo: async (repoId: string, options?: ScanRepoOptions): Promise<ScanResult> => {
    return apiClient.postValidatedData<ScanResult>(
      scanResultSchema,
      `/v1/github/repos/${repoId}/scan`,
      options
    );
  },

  listBranches: async (repoId: string): Promise<Branch[]> => {
    return apiClient.getValidatedData<Branch[]>(
      branchListSchema,
      `/v1/github/repos/${repoId}/branches`
    );
  },

  getTree: async (repoId: string, params?: GetTreeParams): Promise<TreeResponse> => {
    return apiClient.getValidatedData<TreeResponse>(
      treeResponseSchema,
      `/v1/github/repos/${repoId}/tree`,
      { params }
    );
  },

  // ==========================================================================
  // Imports
  // ==========================================================================

  startImport: async (data: StartImportRequest): Promise<GitHubImport> => {
    return apiClient.postValidatedData<GitHubImport>(
      githubImportSchema,
      '/v1/github/imports',
      data
    );
  },

  previewImport: async (data: StartImportRequest): Promise<ImportPreview> => {
    const response = await apiClient.post<ImportPreview>(
      '/v1/github/imports/preview',
      data
    );
    return importPreviewSchema.parse(response);
  },

  previewBulkImport: async (data: BulkImportRequest): Promise<ImportPreview[]> => {
    const response = await apiClient.post<ImportPreview[]>(
      '/v1/github/imports/preview/bulk',
      data
    );
    return (response as ImportPreview[]).map((p) => importPreviewSchema.parse(p));
  },

  bulkImport: async (data: BulkImportRequest): Promise<BulkImportResult[]> => {
    const response = await apiClient.post<{ imports: unknown[] }>(
      '/v1/github/imports/bulk',
      data
    );
    return z.array(bulkImportResultSchema).parse(response.imports);
  },

  listImports: async (params?: ListImportsParams): Promise<PaginatedImportsResponse> => {
    const response = await apiClient.get<{
      imports: unknown[];
      total: number;
      page: number;
      per_page: number;
    }>('/v1/github/imports', { params });
    return {
      imports: githubImportListSchema.parse(response.imports),
      total: response.total,
      page: response.page,
      per_page: response.per_page,
    };
  },

  getImport: async (importId: string): Promise<GitHubImport> => {
    return apiClient.getValidatedData<GitHubImport>(
      githubImportSchema,
      `/v1/github/imports/${importId}`
    );
  },

  cancelImport: async (importId: string): Promise<{ status: string }> => {
    return apiClient.post<{ status: string }>(
      `/v1/github/imports/${importId}/cancel`
    );
  },

  retryImport: async (importId: string): Promise<{ status: string }> => {
    return apiClient.post<{ status: string }>(
      `/v1/github/imports/${importId}/retry`
    );
  },

  resyncImport: async (importId: string): Promise<{ status: string }> => {
    return apiClient.post<{ status: string }>(
      `/v1/github/imports/${importId}/resync`
    );
  },

  // ==========================================================================
  // Sync
  // ==========================================================================

  updateSync: async (importId: string, data: UpdateSyncRequest): Promise<void> => {
    await apiClient.put(`/v1/github/imports/${importId}/sync`, data);
  },

  getSyncLogs: async (
    importId: string,
    params?: ListSyncLogsParams
  ): Promise<PaginatedSyncLogsResponse> => {
    const response = await apiClient.get<{
      logs: unknown[];
      total: number;
      page: number;
      per_page: number;
    }>(`/v1/github/imports/${importId}/sync-logs`, { params });
    return {
      logs: githubSyncLogListSchema.parse(response.logs),
      total: response.total,
      page: response.page,
      per_page: response.per_page,
    };
  },

  // ==========================================================================
  // Templates
  // ==========================================================================

  listTemplates: async (): Promise<GitHubTemplate[]> => {
    return apiClient.getValidatedData<GitHubTemplate[]>(
      githubTemplateListSchema,
      '/v1/github/templates'
    );
  },

  createTemplate: async (data: CreateTemplateRequest): Promise<GitHubTemplate> => {
    return apiClient.postValidatedData<GitHubTemplate>(
      githubTemplateSchema,
      '/v1/github/templates',
      data
    );
  },

  updateTemplate: async (
    templateId: string,
    data: UpdateTemplateRequest
  ): Promise<GitHubTemplate> => {
    return apiClient.putValidatedData<GitHubTemplate>(
      githubTemplateSchema,
      `/v1/github/templates/${templateId}`,
      data
    );
  },

  deleteTemplate: async (templateId: string): Promise<void> => {
    await apiClient.delete(`/v1/github/templates/${templateId}`);
  },

  // ==========================================================================
  // SSE Import Progress
  // ==========================================================================

  streamImportProgress: async (
    importId: string,
    onProgress: (event: ImportProgressEvent) => void,
    onComplete: (event: ImportCompleteEvent) => void,
    onError: (event: ImportErrorEvent) => void
  ): Promise<EventSource> => {
    const baseUrl = apiClient.getBaseUrl();
    await tokenVault.initialize();
    const token = await tokenVault.getAccessToken();
    const url = `${baseUrl}/v1/github/imports/${importId}/progress?token=${encodeURIComponent(token || '')}`;

    const eventSource = new EventSource(url);

    eventSource.addEventListener('progress', (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data) as ImportProgressEvent;
        onProgress(data);
      } catch {
        console.error('[GitHub SSE] Failed to parse progress event');
      }
    });

    eventSource.addEventListener('complete', (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data) as ImportCompleteEvent;
        onComplete(data);
      } catch {
        console.error('[GitHub SSE] Failed to parse complete event');
      }
      eventSource.close();
    });

    eventSource.addEventListener('error', (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data) as ImportErrorEvent;
        onError(data);
      } catch {
        // Native EventSource error (connection lost, etc.)
        onError({
          stage: 'error',
          progress: 0,
          message: 'Connection to import progress stream lost',
        });
      }
      eventSource.close();
    });

    return eventSource;
  },
};
