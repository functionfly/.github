export const PROFILE_REPORT_REASONS = [
  { value: 'tos_violation', label: 'Violates Terms of Service' },
  { value: 'harassment', label: 'Harassment or abuse' },
  { value: 'spam', label: 'Spam or misleading content' },
  { value: 'impersonation', label: 'Impersonation' },
  { value: 'other', label: 'Other (describe below)' },
] as const;

export type ProfileReportReason = (typeof PROFILE_REPORT_REASONS)[number]['value'];
