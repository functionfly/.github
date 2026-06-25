import { Calendar, MapPin, Briefcase, Star, TrendingUp } from 'lucide-react';
import { IdentitySignature } from './IdentitySignature';
import { ClearanceBadge } from './ClearanceBadge';
import { formatDate } from '@/lib/utils';
import type { Employee } from '@/api/employees';

interface IdentityCardProps {
  employee: Employee;
  identitySignature: string;
  clearanceLevel: number;
  reputationTotal: number;
  trustScore?: number;
  className?: string;
}

const statusColors: Record<string, string> = {
  active: 'bg-green-500/20 text-green-400',
  inactive: 'bg-gray-500/20 text-gray-400',
  on_leave: 'bg-yellow-500/20 text-yellow-400',
  terminated: 'bg-red-500/20 text-red-400',
};

export function IdentityCard({
  employee,
  identitySignature,
  clearanceLevel,
  reputationTotal,
  trustScore,
  className = '',
}: IdentityCardProps) {
  return (
    <div className={`rounded-xl border border-gray-800 bg-gray-900 ${className}`}>
      <div className="flex items-center gap-2 border-b border-gray-800 px-6 py-3">
        <span className="text-xl font-bold text-blue-400">&#x2314;</span>
        <span className="text-sm font-semibold tracking-wider text-gray-300">FUNCTIONFLY IDENTITY</span>
      </div>

      <div className="p-6">
        <div className="flex items-start gap-6">
          <div className="flex h-20 w-20 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-blue-600 to-purple-600">
            <span className="text-2xl font-bold text-white">{employee.ffid.slice(-4)}</span>
          </div>

          <div className="flex-1 min-w-0">
            <div className="flex flex-wrap items-center gap-3">
              <h2 className="text-xl font-bold text-gray-100">{employee.ffid}</h2>
              <span
                className={`rounded-full px-2 py-0.5 text-xs font-medium ${statusColors[employee.status] || 'bg-gray-500/20 text-gray-400'}`}
              >
                {employee.status}
              </span>
              <ClearanceBadge level={clearanceLevel} />
            </div>

            <div className="mt-2">
              <IdentitySignature signature={identitySignature} />
            </div>

            <div className="mt-3 flex flex-wrap gap-4 text-sm text-gray-400">
              <span className="flex items-center gap-1">
                <Briefcase className="h-4 w-4" />
                {employee.employment_type.replace('_', ' ')}
              </span>
              {employee.work_location && (
                <span className="flex items-center gap-1">
                  <MapPin className="h-4 w-4" />
                  {employee.work_location}
                </span>
              )}
              {employee.hire_date && (
                <span className="flex items-center gap-1">
                  <Calendar className="h-4 w-4" />
                  Joined {formatDate(employee.hire_date)}
                </span>
              )}
            </div>
          </div>

          <div className="flex shrink-0 gap-4">
            <div className="text-center">
              <div className="flex items-center gap-1 text-yellow-400">
                <Star className="h-4 w-4" />
                <span className="text-lg font-bold">{reputationTotal}</span>
              </div>
              <span className="text-xs text-gray-500">Reputation</span>
            </div>
            {trustScore !== undefined && (
              <div className="text-center">
                <div className="flex items-center gap-1 text-green-400">
                  <TrendingUp className="h-4 w-4" />
                  <span className="text-lg font-bold">{trustScore}</span>
                </div>
                <span className="text-xs text-gray-500">Trust</span>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
