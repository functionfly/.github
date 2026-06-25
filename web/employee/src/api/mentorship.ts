import apiClient from './client';

export interface MentorshipMatch {
  id: string;
  tenant_id: string;
  mentor_id: string;
  mentee_id: string;
  focus_area?: string;
  status: string;
  started_at: string;
  meeting_frequency?: string;
  notes?: string;
}

export const mentorshipApi = {
  list: () => apiClient.get<{ matches: MentorshipMatch[] }>('/v1/mentorship'),
  request: (mentorId: string, focusArea?: string) => apiClient.post('/v1/mentorship', { mentor_id: mentorId, focus_area: focusArea }),
  update: (id: string, data: Partial<MentorshipMatch>) => apiClient.patch(`/v1/mentorship/${id}`, data),
};
