import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { universityRankingsApi } from '@/api/universityRankings';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { ArrowLeft, GraduationCap, Globe, MapPin, Users, Zap, Code2, TrendingUp, Building2 } from 'lucide-react';

export function UniversityRankingsDetailPage() {
  const { slug } = useParams<{ slug: string }>();
  const { data, isLoading, error } = useQuery({
    queryKey: ['university-rankings', 'detail', slug],
    queryFn: () => universityRankingsApi.getDetail(slug!),
    enabled: !!slug,
  });

  if (isLoading) {
    return (
      <div className="container mx-auto max-w-5xl px-4 py-8">
        <Skeleton className="mb-4 h-8 w-64" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }
  if (error || !data) {
    return (
      <div className="container mx-auto max-w-5xl px-4 py-8">
        <Link to="/university-rankings" className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
          <ArrowLeft className="h-4 w-4" /> Back to rankings
        </Link>
        <Card>
          <CardContent className="py-12 text-center text-sm text-muted-foreground">
            University not found.
          </CardContent>
        </Card>
      </div>
    );
  }

  const e = data.entry;
  return (
    <div className="container mx-auto max-w-5xl px-4 py-8">
      <Link
        to="/university-rankings"
        className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" /> Back to rankings
      </Link>

      <header className="mb-6 flex items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2 text-xs uppercase tracking-wider text-muted-foreground">
            <GraduationCap className="h-3.5 w-3.5" />
            <span>{e.country_code ?? '—'}</span>
            {e.state_code && <span>· {e.state_code}</span>}
            <span>· University</span>
          </div>
          <h1 className="mt-1 text-3xl font-bold tracking-tight">{e.name}</h1>
          {e.short_name && e.short_name !== e.name && (
            <p className="text-lg text-muted-foreground">{e.short_name}</p>
          )}
          {e.city_slug && (
            <Link
              to={`/rankings/cities/${e.city_slug}`}
              className="mt-1 inline-flex items-center gap-1 text-sm text-amber-500 hover:underline"
            >
              <Building2 className="h-3.5 w-3.5" />
              View city ranking
            </Link>
          )}
        </div>
        {e.rank > 0 && (
          <div className="text-right">
            <div className="text-5xl font-bold text-amber-500">#{e.rank}</div>
            <div className="text-xs text-muted-foreground">per-capita rank</div>
          </div>
        )}
      </header>

      <Card className="mb-6">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <TrendingUp className="h-5 w-5 text-amber-500" />
            Score breakdown
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <StatCard label="Per-capita score" value={e.score_per_capita.toFixed(4)} icon={<TrendingUp className="h-4 w-4 text-amber-500" />} />
            <StatCard label="Raw score" value={e.score_raw.toFixed(3)} icon={<Zap className="h-4 w-4 text-amber-500" />} />
            <StatCard label="Active builders (30d)" value={e.active_users.toLocaleString()} icon={<Users className="h-4 w-4 text-amber-500" />} />
            <StatCard label="Deployments" value={e.deployments.toLocaleString()} icon={<Code2 className="h-4 w-4 text-amber-500" />} />
            <StatCard label="Executions (30d)" value={e.executions_30d.toLocaleString()} icon={<Zap className="h-4 w-4 text-amber-500" />} />
            <StatCard label="New users (30d)" value={e.new_users_30d.toLocaleString()} icon={<MapPin className="h-4 w-4 text-amber-500" />} />
          </div>
          {e.student_count && e.student_count > 0 && (
            <div className="mt-4 border-t border-white/5 pt-4 text-xs text-muted-foreground">
              Student count: {e.student_count.toLocaleString()} · per-capita = raw × 100,000 / {e.student_count.toLocaleString()}
            </div>
          )}
        </CardContent>
      </Card>

      <Card className="border-amber-500/20 bg-amber-500/5">
        <CardContent className="flex items-center gap-4 py-4">
          <Globe className="h-8 w-8 text-amber-500" />
          <div className="flex-1">
            <p className="text-xs uppercase tracking-wider text-muted-foreground">
              Institution
            </p>
            <p className="font-medium">
              {e.name}
              {e.short_name && e.short_name !== e.name && (
                <span className="ml-2 text-muted-foreground">({e.short_name})</span>
              )}
            </p>
            <p className="text-xs text-muted-foreground">
              {e.country_code}
              {e.state_code && ` · ${e.state_code}`}
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function StatCard({ label, value, icon }: { label: string; value: string; icon: React.ReactNode }) {
  return (
    <Card>
      <CardContent className="flex items-center justify-between p-4">
        <div>
          <div className="text-xs uppercase tracking-wider text-muted-foreground">{label}</div>
          <div className="mt-1 font-mono text-xl">{value}</div>
        </div>
        {icon}
      </CardContent>
    </Card>
  );
}
