import React, { useState, useEffect } from "react";
import { API_ORIGIN } from "../config";

const PROVIDER_ICONS: Record<string, React.ReactNode> = {
  github: (
    <svg
      width="18"
      height="18"
      viewBox="0 0 24 24"
      fill="currentColor"
      style={{ flexShrink: 0 }}
    >
      <path d="M12 0C5.374 0 0 5.373 0 12c0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0 1 12 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.3 24 12c0-6.627-5.373-12-12-12z" />
    </svg>
  ),
  google: (
    <svg width="18" height="18" viewBox="0 0 24 24" style={{ flexShrink: 0 }}>
      <path
        fill="#4285F4"
        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
      />
      <path
        fill="#34A853"
        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
      />
      <path
        fill="#FBBC05"
        d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
      />
      <path
        fill="#EA4335"
        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
      />
    </svg>
  ),
};

const PROVIDER_NAMES: Record<string, string> = {
  github: "GitHub",
  google: "Google",
};

interface Props {
  inviteCode?: string;
  redirectAfterLogin?: string;
}

export default function OAuthProviders({
  inviteCode,
  redirectAfterLogin,
}: Props) {
  const [loadingProvider, setLoadingProvider] = useState<string | null>(null);
  const [providers, setProviders] = useState<string[]>([]);

  useEffect(() => {
    // Fetch providers client-side since API isn't available at build time
    fetch(`${API_ORIGIN}/auth/oauth/providers`, { credentials: "include" })
      .then((res) => res.json())
      .then((data) => {
        if (data.providers && Array.isArray(data.providers)) {
          setProviders(data.providers);
        }
      })
      .catch(() => {
        // Silently fail - OAuth providers are optional
      });
  }, []);

  if (!providers.length) return null;

  const callbackOrigin =
    typeof window !== "undefined"
      ? window.location.origin
      : "https://auth.functionfly.com";

  const buildOAuthUrl = (provider: string) => {
    const params = new URLSearchParams({ provider });
    if (inviteCode) params.set("invite_code", inviteCode);
    if (redirectAfterLogin) params.set("redirect_uri", redirectAfterLogin);
    else params.set("redirect_uri", `${callbackOrigin}/auth/callback`);
    return `${API_ORIGIN}/auth/oauth/url?${params.toString()}`;
  };

  const handleClick = (provider: string) => (e: React.MouseEvent) => {
    // Don't prevent default - let the link work normally
    // Just set loading state for visual feedback
    setLoadingProvider(provider);
  };

  return (
    <>
      <div className="oauth-divider">
        <span>or continue with</span>
      </div>
      <div className="oauth-providers">
        {providers.map((p) => (
          <a
            key={p}
            href={buildOAuthUrl(p)}
            className={`oauth-btn oauth-btn-${p} ${loadingProvider === p ? "oauth-btn-loading" : ""}`}
            aria-label={`Sign in with ${PROVIDER_NAMES[p] || p}`}
            onClick={handleClick(p)}
          >
            {loadingProvider === p ? (
              <svg
                width="18"
                height="18"
                viewBox="0 0 24 24"
                fill="none"
                style={{ flexShrink: 0, animation: "spin 1s linear infinite" }}
              >
                <circle
                  opacity="0.25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  strokeWidth="4"
                ></circle>
                <path
                  opacity="0.75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
            ) : (
              (PROVIDER_ICONS[p] ?? null)
            )}
            <span>
              {PROVIDER_NAMES[p] || p.charAt(0).toUpperCase() + p.slice(1)}
            </span>
          </a>
        ))}
      </div>
      <style>{`
        .oauth-divider {
          display: flex;
          align-items: center;
          gap: 1rem;
          margin: 1.25rem 0;
          color: #71717a;
          font-size: 0.8125rem;
        }
        .oauth-divider::before, .oauth-divider::after {
          content: "";
          flex: 1;
          height: 1px;
          background: linear-gradient(90deg, transparent, #27272a 50%, transparent);
        }
        .oauth-providers {
          display: flex;
          gap: 0.75rem;
        }
        .oauth-btn {
          flex: 1;
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 0.5rem;
          padding: 0.625rem 0.875rem;
          border-radius: 8px;
          border: 1px solid #27272a;
          background: #09090b;
          color: #e4e4e7;
          text-decoration: none;
          font-size: 0.875rem;
          font-weight: 500;
          transition: all 0.15s ease;
          cursor: pointer;
        }
        .oauth-btn:hover {
          background: #18181b;
          border-color: #3f3f46;
          transform: translateY(-1px);
        }
        .oauth-btn:active {
          transform: translateY(0);
        }
        .oauth-btn-loading {
          opacity: 0.7;
          cursor: wait;
        }
        .oauth-btn-github:hover {
          border-color: #6e7681;
          background: #161b22;
        }
        .oauth-btn-google:hover {
          border-color: #4285F4;
          background: rgba(66, 133, 244, 0.1);
        }
        @keyframes spin {
          from { transform: rotate(0deg); }
          to { transform: rotate(360deg); }
        }
      `}</style>
    </>
  );
}
