import { useEffect, useState } from 'react';
import { Trophy, Swords } from 'lucide-react';
import {
  Chamber,
  CornerBrace,
  StatusPill,
  AnnotationTag,
  SealedButton,
  TrustSeal,
  ReducedMotionGate,
} from './rankings';
import type { War, WarsResponse, ChampionsResponse } from './cityWarTypes';

const API_BASE = (import.meta as any).env?.PUBLIC_API_URL || 'https://api.functionfly.com';

export default function CityWarsPage() {
  const [activeWar, setActiveWar] = useState<War | null>(null);
  const [champions, setChampions] = useState<WarChampion[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      fetch(`${API_BASE}/v1/city-wars/latest`).then(r => r.json()),
      fetch(`${API_BASE}/v1/city-wars/champions`).then(r => r.json()),
    ])
      .then(([warData, champData]) => {
        const warResp = warData as WarResponse;
        const champResp = champData as ChampionsResponse;
        if (warResp.war) setActiveWar(warResp.war);
        setChampions(champResp.champions || []);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', minHeight: '50vh' }}>
        <div style={{ width: 32, height: 32, borderRadius: '50%', border: '2px solid var(--chamber-panel-edge)', borderTopColor: 'var(--chamber-status-ok)', animation: 'spin 1s linear infinite' }} />
        <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
      </div>
    );
  }

  return (
    <div style={{ background: 'var(--chamber-bg)', minHeight: '100vh' }}>
      {/* Header */}
      <header className="city-wars-header" style={{ background: 'var(--chamber-panel)', borderBottom: '1px solid var(--chamber-panel-edge)' }}>
        <div style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0, background: 'radial-gradient(ellipse 80% 60% at 20% 0%, rgba(143, 255, 208, 0.08) 0%, transparent 70%)', pointerEvents: 'none' }} />
        <div style={{ position: 'relative', zIndex: 1 }}>
          <div className="city-wars-header__eyebrow">
            <Swords />
            <span>Quarterly Bracket</span>
          </div>
          <h1 className="city-wars-header__title">City Wars</h1>
          <p className="city-wars-header__subtitle">
            Top 8 metros face off in a single-elimination bracket. Winner takes the crown.
          </p>
        </div>
      </header>

      <main style={{ padding: '48px 24px', maxWidth: '1200px', margin: '0 auto' }}>
        {activeWar ? (
          <div>
            <Chamber ribs={true} animate={true}>
              <CornerBrace position="tl" />
              <CornerBrace position="br" />
              <AnnotationTag label={`${activeWar.season} · BRACKET CHALLENGE`} position="tl" />
              <div style={{ padding: '24px' }}>
                <WarBracket war={activeWar} />
              </div>
            </Chamber>
          </div>
        ) : (
          <Chamber animate={true}>
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div className="war-empty">
              <Swords className="war-empty__icon" />
              <h2 className="war-empty__title">No Active War</h2>
              <p className="war-empty__text">
                The next City Wars bracket will begin soon. Follow the top metros to see when voting starts.
              </p>
            </div>
          </Chamber>
        )}

        {champions.length > 0 && (
          <section style={{ marginTop: '64px' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '24px' }}>
              <Trophy style={{ width: 24, height: 24, color: 'var(--chamber-status-ok)' }} />
              <h2 style={{ fontFamily: "'Space Grotesk', sans-serif", fontSize: '24px', fontWeight: 700, color: 'var(--chamber-text)', margin: 0 }}>
                Hall of Champions
              </h2>
            </div>
            <div className="hall-of-champions">
              {champions.map((c) => (
                <a
                  key={c.war_slug}
                  href={`/rankings/cities/${c.metro_slug}`}
                  className="champion-card"
                >
                  <div className="champion-card__season">{c.war_season}</div>
                  <div className="champion-card__name">{c.metro_name}</div>
                  <div className="champion-card__location">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />
                      <circle cx="12" cy="10" r="3" />
                    </svg>
                    {c.country_code}
                  </div>
                </a>
              ))}
            </div>
          </section>
        )}
      </main>
    </div>
  );
}

function WarBracket({ war }: { war: War }) {
  const getStatusType = (): 'live' | 'pending' | 'revoked' => {
    switch (war.status) {
      case 'active': return 'live';
      case 'scheduled': return 'pending';
      case 'complete': return 'revoked';
      default: return 'pending';
    }
  };

  return (
    <div>
      <div className="war-info">
        <div>
          <h2 className="war-info__title">{war.name}</h2>
          <div className="war-info__meta">
            <span className="war-info__meta-item">{war.season}</span>
            <span className="war-info__meta-item">·</span>
            <span className="war-info__meta-item">Round: {war.round}</span>
          </div>
        </div>
        <StatusPill status={getStatusType()} label={war.status} />
      </div>

      <div className="city-wars-bracket">
        {/* Quarterfinals */}
        <div className="bracket-round">
          <div className="bracket-round__label">Quarterfinals</div>
          {war.quarterfinals?.map(match => (
            <MatchCard key={match.id} match={match} />
          ))}
        </div>

        {/* Semifinals */}
        <div className="bracket-round">
          <div className="bracket-round__label">Semifinals</div>
          {war.semifinals?.length === 0 && (
            <div className="match-pending">
              <Swords className="match-pending__icon" />
              <p className="match-pending__text">Pending quarterfinal results</p>
            </div>
          )}
          {war.semifinals?.map(match => (
            <MatchCard key={match.id} match={match} />
          ))}
        </div>

        {/* Final */}
        <div className="bracket-round">
          <div className="bracket-round__label bracket-round__label--final">Final</div>
          {war.final ? (
            <MatchCard match={war.final} isFinal />
          ) : (
            <div className="match-pending">
              <Trophy className="match-pending__icon" />
              <p className="match-pending__text">Final match pending</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function MatchCard({ match, isFinal = false }: { match: WarMatch; isFinal?: boolean }) {
  const winnerA = match.winner_metro_id === match.metro_a_id;
  const winnerB = match.winner_metro_id === match.metro_b_id;
  const pending = !match.winner_metro_id;

  return (
    <div className={`match-card ${isFinal ? 'match-card--final' : ''}`}>
      <CompetitorRow
        name={match.metro_a_name}
        country={match.metro_a_country}
        score={match.score_a}
        builders={match.active_users_a}
        isWinner={winnerA}
        isLoser={!pending && !winnerA}
        isPending={pending}
        href={`/rankings/cities/${match.metro_a_slug}`}
      />
      <div className="match-card__vs">vs</div>
      <CompetitorRow
        name={match.metro_b_name}
        country={match.metro_b_country}
        score={match.score_b}
        builders={match.active_users_b}
        isWinner={winnerB}
        isLoser={!pending && !winnerB}
        isPending={pending}
        href={`/rankings/cities/${match.metro_b_slug}`}
      />
    </div>
  );
}

function CompetitorRow({
  name,
  country,
  score,
  builders,
  isWinner,
  isLoser,
  isPending,
  href,
}: {
  name: string;
  country: string;
  score: number;
  builders: number;
  isWinner: boolean;
  isLoser: boolean;
  isPending: boolean;
  href: string;
}) {
  let cardClass = 'match-card__competitor';
  if (isWinner) cardClass += ' match-card__competitor--winner';
  if (isLoser) cardClass += ' match-card__competitor--loser';

  let nameClass = 'match-card__name';
  if (isWinner) nameClass += ' match-card__name--winner';

  let scoreClass = 'match-card__score-value';
  if (isWinner) scoreClass += ' match-card__score-value--winner';

  return (
    <a href={href} className={cardClass}>
      <div className="match-card__info">
        <p className={nameClass}>{name}</p>
        <p className="match-card__country">{country}</p>
      </div>
      <div className="match-card__score">
        <p className={scoreClass}>
          {isPending ? '—' : score > 0 ? score.toFixed(4) : '—'}
        </p>
        <p className="match-card__score-label">
          {isPending ? '' : `${builders} builders`}
        </p>
      </div>
    </a>
  );
}

type WarChampion = ChampionsResponse['champions'][number];
