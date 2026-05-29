/** Platform admin roles allowed to access the admin dashboard SPA. */
const ADMIN_ROLES = new Set([
  'super_admin',
  'admin',
  'support',
  'billing_admin',
  'developer_admin',
]);

export function isAdminRole(role: string | undefined | null): boolean {
  if (!role) return false;
  return ADMIN_ROLES.has(role.toLowerCase());
}
