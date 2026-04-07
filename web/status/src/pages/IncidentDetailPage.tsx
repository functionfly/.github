import { Logo } from "@/components/Logo";
import { StatusBadge, StatusDot } from "@/components/StatusBadge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  getSeverityColor,
  getStatusColor,
  statusAPI,
  type Incident,
  type IncidentUpdate,
} from "@/lib/api";
import { cn } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { format, formatDistanceToNow, parseISO } from "date-fns";
import { AnimatePresence, motion } from "framer-motion";
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
  Share2,
} from "lucide-react";
import { useState } from "react";
import { Link, useParams } from "react-router-dom";

const statusConfig = {
  investigating: {
    icon: AlertCircle,
    color: "text-amber-400",
    bg: "bg-amber-500/10",
    border: "border-amber-500/30",
  },
  identified: {
    icon: AlertTriangle,
    color: "text-blue-400",
    bg: "bg-blue-500/10",
    border: "border-blue-500/30",
  },
  monitoring: {
    icon: Activity,
    color: "text-purple-400",
    bg: "bg-purple-500/10",
    border: "border-purple-500/30",
  },
  resolved: {
    icon: CheckCircle,
    color: "text-emerald-400",
    bg: "bg-emerald-500/10",
    border: "border-emerald-500/30",
  },
};

const severityConfig = {
  low: { label: "Low", color: "text-emerald-400", bg: "bg-emerald-500" },
  medium: { label: "Medium", color: "text-amber-400", bg: "bg-amber-500" },
  high: { label: "High", color: "text-orange-400", bg: "bg-orange-500" },
  critical: { label: "Critical", color: "text-red-400", bg: "bg-red-500" },
};

function AnimatedBackground() {
  return (
    <div className="fixed inset-0 pointer-events-none overflow-hidden -z-10">
      <motion.div
        className="absolute w-[500px] h-[500px] rounded-full opacity-10 blur-[100px]"
        style={{
          background:
            "radial-gradient(circle, rgba(99, 102, 241, 0.3) 0%, transparent 70%)",
          top: "20%",
          right: "-10%",
        }}
        animate={{
          x: [0, -50, 0],
          y: [0, 30, 0],
          scale: [1, 1.1, 1],
        }}
        transition={{
          duration: 15,
          repeat: Infinity,
          ease: "easeInOut",
        }}
      />
    </div>
  );
}

function Header() {
  return (
    <motion.header
      className="fixed top-0 left-0 right-0 z-50 bg-bg-glass-strong/90 backdrop-blur-xl border-b border-border-subtle shadow-lg"
      initial={{ y: -100 }}
      animate={{ y: 0 }}
      transition={{ duration: 0.5 }}
    >
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex h-16 items-center justify-between">
          <Link
            to="/"
            className="inline-flex items-center gap-2 text-text-secondary hover:text-text-primary transition-colors group"
          >
            <ArrowLeft className="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
            <span className="text-sm font-medium">Back to Status</span>
          </Link>

          <div className="flex items-center gap-3">
            <Logo size="sm" showText={false} />
            <span className="font-semibold text-text-primary hidden sm:inline">
              Incident Report
            </span>
          </div>

          <div className="flex items-center gap-2">
            <Button variant="ghost" size="sm" className="hidden sm:flex">
              <Share2 className="w-4 h-4 mr-2" />
              Share
            </Button>
          </div>
        </div>
      </div>
    </motion.header>
  );
}

function TimelineItem({
  update,
  isLast,
}: {
  update: IncidentUpdate;
  isLast: boolean;
}) {
  const config =
    statusConfig[update.status as keyof typeof statusConfig] ||
    statusConfig.investigating;
  const Icon = config.icon;

  return (
    <motion.div
      initial={{ opacity: 0, x: -20 }}
      animate={{ opacity: 1, x: 0 }}
      className="relative flex gap-4"
    >
      <div className="flex flex-col items-center">
        <div
          className={cn(
            "w-10 h-10 rounded-full flex items-center justify-center border",
            config.bg,
            config.border,
            config.color,
          )}
        >
          <Icon className="w-5 h-5" />
        </div>
        {!isLast && <div className="w-0.5 flex-1 bg-border-subtle my-2" />}
      </div>

      <div className={cn("flex-1 pb-8", isLast && "pb-0")}>
        <div className="flex items-center gap-3 mb-2">
          <span
            className={cn(
              "font-semibold text-sm",
              getStatusColor(update.status),
            )}
          >
            {update.status.charAt(0).toUpperCase() + update.status.slice(1)}
          </span>
          <span className="text-xs text-text-muted">
            {formatDistanceToNow(parseISO(update.created_at), {
              addSuffix: true,
            })}
          </span>
          <span className="text-xs text-text-muted">•</span>
          <span className="text-xs text-text-muted">
            {format(parseISO(update.created_at), "MMM d, HH:mm")}
          </span>
        </div>

        <p className="text-text-primary leading-relaxed">{update.message}</p>

        {update.created_by && (
          <div className="mt-2 flex items-center gap-2 text-xs text-text-muted">
            <span className="w-5 h-5 rounded-full bg-brand-500/20 flex items-center justify-center text-brand-400 font-medium">
              {update.created_by.name.charAt(0)}
            </span>
            <span>{update.created_by.name}</span>
          </div>
        )}
      </div>
    </motion.div>
  );
}

function LoadingSkeleton() {
  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 space-y-6">
      <Card className="border-border-subtle bg-bg-tertiary/50 p-6">
        <div className="flex flex-col md:flex-row md:items-start gap-6">
          <Skeleton className="w-16 h-16 rounded-2xl" />
          <div className="flex-1">
            <div className="flex gap-2 mb-3">
              <Skeleton className="h-6 w-24 rounded-full" />
              <Skeleton className="h-6 w-24 rounded-full" />
              <Skeleton className="h-6 w-20" />
            </div>
            <Skeleton className="h-8 w-3/4 mb-2" />
            <Skeleton className="h-5 w-full" />
          </div>
        </div>
        <div className="mt-6 pt-6 border-t border-border-subtle grid grid-cols-2 md:grid-cols-4 gap-4">
          {[1, 2, 3, 4].map((i) => (
            <div key={i}>
              <Skeleton className="h-4 w-20 mb-1" />
              <Skeleton className="h-6 w-32" />
              <Skeleton className="h-4 w-24 mt-1" />
            </div>
          ))}
        </div>
      </Card>

      <Card className="border-border-subtle bg-bg-tertiary/50">
        <CardHeader>
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-4 w-64" />
        </CardHeader>
        <CardContent>
          <div className="flex gap-2">
            {[1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-10 w-32 rounded-lg" />
            ))}
          </div>
        </CardContent>
      </Card>

      <Card className="border-border-subtle bg-bg-tertiary/50">
        <CardHeader>
          <Skeleton className="h-6 w-40" />
          <Skeleton className="h-4 w-56" />
        </CardHeader>
        <CardContent>
          <div className="space-y-6">
            {[1, 2, 3].map((i) => (
              <div key={i} className="flex gap-4">
                <Skeleton className="w-10 h-10 rounded-full shrink-0" />
                <div className="flex-1">
                  <Skeleton className="h-5 w-32 mb-2" />
                  <Skeleton className="h-4 w-full" />
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export default function IncidentDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [showSubscribe, setShowSubscribe] = useState(false);

  const {
    data: incident,
    isLoading,
    error,
  } = useQuery<Incident, Error>({
    queryKey: ["incident", id],
    queryFn: () =>
      id ? statusAPI.getIncident(id) : Promise.reject("No ID provided"),
    enabled: !!id,
    retry: 2,
  });

  if (isLoading) {
    return (
      <div className="min-h-screen bg-bg-primary">
        <AnimatedBackground />
        <Header />
        <main className="pt-24 pb-12">
          <LoadingSkeleton />
        </main>
      </div>
    );
  }

  if (error || !incident) {
    return (
      <div className="min-h-screen bg-bg-primary flex items-center justify-center">
        <Card className="border-border-subtle bg-bg-tertiary/50 backdrop-blur-sm p-8 text-center max-w-md">
          <AlertCircle className="w-12 h-12 text-amber-400 mx-auto mb-4" />
          <h2 className="text-xl font-semibold text-text-primary mb-2">
            Incident Not Found
          </h2>
          <p className="text-text-secondary mb-6">
            The incident you're looking for doesn't exist or has been removed.
          </p>
          <Button asChild className="bg-brand-500 hover:bg-brand-600">
            <Link to="/">Return to Status Page</Link>
          </Button>
        </Card>
      </div>
    );
  }

  const severity =
    severityConfig[incident.severity as keyof typeof severityConfig] ||
    severityConfig.medium;
  const isResolved = incident.status === "resolved";
  const duration = incident.resolved_at
    ? Math.round(
        (new Date(incident.resolved_at).getTime() -
          new Date(incident.created_at).getTime()) /
          (1000 * 60),
      )
    : undefined;

  return (
    <div className="min-h-screen bg-bg-primary">
      <AnimatedBackground />
      <Header />

      <main className="pt-24 pb-12">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 space-y-6">
          {/* Incident Header Card */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
          >
            <Card
              className={cn(
                "border overflow-hidden",
                isResolved
                  ? "border-emerald-500/20 bg-gradient-to-br from-emerald-500/10 via-bg-tertiary/50 to-bg-tertiary/50"
                  : "border-amber-500/20 bg-gradient-to-br from-amber-500/10 via-bg-tertiary/50 to-bg-tertiary/50",
                "backdrop-blur-sm",
              )}
            >
              <CardContent className="p-6">
                <div className="flex flex-col md:flex-row md:items-start gap-6">
                  <div
                    className={cn(
                      "w-16 h-16 rounded-2xl flex items-center justify-center shrink-0",
                      isResolved ? "bg-emerald-500/20" : "bg-amber-500/20",
                    )}
                  >
                    {isResolved ? (
                      <CheckCircle className="w-8 h-8 text-emerald-400" />
                    ) : (
                      <AlertTriangle className="w-8 h-8 text-amber-400" />
                    )}
                  </div>

                  <div className="flex-1 min-w-0">
                    <div className="flex flex-wrap items-center gap-2 mb-3">
                      <StatusBadge status={incident.status} size="md" />
                      <span
                        className={cn(
                          "text-xs px-2 py-1 rounded-full border font-medium",
                          getSeverityColor(incident.severity),
                        )}
                      >
                        {severity.label} Severity
                      </span>
                      <span className="text-xs text-text-muted">
                        {incident.id}
                      </span>
                    </div>

                    <h1 className="text-2xl md:text-3xl font-bold text-text-primary mb-2">
                      {incident.title}
                    </h1>

                    <p className="text-text-secondary leading-relaxed">
                      {incident.description}
                    </p>
                  </div>
                </div>

                {/* Timeline Stats */}
                <div className="mt-6 pt-6 border-t border-border-subtle grid grid-cols-2 md:grid-cols-4 gap-4">
                  <div>
                    <div className="flex items-center gap-2 text-text-muted text-sm mb-1">
                      <Clock className="w-4 h-4" />
                      Started
                    </div>
                    <div className="text-text-primary font-medium">
                      {format(parseISO(incident.created_at), "MMM d, HH:mm")}
                    </div>
                    <div className="text-xs text-text-muted">
                      {formatDistanceToNow(parseISO(incident.created_at), {
                        addSuffix: true,
                      })}
                    </div>
                  </div>

                  {incident.resolved_at && (
                    <>
                      <div>
                        <div className="flex items-center gap-2 text-text-muted text-sm mb-1">
                          <CheckCircle className="w-4 h-4" />
                          Resolved
                        </div>
                        <div className="text-emerald-400 font-medium">
                          {format(
                            parseISO(incident.resolved_at),
                            "MMM d, HH:mm",
                          )}
                        </div>
                        <div className="text-xs text-text-muted">
                          {formatDistanceToNow(parseISO(incident.resolved_at), {
                            addSuffix: true,
                          })}
                        </div>
                      </div>

                      <div>
                        <div className="flex items-center gap-2 text-text-muted text-sm mb-1">
                          <Activity className="w-4 h-4" />
                          Duration
                        </div>
                        <div className="text-text-primary font-medium">
                          {duration} minutes
                        </div>
                        <div className="text-xs text-text-muted">
                          ~{Math.round(((duration || 0) / 60) * 10) / 10} hours
                        </div>
                      </div>
                    </>
                  )}

                  <div>
                    <div className="flex items-center gap-2 text-text-muted text-sm mb-1">
                      <Server className="w-4 h-4" />
                      Affected
                    </div>
                    <div className="text-text-primary font-medium">
                      {incident.affected_components?.length || 0}{" "}
                      {incident.affected_components?.length === 1
                        ? "service"
                        : "services"}
                    </div>
                    <div className="text-xs text-text-muted truncate">
                      {incident.affected_components?.join(", ") || "None"}
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </motion.div>

          {/* Affected Services */}
          {incident.affected_components &&
            incident.affected_components.length > 0 && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.1 }}
              >
                <Card className="border-border-subtle bg-bg-tertiary/50 backdrop-blur-sm">
                  <CardHeader>
                    <CardTitle className="text-lg flex items-center gap-2">
                      <Server className="w-5 h-5 text-brand-400" />
                      Affected Services
                    </CardTitle>
                    <CardDescription>
                      Services impacted during this incident
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="flex flex-wrap gap-2">
                      {incident.affected_components.map((service, index) => (
                        <motion.div
                          key={service}
                          initial={{ opacity: 0, scale: 0.9 }}
                          animate={{ opacity: 1, scale: 1 }}
                          transition={{ delay: 0.2 + index * 0.05 }}
                          className={cn(
                            "flex items-center gap-2 px-4 py-2 rounded-lg border",
                            "bg-bg-secondary border-border-subtle",
                            isResolved
                              ? "hover:border-emerald-500/30"
                              : "hover:border-amber-500/30",
                            "transition-colors",
                          )}
                        >
                          <StatusDot
                            status={isResolved ? "operational" : "degraded"}
                            size="sm"
                            pulse={!isResolved}
                          />
                          <span className="text-text-primary font-medium capitalize">
                            {service.replace(/-/g, " ")}
                          </span>
                          <span
                            className={cn(
                              "text-xs",
                              isResolved
                                ? "text-emerald-400"
                                : "text-amber-400",
                            )}
                          >
                            {isResolved ? "Operational" : "Affected"}
                          </span>
                        </motion.div>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              </motion.div>
            )}

          {/* Timeline */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
          >
            <Card className="border-border-subtle bg-bg-tertiary/50 backdrop-blur-sm">
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2">
                  <MessageSquare className="w-5 h-5 text-brand-400" />
                  Incident Timeline
                </CardTitle>
                <CardDescription>
                  Chronological updates and status changes
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-0">
                  {incident.updates?.map((update, index) => (
                    <TimelineItem
                      key={update.id}
                      update={update}
                      isLast={index === (incident.updates?.length || 0) - 1}
                    />
                  )) || (
                    <p className="text-text-muted text-center py-8">
                      No updates available for this incident.
                    </p>
                  )}
                </div>
              </CardContent>
            </Card>
          </motion.div>

          {/* Actions */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.4 }}
            className={cn(
              "rounded-2xl p-6 border",
              "bg-gradient-to-br from-brand-500/10 via-purple-500/5 to-transparent",
              "border-brand-500/20",
            )}
          >
            <div className="flex flex-col md:flex-row items-center justify-between gap-4">
              <div>
                <h3 className="text-lg font-semibold text-text-primary mb-1">
                  Stay informed about incidents
                </h3>
                <p className="text-text-secondary text-sm">
                  Subscribe to get notified when we create, update or resolve
                  incidents.
                </p>
              </div>
              <div className="flex gap-3">
                <Button
                  variant="outline"
                  className="border-border-default hover:bg-bg-hover"
                  onClick={() => setShowSubscribe(!showSubscribe)}
                >
                  <Bell className="w-4 h-4 mr-2" />
                  Subscribe
                </Button>
                <Button
                  className="bg-brand-500 hover:bg-brand-600 text-white shadow-lg shadow-brand-500/25"
                  onClick={() => window.open("/api/v1/status/rss", "_blank")}
                >
                  <ExternalLink className="w-4 h-4 mr-2" />
                  RSS Feed
                </Button>
              </div>
            </div>

            <AnimatePresence>
              {showSubscribe && (
                <motion.div
                  initial={{ height: 0, opacity: 0 }}
                  animate={{ height: "auto", opacity: 1 }}
                  exit={{ height: 0, opacity: 0 }}
                  className="mt-4 pt-4 border-t border-border-subtle overflow-hidden"
                >
                  <div className="flex gap-3">
                    <input
                      type="email"
                      placeholder="Enter your email"
                      className={cn(
                        "flex-1 h-10 px-4 rounded-lg border text-sm",
                        "bg-bg-secondary border-border-subtle text-text-primary",
                        "placeholder:text-text-muted",
                        "focus:outline-none focus:border-brand-500/50 focus:ring-2 focus:ring-brand-500/20",
                        "transition-all duration-200",
                      )}
                    />
                    <Button className="bg-brand-500 hover:bg-brand-600 text-white">
                      Subscribe
                    </Button>
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </motion.div>
        </div>
      </main>

      {/* Footer */}
      <footer className="border-t border-border-subtle bg-bg-secondary/50 mt-16">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <div className="flex items-center gap-2 text-sm text-text-muted">
              <Logo size="sm" showText={false} />
              <span>© {new Date().getFullYear()} FunctionFly</span>
            </div>
            <div className="flex items-center gap-6 text-sm text-text-muted">
              <Link
                to="/"
                className="hover:text-text-primary transition-colors"
              >
                Status
              </Link>
              <Link
                to="/history"
                className="hover:text-text-primary transition-colors"
              >
                History
              </Link>
              <a
                href="https://docs.functionfly.com"
                className="hover:text-text-primary transition-colors"
              >
                API
              </a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}
