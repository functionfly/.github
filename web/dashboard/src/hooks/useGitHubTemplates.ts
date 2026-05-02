import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { githubApi } from '@/api/github';
import { githubKeys } from '@/hooks/useGitHubConnection';
import type { CreateTemplateRequest, UpdateTemplateRequest } from '@/types/github';

export function useGitHubTemplates() {
  return useQuery({
    queryKey: githubKeys.templates(),
    queryFn: () => githubApi.listTemplates(),
    staleTime: 1000 * 60 * 5,
  });
}

export function useCreateTemplate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateTemplateRequest) => githubApi.createTemplate(data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: githubKeys.templates() });
      toast.success(`Template "${data.name}" created`);
    },
    onError: (error: Error) => {
      toast.error(`Failed to create template: ${error.message}`);
    },
  });
}

export function useUpdateTemplate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateTemplateRequest }) =>
      githubApi.updateTemplate(id, data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: githubKeys.templates() });
      toast.success(`Template "${data.name}" updated`);
    },
    onError: (error: Error) => {
      toast.error(`Failed to update template: ${error.message}`);
    },
  });
}

export function useDeleteTemplate() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (templateId: string) => githubApi.deleteTemplate(templateId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: githubKeys.templates() });
      toast.success('Template deleted');
    },
    onError: (error: Error) => {
      toast.error(`Failed to delete template: ${error.message}`);
    },
  });
}
