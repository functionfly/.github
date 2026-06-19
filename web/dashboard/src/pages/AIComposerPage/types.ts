import type { FunctionGenerationResult } from '@/api/composer';

export interface RefinementChunk {
  type: 'chunk' | 'done' | 'error';
  data?: string;
  result?: FunctionGenerationResult;
  error?: string;
  generation_id?: string;
  refinement_id?: string;
  latency_ms?: number;
  tokens_used?: {
    prompt: number;
    completion: number;
    total: number;
  };
}

export interface RefinementHistoryItem {
  id: string;
  refinementId?: string;
  timestamp: number | Date;
  prompt?: string;
  request?: string;
  result?: FunctionGenerationResult;
  latencyMs?: number;
  tokensUsed?: number;
  success?: boolean;
}

export interface EditableManifest {
  id?: string;
  name?: string;
  description?: string;
  runtime?: string;
  inputs?: Array<{ name: string; type: string; required?: boolean }>;
  outputs?: Array<{ name: string; type: string }>;
  constraints?: string;
  examples?: string[];
  capabilities?: string[];
  timeout_seconds?: number;
  memory_mb?: number;
  version?: string;
}

export interface DraftData {
  id: string;
  description?: string;
  manifest: EditableManifest;
  refinementHistory: RefinementHistoryItem[];
  timestamp?: number;
  createdAt: number;
  updatedAt: number;
  status: 'draft' | 'generating' | 'complete' | 'error';
  constraints?: string;
  runtime?: string;
}

export interface StoredDraft {
  description?: string;
  constraints?: string;
  runtime?: string;
  timestamp?: number;
}

export interface StreamChunk {
  type: 'chunk' | 'done' | 'error';
  data?: string;
  content?: string;
  result?: FunctionGenerationResult;
  error?: string;
  generation_id?: string;
  refinement_id?: string;
  latency_ms?: number;
  tokens_used?: {
    prompt: number;
    completion: number;
    total: number;
  };
  confidence?: number;
  timestamp?: number;
}

export interface ManifestFlow {
  inputs?: Array<{ name: string; type: string }>;
  outputs?: Array<{ name: string; type: string }>;
  nodes?: Array<{
    id: string;
    type: 'input' | 'output' | 'function' | 'condition';
    label: string;
    position: { x: number; y: number };
  }>;
  edges?: Array<{
    source: string;
    target: string;
    label?: string;
  }>;
}
