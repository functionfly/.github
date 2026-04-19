import React, { useCallback, useEffect, useRef, useState } from "react";
import { API_ORIGIN } from "../config";
import { cn } from "../lib/utils";

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
        setValidation({ valid: true });
      }
    },
    [required],
  );

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
    <div className="ff-field">
      <label className="ff-field__label" htmlFor="invite_code">
        Invite code {required && <span className="text-[var(--ff-error)]">*</span>}
      </label>
      <div className="relative min-w-0">
        <input
          id="invite_code"
          name="invite_code"
          type="text"
          className={cn(
            "ff-input",
            showError && "ff-input--error",
            showSuccess && "ff-input--success"
          )}
          placeholder="Enter your invite code"
          autoComplete="off"
          value={value}
          onChange={handleChange}
          onBlur={handleBlur}
        />
      </div>
      <p
        id="invite-validation-msg"
        className={cn(
          "text-xs mt-1.5 min-h-[1.2rem]",
          showSuccess && "text-[var(--ff-success)]",
          showError && "text-[var(--ff-error)]",
          !showSuccess && !showError && "text-transparent"
        )}
      >
        {showSuccess
          ? "✓ Valid invite code"
          : showError
            ? `✗ ${validation.error}`
            : ""}
      </p>
    </div>
  );
}
