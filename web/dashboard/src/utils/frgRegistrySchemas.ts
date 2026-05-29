/** Extract JSON schemas from registry function manifests and API payloads. */

const DEFAULT_SCHEMA: Record<string, unknown> = { type: 'object', properties: {} };

function asObject(value: unknown): Record<string, unknown> | undefined {
  return value && typeof value === 'object' && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : undefined;
}

export function schemaFromManifestField(
  manifest: Record<string, unknown> | undefined,
  schemaKey: 'input_schema' | 'output_schema',
  legacyKey: 'input' | 'output'
): Record<string, unknown> {
  if (!manifest) return DEFAULT_SCHEMA;

  const direct = asObject(manifest[schemaKey]);
  if (direct) return direct;

  const legacy = asObject(manifest[legacyKey]);
  if (legacy?.schema) {
    const nested = asObject(legacy.schema);
    if (nested) return nested;
  }

  if (legacy?.properties) {
    return { type: 'object', properties: legacy.properties };
  }

  return DEFAULT_SCHEMA;
}

export function schemasFromRegistryManifest(
  manifest: unknown
): { inputSchema: Record<string, unknown>; outputSchema: Record<string, unknown> } {
  const parsed = asObject(manifest);
  return {
    inputSchema: schemaFromManifestField(parsed, 'input_schema', 'input'),
    outputSchema: schemaFromManifestField(parsed, 'output_schema', 'output'),
  };
}

export async function enrichCatalogItemSchemas(
  item: {
    author: string;
    name: string;
    inputSchema: Record<string, unknown>;
    outputSchema: Record<string, unknown>;
  },
  fetchDetail: (
    author: string,
    name: string
  ) => Promise<{ versions?: Array<{ manifest?: unknown }> }>
): Promise<{ inputSchema: Record<string, unknown>; outputSchema: Record<string, unknown> }> {
  const hasProperties = (schema: Record<string, unknown>) =>
    Object.keys(asObject(schema.properties) ?? {}).length > 0;

  if (hasProperties(item.inputSchema) || hasProperties(item.outputSchema)) {
    return { inputSchema: item.inputSchema, outputSchema: item.outputSchema };
  }

  try {
    const detail = await fetchDetail(item.author, item.name);
    const manifest = detail.versions?.[0]?.manifest;
    return schemasFromRegistryManifest(manifest);
  } catch {
    return { inputSchema: item.inputSchema, outputSchema: item.outputSchema };
  }
}
