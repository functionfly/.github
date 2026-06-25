import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { employeesApi } from '@/api/employees';
import { Users, Search, User, MapPin } from 'lucide-react';
import { useState } from 'react';

export function TeamPage() {
  const navigate = useNavigate();
  const [search, setSearch] = useState('');
  const { data, isLoading } = useQuery({
    queryKey: ['employees', { search }],
    queryFn: () => employeesApi.list({ search: search || undefined }),
  });

  const employees = data?.data?.employees || [];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Team Directory</h1>

      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
        <input
          type="text"
          placeholder="Search team members..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full rounded-lg border border-gray-700 bg-gray-800 py-2 pl-10 pr-4 text-sm text-gray-100 placeholder-gray-500 focus:border-blue-500 focus:outline-none"
        />
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : employees.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <Users className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">No team members found</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-3">
          {employees.map((emp) => (
            <div
              key={emp.id}
              onClick={() => navigate(`/profile/${emp.id}`)}
              className="cursor-pointer rounded-xl border border-gray-800 bg-gray-900 p-4 transition-colors hover:border-gray-700"
            >
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-gray-700">
                  <User className="h-5 w-5 text-gray-400" />
                </div>
                <div>
                  <p className="font-medium text-gray-200">{emp.ffid}</p>
                  <p className="text-xs text-gray-500 capitalize">{emp.employment_type.replace('_', ' ')}</p>
                </div>
              </div>
              <div className="mt-3 flex items-center gap-2 text-xs text-gray-500">
                {emp.work_location && (
                  <span className="flex items-center gap-1">
                    <MapPin className="h-3 w-3" />
                    {emp.work_location}
                  </span>
                )}
                <span className="rounded bg-gray-800 px-1.5 py-0.5">{emp.clearance_level}</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
