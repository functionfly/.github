const isDev = import.meta.env.DEV;

export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

interface LogEntry {
  level: LogLevel;
  message: string;
  args: unknown[];
  timestamp: string;
}

const logBuffer: LogEntry[] = [];
const MAX_BUFFER_SIZE = 100;

function createEntry(level: LogLevel, message: string, args: unknown[]): LogEntry {
  return {
    level,
    message,
    args,
    timestamp: new Date().toISOString(),
  };
}

function buffer(entry: LogEntry) {
  logBuffer.push(entry);
  if (logBuffer.length > MAX_BUFFER_SIZE) {
    logBuffer.shift();
  }
}

function makeTagged(
  level: LogLevel,
  consoleFn: typeof console.log,
) {
  return (strings: TemplateStringsArray, ...values: unknown[]) => {
    const message = strings.reduce(
      (acc, s, i) => acc + s + (i < values.length ? String(values[i]) : ''),
      '',
    );
    const entry = createEntry(level, message, values);
    buffer(entry);
    if (isDev) {
      consoleFn(message);
    }
  };
}

export const logger = {
  debug(message: string, ...args: unknown[]) {
    const entry = createEntry('debug', message, args);
    buffer(entry);
    if (isDev) {
      console.debug(message, ...args);
    }
  },

  log(message: string, ...args: unknown[]) {
    const entry = createEntry('info', message, args);
    buffer(entry);
    if (isDev) {
      console.log(message, ...args);
    }
  },

  info(message: string, ...args: unknown[]) {
    const entry = createEntry('info', message, args);
    buffer(entry);
    if (isDev) {
      console.info(message, ...args);
    }
  },

  warn(message: string, ...args: unknown[]) {
    const entry = createEntry('warn', message, args);
    buffer(entry);
    console.warn(message, ...args);
  },

  error(message: string, ...args: unknown[]) {
    const entry = createEntry('error', message, args);
    buffer(entry);
    console.error(message, ...args);
  },

  /** Tagged-template helpers: logger.debug$`message ${val}` */
  debug$: makeTagged('debug', console.debug),
  log$: makeTagged('info', console.log),
  info$: makeTagged('info', console.info),
  warn$: makeTagged('warn', console.warn),
  error$: makeTagged('error', console.error),
};

export function getLogBuffer(): ReadonlyArray<LogEntry> {
  return logBuffer;
}

export function clearLogBuffer() {
  logBuffer.length = 0;
}
