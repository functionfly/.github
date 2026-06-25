import React, { useState, useEffect, useCallback, createContext, useContext } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

type ToastStatus = 'ok' | 'pending' | 'revoked';

interface Toast {
  id: string;
  message: string;
  status: ToastStatus;
  duration?: number;
}

interface ToastContextValue {
  toasts: Toast[];
  addToast: (toast: Omit<Toast, 'id'>) => void;
  removeToast: (id: string) => void;
}

const ToastContext = createContext<ToastContextValue | null>(null);

interface ToastProviderProps {
  children: React.ReactNode;
}

export const ToastProvider: React.FC<ToastProviderProps> = ({ children }) => {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const addToast = useCallback((toast: Omit<Toast, 'id'>) => {
    const id = Math.random().toString(36).slice(2, 9);
    setToasts((prev) => [...prev, { ...toast, id }]);
  }, []);

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  return (
    <ToastContext.Provider value={{ toasts, addToast, removeToast }}>
      {children}
      <ToastContainer />
    </ToastContext.Provider>
  );
};

export const useToast = () => {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
};

const ToastContainer: React.FC = () => {
  const { toasts, removeToast } = useToast();

  return (
    <div
      className="fixed bottom-[var(--space-5)] right-[var(--space-5)] z-[var(--z-toast)] flex flex-col gap-[var(--space-3)]"
      aria-live="polite"
      aria-label="Notifications"
    >
      <AnimatePresence mode="popLayout">
        {toasts.map((toast) => (
          <ToastItem
            key={toast.id}
            toast={toast}
            onDismiss={() => removeToast(toast.id)}
          />
        ))}
      </AnimatePresence>
    </div>
  );
};

interface ToastItemProps {
  toast: Toast;
  onDismiss: () => void;
}

const STATUS_COLORS: Record<ToastStatus, string> = {
  ok: 'var(--status-ok)',
  pending: 'var(--status-pending)',
  revoked: 'var(--status-revoked)',
};

const ToastItem: React.FC<ToastItemProps> = ({ toast, onDismiss }) => {
  const [isPaused, setIsPaused] = useState(false);
  const duration = toast.duration ?? 5000;

  useEffect(() => {
    if (isPaused) return;

    const timer = setTimeout(onDismiss, duration);
    return () => clearTimeout(timer);
  }, [duration, onDismiss, isPaused]);

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 8, scale: 0.95 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      exit={{ opacity: 0, y: -8, scale: 0.95 }}
      transition={{ duration: 0.2, ease: 'easeOut' }}
      onMouseEnter={() => setIsPaused(true)}
      onMouseLeave={() => setIsPaused(false)}
      className="relative overflow-hidden"
      style={{
        background: 'var(--panel-raised)',
        borderRadius: 'var(--radius)',
        border: '1px solid var(--panel-edge)',
        boxShadow: 'var(--shadow-chamber)',
        paddingLeft: 'var(--space-4)',
        paddingRight: 'var(--space-4)',
        paddingTop: 'var(--space-3)',
        paddingBottom: 'var(--space-3)',
        minWidth: '280px',
        maxWidth: '400px',
      }}
    >
      <div
        className="absolute left-0 top-0 bottom-0 w-[3px]"
        style={{ backgroundColor: STATUS_COLORS[toast.status] }}
      />
      <div className="flex items-start gap-[var(--space-3)]">
        <span
          className="mt-[2px] w-[6px] h-[6px] rounded-full flex-shrink-0"
          style={{ backgroundColor: STATUS_COLORS[toast.status] }}
        />
        <p
          className="text-[13px] text-[var(--text-dim)] flex-1"
          role="alert"
        >
          {toast.message}
        </p>
        <button
          type="button"
          onClick={onDismiss}
          className="text-[var(--text-faint)] hover:text-[var(--text-dim)] transition-colors duration-[var(--duration-fast)] flex-shrink-0"
          aria-label="Dismiss notification"
        >
          <svg
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>
      </div>
    </motion.div>
  );
};
