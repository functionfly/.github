import { createClient, SupabaseAuthAdapter } from '@neondatabase/neon-js';
import { getApiBaseUrl } from '@/lib/constants';

// Neon Auth configuration — BetterAuth requires an absolute URL (relative paths like /api/v1/auth throw)
function resolveAuthUrl(): string {
  const raw = getApiBaseUrl();
  if (raw.startsWith('http://') || raw.startsWith('https://')) return raw;
  const origin = typeof window !== 'undefined' ? window.location.origin : (import.meta.env.VITE_APP_ORIGIN ||
    (() => { throw new Error('VITE_APP_ORIGIN environment variable is required'); })());
  return `${origin}${raw.startsWith('/') ? raw : `/${raw}`}`;
}

const neonAuthUrl = resolveAuthUrl();
const neonDataApiUrl = import.meta.env.VITE_NEON_DATA_API_URL;

// Neon is production-only; in dev, use the Go API backend directly
const client = neonDataApiUrl ? createClient({
  auth: {
    url: neonAuthUrl,
    adapter: SupabaseAuthAdapter(),
  },
  dataApi: {
    url: neonDataApiUrl,
  },
}) : null;

if (!neonDataApiUrl) {
  console.warn('[neon] VITE_NEON_DATA_API_URL not set — Neon client disabled. Using Go API backend for all data.');
}

// Legacy alias for backward compatibility during migration
export const supabase = client;

function requireNeon(): any {
  if (!client) throw new Error('Neon client not initialized — set VITE_NEON_DATA_API_URL for production');
  return client;
}

// Helper function to get current user session
export const getCurrentUser = async () => {
  const { data: { user }, error } = await requireNeon().auth.getUser();
  if (error) throw error;
  return user;
};

// Helper function to get current session
export const getCurrentSession = async () => {
  const { data: { session }, error } = await requireNeon().auth.getSession();
  if (error) throw error;
  return session;
};

// Helper function to sign out
export const signOut = async () => {
  const { error } = await requireNeon().auth.signOut();
  if (error) throw error;
};

// Database query helpers using Neon Data API
export const getUserProfile = async (userId: string) => {
  const { data, error } = await requireNeon()
    .from('user_profiles')
    .select('*')
    .eq('user_id', userId)
    .single();

  if (error) throw error;
  return data;
};

export const updateUserProfile = async (userId: string, updates: any) => {
  const { data, error } = await requireNeon()
    .from('user_profiles')
    .update({ ...updates, updated_at: new Date().toISOString() })
    .eq('user_id', userId)
    .select()
    .single();

  if (error) throw error;
  return data;
};

export const getUserNotifications = async (userId: string, includeRead = false) => {
  let query = requireNeon()
    .from('user_notifications')
    .select('*')
    .eq('user_id', userId)
    .order('created_at', { ascending: false });

  if (!includeRead) {
    query = query.is('read_at', null);
  }

  const { data, error } = await query;
  if (error) throw error;
  return data;
};

export const markNotificationRead = async (notificationId: string, userId: string) => {
  const { data, error } = await requireNeon()
    .from('user_notifications')
    .update({ read_at: new Date().toISOString() })
    .eq('id', notificationId)
    .eq('user_id', userId)
    .select()
    .single();

  if (error) throw error;
  return data;
};

export const getOnlineUsers = async (tenantId?: string) => {
  const { data, error } = await requireNeon().rpc('get_online_users', {
    tenant_filter: tenantId || null,
  });

  if (error) throw error;
  return data;
};

export const getUsersByIds = async (userIds: string[]) => {
  if (userIds.length === 0) return [];

  const { data, error } = await requireNeon()
    .from('users')
    .select(`
      id,
      email,
      profile:user_profiles(
        first_name,
        last_name,
        avatar_url,
        last_active_at
      )
    `)
    .in('id', userIds);

  if (error) throw error;

  // Transform the data to match the expected format (profile is an array from the join)
  return data.map(user => {
    const profile = Array.isArray(user.profile) ? user.profile[0] : user.profile;
    return {
      id: user.id,
      name: profile?.first_name && profile?.last_name
        ? `${profile.first_name} ${profile.last_name}`
        : user.email.split('@')[0], // Fallback to email username
      email: user.email,
      avatar_url: profile?.avatar_url,
      last_active_at: profile?.last_active_at,
    };
  });
};

export const getUserLastActive = async (userId: string) => {
  const { data, error } = await requireNeon()
    .from('user_profiles')
    .select('last_active_at')
    .eq('user_id', userId)
    .single();

  if (error) {
    // If no profile exists, return null
    if (error.code === 'PGRST116') return null;
    throw error;
  }

  return data?.last_active_at;
};

export const createNotification = async (notification: {
  user_id: string;
  type: string;
  title: string;
  message?: string;
  data?: any;
  expires_at?: string;
}) => {
  const { data, error } = await requireNeon()
    .from('user_notifications')
    .insert([notification])
    .select()
    .single();

  if (error) throw error;
  return data;
};

// WebSocket-based real-time subscription helpers for Neon
export const subscribeToUserNotifications = (userId: string, callback: (payload: any) => void) => {
  // This is now handled by the useRealtimeSubscription hook
  // Individual hooks should use the generic useRealtimeSubscription hook
  console.warn('Use useRealtimeSubscription hook instead of subscribeToUserNotifications');
  return { unsubscribe: () => {} };
};

export const subscribeToUserProfile = (userId: string, callback: (payload: any) => void) => {
  // This is now handled by the useRealtimeSubscription hook
  // Individual hooks should use the generic useRealtimeSubscription hook
  return { unsubscribe: () => {} };
};

export const subscribeToTenantUsers = (tenantId: string, callback: (payload: any) => void) => {
  // This is now handled by the useRealtimeSubscription hook
  // Individual hooks should use the generic useRealtimeSubscription hook
  return { unsubscribe: () => {} };
};

export const subscribeToUserPresence = (tenantId: string, callback: (payload: any) => void) => {
  // This is now handled by the useRealtimeSubscription hook
  // Individual hooks should use the generic useRealtimeSubscription hook
  return {
    unsubscribe: () => {},
    track: () => {},
    untrack: () => {}
  };
};

// WebSocket-based real-time channel management for Neon
export class RealtimeManager {
  private connections: Map<string, WebSocket> = new Map();

  subscribe(channelName: string, config: {
    onEvent?: (payload: any) => void;
    onPresence?: (event: string, payload: any) => void;
  }) {
    const baseUrl = getApiBaseUrl();
    const realtimePath = '/v1/monitoring/realtime';
    const fullWsUrl = baseUrl.startsWith('http')
      ? `${baseUrl.replace(/^http/, 'ws').replace(/\/$/, '')}${realtimePath}`
      : `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}${baseUrl.startsWith('/') ? baseUrl : `/${baseUrl || 'api'}`}${realtimePath}`;

    try {
      const ws = new WebSocket(fullWsUrl);

      ws.onopen = () => {
        // Subscribe to the channel
        ws.send(JSON.stringify({
          type: 'subscribe',
          channel: channelName,
        }));
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);

          if (data.type === 'broadcast' && config.onEvent) {
            config.onEvent(data.payload);
          }

          if ((data.type === 'presence_join' || data.type === 'presence_leave') && config.onPresence) {
            config.onPresence(data.type, data.payload);
          }
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

      this.connections.set(channelName, ws);

      return {
        unsubscribe: () => this.unsubscribe(channelName),
        send: (event: string, payload: any) => this.broadcast(channelName, event, payload),
        track: (payload: any) => {
          if (ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({
              type: 'presence',
              event: 'track',
              payload,
              channel: channelName,
            }));
          }
        },
        untrack: () => {
          if (ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({
              type: 'presence',
              event: 'untrack',
              channel: channelName,
            }));
          }
        }
      };
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      return {
        unsubscribe: () => {},
        send: () => {},
        track: () => {},
        untrack: () => {}
      };
    }
  }

  unsubscribe(channelName: string) {
    const ws = this.connections.get(channelName);
    if (ws) {
      ws.send(JSON.stringify({
        type: 'unsubscribe',
        channel: channelName,
      }));
      ws.close();
      this.connections.delete(channelName);
    }
  }

  broadcast(channelName: string, event: string, payload: any) {
    const ws = this.connections.get(channelName);
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        type: 'broadcast',
        event,
        payload,
        channel: channelName,
      }));
    }
  }

  cleanup() {
    for (const [channelName, ws] of this.connections) {
      ws.close();
    }
    this.connections.clear();
  }
}

// Global realtime manager instance
export const realtimeManager = new RealtimeManager();
