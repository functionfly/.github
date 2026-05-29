import type { StudioFile, StudioProject, StudioProjectsState } from '@/lib/studio-projects';
import { apiClient } from './client';

export interface StudioProjectFileApi {
  id: string;
  name: string;
  path: string;
  content: string;
  language: string;
  created_at?: string;
  updated_at?: string;
}

export interface StudioProjectApi {
  id: string;
  name: string;
  files: StudioProjectFileApi[];
  created_at: string;
  updated_at: string;
}

export interface StudioWorkspaceApi {
  projects: StudioProjectApi[];
  active_project_id: string | null;
  active_file_id: string | null;
}

function mapFile(file: StudioProjectFileApi): StudioFile {
  return {
    id: file.id,
    name: file.name,
    path: file.path,
    content: file.content,
    language: file.language,
  };
}

function mapProject(project: StudioProjectApi): StudioProject {
  return {
    id: project.id,
    name: project.name,
    files: (project.files ?? []).map(mapFile),
    createdAt: project.created_at,
    updatedAt: project.updated_at,
  };
}

export function mapWorkspace(api: StudioWorkspaceApi): StudioProjectsState {
  return {
    projects: (api.projects ?? []).map(mapProject),
    activeProjectId: api.active_project_id,
    activeFileId: api.active_file_id,
  };
}

export const studioProjectsApi = {
  getWorkspace: async (): Promise<StudioProjectsState> => {
    const res = await apiClient.get<StudioWorkspaceApi>('/v1/studio/workspace');
    return mapWorkspace(res);
  },

  saveSession: (activeProjectId: string | null, activeFileId: string | null) =>
    apiClient.put<{ active_project_id: string | null; active_file_id: string | null }>(
      '/v1/studio/workspace/session',
      {
        active_project_id: activeProjectId,
        active_file_id: activeFileId,
      }
    ),

  createProject: async (name: string, starterContent?: string) => {
    const res = await apiClient.post<{ project: StudioProjectApi }>('/v1/studio/projects', {
      name,
      starter_content: starterContent,
    });
    return mapProject(res.project);
  },

  updateProject: async (projectId: string, name: string) => {
    const res = await apiClient.patch<{ project: StudioProjectApi }>(
      `/v1/studio/projects/${projectId}`,
      { name }
    );
    return mapProject(res.project);
  },

  deleteProject: (projectId: string) =>
    apiClient.delete<{ message: string }>(`/v1/studio/projects/${projectId}`),

  duplicateProject: async (projectId: string) => {
    const res = await apiClient.post<{ project: StudioProjectApi }>(
      `/v1/studio/projects/${projectId}/duplicate`
    );
    return mapProject(res.project);
  },

  createFile: async (
    projectId: string,
    name: string,
    options?: { dir?: string; content?: string }
  ) => {
    const res = await apiClient.post<{ file: StudioProjectFileApi }>(
      `/v1/studio/projects/${projectId}/files`,
      {
        name,
        dir: options?.dir ?? 'src',
        content: options?.content,
      }
    );
    return mapFile(res.file);
  },

  updateFile: async (
    projectId: string,
    fileId: string,
    updates: Partial<Pick<StudioFile, 'name' | 'path' | 'content' | 'language'>>
  ) => {
    const res = await apiClient.patch<{ file: StudioProjectFileApi }>(
      `/v1/studio/projects/${projectId}/files/${fileId}`,
      updates
    );
    return mapFile(res.file);
  },

  deleteFile: (projectId: string, fileId: string) =>
    apiClient.delete<{ message: string }>(`/v1/studio/projects/${projectId}/files/${fileId}`),
};
