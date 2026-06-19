import { describe, expect, it } from 'vitest';

import type { FunctionCatalogItem } from '@/types/frg';
import { buildFrgExportPayload, parseFrgImportFile } from '@/utils/frgImportExport';
import { countSchemaPorts, listSchemaPropertyNames } from '@/utils/frgSchema';
import { schemaFromManifestField } from '@/utils/frgRegistrySchemas';
import { nextStepNodeId, topologicalNodeOrder } from '@/utils/frgExecutionOrder';
import { applyFrgTemplate, FRG_TEMPLATES } from '@/utils/frgTemplates';
import {
  isFunctionRefConfigured,
  validateGraphName,
  validateRouteSegment,
} from '@/utils/frgValidation';

const mockLibrary: FunctionCatalogItem[] = [
  {
    id: 'fn-1',
    author: 'alice',
    name: 'echo',
    version: '1.0.0',
    description: 'Echo input',
    category: 'api',
    tags: [],
    inputSchema: { type: 'object', properties: { message: { type: 'string' } } },
    outputSchema: { type: 'object', properties: { result: { type: 'string' } } },
    trustScore: 4.5,
    usageCount: 10,
    avgExecutionTimeMs: 50,
  },
  {
    id: 'fn-2',
    author: 'alice',
    name: 'transform',
    version: '1.0.0',
    description: 'Transform data',
    category: 'data',
    tags: [],
    inputSchema: { type: 'object', properties: { data: { type: 'object' } } },
    outputSchema: { type: 'object', properties: { data: { type: 'object' } } },
    trustScore: 4.2,
    usageCount: 5,
    avgExecutionTimeMs: 80,
  },
  {
    id: 'fn-3',
    author: 'alice',
    name: 'respond',
    version: '1.0.0',
    description: 'Send response',
    category: 'text',
    tags: [],
    inputSchema: { type: 'object', properties: { body: { type: 'string' } } },
    outputSchema: { type: 'object', properties: { status: { type: 'number' } } },
    trustScore: 4.0,
    usageCount: 3,
    avgExecutionTimeMs: 30,
  },
];

describe('frgValidation', () => {
  it('accepts valid graph names', () => {
    expect(validateGraphName('my-graph_1')).toBeNull();
  });

  it('rejects invalid graph names', () => {
    expect(validateGraphName('')).toMatch(/required/i);
    expect(validateGraphName('1bad')).toMatch(/letter/i);
  });

  it('validates route segments', () => {
    expect(validateRouteSegment('author/name')).toBe(false);
    expect(validateRouteSegment('my-graph')).toBe(true);
  });

  it('requires author and name for configured refs', () => {
    expect(isFunctionRefConfigured({ author: 'alice', name: 'echo' })).toBe(true);
    expect(isFunctionRefConfigured({ author: '', name: 'echo' })).toBe(false);
  });
});

describe('frgSchema', () => {
  it('counts schema ports', () => {
    expect(countSchemaPorts({ type: 'object', properties: { a: {}, b: {} } })).toBe(2);
    expect(countSchemaPorts(undefined)).toBe(1);
  });

  it('lists schema property names', () => {
    expect(listSchemaPropertyNames({ type: 'object', properties: { foo: {}, bar: {} } })).toEqual([
      'foo',
      'bar',
    ]);
  });
});

describe('frgRegistrySchemas', () => {
  it('extracts schemas from manifest fields', () => {
    const input = schemaFromManifestField(
      {
        input_schema: { type: 'object', properties: { message: { type: 'string' } } },
      },
      'input_schema',
      'input'
    );
    expect(Object.keys((input.properties as Record<string, unknown>) ?? {})).toContain('message');
  });
});

describe('frgExecutionOrder', () => {
  it('orders nodes topologically', () => {
    const nodes = [
      { id: 'a', type: 'functionNode', position: { x: 0, y: 0 }, data: {} as any },
      { id: 'b', type: 'functionNode', position: { x: 0, y: 0 }, data: {} as any },
    ];
    const edges = [
      {
        id: 'e1',
        source: 'a',
        target: 'b',
        type: 'custom',
        data: { mapping: { sourcePath: '*', targetPath: '*' }, isValid: true },
      },
    ];
    expect(topologicalNodeOrder(nodes, edges)).toEqual(['a', 'b']);
    expect(nextStepNodeId(nodes, edges, {})).toBe('a');
  });
});

describe('frgTemplates', () => {
  it('applies known templates with real library functions', () => {
    const result = applyFrgTemplate('ai-pipeline', mockLibrary);
    expect(result.ok).toBe(true);
    if (result.ok) {
      expect(result.nodes).toHaveLength(3);
      expect(result.edges).toHaveLength(2);
      expect(result.nodes.every((node) => node.data.functionRef.author)).toBe(true);
    }
  });

  it('rejects templates when the library is too small', () => {
    const result = applyFrgTemplate('ai-pipeline', mockLibrary.slice(0, 1));
    expect(result.ok).toBe(false);
    if (!result.ok && 'reason' in result) {
      expect(result.reason).toBe('insufficient_functions');
    }
  });

  it('returns unknown template errors', () => {
    const result = applyFrgTemplate('missing', mockLibrary);
    expect(result.ok).toBe(false);
    if (!result.ok && 'reason' in result) {
      expect(result.reason).toBe('unknown_template');
    }
  });

  it('keeps template ids unique', () => {
    const ids = FRG_TEMPLATES.map((t) => t.id);
    expect(new Set(ids).size).toBe(ids.length);
  });
});

describe('frgImportExport', () => {
  it('round-trips export payload through import parser', () => {
    const payload = buildFrgExportPayload(
      'demo-graph',
      [
        {
          id: 'node-1',
          type: 'functionNode',
          position: { x: 0, y: 0 },
          data: {
            functionRef: {
              nodeId: 'node-1',
              author: 'alice',
              name: 'echo',
              version: '1.0.0',
              config: { apiKey: 'secret' },
              metadata: {},
            },
            isSelected: false,
            isEditable: true,
          },
        },
      ],
      [
        {
          id: 'edge-1',
          source: 'node-1',
          target: 'node-1',
          type: 'custom',
          data: { mapping: { sourcePath: '*', targetPath: '*' }, isValid: true },
        },
      ],
      null
    );

    const imported = parseFrgImportFile(JSON.stringify(payload));
    expect(imported.name).toBe('demo-graph');
    expect(imported.nodes).toHaveLength(1);
    expect(imported.nodes[0].data.functionRef.config.apiKey).toBe('[REDACTED]');
  });

  it('rejects oversized imports', () => {
    expect(() => parseFrgImportFile('{"nodes":[],"edges":[]}'.padEnd(1_000_001, ' '))).toThrow(
      /1MB/i
    );
  });
});
