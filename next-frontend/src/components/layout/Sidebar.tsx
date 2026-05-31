'use client';

import { useRef } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { Plus, Calendar, BookOpen, ShieldAlert, Search, User } from 'lucide-react';
import { Button } from '@/components/ui';
import { EntryCard } from './EntryCard';
import { useDiaryEntries, useHealthIssues } from '@/hooks';
import { getTodayString } from '@/lib/utils/date';
import { cn } from '@/lib/utils';
import type { DiaryEntry } from '@/types';

export interface SidebarProps {
  selectedDate?: string | null;
  onSelectEntry?: (date: string) => void;
  onHealthClick?: () => void;
  className?: string;
}

export function Sidebar({ selectedDate, onSelectEntry, onHealthClick, className }: SidebarProps) {
  const router = useRouter();
  const pathname = usePathname();
  const datePickerRef = useRef<HTMLInputElement>(null);
  const { data, isLoading } = useDiaryEntries();
  const { data: healthData } = useHealthIssues();

  const issueCount = healthData?.issues?.length ?? 0;

  const entries = data?.items || [];

  const handleSelectEntry = (date: string) => {
    router.push(`/diary/${date}`);
    // Also call callback (e.g., to close drawer on mobile)
    onSelectEntry?.(date);
  };

  const handleNewEntry = () => {
    const today = getTodayString();
    router.push(`/diary/${today}?edit=true`);
  };

  const handleOpenDatePicker = () => {
    const input = datePickerRef.current;
    if (!input) return;
    input.value = getTodayString();
    if (typeof input.showPicker === 'function') {
      input.showPicker();
    } else {
      input.click();
    }
  };

  const handleDatePickerChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.value) {
      router.push(`/diary/${e.target.value}`);
      onSelectEntry?.(e.target.value);
    }
  };

  // Determine selected date from URL if not provided
  const currentSelectedDate = selectedDate ?? pathname?.match(/\/diary\/(\d{4}-\d{2}-\d{2})/)?.[1];

  return (
    <aside className={className}>
      {/* Header */}
      <div className="p-4 border-b border-zinc-200 dark:border-zinc-800">
        <div className="flex items-center gap-2 mb-4">
          <BookOpen className="h-6 w-6 text-zinc-900 dark:text-white" />
          <h1 className="text-xl font-semibold text-zinc-900 dark:text-white">
            Diary
          </h1>
          {issueCount > 0 && (
            <button
              onClick={onHealthClick}
              className="relative ml-auto flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-amber-600 hover:bg-amber-50 dark:text-amber-400 dark:hover:bg-amber-900/20"
              title={`${issueCount} storage issue${issueCount > 1 ? 's' : ''}`}
            >
              <ShieldAlert className="h-4 w-4" />
              <span>{issueCount}</span>
            </button>
          )}
        </div>

        <div className="flex gap-1">
          <Button onClick={handleNewEntry} className="flex-1 gap-2">
            <Plus className="h-4 w-4" />
            New Entry
          </Button>
          <div className="relative">
            <input
              ref={datePickerRef}
              type="date"
              onChange={handleDatePickerChange}
              className="absolute opacity-0 pointer-events-none w-px h-px"
              tabIndex={-1}
              aria-hidden="true"
            />
            <Button
              variant="secondary"
              onClick={handleOpenDatePicker}
              title="Jump to date"
              className="h-full px-2.5"
            >
              <Calendar className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>

      {/* Entry List */}
      <div className="flex-1 overflow-y-auto">
        {isLoading ? (
          <div className="p-4 text-center text-zinc-500">Loading...</div>
        ) : entries.length === 0 ? (
          <div className="p-4 text-center text-zinc-500 dark:text-zinc-400">
            <p>No entries yet.</p>
            <p className="text-sm mt-1">Create your first entry!</p>
          </div>
        ) : (
          <div className="divide-y divide-zinc-100 dark:divide-zinc-800">
            {entries.map((entry: DiaryEntry) => (
              <EntryCard
                key={entry.date}
                entry={entry}
                isSelected={currentSelectedDate === entry.date}
                onClick={() => handleSelectEntry(entry.date)}
              />
            ))}
          </div>
        )}
      </div>

      {/* Footer nav */}
      <div className="border-t border-zinc-200 dark:border-zinc-800 p-2">
        {[
          { icon: Search, label: 'Search', href: '/search' },
          { icon: User, label: 'Profile', href: '/profile' },
        ].map(({ icon: Icon, label, href }) => (
          <button
            key={label}
            onClick={() => router.push(href)}
            className={cn(
              'flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
              pathname === href
                ? 'bg-zinc-100 text-zinc-900 dark:bg-zinc-800 dark:text-white'
                : 'text-zinc-500 hover:bg-zinc-100 hover:text-zinc-900 dark:text-zinc-400 dark:hover:bg-zinc-800 dark:hover:text-white'
            )}
          >
            <Icon className="h-4 w-4" />
            {label}
          </button>
        ))}
      </div>
    </aside>
  );
}

