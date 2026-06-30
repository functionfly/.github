/**
 * MCP Center - Summary Cards Component
 * KPI cards showing MCP registry summary
 */

import { Link } from 'react-router-dom';
import { Loader2, Zap, Shield, Activity, Globe } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

interface SummaryCardsProps {
  stats: {
    total: number;
    enabled: number;
    verified: number;
    totalInvocations: number;
    transportsCount: number;
  };
  isLoading: boolean;
}

export function SummaryCards({ stats, isLoading }: SummaryCardsProps) {
  const cards = [
    {
      title: 'MCP Functions',
      value: stats.total,
      subtitle: `${stats.enabled} enabled`,
      icon: Zap,
      href: '/mcp?filter=enabled',
      color: 'text-[var(--status-ok)]',
    },
    {
      title: 'Verified MCP',
      value: stats.verified,
      subtitle: 'trusted by clients',
      icon: Shield,
      href: '/mcp?filter=verified',
      color: 'text-[var(--status-ok)]',
    },
    {
      title: 'Total Invocations',
      value: stats.totalInvocations.toLocaleString(),
      subtitle: 'last 30 days',
      icon: Activity,
      href: '/mcp?tab=analytics',
      color: 'text-[var(--status-pending)]',
    },
    {
      title: 'Active Transports',
      value: stats.transportsCount,
      subtitle: 'protocols',
      icon: Globe,
      href: '/mcp?tab=settings',
      color: 'text-[var(--foil-a)]',
    },
  ];

  if (isLoading) {
    return (
      <div className="mcp-cards-grid">
        {cards.map((_, i) => (
          <Card key={i} className="mcp-metric-card">
            <CardHeader className="pb-2">
              <div className="h-4 w-20 bg-muted animate-pulse rounded" />
            </CardHeader>
            <CardContent>
              <div className="h-8 w-16 bg-muted animate-pulse rounded" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  return (
    <div className="mcp-cards-grid">
      {cards.map((card) => {
        const Icon = card.icon;
        return (
          <Link key={card.title} to={card.href} className="block">
            <Card className="mcp-metric-card">
              <CardHeader className="pb-2">
                <CardTitle className="mcp-metric-card-title flex items-center gap-2">
                  <Icon className={`h-4 w-4 ${card.color}`} />
                  {card.title}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="mcp-metric-card-value">{card.value}</p>
                <p className="mcp-metric-card-subtitle">{card.subtitle}</p>
              </CardContent>
            </Card>
          </Link>
        );
      })}
    </div>
  );
}

// Individual metric card component
interface MCPMetricsCardProps {
  title: string;
  value: string | number;
  change?: number;
  subtitle?: string;
  icon?: React.ReactNode;
  iconColor?: string;
  lowerIsBetter?: boolean;
  higherIsBetter?: boolean;
  isLoading?: boolean;
}

export function MCPMetricsCard({
  title,
  value,
  change,
  subtitle,
  icon,
  iconColor = 'text-[var(--status-ok)]',
  lowerIsBetter,
  higherIsBetter,
  isLoading,
}: MCPMetricsCardProps) {
  if (isLoading) {
    return (
      <Card className="border-theme bg-card">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm font-medium text-text-secondary">{title}</CardTitle>
        </CardHeader>
        <CardContent>
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    );
  }

  const getChangeColor = () => {
    if (change === undefined || change === 0) return 'text-text-secondary';
    if (lowerIsBetter) return change < 0 ? 'text-[var(--status-ok)]' : 'text-[var(--status-pending)]';
    if (higherIsBetter) return change > 0 ? 'text-[var(--status-ok)]' : 'text-[var(--status-pending)]';
    return 'text-text-secondary';
  };

  const formatChange = () => {
    if (change === undefined) return null;
    const sign = change > 0 ? '+' : '';
    return `${sign}${change.toFixed(1)}%`;
  };

  return (
    <Card className="border-theme bg-card">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium text-text-secondary flex items-center gap-2">
          {icon && <span className={iconColor}>{icon}</span>}
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-2xl font-semibold text-text-primary">{value}</p>
        {change !== undefined && (
          <p className={`text-xs mt-1 ${getChangeColor()}`}>
            {formatChange()} {subtitle}
          </p>
        )}
        {subtitle && change === undefined && (
          <p className="text-xs text-muted-foreground mt-1">{subtitle}</p>
        )}
      </CardContent>
    </Card>
  );
}
