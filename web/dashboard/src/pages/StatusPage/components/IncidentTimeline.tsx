import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  AlertTriangle,
  AlertCircle,
  Info,
  CheckCircle2,
  ChevronDown,
  Clock,
  Calendar,
  Search,
  X,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/utils';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import type { Incident, IncidentSeverity, IncidentStatus } from '@/api/status';

interface IncidentTimelineProps {
  incidents: Incident[];
  isLoading?: boolean;
  showFilters?: boolean;
  maxItems?: number;
}

const severityIcons: Record<IncidentSeverity, React.ComponentType<{ className?: string }>> = {
  critical: AlertTriangle,
  high: AlertCircle,
  medium: Info,
  low: Info,
};

const severityColors: Record<IncidentSeverity, string> = {
  critical: 'text-red-400 border-red-500/30 bg-red-500/10',
  high: 'text-orange-400 border-orange-500/30 bg-orange-500/10',
  medium: 'text-amber-400 border-amber-500/30 bg-amber-500/10',
  low: 'text-blue-400 border-blue-500/30 bg-blue-500/10',
};

const statusColors: Record<IncidentStatus, { color: string; bgColor: string }> = {
  investigating: { color: 'text-red-400', bgColor: 'bg-red-500/10' },
  identified: { color: 'text-amber-400', bgColor: 'bg-amber-500/10' },
  monitoring: { color: 'text-blue-400', bgColor: 'bg-blue-500/10' },
  resolved: { color: 'text-emerald-400', bgColor: 'bg-emerald-500/10' },
};

const statusLabels: Record<IncidentStatus, string> = {
  investigating: 'investigating',
  identified: 'identified',
  monitoring: 'monitoring',
  resolved: 'resolved',
};

function formatDuration(startDate: string, endDate?: string): string {
  const start = new Date(startDate);
  const end = endDate ? new Date(endDate) : new Date();
  const diff = end.getTime() - start.getTime();

  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }
  return `${minutes}m`;
}

function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  });
}

interface IncidentCardProps {
  incident: Incident;
  index: number;
}

function IncidentCard({ incident, index }: IncidentCardProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const { t } = useTranslation();
  const severityColor = severityColors[incident.severity];
  const statusColor = statusColors[incident.status];
  const SeverityIcon = severityIcons[incident.severity];
  const severityLabel = t(`statusPage.${incident.severity}`);
  const statusLabel = t(`statusPage.${statusLabels[incident.status]}`);
  const detailsId = `incident-details-${incident.id}`;

  return (
    <motion.div
      initial={{ opacity: 0, x: -20 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ delay: index * 0.1, duration: 0.3 }}
      className="relative pl-8"
    >
      {/* Timeline line */}
      <div className="absolute left-3 top-0 bottom-0 w-px bg-border-subtle" />

      {/* Timeline dot */}
      <div
        className={cn(
          'absolute left-0 top-4 h-6 w-6 rounded-full border-2 flex items-center justify-center',
          'bg-bg-primary',
          severityColor
        )}
      >
        <SeverityIcon className="h-3 w-3" />
      </div>

      {/* Card */}
      <Card
        className={cn(
          'mb-4 transition-all duration-200',
          'hover:border-border-default',
          isExpanded && 'ring-1 ring-border-default'
        )}
      >
        <CardHeader
          className="pb-3 cursor-pointer"
          role="button"
          tabIndex={0}
          aria-expanded={isExpanded}
          aria-controls={detailsId}
          onClick={() => setIsExpanded(!isExpanded)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault();
              setIsExpanded(!isExpanded);
            }
          }}
        >
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 flex-wrap">
                <Badge
                  variant="secondary"
                  className={cn('text-xs', statusColor.bgColor, statusColor.color)}
                >
                  {statusLabel}
                </Badge>
                <Badge
                  variant="outline"
                  className={cn('text-xs', severityColor)}
                >
                  {severityLabel}
                </Badge>
              </div>
              <h3 className="mt-2 font-semibold text-text-primary line-clamp-1">
                {incident.title}
              </h3>
              <p className="text-sm text-text-secondary line-clamp-1">
                {incident.description}
              </p>
            </div>
            <Button
              variant="ghost"
              size="sm"
              className="h-8 w-8 p-0 shrink-0"
              onClick={(e) => {
                e.stopPropagation();
                setIsExpanded(!isExpanded);
              }}
            >
              <motion.div
                animate={{ rotate: isExpanded ? 180 : 0 }}
                transition={{ duration: 0.2 }}
              >
                <ChevronDown className="h-4 w-4 text-text-muted" />
              </motion.div>
            </Button>
          </div>

          {/* Meta info */}
          <div className="mt-3 flex flex-wrap items-center gap-4 text-xs text-text-muted">
            <span className="flex items-center gap-1">
              <Calendar className="h-3 w-3" />
              {t('statusPage.started')} {formatDate(incident.created_at)}
            </span>
            <span className="flex items-center gap-1">
              <Clock className="h-3 w-3" />
              {incident.status === 'resolved'
                ? t('statusPage.resolvedAfter', { duration: formatDuration(incident.created_at, incident.resolved_at) })
                : t('statusPage.ongoingFor', { duration: formatDuration(incident.created_at) })}
            </span>
            {incident.affected_components.length > 0 && (
              <span className="flex items-center gap-1">
                <span className="text-text-secondary">
                  {t('statusPage.affects')} {incident.affected_components.join(', ')}
                </span>
              </span>
            )}
          </div>
        </CardHeader>

        <AnimatePresence>
          {isExpanded && (
            <motion.div
              id={detailsId}
              initial={{ height: 0, opacity: 0 }}
              animate={{ height: 'auto', opacity: 1 }}
              exit={{ height: 0, opacity: 0 }}
              transition={{ duration: 0.2 }}
              className="overflow-hidden"
            >
              <CardContent className="pt-0">
                <div className="border-t border-border-subtle pt-4">
                  <h4 className="text-sm font-medium text-text-primary mb-2">
                    {t('statusPage.description')}
                  </h4>
                  <p className="text-sm text-text-secondary whitespace-pre-wrap">
                    {incident.description}
                  </p>

                  {incident.updates && incident.updates.length > 0 && (
                    <div className="mt-4">
                      <h4 className="text-sm font-medium text-text-primary mb-2">
                        {t('statusPage.updates')}
                      </h4>
                      <div className="space-y-3">
                        {incident.updates.map((update, idx) => (
                          <div
                            key={update.id}
                            className="flex gap-3 text-sm"
                          >
                            <div className="flex flex-col items-center">
                              <div className="h-2 w-2 rounded-full bg-brand-500" />
                              {idx < incident.updates!.length - 1 && (
                                <div className="flex-1 w-px bg-border-subtle my-1" />
                              )}
                            </div>
                            <div className="flex-1 pb-3">
                              <p className="text-text-secondary">{update.message}</p>
                              <p className="text-xs text-text-muted mt-1">
                                {formatDate(update.created_at)} · {t(`statusPage.${statusLabels[update.status]}`)}
                              </p>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}

                  {incident.resolved_at && (
                    <div className="mt-4 flex items-center gap-2 rounded-lg bg-emerald-500/10 p-3">
                      <CheckCircle2 className="h-4 w-4 text-emerald-400" />
                      <span className="text-sm text-emerald-400">
                        {t('statusPage.resolvedOn', { date: formatDate(incident.resolved_at) })}
                      </span>
                    </div>
                  )}
                </div>
              </CardContent>
            </motion.div>
          )}
        </AnimatePresence>
      </Card>
    </motion.div>
  );
}

function IncidentSkeleton({ index }: { index: number }) {
  return (
    <div className="relative pl-8">
      <div className="absolute left-3 top-0 bottom-0 w-px bg-border-subtle" />
      <Skeleton className="absolute left-0 top-4 h-6 w-6 rounded-full" />
      <Card className="mb-4">
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1 space-y-2">
              <div className="flex gap-2">
                <Skeleton className="h-5 w-20" />
                <Skeleton className="h-5 w-16" />
              </div>
              <Skeleton className="h-5 w-3/4" />
              <Skeleton className="h-4 w-1/2" />
            </div>
          </div>
          <div className="mt-3 flex gap-4">
            <Skeleton className="h-3 w-24" />
            <Skeleton className="h-3 w-32" />
          </div>
        </CardHeader>
      </Card>
    </div>
  );
}

export function IncidentTimeline({
  incidents,
  isLoading,
  showFilters = true,
  maxItems,
}: IncidentTimelineProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<IncidentStatus | 'all'>('all');
  const [severityFilter, setSeverityFilter] = useState<IncidentSeverity | 'all'>('all');
  const { t } = useTranslation();

  // Filter incidents
  const filteredIncidents = incidents.filter((incident) => {
    const matchesSearch =
      !searchQuery ||
      incident.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      incident.description.toLowerCase().includes(searchQuery.toLowerCase());

    const matchesStatus = statusFilter === 'all' || incident.status === statusFilter;
    const matchesSeverity =
      severityFilter === 'all' || incident.severity === severityFilter;

    return matchesSearch && matchesStatus && matchesSeverity;
  });

  const displayedIncidents = maxItems
    ? filteredIncidents.slice(0, maxItems)
    : filteredIncidents;

  const hasActiveFilters =
    searchQuery || statusFilter !== 'all' || severityFilter !== 'all';

  const clearFilters = () => {
    setSearchQuery('');
    setStatusFilter('all');
    setSeverityFilter('all');
  };

  return (
    <section aria-label={t('statusPage.incidentTimeline')}>
      <div className="mb-6">
        <div className="flex items-center justify-between flex-wrap gap-4">
          <div>
            <h2 className="text-xl font-semibold text-text-primary">{t('statusPage.recentIncidents')}</h2>
            <p className="mt-1 text-sm text-text-secondary">
              {t('statusPage.incidentHistory')}
            </p>
          </div>
          {incidents.filter((i) => i.status !== 'resolved').length > 0 && (
            <Badge variant="destructive" className="animate-pulse">
              {incidents.filter((i) => i.status !== 'resolved').length} {t('statusPage.active')}
            </Badge>
          )}
        </div>

        {showFilters && (
          <div className="mt-4 flex flex-wrap gap-3">
            <div className="relative flex-1 min-w-[200px] max-w-sm">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-text-muted" />
              <Input
                placeholder={t('statusPage.searchIncidents')}
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-9"
              />
              {searchQuery && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="absolute right-1 top-1/2 -translate-y-1/2 h-6 w-6 p-0"
                  onClick={() => setSearchQuery('')}
                >
                  <X className="h-3 w-3" />
                </Button>
              )}
            </div>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value as IncidentStatus | 'all')}
              className="h-10 rounded-md border border-border-subtle bg-bg-secondary px-3 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-brand-500"
            >
              <option value="all">{t('statusPage.allStatuses')}</option>
              <option value="investigating">{t('statusPage.investigating')}</option>
              <option value="identified">{t('statusPage.identified')}</option>
              <option value="monitoring">{t('statusPage.monitoring')}</option>
              <option value="resolved">{t('statusPage.resolved')}</option>
            </select>
            <select
              value={severityFilter}
              onChange={(e) => setSeverityFilter(e.target.value as IncidentSeverity | 'all')}
              className="h-10 rounded-md border border-border-subtle bg-bg-secondary px-3 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-brand-500"
            >
              <option value="all">{t('statusPage.allSeverities')}</option>
              <option value="critical">{t('statusPage.critical')}</option>
              <option value="high">{t('statusPage.high')}</option>
              <option value="medium">{t('statusPage.medium')}</option>
              <option value="low">{t('statusPage.low')}</option>
            </select>
            {hasActiveFilters && (
              <Button variant="ghost" size="sm" onClick={clearFilters}>
                {t('statusPage.clearFilters')}
              </Button>
            )}
          </div>
        )}
      </div>

      {isLoading ? (
        <div>
          {[1, 2, 3].map((i) => (
            <IncidentSkeleton key={i} index={i} />
          ))}
        </div>
      ) : displayedIncidents.length === 0 ? (
        <Card className="p-8 text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-emerald-500/10">
            <CheckCircle2 className="h-6 w-6 text-emerald-400" />
          </div>
          <h3 className="font-medium text-text-primary">{t('statusPage.noIncidentsFound')}</h3>
          <p className="mt-1 text-sm text-text-secondary">
            {hasActiveFilters
              ? t('statusPage.adjustFilters')
              : t('statusPage.allSystemsRunning')}
          </p>
        </Card>
      ) : (
        <div>
          {displayedIncidents.map((incident, index) => (
            <IncidentCard key={incident.id} incident={incident} index={index} />
          ))}
        </div>
      )}
    </section>
  );
}
