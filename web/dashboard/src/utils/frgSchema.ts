/** JSON Schema helpers for FRG node port rendering. */

export function countSchemaPorts(schema?: Record<string, unknown>): number {
  if (!schema || typeof schema !== 'object') return 1;

  const properties = schema.properties;
  if (properties && typeof properties === 'object') {
    const keys = Object.keys(properties as Record<string, unknown>);
    return keys.length > 0 ? keys.length : 1;
  }

  return 1;
}

export function listSchemaPropertyNames(schema?: Record<string, unknown>): string[] {
  if (!schema || typeof schema !== 'object') return [];

  const properties = schema.properties;
  if (!properties || typeof properties !== 'object') return [];

  return Object.keys(properties as Record<string, unknown>);
}
