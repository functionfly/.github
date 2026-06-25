import apiClient from './client';

export interface LearningCourse {
  id: string;
  tenant_id: string;
  title: string;
  description?: string;
  category?: string;
  difficulty?: string;
  duration_min?: number;
  content_url?: string;
  thumbnail_url?: string;
  is_mandatory: boolean;
  is_active: boolean;
}

export interface EmployeeLearning {
  id: number;
  employee_id: string;
  course_id: string;
  course?: LearningCourse;
  status: string;
  progress_pct: number;
  started_at?: string;
  completed_at?: string;
  score?: number;
}

export const learningApi = {
  listCourses: () =>
    apiClient.get<{ courses: LearningCourse[] }>('/v1/learning/courses'),

  getMyProgress: () =>
    apiClient.get<{ progress: EmployeeLearning[] }>('/v1/learning/progress'),

  enroll: (courseId: string) =>
    apiClient.post(`/v1/learning/courses/${courseId}/enroll`),

  updateProgress: (id: number, data: { progress_pct?: number; status?: string }) =>
    apiClient.patch(`/v1/learning/progress/${id}`, data),
};
