import { Navbar } from '@/components/common/Navbar';
import { Button } from '@/components/ui/button';
import { getApiBaseUrl, getMarketingPageUrl } from '@/lib/constants';
import { motion } from 'framer-motion';
import { Github, Loader2 } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { toast } from 'sonner';
import { AuthHeroAnimation } from './AuthHeroAnimation';
import { LoginForm } from './LoginForm';
import { SignupForm } from './SignupForm';

// OAuth Provider type
interface OAuthProvider {
  id: string;
  name: string;
  clientId: string;
}

// Fetch available OAuth providers
async function fetchOAuthProviders(): Promise<OAuthProvider[]> {
  try {
    const base = getApiBaseUrl();
    const response = await fetch(`${base}/v1/auth/oauth/providers`);
    if (!response.ok) return [];
    const data = await response.json();
    return data.providers || [];
  } catch {
    return [];
  }
}

async function handleSocialLogin(provider: string) {
  try {
    const base = getApiBaseUrl();
    const response = await fetch(`${base}/v1/auth/oauth/url?provider=${provider}`);
    if (!response.ok) throw new Error(`Failed to get OAuth URL: ${response.statusText}`);
    const data = await response.json();
    if (data.url) {
      window.location.href = data.url;
    } else {
      console.error('No OAuth URL returned:', data);
    }
  } catch (error) {
    console.error('Social login failed:', error);
    toast.error(
      `Social login with ${provider} is not yet configured. Please check the backend OAuth settings.`
    );
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

  const getIcon = () => {
    if (provider.id === 'github') {
      return <Github className="w-4 h-4" />;
    }
    if (provider.id === 'google') {
      return (
        <svg className="w-4 h-4" viewBox="0 0 24 24">
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
      );
    }
    return null;
  };

  const getButtonClass = () => {
    if (provider.id === 'github') {
      return 'social-btn-github w-full gap-2 focus-visible:ring-2 focus-visible:ring-gray-600 focus-visible:ring-offset-2';
    }
    if (provider.id === 'google') {
      return 'social-btn-google w-full gap-2 focus-visible:ring-2 focus-visible:ring-gray-400 focus-visible:ring-offset-2';
    }
    return 'oauth-button w-full gap-2';
  };

  return (
    <Button type="button" className={getButtonClass()} onClick={handleClick} disabled={isLoading}>
      {isLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : getIcon()}
      {provider.name}
    </Button>
  );
}

export function AuthPage() {
  const location = useLocation();
  const isLogin = location.pathname === '/login' || location.pathname === '/auth/login';
  const [activeTab, setActiveTab] = useState<'login' | 'signup'>(isLogin ? 'login' : 'signup');
  const [authMode, setAuthMode] = useState<'email' | 'social'>('email');
  const [oauthProviders, setOauthProviders] = useState<OAuthProvider[]>([]);
  const [isLoadingProviders, setIsLoadingProviders] = useState(true);

  // Sync tab with route (/login or /auth/login → login tab)
  useEffect(() => {
    setActiveTab(
      location.pathname === '/login' || location.pathname === '/auth/login' ? 'login' : 'signup'
    );
  }, [location.pathname]);

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
      <main className="flex-1 flex pt-16">
        {/* Left Side - Form */}
        <div className="flex-1 flex flex-col justify-center px-4 sm:px-6 lg:px-8 xl:px-12 py-12">
          <div className="auth-form-section w-full max-w-md mx-auto">
            {/* Header */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5 }}
            >
              <h1 className="text-2xl font-bold text-text-primary mb-2">
                {activeTab === 'login' ? 'Welcome back' : 'Create your account'}
              </h1>
              <p className="text-text-secondary mb-6">
                {activeTab === 'login'
                  ? 'Sign in with your email or use GitHub / Google'
                  : 'Start your journey with FunctionFly'}
              </p>
            </motion.div>

            {/* Email / Social login tabs (gradient buttons) - shown on login */}
            {activeTab === 'login' && (
              <div className="flex gap-2 mb-6">
                <Button
                  type="button"
                  size="default"
                  onClick={() => setAuthMode('email')}
                  className={`auth-mode-btn flex-1 transition-all duration-300 ${
                    authMode === 'email'
                      ? 'auth-mode-btn-active'
                      : 'auth-mode-btn-inactive hover:bg-bg-hover'
                  }`}
                >
                  Email Login
                </Button>
                <Button
                  type="button"
                  size="default"
                  onClick={() => setAuthMode('social')}
                  className={`auth-mode-btn flex-1 transition-all duration-300 ${
                    authMode === 'social'
                      ? 'auth-mode-btn-active'
                      : 'auth-mode-btn-inactive hover:bg-bg-hover'
                  }`}
                >
                  Social Login
                </Button>
              </div>
            )}

            {/* Form */}
            <motion.div
              key={activeTab}
              initial={{ opacity: 0, x: activeTab === 'login' ? -20 : 20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.3 }}
              className="mb-4"
            >
              {activeTab === 'login' ? (
                <LoginForm authMode={authMode} setAuthMode={setAuthMode} />
              ) : (
                <SignupForm />
              )}
            </motion.div>

            {/* Switch between login and signup */}
            <p className="text-sm text-text-secondary text-center mt-4">
              {activeTab === 'login' ? (
                <>
                  Don&apos;t have an account?{' '}
                  <Link
                    to="/signup"
                    onClick={() => setActiveTab('signup')}
                    className="text-brand-500 font-medium hover:underline"
                  >
                    Sign up
                  </Link>
                </>
              ) : (
                <>
                  Already have an account?{' '}
                  <Link
                    to="/login"
                    onClick={() => setActiveTab('login')}
                    className="text-brand-500 font-medium hover:underline"
                  >
                    Sign in
                  </Link>
                </>
              )}
            </p>
          </div>
        </div>

        {/* Right Side - Custom animation */}
        <div className="auth-testimonial-panel testimonial-mesh-gradient hidden lg:flex flex-1 relative overflow-hidden">
          {/* Background gradient */}
          <div className="absolute inset-0 bg-linear-to-br from-[#6366f1]/20 via-[#8b5cf6]/10 to-bg-primary" />
          {/* Custom function-flow animation */}
          <AuthHeroAnimation />
        </div>
      </main>

      {/* Minimal Footer for Auth Pages */}
      <footer className="auth-footer">
        <div className="auth-footer-content">
          <div className="flex items-center gap-2 text-sm text-text-secondary">
            <span>© {new Date().getFullYear()} FunctionFly LLC</span>
          </div>
          <div className="flex items-center gap-6 text-sm">
            <a
              href={getMarketingPageUrl('/privacy')}
              className="auth-footer-link focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2 rounded-sm"
              rel="noopener noreferrer"
            >
              Privacy Policy
            </a>
            <a
              href={getMarketingPageUrl('/terms')}
              className="auth-footer-link focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2 rounded-sm"
              rel="noopener noreferrer"
            >
              Terms of Service
            </a>
          </div>
        </div>
      </footer>
    </div>
  );
}
