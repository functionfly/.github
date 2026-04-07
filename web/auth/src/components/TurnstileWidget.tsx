import React, { useEffect, useRef, useState } from "react";

interface TurnstileWidgetProps {
  onVerify: (token: string) => void;
  onError?: () => void;
  onExpire?: () => void;
  action?: string;
  theme?: "light" | "dark" | "auto";
}

// Cloudflare Turnstile site key from environment or default
const TURNSTILE_SITE_KEY =
  (import.meta.env.PUBLIC_TURNSTILE_SITE_KEY as string | undefined) ||
  "1x00000000000000000000AA"; // Test key - always passes

export const TurnstileWidget: React.FC<TurnstileWidgetProps> = ({
  onVerify,
  onError,
  onExpire,
  action = "auth",
  theme = "dark",
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const widgetIdRef = useRef<string | null>(null);
  const [isLoaded, setIsLoaded] = useState(false);
  const [hasError, setHasError] = useState(false);

  useEffect(() => {
    // Load Turnstile script if not already loaded
    if (!document.getElementById("turnstile-script")) {
      const script = document.createElement("script");
      script.id = "turnstile-script";
      script.src = "https://challenges.cloudflare.com/turnstile/v0/api.js";
      script.async = true;
      script.defer = true;
      script.onload = () => setIsLoaded(true);
      script.onerror = () => {
        setHasError(true);
        onError?.();
      };
      document.body.appendChild(script);
    } else {
      setIsLoaded(true);
    }

    return () => {
      // Cleanup widget on unmount
      if (widgetIdRef.current && window.turnstile) {
        window.turnstile.remove(widgetIdRef.current);
      }
    };
  }, [onError]);

  useEffect(() => {
    if (!isLoaded || !containerRef.current || !window.turnstile) return;

    // Render Turnstile widget
    try {
      widgetIdRef.current = window.turnstile.render(containerRef.current, {
        sitekey: TURNSTILE_SITE_KEY,
        callback: (token: string) => {
          setHasError(false);
          onVerify(token);
        },
        "error-callback": () => {
          setHasError(true);
          onError?.();
        },
        "expired-callback": () => {
          onExpire?.();
        },
        action,
        theme,
        size: "normal",
      });
    } catch (err) {
      setHasError(true);
      onError?.();
    }
  }, [isLoaded, action, theme, onVerify, onError, onExpire]);

  const handleRetry = () => {
    setHasError(false);
    if (widgetIdRef.current && window.turnstile) {
      window.turnstile.reset(widgetIdRef.current);
    }
  };

  return (
    <div className="turnstile-container">
      {!isLoaded && (
        <div className="turnstile-loading">
          <div className="turnstile-spinner" />
          <span>Loading verification...</span>
        </div>
      )}

      {hasError && (
        <div className="turnstile-error">
          <p>Verification failed. Please try again.</p>
          <button
            type="button"
            onClick={handleRetry}
            className="turnstile-retry"
          >
            Retry
          </button>
        </div>
      )}

      <div ref={containerRef} className="turnstile-widget" />

      <style>{`
        .turnstile-container {
          margin: 1rem 0;
          min-height: 65px;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
        }
        .turnstile-loading {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          color: #71717a;
          font-size: 0.875rem;
        }
        .turnstile-spinner {
          width: 16px;
          height: 16px;
          border: 2px solid #27272a;
          border-top-color: #6366f1;
          border-radius: 50%;
          animation: turnstile-spin 0.8s linear infinite;
        }
        @keyframes turnstile-spin {
          to { transform: rotate(360deg); }
        }
        .turnstile-error {
          text-align: center;
          padding: 1rem;
          background: rgba(239, 68, 68, 0.1);
          border: 1px solid rgba(239, 68, 68, 0.3);
          border-radius: 8px;
          margin-bottom: 0.5rem;
        }
        .turnstile-error p {
          color: #fca5a5;
          font-size: 0.875rem;
          margin: 0 0 0.75rem 0;
        }
        .turnstile-retry {
          padding: 0.375rem 1rem;
          background: #27272a;
          border: 1px solid #3f3f46;
          border-radius: 6px;
          color: #e4e4e7;
          font-size: 0.875rem;
          cursor: pointer;
          transition: all 0.15s;
        }
        .turnstile-retry:hover {
          background: #3f3f46;
        }
        .turnstile-widget {
          min-height: 65px;
        }
        .turnstile-widget iframe {
          border-radius: 8px;
        }
      `}</style>
    </div>
  );
};

// Type declaration for Turnstile global
declare global {
  interface Window {
    turnstile?: {
      render: (container: HTMLElement, options: TurnstileOptions) => string;
      reset: (widgetId: string) => void;
      remove: (widgetId: string) => void;
      getResponse: (widgetId: string) => string | undefined;
    };
  }
}

interface TurnstileOptions {
  sitekey: string;
  callback: (token: string) => void;
  "error-callback"?: () => void;
  "expired-callback"?: () => void;
  action?: string;
  theme?: "light" | "dark" | "auto";
  size?: "normal" | "compact" | "invisible";
  tabindex?: number;
}

export default TurnstileWidget;
