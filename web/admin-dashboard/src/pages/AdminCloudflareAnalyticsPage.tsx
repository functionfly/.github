import { adminApiClient } from '@/lib/api/adminClient';
import { API_ROUTES } from '@/lib/constants';
import { useQuery } from '@tanstack/react-query';
import {
  AlertTriangle,
  ArrowDownRight,
  ArrowUpRight,
  Cloud,
  Globe,
  Lock,
  RefreshCw,
  Shield,
  Wifi,
  Zap,
} from 'lucide-react';
import { useState } from 'react';

// ── Types ────────────────────────────────────────────────────────────────────

interface TrendStat {
  value: number;
  change_pct: number;
  unit?: string;
}

interface CountryStat {
  country: string;
  country_code: string;
  requests: number;
  bandwidth_bytes: number;
}

interface NetworkBreakdown {
  label: string;
  count: number;
}

interface CloudflareAnalytics {
  period: string;
  traffic: {
    requests: TrendStat;
    bandwidth: TrendStat;
    visits: TrendStat;
    page_views: TrendStat;
  };
  security: {
    encrypted_requests: TrendStat;
    encrypted_requests_rate: TrendStat;
    encrypted_bandwidth: TrendStat;
    encrypted_bandwidth_rate: TrendStat;
  };
  cache: {
    cached_requests: TrendStat;
    cached_requests_rate: TrendStat;
    cached_bandwidth: TrendStat;
    cached_bandwidth_rate: TrendStat;
  };
  errors: {
    errors_4xx: TrendStat;
    error_rate_4xx: TrendStat;
    errors_5xx: TrendStat;
    error_rate_5xx: TrendStat;
  };
  top_countries: CountryStat[];
  http_versions: NetworkBreakdown[];
  tls_versions: NetworkBreakdown[];
  content_types: NetworkBreakdown[];
}

// ── Helpers ──────────────────────────────────────────────────────────────────

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${units[i]}`;
}

function formatNumber(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(2)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(2)}k`;
  return n.toString();
}

function formatPct(n: number): string {
  return `${n.toFixed(2)}%`;
}

function clamp(n: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, n));
}

const PERIODS = [
  { label: '24h', value: '24h' },
  { label: '7d', value: '7d' },
  { label: '30d', value: '30d' },
];

const COUNTRY_FLAGS: Record<string, string> = {
  FR: '🇫🇷',
  US: '🇺🇸',
  NL: '🇳🇱',
  CH: '🇨🇭',
  IN: '🇮🇳',
  CA: '🇨🇦',
  GB: '🇬🇧',
  CN: '🇨🇳',
  BR: '🇧🇷',
  DE: '🇩🇪',
  JP: '🇯🇵',
  AU: '🇦🇺',
  SG: '🇸🇬',
  RU: '🇷🇺',
  KR: '🇰🇷',
};

// ── Sub-components ────────────────────────────────────────────────────────────

function TrendBadge({ change_pct }: { change_pct: number }) {
  const up = change_pct >= 0;
  return (
    <span
      className={`inline-flex items-center gap-0.5 text-xs font-mono font-semibold ${
        up ? 'text-emerald-600' : 'text-red-500'
      }`}
    >
      {up ? <ArrowUpRight className="w-3 h-3" /> : <ArrowDownRight className="w-3 h-3" />}
      {Math.abs(change_pct).toFixed(2)}%
    </span>
  );
}

interface StatCardProps {
  label: string;
  stat: TrendStat;
  format?: 'number' | 'bytes' | 'pct';
  warn?: boolean;
}

function StatCard({ label, stat, format = 'number', warn = false }: StatCardProps) {
  const display =
    format === 'bytes'
      ? formatBytes(stat.value)
      : format === 'pct'
        ? formatPct(stat.value)
        : formatNumber(stat.value);

  return (
    <div
      className={`rounded-xl border p-5 flex flex-col gap-3 transition-shadow hover:shadow-md ${
        warn ? 'bg-amber-50 border-amber-200' : 'bg-white border-gray-100'
      }`}
    >
      <p className="text-xs font-semibold uppercase tracking-widest text-gray-400">{label}</p>
      <p
        className={`text-2xl font-mono font-bold leading-none ${warn ? 'text-amber-700' : 'text-gray-900'}`}
      >
        {display}
      </p>
      <TrendBadge change_pct={stat.change_pct} />
    </div>
  );
}

function SectionHeader({ icon: Icon, title }: { icon: React.ElementType; title: string }) {
  return (
    <div className="flex items-center gap-2 mb-4">
      <Icon className="w-4 h-4 text-gray-400" />
      <h2 className="text-xs font-bold uppercase tracking-widest text-gray-400">{title}</h2>
    </div>
  );
}

function MiniBar({ value, max }: { value: number; max: number }) {
  const pct = max > 0 ? (value / max) * 100 : 0;
  return (
    <div className="h-1.5 w-24 bg-gray-100 rounded-full overflow-hidden">
      <div className="h-full bg-blue-500 rounded-full" style={{ width: `${pct}%` }} />
    </div>
  );
}

function DonutChart({
  valuePct,
  color,
  size = 80,
  strokeWidth = 10,
  centerLabel,
}: {
  valuePct: number;
  color: string;
  size?: number;
  strokeWidth?: number;
  centerLabel: string;
}) {
  const v = clamp(valuePct, 0, 100);
  const radius = 34 / 2; // 17
  const normalizedRadius = radius - strokeWidth / 2;
  const circumference = 2 * Math.PI * normalizedRadius;
  const strokeDasharray = `${(v / 100) * circumference} ${circumference}`;

  return (
    <div className="relative" style={{ width: size, height: size }}>
      <svg
        className="-rotate-90"
        viewBox="0 0 34 34"
        width={size}
        height={size}
        aria-hidden="true"
      >
        <circle cx="17" cy="17" r={normalizedRadius} fill="none" stroke="#E5E7EB" strokeWidth={strokeWidth} />
        <circle
          cx="17"
          cy="17"
          r={normalizedRadius}
          fill="none"
          stroke={color}
          strokeWidth={strokeWidth}
          strokeLinecap="round"
          strokeDasharray={strokeDasharray}
        />
      </svg>
      <div className="absolute inset-0 flex items-center justify-center">
        <span className="text-sm font-semibold font-mono text-gray-900">{centerLabel}</span>
      </div>
    </div>
  );
}

function DonutStatCard({
  label,
  stat,
  valueFormatter,
  color,
  warn,
}: {
  label: string;
  stat: TrendStat;
  valueFormatter: (value: number) => string;
  color: string;
  warn?: boolean;
}) {
  return (
    <div
      className={`rounded-2xl border p-5 flex flex-col gap-4 transition-shadow ${
        warn ? 'bg-amber-50 border-amber-200' : 'bg-white border-gray-100'
      }`}
    >
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <p className="text-xs font-semibold uppercase tracking-widest text-gray-400">{label}</p>
          <div className="mt-1">
            <TrendBadge change_pct={stat.change_pct} />
          </div>
        </div>
        <DonutChart valuePct={stat.value} color={color} centerLabel={valueFormatter(stat.value)} />
      </div>
    </div>
  );
}

function VerticalBarsChart({
  items,
  color,
  ariaLabel,
}: {
  items: NetworkBreakdown[];
  color: string;
  ariaLabel: string;
}) {
  const max = Math.max(...items.map((i) => i.count), 1);
  return (
    <div className="px-6 py-5">
      <div className="flex flex-wrap gap-x-6 gap-y-7 items-end">
        {items.map((v) => {
          const h = max > 0 ? (v.count / max) * 100 : 0;
          const heightPct = `${clamp(h, 0, 100)}%`;
          return (
            <div key={v.label} className="w-16">
              <div className="h-40 flex items-end justify-center">
                <div
                  className="w-3 rounded-md"
                  style={{ height: heightPct, backgroundColor: color }}
                  aria-label={`${ariaLabel}: ${v.label} (${formatNumber(v.count)})`}
                  role="img"
                />
              </div>
              <div
                className="mt-3 text-[10px] font-mono text-gray-500 text-center leading-tight truncate"
                title={v.label}
              >
                {v.label}
              </div>
              <div className="mt-1 text-[10px] font-mono text-gray-700 text-center">
                {formatNumber(v.count)}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// ── Skeleton ──────────────────────────────────────────────────────────────────

function Skeleton({ className = '' }: { className?: string }) {
  return <div className={`animate-pulse bg-gray-100 rounded-lg ${className}`} />;
}

function PageSkeleton() {
  return (
    <div className="space-y-8">
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-28" />
        ))}
      </div>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-28" />
        ))}
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Skeleton className="h-64" />
        <Skeleton className="h-64" />
      </div>
    </div>
  );
}

// ── Setup Banner ──────────────────────────────────────────────────────────────

function SetupBanner() {
  return (
    <div className="rounded-2xl border-2 border-dashed border-gray-200 bg-gray-50 p-12 flex flex-col items-center gap-4 text-center">
      <div className="w-16 h-16 rounded-2xl bg-orange-100 flex items-center justify-center">
        <Cloud className="w-8 h-8 text-orange-500" />
      </div>
      <div>
        <h3 className="text-lg font-bold text-gray-900">Cloudflare Analytics not configured</h3>
        <p className="text-sm text-gray-500 mt-1 max-w-md">
          Add your Cloudflare API token and Zone ID to the server environment to enable live traffic
          analytics.
        </p>
      </div>
      <div className="mt-2 rounded-xl bg-gray-900 text-gray-100 font-mono text-xs px-6 py-4 text-left leading-relaxed max-w-lg w-full">
        <span className="text-gray-500"># .env</span>
        <br />
        <span className="text-emerald-400">CF_API_TOKEN</span>=your_token_here
        <br />
        <span className="text-emerald-400">CF_ZONE_ID</span>=your_zone_id_here
      </div>
    </div>
  );
}

// ── Main Page ─────────────────────────────────────────────────────────────────

export function AdminCloudflareAnalyticsPage() {
  const [period, setPeriod] = useState('24h');

  const { data, isLoading, isError, error, refetch, isFetching } = useQuery<CloudflareAnalytics>({
    queryKey: ['cloudflare-analytics', period],
    queryFn: async () => {
      const res = await adminApiClient.get<CloudflareAnalytics>(
        `${API_ROUTES.ADMIN_CLOUDFLARE_ANALYTICS}?period=${period}`
      );
      return res.data;
    },
    refetchInterval: 5 * 60 * 1000, // 5 minutes
    retry: 1,
  });

  const notConfigured = isError && (error as { status?: number })?.status === 404;

  const maxCountryRequests = Math.max(...(data?.top_countries?.map((c) => c.requests) ?? [1]), 1);

  const has4xxAlert = (data?.errors?.error_rate_4xx?.value ?? 0) > 5;

  return (
    <div className="space-y-8">
      {/* ── Header ── */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-orange-100 flex items-center justify-center">
              <Cloud className="w-4 h-4 text-orange-500" />
            </div>
            <h1 className="text-2xl font-bold text-gray-900 tracking-tight">
              Cloudflare Analytics
            </h1>
          </div>
          <p className="text-sm text-gray-500 mt-1 ml-10">
            Live CDN traffic, security, and performance data
          </p>
        </div>

        <div className="flex items-center gap-3">
          {/* Period selector */}
          <div className="flex bg-gray-100 rounded-lg p-1 gap-1">
            {PERIODS.map((p) => (
              <button
                key={p.value}
                onClick={() => setPeriod(p.value)}
                className={`px-3 py-1.5 rounded-md text-sm font-medium transition-all ${
                  period === p.value
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-500 hover:text-gray-700'
                }`}
              >
                {p.label}
              </button>
            ))}
          </div>

          <button
            onClick={() => refetch()}
            disabled={isFetching}
            className="flex items-center gap-2 px-3 py-2 bg-white border border-gray-200 text-gray-600 rounded-lg hover:bg-gray-50 disabled:opacity-50 text-sm"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isFetching ? 'animate-spin' : ''}`} />
            Refresh
          </button>
        </div>
      </div>

      {/* ── Alert banner ── */}
      {has4xxAlert && data && (
        <div className="flex items-start gap-3 bg-amber-50 border border-amber-200 rounded-xl px-5 py-4">
          <AlertTriangle className="w-5 h-5 text-amber-500 mt-0.5 flex-shrink-0" />
          <div>
            <p className="text-sm font-semibold text-amber-800">
              Elevated 4xx error rate — {formatPct(data.errors.error_rate_4xx.value)} (
              {formatNumber(data.errors.errors_4xx.value)} errors in the last {period})
            </p>
            <p className="text-xs text-amber-600 mt-0.5">
              Check origin server logs or Cloudflare firewall rules for blocked requests.
            </p>
          </div>
        </div>
      )}

      {/* ── Body ── */}
      {isLoading ? (
        <PageSkeleton />
      ) : notConfigured ? (
        <SetupBanner />
      ) : isError ? (
        <div className="text-center py-16 text-gray-400">
          <p className="font-medium">Failed to load analytics</p>
          <button onClick={() => refetch()} className="mt-2 text-sm text-blue-500 hover:underline">
            Try again
          </button>
        </div>
      ) : data ? (
        <div className="space-y-8">
          {/* ── Key Ratios (Graphs) ── */}
          <section>
            <div className="flex items-center justify-between gap-4 mb-4">
              <div className="flex items-center gap-2">
                <Zap className="w-4 h-4 text-gray-400" />
                <h2 className="text-xs font-bold uppercase tracking-widest text-gray-400">Key Ratios</h2>
              </div>
              <p className="text-xs text-gray-500">Percentages over the selected period</p>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <DonutStatCard
                label="Encrypted Requests Rate"
                stat={data.security.encrypted_requests_rate}
                valueFormatter={(n) => formatPct(n)}
                color="#3B82F6" // blue-500
              />
              <DonutStatCard
                label="Cache Hit Rate"
                stat={data.cache.cached_requests_rate}
                valueFormatter={(n) => formatPct(n)}
                color="#10B981" // emerald-500
                warn={data.cache.cached_requests_rate.value < 20}
              />
              <DonutStatCard
                label="4xx Error Rate"
                stat={data.errors.error_rate_4xx}
                valueFormatter={(n) => formatPct(n)}
                color="#F97316" // orange-500
                warn={(data.errors.error_rate_4xx.value ?? 0) > 2}
              />
            </div>
          </section>

          {/* ── Traffic ── */}
          <section>
            <SectionHeader icon={Globe} title="Traffic" />
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <StatCard label="Requests" stat={data.traffic.requests} />
              <StatCard label="Bandwidth" stat={data.traffic.bandwidth} format="bytes" />
              <StatCard label="Unique Visits" stat={data.traffic.visits} />
              <StatCard label="Page Views" stat={data.traffic.page_views} />
            </div>
          </section>

          {/* ── Errors ── */}
          <section>
            <SectionHeader icon={AlertTriangle} title="Errors" />
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <StatCard
                label="4xx Errors"
                stat={data.errors.errors_4xx}
                warn={data.errors.errors_4xx.value > 0}
              />
              <StatCard
                label="4xx Error Rate"
                stat={data.errors.error_rate_4xx}
                format="pct"
                warn={data.errors.error_rate_4xx.value > 2}
              />
              <StatCard label="5xx Errors" stat={data.errors.errors_5xx} />
              <StatCard label="5xx Error Rate" stat={data.errors.error_rate_5xx} format="pct" />
            </div>
          </section>

          {/* ── Security ── */}
          <section>
            <SectionHeader icon={Lock} title="Security" />
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <StatCard label="Encrypted Requests" stat={data.security.encrypted_requests} />
              <StatCard
                label="Encrypted Rate"
                stat={data.security.encrypted_requests_rate}
                format="pct"
              />
              <StatCard
                label="Encrypted Bandwidth"
                stat={data.security.encrypted_bandwidth}
                format="bytes"
              />
              <StatCard
                label="Enc. Bandwidth Rate"
                stat={data.security.encrypted_bandwidth_rate}
                format="pct"
              />
            </div>
          </section>

          {/* ── Cache ── */}
          <section>
            <SectionHeader icon={Zap} title="Cache" />
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <StatCard label="Cached Requests" stat={data.cache.cached_requests} />
              <StatCard
                label="Cache Hit Rate"
                stat={data.cache.cached_requests_rate}
                format="pct"
                warn={data.cache.cached_requests_rate.value < 20}
              />
              <StatCard
                label="Cached Bandwidth"
                stat={data.cache.cached_bandwidth}
                format="bytes"
              />
              <StatCard
                label="Cached BW Rate"
                stat={data.cache.cached_bandwidth_rate}
                format="pct"
              />
            </div>
          </section>

          {/* ── Countries + Network ── */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Top Countries */}
            <section className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
              <div className="px-6 py-4 border-b border-gray-50 flex items-center gap-2">
                <Globe className="w-4 h-4 text-gray-400" />
                <h2 className="text-xs font-bold uppercase tracking-widest text-gray-400">
                  Top Countries
                </h2>
              </div>
              <div className="divide-y divide-gray-50">
                {data.top_countries.map((c) => (
                  <div key={c.country_code} className="flex items-center gap-4 px-6 py-3">
                    <span className="text-lg w-7 text-center flex-shrink-0">
                      {COUNTRY_FLAGS[c.country_code] ?? '🌐'}
                    </span>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between mb-1">
                        <span className="text-sm font-medium text-gray-800 truncate">
                          {c.country}
                        </span>
                        <span className="font-mono text-xs text-gray-500 ml-2 flex-shrink-0">
                          {formatNumber(c.requests)} req
                        </span>
                      </div>
                      <MiniBar value={c.requests} max={maxCountryRequests} />
                    </div>
                    <div className="text-right flex-shrink-0 w-20">
                      <span className="font-mono text-xs text-gray-400">
                        {formatBytes(c.bandwidth_bytes)}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </section>

            {/* Network Breakdown */}
            <div className="space-y-4">
              {/* HTTP Versions */}
              <section className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
                <div className="px-6 py-4 border-b border-gray-50 flex items-center gap-2">
                  <Wifi className="w-4 h-4 text-gray-400" />
                  <h2 className="text-xs font-bold uppercase tracking-widest text-gray-400">
                    HTTP Versions
                  </h2>
                </div>
                <VerticalBarsChart
                  items={data.http_versions}
                  color="#60A5FA" // blue-400
                  ariaLabel="HTTP versions requests"
                />
              </section>

              {/* TLS Versions */}
              <section className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
                <div className="px-6 py-4 border-b border-gray-50 flex items-center gap-2">
                  <Shield className="w-4 h-4 text-gray-400" />
                  <h2 className="text-xs font-bold uppercase tracking-widest text-gray-400">
                    TLS Versions
                  </h2>
                </div>
                <VerticalBarsChart
                  items={data.tls_versions}
                  color="#34D399" // emerald-400
                  ariaLabel="TLS versions requests"
                />
              </section>

              {/* Content Types */}
              <section className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
                <div className="px-6 py-4 border-b border-gray-50 flex items-center gap-2">
                  <Globe className="w-4 h-4 text-gray-400" />
                  <h2 className="text-xs font-bold uppercase tracking-widest text-gray-400">
                    Top Content Types
                  </h2>
                </div>
                <VerticalBarsChart
                  items={data.content_types}
                  color="#A78BFA" // purple-400
                  ariaLabel="Top content types requests"
                />
              </section>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}

export default AdminCloudflareAnalyticsPage;
