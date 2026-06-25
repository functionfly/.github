import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { identityApi } from '@/api/identity';
import { useAuthStore } from '@/stores/authStore';
import { IdentityCard } from '@/components/identity/IdentityCard';
import { IdentityLayers } from '@/components/identity/IdentityLayers';
import { ReputationRadar } from '@/components/identity/ReputationRadar';
import { AchievementGrid } from '@/components/identity/AchievementGrid';
import { CareerTimeline } from '@/components/identity/CareerTimeline';
import { employeesApi } from '@/api/employees';
import { BookOpen } from 'lucide-react';

export function PassportPage() {
  const { id } = useParams<{ id: string }>();
  const { user } = useAuthStore();

  const { data: employeeData } = useQuery({
    queryKey: ['employee', id || 'me'],
    queryFn: () =>
      id
        ? employeesApi.get(id)
        : employeesApi.list({ limit: 1 }).then((r) => ({ data: { employee: r.data.employees[0] } })),
  });

  const employee = employeeData?.data?.employee;
  const employeeId = employee?.id || '';

  const { data: cardData, isLoading: cardLoading } = useQuery({
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

  if (cardLoading) {
    return (
      <div className="flex justify-center py-12">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
      </div>
    );
  }

  if (!employee || !card) {
    return (
      <div className="flex flex-col items-center justify-center py-12">
        <BookOpen className="mb-4 h-12 w-12 text-gray-600" />
        <p className="text-gray-400">Passport data not available</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <BookOpen className="h-6 w-6 text-blue-400" />
        <h1 className="text-2xl font-bold">FunctionFly Passport</h1>
      </div>

      <IdentityCard
        employee={card.employee}
        identitySignature={card.identity_signature}
        clearanceLevel={card.clearance_level_num}
        reputationTotal={card.reputation_total}
        trustScore={card.trust_score}
      />

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <IdentityLayers employee={card.employee} />
        </div>
        <div className="space-y-6">
          <ReputationRadar scores={radarScores} />
          <AchievementGrid achievements={card.achievements} />
        </div>
      </div>

      <CareerTimeline events={card.recent_timeline} />
    </div>
  );
}
