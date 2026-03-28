import { useState } from "react";
import { UserPlus, UserMinus, Loader2, Bell, BellOff, Users } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  useUserFollowStatus,
  useFollowUser,
  useUnfollowUser,
  useFunctionFollowStatus,
  useFollowFunction,
  useUnfollowFunction,
} from "@/hooks/useFollow";

interface FollowUserButtonProps {
  username: string;
  variant?: "default" | "outline" | "ghost";
  size?: "default" | "sm" | "lg";
  showDropdown?: boolean;
  className?: string;
}

export function FollowUserButton({
  username,
  variant = "default",
  size = "default",
  showDropdown = true,
  className,
}: FollowUserButtonProps) {
  const { data: status, isLoading: isLoadingStatus } = useUserFollowStatus(username);
  const followMutation = useFollowUser(username);
  const unfollowMutation = useUnfollowUser(username);
  const [isOpen, setIsOpen] = useState(false);

  const isFollowing = status?.is_following ?? false;
  const isLoading = isLoadingStatus || followMutation.isPending || unfollowMutation.isPending;

  const handleFollow = () => {
    followMutation.mutate(undefined);
  };

  const handleUnfollow = () => {
    unfollowMutation.mutate();
    setIsOpen(false);
  };

  if (isLoadingStatus) {
    return (
      <Button variant={variant} size={size} disabled className={cn("gap-2", className)}>
        <Loader2 className="w-4 h-4 animate-spin mr-2" />
        Loading...
      </Button>
    );
  }

  if (!isFollowing) {
    return (
      <Button
        variant={variant}
        size={size}
        onClick={handleFollow}
        disabled={followMutation.isPending}
        className={cn("gap-2 bg-brand-600 hover:bg-brand-700", className)}
      >
        {followMutation.isPending ? (
          <Loader2 className="w-4 h-4 animate-spin" />
        ) : (
          <UserPlus className="w-4 h-4" />
        )}
        Follow
      </Button>
    );
  }

  // User is following - show unfollow dropdown or simple button
  if (showDropdown) {
    return (
      <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size={size} className={cn("gap-2", className)}>
            <Bell className="w-4 h-4 mr-2" />
            Following
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onClick={handleUnfollow} className="text-destructive">
            <UserMinus className="w-4 h-4 mr-2" />
            Unfollow
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    );
  }

  return (
    <Button
      variant="outline"
      size={size}
      onClick={handleUnfollow}
      disabled={unfollowMutation.isPending}
      className={cn("gap-2 border-brand-500 text-brand-400", className)}
    >
      {unfollowMutation.isPending ? (
        <Loader2 className="w-4 h-4 animate-spin" />
      ) : (
        <Users className="w-4 h-4" />
      )}
      Following
    </Button>
  );
}

interface FollowFunctionButtonProps {
  functionId: string;
  functionName?: string;
  variant?: "default" | "outline" | "ghost";
  size?: "default" | "sm" | "lg";
}

export function FollowFunctionButton({
  functionId,
  functionName,
  variant = "outline",
  size = "sm",
}: FollowFunctionButtonProps) {
  const { data: status, isLoading: isLoadingStatus } = useFunctionFollowStatus(functionId);
  const followMutation = useFollowFunction(functionId);
  const unfollowMutation = useUnfollowFunction(functionId);

  const isFollowing = status?.is_following ?? false;
  const followerCount = status?.follower_count ?? 0;
  const isLoading = isLoadingStatus || followMutation.isPending || unfollowMutation.isPending;

  const handleToggle = () => {
    if (isFollowing) {
      unfollowMutation.mutate();
    } else {
      followMutation.mutate(undefined);
    }
  };

  if (isLoadingStatus) {
    return (
      <Button variant={variant} size={size} disabled>
        <Loader2 className="w-3.5 h-3.5 animate-spin" />
      </Button>
    );
  }

  return (
    <Button
      variant={isFollowing ? "default" : variant}
      size={size}
      onClick={handleToggle}
      disabled={isLoading}
      className={isFollowing ? "bg-brand-600 hover:bg-brand-700" : ""}
    >
      {isLoading ? (
        <Loader2 className="w-3.5 h-3.5 animate-spin mr-1.5" />
      ) : isFollowing ? (
        <Bell className="w-3.5 h-3.5 mr-1.5" />
      ) : (
        <BellOff className="w-3.5 h-3.5 mr-1.5" />
      )}
      <span className="mr-1">{isFollowing ? "Following" : "Follow"}</span>
      {followerCount > 0 && (
        <span className="text-xs opacity-80">({followerCount})</span>
      )}
    </Button>
  );
}
