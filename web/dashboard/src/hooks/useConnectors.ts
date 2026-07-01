import { useQuery } from '@tanstack/react-query';
import { connectorsApi, type UserConnector } from '@/api/connectors';

export const connectorKeys = {
  all: ['connectors'] as const,
  catalog: () => [...connectorKeys.all, 'catalog'] as const,
  user: () => [...connectorKeys.all, 'user'] as const,
};

export function useConnectorCatalog() {
  return useQuery({
    queryKey: connectorKeys.catalog(),
    queryFn: () => connectorsApi.listCatalog(),
  });
}

export function useUserConnectors() {
  return useQuery({
    queryKey: connectorKeys.user(),
    queryFn: () => connectorsApi.listUserConnectors(),
  });
}

export interface ConnectorStatus {
  name: string;
  slug: string;
  connected: boolean;
  userConnector?: UserConnector;
}

export function useConnectorStatuses() {
  const catalog = useConnectorCatalog();
  const userConnectors = useUserConnectors();

  const statuses: ConnectorStatus[] = (catalog.data ?? []).map((conn) => {
    const uc = (userConnectors.data ?? []).find(
      (u) => u.connector_id === conn.id || u.connector_slug === conn.slug
    );
    return {
      name: conn.name,
      slug: conn.slug,
      connected: !!uc && uc.status !== 'disabled' && uc.status !== 'revoked',
      userConnector: uc,
    };
  });

  return {
    statuses,
    isLoading: catalog.isLoading || userConnectors.isLoading,
  };
}
