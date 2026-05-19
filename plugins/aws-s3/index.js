/**
 * FunctionFly AWS S3 Storage Plugin
 * Production-ready with comprehensive security measures
 */

import {
  S3Client,
  PutObjectCommand,
  GetObjectCommand,
  DeleteObjectCommand
} from "@aws-sdk/client-s3";
import { getSignedUrl } from "@aws-sdk/s3-request-presigner";

const MAX_KEY_LENGTH = 1024;
const MAX_BUCKET_NAME_LENGTH = 63;
const MAX_SIGNED_URL_EXPIRY = 86400;
const VALID_REGIONS = new Set([
  "us-east-1", "us-east-2", "us-west-1", "us-west-2",
  "eu-west-1", "eu-west-2", "eu-west-3", "eu-central-1",
  "ap-northeast-1", "ap-northeast-2", "ap-southeast-1", "ap-southeast-2",
  "ap-south-1", "sa-east-1", "ca-central-1"
]);
const BLOCKED_EXTENSIONS = new Set(["exe", "dll", "bat", "cmd", "sh", "ps1"]);
const REQUEST_TIMEOUT_MS = 30000;

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const path = url.pathname;

    if (path === "/upload" && request.method === "POST") {
      return handleUpload(request, env);
    }

    if (path === "/download" && request.method === "GET") {
      return handleDownload(request, env);
    }

    if (path === "/signed-url" && request.method === "GET") {
      return handleSignedUrl(request, env);
    }

    return jsonResponse({ error: "Not found" }, 404);
  }
};

function validateCredentials(env) {
  if (!env.AWS_ACCESS_KEY_ID || !env.AWS_SECRET_ACCESS_KEY) {
    return false;
  }
  if (typeof env.AWS_ACCESS_KEY_ID !== "string" || env.AWS_ACCESS_KEY_ID.length < 16) {
    return false;
  }
  if (typeof env.AWS_SECRET_ACCESS_KEY !== "string" || env.AWS_SECRET_ACCESS_KEY.length < 16) {
    return false;
  }
  return true;
}

function validateRegion(region) {
  if (!region) return "us-east-1";
  return VALID_REGIONS.has(region) ? region : null;
}

function getS3Client(env) {
  const region = validateRegion(env.REGION) || "us-east-1";

  return new S3Client({
    region,
    credentials: {
      accessKeyId: env.AWS_ACCESS_KEY_ID,
      secretAccessKey: env.AWS_SECRET_ACCESS_KEY
    },
    maxAttempts: 3
  });
}

function sanitizeBucketName(name) {
  if (typeof name !== "string") return null;

  const cleaned = name.toLowerCase()
    .replace(/[^a-z0-9.-]/g, "")
    .slice(0, MAX_BUCKET_NAME_LENGTH);

  if (cleaned.length < 3 || cleaned.length > MAX_BUCKET_NAME_LENGTH) {
    return null;
  }

  if (cleaned.startsWith("xn--") || cleaned.endsWith("-s3alias")) {
    return null;
  }

  return cleaned;
}

function sanitizeKey(key) {
  if (typeof key !== "string") return null;

  if (key.includes("..") || key.startsWith("/")) {
    return null;
  }

  const ext = key.split(".").pop()?.toLowerCase();
  if (ext && BLOCKED_EXTENSIONS.has(ext)) {
    return null;
  }

  return key.slice(0, MAX_KEY_LENGTH);
}

function sanitizeContentType(type) {
  if (typeof type !== "string") return "application/octet-stream";

  const validTypes = [
    "application/json", "application/xml", "text/plain", "text/html",
    "application/octet-stream", "image/jpeg", "image/png", "image/gif",
    "image/webp", "video/mp4", "audio/mpeg", "application/pdf",
    "application/zip", "application/gzip"
  ];

  return validTypes.includes(type) ? type : "application/octet-stream";
}

async function handleUpload(request, env) {
  if (!validateCredentials(env)) {
    return jsonResponse({ error: "Invalid AWS configuration" }, 500);
  }

  const contentLength = request.headers.get("content-length");
  if (contentLength && parseInt(contentLength) > 100_000_000) {
    return jsonResponse({ error: "File too large (max 100MB)" }, 413);
  }

  let body;
  try {
    body = await request.json();
  } catch {
    return jsonResponse({ error: "Invalid JSON" }, 400);
  }

  const { bucket, key, body: fileBody, contentType, metadata = {} } = body;

  const bucketName = sanitizeBucketName(bucket || env.BUCKET_NAME);
  if (!bucketName) {
    return jsonResponse({ error: "Invalid bucket name" }, 400);
  }

  const sanitizedKey = sanitizeKey(key);
  if (!sanitizedKey) {
    return jsonResponse({ error: "Invalid key" }, 400);
  }

  if (!fileBody || typeof fileBody !== "string") {
    return jsonResponse({ error: "Invalid file body" }, 400);
  }

  const sanitizedContentType = sanitizeContentType(contentType);
  const sanitizedMetadata = sanitizeMetadata(metadata);

  const s3 = getS3Client(env);
  const command = new PutObjectCommand({
    Bucket: bucketName,
    Key: sanitizedKey,
    Body: Buffer.from(fileBody, "base64"),
    ContentType: sanitizedContentType,
    Metadata: sanitizedMetadata,
    ServerSideEncryption: "AES256"
  });

  try {
    await s3.send(command);
  } catch (err) {
    ctx.logger.error("S3 upload error", { error: err.message, bucket: bucketName });
    return jsonResponse({ error: "Failed to upload to S3" }, 500);
  }

  return jsonResponse({
    success: true,
    bucket: bucketName,
    key: sanitizedKey,
    uploadedAt: Date.now()
  });
}

async function handleDownload(request, env) {
  if (!validateCredentials(env)) {
    return jsonResponse({ error: "Invalid AWS configuration" }, 500);
  }

  const url = new URL(request.url);
  const bucket = url.searchParams.get("bucket");
  const key = url.searchParams.get("key");

  if (!bucket || !key) {
    return jsonResponse({ error: "Missing bucket or key" }, 400);
  }

  const bucketName = sanitizeBucketName(bucket);
  const sanitizedKey = sanitizeKey(key);

  if (!bucketName || !sanitizedKey) {
    return jsonResponse({ error: "Invalid bucket or key" }, 400);
  }

  const s3 = getS3Client(env);
  const command = new GetObjectCommand({
    Bucket: bucketName,
    Key: sanitizedKey
  });

  try {
    const signedUrl = await getSignedUrl(s3, command, { expiresIn: 3600 });
    return jsonResponse({ signedUrl, expiresAt: Date.now() + 3600000 });
  } catch (err) {
    return jsonResponse({ error: "Failed to generate signed URL" }, 500);
  }
}

async function handleSignedUrl(request, env) {
  if (!validateCredentials(env)) {
    return jsonResponse({ error: "Invalid AWS configuration" }, 500);
  }

  const url = new URL(request.url);
  const bucket = url.searchParams.get("bucket");
  const key = url.searchParams.get("key");
  const expiresIn = Math.min(parseInt(url.searchParams.get("expiresIn")) || 3600, MAX_SIGNED_URL_EXPIRY);

  if (!bucket || !key) {
    return jsonResponse({ error: "Missing bucket or key" }, 400);
  }

  const bucketName = sanitizeBucketName(bucket);
  const sanitizedKey = sanitizeKey(key);

  if (!bucketName || !sanitizedKey) {
    return jsonResponse({ error: "Invalid bucket or key" }, 400);
  }

  const s3 = getS3Client(env);
  const command = new GetObjectCommand({
    Bucket: bucketName,
    Key: sanitizedKey
  });

  try {
    const signedUrl = await getSignedUrl(s3, command, { expiresIn });
    return jsonResponse({
      signedUrl,
      expiresAt: Date.now() + (expiresIn * 1000)
    });
  } catch (err) {
    return jsonResponse({ error: "Failed to generate signed URL" }, 500);
  }
}

function sanitizeMetadata(meta) {
  if (typeof meta !== "object" || !meta) return {};

  const sanitized = {};
  for (const [key, value] of Object.entries(meta)) {
    if (typeof key === "string" && key.length <= 50 && typeof value === "string") {
      sanitized[key.slice(0, 50)] = value.slice(0, 200);
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

export async function uploadFile(env, { bucket, key, body, contentType }) {
  if (!validateCredentials(env)) {
    throw new Error("Invalid AWS configuration");
  }

  const bucketName = sanitizeBucketName(bucket || env.BUCKET_NAME);
  const sanitizedKey = sanitizeKey(key);

  if (!bucketName || !sanitizedKey) {
    throw new Error("Invalid bucket or key");
  }

  const s3 = getS3Client(env);
  const command = new PutObjectCommand({
    Bucket: bucketName,
    Key: sanitizedKey,
    Body: body,
    ContentType: sanitizeContentType(contentType),
    ServerSideEncryption: "AES256"
  });

  await s3.send(command);
  return { success: true, bucket: bucketName, key: sanitizedKey };
}

export async function getSignedUrl(env, { bucket, key, expiresIn = 3600 }) {
  if (!validateCredentials(env)) {
    throw new Error("Invalid AWS configuration");
  }

  const bucketName = sanitizeBucketName(bucket || env.BUCKET_NAME);
  const sanitizedKey = sanitizeKey(key);

  if (!bucketName || !sanitizedKey) {
    throw new Error("Invalid bucket or key");
  }

  const s3 = getS3Client(env);
  const command = new GetObjectCommand({
    Bucket: bucketName,
    Key: sanitizedKey
  });

  return getSignedUrl(s3, command, { expiresIn });
}