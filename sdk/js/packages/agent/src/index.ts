/**
 * FunctionFly Agent SDK
 *
 * Core client for AI agent integrations with FunctionFly function registry.
 * Provides function discovery, trust-aware routing, and automatic retry with fallback.
 *
 * Example:
 *     import { AgentClient } from '@functionfly/agent';
 *
 *     const client = new AgentClient({ apiKey: 'your-key', baseUrl: 'https://api.functionfly.com' });
 *
 *     // Discover functions
 *     const functions = await client.discoverFunctions({ category: 'data-processing', minTrustScore: 80 });
 *
 *     // Execute with automatic retry and fallback
 *     const result = await client.executeWithRetry(functionId, input);
 */

export * from "./client.js";
export * from "./discovery.js";
export * from "./execution.js";
export * from "./trust.js";
export * from "./types.js";

// SDK version
export const VERSION = "1.0.0";
