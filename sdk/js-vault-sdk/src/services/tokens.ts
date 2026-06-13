import type { VaultClient } from "../client";
import type { Token, TokenCreate, TokenList } from "../types/tokens";

export class TokensService {
  constructor(private readonly client: VaultClient) {}

  /** The plaintext token is shown exactly once — store it immediately. */
  public async create(input: TokenCreate): Promise<Token> {
    if (!input.secretId) {
      throw new Error("secretId is required");
    }
    return this.client.request<Token>(
      "POST",
      `/v1/vault/secrets/${encodeURIComponent(input.secretId)}/tokens`,
      input,
    );
  }

  public async list(secretId: string): Promise<TokenList> {
    return this.client.request<TokenList>(
      "GET",
      `/v1/vault/secrets/${encodeURIComponent(secretId)}/tokens`,
    );
  }

  public async revoke(tokenId: string): Promise<void> {
    await this.client.request<void>("DELETE", `/v1/vault/tokens/${encodeURIComponent(tokenId)}`);
  }
}
