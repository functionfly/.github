import React, { useEffect, useRef, useState } from "react";
import { cn } from "../lib/utils";

interface TurnstileWidgetProps {
  onVerify: (token: string) => void;
  onError?: () => void;
  onExpire?: () => void;
  action?: string;
  theme?: "light" | "dark" | "auto";
}

const TURNSTILE_SITE_KEY =
  (import.meta.env.PUBLIC_TURNSTILE_SITE_KEY as string | undefined) ||
  "1x00000000000000000000AA";

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
      if (widgetIdRef.current && window.turnstile) {
        window.turnstile.remove(widgetIdRef.current);
      }
    };
  }, [onError]);

  useEffect(() => {
    if (!isLoaded || !containerRef.current || !window.turnstile) return;

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
    <div className="my-4 min-h-[65px] flex flex-col items-center justify-center">
      {!isLoaded && (
        <div className="flex items-center gap-2 text-sm text-[var(--ff-muted-text)]">
          <span className="ff-spinner" />
          <span>Loading verification...</span>
        </div>
      )}

      {hasError && (
        <div className="text-center p-4 rounded-lg mb-2 bg-red-500/10 border border-red-500/30">
          <p className="text-sm text-[var(--ff-error)] mb-3">
            Verification failed. Please try again.
          </p>
          <button
            type="button"
            onClick={handleRetry}
            className="ff-btn ff-btn--secondary text-xs px-3 py-1.5"
          >
            Retry
          </button>
        </div>
      )}

      <div ref={containerRef} className="min-h-[65px]" />
    </div>
  );
};

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
