import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { useMyCity, useSetCityOptOut, useCityOptOut } from '@/hooks/useCityRankings';
import { TrendingUp, MapPin, Shield, ShieldOff } from 'lucide-react';
import { Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';

export function MyCityWidget() {
  const { data, isLoading } = useMyCity();
  const { data: optOut } = useCityOptOut();
  const setOptOut = useSetCityOptOut();

  if (isLoading) {
    return (
      <Card className="border-white/10 bg-gradient-to-br from-amber-500/5 to-orange-500/5">
        <CardContent className="p-5">
          <Skeleton className="h-4 w-32 mb-3" />
          <Skeleton className="h-8 w-48 mb-2" />
          <Skeleton className="h-3 w-24" />
        </CardContent>
      </Card>
    );
  }

  if (!data?.has_city) {
    return (
      <Card className="border-white/10 bg-gradient-to-br from-amber-500/5 to-orange-500/5">
        <CardContent className="p-5">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <MapPin className="h-4 w-4" />
            <span>Your city isn't ranked yet</span>
          </div>
          <p className="mt-2 text-xs text-muted-foreground">
            {optOut?.opted_out
              ? "You're opted out of city rankings. Set a location on your profile to be counted in."
              : 'Set your location in profile settings to see how your city ranks.'}
          </p>
          <Button variant="outline" size="sm" className="mt-3" asChild>
            <Link to="/settings/profile">Set location</Link>
          </Button>
        </CardContent>
      </Card>
    );
  }

  const m = data.metro!;
  return (
    <Card className="border-white/10 bg-gradient-to-br from-amber-500/10 to-orange-500/10">
      <CardContent className="p-5">
        <div className="flex items-start justify-between gap-3">
          <div>
            <div className="flex items-center gap-2 text-xs uppercase tracking-wider text-muted-foreground">
              <MapPin className="h-3 w-3" /> Your city
            </div>
            <Link
              to={`/rankings/cities/${m.metro_slug}`}
              className="mt-1 block text-2xl font-semibold hover:underline"
            >
              {m.metro_name}
            </Link>
            <div className="mt-1 flex items-center gap-3 text-sm text-muted-foreground">
              <span className="flex items-center gap-1">
                <TrendingUp className="h-3.5 w-3.5 text-amber-500" /> Rank #{m.rank}
              </span>
              {m.rank_delta !== 0 && (
                <span className={m.rank_delta > 0 ? 'text-emerald-400' : 'text-rose-400'}>
                  {m.rank_delta > 0 ? '▲' : '▼'} {Math.abs(m.rank_delta)}
                </span>
              )}
            </div>
          </div>
          <div className="text-right">
            <div className="text-xs uppercase tracking-wider text-muted-foreground">Per capita</div>
            <div className="mt-1 font-mono text-lg">{m.score_per_capita.toFixed(4)}</div>
            <div className="text-[10px] text-muted-foreground">raw {m.score_raw.toFixed(3)}</div>
          </div>
        </div>
        <div className="mt-3 flex items-center justify-between border-t border-white/5 pt-3 text-xs">
          <span className="text-muted-foreground">
            {m.active_users} active builders · {m.executions_30d.toLocaleString()} executions / 30d
          </span>
          <button
            type="button"
            onClick={() => setOptOut.mutate(!optOut?.opted_out)}
            className="flex items-center gap-1 text-muted-foreground hover:text-foreground"
            disabled={setOptOut.isPending}
          >
            {optOut?.opted_out ? (
              <>
                <ShieldOff className="h-3 w-3" /> Opted out
              </>
            ) : (
              <>
                <Shield className="h-3 w-3" /> Hide me
              </>
            )}
          </button>
        </div>
      </CardContent>
    </Card>
  );
}
