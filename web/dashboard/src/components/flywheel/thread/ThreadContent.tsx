/**
 * ThreadContent - Markdown + code rendering for thread content
 */

import { useState } from 'react';
import { cn } from '@/lib/utils';
import { Badge } from '@/components/ui/badge';
import {
  Lock,
  Unlock,
  Clock,
  Cpu,
  HardDrive,
  Globe,
  AlertCircle,
} from 'lucide-react';
import type { ProblemData, EnvironmentSpecs, TestCase } from '../types';

interface ThreadContentProps {
  problemData?: ProblemData;
  environmentSpecs?: EnvironmentSpecs;
  className?: string;
}

function TestCaseCard({ testCase, index }: { testCase: TestCase; index: number }) {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <div className="rounded-lg border border-slate-800 bg-slate-900/50">
      <button
        onClick={() => testCase.isPublic && setIsExpanded(!isExpanded)}
        className={cn(
          'flex w-full items-center justify-between p-3 text-left',
          testCase.isPublic && 'hover:bg-slate-800/50'
        )}
      >
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-slate-300">
            Test #{index + 1}: {testCase.name}
          </span>
          {testCase.isPublic ? (
            <Badge variant="outline" className="border-emerald-500/30 bg-emerald-500/10 text-emerald-400 text-xs">
              <Unlock className="mr-1 h-3 w-3" />
              Public
            </Badge>
          ) : (
            <Badge variant="outline" className="border-amber-500/30 bg-amber-500/10 text-amber-400 text-xs">
              <Lock className="mr-1 h-3 w-3" />
              Hidden
            </Badge>
          )}
        </div>
        {testCase.isPublic && (
          <span className="text-xs text-slate-500">
            {isExpanded ? 'Hide' : 'Show'}
          </span>
        )}
      </button>

      {isExpanded && testCase.isPublic && (
        <div className="border-t border-slate-800 p-3 space-y-2">
          {testCase.description && (
            <p className="text-sm text-slate-400">{testCase.description}</p>
          )}
          <div className="grid gap-2 sm:grid-cols-2">
            <div>
              <p className="text-xs font-medium text-slate-500 mb-1">Input:</p>
              <pre className="rounded bg-slate-950 p-2 text-xs text-slate-300 overflow-x-auto">
                {testCase.input}
              </pre>
            </div>
            <div>
              <p className="text-xs font-medium text-slate-500 mb-1">Expected Output:</p>
              <pre className="rounded bg-slate-950 p-2 text-xs text-slate-300 overflow-x-auto">
                {testCase.expectedOutput}
              </pre>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export function ThreadContent({
  problemData,
  environmentSpecs,
  className,
}: ThreadContentProps) {
  if (!problemData && !environmentSpecs) {
    return null;
  }

  return (
    <div className={cn('space-y-6', className)}>
      {/* Problem Description */}
      {problemData?.description && (
        <div className="prose prose-invert max-w-none">
          <h2 className="text-xl font-semibold text-white mb-3">Problem Description</h2>
          <div className="text-slate-300 whitespace-pre-wrap">
            {problemData.description}
          </div>
        </div>
      )}

      {/* Constraints */}
      {problemData?.constraints && (
        <div className="rounded-lg border border-slate-800 bg-slate-900/50 p-4">
          <h3 className="flex items-center gap-2 text-sm font-semibold text-white mb-3">
            <AlertCircle className="h-4 w-4 text-amber-400" />
            Constraints
          </h3>
          <div className="grid gap-3 sm:grid-cols-2">
            {problemData.constraints.timeComplexity && (
              <div className="flex items-center gap-2">
                <Clock className="h-4 w-4 text-slate-500" />
                <span className="text-sm text-slate-400">Time Complexity:</span>
                <Badge variant="outline" className="border-slate-700 text-slate-300">
                  {problemData.constraints.timeComplexity}
                </Badge>
              </div>
            )}
            {problemData.constraints.spaceComplexity && (
              <div className="flex items-center gap-2">
                <HardDrive className="h-4 w-4 text-slate-500" />
                <span className="text-sm text-slate-400">Space Complexity:</span>
                <Badge variant="outline" className="border-slate-700 text-slate-300">
                  {problemData.constraints.spaceComplexity}
                </Badge>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Test Cases */}
      {problemData?.testCases && problemData.testCases.length > 0 && (
        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-white">Test Cases</h3>
          <div className="space-y-2">
            {problemData.testCases.map((testCase, index) => (
              <TestCaseCard key={testCase.id} testCase={testCase} index={index} />
            ))}
          </div>
        </div>
      )}

      {/* Environment Specs */}
      {environmentSpecs && (
        <div className="rounded-lg border border-slate-800 bg-slate-900/50 p-4">
          <h3 className="flex items-center gap-2 text-sm font-semibold text-white mb-3">
            <Cpu className="h-4 w-4 text-indigo-400" />
            Environment Specifications
          </h3>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <div className="flex items-center gap-2">
              <span className="text-sm text-slate-500">Runtime:</span>
              <Badge variant="outline" className="border-slate-700 text-slate-300">
                {environmentSpecs.runtime} {environmentSpecs.runtimeVersion}
              </Badge>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-sm text-slate-500">Timeout:</span>
              <Badge variant="outline" className="border-slate-700 text-slate-300">
                {environmentSpecs.timeoutMs / 1000}s
              </Badge>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-sm text-slate-500">Memory:</span>
              <Badge variant="outline" className="border-slate-700 text-slate-300">
                {environmentSpecs.memoryMb}MB
              </Badge>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-sm text-slate-500">Network:</span>
              <Badge variant="outline" className="border-slate-700 text-slate-300">
                <Globe className="mr-1 h-3 w-3" />
                {environmentSpecs.networkAccess === 'full' ? 'Allowed' : 'Restricted'}
              </Badge>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
