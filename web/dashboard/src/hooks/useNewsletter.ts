import { useState, useCallback } from 'react';
import { API_URLS } from '@/lib/api-urls';

interface NewsletterState {
  isLoading: boolean;
  isSuccess: boolean;
  error: string | null;
}

interface SubscribeParams {
  email: string;
  name?: string;
}

/**
 * Hook for newsletter subscription functionality
 * Uses the public newsletter API endpoints (no auth required)
 */
export function useNewsletter() {
  const [state, setState] = useState<NewsletterState>({
    isLoading: false,
    isSuccess: false,
    error: null,
  });

  const reset = useCallback(() => {
    setState({
      isLoading: false,
      isSuccess: false,
      error: null,
    });
  }, []);

  /**
   * Subscribe to the newsletter
   * @param params - Email and optional name
   * @returns Promise<boolean> indicating success/failure
   */
  const subscribe = useCallback(async (params: SubscribeParams): Promise<boolean> => {
    setState({
      isLoading: true,
      isSuccess: false,
      error: null,
    });

    try {
      const response = await fetch(API_URLS.newsletter.subscribe, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: params.email,
          name: params.name || '',
        }),
      });

      const data = await response.json().catch(() => ({}));

      if (!response.ok) {
        const errorMessage = data.error || data.message || 'Failed to subscribe. Please try again.';
        setState({
          isLoading: false,
          isSuccess: false,
          error: errorMessage,
        });
        return false;
      }

      setState({
        isLoading: false,
        isSuccess: true,
        error: null,
      });
      return true;
    } catch (err) {
      setState({
        isLoading: false,
        isSuccess: false,
        error: err instanceof Error ? err.message : 'An unexpected error occurred. Please try again.',
      });
      return false;
    }
  }, []);

  /**
   * Unsubscribe from the newsletter
   * @param email - Email to unsubscribe
   * @returns Promise<boolean> indicating success/failure
   */
  const unsubscribe = useCallback(async (email: string): Promise<boolean> => {
    setState({
      isLoading: true,
      isSuccess: false,
      error: null,
    });

    try {
      const response = await fetch(API_URLS.newsletter.unsubscribe, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email,
        }),
      });

      const data = await response.json().catch(() => ({}));

      if (!response.ok) {
        const errorMessage = data.error || data.message || 'Failed to unsubscribe. Please try again.';
        setState({
          isLoading: false,
          isSuccess: false,
          error: errorMessage,
        });
        return false;
      }

      setState({
        isLoading: false,
        isSuccess: true,
        error: null,
      });
      return true;
    } catch (err) {
      setState({
        isLoading: false,
        isSuccess: false,
        error: err instanceof Error ? err.message : 'An unexpected error occurred. Please try again.',
      });
      return false;
    }
  }, []);

  return {
    ...state,
    subscribe,
    unsubscribe,
    reset,
  };
}
