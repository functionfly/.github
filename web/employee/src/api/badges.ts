import apiClient from './client';

export interface DigitalBadge {
  id: string;
  slug: string;
  name: string;
  description?: string;
  icon_url?: string;
  category: string;
  points: number;
}

export interface EmployeeBadge {
  id: number;
  badge_id: string;
  awarded_at: string;
  badge?: DigitalBadge;
}

export const badgesApi = {
  list: () => apiClient.get<{ badges: DigitalBadge[] }>('/v1/badges'),
  getMyBadges: () => apiClient.get<{ badges: EmployeeBadge[] }>('/v1/badges/mine'),
  award: (employeeId: string, badgeId: string) => apiClient.post('/v1/badges/award', { employee_id: employeeId, badge_id: badgeId }),
};
