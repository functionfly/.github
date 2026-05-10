import { apiClient } from './client';

// ── Types ──────────────────────────────────────────────────────────────────────

export interface CertTier {
  id: string;
  slug: string;
  name: string;
  description: string;
  icon: string;
  color: string;
  sort_order: number;
  price_cents: number;
  currency: string;
  pass_threshold: number;
  time_limit_minutes: number;
  question_count: number;
  practical_count: number;
  validity_months: number;
  is_coming_soon?: boolean;
}

export interface CertQuestionPublic {
  id: string;
  category: string;
  difficulty: string;
  question_text: string;
  question_format: 'multiple_choice' | 'true_false' | 'multi_select';
  options: { id: string; text: string }[];
  points: number;
}

export interface CertPracticalChallengePublic {
  id: string;
  name: string;
  description: string;
  category: string;
  difficulty: string;
  points: number;
  time_limit_minutes: number;
}

export interface CertExam {
  id: string;
  tier_id?: string;
  tier?: { slug: string; name: string };
  status: string;
  started_at: string;
  expires_at: string;
  time_limit_minutes?: number;
  question_count?: number;
  practical_count?: number;
  time_remaining_seconds?: number;
  answers?: Record<string, string>;
  questions?: CertQuestionPublic[];
  practical_challenges?: CertPracticalChallengePublic[];
  submitted_at?: string;
  graded_at?: string;
  knowledge_score?: number;
  practical_score?: number;
  total_score?: number;
  passed?: boolean;
  amount_cents?: number;
  created_at?: string;
}

export interface CertCredential {
  id: string;
  tier_id?: string;
  tier?: { slug: string; name: string };
  credential_number: string;
  status: string;
  issued_at: string;
  expires_at: string;
  verification_url?: string;
  verification_hash: string;
  oba_credential?: Record<string, unknown>;
  revoked_at?: string;
  revoked_reason?: string;
}

export interface VerifyResult {
  user: { id: string; username?: string; name?: string };
  credentials: {
    credential_number: string;
    tier?: { slug: string; name: string };
    status: string;
    issued_at: string;
    expires_at: string;
    verification_hash: string;
  }[];
}

export interface PublicBadge {
  tier_slug: string;
  tier_name: string;
  tier_color: string;
  tier_icon: string;
  credential_number: string;
  issued_at: string;
  expires_at: string;
}

// ── API Client ─────────────────────────────────────────────────────────────────

export const certificationApi = {
  // Public
  listTiers: async (): Promise<{ tiers: CertTier[] }> => {
    return apiClient.get('/v1/certification/tiers');
  },

  verifyCredential: async (username: string): Promise<VerifyResult> => {
    return apiClient.get(`/v1/certification/verify/${username}`);
  },

  verifyByNumber: async (number: string): Promise<VerifyResult> => {
    return apiClient.get(`/v1/certification/verify/number/${number}`);
  },

  getPublicBadges: async (username: string): Promise<{ username: string; badges: PublicBadge[]; count: number }> => {
    return apiClient.get(`/v1/certification/credentials/${username}/badges`);
  },

  // Authenticated — Exams
  startExam: async (tierSlug: string): Promise<{ exam: CertExam; checkout_url?: string }> => {
    return apiClient.post(`/v1/certification/tiers/${tierSlug}/start`, {});
  },

  getExam: async (examId: string): Promise<{ exam: CertExam }> => {
    return apiClient.get(`/v1/certification/exams/${examId}`);
  },

  submitAnswer: async (examId: string, questionId: string, answer: string): Promise<{ saved: boolean; answers_submitted: number }> => {
    return apiClient.put(`/v1/certification/exams/${examId}/answer`, { question_id: questionId, answer });
  },

  submitExam: async (examId: string, answers?: Record<string, string>): Promise<{ status: string; message: string }> => {
    return apiClient.post(`/v1/certification/exams/${examId}/submit`, answers ? { answers } : {});
  },

  abandonExam: async (examId: string): Promise<{ status: string; message: string }> => {
    return apiClient.post(`/v1/certification/exams/${examId}/abandon`, {});
  },

  listMyExams: async (limit = 20, offset = 0): Promise<{ exams: CertExam[]; total: number; limit: number; offset: number }> => {
    return apiClient.get(`/v1/certification/exams?limit=${limit}&offset=${offset}`);
  },

  // Authenticated — Credentials
  listMyCredentials: async (): Promise<{ credentials: CertCredential[] }> => {
    return apiClient.get('/v1/certification/credentials');
  },

  getCredential: async (credentialId: string): Promise<CertCredential> => {
    return apiClient.get(`/v1/certification/credentials/${credentialId}`);
  },
};
