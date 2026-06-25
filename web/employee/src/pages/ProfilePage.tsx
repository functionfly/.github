import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { identityApi } from '@/api/identity';
import { employeesApi } from '@/api/employees';
import { useAuthStore } from '@/stores/authStore';
import { IdentityCard } from '@/components/identity/IdentityCard';
import { IdentityLayers } from '@/components/identity/IdentityLayers';
import { ReputationRadar } from '@/components/identity/ReputationRadar';
import { AchievementGrid } from '@/components/identity/AchievementGrid';
import { CareerTimeline } from '@/components/identity/CareerTimeline';
import { User } from 'lucide-react';

export function ProfilePage() {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuthStore();

  const { data: employeeData, isLoading } = useQuery({
    queryKey: ['employee', id || 'me'],
    queryFn: () =>
      id
        ? employeesApi.get(id)
        : employeesApi.list({ limit: 1 }).then((r) => ({ data: { employee: r.data.employees[0] } })),
  });

  const employee = employeeData?.data?.employee;
  const employeeId = employee?.id || '';

  const { data: cardData } = useQuery({
    queryKey: ['identity-card', employeeId],
    queryFn: () => identityApi.getCard(employeeId),
    enabled: !!employeeId,
  });

  const { data: trendsData } = useQuery({
    queryKey: ['reputation-trends', employeeId],
    queryFn: () => identityApi.getReputationTrends(employeeId),
    enabled: !!employeeId,
  });

  const card = cardData?.data?.card;
  const trends = trendsData?.data?.trends || [];

  const radarScores: Record<string, number> = {};
  for (const trend of trends) {
    const latest = trend.history[trend.history.length - 1];
    if (latest) radarScores[trend.category] = latest.score;
  }

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
      </div>
    );
  }

  if (!employee) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <User className="mb-4 h-12 w-12 text-gray-600" />
        <p className="text-gray-400">Employee profile not found</p>
      </div>
    );
  }

  if (!card) {
    return (
      <div className="space-y-6">
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-6">
          <div className="flex items-start gap-6">
            <div className="flex h-20 w-20 items-center justify-center rounded-full bg-gradient-to-br from-blue-600 to-purple-600">
              <span className="text-2xl font-bold text-white">{employee.ffid.slice(-4)}</span>
            </div>
            <div className="flex-1">
              <h1 className="text-2xl font-bold text-gray-100">{user?.name || employee.ffid}</h1>
              <p className="mt-1 font-mono text-lg text-blue-400">{employee.ffid}</p>
            </div>
          </div>
        </div>
        <IdentityLayers employee={employee} />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <IdentityCard
        employee={card.employee}
        identitySignature={card.identity_signature}
        clearanceLevel={card.clearance_level_num}
        reputationTotal={card.reputation_total}
        trustScore={card.trust_score}
      />

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          <IdentityLayers employee={card.employee} />
          <CareerTimeline events={card.recent_timeline} />
        </div>
        <div className="space-y-6">
          <ReputationRadar scores={radarScores} />
          <AchievementGrid achievements={card.achievements} />
        </div>
      </div>
    </div>
  );
}
