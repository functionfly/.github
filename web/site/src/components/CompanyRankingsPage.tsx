import { useEffect, useState } from 'react';
import { Building2, TrendingUp, TrendingDown, Zap } from 'lucide-react';
import {
  Chamber,
  CornerBrace,
  GaugeStrip,
  AnnotationTag,
} from './rankings';

interface CompanyRanking {
  rank: number;
  previous_rank?: number;
  rank_delta: number;
  company_id: number;
  slug: string;
  name: string;
  country_code: string;
  city_slug?: string;
  employee_count: number;
  score_raw: number;
  score_per_capita: number;
  active_users: number;
  deployments: number;
  executions_30d: number;
  revenue_cents: number;
  new_users_30d: number;
  category: string;
}

interface LeaderboardResponse {
  total_ranked: number;
  entries: CompanyRanking[];
  country?: string;
  category: string;
  cache_hit: boolean;
}

const COUNTRY_OPTIONS = [
  { code: '', label: 'World' },
  { code: 'US', label: 'United States' },
  { code: 'GB', label: 'United Kingdom' },
  { code: 'CA', label: 'Canada' },
  { code: 'IN', label: 'India' },
  { code: 'DE', label: 'Germany' },
  { code: 'BR', label: 'Brazil' },
  { code: 'JP', label: 'Japan' },
  { code: 'CN', label: 'China' },
  { code: 'SG', label: 'Singapore' },
  { code: 'AU', label: 'Australia' },
];

const CATEGORY_OPTIONS = [
  { slug: 'composite', label: 'All Activity' },
  { slug: 'agents', label: 'Agent Capital' },
  { slug: 'automation', label: 'Automation Capital' },
  { slug: 'startups', label: 'Startup Capital' },
  { slug: 'open_source', label: 'Open Source Capital' },
  { slug: 'robotics', label: 'Robotics Capital' },
];

const API_BASE =
  (typeof window !== 'undefined' && (window as any).__FUNCTIONFLY_API__) ||
  (typeof import.meta !== 'undefined' && (import.meta as any).env?.PUBLIC_API_URL) ||
  'https://api.functionfly.com';

export default function CompanyRankingsPage() {
  const [country, setCountry] = useState('');
  const [category, setCategory] = useState('composite');
  const [data, setData] = useState<LeaderboardResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    const params = new URLSearchParams({ category, limit: '100' });
    if (country) params.set('country', country);
    fetch(`${API_BASE}/v1/company-rankings?${params.toString()}`)
      .then((r) => (r.ok ? r.json() : Promise.reject(r.statusText)))
      .then((d: LeaderboardResponse) => {
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
  }, [country, category]);

  return (
    <div style={{ background: 'var(--chamber-bg)', minHeight: '100vh' }}>
      {/* Header */}
      <header style={{ background: 'var(--chamber-panel)', borderBottom: '1px solid var(--chamber-panel-edge)' }}>
        <div style={{ position: 'relative', padding: '48px 24px', textAlign: 'center', maxWidth: '1200px', margin: '0 auto' }}>
          <div style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0, background: 'radial-gradient(ellipse 80% 60% at 20% 0%, rgba(143, 255, 208, 0.08) 0%, transparent 70%)', pointerEvents: 'none' }} />
          <div style={{ position: 'relative', zIndex: 1 }}>
            <div className="company-header__badge">
              <Building2 />
              <span>FunctionFly Company Rankings™</span>
            </div>
            <h1 className="company-header__title">
              Top businesses by AI activity
            </h1>
            <p className="company-header__subtitle">
              See which companies are leading in AI agent deployments, automation workflows, 
              and builder productivity. Updated hourly.
            </p>
          </div>
        </div>
      </header>

      <main style={{ padding: '32px 24px', maxWidth: '1200px', margin: '0 auto' }}>
        {/* Filters */}
        <div className="filter-section">
          <div className="filter-row">
            {COUNTRY_OPTIONS.map((c) => (
              <button
                key={c.code}
                onClick={() => setCountry(c.code)}
                className={`filter-btn ${country === c.code ? 'filter-btn--active' : ''}`}
              >
                {c.label}
              </button>
            ))}
          </div>
        </div>

        <div className="filter-section">
          <div className="filter-row">
            {CATEGORY_OPTIONS.map((cat) => (
              <button
                key={cat.slug}
                onClick={() => setCategory(cat.slug)}
                className={`filter-btn ${category === cat.slug ? 'filter-btn--active' : ''}`}
              >
                {cat.label}
              </button>
            ))}
          </div>
        </div>

        {/* Stats Strip */}
        <GaugeStrip
          gauges={[
            { value: data?.total_ranked ?? 0, label: 'Companies Ranked' },
            { value: data?.entries?.[0]?.score_per_capita?.toFixed(4) ?? '—', label: 'Top Per Capita' },
            { value: data?.entries?.[0]?.deployments?.toLocaleString() ?? 0, label: 'Top Deployments' },
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
                  style={{ height: '56px', marginBottom: '8px', borderRadius: 'var(--chamber-radius)', background: 'var(--chamber-bg)', opacity: 0.5 }}
                />
              ))}
            </div>
          </Chamber>
        ) : data && data.entries.length > 0 ? (
          <Chamber ribs={true} animate={true}>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <AnnotationTag label={`COMPANY RANKINGS · ${category.toUpperCase()}`} position="tr" />
            <div style={{ overflowX: 'auto' }}>
              <table className="company-table">
                <thead>
                  <tr>
                    <th style={{ width: '60px' }}>#</th>
                    <th>Company</th>
                    <th style={{ textAlign: 'center' }}>Category</th>
                    <th style={{ textAlign: 'right' }}>Employees</th>
                    <th style={{ textAlign: 'right' }}>Per Capita</th>
                    <th style={{ textAlign: 'right' }}>Deployments</th>
                    <th style={{ textAlign: 'right' }}>Execs / 30d</th>
                    <th style={{ textAlign: 'right' }}>Delta</th>
                  </tr>
                </thead>
                <tbody>
                  {data.entries.map((row) => (
                    <tr key={row.slug}>
                      <td>
                        <span className="company-table__rank">#{row.rank}</span>
                      </td>
                      <td>
                        <div className="company-table__name">{row.name}</div>
                        <div className="company-table__meta">
                          {row.country_code}
                          {row.city_slug && ` · ${row.city_slug}`}
                        </div>
                      </td>
                      <td style={{ textAlign: 'center' }}>
                        <span className={`company-table__category ${row.category === 'composite' ? 'company-table__category--composite' : ''}`}>
                          <Zap />
                          {row.category}
                        </span>
                      </td>
                      <td style={{ textAlign: 'right' }}>
                        <span className="company-table__number">
                          {row.employee_count > 0 ? row.employee_count.toLocaleString() : '—'}
                        </span>
                      </td>
                      <td style={{ textAlign: 'right' }}>
                        <span className="company-table__score">
                          {row.score_per_capita.toFixed(4)}
                        </span>
                      </td>
                      <td style={{ textAlign: 'right' }}>
                        <span className="company-table__secondary">
                          {row.deployments.toLocaleString()}
                        </span>
                      </td>
                      <td style={{ textAlign: 'right' }}>
                        <span className="company-table__secondary">
                          {row.executions_30d.toLocaleString()}
                        </span>
                      </td>
                      <td style={{ textAlign: 'right' }}>
                        <DeltaBadge delta={row.rank_delta} />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Chamber>
        ) : (
          <Chamber animate={true}>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div className="company-empty">
              <Building2 className="company-empty__icon" />
              <h2 className="company-empty__title">No companies ranked yet</h2>
              <p className="company-empty__text">
                Businesses can join by having their team use FunctionFly.
              </p>
            </div>
          </Chamber>
        )}

        <div className="company-footer">
          Updated hourly · {data?.total_ranked ?? 0} companies ranked
          {data?.category && data.category !== 'composite' && ` · ${data.category}`}
        </div>
      </main>
    </div>
  );
}

function DeltaBadge({ delta }: { delta: number }) {
  if (delta === 0) {
    return <span style={{ fontFamily: "'IBM Plex Mono', monospace", fontSize: '12px', color: 'var(--chamber-text-faint)' }}>—</span>;
  }

  if (delta > 0) {
    return (
      <span className="delta-badge delta-badge--up">
        <TrendingUp className="delta-badge__icon" />
        {delta}
      </span>
    );
  }

  return (
    <span className="delta-badge delta-badge--down">
      <TrendingDown className="delta-badge__icon" />
      {Math.abs(delta)}
    </span>
  );
}
