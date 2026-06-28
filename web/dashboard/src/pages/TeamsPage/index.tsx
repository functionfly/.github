import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import {
  useMutation, useQuery, useQueryClient,
} from '@tanstack/react-query';
import { AnimatePresence, motion } from 'framer-motion';
import {
  ChevronRight, Clock, Crown, Edit3, Eye, LayoutGrid, List,
  Mail, MoreVertical, Plus, Search, Settings, Shield, Trash2,
  User, UserPlus, Users, X,
} from 'lucide-react';
import { useState } from 'react';
import { usePageTitle } from '@/hooks';
import { teamsApi, type Team, type TeamMember } from '@/api/teams';
import { Chamber, CornerBrace, PageGrid, SealedButton, FrameButton, StatusPill, Modal, Input as ScInput } from '@/components/containment';
import './styles.css';
import { useAuthStore } from '@/stores/authStore';
import { format, formatDistanceToNow } from 'date-fns';
import { toast } from 'sonner';

const roleIcons: Record<string, typeof Crown> = { owner: Crown, admin: Shield, member: User, viewer: Eye };
const roleLabels: Record<string, string> = { owner: 'Owner', admin: 'Admin', member: 'Member', viewer: 'Viewer' };

const getInitials = (name?: string, email?: string) => {
  if (name) return name.split(' ').map((n) => n[0]).join('').toUpperCase().slice(0, 2);
  if (email) return email[0].toUpperCase();
  return '?';
};

export function TeamsPage() {
  usePageTitle('Teams');
  const queryClient = useQueryClient();
  const { user } = useAuthStore();
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false);
  const [isInviteDialogOpen, setIsInviteDialogOpen] = useState(false);
  const [selectedTeam, setSelectedTeam] = useState<Team | null>(null);
  const [newTeamName, setNewTeamName] = useState('');
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteRole, setInviteRole] = useState<'admin' | 'member' | 'viewer'>('member');
  const [searchQuery, setSearchQuery] = useState('');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [activeTab, setActiveTab] = useState('members');

  const { data: teamsData, isLoading } = useQuery({ queryKey: ['teams'], queryFn: () => teamsApi.list() });
  const teams = teamsData?.teams ?? [];
  const filteredTeams = teams.filter((t) => t.name.toLowerCase().includes(searchQuery.toLowerCase()));

  const { data: membersData, isLoading: membersLoading } = useQuery({
    queryKey: ['team-members', selectedTeam?.id], queryFn: () => teamsApi.listMembers(selectedTeam!.id), enabled: !!selectedTeam,
  });
  const { data: invitesData, isLoading: invitesLoading } = useQuery({
    queryKey: ['team-invites', selectedTeam?.id], queryFn: () => teamsApi.listInvites(selectedTeam!.id), enabled: !!selectedTeam,
  });
  const members = membersData?.members ?? [];
  const invites = invitesData?.invites ?? [];

  const createTeamMutation = useMutation({
    mutationFn: (name: string) => teamsApi.create({ name }),
    onSuccess: (data) => {
      queryClient.setQueryData<{ teams: Team[] }>(['teams'], (old) => ({ teams: old ? [...old.teams, { ...data.team, members: [{ id: '', user_id: user?.id ?? '', team_id: data.team.id, role: 'owner' as const, created_at: new Date().toISOString(), user: user ? { id: user.id, email: user.email, name: user.name, username: user.username } : undefined }] }] : [{ ...data.team }] }));
      setIsCreateDialogOpen(false); setNewTeamName(''); toast.success('Team created successfully');
    },
    onError: (e: Error) => toast.error(e.message || 'Failed to create team'),
  });

  const deleteTeamMutation = useMutation({
    mutationFn: (teamId: string) => teamsApi.delete(teamId),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['teams'] }); if (selectedTeam) setSelectedTeam(null); toast.success('Team deleted successfully'); },
    onError: (e: Error) => toast.error(e.message || 'Failed to delete team'),
  });

  const inviteMemberMutation = useMutation({
    mutationFn: ({ teamId, email, role }: { teamId: string; email: string; role: string }) => teamsApi.addMember(teamId, { email, role: role as 'admin' | 'member' | 'viewer' }),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['team-members', selectedTeam?.id] }); queryClient.invalidateQueries({ queryKey: ['team-invites', selectedTeam?.id] }); setIsInviteDialogOpen(false); setInviteEmail(''); setInviteRole('member'); toast.success('Invitation sent successfully'); },
    onError: (e: Error) => toast.error(e.message || 'Failed to send invitation'),
  });

  const updateMemberMutation = useMutation({
    mutationFn: ({ teamId, memberId, role }: { teamId: string; memberId: string; role: string }) => teamsApi.updateMember(teamId, memberId, { role: role as 'admin' | 'member' | 'viewer' }),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['team-members', selectedTeam?.id] }); toast.success('Member role updated'); },
    onError: (e: Error) => toast.error(e.message || 'Failed to update role'),
  });

  const removeMemberMutation = useMutation({
    mutationFn: ({ teamId, memberId }: { teamId: string; memberId: string }) => teamsApi.removeMember(teamId, memberId),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['team-members', selectedTeam?.id] }); toast.success('Member removed from team'); },
    onError: (e: Error) => toast.error(e.message || 'Failed to remove member'),
  });

  const cancelInviteMutation = useMutation({
    mutationFn: ({ teamId, inviteId }: { teamId: string; inviteId: string }) => teamsApi.cancelInvite(teamId, inviteId),
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ['team-invites', selectedTeam?.id] }); toast.success('Invitation cancelled'); },
    onError: (e: Error) => toast.error(e.message || 'Failed to cancel invitation'),
  });

  const isCurrentUserOwner = (m: TeamMember) => m.user_id === user?.id && m.role === 'owner';
  const getUserRoleInTeam = (t: Team) => { if (t.owner_id === user?.id) return 'owner'; return t.members?.find((m) => m.user_id === user?.id)?.role || 'member'; };

  const TabButton = ({ value, icon: Icon, label }: { value: string; icon: typeof Users; label: string }) => (
    <button className="team-tab" data-state={activeTab === value ? 'active' : 'inactive'} onClick={() => setActiveTab(value)}>
      <Icon style={{ width: 14, height: 14 }} />{label}
    </button>
  );

  return (
    <div style={{ maxWidth: 1180, margin: '0 auto', padding: 'var(--space-7)', display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }} className="animate-fade-in">
      <PageGrid />

      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 'var(--space-4)' }}>
        <div>
          <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 36, fontWeight: 700, letterSpacing: '-0.005em', color: 'var(--text)' }}>Teams</h1>
          <p style={{ fontSize: 14, color: 'var(--text-dim)', marginTop: 'var(--space-2)' }}>Manage your teams, members, and permissions</p>
        </div>
        {teams.length > 0 && (
          <SealedButton onClick={() => setIsCreateDialogOpen(true)} iconLeft={<Plus style={{ width: 14, height: 14 }} />}>Create Team</SealedButton>
        )}
      </div>

      {/* Search + View Toggle */}
      {teams.length > 0 && (
        <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-4)' }}>
          <div style={{ position: 'relative', flex: 1, maxWidth: 400 }}>
            <Search style={{ position: 'absolute', left: 12, top: '50%', transform: 'translateY(-50%)', width: 14, height: 14, color: 'var(--text-faint)', pointerEvents: 'none' }} />
            <input type="text" placeholder="Search teams..." value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} className="team-input" style={{ paddingLeft: 36 }} />
            {searchQuery && <button onClick={() => setSearchQuery('')} style={{ position: 'absolute', right: 12, top: '50%', transform: 'translateY(-50%)', background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-faint)' }}><X style={{ width: 14, height: 14 }} /></button>}
          </div>
          <div style={{ display: 'flex', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)', overflow: 'hidden', marginLeft: 'auto' }}>
            {([['grid', LayoutGrid], ['list', List]] as const).map(([mode, Icon]) => (
              <button key={mode} onClick={() => setViewMode(mode as 'grid' | 'list')} style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', width: 36, height: 36, background: viewMode === mode ? 'var(--panel-raised)' : 'transparent', border: 'none', cursor: 'pointer', color: viewMode === mode ? 'var(--status-ok)' : 'var(--text-faint)', transition: 'all var(--duration-fast) var(--ease-out)' }}>
                <Icon style={{ width: 14, height: 14 }} />
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Teams Grid/List */}
      {isLoading ? (
        <div className="team-loading-container"><div className="team-loading-spinner" /></div>
      ) : teams.length === 0 ? (
        <Chamber>
          <div style={{ textAlign: 'center', padding: 'var(--space-8) 0' }}>
            <div className="team-empty-icon"><Users style={{ width: 32, height: 32 }} /></div>
            <h3 className="team-empty-title">No teams yet</h3>
            <p className="team-empty-description">Create your first team to start collaborating with your colleagues.</p>
            <SealedButton onClick={() => setIsCreateDialogOpen(true)} iconLeft={<Plus style={{ width: 14, height: 14 }} />}>Create Team</SealedButton>
          </div>
        </Chamber>
      ) : (
        <div style={{ display: 'grid', gridTemplateColumns: viewMode === 'grid' ? 'repeat(auto-fill, minmax(320px, 1fr))' : '1fr', gap: 'var(--space-4)' }}>
          <AnimatePresence mode="popLayout">
            {filteredTeams.map((team, index) => {
              const userRole = getUserRoleInTeam(team);
              const RoleIcon = roleIcons[userRole] || User;
              const isSelected = selectedTeam?.id === team.id;
              return (
                <motion.div key={team.id} initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: index * 0.05 }}>
                  <div className={`team-card ${isSelected ? 'ring-2' : ''}`} style={{ cursor: 'pointer', padding: 'var(--space-5)', borderColor: isSelected ? 'var(--status-ok)' : undefined }} onClick={() => setSelectedTeam(team)}>
                    <CornerBrace position="tl" /><CornerBrace position="br" />
                    <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: 'var(--space-3)' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
                        <div className="team-avatar" style={{ width: viewMode === 'list' ? 40 : 48, height: viewMode === 'list' ? 40 : 48 }}>
                          <Users className="team-avatar-icon" style={{ width: viewMode === 'list' ? 18 : 22, height: viewMode === 'list' ? 18 : 22 }} />
                        </div>
                        <div>
                          <h3 style={{ fontFamily: 'var(--font-display)', fontSize: 16, fontWeight: 600, color: 'var(--text)' }}>{team.name}</h3>
                          <p style={{ fontSize: 13, color: 'var(--text-dim)' }}>{team.members?.length ?? 0} member{(team.members?.length ?? 0) !== 1 ? 's' : ''}</p>
                        </div>
                      </div>
                      <span className={`role-badge role-badge-${userRole}`}><RoleIcon style={{ width: 10, height: 10 }} />{roleLabels[userRole]}</span>
                    </div>
                    {viewMode === 'grid' && (
                      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', fontSize: 11, color: 'var(--text-faint)' }}>
                        <span>Created {formatDistanceToNow(new Date(team.created_at), { addSuffix: true })}</span>
                        <ChevronRight style={{ width: 14, height: 14, color: isSelected ? 'var(--status-ok)' : 'var(--text-faint)' }} />
                      </div>
                    )}
                  </div>
                </motion.div>
              );
            })}
          </AnimatePresence>
        </div>
      )}

      {/* Team Detail Panel */}
      <AnimatePresence mode="wait">
        {selectedTeam && (
          <motion.div key={selectedTeam.id} initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0, y: -20 }} transition={{ duration: 0.2 }}>
            <Chamber>
              <CornerBrace position="tl" /><CornerBrace position="br" />
              <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: 'var(--space-5)' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-4)' }}>
                  <div className="team-avatar team-avatar-lg"><Users className="team-avatar-icon" style={{ width: 24, height: 24 }} /></div>
                  <div>
                    <h2 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 500, color: 'var(--text)' }}>{selectedTeam.name}</h2>
                    <p style={{ fontSize: 13, color: 'var(--text-dim)' }}>Manage team members, invites, and settings</p>
                  </div>
                </div>
                <div style={{ display: 'flex', gap: 'var(--space-2)' }}>
                  <FrameButton size="sm" onClick={() => setSelectedTeam(null)} iconLeft={<X style={{ width: 14, height: 14 }} />}>Close</FrameButton>
                  <SealedButton size="sm" onClick={() => setIsInviteDialogOpen(true)} iconLeft={<UserPlus style={{ width: 14, height: 14 }} />}>Invite</SealedButton>
                </div>
              </div>

              {/* Tabs */}
              <div className="team-tabs">
                <TabButton value="members" icon={Users} label={`Members (${members.length})`} />
                <TabButton value="invites" icon={Mail} label={`Invites (${invites.length})`} />
                <TabButton value="settings" icon={Settings} label="Settings" />
              </div>

              <div style={{ padding: 'var(--space-5)' }}>
                {/* Members Tab */}
                {activeTab === 'members' && (
                  membersLoading ? <div className="team-loading-container"><div className="team-loading-spinner" /></div>
                  : members.length === 0 ? (
                    <div style={{ textAlign: 'center', padding: 'var(--space-7) 0' }}>
                      <div className="team-empty-icon" style={{ margin: '0 auto var(--space-4)' }}><Users style={{ width: 40, height: 40 }} /></div>
                      <h3 className="team-empty-title">No members yet</h3>
                      <p style={{ color: 'var(--text-dim)', marginBottom: 'var(--space-5)' }}>Invite your first team member to start collaborating.</p>
                      <SealedButton onClick={() => setIsInviteDialogOpen(true)} iconLeft={<UserPlus style={{ width: 14, height: 14 }} />}>Invite Member</SealedButton>
                    </div>
                  ) : (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
                      {members.map((member) => {
                        const RoleIcon = roleIcons[member.role];
                        return (
                          <div key={member.id} className="member-list-item">
                            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-4)' }}>
                              <Avatar className="member-avatar"><AvatarFallback className="member-avatar-fallback">{getInitials(member.user?.name, member.user?.email)}</AvatarFallback></Avatar>
                              <div>
                                <div className="member-name" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)' }}>
                                  {member.user?.name || member.user?.username || member.user?.email}
                                  {isCurrentUserOwner(member) && <span className="role-badge role-badge-member" style={{ fontSize: 9 }}>You</span>}
                                </div>
                                <div className="member-email">{member.user?.email}</div>
                              </div>
                            </div>
                            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
                              <span className={`role-badge role-badge-${member.role}`}><RoleIcon style={{ width: 10, height: 10 }} />{roleLabels[member.role]}</span>
                              {!isCurrentUserOwner(member) && (
                                <div style={{ display: 'flex', gap: 'var(--space-1)' }}>
                                  {(['admin', 'member', 'viewer'] as const).map((r) => (
                                    <button key={r} onClick={() => updateMemberMutation.mutate({ teamId: selectedTeam.id, memberId: member.id, role: r })} style={{ fontSize: 9, padding: '2px 6px', borderRadius: 'var(--radius-sm)', border: member.role === r ? '1px solid var(--status-ok)' : '1px solid var(--panel-edge)', background: member.role === r ? 'rgba(143,255,208,0.06)' : 'transparent', color: member.role === r ? 'var(--status-ok)' : 'var(--text-faint)', cursor: 'pointer', fontFamily: 'var(--font-mono)', textTransform: 'uppercase', letterSpacing: '0.04em' }}>
                                      {r}
                                    </button>
                                  ))}
                                  <button onClick={() => { if (confirm(`Remove ${member.user?.email} from this team?`)) removeMemberMutation.mutate({ teamId: selectedTeam.id, memberId: member.id }); }} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--status-revoked)', padding: 4 }} title="Remove"><Trash2 style={{ width: 12, height: 12 }} /></button>
                                </div>
                              )}
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  )
                )}

                {/* Invites Tab */}
                {activeTab === 'invites' && (
                  invitesLoading ? <div className="team-loading-container"><div className="team-loading-spinner" /></div>
                  : invites.length === 0 ? (
                    <div style={{ textAlign: 'center', padding: 'var(--space-7) 0' }}>
                      <div className="team-empty-icon" style={{ margin: '0 auto var(--space-4)' }}><Mail style={{ width: 40, height: 40 }} /></div>
                      <h3 className="team-empty-title">No pending invites</h3>
                      <p style={{ color: 'var(--text-dim)', marginBottom: 'var(--space-5)' }}>Invite team members to see pending invitations here.</p>
                      <SealedButton onClick={() => setIsInviteDialogOpen(true)} iconLeft={<UserPlus style={{ width: 14, height: 14 }} />}>Invite Member</SealedButton>
                    </div>
                  ) : (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-2)' }}>
                      {invites.map((invite) => (
                        <div key={invite.id} className="member-list-item">
                          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-4)' }}>
                            <div style={{ width: 40, height: 40, borderRadius: 'var(--radius)', background: 'rgba(232,196,104,0.08)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                              <Mail style={{ width: 18, height: 18, color: 'var(--status-pending)' }} />
                            </div>
                            <div>
                              <div className="member-name">{invite.email}</div>
                              <div className="member-email" style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-1)' }}>
                                <Clock style={{ width: 11, height: 11 }} />
                                Invited {formatDistanceToNow(new Date(invite.created_at), { addSuffix: true })} · Expires {format(new Date(invite.expires_at), 'MMM d')}
                              </div>
                            </div>
                          </div>
                          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-3)' }}>
                            <StatusPill status="pending" label="Pending" />
                            <button onClick={() => cancelInviteMutation.mutate({ teamId: selectedTeam.id, inviteId: invite.id })} disabled={cancelInviteMutation.isPending} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-faint)', padding: 4 }}><X style={{ width: 14, height: 14 }} /></button>
                          </div>
                        </div>
                      ))}
                    </div>
                  )
                )}

                {/* Settings Tab */}
                {activeTab === 'settings' && (
                  <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: 'var(--space-5)' }}>
                    <div className="settings-card">
                      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', marginBottom: 'var(--space-3)' }}>
                        <Shield style={{ width: 18, height: 18, color: 'var(--status-ok)' }} />
                        <h3 className="settings-card-title">Team Permissions</h3>
                      </div>
                      <p className="settings-card-description" style={{ marginBottom: 'var(--space-4)' }}>Understanding roles and access levels</p>
                      <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
                        {[
                          { role: 'owner', desc: 'Full control — manage members, settings, billing, and delete team', color: 'var(--status-pending)' },
                          { role: 'admin', desc: 'Can manage members and resources, but cannot delete team', color: 'var(--foil-a)' },
                          { role: 'member', desc: 'Can create and manage functions and resources', color: 'var(--status-ok)' },
                          { role: 'viewer', desc: 'Read-only access to view team resources and activity', color: 'var(--text-dim)' },
                        ].map(({ role, desc, color }) => {
                          const RI = roleIcons[role];
                          return (
                            <div key={role} style={{ display: 'flex', alignItems: 'flex-start', gap: 'var(--space-3)' }}>
                              <div style={{ width: 32, height: 32, borderRadius: 'var(--radius)', background: `${color}11`, border: `1px solid ${color}33`, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                                <RI style={{ width: 14, height: 14, color }} />
                              </div>
                              <div>
                                <div style={{ fontWeight: 600, fontSize: 13, color: 'var(--text)', textTransform: 'capitalize' }}>{role}</div>
                                <div style={{ fontSize: 13, color: 'var(--text-dim)' }}>{desc}</div>
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    </div>

                    <div className="settings-card">
                      <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', marginBottom: 'var(--space-3)' }}>
                        <Settings style={{ width: 18, height: 18, color: 'var(--status-ok)' }} />
                        <h3 className="settings-card-title">Team Settings</h3>
                      </div>
                      <p className="settings-card-description" style={{ marginBottom: 'var(--space-4)' }}>Manage your team configuration</p>
                      <div style={{ padding: 'var(--space-4)', background: 'var(--panel)', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)', marginBottom: 'var(--space-3)' }}>
                        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 'var(--space-2)' }}>
                          <span style={{ fontWeight: 500, color: 'var(--text)' }}>Team Name</span>
                          <button style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--status-ok)', display: 'flex', alignItems: 'center', gap: 4, fontSize: 12 }}><Edit3 style={{ width: 12, height: 12 }} />Edit</button>
                        </div>
                        <p style={{ fontSize: 13, color: 'var(--text-dim)' }}>{selectedTeam.name}</p>
                      </div>
                      <div style={{ padding: 'var(--space-4)', background: 'var(--panel)', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)', marginBottom: 'var(--space-3)' }}>
                        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 'var(--space-2)' }}>
                          <span style={{ fontWeight: 500, color: 'var(--text)' }}>Team ID</span>
                          <span style={{ fontFamily: 'var(--font-mono)', fontSize: 11, color: 'var(--text-faint)', padding: '2px 8px', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius-sm)' }}>{selectedTeam.id.slice(0, 8)}...</span>
                        </div>
                        <p style={{ fontSize: 13, color: 'var(--text-dim)' }}>Unique identifier for API access</p>
                      </div>
                      <div style={{ padding: 'var(--space-4)', background: 'var(--panel)', border: '1px solid var(--panel-edge)', borderRadius: 'var(--radius)' }}>
                        <span style={{ fontWeight: 500, color: 'var(--text)', display: 'block', marginBottom: 'var(--space-2)' }}>Created</span>
                        <p style={{ fontSize: 13, color: 'var(--text-dim)' }}>{format(new Date(selectedTeam.created_at), 'MMMM d, yyyy')}</p>
                      </div>

                      <div style={{ height: 1, background: 'var(--panel-edge)', margin: 'var(--space-5) 0' }} />

                      <div style={{ padding: 'var(--space-4)', background: 'rgba(255, 107, 107, 0.04)', border: '1px solid rgba(255, 107, 107, 0.2)', borderRadius: 'var(--radius)' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--space-2)', marginBottom: 'var(--space-2)' }}>
                          <span style={{ fontWeight: 600, color: 'var(--status-revoked)' }}>Danger Zone</span>
                        </div>
                        <p style={{ fontSize: 13, color: 'var(--text-dim)', marginBottom: 'var(--space-3)' }}>Deleting a team will remove all members and data. This cannot be undone.</p>
                        <button className="btn-team-danger" onClick={() => { if (confirm(`Are you sure you want to delete "${selectedTeam.name}"? This cannot be undone.`)) deleteTeamMutation.mutate(selectedTeam.id); }} disabled={deleteTeamMutation.isPending}>
                          <Trash2 style={{ width: 14, height: 14 }} />Delete Team
                        </button>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </Chamber>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Create Team Modal */}
      <Modal open={isCreateDialogOpen} onClose={() => setIsCreateDialogOpen(false)} title="Create New Team">
        <p style={{ fontSize: 13, color: 'var(--text-dim)', marginBottom: 'var(--space-4)' }}>Create a new team to organize members and manage permissions.</p>
        <div style={{ marginBottom: 'var(--space-5)' }}>
          <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Team Name</label>
          <input type="text" placeholder="e.g., Engineering, Marketing" value={newTeamName} onChange={(e) => setNewTeamName(e.target.value)} className="team-input" />
        </div>
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 'var(--space-3)' }}>
          <FrameButton onClick={() => setIsCreateDialogOpen(false)}>Cancel</FrameButton>
          <SealedButton onClick={() => { if (!newTeamName.trim()) { toast.error('Team name is required'); return; } createTeamMutation.mutate(newTeamName); }} disabled={createTeamMutation.isPending}>
            {createTeamMutation.isPending ? 'Creating...' : 'Create Team'}
          </SealedButton>
        </div>
      </Modal>

      {/* Invite Modal */}
      <Modal open={isInviteDialogOpen} onClose={() => setIsInviteDialogOpen(false)} title="Invite Team Member">
        <p style={{ fontSize: 13, color: 'var(--text-dim)', marginBottom: 'var(--space-4)' }}>Send an invitation to join {selectedTeam?.name}.</p>
        <div style={{ marginBottom: 'var(--space-4)' }}>
          <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Email Address</label>
          <input type="email" placeholder="colleague@company.com" value={inviteEmail} onChange={(e) => setInviteEmail(e.target.value)} className="team-input" />
        </div>
        <div style={{ marginBottom: 'var(--space-5)' }}>
          <label style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>Role</label>
          <select value={inviteRole} onChange={(e) => setInviteRole(e.target.value as 'admin' | 'member' | 'viewer')} className="team-input" style={{ cursor: 'pointer', appearance: 'none' }}>
            <option value="admin">Admin — Full access</option>
            <option value="member">Member — Can manage resources</option>
            <option value="viewer">Viewer — Read-only access</option>
          </select>
        </div>
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 'var(--space-3)' }}>
          <FrameButton onClick={() => setIsInviteDialogOpen(false)}>Cancel</FrameButton>
          <SealedButton onClick={() => { if (!inviteEmail.trim() || !selectedTeam) { toast.error('Email is required'); return; } inviteMemberMutation.mutate({ teamId: selectedTeam.id, email: inviteEmail, role: inviteRole }); }} disabled={inviteMemberMutation.isPending || !inviteEmail.trim()}>
            {inviteMemberMutation.isPending ? 'Sending...' : 'Send Invitation'}
          </SealedButton>
        </div>
      </Modal>
    </div>
  );
}

export default TeamsPage;
