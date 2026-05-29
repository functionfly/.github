import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import type { Permission } from '@/hooks/useAccessControl';
import { Suspense, type ComponentType } from 'react';

interface AdminPageProps {
  component: ComponentType;
  permission?: Permission;
  featureName?: string;
}

export function AdminPage({ component: Page, permission, featureName }: AdminPageProps) {
  return (
    <ProtectedRoute requiredPermission={permission} featureName={featureName}>
      <Suspense fallback={<LoadingScreen />}>
        <Page />
      </Suspense>
    </ProtectedRoute>
  );
}
