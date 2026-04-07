import { apiClient } from './client';

// ==================== Types ====================

export interface PaymentMethod {
  brand: string;
  last4: string;
  exp_month: number;
  exp_year: number;
}

export interface CreatePortalSessionResponse {
  url: string;
}

export interface CreateCheckoutSessionResponse {
  session_id: string;
  url: string;
}

export interface Subscription {
  id: string;
  tenant_id: string;
  plan: string;
  status: string;
  stripe_subscription_id: string | null;
  stripe_price_id: string | null;
  current_period_start: string | null;
  current_period_end: string | null;
  cancel_at_period_end: boolean;
  canceled_at: string | null;
  trial_end: string | null;
  is_trialing: boolean;
  trial_days_remaining: number;
  created_at: string;
  updated_at: string;
  payment_method?: PaymentMethod | null;
}

export interface Invoice {
  id: string;
  tenant_id: string;
  stripe_invoice_id: string | null;
  /** Amount in smallest currency unit (cents for USD); matches server and formatCurrency in settings. */
  amount: number;
  currency: string;
  status: string;
  invoice_date: string | null;
  due_date: string | null;
  invoice_pdf: string | null;
  hosted_invoice_url: string | null;
  created_at: string;
}

export interface InvoicesResponse {
  invoices: Invoice[];
  limit: number;
  offset: number;
  total: number;
}

export interface CreateCheckoutRequest {
  price_id: string;
  success_url?: string;
  cancel_url?: string;
}

// ==================== Helper Functions ====================

/** Extract error message from API error (e.g. axios 4xx/5xx). */
export function getBillingPortalErrorMessage(e: unknown): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const res = (e as { response?: { data?: { error?: string }; status?: number } }).response;
    if (res?.data?.error) return res.data.error;
    if (res?.status === 503)
      return 'Billing is not configured. Set STRIPE_SECRET_KEY on the server.';
  }
  return 'Could not open billing portal.';
}

/** Extract error message from checkout API error. */
export function getCheckoutErrorMessage(e: unknown): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const res = (e as { response?: { data?: { error?: string }; status?: number } }).response;
    if (res?.data?.error) return res.data.error;
    if (res?.status === 503)
      return 'Billing is not configured. Set STRIPE_SECRET_KEY on the server.';
    if (res?.status === 400) return 'Invalid price ID. Please try again.';
  }
  return 'Could not create checkout session.';
}

/** Extract error message from subscription API error. */
export function getSubscriptionErrorMessage(e: unknown): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const res = (e as { response?: { data?: { error?: string }; status?: number } }).response;
    if (res?.data?.error) return res.data.error;
  }
  return 'Could not load subscription details.';
}

/** Extract error message from invoices API error. */
export function getInvoicesErrorMessage(e: unknown): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const res = (e as { response?: { data?: { error?: string }; status?: number } }).response;
    if (res?.data?.error) return res.data.error;
  }
  return 'Could not load invoices.';
}

// ==================== API Functions ====================

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
    '/v1/billing/portal-session',
    body ?? {}
  );
  return response;
}

/**
 * Create a Stripe Checkout session for subscription checkout.
 * Returns the URL to redirect the user to complete the checkout on Stripe.
 * @param priceId - The Stripe price ID for the subscription plan
 * @param successUrl - URL to redirect after successful checkout (optional)
 * @param cancelUrl - URL to redirect after cancelled checkout (optional)
 */
export async function createCheckoutSession(
  priceId: string,
  successUrl?: string,
  cancelUrl?: string
): Promise<CreateCheckoutSessionResponse> {
  const body: CreateCheckoutRequest = {
    price_id: priceId,
  };
  if (successUrl) body.success_url = successUrl;
  if (cancelUrl) body.cancel_url = cancelUrl;

  const response = await apiClient.post<CreateCheckoutSessionResponse>(
    '/v1/billing/checkout',
    body
  );
  return response;
}

/**
 * Get the current user's subscription details.
 * Returns subscription information including plan, status, and billing period.
 */
export async function getSubscription(): Promise<Subscription> {
  const response = await apiClient.get<Subscription>('/v1/billing/subscription');
  return response;
}

/**
 * Get the current user's invoices.
 * @param limit - Number of invoices to return (default: 10)
 * @param offset - Number of invoices to skip for pagination (default: 0)
 */
export async function listInvoices(
  limit: number = 10,
  offset: number = 0
): Promise<InvoicesResponse> {
  const params = new URLSearchParams({
    limit: limit.toString(),
    offset: offset.toString(),
  });
  const response = await apiClient.get<InvoicesResponse>(`/v1/billing/invoices?${params}`);
  return response;
}

/**
 * Get the current user's usage details.
 * Returns usage information for the specified period.
 */
export interface UsageDataPoint {
  event_type: string;
  quantity: number;
  unit_price_cents: number;
  total_cost_cents: number;
  timestamp: string;
}

export interface UsageSummary {
  start: string;
  end: string;
  total_events: number;
  total_cost_usd: number;
  events: UsageDataPoint[];
}

export async function getUsage(
  startDate?: string,
  endDate?: string
): Promise<UsageSummary> {
  const params = new URLSearchParams();
  if (startDate) params.set('start', startDate);
  if (endDate) params.set('end', endDate);

  const queryString = params.toString();
  const url = queryString ? `/v1/billing/usage?${queryString}` : '/v1/billing/usage';

  const response = await apiClient.get<UsageSummary>(url);
  return response;
}

/**
 * Cancel the current user's subscription.
 * @param immediately - If true, cancels immediately. If false, cancels at period end.
 */
export async function cancelSubscription(
  immediately: boolean = false
): Promise<{ message: string }> {
  const response = await apiClient.post<{ message: string }>('/v1/billing/subscription/cancel', {
    immediately,
  });
  return response;
}

// ==================== Wallet & Platform Fees ====================

export interface WalletInfo {
  user_id: string;
  balance_usd: number;
  lifetime_earnings_usd: number;
  lifetime_fees_usd: number;
}

export interface PlatformFee {
  id: string;
  user_id: string;
  fee_type: string;
  amount_usd: number;
  description: string;
  created_at: string;
}

export interface PlatformFeesResponse {
  fees: PlatformFee[];
  limit: number;
  offset: number;
  total: number;
}

export interface TopUpResponse {
  checkout_url: string;
  session_id: string;
}

/**
 * Get the current user's wallet information.
 * Returns balance and lifetime earnings/fees.
 */
export async function getWalletInfo(): Promise<WalletInfo> {
  const response = await apiClient.get<WalletInfo>('/v1/billing/wallet');
  return response;
}

/**
 * Get the current user's platform fees history.
 * @param limit - Number of fees to return (default: 20)
 * @param offset - Number of fees to skip for pagination (default: 0)
 */
export async function listPlatformFees(
  limit: number = 20,
  offset: number = 0
): Promise<PlatformFeesResponse> {
  const params = new URLSearchParams({
    limit: limit.toString(),
    offset: offset.toString(),
  });
  const response = await apiClient.get<PlatformFeesResponse>(`/v1/billing/fees?${params}`);
  return response;
}

/**
 * Create a Stripe checkout session to add funds to wallet.
 * Returns the URL to redirect the user to complete the checkout on Stripe.
 * @param amountUsd - Amount in USD to add (minimum $1.00)
 * @param successUrl - URL to redirect after successful checkout (optional)
 * @param cancelUrl - URL to redirect after cancelled checkout (optional)
 */
export async function topUpWallet(
  amountUsd: number,
  successUrl?: string,
  cancelUrl?: string
): Promise<TopUpResponse> {
  const body: {
    amount_usd: number;
    success_url?: string;
    cancel_url?: string;
  } = {
    amount_usd: amountUsd,
  };
  if (successUrl) body.success_url = successUrl;
  if (cancelUrl) body.cancel_url = cancelUrl;

  const response = await apiClient.post<TopUpResponse>('/v1/billing/wallet/top-up', body);
  return response;
}

/** Extract error message from wallet API error. */
export function getWalletErrorMessage(e: unknown): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const res = (e as { response?: { data?: { error?: string }; status?: number } }).response;
    if (res?.data?.error) return res.data.error;
    if (res?.status === 503)
      return 'Billing is not configured. Set STRIPE_SECRET_KEY on the server.';
  }
  return 'Could not load wallet information.';
}

/** Extract error message from top-up API error. */
export function getTopUpErrorMessage(e: unknown): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const res = (e as { response?: { data?: { error?: string }; status?: number } }).response;
    if (res?.data?.error) return res.data.error;
    if (res?.status === 400) return 'Invalid amount. Minimum top-up is $1.00.';
    if (res?.status === 503)
      return 'Billing is not configured. Set STRIPE_SECRET_KEY on the server.';
  }
  return 'Could not create top-up session.';
}

// ==================== State Fabric add-ons ====================

export interface StateFabricAddOnDTO {
  id: string;
  name: string;
  price: string;
  period: string;
  description: string;
  stripe_price_id?: string;
}

export interface StateFabricAddOnCatalogResponse {
  add_ons: StateFabricAddOnDTO[];
}

/**
 * Public catalog of State Fabric add-ons (pricing page). No auth required.
 * GET /v1/billing/state-fabric/add-ons
 */
export async function listStateFabricAddOnCatalog(): Promise<StateFabricAddOnCatalogResponse> {
  return apiClient.get<StateFabricAddOnCatalogResponse>('/v1/billing/state-fabric/add-ons');
}

/**
 * Active add-on IDs for the current tenant (from state_fabric_addon_entitlements).
 * GET /v1/billing/state-fabric/add-ons/entitlements
 */
export async function getStateFabricAddOnEntitlements(): Promise<{ addon_ids: string[] }> {
  return apiClient.get<{ addon_ids: string[] }>('/v1/billing/state-fabric/add-ons/entitlements');
}

/**
 * Start checkout for a State Fabric add-on subscription.
 * POST /v1/billing/state-fabric/add-ons/checkout
 */
export async function createStateFabricAddOnCheckout(
  addonId: string,
  successUrl?: string,
  cancelUrl?: string
): Promise<CreateCheckoutSessionResponse> {
  return apiClient.post<CreateCheckoutSessionResponse>(
    '/v1/billing/state-fabric/add-ons/checkout',
    {
      addon_id: addonId,
      success_url: successUrl,
      cancel_url: cancelUrl,
    }
  );
}
