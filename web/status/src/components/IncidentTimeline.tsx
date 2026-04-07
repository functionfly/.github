import { cn } from "@/lib/utils";
import { Clock, AlertCircle, CheckCircle2, Wrench, ExternalLink } from "lucide-react";
import { Link } from "react-router-dom";

interface Incident {
  id: string;
  title: string;
  status: "investigating" | "identified" | "monitoring" | "resolved";
  severity: "minor" | "major" | "critical";
  affectedServices: string[];
  createdAt: string;
  resolvedAt?: string;
  updates: IncidentUpdate[];
}

interface IncidentUpdate {
  id: string;
  status: string;
  message: string;
  timestamp: string;
}

interface IncidentTimelineProps {
  incidents: Incident[];
  className?: string;
}

const statusConfig = {
  investigating: {
    icon: AlertCircle,
    color: "text-amber-400",
    bg: "bg-amber-500/20",
    border: "border-amber-500/30",
    label: "Investigating",
  },
  identified: {
    icon: AlertCircle,
    color: "text-orange-400",
    bg: "bg-orange-500/20",
    border: "border-orange-500/30",
    label: "Identified",
  },
  monitoring: {
    icon: Clock,
    color: "text-blue-400",
    bg: "bg-blue-500/20",
    border: "border-blue-500/30",
    label: "Monitoring",
  },
  resolved: {
    icon: CheckCircle2,
    color: "text-emerald-400",
    bg: "bg-emerald-500/20",
    border: "border-emerald-500/30",
    label: "Resolved",
  },
};

const severityConfig = {
  minor: {
    color: "text-yellow-400",
    bg: "bg-yellow-500/20",
    label: "Minor",
  },
  major: {
    color: "text-orange-400",
    bg: "bg-orange-500/20",
    label: "Major",
  },
  critical: {
    color: "text-red-400",
    bg: "bg-red-500/20",
    label: "Critical",
  },
};

function IncidentCard({ incident, index }: { incident: Incident; index: number }) {
  const status = statusConfig[incident.status];
  const severity = severityConfig[incident.severity];
  const StatusIcon = status.icon;
  
  const formatTime = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(hours / 24);
    
    if (days > 0) return `${days}d ago`;
    if (hours > 0) return `${hours}h ago`;
    return "Just now";
  };

  return (
    <div 
      className={cn(
        "relative glass-card p-6",
        "animate-fade-in-up transition-all duration-300 hover-lift"
      )}
      style={{ animationDelay: `${index * 100}ms` }}
    >
      {/* Status indicator line */}
      <div className={cn(
        "absolute left-0 top-0 bottom-0 w-1 rounded-l-xl",
        incident.status === "resolved" ? "bg-emerald-500" : 
        incident.severity === "critical" ? "bg-red-500" : 
        incident.severity === "major" ? "bg-orange-500" : "bg-yellow-500"
      )} />
      
      <div className="pl-4">
        {/* Header */}
        <div className="flex items-start justify-between gap-4">
          <div className="flex-1">
            <div className="flex items-center gap-2 mb-2">
              <span className={cn(
                "inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium border",
                status.bg,
                status.color,
                status.border
              )}>
                <StatusIcon className="w-3 h-3" />
                {status.label}
              </span>
              <span className={cn(
                "inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium",
                severity.bg,
                severity.color
              )}>
                {severity.label}
              </span>
            </div>
            
            <h3 className="text-lg font-semibold text-text-primary mb-2">
              {incident.title}
            </h3>
            
            {/* Affected services */}
            <div className="flex flex-wrap gap-1.5 mb-3">
              {incident.affectedServices.map((service) => (
                <span 
                  key={service}
                  className="px-2 py-0.5 rounded-md text-xs bg-bg-elevated text-text-secondary"
                >
                  {service}
                </span>
              ))}
            </div>
          </div>
          
          <div className="text-right">
            <p className="text-text-muted text-sm">{formatTime(incident.createdAt)}</p>
            {incident.resolvedAt && (
              <p className="text-emerald-400 text-xs mt-1">
                Resolved {formatTime(incident.resolvedAt)}
              </p>
            )}
          </div>
        </div>
        
        {/* Latest update */}
        {incident.updates.length > 0 && (
          <div className="mt-4 p-3 rounded-lg bg-bg-elevated/50 border border-border-subtle">
            <p className="text-text-secondary text-sm">
              <span className="text-text-muted font-medium">
                {new Date(incident.updates[incident.updates.length - 1].timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}:
              </span>{" "}
              {incident.updates[incident.updates.length - 1].message}
            </p>
          </div>
        )}
        
        {/* View details link */}
        <Link
          to={`/incidents/${incident.id}`}
          className="inline-flex items-center gap-1.5 mt-4 text-sm text-brand-400 hover:text-brand-300 transition-colors"
        >
          View details
          <ExternalLink className="w-3.5 h-3.5" />
        </Link>
      </div>
    </div>
  );
}

export function IncidentTimeline({ incidents, className }: IncidentTimelineProps) {
  const activeIncidents = incidents.filter(i => i.status !== "resolved");
  const resolvedIncidents = incidents.filter(i => i.status === "resolved").slice(0, 5);
  
  return (
    <div className={cn("space-y-6", className)}>
      {/* Active Incidents */}
      {activeIncidents.length > 0 && (
        <div>
          <div className="flex items-center gap-2 mb-4">
            <div className="w-2 h-2 rounded-full bg-amber-500 animate-pulse" />
            <h2 className="text-lg font-semibold text-text-primary">
              Active Incidents
            </h2>
            <span className="text-text-muted text-sm">({activeIncidents.length})</span>
          </div>
          
          <div className="space-y-4">
            {activeIncidents.map((incident, index) => (
              <IncidentCard 
                key={incident.id} 
                incident={incident} 
                index={index}
              />
            ))}
          </div>
        </div>
      )}
      
      {/* Recent Resolved */}
      {resolvedIncidents.length > 0 && (
        <div>
          <div className="flex items-center gap-2 mb-4">
            <CheckCircle2 className="w-5 h-5 text-emerald-400" />
            <h2 className="text-lg font-semibold text-text-primary">
              Recently Resolved
            </h2>
          </div>
          
          <div className="space-y-4">
            {resolvedIncidents.map((incident, index) => (
              <IncidentCard 
                key={incident.id} 
                incident={incident} 
                index={index + activeIncidents.length}
              />
            ))}
          </div>
        </div>
      )}
      
      {/* Empty state */}
      {incidents.length === 0 && (
        <div className="glass-card p-8 text-center">
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-emerald-500/20 mb-4">
            <CheckCircle2 className="w-8 h-8 text-emerald-400" />
          </div>
          <h3 className="text-lg font-semibold text-text-primary mb-2">
            No Active Incidents
          </h3>
          <p className="text-text-secondary">
            All systems are operating normally.
          </p>
        </div>
      )}
    </div>
  );
}

// Maintenance schedule component
interface MaintenanceItem {
  id: string;
  title: string;
  scheduledFor: string;
  duration: string;
  affectedServices: string[];
  description: string;
}

interface MaintenanceScheduleProps {
  items: MaintenanceItem[];
  className?: string;
}

export function MaintenanceSchedule({ items, className }: MaintenanceScheduleProps) {
  return (
    <div className={cn("space-y-4", className)}>
      {items.map((item, index) => (
        <div 
          key={item.id}
          className={cn(
            "glass-card p-5",
            "animate-fade-in-up"
          )}
          style={{ animationDelay: `${index * 100}ms` }}
        >
          <div className="flex items-start gap-4">
            <div className="p-2.5 rounded-xl bg-purple-500/20 border border-purple-500/30">
              <Wrench className="w-5 h-5 text-purple-400" />
            </div>
            
            <div className="flex-1 min-w-0">
              <h4 className="font-medium text-text-primary mb-1">{item.title}</h4>
              <p className="text-text-secondary text-sm mb-3">{item.description}</p>
              
              <div className="flex flex-wrap gap-2">
                <span className="inline-flex items-center gap-1 px-2 py-1 rounded-md text-xs bg-bg-elevated text-text-secondary">
                  <Clock className="w-3 h-3" />
                  {new Date(item.scheduledFor).toLocaleString([], { 
                    month: 'short', 
                    day: 'numeric',
                    hour: '2-digit',
                    minute: '2-digit'
                  })}
                </span>
                <span className="inline-flex items-center gap-1 px-2 py-1 rounded-md text-xs bg-bg-elevated text-text-secondary">
                  Duration: {item.duration}
                </span>
              </div>
              
              <div className="flex flex-wrap gap-1.5 mt-3">
                {item.affectedServices.map((service) => (
                  <span 
                    key={service}
                    className="px-2 py-0.5 rounded-md text-xs bg-purple-500/10 text-purple-300 border border-purple-500/20"
                  >
                    {service}
                  </span>
                ))}
              </div>
            </div>
          </div>
        </div>
      ))}
      
      {items.length === 0 && (
        <div className="glass-card p-6 text-center text-text-secondary">
          <Wrench className="w-8 h-8 mx-auto mb-3 text-text-muted" />
          <p>No scheduled maintenance</p>
        </div>
      )}
    </div>
  );
}
