import { useState } from 'react';
import { useCityLeaderboard, useCityMovers } from '@/hooks/useCityRankings';
import { Globe, TrendingUp, TrendingDown, Zap, MapPin, RefreshCw, AlertCircle, Sparkles } from 'lucide-react';
import { Link } from 'react-router-dom';
import type { CityRankingEntry } from '@/api/cityRankings';
import '@/styles/sc-rankings.css';

type Category = 'composite' | 'agents' | 'automation' | 'startups' | 'open_source' | 'robotics';

const CATEGORIES: { value: Category; label: string }[] = [
  { value: 'composite', label: 'Composite' },
  { value: 'agents', label: 'Agents' },
  { value: 'automation', label: 'Automation' },
  { value: 'startups', label: 'Startups' },
  { value: 'open_source', label: 'Open Source' },
  { value: 'robotics', label: 'Robotics' },
];

const COUNTRY_OPTIONS: { code: string; label: string }[] = [
  { code: '', label: 'World' },
  { code: 'US', label: 'United States' },
  { code: 'GB', label: 'United Kingdom' },
  { code: 'CA', label: 'Canada' },
  { code: 'DE', label: 'Germany' },
  { code: 'FR', label: 'France' },
  { code: 'IN', label: 'India' },
  { code: 'JP', label: 'Japan' },
  { code: 'AU', label: 'Australia' },
];

export function CityRankingsPage() {
  const [category, setCategory] = useState<Category>('composite');
  const [country, setCountry] = useState('');
  const { data, isLoading, error } = useCityLeaderboard({ country, limit: 100 });
  const gainers = useCityMovers('gainers');
  const losers = useCityMovers('losers');

  const handleRetry = () => {
    window.location.reload();
  };

  return (
    <div className="sc-rankings-page sc-rankings-fade-in">
      {/* Header */}
      <header className="sc-rankings-header">
        <h1 className="sc-rankings-title">
          <Globe className="h-6 w-6" style={{ color: 'var(--status-ok)' }} />
          City Rankings
        </h1>
        <p className="sc-rankings-subtitle">
          Live map of human + AI productivity. Updated every hour.
        </p>
      </header>

      {/* Filter Bar */}
      <div className="sc-rankings-filters">
        <div className="sc-rankings-filter-group">
          <span className="sc-rankings-filter-label">Category:</span>
          {CATEGORIES.map((cat) => (
            <button
              key={cat.value}
              onClick={() => setCategory(cat.value)}
              className={`sc-rankings-filter-btn ${category === cat.value ? 'sc-rankings-filter-btn-active' : ''}`}
            >
              {cat.label}
            </button>
          ))}
        </div>
        <div className="sc-rankings-filter-group">
          <MapPin className="h-4 w-4" style={{ color: 'var(--text-faint)' }} />
          <select
            value={country}
            onChange={(e) => setCountry(e.target.value)}
            className="rounded-md border border-white/10 bg-white/5 px-3 py-1.5 text-sm focus:border-amber-500 focus:outline-none"
            style={{
              fontFamily: 'var(--font-mono)',
              fontSize: '0.75rem',
              color: 'var(--text-dim)',
              backgroundColor: 'var(--chamber-bg)',
            }}
          >
            {COUNTRY_OPTIONS.map((c) => (
              <option key={c.code} value={c.code}>
                {c.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Error Banner */}
      {error && (
        <div className="sc-rankings-error-banner">
          <AlertCircle />
          <span>Failed to load rankings. </span>
          <button
            onClick={handleRetry}
            style={{
              background: 'none',
              border: 'none',
              color: 'inherit',
              textDecoration: 'underline',
              cursor: 'pointer',
              font: 'inherit',
            }}
          >
            Retry
          </button>
          <span>or refresh the page.</span>
        </div>
      )}

      {/* Main Content Grid */}
      <div className="sc-rankings-main-grid">
        {/* Leaderboard Chamber */}
        <div className="sc-rankings-chamber">
          <div className="sc-rankings-chamber-header">
            <span className="sc-rankings-chamber-title">
              <TrendingUp />
              Composite Ranking · 30D Activity
            </span>
            {data && (
              <span style={{ fontFamily: 'var(--font-mono)', fontSize: '0.6875rem', color: 'var(--text-faint)' }}>
                {data.total_ranked} cities
              </span>
            )}
          </div>

          {isLoading ? (
            <RankingsSkeleton />
          ) : !data || data.entries.length === 0 ? (
            <div className="sc-rankings-empty">
              <div className="sc-rankings-empty-icon">
                <Sparkles />
              </div>
              <p className="sc-rankings-empty-text">
                No metros in this region have hit the privacy threshold yet.
                Be the first to set your location!
              </p>
            </div>
          ) : (
            <RankingsTable entries={data.entries} />
          )}
        </div>

        {/* Sidebar */}
        <aside className="sc-rankings-sidebar">
          {/* Top Gainers */}
          <MoversCard
            title="Top Gainers"
            icon={<TrendingUp className="h-4 w-4" />}
            iconClass="sc-rankings-sidebar-title-up"
            data={gainers.data}
            isLoading={gainers.isLoading}
          />

          {/* Top Losers */}
          <MoversCard
            title="Top Losers"
            icon={<TrendingDown className="h-4 w-4" />}
            iconClass="sc-rankings-sidebar-title-down"
            data={losers.data}
            isLoading={losers.isLoading}
          />

          {/* How It Works */}
          <HowItWorksCard />
        </aside>
      </div>

      {/* Footer Info */}
      {data && (
        <div style={{
          fontFamily: 'var(--font-mono)',
          fontSize: '0.6875rem',
          color: 'var(--text-faint)',
          textAlign: 'center',
          paddingTop: '1rem',
          borderTop: '1px solid var(--chamber-edge)',
        }}>
          Updated {new Date(data.period_end).toLocaleString()} · {data.total_ranked} cities ranked
          {data.cache_hit && ' · cached'} · {data.category}
        </div>
      )}
    </div>
  );
}

function RankingsTable({ entries }: { entries: CityRankingEntry[] }) {
  return (
    <div style={{ overflow: 'auto' }}>
      <table className="sc-rankings-table">
        <thead>
          <tr>
            <th>#</th>
            <th>Metro</th>
            <th style={{ textAlign: 'right' }}>Per Capita</th>
            <th style={{ textAlign: 'right' }} className="sm:table-cell">Raw</th>
            <th style={{ textAlign: 'right' }} className="md:table-cell">Active</th>
            <th style={{ textAlign: 'right' }} className="md:table-cell">Deploys</th>
          </tr>
        </thead>
        <tbody>
          {entries.map((e) => (
            <tr key={e.metro_slug}>
              <td>
                <div className="sc-rankings-rank">
                  <span className="sc-rankings-rank-number">{e.rank}</span>
                  {e.rank_delta > 0 && (
                    <span className="sc-rankings-rank-delta sc-rankings-rank-delta-up">
                      ▲{e.rank_delta}
                    </span>
                  )}
                  {e.rank_delta < 0 && (
                    <span className="sc-rankings-rank-delta sc-rankings-rank-delta-down">
                      ▼{Math.abs(e.rank_delta)}
                    </span>
                  )}
                </div>
              </td>
              <td>
                <Link
                  to={`/rankings/cities/${e.metro_slug}`}
                  className="sc-rankings-metro-link"
                >
                  {e.metro_name}
                </Link>
                <span className="sc-rankings-metro-country">{e.country_code}</span>
              </td>
              <td style={{ textAlign: 'right' }}>
                <span className="sc-rankings-score">{e.score_per_capita.toFixed(4)}</span>
              </td>
              <td style={{ textAlign: 'right' }} className="sm:table-cell">
                <span className="sc-rankings-score-secondary">{e.score_raw.toFixed(3)}</span>
              </td>
              <td style={{ textAlign: 'right' }} className="md:table-cell">
                <span className="sc-rankings-score-secondary">{e.active_users}</span>
              </td>
              <td style={{ textAlign: 'right' }} className="md:table-cell">
                <span className="sc-rankings-score-secondary">{e.deployments.toLocaleString()}</span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function RankingsSkeleton() {
  return (
    <div className="sc-rankings-skeleton">
      {Array.from({ length: 10 }).map((_, i) => (
        <div key={i} className="sc-rankings-skeleton-row">
          <div className="sc-rankings-skeleton-cell sc-rankings-skeleton-rank" />
          <div className="sc-rankings-skeleton-cell sc-rankings-skeleton-metro" />
          <div className="sc-rankings-skeleton-cell sc-rankings-skeleton-score" />
        </div>
      ))}
    </div>
  );
}

interface MoversCardProps {
  title: string;
  icon: React.ReactNode;
  iconClass: string;
  data?: { entries: CityRankingEntry[] };
  isLoading: boolean;
}

function MoversCard({ title, icon, iconClass, data, isLoading }: MoversCardProps) {
  return (
    <div className="sc-rankings-sidebar-card">
      <h3 className={`sc-rankings-sidebar-title ${iconClass}`}>
        {icon}
        {title}
      </h3>
      {isLoading ? (
        <div className="sc-rankings-skeleton">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="sc-rankings-skeleton-row">
              <div className="sc-rankings-skeleton-cell" style={{ flex: 1 }} />
              <div className="sc-rankings-skeleton-cell" style={{ width: '2rem' }} />
            </div>
          ))}
        </div>
      ) : !data || data.entries.length === 0 ? (
        <p style={{ fontSize: '0.8125rem', color: 'var(--text-faint)', padding: '0.5rem 0' }}>
          No movers yet — check back after the next hourly recompute.
        </p>
      ) : (
        <ol className="sc-rankings-movers-list">
          {data.entries.slice(0, 5).map((e) => (
            <li key={e.metro_slug} className="sc-rankings-movers-item">
              <Link
                to={`/rankings/cities/${e.metro_slug}`}
                className="sc-rankings-movers-link"
              >
                {e.metro_name}
              </Link>
              <span className={`sc-rankings-movers-delta ${e.rank_delta > 0 ? 'sc-rankings-movers-delta-up' : 'sc-rankings-movers-delta-down'}`}>
                {e.rank_delta > 0 ? '▲' : '▼'} {Math.abs(e.rank_delta)}
              </span>
            </li>
          ))}
        </ol>
      )}
    </div>
  );
}

function HowItWorksCard() {
  return (
    <div className="sc-rankings-howitworks">
      <h3 className="sc-rankings-howitworks-title">
        <Zap />
        How it works
      </h3>
      <div className="sc-rankings-howitworks-text">
        <p>
          Score is recomputed hourly from real FunctionFly activity: active users,
          deployments, executions, founder earnings, and new user growth (log-scaled
          and weighted).
        </p>
        <p>
          Per-capita normalization lets dense AI hubs compete with megacities.
          Cities with fewer than 5 active builders are hidden.
        </p>
        <p>
          <Link to="/settings/profile" className="sc-rankings-howitworks-link">
            Set your location →
          </Link>
        </p>
      </div>
    </div>
  );
}
