import { useQuery } from "@tanstack/react-query";
import { TrendingUp } from "lucide-react";
import { cn } from "@/lib/utils";

export interface ExecutionTimelineOverlayProps {
  author: string;
  name: string;
  metric?: "latency" | "errors" | "trust";
  className?: string;
}

async function fetchTimeline(
  author: string,
  name: string,
  metric: string
): Promise<{ insight?: string; buckets?: Array<{ bucket: string; value: number }> }> {
  const base = import.meta.env.VITE_API_URL || "";
  const url = `${base}/v1/registry/${encodeURIComponent(author)}/${encodeURIComponent(name)}/executions/timeline?metric=${metric}`;
  const res = await fetch(url, { credentials: "include" });
  if (!res.ok) return {};
  const data = await res.json();
  return { insight: data.insight, buckets: data.buckets };
}

export function ExecutionTimelineOverlay({
  author,
  name,
  metric = "latency",
  className,
}: ExecutionTimelineOverlayProps) {
  const { data } = useQuery({
    queryKey: ["execution-timeline", author, name, metric],
    queryFn: () => fetchTimeline(author, name, metric),
    enabled: Boolean(author && name),
  });

  if (!data?.buckets?.length && !data?.insight) return null;

  const latest = data.buckets?.length ? data.buckets[data.buckets.length - 1] : null;
  const prev = data.buckets?.length && data.buckets.length >= 2 ? data.buckets[data.buckets.length - 2] : null;
  const insight =
    data.insight ||
    (latest && prev && latest.value !== prev.value
      ? `${metric}: ${latest.value > prev.value ? "+" : ""}${((latest.value - prev.value) / prev.value) * 100}% vs previous`
      : null);

  return (
    <div
      className={cn(
        "flex items-center gap-2 rounded-md border border-border/60 bg-muted/30 px-2 py-1.5 text-xs",
        className
      )}
    >
      <TrendingUp className="h-3.5 w-3.5 text-muted-foreground" />
      <span className="text-muted-foreground">
        fx://{author}/{name}
      </span>
      {insight && <span className="font-medium">{insight}</span>}
    </div>
  );
}
