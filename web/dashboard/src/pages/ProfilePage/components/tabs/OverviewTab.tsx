/**
 * Overview Tab Component
 *
 * Displays summary stats, featured functions, and key profile information.
 */

import { FunctionCard } from '@/components/functions/FunctionCard';
import type { UserProfile } from '@/types';
import { motion } from 'framer-motion';
import { Star } from 'lucide-react';
import { Link } from 'react-router-dom';
import { tabContentVariants } from '../../animations';
import { AchievementsSection } from '../AchievementsSection';
import { ContributionActivity } from '../ContributionActivity';
import { SkillsSection } from '../SkillsSection';
import { TrustMetricsSection } from '../TrustMetricsSection';
import { CertificationsSection } from '../CertificationsSection';

export interface OverviewTabProps {
  profile: UserProfile;
}

export function OverviewTab({ profile }: OverviewTabProps) {
  const featuredFunctions = profile.publishedFunctions
    .filter((f) => f.isFeatured || f.metrics.executionCount > 1000)
    .slice(0, 4);

  return (
    <motion.div
      variants={tabContentVariants}
      initial="hidden"
      animate="visible"
      exit="exit"
      className="space-y-6 px-4 md:px-8 pb-8"
    >
      {/* Featured Functions */}
      {featuredFunctions.length > 0 && (
        <section>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold font-display text-text-primary flex items-center gap-2">
              <Star className="w-5 h-5 text-brand-500" />
              Featured Functions
            </h2>
            <Link to="?tab=functions" className="text-sm text-brand-400 hover:text-brand-300">
              View all
            </Link>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {featuredFunctions.map((fn) => (
              <FunctionCard key={fn.id} data={fn} variant="compact" />
            ))}
          </div>
        </section>
      )}

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column - Contribution Activity */}
        <div className="lg:col-span-2">
          <ContributionActivity
            profile={profile}
            maxActivities={6}
            compact
          />
        </div>

        {/* Right Column - Achievements, Skills, Trust */}
        <div className="space-y-6">
          <AchievementsSection achievements={profile.achievements} />
          <CertificationsSection badges={profile.certifications} username={profile.username} />
          <TrustMetricsSection trustScore={profile.stats.trustScore} />
          <SkillsSection skills={profile.skills} />
        </div>
      </div>
    </motion.div>
  );
}
