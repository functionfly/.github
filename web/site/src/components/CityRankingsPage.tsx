import { useEffect, useState, useCallback } from 'react';
import { Globe, TrendingUp, TrendingDown, MapPin, Sparkles, Users, Zap, RefreshCw, AlertCircle } from 'lucide-react';
import CityRankingsGlobe, { type GlobePoint } from './CityRankingsGlobe';
import {
  Chamber,
  CornerBrace,
  StatusPill,
  GaugeStrip,
  AnnotationTag,
  SealedButton,
  TrustBadge,
  RankDisplay,
} from './rankings';
import { APP_DASHBOARD_ORIGIN } from '../config';

interface CityRankingEntry {
  rank: number;
  previous_rank: number;
  rank_delta: number;
  metro_slug: string;
  metro_name: string;
  country_code: string;
  population: number;
  score_raw: number;
  score_per_capita: number;
  active_users: number;
  deployments: number;
  executions_30d: number;
  founder_earnings_cents: number;
  new_users_30d: number;
  period_end: string;
}

interface LeaderboardResponse {
  period_end: string;
  total_ranked: number;
  entries: CityRankingEntry[];
  category: string;
  country?: string;
  cache_hit: boolean;
}

interface StateRankingEntry {
  rank: number;
  state_code: string;
  state_name: string;
  country_code: string;
  population: number;
  score_raw: number;
  score_per_capita: number;
  active_users: number;
  deployments: number;
  executions_30d: number;
  metro_count: number;
  ranked_metros: number;
  period_end: string;
}

interface StatesResponse {
  period_end: string;
  total_states: number;
  entries: StateRankingEntry[];
  category: string;
  country?: string;
  cache_hit: boolean;
}

interface MoversResponse {
  entries: CityRankingEntry[];
  category: string;
  cache_hit: boolean;
}

interface MapPointsResponse {
  period_end: string;
  points: GlobePoint[];
  category: string;
  cache_hit: boolean;
}

interface FetchState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
}

type Category = 'composite' | 'agents' | 'automation' | 'startups' | 'open_source' | 'robotics';

const CATEGORIES: { value: Category; label: string }[] = [
  { value: 'composite', label: 'Composite' },
  { value: 'agents', label: 'Agents' },
  { value: 'automation', label: 'Automation' },
  { value: 'startups', label: 'Startups' },
  { value: 'open_source', label: 'Open Source' },
  { value: 'robotics', label: 'Robotics' },
];

const API_BASE = (import.meta as any).env?.PUBLIC_API_URL || 'https://api.functionfly.com';

const MAX_RETRIES = 2;
const RETRY_DELAY = 1000;

async function fetchWithRetry<T>(url: string, retries = MAX_RETRIES): Promise<T> {
  let lastError: Error | null = null;
  for (let i = 0; i <= retries; i++) {
    try {
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      const data = await response.json();
      if (!data || typeof data !== 'object') {
        throw new Error('Invalid JSON response');
      }
      return data as T;
    } catch (err) {
      lastError = err as Error;
      if (i < retries) {
        await new Promise((r) => setTimeout(r, RETRY_DELAY * (i + 1)));
      }
    }
  }
  throw lastError || new Error('Failed to fetch');
}

function createFetchState<T>(): FetchState<T> {
  return { data: null, loading: true, error: null };
}

export default function CityRankingsPage() {
  const [category, setCategory] = useState<Category>('composite');
  const [country, setCountry] = useState<string>('');
  const [leaderboard, setLeaderboard] = useState<FetchState<LeaderboardResponse>>(createFetchState());
  const [gainers, setGainers] = useState<FetchState<MoversResponse>>(createFetchState());
  const [losers, setLosers] = useState<FetchState<MoversResponse>>(createFetchState());
  const [states, setStates] = useState<FetchState<StatesResponse>>(createFetchState());
  const [mapPoints, setMapPoints] = useState<FetchState<MapPointsResponse>>(createFetchState());
  const [visibleTiers, setVisibleTiers] = useState<Array<'gold' | 'blue' | 'green'>>([]);
  const [retryKey, setRetryKey] = useState(0);

  const toggleTier = (tier: 'gold' | 'blue' | 'green') => {
    setVisibleTiers((prev) =>
      prev.includes(tier) ? prev.filter((t) => t !== tier) : [...prev, tier]
    );
  };

  const buildUrl = useCallback((path: string) => {
    const params = new URLSearchParams();
    if (category) params.set('category', category);
    if (country) params.set('country', country);
    const qs = params.toString();
    return `${API_BASE}/v1/${path}${qs ? `?${qs}` : ''}`;
  }, [category, country]);

  useEffect(() => {
    let cancelled = false;

    const loadData = async () => {
      setLeaderboard(createFetchState());
      setGainers(createFetchState());
      setLosers(createFetchState());
      setStates(createFetchState());
      setMapPoints(createFetchState());

      try {
        const [lb, g, l, s, m] = await Promise.all([
          fetchWithRetry<LeaderboardResponse>(buildUrl('city-rankings?limit=100')),
          fetchWithRetry<MoversResponse>(buildUrl('city-rankings/movers?direction=gainers&limit=10')),
          fetchWithRetry<MoversResponse>(buildUrl('city-rankings/movers?direction=losers&limit=10')),
          fetchWithRetry<StatesResponse>(buildUrl('city-rankings/states?country=US')),
          fetchWithRetry<MapPointsResponse>(buildUrl('city-rankings/map')),
        ]);

        if (cancelled) return;

        setLeaderboard({ data: lb, loading: false, error: null });
        setGainers({ data: g, loading: false, error: null });
        setLosers({ data: l, loading: false, error: null });
        setStates({ data: s, loading: false, error: null });
        setMapPoints({ data: m, loading: false, error: null });
      } catch (err) {
        if (cancelled) return;
        const msg = err instanceof Error ? err.message : 'Failed to load data';
        setLeaderboard({ data: null, loading: false, error: msg });
        setGainers({ data: null, loading: false, error: msg });
        setLosers({ data: null, loading: false, error: msg });
        setStates({ data: null, loading: false, error: msg });
        setMapPoints({ data: null, loading: false, error: msg });
      }
    };

    loadData();
    return () => { cancelled = true; };
  }, [buildUrl, retryKey]);

  const handleRetry = () => setRetryKey((k) => k + 1);

  const onSelectMetro = (p: GlobePoint) => {
    window.location.href = `${APP_DASHBOARD_ORIGIN}/rankings/cities/${p.metro_slug}`;
  };

  const isLoading = leaderboard.loading || gainers.loading || losers.loading || states.loading || mapPoints.loading;
  const hasError = leaderboard.error || gainers.error || losers.error || states.error || mapPoints.error;

  return (
    <div className="min-h-screen" style={{ background: 'var(--chamber-bg)' }}>
      {/* Header Section */}
      <header className="rankings-header" style={{ background: 'var(--chamber-panel)', borderBottom: '1px solid var(--chamber-panel-edge)' }}>
        <div style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0, background: 'radial-gradient(ellipse 80% 60% at 20% 0%, rgba(143, 255, 208, 0.08) 0%, transparent 70%)', pointerEvents: 'none' }} />
        <div style={{ position: 'relative', zIndex: 1 }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '12px', marginBottom: '16px' }}>
            <Globe style={{ width: 24, height: 24, color: 'var(--chamber-status-ok)' }} />
            <span style={{ fontFamily: "'IBM Plex Mono', monospace", fontSize: '12px', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.08em', color: 'var(--chamber-status-ok)' }}>
              FunctionFly City Rankings™
            </span>
          </div>
          <h1 className="rankings-header__title">
            The live map of human <span style={{ color: 'var(--chamber-status-ok)' }}>+ AI</span> productivity.
          </h1>
          <p className="rankings-header__subtitle">
            We rank cities, not vibes. Every city on this list earned its place from real
            FunctionFly activity — function executions, deployments, founders, and new builders
            in the last 30 days.
          </p>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '24px', marginTop: '24px', fontFamily: "'IBM Plex Mono', monospace", fontSize: '11px', color: 'var(--chamber-text-dim)' }}>
            <span style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
              <Zap style={{ width: 14, height: 14, color: 'var(--chamber-status-ok)' }} />
              Updated every hour
            </span>
            <span style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
              <Users style={{ width: 14, height: 14, color: 'var(--chamber-status-ok)' }} />
              Privacy threshold: k ≥ 5
            </span>
            <span style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
              <MapPin style={{ width: 14, height: 14, color: 'var(--chamber-status-ok)' }} />
              Opt-out any time
            </span>
          </div>
        </div>
      </header>

      {/* Filter Bar */}
      <section style={{ padding: '16px 24px', maxWidth: '1200px', margin: '0 auto' }}>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: '16px', alignItems: 'center', justifyContent: 'space-between' }}>
          {/* Category Filter */}
          <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
            <span style={{ fontFamily: "'IBM Plex Mono', monospace", fontSize: '11px', color: 'var(--chamber-text-faint)', alignSelf: 'center' }}>Category:</span>
            {CATEGORIES.map((cat) => (
              <button
                key={cat.value}
                onClick={() => setCategory(cat.value)}
                style={{
                  padding: '6px 12px',
                  borderRadius: 'var(--chamber-radius)',
                  fontFamily: "'IBM Plex Mono', monospace",
                  fontSize: '11px',
                  fontWeight: 500,
                  cursor: 'pointer',
                  transition: 'all 0.15s ease',
                  border: '1px solid',
                  borderColor: category === cat.value ? 'var(--chamber-status-ok)' : 'var(--chamber-panel-edge)',
                  background: category === cat.value ? 'var(--chamber-status-ok)' : 'transparent',
                  color: category === cat.value ? 'var(--chamber-bg)' : 'var(--chamber-text-dim)',
                }}
              >
                {cat.label}
              </button>
            ))}
          </div>
          {/* Retry Button */}
          <button
            onClick={handleRetry}
            disabled={isLoading}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              padding: '6px 12px',
              borderRadius: 'var(--chamber-radius)',
              fontFamily: "'IBM Plex Mono', monospace",
              fontSize: '11px',
              cursor: isLoading ? 'wait' : 'pointer',
              border: '1px solid var(--chamber-panel-edge)',
              background: 'transparent',
              color: 'var(--chamber-text-dim)',
              opacity: isLoading ? 0.6 : 1,
            }}
          >
            <RefreshCw style={{ width: 12, height: 12, animation: isLoading ? 'spin 1s linear infinite' : 'none' }} />
            Refresh
          </button>
        </div>
      </section>

      {/* Global Error Banner */}
      {hasError && (
        <section style={{ padding: '0 24px', maxWidth: '1200px', margin: '0 auto' }}>
          <div style={{
            padding: '12px 16px',
            borderRadius: 'var(--chamber-radius)',
            background: 'rgba(255, 107, 107, 0.1)',
            border: '1px solid rgba(255, 107, 107, 0.3)',
            color: 'var(--chamber-accent)',
            fontFamily: "'IBM Plex Mono', monospace",
            fontSize: '12px',
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            marginBottom: '16px',
          }}>
            <AlertCircle style={{ width: 16, height: 16, flexShrink: 0 }} />
            <span>
              Failed to load some data.{' '}
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
              {' '}or refresh the page.
            </span>
          </div>
        </section>
      )}

      {/* AI World Map Section */}
      <section style={{ padding: '0 24px 48px', maxWidth: '1200px', margin: '0 auto' }}>
        <Chamber ribs={true} animate={true}>
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <AnnotationTag label="AI WORLD MAP · LIVE DATA" position="tl" />

          <div style={{ padding: '24px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: '16px' }}>
              <div>
                <h2 style={{ fontFamily: "'Space Grotesk', sans-serif", fontSize: '20px', fontWeight: 600, color: 'var(--chamber-text)', marginBottom: '8px', marginTop: '16px' }}>
                  AI World Map
                </h2>
                <p style={{ fontFamily: "'IBM Plex Mono', monospace", fontSize: '11px', color: 'var(--chamber-text-dim)', marginBottom: '0' }}>
                  Every glowing dot is a ranked metro. Gold is the top tier, blue mid, green emerging.
                </p>
              </div>
              {mapPoints.data?.category && (
                <span style={{
                  fontFamily: "'IBM Plex Mono', monospace",
                  fontSize: '10px',
                  padding: '4px 8px',
                  borderRadius: 'var(--chamber-radius)',
                  background: 'var(--chamber-panel)',
                  color: 'var(--chamber-text-faint)',
                }}>
                  {mapPoints.data.category}
                </span>
              )}
            </div>

            <div style={{ borderRadius: 'var(--chamber-radius)', overflow: 'hidden', border: '1px solid var(--chamber-panel-edge)', marginTop: '16px' }}>
              <CityRankingsGlobe
                points={mapPoints.data?.points ?? []}
                onSelect={onSelectMetro}
                height={480}
                visibleTiers={visibleTiers.length > 0 ? visibleTiers : undefined}
                isLoading={mapPoints.loading}
              />
            </div>

            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '24px', marginTop: '16px', fontFamily: "'IBM Plex Mono', monospace", fontSize: '10px', color: 'var(--chamber-text-faint)' }}>
              <button
                onClick={() => toggleTier('gold')}
                style={{
                  display: 'flex', alignItems: 'center', gap: '6px',
                  background: visibleTiers.includes('gold') ? 'rgba(255, 183, 77, 0.15)' : 'transparent',
                  border: visibleTiers.includes('gold') ? '1px solid rgba(255, 183, 77, 0.4)' : '1px solid transparent',
                  borderRadius: '6px', padding: '4px 8px', cursor: 'pointer',
                  color: visibleTiers.includes('gold') ? '#ffb74d' : 'var(--chamber-text-faint)',
                  transition: 'all 0.15s ease',
                }}
              >
                <span style={{ width: 10, height: 10, borderRadius: '50%', background: '#ffb74d', display: 'inline-block' }} />
                Gold tier (top)
              </button>
              <button
                onClick={() => toggleTier('blue')}
                style={{
                  display: 'flex', alignItems: 'center', gap: '6px',
                  background: visibleTiers.includes('blue') ? 'rgba(79, 195, 247, 0.15)' : 'transparent',
                  border: visibleTiers.includes('blue') ? '1px solid rgba(79, 195, 247, 0.4)' : '1px solid transparent',
                  borderRadius: '6px', padding: '4px 8px', cursor: 'pointer',
                  color: visibleTiers.includes('blue') ? '#4fc3f7' : 'var(--chamber-text-faint)',
                  transition: 'all 0.15s ease',
                }}
              >
                <span style={{ width: 10, height: 10, borderRadius: '50%', background: '#4fc3f7', display: 'inline-block' }} />
                Blue tier
              </button>
              <button
                onClick={() => toggleTier('green')}
                style={{
                  display: 'flex', alignItems: 'center', gap: '6px',
                  background: visibleTiers.includes('green') ? 'rgba(102, 187, 106, 0.15)' : 'transparent',
                  border: visibleTiers.includes('green') ? '1px solid rgba(102, 187, 106, 0.4)' : '1px solid transparent',
                  borderRadius: '6px', padding: '4px 8px', cursor: 'pointer',
                  color: visibleTiers.includes('green') ? '#66bb6a' : 'var(--chamber-text-faint)',
                  transition: 'all 0.15s ease',
                }}
              >
                <span style={{ width: 10, height: 10, borderRadius: '50%', background: '#66bb6a', display: 'inline-block' }} />
                Green tier
              </button>
            </div>
          </div>
        </Chamber>
      </section>

      {/* Stats Strip */}
      <section style={{ padding: '0 24px 48px', maxWidth: '1200px', margin: '0 auto' }}>
        <GaugeStrip
          gauges={[
            { value: leaderboard.data?.total_ranked ?? 0, label: 'Cities Ranked' },
            { value: states.data?.total_states ?? 0, label: 'States' },
            { value: leaderboard.data?.entries?.[0]?.score_per_capita?.toFixed(4) ?? '—', label: 'Top Per Capita' },
            { value: '1/hr', label: 'Update Freq' },
          ]}
        />
      </section>

      {/* Main Content Grid */}
      <section style={{ padding: '0 24px 48px', maxWidth: '1200px', margin: '0 auto' }}>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 360px', gap: '32px' }}>
          {/* Leaderboard */}
          <div>
            <Chamber animate={true}>
              <CornerBrace position="tl" />
              <CornerBrace position="br" />
              <AnnotationTag label={`${category.toUpperCase()} RANKING · 30D ACTIVITY`} position="tr" />

              {leaderboard.loading ? (
                <LeaderboardSkeleton />
              ) : leaderboard.error ? (
                <ErrorState message={leaderboard.error} onRetry={handleRetry} />
              ) : !leaderboard.data || leaderboard.data.entries.length === 0 ? (
                <EmptyState message="No cities ranked yet. Check back after the next hourly recompute." />
              ) : (
                <Leaderboard entries={leaderboard.data.entries} />
              )}
            </Chamber>
          </div>

          {/* Sidebar - Movers */}
          <aside style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
            <Chamber animate={true}>
              <CornerBrace position="tl" />
              <MoversSection
                title="Top Gainers"
                icon={<TrendingUp style={{ width: 16, height: 16, color: 'var(--chamber-status-ok)' }} />}
                entries={gainers.data?.entries ?? []}
                loading={gainers.loading}
                error={gainers.error}
                onRetry={handleRetry}
              />
            </Chamber>

            <Chamber animate={true}>
              <CornerBrace position="tl" />
              <MoversSection
                title="Top Losers"
                icon={<TrendingDown style={{ width: 16, height: 16, color: 'var(--chamber-accent)' }} />}
                entries={losers.data?.entries ?? []}
                loading={losers.loading}
                error={losers.error}
                onRetry={handleRetry}
              />
            </Chamber>

            <Chamber animate={true}>
              <CornerBrace position="tl" />
              <CornerBrace position="br" />
              <AnnotationTag label="HOW IT WORKS" position="tr" />
              <div style={{ padding: '20px', paddingTop: '32px' }}>
                <HowItWorks />
              </div>
            </Chamber>
          </aside>
        </div>
      </section>

      {/* State Rankings */}
      <section style={{ padding: '0 24px 48px', maxWidth: '1200px', margin: '0 auto' }}>
        <Chamber animate={true}>
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <AnnotationTag label="STATE RANKINGS · US" position="tr" />
          <div style={{ padding: '24px' }}>
            <h2 style={{ fontFamily: "'Space Grotesk', sans-serif", fontSize: '20px', fontWeight: 600, color: 'var(--chamber-text)', marginBottom: '8px' }}>
              Top US States
            </h2>
            <p style={{ fontFamily: "'IBM Plex Mono', monospace", fontSize: '11px', color: 'var(--chamber-text-dim)', marginBottom: '24px' }}>
              Rolled up from the metros below.
            </p>
            {states.loading ? (
              <StatesSkeleton />
            ) : states.error ? (
              <ErrorState message={states.error} onRetry={handleRetry} />
            ) : !states.data || states.data.entries.length === 0 ? (
              <EmptyState message="No US states have enough ranked metros yet." />
            ) : (
              <StatesTable entries={states.data.entries} />
            )}
          </div>
        </Chamber>
      </section>

      {/* CTA Section */}
      <section style={{ borderTop: '1px solid var(--chamber-panel-edge)', padding: '48px 24px', textAlign: 'center' }}>
        <div style={{ maxWidth: '600px', margin: '0 auto' }}>
          <Sparkles style={{ width: 32, height: 32, color: 'var(--chamber-status-ok)', margin: '0 auto 16px' }} />
          <h2 style={{ fontFamily: "'Space Grotesk', sans-serif", fontSize: '28px', fontWeight: 700, color: 'var(--chamber-text)', marginBottom: '12px' }}>
            Your city should be on this list.
          </h2>
          <p style={{ color: 'var(--chamber-text-dim)', marginBottom: '24px', lineHeight: 1.6 }}>
            Set your location in profile settings and watch your city climb the leaderboard.
            Free, open to every builder.
          </p>
          <SealedButton onClick={() => window.location.href = `${APP_DASHBOARD_ORIGIN}/settings/profile`}>
            Get on the leaderboard →
          </SealedButton>
        </div>
      </section>

      {/* Footer */}
      <footer style={{ borderTop: '1px solid var(--chamber-panel-edge)', padding: '24px', textAlign: 'center' }}>
        <div style={{ fontFamily: "'IBM Plex Mono', monospace", fontSize: '11px', color: 'var(--chamber-text-faint)' }}>
          Updated {leaderboard.data?.period_end ? new Date(leaderboard.data.period_end).toLocaleString() : '—'} ·
          {' '}{leaderboard.data?.total_ranked ?? 0} cities ranked
          {leaderboard.data?.cache_hit && ' · cached'} ·
          {' '}{states.data?.total_states ?? 0} states ranked
          {' · '}{leaderboard.data?.category ?? 'composite'}
        </div>
      </footer>
    </div>
  );
}

function ErrorState({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <div style={{ padding: '48px', textAlign: 'center', color: 'var(--chamber-accent)' }}>
      <AlertCircle style={{ width: 24, height: 24, margin: '0 auto 12px', opacity: 0.7 }} />
      <p style={{ fontFamily: "'IBM Plex Mono', monospace", fontSize: '12px', marginBottom: '12px' }}>
        {message}
      </p>
      <button
        onClick={onRetry}
        style={{
          fontFamily: "'IBM Plex Mono', monospace",
          fontSize: '11px',
          padding: '6px 12px',
          borderRadius: 'var(--chamber-radius)',
          border: '1px solid var(--chamber-panel-edge)',
          background: 'transparent',
          color: 'var(--chamber-text-dim)',
          cursor: 'pointer',
        }}
      >
        Retry
      </button>
    </div>
  );
}

function EmptyState({ message }: { message: string }) {
  return (
    <div style={{ padding: '48px', textAlign: 'center', color: 'var(--chamber-text-dim)' }}>
      <p style={{ fontFamily: "'IBM Plex Mono', monospace", fontSize: '12px' }}>
        {message}
      </p>
    </div>
  );
}

function Leaderboard({ entries }: { entries: CityRankingEntry[] }) {
  return (
    <div style={{ overflow: 'hidden' }}>
      <table className="rankings-table">
        <thead>
          <tr>
            <th style={{ width: '60px' }}>#</th>
            <th>Metro</th>
            <th style={{ textAlign: 'right' }}>Per Capita</th>
            <th style={{ textAlign: 'right', display: 'none' }} className="sm:table-cell">Raw</th>
            <th style={{ textAlign: 'right', display: 'none' }} className="sm:table-cell">Active</th>
            <th style={{ textAlign: 'right', display: 'none' }} className="md:table-cell">Execs / 30d</th>
          </tr>
        </thead>
        <tbody>
          {entries.map((e) => (
            <tr key={e.metro_slug}>
              <td>
                <RankDisplay rank={e.rank} delta={e.rank_delta} />
              </td>
              <td>
                <div style={{ fontWeight: 500, color: 'var(--chamber-text)' }}>{e.metro_name}</div>
                <div style={{ fontSize: '11px', color: 'var(--chamber-text-faint)' }}>
                  {e.country_code} · pop {Intl.NumberFormat().format(e.population)}
                </div>
              </td>
              <td style={{ textAlign: 'right' }}>
                <span className="rankings-table__score">{e.score_per_capita.toFixed(4)}</span>
              </td>
              <td style={{ textAlign: 'right', display: 'none' }} className="sm:table-cell">
                <span className="rankings-table__secondary">{e.score_raw.toFixed(3)}</span>
              </td>
              <td style={{ textAlign: 'right', display: 'none' }} className="sm:table-cell">
                <span className="rankings-table__secondary">{e.active_users}</span>
              </td>
              <td style={{ textAlign: 'right', display: 'none' }} className="md:table-cell">
                <span className="rankings-table__secondary">{e.executions_30d.toLocaleString()}</span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

interface MoversSectionProps {
  title: string;
  icon: React.ReactNode;
  entries: CityRankingEntry[];
  loading: boolean;
  error: string | null;
  onRetry: () => void;
}

function MoversSection({ title, icon, entries, loading, error, onRetry }: MoversSectionProps) {
  return (
    <div style={{ padding: '20px' }}>
      <h3 style={{ fontFamily: "'IBM Plex Mono', monospace", fontSize: '10px', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.08em', color: 'var(--chamber-text-dim)', marginBottom: '16px', display: 'flex', alignItems: 'center', gap: '8px' }}>
        {icon}
        {title}
      </h3>
      {loading ? (
        <MoversSkeleton />
      ) : error ? (
        <div style={{ textAlign: 'center', padding: '16px 0' }}>
          <p style={{ fontSize: '11px', color: 'var(--chamber-accent)', marginBottom: '8px' }}>Failed to load</p>
          <button
            onClick={onRetry}
            style={{
              fontSize: '10px',
              padding: '4px 8px',
              borderRadius: 'var(--chamber-radius)',
              border: '1px solid var(--chamber-panel-edge)',
              background: 'transparent',
              color: 'var(--chamber-text-dim)',
              cursor: 'pointer',
            }}
          >
            Retry
          </button>
        </div>
      ) : entries.length === 0 ? (
        <p style={{ fontSize: '12px', color: 'var(--chamber-text-faint)', padding: '16px 0' }}>
          No movers yet — check back after the next hourly recompute.
        </p>
      ) : (
        <ol style={{ listStyle: 'none', padding: 0, margin: 0 }}>
          {entries.slice(0, 5).map((e) => (
            <li key={e.metro_slug} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '10px 0', borderBottom: '1px solid var(--chamber-panel-edge)' }}>
              <span style={{ fontWeight: 500, color: 'var(--chamber-text)' }}>{e.metro_name}</span>
              <span style={{ fontFamily: "'IBM Plex Mono', monospace", fontSize: '12px', color: e.rank_delta > 0 ? 'var(--chamber-status-ok)' : 'var(--chamber-accent)' }}>
                {e.rank_delta > 0 ? '▲' : '▼'} {Math.abs(e.rank_delta)}
              </span>
            </li>
          ))}
        </ol>
      )}
    </div>
  );
}

function StatesTable({ entries }: { entries: StateRankingEntry[] }) {
  return (
    <table className="rankings-table">
      <thead>
        <tr>
          <th style={{ width: '60px' }}>#</th>
          <th>State</th>
          <th style={{ textAlign: 'center', display: 'none' }} className="sm:table-cell">Metros</th>
          <th style={{ textAlign: 'right' }}>Per Capita</th>
          <th style={{ textAlign: 'right', display: 'none' }} className="md:table-cell">Active</th>
          <th style={{ textAlign: 'right', display: 'none' }} className="md:table-cell">Execs / 30d</th>
        </tr>
      </thead>
      <tbody>
        {entries.map((e) => (
          <tr key={`${e.country_code}-${e.state_code}`}>
            <td>
              <span style={{ fontFamily: "'IBM Plex Mono', monospace", fontSize: '16px', fontWeight: 600 }}>{e.rank}</span>
            </td>
            <td>
              <div style={{ fontWeight: 500, color: 'var(--chamber-text)' }}>{e.state_name}</div>
              <div style={{ fontSize: '11px', color: 'var(--chamber-text-faint)' }}>
                {e.state_code} · pop {Intl.NumberFormat().format(e.population)}
              </div>
            </td>
            <td style={{ textAlign: 'center', display: 'none' }} className="sm:table-cell">
              <span style={{ fontFamily: "'IBM Plex Mono', monospace", fontSize: '11px', padding: '2px 8px', borderRadius: 'var(--chamber-radius)', background: 'var(--chamber-bg)' }}>
                {e.ranked_metros}/{e.metro_count}
              </span>
            </td>
            <td style={{ textAlign: 'right' }}>
              <span className="rankings-table__score">{e.score_per_capita.toFixed(4)}</span>
            </td>
            <td style={{ textAlign: 'right', display: 'none' }} className="md:table-cell">
              <span className="rankings-table__secondary">{e.active_users}</span>
            </td>
            <td style={{ textAlign: 'right', display: 'none' }} className="md:table-cell">
              <span className="rankings-table__secondary">{e.executions_30d.toLocaleString()}</span>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function HowItWorks() {
  return (
    <div style={{ padding: '20px' }}>
      <ul style={{ listStyle: 'none', padding: 0, margin: 0, fontSize: '12px', color: 'var(--chamber-text-dim)', lineHeight: 1.8 }}>
        <li>Composite score = log-scaled activity × weights:</li>
        <li style={{ paddingLeft: '12px', color: 'var(--chamber-text-faint)' }}>· 30% active users (30d)</li>
        <li style={{ paddingLeft: '12px', color: 'var(--chamber-text-faint)' }}>· 25% deployments (30d)</li>
        <li style={{ paddingLeft: '12px', color: 'var(--chamber-text-faint)' }}>· 20% function executions (30d)</li>
        <li style={{ paddingLeft: '12px', color: 'var(--chamber-text-faint)' }}>· 15% founder / referral earnings</li>
        <li style={{ paddingLeft: '12px', color: 'var(--chamber-text-faint)' }}>· 10% new user growth</li>
        <li style={{ marginTop: '12px' }}>Per-capita = raw × 100k / metro population.</li>
        <li>Privacy: metros with fewer than 5 active users are hidden. Users can opt out any time.</li>
      </ul>
    </div>
  );
}

function LeaderboardSkeleton() {
  return (
    <div style={{ padding: '24px' }}>
      {Array.from({ length: 8 }).map((_, i) => (
        <div key={i} style={{ height: '48px', marginBottom: '8px', borderRadius: 'var(--chamber-radius)', background: 'var(--chamber-bg)', opacity: 0.5 }} />
      ))}
    </div>
  );
}

function MoversSkeleton() {
  return (
    <div>
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} style={{ display: 'flex', justifyContent: 'space-between', padding: '10px 0', borderBottom: '1px solid var(--chamber-panel-edge)' }}>
          <div style={{ height: 16, width: '60%', borderRadius: 4, background: 'var(--chamber-bg)', opacity: 0.5 }} />
          <div style={{ height: 16, width: '15%', borderRadius: 4, background: 'var(--chamber-bg)', opacity: 0.5 }} />
        </div>
      ))}
    </div>
  );
}

function StatesSkeleton() {
  return (
    <div style={{ padding: '16px 0' }}>
      {Array.from({ length: 10 }).map((_, i) => (
        <div key={i} style={{ display: 'flex', alignItems: 'center', gap: '16px', padding: '12px 0', borderBottom: '1px solid var(--chamber-panel-edge)' }}>
          <div style={{ height: 16, width: '30px', borderRadius: 4, background: 'var(--chamber-bg)', opacity: 0.5 }} />
          <div style={{ flex: 1, height: 16, borderRadius: 4, background: 'var(--chamber-bg)', opacity: 0.5 }} />
          <div style={{ height: 16, width: '80px', borderRadius: 4, background: 'var(--chamber-bg)', opacity: 0.5 }} />
        </div>
      ))}
    </div>
  );
}
