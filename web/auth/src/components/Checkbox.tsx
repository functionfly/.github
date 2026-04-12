import React, { type InputHTMLAttributes, type ReactNode } from "react";

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
        <span className="checkbox-text">
          {labelContent}
          <a href={termsUrl} target="_blank" rel="noopener">Terms of Service</a>
          {" and "}
          <a href={privacyUrl} target="_blank" rel="noopener">Privacy Policy</a>
        </span>
      );
    }
    return <span className="checkbox-text">{labelContent}</span>;
  };

  return (
    <div className={`checkbox-field ${className}`}>
      <label className="checkbox-label" htmlFor={inputId}>
        <input
          id={inputId}
          name={name}
          type="checkbox"
          className={`checkbox-input ${error ? "checkbox-input--error" : ""}`}
          aria-invalid={error ? "true" : undefined}
          aria-describedby={error ? errorId : helperId}
          {...rest}
        />
        {renderLabel(label)}
      </label>

      {helperText && !error && (
        <p className="checkbox-helper" id={helperId}>
          {helperText}
        </p>
      )}

      {error && (
        <p className="checkbox-error" id={errorId} role="alert">
          {error}
        </p>
      )}

      <style>{`
        .checkbox-field {
          margin-bottom: 1rem;
        }
        .checkbox-label {
          display: flex;
          align-items: flex-start;
          gap: 0.625rem;
          font-size: 0.875rem;
          color: #a1a1aa;
          cursor: pointer;
          line-height: 1.5;
        }
        .checkbox-input {
          margin-top: 0.15rem;
          width: 16px;
          height: 16px;
          accent-color: #6366f1;
          cursor: pointer;
          flex-shrink: 0;
        }
        .checkbox-input--error {
          outline: 2px solid #ef4444;
          outline-offset: 1px;
        }
        .checkbox-text {
          flex: 1;
        }
        .checkbox-text a {
          color: #818cf8;
          text-decoration: none;
          font-weight: 500;
        }
        .checkbox-text a:hover {
          text-decoration: underline;
          color: #a5b4fc;
        }
        .checkbox-helper {
          margin: 0.25rem 0 0 1.625rem;
          font-size: 0.75rem;
          color: #71717a;
        }
        .checkbox-error {
          margin: 0.25rem 0 0 1.625rem;
          font-size: 0.75rem;
          color: #f87171;
        }
      `}</style>
    </div>
  );
};

export default Checkbox;
