import { useCallback, useRef, useState } from "react";
import { motion } from "framer-motion";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { StatusBadge } from "@/components/StatusBadge";
import { UptimeMiniBar } from "@/components/UptimeBar";
import { cn } from "@/lib/utils";
import {
  Activity,
  Box,
  Brain,
  Cloud,
  Container,
  CreditCard,
  Database,
  Globe,
  HardDrive,
  Mail,
  Server,
  Shield,
  Sparkles,
  Zap,
} from "lucide-react";
import type { ServiceCardProps, ProviderCardProps } from "./types";

const statusColors = {
  operational: "text-emerald-400 border-emerald-500/30 bg-emerald-500/10",
  degraded: "text-amber-400 border-amber-500/30 bg-amber-500/10",
  maintenance: "text-purple-400 border-purple-500/30 bg-purple-500/10",
  major_outage: "text-red-400 border-red-500/30 bg-red-500/10",
  partial_outage: "text-orange-400 border-orange-500/30 bg-orange-500/10",
};

function getComponentIcon(type: string) {
  switch (type) {
    case "api":
      return <Globe className="w-4 h-4" />;
    case "database":
      return <Database className="w-4 h-4" />;
    case "cache":
      return <Zap className="w-4 h-4" />;
    case "ai":
      return <Brain className="w-4 h-4" />;
    case "email":
      return <Mail className="w-4 h-4" />;
    case "billing":
      return <CreditCard className="w-4 h-4" />;
    case "storage":
      return <HardDrive className="w-4 h-4" />;
    case "cdn":
      return <Cloud className="w-4 h-4" />;
    case "monitoring":
      return <Activity className="w-4 h-4" />;
    case "runtime":
      return <Container className="w-4 h-4" />;
    case "worker":
      return <Box className="w-4 h-4" />;
    case "backup":
      return <HardDrive className="w-4 h-4" />;
    case "infrastructure":
      return <Server className="w-4 h-4" />;
    case "security":
      return <Shield className="w-4 h-4" />;
    case "service":
      return <Sparkles className="w-4 h-4" />;
    default:
      return <Server className="w-4 h-4" />;
  }
}

function getProviderIcon(type: string) {
  switch (type) {
    case "cloud":
      return <Cloud className="w-4 h-4" />;
    case "cdn":
      return <Globe className="w-4 h-4" />;
    case "database":
      return <Database className="w-4 h-4" />;
    case "storage":
      return <HardDrive className="w-4 h-4" />;
    case "compute":
      return <Server className="w-4 h-4" />;
    case "ai":
      return <Brain className="w-4 h-4" />;
    case "security":
      return <Shield className="w-4 h-4" />;
    case "edge":
      return <Zap className="w-4 h-4" />;
    default:
      return <Server className="w-4 h-4" />;
  }
}

function getHealthColor(score: number) {
  if (score >= 98) return "text-emerald-400";
  if (score >= 95) return "text-amber-400";
  return "text-red-400";
}

export function ServiceCard({ component, index }: ServiceCardProps) {
  const cardRef = useRef<HTMLDivElement>(null);
  const [mousePosition, setMousePosition] = useState({ x: 0, y: 0 });

  const handleMouseMove = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
    if (!cardRef.current) return;
    const rect = cardRef.current.getBoundingClientRect();
    const x = ((e.clientX - rect.left) / rect.width) * 100;
    const y = ((e.clientY - rect.top) / rect.height) * 100;
    setMousePosition({ x, y });
  }, []);

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.05, duration: 0.4 }}
      whileHover={{ y: -4, scale: 1.02 }}
      className="group relative"
      onMouseMove={handleMouseMove}
    >
      <Card
        ref={cardRef}
        className={cn(
          "relative overflow-hidden transition-all duration-300",
          "border border-border-subtle bg-bg-tertiary/50 backdrop-blur-sm",
          "hover:border-border-default hover:bg-bg-elevated/80",
          "hover:shadow-lg hover:shadow-brand-500/10",
        )}
        style={
          {
            "--mouse-x": `${mousePosition.x}%`,
            "--mouse-y": `${mousePosition.y}%`,
          } as React.CSSProperties
        }
      >
        <div
          className={cn(
            "absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none",
          )}
          style={{
            background: `radial-gradient(circle at ${mousePosition.x}% ${mousePosition.y}%, rgba(99, 102, 241, 0.15), transparent 50%)`,
          }}
        />

        <CardHeader className="pb-3">
          <div className="flex items-start justify-between gap-4">
            <div className="flex items-center gap-3 min-w-0">
              <div
                className={cn(
                  "p-2 rounded-lg border shrink-0",
                  statusColors[component.status as keyof typeof statusColors],
                )}
              >
                {getComponentIcon(component.type)}
              </div>
              <div className="min-w-0">
                <CardTitle className="text-base font-semibold text-text-primary truncate">
                  {component.name}
                </CardTitle>
                <CardDescription className="text-xs text-text-muted mt-0.5">
                  {component.type}
                </CardDescription>
              </div>
            </div>
            <div className="shrink-0 pt-0.5">
              <StatusBadge status={component.status} />
            </div>
          </div>
        </CardHeader>

        <CardContent className="pb-4">
          <p className="text-sm text-text-secondary mb-4">
            {component.description || `${component.name} service`}
          </p>

          <div className="grid grid-cols-2 gap-3">
            <div className="flex items-center gap-2 text-xs">
              <Activity className="w-3.5 h-3.5 text-text-muted" />
              <span className="text-text-secondary">
                {component.response_time_ms}ms
              </span>
            </div>
            <div className="flex items-center gap-2 text-xs">
              <Zap className="w-3.5 h-3.5 text-emerald-400" />
              <span className="text-emerald-400 font-medium">
                {component.uptime_30d != null ? `${component.uptime_30d.toFixed(2)}%` : 'N/A'}
              </span>
            </div>
          </div>

          <div className="mt-4">
            <UptimeMiniBar
              days={30}
              uptime={component.uptime_30d ?? undefined}
              className="opacity-60 group-hover:opacity-100 transition-opacity"
            />
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}

export function ProviderCard({ provider, index }: ProviderCardProps) {
  const cardRef = useRef<HTMLDivElement>(null);
  const [mousePosition, setMousePosition] = useState({ x: 0, y: 0 });

  const handleMouseMove = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
    if (!cardRef.current) return;
    const rect = cardRef.current.getBoundingClientRect();
    const x = ((e.clientX - rect.left) / rect.width) * 100;
    const y = ((e.clientY - rect.top) / rect.height) * 100;
    setMousePosition({ x, y });
  }, []);

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.05, duration: 0.4 }}
      whileHover={{ y: -4, scale: 1.02 }}
      className="group relative"
      onMouseMove={handleMouseMove}
    >
      <Card
        ref={cardRef}
        className={cn(
          "relative overflow-hidden transition-all duration-300",
          "border border-border-subtle bg-bg-tertiary/50 backdrop-blur-sm",
          "hover:border-border-default hover:bg-bg-elevated/80",
          "hover:shadow-lg hover:shadow-brand-500/10",
        )}
        style={
          {
            "--mouse-x": `${mousePosition.x}%`,
            "--mouse-y": `${mousePosition.y}%`,
          } as React.CSSProperties
        }
      >
        <div
          className={cn(
            "absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none",
          )}
          style={{
            background: `radial-gradient(circle at ${mousePosition.x}% ${mousePosition.y}%, rgba(99, 102, 241, 0.15), transparent 50%)`,
          }}
        />

        <CardHeader className="pb-3">
          <div className="flex items-start justify-between gap-4">
            <div className="flex items-center gap-3 min-w-0">
              <div
                className={cn(
                  "p-2 rounded-lg border shrink-0",
                  statusColors[provider.status as keyof typeof statusColors],
                )}
              >
                {getProviderIcon(provider.type)}
              </div>
              <div className="min-w-0">
                <CardTitle className="text-base font-semibold text-text-primary truncate">
                  {provider.name}
                </CardTitle>
                <CardDescription className="text-xs text-text-muted mt-0.5 flex items-center gap-1">
                  <span className="capitalize">{provider.type}</span>
                  <span className="text-border-default">•</span>
                  <span>{provider.region}</span>
                </CardDescription>
              </div>
            </div>
            <div className="shrink-0 pt-0.5">
              <StatusBadge status={provider.status} />
            </div>
          </div>
        </CardHeader>

        <CardContent className="pb-4">
          <p className="text-sm text-text-secondary mb-4">
            {provider.description}
          </p>

          <div className="grid grid-cols-2 gap-3">
            <div className="flex items-center gap-2 text-xs">
              <Zap className="w-3.5 h-3.5 text-text-muted" />
              <span className="text-text-secondary">
                {provider.latency}ms latency
              </span>
            </div>
            <div className="flex items-center gap-2 text-xs">
              <Activity className="w-3.5 h-3.5 text-text-muted" />
              <span className={cn("font-medium", getHealthColor(provider.healthScore))}>
                {provider.healthScore}% health
              </span>
            </div>
          </div>

          <div className="mt-4">
            <div className="h-1.5 w-full rounded-full bg-bg-secondary overflow-hidden">
              <motion.div
                className={cn(
                  "h-full rounded-full",
                  provider.healthScore >= 98
                    ? "bg-emerald-500"
                    : provider.healthScore >= 95
                      ? "bg-amber-500"
                      : "bg-red-500",
                )}
                initial={{ width: 0 }}
                animate={{ width: `${provider.healthScore}%` }}
                transition={{ duration: 1, delay: 0.5 + index * 0.1, ease: "easeOut" }}
              />
            </div>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}
