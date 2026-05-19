/**
 * FunctionFly Integration for CrewAI
 *
 * Provides FunctionFly toolkit for CrewAI agents with trust-based routing.
 *
 * Example:
 *     import { FunctionFlyToolkit, createCrew } from '@functionfly/crewai';
 *
 *     // Create toolkit
 *     const toolkit = new FunctionFlyToolkit({
 *       apiKey: 'your-key',
 *       minTrustScore: 70,
 *     });
 *
 *     // Create crew with toolkit
 *     const crew = createCrew({
 *       agents: [agent],
 *       tasks: [task],
 *       tools: toolkit.getTools(),
 *     });
 */

import type { FunctionFlyToolConfig } from "@functionfly/langchain";
import { FunctionFlyTool, createToolsFromSearch } from "@functionfly/langchain";

/**
 * Configuration for FunctionFly CrewAI integration
 */
export interface FunctionFlyCrewAIConfig {
  /** FunctionFly API key */
  apiKey: string;
  /** FunctionFly API base URL */
  baseUrl?: string;
  /** Minimum trust score for tools (0-100) */
  minTrustScore?: number;
  /** Maximum number of tools */
  maxTools?: number;
  /** Enable automatic fallback */
  enableFallback?: boolean;
  /** Custom prefix for tool names */
  toolNamePrefix?: string;
}

/**
 * Trust level
 */
export type TrustLevel =
  | "highly_trusted"
  | "verified"
  | "trusted"
  | "unverified";

/**
 * CrewAI-compatible tool definition
 */
export interface CrewAITool {
  /** Tool name (slug format) */
  name: string;
  /** Human-readable description */
  description: string;
  /** Parameters schema */
  parameters: Record<string, unknown>;
  /** Metadata */
  metadata?: {
    trustScore: number;
    trustLevel: TrustLevel;
    verified: boolean;
    functionId: string;
    author: string;
  };
}

/**
 * Toolkit for CrewAI
 */
export class FunctionFlyToolkit {
  private tools: CrewAITool[] = [];
  private apiKey: string;
  private baseUrl?: string;
  private minTrustScore?: number;
  private enableFallback: boolean;
  private toolNamePrefix: string;

  constructor(config: FunctionFlyCrewAIConfig) {
    this.apiKey = config.apiKey;
    this.baseUrl = config.baseUrl;
    this.minTrustScore = config.minTrustScore;
    this.enableFallback = config.enableFallback ?? true;
    this.toolNamePrefix = config.toolNamePrefix || "functionfly";
  }

  /**
   * Get all registered tools
   */
  getTools(): CrewAITool[] {
    return this.tools;
  }

  /**
   * Get tools filtered by trust level
   */
  getToolsByTrustLevel(level: TrustLevel): CrewAITool[] {
    return this.tools.filter((t) => t.metadata?.trustLevel === level);
  }

  /**
   * Get tool count
   */
  getToolCount(): number {
    return this.tools.length;
  }

  /**
   * Initialize toolkit by loading tools from FunctionFly
   */
  async initialize(config?: {
    category?: string;
    query?: string;
    limit?: number;
  }): Promise<void> {
    const langchainTools = await createToolsFromSearch({
      apiKey: this.apiKey,
      baseUrl: this.baseUrl,
      category: config?.category,
      query: config?.query,
      minTrustScore: this.minTrustScore,
      limit: config?.limit || this.minTrustScore ? 50 : 20,
      enableFallback: this.enableFallback,
    });

    this.tools = langchainTools.map((tool) => this.convertToCrewAITool(tool));
  }

  /**
   * Initialize from a specific function
   */
  async initializeFromFunction(config: {
    functionId?: string;
    author?: string;
    name?: string;
  }): Promise<void> {
    const toolConfig: FunctionFlyToolConfig = {
      apiKey: this.apiKey,
      baseUrl: this.baseUrl,
      functionId: config.functionId,
      author: config.author,
      name: config.name,
      minTrustScore: this.minTrustScore,
      enableFallback: this.enableFallback,
    };

    const tool = new FunctionFlyTool(toolConfig);
    await tool.initialize();

    this.tools = [this.convertToCrewAITool(tool)];
  }

  /**
   * Add a single tool manually
   */
  addTool(config: {
    functionId: string;
    author: string;
    name: string;
    description?: string;
    parameters?: Record<string, unknown>;
    trustScore?: number;
    verified?: boolean;
  }): void {
    const tool: CrewAITool = {
      name: `${this.toolNamePrefix}_${config.author}_${config.name}`
        .toLowerCase()
        .replace(/[^a-z0-9_]/g, "_"),
      description:
        config.description || `${config.name} function by ${config.author}`,
      parameters: config.parameters || { type: "object", properties: {} },
      metadata: {
        trustScore: config.trustScore || 0,
        trustLevel: getTrustLevel(
          config.trustScore || 0,
          config.verified || false,
        ),
        verified: config.verified || false,
        functionId: config.functionId,
        author: config.author,
      },
    };

    this.tools.push(tool);
  }

  /**
   * Convert FunctionFlyTool to CrewAI format
   */
  private convertToCrewAITool(tool: FunctionFlyTool): CrewAITool {
    return {
      name: `${this.toolNamePrefix}_${tool.author}_${tool.functionName}`
        .toLowerCase()
        .replace(/[^a-z0-9_]/g, "_"),
      description: tool.description,
      parameters: tool.schema,
      metadata: {
        trustScore: tool.trustScore,
        trustLevel: tool.trustBadge,
        verified: tool.verified,
        functionId: tool.functionId,
        author: tool.author,
      },
    };
  }

  /**
   * Get tools by category (if metadata available)
   */
  getToolsByAuthor(author: string): CrewAITool[] {
    return this.tools.filter((t) => t.metadata?.author === author);
  }

  /**
   * Get the most trusted tools
   */
  getMostTrustedTools(count: number = 10): CrewAITool[] {
    return [...this.tools]
      .sort(
        (a, b) => (b.metadata?.trustScore || 0) - (a.metadata?.trustScore || 0),
      )
      .slice(0, count);
  }

  /**
   * Get toolkit summary
   */
  getSummary(): {
    totalTools: number;
    byTrustLevel: Record<TrustLevel, number>;
    averageTrustScore: number;
    verifiedCount: number;
  } {
    const byTrustLevel: Record<TrustLevel, number> = {
      highly_trusted: 0,
      verified: 0,
      trusted: 0,
      unverified: 0,
    };

    let totalTrust = 0;
    let verifiedCount = 0;

    for (const tool of this.tools) {
      const level = tool.metadata?.trustLevel || "unverified";
      byTrustLevel[level]++;
      totalTrust += tool.metadata?.trustScore || 0;
      if (tool.metadata?.verified) verifiedCount++;
    }

    return {
      totalTools: this.tools.length,
      byTrustLevel,
      averageTrustScore:
        this.tools.length > 0 ? totalTrust / this.tools.length : 0,
      verifiedCount,
    };
  }
}

/**
 * Get trust level from score
 */
export function getTrustLevel(score: number, verified: boolean): TrustLevel {
  if (score >= 90 && verified) return "highly_trusted";
  if (score >= 70) return "verified";
  if (score >= 50) return "trusted";
  return "unverified";
}

/**
 * Create a task for CrewAI
 */
export function createFunctionFlyTask(config: {
  description: string;
  expectedOutput: string;
  toolName?: string;
  agentName?: string;
}): {
  description: string;
  expected_output: string;
  tool?: string;
  agent?: string;
} {
  return {
    description: config.description,
    expected_output: config.expectedOutput,
    tool: config.toolName,
    agent: config.agentName,
  };
}

/**
 * Validate toolkit for production
 */
export function validateToolkit(
  toolkit: FunctionFlyToolkit,
  options: {
    requireMinimumTrustScore?: number;
    requireAllVerified?: boolean;
  } = {},
): {
  valid: boolean;
  issues: string[];
  warnings: string[];
} {
  const issues: string[] = [];
  const warnings: string[] = [];
  const tools = toolkit.getTools();

  if (tools.length === 0) {
    issues.push("Toolkit has no tools registered");
  }

  for (const tool of tools) {
    if (options.requireMinimumTrustScore) {
      if ((tool.metadata?.trustScore || 0) < options.requireMinimumTrustScore) {
        issues.push(`Tool ${tool.name} has trust score below minimum`);
      }
    }

    if (options.requireAllVerified && !tool.metadata?.verified) {
      warnings.push(`Tool ${tool.name} is not verified`);
    }

    if ((tool.metadata?.trustScore || 0) < 50) {
      warnings.push(`Tool ${tool.name} has low trust score`);
    }
  }

  return {
    valid: issues.length === 0,
    issues,
    warnings,
  };
}

/**
 * Format tools for CrewAI display
 */
export function formatToolsForDisplay(tools: CrewAITool[]): string {
  const lines: string[] = [];

  for (const tool of tools) {
    const trust = tool.metadata
      ? `[${tool.metadata.trustLevel}] ${tool.metadata.trustScore}%`
      : "[unknown]";
    const verified = tool.metadata?.verified ? "✓" : "✗";

    lines.push(`${tool.name} ${verified} ${trust}`);
    lines.push(`  ${tool.description}`);
    lines.push("");
  }

  return lines.join("\n");
}
