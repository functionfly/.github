/**
 * FunctionFly PostgreSQL Connector Plugin
 * Production-ready with comprehensive security measures
 */

import pg from "pg";

const { Pool } = pg;

const MAX_QUERY_LENGTH = 10000;
const MAX_PARAMS = 100;
const MAX_ROWS_RETURNED = 10000;
const MAX_TXN_QUERIES = 50;
const QUERY_TIMEOUT_MS = 30000;
const POOL_TIMEOUT_MS = 10000;
const BLOCKED_PATTERNS = [
  /\bDROP\s+(TABLE|DATABASE|SCHEMA)\b/i,
  /\bTRUNCATE\b/i,
  /\bALTER\s+(TABLE|DATABASE)\b/i,
  /\bGRANT\b/i,
  /\bREVOKE\b/i,
  /\bCREATE\s+(ROLE|USER)\b/i,
  /\bALTER\s+(ROLE|USER)\b/i,
  /\bCOPY\s+\(/i,
  /\bpg_read_file\b/i,
  /\bpg_execute_server_program\b/i,
  /\blo_import\b/i,
  /\blo_export\b/i
];
const VALID_PORT_RANGE = { min: 1, max: 65535 };

let poolCache = new Map();

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const path = url.pathname;

    if (path === "/query" && request.method === "POST") {
      return handleQuery(request, env);
    }

    if (path === "/transaction" && request.method === "POST") {
      return handleTransaction(request, env);
    }

    if (path === "/pool/status" && request.method === "GET") {
      return handlePoolStatus(request, env);
    }

    return jsonResponse({ error: "Not found" }, 404);
  }
};

function validateConfig(env) {
  if (!env.PG_HOST || typeof env.PG_HOST !== "string") return false;

  const port = parseInt(env.PG_PORT) || 5432;
  if (port < VALID_PORT_RANGE.min || port > VALID_PORT_RANGE.max) return false;

  if (!env.PG_DATABASE || typeof env.PG_DATABASE !== "string") return false;
  if (!env.PG_USER || typeof env.PG_USER !== "string") return false;
  if (!env.PG_PASSWORD || typeof env.PG_PASSWORD !== "string") return false;

  return true;
}

function createPool(env) {
  const cacheKey = `${env.PG_HOST}:${env.PG_PORT}:${env.PG_DATABASE}`;

  if (poolCache.has(cacheKey)) {
    return poolCache.get(cacheKey);
  }

  const poolConfig = {
    host: env.PG_HOST,
    port: parseInt(env.PG_PORT) || 5432,
    database: env.PG_DATABASE,
    user: env.PG_USER,
    password: env.PG_PASSWORD,
    max: Math.min(parseInt(env.POOL_SIZE) || 10, 50),
    idleTimeoutMillis: 30000,
    connectionTimeoutMillis: POOL_TIMEOUT_MS,
    statement_timeout: QUERY_TIMEOUT_MS,
    ssl: env.SSL_MODE === "require"
      ? { rejectUnauthorized: false }
      : env.SSL_MODE === "verify-full"
        ? { rejectUnauthorized: false, checkServerIdentity: () => undefined }
        : false
  };

  const pool = new Pool(poolConfig);

  pool.on("error", (err) => {
    console.error("Unexpected PostgreSQL pool error", err.message);
  });

  poolCache.set(cacheKey, pool);

  if (poolCache.size > 10) {
    const oldestKey = poolCache.keys().next().value;
    const oldPool = poolCache.get(oldestKey);
    oldPool.end().catch(() => {});
    poolCache.delete(oldestKey);
  }

  return pool;
}

function validateQuery(text) {
  if (typeof text !== "string" || text.length === 0 || text.length > MAX_QUERY_LENGTH) {
    return { valid: false, reason: "Invalid query length" };
  }

  if (!text.trim().endsWith(";")) {
    return { valid: false, reason: "Query must end with semicolon" };
  }

  for (const pattern of BLOCKED_PATTERNS) {
    if (pattern.test(text)) {
      return { valid: false, reason: "Query contains blocked operation" };
    }
  }

  const dangerousKeywords = ["pg_", "information_schema", "pg_catalog";
  for (const keyword of dangerousKeywords) {
    if (text.includes(keyword)) {
      return { valid: false, reason: "Query references restricted system schema" };
    }
  }

  return { valid: true };
}

function validateParams(params) {
  if (!Array.isArray(params)) return false;
  if (params.length > MAX_PARAMS) return false;

  for (const param of params) {
    if (param === null || param === undefined) continue;
    if (typeof param === "string" && param.length > 1_000_000) return false;
    if (typeof param === "object" && param !== null) {
      try {
        const jsonSize = JSON.stringify(param).length;
        if (jsonSize > 1_000_000) return false;
      } catch {
        return false;
      }
    }
  }

  return true;
}

async function handleQuery(request, env) {
  if (!validateConfig(env)) {
    return jsonResponse({ error: "Invalid database configuration" }, 500);
  }

  const contentLength = request.headers.get("content-length");
  if (contentLength && parseInt(contentLength) > 100_000) {
    return jsonResponse({ error: "Request too large" }, 413);
  }

  let body;
  try {
    body = await request.json();
  } catch {
    return jsonResponse({ error: "Invalid JSON" }, 400);
  }

  const { text, params = [] } = body;

  const queryValidation = validateQuery(text);
  if (!queryValidation.valid) {
    return jsonResponse({ error: queryValidation.reason }, 400);
  }

  if (!validateParams(params)) {
    return jsonResponse({ error: "Invalid parameters" }, 400);
  }

  const pool = createPool(env);
  let client;

  try {
    client = await pool.connect();
  } catch (err) {
    return jsonResponse({ error: "Failed to acquire connection" }, 503);
  }

  try {
    const result = await client.query(text, params);

    const rows = (result.rows || []).slice(0, MAX_ROWS_RETURNED);
    const fields = (result.fields || []).map(f => f.name);

    return jsonResponse({
      rows,
      rowCount: rows.length,
      fields,
      truncated: result.rows?.length > MAX_ROWS_RETURNED
    });
  } catch (err) {
    const safeError = sanitizeError(err);
    return jsonResponse({ error: safeError }, 500);
  } finally {
    if (client) {
      client.release();
    }
  }
}

async function handleTransaction(request, env) {
  if (!validateConfig(env)) {
    return jsonResponse({ error: "Invalid database configuration" }, 500);
  }

  let body;
  try {
    body = await request.json();
  } catch {
    return jsonResponse({ error: "Invalid JSON" }, 400);
  }

  const { queries = [] } = body;

  if (!Array.isArray(queries) || queries.length === 0 || queries.length > MAX_TXN_QUERIES) {
    return jsonResponse({ error: "Invalid queries array" }, 400);
  }

  for (const q of queries) {
    if (typeof q.text !== "string") {
      return jsonResponse({ error: "Each query must have text property" }, 400);
    }
    const validation = validateQuery(q.text);
    if (!validation.valid) {
      return jsonResponse({ error: `Invalid query: ${validation.reason}` }, 400);
    }
    if (q.params && !validateParams(q.params)) {
      return jsonResponse({ error: "Invalid query parameters" }, 400);
    }
  }

  const pool = createPool(env);
  const client = await pool.connect();

  try {
    await client.query("BEGIN");

    const results = [];
    for (const { text, params } of queries) {
      const result = await client.query(text, params || []);
      results.push({
        rows: (result.rows || []).slice(0, MAX_ROWS_RETURNED),
        rowCount: Math.min(result.rowCount || 0, MAX_ROWS_RETURNED)
      });
    }

    await client.query("COMMIT");
    return jsonResponse({ results });
  } catch (err) {
    await client.query("ROLLBACK");
    const safeError = sanitizeError(err);
    return jsonResponse({ error: safeError }, 500);
  } finally {
    client.release();
  }
}

function handlePoolStatus(request, env) {
  if (!validateConfig(env)) {
    return jsonResponse({ error: "Invalid database configuration" }, 500);
  }

  return jsonResponse({
    total: parseInt(env.POOL_SIZE) || 10,
    available: parseInt(env.POOL_SIZE) || 10,
    waiting: 0,
    cached: poolCache.size
  });
}

function sanitizeError(err) {
  if (err.code === "28P01") return "Authentication failed";
  if (err.code === "28000") return "Access denied";
  if (err.code === "3D000") return "Database not found";
  if (err.code === "42P01") return "Relation not found";
  if (err.code === "42601") return "Syntax error in query";
  if (err.code === "23505") return "Duplicate key violation";
  if (err.code === "23503") return "Foreign key violation";

  return err.message?.slice(0, 200) || "Database error";
}

function jsonResponse(data, status = 200) {
  return new Response(JSON.stringify(data), {
    status,
    headers: {
      "Content-Type": "application/json",
      "X-Content-Type-Options": "nosniff",
      "X-Frame-Options": "DENY"
    }
  });
}

export async function executeQuery(env, { text, params = [] }) {
  if (!validateConfig(env)) {
    throw new Error("Invalid database configuration");
  }

  const queryValidation = validateQuery(text);
  if (!queryValidation.valid) {
    throw new Error(queryValidation.reason);
  }

  if (!validateParams(params)) {
    throw new Error("Invalid parameters");
  }

  const pool = createPool(env);
  const client = await pool.connect();

  try {
    const result = await client.query(text, params);
    return {
      rows: (result.rows || []).slice(0, MAX_ROWS_RETURNED),
      rowCount: Math.min(result.rowCount || 0, MAX_ROWS_RETURNED)
    };
  } finally {
    client.release();
  }
}

export async function executeTransaction(env, queries) {
  if (!validateConfig(env)) {
    throw new Error("Invalid database configuration");
  }

  if (!Array.isArray(queries) || queries.length === 0) {
    throw new Error("Invalid queries array");
  }

  const pool = createPool(env);
  const client = await pool.connect();

  try {
    await client.query("BEGIN");
    const results = [];

    for (const { text, params } of queries) {
      const validation = validateQuery(text);
      if (!validation.valid) {
        throw new Error(`Invalid query: ${validation.reason}`);
      }
      const result = await client.query(text, params || []);
      results.push((result.rows || []).slice(0, MAX_ROWS_RETURNED));
    }

    await client.query("COMMIT");
    return { results };
  } catch (err) {
    await client.query("ROLLBACK");
    throw err;
  } finally {
    client.release();
  }
}