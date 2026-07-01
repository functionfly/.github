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
  const cardStyle = { background: 'var(--panel-raised)', borderColor: 'var(--panel-edge)', borderRadius: 'var(--radius-lg)' };
  const iconOk = { color: 'var(--status-ok)' };
  const iconFoil = { color: 'var(--foil-b)' };

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
          <Card style={cardStyle}>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2" style={{ fontFamily: 'var(--font-display)' }}>
                <User className="w-5 h-5" style={iconOk} />
                About
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="whitespace-pre-wrap" style={{ color: 'var(--text-dim)' }}>
                {profile.bio || "No bio provided yet."}
              </p>
            </CardContent>
          </Card>

          {/* Experience */}
          {profile.experience && profile.experience.length > 0 && (
            <Card style={cardStyle}>
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2" style={{ fontFamily: 'var(--font-display)' }}>
                  <Briefcase className="w-5 h-5" style={iconOk} />
                  Experience
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                {profile.experience.map((exp, index) => (
                  <div key={index} className="flex gap-4">
                    <div className="w-10 h-10 rounded-[var(--radius)] flex items-center justify-center shrink-0" style={{ background: 'rgba(143, 255, 208, 0.06)' }}>
                      <Building2 className="w-5 h-5" style={iconOk} />
                    </div>
                    <div>
                      <h4 className="font-medium" style={{ color: 'var(--text)' }}>{exp.title}</h4>
                      <p className="text-sm" style={{ color: 'var(--text-dim)' }}>{exp.company}</p>
                      <p className="text-xs mt-1" style={{ color: 'var(--text-faint)' }}>
                        {format(new Date(exp.startDate), "MMM yyyy")} -{" "}
                        {exp.current ? "Present" : exp.endDate ? format(new Date(exp.endDate), "MMM yyyy") : ""}
                      </p>
                      {exp.description && (
                        <p className="text-sm mt-2" style={{ color: 'var(--text-faint)' }}>{exp.description}</p>
                      )}
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>
          )}

          {/* Education */}
          {profile.education && profile.education.length > 0 && (
            <Card style={cardStyle}>
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2" style={{ fontFamily: 'var(--font-display)' }}>
                  <GraduationCap className="w-5 h-5" style={iconFoil} />
                  Education
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                {profile.education.map((edu, index) => (
                  <div key={index} className="flex gap-4">
                    <div className="w-10 h-10 rounded-[var(--radius)] flex items-center justify-center shrink-0" style={{ background: 'rgba(217, 196, 255, 0.08)' }}>
                      <GraduationCap className="w-5 h-5" style={iconFoil} />
                    </div>
                    <div>
                      <h4 className="font-medium" style={{ color: 'var(--text)' }}>{edu.institution}</h4>
                      <p className="text-sm" style={{ color: 'var(--text-dim)' }}>{edu.degree} in {edu.field}</p>
                      <p className="text-xs mt-1" style={{ color: 'var(--text-faint)' }}>
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
            <Card style={cardStyle}>
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2" style={{ fontFamily: 'var(--font-display)' }}>
                  <Icon icon="simple-icons:github" className="w-5 h-5" style={iconOk} />
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
                      className="flex items-center justify-between p-3 rounded-[var(--radius)] transition-colors"
                      style={{ background: 'var(--panel)' }}
                    >
                      <span className="font-medium" style={{ color: 'var(--text)' }}>
                        {contrib.project}
                      </span>
                      <span className="text-sm" style={{ color: 'var(--text-faint)' }}>
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
          <Card style={cardStyle}>
            <CardHeader>
              <CardTitle className="text-lg" style={{ fontFamily: 'var(--font-display)' }}>Contact</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {profile.socialLinks.github && (
                <a
                  href={profile.socialLinks.github}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center gap-3 transition-colors"
                  style={{ color: 'var(--text-dim)' }}
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
                  className="flex items-center gap-3 transition-colors"
                  style={{ color: 'var(--text-dim)' }}
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
                  className="flex items-center gap-3 transition-colors"
                  style={{ color: 'var(--text-dim)' }}
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
                  className="flex items-center gap-3 transition-colors"
                  style={{ color: 'var(--text-dim)' }}
                >
                  <Globe className="w-5 h-5" />
                  <span className="text-sm">Website</span>
                </a>
              )}
            </CardContent>
          </Card>

          {/* Languages */}
          {profile.languages && profile.languages.length > 0 && (
            <Card style={cardStyle}>
              <CardHeader>
                <CardTitle className="text-lg" style={{ fontFamily: 'var(--font-display)' }}>Languages</CardTitle>
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

          {/* Skills */}
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
