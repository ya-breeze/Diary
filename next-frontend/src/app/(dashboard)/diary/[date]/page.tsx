'use client';

import { useEffect, useState } from 'react';
import { useParams, useSearchParams, useRouter } from 'next/navigation';
import { EntryViewer, EntryEditor } from '@/components/diary';
import { useDiaryEntry } from '@/hooks';

export default function DiaryEntryPage() {
  const params = useParams();
  const searchParams = useSearchParams();
  const router = useRouter();

  const date = params.date as string;
  const editParam = searchParams.get('edit');

  const { data: entry, isLoading, refetch } = useDiaryEntry(date);
  const [isEditing, setIsEditing] = useState(editParam === 'true');

  // Sync edit state with URL
  useEffect(() => {
    setIsEditing(editParam === 'true');
  }, [editParam]);

  const handleEdit = () => {
    router.push(`/diary/${date}?edit=true`);
  };

  const handleCloseEdit = () => {
    router.push(`/diary/${date}`);
  };

  const handleSaved = () => {
    refetch();
    handleCloseEdit();
  };


  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center p-8">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-zinc-300 border-t-zinc-900" />
      </div>
    );
  }

  // No entry found - show editor for new entry
  if (!entry) {
    return (
      <div className="h-full">
        <EntryEditor
          initialDate={date}
          onClose={() => router.push('/diary')}
          onSave={handleSaved}
        />
      </div>
    );
  }

  return (
    <>
      {/* Entry view */}
      <div className="mx-auto max-w-3xl px-4 py-6 md:px-8 md:py-8">
        {/* Entry content (Edit button is now inside) */}
        <EntryViewer entry={entry} onEdit={handleEdit} />
      </div>

      {/* Edit modal/page */}
      {isEditing && (
        <div className="fixed inset-0 z-50 bg-white dark:bg-zinc-900">
          <EntryEditor
            entry={entry}
            onClose={handleCloseEdit}
            onSave={handleSaved}
          />
        </div>
      )}
    </>
  );
}
