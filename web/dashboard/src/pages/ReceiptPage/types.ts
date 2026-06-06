// Types for the Execution Receipt feature. Mirrors the Go
// PublicResponse in internal/api/handlers/receipt/handler.go.
//
// Keep this in sync with the backend — both sides treat the public ID
// as an opaque nanoid.

export type VerificationStatus = "verified" | "failed" | "pending" | "mismatched" | "skipped";

export interface ReceiptFunction {
  name: string;
  author: string;
  runtime: string;
  version: string;
  visibility: string;
  description?: string;
  input_schema?: unknown;
  output_schema?: unknown;
}

export interface ReceiptVerification {
  status: VerificationStatus;
  verified_at?: string;
  error?: string;
}

export interface ReceiptExecution {
  input: unknown;
  output: unknown;
  duration_ms: number;
  cached: boolean;
  created_at: string;
  verification?: ReceiptVerification;
}

export interface ReceiptOGMeta {
  title: string;
  description: string;
  image: string;
}

export interface ReceiptShare {
  url: string;
  embed_url: string;
  tweet_intent_url: string;
  og_meta: ReceiptOGMeta;
}

export interface Receipt {
  id: string;
  function: ReceiptFunction;
  execution: ReceiptExecution;
  share: ReceiptShare;
  can_run: boolean;
  is_paid: boolean;
  price_per_call_usd: number;
}

export interface ReceiptForkPayload {
  function: {
    author: string;
    name: string;
    version: string;
    runtime: string;
  };
  fork: {
    source_b64: string;
    source_bytes: number;
    readme: string;
    editor_url: string;
  };
}

export interface ReceiptRunResponse {
  ok: boolean;
  data?: unknown;
  cached?: boolean;
  duration_ms?: number;
  version?: string;
  execution_id?: string;
  error?: { code: string; message: string };
}

export interface TrendingReceipt {
  id: string;
  function_name: string;
  function_author: string;
  runtime: string;
  view_count: number;
  created_at: string;
  url: string;
}

export interface TrendingReceiptsResponse {
  receipts: TrendingReceipt[];
}
