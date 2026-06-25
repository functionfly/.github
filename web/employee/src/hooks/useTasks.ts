import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { tasksApi, type Task } from '@/api/projects';

export function useTasks(params?: { project_id?: string; assignee_id?: string; status?: string; limit?: number; offset?: number }) {
  return useQuery({
    queryKey: ['tasks', params],
    queryFn: () => tasksApi.list(params),
  });
}

export function useTask(id: string) {
  return useQuery({
    queryKey: ['task', id],
    queryFn: () => tasksApi.get(id),
    enabled: !!id,
  });
}

export function useCreateTask() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Task>) => tasksApi.create(data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['tasks'] }),
  });
}

export function useUpdateTask() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Task> }) =>
      tasksApi.update(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      queryClient.invalidateQueries({ queryKey: ['task', id] });
    },
  });
}

export function useAssignTask() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, assigneeId }: { id: string; assigneeId: string }) =>
      tasksApi.assign(id, assigneeId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['tasks'] }),
  });
}
