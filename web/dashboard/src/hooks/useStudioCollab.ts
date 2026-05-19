import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { studioCollabApi, type CollabEvent, type CreateCollabEventRequest, type CreateActivityRequest } from '@/api/studioCollab';

const COLLAB_KEY = 'studio-collab';

export function useStudioCollabEvents(filters?: { type?: string; limit?: number; offset?: number }) {
  return useQuery({
    queryKey: [COLLAB_KEY, 'events', filters],
    queryFn: () => studioCollabApi.listEvents(filters),
    staleTime: 1000 * 10,
  });
}

export function useStudioCollabActivity(limit?: number) {
  return useQuery({
    queryKey: [COLLAB_KEY, 'activity', limit],
    queryFn: () => studioCollabApi.getActivityFeed(limit),
    staleTime: 1000 * 10,
    refetchInterval: 15000,
  });
}

export function useCreateCollabEvent() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateCollabEventRequest) => studioCollabApi.createEvent(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [COLLAB_KEY] });
    },
  });
}

export function useCreateCollabActivity() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateActivityRequest) => studioCollabApi.createActivity(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [COLLAB_KEY, 'activity'] });
    },
  });
}

export function useUpdateCollabEvent() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CollabEvent['metadata']> }) =>
      studioCollabApi.updateEvent(id, { metadata: data }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [COLLAB_KEY] });
    },
  });
}

export function useDeleteCollabEvent() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => studioCollabApi.deleteEvent(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [COLLAB_KEY] });
    },
  });
}
