/**
 * FunctionFly Edge Worker
 *
 * Handles wildcard subdomain routing for {slug}.functionfly.com.
 * Extracts the app slug from the subdomain and forwards the request
 * to the orchestrator API with the slug in a header.
 *
 * This runs at 300+ Cloudflare PoPs with zero cold-start, so every
 * user app gets global edge routing for free.
 */

interface Env {
  UPSTREAM_HOST: string;
  ENVIRONMENT: string;
}

// Reserved subdomains that should NOT be treated as app slugs.
const RESERVED_SUBDOMAINS = new Set([
  "www",
  "api",
  "app",
  "admin",
  "auth",
  "docs",
  "cdn",
  "edge",
  "run",
  "registry",
  "status",
  "blog",
  "mail",
  "ftp",
  "smtp",
  "pop",
  "imap",
  "ns1",
  "ns2",
  "mx",
  "staging",
  "dev",
  "test",
]);

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url);
    const hostname = url.hostname;

    // Extract subdomain: saas-starter.functionfly.com → saas-starter
    const parts = hostname.split(".");
    // Need at least 3 parts: subdomain.domain.tld
    if (parts.length < 3) {
      return new Response("Not Found", { status: 404 });
    }

    const subdomain = parts[0];

    // Skip reserved subdomains — these are handled by other services
    if (RESERVED_SUBDOMAINS.has(subdomain.toLowerCase())) {
      return new Response("Not Found", { status: 404 });
    }

    // Build upstream URL — forward to orchestrator API
    const upstream = new URL(url.pathname + url.search, `https://${env.UPSTREAM_HOST}`);

    // Clone the request with upstream target
    const headers = new Headers(request.headers);
    headers.set("X-FF-Slug", subdomain);
    headers.set("X-Forwarded-Host", hostname);
    headers.set("X-Forwarded-Proto", "https");
    // Preserve real client IP
    const clientIP = request.headers.get("CF-Connecting-IP");
    if (clientIP) {
      headers.set("X-Real-IP", clientIP);
    }

    const upstreamRequest = new Request(upstream.toString(), {
      method: request.method,
      headers,
      body: request.body,
      redirect: "follow",
    });

    try {
      const response = await fetch(upstreamRequest);

      // Clone response to add edge headers
      const newResponse = new Response(response.body, response);
      newResponse.headers.set("X-FF-Edge", "cloudflare");
      newResponse.headers.set("X-FF-Slug", subdomain);

      return newResponse;
    } catch (err) {
      return new Response(
        JSON.stringify({ error: "Edge routing error", message: String(err) }),
        {
          status: 502,
          headers: { "Content-Type": "application/json" },
        }
      );
    }
  },
};
