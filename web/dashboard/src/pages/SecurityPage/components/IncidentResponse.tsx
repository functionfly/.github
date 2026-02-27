import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { AlertTriangle, Eye, Zap, Users, Database, Monitor, AlertCircle } from 'lucide-react';
import { securityApi } from '@/api/security';
import type { IncidentResponse as IncidentResponseType } from '../types';

export function IncidentResponse() {
  const [incidentResponse, setIncidentResponse] = useState<IncidentResponseType | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchIncidentResponse = async () => {
      try {
        setLoading(true);
        const response = await securityApi.getIncidentResponse();
        setIncidentResponse(response);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load incident response data');
      } finally {
        setLoading(false);
      }
    };

    fetchIncidentResponse();
  }, []);

  const responseSteps = incidentResponse ? [
    {
      icon: Eye,
      title: 'Detection',
      description: incidentResponse.detection,
      color: 'blue'
    },
    {
      icon: Zap,
      title: 'Response',
      description: incidentResponse.response,
      color: 'orange'
    },
    {
      icon: Users,
      title: 'Communication',
      description: incidentResponse.communication,
      color: 'green'
    },
    {
      icon: Database,
      title: 'Recovery',
      description: incidentResponse.recovery,
      color: 'purple'
    },
    {
      icon: Monitor,
      title: 'Learning',
      description: incidentResponse.learning,
      color: 'indigo'
    }
  ] : [];

  if (loading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {[1, 2, 3, 4, 5].map((i) => (
          <div key={i} className="text-center">
            <div className="w-12 h-12 bg-muted animate-pulse rounded-full mx-auto mb-3"></div>
            <div className="h-4 bg-muted animate-pulse rounded w-20 mx-auto mb-2"></div>
            <div className="h-3 bg-muted animate-pulse rounded w-full"></div>
          </div>
        ))}
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
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      {responseSteps.map((step, _index) => (
        <div key={step.title} className="text-center">
          <div className={`w-12 h-12 bg-${step.color}-500/10 rounded-full flex items-center justify-center mx-auto mb-3`}>
            <step.icon className={`h-6 w-6 text-${step.color}-500`} />
          </div>
          <h4 className="font-semibold mb-2">{step.title}</h4>
          <p className="text-sm text-muted-foreground">{step.description}</p>
        </div>
      ))}
    </div>
  );
}