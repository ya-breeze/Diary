'use client';

import { Calendar, Clock } from 'lucide-react';
import { Badge } from '@/components/ui';
import { MarkdownRenderer } from './MarkdownRenderer';
import { formatFullDate, formatTime } from '@/lib/utils/date';
import type { DiaryEntry } from '@/types';

export interface EntryViewerProps {
  entry: DiaryEntry;
  className?: string;
}


export function EntryViewer({ entry, className }: EntryViewerProps) {
  // Get mood from first tag
  const mood = entry.tags?.[0];

  return (
    <article className={className}>
      {/* Header badges */}
      <div className="mb-4 flex flex-wrap items-center gap-2">
        <Badge variant="outline" className="gap-1.5">
          <Calendar className="h-3 w-3" />
          {formatFullDate(entry.date)}
        </Badge>

        <Badge variant="outline" className="gap-1.5">
          <Clock className="h-3 w-3" />
          {formatTime(new Date())}
        </Badge>

        {mood && <Badge variant="mood">{mood}</Badge>}

        {entry.tags?.slice(1).map((tag) => (
          <Badge key={tag} variant="default">
            {tag}
          </Badge>
        ))}
      </div>

      {/* Title */}
      <h1 className="mb-6 font-serif text-3xl font-bold text-zinc-900 dark:text-white md:text-4xl">
        {entry.title || 'Untitled'}
      </h1>

      {/* Body */}
      <MarkdownRenderer content={entry.body} />
    </article>
  );
}
