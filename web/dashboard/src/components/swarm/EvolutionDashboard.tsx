import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { 
  Brain, 
  TrendingUp, 
  TrendingDown, 
  AlertTriangle,
  CheckCircle,
  XCircle,
  Clock,
  Activity,
  Zap,
  Target,
  Sparkles,
  ChevronRight
} from 'lucide-react';

// Types for Evolution Dashboard
interface EvolutionProposal {
  id: string;
  type: 'spawn_specialist' | 'modify_policy' | 'adjust_timeout' | 'generate_function' | 'retire_child';
  status: 'pending' | 'approved' | 'rejected' | 'implemented' | 'expired';
  createdAt: string;
  description: string;
  impact?: {
    successRate?: number;
    latency?: number;
    cost?: number;
  };
}

interface PerformanceMetrics {
  totalExecutions: number;
  successRate: number;
  avgLatency: number;
  avgCost: number;
  failureCategories: Record<string, number>;
}

interface TrendData {
  date: string;
  successRate: number;
  latency: number;
  cost: number;
}

export function EvolutionDashboard({ agentId }: { agentId: string }) {
  const [proposals, setProposals] = useState<EvolutionProposal[]>([]);
  const [metrics, setMetrics] = useState<PerformanceMetrics | null>(null);
  const [trends, setTrends] = useState<TrendData[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Mock data - in production fetch from API
    setTimeout(() => {
      setMetrics({
        totalExecutions: 15432,
        successRate: 94.5,
        avgLatency: 1250,
        avgCost: 0.012,
        failureCategories: {
          timeout: 45,
          policy_violation: 23,
          network: 12,
          unknown: 8
        }
      });
      setProposals([
        {
          id: '1',
          type: 'spawn_specialist',
          status: 'pending',
          createdAt: '2026-03-01T10:30:00Z',
          description: 'Spawn error handler specialist to improve failure recovery',
          impact: { successRate: 5 }
        },
        {
          id: '2',
          type: 'modify_policy',
          status: 'approved',
          createdAt: '2026-02-28T14:20:00Z',
          description: 'Increase timeout to handle larger payloads',
          impact: { latency: -200 }
        },
        {
          id: '3',
          type: 'generate_function',
          status: 'implemented',
          createdAt: '2026-02-25T09:15:00Z',
          description: 'Generated custom data transformation function',
          impact: { successRate: 3, cost: -15 }
        }
      ]);
      setTrends([
        { date: '2026-02-22', successRate: 91, latency: 1400, cost: 0.015 },
        { date: '2026-02-23', successRate: 92, latency: 1350, cost: 0.014 },
        { date: '2026-02-24', successRate: 93, latency: 1300, cost: 0.013 },
        { date: '2026-02-25', successRate: 94, latency: 1280, cost: 0.013 },
        { date: '2026-02-26', successRate: 94.2, latency: 1260, cost: 0.012 },
        { date: '2026-02-27', successRate: 94.5, latency: 1250, cost: 0.012 },
        { date: '2026-02-28', successRate: 94.8, latency: 1240, cost: 0.011 }
      ]);
      setLoading(false);
    }, 800);
  }, [agentId]);

  const getStatusBadge = (status: string) => {
    const config = {
      pending: { color: 'bg-yellow-500', icon: <Clock className="h-3 w-3" /> },
      approved: { color: 'bg-blue-500', icon: <CheckCircle className="h-3 w-3" /> },
      rejected: { color: 'bg-red-500', icon: <XCircle className="h-3 w-3" /> },
      implemented: { color: 'bg-green-500', icon: <CheckCircle className="h-3 w-3" /> },
      expired: { color: 'bg-gray-500', icon: <Clock className="h-3 w-3" /> }
    };
    const c = config[status as keyof typeof config];
    return (
      <Badge className={`${c.color} text-white`}>
        {c.icon}
        <span className="ml-1 capitalize">{status}</span>
      </Badge>
    );
  };

  const getTypeLabel = (type: string) => {
    const labels = {
      spawn_specialist: 'Spawn Specialist',
      modify_policy: 'Modify Policy',
      adjust_timeout: 'Adjust Timeout',
      generate_function: 'Generate Function',
      retire_child: 'Retire Child'
    };
    return labels[type as keyof typeof labels] || type;
  };

  const getTypeIcon = (type: string) => {
    const icons = {
      spawn_specialist: <Zap className="h-4 w-4" />,
      modify_policy: <Target className="h-4 w-4" />,
      adjust_timeout: <Clock className="h-4 w-4" />,
      generate_function: <Sparkles className="h-4 w-4" />,
      retire_child: <TrendingDown className="h-4 w-4" />
    };
    return icons[type as keyof typeof icons] || <Brain className="h-4 w-4" />;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <Brain className="h-8 w-8" />
            Evolution Center
          </h1>
          <p className="text-muted-foreground mt-1">
            Agent learning and autonomous improvement
          </p>
        </div>
        <Button>
          <Sparkles className="h-4 w-4 mr-2" />
          Analyze & Propose
        </Button>
      </div>

      {/* Performance Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Total Executions</p>
                <p className="text-2xl font-bold">{metrics?.totalExecutions.toLocaleString()}</p>
              </div>
              <Activity className="h-8 w-8 text-blue-500" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Success Rate</p>
                <p className="text-2xl font-bold">{metrics?.successRate}%</p>
              </div>
              <TrendingUp className="h-8 w-8 text-green-500" />
            </div>
            <Progress value={metrics?.successRate ?? 0} className="mt-2" />
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Avg Latency</p>
                <p className="text-2xl font-bold">{(metrics?.avgLatency ?? 0)}ms</p>
              </div>
              <Clock className="h-8 w-8 text-purple-500" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Avg Cost</p>
                <p className="text-2xl font-bold">${(metrics?.avgCost ?? 0).toFixed(3)}</p>
              </div>
              <Zap className="h-8 w-8 text-yellow-500" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="proposals" className="space-y-4">
        <TabsList>
          <TabsTrigger value="proposals">Proposals</TabsTrigger>
          <TabsTrigger value="trends">Performance Trends</TabsTrigger>
          <TabsTrigger value="failures">Failure Analysis</TabsTrigger>
        </TabsList>

        <TabsContent value="proposals" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Evolution Proposals</CardTitle>
              <CardDescription>
                AI-generated suggestions to improve agent performance
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {proposals.map((proposal) => (
                  <div 
                    key={proposal.id} 
                    className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 transition-colors"
                  >
                    <div className="flex items-center gap-4">
                      <div className="p-2 bg-primary/10 rounded-lg">
                        {getTypeIcon(proposal.type)}
                      </div>
                      <div>
                        <p className="font-medium">{getTypeLabel(proposal.type)}</p>
                        <p className="text-sm text-muted-foreground">{proposal.description}</p>
                        <div className="flex items-center gap-2 mt-1">
                          <Clock className="h-3 w-3 text-muted-foreground" />
                          <span className="text-xs text-muted-foreground">
                            {new Date(proposal.createdAt).toLocaleDateString()}
                          </span>
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-4">
                      {proposal.impact && (
                        <div className="text-right text-sm">
                          {proposal.impact.successRate && (
                            <span className={proposal.impact.successRate > 0 ? 'text-green-500' : 'text-red-500'}>
                              {proposal.impact.successRate > 0 ? '+' : ''}{proposal.impact.successRate}% success
                            </span>
                          )}
                          {proposal.impact.latency && (
                            <span className="ml-2 text-blue-500">
                              {proposal.impact.latency}ms latency
                            </span>
                          )}
                        </div>
                      )}
                      {getStatusBadge(proposal.status)}
                      <Button variant="ghost" size="sm">
                        <ChevronRight className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="trends">
          <Card>
            <CardHeader>
              <CardTitle>Performance Trends</CardTitle>
              <CardDescription>
                Agent performance over the last 7 days
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="h-64 flex items-end justify-between gap-2">
                {trends.map((day, i) => (
                  <div key={day.date} className="flex-1 flex flex-col items-center gap-2">
                    <div 
                      className="w-full bg-blue-500 rounded-t"
                      style={{ height: `${day.successRate}%` }}
                    />
                    <span className="text-xs text-muted-foreground">
                      {new Date(day.date).toLocaleDateString('en-US', { weekday: 'short' })}
                    </span>
                  </div>
                ))}
              </div>
              <div className="flex justify-center gap-6 mt-4">
                <div className="flex items-center gap-2">
                  <div className="h-3 w-3 bg-blue-500 rounded" />
                  <span className="text-sm">Success Rate</span>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="failures">
          <Card>
            <CardHeader>
              <CardTitle>Failure Analysis</CardTitle>
              <CardDescription>
                Breakdown of failure categories
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {Object.entries(metrics?.failureCategories ?? {}).map(([category, count]) => (
                  <div key={category} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <AlertTriangle className="h-4 w-4 text-yellow-500" />
                      <span className="capitalize">{category.replace('_', ' ')}</span>
                    </div>
                    <div className="flex items-center gap-4">
                      <Progress 
                        value={(count / Object.values(metrics?.failureCategories ?? {}).reduce((a, b) => a + b, 0)) * 100} 
                        className="w-32" 
                      />
                      <span className="font-medium">{count}</span>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default EvolutionDashboard;
