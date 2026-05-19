/**
 * Unit tests for trust utilities
 */

import { describe, expect, it } from "vitest";
import {
  calculateBackoffDelay,
  DEFAULT_RETRY_CONFIG,
  filterByTrustScore,
  formatTrustScore,
  getTrustLevelColor,
  getTrustTierFromScore,
  meetsTrustThreshold,
  parseTrustTier,
  selectBestFallback,
  sortByTrustScore,
} from "./trust.js";
import { TrustTier } from "./types.js";

describe("Trust Utilities", () => {
  describe("getTrustTierFromScore", () => {
    it("should return highly_trusted for score >= 90 with verification", () => {
      expect(getTrustTierFromScore(90, true)).toBe(TrustTier.HighlyTrusted);
      expect(getTrustTierFromScore(95, true)).toBe(TrustTier.HighlyTrusted);
      expect(getTrustTierFromScore(100, true)).toBe(TrustTier.HighlyTrusted);
    });

    it("should return verified for score >= 90 without verification", () => {
      expect(getTrustTierFromScore(90, false)).toBe(TrustTier.Verified);
      expect(getTrustTierFromScore(95, false)).toBe(TrustTier.Verified);
    });

    it("should return verified for score >= 70", () => {
      expect(getTrustTierFromScore(70, false)).toBe(TrustTier.Verified);
      expect(getTrustTierFromScore(85, false)).toBe(TrustTier.Verified);
    });

    it("should return trusted for score >= 50", () => {
      expect(getTrustTierFromScore(50, false)).toBe(TrustTier.Trusted);
      expect(getTrustTierFromScore(65, false)).toBe(TrustTier.Trusted);
    });

    it("should return untrusted for score < 50", () => {
      expect(getTrustTierFromScore(0, false)).toBe(TrustTier.Untrusted);
      expect(getTrustTierFromScore(49, false)).toBe(TrustTier.Untrusted);
    });
  });

  describe("meetsTrustThreshold", () => {
    it("should pass when score meets minimum", () => {
      expect(meetsTrustThreshold(80, 70)).toBe(true);
      expect(meetsTrustThreshold(70, 70)).toBe(true);
    });

    it("should fail when score is below minimum", () => {
      expect(meetsTrustThreshold(60, 70)).toBe(false);
      expect(meetsTrustThreshold(50, 70)).toBe(false);
    });

    it("should pass when no minimum is set", () => {
      expect(meetsTrustThreshold(30)).toBe(true);
    });

    it("should consider preferred trust tier", () => {
      expect(meetsTrustThreshold(75, undefined, TrustTier.Verified)).toBe(true);
      expect(meetsTrustThreshold(65, undefined, TrustTier.Verified)).toBe(
        false,
      );
      expect(meetsTrustThreshold(95, undefined, TrustTier.Verified)).toBe(true);
    });
  });

  describe("sortByTrustScore", () => {
    it("should sort functions in descending order", () => {
      const functions = [
        { trustScore: 50 },
        { trustScore: 90 },
        { trustScore: 70 },
      ];

      const sorted = sortByTrustScore(functions);

      expect(sorted[0].trustScore).toBe(90);
      expect(sorted[1].trustScore).toBe(70);
      expect(sorted[2].trustScore).toBe(50);
    });

    it("should not mutate original array", () => {
      const functions = [{ trustScore: 50 }, { trustScore: 90 }];

      const sorted = sortByTrustScore(functions);

      expect(functions[0].trustScore).toBe(50);
      expect(sorted).not.toBe(functions);
    });
  });

  describe("filterByTrustScore", () => {
    it("should filter functions by minimum score", () => {
      const functions = [
        { trustScore: 30 },
        { trustScore: 60 },
        { trustScore: 80 },
      ];

      const filtered = filterByTrustScore(functions, 60);

      expect(filtered.length).toBe(2);
      expect(filtered[0].trustScore).toBe(60);
      expect(filtered[1].trustScore).toBe(80);
    });
  });

  describe("selectBestFallback", () => {
    it("should select highest trust score function", () => {
      const candidates = [
        { id: "1", trustScore: 60 },
        { id: "2", trustScore: 90 },
        { id: "3", trustScore: 75 },
      ];

      const selected = selectBestFallback(candidates);

      expect(selected?.id).toBe("2");
    });

    it("should exclude specified IDs", () => {
      const candidates = [
        { id: "1", trustScore: 60 },
        { id: "2", trustScore: 90 },
        { id: "3", trustScore: 75 },
      ];

      const selected = selectBestFallback(candidates, ["2"]);

      expect(selected?.id).toBe("3");
    });

    it("should return undefined for empty array", () => {
      const selected = selectBestFallback([]);
      expect(selected).toBeUndefined();
    });
  });

  describe("calculateBackoffDelay", () => {
    it("should calculate exponential backoff", () => {
      const delay1 = calculateBackoffDelay(0, 1000, 10000, 2);
      const delay2 = calculateBackoffDelay(1, 1000, 10000, 2);
      const delay3 = calculateBackoffDelay(2, 1000, 10000, 2);

      expect(delay1).toBeGreaterThan(0);
      expect(delay2).toBeGreaterThan(delay1);
      expect(delay3).toBeGreaterThan(delay2);
    });

    it("should not exceed max delay", () => {
      const delay = calculateBackoffDelay(10, 1000, 5000, 2);
      expect(delay).toBeLessThanOrEqual(5000);
    });
  });

  describe("DEFAULT_RETRY_CONFIG", () => {
    it("should have correct default values", () => {
      expect(DEFAULT_RETRY_CONFIG.maxRetries).toBe(3);
      expect(DEFAULT_RETRY_CONFIG.baseDelayMs).toBe(1000);
      expect(DEFAULT_RETRY_CONFIG.maxDelayMs).toBe(10000);
      expect(DEFAULT_RETRY_CONFIG.backoffMultiplier).toBe(2);
      expect(DEFAULT_RETRY_CONFIG.retryOnTimeout).toBe(true);
      expect(DEFAULT_RETRY_CONFIG.retryOnError).toBe(true);
    });
  });

  describe("parseTrustTier", () => {
    it("should parse trust tier strings", () => {
      expect(parseTrustTier("highly_trusted")).toBe(TrustTier.HighlyTrusted);
      expect(parseTrustTier("highly-trusted")).toBe(TrustTier.HighlyTrusted);
      expect(parseTrustTier("verified")).toBe(TrustTier.Verified);
      expect(parseTrustTier("trusted")).toBe(TrustTier.Trusted);
      expect(parseTrustTier("untrusted")).toBe(TrustTier.Untrusted);
    });

    it("should return unknown for invalid tier", () => {
      expect(parseTrustTier("invalid")).toBe(TrustTier.Unknown);
      expect(parseTrustTier("")).toBe(TrustTier.Unknown);
      expect(parseTrustTier(undefined)).toBe(TrustTier.Unknown);
    });
  });

  describe("formatTrustScore", () => {
    it("should format trust scores correctly", () => {
      expect(formatTrustScore(95)).toBe("Excellent");
      expect(formatTrustScore(75)).toBe("Good");
      expect(formatTrustScore(55)).toBe("Fair");
      expect(formatTrustScore(35)).toBe("Low");
    });
  });

  describe("getTrustLevelColor", () => {
    it("should return correct colors", () => {
      expect(getTrustLevelColor(95)).toBe("#10b981"); // green
      expect(getTrustLevelColor(75)).toBe("#3b82f6"); // blue
      expect(getTrustLevelColor(55)).toBe("#f59e0b"); // amber
      expect(getTrustLevelColor(35)).toBe("#ef4444"); // red
    });
  });
});
