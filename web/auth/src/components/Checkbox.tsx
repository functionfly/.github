import React, { type InputHTMLAttributes, type ReactNode } from "react";
import { cn } from "../lib/utils";

interface CheckboxProps extends Omit<InputHTMLAttributes<HTMLInputElement>, "type"> {
  label: ReactNode;
  name: string;
  error?: string;
  helperText?: string;
  termsUrl?: string;
  privacyUrl?: string;
}

export const Checkbox: React.FC<CheckboxProps> = ({
  label,
  name,
  error,
  helperText,
  termsUrl,
  privacyUrl,
  className = "",
  id,
  ...rest
}) => {
  const inputId = id || name;
  const errorId = error ? `${inputId}-error` : undefined;
  const helperId = helperText ? `${inputId}-helper` : undefined;

  const renderLabel = (labelContent: ReactNode) => {
    if (React.isValidElement(labelContent)) {
      return labelContent;
    }
    if (termsUrl && privacyUrl) {
      return (
        <span className="text-sm text-[var(--ff-secondary-text)]">
          {labelContent}
          <a 
            href={termsUrl} 
            target="_blank" 
            rel="noopener"
            className="text-[var(--ff-cyan)] hover:text-[var(--ff-flame)] hover:underline font-medium"
          >
            Terms of Service
          </a>
          {" and "}
          <a 
            href={privacyUrl} 
            target="_blank" 
            rel="noopener"
            className="text-[var(--ff-cyan)] hover:text-[var(--ff-flame)] hover:underline font-medium"
          >
            Privacy Policy
          </a>
        </span>
      );
    }
    return <span className="text-sm text-[var(--ff-secondary-text)]">{labelContent}</span>;
  };

  return (
    <div className={cn("mb-4", className)}>
      <label className="ff-checkbox" htmlFor={inputId}>
        <input
          id={inputId}
          name={name}
          type="checkbox"
          className={cn(
            "ff-checkbox__input",
            error && "border-[var(--ff-error)] shadow-[0_0_0_2px_rgba(255,45,85,0.15)]"
          )}
          aria-invalid={error ? "true" : undefined}
          aria-describedby={error ? errorId : helperId}
          {...rest}
        />
        {renderLabel(label)}
      </label>

      {helperText && !error && (
        <p 
          className="text-xs text-[var(--ff-muted-text)] mt-1 ml-7" 
          id={helperId}
        >
          {helperText}
        </p>
      )}

      {error && (
        <p 
          className="text-xs text-[var(--ff-error)] mt-1 ml-7 flex items-center gap-1" 
          id={errorId} 
          role="alert"
        >
          <svg 
            width="12" 
            height="12" 
            viewBox="0 0 24 24" 
            fill="none" 
            stroke="currentColor" 
            strokeWidth="2"
          >
            <circle cx="12" cy="12" r="10" />
            <line x1="12" y1="8" x2="12" y2="12" />
            <line x1="12" y1="16" x2="12.01" y2="16" />
          </svg>
          {error}
        </p>
      )}
    </div>
  );
};

export default Checkbox;
