// useReceiptRun — TanStack Query mutation for POST /v1/receipts/:id/run
//
// Calls the function with the recorded input (or an override), returns
// the new execution_id so the caller can build a fresh /r/:id URL.
import { useMutation, type UseMutationResult } from "@tanstack/react-query";

import { API_URLS } from "@/lib/api-urls";

import { ApiError } from "../lib/api-error";
import type { ReceiptRunResponse } from "../types";

export interface UseReceiptRunInput {
  receiptId: string;
  input?: unknown;
  signal?: AbortSignal;
}

async function runReceipt(input: UseReceiptRunInput): Promise<ReceiptRunResponse> {
  const res = await fetch(API_URLS.receipt.run(input.receiptId), {
    method: "POST",
    credentials: "omit",
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    body: JSON.stringify({ input: input.input ?? null }),
    signal: input.signal,
  });
  // The /run endpoint delegates to /v1/fx/... and returns the same
  // ExecutionResponse shape. We always parse the body, even on error,
  // so the UI can show the error code.
  let body: ReceiptRunResponse | null = null;
  try {
    body = (await res.json()) as ReceiptRunResponse;
  } catch {
    // ignore
  }
  if (!res.ok || (body && body.ok === false)) {
    const code = body?.error?.code ?? `HTTP_${res.status}`;
    const message = body?.error?.message ?? res.statusText ?? "Execution failed";
    throw new ApiError(res.status, code, message);
  }
  return body ?? { ok: false, error: { code: "EMPTY_RESPONSE", message: "Empty response" } };
}

export function useReceiptRun(): UseMutationResult<ReceiptRunResponse, ApiError, UseReceiptRunInput> {
  return useMutation<ReceiptRunResponse, ApiError, UseReceiptRunInput>({
    mutationFn: runReceipt,
  });
}
