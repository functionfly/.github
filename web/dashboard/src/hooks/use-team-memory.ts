import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { teamMemoryService, type TeamMemory, type CreateMemoryRequest } from '@/services/team-memory.service';
import { useToast } from '@/components/ui/use-toast';

const MEMORIES_QUERY_KEY = 'team-memories';

export function useTeamMemories(teamId: string, filters?: Parameters<typeof teamMemoryService.listMemories>[1]) {
  const { toast } = useToast();

  const query = useQuery({
    queryKey: [MEMORIES_QUERY_KEY, teamId, filters],
    queryFn: () => teamMemoryService.listMemories(teamId, filters),
    enabled: !!teamId,
  });

  return {
    ...query,
    memories: query.data?.memories ?? [],
    total: query.data?.total ?? 0,
  };
}

export function useSearchMemories(teamId: string) {
  const queryClient = useQueryClient();

  const searchMutation = useMutation({
    mutationFn: (params: {
      query: string;
      options?: Parameters<typeof teamMemoryService.searchMemories>[2];
    }) => teamMemoryService.searchMemories(teamId, params.query, params.options),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [MEMORIES_QUERY_KEY, teamId] });
    },
  });

  return searchMutation;
}

export function useCreateMemory(teamId: string) {
  const queryClient = useQueryClient();
  const { toast } = useToast();

  return useMutation({
    mutationFn: (data: CreateMemoryRequest) =>
      teamMemoryService.createMemory(teamId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [MEMORIES_QUERY_KEY, teamId] });
      toast({
        title: 'Memory created',
        description: 'Team memory has been added successfully.',
      });
    },
    onError: (error: Error) => {
      toast({
        title: 'Failed to create memory',
        description: error.message,
        variant: 'destructive',
      });
    },
  });
}

export function useUpdateMemory(teamId: string) {
  const queryClient = useQueryClient();
  const { toast } = useToast();

  return useMutation({
    mutationFn: ({
      memoryId,
      data,
    }: {
      memoryId: string;
      data: Partial<CreateMemoryRequest>;
    }) => teamMemoryService.updateMemory(teamId, memoryId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [MEMORIES_QUERY_KEY, teamId] });
      toast({
        title: 'Memory updated',
        description: 'Changes have been saved.',
      });
    },
    onError: (error: Error) => {
      toast({
        title: 'Failed to update memory',
        description: error.message,
        variant: 'destructive',
      });
    },
  });
}

export function useDeleteMemory(teamId: string) {
  const queryClient = useQueryClient();
  const { toast } = useToast();

  return useMutation({
    mutationFn: (memoryId: string) => teamMemoryService.deleteMemory(teamId, memoryId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [MEMORIES_QUERY_KEY, teamId] });
      toast({
        title: 'Memory deleted',
        description: 'Team memory has been removed.',
      });
    },
    onError: (error: Error) => {
      toast({
        title: 'Failed to delete memory',
        description: error.message,
        variant: 'destructive',
      });
    },
  });
}

export function useValidateMemory(teamId: string) {
  const queryClient = useQueryClient();
  const { toast } = useToast();

  return useMutation({
    mutationFn: ({
      memoryId,
      validated,
    }: {
      memoryId: string;
      validated: boolean;
    }) => teamMemoryService.validateMemory(teamId, memoryId, validated),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: [MEMORIES_QUERY_KEY, teamId] });
      toast({
        title: variables.validated ? 'Memory validated' : 'Memory unvalidated',
        description: variables.validated
          ? 'This memory is now marked as validated.'
          : 'This memory is now marked as unvalidated.',
      });
    },
    onError: (error: Error) => {
      toast({
        title: 'Failed to update validation status',
        description: error.message,
        variant: 'destructive',
      });
    },
  });
}

export function useMemoryExtractions(teamId: string, status: string = 'pending') {
  return useQuery({
    queryKey: [MEMORIES_QUERY_KEY, 'extractions', teamId, status],
    queryFn: () => teamMemoryService.listExtractions(teamId, status),
    enabled: !!teamId,
  });
}

export function useApproveExtraction(teamId: string) {
  const queryClient = useQueryClient();
  const { toast } = useToast();

  return useMutation({
    mutationFn: (extractionId: string) =>
      teamMemoryService.approveExtraction(teamId, extractionId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [MEMORIES_QUERY_KEY, 'extractions', teamId],
      });
      queryClient.invalidateQueries({ queryKey: [MEMORIES_QUERY_KEY, teamId] });
      toast({
        title: 'Extraction approved',
        description: 'Memory has been added to team knowledge.',
      });
    },
    onError: (error: Error) => {
      toast({
        title: 'Failed to approve extraction',
        description: error.message,
        variant: 'destructive',
      });
    },
  });
}

export function useRejectExtraction(teamId: string) {
  const queryClient = useQueryClient();
  const { toast } = useToast();

  return useMutation({
    mutationFn: ({
      extractionId,
      reason,
    }: {
      extractionId: string;
      reason?: string;
    }) => teamMemoryService.rejectExtraction(teamId, extractionId, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: [MEMORIES_QUERY_KEY, 'extractions', teamId],
      });
      toast({
        title: 'Extraction rejected',
        description: 'This extraction has been rejected.',
      });
    },
    onError: (error: Error) => {
      toast({
        title: 'Failed to reject extraction',
        description: error.message,
        variant: 'destructive',
      });
    },
  });
}
