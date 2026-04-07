/**
 * Shared validation utilities for auth forms
 */

export interface ValidationResult {
  valid: boolean;
  error?: string;
}

export const validators = {
  /**
   * Validate email format
   */
  email: (value: string): ValidationResult => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!value || value.trim() === "") {
      return { valid: false, error: "Email is required" };
    }
    if (!emailRegex.test(value)) {
      return { valid: false, error: "Please enter a valid email address" };
    }
    return { valid: true };
  },

  /**
   * Validate password strength
   */
  password: (
    value: string,
    options: { minLength?: number; requireSpecial?: boolean } = {}
  ): ValidationResult & { strength: PasswordStrength; checks: PasswordChecks } => {
    const { minLength = 8, requireSpecial = false } = options;

    const checks: PasswordChecks = {
      length: value.length >= minLength,
      upper: /[A-Z]/.test(value),
      lower: /[a-z]/.test(value),
      number: /[0-9]/.test(value),
      special: /[!@#$%^&*(),.?":{}|<>]/.test(value),
    };

    const score = Object.values(checks).filter(Boolean).length;
    let strength: PasswordStrength = "weak";
    if (score >= 5 || (score >= 4 && !requireSpecial)) {
      strength = "strong";
    } else if (score >= 4 || (score >= 3 && !requireSpecial)) {
      strength = "good";
    } else if (score >= 3) {
      strength = "fair";
    }

    if (!value || value.length < minLength) {
      return {
        valid: false,
        error: `Password must be at least ${minLength} characters`,
        strength,
        checks,
      };
    }

    return { valid: true, strength, checks };
  },

  /**
   * Confirm password matches
   */
  passwordMatch: (password: string, confirm: string): ValidationResult => {
    if (confirm && password !== confirm) {
      return { valid: false, error: "Passwords do not match" };
    }
    return { valid: true };
  },

  /**
   * Validate username format
   */
  username: (value: string): ValidationResult => {
    const usernameRegex = /^[a-zA-Z0-9_-]+$/;
    if (!value || value.trim() === "") {
      return { valid: false, error: "Username is required" };
    }
    if (value.length < 3) {
      return { valid: false, error: "Username must be at least 3 characters" };
    }
    if (value.length > 32) {
      return { valid: false, error: "Username must be no more than 32 characters" };
    }
    if (!usernameRegex.test(value)) {
      return {
        valid: false,
        error: "Username can only contain letters, numbers, underscores, and hyphens",
      };
    }
    return { valid: true };
  },

  /**
   * Validate date of birth (must be at least 13 years old)
   */
  dateOfBirth: (value: string): ValidationResult => {
    if (!value) {
      return { valid: false, error: "Date of birth is required" };
    }

    const dob = new Date(value);
    const today = new Date();
    let age = today.getFullYear() - dob.getFullYear();
    const month = today.getMonth() - dob.getMonth();

    if (month < 0 || (month === 0 && today.getDate() < dob.getDate())) {
      age--;
    }

    if (age < 13) {
      return { valid: false, error: "You must be at least 13 years old" };
    }

    if (dob > today) {
      return { valid: false, error: "Date of birth cannot be in the future" };
    }

    return { valid: true };
  },

  /**
   * Validate invite code format
   */
  inviteCode: (value: string): ValidationResult => {
    if (!value || value.trim() === "") {
      return { valid: false, error: "Invite code is required" };
    }
    if (value.length < 6) {
      return { valid: false, error: "Invalid invite code format" };
    }
    return { valid: true };
  },

  /**
   * Validate required field
   */
  required: (value: string, fieldName = "Field"): ValidationResult => {
    if (!value || value.trim() === "") {
      return { valid: false, error: `${fieldName} is required` };
    }
    return { valid: true };
  },
};

export type PasswordStrength = "weak" | "fair" | "good" | "strong";

export interface PasswordChecks {
  length: boolean;
  upper: boolean;
  lower: boolean;
  number: boolean;
  special: boolean;
}

/**
 * Get maximum date for date picker (13 years ago today)
 */
export function getMaxDateOfBirth(): string {
  const maxDate = new Date();
  maxDate.setFullYear(maxDate.getFullYear() - 13);
  return maxDate.toISOString().split("T")[0];
}

/**
 * Format validation errors for display
 */
export function formatErrors(results: ValidationResult[]): string[] {
  return results.filter((r) => !r.valid && r.error).map((r) => r.error!);
}

/**
 * Run all validations and return aggregated result
 */
export function validateAll(
  validations: Array<{ name: string; result: ValidationResult }>
): { valid: boolean; errors: Record<string, string> } {
  const errors: Record<string, string> = {};

  for (const { name, result } of validations) {
    if (!result.valid && result.error) {
      errors[name] = result.error;
    }
  }

  return {
    valid: Object.keys(errors).length === 0,
    errors,
  };
}
