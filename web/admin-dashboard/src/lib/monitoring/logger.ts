/**
 * Centralized logger for the admin dashboard.
 *
 * Why this exists:
 *  - console.* output in the browser bundle can leak sensitive data
 *    (auth errors, session info, etc.) and pollutes the console for
 *    every user. Most of our `console.error('X failed:', err)` calls
 *    are dev-time signals, not user-facing messages.
 *  - We want a single switch to silence everything in production.
 *  - We want warnings to be sent to a remote sink eventually (Sentry,
 *    Datadog, etc.) without touching every call site.
 *
 * Behavior:
 *  - In production (import.meta.env.PROD === true), all levels are
 *    silenced so the user console stays clean.
 *  - In dev, the logger proxies to console.{log,warn,error} so the
 *    existing DX is preserved.
 *  - error() and warn() are also pushed to a pluggable sink if one is
 *    registered (default: no-op).
 */

export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

export interface LogEntry {
  level: LogLevel;
  message: string;
  context?: unknown;
  timestamp: number;
}

type Sink = (entry: LogEntry) => void;

const sinks: Sink[] = [];

export function registerLogSink(sink: Sink): () => void {
  sinks.push(sink);
  return () => {
    const i = sinks.indexOf(sink);
    if (i >= 0) sinks.splice(i, 1);
  };
}

function isProd(): boolean {
  try {
    return Boolean((import.meta as unknown as { env: { PROD?: boolean } }).env?.PROD);
  } catch {
    return false;
  }
}

function emit(level: LogLevel, message: string, context?: unknown): void {
  const entry: LogEntry = { level, message, context, timestamp: Date.now() };
  if (sinks.length > 0) {
    for (const sink of sinks) {
      try {
        sink(entry);
      } catch {
        // A misbehaving sink must never break the app.
      }
    }
  }
  if (isProd()) return; // Silent in production.

  const fn =
    level === 'error' ? console.error : level === 'warn' ? console.warn : console.log;
  if (context !== undefined) {
    fn(`[admin:${level}] ${message}`, context);
  } else {
    fn(`[admin:${level}] ${message}`);
  }
}

export const logger = {
  debug: (msg: string, ctx?: unknown) => emit('debug', msg, ctx),
  info: (msg: string, ctx?: unknown) => emit('info', msg, ctx),
  warn: (msg: string, ctx?: unknown) => emit('warn', msg, ctx),
  error: (msg: string, ctx?: unknown) => emit('error', msg, ctx),
};
