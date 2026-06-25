import apiClient from './client';

export interface PerformanceGoal {
  id: string;
  employee_id: string;
  title: string;
  description?: string;
  category: string;
  status: string;
  priority: string;
  target_date?: string;
  progress_pct: number;
  created_at: string;
}

export interface PerformanceReview {
  id: string;
  employee_id: string;
  reviewer_id: string;
  review_period: string;
  review_type: string;
  status: string;
  strengths?: string;
  areas_for_improvement?: string;
  overall_rating?: number;
  comments?: string;
  created_at: string;
}

export interface PeerFeedback {
  id: string;
  from_employee_id: string;
  to_employee_id: string;
  feedback_text: string;
  rating?: number;
  is_anonymous: boolean;
  created_at: string;
}

export const performanceApi = {
  listGoals: () => apiClient.get<{ goals: PerformanceGoal[] }>('/v1/performance/goals'),
  createGoal: (data: Partial<PerformanceGoal>) => apiClient.post<{ goal: PerformanceGoal }>('/v1/performance/goals', data),
  updateGoal: (id: string, data: Partial<PerformanceGoal>) => apiClient.patch(`/v1/performance/goals/${id}`, data),
  listReviews: () => apiClient.get<{ reviews: PerformanceReview[] }>('/v1/performance/reviews'),
  createReview: (data: Partial<PerformanceReview>) => apiClient.post<{ review: PerformanceReview }>('/v1/performance/reviews', data),
  submitReview: (id: string) => apiClient.post(`/v1/performance/reviews/${id}/submit`),
  listFeedback: (employeeId: string) => apiClient.get<{ feedback: PeerFeedback[] }>(`/v1/performance/feedback/${employeeId}`),
  giveFeedback: (data: Partial<PeerFeedback>) => apiClient.post('/v1/performance/feedback', data),
};
