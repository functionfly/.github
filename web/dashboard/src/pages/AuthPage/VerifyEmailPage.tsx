import { useEffect, useState } from "react";
import { useLocation, useNavigate, Link } from "react-router-dom";
import { motion } from "framer-motion";
import { CheckCircle, Mail, ArrowRight, RefreshCw, ExternalLink } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuthStore } from "@/stores/authStore";
import { authApi } from "@/api/auth";

export function VerifyEmailPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const [isResending, setIsResending] = useState(false);
  const [resendMessage, setResendMessage] = useState<string | null>(null);
  const { clearError } = useAuthStore();

  // Get data from navigation state
  const { message, email, emailSent } = (location.state as any) || {};

  useEffect(() => {
    // If no state data, redirect to signup
    if (!message && !email) {
      navigate("/signup");
    }
  }, [message, email, navigate]);

  const handleResendEmail = async () => {
    if (!email) return;

    setIsResending(true);
    setResendMessage(null);
    clearError();

    try {
      await authApi.resendVerification(email);
      setResendMessage('Verification email sent successfully!');
    } catch (error) {
      setResendMessage(error instanceof Error ? error.message : 'Failed to send verification email. Please try again.');
    } finally {
      setIsResending(false);
    }
  };

  return (
    <main className="min-h-screen bg-bg-primary flex items-center justify-center px-4">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="w-full max-w-md"
      >
        <Card className="border-white/8 bg-bg-secondary">
          <CardHeader className="text-center pb-2">
            <div className="mx-auto w-16 h-16 bg-brand-500/20 rounded-full flex items-center justify-center mb-4">
              <Mail className="w-8 h-8 text-brand-500" />
            </div>
            <CardTitle className="text-xl text-white">
              Check Your Email
            </CardTitle>
            <CardDescription className="text-text-secondary">
              We've sent you a verification link
            </CardDescription>
          </CardHeader>

          <CardContent className="space-y-6">
            <div className="text-center">
              <p className="text-text-secondary mb-4">
                {message || "Please check your email and click the verification link to complete your registration."}
              </p>

              {email && (
                <div className="bg-bg-tertiary rounded-lg p-3 mb-4">
                  <p className="text-sm text-text-muted">
                    Email sent to: <span className="text-white font-medium">{email}</span>
                  </p>
                </div>
              )}

              {emailSent && (
                <div className="flex items-center justify-center gap-2 text-green-400 text-sm">
                  <CheckCircle className="w-4 h-4" />
                  Email sent successfully
                </div>
              )}

              {resendMessage && (
                <div className={`text-center text-sm ${resendMessage.includes('successfully') ? 'text-green-400' : 'text-red-400'}`}>
                  {resendMessage}
                </div>
              )}

              {/* Mailpit link for development */}
              {import.meta.env.DEV && (
                <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-3">
                  <p className="text-sm text-blue-400 mb-2">
                    📧 Development Mode: Check your emails in Mailpit
                  </p>
                  <a
                    href="http://localhost:8025"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1 text-xs text-blue-300 hover:text-blue-200 underline"
                  >
                    Open Mailpit <ExternalLink className="w-3 h-3" />
                  </a>
                </div>
              )}
            </div>

            <div className="space-y-3">
              <Button
                onClick={handleResendEmail}
                disabled={isResending}
                variant="outline"
                className="w-full"
              >
                {isResending ? (
                  <>
                    <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
                    Sending...
                  </>
                ) : (
                  <>
                    <Mail className="w-4 h-4 mr-2" />
                    Resend Verification Email
                  </>
                )}
              </Button>

              <div className="text-center">
                <Link to="/login">
                  <Button variant="ghost" className="text-text-secondary hover:text-white">
                    Already have an account? Sign in
                    <ArrowRight className="w-4 h-4 ml-2" />
                  </Button>
                </Link>
              </div>
            </div>

            <div className="text-xs text-text-muted text-center space-y-2">
              <p>
                Didn't receive the email? Check your spam folder or try resending.
              </p>
              <p>
                The verification link will expire in 24 hours.
              </p>
            </div>
          </CardContent>
        </Card>
      </motion.div>
    </main>
  );
}
