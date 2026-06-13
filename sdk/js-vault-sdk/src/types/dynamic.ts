export enum DynamicDBType {
  Postgres = "postgres",
  MySQL = "mysql",
}

export interface DynamicTarget {
  id: string;
  name: string;
  description?: string;
  dbType: DynamicDBType;
  host: string;
  port: number;
  databaseName: string;
  adminUsername: string;
  sslMode: string;
  allowedRoles?: string[];
  defaultTtlSeconds: number;
  maxTtlSeconds: number;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface DynamicTargetCreate {
  name: string;
  description?: string;
  dbType: DynamicDBType;
  host: string;
  port: number;
  databaseName: string;
  adminUsername: string;
  /** Plaintext password; encrypted server-side. Never returned by the API. */
  adminPassword: string;
  sslMode?: string;
  allowedRoles?: string[];
  defaultTtlSeconds?: number;
  maxTtlSeconds?: number;
}

export interface DynamicTargetsList {
  targets: DynamicTarget[];
  total: number;
}

export interface DynamicCredential {
  id: string;
  targetId: string;
  name: string;
  description?: string;
  roleTemplate?: string;
  ttlSeconds: number;
  maxTtlSeconds: number;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface DynamicCredentialCreate {
  targetId: string;
  name: string;
  description?: string;
  roleTemplate?: string;
  ttlSeconds?: number;
  maxTtlSeconds?: number;
}

export interface GeneratedCredential {
  leaseId: string;
  username: string;
  password: string;
  host: string;
  port: number;
  database: string;
  expiresAt: string;
  credential: DynamicCredential;
  target: DynamicTarget;
}

export interface GenerateOptions {
  ttlSeconds?: number;
}

export interface RenewOptions {
  ttlSeconds?: number;
}
