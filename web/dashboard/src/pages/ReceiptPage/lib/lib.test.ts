// Unit tests for the receipt lib utilities. These cover the schema
// renderer, OG meta builder, and brand icon registration — the pure
// functions that the page components depend on. Network and component
// rendering tests live alongside the components.

import { describe, expect, it } from "vitest";

import { encodeInputForUrl, buildShareText, buildOgTitle } from "../lib/og-meta";
import { flattenSchema, prettyJSON, truncatedPretty, truncate } from "../lib/schema-render";
import { getRuntimeStyle } from "../lib/runtime-badge";
import type { Receipt } from "../types";

const sampleReceipt: Receipt = {
  id: "V1StGXR8_Z5jHi3B-myT",
  function: {
    name: "summarize-url",
    author: "ada",
    runtime: "python3.11",
    version: "1.4.2",
    visibility: "public",
    description: "Fetches a URL and returns a 3-sentence summary.",
  },
  execution: {
    input: { url: "https://example.com" },
    output: { summary: "An example domain." },
    duration_ms: 142,
    cached: false,
    created_at: "2026-06-01T18:42:11Z",
  },
  share: {
    url: "https://functionfly.com/r/V1StGXR8_Z5jHi3B-myT",
    embed_url: "https://functionfly.com/r/V1StGXR8_Z5jHi3B-myT/embed",
    tweet_intent_url: "https://twitter.com/intent/tweet?text=hi",
    og_meta: { title: "t", description: "d", image: "i" },
  },
  can_run: true,
  is_paid: false,
  price_per_call_usd: 0,
};

describe("schema-render", () => {
  it("flattens a simple object schema", () => {
    const schema = {
      type: "object",
      properties: {
        url: { type: "string", description: "URL to summarize" },
        max: { type: "number", default: 3 },
      },
      required: ["url"],
    };
    const lines = flattenSchema(schema);
    expect(lines.length).toBe(3); // root + 2 properties
    const urlLine = lines.find((l) => l.key === "url");
    expect(urlLine).toBeDefined();
    expect(urlLine?.required).toBe(true);
    expect(urlLine?.type).toBe("string");
  });

  it("returns empty list for null schema", () => {
    expect(flattenSchema(null)).toEqual([]);
    expect(flattenSchema(undefined)).toEqual([]);
  });

  it("prettyJSON handles null and objects", () => {
    expect(prettyJSON(null)).toBe("null");
    expect(prettyJSON({ a: 1 })).toBe('{\n  "a": 1\n}');
  });

  it("truncatedPretty truncates with marker", () => {
    const big = { a: "x".repeat(10000) };
    const out = truncatedPretty(big, 100);
    expect(out.truncated).toBe(true);
    expect(out.text.length).toBe(100);
  });

  it("truncate adds ellipsis at length-1", () => {
    expect(truncate("hello world", 8)).toBe("hello w…");
    expect(truncate("hi", 10)).toBe("hi");
  });
});

describe("og-meta", () => {
  it("buildShareText uses description when present", () => {
    const text = buildShareText(sampleReceipt);
    expect(text).toContain("ada/summarize-url");
    expect(text).toContain("@functionfly");
    expect(text).toContain("Fetches a URL");
  });

  it("buildShareText milestone variant", () => {
    const text = buildShareText(sampleReceipt, { variant: "milestone", threshold: 100 });
    expect(text).toContain("🎉");
    expect(text).toContain("100");
  });

  it("buildOgTitle includes duration", () => {
    const title = buildOgTitle(sampleReceipt);
    expect(title).toContain("ada/summarize-url");
    expect(title).toContain("142ms");
  });

  it("encodeInputForUrl is URL-safe", () => {
    const encoded = encodeInputForUrl({ url: "https://example.com/?q=hi" });
    expect(encoded).toBeTruthy();
    // No '+', '/', or '=' in URL-safe base64.
    expect(encoded).not.toMatch(/[+/=]/);
  });
});

describe("runtime-badge", () => {
  it("returns Python badge for python3.11", () => {
    const s = getRuntimeStyle("python3.11");
    expect(s.label).toBe("Python 3.11");
    expect(s.className).toContain("blue");
  });

  it("falls back to default for unknown runtime", () => {
    const s = getRuntimeStyle("totally-unknown-runtime");
    expect(s.label).toBe("totally-unknown-runtime");
  });

  it("handles empty runtime", () => {
    const s = getRuntimeStyle("");
    expect(s.label).toBe("Function");
  });

  it("matches prefix for fuzzy runtime strings", () => {
    const s = getRuntimeStyle("python3.12-coldstart");
    // Should fuzzy-match to python3_12
    expect(s.label.toLowerCase()).toContain("python");
  });
});
