/**
 * Security Event Monitoring
 * Tracks security-related client events and reports suspicious activity to backend.
 */

import { adminApiClient } from '@/lib/api/adminClient';

// Small helper to build DOM elements with inline styles. Using DOM APIs keeps
// the markup out of HTML strings so CSP doesn't need 'unsafe-inline' for the
// React app, and so any future dynamic content cannot accidentally become
// script injection.
function el<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  options: { className?: string; styles?: Partial<CSSStyleDeclaration>; text?: string } = {}
): HTMLElementTagNameMap[K] {
  const node = document.createElement(tag);
  if (options.className) node.className = options.className;
  if (options.styles) {
    Object.assign(node.style, options.styles);
  }
  if (options.text != null) node.textContent = options.text;
  return node;
}

export type SecurityEventType =
  | 'login_success'
  | 'login_failed'
  | 'logout'
  | 'session_expired'
  | 'mfa_verified'
  | 'mfa_failed'
  | 'mfa_verify_failed'
  | 'mfa_required'
  | 'ip_blocked'
  | 'suspicious_activity'
  | 'session_timeout_warning'
  | 'new_device_detected'
  | 'fingerprint_mismatch';

export interface SecurityEvent {
  type: SecurityEventType;
  timestamp: number;
  metadata?: Record<string, unknown>;
}

export interface SecurityEventPayload {
  event_type: SecurityEventType;
  timestamp: string;
  ip_address?: string;
  user_agent?: string;
  device_fingerprint?: string;
  session_id?: string;
  metadata?: Record<string, unknown>;
}

/**
 * In-memory event buffer to batch reports
 */
const eventBuffer: SecurityEvent[] = [];
const MAX_BUFFER_SIZE = 10;
let flushScheduled = false;

/**
 * Track a security event locally and queue for backend reporting.
 */
export function trackSecurityEvent(
  type: SecurityEventType,
  metadata?: Record<string, unknown>
): void {
  const event: SecurityEvent = {
    type,
    timestamp: Date.now(),
    metadata,
  };

  eventBuffer.push(event);

  if (eventBuffer.length >= MAX_BUFFER_SIZE) {
    flushEvents();
  } else if (!flushScheduled) {
    // Debounce flush to avoid excessive requests
    flushScheduled = true;
    setTimeout(() => {
      flushScheduled = false;
      flushEvents();
    }, 5000);
  }
}

/**
 * Flush buffered events to backend.
 */
async function flushEvents(): Promise<void> {
  if (eventBuffer.length === 0) return;

  const events = eventBuffer.splice(0, eventBuffer.length);

  try {
    await adminApiClient.post('/security/events', {
      events: events.map((e) => ({
        event_type: e.type,
        timestamp: new Date(e.timestamp).toISOString(),
        metadata: e.metadata,
      })),
    });
  } catch {
    // Silently fail - security events should not break UX
    // Re-queue critical events if needed in production
  }
}

/**
 * Show session timeout warning to user.
 * Returns a cleanup function.
 */
export function showSessionTimeoutWarning(
  onContinue: () => void,
  onLogout: () => void,
  timeoutMs = 60000 // 1 minute warning
): () => void {
  const overlay = document.createElement('div');
  overlay.style.cssText = `
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
  `;

  const card = document.createElement('div');
  Object.assign(card.style, {
    background: 'white',
    padding: '24px',
    borderRadius: '12px',
    maxWidth: '400px',
    width: '90%',
    boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
  });

  const heading = el('h2', {
    styles: { margin: '0 0 12px', fontSize: '18px', fontWeight: '600', color: '#111' },
    text: 'Session Expiring Soon',
  });
  const body = el('p', {
    styles: { margin: '0 0 20px', color: '#666', fontSize: '14px' },
    text: 'Your session will expire in less than 1 minute due to inactivity. Would you like to continue?',
  });

  const continueBtn = el('button', {
    styles: {
      flex: '1',
      padding: '10px 16px',
      background: '#2563eb',
      color: 'white',
      border: 'none',
      borderRadius: '6px',
      fontSize: '14px',
      fontWeight: '500',
      cursor: 'pointer',
    },
    text: 'Continue Session',
  });
  continueBtn.id = 'session-continue';

  const logoutBtn = el('button', {
    styles: {
      flex: '1',
      padding: '10px 16px',
      background: '#f3f4f6',
      color: '#374151',
      border: 'none',
      borderRadius: '6px',
      fontSize: '14px',
      fontWeight: '500',
      cursor: 'pointer',
    },
    text: 'Logout',
  });
  logoutBtn.id = 'session-logout';

  const row = el('div', { styles: { display: 'flex', gap: '12px' } });
  row.appendChild(continueBtn);
  row.appendChild(logoutBtn);

  card.appendChild(heading);
  card.appendChild(body);
  card.appendChild(row);

  overlay.appendChild(card);
  document.body.appendChild(overlay);

  const handleContinue = () => {
    document.body.removeChild(overlay);
    onContinue();
  };

  const handleLogout = () => {
    document.body.removeChild(overlay);
    onLogout();
  };

  continueBtn.addEventListener('click', handleContinue);
  logoutBtn.addEventListener('click', handleLogout);

  const timeoutId = setTimeout(handleLogout, timeoutMs);

  // Return cleanup function
  return () => {
    clearTimeout(timeoutId);
    if (document.body.contains(overlay)) {
      document.body.removeChild(overlay);
    }
  };
}

/**
 * Request MFA re-verification.
 * Returns a cleanup function.
 */
export function showMFAReverificationPrompt(
  onVerified: () => void,
  onCancel: () => void
): () => void {
  const overlay = document.createElement('div');
  overlay.style.cssText = `
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 9999;
  `;

  const card = document.createElement('div');
  Object.assign(card.style, {
    background: 'white',
    padding: '24px',
    borderRadius: '12px',
    maxWidth: '400px',
    width: '90%',
    boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
  });

  const heading = el('h2', {
    styles: { margin: '0 0 12px', fontSize: '18px', fontWeight: '600', color: '#111' },
    text: 'Re-verify Your Identity',
  });
  const body = el('p', {
    styles: { margin: '0 0 20px', color: '#666', fontSize: '14px' },
    text: 'For your security, please verify your identity again to continue.',
  });

  const verifyBtn = el('button', {
    styles: {
      flex: '1',
      padding: '10px 16px',
      background: '#2563eb',
      color: 'white',
      border: 'none',
      borderRadius: '6px',
      fontSize: '14px',
      fontWeight: '500',
      cursor: 'pointer',
    },
    text: 'Verify Now',
  });
  verifyBtn.id = 'mfa-verify';

  const cancelBtn = el('button', {
    styles: {
      flex: '1',
      padding: '10px 16px',
      background: '#f3f4f6',
      color: '#374151',
      border: 'none',
      borderRadius: '6px',
      fontSize: '14px',
      fontWeight: '500',
      cursor: 'pointer',
    },
    text: 'Cancel',
  });
  cancelBtn.id = 'mfa-cancel';

  const row = el('div', { styles: { display: 'flex', gap: '12px' } });
  row.appendChild(verifyBtn);
  row.appendChild(cancelBtn);

  card.appendChild(heading);
  card.appendChild(body);
  card.appendChild(row);

  overlay.appendChild(card);
  document.body.appendChild(overlay);

  const handleVerify = () => {
    document.body.removeChild(overlay);
    onVerified();
  };

  const handleCancel = () => {
    document.body.removeChild(overlay);
    onCancel();
  };

  verifyBtn.addEventListener('click', handleVerify);
  cancelBtn.addEventListener('click', handleCancel);

  // Return cleanup function
  return () => {
    if (document.body.contains(overlay)) {
      document.body.removeChild(overlay);
    }
  };
}

/**
 * Report suspicious login activity to backend.
 */
export async function reportSuspiciousLogin(
  sessionId: string,
  details: {
    ip_address?: string;
    device_fingerprint?: string;
    user_agent?: string;
    reason: string;
  }
): Promise<void> {
  try {
    await adminApiClient.post('/security/report-suspicious-login', {
      session_id: sessionId,
      ...details,
    });
  } catch {
    // Silently fail
  }
}
