import type { VaultClient } from "../client";
import type {
  Secret,
  SecretCreate,
  SecretList,
  SecretListOptions,
  SecretRotate,
  SecretUpdate,
} from "../types/secrets";
import { SecretType } from "../types/secrets";

/**
 * SecretsService manages encrypted secrets in the FunctionFly vault.
 */
export class SecretsService {
  constructor(private readonly client: VaultClient) {}

  /**
   * Create persists a new secret. The {@link SecretCreate.encryptedData}
   * must be produced by the caller's encryption layer; the server never
   * sees plaintext.
   */
  public async create(input: SecretCreate): Promise<Secret> {
    if (!Object.values(SecretType).includes(input.secretType)) {
      throw new Error(`invalid secretType: ${String(input.secretType)}`);
    }
    return this.client.request<Secret>("POST", "/v1/vault/secrets", input);
  }

  public async get(id: string): Promise<Secret> {
    return this.client.request<Secret>("GET", `/v1/vault/secrets/${encodeURIComponent(id)}`);
  }

  public async update(id: string, input: SecretUpdate): Promise<Secret> {
    return this.client.request<Secret>("PATCH", `/v1/vault/secrets/${encodeURIComponent(id)}`, input);
  }

  public async rotate(id: string, input: SecretRotate): Promise<Secret> {
    return this.client.request<Secret>(
      "PATCH",
      `/v1/vault/secrets/${encodeURIComponent(id)}/rotate`,
      input,
    );
  }

  public async delete(id: string): Promise<void> {
    await this.client.request<void>("DELETE", `/v1/vault/secrets/${encodeURIComponent(id)}`);
  }

  public async list(options: SecretListOptions = {}): Promise<SecretList> {
    const params: string[] = [];
    if (options.limit !== undefined) params.push(`limit=${options.limit}`);
    if (options.offset !== undefined) params.push(`offset=${options.offset}`);
    if (options.secretType) params.push(`secret_type=${options.secretType}`);
    const qs = params.length ? `?${params.join("&")}` : "";
    return this.client.request<SecretList>("GET", `/v1/vault/secrets${qs}`);
  }
}
