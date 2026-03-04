'use client';

import { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Mail, Check, Loader2, ArrowRight, Sparkles } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { GlassmorphismCard, ShinyButton, TextGradient } from '@/components/ui';

interface NewsletterSignupProps {
  variant?: 'inline' | 'modal';
  onSubscribed?: (email: string) => void;
}

const STORAGE_KEY = 'newsletter_subscribed_email';

export function NewsletterSignup({ variant = 'inline', onSubscribed }: NewsletterSignupProps) {
  const [email, setEmail] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isVisible, setIsVisible] = useState(true);

  // Check if user already subscribed
  useEffect(() => {
    const subscribedEmail = localStorage.getItem(STORAGE_KEY);
    if (subscribedEmail) {
      setIsVisible(false);
    }
  }, []);

  const validateEmail = (email: string): boolean => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  };

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!email.trim()) {
      setError('Please enter your email address');
      return;
    }

    if (!validateEmail(email)) {
      setError('Please enter a valid email address');
      return;
    }

    setIsLoading(true);

    try {
      // Simulate API call - replace with actual newsletter API
      await new Promise((resolve) => setTimeout(resolve, 1000));

      // Store in localStorage to avoid showing again
      localStorage.setItem(STORAGE_KEY, email);

      setIsSuccess(true);
      onSubscribed?.(email);
    } catch (err) {
      setError('Something went wrong. Please try again.');
    } finally {
      setIsLoading(false);
    }
  }, [email, onSubscribed]);

  if (!isVisible) {
    return null;
  }

  return (
    <AnimatePresence>
      {isVisible && (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -20 }}
          transition={{ duration: 0.4 }}
          className={variant === 'modal' ? 'max-w-md mx-auto' : ''}
        >
          <GlassmorphismCard
            blurAmount={16}
            className={`blog-newsletter-card overflow-hidden ${variant === 'modal' ? '' : ''}`}
          >
            {isSuccess ? (
              <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                className="text-center py-4"
              >
                <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-green-500/10 text-green-500 mb-4">
                  <Check className="h-8 w-8" />
                </div>
                <h3 className="text-xl font-semibold mb-2">You're subscribed!</h3>
                <p className="text-muted-foreground">
                  Thanks for subscribing. Check your inbox for a confirmation email.
                </p>
              </motion.div>
            ) : (
              <div className="space-y-4">
                <div className="text-center sm:text-left">
                  <div className="flex items-center gap-2 mb-2">
                    <Sparkles className="h-5 w-5 text-brand-500" />
                    <TextGradient
                      size="subheading"
                      animate={true}
                      className="text-xl font-semibold"
                    >
                      Subscribe to our newsletter
                    </TextGradient>
                  </div>
                  <p className="text-muted-foreground text-sm">
                    Get the latest articles and insights delivered directly to your inbox. No spam, ever.
                  </p>
                </div>

                <form onSubmit={handleSubmit} className="space-y-3">
                  <div className="flex flex-col sm:flex-row gap-3">
                    <div className="relative flex-1">
                      <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                      <Input
                        type="email"
                        placeholder="Enter your email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        disabled={isLoading}
                        className="pl-10 bg-background/60 border-border/50 dark:border-white/10"
                        aria-label="Email address"
                      />
                    </div>
                    <ShinyButton
                      type="submit"
                      disabled={isLoading}
                      size="default"
                    >
                      {isLoading ? (
                        <>
                          <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                          Subscribing...
                        </>
                      ) : (
                        <>
                          Subscribe
                          <ArrowRight className="h-4 w-4 ml-2" />
                        </>
                      )}
                    </ShinyButton>
                  </div>

                  {error && (
                    <motion.p
                      initial={{ opacity: 0, y: -10 }}
                      animate={{ opacity: 1, y: 0 }}
                      className="text-sm text-destructive"
                    >
                      {error}
                    </motion.p>
                  )}
                </form>

                <p className="text-xs text-muted-foreground text-center sm:text-left">
                  We respect your privacy. Unsubscribe at any time.
                </p>
              </div>
            )}
          </GlassmorphismCard>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
