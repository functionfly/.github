import React, { useState } from 'react';
import { useParams } from 'react-router-dom';
import { cn } from '@/lib/utils';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { LoadingSpinner } from '@/components/ui/loading-spinner';
import {
  DNAHelix,
  EvolutionTimeline,
  DNAVariantDiff,
  DNAInsightsDashboard,
} from '@/components/dna';
import {
  useDNAProfile,
  useDNAMutations,
  useDNAMutation,
  useDNAInsights,
  useAcceptDNAVariant,
  useRejectDNAVariant,
  useTriggerDNAAnalysis,
  useToggleDNAEvolution,
} from '@/hooks/useFunctionDNA';
import { Link } from 'react-router-dom';
import { useAuthStore } from '@/stores/authStore';
import { Dna, GitBranch, BarChart3, Code2, ChevronRight, Sparkles } from 'lucide-react';

// ──────────────────────────────────────────────────────────────────────────────
// FunctionDNAPage — full-page DNA experience for a function
// ──────────────────────────────────────────────────────────────────────────────

export default function FunctionDNAPage() {
  const { id: functionId } = useParams<{ id: string }>();
  const [activeTab, setActiveTab] = useState('profile');
  const [selectedMutationId, setSelectedMutationId] = useState<string | null>(null);
  const [mutationFilter, setMutationFilter] = useState<string>('');

  // Data fetching
  const { data: profile, isLoading: profileLoading, error: profileError } = useDNAProfile(functionId || '');
  const { data: mutationsData, isLoading: mutationsLoading } = useDNAMutations(
    functionId || '',
    { status: mutationFilter || undefined, limit: 50 }
  );
  const { data: selectedMutation } = useDNAMutation(
    functionId || '',
    selectedMutationId || ''
  );
  const { data: insights, isLoading: insightsLoading } = useDNAInsights(functionId || '');

  // Mutations
  const acceptVariant = useAcceptDNAVariant(functionId || '');
  const rejectVariant = useRejectDNAVariant(functionId || '');
  const triggerAnalysis = useTriggerDNAAnalysis(functionId || '');
  const toggleEvolution = useToggleDNAEvolution(functionId || '');

  const username = useAuthStore((s) => s.user?.username) ?? '';

  if (!functionId) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-text-muted">No function selected</p>
      </div>
    );
  }

  if (profileLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner />
      </div>
    );
  }

  if (profileError) {
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
            <h3 className="text-lg font-semibold text-text-primary mb-2">Function Not Found</h3>
            <p className="text-sm text-text-secondary mb-4 text-center max-w-md">
              Could not load DNA profile for this function. The function may not exist or you may not have access.
            </p>
            <Button variant="outline" onClick={() => window.history.back()}>
              Go Back
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!profile) {
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
            <h3 className="text-lg font-semibold text-text-primary mb-2">DNA Not Enabled</h3>
            <p className="text-sm text-text-secondary mb-6 text-center max-w-md">
              Enable Function DNA from your Platform Settings to track execution patterns and receive AI-powered evolution suggestions.
            </p>
            <Link to={`/u/${username}/settings`}>
              <Button variant="outline" className="gap-1.5">
                <Sparkles className="h-4 w-4" />
                Enable in Platform Settings
                <ChevronRight className="h-4 w-4 opacity-60" />
              </Button>
            </Link>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6 max-w-6xl mx-auto">
      {/* Page header */}
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
        {profile && (
          <Badge variant="outline" className="ml-auto font-mono">
            {profile.function_type} / {profile.function_id}
          </Badge>
        )}
      </div>

      {/* Main content */}
      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="profile" className="gap-1.5">
            <Dna className="h-3.5 w-3.5" />
            DNA Profile
          </TabsTrigger>
          <TabsTrigger value="evolution" className="gap-1.5">
            <GitBranch className="h-3.5 w-3.5" />
            Evolution History
            {mutationsData && mutationsData.total > 0 && (
              <span className="ml-1 px-1.5 py-0.5 rounded-full bg-velocity-500/10 text-velocity-500 text-[10px] font-mono">
                {mutationsData.total}
              </span>
            )}
          </TabsTrigger>
          <TabsTrigger value="insights" className="gap-1.5">
            <BarChart3 className="h-3.5 w-3.5" />
            Insights
          </TabsTrigger>
        </TabsList>

        {/* DNA Profile Tab */}
        <TabsContent value="profile" className="space-y-6">
          {profile && (
            <DNAHelix
              profile={profile}
              onToggleEvolution={(enabled) => toggleEvolution.mutate(enabled)}
              onTriggerAnalysis={() => triggerAnalysis.mutate()}
              isToggling={toggleEvolution.isPending}
              isAnalyzing={triggerAnalysis.isPending}
            />
          )}

          {/* Pending mutation preview */}
          {mutationsData?.mutations?.some((m) => m.status === 'proposed') && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2 text-base">
                  <Code2 className="h-4 w-4 text-velocity-500" />
                  Pending Evolution
                  <Badge variant="default" className="ml-1">New</Badge>
                </CardTitle>
              </CardHeader>
              <CardContent>
                {mutationsData.mutations
                  .filter((m) => m.status === 'proposed')
                  .slice(0, 1)
                  .map((m) => (
                    <div key={m.id} className="space-y-3">
                      <p className="text-sm text-text-secondary">{m.trigger_reason}</p>
                      <button
                        onClick={() => {
                          setSelectedMutationId(m.id);
                          setActiveTab('evolution');
                        }}
                        className="text-sm text-velocity-500 hover:text-velocity-400 transition-colors"
                      >
                        Review and decide →
                      </button>
                    </div>
                  ))}
              </CardContent>
            </Card>
          )}
        </TabsContent>

        {/* Evolution History Tab */}
        <TabsContent value="evolution" className="space-y-4">
          {/* Filters */}
          <div className="flex items-center gap-2 flex-wrap">
            {['', 'proposed', 'accepted', 'rejected', 'deployed'].map((status) => (
              <button
                key={status}
                onClick={() => setMutationFilter(status)}
                className={cn(
                  'px-3 py-1 text-xs font-medium rounded-lg transition-all',
                  'border',
                  mutationFilter === status
                    ? 'border-velocity-500/30 bg-velocity-500/10 text-velocity-500'
                    : 'border-border-subtle bg-bg-tertiary text-text-muted hover:text-text-secondary'
                )}
              >
                {status || 'All'}
              </button>
            ))}
          </div>

          {selectedMutation ? (
            <div className="space-y-4">
              <button
                onClick={() => setSelectedMutationId(null)}
                className="text-sm text-text-muted hover:text-text-primary transition-colors"
              >
                ← Back to timeline
              </button>
              <DNAVariantDiff
                mutation={selectedMutation}
                onAccept={(canaryPct) =>
                  acceptVariant.mutate({ mutationId: selectedMutation.id, canaryPercentage: canaryPct })
                }
                onReject={(reason) =>
                  rejectVariant.mutate({ mutationId: selectedMutation.id, reason })
                }
                isAccepting={acceptVariant.isPending}
                isRejecting={rejectVariant.isPending}
              />
            </div>
          ) : (
            <EvolutionTimeline
              mutations={mutationsData?.mutations || []}
              loading={mutationsLoading}
              onViewDiff={(mutationId) => setSelectedMutationId(mutationId)}
            />
          )}
        </TabsContent>

        {/* Insights Tab */}
        <TabsContent value="insights">
          <DNAInsightsDashboard
            functionInsights={insights}
            loading={insightsLoading}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
