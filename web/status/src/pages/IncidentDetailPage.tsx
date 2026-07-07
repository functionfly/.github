import { Chamber, CornerBrace, StatusPill, FrameButton, SealedButton, PageGrid, ReducedMotionGate, AnnotationTag } from '@/components/containment';
import { Nav } from '@/components/Nav';
import { Footer } from '@/components/Footer';
import {
  statusAPI,
  type Incident,
  type IncidentUpdate,
} from '@/lib/api';
import { useQuery } from '@tanstack/react-query';
import { format, formatDistanceToNow, parseISO } from 'date-fns';
import {
  Activity,
  AlertCircle,
  AlertTriangle,
  ArrowLeft,
  Bell,
  CheckCircle,
  Clock,
  ExternalLink,
  MessageSquare,
  Server,
} from 'lucide-react';
import { useState } from 'react';
import { Link, useParams } from 'react-router-dom';

function mapIncidentStatus(s: string): 'live' | 'pending' | 'revoked' {
  if (s === 'resolved') return 'live';
  if (s === 'monitoring' || s === 'identified') return 'pending';
  return 'revoked';
}

function TimelineItem({ update, isLast }: { update: IncidentUpdate; isLast: boolean }) {
  const statusColors: Record<string, string> = {
    investigating: 'var(--status-pending)',
    identified: 'var(--foil-a)',
    monitoring: 'var(--foil-b)',
    resolved: 'var(--status-ok)',
  };

  const color = statusColors[update.status] || 'var(--status-pending)';

  return (
    <div className="relative flex gap-4">
      <div className="flex flex-col items-center">
        <div
          className="flex items-center justify-center rounded-full"
          style={{
            width: '40px',
            height: '40px',
            border: `1px solid ${color}`,
            background: `${color}10`,
            color,
          }}
        >
          {update.status === 'resolved' ? <CheckCircle style={{ width: 20, height: 20 }} /> :
           update.status === 'monitoring' ? <Activity style={{ width: 20, height: 20 }} /> :
           update.status === 'identified' ? <AlertTriangle style={{ width: 20, height: 20 }} /> :
           <AlertCircle style={{ width: 20, height: 20 }} />}
        </div>
        {!isLast && <div style={{ width: '2px', flex: 1, background: 'var(--panel-edge)', margin: 'var(--space-2) 0' }} />}
      </div>

      <div className="flex-1" style={{ paddingBottom: isLast ? 0 : 'var(--space-6)' }}>
        <div className="flex items-center gap-3" style={{ marginBottom: 'var(--space-2)' }}>
          <span style={{ fontWeight: 600, fontSize: '14px', color }}>
            {update.status.charAt(0).toUpperCase() + update.status.slice(1)}
          </span>
          <span style={{ fontSize: '12px', color: 'var(--text-faint)' }}>
            {formatDistanceToNow(parseISO(update.created_at), { addSuffix: true })}
          </span>
          <span style={{ fontSize: '12px', color: 'var(--text-faint)' }}>&bull;</span>
          <span style={{ fontSize: '12px', color: 'var(--text-faint)' }}>
            {format(parseISO(update.created_at), 'MMM d, HH:mm')}
          </span>
        </div>

        <p style={{ color: 'var(--text)', lineHeight: 1.6 }}>{update.message}</p>

        {update.created_by && (
          <div className="flex items-center gap-2" style={{ marginTop: 'var(--space-2)', fontSize: '12px', color: 'var(--text-faint)' }}>
            <span
              className="flex items-center justify-center rounded-full"
              style={{ width: '20px', height: '20px', background: 'var(--foil-a)', color: 'var(--bg)', fontSize: '10px', fontWeight: 600 }}
            >
              {update.created_by.name.charAt(0)}
            </span>
            <span>{update.created_by.name}</span>
          </div>
        )}
      </div>
    </div>
  );
}

export default function IncidentDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [showSubscribe, setShowSubscribe] = useState(false);

  const { data: incident, isLoading, error } = useQuery<Incident, Error>({
    queryKey: ['incident', id],
    queryFn: () => (id ? statusAPI.getIncident(id) : Promise.reject('No ID provided')),
    enabled: !!id,
    retry: 2,
  });

  if (isLoading) {
    return (
      <ReducedMotionGate>
        <PageGrid />
        <div className="min-h-screen" style={{ background: 'var(--bg)', position: 'relative', zIndex: 1 }}>
          <Nav />
          <main style={{ paddingTop: '120px', paddingBottom: 'var(--space-8)' }}>
            <div className="mx-auto" style={{ maxWidth: '860px', padding: '0 var(--space-7)' }}>
              <Chamber>
                <div className="animate-pulse space-y-4">
                  <div style={{ height: 24, width: 200, background: 'var(--panel-raised)', borderRadius: 'var(--radius)' }} />
                  <div style={{ height: 32, width: '75%', background: 'var(--panel-raised)', borderRadius: 'var(--radius)' }} />
                  <div style={{ height: 16, width: '100%', background: 'var(--panel-raised)', borderRadius: 'var(--radius)' }} />
                </div>
              </Chamber>
            </div>
          </main>
        </div>
      </ReducedMotionGate>
    );
  }

  if (error || !incident) {
    return (
      <ReducedMotionGate>
        <PageGrid />
        <div className="min-h-screen flex items-center justify-center" style={{ background: 'var(--bg)', position: 'relative', zIndex: 1 }}>
          <Chamber>
            <div className="text-center">
              <AlertCircle style={{ width: 48, height: 48, color: 'var(--status-pending)', margin: '0 auto var(--space-4)' }} />
              <h2 style={{ fontFamily: 'var(--font-display)', fontSize: '22px', fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                Incident Not Found
              </h2>
              <p style={{ fontSize: '15px', color: 'var(--text-dim)', marginBottom: 'var(--space-6)' }}>
                The incident you're looking for doesn't exist or has been removed.
              </p>
              <SealedButton onClick={() => window.location.href = '/'}>
                Return to Status
              </SealedButton>
            </div>
          </Chamber>
        </div>
      </ReducedMotionGate>
    );
  }

  const isResolved = incident.status === 'resolved';
  const duration = incident.resolved_at
    ? Math.round(
        (new Date(incident.resolved_at).getTime() - new Date(incident.created_at).getTime()) / (1000 * 60),
      )
    : undefined;

  return (
    <ReducedMotionGate>
      <PageGrid />
      <div className="min-h-screen" style={{ background: 'var(--bg)', position: 'relative', zIndex: 1 }}>
        <Nav />

        <main style={{ paddingTop: '120px', paddingBottom: 'var(--space-8)' }}>
          <div className="mx-auto space-y-6" style={{ maxWidth: '860px', padding: '0 var(--space-7)' }}>
            {/* Back link */}
            <Link to="/" className="inline-flex items-center gap-2" style={{ color: 'var(--text-dim)', textDecoration: 'none', fontSize: '14px' }}>
              <ArrowLeft style={{ width: 14, height: 14 }} />
              Back to Status
            </Link>

            {/* Incident Header */}
            <Chamber>
              <CornerBrace position="tl" />
              <CornerBrace position="br" />
              <AnnotationTag label="INCIDENT" detail={incident.id.slice(0, 8)} />

              <div className="flex flex-col md:flex-row md:items-start gap-6">
                <div
                  className="flex items-center justify-center shrink-0 rounded-lg"
                  style={{
                    width: '64px',
                    height: '64px',
                    background: isResolved ? 'rgba(143,255,208,0.1)' : 'rgba(232,196,104,0.1)',
                  }}
                >
                  {isResolved ? (
                    <CheckCircle style={{ width: 32, height: 32, color: 'var(--status-ok)' }} />
                  ) : (
                    <AlertTriangle style={{ width: 32, height: 32, color: 'var(--status-pending)' }} />
                  )}
                </div>

                <div className="flex-1 min-w-0">
                  <div className="flex flex-wrap items-center gap-2" style={{ marginBottom: 'var(--space-3)' }}>
                    <StatusPill status={mapIncidentStatus(incident.status)} label={incident.status} />
                    <span
                      className="uppercase"
                      style={{
                        fontFamily: 'var(--font-mono)',
                        fontSize: '11px',
                        fontWeight: 500,
                        letterSpacing: '0.06em',
                        color: 'var(--status-pending)',
                        padding: '3px 8px',
                        borderRadius: 'var(--radius-sm)',
                        border: '1px solid rgba(232,196,104,0.3)',
                        background: 'rgba(232,196,104,0.06)',
                      }}
                    >
                      {incident.severity}
                    </span>
                  </div>

                  <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 'clamp(22px, 3vw, 32px)', fontWeight: 700, color: 'var(--text)', marginBottom: 'var(--space-2)' }}>
                    {incident.title}
                  </h1>

                  <p style={{ fontSize: '15px', color: 'var(--text-dim)', lineHeight: 1.6 }}>
                    {incident.description}
                  </p>
                </div>
              </div>

              {/* Stats */}
              <div
                className="grid grid-cols-2 md:grid-cols-4 gap-4"
                style={{ marginTop: 'var(--space-6)', paddingTop: 'var(--space-6)', borderTop: '1px solid var(--panel-edge)' }}
              >
                <div>
                  <div className="flex items-center gap-2" style={{ color: 'var(--text-faint)', fontSize: '13px', marginBottom: 'var(--space-1)' }}>
                    <Clock style={{ width: 14, height: 14 }} />
                    Started
                  </div>
                  <div style={{ color: 'var(--text)', fontWeight: 500 }}>{format(parseISO(incident.created_at), 'MMM d, HH:mm')}</div>
                  <div style={{ fontSize: '12px', color: 'var(--text-faint)' }}>
                    {formatDistanceToNow(parseISO(incident.created_at), { addSuffix: true })}
                  </div>
                </div>

                {incident.resolved_at && (
                  <>
                    <div>
                      <div className="flex items-center gap-2" style={{ color: 'var(--text-faint)', fontSize: '13px', marginBottom: 'var(--space-1)' }}>
                        <CheckCircle style={{ width: 14, height: 14 }} />
                        Resolved
                      </div>
                      <div style={{ color: 'var(--status-ok)', fontWeight: 500 }}>{format(parseISO(incident.resolved_at), 'MMM d, HH:mm')}</div>
                      <div style={{ fontSize: '12px', color: 'var(--text-faint)' }}>
                        {formatDistanceToNow(parseISO(incident.resolved_at), { addSuffix: true })}
                      </div>
                    </div>
                    <div>
                      <div className="flex items-center gap-2" style={{ color: 'var(--text-faint)', fontSize: '13px', marginBottom: 'var(--space-1)' }}>
                        <Activity style={{ width: 14, height: 14 }} />
                        Duration
                      </div>
                      <div style={{ color: 'var(--text)', fontWeight: 500 }}>{duration} minutes</div>
                      <div style={{ fontSize: '12px', color: 'var(--text-faint)' }}>~{Math.round(((duration || 0) / 60) * 10) / 10} hours</div>
                    </div>
                  </>
                )}

                <div>
                  <div className="flex items-center gap-2" style={{ color: 'var(--text-faint)', fontSize: '13px', marginBottom: 'var(--space-1)' }}>
                    <Server style={{ width: 14, height: 14 }} />
                    Affected
                  </div>
                  <div style={{ color: 'var(--text)', fontWeight: 500 }}>
                    {incident.affected_components?.length || 0} {incident.affected_components?.length === 1 ? 'service' : 'services'}
                  </div>
                  <div style={{ fontSize: '12px', color: 'var(--text-faint)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {incident.affected_components?.join(', ') || 'None'}
                  </div>
                </div>
              </div>
            </Chamber>

            {/* Affected Services */}
            {incident.affected_components && incident.affected_components.length > 0 && (
              <Chamber nested>
                <h3 className="flex items-center gap-2" style={{ fontFamily: 'var(--font-display)', fontSize: '18px', fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-4)' }}>
                  <Server style={{ width: 18, height: 18, color: 'var(--foil-a)' }} />
                  Affected Services
                </h3>
                <div className="flex flex-wrap gap-2">
                  {incident.affected_components.map((service) => (
                    <div
                      key={service}
                      className="flex items-center gap-2"
                      style={{
                        padding: 'var(--space-2) var(--space-4)',
                        background: 'var(--panel)',
                        border: '1px solid var(--panel-edge)',
                        borderRadius: 'var(--radius)',
                        fontSize: '14px',
                        color: 'var(--text)',
                        fontWeight: 500,
                      }}
                    >
                      <span
                        className="inline-block rounded-full"
                        style={{
                          width: '6px',
                          height: '6px',
                          background: isResolved ? 'var(--status-ok)' : 'var(--status-pending)',
                        }}
                      />
                      {service.replace(/-/g, ' ')}
                      <span style={{ fontSize: '12px', color: isResolved ? 'var(--status-ok)' : 'var(--status-pending)' }}>
                        {isResolved ? 'Operational' : 'Affected'}
                      </span>
                    </div>
                  ))}
                </div>
              </Chamber>
            )}

            {/* Timeline */}
            <Chamber>
              <CornerBrace position="tr" />
              <CornerBrace position="bl" />
              <h3 className="flex items-center gap-2" style={{ fontFamily: 'var(--font-display)', fontSize: '18px', fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-5)' }}>
                <MessageSquare style={{ width: 18, height: 18, color: 'var(--foil-a)' }} />
                Incident Timeline
              </h3>

              {incident.updates?.length ? (
                <div>
                  {incident.updates.map((update, index) => (
                    <TimelineItem key={update.id} update={update} isLast={index === incident.updates!.length - 1} />
                  ))}
                </div>
              ) : (
                <div className="flex items-center justify-center" style={{ minHeight: 120 }}>
                  <p style={{ fontFamily: 'var(--font-mono)', fontSize: '11px', textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--text-faint)' }}>
                    No updates available
                  </p>
                </div>
              )}
            </Chamber>

            {/* Subscribe CTA */}
            <Chamber nested>
              <div className="flex flex-col md:flex-row items-center justify-between gap-4">
                <div>
                  <h3 style={{ fontFamily: 'var(--font-display)', fontSize: '18px', fontWeight: 500, color: 'var(--text)', marginBottom: 'var(--space-1)' }}>
                    Stay informed about incidents
                  </h3>
                  <p style={{ fontSize: '14px', color: 'var(--text-dim)' }}>
                    Subscribe to get notified when we create, update or resolve incidents.
                  </p>
                </div>
                <div className="flex gap-3">
                  <FrameButton onClick={() => setShowSubscribe(!showSubscribe)} iconLeft={<Bell style={{ width: 14, height: 14 }} />}>
                    Subscribe
                  </FrameButton>
                  <SealedButton onClick={() => window.open('/api/v1/status/rss', '_blank')} iconLeft={<ExternalLink style={{ width: 14, height: 14 }} />}>
                    RSS Feed
                  </SealedButton>
                </div>
              </div>

              {showSubscribe && (
                <div style={{ marginTop: 'var(--space-4)', paddingTop: 'var(--space-4)', borderTop: '1px solid var(--panel-edge)' }}>
                  <div className="flex gap-3">
                    <input
                      type="email"
                      placeholder="Enter your email"
                      style={{
                        flex: 1,
                        fontFamily: 'var(--font-body)',
                        fontSize: '14px',
                        color: 'var(--text)',
                        background: 'var(--panel)',
                        border: '1px solid var(--steel)',
                        borderRadius: 'var(--radius)',
                        padding: 'var(--space-3) var(--space-4)',
                        outline: 'none',
                      }}
                    />
                    <SealedButton>Subscribe</SealedButton>
                  </div>
                </div>
              )}
            </Chamber>
          </div>
        </main>

        <Footer />
      </div>
    </ReducedMotionGate>
  );
}
