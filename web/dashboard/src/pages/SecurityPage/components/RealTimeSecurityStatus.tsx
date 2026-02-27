import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Activity, RefreshCw, Radar, ShieldCheck } from 'lucide-react';
import { ServiceStatus } from './ServiceStatus';
import { SSLCertificates } from './SSLCertificates';
import { RecentIncidents } from './RecentIncidents';
import type { ServiceStatus as ServiceStatusType, SSLCertificate, SecurityIncident } from '../types';

interface RealTimeSecurityStatusProps {
  serviceStatus: ServiceStatusType[];
  sslCertificates: SSLCertificate[];
  recentIncidents: SecurityIncident[];
}

export function RealTimeSecurityStatus({
  serviceStatus,
  sslCertificates,
  recentIncidents
}: RealTimeSecurityStatusProps) {
  return (
    <div className="space-y-6">
      {/* Status indicator */}
      <div className="flex items-center justify-end mb-4 md:mb-4">
        <div className="flex items-center gap-2 text-sm text-muted-foreground bg-muted/50 px-3 py-1 rounded-full">
          <RefreshCw className="h-4 w-4 animate-spin" />
          <span>Live • 30s updates</span>
        </div>
      </div>
        {/* Service Uptime Indicators */}
        <ServiceStatus services={serviceStatus} />

        {/* SSL Certificate Status */}
        <SSLCertificates certificates={sslCertificates} />

        {/* Recent Security Events */}
        <RecentIncidents incidents={recentIncidents} />
    </div>
  );
}