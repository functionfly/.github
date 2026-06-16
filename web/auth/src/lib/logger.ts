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

const STYLES = {
  base: 'font-family: "JetBrains Mono", monospace; font-size: 12px; padding: 2px 4px; border-radius: 3px;',
  debug: 'background: #21262D; color: #6E7681; border: 1px solid #30363D;',
  info: 'background: #00D4FF; color: #0D1117; border: 1px solid #00D4FF; font-weight: bold;',
  warn: 'background: #FF9500; color: #0D1117; border: 1px solid #FF9500; font-weight: bold;',
  error: 'background: #FF2D55; color: #FFFFFF; border: 1px solid #FF2D55; font-weight: bold;',
  debugMsg: 'color: #6E7681;',
  infoMsg: 'color: #00D4FF;',
  warnMsg: 'color: #FF9500;',
  errorMsg: 'color: #FF2D55; font-weight: bold;',
  timestamp: 'color: #484F58; font-size: 10px; font-style: italic;',
};

function createEntry(level: LogLevel, message: string, args: unknown[]): LogEntry {
  return { level, message, args, timestamp: new Date().toISOString() };
}

function buffer(entry: LogEntry) {
  logBuffer.push(entry);
  if (logBuffer.length > MAX_BUFFER_SIZE) logBuffer.shift();
}

function formatTime(): string {
  const now = new Date();
  return `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}.${now.getMilliseconds().toString().padStart(3, '0')}`;
}

function styledLog(level: LogLevel, consoleFn: typeof console.log, badge: string, badgeStyle: string, msgStyle: string, message: string, args: unknown[]) {
  const time = formatTime();
  if (args.length === 0) {
    consoleFn(`%c${badge}%c ${time} %c${message}`, `${STYLES.base} ${badgeStyle}`, STYLES.timestamp, msgStyle);
  } else {
    consoleFn(`%c${badge}%c ${time} %c${message}`, `${STYLES.base} ${badgeStyle}`, STYLES.timestamp, msgStyle, ...args);
  }
}

export const logger = {
  debug(message: string, ...args: unknown[]) {
    const entry = createEntry('debug', message, args);
    buffer(entry);
    if (isDev) styledLog('debug', console.debug, 'DEBUG', STYLES.debug, STYLES.debugMsg, message, args);
  },
  info(message: string, ...args: unknown[]) {
    const entry = createEntry('info', message, args);
    buffer(entry);
    if (isDev) styledLog('info', console.info, 'INFO', STYLES.info, STYLES.infoMsg, message, args);
  },
  warn(message: string, ...args: unknown[]) {
    const entry = createEntry('warn', message, args);
    buffer(entry);
    styledLog('warn', console.warn, 'WARN', STYLES.warn, STYLES.warnMsg, message, args);
  },
  error(message: string, ...args: unknown[]) {
    const entry = createEntry('error', message, args);
    buffer(entry);
    styledLog('error', console.error, 'ERROR', STYLES.error, STYLES.errorMsg, message, args);
  },
};

export function getLogBuffer(): ReadonlyArray<LogEntry> {
  return logBuffer;
}
