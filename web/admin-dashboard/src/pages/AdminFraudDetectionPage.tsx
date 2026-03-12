import { useQuery } from '@tanstack/react-query';
import { adminApiClient } from '@/lib/api/adminClient';
import { LoadingScreen } from '@/components/common/LoadingScreen';
import { Shield, Bot, Users, Globe, RefreshCw, AlertTriangle } from 'lucide-react';

interface FraudSummary {
  totalBotPatterns?: number;
  highRiskClusters?: number;
  suspiciousTenants?: number;
  washUsageDetected?: number;
}

interface BotPatternItem {
  id: string;
  patternType: string;
  confidenceScore: number;
  affectedFunctions: string[];
  affectedTenants: string[];
  detectedAt: string;
  pattern: string;
}

interface FakeDiversityAlertItem {
  id: string;
  tenantGroup: string[];
  indicators: string[];
  riskLevel: string;
  detectedAt: string;
  commonPatterns: string[];
}

interface IPClusterItem {
  id: string;
  ipRange: string;
  associatedTenants: string[];
  riskLevel: string;
  commonPatterns: string[];
  firstSeen: string;
  lastSeen: string;
}

interface WashUsagePatternItem {
  id: string;
  tenantA: string;
  tenantB: string;
  function: string;
  pattern: string;
  confidence: number;
  reciprocalExecutions: number;
  detectedAt: string;
}

interface FraudDetectionPayload {
  botPatterns?: BotPatternItem[];
  fakeDiversityAlerts?: FakeDiversityAlertItem[];
  ipClusters?: IPClusterItem[];
  washUsagePatterns?: WashUsagePatternItem[];
  summary?: FraudSummary;
}

export function AdminFraudDetectionPage() {
  const { data: raw, isLoading, isError } = useQuery({
    queryKey: ['admin-oversight-fraud-detection'],
    queryFn: async () => {
      try {
        return await adminApiClient.get<FraudDetectionPayload>('/oversight/fraud-detection');
      } catch {
        return null;
      }
    },
  });

  const hasDataWrapper = raw && typeof raw === 'object' && 'data' in raw;
  const payload: FraudDetectionPayload = hasDataWrapper
    ? (raw as { data?: FraudDetectionPayload }).data ?? {}
    : ((raw ?? {}) as unknown as FraudDetectionPayload);

  const summary = payload.summary ?? {};
  const botPatterns = payload.botPatterns ?? [];
  const fakeAlerts = payload.fakeDiversityAlerts ?? [];
  const ipClusters = payload.ipClusters ?? [];
  const washPatterns = payload.washUsagePatterns ?? [];

  if (isLoading) return <LoadingScreen />;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Fraud Detection</h1>
        <p className="mt-2 text-gray-600">Fraud risk insights and suspicious activity signals.</p>
      </div>

      {isError || raw == null ? (
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 text-amber-800">
          <p className="font-medium">Unable to load fraud detection data.</p>
          <p className="text-sm mt-1">The oversight service or registry may be unavailable.</p>
        </div>
      ) : (
        <>
          {/* Summary */}
          <div>
            <h2 className="text-lg font-semibold text-gray-900 mb-3 flex items-center gap-2">
              <Shield className="w-5 h-5 text-emerald-600" />
              Summary
            </h2>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <StatCard label="Bot patterns" value={summary.totalBotPatterns ?? 0} />
              <StatCard label="High-risk clusters" value={summary.highRiskClusters ?? 0} />
              <StatCard label="Suspicious tenants" value={summary.suspiciousTenants ?? 0} />
              <StatCard label="Wash usage detected" value={summary.washUsageDetected ?? 0} />
            </div>
          </div>

          {/* Bot patterns */}
          <Section title="Bot patterns" icon={Bot} count={botPatterns.length}>
            {botPatterns.length === 0 ? (
              <EmptyBlock />
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="bg-gray-50 border-b border-gray-200">
                      <th className="px-4 py-2 text-left font-semibold text-gray-700">Type</th>
                      <th className="px-4 py-2 text-left font-semibold text-gray-700">Confidence</th>
                      <th className="px-4 py-2 text-left font-semibold text-gray-700">Affected</th>
                      <th className="px-4 py-2 text-left font-semibold text-gray-700">Detected</th>
                    </tr>
                  </thead>
                  <tbody>
                    {botPatterns.map((p) => (
                      <tr key={p.id} className="border-b border-gray-100">
                        <td className="px-4 py-2 text-gray-900">{p.patternType || '—'}</td>
                        <td className="px-4 py-2 text-gray-600">{p.confidenceScore ?? 0}%</td>
                        <td className="px-4 py-2 text-gray-600">
                          {(p.affectedTenants?.length ?? 0) + (p.affectedFunctions?.length ?? 0)} items
                        </td>
                        <td className="px-4 py-2 text-gray-500">
                          {p.detectedAt ? new Date(p.detectedAt).toLocaleString() : '—'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Section>

          {/* Fake diversity alerts */}
          <Section title="Fake diversity alerts" icon={Users} count={fakeAlerts.length}>
            {fakeAlerts.length === 0 ? (
              <EmptyBlock />
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="bg-gray-50 border-b border-gray-200">
                      <th className="px-4 py-2 text-left font-semibold text-gray-700">Risk</th>
                      <th className="px-4 py-2 text-left font-semibold text-gray-700">Tenants</th>
                      <th className="px-4 py-2 text-left font-semibold text-gray-700">Detected</th>
                    </tr>
                  </thead>
                  <tbody>
                    {fakeAlerts.map((a) => (
                      <tr key={a.id} className="border-b border-gray-100">
                        <td className="px-4 py-2">
                          <RiskBadge level={a.riskLevel} />
                        </td>
                        <td className="px-4 py-2 text-gray-600">{(a.tenantGroup ?? []).length} tenants</td>
                        <td className="px-4 py-2 text-gray-500">
                          {a.detectedAt ? new Date(a.detectedAt).toLocaleString() : '—'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Section>

          {/* IP clusters */}
          <Section title="IP clusters" icon={Globe} count={ipClusters.length}>
            {ipClusters.length === 0 ? (
              <EmptyBlock />
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="bg-gray-50 border-b border-gray-200">
                      <th className="px-4 py-2 text-left font-semibold text-gray-700">IP range</th>
                      <th className="px-4 py-2 text-left font-semibold text-gray-700">Risk</th>
                      <th className="px-4 py-2 text-left font-semibold text-gray-700">Tenants</th>
                      <th className="px-4 py-2 text-left font-semibold text-gray-700">First / Last seen</th>
                    </tr>
                  </thead>
                  <tbody>
                    {ipClusters.map((c) => (
                      <tr key={c.id} className="border-b border-gray-100">
                        <td className="px-4 py-2 font-mono text-gray-900">{c.ipRange || '—'}</td>
                        <td className="px-4 py-2"><RiskBadge level={c.riskLevel} /></td>
                        <td className="px-4 py-2 text-gray-600">{(c.associatedTenants ?? []).length}</td>
                        <td className="px-4 py-2 text-gray-500 text-xs">
                          {c.firstSeen ? new Date(c.firstSeen).toLocaleDateString() : '—'} –{' '}
                          {c.lastSeen ? new Date(c.lastSeen).toLocaleDateString() : '—'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Section>

          {/* Wash usage patterns */}
          <Section title="Wash usage patterns" icon={RefreshCw} count={washPatterns.length}>
            {washPatterns.length === 0 ? (
              <EmptyBlock />
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="bg-gray-50 border-b border-gray-200">
                      <th className="px-4 py-2 text-left font-semibold text-gray-700">Tenant A / B</th>
                      <th className="px-4 py-2 text-left font-semibold text-gray-700">Function</th>
                      <th className="px-4 py-2 text-left font-semibold text-gray-700">Confidence</th>
                      <th className="px-4 py-2 text-left font-semibold text-gray-700">Reciprocal</th>
                      <th className="px-4 py-2 text-left font-semibold text-gray-700">Detected</th>
                    </tr>
                  </thead>
                  <tbody>
                    {washPatterns.map((w) => (
                      <tr key={w.id} className="border-b border-gray-100">
                        <td className="px-4 py-2 text-gray-900">{w.tenantA ?? '—'} / {w.tenantB ?? '—'}</td>
                        <td className="px-4 py-2 text-gray-600">{w.function ?? '—'}</td>
                        <td className="px-4 py-2 text-gray-600">{w.confidence ?? 0}%</td>
                        <td className="px-4 py-2 text-gray-600">{w.reciprocalExecutions ?? 0}</td>
                        <td className="px-4 py-2 text-gray-500">
                          {w.detectedAt ? new Date(w.detectedAt).toLocaleString() : '—'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Section>
        </>
      )}
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4">
      <p className="text-sm text-gray-600">{label}</p>
      <p className="text-2xl font-bold text-gray-900 mt-1">{value}</p>
    </div>
  );
}

function Section({
  title,
  icon: Icon,
  count,
  children,
}: {
  title: string;
  icon: React.ComponentType<{ className?: string }>;
  count: number;
  children: React.ReactNode;
}) {
  return (
    <div>
      <h2 className="text-lg font-semibold text-gray-900 mb-3 flex items-center gap-2">
        <Icon className="w-5 h-5 text-gray-600" />
        {title}
        {count > 0 && (
          <span className="text-sm font-normal text-gray-500">({count})</span>
        )}
      </h2>
      {children}
    </div>
  );
}

function EmptyBlock() {
  return (
    <div className="bg-white border border-gray-200 rounded-lg p-6 text-center text-gray-500 text-sm">
      No items in this category.
    </div>
  );
}

function RiskBadge({ level }: { level: string }) {
  const l = (level ?? '').toLowerCase();
  const isHigh = l.includes('high') || l === 'critical';
  const isMedium = l.includes('medium') || l.includes('moderate');
  const cls = isHigh ? 'bg-red-100 text-red-800' : isMedium ? 'bg-amber-100 text-amber-800' : 'bg-gray-100 text-gray-800';
  return (
    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium ${cls}`}>
      {isHigh && <AlertTriangle className="w-3 h-3" />}
      {level || '—'}
    </span>
  );
}
