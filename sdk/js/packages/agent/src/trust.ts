/**
 * Trust score utilities for FunctionFly Agent SDK
 */

import { TrustTier } from "./types.js";

/**
 * Maps trust score (0-100) to trust tier
 */
export function getTrustTierFromScore(
  score: number,
  isVerified: boolean = false,
): TrustTier {
  if (score >= 90) {
    return isVerified ? TrustTier.HighlyTrusted : TrustTier.Verified;
  } else if (score >= 70) {
    return TrustTier.Verified;
  } else if (score >= 50) {
    return TrustTier.Trusted;
  }
  return TrustTier.Untrusted;
}

/**
 * Checks if a trust score meets the minimum threshold
 */
export function meetsTrustThreshold(
  trustScore: number,
  minTrustScore?: number,
  preferredTrustTier?: TrustTier,
): boolean {
  if (minTrustScore !== undefined && trustScore < minTrustScore) {
    return false;
  }

  if (preferredTrustTier !== undefined) {
    const currentTier = getTrustTierFromScore(trustScore);
    const tierPriority: Record<TrustTier, number> = {
      [TrustTier.HighlyTrusted]: 4,
      [TrustTier.Verified]: 3,
      [TrustTier.Trusted]: 2,
      [TrustTier.Untrusted]: 1,
      [TrustTier.Unknown]: 0,
    };

    if (tierPriority[currentTier] < tierPriority[preferredTrustTier]) {
      return false;
    }
  }

  return true;
}

/**
 * Sorts functions by trust score (descending)
 */
export function sortByTrustScore<T extends { trustScore: number }>(
  functions: T[],
): T[] {
  return [...functions].sort((a, b) => b.trustScore - a.trustScore);
}

/**
 * Filters functions by trust score threshold
 */
export function filterByTrustScore<T extends { trustScore: number }>(
  functions: T[],
  minTrustScore: number,
): T[] {
  return functions.filter((f) => f.trustScore >= minTrustScore);
}

/**
 * Selects the best fallback function based on trust score
 */
export function selectBestFallback<
  T extends { trustScore: number; id: string },
>(candidates: T[], excludeIds: string[] = []): T | undefined {
  const available = candidates.filter((c) => !excludeIds.includes(c.id));
  return sortByTrustScore(available)[0];
}

/**
 * Calculates exponential backoff delay
 */
export function calculateBackoffDelay(
  attempt: number,
  baseDelayMs: number,
  maxDelayMs: number,
  multiplier: number,
): number {
  const delay = Math.min(
    baseDelayMs * Math.pow(multiplier, attempt),
    maxDelayMs,
  );
  // Add jitter to prevent thundering herd
  return delay * (0.5 + Math.random() * 0.5);
}

/**
 * Default retry configuration
 */
export const DEFAULT_RETRY_CONFIG = {
  maxRetries: 3,
  baseDelayMs: 1000,
  maxDelayMs: 10000,
  backoffMultiplier: 2,
  retryOnTimeout: true,
  retryOnError: true,
};

/**
 * Parses trust tier from API response string
 */
export function parseTrustTier(tierString: string | undefined): TrustTier {
  if (!tierString) return TrustTier.Unknown;

  const normalized = tierString.toLowerCase().replace(" ", "_");

  switch (normalized) {
    case "highly_trusted":
    case "highly-trusted":
      return TrustTier.HighlyTrusted;
    case "verified":
      return TrustTier.Verified;
    case "trusted":
      return TrustTier.Trusted;
    case "untrusted":
      return TrustTier.Untrusted;
    default:
      return TrustTier.Unknown;
  }
}

/**
 * Formats trust score for display
 */
export function formatTrustScore(score: number): string {
  if (score >= 90) return "Excellent";
  if (score >= 70) return "Good";
  if (score >= 50) return "Fair";
  return "Low";
}

/**
 * Gets trust level color for UI (hex values)
 */
export function getTrustLevelColor(score: number): string {
  if (score >= 90) return "#10b981"; // green
  if (score >= 70) return "#3b82f6"; // blue
  if (score >= 50) return "#f59e0b"; // amber
  return "#ef4444"; // red
}
