/**
 * Profile Header Component
 *
 * Displays the user's cover image, avatar, name, bio, and action buttons.
 */

import { useState } from "react";
import { motion } from "framer-motion";
import { format, formatDistanceToNow } from "date-fns";
import {
  Calendar,
  MapPin,
  Building2,
  Clock,
  Users,
  MessageCircle,
  Share2,
  MoreHorizontal,
  ExternalLink,
  Eye,
  CheckCircle2,
  Edit3,
} from "lucide-react";
import { cn, formatNumber } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { SocialLinks } from "@/components/profile/SocialLinks";
import type { UserProfile } from "@/types";

export interface ProfileHeaderProps {
  profile: UserProfile;
  isOwnProfile: boolean;
  /** When false, Follow and Message buttons are hidden (signed-out viewers). */
  isViewerSignedIn?: boolean;
  onEditProfile?: () => void;
  onAvatarClick?: () => void;
}

export function ProfileHeader({ profile, isOwnProfile, isViewerSignedIn = false, onEditProfile, onAvatarClick }: ProfileHeaderProps) {
  const [isFollowing, setIsFollowing] = useState(false);
  const [followersCount, setFollowersCount] = useState(profile.stats.followersCount);

  const handleFollow = () => {
    setIsFollowing(!isFollowing);
    setFollowersCount(prev => isFollowing ? prev - 1 : prev + 1);
  };

  const handleShare = async () => {
    try {
      await navigator.clipboard.writeText(window.location.href);
      // Could show toast here
    } catch (err) {
      console.error("Failed to copy URL", err);
    }
  };

  const joinedDate = format(new Date(profile.createdAt), "MMMM yyyy");
  const lastActiveText = profile.lastActive
    ? formatDistanceToNow(new Date(profile.lastActive), { addSuffix: true })
    : null;

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="relative"
    >
      {/* Cover Image */}
      <div className="h-48 md:h-64 rounded-t-xl overflow-hidden relative">
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

        {/* Gradient overlay */}
        <div className="absolute inset-0 bg-gradient-to-t from-background/80 via-transparent to-transparent" />
      </div>

      {/* Profile Info */}
      <div className="px-4 md:px-8 pb-6">
        <div className="flex flex-col md:flex-row md:items-end -mt-16 md:-mt-20 gap-4 md:gap-6">
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
                "absolute bottom-2 right-2 w-6 h-6 rounded-full border-4 border-background",
                profile.isOnline ? "bg-green-500" : "bg-gray-400"
              )}
              title={profile.isOnline ? "Online" : `Last active ${lastActiveText}`}
            />
          </div>

          {/* User Info */}
          <div className="flex-1 min-w-0 space-y-2">
            <div className="flex items-center gap-3 flex-wrap">
              <h1 className="text-2xl md:text-3xl font-bold text-text-primary">
                {profile.name || profile.username}
              </h1>
              {profile.stats.trustScore >= 80 && (
                <Badge variant="default" className="bg-brand-500/20 text-brand-400 border-brand-500/30">
                  <CheckCircle2 className="w-3 h-3 mr-1" />
                  Verified
                </Badge>
              )}
            </div>

            <p className="text-lg text-brand-400 font-medium">@{profile.username}</p>

            {profile.bio && (
              <p className="text-text-secondary max-w-2xl">{profile.bio}</p>
            )}

            {/* Meta info */}
            <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm text-text-muted">
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
            <SocialLinks
              links={profile.socialLinks}
              variant="compact"
              className="pt-1"
            />
          </div>

          {/* Action Buttons */}
          <div className="flex items-center gap-2 shrink-0">
            {!isOwnProfile && isViewerSignedIn ? (
              <>
                <Button
                  variant={isFollowing ? "outline" : "default"}
                  onClick={handleFollow}
                  className={cn(
                    "gap-2",
                    isFollowing && "border-brand-500 text-brand-400"
                  )}
                >
                  <Users className="w-4 h-4" />
                  {isFollowing ? "Following" : "Follow"}
                </Button>
                <Button variant="outline" className="gap-2">
                  <MessageCircle className="w-4 h-4" />
                  Message
                </Button>
              </>
            ) : !isOwnProfile ? null : (
              <Button variant="outline" onClick={onEditProfile} className="gap-2">
                <Edit3 className="w-4 h-4" />
                Edit Profile
              </Button>
            )}

            <Button variant="outline" size="icon" onClick={handleShare} aria-label="Share profile">
              <Share2 className="w-4 h-4" />
            </Button>

            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="icon" aria-label="Profile options">
                  <MoreHorizontal className="w-4 h-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem>
                  <ExternalLink className="w-4 h-4 mr-2" />
                  Copy Profile URL
                </DropdownMenuItem>
                <DropdownMenuItem>
                  <Eye className="w-4 h-4 mr-2" />
                  View in Registry
                </DropdownMenuItem>
                {!isOwnProfile && (
                  <>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem className="text-red-500">
                      Report User
                    </DropdownMenuItem>
                  </>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
      </div>
    </motion.div>
  );
}
