import { FrameButton } from "@/components/containment/FrameButton";
import { RefreshCw, Wifi, WifiOff } from "lucide-react";
import React, { useEffect, useState } from "react";
import { Link, useLocation } from "react-router-dom";

interface NavProps {
  onRefresh?: () => void;
  isRefreshing?: boolean;
  isDedicatedServerConnected?: boolean;
}

export const Nav: React.FC<NavProps> = ({
  onRefresh,
  isRefreshing = false,
  isDedicatedServerConnected,
}) => {
  const location = useLocation();
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 80);
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, []);

  const isActive = (path: string) => location.pathname === path;

  return (
    <header
      className="fixed top-0 left-0 right-0"
      style={{
        zIndex: "var(--z-nav)",
        borderBottom: scrolled
          ? "1px solid var(--panel-edge)"
          : "1px solid transparent",
        background: scrolled ? "rgba(10,13,17,0.85)" : "transparent",
        backdropFilter: scrolled ? "blur(8px)" : "none",
        WebkitBackdropFilter: scrolled ? "blur(8px)" : "none",
        transition:
          "background var(--duration-base), border-color var(--duration-base)",
      }}
    >
      <div
        className="mx-auto flex items-center justify-between"
        style={{
          maxWidth: "1180px",
          padding: "22px var(--space-7)",
        }}
      >
        {/* Logo */}
        <Link
          to="/"
          className="inline-flex items-center gap-2 no-underline"
          style={{ textDecoration: "none" }}
        >
          <div
            style={{
              width: "8px",
              height: "8px",
              clipPath: "polygon(50% 0%, 100% 50%, 50% 100%, 0% 50%)",
              backgroundColor: "var(--accent)",
            }}
          />
          <span
            style={{
              fontFamily: "var(--font-display)",
              fontWeight: 700,
              fontSize: "16px",
              color: "var(--text)",
            }}
          >
            FunctionFly
          </span>
        </Link>

        {/* Nav Links */}
        <nav className="hidden md:flex items-center gap-5">
          <Link
            to="/"
            className={`nav-link ${isActive("/") ? "nav-link-active" : ""}`}
          >
            Status
          </Link>
          <Link
            to="/history"
            className={`nav-link ${isActive("/history") ? "nav-link-active" : ""}`}
          >
            History
          </Link>
          <a
            href="https://functionfly.com"
            className="nav-link"
            target="_blank"
            rel="noopener noreferrer"
          >
            Home
          </a>
        </nav>

        {/* Actions */}
        <div className="flex items-center gap-3">
          {isDedicatedServerConnected !== undefined && (
            <span
              className="inline-flex items-center gap-1.5"
              title={`Dedicated server: ${isDedicatedServerConnected ? "connected" : "disconnected"}`}
              style={{
                fontFamily: "var(--font-mono)",
                fontSize: "10px",
                color: isDedicatedServerConnected
                  ? "var(--status-ok)"
                  : "var(--text-faint)",
                letterSpacing: "0.06em",
                textTransform: "uppercase",
              }}
            >
              {isDedicatedServerConnected ? (
                <Wifi style={{ width: 12, height: 12 }} />
              ) : (
                <WifiOff style={{ width: 12, height: 12 }} />
              )}
              <span className="hidden sm:inline">
                {isDedicatedServerConnected ? "Connected" : "Offline"}
              </span>
            </span>
          )}
          {onRefresh && (
            <FrameButton
              size="sm"
              onClick={onRefresh}
              disabled={isRefreshing}
              iconLeft={
                <RefreshCw
                  style={{
                    width: 14,
                    height: 14,
                    transform: isRefreshing ? "rotate(360deg)" : undefined,
                    transition: "transform 0.5s ease",
                  }}
                />
              }
            >
              Refresh
            </FrameButton>
          )}
        </div>
      </div>
    </header>
  );
};
