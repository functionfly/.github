import { Badge } from '@/components/ui/badge';
import { Wifi, WifiOff } from 'lucide-react';
import type { ServiceStatus as ServiceStatusType } from '../types';
import { RISK_LEVELS, getStatusRiskLevel } from '../utils/riskColors';

interface ServiceStatusProps {
  services: ServiceStatusType[];
}

export function ServiceStatus({ services }: ServiceStatusProps) {
  return (
    <div>
      <h4 className="font-semibold mb-4 flex items-center gap-2">
        <Wifi className="h-4 w-4" />
        Service Status
      </h4>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 md:gap-4">
        {services.map((service) => {
          const riskLevel = getStatusRiskLevel(service.status);
          const riskColors = RISK_LEVELS[riskLevel];

          return (
            <div key={service.name} className="flex items-center justify-between p-3 border rounded-lg" style={{ borderColor: riskColors.color + '20' }}>
              <div className="flex items-center gap-3">
                {service.status === 'operational' ? (
                  <Wifi className="h-4 w-4" style={{ color: riskColors.color }} />
                ) : (
                  <WifiOff className="h-4 w-4" style={{ color: riskColors.color }} />
                )}
                <div>
                  <p className="font-medium text-sm">{service.name}</p>
                  <p className="text-xs text-muted-foreground">{service.uptime} uptime</p>
                </div>
              </div>
              <div className="text-right">
                <Badge
                  variant={riskLevel === 'excellent' ? 'default' : riskLevel === 'critical' ? 'destructive' : 'secondary'}
                  style={{
                    backgroundColor: riskColors.bgColor,
                    color: riskColors.textColor,
                    borderColor: riskColors.color + '40'
                  }}
                >
                  {service.status}
                </Badge>
                <p className="text-xs text-muted-foreground mt-1">{service.responseTime}</p>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}