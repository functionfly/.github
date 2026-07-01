import { useState, type InputHTMLAttributes, type ReactNode } from "react";
import { cn } from "../lib/utils";
import { Input } from "./sc";

interface Props extends InputHTMLAttributes<HTMLInputElement> {
  label: ReactNode;
  hint?: string;
  error?: string;
  /** Optional custom ID (defaults to name) */
  id?: string;
  // Note: 'name' is inherited from InputHTMLAttributes
}

export default function InputField(props: Props) {
  const {
    label,
    hint,
    error,
    type,
    id: customId,
    className,
    name,
    ...rest
  } = props;
  const [showPwd, setShowPwd] = useState(false);
  const isPassword = type === "password";

  // Generate IDs for accessibility
  const id = customId || name;
  const errorId = error ? `${id}-error` : undefined;
  const hintId = hint ? `${id}-hint` : undefined;
  const describedBy = [hintId, errorId].filter(Boolean).join(" ") || undefined;

  const autoCompleteValue =
    ((rest as Record<string, unknown>).autoComplete as string) ||
    ((rest as Record<string, unknown>).autocomplete as string);
  const {
    autoComplete: _ac1,
    autocomplete: _ac2,
    type: _type,
    label: _label,
    ...restWithoutAutoComplete
  } = rest as Props & {
    autoComplete?: string;
    autocomplete?: string;
    type?: string;
    label?: string;
  };

  return (
    <div className={cn("ff-field", className)}>
      <Input
        id={id}
        name={name}
        type={isPassword && showPwd ? "text" : type}
        label={typeof label === "string" ? label : undefined}
        error={error}
        autoComplete={autoCompleteValue}
        aria-describedby={describedBy}
        {...restWithoutAutoComplete}
      >
        {isPassword && (
          <button
            type="button"
            className="ff-input__toggle"
            onClick={() => setShowPwd((v) => !v)}
            aria-label={showPwd ? "Hide password" : "Show password"}
            aria-pressed={showPwd}
            style={{
              position: "absolute",
              right: "var(--space-3)",
              top: "50%",
              transform: "translateY(-50%)",
              background: "none",
              border: "none",
              cursor: "pointer",
              color: "var(--text-faint)",
              padding: "var(--space-1)",
            }}
          >
            {showPwd ? (
              <svg
                width="18"
                height="18"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94" />
                <path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19" />
                <line x1="1" y1="1" x2="23" y2="23" />
              </svg>
            ) : (
              <svg
                width="18"
                height="18"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                <circle cx="12" cy="12" r="3" />
              </svg>
            )}
          </button>
        )}
      </Input>
    </div>
  );
}
