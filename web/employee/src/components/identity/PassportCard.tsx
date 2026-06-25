import { Download } from 'lucide-react';
import { IdentityCard } from './IdentityCard';
import { IdentitySignature } from './IdentitySignature';
import { ClearanceBadge } from './ClearanceBadge';
import type { Employee } from '@/api/employees';
import type { Achievement } from '@/api/identity';

interface PassportCardProps {
  employee: Employee;
  identitySignature: string;
  clearanceLevel: number;
  reputationTotal: number;
  trustScore?: number;
  achievements?: Achievement[];
  onExport?: () => void;
  className?: string;
}

export function PassportCard({
  employee,
  identitySignature,
  clearanceLevel,
  reputationTotal,
  trustScore,
  achievements,
  onExport,
  className = '',
}: PassportCardProps) {
  return (
    <div className={`space-y-4 ${className}`}>
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-gray-100">FunctionFly Passport</h2>
        {onExport && (
          <button
            onClick={onExport}
            className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
          >
            <Download className="h-4 w-4" />
            Export PDF
          </button>
        )}
      </div>

      <div className="rounded-xl border-2 border-blue-500/20 bg-gradient-to-br from-gray-900 via-gray-900 to-blue-950/30">
        <div className="flex items-center gap-2 border-b border-gray-800 px-6 py-3">
          <span className="text-xl font-bold text-blue-400">&#x2314;</span>
          <span className="text-sm font-semibold tracking-wider text-gray-300">FUNCTIONFLY PASSPORT</span>
          <span className="ml-auto text-xs text-gray-500">Digital Identity Document</span>
        </div>

        <div className="p-6">
          <div className="flex items-start gap-6">
            <div className="flex h-24 w-24 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-blue-600 to-purple-600 ring-4 ring-blue-500/20">
              <span className="text-3xl font-bold text-white">{employee.ffid.slice(-4)}</span>
            </div>

            <div className="flex-1">
              <h3 className="text-xl font-bold text-gray-100">{employee.ffid}</h3>
              <div className="mt-2 flex flex-wrap items-center gap-3">
                <IdentitySignature signature={identitySignature} />
                <ClearanceBadge level={clearanceLevel} />
              </div>
              <div className="mt-3 grid grid-cols-2 gap-3 text-sm sm:grid-cols-4">
                <div>
                  <span className="text-xs text-gray-500">Status</span>
                  <p className="text-gray-200 capitalize">{employee.status}</p>
                </div>
                <div>
                  <span className="text-xs text-gray-500">Type</span>
                  <p className="text-gray-200 capitalize">{employee.employment_type.replace('_', ' ')}</p>
                </div>
                <div>
                  <span className="text-xs text-gray-500">Reputation</span>
                  <p className="text-yellow-400">{reputationTotal}</p>
                </div>
                {trustScore !== undefined && (
                  <div>
                    <span className="text-xs text-gray-500">Trust Score</span>
                    <p className="text-green-400">{trustScore}</p>
                  </div>
                )}
              </div>
            </div>
          </div>

          {achievements && achievements.length > 0 && (
            <div className="mt-6 border-t border-gray-800 pt-4">
              <span className="text-xs text-gray-500">Recent Achievements</span>
              <div className="mt-2 flex flex-wrap gap-2">
                {achievements.filter((a) => a.earned).slice(0, 6).map((a) => (
                  <span key={a.id} className="rounded-full bg-yellow-500/10 px-3 py-1 text-xs text-yellow-400">
                    {a.icon || '🏆'} {a.name}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
