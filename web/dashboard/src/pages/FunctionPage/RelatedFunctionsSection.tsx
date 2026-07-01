import { useQuery } from '@tanstack/react-query';
import { registryApi, type RegistryFunction } from '@/api/registry';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { motion } from 'framer-motion';
import { ChevronRight, Package, Shield, Star, Zap } from 'lucide-react';
import { Link } from 'react-router-dom';

interface RelatedFunctionsSectionProps {
  author: string;
  name: string;
  category?: string;
  tags?: string[];
}

export function RelatedFunctionsSection({ author, name, category, tags }: RelatedFunctionsSectionProps) {
  const { data, isLoading } = useQuery({
    queryKey: ['related-functions', category, tags?.join(',')],
    queryFn: async () => {
      try {
        const result = await registryApi.getFunctions({
          category: category || undefined,
          tags: tags && tags.length > 0 ? tags : undefined,
          limit: 5,
        });
        return (result?.functions ?? []).filter(
          (f) => !(f.author === author && f.name === name)
        ).slice(0, 4);
      } catch {
        return [];
      }
    },
    enabled: !!category || (tags != null && tags.length > 0),
    staleTime: 5 * 60 * 1000,
  });

  const related = data ?? [];

  if (!isLoading && related.length === 0) {
    return (
      <Card style={{ background: 'var(--panel-raised)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)' }}>
        <CardContent className="p-6">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 rounded-[var(--radius-lg)] flex items-center justify-center" style={{ background: 'rgba(143, 255, 208, 0.08)' }}>
              <Package className="w-6 h-6" style={{ color: 'var(--status-ok)' }} />
            </div>
            <div className="flex-1">
              <h3 className="font-semibold text-lg" style={{ fontFamily: 'var(--font-display)' }}>Explore More Functions</h3>
              <p className="text-sm" style={{ color: 'var(--text-dim)' }}>
                Discover related functions in the registry to build powerful workflows
              </p>
            </div>
            <Link
              to="/registry"
              className="function-page-related-cta-button flex items-center"
            >
              Browse Registry
              <ChevronRight className="w-4 h-4 ml-1" />
            </Link>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
      className="function-page-section"
    >
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <Package className="h-6 w-6" style={{ color: 'var(--status-ok)' }} />
          <h2 className="text-2xl font-bold" style={{ fontFamily: 'var(--font-display)' }}>
            Related Functions
          </h2>
        </div>
        <Link
          to="/registry"
          className="text-sm flex items-center gap-1 hover:underline"
          style={{ color: 'var(--foil-b)' }}
        >
          Browse Registry
          <ChevronRight className="w-3 h-3" />
        </Link>
      </div>

      {isLoading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-36 rounded-[var(--radius-lg)]" />
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {related.map((fn) => (
            <Link
              key={fn.id}
              to={`/fx/${fn.author}/${fn.name}`}
              className="block group"
            >
              <Card
                className="h-full transition-all duration-200 hover:scale-[1.02]"
                style={{
                  background: 'var(--panel-raised)',
                  borderColor: 'var(--panel-edge)',
                  borderRadius: 'var(--radius-lg)',
                }}
              >
                <CardContent className="p-4 space-y-3">
                  <div className="flex items-center justify-between">
                    <p
                      className="text-sm font-semibold truncate"
                      style={{ color: 'var(--text)' }}
                    >
                      {fn.title || fn.name}
                    </p>
                    {fn.trust_score != null && fn.trust_score >= 80 && (
                      <Shield className="w-3.5 h-3.5 shrink-0" style={{ color: 'var(--status-ok)' }} />
                    )}
                  </div>
                  <p
                    className="text-xs line-clamp-2"
                    style={{ color: 'var(--text-faint)' }}
                  >
                    {fn.description || 'No description'}
                  </p>
                  <div className="flex items-center gap-2 flex-wrap">
                    {fn.tags?.slice(0, 2).map((tag) => (
                      <Badge
                        key={tag}
                        variant="secondary"
                        className="text-[10px] px-1.5 py-0"
                      >
                        {tag}
                      </Badge>
                    ))}
                  </div>
                  <div className="flex items-center gap-3 text-[10px]" style={{ color: 'var(--text-dim)' }}>
                    <span className="flex items-center gap-1">
                      <Zap className="w-3 h-3" />
                      {fn.popularity_score > 0 ? fn.popularity_score : '—'}
                    </span>
                    <span className="flex items-center gap-1">
                      <Star className="w-3 h-3" />
                      {fn.total_ratings > 0 ? fn.total_ratings : '—'}
                    </span>
                    <span className="ml-auto font-mono">
                      @{fn.author}
                    </span>
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </motion.div>
  );
}
