/**
 * OutputComparisonViewer - Side-by-side diff for test results
 */

import { useState } from 'react';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  CheckCircle2,
  XCircle,
  AlertTriangle,
  ChevronDown,
  ChevronUp,
  Copy,
} from 'lucide-react';
import type { TestResult } from '../types';

interface OutputComparisonViewerProps {
  testResult: TestResult;
  expanded?: boolean;
  className?: string;
}

export function OutputComparisonViewer({
  testResult,
  expanded: defaultExpanded = false,
  className,
}: OutputComparisonViewerProps) {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);
  const [copiedField, setCopiedField] = useState<string | null>(null);

  const handleCopy = async (text: string, field: string) => {
    await navigator.clipboard.writeText(text);
    setCopiedField(field);
    setTimeout(() => setCopiedField(null), 2000);
  };

  const statusConfig = {
    passed: {
      icon: CheckCircle2,
      color: 'text-emerald-400',
      bgColor: 'bg-emerald-500/10',
      borderColor: 'border-emerald-500/30',
      label: 'Passed',
    },
    failed: {
      icon: XCircle,
      color: 'text-red-400',
      bgColor: 'bg-red-500/10',
      borderColor: 'border-red-500/30',
      label: 'Failed',
    },
    error: {
      icon: AlertTriangle,
      color: 'text-amber-400',
      bgColor: 'bg-amber-500/10',
      borderColor: 'border-amber-500/30',
      label: 'Error',
    },
  };

  const config = statusConfig[testResult.status];
  const StatusIcon = config.icon;

  return (
    <div
      className={cn(
        'rounded-lg border bg-slate-900/50 overflow-hidden',
        config.borderColor,
        className
      )}
    >
      {/* Header */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="flex w-full items-center justify-between p-3 hover:bg-slate-800/50 transition-colors"
      >
        <div className="flex items-center gap-2">
          <StatusIcon className={cn('h-4 w-4', config.color)} />
          <span className="font-medium text-slate-200">{testResult.testName}</span>
          <Badge variant="outline" className={cn('text-xs', config.bgColor, config.color, config.borderColor)}>
            {config.label}
          </Badge>
        </div>
        <div className="flex items-center gap-2">
          {testResult.executionTimeMs && (
            <span className="text-xs text-slate-400">
              {testResult.executionTimeMs.toFixed(2)}ms
            </span>
          )}
          {isExpanded ? (
            <ChevronUp className="h-4 w-4 text-slate-400" />
          ) : (
            <ChevronDown className="h-4 w-4 text-slate-400" />
          )}
        </div>
      </button>

      {/* Content */}
      {isExpanded && (
        <div className="border-t border-slate-800">
          {testResult.errorMessage ? (
            <div className="p-3">
              <p className="text-xs font-medium text-red-400 mb-1">Error:</p>
              <pre className="rounded bg-red-950/30 border border-red-500/20 p-3 text-xs text-red-300 overflow-x-auto">
                {testResult.errorMessage}
              </pre>
            </div>
          ) : (
            <div className="grid sm:grid-cols-2 gap-px bg-slate-800">
              {/* Input */}
              {testResult.input && (
                <div className="bg-slate-900/50 p-3">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-xs font-medium text-slate-400">Input</span>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 px-2 text-xs text-slate-400 hover:text-slate-300"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleCopy(testResult.input || '', 'input');
                      }}
                    >
                      {copiedField === 'input' ? 'Copied!' : <Copy className="h-3 w-3" />}
                    </Button>
                  </div>
                  <pre className="text-xs text-slate-300 overflow-x-auto whitespace-pre-wrap">
                    {testResult.input}
                  </pre>
                </div>
              )}

              {/* Expected Output */}
              {testResult.expectedOutput && (
                <div className="bg-slate-900/50 p-3">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-xs font-medium text-slate-400">Expected Output</span>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 px-2 text-xs text-slate-400 hover:text-slate-300"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleCopy(testResult.expectedOutput || '', 'expected');
                      }}
                    >
                      {copiedField === 'expected' ? 'Copied!' : <Copy className="h-3 w-3" />}
                    </Button>
                  </div>
                  <pre className="text-xs text-emerald-300 overflow-x-auto whitespace-pre-wrap">
                    {testResult.expectedOutput}
                  </pre>
                </div>
              )}

              {/* Actual Output */}
              {testResult.actualOutput && (
                <div className={cn(
                  'bg-slate-900/50 p-3',
                  testResult.status === 'failed' && 'sm:col-span-2'
                )}>
                  <div className="flex items-center justify-between mb-2">
                    <span className={cn(
                      'text-xs font-medium',
                      testResult.status === 'passed' ? 'text-emerald-400' : 'text-red-400'
                    )}>
                      Actual Output
                    </span>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 px-2 text-xs text-slate-400 hover:text-slate-300"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleCopy(testResult.actualOutput || '', 'actual');
                      }}
                    >
                      {copiedField === 'actual' ? 'Copied!' : <Copy className="h-3 w-3" />}
                    </Button>
                  </div>
                  <pre className={cn(
                    'text-xs overflow-x-auto whitespace-pre-wrap',
                    testResult.status === 'passed' ? 'text-emerald-300' : 'text-red-300'
                  )}>
                    {testResult.actualOutput}
                  </pre>
                </div>
              )}
            </div>
          )}

          {/* Match Info */}
          {testResult.matchScore !== undefined && (
            <div className="border-t border-slate-800 px-3 py-2">
              <div className="flex items-center justify-between text-xs">
                <span className="text-slate-400">Match Score:</span>
                <span className={cn(
                  'font-medium',
                  testResult.matchScore >= 90 ? 'text-emerald-400' :
                  testResult.matchScore >= 70 ? 'text-amber-400' : 'text-red-400'
                )}>
                  {testResult.matchScore.toFixed(1)}%
                </span>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

/**
 * Summary of all test results
 */
interface TestResultsSummaryProps {
  results: TestResult[];
  className?: string;
}

export function TestResultsSummary({ results, className }: TestResultsSummaryProps) {
  const passed = results.filter(r => r.status === 'passed').length;
  const failed = results.filter(r => r.status === 'failed').length;
  const errors = results.filter(r => r.status === 'error').length;
  const total = results.length;

  return (
    <div className={cn('flex items-center gap-4 text-sm', className)}>
      <span className="flex items-center gap-1.5">
        <span className="h-2 w-2 rounded-full bg-emerald-400" />
        <span className="text-slate-400">{passed} passed</span>
      </span>
      {failed > 0 && (
        <span className="flex items-center gap-1.5">
          <span className="h-2 w-2 rounded-full bg-red-400" />
          <span className="text-slate-400">{failed} failed</span>
        </span>
      )}
      {errors > 0 && (
        <span className="flex items-center gap-1.5">
          <span className="h-2 w-2 rounded-full bg-amber-400" />
          <span className="text-slate-400">{errors} errors</span>
        </span>
      )}
      <span className="text-slate-600">•</span>
      <span className="text-slate-400">{total} total</span>
    </div>
  );
}
