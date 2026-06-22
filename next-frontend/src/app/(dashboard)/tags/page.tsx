'use client';

import { Suspense, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { ArrowLeft, Pencil, Trash2, Check, X, Tag } from 'lucide-react';
import { Button, Modal } from '@/components/ui';
import { EntryCard } from '@/components/layout';
import {
  useTagStats,
  useRenameTag,
  useDeleteTag,
  useSearchEntries,
} from '@/hooks';
import type { TagStat } from '@/types';

export default function TagsPage() {
  // useSearchParams requires a Suspense boundary in the App Router.
  return (
    <Suspense fallback={<div className="p-8 text-center text-zinc-500">Loading…</div>}>
      <TagsPageInner />
    </Suspense>
  );
}

function TagsPageInner() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { data, isLoading } = useTagStats();
  const renameTag = useRenameTag();
  const deleteTag = useDeleteTag();

  // The browsed tag lives in the URL (?tag=) so the view is deep-linkable.
  const browseTag = searchParams.get('tag');
  const [editing, setEditing] = useState<string | null>(null);
  const [editValue, setEditValue] = useState('');
  const [pendingDelete, setPendingDelete] = useState<TagStat | null>(null);
  const [error, setError] = useState<string | null>(null);

  const tags = data?.tags ?? [];

  const startRename = (name: string) => {
    setError(null);
    setEditing(name);
    setEditValue(name);
  };

  const submitRename = async (name: string) => {
    const newName = editValue.trim();
    if (!newName || newName === name) {
      setEditing(null);
      return;
    }
    try {
      await renameTag.mutateAsync({ name, newName });
      setEditing(null);
    } catch {
      setError(`Failed to rename "${name}".`);
    }
  };

  const confirmDelete = async () => {
    if (!pendingDelete) return;
    try {
      await deleteTag.mutateAsync(pendingDelete.name);
      setPendingDelete(null);
    } catch {
      setError(`Failed to delete "${pendingDelete.name}".`);
      setPendingDelete(null);
    }
  };

  if (browseTag) {
    return <TagBrowse tag={browseTag} onBack={() => router.push('/tags')} />;
  }

  return (
    <div className="min-h-full bg-zinc-50 dark:bg-zinc-950">
      {/* Header */}
      <div className="sticky top-0 z-10 flex items-center gap-3 border-b border-zinc-200 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900">
        <button
          onClick={() => router.back()}
          className="rounded-md p-1 text-zinc-500 hover:bg-zinc-100 dark:text-zinc-400 dark:hover:bg-zinc-800"
          aria-label="Back"
        >
          <ArrowLeft className="h-5 w-5" />
        </button>
        <h1 className="flex items-center gap-2 text-lg font-semibold text-zinc-900 dark:text-white">
          <Tag className="h-5 w-5 text-zinc-400" />
          Tags{!isLoading && ` (${tags.length})`}
        </h1>
      </div>

      {error && (
        <p className="px-4 pt-3 text-sm text-red-600 dark:text-red-400">{error}</p>
      )}

      <div className="p-4">
        {isLoading ? (
          <p className="py-8 text-center text-zinc-500">Loading…</p>
        ) : tags.length === 0 ? (
          <p className="py-8 text-center text-zinc-500 dark:text-zinc-400">
            No tags yet. Add tags to your entries to see them here.
          </p>
        ) : (
          <ul className="divide-y divide-zinc-200 overflow-hidden rounded-xl border border-zinc-200 bg-white dark:divide-zinc-800 dark:border-zinc-800 dark:bg-zinc-900">
            {tags.map((stat) => (
              <li
                key={stat.name}
                className="flex items-center gap-2 px-4 py-3"
                data-testid="tag-row"
              >
                {editing === stat.name ? (
                  <>
                    <input
                      autoFocus
                      value={editValue}
                      onChange={(e) => setEditValue(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') void submitRename(stat.name);
                        if (e.key === 'Escape') setEditing(null);
                      }}
                      className="flex-1 rounded-md border border-zinc-300 bg-white px-2 py-1 text-sm text-zinc-900 focus:border-zinc-500 focus:outline-none focus:ring-1 focus:ring-zinc-500 dark:border-zinc-700 dark:bg-zinc-800 dark:text-white"
                      data-testid="tag-rename-input"
                    />
                    <button
                      onClick={() => void submitRename(stat.name)}
                      disabled={renameTag.isPending}
                      className="rounded-md p-1.5 text-green-600 hover:bg-green-50 disabled:opacity-50 dark:hover:bg-green-900/30"
                      aria-label="Save"
                      data-testid="tag-rename-save"
                    >
                      <Check className="h-4 w-4" />
                    </button>
                    <button
                      onClick={() => setEditing(null)}
                      className="rounded-md p-1.5 text-zinc-500 hover:bg-zinc-100 dark:hover:bg-zinc-800"
                      aria-label="Cancel"
                    >
                      <X className="h-4 w-4" />
                    </button>
                  </>
                ) : (
                  <>
                    <button
                      onClick={() => router.push(`/tags?tag=${encodeURIComponent(stat.name)}`)}
                      className="flex-1 text-left text-sm font-medium text-zinc-900 hover:underline dark:text-zinc-100"
                      data-testid="tag-browse"
                    >
                      {stat.name}
                    </button>
                    <span
                      className="rounded-full bg-zinc-100 px-2 py-0.5 text-xs text-zinc-600 dark:bg-zinc-800 dark:text-zinc-400"
                      data-testid="tag-count"
                    >
                      {stat.count}
                    </span>
                    <button
                      onClick={() => startRename(stat.name)}
                      className="rounded-md p-1.5 text-zinc-400 hover:bg-zinc-100 hover:text-zinc-700 dark:hover:bg-zinc-800 dark:hover:text-zinc-200"
                      aria-label={`Rename ${stat.name}`}
                      data-testid="tag-rename"
                    >
                      <Pencil className="h-4 w-4" />
                    </button>
                    <button
                      onClick={() => { setError(null); setPendingDelete(stat); }}
                      className="rounded-md p-1.5 text-zinc-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/30 dark:hover:text-red-400"
                      aria-label={`Delete ${stat.name}`}
                      data-testid="tag-delete"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Delete confirmation */}
      <Modal
        isOpen={!!pendingDelete}
        onClose={() => setPendingDelete(null)}
        title="Delete tag"
        fullScreenOnMobile={false}
      >
        <div className="p-6">
          <p className="mb-6 text-zinc-600 dark:text-zinc-400">
            Remove the tag{' '}
            <span className="font-medium text-zinc-900 dark:text-white">
              {pendingDelete?.name}
            </span>{' '}
            from {pendingDelete?.count} entr{pendingDelete?.count === 1 ? 'y' : 'ies'}? This
            cannot be undone.
          </p>
          <div className="flex flex-col gap-3">
            <Button
              onClick={() => void confirmDelete()}
              isLoading={deleteTag.isPending}
              className="w-full bg-red-600 hover:bg-red-700"
              data-testid="tag-delete-confirm"
            >
              Delete
            </Button>
            <Button onClick={() => setPendingDelete(null)} variant="ghost" className="w-full">
              Cancel
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}

function TagBrowse({ tag, onBack }: { tag: string; onBack: () => void }) {
  const router = useRouter();
  const { data, isLoading } = useSearchEntries('', tag);
  const entries = data?.items ?? [];

  return (
    <div className="min-h-full bg-zinc-50 dark:bg-zinc-950">
      <div className="sticky top-0 z-10 flex items-center gap-3 border-b border-zinc-200 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900">
        <button
          onClick={onBack}
          className="rounded-md p-1 text-zinc-500 hover:bg-zinc-100 dark:text-zinc-400 dark:hover:bg-zinc-800"
          aria-label="Back to tags"
        >
          <ArrowLeft className="h-5 w-5" />
        </button>
        <h1 className="flex items-center gap-2 text-lg font-semibold text-zinc-900 dark:text-white">
          <Tag className="h-5 w-5 text-zinc-400" />
          {tag}
        </h1>
      </div>

      <div className="bg-white dark:bg-zinc-900">
        {isLoading ? (
          <p className="p-8 text-center text-zinc-500">Loading…</p>
        ) : entries.length === 0 ? (
          <p className="p-8 text-center text-zinc-500 dark:text-zinc-400">
            No entries tagged &quot;{tag}&quot;.
          </p>
        ) : (
          <div className="divide-y divide-zinc-200 dark:divide-zinc-800">
            <div className="px-4 py-2 text-sm text-zinc-500 dark:text-zinc-400">
              {data?.totalCount ?? entries.length} entr
              {(data?.totalCount ?? entries.length) === 1 ? 'y' : 'ies'}
            </div>
            {entries.map((entry) => (
              <EntryCard
                key={entry.date}
                entry={entry}
                onClick={() => router.push(`/diary/${entry.date}`)}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
