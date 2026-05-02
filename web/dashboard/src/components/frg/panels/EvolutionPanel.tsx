/**
 * EvolutionPanel - Agent Evolution Mode UI
 * 
 * Displays optimization suggestions from the SAR runtime's GraphOptimizer,
 * allows users to approve/reject suggestions, and shows evolution history.
 * 
 * Phase 1: Agent Evolution Mode feature
 */

import { useEffect, useState } from 'react';
import {
  Sparkles,
  Check,
  X,
  History,
  RefreshCw,
  ChevronDown,
  ChevronUp,
  Lightbulb,
  AlertCircle,
  Zap,
} from 'lucide-react';
import { useFRGStore, type EvolutionSuggestion } from '@/stores/frgStore';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { Separator } from '@/components/ui/separator';
import { cn } from '@/lib/utils';

interface SuggestionCardProps {
  suggestion: EvolutionSuggestion;
  onApprove: (id: string) => void;
  onReject: (id: string) => void;
}

function SuggestionCard({ suggestion, onApprove, onReject }: SuggestionCardProps) {
  const [expanded, setExpanded] = useState(false);

  const getConfidenceVariant = (confidence: string) => {
    switch (confidence) {
      case 'high': return 'success' as const;
      case 'medium': return 'warning' as const;
      case 'low': return 'destructive' as const;
      default: return 'secondary' as const;
    }
  };

  const getImpactDisplay = (impact: number) => {
    if (impact >= 0.2) return { label: 'High Impact', variant: 'success' as const };
    if (impact >= 0.1) return { label: 'Medium Impact', variant: 'warning' as const };
    return { label: 'Low Impact', variant: 'secondary' as const };
  };

  const impact = getImpactDisplay(suggestion.expectedImpact);

  return (
    <Card className="mb-3 overflow-hidden">
      <CardContent className="p-4">
        <div className="flex items-center gap-2 mb-2">
          <Lightbulb className="h-4 w-4 text-brand-500" />
          <span className="text-sm font-medium flex-1">{suggestion.type}</span>
          <Badge variant={getConfidenceVariant(suggestion.confidence)}>
            {suggestion.confidence}
          </Badge>
          <Badge variant={impact.variant} className="border border-border-default">
            {impact.label}
          </Badge>
        </div>

        <p className="text-sm text-text-secondary mb-3">
          {suggestion.description}
        </p>

        <div className="flex items-center justify-between">
          <span className="text-xs text-text-muted">
            {new Date(suggestion.createdAt).toLocaleDateString()}
          </span>
          
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => setExpanded(!expanded)}
            >
              {expanded ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
            </Button>
            
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-success hover:text-success hover:bg-success-glow"
                    onClick={() => onApprove(suggestion.id)}
                  >
                    <Check className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Approve and apply</TooltipContent>
              </Tooltip>
            </TooltipProvider>
            
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-error hover:text-error hover:bg-error-glow"
                    onClick={() => onReject(suggestion.id)}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>Reject</TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        </div>

        {expanded && (
          <>
            <Separator className="my-3" />
            <div className="bg-bg-secondary rounded-lg p-3">
              <span className="text-xs font-medium block mb-2">Suggestion Data:</span>
              <pre className="text-xs bg-bg-primary p-2 rounded overflow-auto max-h-[200px]">
                {JSON.stringify(suggestion.data, null, 2)}
              </pre>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
}

interface HistoryEntryProps {
  entry: EvolutionSuggestion;
}

function HistoryEntry({ entry }: HistoryEntryProps) {
  const getStatusVariant = (status: string) => {
    switch (status) {
      case 'implemented': return 'success' as const;
      case 'approved': return 'default' as const;
      case 'rejected': return 'destructive' as const;
      default: return 'secondary' as const;
    }
  };

  return (
    <div className="flex items-start gap-3 py-3 border-b border-border-default last:border-0">
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium">{entry.type}</p>
        <p className="text-xs text-text-secondary truncate">
          {entry.description.substring(0, 60)}...
        </p>
        <p className="text-xs text-text-muted mt-1">
          {entry.implementedAt 
            ? `Implemented: ${new Date(entry.implementedAt).toLocaleDateString()}`
            : new Date(entry.createdAt).toLocaleDateString()
          }
        </p>
      </div>
      <Badge variant={getStatusVariant(entry.status)}>
        {entry.status}
      </Badge>
    </div>
  );
}

export function EvolutionPanel() {
  const {
    evolutionStatus,
    evolutionSuggestions,
    evolutionHistory,
    isEvolutionLoading,
    evolutionError,
    toggleEvolutionMode,
    fetchEvolutionStatus,
    fetchEvolutionSuggestions,
    fetchEvolutionHistory,
    triggerEvolutionAnalysis,
    approveEvolutionSuggestion,
    rejectEvolutionSuggestion,
  } = useFRGStore();

  const [activeTab, setActiveTab] = useState<'suggestions' | 'history'>('suggestions');
  const [showEnableHint, setShowEnableHint] = useState(true);

  useEffect(() => {
    // Fetch evolution data on mount
    fetchEvolutionStatus();
    fetchEvolutionSuggestions();
    fetchEvolutionHistory();
  }, [fetchEvolutionStatus, fetchEvolutionSuggestions, fetchEvolutionHistory]);

  const handleToggleEvolution = async () => {
    const newEnabled = !evolutionStatus?.evolutionEnabled;
    await toggleEvolutionMode(newEnabled);
    if (newEnabled) {
      setShowEnableHint(false);
    }
  };

  const handleApprove = async (id: string) => {
    approveEvolutionSuggestion(id);
    // Call API to actually approve
    try {
      const response = await fetch(`/api/agents/${evolutionStatus?.agentId}/evolution/suggestions/${id}/approve`, {
        method: 'POST',
      });
      if (!response.ok) throw new Error('Failed to approve');
      // Refresh suggestions
      fetchEvolutionSuggestions();
      fetchEvolutionHistory();
    } catch (err) {
      console.error('Failed to approve suggestion:', err);
    }
  };

  const handleReject = async (id: string) => {
    rejectEvolutionSuggestion(id);
    // Call API to actually reject
    try {
      const response = await fetch(`/api/agents/${evolutionStatus?.agentId}/evolution/suggestions/${id}/reject`, {
        method: 'POST',
      });
      if (!response.ok) throw new Error('Failed to reject');
      // Refresh suggestions
      fetchEvolutionSuggestions();
    } catch (err) {
      console.error('Failed to reject suggestion:', err);
    }
  };

  const pendingSuggestions = evolutionSuggestions.filter(s => s.status === 'pending');

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="p-4 border-b border-border-default">
        <div className="flex items-center gap-2 mb-4">
          <Sparkles className="h-5 w-5" />
          <h3 className="text-lg font-semibold flex-1">Agent Evolution</h3>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  onClick={() => {
                    fetchEvolutionStatus();
                    fetchEvolutionSuggestions();
                  }}
                  disabled={isEvolutionLoading}
                >
                  <RefreshCw className={cn("h-4 w-4", isEvolutionLoading && "animate-spin")} />
                </Button>
              </TooltipTrigger>
              <TooltipContent>Refresh</TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>

        {/* Evolution Mode Toggle */}
        <Card className="mb-4 border-border-subtle">
          <CardContent className="p-3 flex items-center justify-between">
            <div>
              <p className="text-sm font-medium">Evolution Mode</p>
              <p className="text-xs text-text-secondary">Let the agent improve itself</p>
            </div>
            <Switch
              checked={evolutionStatus?.evolutionEnabled || false}
              onCheckedChange={handleToggleEvolution}
              disabled={isEvolutionLoading}
            />
          </CardContent>
        </Card>

        {/* Stats */}
        {evolutionStatus && (
          <div className="flex gap-2">
            <Badge 
              variant={activeTab === 'suggestions' ? 'default' : 'secondary'}
              className="cursor-pointer"
              onClick={() => setActiveTab('suggestions')}
            >
              <Lightbulb className="h-3 w-3 mr-1" />
              {evolutionStatus.pendingCount} Pending
            </Badge>
            <Badge 
              variant={activeTab === 'history' ? 'default' : 'secondary'}
              className="cursor-pointer"
              onClick={() => setActiveTab('history')}
            >
              <History className="h-3 w-3 mr-1" />
              {evolutionStatus.implementedCount} Applied
            </Badge>
          </div>
        )}
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto p-4">
        {isEvolutionLoading && (
          <div className="flex justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-brand-500" />
          </div>
        )}

        {evolutionError && (
          <Alert variant="destructive" className="mb-4">
            <AlertDescription>{evolutionError}</AlertDescription>
          </Alert>
        )}

        {!evolutionStatus?.evolutionEnabled && showEnableHint && (
          <Alert className="mb-4">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription className="flex items-center justify-between">
              <span>Enable Evolution Mode to receive AI-powered optimization suggestions based on your agent's performance.</span>
              <Button 
                variant="ghost" 
                size="icon" 
                className="h-6 w-6 -mr-2"
                onClick={() => setShowEnableHint(false)}
              >
                <X className="h-3 w-3" />
              </Button>
            </AlertDescription>
          </Alert>
        )}

        {activeTab === 'suggestions' && (
          <>
            {/* Manual Analysis Trigger */}
            <Button
              variant="outline"
              className="w-full mb-4"
              onClick={triggerEvolutionAnalysis}
              disabled={isEvolutionLoading || !evolutionStatus?.evolutionEnabled}
            >
              <Sparkles className="h-4 w-4 mr-2" />
              Analyze Performance Now
            </Button>

            {pendingSuggestions.length === 0 ? (
              <div className="text-center py-8 text-text-secondary">
                <Lightbulb className="h-12 w-12 mx-auto mb-2 opacity-30" />
                <p className="text-sm">No pending suggestions</p>
                <p className="text-xs mt-1">
                  Suggestions appear when the optimizer detects patterns
                </p>
              </div>
            ) : (
              pendingSuggestions.map((suggestion) => (
                <SuggestionCard
                  key={suggestion.id}
                  suggestion={suggestion}
                  onApprove={handleApprove}
                  onReject={handleReject}
                />
              ))
            )}
          </>
        )}

        {activeTab === 'history' && (
          <>
            {evolutionHistory.length === 0 ? (
              <div className="text-center py-8 text-text-secondary">
                <History className="h-12 w-12 mx-auto mb-2 opacity-30" />
                <p className="text-sm">No evolution history yet</p>
                <p className="text-xs mt-1">
                  Applied suggestions will appear here
                </p>
              </div>
            ) : (
              <div className="space-y-1">
                {evolutionHistory.map((entry) => (
                  <HistoryEntry key={entry.id} entry={entry} />
                ))}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
}
