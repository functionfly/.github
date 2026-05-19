/**
 * FunctionFly Rate Limiter Plugin
 * Production-ready with comprehensive security measures
 */

const ALGORITHM = ["token_bucket", "sliding_window", "fixed_window"].includes(env.ALGORITHM)
  ? env.ALGORITHM
  : "token_bucket";
const DEFAULT_LIMIT = Math.max(1, Math.min(parseInt(env.DEFAULT_LIMIT) || 100, 10000);
const WINDOW_SIZE = Math.max(1, Math.min(parseInt(env.WINDOW_SIZE_SECONDS) || 60, 3600));
const MAX_KEY_LENGTH = 200;
const MAX_TOKENS_PER_REQUEST = 10;
const REQUEST_TIMEOUT_MS = 5000;

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const path = url.pathname;

    if (path === "/limit/check" && request.method === "POST") {
      return handleCheckLimit(request, env);
    }

    const resetMatch = path.match(/^\/limit\/([^/]+)\/reset$/);
    if (resetMatch && request.method === "POST") {
      return handleResetLimit(resetMatch[1], request, env);
    }

    return jsonResponse({ error: "Not found" }, 404);
  }
};

function sanitizeKey(key) {
  if (typeof key !== "string") return null;
  const cleaned = key.replace(/[^a-zA-Z0-9_:.@#$%&/-]/g, "").slice(0, MAX_KEY_LENGTH);
  return cleaned.length >= 1 ? cleaned : null;
}

function sanitizeLimit(limit) {
  const parsed = parseInt(limit);
  if (isNaN(parsed) || parsed < 1) return DEFAULT_LIMIT;
  return Math.min(parsed, 10000);
}

function sanitizeAlgorithm(algorithm) {
  if (typeof algorithm !== "string") return ALGORITHM;
  return ["token_bucket", "sliding_window", "fixed_window"].includes(algorithm)
    ? algorithm
    : ALGORITHM;
}

function getStorageKey(key, algorithm) {
  return `ratelimit:${algorithm}:${key}`;
}

async function handleCheckLimit(request, env) {
  if (!env.KV) {
    return jsonResponse({ error: "Storage not configured" }, 500);
  }

  const contentLength = request.headers.get("content-length");
  if (contentLength && parseInt(contentLength) > 10_000) {
    return jsonResponse({ error: "Request too large" }, 413);
  }

  let body;
  try {
    body = await request.json();
  } catch {
    return jsonResponse({ error: "Invalid JSON" }, 400);
  }

  const { key, limit: customLimit, algorithm: customAlgorithm } = body;

  const sanitizedKey = sanitizeKey(key);
  if (!sanitizedKey) {
    return jsonResponse({ error: "Missing or invalid key" }, 400);
  }

  const limit = sanitizeLimit(customLimit);
  const algorithm = sanitizeAlgorithm(customAlgorithm);
  const storageKey = getStorageKey(sanitizedKey, algorithm);

  let result;
  try {
    switch (algorithm) {
      case "token_bucket":
        result = await tokenBucket(env, storageKey, limit);
        break;
      case "sliding_window":
        result = await slidingWindow(env, storageKey, limit);
        break;
      case "fixed_window":
        result = await fixedWindow(env, storageKey, limit);
        break;
      default:
        result = await tokenBucket(env, storageKey, limit);
    }
  } catch (err) {
    return jsonResponse({ error: "Failed to check rate limit" }, 500);
  }

  return new Response(JSON.stringify(result), {
    status: 200,
    headers: {
      "Content-Type": "application/json",
      "X-RateLimit-Limit": limit.toString(),
      "X-RateLimit-Remaining": result.remaining.toString(),
      "X-RateLimit-Reset": result.reset.toString(),
      "X-RateLimit-Algorithm": algorithm,
      "X-Content-Type-Options": "nosniff",
      "X-Frame-Options": "DENY"
    }
  });
}

async function handleResetLimit(key, request, env) {
  if (!env.KV) {
    return jsonResponse({ error: "Storage not configured" }, 500);
  }

  const sanitizedKey = sanitizeKey(key);
  if (!sanitizedKey) {
    return jsonResponse({ error: "Invalid key" }, 400);
  }

  const algorithm = sanitizeAlgorithm(env.ALGORITHM);
  const storageKey = getStorageKey(sanitizedKey, algorithm);

  try {
    await env.KV.delete(storageKey);
    return jsonResponse({ reset: true, key: sanitizedKey });
  } catch {
    return jsonResponse({ error: "Failed to reset limit" }, 500);
  }
}

async function tokenBucket(env, key, limit) {
  const now = Date.now();
  const bucket = await env.KV.get(key);

  if (!bucket) {
    const newBucket = {
      tokens: limit - 1,
      lastRefill: now
    };
    await env.KV.put(key, JSON.stringify(newBucket), {
      expirationTtl: WINDOW_SIZE * 2
    });
    return {
      allowed: true,
      remaining: limit - 1,
      limit,
      reset: now + (WINDOW_SIZE * 1000)
    };
  }

  let data;
  try {
    data = JSON.parse(bucket);
  } catch {
    data = { tokens: limit - 1, lastRefill: now };
  }

  const elapsed = Math.max(0, now - (data.lastRefill || now));
  const tokensPerMs = limit / (WINDOW_SIZE * 1000);
  const newTokens = Math.min(limit, (data.tokens || 0) + (elapsed * tokensPerMs));

  if (newTokens < 1) {
    const resetTime = (data.lastRefill || now) + (WINDOW_SIZE * 1000);
    return {
      allowed: false,
      remaining: 0,
      limit,
      reset: resetTime
    };
  }

  const tokensToConsume = Math.min(MAX_TOKENS_PER_REQUEST, 1);
  const remaining = Math.floor(newTokens) - tokensToConsume;

  const newBucket = {
    tokens: remaining,
    lastRefill: now
  };

  await env.KV.put(key, JSON.stringify(newBucket), {
    expirationTtl: WINDOW_SIZE * 2
  });

  return {
    allowed: true,
    remaining: Math.max(0, remaining),
    limit,
    reset: now + (WINDOW_SIZE * 1000)
  };
}

async function slidingWindow(env, key, limit) {
  const now = Date.now();
  const windowStart = now - (WINDOW_SIZE * 1000);

  const requestsData = await env.KV.get(key);
  let requestList = [];

  if (requestsData) {
    try {
      requestList = JSON.parse(requestsData);
    } catch {
      requestList = [];
    }
  }

  const recentRequests = requestList.filter(ts => typeof ts === "number" && ts > windowStart);

  if (recentRequests.length >= limit) {
    const oldestInWindow = recentRequests[0];
    return {
      allowed: false,
      remaining: 0,
      limit,
      reset: (oldestInWindow || now) + (WINDOW_SIZE * 1000)
    };
  }

  recentRequests.push(now);

  const reduced = recentRequests.slice(-limit * 2);

  await env.KV.put(key, JSON.stringify(reduced), {
    expirationTtl: WINDOW_SIZE * 2
  });

  return {
    allowed: true,
    remaining: limit - recentRequests.length,
    limit,
    reset: now + (WINDOW_SIZE * 1000)
  };
}

async function fixedWindow(env, key, limit) {
  const now = Date.now();
  const windowKey = Math.floor(now / (WINDOW_SIZE * 1000));
  const counterKey = `${key}:${windowKey}`;

  const countData = await env.KV.get(counterKey);
  let currentCount = 0;

  if (countData) {
    const parsed = parseInt(countData);
    currentCount = isNaN(parsed) ? 0 : parsed;
  }

  if (currentCount >= limit) {
    const nextWindow = (windowKey + 1) * WINDOW_SIZE * 1000;
    return {
      allowed: false,
      remaining: 0,
      limit,
      reset: nextWindow
    };
  }

  await env.KV.put(counterKey, (currentCount + 1).toString(), {
    expirationTtl: WINDOW_SIZE * 2
  });

  return {
    allowed: true,
    remaining: limit - currentCount - 1,
    limit,
    reset: (windowKey + 1) * WINDOW_SIZE * 1000
  };
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

export async function checkLimit(env, key, limit = DEFAULT_LIMIT) {
  if (!env.KV) throw new Error("Storage not configured");

  const sanitizedKey = sanitizeKey(key);
  if (!sanitizedKey) throw new Error("Invalid key");

  const sanitizedLimit = sanitizeLimit(limit);
  const storageKey = getStorageKey(sanitizedKey, ALGORITHM);

  let result;
  switch (ALGORITHM) {
    case "sliding_window":
      result = await slidingWindow(env, storageKey, sanitizedLimit);
      break;
    case "fixed_window":
      result = await fixedWindow(env, storageKey, sanitizedLimit);
      break;
    default:
      result = await tokenBucket(env, storageKey, sanitizedLimit);
  }

  return result;
}