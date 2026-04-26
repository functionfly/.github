import React from 'react';
import { usePresence, useUserOnlineStatus, useTenantOnlineUsers } from '../../hooks/usePresence';
import { Avatar } from '../common/Avatar';
import { Badge } from '../common/Badge';
import { Tooltip } from '../common/Tooltip';
import { cn } from '@/lib/utils';

interface UserPresenceIndicatorProps {
  userId: string;
  userName: string;
  userEmail: string;
  avatarUrl?: string;
  size?: 'sm' | 'md' | 'lg';
  showDetails?: boolean;
  showLastSeen?: boolean;
}

export function UserPresenceIndicator({
  userId,
  userName,
  userEmail,
  avatarUrl,
  size = 'md',
  showDetails = false,
  showLastSeen = true,
}: UserPresenceIndicatorProps) {
  const { isOnline, isAway, status } = useUserOnlineStatus(userId);
  const { isConnected } = usePresence();

  const sizeClasses = {
    sm: 'w-6 h-6',
    md: 'w-8 h-8',
    lg: 'w-10 h-10',
  };

  const statusColor = status === 'online' ? 'bg-emerald-500' : status === 'away' ? 'bg-amber-500' : 'bg-slate-400';

  const statusIndicator = (
    <div
      className={cn(
        'absolute -bottom-0.5 -right-0.5 rounded-full border-2 border-white',
        size === 'sm' ? 'w-2.5 h-2.5' : 'w-3 h-3',
        statusColor,
        !isConnected && 'opacity-50'
      )}
    />
  );

  if (showDetails) {
    return (
      <div className="flex items-center gap-3 p-2 rounded-lg hover:bg-bg-hover transition-colors">
        <div className="relative">
          <Avatar
            src={avatarUrl}
            alt={userName}
            fallback={userName.charAt(0).toUpperCase()}
            className={sizeClasses[size]}
          />
          {statusIndicator}
        </div>

        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium text-text-primary truncate">
            {userName}
          </p>
          <p className="text-xs text-text-muted truncate">
            {userEmail}
          </p>
          <div className="flex items-center gap-2 mt-1">
            <Badge
              variant={isOnline ? 'success' : isAway ? 'warning' : 'secondary'}
              className="text-xs"
            >
              {isOnline ? 'Online' : isAway ? 'Away' : 'Offline'}
            </Badge>
          </div>
        </div>
      </div>
    );
  }

  const tooltipContent = (
    <div className="text-center">
      <div className="font-medium">{userName}</div>
      <div className="text-sm text-text-muted">{userEmail}</div>
      <div className="text-xs mt-1">
        {isOnline ? (
          <span className="text-emerald-400">● Online</span>
        ) : isAway ? (
          <span className="text-amber-400">● Away</span>
        ) : (
          <span className="text-text-muted">● Offline</span>
        )}
      </div>
    </div>
  );

  return (
    <Tooltip content={tooltipContent}>
      <div className="relative inline-block">
        <Avatar
          src={avatarUrl}
          alt={userName}
          fallback={userName.charAt(0).toUpperCase()}
          className={sizeClasses[size]}
        />
        {statusIndicator}
      </div>
    </Tooltip>
  );
}

interface OnlineUsersListProps {
  className?: string;
  maxVisible?: number;
  showCount?: boolean;
}

export function OnlineUsersList({ className = '', maxVisible = 5, showCount = true }: OnlineUsersListProps) {
  const { onlineUsers, isConnected, isLoading, error } = useTenantOnlineUsers();

  const visibleUsers = onlineUsers.slice(0, maxVisible);
  const hiddenCount = Math.max(0, onlineUsers.length - maxVisible);

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <div className="flex -space-x-2">
        {visibleUsers.map((user) => (
          <UserPresenceIndicator
            key={user.userId}
            userId={user.userId}
            userName={user.displayName || user.username || 'User'}
            userEmail=""
            avatarUrl={user.avatar}
            size="sm"
          />
        ))}
      </div>

      {hiddenCount > 0 && (
        <Tooltip content={`${hiddenCount} more user${hiddenCount !== 1 ? 's' : ''} online`}>
          <div className="w-6 h-6 rounded-full bg-bg-secondary border-2 border-bg-primary flex items-center justify-center text-xs font-medium text-text-secondary">
            +{hiddenCount}
          </div>
        </Tooltip>
      )}

      {showCount && (
        <div className="flex items-center gap-2 text-sm text-text-secondary">
          <div className={cn('w-2 h-2 rounded-full', isConnected ? 'bg-emerald-500' : 'bg-amber-500')} />
          <span>
            {isLoading ? 'Loading...' : error ? 'Error loading users' : `${onlineUsers.length} online`}
          </span>
        </div>
      )}
    </div>
  );
}
