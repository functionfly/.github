/**
 * About Tab Component
 *
 * Displays user's bio, experience, education, skills, and contact info.
 */

import { format } from "date-fns";
import { motion } from "framer-motion";
import {
  User,
  Building2,
  Briefcase,
  GraduationCap,
  Globe,
} from "lucide-react";
import { Icon } from '@iconify/react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { SkillsManager } from "@/components/profile/SkillsManager";
import { tabContentVariants } from "../../animations";
import { SkillsSection } from "../SkillsSection";
import type { UserProfile } from "@/types";
import type { UserSkill } from "@/api/users";

export interface AboutTabProps {
  profile: UserProfile;
  isOwnProfile: boolean;
  userSkills?: UserSkill[];
  onAddSkill?: (skill: { name: string; level?: string; category?: string }) => Promise<void>;
  onRemoveSkill?: (skillId: string) => Promise<void>;
  isSkillsLoading?: boolean;
}

export function AboutTab({ profile, isOwnProfile, userSkills, onAddSkill, onRemoveSkill, isSkillsLoading }: AboutTabProps) {
  return (
    <motion.div
      variants={tabContentVariants}
      initial="hidden"
      animate="visible"
      exit="exit"
      className="space-y-6 px-4 md:px-8 pb-8"
    >
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Bio */}
          <Card className="border-border-subtle">
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <User className="w-5 h-5 text-brand-500" />
                About
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-text-secondary whitespace-pre-wrap">
                {profile.bio || "No bio provided yet."}
              </p>
            </CardContent>
          </Card>

          {/* Experience */}
          {profile.experience && profile.experience.length > 0 && (
            <Card className="border-border-subtle">
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2">
                  <Briefcase className="w-5 h-5 text-brand-500" />
                  Experience
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                {profile.experience.map((exp, index) => (
                  <div key={index} className="flex gap-4">
                    <div className="w-10 h-10 rounded-lg bg-brand-500/10 flex items-center justify-center shrink-0">
                      <Building2 className="w-5 h-5 text-brand-500" />
                    </div>
                    <div>
                      <h4 className="font-medium text-text-primary">{exp.title}</h4>
                      <p className="text-sm text-text-secondary">{exp.company}</p>
                      <p className="text-xs text-text-muted mt-1">
                        {format(new Date(exp.startDate), "MMM yyyy")} -{" "}
                        {exp.current ? "Present" : exp.endDate ? format(new Date(exp.endDate), "MMM yyyy") : ""}
                      </p>
                      {exp.description && (
                        <p className="text-sm text-text-muted mt-2">{exp.description}</p>
                      )}
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          )}

          {/* Education */}
          {profile.education && profile.education.length > 0 && (
            <Card className="border-border-subtle">
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2">
                  <GraduationCap className="w-5 h-5 text-brand-500" />
                  Education
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                {profile.education.map((edu, index) => (
                  <div key={index} className="flex gap-4">
                    <div className="w-10 h-10 rounded-lg bg-purple-500/10 flex items-center justify-center shrink-0">
                      <GraduationCap className="w-5 h-5 text-purple-500" />
                    </div>
                    <div>
                      <h4 className="font-medium text-text-primary">{edu.institution}</h4>
                      <p className="text-sm text-text-secondary">{edu.degree} in {edu.field}</p>
                      <p className="text-xs text-text-muted mt-1">
                        {format(new Date(edu.startDate), "yyyy")} -{" "}
                        {edu.endDate ? format(new Date(edu.endDate), "yyyy") : "Present"}
                      </p>
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          )}

          {/* Open Source */}
          {profile.openSourceContributions && profile.openSourceContributions.length > 0 && (
            <Card className="border-border-subtle">
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2">
                  <Icon icon="simple-icons:github" className="w-5 h-5 text-brand-500" />
                  Open Source Contributions
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {profile.openSourceContributions.map((contrib, index) => (
                    <a
                      key={index}
                      href={contrib.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center justify-between p-3 rounded-lg bg-secondary/50 hover:bg-secondary transition-colors group"
                    >
                      <span className="font-medium text-text-primary group-hover:text-brand-400 transition-colors">
                        {contrib.project}
                      </span>
                      <span className="text-sm text-text-muted">
                        {contrib.contributions} contributions
                      </span>
                    </a>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Contact Info */}
          <Card className="border-border-subtle">
            <CardHeader>
              <CardTitle className="text-lg">Contact</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {profile.socialLinks.github && (
                <a
                  href={profile.socialLinks.github}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-3 text-text-secondary hover:text-text-primary transition-colors"
                >
                  <Icon icon="simple-icons:github" className="w-5 h-5" />
                  <span className="text-sm">GitHub</span>
                </a>
              )}
              {profile.socialLinks.twitter && (
                <a
                  href={profile.socialLinks.twitter}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-3 text-text-secondary hover:text-text-primary transition-colors"
                >
                  <Icon icon="simple-icons:x" className="w-5 h-5" />
                  <span className="text-sm">Twitter</span>
                </a>
              )}
              {profile.socialLinks.linkedin && (
                <a
                  href={profile.socialLinks.linkedin}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-3 text-text-secondary hover:text-text-primary transition-colors"
                >
                  <Icon icon="simple-icons:linkedin" className="w-5 h-5" />
                  <span className="text-sm">LinkedIn</span>
                </a>
              )}
              {profile.website && (
                <a
                  href={profile.website}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-3 text-text-secondary hover:text-text-primary transition-colors"
                >
                  <Globe className="w-5 h-5" />
                  <span className="text-sm">Website</span>
                </a>
              )}
            </CardContent>
          </Card>

          {/* Languages */}
          {profile.languages && profile.languages.length > 0 && (
            <Card className="border-border-subtle">
              <CardHeader>
                <CardTitle className="text-lg">Languages</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex flex-wrap gap-2">
                  {profile.languages.map((lang) => (
                    <Badge key={lang} variant="secondary">
                      {lang}
                    </Badge>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Skills - Show SkillsManager for own profile, SkillsSection for others */}
          {isOwnProfile ? (
            <SkillsManager
              skills={userSkills || []}
              isOwnProfile={true}
              onAddSkill={onAddSkill}
              onRemoveSkill={onRemoveSkill}
              isLoading={isSkillsLoading}
            />
          ) : (
            <SkillsSection skills={profile.skills.slice(0, 8)} />
          )}
        </div>
      </div>
    </motion.div>
  );
}
