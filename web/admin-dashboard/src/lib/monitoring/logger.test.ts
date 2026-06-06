// @vitest-environment happy-dom

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { logger, registerLogSink, type LogEntry } from './logger';

describe('logger', () => {
  let errorSpy: ReturnType<typeof vi.spyOn>;
  let warnSpy: ReturnType<typeof vi.spyOn>;
  let logSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
    logSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
  });

  afterEach(() => {
    errorSpy.mockRestore();
    warnSpy.mockRestore();
    logSpy.mockRestore();
  });

  it('routes error() to console.error in dev', () => {
    // import.meta.env.PROD is false in vitest by default
    logger.error('boom', { code: 500 });
    expect(errorSpy).toHaveBeenCalledWith('[admin:error] boom', { code: 500 });
  });

  it('routes warn() to console.warn in dev', () => {
    logger.warn('careful');
    expect(warnSpy).toHaveBeenCalledWith('[admin:warn] careful');
  });

  it('routes debug() to console.log in dev', () => {
    logger.debug('hello');
    expect(logSpy).toHaveBeenCalledWith('[admin:debug] hello');
  });

  it('does not log when context is undefined and no context arg is provided', () => {
    logger.info('hi');
    expect(logSpy).toHaveBeenCalledWith('[admin:info] hi');
  });

  it('always invokes registered sinks (even if they will be silenced in prod)', () => {
    const sink = vi.fn<(entry: LogEntry) => void>();
    const unregister = registerLogSink(sink);
    try {
      logger.error('x', { extra: true });
      expect(sink).toHaveBeenCalledTimes(1);
      const entry = sink.mock.calls[0]?.[0];
      expect(entry?.level).toBe('error');
      expect(entry?.message).toBe('x');
      expect(entry?.context).toEqual({ extra: true });
      expect(typeof entry?.timestamp).toBe('number');
    } finally {
      unregister();
    }
  });

  it('a misbehaving sink does not break the logger', () => {
    const unregister = registerLogSink(() => {
      throw new Error('sink boom');
    });
    try {
      // The wrapped try/catch should swallow the error.
      expect(() => logger.error('still logged')).not.toThrow();
      expect(errorSpy).toHaveBeenCalled();
    } finally {
      unregister();
    }
  });

  it('unregister removes the sink', () => {
    const sink = vi.fn<(entry: LogEntry) => void>();
    const unregister = registerLogSink(sink);
    logger.error('one');
    unregister();
    logger.error('two');
    expect(sink).toHaveBeenCalledTimes(1);
  });
});
