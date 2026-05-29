import { ChevronLeft, ChevronRight, Flame } from 'lucide-react';
import { useCallback, useRef } from 'react';
import type { GalleryFunction } from '@/api/composer';
import { Button } from '@/components/ui/button';
import { RUNTIME_COLORS, RUNTIME_ICONS } from '../constants';
import { TrustGauge } from './TrustGauge';

interface RunwayStripProps {
  functions: GalleryFunction[];
  onSelect: (fn: GalleryFunction) => void;
  title?: string;
  fullBleed?: boolean;
}

export function RunwayStrip({
  functions,
  onSelect,
  title = 'Runway — Top Trusted',
  fullBleed = false,
}: RunwayStripProps) {
  const trackRef = useRef<HTMLDivElement>(null);

  const scroll = useCallback((direction: 'left' | 'right') => {
    const track = trackRef.current;
    if (!track) return;
    const amount = Math.max(320, track.clientWidth * 0.75);
    track.scrollBy({ left: direction === 'left' ? -amount : amount, behavior: 'smooth' });
  }, []);

  if (functions.length === 0) {
    return (
      <section className="flyway-runway flyway-runway-empty" aria-label="Runway functions">
        <p className="text-sm text-muted-foreground text-center py-12">No functions cleared for the runway.</p>
      </section>
    );
  }

  return (
    <section
      className={fullBleed ? 'flyway-runway flyway-runway-full' : 'flyway-runway'}
      aria-label="Runway functions"
    >
      <div className="flyway-runway-header">
        <div className="flex items-center gap-2">
          <Flame className="w-4 h-4 text-[var(--flyway-flame)]" />
          <h2 className="text-sm font-semibold uppercase tracking-wider text-muted-foreground">{title}</h2>
          <span className="text-xs font-mono text-muted-foreground/70">({functions.length})</span>
        </div>
        <div className="flyway-runway-scroll-controls">
          <Button
            type="button"
            variant="outline"
            size="icon"
            className="h-8 w-8"
            aria-label="Scroll runway left"
            onClick={() => scroll('left')}
          >
            <ChevronLeft className="w-4 h-4" />
          </Button>
          <Button
            type="button"
            variant="outline"
            size="icon"
            className="h-8 w-8"
            aria-label="Scroll runway right"
            onClick={() => scroll('right')}
          >
            <ChevronRight className="w-4 h-4" />
          </Button>
        </div>
      </div>

      <div ref={trackRef} className="flyway-runway-track" tabIndex={0}>
        {functions.map((fn, i) => {
          const runtime = fn.runtime || 'python';
          const colors = RUNTIME_COLORS[runtime] || RUNTIME_COLORS.python;
          return (
            <article
              key={fn.id}
              className="flyway-runway-card"
              role="button"
              tabIndex={0}
              onClick={() => onSelect(fn)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  onSelect(fn);
                }
              }}
              style={{ borderTop: `2px solid ${colors.primary}` }}
            >
              <span className="flyway-runway-card-rank">#{i + 1}</span>
              <div className="flex items-start gap-3">
                <span className="text-2xl shrink-0">{RUNTIME_ICONS[runtime] || '⚡'}</span>
                <div className="flex-1 min-w-0 pr-8">
                  <h3 className="font-semibold truncate">{fn.title || fn.name}</h3>
                  <p className="text-xs text-muted-foreground">@{fn.author}</p>
                  <p className="text-sm text-muted-foreground mt-2 line-clamp-2">
                    {fn.description || fn.name}
                  </p>
                </div>
                <TrustGauge score={fn.trust_score || 0} runtime={runtime} size={40} />
              </div>
              <div className="flex items-center gap-4 mt-3 text-xs text-muted-foreground font-mono">
                <span>{fn.remix_count || 0} remixes</span>
                <span>{fn.like_count || 0} likes</span>
                <span className="capitalize" style={{ color: colors.glow }}>
                  {runtime}
                </span>
              </div>
            </article>
          );
        })}
      </div>

      <div className="flyway-runway-lights" aria-hidden="true">
        {Array.from({ length: 12 }).map((_, i) => (
          <div key={i} className="flyway-runway-light" style={{ animationDelay: `${i * 0.12}s` }} />
        ))}
      </div>
    </section>
  );
}
