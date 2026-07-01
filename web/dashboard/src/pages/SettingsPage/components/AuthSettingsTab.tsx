import { apiClient } from '@/api/client';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useAuthStore } from '@/stores/authStore';
import type {
  PasswordPolicy,
  TenantAuthSettings,
  TenantInviteCode,
  TenantMembership,
  TenantOAuthProvider,
} from '@/types';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Building, Copy, Eye, EyeOff, Key, Shield, Trash2, UserPlus } from 'lucide-react';
import { useState } from 'react';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

interface OAuthProviderInput {
  provider: string;
  client_id: string;
  client_secret: string;
  enabled: boolean;
  callback_url?: string;
}

async function fetchAuthSettings(): Promise<TenantAuthSettings> {
  return apiClient.get<TenantAuthSettings>('/v1/auth/settings');
}

async function updateAuthSettings(
  updates: Partial<TenantAuthSettings>
): Promise<TenantAuthSettings> {
  return apiClient.patch<TenantAuthSettings>('/v1/auth/settings', updates);
}

async function fetchOAuthProviders(): Promise<TenantOAuthProvider[]> {
  const data = await apiClient.get<{ providers: TenantOAuthProvider[] }>('/v1/auth/oauth');
  return data.providers || [];
}

async function configureOAuthProvider(config: OAuthProviderInput): Promise<void> {
  await apiClient.post('/v1/auth/oauth', config);
}

async function deleteOAuthProvider(provider: string): Promise<void> {
  await apiClient.delete(`/v1/auth/oauth/${provider}`);
}

async function fetchMembers(): Promise<TenantMembership[]> {
  const data = await apiClient.get<{ members: TenantMembership[] }>('/v1/auth/members');
  return data.members || [];
}

async function inviteMember(email: string, role: string): Promise<TenantInviteCode> {
  const data = await apiClient.post<{ invite: TenantInviteCode }>('/v1/auth/members/invite', {
    email,
    role,
  });
  return data.invite;
}

async function updateMemberRole(userId: string, role: string): Promise<void> {
  await apiClient.patch(`/v1/auth/members/${userId}/role`, { role });
}

async function removeMember(userId: string): Promise<void> {
  await apiClient.delete(`/v1/auth/members/${userId}`);
}

async function fetchPendingInvites(): Promise<TenantInviteCode[]> {
  const data = await apiClient.get<{ invites: TenantInviteCode[] }>('/v1/auth/invites');
  return data.invites || [];
}

async function revokeInvite(code: string): Promise<void> {
  await apiClient.delete(`/v1/auth/invites/${code}`);
}

// Password Policy Editor Component
function PasswordPolicyEditor({
  policy,
  onSave,
}: {
  policy: PasswordPolicy;
  onSave: (policy: PasswordPolicy) => void;
}) {
  const { t } = useTranslation();
  const [localPolicy, setLocalPolicy] = useState(policy);
  const [saving, setSaving] = useState(false);

  const handleSave = async () => {
    setSaving(true);
    try {
      await onSave(localPolicy);
      toast.success(t('authSettings.toastPolicyUpdated'));
    } catch {
      toast.error(t('authSettings.toastPolicyFailed'));
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div>
          <Label htmlFor="min_length">{t('authSettings.minLength')}</Label>
          <Input
            id="min_length"
            type="number"
            min={8}
            max={128}
            value={localPolicy.min_length}
            onChange={(e) =>
              setLocalPolicy({ ...localPolicy, min_length: parseInt(e.target.value) || 8 })
            }
          />
        </div>
      </div>

      <div className="space-y-3">
        <div className="flex items-center gap-3">
          <Switch
            id="require_uppercase"
            checked={localPolicy.require_uppercase}
            onCheckedChange={(checked) =>
              setLocalPolicy({ ...localPolicy, require_uppercase: checked })
            }
          />
          <Label htmlFor="require_uppercase">{t('authSettings.requireUppercase')}</Label>
        </div>
        <div className="flex items-center gap-3">
          <Switch
            id="require_lowercase"
            checked={localPolicy.require_lowercase}
            onCheckedChange={(checked) =>
              setLocalPolicy({ ...localPolicy, require_lowercase: checked })
            }
          />
          <Label htmlFor="require_lowercase">{t('authSettings.requireLowercase')}</Label>
        </div>
        <div className="flex items-center gap-3">
          <Switch
            id="require_digit"
            checked={localPolicy.require_digit}
            onCheckedChange={(checked) =>
              setLocalPolicy({ ...localPolicy, require_digit: checked })
            }
          />
          <Label htmlFor="require_digit">{t('authSettings.requireDigit')}</Label>
        </div>
        <div className="flex items-center gap-3">
          <Switch
            id="require_special"
            checked={localPolicy.require_special}
            onCheckedChange={(checked) =>
              setLocalPolicy({ ...localPolicy, require_special: checked })
            }
          />
          <Label htmlFor="require_special">{t('authSettings.requireSpecial')}</Label>
        </div>
      </div>

      <Button
        onClick={handleSave}
        disabled={saving}
        style={{
          background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
          color: 'var(--text-on-light)',
          boxShadow: 'var(--shadow-btn-primary-rest)',
        }}
      >
        {saving ? t('authSettings.saving') : t('authSettings.savePolicy')}
      </Button>
    </div>
  );
}

// OAuth Provider Card Component
function OAuthProviderCard({
  provider,
  onEdit,
  onDelete,
}: {
  provider: TenantOAuthProvider;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const { t } = useTranslation();
  const providerIcons: Record<string, string> = {
    github: '🐙',
    google: '🔵',
    microsoft: '🪟',
    apple: '🍎',
  };

  return (
    <Card className="relative">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-bg-secondary flex items-center justify-center text-xl">
              {providerIcons[provider.provider] || '🔐'}
            </div>
            <div>
              <CardTitle className="text-base capitalize">{provider.provider}</CardTitle>
              <CardDescription>{provider.enabled ? t('authSettings.connected') : t('authSettings.disabled')}</CardDescription>
            </div>
          </div>
          <Badge variant={provider.enabled ? 'default' : 'secondary'}>
            {provider.enabled ? t('authSettings.active') : t('authSettings.inactive')}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={onEdit}
            style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
          >
            {t('authSettings.edit')}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={onDelete}
            style={{ color: 'var(--status-revoked)' }}
          >
            <Trash2 className="w-4 h-4" />
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

// OAuth Provider Form Modal
function OAuthProviderForm({
  provider,
  onSave,
  onCancel,
}: {
  provider?: OAuthProviderInput;
  onSave: (config: OAuthProviderInput) => void;
  onCancel: () => void;
}) {
  const { t } = useTranslation();
  const [formData, setFormData] = useState<OAuthProviderInput>(
    provider || {
      provider: 'github',
      client_id: '',
      client_secret: '',
      enabled: true,
    }
  );
  const [showSecret, setShowSecret] = useState(false);

  const handleSave = () => {
    if (!formData.client_id || !formData.client_secret) {
      toast.error(t('authSettings.toastClientRequired'));
      return;
    }
    onSave(formData);
  };

  return (
    <div className="space-y-4">
      <div>
        <Label>Provider</Label>
        <Select
          value={formData.provider}
          onValueChange={(value) => setFormData({ ...formData, provider: value })}
          disabled={!!provider}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="github">GitHub</SelectItem>
            <SelectItem value="google">Google</SelectItem>
            <SelectItem value="microsoft">Microsoft</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div>
        <Label htmlFor="client_id">{t('authSettings.clientID')}</Label>
        <Input
          id="client_id"
          value={formData.client_id}
          onChange={(e) => setFormData({ ...formData, client_id: e.target.value })}
          placeholder={t('authSettings.enterClientID')}
        />
      </div>

      <div>
        <Label htmlFor="client_secret">{t('authSettings.clientSecret')}</Label>
        <div className="relative">
          <Input
            id="client_secret"
            type={showSecret ? 'text' : 'password'}
            value={formData.client_secret}
            onChange={(e) => setFormData({ ...formData, client_secret: e.target.value })}
            placeholder={provider ? '••••••••' : t('authSettings.enterClientSecret')}
          />
          <button
            type="button"
            className="absolute right-2 top-1/2 -translate-y-1/2 text-text-muted"
            onClick={() => setShowSecret(!showSecret)}
          >
            {showSecret ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
          </button>
        </div>
        {provider && (
          <p className="text-xs text-text-muted mt-1">{t('authSettings.keepSecretHint')}</p>
        )}
      </div>

      <div className="flex items-center gap-3">
        <Switch
          id="enabled"
          checked={formData.enabled}
          onCheckedChange={(checked) => setFormData({ ...formData, enabled: checked })}
        />
        <Label htmlFor="enabled">{t('authSettings.enableProvider')}</Label>
      </div>

      <div className="flex justify-end gap-2">
        <Button
          variant="outline"
          onClick={onCancel}
          style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
        >
          Cancel
        </Button>
        <Button
          onClick={handleSave}
          style={{
            background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
            color: 'var(--text-on-light)',
            boxShadow: 'var(--shadow-btn-primary-rest)',
          }}
        >
          {t('authSettings.saveProvider')}
        </Button>
      </div>
    </div>
  );
}

// Team Member Row Component
function MemberRow({
  member,
  onRoleChange,
  onRemove,
  canManage,
}: {
  member: TenantMembership;
  onRoleChange: (role: string) => void;
  onRemove: () => void;
  canManage: boolean;
}) {
  const { t } = useTranslation();
  const roleColors: Record<string, string> = {
    team_owner: 'bg-amber-500/10 text-amber-500 border-amber-500/20',
    team_admin: 'bg-blue-500/10 text-blue-500 border-blue-500/20',
    team_member: 'bg-green-500/10 text-green-500 border-green-500/20',
    team_viewer: 'bg-gray-500/10 text-gray-500 border-gray-500/20',
  };

  return (
    <div className="flex items-center justify-between py-3 border-b border-border last:border-0">
      <div className="flex items-center gap-3">
        <div className="w-8 h-8 rounded-full bg-brand-500/20 flex items-center justify-center">
          <span className="text-sm font-medium text-brand-400">
            {member.user?.email?.[0]?.toUpperCase() || 'U'}
          </span>
        </div>
        <div>
          <p className="font-medium text-text-primary">{member.user?.name || member.user?.email}</p>
          <p className="text-sm text-text-muted">{member.user?.email}</p>
        </div>
      </div>
      <div className="flex items-center gap-3">
        {canManage ? (
          <Select value={member.role} onValueChange={onRoleChange}>
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="team_owner">{t('authSettings.roles.owner')}</SelectItem>
              <SelectItem value="team_admin">{t('authSettings.roles.admin')}</SelectItem>
              <SelectItem value="team_member">{t('authSettings.roles.member')}</SelectItem>
              <SelectItem value="team_viewer">{t('authSettings.roles.viewer')}</SelectItem>
            </SelectContent>
          </Select>
        ) : (
          <Badge className={roleColors[member.role] || ''}>{member.role.replace('_', ' ')}</Badge>
        )}
        {canManage && (
          <Button variant="ghost" size="sm" onClick={onRemove} className="text-destructive">
            <Trash2 className="w-4 h-4" />
          </Button>
        )}
      </div>
    </div>
  );
}

// Invite Row Component
function InviteRow({
  invite,
  onRevoke,
  onCopy,
}: {
  invite: TenantInviteCode;
  onRevoke: () => void;
  onCopy: () => void;
}) {
  const { t } = useTranslation();
  const expiresAt = new Date(invite.expires_at);
  const isExpired = expiresAt < new Date();

  return (
    <div className="flex items-center justify-between py-3 border-b border-border last:border-0">
      <div>
        <p className="font-medium text-text-primary">{invite.email}</p>
        <p className="text-sm text-text-muted">
          {isExpired ? t('authSettings.expired') : t('authSettings.expires', { date: expiresAt.toLocaleDateString() })} ·{' '}
          {invite.role.replace('_', ' ')}
        </p>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm" onClick={onCopy}>
          <Copy className="w-4 h-4" />
        </Button>
        {!isExpired && (
          <Button variant="ghost" size="sm" onClick={onRevoke} className="text-destructive">
            <Trash2 className="w-4 h-4" />
          </Button>
        )}
      </div>
    </div>
  );
}

// Main Auth Settings Tab Component
export function AuthSettingsTab() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const user = useAuthStore((s) => s.user);
  const [showOAuthForm, setShowOAuthForm] = useState(false);
  const [editingProvider, setEditingProvider] = useState<OAuthProviderInput | null>(null);
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteRole, setInviteRole] = useState('team_member');
  const [showInviteForm, setShowInviteForm] = useState(false);
  const [confirmRemoveMember, setConfirmRemoveMember] = useState<string | null>(null);
  const [confirmDeleteProvider, setConfirmDeleteProvider] = useState<string | null>(null);
  const [confirmRevokeInvite, setConfirmRevokeInvite] = useState<string | null>(null);

  // Fetch auth settings
  const { data: settings, isLoading: settingsLoading } = useQuery({
    queryKey: ['auth-settings'],
    queryFn: fetchAuthSettings,
  });

  // Fetch OAuth providers
  const { data: oauthProviders = [] } = useQuery({
    queryKey: ['oauth-providers'],
    queryFn: fetchOAuthProviders,
  });

  // Fetch members
  const { data: members = [] } = useQuery({
    queryKey: ['team-members'],
    queryFn: fetchMembers,
  });

  // Fetch pending invites
  const { data: pendingInvites = [] } = useQuery({
    queryKey: ['pending-invites'],
    queryFn: fetchPendingInvites,
  });

  // Mutations
  const updateSettingsMutation = useMutation({
    mutationFn: updateAuthSettings,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['auth-settings'] });
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to update settings');
    },
  });

  const revokeInviteMutation = useMutation({
    mutationFn: revokeInvite,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pending-invites'] });
      toast.success(t('authSettings.toastInviteRevoked'));
    },
  });

  const oauthMutation = useMutation({
    mutationFn: configureOAuthProvider,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['oauth-providers'] });
      setShowOAuthForm(false);
      setEditingProvider(null);
      toast.success(t('authSettings.toastOAuthSaved'));
    },
  });

  const deleteOAuthMutation = useMutation({
    mutationFn: deleteOAuthProvider,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['oauth-providers'] });
      toast.success(t('authSettings.toastProviderRemoved'));
    },
  });

  const inviteMutation = useMutation({
    mutationFn: ({ email, role }: { email: string; role: string }) => inviteMember(email, role),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pending-invites'] });
      setShowInviteForm(false);
      setInviteEmail('');
      toast.success(t('authSettings.toastInviteSent'));
    },
  });

  const updateRoleMutation = useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: string }) =>
      updateMemberRole(userId, role),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-members'] });
      toast.success(t('authSettings.toastRoleUpdated'));
    },
  });

  const removeMemberMutation = useMutation({
    mutationFn: (userId: string) => removeMember(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-members'] });
      toast.success(t('authSettings.toastMemberRemoved'));
    },
  });

  const canManageTeam = members.some(
    (m) => m.user_id === user?.id && (m.role === 'team_owner' || m.role === 'team_admin')
  );

  if (settingsLoading) {
    return (
      <div className="settings-page space-y-6 animate-pulse">
        <div className="flex gap-2">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-9 w-24 rounded-md bg-bg-secondary" />
          ))}
        </div>
        <div className="rounded-lg p-5 space-y-4" style={{ background: 'var(--panel)', border: '1px solid var(--panel-edge)' }}>
          <div className="h-5 w-32 rounded bg-bg-secondary" />
          <div className="h-4 w-64 rounded bg-bg-secondary/60" />
          <div className="space-y-3 pt-2">
            {[1, 2, 3].map((i) => (
              <div key={i} className="flex items-center justify-between">
                <div className="h-4 w-40 rounded bg-bg-secondary" />
                <div className="h-6 w-10 rounded-full bg-bg-secondary" />
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="settings-page space-y-6">
      <Tabs defaultValue="security" className="space-y-6">
        <TabsList>
          <TabsTrigger value="security">
            <Shield className="w-4 h-4 mr-2" />
            {t('authSettings.security')}
          </TabsTrigger>
          <TabsTrigger value="oauth">
            <Key className="w-4 h-4 mr-2" />
            {t('authSettings.oauth')}
          </TabsTrigger>
          <TabsTrigger value="team">
            <Building className="w-4 h-4 mr-2" />
            {t('authSettings.team')}
          </TabsTrigger>
        </TabsList>

        {/* Security Tab */}
        <TabsContent value="security" className="space-y-6">
          <div
            className="rounded-lg p-5"
            style={{
              background: 'var(--panel)',
              border: '1px solid var(--panel-edge)',
              boxShadow: 'var(--shadow-chamber)',
            }}
          >
            <div className="mb-4">
              <h3 className="font-display text-lg font-semibold" style={{ color: 'var(--text)' }}>
                {t('authSettings.passwordPolicy')}
              </h3>
              <p className="text-sm mt-1" style={{ color: 'var(--text-dim)' }}>
                {t('authSettings.passwordPolicyDesc')}
              </p>
            </div>
            <div className="space-y-4">
              <PasswordPolicyEditor
                policy={
                  settings?.password_policy || {
                    min_length: 8,
                    require_uppercase: true,
                    require_lowercase: true,
                    require_digit: true,
                    require_special: true,
                  }
                }
                onSave={(policy) => updateSettingsMutation.mutate({ password_policy: policy })}
              />
            </div>
          </div>

          <div
            className="rounded-lg p-5"
            style={{
              background: 'var(--panel)',
              border: '1px solid var(--panel-edge)',
              boxShadow: 'var(--shadow-chamber)',
            }}
          >
            <div className="mb-4">
              <h3 className="font-display text-lg font-semibold" style={{ color: 'var(--text)' }}>
                {t('authSettings.sessionSettings')}
              </h3>
              <p className="text-sm mt-1" style={{ color: 'var(--text-dim)' }}>
                {t('authSettings.sessionSettingsDesc')}
              </p>
            </div>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <Label style={{ color: 'var(--text)' }}>{t('authSettings.sessionTimeout')}</Label>
                  <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
                    {t('authSettings.sessionTimeoutDesc')}
                  </p>
                </div>
                <Select
                  value={String(settings?.session_timeout_minutes || 480)}
                  onValueChange={(value) =>
                    updateSettingsMutation.mutate({ session_timeout_minutes: parseInt(value) })
                  }
                >
                  <SelectTrigger className="w-40">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="60">{t('authSettings.sessionTimeouts.1hour')}</SelectItem>
                    <SelectItem value="240">{t('authSettings.sessionTimeouts.4hours')}</SelectItem>
                    <SelectItem value="480">{t('authSettings.sessionTimeouts.8hours')}</SelectItem>
<SelectItem value="1440">{t('authSettings.sessionTimeouts.24hours')}</SelectItem>
                    <SelectItem value="10080">{t('authSettings.sessionTimeouts.1week')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <Label style={{ color: 'var(--text)' }}>
                    {t('authSettings.requireMFA')}
                  </Label>
                  <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
                    {t('authSettings.requireMFADesc')}
                  </p>
                </div>
                <Select
                  value={settings?.mfa_mode || 'optional'}
                  onValueChange={(value) =>
                    updateSettingsMutation.mutate({
                      mfa_mode: value as 'optional' | 'required' | 'enforced',
                    })
                  }
                >
                  <SelectTrigger className="w-40">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="optional">{t('authSettings.mfaModes.optional')}</SelectItem>
                    <SelectItem value="required">{t('authSettings.mfaModes.required')}</SelectItem>
                    <SelectItem value="enforced">{t('authSettings.mfaModes.enforced')}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>

          <div
            className="rounded-lg p-5"
            style={{
              background: 'var(--panel)',
              border: '1px solid var(--panel-edge)',
              boxShadow: 'var(--shadow-chamber)',
            }}
          >
            <div className="mb-4">
              <h3 className="font-display text-lg font-semibold" style={{ color: 'var(--text)' }}>
                {t('authSettings.loginMethods')}
              </h3>
              <p className="text-sm mt-1" style={{ color: 'var(--text-dim)' }}>
                {t('authSettings.loginMethodsDesc')}
              </p>
            </div>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <Label style={{ color: 'var(--text)' }}>{t('authSettings.passwordLogin')}</Label>
                  <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
                    {t('authSettings.passwordLoginDesc')}
                  </p>
                </div>
                <Switch
                  checked={settings?.allow_password_login ?? true}
                  onCheckedChange={(checked) =>
                    updateSettingsMutation.mutate({ allow_password_login: checked })
                  }
                />
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <Label style={{ color: 'var(--text)' }}>{t('authSettings.magicLink')}</Label>
                  <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
                    {t('authSettings.magicLinkDesc')}
                  </p>
                </div>
                <Switch
                  checked={settings?.allow_magic_link ?? true}
                  onCheckedChange={(checked) =>
                    updateSettingsMutation.mutate({ allow_magic_link: checked })
                  }
                />
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <Label style={{ color: 'var(--text)' }}>{t('authSettings.emailVerification')}</Label>
                  <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
                    {t('authSettings.emailVerificationDesc')}
                  </p>
                </div>
                <Switch
                  checked={settings?.require_email_verification ?? true}
                  onCheckedChange={(checked) =>
                    updateSettingsMutation.mutate({ require_email_verification: checked })
                  }
                />
              </div>
            </div>
          </div>
        </TabsContent>

        {/* OAuth Tab */}
        <TabsContent value="oauth" className="space-y-6">
          <div
            className="rounded-lg p-5"
            style={{
              background: 'var(--panel)',
              border: '1px solid var(--panel-edge)',
              boxShadow: 'var(--shadow-chamber)',
            }}
          >
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="font-display text-lg font-semibold" style={{ color: 'var(--text)' }}>
                  {t('authSettings.socialLogin')}
                </h3>
                <p className="text-sm mt-1" style={{ color: 'var(--text-dim)' }}>
                  {t('authSettings.socialLoginDesc')}
                </p>
              </div>
              <Button
                onClick={() => setShowOAuthForm(true)}
                style={{
                  background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                  color: 'var(--text-on-light)',
                  boxShadow: 'var(--shadow-btn-primary-rest)',
                }}
              >
                <Key className="w-4 h-4 mr-2" />
                {t('authSettings.addProvider')}
              </Button>
            </div>
            <div>
              {showOAuthForm ? (
                <OAuthProviderForm
                  provider={editingProvider || undefined}
                  onSave={(config) => oauthMutation.mutate(config)}
                  onCancel={() => {
                    setShowOAuthForm(false);
                    setEditingProvider(null);
                  }}
                />
              ) : (
                <div className="grid md:grid-cols-2 gap-4">
                  {oauthProviders.map((provider) => (
                    <OAuthProviderCard
                      key={provider.provider}
                      provider={provider}
                      onEdit={() => {
                        setEditingProvider({
                          provider: provider.provider,
                          client_id: provider.client_id,
                          client_secret: '',
                          enabled: provider.enabled,
                          callback_url: provider.callback_url,
                        });
                        setShowOAuthForm(true);
                      }}
                      onDelete={() => setConfirmDeleteProvider(provider.provider)}
                    />
                  ))}
                  {oauthProviders.length === 0 && (
                    <div
                      className="col-span-2 text-center py-8"
                      style={{ color: 'var(--text-dim)' }}
                    >
                      {t('authSettings.noProviders')}
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
        </TabsContent>

        {/* Team Tab */}
        <TabsContent value="team" className="space-y-6">
          <div
            className="rounded-lg p-5"
            style={{
              background: 'var(--panel)',
              border: '1px solid var(--panel-edge)',
              boxShadow: 'var(--shadow-chamber)',
            }}
          >
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="font-display text-lg font-semibold" style={{ color: 'var(--text)' }}>
                  {t('authSettings.teamMembers')}
                </h3>
                <p className="text-sm mt-1" style={{ color: 'var(--text-dim)' }}>
                  {t('authSettings.teamMembersDesc')}
                </p>
              </div>
              {canManageTeam && (
                <Button
                  onClick={() => setShowInviteForm(true)}
                  style={{
                    background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                    color: 'var(--text-on-light)',
                    boxShadow: 'var(--shadow-btn-primary-rest)',
                  }}
                >
                  <UserPlus className="w-4 h-4 mr-2" />
                  {t('authSettings.invite')}
                </Button>
              )}
            </div>
            <div>
              {showInviteForm && (
                <div
                  className="mb-6 p-4 rounded-lg space-y-4"
                  style={{
                    background: 'var(--panel-raised)',
                    border: '1px solid var(--panel-edge)',
                  }}
                >
                  <div className="flex gap-4">
                    <div className="flex-1">
                      <Label style={{ color: 'var(--text)' }}>{t('authSettings.inviteEmail')}</Label>
                      <Input
                        type="email"
                        value={inviteEmail}
                        onChange={(e) => setInviteEmail(e.target.value)}
                        placeholder="colleague@company.com"
                      />
                    </div>
                    <div className="w-40">
                      <Label style={{ color: 'var(--text)' }}>{t('authSettings.inviteRole')}</Label>
                      <Select value={inviteRole} onValueChange={setInviteRole}>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="team_admin">{t('authSettings.roles.admin')}</SelectItem>
                          <SelectItem value="team_member">{t('authSettings.roles.member')}</SelectItem>
                          <SelectItem value="team_viewer">{t('authSettings.roles.viewer')}</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                  <div className="flex justify-end gap-2">
                    <Button
                      variant="outline"
                      onClick={() => {
                        setShowInviteForm(false);
                        setInviteEmail('');
                      }}
                      style={{ borderColor: 'var(--steel)', color: 'var(--text)' }}
                    >
                      {t('billingSettings.cancelBtn')}
                    </Button>
                    <Button
                      onClick={() =>
                        inviteMutation.mutate({ email: inviteEmail, role: inviteRole })
                      }
                      disabled={!inviteEmail}
                      style={{
                        background: 'linear-gradient(180deg, #ffffff, #d8dee2)',
                        color: 'var(--text-on-light)',
                        boxShadow: 'var(--shadow-btn-primary-rest)',
                      }}
                    >
                      {t('authSettings.sendInvite')}
                    </Button>
                  </div>
                </div>
              )}

              <div>
                {members.map((member) => (
                  <MemberRow
                    key={member.id}
                    member={member}
                    onRoleChange={(role) =>
                      updateRoleMutation.mutate({ userId: member.user_id, role })
                    }
                    onRemove={() => setConfirmRemoveMember(member.user_id)}
                    canManage={canManageTeam}
                  />
                ))}
                {members.length === 0 && (
                  <div className="text-center py-8" style={{ color: 'var(--text-dim)' }}>
                    {t('authSettings.noMembers')}
                  </div>
                )}
              </div>
            </div>
          </div>

          <div
            className="rounded-lg p-5"
            style={{
              background: 'var(--panel)',
              border: '1px solid var(--panel-edge)',
              boxShadow: 'var(--shadow-chamber)',
            }}
          >
            <div className="mb-4">
              <h3 className="font-display text-lg font-semibold" style={{ color: 'var(--text)' }}>
                {t('authSettings.pendingInvitations')}
              </h3>
              <p className="text-sm mt-1" style={{ color: 'var(--text-dim)' }}>
                {t('authSettings.pendingInvitationsDesc')}
              </p>
            </div>
            <div>
              <div>
                {pendingInvites.map((invite) => (
                  <InviteRow
                    key={invite.id}
                    invite={invite}
                    onCopy={() => {
                      navigator.clipboard.writeText(
                        `${window.location.origin}/invite/${invite.code}`
                      );
                      toast.success(t('authSettings.toastLinkCopied'));
                    }}
                    onRevoke={() => setConfirmRevokeInvite(invite.code)}
                  />
                ))}
                {pendingInvites.length === 0 && (
                  <div className="text-center py-8" style={{ color: 'var(--text-dim)' }}>
                    {t('authSettings.noPendingInvites')}
                  </div>
                )}
              </div>
            </div>
          </div>
        </TabsContent>
      </Tabs>

      <AlertDialog open={!!confirmRemoveMember} onOpenChange={() => setConfirmRemoveMember(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('authSettings.removeMemberTitle')}</AlertDialogTitle>
            <AlertDialogDescription>
              {t('authSettings.removeMemberDesc')}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t('billingSettings.cancelBtn')}</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                if (confirmRemoveMember) {
                  removeMemberMutation.mutate(confirmRemoveMember);
                  setConfirmRemoveMember(null);
                }
              }}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {t('authSettings.remove')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog open={!!confirmDeleteProvider} onOpenChange={() => setConfirmDeleteProvider(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('authSettings.deleteProviderTitle')}</AlertDialogTitle>
            <AlertDialogDescription>
              {t('authSettings.deleteProviderDesc')}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t('billingSettings.cancelBtn')}</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                if (confirmDeleteProvider) {
                  deleteOAuthMutation.mutate(confirmDeleteProvider);
                  setConfirmDeleteProvider(null);
                }
              }}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {t('authSettings.delete')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <AlertDialog open={!!confirmRevokeInvite} onOpenChange={() => setConfirmRevokeInvite(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('authSettings.revokeInviteTitle')}</AlertDialogTitle>
            <AlertDialogDescription>
              {t('authSettings.revokeInviteDesc')}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t('billingSettings.cancelBtn')}</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                if (confirmRevokeInvite) {
                  revokeInviteMutation.mutate(confirmRevokeInvite);
                  setConfirmRevokeInvite(null);
                }
              }}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {t('authSettings.revoke')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
