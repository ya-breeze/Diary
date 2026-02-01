'use client';

import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import { Plus } from 'lucide-react';
import { Button } from '@/components/ui';
import { EntryCard } from '@/components/layout';
import { useDiaryEntries } from '@/hooks';
import { useIsMobile } from '@/hooks';
import { getTodayString } from '@/lib/utils/date';

export default function DiaryListPage() {
  const router = useRouter();
  const { data, isLoading } = useDiaryEntries();
  const isMobile = useIsMobile();

  const entries = data?.items || [];

  // On desktop, redirect to first entry or today
  useEffect(() => {
    if (!isMobile && entries.length > 0) {
      router.replace(`/diary/${entries[0].date}`);
    }
  }, [isMobile, entries, router]);

  const handleSelectEntry = (date: string) => {
    router.push(`/diary/${date}`);
  };

  const handleNewEntry = () => {
    router.push(`/diary/${getTodayString()}?edit=true`);
  };

  // Show entry list on mobile
  return (
    <div className="min-h-full bg-zinc-50 dark:bg-zinc-950">
      {/* Header for mobile list view */}
      <div className="sticky top-0 z-10 border-b border-zinc-200 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900 lg:hidden">
        <Button onClick={handleNewEntry} className="w-full gap-2">
          <Plus className="h-4 w-4" />
          New Entry
        </Button>
      </div>

      {/* Entry list */}
      <div className="divide-y divide-zinc-200 bg-white dark:divide-zinc-800 dark:bg-zinc-900">
        {isLoading ? (
          <div className="p-8 text-center text-zinc-500">Loading entries...</div>
        ) : entries.length === 0 ? (
          <div className="p-8 text-center">
            <p className="text-zinc-500 dark:text-zinc-400">No entries yet.</p>
            <p className="mt-1 text-sm text-zinc-400 dark:text-zinc-500">
              Create your first journal entry!
            </p>
            <Button onClick={handleNewEntry} className="mt-4 gap-2">
              <Plus className="h-4 w-4" />
              Create Entry
            </Button>
          </div>
        ) : (
          entries.map((entry) => (
            <EntryCard
              key={entry.date}
              entry={entry}
              onClick={() => handleSelectEntry(entry.date)}
            />
          ))
        )}
      </div>
    </div>
  );
}
