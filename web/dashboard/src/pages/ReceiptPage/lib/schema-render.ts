// Tiny JSON-Schema → human readable renderer.
//
// We intentionally do NOT pull in a heavy schema-rendering dependency.
// The receipt page is the top of the marketing funnel and must hydrate
// fast; we render the schema as a tree of type + required + description
// labels and fall back to a pretty-printed JSON view for unknown shapes.

import type { ReactNode } from "react";

export interface SchemaNode {
  type: "object" | "array" | "string" | "number" | "boolean" | "null" | "any";
  description?: string;
  required?: boolean;
  default?: unknown;
  properties?: Record<string, SchemaNode>;
  items?: SchemaNode;
  enum?: unknown[];
  example?: unknown;
  format?: string;
  /** Names of child `properties` that the JSON Schema marks as required. */
  requiredFields?: string[];
}

export interface RenderedSchemaLine {
  key: string;
  type: string;
  required: boolean;
  description?: string;
  example?: string;
  depth: number;
  isLeaf: boolean;
}

const TYPE_FALLBACK = "any";

function detectType(value: unknown): string {
  if (value === null) return "null";
  if (Array.isArray(value)) return "array";
  if (typeof value === "object") return "object";
  if (typeof value === "string") return "string";
  if (typeof value === "number") return "number";
  if (typeof value === "boolean") return "boolean";
  return TYPE_FALLBACK;
}

function normalize(node: unknown, required = false): SchemaNode | null {
  if (!node) return null;
  if (typeof node !== "object") return null;
  const obj = node as Record<string, unknown>;
  const t = (obj.type as string) ?? detectType(obj);
  const requiredFields = Array.isArray(obj.required) ? (obj.required as string[]) : undefined;
  return {
    type: (t as SchemaNode["type"]) ?? TYPE_FALLBACK,
    description: obj.description as string | undefined,
    required,
    default: obj.default,
    properties: obj.properties as Record<string, SchemaNode> | undefined,
    items: obj.items as SchemaNode | undefined,
    enum: obj.enum as unknown[] | undefined,
    example: obj.example,
    format: obj.format as string | undefined,
    requiredFields,
  };
}

function walk(
  node: SchemaNode,
  lines: RenderedSchemaLine[],
  depth: number,
  key: string,
  parentRequired: string[] = [],
): void {
  const isLeaf =
    node.type !== "object" &&
    node.type !== "array" &&
    !node.properties &&
    !node.items;
  const required = parentRequired.includes(key) || !!node.required;
  lines.push({
    key,
    type: node.type + (node.format ? `<${node.format}>` : ""),
    required,
    description: node.description,
    example: node.example !== undefined ? JSON.stringify(node.example) : undefined,
    depth,
    isLeaf,
  });
  if (node.properties) {
    // The "required" array of the current node lists the child keys
    // that are required. We preserved it on the SchemaNode during
    // normalize so the walker doesn't have to reach back into the raw
    // schema object.
    const childRequired = node.requiredFields ?? [];
    const requiredSet = new Set(childRequired);
    for (const [childKey, childNode] of Object.entries(node.properties)) {
      const child = normalize(childNode, requiredSet.has(childKey));
      if (child) walk(child, lines, depth + 1, childKey, childRequired);
    }
  }
  if (node.items) {
    const item = normalize(node.items);
    if (item) walk(item, lines, depth + 1, "[]", parentRequired);
  }
}

/**
 * Flatten a JSON Schema into a list of tree rows. The caller renders these
 * with appropriate indentation.
 */
export function flattenSchema(schema: unknown): RenderedSchemaLine[] {
  const root = normalize(schema);
  if (!root) return [];
  const lines: RenderedSchemaLine[] = [];
  walk(root, lines, 0, '');
  return lines;
}

/**
 * Pretty-print JSON for the input/output panels. Falls back to
 * `String(value)` if JSON.stringify throws (e.g. circular refs).
 */
export function prettyJSON(value: unknown, indent = 2): string {
  try {
    return JSON.stringify(value, null, indent);
  } catch {
    return String(value);
  }
}

/**
 * Truncate a string to N characters with an ellipsis.
 */
export function truncate(s: string, n: number): string {
  if (s.length <= n) return s;
  return s.slice(0, Math.max(0, n - 1)) + "…";
}

/**
 * Build a "view" of JSON that the UI can render: trims large arrays, marks
 * truncation so the user can click "show more".
 */
export function truncatedPretty(value: unknown, maxBytes = 8 * 1024): { text: string; truncated: boolean } {
  const full = prettyJSON(value, 2);
  if (full.length <= maxBytes) return { text: full, truncated: false };
  return { text: full.slice(0, maxBytes), truncated: true };
}

/**
 * Render a "Powered by" badge as a linkable React node. Kept here so the
 * visual style is consistent between the receipt page and the embed.
 */
export function poweredByNode(opts?: { href?: string }): ReactNode {
  return null; // rendered by ReceiptPoweredBy component to keep JSX out of lib
}
