import { useState, useEffect, useCallback } from 'react';
import { authApi } from '@/api/auth';
import { useDebounce } from '@/hooks/useInfiniteScroll';

export interface UsernameValidationResult {
  isValid: boolean;
  isAvailable?: boolean;
  isLoading: boolean;
  error?: string;
}

export interface UseUsernameValidationOptions {
  debounceMs?: number;
  minLength?: number;
  maxLength?: number;
}

export function useUsernameValidation(
  username: string,
  options: UseUsernameValidationOptions = {}
) {
  const {
    debounceMs = 500,
    minLength = 1,
    maxLength = 50
  } = options;

  const [validation, setValidation] = useState<UsernameValidationResult>({
    isValid: false,
    isLoading: false,
  });

  // Debounce the username input
  const debouncedUsername = useDebounce(username, debounceMs);

  const validateUsername = useCallback(async (value: string) => {
    // Clear previous state
    setValidation(prev => ({ ...prev, isLoading: true, error: undefined }));

    // For required field, empty is invalid
    if (!value) {
      setValidation({
        isValid: false,
        isLoading: false,
        error: 'Username is required',
      });
      return;
    }

    // Check length constraints
    if (value.length < minLength) {
      setValidation({
        isValid: false,
        isLoading: false,
        error: `Username must be at least ${minLength} character${minLength === 1 ? '' : 's'} long`,
      });
      return;
    }

    // Check length constraints
    if (value.length > maxLength) {
      setValidation({
        isValid: false,
        isLoading: false,
        error: `Username must be less than ${maxLength} characters long`,
      });
      return;
    }

    // Check format (letters, numbers, underscores, hyphens only)
    const usernameRegex = /^[a-zA-Z0-9_-]*$/;
    if (!usernameRegex.test(value)) {
      setValidation({
        isValid: false,
        isLoading: false,
        error: 'Username can only contain letters, numbers, underscores, and hyphens',
      });
      return;
    }

    // Check if username contains only valid characters (no spaces, etc.)
    if (value !== value.trim()) {
      setValidation({
        isValid: false,
        isLoading: false,
        error: 'Username cannot contain leading or trailing spaces',
      });
      return;
    }

    // If we pass all basic validation, check availability via API
    try {
      setValidation(prev => ({ ...prev, isLoading: true })); // Show loading while checking API
      const response = await authApi.checkUsernameAvailability(value);

      setValidation({
        isValid: response.available,
        isAvailable: response.available,
        isLoading: false,
        error: response.available ? undefined : 'This username is already taken',
      });
    } catch (error) {
      console.error('Username validation error:', error);
      setValidation({
        isValid: false,
        isLoading: false,
        error: 'Unable to check username availability. Please try again.',
      });
    }
  }, [minLength, maxLength]);

  // Trigger validation when debounced username changes
  useEffect(() => {
    if (debouncedUsername || debouncedUsername === '') {
      validateUsername(debouncedUsername);
    }
  }, [debouncedUsername, validateUsername]);

  return validation;
}
