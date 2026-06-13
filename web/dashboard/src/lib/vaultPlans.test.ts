import { describe, expect, it } from "vitest";
import {
  FEATURE_MIN_PLAN,
  PLAN_LIMITS,
  getPlanLimits,
  hasFeature,
  isFeatureAvailable,
  minPlanForFeature,
  PLAN_META,
} from "@/lib/vaultPlans";

describe("vaultPlans", () => {
  describe("getPlanLimits", () => {
    it("returns the canonical plan matrix for known plans", () => {
      expect(getPlanLimits("free").maxSecrets).toBe(25);
      expect(getPlanLimits("pro").maxSecrets).toBe(500);
      expect(getPlanLimits("team").maxSecrets).toBe(5_000);
      expect(getPlanLimits("enterprise").maxSecrets).toBe(1_000_000);
    });

    it("falls back to free for unknown plans", () => {
      const lim = getPlanLimits("unknown" as never);
      expect(lim.maxSecrets).toBe(25);
    });

    it("higher plans have more dynamic-cred capacity", () => {
      const tiers = ["free", "pro", "team", "enterprise"] as const;
      let prev = -1;
      for (const t of tiers) {
        expect(getPlanLimits(t).maxDynamicCreds).toBeGreaterThan(prev);
        prev = getPlanLimits(t).maxDynamicCreds;
      }
    });
  });

  describe("hasFeature / isFeatureAvailable", () => {
    it("gates SSO behind enterprise", () => {
      expect(hasFeature("free", "sso")).toBe(false);
      expect(hasFeature("pro", "sso")).toBe(false);
      expect(hasFeature("team", "sso")).toBe(false);
      expect(hasFeature("enterprise", "sso")).toBe(true);
    });

    it("gates MFA behind pro", () => {
      expect(hasFeature("free", "mfa")).toBe(false);
      expect(hasFeature("pro", "mfa")).toBe(true);
    });

    it("gates RBAC behind team", () => {
      expect(hasFeature("pro", "rbac")).toBe(false);
      expect(hasFeature("team", "rbac")).toBe(true);
    });

    it("makes namespaces available to all plans", () => {
      for (const p of ["free", "pro", "team", "enterprise"] as const) {
        expect(hasFeature(p, "namespaces")).toBe(true);
      }
    });

    it("makes the quota widget available to all plans", () => {
      for (const p of ["free", "pro", "team", "enterprise"] as const) {
        expect(hasFeature(p, "quotaWidget")).toBe(true);
      }
    });

    it("isFeatureAvailable mirrors hasFeature", () => {
      expect(isFeatureAvailable("pro", "mfa")).toBe(hasFeature("pro", "mfa"));
      expect(isFeatureAvailable("free", "mfa")).toBe(hasFeature("free", "mfa"));
    });
  });

  describe("minPlanForFeature", () => {
    it("returns the lowest plan that unlocks each feature", () => {
      expect(minPlanForFeature("mfa")).toBe("pro");
      expect(minPlanForFeature("ipAllowlist")).toBe("pro");
      expect(minPlanForFeature("breakGlass")).toBe("pro");
      expect(minPlanForFeature("auditExport")).toBe("pro");
      expect(minPlanForFeature("cacheStats")).toBe("pro");
      expect(minPlanForFeature("tokenMonitor")).toBe("pro");

      expect(minPlanForFeature("rbac")).toBe("team");
      expect(minPlanForFeature("escrow")).toBe("team");
      expect(minPlanForFeature("shares")).toBe("team");
      expect(minPlanForFeature("siemWebhooks")).toBe("team");
      expect(minPlanForFeature("dependencyGraph")).toBe("team");

      expect(minPlanForFeature("sso")).toBe("enterprise");
      expect(minPlanForFeature("haStatus")).toBe("enterprise");

      expect(minPlanForFeature("namespaces")).toBe("free");
      expect(minPlanForFeature("expiration")).toBe("free");
      expect(minPlanForFeature("expirationDashboard")).toBe("free");
      expect(minPlanForFeature("quotaWidget")).toBe("free");
    });

    it("the matrix is internally consistent (every feature has a min plan)", () => {
      const matrix = PLAN_LIMITS.free.features;
      for (const feature of Object.keys(matrix) as Array<keyof typeof matrix>) {
        // The matrix and FEATURE_MIN_PLAN should agree on which
        // plans include a feature.
        const min = FEATURE_MIN_PLAN[feature];
        expect(hasFeature(min, feature)).toBe(true);
      }
    });
  });

  describe("PLAN_META", () => {
    it("every plan has a name, price, and tagline", () => {
      for (const p of ["free", "pro", "team", "enterprise"] as const) {
        expect(PLAN_META[p].name).toBeTruthy();
        expect(PLAN_META[p].price).toBeTruthy();
        expect(PLAN_META[p].tagline).toBeTruthy();
      }
    });
  });
});
