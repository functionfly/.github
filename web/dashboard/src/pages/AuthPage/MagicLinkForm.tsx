import { Button } from '@/components/ui/button';
import { FormError } from '@/components/ui/form-error';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { LoadingSpinner } from '@/components/ui/loading-spinner';
import { trackEvent } from '@/lib/analytics';
import { auth } from '@/lib/auth';
import { cn } from '@/lib/utils';
import { zodResolver } from '@hookform/resolvers/zod';
import { AlertCircle, ArrowLeft, CheckCircle2, Mail, Sparkles } from 'lucide-react';
import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';
import { z } from 'zod';

const magicLinkSchema = z.object({
  email: z.string().email('Please enter a valid email address'),
});

type MagicLinkFormData = z.infer<typeof magicLinkSchema>;

interface MagicLinkFormProps {
  onBack?: () => void;
}

export function MagicLinkForm({ onBack }: MagicLinkFormProps): React.JSX.Element {
  const [searchParams] = useSearchParams();
  const redirectTo = searchParams.get('redirect');

  const [isLoading, setIsLoading] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isValid },
    clearErrors,
  } = useForm<MagicLinkFormData>({
    resolver: zodResolver(magicLinkSchema),
    mode: 'onChange',
  });

  const onSubmit = async (data: MagicLinkFormData) => {
    setError(null);
    clearErrors();
    setIsLoading(true);

    try {
      trackEvent('auth_magic_link_requested');
      const result = await auth.requestMagicLink(data.email, redirectTo ?? '/overview');

      if (result.email_sent) {
        setIsSuccess(true);
        trackEvent('auth_magic_link_sent');
        toast.success('Magic link sent! Check your email.');
      } else {
        // Still show success to prevent email enumeration
        setIsSuccess(true);
      }
    } catch (err) {
      trackEvent('auth_magic_link_failed');
      const message = err instanceof Error ? err.message : 'Failed to send magic link';
      setError(message);
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  };

  if (isSuccess) {
    return (
      <div className="space-y-6 text-center">
        <div className="flex justify-center">
          <div className="rounded-full bg-green-100 p-3 dark:bg-green-900/30">
            <CheckCircle2 className="h-8 w-8 text-green-600 dark:text-green-400" />
          </div>
        </div>
        <div className="space-y-2">
          <h3 className="text-lg font-semibold">Check your email</h3>
          <p className="text-sm text-text-muted">
            We've sent a magic link to your email address. Click the link in the email to sign in instantly.
          </p>
        </div>
        <div className="rounded-lg border border-border-default bg-bg-secondary p-4 text-left">
          <div className="flex items-start gap-3">
            <Mail className="mt-0.5 h-4 w-4 text-text-muted" />
            <div className="text-sm text-text-secondary">
              <p>The magic link expires in 15 minutes and can only be used once.</p>
            </div>
          </div>
        </div>
        {onBack && (
          <Button
            type="button"
            variant="outline"
            onClick={onBack}
            className="w-full"
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to sign in
          </Button>
        )}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="text-center space-y-2">
        <div className="flex justify-center">
          <div className="rounded-full bg-brand-100 p-3 dark:bg-brand-900/30">
            <Sparkles className="h-6 w-6 text-brand-600 dark:text-brand-400" />
          </div>
        </div>
        <h3 className="text-lg font-semibold">Sign in with Magic Link</h3>
        <p className="text-sm text-text-muted">
          No password needed. We'll send a secure sign-in link to your email.
        </p>
      </div>

      {error && (
        <FormError
          error={error}
          className="animate-in fade-in slide-in-from-top-2 duration-200"
        />
      )}

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div className="space-y-2">
          <Label
            htmlFor="magic-link-email"
            className={cn(
              'flex items-center gap-2',
              errors.email && 'text-error',
            )}
          >
            Email Address <span className="text-error">*</span>
          </Label>
          <Input
            id="magic-link-email"
            type="email"
            placeholder="you@example.com"
            className={cn(
              'focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2 transition-all duration-200',
              errors.email && 'auth-input-error',
            )}
            {...register('email')}
            aria-invalid={errors.email ? 'true' : 'false'}
            aria-describedby={errors.email ? 'email-error' : undefined}
          />
          {errors.email && (
            <div id="email-error" className="auth-error-text">
              <AlertCircle className="w-3 h-3" />
              {errors.email.message}
            </div>
          )}
        </div>

        <Button
          type="submit"
          className="w-full focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2 transition-all duration-200"
          disabled={isLoading || !isValid}
        >
          {isLoading ? (
            <LoadingSpinner text="Sending magic link..." />
          ) : (
            <>
              <Sparkles className="mr-2 h-4 w-4" />
              Send Magic Link
            </>
          )}
        </Button>
      </form>

      {onBack && (
        <Button
          type="button"
          variant="ghost"
          onClick={onBack}
          className="w-full"
        >
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back to other sign in options
        </Button>
      )}

      <div className="text-center">
        <p className="text-xs text-text-muted">
          New to FunctionFly?{' '}
          <a
            href="/auth/signup"
            className="text-brand-500 hover:text-brand-400 hover:underline"
          >
            Create an account
          </a>
        </p>
      </div>
    </div>
  );
}
