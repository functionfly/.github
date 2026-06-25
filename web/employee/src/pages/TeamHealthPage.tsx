import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { teamHealthApi, type TeamHealthMetric } from '@/api/team_health';
import { HeartPulse, AlertTriangle, Users, Clock, Flame } from 'lucide-react';
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  LineChart, Line, RadarChart, Radar, PolarGrid, PolarAngleAxis, PolarRadiusAxis,
  PieChart, Pie, Cell, Legend,
} from 'recharts';

const COLORS = ['#3b82f6', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899'];

export function TeamHealthPage() {
  const [deptFilter, setDeptFilter] = useState<number | undefined>();

  const { data: metricsData, isLoading } = useQuery({
    queryKey: ['team-health', deptFilter],
    queryFn: () => teamHealthApi.get(deptFilter ? { department_id: deptFilter } : undefined),
  });

  const { data: burnoutData } = useQuery({
    queryKey: ['team-health', 'burnout-risk'],
    queryFn: () => teamHealthApi.getBurnoutRisk(),
  });

  const metrics = metricsData?.data?.metrics || [];
  const atRisk = burnoutData?.data?.at_risk || [];

  const avgMetric = (fn: (m: TeamHealthMetric) => number) =>
    metrics.length > 0 ? metrics.reduce((s, m) => s + fn(m), 0) / metrics.length : 0;

  const workloadData = metrics.slice(0, 10).map((m) => ({
    name: m.metric_date.slice(5),
    workload: m.workload_score,
    velocity: m.velocity_score,
  }));

  const velocityTrend = metrics.slice(0, 12).map((m) => ({
    date: m.metric_date.slice(5),
    velocity: m.velocity_score,
    collaboration: m.collaboration_score,
  })).reverse();

  const radarData = [
    { subject: 'Workload', score: Math.round(avgMetric((m) => m.workload_score)), fullMark: 100 },
    { subject: 'Velocity', score: Math.round(avgMetric((m) => m.velocity_score)), fullMark: 100 },
    { subject: 'Collaboration', score: Math.round(avgMetric((m) => m.collaboration_score)), fullMark: 100 },
    { subject: 'Knowledge', score: Math.round(avgMetric((m) => m.knowledge_sharing_score)), fullMark: 100 },
    { subject: 'PTO Util', score: Math.round(avgMetric((m) => m.pto_utilization_pct)), fullMark: 100 },
  ];

  const ptoAvg = avgMetric((m) => m.pto_utilization_pct);
  const ptoPie = [
    { name: 'Used', value: Math.round(ptoAvg) },
    { name: 'Available', value: Math.round(100 - ptoAvg) },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <HeartPulse className="h-6 w-6 text-red-400" />
          <h1 className="text-2xl font-bold">Team Health</h1>
        </div>
        <select
          value={deptFilter ?? ''}
          onChange={(e) => setDeptFilter(e.target.value ? Number(e.target.value) : undefined)}
          className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
        >
          <option value="">All Departments</option>
          {[1, 2, 3, 4, 5].map((id) => (
            <option key={id} value={id}>Department {id}</option>
          ))}
        </select>
      </div>

      {atRisk.length > 0 && (
        <div className="space-y-2">
          <h3 className="flex items-center gap-2 text-sm font-medium text-gray-300">
            <AlertTriangle className="h-4 w-4 text-yellow-400" />
            Burnout Risk Alerts
          </h3>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {atRisk.map((emp) => {
              const riskPct = Math.round(emp.risk_score * 100);
              const color = riskPct > 70 ? 'border-red-500/50 bg-red-500/10' :
                            riskPct > 40 ? 'border-yellow-500/50 bg-yellow-500/10' :
                            'border-green-500/50 bg-green-500/10';
              const textColor = riskPct > 70 ? 'text-red-400' : riskPct > 40 ? 'text-yellow-400' : 'text-green-400';
              return (
                <div key={emp.employee_id} className={`rounded-xl border p-4 ${color}`}>
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium text-gray-100">{emp.ffid}</span>
                    <span className={`text-sm font-bold ${textColor}`}>{riskPct}%</span>
                  </div>
                  <div className="mt-2 h-1.5 rounded-full bg-gray-800">
                    <div
                      className={`h-full rounded-full ${riskPct > 70 ? 'bg-red-500' : riskPct > 40 ? 'bg-yellow-500' : 'bg-green-500'}`}
                      style={{ width: `${riskPct}%` }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : metrics.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <HeartPulse className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">No team health metrics available</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <div className="mb-2 flex items-center gap-2 text-gray-400">
                <Flame className="h-4 w-4" />
                <span className="text-sm">Avg Burnout Risk</span>
              </div>
              <p className="text-2xl font-bold text-gray-100">{Math.round(avgMetric((m) => m.burnout_risk) * 100)}%</p>
            </div>
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <div className="mb-2 flex items-center gap-2 text-gray-400">
                <Clock className="h-4 w-4" />
                <span className="text-sm">Avg Overtime</span>
              </div>
              <p className="text-2xl font-bold text-gray-100">{avgMetric((m) => m.avg_overtime_hours).toFixed(1)}h</p>
            </div>
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <div className="mb-2 flex items-center gap-2 text-gray-400">
                <Users className="h-4 w-4" />
                <span className="text-sm">Total Headcount</span>
              </div>
              <p className="text-2xl font-bold text-gray-100">{metrics.reduce((s, m) => s + m.headcount, 0)}</p>
            </div>
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <div className="mb-2 flex items-center gap-2 text-gray-400">
                <HeartPulse className="h-4 w-4" />
                <span className="text-sm">Collaboration</span>
              </div>
              <p className="text-2xl font-bold text-gray-100">{Math.round(avgMetric((m) => m.collaboration_score))}</p>
            </div>
          </div>

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <h3 className="mb-4 text-sm font-medium text-gray-300">Workload Distribution</h3>
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={workloadData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                    <XAxis dataKey="name" stroke="#9ca3af" fontSize={12} />
                    <YAxis stroke="#9ca3af" fontSize={12} />
                    <Tooltip
                      contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '8px' }}
                      labelStyle={{ color: '#f3f4f6' }}
                    />
                    <Bar dataKey="workload" fill="#3b82f6" radius={[4, 4, 0, 0]} name="Workload" />
                    <Bar dataKey="velocity" fill="#22c55e" radius={[4, 4, 0, 0]} name="Velocity" />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>

            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <h3 className="mb-4 text-sm font-medium text-gray-300">Velocity Trend</h3>
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={velocityTrend}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                    <XAxis dataKey="date" stroke="#9ca3af" fontSize={12} />
                    <YAxis stroke="#9ca3af" fontSize={12} />
                    <Tooltip
                      contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '8px' }}
                      labelStyle={{ color: '#f3f4f6' }}
                    />
                    <Line type="monotone" dataKey="velocity" stroke="#3b82f6" strokeWidth={2} dot={false} />
                    <Line type="monotone" dataKey="collaboration" stroke="#8b5cf6" strokeWidth={2} dot={false} />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </div>

            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <h3 className="mb-4 text-sm font-medium text-gray-300">Collaboration Radar</h3>
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <RadarChart data={radarData}>
                    <PolarGrid stroke="#374151" />
                    <PolarAngleAxis dataKey="subject" stroke="#9ca3af" fontSize={12} />
                    <PolarRadiusAxis stroke="#374151" fontSize={10} />
                    <Radar name="Team" dataKey="score" stroke="#8b5cf6" fill="#8b5cf6" fillOpacity={0.3} />
                  </RadarChart>
                </ResponsiveContainer>
              </div>
            </div>

            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <h3 className="mb-4 text-sm font-medium text-gray-300">PTO Utilization</h3>
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie data={ptoPie} cx="50%" cy="50%" innerRadius={50} outerRadius={80} paddingAngle={4} dataKey="value">
                      <Cell fill="#22c55e" />
                      <Cell fill="#374151" />
                    </Pie>
                    <Legend formatter={(value: string) => <span style={{ color: '#d1d5db', fontSize: '12px' }}>{value}</span>} />
                    <Tooltip contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '8px' }} />
                  </PieChart>
                </ResponsiveContainer>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
