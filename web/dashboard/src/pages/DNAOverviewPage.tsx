import React from 'react';
import { Dna, BarChart3, GitBranch, Sparkles, ChevronRight, DollarSign, TrendingUp } from 'lucide-react';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { LoadingSpinner } from '@/components/ui/loading-spinner';
import { useEnterpriseDNAInsights } from '@/hooks/useFunctionDNA';
import { Link } from 'react-router-dom';

export default function DNAOverviewPage() {
  const { data: insights, isLoading, error } = useEnterpriseDNAInsights('30d');

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner />
      </div>
    );
  }

  if (error || !insights) {
    return (
      <div className="space-y-6 max-w-6xl mx-auto">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-xl bg-velocity-500/10">
            <Dna className="h-6 w-6 text-velocity-500" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-text-primary font-display">Function DNA</h1>
            <p className="text-sm text-text-secondary">
              Living code that evolves based on real production traffic
            </p>
          </div>
        </div>
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-16">
            <Dna className="h-12 w-12 text-text-muted mb-4" />
            <h3 className="text-lg font-semibold text-text-primary mb-2">DNA Not Available</h3>
            <p className="text-sm text-text-secondary mb-4 text-center max-w-md">
              Unable to load DNA insights. Please ensure you have functions with DNA enabled.
            </p>
            <Link to="/functions/my">
              <Button variant="outline" className="gap-1.5">
                Go to Functions
                <ChevronRight className="h-4 w-4 opacity-60" />
              </Button>
            </Link>
          </CardContent>
        </Card>
      </div>
    );
  }

  const totalFunctions = insights.total_functions_analyzed || 0;
  const totalMutations = insights.total_mutations_proposed || 0;
  const acceptedMutations = insights.total_mutations_accepted || 0;
  const avgFitness = insights.avg_fitness_score || 0;
  const costSavings = insights.total_cost_savings_usd || 0;
  const latencyImprovement = insights.avg_latency_improvement_pct || 0;

  return (
    <div className="space-y-6 max-w-6xl mx-auto">
      <div className="flex items-center gap-3">
        <div className="p-2 rounded-xl bg-velocity-500/10">
          <Dna className="h-6 w-6 text-velocity-500" />
        </div>
        <div>
          <h1 className="text-2xl font-bold text-text-primary font-display">Function DNA</h1>
          <p className="text-sm text-text-secondary">
            Living code that evolves based on real production traffic
          </p>
        </div>
        <Badge variant="outline" className="ml-auto">
          Enterprise Overview
        </Badge>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">Functions Analyzed</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-text-primary">{totalFunctions}</div>
            <p className="text-xs text-text-muted mt-1">with DNA profiles</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">Mutations Proposed</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-text-primary">{totalMutations}</div>
            <p className="text-xs text-text-muted mt-1">
              <span className="text-velocity-500">{acceptedMutations} accepted</span>
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">Avg Fitness Score</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-text-primary">{avgFitness.toFixed(1)}%</div>
            <p className="text-xs text-text-muted mt-1">across all functions</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">Latency Improvement</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-velocity-500">+{latencyImprovement.toFixed(1)}%</div>
            <p className="text-xs text-text-muted mt-1">average improvement</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">Cost Savings</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-velocity-500">
              <DollarSign className="inline h-5 w-5" />
              {costSavings.toFixed(0)}
            </div>
            <p className="text-xs text-text-muted mt-1">USD saved via optimizations</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-text-secondary">Top Categories</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-1">
              {insights.top_bottleneck_categories?.slice(0, 3).map((cat, i) => (
                <div key={i} className="flex items-center justify-between text-sm">
                  <span className="text-text-secondary truncate">{cat.category}</span>
                  <Badge variant="outline" className="text-xs">{cat.count}</Badge>
                </div>
              )) || <span className="text-text-muted text-sm">No data yet</span>}
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Sparkles className="h-4 w-4 text-velocity-500" />
            Quick Actions
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between p-4 rounded-lg border border-border-subtle bg-card hover:border-velocity-500/30 transition-colors">
            <div className="flex items-center gap-3">
              <Dna className="h-5 w-5 text-velocity-500" />
              <div>
                <p className="font-medium text-text-primary">Browse Functions</p>
                <p className="text-sm text-text-secondary">View functions with DNA profiles</p>
              </div>
            </div>
            <Link to="/functions/my">
              <Button variant="outline" size="sm" className="gap-1.5">
                Go to Functions
                <ChevronRight className="h-4 w-4" />
              </Button>
            </Link>
          </div>

          <div className="flex items-center justify-between p-4 rounded-lg border border-border-subtle bg-card hover:border-velocity-500/30 transition-colors">
            <div className="flex items-center gap-3">
              <TrendingUp className="h-5 w-5 text-velocity-500" />
              <div>
                <p className="font-medium text-text-primary">Evolution Leaderboard</p>
                <p className="text-sm text-text-secondary">See top performing functions</p>
              </div>
            </div>
            <Button variant="outline" size="sm" className="gap-1.5" disabled>
              Coming Soon
            </Button>
          </div>

          <div className="flex items-center justify-between p-4 rounded-lg border border-border-subtle bg-card hover:border-velocity-500/30 transition-colors">
            <div className="flex items-center gap-3">
              <GitBranch className="h-5 w-5 text-velocity-500" />
              <div>
                <p className="font-medium text-text-primary">Evolution History</p>
                <p className="text-sm text-text-secondary">Track all mutation events</p>
              </div>
            </div>
            <Link to="/evolution">
              <Button variant="outline" size="sm" className="gap-1.5">
                View Evolution
                <ChevronRight className="h-4 w-4" />
              </Button>
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}