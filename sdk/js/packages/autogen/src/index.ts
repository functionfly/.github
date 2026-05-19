/**
 * FunctionFly Integration for AutoGen
 *
 * Provides FunctionFly agent registration and tool discovery for AutoGen agents.
 *
 * Example:
 *     import { FunctionFlyAgent, registerFunctionFlyTools } from '@functionfly/autogen';
 *
 *     // Register tools with AutoGen
 *     const tools = await registerFunctionFlyTools({
 *       apiKey: 'your-key',
 *       minTrustScore: 70,
 *     });
 *
 *     // Create agent with tools
 *     const agent = new FunctionFlyAgent({
 *       name: 'functionfly_assistant',
 *       tools,
 *     });
 */

import { createToolsFromSearch } from "@functionfly/langchain";

/**
 * Configuration for FunctionFly AutoGen integration
 */
export interface FunctionFlyAutoGenConfig {
  /** FunctionFly API key */
  apiKey: string;
  /** FunctionFly API base URL */
  baseUrl?: string;
  /** Minimum trust score for tools (0-100) */
  minTrustScore?: number;
  /** Maximum number of tools to register */
  maxTools?: number;
  /** Enable automatic fallback */
  enableFallback?: boolean;
}

/**
 * FunctionFly agent for AutoGen
 */
export interface FunctionFlyAutoGenAgent {
  /** Agent name */
  name: string;
  /** Agent description */
  description: string;
  /** Registered tools */
  tools: AutoGenTool[];
  /** System message template */
  systemMessage: string;
  /** Minimum trust score */
  minTrustScore: number;
}

/**
 * AutoGen-compatible tool definition
 */
export interface AutoGenTool {
  /** Tool name */
  name: string;
  /** Tool description */
  description: string;
  /** Parameters schema */
  parameters: Record<string, unknown>;
  /** Whether function is verified */
  verified?: boolean;
  /** Trust score (0-100) */
  trustScore?: number;
}

/**
 * Tool registration result
 */
export interface ToolRegistrationResult {
  tools: AutoGenTool[];
  count: number;
  byCategory: Record<string, number>;
  trustDistribution: {
    highlyTrusted: number;
    verified: number;
    trusted: number;
    unverified: number;
  };
}

/**
 * Trust level for categorization
 */
type TrustLevel = "highly_trusted" | "verified" | "trusted" | "unverified";

/**
 * Get trust level from score
 */
function getTrustLevel(score: number, verified: boolean): TrustLevel {
  if (score >= 90 && verified) return "highly_trusted";
  if (score >= 70) return "verified";
  if (score >= 50) return "trusted";
  return "unverified";
}

/**
 * Register FunctionFly tools with AutoGen
 */
export async function registerFunctionFlyTools(
  config: FunctionFlyAutoGenConfig,
): Promise<ToolRegistrationResult> {
  const {
    apiKey,
    baseUrl,
    minTrustScore = 0,
    maxTools = 50,
    enableFallback = true,
  } = config;

  const tools = await createToolsFromSearch({
    apiKey,
    baseUrl,
    minTrustScore,
    limit: maxTools,
    enableFallback,
  });

  const autoGenTools: AutoGenTool[] = tools.map((tool) => ({
    name: tool.name,
    description: tool.description,
    parameters: tool.schema,
    verified: tool.verified,
    trustScore: tool.trustScore,
  }));

  // Categorize tools
  const byCategory: Record<string, number> = {};
  const trustDistribution = {
    highlyTrusted: 0,
    verified: 0,
    trusted: 0,
    unverified: 0,
  };

  for (const tool of autoGenTools) {
    const level = getTrustLevel(tool.trustScore || 0, tool.verified || false);
    trustDistribution[level]++;
  }

  return {
    tools: autoGenTools,
    count: autoGenTools.length,
    byCategory,
    trustDistribution,
  };
}

/**
 * Register tools from a specific category
 */
export async function registerToolsByCategory(
  config: FunctionFlyAutoGenConfig & { category: string },
): Promise<AutoGenTool[]> {
  const tools = await createToolsFromSearch({
    apiKey: config.apiKey,
    baseUrl: config.baseUrl,
    category: config.category,
    minTrustScore: config.minTrustScore,
    limit: config.maxTools,
    enableFallback: config.enableFallback,
  });

  return tools.map((tool) => ({
    name: tool.name,
    description: tool.description,
    parameters: tool.schema,
    verified: tool.verified,
    trustScore: tool.trustScore,
  }));
}

/**
 * Create a FunctionFly agent for AutoGen
 */
export function createFunctionFlyAgent(config: {
  name: string;
  description?: string;
  tools: AutoGenTool[];
  systemMessage?: string;
}): FunctionFlyAutoGenAgent {
  return {
    name: config.name,
    description:
      config.description ||
      `FunctionFly agent with ${config.tools.length} tools`,
    tools: config.tools,
    systemMessage:
      config.systemMessage || getDefaultSystemMessage(config.tools),
    minTrustScore: Math.min(...config.tools.map((t) => t.trustScore || 0)),
  };
}

/**
 * Get default system message for agent
 */
function getDefaultSystemMessage(tools: AutoGenTool[]): string {
  const toolList = tools
    .map((t) => `  - ${t.name}: ${t.description}`)
    .join("\n");

  return `You are a FunctionFly AI assistant with access to the following trusted functions:

${toolList}

Guidelines:
- Always check the trust score of tools before using them
- Prefer highly_trusted and verified tools when available
- If a tool fails, try an alternative trusted tool
- Provide clear responses based on function outputs
- If no suitable function is available, explain this to the user`;
}

/**
 * Filter tools by minimum trust score
 */
export function filterByTrustScore(
  tools: AutoGenTool[],
  minTrustScore: number,
): AutoGenTool[] {
  return tools.filter((t) => (t.trustScore || 0) >= minTrustScore);
}

/**
 * Sort tools by trust score (descending)
 */
export function sortByTrustScore(tools: AutoGenTool[]): AutoGenTool[] {
  return [...tools].sort((a, b) => (b.trustScore || 0) - (a.trustScore || 0));
}

/**
 * Select best tool for a task based on description matching and trust
 */
export function selectBestTool(
  tools: AutoGenTool[],
  taskDescription: string,
): AutoGenTool | undefined {
  if (tools.length === 0) return undefined;

  // Score each tool based on trust and description relevance
  const scored = tools.map((tool) => {
    let score = (tool.trustScore || 0) * 0.7; // Trust is 70% of score

    // Boost score if task description matches tool description
    const descLower = tool.description.toLowerCase();
    const taskLower = taskDescription.toLowerCase();

    const keywords = taskLower.split(/\s+/).filter((w) => w.length > 3);
    const matches = keywords.filter((k) => descLower.includes(k)).length;
    score += (matches / keywords.length) * 30; // 30% for relevance

    return { tool, score };
  });

  scored.sort((a, b) => b.score - a.score);
  return scored[0]?.tool;
}

/**
 * Get tools by trust level
 */
export function getToolsByTrustLevel(
  tools: AutoGenTool[],
  level: TrustLevel,
): AutoGenTool[] {
  return tools.filter(
    (t) => getTrustLevel(t.trustScore || 0, t.verified || false) === level,
  );
}

/**
 * Validate tools for production use
 */
export function validateToolsForProduction(
  tools: AutoGenTool[],
  requireVerified: boolean = false,
): { valid: AutoGenTool[]; invalid: AutoGenTool[] } {
  const valid: AutoGenTool[] = [];
  const invalid: AutoGenTool[] = [];

  for (const tool of tools) {
    const level = getTrustLevel(tool.trustScore || 0, tool.verified || false);

    if (requireVerified && level === "unverified") {
      invalid.push(tool);
      continue;
    }

    if (tool.trustScore && tool.trustScore < 50) {
      invalid.push(tool);
      continue;
    }

    valid.push(tool);
  }

  return { valid, invalid };
}
