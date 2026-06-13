import type { VaultClient } from "../client";
import {
  DynamicDBType,
  type DynamicCredential,
  type DynamicCredentialCreate,
  type DynamicTarget,
  type DynamicTargetCreate,
  type DynamicTargetsList,
  type GeneratedCredential,
  type GenerateOptions,
  type RenewOptions,
} from "../types/dynamic";

export class DynamicTargetsService {
  constructor(private readonly client: VaultClient) {}

  public async create(input: DynamicTargetCreate): Promise<DynamicTarget> {
    if (!input.adminPassword) {
      throw new Error("adminPassword is required");
    }
    if (input.dbType !== DynamicDBType.Postgres && input.dbType !== DynamicDBType.MySQL) {
      throw new Error(`invalid dbType: ${String(input.dbType)}`);
    }
    return this.client.request<DynamicTarget>(
      "POST",
      "/v1/vault/dynamic-secret-targets",
      input,
    );
  }

  public async list(): Promise<DynamicTargetsList> {
    return this.client.request<DynamicTargetsList>("GET", "/v1/vault/dynamic-secret-targets");
  }

  public async delete(id: string): Promise<void> {
    await this.client.request<void>(
      "DELETE",
      `/v1/vault/dynamic-secret-targets/${encodeURIComponent(id)}`,
    );
  }

  public async test(id: string): Promise<void> {
    await this.client.request<void>(
      "POST",
      `/v1/vault/dynamic-secret-targets/${encodeURIComponent(id)}/test`,
    );
  }
}

export class DynamicCredentialsService {
  constructor(private readonly client: VaultClient) {}

  public async create(input: DynamicCredentialCreate): Promise<DynamicCredential> {
    if (!input.targetId) {
      throw new Error("targetId is required");
    }
    return this.client.request<DynamicCredential>("POST", "/v1/vault/dynamic-credentials", input);
  }

  public async generate(credentialId: string, options: GenerateOptions = {}): Promise<GeneratedCredential> {
    return this.client.request<GeneratedCredential>(
      "POST",
      `/v1/vault/dynamic-credentials/${encodeURIComponent(credentialId)}/generate`,
      Object.keys(options).length > 0 ? options : undefined,
    );
  }

  public async revokeAll(credentialId: string): Promise<void> {
    await this.client.request<void>(
      "POST",
      `/v1/vault/dynamic-credentials/${encodeURIComponent(credentialId)}/revoke`,
    );
  }
}

export class LeasesService {
  constructor(private readonly client: VaultClient) {}

  public async renew(credentialId: string, leaseId: string, options: RenewOptions = {}): Promise<{ leaseId: string; expiresAt: string }> {
    return this.client.request<{ leaseId: string; expiresAt: string }>(
      "POST",
      `/v1/vault/dynamic-credentials/${encodeURIComponent(credentialId)}/leases/${encodeURIComponent(leaseId)}/renew`,
      Object.keys(options).length > 0 ? options : undefined,
    );
  }

  public async revoke(credentialId: string, leaseId: string): Promise<void> {
    await this.client.request<void>(
      "POST",
      `/v1/vault/dynamic-credentials/${encodeURIComponent(credentialId)}/leases/${encodeURIComponent(leaseId)}/revoke`,
    );
  }
}
