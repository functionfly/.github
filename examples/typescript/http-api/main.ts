/**
 * HTTP API TypeScript Function
 *
 * Demonstrates building a RESTful API with proper routing,
 * request parsing, and response formatting.
 *
 * @module http-api
 */

import type {
  Context,
  Env,
  Handler,
  Request,
  Response,
} from "./types/functionfly";

/**
 * In-memory data store for demonstration
 */
const users = new Map<string, { id: string; name: string; email: string }>();

// Initialize with some sample data
users.set("1", { id: "1", name: "Alice", email: "alice@example.com" });
users.set("2", { id: "2", name: "Bob", email: "bob@example.com" });

/**
 * Parse URL and extract path and query params (no global URL dependency)
 */
function parsePath(url: string): {
  path: string;
  params: Record<string, string>;
} {
  const params: Record<string, string> = {};
  const [pathPart, searchPart] = url.split("?", 2);
  const path = pathPart || "/";

  if (searchPart) {
    for (const pair of searchPart.split("&")) {
      const eq = pair.indexOf("=");
      const key = eq === -1 ? pair : pair.slice(0, eq);
      const value = eq === -1 ? "" : decodeURIComponent(pair.slice(eq + 1));
      if (key.startsWith(":")) {
        params[key.slice(1)] = value;
      }
    }
  }

  return { path, params };
}

/**
 * GET /users - List all users
 */
async function listUsers(_req: Request, _ctx: Context): Promise<Response> {
  const allUsers = Array.from(users.values());

  return {
    status: 200,
    headers: { "Content-Type": "application/json" },
    body: {
      data: allUsers,
      total: allUsers.length,
    },
  };
}

/**
 * GET /users/:id - Get a specific user
 */
async function getUser(
  _req: Request,
  _ctx: Context,
  id: string,
): Promise<Response> {
  const user = users.get(id);

  if (!user) {
    return {
      status: 404,
      headers: { "Content-Type": "application/json" },
      body: { error: "User not found" },
    };
  }

  return {
    status: 200,
    headers: { "Content-Type": "application/json" },
    body: { data: user },
  };
}

/**
 * POST /users - Create a new user
 */
async function createUser(req: Request, _ctx: Context): Promise<Response> {
  let body: { name?: string; email?: string };

  try {
    body = await req.json();
  } catch {
    return {
      status: 400,
      headers: { "Content-Type": "application/json" },
      body: { error: "Invalid JSON body" },
    };
  }

  if (!body.name || !body.email) {
    return {
      status: 400,
      headers: { "Content-Type": "application/json" },
      body: { error: "Name and email are required" },
    };
  }

  const id = String(users.size + 1);
  const user = { id, name: body.name, email: body.email };

  users.set(id, user);

  return {
    status: 201,
    headers: { "Content-Type": "application/json" },
    body: { data: user },
  };
}

/**
 * PUT /users/:id - Update a user
 */
async function updateUser(
  req: Request,
  _ctx: Context,
  id: string,
): Promise<Response> {
  const existing = users.get(id);

  if (!existing) {
    return {
      status: 404,
      headers: { "Content-Type": "application/json" },
      body: { error: "User not found" },
    };
  }

  let body: { name?: string; email?: string };

  try {
    body = await req.json();
  } catch {
    return {
      status: 400,
      headers: { "Content-Type": "application/json" },
      body: { error: "Invalid JSON body" },
    };
  }

  const updated = {
    id,
    name: body.name ?? existing.name,
    email: body.email ?? existing.email,
  };

  users.set(id, updated);

  return {
    status: 200,
    headers: { "Content-Type": "application/json" },
    body: { data: updated },
  };
}

/**
 * DELETE /users/:id - Delete a user
 */
async function deleteUser(
  _req: Request,
  _ctx: Context,
  id: string,
): Promise<Response> {
  const deleted = users.delete(id);

  if (!deleted) {
    return {
      status: 404,
      headers: { "Content-Type": "application/json" },
      body: { error: "User not found" },
    };
  }

  return {
    status: 204,
    headers: { "Content-Type": "application/json" },
    body: "",
  };
}

/**
 * Health check endpoint
 */
async function healthCheck(_req: Request, ctx: Context): Promise<Response> {
  return {
    status: 200,
    headers: { "Content-Type": "application/json" },
    body: {
      status: "healthy",
      version: ctx.env.API_VERSION ?? "v1",
      timestamp: Date.now(),
    },
  };
}

/**
 * Not found handler
 */
async function notFound(_req: Request, ctx: Context): Promise<Response> {
  return {
    status: 404,
    headers: { "Content-Type": "application/json" },
    body: {
      error: "Not found",
      path: ctx.request.url,
      method: ctx.request.method,
    },
  };
}

/**
 * Route the request to the appropriate handler
 */
async function routeRequest(req: Request, ctx: Context): Promise<Response> {
  const { path } = parsePath(req.url);
  const method = req.method.toUpperCase();

  // Health check
  if (method === "GET" && path === "/health") {
    return healthCheck(req, ctx);
  }

  // User routes
  if (path === "/users") {
    if (method === "GET") {
      return listUsers(req, ctx);
    }
    if (method === "POST") {
      return createUser(req, ctx);
    }
  }

  // User by ID routes
  const userIdMatch = path.match(/^\/users\/(.+)$/);
  if (userIdMatch) {
    const id = userIdMatch[1];

    if (method === "GET") {
      return getUser(req, ctx, id);
    }
    if (method === "PUT" || method === "PATCH") {
      return updateUser(req, ctx, id);
    }
    if (method === "DELETE") {
      return deleteUser(req, ctx, id);
    }
  }

  // API info endpoint
  if (method === "GET" && path === "/") {
    return {
      status: 200,
      headers: { "Content-Type": "application/json" },
      body: {
        name: "FunctionFly HTTP API",
        version: ctx.env.API_VERSION ?? "v1",
        endpoints: [
          "GET /",
          "GET /health",
          "GET /users",
          "POST /users",
          "GET /users/:id",
          "PUT /users/:id",
          "DELETE /users/:id",
        ],
      },
    };
  }

  return notFound(req, ctx);
}

/**
 * Main handler function for the http-api function.
 *
 * This demonstrates a RESTful API with proper routing,
 * method handling, and JSON responses.
 *
 * @param request - The incoming HTTP request
 * @param env - Environment variables
 * @param context - Function execution context
 * @returns Response with API result
 */
const handler: Handler = async (
  request: Request,
  _env: Env,
  context: Context,
): Promise<Response> => {
  try {
    return await routeRequest(request, context);
  } catch (error) {
    const message = error instanceof Error ? error.message : "Unknown error";

    return {
      status: 500,
      headers: { "Content-Type": "application/json" },
      body: {
        error: "Internal server error",
        message,
      },
    };
  }
};

export { handler };
export default handler;
