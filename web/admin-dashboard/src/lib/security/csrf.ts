/**
 * CSRF token helpers for admin mutating requests.
 */

function readCookie(name: string): string | null {
  const encodedName = `${encodeURIComponent(name)}=`;
  const parts = document.cookie.split(';');

  for (const part of parts) {
    const trimmed = part.trim();
    if (trimmed.startsWith(encodedName)) {
      return decodeURIComponent(trimmed.slice(encodedName.length));
    }
  }

  return null;
}

export function getCsrfToken(): string | null {
  const fromMeta = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content');
  if (fromMeta) {
    return fromMeta;
  }

  const fromCookie = readCookie('csrf_token') || readCookie('XSRF-TOKEN');
  if (fromCookie) {
    return fromCookie;
  }

  return null;
}
