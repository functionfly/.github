import React, { useEffect, useRef, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  children: React.ReactNode;
  className?: string;
}

export const Modal: React.FC<ModalProps> = ({
  isOpen,
  onClose,
  children,
  className = '',
}) => {
  const modalRef = useRef<HTMLDivElement>(null);
  const previousActiveElement = useRef<HTMLElement | null>(null);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }

      if (e.key === 'Tab' && modalRef.current) {
        const focusableElements = modalRef.current.querySelectorAll<HTMLElement>(
          'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
        );
        const firstElement = focusableElements[0];
        const lastElement = focusableElements[focusableElements.length - 1];

        if (e.shiftKey && document.activeElement === firstElement) {
          e.preventDefault();
          lastElement?.focus();
        } else if (!e.shiftKey && document.activeElement === lastElement) {
          e.preventDefault();
          firstElement?.focus();
        }
      }
    },
    [onClose]
  );

  useEffect(() => {
    if (isOpen) {
      previousActiveElement.current = document.activeElement as HTMLElement;
      document.addEventListener('keydown', handleKeyDown);
      document.body.style.overflow = 'hidden';

      const firstFocusable = modalRef.current?.querySelector<HTMLElement>(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
      );
      firstFocusable?.focus();
    }

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      document.body.style.overflow = '';
      previousActiveElement.current?.focus();
    };
  }, [isOpen, handleKeyDown]);

  return (
    <AnimatePresence>
      {isOpen && (
        <>
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.2, ease: 'easeOut' }}
            className="fixed inset-0 bg-[rgba(10,13,17,0.7)] z-[var(--z-modal)]"
            onClick={onClose}
            aria-hidden="true"
          />
          <div
            ref={modalRef}
            role="dialog"
            aria-modal="true"
            className={`fixed inset-0 z-[var(--z-modal)] flex items-center justify-center p-[var(--space-4)] ${className}`}
          >
            <motion.div
              initial={{ opacity: 0, scale: 0.98 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.98 }}
              transition={{ duration: 0.2, ease: 'easeOut' }}
              className="relative w-full max-w-[480px]"
              onClick={(e) => e.stopPropagation()}
            >
              <div
                className="relative"
                style={{
                  background: 'var(--panel)',
                  borderRadius: 'var(--radius-lg)',
                  border: '1px solid var(--panel-edge)',
                  boxShadow: 'var(--shadow-chamber)',
                  padding: 'var(--space-7)',
                }}
              >
                {children}
              </div>
            </motion.div>
          </div>
        </>
      )}
    </AnimatePresence>
  );
};

interface ModalHeaderProps {
  children: React.ReactNode;
}

export const ModalHeader: React.FC<ModalHeaderProps> = ({ children }) => {
  return (
    <div className="mb-[var(--space-5)]">
      <h2
        className="text-[22px] font-[var(--font-display)] font-medium text-[var(--text)]"
        id="modal-title"
      >
        {children}
      </h2>
    </div>
  );
};

interface ModalBodyProps {
  children: React.ReactNode;
}

export const ModalBody: React.FC<ModalBodyProps> = ({ children }) => {
  return <div className="text-[15px] text-[var(--text-dim)]">{children}</div>;
};

interface ModalFooterProps {
  children: React.ReactNode;
}

export const ModalFooter: React.FC<ModalFooterProps> = ({ children }) => {
  return (
    <div
      className="mt-[var(--space-6)] flex justify-end gap-[var(--space-3)]"
      onClick={(e) => e.stopPropagation()}
    >
      {children}
    </div>
  );
};

interface ModalCloseButtonProps {
  onClose: () => void;
}

export const ModalCloseButton: React.FC<ModalCloseButtonProps> = ({
  onClose,
}) => {
  return (
    <button
      type="button"
      onClick={onClose}
      className="absolute top-[var(--space-4)] right-[var(--space-4)] text-[var(--text-dim)] hover:text-[var(--text)] transition-colors duration-[var(--duration-fast)]"
      aria-label="Close modal"
    >
      <svg
        width="20"
        height="20"
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
  );
};
