import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { FunctionSquare, Database, Bot } from 'lucide-react';
import { formatLimit } from '../utils';

interface ResourcesTabProps {
  displayPlan: string;
  displayName?: string;

  // Limits
  requestLimit: number;
  functionsLimit: number;
  providersLimit: number;
  stateFabricsLimit: number;
  agentsLimit: number;
  customDomainsLimit?: number;

  // Usage counts
  functionsCount: number;
  providersCount: number;
  stateFabricsCount: number;
  agentIds: string[];
  agentsUsageAndBalance?: {
    totalCallsToday?: number;
    totalSpendToday?: number;
    totalBalanceUsd?: number;
  } | null;

  // State fabric metrics
  stateFabricTotals: {
    operations: number;
    storage: number;
  };
}

export function ResourcesTab({
  displayPlan,
  displayName,
  requestLimit,
  functionsLimit,
  providersLimit,
  stateFabricsLimit,
  agentsLimit,
  customDomainsLimit = 0,
  functionsCount,
  providersCount,
  stateFabricsCount,
  agentIds,
  agentsUsageAndBalance,
  stateFabricTotals,
}: ResourcesTabProps) {
  return (
    <div className="space-y-4">
      {/* Plan Limits */}
      <Card className="border-theme bg-card">
        <CardHeader>
          <CardTitle className="text-base">Your plan limits</CardTitle>
          <CardDescription>
            Maximum included in your {displayName || displayPlan} plan.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3 text-sm">
            <div className="rounded-lg border border-border-subtle bg-bg-secondary p-3">
              <p className="text-text-muted font-medium">Requests/mo</p>
              <p className="text-text-primary font-semibold">{formatLimit(requestLimit)}</p>
            </div>
            <div className="rounded-lg border border-border-subtle bg-bg-secondary p-3">
              <p className="text-text-muted font-medium">Functions</p>
              <p className="text-text-primary font-semibold">{formatLimit(functionsLimit)}</p>
            </div>
            <div className="rounded-lg border border-border-subtle bg-bg-secondary p-3">
              <p className="text-text-muted font-medium">Providers</p>
              <p className="text-text-primary font-semibold">{formatLimit(providersLimit)}</p>
            </div>
            <div className="rounded-lg border border-border-subtle bg-bg-secondary p-3">
              <p className="text-text-muted font-medium">Custom domains</p>
              <p className="text-text-primary font-semibold">{formatLimit(customDomainsLimit)}</p>
            </div>
            <div className="rounded-lg border border-border-subtle bg-bg-secondary p-3">
              <p className="text-text-muted font-medium">State Fabrics</p>
              <p className="text-text-primary font-semibold">{formatLimit(stateFabricsLimit)}</p>
            </div>
            <div className="rounded-lg border border-border-subtle bg-bg-secondary p-3">
              <p className="text-text-muted font-medium">Agents</p>
              <p className="text-text-primary font-semibold">{formatLimit(agentsLimit)}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Resource Usage Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {/* Functions & Providers */}
        <Card className="border-theme bg-card">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <FunctionSquare className="h-4 w-4" />
              Functions & Providers
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex justify-between items-center">
                <span className="text-sm text-text-secondary">Functions</span>
                <span className="font-medium">
                  {functionsCount} / {formatLimit(functionsLimit)}
                </span>
              </div>
              <Progress
                value={functionsLimit > 0 ? (functionsCount / functionsLimit) * 100 : 0}
              />
              <div className="flex justify-between items-center">
                <span className="text-sm text-text-secondary">Providers</span>
                <span className="font-medium">
                  {providersCount} / {formatLimit(providersLimit)}
                </span>
              </div>
              <Progress
                value={providersLimit > 0 ? (providersCount / providersLimit) * 100 : 0}
              />
            </div>
          </CardContent>
        </Card>

        {/* State Fabric */}
        <Card className="border-theme bg-card">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Database className="h-4 w-4" />
              State Fabric
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex justify-between items-center">
                <span className="text-sm text-text-secondary">Fabrics</span>
                <span className="font-medium">
                  {stateFabricsCount} / {formatLimit(stateFabricsLimit)}
                </span>
              </div>
              <Progress
                value={
                  stateFabricsLimit > 0 ? (stateFabricsCount / stateFabricsLimit) * 100 : 0
                }
              />
              <div className="flex justify-between items-center">
                <span className="text-sm text-text-secondary">Operations</span>
                <span className="font-medium">
                  {stateFabricTotals.operations.toLocaleString()}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-text-secondary">Storage</span>
                <span className="font-medium">
                  {stateFabricTotals.storage >= 1024
                    ? `${(stateFabricTotals.storage / 1024).toFixed(1)} GB`
                    : `${stateFabricTotals.storage} MB`}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Agents */}
        <Card className="border-theme bg-card">
          <CardHeader>
            <CardTitle className="text-base flex items-center gap-2">
              <Bot className="h-4 w-4" />
              Agents
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex justify-between items-center">
                <span className="text-sm text-text-secondary">Agents</span>
                <span className="font-medium">
                  {agentIds.length} / {formatLimit(agentsLimit)}
                </span>
              </div>
              <Progress
                value={agentsLimit > 0 ? (agentIds.length / agentsLimit) * 100 : 0}
              />
              <div className="flex justify-between items-center">
                <span className="text-sm text-text-secondary">Calls today</span>
                <span className="font-medium">
                  {(agentsUsageAndBalance?.totalCallsToday ?? 0).toLocaleString()}
                </span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-text-secondary">Spend today</span>
                <span className="font-medium">
                  ${(agentsUsageAndBalance?.totalSpendToday ?? 0).toFixed(2)}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
