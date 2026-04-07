import React, { useCallback, useEffect, useRef, useState } from "react";
import { API_ORIGIN } from "../config";

interface Props {
  initialValue?: string;
  required?: boolean;
}

interface ValidationResult {
  valid: boolean;
  error?: string;
}

export default function InviteCodeField({
  initialValue = "",
  required = true,
}: Props) {
  const [value, setValue] = useState(initialValue);
  const [touched, setTouched] = useState(false);
  const [validation, setValidation] = useState<ValidationResult>({
    valid: false,
  });
  const abortRef = useRef<AbortController | null>(null);

  const validate = useCallback(
    async (code: string) => {
      if (!required && !code.trim()) {
        setValidation({ valid: true });
        return;
      }
      if (!code.trim()) {
        setValidation({ valid: false, error: "invite code is required" });
        return;
      }

      if (abortRef.current) abortRef.current.abort();
      abortRef.current = new AbortController();

      setValidation((prev) => ({ ...prev, error: undefined }));

      try {
        const res = await fetch(
          `${API_ORIGIN}/auth/check-invite-code?code=${encodeURIComponent(code)}`,
          { credentials: "include", signal: abortRef.current.signal },
        );
        const data = await res.json().catch(() => ({}));
        setValidation({ valid: data.valid !== false, error: data.error });
      } catch (err) {
        if (err instanceof Error && err.name === "AbortError") return;
        // Network error — don't block, clear feedback
        setValidation({ valid: true });
      }
    },
    [required],
  );

  // Debounce validation: wait 400ms after user stops typing
  useEffect(() => {
    if (!touched) return;
    const timer = setTimeout(() => validate(value), 400);
    return () => clearTimeout(timer);
  }, [value, touched, validate]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setValue(e.target.value);
  };

  const handleBlur = () => {
    setTouched(true);
    validate(value);
  };

  const showError = touched && validation.error;
  const showSuccess = touched && validation.valid && value.trim();

  return (
    <div className="field">
      <label className="field-label" htmlFor="invite_code">
        Invite code {required && <span style={{ color: "#ef4444" }}>*</span>}
      </label>
      <div className="field-wrap">
        <input
          id="invite_code"
          name="invite_code"
          type="text"
          className={`field-input${showError ? " field-input--error" : showSuccess ? " field-input--success" : ""}`}
          placeholder="Enter your invite code"
          autoComplete="off"
          value={value}
          onChange={handleChange}
          onBlur={handleBlur}
          style={{
            width: "100%",
            maxWidth: "100%",
            padding: "0.625rem 0.875rem",
            background: "#09090b",
            border: "none",
            boxShadow: `0 0 0 1px ${showError ? "#ef4444" : showSuccess ? "#22c55e" : "#27272a"} inset`,
            borderRadius: "8px",
            color: "#fafafa",
            fontSize: "0.9375rem",
            outline: "none",
            transition: "box-shadow 0.15s",
            fontFamily: "inherit",
            boxSizing: "border-box",
          }}
        />
      </div>
      <p
        id="invite-validation-msg"
        className={`invite-validation-msg ${showSuccess ? "valid" : showError ? "invalid" : ""}`}
        style={{
          fontSize: "0.8125rem",
          margin: "0.375rem 0 0",
          minHeight: "1.2rem",
          color: showSuccess
            ? "#22c55e"
            : showError
              ? "#f87171"
              : "transparent",
        }}
      >
        {showSuccess
          ? "✓ Valid invite code"
          : showError
            ? `✗ ${validation.error}`
            : ""}
      </p>

      <style>{`
        .field { display: flex; flex-direction: column; gap: 0.375rem; margin-bottom: 1rem; min-width: 0; }
        .field-label { font-size: 0.875rem; font-weight: 500; color: #e4e4e7; }
        .field-wrap { position: relative; min-width: 0; }
        .field-input { width: 100%; max-width: 100%; box-sizing: border-box; }
        .field-input:focus { box-shadow: 0 0 0 1px #6366f1 inset; }
        .field-input--error { box-shadow: 0 0 0 1px #ef4444 inset !important; }
        .field-input--success { box-shadow: 0 0 0 1px #22c55e inset !important; }
        .field-input::placeholder { color: #52525b; }
      `}</style>
    </div>
  );
}
