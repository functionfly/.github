import { useState, useEffect, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import {
  Shield,
  AlertTriangle,
  AlertCircle,
  CheckCircle,
  XCircle,
  RefreshCw,
  Filter,
  Eye,
  EyeOff,
  Power,
  PowerOff,
  Activity,
  X,
} from 'lucide-react';
import { agentApi } from '@/api/agent';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

interface SecurityAlert {
  id: string;
  type: string;
  severity: 'critical' | 'high' | 'medium' | 'low';
  description: string;
  agentId?: string;
  agentName?: string;
  timestamp: string;
  status: 'active' | 'acknowledged' | 'resolved';
  metadata?: Record<string, unknown>;
}

interface KillSwitchConfirmation {
  alert: SecurityAlert | null;
  open: boolean;
}

const SEVERITY_CONFIG = {
  critical: {
    color: 'error' as const,
    icon: XCircle,
    label: 'Critical',
  },
  high: {
    color: 'destructive' as const,
    icon: AlertTriangle,
    label: 'High',
  },
  medium: {
    color: 'warning' as const,
    icon: AlertCircle,
    label: 'Medium',
  },
  low: {
    color: 'secondary' as const,
    icon: Activity,
    label: 'Low',
  },
};

const STATUS_CONFIG = {
  active: {
    color: 'error' as const,
    label: 'Active',
  },
  acknowledged: {
    color: 'warning' as const,
    label: 'Acknowledged',
  },
  resolved: {
    color: 'success' as const,
    label: 'Resolved',
  },
};

const generateMockAlerts = (): SecurityAlert[] => {
  const now = new Date();
  return [
    {
      id: 'alert-001',
      type: 'anomaly_detection',
      severity: 'critical',
      description: 'Unusual API call pattern detected from agent "worker-alpha-7". Multiple failed authentication attempts followed by successful admin actions.',
      agentId: 'agent-123',
      agentName: 'worker-alpha-7',
      timestamp: new Date(now.getTime() - 5 * 60000).toISOString(),
      status: 'active',
    },
    {
      id: 'alert-002',
      type: 'permission_escalation',
      severity: 'high',
      description: 'Agent "data-processor-3" attempted to escalate privileges beyond configured limits.',
      agentId: 'agent-456',
      agentName: 'data-processor-3',
      timestamp: new Date(now.getTime() - 15 * 60000).toISOString(),
      status: 'active',
    },
    {
      id: 'alert-003',
      type: 'resource_abuse',
      severity: 'medium',
      description: 'Agent "batch-runner-2" exceeded memory allocation by 200% for extended period.',
      agentId: 'agent-789',
      agentName: 'batch-runner-2',
      timestamp: new Date(now.getTime() - 45 * 60000).toISOString(),
      status: 'acknowledged',
    },
    {
      id: 'alert-004',
      type: 'unusual_behavior',
      severity: 'low',
      description: 'Agent "listener-bot-1" has been idle for 48 hours without proper shutdown sequence.',
      agentId: 'agent-101',
      agentName: 'listener-bot-1',
      timestamp: new Date(now.getTime() - 120 * 60000).toISOString(),
      status: 'resolved',
    },
    {
      id: 'alert-005',
      type: 'network_anomaly',
      severity: 'high',
      description: 'Suspicious outbound traffic pattern detected from agent swarm. Potential data exfiltration attempt.',
      agentId: 'agent-202',
      agentName: 'swarm-coordinator',
      timestamp: new Date(now.getTime() - 2 * 60000).toISOString(),
      status: 'active',
    },
    {
      id: 'alert-006',
      type: 'quota_violation',
      severity: 'medium',
      description: 'Agent "compute-optimizer" exceeded hourly API call quota by 150%.',
      agentId: 'agent-303',
      agentName: 'compute-optimizer',
      timestamp: new Date(now.getTime() - 30 * 60000).toISOString(),
      status: 'acknowledged',
    },
  ];
};

export function SecurityAlertsPanel({ className }: { className?: string }) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  const [alerts, setAlerts] = useState<SecurityAlert[]>([]);
  const [selectedAlert, setSelectedAlert] = useState<SecurityAlert | null>(null);
  const [severityFilter, setSeverityFilter] = useState<string>('all');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [autoRefreshInterval, setAutoRefreshInterval] = useState(30);
  const [killSwitchConfirm, setKillSwitchConfirm] = useState<KillSwitchConfirmation>({
    alert: null,
    open: false,
  });
  const [isDetailsPanelOpen, setIsDetailsPanelOpen] = useState(false);

  const { data: agentsData } = useQuery({
    queryKey: ['agents'],
    queryFn: async () => {
      const response = await agentApi.listAgents({ limit: 100 });
      return response.agents;
    },
    staleTime: 30000,
  });

  const killSwitchMutation = useMutation({
    mutationFn: async ({ agentId, reason }: { agentId: string; reason?: string }) => {
      const response = await agentApi.triggerKillSwitch(agentId, reason);
      return response;
    },
    onSuccess: (data, variables) => {
      toast.success(`Kill switch activated`, {
        description: `Agent ${variables.agentId} has been terminated. ${data.agents_killed} related processes stopped.`,
      });
      setKillSwitchConfirm({ alert: null, open: false });
      queryClient.invalidateQueries({ queryKey: ['agents'] });
    },
    onError: (error) => {
      toast.error('Failed to activate kill switch', {
        description: error instanceof Error ? error.message : 'Unknown error occurred',
      });
    },
  });

  const acknowledgeMutation = useMutation({
    mutationFn: async (alertId: string) => {
      setAlerts((prev) =>
        prev.map((alert) =>
          alert.id === alertId ? { ...alert, status: 'acknowledged' as const } : alert
        )
      );
      return { success: true };
    },
    onSuccess: () => {
      toast.success('Alert acknowledged');
    },
  });

  const resolveMutation = useMutation({
    mutationFn: async (alertId: string) => {
      setAlerts((prev) =>
        prev.map((alert) =>
          alert.id === alertId ? { ...alert, status: 'resolved' as const } : alert
        )
      );
      return { success: true };
    },
    onSuccess: () => {
      toast.success('Alert resolved');
    },
  });

  useEffect(() => {
    const initialAlerts = generateMockAlerts();
    setAlerts(initialAlerts);
  }, []);

  useEffect(() => {
    if (!autoRefresh) return;

    const interval = setInterval(() => {
      setAlerts((prev) => {
        const newAlerts = [...prev];
        if (Math.random() > 0.7 && newAlerts.length < 20) {
          const severities: ('critical' | 'high' | 'medium' | 'low')[] = ['critical', 'high', 'medium', 'low'];
          const types = ['anomaly_detection', 'permission_escalation', 'resource_abuse', 'network_anomaly'];
          const newAlert: SecurityAlert = {
            id: `alert-${Date.now()}`,
            type: types[Math.floor(Math.random() * types.length)],
            severity: severities[Math.floor(Math.random() * severities.length)],
            description: 'New security anomaly detected during auto-monitoring.',
            agentId: `agent-${Math.floor(Math.random() * 1000)}`,
            agentName: `agent-${Math.floor(Math.random() * 100)}`,
            timestamp: new Date().toISOString(),
            status: 'active',
          };
          newAlerts.unshift(newAlert);
        }
        return newAlerts.slice(0, 50);
      });
    }, autoRefreshInterval * 1000);

    return () => clearInterval(interval);
  }, [autoRefresh, autoRefreshInterval]);

  const filteredAlerts = alerts.filter((alert) => {
    if (severityFilter !== 'all' && alert.severity !== severityFilter) return false;
    if (statusFilter !== 'all' && alert.status !== statusFilter) return false;
    return true;
  });

  const handleKillSwitch = useCallback((alert: SecurityAlert) => {
    setKillSwitchConfirm({ alert, open: true });
  }, []);

  const confirmKillSwitch = useCallback(() => {
    if (killSwitchConfirm.alert?.agentId) {
      killSwitchMutation.mutate({
        agentId: killSwitchConfirm.alert.agentId,
        reason: `Security alert: ${killSwitchConfirm.alert.type}`,
      });
    }
  }, [killSwitchConfirm.alert, killSwitchMutation]);

  const handleAlertClick = useCallback((alert: SecurityAlert) => {
    setSelectedAlert(alert);
    setIsDetailsPanelOpen(true);
  }, []);

  const handleCloseDetails = useCallback(() => {
    setIsDetailsPanelOpen(false);
    setTimeout(() => setSelectedAlert(null), 200);
  }, []);

  const activeAlertCount = alerts.filter((a) => a.status === 'active').length;
  const criticalAlertCount = alerts.filter((a) => a.severity === 'critical' && a.status === 'active').length;

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffMins < 1440) return `${Math.floor(diffMins / 60)}h ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className={className}>
      <Card className="border-orange-500/20">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-orange-500 to-red-500">
                <Shield className="h-5 w-5 text-white" />
              </div>
              <div>
                <CardTitle className="flex items-center gap-2">
                  Security Alerts
                  {criticalAlertCount > 0 && (
                    <Badge variant="error" className="animate-pulse">
                      {criticalAlertCount} Critical
                    </Badge>
                  )}
                </CardTitle>
                <CardDescription>
                  {activeAlertCount} active alerts · Auto-refresh every {autoRefreshInterval}s
                </CardDescription>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setAutoRefresh(!autoRefresh)}
                title={autoRefresh ? 'Pause auto-refresh' : 'Enable auto-refresh'}
              >
                <RefreshCw className={`h-4 w-4 ${autoRefresh ? 'animate-spin' : ''}`} />
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-wrap items-center gap-3">
            <div className="flex items-center gap-2">
              <Filter className="h-4 w-4 text-text-muted" />
              <Select value={severityFilter} onValueChange={setSeverityFilter}>
                <SelectTrigger className="w-[140px]">
                  <SelectValue placeholder="Severity" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Severities</SelectItem>
                  <SelectItem value="critical">Critical</SelectItem>
                  <SelectItem value="high">High</SelectItem>
                  <SelectItem value="medium">Medium</SelectItem>
                  <SelectItem value="low">Low</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-[140px]">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Status</SelectItem>
                <SelectItem value="active">Active</SelectItem>
                <SelectItem value="acknowledged">Acknowledged</SelectItem>
                <SelectItem value="resolved">Resolved</SelectItem>
              </SelectContent>
            </Select>
            <div className="ml-auto flex items-center gap-2 text-sm text-text-muted">
              <span>{filteredAlerts.length} alerts</span>
            </div>
          </div>

          {criticalAlertCount > 0 && (
            <Alert variant="destructive" className="border-red-500/50">
              <AlertTriangle className="h-4 w-4" />
              <AlertDescription>
                <strong>{criticalAlertCount} critical security alert{criticalAlertCount > 1 ? 's' : ''}</strong> require immediate attention.
                {criticalAlertCount > 0 && ' Consider activating kill switches for compromised agents.'}
              </AlertDescription>
            </Alert>
          )}

          <div className="space-y-2 max-h-[500px] overflow-y-auto">
            {filteredAlerts.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-12 text-center">
                <CheckCircle className="h-12 w-12 text-success mb-3" />
                <p className="text-text-muted">No security alerts match your filters</p>
              </div>
            ) : (
              filteredAlerts.map((alert) => {
                const severityConfig = SEVERITY_CONFIG[alert.severity];
                const statusConfig = STATUS_CONFIG[alert.status];
                const SeverityIcon = severityConfig.icon;

                return (
                  <div
                    key={alert.id}
                    className="group relative rounded-lg border border-border-default bg-card p-4 transition-all duration-200 hover:border-border-focus hover:shadow-md cursor-pointer"
                    onClick={() => handleAlertClick(alert)}
                  >
                    <div className="flex items-start gap-3">
                      <div className={`mt-0.5 rounded-full p-1.5 ${
                        alert.severity === 'critical' ? 'bg-error/10 text-error' :
                        alert.severity === 'high' ? 'bg-red-500/10 text-red-500' :
                        alert.severity === 'medium' ? 'bg-amber-500/10 text-amber-500' :
                        'bg-blue-500/10 text-blue-500'
                      }`}>
                        <SeverityIcon className="h-4 w-4" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 flex-wrap">
                          <span className="font-medium text-text-primary">
                            {alert.type.replace(/_/g, ' ')}
                          </span>
                          <Badge variant={statusConfig.color} className="text-xs">
                            {statusConfig.label}
                          </Badge>
                          <Badge variant={severityConfig.color} className="text-xs">
                            {severityConfig.label}
                          </Badge>
                        </div>
                        <p className="mt-1 text-sm text-text-secondary line-clamp-2">
                          {alert.description}
                        </p>
                        <div className="mt-2 flex items-center gap-3 text-xs text-text-muted">
                          {alert.agentName && (
                            <span className="flex items-center gap-1">
                              <Power className="h-3 w-3" />
                              {alert.agentName}
                            </span>
                          )}
                          <span>{formatTimestamp(alert.timestamp)}</span>
                        </div>
                      </div>
                    </div>
                    {alert.status !== 'resolved' && (
                      <div className="absolute right-3 top-3 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                        {alert.status === 'active' && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={(e) => {
                              e.stopPropagation();
                              acknowledgeMutation.mutate(alert.id);
                            }}
                          >
                            <Eye className="h-3 w-3 mr-1" />
                            Ack
                          </Button>
                        )}
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={(e) => {
                            e.stopPropagation();
                            resolveMutation.mutate(alert.id);
                          }}
                        >
                          <CheckCircle className="h-3 w-3 mr-1" />
                          Resolve
                        </Button>
                        {alert.agentId && alert.status === 'active' && (
                          <Button
                            variant="destructive"
                            size="sm"
                            onClick={(e) => {
                              e.stopPropagation();
                              handleKillSwitch(alert);
                            }}
                          >
                            <PowerOff className="h-3 w-3 mr-1" />
                            Kill
                          </Button>
                        )}
                      </div>
                    )}
                  </div>
                );
              })
            )}
          </div>
        </CardContent>
      </Card>

      <Dialog open={killSwitchConfirm.open} onOpenChange={(open) => setKillSwitchConfirm({ alert: null, open })}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-error">
              <PowerOff className="h-5 w-5" />
              Activate Kill Switch
            </DialogTitle>
            <DialogDescription>
              This will immediately terminate agent "{killSwitchConfirm.alert?.agentName}" 
              and all related processes. This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <Alert variant="destructive">
              <AlertTriangle className="h-4 w-4" />
              <AlertDescription>
                You are about to kill switch on alert: <strong>{killSwitchConfirm.alert?.type}</strong>
                {killSwitchConfirm.alert?.agentId && (
                  <> (Agent ID: {killSwitchConfirm.alert.agentId})</>
                )}
              </AlertDescription>
            </Alert>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setKillSwitchConfirm({ alert: null, open: false })}
              disabled={killSwitchMutation.isPending}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              isLoading={killSwitchMutation.isPending}
              onClick={confirmKillSwitch}
            >
              <PowerOff className="h-4 w-4 mr-2" />
              Confirm Kill Switch
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={isDetailsPanelOpen} onOpenChange={(open) => !open && handleCloseDetails()}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <div className="flex items-center justify-between">
              <DialogTitle className="flex items-center gap-2">
                <Shield className="h-5 w-5 text-orange-500" />
                Alert Details
              </DialogTitle>
              <Button variant="ghost" size="icon" onClick={handleCloseDetails}>
                <X className="h-4 w-4" />
              </Button>
            </div>
          </DialogHeader>
          {selectedAlert && (
            <div className="space-y-4 py-4">
              <div className="flex items-center gap-3">
                {(() => {
                  const config = SEVERITY_CONFIG[selectedAlert.severity];
                  const Icon = config.icon;
                  return (
                    <div className={`rounded-full p-2 ${
                      selectedAlert.severity === 'critical' ? 'bg-error/10 text-error' :
                      selectedAlert.severity === 'high' ? 'bg-red-500/10 text-red-500' :
                      selectedAlert.severity === 'medium' ? 'bg-amber-500/10 text-amber-500' :
                      'bg-blue-500/10 text-blue-500'
                    }`}>
                      <Icon className="h-5 w-5" />
                    </div>
                  );
                })()}
                <div>
                  <h3 className="font-semibold text-lg capitalize">
                    {selectedAlert.type.replace(/_/g, ' ')}
                  </h3>
                  <div className="flex items-center gap-2 mt-1">
                    <Badge variant={SEVERITY_CONFIG[selectedAlert.severity].color}>
                      {SEVERITY_CONFIG[selectedAlert.severity].label}
                    </Badge>
                    <Badge variant={STATUS_CONFIG[selectedAlert.status].color}>
                      {STATUS_CONFIG[selectedAlert.status].label}
                    </Badge>
                  </div>
                </div>
              </div>

              <div className="rounded-lg bg-bg-tertiary p-4">
                <h4 className="text-sm font-medium text-text-muted mb-2">Description</h4>
                <p className="text-text-primary">{selectedAlert.description}</p>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="rounded-lg bg-bg-tertiary p-4">
                  <h4 className="text-sm font-medium text-text-muted mb-1">Agent</h4>
                  <p className="text-text-primary font-mono text-sm">
                    {selectedAlert.agentName || 'Unknown'}
                  </p>
                  {selectedAlert.agentId && (
                    <p className="text-xs text-text-muted mt-1">
                      ID: {selectedAlert.agentId}
                    </p>
                  )}
                </div>
                <div className="rounded-lg bg-bg-tertiary p-4">
                  <h4 className="text-sm font-medium text-text-muted mb-1">Timestamp</h4>
                  <p className="text-text-primary">
                    {new Date(selectedAlert.timestamp).toLocaleString()}
                  </p>
                  <p className="text-xs text-text-muted mt-1">
                    {formatTimestamp(selectedAlert.timestamp)}
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-3 pt-4 border-t border-border-default">
                {selectedAlert.status === 'active' && (
                  <>
                    <Button
                      variant="outline"
                      onClick={() => {
                        acknowledgeMutation.mutate(selectedAlert.id);
                        handleCloseDetails();
                      }}
                    >
                      <Eye className="h-4 w-4 mr-2" />
                      Acknowledge
                    </Button>
                    <Button
                      variant="outline"
                      onClick={() => {
                        resolveMutation.mutate(selectedAlert.id);
                        handleCloseDetails();
                      }}
                    >
                      <CheckCircle className="h-4 w-4 mr-2" />
                      Resolve
                    </Button>
                  </>
                )}
                {selectedAlert.agentId && selectedAlert.status === 'active' && (
                  <Button
                    variant="destructive"
                    className="ml-auto"
                    onClick={() => {
                      handleCloseDetails();
                      setTimeout(() => handleKillSwitch(selectedAlert), 300);
                    }}
                  >
                    <PowerOff className="h-4 w-4 mr-2" />
                    Activate Kill Switch
                  </Button>
                )}
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
