import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { 
  Users, 
  GitBranch, 
  MessageSquare, 
  Wallet, 
  TrendingUp, 
  Shield,
  Activity,
  Plus,
  ChevronRight,
  Brain,
  Clock,
  DollarSign
} from 'lucide-react';

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

export function SwarmDashboard({ agentId }: SwarmDashboardProps) {
  const [stats, setStats] = useState<SwarmStats | null>(null);
  const [children, setChildren] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // In production, fetch from API
    setTimeout(() => {
      setStats({
        totalAgents: 12,
        activeAgents: 10,
        totalMessages: 1543,
        pendingMessages: 7,
        walletBalance: 1250.00,
        revenueThisMonth: 3420.50
      });
      setChildren([
        {
          id: 'agent-1',
          name: 'Data Processor',
          status: 'active',
          swarmRole: 'worker',
          trustScore: 95,
          economicScore: 88
        },
        {
          id: 'agent-2',
          name: 'Analytics Worker',
          status: 'active',
          swarmRole: 'worker',
          trustScore: 92,
          economicScore: 85
        },
        {
          id: 'agent-3',
          name: 'Error Handler',
          status: 'active',
          swarmRole: 'worker',
          trustScore: 98,
          economicScore: 91
        }
      ]);
      setLoading(false);
    }, 1000);
  }, [agentId]);

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
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
            <CardDescription>
              Agents spawned by this manager
            </CardDescription>
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
          href="/marketplace/functions"
        />
        <ActionCard
          title="Evolution Center"
          description="Manage agent learning"
          icon={<Activity className="h-5 w-5" />}
          href="/evolution"
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
  trend 
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
            {description && (
              <p className="text-xs text-muted-foreground">{description}</p>
            )}
          </div>
          <div className="p-2 bg-primary/10 rounded-lg text-primary">
            {icon}
          </div>
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
    pending: 'bg-yellow-500'
  };

  const roleColors = {
    worker: 'bg-blue-500',
    manager: 'bg-purple-500',
    infrastructure: 'bg-orange-500'
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
  href 
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
          <div className="p-3 bg-primary/10 rounded-lg">
            {icon}
          </div>
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
