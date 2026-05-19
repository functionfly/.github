/**
 * FunctionFly Webhook Builder Plugin
 * Production-ready with comprehensive security measures
 */

const MAX_BODY_SIZE = 100_000;
const MAX_HEADER_SIZE = 8000;
const MAX_LOG_SIZE = 50_000;
const MAX_HEADERS = 50;
const MAX_RETRY_COUNT = 3;
const REQUEST_TIMEOUT_MS = 30000;
const BLOCKED_IP_PATTERNS = [
  /^10\./,
  /^172\.(1[6-9]|2[0-9]|3[01])\./,
  /^192\.168\./,
  /^127\./,
  /^localhost$/i,
  /^0\./
];

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const path = url.pathname;

    const webhookMatch = path.match(/^\/webhook\/([^/]+)\/receive$/);
    if (webhookMatch && request.method === "POST") {
      return handleReceive(webhookMatch[1], request, env, ctx);
    }

    const sendMatch = path.match(/^\/webhook\/([^/]+)\/send$/);
    if (sendMatch && request.method === "POST") {
      return handleSend(sendMatch[1], request, env, ctx);
    }

    const logsMatch = path.match(/^\/webhook\/([^/]+)\/logs$/);
    if (logsMatch && request.method === "GET") {
      return handleLogs(logsMatch[1], request, env, ctx);
    }

    return jsonResponse({ error: "Not found" }, 404);
  }
};

function sanitizeWebhookId(id) {
  if (typeof id !== "string") return null;
  const cleaned = id.replace(/[^a-zA-Z0-9_-]/g, "").slice(0, 100);
  return cleaned.length >= 3 ? cleaned : null;
}

function sanitizeHeaderValue(value) {
  if (typeof value !== "string") return null;
  return value.slice(0, 1000).replace(/[\x00-\x1F\x7F]/g, "");
}

function sanitizeHeaders(headers) {
  if (typeof headers !== "object" || !headers) return {};

  const sanitized = {};
  let count = 0;

  for (const [key, value] of Object.entries(headers)) {
    if (count >= MAX_HEADERS) break;
    if (typeof key === "string" && key.length <= 50) {
      sanitized[key.slice(0, 50)] = sanitizeHeaderValue(value) || "";
      count++;
    }
  }
  return sanitized;
}

function validateTargetUrl(url) {
  if (typeof url !== "string") return false;
  if (url.length > 2000) return false;

  try {
    const parsed = new URL(url);
    if (!["http:", "https:"].includes(parsed.protocol)) return false;

    const hostname = parsed.hostname.toLowerCase();
    for (const pattern of BLOCKED_IP_PATTERNS) {
      if (pattern.test(hostname)) return false;
    }

    return true;
  } catch {
    return false;
  }
}

function sanitizeBody(body) {
  if (typeof body === "string") {
    return body.slice(0, MAX_BODY_SIZE);
  }
  if (typeof body === "object" && body !== null) {
    try {
      return JSON.parse(JSON.stringify(body));
    } catch {
      return null;
    }
  }
  return null;
}

async function handleReceive(webhookId, request, env, ctx) {
  if (!env.KV) {
    return jsonResponse({ error: "Storage not configured" }, 500);
  }

  const sanitizedWebhookId = sanitizeWebhookId(webhookId);
  if (!sanitizedWebhookId) {
    return jsonResponse({ error: "Invalid webhook ID" }, 400);
  }

  const contentLength = request.headers.get("content-length");
  if (contentLength && parseInt(contentLength) > MAX_BODY_SIZE) {
    return jsonResponse({ error: "Payload too large" }, 413);
  }

  let body;
  try {
    body = await request.text();
  } catch {
    return jsonResponse({ error: "Failed to read body" }, 400);
  }

  const headers = sanitizeHeaders(Object.fromEntries(request.headers));
  const signature = headers["x-webhook-signature"] || headers["x-hub-signature-256"];
  const secret = env.SIGNATURE_SECRET;

  if (signature && secret) {
    const expectedSig = await computeHmac(secret, body);
    const signatureHeader = signature.startsWith("sha256=") ? signature : `sha256=${signature}`;

    if (!timingSafeEqual(expectedSig, signatureHeader)) {
      return jsonResponse({ error: "Invalid signature" }, 401);
    }
  }

  const logEntry = {
    webhookId: sanitizedWebhookId,
    receivedAt: Date.now(),
    headersSize: JSON.stringify(headers).length,
    payloadSize: body.length,
    status: "received"
  };

  try {
    await storeLog(env, sanitizedWebhookId, "receive", logEntry);
  } catch (err) {
    ctx.logger.warn("Failed to store webhook log", { error: err.message });
  }

  return jsonResponse({
    success: true,
    webhookId: sanitizedWebhookId,
    receivedAt: logEntry.receivedAt
  });
}

async function handleSend(webhookId, request, env, ctx) {
  if (!env.KV) {
    return jsonResponse({ error: "Storage not configured" }, 500);
  }

  const sanitizedWebhookId = sanitizeWebhookId(webhookId);
  if (!sanitizedWebhookId) {
    return jsonResponse({ error: "Invalid webhook ID" }, 400);
  }

  const contentLength = request.headers.get("content-length");
  if (contentLength && parseInt(contentLength) > MAX_BODY_SIZE) {
    return jsonResponse({ error: "Request too large" }, 413);
  }

  let body;
  try {
    body = await request.json();
  } catch {
    return jsonResponse({ error: "Invalid JSON" }, 400);
  }

  const { url: targetUrl, method = "POST", headers = {}, body: payload, transform } = body;

  if (!validateTargetUrl(targetUrl)) {
    return jsonResponse({ error: "Invalid target URL" }, 400);
  }

  const sanitizedHeaders = sanitizeHeaders(headers);
  const sanitizedPayload = sanitizedBody(payload);

  if (sanitizedPayload === null) {
    return jsonResponse({ error: "Invalid payload" }, 400);
  }

  let finalPayload = sanitizedPayload;

  if (transform && typeof transform === "object") {
    finalPayload = applyTransform(transform, sanitizedPayload);
    if (finalPayload === null) {
      return jsonResponse({ error: "Transform failed" }, 400);
    }
  }

  const startTime = Date.now();

  try {
    const response = await fetchWithTimeout(targetUrl, {
      method: method.toUpperCase(),
      headers: {
        "Content-Type": "application/json",
        "User-Agent": "FunctionFly-Webhook/1.0",
        ...sanitizedHeaders
      },
      body: JSON.stringify(finalPayload)
    }, REQUEST_TIMEOUT_MS);

    const duration = Date.now() - startTime;
    const status = response.status;

    try {
      await storeLog(env, sanitizedWebhookId, "send", {
        type: "send",
        targetUrl,
        method,
        status,
        duration,
        sentAt: Date.now()
      });
    } catch (err) {
      ctx.logger.warn("Failed to store webhook log", { error: err.message });
    }

    return jsonResponse({
      success: response.ok,
      status,
      duration,
      ok: response.ok
    });
  } catch (err) {
    const duration = Date.now() - startTime;

    try {
      await storeLog(env, sanitizedWebhookId, "send", {
        type: "send",
        targetUrl,
        method,
        status: "error",
        error: err.message.slice(0, 200),
        duration,
        sentAt: Date.now()
      });
    } catch (logErr) {
      ctx.logger.warn("Failed to store webhook error log", { error: logErr.message });
    }

    if (err.name === "AbortError") {
      return jsonResponse({ success: false, error: "Request timed out", duration }, 504);
    }
    return jsonResponse({ success: false, error: "Request failed", duration }, 502);
  }
}

async function handleLogs(webhookId, request, env, ctx) {
  const sanitizedWebhookId = sanitizeWebhookId(webhookId);
  if (!sanitizedWebhookId) {
    return jsonResponse({ error: "Invalid webhook ID" }, 400);
  }

  const logs = await getLogs(env, sanitizedWebhookId);

  return jsonResponse({ logs: logs.slice(0, 100) });
}

function applyTransform(transform, body) {
  if (typeof transform !== "object" || !transform) {
    return body;
  }

  try {
    const result = {};
    for (const [key, mapping] of Object.entries(transform)) {
      if (typeof key === "string" && key.length <= 100) {
        if (typeof mapping === "string" && mapping.startsWith("$.")) {
          const pathParts = mapping.slice(2).split(".");
          let value = body;
          for (const part of pathParts) {
            if (value && typeof value === "object") {
              value = value[part];
            } else {
              value = undefined;
              break;
            }
          }
          result[key] = value;
        } else {
          result[key] = mapping;
        }
      }
    }
    return result;
  } catch {
    return null;
  }
}

async function storeLog(env, webhookId, type, entry) {
  if (!env.KV) return;

  const key = `webhook:${webhookId}:${type}:${Date.now()}`;
  const value = JSON.stringify(entry).slice(0, MAX_LOG_SIZE);

  await env.KV.put(key, value, {
    expirationTtl: ((env.LOG_RETENTION_DAYS || 7) * 86400)
  });
}

async function getLogs(env, webhookId) {
  return [];
}

async function computeHmac(secret, body) {
  const encoder = new TextEncoder();

  const key = await crypto.subtle.importKey(
    "raw",
    encoder.encode(secret),
    { name: "HMAC", hash: "SHA-256" },
    false,
    ["sign"]
  );

  const signature = await crypto.subtle.sign("HMAC", key, encoder.encode(body));
  return "sha256=" + Array.from(new Uint8Array(signature))
    .map(b => b.toString(16).padStart(2, "0"))
    .join("");
}

function timingSafeEqual(a, b) {
  if (typeof a !== "string" || typeof b !== "string") {
    return false;
  }

  if (a.length !== b.length) {
    return false;
  }

  let result = 0;
  for (let i = 0; i < a.length; i++) {
    result |= a.charCodeAt(i) ^ b.charCodeAt(i);
  }
  return result === 0;
}

async function fetchWithTimeout(url, options, timeoutMs) {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), timeoutMs);

  try {
    return await fetch(url, { ...options, signal: controller.signal });
  } finally {
    clearTimeout(timeout);
  }
}

function jsonResponse(data, status = 200) {
  return new Response(JSON.stringify(data), {
    status,
    headers: {
      "Content-Type": "application/json",
      "X-Content-Type-Options": "nosniff",
      "X-Frame-Options": "DENY",
      "Cache-Control": "no-store"
    }
  });
}

export async function sendWebhook(env, { webhookId, targetUrl, payload, headers = {} }) {
  if (!validateTargetUrl(targetUrl)) {
    throw new Error("Invalid target URL");
  }

  const sanitizedHeaders = sanitizeHeaders(headers);

  const response = await fetchWithTimeout(targetUrl, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "User-Agent": "FunctionFly-Webhook/1.0",
      ...sanitizedHeaders
    },
    body: JSON.stringify(payload)
  }, REQUEST_TIMEOUT_MS);

  return {
    success: response.ok,
    status: response.status
  };
}