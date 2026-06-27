/**
 * DEPRECATED: This file is kept as a fallback for API unavailability.
 * The primary source of truth is now the backend API at /v1/content/changelog.
 * Manage entries via the admin dashboard at /admin/changelog.
 *
 * This fallback data contains only the latest 2 releases and should only be
 * used when the API is unreachable (e.g., build-time, network failures).
 */

export type ChangeType = "new" | "improved" | "fixed" | "security" | "breaking";

export interface Change {
  type: ChangeType;
  description: string;
  details?: string;
}

export interface Release {
  version: string;
  date: string;
  isLatest?: boolean;
  changes: Change[];
}

/**
 * @deprecated Use API at /v1/content/changelog instead. This is only a fallback.
 */
export const changelogData: Release[] = [
  {
    version: "v2.4.0",
    date: "2026-06-10",
    isLatest: true,
    changes: [
      {
        type: "new",
        description: "Zero-knowledge Secrets Vault",
        details:
          "Client-side AES-256-GCM encryption. Server never sees plaintext or decryption passphrase.",
      },
      {
        type: "new",
        description: "Trust Score API",
        details:
          "Real-time trust scoring based on execution history, verification level, and revocation status.",
      },
      {
        type: "improved",
        description: "Dashboard Performance",
        details:
          "Reduced initial load time by 60% with code-splitting and lazy loading of non-critical routes.",
      },
      {
        type: "fixed",
        description: "WebAuthn Registration Flow",
        details:
          "Fixed race condition during credential creation that caused intermittent failures.",
      },
    ],
  },
  {
    version: "v2.3.0",
    date: "2026-05-28",
    changes: [
      {
        type: "new",
        description: "MCP Server Integration",
        details:
          "Full Model Context Protocol support for seamless agent-to-tool communication.",
      },
      {
        type: "new",
        description: "Execution Trace Viewer",
        details:
          "Visual timeline of function executions with trust signal attribution.",
      },
      {
        type: "improved",
        description: "Verification Workflows",
        details:
          "Multi-level verification now supports custom policies and team-specific rules.",
      },
      {
        type: "fixed",
        description: "Rate Limiting on Bulk Operations",
        details:
          "Correctly handles distributed rate limits across multiple function providers.",
      },
      {
        type: "security",
        description: "Updated CORS Policy",
        details:
          "Stricter origin validation for cross-origin requests in the trust API.",
      },
    ],
  },
];

export const changeTypeLabels: Record<ChangeType, string> = {
  new: "New",
  improved: "Improved",
  fixed: "Fixed",
  security: "Security",
  breaking: "Breaking",
};

export const changeTypeColors: Record<
  ChangeType,
  { bg: string; border: string; text: string; dot: string }
> = {
  new: {
    bg: "rgba(143, 255, 208, 0.12)",
    border: "rgba(143, 255, 208, 0.3)",
    text: "#8fffd0",
    dot: "#8fffd0",
  },
  improved: {
    bg: "rgba(159, 216, 255, 0.12)",
    border: "rgba(159, 216, 255, 0.3)",
    text: "#9fd8ff",
    dot: "#9fd8ff",
  },
  fixed: {
    bg: "rgba(232, 196, 104, 0.12)",
    border: "rgba(232, 196, 104, 0.3)",
    text: "#e8c468",
    dot: "#e8c468",
  },
  security: {
    bg: "rgba(255, 107, 107, 0.12)",
    border: "rgba(255, 107, 107, 0.3)",
    text: "#ff6b6b",
    dot: "#ff6b6b",
  },
  breaking: {
    bg: "rgba(255, 122, 61, 0.12)",
    border: "rgba(255, 122, 61, 0.3)",
    text: "#ff7a3d",
    dot: "#ff7a3d",
  },
};
