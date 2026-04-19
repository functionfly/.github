import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { getAnalyticsSettings, updateAnalyticsSettings, type UpdateAnalyticsRequest } from '@/api/analytics';

// Query keys
export const analyticsKeys = {
  all: ['analytics'] as const,
  settings: () => [...analyticsKeys.all, 'settings'] as const,
};

// Get analytics settings
export function useAnalyticsSettings() {
  return useQuery({
    queryKey: analyticsKeys.settings(),
    queryFn: getAnalyticsSettings,
    staleTime: 1000 * 60 * 5,
  });
}

// Update analytics settings
export function useUpdateAnalyticsSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (settings: UpdateAnalyticsRequest) => updateAnalyticsSettings(settings),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: analyticsKeys.settings() });
      toast.success('Analytics settings updated');
    },
    onError: (error: Error) => {
      toast.error(`Failed to update analytics settings: ${error.message}`);
    },
  });
}
