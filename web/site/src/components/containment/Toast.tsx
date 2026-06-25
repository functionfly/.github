import type { ReactNode } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

interface ToastProps {
  isVisible: boolean;
  onDismiss: () => void;
  status: 'live' | 'pending' | 'revoked';
  title: string;
  message?: string;
}

export function Toast({ isVisible, onDismiss, status, title, message }: ToastProps) {
  return (
    <AnimatePresence>
      {isVisible && (
        <motion.div
          className="toast"
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: 8 }}
          transition={{ duration: 0.2 }}
        >
          <div className={`toast__accent toast__accent--${status}`} />
          <div style={{ flex: 1 }}>
            <div className="toast__title">{title}</div>
            {message && <div className="toast__message">{message}</div>}
          </div>
          <button className="toast__dismiss" onClick={onDismiss} aria-label="Dismiss">
            ×
          </button>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
