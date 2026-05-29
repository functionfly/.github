import { Button } from '@/components/ui/button';
import { Activity, GitFork, Globe, Radar, Sparkles } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

interface GalleryHeroProps {
  totalFunctions: number;
  totalRemixes: number;
  activeRuntimes: number;
}

export function GalleryHero({ totalFunctions, totalRemixes, activeRuntimes }: GalleryHeroProps) {
  const navigate = useNavigate();

  return (
    <header className="flyway-hero">
      <div className="flyway-hero-radar" aria-hidden="true">
        <div className="flyway-hero-radar-sweep" />
      </div>

      <div className="flex flex-col lg:flex-row lg:items-end lg:justify-between gap-6 px-0">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <Radar className="w-4 h-4 text-[var(--flyway-cyan)]" />
            <span className="text-xs font-mono uppercase tracking-[0.2em] text-muted-foreground">
              The Flyway Registry
            </span>
          </div>
          <h1 className="flyway-title">Function Gallery</h1>
          <p className="flyway-subtitle">
            Discover verified serverless functions from the community. Remix trusted code, explore
            multi-runtime workflows, and deploy to the edge in one click.
          </p>

          <div className="flyway-hud-stats">
            <div className="flyway-hud-stat">
              <Globe className="w-3.5 h-3.5 text-[var(--flyway-flame)]" />
              <span className="flyway-hud-stat-value">{totalFunctions.toLocaleString()}</span>
              <span className="flyway-hud-stat-label">Functions</span>
            </div>
            <div className="flyway-hud-stat">
              <GitFork className="w-3.5 h-3.5 text-[var(--flyway-cyan)]" />
              <span className="flyway-hud-stat-value">{totalRemixes.toLocaleString()}</span>
              <span className="flyway-hud-stat-label">Remixes</span>
            </div>
            <div className="flyway-hud-stat">
              <Activity className="w-3.5 h-3.5 text-emerald-400" />
              <span className="flyway-hud-stat-value">{activeRuntimes}</span>
              <span className="flyway-hud-stat-label">Runtimes</span>
            </div>
          </div>
        </div>

        <div className="flex gap-2 shrink-0">
          <Button variant="outline" size="sm" onClick={() => navigate('/registry')}>
            Full Registry
          </Button>
          <Button
            size="sm"
            className="bg-gradient-to-r from-[var(--flyway-flame)] to-[var(--flyway-afterburner)] hover:opacity-90"
            onClick={() => navigate('/ai/composer')}
          >
            <Sparkles className="w-4 h-4 mr-1.5" />
            Create with AI
          </Button>
        </div>
      </div>
    </header>
  );
}
