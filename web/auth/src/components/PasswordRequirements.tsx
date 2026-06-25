import React, { useState, useEffect, useRef } from "react";

interface PasswordRequirementsProps {
  minLength?: number;
  inputId?: string;
}

interface Requirements {
  length: boolean;
  upper: boolean;
  lower: boolean;
  number: boolean;
  special: boolean;
}

function checkPassword(password: string, minLength = 8): Requirements {
  return {
    length: password.length >= minLength,
    upper: /[A-Z]/.test(password),
    lower: /[a-z]/.test(password),
    number: /[0-9]/.test(password),
    special: /[!@#$%^&*(),.?":{}|<>]/.test(password),
  };
}

const requirementLabels = [
  { key: "length" as const, label: "At least 8 characters" },
  { key: "upper" as const, label: "One uppercase letter" },
  { key: "lower" as const, label: "One lowercase letter" },
  { key: "number" as const, label: "One number" },
  { key: "special" as const, label: "One special character" },
];

export default function PasswordRequirements({ minLength = 8, inputId = "password-field" }: PasswordRequirementsProps) {
  const [password, setPassword] = useState("");
  const [isExpanded, setIsExpanded] = useState(false);
  const intervalRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    const updatePassword = () => {
      const field = document.getElementById(inputId) as HTMLInputElement;
      if (field && field.value !== password) {
        setPassword(field.value);
      }
    };

    updatePassword();
    intervalRef.current = setInterval(updatePassword, 100);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [inputId]);

  const requirements = checkPassword(password, minLength);
  const metCount = Object.values(requirements).filter(Boolean).length;
  const totalCount = Object.values(requirements).length;
  const allMet = metCount === totalCount;

  return (
    <div className="password-requirements">
      {/* Header - clickable to expand/collapse */}
      <button
        type="button"
        className="password-requirements__header"
        onClick={() => setIsExpanded(!isExpanded)}
        aria-expanded={isExpanded}
      >
        <span className="password-requirements__summary">
          {allMet ? (
            <svg
              className="password-requirements__check-all"
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2.5"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <polyline points="20 6 9 17 4 12" />
            </svg>
          ) : (
            <span className="password-requirements__count">
              {metCount}/{totalCount}
            </span>
          )}
          <span className="password-requirements__text">
            {allMet ? "All requirements met" : "Password requirements"}
          </span>
        </span>
        <svg
          className={`password-requirements__chevron ${isExpanded ? "expanded" : ""}`}
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <polyline points="6 9 12 15 18 9" />
        </svg>
      </button>

      {/* Requirements list */}
      <ul
        className={`password-requirements__list ${isExpanded ? "expanded" : ""}`}
        aria-label="Password requirements"
      >
        {requirementLabels.map(({ key, label }) => (
          <li
            key={key}
            className={`password-requirements__item ${requirements[key] ? "met" : ""}`}
          >
            <span className="password-requirements__icon">
              {requirements[key] ? (
                <svg
                  width="12"
                  height="12"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  strokeWidth="3"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                >
                  <polyline points="20 6 9 17 4 12" />
                </svg>
              ) : (
                <span className="password-requirements__dot" />
              )}
            </span>
            <span className="password-requirements__label">{label}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
