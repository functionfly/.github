/**
 * Admin System Page
 * System configuration, health checks, and monitoring
 */

import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { Activity, HardDrive } from 'lucide-react';

interface SystemMetrics {
  status: 'healthy' | 'degraded' | 'down';
  uptime: number;
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
  apiResponsiveness: number;
  databaseHealth: 'connected' | 'disconnected';
}

export function AdminSystemPage() {
  const { data: metricsResponse } = useQuery({
    queryKey: ['admin-system-metrics'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<SystemMetrics>('/system/metrics');
      } catch {
        return { data: null, success: false };
      }
    },
    staleTime: 1000 * 30,
  });

  const metrics = metricsResponse?.data;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">System Health & Configuration</h1>
        <p className="mt-2 text-gray-600">Monitor system resources, health checks, and configurations</p>
      </div>

      {/* Health Status */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 text-sm">System Status</p>
              <p className={`text-2xl font-bold ${
                metrics?.status === 'healthy' ? 'text-green-600' :
                metrics?.status === 'degraded' ? 'text-yellow-600' :
                'text-red-600'
              }`}>
                {metrics?.status || 'checking...'}
              </p>
            </div>
            <Activity className={`w-8 h-8 ${
              metrics?.status === 'healthy' ? 'text-green-600' :
              metrics?.status === 'degraded' ? 'text-yellow-600' :
              'text-red-600'
            }`} />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 text-sm">Database</p>
              <p className={`text-2xl font-bold ${
                metrics?.databaseHealth === 'connected' ? 'text-green-600' : 'text-red-600'
              }`}>
                {metrics?.databaseHealth || 'checking...'}
              </p>
            </div>
            <HardDrive className={`w-8 h-8 ${
              metrics?.databaseHealth === 'connected' ? 'text-green-600' : 'text-red-600'
            }`} />
          </div>
        </div>
      </div>

      {/* Resource Usage */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <p className="text-gray-600 text-sm mb-3">CPU Usage</p>
          <div className="relative w-24 h-24 mx-auto mb-4">
            <svg className="w-full h-full transform -rotate-90" viewBox="0 0 100 100">
              <circle cx="50" cy="50" r="45" fill="none" stroke="#e5e7eb" strokeWidth="8" />
              <circle
                cx="50"
                cy="50"
                r="45"
                fill="none"
                stroke="#3b82f6"
                strokeWidth="8"
                strokeDasharray={`${(metrics?.cpuUsage || 0) * 2.827} 282.7`}
              />
            </svg>
            <span className="absolute inset-0 flex items-center justify-center font-bold">
              {metrics?.cpuUsage}%
            </span>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <p className="text-gray-600 text-sm mb-3">Memory Usage</p>
          <div className="relative w-24 h-24 mx-auto mb-4">
            <svg className="w-full h-full transform -rotate-90" viewBox="0 0 100 100">
              <circle cx="50" cy="50" r="45" fill="none" stroke="#e5e7eb" strokeWidth="8" />
              <circle
                cx="50"
                cy="50"
                r="45"
                fill="none"
                stroke="#8b5cf6"
                strokeWidth="8"
                strokeDasharray={`${(metrics?.memoryUsage || 0) * 2.827} 282.7`}
              />
            </svg>
            <span className="absolute inset-0 flex items-center justify-center font-bold">
              {metrics?.memoryUsage}%
            </span>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <p className="text-gray-600 text-sm mb-3">Disk Usage</p>
          <div className="relative w-24 h-24 mx-auto mb-4">
            <svg className="w-full h-full transform -rotate-90" viewBox="0 0 100 100">
              <circle cx="50" cy="50" r="45" fill="none" stroke="#e5e7eb" strokeWidth="8" />
              <circle
                cx="50"
                cy="50"
                r="45"
                fill="none"
                stroke="#ec4899"
                strokeWidth="8"
                strokeDasharray={`${(metrics?.diskUsage || 0) * 2.827} 282.7`}
              />
            </svg>
            <span className="absolute inset-0 flex items-center justify-center font-bold">
              {metrics?.diskUsage}%
            </span>
          </div>
        </div>
      </div>

      {/* Configuration */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">System Configuration</h2>
        <div className="space-y-4">
          <div className="flex justify-between items-center pb-4 border-b border-gray-200">
            <span className="text-gray-600">API Version</span>
            <span className="font-medium text-gray-900">v1.0.0</span>
          </div>
          <div className="flex justify-between items-center pb-4 border-b border-gray-200">
            <span className="text-gray-600">Environment</span>
            <span className="font-medium text-gray-900">Production</span>
          </div>
          <div className="flex justify-between items-center pb-4 border-b border-gray-200">
            <span className="text-gray-600">Uptime</span>
            <span className="font-medium text-gray-900">{metrics?.uptime}h</span>
          </div>
          <div className="flex justify-between items-center">
            <span className="text-gray-600">API Responsiveness</span>
            <span className="font-medium text-gray-900">{metrics?.apiResponsiveness}ms</span>
          </div>
        </div>
      </div>
    </div>
  );
}
