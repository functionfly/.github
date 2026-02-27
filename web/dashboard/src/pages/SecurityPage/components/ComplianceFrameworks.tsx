import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { FileText, AlertCircle, Award, ShieldCheck } from 'lucide-react';
import { securityApi } from '@/api/security';
import type { ComplianceFramework } from '../types';
import { RISK_LEVELS, getStatusRiskLevel } from '../utils/riskColors';
import { SwipeableBadge } from './SwipeableBadge';

export function ComplianceFrameworks() {
  const [frameworks, setFrameworks] = useState<ComplianceFramework[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchFrameworks = async () => {
      try {
        setLoading(true);
        const response = await securityApi.getComplianceFrameworks();
        setFrameworks(response.frameworks);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load compliance frameworks');
      } finally {
        setLoading(false);
      }
    };

    fetchFrameworks();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8">
        <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center py-8 text-center">
        <div>
          <AlertCircle className="h-8 w-8 text-red-500 mx-auto mb-2" />
          <p className="text-sm text-muted-foreground">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <SwipeableBadge frameworks={frameworks} />
  );
}