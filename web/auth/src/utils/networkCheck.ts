// web/auth/src/utility/NetworkCheck.ts
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