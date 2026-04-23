/**
 * Admin Functions Page
 * Manage deployed functions and view registry metrics
 */

import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { Plus, Search, Star, Eye, DollarSign } from 'lucide-react';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { DeployFunctionModal } from '@/components/common/DeployFunctionModal';

interface RegistryFunctionData {
  id: string;
  author: string;
  name: string;
  title: string;
  description: string;
  category: string;
  visibility: string;
  price_per_call: number;
  popularity_score: number;
  reliability_score: number;
  deterministic_score: number;
  latest_version: string;
  total_ratings: number;
  overall_score: number;
  is_flagged: boolean;
  flag_reason: string | null;
  created_at: string;
  updated_at: string;
}

interface FunctionsAPIResponse {
  functions: RegistryFunctionData[];
  total: number;
}

export function AdminFunctionsPage() {
  const [searchTerm, setSearchTerm] = useState('');
  const [categoryFilter, setCategoryFilter] = useState<string>('all');
  const [visibilityFilter, setVisibilityFilter] = useState<string>('all');
  const [deployModalOpen, setDeployModalOpen] = useState(false);

  const { data, isLoading, isError } = useQuery({
    queryKey: ['admin-functions'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<FunctionsAPIResponse>('/registry/functions');
      } catch {
        return null;
      }
    },
    staleTime: 1000 * 60,
  });

  const functions = data?.functions || [];

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
    const matchesSearch =
      func.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      func.author.toLowerCase().includes(searchTerm.toLowerCase()) ||
      func.title.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesCategory = categoryFilter === 'all' || func.category === categoryFilter;
    const matchesVisibility = visibilityFilter === 'all' || func.visibility === visibilityFilter;
    return matchesSearch && matchesCategory && matchesVisibility;
  });

  const categories = [...new Set(functions.map((f) => f.category).filter(Boolean))];

  const avgRating = functions.length > 0
    ? (functions.reduce((sum, f) => sum + f.overall_score, 0) / functions.length).toFixed(1)
    : '0';
  const totalPrice = functions.reduce((sum, f) => sum + f.price_per_call, 0);
  const flaggedCount = functions.filter((f) => f.is_flagged).length;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Functions</h1>
          <p className="mt-2 text-gray-600 dark:text-gray-400">Manage deployed functions and view metrics</p>
        </div>
        <button
          onClick={() => setDeployModalOpen(true)}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
        >
          <Plus className="w-5 h-5" />
          Deploy Function
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <p className="text-gray-600 dark:text-gray-400 text-sm">Total Functions</p>
          <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">{data?.total ?? functions.length}</p>
        </div>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <p className="text-gray-600 dark:text-gray-400 text-sm">Avg Rating</p>
          <p className="text-2xl font-bold text-gray-900 dark:text-gray-100">
            <Star className="inline w-5 h-5 text-yellow-500 mr-1" />
            {avgRating}
          </p>
        </div>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <p className="text-gray-600 dark:text-gray-400 text-sm">Total Revenue</p>
          <p className="text-2xl font-bold text-green-600 dark:text-green-400">
            <DollarSign className="inline w-5 h-5 mr-1" />
            {totalPrice.toFixed(2)}
          </p>
        </div>
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
          <p className="text-gray-600 dark:text-gray-400 text-sm">Flagged</p>
          <p className="text-2xl font-bold text-red-600 dark:text-red-400">{flaggedCount}</p>
        </div>
      </div>

      <div className="flex gap-4 flex-wrap">
        <div className="flex-1 min-w-[200px] relative">
          <Search className="absolute left-3 top-3 w-5 h-5 text-gray-400" />
          <input
            type="text"
            placeholder="Search functions..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
          />
        </div>
        <select
          value={categoryFilter}
          onChange={(e) => setCategoryFilter(e.target.value)}
          className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        >
          <option value="all">All Categories</option>
          {categories.map((cat) => (
            <option key={cat} value={cat}>
              {cat}
            </option>
          ))}
        </select>
        <select
          value={visibilityFilter}
          onChange={(e) => setVisibilityFilter(e.target.value)}
          className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
        >
          <option value="all">All Visibility</option>
          <option value="public">Public</option>
          <option value="private">Private</option>
          <option value="unlisted">Unlisted</option>
        </select>
      </div>

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Name</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Author</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Category</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Visibility</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Price/Call</th>
              <th className="px-6 py-3 text-left text-sm font-semibold text-gray-700 dark:text-gray-300">Rating</th>
            </tr>
          </thead>
          <tbody>
            {filteredFunctions.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-6 py-8 text-center text-gray-500 dark:text-gray-400">
                  No functions found
                </td>
              </tr>
            ) : (
              filteredFunctions.map((func) => (
                <tr key={func.id} className="border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800">
                  <td className="px-6 py-4 text-sm">
                    <div>
                      <Link
                        to={`/functions/${func.id}`}
                        className="font-medium text-blue-700 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 hover:underline"
                      >
                        {func.name}
                      </Link>
                      {func.title && (
                        <p className="text-xs text-gray-500 dark:text-gray-400 truncate max-w-[200px]">{func.title}</p>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-600 dark:text-gray-400">{func.author}</td>
                  <td className="px-6 py-4 text-sm text-gray-600 dark:text-gray-400">
                    {func.category || <span className="text-gray-400">—</span>}
                  </td>
                  <td className="px-6 py-4 text-sm">
                    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${
                      func.visibility === 'public'
                        ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
                        : func.visibility === 'private'
                        ? 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
                        : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
                    }`}>
                      <Eye className="w-3 h-3" />
                      {func.visibility}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-600 dark:text-gray-400">
                    {func.price_per_call > 0 ? `$${func.price_per_call.toFixed(4)}` : <span className="text-green-600">Free</span>}
                  </td>
                  <td className="px-6 py-4 text-sm">
                    <div className="flex items-center gap-1">
                      <Star className="w-4 h-4 text-yellow-500" />
                      <span className={func.overall_score > 0 ? 'text-gray-900 dark:text-gray-100' : 'text-gray-400'}>
                        {func.overall_score > 0 ? func.overall_score.toFixed(1) : 'N/A'}
                      </span>
                      {func.total_ratings > 0 && (
                        <span className="text-xs text-gray-400">({func.total_ratings})</span>
                      )}
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <DeployFunctionModal
        open={deployModalOpen}
        onOpenChange={setDeployModalOpen}
      />
    </div>
  );
}