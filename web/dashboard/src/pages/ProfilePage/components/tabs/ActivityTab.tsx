/**
 * Activity Tab Component
 *
 * Displays user's full contribution activity with filtering and view options.
 */

import type { UserProfile } from '@/types';
import { motion } from 'framer-motion';
import { tabContentVariants } from '../../animations';
import { ContributionActivity } from '../ContributionActivity';

export interface ActivityTabProps {
  profile: UserProfile;
}

export function ActivityTab({ profile }: ActivityTabProps) {
  return (
    <motion.div
      variants={tabContentVariants}
      initial="hidden"
      animate="visible"
      exit="exit"
      className="space-y-6 px-4 md:px-8 pb-8"
    >
      <ContributionActivity
        profile={profile}
        showFilter
      />
    </motion.div>
  );
}
