/**
 * KV Store TypeScript Function
 *
 * Demonstrates using KV (key-value) storage in FunctionFly functions.
 * Supports get, set, delete, and list operations.
 *
 * @module kv-store
 */

import type {
  Handler,
  Request,
  Env,
  Context,
  Response,
  KVNamespace,
} from "./types/functionfly";

/**
 * Input actions for KV operations
 */
type KVAction = "get" | "set" | "delete" | "list";

/**
 * Input interface for the kv-store function
 */
interface KVStoreInput {
  action: KVAction;
  key?: string;
  value?: string;
  ttl?: number;
  prefix?: string;
}

/**
 * Output interface for the kv-store function
 */
interface KVStoreOutput {
  success: boolean;
  value?: string;
  keys?: string[];
  error?: string;
}

/**
 * Prepends the configured prefix to a key
 */
function getPrefixedKey(env: Env, key: string): string {
  const prefix = env.KV_PREFIX ?? "demo:";
  return `${prefix}${key}`;
}

/**
 * Handle GET operation
 */
async function handleGet(
  kv: KVNamespace,
  env: Env,
  key: string,
): Promise<KVStoreOutput> {
  if (!key) {
    return { success: false, error: "Key is required for GET operation" };
  }

  const prefixedKey = getPrefixedKey(env, key);
  const value = await kv.get(prefixedKey);

  if (value === null) {
    return { success: false, error: "Key not found" };
  }

  return { success: true, value };
}

/**
 * Handle SET operation
 */
async function handleSet(
  kv: KVNamespace,
  env: Env,
  key: string,
  value: string,
  ttl?: number,
): Promise<KVStoreOutput> {
  if (!key) {
    return { success: false, error: "Key is required for SET operation" };
  }

  if (value === undefined) {
    return { success: false, error: "Value is required for SET operation" };
  }

  const prefixedKey = getPrefixedKey(env, key);

  await kv.put(prefixedKey, value, ttl ? { expirationTtl: ttl } : undefined);

  return { success: true };
}

/**
 * Handle DELETE operation
 */
async function handleDelete(
  kv: KVNamespace,
  env: Env,
  key: string,
): Promise<KVStoreOutput> {
  if (!key) {
    return { success: false, error: "Key is required for DELETE operation" };
  }

  const prefixedKey = getPrefixedKey(env, key);

  await kv.delete(prefixedKey);

  return { success: true };
}

/**
 * Handle LIST operation
 */
async function handleList(
  kv: KVNamespace,
  env: Env,
  prefix?: string,
): Promise<KVStoreOutput> {
  const listPrefix = prefix
    ? getPrefixedKey(env, prefix)
    : (env.KV_PREFIX ?? "demo:");

  const result = await kv.list({ prefix: listPrefix });

  const keys = result.keys.map((k) => k.name);

  return { success: true, keys };
}

/**
 * Main handler function for the kv-store function.
 *
 * @param request - The incoming HTTP request
 * @param env - Environment variables
 * @param context - Function execution context with KV storage
 * @returns Response with KV operation result
 */
const handler: Handler = async (
  request: Request,
  env: Env,
  context: Context,
): Promise<Response> => {
  // Parse the request body
  let input: KVStoreInput;

  try {
    input = await request.json<KVStoreInput>();
  } catch {
    return {
      status: 400,
      headers: { "Content-Type": "application/json" },
      body: { success: false, error: "Invalid JSON body" },
    };
  }

  // Validate action
  const validActions: KVAction[] = ["get", "set", "delete", "list"];
  if (!input.action || !validActions.includes(input.action)) {
    return {
      status: 400,
      headers: { "Content-Type": "application/json" },
      body: {
        success: false,
        error: `Invalid action. Must be one of: ${validActions.join(", ")}`,
      },
    };
  }

  // Execute the requested operation
  let output: KVStoreOutput;

  try {
    switch (input.action) {
      case "get":
        output = await handleGet(context.kv, env, input.key!);
        break;
      case "set":
        output = await handleSet(
          context.kv,
          env,
          input.key!,
          input.value!,
          input.ttl,
        );
        break;
      case "delete":
        output = await handleDelete(context.kv, env, input.key!);
        break;
      case "list":
        output = await handleList(context.kv, env, input.prefix);
        break;
      default:
        output = { success: false, error: "Unknown action" };
    }
  } catch (error) {
    const errorMessage =
      error instanceof Error ? error.message : "Unknown error";
    return {
      status: 500,
      headers: { "Content-Type": "application/json" },
      body: { success: false, error: `KV operation failed: ${errorMessage}` },
    };
  }

  // Return the result
  return {
    status: output.success ? 200 : 400,
    headers: { "Content-Type": "application/json" },
    body: output,
  };
};

export { handler };
export default handler;
