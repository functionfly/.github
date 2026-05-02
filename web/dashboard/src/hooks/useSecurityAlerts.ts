import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { agentApi } from "@/api/agent";

// Query keys for security alerts
export const securityAlertKeys = {
  all: ["securityAlerts"] as const,
  list: (filters?: { severity?: string; status?: string }) => [...securityAlertKeys.all, "list", filters] as const,
  detail: (id: string) => [...securityAlertKeys.all, "detail", id] as const,
};

/**
 * Get security alerts for an agent
 * GET /v1/agent/{agent_id}/security/alerts
 */
export function useSecurityAlerts(agentId: string, filters?: { severity?: string; status?: string }) {
  return useQuery({
    queryKey: securityAlertKeys.list(filters),
    queryFn: () => agentApi.getSecurityAlerts(agentId, filters),
    enabled: !!agentId,
    staleTime: 1000 * 30, // 30 seconds
  });
}

/**
 * Acknowledge a security alert
 * POST /v1/agent/{agent_id}/security/alerts/{alert_id}/acknowledge
 */
export function useAcknowledgeAlert() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { agentId: string; alertId: string }) =>
      agentApi.acknowledgeSecurityAlert(params.agentId, params.alertId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: securityAlertKeys.all });
      toast.success("Alert acknowledged");
    },
    onError: (error: Error) => {
      toast.error(`Failed to acknowledge alert: ${error.message}`);
    },
  });
}

/**
 * Resolve a security alert
 * POST /v1/agent/{agent_id}/security/alerts/{alert_id}/resolve
 */
export function useResolveAlert() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { agentId: string; alertId: string; resolution?: string }) =>
      agentApi.resolveSecurityAlert(params.agentId, params.alertId, params.resolution),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: securityAlertKeys.all });
      toast.success("Alert resolved");
    },
    onError: (error: Error) => {
      toast.error(`Failed to resolve alert: ${error.message}`);
    },
  });
}

/**
 * Trigger kill switch on an agent
 * POST /v1/agent/{agent_id}/kill-switch
 */
export function useKillSwitch() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (params: { agentId: string; reason?: string }) =>
      agentApi.triggerKillSwitch(params.agentId, params.reason),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: securityAlertKeys.all });
      queryClient.invalidateQueries({ queryKey: ["agents"] });
      toast.success("Kill switch activated", {
        description: `Agent ${variables.agentId} has been stopped`,
      });
    },
    onError: (error: Error) => {
      toast.error(`Failed to trigger kill switch: ${error.message}`);
    },
  });
}
