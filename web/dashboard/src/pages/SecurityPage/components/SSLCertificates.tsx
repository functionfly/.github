import { Badge } from '@/components/ui/badge';
import { Lock, AlertCircle, CheckCircle2, XCircle } from 'lucide-react';
import type { SSLCertificate } from '../types';
import { EXPIRING_CERT_THRESHOLD } from '../constants';
import { RISK_LEVELS, getStatusRiskLevel } from '../utils/riskColors';

interface SSLCertificatesProps {
  certificates: SSLCertificate[];
}

export function SSLCertificates({ certificates }: SSLCertificatesProps) {
  return (
    <div>
      <h4 className="font-semibold mb-4 flex items-center gap-2">
        <Lock className="h-4 w-4" />
        SSL/TLS Certificates
      </h4>
      <div className="space-y-3">
        {certificates.map((cert) => {
          const expiryDate = new Date(cert.expiryDate);
          const daysUntilExpiry = Math.ceil((expiryDate.getTime() - Date.now()) / (1000 * 60 * 60 * 24));
          const isExpiringSoon = daysUntilExpiry < EXPIRING_CERT_THRESHOLD;
          const isExpired = daysUntilExpiry < 0;

          const status = isExpired ? 'expired' : isExpiringSoon ? 'expiring' : 'valid';
          const riskLevel = getStatusRiskLevel(status);
          const riskColors = RISK_LEVELS[riskLevel];

          return (
            <div key={cert.domain} className="flex flex-col sm:flex-row items-start sm:items-center justify-between p-3 md:p-3 gap-2 sm:gap-0 border rounded-lg" style={{ borderColor: riskColors.color + '20' }}>
              <div className="flex items-center gap-3 flex-1 min-w-0">
                {isExpired ? (
                  <XCircle className="h-4 w-4 flex-shrink-0" style={{ color: riskColors.color }} />
                ) : isExpiringSoon ? (
                  <AlertCircle className="h-4 w-4 flex-shrink-0" style={{ color: riskColors.color }} />
                ) : (
                  <CheckCircle2 className="h-4 w-4 flex-shrink-0" style={{ color: riskColors.color }} />
                )}
                <div className="min-w-0 flex-1">
                  <p className="font-medium text-sm truncate">{cert.domain}</p>
                  <p className="text-xs text-muted-foreground truncate">{cert.issuer}</p>
                </div>
              </div>
              <div className="flex flex-col sm:items-end gap-1 sm:gap-0">
                <div
                  className="px-2 py-1 rounded-full text-xs font-medium self-start sm:self-end"
                  style={{
                    backgroundColor: riskColors.bgColor,
                    color: riskColors.textColor,
                    border: `1px solid ${riskColors.color}40`
                  }}
                >
                  {isExpired ? 'Expired' : `${daysUntilExpiry} days`}
                </div>
                {cert.autoRenewal && (
                  <p className="text-xs self-start sm:self-end" style={{ color: RISK_LEVELS.excellent.color }}>Auto-renewal</p>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}