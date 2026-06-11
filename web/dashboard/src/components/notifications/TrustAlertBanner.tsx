/**
 * TrustAlertBanner Component
 *
 * A critical alert banner that appears when trust-related issues occur.
 * Supports dismissible alerts, expandable details, auto-collapse with progress bar,
 * and severity-based styling.
 */

'use client';

import React, { useState, useEffect, useCallback, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  ShieldAlert,
  AlertTriangle,
  ShieldQuestion,
  X,
  ChevronDown,
  ChevronUp,
  Pin,
  PinOff,
  Eye,
  FileText,
  RotateCcw,
  CheckCircle,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { TrustAlert } from '@/types/notifications';

// ============================================================================
// Types & Interfaces
// ============================================================================

interface TrustAlertBannerProps {
  alert?: TrustAlert;
  onDismiss?: (id: string) => void;
  onAction?: (action: string, alert: TrustAlert) => void;
  className?: string;
}

interface SeverityConfig {
  level: 'critical' | 'high' | 'medium' | 'low';
  icon: React.ReactNode;
  gradient: string;
  borderColor: string;
  glowColor: string;
  progressColor: string;
  textColor: string;
}

// ============================================================================
// Constants
// ============================================================================

const AUTO_COLLAPSE_DURATION = 10000; // 10 seconds
function getSeverityLevel(alert: TrustAlert): SeverityConfig['level'] {
  // Critical conditions
  if (
    alert.severity === 'emergency' ||
    alert.type === 'replay_failed' ||
    alert.type === 'determinism_broken' ||
    (alert.trustDropAmount && alert.trustDropAmount > 20)
  ) {
    return 'critical';
  }

  // High severity
  if (alert.trustDropAmount && alert.trustDropAmount > 10) {
    return 'high';
  }

  // Medium severity
  if (alert.trustDropAmount && alert.trustDropAmount > 5) {
    return 'medium';
  }

  // Low severity
  return 'low';
}

/**
 * Get severity configuration based on level
 */
function getSeverityConfig(level: SeverityConfig['level']): SeverityConfig {
  switch (level) {
    case 'critical':
      return {
        level: 'critical',
        icon: <ShieldAlert className="h-6 w-6" />,
        gradient: 'from-red-900/50 to-red-800/30',
        borderColor: 'border-red-500/50',
        glowColor: 'shadow-red-500/20',
        progressColor: 'bg-red-500',
        textColor: 'text-red-200',
      };
    case 'high':
      return {
        level: 'high',
        icon: <AlertTriangle className="h-6 w-6" />,
        gradient: 'from-amber-900/50 to-amber-800/30',
        borderColor: 'border-amber-500/50',
        glowColor: 'shadow-amber-500/20',
        progressColor: 'bg-amber-500',
        textColor: 'text-amber-200',
      };
    case 'medium':
      return {
        level: 'medium',
        icon: <ShieldQuestion className="h-6 w-6" />,
        gradient: 'from-blue-900/50 to-blue-800/30',
        borderColor: 'border-blue-500/50',
        glowColor: 'shadow-blue-500/20',
        progressColor: 'bg-blue-500',
        textColor: 'text-blue-200',
      };
    case 'low':
      return {
        level: 'low',
        icon: <ShieldQuestion className="h-5 w-5" />,
        gradient: 'from-gray-800/50 to-gray-700/30',
        borderColor: 'border-gray-500/50',
        glowColor: 'shadow-gray-500/20',
        progressColor: 'bg-gray-500',
        textColor: 'text-gray-300',
      };
  }
}



/**
 * Format timestamp for display
 */
function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  return date.toLocaleDateString();
}

// ============================================================================
// Component
// ============================================================================

export function TrustAlertBanner({
  alert,
  onDismiss,
  onAction,
  className,
}: TrustAlertBannerProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [isPinned, setIsPinned] = useState(false);
  const [progress, setProgress] = useState(100);
  const [isVisible, setIsVisible] = useState(true);
  const progressIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const autoCollapseTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Auto-collapse timer
  useEffect(() => {
    if (!alert || isPinned || isExpanded) {
      return;
    }

    const startTime = Date.now();
    const updateProgress = () => {
      const elapsed = Date.now() - startTime;
      const remaining = Math.max(0, AUTO_COLLAPSE_DURATION - elapsed);
      const newProgress = (remaining / AUTO_COLLAPSE_DURATION) * 100;
      setProgress(newProgress);

      if (remaining <= 0) {
        handleDismiss();
      }
    };

    progressIntervalRef.current = setInterval(updateProgress, 50);
    autoCollapseTimeoutRef.current = setTimeout(handleDismiss, AUTO_COLLAPSE_DURATION);

    return () => {
      if (progressIntervalRef.current) {
        clearInterval(progressIntervalRef.current);
      }
      if (autoCollapseTimeoutRef.current) {
        clearTimeout(autoCollapseTimeoutRef.current);
      }
    };
  }, [alert, isPinned, isExpanded]);

  const handleDismiss = useCallback(() => {
    if (alert) {
      // Call onDismiss to update Zustand store (source of truth for dismissal state)
      onDismiss?.(alert.id);
    }
    setIsVisible(false);
  }, [alert, onDismiss]);

  const handleAction = useCallback(
    (action: string) => {
      if (alert) {
        onAction?.(action, alert);
      }
    },
    [alert, onAction]
  );

  const toggleExpand = useCallback(() => {
    setIsExpanded((prev) => !prev);
  }, []);

  const togglePin = useCallback(() => {
    setIsPinned((prev) => !prev);
  }, []);

  // Don't render if no alert or visibility is hidden
  // Note: Dismissal state is managed by parent via onDismiss callback and Zustand store
  if (!alert || !isVisible) {
    return null;
  }

  const severityLevel = getSeverityLevel(alert);
  const config = getSeverityConfig(severityLevel);
  const isCritical = severityLevel === 'critical';

  // Determine available actions based on alert type
  const getActionButtons = () => {
    const buttons = [];

    buttons.push({
      id: 'view_details',
      label: 'View Details',
      icon: <Eye className="h-4 w-4 mr-1.5" />,
      variant: 'default' as const,
    });

    if (alert.type === 'replay_failed') {
      buttons.push({
        id: 'revert',
        label: 'Revert',
        icon: <RotateCcw className="h-4 w-4 mr-1.5" />,
        variant: 'outline' as const,
      });
    }

    buttons.push({
      id: 'check_logs',
      label: 'Check Logs',
      icon: <FileText className="h-4 w-4 mr-1.5" />,
      variant: 'outline' as const,
    });

    buttons.push({
      id: 'acknowledge',
      label: 'Acknowledge',
      icon: <CheckCircle className="h-4 w-4 mr-1.5" />,
      variant: 'ghost' as const,
    });

    return buttons;
  };

  const actionButtons = getActionButtons();

  return (
    <AnimatePresence>
      {isVisible && (
        <motion.div
          initial={{ y: -100, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          exit={{ y: -100, opacity: 0 }}
          transition={{
            type: 'spring',
            stiffness: 300,
            damping: 25,
          }}
          className={cn(
            'relative w-full overflow-hidden border-b backdrop-blur-sm',
            'bg-gradient-to-r',
            config.gradient,
            config.borderColor,
            isCritical && `shadow-lg ${config.glowColor}`,
            className
          )}
        >
          {/* Progress bar for auto-collapse */}
          {!isPinned && !isExpanded && (
            <motion.div
              className={cn('absolute bottom-0 left-0 h-0.5', config.progressColor)}
              initial={{ width: '100%' }}
              animate={{ width: `${progress}%` }}
              transition={{ duration: 0.05, ease: 'linear' }}
            />
          )}

          {/* Main content */}
          <div className="relative px-4 py-3 sm:px-6">
            <div className="flex items-start gap-3">
              {/* Icon with pulse animation for critical */}
              <div
                className={cn(
                  'flex-shrink-0 mt-0.5',
                  config.textColor,
                  isCritical && 'animate-pulse'
                )}
              >
                {config.icon}
              </div>

              {/* Content */}
              <div className="flex-1 min-w-0">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1">
                    {/* Header */}
                    <div className="flex items-center gap-2 flex-wrap">
                      <h3
                        className={cn(
                          'font-semibold text-sm sm:text-base',
                          config.textColor
                        )}
                      >
                        {alert.title}
                      </h3>
                      <Badge
                        variant="outline"
                        className={cn(
                          'text-xs capitalize',
                          config.borderColor,
                          config.textColor
                        )}
                      >
                        {severityLevel}
                      </Badge>
                      {alert.trustDropAmount && (
                        <Badge
                          variant="outline"
                          className={cn(
                            'text-xs',
                            config.borderColor,
                            config.textColor
                          )}
                        >
                          -{alert.trustDropAmount}% trust
                        </Badge>
                      )}
                    </div>

                    {/* Description */}
                    <p
                      className={cn(
                        'mt-1 text-sm opacity-90',
                        config.textColor
                      )}
                    >
                      {alert.description}
                    </p>

                    {/* Timestamp */}
                    <p
                      className={cn(
                        'mt-1 text-xs opacity-70',
                        config.textColor
                      )}
                    >
                      {formatTimestamp(alert.triggeredAt)}
                    </p>
                  </div>

                  {/* Action buttons row */}
                  <div className="flex items-center gap-1 flex-shrink-0">
                    {/* Pin button */}
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={togglePin}
                      className={cn(
                        'h-8 w-8',
                        config.textColor,
                        'hover:bg-white/10'
                      )}
                      title={isPinned ? 'Unpin alert' : 'Pin alert'}
                    >
                      {isPinned ? (
                        <PinOff className="h-4 w-4" />
                      ) : (
                        <Pin className="h-4 w-4" />
                      )}
                    </Button>

                    {/* Expand/collapse button */}
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={toggleExpand}
                      className={cn(
                        'h-8 w-8',
                        config.textColor,
                        'hover:bg-white/10'
                      )}
                      title={isExpanded ? 'Collapse details' : 'Expand details'}
                    >
                      {isExpanded ? (
                        <ChevronUp className="h-4 w-4" />
                      ) : (
                        <ChevronDown className="h-4 w-4" />
                      )}
                    </Button>

                    {/* Dismiss button */}
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={handleDismiss}
                      className={cn(
                        'h-8 w-8',
                        config.textColor,
                        'hover:bg-white/10'
                      )}
                      title="Dismiss alert"
                      aria-label="Dismiss alert"
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </div>
                </div>

                {/* Expanded details */}
                <AnimatePresence>
                  {isExpanded && (
                    <motion.div
                      initial={{ height: 0, opacity: 0 }}
                      animate={{ height: 'auto', opacity: 1 }}
                      exit={{ height: 0, opacity: 0 }}
                      transition={{ duration: 0.2 }}
                      className="overflow-hidden"
                    >
                      <div className="pt-3 mt-3 border-t border-white/10">
                        {/* Affected functions */}
                        {alert.affectedFunctions.length > 0 && (
                          <div className="mb-3">
                            <h4
                              className={cn(
                                'text-xs font-medium uppercase tracking-wider mb-2',
                                config.textColor,
                                'opacity-80'
                              )}
                            >
                              Affected Functions
                            </h4>
                            <div className="flex flex-wrap gap-1.5">
                              {alert.affectedFunctions.map((func, index) => (
                                <Badge
                                  key={func}
                                  variant="secondary"
                                  className="bg-white/10 text-white/90 hover:bg-white/20"
                                >
                                  {alert.affectedFunctionNames?.[index] || func}
                                </Badge>
                              ))}
                            </div>
                          </div>
                        )}

                        {/* Recommended action */}
                        <div className="mb-3">
                          <h4
                            className={cn(
                              'text-xs font-medium uppercase tracking-wider mb-1',
                              config.textColor,
                              'opacity-80'
                            )}
                          >
                            Recommended Action
                          </h4>
                          <p className={cn('text-sm', config.textColor)}>
                            {alert.recommendedAction}
                          </p>
                        </div>

                        {/* Trust score info */}
                        {alert.currentTrustScore !== undefined && (
                          <div className="mb-3">
                            <h4
                              className={cn(
                                'text-xs font-medium uppercase tracking-wider mb-1',
                                config.textColor,
                                'opacity-80'
                              )}
                            >
                              Current Trust Score
                            </h4>
                            <div className="flex items-center gap-2">
                              <div className="flex-1 h-2 bg-white/20 rounded-full overflow-hidden">
                                <div
                                  className={cn(
                                    'h-full rounded-full transition-all duration-500',
                                    alert.currentTrustScore > 80
                                      ? 'bg-green-500'
                                      : alert.currentTrustScore > 50
                                      ? 'bg-yellow-500'
                                      : 'bg-red-500'
                                  )}
                                  style={{
                                    width: `${alert.currentTrustScore}%`,
                                  }}
                                />
                              </div>
                              <span
                                className={cn(
                                  'text-sm font-medium',
                                  config.textColor
                                )}
                              >
                                {alert.currentTrustScore}%
                              </span>
                            </div>
                          </div>
                        )}

                        {/* Action buttons */}
                        <div className="flex flex-wrap gap-2 pt-2">
                          {actionButtons.map((button) => (
                            <Button
                              key={button.id}
                              variant={button.variant}
                              size="sm"
                              onClick={() => handleAction(button.id)}
                              className={cn(
                                button.variant === 'default' &&
                                  severityLevel === 'critical' &&
                                  'bg-red-600 hover:bg-red-700',
                                button.variant === 'default' &&
                                  severityLevel === 'high' &&
                                  'bg-amber-600 hover:bg-amber-700',
                                button.variant === 'default' &&
                                  severityLevel === 'medium' &&
                                  'bg-blue-600 hover:bg-blue-700',
                                button.variant === 'outline' &&
                                  'border-white/30 text-white hover:bg-white/10',
                                button.variant === 'ghost' &&
                                  'text-white/70 hover:text-white hover:bg-white/10'
                              )}
                            >
                              {button.icon}
                              {button.label}
                            </Button>
                          ))}
                        </div>
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>
            </div>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}

export default TrustAlertBanner;
