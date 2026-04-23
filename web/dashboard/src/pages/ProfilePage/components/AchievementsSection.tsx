/**
 * Achievements Section Component
 *
 * Displays user's earned achievements/badges.
 */

import { Award } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { AchievementBadge } from "@/components/profile/AchievementBadge";
import type { Achievement as AchievementType } from "@/types";

export interface AchievementsSectionProps {
  achievements: AchievementType[];
}

export function AchievementsSection({ achievements }: AchievementsSectionProps) {
  return (
    <Card className="border-border-subtle">
      <CardHeader className="pb-3">
        <CardTitle className="text-lg font-display flex items-center gap-2">
          <Award className="w-5 h-5 text-brand-500" />
          Achievements
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex flex-wrap gap-3 justify-center">
          {achievements.slice(0, 8).map((achievement) => (
            <AchievementBadge
              key={achievement.id}
              achievement={achievement}
              size="md"
              showProgress
              showRarity
            />
          ))}
        </div>

        {achievements.length > 8 && (
          <Button variant="ghost" className="w-full mt-4 text-sm text-brand-400">
            View all {achievements.length} achievements
          </Button>
        )}
      </CardContent>
    </Card>
  );
}
