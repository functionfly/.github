// Function Registry API Types

export interface FunctionDocSummary {
  author: string
  name: string
  title: string
  description: string
  category: string
  version: string
  trust_score: number
}

export interface IOType {
  type: string
  example?: unknown
  schema?: Record<string, unknown>
  required?: boolean
}

export interface FunctionManifest {
  name: string
  version: string
  runtime: string
  title?: string
  description?: string
  input?: IOType
  output?: IOType
  timeout_ms?: number
  memory_mb?: number
  capabilities?: string[]
  category?: string
  tags?: string[]
}

export interface FunctionDocs {
  function: FunctionDocSummary
  manifest: FunctionManifest
  runtime: string
  trust_score: number
  success_rate: number
  avg_latency_ms: number
  examples: ExecutionExample[]
  capabilities: string[]
}

export interface ExecutionExample {
  input: unknown
  output: unknown
  cached: boolean
  duration_ms: number
}

export interface FunctionVersion {
  version: string
  published_at: string
  runtime: string
  manifest: FunctionManifest
}

export interface PlaygroundResponse {
  ok: boolean
  data?: unknown
  error?: {
    code: string
    message: string
  }
  cached: boolean
  duration_ms: number
  version: string
}

export interface Category {
  name: string
  count: number
}
