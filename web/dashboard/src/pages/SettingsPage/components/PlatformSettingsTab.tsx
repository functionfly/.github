import { usersApi } from '@/api/users';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { Switch } from '@/components/ui/switch';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  AlertTriangle,
  Bell,
  CheckCircle2,
  ChevronRight,
  Clock,
  Dna,
  Gauge,
  GitBranch,
  Loader2,
  Server,
  Shield,
  Zap,
} from 'lucide-react';
import { useState } from 'react';
import { Trans, useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { toast } from 'sonner';

export function PlatformSettingsTab() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();

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

      const changes: string[] = [];
      if ('auto_evolve' in variables)
        changes.push(t('platformSettings.autoEvolution') + ' ' + (variables.auto_evolve ? t('platformSettings.autoEvolutionEnabled') : t('platformSettings.autoEvolutionDisabled')));
      if ('require_approval' in variables)
        changes.push(
          variables.require_approval ? t('platformSettings.manualApprovalRequired') : t('platformSettings.mutationsAutoApproved')
        );
      if ('sandbox_validation' in variables)
        changes.push(t('platformSettings.sandboxValidation') + ' ' + (variables.sandbox_validation ? t('platformSettings.sandboxEnabled') : t('platformSettings.sandboxDisabled')));
      if ('default_canary_pct' in variables)
        changes.push(t('platformSettings.defaultCanaryPct') + ': ' + variables.default_canary_pct + '%');
      if ('max_mutations_per_day' in variables)
        changes.push(t('platformSettings.maxMutationsPerDay') + ': ' + variables.max_mutations_per_day + '/day');
      if ('notify_on_proposal' in variables)
        changes.push(t('platformSettings.newMutationProposed') + ' ' + (variables.notify_on_proposal ? t('platformSettings.toastEnabled', { label: '' }).trim() : t('platformSettings.toastDisabled', { label: '' }).trim()));
      if ('notify_on_deploy' in variables)
        changes.push(t('platformSettings.mutationDeployed') + ' ' + (variables.notify_on_deploy ? t('platformSettings.toastEnabled', { label: '' }).trim() : t('platformSettings.toastDisabled', { label: '' }).trim()));
      if ('notify_on_rollback' in variables)
        changes.push(t('platformSettings.mutationRolledBack') + ' ' + (variables.notify_on_rollback ? t('platformSettings.toastEnabled', { label: '' }).trim() : t('platformSettings.toastDisabled', { label: '' }).trim()));
      if ('auto_rollback_on_error' in variables)
        changes.push(t('platformSettings.autoRollback') + ' ' + (variables.auto_rollback_on_error ? t('platformSettings.autoRollbackEnabled') : t('platformSettings.autoRollbackDisabled')));
      if ('auto_rollback_error_threshold' in variables)
        changes.push(t('platformSettings.errorRateThreshold') + ': ' + variables.auto_rollback_error_threshold + '%');

      if (changes.length > 0) {
        toast.success(t('platformSettings.toastSaved'), {
          description:
            changes.slice(0, 5).join(' · ') +
            (changes.length > 5 ? ' · ' + t('platformSettings.toastSavedDescMore', { count: changes.length - 5 }) : ''),
          icon: <CheckCircle2 className="h-4 w-4 text-success" />,
          duration: 4000,
        });
      } else {
        toast.success(t('platformSettings.toastSaved'));
      }
    },
    onError: (error: Error) => {
      toast.error(t('platformSettings.toastFailed'), {
        description: error?.message || t('platformSettings.toastFailed'),
        icon: <AlertTriangle className="h-4 w-4 text-error" />,
      });
    },
  });

  const [autoEvolve, setAutoEvolve] = useState((dna.auto_evolve as boolean) ?? true);
  const [requireApproval, setRequireApproval] = useState((dna.require_approval as boolean) ?? true);
  const [defaultCanaryPct, setDefaultCanaryPct] = useState(
    String((dna.default_canary_pct as number) ?? 10)
  );
  const [notifyOnProposal, setNotifyOnProposal] = useState(
    (dna.notify_on_proposal as boolean) ?? true
  );
  const [notifyOnDeploy, setNotifyOnDeploy] = useState((dna.notify_on_deploy as boolean) ?? true);
  const [notifyOnRollback, setNotifyOnRollback] = useState(
    (dna.notify_on_rollback as boolean) ?? true
  );
  const [maxMutationsPerDay, setMaxMutationsPerDay] = useState(
    String((dna.max_mutations_per_day as number) ?? 5)
  );
  const [sandboxValidation, setSandboxValidation] = useState(
    (dna.sandbox_validation as boolean) ?? true
  );
  const [autoRollback, setAutoRollback] = useState((dna.auto_rollback_on_error as boolean) ?? true);
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
    <div className="settings-page space-y-6">
      <Card className="settings-panel">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Dna className="h-5 w-5 text-brand-500" />
            {t('platformSettings.functionDNA')}
          </CardTitle>
          <CardDescription>
            {t('platformSettings.functionDNADesc')}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label className="text-base flex items-center gap-2">
                <Zap className="h-4 w-4 text-velocity-500" />
                {t('platformSettings.autoEvolution')}
              </Label>
              <p className="text-sm text-text-muted">
                {t('platformSettings.autoEvolutionDesc')}
              </p>
            </div>
            <Switch
              checked={autoEvolve}
              onCheckedChange={(val) => setAutoEvolve(val)}
            />
          </div>

          <Separator />

          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label className="text-base flex items-center gap-2">
                <Shield className="h-4 w-4 text-info" />
                {t('platformSettings.requireApproval')}
              </Label>
              <p className="text-sm text-text-muted">
                {t('platformSettings.requireApprovalDesc')}
              </p>
            </div>
            <Switch
              checked={requireApproval}
              onCheckedChange={(val) => setRequireApproval(val)}
            />
          </div>

          <Separator />

          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label className="text-base flex items-center gap-2">
                <Shield className="h-4 w-4 text-success" />
                {t('platformSettings.sandboxValidation')}
              </Label>
              <p className="text-sm text-text-muted">
                {t('platformSettings.sandboxValidationDesc')}
              </p>
            </div>
            <Switch
              checked={sandboxValidation}
              onCheckedChange={(val) => setSandboxValidation(val)}
            />
          </div>

          <Separator />

          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label className="text-base flex items-center gap-2">
                <GitBranch className="h-4 w-4 text-warning" />
                {t('platformSettings.defaultCanaryPct')}
              </Label>
              <p className="text-sm text-text-muted">
                {t('platformSettings.defaultCanaryPctDesc')}
              </p>
            </div>
            <Select
              value={defaultCanaryPct}
              onValueChange={(val) => setDefaultCanaryPct(val)}
            >
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

          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label className="text-base flex items-center gap-2">
                <Clock className="h-4 w-4 text-text-muted" />
                {t('platformSettings.maxMutationsPerDay')}
              </Label>
              <p className="text-sm text-text-muted">
                {t('platformSettings.maxMutationsPerDayDesc')}
              </p>
            </div>
            <Select
              value={maxMutationsPerDay}
              onValueChange={(val) => setMaxMutationsPerDay(val)}
            >
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

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-warning" />
            {t('platformSettings.autoRollback')}
          </CardTitle>
          <CardDescription>
            {t('platformSettings.autoRollbackDesc')}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex items-center justify-between">
            <div className="space-y-0.5">
              <Label className="text-base flex items-center gap-2">
                <CheckCircle2 className="h-4 w-4 text-success" />
                {t('platformSettings.enableAutoRollback')}
              </Label>
              <p className="text-sm text-text-muted">
                {t('platformSettings.enableAutoRollbackDesc')}
              </p>
            </div>
            <Switch
              checked={autoRollback}
              onCheckedChange={(val) => setAutoRollback(val)}
            />
          </div>

          {autoRollback && (
            <>
              <Separator />
              <div className="flex items-center justify-between">
                <div className="space-y-0.5">
                  <Label className="text-base">{t('platformSettings.errorRateThreshold')}</Label>
                  <p className="text-sm text-text-muted">
                    {t('platformSettings.errorRateThresholdDesc')}
                  </p>
                </div>
                <Select
                  value={autoRollbackThreshold}
                  onValueChange={(val) => setAutoRollbackThreshold(val)}
                >
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

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Bell className="h-5 w-5 text-brand-500" />
            {t('platformSettings.dnaNotifications')}
          </CardTitle>
          <CardDescription>{t('platformSettings.dnaNotificationsDesc')}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {[
            {
              key: 'notify_on_proposal',
              label: t('platformSettings.newMutationProposed'),
              description: t('platformSettings.newMutationProposedDesc'),
              state: notifyOnProposal,
              setter: setNotifyOnProposal,
              icon: '✨',
            },
            {
              key: 'notify_on_deploy',
              label: t('platformSettings.mutationDeployed'),
              description: t('platformSettings.mutationDeployedDesc'),
              state: notifyOnDeploy,
              setter: setNotifyOnDeploy,
              icon: '🚀',
            },
            {
              key: 'notify_on_rollback',
              label: t('platformSettings.mutationRolledBack'),
              description: t('platformSettings.mutationRolledBackDesc'),
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
                onCheckedChange={(val) => setter(val)}
              />
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Server className="h-5 w-5 text-brand-500" />
            {t('platformSettings.quickLinks')}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {[
            { label: t('platformSettings.viewDNADashboard'), href: '/dna/overview', icon: Dna },
            { label: t('platformSettings.evolutionHistory'), href: '/dna/overview', icon: GitBranch },
            { label: t('platformSettings.performanceAnalytics'), href: '/analytics', icon: Gauge },
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

      <div className="rounded-xl border border-border-subtle bg-bg-secondary/50 p-4">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-lg bg-velocity-500/10">
            <Zap className="h-4 w-4 text-velocity-500" />
          </div>
          <div>
            <p className="text-sm font-medium text-text-primary">{t('platformSettings.mutationCost')}</p>
            <p className="text-xs text-text-muted">
              <Trans i18nKey="platformSettings.mutationCostDesc" components={{ 1: <span className="font-mono text-velocity-500" /> }} values={{ cost: '50 credits' }} />
            </p>
          </div>
          <Badge
            variant="outline"
            className="ml-auto font-mono text-velocity-500 border-velocity-500/30"
          >
            {t('platformSettings.mutationCostBadge')}
          </Badge>
        </div>
      </div>

      <div className="flex justify-end">
        <Button
          onClick={handleSave}
          disabled={updateSettings.isPending}
          className="gap-2"
          style={{
            background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
            color: 'var(--text-on-light)',
            boxShadow: 'var(--shadow-btn-primary-rest)',
          }}
        >
          {updateSettings.isPending ? (
            <>
              <Loader2 className="h-4 w-4 animate-spin" />
              {t('platformSettings.saving')}
            </>
          ) : (
            <>
              <CheckCircle2 className="h-4 w-4" />
              {t('platformSettings.savePlatformSettings')}
            </>
          )}
        </Button>
      </div>
    </div>
  );
}
