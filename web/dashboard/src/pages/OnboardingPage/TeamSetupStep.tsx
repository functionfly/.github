import { useState } from "react";
import { motion } from "framer-motion";
import { useNavigate } from "react-router-dom";
import { Users, Mail, Check, Loader2, Crown, User, Eye, Plus, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card } from "@/components/ui/card";
import { HelpTooltip } from "@/components/ui/help-tooltip";
import { useOnboardingStore } from "@/stores/onboardingStore";
import { toast } from "sonner";

type TeamRole = 'admin' | 'member' | 'viewer';

interface TeamInvite {
  email: string;
  role: TeamRole;
}

export function TeamSetupStep() {
  const navigate = useNavigate();
  const { updateStepData, setTeamInvites, setUserRole, skipOnboarding } = useOnboardingStore();
  const [userRole, setLocalUserRole] = useState<TeamRole>('admin');
  const [invites, setInvites] = useState<TeamInvite[]>([]);
  const [newEmail, setNewEmail] = useState('');
  const [newRole, setNewRole] = useState<TeamRole>('member');
  const [isSendingInvites, setIsSendingInvites] = useState(false);
  const [isSkipping, setIsSkipping] = useState(false);

  const roleOptions = [
    {
      value: 'admin' as TeamRole,
      label: 'Admin',
      description: 'Full access to manage team, providers, and functions',
      icon: Crown,
      color: 'text-aviation-stratosphere',
    },
    {
      value: 'member' as TeamRole,
      label: 'Member',
      description: 'Can deploy and manage functions, view team resources',
      icon: User,
      color: 'text-aviation-cyan',
    },
    {
      value: 'viewer' as TeamRole,
      label: 'Viewer',
      description: 'Read-only access to team functions and metrics',
      icon: Eye,
      color: 'text-aviation-green',
    },
  ];

  const addInvite = () => {
    if (!newEmail.trim()) return;

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(newEmail)) {
      toast.error('Please enter a valid email address');
      return;
    }

    if (invites.some(invite => invite.email === newEmail)) {
      toast.error('This email has already been added');
      return;
    }

    setInvites([...invites, { email: newEmail, role: newRole }]);
    setNewEmail('');
    setNewRole('member');
    toast.success('Team member added');
  };

  const removeInvite = (email: string) => {
    setInvites(invites.filter(invite => invite.email !== email));
  };

  const handleSendInvites = async () => {
    if (invites.length === 0) {
      handleSkip();
      return;
    }

    setIsSendingInvites(true);

    try {
      const response = await fetch('/v1/teams/invites', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          emails: invites.map(invite => invite.email),
          role: invites[0].role,
          message: 'Welcome to our FunctionFly team! Please join us to start deploying functions together.',
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to send invites');
      }

      const result = await response.json();

      setTeamInvites(result.invites.map((invite: any) => ({
        email: invite.email,
        token: invite.token,
        role: invites.find(i => i.email === invite.email)?.role || 'member',
        expires: invite.expires,
      })));

      setUserRole(userRole);

      updateStepData("team-setup", {
        userRole,
        invitesSent: result.invites.length,
        teamCreated: true,
        inviteTokens: result.invites,
      });

      toast.success(`Team invites sent to ${result.invites.length} member${result.invites.length > 1 ? 's' : ''}!`);
    } catch (error) {
      console.error('Team invite error:', error);
      toast.error(error instanceof Error ? error.message : 'Failed to send team invites');
    } finally {
      setIsSendingInvites(false);
    }
  };

  const handleSkip = () => {
    setIsSkipping(true);

    setUserRole('admin');

    updateStepData("team-setup", {
      userRole: 'admin',
      skipped: true,
      teamCreated: false,
    });

    skipOnboarding();
    setIsSkipping(false);
    navigate('/overview', { replace: true });
  };

  return (
    <div className="space-y-6">
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <Crown className="w-5 h-5 text-aviation-stratosphere" />
          <h3 className="text-lg font-mono font-semibold text-aviation-text-primary">Your Role</h3>
          <HelpTooltip content="Choose your role in the team. You can change this later in team settings." />
        </div>

        <div className="grid gap-3">
          {roleOptions.map((option) => {
            const Icon = option.icon;
            return (
              <Card
                key={option.value}
                className={`aviation-instrument p-4 cursor-pointer transition-all ${
                  userRole === option.value
                    ? "border-aviation-stratosphere ring-1 ring-aviation-stratosphere bg-aviation-stratosphere/5"
                    : "hover:border-aviation-border-glow"
                }`}
                onClick={() => setLocalUserRole(option.value)}
              >
                <div className="flex items-center gap-3">
                  <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
                    userRole === option.value ? "bg-aviation-stratosphere/20" : "bg-aviation-bg-tertiary"
                  }`}>
                    <Icon className={`w-5 h-5 ${userRole === option.value ? "text-aviation-stratosphere" : "text-aviation-text-muted"}`} />
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <h4 className="font-mono font-semibold text-aviation-text-primary">{option.label}</h4>
                      {userRole === option.value && <Check className="w-4 h-4 text-aviation-stratosphere" />}
                    </div>
                    <p className="text-sm font-mono text-aviation-text-secondary">{option.description}</p>
                  </div>
                </div>
              </Card>
            );
          })}
        </div>
      </div>

      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <Users className="w-5 h-5 text-aviation-amber" />
          <h3 className="text-lg font-mono font-semibold text-aviation-text-primary">Invite Team Members</h3>
          <HelpTooltip content="Add team members to collaborate on functions and share provider configurations." />
        </div>

        <Card className="aviation-panel p-4">
          <div className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div className="md:col-span-2">
                <Label htmlFor="email" className="flex items-center gap-2 font-mono text-aviation-text-secondary">
                  <Mail className="w-4 h-4" />
                  Email Address
                </Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="colleague@company.com"
                  value={newEmail}
                  onChange={(e) => setNewEmail(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && addInvite()}
                  className="aviation-input mt-1"
                />
              </div>
              <div>
                <Label htmlFor="role" className="font-mono text-aviation-text-secondary">Role</Label>
                <select
                  id="role"
                  value={newRole}
                  onChange={(e) => setNewRole(e.target.value as TeamRole)}
                  className="mt-1 w-full px-3 py-2 bg-aviation-bg-instrument border border-aviation-border-instrument rounded-md text-aviation-text-primary font-mono focus:border-aviation-amber focus:ring-1 focus:ring-aviation-amber"
                >
                  {roleOptions.slice(1).map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>
            <Button
              onClick={addInvite}
              disabled={!newEmail.trim()}
              className="aviation-button w-full md:w-auto"
            >
              <Plus className="w-4 h-4 mr-2" />
              Add Team Member
            </Button>
          </div>
        </Card>

        {invites.length > 0 && (
          <div className="space-y-3">
            <h4 className="font-mono font-medium text-aviation-text-primary">Pending Invites ({invites.length})</h4>
            {invites.map((invite) => {
              const roleData = roleOptions.find(r => r.value === invite.role);
              const Icon = roleData?.icon || User;
              return (
                <Card key={invite.email} className="aviation-instrument p-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-full bg-aviation-bg-tertiary flex items-center justify-center">
                        <Icon className="w-4 h-4 text-aviation-text-muted" />
                      </div>
                      <div>
                        <p className="text-sm font-mono font-medium text-aviation-text-primary">{invite.email}</p>
                        <p className="text-xs font-mono text-aviation-text-muted">{roleData?.label}</p>
                      </div>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => removeInvite(invite.email)}
                      className="text-aviation-red hover:text-aviation-red hover:bg-aviation-red-dim"
                    >
                      <X className="w-4 h-4" />
                    </Button>
                  </div>
                </Card>
              );
            })}
          </div>
        )}

        <div className="flex gap-3 pt-4">
          <Button
            onClick={handleSendInvites}
            disabled={isSendingInvites || isSkipping}
            className="aviation-button-primary flex-1 font-mono"
          >
            {isSendingInvites ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Sending Invites...
              </>
            ) : invites.length > 0 ? (
              <>
                <Mail className="w-4 h-4 mr-2" />
                Send {invites.length} Invite{invites.length > 1 ? 's' : ''}
              </>
            ) : (
              <>
                <Check className="w-4 h-4 mr-2" />
                Continue Solo
              </>
            )}
          </Button>
          <Button
            variant="outline"
            onClick={handleSkip}
            disabled={isSendingInvites || isSkipping}
            className="flex-1 font-mono border-aviation-border-instrument text-aviation-text-primary hover:border-aviation-amber hover:text-aviation-amber"
          >
            {isSkipping ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Skipping...
              </>
            ) : (
              'Skip for Now'
            )}
          </Button>
        </div>

        {invites.length === 0 && (
          <div className="bg-aviation-cyan-dim border border-aviation-cyan/30 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <Users className="w-5 h-5 text-aviation-cyan flex-shrink-0 mt-0.5" />
              <div>
                <h4 className="font-mono font-medium text-aviation-cyan mb-1">Team Benefits</h4>
                <p className="text-sm font-mono text-aviation-text-secondary">
                  Invite team members to collaborate on functions, share provider configurations,
                  and manage deployments together. You can always add members later from your dashboard.
                </p>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}