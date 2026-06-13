export interface AuditEntry {
  id: string;
  secretId?: string;
  action: string;
  actorId: string;
  actorType: string;
  ipAddress?: string;
  userAgent?: string;
  metadata?: Record<string, unknown>;
  success: boolean;
  errorMessage?: string;
  createdAt: string;
}

export interface AuditListOptions {
  secretId?: string;
  action?: string;
  actorId?: string;
  limit?: number;
  offset?: number;
}

export interface AuditList {
  entries: AuditEntry[];
  total: number;
  limit: number;
  offset: number;
}
