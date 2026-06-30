import { apiClient } from './client';

// ==================== Types ====================

export interface PaymentMethod {
  brand: string;
  last4: string;
  exp_month: number;
  exp_year: number;
  is_default?: boolean;
  stripe_payment_method_id?: string;
}

export interface PaymentMethodsResponse {
  payment_methods: PaymentMethod[];
}

export interface SetupIntentResponse {
  client_secret: string;
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

  // Fetch CSRF token for protected billing endpoint
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  const response = await apiClient.post<CreatePortalSessionResponse>(
    '/v1/billing/portal-session',
    body ?? {},
    { headers }
  );
  return response;
}

/**
 * Validates that the provided ID is a valid Stripe price ID.
 * Prevents common mistakes like using product IDs (prod_*) instead of price IDs.
 */
function validateStripePriceId(priceId: string): void {
  if (!priceId) {
    throw new Error('Price ID is required');
  }

  // Check for common mistakes
  if (priceId.startsWith('prod_')) {
    throw new Error(
      `Invalid price ID: received product ID (${priceId}) instead of price ID. ` +
        'Product IDs cannot be used for checkout - use the associated price ID (price_*) from Stripe Dashboard.'
    );
  }
  if (priceId.startsWith('sub_')) {
    throw new Error(`Invalid price ID: received subscription ID (${priceId}) instead of price ID`);
  }
  if (priceId.startsWith('plan_')) {
    throw new Error(`Invalid price ID: received plan ID (${priceId}) instead of price ID`);
  }

  if (!priceId.startsWith('price_')) {
    throw new Error(`Invalid price ID: must start with 'price_', got: ${priceId}`);
  }
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
  // Validate price ID before sending to backend
  validateStripePriceId(priceId);

  const body: CreateCheckoutRequest = {
    price_id: priceId,
  };
  if (successUrl) body.success_url = successUrl;
  if (cancelUrl) body.cancel_url = cancelUrl;

  // Fetch CSRF token for protected billing endpoint
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  const response = await apiClient.post<CreateCheckoutSessionResponse>(
    '/v1/billing/checkout',
    body,
    { headers }
  );
  return response;
}

/**
 * Get the current user's payment methods.
 */
export async function listPaymentMethods(): Promise<PaymentMethodsResponse> {
  return apiClient.get<PaymentMethodsResponse>('/v1/billing/payment-methods');
}

/**
 * Create a Stripe SetupIntent for collecting payment methods client-side.
 */
export async function createSetupIntent(): Promise<SetupIntentResponse> {
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  return apiClient.post<SetupIntentResponse>(
    '/v1/billing/payment-methods/setup-intent',
    {},
    { headers }
  );
}

/**
 * Set a payment method as the default for the customer.
 */
export async function setDefaultPaymentMethod(
  paymentMethodId: string
): Promise<{ message: string }> {
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  return apiClient.post<{ message: string }>(
    '/v1/billing/payment-methods/default',
    { payment_method_id: paymentMethodId },
    { headers }
  );
}

/**
 * Remove a payment method from the customer's account.
 */
export async function removePaymentMethod(paymentMethodId: string): Promise<{ message: string }> {
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  return apiClient.delete<{ message: string }>(`/v1/billing/payment-methods/${paymentMethodId}`, {
    headers,
  });
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

export async function getUsage(startDate?: string, endDate?: string): Promise<UsageSummary> {
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
  // Fetch CSRF token for protected billing endpoint
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  const response = await apiClient.post<{ message: string }>(
    '/v1/billing/subscription/cancel',
    { immediately },
    { headers }
  );
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

export interface WalletTransaction {
  id: string;
  type: 'credit' | 'debit' | 'top_up' | 'refund' | 'payout';
  amount: number;
  description: string;
  timestamp: string;
  status: 'completed' | 'pending' | 'failed';
}

export interface WalletTransactionsResponse {
  transactions: WalletTransaction[];
  total: number;
}

/**
 * Get wallet transaction history.
 */
export async function getWalletTransactions(limit = 50): Promise<WalletTransactionsResponse> {
  try {
    const params = new URLSearchParams({ limit: String(limit) });
    const response = await apiClient.get<WalletTransactionsResponse>(
      `/v1/billing/wallet/transactions?${params}`
    );
    return response;
  } catch {
    return { transactions: [], total: 0 };
  }
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

  // Fetch CSRF token for protected billing endpoint
  const csrfToken = await apiClient.fetchCSRFTokenWithRetry();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  const response = await apiClient.post<TopUpResponse>('/v1/billing/wallet/top-up', body, {
    headers,
  });
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
  // Fetch CSRF token for protected billing endpoint
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  return apiClient.post<CreateCheckoutSessionResponse>(
    '/v1/billing/state-fabric/add-ons/checkout',
    {
      addon_id: addonId,
      success_url: successUrl,
      cancel_url: cancelUrl,
    },
    { headers }
  );
}

// ==================== Backend-in-a-Box Pricing Bundles ====================

export interface Bundle {
  id: string;
  slug: string;
  name: string;
  display_name: string;
  description: string;
  short_description: string;
  price_cents: number;
  price_usd: string;
  billing_interval: string;
  icon: string;
  color: string;
  features_included: string[];
  feature_limits: Record<string, number>;
  provisioning_steps: string[];
  is_popular: boolean;
}

export interface FounderModeRegistration {
  id: string;
  bundle_id: string;
  bundle_slug: string;
  mode_type: string;
  status: string;
  started_at: string;
  ends_at?: string;
  free_days: number;
  mrr_threshold_cents: number;
  days_remaining: number;
}

export interface DeferredBillingStatus {
  bundle_id: string;
  status: string;
  trigger_thresholds: Record<string, unknown>;
  current_progress: Record<string, unknown>;
  progress_percent: number;
  estimated_days_left?: number;
}

export interface FounderModeRequest {
  mode_type: 'time_based' | 'revenue_based' | 'hybrid';
  free_days: number;
  mrr_threshold: number;
  success_url: string;
  cancel_url: string;
}

/**
 * Get all active Backend-in-a-Box pricing bundles.
 * GET /v1/billing/bundles
 */
export async function getBundles(): Promise<{ bundles: Bundle[] }> {
  return apiClient.get<{ bundles: Bundle[] }>('/v1/billing/bundles');
}

export interface BundleStats {
  active_founders: number;
  recent_deployments: number;
}

/**
 * Get public stats about bundles (founder count, deployments).
 * GET /v1/billing/bundles/stats
 */
export async function getBundleStats(): Promise<BundleStats> {
  return apiClient.get<BundleStats>('/v1/billing/bundles/stats');
}

/**
 * Get a specific bundle by slug.
 * GET /v1/billing/bundles/:slug
 */
export async function getBundle(slug: string): Promise<Bundle> {
  return apiClient.get<Bundle>(`/v1/billing/bundles/${slug}`);
}

/**
 * Create a checkout session for immediate bundle subscription.
 * POST /v1/billing/bundles/:slug/checkout
 */
export async function createBundleCheckout(
  slug: string,
  successUrl: string,
  cancelUrl: string
): Promise<CreateCheckoutSessionResponse> {
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  return apiClient.post<CreateCheckoutSessionResponse>(
    `/v1/billing/bundles/${slug}/checkout`,
    { success_url: successUrl, cancel_url: cancelUrl },
    { headers }
  );
}

/**
 * Register for founder mode ("Build Now, Pay Later" / "Free until $1K MRR").
 * POST /v1/billing/bundles/:slug/founder
 */
export async function registerFounderMode(
  slug: string,
  data: FounderModeRequest
): Promise<FounderModeRegistration> {
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  return apiClient.post<FounderModeRegistration>(`/v1/billing/bundles/${slug}/founder`, data, {
    headers,
  });
}

/**
 * Get founder mode status for current tenant.
 * GET /v1/billing/founder-mode
 */
export async function getFounderModeStatus(): Promise<{
  founder_modes: FounderModeRegistration[];
}> {
  return apiClient.get<{ founder_modes: FounderModeRegistration[] }>('/v1/billing/founder-mode');
}

/**
 * Get deferred billing progress toward trigger thresholds.
 * GET /v1/billing/deferred-status
 */
export async function getDeferredBillingStatus(): Promise<DeferredBillingStatus> {
  return apiClient.get<DeferredBillingStatus>('/v1/billing/deferred-status');
}

/**
 * Manually convert from founder mode to paid subscription.
 * POST /v1/billing/convert-to-paid
 */
export async function convertToPaid(bundleId: string): Promise<CreateCheckoutSessionResponse> {
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  return apiClient.post<CreateCheckoutSessionResponse>(
    '/v1/billing/convert-to-paid',
    { bundle_id: bundleId },
    { headers }
  );
}

// ==================== Backend-in-a-Box Bundle Deployment ====================

export interface DeploymentStep {
  id: string;
  name: string;
  description: string;
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
  error?: string;
}

export interface DeploymentResponse {
  deployment_id: string;
  status: string;
  message?: string;
  app_id?: string;
  backend_id?: string;
  steps?: DeploymentStep[];
}

export interface DeploymentStatusResponse {
  deployment_id: string;
  status: string;
  message?: string;
  progress: number;
  current_step?: string;
  steps: DeploymentStep[];
  error?: string;
  app_id?: string;
  backend_id?: string;
}

/**
 * Initiate a bundle deployment.
 * POST /v1/billing/bundles/{slug}/deploy
 */
export async function deployBundle(slug: string, region?: string): Promise<DeploymentResponse> {
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  return apiClient.post<DeploymentResponse>(
    `/v1/billing/bundles/${slug}/deploy`,
    { bundle_slug: slug, region },
    { headers }
  );
}

/**
 * Get deployment status by deployment ID.
 * GET /v1/billing/deployments/{deploymentId}/status
 */
export async function getDeploymentStatus(deploymentId: string): Promise<DeploymentStatusResponse> {
  return apiClient.get<DeploymentStatusResponse>(`/v1/billing/deployments/${deploymentId}/status`);
}

// ==================== Revenue Recognition (ASC 606/IFRS 15) ====================

export interface DeferredRevenueResponse {
  tenant_id: string;
  period: string;
  opening_balance_cents: number;
  new_deferred_cents: number;
  recognized_cents: number;
  closing_balance_cents: number;
}

export interface RecognizedRevenueResponse {
  tenant_id: string;
  period: string;
  subscription_revenue_cents: number;
  usage_revenue_cents: number;
  one_time_revenue_cents: number;
  total_cents: number;
}

export interface RevenueReportResponse {
  report_id: string;
  period: string;
  total_revenue_cents: number;
  total_deferred_cents: number;
  total_recognized_cents: number;
  opening_deferred_cents: number;
  new_deferred_cents: number;
  recognized_from_deferred_cents: number;
  closing_deferred_cents: number;
  over_time_revenue_cents: number;
  point_in_time_revenue_cents: number;
}

export interface UnbilledRevenueResponse {
  tenant_id: string;
  unbilled_revenue_cents: number;
}

export interface AllocationRequest {
  invoice_id: string;
  invoice_amount_cents: number;
  currency: string;
  line_items: AllocationLineItem[];
}

export interface AllocationLineItem {
  description: string;
  amount_cents: number;
  revenue_type: 'subscription' | 'usage' | 'one_time';
  ssp_cents: number;
  recognition_method: 'over_time' | 'point_in_time';
  start_date: string;
  end_date?: string;
  delivery_pattern: 'linear' | 'milestone' | 'usage_based';
}

export interface AllocationResponse {
  performance_obligation_ids: string[];
  schedule_count: number;
}

/**
 * Get deferred revenue summary for a period.
 * GET /v1/billing/revenue/deferred?period=YYYY-MM
 */
export async function getDeferredRevenue(period?: string): Promise<DeferredRevenueResponse> {
  const params = period ? `?period=${period}` : '';
  return apiClient.get<DeferredRevenueResponse>(`/v1/billing/revenue/deferred${params}`);
}

/**
 * Get recognized revenue summary for a period.
 * GET /v1/billing/revenue/recognized?period=YYYY-MM
 */
export async function getRecognizedRevenue(period?: string): Promise<RecognizedRevenueResponse> {
  const params = period ? `?period=${period}` : '';
  return apiClient.get<RecognizedRevenueResponse>(`/v1/billing/revenue/recognized${params}`);
}

/**
 * Get full revenue recognition report for a period.
 * GET /v1/billing/revenue/report?period=YYYY-MM
 */
export async function getRevenueReport(period?: string): Promise<RevenueReportResponse> {
  const params = period ? `?period=${period}` : '';
  return apiClient.get<RevenueReportResponse>(`/v1/billing/revenue/report${params}`);
}

/**
 * Get unbilled revenue amount (remaining to be recognized).
 * GET /v1/billing/revenue/unbilled
 */
export async function getUnbilledRevenue(): Promise<UnbilledRevenueResponse> {
  return apiClient.get<UnbilledRevenueResponse>('/v1/billing/revenue/unbilled');
}

/**
 * Manually trigger revenue recognition for a schedule.
 * POST /v1/billing/revenue/recognize
 */
export async function recognizeRevenue(scheduleId: string): Promise<{ status: string }> {
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  return apiClient.post<{ status: string }>(
    '/v1/billing/revenue/recognize',
    { schedule_id: scheduleId },
    { headers }
  );
}

/**
 * Allocate invoice to performance obligations and create recognition schedules.
 * POST /v1/billing/revenue/allocate
 */
export async function allocateRevenue(request: AllocationRequest): Promise<AllocationResponse> {
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  return apiClient.post<AllocationResponse>('/v1/billing/revenue/allocate', request, { headers });
}

// ==================== Affiliate / Referral API ====================

export interface AffiliateCode {
  id: string;
  code: string;
  publisher_id: string;
  tenant_id?: string;
  name: string;
  description?: string;
  commission_type: 'percent' | 'fixed';
  commission_value: number;
  max_commissions?: number;
  max_referrals?: number;
  total_referrals: number;
  total_commissions: number;
  pending_commissions: number;
  pending_earnings_cents: number;
  total_earnings_cents: number;
  paid_out_earnings_cents: number;
  valid_from?: string;
  valid_until?: string;
  is_active: boolean;
  utm_source?: string;
  utm_campaign?: string;
  created_at: string;
  updated_at: string;
}

export interface AffiliateReferral {
  id: string;
  affiliate_code_id: string;
  referred_tenant_id: string;
  subscription_id?: string;
  utm_source?: string;
  utm_campaign?: string;
  utm_content?: string;
  utm_term?: string;
  ip_address?: string;
  user_agent?: string;
  status: 'pending' | 'converted' | 'qualified' | 'canceled';
  referred_at: string;
  converted_at?: string;
  created_at: string;
  updated_at: string;
}

export interface AffiliateCommission {
  id: string;
  affiliate_code_id: string;
  referral_id: string;
  commission_type: 'percent' | 'fixed';
  commission_value: number;
  base_amount_cents: number;
  base_amount_usd: number;
  commission_cents: number;
  commission_usd: number;
  status: 'pending' | 'approved' | 'paid' | 'canceled';
  paid_at?: string;
  payment_batch_id?: string;
  payment_batch?: string;
  subscription_id?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface AffiliateEarningsSummary {
  pending_earnings_cents: number;
  total_earnings_cents: number;
  paid_out_cents: number;
  total_referrals: number;
  pending_commissions: number;
  codes_count: number;
}

export interface ApplyAffiliateCodeResponse {
  success: boolean;
  code: string;
  name: string;
  commission_type: string;
  commission_value: number;
}

export async function getMyAffiliateCodes(): Promise<AffiliateCode[]> {
  return apiClient.get<AffiliateCode[]>('/v1/affiliate/my-codes');
}

export async function getMyAffiliateCommissions(): Promise<AffiliateCommission[]> {
  const res = await apiClient.get<{ commissions: AffiliateCommission[] }>('/v1/affiliate/my-commissions');
  return res?.commissions ?? [];
}

export async function getMyAffiliateReferrals(): Promise<AffiliateReferral[]> {
  return apiClient.get<AffiliateReferral[]>('/v1/affiliate/referrals');
}

// ==================== Founders Types ====================

export interface FounderReferralCode {
  code: string;
  url: string;
  share_url: string;
  total_referrals: number;
  total_commission_earned: number;
  commission_rate: number;
  created_at: string;
}

export interface FounderReferralStats {
  total_referrals: number;
  active_referrals: number;
  total_commission_earned: number;
  total_commission_paid: number;
  pending_commission: number;
  converted_count: number;
  conversion_rate: number;
  average_commission_cents: number;
  referral_details: Array<{
    user_id: string;
    email: string;
    joined_at: string;
    status: string;
    commission_earned: number;
  }>;
}

export async function getMyReferralCode(): Promise<FounderReferralCode> {
  return apiClient.get<FounderReferralCode>('/v1/founders/referral-code');
}

export async function getMyReferralStats(): Promise<FounderReferralStats> {
  return apiClient.get<FounderReferralStats>('/v1/founders/referral-stats');
}

export async function getAffiliateEarningsSummary(): Promise<AffiliateEarningsSummary> {
  return apiClient.get<AffiliateEarningsSummary>('/v1/affiliate/earnings-summary');
}

export async function applyAffiliateCode(request: {
  code: string;
  utm_source?: string;
  utm_campaign?: string;
  utm_content?: string;
  utm_term?: string;
}): Promise<ApplyAffiliateCodeResponse> {
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  return apiClient.post<ApplyAffiliateCodeResponse>('/v1/affiliate/apply-code', request, {
    headers,
  });
}

// ==================== Live Reconciliation (Enterprise) ====================

export interface LiveReconciliationStatus {
  plan: string;
  live_reconciliation_active: boolean;
  auto_reconcile_enabled: boolean;
  scheduled_reconcile_enabled: boolean;
  schedule_cron?: string;
  last_reconciliation_at?: string;
  next_scheduled_reconcile?: string;
  total_reconciliations: number;
  successful_reconciliations: number;
  failed_reconciliations: number;
  audit_export_enabled: boolean;
  soc2_compliant: boolean;
  hipaa_compliant: boolean;
}

export interface LiveReconciliationSettings {
  plan: string;
  auto_reconcile_enabled: boolean;
  scheduled_reconcile_enabled: boolean;
  schedule_cron: string;
  audit_export_enabled: boolean;
  notify_on_completion: boolean;
  notify_on_failure: boolean;
}

export interface UpdateReconciliationSettingsRequest {
  auto_reconcile_enabled: boolean;
  scheduled_reconcile_enabled: boolean;
  schedule_cron: string;
  audit_export_enabled: boolean;
  notify_on_completion: boolean;
  notify_on_failure: boolean;
}

export interface LiveReconciliationUsageResponse {
  plan: string;
  period_start: string;
  period_end: string;
  total_reconciliations: number;
  total_executions_reconciled: number;
  avg_duration_ms: number;
  success_rate: number;
}

export async function getLiveReconciliationStatus(): Promise<LiveReconciliationStatus> {
  return apiClient.get<LiveReconciliationStatus>('/v1/billing/live-reconciliation/status');
}

export async function getLiveReconciliationSettings(): Promise<LiveReconciliationSettings> {
  return apiClient.get<LiveReconciliationSettings>('/v1/billing/live-reconciliation/settings');
}

export async function updateLiveReconciliationSettings(
  settings: UpdateReconciliationSettingsRequest
): Promise<LiveReconciliationSettings> {
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }
  return apiClient.post<LiveReconciliationSettings>(
    '/v1/billing/live-reconciliation/settings',
    settings,
    { headers }
  );
}

export async function getLiveReconciliationUsage(params?: {
  start?: string;
  end?: string;
}): Promise<LiveReconciliationUsageResponse> {
  return apiClient.get<LiveReconciliationUsageResponse>(
    '/v1/billing/live-reconciliation/usage',
    { params }
  );
}
