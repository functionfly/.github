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
