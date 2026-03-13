import { useQuery } from "@tanstack/react-query";
import { User } from "lucide-react";
import { cn } from "@/lib/utils";

export interface CollaborationInsightPanelProps {
  userId: string;
  className?: string;
}

async function fetchCollaborationProfile(userId: string): Promise<{
  reputation?: Record<string, number>;
  shared_threads?: number;
  functions_overlap?: string[];
}> {
  const base = import.meta.env.VITE_API_URL || "";
  const url = `${base}/v1/conversations/collaboration-profile/${userId}`;
  const res = await fetch(url, { credentials: "include" });
  if (!res.ok) return {};
  return res.json();
}

export function CollaborationInsightPanel({ userId, className }: CollaborationInsightPanelProps) {
  const { data } = useQuery({
    queryKey: ["collaboration-profile", userId],
    queryFn: () => fetchCollaborationProfile(userId),
    enabled: Boolean(userId),
  });

  if (!data) return null;

  return (
    <div className={cn("rounded-lg border border-border bg-card p-3 text-sm space-y-2", className)}>
      <div className="flex items-center gap-2 font-medium">
        <User className="h-4 w-4" />
        Profile
      </div>
      {data.reputation && Object.keys(data.reputation).length > 0 && (
        <div>
          <span className="text-muted-foreground text-xs">Reputation</span>
          <div className="flex flex-wrap gap-1 mt-0.5">
            {Object.entries(data.reputation).map(([k, v]) => (
              <span key={k} className="text-xs">
                {k}: {v}
              </span>
            ))}
          </div>
        </div>
      )}
      {data.shared_threads != null && data.shared_threads > 0 && (
        <p className="text-xs text-muted-foreground">{data.shared_threads} shared threads</p>
      )}
    </div>
  );
}
