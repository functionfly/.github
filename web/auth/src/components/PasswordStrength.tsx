import React, { useState, useEffect, useRef } from "react";

interface PasswordStrengthProps {
  minLength?: number;
}

interface StrengthResult {
  valid: boolean;
  error?: string;
  strength: "weak" | "fair" | "good" | "strong";
  checks: {
    length: boolean;
    upper: boolean;
    lower: boolean;
    number: boolean;
    special: boolean;
  };
}

function getPasswordStrength(password: string, minLength = 8): StrengthResult {
  const checks = {
    length: password.length >= minLength,
    upper: /[A-Z]/.test(password),
    lower: /[a-z]/.test(password),
    number: /[0-9]/.test(password),
    special: /[!@#$%^&*(),.?":{}|<>]/.test(password),
  };
  const score = Object.values(checks).filter(Boolean).length;
  let strength: "weak" | "fair" | "good" | "strong" = "weak";
  if (score >= 5) strength = "strong";
  else if (score >= 4) strength = "good";
  else if (score >= 3) strength = "fair";

  if (!password || password.length < minLength) {
    return { valid: false, error: `Password must be at least ${minLength} characters`, strength, checks };
  }
  return { valid: true, strength, checks };
}

export default function PasswordStrength({ minLength = 8 }: PasswordStrengthProps) {
  const [password, setPassword] = useState("");
  const [strength, setStrength] = useState<StrengthResult | null>(null);
  const intervalRef = useRef<NodeJS.Timeout | null>(null);
  const passwordRef = useRef("");

  useEffect(() => {
    const updatePassword = () => {
      const field = document.getElementById("password-field") as HTMLInputElement;
      if (field && field.value !== passwordRef.current) {
        passwordRef.current = field.value;
        const pwd = field.value;
        setPassword(pwd);
        setStrength(getPasswordStrength(pwd, minLength));
      }
    };

    // Try to read initial value
    updatePassword();

    // Poll for changes since React islands may not share state
    intervalRef.current = setInterval(updatePassword, 100);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [minLength]);

  const widths = { weak: "25%", fair: "50%", good: "75%", strong: "100%" };
  const colors = {
    weak: "var(--ff-error)",
    fair: "var(--ff-warning)",
    good: "var(--ff-flame)",
    strong: "var(--ff-success)",
  };
  const labels = {
    weak: "Weak password",
    fair: "Fair password",
    good: "Good password",
    strong: "Strong password",
  };

  if (!password) {
    return (
      <div className="password-strength hidden">
<div className="strength-bar" style={{ height: "4px", background: "var(--ff-cockpit)", borderRadius: "2px", overflow: "hidden", marginBottom: "0.5rem" }} aria-hidden="true">
          <div className="strength-fill" style={{ width: 0 }} />
        </div>
        <p className="strength-text">Enter a password</p>
        <ul className="password-requirements" aria-label="Password requirements">
          <li>At least 8 characters</li>
          <li>One uppercase letter</li>
          <li>One lowercase letter</li>
          <li>One number</li>
          <li>One special character</li>
        </ul>
      </div>
    );
  }

  return (
    <div className="password-strength" style={{ display: password ? "block" : "none" }}>
      <div className="strength-bar" aria-hidden="true">
        <div
          className="strength-fill"
          style={{
            width: strength?.strength ? widths[strength.strength] : "0%",
            background: strength?.strength ? colors[strength.strength] : "#6E7681",
            transition: "width 0.3s, background-color 0.3s",
            borderRadius: "2px",
          }}
        />
      </div>
      <p
        className="strength-text"
        style={{
          color: strength?.strength ? colors[strength.strength] : "var(--ff-secondary-text)",
          fontSize: "0.75rem",
          fontWeight: 500,
          marginBottom: "0.5rem",
        }}
      >
        {strength?.strength ? labels[strength.strength] : "Enter a password"}
      </p>
      <ul className="password-requirements" aria-label="Password requirements">
        <li className={strength?.checks.length ? "met" : ""}>
          At least 8 characters
        </li>
        <li className={strength?.checks.upper ? "met" : ""}>
          One uppercase letter
        </li>
        <li className={strength?.checks.lower ? "met" : ""}>
          One lowercase letter
        </li>
        <li className={strength?.checks.number ? "met" : ""}>
          One number
        </li>
        <li className={strength?.checks.special ? "met" : ""}>
          One special character
        </li>
      </ul>
    </div>
  );
}
