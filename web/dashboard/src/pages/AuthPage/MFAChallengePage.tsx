import { useState, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { Zap, Shield, ArrowLeft } from "lucide-react";
import { Link } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { OTPInput } from "@/components/auth/OTPInput";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import { useAuthStore } from "@/stores/authStore";

interface MFAChallengePageProps {
  onVerify?: (code: string) => Promise<void>;
}

export function MFAChallengePage({ onVerify }: MFAChallengePageProps) {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { verifyMFA, isLoading, error, clearError, user } = useAuthStore();
  
  const [mfaCode, setMfaCode] = useState("");
  const [attempts, setAttempts] = useState(0);
  const [maxAttempts] = useState(5);
  const [isVerifying, setIsVerifying] = useState(false);
  
  const email = searchParams.get("email") || user?.email || "";
  const isRateLimited = attempts >= maxAttempts;

  const handleVerify = async (code: string) => {
    if (isRateLimited || isVerifying) return;
    
    setIsVerifying(true);
    clearError();
    
    try {
      if (onVerify) {
        await onVerify(code);
      } else {
        await verifyMFA(code);
      }
      // MFA verified successfully - navigate to dashboard
      navigate("/dashboard", { replace: true });
    } catch (err) {
      setAttempts((prev) => prev + 1);
      setMfaCode("");
    } finally {
      setIsVerifying(false);
    }
  };

  const handleResend = async () => {
    // Would call API to resend MFA code
    console.log("Resending MFA code...");
  };

  return (
    <div className="min-h-screen bg-bg-primary flex flex-col">
      {/* Simple Header */}
      <header className="border-b border-border-subtle bg-bg-secondary">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <Link to="/" className="flex items-center gap-2">
              <div className="w-8 h-8 rounded-lg bg-linear-to-br from-[#6366f1] to-[#8b5cf6] flex items-center justify-center">
                <Zap className="w-5 h-5 text-white" fill="currentColor" />
              </div>
              <span className="text-xl font-bold gradient-text">FunctionFly</span>
            </Link>
            <Link to="/login">
              <Button variant="ghost" size="sm">
                <ArrowLeft className="w-4 h-4 mr-2" />
                Back to Login
              </Button>
            </Link>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="flex-1 flex items-center justify-center px-4 py-12">
        <div className="w-full max-w-md">
          {/* Icon */}
          <div className="flex justify-center mb-8">
            <div className="w-16 h-16 rounded-full bg-[#6366f1]/10 flex items-center justify-center">
              <Shield className="w-8 h-8 text-[#6366f1]" />
            </div>
          </div>

          {/* Error Message */}
          {error && (
            <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          {/* Rate Limited Message */}
          {isRateLimited && (
            <div className="mb-6 p-4 bg-orange-50 border border-orange-200 rounded-lg">
              <p className="text-sm text-orange-800">
                Too many failed attempts. Please try again later or contact support.
              </p>
            </div>
          )}

          {/* OTP Input */}
          {!isRateLimited && (
            <OTPInput
              length={6}
              onComplete={handleVerify}
              onResend={handleResend}
              isLoading={isVerifying}
              title="Two-Factor Authentication"
              description={`Enter the 6-digit code from your authenticator app`}
              error={error || undefined}
            />
          )}

          {/* Attempts Counter */}
          {!isRateLimited && attempts > 0 && (
            <p className="text-center text-sm text-text-muted mt-4">
              Attempts remaining: {maxAttempts - attempts}
            </p>
          )}

          {/* Loading State */}
          {isVerifying && (
            <div className="flex justify-center py-4">
              <LoadingSpinner text="Verifying..." />
            </div>
          )}

          {/* Help Text */}
          <div className="mt-8 text-center text-sm text-text-muted">
            <p>
              Having trouble?{" "}
              <a href="mailto:support@functionfly.com" className="text-brand-500 hover:underline">
                Contact Support
              </a>
            </p>
          </div>
        </div>
      </main>
    </div>
  );
}
