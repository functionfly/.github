/**
 * Profile Header Component
 *
 * Displays the user's cover image, avatar, name, bio, and action buttons.
 */

import { conversationsApi } from '@/api/conversations';
import { FollowUserButton } from '@/components/follow';
import { AdminBadge } from '@/components/profile/AdminBadge';
import { ReportProfileDialog } from '@/components/profile/ReportProfileDialog';
import { EnterpriseBadge } from '@/components/profile/EnterpriseBadge';
import { FounderBadge } from '@/components/profile/FounderBadge';
import { SocialLinks } from '@/components/profile/SocialLinks';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { FrameButton } from '@/components/containment';
import { isPlatformAdminRole } from '@/lib/platform-admin';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/stores/authStore';
import type { UserProfile } from '@/types';
import { format, formatDistanceToNow } from 'date-fns';
import { motion } from 'framer-motion';
import {
  Building2,
  Calendar,
  Check,
  CheckCircle2,
  Clock,
  Copy,
  Edit3,
  ExternalLink,
  Eye,
  Flag,
  Loader2,
  Mail,
  MapPin,
  MessageCircle,
  MoreHorizontal,
  Share2,
  X,
} from 'lucide-react';
import { Icon } from '@iconify/react';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

export interface ProfileHeaderProps {
  profile: UserProfile;
  isOwnProfile: boolean;
  /** When false, Follow and Message buttons are hidden (signed-out viewers). */
  isViewerSignedIn?: boolean;
  onEditProfile?: () => void;
  onAvatarClick?: () => void;
  /** Whether user is on enterprise plan - shows animated enterprise badge */
  isEnterprise?: boolean;
}

// Share dialog component — Sealed Containment
function ShareDialog({ profile, children }: { profile: UserProfile; children: React.ReactNode }) {
  const [open, setOpen] = useState(false);
  const [copied, setCopied] = useState(false);
  const profileUrl = `${window.location.origin}/u/${encodeURIComponent(profile.username)}`;
  const shareText = `Check out ${profile.name || profile.username}'s profile on FunctionFly — the platform for serverless functions`;

  const handleCopy = async () => {
    try { await navigator.clipboard.writeText(profileUrl); setCopied(true); setTimeout(() => setCopied(false), 2000); }
    catch (err) { console.error('Failed to copy URL', err); }
  };

  const shareLinks = [
    { name: 'Twitter', icon: () => <Icon icon="simple-icons:x" className="ph-share-icon" />, url: `https://twitter.com/intent/tweet?text=${encodeURIComponent(shareText)}&url=${encodeURIComponent(profileUrl)}` },
    { name: 'LinkedIn', icon: () => <Icon icon="simple-icons:linkedin" className="ph-share-icon" />, url: `https://www.linkedin.com/sharing/share-offsite/?url=${encodeURIComponent(profileUrl)}` },
    { name: 'Facebook', icon: () => (<svg className="ph-share-icon" fill="currentColor" viewBox="0 0 24 24"><path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z" /></svg>), url: `https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(profileUrl)}` },
    { name: 'WhatsApp', icon: () => (<svg className="ph-share-icon" fill="currentColor" viewBox="0 0 24 24"><path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z" /></svg>), url: `https://wa.me/?text=${encodeURIComponent(`${shareText} ${profileUrl}`)}` },
    { name: 'Telegram', icon: () => (<svg className="ph-share-icon" fill="currentColor" viewBox="0 0 24 24"><path d="M11.944 0A12 12 0 0 0 0 12a12 12 0 0 0 12 12 12 12 0 0 0 12-12A12 12 0 0 0 12 0a12 12 0 0 0-.056 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 0 1 .171.325c.016.093.036.306.02.472-.18 1.898-.962 6.502-1.36 8.627-.168.9-.499 1.201-.82 1.23-.696.065-1.225-.46-1.9-.902-1.056-.693-1.653-1.124-2.678-1.8-1.185-.78-.417-1.21.258-1.91.177-.184 3.247-2.977 3.307-3.23.007-.032.014-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.14-5.061 3.345-.48.33-.913.49-1.302.48-.428-.008-1.252-.241-1.865-.44-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.893-.663 3.498-1.524 5.83-2.529 6.998-3.014 3.332-1.386 4.025-1.627 4.476-1.635z" /></svg>), url: `https://t.me/share/url?url=${encodeURIComponent(profileUrl)}&text=${encodeURIComponent(shareText)}` },
    { name: 'Email', icon: Mail, url: `mailto:?subject=${encodeURIComponent(`FunctionFly Profile: ${profile.name || profile.username}`)}&body=${encodeURIComponent(`${shareText}\n\n${profileUrl}`)}` },
  ];

  const handleNativeShare = async () => {
    if (navigator.share) { try { await navigator.share({ title: `${profile.name || profile.username} - FunctionFly`, text: shareText, url: profileUrl }); } catch {} }
  };

  return (
    <>
      <button type="button" className="ph-share-trigger" onClick={() => setOpen(true)} aria-label="Share profile">
        {children}
      </button>
      {open && (
        <div className="ph-share-overlay" onClick={() => setOpen(false)}>
          <div className="ph-share-modal" onClick={(e) => e.stopPropagation()}>
            <div className="ph-share-header">
              <h2 className="ph-share-title">Share Profile</h2>
              <p className="ph-share-desc">Spread the word about {profile.name || profile.username}&apos;s work</p>
              <button className="ph-share-close" onClick={() => setOpen(false)} aria-label="Close">×</button>
            </div>

            <div className="ph-share-body">
              <div className="ph-share-preview">
                <div className="ph-share-avatar">{(profile.name || profile.username).charAt(0).toUpperCase()}</div>
                <div className="ph-share-preview-info">
                  <p className="ph-share-preview-name">{profile.name || profile.username}</p>
                  <p className="ph-share-preview-handle">@{profile.username}</p>
                </div>
                {profile.profileNumber && <span className="ph-share-member">#{profile.profileNumber}</span>}
              </div>

              <div className="ph-share-grid">
                {shareLinks.map((link) => {
                  const LinkIcon = link.icon;
                  return (
                    <a key={link.name} href={link.url} target="_blank" rel="noopener noreferrer" className="ph-share-link">
                      <LinkIcon />
                      <span>{link.name}</span>
                    </a>
                  );
                })}
                <button type="button" className="ph-share-link" onClick={handleCopy}>
                  {copied ? <Check className="ph-share-icon" /> : <Copy className="ph-share-icon" />}
                  <span>{copied ? 'Copied' : 'Copy'}</span>
                </button>
              </div>

              {navigator.share && (
                <FrameButton onClick={handleNativeShare} iconLeft={<Share2 className="ph-share-icon" />}>
                  More Sharing Options
                </FrameButton>
              )}

              <div className="ph-share-url-bar">
                <ExternalLink className="ph-share-url-icon" />
                <input type="text" readOnly value={profileUrl} className="ph-share-url-input" />
                <button type="button" className="ph-share-url-copy" onClick={handleCopy}>
                  {copied ? <><Check className="ph-share-icon" /> Copied</> : <><Copy className="ph-share-icon" /> Copy</>}
                </button>
              </div>

              <p className="ph-share-footer-text">Share this profile to help {profile.username} grow their network on FunctionFly</p>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

export function ProfileHeader({
  profile,
  isOwnProfile,
  isViewerSignedIn = false,
  onEditProfile,
  onAvatarClick,
  isEnterprise = false,
}: ProfileHeaderProps) {
  const navigate = useNavigate();
  const currentUser = useAuthStore((s) => s.user);
  const [messageLoading, setMessageLoading] = useState(false);
  const [reportOpen, setReportOpen] = useState(false);

  const handleMessageUser = async () => {
    if (!currentUser?.id || !currentUser?.username) {
      toast.error('Sign in to send a message');
      return;
    }
    if (currentUser.id === profile.id) {
      return;
    }
    setMessageLoading(true);
    try {
      const { conversations } = await conversationsApi.listConversations(currentUser.username, { limit: 100 });
      const existingDm = conversations.find(
        (c) =>
          c.type === 'dm' &&
          c.participant_ids.length === 2 &&
          c.participant_ids.includes(profile.id) &&
          c.participant_ids.includes(currentUser.id)
      );
      if (existingDm) {
        navigate(`/u/${currentUser.username}/conversations/${existingDm.id}`);
        return;
      }
      const conv = await conversationsApi.createConversation(currentUser.username, {
        type: 'dm',
        participant_ids: [currentUser.id, profile.id],
      });
      navigate(`/u/${currentUser.username}/conversations/${conv.id}`);
    } catch (e: unknown) {
      const res =
        e && typeof e === 'object' && 'response' in e
          ? (e as { response?: { data?: unknown } }).response
          : null;
      const data = res?.data as { error?: string; message?: string } | undefined;
      const msg = data?.error || data?.message;
      toast.error(typeof msg === 'string' && msg.trim() ? msg : 'Could not open conversation');
    } finally {
      setMessageLoading(false);
    }
  };

  const handleCopyProfileUrl = () => {
    const url = `${window.location.origin}/u/${encodeURIComponent(profile.username)}`;
    void navigator.clipboard.writeText(url).then(
      () => toast.success('Profile link copied'),
      () => toast.error('Could not copy link')
    );
  };

  const handleViewInRegistry = () => {
    navigate(`/registry?author=${encodeURIComponent(profile.username)}`);
  };

  // Safely format the joined date
  const joinedDate = (() => {
    try {
      if (profile.createdAt) {
        return format(new Date(profile.createdAt), 'MMMM yyyy');
      }
      return 'Unknown';
    } catch {
      return 'Unknown';
    }
  })();

  // Safely format the last active time (handles both date strings and relative text like "Just now")
  const lastActiveText = (() => {
    if (!profile.lastActive) return null;
    // If it's already a relative string (not a date), just display it
    if (['just now', 'online', 'active now'].includes(profile.lastActive.toLowerCase())) {
      return profile.lastActive;
    }
    try {
      const date = new Date(profile.lastActive);
      // Check if valid date
      if (isNaN(date.getTime())) {
        return profile.lastActive; // Return original string if not a valid date
      }
      return formatDistanceToNow(date, { addSuffix: true });
    } catch {
      return profile.lastActive; // Return original string on error
    }
  })();

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="relative isolate"
    >
      {/* Cover Image — keep below profile row so overlapping name/avatar/badge stay visible */}
      <div className="relative z-0 h-48 md:h-64 rounded-t-xl overflow-hidden">
        {profile.coverImage ? (
          <img
            src={profile.coverImage}
            alt="Profile cover"
            className="w-full h-full object-cover"
          />
        ) : (
          <div className="w-full h-full bg-gradient-to-br from-brand-500 via-brand-600 to-indigo-700">
            <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjxnIGZpbGw9IiNmZmZmZmYiIGZpbGwtb3BhY2l0eT0iMC4wNSI+PGNpcmNsZSBjeD0iMzAiIGN5PSIzMCIgcj0iMiIvPjwvZz48L2c+PC9zdmc+')] opacity-30" />
          </div>
        )}

        {/* Gradient overlay — stronger at bottom in dark mode so edge meets content cleanly */}
        <div className="absolute inset-0 bg-gradient-to-t from-background/95 via-background/35 to-transparent" />
      </div>

      {/* Profile Info — must stack above cover (negative margin overlap) */}
      <div className="relative z-10 px-4 md:px-8 pb-6">
        <div className="flex flex-col md:flex-row md:items-start -mt-16 md:-mt-20 gap-4 md:gap-6">
          {/* Avatar */}
          <div className="relative shrink-0">
            {isOwnProfile && onAvatarClick ? (
              <button
                type="button"
                onClick={onAvatarClick}
                className="rounded-full focus:outline-none focus:ring-2 focus:ring-brand-500 focus:ring-offset-2 focus:ring-offset-background"
                title="Change profile picture"
              >
                <Avatar className="w-32 h-32 md:w-40 md:h-40 rounded-full border-4 border-background shadow-xl ring-0 bg-gradient-to-br from-brand-500 to-brand-600 cursor-pointer transition opacity-90 hover:opacity-100">
                  <AvatarImage
                    src={profile.avatar}
                    alt={profile.name || profile.username}
                    className="object-cover"
                  />
                  <AvatarFallback className="bg-gradient-to-br from-brand-500 to-brand-600 text-white text-4xl md:text-5xl font-bold">
                    {(profile.name || profile.username).charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
              </button>
            ) : (
              <Avatar className="w-32 h-32 md:w-40 md:h-40 rounded-full border-4 border-background shadow-xl ring-0 bg-gradient-to-br from-brand-500 to-brand-600">
                <AvatarImage
                  src={profile.avatar}
                  alt={profile.name || profile.username}
                  className="object-cover"
                />
                <AvatarFallback className="bg-gradient-to-br from-brand-500 to-brand-600 text-white text-4xl md:text-5xl font-bold">
                  {(profile.name || profile.username).charAt(0).toUpperCase()}
                </AvatarFallback>
              </Avatar>
            )}

            {/* Online status indicator */}
            <div
              className={cn(
                'absolute bottom-2 right-2 w-6 h-6 rounded-full border-4 border-background z-20',
                profile.isOnline
                  ? 'bg-green-400 shadow-[0_0_8px_rgba(74,222,128,0.6)]'
                  : 'bg-gray-500'
              )}
              title={
                profile.isOnline
                  ? 'Online'
                  : lastActiveText
                    ? `Last active ${lastActiveText}`
                    : 'Offline'
              }
            />
          </div>

          {/* User Info — md:pt-20 matches md:-mt-20 on the row so the title row clears the cover (avatar still overlaps) */}
          <div className="flex-1 min-w-0 space-y-2 md:pt-20">
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="text-2xl md:text-3xl font-bold font-display text-text-primary">
                {profile.name || profile.username}
              </h1>
              {/* Show all three badges for FunctionFly profile */}
              {(isEnterprise || profile.username === 'FunctionFly') && (
                <EnterpriseBadge size="md" showParticles />
              )}
              {isPlatformAdminRole(profile.role) ? (
                <AdminBadge
                  size="md"
                  showParticles
                  variant={
                    profile.role === 'super_admin'
                      ? 'super_admin'
                      : profile.role === 'support'
                        ? 'support'
                        : 'admin'
                  }
                />
              ) : profile.username === 'FunctionFly' ? (
                <AdminBadge size="md" showParticles variant="super_admin" />
              ) : null}
              {profile.username === 'FunctionFly' && (
                <AdminBadge size="md" showParticles variant="support" />
              )}
              {profile.founderNumber && profile.founderNumber > 0 && (
                <FounderBadge founderNumber={profile.founderNumber} size="md" />
              )}
              {profile.stats.trustScore >= 80 && (
                <Badge
                  variant="default"
                  className="bg-brand-500/20 text-brand-400 border-brand-500/30"
                >
                  <CheckCircle2 className="w-3 h-3 mr-1" />
                  Verified
                </Badge>
              )}
            </div>

            <div className="flex items-center gap-2 flex-wrap">
              <p className="text-base md:text-lg text-text-secondary font-medium tracking-tight">
                @{profile.username}
              </p>
              {profile.profileNumber && profile.profileNumber > 0 && (
                <Badge
                  variant="outline"
                  className="text-xs font-medium text-text-muted border-border-subtle bg-background/50 font-mono"
                  title={`Member #${profile.profileNumber.toLocaleString()} - Early adopter badge`}
                >
                  #{profile.profileNumber.toLocaleString()}
                </Badge>
              )}
            </div>

            {profile.bio && (
              <p className="text-text-secondary max-w-2xl leading-relaxed">{profile.bio}</p>
            )}

            {/* Meta info — text-muted is too dim on dark card; secondary meets body contrast */}
            <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm text-text-secondary [&_svg]:shrink-0">
              {profile.location && (
                <span className="flex items-center gap-1">
                  <MapPin className="w-4 h-4" />
                  {profile.location}
                </span>
              )}
              {profile.company && (
                <span className="flex items-center gap-1">
                  <Building2 className="w-4 h-4" />
                  {profile.company}
                  {profile.jobTitle && ` · ${profile.jobTitle}`}
                </span>
              )}
              <span className="flex items-center gap-1">
                <Calendar className="w-4 h-4" />
                Joined {joinedDate}
              </span>
              {lastActiveText && !profile.isOnline && (
                <span className="flex items-center gap-1">
                  <Clock className="w-4 h-4" />
                  Active {lastActiveText}
                </span>
              )}
            </div>

            {/* Social Links */}
            <SocialLinks links={profile.socialLinks} variant="compact" className="pt-1" />
          </div>

          {/* Action Buttons — align with avatar bottom when row uses items-start */}
          <div className="flex items-center gap-2 shrink-0 md:self-end">
            {!isOwnProfile && isViewerSignedIn ? (
              <>
                <FollowUserButton username={profile.username} showDropdown={false} />
                <Button
                  type="button"
                  variant="outline"
                  className="gap-2"
                  disabled={messageLoading}
                  onClick={() => void handleMessageUser()}
                >
                  {messageLoading ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Opening…
                    </>
                  ) : (
                    <>
                      <MessageCircle className="w-4 h-4" />
                      Message
                    </>
                  )}
                </Button>
              </>
            ) : !isOwnProfile ? null : (
              <Button variant="outline" onClick={onEditProfile} className="gap-2">
                <Edit3 className="w-4 h-4" />
                Edit Profile
              </Button>
            )}

            <ShareDialog profile={profile}>
              <Button variant="outline" size="icon" aria-label="Share profile">
                <Share2 className="w-4 h-4" />
              </Button>
            </ShareDialog>

            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="icon" aria-label="Profile options">
                  <MoreHorizontal className="w-4 h-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem onSelect={handleCopyProfileUrl}>
                  <ExternalLink className="w-4 h-4 mr-2" />
                  Copy Profile URL
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={handleViewInRegistry}>
                  <Eye className="w-4 h-4 mr-2" />
                  View in Registry
                </DropdownMenuItem>
                {!isOwnProfile && isViewerSignedIn && (
                  <>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                      className="text-red-600 focus:text-red-600"
                      onSelect={(e) => {
                        e.preventDefault();
                        setReportOpen(true);
                      }}
                    >
                      <Flag className="w-4 h-4 mr-2" />
                      Report profile…
                    </DropdownMenuItem>
                  </>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </div>

      {!isOwnProfile && isViewerSignedIn && (
        <ReportProfileDialog
          open={reportOpen}
          onOpenChange={setReportOpen}
          username={profile.username}
          displayName={profile.name || profile.username}
        />
      )}
    </motion.div>
  );
}
