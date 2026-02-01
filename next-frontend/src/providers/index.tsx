'use client';

import { QueryProvider } from './QueryProvider';
import type { ReactNode } from 'react';

export function Providers({ children }: { children: ReactNode }) {
  return <QueryProvider>{children}</QueryProvider>;
}

export { QueryProvider };
