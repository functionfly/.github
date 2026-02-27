import React, { useEffect, useState } from 'react';
import { useUserPresenceInfo, useUserPresenceDetails, useUserPresence } from '../../hooks/usePresence';
import { Avatar } from '../common/Avatar';
import { Badge } from '../common/Badge';
import { Tooltip } from '../common/Tooltip';

interface UserPresenceIndicatorProps {
  userId: string;
  userName: string;
  userEmail: string;
  avatarUrl?: string;
  size?: 'sm' | 'md' | 'lg';
  showDetails?: boolean;
}

export function UserPresenceIndicator({
  userId,
  userName,
  userEmail,
  avatarUrl,
  size = 'md',
  showDetails = false,
}: UserPresenceIndicatorProps) {
  const { isOnline, isRealtimeEnabled, lastSeen, loading } = useUserPresenceInfo(userId);
  const { isConnected } = useUserPresence(); // For connection status

  const sizeClasses = {
    sm: 'w-6 h-6',
    md: 'w-8 h-8',
    lg: 'w-10 h-10',
  };

  const statusIndicator = (
    <div className={`absolute -bottom-0.5 -right-0.5 ${size === 'sm' ? 'w-2.5 h-2.5' : 'w-3 h-3'} rounded-full border-2 border-white ${
      isOnline ? 'bg-green-500' : 'bg-gray-400'
    } ${!isConnected ? 'opacity-50' : ''}`} />
  );

  if (showDetails) {
    return (
      <div className="flex items-center space-x-3 p-2 rounded-lg hover:bg-gray-50 transition-colors">
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
          <p className="text-sm font-medium text-gray-900 truncate">
            {userName}
          </p>
          <p className="text-xs text-gray-500 truncate">
            {userEmail}
          </p>
          <div className="flex items-center mt-1">
            <Badge
              variant={isOnline ? 'success' : 'secondary'}
              className="text-xs"
            >
              {loading ? 'Loading...' : isOnline ? 'Online' : 'Offline'}
            </Badge>
            {lastSeen && !isOnline && !loading && (
              <span className="text-xs text-gray-400 ml-2">
                Last seen {lastSeen}
              </span>
            )}
          </div>
        </div>
      </div>
    );
  }

  return (
    <Tooltip content={
      <div className="text-center">
        <div className="font-medium">{userName}</div>
        <div className="text-sm text-gray-300">{userEmail}</div>
        <div className="text-xs mt-1">
          {loading ? (
            <span className="text-gray-400">Loading...</span>
          ) : isOnline ? (
            <span className="text-green-400">● Online</span>
          ) : (
            <span className="text-gray-400">
              ● Offline {lastSeen && `• Last seen ${lastSeen}`}
            </span>
          )}
        </div>
        {!isRealtimeEnabled && (
          <div className="text-xs text-yellow-400 mt-1">
            Real-time updates paused
          </div>
        )}
      </div>
    }>
      <div className="relative inline-block">
        <Avatar
          src={avatarUrl}
          alt={userName}
          fallback={userName.charAt(0).toUpperCase()}
          className={sizeClasses[size]}
        />
        {loading ? (
          <div className="absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-white bg-gray-400 animate-pulse" />
        ) : (
          statusIndicator
        )}
      </div>
    </Tooltip>
  );
}

// Component to show a list of online users
interface OnlineUsersListProps {
  className?: string;
  maxVisible?: number;
  showCount?: boolean;
}

export function OnlineUsersList({ className = '', maxVisible = 5, showCount = true }: OnlineUsersListProps) {
  const { onlineUsers, isConnected } = useUserPresence();
  const { userDetails, loading, error } = useUserPresenceDetails(onlineUsers);

  const visibleUsers = userDetails.slice(0, maxVisible);
  const hiddenCount = Math.max(0, onlineUsers.length - maxVisible);

  return (
    <div className={`flex items-center space-x-2 ${className}`}>
      <div className="flex -space-x-2">
        {visibleUsers.map((user) => (
          <UserPresenceIndicator
            key={user.id}
            userId={user.id}
            userName={user.name}
            userEmail={user.email}
            avatarUrl={user.avatar_url}
            size="sm"
          />
        ))}
      </div>

      {hiddenCount > 0 && (
        <Tooltip content={`${hiddenCount} more user${hiddenCount !== 1 ? 's' : ''} online`}>
          <div className="w-6 h-6 rounded-full bg-gray-200 border-2 border-white flex items-center justify-center text-xs font-medium text-gray-600">
            +{hiddenCount}
          </div>
        </Tooltip>
      )}

      {showCount && (
        <div className="flex items-center space-x-2 text-sm text-gray-600">
          <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-yellow-500'}`} />
          <span>
            {loading ? 'Loading...' : error ? 'Error loading users' : `${onlineUsers.length} online`}
          </span>
        </div>
      )}
    </div>
  );
}

// Hook to get online status for a specific user
export function useUserOnlineStatus(userId: string) {
  const { onlineUsers, isConnected } = useUserPresence();

  return {
    isOnline: onlineUsers.includes(userId),
    isRealtimeEnabled: isConnected,
  };
}