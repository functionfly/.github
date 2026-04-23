import { CheckCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { motion, AnimatePresence } from "framer-motion";
import { useCallback, useEffect, useState } from "react";
import ReactConfetti from "react-confetti";

export interface ResolutionBannerProps {
  /** Short message, e.g. "This solution improved error rate by 12%" */
  message?: string;
  /** Reputation points awarded (optional; when integrated with flywheel) */
  reputationAwarded?: number;
  resolvedAt: string;
  /** Set to true to show the confetti celebration on mount */
  celebrate?: boolean;
  className?: string;
}

export function ResolutionBanner({
  message,
  reputationAwarded,
  resolvedAt,
  celebrate = true,
  className,
}: ResolutionBannerProps) {
  const [showConfetti, setShowConfetti] = useState(false);
  const [windowSize, setWindowSize] = useState({ width: 0, height: 0 });

  useEffect(() => {
    if (!celebrate) return;
    setShowConfetti(true);
    setWindowSize({ width: window.innerWidth, height: window.innerHeight });

    const handleResize = () => {
      setWindowSize({ width: window.innerWidth, height: window.innerHeight });
    };
    window.addEventListener("resize", handleResize);

    const timer = setTimeout(() => setShowConfetti(false), 4000);
    return () => {
      window.removeEventListener("resize", handleResize);
      clearTimeout(timer);
    };
  }, [celebrate]);

  return (
    <>
      <AnimatePresence>
        {showConfetti && (
          <ReactConfetti
            width={windowSize.width}
            height={windowSize.height}
            numberOfPieces={120}
            recycle={false}
            gravity={0.3}
            style={{ position: "fixed", top: 0, left: 0, zIndex: 9999, pointerEvents: "none" }}
          />
        )}
      </AnimatePresence>

      <motion.div
        initial={celebrate ? { opacity: 0, scale: 0.95, y: -8 } : false}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        transition={{ duration: 0.4, ease: "easeOut" }}
        className={cn(
          "flex flex-wrap items-center gap-2 rounded-lg border border-green-500/30 bg-green-500/10 px-3 py-2 text-sm",
          className
        )}
      >
        <motion.div
          initial={celebrate ? { scale: 0, rotate: -180 } : false}
          animate={{ scale: 1, rotate: 0 }}
          transition={{ delay: 0.2, type: "spring", stiffness: 200, damping: 15 }}
        >
          <CheckCircle className="h-4 w-4 text-green-600 shrink-0" />
        </motion.div>
        <span className="text-green-800 dark:text-green-200">
          {message ?? "Conversation resolved"}
        </span>
        {resolvedAt && (
          <span className="text-muted-foreground text-xs">
            {new Date(resolvedAt).toLocaleString()}
          </span>
        )}
        {reputationAwarded != null && reputationAwarded > 0 && (
          <motion.span
            initial={celebrate ? { opacity: 0, x: -10 } : false}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.5, duration: 0.3 }}
            className="font-medium text-green-700 dark:text-green-300"
          >
            +{reputationAwarded} Community Reputation awarded
          </motion.span>
        )}
      </motion.div>
    </>
  );
}
