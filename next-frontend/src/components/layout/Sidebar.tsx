'use client';

import { useRouter, usePathname } from 'next/navigation';
import { Plus, BookOpen } from 'lucide-react';
import { Button } from '@/components/ui';
import { EntryCard } from './EntryCard';
import { useDiaryEntries } from '@/hooks';
import { getTodayString } from '@/lib/utils/date';
import type { DiaryEntry } from '@/types';

export interface SidebarProps {
  selectedDate?: string | null;
  onSelectEntry?: (date: string) => void;
  className?: string;
}

export function Sidebar({ selectedDate, onSelectEntry, className }: SidebarProps) {
  const router = useRouter();
  const pathname = usePathname();
  const { data, isLoading } = useDiaryEntries();

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
        </div>

        <Button onClick={handleNewEntry} className="w-full gap-2">
          <Plus className="h-4 w-4" />
          New Entry
        </Button>
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
    </aside>
  );
}
