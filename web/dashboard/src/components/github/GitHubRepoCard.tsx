import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Star,
  GitFork,
  Lock,
  Globe,
  Clock,
  Download,
  ScanSearch,
  ChevronDown,
  ChevronUp,
  Code2,
} from 'lucide-react';
import { Card, CardContent, CardFooter, CardHeader } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { cn, formatDistanceToNow, formatNumber } from '@/lib/utils';
import type { GitHubRepo, DetectedFunction } from '@/types/github';

const LANGUAGE_COLORS: Record<string, string> = {
  JavaScript: '#f1e05a',
  TypeScript: '#3178c6',
  Python: '#3572A5',
  Go: '#00ADD8',
  Rust: '#dea584',
  Ruby: '#701516',
  Java: '#b07219',
  'C++': '#f34b7d',
  C: '#555555',
  PHP: '#4F5D95',
  Swift: '#F05138',
  Kotlin: '#A97BFF',
  Dart: '#00B4AB',
  Shell: '#89e051',
  Lua: '#000080',
};

function getConfidenceColor(confidence: number): string {
  if (confidence >= 80) return 'bg-emerald-500';
  if (confidence >= 50) return 'bg-amber-500';
  return 'bg-red-500';
}

function getConfidenceTextColor(confidence: number): string {
  if (confidence >= 80) return 'text-emerald-500';
  if (confidence >= 50) return 'text-amber-500';
  return 'text-red-500';
}

function getRuntimeIcon(runtime: string): string {
  const lower = runtime.toLowerCase();
  if (lower.includes('node') || lower.includes('javascript') || lower.includes('typescript')) return 'JS';
  if (lower.includes('python')) return 'PY';
  if (lower.includes('go')) return 'GO';
  if (lower.includes('rust')) return 'RS';
  if (lower.includes('ruby')) return 'RB';
  return runtime.slice(0, 2).toUpperCase();
}

interface DetectedFunctionsListProps {
  functions: DetectedFunction[];
}

function DetectedFunctionsList({ functions }: DetectedFunctionsListProps) {
  return (
    <div className="space-y-2 pt-2 border-t border-border-subtle">
      {functions.map((fn) => (
        <div
          key={fn.name}
          className="flex items-center gap-2 text-xs p-2 rounded-md bg-bg-secondary/50"
        >
          <span className="inline-flex items-center justify-center h-5 w-8 rounded text-[10px] font-bold bg-muted text-text-secondary shrink-0">
            {getRuntimeIcon(fn.runtime)}
          </span>
          <span className="font-medium text-text-primary truncate flex-1">{fn.name}</span>
          <span className={cn('font-mono text-[10px]', getConfidenceTextColor(fn.confidence))}>
            {fn.confidence}%
          </span>
        </div>
      ))}
    </div>
  );
}

interface GitHubRepoCardProps {
  repo: GitHubRepo;
  onImport?: (repo: GitHubRepo) => void;
  onScan?: (repo: GitHubRepo) => void;
  className?: string;
}

export function GitHubRepoCard({ repo, onImport, onScan, className }: GitHubRepoCardProps) {
  const [showFunctions, setShowFunctions] = useState(false);
  const languages = repo.languages ? (Object.entries(repo.languages) as [string, number][]).sort(([, a], [, b]) => b - a) : [];
  const detectedFunctions = repo.detected_functions ?? [];
  const hasFunctions = detectedFunctions.length > 0;

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className={className}
    >
      <Card className="group hover:-translate-y-0.5 hover:shadow-lg transition-all duration-200 h-full flex flex-col">
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between gap-2">
            <div className="flex items-center gap-2 min-w-0">
              <svg role="img" viewBox="0 0 24 24" className="h-4 w-4 text-text-secondary shrink-0" xmlns="http://www.w3.org/2000/svg"><path fill="currentColor" d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/></svg>
              <a
                href={repo.html_url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm font-semibold text-text-primary hover:text-brand-500 truncate transition-colors"
              >
                {repo.full_name}
              </a>
            </div>
            <div className="flex items-center gap-1.5 shrink-0">
              {repo.is_private ? (
                <Badge variant="secondary" className="text-[10px] gap-1">
                  <Lock className="h-2.5 w-2.5" />
                  Private
                </Badge>
              ) : (
                <Badge variant="outline" className="text-[10px] gap-1">
                  <Globe className="h-2.5 w-2.5" />
                  Public
                </Badge>
              )}
              {repo.is_archived && (
                <Badge variant="secondary" className="text-[10px]">Archived</Badge>
              )}
            </div>
          </div>

          {repo.description && (
            <p className="text-xs text-text-secondary line-clamp-2 mt-1">{repo.description}</p>
          )}
        </CardHeader>

        <CardContent className="flex-1 space-y-3">
          <div className="flex items-center gap-3 text-xs text-text-muted">
            {repo.stars_count > 0 && (
              <span className="flex items-center gap-1">
                <Star className="h-3 w-3" />
                {formatNumber(repo.stars_count)}
              </span>
            )}
            {repo.forks_count > 0 && (
              <span className="flex items-center gap-1">
                <GitFork className="h-3 w-3" />
                {formatNumber(repo.forks_count)}
              </span>
            )}
            {repo.pushed_at && (
              <span className="flex items-center gap-1">
                <Clock className="h-3 w-3" />
                {formatDistanceToNow(repo.pushed_at, { addSuffix: true })}
              </span>
            )}
          </div>

          {languages.length > 0 && (
            <div className="flex flex-wrap gap-1">
              {languages.slice(0, 4).map(([lang, bytes]) => (
                <TooltipProvider key={lang}>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Badge
                        variant="secondary"
                        className="text-[10px] gap-1 cursor-default"
                      >
                        <span
                          className="h-2 w-2 rounded-full shrink-0"
                          style={{ backgroundColor: LANGUAGE_COLORS[lang] ?? '#6b7280' }}
                        />
                        {lang}
                      </Badge>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>{formatNumber(bytes)} bytes</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              ))}
              {languages.length > 4 && (
                <Badge variant="secondary" className="text-[10px]">
                  +{languages.length - 4}
                </Badge>
              )}
            </div>
          )}

          {hasFunctions && (
            <div>
              <button
                onClick={() => setShowFunctions(!showFunctions)}
                className="flex items-center gap-1.5 text-xs text-text-secondary hover:text-text-primary transition-colors w-full"
                aria-expanded={showFunctions}
                aria-label={`${showFunctions ? 'Hide' : 'Show'} ${detectedFunctions.length} detected functions`}
              >
                <Code2 className="h-3 w-3" />
                <span className="font-medium">{detectedFunctions.length} function{detectedFunctions.length !== 1 ? 's' : ''} detected</span>
                {showFunctions ? <ChevronUp className="h-3 w-3 ml-auto" /> : <ChevronDown className="h-3 w-3 ml-auto" />}
              </button>
              <AnimatePresence>
                {showFunctions && (
                  <motion.div
                    initial={{ height: 0, opacity: 0 }}
                    animate={{ height: 'auto', opacity: 1 }}
                    exit={{ height: 0, opacity: 0 }}
                    transition={{ duration: 0.2 }}
                    className="overflow-hidden"
                  >
                    <DetectedFunctionsList functions={detectedFunctions} />
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
          )}
        </CardContent>

        <CardFooter className="pt-3 border-t border-border-subtle gap-2">
          <Button
            variant="default"
            size="sm"
            onClick={() => onImport?.(repo)}
            disabled={!hasFunctions}
            className="flex-1"
            aria-label={`Import functions from ${repo.name}`}
          >
            <Download className="h-3.5 w-3.5 mr-1.5" />
            Import
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onScan?.(repo)}
            aria-label={`Scan ${repo.name} for functions`}
          >
            <ScanSearch className="h-3.5 w-3.5 mr-1.5" />
            Scan
          </Button>
        </CardFooter>
      </Card>
    </motion.div>
  );
}
