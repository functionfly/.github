import type { VaultClient } from "../client";
import type { AuditList, AuditListOptions } from "../types/audit";

export class AuditService {
  constructor(private readonly client: VaultClient) {}

  public async list(options: AuditListOptions = {}): Promise<AuditList> {
    const params: string[] = [];
    if (options.secretId) params.push(`secret_id=${encodeURIComponent(options.secretId)}`);
    if (options.action) params.push(`action=${encodeURIComponent(options.action)}`);
    if (options.actorId) params.push(`actor_id=${encodeURIComponent(options.actorId)}`);
    if (options.limit !== undefined) params.push(`limit=${options.limit}`);
    if (options.offset !== undefined) params.push(`offset=${options.offset}`);
    const qs = params.length ? `?${params.join("&")}` : "";
    return this.client.request<AuditList>("GET", `/v1/vault/audit${qs}`);
  }
}
