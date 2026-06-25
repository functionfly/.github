import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { cityRankingsApi } from '@/api/cityRankings';
import { ambassadorsApi } from '@/api/ambassadors';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { ArrowLeft, Globe, MapPin, Users, Zap, Code2, TrendingUp, Award, GraduationCap } from 'lucide-react';

export function CityRankingsDetailPage() {
  const { slug } = useParams<{ slug: string }>();
  const { data, isLoading, error } = useQuery({
    queryKey: ['city-rankings', 'metro', slug],
    queryFn: () => cityRankingsApi.getMetro(slug!),
    enabled: !!slug,
  });
  const { data: ambassador } = useQuery({
    queryKey: ['city-rankings', 'ambassador', slug],
    queryFn: () => ambassadorsApi.getForMetro(slug!).catch(() => null),
    enabled: !!slug,
    retry: false,
  });
  const { data: universities } = useQuery({
    queryKey: ['city-rankings', 'universities', slug],
    queryFn: () => cityRankingsApi.getUniversitiesForMetro(slug!).catch(() => null),
    enabled: !!slug,
    retry: false,
  });

  if (isLoading) {
    return (
      <div className="container mx-auto max-w-5xl px-4 py-8">
        <Skeleton className="h-8 w-64 mb-4" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }
  if (error || !data) {
    return (
      <div className="container mx-auto max-w-5xl px-4 py-8">
        <Link to="/rankings" className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
          <ArrowLeft className="h-4 w-4" /> Back to rankings
        </Link>
        <Card>
          <CardContent className="py-12 text-center text-sm text-muted-foreground">
            Metro not found.
          </CardContent>
        </Card>
      </div>
    );
  }

  const c = data.current;
  return (
    <div className="container mx-auto max-w-5xl px-4 py-8">
      <Link
        to="/rankings"
        className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" /> Back to rankings
      </Link>

      <header className="mb-6 flex items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2 text-xs uppercase tracking-wider text-muted-foreground">
            <Globe className="h-3.5 w-3.5" /> {c?.country_code ?? '—'} · Metro
          </div>
          <h1 className="mt-1 text-3xl font-bold tracking-tight">{c?.metro_name ?? slug}</h1>
        </div>
        {c && (
          <div className="text-right">
            <div className="text-5xl font-bold text-amber-500">#{c.rank}</div>
            <div className="text-xs text-muted-foreground">global per-capita</div>
          </div>
        )}
      </header>

      {data.not_ranked ? (
        <Card className="border-amber-500/20">
          <CardContent className="py-8 text-center text-sm text-muted-foreground">
            This metro has fewer than {data.privacy_min_active_users} active builders in the
            last 30 days. Once it crosses the privacy threshold it will appear on the
            public leaderboard.
          </CardContent>
        </Card>
      ) : (
        c && <CityStats current={c} />
      )}

      {data.history.length > 0 && (
        <Card className="mt-6">
          <CardHeader>
            <CardTitle>Last 30 days</CardTitle>
          </CardHeader>
          <CardContent>
            <ScoreSparkline history={data.history} />
          </CardContent>
        </Card>
      )}

      {ambassador?.ambassador && (
        <Card className="mt-6 border-amber-500/30 bg-amber-500/5">
          <CardContent className="flex items-center gap-4 py-4">
            <Award className="h-8 w-8 text-amber-500" />
            <div className="flex-1">
              <p className="text-xs uppercase tracking-wider text-muted-foreground">
                City Ambassador
              </p>
              <p className="text-base font-semibold">{ambassador.ambassador.name}</p>
              <p className="text-xs text-muted-foreground">
                Promoted {new Date(ambassador.ambassador.promoted_at).toLocaleDateString()} · source: {ambassador.ambassador.source}
              </p>
            </div>
          </CardContent>
        </Card>
      )}

      {universities?.entries?.length > 0 && (
        <Card className="mt-6">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <GraduationCap className="h-5 w-5 text-amber-500" />
              Top universities in this metro
            </CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="divide-y divide-white/5">
              {universities.entries?.slice(0, 10).map((uni) => (
                <li key={uni.university_id}>
                  <Link
                    to={`/university-rankings/${uni.slug}`}
                    className="flex items-center justify-between py-3 transition-colors hover:bg-white/5"
                  >
                    <div className="flex items-center gap-3">
                      <span className="w-8 text-right font-mono text-amber-500">#{uni.rank}</span>
                      <div>
                        <p className="font-medium">{uni.short_name || uni.name}</p>
                        <p className="text-xs text-muted-foreground">
                          {uni.country_code}
                          {uni.state_code && ` · ${uni.state_code}`}
                        </p>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="font-mono text-sm">{uni.score_per_capita.toFixed(4)}</p>
                      <p className="text-xs text-muted-foreground">{uni.active_users} builders</p>
                    </div>
                  </Link>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function CityStats({ current }: { current: import('@/api/cityRankings').CityRankingEntry }) {
  const stats = [
    { label: 'Per-capita score', value: current.score_per_capita.toFixed(4), icon: <TrendingUp className="h-4 w-4 text-amber-500" /> },
    { label: 'Raw score', value: current.score_raw.toFixed(3), icon: <Zap className="h-4 w-4 text-amber-500" /> },
    { label: 'Active users (30d)', value: current.active_users.toLocaleString(), icon: <Users className="h-4 w-4 text-amber-500" /> },
    { label: 'Deployments', value: current.deployments.toLocaleString(), icon: <Code2 className="h-4 w-4 text-amber-500" /> },
    { label: 'Executions (30d)', value: current.executions_30d.toLocaleString(), icon: <Zap className="h-4 w-4 text-amber-500" /> },
    { label: 'New users (30d)', value: current.new_users_30d.toLocaleString(), icon: <MapPin className="h-4 w-4 text-amber-500" /> },
  ];
  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
      {stats.map((s) => (
        <Card key={s.label}>
          <CardContent className="flex items-center justify-between p-4">
            <div>
              <div className="text-xs uppercase tracking-wider text-muted-foreground">{s.label}</div>
              <div className="mt-1 font-mono text-xl">{s.value}</div>
            </div>
            {s.icon}
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

function ScoreSparkline({ history }: { history: import('@/api/cityRankings').CityRankingEntry[] }) {
  if (history.length < 2) {
    return <p className="text-xs text-muted-foreground">Not enough history yet.</p>;
  }
  const scores = history.map((h) => h.score_per_capita);
  const min = Math.min(...scores);
  const max = Math.max(...scores);
  const range = max - min || 1;
  const width = 600;
  const height = 80;
  const stepX = width / (scores.length - 1);
  const points = scores
    .map((v, i) => `${(i * stepX).toFixed(1)},${(height - ((v - min) / range) * height).toFixed(1)}`)
    .join(' ');
  return (
    <div className="overflow-x-auto">
      <svg width={width} height={height} className="text-amber-500">
        <polyline
          points={points}
          fill="none"
          stroke="currentColor"
          strokeWidth={2}
          strokeLinejoin="round"
        />
      </svg>
      <div className="mt-1 flex justify-between text-[10px] text-muted-foreground">
        <span>{new Date(history[0].period_end).toLocaleDateString()}</span>
        <span>min {min.toFixed(4)} · max {max.toFixed(4)}</span>
        <span>{new Date(history[history.length - 1].period_end).toLocaleDateString()}</span>
      </div>
    </div>
  );
}
