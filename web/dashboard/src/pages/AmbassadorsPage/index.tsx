import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Award, MapPin, Globe } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { useQuery } from '@tanstack/react-query';
import { ambassadorsApi } from '@/api/ambassadors';

const COUNTRY_OPTIONS: { code: string; label: string }[] = [
  { code: '', label: 'World' },
  { code: 'US', label: 'United States' },
  { code: 'GB', label: 'United Kingdom' },
  { code: 'CA', label: 'Canada' },
  { code: 'IN', label: 'India' },
  { code: 'DE', label: 'Germany' },
  { code: 'BR', label: 'Brazil' },
  { code: 'JP', label: 'Japan' },
];

export function AmbassadorsPage() {
  const [country, setCountry] = useState('');
  const { data, isLoading } = useQuery({
    queryKey: ['ambassadors', country],
    queryFn: () => ambassadorsApi.list(country || undefined),
  });

  return (
    <div className="container mx-auto max-w-5xl px-4 py-8">
      <header className="mb-8">
        <div className="flex items-center gap-3">
          <Award className="h-8 w-8 text-amber-500" />
          <div>
            <h1 className="text-3xl font-bold tracking-tight">City Ambassadors</h1>
            <p className="text-sm text-muted-foreground">
              Top builder per city. Promoted automatically every hour from the
              leaderboard; k≥5 privacy threshold enforced.
            </p>
          </div>
        </div>
      </header>

      <div className="mb-6 flex items-center gap-2 text-sm">
        <Globe className="h-4 w-4 text-muted-foreground" />
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
        <span className="text-xs text-muted-foreground">
          {data && `· ${data.total} ambassadors`}
        </span>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-lg">
            <Award className="h-5 w-5 text-amber-500" />
            {data?.total ?? 0} ambassadors
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
                <li key={`${row.metro_id}-${row.user_id}`}>
                  <Link
                    to={`/rankings/cities/${row.metro_slug}`}
                    className="flex items-center gap-4 px-4 py-3 transition-colors hover:bg-white/5"
                  >
                    <Award className="h-5 w-5 text-amber-500" />
                    <div className="flex-1 min-w-0">
                      <p className="truncate font-medium">{row.name}</p>
                      <p className="truncate text-xs text-muted-foreground">
                        {row.metro_name}
                        {row.state_code ? ` · ${row.state_code}` : ''}
                        {' · '}
                        {row.country_code}
                      </p>
                    </div>
                    <div className="hidden text-right text-xs text-muted-foreground md:block">
                      <p className="flex items-center gap-1">
                        <MapPin className="h-3 w-3" />
                        {new Date(row.promoted_at).toLocaleDateString()}
                      </p>
                      <p className="text-xs">{row.source}</p>
                    </div>
                  </Link>
                </li>
              ))}
            </ul>
          ) : (
            <p className="px-4 py-12 text-center text-sm text-muted-foreground">
              No ambassadors yet. Promote a top builder by inviting more
              activity to your city.
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
