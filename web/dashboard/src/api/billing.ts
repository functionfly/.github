import { apiClient } from "./client";

export interface CreatePortalSessionResponse {
  url: string;
}

/** Extract error message from API error (e.g. axios 4xx/5xx). */
export function getBillingPortalErrorMessage(e: unknown): string {
  if (e && typeof e === "object" && "response" in e) {
    const res = (e as { response?: { data?: { error?: string }; status?: number } }).response;
    if (res?.data?.error) return res.data.error;
    if (res?.status === 503) return "Billing is not configured. Set STRIPE_SECRET_KEY on the server.";
  }
  return "Could not open billing portal.";
}

/**
 * Create a Stripe Customer Billing Portal session.
 * Returns the URL to redirect the user to manage subscription, payment methods, and invoices.
 * Fails with 503 if STRIPE_SECRET_KEY is not set on the API server.
 */
export async function createBillingPortalSession(
  returnUrl?: string
): Promise<CreatePortalSessionResponse> {
  const body = returnUrl ? { return_url: returnUrl } : undefined;
  const response = await apiClient.post<CreatePortalSessionResponse>(
    "/v1/billing/portal-session",
    body ?? {}
  );
  return response;
}
