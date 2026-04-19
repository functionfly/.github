const isDev = import.meta.env.DEV;

export type LogLevel = 'debug' | 'info' | 'warn' | 'error' | 'success' | 'api' | 'ws' | 'auth';

interface LogEntry {
  level: LogLevel;
  message: string;
  args: unknown[];
  timestamp: string;
}

const logBuffer: LogEntry[] = [];
const MAX_BUFFER_SIZE = 100;

// Velocity Brand Colors for Console Styling
const STYLES = {
  // Base styles
  base: 'font-family: "JetBrains Mono", monospace; font-size: 12px; padding: 2px 4px; border-radius: 3px;',

  // Log level badges
  debug: 'background: #21262D; color: #6E7681; border: 1px solid #30363D;',
  info: 'background: #00D4FF; color: #0D1117; border: 1px solid #00D4FF; font-weight: bold;',
  warn: 'background: #FF9500; color: #0D1117; border: 1px solid #FF9500; font-weight: bold;',
  error: 'background: #FF2D55; color: #FFFFFF; border: 1px solid #FF2D55; font-weight: bold;',
  success: 'background: #00FF9D; color: #0D1117; border: 1px solid #00FF9D; font-weight: bold;',

  // Domain badges
  api: 'background: #FF6B35; color: #FFFFFF; border: 1px solid #FF6B35; font-weight: bold;',
  ws: 'background: #00D4FF; color: #0D1117; border: 1px solid #00D4FF; font-weight: bold;',
  auth: 'background: #FF4F5E; color: #FFFFFF; border: 1px solid #FF4F5E; font-weight: bold;',

  // Message styles by level
  debugMsg: 'color: #6E7681; font-weight: normal;',
  infoMsg: 'color: #00D4FF; font-weight: normal;',
  warnMsg: 'color: #FF9500; font-weight: normal;',
  errorMsg: 'color: #FF2D55; font-weight: bold;',
  successMsg: 'color: #00FF9D; font-weight: normal;',
  apiMsg: 'color: #FF6B35; font-weight: normal;',
  wsMsg: 'color: #00D4FF; font-weight: normal;',
  authMsg: 'color: #FF4F5E; font-weight: normal;',

  // Timestamp
  timestamp: 'color: #484F58; font-size: 10px; font-style: italic;',

  // Object preview box
  objectBox: 'background: #161B22; color: #F0F6FC; border: 1px solid #30363D; border-radius: 4px; padding: 4px 8px;',
};

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

function formatTime(): string {
  const now = new Date();
  return `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}.${now.getMilliseconds().toString().padStart(3, '0')}`;
}

function styledLog(
  level: LogLevel,
  consoleFn: typeof console.log,
  badge: string,
  badgeStyle: string,
  msgStyle: string,
  message: string,
  args: unknown[]
) {
  const time = formatTime();

  if (args.length === 0) {
    consoleFn(
      `%c${badge}%c ${time} %c${message}`,
      `${STYLES.base} ${badgeStyle}`,
      STYLES.timestamp,
      msgStyle
    );
  } else {
    consoleFn(
      `%c${badge}%c ${time} %c${message}`,
      `${STYLES.base} ${badgeStyle}`,
      STYLES.timestamp,
      msgStyle,
      ...args
    );
  }
}

function makeTagged(
  level: LogLevel,
  badge: string,
  badgeStyle: string,
  msgStyle: string,
  consoleFn: typeof console.log,
) {
  return (strings: TemplateStringsArray, ...values: unknown[]) => {
    const message = strings.reduce(
      (acc, s, i) => acc + s + (i < values.length ? String(values[i]) : ''),
      '',
    );
    const entry = createEntry(level, message, values);
    buffer(entry);
    if (isDev || level === 'error' || level === 'warn') {
      styledLog(level, consoleFn, badge, badgeStyle, msgStyle, message, values);
    }
  };
}

export const logger = {
  debug(message: string, ...args: unknown[]) {
    const entry = createEntry('debug', message, args);
    buffer(entry);
    if (isDev) {
      styledLog('debug', console.debug, 'DEBUG', STYLES.debug, STYLES.debugMsg, message, args);
    }
  },

  log(message: string, ...args: unknown[]) {
    const entry = createEntry('info', message, args);
    buffer(entry);
    if (isDev) {
      styledLog('info', console.log, 'INFO', STYLES.info, STYLES.infoMsg, message, args);
    }
  },

  info(message: string, ...args: unknown[]) {
    const entry = createEntry('info', message, args);
    buffer(entry);
    if (isDev) {
      styledLog('info', console.info, 'INFO', STYLES.info, STYLES.infoMsg, message, args);
    }
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

  success(message: string, ...args: unknown[]) {
    const entry = createEntry('success', message, args);
    buffer(entry);
    if (isDev) {
      styledLog('success', console.log, 'OK', STYLES.success, STYLES.successMsg, message, args);
    }
  },

  api(message: string, ...args: unknown[]) {
    const entry = createEntry('api', message, args);
    buffer(entry);
    if (isDev) {
      styledLog('api', console.log, 'API', STYLES.api, STYLES.apiMsg, message, args);
    }
  },

  ws(message: string, ...args: unknown[]) {
    const entry = createEntry('ws', message, args);
    buffer(entry);
    if (isDev) {
      styledLog('ws', console.log, 'WS', STYLES.ws, STYLES.wsMsg, message, args);
    }
  },

  auth(message: string, ...args: unknown[]) {
    const entry = createEntry('auth', message, args);
    buffer(entry);
    if (isDev) {
      styledLog('auth', console.log, 'AUTH', STYLES.auth, STYLES.authMsg, message, args);
    }
  },

  group(label: string, collapsed = false) {
    const fn = collapsed ? console.groupCollapsed : console.group;
    fn(`%c ${label} `, `${STYLES.base} background: #21262D; color: #F0F6FC; border: 1px solid #30363D;`);
  },

  groupEnd() {
    console.groupEnd();
  },

  table(data: unknown[]) {
    console.table(data);
  },

  // Tagged-template helpers with full styling
  debug$: makeTagged('debug', 'DEBUG', STYLES.debug, STYLES.debugMsg, console.debug),
  log$: makeTagged('info', 'INFO', STYLES.info, STYLES.infoMsg, console.log),
  info$: makeTagged('info', 'INFO', STYLES.info, STYLES.infoMsg, console.info),
  warn$: makeTagged('warn', 'WARN', STYLES.warn, STYLES.warnMsg, console.warn),
  error$: makeTagged('error', 'ERROR', STYLES.error, STYLES.errorMsg, console.error),
  success$: makeTagged('success', 'OK', STYLES.success, STYLES.successMsg, console.log),
  api$: makeTagged('api', 'API', STYLES.api, STYLES.apiMsg, console.log),
  ws$: makeTagged('ws', 'WS', STYLES.ws, STYLES.wsMsg, console.log),
  auth$: makeTagged('auth', 'AUTH', STYLES.auth, STYLES.authMsg, console.log),
};

/** Quick console styling for inline usage */
export const cc = {
  /** Log with flame (brand) styling: cc.flame('message', data) */
  flame: (msg: string, ...args: unknown[]) => {
    console.log(`%c${msg}`, 'color: #FF6B35; font-weight: bold; font-family: "JetBrains Mono", monospace;', ...args);
  },
  /** Log with cyan styling: cc.cyan('message', data) */
  cyan: (msg: string, ...args: unknown[]) => {
    console.log(`%c${msg}`, 'color: #00D4FF; font-weight: bold; font-family: "JetBrains Mono", monospace;', ...args);
  },
  /** Log with success styling: cc.success('message', data) */
  success: (msg: string, ...args: unknown[]) => {
    console.log(`%c${msg}`, 'color: #00FF9D; font-weight: bold; font-family: "JetBrains Mono", monospace;', ...args);
  },
  /** Log with error styling: cc.error('message', data) */
  error: (msg: string, ...args: unknown[]) => {
    console.log(`%c${msg}`, 'color: #FF2D55; font-weight: bold; font-family: "JetBrains Mono", monospace;', ...args);
  },
  /** Log with warning styling: cc.warn('message', data) */
  warn: (msg: string, ...args: unknown[]) => {
    console.log(`%c${msg}`, 'color: #FF9500; font-weight: bold; font-family: "JetBrains Mono", monospace;', ...args);
  },
  /** Log with object preview box */
  box: (label: string, data: unknown) => {
    console.log(`%c ${label} `, `${STYLES.base} ${STYLES.objectBox}`, data);
  },
  /** Styled badge for custom domains */
  badge: (badge: string, color: string, msg: string, ...args: unknown[]) => {
    console.log(
      `%c${badge}%c ${msg}`,
      `${STYLES.base} background: ${color}; color: #0D1117; border: 1px solid ${color}; font-weight: bold;`,
      'color: ' + color + '; font-family: "JetBrains Mono", monospace;',
      ...args
    );
  },
};

export function getLogBuffer(): ReadonlyArray<LogEntry> {
  return logBuffer;
}

export function clearLogBuffer() {
  logBuffer.length = 0;
}
