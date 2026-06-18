import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { certificationApi } from '@/api/certification';
import type { CertTier, CertExam, CertCredential, VerifyResult, PublicBadge } from '@/api/certification';

// ── Query Keys ─────────────────────────────────────────────────────────────────

export const certificationKeys = {
  all: ['certification'] as const,
  tiers: () => [...certificationKeys.all, 'tiers'] as const,
  exams: () => [...certificationKeys.all, 'exams'] as const,
  exam: (id: string) => [...certificationKeys.all, 'exam', id] as const,
  credentials: () => [...certificationKeys.all, 'credentials'] as const,
  credential: (id: string) => [...certificationKeys.all, 'credential', id] as const,
  verify: (username: string) => [...certificationKeys.all, 'verify', username] as const,
  badges: (username: string) => [...certificationKeys.all, 'badges', username] as const,
};

// ── Tier Hooks ─────────────────────────────────────────────────────────────────

export function useCertTiers() {
  return useQuery({
    queryKey: certificationKeys.tiers(),
    queryFn: () => certificationApi.listTiers(),
    staleTime: 1000 * 60 * 5, // 5 minutes — tiers rarely change
  });
}

// ── Exam Hooks ─────────────────────────────────────────────────────────────────

export function useStartExam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (tierSlug: string) => certificationApi.startExam(tierSlug),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: certificationKeys.exams() });
      const checkoutUrl = (data as { checkout_url?: string; exam?: { checkout_url?: string } }).checkout_url || (data as { exam?: { checkout_url?: string } }).exam?.checkout_url;
      if (checkoutUrl) {
        // Redirect to Stripe checkout
        window.location.href = checkoutUrl;
      } else {
        toast.success('Exam session started!');
      }
    },
    onError: (error: { message: string; response?: { status?: number } }) => {
      const status = error?.response?.status;
      const message = status === 409
        ? 'You already have an in-progress exam for this tier. Please complete or abandon it first.'
        : error?.message || 'Unknown error';
      toast.error(`Failed to start exam: ${message}`);
    },
  });
}

export function useExam(examId: string) {
  return useQuery({
    queryKey: certificationKeys.exam(examId),
    queryFn: () => certificationApi.getExam(examId),
    enabled: !!examId,
    refetchInterval: (query) => {
      // Poll while exam is in progress
      const status = query.state.data?.exam?.status;
      if (status === 'in_progress') return 30000; // 30 seconds
      return false;
    },
  });
}

export function useSubmitAnswer() {
  return useMutation({
    mutationFn: ({ examId, questionId, answer }: { examId: string; questionId: string; answer: string }) =>
      certificationApi.submitAnswer(examId, questionId, answer),
    onError: (error: Error) => {
      toast.error(`Failed to save answer: ${error.message}`);
    },
  });
}

export function useSubmitExam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ examId, answers }: { examId: string; answers?: Record<string, string> }) =>
      certificationApi.submitExam(examId, answers),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: certificationKeys.exam(variables.examId) });
      queryClient.invalidateQueries({ queryKey: certificationKeys.exams() });
      toast.success('Exam submitted successfully!');
    },
    onError: (error: Error) => {
      toast.error(`Failed to submit exam: ${error.message}`);
    },
  });
}

export function useAbandonExam() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (examId: string) => certificationApi.abandonExam(examId),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: certificationKeys.exam(variables) });
      queryClient.invalidateQueries({ queryKey: certificationKeys.exams() });
      toast.success('Exam abandoned.');
    },
    onError: (error: Error) => {
      toast.error(`Failed to abandon exam: ${error.message}`);
    },
  });
}

export function useMyExams(limit = 20, offset = 0) {
  return useQuery({
    queryKey: [...certificationKeys.exams(), { limit, offset }],
    queryFn: () => certificationApi.listMyExams(limit, offset),
    staleTime: 1000 * 30,
  });
}

// ── Credential Hooks ───────────────────────────────────────────────────────────

export function useMyCredentials() {
  return useQuery({
    queryKey: certificationKeys.credentials(),
    queryFn: () => certificationApi.listMyCredentials(),
    staleTime: 1000 * 60,
  });
}

export function useCredential(credentialId: string) {
  return useQuery({
    queryKey: certificationKeys.credential(credentialId),
    queryFn: () => certificationApi.getCredential(credentialId),
    enabled: !!credentialId,
  });
}

// ── Verification Hooks ─────────────────────────────────────────────────────────

export function useVerifyCredential(username: string) {
  return useQuery({
    queryKey: certificationKeys.verify(username),
    queryFn: () => certificationApi.verifyCredential(username),
    enabled: !!username,
    staleTime: 1000 * 60 * 2,
  });
}

export function usePublicBadges(username: string) {
  return useQuery({
    queryKey: certificationKeys.badges(username),
    queryFn: () => certificationApi.getPublicBadges(username),
    enabled: !!username,
    staleTime: 1000 * 60 * 2,
  });
}
