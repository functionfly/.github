import { motion } from 'framer-motion';
import {
  ScanSearch,
  CheckSquare,
  Square,
  AlertTriangle,
  Clock,
  DollarSign,
  Zap,
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';
import { DetectedFunctionCard } from '@/components/github/DetectedFunctionCard';
import { useGitHubStore } from '@/stores/githubStore';
import type { ScanResult } from '@/types/github';

function getConfidenceColor(confidence: number): string {
  if (confidence >= 80) return 'text-emerald-500';
  if (confidence >= 50) return 'text-amber-500';
  return 'text-red-500';
}

function getConfidenceBarColor(confidence: number): string {
  if (confidence >= 80) return 'bg-emerald-500';
  if (confidence >= 50) return 'bg-amber-500';
  return 'bg-red-500';
}

interface FunctionDetectionResultsProps {
  scanResult: ScanResult;
  className?: string;
}

export function FunctionDetectionResults({ scanResult, className }: FunctionDetectionResultsProps) {
  const {
    importConfig,
    toggleFunction,
    setSelectedFunctions,
    setVisibilityOverride,
  } = useGitHubStore();

  const { selectedFunctions, visibilityOverrides, globalVisibility } = importConfig;

  const allSelected = scanResult.functions.every((f) => selectedFunctions.includes(f.name));
  const noneSelected = selectedFunctions.length === 0;

  const handleSelectAll = () => {
    setSelectedFunctions(scanResult.functions.map((f) => f.name));
  };

  const handleDeselectAll = () => {
    setSelectedFunctions([]);
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className={className}
    >
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between flex-wrap gap-3">
            <div className="flex items-center gap-3">
              <div className="h-9 w-9 rounded-lg bg-brand-500/10 flex items-center justify-center">
                <ScanSearch className="h-5 w-5 text-brand-500" />
              </div>
              <div>
                <CardTitle className="text-base">Scan Results</CardTitle>
                <p className="text-xs text-text-muted mt-0.5">
                  {scanResult.functions.length} function{scanResult.functions.length !== 1 ? 's' : ''} detected
                </p>
              </div>
            </div>

            <div className="flex items-center gap-3">
              <div className="flex items-center gap-2">
                <span className="text-xs text-text-muted">Overall confidence:</span>
                <span className={cn('text-sm font-bold', getConfidenceColor(scanResult.overall_confidence))}>
                  {scanResult.overall_confidence}%
                </span>
              </div>
              <Badge variant="outline" className="text-xs font-mono">
                <Zap className="h-3 w-3 mr-1" />
                {scanResult.strategy_used}
              </Badge>
            </div>
          </div>
        </CardHeader>

        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                size="sm"
                onClick={handleSelectAll}
                disabled={allSelected}
                aria-label="Select all functions"
              >
                <CheckSquare className="h-3.5 w-3.5 mr-1.5" />
                Select All
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={handleDeselectAll}
                disabled={noneSelected}
                aria-label="Deselect all functions"
              >
                <Square className="h-3.5 w-3.5 mr-1.5" />
                Deselect All
              </Button>
            </div>
            <span className="text-xs text-text-muted">
              {selectedFunctions.length} of {scanResult.functions.length} selected
            </span>
          </div>

          <ScrollArea className="max-h-[400px]">
            <div className="space-y-2 pr-2">
              {scanResult.functions.map((fn, index) => (
                <motion.div
                  key={fn.name}
                  initial={{ opacity: 0, x: -10 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ duration: 0.2, delay: index * 0.05 }}
                >
                  <DetectedFunctionCard
                    fn={fn}
                    selected={selectedFunctions.includes(fn.name)}
                    visibility={visibilityOverrides[fn.name] ?? globalVisibility}
                    onSelect={(selected) => toggleFunction(fn.name)}
                    onNameChange={(name) => {
                      // Name change handled at parent level via store
                    }}
                    onVisibilityChange={(vis) => setVisibilityOverride(fn.name, vis)}
                  />
                </motion.div>
              ))}
            </div>
          </ScrollArea>

          {scanResult.warnings.length > 0 && (
            <div className="space-y-2 pt-2 border-t border-border-subtle">
              <div className="flex items-center gap-2 text-xs text-amber-500">
                <AlertTriangle className="h-3.5 w-3.5" />
                <span className="font-medium">Warnings</span>
              </div>
              <div className="flex flex-wrap gap-1.5">
                {scanResult.warnings.map((warning, i) => (
                  <Badge key={i} variant="warning" className="text-[10px]">
                    {warning}
                  </Badge>
                ))}
              </div>
            </div>
          )}

          <div className="flex items-center gap-4 pt-2 border-t border-border-subtle text-xs text-text-muted">
            <div className="flex items-center gap-1.5">
              <Clock className="h-3 w-3" />
              <span>Est. time: {scanResult.estimated_import_time_seconds}s</span>
            </div>
            <div className="flex items-center gap-1.5">
              <DollarSign className="h-3 w-3" />
              <span>Est. cost: ${scanResult.estimated_cost_usd.toFixed(4)}</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}
