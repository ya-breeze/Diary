'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Search } from 'lucide-react';
import { Input } from '@/components/ui';
import { EntryCard } from '@/components/layout';
import { useSearchEntries } from '@/hooks';
import { useDebounce } from '@/hooks/useDebounce';

export default function SearchPage() {
  const router = useRouter();
  const [searchText, setSearchText] = useState('');
  const debouncedSearch = useDebounce(searchText, 300);

  const { data, isLoading } = useSearchEntries(debouncedSearch, debouncedSearch.length >= 2);

  const entries = data?.items || [];

  const handleSelectEntry = (date: string) => {
    router.push(`/diary/${date}`);
  };

  return (
    <div className="min-h-full bg-zinc-50 dark:bg-zinc-950">
      {/* Search header */}
      <div className="sticky top-0 z-10 border-b border-zinc-200 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-zinc-400" />
          <Input
            type="search"
            placeholder="Search your journal..."
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            className="pl-10"
            autoFocus
          />
        </div>
      </div>

      {/* Results */}
      <div className="bg-white dark:bg-zinc-900">
        {searchText.length < 2 ? (
          <div className="p-8 text-center text-zinc-500 dark:text-zinc-400">
            Enter at least 2 characters to search
          </div>
        ) : isLoading ? (
          <div className="p-8 text-center text-zinc-500">Searching...</div>
        ) : entries.length === 0 ? (
          <div className="p-8 text-center text-zinc-500 dark:text-zinc-400">
            No entries found for &quot;{searchText}&quot;
          </div>
        ) : (
          <div className="divide-y divide-zinc-200 dark:divide-zinc-800">
            <div className="px-4 py-2 text-sm text-zinc-500 dark:text-zinc-400">
              {data?.totalCount || entries.length} result{entries.length !== 1 ? 's' : ''} found
            </div>
            {entries.map((entry) => (
              <EntryCard
                key={entry.date}
                entry={entry}
                onClick={() => handleSelectEntry(entry.date)}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
