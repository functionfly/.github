import React, { useCallback, useEffect, useRef, useState } from "react";
import { API_ORIGIN } from "../config";

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
        error: undefined, // Don't show "required" error immediately
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
          // Endpoint not available - skip availability check
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
        return; // Request was cancelled
      }
      // Gracefully degrade - don't block user if API fails
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

  const getInputClass = () => {
    const baseClass = "field-input";
    if (!touched && !username) return baseClass;
    if (validation.isLoading) return baseClass;
    if (validation.isValid && validation.isAvailable)
      return `${baseClass} field-input--success`;
    if (validation.error) return `${baseClass} field-input--error`;
    return baseClass;
  };

  const getStatusIcon = () => {
    if (validation.isLoading) {
      return (
        <svg
          className="status-icon status-icon--loading"
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
          ></circle>
          <path
            opacity="0.75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          ></path>
        </svg>
      );
    }
    if (validation.isValid && validation.isAvailable) {
      return (
        <svg
          className="status-icon status-icon--success"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="3"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <polyline points="20 6 9 17 4 12"></polyline>
        </svg>
      );
    }
    if (validation.error && touched) {
      return (
        <svg
          className="status-icon status-icon--error"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      );
    }
    return null;
  };

  const getHintText = () => {
    if (validation.isLoading) return "Checking availability...";
    if (validation.isValid && validation.isAvailable)
      return "✓ Username is available";
    if (validation.error && touched) return validation.error;
    return "This will be your @username on FunctionFly™";
  };

  const getHintClass = () => {
    if (validation.isLoading) return "field-hint field-hint--loading";
    if (validation.isValid && validation.isAvailable)
      return "field-hint field-hint--success";
    if (validation.error && touched) return "field-hint field-hint--error";
    return "field-hint";
  };

  return (
    <div className="field">
      <label className="field-label" htmlFor="username">
        Username
      </label>
      <div className="field-wrap field-wrap--with-icon">
        <input
          id="username"
          name="username"
          type="text"
          className={getInputClass()}
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

      <style>{`
        .field { display: flex; flex-direction: column; gap: 0.375rem; margin-bottom: 1rem; min-width: 0; }
        .field-label { font-size: 0.875rem; font-weight: 500; color: #e4e4e7; }
        .field-wrap { position: relative; min-width: 0; }
        .field-wrap--with-icon { display: flex; align-items: center; min-width: 0; }
        .field-input {
          width: 100%;
          max-width: 100%;
          padding: 0.625rem 0.875rem;
          padding-right: 2.5rem;
          background: #09090b;
          border: none;
          border-radius: 8px;
          box-shadow: 0 0 0 1px #27272a inset;
          color: #fafafa;
          font-size: 0.9375rem;
          outline: none;
          transition: box-shadow 0.15s;
          font-family: inherit;
          box-sizing: border-box;
        }
        .field-input:focus { box-shadow: 0 0 0 1px #6366f1 inset; }
        .field-input--error { box-shadow: 0 0 0 1px #ef4444 inset; }
        .field-input--error:focus { box-shadow: 0 0 0 1px #ef4444 inset; }
        .field-input--success { box-shadow: 0 0 0 1px #22c55e inset; }
        .field-input--success:focus { box-shadow: 0 0 0 1px #22c55e inset; }
        .field-input::placeholder { color: #52525b; }

        .status-icon {
          position: absolute;
          right: 0.75rem;
          top: 50%;
          transform: translateY(-50%);
          pointer-events: none;
        }
        .status-icon--loading { color: #71717a; animation: spin 1s linear infinite; }
        .status-icon--success { color: #22c55e; }
        .status-icon--error { color: #ef4444; }

        .field-hint { font-size: 0.8125rem; color: #71717a; margin: 0; }
        .field-hint--success { color: #22c55e; }
        .field-hint--error { color: #f87171; }
        .field-hint--loading { color: #a1a1aa; }

        @keyframes spin {
          from { transform: translateY(-50%) rotate(0deg); }
          to { transform: translateY(-50%) rotate(360deg); }
        }
      `}</style>
    </div>
  );
}
