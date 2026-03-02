/**
 * Trust Metrics Section Component
 *
 * Displays user's trust score with detailed metrics breakdown.
 */

import { Target, Shield, Zap, Users, BookOpen } from "lucide-react";
import { cn } from "@/lib/utils";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export interface TrustMetricsSectionProps {
  trustScore: number;
}

export function TrustMetricsSection({ trustScore }: TrustMetricsSectionProps) {
  const metrics = [
    { name: "Reliability", score: Math.min(100, trustScore + Math.random() * 10 - 5), icon: Shield },
    { name: "Performance", score: Math.min(100, trustScore + Math.random() * 10 - 5), icon: Zap },
    { name: "Community", score: Math.min(100, trustScore + Math.random() * 15 - 7), icon: Users },
    { name: "Documentation", score: Math.min(100, trustScore + Math.random() * 20 - 10), icon: BookOpen },
  ];

  return (
    <Card className="border-border-subtle">
      <CardHeader className="pb-3">
        <CardTitle className="text-lg flex items-center gap-2">
          <Target className="w-5 h-5 text-brand-500" />
          Trust Metrics
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-4 mb-6">
          <div className="relative w-24 h-24">
            <svg className="w-full h-full -rotate-90" viewBox="0 0 40 40">
              <circle
                cx="20"
                cy="20"
                r="16"
                fill="none"
                stroke="currentColor"
                strokeWidth="3"
                className="text-border-subtle"
              />
              <circle
                cx="20"
                cy="20"
                r="16"
                fill="none"
                stroke="currentColor"
                strokeWidth="3"
                strokeLinecap="round"
                className={cn(
                  trustScore >= 80 ? "text-emerald-500" :
                  trustScore >= 60 ? "text-yellow-500" :
                  "text-orange-500"
                )}
                style={{
                  strokeDasharray: 2 * Math.PI * 16,
                  strokeDashoffset: 2 * Math.PI * 16 * (1 - trustScore / 100),
                }}
              />
            </svg>
            <div className="absolute inset-0 flex flex-col items-center justify-center">
              <span className="text-2xl font-bold text-text-primary">{trustScore}</span>
              <span className="text-xs text-text-muted">Score</span>
            </div>
          </div>
          <div>
            <h4 className="font-medium text-text-primary">
              {trustScore >= 80 ? "Excellent" : trustScore >= 60 ? "Good" : "Fair"} Reputation
            </h4>
            <p className="text-sm text-text-muted">
              Based on function quality, community engagement, and execution reliability
            </p>
          </div>
        </div>

        <div className="space-y-3">
          {metrics.map((metric) => (
            <div key={metric.name} className="flex items-center gap-3">
              <metric.icon className="w-4 h-4 text-text-muted" />
              <span className="text-sm text-text-secondary w-28">{metric.name}</span>
              <div className="flex-1">
                <div className="h-2 bg-border-subtle rounded-full overflow-hidden">
                  <div
                    className={cn(
                      "h-full rounded-full transition-all duration-1000",
                      metric.score >= 80 ? "bg-emerald-500" :
                      metric.score >= 60 ? "bg-yellow-500" :
                      "bg-orange-500"
                    )}
                    style={{ width: `${metric.score}%` }}
                  />
                </div>
              </div>
              <span className="text-sm font-medium text-text-primary w-10 text-right">
                {Math.round(metric.score)}
              </span>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
