/**
 * Helpers for platform staff accessing the standalone admin SPA (`web/admin-dashboard`, dev :3002,
 * production `admin.functionfly.com`). The main user dashboard links here via `ADMIN_DASHBOARD_URL`.
 */
import { ADMIN_DASHBOARD_URL } from '@/lib/constants';
import { tokenVault } from '@/utils/token-vault';
import { toast } from 'sonner';

/** Matches `internal/auth/roles.go` IsAdminRole — platform staff who may use the standalone admin app. */
const PLATFORM_ADMIN_ROLES = [
  'super_admin',
  'admin',
  'support',
  'billing_admin',
  'developer_admin',
] as const;

export function isPlatformAdminRole(role: string | undefined): boolean {
  if (!role) return false;
  return (PLATFORM_ADMIN_ROLES as readonly string[]).includes(role);
}

/** Read `role` from a JWT access token (same claim as Go `Claims.Role`). */
export function decodeJwtRole(token: string | null | undefined): string | undefined {
  if (!token) return undefined;
  try {
    const payload = JSON.parse(atob(token.split('.')[1])) as { role?: string };
    return typeof payload.role === 'string' ? payload.role : undefined;
  } catch {
    return undefined;
  }
}

/**
 * After email/OAuth/MFA sign-in, prompt platform staff to open the admin SPA (local :3002 or production host).
 * Uses JWT if `role` is missing (e.g. MFA path before user is hydrated).
 */
export async function notifyAdminPanelAfterLogin(role: string | undefined): Promise<void> {
  const resolved =
    role ??
    (await (async () => {
      await tokenVault.initialize();
      const token = await tokenVault.getAccessToken();
      return token ? decodeJwtRole(token) : undefined;
    })());
  if (!isPlatformAdminRole(resolved) || !ADMIN_DASHBOARD_URL) return;
  toast.success('Signed in', {
    description: 'Open the admin dashboard to manage the platform.',
    action: {
      label: 'Open admin dashboard',
      onClick: () => {
        window.location.assign(ADMIN_DASHBOARD_URL);
      },
    },
  });
}
