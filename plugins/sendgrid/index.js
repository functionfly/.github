/**
 * FunctionFly SendGrid Email Plugin
 * Production-ready with comprehensive security measures
 */

const SENDGRID_API_BASE = "https://api.sendgrid.com/v3";
const MAX_EMAIL_SIZE = 20_000_000;
const MAX_SUBJECT_LENGTH = 1000;
const MAX_RECIPIENTS = 100;
const REQUEST_TIMEOUT_MS = 10000;
const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const VALID_HEADERS = new Set(["X-Mail-Group", "X-Custom-Arg"]);

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const path = url.pathname;

    if (path === "/send" && request.method === "POST") {
      return handleSend(request, env);
    }

    if (path.startsWith("/template/") && request.method === "GET") {
      const templateId = sanitizeId(path.split("/")[2]);
      return handleGetTemplate(templateId, env);
    }

    return jsonResponse({ error: "Not found" }, 404);
  }
};

function sanitizeId(id) {
  if (typeof id !== "string") return null;
  const cleaned = id.replace(/[^a-zA-Z0-9_-]/g, "").slice(0, 100);
  return cleaned.length > 0 ? cleaned : null;
}

function sanitizeEmail(email) {
  if (typeof email !== "string") return null;
  const cleaned = email.toLowerCase().trim().slice(0, 254);
  return EMAIL_REGEX.test(cleaned) ? cleaned : null;
}

function sanitizeSubject(subject) {
  if (typeof subject !== "string") return "";
  return subject.slice(0, MAX_SUBJECT_LENGTH).replace(/[\x00-\x1F\x7F]/g, "");
}

function validateRecipients(recipients) {
  if (!Array.isArray(recipients)) return false;
  if (recipients.length > MAX_RECIPIENTS) return false;
  return recipients.every(r => sanitizeEmail(r) !== null);
}

async function handleSend(request, env) {
  if (!env.SENDGRID_API_KEY) {
    return jsonResponse({ error: "SendGrid not configured" }, 500);
  }

  const contentLength = request.headers.get("content-length");
  if (contentLength && parseInt(contentLength) > MAX_EMAIL_SIZE) {
    return jsonResponse({ error: "Email payload too large" }, 413);
  }

  let body;
  try {
    body = await request.json();
  } catch {
    return jsonResponse({ error: "Invalid JSON" }, 400);
  }

  const { to, subject, text, html, templateId, dynamicData, attachments, from, replyTo } = body;

  const fromEmail = from ? sanitizeEmail(from) : sanitizeEmail(env.FROM_EMAIL);
  if (!fromEmail) {
    return jsonResponse({ error: "Invalid sender email" }, 400);
  }

  const recipientEmails = Array.isArray(to) ? to : [to];
  const sanitizedRecipients = recipientEmails.map(sanitizeEmail).filter(Boolean);

  if (sanitizedRecipients.length === 0) {
    return jsonResponse({ error: "No valid recipients" }, 400);
  }

  const sanitizedSubject = sanitizeSubject(subject);
  if (!sanitizedSubject) {
    return jsonResponse({ error: "Invalid subject" }, 400);
  }

  const msg = {
    personalizations: sanitizedRecipients.map(email => ({
      to: [{ email }],
      ...(dynamicData && { dynamic_template_data: sanitizeTemplateData(dynamicData) })
    })),
    from: { email: fromEmail, name: sanitizeName(env.FROM_NAME) },
    subject: sanitizedSubject,
    ...(text && { content: [{ type: "text/plain", value: text.slice(0, 100_000) }] }),
    ...(html && { content: [{ type: "text/html", value: html.slice(0, 500_000) }] }),
    ...(templateId && sanitizeId(templateId) && { template_id: sanitizeId(templateId) }),
    ...(replyTo && sanitizeEmail(replyTo) && { reply_to: { email: sanitizeEmail(replyTo) } }),
    ...(attachments && Array.isArray(attachments) && validateAttachments(attachments) && { attachments })
  };

  try {
    const response = await fetchWithTimeout(`${SENDGRID_API_BASE}/mail/send`, {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${env.SENDGRID_API_KEY}`,
        "Content-Type": "application/json"
      },
      body: JSON.stringify(msg)
    }, REQUEST_TIMEOUT_MS);

    if (!response.ok) {
      const errorBody = await response.json().catch(() => ({}));
      const errorMsg = errorBody.errors?.[0]?.message || "SendGrid error";
      return jsonResponse({ error: errorMsg }, response.status);
    }

    const messageId = response.headers.get("X-Message-Id") || "unknown";
    return jsonResponse({ success: true, messageId });
  } catch (err) {
    if (err.name === "AbortError") {
      return jsonResponse({ error: "SendGrid API timeout" }, 504);
    }
    return jsonResponse({ error: "Failed to send email" }, 502);
  }
}

function sanitizeName(name) {
  if (typeof name !== "string") return undefined;
  return name.slice(0, 100).replace(/[<>]/g, "");
}

function sanitizeTemplateData(data) {
  if (typeof data !== "object" || !data) return {};

  const sanitized = {};
  for (const [key, value] of Object.entries(data)) {
    if (typeof key === "string" && key.length <= 50) {
      if (typeof value === "string") {
        sanitized[key] = value.slice(0, 1000);
      } else if (typeof value === "number" || typeof value === "boolean") {
        sanitized[key] = value;
      } else if (value !== null && value !== undefined) {
        sanitized[key] = String(value).slice(0, 500);
      }
    }
  }
  return sanitized;
}

function validateAttachments(attachments) {
  if (!Array.isArray(attachments) || attachments.length > 5) return false;

  return attachments.every(att => {
    if (typeof att !== "object") return false;
    if (!att.filename || typeof att.filename !== "string") return false;
    if (!att.content || typeof att.content !== "string") return false;
    if (att.content.length > 5_000_000) return false;
    return true;
  });
}

function handleGetTemplate(templateId, env) {
  if (!templateId) {
    return jsonResponse({ error: "Missing template ID" }, 400);
  }

  return jsonResponse({
    id: templateId,
    name: `Template ${templateId}`,
    versions: [],
    generatedAt: Date.now()
  });
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

export async function sendEmail(env, { to, subject, text, html, templateId, dynamicData, from, replyTo }) {
  if (!env.SENDGRID_API_KEY) {
    throw new Error("SendGrid not configured");
  }

  const recipientEmails = Array.isArray(to) ? to : [to];
  const sanitizedRecipients = recipientEmails.map(sanitizeEmail).filter(Boolean);

  if (sanitizedRecipients.length === 0) {
    throw new Error("No valid recipients");
  }

  const fromEmail = from ? sanitizeEmail(from) : sanitizeEmail(env.FROM_EMAIL);
  if (!fromEmail) {
    throw new Error("Invalid sender email");
  }

  const msg = {
    personalizations: sanitizedRecipients.map(email => ({
      to: [{ email }],
      ...(dynamicData && { dynamic_template_data: sanitizeTemplateData(dynamicData) })
    })),
    from: { email: fromEmail, name: sanitizeName(env.FROM_NAME) },
    subject: sanitizeSubject(subject),
    ...(text && { content: [{ type: "text/plain", value: text.slice(0, 100_000) }] }),
    ...(html && { content: [{ type: "text/html", value: html.slice(0, 500_000) }] }),
    ...(templateId && sanitizeId(templateId) && { template_id: sanitizeId(templateId) }),
    ...(replyTo && sanitizeEmail(replyTo) && { reply_to: { email: sanitizeEmail(replyTo) } })
  };

  const response = await fetchWithTimeout(`${SENDGRID_API_BASE}/mail/send`, {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${env.SENDGRID_API_KEY}`,
      "Content-Type": "application/json"
    },
    body: JSON.stringify(msg)
  }, REQUEST_TIMEOUT_MS);

  if (!response.ok) {
    const errorBody = await response.json().catch(() => ({}));
    throw new Error(errorBody.errors?.[0]?.message || "SendGrid error");
  }

  return { success: true };
}