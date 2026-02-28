/**
 * Schema traversal and helper utilities for JSON Schema.
 */

export interface SchemaNode {
  key: string;
  path: string;
  type: string | string[];
  title?: string;
  description?: string;
  required: boolean;
  example?: unknown;
  enum?: unknown[];
  properties?: SchemaNode[];
  items?: SchemaNode;
  default?: unknown;
}

export function flattenSchema(
  schema: Record<string, unknown>,
  path = '',
  requiredFields: string[] = []
): SchemaNode[] {
  const nodes: SchemaNode[] = [];

  if (!schema || typeof schema !== 'object') return nodes;

  const properties = schema.properties as Record<string, Record<string, unknown>> | undefined;
  if (!properties) return nodes;

  const required = (schema.required as string[]) || requiredFields;

  for (const [key, propSchema] of Object.entries(properties)) {
    const childPath = path ? `${path}.${key}` : key;
    const isRequired = required.includes(key);

    const node: SchemaNode = {
      key,
      path: childPath,
      type: (propSchema.type as string | string[]) || 'unknown',
      title: propSchema.title as string | undefined,
      description: propSchema.description as string | undefined,
      required: isRequired,
      example: propSchema.example,
      enum: propSchema.enum as unknown[] | undefined,
      default: propSchema.default,
    };

    if (propSchema.type === 'object' && propSchema.properties) {
      node.properties = flattenSchema(
        propSchema as Record<string, unknown>,
        childPath,
        (propSchema.required as string[]) || []
      );
    }

    if (propSchema.type === 'array' && propSchema.items) {
      const itemSchema = propSchema.items as Record<string, unknown>;
      node.items = {
        key: 'items',
        path: `${childPath}[]`,
        type: (itemSchema.type as string) || 'unknown',
        title: itemSchema.title as string | undefined,
        description: itemSchema.description as string | undefined,
        required: false,
        example: itemSchema.example,
      };
    }

    nodes.push(node);
  }

  return nodes;
}

export function getTypeColor(type: string | string[]): string {
  const t = Array.isArray(type) ? type[0] : type;
  switch (t) {
    case 'string':
      return 'text-blue-400';
    case 'number':
    case 'integer':
      return 'text-green-400';
    case 'boolean':
      return 'text-yellow-400';
    case 'object':
      return 'text-purple-400';
    case 'array':
      return 'text-orange-400';
    case 'null':
      return 'text-gray-400';
    default:
      return 'text-text-secondary';
  }
}

export function getTypeBadgeColor(type: string | string[]): string {
  const t = Array.isArray(type) ? type[0] : type;
  switch (t) {
    case 'string':
      return 'bg-blue-500/10 text-blue-400 border-blue-500/20';
    case 'number':
    case 'integer':
      return 'bg-green-500/10 text-green-400 border-green-500/20';
    case 'boolean':
      return 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20';
    case 'object':
      return 'bg-purple-500/10 text-purple-400 border-purple-500/20';
    case 'array':
      return 'bg-orange-500/10 text-orange-400 border-orange-500/20';
    default:
      return 'bg-gray-500/10 text-gray-400 border-gray-500/20';
  }
}

export function getValueType(value: unknown): string {
  if (value === null) return 'null';
  if (Array.isArray(value)) return 'array';
  return typeof value;
}

export function formatValue(value: unknown): string {
  if (value === null) return 'null';
  if (value === undefined) return 'undefined';
  if (typeof value === 'string') return `"${value}"`;
  if (typeof value === 'boolean') return value.toString();
  if (typeof value === 'number') return value.toString();
  if (Array.isArray(value)) return `[${value.length} items]`;
  if (typeof value === 'object') return `{${Object.keys(value as object).length} keys}`;
  return String(value);
}

export function getSchemaExamples(
  schema: Record<string, unknown>
): Array<{ name: string; input: unknown; description?: string }> {
  // Check for top-level examples array
  if (Array.isArray(schema.examples)) {
    return schema.examples as Array<{ name: string; input: unknown; description?: string }>;
  }

  // Build example from schema properties
  const example = buildExampleFromSchema(schema);
  if (example && Object.keys(example as object).length > 0) {
    return [{ name: 'Default Example', input: example }];
  }

  return [];
}

export function buildExampleFromSchema(schema: Record<string, unknown>): unknown {
  if (!schema) return null;

  if (schema.example !== undefined) return schema.example;

  const type = schema.type as string;

  switch (type) {
    case 'object': {
      const props = schema.properties as Record<string, Record<string, unknown>> | undefined;
      if (!props) return {};
      const result: Record<string, unknown> = {};
      for (const [key, propSchema] of Object.entries(props)) {
        result[key] = buildExampleFromSchema(propSchema);
      }
      return result;
    }
    case 'array': {
      const items = schema.items as Record<string, unknown> | undefined;
      if (!items) return [];
      return [buildExampleFromSchema(items)];
    }
    case 'string':
      return schema.default || '';
    case 'number':
    case 'integer':
      return schema.default || 0;
    case 'boolean':
      return schema.default || false;
    default:
      return null;
  }
}
