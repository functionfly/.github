import { formatDistanceToNow } from "date-fns";
import { CheckCircle, XCircle, Clock, AlertTriangle } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { ProviderIcon } from "./ProviderIcon";

interface ProviderStatusData {
  provider: string;
  status: "connected" | "disconnected" | "connecting" | "error";
  lastChecked: Date | string;
  latency?: number;
  uptime?: number;
  errorMessage?: string;
  functionsCount?: number;
  deploymentsCount?: number;
}

interface ProviderStatusProps {
  providers: ProviderStatusData[];
  className?: string;
}

const statusConfig = {
  connected: {
    icon: CheckCircle,
    color: "text-green-400",
    bgColor: "bg-green-400/10",
    borderColor: "border-green-400/20",
    label: "Connected",
  },
  disconnected: {
    icon: XCircle,
    color: "text-red-400",
    bgColor: "bg-red-400/10",
    borderColor: "border-red-400/20",
    label: "Disconnected",
  },
  connecting: {
    icon: Clock,
    color: "text-yellow-400",
    bgColor: "bg-yellow-400/10",
    borderColor: "border-yellow-400/20",
    label: "Connecting",
  },
  error: {
    icon: AlertTriangle,
    color: "text-red-400",
    bgColor: "bg-red-400/10",
    borderColor: "border-red-400/20",
    label: "Error",
  },
};

export function ProviderStatus({ providers, className }: ProviderStatusProps) {
  return (
    <div className={cn("grid gap-4 md:grid-cols-2 lg:grid-cols-3", className)}>
      {providers.map((providerData) => {
        const status = statusConfig[providerData.status];
        const StatusIcon = status.icon;

        return (
          <Card
            key={providerData.provider}
            className={cn(
              "transition-all duration-200 hover:shadow-lg",
              status.bgColor,
              status.borderColor
            )}
          >
            <CardContent className="p-4">
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-3">
                  <ProviderIcon provider={providerData.provider} size="md" />
                  <div>
                    <h3 className="font-medium text-white capitalize">
                      {providerData.provider}
                    </h3>
                    <div className="flex items-center gap-2 mt-1">
                      <StatusIcon className={cn("w-4 h-4", status.color)} />
                      <Badge
                        variant="secondary"
                        className={cn("text-xs", status.color, status.bgColor)}
                      >
                        {status.label}
                      </Badge>
                    </div>
                  </div>
                </div>
              </div>

              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span className="text-text-secondary">Last checked</span>
                  <span className="text-white">
                    {formatDistanceToNow(new Date(providerData.lastChecked), { addSuffix: true })}
                  </span>
                </div>

                {providerData.latency !== undefined && (
                  <div className="flex justify-between text-sm">
                    <span className="text-text-secondary">Latency</span>
                    <span className="text-white">{providerData.latency}ms</span>
                  </div>
                )}

                {providerData.uptime !== undefined && (
                  <div className="flex justify-between text-sm">
                    <span className="text-text-secondary">Uptime</span>
                    <span className="text-white">{providerData.uptime.toFixed(1)}%</span>
                  </div>
                )}

                {providerData.functionsCount !== undefined && (
                  <div className="flex justify-between text-sm">
                    <span className="text-text-secondary">Functions</span>
                    <span className="text-white">{providerData.functionsCount}</span>
                  </div>
                )}

                {providerData.deploymentsCount !== undefined && (
                  <div className="flex justify-between text-sm">
                    <span className="text-text-secondary">Deployments</span>
                    <span className="text-white">{providerData.deploymentsCount}</span>
                  </div>
                )}

                {providerData.errorMessage && providerData.status === "error" && (
                  <div className="mt-3 p-2 bg-red-400/10 border border-red-400/20 rounded-md">
                    <p className="text-xs text-red-400">{providerData.errorMessage}</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}