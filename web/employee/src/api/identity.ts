import apiClient from './client';
import type { Employee } from './employees';

export interface Achievement {
  id: string;
  slug: string;
  name: string;
  description?: string;
  icon?: string;
  category: string;
  points: number;
  tier: number;
  earned?: boolean;
  earned_at?: string;
  progress?: number;
  threshold?: number;
}

export interface TimelineEvent {
  id: string;
  event_type: string;
  title: string;
  description?: string;
  event_date: string;
  metadata?: Record<string, unknown>;
}

export interface ReputationTrend {
  category: string;
  history: { score: number; recorded_at: string }[];
}

export interface IdentityCard {
  employee: Employee;
  identity_signature: string;
  clearance_level_num: number;
  reputation_total: number;
  trust_score: number;
  achievements: Achievement[];
  skills: { name: string; level: number }[];
  recent_timeline: TimelineEvent[];
}

export interface AchievementProgress {
  achievement_id: string;
  current_value: number;
  awarded: boolean;
}

export const identityApi = {
  getCard: (employeeId: string) =>
    apiClient.get<{ card: IdentityCard }>(`/v1/employees/${employeeId}/identity-card`),

  getAchievements: () =>
    apiClient.get<{ definitions: Achievement[]; progress: AchievementProgress[] }>('/v1/achievements'),

  getTimeline: (employeeId: string) =>
    apiClient.get<{ events: TimelineEvent[] }>(`/v1/employees/${employeeId}/timeline`),

  getReputationTrends: (employeeId: string) =>
    apiClient.get<{ trends: ReputationTrend[] }>(`/v1/reputation/${employeeId}/trends`),

  getPassport: (employeeId: string) =>
    apiClient.get<{ passport: IdentityCard }>(`/v1/employees/${employeeId}/passport`),
};
