import React, { useCallback, useEffect, useRef, useState } from "react";
import { API_ORIGIN } from "../config";
import { cn } from "../lib/utils";

interface ValidationState {
  isValid: boolean;
  isAvailable?: boolean;
  isLoading: boolean;
  error?: string;
}

// Debounce hook
function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedValue(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);
  return debouncedValue;
}

export default function UsernameField() {
  const [username, setUsername] = useState("");
  const [touched, setTouched] = useState(false);
  const [validation, setValidation] = useState<ValidationState>({
    isValid: false,
    isLoading: false,
  });
  const abortControllerRef = useRef<AbortController | null>(null);

  const debouncedUsername = useDebounce(username, 300);

  const validateUsername = useCallback(async (value: string) => {
    // Cancel previous request
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
    abortControllerRef.current = new AbortController();

    // Clear previous state
    setValidation((prev) => ({ ...prev, isLoading: true, error: undefined }));

    const minLength = 3;
    const maxLength = 50;

    // Check empty - only show error if user has typed before
    if (!value) {
      setValidation({
        isValid: false,
        isLoading: false,
        error: undefined,
      });
      return;
    }

    // Check length
    if (value.length < minLength) {
      setValidation({
        isValid: false,
        isLoading: false,
        error: `Username must be at least ${minLength} characters`,
      });
      return;
    }

    if (value.length > maxLength) {
      setValidation({
        isValid: false,
        isLoading: false,
        error: `Username must be less than ${maxLength} characters`,
      });
      return;
    }

    // Check format (letters, numbers, underscores, hyphens only)
    const usernameRegex = /^[a-zA-Z0-9_-]+$/;
    if (!usernameRegex.test(value)) {
      setValidation({
        isValid: false,
        isLoading: false,
        error:
          "Username can only contain letters, numbers, underscores, and hyphens",
      });
      return;
    }

    // Check if starts with letter or number
    if (!/^[a-zA-Z0-9]/.test(value)) {
      setValidation({
        isValid: false,
        isLoading: false,
        error: "Username must start with a letter or number",
      });
      return;
    }

    // Check no consecutive special characters
    if (/[_-]{2,}/.test(value)) {
      setValidation({
        isValid: false,
        isLoading: false,
        error: "Username cannot have consecutive underscores or hyphens",
      });
      return;
    }

    // Check doesn't end with special character
    if (/[_-]$/.test(value)) {
      setValidation({
        isValid: false,
        isLoading: false,
        error: "Username cannot end with underscore or hyphen",
      });
      return;
    }

    // Check availability via API
    try {
      const response = await fetch(
        `${API_ORIGIN}/auth/check-username?username=${encodeURIComponent(value)}`,
        { signal: abortControllerRef.current?.signal },
      );

      if (!response.ok) {
        if (response.status === 404) {
          setValidation({
            isValid: true,
            isLoading: false,
            error: undefined,
          });
          return;
        }
        throw new Error("Failed to check availability");
      }

      const data = await response.json();

      setValidation({
        isValid: data.available,
        isAvailable: data.available,
        isLoading: false,
        error: data.available ? undefined : "This username is already taken",
      });
    } catch (error) {
      if (error instanceof Error && error.name === "AbortError") {
        return;
      }
      // Gracefully degrade
      setValidation({
        isValid: true,
        isLoading: false,
        error: undefined,
      });
    }
  }, []);

  useEffect(() => {
    if (touched || debouncedUsername) {
      validateUsername(debouncedUsername);
    }
  }, [debouncedUsername, validateUsername, touched]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value.toLowerCase().trim();
    setUsername(value);
    setTouched(true);
  };

  const getStatusIcon = () => {
    if (validation.isLoading) {
      return (
        <svg
          className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--ff-muted-text)] animate-spin pointer-events-none"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
        >
          <circle
            opacity="0.25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            opacity="0.75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      );
    }
    if (validation.isValid && validation.isAvailable) {
      return (
        <svg
          className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--ff-success)] pointer-events-none"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="3"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <polyline points="20 6 9 17 4 12" />
        </svg>
      );
    }
    if (validation.error && touched) {
      return (
        <svg
          className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--ff-error)] pointer-events-none"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <line x1="18" y1="6" x2="6" y2="18" />
          <line x1="6" y1="6" x2="18" y2="18" />
        </svg>
      );
    }
    return null;
  };

  const getHintClass = () => {
    return cn(
      "ff-field__hint",
      validation.isLoading && "text-[var(--ff-secondary-text)]",
      validation.isValid && validation.isAvailable && "text-[var(--ff-success)]",
      validation.error && touched && "text-[var(--ff-error)]"
    );
  };

  const getHintText = () => {
    if (validation.isLoading) return "Checking availability...";
    if (validation.isValid && validation.isAvailable)
      return "✓ Username is available";
    if (validation.error && touched) return validation.error;
    return "This will be your @username on FunctionFly™";
  };

  return (
    <div className="ff-field">
      <label className="ff-field__label" htmlFor="username">
        Username <span className="ff-required" aria-label="required">*</span>
      </label>
      <div className="relative min-w-0">
        <input
          id="username"
          name="username"
          type="text"
          className={cn(
            "ff-input pr-10",
            !touched && !username && "ff-input",
            validation.isValid && validation.isAvailable && "ff-input--success",
            validation.error && touched && "ff-input--error"
          )}
          placeholder="yourhandle"
          required
          autoComplete="username"
          value={username}
          onChange={handleChange}
          minLength={3}
          maxLength={50}
          pattern="[a-zA-Z0-9_\-]+"
        />
        {getStatusIcon()}
      </div>
      <p className={getHintClass()}>{getHintText()}</p>
    </div>
  );
}
