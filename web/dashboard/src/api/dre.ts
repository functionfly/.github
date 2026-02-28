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
}

export const dreApi = new DreApi();
