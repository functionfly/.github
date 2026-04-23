/**
 * Contribution Graph Component
 *
 * GitHub-style heatmap showing user contribution activity over the past year.
 */

import { useMemo } from "react";
import { format } from "date-fns";
import { GitBranch } from "lucide-react";
import { cn, formatNumber } from "@/lib/utils";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { UserProfile } from "@/types";

export interface ContributionGraphProps {
  data: UserProfile["stats"]["contributionGraph"];
}

export function ContributionGraph({ data }: ContributionGraphProps) {
  const levelColors = [
    "bg-border-subtle",
    "bg-brand-500/20",
    "bg-brand-500/40",
    "bg-brand-500/60",
    "bg-brand-500",
  ];

  // Group by weeks
  const weeks = useMemo(() => {
    const result: typeof data[] = [];
    for (let i = 0; i < data.length; i += 7) {
      result.push(data.slice(i, i + 7));
    }
    return result;
  }, [data]);

  const totalContributions = data.reduce((sum, day) => sum + day.count, 0);

  return (
    <Card className="border-border-subtle">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg font-display flex items-center gap-2">
            <GitBranch className="w-5 h-5 text-brand-500" />
            Contribution Activity
          </CardTitle>
          <span className="text-sm text-text-muted">
            {formatNumber(totalContributions)} contributions in the last year
          </span>
        </div>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto">
          <div className="flex gap-1 min-w-max">
            {weeks.map((week, weekIndex) => (
              <div key={weekIndex} className="flex flex-col gap-1">
                {week.map((day, dayIndex) => (
                  <div
                    key={dayIndex}
                    className={cn(
                      "w-3 h-3 rounded-sm transition-all duration-200 hover:ring-2 hover:ring-brand-500/50 cursor-pointer",
                      levelColors[day.level]
                    )}
                    title={`${day.count} contributions on ${format(new Date(day.date), "MMM d, yyyy")}`}
                  />
                ))}
              </div>
            ))}
          </div>
        </div>

        {/* Legend */}
        <div className="flex items-center gap-2 mt-4 text-xs text-text-muted">
          <span>Less</span>
          {levelColors.map((color, i) => (
            <div key={i} className={cn("w-3 h-3 rounded-sm", color)} />
          ))}
          <span>More</span>
        </div>
      </CardContent>
    </Card>
  );
}
