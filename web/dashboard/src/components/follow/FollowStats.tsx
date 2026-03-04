import { useState } from "react";
import { Users, UserPlus, FunctionSquare, X } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Link } from "react-router-dom";
import {
  useUserFollowers,
  useUserFollowing,
  useMyFollowStats,
} from "@/hooks/useFollow";
import { formatDistanceToNow } from "@/lib/utils";

interface FollowStatsProps {
  username: string;
  followerCount?: number;
  followingCount?: number;
}

export function FollowStats({ username, followerCount, followingCount }: FollowStatsProps) {
  const [showFollowers, setShowFollowers] = useState(false);
  const [showFollowing, setShowFollowing] = useState(false);

  return (
    <>
      <div className="flex items-center gap-4 text-sm">
        <button
          onClick={() => setShowFollowers(true)}
          className="hover:text-brand-400 transition-colors flex items-center gap-1.5"
        >
          <span className="font-semibold text-text-primary">
            {followerCount?.toLocaleString() ?? 0}
          </span>
          <span className="text-text-secondary">followers</span>
        </button>
        <span className="text-text-muted">·</span>
        <button
          onClick={() => setShowFollowing(true)}
          className="hover:text-brand-400 transition-colors flex items-center gap-1.5"
        >
          <span className="font-semibold text-text-primary">
            {followingCount?.toLocaleString() ?? 0}
          </span>
          <span className="text-text-secondary">following</span>
        </button>
      </div>

      <FollowersDialog
        username={username}
        open={showFollowers}
        onOpenChange={setShowFollowers}
      />
      <FollowingDialog
        username={username}
        open={showFollowing}
        onOpenChange={setShowFollowing}
      />
    </>
  );
}

interface FollowersDialogProps {
  username: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

function FollowersDialog({ username, open, onOpenChange }: FollowersDialogProps) {
  const { data, isLoading } = useUserFollowers(username, 1, 50);
  const followers = data?.data ?? [];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md max-h-[80vh] p-0 gap-0">
        <DialogHeader className="p-4 pb-2 border-b border-border-subtle">
          <DialogTitle className="flex items-center gap-2">
            <Users className="w-5 h-5 text-brand-500" />
            Followers
          </DialogTitle>
        </DialogHeader>
        <ScrollArea className="max-h-[60vh]">
          <div className="p-4 space-y-3">
            {isLoading ? (
              <FollowListSkeleton />
            ) : followers.length === 0 ? (
              <div className="text-center py-8 text-text-muted">
                <UserPlus className="w-10 h-10 mx-auto mb-3 opacity-40" />
                <p>No followers yet</p>
              </div>
            ) : (
              followers.map((follower) => (
                <FollowerItem key={follower.id} follower={follower} />
              ))
            )}
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
}

interface FollowingDialogProps {
  username: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

function FollowingDialog({ username, open, onOpenChange }: FollowingDialogProps) {
  const { data, isLoading } = useUserFollowing(username, 1, 50);
  const following = data?.data ?? [];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md max-h-[80vh] p-0 gap-0">
        <DialogHeader className="p-4 pb-2 border-b border-border-subtle">
          <DialogTitle className="flex items-center gap-2">
            <UserPlus className="w-5 h-5 text-brand-500" />
            Following
          </DialogTitle>
        </DialogHeader>
        <ScrollArea className="max-h-[60vh]">
          <div className="p-4 space-y-3">
            {isLoading ? (
              <FollowListSkeleton />
            ) : following.length === 0 ? (
              <div className="text-center py-8 text-text-muted">
                <Users className="w-10 h-10 mx-auto mb-3 opacity-40" />
                <p>Not following anyone yet</p>
              </div>
            ) : (
              following.map((follow) => (
                <FollowingItem key={follow.id} follow={follow} />
              ))
            )}
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
}

interface FollowerItemProps {
  follower: {
    id: string;
    follower_id: string;
    follower_name?: string;
    follower_avatar?: string;
    created_at: string;
  };
}

function FollowerItem({ follower }: FollowerItemProps) {
  const displayName = follower.follower_name || "Unknown User";
  const initial = displayName.charAt(0).toUpperCase();

  return (
    <div className="flex items-center gap-3 p-2 rounded-lg hover:bg-bg-secondary transition-colors">
      <Avatar className="w-10 h-10">
        <AvatarImage src={follower.follower_avatar} alt={displayName} />
        <AvatarFallback className="bg-brand-500/20 text-brand-400 text-sm">
          {initial}
        </AvatarFallback>
      </Avatar>
      <div className="flex-1 min-w-0">
        <Link
          to={`/@/${follower.follower_name || follower.follower_id}`}
          className="font-medium text-text-primary hover:text-brand-400 transition-colors block truncate"
        >
          {displayName}
        </Link>
        <p className="text-xs text-text-muted">
          Followed {formatDistanceToNow(follower.created_at)}
        </p>
      </div>
    </div>
  );
}

interface FollowingItemProps {
  follow: {
    id: string;
    followed_id?: string;
    followed_name?: string;
    followed_avatar?: string;
    created_at: string;
  };
}

function FollowingItem({ follow }: FollowingItemProps) {
  const displayName = follow.followed_name || "Unknown User";
  const initial = displayName.charAt(0).toUpperCase();

  return (
    <div className="flex items-center gap-3 p-2 rounded-lg hover:bg-bg-secondary transition-colors">
      <Avatar className="w-10 h-10">
        <AvatarImage src={follow.followed_avatar} alt={displayName} />
        <AvatarFallback className="bg-brand-500/20 text-brand-400 text-sm">
          {initial}
        </AvatarFallback>
      </Avatar>
      <div className="flex-1 min-w-0">
        <Link
          to={`/@/${follow.followed_name || follow.followed_id}`}
          className="font-medium text-text-primary hover:text-brand-400 transition-colors block truncate"
        >
          {displayName}
        </Link>
        <p className="text-xs text-text-muted">
          Following since {formatDistanceToNow(follow.created_at)}
        </p>
      </div>
    </div>
  );
}

function FollowListSkeleton() {
  return (
    <>
      {[1, 2, 3].map((i) => (
        <div key={i} className="flex items-center gap-3 p-2">
          <div className="w-10 h-10 rounded-full bg-bg-secondary animate-pulse" />
          <div className="flex-1 space-y-2">
            <div className="h-4 bg-bg-secondary rounded w-32 animate-pulse" />
            <div className="h-3 bg-bg-secondary rounded w-24 animate-pulse" />
          </div>
        </div>
      ))}
    </>
  );
}

// Compact stats for my profile sidebar
export function MyFollowStats() {
  const { data: stats, isLoading } = useMyFollowStats();

  if (isLoading) {
    return (
      <div className="grid grid-cols-3 gap-4 py-2">
        {[1, 2, 3].map((i) => (
          <div key={i} className="text-center">
            <div className="h-6 bg-bg-secondary rounded animate-pulse mb-1" />
            <div className="h-3 bg-bg-secondary rounded w-16 mx-auto animate-pulse" />
          </div>
        ))}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-3 gap-4 py-2">
      <div className="text-center">
        <p className="text-lg font-bold text-text-primary">{stats?.followers ?? 0}</p>
        <p className="text-xs text-text-secondary">Followers</p>
      </div>
      <div className="text-center">
        <p className="text-lg font-bold text-text-primary">{stats?.following ?? 0}</p>
        <p className="text-xs text-text-secondary">Following</p>
      </div>
      <div className="text-center">
        <p className="text-lg font-bold text-text-primary">{stats?.functions_followed ?? 0}</p>
        <p className="text-xs text-text-secondary">Functions</p>
      </div>
    </div>
  );
}
