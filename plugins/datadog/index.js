/**
 * FunctionFly Datadog Monitoring Plugin
 * Production-ready with comprehensive security measures
 */

const DATADOG_API_BASE = "https://api.datadoghq.com";
const DATADOG_LOG_INTAKE = "https://http-intake.logs.datadoghq.com";
const VALID_SITES = new Set(["datadoghq.com", "datadoghq.eu", "us3.datadoghq.com", "ddog-gov.com"]);
const VALID_METRIC_TYPES = new Set(["gauge", "count", "histogram", "distribution"]);
const MAX_METRIC_NAME_LENGTH = 200;
const MAX_TAG_COUNT = 20;
const MAX_TAG_LENGTH = 100;
const MAX_LOG_LENGTH = 256_000;
const REQUEST_TIMEOUT_MS = 5000;

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);

    if (url.pathname === "/metric" && request.method === "POST") {
      return handleMetric(request, env);
    }

    return jsonResponse({ error: "Not found" }, 404);
  }
};

function validateApiKeys(apiKey, appKey) {
  if (!apiKey || typeof apiKey !== "string" || apiKey.length < 32) return false;
  if (appKey && (typeof appKey !== "string" || appKey.length < 40)) return false;
  return true;
}

function validateSite(site) {
  return VALID_SITES.has(site);
}

function sanitizeMetricName(name) {
  if (typeof name !== "string") return "";
  return name.replace(/[^a-zA-Z0-9_.]/g, "").slice(0, MAX_METRIC_NAME_LENGTH);
}

function sanitizeTag(tag) {
  if (typeof tag !== "string") return null;
  const cleaned = tag.replace(/[^a-zA-Z0-9_./:-]/g, "").slice(0, MAX_TAG_LENGTH);
  return cleaned.length > 0 ? cleaned : null;
}

function sanitizeValue(value) {
  if (typeof value !== "number" || isNaN(value) || !isFinite(value)) return null;
  const absValue = Math.abs(value);
  if (absValue > 1e12) return null;
  return value;
}

async function handleMetric(request, env) {
  if (!env.DATADOG_API_KEY || !validateApiKeys(env.DATADOG_API_KEY, env.DATADOG_APP_KEY)) {
    return jsonResponse({ error: "Invalid Datadog configuration" }, 500);
  }

  const site = VALID_SITES.has(env.SITE) ? env.SITE : "datadoghq.com";

  const contentLength = request.headers.get("content-length");
  if (contentLength && parseInt(contentLength) > 50_000) {
    return jsonResponse({ error: "Request payload too large" }, 413);
  }

  let body;
  try {
    body = await request.json();
  } catch {
    return jsonResponse({ error: "Invalid JSON" }, 400);
  }

  const { metric, value, tags = [], type = "gauge" } = body;

  const sanitizedMetric = sanitizeMetricName(metric);
  if (!sanitizedMetric) {
    return jsonResponse({ error: "Invalid metric name" }, 400);
  }

  if (!VALID_METRIC_TYPES.has(type)) {
    return jsonResponse({ error: "Invalid metric type" }, 400);
  }

  const sanitizedValue = sanitizeValue(value);
  if (sanitizedValue === null) {
    return jsonResponse({ error: "Invalid metric value" }, 400);
  }

  if (!Array.isArray(tags) || tags.length > MAX_TAG_COUNT) {
    return jsonResponse({ error: "Invalid tags" }, 400);
  }

  const sanitizedTags = tags
    .map(sanitizeTag)
    .filter(Boolean)
    .slice(0, MAX_TAG_COUNT);

  const point = {
    metric: sanitizedMetric,
    points: [[Math.floor(Date.now() / 1000), sanitizedValue]],
    type,
    tags: sanitizedTags
  };

  const apiKey = env.DATADOG_API_KEY;

  let response;
  try {
    response = await fetchWithTimeout(
      `${DATADOG_API_BASE}/api/v1/series`,
      {
        method: "POST",
        headers: {
          "DD-API-KEY": apiKey,
          "DD-APPLICATION-KEY": env.DATADOG_APP_KEY || "",
          "Content-Type": "application/json"
        },
        body: JSON.stringify({ series: [point] })
      },
      REQUEST_TIMEOUT_MS
    );
  } catch (err) {
    if (err.name === "AbortError") {
      return jsonResponse({ error: "Datadog API timeout" }, 504);
    }
    return jsonResponse({ error: "Failed to send metric" }, 502);
  }

  if (!response.ok) {
    const error = await response.text();
    ctx.logger.error("Datadog metric error", { status: response.status, error });
    return jsonResponse({ error: "Failed to submit metric" }, 502);
  }

  return jsonResponse({ success: true });
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

async function fetchWithTimeout(url, options, timeoutMs) {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), timeoutMs);

  try {
    return await fetch(url, { ...options, signal: controller.signal });
  } finally {
    clearTimeout(timeout);
  }
}

export async function sendMetric(env, { metric, value, tags = [], type = "gauge" }) {
  if (!env.DATADOG_API_KEY) {
    throw new Error("Datadog not configured");
  }

  const sanitizedMetric = sanitizeMetricName(metric);
  const sanitizedValue = sanitizeValue(value);

  if (!sanitizedMetric || sanitizedValue === null) {
    throw new Error("Invalid metric data");
  }

  const sanitizedTags = tags.map(sanitizeTag).filter(Boolean).slice(0, MAX_TAG_COUNT);

  const point = {
    metric: sanitizedMetric,
    points: [[Math.floor(Date.now() / 1000), sanitizedValue]],
    type: VALID_METRIC_TYPES.has(type) ? type : "gauge",
    tags: sanitizedTags
  };

  const site = VALID_SITES.has(env.SITE) ? env.SITE : "datadoghq.com";

  const response = await fetchWithTimeout(
    `${DATADOG_API_BASE}/api/v1/series`,
    {
      method: "POST",
      headers: {
        "DD-API-KEY": env.DATADOG_API_KEY,
        "DD-APPLICATION-KEY": env.DATADOG_APP_KEY || "",
        "Content-Type": "application/json"
      },
      body: JSON.stringify({ series: [point] })
    },
    REQUEST_TIMEOUT_MS
  );

  if (!response.ok) {
    throw new Error("Failed to send metric to Datadog");
  }

  return { success: true };
}

export async function sendLog(env, { message, level = "info", tags = {}, metadata = {} }) {
  if (!env.DATADOG_API_KEY) {
    throw new Error("Datadog not configured");
  }

  const validLevels = new Set(["debug", "info", "warn", "error", "critical"]);
  const sanitizedLevel = validLevels.has(level) ? level : "info";

  const sanitizedMessage = typeof message === "string"
    ? message.slice(0, MAX_LOG_LENGTH)
    : JSON.stringify(message).slice(0, MAX_LOG_LENGTH);

  const sanitizedTags = {};
  for (const [key, value] of Object.entries(tags)) {
    if (typeof key === "string" && key.length <= 50) {
      sanitizedTags[key.slice(0, 50)] = typeof value === "string" ? value.slice(0, 100) : String(value).slice(0, 100);
    }
  }

  const site = VALID_SITES.has(env.SITE) ? env.SITE : "datadoghq.com";

  const logEntry = {
    ddtags: Object.entries(sanitizedTags).map(([k, v]) => `${k}:${v}`).join(","),
    msg: sanitizedMessage,
    status: sanitizedLevel,
    ...sanitizeMetadata(metadata)
  };

  const response = await fetchWithTimeout(
    `${DATADOG_LOG_INTAKE}/v1/input`,
    {
      method: "POST",
      headers: {
        "DD-API-KEY": env.DATADOG_API_KEY,
        "Content-Type": "application/json"
      },
      body: JSON.stringify(logEntry)
    },
    REQUEST_TIMEOUT_MS
  );

  if (!response.ok) {
    throw new Error("Failed to send log to Datadog");
  }

  return { success: true };
}

function sanitizeMetadata(meta) {
  if (typeof meta !== "object" || !meta) return {};

  const sanitized = {};
  for (const [key, value] of Object.entries(meta)) {
    if (typeof key === "string" && key.length <= 50) {
      if (typeof value === "string") {
        sanitized[key] = value.slice(0, 200);
      } else if (typeof value === "number" || typeof value === "boolean") {
        sanitized[key] = value;
      }
    }
  }
  return sanitized;
}

export async function sendTrace(env, { name, duration, tags = {}, error = false }) {
  if (!env.DATADOG_API_KEY) {
    throw new Error("Datadog not configured");
  }

  const sanitizedName = typeof name === "string" ? name.replace(/[^a-zA-Z0-9_.]/g, "").slice(0, 100) : "unnamed";
  const sanitizedDuration = typeof duration === "number" && duration >= 0 && duration < 1e12 ? duration : 0;

  const traceTags = {
    ...tags,
    error: String(Boolean(error)),
    service: "functionfly-plugin"
  };

  const sanitizedTraceTags = {};
  for (const [key, value] of Object.entries(traceTags)) {
    if (typeof key === "string") {
      sanitizedTraceTags[key.slice(0, 50)] = String(value).slice(0, 100);
    }
  }

  const trace = [{
    name: sanitizedName,
    duration: sanitizedDuration,
    tags: sanitizedTraceTags
  }];

  const site = VALID_SITES.has(env.SITE) ? env.SITE : "datadoghq.com";

  const response = await fetchWithTimeout(
    `${DATADOG_API_BASE}/api/v0/traces`,
    {
      method: "POST",
      headers: {
        "DD-API-KEY": env.DATADOG_API_KEY,
        "Content-Type": "application/json"
      },
      body: JSON.stringify(trace)
    },
    REQUEST_TIMEOUT_MS
  );

  if (!response.ok) {
    throw new Error("Failed to send trace to Datadog");
  }

  return { success: true };
}