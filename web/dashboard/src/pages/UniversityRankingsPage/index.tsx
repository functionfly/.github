import { useState } from 'react';
import { GraduationCap, TrendingUp, TrendingDown, Globe, MapPin } from 'lucide-react';
import { Link } from 'react-router-dom';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { useUniversityLeaderboard, useMyUniversity, UNIVERSITY_CATEGORIES } from '@/hooks/useUniversityRankings';

const COUNTRY_OPTIONS: { code: string; label: string }[] = [
  { code: '', label: 'World' },
  { code: 'US', label: 'United States' },
  { code: 'GB', label: 'United Kingdom' },
  { code: 'CA', label: 'Canada' },
  { code: 'CN', label: 'China' },
  { code: 'JP', label: 'Japan' },
  { code: 'SG', label: 'Singapore' },
  { code: 'AU', label: 'Australia' },
];

export function UniversityRankingsPage() {
  const [country, setCountry] = useState('');
  const [category, setCategory] = useState('composite');
  const { data, isLoading, error } = useUniversityLeaderboard({ country, limit: 100, category });
  const { data: myUni } = useMyUniversity();

  return (
    <div className="container mx-auto max-w-7xl px-4 py-8">
      <header className="mb-8">
        <div className="flex items-center gap-3">
          <GraduationCap className="h-8 w-8 text-amber-500" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">University Rankings</h1>
            <p className="text-sm text-muted-foreground">
              Top schools by FunctionFly activity. Per-capita, so a small CS dept
              can outrank a large state U. Updated every hour.
            </p>
          </div>
        </div>
      </header>

      <div className="mb-6 flex flex-wrap items-center gap-3 text-sm">
        <div className="flex items-center gap-2">
          <MapPin className="h-4 w-4 text-muted-foreground" />
          <select
            value={country}
            onChange={(e) => setCountry(e.target.value)}
            className="rounded-md border border-white/10 bg-white/5 px-3 py-1.5 text-sm focus:border-amber-500 focus:outline-none"
          >
            {COUNTRY_OPTIONS.map((c) => (
              <option key={c.code} value={c.code}>
                {c.label}
              </option>
            ))}
          </select>
        </div>
        <div className="flex items-center gap-1 rounded-md border border-white/10 bg-white/5 p-1">
          {UNIVERSITY_CATEGORIES.map((cat) => (
            <button
              key={cat.slug}
              onClick={() => setCategory(cat.slug)}
              className={`rounded px-2 py-1 text-xs transition-colors ${
                category === cat.slug
                  ? 'bg-amber-500 text-black font-medium'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              {cat.label}
            </button>
          ))}
        </div>
        <span className="text-xs text-muted-foreground">
          {data && `${data.total_ranked} ranked · k≥${data.privacy_min_active_users} · ${data.category}`}
        </span>
      </div>

      {myUni?.university && (
        <Card className="mb-6 border-amber-500/30 bg-amber-500/5">
          <CardContent className="flex items-center justify-between gap-4 py-3">
            <div>
              <p className="text-xs uppercase tracking-wider text-muted-foreground">
                Your university
              </p>
              <p className="text-base font-semibold">
                {myUni.university.name}
                {myUni.ranking && (
                  <span className="ml-2 text-sm font-normal text-muted-foreground">
                    Rank #{myUni.ranking.rank}
                  </span>
                )}
              </p>
            </div>
            <Link
              to={`/university-rankings/${myUni.university.slug}`}
              className="text-sm text-amber-500 hover:underline"
            >
              View →
            </Link>
          </CardContent>
        </Card>
      )}

      {error && (
        <Card className="mb-6 border-red-500/30 bg-red-500/5">
          <CardContent className="py-3 text-sm text-red-300">
            Failed to load university rankings.
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <Globe className="h-5 w-5 text-amber-500" />
            Top {data?.total_ranked ?? 0} universities
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="space-y-2 p-4">
              {Array.from({ length: 10 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : data && data.entries.length > 0 ? (
            <ul className="divide-y divide-white/5">
              {data.entries.map((row) => (
                <li key={row.university_id}>
                  <Link
                    to={`/university-rankings/${row.slug}`}
                    className="flex items-center gap-4 px-4 py-3 transition-colors hover:bg-white/5"
                  >
                    <span className="w-10 text-right font-mono text-lg tabular-nums text-amber-500">
                      #{row.rank}
                    </span>
                    <div className="flex-1 min-w-0">
                      <p className="truncate font-medium">
                        {row.short_name || row.name}
                      </p>
                      <p className="truncate text-xs text-muted-foreground">
                        {row.country_code}
                        {row.state_code ? ` · ${row.state_code}` : ''}
                        {' · '}
                        {(row.student_count || 0).toLocaleString()} students
                      </p>
                    </div>
                    <div className="hidden text-right md:block">
                      <p className="font-mono text-sm tabular-nums">
                        {row.score_per_capita.toFixed(4)}
                      </p>
                      <p className="text-xs text-muted-foreground">per capita</p>
                    </div>
                    <div className="hidden text-right md:block">
                      <p className="font-mono text-sm tabular-nums">
                        {row.active_users.toLocaleString()}
                      </p>
                      <p className="text-xs text-muted-foreground">builders</p>
                    </div>
                    <div className="w-12 text-right">
                      {row.rank_delta > 0 && (
                        <span className="inline-flex items-center gap-0.5 text-xs text-emerald-400">
                          <TrendingUp className="h-3 w-3" />
                          {row.rank_delta}
                        </span>
                      )}
                      {row.rank_delta < 0 && (
                        <span className="inline-flex items-center gap-0.5 text-xs text-red-400">
                          <TrendingDown className="h-3 w-3" />
                          {Math.abs(row.rank_delta)}
                        </span>
                      )}
                      {row.rank_delta === 0 && (
                        <span className="text-xs text-muted-foreground">—</span>
                      )}
                    </div>
                  </Link>
                </li>
              ))}
            </ul>
          ) : (
            <p className="px-4 py-12 text-center text-sm text-muted-foreground">
              No universities ranked yet — invite builders from your school to get on the board.
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
