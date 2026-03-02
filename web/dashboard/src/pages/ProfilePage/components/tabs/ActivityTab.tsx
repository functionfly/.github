/**
 * Activity Tab Component
 *
 * Displays user's activity feed with filtering options.
 */

import { useState } from "react";
import { motion } from "framer-motion";
import { Activity, Filter } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ActivityTimeline } from "../ActivityTimeline";
import { tabContentVariants } from "../../animations";
import type { UserProfile, ActivityType } from "@/types";

export interface ActivityTabProps {
  profile: UserProfile;
}

export function ActivityTab({ profile }: ActivityTabProps) {
  const [activityFilter, setActivityFilter] = useState<ActivityType | "all">("all");

  const filterOptions: { value: ActivityType | "all"; label: string }[] = [
    { value: "all", label: "All Activity" },
    { value: "joined", label: "Joined" },
    { value: "function_published", label: "Functions" },
    { value: "achievement_earned", label: "Achievements" },
    { value: "review_received", label: "Reviews" },
    { value: "milestone_reached", label: "Milestones" },
  ];

  return (
    <motion.div
      variants={tabContentVariants}
      initial="hidden"
      animate="visible"
      exit="exit"
      className="space-y-6 px-4 md:px-8 pb-8"
    >
      <Card className="border-border-subtle">
        <CardHeader className="pb-3">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <CardTitle className="text-lg flex items-center gap-2">
              <Activity className="w-5 h-5 text-brand-500" />
              Activity Timeline
            </CardTitle>
            <Select value={activityFilter} onValueChange={(v) => setActivityFilter(v as ActivityType | "all")}>
              <SelectTrigger className="w-full sm:w-40">
                <Filter className="w-4 h-4 mr-2" />
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {filterOptions.map(opt => (
                  <SelectItem key={opt.value} value={opt.value}>{opt.label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </CardHeader>
        <CardContent>
          <ActivityTimeline activities={profile.recentActivity} filter={activityFilter} />
        </CardContent>
      </Card>
    </motion.div>
  );
}
