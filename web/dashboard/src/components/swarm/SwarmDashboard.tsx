import { agentApi } from '@/api/agent';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import {
  Activity,
  Brain,
  ChevronRight,
  GitBranch,
  MessageSquare,
  Plus,
  Shield,
  TrendingUp,
  Users,
  Wallet,
} from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';

// Types for Swarm Dashboard
interface Agent {
  id: string;
  name: string;
  status: 'active' | 'suspended' | 'pending';
  swarmRole: 'worker' | 'manager' | 'infrastructure';
  trustScore: number;
  economicScore: number;
  parentAgentId?: string;
  children?: Agent[];
}

interface SwarmStats {
  totalAgents: number;
  activeAgents: number;
  totalMessages: number;
  pendingMessages: number;
  walletBalance: number;
  revenueThisMonth: number;
}

interface SwarmDashboardProps {
  agentId: string;
}

// Map API child (snake_case) to dashboard Agent
function mapChildToAgent(c: {
  id?: string;
  agent_id?: string;
  name: string;
  status: string;
  swarm_role?: string;
  trust_score?: number;
  economic_score?: number;
}): Agent {
  const status = (
    c.status === 'active' || c.status === 'suspended' || c.status === 'pending'
      ? c.status
      : 'active'
  ) as Agent['status'];
  const swarmRole = (
    c.swarm_role === 'worker' || c.swarm_role === 'manager' || c.swarm_role === 'infrastructure'
      ? c.swarm_role
      : 'worker'
  ) as Agent['swarmRole'];
  return {
    id: c.id ?? c.agent_id ?? '',
    name: c.name,
    status,
    swarmRole,
    trustScore: Number(c.trust_score) || 0,
    economicScore: Number(c.economic_score) || 0,
  };
}

export function SwarmDashboard({ agentId }: SwarmDashboardProps) {
  const [stats, setStats] = useState<SwarmStats | null>(null);
  const [children, setChildren] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    if (!agentId) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      const [childrenRes, walletRes, inboxRes, listRes] = await Promise.allSettled([
        agentApi.getChildren(agentId),
        agentApi.getWallet(agentId),
        agentApi.getInbox(agentId),
        agentApi.listAgents({ limit: 1000 }),
      ]);

      const childrenData =
        childrenRes.status === 'fulfilled' && childrenRes.value?.children
          ? (childrenRes.value.children as unknown[]).map(mapChildToAgent)
          : [];
      setChildren(childrenData);

      const wallet =
        walletRes.status === 'fulfilled' && walletRes.value?.wallet
          ? (walletRes.value.wallet as unknown as Record<string, unknown>)
          : null;
      const messages =
        inboxRes.status === 'fulfilled' && inboxRes.value?.messages
          ? (inboxRes.value.messages as { status?: string }[])
          : [];
      const agents =
        listRes.status === 'fulfilled' && listRes.value?.agents
          ? (listRes.value.agents as { status?: string }[])
          : [];

      const activeCount = agents.filter((a) => a.status === 'active').length;
      const pendingMessages = messages.filter((m) => m.status === 'pending' || !m.status).length;

      const balanceUSD =
        wallet &&
        (typeof wallet.balance_usd === 'number'
          ? wallet.balance_usd
          : (wallet as { balanceUSD?: number }).balanceUSD);
      const totalEarnedUSD =
        wallet &&
        (typeof wallet.total_earned_usd === 'number'
          ? wallet.total_earned_usd
          : (wallet as { totalEarnedUSD?: number }).totalEarnedUSD);
      setStats({
        totalAgents: agents.length,
        activeAgents: activeCount,
        totalMessages: messages.length,
        pendingMessages,
        walletBalance: Number(balanceUSD) || 0,
        revenueThisMonth: Number(totalEarnedUSD) || 0,
      });
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load swarm data');
      setStats(null);
      setChildren([]);
    } finally {
      setLoading(false);
    }
  }, [agentId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-destructive">
        <p className="font-medium">Failed to load swarm data</p>
        <p className="text-sm mt-1">{error}</p>
        <Button variant="outline" size="sm" className="mt-3" onClick={() => fetchData()}>
          Retry
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Stats Overview */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total Agents"
          value={stats?.totalAgents ?? 0}
          icon={<Users className="h-4 w-4" />}
          description={`${stats?.activeAgents ?? 0} active`}
        />
        <StatCard
          title="Messages"
          value={stats?.totalMessages ?? 0}
          icon={<MessageSquare className="h-4 w-4" />}
          description={`${stats?.pendingMessages ?? 0} pending`}
          trend="+12%"
        />
        <StatCard
          title="Wallet Balance"
          value={`$${(stats?.walletBalance ?? 0).toFixed(2)}`}
          icon={<Wallet className="h-4 w-4" />}
          description="Available"
        />
        <StatCard
          title="Revenue (Month)"
          value={`$${(stats?.revenueThisMonth ?? 0).toFixed(2)}`}
          icon={<TrendingUp className="h-4 w-4" />}
          description="This period"
          trend="+23%"
        />
      </div>

      {/* Child Agents */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <GitBranch className="h-5 w-5" />
              Child Agents
            </CardTitle>
            <CardDescription>Agents spawned by this manager</CardDescription>
          </div>
          <Button size="sm">
            <Plus className="h-4 w-4 mr-2" />
            Spawn Agent
          </Button>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {children.map((child) => (
              <AgentCard key={child.id} agent={child} />
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Quick Actions */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <ActionCard
          title="Agent Marketplace"
          description="Browse and hire agents"
          icon={<Shield className="h-5 w-5" />}
          href="/marketplace/agents"
        />
        <ActionCard
          title="Function Marketplace"
          description="Discover AI-generated functions"
          icon={<Brain className="h-5 w-5" />}
          href="/dashboard"
        />
        <ActionCard
          title="Evolution Center"
          description="Manage agent learning"
          icon={<Activity className="h-5 w-5" />}
          href="/evolution"
        />
        <ActionCard
          title="Agent Wallet"
          description="View balance and transactions"
          icon={<Wallet className="h-5 w-5" />}
          href="/wallet"
        />
      </div>
    </div>
  );
}

function StatCard({
  title,
  value,
  icon,
  description,
  trend,
}: {
  title: string;
  value: string | number;
  icon: React.ReactNode;
  description?: string;
  trend?: string;
}) {
  return (
    <Card>
      <CardContent className="pt-6">
        <div className="flex items-start justify-between">
          <div className="space-y-1">
            <p className="text-sm font-medium text-muted-foreground">{title}</p>
            <p className="text-2xl font-bold">{value}</p>
            {description && <p className="text-xs text-muted-foreground">{description}</p>}
          </div>
          <div className="p-2 bg-primary/10 rounded-lg text-primary">{icon}</div>
        </div>
        {trend && (
          <div className="mt-2 flex items-center text-xs text-green-600">
            <TrendingUp className="h-3 w-3 mr-1" />
            {trend}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function AgentCard({ agent }: { agent: Agent }) {
  const statusColors = {
    active: 'bg-green-500',
    suspended: 'bg-red-500',
    pending: 'bg-yellow-500',
  };

  const roleColors = {
    worker: 'bg-blue-500',
    manager: 'bg-purple-500',
    infrastructure: 'bg-orange-500',
  };

  return (
    <div className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 transition-colors">
      <div className="flex items-center gap-4">
        <div className={`h-3 w-3 rounded-full ${statusColors[agent.status]}`} />
        <div>
          <p className="font-medium">{agent.name}</p>
          <div className="flex items-center gap-2 mt-1">
            <Badge variant="outline" className={roleColors[agent.swarmRole]}>
              {agent.swarmRole}
            </Badge>
          </div>
        </div>
      </div>
      <div className="flex items-center gap-6">
        <div className="text-center">
          <p className="text-xs text-muted-foreground">Trust</p>
          <Progress value={agent.trustScore} className="h-2 w-20" />
          <p className="text-xs text-right mt-1">{agent.trustScore}%</p>
        </div>
        <div className="text-center">
          <p className="text-xs text-muted-foreground">Economic</p>
          <Progress value={agent.economicScore} className="h-2 w-20" />
          <p className="text-xs text-right mt-1">{agent.economicScore}%</p>
        </div>
        <Button variant="ghost" size="sm">
          <ChevronRight className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}

function ActionCard({
  title,
  description,
  icon,
  href,
}: {
  title: string;
  description: string;
  icon: React.ReactNode;
  href: string;
}) {
  return (
    <Card className="hover:border-primary/50 transition-colors cursor-pointer">
      <CardContent className="pt-6">
        <div className="flex items-start gap-4">
          <div className="p-3 bg-primary/10 rounded-lg">{icon}</div>
          <div>
            <h3 className="font-medium">{title}</h3>
            <p className="text-sm text-muted-foreground mt-1">{description}</p>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export default SwarmDashboard;
