/**
 * FunctionFly Tool for LangChain
 *
 * Integrates FunctionFly functions as LangChain tools with trust-aware selection.
 *
 * Example:
 *     import { FunctionFlyTool } from '@functionfly/langchain';
 *
 *     const tool = new FunctionFlyTool({
 *       apiKey: 'your-key',
 *       functionId: 'func_xxx',
 *     });
 *
 *     // Use as LangChain tool
 *     const result = await tool.invoke({ input: 'hello world' });
 */

/**
 * Configuration for FunctionFlyTool
 */
export interface FunctionFlyToolConfig {
  /** FunctionFly API key */
  apiKey: string;
  /** Function ID to execute */
  functionId?: string;
  /** Function author (for name-based lookup) */
  author?: string;
  /** Function name (for name-based lookup) */
  name?: string;
  /** FunctionFly API base URL */
  baseUrl?: string;
  /** Minimum trust score (0-100) */
  minTrustScore?: number;
  /** Whether to prefer verified functions */
  preferVerified?: boolean;
  /** Enable automatic retry with fallback */
  enableFallback?: boolean;
  /** Custom name for the tool (defaults to function name) */
  nameOverride?: string;
  /** Custom description for the tool */
  descriptionOverride?: string;
}

/**
 * Trust tier badge for display
 */
export type TrustBadge =
  | "highly_trusted"
  | "verified"
  | "trusted"
  | "unverified";

/**
 * Tool metadata for display and selection
 */
export interface FunctionFlyToolMetadata {
  name: string;
  description: string;
  functionId: string;
  author: string;
  functionName: string;
  trustScore: number;
  trustBadge: TrustBadge;
  verified: boolean;
  schema: Record<string, unknown>;
}

interface FunctionMetadata {
  id: string;
  author: string;
  name: string;
  description: string;
  inputs: Record<string, unknown>;
  trustScore: number;
  verified: boolean;
}

interface FunctionExecutionResult {
  success: boolean;
  output: Record<string, unknown>;
  error?: string;
}

interface ExecuteOptions {
  enableFallback?: boolean;
  minTrustScore?: number;
  retry?: {
    maxRetries: number;
    baseDelayMs: number;
    maxDelayMs: number;
    backoffMultiplier: number;
  };
}

/**
 * Simple HTTP client for FunctionFly API
 */
class SimpleFunctionFlyClient {
  constructor(
    private apiKey: string,
    private baseUrl: string = "https://api.functionfly.com",
  ) {}

  async getFunctions(params: {
    query?: string;
    category?: string;
    minTrustScore?: number;
    limit?: number;
  }): Promise<FunctionMetadata[]> {
    const searchParams = new URLSearchParams();
    if (params.query) searchParams.set("q", params.query);
    if (params.category) searchParams.set("category", params.category);
    if (params.minTrustScore)
      searchParams.set("min_rating", String(params.minTrustScore / 100));
    if (params.limit) searchParams.set("limit", String(params.limit));

    const response = await this.request<{
      functions: Record<string, unknown>[];
    }>(`/v1/registry/functions/search?${searchParams.toString()}`);

    return (response.functions || []).map(this.transformFunction);
  }

  async getFunctionById(functionId: string): Promise<FunctionMetadata> {
    const response = await this.request<{
      functions: Record<string, unknown>[];
    }>(`/v1/registry/functions?limit=100`);
    const found = response.functions?.find(
      (f: Record<string, unknown>) => String(f.id) === functionId,
    );

    if (!found) {
      throw new Error(`Function not found: ${functionId}`);
    }

    return this.transformFunction(found);
  }

  async getFunction(author: string, name: string): Promise<FunctionMetadata> {
    const response = await this.request<Record<string, unknown>>(
      `/v1/registry/functions/${encodeURIComponent(author)}/${encodeURIComponent(name)}`,
    );
    return this.transformFunction(response);
  }

  private transformFunction = (
    data: Record<string, unknown>,
  ): FunctionMetadata => {
    return {
      id: String(data.id || ""),
      author: String(data.author || ""),
      name: String(data.name || ""),
      description: String(data.description || ""),
      inputs: (data.inputs || data.input_schema || {}) as Record<
        string,
        unknown
      >,
      trustScore: Number(data.trust_score || data.trustScore || 0),
      verified: Boolean(data.verified || data.is_verified || false),
    };
  };

  async execute(
    functionId: string,
    input: Record<string, unknown>,
  ): Promise<FunctionExecutionResult> {
    const response = await this.request<Record<string, unknown>>(
      `/v1/functions/${functionId}/execute`,
      {
        method: "POST",
        body: JSON.stringify({ input }),
      },
    );

    return {
      success: Boolean(response.success ?? true),
      output: (response.output || response.result || {}) as Record<
        string,
        unknown
      >,
      error: response.error ? String(response.error) : undefined,
    };
  }

  async executeWithRetry(
    functionId: string,
    input: Record<string, unknown>,
    options: ExecuteOptions = {},
  ): Promise<FunctionExecutionResult> {
    const retryConfig = {
      maxRetries: options.retry?.maxRetries ?? 3,
      baseDelayMs: options.retry?.baseDelayMs ?? 1000,
      maxDelayMs: options.retry?.maxDelayMs ?? 10000,
      backoffMultiplier: options.retry?.backoffMultiplier ?? 2,
    };

    let lastError: Error | undefined;

    for (let attempt = 0; attempt <= retryConfig.maxRetries; attempt++) {
      try {
        const result = await this.execute(functionId, input);
        if (result.success) {
          return result;
        }
        lastError = new Error(result.error || "Execution failed");
      } catch (error) {
        lastError = error instanceof Error ? error : new Error(String(error));

        if (attempt < retryConfig.maxRetries) {
          const delay =
            Math.min(
              retryConfig.baseDelayMs *
                Math.pow(retryConfig.backoffMultiplier, attempt),
              retryConfig.maxDelayMs,
            ) *
            (0.5 + Math.random() * 0.5);

          await new Promise((resolve) => setTimeout(resolve, delay));
          continue;
        }
      }
    }

    throw lastError || new Error("Execution failed after retries");
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    const response = await fetch(url, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${this.apiKey}`,
        ...options.headers,
      },
    });

    if (!response.ok) {
      throw new Error(
        `API request failed: ${response.status} ${response.statusText}`,
      );
    }

    return (await response.json()) as T;
  }
}

/**
 * Get trust badge from score
 */
export function getTrustBadge(score: number, verified: boolean): TrustBadge {
  if (score >= 90 && verified) return "highly_trusted";
  if (score >= 70) return "verified";
  if (score >= 50) return "trusted";
  return "unverified";
}

/**
 * Build JSON schema from function inputs
 */
export function buildSchema(
  inputs: Record<string, unknown>,
): Record<string, unknown> {
  if (Object.keys(inputs).length === 0) {
    return {
      type: "object",
      properties: {
        input: { type: "string", description: "Input data" },
      },
      required: [],
    };
  }

  const properties: Record<string, unknown> = {};
  const required: string[] = [];

  for (const [key, value] of Object.entries(inputs)) {
    properties[key] = {
      type: inferType(value),
      description: `Parameter: ${key}`,
    };
    required.push(key);
  }

  return {
    type: "object",
    properties,
    required,
  };
}

/**
 * Infer JSON type from value
 */
function inferType(value: unknown): string {
  if (value === null || value === undefined) return "string";
  if (typeof value === "boolean") return "boolean";
  if (typeof value === "number") return "number";
  if (Array.isArray(value)) return "array";
  if (typeof value === "object") return "object";
  return "string";
}

/**
 * LangChain-compatible tool wrapping a FunctionFly function
 */
export class FunctionFlyTool {
  /** Tool name */
  name: string;

  /** Tool description */
  description: string;

  /** JSON Schema for tool input */
  schema: Record<string, unknown>;

  /** Whether the function is verified */
  verified: boolean;

  /** Trust score (0-100) */
  trustScore: number;

  /** Trust tier badge */
  trustBadge: TrustBadge;

  /** Function ID */
  functionId: string;

  /** Function author */
  author: string;

  /** Function name */
  functionName: string;

  /** Whether the tool has been initialized */
  isInitialized: boolean = false;

  private client: SimpleFunctionFlyClient;
  private enableFallback: boolean;
  private minTrustScore?: number;
  private config: FunctionFlyToolConfig;

  constructor(config: FunctionFlyToolConfig) {
    if (!config.functionId && !(config.author && config.name)) {
      throw new Error(
        "Either functionId or (author and name) must be provided",
      );
    }

    this.config = config;
    this.client = new SimpleFunctionFlyClient(config.apiKey, config.baseUrl);
    this.enableFallback = config.enableFallback ?? true;
    this.minTrustScore = config.minTrustScore;

    this.name = config.nameOverride || "";
    this.description = config.descriptionOverride || "";
    this.schema = {};
    this.verified = false;
    this.trustScore = 0;
    this.trustBadge = "unverified";
    this.functionId = config.functionId || "";
    this.author = config.author || "";
    this.functionName = config.name || "";
  }

  /**
   * Initialize the tool by fetching function metadata
   */
  async initialize(): Promise<void> {
    if (this.isInitialized) return;

    let functionInfo;

    if (this.functionId) {
      functionInfo = await this.client.getFunctionById(this.functionId);
    } else if (this.author && this.functionName) {
      functionInfo = await this.client.getFunction(
        this.author,
        this.functionName,
      );
      this.functionId = functionInfo.id;
    }

    if (!functionInfo) {
      throw new Error(`Function not found`);
    }

    this.author = functionInfo.author;
    this.functionName = functionInfo.name;
    this.name =
      this.config.nameOverride ||
      `${functionInfo.author}_${functionInfo.name}`.replace(
        /[^a-zA-Z0-9_]/g,
        "_",
      );
    this.description =
      this.config.descriptionOverride ||
      functionInfo.description ||
      `${functionInfo.name} function by ${functionInfo.author}`;
    this.verified = functionInfo.verified;
    this.trustScore = functionInfo.trustScore;
    this.trustBadge = getTrustBadge(
      functionInfo.trustScore,
      functionInfo.verified,
    );
    this.schema = buildSchema(functionInfo.inputs || {});

    this.isInitialized = true;
  }

  /**
   * Get the Zod schema for this tool (compatible with LangChain)
   */
  getInputSchema(): Record<string, unknown> {
    return this.schema;
  }

  /**
   * Execute the tool
   */
  async _call(input: Record<string, unknown>): Promise<string> {
    if (!this.isInitialized) {
      await this.initialize();
    }

    if (this.minTrustScore && this.trustScore < this.minTrustScore) {
      throw new Error(
        `Trust score ${this.trustScore} is below minimum ${this.minTrustScore} for function ${this.name}`,
      );
    }

    let normalizedInput: Record<string, unknown>;
    if (typeof input.input === "string") {
      try {
        normalizedInput = JSON.parse(input.input);
      } catch {
        normalizedInput = { data: input.input };
      }
    } else if (typeof input.input === "object") {
      normalizedInput = input.input as Record<string, unknown>;
    } else {
      normalizedInput = { data: input.input };
    }

    const response = await this.client.executeWithRetry(
      this.functionId,
      normalizedInput,
      {
        enableFallback: this.enableFallback,
        minTrustScore: this.minTrustScore,
        retry: {
          maxRetries: 3,
          baseDelayMs: 1000,
          maxDelayMs: 10000,
          backoffMultiplier: 2,
        },
      },
    );

    if (!response.success) {
      throw new Error(response.error || "Function execution failed");
    }

    return JSON.stringify(response.output);
  }

  /**
   * Invoke alias for LangChain compatibility
   */
  async invoke(input: Record<string, unknown>): Promise<string> {
    return this._call(input);
  }

  /**
   * Get tool metadata for display
   */
  getMetadata(): FunctionFlyToolMetadata {
    return {
      name: this.name,
      description: this.description,
      functionId: this.functionId,
      author: this.author,
      functionName: this.functionName,
      trustScore: this.trustScore,
      trustBadge: this.trustBadge,
      verified: this.verified,
      schema: this.schema,
    };
  }
}

/**
 * Create multiple tools from search criteria
 */
export async function createToolsFromSearch(
  config: FunctionFlyToolConfig & {
    query?: string;
    category?: string;
    limit?: number;
  },
): Promise<FunctionFlyTool[]> {
  const client = new SimpleFunctionFlyClient(config.apiKey, config.baseUrl);
  const functions = await client.getFunctions({
    query: config.query,
    category: config.category,
    minTrustScore: config.minTrustScore,
    limit: config.limit || 20,
  });

  return functions.map((fn) => {
    const tool = new FunctionFlyTool({
      apiKey: config.apiKey,
      functionId: fn.id,
      baseUrl: config.baseUrl,
      minTrustScore: config.minTrustScore,
      preferVerified: config.preferVerified,
      enableFallback: config.enableFallback,
    });

    tool.name = `${fn.author}_${fn.name}`.replace(/[^a-zA-Z0-9_]/g, "_");
    tool.description = fn.description || `${fn.name} function by ${fn.author}`;
    tool.schema = buildSchema(fn.inputs || {});
    tool.verified = fn.verified;
    tool.trustScore = fn.trustScore;
    tool.trustBadge = getTrustBadge(fn.trustScore, fn.verified);
    tool.author = fn.author;
    tool.functionName = fn.name;
    tool.isInitialized = true;

    return tool;
  });
}

/**
 * Create tools from a list of functions
 */
export function createToolsFromFunctions(
  apiKey: string,
  baseUrl: string | undefined,
  functions: Array<{
    id: string;
    author: string;
    name: string;
    description: string;
    inputs: Record<string, unknown>;
    trustScore: number;
    verified: boolean;
  }>,
): FunctionFlyTool[] {
  return functions.map((fn) => {
    const tool = new FunctionFlyTool({
      apiKey,
      functionId: fn.id,
      baseUrl,
    });

    tool.name = `${fn.author}_${fn.name}`.replace(/[^a-zA-Z0-9_]/g, "_");
    tool.description = fn.description || `${fn.name} function by ${fn.author}`;
    tool.schema = buildSchema(fn.inputs || {});
    tool.verified = fn.verified;
    tool.trustScore = fn.trustScore;
    tool.trustBadge = getTrustBadge(fn.trustScore, fn.verified);
    tool.author = fn.author;
    tool.functionName = fn.name;
    tool.isInitialized = true;

    return tool;
  });
}
