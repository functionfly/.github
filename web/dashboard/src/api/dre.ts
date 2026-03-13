import { apiClient } from "./client";

export interface ComponentHashes {
  input: string;
  output: string;
  environment: string;
  dependency: string;
  trace: string;
  resource: string;
  metadata: string;
}

export interface Execution {
  execution_id: string;
  execution_root_hash: string;
  version: string;
  created_at: string;
  replay_verified: boolean;
  roots_match: boolean;
  determinism_tier: string;
  protocol_version: string;
  component_hashes: ComponentHashes;
}

export interface ExecutionCertificate {
  certificate_id: string;
  cert_level: string;
  certificate_hash: string;
  created_at: string;
  anchored: boolean;
}

export interface ExecutionDetail {
  execution_id: string;
  execution_root_hash: string;
  version: string;
  created_at: string;
  determinism_tier: string;
  protocol_version: string;
  replay_verified_at: string | null;
  replay_root_hash: string;
  replay_node_id: string;
  roots_match: boolean;
  component_hashes: ComponentHashes;
  certificate?: ExecutionCertificate;
}

export interface ExecutionListParams {
  limit?: number;
  offset?: number;
  version?: string;
  from?: string; // ISO 8601 datetime
  to?: string; // ISO 8601 datetime
  verified_only?: boolean;
}

export interface ExecutionListResponse {
  function: string;
  executions: Execution[];
  total: number;
  limit: number;
  offset: number;
}

export interface ExecutionDetailResponse {
  execution: ExecutionDetail;
}

// Certificate list (from GET /registry/{author}/{name}/certs)
export interface CertificateListItem {
  certificate_id: string;
  cert_level: string;
  execution_root_hash: string;
  certificate_hash: string;
  anchored: boolean;
  created_at: string;
}

export interface CertificateListResponse {
  function: string;
  certs: CertificateListItem[];
  limit: number;
  offset: number;
}

// Full FXCERT from backend (cert field in GET /registry/.../cert/{cert_id})
export interface FXCertExecutionSection {
  execution_id: string;
  function_id: string;
  owner_id: string;
  caller_id: string;
  node_id: string;
  region: string;
  timestamp_virtual: string;
  timestamp_real_utc: string;
  protocol_version: string;
}

export interface FXCertCapsuleSection {
  capsule_descriptor_hash: string;
  determinism_tier: string;
  protocol_version: string;
}

export interface FXCertIntegritySection {
  execution_root_hash: string;
  input_hash: string;
  environment_hash: string;
  dependency_hash: string;
  trace_hash: string;
  resource_hash: string;
  output_hash: string;
  metadata_hash: string;
  certificate_hash: string;
}

export interface FXCertTrustSection {
  trust_score: number;
  determinism_score: number;
  replay_consistency_score: number;
  drift_incidents_total: number;
  verified_executions_total: number;
}

export interface FXCertSignature {
  algorithm: string;
  public_key: string;
  signature: string;
}

export interface FXCertSignatureSection {
  node_signature?: FXCertSignature | null;
  platform_signature?: FXCertSignature | null;
}

export interface FXCertAnchoringSection {
  anchored: boolean;
  anchor_chain?: string;
  anchor_block_number?: number;
  anchor_tx_hash?: string;
  anchor_merkle_root?: string;
  anchored_at?: string;
}

export interface FXCertRaw {
  fxcert_version: string;
  certificate_id: string;
  execution: FXCertExecutionSection;
  capsule: FXCertCapsuleSection;
  integrity: FXCertIntegritySection;
  trust: FXCertTrustSection;
  signatures: FXCertSignatureSection;
  anchoring?: FXCertAnchoringSection;
  replay_certification?: unknown;
}

export interface CertificateDetailResponse {
  certificate_id: string;
  cert_level: string;
  execution_root_hash: string;
  certificate_hash: string;
  created_at: string;
  anchored: boolean;
  cert: FXCertRaw;
}

export interface CertificateListParams {
  limit?: number;
  offset?: number;
}

class DreApi {
  // List executions for a function
  async listExecutions(
    author: string,
    name: string,
    params?: ExecutionListParams
  ): Promise<ExecutionListResponse> {
    const queryParams = new URLSearchParams();
    if (params?.limit !== undefined)
      queryParams.append("limit", params.limit.toString());
    if (params?.offset !== undefined)
      queryParams.append("offset", params.offset.toString());
    if (params?.version) queryParams.append("version", params.version);
    if (params?.from) queryParams.append("from", params.from);
    if (params?.to) queryParams.append("to", params.to);
    if (params?.verified_only !== undefined)
      queryParams.append("verified_only", params.verified_only.toString());

    const query = queryParams.toString();
    const url = `/v1/registry/${author}/${name}/executions${query ? `?${query}` : ""}`;
    return apiClient.get<ExecutionListResponse>(url);
  }

  // Get single execution details
  async getExecution(
    author: string,
    name: string,
    executionId: string
  ): Promise<ExecutionDetailResponse> {
    return apiClient.get<ExecutionDetailResponse>(
      `/v1/registry/${author}/${name}/executions/${executionId}`
    );
  }

  // Get execution by root hash (for conversation execution refs)
  async getExecutionByHash(
    author: string,
    name: string,
    executionRootHash: string
  ): Promise<ExecutionDetailResponse> {
    const q = new URLSearchParams({ execution_root_hash: executionRootHash });
    return apiClient.get<ExecutionDetailResponse>(
      `/v1/registry/${author}/${name}/executions/by-hash?${q.toString()}`
    );
  }

  // List certificates (FXCERTs) for a function
  async listCertificates(
    author: string,
    name: string,
    params?: CertificateListParams
  ): Promise<CertificateListResponse> {
    const queryParams = new URLSearchParams();
    if (params?.limit !== undefined)
      queryParams.append("limit", params.limit.toString());
    if (params?.offset !== undefined)
      queryParams.append("offset", params.offset.toString());
    const query = queryParams.toString();
    const url = `/v1/registry/${author}/${name}/certs${query ? `?${query}` : ""}`;
    return apiClient.get<CertificateListResponse>(url);
  }

  // Get single FXCERT by certificate_id
  async getCertificate(
    author: string,
    name: string,
    certId: string
  ): Promise<CertificateDetailResponse> {
    return apiClient.get<CertificateDetailResponse>(
      `/v1/registry/${author}/${name}/cert/${encodeURIComponent(certId)}`
    );
  }
}

export const dreApi = new DreApi();
