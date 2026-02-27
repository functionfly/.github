import { useState, useEffect } from 'react';
import { securityApi } from '@/api/security';
import type { ServiceStatus, SSLCertificate, SecurityIncident } from './types';
import { SECURITY_UPDATE_INTERVAL, MIN_SECURITY_SCORE, MAX_SECURITY_SCORE } from './constants';

interface SecurityState {
  expandedSection: string | null;
  securityScore: number;
  lastUpdated: Date;
  serviceStatus: ServiceStatus[];
  recentIncidents: SecurityIncident[];
  sslCertificates: SSLCertificate[];
  loading: boolean;
  error: string | null;
}

export function useSecurityState() {
  const [state, setState] = useState<SecurityState>({
    expandedSection: null,
    securityScore: 98.5,
    lastUpdated: new Date(),
    serviceStatus: [],
    recentIncidents: [],
    sslCertificates: [],
    loading: true,
    error: null
  });

  // Fetch initial security data
  useEffect(() => {
    const fetchSecurityData = async () => {
      try {
        setState(prev => ({ ...prev, loading: true, error: null }));

        const [metricsResponse] = await Promise.all([
          securityApi.getSecurityMetrics()
        ]);

        setState(prev => ({
          ...prev,
          securityScore: metricsResponse.overallScore,
          lastUpdated: new Date(metricsResponse.lastUpdated),
          serviceStatus: metricsResponse.services,
          recentIncidents: metricsResponse.recentIncidents,
          sslCertificates: metricsResponse.certificates,
          loading: false
        }));
      } catch (error) {
        console.error('Failed to fetch security data:', error);
        setState(prev => ({
          ...prev,
          loading: false,
          error: error instanceof Error ? error.message : 'Failed to load security data'
        }));
      }
    };

    fetchSecurityData();
  }, []);

  // Simulate real-time updates
  useEffect(() => {
    if (state.loading || state.error) return;

    const updateInterval = setInterval(async () => {
      try {
        // Fetch updated metrics
        const metricsResponse = await securityApi.getSecurityMetrics();

        setState(prev => ({
          ...prev,
          securityScore: metricsResponse.overallScore,
          lastUpdated: new Date(metricsResponse.lastUpdated),
          serviceStatus: metricsResponse.services,
          recentIncidents: metricsResponse.recentIncidents,
          sslCertificates: metricsResponse.certificates
        }));
      } catch (error) {
        console.error('Failed to update security data:', error);
        // Continue with existing data rather than showing error for background updates
      }
    }, SECURITY_UPDATE_INTERVAL);

    return () => clearInterval(updateInterval);
  }, [state.loading, state.error]);

  const toggleSection = (sectionId: string) => {
    setState(prev => ({
      ...prev,
      expandedSection: prev.expandedSection === sectionId ? null : sectionId
    }));
  };

  const retry = () => {
    setState(prev => ({ ...prev, loading: true, error: null }));
    // Trigger a re-fetch by calling the effect again
    window.location.reload();
  };

  return {
    ...state,
    toggleSection,
    retry
  };
}