import React, { useState, type InputHTMLAttributes, type ReactNode } from "react";

interface Props extends InputHTMLAttributes<HTMLInputElement> {
  label: ReactNode;
  hint?: string;
  error?: string;
  name: string;
}

export default function InputField({ label, hint, error, name, type, ...rest }: Props) {
  const [showPwd, setShowPwd] = useState(false);
  const isPassword = type === "password";

  return (
    <div className="field">
      <label className="field-label" htmlFor={name}>
        {label}
      </label>
      <div className="field-wrap">
        <input
          id={name}
          name={name}
          type={isPassword && showPwd ? "text" : type}
          className={`field-input${error ? " field-input--error" : ""}`}
          {...rest}
        />
        {isPassword && (
          <button
            type="button"
            className="pwd-toggle"
            onClick={() => setShowPwd((v) => !v)}
            aria-label={showPwd ? "Hide password" : "Show password"}
          >
            {showPwd ? (
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94"/><path d="M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19"/><line x1="1" y1="1" x2="23" y2="23"/></svg>
            ) : (
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
            )}
          </button>
        )}
      </div>
      {hint && !error && <p className="field-hint">{hint}</p>}
      {error && <p className="field-error">{error}</p>}
      <style>{`
        .field { display: flex; flex-direction: column; gap: 0.375rem; margin-bottom: 1rem; }
        .field-label { font-size: 0.875rem; font-weight: 500; color: #e4e4e7; }
        .field-wrap { position: relative; }
        .field-input {
          width: 100%;
          padding: 0.625rem 0.875rem;
          background: #09090b;
          border: 1px solid #27272a;
          border-radius: 8px;
          color: #fafafa;
          font-size: 0.9375rem;
          outline: none;
          transition: border-color 0.15s;
          font-family: inherit;
        }
        .field-input:focus { border-color: #6366f1; }
        .field-input--error { border-color: #ef4444; }
        .field-input::placeholder { color: #52525b; }
        .pwd-toggle {
          position: absolute; right: 0.625rem; top: 50%; transform: translateY(-50%);
          background: none; border: none; cursor: pointer; color: #71717a; padding: 0.25rem;
          display: flex; align-items: center;
        }
        .pwd-toggle:hover { color: #a1a1aa; }
        .field-hint { font-size: 0.8125rem; color: #71717a; }
        .field-error { font-size: 0.8125rem; color: #f87171; }
      `}</style>
    </div>
  );
}
