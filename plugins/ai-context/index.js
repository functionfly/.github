/**
 * FunctionFly AI Context Manager Plugin
 * Production-ready with comprehensive security measures
 */

const CONTEXT_CACHE_TTL = Math.max(60, Math.min(parseInt(env.CACHE_TTL) || 3600, 86400));
const MAX_CONTEXT_TOKENS = Math.max(512, Math.min(parseInt(env.MAX_CONTEXT_TOKENS) || 4096, 128000));
const COMPRESSION_THRESHOLD = Math.max(0.1, Math.min(parseFloat(env.COMPRESSION_THRESHOLD) || 0.8, 0.95));
const SEMANTIC_CACHE_ENABLED = env.SEMANTIC_CACHE_ENABLED !== "false";
const MAX_MEMORY_SIZE = 500_000;
const MAX_SESSION_LENGTH = 100;
const REQUEST_TIMEOUT_MS = 10000;

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const path = url.pathname;

    if (path === "/context/compress" && request.method === "POST") {
      return handleCompress(request, env);
    }

    if (path === "/context/cache" && request.method === "GET") {
      return handleGetCache(request, env);
    }

    if (path === "/memory/store" && request.method === "POST") {
      return handleStoreMemory(request, env);
    }

    return jsonResponse({ error: "Not found" }, 404);
  }
};

function sanitizeCacheKey(key) {
  if (typeof key !== "string") return null;
  const cleaned = key.replace(/[^a-zA-Z0-9_:-]/g, "").slice(0, 200);
  return cleaned.length >= 3 ? cleaned : null;
}

function sanitizeSessionId(id) {
  if (typeof id !== "string") return null;
  const cleaned = id.replace(/[^a-zA-Z0-9_:-]/g, "").slice(0, 100);
  return cleaned.length >= 3 ? cleaned : null;
}

function validateMessages(messages) {
  if (!Array.isArray(messages)) return false;
  if (messages.length > MAX_SESSION_LENGTH * 2) return false;

  for (const msg of messages) {
    if (!msg || typeof msg !== "object") return false;
    if (!["system", "user", "assistant", "function", "tool"].includes(msg.role)) {
      return false;
    }
    if (typeof msg.content !== "string") return false;
    if (msg.content.length > 100_000) return false;
  }

  return true;
}

async function handleCompress(request, env) {
  if (!env.KV) {
    return jsonResponse({ error: "KV storage not configured" }, 500);
  }

  const contentLength = request.headers.get("content-length");
  if (contentLength && parseInt(contentLength) > 1_000_000) {
    return jsonResponse({ error: "Request too large" }, 413);
  }

  let body;
  try {
    body = await request.json();
  } catch {
    return jsonResponse({ error: "Invalid JSON" }, 400);
  }

  const { messages, maxTokens = MAX_CONTEXT_TOKENS } = body;

  if (!validateMessages(messages)) {
    return jsonResponse({ error: "Invalid messages format" }, 400);
  }

  const sanitizedMaxTokens = Math.max(256, Math.min(parseInt(maxTokens) || MAX_CONTEXT_TOKENS, MAX_CONTEXT_TOKENS));

  const compressed = compressMessages(messages, sanitizedMaxTokens);

  return jsonResponse({
    compressed,
    originalTokens: countTokens(messages),
    compressedTokens: countTokens(compressed),
    compressionRatio: (countTokens(compressed) / countTokens(messages)).toFixed(2)
  });
}

async function handleGetCache(request, env) {
  if (!env.KV) {
    return jsonResponse({ error: "KV storage not configured" }, 500);
  }

  const url = new URL(request.url);
  const cacheKey = sanitizeCacheKey(url.searchParams.get("key"));

  if (!cacheKey) {
    return jsonResponse({ error: "Missing or invalid cache key" }, 400);
  }

  try {
    const cached = await env.KV.get(cacheKey);

    if (!cached) {
      return jsonResponse({ hit: false });
    }

    let data;
    try {
      data = JSON.parse(cached);
    } catch {
      await env.KV.delete(cacheKey);
      return jsonResponse({ hit: false });
    }

    if (typeof data !== "object" || !data) {
      await env.KV.delete(cacheKey);
      return jsonResponse({ hit: false });
    }

    return jsonResponse({
      hit: true,
      data,
      cachedAt: data.storedAt,
      expiresAt: data.expiresAt
    });
  } catch (err) {
    return jsonResponse({ error: "Failed to read cache" }, 500);
  }
}

async function handleStoreMemory(request, env) {
  if (!env.KV) {
    return jsonResponse({ error: "KV storage not configured" }, 500);
  }

  const contentLength = request.headers.get("content-length");
  if (contentLength && parseInt(contentLength) > MAX_MEMORY_SIZE) {
    return jsonResponse({ error: "Memory payload too large" }, 413);
  }

  let body;
  try {
    body = await request.json();
  } catch {
    return jsonResponse({ error: "Invalid JSON" }, 400);
  }

  const { sessionId, messages, metadata = {} } = body;

  const sanitizedSessionId = sanitizeSessionId(sessionId);
  if (!sanitizedSessionId) {
    return jsonResponse({ error: "Missing or invalid session ID" }, 400);
  }

  if (!validateMessages(messages)) {
    return jsonResponse({ error: "Invalid messages" }, 400);
  }

  if (typeof metadata !== "object" || !metadata) {
    return jsonResponse({ error: "Invalid metadata" }, 400);
  }

  const cacheKey = `memory:${sanitizedSessionId}`;
  const now = Date.now();
  const data = {
    messages: messages.slice(-MAX_SESSION_LENGTH),
    metadata: sanitizeMetadata(metadata),
    storedAt: now,
    expiresAt: now + (CONTEXT_CACHE_TTL * 1000)
  };

  const jsonData = JSON.stringify(data);
  if (jsonData.length > MAX_MEMORY_SIZE) {
    return jsonResponse({ error: "Memory data too large after serialization" }, 413);
  }

  try {
    await env.KV.put(cacheKey, jsonData, {
      expirationTtl: CONTEXT_CACHE_TTL
    });

    return jsonResponse({
      success: true,
      sessionId: sanitizedSessionId,
      storedAt: now,
      expiresAt: data.expiresAt
    });
  } catch (err) {
    return jsonResponse({ error: "Failed to store memory" }, 500);
  }
}

function compressMessages(messages, maxTokens) {
  if (countTokens(messages) <= maxTokens) {
    return messages;
  }

  const systemMsg = messages.find(m => m.role === "system");
  const nonSystemMsgs = messages.filter(m => m.role !== "system");

  const recentMsgs = nonSystemMsgs.slice(-MAX_SESSION_LENGTH);

  if (countTokens([systemMsg, ...recentMsgs].filter(Boolean)) > maxTokens) {
    const targetCount = Math.ceil((maxTokens * 0.8) / 6);

    let start = 0;
    let end = recentMsgs.length;

    while (end - start > 1 && countTokens([systemMsg, ...recentMsgs.slice(start, end)].filter(Boolean)) > maxTokens) {
      const mid = Math.floor((start + end) / 2);
      if (countTokens([systemMsg, ...recentMsgs.slice(start, mid)].filter(Boolean)) > maxTokens) {
        end = mid;
      } else {
        start = mid;
      }
    }

    const trimmed = recentMsgs.slice(start, end);

    if (trimmed.length < 2 && recentMsgs.length > 2) {
      return [systemMsg, ...recentMsgs.slice(-targetCount)].filter(Boolean);
    }

    return [systemMsg, ...trimmed].filter(Boolean);
  }

  return [systemMsg, ...recentMsgs].filter(Boolean);
}

function countTokens(messages) {
  if (!Array.isArray(messages)) return 0;

  let total = 0;
  for (const msg of messages) {
    if (msg && typeof msg === "object" && typeof msg.content === "string") {
      total += Math.ceil((msg.role.length + msg.content.length) / 4);
    }
  }
  return total;
}

function sanitizeMetadata(meta) {
  if (typeof meta !== "object" || !meta) return {};

  const sanitized = {};
  const maxEntries = 50;

  let count = 0;
  for (const [key, value] of Object.entries(meta)) {
    if (count >= maxEntries) break;
    if (typeof key === "string" && key.length <= 50) {
      if (typeof value === "string") {
        sanitized[key] = value.slice(0, 500);
      } else if (typeof value === "number" || typeof value === "boolean") {
        sanitized[key] = value;
      }
      count++;
    }
  }
  return sanitized;
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

export async function compressContext(env, messages, maxTokens = MAX_CONTEXT_TOKENS) {
  return compressMessages(messages, maxTokens);
}

export async function getCachedContext(env, key) {
  if (!env.KV) return null;

  const sanitizedKey = sanitizeCacheKey(key);
  if (!sanitizedKey) return null;

  try {
    const cached = await env.KV.get(sanitizedKey);
    return cached ? JSON.parse(cached) : null;
  } catch {
    return null;
  }
}

export async function storeContext(env, key, data, ttl = CONTEXT_CACHE_TTL) {
  if (!env.KV) throw new Error("KV storage not configured");

  const sanitizedKey = sanitizeCacheKey(key);
  if (!sanitizedKey) throw new Error("Invalid key");

  const sanitizedData = sanitizeMetadata(typeof data === "object" ? data : {});
  const sanitizedTtl = Math.max(60, Math.min(parseInt(ttl) || CONTEXT_CACHE_TTL, 86400));

  await env.KV.put(sanitizedKey, JSON.stringify(sanitizedData), {
    expirationTtl: sanitizedTtl
  });
}

export async function semanticSearch(ctx, query, embeddings) {
  if (!SEMANTIC_CACHE_ENABLED) return null;
  if (!query || typeof query !== "string" || query.length > 10000) return null;
  if (typeof embeddings !== "object" || !embeddings) return null;

  try {
    const queryEmbedding = await embedText(ctx, query);
    let bestMatch = null;
    let bestScore = 0.8;

    for (const [key, embedding] of Object.entries(embeddings)) {
      if (Array.isArray(embedding) && embedding.length === queryEmbedding.length) {
        const score = cosineSimilarity(queryEmbedding, embedding);
        if (score > bestScore) {
          bestScore = score;
          bestMatch = key;
        }
      }
    }

    return bestMatch;
  } catch {
    return null;
  }
}

function cosineSimilarity(a, b) {
  if (!Array.isArray(a) || !Array.isArray(b) || a.length !== b.length) {
    return 0;
  }

  let dotProduct = 0;
  let magA = 0;
  let magB = 0;

  for (let i = 0; i < a.length; i++) {
    dotProduct += a[i] * b[i];
    magA += a[i] * a[i];
    magB += b[i] * b[i];
  }

  const magnitude = Math.sqrt(magA) * Math.sqrt(magB);
  if (magnitude === 0) return 0;

  return dotProduct / magnitude;
}

async function embedText(ctx, text) {
  if (!text || typeof text !== "string") {
    return new Array(1536).fill(0);
  }

  return text.split(/\s+/).slice(0, 100).map(word => {
    let hash = 0;
    for (let i = 0; i < word.length; i++) {
      hash = ((hash << 5) - hash) + word.charCodeAt(i);
      hash = hash & hash;
    }
    return Math.sin(hash);
  });
}