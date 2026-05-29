import type { FRGEdge, FRGNode, FunctionCatalogItem } from '@/types/frg';

export interface FrgTemplateDefinition {
  id: string;
  name: string;
  description: string;
  nodeCount: number;
  slots: Array<{
    id: string;
    label: string;
    categories: string[];
    x: number;
    y: number;
  }>;
  edges: Array<{ source: string; target: string }>;
}

export const FRG_TEMPLATES: FrgTemplateDefinition[] = [
  {
    id: 'webhook-api',
    name: 'Webhook → Function → Response',
    description: 'API endpoint with validation and response',
    nodeCount: 3,
    slots: [
      { id: 'input', label: 'HTTP Input', categories: ['api', 'default'], x: 80, y: 160 },
      { id: 'process', label: 'Process', categories: ['code', 'data', 'default'], x: 320, y: 160 },
      { id: 'output', label: 'Response', categories: ['api', 'default'], x: 560, y: 160 },
    ],
    edges: [
      { source: 'input', target: 'process' },
      { source: 'process', target: 'output' },
    ],
  },
  {
    id: 'ai-pipeline',
    name: 'AI Pipeline',
    description: 'Input → model → output processing chain',
    nodeCount: 3,
    slots: [
      { id: 'input', label: 'User Input', categories: ['api', 'text', 'default'], x: 80, y: 160 },
      { id: 'model', label: 'AI Model', categories: ['ml', 'text', 'default'], x: 320, y: 160 },
      { id: 'output', label: 'Output', categories: ['text', 'data', 'default'], x: 560, y: 160 },
    ],
    edges: [
      { source: 'input', target: 'model' },
      { source: 'model', target: 'output' },
    ],
  },
  {
    id: 'data-processing',
    name: 'Data Processing',
    description: 'Upload → transform → store workflow',
    nodeCount: 3,
    slots: [
      { id: 'upload', label: 'Upload', categories: ['api', 'data', 'default'], x: 80, y: 160 },
      {
        id: 'transform',
        label: 'Transform',
        categories: ['data', 'code', 'default'],
        x: 320,
        y: 160,
      },
      { id: 'store', label: 'Store', categories: ['data', 'default'], x: 560, y: 160 },
    ],
    edges: [
      { source: 'upload', target: 'transform' },
      { source: 'transform', target: 'store' },
    ],
  },
  {
    id: 'scheduled-job',
    name: 'Scheduled Job',
    description: 'Timer → function → notification workflow',
    nodeCount: 3,
    slots: [
      { id: 'timer', label: 'Schedule', categories: ['code', 'default'], x: 80, y: 160 },
      { id: 'execute', label: 'Execute', categories: ['code', 'data', 'default'], x: 320, y: 160 },
      { id: 'notify', label: 'Notify', categories: ['api', 'text', 'default'], x: 560, y: 160 },
    ],
    edges: [
      { source: 'timer', target: 'execute' },
      { source: 'execute', target: 'notify' },
    ],
  },
  {
    id: 'etl-workflow',
    name: 'ETL Workflow',
    description: 'Extract, transform, load pipeline',
    nodeCount: 3,
    slots: [
      { id: 'extract', label: 'Extract', categories: ['data', 'api', 'default'], x: 80, y: 160 },
      {
        id: 'transform',
        label: 'Transform',
        categories: ['data', 'code', 'default'],
        x: 320,
        y: 160,
      },
      { id: 'load', label: 'Load', categories: ['data', 'default'], x: 560, y: 160 },
    ],
    edges: [
      { source: 'extract', target: 'transform' },
      { source: 'transform', target: 'load' },
    ],
  },
];

export type ApplyFrgTemplateResult =
  | { ok: true; nodes: FRGNode[]; edges: FRGEdge[] }
  | { ok: false; reason: 'unknown_template' | 'insufficient_functions'; message: string };

function pickLibraryFunction(
  categories: string[],
  library: FunctionCatalogItem[],
  used: Set<string>
): FunctionCatalogItem | undefined {
  for (const category of categories) {
    const match = library.find((fn) => fn.category === category && !used.has(fn.id));
    if (match) return match;
  }
  return library.find((fn) => !used.has(fn.id));
}

function createNodeFromFunction(
  slot: FrgTemplateDefinition['slots'][number],
  fn: FunctionCatalogItem,
  nodeId: string
): FRGNode {
  return {
    id: nodeId,
    type: 'functionNode',
    position: { x: slot.x, y: slot.y },
    data: {
      functionRef: {
        nodeId,
        author: fn.author,
        name: fn.name,
        version: fn.version || 'latest',
        config: {},
        metadata: {
          label: slot.label,
          category: fn.category,
          inputSchema: fn.inputSchema,
          outputSchema: fn.outputSchema,
        },
      },
      isSelected: false,
      isEditable: true,
    },
  };
}

export function applyFrgTemplate(
  templateId: string,
  library: FunctionCatalogItem[] = []
): ApplyFrgTemplateResult {
  const template = FRG_TEMPLATES.find((t) => t.id === templateId);
  if (!template) {
    return { ok: false, reason: 'unknown_template', message: 'Template not found' };
  }

  if (library.length < template.slots.length) {
    return {
      ok: false,
      reason: 'insufficient_functions',
      message: `Add at least ${template.slots.length} functions to your library before using this template`,
    };
  }

  const used = new Set<string>();
  const slotToNodeId: Record<string, string> = {};
  const nodes: FRGNode[] = [];

  for (const slot of template.slots) {
    const fn = pickLibraryFunction(slot.categories, library, used);
    if (!fn) {
      return {
        ok: false,
        reason: 'insufficient_functions',
        message: `Could not resolve all template slots. Publish functions in categories: ${slot.categories.join(', ')}`,
      };
    }
    used.add(fn.id);
    const nodeId = `node-${slot.id}-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;
    slotToNodeId[slot.id] = nodeId;
    nodes.push(createNodeFromFunction(slot, fn, nodeId));
  }

  const edges: FRGEdge[] = template.edges.map((edge, index) => ({
    id: `e-${slotToNodeId[edge.source]}-${slotToNodeId[edge.target]}-${index}`,
    source: slotToNodeId[edge.source],
    target: slotToNodeId[edge.target],
    type: 'custom',
    data: {
      mapping: { sourcePath: '*', targetPath: '*' },
      isValid: true,
      runtimeState: {
        status: 'idle',
        recordsTransferred: 0,
        bytesTransferred: 0,
        isDataFlowing: false,
        flowProgress: 0,
      },
    },
  }));

  return { ok: true, nodes, edges };
}

export function getFrgTemplate(templateId: string): FrgTemplateDefinition | undefined {
  return FRG_TEMPLATES.find((t) => t.id === templateId);
}
