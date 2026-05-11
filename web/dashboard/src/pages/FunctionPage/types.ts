export interface FunctionManifest {
  input?: {
    type?: string;
    properties?: Record<string, ManifestProperty>;
    required?: string[];
    description?: string;
    schema?: Record<string, unknown>;
    example?: unknown;
  };
  output?: {
    type?: string;
    properties?: Record<string, ManifestProperty>;
    required?: string[];
    description?: string;
    schema?: Record<string, unknown>;
    example?: unknown;
  };
  deterministic?: boolean;
  deprecated?: boolean;
  successor?: string;
  rate_limit?: number;
  auth?: string;
  endpoint?: string;
}

export interface ManifestProperty {
  type?: string;
  description?: string;
  default?: unknown;
  example?: unknown;
  title?: string;
  enum?: string[];
  properties?: Record<string, ManifestProperty>;
  required?: string[];
  items?: ManifestProperty;
  [key: string]: unknown;
}

export interface FunctionInfo {
  id: string;
  author: string;
  name: string;
  version: string;
  title?: string;
  description?: string;
  runtime: string;
  category?: string;
  tags: string[];
  price_per_call: number;
  reliability: number;
  deterministic: boolean;
  cache_ttl: number;
  input_type?: string;
  output_type?: string;
  input_example?: unknown;
  output_example?: unknown;
  manifest?: FunctionManifest;
  stars?: number;
  executions?: number;
  created_at?: string;
  updated_at?: string;
  popularity_score?: number;
  readme?: string;
  documentation_url?: string;
  repo_url?: string;
  trust_score?: number;
  trust_level?: string;
  verified?: boolean;
  capabilities?: string[];
  source_hash?: string;
  dna_generation?: number;
  dna_fitness_score?: number;
  dna_total_mutations?: number;
  dna_total_executions?: number;
}
