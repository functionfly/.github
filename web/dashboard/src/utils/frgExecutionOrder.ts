/** Topological ordering for FRG debug step execution. */

import type { FRGEdge, FRGNode } from '@/types/frg';

export function topologicalNodeOrder(nodes: FRGNode[], edges: FRGEdge[]): string[] {
  const nodeIds = nodes.map((n) => n.id);
  const indegree = new Map<string, number>();
  const adjacency = new Map<string, string[]>();

  for (const id of nodeIds) {
    indegree.set(id, 0);
    adjacency.set(id, []);
  }

  for (const edge of edges) {
    if (!indegree.has(edge.source) || !indegree.has(edge.target)) continue;
    adjacency.get(edge.source)?.push(edge.target);
    indegree.set(edge.target, (indegree.get(edge.target) ?? 0) + 1);
  }

  const queue = nodeIds.filter((id) => (indegree.get(id) ?? 0) === 0);
  const ordered: string[] = [];

  while (queue.length > 0) {
    const current = queue.shift();
    if (!current) break;
    ordered.push(current);
    for (const next of adjacency.get(current) ?? []) {
      const nextDegree = (indegree.get(next) ?? 0) - 1;
      indegree.set(next, nextDegree);
      if (nextDegree === 0) queue.push(next);
    }
  }

  if (ordered.length < nodeIds.length) {
    return nodeIds;
  }

  return ordered;
}

export function nextStepNodeId(
  nodes: FRGNode[],
  edges: FRGEdge[],
  nodeRuntimeStates: Record<string, { status?: string }>
): string | null {
  const order = topologicalNodeOrder(nodes, edges);
  return (
    order.find((nodeId) => {
      const status = nodeRuntimeStates[nodeId]?.status ?? 'idle';
      return status === 'idle' || status === 'pending';
    }) ?? null
  );
}
