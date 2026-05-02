import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useAuthStore } from '@/stores/authStore';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Building, Copy, Eye, EyeOff, Key, Shield, Trash2, UserPlus } from 'lucide-react';
import { useState } from 'react';
import { toast } from 'react-hot-toast';
import { useTranslation } from 'react-i18next';
import { apiClient } from '@/api/client';
import type { TenantAuthSettings, TenantOAuthProvider, TenantMembership, TenantInviteCode, PasswordPolicy } from '@/types';

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

async function updateAuthSettings(updates: Partial<TenantAuthSettings>): Promise<TenantAuthSettings> {
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
  const data = await apiClient.post<{ invite: TenantInviteCode }>('/v1/auth/members/invite', { email, role });
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
  const [localPolicy, setLocalPolicy] = useState(policy);
  const [saving, setSaving] = useState(false);

  const handleSave = async () => {
    setSaving(true);
    try {
      await onSave(localPolicy);
      toast.success('Password policy updated');
    } catch {
      toast.error('Failed to update policy');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div>
          <Label htmlFor="min_length">Minimum Length</Label>
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
          <Label htmlFor="require_uppercase">Require uppercase letter</Label>
        </div>
        <div className="flex items-center gap-3">
          <Switch
            id="require_lowercase"
            checked={localPolicy.require_lowercase}
            onCheckedChange={(checked) =>
              setLocalPolicy({ ...localPolicy, require_lowercase: checked })
            }
          />
          <Label htmlFor="require_lowercase">Require lowercase letter</Label>
        </div>
        <div className="flex items-center gap-3">
          <Switch
            id="require_digit"
            checked={localPolicy.require_digit}
            onCheckedChange={(checked) =>
              setLocalPolicy({ ...localPolicy, require_digit: checked })
            }
          />
          <Label htmlFor="require_digit">Require digit</Label>
        </div>
        <div className="flex items-center gap-3">
          <Switch
            id="require_special"
            checked={localPolicy.require_special}
            onCheckedChange={(checked) =>
              setLocalPolicy({ ...localPolicy, require_special: checked })
            }
          />
          <Label htmlFor="require_special">Require special character</Label>
        </div>
      </div>

      <Button onClick={handleSave} disabled={saving}>
        {saving ? 'Saving...' : 'Save Policy'}
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
              <CardDescription>{provider.enabled ? 'Connected' : 'Disabled'}</CardDescription>
            </div>
          </div>
          <Badge variant={provider.enabled ? 'default' : 'secondary'}>
            {provider.enabled ? 'Active' : 'Inactive'}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={onEdit}>
            Edit
          </Button>
          <Button variant="ghost" size="sm" onClick={onDelete} className="text-destructive">
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
      toast.error('Client ID and Secret are required');
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
        <Label htmlFor="client_id">Client ID</Label>
        <Input
          id="client_id"
          value={formData.client_id}
          onChange={(e) => setFormData({ ...formData, client_id: e.target.value })}
          placeholder="Enter OAuth client ID"
        />
      </div>

      <div>
        <Label htmlFor="client_secret">Client Secret</Label>
        <div className="relative">
          <Input
            id="client_secret"
            type={showSecret ? 'text' : 'password'}
            value={formData.client_secret}
            onChange={(e) => setFormData({ ...formData, client_secret: e.target.value })}
            placeholder={provider ? '••••••••' : 'Enter OAuth client secret'}
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
          <p className="text-xs text-text-muted mt-1">Leave blank to keep existing secret</p>
        )}
      </div>

      <div className="flex items-center gap-3">
        <Switch
          id="enabled"
          checked={formData.enabled}
          onCheckedChange={(checked) => setFormData({ ...formData, enabled: checked })}
        />
        <Label htmlFor="enabled">Enable this provider</Label>
      </div>

      <div className="flex justify-end gap-2">
        <Button variant="outline" onClick={onCancel}>
          Cancel
        </Button>
        <Button onClick={handleSave}>Save Provider</Button>
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
              <SelectItem value="team_owner">Owner</SelectItem>
              <SelectItem value="team_admin">Admin</SelectItem>
              <SelectItem value="team_member">Member</SelectItem>
              <SelectItem value="team_viewer">Viewer</SelectItem>
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
  const expiresAt = new Date(invite.expires_at);
  const isExpired = expiresAt < new Date();

  return (
    <div className="flex items-center justify-between py-3 border-b border-border last:border-0">
      <div>
        <p className="font-medium text-text-primary">{invite.email}</p>
        <p className="text-sm text-text-muted">
          {isExpired ? 'Expired' : `Expires ${expiresAt.toLocaleDateString()}`} ·{' '}
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
  });

  const oauthMutation = useMutation({
    mutationFn: configureOAuthProvider,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['oauth-providers'] });
      setShowOAuthForm(false);
      setEditingProvider(null);
      toast.success('OAuth provider saved');
    },
  });

  const deleteOAuthMutation = useMutation({
    mutationFn: deleteOAuthProvider,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['oauth-providers'] });
      toast.success('Provider removed');
    },
  });

  const inviteMutation = useMutation({
    mutationFn: ({ email, role }: { email: string; role: string }) => inviteMember(email, role),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pending-invites'] });
      setShowInviteForm(false);
      setInviteEmail('');
      toast.success('Invitation sent');
    },
  });

  const updateRoleMutation = useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: string }) => updateMemberRole(userId, role),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-members'] });
      toast.success('Role updated');
    },
  });

  const removeMemberMutation = useMutation({
    mutationFn: (userId: string) => removeMember(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-members'] });
      toast.success('Member removed');
    },
  });

  const revokeInviteMutation = useMutation({
    mutationFn: revokeInvite,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['pending-invites'] });
      toast.success('Invitation revoked');
    },
  });

  const canManageTeam = members.some(
    (m) => m.user_id === user?.id && (m.role === 'team_owner' || m.role === 'team_admin')
  );

  if (settingsLoading) {
    return <div className="animate-pulse space-y-4">{/* Loading skeleton */}</div>;
  }

  return (
    <Tabs defaultValue="security" className="space-y-6">
      <TabsList>
        <TabsTrigger value="security">
          <Shield className="w-4 h-4 mr-2" />
          Security
        </TabsTrigger>
        <TabsTrigger value="oauth">
          <Key className="w-4 h-4 mr-2" />
          OAuth
        </TabsTrigger>
        <TabsTrigger value="team">
          <Building className="w-4 h-4 mr-2" />
          Team
        </TabsTrigger>
      </TabsList>

      {/* Security Tab */}
      <TabsContent value="security" className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Password Policy</CardTitle>
            <CardDescription>Configure minimum password requirements</CardDescription>
          </CardHeader>
          <CardContent>
            <PasswordPolicyEditor
              policy={settings?.password_policy || {
                min_length: 8,
                require_uppercase: true,
                require_lowercase: true,
                require_digit: true,
                require_special: true,
              }}
              onSave={(policy) => updateSettingsMutation.mutate({ password_policy: policy })}
            />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Session Settings</CardTitle>
            <CardDescription>Configure session timeout and security</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <Label>Session Timeout</Label>
                <p className="text-sm text-text-muted">
                  How long until a session expires
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
                  <SelectItem value="60">1 hour</SelectItem>
                  <SelectItem value="240">4 hours</SelectItem>
                  <SelectItem value="480">8 hours</SelectItem>
                  <SelectItem value="1440">24 hours</SelectItem>
                  <SelectItem value="10080">1 week</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-center justify-between">
              <div>
                <Label>Require Multi-Factor Authentication</Label>
                <p className="text-sm text-text-muted">
                  Require MFA for all users in your organization
                </p>
              </div>
              <Select
                value={settings?.mfa_mode || 'optional'}
                onValueChange={(value) =>
                  updateSettingsMutation.mutate({ mfa_mode: value as 'optional' | 'required' | 'enforced' })
                }
              >
                <SelectTrigger className="w-40">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="optional">Optional</SelectItem>
                  <SelectItem value="required">Required</SelectItem>
                  <SelectItem value="enforced">Enforced</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Login Methods</CardTitle>
            <CardDescription>Configure how users can sign in</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <Label>Password Login</Label>
                <p className="text-sm text-text-muted">
                  Allow users to sign in with email and password
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
                <Label>Magic Link</Label>
                <p className="text-sm text-text-muted">
                  Allow passwordless sign-in via email link
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
                <Label>Email Verification</Label>
                <p className="text-sm text-text-muted">
                  Require users to verify their email
                </p>
              </div>
              <Switch
                checked={settings?.require_email_verification ?? true}
                onCheckedChange={(checked) =>
                  updateSettingsMutation.mutate({ require_email_verification: checked })
                }
              />
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      {/* OAuth Tab */}
      <TabsContent value="oauth" className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Social Login</CardTitle>
                <CardDescription>
                  Configure OAuth providers for social login
                </CardDescription>
              </div>
              <Button onClick={() => setShowOAuthForm(true)}>
                <Key className="w-4 h-4 mr-2" />
                Add Provider
              </Button>
            </div>
          </CardHeader>
          <CardContent>
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
                    onDelete={() => deleteOAuthMutation.mutate(provider.provider)}
                  />
                ))}
                {oauthProviders.length === 0 && (
                  <div className="col-span-2 text-center py-8 text-text-muted">
                    No OAuth providers configured. Click "Add Provider" to get started.
                  </div>
                )}
              </div>
            )}
          </CardContent>
        </Card>
      </TabsContent>

      {/* Team Tab */}
      <TabsContent value="team" className="space-y-6">
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>Team Members</CardTitle>
                <CardDescription>Manage who has access to your organization</CardDescription>
              </div>
              {canManageTeam && (
                <Button onClick={() => setShowInviteForm(true)}>
                  <UserPlus className="w-4 h-4 mr-2" />
                  Invite
                </Button>
              )}
            </div>
          </CardHeader>
          <CardContent>
            {showInviteForm && (
              <div className="mb-6 p-4 bg-bg-secondary rounded-lg space-y-4">
                <div className="flex gap-4">
                  <div className="flex-1">
                    <Label>Email</Label>
                    <Input
                      type="email"
                      value={inviteEmail}
                      onChange={(e) => setInviteEmail(e.target.value)}
                      placeholder="colleague@company.com"
                    />
                  </div>
                  <div className="w-40">
                    <Label>Role</Label>
                    <Select value={inviteRole} onValueChange={setInviteRole}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="team_admin">Admin</SelectItem>
                        <SelectItem value="team_member">Member</SelectItem>
                        <SelectItem value="team_viewer">Viewer</SelectItem>
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
                  >
                    Cancel
                  </Button>
                  <Button
                    onClick={() => inviteMutation.mutate({ email: inviteEmail, role: inviteRole })}
                    disabled={!inviteEmail}
                  >
                    Send Invite
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
                  onRemove={() => removeMemberMutation.mutate(member.user_id)}
                  canManage={canManageTeam}
                />
              ))}
              {members.length === 0 && (
                <div className="text-center py-8 text-text-muted">No team members found</div>
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Pending Invitations</CardTitle>
            <CardDescription>Invitations that haven't been accepted yet</CardDescription>
          </CardHeader>
          <CardContent>
            <div>
              {pendingInvites.map((invite) => (
                <InviteRow
                  key={invite.id}
                  invite={invite}
                  onCopy={() => {
                    navigator.clipboard.writeText(`${window.location.origin}/invite/${invite.code}`);
                    toast.success('Link copied');
                  }}
                  onRevoke={() => revokeInviteMutation.mutate(invite.code)}
                />
              ))}
              {pendingInvites.length === 0 && (
                <div className="text-center py-8 text-text-muted">No pending invitations</div>
              )}
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  );
}
