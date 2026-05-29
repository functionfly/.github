import type { FRGEdge, FRGNode, GraphDefinition } from '@/types/frg';

const MAX_IMPORT_BYTES = 1_000_000;
const MAX_NODES = 100;
const MAX_EDGES = 500;

const SECRET_KEYS = [
  'apiKey',
  'api_key',
  'apiSecret',
  'api_secret',
  'password',
  'secret',
  'token',
  'privateKey',
  'private_key',
];

export interface FrgExportPayload {
  format: 'functionfly-frg';
  version: 1;
  exportedAt: string;
  graph: {
    name: string;
    executionMode?: GraphDefinition['executionMode'];
    visibility?: GraphDefinition['visibility'];
    nodeRefs: Array<{
      nodeId: string;
      author: string;
      name: string;
      version: string;
      config: Record<string, unknown>;
      metadata: Record<string, unknown>;
    }>;
    edges: Array<{
      id: string;
      sourceNodeId: string;
      targetNodeId: string;
      mapping: { sourcePath?: string; targetPath?: string };
    }>;
  };
}

function redactSecrets(obj: Record<string, unknown>): Record<string, unknown> {
  const out: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(obj)) {
    if (SECRET_KEYS.some((sk) => key.toLowerCase().includes(sk.toLowerCase()))) {
      out[key] = '[REDACTED]';
    } else {
      out[key] = value;
    }
  }
  return out;
}

export function buildFrgExportPayload(
  graphName: string,
  nodes: FRGNode[],
  edges: FRGEdge[],
  definition: Partial<GraphDefinition> | null
): FrgExportPayload {
  return {
    format: 'functionfly-frg',
    version: 1,
    exportedAt: new Date().toISOString(),
    graph: {
      name: graphName,
      executionMode: definition?.executionMode,
      visibility: definition?.visibility,
      nodeRefs: nodes.map((node) => ({
        nodeId: node.id,
        author: node.data.functionRef?.author ?? '',
        name: node.data.functionRef?.name ?? '',
        version: node.data.functionRef?.version ?? 'latest',
        config: redactSecrets(node.data.functionRef?.config ?? {}),
        metadata: node.data.functionRef?.metadata ?? {},
      })),
      edges: edges.map((edge) => ({
        id: edge.id,
        sourceNodeId: edge.source,
        targetNodeId: edge.target,
        mapping: edge.data?.mapping ?? { sourcePath: '*', targetPath: '*' },
      })),
    },
  };
}

export function downloadFrgExport(payload: FrgExportPayload, filename: string): void {
  const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = filename.endsWith('.json') ? filename : `${filename}.json`;
  anchor.click();
  URL.revokeObjectURL(url);
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

export function parseFrgImportFile(raw: string): {
  nodes: FRGNode[];
  edges: FRGEdge[];
  name: string;
} {
  if (raw.length > MAX_IMPORT_BYTES) {
    throw new Error('Import file exceeds 1MB limit');
  }

  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    throw new Error('Invalid JSON file');
  }

  if (!isRecord(parsed)) {
    throw new Error('Invalid graph export format');
  }

  const graph = isRecord(parsed.graph) ? parsed.graph : parsed;
  const nodeRefs = Array.isArray(graph.nodeRefs)
    ? graph.nodeRefs
    : Array.isArray(graph.nodes)
      ? graph.nodes
      : null;
  const edgeDefs = Array.isArray(graph.edges) ? graph.edges : null;

  if (!nodeRefs || !edgeDefs) {
    throw new Error('Graph must include nodes and edges');
  }
  if (nodeRefs.length > MAX_NODES) {
    throw new Error(`Graph cannot have more than ${MAX_NODES} nodes`);
  }
  if (edgeDefs.length > MAX_EDGES) {
    throw new Error(`Graph cannot have more than ${MAX_EDGES} edges`);
  }

  const name =
    typeof graph.name === 'string' && graph.name.trim() ? graph.name.trim() : 'imported-graph';

  const nodes: FRGNode[] = nodeRefs.map((ref, index) => {
    if (!isRecord(ref)) throw new Error(`Invalid node at index ${index}`);
    const nodeId = String(ref.nodeId ?? ref.node_id ?? `node-import-${index}`);
    return {
      id: nodeId,
      type: 'functionNode',
      position: {
        x: 120 + (index % 4) * 180,
        y: 120 + Math.floor(index / 4) * 140,
      },
      data: {
        functionRef: {
          nodeId,
          author: String(ref.author ?? ''),
          name: String(ref.name ?? 'unnamed'),
          version: String(ref.version ?? 'latest'),
          config: isRecord(ref.config) ? ref.config : {},
          metadata: isRecord(ref.metadata) ? ref.metadata : {},
        },
        isSelected: false,
        isEditable: true,
      },
    };
  });

  const nodeIds = new Set(nodes.map((n) => n.id));
  const edges: FRGEdge[] = edgeDefs.map((edge, index) => {
    if (!isRecord(edge)) throw new Error(`Invalid edge at index ${index}`);
    const source = String(edge.sourceNodeId ?? edge.source_node_id ?? edge.source ?? '');
    const target = String(edge.targetNodeId ?? edge.target_node_id ?? edge.target ?? '');
    if (!nodeIds.has(source) || !nodeIds.has(target)) {
      throw new Error(`Edge references unknown node at index ${index}`);
    }
    return {
      id: String(edge.id ?? `e-import-${index}`),
      source,
      target,
      type: 'custom',
      data: {
        mapping: isRecord(edge.mapping)
          ? {
              sourcePath:
                typeof edge.mapping.sourcePath === 'string' ? edge.mapping.sourcePath : '*',
              targetPath:
                typeof edge.mapping.targetPath === 'string' ? edge.mapping.targetPath : '*',
            }
          : { sourcePath: '*', targetPath: '*' },
        isValid: true,
        runtimeState: {
          status: 'idle',
          recordsTransferred: 0,
          bytesTransferred: 0,
          isDataFlowing: false,
          flowProgress: 0,
        },
      },
    };
  });

  return { nodes, edges, name };
}
