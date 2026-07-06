'use client';

import { QueryProvider } from './QueryProvider';
import { ToastProvider } from './ToastProvider';
import type { ReactNode } from 'react';

export function Providers({ children }: { children: ReactNode }) {
  return (
    <QueryProvider>
      <ToastProvider>{children}</ToastProvider>
    </QueryProvider>
  );
}

export { QueryProvider, ToastProvider };
export { useToast } from './ToastProvider';
