// Supabase TypeScript types generated from your database schema

export type Json =
  | string
  | number
  | boolean
  | null
  | { [key: string]: Json | undefined }
  | Json[]

export interface Database {
  public: {
    Tables: {
      tenants: {
        Row: {
          id: string
          name: string
          plan: string | null
          status: string
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          name: string
          plan?: string | null
          status?: string
          created_at?: string
          updated_at?: string
        }
        Update: {
          id?: string
          name?: string
          plan?: string | null
          status?: string
          created_at?: string
          updated_at?: string
        }
        Relationships: []
      }
      users: {
        Row: {
          id: string
          tenant_id: string
          email: string
          password_hash: string | null
          role: string | null
          email_verified: boolean
          verification_token: string | null
          verification_expires_at: string | null
          provider: string | null
          provider_id: string | null
          provider_data: Json | null
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          tenant_id: string
          email: string
          password_hash?: string | null
          role?: string | null
          email_verified?: boolean
          verification_token?: string | null
          verification_expires_at?: string | null
          provider?: string | null
          provider_id?: string | null
          provider_data?: Json | null
          created_at?: string
          updated_at?: string
        }
        Update: {
          id?: string
          tenant_id?: string
          email?: string
          password_hash?: string | null
          role?: string | null
          email_verified?: boolean
          verification_token?: string | null
          verification_expires_at?: string | null
          provider?: string | null
          provider_id?: string | null
          provider_data?: Json | null
          created_at?: string
          updated_at?: string
        }
        Relationships: [
          {
            foreignKeyName: "users_tenant_id_fkey"
            columns: ["tenant_id"]
            isOneToOne: false
            referencedRelation: "tenants"
            referencedColumns: ["id"]
          }
        ]
      }
      user_profiles: {
        Row: {
          id: string
          user_id: string
          first_name: string | null
          last_name: string | null
          avatar_url: string | null
          bio: string | null
          timezone: string
          preferences: Json
          last_active_at: string | null
          created_at: string
          updated_at: string
        }
        Insert: {
          id?: string
          user_id: string
          first_name?: string | null
          last_name?: string | null
          avatar_url?: string | null
          bio?: string | null
          timezone?: string
          preferences?: Json
          last_active_at?: string | null
          created_at?: string
          updated_at?: string
        }
        Update: {
          id?: string
          user_id?: string
          first_name?: string | null
          last_name?: string | null
          avatar_url?: string | null
          bio?: string | null
          timezone?: string
          preferences?: Json
          last_active_at?: string | null
          created_at?: string
          updated_at?: string
        }
        Relationships: [
          {
            foreignKeyName: "user_profiles_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: true
            referencedRelation: "users"
            referencedColumns: ["id"]
          }
        ]
      }
      user_sessions: {
        Row: {
          id: string
          user_id: string
          session_token: string
          ip_address: string | null
          user_agent: string | null
          expires_at: string
          created_at: string
          last_activity_at: string
        }
        Insert: {
          id?: string
          user_id: string
          session_token: string
          ip_address?: string | null
          user_agent?: string | null
          expires_at: string
          created_at?: string
          last_activity_at?: string
        }
        Update: {
          id?: string
          user_id?: string
          session_token?: string
          ip_address?: string | null
          user_agent?: string | null
          expires_at?: string
          created_at?: string
          last_activity_at?: string
        }
        Relationships: [
          {
            foreignKeyName: "user_sessions_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: false
            referencedRelation: "users"
            referencedColumns: ["id"]
          }
        ]
      }
      user_notifications: {
        Row: {
          id: string
          user_id: string
          type: string
          title: string
          message: string | null
          data: Json
          read_at: string | null
          expires_at: string | null
          created_at: string
        }
        Insert: {
          id?: string
          user_id: string
          type: string
          title: string
          message?: string | null
          data?: Json
          read_at?: string | null
          expires_at?: string | null
          created_at?: string
        }
        Update: {
          id?: string
          user_id?: string
          type?: string
          title?: string
          message?: string | null
          data?: Json
          read_at?: string | null
          expires_at?: string | null
          created_at?: string
        }
        Relationships: [
          {
            foreignKeyName: "user_notifications_user_id_fkey"
            columns: ["user_id"]
            isOneToOne: false
            referencedRelation: "users"
            referencedColumns: ["id"]
          }
        ]
      }
      audit_events: {
        Row: {
          id: string
          actor_user_id: string | null
          actor_email: string | null
          tenant_id: string | null
          action: string
          resource_type: string
          resource_id: string | null
          request_id: string | null
          before_state: Json | null
          after_state: Json | null
          ip_address: string | null
          user_agent: string | null
          timestamp: string
          success: boolean
        }
        Insert: {
          id?: string
          actor_user_id?: string | null
          actor_email?: string | null
          tenant_id?: string | null
          action: string
          resource_type: string
          resource_id?: string | null
          request_id?: string | null
          before_state?: Json | null
          after_state?: Json | null
          ip_address?: string | null
          user_agent?: string | null
          timestamp?: string
          success?: boolean
        }
        Update: {
          id?: string
          actor_user_id?: string | null
          actor_email?: string | null
          tenant_id?: string | null
          action?: string
          resource_type?: string
          resource_id?: string | null
          request_id?: string | null
          before_state?: Json | null
          after_state?: Json | null
          ip_address?: string | null
          user_agent?: string | null
          timestamp?: string
          success?: boolean
        }
        Relationships: [
          {
            foreignKeyName: "audit_events_actor_user_id_fkey"
            columns: ["actor_user_id"]
            isOneToOne: false
            referencedRelation: "users"
            referencedColumns: ["id"]
          },
          {
            foreignKeyName: "audit_events_tenant_id_fkey"
            columns: ["tenant_id"]
            isOneToOne: false
            referencedRelation: "tenants"
            referencedColumns: ["id"]
          }
        ]
      }
    }
    Views: {
      [_ in never]: never
    }
    Functions: {
      get_online_users: {
        Args: {
          tenant_filter: string | null
        }
        Returns: {
          user_id: string
          email: string
          last_active_at: string | null
          tenant_id: string
        }[]
      }
      get_user_activity_summary: {
        Args: {
          user_id: string
          days_back?: number
        }
        Returns: Json
      }
      cleanup_expired_sessions: {
        Args: Record<PropertyKey, never>
        Returns: number
      }
      cleanup_expired_notifications: {
        Args: Record<PropertyKey, never>
        Returns: number
      }
    }
    Enums: {
      [_ in never]: never
    }
    CompositeTypes: {
      [_ in never]: never
    }
  }
}

// Helper types for the application
export type User = Database['public']['Tables']['users']['Row'] & {
  profile?: Database['public']['Tables']['user_profiles']['Row']
}

export type UserProfile = Database['public']['Tables']['user_profiles']['Row']
export type UserSession = Database['public']['Tables']['user_sessions']['Row']
export type UserNotification = Database['public']['Tables']['user_notifications']['Row']
export type AuditEvent = Database['public']['Tables']['audit_events']['Row']
export type Tenant = Database['public']['Tables']['tenants']['Row']

// Real-time event types
export type RealtimeUserEvent = {
  type: 'INSERT' | 'UPDATE' | 'DELETE'
  table: string
  user_id?: string
  tenant_id?: string
  data: User
}

export type RealtimeNotificationEvent = {
  type: 'INSERT' | 'UPDATE' | 'DELETE'
  table: string
  user_id: string
  data: UserNotification
}

export type RealtimeProfileEvent = {
  type: 'INSERT' | 'UPDATE' | 'DELETE'
  table: string
  user_id: string
  data: UserProfile
}