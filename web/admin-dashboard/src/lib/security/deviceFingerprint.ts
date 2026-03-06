/**
 * Device fingerprint helper.
 *
 * This is not intended as a cryptographic identity. It is a stable, best-effort
 * client fingerprint used for admin-session hardening signals.
 */

async function sha256Hex(input: string): Promise<string> {
  if (typeof crypto !== 'undefined' && crypto.subtle) {
    const data = new TextEncoder().encode(input);
    const hash = await crypto.subtle.digest('SHA-256', data);
    const bytes = Array.from(new Uint8Array(hash));
    return bytes.map((b) => b.toString(16).padStart(2, '0')).join('');
  }

  // Fallback for environments without SubtleCrypto.
  let h = 0;
  for (let i = 0; i < input.length; i += 1) {
    h = (Math.imul(31, h) + input.charCodeAt(i)) | 0;
  }
  return Math.abs(h).toString(16);
}

export async function generateDeviceFingerprint(): Promise<string> {
  const tz = Intl.DateTimeFormat().resolvedOptions().timeZone || 'unknown';
  const language = navigator.language || 'unknown';
  const languages = Array.isArray(navigator.languages) ? navigator.languages.join(',') : language;
  const platform = navigator.platform || 'unknown';
  const ua = navigator.userAgent || 'unknown';
  const screenPart = `${window.screen?.width || 0}x${window.screen?.height || 0}x${window.screen?.colorDepth || 0}`;
  const memory = (navigator as Navigator & { deviceMemory?: number }).deviceMemory ?? 0;
  const cpu = navigator.hardwareConcurrency || 0;

  const seed = [tz, language, languages, platform, ua, screenPart, String(memory), String(cpu)].join('|');
  return sha256Hex(seed);
}
