import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { studioCollabApi, type CreateCollabEventRequest, type CreateActivityRequest } from '@/api/studioCollab';

export function useResolveComment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { commentId: string; resolved: boolean }) =>
      studioCollabApi.updateEvent(params.commentId, {
        metadata: { resolved: params.resolved },
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['studio', 'collab', 'events'] });
      toast.success('Comment resolved');
    },
    onError: (error: Error) => {
      toast.error(`Failed to resolve comment: ${error.message}`);
    },
  });
}

export function useCreateComment() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { content: string; line?: number; targetId?: string }) =>
      studioCollabApi.createEvent({
        event_type: 'comment',
        metadata: {
          content: data.content,
          line: data.line,
          target_id: data.targetId,
          user_name: 'You',
          user_color: '#f97316',
        },
      } as CreateCollabEventRequest),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['studio', 'collab', 'events'] });
      toast.success('Comment added');
    },
    onError: (error: Error) => {
      toast.error(`Failed to add comment: ${error.message}`);
    },
  });
}

export function useCreateAnnotation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { content: string; targetId: string; targetType: string; position: { x: number; y: number } }) =>
      studioCollabApi.createEvent({
        event_type: 'annotation',
        metadata: {
          content: data.content,
          target_id: data.targetId,
          target_type: data.targetType,
          position: data.position,
          user_name: 'You',
          user_color: '#f97316',
        },
      } as CreateCollabEventRequest),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['studio', 'collab', 'events'] });
      toast.success('Annotation added');
    },
    onError: (error: Error) => {
      toast.error(`Failed to add annotation: ${error.message}`);
    },
  });
}

export function useResolveAnnotation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { annotationId: string; resolved: boolean }) =>
      studioCollabApi.updateEvent(params.annotationId, {
        metadata: { resolved: params.resolved },
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['studio', 'collab', 'events'] });
      toast.success('Annotation resolved');
    },
    onError: (error: Error) => {
      toast.error(`Failed to resolve annotation: ${error.message}`);
    },
  });
}

export function useStartPairSession() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data?: { guestName?: string }) =>
      studioCollabApi.createEvent({
        event_type: 'pair_session',
        metadata: {
          host_name: 'You',
          host_color: '#f97316',
          guest_name: data?.guestName,
          status: 'active',
          started_at: new Date().toISOString(),
        },
      } as CreateCollabEventRequest),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['studio', 'collab', 'events'] });
      toast.success('Pair session started');
    },
    onError: (error: Error) => {
      toast.error(`Failed to start pair session: ${error.message}`);
    },
  });
}

export function useEndPairSession() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (sessionId: string) =>
      studioCollabApi.updateEvent(sessionId, {
        metadata: { status: 'ended', ended_at: new Date().toISOString() },
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['studio', 'collab', 'events'] });
      toast.success('Pair session ended');
    },
    onError: (error: Error) => {
      toast.error(`Failed to end pair session: ${error.message}`);
    },
  });
}

export function useUpdatePromptVersion() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { prompt: string; changes?: string }) =>
      studioCollabApi.createEvent({
        event_type: 'prompt_version',
        metadata: {
          prompt: data.prompt,
          changes: data.changes || 'Update',
          user_name: 'You',
          user_color: '#f97316',
        },
      } as CreateCollabEventRequest),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['studio', 'collab', 'events'] });
      toast.success('Prompt version saved');
    },
    onError: (error: Error) => {
      toast.error(`Failed to save prompt version: ${error.message}`);
    },
  });
}

export function useRecordGraphEdit() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: { nodeId: string; field: string; oldValue: string; newValue: string }) =>
      studioCollabApi.createEvent({
        event_type: 'graph_edit',
        metadata: {
          node_id: data.nodeId,
          field: data.field,
          old_value: data.oldValue,
          new_value: data.newValue,
          user_name: 'You',
          user_color: '#f97316',
        },
      } as CreateCollabEventRequest),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['studio', 'collab', 'events'] });
    },
    onError: (error: Error) => {
      toast.error(`Failed to record graph edit: ${error.message}`);
    },
  });
}

export function useResolveConflict() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { conflictId: string; resolution: 'current' | 'incoming' }) =>
      studioCollabApi.updateEvent(params.conflictId, {
        metadata: { resolution: params.resolution, resolved_at: new Date().toISOString() },
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['studio', 'collab', 'events'] });
      toast.success('Conflict resolved');
    },
    onError: (error: Error) => {
      toast.error(`Failed to resolve conflict: ${error.message}`);
    },
  });
}

export function useRecordActivity() {
  return useMutation({
    mutationFn: (data: CreateActivityRequest) => studioCollabApi.createActivity(data),
  });
}