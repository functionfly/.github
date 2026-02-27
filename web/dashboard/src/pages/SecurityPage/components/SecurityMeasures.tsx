import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { CheckCircle, AlertCircle } from 'lucide-react';
import { securityApi } from '@/api/security';
import type { SecurityMeasure } from '../types';
import * as LucideIcons from 'lucide-react';

export function SecurityMeasures() {
  const [measures, setMeasures] = useState<SecurityMeasure[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchMeasures = async () => {
      try {
        setLoading(true);
        const response = await securityApi.getSecurityMeasures();
        setMeasures(response.measures);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load security measures');
      } finally {
        setLoading(false);
      }
    };

    fetchMeasures();
  }, []);

  const getIcon = (iconName: string) => {
    const IconComponent = (LucideIcons as any)[iconName];
    return IconComponent || LucideIcons.Shield;
  };

  if (loading) {
    return (
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="border rounded-lg p-6">
            <div className="flex items-center gap-2 mb-4">
              <div className="h-5 w-5 bg-muted animate-pulse rounded"></div>
              <div className="h-5 bg-muted animate-pulse rounded w-32"></div>
            </div>
            <div className="space-y-2">
              {[1, 2, 3].map((j) => (
                <div key={j} className="flex items-start gap-2">
                  <div className="h-4 w-4 bg-muted animate-pulse rounded mt-0.5"></div>
                  <div className="h-4 bg-muted animate-pulse rounded w-full"></div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center py-8">
        <div className="text-center">
          <AlertCircle className="h-8 w-8 text-red-500 mx-auto mb-2" />
          <p className="text-sm text-muted-foreground">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
      {measures.map((category) => {
        const IconComponent = getIcon(category.icon);
        return (
          <div key={category.category} className="border rounded-lg p-6">
            <div className="flex items-center gap-2 mb-4">
              <IconComponent className="h-5 w-5 text-emerald-600" />
              <h3 className="font-semibold">{category.category}</h3>
            </div>
            <ul className="space-y-2">
              {category.measures.map((measure, index) => (
                <li key={index} className="flex items-start gap-2 text-sm">
                  <CheckCircle className="h-4 w-4 text-green-500 mt-0.5 flex-shrink-0" />
                  <span className="text-muted-foreground">{measure}</span>
                </li>
              ))}
            </ul>
          </div>
        );
      })}
    </div>
  );
}