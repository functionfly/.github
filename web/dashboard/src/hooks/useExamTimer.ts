import { useState, useEffect, useCallback, useRef } from 'react';

interface UseExamTimerOptions {
  expiresAt: string; // ISO 8601
  onExpire?: () => void;
}

interface UseExamTimerReturn {
  timeRemainingSeconds: number;
  isExpired: boolean;
  isWarning: boolean; // < 5 minutes
  isCritical: boolean; // < 1 minute
  formatted: string; // MM:SS or HH:MM:SS
  progress: number; // 0-100 (percentage of time elapsed)
}

export function useExamTimer({ expiresAt, onExpire }: UseExamTimerOptions): UseExamTimerReturn {
  const [timeRemaining, setTimeRemaining] = useState(() => {
    const diff = new Date(expiresAt).getTime() - Date.now();
    return Math.max(0, Math.floor(diff / 1000));
  });
  const onExpireRef = useRef(onExpire);
  onExpireRef.current = onExpire;

  useEffect(() => {
    const interval = setInterval(() => {
      const diff = new Date(expiresAt).getTime() - Date.now();
      const remaining = Math.max(0, Math.floor(diff / 1000));
      setTimeRemaining(remaining);

      if (remaining <= 0) {
        clearInterval(interval);
        onExpireRef.current?.();
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [expiresAt]);

  const formatTime = useCallback((seconds: number): string => {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = seconds % 60;

    if (h > 0) {
      return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
    }
    return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
  }, []);

  // Calculate total duration for progress
  const totalDuration = useRef(
    Math.floor((new Date(expiresAt).getTime() - Date.now()) / 1000) + timeRemaining
  );

  const elapsed = totalDuration.current - timeRemaining;
  const progress = totalDuration.current > 0 ? Math.min(100, (elapsed / totalDuration.current) * 100) : 100;

  return {
    timeRemainingSeconds: timeRemaining,
    isExpired: timeRemaining <= 0,
    isWarning: timeRemaining > 0 && timeRemaining <= 300, // < 5 min
    isCritical: timeRemaining > 0 && timeRemaining <= 60,  // < 1 min
    formatted: formatTime(timeRemaining),
    progress,
  };
}
