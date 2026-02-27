import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Eye, EyeOff, Shield, Check, X, User, Mail, Key } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { FormError } from "@/components/ui/form-error";
import { LoadingSpinner } from "@/components/ui/loading-spinner";
import { useSignupForm } from "@/hooks/useAuthForms";
import { useAuthStore } from "@/stores/authStore";
import { cn } from "@/lib/utils";

// New authentication libraries
import { useGoogleReCaptcha } from "react-google-recaptcha-v3";
import { useAutoAnimate } from "@formkit/auto-animate/react";

// Enhanced password strength bar
import PasswordStrengthBar from 'react-password-strength-bar';

// Password requirement checklist component
function PasswordRequirements({ password }: { password: string }) {
  const requirements = [
    { key: 'length', label: 'At least 8 characters', test: (p: string) => p.length >= 8 },
    { key: 'uppercase', label: 'One uppercase letter', test: (p: string) => /[A-Z]/.test(p) },
    { key: 'lowercase', label: 'One lowercase letter', test: (p: string) => /[a-z]/.test(p) },
    { key: 'number', label: 'One number', test: (p: string) => /[0-9]/.test(p) },
    { key: 'special', label: 'One special character', test: (p: string) => /[^A-Za-z0-9]/.test(p) },
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
            {passed ? <Check className="w-3 h-3" /> : <X className="w-3 h-3" />}
            <span>{req.label}</span>
          </div>
        );
      })}
    </div>
  );
}

export function SignupForm() {
  const navigate = useNavigate();
  const { signup, isLoading, error, clearError } = useAuthStore();
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  // New authentication states
  const [recaptchaToken, setRecaptchaToken] = useState<string | null>(null);

  // Auto-animate refs
  const [formRef] = useAutoAnimate();

  // reCAPTCHA hook - only enabled in production
  const { executeRecaptcha } = useGoogleReCaptcha();

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isValid, isSubmitting },
    clearErrors,
  } = useSignupForm();

  const password = watch('password');
  const watchedErrors = Object.keys(errors).length > 0;

  // Execute reCAPTCHA when form is valid (only in production)
  useEffect(() => {
    const executeCaptcha = async () => {
      if (executeRecaptcha && isValid && !watchedErrors && import.meta.env.PROD) {
        try {
          const token = await executeRecaptcha('signup');
          setRecaptchaToken(token);
        } catch (error) {
          console.error('reCAPTCHA execution failed:', error);
        }
      }
    };
    executeCaptcha();
  }, [executeRecaptcha, isValid, watchedErrors]);


  const onSubmit = async (data: any) => {
    clearError();
    clearErrors();

    // Verify reCAPTCHA token (only in production)
    if (import.meta.env.PROD && !recaptchaToken) {
      alert('Please complete the security verification');
      return;
    }

    try {
      // Include reCAPTCHA token in signup data (only in production)
      const signupData = {
        ...data,
        ...(import.meta.env.PROD && recaptchaToken ? { recaptchaToken } : {}),
      };

      // Use the proper signup endpoint - now returns SignupResponse
      const response = await signup(signupData);

      // Show success message and navigate to verification page
      if (response.requiresVerification) {
        navigate("/auth/verify-email", {
          state: {
            message: response.message,
            email: data.email,
            emailSent: response.emailSent
          }
        });
      }
    } catch {
      // Error is handled by the store
    }
  };

  return (
    <div className="space-y-6">
      <form ref={formRef} onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        <FormError error={error} />

        {/* reCAPTCHA Badge - only shown in production */}
        {import.meta.env.PROD && (
          <div className="flex items-center justify-center text-xs text-text-muted">
            <Shield className="w-3 h-3 mr-1" />
            Protected by reCAPTCHA
          </div>
        )}

        {/* Name Field */}
        <div className="space-y-2">
          <Label htmlFor="name" className={cn(
            'flex items-center gap-2',
            errors.name && 'text-error',
            !errors.name && watch('name') && 'text-success'
          )}>
            <User className="w-4 h-4" />
            Full Name <span className="text-error">*</span>
          </Label>
          <Input
            id="name"
            type="text"
            placeholder="John Doe"
            className={cn(
              errors.name && 'border-error focus:border-error focus:ring-error',
              !errors.name && watch('name') && 'border-success focus:border-success focus:ring-success'
            )}
            {...register('name')}
          />
          {errors.name && (
            <div className="text-xs text-error">
              {typeof errors.name.message === 'string' ? errors.name.message : 'Invalid name'}
            </div>
          )}
        </div>

      <div className="space-y-2">
        <Label htmlFor="email" className={cn(
          'flex items-center gap-2',
          errors.email && 'text-error',
          !errors.email && watch('email') && 'text-success'
        )}>
          <Mail className="w-4 h-4" />
          Email <span className="text-error">*</span>
        </Label>
        <Input
          id="email"
          type="email"
          placeholder="you@example.com"
          className={cn(
            errors.email && 'border-error focus:border-error focus:ring-error',
            !errors.email && watch('email') && 'border-success focus:border-success focus:ring-success'
          )}
          {...register('email')}
        />
        {errors.email && (
          <div className="text-xs text-error">
            {typeof errors.email.message === 'string' ? errors.email.message : 'Invalid email'}
          </div>
        )}
      </div>

      {/* Invite Code Field */}
      <div className="space-y-2">
        <Label htmlFor="inviteCode" className="flex items-center gap-2 text-text-secondary">
          <Key className="w-4 h-4" />
          Invite Code <span className="text-text-muted text-xs">(optional)</span>
        </Label>
        <Input
          id="inviteCode"
          type="text"
          placeholder="Enter your invite code"
          className={cn(
            errors.inviteCode && 'border-error focus:border-error focus:ring-error'
          )}
          {...register('inviteCode')}
        />
        {errors.inviteCode && (
          <div className="text-xs text-error">
            {typeof errors.inviteCode.message === 'string' ? errors.inviteCode.message : 'Invalid invite code'}
          </div>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="password" className={cn(
          'flex items-center gap-2',
          errors.password && 'text-error',
          !errors.password && password && 'text-success'
        )}>
          Password <span className="text-error">*</span>
        </Label>
        <div className="relative">
          <Input
            id="password"
            type={showPassword ? "text" : "password"}
            placeholder="••••••••"
            className={cn(
              'pr-10',
              errors.password && 'border-error focus:border-error focus:ring-error',
              !errors.password && password && 'border-success focus:border-success focus:ring-success'
            )}
            {...register('password')}
          />
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-secondary transition-colors"
          >
            {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
          </button>
        </div>
        {errors.password && (
          <div className="text-xs text-error">
            {typeof errors.password.message === 'string' ? errors.password.message : 'Invalid password'}
          </div>
        )}
      </div>

      {/* Inline Password Requirements */}
      {password && (
        <div className="mt-2">
          <PasswordRequirements password={password} />
          <div className="mt-3">
            <PasswordStrengthBar
              password={password}
              scoreWords={['Very Weak', 'Weak', 'Fair', 'Good', 'Strong']}
              shortScoreWord="Too short"
              minLength={8}
            />
          </div>
        </div>
      )}

      <div className="space-y-2">
        <Label htmlFor="confirmPassword" className={cn(
          'flex items-center gap-2',
          errors.confirmPassword && 'text-error',
          !errors.confirmPassword && watch('confirmPassword') && 'text-success'
        )}>
          Confirm Password <span className="text-error">*</span>
        </Label>
        <div className="relative">
          <Input
            id="confirmPassword"
            type={showConfirmPassword ? "text" : "password"}
            placeholder="••••••••"
            className={cn(
              'pr-10',
              errors.confirmPassword && 'border-error focus:border-error focus:ring-error',
              !errors.confirmPassword && watch('confirmPassword') && 'border-success focus:border-success focus:ring-success'
            )}
            {...register('confirmPassword')}
          />
          <button
            type="button"
            onClick={() => setShowConfirmPassword(!showConfirmPassword)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-secondary transition-colors"
          >
            {showConfirmPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
          </button>
        </div>
        {errors.confirmPassword && (
          <div className="text-xs text-error">
            {typeof errors.confirmPassword?.message === 'string' ? errors.confirmPassword.message : 'Invalid confirm password'}
          </div>
        )}
      </div>

      <div className="flex items-start gap-3">
        <Checkbox
          id="terms"
          {...register('termsAccepted')}
          className="mt-1"
        />
        <div className="space-y-1">
          <label htmlFor="terms" className="text-sm text-text-secondary leading-tight cursor-pointer">
            I agree to the{" "}
            <a href="/terms" className="text-brand-500 hover:underline">
              Terms of Service
            </a>{" "}
            and{" "}
            <a href="/privacy" className="text-brand-500 hover:underline">
              Privacy Policy
            </a>
          </label>
          {errors.termsAccepted && (
            <div className="text-xs text-error">
              {typeof errors.termsAccepted?.message === 'string' ? errors.termsAccepted.message : 'You must accept the terms'}
            </div>
          )}
        </div>
      </div>

      <Button
        type="submit"
        className="w-full"
        disabled={isLoading || isSubmitting || !isValid || watchedErrors}
      >
        {isLoading || isSubmitting ? (
          <LoadingSpinner text="Creating account..." />
        ) : (
          "Create Account"
        )}
      </Button>
      </form>
    </div>
  );
}
