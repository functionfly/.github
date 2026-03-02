/**
 * Settings Tab Component
 *
 * Displays profile settings for the user's own profile.
 */

import { motion } from "framer-motion";
import { Settings } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { tabContentVariants } from "../../animations";

export function SettingsTab() {
  return (
    <motion.div
      variants={tabContentVariants}
      initial="hidden"
      animate="visible"
      exit="exit"
      className="space-y-6 px-4 md:px-8 pb-8"
    >
      <Card className="border-border-subtle">
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Settings className="w-5 h-5 text-brand-500" />
            Profile Settings
          </CardTitle>
          <CardDescription>
            Manage your profile visibility and preferences
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-text-muted">
            Profile settings would be managed here. This is a placeholder for future implementation.
          </p>
        </CardContent>
      </Card>
    </motion.div>
  );
}
