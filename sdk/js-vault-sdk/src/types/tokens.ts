export interface Token {
  tokenId: string;
  /** Plaintext token — shown exactly once. Store immediately. */
  token: string;
  secretId: string;
  name?: string;
  expiresAt: string;
  createdAt: string;
}

export interface TokenInfo {
  id: string;
  secretId: string;
  name?: string;
  expiresAt: string;
  isRevoked: boolean;
  revokedAt?: string;
  revokedReason?: string;
  lastUsedAt?: string;
  useCount: number;
  createdAt: string;
}

export interface TokenCreate {
  secretId: string;
  expiresInHours: number;
  scopes?: string[];
  name?: string;
}

export interface TokenList {
  tokens: TokenInfo[];
  total: number;
}
