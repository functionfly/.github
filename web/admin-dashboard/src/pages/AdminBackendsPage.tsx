/**
 * Admin Backends Page
 * Manage execution backends and worker pools
 */

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { Plus, Search, Activity } from 'lucide-react';
import { LoadingScreen } from '@/components/common/LoadingScreen';

interface Backend {
  id: string;
  name: string;
  type: string;
  status: 'active' | 'inactive' | 'unhealthy';
  workers: number;
  utilization: number;
  created_at: string;
}

export function AdminBackendsPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');

  const { data: backendsResponse, isLoading, isError } = useQuery({
    queryKey: ['admin-backends'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<Backend[]>('/backends');
      } catch {
        return { data: [], success: false };
      }
    },
    staleTime: 1000 * 60,
  });

  const backends = backendsResponse?.data || [];

  if (isLoading) {
    return <LoadingScreen />;
  }

  if (isError) {
    return (
      <div className="p-8 bg-red-50 border border-red-200 rounded-lg">
        <h3 className="font-semibold text-red-900">Error loading backends</h3>
      </div>
    );
  }

  const filteredBackends = backends.filter((backend) => {
    const matchesSearch = backend.name.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = statusFilter === 'all' || backend.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Execution Backends</h1>
          <p className="mt-2 text-gray-600">Manage worker pools and execution infrastructure</p>
        </div>
        <button className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
          <Plus className="w-5 h-5" />
          Add Backend
        </button>
      </div>

      <div className="flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-3 w-5 h-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search backends..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg"
          />
        </div>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="px-4 py-2 border border-gray-300 rounded-lg"
        >
          <option value="all">All Status</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
          <option value="unhealthy">Unhealthy</option>
        </select>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {filteredBackends.map((backend) => (
          <div key={backend.id} className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold text-gray-900">{backend.name}</h3>
              <Activity className={`w-5 h-5 ${
                backend.status === 'active' ? 'text-green-600' :
                backend.status === 'unhealthy' ? 'text-red-600' :
                'text-gray-400'
              }`} />
            </div>
            <div className="space-y-2">
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Type</span>
                <span className="font-medium">{backend.type}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-gray-600">Workers</span>
                <span className="font-medium">{backend.workers}</span>
              </div>
              <div className="flex justify-between text-sm mb-3">
                <span className="text-gray-600">Utilization</span>
                <span className="font-medium">{backend.utilization}%</span>
              </div>
              <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                  className="bg-blue-600 h-2 rounded-full"
                  style={{ width: `${backend.utilization}%` }}
                />
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
