// useReceipt — TanStack Query wrapper around GET /v1/receipts/:id
//
// Public, no auth. Caches the result for 30s to match the CDN edge TTL.
import { useQuery, type UseQueryResult } from "@tanstack/react-query";

import { API_URLS } from "@/lib/api-urls";

import { ApiError } from "../lib/api-error";
import type { Receipt } from "../types";

export function useReceipt(execId: string | undefined): UseQueryResult<Receipt, ApiError> {
  return useQuery<Receipt, ApiError>({
    queryKey: ["receipt", execId],
    queryFn: async () => {
      if (!execId) throw new ApiError(0, "MISSING_ID", "Receipt id is required");
      const res = await fetch(API_URLS.receipts.get(execId), {
        credentials: "omit",
        headers: { Accept: "application/json" },
      });
      if (!res.ok) {
        let code = "UNKNOWN";
        let message = res.statusText || "Failed to load receipt";
        try {
          const body = await res.json();
          code = body?.error?.code ?? code;
          message = body?.error?.message ?? message;
        } catch {
          // ignore JSON parse errors
        }
        throw new ApiError(res.status, code, message);
      }
      return (await res.json()) as Receipt;
    },
    enabled: !!execId,
    staleTime: 30_000,
    gcTime: 5 * 60_000,
    retry: (count, err) => {
      const status = (err as ApiError)?.status;
      if (status === 404 || status === 410) return false;
      return count < 2;
    },
  });
}
