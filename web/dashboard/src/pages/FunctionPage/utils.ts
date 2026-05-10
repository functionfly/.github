import type { FraudRiskLevel, FunctionCardData, FunctionHeaderData, TrustMetrics, TrustTier } from '@/types';
import type { FunctionInfo } from './types';

export interface DNATrustData {
  generation: number;
  fitnessScore: number;
  totalMutations: number;
  totalExecutions: number;
}

export function mapToFunctionHeaderData(
  data: FunctionInfo,
  trustTier: TrustTier = 'medium'
): FunctionHeaderData {
  const executionRootHash = data.source_hash
    ? data.source_hash.startsWith('0x')
      ? data.source_hash
      : `0x${data.source_hash}`
    : `0x${data.author}${data.name}`.slice(0, 66).padEnd(66, '0');

  const resourceSignature = `res_sig_${data.author.slice(0, 8)}${data.name.slice(0, 8)}`;

  const mappedTrustTier: TrustTier =
    (data.trust_level?.toLowerCase() as TrustTier) ||
    (data.trust_score != null && data.trust_score >= 80
      ? 'high'
      : data.trust_score != null && data.trust_score >= 50
        ? 'medium'
        : data.trust_score != null && data.trust_score >= 20
          ? 'low'
          : trustTier);

  const economicScore = Math.round((data.reliability || 0.5) * 100);

  return {
    name: data.title || data.name,
    id: `${data.author}/${data.name}`,
    executionRootHash,
    trustTier: mappedTrustTier,
    economicScore,
    runtime: data.runtime,
    resourceSignature,
    fxcert: {
      verified: data.verified === true || (data.verified !== false && data.deterministic) || false,
      issuedAt: data.created_at,
      issuer: 'FunctionFly Registry',
    },
    status: 'online',
    version: `v${data.version}`,
    description: data.description,
  };
}

export function mapToFunctionCardData(data: FunctionInfo): FunctionCardData {
  return {
    id: `${data.author}/${data.name}`,
    name: data.title || data.name,
    description: data.description || 'No description available',
    author: {
      id: data.author,
      username: data.author,
      name: data.author,
    },
    trustScore: data.trust_score ?? Math.round((data.reliability || 0.5) * 100),
    metrics: {
      executionCount: data.executions || 0,
      executionTrend: data.popularity_score ? [data.popularity_score] : undefined,
    },
    pricing: {
      model: data.price_per_call === 0 ? 'free' : 'per_call',
      pricePerCall: data.price_per_call,
      currency: 'USD',
    },
    isVerified: data.deterministic || false,
    isDeterministic: data.deterministic || false,
    rating: {
      average: data.stars ? Math.min(data.stars / 20, 5) : 0,
      count: data.stars || 0,
    },
    tags: data.tags,
    category: data.category,
    language: data.runtime,
    lastUpdated: data.updated_at,
    version: data.version,
    isFavorite: false,
    isFeatured: data.popularity_score ? data.popularity_score > 80 : false,
  };
}

export function mapToTrustMetrics(data: FunctionInfo): TrustMetrics {
  const overallScore = data.trust_score ?? Math.round((data.reliability || 0.5) * 100);
  const reliability = Math.round((data.reliability || 0.5) * 100);
  const latency: number | undefined = undefined;
  const determinism = data.deterministic ? 95 : 50;
  const communityReputation = data.stars ? Math.min(data.stars, 100) : 50;
  const fraudRisk: FraudRiskLevel | undefined = undefined;

  return {
    overallScore,
    reliability,
    latency,
    determinism,
    communityReputation,
    fraudRisk,
    details: {
      totalExecutions: data.executions,
      lastUpdated: data.updated_at,
    },
  };
}

export function mapToDNATrustData(data: FunctionInfo): DNATrustData | null {
  if (!data.dna_generation || data.dna_generation === 0) {
    return null;
  }
  return {
    generation: data.dna_generation,
    fitnessScore: data.dna_fitness_score ?? 0,
    totalMutations: data.dna_total_mutations ?? 0,
    totalExecutions: data.dna_total_executions ?? 0,
  };
}

export function generateCodeExamples(fn: FunctionInfo) {
  const baseUrl = window.location.origin;
  const executeUrl = `${baseUrl}/v1/fx/${fn.author}/${fn.name}`;
  const inputExample = fn.input_example || fn.manifest?.input?.example || {};

  return {
    curl: `curl -X POST "${executeUrl}" \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer <your-token>" \\
  -d '${JSON.stringify(inputExample)}'`,
    javascript: `const response = await fetch('${executeUrl}', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer <your-token>',
  },
  body: JSON.stringify(${JSON.stringify(inputExample, null, 2)})
});

const result = await response.json();
console.log(result);`,
    python: `import requests

response = requests.post('${executeUrl}', json=${JSON.stringify(inputExample)}, headers={
    'Authorization': 'Bearer <your-token>',
})
result = response.json()
print(result)`,
  };
}
