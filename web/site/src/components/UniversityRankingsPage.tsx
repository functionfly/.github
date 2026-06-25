import { useEffect, useState } from 'react';
import { GraduationCap, TrendingUp, TrendingDown } from 'lucide-react';
import {
  Chamber,
  CornerBrace,
  GaugeStrip,
  AnnotationTag,
} from './rankings';

interface UniversityRanking {
  rank: number;
  rank_delta: number;
  slug: string;
  name: string;
  short_name?: string;
  country_code: string;
  state_code?: string;
  student_count?: number;
  score_per_capita: number;
  active_users: number;
  executions_30d: number;
}

interface Leaderboard {
  period_end: string;
  total_ranked: number;
  entries: UniversityRanking[];
  privacy_min_active_users: number;
}

const COUNTRY_OPTIONS = [
  { code: '', label: 'World' },
  { code: 'US', label: 'United States' },
  { code: 'GB', label: 'United Kingdom' },
  { code: 'CA', label: 'Canada' },
  { code: 'CN', label: 'China' },
  { code: 'JP', label: 'Japan' },
  { code: 'SG', label: 'Singapore' },
  { code: 'AU', label: 'Australia' },
];

const API_BASE =
  (typeof window !== 'undefined' && (window as any).__FUNCTIONFLY_API__) ||
  (typeof import.meta !== 'undefined' && (import.meta as any).env?.PUBLIC_API_URL) ||
  'https://api.functionfly.com';

export default function UniversityRankingsPage() {
  const [country, setCountry] = useState('');
  const [data, setData] = useState<Leaderboard | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    const params = new URLSearchParams({ category: 'composite', limit: '100' });
    if (country) params.set('country', country);
    fetch(`${API_BASE}/v1/university-rankings?${params.toString()}`)
      .then((r) => (r.ok ? r.json() : Promise.reject(r.statusText)))
      .then((d: Leaderboard) => {
        if (!cancelled) {
          setData(d);
          setLoading(false);
        }
      })
      .catch((e) => {
        if (!cancelled) {
          setError(String(e));
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [country]);

  return (
    <div style={{ background: 'var(--chamber-bg)', minHeight: '100vh' }}>
      {/* Header */}
      <header style={{ background: 'var(--chamber-panel)', borderBottom: '1px solid var(--chamber-panel-edge)' }}>
        <div style={{ position: 'relative', padding: '48px 24px', textAlign: 'center', maxWidth: '1200px', margin: '0 auto' }}>
          <div style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0, background: 'radial-gradient(ellipse 80% 60% at 20% 0%, rgba(143, 255, 208, 0.08) 0%, transparent 70%)', pointerEvents: 'none' }} />
          <div style={{ position: 'relative', zIndex: 1 }}>
            <div className="university-header__badge">
              <GraduationCap />
              <span>FunctionFly University Rankings™</span>
            </div>
            <h1 className="university-header__title">
              Top schools by AI builder activity
            </h1>
            <p className="university-header__subtitle">
              Per-capita scoring, so a small CS lab can outrank a state flagship. 
              Privacy-thresholded (k≥{data?.privacy_min_active_users ?? 5} active builders) and updated hourly.
            </p>
          </div>
        </div>
      </header>

      <main style={{ padding: '32px 24px', maxWidth: '1200px', margin: '0 auto' }}>
        {/* Country Filter */}
        <div className="country-filter">
          {COUNTRY_OPTIONS.map((c) => (
            <button
              key={c.code}
              onClick={() => setCountry(c.code)}
              className={`country-btn ${country === c.code ? 'country-btn--active' : ''}`}
            >
              {c.label}
            </button>
          ))}
        </div>

        {/* Stats Strip */}
        <GaugeStrip
          gauges={[
            { value: data?.total_ranked ?? 0, label: 'Universities Ranked' },
            { value: data?.entries?.[0]?.score_per_capita?.toFixed(4) ?? '—', label: 'Top Per Capita' },
            { value: data?.entries?.[0]?.active_users ?? 0, label: 'Top Builders' },
            { value: '1/hr', label: 'Update Freq' },
          ]}
          style={{ marginBottom: '24px' }}
        />

        {error && (
          <Chamber style={{ padding: '16px 20px', marginBottom: '24px' }}>
            <p style={{ color: 'var(--chamber-accent)', fontFamily: "'IBM Plex Mono', monospace", fontSize: '12px' }}>
              Failed to load leaderboard: {error}
            </p>
        </Chamber>
        )}

        {loading ? (
          <Chamber ribs={true} animate={true}>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div style={{ padding: '24px' }}>
              {Array.from({ length: 8 }).map((_, i) => (
                <div
                  key={i}
                  style={{ height: '64px', marginBottom: '8px', borderRadius: 'var(--chamber-radius)', background: 'var(--chamber-bg)', opacity: 0.5 }}
                />
              ))}
            </div>
          </Chamber>
        ) : data && data.entries.length > 0 ? (
          <Chamber ribs={true} animate={true}>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <AnnotationTag label="TOP UNIVERSITIES · PER CAPITA SCORING" position="tr" />
            <ol className="university-list">
              {data.entries.map((row) => (
                <UniversityItem key={row.slug} university={row} />
              ))}
            </ol>
          </Chamber>
        ) : (
          <Chamber animate={true}>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div className="university-empty">
              <GraduationCap className="university-empty__icon" />
              <h2 className="university-empty__title">No universities ranked yet</h2>
              <p className="university-empty__text">
                Invite builders from your school to get on the board.
              </p>
            </div>
          </Chamber>
        )}

        <div className="university-footer">
          Updated hourly · {data?.total_ranked ?? 0} universities ranked
        </div>
      </main>
    </div>
  );
}

function UniversityItem({ university }: { university: UniversityRanking }) {
  const isTop3 = university.rank <= 3;
  const delta = university.rank_delta;

  return (
    <li className="university-item">
      <span className={`university-item__rank ${isTop3 ? 'university-item__rank--top3' : ''}`}>
        #{university.rank}
      </span>

      <div className="university-item__info">
        <p className="university-item__name">
          {university.short_name || university.name}
        </p>
        <div className="university-item__meta">
          <span>{university.country_code}</span>
          {university.state_code && (
            <>
              <span className="university-item__meta-sep">·</span>
              <span>{university.state_code}</span>
            </>
          )}
          {university.student_count && (
            <>
              <span className="university-item__meta-sep">·</span>
              <span>{university.student_count.toLocaleString()} students</span>
            </>
          )}
        </div>
      </div>

      <div className="university-item__stat">
        <p className="university-item__stat-value">
          {university.score_per_capita.toFixed(4)}
        </p>
        <p className="university-item__stat-label">per capita</p>
      </div>

      <div className="university-item__builders">
        <p className="university-item__builders-value">
          {university.active_users.toLocaleString()}
        </p>
        <p className="university-item__builders-label">builders</p>
      </div>

      <div className="university-item__delta">
        {delta > 0 && (
          <span className="university-item__delta--up">
            <TrendingUp className="university-item__delta-icon" />
          </span>
        )}
        {delta < 0 && (
          <span className="university-item__delta--down">
            <TrendingDown className="university-item__delta-icon" />
          </span>
        )}
        {delta === 0 && (
          <span style={{ color: 'var(--chamber-text-faint)', fontFamily: "'IBM Plex Mono', monospace", fontSize: '12px' }}>—</span>
        )}
      </div>
    </li>
  );
}
