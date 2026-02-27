import { useState } from "react";
import { motion } from "framer-motion";
import { Users, Mail, Check, Loader2, AlertTriangle, Crown, User, Eye, Plus, X } from "lucide-react";
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
  const { updateStepData, setTeamInvites, setUserRole } = useOnboardingStore();
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
      color: 'text-purple-500',
    },
    {
      value: 'member' as TeamRole,
      label: 'Member',
      description: 'Can deploy and manage functions, view team resources',
      icon: User,
      color: 'text-blue-500',
    },
    {
      value: 'viewer' as TeamRole,
      label: 'Viewer',
      description: 'Read-only access to team functions and metrics',
      icon: Eye,
      color: 'text-green-500',
    },
  ];

  const addInvite = () => {
    if (!newEmail.trim()) return;

    // Basic email validation
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(newEmail)) {
      toast.error('Please enter a valid email address');
      return;
    }

    // Check if email already exists
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
      // Skip team setup
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
          role: invites[0].role, // For now, use the same role for all invites
          message: 'Welcome to our FunctionFly team! Please join us to start deploying functions together.',
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to send invites');
      }

      const result = await response.json();

      // Update store with team invites
      setTeamInvites(result.invites.map((invite: any) => ({
        email: invite.email,
        token: invite.token,
        role: invites.find(i => i.email === invite.email)?.role || 'member',
        expires: invite.expires,
      })));

      // Update user role
      setUserRole(userRole);

      // Save step data
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

    // Update store for solo user
    setUserRole('admin');

    // Save step data
    updateStepData("team-setup", {
      userRole: 'admin',
      skipped: true,
      teamCreated: false,
    });

    // Simulate brief loading
    setTimeout(() => {
      setIsSkipping(false);
      toast.success('Continuing with solo setup');
    }, 1000);
  };

  return (
    <div className="space-y-6">
      {/* User Role Selection */}
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <Crown className="w-5 h-5 text-[#6366f1]" />
          <h3 className="text-lg font-medium text-text-primary">Your Role</h3>
          <HelpTooltip content="Choose your role in the team. You can change this later in team settings." />
        </div>

        <div className="grid gap-3">
          {roleOptions.map((option) => {
            const Icon = option.icon;
            return (
              <Card
                key={option.value}
                className={`card p-4 cursor-pointer transition-all ${
                  userRole === option.value
                    ? "border-[#6366f1] ring-1 ring-[#6366f1] bg-[#6366f1]/5"
                    : "hover:border-border-default"
                }`}
                onClick={() => setLocalUserRole(option.value)}
              >
                <div className="flex items-center gap-3">
                  <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
                    userRole === option.value ? "bg-[#6366f1]/20" : "bg-bg-tertiary"
                  }`}>
                    <Icon className={`w-5 h-5 ${userRole === option.value ? "text-[#6366f1]" : "text-text-muted"}`} />
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <h4 className="font-medium text-text-primary">{option.label}</h4>
                      {userRole === option.value && <Check className="w-4 h-4 text-[#6366f1]" />}
                    </div>
                    <p className="text-sm text-text-secondary">{option.description}</p>
                  </div>
                </div>
              </Card>
            );
          })}
        </div>
      </div>

      {/* Team Invitations */}
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <Users className="w-5 h-5 text-[#6366f1]" />
          <h3 className="text-lg font-medium text-text-primary">Invite Team Members</h3>
          <HelpTooltip content="Add team members to collaborate on functions and share provider configurations." />
        </div>

        {/* Add Invite Form */}
        <Card className="card p-4">
          <div className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div className="md:col-span-2">
                <Label htmlFor="email" className="flex items-center gap-2">
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
                  className="mt-1"
                />
              </div>
              <div>
                <Label htmlFor="role">Role</Label>
                <select
                  id="role"
                  value={newRole}
                  onChange={(e) => setNewRole(e.target.value as TeamRole)}
                  className="mt-1 w-full px-3 py-2 bg-bg-primary border border-border-subtle rounded-md text-text-primary focus:border-[#6366f1] focus:ring-1 focus:ring-[#6366f1]"
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
              className="w-full md:w-auto"
            >
              <Plus className="w-4 h-4 mr-2" />
              Add Team Member
            </Button>
          </div>
        </Card>

        {/* Pending Invites */}
        {invites.length > 0 && (
          <div className="space-y-3">
            <h4 className="font-medium text-text-primary">Pending Invites ({invites.length})</h4>
            {invites.map((invite) => {
              const roleData = roleOptions.find(r => r.value === invite.role);
              const Icon = roleData?.icon || User;
              return (
                <Card key={invite.email} className="card p-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 rounded-full bg-bg-tertiary flex items-center justify-center">
                        <Icon className="w-4 h-4 text-text-muted" />
                      </div>
                      <div>
                        <p className="text-sm font-medium text-text-primary">{invite.email}</p>
                        <p className="text-xs text-text-muted">{roleData?.label}</p>
                      </div>
                    </div>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => removeInvite(invite.email)}
                      className="text-red-500 hover:text-red-600 hover:bg-red-500/10"
                    >
                      <X className="w-4 h-4" />
                    </Button>
                  </div>
                </Card>
              );
            })}
          </div>
        )}

        {/* Action Buttons */}
        <div className="flex gap-3 pt-4">
          <Button
            onClick={handleSendInvites}
            disabled={isSendingInvites || isSkipping}
            className="btn-primary flex-1"
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
            className="flex-1"
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
          <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <Users className="w-5 h-5 text-blue-500 flex-shrink-0 mt-0.5" />
              <div>
                <h4 className="font-medium text-blue-400 mb-1">Team Benefits</h4>
                <p className="text-sm text-text-secondary">
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