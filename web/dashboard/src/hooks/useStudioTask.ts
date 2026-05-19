import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { API_BASE_URL } from '@/lib/constants';

export interface StudioTask {
  id: string;
  title: string;
  description: string;
  status: 'todo' | 'in-progress' | 'done' | 'blocked' | 'review';
  priority: 'low' | 'medium' | 'high';
  createdAt: string;
  updatedAt: string;
  assigneeId?: string;
}

export interface CreateTaskRequest {
  title: string;
  description?: string;
  priority?: 'low' | 'medium' | 'high';
  assigneeId?: string;
}

export interface UpdateTaskRequest {
  title?: string;
  description?: string;
  status?: 'todo' | 'in-progress' | 'done' | 'blocked' | 'review';
  priority?: 'low' | 'medium' | 'high';
  assigneeId?: string;
}

async function taskFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    credentials: 'include',
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }));
    throw new Error(error.message || `HTTP ${response.status}`);
  }

  return response.json();
}

export function useStudioTasks() {
  return useQuery({
    queryKey: ['studio', 'tasks'],
    queryFn: () => taskFetch<{ tasks: StudioTask[] }>('/v1/studio/tasks'),
    staleTime: 1000 * 30,
  });
}

export function useCreateTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateTaskRequest) =>
      taskFetch<{ task: StudioTask }>('/v1/studio/tasks', {
        method: 'POST',
        body: JSON.stringify(data),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['studio', 'tasks'] });
      toast.success('Task created');
    },
    onError: (error: Error) => {
      toast.error(`Failed to create task: ${error.message}`);
    },
  });
}

export function useUpdateTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { taskId: string; updates: UpdateTaskRequest }) =>
      taskFetch<{ task: StudioTask }>(`/v1/studio/tasks/${params.taskId}`, {
        method: 'PATCH',
        body: JSON.stringify(params.updates),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['studio', 'tasks'] });
      toast.success('Task updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update task: ${error.message}`);
    },
  });
}

export function useDeleteTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (taskId: string) =>
      taskFetch<{ ok: boolean }>(`/v1/studio/tasks/${taskId}`, {
        method: 'DELETE',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['studio', 'tasks'] });
      toast.success('Task deleted');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete task: ${error.message}`);
    },
  });
}

export function useAssignTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { taskId: string; assigneeId: string }) =>
      taskFetch<{ ok: boolean }>(`/v1/studio/tasks/${params.taskId}/assign`, {
        method: 'POST',
        body: JSON.stringify({ assignee_id: params.assigneeId }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['studio', 'tasks'] });
      toast.success('Task assigned');
    },
    onError: (error: Error) => {
      toast.error(`Failed to assign task: ${error.message}`);
    },
  });
}