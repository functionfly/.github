import { useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { Menu, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Logo } from "@/components/common/Logo";
import { UserMenu } from "@/components/layout/UserMenu";
import { SearchButton } from "@/components/layout/SearchButton";
import { ThemeToggle } from "@/components/common/ThemeToggle";
import { ProductsDropdown } from "@/components/common/ProductsDropdown";
import { MarketplaceDropdown } from "@/components/common/MarketplaceDropdown";
import { NotificationBell } from "@/components/notifications";
import { useAuthStore } from "@/stores/authStore";
import { useThemeStore } from "@/stores/themeStore";
import { cn } from "@/lib/utils";
import { DOCS_SITE_URL } from "@/lib/constants";

interface NavbarProps {
  variant?: "landing" | "dashboard";
  className?: string;
  onMenuClick?: () => void;
}

export function Navbar({ variant = "landing", className, onMenuClick }: NavbarProps) {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const location = useLocation();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const user = useAuthStore((state) => state.user);
  const theme = useThemeStore((state) => state.theme);

  const isAdmin = user?.role && ["super_admin", "support", "billing_admin", "developer_admin"].includes(user.role);

  const navigationItems = isAuthenticated ? [
    { path: "/dashboard", label: "Dashboard" },
    { path: "/functions", label: "Functions" },
    { path: "/providers", label: "Providers" },
    { path: "/analytics", label: "Analytics" },
    { path: "/settings", label: "Settings" },
    ...(isAdmin ? [{ path: "/admin", label: "Admin" }] : [])
  ] : [
    { path: "/", label: "Home" },
    { path: "/registry", label: "Functions" },
    { path: "/pricing", label: "Pricing" },
    { path: DOCS_SITE_URL, label: "Docs", external: true }
  ];

  const toggleMobileMenu = () => setIsMobileMenuOpen(!isMobileMenuOpen);

  return (
    <>
      <nav
        className={cn(
          "fixed top-0 left-0 right-0 z-50 glass-navbar",
          className
        )}
      >
        <div className="max-w-7xl mx-auto px-4 lg:px-6 h-16 flex items-center justify-between gap-4">
          {/* Left: menu button + logo with spacing so nav never touches logo on small screens */}
          <div className="flex items-center shrink-0 min-w-0">
            {/* Mobile Menu Button (Dashboard only) */}
            {variant === "dashboard" && (
              <Button
                variant="ghost"
                size="icon"
                className="lg:hidden text-text-secondary hover:text-text-primary mr-2"
                onClick={onMenuClick}
                aria-label="Open navigation menu"
                style={theme === 'light' ? {
                  color: '#1a1a2e',
                } : {}}
              >
                <Menu className="w-5 h-5" />
              </Button>
            )}

            {/* Logo - margin-right keeps space from nav/right section on small screens */}
            <Link to={isAuthenticated ? "/dashboard" : "/"} className="shrink-0 mr-4 md:mr-6" aria-label="FunctionFly home">
              <Logo />
            </Link>
          </div>

          {/* Desktop Navigation - gap-4 on parent + logo mr keeps first nav link clear of logo */}
          <div className="hidden md:flex items-center gap-6">
            {isAuthenticated ? (
              <>
                <ProductsDropdown />
                <MarketplaceDropdown />
                <Link
                  to="/settings"
                  className={cn(
                    "text-text-secondary hover:text-text-primary transition-colors font-medium",
                    location.pathname === "/settings" && "text-text-primary"
                  )}
                  style={theme === 'light' ? {
                    color: '#1a1a2e',
                  } : {}}
                >
                  Settings
                </Link>
                {isAdmin && (
                  <Link
                    to="/admin"
                    className={cn(
                      "text-text-secondary hover:text-text-primary transition-colors font-medium",
                      location.pathname === "/admin" && "text-text-primary"
                    )}
                    style={theme === 'light' ? {
                      color: '#1a1a2e',
                    } : {}}
                  >
                    Admin
                  </Link>
                )}
              </>
            ) : (
              <>
                <ProductsDropdown />
                <Link
                  to="/registry"
                  className={cn(
                    "text-text-secondary hover:text-text-primary transition-colors font-medium",
                    location.pathname === "/registry" && "text-text-primary"
                  )}
                  style={theme === 'light' ? {
                    color: '#1a1a2e',
                  } : {}}
                >
                  Functions
                </Link>
                <MarketplaceDropdown />
                <Link
                  to="/pricing"
                  className={cn(
                    "text-text-secondary hover:text-text-primary transition-colors font-medium",
                    location.pathname === "/pricing" && "text-text-primary"
                  )}
                  style={theme === 'light' ? {
                    color: '#1a1a2e',
                  } : {}}
                >
                  Pricing
                </Link>
                <a
                  href={DOCS_SITE_URL}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-text-secondary hover:text-text-primary transition-colors font-medium"
                  style={theme === 'light' ? {
                    color: '#1a1a2e',
                  } : {}}
                >
                  Docs
                </a>
              </>
            )}
          </div>

          {/* Right Section - shrink-0 so it never overlaps logo on small screens */}
          <div className="flex items-center gap-4 shrink-0">
            {isAuthenticated ? (
              <>
                {/* Search (Dashboard only) */}
                {variant === "dashboard" && <SearchButton />}

                {/* Theme Toggle */}
                <ThemeToggle />

                {/* Notifications (Dashboard only) */}
                {variant === "dashboard" && (
                  <NotificationBell
                    variant="ghost"
                    size="md"
                    className={cn(
                      "relative text-text-secondary hover:text-text-primary hover:bg-bg-hover",
                      theme === 'light' && "text-[#1a1a2e]"
                    )}
                  />
                )}

                {/* User Menu */}
                <UserMenu />
              </>
            ) : (
              <>
                {/* Theme Toggle */}
                <ThemeToggle />

                {/* Auth Buttons */}
                <Link to="/login">
                  <Button
                    variant="ghost"
                    className="text-text-secondary hover:text-text-primary"
                    style={theme === 'light' ? {
                      color: '#1a1a2e',
                    } : {}}
                  >
                    Sign In
                  </Button>
                </Link>
                <Link to="/signup">
                  <Button className={variant === "landing" ? "hero-primary-button text-white" : ""}>Get Started</Button>
                </Link>
              </>
            )}

            {/* Mobile Menu Toggle */}
            <Button
              variant="ghost"
              size="icon"
              className="md:hidden text-text-secondary hover:text-text-primary"
              onClick={toggleMobileMenu}
              aria-label={isMobileMenuOpen ? 'Close navigation menu' : 'Open navigation menu'}
              aria-expanded={isMobileMenuOpen}
              style={theme === 'light' ? {
                color: '#1a1a2e',
              } : {}}
            >
              {isMobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </Button>
          </div>
        </div>
      </nav>

      {/* Mobile Menu */}
      {isMobileMenuOpen && (
        <div className="fixed inset-0 z-40 md:hidden">
          {/* Overlay */}
          <div
            className="fixed inset-0 bg-black/50"
            onClick={() => setIsMobileMenuOpen(false)}
          />

          {/* Menu */}
          <div className="fixed top-16 left-0 right-0 bg-bg-primary border-b border-border-subtle">
            <div className="px-4 py-4 space-y-4">
              {isAuthenticated ? (
                <>
                  {/* Products Section */}
                  <div className="space-y-2">
                    <div className="text-sm font-semibold text-text-primary px-2">
                      Products
                    </div>
                    <Link
                      to="/functions"
                      onClick={() => setIsMobileMenuOpen(false)}
                      className={cn(
                        "block py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium",
                        location.pathname === "/functions" && "text-text-primary"
                      )}
                      style={theme === 'light' ? {
                        color: '#1a1a2e',
                      } : {}}
                    >
                      ⚡ Functions
                    </Link>
                    <Link
                      to="/providers"
                      onClick={() => setIsMobileMenuOpen(false)}
                      className={cn(
                        "block py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium",
                        location.pathname === "/providers" && "text-text-primary"
                      )}
                      style={theme === 'light' ? {
                        color: '#1a1a2e',
                      } : {}}
                    >
                      ☁️ Providers
                    </Link>
                    <Link
                      to="/analytics"
                      onClick={() => setIsMobileMenuOpen(false)}
                      className={cn(
                        "block py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium",
                        location.pathname === "/analytics" && "text-text-primary"
                      )}
                      style={theme === 'light' ? {
                        color: '#1a1a2e',
                      } : {}}
                    >
                      📊 Analytics
                    </Link>
                    <Link
                      to="/api-gateway"
                      onClick={() => setIsMobileMenuOpen(false)}
                      className={cn(
                        "block py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium",
                        location.pathname === "/api-gateway" && "text-text-primary"
                      )}
                      style={theme === 'light' ? {
                        color: '#1a1a2e',
                      } : {}}
                    >
                      🚪 API Gateway
                    </Link>
                    <Link
                      to="/monitoring"
                      onClick={() => setIsMobileMenuOpen(false)}
                      className={cn(
                        "block py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium",
                        location.pathname === "/monitoring" && "text-text-primary"
                      )}
                      style={theme === 'light' ? {
                        color: '#1a1a2e',
                      } : {}}
                    >
                      👁️ Monitoring
                    </Link>
                  </div>

                  {/* Marketplace Section */}
                  <div className="space-y-2">
                    <div className="text-sm font-semibold text-text-primary px-2">
                      Marketplace
                    </div>
                    <Link
                      to="/marketplace/functions"
                      onClick={() => setIsMobileMenuOpen(false)}
                      className={cn(
                        "block py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium",
                        location.pathname === "/marketplace/functions" && "text-text-primary"
                      )}
                      style={theme === 'light' ? {
                        color: '#1a1a2e',
                      } : {}}
                    >
                      ⚡ Function Marketplace
                    </Link>
                    <Link
                      to="/marketplace/agents"
                      onClick={() => setIsMobileMenuOpen(false)}
                      className={cn(
                        "block py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium",
                        location.pathname === "/marketplace/agents" && "text-text-primary"
                      )}
                      style={theme === 'light' ? {
                        color: '#1a1a2e',
                      } : {}}
                    >
                      🤖 Agent Marketplace
                    </Link>
                  </div>

                  <Link
                    to="/settings"
                    onClick={() => setIsMobileMenuOpen(false)}
                    className={cn(
                      "block py-2 text-text-secondary hover:text-text-primary transition-colors font-medium",
                      location.pathname === "/settings" && "text-text-primary"
                    )}
                    style={theme === 'light' ? {
                      color: '#1a1a2e',
                    } : {}}
                  >
                    Settings
                  </Link>

                  {isAdmin && (
                    <Link
                      to="/admin"
                      onClick={() => setIsMobileMenuOpen(false)}
                      className={cn(
                        "block py-2 text-text-secondary hover:text-text-primary transition-colors font-medium",
                        location.pathname === "/admin" && "text-text-primary"
                      )}
                      style={theme === 'light' ? {
                        color: '#1a1a2e',
                      } : {}}
                    >
                      Admin
                    </Link>
                  )}
                </>
              ) : (
                <>
                  <Link
                    to="/registry"
                    onClick={() => setIsMobileMenuOpen(false)}
                    className={cn(
                      "block py-2 text-text-secondary hover:text-text-primary transition-colors font-medium",
                      location.pathname === "/registry" && "text-text-primary"
                    )}
                    style={theme === 'light' ? {
                      color: '#1a1a2e',
                    } : {}}
                  >
                    ⚡ Browse Functions
                  </Link>
                  {/* Marketplace Section */}
                  <div className="space-y-2">
                    <div className="text-sm font-semibold text-text-primary px-2">
                      Marketplace
                    </div>
                    <Link
                      to="/marketplace/functions"
                      onClick={() => setIsMobileMenuOpen(false)}
                      className={cn(
                        "block py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium",
                        location.pathname === "/marketplace/functions" && "text-text-primary"
                      )}
                      style={theme === 'light' ? {
                        color: '#1a1a2e',
                      } : {}}
                    >
                      ⚡ Function Marketplace
                    </Link>
                    <Link
                      to="/marketplace/agents"
                      onClick={() => setIsMobileMenuOpen(false)}
                      className={cn(
                        "block py-2 pl-4 text-text-secondary hover:text-text-primary transition-colors font-medium",
                        location.pathname === "/marketplace/agents" && "text-text-primary"
                      )}
                      style={theme === 'light' ? {
                        color: '#1a1a2e',
                      } : {}}
                    >
                      🤖 Agent Marketplace
                    </Link>
                  </div>

                  <Link
                    to="/pricing"
                    onClick={() => setIsMobileMenuOpen(false)}
                    className={cn(
                      "block py-2 text-text-secondary hover:text-text-primary transition-colors font-medium",
                      location.pathname === "/pricing" && "text-text-primary"
                    )}
                    style={theme === 'light' ? {
                      color: '#1a1a2e',
                    } : {}}
                  >
                    Pricing
                  </Link>
                  <a
                    href={DOCS_SITE_URL}
                    target="_blank"
                    rel="noopener noreferrer"
                    onClick={() => setIsMobileMenuOpen(false)}
                    className="block py-2 text-text-secondary hover:text-text-primary transition-colors font-medium"
                    style={theme === 'light' ? {
                      color: '#1a1a2e',
                    } : {}}
                  >
                    Docs
                  </a>
                </>
              )}

              {/* Theme Toggle */}
              <div className="py-2">
                <ThemeToggle />
              </div>

              {!isAuthenticated && (
                <div className="pt-4 border-t border-border-subtle space-y-2">
                  <Link to="/login" onClick={() => setIsMobileMenuOpen(false)}>
                    <Button variant="ghost" className="w-full justify-start">
                      Sign In
                    </Button>
                  </Link>
                  <Link to="/signup" onClick={() => setIsMobileMenuOpen(false)}>
                    <Button className="w-full">Get Started</Button>
                  </Link>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
