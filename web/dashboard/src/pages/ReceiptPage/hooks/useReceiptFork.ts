// useReceiptFork — builds the "fork this function" deep link.
//
// Two surfaces:
//   1. Signed-in users go straight to the editor with a `?fork=` query
//      parameter that pre-fills the source.
//   2. Signed-out users are routed through the auth flow with a
//      `?next=` that lands them on the editor post-signup.
//
// The fork payload itself (base64 source) is fetched from
// GET /v1/receipts/:id/fork-payload on demand so the URL doesn't
// balloon for receipt pages with large sources.
import { useQuery } from "@tanstack/react-query";

import { API_URLS } from "@/lib/api-urls";

import { ApiError } from "../lib/api-error";
import type { ReceiptForkPayload } from "../types";

export interface UseReceiptForkInput {
  receiptId: string;
  isAuthenticated: boolean;
  /** Origin of the dashboard SPA — used to build absolute next-URLs. */
  appOrigin?: string;
}

export interface ReceiptForkLink {
  href: string;
  target: "editor" | "signup" | "signin";
  sourceBytes: number;
}

async function fetchForkPayload(receiptId: string, signal?: AbortSignal): Promise<ReceiptForkPayload> {
  const res = await fetch(API_URLS.receipt.fork(receiptId), {
    credentials: "omit",
    headers: { Accept: "application/json" },
    signal,
  });
  if (!res.ok) {
    let code = "UNKNOWN";
    let message = res.statusText || "Failed to load fork payload";
    try {
      const body = await res.json();
      code = body?.error?.code ?? code;
      message = body?.error?.message ?? message;
    } catch {
      // ignore
    }
    throw new ApiError(res.status, code, message);
  }
  return (await res.json()) as ReceiptForkPayload;
}

/**
 * Returns a `buildForkLink()` function that the caller invokes at click
 * time. Pre-fetches the fork payload so the click feels instant.
 */
export function useReceiptFork(input: UseReceiptForkInput) {
  const q = useQuery<ReceiptForkPayload, ApiError>({
    queryKey: ["receipt-fork", input.receiptId],
    queryFn: ({ signal }) => fetchForkPayload(input.receiptId, signal),
    enabled: !!input.receiptId,
    staleTime: 5 * 60_000,
    gcTime: 30 * 60_000,
    retry: (count, err) => {
      const status = (err as ApiError)?.status;
      if (status === 404 || status === 403) return false;
      return count < 1;
    },
  });

  const buildForkLink = (): ReceiptForkLink | null => {
    const payload = q.data;
    if (!payload) return null;
    const origin = input.appOrigin ?? (typeof window !== "undefined" ? window.location.origin : "");
    const editorPath = payload.fork.editor_url;
    const editorAbsolute = origin ? `${origin}${editorPath}` : editorPath;
    if (input.isAuthenticated) {
      return { href: editorAbsolute, target: "editor", sourceBytes: payload.fork.source_bytes };
    }
    const next = encodeURIComponent(editorAbsolute);
    return {
      href: `${origin}/auth/signup?next=${next}`,
      target: "signup",
      sourceBytes: payload.fork.source_bytes,
    };
  };

  return { ...q, buildForkLink };
}
