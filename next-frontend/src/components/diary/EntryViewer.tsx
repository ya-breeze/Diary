'use client';

import { useRef } from 'react';
import { useRouter } from 'next/navigation';
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
  const router = useRouter();
  const datePickerRef = useRef<HTMLInputElement>(null);

  const mood = entry.tags?.[0];

  const handleDateBadgeClick = () => {
    const input = datePickerRef.current;
    if (!input) return;
    input.value = entry.date;
    if (typeof input.showPicker === 'function') {
      input.showPicker();
    } else {
      input.click();
    }
  };

  const handleDateChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.value) {
      router.push(`/diary/${e.target.value}`);
    }
  };

  return (
    <article className={className}>
      {/* Header badges & Actions */}
      <div className="mb-8 flex flex-wrap items-center justify-between gap-x-4 gap-y-3 md:flex-nowrap">
        {/* Left: Tags */}
        <div className="order-1 flex w-full flex-wrap items-center gap-2 md:order-none md:w-auto md:flex-1 md:flex-nowrap md:overflow-hidden">
          {mood && <Badge variant="mood" className="shrink-0">{mood}</Badge>}
          <div className="flex flex-wrap gap-2">
            {entry.tags?.slice(1).map((tag) => (
              <Badge key={tag} variant="default" className="shrink-0">
                {tag}
              </Badge>
            ))}
          </div>
        </div>

        {/* Center: Date Navigation */}
        <div className="order-2 flex flex-none items-center gap-0.5 md:order-none">
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

          <div className="relative">
            <input
              ref={datePickerRef}
              type="date"
              onChange={handleDateChange}
              className="absolute opacity-0 pointer-events-none w-px h-px"
              tabIndex={-1}
              aria-hidden="true"
            />
            <Badge
              variant="outline"
              className="cursor-pointer gap-1.5 px-3 py-1.5 text-sm font-medium text-zinc-600 transition-colors hover:bg-zinc-100 hover:ring-1 hover:ring-zinc-300 dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:ring-zinc-600"
              onClick={handleDateBadgeClick}
              title="Jump to date"
            >
              <Calendar className="h-3.5 w-3.5" />
              {formatFullDate(entry.date)}
            </Badge>
          </div>

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
        <div className="order-3 ml-auto flex justify-end md:order-none md:ml-0 md:flex-1">
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
    </article>
  );
}
