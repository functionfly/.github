import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Separator } from '@/components/ui/separator';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dna,
  Zap,
  Shield,
  GitBranch,
  Server,
  Gauge,
  Bell,
  AlertTriangle,
  CheckCircle2,
  ChevronRight,
  Loader2,
  Clock,
  Info,
} from 'lucide-react';
import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { usersApi } from '@/api/users';
import { Link } from 'react-router-dom';

// ──────────────────────────────────────────────────────────────────────────────
// PlatformSettingsTab — Function DNA, runtime defaults, canary, and security
// ──────────────────────────────────────────────────────────────────────────────

export function PlatformSettingsTab() {
  const queryClient = useQueryClient();

  // Load platform settings
  const { data: settingsData, isLoading } = useQuery({
    queryKey: ['platform-settings'],
    queryFn: async () => {
      try {
        const data = await usersApi.getMySettings();
        return (data as { settings?: Record<string, unknown> }).settings ?? {};
      } catch {
        return {};
      }
    },
    staleTime: 60 * 1000,
  });

  const platformSettings = (settingsData as Record<string, unknown>) ?? {};
  const dna = (platformSettings.dna as Record<string, unknown>) ?? {};

  // Update settings mutation
  const updateSettings = useMutation({
    mutationFn: async (patch: Record<string, unknown>) => {
      const current = (settingsData as Record<string, unknown>) ?? {};
      const updated = {
        ...current,
        dna: { ...((current.dna as Record<string, unknown>) ?? {}), ...patch },
      };
      await usersApi.updateMyPlatformSettings(updated);
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['platform-settings'] });

      // Build a summary of what changed
      const changes: string[] = [];
      if ('auto_evolve' in variables) changes.push(`Auto-Evolution ${variables.auto_evolve ? 'enabled' : 'disabled'}`);
      if ('require_approval' in variables) changes.push(`Manual approval ${variables.require_approval ? 'required' : 'auto-approved'}`);
      if ('sandbox_validation' in variables) changes.push(`Sandbox validation ${variables.sandbox_validation ? 'enabled' : 'disabled'}`);
      if ('default_canary_pct' in variables) changes.push(`Canary default: ${variables.default_canary_pct}%`);
      if ('max_mutations_per_day' in variables) changes.push(`Max mutations: ${variables.max_mutations_per_day}/day`);
      if ('notify_on_proposal' in variables) changes.push(`Proposal notifications ${variables.notify_on_proposal ? 'on' : 'off'}`);
      if ('notify_on_deploy' in variables) changes.push(`Deploy notifications ${variables.notify_on_deploy ? 'on' : 'off'}`);
      if ('notify_on_rollback' in variables) changes.push(`Rollback notifications ${variables.notify_on_rollback ? 'on' : 'off'}`);
      if ('auto_rollback_on_error' in variables) changes.push(`Auto-rollback ${variables.auto_rollback_on_error ? 'enabled' : 'disabled'}`);
      if ('auto_rollback_error_threshold' in variables) changes.push(`Rollback threshold: ${variables.auto_rollback_error_threshold}%`);

      if (changes.length > 0) {
        toast.success('Platform settings saved', {
          description: changes.slice(0, 5).join(' · ') + (changes.length > 5 ? ` · +${changes.length - 5} more` : ''),
          icon: <CheckCircle2 className="h-4 w-4 text-success" />,
          duration: 4000,
        });
      } else {
        toast.success('Platform settings saved');
      }
    },
    onError: (error: Error) => {
      toast.error('Failed to save platform settings', {
        description: error?.message || 'Please try again',
        icon: <AlertTriangle className="h-4 w-4 text-error" />,
      });
    },
  });

  // Local state for optimistic updates
  const [autoEvolve, setAutoEvolve] = useState(
    (dna.auto_evolve as boolean) ?? true
  );
  const [requireApproval, setRequireApproval] = useState(
    (dna.require_approval as boolean) ?? true
  );
  const [defaultCanaryPct, setDefaultCanaryPct] = useState(
    String((dna.default_canary_pct as number) ?? 10)
  );
  const [notifyOnProposal, setNotifyOnProposal] = useState(
    (dna.notify_on_proposal as boolean) ?? true
  );
  const [notifyOnDeploy, setNotifyOnDeploy] = useState(
    (dna.notify_on_deploy as boolean) ?? true
  );
  const [notifyOnRollback, setNotifyOnRollback] = useState(
    (dna.notify_on_rollback as boolean) ?? true
  );
  const [maxMutationsPerDay, setMaxMutationsPerDay] = useState(
    String((dna.max_mutations_per_day as number) ?? 5)
  );
  const [sandboxValidation, setSandboxValidation] = useState(
    (dna.sandbox_validation as boolean) ?? true
  );
  const [autoRollback, setAutoRollback] = useState(
    (dna.auto_rollback_on_error as boolean) ?? true
  );
  const [autoRollbackThreshold, setAutoRollbackThreshold] = useState(
    String((dna.auto_rollback_error_threshold as number) ?? 5)
  );

  const handleSave = () => {
    updateSettings.mutate({
      auto_evolve: autoEvolve,
      require_approval: requireApproval,
      default_canary_pct: Number(defaultCanaryPct),
      notify_on_proposal: notifyOnProposal,
      notify_on_deploy: notifyOnDeploy,
      notify_on_rollback: notifyOnRollback,
      max_mutations_per_day: Number(maxMutationsPerDay),
      sandbox_validation: sandboxValidation,
      auto_rollback_on_error: autoRollback,
      auto_rollback_error_threshold: Number(autoRollbackThreshold),
    });
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-16">
        <Loader2 className="h-6 w-6 animate-spin text-text-muted" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* ─── Function DNA: Evolution ─── */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Dna className="h-5 w-5 text-brand-500" />
            Function DNA
          </CardTitle>
          <CardDescription>
            Control how your functions evolve. Function DNA tracks execution patterns
            and proposes AI-powered code optimizations.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Auto-evolve master switch */}
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label className="text-base flex items-center gap-2">
                <Zap className="h-4 w-4 text-velocity-500" />
                Auto-Evolution
              </Label>
              <p className="text-sm text-text-muted">
                Allow DNA to propose code mutations based on execution patterns
              </p>
            </div>
            <Switch
              checked={autoEvolve}
              onCheckedChange={(val) => {
                setAutoEvolve(val);
                toast.info(val ? 'Auto-evolution enabled' : 'Auto-evolution disabled', {
                  description: val
                    ? 'DNA will analyze functions and propose optimizations'
                    : 'DNA analysis paused — manual trigger still works',
                  icon: <Info className="h-4 w-4 text-info" />,
                  duration: 3000,
                });
              }}
            />
          </div>

          <Separator />

          {/* Require approval */}
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label className="text-base flex items-center gap-2">
                <Shield className="h-4 w-4 text-info" />
                Require Manual Approval
              </Label>
              <p className="text-sm text-text-muted">
                Mutations stay in proposed state until you accept or reject them
              </p>
            </div>
            <Switch
              checked={requireApproval}
              onCheckedChange={(val) => {
                setRequireApproval(val);
                toast.info(
                  val ? 'Manual approval required' : 'Mutations auto-approved',
                  {
                    description: val
                      ? 'All proposed mutations require your review before deployment'
                      : 'Accepted mutations will automatically trigger canary deployment',
                    icon: <Info className="h-4 w-4 text-info" />,
                    duration: 3000,
                  }
                );
              }}
            />
          </div>

          <Separator />

          {/* Sandbox validation */}
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label className="text-base flex items-center gap-2">
                <Shield className="h-4 w-4 text-success" />
                Sandbox Validation
              </Label>
              <p className="text-sm text-text-muted">
                Run mutations in a sandbox before acceptance to verify behavioral equivalence
              </p>
            </div>
            <Switch
              checked={sandboxValidation}
              onCheckedChange={(val) => {
                setSandboxValidation(val);
                toast.info(
                  val ? 'Sandbox validation enabled' : 'Sandbox validation disabled',
                  {
                    description: val
                      ? 'Mutations will be tested in a sandbox before acceptance'
                      : 'Mutations will be accepted without sandbox testing',
                    icon: <Info className="h-4 w-4 text-info" />,
                    duration: 3000,
                  }
                );
              }}
            />
          </div>

          <Separator />

          {/* Default canary percentage */}
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label className="text-base flex items-center gap-2">
                <GitBranch className="h-4 w-4 text-warning" />
                Default Canary Percentage
              </Label>
              <p className="text-sm text-text-muted">
                Percentage of traffic routed to the mutated code during canary deployment
              </p>
            </div>
            <Select value={defaultCanaryPct} onValueChange={(val) => {
              setDefaultCanaryPct(val);
              toast.success(`Canary default set to ${val}%`, {
                description: `${val}% of traffic will route to new mutations during canary`,
                icon: <GitBranch className="h-4 w-4 text-success" />,
                duration: 2500,
              });
            }}>
              <SelectTrigger className="w-24">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {[5, 10, 25, 50, 100].map((v) => (
                  <SelectItem key={v} value={String(v)}>
                    {v}%
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <Separator />

          {/* Max mutations per day */}
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label className="text-base flex items-center gap-2">
                <Clock className="h-4 w-4 text-text-muted" />
                Max Mutations Per Day
              </Label>
              <p className="text-sm text-text-muted">
                Limit how many mutations the AI can propose per function per day
              </p>
            </div>
            <Select value={maxMutationsPerDay} onValueChange={(val) => {
              setMaxMutationsPerDay(val);
              toast.success(`Max mutations set to ${val}/day`, {
                description: `AI will propose at most ${val} mutations per function per day`,
                icon: <Clock className="h-4 w-4 text-success" />,
                duration: 2500,
              });
            }}>
              <SelectTrigger className="w-24">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {[1, 3, 5, 10, 25, 50].map((v) => (
                  <SelectItem key={v} value={String(v)}>
                    {v}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* ─── Auto-Rollback ─── */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-warning" />
            Auto-Rollback
          </CardTitle>
          <CardDescription>
            Automatically revert mutations if the canary deployment shows elevated error rates
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label className="text-base flex items-center gap-2">
                <CheckCircle2 className="h-4 w-4 text-success" />
                Enable Auto-Rollback
              </Label>
              <p className="text-sm text-text-muted">
                Cancel canary and revert to original code when error rate exceeds threshold
              </p>
            </div>
            <Switch
              checked={autoRollback}
              onCheckedChange={(val) => {
                setAutoRollback(val);
                toast.info(
                  val ? 'Auto-rollback enabled' : 'Auto-rollback disabled',
                  {
                    description: val
                      ? 'Mutations will automatically roll back if error rate exceeds threshold'
                      : 'Manual rollback required — errors will not auto-revert',
                    icon: <Info className="h-4 w-4 text-info" />,
                    duration: 3000,
                  }
                );
              }}
            />
          </div>

          {autoRollback && (
            <>
              <Separator />
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label className="text-base">Error Rate Threshold</Label>
                  <p className="text-sm text-text-muted">
                    Roll back when canary error rate exceeds this percentage
                  </p>
                </div>
                <Select value={autoRollbackThreshold} onValueChange={(val) => {
                  setAutoRollbackThreshold(val);
                  toast.success(`Rollback threshold set to ${val}%`, {
                    description: `Canary will auto-rollback if error rate exceeds ${val}%`,
                    icon: <AlertTriangle className="h-4 w-4 text-success" />,
                    duration: 2500,
                  });
                }}>
                  <SelectTrigger className="w-24">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {[2, 5, 10, 15, 25].map((v) => (
                      <SelectItem key={v} value={String(v)}>
                        {v}%
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* ─── Notifications ─── */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Bell className="h-5 w-5 text-brand-500" />
            DNA Notifications
          </CardTitle>
          <CardDescription>
            Choose which DNA events trigger notifications
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {[
            {
              key: 'notify_on_proposal',
              label: 'New Mutation Proposed',
              description: 'When AI proposes a new code optimization',
              state: notifyOnProposal,
              setter: setNotifyOnProposal,
              icon: '✨',
            },
            {
              key: 'notify_on_deploy',
              label: 'Mutation Deployed',
              description: 'When a canary deployment succeeds and code goes live',
              state: notifyOnDeploy,
              setter: setNotifyOnDeploy,
              icon: '🚀',
            },
            {
              key: 'notify_on_rollback',
              label: 'Mutation Rolled Back',
              description: 'When a deployment is automatically or manually rolled back',
              state: notifyOnRollback,
              setter: setNotifyOnRollback,
              icon: '↩️',
            },
          ].map(({ key, label, description, state, setter }) => (
            <div key={key} className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label className="text-sm">{label}</Label>
                <p className="text-xs text-text-muted">{description}</p>
              </div>
              <Switch
                checked={state}
                onCheckedChange={(val) => {
                  setter(val);
                  const action = val ? 'enabled' : 'disabled';
                  toast.success(`${label} ${action}`, {
                    description: val
                      ? `You'll be notified when ${label.toLowerCase()}`
                      : `Notifications disabled for ${label.toLowerCase()}`,
                    icon: <Bell className="h-4 w-4 text-success" />,
                    duration: 2500,
                  });
                }}
              />
            </div>
          ))}
        </CardContent>
      </Card>

      {/* ─── Quick Links ─── */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Server className="h-5 w-5 text-brand-500" />
            Quick Links
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {[
            { label: 'View Function DNA Dashboard', href: '/dna/overview', icon: Dna },
            { label: 'Evolution History', href: '/dna/overview', icon: GitBranch },
            { label: 'Performance Analytics', href: '/analytics', icon: Gauge },
          ].map(({ label, href, icon: Icon }) => (
            <Link
              key={label}
              to={href}
              className="flex items-center justify-between rounded-lg border border-border-subtle px-4 py-3 hover:border-border-default hover:bg-bg-tertiary/50 transition-colors group"
            >
              <div className="flex items-center gap-3">
                <Icon className="h-4 w-4 text-text-muted" />
                <span className="text-sm text-text-primary">{label}</span>
              </div>
              <ChevronRight className="h-4 w-4 text-text-muted group-hover:text-text-primary transition-colors" />
            </Link>
          ))}
        </CardContent>
      </Card>

      {/* ─── Cost Indicator ─── */}
      <div className="rounded-xl border border-border-subtle bg-bg-secondary/50 p-4">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-lg bg-velocity-500/10">
            <Zap className="h-4 w-4 text-velocity-500" />
          </div>
          <div>
            <p className="text-sm font-medium text-text-primary">Mutation Cost</p>
            <p className="text-xs text-text-muted">
              Each accepted mutation costs <span className="font-mono text-velocity-500">50 credits</span> from your wallet.
              Rejected mutations are free.
            </p>
          </div>
          <Badge variant="outline" className="ml-auto font-mono text-velocity-500 border-velocity-500/30">
            50 cr
          </Badge>
        </div>
      </div>

      {/* ─── Save Button ─── */}
      <div className="flex justify-end">
        <Button
          onClick={handleSave}
          disabled={updateSettings.isPending}
          className="gap-2"
        >
          {updateSettings.isPending ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin" />
              Saving...
            </>
          ) : (
            <>
              <CheckCircle2 className="h-4 w-4" />
              Save Platform Settings
            </>
          )}
        </Button>
      </div>
    </div>
  );
}
