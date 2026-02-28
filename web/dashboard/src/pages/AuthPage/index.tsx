import { useState, useEffect } from "react";
import { Link, useLocation } from "react-router-dom";
import { motion } from "framer-motion";
import { Zap, Github, Chrome, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Navbar } from "@/components/common/Navbar";
import { Footer } from "@/pages/LandingPage/components/Footer";
import { LoginForm } from "./LoginForm";
import { SignupForm } from "./SignupForm";

// OAuth Provider type
interface OAuthProvider {
  id: string;
  name: string;
  clientId: string;
}

// Fetch available OAuth providers
async function fetchOAuthProviders(): Promise<OAuthProvider[]> {
  try {
    const response = await fetch(`/v1/auth/oauth/providers`);
    if (!response.ok) return [];
    const data = await response.json();
    return data.providers || [];
  } catch {
    return [];
  }
}

async function handleSocialLogin(provider: string) {
  try {
    const response = await fetch(`/v1/auth/oauth/url?provider=${provider}`);
    if (!response.ok) throw new Error(`Failed to get OAuth URL: ${response.statusText}`);
    const data = await response.json();
    if (data.url) {
      window.location.href = data.url;
    } else {
      console.error("No OAuth URL returned:", data);
    }
  } catch (error) {
    console.error("Social login failed:", error);
  }
}

// OAuth Button component
function OAuthButton({ provider }: { provider: OAuthProvider }) {
  const [isLoading, setIsLoading] = useState(false);

  const handleClick = async () => {
    setIsLoading(true);
    await handleSocialLogin(provider.id);
    setIsLoading(false);
  };

  const icon = provider.id === "github" ? (
    <Github className="w-4 h-4" />
  ) : provider.id === "google" ? (
    <Chrome className="w-4 h-4" />
  ) : null;

  return (
    <Button
      type="button"
      variant="outline"
      className="oauth-button w-full gap-2"
      onClick={handleClick}
      disabled={isLoading}
    >
      {isLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : icon}
      {provider.name}
    </Button>
  );
}

export function AuthPage() {
  const location = useLocation();
  const isLogin = location.pathname === "/login";
  const [activeTab, setActiveTab] = useState<"login" | "signup">(isLogin ? "login" : "signup");
  const [oauthProviders, setOauthProviders] = useState<OAuthProvider[]>([]);
  const [isLoadingProviders, setIsLoadingProviders] = useState(true);

  // Fetch OAuth providers on mount
  useEffect(() => {
    const loadProviders = async () => {
      const providers = await fetchOAuthProviders();
      setOauthProviders(providers);
      setIsLoadingProviders(false);
    };
    loadProviders();
  }, []);

  return (
    <div className="auth-page min-h-screen bg-bg-primary flex flex-col">
      {/* Navbar */}
      <Navbar variant="landing" />

      {/* Main Content */}
      <div className="flex-1 flex pt-16">
        {/* Left Side - Form */}
        <div className="flex-1 flex flex-col justify-center px-4 sm:px-6 lg:px-8 xl:px-12 py-12">
        <div className="auth-form-section w-full max-w-md mx-auto">
          {/* Logo */}
          <Link to="/" className="flex items-center gap-2 mb-8">
            <div className="w-8 h-8 rounded-lg bg-linear-to-br from-[#6366f1] to-[#8b5cf6] flex items-center justify-center">
              <Zap className="w-5 h-5 text-white" fill="currentColor" />
            </div>
            <span className="text-xl font-bold gradient-text">FunctionFly</span>
          </Link>

          {/* Header */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5 }}
          >
            <h1 className="text-2xl font-bold text-text-primary mb-2">
              {activeTab === "login" ? "Welcome back" : "Create your account"}
            </h1>
            <p className="text-text-secondary mb-8">
              {activeTab === "login"
                ? "Sign in to access your dashboard"
                : "Start your journey with FunctionFly"}
            </p>
          </motion.div>

          {/* Tabs */}
          <div className="flex gap-2 p-1 bg-bg-secondary rounded-lg mb-6">
            <Link to="/login" className="flex-1" onClick={() => setActiveTab("login")}>
              <Button
                variant={activeTab === "login" ? "secondary" : "ghost"}
                className={`w-full ${
                  activeTab === "login"
                    ? "bg-bg-tertiary text-text-primary"
                    : "text-text-secondary hover:text-text-primary"
                }`}
              >
                Sign In
              </Button>
            </Link>
            <Link to="/signup" className="flex-1" onClick={() => setActiveTab("signup")}>
              <Button
                variant={activeTab === "signup" ? "secondary" : "ghost"}
                className={`w-full ${
                  activeTab === "signup"
                    ? "bg-bg-tertiary text-text-primary"
                    : "text-text-secondary hover:text-text-primary"
                }`}
              >
                Sign Up
              </Button>
            </Link>
          </div>

          {/* Form */}
          <motion.div
            key={activeTab}
            initial={{ opacity: 0, x: activeTab === "login" ? -20 : 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.3 }}
          >
            {activeTab === "login" ? <LoginForm /> : <SignupForm />}
          </motion.div>

          {/* Divider */}
          <div className="relative my-6">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-border-subtle" />
            </div>
            <div className="relative flex justify-center text-xs">
              <span className="px-2 bg-bg-primary text-text-muted">
                Or continue with
              </span>
            </div>
          </div>

          {/* OAuth Buttons - Show only configured providers */}
          {oauthProviders.length > 0 && (
            <div className="grid grid-cols-2 gap-3">
              {oauthProviders.map((provider) => (
                <OAuthButton key={provider.id} provider={provider} />
              ))}
            </div>
          )}
          
          {/* Fallback: Show static buttons if providers fail to load */}
          {oauthProviders.length === 0 && !isLoadingProviders && (
            <div className="grid grid-cols-2 gap-3">
              <Button
                type="button"
                variant="outline"
                className="oauth-button w-full gap-2"
                onClick={() => handleSocialLogin("github")}
              >
                <Github className="w-4 h-4" />
                GitHub
              </Button>
              <Button
                type="button"
                variant="outline"
                className="oauth-button w-full gap-2"
                onClick={() => handleSocialLogin("google")}
              >
                <Chrome className="w-4 h-4" />
                Google
              </Button>
            </div>
          )}

          {/* Loading state */}
          {isLoadingProviders && (
            <div className="flex justify-center py-2">
              <Loader2 className="w-4 h-4 animate-spin text-text-muted" />
            </div>
          )}
        </div>
      </div>

      {/* Right Side - Illustration */}
      <div className="auth-testimonial-panel hidden lg:flex flex-1 relative overflow-hidden">
        {/* Background */}
        <div className="absolute inset-0 bg-linear-to-br from-[#6366f1]/20 via-[#8b5cf6]/10 to-bg-primary" />
        
        {/* Decorative Elements */}
        <div className="absolute top-1/4 left-1/4 w-64 h-64 bg-[#6366f1]/30 rounded-full blur-[100px]" />
        <div className="absolute bottom-1/4 right-1/4 w-64 h-64 bg-[#8b5cf6]/30 rounded-full blur-[100px]" />

        {/* Content */}
        <div className="relative z-10 flex flex-col justify-center px-12">
          <blockquote className="text-2xl font-medium text-text-primary mb-6">
            "FunctionFly has been a game-changer for our startup. We went from
            worrying about downtime to never thinking about it."
          </blockquote>
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-full bg-linear-to-br from-[#6366f1] to-[#8b5cf6]" />
            <div>
              <p className="font-medium text-text-primary">Alex Chen</p>
              <p className="text-sm text-text-secondary">Founder, TechStart</p>
            </div>
          </div>
        </div>
      </div>
      </div>

      {/* Footer */}
      <Footer />
    </div>
  );
}
