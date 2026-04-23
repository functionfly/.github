/**
 * Skills Section Component
 *
 * Displays user's skills and technologies grouped by category.
 */

import { Code2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import type { Skill } from "@/types";

export interface SkillsSectionProps {
  skills: Skill[];
}

export function SkillsSection({ skills }: SkillsSectionProps) {
  const levelColors = {
    beginner: "bg-gray-500/20 text-gray-400",
    intermediate: "bg-blue-500/20 text-blue-400",
    advanced: "bg-purple-500/20 text-purple-400",
    expert: "bg-brand-500/20 text-brand-400",
  };

  const categories = {
    language: "Languages",
    framework: "Frameworks",
    tool: "Tools",
    platform: "Platforms",
    concept: "Concepts",
  };

  const groupedSkills = skills.reduce((acc, skill) => {
    if (!acc[skill.category]) acc[skill.category] = [];
    acc[skill.category].push(skill);
    return acc;
  }, {} as Record<string, Skill[]>);

  return (
    <Card className="border-border-subtle">
      <CardHeader className="pb-3">
        <CardTitle className="text-lg font-display flex items-center gap-2">
          <Code2 className="w-5 h-5 text-brand-500" />
          Skills & Technologies
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {Object.entries(groupedSkills).map(([category, categorySkills]) => (
          <div key={category}>
            <h4 className="text-sm font-medium text-text-muted mb-2">
              {categories[category as keyof typeof categories]}
            </h4>
            <div className="flex flex-wrap gap-2">
              {categorySkills.map((skill) => (
                <Badge
                  key={skill.name}
                  variant="secondary"
                  className={cn(
                    "px-2 py-1 text-xs cursor-default transition-all hover:scale-105",
                    levelColors[skill.level]
                  )}
                  title={`${skill.level}${skill.endorsements ? ` · ${skill.endorsements} endorsements` : ""}`}
                >
                  {skill.name}
                </Badge>
              ))}
            </div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
