// web/auth/src/utility/NetworkCheck.ts
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
    const response = await fetch('/api/health', {
      method: 'HEAD',
      cache: 'no-cache',
      signal: AbortSignal.timeout(timeoutMs),
    });
    return response.ok;
  } catch (error) {
    return false;
  }
}