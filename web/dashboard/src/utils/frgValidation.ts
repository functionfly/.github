/** Validation helpers for FRG graph names and import payloads. */

import type { FRGNode } from '@/types/frg';

const GRAPH_NAME_PATTERN = /^[a-zA-Z][a-zA-Z0-9_-]{0,99}$/;

export function validateGraphName(name: string): string | null {
  const trimmed = name.trim();
  if (!trimmed) return 'Graph name is required';
  if (trimmed.length > 100) return 'Graph name must be 100 characters or fewer';
  if (!GRAPH_NAME_PATTERN.test(trimmed)) {
    return 'Graph name must start with a letter and contain only letters, numbers, hyphens, and underscores';
  }
  return null;
}

export function validateRouteSegment(segment: string): boolean {
  return /^[a-zA-Z0-9_-]+$/.test(segment) && segment.length <= 100;
}

export function sanitizeGraphName(name: string): string {
  const trimmed = name.trim().replace(/\s+/g, '-');
  const sanitized = trimmed.replace(/[^a-zA-Z0-9_-]/g, '');
  if (!sanitized || !/^[a-zA-Z]/.test(sanitized)) {
    return 'untitled-graph';
  }
  return sanitized.slice(0, 100);
}

export function isFunctionRefConfigured(ref?: { author?: string; name?: string }): boolean {
  return Boolean(ref?.author?.trim() && ref?.name?.trim());
}

export function getUnconfiguredNodes(nodes: FRGNode[]): FRGNode[] {
  return nodes.filter((node) => !isFunctionRefConfigured(node.data?.functionRef));
}

export function formatUnconfiguredNodesMessage(nodes: FRGNode[]): string {
  const labels = getUnconfiguredNodes(nodes).map(
    (node) =>
      (node.data?.functionRef?.metadata?.label as string | undefined) ||
      node.data?.functionRef?.name ||
      node.id
  );
  if (labels.length === 0) {
    return 'Each node must reference a function from the library before saving';
  }
  return `Assign functions to: ${labels.join(', ')}`;
}
