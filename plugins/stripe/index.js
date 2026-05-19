/**
 * FunctionFly Stripe Billing Plugin
 * Production-ready with comprehensive security measures
 */

const STRIPE_API_BASE = "https://api.stripe.com";
const MAX_AMOUNT = 999999999;
const MIN_AMOUNT = 0.5;
const REQUEST_TIMEOUT_MS = 10000;
const VALID_CURRENCIES = new Set(["usd", "eur", "gbp", "cad", "aud", "jpy"]);

let stripeClient = null;

function getStripeClient(secretKey) {
  if (!secretKey) return null;

  if (!stripeClient || stripeClient._secretKey !== secretKey) {
    const Stripe = require("stripe");
    stripeClient = new Stripe(secretKey, {
      apiVersion: "2024-11-20.acacia",
      timeout: REQUEST_TIMEOUT_MS,
      maxNetworkRetries: 2
    });
    stripeClient._secretKey = secretKey;
  }
  return stripeClient;
}

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);

    if (url.pathname === "/webhook/stripe" && request.method === "POST") {
      return handleWebhook(request, env);
    }

    return jsonResponse({ error: "Not found" }, 404);
  }
};

async function handleWebhook(request, env) {
  if (!env.STRIPE_SECRET_KEY) {
    return jsonResponse({ error: "Stripe not configured" }, 500);
  }

  if (!env.WEBHOOK_SECRET) {
    return jsonResponse({ error: "Webhook secret not configured" }, 500);
  }

  const signature = request.headers.get("stripe-signature");
  if (!signature) {
    return jsonResponse({ error: "Missing signature" }, 400);
  }

  const contentLength = request.headers.get("content-length");
  if (contentLength && parseInt(contentLength) > 65536) {
    return jsonResponse({ error: "Payload too large" }, 413);
  }

  let body;
  try {
    body = await request.text();
  } catch {
    return jsonResponse({ error: "Failed to read body" }, 400);
  }

  if (!body || body.length === 0) {
    return jsonResponse({ error: "Empty body" }, 400);
  }

  const stripe = getStripeClient(env.STRIPE_SECRET_KEY);

  let event;
  try {
    event = stripe.webhooks.constructEvent(body, signature, env.WEBHOOK_SECRET);
  } catch (err) {
    const errorMessage = err.message.includes("No signatures found") ||
                         err.message.includes("signature does not match")
      ? "Invalid signature"
      : "Webhook processing failed";

    return jsonResponse({ error: errorMessage }, 400);
  }

  const validEventTypes = new Set([
    "payment_intent.succeeded",
    "payment_intent.payment_failed",
    "payment_intent.processing",
    "payment_intent.canceled",
    "customer.subscription.created",
    "customer.subscription.updated",
    "customer.subscription.deleted",
    "customer.subscription.trial_will_end",
    "invoice.paid",
    "invoice.payment_failed",
    "invoice.finalized"
  ]);

  if (!validEventTypes.has(event.type)) {
    return jsonResponse({ received: true, skipped: true });
  }

  try {
    await processEvent(event, env, ctx);
  } catch (err) {
    ctx.waitUntil(logError(env, `Event processing failed: ${err.message}`, event.id));
    return jsonResponse({ received: true, error: "Processing failed" }, 500);
  }

  return jsonResponse({ received: true });
}

async function processEvent(event, env, ctx) {
  const { id, type, data } = event;

  switch (type) {
    case "payment_intent.succeeded":
      await handlePaymentSuccess(data.object, env, ctx);
      break;
    case "payment_intent.payment_failed":
      await handlePaymentFailed(data.object, env, ctx);
      break;
    case "customer.subscription.updated":
      await handleSubscriptionUpdated(data.object, env, ctx);
      break;
    case "customer.subscription.deleted":
      await handleSubscriptionDeleted(data.object, env, ctx);
      break;
    case "invoice.paid":
      await handleInvoicePaid(data.object, env, ctx);
      break;
    case "invoice.payment_failed":
      await handleInvoicePaymentFailed(data.object, env, ctx);
      break;
  }
}

async function handlePaymentSuccess(paymentIntent, env, ctx) {
  const safeId = sanitizeId(paymentIntent.id);
  ctx.logger.info(`Payment succeeded: ${safeId}`, {
    amount: paymentIntent.amount,
    currency: paymentIntent.currency,
    customerId: sanitizeId(paymentIntent.customer)
  });
}

async function handlePaymentFailed(paymentIntent, env, ctx) {
  const safeId = sanitizeId(paymentIntent.id);
  ctx.logger.warn(`Payment failed: ${safeId}`, {
    customerId: sanitizeId(paymentIntent.customer),
    lastError: paymentIntent.last_payment_error?.message
  });
}

async function handleSubscriptionUpdated(subscription, env, ctx) {
  ctx.logger.info(`Subscription updated: ${sanitizeId(subscription.id)}`, {
    status: subscription.status,
    customerId: sanitizeId(subscription.customer)
  });
}

async function handleSubscriptionDeleted(subscription, env, ctx) {
  ctx.logger.info(`Subscription deleted: ${sanitizeId(subscription.id)}`);
}

async function handleInvoicePaid(invoice, env, ctx) {
  ctx.logger.info(`Invoice paid: ${sanitizeId(invoice.id)}`, {
    amount: invoice.amount_paid,
    customerId: sanitizeId(invoice.customer)
  });
}

async function handleInvoicePaymentFailed(invoice, env, ctx) {
  ctx.logger.warn(`Invoice payment failed: ${sanitizeId(invoice.id)}`, {
    customerId: sanitizeId(invoice.customer),
    attemptCount: invoice.attempt_count
  });
}

async function logError(env, message, eventId) {
  ctx.logger.error(message, { eventId });
}

function sanitizeId(id) {
  if (typeof id !== "string") return "unknown";
  if (id.startsWith("pi_") || id.startsWith("sub_") || id.startsWith("in_") || id.startsWith("cu_")) {
    return id.slice(0, 14) + "...";
  }
  return id.slice(0, 20);
}

function sanitizeAmount(amount, currency) {
  if (typeof amount !== "number" || isNaN(amount)) return 0;
  if (amount < MIN_AMOUNT * 100 || amount > MAX_AMOUNT) return null;
  if (!VALID_CURRENCIES.has(currency)) return null;
  return amount;
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

export async function createPayment(env, { amount, currency = "usd", customer, metadata = {} }) {
  if (!env.STRIPE_SECRET_KEY) {
    throw new Error("Stripe not configured");
  }

  const sanitizedAmount = sanitizeAmount(Math.round(amount * 100), currency);
  if (sanitizedAmount === null) {
    throw new Error("Invalid amount or currency");
  }

  if (!customer || typeof customer !== "string") {
    throw new Error("Invalid customer ID");
  }

  const sanitizedMetadata = sanitizeMetadata(metadata);

  const stripe = getStripeClient(env.STRIPE_SECRET_KEY);

  const paymentIntent = await stripe.paymentIntents.create({
    amount: sanitizedAmount,
    currency: currency.toLowerCase(),
    customer: customer,
    metadata: sanitizedMetadata,
    automatic_payment_methods: {
      enabled: true
    }
  });

  return {
    id: paymentIntent.id,
    client_secret: paymentIntent.client_secret,
    status: paymentIntent.status
  };
}

export async function createSubscription(env, { customer, priceId, metadata = {} }) {
  if (!env.STRIPE_SECRET_KEY) {
    throw new Error("Stripe not configured");
  }

  if (!customer || typeof customer !== "string") {
    throw new Error("Invalid customer ID");
  }

  if (!priceId || typeof priceId !== "string") {
    throw new Error("Invalid price ID");
  }

  const sanitizedMetadata = sanitizeMetadata(metadata);

  const stripe = getStripeClient(env.STRIPE_SECRET_KEY);

  const subscription = await stripe.subscriptions.create({
    customer,
    items: [{ price: priceId }],
    metadata: sanitizedMetadata,
    payment_behavior: "default_incomplete",
    expand: ["latest_invoice.payment_intent"]
  });

  return {
    id: subscription.id,
    status: subscription.status,
    latestInvoiceId: subscription.latest_invoice?.id
  };
}

function sanitizeMetadata(meta) {
  if (typeof meta !== "object" || !meta) return {};

  const sanitized = {};
  for (const [key, value] of Object.entries(meta)) {
    if (typeof key === "string" && key.length <= 40) {
      sanitized[key.slice(0, 40)] = typeof value === "string" ? value.slice(0, 500) : String(value).slice(0, 500);
    }
  }
  return sanitized;
}