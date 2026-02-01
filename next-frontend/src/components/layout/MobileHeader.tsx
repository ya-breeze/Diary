'use client';

import { Menu, BookOpen } from 'lucide-react';
import { cn } from '@/lib/utils';

export interface MobileHeaderProps {
  title?: string;
  onMenuClick: () => void;
  className?: string;
  rightContent?: React.ReactNode;
}

export function MobileHeader({
  title = 'My Journal',
  onMenuClick,
  className,
  rightContent,
}: MobileHeaderProps) {
  return (
    <header
      className={cn(
        'sticky top-0 z-30 flex h-14 items-center justify-between border-b border-zinc-200 bg-white px-4 dark:border-zinc-800 dark:bg-zinc-900',
        className
      )}
    >
      <div className="flex items-center gap-3">
        <button
          onClick={onMenuClick}
          className="rounded-lg p-2 text-zinc-600 hover:bg-zinc-100 dark:text-zinc-400 dark:hover:bg-zinc-800"
          aria-label="Open menu"
        >
          <Menu className="h-5 w-5" />
        </button>

        <div className="flex items-center gap-2">
          <BookOpen className="h-5 w-5 text-zinc-900 dark:text-white" />
          <span className="font-semibold text-zinc-900 dark:text-white">
            {title}
          </span>
        </div>
      </div>

      {rightContent && <div className="flex items-center gap-2">{rightContent}</div>}
    </header>
  );
}
