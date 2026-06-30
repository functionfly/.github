import { Activity, Globe, Wifi } from 'lucide-react';
import { useEffect, useState } from 'react';
import { API_URLS } from '@/lib/api-urls';
import type { PlatformMetrics } from '../hooks/usePlatformStream';

interface PlatformComponent {
  name: string;
  status: string;
  uptime_24h?: number;
  response_time_ms?: number;
}

interface PlatformStateProps {
  metrics: PlatformMetrics;
}

export function PlatformState({ metrics }: PlatformStateProps) {
  const [components, setComponents] = useState<PlatformComponent[]>([]);
  const [edgeNodes, setEdgeNodes] = useState<{ host: string; region: string; status: string }[]>([]);

  useEffect(() => {
    fetch(API_URLS.platform.statusComponents)
      .then((r) => r.json())
      .then((data) => {
        if (data.components) setComponents(data.components);
      })
      .catch(() => {});
    fetch(API_URLS.platform.statusEdge)
      .then((r) => r.json())
      .then((data) => {
        if (data.nodes) setEdgeNodes(data.nodes);
      })
      .catch(() => {});
  }, []);

  const statusColor = (s: string) => {
    if (s === 'operational' || s === 'ok') return 'ok';
    if (s === 'degraded' || s === 'warning') return 'warning';
    return 'error';
  };

  return (
    <section className="founders-section">
      <div className="founders-section__title">
        <Activity size={14} />
        Platform State
      </div>

      <div className="platform-header">
        <div className="platform-header__metric">
          <span className="platform-header__value">{metrics.uptime.toFixed(2)}%</span>
          <span className="platform-header__label">Uptime</span>
        </div>
        <div className="platform-header__metric">
          <span className="platform-header__value">{metrics.latency}ms</span>
          <span className="platform-header__label">Latency</span>
        </div>
        <div className="platform-header__metric">
          <span className={`platform-header__badge platform-header__badge--${statusColor(metrics.status)}`}>
            {metrics.status.toUpperCase()}
          </span>
        </div>
        {edgeNodes.length > 0 && (
          <div className="platform-header__metric">
            <span className="platform-header__value">
              <Globe size={14} style={{ display: 'inline', verticalAlign: 'middle', marginRight: 4 }} />
              {edgeNodes.filter((n) => n.status === 'ok').length}
            </span>
            <span className="platform-header__label">Edge Nodes</span>
          </div>
        )}
      </div>

      {components.length > 0 && (
        <div className="platform-grid">
          {components.slice(0, 20).map((comp) => (
            <div key={comp.name} className="platform-tile">
              <div className={`platform-tile__dot platform-tile__dot--${statusColor(comp.status)}`} />
              <span className="platform-tile__name">{comp.name}</span>
              {comp.uptime_24h != null && (
                <span className="platform-tile__uptime">{comp.uptime_24h.toFixed(1)}%</span>
              )}
              {comp.response_time_ms != null && (
                <span className="platform-tile__latency">{comp.response_time_ms}ms</span>
              )}
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
