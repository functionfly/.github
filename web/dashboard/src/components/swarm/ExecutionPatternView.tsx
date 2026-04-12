import { useState, useEffect, useCallback } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Progress } from '@/components/ui/progress';
import { 
  Activity, 
  TrendingUp,
  TrendingDown,
  AlertTriangle,
  CheckCircle,
  Clock,
  DollarSign,
  Zap,
  RefreshCw,
  Loader2,
  BarChart3,
  PieChart,
  ArrowUpRight,
  ArrowDownRight,
  Lightbulb
} from 'lucide-react';
import { agentApi } from '@/api/agent';
import { toast } from 'sonner';

interface Pattern {
  id: string;
  pattern_type: string;
  confidence: number;
  occurrence_count: number;
  recommendations: string[];
  first_seen_at: string;
  last_seen_at: string;
}

interface Optimization {
  id: string;
  optimization_type: string;
  description: string;
  expected_impact: Record<string, number>;
  implementation: 'low' | 'medium' | 'high';
  status: 'pending' | 'approved' | 'rejected' | 'applied';
  created_at: string;
}

interface AnalysisData {
  agent_id: string;
  total_executions: number;
  patterns: Pattern[];
  insights: string[];
  success_rate: number;
  avg_latency_ms: number;
  avg_cost_usd: number;
}

interface ExecutionPatternViewProps {
  agentId: string;
  agentName: string;
}

const patternIcons: Record<string, typeof Activity> = {
  frequent_failure: AlertTriangle,
  slow_execution: Clock,
  cost_inefficient: DollarSign,
  high_retry_rate: RefreshCw,
  resource_contention: Zap,
  successful: CheckCircle,
  optimal: TrendingUp,
};

const patternColors: Record<string, string> = {
  frequent_failure: 'bg-red-100 text-red-800 border-red-200',
  slow_execution: 'bg-yellow-100 text-yellow-800 border-yellow-200',
  cost_inefficient: 'bg-orange-100 text-orange-800 border-orange-200',
  high_retry_rate: 'bg-purple-100 text-purple-800 border-purple-200',
  resource_contention: 'bg-blue-100 text-blue-800 border-blue-200',
  successful: 'bg-green-100 text-green-800 border-green-200',
  optimal: 'bg-emerald-100 text-emerald-800 border-emerald-200',
};

const optimizationTypes: Record<string, string> = {
  timeout_adjustment: 'Timeout Adjustment',
  caching: 'Caching Strategy',
  batch_processing: 'Batch Processing',
  resource_upgrade: 'Resource Upgrade',
  policy_change: 'Policy Change',
  retry_strategy: 'Retry Strategy',
  query_optimization: 'Query Optimization',
};

const implementationColors = {
  low: 'bg-green-500',
  medium: 'bg-yellow-500',
  high: 'bg-red-500',
};

export function ExecutionPatternView({ agentId, agentName }: ExecutionPatternViewProps) {
  const [analysis, setAnalysis] = useState<AnalysisData | null>(null);
  const [optimizations, setOptimizations] = useState<Optimization[]>([]);
  const [loading, setLoading] = useState(true);
  const [optimizing, setOptimizing] = useState(false);
  const [timeRange, setTimeRange] = useState(7);

  const loadAnalysis = useCallback(async () => {
    setLoading(true);
    try {
      const data = await agentApi.analyzeAgent(agentId, { days: timeRange });
      setAnalysis(data.analysis);
    } catch (err) {
      console.error('Failed to load analysis:', err);
      toast.error('Failed to load execution patterns');
    } finally {
      setLoading(false);
    }
  }, [agentId, timeRange]);

  const loadInsights = useCallback(async () => {
    try {
      const data = await agentApi.getInsights(agentId);
      setOptimizations(data.optimizations as Optimization[]);
    } catch (err) {
      console.error('Failed to load insights:', err);
    }
  }, [agentId]);

  useEffect(() => {
    loadAnalysis();
    loadInsights();
  }, [loadAnalysis, loadInsights]);

  const handleOptimize = async () => {
    setOptimizing(true);
    try {
      await agentApi.optimizeAgent(agentId);
      toast.success('Optimization recommendations generated');
      await loadInsights();
    } catch (err) {
      toast.error('Failed to generate optimizations');
    } finally {
      setOptimizing(false);
    }
  };

  const getHealthScore = () => {
    if (!analysis) return 0;
    let score = 100;
    
    // Deduct for failures
    score -= (100 - analysis.success_rate) * 0.5;
    
    // Deduct for high latency (> 5s is bad)
    if (analysis.avg_latency_ms > 5000) {
      score -= Math.min(30, (analysis.avg_latency_ms - 5000) / 100);
    }
    
    // Deduct for high cost (>$0.10 is expensive)
    if (analysis.avg_cost_usd > 0.10) {
      score -= Math.min(20, (analysis.avg_cost_usd - 0.10) * 200);
    }
    
    return Math.max(0, Math.round(score));
  };

  const healthScore = getHealthScore();
  const healthStatus = healthScore >= 80 ? 'healthy' : healthScore >= 50 ? 'degraded' : 'critical';
  const healthColor = healthScore >= 80 ? 'bg-green-500' : healthScore >= 50 ? 'bg-yellow-500' : 'bg-red-500';

  if (loading && !analysis) {
    return (
      <Card className="min-h-[400px]">
        <CardContent className="flex items-center justify-center py-20">
          <Loader2 className="h-10 w-10 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold flex items-center gap-2">
            <Activity className="h-6 w-6" />
            Execution Patterns & Optimization
          </h2>
          <p className="text-muted-foreground">
            Analysis and recommendations for {agentName}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={timeRange}
            onChange={(e) => setTimeRange(Number(e.target.value))}
            className="h-9 rounded-md border border-input bg-background px-3 text-sm"
          >
            <option value={1}>Last 24 hours</option>
            <option value={7}>Last 7 days</option>
            <option value={30}>Last 30 days</option>
            <option value={90}>Last 90 days</option>
          </select>
          <Button onClick={handleOptimize} disabled={optimizing}>
            {optimizing ? (
              <Loader2 className="h-4 w-4 animate-spin mr-2" />
            ) : (
              <Zap className="h-4 w-4 mr-2" />
            )}
            Optimize
          </Button>
        </div>
      </div>

      {/* Health Overview */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium text-muted-foreground">Health Score</span>
              {healthScore >= 80 ? (
                <CheckCircle className="h-4 w-4 text-green-500" />
              ) : healthScore >= 50 ? (
                <AlertTriangle className="h-4 w-4 text-yellow-500" />
              ) : (
                <AlertTriangle className="h-4 w-4 text-red-500" />
              )}
            </div>
            <div className="flex items-end gap-2">
              <span className="text-3xl font-bold">{healthScore}</span>
              <span className="text-xs text-muted-foreground mb-1">/100</span>
            </div>
            <Progress value={healthScore} className={`mt-2 ${healthColor}`} />
            <p className="text-xs text-muted-foreground mt-2 capitalize">{healthStatus}</p>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium text-muted-foreground">Success Rate</span>
              <CheckCircle className="h-4 w-4 text-muted-foreground" />
            </div>
            <div className="flex items-end gap-2">
              <span className="text-3xl font-bold">
                {analysis?.success_rate.toFixed(1)}%
              </span>
            </div>
            <div className="flex items-center gap-1 mt-2 text-xs">
              {analysis && analysis.success_rate >= 95 ? (
                <ArrowUpRight className="h-3 w-3 text-green-500" />
              ) : (
                <ArrowDownRight className="h-3 w-3 text-red-500" />
              )}
              <span className={analysis && analysis.success_rate >= 95 ? 'text-green-600' : 'text-red-600'}>
                {analysis && analysis.success_rate >= 95 ? 'Excellent' : 'Needs improvement'}
              </span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium text-muted-foreground">Avg Latency</span>
              <Clock className="h-4 w-4 text-muted-foreground" />
            </div>
            <div className="flex items-end gap-2">
              <span className="text-3xl font-bold">
                {(analysis?.avg_latency_ms || 0).toFixed(0)}ms
              </span>
            </div>
            <div className="flex items-center gap-1 mt-2 text-xs">
              {analysis && analysis.avg_latency_ms < 1000 ? (
                <ArrowUpRight className="h-3 w-3 text-green-500" />
              ) : (
                <ArrowDownRight className="h-3 w-3 text-yellow-500" />
              )}
              <span className={analysis && analysis.avg_latency_ms < 1000 ? 'text-green-600' : 'text-yellow-600'}>
                {analysis && analysis.avg_latency_ms < 1000 ? 'Fast' : 'Moderate'}
              </span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium text-muted-foreground">Avg Cost</span>
              <DollarSign className="h-4 w-4 text-muted-foreground" />
            </div>
            <div className="flex items-end gap-2">
              <span className="text-3xl font-bold">
                ${(analysis?.avg_cost_usd || 0).toFixed(4)}
              </span>
            </div>
            <div className="flex items-center gap-1 mt-2 text-xs">
              {analysis && analysis.avg_cost_usd < 0.05 ? (
                <ArrowUpRight className="h-3 w-3 text-green-500" />
              ) : (
                <ArrowDownRight className="h-3 w-3 text-orange-500" />
              )}
              <span className={analysis && analysis.avg_cost_usd < 0.05 ? 'text-green-600' : 'text-orange-600'}>
                {analysis && analysis.avg_cost_usd < 0.05 ? 'Efficient' : 'Expensive'}
              </span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Main Content */}
      <Tabs defaultValue="patterns" className="space-y-4">
        <TabsList>
          <TabsTrigger value="patterns" className="flex items-center gap-2">
            <PieChart className="h-4 w-4" />
            Detected Patterns ({analysis?.patterns.length || 0})
          </TabsTrigger>
          <TabsTrigger value="optimizations" className="flex items-center gap-2">
            <Lightbulb className="h-4 w-4" />
            Recommendations ({optimizations.length})
          </TabsTrigger>
          <TabsTrigger value="insights" className="flex items-center gap-2">
            <BarChart3 className="h-4 w-4" />
            Insights ({analysis?.insights.length || 0})
          </TabsTrigger>
        </TabsList>

        <TabsContent value="patterns" className="space-y-4">
          {!analysis?.patterns.length ? (
            <Card>
              <CardContent className="py-12 text-center">
                <CheckCircle className="h-12 w-12 mx-auto text-green-500 mb-4" />
                <p className="text-lg font-medium">No patterns detected</p>
                <p className="text-muted-foreground">Agent is performing within expected parameters</p>
              </CardContent>
            </Card>
          ) : (
            <div className="grid gap-4">
              {analysis.patterns.map((pattern) => {
                const Icon = patternIcons[pattern.pattern_type] || Activity;
                const colorClass = patternColors[pattern.pattern_type] || patternColors.optimal;
                
                return (
                  <Card key={pattern.id}>
                    <CardContent className="p-4">
                      <div className="flex items-start gap-4">
                        <div className={`p-3 rounded-lg ${colorClass.split(' ')[0]}`}>
                          <Icon className="h-5 w-5" />
                        </div>
                        
                        <div className="flex-1">
                          <div className="flex items-center gap-2 mb-1">
                            <Badge className={`capitalize ${colorClass}`}>
                              {pattern.pattern_type.replace(/_/g, ' ')}
                            </Badge>
                            <Badge variant="outline" className="text-xs">
                              {pattern.confidence.toFixed(0)}% confidence
                            </Badge>
                            <span className="text-xs text-muted-foreground">
                              {pattern.occurrence_count} occurrences
                            </span>
                          </div>
                          
                          <div className="space-y-2 mt-3">
                            <p className="text-sm font-medium">Recommendations:</p>
                            <ul className="space-y-1">
                              {pattern.recommendations.map((rec, i) => (
                                <li key={i} className="text-sm text-muted-foreground flex items-start gap-2">
                                  <ArrowUpRight className="h-3 w-3 mt-0.5 text-blue-500" />
                                  {rec}
                                </li>
                              ))}
                            </ul>
                          </div>
                          
                          <div className="flex items-center gap-4 mt-3 text-xs text-muted-foreground">
                            <span>First seen: {new Date(pattern.first_seen_at).toLocaleDateString()}</span>
                            <span>Last seen: {new Date(pattern.last_seen_at).toLocaleDateString()}</span>
                          </div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                );
              })}
            </div>
          )}
        </TabsContent>

        <TabsContent value="optimizations" className="space-y-4">
          {!optimizations.length ? (
            <Card>
              <CardContent className="py-12 text-center">
                <Lightbulb className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
                <p className="text-lg font-medium">No optimizations yet</p>
                <p className="text-muted-foreground mb-4">
                  Click "Optimize" to generate recommendations
                </p>
                <Button onClick={handleOptimize} disabled={optimizing}>
                  {optimizing ? <Loader2 className="h-4 w-4 animate-spin mr-2" /> : <Zap className="h-4 w-4 mr-2" />}
                  Generate Optimizations
                </Button>
              </CardContent>
            </Card>
          ) : (
            <div className="grid gap-4">
              {optimizations.map((opt) => (
                <Card key={opt.id}>
                  <CardContent className="p-4">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-2">
                          <Badge variant="outline" className="capitalize">
                            {optimizationTypes[opt.optimization_type] || opt.optimization_type.replace(/_/g, ' ')}
                          </Badge>
                          <Badge 
                            variant={opt.status === 'applied' ? 'default' : opt.status === 'approved' ? 'secondary' : 'outline'}
                            className="capitalize"
                          >
                            {opt.status}
                          </Badge>
                        </div>
                        
                        <p className="text-sm font-medium mb-2">{opt.description}</p>
                        
                        {Object.entries(opt.expected_impact).length > 0 && (
                          <div className="flex flex-wrap gap-2 mt-3">
                            {Object.entries(opt.expected_impact).map(([key, value]) => (
                              <div key={key} className="flex items-center gap-1 text-xs bg-muted px-2 py-1 rounded">
                                <span className="capitalize text-muted-foreground">{key.replace(/_/g, ' ')}:</span>
                                <span className="font-medium">{(value * 100).toFixed(0)}%</span>
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                      
                      <div className="flex flex-col items-end gap-2">
                        <div className="flex items-center gap-2">
                          <span className="text-xs text-muted-foreground">Implementation:</span>
                          <div className={`w-3 h-3 rounded-full ${implementationColors[opt.implementation]}`} />
                          <span className="text-xs capitalize">{opt.implementation}</span>
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="insights" className="space-y-4">
          {!analysis?.insights.length ? (
            <Card>
              <CardContent className="py-12 text-center text-muted-foreground">
                No insights available for this time period
              </CardContent>
            </Card>
          ) : (
            <div className="grid gap-3">
              {analysis.insights.map((insight, i) => (
                <div 
                  key={i} 
                  className="flex items-start gap-3 p-3 rounded-lg bg-muted/50"
                >
                  <BarChart3 className="h-4 w-4 mt-0.5 text-muted-foreground" />
                  <p className="text-sm">{insight}</p>
                </div>
              ))}
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default ExecutionPatternView;
