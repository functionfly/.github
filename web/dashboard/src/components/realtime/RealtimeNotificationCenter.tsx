import React, { useState, useEffect } from 'react';
import { Bell, Check, CheckCheck, X, AlertCircle, Info, AlertTriangle, CheckCircle } from 'lucide-react';
import { useUserNotifications } from '../../hooks/useRealtime';
import { Button } from '../common/Button';
import { Card } from '../common/Card';
import { Badge } from '../common/Badge';

interface NotificationItemProps {
  notification: {
    notification_id: string;
    type: 'info' | 'warning' | 'error' | 'success';
    title: string;
    message?: string;
    created_at: string;
    read_at?: string;
  };
  onMarkAsRead: (id: string) => void;
  onDismiss: (id: string) => void;
}

function NotificationItem({ notification, onMarkAsRead, onDismiss }: NotificationItemProps) {
  const getIcon = () => {
    switch (notification.type) {
      case 'error':
        return <AlertCircle className="w-5 h-5 text-red-500" />;
      case 'warning':
        return <AlertTriangle className="w-5 h-5 text-yellow-500" />;
      case 'success':
        return <CheckCircle className="w-5 h-5 text-green-500" />;
      case 'info':
      default:
        return <Info className="w-5 h-5 text-blue-500" />;
    }
  };

  const getBorderColor = () => {
    switch (notification.type) {
      case 'error':
        return 'border-red-200';
      case 'warning':
        return 'border-yellow-200';
      case 'success':
        return 'border-green-200';
      case 'info':
      default:
        return 'border-blue-200';
    }
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / (1000 * 60));
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className={`p-4 border-l-4 ${getBorderColor()} bg-white rounded-lg shadow-sm hover:shadow-md transition-shadow`}>
      <div className="flex items-start justify-between">
        <div className="flex items-start space-x-3">
          {getIcon()}
          <div className="flex-1 min-w-0">
            <h4 className={`text-sm font-medium ${notification.read_at ? 'text-gray-600' : 'text-gray-900'}`}>
              {notification.title}
            </h4>
            {notification.message && (
              <p className={`text-sm mt-1 ${notification.read_at ? 'text-gray-500' : 'text-gray-700'}`}>
                {notification.message}
              </p>
            )}
            <p className="text-xs text-gray-400 mt-2">
              {formatTimestamp(notification.created_at)}
            </p>
          </div>
        </div>
        <div className="flex items-center space-x-1 ml-4">
          {!notification.read_at && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onMarkAsRead(notification.notification_id)}
              className="p-1 h-8 w-8"
            >
              <Check className="w-4 h-4" />
            </Button>
          )}
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onDismiss(notification.notification_id)}
            className="p-1 h-8 w-8 text-gray-400 hover:text-gray-600"
          >
            <X className="w-4 h-4" />
          </Button>
        </div>
      </div>
    </div>
  );
}

interface RealtimeNotificationCenterProps {
  className?: string;
}

export function RealtimeNotificationCenter({ className = '' }: RealtimeNotificationCenterProps) {
  const {
    notifications,
    unreadCount,
    isConnected,
    markAsRead,
    markAllAsRead,
  } = useUserNotifications();

  const [isOpen, setIsOpen] = useState(false);
  const [showUnreadOnly, setShowUnreadOnly] = useState(false);

  const displayedNotifications = showUnreadOnly
    ? notifications.filter(n => !n.read_at)
    : notifications;

  const handleMarkAsRead = async (id: string) => {
    await markAsRead(id);
  };

  const handleDismiss = async (id: string) => {
    // For now, just mark as read. In a real app, you might have a separate dismiss action
    await markAsRead(id);
  };

  const handleMarkAllAsRead = async () => {
    await markAllAsRead();
  };

  return (
    <div className={`relative ${className}`}>
      {/* Notification Bell Button */}
      <Button
        variant="ghost"
        size="sm"
        onClick={() => setIsOpen(!isOpen)}
        className="relative p-2"
      >
        <Bell className="w-5 h-5" />
        {unreadCount > 0 && (
          <Badge
            variant="danger"
            className="absolute -top-1 -right-1 min-w-[1.25rem] h-5 flex items-center justify-center text-xs"
          >
            {unreadCount > 99 ? '99+' : unreadCount}
          </Badge>
        )}
        {!isConnected && (
          <div className="absolute -bottom-1 -right-1 w-2 h-2 bg-yellow-400 rounded-full" />
        )}
      </Button>

      {/* Notification Dropdown */}
      {isOpen && (
        <Card className="absolute right-0 top-12 w-96 max-h-96 overflow-hidden shadow-lg z-50">
          <div className="p-4 border-b border-gray-200">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold">Notifications</h3>
              <div className="flex items-center space-x-2">
                <Badge variant={isConnected ? 'success' : 'warning'} className="text-xs">
                  {isConnected ? 'Live' : 'Offline'}
                </Badge>
              </div>
            </div>

            <div className="flex items-center justify-between mt-3">
              <div className="flex items-center space-x-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowUnreadOnly(!showUnreadOnly)}
                  className={`text-xs ${showUnreadOnly ? 'bg-blue-50 text-blue-700' : ''}`}
                >
                  Unread only
                </Button>
                <span className="text-sm text-gray-500">
                  {displayedNotifications.length} notification{displayedNotifications.length !== 1 ? 's' : ''}
                </span>
              </div>

              {unreadCount > 0 && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={handleMarkAllAsRead}
                  className="text-xs text-blue-600 hover:text-blue-800"
                >
                  <CheckCheck className="w-4 h-4 mr-1" />
                  Mark all read
                </Button>
              )}
            </div>
          </div>

          <div className="max-h-80 overflow-y-auto">
            {displayedNotifications.length === 0 ? (
              <div className="p-6 text-center text-gray-500">
                <Bell className="w-8 h-8 mx-auto mb-2 text-gray-300" />
                <p className="text-sm">
                  {showUnreadOnly ? 'No unread notifications' : 'No notifications yet'}
                </p>
              </div>
            ) : (
              <div className="divide-y divide-gray-100">
                {displayedNotifications.map((notification) => (
                  <div key={notification.notification_id} className="p-2">
                    <NotificationItem
                      notification={notification}
                      onMarkAsRead={handleMarkAsRead}
                      onDismiss={handleDismiss}
                    />
                  </div>
                ))}
              </div>
            )}
          </div>

          {!isConnected && (
            <div className="p-3 bg-yellow-50 border-t border-yellow-200">
              <p className="text-xs text-yellow-800 text-center">
                Real-time updates paused - check your connection
              </p>
            </div>
          )}
        </Card>
      )}

      {/* Click outside to close */}
      {isOpen && (
        <div
          className="fixed inset-0 z-40"
          onClick={() => setIsOpen(false)}
        />
      )}
    </div>
  );
}