/**
 * UserNotFoundView – Empty state when a profile username doesn't exist.
 * Used on public user profile and dashboard profile pages.
 */

import { Link } from "react-router-dom";
import { motion } from "framer-motion";
import { UserX, Zap, Home, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { SpotlightCard } from "@/components/ui/SpotlightCard";
import { cn } from "@/lib/utils";

export interface UserNotFoundViewProps {
  /** Username that was not found (e.g. "micro" for @micro) */
  username?: string;
  /** Whether the error was a 404 (user missing) vs generic error */
  is404?: boolean;
  /** Optional class for the container */
  className?: string;
  /** Use compact layout (e.g. inside dashboard) */
  compact?: boolean;
}

const containerVariants = {
  hidden: { opacity: 0, y: 24 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.35 },
  },
};

const iconVariants = {
  hidden: { scale: 0.8, opacity: 0 },
  visible: {
    scale: 1,
    opacity: 1,
    transition: { delay: 0.1, duration: 0.4 },
  },
};

export function UserNotFoundView({
  username,
  is404 = true,
  className,
  compact = false,
}: UserNotFoundViewProps) {
  const message = is404 && username
    ? `No user with username "@${username}" exists.`
    : "Failed to load this profile. Please try again.";

  const hint = is404 && username
    ? "Check the spelling or explore the function registry."
    : "You can try again later or browse functions.";

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="visible"
      className={cn("user-not-found-view flex flex-col items-center justify-center text-center", className)}
    >
      <SpotlightCard
        spotlightColor="rgba(99, 102, 241, 0.12)"
        spotlightSize={280}
        hoverOnly
        className={cn(
          "user-not-found-view__card w-full max-w-md",
          compact ? "p-6" : "p-8 sm:p-10"
        )}
      >
        <div className="space-y-6">
          {/* Icon */}
          <motion.div
            variants={iconVariants}
            initial="hidden"
            animate="visible"
            className="flex justify-center"
          >
            <div
              className={cn(
                "user-not-found-view__icon rounded-full flex items-center justify-center",
                compact ? "w-14 h-14" : "w-20 h-20"
              )}
            >
              <UserX className={compact ? "w-7 h-7" : "w-10 h-10"} strokeWidth={1.75} />
            </div>
          </motion.div>

          {/* Copy */}
          <div className="space-y-2">
            <h1
              className={cn(
                "user-not-found-view__title font-bold",
                compact ? "text-xl" : "text-2xl sm:text-3xl"
              )}
            >
              User not found
            </h1>
            <p className="user-not-found-view__message text-sm sm:text-base max-w-sm mx-auto">
              {message}
            </p>
            <p className="user-not-found-view__hint text-xs sm:text-sm max-w-xs mx-auto">
              {hint}
            </p>
          </div>

          {/* Actions */}
          <div className="flex flex-col sm:flex-row items-center justify-center gap-3 pt-2">
            <Link to="/registry" className="w-full sm:w-auto">
              <Button className="user-not-found-view__btn-primary w-full sm:w-auto gap-2 border-0 transition-all duration-200">
                <Zap className="w-4 h-4" />
                Browse Functions
              </Button>
            </Link>
            <Link to="/" className="w-full sm:w-auto">
              <Button
                variant="outline"
                className="user-not-found-view__btn-secondary w-full sm:w-auto gap-2"
              >
                <Home className="w-4 h-4" />
                Back to Home
              </Button>
            </Link>
          </div>

          {/* Optional search hint */}
          {is404 && username && (
            <p className="user-not-found-view__hint text-xs flex items-center justify-center gap-1.5 pt-2">
              <Search className="w-3.5 h-3.5" />
              Try searching for a different username in the registry
            </p>
          )}
        </div>
      </SpotlightCard>
    </motion.div>
  );
}
