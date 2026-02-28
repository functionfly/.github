import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ArrowLeft, ChevronRight, Zap, Clock, Shield } from 'lucide-react';
import { FunctionInfo } from '../store/playgroundStore';

interface PlaygroundHeaderProps {
  functionInfo: FunctionInfo;
}

const runtimeColors: Record<string, string> = {
  python: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
  python3: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
  'python3.11': 'bg-blue-500/10 text-blue-400 border-blue-500/20',
  node: 'bg-green-500/10 text-green-400 border-green-500/20',
  node20: 'bg-green-500/10 text-green-400 border-green-500/20',
  javascript: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20',
  rust: 'bg-orange-500/10 text-orange-400 border-orange-500/20',
  wasm: 'bg-purple-500/10 text-purple-400 border-purple-500/20',
};

export function PlaygroundHeader({ functionInfo }: PlaygroundHeaderProps) {
  const runtimeColor =
    runtimeColors[functionInfo.runtime?.toLowerCase() || ''] ||
    'bg-gray-500/10 text-gray-400 border-gray-500/20';

  return (
    <motion.div
      initial={{ opacity: 0, y: -8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className="border-b border-border-subtle bg-bg-primary px-4 py-3"
    >
      {/* Breadcrumb */}
      <div className="flex items-center gap-1.5 text-xs text-text-muted mb-2">
        <Link
          to="/registry"
          className="hover:text-text-primary transition-colors"
        >
          Registry
        </Link>
        <ChevronRight className="w-3 h-3" />
        <Link
          to={`/fx/${functionInfo.author}/${functionInfo.name}`}
          className="hover:text-text-primary transition-colors"
        >
          {functionInfo.author}/{functionInfo.name}
        </Link>
        <ChevronRight className="w-3 h-3" />
        <span className="text-text-secondary">Playground</span>
      </div>

      {/* Title row */}
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-3 min-w-0">
          <Link to={`/fx/${functionInfo.author}/${functionInfo.name}`}>
            <Button
              variant="ghost"
              size="sm"
              className="h-7 px-2 text-text-muted hover:text-text-primary shrink-0"
            >
              <ArrowLeft className="w-3.5 h-3.5" />
            </Button>
          </Link>

          <div className="min-w-0">
            <h1 className="text-lg font-semibold text-text-primary truncate">
              {functionInfo.title || `${functionInfo.author}/${functionInfo.name}`}
            </h1>
            {functionInfo.description && (
              <p className="text-xs text-text-muted truncate max-w-xl">
                {functionInfo.description}
              </p>
            )}
          </div>
        </div>

        {/* Badges */}
        <div className="flex items-center gap-2 shrink-0">
          <Badge variant="outline" className="text-xs">
            v{functionInfo.version}
          </Badge>

          {functionInfo.runtime && (
            <Badge
              variant="outline"
              className={`text-xs border ${runtimeColor}`}
            >
              {functionInfo.runtime}
            </Badge>
          )}

          {functionInfo.cache_ttl && functionInfo.cache_ttl > 0 && (
            <Badge
              variant="outline"
              className="text-xs border bg-amber-500/10 text-amber-400 border-amber-500/20 gap-1"
            >
              <Clock className="w-3 h-3" />
              Cached
            </Badge>
          )}

          {functionInfo.reliability_score && functionInfo.reliability_score >= 99 && (
            <Badge
              variant="outline"
              className="text-xs border bg-green-500/10 text-green-400 border-green-500/20 gap-1"
            >
              <Shield className="w-3 h-3" />
              {functionInfo.reliability_score}%
            </Badge>
          )}

          <Badge
            variant="outline"
            className="text-xs border bg-indigo-500/10 text-indigo-400 border-indigo-500/20 gap-1"
          >
            <Zap className="w-3 h-3" />
            Playground
          </Badge>
        </div>
      </div>
    </motion.div>
  );
}
