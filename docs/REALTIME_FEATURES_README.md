# Real-time Features with Supabase

This guide demonstrates how to implement real-time user data features using Supabase subscriptions in your FunctionFly application.

## Overview

The real-time implementation provides:
- **Live Notifications**: Instant notification delivery and status updates
- **User Presence**: Real-time online/offline status tracking
- **Profile Updates**: Live profile change broadcasts
- **Activity Feeds**: Combined real-time activity streams
- **Connection Status**: Real-time connection health monitoring

## Backend Implementation

### 1. Supabase Client Setup

```go
// Initialize Supabase client
config := &supabase.Config{
    URL:       "https://your-project.supabase.co",
    AnonKey:   "your-anon-key",
    ServiceKey: "your-service-key",
}

client, err := supabase.NewClient(config)
if err != nil {
    log.Fatal(err)
}
```

### 2. Real-time Subscriptions

```go
// Create real-time handler
handler := supabase.NewRealtimeHandler(client)

// Subscribe to user notifications
err := handler.SubscribeToUserNotifications(context.Background(), userID, func(event supabase.RealtimeEvent) {
    log.Printf("New notification: %s", event.Data["title"])
})

// Subscribe to tenant user changes
err = handler.SubscribeToTenantUsers(context.Background(), tenantID, func(event supabase.RealtimeEvent) {
    log.Printf("User status changed: %s", event.Type)
})
```

### 3. Broadcasting Events

```go
// Broadcast a notification
notification := &supabase.UserNotification{
    UserID:    userID,
    Type:      "success",
    Title:     "Welcome!",
    Message:   "Your account has been verified",
}

err := client.CreateNotification(context.Background(), *notification)
// This automatically triggers real-time broadcasts
```

## Frontend Implementation

### 1. Supabase Client Configuration

```typescript
// lib/supabase.ts
import { createClient } from '@supabase/supabase-js';

const supabase = createClient(
  import.meta.env.VITE_SUPABASE_URL,
  import.meta.env.VITE_SUPABASE_ANON_KEY,
  {
    realtime: {
      params: {
        eventsPerSecond: 10,
      },
    },
  }
);

export { supabase };
```

### 2. Real-time Hooks

#### User Notifications

```tsx
import { useUserNotifications } from '../hooks/useRealtime';

function NotificationComponent() {
  const {
    notifications,
    unreadCount,
    isConnected,
    markAsRead,
    markAllAsRead
  } = useUserNotifications();

  return (
    <div>
      <div className="flex items-center space-x-2">
        <Bell className="w-5 h-5" />
        <span>{unreadCount} unread</span>
        {!isConnected && <span className="text-yellow-500">Offline</span>}
      </div>

      {notifications.map(notification => (
        <div key={notification.notification_id}>
          <h4>{notification.title}</h4>
          <p>{notification.message}</p>
          <button onClick={() => markAsRead(notification.notification_id)}>
            Mark as read
          </button>
        </div>
      ))}
    </div>
  );
}
```

#### User Presence

```tsx
import { useUserPresence, useUserOnlineStatus } from '../hooks/useRealtime';

function PresenceComponent() {
  const { onlineUsers, isConnected } = useUserPresence();
  const { isOnline } = useUserOnlineStatus('user-id');

  return (
    <div>
      <div className={`w-3 h-3 rounded-full ${isOnline ? 'bg-green-500' : 'bg-gray-400'}`} />
      <span>{onlineUsers.length} users online</span>
      {!isConnected && <span>Real-time updates paused</span>}
    </div>
  );
}
```

#### Profile Updates

```tsx
import { useProfileUpdates } from '../hooks/useRealtime';

function ProfileComponent() {
  const { profileUpdates } = useProfileUpdates();

  useEffect(() => {
    profileUpdates.forEach(update => {
      if (update.changes.avatar_url) {
        // Refresh avatar
      }
    });
  }, [profileUpdates]);

  return <div>Profile updates: {profileUpdates.length}</div>;
}
```

### 3. Real-time Components

#### Notification Center

```tsx
import { RealtimeNotificationCenter } from '../components/realtime/RealtimeNotificationCenter';

function Dashboard() {
  return (
    <div>
      <header>
        <RealtimeNotificationCenter />
      </header>
      {/* Dashboard content */}
    </div>
  );
}
```

#### User Presence Indicators

```tsx
import { UserPresenceIndicator, OnlineUsersList } from '../components/realtime/UserPresenceIndicator';

function TeamPage() {
  return (
    <div>
      <OnlineUsersList maxVisible={5} showCount={true} />

      <div className="team-members">
        {teamMembers.map(member => (
          <UserPresenceIndicator
            key={member.id}
            userId={member.id}
            userName={member.name}
            userEmail={member.email}
            avatarUrl={member.avatarUrl}
            showDetails={true}
          />
        ))}
      </div>
    </div>
  );
}
```

## Real-time Event Types

### User Notifications

```typescript
interface NewNotificationEvent {
  type: 'new_notification';
  notification_id: string;
  notification_type: 'info' | 'warning' | 'error' | 'success';
  title: string;
  timestamp: string;
}
```

### Profile Updates

```typescript
interface ProfileUpdateEvent {
  type: 'profile_update';
  user_id: string;
  changes: {
    first_name?: boolean;
    last_name?: boolean;
    avatar_url?: boolean;
    bio?: boolean;
  };
  timestamp: string;
}
```

### User Status Changes

```typescript
interface UserStatusChangeEvent {
  type: 'user_status_change';
  user_id: string;
  tenant_id: string;
  old_status: 'verified' | 'unverified';
  new_status: 'verified' | 'unverified';
  timestamp: string;
}
```

### Presence Events

```typescript
interface PresenceEvent {
  type: 'presence_join' | 'presence_leave';
  user_id: string;
  timestamp: string;
}
```

## Database Schema

### Real-time Tables

```sql
-- Enable real-time on user tables
ALTER PUBLICATION supabase_realtime ADD TABLE users, user_profiles, user_notifications;

-- RLS Policies ensure users only see their own data
CREATE POLICY "Users can view own notifications" ON user_notifications
  FOR SELECT USING (auth.uid() = user_id);
```

### Triggers for Broadcasting

```sql
-- Function to broadcast notification creation
CREATE OR REPLACE FUNCTION broadcast_new_notification()
RETURNS TRIGGER AS $$
BEGIN
  PERFORM pg_notify(
    'user_' || NEW.user_id || '_notifications',
    jsonb_build_object(
      'type', 'new_notification',
      'notification_id', NEW.id,
      'title', NEW.title,
      'timestamp', NOW()
    )::text
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for notifications
CREATE TRIGGER broadcast_notification_trigger
  AFTER INSERT ON user_notifications
  FOR EACH ROW EXECUTE FUNCTION broadcast_new_notification();
```

## Best Practices

### 1. Connection Management

```typescript
// Monitor connection status
const { connectionStatus } = useRealtimeConnection();

if (connectionStatus === 'disconnected') {
  // Show offline indicator
  // Implement retry logic
}
```

### 2. Error Handling

```typescript
// Handle subscription errors
useEffect(() => {
  const subscription = supabase
    .channel('user_updates')
    .on('broadcast', handleEvent)
    .subscribe((status) => {
      if (status === 'SUBSCRIPTION_ERROR') {
        // Handle error, maybe retry
        console.error('Subscription failed');
      }
    });

  return () => subscription.unsubscribe();
}, []);
```

### 3. Performance Optimization

```typescript
// Limit subscriptions per user
const MAX_SUBSCRIPTIONS = 10;

// Debounce rapid updates
const debouncedUpdate = useMemo(
  () => debounce((data) => updateUI(data), 100),
  []
);
```

### 4. Memory Management

```typescript
// Clean up subscriptions on unmount
useEffect(() => {
  const subscriptions = [
    subscribeToNotifications(),
    subscribeToPresence(),
  ];

  return () => {
    subscriptions.forEach(sub => sub.unsubscribe());
  };
}, []);
```

## Environment Variables

```bash
# Frontend
VITE_SUPABASE_URL=https://your-project.supabase.co
VITE_SUPABASE_ANON_KEY=your-anon-key

# Backend (optional, if using service key)
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_SERVICE_KEY=your-service-key
```

## Testing Real-time Features

### Unit Tests

```typescript
describe('useUserNotifications', () => {
  it('should handle new notifications', () => {
    // Mock Supabase client
    // Test notification state updates
  });
});
```

### Integration Tests

```typescript
// Test real-time broadcasts
test('notification broadcast', async () => {
  const { notifications } = renderHook(() => useUserNotifications());

  // Simulate notification creation
  await supabase.from('user_notifications').insert(testNotification);

  // Assert notification appears
  await waitFor(() => {
    expect(notifications.current).toContain(testNotification);
  });
});
```

## Troubleshooting

### Common Issues

1. **Subscriptions not working**
   - Check RLS policies
   - Verify publication includes the table
   - Check network connectivity

2. **Performance issues**
   - Too many subscriptions per user
   - Frequent broadcasts without throttling
   - Large payload sizes

3. **Connection drops**
   - Implement reconnection logic
   - Show offline indicators
   - Cache offline changes

### Debug Tools

```typescript
// Enable debug logging
const supabase = createClient(url, key, {
  realtime: {
    logger: console.log,
  },
});

// Monitor subscription status
channel.subscribe((status) => {
  console.log('Subscription status:', status);
});
```

## Migration Path

1. **Phase 1**: Add Supabase alongside existing database
2. **Phase 2**: Implement real-time features with feature flags
3. **Phase 3**: Gradual migration of read operations
4. **Phase 4**: Switch to Supabase as primary database
5. **Phase 5**: Remove old database integration

This implementation provides a robust real-time user data system with Supabase while maintaining backward compatibility during migration.