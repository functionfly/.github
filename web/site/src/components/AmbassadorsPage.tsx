import { useEffect, useState } from 'react';
import { Award, MapPin } from 'lucide-react';
import {
  Chamber,
  CornerBrace,
  GaugeStrip,
  AnnotationTag,
  TrustSeal,
  ReducedMotionGate,
} from './rankings';

interface Ambassador {
  metro_id: number;
  metro_slug: string;
  metro_name: string;
  country_code: string;
  state_code?: string;
  city_slug?: string;
  user_id: string;
  name: string;
  username?: string;
  profile_number?: number;
  promoted_at: string;
  source: string;
}

interface List {
  total: number;
  entries: Ambassador[];
  privacy_min_active_users: number;
}

interface CountryOption {
  code: string;
  name: string;
}

interface CountriesResponse {
  countries: CountryOption[];
}

const API_BASE =
  (typeof import.meta !== 'undefined' && (import.meta as any).env?.PUBLIC_API_URL) ||
  'https://api.functionfly.com';

export default function AmbassadorsPage() {
  const [countries, setCountries] = useState<CountryOption[]>([]);
  const [country, setCountry] = useState('');
  const [data, setData] = useState<List | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    fetch(`${API_BASE}/v1/city-rankings/countries`)
      .then((r) => (r.ok ? r.json() : Promise.reject(r.statusText)))
      .then((d: CountriesResponse) => {
        if (!cancelled) {
          setCountries([{ code: '', name: 'World' }, ...d.countries]);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setCountries([{ code: '', name: 'World' }]);
        }
      });
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    const params = new URLSearchParams();
    if (country) params.set('country', country);
    params.set('limit', '300');
    fetch(`${API_BASE}/v1/city-rankings/ambassadors?${params.toString()}`)
      .then((r) => (r.ok ? r.json() : Promise.reject(r.statusText)))
      .then((d: List) => {
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
            <div className="ambassadors-header__badge">
              <Award />
              <span>FunctionFly City Ambassadors™</span>
            </div>
            <h1 className="ambassadors-header__title">
              Top builder in every city
            </h1>
            <p className="ambassadors-header__subtitle">
              Promoted automatically every hour from the live leaderboard. 
              Privacy-thresholded (k≥{data?.privacy_min_active_users ?? 5} active builders).
            </p>
          </div>
        </div>
      </header>

      <main style={{ padding: '32px 24px', maxWidth: '1200px', margin: '0 auto' }}>
        {/* Country Filter */}
        <div className="country-filter" style={{ marginBottom: '24px' }}>
          <span className="country-filter__label">
            <MapPin style={{ width: 12, height: 12, display: 'inline', marginRight: '4px' }} />
            Filter by country:
          </span>
          {countries.map((c) => (
            <button
              key={c.code}
              onClick={() => setCountry(c.code)}
              className={`country-btn ${country === c.code ? 'country-btn--active' : ''}`}
            >
              {c.name}
            </button>
          ))}
        </div>

        {/* Stats Strip */}
        <GaugeStrip
          gauges={[
            { value: data?.total ?? 0, label: 'Ambassadors' },
            { value: data?.entries?.[0] ? new Date(data.entries[0].promoted_at).toLocaleDateString() : '—', label: 'Last Promoted' },
            { value: '1/hr', label: 'Update Freq' },
            { value: `${data?.privacy_min_active_users ?? 5}+`, label: 'Min Builders' },
          ]}
          style={{ marginBottom: '24px' }}
        />

        {error && (
          <Chamber style={{ padding: '16px 20px', marginBottom: '24px' }}>
            <p style={{ color: 'var(--chamber-accent)', fontFamily: "'IBM Plex Mono', monospace", fontSize: '12px' }}>
              Failed to load ambassadors: {error}
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
            <AnnotationTag label="CITY AMBASSADORS · AUTO-PROMOTED" position="tr" />
            <ol className="ambassador-list">
              {data.entries.map((row) => (
                <AmbassadorItem key={`${row.metro_id}-${row.user_id}`} ambassador={row} />
              ))}
            </ol>
          </Chamber>
        ) : (
          <Chamber animate={true}>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div className="ambassadors-empty">
              <Award className="ambassadors-empty__icon" />
              <h2 className="ambassadors-empty__title">No ambassadors yet</h2>
              <p className="ambassadors-empty__text">
                Keep building and you might be the first.
              </p>
            </div>
          </Chamber>
        )}

        <div className="ambassadors-footer">
          Updated hourly · {data?.total ?? 0} ambassadors
        </div>
      </main>
    </div>
  );
}

function AmbassadorItem({ ambassador }: { ambassador: Ambassador }) {
  return (
    <li className="ambassador-item">
      <div className="ambassador-item__seal">
        <ReducedMotionGate
          fallback={<Award style={{ width: 20, height: 20, color: 'var(--chamber-status-ok)' }} />}
        >
          <TrustSeal size="sm" showShimmer={false} onHover={false} />
        </ReducedMotionGate>
      </div>

      <div className="ambassador-item__info">
        <p className="ambassador-item__name">
          {ambassador.name}
          {ambassador.username && (
            <span className="ambassador-item__username">@{ambassador.username}</span>
          )}
        </p>
        <div className="ambassador-item__location">
          <MapPin style={{ width: 12, height: 12 }} />
          <span>{ambassador.metro_name}</span>
          {ambassador.state_code && (
            <>
              <span className="ambassador-item__location-sep">·</span>
              <span>{ambassador.state_code}</span>
            </>
          )}
          <span className="ambassador-item__location-sep">·</span>
          <span>{ambassador.country_code}</span>
        </div>
      </div>

      <div className="ambassador-item__meta">
        <p className="ambassador-item__date">
          {new Date(ambassador.promoted_at).toLocaleDateString()}
        </p>
        <p className="ambassador-item__source">{ambassador.source}</p>
      </div>
    </li>
  );
}
