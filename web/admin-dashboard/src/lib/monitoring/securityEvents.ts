/**
 * Security Event Monitoring
 * Tracks security-related client events and reports suspicious activity to backend.
 */

import { adminApiClient } from '@/lib/api/adminClient';

export type SecurityEventType =
  | 'login_success'
  | 'login_failed'
  | 'logout'
  | 'session_expired'
  | 'mfa_verified'
  | 'mfa_failed'
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
  card.style.cssText = `
    background: white;
    padding: 24px;
    border-radius: 12px;
    max-width: 400px;
    width: 90%;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  `;

  card.innerHTML = `
    <h2 style="margin: 0 0 12px; font-size: 18px; font-weight: 600; color: #111;">
      Session Expiring Soon
    </h2>
    <p style="margin: 0 0 20px; color: #666; font-size: 14px;">
      Your session will expire in less than 1 minute due to inactivity.
      Would you like to continue?
    </p>
    <div style="display: flex; gap: 12px;">
      <button id="session-continue" style="
        flex: 1;
        padding: 10px 16px;
        background: #2563eb;
        color: white;
        border: none;
        border-radius: 6px;
        font-size: 14px;
        font-weight: 500;
        cursor: pointer;
      ">Continue Session</button>
      <button id="session-logout" style="
        flex: 1;
        padding: 10px 16px;
        background: #f3f4f6;
        color: #374151;
        border: none;
        border-radius: 6px;
        font-size: 14px;
        font-weight: 500;
        cursor: pointer;
      ">Logout</button>
    </div>
  `;

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

  card.querySelector('#session-continue')?.addEventListener('click', handleContinue);
  card.querySelector('#session-logout')?.addEventListener('click', handleLogout);

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
  card.style.cssText = `
    background: white;
    padding: 24px;
    border-radius: 12px;
    max-width: 400px;
    width: 90%;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  `;

  card.innerHTML = `
    <h2 style="margin: 0 0 12px; font-size: 18px; font-weight: 600; color: #111;">
      Re-verify Your Identity
    </h2>
    <p style="margin: 0 0 20px; color: #666; font-size: 14px;">
      For your security, please verify your identity again to continue.
    </p>
    <div style="display: flex; gap: 12px;">
      <button id="mfa-verify" style="
        flex: 1;
        padding: 10px 16px;
        background: #2563eb;
        color: white;
        border: none;
        border-radius: 6px;
        font-size: 14px;
        font-weight: 500;
        cursor: pointer;
      ">Verify Now</button>
      <button id="mfa-cancel" style="
        flex: 1;
        padding: 10px 16px;
        background: #f3f4f6;
        color: #374151;
        border: none;
        border-radius: 6px;
        font-size: 14px;
        font-weight: 500;
        cursor: pointer;
      ">Cancel</button>
    </div>
  `;

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

  card.querySelector('#mfa-verify')?.addEventListener('click', handleVerify);
  card.querySelector('#mfa-cancel')?.addEventListener('click', handleCancel);

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
