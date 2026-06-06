// OG meta + share-text builders. Mirrors the server-side copy in
// internal/api/handlers/receipt/handler.go so client-side fallbacks stay
// consistent when the API is offline.

import type { Receipt } from "../types";

export interface ShareTextOptions {
  variant?: "default" | "milestone";
  threshold?: number;
}

export function buildShareText(receipt: Receipt, opts: ShareTextOptions = {}): string {
  const { function: fn, execution } = receipt;
  if (opts.variant === "milestone" && opts.threshold) {
    return `🎉 My function just hit ${opts.threshold} executions on @functionfly — see the public receipt:`;
  }
  if (fn.description) {
    const d = fn.description.length > 180 ? fn.description.slice(0, 177) + "..." : fn.description;
    return `I just ran ${fn.author}/${fn.name} on @functionfly — ${d}`;
  }
  return `I just ran ${fn.author}/${fn.name} on @functionfly in ${execution.duration_ms}ms.`;
}

export function buildOgTitle(receipt: Receipt): string {
  const { function: fn, execution } = receipt;
  return `${fn.author}/${fn.name} ran in ${execution.duration_ms}ms · FunctionFly`;
}

export function buildOgDescription(receipt: Receipt): string {
  return buildShareText(receipt);
}

/**
 * Encode a JSON value as a URL-safe base64 string. Used for the
 * "Run with this input" pre-fill on the playground link.
 */
export function encodeInputForUrl(value: unknown): string {
  try {
    const json = JSON.stringify(value);
    if (typeof window === "undefined") {
      // Node-side fallback (SSR/edge): use Buffer.
      return Buffer.from(json, "utf-8").toString("base64")
        .replace(/\+/g, "-")
        .replace(/\//g, "_")
        .replace(/=+$/, "");
    }
    return btoa(unescape(encodeURIComponent(json)))
      .replace(/\+/g, "-")
      .replace(/\//g, "_")
      .replace(/=+$/, "");
  } catch {
    return "";
  }
}
