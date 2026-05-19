/**
 * Unit tests for FunctionFly Agent SDK
 */

import { describe, expect, it } from "vitest";
import { AgentClient } from "./client.js";
import { TrustTier } from "./types.js";

describe("AgentClient", () => {
  const mockApiKey = "test-api-key";
  const mockBaseUrl = "https://test.functionfly.com";

  describe("constructor", () => {
    it("should create client with default base URL", () => {
      const client = new AgentClient({ apiKey: mockApiKey });
      expect(client).toBeDefined();
    });

    it("should create client with custom base URL", () => {
      const client = new AgentClient({
        apiKey: mockApiKey,
        baseUrl: mockBaseUrl,
      });
      expect(client).toBeDefined();
    });

    it("should create client with custom timeout", () => {
      const client = new AgentClient({
        apiKey: mockApiKey,
        timeout: 60000,
      });
      expect(client).toBeDefined();
    });
  });
});

describe("Trust Tier", () => {
  it("should have correct trust tier values", () => {
    expect(TrustTier.HighlyTrusted).toBe("highly_trusted");
    expect(TrustTier.Verified).toBe("verified");
    expect(TrustTier.Trusted).toBe("trusted");
    expect(TrustTier.Untrusted).toBe("untrusted");
    expect(TrustTier.Unknown).toBe("unknown");
  });
});

describe("Type transformations", () => {
  it("should export all required types", async () => {
    const { types } = await import("./types.js");

    expect(types.FunctionSearchOptions).toBeDefined();
    expect(types.ExecuteRequest).toBeDefined();
    expect(types.ExecuteResponse).toBeDefined();
    expect(types.TrustScoreResponse).toBeDefined();
  });
});
