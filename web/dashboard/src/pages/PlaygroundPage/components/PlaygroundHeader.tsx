import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { ArrowLeft, ChevronRight, Zap, Clock, Shield } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { FunctionInfo } from '../store/playgroundStore';

interface PlaygroundHeaderProps {
  functionInfo: FunctionInfo;
}

const runtimeColors: Record<string, string> = {
  python: 'bg-blue-500/10 dark:bg-blue-500/20 text-blue-500 dark:text-blue-400 border-blue-500/20 dark:border-blue-500/30',
  python3: 'bg-blue-500/10 dark:bg-blue-500/20 text-blue-500 dark:text-blue-400 border-blue-500/20 dark:border-blue-500/30',
  'python3.11': 'bg-blue-500/10 dark:bg-blue-500/20 text-blue-500 dark:text-blue-400 border-blue-500/20 dark:border-blue-500/30',
  node: 'bg-green-500/10 dark:bg-green-500/20 text-green-500 dark:text-green-400 border-green-500/20 dark:border-green-500/30',
  node20: 'bg-green-500/10 dark:bg-green-500/20 text-green-500 dark:text-green-400 border-green-500/20 dark:border-green-500/30',
  javascript: 'bg-yellow-500/10 dark:bg-yellow-500/20 text-yellow-500 dark:text-yellow-400 border-yellow-500/20 dark:border-yellow-500/30',
  rust: 'bg-orange-500/10 dark:bg-orange-500/20 text-orange-500 dark:text-orange-400 border-orange-500/20 dark:border-orange-500/30',
  wasm: 'bg-purple-500/10 dark:bg-purple-500/20 text-purple-500 dark:text-purple-400 border-purple-500/20 dark:border-purple-500/30',
};

export function PlaygroundHeader({ functionInfo }: PlaygroundHeaderProps) {
  const { t } = useTranslation();
  const runtimeColor =
    runtimeColors[functionInfo.runtime?.toLowerCase() || ''] ||
    'bg-gray-500/10 dark:bg-gray-500/20 text-gray-500 dark:text-gray-400 border-gray-500/20 dark:border-gray-500/30';

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
          {t('playground.registry')}
        </Link>
        <ChevronRight className="w-3 h-3" />
        <Link
          to={`/fx/${functionInfo.author}/${functionInfo.name}`}
          className="hover:text-text-primary transition-colors"
        >
          {functionInfo.author}/{functionInfo.name}
        </Link>
        <ChevronRight className="w-3 h-3" />
        <span className="text-text-secondary">{t('playground.playground')}</span>
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
              className="text-xs border bg-amber-500/10 dark:bg-amber-500/20 text-amber-500 dark:text-amber-400 border-amber-500/20 dark:border-amber-500/30 gap-1"
            >
              <Clock className="w-3 h-3" />
              {t('playground.cached')}
            </Badge>
          )}

          {functionInfo.reliability_score && functionInfo.reliability_score >= 99 && (
            <Badge
              variant="outline"
              className="text-xs border bg-green-500/10 dark:bg-green-500/20 text-green-500 dark:text-green-400 border-green-500/20 dark:border-green-500/30 gap-1"
            >
              <Shield className="w-3 h-3" />
              {functionInfo.reliability_score}%
            </Badge>
          )}

          <Badge
            variant="outline"
            className="text-xs border bg-indigo-500/10 dark:bg-indigo-500/20 text-indigo-500 dark:text-indigo-400 border-indigo-500/20 dark:border-indigo-500/30 gap-1"
          >
            <Zap className="w-3 h-3" />
            {t('playground.playground')}
          </Badge>
        </div>
      </div>
    </motion.div>
  );
}
