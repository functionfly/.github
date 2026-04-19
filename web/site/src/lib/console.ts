// Console styling utility - Velocity Brand Colors
// Usage: cc.flame('message'), cc.cyan('message'), cc.badge('API', '#FF6B35', 'Request sent')

const BASE_STYLE = 'font-family: "JetBrains Mono", monospace; font-size: 12px; padding: 2px 4px; border-radius: 3px;';

const BADGE_STYLES: Record<string, string> = {
  debug: 'background: #21262D; color: #6E7681; border: 1px solid #30363D;',
  info: 'background: #00D4FF; color: #0D1117; border: 1px solid #00D4FF; font-weight: bold;',
  warn: 'background: #FF9500; color: #0D1117; border: 1px solid #FF9500; font-weight: bold;',
  error: 'background: #FF2D55; color: #FFFFFF; border: 1px solid #FF2D55; font-weight: bold;',
  success: 'background: #00FF9D; color: #0D1117; border: 1px solid #00FF9D; font-weight: bold;',
  flame: 'background: #FF6B35; color: #FFFFFF; border: 1px solid #FF6B35; font-weight: bold;',
  cyan: 'background: #00D4FF; color: #0D1117; border: 1px solid #00D4FF; font-weight: bold;',
  afterburner: 'background: #FF4F5E; color: #FFFFFF; border: 1px solid #FF4F5E; font-weight: bold;',
};

const MSG_STYLES: Record<string, string> = {
  debug: 'color: #6E7681;',
  info: 'color: #00D4FF;',
  warn: 'color: #FF9500;',
  error: 'color: #FF2D55; font-weight: bold;',
  success: 'color: #00FF9D;',
  flame: 'color: #FF6B35;',
  cyan: 'color: #00D4FF;',
  afterburner: 'color: #FF4F5E;',
};

function formatTime(): string {
  const now = new Date();
  return `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`;
}

/** Styled console logger - Velocity Brand colors */
export const cc = {
  /** Debug message with gray styling */
  debug: (msg: string, ...args: unknown[]) => {
    const time = formatTime();
    console.log(
      `%cDEBUG%c ${time} %c${msg}`,
      `${BASE_STYLE} ${BADGE_STYLES.debug}`,
      'color: #484F58; font-size: 10px; font-style: italic;',
      `${MSG_STYLES.debug} ${BASE_STYLE}`,
      ...args
    );
  },

  /** Info message with cyan styling */
  info: (msg: string, ...args: unknown[]) => {
    const time = formatTime();
    console.log(
      `%cINFO%c ${time} %c${msg}`,
      `${BASE_STYLE} ${BADGE_STYLES.info}`,
      'color: #484F58; font-size: 10px; font-style: italic;',
      `${MSG_STYLES.info} ${BASE_STYLE}`,
      ...args
    );
  },

  /** Warning message with amber styling */
  warn: (msg: string, ...args: unknown[]) => {
    const time = formatTime();
    console.warn(
      `%cWARN%c ${time} %c${msg}`,
      `${BASE_STYLE} ${BADGE_STYLES.warn}`,
      'color: #484F58; font-size: 10px; font-style: italic;',
      `${MSG_STYLES.warn} ${BASE_STYLE}`,
      ...args
    );
  },

  /** Error message with red styling */
  error: (msg: string, ...args: unknown[]) => {
    const time = formatTime();
    console.error(
      `%cERROR%c ${time} %c${msg}`,
      `${BASE_STYLE} ${BADGE_STYLES.error}`,
      'color: #484F58; font-size: 10px; font-style: italic;',
      `${MSG_STYLES.error} ${BASE_STYLE}`,
      ...args
    );
  },

  /** Success message with green styling */
  success: (msg: string, ...args: unknown[]) => {
    const time = formatTime();
    console.log(
      `%cOK%c ${time} %c${msg}`,
      `${BASE_STYLE} ${BADGE_STYLES.success}`,
      'color: #484F58; font-size: 10px; font-style: italic;',
      `${MSG_STYLES.success} ${BASE_STYLE}`,
      ...args
    );
  },

  /** Flame brand color message */
  flame: (msg: string, ...args: unknown[]) => {
    console.log(`%c${msg}`, `${MSG_STYLES.flame} font-weight: bold; ${BASE_STYLE}`, ...args);
  },

  /** Cyan accent color message */
  cyan: (msg: string, ...args: unknown[]) => {
    console.log(`%c${msg}`, `${MSG_STYLES.cyan} font-weight: bold; ${BASE_STYLE}`, ...args);
  },

  /** Afterburner red message */
  afterburner: (msg: string, ...args: unknown[]) => {
    console.log(`%c${msg}`, `${MSG_STYLES.afterburner} font-weight: bold; ${BASE_STYLE}`, ...args);
  },

  /** Create a custom styled badge */
  badge: (badgeText: string, color: string, msg: string, ...args: unknown[]) => {
    const time = formatTime();
    console.log(
      `%c${badgeText}%c ${time} %c${msg}`,
      `${BASE_STYLE} background: ${color}; color: ${isLightColor(color) ? '#0D1117' : '#FFFFFF'}; border: 1px solid ${color}; font-weight: bold;`,
      'color: #484F58; font-size: 10px; font-style: italic;',
      `color: ${color}; ${BASE_STYLE}`,
      ...args
    );
  },

  /** Group related logs */
  group: (label: string, collapsed = false) => {
    const fn = collapsed ? console.groupCollapsed : console.group;
    fn(`%c ${label} `, `${BASE_STYLE} background: #21262D; color: #F0F6FC; border: 1px solid #30363D;`);
  },

  groupEnd: () => console.groupEnd(),

  /** Display data in a styled box */
  box: (label: string, data: unknown) => {
    console.log(
      `%c ${label} `,
      `${BASE_STYLE} background: #161B22; color: #F0F6FC; border: 1px solid #30363D; border-radius: 4px; padding: 4px 8px;`,
      data
    );
  },

  /** API call logging */
  api: (method: string, url: string, data?: unknown) => {
    const time = formatTime();
    const methodColors: Record<string, string> = {
      GET: '#00FF9D',
      POST: '#FF6B35',
      PUT: '#00D4FF',
      PATCH: '#FF9500',
      DELETE: '#FF2D55',
    };
    const color = methodColors[method] || '#F0F6FC';
    console.log(
      `%c${method}%c ${time} %c${url}`,
      `${BASE_STYLE} background: ${color}; color: #0D1117; border: 1px solid ${color}; font-weight: bold;`,
      'color: #484F58; font-size: 10px; font-style: italic;',
      `${BASE_STYLE} color: ${color};`,
      data ? '\n' : '',
      data || ''
    );
  },
};

// Helper to determine if a color is light or dark
function isLightColor(hex: string): boolean {
  const clean = hex.replace('#', '');
  const r = parseInt(clean.slice(0, 2), 16);
  const g = parseInt(clean.slice(2, 4), 16);
  const b = parseInt(clean.slice(4, 6), 16);
  const brightness = (r * 299 + g * 587 + b * 114) / 1000;
  return brightness > 128;
}

// Make available globally for browser console usage
if (typeof window !== 'undefined') {
  (window as unknown as Record<string, unknown>).cc = cc;
}
