import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { FormError } from "@/components/ui/form-error";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import { OTPInput } from './OTPInput';
import { useAutoAnimate } from '@formkit/auto-animate/react';
import { ArrowLeft, Mail, CheckCircle, AlertCircle, Lock, Eye, EyeOff } from 'lucide-react';
import { usersApi } from '@/api/users';
import { cn } from '@/lib/utils';

const emailSchema = z.object({
  email: z.string().email('Please enter a valid email address'),
});

const passwordSchema = z.object({
  password: z.string()
    .min(8, 'Password must be at least 8 characters')
    .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
    .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
    .regex(/[0-9]/, 'Password must contain at least one number'),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

type EmailFormData = z.infer<typeof emailSchema>;
type PasswordFormData = z.infer<typeof passwordSchema>;

// Password requirements component
function PasswordRequirements({ password }: { password: string }) {
  const requirements = [
    { key: 'length', label: 'At least 8 characters', test: (p: string) => p.length >= 8 },
    { key: 'uppercase', label: 'One uppercase letter', test: (p: string) => /[A-Z]/.test(p) },
    { key: 'lowercase', label: 'One lowercase letter', test: (p: string) => /[a-z]/.test(p) },
    { key: 'number', label: 'One number', test: (p: string) => /[0-9]/.test(p) },
  ];

  return (
    <div className="mt-2 space-y-1">
      {requirements.map((req) => {
        const passed = req.test(password);
        return (
          <div key={req.key} className={cn(
            "flex items-center gap-2 text-xs transition-colors",
            passed ? "text-green-500" : "text-text-muted"
          )}>
            {passed ? <CheckCircle className="w-3 h-3" /> : <AlertCircle className="w-3 h-3" />}
            <span>{req.label}</span>
          </div>
        );
      })}
    </div>
  );
}

interface PasswordResetFlowProps {
  onBack: () => void;
  onComplete: () => void;
}

export function PasswordResetFlow({ onBack, onComplete }: PasswordResetFlowProps) {
  const [step, setStep] = useState<'email' | 'otp' | 'password' | 'success'>('email');
  const [email, setEmail] = useState('');
  const [resetToken, setResetToken] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [animateRef] = useAutoAnimate();

  // Email form
  const emailForm = useForm<EmailFormData>({
    resolver: zodResolver(emailSchema),
  });

  // Password form
  const passwordForm = useForm<PasswordFormData>({
    resolver: zodResolver(passwordSchema),
  });

  const passwordValue = passwordForm.watch('password', '');

  const handleEmailSubmit = async (data: EmailFormData) => {
    setIsLoading(true);
    setError(null);

    try {
      await usersApi.requestPasswordReset(data.email);
      setEmail(data.email);
      setStep('otp');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to send reset email. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleOTPComplete = async (otp: string) => {
    setIsLoading(true);
    setError(null);

    try {
      // The OTP IS the reset token returned via email link.
      // Store it so the password step can use it.
      setResetToken(otp);
      setStep('password');
    } catch (err) {
      setError('Invalid verification code. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handlePasswordSubmit = async (data: PasswordFormData) => {
    setIsLoading(true);
    setError(null);

    try {
      await usersApi.confirmPasswordReset(resetToken, data.password);
      setStep('success');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to reset password. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleResendOTP = async () => {
    setIsLoading(true);
    setError(null);

    try {
      await usersApi.requestPasswordReset(email);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to resend code. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div ref={animateRef} className="space-y-6">
      {/* Back Button */}
      <Button
        type="button"
        variant="ghost"
        onClick={onBack}
        className="flex items-center gap-2 text-text-muted hover:text-text-secondary"
      >
        <ArrowLeft className="w-4 h-4" />
        Back to Login
      </Button>

      {/* Email Step */}
      {step === 'email' && (
        <form onSubmit={emailForm.handleSubmit(handleEmailSubmit)} className="space-y-6">
          <div className="text-center">
            <Mail className="w-12 h-12 mx-auto text-[#6366f1] mb-4" />
            <h3 className="text-lg font-semibold">Reset Your Password</h3>
            <p className="text-sm text-text-muted mt-1">
              Enter your email address and we'll send you a verification code
            </p>
          </div>

          <FormError error={error} />

          <div className="space-y-2">
            <Label htmlFor="reset-email">Email</Label>
            <Input
              id="reset-email"
              type="email"
              placeholder="you@example.com"
              {...emailForm.register('email')}
              className={emailForm.formState.errors.email ? 'border-red-500' : ''}
            />
            {emailForm.formState.errors.email && (
              <p className="text-xs text-red-600 dark:text-red-400">
                {emailForm.formState.errors.email.message}
              </p>
            )}
          </div>

          <Button
            type="submit"
            className="w-full"
            disabled={isLoading}
          >
            {isLoading ? (
              <LoadingSpinner text="Sending..." />
            ) : (
              'Send Reset Code'
            )}
          </Button>
        </form>
      )}

      {/* OTP Step */}
      {step === 'otp' && (
        <OTPInput
          onComplete={handleOTPComplete}
          onResend={handleResendOTP}
          error={error || undefined}
          isLoading={isLoading}
          title="Verify Your Email"
          description={`We've sent a 6-digit code to ${email}`}
        />
      )}

      {/* Password Step */}
      {step === 'password' && (
        <form onSubmit={passwordForm.handleSubmit(handlePasswordSubmit)} className="space-y-6">
          <div className="text-center">
            <h3 className="text-lg font-semibold">Set New Password</h3>
            <p className="text-sm text-text-muted mt-1">
              Choose a strong password for your account
            </p>
          </div>

          <FormError error={error} />

          <div className="space-y-2">
            <Label htmlFor="new-password">New Password</Label>
            <Input
              id="new-password"
              type="password"
              placeholder="••••••••"
              {...passwordForm.register('password')}
              className={passwordForm.formState.errors.password ? 'border-red-500' : ''}
            />
            {passwordValue && (
              <PasswordRequirements password={passwordValue} />
            )}
            {passwordForm.formState.errors.password && (
              <p className="text-xs text-red-600 dark:text-red-400">
                {passwordForm.formState.errors.password.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="confirm-new-password">Confirm New Password</Label>
            <Input
              id="confirm-new-password"
              type="password"
              placeholder="••••••••"
              {...passwordForm.register('confirmPassword')}
              className={passwordForm.formState.errors.confirmPassword ? 'border-red-500' : ''}
            />
            {passwordForm.formState.errors.confirmPassword && (
              <p className="text-xs text-red-600 dark:text-red-400">
                {passwordForm.formState.errors.confirmPassword.message}
              </p>
            )}
          </div>

          <Button
            type="submit"
            className="w-full"
            disabled={isLoading}
          >
            {isLoading ? (
              <LoadingSpinner text="Updating..." />
            ) : (
              'Update Password'
            )}
          </Button>
        </form>
      )}

      {/* Success Step */}
      {step === 'success' && (
        <div className="text-center py-6">
          <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-green-100 flex items-center justify-center">
            <CheckCircle className="w-8 h-8 text-green-600" />
          </div>
          <h3 className="text-lg font-semibold">Password Reset Complete!</h3>
          <p className="text-sm text-text-muted mt-2 mb-6">
            Your password has been successfully updated. You can now sign in with your new password.
          </p>
          <Button onClick={onBack} className="w-full">
            Go to Login
          </Button>
        </div>
      )}
    </div>
  );
}