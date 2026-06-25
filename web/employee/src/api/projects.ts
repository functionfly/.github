import apiClient from './client';

export interface Project {
  id: string;
  tenant_id: string;
  name: string;
  slug: string;
  description?: string;
  status: string;
  priority: string;
  owner_id: string;
  start_date?: string;
  target_date?: string;
  tags?: string[];
  created_at: string;
  updated_at: string;
}

export interface Task {
  id: string;
  project_id: string;
  tenant_id: string;
  title: string;
  description?: string;
  status: string;
  priority: string;
  assignee_id?: string;
  reporter_id: string;
  parent_id?: string;
  due_date?: string;
  estimated_hours?: number;
  actual_hours?: number;
  tags?: string[];
  position: number;
  created_at: string;
  updated_at: string;
}

export interface TaskComment {
  id: number;
  task_id: string;
  author_id: string;
  body: string;
  created_at: string;
}

export const projectsApi = {
  list: (params?: { status?: string; limit?: number; offset?: number }) =>
    apiClient.get<{ projects: Project[]; total: number }>('/v1/projects', { params }),

  get: (id: string) =>
    apiClient.get<{ project: Project }>(`/v1/projects/${id}`),

  create: (data: Partial<Project>) =>
    apiClient.post<{ project: Project }>('/v1/projects', data),

  update: (id: string, data: Partial<Project>) =>
    apiClient.patch<{ project: Project }>(`/v1/projects/${id}`, data),
};

export const tasksApi = {
  list: (params?: { project_id?: string; assignee_id?: string; status?: string; limit?: number; offset?: number }) =>
    apiClient.get<{ tasks: Task[]; total: number }>('/v1/tasks', { params }),

  get: (id: string) =>
    apiClient.get<{ task: Task }>(`/v1/tasks/${id}`),

  create: (data: Partial<Task>) =>
    apiClient.post<{ task: Task }>('/v1/tasks', data),

  update: (id: string, data: Partial<Task>) =>
    apiClient.patch<{ task: Task }>(`/v1/tasks/${id}`, data),

  assign: (id: string, assigneeId: string) =>
    apiClient.post(`/v1/tasks/${id}/assign`, { assignee_id: assigneeId }),

  listComments: (taskId: string) =>
    apiClient.get<{ comments: TaskComment[] }>(`/v1/tasks/${taskId}/comments`),

  addComment: (taskId: string, body: string) =>
    apiClient.post<{ comment: TaskComment }>(`/v1/tasks/${taskId}/comments`, { body }),
};
