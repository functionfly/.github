import { motion } from 'framer-motion';
import { GitFork, Heart, Star } from 'lucide-react';
import type { GalleryFunction } from '@/api/composer';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import {
  CATEGORY_META,
  RUNTIME_COLORS,
  RUNTIME_ICONS,
} from '../constants';
import { TrustGauge } from './TrustGauge';

interface FunctionTileProps {
  fn: GalleryFunction;
  featured?: boolean;
  onClick: () => void;
  index?: number;
}

export function FunctionTile({ fn, featured = false, onClick, index = 0 }: FunctionTileProps) {
  const runtime = fn.runtime || 'python';
  const colors = RUNTIME_COLORS[runtime] || RUNTIME_COLORS.python;
  const categoryMeta = CATEGORY_META[fn.category || 'default'] || CATEGORY_META.default;

  return (
    <motion.article
      className={cn('flyway-tile', featured && 'flyway-tile-featured')}
      onClick={onClick}
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.04, duration: 0.35 }}
      whileHover={{ scale: featured ? 1.01 : 1.02 }}
    >
      <div className="flyway-tile-livery" style={{ background: `linear-gradient(90deg, ${colors.primary}, ${colors.glow})` }} />
      <div className="flyway-tile-glow" style={{ background: colors.primary }} />

      <div className={cn('relative p-5 flex flex-col', featured ? 'h-full min-h-[280px]' : 'h-full min-h-[180px]')}>
        <div className="flex items-start justify-between gap-3 mb-3">
          <div className="flex items-center gap-2.5 min-w-0">
            <span className="text-2xl shrink-0">{RUNTIME_ICONS[runtime] || '⚡'}</span>
            <div className="min-w-0">
              <h3 className={cn('font-semibold text-foreground truncate', featured && 'text-lg')}>
                {fn.title || fn.name}
              </h3>
              <p className="text-xs text-muted-foreground truncate">@{fn.author}</p>
            </div>
          </div>
          <TrustGauge score={fn.trust_score || 0} runtime={runtime} />
        </div>

        <p className={cn('text-sm text-muted-foreground leading-relaxed flex-1', featured ? 'line-clamp-4' : 'line-clamp-2')}>
          {fn.description || 'No description available'}
        </p>

        <div className="flex items-center justify-between mt-4 pt-3 border-t border-white/5">
          <div className="flex items-center gap-2">
            <Badge
              variant="outline"
              className="text-xs border-0 capitalize"
              style={{ backgroundColor: `${categoryMeta.color}20`, color: categoryMeta.color }}
            >
              {categoryMeta.label}
            </Badge>
            <Badge variant="outline" className="text-xs border-0 font-mono" style={{ backgroundColor: `${colors.primary}15`, color: colors.glow }}>
              {runtime}
            </Badge>
          </div>

          <div className="flex items-center gap-3 text-xs text-muted-foreground">
            <span className="flex items-center gap-1" title="Remixes">
              <GitFork className="w-3 h-3" />
              {fn.remix_count || 0}
            </span>
            <span className="flex items-center gap-1" title="Likes">
              <Heart className="w-3 h-3" />
              {fn.like_count || 0}
            </span>
            {featured && (
              <span className="flex items-center gap-1 text-amber-400" title="Trust">
                <Star className="w-3 h-3 fill-current" />
              </span>
            )}
          </div>
        </div>

        {fn.tags && fn.tags.length > 0 && (
          <div className="flex flex-wrap gap-1 mt-2">
            {fn.tags.slice(0, featured ? 4 : 2).map((tag) => (
              <span key={tag} className="text-[10px] px-1.5 py-0.5 rounded bg-white/5 text-muted-foreground">
                #{tag}
              </span>
            ))}
          </div>
        )}
      </div>
    </motion.article>
  );
}
