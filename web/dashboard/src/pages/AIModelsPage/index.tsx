import { aiModelsApi, type ModelSelection, type TenantAIPreferences } from '@/api/aiModels';
import { apiClient } from '@/api/client';
import { usersApi } from '@/api/users';
import { ModelPicker } from '@/components/ai/ModelPicker';
import { PageHeader } from '@/components/layout/PageHeader';
import { PageLayout } from '@/components/layout/PageLayout';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { useAuthStore } from '@/stores/authStore';
import type { TenantMembership } from '@/types';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { AlertTriangle, Brain, CheckCircle2, Globe, Loader2, Lock, Save, Shield } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { toast } from 'sonner';
import { GlobalDefaultSection } from './GlobalDefaultSection';
import { ModelCatalogTable } from './ModelCatalogTable';
import { ModelProfileSelector } from './ModelProfileSelector';
import { UserOverridesSection } from './UserOverridesSection';
import {
  clearAllowlist,
  clearProviderAllowlist,
  enableAllModels,
  enableAllProviders,
  expandProfileDefaults,
  FEATURE_DEFAULTS,
  formatModelLabel,
  isProviderEnabled,
  preferencesEqual,
  toggleModelAllowlist,
  toggleProviderAllowlist,
} from './utils';

const emptyPreferences: TenantAIPreferences = {
  profile: 'balanced',
  use_same_model_everywhere: false,
  defaults: {},
  enabled_models: [],
  enabled_providers: [],
  allow_user_overrides: true,
  routing_strategy: 'quality_first',
};

async function fetchTenantMembers(): Promise<TenantMembership[]> {
  const data = await apiClient.get<{ members: TenantMembership[] }>('/v1/auth/members');
  return data.members || [];
}

function parseUserOverrides(
  settings: Record<string, unknown> | undefined
): Record<string, ModelSelection | undefined> {
  const ai = settings?.ai as { overrides?: Record<string, ModelSelection> } | undefined;
  const overrides = ai?.overrides ?? {};
  return {
    composer: overrides.composer,
    frg: overrides.frg,
  };
}

export default function AIModelsPage() {
  const queryClient = useQueryClient();
  const user = useAuthStore((s) => s.user);
  const [draft, setDraft] = useState<TenantAIPreferences | null>(null);
  const [userOverrides, setUserOverrides] = useState<Record<string, ModelSelection | undefined>>(
    {}
  );

  const {
    data: prefsData,
    isLoading: prefsLoading,
    refetch: refetchPrefs,
  } = useQuery({
    queryKey: ['ai-model-preferences'],
    queryFn: aiModelsApi.getPreferences,
  });

  const {
    data: catalog = [],
    isLoading: catalogLoading,
    error: catalogError,
    refetch: refetchCatalog,
  } = useQuery({
    queryKey: ['ai-model-catalog', 'all'],
    queryFn: () => aiModelsApi.getCatalog(),
    retry: 1,
  });

  const { data: members = [] } = useQuery({
    queryKey: ['tenant-members'],
    queryFn: fetchTenantMembers,
  });

  const { data: mySettings } = useQuery({
    queryKey: ['my-settings-ai'],
    queryFn: async () => {
      const data = await usersApi.getMySettings();
      return (data as { settings?: Record<string, unknown> }).settings;
    },
  });

  useEffect(() => {
    if (mySettings) {
      setUserOverrides(parseUserOverrides(mySettings));
    }
  }, [mySettings]);

  const saved = prefsData ?? emptyPreferences;
  const value = draft ?? saved;

  const canManageOrg = useMemo(() => {
    const me = members.find((m) => m.user_id === user?.id);
    if (!me) return members.length === 0;
    return me.role === 'team_owner' || me.role === 'team_admin';
  }, [members, user?.id]);

  const isDirty = draft !== null && !preferencesEqual(draft, saved);

  const saveOrgMutation = useMutation({
    mutationFn: (payload: TenantAIPreferences) => aiModelsApi.updatePreferences(payload),
    onSuccess: async () => {
      toast.success('Organization AI model preferences saved');
      setDraft(null);
      await refetchPrefs();
      await refetchCatalog();
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to save preferences');
    },
  });

  const refreshCatalogMutation = useMutation({
    mutationFn: aiModelsApi.refreshCatalog,
    onSuccess: async () => {
      toast.success('Model catalog refreshed');
      await refetchCatalog();
    },
    onError: () => toast.error('Failed to refresh catalog'),
  });

  const saveOverridesMutation = useMutation({
    mutationFn: (overrides: Record<string, ModelSelection>) =>
      aiModelsApi.updateMyOverrides(overrides),
    onSuccess: () => {
      toast.success('Your AI model overrides saved');
      queryClient.invalidateQueries({ queryKey: ['my-settings-ai'] });
    },
    onError: (err: Error) => toast.error(err.message || 'Failed to save overrides'),
  });

  const updateDraft = (patch: Partial<TenantAIPreferences>) => {
    setDraft({ ...value, ...patch });
  };

  const handleToggleAllowlist = (provider: string, modelId: string, enabled: boolean) => {
    updateDraft({
      enabled_models: toggleModelAllowlist(
        value.enabled_models,
        catalog,
        provider,
        modelId,
        enabled
      ),
    });
  };

  const handleToggleProvider = (provider: string, enabled: boolean) => {
    updateDraft({
      enabled_providers: toggleProviderAllowlist(
        value.enabled_providers,
        catalog,
        provider,
        enabled
      ),
    });
  };

  const providers = useMemo(() => {
    const map = new Map<string, { count: number; available: number }>();
    for (const m of catalog) {
      const entry = map.get(m.provider) ?? { count: 0, available: 0 };
      entry.count++;
      if (m.provider_available !== false) entry.available++;
      map.set(m.provider, entry);
    }
    return [...map.entries()].sort(([a], [b]) => a.localeCompare(b));
  }, [catalog]);

  const handleSaveOverrides = () => {
    const payload: Record<string, ModelSelection> = {};
    for (const [feature, selection] of Object.entries(userOverrides)) {
      if (selection?.provider && selection?.model_id) {
        payload[feature] = selection;
      }
    }
    saveOverridesMutation.mutate(payload);
  };

  if (prefsLoading) {
    return (
      <PageLayout maxWidth="5xl">
        <div className="flex items-center justify-center py-24">
          <Loader2 className="h-8 w-8 animate-spin text-text-muted" />
        </div>
      </PageLayout>
    );
  }

  return (
    <PageLayout maxWidth="5xl">
      <PageHeader
        title="AI Models"
        subtitle="Configure organization-wide FlyMind model defaults, allowlists, and personal overrides."
        badges={[{ label: 'FlyMind Gateway', variant: 'outline', icon: Brain }]}
      />

      {!canManageOrg && (
        <div className="mb-6 flex items-start gap-3 rounded-xl border border-border-subtle bg-bg-secondary/40 p-4 text-sm">
          <Shield className="mt-0.5 h-4 w-4 shrink-0 text-text-muted" />
          <p className="text-text-muted">
            You can view org settings and manage your personal overrides. Contact a tenant admin to
            change organization defaults.
          </p>
        </div>
      )}

      {/* Provider toggles — visible to admins */}
      {canManageOrg && providers.length > 0 && (
        <Card className="mb-6">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Globe className="h-5 w-5 text-brand-500" />
              Providers
            </CardTitle>
            <CardDescription>
              Enable or disable AI providers. Disabled providers hide all their models from selectors. Empty means all providers are allowed.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-2 mb-4">
              <Button
                variant="outline"
                size="sm"
                onClick={() => updateDraft({ enabled_providers: clearProviderAllowlist() })}
              >
                Allow all
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => updateDraft({ enabled_providers: enableAllProviders(catalog) })}
              >
                Restrict to listed
              </Button>
              {value.enabled_providers.length > 0 && (
                <Badge variant="outline" className="gap-1">
                  {value.enabled_providers.length} of {providers.length} enabled
                </Badge>
              )}
            </div>
            <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
              {providers.map(([name, stats]) => (
                <div
                  key={name}
                  className="flex items-center justify-between rounded-lg border border-border-subtle px-4 py-3"
                >
                  <div className="space-y-0.5">
                    <p className="text-sm font-medium capitalize">{name}</p>
                    <p className="text-xs text-text-muted">
                      {stats.count} model{stats.count !== 1 ? 's' : ''}
                      {stats.available < stats.count && (
                        <span className="ml-1 text-warning">({stats.available} available)</span>
                      )}
                    </p>
                  </div>
                  <Switch
                    checked={isProviderEnabled(value.enabled_providers, name)}
                    onCheckedChange={(checked) => handleToggleProvider(name, checked)}
                  />
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Model catalog — visible to everyone */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Brain className="h-5 w-5 text-brand-500" />
            Model catalog
          </CardTitle>
          <CardDescription>
            Models available through FlyMind for your organization. Admins can restrict which models
            members may select.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <ModelCatalogTable
            catalog={catalog}
            enabledModels={canManageOrg ? value.enabled_models : saved.enabled_models}
            isLoading={catalogLoading}
            error={catalogError instanceof Error ? catalogError.message : null}
            onToggle={handleToggleAllowlist}
            onEnableAll={() => updateDraft({ enabled_models: enableAllModels(catalog) })}
            onAllowAll={() => updateDraft({ enabled_models: clearAllowlist() })}
            onRefresh={() => refreshCatalogMutation.mutate()}
            isRefreshing={refreshCatalogMutation.isPending}
            disabled={!canManageOrg}
          />
        </CardContent>
      </Card>

      {canManageOrg && (
        <Card className="mb-6">
          <CardHeader>
            <CardTitle>Organization profile</CardTitle>
            <CardDescription>
              Choose a preset that maps to recommended models per feature, or customize defaults
              below.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <ModelProfileSelector
              value={value.profile}
              onChange={(profile) => {
                if (profile === 'custom') return;
                updateDraft({
                  profile,
                  defaults: { ...value.defaults, ...expandProfileDefaults(profile) },
                });
              }}
            />
            {value.profile === 'custom' && (
              <p className="text-sm text-text-muted">
                Custom profile — per-feature defaults below override preset mappings.
              </p>
            )}

            <GlobalDefaultSection
              useSameModelEverywhere={value.use_same_model_everywhere}
              globalDefault={value.global_default}
              onToggleSameModel={(checked) => updateDraft({ use_same_model_everywhere: checked })}
              onGlobalDefaultChange={(next) => updateDraft({ global_default: next })}
            />

            {!value.use_same_model_everywhere && (
              <div className="grid gap-4 sm:grid-cols-2">
                {FEATURE_DEFAULTS.map((feature) => (
                  <div
                    key={feature.key}
                    className="space-y-2 rounded-lg border border-border-subtle p-4"
                  >
                    <Label>{feature.label}</Label>
                    <ModelPicker
                      feature={feature.key}
                      capability={feature.capability}
                      value={value.defaults?.[feature.key]}
                      onChange={(next) =>
                        updateDraft({
                          profile: 'custom',
                          defaults: {
                            ...value.defaults,
                            ...(next ? { [feature.key]: next } : {}),
                          },
                        })
                      }
                      showOrgDefaultOption={false}
                    />
                    <p className="text-xs text-text-muted">
                      Current: {formatModelLabel(value.defaults?.[feature.key])}
                    </p>
                  </div>
                ))}
              </div>
            )}

            <div className="flex items-center justify-between rounded-lg border border-border-subtle p-4">
              <div className="space-y-1">
                <Label className="flex items-center gap-2 text-base">
                  <Lock className="h-4 w-4" />
                  Allow user overrides
                </Label>
                <p className="text-sm text-text-muted">
                  When disabled, members cannot set personal model preferences (compliance lock).
                </p>
              </div>
              <Switch
                checked={value.allow_user_overrides}
                onCheckedChange={(checked) => updateDraft({ allow_user_overrides: checked })}
              />
            </div>

            <div className="flex flex-wrap items-center gap-3">
              <Button
                onClick={() => saveOrgMutation.mutate(value)}
                disabled={!isDirty || saveOrgMutation.isPending}
                className="gap-2"
              >
                {saveOrgMutation.isPending ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Save className="h-4 w-4" />
                )}
                Save organization settings
              </Button>
              {isDirty && (
                <Badge variant="outline" className="gap-1">
                  <AlertTriangle className="h-3 w-3" />
                  Unsaved changes
                </Badge>
              )}
              {!isDirty && (
                <span className="inline-flex items-center gap-1 text-sm text-text-muted">
                  <CheckCircle2 className="h-4 w-4 text-success" />
                  Saved
                </span>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {!canManageOrg && (
        <Card className="mb-6">
          <CardHeader>
            <CardTitle>Organization defaults</CardTitle>
            <CardDescription>Read-only view of your org&apos;s AI configuration.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex flex-wrap gap-2">
              <Badge variant="outline">Profile: {saved.profile}</Badge>
              {saved.use_same_model_everywhere && (
                <Badge variant="outline">Same model everywhere</Badge>
              )}
              {!saved.allow_user_overrides && (
                <Badge variant="outline" className="border-warning/40 text-warning">
                  Overrides locked
                </Badge>
              )}
            </div>
            <p className="text-sm text-text-muted">
              Composer default: {formatModelLabel(saved.defaults?.composer ?? saved.global_default)}
            </p>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>My overrides</CardTitle>
          <CardDescription>
            Personal model picks for Composer and FRG. Used only when your org allows overrides.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {!saved.allow_user_overrides ? (
            <div className="flex items-start gap-3 rounded-lg border border-warning/30 bg-warning/5 p-4 text-sm">
              <Lock className="mt-0.5 h-4 w-4 shrink-0 text-warning" />
              <p className="text-text-muted">
                Your organization has disabled personal model overrides.
              </p>
            </div>
          ) : (
            <>
              <UserOverridesSection
                overrides={userOverrides}
                onChange={(feature, next) =>
                  setUserOverrides((prev) => ({ ...prev, [feature]: next }))
                }
              />
              <Button
                onClick={handleSaveOverrides}
                disabled={saveOverridesMutation.isPending}
                className="gap-2"
              >
                {saveOverridesMutation.isPending ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Save className="h-4 w-4" />
                )}
                Save my overrides
              </Button>
            </>
          )}
        </CardContent>
      </Card>
    </PageLayout>
  );
}
