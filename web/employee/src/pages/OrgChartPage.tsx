import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { orgchartApi, type OrgChartEmployee } from '@/api/orgchart';
import { Network, ChevronRight, ChevronDown, Search, User } from 'lucide-react';

interface TreeNodeProps {
  employee: OrgChartEmployee;
  allEmployees: OrgChartEmployee[];
  level: number;
}

function TreeNode({ employee, allEmployees, level }: TreeNodeProps) {
  const [expanded, setExpanded] = useState(level < 2);
  const reports = allEmployees.filter((e) => e.manager_id === employee.id);
  const hasReports = reports.length > 0;

  return (
    <div className={level > 0 ? 'ml-6 border-l border-gray-700 pl-4' : ''}>
      <div
        className="flex cursor-pointer items-center gap-3 rounded-lg p-2 hover:bg-gray-800/50"
        onClick={() => hasReports && setExpanded(!expanded)}
      >
        {hasReports ? (
          expanded ? (
            <ChevronDown className="h-4 w-4 text-gray-500" />
          ) : (
            <ChevronRight className="h-4 w-4 text-gray-500" />
          )
        ) : (
          <div className="w-4" />
        )}
        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gray-700">
          {employee.avatar_url ? (
            <img src={employee.avatar_url} alt="" className="h-full w-full rounded-full object-cover" />
          ) : (
            <User className="h-4 w-4 text-gray-400" />
          )}
        </div>
        <div>
          <p className="text-sm font-medium text-gray-200">{employee.name}</p>
          <p className="text-xs text-gray-500">
            {employee.job_title || 'No title'}
            {employee.department_name && ` · ${employee.department_name}`}
          </p>
        </div>
        <span className="ml-auto rounded bg-gray-800 px-1.5 py-0.5 text-xs text-gray-500">
          {employee.ffid}
        </span>
      </div>
      {expanded && hasReports && (
        <div className="mt-1">
          {reports.map((report) => (
            <TreeNode key={report.id} employee={report} allEmployees={allEmployees} level={level + 1} />
          ))}
        </div>
      )}
    </div>
  );
}

export function OrgChartPage() {
  const [search, setSearch] = useState('');
  const { data, isLoading } = useQuery({
    queryKey: ['orgchart'],
    queryFn: () => orgchartApi.getOrgChart(),
  });

  const employees = data?.data?.employees || [];
  const roots = employees.filter((e) => !e.manager_id || !employees.find((m) => m.id === e.manager_id));

  const filtered = search
    ? employees.filter(
        (e) =>
          e.name.toLowerCase().includes(search.toLowerCase()) ||
          e.ffid.toLowerCase().includes(search.toLowerCase())
      )
    : [];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Organization Chart</h1>

      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
        <input
          type="text"
          placeholder="Search by name or FFID..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full rounded-lg border border-gray-700 bg-gray-800 py-2 pl-10 pr-4 text-sm text-gray-100 placeholder-gray-500 focus:border-blue-500 focus:outline-none"
        />
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : search ? (
        <div className="space-y-2">
          {filtered.map((emp) => (
            <div key={emp.id} className="flex items-center gap-3 rounded-lg border border-gray-800 bg-gray-900 p-3">
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gray-700">
                {emp.avatar_url ? (
                  <img src={emp.avatar_url} alt="" className="h-full w-full rounded-full object-cover" />
                ) : (
                  <User className="h-4 w-4 text-gray-400" />
                )}
              </div>
              <div>
                <p className="text-sm font-medium text-gray-200">{emp.name}</p>
                <p className="text-xs text-gray-500">
                  {emp.job_title || 'No title'}
                  {emp.department_name && ` · ${emp.department_name}`}
                </p>
              </div>
              <span className="ml-auto rounded bg-gray-800 px-1.5 py-0.5 text-xs text-gray-500">
                {emp.ffid}
              </span>
            </div>
          ))}
          {filtered.length === 0 && (
            <p className="py-8 text-center text-gray-400">No employees found</p>
          )}
        </div>
      ) : (
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
          {roots.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12">
              <Network className="mb-4 h-12 w-12 text-gray-600" />
              <p className="text-gray-400">No organization data available</p>
            </div>
          ) : (
            roots.map((root) => (
              <TreeNode key={root.id} employee={root} allEmployees={employees} level={0} />
            ))
          )}
        </div>
      )}
    </div>
  );
}
