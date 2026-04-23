import { Badge } from '@/components/ui/badge';
import { AlertTriangle, CheckCircle } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import type { SecurityIncident } from '../types';
import { RISK_LEVELS, getSeverityRiskLevel, getStatusRiskLevel } from '../utils/riskColors';

interface RecentIncidentsProps {
  incidents: SecurityIncident[];
}

export function RecentIncidents({ incidents }: RecentIncidentsProps) {
  const { t } = useTranslation();
  return (
    <div>
      <h4 className="font-semibold mb-4 flex items-center gap-2">
        <AlertTriangle className="h-4 w-4" />
        {t('securityPage.recentSecurityEvents')}
      </h4>
      <div className="space-y-3">
        {incidents.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <CheckCircle className="h-8 w-8 mx-auto mb-2 text-green-500" />
            <p>{t('securityPage.noSecurityIncidents')}</p>
          </div>
        ) : (
          incidents.map((incident) => {
            const severityRiskLevel = getSeverityRiskLevel(incident.severity);
            const statusRiskLevel = getStatusRiskLevel(incident.status);
            const severityColors = RISK_LEVELS[severityRiskLevel];
            const statusColors = RISK_LEVELS[statusRiskLevel];

            return (
              <div key={incident.id} className="border rounded-lg p-4" style={{ borderColor: severityColors.color + '20' }}>
                <div className="flex items-start justify-between mb-2">
                  <div>
                    <h5 className="font-medium">{incident.title}</h5>
                    <p className="text-sm text-muted-foreground">{incident.id}</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge
                      variant={severityRiskLevel === 'critical' ? 'destructive' : severityRiskLevel === 'warning' ? 'secondary' : 'outline'}
                      style={{
                        backgroundColor: severityColors.bgColor,
                        color: severityColors.textColor,
                        borderColor: severityColors.color + '40'
                      }}
                    >
                      {incident.severity}
                    </Badge>
                    <Badge
                      variant={statusRiskLevel === 'excellent' ? 'default' : 'secondary'}
                      style={{
                        backgroundColor: statusColors.bgColor,
                        color: statusColors.textColor,
                        borderColor: statusColors.color + '40'
                      }}
                    >
                      {incident.status}
                    </Badge>
                  </div>
                </div>
                <p className="text-sm text-muted-foreground mb-2">{incident.description}</p>
                <div className="flex items-center gap-4 text-xs text-muted-foreground">
                  <span>{t('securityPage.impact')}: {incident.impact}</span>
                  <span>{t('securityPage.duration')}: {incident.duration}</span>
                  <span>{new Date(incident.timestamp).toLocaleString()}</span>
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}