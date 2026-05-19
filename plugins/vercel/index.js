/**
 * FunctionFly Vercel Deployments Plugin
 * Production-ready with comprehensive security measures
 */

const VERCEL_API_BASE = "https://api.vercel.com";
const REQUEST_TIMEOUT_MS = 30000;
const MAX_BODY_SIZE = 100_000;
const VALID_ENVIRONMENTS = new Set(["production", "preview", "development"]);

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const path = url.pathname;

    if (path === "/deploy" && request.method === "POST") {
      return handleDeploy(request, env);
    }

    if (path === "/deployments" && request.method === "GET") {
      return handleListDeployments(request, env);
    }

    if (path === "/rollback" && request.method === "POST") {
      return handleRollback(request, env);
    }

    return jsonResponse({ error: "Not found" }, 404);
  }
};

function validateToken(token) {
  if (!token || typeof token !== "string") return false;
  return token.match(/^[A-Za-z0-9_-]{24,}$/) !== null;
}

function validateDeploymentId(id) {
  if (!id || typeof id !== "string") return false;
  return id.match(/^[a-zA-Z0-9]{20}$/) !== null;
}

function validateProjectName(name) {
  if (!name || typeof name !== "string") return true;
  return name.match(/^[a-zA-Z0-9][a-zA-Z0-9-_]{0,100}$/) !== null;
}

async function handleDeploy(request, env) {
  if (!env.VERCEL_TOKEN || !validateToken(env.VERCEL_TOKEN)) {
    return jsonResponse({ error: "Invalid Vercel configuration" }, 500);
  }

  const contentLength = request.headers.get("content-length");
  if (contentLength && parseInt(contentLength) > MAX_BODY_SIZE) {
    return jsonResponse({ error: "Request body too large" }, 413);
  }

  let body;
  try {
    body = await request.json();
  } catch {
    return jsonResponse({ error: "Invalid JSON" }, 400);
  }

  const { projectName, gitSource, targetEnv } = body;

  if (!validateProjectName(projectName || env.PROJECT_NAME)) {
    return jsonResponse({ error: "Invalid project name" }, 400);
  }

  if (targetEnv && !VALID_ENVIRONMENTS.has(targetEnv)) {
    return jsonResponse({ error: "Invalid environment" }, 400);
  }

  if (gitSource) {
    if (typeof gitSource !== "object") {
      return jsonResponse({ error: "Invalid gitSource format" }, 400);
    }
    if (gitSource.repo && !isValidRepoUrl(gitSource.repo)) {
      return jsonResponse({ error: "Invalid repository URL" }, 400);
    }
  }

  const token = env.VERCEL_TOKEN;
  const finalProjectName = sanitizeProjectName(projectName || env.PROJECT_NAME || "my-project");
  const teamId = env.TEAM_ID;
  const finalEnv = targetEnv || "production";

  let apiUrl = `${VERCEL_API_BASE}/v13/deployments`;
  const queryParams = teamId ? `?teamId=${encodeURIComponent(teamId)}` : "";

  const deployPayload = {
    name: finalProjectName,
    target: finalEnv,
    ...(gitSource && {
      gitSource: {
        type: gitSource.type || "github",
        repo: gitSource.repo,
        ref: gitSource.ref || "main"
      }
    })
  };

  let response;
  try {
    response = await fetchWithTimeout(`${apiUrl}${queryParams}`, {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${token}`,
        "Content-Type": "application/json"
      },
      body: JSON.stringify(deployPayload)
    }, REQUEST_TIMEOUT_MS);
  } catch (err) {
    if (err.name === "AbortError") {
      return jsonResponse({ error: "Deployment request timed out" }, 504);
    }
    return jsonResponse({ error: "Failed to connect to Vercel" }, 502);
  }

  const deployment = await response.json();

  if (!response.ok) {
    const errorMsg = deployment.error?.message || "Deployment failed";
    ctx.logger.error("Vercel deployment error", { error: errorMsg, status: response.status });
    return jsonResponse({ error: errorMsg }, response.status);
  }

  return jsonResponse({
    id: deployment.id,
    url: deployment.url,
    status: deployment.status,
    createdAt: deployment.createdAt
  });
}

async function handleListDeployments(request, env) {
  if (!env.VERCEL_TOKEN || !validateToken(env.VERCEL_TOKEN)) {
    return jsonResponse({ error: "Invalid Vercel configuration" }, 500);
  }

  const url = new URL(request.url);
  const limit = Math.min(parseInt(url.searchParams.get("limit")) || 20, 100);
  const since = url.searchParams.get("since");

  const token = env.VERCEL_TOKEN;
  let apiUrl = `${VERCEL_API_BASE}/v13/deployments?limit=${limit}`;
  if (env.TEAM_ID) {
    apiUrl += `&teamId=${encodeURIComponent(env.TEAM_ID)}`;
  }
  if (since) {
    const sinceTs = parseInt(since);
    if (!isNaN(sinceTs) && sinceTs > 0) {
      apiUrl += `&since=${sinceTs}`;
    }
  }

  let response;
  try {
    response = await fetchWithTimeout(apiUrl, {
      headers: { "Authorization": `Bearer ${token}` }
    }, REQUEST_TIMEOUT_MS);
  } catch (err) {
    return jsonResponse({ error: "Failed to connect to Vercel" }, 502);
  }

  const data = await response.json();

  if (!response.ok) {
    return jsonResponse({ error: data.error?.message || "Failed to list deployments" }, response.status);
  }

  return jsonResponse({
    deployments: (data.deployments || []).map(d => ({
      id: d.id,
      url: d.url,
      status: d.status,
      createdAt: d.createdAt,
      readyState: d.readyState
    }))
  });
}

async function handleRollback(request, env) {
  if (!env.VERCEL_TOKEN || !validateToken(env.VERCEL_TOKEN)) {
    return jsonResponse({ error: "Invalid Vercel configuration" }, 500);
  }

  let body;
  try {
    body = await request.json();
  } catch {
    return jsonResponse({ error: "Invalid JSON" }, 400);
  }

  const { deploymentId } = body;

  if (!deploymentId || !validateDeploymentId(deploymentId)) {
    return jsonResponse({ error: "Invalid deployment ID" }, 400);
  }

  const token = env.VERCEL_TOKEN;

  let response;
  try {
    response = await fetchWithTimeout(
      `${VERCEL_API_BASE}/v13/deployments/${deploymentId}/rollback`,
      {
        method: "POST",
        headers: { "Authorization": `Bearer ${token}` }
      }
    , REQUEST_TIMEOUT_MS);
  } catch (err) {
    return jsonResponse({ error: "Failed to connect to Vercel" }, 502);
  }

  const result = await response.json();

  if (!response.ok) {
    return jsonResponse({ error: result.error?.message || "Rollback failed" }, response.status);
  }

  return jsonResponse({
    id: result.id,
    url: result.url,
    status: result.status,
    readyState: result.readyState
  });
}

function isValidRepoUrl(url) {
  if (typeof url !== "string") return false;
  const validHosts = ["github.com", "gitlab.com", "bitbucket.org"];
  try {
    const parsed = new URL(url);
    return validHosts.includes(parsed.hostname) && parsed.protocol === "https:";
  } catch {
    return false;
  }
}

function sanitizeProjectName(name) {
  if (!name || typeof name !== "string") return "project";
  return name.replace(/[^a-zA-Z0-9-_]/g, "").slice(0, 100) || "project";
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