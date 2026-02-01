'use client';

import { Calendar, ChevronLeft, ChevronRight, Edit } from 'lucide-react';
import { Badge, Button } from '@/components/ui';
import { MarkdownRenderer } from './MarkdownRenderer';
import { formatFullDate } from '@/lib/utils/date';
import Link from 'next/link';
import type { DiaryEntry } from '@/types';

export interface EntryViewerProps {
  entry: DiaryEntry;
  className?: string;
  onEdit?: () => void;
}

export function EntryViewer({ entry, className, onEdit }: EntryViewerProps) {
  // Get mood from first tag
  const mood = entry.tags?.[0];

  return (
    <article className={className}>
      {/* Header badges & Actions */}
      <div className="mb-8 flex items-center justify-between gap-4">
        {/* Left: Tags */}
        <div className="flex flex-1 items-center gap-2 overflow-hidden">
          {mood && <Badge variant="mood" className="shrink-0">{mood}</Badge>}
          <div className="flex flex-wrap gap-2 overflow-hidden">
            {entry.tags?.slice(1).map((tag) => (
              <Badge key={tag} variant="default" className="shrink-0">
                {tag}
              </Badge>
            ))}
          </div>
        </div>

        {/* Center: Date Navigation */}
        <div className="flex flex-none items-center gap-0.5">
          {entry.previousDate ? (
            <Link
              href={`/diary/${entry.previousDate}`}
              className="flex h-9 w-9 items-center justify-center rounded-full text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-zinc-900 dark:hover:bg-zinc-800 dark:hover:text-white"
              title="Previous entry"
            >
              <ChevronLeft className="h-5 w-5" />
            </Link>
          ) : (
            <span className="flex h-9 w-9 items-center justify-center text-zinc-200 dark:text-zinc-800" aria-hidden="true">
              <ChevronLeft className="h-5 w-5" />
            </span>
          )}

          <Badge variant="outline" className="gap-1.5 px-3 py-1.5 text-sm font-medium text-zinc-600 dark:text-zinc-400">
            <Calendar className="h-3.5 w-3.5" />
            {formatFullDate(entry.date)}
          </Badge>

          {entry.nextDate ? (
            <Link
              href={`/diary/${entry.nextDate}`}
              className="flex h-9 w-9 items-center justify-center rounded-full text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-zinc-900 dark:hover:bg-zinc-800 dark:hover:text-white"
              title="Next entry"
            >
              <ChevronRight className="h-5 w-5" />
            </Link>
          ) : (
            <span className="flex h-9 w-9 items-center justify-center text-zinc-200 dark:text-zinc-800" aria-hidden="true">
              <ChevronRight className="h-5 w-5" />
            </span>
          )}
        </div>

        {/* Right: Edit Action */}
        <div className="flex flex-1 justify-end">
          {onEdit && (
            <Button
              variant="ghost"
              size="sm"
              onClick={onEdit}
              className="h-9 gap-1.5 text-zinc-500 hover:text-zinc-900 dark:hover:text-white"
            >
              <Edit className="h-4 w-4" />
              <span className="hidden sm:inline">Edit</span>
            </Button>
          )}
        </div>
      </div>

      {/* Title */}
      <h1 className="mb-6 font-serif text-3xl font-bold text-zinc-900 dark:text-white md:text-4xl">
        {entry.title || 'Untitled'}
      </h1>

      {/* Body */}
      <MarkdownRenderer content={entry.body} />
    </article >
  );
}
