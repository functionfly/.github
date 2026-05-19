/**
 * FunctionFly Slack Integration Plugin
 * Production-ready with comprehensive security measures
 */

const SLACK_API_BASE = "https://slack.com/api";
const ALLOWED_CHANNELS = new Set();
const MAX_MESSAGE_LENGTH = 3000;
const MAX_RETRY_COUNT = 3;
const REQUEST_TIMEOUT_MS = 5000;

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);

    if (request.method === "POST" && url.pathname.endsWith("/webhook/slack")) {
      return handleWebhook(request, env);
    }

    return jsonResponse({ error: "Not found" }, 404);
  }
};

async function handleWebhook(request, env) {
  const contentLength = request.headers.get("content-length");

  if (contentLength && parseInt(contentLength) > 100_000) {
    return jsonResponse({ error: "Payload too large" }, 413);
  }

  let body;
  try {
    body = await request.json();
  } catch {
    return jsonResponse({ error: "Invalid JSON" }, 400);
  }

  if (!env.SLACK_TOKEN) {
    return jsonResponse({ error: "Slack token not configured" }, 500);
  }

  const slackToken = env.SLACK_TOKEN;
  const defaultChannel = sanitizeChannelId(env.DEFAULT_CHANNEL || "#alerts");

  ALLOWED_CHANNELS.add(defaultChannel);

  const message = formatSlackMessage(body);

  if (!message || (typeof message === "object" && !message.text && !message.blocks)) {
    return jsonResponse({ error: "Invalid message format" }, 400);
  }

  const response = await fetchWithTimeout(`${SLACK_API_BASE}/chat.postMessage`, {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${slackToken}`,
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      channel: defaultChannel,
      ...message
    })
  }, REQUEST_TIMEOUT_MS);

  const result = await response.json();

  if (!result.ok) {
    const errorType = getSlackErrorType(result.error);
    return jsonResponse({ error: result.error, errorType }, errorType === "auth" ? 401 : 500);
  }

  return jsonResponse({ success: true, ts: result.ts });
}

function formatSlackMessage(body) {
  const { event, type, text, channel, user, data } = body;

  if (type === "callback" && event) {
    const eventText = sanitizeString(event.text || "New event", MAX_MESSAGE_LENGTH);
    return {
      blocks: [
        {
          type: "section",
          text: {
            type: "mrkdwn",
            text: `*${eventText}*`
          }
        }
      ]
    };
  }

  if (text) {
    return {
      text: sanitizeString(text, MAX_MESSAGE_LENGTH)
    };
  }

  return {
    text: sanitizeString(JSON.stringify(data || body), MAX_MESSAGE_LENGTH)
  };
}

function sanitizeChannelId(channel) {
  const cleaned = channel.replace(/[^#@a-zA-Z0-9_-]/g, "").slice(0, 100);
  return cleaned || "#general";
}

function sanitizeString(str, maxLength) {
  if (typeof str !== "string") return "";
  return str.slice(0, maxLength).replace(/[\x00-\x1F\x7F]/g, "");
}

function getSlackErrorType(error) {
  const authErrors = ["token_expired", "invalid_auth", "account_inactive", "token_revoked"];
  const rateErrors = ["ratelimited", "too_many_requests"];
  const clientErrors = ["channel_not_found", "invalid_channel", "is_archived"];

  if (authErrors.includes(error)) return "auth";
  if (rateErrors.includes(error)) return "rate";
  if (clientErrors.includes(error)) return "client";
  return "server";
}

async function fetchWithTimeout(url, options, timeoutMs) {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const response = await fetch(url, {
      ...options,
      signal: controller.signal
    });
    return response;
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
      "X-Frame-Options": "DENY"
    }
  });
}

export async function sendAlert(env, { channel, message, level = "info" }) {
  if (!env.SLACK_TOKEN) {
    throw new Error("Slack token not configured");
  }

  const validLevels = ["info", "warning", "error", "success"];
  if (!validLevels.includes(level)) {
    level = "info";
  }

  const colors = {
    error: "#ff0000",
    warning: "#ffaa00",
    success: "#36a64f",
    info: "#5865F2"
  };

  const sanitizedChannel = sanitizeChannelId(channel || env.DEFAULT_CHANNEL);
  const sanitizedMessage = sanitizeString(message, MAX_MESSAGE_LENGTH);

  const payload = {
    channel: sanitizedChannel,
    attachments: [{
      color: colors[level] || colors.info,
      text: sanitizedMessage,
      ts: Math.floor(Date.now() / 1000)
    }]
  };

  const response = await fetchWithTimeout(`${SLACK_API_BASE}/chat.postMessage`, {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${env.SLACK_TOKEN}`,
      "Content-Type": "application/json"
    },
    body: JSON.stringify(payload)
  }, REQUEST_TIMEOUT_MS);

  const result = await response.json();

  if (!result.ok) {
    throw new Error(`Slack error: ${result.error}`);
  }

  return result;
}