import { useQuery } from '@tanstack/react-query';
import { identityApi } from '@/api/identity';
import { employeesApi } from '@/api/employees';
import { WalletSection } from '@/components/identity/WalletSection';
import { IdentitySignature } from '@/components/identity/IdentitySignature';
import { ClearanceBadge } from '@/components/identity/ClearanceBadge';
import { Wallet, Shield, Award, Banknote, GraduationCap } from 'lucide-react';

export function IdentityWalletPage() {
  const { data: employeeData } = useQuery({
    queryKey: ['employee', 'me'],
    queryFn: () => employeesApi.list({ limit: 1 }).then((r) => ({ data: { employee: r.data.employees[0] } })),
  });

  const employee = employeeData?.data?.employee;
  const employeeId = employee?.id || '';

  const { data: cardData } = useQuery({
    queryKey: ['identity-card', employeeId],
    queryFn: () => identityApi.getCard(employeeId),
    enabled: !!employeeId,
  });

  const { data: achievementsData } = useQuery({
    queryKey: ['achievements'],
    queryFn: () => identityApi.getAchievements(),
  });

  const card = cardData?.data?.card;
  const achievements = achievementsData?.data?.definitions || [];
  const earnedAchievements = achievements
    .filter((a) => a.earned)
    .map((a) => ({
      id: a.id,
      name: a.name,
      description: a.description,
      icon: a.icon,
      earned_at: a.earned_at,
    }));

  const badges = earnedAchievements.filter((a) => a.icon?.includes('badge') || false);
  const awards = earnedAchievements.filter((a) => !a.icon?.includes('badge'));
  const credentials = card?.skills?.map((s) => ({
    id: s.name,
    name: s.name,
    description: `Level ${s.level}`,
    icon: '📜',
  })) || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Wallet className="h-6 w-6 text-emerald-400" />
        <h1 className="text-2xl font-bold">Identity Wallet</h1>
      </div>

      {employee && card && (
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-5">
          <div className="flex items-center gap-4">
            <div className="flex h-14 w-14 items-center justify-center rounded-full bg-gradient-to-br from-blue-600 to-purple-600">
              <span className="text-lg font-bold text-white">{employee.ffid.slice(-4)}</span>
            </div>
            <div>
              <h2 className="font-semibold text-gray-100">{employee.ffid}</h2>
              <div className="mt-1 flex flex-wrap items-center gap-2">
                <IdentitySignature signature={card.identity_signature} />
                <ClearanceBadge level={card.clearance_level_num} />
              </div>
            </div>
            <div className="ml-auto text-right">
              <p className="text-sm text-gray-500">Total Assets</p>
              <p className="text-xl font-bold text-gray-100">
                {credentials.length + earnedAchievements.length}
              </p>
            </div>
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <WalletSection title="Credentials" type="credentials" items={credentials} />
        <WalletSection title="Badges" type="badges" items={badges} />
        <WalletSection title="Awards" type="awards" items={awards} />
        <WalletSection
          title="Training"
          type="training"
          items={[]}
        />
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        {[
          { icon: Shield, label: 'Clearance Level', value: card ? `L${card.clearance_level_num}` : '—', color: 'text-blue-400' },
          { icon: Award, label: 'Achievements', value: earnedAchievements.length.toString(), color: 'text-yellow-400' },
          { icon: Banknote, label: 'Reputation', value: card?.reputation_total?.toString() || '—', color: 'text-green-400' },
        ].map((stat) => (
          <div key={stat.label} className="rounded-xl border border-gray-800 bg-gray-900 p-5 text-center">
            <stat.icon className={`mx-auto h-6 w-6 ${stat.color}`} />
            <p className="mt-2 text-2xl font-bold text-gray-100">{stat.value}</p>
            <p className="text-xs text-gray-500">{stat.label}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
