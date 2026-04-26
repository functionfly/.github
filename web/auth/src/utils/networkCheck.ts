// web/auth/src/utils/networkCheck.ts
import { API_ORIGIN } from '../config';

/**
 * Performs a reliable network connectivity check
 *
 * @param timeoutMs Timeout in milliseconds for the connectivity check (default: 3000)
 * @returns Promise<boolean> True if online and can reach our servers, false otherwise
 */
export async function checkNetworkConnectivity(timeoutMs = 3000): Promise<boolean> {
  if (!navigator.onLine) {
    return false;
  }

  try {
    const response = await fetch(`${API_ORIGIN}/health`, {
      method: 'GET',
      cache: 'no-cache',
      signal: AbortSignal.timeout(timeoutMs),
    });
    return response.ok;
  } catch (error) {
    return false;
  }
}

export const DEFAULT_REQUEST_TIMEOUT = 10000;

export async function fetchWithTimeout(
  url: string,
  options: RequestInit & { timeout?: number } = {}
): Promise<Response> {
  const { timeout = DEFAULT_REQUEST_TIMEOUT, ...fetchOptions } = options;
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);

  try {
    const response = await fetch(url, {
      ...fetchOptions,
      signal: controller.signal,
    });
    return response;
  } finally {
    clearTimeout(timeoutId);
  }
}