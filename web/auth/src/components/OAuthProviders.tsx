import React, { useEffect, useState } from "react";
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

interface SignupConfig {
  invite_required?: boolean;
  turnstile_required?: boolean;
  turnstile_site_key?: string;
}

export default function OAuthProviders({
  inviteCode,
  redirectAfterLogin,
}: Props) {
  const [loadingProvider, setLoadingProvider] = useState<string | null>(null);
  const [providers, setProviders] = useState<string[]>(["github", "google"]);
  const [oauthError, setOauthError] = useState<string | null>(null);
  const [signupConfig, setSignupConfig] = useState<SignupConfig | null>(null);
  const [configLoading, setConfigLoading] = useState(true);
  const [pendingProvider, setPendingProvider] = useState<string | null>(null);
  const [modalInviteCode, setModalInviteCode] = useState("");
  const [modalError, setModalError] = useState<string | null>(null);
  const [isValidatingInvite, setIsValidatingInvite] = useState(false);

  useEffect(() => {
    // Fetch providers client-side since API isn't available at build time
    fetch(`${API_ORIGIN}/auth/oauth/providers`, { credentials: "include" })
      .then((res) => res.json())
      .then((data) => {
        if (
          data.providers &&
          Array.isArray(data.providers) &&
          data.providers.length > 0
        ) {
          setProviders(data.providers);
        }
      })
      .catch(() => {
        // Keep default providers - OAuth providers are optional
      });

    // Fetch signup config to check if invite is required
    fetch(`${API_ORIGIN}/auth/signup-config`, { credentials: "include" })
      .then((res) => res.json())
      .then((data) => {
        setSignupConfig(data);
        setConfigLoading(false);
      })
      .catch(() => {
        // If config fails to load, assume invite is required (safer default)
        setSignupConfig({ invite_required: true });
        setConfigLoading(false);
      });
  }, []);

  const isInviteRequired =
    configLoading || (signupConfig?.invite_required ?? true);
  const hasInviteCode = Boolean(inviteCode?.trim());
  const needsInviteModal = isInviteRequired && !hasInviteCode;

  const buildOAuthUrl = (provider: string, code?: string) => {
    const params = new URLSearchParams({ provider });
    const effectiveInviteCode = code || inviteCode;
    if (effectiveInviteCode) params.set("invite_code", effectiveInviteCode);
    // NOTE: redirect_uri is only for CLI tools (localhost/127.0.0.1).
    // The web flow uses the backend's configured RedirectURL; do NOT send
    // redirect_uri here or the API will reject it with:
    // "redirect_uri must be http://127.0.0.1 or http://localhost"
    return `${API_ORIGIN}/auth/oauth/url?${params.toString()}`;
  };

  const validateInviteCode = async (code: string): Promise<boolean> => {
    try {
      const res = await fetch(
        `${API_ORIGIN}/auth/check-invite-code?code=${encodeURIComponent(code)}`,
        { credentials: "include" },
      );
      const data = await res.json();
      return data.valid === true;
    } catch {
      return false;
    }
  };

  const startOAuth = async (provider: string, code?: string) => {
    setOauthError(null);
    setLoadingProvider(provider);

    // Analytics tracking
    const trackAuth = (event: string, properties: Record<string, any> = {}) => {
      const eventName = `auth_${event}`;
      if (typeof window !== "undefined" && ((window as any).plausible || (window as any).posthog)) {
        if ((window as any).plausible) {
          (window as any).plausible(eventName, { props: properties });
        }
        if ((window as any).posthog) {
          (window as any).posthog.capture(eventName, properties);
        }
      }
      if (typeof process !== "undefined" && process.env?.NODE_ENV === "development") {
        console.log(`[Analytics] ${eventName}`, properties);
      }
    };

    trackAuth("oauth_start", { provider });
    try {
      const res = await fetch(buildOAuthUrl(provider, code), {
        credentials: "include",
      });
      const data = (await res.json().catch(() => ({}))) as {
        url?: string;
        message?: string;
      };
      if (!res.ok) {
        throw new Error(
          data.message || res.statusText || "Could not start sign-in",
        );
      }
      if (!data.url) {
        throw new Error("No OAuth URL returned from the server");
      }
      window.location.href = data.url;
    } catch (err) {
      setLoadingProvider(null);
      setOauthError(
        err instanceof Error
          ? err.message
          : "Sign-in failed. Please try again.",
      );
    }
  };

  const handleProviderClick = (provider: string) => {
    // Reset any previous errors
    setOauthError(null);

    // If we need an invite code, show the modal
    if (needsInviteModal) {
      setPendingProvider(provider);
      setModalInviteCode("");
      setModalError(null);
      return;
    }

    // Otherwise proceed directly
    void startOAuth(provider);
  };

  const handleModalSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!modalInviteCode.trim() || !pendingProvider) return;

    setIsValidatingInvite(true);
    setModalError(null);

    const isValid = await validateInviteCode(modalInviteCode.trim());

    if (!isValid) {
      setIsValidatingInvite(false);
      setModalError("Invalid or expired invite code");
      return;
    }

    // Close modal and start OAuth with the invite code
    const provider = pendingProvider;
    setPendingProvider(null);
    setIsValidatingInvite(false);
    void startOAuth(provider, modalInviteCode.trim());
  };

  const closeModal = () => {
    setPendingProvider(null);
    setModalInviteCode("");
    setModalError(null);
  };

  // Close modal on escape key
  useEffect(() => {
    if (!pendingProvider) return;
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape") closeModal();
    };
    document.addEventListener("keydown", handleEscape);
    return () => document.removeEventListener("keydown", handleEscape);
  }, [pendingProvider]);

  return (
    <>
      <div className="oauth-divider">
        <span>or continue with</span>
      </div>
      {oauthError ? (
        <p className="oauth-error" role="alert">
          {oauthError}
        </p>
      ) : null}
      <div className="oauth-providers">
        {providers.map((p) => (
          <button
            key={p}
            type="button"
            className={`oauth-btn oauth-btn-${p} ${loadingProvider === p ? "oauth-btn-loading" : ""}`}
            aria-label={`Sign in with ${PROVIDER_NAMES[p] || p}`}
            disabled={loadingProvider !== null || configLoading}
            onClick={() => handleProviderClick(p)}
          >
            {loadingProvider === p ||
            (configLoading && loadingProvider === null) ? (
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
          </button>
        ))}
      </div>

      {/* Invite Code Modal */}
      {pendingProvider && (
        <div
          className="invite-modal-overlay"
          onClick={closeModal}
          role="dialog"
          aria-modal="true"
          aria-labelledby="invite-modal-title"
        >
          <div className="invite-modal" onClick={(e) => e.stopPropagation()}>
            <div className="invite-modal-header">
              <h3 id="invite-modal-title">Invite Required</h3>
              <button
                type="button"
                className="invite-modal-close"
                onClick={closeModal}
                aria-label="Close"
              >
                ×
              </button>
            </div>
            <form onSubmit={handleModalSubmit}>
              <p className="invite-modal-description">
                Sign-ups are currently invite-only. Please enter your invite
                code to continue with{" "}
                {PROVIDER_NAMES[pendingProvider] || pendingProvider}.
              </p>
              <div className="invite-modal-input-wrap">
                <input
                  type="text"
                  value={modalInviteCode}
                  onChange={(e) => setModalInviteCode(e.target.value)}
                  placeholder="Enter invite code"
                  className={`invite-modal-input ${modalError ? "invite-modal-input-error" : ""}`}
                  disabled={isValidatingInvite}
                  autoFocus
                />
                {modalError && (
                  <p className="invite-modal-error">{modalError}</p>
                )}
              </div>
              <div className="invite-modal-actions">
                <button
                  type="button"
                  className="invite-modal-btn invite-modal-btn-secondary"
                  onClick={closeModal}
                  disabled={isValidatingInvite}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="invite-modal-btn invite-modal-btn-primary"
                  disabled={!modalInviteCode.trim() || isValidatingInvite}
                >
                  {isValidatingInvite ? (
                    <>
                      <span className="spinner-small" />
                      Checking...
                    </>
                  ) : (
                    "Continue"
                  )}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      <style>{`
        .oauth-divider {
          display: flex;
          align-items: center;
          gap: 1rem;
          margin: 1.25rem 0;
          color: #71717a;
          font-size: 0.8125rem;
        }
        .oauth-divider::before,
        .oauth-divider::after {
          content: "";
          flex: 1;
          height: 1px;
          background: linear-gradient(90deg, transparent, #27272a 50%, transparent);
        }
        .oauth-providers {
          display: flex;
          gap: 0.75rem;
        }
        .oauth-error {
          margin: 0 0 0.75rem;
          padding: 0.5rem 0.75rem;
          border-radius: 8px;
          border: 1px solid #7f1d1d;
          background: rgba(127, 29, 29, 0.2);
          color: #fca5a5;
          font-size: 0.8125rem;
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
          font: inherit;
          font-size: 0.875rem;
          font-weight: 500;
          transition: all 0.15s ease;
          cursor: pointer;
        }
        .oauth-btn:disabled {
          cursor: wait;
          opacity: 0.85;
        }
        .oauth-btn:hover:not(:disabled) {
          background: #18181b;
          border-color: #3f3f46;
          transform: translateY(-1px);
        }
        .oauth-btn:active:not(:disabled) {
          transform: translateY(0);
        }
        .oauth-btn-loading {
          opacity: 0.7;
          cursor: wait;
        }
        .oauth-btn-github:hover:not(:disabled) {
          border-color: #6e7681;
          background: #161b22;
        }
        .oauth-btn-google:hover:not(:disabled) {
          border-color: #4285f4;
          background: rgba(66, 133, 244, 0.1);
        }
        @keyframes spin {
          from {
            transform: rotate(0deg);
          }
          to {
            transform: rotate(360deg);
          }
        }

        /* Invite Modal */
        .invite-modal-overlay {
          position: fixed;
          inset: 0;
          background: rgba(0, 0, 0, 0.7);
          backdrop-filter: blur(4px);
          display: flex;
          align-items: center;
          justify-content: center;
          z-index: 1000;
          animation: fadeIn 0.2s ease;
        }
        @keyframes fadeIn {
          from {
            opacity: 0;
          }
          to {
            opacity: 1;
          }
        }
        .invite-modal {
          background: #18181b;
          border: 1px solid #27272a;
          border-radius: 12px;
          padding: 1.5rem;
          width: 90%;
          max-width: 400px;
          animation: slideUp 0.2s ease;
        }
        @keyframes slideUp {
          from {
            opacity: 0;
            transform: translateY(16px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
        .invite-modal-header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          margin-bottom: 1rem;
        }
        .invite-modal-header h3 {
          margin: 0;
          font-size: 1.125rem;
          font-weight: 600;
          color: #fafafa;
        }
        .invite-modal-close {
          background: transparent;
          border: none;
          color: #71717a;
          font-size: 1.5rem;
          line-height: 1;
          cursor: pointer;
          padding: 0.25rem;
          transition: color 0.15s;
        }
        .invite-modal-close:hover {
          color: #e4e4e7;
        }
        .invite-modal-description {
          margin: 0 0 1rem;
          font-size: 0.875rem;
          color: #a1a1aa;
          line-height: 1.5;
        }
        .invite-modal-input-wrap {
          margin-bottom: 1.25rem;
        }
        .invite-modal-input {
          width: 100%;
          padding: 0.625rem 0.875rem;
          background: #09090b;
          border: 1px solid #27272a;
          border-radius: 8px;
          color: #e4e4e7;
          font-size: 0.875rem;
          transition: border-color 0.15s, box-shadow 0.15s;
          box-sizing: border-box;
        }
        .invite-modal-input:focus {
          outline: none;
          border-color: #6366f1;
          box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.2);
        }
        .invite-modal-input-error {
          border-color: #ef4444;
        }
        .invite-modal-input-error:focus {
          border-color: #ef4444;
          box-shadow: 0 0 0 2px rgba(239, 68, 68, 0.2);
        }
        .invite-modal-error {
          margin: 0.375rem 0 0;
          font-size: 0.75rem;
          color: #fca5a5;
        }
        .invite-modal-actions {
          display: flex;
          gap: 0.75rem;
          justify-content: flex-end;
        }
        .invite-modal-btn {
          padding: 0.5rem 1rem;
          border-radius: 8px;
          font-size: 0.875rem;
          font-weight: 500;
          cursor: pointer;
          transition: all 0.15s;
          display: flex;
          align-items: center;
          gap: 0.5rem;
        }
        .invite-modal-btn:disabled {
          opacity: 0.6;
          cursor: not-allowed;
        }
        .invite-modal-btn-secondary {
          background: transparent;
          border: 1px solid #27272a;
          color: #a1a1aa;
        }
        .invite-modal-btn-secondary:hover:not(:disabled) {
          background: #27272a;
          color: #e4e4e7;
        }
        .invite-modal-btn-primary {
          background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%);
          border: none;
          color: #fff;
        }
        .invite-modal-btn-primary:hover:not(:disabled) {
          opacity: 0.9;
          transform: translateY(-1px);
        }
        .spinner-small {
          width: 14px;
          height: 14px;
          border: 2px solid rgba(255, 255, 255, 0.3);
          border-top-color: #fff;
          border-radius: 50%;
          animation: spin 0.8s linear infinite;
        }
      `}</style>
    </>
  );
}
