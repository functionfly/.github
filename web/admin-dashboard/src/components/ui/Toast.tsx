import React, { createContext, useContext, useState, useCallback } from 'react';
import { cn } from '@/lib/utils';
import { X, CheckCircle, AlertCircle, Info, AlertTriangle } from 'lucide-react';

interface Toast {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  title: string;
  description?: React.ReactNode;
  duration?: number;
}

interface ToastContextType {
  toasts: Toast[];
  showToast: (toast: Omit<Toast, 'id'>) => void;
  dismissToast: (id: string) => void;
}

const ToastContext = createContext<ToastContextType | null>(null);

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);

  const showToast = useCallback((toast: Omit<Toast, 'id'>) => {
    const id = Math.random().toString(36).substring(7);
    const newToast: Toast = {
      ...toast,
      id,
      duration: toast.duration ?? 5000,
    };

    setToasts((prev) => [...prev, newToast]);

    // Auto dismiss
    if (newToast.duration !== Infinity) {
      setTimeout(() => {
        dismissToast(id);
      }, newToast.duration);
    }
  }, []);

  const dismissToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  return (
    <ToastContext.Provider value={{ toasts, showToast, dismissToast }}>
      {children}
      <ToastContainer toasts={toasts} onDismiss={dismissToast} />
    </ToastContext.Provider>
  );
}

export function useToast() {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
}

interface ToastContainerProps {
  toasts: Toast[];
  onDismiss: (id: string) => void;
}

function ToastContainer({ toasts, onDismiss }: ToastContainerProps) {
  return (
    <div className="fixed top-4 right-4 z-50 flex flex-col gap-2 w-full max-w-sm">
      {toasts.map((toast) => (
        <ToastItem key={toast.id} toast={toast} onDismiss={() => onDismiss(toast.id)} />
      ))}
    </div>
  );
}

interface ToastItemProps {
  toast: Toast;
  onDismiss: () => void;
}

function ToastItem({ toast, onDismiss }: ToastItemProps) {
  const iconConfig = {
    success: { icon: CheckCircle, className: 'text-green-600 bg-green-100' },
    error: { icon: AlertCircle, className: 'text-red-600 bg-red-100' },
    warning: { icon: AlertTriangle, className: 'text-yellow-600 bg-yellow-100' },
    info: { icon: Info, className: 'text-blue-600 bg-blue-100' },
  };

  const { icon: Icon, className: iconClass } = iconConfig[toast.type];

  return (
    <div
      className={cn(
        'flex items-start gap-3 p-4 bg-white rounded-lg shadow-lg border',
        'animate-in slide-in-from-right-full duration-300',
        toast.type === 'success' && 'border-green-200',
        toast.type === 'error' && 'border-red-200',
        toast.type === 'warning' && 'border-yellow-200',
        toast.type === 'info' && 'border-blue-200'
      )}
      role="alert"
    >
      <div className={cn('flex items-center justify-center w-8 h-8 rounded-full shrink-0', iconClass)}>
        <Icon className="w-5 h-5" />
      </div>
      <div className="flex-1 min-w-0">
        <h4 className="font-semibold text-gray-900">{toast.title}</h4>
         {toast.description && (
           <p className="mt-0.5 text-sm text-gray-600">{toast.description}</p>
         )}
      </div>
      <button
        onClick={onDismiss}
        className="shrink-0 p-1 text-gray-400 hover:text-gray-600 transition-colors"
        aria-label="Dismiss notification"
      >
        <X className="w-4 h-4" />
      </button>
    </div>
  );
}

// Convenience hooks for common toast patterns
export function useToastHelpers() {
  const { showToast } = useToast();

  return {
    success: (title: string, arg2?: string | React.ReactNode | { description?: string | React.ReactNode; duration?: number }) => {
      const isOpt = typeof arg2 === 'object' && arg2 !== null && !('type' in (arg2 as any));
      const description = isOpt ? (arg2 as any).description : arg2;
      const duration = isOpt ? (arg2 as any).duration : undefined;
      showToast({ type: 'success', title, description, duration });
    },
    error: (title: string, arg2?: string | React.ReactNode | { description?: string | React.ReactNode; duration?: number }) => {
      const isOpt = typeof arg2 === 'object' && arg2 !== null && !('type' in (arg2 as any));
      const description = isOpt ? (arg2 as any).description : arg2;
      const duration = isOpt ? (arg2 as any).duration : 8000;
      showToast({ type: 'error', title, description, duration });
    },
    warning: (title: string, arg2?: string | React.ReactNode | { description?: string | React.ReactNode; duration?: number }) => {
      const isOpt = typeof arg2 === 'object' && arg2 !== null && !('type' in (arg2 as any));
      const description = isOpt ? (arg2 as any).description : arg2;
      showToast({ type: 'warning', title, description });
    },
    info: (title: string, arg2?: string | React.ReactNode | { description?: string | React.ReactNode; duration?: number }) => {
      const isOpt = typeof arg2 === 'object' && arg2 !== null && !('type' in (arg2 as any));
      const description = isOpt ? (arg2 as any).description : arg2;
      showToast({ type: 'info', title, description });
    },
  };
}
