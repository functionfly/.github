import { z, ZodError, ZodSchema } from 'zod';
import * as schemas from './api-validation';

// Validation result type
export interface ValidationResult<T> {
  success: boolean;
  data?: T;
  error?: string;
  fallbackUsed?: boolean;
}

// Logger interface for validation errors
interface ValidationLogger {
  warn: (message: string, error?: unknown) => void;
  error: (message: string, error?: unknown) => void;
  info: (message: string, data?: unknown) => void;
}

// Default logger (can be overridden)
let logger: ValidationLogger = {
  warn: (message: string, error?: unknown) => {
    console.warn(`[Validation] ${message}`, error);
  },
  error: (message: string, error?: unknown) => {
    console.error(`[Validation] ${message}`, error);
  },
  info: (message: string, data?: unknown) => {
    console.info(`[Validation] ${message}`, data);
  }
};

// Set custom logger
export function setValidationLogger(customLogger: ValidationLogger) {
  logger = customLogger;
}

// Generic safe parse function with fallback
export function safeParse<T>(
  schema: ZodSchema<T>,
  data: unknown,
  fallback?: T,
  context?: string
): ValidationResult<T> {
  try {
    const result = schema.parse(data);
    return { success: true, data: result };
  } catch (error) {
    const errorMessage = error instanceof ZodError
      ? `Validation failed: ${error.errors.map(e => `${e.path.join('.')}: ${e.message}`).join(', ')}`
      : 'Unknown validation error';

    logger.warn(`${context ? `[${context}] ` : ''}${errorMessage}`, error);

    if (fallback !== undefined) {
      logger.warn(`${context ? `[${context}] ` : ''}Using fallback data`);
      return {
        success: false,
        error: errorMessage,
        data: fallback,
        fallbackUsed: true
      };
    }

    return { success: false, error: errorMessage };
  }
}

// Safe parse with default fallback generation
export function safeParseWithDefaults<T>(
  schema: ZodSchema<T>,
  data: unknown,
  defaultGenerator?: () => T,
  context?: string
): T | null {
  const result = safeParse(schema, data, undefined, context);

  if (result.success && result.data) {
    return result.data;
  }

  if (defaultGenerator) {
    logger.warn(`${context ? `[${context}] ` : ''}Using generated default data`);
    return defaultGenerator();
  }

  return null;
}

// Array validation with individual item validation
export function safeParseArray<T>(
  schema: ZodSchema<T>,
  data: unknown,
  context?: string
): T[] {
  if (!Array.isArray(data)) {
    logger.warn(`${context ? `[${context}] ` : ''}Expected array, got ${typeof data}`);
    return [];
  }

  const validItems: T[] = [];
  let invalidCount = 0;

  for (let i = 0; i < data.length; i++) {
    const result = safeParse(schema, data[i], undefined, `${context}[${i}]`);
    if (result.success && result.data !== undefined) {
      validItems.push(result.data);
    } else {
      invalidCount++;
    }
  }

  if (invalidCount > 0) {
    logger.warn(`${context ? `[${context}] ` : ''}Filtered out ${invalidCount} invalid items from array`);
  }

  return validItems;
}

// Specific validation functions for common API responses
export const validateDatabaseHealth = (data: unknown) =>
  safeParse(schemas.databaseHealthSchema, data, undefined, 'DatabaseHealth');

export const validateDatabaseAlert = (data: unknown) =>
  safeParse(schemas.databaseAlertSchema, data, undefined, 'DatabaseAlert');

export const validateDatabaseMetric = (data: unknown) =>
  safeParse(schemas.databaseMetricSchema, data, undefined, 'DatabaseMetric');

export const validateDatabaseAlerts = (data: unknown) =>
  safeParseArray(schemas.databaseAlertSchema, data, 'DatabaseAlerts');

export const validateDatabaseMetrics = (data: unknown) =>
  safeParseArray(schemas.databaseMetricSchema, data, 'DatabaseMetrics');

export const validateUser = (data: unknown) =>
  safeParse(schemas.userSchema, data, undefined, 'User');

export const validateFunctionConfig = (data: unknown) =>
  safeParse(schemas.functionConfigSchema, data, undefined, 'FunctionConfig');

export const validateFunctionConfigs = (data: unknown) =>
  safeParseArray(schemas.functionConfigSchema, data, 'FunctionConfigs');

export const validateFunctionDeployment = (data: unknown) =>
  safeParse(schemas.functionDeploymentSchema, data, undefined, 'FunctionDeployment');

export const validateFunctionDeployments = (data: unknown) =>
  safeParseArray(schemas.functionDeploymentSchema, data, 'FunctionDeployments');

export const validateFunctionLog = (data: unknown) =>
  safeParse(schemas.functionLogSchema, data, undefined, 'FunctionLog');

export const validateFunctionLogs = (data: unknown) =>
  safeParseArray(schemas.functionLogSchema, data, 'FunctionLogs');

export const validateAnalyticsSettings = (data: unknown) =>
  safeParse(schemas.analyticsSettingsSchema, data, undefined, 'AnalyticsSettings');

export const validateTenant = (data: unknown) =>
  safeParse(schemas.tenantSchema, data, undefined, 'Tenant');

export const validateAuditEvent = (data: unknown) =>
  safeParse(schemas.auditEventSchema, data, undefined, 'AuditEvent');

export const validateRealtimeEvent = (data: unknown) =>
  safeParse(schemas.realtimeEventSchema, data, undefined, 'RealtimeEvent');

// Input validation functions
export const validateTimeRange = (timeRange: unknown) =>
  safeParse(schemas.timeRangeSchema, timeRange, '1h' as const, 'TimeRange');

export const validateTableName = (tableName: unknown) =>
  safeParse(schemas.tableNameSchema, tableName, undefined, 'TableName');

// Validation middleware for API responses
export function createValidatedApiResponse<T>(
  schema: ZodSchema<T>,
  fallback?: T
) {
  return (data: unknown, context?: string): T | null => {
    const result = safeParse(schema, data, fallback, context);
    return result.data || null;
  };
}

// Batch validation for multiple items
export function validateBatch<T>(
  schema: ZodSchema<T>,
  items: unknown[],
  context?: string
): { valid: T[], invalid: { index: number, error: string }[] } {
  const valid: T[] = [];
  const invalid: { index: number, error: string }[] = [];

  items.forEach((item, index) => {
    const result = safeParse(schema, item, undefined, `${context}[${index}]`);
    if (result.success && result.data !== undefined) {
      valid.push(result.data);
    } else {
      invalid.push({ index, error: result.error || 'Unknown error' });
    }
  });

  return { valid, invalid };
}

// Validation statistics (for monitoring)
export class ValidationStats {
  private static instance: ValidationStats;
  private stats = {
    totalValidations: 0,
    successfulValidations: 0,
    failedValidations: 0,
    fallbackUsed: 0,
    errorsByType: new Map<string, number>()
  };

  static getInstance(): ValidationStats {
    if (!ValidationStats.instance) {
      ValidationStats.instance = new ValidationStats();
    }
    return ValidationStats.instance;
  }

  recordValidation(success: boolean, fallbackUsed: boolean = false, errorType?: string) {
    this.stats.totalValidations++;
    if (success) {
      this.stats.successfulValidations++;
    } else {
      this.stats.failedValidations++;
      if (errorType) {
        this.stats.errorsByType.set(errorType, (this.stats.errorsByType.get(errorType) || 0) + 1);
      }
    }
    if (fallbackUsed) {
      this.stats.fallbackUsed++;
    }
  }

  getStats() {
    return { ...this.stats, errorsByType: Object.fromEntries(this.stats.errorsByType) };
  }

  reset() {
    this.stats = {
      totalValidations: 0,
      successfulValidations: 0,
      failedValidations: 0,
      fallbackUsed: 0,
      errorsByType: new Map()
    };
  }
}

// Helper to record validation stats automatically
export function withValidationStats<T>(
  schema: ZodSchema<T>,
  data: unknown,
  fallback?: T,
  context?: string
): ValidationResult<T> {
  const result = safeParse(schema, data, fallback, context);
  ValidationStats.getInstance().recordValidation(
    result.success,
    result.fallbackUsed || false,
    result.error ? 'validation_error' : undefined
  );
  return result;
}

// Log validation statistics (useful for monitoring and debugging)
export function logValidationStats() {
  const stats = ValidationStats.getInstance().getStats();
  logger.info('Validation Statistics', stats);
  return stats;
}

// Enhanced validation error reporting
export function reportValidationError(
  error: unknown,
  context?: string,
  additionalData?: Record<string, unknown>
) {
  const errorMessage = error instanceof Error ? error.message : String(error);
  const fullContext = context ? `[${context}]` : '';

  logger.error(`Validation error ${fullContext}: ${errorMessage}`, {
    context,
    error: error instanceof Error ? {
      name: error.name,
      message: error.message,
      stack: error.stack
    } : error,
    additionalData,
    timestamp: new Date().toISOString()
  });

  if (import.meta.env.PROD && typeof window !== 'undefined') {
    const err = error instanceof Error ? error : new Error(errorMessage);
    import('@sentry/react')
      .then((Sentry) => {
        Sentry.captureException(err, {
          tags: { type: 'validation' },
          extra: { context, ...additionalData },
        });
      })
      .catch(() => {});
  }
}

// Create a validated fetch wrapper
export async function validatedFetch<T>(
  schema: ZodSchema<T>,
  url: string,
  options?: RequestInit,
  fallback?: T,
  context?: string
): Promise<ValidationResult<T>> {
  try {
    const response = await fetch(url, options);

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    const data = await response.json();
    return safeParse(schema, data, fallback, context || `fetch ${url}`);
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : 'Unknown fetch error';
    logger.error(`Fetch validation error: ${errorMessage}`, { url, context, error });

    if (fallback !== undefined) {
      return { success: false, data: fallback, fallbackUsed: true, error: errorMessage };
    }

    return { success: false, error: errorMessage };
  }
}