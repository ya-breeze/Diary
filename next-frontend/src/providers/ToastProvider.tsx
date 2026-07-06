'use client';

import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react';
import { cn } from '@/lib/utils';

type ToastVariant = 'error' | 'success' | 'info';

interface ToastItem {
  id: number;
  message: string;
  variant: ToastVariant;
}

interface ToastApi {
  show: (message: string, variant?: ToastVariant) => void;
  error: (message: string) => void;
  success: (message: string) => void;
}

const ToastContext = createContext<ToastApi | null>(null);

const AUTO_DISMISS_MS = 5000;

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastItem[]>([]);
  const nextId = useRef(0);

  const remove = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const show = useCallback(
    (message: string, variant: ToastVariant = 'error') => {
      const id = nextId.current++;
      setToasts((prev) => [...prev, { id, message, variant }]);
      setTimeout(() => remove(id), AUTO_DISMISS_MS);
    },
    [remove]
  );

  const api = useMemo<ToastApi>(
    () => ({
      show,
      error: (message: string) => show(message, 'error'),
      success: (message: string) => show(message, 'success'),
    }),
    [show]
  );

  return (
    <ToastContext.Provider value={api}>
      {children}
      {/* Live region so screen readers announce new toasts without focus. */}
      <div
        role="alert"
        aria-live="assertive"
        aria-atomic="false"
        className="pointer-events-none fixed inset-x-0 top-4 z-50 flex flex-col items-center gap-2 px-4"
      >
        {toasts.map((toast) => (
          <div
            key={toast.id}
            data-testid={`toast-${toast.variant}`}
            className={cn(
              'pointer-events-auto flex w-full max-w-sm items-start justify-between gap-3 rounded-lg px-4 py-3 text-sm shadow-lg',
              toast.variant === 'error' && 'bg-red-600 text-white',
              toast.variant === 'success' && 'bg-green-600 text-white',
              toast.variant === 'info' && 'bg-zinc-800 text-white'
            )}
          >
            <span className="min-w-0 break-words">{toast.message}</span>
            <button
              type="button"
              onClick={() => remove(toast.id)}
              aria-label="Dismiss notification"
              className="shrink-0 text-white/80 hover:text-white"
            >
              ×
            </button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}

export function useToast(): ToastApi {
  const ctx = useContext(ToastContext);
  if (!ctx) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return ctx;
}
