import apiClient from './client';

export interface FeedbackRound {
  id: string;
  tenant_id: string;
  name: string;
  description?: string;
  review_period: string;
  round_type: string;
  status: string;
  start_date: string;
  end_date: string;
  questions: { question: string; type: string; required: boolean }[];
  created_at: string;
}

export interface FeedbackRoundAssignment {
  id: number;
  round_id: string;
  reviewer_id: string;
  reviewee_id: string;
  status: string;
}

export const feedbackRoundsApi = {
  list: () => apiClient.get<{ rounds: FeedbackRound[] }>('/v1/feedback-rounds'),
  create: (data: Partial<FeedbackRound>) => apiClient.post<{ round: FeedbackRound }>('/v1/feedback-rounds', data),
  start: (id: string) => apiClient.post(`/v1/feedback-rounds/${id}/start`),
  submitResponse: (assignmentId: number, responses: { question_index: number; response_text?: string; response_rating?: number }[]) => apiClient.post(`/v1/feedback-rounds/assignments/${assignmentId}/submit`, { responses }),
  getResults: (id: string) => apiClient.get<{ results: { reviewee_id: string; avg_ratings: Record<string, number>; comments: string[] }[] }>(`/v1/feedback-rounds/${id}/results`),
};
