'use client';

import { cn } from '@/lib/utils';
import { formatDate } from '@/lib/utils/date';
import { Badge } from '@/components/ui';
import type { DiaryEntry } from '@/types';

export interface EntryCardProps {
  entry: DiaryEntry;
  isSelected?: boolean;
  onClick?: () => void;
}

export function EntryCard({ entry, isSelected, onClick }: EntryCardProps) {
  // Get first tag as mood indicator
  const mood = entry.tags?.[0];

  // Get preview text (first 100 chars of body, stripped of markdown)
  const previewText = entry.body
    .replace(/[#*_`~\[\]]/g, '')
    .replace(/\n+/g, ' ')
    .trim()
    .slice(0, 100);

  return (
    <button
      onClick={onClick}
      className={cn(
        'w-full text-left p-4 border-l-4 transition-colors',
        'hover:bg-zinc-50 dark:hover:bg-zinc-800/50',
        isSelected
          ? 'border-l-zinc-900 bg-zinc-50 dark:border-l-white dark:bg-zinc-800/50'
          : 'border-l-transparent'
      )}
    >
      <h3 className="font-medium text-zinc-900 dark:text-white truncate">
        {entry.title || 'Untitled'}
      </h3>

      <div className="mt-1 flex items-center gap-2">
        <span className="text-sm text-zinc-500 dark:text-zinc-400">
          {formatDate(entry.date, 'MMM d, yyyy')}
        </span>
        {mood && (
          <Badge variant="outline" className="text-xs py-0.5">
            {mood}
          </Badge>
        )}
      </div>

      {previewText && (
        <p className="mt-2 text-sm text-zinc-600 dark:text-zinc-400 line-clamp-2">
          {previewText}...
        </p>
      )}
    </button>
  );
}
