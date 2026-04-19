import React, { useState } from "react";
import "./ApiReference.css";

const Icons = {
  Terminal: () => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <polyline points="4 17 10 11 4 5" />
      <line x1="12" x2="20" y1="19" y2="19" />
    </svg>
  ),
  Copy: () => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <rect width="14" height="14" x="8" y="8" rx="2" ry="2" />
      <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
    </svg>
  ),
  Check: () => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <polyline points="20 6 9 17 4 12" />
    </svg>
  ),
  ExternalLink: () => (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="14"
      height="14"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
      <polyline points="15 3 21 3 21 9" />
      <line x1="10" x2="21" y1="14" y2="3" />
    </svg>
  ),
};

interface Endpoint {
  method: string;
  path: string;
  description: string;
  request?: string;
  response?: string;
}

interface Section {
  id: string;
  title: string;
  description?: string;
  endpoints: Endpoint[];
}

const API_SECTIONS: Section[] = [
  {
    id: "base-url",
    title: "Base URL",
    description: "All API requests should be made to the following base URLs based on your environment:",
    endpoints: [],
  },
  {
    id: "authentication",
    title: "Authentication",
    description: "FunctionFly supports multiple authentication methods:",
    endpoints: [
      {
        method: "POST",
        path: "/v1/auth/login",
        description: "Login with email and password",
        request: `{
  "email": "user@example.com",
  "password": "password"
}`,
        response: `{
  "token": "<jwt>",
  "expiresIn": 3600
}`,
      },
      {
        method: "POST",
        path: "/v1/auth/api-key",
        description: "Authenticate with API key",
        request: `// Header:
Authorization: Bearer <api_key>`,
      },
      {
        method: "POST",
        path: "/v1/auth/refresh",
        description: "Refresh access token",
        request: `{
  "refresh_token": "..."
}`,
      },
    ],
  },
  {
    id: "users",
    title: "Users",
    description: "Manage user profiles and settings",
    endpoints: [
      {
        method: "GET",
        path: "/v1/users/me",
        description: "Get current user profile",
      },
      {
        method: "PATCH",
        path: "/v1/users/me",
        description: "Update current user",
        request: `{
  "display_name": "...",
  "bio": "...",
  "avatar_url": "..."
}`,
      },
      {
        method: "GET",
        path: "/v1/users/{username}",
        description: "Get public user profile",
      },
    ],
  },
  {
    id: "functions",
    title: "Functions",
    description: "Create, manage, and deploy serverless functions",
    endpoints: [
      {
        method: "GET",
        path: "/v1/functions",
        description: "List all functions",
      },
      {
        method: "GET",
        path: "/v1/functions/{id}",
        description: "Get function details",
      },
      {
        method: "POST",
        path: "/v1/functions",
        description: "Create a new function",
        request: `{
  "name": "my-function",
  "runtime": "python",
  "code": "..."
}`,
      },
      {
        method: "PATCH",
        path: "/v1/functions/{id}",
        description: "Update function",
      },
      {
        method: "DELETE",
        path: "/v1/functions/{id}",
        description: "Delete function",
      },
      {
        method: "POST",
        path: "/v1/functions/{id}/deploy",
        description: "Deploy function",
      },
    ],
  },
  {
    id: "execution",
    title: "Execution",
    description: "Execute functions and retrieve results",
    endpoints: [
      {
        method: "POST",
        path: "/v1/execute/{functionId}",
        description: "Execute function (public)",
        request: `{
  "data": {}
}`,
      },
      {
        method: "POST",
        path: "/v1/run/{functionId}",
        description: "Execute function (authenticated)",
        request: `// Header:
Authorization: Bearer <token>

{
  "input": {}
}`,
      },
      {
        method: "GET",
        path: "/v1/executions/{id}",
        description: "Get execution result",
      },
    ],
  },
  {
    id: "registry",
    title: "Registry",
    description: "Browse and publish functions to the public registry",
    endpoints: [
      {
        method: "GET",
        path: "/v1/registry/search?q=...",
        description: "Search registry",
      },
      {
        method: "GET",
        path: "/v1/registry/{id}",
        description: "Get function from registry",
      },
      {
        method: "POST",
        path: "/v1/registry/publish",
        description: "Publish function",
        request: `{
  "name": "my-function",
  "description": "...",
  "code": "...",
  "runtime": "python",
  "version": "1.0.0"
}`,
      },
    ],
  },
  {
    id: "api-keys",
    title: "API Keys",
    description: "Manage API keys for programmatic access",
    endpoints: [
      {
        method: "GET",
        path: "/v1/api-keys",
        description: "List API keys",
      },
      {
        method: "POST",
        path: "/v1/api-keys",
        description: "Create API key",
        request: `{
  "name": "Production Key",
  "permissions": ["read", "execute"],
  "environments": ["production"]
}`,
      },
      {
        method: "DELETE",
        path: "/v1/api-keys/{id}",
        description: "Delete API key",
      },
      {
        method: "POST",
        path: "/v1/api-keys/{id}/rotate",
        description: "Rotate API key",
      },
    ],
  },
  {
    id: "monitoring",
    title: "Monitoring",
    description: "Health checks and metrics",
    endpoints: [
      {
        method: "GET",
        path: "/healthz",
        description: "Health check",
      },
      {
        method: "GET",
        path: "/readyz",
        description: "Readiness check",
      },
      {
        method: "GET",
        path: "/metrics",
        description: "Prometheus metrics",
      },
    ],
  },
];

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Ignore copy errors
    }
  };

  return (
    <button
      onClick={handleCopy}
      className="copy-button"
      title="Copy to clipboard"
    >
      {copied ? <Icons.Check /> : <Icons.Copy />}
    </button>
  );
}

function EndpointCard({ endpoint }: { endpoint: Endpoint }) {
  const methodColors: Record<string, string> = {
    GET: "var(--color-success)",
    POST: "var(--color-brand-500)",
    PATCH: "var(--color-warning)",
    DELETE: "var(--color-error)",
    PUT: "var(--color-info)",
  };

  return (
    <div className="endpoint-card">
      <div className="endpoint-header">
        <span
          className="endpoint-method"
          style={{ backgroundColor: methodColors[endpoint.method] || "var(--color-brand-500)" }}
        >
          {endpoint.method}
        </span>
        <code className="endpoint-path">{endpoint.path}</code>
      </div>
      <p className="endpoint-description">{endpoint.description}</p>
      {endpoint.request && (
        <div className="code-block-wrapper">
          <div className="code-block-header">
            <span className="code-block-label">Request</span>
            <CopyButton text={endpoint.request.replace(/\/\/.*/g, "").trim()} />
          </div>
          <pre className="code-block">
            <code>{endpoint.request}</code>
          </pre>
        </div>
      )}
      {endpoint.response && (
        <div className="code-block-wrapper">
          <div className="code-block-header">
            <span className="code-block-label">Response</span>
            <CopyButton text={endpoint.response} />
          </div>
          <pre className="code-block">
            <code>{endpoint.response}</code>
          </pre>
        </div>
      )}
    </div>
  );
}

export default function ApiReference() {
  const [activeSection, setActiveSection] = useState<string>("authentication");

  return (
    <div className="api-reference-container">
      <div className="api-reference-sidebar">
        <div className="sidebar-header">
          <Icons.Terminal />
          <span>API Reference</span>
        </div>
        <nav className="sidebar-nav">
          {API_SECTIONS.map((section) => (
            <button
              key={section.id}
              onClick={() => setActiveSection(section.id)}
              className={`sidebar-link ${activeSection === section.id ? "active" : ""}`}
            >
              {section.title}
            </button>
          ))}
        </nav>
      </div>

      <div className="api-reference-content">
        <div className="api-reference-header">
          <h1>API Reference</h1>
          <p>
            Complete reference for the FunctionFly REST API. Base URLs for each environment:
          </p>
        </div>

        <div className="base-url-table">
          <table>
            <thead>
              <tr>
                <th>Environment</th>
                <th>URL</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>Production</td>
                <td>
                  <code>https://api.functionfly.com</code>
                </td>
              </tr>
              <tr>
                <td>Staging</td>
                <td>
                  <code>https://api.staging.functionfly.com</code>
                </td>
              </tr>
              <tr>
                <td>Development</td>
                <td>
                  <code>http://localhost:8080</code>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div className="api-sections">
          {API_SECTIONS.filter((s) => s.id !== "base-url").map((section) => (
            <section
              key={section.id}
              id={section.id}
              className={`api-section ${activeSection === section.id ? "active" : ""}`}
            >
              <h2 className="section-title">{section.title}</h2>
              {section.description && (
                <p className="section-description">{section.description}</p>
              )}
              <div className="endpoints-list">
                {section.endpoints.map((endpoint, index) => (
                  <EndpointCard key={`${section.id}-${index}`} endpoint={endpoint} />
                ))}
              </div>
            </section>
          ))}
        </div>

        <div className="api-footer">
          <p>
            For detailed SDK documentation, visit the{" "}
            <a
              href="https://github.com/functionfly/functionfly/tree/main/sdk"
              target="_blank"
              rel="noopener noreferrer"
            >
              SDK repositories <Icons.ExternalLink />
            </a>
          </p>
          <p>
            Full API specification available on{" "}
            <a
              href="https://github.com/functionfly/functionfly/blob/main/docs/API.md"
              target="_blank"
              rel="noopener noreferrer"
            >
              GitHub <Icons.ExternalLink />
            </a>
          </p>
        </div>
      </div>
    </div>
  );
}
