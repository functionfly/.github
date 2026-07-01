import { apiClient } from './client';

// ==================== Types ====================

export interface TrustAPITierPricing {
  tier: string;
  monthly_price_usd: number;
  included_requests: number;
  overage_price_per_1000: number;
  has_overage_billing: boolean;
  rate_limit_per_minute: number;
  rate_limit_per_day: number;
  monthly_request_limit: number;
  description: string;
}

export interface TrustAPITierPricingResponse {
  tiers: TrustAPITierPricing[];
}

export interface TrustAPIBillingStatus {
  partner_id: string;
  tier: string;
  billing_status: string;
  is_founder_mode: boolean;
  monthly_price_usd: number;
  included_requests: number;
  current_usage: number;
  remaining_requests: number;
  overage_requests: number;
  overage_charge_usd: number;
  billing_period_start: string;
  billing_period_end: string;
  is_hard_limit: boolean;
  founder_mode_started_at?: string;
  founder_mode_ends_at?: string;
  usage_threshold?: number;
}

export interface TrustAPICheckoutRequest {
  tier: string;
  success_url?: string;
  cancel_url?: string;
}

export interface TrustAPICheckoutResponse {
  session_id: string;
  url: string;
  status: string;
}

export interface TrustAPIUsageRecord {
  date: string;
  requests: number;
  cost_usd: number;
}

export interface TrustAPIUsageReport {
  partner_id: string;
  tier: string;
  period_start: string;
  period_end: string;
  total_requests: number;
  included_requests: number;
  overage_requests: number;
  overage_cost_usd: number;
  total_cost_usd: number;
  records: TrustAPIUsageRecord[];
}

export interface TrustAPIInvoice {
  id: string;
  partner_id: string;
  stripe_invoice_id: string | null;
  amount: number;
  currency: string;
  status: string;
  invoice_date: string | null;
  due_date: string | null;
  invoice_pdf: string | null;
  hosted_invoice_url: string | null;
  created_at: string;
}

export interface TrustAPIInvoicesResponse {
  invoices: TrustAPIInvoice[];
  limit: number;
  offset: number;
  total: number;
}

export interface TrustAPIPartnerCreateRequest {
  name: string;
  slug: string;
  description?: string;
  contact_email: string;
  contact_name?: string;
  website_url?: string;
  tier?: string;
}

export interface TrustAPIPartnerResponse {
  id: string;
  name: string;
  slug: string;
  description: string;
  contact_email: string;
  contact_name: string;
  website_url: string;
  tier: string;
  rate_limit_per_minute: number;
  rate_limit_per_day: number;
  monthly_request_limit: number;
  current_month_usage: number;
  status: string;
  created_at: string;
  activated_at: string | null;
}

export interface TrustAPIFounderModeRequest {
  mode_type: 'time_based' | 'revenue_based' | 'hybrid';
  free_days: number;
  mrr_threshold: number;
  success_url: string;
  cancel_url: string;
}

export interface TrustAPIFounderModeStatus {
  id: string;
  partner_id: string;
  mode_type: string;
  status: string;
  started_at: string;
  ends_at?: string;
  free_days: number;
  mrr_threshold_cents: number;
  days_remaining: number;
}

// ==================== Helper Functions ====================

export function getTrustAPIErrorMessage(e: unknown, defaultMsg = 'An error occurred'): string {
  if (e && typeof e === 'object' && 'response' in e) {
    const res = (e as { response?: { data?: { error?: string; message?: string }; status?: number } }).response;
    if (res?.data?.error) return res.data.error;
    if (res?.data?.message) return res.data.message;
    if (res?.status === 401) return 'Unauthorized. Please log in again.';
    if (res?.status === 403) return 'Access denied.';
    if (res?.status === 404) return 'Resource not found.';
    if (res?.status === 503) return 'Billing is not configured. Contact support.';
  }
  return defaultMsg;
}

// ==================== API Functions ====================

/**
 * Get all Trust API tier pricing.
 * GET /v1/partners/tiers
 */
export async function getTrustAPITierPricing(): Promise<TrustAPITierPricingResponse> {
  return apiClient.get<TrustAPITierPricingResponse>('/v1/partners/tiers');
}

/**
 * Get current partner's billing status.
 * GET /v1/partners/{partner_id}/billing
 */
export async function getTrustAPIBillingStatus(): Promise<TrustAPIBillingStatus> {
  return apiClient.get<TrustAPIBillingStatus>('/v1/partners/me/billing');
}

/**
 * Create a Stripe checkout session for tier upgrade.
 * POST /v1/partners/{partner_id}/billing/checkout
 */
export async function createTrustAPICheckout(
  tier: string,
  successUrl?: string,
  cancelUrl?: string
): Promise<TrustAPICheckoutResponse> {
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  const body: TrustAPICheckoutRequest = {
    tier,
  };
  if (successUrl) body.success_url = successUrl;
  if (cancelUrl) body.cancel_url = cancelUrl;

  return apiClient.post<TrustAPICheckoutResponse>(
    '/v1/partners/me/billing/checkout',
    body,
    { headers }
  );
}

/**
 * Get partner's usage report.
 * GET /v1/partners/{partner_id}/billing/usage
 */
export async function getTrustAPIUsageReport(): Promise<TrustAPIUsageReport> {
  return apiClient.get<TrustAPIUsageReport>('/v1/partners/me/billing/usage');
}

/**
 * Get partner's invoices.
 * GET /v1/partners/{partner_id}/billing/invoices
 */
export async function getTrustAPIInvoices(
  limit = 10,
  offset = 0
): Promise<TrustAPIInvoicesResponse> {
  const params = new URLSearchParams({
    limit: limit.toString(),
    offset: offset.toString(),
  });
  return apiClient.get<TrustAPIInvoicesResponse>(`/v1/partners/me/billing/invoices?${params}`);
}

/**
 * Enroll in founder mode.
 * POST /v1/partners/{partner_id}/founder
 */
export async function enrollTrustAPIFounderMode(
  data: TrustAPIFounderModeRequest
): Promise<TrustAPIFounderModeStatus> {
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) {
    headers['X-CSRF-Token'] = csrfToken;
  }

  return apiClient.post<TrustAPIFounderModeStatus>(
    '/v1/partners/me/founder',
    data,
    { headers }
  );
}

/**
 * Register a new Trust API partner.
 * POST /v1/partners
 */
export async function createTrustAPIPartner(
  data: TrustAPIPartnerCreateRequest
): Promise<TrustAPIPartnerResponse> {
  return apiClient.post<TrustAPIPartnerResponse>('/v1/partners', data);
}

// ==================== Tier Display Helpers ====================

export const TRUST_API_TIER_LABELS: Record<string, string> = {
  developer: 'Developer',
  payg: 'Pay-as-you-Go',
  startup: 'Startup',
  business: 'Business',
  enterprise: 'Enterprise',
};

export const TRUST_API_TIER_COLORS: Record<string, string> = {
  developer: '#6B7280',
  payg: '#3B82F6',
  startup: '#10B981',
  business: '#F59E0B',
  enterprise: '#8B5CF6',
};

export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(amount);
}

export function formatNumber(num: number): string {
  if (num >= 1000000) {
    return `${(num / 1000000).toFixed(1)}M`;
  }
  if (num >= 1000) {
    return `${(num / 1000).toFixed(0)}K`;
  }
  return num.toString();
}

export function getUsagePercentage(current: number, included: number): number {
  if (included === 0) return 0;
  return Math.min((current / included) * 100, 100);
}

// ==================== Attestation Types ====================

export type AttestationType =
  | 'verification'
  | 'security_scan'
  | 'code_review'
  | 'execution'
  | 'compliance'
  | 'signature'
  | 'delegation';

export type AttestationStatus = 'valid' | 'revoked' | 'expired';

export interface Attestation {
  id: string;
  attestation_id: string;
  function_id: string;
  function_version?: string;
  function_author?: string;
  function_name?: string;
  type: AttestationType;
  status: AttestationStatus;
  title: string;
  description?: string;
  results?: Record<string, unknown>;
  attester_id: string;
  attester_type: string;
  attester_name?: string;
  verification_level?: string;
  proof_hash: string;
  signature?: string;
  public_key_id?: string;
  code_hash?: string;
  input_hash?: string;
  output_hash?: string;
  source_data_hash?: string;
  previous_hash?: string;
  attested_at: string;
  valid_until?: string;
  revoked_at?: string;
  revoke_reason?: string;
  is_valid: boolean;
  signature_valid: boolean;
  chain_valid: boolean;
}

export interface AttestationListResponse {
  attestations: Attestation[];
  total_count: number;
  page: number;
  page_size: number;
}

export interface AttestationCreateRequest {
  function_id: string;
  function_version?: string;
  type: AttestationType;
  title: string;
  description?: string;
  results?: Record<string, unknown>;
  attester_type?: string;
  attester_name?: string;
  verification_level?: string;
  code_hash?: string;
  input_hash?: string;
  output_hash?: string;
  source_data_hash?: string;
  valid_until?: string;
}

export interface AttestationPublicKey {
  public_key: string;
  key_id: string;
  algorithm: string;
  key_encoding: string;
}

export interface ChainVerificationResult {
  function_id: string;
  chain_valid: boolean;
  chain_length: number;
  verified_at: string;
  signing_key_id: string;
  algorithm: string;
}

export const ATTESTATION_TYPE_LABELS: Record<AttestationType, string> = {
  verification: 'Verification',
  security_scan: 'Security Scan',
  code_review: 'Code Review',
  execution: 'Execution',
  compliance: 'Compliance',
  signature: 'Signature',
  delegation: 'Delegation',
};

export const ATTESTATION_TYPE_COLORS: Record<AttestationType, string> = {
  verification: '#10B981',
  security_scan: '#EF4444',
  code_review: '#3B82F6',
  execution: '#8B5CF6',
  compliance: '#F59E0B',
  signature: '#6366F1',
  delegation: '#14B8A6',
};

// ==================== Attestation API Functions ====================

/**
 * List attestations for a function.
 * GET /v1/trust/attestations?function_id=...
 */
export async function listAttestations(
  functionId: string,
  params?: { type?: string; status?: string; include_revoked?: boolean; page?: number; page_size?: number }
): Promise<AttestationListResponse> {
  const searchParams = new URLSearchParams({ function_id: functionId });
  if (params?.type) searchParams.set('type', params.type);
  if (params?.status) searchParams.set('status', params.status);
  if (params?.include_revoked) searchParams.set('include_revoked', 'true');
  if (params?.page) searchParams.set('page', params.page.toString());
  if (params?.page_size) searchParams.set('page_size', params.page_size.toString());
  return apiClient.get<AttestationListResponse>(`/v1/trust/attestations?${searchParams}`);
}

/**
 * Get a specific attestation by ID.
 * GET /v1/trust/attestations/{attestation_id}
 */
export async function getAttestation(attestationId: string): Promise<Attestation> {
  return apiClient.get<Attestation>(`/v1/trust/attestations/${attestationId}`);
}

/**
 * Create a new attestation.
 * POST /v1/trust/attestations
 */
export async function createAttestation(
  data: AttestationCreateRequest
): Promise<Attestation> {
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) headers['X-CSRF-Token'] = csrfToken;
  return apiClient.post<Attestation>('/v1/trust/attestations', data, { headers });
}

/**
 * Revoke an attestation.
 * POST /v1/trust/attestations/{attestation_id}/revoke
 */
export async function revokeAttestation(
  attestationId: string,
  reason: string
): Promise<{ attestation_id: string; status: string; revoked_at: string; revoke_reason: string }> {
  const csrfToken = await apiClient.fetchCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) headers['X-CSRF-Token'] = csrfToken;
  return apiClient.post(`/v1/trust/attestations/${attestationId}/revoke`, { reason }, { headers });
}

/**
 * Verify an attestation's integrity and signature.
 * GET /v1/trust/attestations/{attestation_id}/verify
 */
export async function verifyAttestation(
  attestationId: string
): Promise<{ attestation_id: string; integrity_verified: boolean; signature_verified: boolean; verified_at: string }> {
  return apiClient.get(`/v1/trust/attestations/${attestationId}/verify`);
}

/**
 * Get the full attestation chain for a function.
 * GET /v1/trust/attestations/chain/{function_id}
 */
export async function getAttestationChain(
  functionId: string
): Promise<{ function_id: string; chain_length: number; chain_valid: boolean; attestations: Attestation[] }> {
  return apiClient.get(`/v1/trust/attestations/chain/${functionId}`);
}

/**
 * Verify the full attestation chain for a function.
 * GET /v1/trust/attestations/chain/{function_id}/verify
 */
export async function verifyAttestationChain(functionId: string): Promise<ChainVerificationResult> {
  return apiClient.get<ChainVerificationResult>(`/v1/trust/attestations/chain/${functionId}/verify`);
}

/**
 * Get the public key for external attestation verification.
 * GET /v1/trust/attestations/public-key
 */
export async function getAttestationPublicKey(): Promise<AttestationPublicKey> {
  return apiClient.get<AttestationPublicKey>('/v1/trust/attestations/public-key');
}

// ==================== Merkle Audit Trail Types ====================

export interface MerkleTreeHead {
  tree_size: number;
  root_hash: string;
  previous_hash?: string;
  timestamp: string;
  signature?: string;
  public_key_id?: string;
  metadata?: Record<string, unknown>;
}

export interface MerkleInclusionProof {
  leaf_index: number;
  leaf_hash: string;
  tree_size: number;
  root_hash: string;
  path: string[];
}

export interface MerkleConsistencyProof {
  old_size: number;
  new_size: number;
  old_root: string;
  new_root: string;
  path: string[];
}

export interface MerkleRoot {
  root_hash: string;
  tree_size: number;
  algorithm: string;
  format: string;
}

// ==================== Merkle Audit Trail API ====================

/**
 * Get the latest signed Merkle tree head.
 * GET /v1/trust/merkle/head
 */
export async function getMerkleTreeHead(): Promise<MerkleTreeHead> {
  return apiClient.get<MerkleTreeHead>('/v1/trust/merkle/head');
}

/**
 * Get the current Merkle root hash.
 * GET /v1/trust/merkle/root
 */
export async function getMerkleRoot(): Promise<MerkleRoot> {
  return apiClient.get<MerkleRoot>('/v1/trust/merkle/root');
}

/**
 * Get an inclusion proof for a specific leaf.
 * GET /v1/trust/merkle/inclusion?leaf_index=N
 */
export async function getMerkleInclusionProof(leafIndex: number): Promise<MerkleInclusionProof> {
  return apiClient.get<MerkleInclusionProof>(`/v1/trust/merkle/inclusion?leaf_index=${leafIndex}`);
}

/**
 * Get a consistency proof between an old tree size and current.
 * GET /v1/trust/merkle/consistency?old_size=N
 */
export async function getMerkleConsistencyProof(oldSize: number): Promise<MerkleConsistencyProof> {
  return apiClient.get<MerkleConsistencyProof>(`/v1/trust/merkle/consistency?old_size=${oldSize}`);
}

/**
 * Verify an inclusion proof client-side.
 * POST /v1/trust/merkle/verify/inclusion
 */
export async function verifyMerkleInclusion(
  proof: MerkleInclusionProof
): Promise<{ valid: boolean; leaf_index: number; tree_size: number; root_hash: string }> {
  return apiClient.post('/v1/trust/merkle/verify/inclusion', proof);
}

// ==================== Delegation Chain of Custody ====================

export interface DelegationChain {
  chain_id: string;
  chain_valid: boolean;
  chain_length: number;
  attestations: DelegationChainEntry[];
}

export interface DelegationChainEntry {
  attestation_id: string;
  depth: number;
  function_id: string;
  function_name?: string;
  function_author?: string;
  delegator_function_id?: string;
  delegator_agent_id?: string;
  delegator_trust_score?: number;
  delegation_input_hash?: string;
  delegation_output_hash?: string;
  proof_hash: string;
  signature?: string;
  parent_attestation_id?: string;
  attested_at: string;
  integrity_verified: boolean;
}

export interface DelegationChainSummary {
  chain_id: string;
  chain_valid: boolean;
  chain_length: number;
}

/**
 * Get the full chain of custody for a delegation sequence.
 * GET /v1/trust/delegation/chain/{chain_id}
 */
export async function getDelegationChain(chainId: string): Promise<DelegationChain> {
  return apiClient.get<DelegationChain>(`/v1/trust/delegation/chain/${chainId}`);
}

/**
 * Verify a delegation chain's cryptographic integrity.
 * GET /v1/trust/delegation/chain/{chain_id}/verify
 */
export async function verifyDelegationChain(
  chainId: string
): Promise<{ chain_id: string; chain_valid: boolean; chain_length: number; signing_key_id: string; algorithm: string; verified_at: string }> {
  return apiClient.get(`/v1/trust/delegation/chain/${chainId}/verify`);
}

/**
 * Get all delegation chains a function participated in.
 * GET /v1/trust/delegation/function/{function_id}
 */
export async function getFunctionDelegationChains(
  functionId: string
): Promise<{ function_id: string; chains: DelegationChainSummary[]; total: number }> {
  return apiClient.get(`/v1/trust/delegation/function/${functionId}`);
}