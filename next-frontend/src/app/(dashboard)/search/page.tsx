'use client';

import { useMemo, useState } from 'react';
import { useRouter } from 'next/navigation';
import { Search, X } from 'lucide-react';
import { Input } from '@/components/ui';
import { EntryCard } from '@/components/layout';
import { useSearchEntries, useTagStats } from '@/hooks';
import { useDebounce } from '@/hooks/useDebounce';

export default function SearchPage() {
  const router = useRouter();
  const [searchText, setSearchText] = useState('');
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [tagInput, setTagInput] = useState('');
  const [tagFocused, setTagFocused] = useState(false);
  const debouncedSearch = useDebounce(searchText, 300);

  const { data: tagStats } = useTagStats();
  const knownTags = useMemo(() => (tagStats?.tags ?? []).map((t) => t.name), [tagStats]);

  // Text must reach 2 chars to search on its own, but a tag filter can run alone.
  const textReady = debouncedSearch.length >= 2;
  const hasTags = selectedTags.length > 0;
  const enabled = textReady || hasTags;

  const { data, isLoading } = useSearchEntries(
    textReady ? debouncedSearch : '',
    selectedTags.join(','),
    enabled
  );

  const entries = data?.items ?? [];

  const tagMatches = knownTags
    .filter((t) => !selectedTags.some((s) => s.toLowerCase() === t.toLowerCase()))
    .filter((t) =>
      tagInput.trim() === ''
        ? true
        : t.toLowerCase().includes(tagInput.trim().toLowerCase())
    )
    .slice(0, 8);

  const addTag = (raw: string) => {
    const name = raw.trim();
    setTagInput('');
    if (!name) return;
    if (selectedTags.some((t) => t.toLowerCase() === name.toLowerCase())) return;
    setSelectedTags((prev) => [...prev, name]);
  };

  const removeTag = (name: string) => {
    setSelectedTags((prev) => prev.filter((t) => t !== name));
  };

  const handleSelectEntry = (date: string) => {
    router.push(`/diary/${date}`);
  };

  return (
    <div className="min-h-full bg-zinc-50 dark:bg-zinc-950">
      {/* Search header */}
      <div className="sticky top-0 z-10 space-y-2 border-b border-zinc-200 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900">
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

        {/* Tag filter chips */}
        <div className="relative">
          <div className="flex flex-wrap items-center gap-1.5 rounded-lg border border-zinc-300 bg-white px-2 py-1.5 dark:border-zinc-700 dark:bg-zinc-800">
            <span className="px-1 text-xs text-zinc-400">Tags:</span>
            {selectedTags.map((tag) => (
              <span
                key={tag}
                className="inline-flex items-center gap-1 rounded-full bg-zinc-100 py-0.5 pl-2.5 pr-1 text-sm text-zinc-700 dark:bg-zinc-700 dark:text-zinc-200"
                data-testid="search-tag-chip"
              >
                {tag}
                <button
                  type="button"
                  onClick={() => removeTag(tag)}
                  className="rounded-full p-0.5 text-zinc-400 transition hover:bg-zinc-200 hover:text-zinc-700 dark:hover:bg-zinc-600 dark:hover:text-zinc-100"
                  aria-label={`Remove ${tag}`}
                  data-testid="search-tag-remove"
                >
                  <X className="h-3 w-3" />
                </button>
              </span>
            ))}
            <input
              type="text"
              value={tagInput}
              placeholder={selectedTags.length === 0 ? 'Filter by tag...' : ''}
              autoComplete="off"
              className="min-w-[6rem] flex-1 border-0 bg-transparent px-1 py-0.5 text-sm text-zinc-900 placeholder:text-zinc-400 focus:outline-none focus:ring-0 dark:text-white dark:placeholder:text-zinc-500"
              data-testid="search-tag-input"
              onChange={(e) => setTagInput(e.target.value)}
              onFocus={() => setTagFocused(true)}
              onBlur={() => setTimeout(() => setTagFocused(false), 150)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ',') {
                  e.preventDefault();
                  addTag(tagInput);
                } else if (e.key === 'Backspace' && tagInput === '' && selectedTags.length > 0) {
                  removeTag(selectedTags[selectedTags.length - 1]);
                }
              }}
            />
          </div>
          {tagFocused && tagMatches.length > 0 && (
            <ul
              className="absolute z-20 mt-1 max-h-56 w-full overflow-auto rounded-lg border border-zinc-200 bg-white py-1 shadow-lg dark:border-zinc-700 dark:bg-zinc-800"
              data-testid="search-tag-autocomplete"
            >
              {tagMatches.map((tag) => (
                <li key={tag}>
                  <button
                    type="button"
                    onMouseDown={(e) => {
                      e.preventDefault();
                      addTag(tag);
                    }}
                    className="block w-full px-3 py-1.5 text-left text-sm text-zinc-700 hover:bg-zinc-100 dark:text-zinc-200 dark:hover:bg-zinc-700"
                  >
                    {tag}
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      {/* Results */}
      <div className="bg-white dark:bg-zinc-900">
        {!enabled ? (
          <div className="p-8 text-center text-zinc-500 dark:text-zinc-400">
            Enter at least 2 characters or pick a tag to search
          </div>
        ) : isLoading ? (
          <div className="p-8 text-center text-zinc-500">Searching...</div>
        ) : entries.length === 0 ? (
          <div className="p-8 text-center text-zinc-500 dark:text-zinc-400">
            No entries found
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
