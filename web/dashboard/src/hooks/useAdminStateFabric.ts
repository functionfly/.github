import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { adminStateFabricApi } from "@/api/stateFabric";

// Query keys
export const adminStateFabricKeys = {
  all: ["admin", "state-fabrics"] as const,
  lists: () => [...adminStateFabricKeys.all, "list"] as const,
  list: (filters: string) => [...adminStateFabricKeys.lists(), { filters }] as const,
  stats: () => [...adminStateFabricKeys.all, "stats"] as const,
};

// List all state fabrics (admin)
export function useAdminStateFabrics(params?: {
  tenantId?: string;
  status?: string;
  limit?: number;
  offset?: number;
}) {
  return useQuery({
    queryKey: adminStateFabricKeys.lists(),
    queryFn: () => adminStateFabricApi.listAll(params),
  });
}

// Get state fabric stats (admin)
export function useAdminStateFabricStats() {
  return useQuery({
    queryKey: adminStateFabricKeys.stats(),
    queryFn: () => adminStateFabricApi.getStats(),
    refetchInterval: 60000, // Refetch every minute
  });
}

// Suspend a state fabric
export function useSuspendStateFabric() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ fabricId, reason }: { fabricId: string; reason: string }) =>
      adminStateFabricApi.suspendFabric(fabricId, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: adminStateFabricKeys.lists() });
      toast.success("State fabric suspended");
    },
    onError: (error: Error) => {
      toast.error(`Failed to suspend state fabric: ${error.message}`);
    },
  });
}

// Resume a state fabric
export function useResumeStateFabric() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (fabricId: string) => adminStateFabricApi.resumeFabric(fabricId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: adminStateFabricKeys.lists() });
      toast.success("State fabric resumed");
    },
    onError: (error: Error) => {
      toast.error(`Failed to resume state fabric: ${error.message}`);
    },
  });
}
