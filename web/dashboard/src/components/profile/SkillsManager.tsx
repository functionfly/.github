/**
 * SkillsManager Component
 *
 * A comprehensive component for managing user skills with tags/chips display,
 * autocomplete suggestions, and add/remove functionality.
 *
 * @example
 * <SkillsManager
 *   skills={userSkills}
 *   isOwnProfile={true}
 *   onAddSkill={handleAddSkill}
 *   onRemoveSkill={handleRemoveSkill}
 *   maxSkills={20}
 * />
 */

import { useState, useRef, useCallback, useMemo } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  X,
  Plus,
  Sparkles,
  TrendingUp,
  Code2,
  Database,
  Cloud,
  Shield,
  Cpu,
  Globe,
  Smartphone,
  Terminal,
  GitBranch,
  Layers,
  Zap,
  Search,
  Loader2,
  AlertCircle,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type { Skill } from "@/types";
import type { UserSkill } from "@/api/users";

interface SkillsManagerProps {
  skills: UserSkill[];
  isOwnProfile: boolean;
  onAddSkill?: (skill: { name: string; level?: string; category?: string }) => Promise<void>;
  onRemoveSkill?: (skillId: string) => Promise<void>;
  maxSkills?: number;
  isLoading?: boolean;
  popularSkills?: string[];
}

// Popular/ suggested skills for developers
const DEFAULT_POPULAR_SKILLS = [
  "JavaScript",
  "TypeScript",
  "Python",
  "Go",
  "Rust",
  "React",
  "Vue",
  "Node.js",
  "Docker",
  "Kubernetes",
  "AWS",
  "PostgreSQL",
  "MongoDB",
  "GraphQL",
  "REST API",
  "WebAssembly",
  "Serverless",
  "CI/CD",
  "Git",
  "Linux",
];

// Skill categories with icons
const SKILL_CATEGORIES = {
  language: { icon: Code2, color: "text-blue-400", bg: "bg-blue-500/10" },
  framework: { icon: Layers, color: "text-purple-400", bg: "bg-purple-500/10" },
  tool: { icon: Terminal, color: "text-green-400", bg: "bg-green-500/10" },
  platform: { icon: Cloud, color: "text-orange-400", bg: "bg-orange-500/10" },
  concept: { icon: Zap, color: "text-yellow-400", bg: "bg-yellow-500/10" },
  database: { icon: Database, color: "text-cyan-400", bg: "bg-cyan-500/10" },
  security: { icon: Shield, color: "text-red-400", bg: "bg-red-500/10" },
  mobile: { icon: Smartphone, color: "text-pink-400", bg: "bg-pink-500/10" },
  devops: { icon: GitBranch, color: "text-indigo-400", bg: "bg-indigo-500/10" },
  ai: { icon: Cpu, color: "text-emerald-400", bg: "bg-emerald-500/10" },
  web: { icon: Globe, color: "text-sky-400", bg: "bg-sky-500/10" },
};

// Skill level colors
const LEVEL_COLORS = {
  beginner: "bg-slate-500/20 text-slate-400 border-slate-500/30",
  intermediate: "bg-blue-500/20 text-blue-400 border-blue-500/30",
  advanced: "bg-purple-500/20 text-purple-400 border-purple-500/30",
  expert: "bg-amber-500/20 text-amber-400 border-amber-500/30",
};

const LEVEL_LABELS = {
  beginner: "Beginner",
  intermediate: "Intermediate",
  advanced: "Advanced",
  expert: "Expert",
};

export function SkillsManager({
  skills,
  isOwnProfile,
  onAddSkill,
  onRemoveSkill,
  maxSkills = 20,
  isLoading = false,
  popularSkills = DEFAULT_POPULAR_SKILLS,
}: SkillsManagerProps) {
  const [newSkill, setNewSkill] = useState("");
  const [isAdding, setIsAdding] = useState(false);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [selectedLevel, setSelectedLevel] = useState<UserSkill["level"]>("intermediate");
  const inputRef = useRef<HTMLInputElement>(null);

  // Filter out already added skills from suggestions
  const availableSuggestions = useMemo(() => {
    const currentSkillNames = new Set(skills.map((s) => s.name.toLowerCase()));
    return popularSkills.filter((skill) => !currentSkillNames.has(skill.toLowerCase()));
  }, [skills, popularSkills]);

  // Filter suggestions based on input
  const filteredSuggestions = useMemo(() => {
    if (!newSkill.trim()) return availableSuggestions.slice(0, 8);
    return availableSuggestions
      .filter((skill) => skill.toLowerCase().includes(newSkill.toLowerCase()))
      .slice(0, 8);
  }, [newSkill, availableSuggestions]);

  // Get skill category icon
  const getSkillIcon = useCallback((category?: string) => {
    const cat = category?.toLowerCase() as keyof typeof SKILL_CATEGORIES;
    const config = SKILL_CATEGORIES[cat] || SKILL_CATEGORIES.concept;
    return config;
  }, []);

  // Handle add skill
  const handleAddSkill = async () => {
    if (!newSkill.trim() || skills.length >= maxSkills || !onAddSkill) return;

    setIsAdding(true);
    try {
      await onAddSkill({
        name: newSkill.trim(),
        level: selectedLevel,
        category: "concept", // Default category, can be updated by API
      });
      setNewSkill("");
      setShowSuggestions(false);
    } finally {
      setIsAdding(false);
    }
  };

  // Handle remove skill
  const handleRemoveSkill = async (skillId: string) => {
    if (!onRemoveSkill) return;
    await onRemoveSkill(skillId);
  };

  // Handle suggestion click
  const handleSuggestionClick = (suggestion: string) => {
    setNewSkill(suggestion);
    setShowSuggestions(false);
    inputRef.current?.focus();
  };

  // Handle key press
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleAddSkill();
    } else if (e.key === "Escape") {
      setShowSuggestions(false);
    }
  };

  const canAddMore = skills.length < maxSkills;
  const isAtLimit = skills.length >= maxSkills;

  return (
    <div className="space-y-4">
      {/* Header with count */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Sparkles className="w-4 h-4 text-brand-500" />
          <span className="text-sm font-medium text-text-primary">
            Skills ({skills.length}/{maxSkills})
          </span>
        </div>
        {isAtLimit && (
          <span className="text-xs text-error flex items-center gap-1">
            <AlertCircle className="w-3 h-3" />
            Maximum reached
          </span>
        )}
      </div>

      {/* Skills Display */}
      <div className="flex flex-wrap gap-2">
        <AnimatePresence mode="popLayout">
          {skills.map((skill) => {
            const iconConfig = getSkillIcon(skill.category);
            const Icon = iconConfig.icon;

            return (
              <motion.div
                key={skill.id}
                initial={{ opacity: 0, scale: 0.8 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.8 }}
                layout
                className={cn(
                  "group inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full border text-sm font-medium transition-all",
                  "hover:shadow-md",
                  LEVEL_COLORS[skill.level]
                )}
              >
                <Icon className={cn("w-3.5 h-3.5", iconConfig.color)} />
                <span>{skill.name}</span>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <span className="text-xs opacity-60 ml-0.5">
                        {LEVEL_LABELS[skill.level]}
                      </span>
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Proficiency: {LEVEL_LABELS[skill.level]}</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
                {isOwnProfile && onRemoveSkill && (
                  <button
                    onClick={() => handleRemoveSkill(skill.id)}
                    disabled={isLoading}
                    className="ml-1 p-0.5 rounded-full opacity-0 group-hover:opacity-100 hover:bg-black/20 transition-all disabled:opacity-50"
                    aria-label={`Remove ${skill.name}`}
                  >
                    <X className="w-3 h-3" />
                  </button>
                )}
              </motion.div>
            );
          })}
        </AnimatePresence>

        {skills.length === 0 && (
          <p className="text-sm text-text-muted italic">
            No skills added yet
          </p>
        )}
      </div>

      {/* Add Skill Input - Only for own profile */}
      {isOwnProfile && canAddMore && (
        <div className="space-y-3">
          {/* Skill level selector */}
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-xs text-text-secondary">Proficiency:</span>
            {(["beginner", "intermediate", "advanced", "expert"] as const).map((level) => (
              <button
                key={level}
                onClick={() => setSelectedLevel(level)}
                className={cn(
                  "px-2 py-0.5 text-xs rounded-full border transition-all",
                  selectedLevel === level
                    ? LEVEL_COLORS[level]
                    : "border-border-subtle text-text-muted hover:border-border-default"
                )}
              >
                {LEVEL_LABELS[level]}
              </button>
            ))}
          </div>

          {/* Input with suggestions */}
          <div className="relative">
            <div className="flex gap-2">
              <div className="relative flex-1">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
                <Input
                  ref={inputRef}
                  value={newSkill}
                  onChange={(e) => {
                    setNewSkill(e.target.value);
                    setShowSuggestions(true);
                  }}
                  onFocus={() => setShowSuggestions(true)}
                  onKeyDown={handleKeyDown}
                  placeholder="Add a skill..."
                  className="pl-10"
                  disabled={isLoading || isAdding}
                />

                {/* Suggestions dropdown */}
                <AnimatePresence>
                  {showSuggestions && filteredSuggestions.length > 0 && (
                    <motion.div
                      initial={{ opacity: 0, y: -10 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -10 }}
                      className="absolute z-50 w-full mt-1 bg-card border border-border-default rounded-lg shadow-lg overflow-hidden"
                    >
                      <div className="max-h-48 overflow-y-auto py-1">
                        {filteredSuggestions.map((suggestion) => (
                          <button
                            key={suggestion}
                            onClick={() => handleSuggestionClick(suggestion)}
                            className="w-full px-3 py-2 text-left text-sm text-text-primary hover:bg-hover transition-colors flex items-center gap-2"
                          >
                            <Plus className="w-3.5 h-3.5 text-brand-500" />
                            {suggestion}
                          </button>
                        ))}
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>

              <Button
                onClick={handleAddSkill}
                disabled={!newSkill.trim() || isAdding || isLoading}
                size="sm"
                className="gap-1"
              >
                {isAdding ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <Plus className="w-4 h-4" />
                )}
                Add
              </Button>
            </div>

            {/* Click outside to close suggestions */}
            {showSuggestions && (
              <div
                className="fixed inset-0 z-40"
                onClick={() => setShowSuggestions(false)}
              />
            )}
          </div>

          {/* Popular skills quick add */}
          <div className="space-y-2">
            <div className="flex items-center gap-2">
              <TrendingUp className="w-3.5 h-3.5 text-text-muted" />
              <span className="text-xs text-text-secondary">Popular skills:</span>
            </div>
            <div className="flex flex-wrap gap-1.5">
              {availableSuggestions.slice(0, 10).map((skill) => (
                <button
                  key={skill}
                  onClick={() => handleSuggestionClick(skill)}
                  disabled={isLoading}
                  className="px-2 py-1 text-xs rounded-md border border-border-subtle text-text-secondary hover:border-brand-500/50 hover:text-brand-400 transition-colors"
                >
                  + {skill}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default SkillsManager;
