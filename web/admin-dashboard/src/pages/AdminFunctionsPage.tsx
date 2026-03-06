/**
 * Admin Functions Page
 * Manage deployed functions and their configurations
 */

import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { Plus, Search, BarChart3 } from 'lucide-react';
import { LoadingScreen } from '@/components/common/LoadingScreen';

interface FunctionData {
  id: string;
  name: string;
  tenant: string;
  runtime: string;
  status: 'active' | 'inactive' | 'error';
  invocations: number;
  errors: number;
  avgLatency: number;
  created_at: string;
}

export function AdminFunctionsPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [runtimeFilter, setRuntimeFilter] = useState<string>('all');

  const { data: functionsResponse, isLoading, isError } = useQuery({
    queryKey: ['admin-functions'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<FunctionData[]>('/functions');
      } catch {
        return { data: [], success: false };
      }
    },
    staleTime: 1000 * 60,
  });

  const functions = functionsResponse?.data || [];

  if (isLoading) {
    return <LoadingScreen />;
  }

  if (isError) {
    return (
      <div className="p-8 bg-red-50 border border-red-200 rounded-lg">
        <h3 className="font-semibold text-red-900">Error loading functions</h3>
      </div>
    );
  }

  const filteredFunctions = functions.filter((func) => {
    const matchesSearch = func.name.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesRuntime = runtimeFilter === 'all' || func.runtime === runtimeFilter;
    return matchesSearch && matchesRuntime;
  });

  const runtimes = [...new Set(functions.map((f) => f.runtime))];

  const totalInvocations = functions.reduce((sum, f) => sum + f.invocations, 0);
  const totalErrors = functions.reduce((sum, f) => sum + f.errors, 0);
  const errorRate = totalInvocations > 0 ? ((totalErrors / totalInvocations) * 100).toFixed(2) : '0';

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Functions</h1>
          <p className="mt-2 text-gray-600">Manage deployed functions and view metrics</p>
        </div>
        <button className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
          <Plus className="w-5 h-5" />
          Deploy Function
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
          <p className="text-gray-600 text-sm">Total Functions</p>
          <p className="text-2xl font-bold text-gray-900">{functions.length}</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
          <p className="text-gray-600 text-sm">Total Invocations</p>
          <p className="text-2xl font-bold text-gray-900">{totalInvocations.toLocaleString()}</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
          <p className="text-gray-600 text-sm">Total Errors</p>
          <p className="text-2xl font-bold text-red-600">{totalErrors.toLocaleString()}</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4">
          <p className="text-gray-600 text-sm">Error Rate</p>
          <p className="text-2xl font-bold text-yellow-600">{errorRate}%</p>
        </div>
      </div>

      <div className="flex gap-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-3 w-5 h-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search functions..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg"
          />
        </div>
        <select
          value={runtimeFilter}
          onChange={(e) => setRuntimeFilter(e.target.value)}
          className="px-4 py-2 border border-gray-300 rounded-lg"
        >
          <option value="all">All Runtimes</option>
          {runtimes.map((runtime) => (
            <option key={runtime} value={runtime}>
              {runtime}
            </option>
          ))}
        </select>
      </div>

      <div className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-200 bg-gray-50">
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Name</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Tenant</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Runtime</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Invocations</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Errors</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700">Avg Latency</th>
            </tr>
          </thead>
          <tbody>
            {filteredFunctions.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-6 py-8 text-center text-gray-500">
                  No functions found
                </td>
              </tr>
            ) : (
              filteredFunctions.map((func) => (
                <tr key={func.id} className="border-b border-gray-100 hover:bg-gray-50">
                  <td className="px-6 py-4 text-sm font-medium text-gray-900">
                    <Link to={`/functions/${func.id}`} className="text-blue-700 hover:text-blue-800 hover:underline">
                      {func.name}
                    </Link>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-600">{func.tenant}</td>
                  <td className="px-6 py-4 text-sm text-gray-600">{func.runtime}</td>
                  <td className="px-6 py-4 text-sm text-gray-600">
                    <div className="flex items-center gap-2">
                      <BarChart3 className="w-4 h-4 text-blue-500" />
                      {func.invocations.toLocaleString()}
                    </div>
                  </td>
                  <td className="px-6 py-4 text-sm">
                    <span className={func.errors > 0 ? 'text-red-600 font-medium' : 'text-gray-600'}>
                      {func.errors.toLocaleString()}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-600">{func.avgLatency}ms</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
